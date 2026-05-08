package httpapi

import (
	"errors"
	"net/http"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

type meActorResponse struct {
	Login   string `json:"login"`
	Subject string `json:"subject,omitempty"`
	Source  string `json:"source"`
}

type meResponse struct {
	User                 appUserResponse   `json:"user"`
	Actor                meActorResponse   `json:"actor"`
	AllowedRoles         []string          `json:"allowed_roles"`
	EffectivePermissions map[string]string `json:"effective_permissions"`
}

func (h Handler) getMe(c *gin.Context) {
	if h.cfg.OrganizationStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "organization store is not configured",
		})
		return
	}

	identity, ok := authenticatedActorIdentity(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "unauthenticated",
			"error":  "authenticated actor is required",
		})
		return
	}

	user, err := h.cfg.OrganizationStore.GetUser(c.Request.Context(), identity.LookupID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.JSON(http.StatusForbidden, gin.H{"status": "forbidden", "error": "actor is not mapped to an active DevHub user"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}
	if user.Status != domain.UserStatusActive {
		c.JSON(http.StatusForbidden, gin.H{"status": "forbidden", "error": "actor user is not active"})
		return
	}

	policy := domain.DefaultRBACPolicy()
	if h.cfg.RBACPolicyStore != nil {
		stored, err := h.cfg.RBACPolicyStore.GetActiveRBACPolicy(c.Request.Context())
		if err != nil && !errors.Is(err, store.ErrNotFound) {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
			return
		}
		if err == nil {
			policy = stored
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data": meResponse{
			User: appUserFromDomain(user),
			Actor: meActorResponse{
				Login:   identity.Login,
				Subject: identity.Subject,
				Source:  identity.Source,
			},
			AllowedRoles:         allowedRolesFor(user.Role),
			EffectivePermissions: effectivePermissionsFor(policy, user.Role),
		},
	})
}

func effectivePermissionsFor(policy domain.RBACPolicy, role domain.AppRole) map[string]string {
	out := map[string]string{}
	for resource, permission := range policy.Matrix[string(role)] {
		out[resource] = string(permission)
	}
	return out
}

func allowedRolesFor(role domain.AppRole) []string {
	switch role {
	case domain.AppRoleSystemAdmin:
		return []string{string(domain.AppRoleDeveloper), string(domain.AppRoleManager), string(domain.AppRoleSystemAdmin)}
	case domain.AppRoleManager:
		return []string{string(domain.AppRoleDeveloper), string(domain.AppRoleManager)}
	default:
		return []string{string(domain.AppRoleDeveloper)}
	}
}
