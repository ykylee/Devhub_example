package httpapi

import (
	"net/http"
	"strconv"

	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

// projectTestResults handler — Sprint B-Tests (kpi-tests-per-domain-scope.md §2.2
// follow-up).
//
// 가중치 적용 test results 정공법: project 의 N개 linked repository 의 build_runs
// status 를 종합 + 가중치 pass rate. 단일 repository 의 raw metric 과 다름
// (Sprint A 의 /repositories/:id/test-results 가 raw, 본 /projects/:id/test-results
// 가 weighted).
//
// Sprint B 1차 PR #627 의 scope = projectKPI 만. 본 handler 는 Sprint B-Tests 의
// follow-up. routePermissionTable 에 /projects/:project_id/test-results row 등록
// 필수 (deny-by-default 회귀 가드 정합).

// projectTestResults — GET /api/v1/projects/:project_id/test-results
//
// Response: domain.ProjectWeightedTestResults 의 JSON 직렬화 + meta (total, limit).
func (h *Handler) projectTestResults(c *gin.Context) {
	storeI, ok := h.platformStoreOrUnavailable(c)
	if !ok {
		return
	}
	projectID := c.Param("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "project_id is required"})
		return
	}
	// Sprint A 의 parseTestResultsWindow helper 활용 (?window=Nd short string
	// + RFC3339 from/to). default 30d.
	windowFrom, windowTo, err := parseTestResultsWindow(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": err.Error()})
		return
	}
	limit := 20
	if v := c.Query("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 50 {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "limit must be 1..50"})
			return
		}
		limit = n
	}
	data, total, err := storeI.ListProjectTestResults(c.Request.Context(), projectID, store.BuildRunListOptions{
		Limit:      limit,
		WindowFrom: windowFrom,
		WindowTo:   windowTo,
	})
	if err != nil {
		writeServerError(c, err, "project.tests.weighted")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   data,
		"meta":   gin.H{"total": total, "limit": limit},
	})
}
