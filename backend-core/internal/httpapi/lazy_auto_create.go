package httpapi

import (
	"strings"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/gin-gonic/gin"
)

// ADR-0020 sub-carve B (sprint -i, issue #209) — lazy auto-create. Keycloak
// admin console 또는 HRDB ETL 로 신규 생성된 user 의 첫 DevHub 진입 시
// `users` row 자동 생성. ADR-0020 §5.2 따름.
//
// 흐름:
//   authenticateActor → token 검증 통과 → GetUser(login) → ErrNotFound
//   → lazyAutoCreateUser → CreateUser + SetIdPSubject + audit emit
//   → finalRole = lazy default 또는 token claim role
//   → c.Next()
//
// 결정 항목 (ADR-0020 §5.2.2):
//   - role 매핑: token realm_access.roles → actor.Role (verifier 가 priority filter).
//     빈 경우 default lazyAutoCreateDefaultRole + audit "user.role_default_assigned"
//   - status: lazyAutoCreateDefaultStatus = active (Keycloak enabled=true 인 token 발급 정합)
//   - unit 매핑: unassigned (primary_unit_id=NULL). admin 후속 배치 또는 HRDB pre-stage
//   - audit action: "account.lazy_provisioned" (DB row 정합 신규 row 만 emit)

// lazyAutoCreateDefaultRole 은 token claim 에 매핑 가능한 role 이 없을 때
// 사용하는 fallback. rbac_policies seed role 4종 중 가장 권한이 약한 developer
// 채택 (ADR-0020 §5.2.2 결정 P1-3).
const lazyAutoCreateDefaultRole = domain.AppRoleDeveloper

// lazyAutoCreateDefaultStatus = active. Keycloak enabled=true 인 token 발급
// 정합. ADR-0020 §5.2.2.
const lazyAutoCreateDefaultStatus = domain.UserStatusActive

// lazyAutoCreateAuditAction is the audit_logs row action for the lazy
// auto-create event. emit-once (다음 진입 시 GetUser hit → 본 분기 미도달).
const lazyAutoCreateAuditAction = "account.lazy_provisioned"

// lazyRoleDefaultAuditAction is emitted alongside lazy_provisioned when the
// token did not carry a mappable Keycloak role and the default developer
// fallback was applied. 별도 audit row 로 발급되어 운영자가 어떤 user 가
// fallback 으로 진입했는지 식별 가능.
const lazyRoleDefaultAuditAction = "user.role_default_assigned"

// lazyAutoCreateUser provisions the DevHub users row + emits audit. Caller
// must have already verified the bearer token and confirmed GetUser returned
// store.ErrNotFound. OrganizationStore != nil 가 caller 책임.
//
// Updates c context (devhub_actor_role) with the resolved role so downstream
// enforceRoutePermission middleware sees consistent state. logRequest 로
// 모든 결정 path 가시화.
func (h Handler) lazyAutoCreateUser(c *gin.Context, login string, actor AuthenticatedActor, fallbackRole string) {
	ctx := c.Request.Context()
	role := domain.AppRole(strings.TrimSpace(actor.Role))
	roleDefaulted := false
	if !isValidLazyRole(role) {
		logRequest(c, "[authenticateActor] lazy auto-create %q: role %q not in rbac_policies seed, applying default %q (ADR-0020 §5.2.2)", login, actor.Role, lazyAutoCreateDefaultRole)
		role = lazyAutoCreateDefaultRole
		roleDefaulted = true
	}

	input := domain.CreateUserInput{
		UserID:      login,
		Email:       strings.TrimSpace(actor.Email),
		DisplayName: strings.TrimSpace(actor.DisplayName),
		Role:        role,
		Status:      lazyAutoCreateDefaultStatus,
		Type:        domain.UserTypeHuman,
		// primary_unit_id / current_unit_id 미할당 (NULL) — ADR-0020 §5.2.2.
		JoinedAt: time.Now().UTC(),
	}
	if input.DisplayName == "" {
		input.DisplayName = login // fallback for audit/UI legibility
	}

	created, err := h.cfg.OrganizationStore.CreateUser(ctx, input)
	if err != nil {
		// Race or schema drift. fall through to token role claim — pre-sprint -i
		// 동작과 정합. log noise 는 surface 보존.
		logRequest(c, "[authenticateActor] lazy auto-create %q failed: %v (fall through to token role claim %q)", login, err, fallbackRole)
		return
	}

	// SetIdPSubject best-effort — CreateUserInput 에 IdPSubject 필드 없어 별도 호출.
	if strings.TrimSpace(actor.Subject) != "" {
		if setErr := h.cfg.OrganizationStore.SetIdPSubject(ctx, login, actor.Subject); setErr != nil {
			logRequest(c, "[authenticateActor] lazy auto-create %q SetIdPSubject failed: %v (best-effort, next request retry)", login, setErr)
		}
	}

	// audit emit (best-effort). recordAuditBestEffort 가 fail 해도 user row 는 보존.
	h.recordAuditBestEffort(c, lazyAutoCreateAuditAction, "user", login, map[string]any{
		"user_id":       created.UserID,
		"email":         created.Email,
		"display_name":  created.DisplayName,
		"role":          string(created.Role),
		"role_source":   roleSourceLabel(roleDefaulted, actor.Role),
		"idp_subject":   actor.Subject,
		"trigger":       "first_login_after_external_provisioning",
	})
	if roleDefaulted {
		h.recordAuditBestEffort(c, lazyRoleDefaultAuditAction, "user", login, map[string]any{
			"user_id":       created.UserID,
			"applied_role":  string(role),
			"token_role":    actor.Role,
			"reason":        "token_realm_access_roles_empty_or_unmapped",
		})
	}

	// downstream middleware (enforceRoutePermission) 가 본 role 을 본다.
	c.Set("devhub_actor_role", string(created.Role))
	logRequest(c, "[authenticateActor] lazy auto-create %q -> role=%q status=%q idp_subject=%q (ADR-0020 §5.2)", login, created.Role, created.Status, actor.Subject)
}

// isValidLazyRole returns true when the role string maps to a known
// rbac_policies seed role. pmo_manager / developer / manager / system_admin
// (organization.go 의 validAppRoles 와 동일 set + lower-case match).
func isValidLazyRole(role domain.AppRole) bool {
	switch role {
	case domain.AppRoleDeveloper, domain.AppRoleManager, domain.AppRoleSystemAdmin, "pmo_manager":
		return true
	}
	return false
}

func roleSourceLabel(defaulted bool, tokenRole string) string {
	if defaulted {
		return "default_fallback"
	}
	if strings.TrimSpace(tokenRole) == "" {
		return "default_fallback"
	}
	return "token_realm_access"
}
