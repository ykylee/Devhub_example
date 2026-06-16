package httpapi

import (
	"net/http"
	"strings"
	"testing"
)

// Project Tests 가중치 종합 handler tests (Sprint B-Tests —
// kpi-tests-per-domain-scope.md §2.2 follow-up).
//
// 정공법: memoryPlatformStore 의 default seed-free (linked repo 0, weightedPassRate
// nil, 7 status 0, recent empty) + 1 invalid window 회귀.
//
// 1) GET /projects/:project_id/test-results — happy (memory store default).
func TestProjectTestResults_Happy(t *testing.T) {
	router := newPlatformsRouter(newMemoryPlatformStore())
	rec := doJSON(t, router, http.MethodGet, "/api/v1/projects/p-1/test-results?window=30d&limit=20", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	// response 정합: status=ok, data.project_id, data.weighted_pass_rate (null),
	// data.totals (7 status 모두 0), data.recent ([]), meta.total + meta.limit
	for _, want := range []string{
		`"status":"ok"`,
		`"project_id":"p-1"`,
		`"weighted_pass_rate":null`,
		`"totals":{`,
		`"success":0`,
		`"failed":0`,
		`"recent":[]`,
		`"total":0`,
		`"limit":20`,
		`"window_from"`,
		`"window_to"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("response missing %q: %s", want, body)
		}
	}
}

// 2) GET /projects/:project_id/test-results — invalid window → 400.
func TestProjectTestResults_InvalidWindow(t *testing.T) {
	router := newPlatformsRouter(newMemoryPlatformStore())
	rec := doJSON(t, router, http.MethodGet, "/api/v1/projects/p-1/test-results?window=bogus", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"status":"rejected"`) {
		t.Errorf("response should be rejected: %s", rec.Body.String())
	}
}

// 3) GET /projects/:project_id/test-results — invalid limit → 400.
func TestProjectTestResults_InvalidLimit(t *testing.T) {
	router := newPlatformsRouter(newMemoryPlatformStore())
	rec := doJSON(t, router, http.MethodGet, "/api/v1/projects/p-1/test-results?limit=999", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"status":"rejected"`) {
		t.Errorf("response should be rejected: %s", rec.Body.String())
	}
}
