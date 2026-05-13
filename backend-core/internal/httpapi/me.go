package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type meResponse struct {
	Login   string `json:"login"`
	Subject string `json:"subject,omitempty"`
	Role    string `json:"role,omitempty"`
	Source  string `json:"actor_source"`
}

// getMe returns the authenticated actor for the current request. Frontend
// uses this to derive the active role after a successful Kratos+Hydra login.
// Returns 401 when the request did not produce an authenticated actor (no
// Authorization header in production, or the dev fallback resolved actor to
// "system"). The legacy X-Devhub-Actor header is intentionally ignored —
// ADR-0004 finalized its removal; SEC-4 already stripped the production
// handling.
func (h Handler) getMe(c *gin.Context) {
	actor := requestActor(c)
	if actor.Login == "" || actor.Login == "system" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"status": "unauthenticated",
			"error":  "no authenticated user in request context",
		})
		return
	}

	subject := ""
	if v, ok := c.Get("devhub_actor_subject"); ok {
		if s, ok := v.(string); ok {
			subject = s
		}
	}
	role := ""
	if v, ok := c.Get("devhub_actor_role"); ok {
		if s, ok := v.(string); ok {
			role = s
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data": meResponse{
			Login:   actor.Login,
			Subject: subject,
			Role:    role,
			Source:  actor.Source,
		},
	})
}
