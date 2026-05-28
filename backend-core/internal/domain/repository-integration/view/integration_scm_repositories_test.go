package view

import (
	"github.com/devhub/backend-core/internal/shared/httphelp"
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

// fakeGiteaServer 는 /api/v1/user/repos 에 고정 repo 1건을 응답하는 SCM 스텁.
func fakeGiteaServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/v1/user/repos" {
			_, _ = w.Write([]byte(`[{"id":1,"name":"test-repo","full_name":"owner/test-repo","clone_url":"http://gitea/owner/test-repo.git","html_url":"http://gitea/owner/test-repo","default_branch":"main","private":false}]`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func seedSCMProvider(t *testing.T, router http.Handler, key, capsJSON, baseURL string) string {
	t.Helper()
	body := `{"provider_key":"` + key + `","provider_type":"scm","display_name":"Gitea","auth_mode":"token","credentials_ref":"hmac_sha256:wh","api_token":"pat","capabilities":` + capsJSON + `,"base_url":"` + baseURL + `"}`
	rec := doJSON(t, router, http.MethodPost, "/api/v1/integration/providers", body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("seed provider failed: %s", rec.Body.String())
	}
	return "prov-" + key // memoryApplicationStore.CreateIntegrationProvider ID 규칙
}

func TestListSCMRepositories_Happy(t *testing.T) {
	srv := fakeGiteaServer(t)
	router := newApplicationsRouter(newMemoryApplicationStore())
	id := seedSCMProvider(t, router, "gitea-list", `["pull"]`, srv.URL)

	rec := doJSON(t, router, http.MethodGet, "/api/v1/integration/providers/"+id+"/scm-repositories", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"full_name":"owner/test-repo"`)) {
		t.Errorf("expected remote repo listed: %s", rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"imported":false`)) {
		t.Errorf("expected imported=false before import: %s", rec.Body.String())
	}
}

func TestListSCMRepositories_RequiresPullCapability(t *testing.T) {
	srv := fakeGiteaServer(t)
	router := newApplicationsRouter(newMemoryApplicationStore())
	// webhook capability 만 — pull 없음.
	id := seedSCMProvider(t, router, "gitea-nopull", `["webhook"]`, srv.URL)

	rec := doJSON(t, router, http.MethodGet, "/api/v1/integration/providers/"+id+"/scm-repositories", "")
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("integration_capability_not_enabled")) {
		t.Errorf("expected capability gate error: %s", rec.Body.String())
	}
}

func TestImportSCMRepositories_Happy(t *testing.T) {
	srv := fakeGiteaServer(t)
	router := newApplicationsRouter(newMemoryApplicationStore())
	id := seedSCMProvider(t, router, "gitea-import", `["pull"]`, srv.URL)

	rec := doJSON(t, router, http.MethodPost, "/api/v1/integration/providers/"+id+"/import-repositories",
		`{"full_names":["owner/test-repo"]}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"imported":1`)) {
		t.Errorf("expected imported=1: %s", rec.Body.String())
	}

	// import 후 listing 은 imported=true.
	list := doJSON(t, router, http.MethodGet, "/api/v1/integration/providers/"+id+"/scm-repositories", "")
	if !bytes.Contains(list.Body.Bytes(), []byte(`"imported":true`)) {
		t.Errorf("expected imported=true after import: %s", list.Body.String())
	}
}

func TestImportSCMRepositories_NoSelection(t *testing.T) {
	srv := fakeGiteaServer(t)
	router := newApplicationsRouter(newMemoryApplicationStore())
	id := seedSCMProvider(t, router, "gitea-empty", `["pull"]`, srv.URL)

	rec := doJSON(t, router, http.MethodPost, "/api/v1/integration/providers/"+id+"/import-repositories", `{"full_names":[]}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("integration_import_no_selection")) {
		t.Errorf("expected no-selection error: %s", rec.Body.String())
	}
}

func TestImportSCMRepositories_UnknownRepoReportedNotFound(t *testing.T) {
	srv := fakeGiteaServer(t)
	router := newApplicationsRouter(newMemoryApplicationStore())
	id := seedSCMProvider(t, router, "gitea-nf", `["pull"]`, srv.URL)

	rec := doJSON(t, router, http.MethodPost, "/api/v1/integration/providers/"+id+"/import-repositories",
		`{"full_names":["owner/does-not-exist"]}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"imported":0`)) ||
		!bytes.Contains(rec.Body.Bytes(), []byte(`"owner/does-not-exist"`)) {
		t.Errorf("expected not_found list with unknown repo: %s", rec.Body.String())
	}
}

// codex #363 P2 — credentials_ref 의 provider_sdk vendor 가 비-gitea(github)면 거부.
func TestListSCMRepositories_RejectsNonGiteaVendor(t *testing.T) {
	srv := fakeGiteaServer(t)
	router := newApplicationsRouter(newMemoryApplicationStore())
	// provider_sdk:github → Gitea-incompatible.
	body := `{"provider_key":"github-main","provider_type":"scm","display_name":"GitHub","auth_mode":"token","credentials_ref":"provider_sdk:github:wh","capabilities":["pull"],"base_url":"` + srv.URL + `"}`
	if rec := doJSON(t, router, http.MethodPost, "/api/v1/integration/providers", body); rec.Code != http.StatusCreated {
		t.Fatalf("seed failed: %s", rec.Body.String())
	}
	rec := doJSON(t, router, http.MethodGet, "/api/v1/integration/providers/prov-github-main/scm-repositories", "")
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("integration_provider_not_gitea_compatible")) {
		t.Errorf("expected gitea-compat gate error: %s", rec.Body.String())
	}
}

// fakeGiteaServerWithCreate 는 POST /api/v1/user/repos 에 생성 repo 를 응답한다.
func fakeGiteaServerWithCreate(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost && r.URL.Path == "/api/v1/user/repos" {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":99,"name":"created-repo","full_name":"alice/created-repo","clone_url":"http://gitea/alice/created-repo.git","html_url":"http://gitea/alice/created-repo","default_branch":"main","private":false}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestCreateSCMRepository_Happy(t *testing.T) {
	srv := fakeGiteaServerWithCreate(t)
	router := newApplicationsRouter(newMemoryApplicationStore())
	id := seedSCMProvider(t, router, "gitea-push", `["pull","push"]`, srv.URL)

	rec := doJSON(t, router, http.MethodPost, "/api/v1/integration/providers/"+id+"/create-repository",
		`{"name":"created-repo","private":false,"auto_init":true}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"full_name":"alice/created-repo"`)) ||
		!bytes.Contains(rec.Body.Bytes(), []byte(`"source":"system"`)) {
		t.Errorf("expected created system repo: %s", rec.Body.String())
	}
}

func TestCreateSCMRepository_RequiresPushCapability(t *testing.T) {
	srv := fakeGiteaServerWithCreate(t)
	router := newApplicationsRouter(newMemoryApplicationStore())
	// pull 만 — push 없음.
	id := seedSCMProvider(t, router, "gitea-nopush", `["pull"]`, srv.URL)

	rec := doJSON(t, router, http.MethodPost, "/api/v1/integration/providers/"+id+"/create-repository",
		`{"name":"x"}`)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("integration_capability_not_enabled")) {
		t.Errorf("expected push capability gate: %s", rec.Body.String())
	}
}

// codex #366 P2 — 비활성 provider 는 scm-repositories 작업(list/import/create) 거부.
func TestSCMRepositories_RejectsDisabledProvider(t *testing.T) {
	srv := fakeGiteaServerWithCreate(t)
	router := newApplicationsRouter(newMemoryApplicationStore())
	id := seedSCMProvider(t, router, "gitea-disabled", `["pull","push"]`, srv.URL)
	patch := doJSON(t, router, http.MethodPatch, "/api/v1/integration/providers/"+id, `{"enabled":false}`)
	if patch.Code != http.StatusOK {
		t.Fatalf("disable failed: %s", patch.Body.String())
	}
	rec := doJSON(t, router, http.MethodPost, "/api/v1/integration/providers/"+id+"/create-repository", `{"name":"x"}`)
	if rec.Code != http.StatusConflict {
		t.Fatalf("create status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("integration_provider_disabled")) {
		t.Errorf("expected disabled rejection: %s", rec.Body.String())
	}
	list := doJSON(t, router, http.MethodGet, "/api/v1/integration/providers/"+id+"/scm-repositories", "")
	if list.Code != http.StatusConflict {
		t.Errorf("list should also reject disabled: status=%d", list.Code)
	}
}

func TestCreateSCMRepository_NameRequired(t *testing.T) {
	srv := fakeGiteaServerWithCreate(t)
	router := newApplicationsRouter(newMemoryApplicationStore())
	id := seedSCMProvider(t, router, "gitea-noname", `["pull","push"]`, srv.URL)

	rec := doJSON(t, router, http.MethodPost, "/api/v1/integration/providers/"+id+"/create-repository", `{"name":"  "}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("integration_repo_name_required")) {
		t.Errorf("expected name-required error: %s", rec.Body.String())
	}
}
