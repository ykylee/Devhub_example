package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/gin-gonic/gin"
)

func TestRepositoryBuildRuns_ReturnsCIRunsData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	startedAt := time.Date(2026, 6, 2, 10, 0, 0, 0, time.UTC)
	storeI := &memoryPlatformStore{
		repositories: map[string]domain.Repository{
			"acme/api": {ID: 101, FullName: "acme/api"},
		},
		nextRepositoryID: 200,
		ciRuns: []domain.BuildRun{
			{
				ID: 1, RepositoryID: 101, RunExternalID: "ci-001",
				Branch: "main", CommitSHA: "abc123", Status: "success",
				DurationSeconds: intPtr(120), StartedAt: startedAt,
			},
			{
				ID: 2, RepositoryID: 101, RunExternalID: "ci-002",
				Branch: "feature/x", CommitSHA: "def456", Status: "failed",
				DurationSeconds: intPtr(45), StartedAt: startedAt.Add(-5 * time.Minute),
			},
		},
	}
	router := testRouter(RouterConfig{PlatformStore: storeI})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/repositories/101/build-runs", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	if !containsAll(body, "ci-001", "ci-002", "success", "failed") {
		t.Errorf("response body missing expected ci-run data:\n%s", body)
	}
}

func TestRepositoryBuildRuns_NoCIRuns(t *testing.T) {
	gin.SetMode(gin.TestMode)
	storeI := &memoryPlatformStore{
		repositories: map[string]domain.Repository{
			"acme/api": {ID: 101, FullName: "acme/api"},
		},
		nextRepositoryID: 200,
	}
	router := testRouter(RouterConfig{PlatformStore: storeI})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/repositories/101/build-runs", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !containsAll(rec.Body.String(), `"data":[]`) {
		t.Errorf("expected empty data array, got %s", rec.Body.String())
	}
}

func TestRepositoryBuildRuns_StatusFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	startedAt := time.Date(2026, 6, 2, 10, 0, 0, 0, time.UTC)
	storeI := &memoryPlatformStore{
		repositories: map[string]domain.Repository{
			"acme/api": {ID: 101, FullName: "acme/api"},
		},
		nextRepositoryID: 200,
		ciRuns: []domain.BuildRun{
			{
				ID: 1, RepositoryID: 101, RunExternalID: "ci-001",
				Branch: "main", CommitSHA: "abc123", Status: "success",
				DurationSeconds: intPtr(120), StartedAt: startedAt,
			},
			{
				ID: 2, RepositoryID: 101, RunExternalID: "ci-002",
				Branch: "feature/x", CommitSHA: "def456", Status: "failed",
				DurationSeconds: intPtr(45), StartedAt: startedAt.Add(-5 * time.Minute),
			},
		},
	}
	router := testRouter(RouterConfig{PlatformStore: storeI})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/repositories/101/build-runs?status=failed", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if containsAll(rec.Body.String(), "ci-001") {
		t.Errorf("expected only failed ci-run, got success ci-run:\n%s", rec.Body.String())
	}
	if !containsAll(rec.Body.String(), "ci-002", "failed") {
		t.Errorf("response missing failed ci-run:\n%s", rec.Body.String())
	}
}

// --- helpers ---

func intPtr(n int) *int { return &n }

func containsAll(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if !strings.Contains(s, sub) {
			return false
		}
	}
	return true
}
