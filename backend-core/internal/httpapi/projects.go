package httpapi

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

// --- Projects ---

func (h *Handler) createProject(c *gin.Context) {
	repoIDStr := c.Param("repository_id")
	repoID, err := strconv.ParseInt(repoIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid repository_id"})
		return
	}

	var req domain.Project // Using domain model directly for simplicity
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.RepositoryID = repoID

	createdProject, err := h.cfg.ApplicationStore.CreateProject(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, store.ErrConflict) {
			c.JSON(http.StatusConflict, gin.H{"error": "project key already exists in this repository"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": createdProject})
}

func (h *Handler) getProject(c *gin.Context) {
	projectID := c.Param("project_id")
	p, err := h.cfg.ApplicationStore.GetProject(c.Request.Context(), projectID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": p})
}

func (h *Handler) listProjects(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	repoID, _ := strconv.ParseInt(c.Query("repository_id"), 10, 64)

	opts := store.ProjectListOptions{
		Status:          c.Query("status"),
		IncludeArchived: c.Query("include_archived") == "true",
		RepositoryID:    repoID,
		Limit:           limit,
		Offset:          offset,
	}

	projects, total, err := h.cfg.ApplicationStore.ListProjects(c.Request.Context(), opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.Header("X-Total-Count", strconv.Itoa(total))
	c.JSON(http.StatusOK, gin.H{"data": projects})
}

type updateProjectRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	OwnerUserID *string `json:"owner_user_id"`
	Visibility  *string `json:"visibility" binding:"omitempty,oneof=public internal restricted"`
	Status      *string `json:"status" binding:"omitempty,oneof=planning active on_hold closed"`
	StartDate   *string `json:"start_date"`
	DueDate     *string `json:"due_date"`
}


func (h *Handler) updateProject(c *gin.Context) {
	projectID := c.Param("project_id")
	var req updateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existing, err := h.cfg.ApplicationStore.GetProject(c.Request.Context(), projectID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.OwnerUserID != nil {
		existing.OwnerUserID = *req.OwnerUserID
	}
	if req.Visibility != nil {
		existing.Visibility = domain.ApplicationVisibility(*req.Visibility)
	}
	if req.Status != nil {
		existing.Status = domain.ProjectStatus(*req.Status)
	}
	if req.StartDate != nil {
		t, err := time.Parse("2006-01-02", *req.StartDate)
		if err == nil {
			existing.StartDate = &t
		}
	}
	if req.DueDate != nil {
		t, err := time.Parse("2006-01-02", *req.DueDate)
		if err == nil {
			existing.DueDate = &t
		}
	}

	updated, err := h.cfg.ApplicationStore.UpdateProject(c.Request.Context(), existing)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": updated})
}

func (h *Handler) archiveProject(c *gin.Context) {
	projectID := c.Param("project_id")
	var req archiveApplicationRequest // Reuse from applications
	_ = c.ShouldBindJSON(&req)

	p, err := h.cfg.ApplicationStore.ArchiveProject(c.Request.Context(), projectID, req.Reason)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found or already archived"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": p})
}
