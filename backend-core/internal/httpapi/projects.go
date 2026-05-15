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

// Project CRUD endpoint (API-55..56, sprint claude/work_260514-c).
// Repository 하위 기간성 운영 단위. concept §13 + REQ-FR-PROJ-001..010.

func projectResponse(p domain.Project) gin.H {
	return gin.H{
		"id":             p.ID,
		"application_id": p.ApplicationID,
		"repository_id":  p.RepositoryID,
		"key":            p.Key,
		"name":           p.Name,
		"description":    p.Description,
		"status":         string(p.Status),
		"visibility":     string(p.Visibility),
		"owner_user_id":  p.OwnerUserID,
		"start_date":     formatDatePtr(p.StartDate),
		"due_date":       formatDatePtr(p.DueDate),
		"archived_at":    formatTimePtr(p.ArchivedAt),
		"created_at":     p.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at":     p.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

// GET /api/v1/repositories/:repository_id/projects
func (h *Handler) listProjects(c *gin.Context) {
	storeI, ok := h.applicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	repoID, err := strconv.ParseInt(c.Param("repository_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "repository_id must be an integer"})
		return
	}
	opts := store.ProjectListOptions{
		RepositoryID:    repoID,
		Status:          c.Query("status"),
		IncludeArchived: c.Query("include_archived") == "true",
	}
	if opts.Status != "" && !validApplicationStatuses[opts.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "status must be one of planning/active/on_hold/closed/archived"})
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
	projects, total, err := storeI.ListProjects(c.Request.Context(), opts)
	if err != nil {
		writeServerError(c, err, "projects.list")
		return
	}
	resp := make([]gin.H, 0, len(projects))
	for _, p := range projects {
		resp = append(resp, projectResponse(p))
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   resp,
		"meta":   gin.H{"total": total},
	})
}

type createProjectRequest struct {
	ApplicationID string `json:"application_id"`
	Key           string `json:"key"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	OwnerUserID   string `json:"owner_user_id"`
	StartDate     string `json:"start_date"`
	DueDate       string `json:"due_date"`
	Visibility    string `json:"visibility"`
	Status        string `json:"status"`
}

// POST /api/v1/repositories/:repository_id/projects
func (h *Handler) createProject(c *gin.Context) {
	storeI, ok := h.applicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	repoID, err := strconv.ParseInt(c.Param("repository_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "repository_id must be an integer"})
		return
	}
	var req createProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": err.Error()})
		return
	}
	if strings.TrimSpace(req.Key) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "key is required"})
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "name is required"})
		return
	}
	if strings.TrimSpace(req.OwnerUserID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "owner_user_id is required"})
		return
	}
	if !validApplicationVisibilities[req.Visibility] {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "visibility must be one of public/internal/restricted"})
		return
	}
	if !validApplicationStatuses[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "status must be one of planning/active/on_hold/closed/archived"})
		return
	}
	startDate, err := parseDate(req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "start_date must be YYYY-MM-DD"})
		return
	}
	dueDate, err := parseDate(req.DueDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "due_date must be YYYY-MM-DD"})
		return
	}
	project := domain.Project{
		ApplicationID: req.ApplicationID,
		RepositoryID:  repoID,
		Key:           req.Key,
		Name:          req.Name,
		Description:   req.Description,
		Status:        domain.ApplicationStatus(req.Status),
		Visibility:    domain.ApplicationVisibility(req.Visibility),
		OwnerUserID:   req.OwnerUserID,
		StartDate:     startDate,
		DueDate:       dueDate,
	}
	created, err := storeI.CreateProject(c.Request.Context(), project)
	if errors.Is(err, store.ErrConflict) {
		c.JSON(http.StatusConflict, gin.H{
			"status": "conflict",
			"error":  "project key already exists or referenced application/repository not found",
			"code":   "project_key_conflict",
		})
		return
	}
	if err != nil {
		writeServerError(c, err, "projects.create")
		return
	}
	h.recordAuditBestEffort(c, "project.created", "project", created.ID, map[string]any{
		"key":           created.Key,
		"repository_id": created.RepositoryID,
		"status":        string(created.Status),
	})
	c.JSON(http.StatusCreated, gin.H{
		"status": "ok",
		"data":   projectResponse(created),
	})
}

// GET /api/v1/projects/:project_id
func (h *Handler) getProject(c *gin.Context) {
	storeI, ok := h.applicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	id := c.Param("project_id")
	p, err := storeI.GetProject(c.Request.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "project not found"})
		return
	}
	if err != nil {
		writeServerError(c, err, "projects.get")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   projectResponse(p),
	})
}

type updateProjectRequest struct {
	Key            *string `json:"key"` // 거부용
	Name           *string `json:"name"`
	Description    *string `json:"description"`
	OwnerUserID    *string `json:"owner_user_id"`
	StartDate      *string `json:"start_date"`
	DueDate        *string `json:"due_date"`
	Visibility     *string `json:"visibility"`
	Status         *string `json:"status"`
	HoldReason     string  `json:"hold_reason"`
	ResumeReason   string  `json:"resume_reason"`
	ArchivedReason string  `json:"archived_reason"`
}

// PATCH /api/v1/projects/:project_id
func (h *Handler) updateProject(c *gin.Context) {
	storeI, ok := h.applicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	id := c.Param("project_id")
	var req updateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": err.Error()})
		return
	}
	if req.Key != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": "rejected",
			"error":  "project key is immutable",
			"code":   "project_key_immutable",
		})
		return
	}
	current, err := storeI.GetProject(c.Request.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "project not found"})
		return
	}
	if err != nil {
		writeServerError(c, err, "projects.update.lookup")
		return
	}

	// ADR-0011 §4.2 row-level 위양 (Application 과 동일 패턴).
	if !h.enforceRowOwnership(c, current.OwnerUserID, string(domain.AppRolePMOManager)) {
		return
	}

	updated := current
	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "name cannot be empty"})
			return
		}
		updated.Name = *req.Name
	}
	if req.Description != nil {
		updated.Description = *req.Description
	}
	if req.OwnerUserID != nil {
		updated.OwnerUserID = *req.OwnerUserID
	}
	if req.Visibility != nil {
		if !validApplicationVisibilities[*req.Visibility] {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "visibility must be one of public/internal/restricted"})
			return
		}
		updated.Visibility = domain.ApplicationVisibility(*req.Visibility)
	}
	if req.StartDate != nil {
		d, err := parseDate(*req.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "start_date must be YYYY-MM-DD"})
			return
		}
		updated.StartDate = d
	}
	if req.DueDate != nil {
		d, err := parseDate(*req.DueDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "due_date must be YYYY-MM-DD"})
			return
		}
		updated.DueDate = d
	}
	if req.Status != nil {
		newStatus := *req.Status
		if !validApplicationStatuses[newStatus] {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "status must be one of planning/active/on_hold/closed/archived"})
			return
		}
		curStatus := string(current.Status)
		if newStatus != curStatus {
			allowed := allowedStatusTransitions[curStatus]
			if !allowed[newStatus] {
				c.JSON(http.StatusUnprocessableEntity, gin.H{
					"status": "rejected",
					"error":  "invalid status transition",
					"code":   "invalid_status_transition",
					"from":   curStatus,
					"to":     newStatus,
				})
				return
			}
			switch {
			case curStatus == "active" && newStatus == "on_hold":
				if strings.TrimSpace(req.HoldReason) == "" {
					c.JSON(http.StatusUnprocessableEntity, gin.H{"status": "rejected", "error": "active→on_hold requires hold_reason", "code": "invalid_status_transition_payload"})
					return
				}
			case curStatus == "on_hold" && newStatus == "active":
				if strings.TrimSpace(req.ResumeReason) == "" {
					c.JSON(http.StatusUnprocessableEntity, gin.H{"status": "rejected", "error": "on_hold→active requires resume_reason", "code": "invalid_status_transition_payload"})
					return
				}
			case newStatus == "archived":
				if strings.TrimSpace(req.ArchivedReason) == "" {
					c.JSON(http.StatusUnprocessableEntity, gin.H{"status": "rejected", "error": "transition to archived requires archived_reason", "code": "invalid_status_transition_payload"})
					return
				}
			}
		}
		updated.Status = domain.ApplicationStatus(newStatus)
	}

	result, err := storeI.UpdateProject(c.Request.Context(), updated)
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "project not found"})
		return
	}
	if err != nil {
		writeServerError(c, err, "projects.update")
		return
	}
	payload := map[string]any{
		"from_status": string(current.Status),
		"to_status":   string(result.Status),
	}
	if req.HoldReason != "" {
		payload["hold_reason"] = req.HoldReason
	}
	if req.ResumeReason != "" {
		payload["resume_reason"] = req.ResumeReason
	}
	if req.ArchivedReason != "" {
		payload["archived_reason"] = req.ArchivedReason
	}
	h.recordAuditBestEffort(c, "project.updated", "project", id, payload)
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   projectResponse(result),
	})
}

type archiveProjectRequest struct {
	ArchivedReason string `json:"archived_reason"`
}

// DELETE /api/v1/projects/:project_id (archive)
func (h *Handler) archiveProject(c *gin.Context) {
	storeI, ok := h.applicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	id := c.Param("project_id")
	var req archiveProjectRequest
	_ = c.ShouldBindJSON(&req)

	// ADR-0011 §4.2 row-level 위양: archive 도 owner-self / pmo_manager 가 가능.
	current, err := storeI.GetProject(c.Request.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "project not found"})
		return
	}
	if err != nil {
		writeServerError(c, err, "projects.archive.lookup")
		return
	}
	if !h.enforceRowOwnership(c, current.OwnerUserID, string(domain.AppRolePMOManager)) {
		return
	}

	archived, err := storeI.ArchiveProject(c.Request.Context(), id, req.ArchivedReason)
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "project not found"})
		return
	}
	if err != nil {
		writeServerError(c, err, "projects.archive")
		return
	}
	h.recordAuditBestEffort(c, "project.archived", "project", id, map[string]any{
		"archived_reason": req.ArchivedReason,
	})
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   projectResponse(archived),
	})
}
