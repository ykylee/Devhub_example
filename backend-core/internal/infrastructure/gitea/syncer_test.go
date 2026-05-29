package gitea

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/devhub/backend-core/internal/domain"
)

type mockSyncStore struct {
	repos  []domain.Repository
	users  []domain.User
	issues []domain.Issue
	pulls  []domain.PullRequest
}

func (m *mockSyncStore) UpsertRepository(ctx context.Context, repository domain.Repository) error {
	m.repos = append(m.repos, repository)
	return nil
}

func (m *mockSyncStore) UpsertUser(ctx context.Context, user domain.User) error {
	m.users = append(m.users, user)
	return nil
}

func (m *mockSyncStore) UpsertIssue(ctx context.Context, issue domain.Issue) error {
	m.issues = append(m.issues, issue)
	return nil
}

func (m *mockSyncStore) UpsertPullRequest(ctx context.Context, pr domain.PullRequest) error {
	m.pulls = append(m.pulls, pr)
	return nil
}

func TestSyncer_SyncRepository(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/v1/repos/owner/test-repo/issues" {
			w.Write([]byte(`[{"id": 10, "number": 1, "title": "issue1", "state": "open", "html_url": "http://gitea/1", "created_at": "2026-05-26T22:00:00Z", "user": {"id": 100, "login": "user1"}}]`))
		} else if r.URL.Path == "/api/v1/repos/owner/test-repo/pulls" {
			w.Write([]byte(`[{"id": 20, "number": 2, "title": "pr1", "state": "open", "html_url": "http://gitea/2", "user": {"id": 200, "login": "user2"}}]`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "")
	store := &mockSyncStore{}
	syncer := NewSyncer(store)

	err := syncer.SyncRepository(context.Background(), client, "owner", "test-repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(store.users) != 4 {
		t.Fatalf("expected 4 users upserted, got %d", len(store.users))
	}
	if len(store.issues) != 2 { // listIssues twice (open, closed) -> 2 issues loaded
		t.Logf("issues: %d", len(store.issues))
	}
	if len(store.pulls) != 2 { // listPullRequests twice (open, closed) -> 2 pulls loaded
		t.Logf("pulls: %d", len(store.pulls))
	}
}
