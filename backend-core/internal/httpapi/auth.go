package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strings"

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

// publicAPIPaths lists /api/v1 routes that pass through authenticateActor without an Authorization header. Webhook endpoints validate their own HMAC signature, so a Bearer token would be redundant.
var publicAPIPaths = map[string]bool{
	"/api/v1/integrations/gitea/webhooks": true,
}

func (h Handler) authenticateActor(c *gin.Context) {
	c.Set("devhub_auth_dev_fallback", h.cfg.AuthDevFallback)

	if publicAPIPaths[c.FullPath()] {
		c.Next()
		return
	}

	header := strings.TrimSpace(c.GetHeader("Authorization"))
	if header == "" {
		if h.cfg.AuthDevFallback {
			c.Header("X-Devhub-Auth", "dev_fallback_no_header")
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
			c.Next()
			return
		}
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"status": "unauthenticated",
			"error":  "bearer token verifier is not configured",
		})
		return
	}

	actor, err := h.cfg.BearerTokenVerifier.VerifyBearerToken(c.Request.Context(), token)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"status": "unauthenticated",
			"error":  "invalid bearer token",
		})
		return
	}

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
	if actor.Subject != "" {
		c.Set("devhub_actor_subject", actor.Subject)
	}
	if actor.Role != "" {
		c.Set("devhub_actor_role", actor.Role)
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
