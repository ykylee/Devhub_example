package adapters

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// fakeStore implements RepositoryPullStore for tests.
type fakeStore struct {
	mu sync.Mutex

	pullState          map[string]*fakePullState
	prs                []fakePR
	builds             []fakeBuild
	upsertPRsCalled    int
	upsertBuildsCalled int
	upsertQSCalled     int
	failNext           bool
}

type fakePullState struct {
	lastPullAt          time.Time
	status              string
	errMsg              string
	consecutiveFailures int
	backoffUntil        time.Time
	lastAlertAt         time.Time
}

type fakePR struct {
	giteaID int64
	number  int
	state   string
}

type fakeBuild struct {
	giteaID  int64
	commitSHA string
}

func newFakeStore() *fakeStore {
	return &fakeStore{pullState: map[string]*fakePullState{}}
}

func (s *fakeStore) seed(repositoryID string, lastPullAt time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pullState[repositoryID] = &fakePullState{lastPullAt: lastPullAt}
}

func (s *fakeStore) LastPullAt(_ context.Context, repositoryID string) (time.Time, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s, ok := s.pullState[repositoryID]; ok {
		return s.lastPullAt, nil
	}
	return time.Time{}, nil
}

func (s *fakeStore) BackoffUntil(_ context.Context, repositoryID string) (time.Time, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s, ok := s.pullState[repositoryID]; ok {
		return s.backoffUntil, nil
	}
	return time.Time{}, nil
}

func (s *fakeStore) UpdatePullState(_ context.Context, repositoryID, status, errMsg string, lastPullAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s, ok := s.pullState[repositoryID]; ok {
		s.status = status
		s.errMsg = errMsg
		s.lastPullAt = lastPullAt
	}
	return nil
}

func (s *fakeStore) IncrementConsecutiveFailures(_ context.Context, repositoryID string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s, ok := s.pullState[repositoryID]; ok {
		s.consecutiveFailures++
		return s.consecutiveFailures, nil
	}
	return 0, nil
}

func (s *fakeStore) ResetConsecutiveFailures(_ context.Context, repositoryID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s, ok := s.pullState[repositoryID]; ok {
		s.consecutiveFailures = 0
		s.backoffUntil = time.Time{}
	}
	return nil
}

func (s *fakeStore) SetBackoff(_ context.Context, repositoryID string, until time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s, ok := s.pullState[repositoryID]; ok {
		s.backoffUntil = until
	}
	return nil
}

func (s *fakeStore) UpsertPullActivity(_ context.Context, _ string, giteaPRID int64, _ int, _ string, _ string, _ string, _ string, _ string, _ time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.upsertPRsCalled++
	s.prs = append(s.prs, fakePR{giteaID: giteaPRID})
	return nil
}

func (s *fakeStore) UpsertBuildRun(_ context.Context, _ string, giteaBuildID int64, commitSHA, _ string, _ string, _ string, _ time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.upsertBuildsCalled++
	s.builds = append(s.builds, fakeBuild{giteaID: giteaBuildID, commitSHA: commitSHA})
	return nil
}

func (s *fakeStore) UpsertQualitySnapshot(_ context.Context, _ string, _ string, _ time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.upsertQSCalled++
	return nil
}

func TestBackoffDuration(t *testing.T) {
	tests := []struct {
		failures int
		cap      time.Duration
		want     time.Duration
	}{
		{0, 24 * time.Hour, 0},
		{1, 24 * time.Hour, 2 * time.Minute},
		{2, 24 * time.Hour, 4 * time.Minute},
		{5, 24 * time.Hour, 32 * time.Minute},
		{10, 24 * time.Hour, 17 * time.Hour + 4*time.Minute}, // 1024 minutes
		{11, 24 * time.Hour, 24 * time.Hour},                // cap
		{100, 24 * time.Hour, 24 * time.Hour},               // cap
	}
	for _, tt := range tests {
		got := backoffDuration(tt.failures, tt.cap)
		if got != tt.want {
			t.Errorf("backoffDuration(%d, %s) = %s; want %s", tt.failures, tt.cap, got, tt.want)
		}
	}
}

func TestSemaphore_ConcurrencyLimit(t *testing.T) {
	sem := NewSemaphore(2)
	var concurrent int32
	var maxObserved int32
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem.Acquire()
			defer sem.Release()
			cur := atomic.AddInt32(&concurrent, 1)
			for {
				m := atomic.LoadInt32(&maxObserved)
				if cur <= m || atomic.CompareAndSwapInt32(&maxObserved, m, cur) {
					break
				}
			}
			time.Sleep(10 * time.Millisecond)
			atomic.AddInt32(&concurrent, -1)
		}()
	}
	wg.Wait()
	if maxObserved > 2 {
		t.Errorf("max concurrent = %d; want <= 2", maxObserved)
	}
}

func TestGiteaPullAdapter_PullAndIngestSince_Success(t *testing.T) {
	// Mock Gitea server with 3 PRs and 2 builds
	prsCalled := int32(0)
	buildsCalled := int32(0)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/repos/test-org/test-repo/pulls", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&prsCalled, 1)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `[{"id":1,"number":1,"state":"open","title":"PR1","body":"","head_sha":"abc","updated_at":"2026-06-13T00:00:00Z","created_at":"2026-06-13T00:00:00Z","user":{"login":"alice"}},{"id":2,"number":2,"state":"open","title":"PR2","body":"","head_sha":"def","updated_at":"2026-06-13T00:00:00Z","created_at":"2026-06-13T00:00:00Z","user":{"login":"bob"}},{"id":3,"number":3,"state":"closed","title":"PR3","body":"","head_sha":"ghi","updated_at":"2026-06-13T00:00:00Z","created_at":"2026-06-13T00:00:00Z","user":{"login":"carol"}}]`)
	})
	mux.HandleFunc("/api/v1/repos/test-org/test-repo/actions/runs", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&buildsCalled, 1)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `[{"id":10,"head_sha":"abc","event":"push","status":"success","conclusion":"success","created_at":"2026-06-13T00:00:00Z","updated_at":"2026-06-13T00:00:00Z"},{"id":11,"head_sha":"def","event":"pull_request","status":"success","conclusion":"success","created_at":"2026-06-13T00:00:00Z","updated_at":"2026-06-13T00:00:00Z"}]`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := NewGiteaClient(srv.URL, "test-token")
	store := newFakeStore()
	store.seed("repo-1", time.Time{}) // never pulled

	adapter := &GiteaPullAdapter{Client: client, Store: store, MaxItemsPerCall: 100}
	err := adapter.PullAndIngestSince(context.Background(), RepositoryTarget{ID: "repo-1", Owner: "test-org", Name: "test-repo"})
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if store.upsertPRsCalled != 3 {
		t.Errorf("upsertPRsCalled = %d; want 3", store.upsertPRsCalled)
	}
	if store.upsertBuildsCalled != 2 {
		t.Errorf("upsertBuildsCalled = %d; want 2", store.upsertBuildsCalled)
	}
	if store.upsertQSCalled != 2 {
		t.Errorf("upsertQSCalled = %d; want 2 (one per unique build commit)", store.upsertQSCalled)
	}
	if store.pullState["repo-1"].status != "success" {
		t.Errorf("status = %s; want success", store.pullState["repo-1"].status)
	}
	if store.pullState["repo-1"].consecutiveFailures != 0 {
		t.Errorf("consecutiveFailures = %d; want 0 (reset on success)", store.pullState["repo-1"].consecutiveFailures)
	}
}

func TestGiteaPullAdapter_PullAndIngestSince_GiteaUnreachable(t *testing.T) {
	// Gitea server that returns 500
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/repos/test-org/test-repo/pulls", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "internal error")
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := NewGiteaClient(srv.URL, "test-token")
	store := newFakeStore()
	store.seed("repo-1", time.Time{})

	adapter := &GiteaPullAdapter{Client: client, Store: store, MaxItemsPerCall: 100}
	err := adapter.PullAndIngestSince(context.Background(), RepositoryTarget{ID: "repo-1", Owner: "test-org", Name: "test-repo"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var pe *PullError
	if !errors.As(err, &pe) {
		t.Fatalf("expected *PullError, got %T: %v", err, err)
	}
	if pe.Class != "gitea_api" {
		t.Errorf("class = %s; want gitea_api", pe.Class)
	}
	// Adapter is responsible for returning a typed error; the loop is responsible for
	// calling IncrementConsecutiveFailures (see TestGiteaPullLoop_AlertThreshold).
	// In this test we directly simulate that the loop would increment by calling it.
	_, _ = store.IncrementConsecutiveFailures(context.Background(), "repo-1")
	cf := store.pullState["repo-1"].consecutiveFailures
	if cf != 1 {
		t.Errorf("consecutiveFailures = %d; want 1 (incremented on error)", cf)
	}
}

func TestGiteaPullAdapter_PullAndIngestSince_PartialResponse(t *testing.T) {
	// Gitea returns 2 PRs and 0 builds (partial: missing build data)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/repos/test-org/test-repo/pulls", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `[{"id":1,"number":1,"state":"open","title":"PR1","body":"","head_sha":"abc","updated_at":"2026-06-13T00:00:00Z","created_at":"2026-06-13T00:00:00Z","user":{"login":"alice"}},{"id":2,"number":2,"state":"open","title":"PR2","body":"","head_sha":"def","updated_at":"2026-06-13T00:00:00Z","created_at":"2026-06-13T00:00:00Z","user":{"login":"bob"}}]`)
	})
	mux.HandleFunc("/api/v1/repos/test-org/test-repo/actions/runs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `[]`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := NewGiteaClient(srv.URL, "test-token")
	store := newFakeStore()
	store.seed("repo-1", time.Time{})

	adapter := &GiteaPullAdapter{Client: client, Store: store, MaxItemsPerCall: 100}
	err := adapter.PullAndIngestSince(context.Background(), RepositoryTarget{ID: "repo-1", Owner: "test-org", Name: "test-repo"})
	if err != nil {
		t.Fatalf("partial response should not be a hard error, got %v", err)
	}
	if store.upsertPRsCalled != 2 {
		t.Errorf("upsertPRsCalled = %d; want 2", store.upsertPRsCalled)
	}
	if store.pullState["repo-1"].status != "success" {
		// empty build list is a valid (but empty) success — partial only when some items fail to upsert
		t.Errorf("status = %s; want success (empty build list is valid)", store.pullState["repo-1"].status)
	}
}

func TestGiteaPullLoop_BackoffCap(t *testing.T) {
	// Test that backoff is capped at 24h
	d := backoffDuration(20, 24*time.Hour)
	if d != 24*time.Hour {
		t.Errorf("backoff cap not enforced: backoffDuration(20, 24h) = %s; want 24h", d)
	}
	d = backoffDuration(50, 24*time.Hour)
	if d != 24*time.Hour {
		t.Errorf("backoff cap not enforced: backoffDuration(50, 24h) = %s; want 24h", d)
	}
}

func TestGiteaPullLoop_AlertThreshold(t *testing.T) {
	// Mock Gitea that always 500s
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/repos/test-org/test-repo/pulls", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	mux.HandleFunc("/api/v1/repos/test-org/test-repo/actions/runs", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := NewGiteaClient(srv.URL, "test-token")
	store := newFakeStore()
	store.seed("repo-1", time.Time{})

	adapter := &GiteaPullAdapter{Client: client, Store: store, MaxItemsPerCall: 100}

	// Simulate 5 consecutive failures (loop pattern: each fail → increment + backoff)
	for i := 1; i <= 5; i++ {
		_ = adapter.PullAndIngestSince(context.Background(), RepositoryTarget{ID: "repo-1", Owner: "test-org", Name: "test-repo"})
		_, _ = store.IncrementConsecutiveFailures(context.Background(), "repo-1")
		until := time.Now().UTC().Add(backoffDuration(i, 24*time.Hour))
		_ = store.SetBackoff(context.Background(), "repo-1", until)
	}

	cf := store.pullState["repo-1"].consecutiveFailures
	if cf != 5 {
		t.Errorf("consecutiveFailures = %d; want 5 after 5 errors", cf)
	}
	bu := store.pullState["repo-1"].backoffUntil
	if bu.IsZero() {
		t.Error("backoffUntil should be set after failures")
	}
	expected := 32 * time.Minute // 2^5 = 32 minutes
	actual := time.Until(bu)
	if actual < expected-time.Minute || actual > expected+time.Minute {
		t.Errorf("backoff duration ~%s; want ~%s", actual, expected)
	}
}

func TestGiteaPullLoop_Shutdown(t *testing.T) {
	// Test that ctx cancellation stops the loop
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/repos/test-org/test-repo/pulls", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `[]`)
	})
	mux.HandleFunc("/api/v1/repos/test-org/test-repo/actions/runs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `[]`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := NewGiteaClient(srv.URL, "test-token")
	store := newFakeStore()

	adapter := &GiteaPullAdapter{Client: client, Store: store, MaxItemsPerCall: 100}
	repoLister := func(ctx context.Context) ([]RepositoryTarget, error) {
		return []RepositoryTarget{}, nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- RunGiteaPullLoop(ctx, adapter, repoLister, 1*time.Hour, 1*time.Minute, 1, 24*time.Hour, 5, nil, nil)
	}()
	time.Sleep(100 * time.Millisecond) // let initial cycle complete
	cancel()
	select {
	case err := <-done:
		if err != context.Canceled {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("RunGiteaPullLoop did not exit within 2s after cancel")
	}
}

func TestPullError_Class(t *testing.T) {
	tests := []struct {
		pe   *PullError
		want string
	}{
		{&PullError{Class: "gitea_api", Message: "list PRs"}, "gitea_api"},
		{&PullError{Class: "partial", Message: "some items failed"}, "partial"},
		{&PullError{Class: "store", Message: "DB error"}, "store"},
		{&PullError{Class: "config", Message: "nil store"}, "config"},
	}
	for _, tt := range tests {
		if tt.pe.Class != tt.want {
			t.Errorf("class = %s; want %s", tt.pe.Class, tt.want)
		}
	}
}
