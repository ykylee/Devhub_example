package httpapi

import (
	"net/http"

	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

// Project KPI sub-section handler (Sprint B — kpi-tests-per-domain-scope.md §2.2
// + §6.2).
//
// 가중치 적용 rollup 정공법: project 의 N개 linked repository 의 raw metric 을
// contribution_weight 로 가중평균. 단일 repository 의 raw metric 과 다름
// (Sprint A 의 /repositories/:id/kpi 가 raw, 본 /projects/:id/kpi 가 weighted).
//
// Sprint B 1차 PR 의 scope = projectKPI 만. projectTestResults handler 는
// follow-up PR (Sprint B-Tests) 에서 동일 패턴으로 구현 (build_runs status 분포
// + 가중치 weighted_pass_rate). 본 PR 에서는 projectKPI 1 endpoint 만 router.go
// 에 등록 + routePermissionTable 의 1 row 만 추가 → deny-by-default 회귀
// 가드 정합.
//
// PR #597 P1 #2 fix 회귀: 신규 1 route 의 routePermissionTable 등록 필수.

// projectKPI — GET /api/v1/projects/:project_id/kpi
//
// Response: domain.ProjectWeightedKPI 의 JSON 직렬화 (application.go §ProjectWeightedKPI).
func (h *Handler) projectKPI(c *gin.Context) {
	storeI, ok := h.platformStoreOrUnavailable(c)
	if !ok {
		return
	}
	projectID := c.Param("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "project_id is required"})
		return
	}
	// ?window=Nd short string + RFC3339 from/to — Sprint A 의 parseTestResultsWindow
	// helper 활용 (정공법 정합). default 30d.
	windowFrom, windowTo, err := parseTestResultsWindow(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": err.Error()})
		return
	}
	opts := store.RepositoryActivityOptions{
		WindowFrom: windowFrom,
		WindowTo:   windowTo,
	}
	kpi, err := storeI.ComputeProjectWeightedKPI(c.Request.Context(), projectID, opts)
	if err != nil {
		writeServerError(c, err, "project.kpi.weighted")
		return
	}
	// PR 가중치 — 별도 query (build_runs 와 무관). store 가 raw 가중치값 반환.
	weightedOpen, weightedMerged, err := storeI.CountProjectOpenAndMergedPRs(c.Request.Context(), projectID, windowFrom, windowTo)
	if err != nil {
		writeServerError(c, err, "project.kpi.pr")
		return
	}
	kpi.WeightedOpenPRCount = weightedOpen
	kpi.WeightedMergedPRCount = weightedMerged

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   kpi,
	})
}
