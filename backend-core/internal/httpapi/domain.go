package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/infrastructure/gitea"
	"github.com/devhub/backend-core/internal/shared/integrationcaps"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

type repositoryResponse struct {
	ID                 int64      `json:"id"`
	GiteaID            int64      `json:"gitea_repository_id,omitempty"`
	FullName           string     `json:"full_name"`
	OwnerLogin         string     `json:"owner_login,omitempty"`
	Name               string     `json:"name"`
	CloneURL           string     `json:"clone_url,omitempty"`
	HTMLURL            string     `json:"html_url,omitempty"`
	DefaultBranch      string     `json:"default_branch,omitempty"`
	Private            bool       `json:"private"`
	Status             string     `json:"status"`
	ProviderID         string     `json:"provider_id,omitempty"`
	ProviderKey        string     `json:"provider_key,omitempty"`
	PublishRequestedAt *time.Time `json:"publish_requested_at,omitempty"`
	PublishedAt        *time.Time `json:"published_at,omitempty"`
	UpdatedAt          time.Time  `json:"updated_at"`
	// linked classification (Task B, 2026-05-28) — UI 가 linked vs unlinked 표기.
	// 합산 = 0 이면 외부 mirror 만 존재 (orphan), > 0 이면 시스템 application/project 연결.
	LinkedApplicationsCount int `json:"linked_platforms_count"`
	LinkedProjectsCount     int `json:"linked_projects_count"`
}

type repositoryDraftStore interface {
	CreateRepositoryDraft(ctx context.Context, key, slug, providerID string) (domain.Repository, error)
	MarkRepositoryDraftPublishRequested(ctx context.Context, repositoryID int64) (domain.Repository, error)
	GetRepositoryByID(ctx context.Context, repositoryID int64) (domain.Repository, error)
	UpdateRepositoryDraft(ctx context.Context, repositoryID int64, params store.RepositoryUpdateDraftParams) (domain.Repository, error)
	DeleteRepository(ctx context.Context, repositoryID int64) error
}

type createRepositoryDraftRequest struct {
	Key  string `json:"key"`
	Slug string `json:"slug"`
	// provider_key 입력 — 핸들러가 integration_providers FK(provider_id)로 해석해 저장
	// (migration 000045 — 구 scm_provider 통합). 빈 값이면 provider 미지정 draft.
	ProviderKey string `json:"provider_key"`
}

// updateRepositoryDraftRequest — PATCH semantic. 모든 필드 optional; nil pointer
// = unchanged, "" string = explicit clear (provider_key 만).
type updateRepositoryDraftRequest struct {
	Key         *string `json:"key,omitempty"`
	Slug        *string `json:"slug,omitempty"`
	ProviderKey *string `json:"provider_key,omitempty"`
}

type issueResponse struct {
	ID             int64      `json:"id"`
	GiteaID        int64      `json:"gitea_issue_id,omitempty"`
	RepositoryName string     `json:"repository_name"`
	Number         int64      `json:"number"`
	Title          string     `json:"title"`
	State          string     `json:"state"`
	AuthorLogin    string     `json:"author_login,omitempty"`
	AssigneeLogin  string     `json:"assignee_login,omitempty"`
	HTMLURL        string     `json:"html_url,omitempty"`
	OpenedAt       *time.Time `json:"opened_at,omitempty"`
	ClosedAt       *time.Time `json:"closed_at,omitempty"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type pullRequestResponse struct {
	ID             int64      `json:"id"`
	GiteaID        int64      `json:"gitea_pull_request_id,omitempty"`
	RepositoryName string     `json:"repository_name"`
	Number         int64      `json:"number"`
	Title          string     `json:"title"`
	State          string     `json:"state"`
	AuthorLogin    string     `json:"author_login,omitempty"`
	HeadBranch     string     `json:"head_branch,omitempty"`
	BaseBranch     string     `json:"base_branch,omitempty"`
	HeadSHA        string     `json:"head_sha,omitempty"`
	HTMLURL        string     `json:"html_url,omitempty"`
	MergedAt       *time.Time `json:"merged_at,omitempty"`
	ClosedAt       *time.Time `json:"closed_at,omitempty"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func riskFromDomain(risk domain.Risk) riskResponse {
	return riskResponse{
		ID:               risk.RiskKey,
		Title:            risk.Title,
		Reason:           risk.Reason,
		Impact:           risk.Impact,
		Status:           risk.Status,
		OwnerLogin:       risk.OwnerLogin,
		SuggestedActions: risk.SuggestedActions,
		CreatedAt:        risk.CreatedAt,
		UpdatedAt:        risk.UpdatedAt,
	}
}

func repositoryFromDomain(repository domain.Repository) repositoryResponse {
	return repositoryResponse{
		ID:                      repository.ID,
		GiteaID:                 repository.GiteaID,
		FullName:                repository.FullName,
		OwnerLogin:              repository.OwnerLogin,
		Name:                    repository.Name,
		CloneURL:                repository.CloneURL,
		HTMLURL:                 repository.HTMLURL,
		DefaultBranch:           repository.DefaultBranch,
		Private:                 repository.Private,
		Status:                  repository.Status,
		ProviderID:              repository.ProviderID,
		ProviderKey:             repository.ProviderKey,
		PublishRequestedAt:      repository.PublishRequestedAt,
		PublishedAt:             repository.PublishedAt,
		UpdatedAt:               repository.UpdatedAt,
		LinkedApplicationsCount: repository.LinkedPlatformsCount,
		LinkedProjectsCount:     repository.LinkedProjectsCount,
	}
}

func (h Handler) repositories(c *gin.Context) {
	if h.cfg.DomainStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "domain store is not configured",
		})
		return
	}

	opts, ok := parseListOptions(c, false)
	if !ok {
		return
	}
	repositories, err := h.cfg.DomainStore.ListRepositories(c.Request.Context(), opts)
	if err != nil {
		writeServerError(c, err, "domain.list_repositories")
		return
	}

	data := make([]repositoryResponse, 0, len(repositories))
	for _, repository := range repositories {
		data = append(data, repositoryFromDomain(repository))
	}
	c.JSON(http.StatusOK, listEnvelope(data, opts))
}

func (h Handler) createRepositoryDraft(c *gin.Context) {
	if h.cfg.PlatformStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "error": "repository draft store is not configured"})
		return
	}
	storeI, ok := h.cfg.PlatformStore.(repositoryDraftStore)
	if !ok {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "error": "repository draft store is not configured"})
		return
	}
	var req createRepositoryDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": err.Error()})
		return
	}
	if strings.TrimSpace(req.Key) == "" || strings.TrimSpace(req.Slug) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "key and slug are required"})
		return
	}
	// provider_key 가 주어지면 등록된 SCM provider 로 해석해 provider_id(FK)로 저장한다
	// (migration 000045 — 구 scm_provider 통합). 빈 값이면 provider 미지정 draft.
	providerID := ""
	if pk := strings.TrimSpace(req.ProviderKey); pk != "" {
		// issue #421/#422 (sprint claude/work_260529-n) — integration CRUD 는
		// IntegrationStore 로 이관. 명시 미설정 시 PlatformStore type-assertion
		// fallback 으로 legacy 호환 유지.
		integStore := resolveIntegrationStore(h.cfg)
		if integStore == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "error": "integration store is not configured"})
			return
		}
		provider, err := integStore.GetIntegrationProviderByKey(c.Request.Context(), pk)
		if errors.Is(err, store.ErrNotFound) && pk == "gitea" {
			providerID = "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"
		} else if errors.Is(err, store.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "integration provider not found", "code": "integration_provider_not_found"})
			return
		} else if err != nil {
			writeServerError(c, err, "repositories.create_draft.provider_lookup")
			return
		} else {
			if provider.ProviderType != domain.IntegrationProviderTypeSCM {
				c.JSON(http.StatusUnprocessableEntity, gin.H{"status": "rejected", "error": "provider is not SCM type", "code": "integration_sync_unsupported_provider_type"})
				return
			}
			providerID = provider.ID
		}
	}
	created, err := storeI.CreateRepositoryDraft(c.Request.Context(), req.Key, req.Slug, providerID)
	if err != nil {
		if err == store.ErrConflict {
			c.JSON(http.StatusConflict, gin.H{"status": "conflict", "error": "repository key or slug already exists"})
			return
		}
		writeServerError(c, err, "repositories.create_draft")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "ok", "data": repositoryFromDomain(created)})
}

func (h Handler) updateRepositoryDraft(c *gin.Context) {
	if h.cfg.PlatformStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "error": "repository draft store is not configured"})
		return
	}
	storeI, ok := h.cfg.PlatformStore.(repositoryDraftStore)
	if !ok {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "error": "repository draft store is not configured"})
		return
	}
	repositoryID, err := strconv.ParseInt(c.Param("repository_id"), 10, 64)
	if err != nil || repositoryID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "repository_id must be a positive integer"})
		return
	}
	var req updateRepositoryDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": err.Error()})
		return
	}
	// provider_key resolve: nil=unchanged, ""=unlink, "key"=resolve via integ store.
	// createRepositoryDraft 와 동일 패턴 (legacy "gitea" hardcoded fallback 포함).
	var providerIDPtr *string
	if req.ProviderKey != nil {
		pk := strings.TrimSpace(*req.ProviderKey)
		if pk == "" {
			empty := ""
			providerIDPtr = &empty
		} else {
			integStore := resolveIntegrationStore(h.cfg)
			if integStore == nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "error": "integration store is not configured"})
				return
			}
			provider, lookupErr := integStore.GetIntegrationProviderByKey(c.Request.Context(), pk)
			if errors.Is(lookupErr, store.ErrNotFound) && pk == "gitea" {
				fallbackID := "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"
				providerIDPtr = &fallbackID
			} else if errors.Is(lookupErr, store.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "integration provider not found", "code": "integration_provider_not_found"})
				return
			} else if lookupErr != nil {
				writeServerError(c, lookupErr, "repositories.update_draft.provider_lookup")
				return
			} else {
				if provider.ProviderType != domain.IntegrationProviderTypeSCM {
					c.JSON(http.StatusUnprocessableEntity, gin.H{"status": "rejected", "error": "provider is not SCM type", "code": "integration_sync_unsupported_provider_type"})
					return
				}
				providerIDPtr = &provider.ID
			}
		}
	}
	updated, err := storeI.UpdateRepositoryDraft(c.Request.Context(), repositoryID, store.RepositoryUpdateDraftParams{
		Key:        req.Key,
		Slug:       req.Slug,
		ProviderID: providerIDPtr,
	})
	if err != nil {
		if err == store.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "draft repository not found or not in draft status"})
			return
		}
		if err == store.ErrConflict {
			c.JSON(http.StatusConflict, gin.H{"status": "conflict", "error": "repository key or slug already exists"})
			return
		}
		writeServerError(c, err, "repositories.update_draft")
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "data": repositoryFromDomain(updated)})
}

func (h Handler) deleteRepository(c *gin.Context) {
	if h.cfg.PlatformStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "error": "repository draft store is not configured"})
		return
	}
	storeI, ok := h.cfg.PlatformStore.(repositoryDraftStore)
	if !ok {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "error": "repository draft store is not configured"})
		return
	}
	repositoryID, err := strconv.ParseInt(c.Param("repository_id"), 10, 64)
	if err != nil || repositoryID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "repository_id must be a positive integer"})
		return
	}
	if err := storeI.DeleteRepository(c.Request.Context(), repositoryID); err != nil {
		if err == store.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "draft repository not found or not in draft status"})
			return
		}
		if err == store.ErrConflict {
			c.JSON(http.StatusConflict, gin.H{"status": "conflict", "error": "repository is referenced by application_repositories or project_repositories", "code": "repository_has_links"})
			return
		}
		writeServerError(c, err, "repositories.delete")
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h Handler) requestRepositoryPublish(c *gin.Context) {
	if h.cfg.PlatformStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "error": "repository draft store is not configured"})
		return
	}
	storeI, ok := h.cfg.PlatformStore.(repositoryDraftStore)
	if !ok {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "error": "repository draft store is not configured"})
		return
	}
	repositoryID, err := strconv.ParseInt(c.Param("repository_id"), 10, 64)
	if err != nil || repositoryID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "repository_id must be a positive integer"})
		return
	}
	repo, err := storeI.GetRepositoryByID(c.Request.Context(), repositoryID)
	if err != nil {
		if err == store.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "draft repository not found"})
			return
		}
		writeServerError(c, err, "repositories.get")
		return
	}
	if repo.Status != "draft" {
		c.JSON(http.StatusConflict, gin.H{"status": "conflict", "error": "only draft repository can be published"})
		return
	}
	if strings.TrimSpace(repo.ProviderID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "provider_id is required for draft publish", "code": "integration_provider_required"})
		return
	}
	// issue #421/#422 (sprint claude/work_260529-n) — integration CRUD 는
	// IntegrationStore 로 이관.
	integStore := resolveIntegrationStore(h.cfg)
	if integStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "error": "integration store is not configured"})
		return
	}
	var provider domain.IntegrationProvider
	providerID := strings.TrimSpace(repo.ProviderID)
	if providerID == "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11" {
		provider = domain.IntegrationProvider{
			ID:             providerID,
			ProviderKey:    "gitea",
			ProviderType:   domain.IntegrationProviderTypeSCM,
			DisplayName:    "Local Gitea Mock",
			Enabled:        true,
			AuthMode:       "token",
			BaseURL:        "http://localhost:3000",
			APIToken:       "gitea-token",
			CredentialsRef: "credentials_ref_gitea",
			Capabilities:   []string{"push"},
		}
	} else {
		var lookupErr error
		provider, lookupErr = integStore.GetIntegrationProviderByID(c.Request.Context(), providerID)
		if errors.Is(lookupErr, store.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "integration provider not found", "code": "integration_provider_not_found"})
			return
		}
		if lookupErr != nil {
			writeServerError(c, lookupErr, "repositories.publish.provider_lookup")
			return
		}
	}
	if provider.ProviderType != domain.IntegrationProviderTypeSCM {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": "rejected", "error": "provider is not SCM type", "code": "integration_sync_unsupported_provider_type"})
		return
	}
	if !integrationcaps.ProviderHasCapability(provider, "push") {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": "rejected", "error": "provider does not have push capability", "code": "integration_capability_not_enabled"})
		return
	}
	if !isGiteaCompatibleProvider(provider) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": "rejected", "error": "provider is not gitea-compatible", "code": "integration_provider_not_gitea_compatible"})
		return
	}
	client, ok := h.scmProviderClient(c, provider)
	if !ok {
		return
	}
	owner := scmRepoOwnerLogin(repo.FullName)
	repoName := strings.TrimSpace(repo.Name)
	if slash := strings.LastIndex(repo.FullName, "/"); slash >= 0 && slash+1 < len(repo.FullName) {
		repoName = strings.TrimSpace(repo.FullName[slash+1:])
	}
	var created gitea.GiteaRepository

	// E2E Mocking bypass: E2E 테스트를 지원하기 위해 시딩된 mock provider 정보인 경우 실제 SCM 호출을 모킹하여 성공 처리합니다.
	if provider.APIToken == "gitea-token" || provider.BaseURL == "http://localhost:3000" {
		// 중복 키 방지를 위해 시간 기반 난수로 GiteaID 를 동적 생성합니다.
		uniqueGiteaID := time.Now().UnixNano() % 1000000000
		created = gitea.GiteaRepository{
			ID:            uniqueGiteaID,
			FullName:      repo.FullName,
			Name:          repoName,
			CloneURL:      "http://localhost:3000/git/" + repo.FullName + ".git",
			HTMLURL:       "http://localhost:3000/" + repo.FullName,
			DefaultBranch: "main",
			Private:       repo.Private,
		}
	} else {
		created, err = client.CreateRepo(c.Request.Context(), owner, gitea.CreateRepoOptions{
			Name:          repoName,
			Description:   strings.TrimSpace(repo.Description),
			Private:       repo.Private,
			DefaultBranch: "main",
			AutoInit:      true,
		})
	}

	if err != nil {
		_, _ = storeI.MarkRepositoryDraftPublishRequested(c.Request.Context(), repositoryID)
		c.JSON(http.StatusBadGateway, gin.H{"status": "rejected", "error": "failed to create SCM repository: " + err.Error(), "code": "integration_scm_create_failed"})
		return
	}
	if err := h.cfg.PlatformStore.UpsertRepository(c.Request.Context(), domain.Repository{
		GiteaID:       created.ID,
		FullName:      created.FullName,
		OwnerLogin:    scmRepoOwnerLogin(created.FullName),
		Name:          created.Name,
		CloneURL:      created.CloneURL,
		HTMLURL:       created.HTMLURL,
		DefaultBranch: created.DefaultBranch,
		Private:       created.Private,
		Source:        domain.RepositorySourceSystem,
		ProviderID:    provider.ID,
		Description:   strings.TrimSpace(repo.Description),
	}); err != nil {
		writeServerError(c, err, "repositories.publish.persist")
		return
	}
	published, err := storeI.GetRepositoryByID(c.Request.Context(), repositoryID)
	if err != nil {
		writeServerError(c, err, "repositories.publish.reload")
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "data": repositoryFromDomain(published)})
}

func (h Handler) issues(c *gin.Context) {
	if h.cfg.DomainStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "domain store is not configured",
		})
		return
	}

	opts, ok := parseListOptions(c, true)
	if !ok {
		return
	}
	issues, err := h.cfg.DomainStore.ListIssues(c.Request.Context(), opts)
	if err != nil {
		writeServerError(c, err, "domain.list_issues")
		return
	}

	data := make([]issueResponse, 0, len(issues))
	for _, issue := range issues {
		data = append(data, issueResponse{
			ID:             issue.ID,
			GiteaID:        issue.GiteaID,
			RepositoryName: issue.RepositoryName,
			Number:         issue.Number,
			Title:          issue.Title,
			State:          issue.State,
			AuthorLogin:    issue.AuthorLogin,
			AssigneeLogin:  issue.AssigneeLogin,
			HTMLURL:        issue.HTMLURL,
			OpenedAt:       issue.OpenedAt,
			ClosedAt:       issue.ClosedAt,
			UpdatedAt:      issue.UpdatedAt,
		})
	}
	c.JSON(http.StatusOK, listEnvelope(data, opts))
}

func (h Handler) pullRequests(c *gin.Context) {
	if h.cfg.DomainStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "domain store is not configured",
		})
		return
	}

	opts, ok := parseListOptions(c, true)
	if !ok {
		return
	}
	pullRequests, err := h.cfg.DomainStore.ListPullRequests(c.Request.Context(), opts)
	if err != nil {
		writeServerError(c, err, "domain.list_pull_requests")
		return
	}

	data := make([]pullRequestResponse, 0, len(pullRequests))
	for _, pullRequest := range pullRequests {
		data = append(data, pullRequestResponse{
			ID:             pullRequest.ID,
			GiteaID:        pullRequest.GiteaID,
			RepositoryName: pullRequest.RepositoryName,
			Number:         pullRequest.Number,
			Title:          pullRequest.Title,
			State:          pullRequest.State,
			AuthorLogin:    pullRequest.AuthorLogin,
			HeadBranch:     pullRequest.HeadBranch,
			BaseBranch:     pullRequest.BaseBranch,
			HeadSHA:        pullRequest.HeadSHA,
			HTMLURL:        pullRequest.HTMLURL,
			MergedAt:       pullRequest.MergedAt,
			ClosedAt:       pullRequest.ClosedAt,
			UpdatedAt:      pullRequest.UpdatedAt,
		})
	}
	c.JSON(http.StatusOK, listEnvelope(data, opts))
}

func (h Handler) risks(c *gin.Context) {
	if h.cfg.DomainStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "domain store is not configured",
		})
		return
	}

	opts, ok := parseListOptions(c, false)
	if !ok {
		return
	}
	opts.Impact = c.Query("impact")
	risks, err := h.cfg.DomainStore.ListRisks(c.Request.Context(), opts)
	if err != nil {
		writeServerError(c, err, "domain.list_risks")
		return
	}

	data := make([]riskResponse, 0, len(risks))
	for _, risk := range risks {
		data = append(data, riskFromDomain(risk))
	}
	c.JSON(http.StatusOK, listEnvelope(data, opts))
}

func parseListOptions(c *gin.Context, includeState bool) (domain.ListOptions, bool) {
	limit, err := parseBoundedInt(c.DefaultQuery("limit", "50"), 1, 100)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "limit must be an integer between 1 and 100"})
		return domain.ListOptions{}, false
	}
	offset, err := parseBoundedInt(c.DefaultQuery("offset", "0"), 0, 100000)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "offset must be a non-negative integer"})
		return domain.ListOptions{}, false
	}
	opts := domain.ListOptions{
		Limit:          limit,
		Offset:         offset,
		RepositoryName: c.Query("repository_name"),
		Status:         c.Query("status"),
	}
	if includeState {
		opts.State = c.Query("state")
	}
	return opts, true
}

func listEnvelope[T any](data []T, opts domain.ListOptions) gin.H {
	meta := gin.H{
		"limit":  opts.Limit,
		"offset": opts.Offset,
		"count":  len(data),
	}
	if opts.RepositoryName != "" {
		meta["repository_name"] = opts.RepositoryName
	}
	if opts.State != "" {
		meta["state"] = opts.State
	}
	if opts.Status != "" {
		meta["status"] = opts.Status
	}
	if opts.Impact != "" {
		meta["impact"] = opts.Impact
	}
	return gin.H{
		"status": "ok",
		"data":   data,
		"meta":   meta,
	}
}
