package gitea_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/gitea"
)

type mockSyncJobStore struct {
	repos  []domain.Repository
	users  []domain.User
	issues []domain.Issue
	pulls  []domain.PullRequest
}

func (m *mockSyncJobStore) UpsertRepository(ctx context.Context, repository domain.Repository) error {
	m.repos = append(m.repos, repository)
	return nil
}

func (m *mockSyncJobStore) UpsertUser(ctx context.Context, user domain.User) error {
	m.users = append(m.users, user)
	return nil
}

func (m *mockSyncJobStore) UpsertIssue(ctx context.Context, issue domain.Issue) error {
	m.issues = append(m.issues, issue)
	return nil
}

func (m *mockSyncJobStore) UpsertPullRequest(ctx context.Context, pr domain.PullRequest) error {
	m.pulls = append(m.pulls, pr)
	return nil
}

func (m *mockSyncJobStore) GetIntegrationProviderByID(ctx context.Context, id string) (domain.IntegrationProvider, error) {
	return domain.IntegrationProvider{
		ID:             id,
		ProviderKey:    "gitea",
		CredentialsRef: "test-token",
		Enabled:        true,
	}, nil
}

func (m *mockSyncJobStore) ListIntegrationBindings(ctx context.Context, opts any) ([]domain.IntegrationBinding, int, error) {
	return []domain.IntegrationBinding{
		{
			ID:          "binding-1",
			ScopeType:   "project",
			ScopeID:     "project-1",
			ExternalKey: "owner/test-repo",
			Enabled:     true,
		},
	}, 1, nil
}

func (m *mockSyncJobStore) AcquireNextQueuedSyncJob(ctx context.Context) (string, string, error) {
	return "", "", nil
}

func (m *mockSyncJobStore) UpdateIntegrationSyncJobStatus(ctx context.Context, jobID string, status string) error {
	return nil
}

func TestSyncWorker_ProcessOnce(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/v1/user/repos" {
			w.Write([]byte(`[{"id": 1, "name": "test-repo", "full_name": "owner/test-repo", "html_url": "http://gitea/test-repo", "clone_url": "http://gitea/test-repo.git", "default_branch": "main", "private": false}]`))
		} else if r.URL.Path == "/api/v1/repos/owner/test-repo/issues" {
			w.Write([]byte(`[{"id": 10, "number": 1, "title": "issue1", "state": "open", "html_url": "http://gitea/1", "created_at": "2026-05-26T22:00:00Z", "user": {"id": 100, "login": "user1"}}]`))
		} else if r.URL.Path == "/api/v1/repos/owner/test-repo/pulls" {
			w.Write([]byte(`[{"id": 20, "number": 2, "title": "pr1", "state": "open", "html_url": "http://gitea/2", "user": {"id": 200, "login": "user2"}}]`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	store := &mockSyncJobStore{}
	worker := gitea.NewSyncWorker(store, server.URL, "test-token")

	err := worker.ProcessOnce(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(store.repos) != 1 {
		t.Errorf("expected 1 repository to be upserted, got %d", len(store.repos))
	}
}
