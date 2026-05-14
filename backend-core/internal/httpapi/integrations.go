package httpapi

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

// Integration CRUD (API-58, sprint claude/work_260514-c).
// project_integrations 테이블의 scope polymorphism (application | project).

var (
	validIntegrationScopes = map[string]bool{
		"application": true, "project": true,
	}
	validIntegrationTypes = map[string]bool{
		"jira": true, "confluence": true,
	}
	validIntegrationPolicies = map[string]bool{
		"summary_only": true, "execution_system": true,
	}
)

func integrationResponse(i domain.ProjectIntegration) gin.H {
	return gin.H{
		"id":               i.ID,
		"scope":            string(i.Scope),
		"project_id":       i.ProjectID,
		"application_id":   i.ApplicationID,
		"integration_type": string(i.IntegrationType),
		"external_key":     i.ExternalKey,
		"url":              i.URL,
		"policy":           string(i.Policy),
		"created_at":       i.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at":       i.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

// GET /api/v1/integrations
func (h *Handler) listIntegrations(c *gin.Context) {
	storeI, ok := h.applicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	opts := store.IntegrationListOptions{
		Scope:           domain.IntegrationScope(c.Query("scope")),
		ApplicationID:   c.Query("application_id"),
		ProjectID:       c.Query("project_id"),
		IntegrationType: domain.IntegrationType(c.Query("integration_type")),
	}
	if string(opts.Scope) != "" && !validIntegrationScopes[string(opts.Scope)] {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "scope must be application or project"})
		return
	}
	if string(opts.IntegrationType) != "" && !validIntegrationTypes[string(opts.IntegrationType)] {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "integration_type must be jira or confluence"})
		return
	}
	if v := c.Query("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "limit must be 1..100"})
			return
		}
		opts.Limit = n
	}
	if v := c.Query("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "offset must be >= 0"})
			return
		}
		opts.Offset = n
	}
	integrations, total, err := storeI.ListIntegrations(c.Request.Context(), opts)
	if err != nil {
		writeServerError(c, err, "integrations.list")
		return
	}
	resp := make([]gin.H, 0, len(integrations))
	for _, i := range integrations {
		resp = append(resp, integrationResponse(i))
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   resp,
		"meta":   gin.H{"total": total},
	})
}

type createIntegrationRequest struct {
	Scope           string `json:"scope"`
	ApplicationID   string `json:"application_id"`
	ProjectID       string `json:"project_id"`
	IntegrationType string `json:"integration_type"`
	ExternalKey     string `json:"external_key"`
	URL             string `json:"url"`
	Policy          string `json:"policy"`
}

// POST /api/v1/integrations
func (h *Handler) createIntegration(c *gin.Context) {
	storeI, ok := h.applicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	var req createIntegrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": err.Error()})
		return
	}
	if !validIntegrationScopes[req.Scope] {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "scope must be application or project"})
		return
	}
	if !validIntegrationTypes[req.IntegrationType] {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "integration_type must be jira or confluence"})
		return
	}
	if !validIntegrationPolicies[req.Policy] {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "policy must be summary_only or execution_system"})
		return
	}
	if strings.TrimSpace(req.ExternalKey) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "external_key is required"})
		return
	}
	if strings.TrimSpace(req.URL) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "url is required"})
		return
	}
	// scope-target 정합 (DB CHECK 와 일치)
	if req.Scope == "application" {
		if strings.TrimSpace(req.ApplicationID) == "" {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"status": "rejected", "error": "scope=application requires application_id", "code": "scope_target_mismatch"})
			return
		}
		req.ProjectID = ""
	} else { // project
		if strings.TrimSpace(req.ProjectID) == "" {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"status": "rejected", "error": "scope=project requires project_id", "code": "scope_target_mismatch"})
			return
		}
		req.ApplicationID = ""
	}
	integration := domain.ProjectIntegration{
		Scope:           domain.IntegrationScope(req.Scope),
		ApplicationID:   req.ApplicationID,
		ProjectID:       req.ProjectID,
		IntegrationType: domain.IntegrationType(req.IntegrationType),
		ExternalKey:     req.ExternalKey,
		URL:             req.URL,
		Policy:          domain.IntegrationPolicy(req.Policy),
	}
	created, err := storeI.CreateIntegration(c.Request.Context(), integration)
	if errors.Is(err, store.ErrConflict) {
		c.JSON(http.StatusConflict, gin.H{
			"status": "conflict",
			"error":  "integration already exists or referenced application/project not found",
			"code":   "integration_conflict",
		})
		return
	}
	if err != nil {
		writeServerError(c, err, "integrations.create")
		return
	}
	target := created.ApplicationID
	if created.Scope == domain.IntegrationScopeProject {
		target = created.ProjectID
	}
	h.recordAuditBestEffort(c, "integration.created", "integration", created.ID, map[string]any{
		"scope":            string(created.Scope),
		"target_id":        target,
		"integration_type": string(created.IntegrationType),
		"policy":           string(created.Policy),
	})
	c.JSON(http.StatusCreated, gin.H{
		"status": "ok",
		"data":   integrationResponse(created),
	})
}

type updateIntegrationRequest struct {
	ExternalKey *string `json:"external_key"`
	URL         *string `json:"url"`
	Policy      *string `json:"policy"`
}

// PATCH /api/v1/integrations/:integration_id
func (h *Handler) updateIntegration(c *gin.Context) {
	storeI, ok := h.applicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	id := c.Param("integration_id")
	var req updateIntegrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": err.Error()})
		return
	}
	current, err := storeI.GetIntegration(c.Request.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "integration not found"})
		return
	}
	if err != nil {
		writeServerError(c, err, "integrations.update.lookup")
		return
	}
	updated := current
	if req.ExternalKey != nil {
		if strings.TrimSpace(*req.ExternalKey) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "external_key cannot be empty"})
			return
		}
		updated.ExternalKey = *req.ExternalKey
	}
	if req.URL != nil {
		if strings.TrimSpace(*req.URL) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "url cannot be empty"})
			return
		}
		updated.URL = *req.URL
	}
	if req.Policy != nil {
		if !validIntegrationPolicies[*req.Policy] {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "policy must be summary_only or execution_system"})
			return
		}
		updated.Policy = domain.IntegrationPolicy(*req.Policy)
	}
	result, err := storeI.UpdateIntegration(c.Request.Context(), updated)
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "integration not found"})
		return
	}
	if errors.Is(err, store.ErrConflict) {
		c.JSON(http.StatusConflict, gin.H{
			"status": "conflict",
			"error":  "integration with the same (scope target, type, external_key) already exists",
			"code":   "integration_conflict",
		})
		return
	}
	if err != nil {
		writeServerError(c, err, "integrations.update")
		return
	}
	h.recordAuditBestEffort(c, "integration.updated", "integration", id, map[string]any{
		"policy": string(result.Policy),
	})
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   integrationResponse(result),
	})
}

// DELETE /api/v1/integrations/:integration_id
func (h *Handler) deleteIntegration(c *gin.Context) {
	storeI, ok := h.applicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	id := c.Param("integration_id")
	if err := storeI.DeleteIntegration(c.Request.Context(), id); errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "integration not found"})
		return
	} else if err != nil {
		writeServerError(c, err, "integrations.delete")
		return
	}
	h.recordAuditBestEffort(c, "integration.deleted", "integration", id, nil)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
