package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// requireOnboardingFlag — feature flag guard (RM-ONBOARD-01 / PR #278 self-review P1 #1).
// `DEVHUB_ONBOARDING_GATE_ENABLED=false` (default) 상태에서 신규 onboarding endpoint
// (POST /me/onboarding / PATCH /me / GET /organizations/search /
// POST /admin/users/:id/review) 호출 시 404 — main 동작 변경 없음 보장.
//
// flag true 시 정상 진입 → handler body 실행.
//
// 본 guard 는 onboardingGate middleware (gate=true 시 미완료 사용자 차단) 와 별개.
// onboardingGate 는 gate 자체의 enforcement layer 이고, 본 guard 는 신규 endpoint
// 자체의 conditional registration 효과를 handler-level 에서 emulate.
func (h Handler) requireOnboardingFlag(c *gin.Context) bool {
	if h.cfg.OnboardingGateEnabled {
		return true
	}
	c.JSON(http.StatusNotFound, gin.H{
		"status": "not_found",
		"code":   "onboarding_feature_disabled",
		"error":  "onboarding endpoints are disabled (DEVHUB_ONBOARDING_GATE_ENABLED=false)",
	})
	return false
}
