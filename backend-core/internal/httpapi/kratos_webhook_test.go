package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/devhub/backend-core/internal/domain"
)

const testKratosWebhookToken = "test-kratos-webhook-secret"

func newKratosWebhookRouter(t *testing.T, store AuditStore, token string) (*memoryAuditStore, http.Handler) {
	t.Helper()
	memStore, _ := store.(*memoryAuditStore)
	if memStore == nil {
		memStore = &memoryAuditStore{}
	}
	router := testRouter(RouterConfig{
		AuditStore:         memStore,
		KratosWebhookToken: token,
	})
	return memStore, router
}

func postKratosPasswordWebhook(t *testing.T, router http.Handler, authHeader string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/kratos/hook/settings/password/after", &buf)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestKratosPasswordWebhookRecordsAuditLog(t *testing.T) {
	store, router := newKratosWebhookRouter(t, nil, testKratosWebhookToken)

	rec := postKratosPasswordWebhook(t, router, "Bearer "+testKratosWebhookToken, map[string]any{
		"identity_id": "kratos-uuid-1",
		"email":       "alice@example.com",
		"occurred_at": "2026-05-12T07:30:00Z",
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s; want 200", rec.Code, rec.Body.String())
	}
	if len(store.logs) != 1 {
		t.Fatalf("audit logs len = %d; want 1", len(store.logs))
	}
	got := store.logs[0]
	if got.SourceType != domain.AuditSourceKratos {
		t.Errorf("source_type = %q; want %q", got.SourceType, domain.AuditSourceKratos)
	}
	if got.Action != "account.password_changed" {
		t.Errorf("action = %q; want account.password_changed", got.Action)
	}
	if got.TargetType != "kratos_identity" {
		t.Errorf("target_type = %q; want kratos_identity", got.TargetType)
	}
	if got.TargetID != "kratos-uuid-1" {
		t.Errorf("target_id = %q; want kratos-uuid-1", got.TargetID)
	}
	if got.ActorLogin != "kratos:kratos-uuid-1" {
		t.Errorf("actor_login = %q; want kratos:kratos-uuid-1", got.ActorLogin)
	}
	if email, _ := got.Payload["email"].(string); email != "alice@example.com" {
		t.Errorf("payload.email = %q; want alice@example.com", email)
	}
}

func TestKratosPasswordWebhookAcceptsBareToken(t *testing.T) {
	_, router := newKratosWebhookRouter(t, nil, testKratosWebhookToken)

	rec := postKratosPasswordWebhook(t, router, testKratosWebhookToken, map[string]any{
		"identity_id": "kratos-uuid-2",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s; want 200 for bare-token auth", rec.Code, rec.Body.String())
	}
}

func TestKratosPasswordWebhookRejectsMissingAuth(t *testing.T) {
	_, router := newKratosWebhookRouter(t, nil, testKratosWebhookToken)

	rec := postKratosPasswordWebhook(t, router, "", map[string]any{"identity_id": "x"})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d; want 401", rec.Code)
	}
}

func TestKratosPasswordWebhookRejectsWrongSecret(t *testing.T) {
	_, router := newKratosWebhookRouter(t, nil, testKratosWebhookToken)

	rec := postKratosPasswordWebhook(t, router, "Bearer wrong-secret", map[string]any{"identity_id": "x"})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d; want 401", rec.Code)
	}
}

func TestKratosPasswordWebhookUnavailableWhenSecretUnset(t *testing.T) {
	_, router := newKratosWebhookRouter(t, nil, "")

	rec := postKratosPasswordWebhook(t, router, "Bearer anything", map[string]any{"identity_id": "x"})
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d; want 503 when DEVHUB_KRATOS_WEBHOOK_TOKEN is empty", rec.Code)
	}
}

func TestKratosPasswordWebhookRejectsMissingIdentityID(t *testing.T) {
	_, router := newKratosWebhookRouter(t, nil, testKratosWebhookToken)

	rec := postKratosPasswordWebhook(t, router, "Bearer "+testKratosWebhookToken, map[string]any{
		"email": "no-id@example.com",
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want 400", rec.Code)
	}
}

func TestKratosPasswordWebhookRejectsInvalidJSON(t *testing.T) {
	_, router := newKratosWebhookRouter(t, nil, testKratosWebhookToken)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/kratos/hook/settings/password/after", bytes.NewBufferString("{not-json"))
	req.Header.Set("Authorization", "Bearer "+testKratosWebhookToken)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want 400", rec.Code)
	}
}
