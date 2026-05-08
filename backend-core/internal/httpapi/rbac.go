package httpapi

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

// writeRBACServerError logs the underlying error with operation context and returns a
// generic 500 (SEC-5 style) so DB internals never leak.
//
// TODO: replace with the package-wide writeServerError helper from M1 PR-A (#20)
// once that PR merges to main and the G-chain rebases on top.
func writeRBACServerError(c *gin.Context, err error, op string) {
	log.Printf("server error: op=%s err=%v", op, err)
	c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": "internal error"})
}

// RBACStore is the subset of *store.PostgresStore that the RBAC handlers depend on. The
// concrete type lives in internal/store; tests inject a fake implementation.
type RBACStore interface {
	ListRBACRoles(ctx context.Context) ([]domain.RBACRole, error)
	GetRBACRole(ctx context.Context, roleID string) (domain.RBACRole, error)
	CreateRBACRole(ctx context.Context, role domain.RBACRole) (domain.RBACRole, error)
	UpdateRBACRolePermissions(ctx context.Context, roleID string, perms domain.PermissionMatrix) (domain.RBACRole, error)
	UpdateRBACRoleMetadata(ctx context.Context, roleID, name, description string) (domain.RBACRole, error)
	DeleteRBACRole(ctx context.Context, roleID string) error
	GetSubjectRoles(ctx context.Context, userID string) ([]string, error)
	SetSubjectRole(ctx context.Context, userID, roleID string) error
}

const (
	rbacPolicyVersion       = "2026-05-08.adr-0002.v1"
	rbacAuditActionUpdated  = "rbac.policy.updated"
	rbacAuditActionAssigned = "rbac.role.assigned"
)

type rbacRoleWire struct {
	ID          string                  `json:"id"`
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	System      bool                    `json:"system"`
	Permissions domain.PermissionMatrix `json:"permissions"`
}

type rbacUpdateRoleWire struct {
	ID          string                  `json:"id"`
	Name        *string                 `json:"name,omitempty"`
	Description *string                 `json:"description,omitempty"`
	Permissions domain.PermissionMatrix `json:"permissions,omitempty"`
}

type rbacUpdatePoliciesRequest struct {
	Roles []rbacUpdateRoleWire `json:"roles"`
}

type rbacCreateRoleRequest struct {
	ID          string                  `json:"id"`
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	Permissions domain.PermissionMatrix `json:"permissions"`
}

type rbacSubjectRolesRequest struct {
	Roles []string `json:"roles"`
}

func wireFromRBACRole(role domain.RBACRole) rbacRoleWire {
	perms := role.Permissions
	if perms == nil {
		perms = domain.PermissionMatrix{}
	}
	// Always emit all 5 resources so clients do not have to backfill.
	out := make(domain.PermissionMatrix, len(domain.AllResources()))
	for _, r := range domain.AllResources() {
		out[r] = perms[r]
	}
	return rbacRoleWire{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		System:      role.System,
		Permissions: out,
	}
}

// getRBACPolicyLegacyGone replaces the M0 GET /api/v1/rbac/policy (singular) endpoint.
// Per docs/backend_api_contract.md section 6 deprecation note and ADR-0002, the legacy
// 1-dimensional response is retired; clients must move to GET /api/v1/rbac/policies.
func (h Handler) getRBACPolicyLegacyGone(c *gin.Context) {
	c.JSON(http.StatusGone, gin.H{
		"status": "gone",
		"error":  "GET /api/v1/rbac/policy was retired by ADR-0002; use GET /api/v1/rbac/policies",
		"meta": gin.H{
			"replacement": "/api/v1/rbac/policies",
			"adr":         "0002",
		},
	})
}

// listRBACPolicies serves docs/backend_api_contract.md section 12.2.
func (h Handler) listRBACPolicies(c *gin.Context) {
	if h.cfg.RBACStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "error": "rbac store is not configured"})
		return
	}

	roles, err := h.cfg.RBACStore.ListRBACRoles(c.Request.Context())
	if err != nil {
		writeRBACServerError(c, err, "rbac.list_policies")
		return
	}

	data := make([]rbacRoleWire, 0, len(roles))
	for _, role := range roles {
		data = append(data, wireFromRBACRole(role))
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   data,
		"meta": gin.H{
			"policy_version": rbacPolicyVersion,
			"source":         "rbac_policies_store",
			"editable":       true,
			"system_roles":   domain.SystemRoleIDs(),
		},
	})
}

// createRBACPolicy serves section 12.4 — POST /api/v1/rbac/policies for custom roles.
func (h Handler) createRBACPolicy(c *gin.Context) {
	if h.cfg.RBACStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "error": "rbac store is not configured"})
		return
	}

	var req rbacCreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "invalid json body"})
		return
	}
	req.ID = strings.TrimSpace(req.ID)
	req.Name = strings.TrimSpace(req.Name)
	if req.ID == "" || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "id and name are required"})
		return
	}
	if domain.IsSystemRole(req.ID) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": "rejected", "error": "system role ids are reserved", "code": "system_role_reserved"})
		return
	}
	if err := domain.ValidateRoleID(req.ID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": err.Error(), "code": "invalid_role_id"})
		return
	}

	created, err := h.cfg.RBACStore.CreateRBACRole(c.Request.Context(), domain.RBACRole{
		ID:          req.ID,
		Name:        req.Name,
		Description: req.Description,
		Permissions: req.Permissions,
	})
	switch {
	case errors.Is(err, store.ErrConflict):
		c.JSON(http.StatusConflict, gin.H{"status": "conflict", "error": "role id already exists", "code": "role_id_conflict"})
		return
	case errors.Is(err, store.ErrAuditInvariantViolation):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": "rejected", "error": err.Error(), "code": "audit_invariant_violation"})
		return
	case err != nil:
		writeRBACServerError(c, err, "rbac.create_policy")
		return
	}

	if h.cfg.PermissionCache != nil {
		h.cfg.PermissionCache.Invalidate()
	}

	auditLog := h.recordAuditBestEffort(c, rbacAuditActionUpdated, "rbac_role", created.ID, map[string]any{
		"change_type": "created",
		"after":       wireFromRBACRole(created),
	})

	response := gin.H{
		"status": "created",
		"data":   wireFromRBACRole(created),
	}
	addAuditMeta(response, auditLog)
	c.JSON(http.StatusCreated, response)
}

// updateRBACPolicies serves section 12.3 — PUT /api/v1/rbac/policies (bulk role update).
// Only the 'permissions' field of each role entry is honored for system roles. Custom roles
// can also have name/description updated in the same call.
func (h Handler) updateRBACPolicies(c *gin.Context) {
	if h.cfg.RBACStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "error": "rbac store is not configured"})
		return
	}

	var req rbacUpdatePoliciesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "invalid json body"})
		return
	}
	if len(req.Roles) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "roles array must contain at least one entry"})
		return
	}

	ctx := c.Request.Context()
	auditEntries := make([]map[string]any, 0, len(req.Roles))

	for _, entry := range req.Roles {
		entry.ID = strings.TrimSpace(entry.ID)
		if entry.ID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "role id is required for each entry"})
			return
		}

		before, err := h.cfg.RBACStore.GetRBACRole(ctx, entry.ID)
		if errors.Is(err, store.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": fmt.Sprintf("role %q not found", entry.ID)})
			return
		}
		if err != nil {
			writeRBACServerError(c, err, "rbac.update_policies.lookup")
			return
		}

		if entry.Permissions != nil {
			updated, err := h.cfg.RBACStore.UpdateRBACRolePermissions(ctx, entry.ID, entry.Permissions)
			switch {
			case errors.Is(err, store.ErrNotFound):
				c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": fmt.Sprintf("role %q not found", entry.ID)})
				return
			case errors.Is(err, store.ErrAuditInvariantViolation):
				c.JSON(http.StatusUnprocessableEntity, gin.H{"status": "rejected", "error": err.Error(), "code": "audit_invariant_violation"})
				return
			case err != nil:
				writeRBACServerError(c, err, "rbac.update_policies.permissions")
				return
			}
			auditEntries = append(auditEntries, map[string]any{
				"role_id":     entry.ID,
				"change_type": "permissions_updated",
				"before":      wireFromRBACRole(before),
				"after":       wireFromRBACRole(updated),
			})
			before = updated
		}

		if entry.Name != nil || entry.Description != nil {
			if before.System {
				c.JSON(http.StatusUnprocessableEntity, gin.H{"status": "rejected", "error": "system role name/description cannot change", "code": "system_role_immutable"})
				return
			}
			name := before.Name
			if entry.Name != nil {
				name = strings.TrimSpace(*entry.Name)
			}
			desc := before.Description
			if entry.Description != nil {
				desc = *entry.Description
			}
			updated, err := h.cfg.RBACStore.UpdateRBACRoleMetadata(ctx, entry.ID, name, desc)
			switch {
			case errors.Is(err, store.ErrSystemRoleImmutable):
				c.JSON(http.StatusUnprocessableEntity, gin.H{"status": "rejected", "error": err.Error(), "code": "system_role_immutable"})
				return
			case errors.Is(err, store.ErrNotFound):
				c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": fmt.Sprintf("role %q not found", entry.ID)})
				return
			case err != nil:
				writeRBACServerError(c, err, "rbac.update_policies.metadata")
				return
			}
			auditEntries = append(auditEntries, map[string]any{
				"role_id":     entry.ID,
				"change_type": "metadata_updated",
				"before":      wireFromRBACRole(before),
				"after":       wireFromRBACRole(updated),
			})
		}
	}

	if h.cfg.PermissionCache != nil {
		h.cfg.PermissionCache.Invalidate()
	}

	auditLog := h.recordAuditBestEffort(c, rbacAuditActionUpdated, "rbac_role", "policies", map[string]any{
		"change_type": "bulk",
		"changes":     auditEntries,
	})

	roles, err := h.cfg.RBACStore.ListRBACRoles(ctx)
	if err != nil {
		writeRBACServerError(c, err, "rbac.update_policies.list_after")
		return
	}
	data := make([]rbacRoleWire, 0, len(roles))
	for _, role := range roles {
		data = append(data, wireFromRBACRole(role))
	}
	response := gin.H{
		"status": "ok",
		"data":   data,
		"meta": gin.H{
			"policy_version": rbacPolicyVersion,
			"source":         "rbac_policies_store",
			"editable":       true,
			"system_roles":   domain.SystemRoleIDs(),
		},
	}
	addAuditMeta(response, auditLog)
	c.JSON(http.StatusOK, response)
}

// deleteRBACPolicy serves section 12.5 — DELETE /api/v1/rbac/policies/:role_id.
func (h Handler) deleteRBACPolicy(c *gin.Context) {
	if h.cfg.RBACStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "error": "rbac store is not configured"})
		return
	}

	roleID := strings.TrimSpace(c.Param("role_id"))
	if roleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "role_id is required"})
		return
	}

	ctx := c.Request.Context()
	before, err := h.cfg.RBACStore.GetRBACRole(ctx, roleID)
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": fmt.Sprintf("role %q not found", roleID)})
		return
	}
	if err != nil {
		writeRBACServerError(c, err, "rbac.delete_policy.lookup")
		return
	}

	if err := h.cfg.RBACStore.DeleteRBACRole(ctx, roleID); err != nil {
		switch {
		case errors.Is(err, store.ErrSystemRoleImmutable):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"status": "rejected", "error": "system roles cannot be deleted", "code": "system_role_not_deletable"})
		case errors.Is(err, store.ErrRoleInUse):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"status": "rejected", "error": err.Error(), "code": "role_in_use"})
		case errors.Is(err, store.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": fmt.Sprintf("role %q not found", roleID)})
		default:
			writeRBACServerError(c, err, "rbac.delete_policy")
		}
		return
	}

	if h.cfg.PermissionCache != nil {
		h.cfg.PermissionCache.Invalidate()
	}

	auditLog := h.recordAuditBestEffort(c, rbacAuditActionUpdated, "rbac_role", roleID, map[string]any{
		"change_type": "deleted",
		"before":      wireFromRBACRole(before),
	})

	response := gin.H{
		"status": "deleted",
		"data":   gin.H{"role_id": roleID},
	}
	addAuditMeta(response, auditLog)
	c.JSON(http.StatusOK, response)
}

// getSubjectRoles serves section 12.6 — GET /api/v1/rbac/subjects/:subject_id/roles.
func (h Handler) getSubjectRoles(c *gin.Context) {
	if h.cfg.RBACStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "error": "rbac store is not configured"})
		return
	}

	subjectID := strings.TrimSpace(c.Param("subject_id"))
	if subjectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "subject_id is required"})
		return
	}

	roles, err := h.cfg.RBACStore.GetSubjectRoles(c.Request.Context(), subjectID)
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": fmt.Sprintf("subject %q not found", subjectID)})
		return
	}
	if err != nil {
		writeRBACServerError(c, err, "rbac.get_subject_roles")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   roles,
		"meta": gin.H{
			"subject_id":       subjectID,
			"single_role_mode": true,
		},
	})
}

// setSubjectRoles serves section 12.7 — PUT /api/v1/rbac/subjects/:subject_id/roles.
// Single-role mode: roles array must have exactly one entry.
func (h Handler) setSubjectRoles(c *gin.Context) {
	if h.cfg.RBACStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "error": "rbac store is not configured"})
		return
	}

	subjectID := strings.TrimSpace(c.Param("subject_id"))
	if subjectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "subject_id is required"})
		return
	}

	var req rbacSubjectRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "invalid json body"})
		return
	}
	if len(req.Roles) != 1 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": "rejected", "error": "exactly one role required (single-role mode)", "code": "single_role_required"})
		return
	}
	roleID := strings.TrimSpace(req.Roles[0])
	if roleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "role id cannot be empty"})
		return
	}

	ctx := c.Request.Context()
	before, err := h.cfg.RBACStore.GetSubjectRoles(ctx, subjectID)
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": fmt.Sprintf("subject %q not found", subjectID)})
		return
	}
	if err != nil {
		writeRBACServerError(c, err, "rbac.set_subject_roles.lookup")
		return
	}

	if err := h.cfg.RBACStore.SetSubjectRole(ctx, subjectID, roleID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": fmt.Sprintf("role %q not found or subject missing", roleID)})
			return
		}
		writeRBACServerError(c, err, "rbac.set_subject_roles.update")
		return
	}

	auditLog := h.recordAuditBestEffort(c, rbacAuditActionAssigned, "user", subjectID, map[string]any{
		"before": before,
		"after":  []string{roleID},
	})

	response := gin.H{
		"status": "ok",
		"data":   []string{roleID},
		"meta": gin.H{
			"subject_id":       subjectID,
			"single_role_mode": true,
		},
	}
	addAuditMeta(response, auditLog)
	c.JSON(http.StatusOK, response)
}
