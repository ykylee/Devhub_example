package httpapi

import (
	"net/http"
	"strings"
	"testing"
)

// Project KPI 가중치 rollup handler tests (Sprint B — kpi-tests-per-domain-scope.md
// §2.2 + §6.2).
//
// 정공법: memoryPlatformStore 의 default seed-free 0,0 + linked_repository_count=0
// 가중치 0.0 응답 검증 (linked repo 0 → 가중치 metric 모두 0, JSON 정합).
//
// 1) GET /projects/:project_id/kpi — happy (memory store default).
func TestProjectKPI_Happy(t *testing.T) {
	router := newPlatformsRouter(newMemoryPlatformStore())
	rec := doJSON(t, router, http.MethodGet, "/api/v1/projects/p-1/kpi?window=30d", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	// response 정합: status=ok, data.project_id, data.linked_repository_count
	for _, want := range []string{
		`"status":"ok"`,
		`"project_id":"p-1"`,
		`"linked_repository_count":0`,
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

// 2) GET /projects/:project_id/kpi — invalid window → 400.
func TestProjectKPI_InvalidWindow(t *testing.T) {
	router := newPlatformsRouter(newMemoryPlatformStore())
	rec := doJSON(t, router, http.MethodGet, "/api/v1/projects/p-1/kpi?window=bogus", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"status":"rejected"`) {
		t.Errorf("response should be rejected: %s", rec.Body.String())
	}
}
