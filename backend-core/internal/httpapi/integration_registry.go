package httpapi

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/gitea"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

var validIntegrationProviderTypes = map[string]bool{
	"alm": true, "scm": true, "ci_cd": true, "doc": true, "infra": true,
}

var validIntegrationAuthModes = map[string]bool{
	"token": true, "basic": true, "oauth2": true, "app_password": true, "agent": true,
}

func integrationProviderResponse(p domain.IntegrationProvider) gin.H {
	return gin.H{
		"provider_id":     p.ID,
		"provider_key":    p.ProviderKey,
		"provider_type":   string(p.ProviderType),
		"display_name":    p.DisplayName,
		"enabled":         p.Enabled,
		"auth_mode":       string(p.AuthMode),
		"credentials_ref": p.CredentialsRef,
		"capabilities":    p.Capabilities,
		"sync_status":     p.SyncStatus,
		"last_sync_at":    nullableRFC3339(p.LastSyncAt),
		"last_error_code": emptyAsNil(p.LastErrorCode),
		"base_url":        emptyAsNil(p.BaseURL),
		"api_token_set":   p.APIToken != "", // write-only — raw token 미노출 (보안)
		// 구조화 outbound auth 자격증명 (auth_mode 별). 비밀 외 필드는 노출,
		// auth_secret 는 write-only (set 여부 bool 만).
		"auth_username":   emptyAsNil(p.AuthUsername),
		"auth_client_id":  emptyAsNil(p.AuthClientID),
		"auth_token_url":  emptyAsNil(p.AuthTokenURL),
		"auth_secret_set": p.AuthSecret != "",
		"created_at":      p.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at":      p.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func integrationBindingResponse(b domain.IntegrationBinding) gin.H {
	return gin.H{
		"binding_id":   b.ID,
		"scope_type":   string(b.ScopeType),
		"scope_id":     b.ScopeID,
		"provider_id":  b.ProviderID,
		"external_key": b.ExternalKey,
		"policy":       string(b.Policy),
		"enabled":      b.Enabled,
		"created_at":   b.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at":   b.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func nullableRFC3339(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.UTC().Format(time.RFC3339)
}

func emptyAsNil(v string) any {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return v
}

func parseOptionalBool(raw string) (*bool, bool) {
	if strings.TrimSpace(raw) == "" {
		return nil, true
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return nil, false
	}
	return &v, true
}

func verifyIntegrationWebhookSignature(provider domain.IntegrationProvider, body []byte, signature string) bool {
	signature = strings.TrimSpace(signature)
	if signature == "" {
		return false
	}
	credentials := strings.TrimSpace(provider.CredentialsRef)
	if credentials == "" {
		return false
	}
	// verifier strategy:
	// - hmac_sha256:<secret> => generic HMAC signature verification
	// - provider_sdk:<provider>[:<secret>] => provider-bound verifier routing
	// - otherwise => shared token constant-time compare
	if strings.HasPrefix(credentials, "hmac_sha256:") {
		secret := strings.TrimPrefix(credentials, "hmac_sha256:")
		return gitea.VerifySignature(body, secret, signature)
	}
	if strings.HasPrefix(credentials, "provider_sdk:") {
		return verifyProviderSDKWebhookSignature(provider, body, signature, strings.TrimPrefix(credentials, "provider_sdk:"))
	}
	if subtle.ConstantTimeCompare([]byte(signature), []byte(credentials)) == 1 {
		return true
	}
	if strings.HasPrefix(signature, "Bearer ") {
		token := strings.TrimSpace(strings.TrimPrefix(signature, "Bearer "))
		return subtle.ConstantTimeCompare([]byte(token), []byte(credentials)) == 1
	}
	return false
}

func (h *Handler) updateProviderSyncStateBestEffort(c *gin.Context, provider domain.IntegrationProvider, status, errorCode string) {
	storeI, ok := h.applicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	updated := provider
	updated.SyncStatus = status
	updated.LastErrorCode = strings.TrimSpace(errorCode)
	now := time.Now().UTC()
	updated.LastSyncAt = &now
	_, _ = storeI.UpdateIntegrationProvider(c.Request.Context(), updated)
}

func verifyProviderSDKWebhookSignature(provider domain.IntegrationProvider, body []byte, signature, strategy string) bool {
	parts := strings.SplitN(strings.TrimSpace(strategy), ":", 2)
	providerKey := strings.TrimSpace(provider.ProviderKey)
	secret := ""
	if len(parts) > 0 && strings.TrimSpace(parts[0]) != "" {
		providerKey = strings.TrimSpace(parts[0])
	}
	if len(parts) == 2 {
		secret = strings.TrimSpace(parts[1])
	}
	if secret == "" {
		secret = strings.TrimSpace(provider.CredentialsRef)
	}
	if strings.HasPrefix(secret, "provider_sdk:") {
		secretParts := strings.SplitN(strings.TrimPrefix(secret, "provider_sdk:"), ":", 2)
		if len(secretParts) == 2 {
			secret = strings.TrimSpace(secretParts[1])
		}
	}
	if secret == "" {
		return false
	}
	switch normalizeProviderSDKKey(providerKey) {
	case "gitea", "forgejo", "github", "gitlab", "jira", "bitbucket", "jenkins", "bamboo":
		return gitea.VerifySignature(body, secret, signature)
	default:
		return false
	}
}

func normalizeProviderSDKKey(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	if i := strings.Index(v, "-"); i > 0 {
		return v[:i]
	}
	return v
}

// API-69
func (h *Handler) listIntegrationProviders(c *gin.Context) {
	storeI, ok := h.applicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	opts := store.IntegrationProviderListOptions{
		ProviderType: domain.IntegrationProviderType(c.Query("provider_type")),
	}
	if string(opts.ProviderType) != "" && !validIntegrationProviderTypes[string(opts.ProviderType)] {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "provider_type must be alm|scm|ci_cd|doc|infra"})
		return
	}
	enabled, ok := parseOptionalBool(c.Query("enabled"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "enabled must be boolean"})
		return
	}
	opts.Enabled = enabled
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
	providers, total, err := storeI.ListIntegrationProviders(c.Request.Context(), opts)
	if err != nil {
		writeServerError(c, err, "integration.providers.list")
		return
	}
	resp := make([]gin.H, 0, len(providers))
	for _, p := range providers {
		resp = append(resp, integrationProviderResponse(p))
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "data": resp, "meta": gin.H{"total": total}})
}

type createIntegrationProviderRequest struct {
	ProviderKey    string   `json:"provider_key"`
	ProviderType   string   `json:"provider_type"`
	DisplayName    string   `json:"display_name"`
	AuthMode       string   `json:"auth_mode"`
	CredentialsRef string   `json:"credentials_ref"`
	Capabilities   []string `json:"capabilities"`
	BaseURL        string   `json:"base_url"`
	APIToken       string   `json:"api_token"`
	// 구조화 outbound auth 자격증명 (auth_mode 별). 모두 optional — sync/pull
	// capability 를 쓰는 provider 만 필요하므로 frontend 가 mode 별로 가이드한다.
	AuthUsername string `json:"auth_username"`
	AuthClientID string `json:"auth_client_id"`
	AuthTokenURL string `json:"auth_token_url"`
	AuthSecret   string `json:"auth_secret"`
}

// validBaseURL — base_url 은 optional (webhook 전용 provider 는 미사용). 제공 시
// http(s) scheme + non-empty host 를 가진 absolute URL 만 허용 (등록 UX 고도화 #2).
// codex review PR #352 P2: scheme-only ("https://") 같은 host 누락 값 거부.
func validBaseURL(raw string) bool {
	v := strings.TrimSpace(raw)
	if v == "" {
		return true
	}
	u, err := url.Parse(v)
	if err != nil {
		return false
	}
	return (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

// API-70
func (h *Handler) createIntegrationProvider(c *gin.Context) {
	storeI, ok := h.applicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	var req createIntegrationProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": err.Error()})
		return
	}
	if strings.TrimSpace(req.ProviderKey) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "provider_key is required"})
		return
	}
	if !validIntegrationProviderTypes[req.ProviderType] {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "invalid provider_type", "code": "invalid_provider_type"})
		return
	}
	if !validIntegrationAuthModes[req.AuthMode] {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "invalid auth_mode"})
		return
	}
	if strings.TrimSpace(req.DisplayName) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "display_name is required"})
		return
	}
	if strings.TrimSpace(req.CredentialsRef) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "credentials_ref is required"})
		return
	}
	if !validBaseURL(req.BaseURL) {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "base_url must be an http(s) URL", "code": "invalid_base_url"})
		return
	}
	if !validBaseURL(req.AuthTokenURL) {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "auth_token_url must be an http(s) URL", "code": "invalid_auth_token_url"})
		return
	}
	created, err := storeI.CreateIntegrationProvider(c.Request.Context(), domain.IntegrationProvider{
		ProviderKey:    req.ProviderKey,
		ProviderType:   domain.IntegrationProviderType(req.ProviderType),
		DisplayName:    req.DisplayName,
		Enabled:        true,
		AuthMode:       domain.IntegrationAuthMode(req.AuthMode),
		CredentialsRef: req.CredentialsRef,
		Capabilities:   req.Capabilities,
		SyncStatus:     "requested",
		BaseURL:        strings.TrimSpace(req.BaseURL),
		APIToken:       strings.TrimSpace(req.APIToken),
		AuthUsername:   strings.TrimSpace(req.AuthUsername),
		AuthClientID:   strings.TrimSpace(req.AuthClientID),
		AuthTokenURL:   strings.TrimSpace(req.AuthTokenURL),
		AuthSecret:     strings.TrimSpace(req.AuthSecret),
	})
	if errors.Is(err, store.ErrConflict) {
		c.JSON(http.StatusConflict, gin.H{"status": "conflict", "error": "provider already exists", "code": "integration_provider_conflict"})
		return
	}
	if err != nil {
		writeServerError(c, err, "integration.providers.create")
		return
	}
	h.recordAuditBestEffort(c, "integration.provider.created", "integration_provider", created.ID, map[string]any{
		"provider_key":  created.ProviderKey,
		"provider_type": string(created.ProviderType),
	})
	c.JSON(http.StatusCreated, gin.H{"status": "created", "data": integrationProviderResponse(created)})
}

type updateIntegrationProviderRequest struct {
	Enabled        *bool    `json:"enabled"`
	DisplayName    *string  `json:"display_name"`
	CredentialsRef *string  `json:"credentials_ref"`
	Capabilities   []string `json:"capabilities"`
	BaseURL        *string  `json:"base_url"`
	APIToken       *string  `json:"api_token"`
	// 구조화 outbound auth 자격증명. nil = 유지 (write-only secret 는 blank=keep).
	AuthUsername *string `json:"auth_username"`
	AuthClientID *string `json:"auth_client_id"`
	AuthTokenURL *string `json:"auth_token_url"`
	AuthSecret   *string `json:"auth_secret"`
}

// API-71
func (h *Handler) updateIntegrationProvider(c *gin.Context) {
	storeI, ok := h.applicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	providerID := c.Param("provider_id")
	var req updateIntegrationProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": err.Error()})
		return
	}
	current, err := storeI.GetIntegrationProviderByID(c.Request.Context(), providerID)
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "provider not found", "code": "integration_provider_not_found"})
		return
	}
	if err != nil {
		writeServerError(c, err, "integration.providers.update.lookup")
		return
	}
	updated := current
	if req.Enabled != nil {
		updated.Enabled = *req.Enabled
	}
	if req.DisplayName != nil {
		if strings.TrimSpace(*req.DisplayName) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "display_name cannot be empty"})
			return
		}
		updated.DisplayName = *req.DisplayName
	}
	if req.CredentialsRef != nil {
		if strings.TrimSpace(*req.CredentialsRef) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "credentials_ref cannot be empty"})
			return
		}
		updated.CredentialsRef = *req.CredentialsRef
	}
	if req.Capabilities != nil {
		updated.Capabilities = req.Capabilities
	}
	if req.BaseURL != nil {
		if !validBaseURL(*req.BaseURL) {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "base_url must be an http(s) URL", "code": "invalid_base_url"})
			return
		}
		updated.BaseURL = strings.TrimSpace(*req.BaseURL)
	}
	if req.APIToken != nil {
		updated.APIToken = strings.TrimSpace(*req.APIToken)
	}
	if req.AuthUsername != nil {
		updated.AuthUsername = strings.TrimSpace(*req.AuthUsername)
	}
	if req.AuthClientID != nil {
		updated.AuthClientID = strings.TrimSpace(*req.AuthClientID)
	}
	if req.AuthTokenURL != nil {
		if !validBaseURL(*req.AuthTokenURL) {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "auth_token_url must be an http(s) URL", "code": "invalid_auth_token_url"})
			return
		}
		updated.AuthTokenURL = strings.TrimSpace(*req.AuthTokenURL)
	}
	if req.AuthSecret != nil {
		updated.AuthSecret = strings.TrimSpace(*req.AuthSecret)
	}
	result, err := storeI.UpdateIntegrationProvider(c.Request.Context(), updated)
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "provider not found", "code": "integration_provider_not_found"})
		return
	}
	if err != nil {
		writeServerError(c, err, "integration.providers.update")
		return
	}
	h.recordAuditBestEffort(c, "integration.provider.updated", "integration_provider", result.ID, nil)
	c.JSON(http.StatusOK, gin.H{"status": "ok", "data": integrationProviderResponse(result)})
}

type testConnectionRequest struct {
	BaseURL string `json:"base_url"`
}

// POST /api/v1/integration/test-connection — 등록 UX 고도화 #5.
// 등록 전/후 외부 시스템 endpoint reachability 검증. system_admin gated
// (ResourceInfrastructure/Edit). reachability 만 확인 (자격증명 검증은 후속) —
// GET + 5s timeout + redirect 미추적.
//
// SSRF: 합법적 sync 대상이 사내 internal endpoint (Gitea/Jenkins 등) 이므로
// internal IP 차단은 하지 않는다. admin 신뢰 경계 + 짧은 timeout + 응답 본문
// 미반환 (status_code/latency 만) 으로 노출 표면을 최소화한다.
func (h *Handler) testIntegrationConnection(c *gin.Context) {
	var req testConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": err.Error()})
		return
	}
	target := strings.TrimSpace(req.BaseURL)
	if target == "" || !validBaseURL(target) {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "base_url must be a non-empty http(s) URL", "code": "invalid_base_url"})
		return
	}
	httpReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, target, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "invalid base_url", "code": "invalid_base_url"})
		return
	}
	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse // redirect 미추적 (SSRF chain 회피)
		},
	}
	start := time.Now()
	resp, err := client.Do(httpReq)
	latencyMs := time.Since(start).Milliseconds()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "reachable": false, "latency_ms": latencyMs, "error": err.Error()})
		return
	}
	defer resp.Body.Close()
	c.JSON(http.StatusOK, gin.H{"status": "ok", "reachable": true, "status_code": resp.StatusCode, "latency_ms": latencyMs})
}

// API-72
func (h *Handler) syncIntegrationProvider(c *gin.Context) {
	storeI, ok := h.applicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	providerID := c.Param("provider_id")
	provider, err := storeI.GetIntegrationProviderByID(c.Request.Context(), providerID)
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "provider not found", "code": "integration_provider_not_found"})
		return
	}
	if err != nil {
		writeServerError(c, err, "integration.providers.sync.lookup")
		return
	}
	if !provider.Enabled {
		c.JSON(http.StatusConflict, gin.H{"status": "rejected", "error": "provider is disabled", "code": "integration_provider_disabled"})
		return
	}
	// 현재 sync worker 는 SCM (Gitea) 한 종류뿐이다 (AcquireNextQueuedSyncJob 가
	// provider_type='scm' 게이트). 비-SCM provider 의 sync job 을 queue 하면 소비할
	// worker 가 없어 영구히 queued 로 남으므로, queue 전에 fast-fail 한다
	// (codex review PR #345 P2). 다른 provider_type 의 sync worker 추가 시 확장.
	if provider.ProviderType != domain.IntegrationProviderTypeSCM {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": "rejected",
			"error":  "sync is only supported for SCM-type providers",
			"code":   "integration_sync_unsupported_provider_type",
		})
		return
	}
	// capability gate (기능 gate 로 전환) — mirror sync 는 pull/sync 권한을 선언한
	// provider 만. 둘 다 없으면 422.
	if !providerHasCapability(provider, "pull", "sync") {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": "rejected",
			"error":  "provider does not have the 'pull' or 'sync' capability enabled",
			"code":   "integration_capability_not_enabled",
		})
		return
	}
	jobID, err := storeI.CreateIntegrationSyncJob(c.Request.Context(), providerID, actorLogin(c))
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "provider not found", "code": "integration_provider_not_found"})
		return
	}
	if err != nil {
		writeServerError(c, err, "integration.providers.sync")
		return
	}
	h.recordAuditBestEffort(c, "integration.provider.sync_requested", "integration_provider", providerID, map[string]any{"job_id": jobID})
	c.JSON(http.StatusAccepted, gin.H{"status": "accepted", "job_id": jobID})
}

// API-80 DELETE /api/v1/integration/providers/:provider_id — provider 삭제.
// FK guard: integration_bindings 가 1건 이상이면 409 `integration_provider_has_bindings`
// (실수 cascade 방지 — binding 정리는 운영자가 명시적으로 수행).
// integration_sync_jobs 는 schema 의 ON DELETE CASCADE 로 자동 정리.
// sprint claude/work_260518-j.
func (h *Handler) deleteIntegrationProvider(c *gin.Context) {
	storeI, ok := h.applicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	providerID := strings.TrimSpace(c.Param("provider_id"))
	if providerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "provider_id is required"})
		return
	}
	// 삭제 전 audit 본문에 담을 정보를 조회. not_found 면 404 직진.
	provider, err := storeI.GetIntegrationProviderByID(c.Request.Context(), providerID)
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"status": "not_found",
			"error":  "provider not found",
			"code":   "integration_provider_not_found",
		})
		return
	}
	if err != nil {
		writeServerError(c, err, "integration.providers.delete.lookup")
		return
	}
	if err := storeI.DeleteIntegrationProvider(c.Request.Context(), providerID); err != nil {
		if errors.Is(err, store.ErrConflict) {
			c.JSON(http.StatusConflict, gin.H{
				"status": "conflict",
				"error":  "provider has active bindings; remove bindings first",
				"code":   "integration_provider_has_bindings",
			})
			return
		}
		if errors.Is(err, store.ErrNotFound) {
			// 사전 조회 후 race 로 사라진 경우 — 404 로 정합.
			c.JSON(http.StatusNotFound, gin.H{
				"status": "not_found",
				"error":  "provider not found",
				"code":   "integration_provider_not_found",
			})
			return
		}
		writeServerError(c, err, "integration.providers.delete")
		return
	}
	h.recordAuditBestEffort(c, "integration.provider.deleted", "integration_provider", providerID, map[string]any{
		"provider_key":  provider.ProviderKey,
		"provider_type": string(provider.ProviderType),
		"display_name":  provider.DisplayName,
	})
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// API-73 minimal ingest (phase-1): provider existence + dedupe + accepted envelope.
func (h *Handler) ingestIntegrationProviderWebhook(c *gin.Context) {
	storeI, ok := h.applicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	providerKey := c.Param("provider_id")
	provider, err := storeI.GetIntegrationProviderByKey(c.Request.Context(), providerKey)
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "provider not found", "code": "integration_provider_not_found"})
		return
	}
	if err != nil {
		writeServerError(c, err, "integration.providers.webhook.lookup")
		return
	}
	if !provider.Enabled {
		c.JSON(http.StatusConflict, gin.H{"status": "rejected", "error": "provider is disabled", "code": "integration_provider_disabled"})
		return
	}
	// Accept DevHub-native header (X-Integration-*) plus provider-native aliases.
	// Gitea/Forgejo send X-Gitea-Signature, Gogs sends X-Gogs-Signature; the
	// signature value itself is provider-agnostic HMAC verified below.
	signature := strings.TrimSpace(firstHeader(c, "X-Integration-Signature", "X-Gitea-Signature", "X-Gogs-Signature"))
	payload, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "cannot read payload"})
		return
	}
	if !verifyIntegrationWebhookSignature(provider, payload, signature) {
		h.updateProviderSyncStateBestEffort(c, provider, "degraded", string(domain.SyncErrorWebhookSignatureInvalid))
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "rejected",
			"error":  "invalid webhook signature",
			"code":   "integration_webhook_signature_invalid",
		})
		return
	}
	deliveryID := strings.TrimSpace(firstHeader(c, "X-Integration-Delivery", "X-Gitea-Delivery", "X-Gogs-Delivery"))
	eventType := strings.TrimSpace(firstHeader(c, "X-Integration-Event", "X-Gitea-Event", "X-Gogs-Event"))
	if eventType == "" {
		eventType = "unknown"
	}
	hash := sha256.Sum256(payload)
	dedupeKey := "integration:" + providerKey + ":" + deliveryID
	if deliveryID == "" {
		dedupeKey = "integration:" + providerKey + ":payload:" + hex.EncodeToString(hash[:])
	}
	if h.cfg.EventStore != nil {
		_, saveErr := h.cfg.EventStore.SaveWebhookEvent(c.Request.Context(), store.WebhookEvent{
			EventType:  eventType,
			DeliveryID: deliveryID,
			DedupeKey:  dedupeKey,
			Payload:    payload,
			Status:     "received",
			ReceivedAt: time.Now().UTC(),
		})
		if errors.Is(saveErr, store.ErrDuplicateEvent) {
			c.JSON(http.StatusConflict, gin.H{"status": "conflict", "error": "duplicate delivery", "code": "integration_event_duplicate"})
			return
		}
		if saveErr != nil {
			writeServerError(c, saveErr, "integration.providers.webhook.save")
			return
		}
	}
	h.updateProviderSyncStateBestEffort(c, provider, "active", "")
	h.recordAuditBestEffort(c, "integration.provider.webhook_ingested", "integration_provider", provider.ID, map[string]any{
		"provider_key": providerKey,
		"event_type":   eventType,
		"delivery_id":  deliveryID,
	})
	c.JSON(http.StatusAccepted, gin.H{"status": "accepted"})
}

// API-74
func (h *Handler) listIntegrationBindings(c *gin.Context) {
	storeI, ok := h.applicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	opts := store.IntegrationBindingListOptions{
		ScopeType:    domain.IntegrationScopeType(c.Query("scope_type")),
		ScopeID:      c.Query("scope_id"),
		ProviderType: domain.IntegrationProviderType(c.Query("provider_type")),
	}
	if string(opts.ScopeType) != "" && opts.ScopeType != domain.IntegrationScopeTypeApplication && opts.ScopeType != domain.IntegrationScopeTypeProject {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "scope_type must be application or project"})
		return
	}
	if string(opts.ProviderType) != "" && !validIntegrationProviderTypes[string(opts.ProviderType)] {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "provider_type must be alm|scm|ci_cd|doc|infra"})
		return
	}
	enabled, ok := parseOptionalBool(c.Query("enabled"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "enabled must be boolean"})
		return
	}
	opts.Enabled = enabled
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
	bindings, total, err := storeI.ListIntegrationBindings(c.Request.Context(), opts)
	if err != nil {
		writeServerError(c, err, "integration.bindings.list")
		return
	}
	resp := make([]gin.H, 0, len(bindings))
	for _, b := range bindings {
		resp = append(resp, integrationBindingResponse(b))
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "data": resp, "meta": gin.H{"total": total}})
}

type createIntegrationBindingRequest struct {
	ScopeType   string `json:"scope_type"`
	ScopeID     string `json:"scope_id"`
	ProviderID  string `json:"provider_id"`
	ExternalKey string `json:"external_key"`
	Policy      string `json:"policy"`
}

// API-75
func (h *Handler) createIntegrationBinding(c *gin.Context) {
	storeI, ok := h.applicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	var req createIntegrationBindingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": err.Error()})
		return
	}
	if req.ScopeType != "application" && req.ScopeType != "project" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "scope_type must be application or project"})
		return
	}
	if strings.TrimSpace(req.ScopeID) == "" || strings.TrimSpace(req.ProviderID) == "" || strings.TrimSpace(req.ExternalKey) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "scope_id/provider_id/external_key are required"})
		return
	}
	if req.Policy != string(domain.IntegrationPolicySummaryOnly) && req.Policy != string(domain.IntegrationPolicyExecutionSystem) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": "rejected", "error": "unsupported policy", "code": "integration_policy_violation"})
		return
	}
	created, err := storeI.CreateIntegrationBinding(c.Request.Context(), domain.IntegrationBinding{
		ScopeType:   domain.IntegrationScopeType(req.ScopeType),
		ScopeID:     req.ScopeID,
		ProviderID:  req.ProviderID,
		ExternalKey: req.ExternalKey,
		Policy:      domain.IntegrationPolicy(req.Policy),
		Enabled:     true,
	})
	if errors.Is(err, store.ErrConflict) {
		c.JSON(http.StatusConflict, gin.H{"status": "conflict", "error": "binding already exists or provider not found", "code": "integration_binding_conflict"})
		return
	}
	if err != nil {
		writeServerError(c, err, "integration.bindings.create")
		return
	}
	h.recordAuditBestEffort(c, "integration.binding.created", "integration_binding", created.ID, nil)
	c.JSON(http.StatusCreated, gin.H{"status": "created", "data": integrationBindingResponse(created)})
}

type updateIntegrationBindingRequest struct {
	ProviderID  *string `json:"provider_id"`
	ExternalKey *string `json:"external_key"`
	Policy      *string `json:"policy"`
	Enabled     *bool   `json:"enabled"`
}

func (h *Handler) updateIntegrationBinding(c *gin.Context) {
	storeI, ok := h.applicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	bindingID := c.Param("binding_id")
	var req updateIntegrationBindingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": err.Error()})
		return
	}

	current, err := storeI.GetIntegrationBindingByID(c.Request.Context(), bindingID)
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "binding not found"})
		return
	}
	if err != nil {
		writeServerError(c, err, "integration.bindings.update.lookup")
		return
	}

	updated := current
	if req.ProviderID != nil {
		updated.ProviderID = *req.ProviderID
	}
	if req.ExternalKey != nil {
		updated.ExternalKey = *req.ExternalKey
	}
	if req.Policy != nil {
		if *req.Policy != string(domain.IntegrationPolicySummaryOnly) && *req.Policy != string(domain.IntegrationPolicyExecutionSystem) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"status": "rejected", "error": "unsupported policy"})
			return
		}
		updated.Policy = domain.IntegrationPolicy(*req.Policy)
	}
	if req.Enabled != nil {
		updated.Enabled = *req.Enabled
	}

	result, err := storeI.UpdateIntegrationBinding(c.Request.Context(), updated)
	if errors.Is(err, store.ErrConflict) {
		c.JSON(http.StatusConflict, gin.H{"status": "conflict", "error": "binding already exists or provider not found"})
		return
	}
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "binding or provider not found"})
		return
	}
	if err != nil {
		writeServerError(c, err, "integration.bindings.update")
		return
	}

	h.recordAuditBestEffort(c, "integration.binding.updated", "integration_binding", result.ID, nil)
	c.JSON(http.StatusOK, gin.H{"status": "ok", "data": integrationBindingResponse(result)})
}

func (h *Handler) deleteIntegrationBinding(c *gin.Context) {
	storeI, ok := h.applicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	bindingID := c.Param("binding_id")
	if err := storeI.DeleteIntegrationBinding(c.Request.Context(), bindingID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "binding not found"})
			return
		}
		writeServerError(c, err, "integration.bindings.delete")
		return
	}
	h.recordAuditBestEffort(c, "integration.binding.deleted", "integration_binding", bindingID, nil)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
