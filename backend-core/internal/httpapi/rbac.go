package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

type rbacRoleResponse struct {
	Role        string `json:"role"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

type rbacResourceResponse struct {
	Resource    string `json:"resource"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

type rbacPermissionResponse struct {
	Permission  string `json:"permission"`
	Label       string `json:"label"`
	Rank        int    `json:"rank"`
	Description string `json:"description"`
}

type rbacPolicyResponse struct {
	Roles       []rbacRoleResponse           `json:"roles"`
	Resources   []rbacResourceResponse       `json:"resources"`
	Permissions []rbacPermissionResponse     `json:"permissions"`
	Matrix      map[string]map[string]string `json:"matrix"`
}

type replaceRBACPolicyRequest struct {
	PolicyVersion string                       `json:"policy_version"`
	Reason        string                       `json:"reason"`
	Matrix        map[string]map[string]string `json:"matrix"`
}

func (h Handler) getRBACPolicy(c *gin.Context) {
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
		"data":   rbacPolicyFromDomain(policy),
		"meta": gin.H{
			"policy_version": policy.PolicyVersion,
			"source":         policy.Source,
			"editable":       policy.Editable,
		},
	})
}

func (h Handler) replaceRBACPolicy(c *gin.Context) {
	if h.cfg.RBACPolicyStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "rbac policy store is not configured",
		})
		return
	}
	if !h.requirePermission(c, "system_config", domain.RBACPermissionAdmin) {
		return
	}

	var req replaceRBACPolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "request body must be valid JSON"})
		return
	}
	req.PolicyVersion = strings.TrimSpace(req.PolicyVersion)
	req.Reason = strings.TrimSpace(req.Reason)
	if req.Reason == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "reason is required"})
		return
	}

	policy, err := rbacPolicyFromRequest(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": err.Error()})
		return
	}
	actor := requestActor(c)
	replaced, err := h.cfg.RBACPolicyStore.ReplaceRBACPolicy(c.Request.Context(), domain.ReplaceRBACPolicyInput{
		PolicyVersion: req.PolicyVersion,
		ActorLogin:    actor.Login,
		Reason:        req.Reason,
		Policy:        policy,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	auditLog, err := h.recordAudit(c, "rbac.policy_replaced", "rbac_policy", replaced.PolicyVersion, map[string]any{
		"policy_version": replaced.PolicyVersion,
		"reason":         req.Reason,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	response := gin.H{
		"status": "ok",
		"data":   rbacPolicyFromDomain(replaced),
		"meta": gin.H{
			"policy_version": replaced.PolicyVersion,
			"source":         replaced.Source,
			"editable":       replaced.Editable,
		},
	}
	addAuditMeta(response, auditLog)
	c.JSON(http.StatusOK, response)
}

func rbacPolicyFromDomain(policy domain.RBACPolicy) rbacPolicyResponse {
	roles := make([]rbacRoleResponse, 0, len(policy.Roles))
	for _, role := range policy.Roles {
		roles = append(roles, rbacRoleResponse{
			Role:        string(role.Role),
			Label:       role.Label,
			Description: role.Description,
		})
	}
	resources := make([]rbacResourceResponse, 0, len(policy.Resources))
	for _, resource := range policy.Resources {
		resources = append(resources, rbacResourceResponse{
			Resource:    resource.Resource,
			Label:       resource.Label,
			Description: resource.Description,
		})
	}
	permissions := make([]rbacPermissionResponse, 0, len(policy.Permissions))
	for _, permission := range policy.Permissions {
		permissions = append(permissions, rbacPermissionResponse{
			Permission:  string(permission.Permission),
			Label:       permission.Label,
			Rank:        permission.Rank,
			Description: permission.Description,
		})
	}
	matrix := make(map[string]map[string]string, len(policy.Matrix))
	for role, resources := range policy.Matrix {
		matrix[role] = map[string]string{}
		for resource, permission := range resources {
			matrix[role][resource] = string(permission)
		}
	}
	return rbacPolicyResponse{
		Roles:       roles,
		Resources:   resources,
		Permissions: permissions,
		Matrix:      matrix,
	}
}

func rbacPolicyFromRequest(req replaceRBACPolicyRequest) (domain.RBACPolicy, error) {
	base := domain.DefaultRBACPolicy()
	if req.Matrix == nil {
		return domain.RBACPolicy{}, errors.New("matrix is required")
	}

	matrix := make(map[string]map[string]domain.RBACPermission, len(base.Roles))
	for _, role := range base.Roles {
		roleKey := string(role.Role)
		inResources, ok := req.Matrix[roleKey]
		if !ok {
			return domain.RBACPolicy{}, errors.New("matrix must include role " + roleKey)
		}
		matrix[roleKey] = map[string]domain.RBACPermission{}
		for _, resource := range base.Resources {
			permission, ok := inResources[resource.Resource]
			if !ok {
				return domain.RBACPolicy{}, errors.New("matrix must include resource " + resource.Resource + " for role " + roleKey)
			}
			permission = strings.TrimSpace(permission)
			if !validRBACPermission(permission) {
				return domain.RBACPolicy{}, errors.New("permission must be one of none, read, write, admin")
			}
			matrix[roleKey][resource.Resource] = domain.RBACPermission(permission)
		}
	}
	base.PolicyVersion = req.PolicyVersion
	base.Matrix = matrix
	return base, nil
}

func validRBACPermission(permission string) bool {
	switch domain.RBACPermission(permission) {
	case domain.RBACPermissionNone, domain.RBACPermissionRead, domain.RBACPermissionWrite, domain.RBACPermissionAdmin:
		return true
	default:
		return false
	}
}
