package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/gitea"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

// providerHasCapability reports whether the provider declares any of the given
// capabilities. capabilities 는 기능 gate 로 사용된다 (sprint scm-repo-sync):
// pull = SCM 으로부터 repo 조회/import/mirror 허용, sync = 주기 mirror 허용,
// push = outbound 생성 허용(Phase C), webhook = inbound webhook 수신.
func providerHasCapability(p domain.IntegrationProvider, caps ...string) bool {
	for _, have := range p.Capabilities {
		for _, want := range caps {
			if have == want {
				return true
			}
		}
	}
	return false
}

// scmProviderForPull resolves the :provider_id provider and enforces the gates
// shared by SCM repository read operations: provider exists, provider_type=scm,
// and the required capability is enabled. Writes the error response and returns
// ok=false on failure.
func (h *Handler) scmProviderForCapability(c *gin.Context, storeI ApplicationStore, capability string) (domain.IntegrationProvider, bool) {
	providerID := c.Param("provider_id")
	provider, err := storeI.GetIntegrationProviderByID(c.Request.Context(), providerID)
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "provider not found", "code": "integration_provider_not_found"})
		return domain.IntegrationProvider{}, false
	}
	if err != nil {
		writeServerError(c, err, "integration.scm.provider.lookup")
		return domain.IntegrationProvider{}, false
	}
	// 비활성 provider 는 거부 (codex #366 P2) — sync/webhook 핸들러와 동일 정책.
	if !provider.Enabled {
		c.JSON(http.StatusConflict, gin.H{
			"status": "rejected",
			"error":  "provider is disabled",
			"code":   "integration_provider_disabled",
		})
		return domain.IntegrationProvider{}, false
	}
	if provider.ProviderType != domain.IntegrationProviderTypeSCM {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": "rejected",
			"error":  "repository operations are only supported for SCM-type providers",
			"code":   "integration_sync_unsupported_provider_type",
		})
		return domain.IntegrationProvider{}, false
	}
	if !providerHasCapability(provider, capability) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": "rejected",
			"error":  "provider does not have the '" + capability + "' capability enabled",
			"code":   "integration_capability_not_enabled",
		})
		return domain.IntegrationProvider{}, false
	}
	if !isGiteaCompatibleProvider(provider) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": "rejected",
			"error":  "repository sync currently supports Gitea-compatible providers (gitea/forgejo/gogs) only",
			"code":   "integration_provider_not_gitea_compatible",
		})
		return domain.IntegrationProvider{}, false
	}
	return provider, true
}

// scmProviderForPull — inbound 조회/import gate (pull capability).
func (h *Handler) scmProviderForPull(c *gin.Context, storeI ApplicationStore) (domain.IntegrationProvider, bool) {
	return h.scmProviderForCapability(c, storeI, "pull")
}

// scmProviderClient builds an authenticated Gitea client from the provider's
// base_url + outbound auth. Writes the error response and returns ok=false on
// missing base_url, auth failure, or unconfigured credentials.
func (h *Handler) scmProviderClient(c *gin.Context, provider domain.IntegrationProvider) (*gitea.Client, bool) {
	if strings.TrimSpace(provider.BaseURL) == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": "rejected", "error": "provider base_url is not set", "code": "integration_base_url_missing"})
		return nil, false
	}
	client, err := gitea.NewClientForAuth(c.Request.Context(), provider.BaseURL, provider.ResolveOutboundAuth())
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"status": "rejected", "error": "SCM authentication failed: " + err.Error(), "code": "integration_scm_auth_failed"})
		return nil, false
	}
	if client == nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": "rejected", "error": "provider outbound credentials are not configured", "code": "integration_outbound_credentials_missing"})
		return nil, false
	}
	return client, true
}

// giteaCompatibleVendors — Gitea REST API 와 호환되는 vendor (현재 구현된 유일한 SCM
// 어댑터). import/create 는 이 client 를 쓰므로 다른 vendor(github/gitlab/bitbucket)는
// 거부한다 (codex #363 P2).
var giteaCompatibleVendors = map[string]bool{"gitea": true, "forgejo": true, "gogs": true}

// isGiteaCompatibleProvider — credentials_ref 의 provider_sdk vendor 가 명시돼 있고
// gitea-family 가 아니면 false. vendor 미명시(hmac_sha256/shared_token, 예: Custom)면
// 사용자가 Gitea-호환 base_url 을 지정했다고 보고 허용(true).
func isGiteaCompatibleProvider(p domain.IntegrationProvider) bool {
	const prefix = "provider_sdk:"
	ref := strings.TrimSpace(p.CredentialsRef)
	if !strings.HasPrefix(ref, prefix) {
		return true
	}
	parts := strings.SplitN(strings.TrimPrefix(ref, prefix), ":", 2)
	vendor := normalizeProviderSDKKey(parts[0])
	if vendor == "" {
		return true
	}
	return giteaCompatibleVendors[vendor]
}

func scmRepoOwnerLogin(fullName string) string {
	if i := strings.Index(fullName, "/"); i > 0 {
		return fullName[:i]
	}
	return ""
}

// API-88 — GET /api/v1/integration/providers/:provider_id/scm-repositories
// SCM(provider)으로부터 원격 repository 목록을 조회한다. 각 항목에 시스템 import 여부
// (imported)를 표시한다. provider_type=scm + pull capability 필요.
func (h *Handler) listSCMRepositories(c *gin.Context) {
	storeI, ok := h.applicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	provider, ok := h.scmProviderForPull(c, storeI)
	if !ok {
		return
	}
	client, ok := h.scmProviderClient(c, provider)
	if !ok {
		return
	}
	remote, err := client.ListUserRepos(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"status": "rejected", "error": "failed to list SCM repositories: " + err.Error(), "code": "integration_scm_unreachable"})
		return
	}
	importedRepos, err := storeI.ListRepositoriesByProvider(c.Request.Context(), provider.ID)
	if err != nil {
		writeServerError(c, err, "integration.scm.repositories.imported")
		return
	}
	importedSet := make(map[string]bool, len(importedRepos))
	for _, r := range importedRepos {
		importedSet[r.FullName] = true
	}
	data := make([]gin.H, 0, len(remote))
	for _, r := range remote {
		data = append(data, gin.H{
			"full_name":      r.FullName,
			"name":           r.Name,
			"clone_url":      r.CloneURL,
			"html_url":       r.HTMLURL,
			"default_branch": r.DefaultBranch,
			"private":        r.Private,
			"imported":       importedSet[r.FullName],
		})
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "data": data, "meta": gin.H{"total": len(data)}})
}

type importSCMRepositoriesRequest struct {
	FullNames []string `json:"full_names"`
}

// API-89 — POST /api/v1/integration/providers/:provider_id/import-repositories
// 선택한 원격 repository 들을 시스템 repositories 로 import/연동한다 (source=scm,
// provider_id 세팅, SCM mirror 필드 채움). 신뢰 가능한 SCM 데이터를 쓰기 위해 요청
// payload 가 아니라 SCM 에서 다시 조회한 값으로 upsert 한다. pull capability 필요.
func (h *Handler) importSCMRepositories(c *gin.Context) {
	storeI, ok := h.applicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	var req importSCMRepositoriesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": err.Error()})
		return
	}
	selected := make(map[string]bool)
	for _, fn := range req.FullNames {
		if t := strings.TrimSpace(fn); t != "" {
			selected[t] = true
		}
	}
	if len(selected) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "full_names is required", "code": "integration_import_no_selection"})
		return
	}
	provider, ok := h.scmProviderForPull(c, storeI)
	if !ok {
		return
	}
	client, ok := h.scmProviderClient(c, provider)
	if !ok {
		return
	}
	remote, err := client.ListUserRepos(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"status": "rejected", "error": "failed to list SCM repositories: " + err.Error(), "code": "integration_scm_unreachable"})
		return
	}
	byName := make(map[string]gitea.GiteaRepository, len(remote))
	for _, r := range remote {
		byName[r.FullName] = r
	}
	imported := make([]gin.H, 0, len(selected))
	notFound := make([]string, 0)
	for fn := range selected {
		r, found := byName[fn]
		if !found {
			notFound = append(notFound, fn)
			continue
		}
		if err := storeI.UpsertRepository(c.Request.Context(), domain.Repository{
			GiteaID:       r.ID,
			FullName:      r.FullName,
			OwnerLogin:    scmRepoOwnerLogin(r.FullName),
			Name:          r.Name,
			CloneURL:      r.CloneURL,
			HTMLURL:       r.HTMLURL,
			DefaultBranch: r.DefaultBranch,
			Private:       r.Private,
			Source:        domain.RepositorySourceSCM,
			ProviderID:    provider.ID,
		}); err != nil {
			writeServerError(c, err, "integration.scm.repositories.import")
			return
		}
		imported = append(imported, gin.H{"full_name": r.FullName, "name": r.Name})
	}
	h.recordAuditBestEffort(c, "integration.provider.repositories_imported", "integration_provider", provider.ID, map[string]any{
		"provider_key": provider.ProviderKey,
		"imported":     len(imported),
		"not_found":    len(notFound),
	})
	c.JSON(http.StatusOK, gin.H{
		"status":       "ok",
		"imported":     len(imported),
		"repositories": imported,
		"not_found":    notFound,
	})
}

type createSCMRepositoryRequest struct {
	Name        string `json:"name"`
	Owner       string `json:"owner"` // optional org; 빈 값이면 인증 사용자 계정
	Description string `json:"description"`
	Private     bool   `json:"private"`
	AutoInit    bool   `json:"auto_init"`
}

// API-90 — POST /api/v1/integration/providers/:provider_id/create-repository (Phase C)
// 시스템에서 선택 SCM(provider)에 실제 저장소를 생성하고 시스템 repositories 로 미러한다
// (source=system, provider_id 세팅 — 시스템이 생성을 주도했으므로 system-owned).
// push capability + Gitea-compatible provider 필요.
func (h *Handler) createSCMRepository(c *gin.Context) {
	storeI, ok := h.applicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	var req createSCMRepositoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": err.Error()})
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "name is required", "code": "integration_repo_name_required"})
		return
	}
	provider, ok := h.scmProviderForCapability(c, storeI, "push")
	if !ok {
		return
	}
	client, ok := h.scmProviderClient(c, provider)
	if !ok {
		return
	}
	created, err := client.CreateRepo(c.Request.Context(), strings.TrimSpace(req.Owner), gitea.CreateRepoOptions{
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		Private:     req.Private,
		AutoInit:    req.AutoInit,
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"status": "rejected", "error": "failed to create SCM repository: " + err.Error(), "code": "integration_scm_create_failed"})
		return
	}
	if err := storeI.UpsertRepository(c.Request.Context(), domain.Repository{
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
		Description:   strings.TrimSpace(req.Description),
	}); err != nil {
		writeServerError(c, err, "integration.scm.repositories.create.persist")
		return
	}
	h.recordAuditBestEffort(c, "integration.provider.repository_created", "integration_provider", provider.ID, map[string]any{
		"provider_key": provider.ProviderKey,
		"full_name":    created.FullName,
	})
	c.JSON(http.StatusCreated, gin.H{
		"status": "created",
		"repository": gin.H{
			"full_name":      created.FullName,
			"name":           created.Name,
			"clone_url":      created.CloneURL,
			"html_url":       created.HTMLURL,
			"default_branch": created.DefaultBranch,
			"private":        created.Private,
			"source":         domain.RepositorySourceSystem,
		},
	})
}
