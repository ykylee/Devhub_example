package gitea

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_ListUserRepos(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "token test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.URL.Path != "/api/v1/user/repos" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"id": 1, "name": "test-repo", "full_name": "owner/test-repo", "html_url": "http://gitea/test-repo", "clone_url": "http://gitea/test-repo.git", "default_branch": "main", "private": false}]`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	repos, err := client.ListUserRepos(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(repos))
	}
	if repos[0].Name != "test-repo" || repos[0].FullName != "owner/test-repo" {
		t.Errorf("unexpected repository details: %+v", repos[0])
	}
}

func TestClient_ListIssues(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/repos/owner/test-repo/issues" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.URL.Query().Get("state") != "all" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"id": 42, "number": 1, "title": "test-issue", "state": "open", "html_url": "http://gitea/test-repo/issues/1"}]`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "")
	issues, err := client.ListIssues(context.Background(), "owner", "test-repo", "all")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].Title != "test-issue" || issues[0].Number != 1 {
		t.Errorf("unexpected issue details: %+v", issues[0])
	}
}
