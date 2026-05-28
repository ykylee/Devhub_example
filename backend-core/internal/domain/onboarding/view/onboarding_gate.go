package view

import (
	"github.com/devhub/backend-core/internal/shared/httphelp"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RM-ONBOARD-01 (ADR-0021 §3.3, ARCH-ONBOARD-03, REQ-FR-ONBOARD-009).
// onboardingGate middleware 가 미완료 사용자 (token-only actor 또는 DB row 의
// onboarding_completed_at IS NULL) 의 allowlist 외 endpoint 호출 시 403 차단.
//
// 동작:
//   - feature flag (RouterConfig.OnboardingGateEnabled) false (default) → no-op.
//     Carve A 단독 머지 후 main 안정성 보장.
//   - context 의 `devhub_onboarding_required` flag = false (완료 사용자 또는
//     legacy mode) → c.Next() (정상).
//   - flag = true + path 가 allowlist (onboardingGateAllowlist) 멤버 → c.Next().
//   - flag = true + path 가 allowlist 외 → 403 + body
//     `{ "status": "forbidden", "code": "onboarding_required" }`.
//
// allowlist 는 backend endpoint 만 — frontend 정적/공통 페이지는 backend 호출
// 없이 렌더되므로 본 정책과 무관 (ARCH §9.3).

// onboardingGateAllowlist — 미완료 사용자가 호출 가능한 backend endpoint set.
// keys 는 gin route pattern (c.FullPath()) 와 정합. /health 는 unauth 가 라
// /api/v1 group 외부 — 본 allowlist 와 무관.
var onboardingGateAllowlist = map[string]bool{
	"/api/v1/me":                          true, // API-32 — onboarding_required flag 응답
	"/api/v1/me/onboarding":               true, // API-83 — 제출 endpoint (POST)
	"/api/v1/organizations/search":        true, // API-84 — typeahead picker
	"/api/v1/organization/hierarchy":      true, // 기존 — tree picker 소스
}

func (h *OnboardingHandler) OnboardingGate(c *gin.Context) {
	if !h.cfg.OnboardingGateEnabled {
		c.Next()
		return
	}

	required, _ := c.Get("devhub_onboarding_required")
	if required != true {
		c.Next()
		return
	}

	if onboardingGateAllowlist[c.FullPath()] {
		c.Next()
		return
	}

	httphelp.LogRequest(c, "[onboardingGate] blocking incomplete actor on %s", c.FullPath())
	ObserveOnboardingGateBlocked("onboarding_required")
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
		"status": "forbidden",
		"code":   "onboarding_required",
		"error":  "user must complete onboarding before accessing this endpoint",
	})
}
