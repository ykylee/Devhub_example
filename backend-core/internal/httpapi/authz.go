package httpapi

import (
	"fmt"
	"net/http"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/gin-gonic/gin"
)

// requireMinRole returns a middleware that lets the request through only when the authenticated actor's role meets or exceeds min. Dev-fallback requests (DEVHUB_AUTH_DEV_FALLBACK=1) bypass the check so local development without a real verifier can still hit protected routes.
func (h Handler) requireMinRole(min domain.AppRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		if devFallbackEnabled(c) {
			c.Next()
			return
		}

		actorValue, _ := c.Get("devhub_actor_role")
		actorRole, _ := actorValue.(string)
		if !roleMeetsMin(actorRole, min) {
			h.recordAuditBestEffort(c, "auth.role_denied", "route", c.FullPath(), map[string]any{
				"required_role": string(min),
				"actor_role":    actorRole,
			})
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"status": "forbidden",
				"error":  fmt.Sprintf("role %q does not meet required minimum %q", actorRole, string(min)),
			})
			return
		}
		c.Next()
	}
}

func roleMeetsMin(actor string, min domain.AppRole) bool {
	return roleRank(actor) >= roleRank(string(min))
}

func roleRank(role string) int {
	switch role {
	case string(domain.AppRoleSystemAdmin):
		return 30
	case string(domain.AppRoleManager):
		return 20
	case string(domain.AppRoleDeveloper):
		return 10
	default:
		return 0
	}
}
