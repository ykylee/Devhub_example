package httpapi

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
)

// UT-httpapi-25 — Integration registry/binding handler tests (API-69..75 baseline).
func TestCreateIntegrationProvider_Happy(t *testing.T) {
	router := newApplicationsRouter(newMemoryApplicationStore())
	rec := doJSON(t, router, http.MethodPost, "/api/v1/integration/providers",
		`{"provider_key":"jira-main","provider_type":"alm","display_name":"Jira","auth_mode":"oauth2","credentials_ref":"secret://jira","capabilities":["issue.read"]}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"provider_key":"jira-main"`)) {
		t.Errorf("unexpected body: %s", rec.Body.String())
	}
}

func TestCreateIntegrationProvider_Duplicate(t *testing.T) {
	router := newApplicationsRouter(newMemoryApplicationStore())
	body := `{"provider_key":"jira-main","provider_type":"alm","display_name":"Jira","auth_mode":"oauth2","credentials_ref":"secret://jira"}`
	first := doJSON(t, router, http.MethodPost, "/api/v1/integration/providers", body)
	if first.Code != http.StatusCreated {
		t.Fatalf("seed failed: %s", first.Body.String())
	}
	rec := doJSON(t, router, http.MethodPost, "/api/v1/integration/providers", body)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestListIntegrationProviders_FilterEnabled(t *testing.T) {
	router := newApplicationsRouter(newMemoryApplicationStore())
	_ = doJSON(t, router, http.MethodPost, "/api/v1/integration/providers",
		`{"provider_key":"jira-main","provider_type":"alm","display_name":"Jira","auth_mode":"oauth2","credentials_ref":"secret://jira"}`)
	_ = doJSON(t, router, http.MethodPost, "/api/v1/integration/providers",
		`{"provider_key":"gitea-main","provider_type":"scm","display_name":"Gitea","auth_mode":"token","credentials_ref":"secret://gitea"}`)
	_ = doJSON(t, router, http.MethodPatch, "/api/v1/integration/providers/prov-gitea-main",
		`{"enabled":false}`)

	rec := doJSON(t, router, http.MethodGet, "/api/v1/integration/providers?enabled=true", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !bytes.Contains([]byte(body), []byte(`"provider_key":"jira-main"`)) || bytes.Contains([]byte(body), []byte(`"provider_key":"gitea-main"`)) {
		t.Errorf("enabled filter mismatch: %s", body)
	}
}

func TestSyncIntegrationProvider_Happy(t *testing.T) {
	router := newApplicationsRouter(newMemoryApplicationStore())
	seed := doJSON(t, router, http.MethodPost, "/api/v1/integration/providers",
		`{"provider_key":"jira-main","provider_type":"alm","display_name":"Jira","auth_mode":"oauth2","credentials_ref":"secret://jira"}`)
	if seed.Code != http.StatusCreated {
		t.Fatalf("seed failed: %s", seed.Body.String())
	}
	rec := doJSON(t, router, http.MethodPost, "/api/v1/integration/providers/prov-jira-main/sync", "{}")
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"job_id":"job-prov-jira-main"`)) {
		t.Errorf("expected job_id: %s", rec.Body.String())
	}
}

func TestCreateIntegrationBinding_Happy(t *testing.T) {
	router := newApplicationsRouter(newMemoryApplicationStore())
	seed := doJSON(t, router, http.MethodPost, "/api/v1/integration/providers",
		`{"provider_key":"jira-main","provider_type":"alm","display_name":"Jira","auth_mode":"oauth2","credentials_ref":"secret://jira"}`)
	if seed.Code != http.StatusCreated {
		t.Fatalf("seed failed: %s", seed.Body.String())
	}
	rec := doJSON(t, router, http.MethodPost, "/api/v1/integration/bindings",
		`{"scope_type":"application","scope_id":"APP-001","provider_id":"prov-jira-main","external_key":"PROJ","policy":"execution_system"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"scope_id":"APP-001"`)) {
		t.Errorf("unexpected binding response: %s", rec.Body.String())
	}
}

func TestCreateIntegrationBinding_ForbiddenForDeveloperRole(t *testing.T) {
	store := newMemoryApplicationStore()
	if _, err := store.CreateIntegrationProvider(context.Background(), domain.IntegrationProvider{
		ID:             "prov-jira-main",
		ProviderKey:    "jira-main",
		ProviderType:   domain.IntegrationProviderType("alm"),
		DisplayName:    "Jira",
		Enabled:        true,
		AuthMode:       domain.IntegrationAuthMode("oauth2"),
		CredentialsRef: "secret://jira",
		SyncStatus:     "requested",
	}); err != nil {
		t.Fatalf("seed provider: %v", err)
	}

	router := NewRouter(RouterConfig{
		ApplicationStore: store,
		BearerTokenVerifier: &fakeBearerTokenVerifier{actor: AuthenticatedActor{
			Login:   "dev-user",
			Subject: "user-dev-user",
			Role:    "developer",
		}},
	})

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/integration/bindings",
		bytes.NewBufferString(`{"scope_type":"application","scope_id":"APP-001","provider_id":"prov-jira-main","external_key":"PROJ","policy":"execution_system"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer t")
	rec := httptestDo(t, router, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

// API-80 DELETE handler tests (sprint claude/work_260518-j).
// FK guard: binding 1건 이상이면 409 integration_provider_has_bindings — 실수
// cascade 방지. integration_sync_jobs 는 schema 의 ON DELETE CASCADE 로 자동 정리.

func TestDeleteIntegrationProvider_Happy(t *testing.T) {
	router := newApplicationsRouter(newMemoryApplicationStore())
	seed := doJSON(t, router, http.MethodPost, "/api/v1/integration/providers",
		`{"provider_key":"jira-main","provider_type":"alm","display_name":"Jira","auth_mode":"oauth2","credentials_ref":"secret://jira"}`)
	if seed.Code != http.StatusCreated {
		t.Fatalf("seed failed: %s", seed.Body.String())
	}
	rec := doJSON(t, router, http.MethodDelete, "/api/v1/integration/providers/prov-jira-main", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	// 삭제 후 GET 은 404 (이미 row 없음).
	check := doJSON(t, router, http.MethodGet, "/api/v1/integration/providers", "")
	if bytes.Contains(check.Body.Bytes(), []byte(`"provider_key":"jira-main"`)) {
		t.Errorf("provider must be gone from list: %s", check.Body.String())
	}
}

func TestDeleteIntegrationProvider_NotFound(t *testing.T) {
	router := newApplicationsRouter(newMemoryApplicationStore())
	rec := doJSON(t, router, http.MethodDelete, "/api/v1/integration/providers/prov-ghost", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"integration_provider_not_found"`)) {
		t.Errorf("expected integration_provider_not_found code: %s", rec.Body.String())
	}
}

func TestDeleteIntegrationProvider_HasBindings(t *testing.T) {
	router := newApplicationsRouter(newMemoryApplicationStore())
	seed := doJSON(t, router, http.MethodPost, "/api/v1/integration/providers",
		`{"provider_key":"jira-main","provider_type":"alm","display_name":"Jira","auth_mode":"oauth2","credentials_ref":"secret://jira"}`)
	if seed.Code != http.StatusCreated {
		t.Fatalf("seed failed: %s", seed.Body.String())
	}
	binding := doJSON(t, router, http.MethodPost, "/api/v1/integration/bindings",
		`{"scope_type":"application","scope_id":"APP-001","provider_id":"prov-jira-main","external_key":"PROJ","policy":"execution_system"}`)
	if binding.Code != http.StatusCreated {
		t.Fatalf("seed binding failed: %s", binding.Body.String())
	}
	rec := doJSON(t, router, http.MethodDelete, "/api/v1/integration/providers/prov-jira-main", "")
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"integration_provider_has_bindings"`)) {
		t.Errorf("expected integration_provider_has_bindings code: %s", rec.Body.String())
	}
}

func TestDeleteIntegrationProvider_ForbiddenForDeveloperRole(t *testing.T) {
	store := newMemoryApplicationStore()
	if _, err := store.CreateIntegrationProvider(context.Background(), domain.IntegrationProvider{
		ID:             "prov-jira-main",
		ProviderKey:    "jira-main",
		ProviderType:   domain.IntegrationProviderType("alm"),
		DisplayName:    "Jira",
		Enabled:        true,
		AuthMode:       domain.IntegrationAuthMode("oauth2"),
		CredentialsRef: "secret://jira",
		SyncStatus:     "requested",
	}); err != nil {
		t.Fatalf("seed provider: %v", err)
	}

	router := NewRouter(RouterConfig{
		ApplicationStore: store,
		BearerTokenVerifier: &fakeBearerTokenVerifier{actor: AuthenticatedActor{
			Login:   "dev-user",
			Subject: "user-dev-user",
			Role:    "developer",
		}},
	})

	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/integration/providers/prov-jira-main", nil)
	req.Header.Set("Authorization", "Bearer t")
	rec := httptestDo(t, router, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestIntegrationProviderWebhook_Happy(t *testing.T) {
	router := newApplicationsRouter(newMemoryApplicationStore())
	seed := doJSON(t, router, http.MethodPost, "/api/v1/integration/providers",
		`{"provider_key":"jira-main","provider_type":"alm","display_name":"Jira","auth_mode":"oauth2","credentials_ref":"hmac_sha256:test-secret"}`)
	if seed.Code != http.StatusCreated {
		t.Fatalf("seed failed: %s", seed.Body.String())
	}
	body := []byte(`{"event":"x"}`)
	mac := hmac.New(sha256.New, []byte("test-secret"))
	mac.Write(body)
	signature := hex.EncodeToString(mac.Sum(nil))

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/integration/providers/jira-main/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Integration-Signature", signature)
	rec := httptestDo(t, router, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestIntegrationProviderWebhook_InvalidSignature(t *testing.T) {
	router := newApplicationsRouter(newMemoryApplicationStore())
	seed := doJSON(t, router, http.MethodPost, "/api/v1/integration/providers",
		`{"provider_key":"jira-main","provider_type":"alm","display_name":"Jira","auth_mode":"oauth2","credentials_ref":"hmac_sha256:test-secret"}`)
	if seed.Code != http.StatusCreated {
		t.Fatalf("seed failed: %s", seed.Body.String())
	}
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/integration/providers/jira-main/webhook", bytes.NewReader([]byte(`{"event":"x"}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Integration-Signature", "bad-signature")
	rec := httptestDo(t, router, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"integration_webhook_signature_invalid"`)) {
		t.Errorf("unexpected body: %s", rec.Body.String())
	}
}

func TestIntegrationProviderWebhook_ProviderSDKHappy(t *testing.T) {
	router := newApplicationsRouter(newMemoryApplicationStore())
	seed := doJSON(t, router, http.MethodPost, "/api/v1/integration/providers",
		`{"provider_key":"jira-main","provider_type":"alm","display_name":"Jira","auth_mode":"oauth2","credentials_ref":"provider_sdk:jira:test-secret"}`)
	if seed.Code != http.StatusCreated {
		t.Fatalf("seed failed: %s", seed.Body.String())
	}
	body := []byte(`{"event":"x"}`)
	mac := hmac.New(sha256.New, []byte("test-secret"))
	mac.Write(body)
	signature := hex.EncodeToString(mac.Sum(nil))

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/integration/providers/jira-main/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Integration-Signature", signature)
	rec := httptestDo(t, router, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestIntegrationProviderWebhook_ProviderSDKUnsupportedProvider(t *testing.T) {
	router := newApplicationsRouter(newMemoryApplicationStore())
	seed := doJSON(t, router, http.MethodPost, "/api/v1/integration/providers",
		`{"provider_key":"custom-main","provider_type":"alm","display_name":"Custom","auth_mode":"oauth2","credentials_ref":"provider_sdk:custom:test-secret"}`)
	if seed.Code != http.StatusCreated {
		t.Fatalf("seed failed: %s", seed.Body.String())
	}

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/integration/providers/custom-main/webhook", bytes.NewReader([]byte(`{"event":"x"}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Integration-Signature", "any-signature")
	rec := httptestDo(t, router, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"integration_webhook_signature_invalid"`)) {
		t.Errorf("unexpected body: %s", rec.Body.String())
	}
}

func TestIntegrationProviderWebhook_DuplicateDeliveryConflict(t *testing.T) {
	eventStore := &dedupeEventStore{seen: map[string]bool{}}
	router := NewRouter(RouterConfig{
		ApplicationStore: newMemoryApplicationStore(),
		AuthDevFallback:  true,
		EventStore:       eventStore,
	})
	seed := doJSON(t, router, http.MethodPost, "/api/v1/integration/providers",
		`{"provider_key":"jira-main","provider_type":"alm","display_name":"Jira","auth_mode":"oauth2","credentials_ref":"hmac_sha256:test-secret"}`)
	if seed.Code != http.StatusCreated {
		t.Fatalf("seed failed: %s", seed.Body.String())
	}
	body := []byte(`{"event":"x"}`)
	mac := hmac.New(sha256.New, []byte("test-secret"))
	mac.Write(body)
	signature := hex.EncodeToString(mac.Sum(nil))

	req1, _ := http.NewRequest(http.MethodPost, "/api/v1/integration/providers/jira-main/webhook", bytes.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("X-Integration-Signature", signature)
	req1.Header.Set("X-Integration-Delivery", "delivery-001")
	rec1 := httptestDo(t, router, req1)
	if rec1.Code != http.StatusAccepted {
		t.Fatalf("first status=%d body=%s", rec1.Code, rec1.Body.String())
	}

	req2, _ := http.NewRequest(http.MethodPost, "/api/v1/integration/providers/jira-main/webhook", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("X-Integration-Signature", signature)
	req2.Header.Set("X-Integration-Delivery", "delivery-001")
	rec2 := httptestDo(t, router, req2)
	if rec2.Code != http.StatusConflict {
		t.Fatalf("second status=%d body=%s", rec2.Code, rec2.Body.String())
	}
	if !bytes.Contains(rec2.Body.Bytes(), []byte(`"code":"integration_event_duplicate"`)) {
		t.Errorf("unexpected body: %s", rec2.Body.String())
	}
}

func TestIntegrationProviderWebhook_InvalidSignatureMarksOnlyTargetProviderDegraded(t *testing.T) {
	router := newApplicationsRouter(newMemoryApplicationStore())
	seedA := doJSON(t, router, http.MethodPost, "/api/v1/integration/providers",
		`{"provider_key":"jira-main","provider_type":"alm","display_name":"Jira","auth_mode":"oauth2","credentials_ref":"hmac_sha256:test-secret"}`)
	if seedA.Code != http.StatusCreated {
		t.Fatalf("seedA failed: %s", seedA.Body.String())
	}
	seedB := doJSON(t, router, http.MethodPost, "/api/v1/integration/providers",
		`{"provider_key":"github-main","provider_type":"scm","display_name":"GitHub","auth_mode":"oauth2","credentials_ref":"hmac_sha256:gh-secret"}`)
	if seedB.Code != http.StatusCreated {
		t.Fatalf("seedB failed: %s", seedB.Body.String())
	}

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/integration/providers/jira-main/webhook", bytes.NewReader([]byte(`{"event":"x"}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Integration-Signature", "bad-signature")
	rec := httptestDo(t, router, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}

	list := doJSON(t, router, http.MethodGet, "/api/v1/integration/providers", "")
	if list.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", list.Code, list.Body.String())
	}
	body := list.Body.String()
	if !bytes.Contains([]byte(body), []byte(`"provider_key":"jira-main"`)) ||
		!bytes.Contains([]byte(body), []byte(`"sync_status":"degraded"`)) ||
		!bytes.Contains([]byte(body), []byte(`"last_error_code":"webhook_signature_invalid"`)) {
		t.Fatalf("jira provider should be degraded: %s", body)
	}
	if !bytes.Contains([]byte(body), []byte(`"provider_key":"github-main"`)) ||
		!bytes.Contains([]byte(body), []byte(`"sync_status":"requested"`)) {
		t.Fatalf("github provider should remain requested: %s", body)
	}
}

type dedupeEventStore struct {
	seen map[string]bool
}

func (s *dedupeEventStore) SaveWebhookEvent(_ context.Context, event store.WebhookEvent) (int64, error) {
	if s.seen[event.DedupeKey] {
		return 0, store.ErrDuplicateEvent
	}
	s.seen[event.DedupeKey] = true
	return 1, nil
}

func (s *dedupeEventStore) ListWebhookEvents(_ context.Context, _ store.ListWebhookEventsOptions) ([]store.WebhookEvent, error) {
	return nil, nil
}

func httptestDo(t *testing.T, router http.Handler, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}
