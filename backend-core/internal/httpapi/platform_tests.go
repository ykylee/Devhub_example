package httpapi

import (
	"net/http"
	"strconv"

	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

// platformTestResults handler — Sprint C (kpi-tests-per-domain-scope.md §2.3
// follow-up).
//
// sub-project 가중치 적용 test results 정공법: platform 의 N개 sub-project 의
// build_runs status 를 종합 + equal average pass rate + multi-project recent.
// Sprint B-Tests 의 projectTestResults 와 정공법 정합 (sub-project 단위만 다름).

// platformTestResults — GET /api/v1/platforms/:platform_id/test-results
//
// Response: domain.PlatformWeightedTestResults 의 JSON 직렬화 + meta (total, limit).
func (h *Handler) platformTestResults(c *gin.Context) {
	storeI, ok := h.platformStoreOrUnavailable(c)
	if !ok {
		return
	}
	platformID := c.Param("platform_id")
	if platformID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "platform_id is required"})
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
	data, total, err := storeI.ListPlatformTestResults(c.Request.Context(), platformID, store.BuildRunListOptions{
		Limit:      limit,
		WindowFrom: windowFrom,
		WindowTo:   windowTo,
	})
	if err != nil {
		writeServerError(c, err, "platform.tests.weighted")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   data,
		"meta":   gin.H{"total": total, "limit": limit},
	})
}
