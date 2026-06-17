package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
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

func TestRepositoryBuildRuns_UnknownRepoReturns404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	storeI := &memoryPlatformStore{
		repositories: map[string]domain.Repository{
			"acme/api": {ID: 101, FullName: "acme/api"},
		},
		nextRepositoryID: 200,
	}
	router := testRouter(RouterConfig{PlatformStore: storeI})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/repositories/999/build-runs", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !containsAll(rec.Body.String(), "not_found", errRepositoryNotFound) {
		t.Errorf("expected not_found response, got %s", rec.Body.String())
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


// #557 (N-9 sub-2): Histogram metric assertion — devhub_repository_build_runs_query_duration_seconds
// {status_filter=<status>} label 의 observe 가 handler 호출 시점에 발생함을 검증.
// 2 case: (a) status filter 있음 (label = "success") (b) status filter 없음 (label = "").
func TestRepositoryBuildRuns_HistogramObserved_WithStatusFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	initBuildRunsMetrics() // test setup — lazy init (production observe 가 이미 발생했을 수 있다)
	storeI := &memoryPlatformStore{
		repositories: map[string]domain.Repository{
			"acme/api": {ID: 101, FullName: "acme/api"},
		},
		nextRepositoryID: 200,
		ciRuns: []domain.BuildRun{
			{ID: 1, RepositoryID: 101, RunExternalID: "ci-001", Branch: "main", CommitSHA: "abc123", Status: "success", StartedAt: time.Now()},
		},
	}
	router := testRouter(RouterConfig{PlatformStore: storeI})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/repositories/101/build-runs?status=success", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if buildRunsDuration == nil {
		t.Fatal("buildRunsDuration not initialized after request")
	}
	// Histogram observe 가 1회 이상 발생 + label "status_filter=success" 가 등록됨을 검증
	// testutil.ToFloat64 는 Histogram 미지원 (counter/gauge/untyped only) → dto.Metric 직접 read
	hist, ok := buildRunsDuration.WithLabelValues("success").(prometheus.Histogram)
	if !ok {
		t.Fatal("buildRunsDuration.WithLabelValues did not return prometheus.Histogram")
	}
	mProto := &dto.Metric{}
	if err := hist.Write(mProto); err != nil {
		t.Fatalf("histogram.Write: %v", err)
	}
	if mProto.GetHistogram() == nil {
		t.Fatal("expected histogram proto, got nil")
	}
	if mProto.GetHistogram().GetSampleCount() < 1 {
		t.Errorf("expected sample_count >= 1, got %d", mProto.GetHistogram().GetSampleCount())
	}
	if mProto.GetHistogram().GetSampleSum() < 0 {
		t.Errorf("expected sample_sum >= 0, got %f", mProto.GetHistogram().GetSampleSum())
	}
}

func TestRepositoryBuildRuns_HistogramObserved_NoStatusFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	initBuildRunsMetrics()
	storeI := &memoryPlatformStore{
		repositories: map[string]domain.Repository{
			"acme/api": {ID: 101, FullName: "acme/api"},
		},
		nextRepositoryID: 200,
		ciRuns:          []domain.BuildRun{},
	}
	router := testRouter(RouterConfig{PlatformStore: storeI})

	// status filter 없이 호출 — label = "" (empty string)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/repositories/101/build-runs", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if buildRunsDuration == nil {
		t.Fatal("buildRunsDuration not initialized after request")
	}
	// label "" 가 등록됨을 검증 (status filter 부재 case)
	hist, ok := buildRunsDuration.WithLabelValues("").(prometheus.Histogram)
	if !ok {
		t.Fatal("buildRunsDuration.WithLabelValues did not return prometheus.Histogram")
	}
	mProto := &dto.Metric{}
	if err := hist.Write(mProto); err != nil {
		t.Fatalf("histogram.Write: %v", err)
	}
	if mProto.GetHistogram() == nil {
		t.Fatal("expected histogram proto, got nil")
	}
	if mProto.GetHistogram().GetSampleCount() < 1 {
		t.Errorf("expected sample_count >= 1 for status=\"\" (no filter), got %d", mProto.GetHistogram().GetSampleCount())
	}
	if mProto.GetHistogram().GetSampleSum() < 0 {
		t.Errorf("expected sample_sum >= 0 for status=\"\" (no filter), got %f", mProto.GetHistogram().GetSampleSum())
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
