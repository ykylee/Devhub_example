package httpapi

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/devhub/backend-core/internal/domain"
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

func (h Handler) authenticateActor(c *gin.Context) {
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
			c.Set(ctxKeySourceType, domain.AuditSourceWebhook)
		case "/api/v1/integration/providers/:provider_id/webhook":
			c.Set(ctxKeySourceType, domain.AuditSourceWebhook)
		default:
			c.Set(ctxKeySourceType, domain.AuditSourceSystem)
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
				entry, ok, err := h.cfg.RealtimeTickets.consume(c.Request.Context(), raw)
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
					c.Set("devhub_actor_login", entry.actorLogin)
					if entry.actorRole != "" {
						c.Set("devhub_actor_role", entry.actorRole)
					}
					c.Set(ctxKeySourceType, entry.sourceType)
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
			c.Set(ctxKeySourceType, domain.AuditSourceSystem)
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

	if h.cfg.BearerTokenVerifier == nil {
		if h.cfg.AuthDevFallback {
			c.Header("X-Devhub-Auth", "bearer_unverified")
			c.Set(ctxKeySourceType, domain.AuditSourceSystem)
			c.Next()
			return
		}
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"status": "unauthenticated",
			"error":  "bearer token verifier is not configured",
		})
		return
	}

	logRequest(c, "[authenticateActor] Verifying token for path: %s", c.FullPath())
	actor, err := h.cfg.BearerTokenVerifier.VerifyBearerToken(c.Request.Context(), token)
	if err != nil {
		logRequest(c, "[authenticateActor] Token verification failed: %v", err)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"status": "unauthenticated",
			"error":  "invalid bearer token",
		})
		return
	}
	logRequest(c, "[authenticateActor] Token verified for login: %s", actor.Login)

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
	c.Set(ctxKeySourceType, domain.AuditSourceOIDC)
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
			finalRole = string(user.Role)
			// sprint -t (PR #188): 자동 idp_subject sync — sprint -j codex review #9 #2 backend
			// 확장 carve 4건 중 4번째. user row 가 있고 idp_subject 가 비어있으면 actor.Subject
			// (Keycloak sub claim) 로 lazy backfill. ADR-0019 §4.3 데이터 모델 정합.
			// best-effort — 실패 시 log only (다음 요청 재시도 idempotent).
			if user.IdPSubject == "" && strings.TrimSpace(actor.Subject) != "" {
				if setErr := h.cfg.OrganizationStore.SetIdPSubject(c.Request.Context(), login, actor.Subject); setErr != nil {
					logRequest(c, "[authenticateActor] SetIdPSubject %q -> %q failed: %v (best-effort, next request retry)", login, actor.Subject, setErr)
				} else {
					logRequest(c, "[authenticateActor] idp_subject lazy backfill for %q -> %q", login, actor.Subject)
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
			logRequest(c, "[authenticateActor] %q DB miss = token-only actor (ADR-0021 §3.3, lazy 폐기 후 unconditional)", login)
		default:
			// Schema drift or store outage. Without this surface, a
			// missing migration (e.g. 000030 rename idp_subject)
			// silently routes every actor to actor.Role's default —
			// that masked the e2e regression where bob/charlie landed
			// on /developer until we found the SQL error by accident.
			logRequest(c, "[authenticateActor] GetUser %q failed: %v; falling back to token role claim %q", login, err, finalRole)
		}
	}

	if finalRole != "" {
		c.Set("devhub_actor_role", finalRole)
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
