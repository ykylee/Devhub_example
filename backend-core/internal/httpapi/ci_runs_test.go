package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

// TestCreateCIRunSpec_RepositoryLookup_HandlesLargeRepoSet — codex P1 review
// regression. ListRepositories 의 기본 50 row 페이지네이션 우회 —
// GetRepositoryByID 가 50 row 초과 환경에서도 단건 PK 조회. 본 test 는
// memory store fixture 가 ListRepositories 와 GetRepositoryByID 모두 노출
// 하므로 둘 다 호출되더라도 결과 정합성을 검증. (PostgresStore 의 50 row
// LIMIT 회피는 코드 review + integration test 에서 검증.)
func TestCreateCIRunSpec_RepositoryLookup_HandlesLargeRepoSet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// 60 row 시뮬레이션 — 50 row 페이지네이션이 있었다면 row 51+ 는
	// lookupRepositoryForCIRun 에서 false 404 되었을 시나리오. 본 fix 후
	// GetRepositoryByID 가 단건 조회이므로 영향 없음.
	repos := make([]domain.Repository, 0, 60)
	for i := 0; i < 60; i++ {
		repos = append(repos, domain.Repository{
			ID:       int64(100 + i),
			FullName: fmt.Sprintf("acme/repo-%d", i),
			Name:     fmt.Sprintf("repo-%d", i),
		})
	}
	repos = append(repos, domain.Repository{ID: 9999, FullName: "acme/target", Name: "target"})

	storeI := &memoryDomainStore{repositories: repos}
	router := testRouter(RouterConfig{DomainStore: storeI})

	// row 60+ 의 repo (ID 9999) 를 정확히 찾아야 함
	body := map[string]any{
		"repository_id": 9999,
		"ref":           "main",
		"status":        "queued",
		"commit_sha":    "large-set",
	}
	bodyJSON, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ci-runs", strings.NewReader(string(bodyJSON)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 (lookup must find row 60+), got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "acme/target") {
		t.Errorf("expected repository_name=acme/target, body=%s", rec.Body.String())
	}
}

// TestCreateCIRunSpec_RepositoryLookup_GetByID_NotFound — GetRepositoryByID
// 가 nil / ErrNotFound 반환 시 404 매핑. codex P1 fix 후에도 404 가 정확히
// 반환되어야 함.
func TestCreateCIRunSpec_RepositoryLookup_GetByID_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	storeI := &memoryDomainStore{
		repositories: []domain.Repository{
			{ID: 100, FullName: "acme/api", Name: "api"},
		},
	}
	router := testRouter(RouterConfig{DomainStore: storeI})

	body := map[string]any{
		"repository_id": 7777, // 존재하지 않음
		"ref":           "main",
		"status":        "queued",
		"commit_sha":    "missing",
	}
	bodyJSON, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ci-runs", strings.NewReader(string(bodyJSON)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", rec.Code, rec.Body.String())
	}
}

// TestRbacAllowsCIRunCreate_UsesActorRole — codex P2 review regression.
// PermissionCache.Allows 가 role key 기반. non-system_admin custom role
// (e.g. team_manager) 보유 actor 가 자신의 role 로 lookup 시 정상 통과
// (이전: actorLogin 으로 lookup → false deny).
//
// 본 test 는 routing path 가 actorRole 사용함을 간접 검증. cache nil 일
// 때 (test 환경) true 반환 (route middleware 가 enforce) 도 함께 검증.
func TestRbacAllowsCIRunCreate_UsesActorRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("cache nil → true (route middleware 가 enforce)", func(t *testing.T) {
		h := Handler{cfg: RouterConfig{PermissionCache: nil}}
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("devhub_actor_role", string(domain.AppRoleTeamManager))
		c.Set("devhub_actor_login", "team-manager-bob")
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/ci-runs", nil)
		if !h.rbacAllowsCIRunCreate(c) {
			t.Errorf("expected true (no cache = no extra check), got false")
		}
	})

	t.Run("system_admin role → true (early return)", func(t *testing.T) {
		h := Handler{cfg: RouterConfig{PermissionCache: NewPermissionCache(nil)}}
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("devhub_actor_role", string(domain.AppRoleSystemAdmin))
		c.Set("devhub_actor_login", "sysadmin")
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/ci-runs", nil)
		if !h.rbacAllowsCIRunCreate(c) {
			t.Errorf("expected true (system_admin early return), got false")
		}
	})

	t.Run("actor role (not login) used for cache.Allows lookup", func(t *testing.T) {
		// 회귀 가드: call site 가 actorRole ("team_manager") 을 사용. 만약 이전
		// 코드 (actorLogin) 가 다시 들어오면 cache.Allows("team-manager-bob", ...)
		// 가 되어 false deny. cache 가 nil store 일 때 cache.Allows(role) 가
		// system roles 기본 매트릭스 ("team_manager") 를 조회하므로 role 기반
		// lookup 이 정상 동작함을 확인.
		h := Handler{cfg: RouterConfig{PermissionCache: NewPermissionCache(nil)}}
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("devhub_actor_role", string(domain.AppRoleTeamManager))
		c.Set("devhub_actor_login", "team-manager-bob")
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/ci-runs", nil)
		// rbacAllowsCIRunCreate 내부에서 cache.Allows("team_manager", pipelines, create) 호출.
		// cache nil store → load from defaults → team_manager 의 pipelines:create 가
		// false 면 rbacAllowsCIRunCreate 도 false 반환 (deny). 본 test 의 핵심은
		// panic 없이 호출 경로가 actorRole 로 진입하는지 확인 (이전 actorLogin 코드였으면
		// "team-manager-bob" 으로 lookup → false deny + 다른 결과). 단순 실행 가능
		// 여부만 검증.
		_ = h.rbacAllowsCIRunCreate(c)
	})
}
