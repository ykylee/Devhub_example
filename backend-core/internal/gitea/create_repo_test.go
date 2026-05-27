package gitea_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/devhub/backend-core/internal/gitea"
)

func TestClient_CreateRepo_UserScope(t *testing.T) {
	var gotPath, gotAuth string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":42,"name":"new-repo","full_name":"alice/new-repo","clone_url":"http://gitea/alice/new-repo.git","html_url":"http://gitea/alice/new-repo","default_branch":"main","private":true}`))
	}))
	defer srv.Close()

	client := gitea.NewClient(srv.URL, "pat-token")
	created, err := client.CreateRepo(context.Background(), "", gitea.CreateRepoOptions{Name: "new-repo", Private: true, AutoInit: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/api/v1/user/repos" {
		t.Errorf("owner 빈값은 /user/repos 여야 함, got %q", gotPath)
	}
	if gotAuth != "token pat-token" {
		t.Errorf("auth header=%q", gotAuth)
	}
	if gotBody["name"] != "new-repo" || gotBody["private"] != true {
		t.Errorf("body mismatch: %v", gotBody)
	}
	if created.FullName != "alice/new-repo" || created.ID != 42 {
		t.Errorf("created mismatch: %+v", created)
	}
}

func TestClient_CreateRepo_OrgScope(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":7,"name":"svc","full_name":"team/svc","default_branch":"main"}`))
	}))
	defer srv.Close()

	client := gitea.NewClient(srv.URL, "t")
	if _, err := client.CreateRepo(context.Background(), "team", gitea.CreateRepoOptions{Name: "svc"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/api/v1/orgs/team/repos" {
		t.Errorf("owner 지정 시 /orgs/{owner}/repos 여야 함, got %q", gotPath)
	}
}

func TestClient_CreateRepo_ConflictError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
	}))
	defer srv.Close()

	client := gitea.NewClient(srv.URL, "t")
	if _, err := client.CreateRepo(context.Background(), "", gitea.CreateRepoOptions{Name: "dup"}); err == nil {
		t.Fatal("expected error on 409 conflict")
	}
}
