package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// RM-ONBOARD-01 — API-84 GET /api/v1/organizations/search (§16.4).
// typeahead 검색 — q (>=2 chars) 의 case-insensitive substring match on
// org_units.label. limit (default 20, max 20). 응답 = unit_id + name 만
// (REQ-FR-ONBOARD-004). 권한 가드 없음.
//
// 인증: OIDC (token-only actor 도 호출 가능). gating: onboardingGate allowlist.

type orgSearchResponseRow struct {
	UnitID string `json:"unit_id"`
	Name   string `json:"name"`
}

func (h Handler) searchOrganizations(c *gin.Context) {
	if h.cfg.OrganizationStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "organization store is not configured",
		})
		return
	}

	q := strings.TrimSpace(c.Query("q"))
	if len(q) < 2 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": "rejected",
			"code":   "invalid_query_params",
			"error":  "q must be at least 2 characters",
		})
		return
	}

	limit := 20
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 || parsed > 20 {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"status": "rejected",
				"code":   "invalid_query_params",
				"error":  "limit must be a positive integer <= 20",
			})
			return
		}
		limit = parsed
	}

	units, err := h.cfg.OrganizationStore.SearchOrgUnits(c.Request.Context(), q, limit)
	if err != nil {
		writeServerError(c, err, "organizations.search")
		return
	}

	rows := make([]orgSearchResponseRow, 0, len(units))
	for _, u := range units {
		rows = append(rows, orgSearchResponseRow{
			UnitID: u.UnitID,
			Name:   u.Label,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   rows,
		"meta": gin.H{
			"limit":         limit,
			"total_matched": len(rows),
		},
	})
}
