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

// Application/Repository/Project 관리 API handler stubs (API-43..50, planned 활성화).
//
// 본 sprint (claude/work_260514-a) 의 carve in 은 handler stub + RBAC gate 등록까지.
// 실 store 호출과 응답 body 는 후속 sprint 의 carve out (state.json carve_out 참조).
// 현재는 모두 `501 not_implemented` envelope 응답 — 단, RBAC matrix 가 호출자 role 을
// 먼저 거부하므로 (ADR-0011 §4.1 의 system_admin 일임) developer/manager 는 `403` 을
//받고 system_admin 만 `501` 까지 도달한다.

func (h *Handler) notImplemented(c *gin.Context, codeHint string) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"status": "error",
		"error": gin.H{
			"code":    "not_implemented",
			"message": "endpoint scaffolded; PostgreSQL store body and request validation pending (sprint claude/work_260514-a carve out).",
			"hint":    codeHint,
		},
	})
}

// SCM Providers (API-41, API-42)

func (h *Handler) listSCMProviders(c *gin.Context) {
	h.notImplemented(c, "API-41")
}

func (h *Handler) updateSCMProvider(c *gin.Context) {
	h.notImplemented(c, "API-42")
}

// Applications (API-43..47)

func (h *Handler) listApplications(c *gin.Context) {
	h.notImplemented(c, "API-43")
}

type createApplicationRequest struct {
	Key         string `json:"key" binding:"required,alphanum,max=10"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	OwnerUserID string `json:"owner_user_id" binding:"required"`
	Visibility  string `json:"visibility" binding:"required,oneof=public internal restricted"`
	Status      string `json:"status" binding:"required,oneof=planning active on_hold closed archived"`
	StartDate   string `json:"start_date"`
	DueDate     string `json:"due_date"`
}

func (h *Handler) createApplication(c *gin.Context) {
	var req createApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	app := domain.Application{
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
		OwnerUserID: req.OwnerUserID,
		Visibility:  domain.ApplicationVisibility(req.Visibility),
		Status:      domain.ApplicationStatus(req.Status),
	}

	if req.StartDate != "" {
		t, err := time.Parse("2006-01-02", req.StartDate)
		if err == nil {
			app.StartDate = &t
		}
	}
	if req.DueDate != "" {
		t, err := time.Parse("2006-01-02", req.DueDate)
		if err == nil {
			app.DueDate = &t
		}
	}

	createdApp, err := h.cfg.ApplicationStore.CreateApplication(c.Request.Context(), app)
	if err != nil {
		if errors.Is(err, store.ErrConflict) {
			c.JSON(http.StatusConflict, gin.H{"error": "application key already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": createdApp})
}

func (h *Handler) getApplication(c *gin.Context) {
	appID := c.Param("application_id")
	app, err := h.cfg.ApplicationStore.GetApplication(c.Request.Context(), appID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": app})
}

type updateApplicationRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	OwnerUserID *string `json:"owner_user_id"`
	Visibility  *string `json:"visibility" binding:"omitempty,oneof=public internal restricted"`
	Status      *string `json:"status" binding:"omitempty,oneof=planning active on_hold closed"`
	StartDate   *string `json:"start_date"`
	DueDate     *string `json:"due_date"`
}

func (h *Handler) updateApplication(c *gin.Context) {
	appID := c.Param("application_id")
	var req updateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existingApp, err := h.cfg.ApplicationStore.GetApplication(c.Request.Context(), appID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if req.Name != nil {
		existingApp.Name = *req.Name
	}
	if req.Description != nil {
		existingApp.Description = *req.Description
	}
	if req.OwnerUserID != nil {
		existingApp.OwnerUserID = *req.OwnerUserID
	}
	if req.Visibility != nil {
		existingApp.Visibility = domain.ApplicationVisibility(*req.Visibility)
	}
	if req.Status != nil {
		existingApp.Status = domain.ApplicationStatus(*req.Status)
	}
	if req.StartDate != nil {
		t, err := time.Parse("2006-01-02", *req.StartDate)
		if err == nil {
			existingApp.StartDate = &t
		}
	}
	if req.DueDate != nil {
		t, err := time.Parse("2006-01-02", *req.DueDate)
		if err == nil {
			existingApp.DueDate = &t
		}
	}

	updatedApp, err := h.cfg.ApplicationStore.UpdateApplication(c.Request.Context(), existingApp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": updatedApp})
}

type archiveApplicationRequest struct {
	Reason string `json:"reason"`
}

func (h *Handler) archiveApplication(c *gin.Context) {
	appID := c.Param("application_id")
	var req archiveApplicationRequest
	_ = c.ShouldBindJSON(&req) // Reason is optional

	app, err := h.cfg.ApplicationStore.ArchiveApplication(c.Request.Context(), appID, req.Reason)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "application not found or already archived"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": app})
}

// Application-Repository link (API-48..50)

func (h *Handler) listApplicationRepositories(c *gin.Context) {
	appID := c.Param("application_id")
	// Basic check if application exists before listing repos.
	if _, err := h.cfg.ApplicationStore.GetApplication(c.Request.Context(), appID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	repos, err := h.cfg.ApplicationStore.ListApplicationRepositories(c.Request.Context(), appID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": repos})
}

type createApplicationRepositoryRequest struct {
	RepoProvider string `json:"repo_provider" binding:"required"`
	RepoFullName string `json:"repo_full_name" binding:"required"`
	Role         string `json:"role" binding:"required,oneof=primary sub shared"`
}

func (h *Handler) createApplicationRepository(c *gin.Context) {
	appID := c.Param("application_id")
	var req createApplicationRepositoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	link := domain.ApplicationRepository{
		ApplicationID: appID,
		RepoProvider:  req.RepoProvider,
		RepoFullName:  req.RepoFullName,
		Role:          domain.ApplicationRepositoryRole(req.Role),
	}

	createdLink, err := h.cfg.ApplicationStore.CreateApplicationRepository(c.Request.Context(), link)
	if err != nil {
		if errors.Is(err, store.ErrConflict) {
			c.JSON(http.StatusConflict, gin.H{"error": "repository is already linked to this application"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": createdLink})
}

func (h *Handler) deleteApplicationRepository(c *gin.Context) {
	appID := c.Param("application_id")
	repoKey, err := strconv.Unquote(c.Param("repo_key"))
	if err != nil {
		repoKey = c.Param("repo_key") // Fallback if not quoted
	}

	parts := strings.SplitN(repoKey, "/", 2)
	if len(parts) != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid repo_key format, expected 'provider/org/repo'"})
		return
	}

	key := store.ApplicationRepositoryLinkKey{
		ApplicationID: appID,
		RepoProvider:  parts[0],
		RepoFullName:  parts[1],
	}

	if err := h.cfg.ApplicationStore.DeleteApplicationRepository(c.Request.Context(), key); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "repository link not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
