package httpapi

import (
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

// Application/Repository/Project 관리 API handlers (API-41..50).
//
// sprint claude/work_260514-b 가 직전 sprint 의 stub (501) 을 정식 응답으로 교체.
// 상태 전이 가드 1차: planning→active 의 활성 Repository ≥1, immutable key, hold/resume/
// archived_reason 필수 (concept §13.2.1). active→closed 의 critical 롤업 0건 검증은
// 롤업 store 의존이라 carve out (다음 sprint).
//
// 권한은 enforceRoutePermission middleware 가 사전 거부. handler 까지 도달하면 ADR-0011
// §4.1 의 system_admin 자격 통과 상태.

var applicationKeyPattern = regexp.MustCompile(`^[A-Za-z0-9]{10}$`)

func (h *Handler) applicationStoreOrUnavailable(c *gin.Context) (ApplicationStore, bool) {
	if h.cfg.ApplicationStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "application store is not configured",
		})
		return nil, false
	}
	return h.cfg.ApplicationStore, true
}

// applicationResponse converts a domain.Application to the wire shape used by §13.2.
func applicationResponse(app domain.Application) gin.H {
	return gin.H{
		"id":                  app.ID,
		"key":                 app.Key,
		"name":                app.Name,
		"description":         app.Description,
		"status":              string(app.Status),
		"visibility":          string(app.Visibility),
		"owner_user_id":       app.OwnerUserID,
		"leader_user_id":      app.LeaderUserID,
		"development_unit_id": app.DevelopmentUnitID,
		"start_date":          formatDatePtr(app.StartDate),
		"due_date":            formatDatePtr(app.DueDate),
		"archived_at":         formatTimePtr(app.ArchivedAt),
		"created_at":          app.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at":          app.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func applicationRepositoryResponse(link domain.ApplicationRepository) gin.H {
	return gin.H{
		"application_id":       link.ApplicationID,
		"repo_provider":        link.RepoProvider,
		"repo_full_name":       link.RepoFullName,
		"external_repo_id":     link.ExternalRepoID,
		"role":                 string(link.Role),
		"sync_status":          string(link.SyncStatus),
		"sync_error_code":      string(link.SyncErrorCode),
		"sync_error_retryable": link.SyncErrorRetryable,
		"sync_error_at":        formatTimePtr(link.SyncErrorAt),
		"last_sync_at":         formatTimePtr(link.LastSyncAt),
		"linked_at":            link.LinkedAt.UTC().Format(time.RFC3339),
	}
}

func scmProviderResponse(p domain.SCMProvider) gin.H {
	return gin.H{
		"provider_key":    p.ProviderKey,
		"display_name":    p.DisplayName,
		"enabled":         p.Enabled,
		"adapter_version": p.AdapterVersion,
		"created_at":      p.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at":      p.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func formatDatePtr(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.UTC().Format("2006-01-02")
}

func formatTimePtr(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.UTC().Format(time.RFC3339)
}

// parseDate 는 RFC3339 ("YYYY-MM-DD") 입력을 *time.Time 으로 변환. 빈 문자열은 nil.
func parseDate(s string) (*time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// validApplicationStatus / validApplicationVisibility helpers — concept §13.2 + api §13.1.
var (
	validApplicationStatuses = map[string]bool{
		"planning": true, "active": true, "on_hold": true, "closed": true, "archived": true,
	}
	validApplicationVisibilities = map[string]bool{
		"public": true, "internal": true, "restricted": true,
	}
	validApplicationRepoRoles = map[string]bool{
		"primary": true, "sub": true, "shared": true,
	}
	allowedStatusTransitions = map[string]map[string]bool{
		"planning": {"active": true, "on_hold": true, "archived": true},
		"active":   {"on_hold": true, "closed": true, "archived": true},
		"on_hold":  {"active": true, "closed": true, "archived": true},
		"closed":   {"archived": true},
		"archived": {}, // 모든 outbound 전이 거부 (api §13.2)
	}
)

// SCM Providers (API-41, API-42) ---

func (h *Handler) listSCMProviders(c *gin.Context) {
	storeI, ok := h.applicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	providers, err := storeI.ListSCMProviders(c.Request.Context())
	if err != nil {
		writeServerError(c, err, "scm_providers.list")
		return
	}
	resp := make([]gin.H, 0, len(providers))
	for _, p := range providers {
		resp = append(resp, scmProviderResponse(p))
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   resp,
	})
}

type updateSCMProviderRequest struct {
	DisplayName    *string `json:"display_name"`
	Enabled        *bool   `json:"enabled"`
	AdapterVersion *string `json:"adapter_version"` // 거부용 — 클라이언트가 보내면 422
}

func (h *Handler) updateSCMProvider(c *gin.Context) {
	storeI, ok := h.applicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	providerKey := c.Param("provider_key")
	if providerKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "provider_key is required"})
		return
	}
	var req updateSCMProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": err.Error()})
		return
	}
	if req.AdapterVersion != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": "rejected",
			"error":  "adapter_version is managed by the deployment pipeline and cannot be set via API",
			"code":   "adapter_version_immutable",
		})
		return
	}
	providers, err := storeI.ListSCMProviders(c.Request.Context())
	if err != nil {
		writeServerError(c, err, "scm_providers.lookup")
		return
	}
	var target *domain.SCMProvider
	for i := range providers {
		if providers[i].ProviderKey == providerKey {
			target = &providers[i]
			break
		}
	}
	if target == nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "scm provider not found"})
		return
	}
	if req.DisplayName != nil {
		target.DisplayName = *req.DisplayName
	}
	if req.Enabled != nil {
		target.Enabled = *req.Enabled
	}
	updated, err := storeI.UpdateSCMProvider(c.Request.Context(), *target)
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "scm provider not found"})
		return
	}
	if err != nil {
		writeServerError(c, err, "scm_providers.update")
		return
	}
	h.recordAuditBestEffort(c, "scm_provider.updated", "scm_provider", providerKey, map[string]any{
		"enabled": updated.Enabled,
	})
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   scmProviderResponse(updated),
	})
}

// Applications (API-43..47) ---

func (h *Handler) listApplications(c *gin.Context) {
	storeI, ok := h.applicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	opts := store.ApplicationListOptions{
		Status:          c.Query("status"),
		IncludeArchived: c.Query("include_archived") == "true",
		Query:           c.Query("q"),
	}
	if s := c.Query("limit"); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil || v < 1 || v > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "limit must be an integer between 1 and 100"})
			return
		}
		opts.Limit = v
	}
	if s := c.Query("offset"); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil || v < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "offset must be a non-negative integer"})
			return
		}
		opts.Offset = v
	}
	if opts.Status != "" && !validApplicationStatuses[opts.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "status must be one of planning/active/on_hold/closed/archived"})
		return
	}
	apps, total, err := storeI.ListApplications(c.Request.Context(), opts)
	if err != nil {
		writeServerError(c, err, "applications.list")
		return
	}
	resp := make([]gin.H, 0, len(apps))
	for _, app := range apps {
		resp = append(resp, applicationResponse(app))
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   resp,
		"meta": gin.H{
			"total": total,
		},
	})
}

type createApplicationRequest struct {
	Key               string `json:"key"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	OwnerUserID       string `json:"owner_user_id"`
	LeaderUserID      string `json:"leader_user_id"`
	DevelopmentUnitID string `json:"development_unit_id"`
	StartDate         string `json:"start_date"`
	DueDate           string `json:"due_date"`
	Visibility        string `json:"visibility"`
	Status            string `json:"status"`
}

func (h *Handler) createApplication(c *gin.Context) {
	storeI, ok := h.applicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	var req createApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": err.Error()})
		return
	}
	if !applicationKeyPattern.MatchString(req.Key) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": "rejected",
			"error":  "key must match ^[A-Za-z0-9]{10}$",
			"code":   "invalid_application_key",
		})
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
	if strings.TrimSpace(req.LeaderUserID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "leader_user_id is required"})
		return
	}
	if strings.TrimSpace(req.DevelopmentUnitID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "development_unit_id is required"})
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
	app := domain.Application{
		Key:               req.Key,
		Name:              req.Name,
		Description:       req.Description,
		Status:            domain.ApplicationStatus(req.Status),
		Visibility:        domain.ApplicationVisibility(req.Visibility),
		OwnerUserID:       req.OwnerUserID,
		LeaderUserID:      req.LeaderUserID,
		DevelopmentUnitID: req.DevelopmentUnitID,
		StartDate:         startDate,
		DueDate:           dueDate,
	}
	created, err := storeI.CreateApplication(c.Request.Context(), app)
	if errors.Is(err, store.ErrConflict) {
		c.JSON(http.StatusConflict, gin.H{
			"status": "conflict",
			"error":  "application key already exists",
			"code":   "application_key_conflict",
		})
		return
	}
	if err != nil {
		writeServerError(c, err, "applications.create")
		return
	}
	h.recordAuditBestEffort(c, "application.created", "application", created.ID, map[string]any{
		"key":    created.Key,
		"status": string(created.Status),
	})
	c.JSON(http.StatusCreated, gin.H{
		"status": "ok",
		"data":   applicationResponse(created),
	})
}

func (h *Handler) getApplication(c *gin.Context) {
	storeI, ok := h.applicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	id := c.Param("application_id")
	app, err := storeI.GetApplication(c.Request.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "application not found"})
		return
	}
	if err != nil {
		writeServerError(c, err, "applications.get")
		return
	}
	links, err := storeI.ListApplicationRepositories(c.Request.Context(), id)
	if err != nil {
		writeServerError(c, err, "applications.get.list_repositories")
		return
	}
	repoResp := make([]gin.H, 0, len(links))
	for _, l := range links {
		repoResp = append(repoResp, applicationRepositoryResponse(l))
	}
	resp := applicationResponse(app)
	resp["repositories"] = repoResp
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   resp,
	})
}

type updateApplicationRequest struct {
	Key               *string `json:"key"` // 거부용
	Name              *string `json:"name"`
	Description       *string `json:"description"`
	OwnerUserID       *string `json:"owner_user_id"`
	LeaderUserID      *string `json:"leader_user_id"`
	DevelopmentUnitID *string `json:"development_unit_id"`
	StartDate         *string `json:"start_date"`
	DueDate           *string `json:"due_date"`
	Visibility        *string `json:"visibility"`
	Status            *string `json:"status"`
	HoldReason        string  `json:"hold_reason"`
	ResumeReason      string  `json:"resume_reason"`
	ArchivedReason    string  `json:"archived_reason"`
}

func (h *Handler) updateApplication(c *gin.Context) {
	storeI, ok := h.applicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	id := c.Param("application_id")
	var req updateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": err.Error()})
		return
	}
	if req.Key != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": "rejected",
			"error":  "application key is immutable",
			"code":   "application_key_immutable",
		})
		return
	}
	current, err := storeI.GetApplication(c.Request.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "application not found"})
		return
	}
	if err != nil {
		writeServerError(c, err, "applications.update.lookup")
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
	if req.LeaderUserID != nil {
		if strings.TrimSpace(*req.LeaderUserID) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "leader_user_id cannot be empty"})
			return
		}
		updated.LeaderUserID = *req.LeaderUserID
	}
	if req.DevelopmentUnitID != nil {
		if strings.TrimSpace(*req.DevelopmentUnitID) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "development_unit_id cannot be empty"})
			return
		}
		updated.DevelopmentUnitID = *req.DevelopmentUnitID
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
			// 전이 가드 (concept §13.2.1, 1차).
			switch {
			case curStatus == "planning" && newStatus == "active":
				cnt, err := storeI.CountActiveApplicationRepositories(c.Request.Context(), id)
				if err != nil {
					writeServerError(c, err, "applications.update.count_active_repos")
					return
				}
				if cnt < 1 {
					c.JSON(http.StatusUnprocessableEntity, gin.H{
						"status": "rejected",
						"error":  "planning→active requires at least one active repository link",
						"code":   "application_activation_precondition_failed",
					})
					return
				}
			case curStatus == "active" && newStatus == "on_hold":
				if strings.TrimSpace(req.HoldReason) == "" {
					c.JSON(http.StatusUnprocessableEntity, gin.H{
						"status": "rejected",
						"error":  "active→on_hold requires hold_reason",
						"code":   "invalid_status_transition_payload",
					})
					return
				}
			case curStatus == "on_hold" && newStatus == "active":
				if strings.TrimSpace(req.ResumeReason) == "" {
					c.JSON(http.StatusUnprocessableEntity, gin.H{
						"status": "rejected",
						"error":  "on_hold→active requires resume_reason",
						"code":   "invalid_status_transition_payload",
					})
					return
				}
			case newStatus == "archived":
				if strings.TrimSpace(req.ArchivedReason) == "" {
					c.JSON(http.StatusUnprocessableEntity, gin.H{
						"status": "rejected",
						"error":  "transition to archived requires archived_reason",
						"code":   "invalid_status_transition_payload",
					})
					return
				}
			case curStatus == "active" && newStatus == "closed":
				// concept §13.2.1 의 "active → closed: 롤업 critical 0건" 가드 (sprint
				// claude/work_260514-c 가 흡수). 롤업 store 의 critical_warning_count
				// 가 1 이상이면 close 거부 — 운영자가 critical 데이터 손실을 모르고
				// closing 하는 사고 방지.
				count, err := storeI.CountApplicationCriticalWarnings(c.Request.Context(), id)
				if err != nil {
					writeServerError(c, err, "applications.update.critical_warnings")
					return
				}
				if count > 0 {
					c.JSON(http.StatusUnprocessableEntity, gin.H{
						"status":                 "rejected",
						"error":                  "active→closed requires critical warning count = 0",
						"code":                   "application_close_precondition_failed",
						"critical_warning_count": count,
					})
					return
				}
			}
		}
		updated.Status = domain.ApplicationStatus(newStatus)
	}

	result, err := storeI.UpdateApplication(c.Request.Context(), updated)
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "application not found"})
		return
	}
	if errors.Is(err, store.ErrConflict) {
		c.JSON(http.StatusConflict, gin.H{"status": "conflict", "error": "owner_user_id not found"})
		return
	}
	if err != nil {
		writeServerError(c, err, "applications.update")
		return
	}
	auditPayload := map[string]any{
		"from_status": string(current.Status),
		"to_status":   string(result.Status),
	}
	if req.HoldReason != "" {
		auditPayload["hold_reason"] = req.HoldReason
	}
	if req.ResumeReason != "" {
		auditPayload["resume_reason"] = req.ResumeReason
	}
	if req.ArchivedReason != "" {
		auditPayload["archived_reason"] = req.ArchivedReason
	}
	h.recordAuditBestEffort(c, "application.updated", "application", id, auditPayload)
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   applicationResponse(result),
	})
}

type archiveApplicationRequest struct {
	ArchivedReason string `json:"archived_reason"`
}

func (h *Handler) archiveApplication(c *gin.Context) {
	storeI, ok := h.applicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	id := c.Param("application_id")
	var req archiveApplicationRequest
	// DELETE body 가 비어도 허용 (archived_reason 은 권장).
	_ = c.ShouldBindJSON(&req)
	archived, err := storeI.ArchiveApplication(c.Request.Context(), id, req.ArchivedReason)
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "application not found"})
		return
	}
	if err != nil {
		writeServerError(c, err, "applications.archive")
		return
	}
	h.recordAuditBestEffort(c, "application.archived", "application", id, map[string]any{
		"archived_reason": req.ArchivedReason,
	})
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   applicationResponse(archived),
	})
}

// Application-Repository link (API-48..50) ---

func (h *Handler) listApplicationRepositories(c *gin.Context) {
	storeI, ok := h.applicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	id := c.Param("application_id")
	links, err := storeI.ListApplicationRepositories(c.Request.Context(), id)
	if err != nil {
		writeServerError(c, err, "application_repositories.list")
		return
	}
	resp := make([]gin.H, 0, len(links))
	for _, l := range links {
		resp = append(resp, applicationRepositoryResponse(l))
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   resp,
	})
}

type createApplicationRepositoryRequest struct {
	RepoProvider   string `json:"repo_provider"`
	RepoFullName   string `json:"repo_full_name"`
	Role           string `json:"role"`
	ExternalRepoID string `json:"external_repo_id"`
}

func (h *Handler) createApplicationRepository(c *gin.Context) {
	storeI, ok := h.applicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	id := c.Param("application_id")
	var req createApplicationRepositoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": err.Error()})
		return
	}
	if strings.TrimSpace(req.RepoProvider) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "repo_provider is required"})
		return
	}
	if strings.TrimSpace(req.RepoFullName) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "repo_full_name is required"})
		return
	}
	if !validApplicationRepoRoles[req.Role] {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "role must be one of primary/sub/shared"})
		return
	}
	providers, err := storeI.ListSCMProviders(c.Request.Context())
	if err != nil {
		writeServerError(c, err, "application_repositories.lookup_provider")
		return
	}
	enabled := false
	for _, p := range providers {
		if p.ProviderKey == req.RepoProvider && p.Enabled {
			enabled = true
			break
		}
	}
	if !enabled {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": "rejected",
			"error":  "repo_provider is not registered or disabled",
			"code":   "unsupported_repo_provider",
		})
		return
	}
	link := domain.ApplicationRepository{
		ApplicationID:  id,
		RepoProvider:   req.RepoProvider,
		RepoFullName:   req.RepoFullName,
		ExternalRepoID: req.ExternalRepoID,
		Role:           domain.ApplicationRepositoryRole(req.Role),
		SyncStatus:     domain.SyncStatusRequested,
	}
	created, err := storeI.CreateApplicationRepository(c.Request.Context(), link)
	if errors.Is(err, store.ErrConflict) {
		c.JSON(http.StatusConflict, gin.H{
			"status": "conflict",
			"error":  "repository link already exists or application not found",
			"code":   "repository_link_conflict",
		})
		return
	}
	if err != nil {
		writeServerError(c, err, "application_repositories.create")
		return
	}
	h.recordAuditBestEffort(c, "application_repository.linked", "application", id, map[string]any{
		"repo_provider":  created.RepoProvider,
		"repo_full_name": created.RepoFullName,
		"role":           string(created.Role),
	})
	c.JSON(http.StatusCreated, gin.H{
		"status": "ok",
		"data":   applicationRepositoryResponse(created),
	})
}

func (h *Handler) deleteApplicationRepository(c *gin.Context) {
	storeI, ok := h.applicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	id := c.Param("application_id")
	// repo_key = "{provider}:{full_name}". gin 의 catch-all (`*repo_key`) 이라 leading `/`
	// 가 붙어 들어옴. 클라이언트가 `//provider:repo` 같은 잘못된 입력을 보내도 leading `/`
	// 를 모두 제거하기 위해 TrimLeft 사용. provider:org/repo 컨벤션 — 콜론으로 분리.
	repoKey := strings.TrimLeft(c.Param("repo_key"), "/")
	parts := strings.SplitN(repoKey, ":", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "rejected",
			"error":  "repo_key must be 'provider:org/repo' (e.g. 'gitea:team/devhub-core')",
		})
		return
	}
	linkKey := store.ApplicationRepositoryLinkKey{
		ApplicationID: id,
		RepoProvider:  parts[0],
		RepoFullName:  parts[1],
	}
	if err := storeI.DeleteApplicationRepository(c.Request.Context(), linkKey); errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "repository link not found"})
		return
	} else if err != nil {
		writeServerError(c, err, "application_repositories.delete")
		return
	}
	h.recordAuditBestEffort(c, "application_repository.unlinked", "application", id, map[string]any{
		"repo_provider":  parts[0],
		"repo_full_name": parts[1],
	})
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
