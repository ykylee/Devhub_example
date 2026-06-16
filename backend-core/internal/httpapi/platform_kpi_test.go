package httpapi

import (
	"net/http"
	"strings"
	"testing"
)

// Platform KPI sub-project rollup handler tests (Sprint C —
// kpi-tests-per-domain-scope.md §2.3 + §6.3).
//
// 정공법: memoryPlatformStore 의 default seed-free (linked project 0, weighted
// metric 0) + 1 invalid window 회귀.

// 1) GET /platforms/:platform_id/kpi — happy (memory store default).
func TestPlatformKPI_Happy(t *testing.T) {
	router := newPlatformsRouter(newMemoryPlatformStore())
	rec := doJSON(t, router, http.MethodGet, "/api/v1/platforms/pl-1/kpi?window=30d", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	// response 정합: status=ok, data.platform_id, data.linked_project_count=0,
	// data.weighted_quality_score=0, data.weighted_build_success_rate=0
	for _, want := range []string{
		`"status":"ok"`,
		`"platform_id":"pl-1"`,
		`"linked_project_count":0`,
		`"weighted_quality_score":0`,
		`"weighted_build_success_rate":0`,
		`"window_from"`,
		`"window_to"`,
		`"weighted_at"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("response missing %q: %s", want, body)
		}
	}
}

// 2) GET /platforms/:platform_id/kpi — invalid window → 400.
func TestPlatformKPI_InvalidWindow(t *testing.T) {
	router := newPlatformsRouter(newMemoryPlatformStore())
	rec := doJSON(t, router, http.MethodGet, "/api/v1/platforms/pl-1/kpi?window=bogus", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"status":"rejected"`) {
		t.Errorf("response should be rejected: %s", rec.Body.String())
	}
}
