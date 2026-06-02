package view

import (
	"errors"
	"github.com/devhub/backend-core/internal/shared/httphelp"
	"net/http"
	"os"
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

var applicationKeyPattern = regexp.MustCompile(`^[A-Za-z0-9]{1,10}$`)

func (h *ApplicationHandler) ApplicationStoreOrUnavailable(c *gin.Context) (ApplicationStore, bool) {
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
		"link_source":          linkSourceOrDefault(link.LinkSource),
	}
}

// linkSourceOrDefault — direct/via_project 외 (legacy row 빈 string 등) 는 "direct"
// fallback. #395/#396 후속 carve P2-#3.
func linkSourceOrDefault(s string) string {
	switch s {
	case "direct", "via_project":
		return s
	default:
		return "direct"
	}
}

func scmProviderResponse(p domain.SCMProvider) gin.H {
	hasCredentials := false
	if p.ProviderKey == "gitea" {
		hasCredentials = os.Getenv("GITEA_URL") != "" && os.Getenv("GITEA_TOKEN") != ""
	}
	return gin.H{
		"provider_key":    p.ProviderKey,
		"display_name":    p.DisplayName,
		"enabled":         p.Enabled,
		"adapter_version": p.AdapterVersion,
		"has_credentials": hasCredentials,
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
	// status 전이 정책 자유화 (2026-05-28) — 운영자가 임의로 어느 상태든 이동
	// 가능. 5종 (planning/active/on_hold/closed/archived) 끼리의 모든 전이를 허용.
	// archived 도 unarchive (다른 상태로 복원) 가능. 이전엔 matrix 가 archived 의
	// 모든 outbound 전이를 거부했음 (api §13.2).
	allowedStatusTransitions = map[string]map[string]bool{
		"planning": {"planning": true, "active": true, "on_hold": true, "closed": true, "archived": true},
		"active":   {"planning": true, "active": true, "on_hold": true, "closed": true, "archived": true},
		"on_hold":  {"planning": true, "active": true, "on_hold": true, "closed": true, "archived": true},
		"closed":   {"planning": true, "active": true, "on_hold": true, "closed": true, "archived": true},
		"archived": {"planning": true, "active": true, "on_hold": true, "closed": true, "archived": true},
	}
)

// SCM Providers (API-41, API-42) ---

func (h *ApplicationHandler) ListSCMProviders(c *gin.Context) {
	storeI, ok := h.ApplicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	providers, err := storeI.ListSCMProviders(c.Request.Context())
	if err != nil {
		httphelp.WriteServerError(c, err, "scm_providers.list")
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

func (h *ApplicationHandler) UpdateSCMProvider(c *gin.Context) {
	storeI, ok := h.ApplicationStoreOrUnavailable(c)
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
		httphelp.WriteServerError(c, err, "scm_providers.lookup")
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
		httphelp.WriteServerError(c, err, "scm_providers.update")
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

func (h *ApplicationHandler) ListApplications(c *gin.Context) {
	storeI, ok := h.ApplicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	opts := store.ApplicationListOptions{
		Status:          c.Query("status"),
		IncludeArchived: c.Query("include_archived") == "true",
		Query:           c.Query("q"),
	}
	if loginVal, ok := c.Get("devhub_actor_login"); ok {
		if login, ok := loginVal.(string); ok {
			opts.ActorLogin = login
		}
	}
	if roleVal, ok := c.Get("devhub_actor_role"); ok {
		if role, ok := roleVal.(string); ok {
			opts.ActorRole = role
		}
	}
	if idsVal, ok := c.Get("devhub_actor_org_unit_ids"); ok {
		if ids, ok := idsVal.([]string); ok {
			opts.OrgUnitIDs = ids
		}
	}
	if idsVal, ok := c.Get("devhub_actor_primary_unit_ids"); ok {
		if ids, ok := idsVal.([]string); ok {
			opts.PrimaryUnitIDs = ids
		}
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
		httphelp.WriteServerError(c, err, "applications.list")
		return
	}
	visible := make([]domain.Application, 0, len(apps))
	for _, app := range apps {
		allowed, _, err := h.actorCanReadApplication(c, storeI, app)
		if err != nil {
			httphelp.WriteServerError(c, err, "applications.list.scope")
			return
		}
		if allowed {
			visible = append(visible, app)
		}
	}
	resp := make([]gin.H, 0, len(visible))
	for _, app := range visible {
		resp = append(resp, applicationResponse(app))
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   resp,
		"meta": gin.H{
			"total":     len(visible),
			"raw_total": total,
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

func (h *ApplicationHandler) CreateApplication(c *gin.Context) {
	storeI, ok := h.ApplicationStoreOrUnavailable(c)
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
			"error":  "key must match ^[A-Za-z0-9]{1,10}$",
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
		httphelp.WriteServerError(c, err, "applications.create")
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

func (h *ApplicationHandler) GetApplication(c *gin.Context) {
	storeI, ok := h.ApplicationStoreOrUnavailable(c)
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
		httphelp.WriteServerError(c, err, "applications.get")
		return
	}
	allowed, deniedReason, err := h.actorCanReadApplication(c, storeI, app)
	if err != nil {
		httphelp.WriteServerError(c, err, "applications.get.scope")
		return
	}
	if !allowed {
		h.denyRowRead(c, deniedReason)
		return
	}
	links, err := storeI.ListApplicationRepositories(c.Request.Context(), id)
	if err != nil {
		httphelp.WriteServerError(c, err, "applications.get.list_repositories")
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

func (h *ApplicationHandler) ApplicationDashboard(c *gin.Context) {
	storeI, ok := h.ApplicationStoreOrUnavailable(c)
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
		httphelp.WriteServerError(c, err, "applications.dashboard.get")
		return
	}
	allowed, deniedReason, err := h.actorCanReadApplication(c, storeI, app)
	if err != nil {
		httphelp.WriteServerError(c, err, "applications.dashboard.scope")
		return
	}
	if !allowed {
		h.denyRowRead(c, deniedReason)
		return
	}

	// data_gaps 는 응답의 부분 데이터 표지 — ARCH-APPDASH-05 Graceful Degradation 의
	// 실 구현 (codex 3b 후속). 외부 데이터 미가용 / 미구현 / sync 비활성 등 모든 누락 사유를
	// 명시 token 으로 누적해 UI 가 분기 표시할 수 있게 한다.
	dataGaps := make([]string, 0)

	// 1. Rollup metrics — 외부 데이터 미존재 시 0 값으로 fallback (ComputeApplicationRollup 자체가
	//    빈 윈도우면 zero-value 반환).
	rollupOpts := domain.ApplicationRollupOptions{
		Policy:     domain.WeightPolicyEqual,
		WindowFrom: time.Now().UTC().AddDate(0, 0, -30),
		WindowTo:   time.Now().UTC(),
	}
	rollup, err := storeI.ComputeApplicationRollup(c.Request.Context(), id, rollupOpts)
	if err != nil {
		httphelp.WriteServerError(c, err, "applications.dashboard.rollup")
		return
	}

	// 2. Linked Application↔Repository links.
	links, err := storeI.ListApplicationRepositories(c.Request.Context(), id)
	if err != nil {
		httphelp.WriteServerError(c, err, "applications.dashboard.list_repositories")
		return
	}

	// 3. Application 의 Projects (milestones).
	projOpts := store.ProjectListOptions{
		ApplicationID:   id,
		IncludeArchived: false,
	}
	if loginVal, ok := c.Get("devhub_actor_login"); ok {
		if login, ok := loginVal.(string); ok {
			projOpts.ActorLogin = login
		}
	}
	if roleVal, ok := c.Get("devhub_actor_role"); ok {
		if role, ok := roleVal.(string); ok {
			projOpts.ActorRole = role
		}
	}
	if idsVal, ok := c.Get("devhub_actor_org_unit_ids"); ok {
		if ids, ok := idsVal.([]string); ok {
			projOpts.OrgUnitIDs = ids
		}
	}
	if idsVal, ok := c.Get("devhub_actor_primary_unit_ids"); ok {
		if ids, ok := idsVal.([]string); ok {
			projOpts.PrimaryUnitIDs = ids
		}
	}
	projects, _, err := storeI.ListProjects(c.Request.Context(), projOpts)
	if err != nil {
		httphelp.WriteServerError(c, err, "applications.dashboard.list_projects")
		return
	}
	// project id set — DREQ 의 project 로 등록된 항목도 본 application 매핑으로 포함.
	projectIDSet := make(map[string]struct{}, len(projects))
	for _, p := range projects {
		projectIDSet[p.ID] = struct{}{}
	}

	// 4. Linked DREQs — 정확한 등록 매칭만 사용 (이전 substring(title/details, app.Key)
	//    휴리스틱은 false positive 와 project 등록 누락이 있어 제거, codex 3c).
	//    DREQ store 미설정 시 503 으로 응답 가로채지 말고 graceful 빈 목록 + data_gap 기록.
	var dreqs []domain.DevRequest
	if h.cfg.DevRequestStore != nil {
		dreqOpts := store.DevRequestListOptions{
			Statuses: []domain.DevRequestStatus{
				domain.DevRequestStatusPending,
				domain.DevRequestStatusInReview,
			},
		}
		allDreqs, _, drErr := h.cfg.DevRequestStore.ListDevRequests(c.Request.Context(), dreqOpts)
		if drErr != nil {
			dataGaps = append(dataGaps, "linked_dev_requests:store_error")
		} else {
			for _, dr := range allDreqs {
				if dr.RegisteredTargetType == domain.DevRequestTargetApplication && dr.RegisteredTargetID == app.ID {
					dreqs = append(dreqs, dr)
					continue
				}
				if dr.RegisteredTargetType == domain.DevRequestTargetProject {
					if _, in := projectIDSet[dr.RegisteredTargetID]; in {
						dreqs = append(dreqs, dr)
					}
				}
			}
		}
	} else {
		dataGaps = append(dataGaps, "linked_dev_requests:store_unavailable")
	}

	// 5. 실패 빌드 런 — `ApplicationRepository` 도메인엔 직접 RepositoryID FK 가 없으므로
	//    (composite PK = ApplicationID + RepoProvider + RepoFullName) 외부 repo id 는
	//    `ListRepositoriesByProvider` 로 해석해야 한다. 이전 구현은 **link 마다** 공급자 전체
	//    repo 를 다시 가져와 클라이언트 매칭(N×M)했지만, 본 PR 은 **provider 별 1회 fetch +
	//    {full_name → id} 맵 캐싱** 으로 M (unique providers) + L (links) 으로 축소 (codex 3d
	//    정정). sync 가 active 아닌 link 는 data_gap (ARCH-APPDASH-05).
	repoIDByProvider := make(map[string]map[string]int64)
	resolveRepoID := func(provider, fullName string) (int64, error) {
		byName, cached := repoIDByProvider[provider]
		if !cached {
			repos, listErr := storeI.ListRepositoriesByProvider(c.Request.Context(), provider)
			if listErr != nil {
				repoIDByProvider[provider] = nil // 음성 캐시 — 다음 link 가 같은 provider 일 때 재요청 회피.
				return 0, listErr
			}
			byName = make(map[string]int64, len(repos))
			for _, r := range repos {
				byName[r.FullName] = r.ID
			}
			repoIDByProvider[provider] = byName
		}
		return byName[fullName], nil
	}

	var buildFailures []gin.H
	for _, link := range links {
		if link.SyncStatus != domain.SyncStatusActive {
			dataGaps = append(dataGaps, "build_runs:"+link.RepoFullName+":sync_"+string(link.SyncStatus))
			continue
		}
		repoID, resolveErr := resolveRepoID(link.RepoProvider, link.RepoFullName)
		if resolveErr != nil {
			dataGaps = append(dataGaps, "build_runs:"+link.RepoFullName+":provider_lookup_error")
			continue
		}
		if repoID == 0 {
			dataGaps = append(dataGaps, "build_runs:"+link.RepoFullName+":repo_not_found")
			continue
		}
		buildOpts := store.BuildRunListOptions{Status: "failed", Limit: 5}
		failedRuns, _, brErr := storeI.ListRepositoryBuildRuns(c.Request.Context(), repoID, buildOpts)
		if brErr != nil {
			dataGaps = append(dataGaps, "build_runs:"+link.RepoFullName+":fetch_error")
			continue
		}
		for _, fr := range failedRuns {
			failedAt := ""
			if fr.FinishedAt != nil {
				failedAt = fr.FinishedAt.UTC().Format(time.RFC3339)
			}
			buildFailures = append(buildFailures, gin.H{
				"repo_provider": link.RepoProvider,
				"repo_slug":     link.RepoFullName,
				"branch":        fr.Branch,
				"build_number":  fr.ID,
				"failed_at":     failedAt,
				"error_snippet": "", // BuildRun 도메인엔 에러 요약 필드 없음 — UI 는 log_url 로 이동.
				"log_url":       "/api/v1/ci-runs/" + strconv.FormatInt(fr.ID, 10) + "/logs",
			})
		}
	}

	// 6. Format Response.
	// codex P2 정합 (#396) — target_branch_build_status 의 source 를 rollup.TargetBranchBuildStatus
	// (각 repo 의 build_runs 최신 1건 종합 derive) 로 통일. 이전 `len(buildFailures) > 0`
	// 휴리스틱은 dashboard 자체의 별도 broken 판단이라 application_rollup endpoint 와
	// 결과가 일치하지 않을 수 있었음 (build_failures 는 별도 store 호출 결과, rollup 은
	// last build status 종합 derive). 빈 문자열은 "unknown" 정규화.
	targetBranchBuildStatus := rollup.TargetBranchBuildStatus
	if targetBranchBuildStatus == "" {
		targetBranchBuildStatus = "unknown"
	}
	metricsOverview := gin.H{
		"target_branch_build_status": targetBranchBuildStatus,
		"avg_build_duration_seconds": rollup.BuildAvgDurationSeconds,
		"quality_score":              rollup.QualityScore,
		"critical_warning_count":     rollup.CriticalWarningCount,
	}

	qualityMetrics := gin.H{
		"normalized_score": rollup.QualityScore,
		"unresolved_issues": gin.H{
			"blocker":  rollup.QualityGateFailedCount,
			"critical": 0,
			"major":    rollup.CriticalWarningCount,
		},
		"comment": "코딩룰/세부 린터 위반 내역은 개별 레포지토리 대시보드에서 상세 제공",
	}

	projectsProgress := make([]gin.H, 0, len(projects))
	for _, p := range projects {
		dDay := 0
		riskLevel := "Healthy"
		riskBadgeColor := "#4CAF50"
		if p.DueDate != nil {
			duration := time.Until(*p.DueDate)
			dDay = int(duration.Hours() / 24)
			if dDay <= 7 {
				riskLevel = "At Risk"
				riskBadgeColor = "#F44336"
			} else if dDay <= 14 {
				riskLevel = "Warning"
				riskBadgeColor = "#FFC107"
			}
		}

		projectsProgress = append(projectsProgress, gin.H{
			"project_id":       p.ID,
			"key":              p.Key,
			"name":             p.Name,
			"progress_percent": 0, // story-point 기반 진척율 계산은 미구현 — data_gap 으로 표기 후 carve.
			"status":           string(p.Status),
			"due_date":         formatDatePtr(p.DueDate),
			"d_day":            dDay,
			"risk_level":       riskLevel,
			"risk_badge_color": riskBadgeColor,
		})
	}
	if len(projects) > 0 {
		dataGaps = append(dataGaps, "projects_progress:story_points_not_implemented")
	}

	linkedDevRequests := make([]gin.H, 0, len(dreqs))
	for _, dr := range dreqs {
		linkedDevRequests = append(linkedDevRequests, gin.H{
			"dreq_id":               dr.ID,
			"title":                 dr.Title,
			"status":                string(dr.Status),
			"assignee_display_name": dr.AssigneeUserID, // user_id 직접 노출 — name resolver 는 후속 carve.
			"created_at":            dr.CreatedAt.UTC().Format(time.RFC3339),
		})
	}

	// 시계열 트렌드는 build_runs/quality_snapshots 의 일자별 집계 store 메서드 부재로
	// 본 PR 에서는 빈 배열로 응답 (이전 구현은 현재 rollup 값에 임의 delta(-0.05/-0.1)
	// 더한 합성 데이터였음 — 실데이터 아니라 noise 라 제거).
	historyTrend := []gin.H{}
	dataGaps = append(dataGaps, "history_trend:not_implemented")

	appliedWeights := make(map[string]float64)
	if len(links) > 0 {
		for _, link := range links {
			appliedWeights[link.RepoFullName] = 1.0 / float64(len(links))
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data": gin.H{
			"application_id":      app.ID,
			"key":                 app.Key,
			"name":                app.Name,
			"status":              string(app.Status),
			"visibility":          string(app.Visibility),
			"leader":              app.LeaderUserID,      // user_id 직접 — name resolver 후속 carve.
			"development_unit":    app.DevelopmentUnitID, // unit_id 직접 — label resolver 후속 carve.
			"updated_at":          app.UpdatedAt.UTC().Format(time.RFC3339),
			"metrics_overview":    metricsOverview,
			"build_failures":      buildFailures,
			"quality_metrics":     qualityMetrics,
			"projects_progress":   projectsProgress,
			"linked_dev_requests": linkedDevRequests,
			"history_trend":       historyTrend,
		},
		"meta": gin.H{
			"weight_policy":   "equal",
			"applied_weights": appliedWeights,
			"fallbacks":       []string{},
			"data_gaps":       dataGaps,
		},
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

func (h *ApplicationHandler) UpdateApplication(c *gin.Context) {
	storeI, ok := h.ApplicationStoreOrUnavailable(c)
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
		httphelp.WriteServerError(c, err, "applications.update.lookup")
		return
	}

	// ADR-0011 §4.2 row-level 위양: system_admin / team_manager / owner-self 만 허용.
	// route-level RBAC gate 가 applications:edit 를 권한 있는 role 만 통과시키고,
	// 본 helper 가 row 단위 owner/위양 검사. caller 가 owner 와 일치하거나 위양
	// role 인 경우 통과, 그 외는 audit auth.row_denied + 403.
	if !h.enforceRowOwnership(c, current.OwnerUserID, string(domain.AppRoleTeamManager)) {
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
			// status 전이 정책 자유화 (2026-05-28) — 모든 전이 가드 (planning→active
			// 의 active repo ≥1, active→closed 의 critical 0건, active→on_hold 의
			// hold_reason 등) 제거. 운영자가 임의로 어느 전이든 가능. reason 필드는
			// audit 기록용 optional 메타로만 유지.
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
		httphelp.WriteServerError(c, err, "applications.update")
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

// ArchiveApplication — DELETE /api/v1/applications/:id. `?hard=true` 면 archived 상태에서만
// hard-delete (project handler 와 동일 패턴), 그 외엔 archive (soft-delete).
func (h *ApplicationHandler) ArchiveApplication(c *gin.Context) {
	storeI, ok := h.ApplicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	id := c.Param("application_id")
	var req archiveApplicationRequest
	// DELETE body 가 비어도 허용 (archived_reason 은 권장).
	_ = c.ShouldBindJSON(&req)

	// ADR-0011 §4.2 row-level 위양: archive 도 owner-self 또는 team_manager 가
	// 가능해야 하므로 archive 직전에 lookup + 검증한다. Application 이 없으면
	// ArchiveApplication 의 ErrNotFound 분기와 동일하게 404.
	current, err := storeI.GetApplication(c.Request.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "application not found"})
		return
	}
	if err != nil {
		httphelp.WriteServerError(c, err, "applications.archive.lookup")
		return
	}
	if !h.enforceRowOwnership(c, current.OwnerUserID, string(domain.AppRoleTeamManager)) {
		return
	}

	isHard := c.Query("hard") == "true"
	if isHard {
		if string(current.Status) != "archived" {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "bad_request",
				"error":  "application must be archived before hard deletion",
				"code":   "application_not_archived",
			})
			return
		}
		if err := storeI.DeleteApplication(c.Request.Context(), id); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "application not found"})
				return
			}
			httphelp.WriteServerError(c, err, "applications.delete")
			return
		}
		h.recordAuditBestEffort(c, "application.deleted", "application", id, nil)
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
		return
	}

	archived, err := storeI.ArchiveApplication(c.Request.Context(), id, req.ArchivedReason)
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "application not found"})
		return
	}
	if err != nil {
		httphelp.WriteServerError(c, err, "applications.archive")
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

func (h *ApplicationHandler) ListApplicationRepositories(c *gin.Context) {
	storeI, ok := h.ApplicationStoreOrUnavailable(c)
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
		httphelp.WriteServerError(c, err, "application_repositories.scope_lookup")
		return
	}
	allowed, deniedReason, err := h.actorCanReadApplication(c, storeI, app)
	if err != nil {
		httphelp.WriteServerError(c, err, "application_repositories.scope")
		return
	}
	if !allowed {
		h.denyRowRead(c, deniedReason)
		return
	}
	links, err := storeI.ListApplicationRepositories(c.Request.Context(), id)
	if err != nil {
		httphelp.WriteServerError(c, err, "application_repositories.list")
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

func (h *ApplicationHandler) CreateApplicationRepository(c *gin.Context) {
	storeI, ok := h.ApplicationStoreOrUnavailable(c)
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
		httphelp.WriteServerError(c, err, "application_repositories.lookup_provider")
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
		httphelp.WriteServerError(c, err, "application_repositories.create")
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

func (h *ApplicationHandler) DeleteApplicationRepository(c *gin.Context) {
	storeI, ok := h.ApplicationStoreOrUnavailable(c)
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
		httphelp.WriteServerError(c, err, "application_repositories.delete")
		return
	}
	h.recordAuditBestEffort(c, "application_repository.unlinked", "application", id, map[string]any{
		"repo_provider":  parts[0],
		"repo_full_name": parts[1],
	})
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
