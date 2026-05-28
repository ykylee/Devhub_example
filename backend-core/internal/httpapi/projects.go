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

// uuidPattern — codex P2 (#397 hotfix) — UUID format pre-check 용. malformed UUID
// 가 store 의 `WHERE id = $1::uuid` cast 까지 도달하면 Postgres invalid UUID error
// 가 500 으로 노출. handler 에서 422 `application_id_invalid` 로 매핑하기 위해 미리 검증.
var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func (h *Handler) projectModel() string {
	mode := strings.ToLower(strings.TrimSpace(h.cfg.ProjectModel))
	switch mode {
	case "legacy", "v2", "hybrid":
		return mode
	default:
		return "hybrid"
	}
}

func (h *Handler) allowLegacyProjectRoutes() bool {
	mode := h.projectModel()
	return mode == "legacy" || mode == "hybrid"
}

func (h *Handler) allowV2ProjectRoutes() bool {
	mode := h.projectModel()
	return mode == "v2" || mode == "hybrid"
}

// Project CRUD endpoint (API-55..56, sprint claude/work_260514-c).
// Repository 하위 기간성 운영 단위. concept §13 + REQ-FR-PROJ-001..010.

func projectResponse(p domain.Project) gin.H {
	repositoryID := any(nil)
	if p.RepositoryID > 0 {
		repositoryID = p.RepositoryID
	}
	return gin.H{
		"id":             p.ID,
		"application_id": p.ApplicationID,
		"repository_id":  repositoryID,
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
	if !h.allowLegacyProjectRoutes() {
		c.JSON(http.StatusGone, gin.H{"status": "gone", "error": "legacy repository-centric project routes are disabled", "code": "project_model_legacy_disabled"})
		return
	}
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
	ApplicationID string  `json:"application_id"`
	RepositoryID  int64   `json:"repository_id"`
	RepositoryIDs []int64 `json:"repository_ids"`
	RepositoryCreatePayload *createRepositoryPayload `json:"repository_create_payload"`
	Key           string  `json:"key"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	OwnerUserID   string  `json:"owner_user_id"`
	StartDate     string  `json:"start_date"`
	DueDate       string  `json:"due_date"`
	Visibility    string  `json:"visibility"`
	Status        string  `json:"status"`
}

type createRepositoryPayload struct {
	Key         string `json:"key"`
	Slug        string `json:"slug"`
	SCMProvider string `json:"scm_provider"`
}

func projectRepositoryResponse(link domain.ProjectRepository) gin.H {
	return gin.H{
		"project_id":    link.ProjectID,
		"repository_id": link.RepositoryID,
		"role":          link.Role,
		"linked_at":     link.LinkedAt.UTC().Format(time.RFC3339),
	}
}

// POST /api/v1/repositories/:repository_id/projects
func (h *Handler) createProject(c *gin.Context) {
	if !h.allowLegacyProjectRoutes() {
		c.JSON(http.StatusGone, gin.H{"status": "gone", "error": "legacy repository-centric project routes are disabled", "code": "project_model_legacy_disabled"})
		return
	}
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
	h.createProjectWithRepoID(c, storeI, repoID, req)
}

func (h *Handler) createProjectWithRepoID(c *gin.Context, storeI ApplicationStore, repoID int64, req createProjectRequest) {
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

// POST /api/v1/projects
// Independent project create (application/repository optional).
func (h *Handler) createProjectStandalone(c *gin.Context) {
	if !h.allowV2ProjectRoutes() {
		c.JSON(http.StatusGone, gin.H{"status": "gone", "error": "v2 project routes are disabled", "code": "project_model_v2_disabled"})
		return
	}
	storeI, ok := h.applicationStoreOrUnavailable(c)
	if !ok {
		return
	}

	var req createProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": err.Error()})
		return
	}

	if strings.TrimSpace(req.Key) == "" || strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.OwnerUserID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "key, name, owner_user_id are required"})
		return
	}
	if !validApplicationVisibilities[req.Visibility] || !validApplicationStatuses[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "visibility/status is invalid"})
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

	primaryRepoID := int64(0)
	if req.RepositoryID > 0 {
		primaryRepoID = req.RepositoryID
	} else if len(req.RepositoryIDs) > 0 {
		primaryRepoID = req.RepositoryIDs[0]
	}
	var repoPayload *store.RepositoryCreatePayload
	if req.RepositoryCreatePayload != nil {
		repoKey := strings.TrimSpace(req.RepositoryCreatePayload.Key)
		repoSlug := strings.TrimSpace(req.RepositoryCreatePayload.Slug)
		scmProvider := strings.TrimSpace(req.RepositoryCreatePayload.SCMProvider)
		if repoKey == "" || repoSlug == "" || scmProvider == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "repository_create_payload.key/slug/scm_provider are required"})
			return
		}
		repoPayload = &store.RepositoryCreatePayload{Key: repoKey, Slug: repoSlug, SCMProvider: scmProvider}
	}

	repoSet := map[int64]struct{}{}
	if primaryRepoID > 0 {
		repoSet[primaryRepoID] = struct{}{}
	}
	for _, id := range req.RepositoryIDs {
		if id > 0 {
			repoSet[id] = struct{}{}
		}
	}
	repoIDs := make([]int64, 0, len(repoSet))
	for id := range repoSet {
		repoIDs = append(repoIDs, id)
	}

	// repo 동반 생성(repoPayload)은 project 생성과 단일 tx — project 실패 시 repo rollback (codex #349 P2).
	created, err := storeI.CreateProjectWithRepositoryPayload(c.Request.Context(), domain.Project{
		ApplicationID: strings.TrimSpace(req.ApplicationID),
		RepositoryID:  primaryRepoID,
		Key:           req.Key,
		Name:          req.Name,
		Description:   req.Description,
		Status:        domain.ApplicationStatus(req.Status),
		Visibility:    domain.ApplicationVisibility(req.Visibility),
		OwnerUserID:   req.OwnerUserID,
		StartDate:     startDate,
		DueDate:       dueDate,
	}, repoIDs, repoPayload)
	if errors.Is(err, store.ErrConflict) {
		c.JSON(http.StatusConflict, gin.H{"status": "conflict", "error": "project key already exists or referenced application/repository not found", "code": "project_key_conflict"})
		return
	}
	if err != nil {
		writeServerError(c, err, "projects.create_standalone")
		return
	}

	h.recordAuditBestEffort(c, "project.created", "project", created.ID, map[string]any{
		"key":            created.Key,
		"repository_id":  created.RepositoryID,
		"application_id": created.ApplicationID,
		"status":         string(created.Status),
	})

	c.JSON(http.StatusCreated, gin.H{"status": "ok", "data": projectResponse(created)})
}

// GET /api/v1/applications/:application_id/projects
func (h *Handler) listApplicationProjects(c *gin.Context) {
	if !h.allowV2ProjectRoutes() {
		c.JSON(http.StatusGone, gin.H{"status": "gone", "error": "v2 application-centric project routes are disabled", "code": "project_model_v2_disabled"})
		return
	}
	storeI, ok := h.applicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	appID := strings.TrimSpace(c.Param("application_id"))
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "application_id is required"})
		return
	}

	opts := store.ProjectListOptions{
		ApplicationID:   appID,
		Status:          c.Query("status"),
		IncludeArchived: c.Query("include_archived") == "true",
	}
	if opts.Status != "" && !validApplicationStatuses[opts.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "status must be one of planning/active/on_hold/closed/archived"})
		return
	}

	projects, total, err := storeI.ListProjects(c.Request.Context(), opts)
	if err != nil {
		writeServerError(c, err, "projects.list_by_application")
		return
	}
	resp := make([]gin.H, 0, len(projects))
	for _, p := range projects {
		resp = append(resp, projectResponse(p))
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "data": resp, "meta": gin.H{"total": total}})
}

// GET /api/v1/projects/standalone — application_id IS NULL projects.
// codex P2 (#397 hotfix) — ApplicationCreationModal 의 "Connected Projects" picker 가
// connected + standalone projects 합쳐 표시할 수 있도록 별도 endpoint.
func (h *Handler) listStandaloneProjects(c *gin.Context) {
	storeI, ok := h.applicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	opts := store.ProjectListOptions{
		StandaloneOnly:  true,
		Status:          c.Query("status"),
		IncludeArchived: c.Query("include_archived") == "true",
	}
	if opts.Status != "" && !validApplicationStatuses[opts.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "status must be one of planning/active/on_hold/closed/archived"})
		return
	}
	projects, total, err := storeI.ListProjects(c.Request.Context(), opts)
	if err != nil {
		writeServerError(c, err, "projects.list_standalone")
		return
	}
	resp := make([]gin.H, 0, len(projects))
	for _, p := range projects {
		resp = append(resp, projectResponse(p))
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "data": resp, "meta": gin.H{"total": total}})
}

// POST /api/v1/applications/:application_id/projects
func (h *Handler) createApplicationProject(c *gin.Context) {
	if !h.allowV2ProjectRoutes() {
		c.JSON(http.StatusGone, gin.H{"status": "gone", "error": "v2 application-centric project routes are disabled", "code": "project_model_v2_disabled"})
		return
	}
	storeI, ok := h.applicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	appID := strings.TrimSpace(c.Param("application_id"))
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "application_id is required"})
		return
	}

	var req createProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": err.Error()})
		return
	}

	primaryRepoID := int64(0)
	if req.RepositoryID > 0 {
		primaryRepoID = req.RepositoryID
	} else if len(req.RepositoryIDs) > 0 {
		primaryRepoID = req.RepositoryIDs[0]
	}
	var repoPayload *store.RepositoryCreatePayload
	if req.RepositoryCreatePayload != nil {
		repoKey := strings.TrimSpace(req.RepositoryCreatePayload.Key)
		repoSlug := strings.TrimSpace(req.RepositoryCreatePayload.Slug)
		scmProvider := strings.TrimSpace(req.RepositoryCreatePayload.SCMProvider)
		if repoKey == "" || repoSlug == "" || scmProvider == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "repository_create_payload.key/slug/scm_provider are required"})
			return
		}
		repoPayload = &store.RepositoryCreatePayload{Key: repoKey, Slug: repoSlug, SCMProvider: scmProvider}
	}

	if strings.TrimSpace(req.Key) == "" || strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.OwnerUserID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "key, name, owner_user_id are required"})
		return
	}
	if !validApplicationVisibilities[req.Visibility] || !validApplicationStatuses[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "visibility/status is invalid"})
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

	repoSet := map[int64]struct{}{}
	if primaryRepoID > 0 {
		repoSet[primaryRepoID] = struct{}{}
	}
	for _, id := range req.RepositoryIDs {
		if id > 0 {
			repoSet[id] = struct{}{}
		}
	}
	repoIDs := make([]int64, 0, len(repoSet))
	for id := range repoSet {
		repoIDs = append(repoIDs, id)
	}

	// repo 동반 생성(repoPayload)은 project 생성과 단일 tx — project 실패 시 repo rollback (codex #349 P2).
	created, err := storeI.CreateProjectWithRepositoryPayload(c.Request.Context(), domain.Project{
		ApplicationID: appID,
		RepositoryID:  primaryRepoID,
		Key:           req.Key,
		Name:          req.Name,
		Description:   req.Description,
		Status:        domain.ApplicationStatus(req.Status),
		Visibility:    domain.ApplicationVisibility(req.Visibility),
		OwnerUserID:   req.OwnerUserID,
		StartDate:     startDate,
		DueDate:       dueDate,
	}, repoIDs, repoPayload)
	if errors.Is(err, store.ErrConflict) {
		c.JSON(http.StatusConflict, gin.H{"status": "conflict", "error": "project key already exists or referenced application/repository not found", "code": "project_key_conflict"})
		return
	}
	if err != nil {
		writeServerError(c, err, "projects.create_by_application")
		return
	}

	h.recordAuditBestEffort(c, "project.created", "project", created.ID, map[string]any{
		"key":            created.Key,
		"repository_id":  created.RepositoryID,
		"application_id": appID,
		"status":         string(created.Status),
	})

	c.JSON(http.StatusCreated, gin.H{"status": "ok", "data": projectResponse(created)})
}

type createProjectRepositoryRequest struct {
	RepositoryID int64  `json:"repository_id"`
	Role         string `json:"role"`
}

// GET /api/v1/projects/:project_id/repositories
func (h *Handler) listProjectRepositories(c *gin.Context) {
	if !h.allowV2ProjectRoutes() {
		c.JSON(http.StatusGone, gin.H{"status": "gone", "error": "v2 project repository routes are disabled", "code": "project_model_v2_disabled"})
		return
	}
	storeI, ok := h.applicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	projectID := c.Param("project_id")
	links, err := storeI.ListProjectRepositories(c.Request.Context(), projectID)
	if err != nil {
		writeServerError(c, err, "projects.list_repositories")
		return
	}
	resp := make([]gin.H, 0, len(links))
	for _, link := range links {
		resp = append(resp, projectRepositoryResponse(link))
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "data": resp})
}

// POST /api/v1/projects/:project_id/repositories
func (h *Handler) createProjectRepository(c *gin.Context) {
	if !h.allowV2ProjectRoutes() {
		c.JSON(http.StatusGone, gin.H{"status": "gone", "error": "v2 project repository routes are disabled", "code": "project_model_v2_disabled"})
		return
	}
	storeI, ok := h.applicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	projectID := c.Param("project_id")
	var req createProjectRepositoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": err.Error()})
		return
	}
	if req.RepositoryID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "repository_id is required"})
		return
	}
	if req.Role == "" {
		req.Role = "linked"
	}
	created, err := storeI.CreateProjectRepository(c.Request.Context(), domain.ProjectRepository{
		ProjectID:    projectID,
		RepositoryID: req.RepositoryID,
		Role:         req.Role,
	})
	if errors.Is(err, store.ErrConflict) {
		c.JSON(http.StatusConflict, gin.H{"status": "conflict", "error": "project repository link conflict", "code": "project_repository_link_conflict"})
		return
	}
	if err != nil {
		writeServerError(c, err, "projects.create_repository")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "ok", "data": projectRepositoryResponse(created)})
}

// DELETE /api/v1/projects/:project_id/repositories/:repository_id
func (h *Handler) deleteProjectRepository(c *gin.Context) {
	if !h.allowV2ProjectRoutes() {
		c.JSON(http.StatusGone, gin.H{"status": "gone", "error": "v2 project repository routes are disabled", "code": "project_model_v2_disabled"})
		return
	}
	storeI, ok := h.applicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	projectID := c.Param("project_id")
	repositoryID, err := strconv.ParseInt(c.Param("repository_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "repository_id must be an integer"})
		return
	}
	if err := storeI.DeleteProjectRepository(c.Request.Context(), projectID, repositoryID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "project repository link not found"})
			return
		}
		writeServerError(c, err, "projects.delete_repository")
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
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
	// ApplicationID — application 이전 / 해제. nil = 변경 안 함, "" = 해제 (NULL),
	// non-empty = 해당 application 으로 이전 (존재 검증 후). #395/#396 후속 carve.
	// migration 000015 의 projects.application_id 는 nullable (ON DELETE SET NULL).
	ApplicationID  *string `json:"application_id"`
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
	// codex P2 정합 (#393): enforceRowOwnership 을 immutable-key 검증보다 먼저
	// 실행. 이전 순서는 비-소유자/PMO 미보유 호출자가 (a) mismatched key 로 PATCH
	// 시 422, (b) 정확한 key 로 PATCH 시 403 이 와서 row-level 가드를 우회 + key
	// 추측 oracle 을 제공했음. 인증/인가 검증을 항상 선행해 미인가 쓰기 시도엔
	// row-write denial 만 일관 노출하도록 정정.
	if !h.enforceRowOwnership(c, current.OwnerUserID, string(domain.AppRolePMOManager)) {
		return
	}
	if req.Key != nil && *req.Key != current.Key {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": "rejected",
			"error":  "project key is immutable",
			"code":   "project_key_immutable",
		})
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
	if req.ApplicationID != nil {
		// application 이전/해제 — #395/#396 후속 carve. ApplicationID nil = 변경 안 함,
		// "" = 해제 (NULL), non-empty = 해당 application 으로 이전 (존재 검증 + audit).
		newAppID := strings.TrimSpace(*req.ApplicationID)
		if newAppID != "" && newAppID != current.ApplicationID {
			// codex P2 (#397 hotfix) — malformed UUID 가 GetApplication 의 `$1::uuid` cast
			// 까지 도달하면 Postgres error → 500. handler 에서 422 로 미리 차단.
			if !uuidPattern.MatchString(newAppID) {
				c.JSON(http.StatusUnprocessableEntity, gin.H{
					"status": "rejected",
					"error":  "application_id must be a valid UUID",
					"code":   "application_id_invalid",
				})
				return
			}
			if _, err := storeI.GetApplication(c.Request.Context(), newAppID); errors.Is(err, store.ErrNotFound) {
				c.JSON(http.StatusUnprocessableEntity, gin.H{
					"status": "rejected",
					"error":  "application_id does not exist",
					"code":   "application_id_invalid",
				})
				return
			} else if err != nil {
				writeServerError(c, err, "projects.update.application_lookup")
				return
			}
		}
		updated.ApplicationID = newAppID
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
	if current.ApplicationID != result.ApplicationID {
		// application 이전/해제 audit. from/to 빈 string 은 NULL 의미 (해제/연결 안 함).
		payload["application_id_from"] = current.ApplicationID
		payload["application_id_to"] = result.ApplicationID
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

	isHard := c.Query("hard") == "true"

	if isHard {
		if current.Status != "archived" {
			c.JSON(http.StatusBadRequest, gin.H{"status": "bad_request", "error": "project must be archived before hard deletion"})
			return
		}
		if err := storeI.DeleteProject(c.Request.Context(), id); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "project not found"})
				return
			}
			writeServerError(c, err, "projects.delete")
			return
		}
		h.recordAuditBestEffort(c, "project.deleted", "project", id, nil)
		c.JSON(http.StatusOK, gin.H{
			"status": "deleted",
		})
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
