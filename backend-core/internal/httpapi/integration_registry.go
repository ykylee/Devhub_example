package httpapi

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"net/http"
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
	// phase-1 verifier strategy:
	// - hmac_sha256:<secret> => HMAC signature verification
	// - otherwise => shared token constant-time compare
	if strings.HasPrefix(credentials, "hmac_sha256:") {
		secret := strings.TrimPrefix(credentials, "hmac_sha256:")
		return gitea.VerifySignature(body, secret, signature)
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
	created, err := storeI.CreateIntegrationProvider(c.Request.Context(), domain.IntegrationProvider{
		ProviderKey:    req.ProviderKey,
		ProviderType:   domain.IntegrationProviderType(req.ProviderType),
		DisplayName:    req.DisplayName,
		Enabled:        true,
		AuthMode:       domain.IntegrationAuthMode(req.AuthMode),
		CredentialsRef: req.CredentialsRef,
		Capabilities:   req.Capabilities,
		SyncStatus:     "requested",
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
	signature := strings.TrimSpace(c.GetHeader("X-Integration-Signature"))
	payload, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "cannot read payload"})
		return
	}
	if !verifyIntegrationWebhookSignature(provider, payload, signature) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "rejected",
			"error":  "invalid webhook signature",
			"code":   "integration_webhook_signature_invalid",
		})
		return
	}
	deliveryID := strings.TrimSpace(c.GetHeader("X-Integration-Delivery"))
	eventType := strings.TrimSpace(c.GetHeader("X-Integration-Event"))
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
