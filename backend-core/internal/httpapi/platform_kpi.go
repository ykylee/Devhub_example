package httpapi

import (
	"net/http"

	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

// platformKPI handler — Sprint C (kpi-tests-per-domain-scope.md §2.3 + §6.3).
//
// sub-project rollup 정공법: platform 의 N개 sub-project 의 가중치 적용 metric
// 을 equal average 로 종합 (sub-project 균등). Sprint B 의 projectKPI 와
// 정공법 분리 (projectKPI 는 sub-repo × contribution_weight, platformKPI 는
// sub-project equal average).
//
// Sprint B 1차 PR #627 + Sprint B-Tests PR #628 의 follow-up. Sprint C 의 1차
// 진입. projectKPI 와 projectTestResults 와 동일하게 routePermissionTable 에
// `/platforms/:platform_id/kpi` row 등록 필수 (deny-by-default 회귀 가드 정합).

// platformKPI — GET /api/v1/platforms/:platform_id/kpi
//
// Response: domain.PlatformWeightedKPI 의 JSON 직렬화.
func (h *Handler) platformKPI(c *gin.Context) {
	storeI, ok := h.platformStoreOrUnavailable(c)
	if !ok {
		return
	}
	platformID := c.Param("platform_id")
	if platformID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "platform_id is required"})
		return
	}
	// ?window=Nd short string + RFC3339 from/to — Sprint A 의 parseTestResultsWindow
	// helper 활용 (정공법 정합). default 30d.
	windowFrom, windowTo, err := parseTestResultsWindow(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": err.Error()})
		return
	}
	opts := store.BuildRunListOptions{
		WindowFrom: windowFrom,
		WindowTo:   windowTo,
	}
	kpi, err := storeI.ComputePlatformWeightedKPI(c.Request.Context(), platformID, opts)
	if err != nil {
		writeServerError(c, err, "platform.kpi.weighted")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   kpi,
	})
}
