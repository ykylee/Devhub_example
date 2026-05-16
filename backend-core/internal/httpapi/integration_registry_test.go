package httpapi

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"
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

func httptestDo(t *testing.T, router http.Handler, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}
