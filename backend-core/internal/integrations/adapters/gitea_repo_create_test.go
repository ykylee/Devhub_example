package adapters

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

type fakeSCMStore struct {
	mu   sync.Mutex
	rows map[string]*fakeSCMRow
}

type fakeSCMRow struct {
	Status     SCMCreateStatus
	ErrMsg     string
	ExternalID int64
	CloneURL   string
	HTMLURL    string
	UpdatedAt  time.Time
}

func newFakeSCMStore() *fakeSCMStore {
	return &fakeSCMStore{rows: map[string]*fakeSCMRow{}}
}

func (s *fakeSCMStore) seed(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rows[id] = &fakeSCMRow{}
}

func (s *fakeSCMStore) UpdateSCMCreateState(_ context.Context, id string, status SCMCreateStatus, errMsg string, externalID int64, cloneURL, htmlURL string, at time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if r, ok := s.rows[id]; ok {
		r.Status = status
		r.ErrMsg = errMsg
		r.ExternalID = externalID
		r.CloneURL = cloneURL
		r.HTMLURL = htmlURL
		r.UpdatedAt = at
	}
	return nil
}

func (s *fakeSCMStore) get(id string) *fakeSCMRow {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.rows[id]
}

func TestGiteaClient_CreateUserRepo_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/user/repos", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %s; want POST", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{"id":42,"name":"new-repo","full_name":"alice/new-repo","clone_url":"https://gitea.example.com/alice/new-repo.git","html_url":"https://gitea.example.com/alice/new-repo","default_branch":"main","private":true}`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := NewGiteaClient(srv.URL, "test-token")
	repo, err := client.CreateUserRepo(context.Background(), "new-repo", GiteaRepoOptions{Private: true, AutoInit: true})
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if repo.ID != 42 {
		t.Errorf("ID = %d; want 42", repo.ID)
	}
	if repo.FullName != "alice/new-repo" {
		t.Errorf("FullName = %s; want alice/new-repo", repo.FullName)
	}
	if !repo.Private {
		t.Error("Private = false; want true")
	}
}

func TestGiteaClient_CreateUserRepo_ValidationError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/user/repos", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, `{"message":"repo name is invalid"}`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := NewGiteaClient(srv.URL, "test-token")
	_, err := client.CreateUserRepo(context.Background(), "bad name!", GiteaRepoOptions{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var ge *GiteaAPIError
	if !errors.As(err, &ge) {
		t.Fatalf("expected *GiteaAPIError, got %T: %v", err, err)
	}
	if ge.Class != "validation" {
		t.Errorf("Class = %s; want validation", ge.Class)
	}
	if ge.HTTPStatus != 422 {
		t.Errorf("HTTPStatus = %d; want 422", ge.HTTPStatus)
	}
}

func TestGiteaClient_CreateUserRepo_PermissionError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/user/repos", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, `{"message":"authentication required"}`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := NewGiteaClient(srv.URL, "invalid-token")
	_, err := client.CreateUserRepo(context.Background(), "new-repo", GiteaRepoOptions{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var ge *GiteaAPIError
	if !errors.As(err, &ge) {
		t.Fatalf("expected *GiteaAPIError")
	}
	if ge.Class != "permission" {
		t.Errorf("Class = %s; want permission", ge.Class)
	}
}

func TestGiteaClient_CreateOrgRepo_OrgNotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/orgs/nonexistent-org/repos", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, `{"message":"org not found"}`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := NewGiteaClient(srv.URL, "test-token")
	_, err := client.CreateOrgRepo(context.Background(), "nonexistent-org", "new-repo", GiteaRepoOptions{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var ge *GiteaAPIError
	if !errors.As(err, &ge) {
		t.Fatalf("expected *GiteaAPIError")
	}
	if ge.Class != "not_found" {
		t.Errorf("Class = %s; want not_found", ge.Class)
	}
}

func TestSCMCreator_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/user/repos", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{"id":42,"name":"new-repo","full_name":"alice/new-repo","clone_url":"https://gitea.example.com/alice/new-repo.git","html_url":"https://gitea.example.com/alice/new-repo","default_branch":"main","private":true}`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := NewGiteaClient(srv.URL, "test-token")
	store := newFakeSCMStore()
	store.seed("repo-1")
	var successCalled bool
	var resultFromCallback SCMCreateResult

	creator := &SCMCreator{
		Client:  client,
		Store:   store,
		Timeout: 5 * time.Second,
		OnSuccess: func(_ context.Context, r SCMCreateResult) {
			successCalled = true
			resultFromCallback = r
		},
		OnError: func(_ context.Context, r SCMCreateResult) {
			t.Errorf("OnError called on success: %+v", r)
		},
	}

	result := creator.CreateRepository(context.Background(), SCMCreateRequest{
		RepositoryID: "repo-1",
		ProjectID:    "proj-1",
		SCMProvider:  "gitea",
		RepoName:     "new-repo",
		Private:      true,
		AutoInit:     true,
	})

	if result.Status != SCMCreateSuccess {
		t.Errorf("Status = %s; want success", result.Status)
	}
	if result.ExternalID != 42 {
		t.Errorf("ExternalID = %d; want 42", result.ExternalID)
	}
	if !successCalled {
		t.Error("OnSuccess was not called")
	}
	if resultFromCallback.Status != SCMCreateSuccess {
		t.Errorf("callback Status = %s; want success", resultFromCallback.Status)
	}
	if store.get("repo-1").Status != SCMCreateSuccess {
		t.Errorf("store status = %s; want success", store.get("repo-1").Status)
	}
	if !strings.HasPrefix(store.get("repo-1").CloneURL, "https://") {
		t.Errorf("CloneURL = %s; want https://...", store.get("repo-1").CloneURL)
	}
}

func TestSCMCreator_Failure_BestEffort(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/user/repos", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `internal error`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := NewGiteaClient(srv.URL, "test-token")
	store := newFakeSCMStore()
	store.seed("repo-1")
	var errorCalled, compensationCalled bool
	var resultFromCallback SCMCreateResult
	_ = resultFromCallback

	creator := &SCMCreator{
		Client:  client,
		Store:   store,
		Timeout: 5 * time.Second,
		OnError: func(_ context.Context, r SCMCreateResult) {
			errorCalled = true
			resultFromCallback = r
		},
		OnCompensation: func(_ context.Context, r SCMCreateResult) {
			compensationCalled = true
			resultFromCallback = r
		},
	}

	result := creator.CreateRepository(context.Background(), SCMCreateRequest{
		RepositoryID: "repo-1",
		SCMProvider:  "gitea",
		RepoName:     "new-repo",
	})

	if result.Status != SCMCreateFailed {
		t.Errorf("Status = %s; want failed", result.Status)
	}
	if !errorCalled {
		t.Error("OnError was not called")
	}
	if !compensationCalled {
		t.Error("OnCompensation was not called")
	}
	if result.CompensationAction != "retry_scheduled" {
		t.Errorf("CompensationAction = %s; want retry_scheduled", result.CompensationAction)
	}
	if result.NextRetryAt == nil {
		t.Error("NextRetryAt should be set")
	}
	if result.HTTPStatus != 500 {
		t.Errorf("HTTPStatus = %d; want 500", result.HTTPStatus)
	}
	if result.ErrorClass != "server" {
		t.Errorf("ErrorClass = %s; want server", result.ErrorClass)
	}
	if store.get("repo-1").Status != SCMCreateFailed {
		t.Errorf("store status = %s; want failed", store.get("repo-1").Status)
	}
}

func TestSCMCreator_NilClient(t *testing.T) {
	store := newFakeSCMStore()
	store.seed("repo-1")
	creator := &SCMCreator{
		Client:  nil,
		Store:   store,
		Timeout: 5 * time.Second,
	}
	result := creator.CreateRepository(context.Background(), SCMCreateRequest{
		RepositoryID: "repo-1",
		SCMProvider:  "gitea",
		RepoName:     "new-repo",
	})
	if result.Status != SCMCreateFailed {
		t.Errorf("Status = %s; want failed (nil client)", result.Status)
	}
	if result.ErrorClass != "config" {
		t.Errorf("ErrorClass = %s; want config", result.ErrorClass)
	}
}

func TestGiteaErrorClass(t *testing.T) {
	tests := []struct {
		status int
		want   string
	}{
		{401, "permission"},
		{403, "permission"},
		{404, "not_found"},
		{422, "validation"},
		{429, "rate_limit"},
		{500, "server"},
		{502, "server"},
		{503, "server"},
		{418, "unknown"},
	}
	for _, tt := range tests {
		got := giteaErrorClass(tt.status)
		if got != tt.want {
			t.Errorf("giteaErrorClass(%d) = %s; want %s", tt.status, got, tt.want)
		}
	}
}

func TestSCMCreator_Timeout(t *testing.T) {
	// Server that hangs
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/user/repos", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := NewGiteaClient(srv.URL, "test-token")
	store := newFakeSCMStore()
	store.seed("repo-1")
	creator := &SCMCreator{
		Client:  client,
		Store:   store,
		Timeout: 100 * time.Millisecond, // very short timeout
	}
	start := time.Now()
	result := creator.CreateRepository(context.Background(), SCMCreateRequest{
		RepositoryID: "repo-1",
		SCMProvider:  "gitea",
		RepoName:     "new-repo",
	})
	elapsed := time.Since(start)
	if elapsed > 2*time.Second {
		t.Errorf("timeout took too long: %s; want < 2s", elapsed)
	}
	if result.Status != SCMCreateFailed {
		t.Errorf("Status = %s; want failed (timeout)", result.Status)
	}
}
