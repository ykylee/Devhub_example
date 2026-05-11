package httpapi

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/gin-gonic/gin"
)

var ErrInvalidBearerToken = errors.New("invalid bearer token")

type AuthenticatedActor struct {
	Login   string
	Subject string
	Role    string
}

type BearerTokenVerifier interface {
	VerifyBearerToken(context.Context, string) (AuthenticatedActor, error)
}

// publicAPIPaths lists /api/v1 routes that pass through authenticateActor without an Authorization header. Webhook endpoints validate their own HMAC signature, so a Bearer token would be redundant. The auth proxy endpoints (login/consent/logout) are called *before* the user has a token, so they cannot require one either; they protect themselves with Hydra's challenge tokens (single-use, lifespan-bound).
var publicAPIPaths = map[string]bool{
	"/api/v1/integrations/gitea/webhooks": true,
	"/api/v1/auth/login":                  true,
	"/api/v1/auth/logout":                 true,
	"/api/v1/auth/token":                  true,
	"/api/v1/auth/signup":                 true,
	"/api/v1/auth/consent":                true,
}

func (h Handler) authenticateActor(c *gin.Context) {
	c.Set("devhub_auth_dev_fallback", h.cfg.AuthDevFallback)

	if publicAPIPaths[c.FullPath()] {
		// Webhook bypass paths run without a Bearer token but still produce
		// audit rows (signature-verified Gitea webhooks). Tag the source so
		// downstream recordAudit picks the right enum (T-M1-04, DEC-2=A).
		// Other public paths (auth proxy endpoints) issue audits via the
		// dedicated handlers and override the source type as needed.
		if c.FullPath() == "/api/v1/integrations/gitea/webhooks" {
			c.Set(ctxKeySourceType, domain.AuditSourceWebhook)
		} else {
			c.Set(ctxKeySourceType, domain.AuditSourceSystem)
		}
		c.Next()
		return
	}

	header := strings.TrimSpace(c.GetHeader("Authorization"))
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

	log.Printf("[authenticateActor] Verifying token for path: %s", c.FullPath())
	actor, err := h.cfg.BearerTokenVerifier.VerifyBearerToken(c.Request.Context(), token)
	if err != nil {
		log.Printf("[authenticateActor] Token verification failed: %v", err)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"status": "unauthenticated",
			"error":  "invalid bearer token",
		})
		return
	}
	log.Printf("[authenticateActor] Token verified for login: %s", actor.Login)

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
		if user, err := h.cfg.OrganizationStore.GetUser(c.Request.Context(), login); err == nil {
			finalRole = string(user.Role)
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
