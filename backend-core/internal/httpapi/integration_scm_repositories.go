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
// and the `pull` capability is enabled. Writes the error response and returns
// ok=false on failure.
func (h *Handler) scmProviderForPull(c *gin.Context, storeI ApplicationStore) (domain.IntegrationProvider, bool) {
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
	if provider.ProviderType != domain.IntegrationProviderTypeSCM {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": "rejected",
			"error":  "repository import is only supported for SCM-type providers",
			"code":   "integration_sync_unsupported_provider_type",
		})
		return domain.IntegrationProvider{}, false
	}
	if !providerHasCapability(provider, "pull") {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": "rejected",
			"error":  "provider does not have the 'pull' capability enabled",
			"code":   "integration_capability_not_enabled",
		})
		return domain.IntegrationProvider{}, false
	}
	return provider, true
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
