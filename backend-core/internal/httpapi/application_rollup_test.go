package httpapi

import (
	"bytes"
	"net/http"
	"net/url"
	"testing"
)

// Application 롤업 handler tests (API-57, sprint claude/work_260514-c).

// 1) GET /applications/:id/rollup — happy (equal policy default).
func TestApplicationRollup_DefaultEqual(t *testing.T) {
	router := newApplicationsRouter(newMemoryApplicationStore())
	rec := doJSON(t, router, http.MethodGet, "/api/v1/applications/some-id/rollup", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !bytes.Contains([]byte(body), []byte(`"weight_policy":"equal"`)) {
		t.Errorf("expected default equal policy: %s", body)
	}
}

// 2) GET /applications/:id/rollup — invalid weight_policy → 400.
func TestApplicationRollup_InvalidPolicy(t *testing.T) {
	router := newApplicationsRouter(newMemoryApplicationStore())
	rec := doJSON(t, router, http.MethodGet, "/api/v1/applications/some-id/rollup?weight_policy=bogus", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

// 3) GET /applications/:id/rollup — custom weights summing to ≠ 1.0 → 422.
func TestApplicationRollup_CustomSumMismatch(t *testing.T) {
	router := newApplicationsRouter(newMemoryApplicationStore())
	weights := url.QueryEscape(`{"team/a":0.4,"team/b":0.3}`) // sum=0.7
	rec := doJSON(t, router, http.MethodGet,
		"/api/v1/applications/some-id/rollup?weight_policy=custom&custom_weights="+weights, "")
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"invalid_weight_policy"`)) {
		t.Errorf("expected invalid_weight_policy: %s", rec.Body.String())
	}
}

// 4) GET /applications/:id/rollup — malformed custom_weights JSON → 400.
func TestApplicationRollup_MalformedCustomWeights(t *testing.T) {
	router := newApplicationsRouter(newMemoryApplicationStore())
	rec := doJSON(t, router, http.MethodGet,
		"/api/v1/applications/some-id/rollup?weight_policy=custom&custom_weights=not-json", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

// 5) PR #107 codex review P1 회귀 guard — custom_weights fallback 후 합이 1.0 으로
// 정규화되어야 함. memoryApplicationStore 의 ComputeApplicationRollup stub 은 실제
// 정규화 로직을 흉내내지 않으므로 본 회귀 guard 는 실 store (PostgresStore) 의
// repository_ops.go::ComputeApplicationRollup 단위 호출로 검증해야 한다. 본 sprint 는
// store integration test 가 carve out 이므로, normalize 로직의 핵심 분기인 application
// 측 호출 (handler → store) 의 흐름은 happy path 만 검증. 실 정규화 검증은 후속
// integration test 에서 SQL pool 을 갖춘 환경에서 수행.
//
// 본 test 는 handler 의 custom_weights query 파싱이 store stub 으로 정상 전달되는지만
// 확인 — fallback 가 0건인 trivial case.
func TestApplicationRollup_CustomWeightsExact(t *testing.T) {
	router := newApplicationsRouter(newMemoryApplicationStore())
	weights := `{"team/a":1.0}` // exact 1.0
	rec := doJSON(t, router, http.MethodGet,
		"/api/v1/applications/some-id/rollup?weight_policy=custom&custom_weights="+weights, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s (expected 200 — exact sum 1.0)", rec.Code, rec.Body.String())
	}
}
