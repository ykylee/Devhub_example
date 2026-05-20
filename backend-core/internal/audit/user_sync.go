package audit

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/httpapi"
	"github.com/devhub/backend-core/internal/store"
)

// ADR-0020 sub-carve C (sprint -k, issue #212) — Keycloak admin event 처리
// 시 DevHub `users` 컬럼 자동 sync. event listener (sprint -u~-y) 가 audit_logs
// 만 emit 하던 동작을 본 sprint 가 확장.
//
// 흐름 (ADR-0020 §5.3.1 매핑):
//   USER:UPDATE 이벤트
//     → KeycloakAdminClient.GetUserDetails(identityID)
//     → KeycloakAdminClient.GetUserGroups(identityID)
//     → DevHub users.email / display_name / status / role 갱신
//     → metric `devhub_keycloak_user_sync_total{action="profile"}` increment
//
//   USER:DELETE 이벤트
//     → users.status = deactivated (soft delete, ADR-0020 §5.3.1 P1-2 결정)
//     → metric `_sync_total{action="status"}` increment
//
//   GROUP_MEMBERSHIP CREATE / DELETE 이벤트
//     → GetUserGroups(identityID)
//     → users.role 재계산 (group composite role priority filter)
//     → metric `_sync_total{action="membership"}` increment

// UserSyncOrgStore is the narrow interface the user sync helpers depend on.
// Implemented by *store.PostgresStore in production; tests inject a memory
// fake. Keeps the audit package import graph small.
//
// GetUserByIdPSubject 는 sprint -k codex P1 hotfix 로 추가 — USER:DELETE 처리
// 시 admin client GetUserDetails 가 404 (user 이미 gone) 라 idp_subject 컬럼
// (Keycloak UUID) 으로 DevHub users 행 lookup 이 유일 경로.
type UserSyncOrgStore interface {
	GetUser(ctx context.Context, userID string) (domain.AppUser, error)
	GetUserByIdPSubject(ctx context.Context, identityID string) (domain.AppUser, error)
	UpdateUser(ctx context.Context, userID string, input domain.UpdateUserInput) (domain.AppUser, error)
}

// UserSyncAdminClient is the narrow interface for the Keycloak admin client.
// Tests inject a fake; production passes *httpapi.KeycloakAdminClient (the
// methods were added in sprint -k Commit 1).
type UserSyncAdminClient interface {
	GetUserDetails(ctx context.Context, identityID string) (httpapi.KeycloakUserDetails, error)
	GetUserGroups(ctx context.Context, identityID string) ([]httpapi.KeycloakGroup, error)
}

// SyncUserAction labels the user_sync metric — `profile` (USER:UPDATE),
// `membership` (GROUP_MEMBERSHIP), `status` (USER:DELETE).
type SyncUserAction string

const (
	SyncActionProfile    SyncUserAction = "profile"
	SyncActionMembership SyncUserAction = "membership"
	SyncActionStatus     SyncUserAction = "status"
)

// SyncUserProfile pulls the latest Keycloak user state and writes
// email/display_name/status onto the DevHub `users` row that matches the
// Keycloak username (= preferred_username = DevHub user_id).
//
// Best-effort. Returns nil even when the DevHub row is missing — lazy
// auto-create (ADR-0020 §5.2, sprint -i) handles first-login provisioning,
// and missing rows here are normal for pre-onboarding accounts.
func SyncUserProfile(ctx context.Context, admin UserSyncAdminClient, orgs UserSyncOrgStore, identityID string) error {
	details, err := admin.GetUserDetails(ctx, identityID)
	if err != nil {
		return fmt.Errorf("user_sync profile fetch %s: %w", identityID, err)
	}
	userID := strings.TrimSpace(details.Username)
	if userID == "" {
		return fmt.Errorf("user_sync profile %s: empty username (Keycloak response missing preferred_username mapper?)", identityID)
	}

	// Look up the DevHub row. Missing row → lazy auto-create scope.
	// PR #241 codex P1 hotfix: ErrNotFound 만 noop, 그 외 (DB 장애 등) propagate
	// — caller 가 metric `_errors_total` 으로 노출. 이전엔 모든 err 를 swallow 해
	// stale data silent 위험.
	if _, err := orgs.GetUser(ctx, userID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("user_sync profile lookup %s: %w", userID, err)
	}

	input := domain.UpdateUserInput{}
	if email := strings.TrimSpace(details.Email); email != "" {
		input.Email = &email
	}
	if display := composeDisplayName(details); display != "" {
		input.DisplayName = &display
	}
	status := domain.UserStatusActive
	if !details.Enabled {
		status = domain.UserStatusDeactivated
	}
	input.Status = &status

	if _, err := orgs.UpdateUser(ctx, userID, input); err != nil {
		return fmt.Errorf("user_sync profile update %s: %w", userID, err)
	}
	return nil
}

// SyncUserMembership pulls the latest group membership for the identity and
// updates the DevHub `users.role` column. Group → role mapping follows
// keycloak_groups_rbac_mapping.md (sprint -f recommended option B — group
// composite realm role priority filter).
//
// Group names with the `devhub-` prefix are treated as DevHub role anchors;
// the rest are ignored. Priority order matches `extractKeycloakRole` of the
// JWT verifier (system_admin > pmo_manager > manager > developer).
func SyncUserMembership(ctx context.Context, admin UserSyncAdminClient, orgs UserSyncOrgStore, identityID string) error {
	groups, err := admin.GetUserGroups(ctx, identityID)
	if err != nil {
		return fmt.Errorf("user_sync membership fetch %s: %w", identityID, err)
	}
	details, err := admin.GetUserDetails(ctx, identityID)
	if err != nil {
		return fmt.Errorf("user_sync membership lookup user %s: %w", identityID, err)
	}
	userID := strings.TrimSpace(details.Username)
	if userID == "" {
		return fmt.Errorf("user_sync membership %s: empty username", identityID)
	}

	// PR #241 codex P1 hotfix: ErrNotFound 만 noop, 그 외 propagate.
	if _, err := orgs.GetUser(ctx, userID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("user_sync membership lookup %s: %w", userID, err)
	}

	role := pickHighestPriorityRole(groups)
	if role == "" {
		// 본인 token claim 기반 role 도 빈 경우 default fallback (developer).
		// ADR-0020 §5.2.2 의 lazyAutoCreateDefaultRole 정합.
		role = string(domain.AppRoleDeveloper)
	}
	roleVal := domain.AppRole(role)
	input := domain.UpdateUserInput{Role: &roleVal}
	if _, err := orgs.UpdateUser(ctx, userID, input); err != nil {
		return fmt.Errorf("user_sync membership update %s: %w", userID, err)
	}
	return nil
}

// MarkUserDeactivated handles USER:DELETE — DevHub does a soft delete by
// setting users.status = deactivated (ADR-0020 §5.3.1 P1-2 decision). The
// DevHub row is preserved so historical audit_logs.actor_login references do
// not dangle.
//
// PR #241 codex P1 hotfix: 받는 인자는 Keycloak identity_id (ResourcePath 의
// users/{id} 에서 파싱된 UUID) — admin client `GetUserDetails` 가 404 (user
// 이미 gone) 이므로 DevHub `users.idp_subject` 컬럼 (UNIQUE) lookup 으로
// user_id 추출. ErrNotFound 만 noop, 그 외 (DB 장애) propagate.
//
// 이전 (PR #241 머지 직전 시점) 의 잘못된 동작: identity UUID 를 user_id 로
// 직접 GetUser 호출 → 매칭 안 됨 → silent noop + metric success → stale data
// 위험. codex inline review P1 이 회귀 식별.
func MarkUserDeactivated(ctx context.Context, orgs UserSyncOrgStore, identityID string) error {
	if strings.TrimSpace(identityID) == "" {
		return fmt.Errorf("user_sync delete: empty identity_id")
	}
	user, err := orgs.GetUserByIdPSubject(ctx, identityID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			// DevHub row 없음 (pre-onboarding 또는 idp_subject lazy backfill 미완) → noop.
			return nil
		}
		return fmt.Errorf("user_sync delete lookup idp_subject=%s: %w", identityID, err)
	}
	deactivated := domain.UserStatusDeactivated
	input := domain.UpdateUserInput{Status: &deactivated}
	if _, err := orgs.UpdateUser(ctx, user.UserID, input); err != nil {
		return fmt.Errorf("user_sync delete %s (idp_subject=%s): %w", user.UserID, identityID, err)
	}
	return nil
}

// composeDisplayName mirrors httpapi/keycloak_verifier.extractDisplayName —
// prefer `firstName lastName`, fall back to whichever is non-empty.
func composeDisplayName(d httpapi.KeycloakUserDetails) string {
	first := strings.TrimSpace(d.FirstName)
	last := strings.TrimSpace(d.LastName)
	switch {
	case first != "" && last != "":
		return first + " " + last
	case first != "":
		return first
	case last != "":
		return last
	}
	return ""
}

// pickHighestPriorityRole reads DevHub-prefixed group names and returns the
// highest-priority role. order: system_admin > pmo_manager > manager >
// developer. Mirrors `extractKeycloakRole` priority filter (sprint -j PR
// #185). Group names follow `devhub-{role}s` plural convention per
// keycloak_groups_rbac_mapping.md (e.g. `devhub-system-admins`).
func pickHighestPriorityRole(groups []httpapi.KeycloakGroup) string {
	priority := map[string]int{
		string(domain.AppRoleSystemAdmin): 4,
		"pmo_manager":                     3,
		string(domain.AppRoleManager):     2,
		string(domain.AppRoleDeveloper):   1,
	}
	best := ""
	bestRank := 0
	for _, g := range groups {
		role := groupNameToRole(g.Name)
		if role == "" {
			continue
		}
		if rank := priority[role]; rank > bestRank {
			best = role
			bestRank = rank
		}
	}
	return best
}

// groupNameToRole maps Keycloak group names (e.g. `devhub-system-admins`) to
// DevHub role identifiers. Returns "" for unrecognized groups.
func groupNameToRole(name string) string {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "devhub-system-admins", "devhub-system-admin":
		return string(domain.AppRoleSystemAdmin)
	case "devhub-pmo-managers", "devhub-pmo-manager":
		return "pmo_manager"
	case "devhub-managers", "devhub-manager":
		return string(domain.AppRoleManager)
	case "devhub-developers", "devhub-developer":
		return string(domain.AppRoleDeveloper)
	}
	return ""
}

// ParseIdentityIDFromResourcePath extracts the Keycloak user UUID from an
// admin event ResourcePath like `users/abc-uuid` or
// `users/abc-uuid/role-mappings/realm`. Returns "" when the path does not
// start with `users/`.
func ParseIdentityIDFromResourcePath(resourcePath string) string {
	const prefix = "users/"
	if !strings.HasPrefix(resourcePath, prefix) {
		return ""
	}
	rest := resourcePath[len(prefix):]
	if idx := strings.IndexByte(rest, '/'); idx >= 0 {
		return rest[:idx]
	}
	return rest
}

// Compile-time check — *store.PostgresStore (the production org store) must
// satisfy UserSyncOrgStore. If this breaks, the postgres store signature
// drifted.
var _ UserSyncOrgStore = (*store.PostgresStore)(nil)
