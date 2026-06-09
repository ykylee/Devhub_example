package view

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/domain/auth-session/repository"
	"github.com/devhub/backend-core/internal/shared/authkey"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

// fakeAPIKeyAdminStore — view/APIKeyAdminStore 의 in-memory fake. repository
// 와 같은 셰이프 (cross-domain import cycle 회피). 생성/회수/갱신은 simple
// map 으로 모킹.
type fakeAPIKeyAdminStore struct {
	keys       map[string]repository.APIKey
	createErr  error
	listErr    error
	revokeErr  error
	updateErr  error
}

func newFakeAPIKeyAdminStore() *fakeAPIKeyAdminStore {
	return &fakeAPIKeyAdminStore{keys: make(map[string]repository.APIKey)}
}

func (f *fakeAPIKeyAdminStore) CreateAPIKey(_ context.Context, keyHash []byte, key repository.APIKey) (repository.APIKey, error) {
	if f.createErr != nil {
		return repository.APIKey{}, f.createErr
	}
	if key.ID == "" {
		key.ID = "test-uuid-1"
	}
	key.Status = repository.APIKeyStatusActive
	f.keys[key.ID] = key
	return key, nil
}

func (f *fakeAPIKeyAdminStore) ListAPIKeys(_ context.Context) ([]repository.APIKey, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	out := make([]repository.APIKey, 0, len(f.keys))
	for _, k := range f.keys {
		out = append(out, k)
	}
	return out, nil
}

func (f *fakeAPIKeyAdminStore) RevokeAPIKey(_ context.Context, id, revokedBy string) error {
	if f.revokeErr != nil {
		return f.revokeErr
	}
	if _, ok := f.keys[id]; !ok {
		return store.ErrNotFound
	}
	k := f.keys[id]
	now := time.Now().UTC().Format(time.RFC3339)
	k.RevokedAt = &now
	k.RevokedBy = &revokedBy
	k.Status = repository.APIKeyStatusRevoked
	f.keys[id] = k
	return nil
}

func (f *fakeAPIKeyAdminStore) UpdateAPIKeyMeta(_ context.Context, id string, update repository.APIKeyUpdateRequest) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	k, ok := f.keys[id]
	if !ok {
		return store.ErrNotFound
	}
	if k.RevokedAt != nil {
		// repository 의 pgAPIKeyStore 와 동일 의미론 — revoked key 는 update 불가.
		return store.ErrNotFound
	}
	k.ExpiresAt = update.ExpiresAt
	k.AllowedCIDRs = update.AllowedCIDRs
	f.keys[id] = k
	return nil
}

// reqWithActor — gin context 에 system_admin actor 주입 + JSON body 까지 한 번에.
// pathParams 가 있으면 c.Params 에 직접 주입 (router 등록 없이 handler 단독 호출).
func reqWithActor(method, path string, body any, pathParams ...gin.Param) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	var reader *bytes.Reader
	if body != nil {
		raw, _ := json.Marshal(body)
		reader = bytes.NewReader(raw)
		c.Request = httptest.NewRequest(method, path, reader)
		c.Request.Header.Set("Content-Type", "application/json")
	} else {
		c.Request = httptest.NewRequest(method, path, nil)
	}
	if len(pathParams) > 0 {
		c.Params = pathParams
	}
	c.Set("devhub_actor_login", "alice")
	c.Set("devhub_actor_role", "system_admin")
	return c, w
}

// TestCreateAPIKey_NilStoreReturns503 — multi-key store 미설정 시 503. nil
// pointer panic 회피 + 운영 가시성.
func TestCreateAPIKey_NilStoreReturns503(t *testing.T) {
	h := NewAuthHandler(AuthConfig{}) // APIKeyStoreAdmin = nil
	c, w := reqWithActor("POST", "/api/v1/admin/api-keys", map[string]any{"name": "k"})
	h.CreateAPIKey(c)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d body=%s", w.Code, w.Body.String())
	}
}

// TestCreateAPIKey_RejectsEmptyName — name 빈 값/공백 400.
func TestCreateAPIKey_RejectsEmptyName(t *testing.T) {
	store := newFakeAPIKeyAdminStore()
	h := NewAuthHandler(AuthConfig{AuditStore: &fakeAuditStore{}, APIKeyStoreAdmin: store})
	c, w := reqWithActor("POST", "/api/v1/admin/api-keys", map[string]any{"name": ""})
	h.CreateAPIKey(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", w.Code, w.Body.String())
	}
}

// TestCreateAPIKey_RejectsInvalidCIDR — 생성 시점에 CIDR sanity. 잘못된 CIDR
// 면 400 (운영자가 즉시 알 수 있게).
func TestCreateAPIKey_RejectsInvalidCIDR(t *testing.T) {
	store := newFakeAPIKeyAdminStore()
	h := NewAuthHandler(AuthConfig{AuditStore: &fakeAuditStore{}, APIKeyStoreAdmin: store})
	c, w := reqWithActor("POST", "/api/v1/admin/api-keys", map[string]any{
		"name":          "k",
		"allowed_cidrs": []string{"not-a-cidr"},
	})
	h.CreateAPIKey(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "invalid CIDR") {
		t.Fatalf("expected CIDR error message, got %s", w.Body.String())
	}
}

// TestCreateAPIKey_RejectsPastExpiry — 과거 시각 expires_at 400.
func TestCreateAPIKey_RejectsPastExpiry(t *testing.T) {
	store := newFakeAPIKeyAdminStore()
	h := NewAuthHandler(AuthConfig{AuditStore: &fakeAuditStore{}, APIKeyStoreAdmin: store})
	past := "2020-01-01T00:00:00Z"
	c, w := reqWithActor("POST", "/api/v1/admin/api-keys", map[string]any{
		"name":       "k",
		"expires_at": past,
	})
	h.CreateAPIKey(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", w.Code, w.Body.String())
	}
}

// TestCreateAPIKey_SuccessAndRawKeyExposed — 성공 시 raw key 1회 응답 +
// summary 의 key_prefix 가 raw key 의 앞 8자와 일치. 보안 invariant.
func TestCreateAPIKey_SuccessAndRawKeyExposed(t *testing.T) {
	store := newFakeAPIKeyAdminStore()
	h := NewAuthHandler(AuthConfig{AuditStore: &fakeAuditStore{}, APIKeyStoreAdmin: store})
	c, w := reqWithActor("POST", "/api/v1/admin/api-keys", map[string]any{"name": "ci-runner"})
	h.CreateAPIKey(c)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", w.Code, w.Body.String())
	}
	var resp struct {
		APIKey struct {
			ID        string `json:"id"`
			KeyPrefix string `json:"key_prefix"`
			Status    string `json:"status"`
		} `json:"api_key"`
		RawKey  string `json:"key"`
		Warning string `json:"warning"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v body=%s", err, w.Body.String())
	}
	if !strings.HasPrefix(resp.RawKey, authkey.APIKeyPrefix) {
		t.Fatalf("raw key must start with %q, got %q", authkey.APIKeyPrefix, resp.RawKey)
	}
	if resp.APIKey.KeyPrefix != resp.RawKey[:authkey.APIKeyDisplayPrefixLength] {
		t.Fatalf("key prefix mismatch: summary=%q, raw[:8]=%q", resp.APIKey.KeyPrefix, resp.RawKey[:authkey.APIKeyDisplayPrefixLength])
	}
	if resp.APIKey.Status != "active" {
		t.Fatalf("expected status=active, got %q", resp.APIKey.Status)
	}
	if resp.Warning == "" {
		t.Fatal("expected warning message for raw key one-shot exposure")
	}
}

// TestListAPIKeys_NilStoreReturns503.
func TestListAPIKeys_NilStoreReturns503(t *testing.T) {
	h := NewAuthHandler(AuthConfig{})
	c, w := reqWithActor("GET", "/api/v1/admin/api-keys", nil)
	h.ListAPIKeys(c)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}

// TestListAPIKeys_RawKeyNotLeaked — list 응답에 raw key / key_hash 절대
// 미포함. 보안 invariant 검증.
func TestListAPIKeys_RawKeyNotLeaked(t *testing.T) {
	store := newFakeAPIKeyAdminStore()
	// pre-populate
	exp := "2030-01-01T00:00:00Z"
	_, _ = store.CreateAPIKey(context.Background(), make([]byte, 32), repository.APIKey{
		ID:        "u-1",
		Name:      "k1",
		KeyPrefix: "dhk_aaaa",
		CreatedBy: "alice",
		ExpiresAt: &exp,
		Status:    repository.APIKeyStatusActive,
	})
	h := NewAuthHandler(AuthConfig{AuditStore: &fakeAuditStore{}, APIKeyStoreAdmin: store})
	c, w := reqWithActor("GET", "/api/v1/admin/api-keys", nil)
	h.ListAPIKeys(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if strings.Contains(strings.ToLower(body), "\"key\":\"dhk_") {
		t.Fatalf("raw key leaked in list response: %s", body)
	}
	if strings.Contains(strings.ToLower(body), "key_hash") {
		t.Fatalf("key_hash leaked in list response: %s", body)
	}
}

// TestRevokeAPIKey_NotFound — 존재하지 않는 id → 404.
func TestRevokeAPIKey_NotFound(t *testing.T) {
	store := newFakeAPIKeyAdminStore()
	h := NewAuthHandler(AuthConfig{AuditStore: &fakeAuditStore{}, APIKeyStoreAdmin: store})
	c, w := reqWithActor("DELETE", "/api/v1/admin/api-keys/ghost-id", nil, gin.Param{Key: "api_key_id", Value: "ghost-id"})
	h.RevokeAPIKey(c)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", w.Code, w.Body.String())
	}
}

// TestRevokeAPIKey_Success — 성공 시 204 + store 의 row 가 revoked 상태.
func TestRevokeAPIKey_Success(t *testing.T) {
	store := newFakeAPIKeyAdminStore()
	_, _ = store.CreateAPIKey(context.Background(), make([]byte, 32), repository.APIKey{
		ID:        "u-2",
		Name:      "k2",
		KeyPrefix: "dhk_bbbb",
		CreatedBy: "alice",
		Status:    repository.APIKeyStatusActive,
	})
	h := NewAuthHandler(AuthConfig{AuditStore: &fakeAuditStore{}, APIKeyStoreAdmin: store})
	c, w := reqWithActor("DELETE", "/api/v1/admin/api-keys/u-2", nil, gin.Param{Key: "api_key_id", Value: "u-2"})
	h.RevokeAPIKey(c)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d body=%s", w.Code, w.Body.String())
	}
	k := store.keys["u-2"]
	if k.Status != repository.APIKeyStatusRevoked {
		t.Fatalf("expected status=revoked, got %q", k.Status)
	}
	if k.RevokedAt == nil || *k.RevokedAt == "" {
		t.Fatal("expected revoked_at to be set")
	}
}

// TestRevokeAPIKey_StoreError — store fault → 500.
func TestRevokeAPIKey_StoreError(t *testing.T) {
	store := newFakeAPIKeyAdminStore()
	store.revokeErr = errors.New("db down")
	h := NewAuthHandler(AuthConfig{AuditStore: &fakeAuditStore{}, APIKeyStoreAdmin: store})
	c, w := reqWithActor("DELETE", "/api/v1/admin/api-keys/u-3", nil, gin.Param{Key: "api_key_id", Value: "u-3"})
	h.RevokeAPIKey(c)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d body=%s", w.Code, w.Body.String())
	}
}

// TestUpdateAPIKeyMeta_RejectsRevoked — revoked key 갱신 시도 → 404
// (repository 가 `revoked_at IS NULL` 조건으로 0 rows → store.ErrNotFound).
func TestUpdateAPIKeyMeta_RejectsRevoked(t *testing.T) {
	store := newFakeAPIKeyAdminStore()
	now := time.Now().UTC().Format(time.RFC3339)
	_, _ = store.CreateAPIKey(context.Background(), make([]byte, 32), repository.APIKey{
		ID:        "u-rev",
		Name:      "k-rev",
		KeyPrefix: "dhk_cccc",
		CreatedBy: "alice",
		RevokedAt: &now,
		Status:    repository.APIKeyStatusRevoked,
	})
	h := NewAuthHandler(AuthConfig{AuditStore: &fakeAuditStore{}, APIKeyStoreAdmin: store})
	c, w := reqWithActor("PATCH", "/api/v1/admin/api-keys/u-rev", map[string]any{
		"allowed_cidrs": []string{"10.0.0.0/8"},
	}, gin.Param{Key: "api_key_id", Value: "u-rev"})
	h.UpdateAPIKeyMeta(c)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", w.Code, w.Body.String())
	}
}

// TestUpdateAPIKeyMeta_Success — 200 + status=updated.
func TestUpdateAPIKeyMeta_Success(t *testing.T) {
	store := newFakeAPIKeyAdminStore()
	_, _ = store.CreateAPIKey(context.Background(), make([]byte, 32), repository.APIKey{
		ID:        "u-ok",
		Name:      "k-ok",
		KeyPrefix: "dhk_dddd",
		CreatedBy: "alice",
		Status:    repository.APIKeyStatusActive,
	})
	h := NewAuthHandler(AuthConfig{AuditStore: &fakeAuditStore{}, APIKeyStoreAdmin: store})
	exp := "2030-12-31T23:59:59Z"
	c, w := reqWithActor("PATCH", "/api/v1/admin/api-keys/u-ok", map[string]any{
		"expires_at":    exp,
		"allowed_cidrs": []string{"10.0.0.0/8"},
	}, gin.Param{Key: "api_key_id", Value: "u-ok"})
	h.UpdateAPIKeyMeta(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	var resp struct {
		Status   string `json:"status"`
		APIKeyID string `json:"api_key_id"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Status != "updated" {
		t.Fatalf("expected status=updated, got %q", resp.Status)
	}
	k := store.keys["u-ok"]
	if k.ExpiresAt == nil || *k.ExpiresAt != exp {
		t.Fatalf("expires_at not updated: %v", k.ExpiresAt)
	}
}
