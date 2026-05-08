package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

type requestActorIdentity struct {
	Login    string
	Subject  string
	LookupID string
	Source   string
}

func (h Handler) requirePermission(c *gin.Context, resource string, required domain.RBACPermission) bool {
	if h.cfg.RBACPolicyStore == nil {
		return true
	}
	role, ok := h.resolveActorRole(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "unauthenticated",
			"error":  "authenticated actor role is required",
		})
		return false
	}

	policy := domain.DefaultRBACPolicy()
	stored, err := h.cfg.RBACPolicyStore.GetActiveRBACPolicy(c.Request.Context())
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return false
	}
	if err == nil {
		policy = stored
	}

	actual := permissionFor(policy, role, resource)
	if !permissionAllows(actual, required) {
		c.JSON(http.StatusForbidden, gin.H{
			"status":              "forbidden",
			"error":               "permission denied",
			"actor_role":          string(role),
			"resource":            resource,
			"required_permission": string(required),
			"actual_permission":   string(actual),
		})
		return false
	}
	return true
}

func (h Handler) actorHasPermission(c *gin.Context, resource string, required domain.RBACPermission) (bool, int, gin.H) {
	if h.cfg.RBACPolicyStore == nil {
		return true, http.StatusOK, nil
	}
	role, ok := h.resolveActorRole(c)
	if !ok {
		return false, http.StatusUnauthorized, gin.H{
			"status": "unauthenticated",
			"error":  "authenticated actor role is required",
		}
	}
	policy := domain.DefaultRBACPolicy()
	stored, err := h.cfg.RBACPolicyStore.GetActiveRBACPolicy(c.Request.Context())
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		return false, http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()}
	}
	if err == nil {
		policy = stored
	}
	actual := permissionFor(policy, role, resource)
	if permissionAllows(actual, required) {
		return true, http.StatusOK, nil
	}
	return false, http.StatusForbidden, gin.H{
		"status":              "forbidden",
		"error":               "permission denied",
		"actor_role":          string(role),
		"resource":            resource,
		"required_permission": string(required),
		"actual_permission":   string(actual),
	}
}

func (h Handler) resolveActorRole(c *gin.Context) (domain.AppRole, bool) {
	if h.cfg.OrganizationStore != nil {
		identity, ok := authenticatedActorIdentity(c)
		if ok {
			user, err := h.cfg.OrganizationStore.GetUser(c.Request.Context(), identity.LookupID)
			if err == nil && user.Status == domain.UserStatusActive {
				return user.Role, true
			}
			if err == nil {
				return "", false
			}
			if !errors.Is(err, store.ErrNotFound) {
				return "", false
			}
			return "", false
		}
		if !h.cfg.AuthDevFallback {
			return "", false
		}
	}
	return requestActorRole(c)
}

func authenticatedActorIdentity(c *gin.Context) (requestActorIdentity, bool) {
	var identity requestActorIdentity
	if value, ok := c.Get("devhub_actor_login"); ok {
		if login, ok := value.(string); ok {
			identity.Login = strings.TrimSpace(login)
		}
	}
	if value, ok := c.Get("devhub_actor_subject"); ok {
		if subject, ok := value.(string); ok {
			identity.Subject = strings.TrimSpace(subject)
		}
	}
	if identity.Subject != "" {
		identity.LookupID = identity.Subject
		identity.Source = "authenticated_context"
		if identity.Login == "" {
			identity.Login = identity.Subject
		}
		return identity, true
	}
	if identity.Login != "" {
		identity.LookupID = identity.Login
		identity.Source = "authenticated_context"
		return identity, true
	}

	actor := strings.TrimSpace(c.GetHeader("X-Devhub-Actor"))
	if actor == "" {
		actor = strings.TrimSpace(c.Query("actor"))
	}
	if actor == "" {
		return requestActorIdentity{}, false
	}
	c.Header("X-Devhub-Actor-Deprecated", "true")
	c.Header("Warning", `299 - "X-Devhub-Actor is a development fallback; use authenticated session or bearer token claims"`)
	return requestActorIdentity{
		Login:    actor,
		LookupID: actor,
		Source:   "x-devhub-actor",
	}, true
}

func requestActorRole(c *gin.Context) (domain.AppRole, bool) {
	if value, ok := c.Get("devhub_actor_role"); ok {
		if role, ok := value.(string); ok {
			return normalizeAppRole(role)
		}
	}

	role := strings.TrimSpace(c.GetHeader("X-Devhub-Role"))
	if role == "" {
		role = strings.TrimSpace(c.Query("role"))
	}
	if role == "" {
		return "", false
	}
	c.Header("X-Devhub-Role-Deprecated", "true")
	c.Header("Warning", `299 - "X-Devhub-Role is a development fallback; use authenticated session or bearer token claims"`)
	return normalizeAppRole(role)
}

func normalizeAppRole(role string) (domain.AppRole, bool) {
	switch domain.AppRole(strings.TrimSpace(role)) {
	case domain.AppRoleDeveloper:
		return domain.AppRoleDeveloper, true
	case domain.AppRoleManager:
		return domain.AppRoleManager, true
	case domain.AppRoleSystemAdmin:
		return domain.AppRoleSystemAdmin, true
	default:
		return "", false
	}
}

func permissionAllows(actual, required domain.RBACPermission) bool {
	return permissionRank(actual) >= permissionRank(required)
}

func permissionFor(policy domain.RBACPolicy, role domain.AppRole, resource string) domain.RBACPermission {
	if policy.Matrix == nil || policy.Matrix[string(role)] == nil {
		return domain.RBACPermissionNone
	}
	permission := policy.Matrix[string(role)][resource]
	if permission == "" {
		return domain.RBACPermissionNone
	}
	return permission
}

func permissionRank(permission domain.RBACPermission) int {
	switch permission {
	case domain.RBACPermissionAdmin:
		return 30
	case domain.RBACPermissionWrite:
		return 20
	case domain.RBACPermissionRead:
		return 10
	default:
		return 0
	}
}
