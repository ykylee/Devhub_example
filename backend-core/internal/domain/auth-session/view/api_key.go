package view

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/devhub/backend-core/internal/domain/auth-session/repository"
	"github.com/devhub/backend-core/internal/shared/authkey"
	"github.com/devhub/backend-core/internal/shared/httphelp"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

// APIKeyAdminStore 는 view handler 가 사용하는 read/write 인터페이스.
// repository.APIKeyStore 와 같은 셰이프이지만 cross-domain import cycle 회피를
// 위해 view-local 로 재선언. main.go wire 시 adapter 로 매핑. auth.go 의
// APIKeyStore (read-only) 와 분리 — handler 는 write (Create/Revoke/Update) 도
// 수행.
type APIKeyAdminStore interface {
	CreateAPIKey(ctx context.Context, keyHash []byte, key repository.APIKey) (repository.APIKey, error)
	ListAPIKeys(ctx context.Context) ([]repository.APIKey, error)
	RevokeAPIKey(ctx context.Context, id string, revokedBy string) error
	UpdateAPIKeyMeta(ctx context.Context, id string, update repository.APIKeyUpdateRequest) error
}

// apiKeyResponseFromRepo — repository.APIKey → response 매핑. raw key / key_hash
// 는 절대 매핑하지 않음 (보안 invariant).
func apiKeyResponseFromRepo(k repository.APIKey) apiKeySummaryResponse {
	return apiKeySummaryResponse{
		ID:           k.ID,
		Name:         k.Name,
		KeyPrefix:    k.KeyPrefix,
		CreatedBy:    k.CreatedBy,
		CreatedAt:    parseTime(k.CreatedAt),
		LastUsedAt:   k.LastUsedAt,
		RevokedAt:    k.RevokedAt,
		RevokedBy:    k.RevokedBy,
		ExpiresAt:    k.ExpiresAt,
		AllowedCIDRs: k.AllowedCIDRs,
		Status:       string(k.Status),
	}
}

// apiKeySummaryResponse — list GET 응답용 (raw key 미포함, key_prefix 만 노출).
// ADR-0029 §6 (f) P3 — 운영자 식별은 prefix 8자 (`dhk_aB3x...`) 로 충분. raw
// key material 은 POST 응답 1회만 노출. 이후 GET 은 절대 재노출 불가.
type apiKeySummaryResponse struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	KeyPrefix    string    `json:"key_prefix"`
	CreatedBy    string    `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
	LastUsedAt   *string   `json:"last_used_at,omitempty"`
	RevokedAt    *string   `json:"revoked_at,omitempty"`
	RevokedBy    *string   `json:"revoked_by,omitempty"`
	ExpiresAt    *string   `json:"expires_at,omitempty"`
	AllowedCIDRs []string  `json:"allowed_cidrs,omitempty"`
	Status       string    `json:"status"`
}

// apiKeyCreateRequest — POST body. raw key 는 client 가 절대 입력할 수 없는
// server-side 생성값.
type apiKeyCreateRequest struct {
	Name         string   `json:"name" binding:"required"`
	ExpiresAt    *string  `json:"expires_at,omitempty"`
	AllowedCIDRs []string `json:"allowed_cidrs,omitempty"`
}

// apiKeyCreateResponse — POST 응답. raw key 1회만 노출 — frontend 가 modal 에
// 표시하고 clipboard copy 후 재호출 시 절대 재노출 안 됨. 이후 GET 은 prefix 만.
type apiKeyCreateResponse struct {
	APIKey     apiKeySummaryResponse `json:"api_key"`
	RawKey     string                `json:"key"`
	KeyExpires *string               `json:"key_expires_at,omitempty"`
	Warning    string                `json:"warning"`
}

// apiKeyUpdateRequest — PATCH body. name 은 변경 불가 (의도적 — 운영자 문서
// invariant). expires_at 와 allowed_cidrs 만 갱신. null/empty = clear.
type apiKeyUpdateRequest struct {
	ExpiresAt    *string  `json:"expires_at"`
	AllowedCIDRs []string `json:"allowed_cidrs"`
}

// parseTime — repository.APIKey.CreatedAt (string) → time.Time. 실패 시 zero time
// (200 epoch) — frontend 가 invalid date 로 표시.
func parseTime(s string) time.Time {
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t
	}
	return time.Time{}
}

// validateCIDRList — list 의 각 entry 가 `net.ParseCIDR` 로 valid 한지 확인.
// invalid 가 하나라도 있으면 첫 invalid entry 의 error message 와 함께 400.
// auth middleware 의 IsCIDRAllowed 와 동일 검증 (allowlist 의미론 정합).
func validateCIDRList(cidrs []string) error {
	for _, c := range cidrs {
		if _, _, err := net.ParseCIDR(c); err != nil {
			return errors.New("invalid CIDR in allowed_cidrs: " + c)
		}
	}
	return nil
}

// validateExpiresAt — nil = unlimited. non-nil = RFC3339 + 미래 시각.
func validateExpiresAt(expiresAt *string) error {
	if expiresAt == nil {
		return nil
	}
	t, err := time.Parse(time.RFC3339, *expiresAt)
	if err != nil {
		return errors.New("expires_at must be RFC3339 (e.g. 2026-12-31T23:59:59Z)")
	}
	if !t.After(time.Now()) {
		return errors.New("expires_at must be in the future")
	}
	return nil
}

// actorLoginFromContext — request context 의 actor login 추출. dev fallback 의
// "system" 또는 빈 값이면 "system_admin_unknown" 으로 치환 (audit 정합).
func actorLoginFromContext(c *gin.Context) string {
	actor := httphelp.RequestActor(c)
	login := strings.TrimSpace(actor.Login)
	if login == "" || login == "system" {
		return "system_admin_unknown"
	}
	return login
}

// CreateAPIKey — POST /api/v1/admin/api-keys. ADR-0029 §6 (f) P3 + sprint plan
// §3.3. system_admin RBAC 일임 (routePermissionTable 정합). 응답에 raw key 1회
// 포함 — frontend modal 에서 clipboard copy 후 폐기. 서버는 hash 만 보관.
func (h *AuthHandler) CreateAPIKey(c *gin.Context) {
	if h.cfg.APIKeyStoreAdmin == nil {
		// multi-key store 미설정 = backend 가 multi-key 비활성 모드. 503.
		// 운영자가 DB store 추가 셋업 후 재시도 필요. nil pointer panic 방지.
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "api key store is not configured (DEVHUB_API_KEY_STORE_PG unset)",
		})
		return
	}
	var req apiKeyCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status": "bad_request",
			"error":  "invalid request body: name is required",
		})
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status": "bad_request",
			"error":  "name must not be empty",
		})
		return
	}
	if err := validateCIDRList(req.AllowedCIDRs); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status": "bad_request",
			"error":  err.Error(),
		})
		return
	}
	if err := validateExpiresAt(req.ExpiresAt); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status": "bad_request",
			"error":  err.Error(),
		})
		return
	}

	login := actorLoginFromContext(c)
	rawKey, keyHash, keyPrefix, err := authkey.GenerateRawAPIKey()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status": "internal_error",
			"error":  "failed to generate api key: " + err.Error(),
		})
		return
	}
	generated := repository.APIKey{
		Name:         name,
		KeyPrefix:    keyPrefix,
		CreatedBy:    login,
		AllowedCIDRs: req.AllowedCIDRs,
		ExpiresAt:    req.ExpiresAt,
		Status:       repository.APIKeyStatusActive,
	}

	persisted, err := h.cfg.APIKeyStoreAdmin.CreateAPIKey(c.Request.Context(), keyHash, generated)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status": "internal_error",
			"error":  "failed to persist api key: " + err.Error(),
		})
		return
	}

	// audit emit — sprint plan §3.8 audit.api_key.created. ADR-0029 §6 (g) 정합.
	h.recordAuditBestEffort(c, "auth.api_key.created", "api_key", persisted.ID, map[string]any{
		"actor":          login,
		"api_key_id":     persisted.ID,
		"key_prefix":     persisted.KeyPrefix,
		"name":           persisted.Name,
		"expires_at":     persisted.ExpiresAt,
		"allowed_cidrs":  persisted.AllowedCIDRs,
		"created_branch": "db_multi_key",
	})

	resp := apiKeyCreateResponse{
		APIKey:     apiKeyResponseFromRepo(persisted),
		RawKey:     rawKey,
		KeyExpires: persisted.ExpiresAt,
		Warning:    "이 키는 다시 볼 수 없습니다. 안전한 곳에 즉시 저장하세요.",
	}
	c.JSON(http.StatusCreated, resp)
}

// ListAPIKeys — GET /api/v1/admin/api-keys. 모든 키 metadata (key_prefix 만, raw
// key 미포함). newest first. routePermissionTable 정합 (system_admin 만).
func (h *AuthHandler) ListAPIKeys(c *gin.Context) {
	if h.cfg.APIKeyStoreAdmin == nil {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "api key store is not configured",
		})
		return
	}
	keys, err := h.cfg.APIKeyStoreAdmin.ListAPIKeys(c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status": "internal_error",
			"error":  "failed to list api keys: " + err.Error(),
		})
		return
	}
	resp := make([]apiKeySummaryResponse, 0, len(keys))
	for _, k := range keys {
		resp = append(resp, apiKeyResponseFromRepo(k))
	}
	c.JSON(http.StatusOK, gin.H{"api_keys": resp, "count": len(resp)})
}

// RevokeAPIKey — DELETE /api/v1/admin/api-keys/:api_key_id. ADR-0029 §6 (f) P3.
// idempotent — repository.RevokeAPIKey 가 이미 회수된 key 도 audit 만 갱신.
// sprint plan §3.3 정합.
func (h *AuthHandler) RevokeAPIKey(c *gin.Context) {
	if h.cfg.APIKeyStoreAdmin == nil {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "api key store is not configured",
		})
		return
	}
	id := strings.TrimSpace(c.Param("api_key_id"))
	if id == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status": "bad_request",
			"error":  "api_key_id is required",
		})
		return
	}
	login := actorLoginFromContext(c)

	err := h.cfg.APIKeyStoreAdmin.RevokeAPIKey(c.Request.Context(), id, login)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"status": "not_found",
				"error":  "api key not found",
			})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status": "internal_error",
			"error":  "failed to revoke api key: " + err.Error(),
		})
		return
	}

	// audit emit — auth.api_key.revoked. ADR-0029 §6 (g).
	h.recordAuditBestEffort(c, "auth.api_key.revoked", "api_key", id, map[string]any{
		"actor":          login,
		"api_key_id":     id,
		"revoked_by":     login,
		"created_branch": "db_multi_key",
	})

	c.Status(http.StatusNoContent)
	c.Writer.WriteHeaderNow()
}

// UpdateAPIKeyMeta — PATCH /api/v1/admin/api-keys/:api_key_id. expires_at /
// allowed_cidrs 갱신. revoked key 는 update 불가 (repository.UpdateAPIKeyMeta 가
// `revoked_at IS NULL` 조건으로 0 rows → store.ErrNotFound).
func (h *AuthHandler) UpdateAPIKeyMeta(c *gin.Context) {
	if h.cfg.APIKeyStoreAdmin == nil {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "api key store is not configured",
		})
		return
	}
	id := strings.TrimSpace(c.Param("api_key_id"))
	if id == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status": "bad_request",
			"error":  "api_key_id is required",
		})
		return
	}
	var req apiKeyUpdateRequest
	if c.Request.Body != nil {
		_ = json.NewDecoder(c.Request.Body).Decode(&req)
	}
	if err := validateCIDRList(req.AllowedCIDRs); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status": "bad_request",
			"error":  err.Error(),
		})
		return
	}
	if err := validateExpiresAt(req.ExpiresAt); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status": "bad_request",
			"error":  err.Error(),
		})
		return
	}
	login := actorLoginFromContext(c)

	update := repository.APIKeyUpdateRequest{
		ExpiresAt:    req.ExpiresAt,
		AllowedCIDRs: req.AllowedCIDRs,
	}
	err := h.cfg.APIKeyStoreAdmin.UpdateAPIKeyMeta(c.Request.Context(), id, update)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"status": "not_found",
				"error":  "api key not found or already revoked",
			})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status": "internal_error",
			"error":  "failed to update api key: " + err.Error(),
		})
		return
	}
	// audit emit — auth.api_key.updated. ADR-0029 §6 (g) 정합.
	h.recordAuditBestEffort(c, "auth.api_key.updated", "api_key", id, map[string]any{
		"actor":          login,
		"api_key_id":     id,
		"expires_at":     req.ExpiresAt,
		"allowed_cidrs":  req.AllowedCIDRs,
		"created_branch": "db_multi_key",
	})
	c.JSON(http.StatusOK, gin.H{"status": "updated", "api_key_id": id})
}
