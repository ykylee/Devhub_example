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

// 등록 UX 고도화 #2 — base_url (endpoint) round-trip.
func TestCreateIntegrationProvider_WithBaseURL(t *testing.T) {
	router := newApplicationsRouter(newMemoryApplicationStore())
	rec := doJSON(t, router, http.MethodPost, "/api/v1/integration/providers",
		`{"provider_key":"gitea-url","provider_type":"scm","display_name":"Gitea","auth_mode":"token","credentials_ref":"provider_sdk:gitea:s","base_url":"https://gitea.example.com"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"base_url":"https://gitea.example.com"`)) {
		t.Errorf("response should echo base_url: %s", rec.Body.String())
	}
}

// Gitea 연동 #3 — api_token 은 write-only: 저장되지만 응답엔 raw 미노출 (api_token_set 만).
func TestCreateIntegrationProvider_APITokenWriteOnly(t *testing.T) {
	router := newApplicationsRouter(newMemoryApplicationStore())
	const secretToken = "gitea-pat-supersecret-xyz"
	rec := doJSON(t, router, http.MethodPost, "/api/v1/integration/providers",
		`{"provider_key":"gitea-tok","provider_type":"scm","display_name":"Gitea","auth_mode":"token","credentials_ref":"provider_sdk:gitea:wh","api_token":"`+secretToken+`"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"api_token_set":true`)) {
		t.Errorf("response should report api_token_set true: %s", rec.Body.String())
	}
	if bytes.Contains(rec.Body.Bytes(), []byte(secretToken)) {
		t.Errorf("raw api_token must NOT be exposed in response: %s", rec.Body.String())
	}
}

// auth_mode 별 구조화 자격증명: 비밀 외 필드(auth_username/client_id/token_url)는
// 응답에 노출, auth_secret 은 write-only (auth_secret_set bool 만).
func TestCreateIntegrationProvider_BasicAuthCredentials(t *testing.T) {
	router := newApplicationsRouter(newMemoryApplicationStore())
	const secret = "basic-password-supersecret"
	rec := doJSON(t, router, http.MethodPost, "/api/v1/integration/providers",
		`{"provider_key":"gitea-basic","provider_type":"scm","display_name":"Gitea","auth_mode":"basic","credentials_ref":"hmac_sha256:wh","auth_username":"alice","auth_secret":"`+secret+`"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"auth_username":"alice"`)) {
		t.Errorf("response should echo auth_username: %s", rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"auth_secret_set":true`)) {
		t.Errorf("response should report auth_secret_set true: %s", rec.Body.String())
	}
	if bytes.Contains(rec.Body.Bytes(), []byte(secret)) {
		t.Errorf("raw auth_secret must NOT be exposed in response: %s", rec.Body.String())
	}
}

// oauth2 token_url 은 http(s) URL 만 허용.
func TestCreateIntegrationProvider_InvalidAuthTokenURL(t *testing.T) {
	router := newApplicationsRouter(newMemoryApplicationStore())
	rec := doJSON(t, router, http.MethodPost, "/api/v1/integration/providers",
		`{"provider_key":"oauth-bad","provider_type":"scm","display_name":"X","auth_mode":"oauth2","credentials_ref":"hmac_sha256:s","auth_client_id":"cid","auth_token_url":"not-a-url","auth_secret":"cs"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("invalid_auth_token_url")) {
		t.Errorf("expected invalid_auth_token_url code: %s", rec.Body.String())
	}
}

// base_url 은 http(s) scheme 만 허용.
func TestCreateIntegrationProvider_InvalidBaseURL(t *testing.T) {
	router := newApplicationsRouter(newMemoryApplicationStore())
	// codex PR #352 P2 — 비-http(s) + scheme-only(host 누락) 모두 거부.
	for _, bad := range []string{"ftp://nope", "https://"} {
		rec := doJSON(t, router, http.MethodPost, "/api/v1/integration/providers",
			`{"provider_key":"bad-url","provider_type":"scm","display_name":"X","auth_mode":"token","credentials_ref":"hmac_sha256:s","base_url":"`+bad+`"}`)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("base_url %q should be 400, got %d body=%s", bad, rec.Code, rec.Body.String())
		}
		if !bytes.Contains(rec.Body.Bytes(), []byte("invalid_base_url")) {
			t.Errorf("base_url %q: expected invalid_base_url code: %s", bad, rec.Body.String())
		}
	}
}

// 등록 UX 고도화 #5 — test-connection reachability.
func TestTestIntegrationConnection_Reachable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	router := newApplicationsRouter(newMemoryApplicationStore())
	rec := doJSON(t, router, http.MethodPost, "/api/v1/integration/test-connection",
		`{"base_url":"`+srv.URL+`"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"reachable":true`)) {
		t.Errorf("expected reachable true: %s", rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"status_code":200`)) {
		t.Errorf("expected status_code 200: %s", rec.Body.String())
	}
}

func TestTestIntegrationConnection_InvalidURL(t *testing.T) {
	router := newApplicationsRouter(newMemoryApplicationStore())
	// codex PR #352 P2 — scheme-only (host 누락) 도 거부.
	for _, body := range []string{`{"base_url":""}`, `{"base_url":"ftp://nope"}`, `{"base_url":"https://"}`} {
		rec := doJSON(t, router, http.MethodPost, "/api/v1/integration/test-connection", body)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("body %s: expected 400, got %d body=%s", body, rec.Code, rec.Body.String())
		}
	}
}

func TestSyncIntegrationProvider_Happy(t *testing.T) {
	router := newApplicationsRouter(newMemoryApplicationStore())
	// SCM provider — sync 가능한 유일한 provider_type (Gitea 워커 대상).
	seed := doJSON(t, router, http.MethodPost, "/api/v1/integration/providers",
		`{"provider_key":"gitea-main","provider_type":"scm","display_name":"Gitea","auth_mode":"token","credentials_ref":"secret://gitea"}`)
	if seed.Code != http.StatusCreated {
		t.Fatalf("seed failed: %s", seed.Body.String())
	}
	rec := doJSON(t, router, http.MethodPost, "/api/v1/integration/providers/prov-gitea-main/sync", "{}")
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"job_id":"job-prov-gitea-main"`)) {
		t.Errorf("expected job_id: %s", rec.Body.String())
	}
}

// codex review PR #345 P2 — 비-SCM provider 의 sync 는 queue 전에 fast-fail.
// (소비할 worker 가 없어 영구 queued 로 남는 zombie job 방지.)
func TestSyncIntegrationProvider_RejectsNonSCM(t *testing.T) {
	router := newApplicationsRouter(newMemoryApplicationStore())
	seed := doJSON(t, router, http.MethodPost, "/api/v1/integration/providers",
		`{"provider_key":"jira-main","provider_type":"alm","display_name":"Jira","auth_mode":"oauth2","credentials_ref":"secret://jira"}`)
	if seed.Code != http.StatusCreated {
		t.Fatalf("seed failed: %s", seed.Body.String())
	}
	rec := doJSON(t, router, http.MethodPost, "/api/v1/integration/providers/prov-jira-main/sync", "{}")
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("non-SCM sync should be rejected with 422, got status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`integration_sync_unsupported_provider_type`)) {
		t.Errorf("expected unsupported_provider_type code: %s", rec.Body.String())
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

// --- PATCH /api/v1/integration/bindings/:binding_id (API-74)
// sprint claude/test-integration-bindings-handlers-2026-05-21.
// 기존엔 TestCreateIntegrationBinding_* 만 cover 됐고 update handler 의
// happy/404/422/409/RBAC 분기가 모두 미가드 상태였다.

// seedBindingFixture — 새 router + seed (provider + binding 1건). 모든 PATCH/
// DELETE happy/neg test 가 동일 진입점을 쓰도록 helper 화.
func seedBindingFixture(t *testing.T) (http.Handler, string) {
	t.Helper()
	router := newApplicationsRouter(newMemoryApplicationStore())
	seed := doJSON(t, router, http.MethodPost, "/api/v1/integration/providers",
		`{"provider_key":"jira-main","provider_type":"alm","display_name":"Jira","auth_mode":"oauth2","credentials_ref":"secret://jira"}`)
	if seed.Code != http.StatusCreated {
		t.Fatalf("seed provider: %s", seed.Body.String())
	}
	created := doJSON(t, router, http.MethodPost, "/api/v1/integration/bindings",
		`{"scope_type":"application","scope_id":"APP-001","provider_id":"prov-jira-main","external_key":"PROJ","policy":"execution_system"}`)
	if created.Code != http.StatusCreated {
		t.Fatalf("seed binding: %s", created.Body.String())
	}
	// fake store 의 ID convention: "bind-<scope_id>-<external_key>". production
	// store 는 UUID 라 별도 응답 body 에서 추출하지만, test 는 fake convention
	// 으로 충분 — 모든 PATCH/DELETE handler 가 path :binding_id 만 사용.
	return router, "bind-APP-001-PROJ"
}

func TestUpdateIntegrationBinding_Happy(t *testing.T) {
	router, bindingID := seedBindingFixture(t)
	rec := doJSON(t, router, http.MethodPatch, "/api/v1/integration/bindings/"+bindingID,
		`{"external_key":"PROJ-RENAMED","policy":"summary_only","enabled":false}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	// handler 반환값으로 store 갱신 자연 검증 (handler 가 store.UpdateIntegrationBinding
	// 결과를 그대로 직렬화). 다른 PATCH happy 테스트의 substring 패턴 정합.
	body := rec.Body.String()
	for _, want := range []string{
		`"external_key":"PROJ-RENAMED"`,
		`"policy":"summary_only"`,
		`"enabled":false`,
	} {
		if !bytes.Contains([]byte(body), []byte(want)) {
			t.Errorf("response missing %s: %s", want, body)
		}
	}
}

func TestUpdateIntegrationBinding_NotFound(t *testing.T) {
	router := newApplicationsRouter(newMemoryApplicationStore())
	rec := doJSON(t, router, http.MethodPatch, "/api/v1/integration/bindings/bind-ghost",
		`{"policy":"summary_only"}`)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUpdateIntegrationBinding_InvalidPolicy(t *testing.T) {
	router, bindingID := seedBindingFixture(t)
	rec := doJSON(t, router, http.MethodPatch, "/api/v1/integration/bindings/"+bindingID,
		`{"policy":"not_a_valid_policy"}`)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("unsupported policy")) {
		t.Errorf("expected 'unsupported policy' message: %s", rec.Body.String())
	}
}

func TestUpdateIntegrationBinding_ConflictDuplicate(t *testing.T) {
	router, firstID := seedBindingFixture(t)
	// 같은 router 에 같은 provider 로 두 번째 binding (scope_id 가 다름) 생성.
	second := doJSON(t, router, http.MethodPost, "/api/v1/integration/bindings",
		`{"scope_type":"application","scope_id":"APP-002","provider_id":"prov-jira-main","external_key":"OTHER","policy":"execution_system"}`)
	if second.Code != http.StatusCreated {
		t.Fatalf("seed second binding: %s", second.Body.String())
	}
	// firstID 의 external_key 를 second 의 것 (OTHER) 으로 변경 시도 — fake store
	// 의 (scope_type, scope_id, provider_id, external_key) 4-tuple 가드는 scope_id
	// 가 다르면 충돌 안 함. 같은 충돌을 일으키려면 second 의 scope_id 와 일치시켜야
	// 하는데, 여기선 different scope_id 라 production store unique index 위반 시나리오와
	// 같이 검증하기 위해 fake 의 conflict 분기를 직접 트리거. firstID 와 동일
	// scope_id 인 별도 binding 을 만들어 conflict 시뮬레이션:
	dup := doJSON(t, router, http.MethodPost, "/api/v1/integration/bindings",
		`{"scope_type":"application","scope_id":"APP-001","provider_id":"prov-jira-main","external_key":"DUPLICATE-CANDIDATE","policy":"execution_system"}`)
	if dup.Code != http.StatusCreated {
		t.Fatalf("seed dup-candidate: %s", dup.Body.String())
	}
	// 이제 firstID 의 external_key 를 "DUPLICATE-CANDIDATE" 로 PATCH → 4-tuple
	// (application, APP-001, prov-jira-main, DUPLICATE-CANDIDATE) 충돌 → 409.
	rec := doJSON(t, router, http.MethodPatch, "/api/v1/integration/bindings/"+firstID,
		`{"external_key":"DUPLICATE-CANDIDATE"}`)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUpdateIntegrationBinding_ForbiddenForDeveloperRole(t *testing.T) {
	store := newMemoryApplicationStore()
	// seed provider + binding 직접 (RBAC bypass 안 되는 router 라 doJSON 사용 불가).
	if _, err := store.CreateIntegrationProvider(context.Background(), domain.IntegrationProvider{
		ID: "prov-jira-main", ProviderKey: "jira-main",
		ProviderType: domain.IntegrationProviderType("alm"), DisplayName: "Jira",
		Enabled: true, AuthMode: domain.IntegrationAuthMode("oauth2"),
		CredentialsRef: "secret://jira", SyncStatus: "requested",
	}); err != nil {
		t.Fatalf("seed provider: %v", err)
	}
	if _, err := store.CreateIntegrationBinding(context.Background(), domain.IntegrationBinding{
		ID: "bind-APP-001-PROJ", ScopeType: "application", ScopeID: "APP-001",
		ProviderID: "prov-jira-main", ExternalKey: "PROJ",
		Policy: domain.IntegrationPolicyExecutionSystem, Enabled: true,
	}); err != nil {
		t.Fatalf("seed binding: %v", err)
	}

	router := NewRouter(RouterConfig{
		ApplicationStore: store,
		BearerTokenVerifier: &fakeBearerTokenVerifier{actor: AuthenticatedActor{
			Login: "dev-user", Subject: "user-dev-user", Role: "developer",
		}},
	})

	req, _ := http.NewRequest(http.MethodPatch, "/api/v1/integration/bindings/bind-APP-001-PROJ",
		bytes.NewBufferString(`{"policy":"summary_only"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer t")
	rec := httptestDo(t, router, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

// --- DELETE /api/v1/integration/bindings/:binding_id (API-74)

func TestDeleteIntegrationBinding_Happy(t *testing.T) {
	router, bindingID := seedBindingFixture(t)
	rec := doJSON(t, router, http.MethodDelete, "/api/v1/integration/bindings/"+bindingID, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	// list 에서 사라졌는지 검증.
	list := doJSON(t, router, http.MethodGet, "/api/v1/integration/bindings", "")
	if bytes.Contains(list.Body.Bytes(), []byte(bindingID)) {
		t.Errorf("deleted binding still in list: %s", list.Body.String())
	}
	// 두 번째 DELETE 는 404 (idempotency 검증 — 같은 store 가드).
	rec2 := doJSON(t, router, http.MethodDelete, "/api/v1/integration/bindings/"+bindingID, "")
	if rec2.Code != http.StatusNotFound {
		t.Fatalf("second delete should 404, status=%d body=%s", rec2.Code, rec2.Body.String())
	}
}

func TestDeleteIntegrationBinding_NotFound(t *testing.T) {
	router := newApplicationsRouter(newMemoryApplicationStore())
	rec := doJSON(t, router, http.MethodDelete, "/api/v1/integration/bindings/bind-ghost", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"error":"binding not found"`)) {
		t.Errorf("expected 'binding not found' error: %s", rec.Body.String())
	}
}

func TestDeleteIntegrationBinding_ForbiddenForDeveloperRole(t *testing.T) {
	store := newMemoryApplicationStore()
	if _, err := store.CreateIntegrationProvider(context.Background(), domain.IntegrationProvider{
		ID: "prov-jira-main", ProviderKey: "jira-main",
		ProviderType: domain.IntegrationProviderType("alm"), DisplayName: "Jira",
		Enabled: true, AuthMode: domain.IntegrationAuthMode("oauth2"),
		CredentialsRef: "secret://jira", SyncStatus: "requested",
	}); err != nil {
		t.Fatalf("seed provider: %v", err)
	}
	if _, err := store.CreateIntegrationBinding(context.Background(), domain.IntegrationBinding{
		ID: "bind-APP-001-PROJ", ScopeType: "application", ScopeID: "APP-001",
		ProviderID: "prov-jira-main", ExternalKey: "PROJ",
		Policy: domain.IntegrationPolicyExecutionSystem, Enabled: true,
	}); err != nil {
		t.Fatalf("seed binding: %v", err)
	}

	router := NewRouter(RouterConfig{
		ApplicationStore: store,
		BearerTokenVerifier: &fakeBearerTokenVerifier{actor: AuthenticatedActor{
			Login: "dev-user", Subject: "user-dev-user", Role: "developer",
		}},
	})

	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/integration/bindings/bind-APP-001-PROJ", nil)
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

// Gitea/Forgejo/Gogs send X-Gitea-Signature (not the DevHub-native
// X-Integration-Signature). The generic per-provider ingest endpoint must
// accept those provider-native header aliases for signature/event/delivery.
func TestIntegrationProviderWebhook_GiteaNativeHeaders(t *testing.T) {
	eventStore := &dedupeEventStore{seen: map[string]bool{}}
	router := NewRouter(RouterConfig{
		ApplicationStore: newMemoryApplicationStore(),
		AuthDevFallback:  true,
		EventStore:       eventStore,
	})
	seed := doJSON(t, router, http.MethodPost, "/api/v1/integration/providers",
		`{"provider_key":"gitea-main","provider_type":"scm","display_name":"Gitea","auth_mode":"token","credentials_ref":"hmac_sha256:test-secret"}`)
	if seed.Code != http.StatusCreated {
		t.Fatalf("seed failed: %s", seed.Body.String())
	}
	body := []byte(`{"repository":{"full_name":"owner/repo"}}`)
	mac := hmac.New(sha256.New, []byte("test-secret"))
	mac.Write(body)
	signature := hex.EncodeToString(mac.Sum(nil))

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/integration/providers/gitea-main/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// Gitea-native headers only — no X-Integration-* present.
	req.Header.Set("X-Gitea-Signature", signature)
	req.Header.Set("X-Gitea-Event", "push")
	req.Header.Set("X-Gitea-Delivery", "gitea-delivery-001")
	rec := httptestDo(t, router, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}

	// Same delivery id must dedupe (event_type carried from X-Gitea-Event).
	req2, _ := http.NewRequest(http.MethodPost, "/api/v1/integration/providers/gitea-main/webhook", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("X-Gitea-Signature", signature)
	req2.Header.Set("X-Gitea-Event", "push")
	req2.Header.Set("X-Gitea-Delivery", "gitea-delivery-001")
	rec2 := httptestDo(t, router, req2)
	if rec2.Code != http.StatusConflict {
		t.Fatalf("dedupe status=%d body=%s", rec2.Code, rec2.Body.String())
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
