package view

import (
	"github.com/devhub/backend-core/internal/shared/httphelp"
	"encoding/json"
	"net/http"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/gin-gonic/gin"
)

// Application 롤업 endpoint (API-57, sprint claude/work_260514-c).
// concept §13.4 weight_policy normalize 룰 + meta(period/filters/weight_policy/
// applied_weights/fallbacks/data_gaps). REQ-FR-APP-012, REQ-NFR-PROJ-006.

func (h *ApplicationHandler) ApplicationRollup(c *gin.Context) {
	storeI, ok := h.ApplicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	id := c.Param("application_id")
	opts := domain.ApplicationRollupOptions{
		Policy: domain.WeightPolicy(c.Query("weight_policy")),
	}
	if opts.Policy == "" {
		opts.Policy = domain.WeightPolicyEqual
	}
	switch opts.Policy {
	case domain.WeightPolicyEqual, domain.WeightPolicyRepoRole, domain.WeightPolicyCustom:
	default:
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "weight_policy must be one of equal/repo_role/custom"})
		return
	}
	if raw := c.Query("custom_weights"); raw != "" {
		var weights map[string]float64
		if err := json.Unmarshal([]byte(raw), &weights); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "custom_weights must be JSON object {\"repo_full_name\": weight}"})
			return
		}
		opts.CustomWeights = weights
	}
	if v := c.Query("from"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "from must be RFC3339"})
			return
		}
		opts.WindowFrom = t
	}
	if v := c.Query("to"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "to must be RFC3339"})
			return
		}
		opts.WindowTo = t
	}

	rollup, err := storeI.ComputeApplicationRollup(c.Request.Context(), id, opts)
	if err != nil {
		// weight_policy validation 실패는 store 가 fmt.Errorf 로 전달 — 422 분기
		errMsg := err.Error()
		if errMsg != "" && (containsAny(errMsg, "invalid weight policy")) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"status": "rejected",
				"error":  errMsg,
				"code":   "invalid_weight_policy",
			})
			return
		}
		httphelp.WriteServerError(c, err, "application.rollup")
		return
	}

	// "마지막 빌드 상태" derive 값 — REQ-FR-APPDASH-001 (단순 % 보다 broken/red 즉시 표기).
	// 빈 문자열은 "unknown" 으로 정규화하여 frontend 가 항상 valid 한 enum 값을 받도록.
	targetBranchBuildStatus := rollup.TargetBranchBuildStatus
	if targetBranchBuildStatus == "" {
		targetBranchBuildStatus = "unknown"
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data": gin.H{
			"pull_request_distribution":  rollup.PullRequestDistribution,
			"build_success_rate":         rollup.BuildSuccessRate,
			"build_avg_duration_seconds": rollup.BuildAvgDurationSeconds,
			"quality_score":              rollup.QualityScore,
			"quality_gate_failed_count":  rollup.QualityGateFailedCount,
			"critical_warning_count":     rollup.CriticalWarningCount,
			"target_branch_build_status": targetBranchBuildStatus,
		},
		"meta": gin.H{
			"period": gin.H{
				"from": rollup.Meta.Period.From.UTC().Format(time.RFC3339),
				"to":   rollup.Meta.Period.To.UTC().Format(time.RFC3339),
			},
			"filters":         rollup.Meta.Filters,
			"weight_policy":   string(rollup.Meta.WeightPolicy),
			"applied_weights": rollup.Meta.AppliedWeights,
			"fallbacks":       rollup.Meta.Fallbacks,
			"data_gaps":       rollup.Meta.DataGaps,
		},
	})
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
	}
	return false
}
