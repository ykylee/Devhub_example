package view

import (
	"context"
	"errors"
	"github.com/devhub/backend-core/internal/shared/authkey"
	"github.com/devhub/backend-core/internal/shared/httphelp"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/shared/metrics"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

var ErrInvalidBearerToken = errors.New("invalid bearer token")

type AuthenticatedActor struct {
	Login   string
	Subject string
	Role    string
	// ADR-0020 sub-carve B (sprint -i, issue #209) — lazy auto-create 가 첫
	// 진입 시 DevHub `users` row 생성에 사용. token claim (email, name)
	// 에서 추출. 빈 값 허용 — 이후 event listener (sub-carve C, sprint -g)
	// 가 USER:UPDATE event 로 sync.
	Email       string
	DisplayName string
}

func normalizeSystemRoleAlias(role string) string {
	switch strings.TrimSpace(role) {
	case "manager", "team_manager":
		return "team_manager"
	case "user":
		// ADR-0026: Keycloak has only a single 'user' realm role; the actual
		// DevHub role comes from DB users.role. Fall back to developer when
		// no DB row exists (token-only actor).
		return "developer"
	default:
		return strings.TrimSpace(role)
	}
}

func isGenericKeycloakRole(role string) bool {
	role = strings.TrimSpace(role)
	if role == "" {
		return true
	}
	switch role {
	case "user", "default-roles-devhub", "offline_access", "uma_authorization":
		return true
	}
	return strings.HasPrefix(role, "default-roles-")
}

type BearerTokenVerifier interface {
	VerifyBearerToken(context.Context, string) (AuthenticatedActor, error)
}

// publicAPIPaths lists /api/v1 routes that pass through authenticateActor
// without an Authorization header.
var publicAPIPaths = map[string]bool{
	"/api/v1/integrations/gitea/webhooks":                true,
	"/api/v1/integration/providers/:provider_id/webhook": true,
	"/api/v1/infra/services/snapshot":                    true,
}

func (h *AuthHandler) AuthenticateActor(c *gin.Context) {
	c.Set("devhub_auth_dev_fallback", h.cfg.AuthDevFallback)

	// ADR-0006 (2026-05-13): legacy X-Devhub-Actor inbound header is rejected
	// outright. ADR-0004 declared the fallback closed at the prod-code level;
	// this middleware turns silent ignore into explicit 400 so client-side
	// usage of the dead header is surfaced immediately. Negative tests in
	// audit_test / commands_test / auth_test / me_test still hold — the
	// header was already being ignored; ADR-0006 only changes "ignore" into
	// "reject + helpful error".
	if c.GetHeader("X-Devhub-Actor") != "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status": "rejected",
			"error":  "X-Devhub-Actor header is removed; use Authorization: Bearer ... (see ADR-0004 / ADR-0006)",
			"code":   "x_devhub_actor_removed",
		})
		return
	}

	if publicAPIPaths[c.FullPath()] {
		// Webhook bypass paths run without a Bearer token but still produce
		// audit rows. Tag the source so downstream recordAudit picks the
		// right enum.
		switch c.FullPath() {
		case "/api/v1/integrations/gitea/webhooks":
			c.Set(httphelp.CtxKeySourceType, domain.AuditSourceWebhook)
		case "/api/v1/integration/providers/:provider_id/webhook":
			c.Set(httphelp.CtxKeySourceType, domain.AuditSourceWebhook)
		default:
			c.Set(httphelp.CtxKeySourceType, domain.AuditSourceSystem)
		}
		c.Next()
		return
	}

	header := strings.TrimSpace(c.GetHeader("Authorization"))
	if header == "" && c.FullPath() == "/api/v1/realtime/ws" {
		// Browser WebSocket API cannot set arbitrary headers like Authorization
		// (ADR-0024). Ticket pattern only (single-use + 60s TTL) — the legacy
		// `?access_token=` query fallback was removed in the ticket-only cutover
		// (ADR-0024 §6 carve 5). No ticket / invalid ticket → 401.
		if h.cfg.RealtimeTickets != nil {
			if raw := strings.TrimSpace(c.Query("ticket")); raw != "" {
				entry, ok, err := h.cfg.RealtimeTickets.Consume(c.Request.Context(), raw)
				if err != nil {
					// Store fault (e.g. transient Postgres outage): the ticket may
					// be valid — do NOT collapse into 401, which would reject a
					// legitimate user and hide the infra signal. 503 instead.
					log.Printf("[realtime] ticket consume store error: %v", err)
					c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
						"status": "unavailable",
						"error":  "realtime ticket store error",
					})
					return
				}
				if ok {
					c.Set("devhub_actor_login", entry.ActorLogin)
					if entry.ActorRole != "" {
						c.Set("devhub_actor_role", entry.ActorRole)
					}
					c.Set(httphelp.CtxKeySourceType, entry.SourceType)
					c.Next()
					return
				}
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"status": "unauthenticated",
					"error":  "realtime ticket invalid or expired",
				})
				return
			}
		}
	}
	if header == "" {
		if h.cfg.AuthDevFallback {
			c.Header("X-Devhub-Auth", "dev_fallback_no_header")
			c.Set(httphelp.CtxKeySourceType, domain.AuditSourceSystem)
			c.Set("devhub_actor_role", "system_admin")
			c.Next()
			return
		}
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"status": "unauthenticated",
			"error":  "authorization header required",
		})
		return
	}

	token, ok := bearerToken(header)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"status": "unauthenticated",
			"error":  "authorization header must use Bearer scheme",
		})
		return
	}

	// ADR-0029: API key fallback. JWT 포맷 (header.payload.signature) 이 아니고
	// cfg.APIKey 가 설정되어 있으면, 정적 키 비교로 인증 통과. Keycloak 도달
	// 불필요. caller actor 는 "api-key" (role=system_admin) 으로 식별 —
	// enforceRoutePermission / RBAC 가 admin-only endpoints 만 허용.
	//
	// ADR-0029 §6 (f) P3 multi-key — cfg.APIKeyStore 가 nil 이 아니면 DB
	// branch 우선 (sha256(token) → GetAPIKeyByHash → revoked/expired/CIDR 검증).
	// env cfg.APIKey (단일 키) 는 migration 기간 동안 legacy fallback 으로
	// 잔존 — DB store 가 nil 일 때만 활성. multi-key 가 primary.
	if h.cfg.APIKeyStore != nil && !looksLikeJWT(token) {
		keyHash := authkey.HashAPIKey(token)
		key, err := h.cfg.APIKeyStore.GetAPIKeyByHash(c.Request.Context(), keyHash)
		if err == nil {
			// CIDR allowlist 검증. nil allowlist = all IPs 허용.
			clientIP := c.ClientIP()
			if allowed, err := authkey.IsCIDRAllowed(clientIP, key.AllowedCIDRs); err == nil && allowed {
				c.Set("devhub_actor_login", "api-key:"+key.KeyPrefix)
				c.Set("devhub_actor_role", "system_admin")
				c.Set("devhub_auth_source", "api_key_db")
				c.Set(httphelp.CtxKeySourceType, domain.AuditSourceSystem)
				c.Set("devhub_api_key_id", key.ID)
				c.Header("X-Devhub-Auth", "api_key_db")
				httphelp.LogRequest(c, "[authenticateActor] API key (DB) authenticated for path: %s prefix=%s", c.FullPath(), key.KeyPrefix)
				// ADR-0029 §6 (g) — audit + metric enrich. SOP [`docs/setup/api_key_rotation.md` §6.1]
				// 의 4 audit 정합. DB branch: api_key_id + key_prefix + allowed_cidrs 분기
				// payload 에 명시 — multi-key 운영 시 key 단위 가시성.
				h.recordAuditBestEffort(c, "auth.api_key_authenticated", "auth", "api-key:"+key.KeyPrefix, map[string]any{
					"actor_role":    "system_admin",
					"path":          c.FullPath(),
					"method":        c.Request.Method,
					"client_ip":     clientIP,
					"request_id":    httphelp.RequestIDFrom(c),
					"api_key_id":    key.ID,
					"key_prefix":    key.KeyPrefix,
					"auth_branch":   "db_multi_key",
				})
				// SOP §6.1 metric 정합 — auth success counter. label value `db` 로
				// env static branch 와 분리.
				metrics.DevhubAPIKeyAuthTotal.WithLabelValues("success_db").Inc()
				// best-effort last_used_at UPDATE — 실패 시 인증 자체는 유지
				// (DB outage 도 인증 통과 — SOP §3.4 정합).
				go func(keyID string) {
					ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
					defer cancel()
					_ = h.cfg.APIKeyStore.UpdateLastUsedAt(ctx, keyID, time.Now())
				}(key.ID)
				c.Next()
				return
			}
			// CIDR allowlist fail — audit emit + 403 (not 401 — auth 는
			// 통과했으나 policy deny. SOP §6.1 의 4 audit 분기 정합).
			h.recordAuditBestEffort(c, "auth.api_key_denied", "auth", "api-key:"+key.KeyPrefix, map[string]any{
				"reason":      "cidr_not_allowed",
				"path":        c.FullPath(),
				"client_ip":   c.ClientIP(),
				"api_key_id":  key.ID,
				"key_prefix":  key.KeyPrefix,
				"auth_branch": "db_multi_key",
			})
			metrics.DevhubAPIKeyAuthTotal.WithLabelValues("denied_cidr").Inc()
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"status": "forbidden",
				"error":  "api key client IP not in allowed_cidrs",
			})
			return
		}
		// DB lookup miss (revoked / expired / not found) — fallback to env
		// branch (아래) 가 시도. 동일 key 로 env static compare 도 가능.
		// 단 audit 로 denied 분기는 skip — env branch 가 401 로 reject.
	}
	if h.cfg.APIKey != "" && !looksLikeJWT(token) {
		if subtleEqual(token, h.cfg.APIKey) {
			c.Set("devhub_actor_login", "api-key")
			c.Set("devhub_actor_role", "system_admin")
			c.Set("devhub_auth_source", "api_key")
			c.Set(httphelp.CtxKeySourceType, domain.AuditSourceSystem)
			c.Header("X-Devhub-Auth", "api_key")
			httphelp.LogRequest(c, "[authenticateActor] API key authenticated for path: %s", c.FullPath())
			// ADR-0029 §6 (g) — audit + metric enrich. SOP [`docs/setup/api_key_rotation.md` §6.1]
			// 의 4 audit 정합. `auth.api_key_authenticated` 1 row per request emit
			// (best-effort — DB outage 도 인증은 통과). payload 에 `actor_role` /
			// `path` / `client_ip` / `request_id` 명시. request_id 는 httphelp.RequestIDFrom
			// 로 cross-correlation. metric increment 는 `c.Next()` 후에 middleware 가
			// 처리 — 본 middleware 는 인증 직후. SOP §6.1 metric 정의 정합.
			h.recordAuditBestEffort(c, "auth.api_key_authenticated", "auth", "api-key", map[string]any{
				"actor_role": "system_admin",
				"path":       c.FullPath(),
				"method":     c.Request.Method,
				"client_ip":  c.ClientIP(),
				"request_id": httphelp.RequestIDFrom(c),
			})
			// SOP §6.1 metric 정합 — auth success counter.
			metrics.DevhubAPIKeyAuthTotal.WithLabelValues("success").Inc()
			c.Next()
			return
		}
		// ADR-0029 §6 (g) — invalid key 거부 시 audit emit (best-effort). SOP §6.1 의
		// 4 audit 정합. `auth.api_key_denied` (system_admin emit, invalid key).
		h.recordAuditBestEffort(c, "auth.api_key_denied", "auth", "api-key", map[string]any{
			"reason":     "invalid_key",
			"path":       c.FullPath(),
			"method":     c.Request.Method,
			"client_ip":  c.ClientIP(),
			"request_id": httphelp.RequestIDFrom(c),
		})
		// SOP §6.1 metric 정합 — auth denied counter (invalid key).
		metrics.DevhubAPIKeyAuthTotal.WithLabelValues("denied").Inc()
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"status": "unauthenticated",
			"error":  "invalid api key",
		})
		return
	}

	if h.cfg.BearerTokenVerifier == nil {
		if h.cfg.AuthDevFallback {
			c.Header("X-Devhub-Auth", "bearer_unverified")
			c.Set(httphelp.CtxKeySourceType, domain.AuditSourceSystem)
			c.Set("devhub_actor_role", "system_admin")
			c.Next()
			return
		}
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"status": "unauthenticated",
			"error":  "bearer token verifier is not configured",
		})
		return
	}

	httphelp.LogRequest(c, "[authenticateActor] Verifying token for path: %s", c.FullPath())
	actor, err := h.cfg.BearerTokenVerifier.VerifyBearerToken(c.Request.Context(), token)
	if err != nil {
		httphelp.LogRequest(c, "[authenticateActor] Token verification failed: %v", err)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"status": "unauthenticated",
			"error":  "invalid bearer token",
		})
		return
	}
	httphelp.LogRequest(c, "[authenticateActor] Token verified for login: %s", actor.Login)

	login := strings.TrimSpace(actor.Login)
	if login == "" {
		login = strings.TrimSpace(actor.Subject)
	}
	if login == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"status": "unauthenticated",
			"error":  "bearer token actor is empty",
		})
		return
	}

	c.Set("devhub_actor_login", login)
	c.Set(httphelp.CtxKeySourceType, domain.AuditSourceOIDC)
	if actor.Subject != "" {
		c.Set("devhub_actor_subject", actor.Subject)
	}

	// Dynamic Role Lookup: Instead of trusting the immutable role claim in the OIDC token,
	// we fetch the latest role from our database to support real-time permission updates.
	finalRole := actor.Role
	if h.cfg.OrganizationStore != nil && login != "" {
		user, err := h.cfg.OrganizationStore.GetUser(c.Request.Context(), login)
		switch {
		case err == nil:
			tokenRole := normalizeSystemRoleAlias(actor.Role)
			dbRole := normalizeSystemRoleAlias(string(user.Role))
			if !isGenericKeycloakRole(actor.Role) && tokenRole != "" && dbRole != "" && tokenRole != dbRole {
				c.Set("devhub_role_sync_required", true)
				c.Set("devhub_role_sync_token_role", tokenRole)
				c.Set("devhub_role_sync_db_role", dbRole)
				h.recordAuditBestEffort(c, "auth.role_sync_required", "user", login, map[string]any{
					"token_role": tokenRole,
					"db_role":    dbRole,
					"reason":     "role_drift_detected",
				})
				httphelp.LogRequest(c, "[authenticateActor] role drift detected for %q: token=%q db=%q", login, tokenRole, dbRole)
			}
			finalRole = string(user.Role)
			// sprint -t (PR #188): 자동 idp_subject sync — sprint -j codex review #9 #2 backend
			// 확장 carve 4건 중 4번째. user row 가 있고 idp_subject 가 비어있으면 actor.Subject
			// (Keycloak sub claim) 로 lazy backfill. ADR-0019 §4.3 데이터 모델 정합.
			// best-effort — 실패 시 log only (다음 요청 재시도 idempotent).
			if user.IdPSubject == "" && strings.TrimSpace(actor.Subject) != "" {
				if setErr := h.cfg.OrganizationStore.SetIdPSubject(c.Request.Context(), login, actor.Subject); setErr != nil {
					httphelp.LogRequest(c, "[authenticateActor] SetIdPSubject %q -> %q failed: %v (best-effort, next request retry)", login, actor.Subject, setErr)
				} else {
					httphelp.LogRequest(c, "[authenticateActor] idp_subject lazy backfill for %q -> %q", login, actor.Subject)
				}
			}
			// RM-ONBOARD-01 (ADR-0021 §3.3) — onboarding_required flag 정합.
			// admin pre-seeded 사용자 (row 존재 + onboarding_completed_at NULL)
			// 도 미완료로 취급 — 첫 로그인 시 onboarding 화면 강제 진입. 2026-05-21
			// lazy 폐기 sprint (issue #284) 이후 본 동작이 unconditional.
			// OnboardingGateEnabled flag 는 onboardingGate middleware 의 차단
			// 동작에만 영향 (rollback path).
			if user.OnboardingCompletedAt == nil {
				c.Set("devhub_onboarding_required", true)
			}

			// 6-P3: org unit scope 주입 — org_head / team_manager 용.
			// PostgresStore 에서만 지원; memory fake 등은 type assertion miss → skip.
			if orgStore, ok := h.cfg.OrganizationStore.(interface {
				ListOrgUnitIDsByLeader(ctx context.Context, leaderUserID string) ([]string, error)
				GetOrgUnitSubtreeIDs(ctx context.Context, unitID string) ([]string, error)
			}); ok {
				// org_head: leader_user_id 로 관리 org unit 조회 → subtree 확장
				leaderIDs, ldErr := orgStore.ListOrgUnitIDsByLeader(c.Request.Context(), login)
				if ldErr != nil {
					log.Printf("[authenticateActor] ListOrgUnitIDsByLeader(%q) failed: %v", login, ldErr)
				}
				if ldErr == nil && len(leaderIDs) > 0 {
					var allIDs []string
					for _, uid := range leaderIDs {
						subIDs, subErr := orgStore.GetOrgUnitSubtreeIDs(c.Request.Context(), uid)
						if subErr != nil {
							log.Printf("[authenticateActor] GetOrgUnitSubtreeIDs(%q) for leader unit %q failed: %v", login, uid, subErr)
						}
						if subErr == nil {
							allIDs = append(allIDs, subIDs...)
						}
					}
					if len(allIDs) > 0 {
						c.Set("devhub_actor_org_unit_ids", allIDs)
					}
				}
				// team_manager: primary_unit_id 기준 subtree 조회
				if user.Role == domain.AppRoleTeamManager && user.PrimaryUnitID != "" {
					subIDs, subErr := orgStore.GetOrgUnitSubtreeIDs(c.Request.Context(), user.PrimaryUnitID)
					if subErr != nil {
						log.Printf("[authenticateActor] GetOrgUnitSubtreeIDs(%q) for primary unit %q failed: %v", login, user.PrimaryUnitID, subErr)
					}
					if subErr == nil && len(subIDs) > 0 {
						c.Set("devhub_actor_primary_unit_ids", subIDs)
					}
				}
			}
		case errors.Is(err, store.ErrNotFound):
			// RM-ONBOARD-01 (ADR-0021 §3.3) — 2026-05-21 lazy 폐기 sprint
			// (issue #284) 이후 unconditional token-only actor 처리.
			// AuthenticatedActor 의 Email/DisplayName 은 keycloak_verifier.go 의
			// extractDisplayName + email claim 처리 가 token 에서 추출.
			// onboardingGate middleware (별도 layer) 가 allowlist 외 endpoint
			// 호출 시 403 onboarding_required 차단.
			if actor.Email != "" {
				c.Set("devhub_actor_email", actor.Email)
			}
			if actor.DisplayName != "" {
				c.Set("devhub_actor_display_name", actor.DisplayName)
			}
			c.Set("devhub_onboarding_required", true)
			httphelp.LogRequest(c, "[authenticateActor] %q DB miss = token-only actor (ADR-0021 §3.3, lazy 폐기 후 unconditional)", login)
		default:
			// Schema drift or store outage. Without this surface, a
			// missing migration (e.g. 000030 rename idp_subject)
			// silently routes every actor to actor.Role's default —
			// that masked the e2e regression where bob/charlie landed
			// on /developer until we found the SQL error by accident.
			httphelp.LogRequest(c, "[authenticateActor] GetUser %q failed: %v; falling back to token role claim %q", login, err, finalRole)
		}
	}

	if finalRole != "" {
		c.Set("devhub_actor_role", normalizeSystemRoleAlias(finalRole))
	}
	c.Next()
}

func bearerToken(header string) (string, bool) {
	scheme, token, ok := strings.Cut(header, " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") {
		return "", false
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return "", false
	}
	return token, true
}

// looksLikeJWT — header.payload.signature 형태 (각 섹션 base64url, 두 개의
// dot 으로 구분) 인지 검사. ADR-0029 API key 분기에서 JWT 와 정적 키를
// 구분하기 위함. 정확성보다 단순성 우선 — base64url charset ([A-Za-z0-9_-])
// 일치만 확인. false negative (진짜 JWT 를 API key 로 잘못 분류) 가 발생해도
// APIKey 미설정 시 분기 자체가 skip 되므로 위험 없음.
func looksLikeJWT(token string) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false
	}
	for _, p := range parts {
		if p == "" {
			return false
		}
		for _, r := range p {
			switch {
			case r >= 'A' && r <= 'Z':
			case r >= 'a' && r <= 'z':
			case r >= '0' && r <= '9':
			case r == '-' || r == '_':
			default:
				return false
			}
		}
	}
	return true
}

// subtleEqual — constant-time string comparison. API key timing attack 완화.
func subtleEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var diff byte
	for i := 0; i < len(a); i++ {
		diff |= a[i] ^ b[i]
	}
	return diff == 0
}
