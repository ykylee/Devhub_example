package gitea_test

import (
	"context"
	"encoding/base64"
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

	// queued job + per-provider config (Phase 3) — 기본 zero 값이면 기존 동작(큐 빈 주기 sync).
	acquireJobID      string
	acquireProviderID string
	providerBaseURL   string
	providerAPIToken  string
	statuses          []string

	// auth_mode 별 outbound 자격증명 (기본 zero → token mode, APIToken 사용).
	providerAuthMode     domain.IntegrationAuthMode
	providerAuthUsername string
	providerAuthSecret   string
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
		BaseURL:        m.providerBaseURL,
		APIToken:       m.providerAPIToken,
		AuthMode:       m.providerAuthMode,
		AuthUsername:   m.providerAuthUsername,
		AuthSecret:     m.providerAuthSecret,
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
	return m.acquireJobID, m.acquireProviderID, nil
}

func (m *mockSyncJobStore) UpdateIntegrationSyncJobStatus(ctx context.Context, jobID string, status string) error {
	m.statuses = append(m.statuses, status)
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

// Phase 3 — queued job 의 provider base_url+api_token 으로 sync (env 비어 있어도 동작).
func TestSyncWorker_ProcessOnce_PerProviderConfig(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/user/repos":
			w.Write([]byte(`[{"id": 1, "name": "test-repo", "full_name": "owner/test-repo", "html_url": "http://gitea/test-repo", "clone_url": "http://gitea/test-repo.git", "default_branch": "main", "private": false}]`))
		case "/api/v1/repos/owner/test-repo/issues", "/api/v1/repos/owner/test-repo/pulls":
			w.Write([]byte(`[]`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	store := &mockSyncJobStore{
		acquireJobID:      "job-1",
		acquireProviderID: "prov-1",
		providerBaseURL:   server.URL, // provider config 가 sync 대상을 결정
		providerAPIToken:  "prov-token",
	}
	// env(GiteaURL/Token) 는 빈 값 — provider config 만으로 동작해야 한다.
	worker := gitea.NewSyncWorker(store, "", "")

	if err := worker.ProcessOnce(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.repos) != 1 {
		t.Fatalf("provider config 로 sync 되어야 함, got %d repos", len(store.repos))
	}
	if len(store.statuses) != 1 || store.statuses[0] != "succeeded" {
		t.Fatalf("job 이 succeeded 로 마킹되어야 함, got %v", store.statuses)
	}
}

// Phase: auth_mode=basic provider 는 outbound 호출에 HTTP Basic 헤더를 사용한다.
func TestSyncWorker_ProcessOnce_BasicAuthOutbound(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/user/repos" {
			gotAuth = r.Header.Get("Authorization")
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/user/repos":
			w.Write([]byte(`[{"id": 1, "name": "test-repo", "full_name": "owner/test-repo", "html_url": "http://gitea/test-repo", "clone_url": "http://gitea/test-repo.git", "default_branch": "main", "private": false}]`))
		case "/api/v1/repos/owner/test-repo/issues", "/api/v1/repos/owner/test-repo/pulls":
			w.Write([]byte(`[]`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	store := &mockSyncJobStore{
		acquireJobID:         "job-1",
		acquireProviderID:    "prov-1",
		providerBaseURL:      server.URL,
		providerAuthMode:     domain.IntegrationAuthModeBasic,
		providerAuthUsername: "alice",
		providerAuthSecret:   "s3cret",
	}
	worker := gitea.NewSyncWorker(store, "", "")

	if err := worker.ProcessOnce(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "Basic " + base64.StdEncoding.EncodeToString([]byte("alice:s3cret"))
	if gotAuth != want {
		t.Fatalf("outbound Authorization=%q want=%q", gotAuth, want)
	}
	if len(store.statuses) != 1 || store.statuses[0] != "succeeded" {
		t.Fatalf("job 이 succeeded 로 마킹되어야 함, got %v", store.statuses)
	}
}
