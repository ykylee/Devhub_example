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
