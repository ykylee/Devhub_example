package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

// Test files for sprint mvs/work_260607-h-486-ci-runs-api (N-7 / P0-4 / issue #486).
// Test IDs 매핑:
//   TC-CI-RUN-01  → TestCreateCIRunSpec_HappyPath
//   TC-CI-RUN-02  → TestCreateCIRunSpec_StatusValidation (7 PASS + 1 FAIL = 422)
//   TC-CI-RUN-03  → TestCreateCIRunSpec_RepositoryNotFound (404)
//   TC-CI-RUN-04  → TestCreateCIRunSpec_Idempotency_409
//   TC-CI-RUN-05  → TestCreateCIRunSpec_BadRequest_MissingFields (400)

func newRepoDomainStoreFixture() (*memoryDomainStore, *gin.Engine) {
	storeI := &memoryDomainStore{
		repositories: []domain.Repository{
			{ID: 100, FullName: "acme/api", Name: "api"},
			{ID: 200, FullName: "acme/web", Name: "web"},
		},
	}
	router := testRouter(RouterConfig{DomainStore: storeI})
	return storeI, router
}

// TC-CI-RUN-01: Happy path — 유효한 repository_id + ref + status + commit_sha +
// started_at + runner 가 201 + 본문 row.
func TestCreateCIRunSpec_HappyPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, router := newRepoDomainStoreFixture()

	startedAt := time.Date(2026, 6, 7, 12, 0, 0, 0, time.UTC)
	body := map[string]any{
		"repository_id": 100,
		"ref":           "main",
		"status":        "running",
		"commit_sha":    "abc1234",
		"runner":        "gitea-runner-01",
		"started_at":    startedAt.Format(time.RFC3339),
	}
	bodyJSON, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ci-runs", strings.NewReader(string(bodyJSON)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Status string `json:"status"`
		Data   struct {
			ExternalID     string `json:"external_id"`
			RepositoryID   int64  `json:"repository_id"`
			Ref            string `json:"ref"`
			Status         string `json:"status"`
			Runner         string `json:"runner"`
			RepositoryName string `json:"repository_name"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v body=%s", err, rec.Body.String())
	}
	if resp.Data.ExternalID == "" {
		t.Errorf("expected external_id in response, body=%s", rec.Body.String())
	}
	if resp.Data.RepositoryID != 100 {
		t.Errorf("expected repository_id=100, got %d", resp.Data.RepositoryID)
	}
	if resp.Data.Ref != "main" {
		t.Errorf("expected ref=main, got %q", resp.Data.Ref)
	}
	if resp.Data.Status != "running" {
		t.Errorf("expected status=running, got %q", resp.Data.Status)
	}
	if resp.Data.Runner != "gitea-runner-01" {
		t.Errorf("expected runner=gitea-runner-01, got %q", resp.Data.Runner)
	}
	if resp.Data.RepositoryName != "acme/api" {
		t.Errorf("expected repository_name=acme/api, got %q", resp.Data.RepositoryName)
	}
}

// TC-CI-RUN-02: status enum validation. 7종 모두 PASS, 1종 (invalid) → 422.
func TestCreateCIRunSpec_StatusValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, router := newRepoDomainStoreFixture()

	validStatuses := []string{"queued", "running", "success", "failed", "cancelled", "skipped", "unknown"}
	for _, status := range validStatuses {
		body := map[string]any{
			"repository_id": 100,
			"ref":           "main",
			"status":        status,
			"commit_sha":    "sha-" + status, // different sha per status so each is unique
		}
		bodyJSON, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ci-runs", strings.NewReader(string(bodyJSON)))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Errorf("status=%q expected 201, got %d body=%s", status, rec.Code, rec.Body.String())
		}
	}

	// invalid status
	body := map[string]any{
		"repository_id": 100,
		"ref":           "main",
		"status":        "nope-not-a-status",
		"commit_sha":    "sha-invalid",
	}
	bodyJSON, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ci-runs", strings.NewReader(string(bodyJSON)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("invalid status expected 422, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "create_ci_run.status_invalid") {
		t.Errorf("expected code=create_ci_run.status_invalid, body=%s", rec.Body.String())
	}
}

// TC-CI-RUN-03: 존재하지 않는 repository_id → 404.
func TestCreateCIRunSpec_RepositoryNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, router := newRepoDomainStoreFixture()

	body := map[string]any{
		"repository_id": 999, // 존재하지 않음
		"ref":           "main",
		"status":        "queued",
		"commit_sha":    "missing-repo",
	}
	bodyJSON, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ci-runs", strings.NewReader(string(bodyJSON)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "create_ci_run.repo_not_found") {
		t.Errorf("expected code=create_ci_run.repo_not_found, body=%s", rec.Body.String())
	}
}

// TC-CI-RUN-04: Idempotency — 동일 (repo, commit, status, started_at) 두번
// POST → 1st 201, 2nd 409 + 기존 row 본문.
func TestCreateCIRunSpec_Idempotency_409(t *testing.T) {
	gin.SetMode(gin.TestMode)
	storeI, router := newRepoDomainStoreFixture()

	startedAt := time.Date(2026, 6, 7, 12, 0, 0, 0, time.UTC)
	body := map[string]any{
		"repository_id": 100,
		"ref":           "main",
		"status":        "running",
		"commit_sha":    "dup-sha",
		"started_at":    startedAt.Format(time.RFC3339),
	}
	bodyJSON, _ := json.Marshal(body)

	// 1st POST → 201
	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/ci-runs", strings.NewReader(string(bodyJSON)))
	req1.Header.Set("Content-Type", "application/json")
	rec1 := httptest.NewRecorder()
	router.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusCreated {
		t.Fatalf("1st POST expected 201, got %d body=%s", rec1.Code, rec1.Body.String())
	}

	// 2nd POST 동일 body → 409 + 기존 row 본문
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/ci-runs", strings.NewReader(string(bodyJSON)))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	router.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusConflict {
		t.Fatalf("2nd POST expected 409, got %d body=%s", rec2.Code, rec2.Body.String())
	}
	if !strings.Contains(rec2.Body.String(), "create_ci_run.duplicate") {
		t.Errorf("expected code=create_ci_run.duplicate, body=%s", rec2.Body.String())
	}
	if !strings.Contains(rec2.Body.String(), "\"data\"") {
		t.Errorf("expected existing row in data field, body=%s", rec2.Body.String())
	}

	// 메모리 스토어 1 row 만 존재 (중복 INSERT 안 됨)
	if len(storeI.ciRuns) != 1 {
		t.Errorf("expected 1 ci_run in memory store, got %d", len(storeI.ciRuns))
	}
}

// TC-CI-RUN-05: 필수 필드 누락 (repository_id, ref, status) → 400.
func TestCreateCIRunSpec_BadRequest_MissingFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, router := newRepoDomainStoreFixture()

	cases := []struct {
		name string
		body string
	}{
		{"missing repository_id", `{"ref":"main","status":"queued"}`},
		{"missing ref", `{"repository_id":100,"status":"queued"}`},
		{"missing status", `{"repository_id":100,"ref":"main"}`},
		{"empty ref", `{"repository_id":100,"ref":"","status":"queued"}`},
		{"malformed JSON", `{"repository_id":100,"ref":}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/ci-runs", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("%s: expected 400, got %d body=%s", tc.name, rec.Code, rec.Body.String())
			}
		})
	}
}

// TC-CI-RUN-bonus: finished_at < started_at → 422.
func TestCreateCIRunSpec_TimeRangeValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, router := newRepoDomainStoreFixture()

	startedAt := time.Date(2026, 6, 7, 12, 0, 0, 0, time.UTC)
	finishedAt := startedAt.Add(-time.Hour) // before started_at
	body := map[string]any{
		"repository_id": 100,
		"ref":           "main",
		"status":        "success",
		"commit_sha":    "time-range",
		"started_at":    startedAt.Format(time.RFC3339),
		"finished_at":   finishedAt.Format(time.RFC3339),
	}
	bodyJSON, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ci-runs", strings.NewReader(string(bodyJSON)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for time range, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "create_ci_run.time_range") {
		t.Errorf("expected code=create_ci_run.time_range, body=%s", rec.Body.String())
	}
}

// TC-CI-RUN-bonus: idempotency key 4-tuple (repo, commit, status, start) —
// commit 만 다르면 별개 row. webhook 재시도 시나리오.
func TestCreateCIRunSpec_Idempotency_DifferentCommit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	storeI, router := newRepoDomainStoreFixture()

	startedAt := time.Date(2026, 6, 7, 12, 0, 0, 0, time.UTC)
	for i, sha := range []string{"commit-A", "commit-B"} {
		body := map[string]any{
			"repository_id": 100,
			"ref":           "main",
			"status":        "running",
			"commit_sha":    sha,
			"started_at":    startedAt.Format(time.RFC3339),
		}
		bodyJSON, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ci-runs", strings.NewReader(string(bodyJSON)))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Errorf("commit %d (%s) expected 201, got %d body=%s", i, sha, rec.Code, rec.Body.String())
		}
	}
	if len(storeI.ciRuns) != 2 {
		t.Errorf("expected 2 ci_runs in memory store, got %d", len(storeI.ciRuns))
	}
}

// TC-CI-RUN-bonus: FK violation handling — PostgresStore 가 IsForeignKeyViolation
// 반환 시 404 매핑. unit test 에서는 memory store 로 시뮬레이션 어려우므로,
// store 가 ErrNotFound 를 반환하는 경로 (= lookup miss) 만 검증. FK 매핑은
// integration test (postgres) 에서 별도 검증.
func TestCreateCIRunSpec_StoreNotConfigured(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// DomainStore 가 nil 인 router
	router := testRouter(RouterConfig{})

	body := `{"repository_id":100,"ref":"main","status":"queued","commit_sha":"x"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ci-runs", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "create_ci_run.no_store") {
		t.Errorf("expected code=create_ci_run.no_store, body=%s", rec.Body.String())
	}
}

// 4-tuple idempotency 매칭 보조 테스트 — computeCIRunExternalID 가
// deterministic 한지 + (repo, commit, status, start) 변경 시 external_id 도
// 변경되는지.
func TestComputeCIRunExternalID_Deterministic(t *testing.T) {
	startedAt := time.Date(2026, 6, 7, 12, 0, 0, 0, time.UTC)
	req := createCIRunSpecRequest{
		RepositoryID: 100,
		Ref:          "main",
		Status:       "running",
		CommitSHA:    "abc",
		StartedAt:    &startedAt,
	}
	id1 := computeCIRunExternalID(req)
	id2 := computeCIRunExternalID(req)
	if id1 != id2 {
		t.Errorf("expected deterministic, got %q vs %q", id1, id2)
	}
	if !strings.HasPrefix(id1, "ci-") {
		t.Errorf("expected ci- prefix, got %q", id1)
	}

	// commit_sha 변경 → external_id 변경
	req2 := req
	req2.CommitSHA = "different"
	if computeCIRunExternalID(req) == computeCIRunExternalID(req2) {
		t.Errorf("expected different external_id for different commit_sha")
	}
}

// GetCIRunByExternalID — memory store 구현 검증.
func TestMemoryDomainStore_GetCIRunByExternalID(t *testing.T) {
	storeI := &memoryDomainStore{
		ciRuns: []domain.CIRun{
			{ID: 1, ExternalID: "ci-aaaa", RepositoryName: "acme/api", Status: "running"},
			{ID: 2, ExternalID: "ci-bbbb", RepositoryName: "acme/web", Status: "success"},
		},
	}
	run, err := storeI.GetCIRunByExternalID(context.Background(), "ci-bbbb")
	if err != nil {
		t.Fatalf("expected hit, got err=%v", err)
	}
	if run.RepositoryName != "acme/web" {
		t.Errorf("expected acme/web, got %q", run.RepositoryName)
	}
	_, err = storeI.GetCIRunByExternalID(context.Background(), "ci-missing")
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
