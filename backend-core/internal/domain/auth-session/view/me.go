package view

import (
	"github.com/devhub/backend-core/internal/shared/httphelp"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

// meResponse — API-32 GET /api/v1/me (§16.2). RM-ONBOARD-01 의 §16.2 확장으로
// onboarding_required + onboarding_completed_at + review_status 3 신규 필드.
// token-only actor (DB row 미존재) 의 경우 display_name/email/role 은 token
// claim 에서 추출 + primary_unit_id 등은 빈 값 + onboarding_required=true.
// MeResponse — meResponse cross-package test 접근용 export alias.
type MeResponse = meResponse

type meResponse struct {
	Login                 string     `json:"login"`
	UserID                string     `json:"user_id,omitempty"`
	Subject               string     `json:"subject,omitempty"`
	Email                 string     `json:"email,omitempty"`
	DisplayName           string     `json:"display_name,omitempty"`
	Role                  string     `json:"role,omitempty"`
	PrimaryUnitID         string     `json:"primary_unit_id,omitempty"`
	CurrentUnitID         string     `json:"current_unit_id,omitempty"`
	Source                string     `json:"actor_source"`
	OnboardingRequired    bool       `json:"onboarding_required"`
	OnboardingCompletedAt *time.Time `json:"onboarding_completed_at"`
	// ReviewStatus — API §16.2 spec ("null | 'pending_review' | 'reviewed'").
	// `*string` 으로 미설정 시 명시적 `null` 응답 — frontend 가 null 분기로 안전 처리.
	ReviewStatus *string `json:"review_status"`
}

// getMe returns the authenticated actor for the current request. Frontend
// uses this to derive the active role after a successful Keycloak OIDC login.
// Returns 401 when the request did not produce an authenticated actor (no
// Authorization header in production, or the dev fallback resolved actor to
// "system"). The legacy X-Devhub-Actor header is intentionally ignored —
// ADR-0004 finalized its removal; SEC-4 already stripped the production
// handling.
//
// RM-ONBOARD-01 (§16.2) — 응답 shape 에 onboarding_required + onboarding_completed_at
// + review_status 3 필드 신규. token-only actor (DB row 미존재) + DB row 의
// completed_at NULL = onboarding_required: true.
func (h *AuthHandler) GetMe(c *gin.Context) {
	actor := httphelp.RequestActor(c)
	if actor.Login == "" || actor.Login == "system" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"status": "unauthenticated",
			"error":  "no authenticated user in request context",
		})
		return
	}

	subject := ""
	if v, ok := c.Get("devhub_actor_subject"); ok {
		if s, ok := v.(string); ok {
			subject = s
		}
	}
	role := ""
	if v, ok := c.Get("devhub_actor_role"); ok {
		if s, ok := v.(string); ok {
			role = s
		}
	}

	resp := meResponse{
		Login:   actor.Login,
		Subject: subject,
		Role:    role,
		Source:  actor.Source,
	}

	// RM-ONBOARD-01 — onboarding_required 계산은 gate flag 와 분리한다.
	// gate flag 는 "차단(onboardingGate middleware)" on/off 만 제어하고,
	// /api/v1/me 응답의 onboarding state 는 항상 hydrate 되어야 frontend
	// (auth callback/AuthGuard)가 온보딩 진입을 정확히 결정할 수 있다.
	if h.cfg.OrganizationStore != nil && actor.Login != "" && actor.Login != "system" {
		user, err := h.cfg.OrganizationStore.GetUser(c.Request.Context(), actor.Login)
		switch {
		case err == nil:
			resp.UserID = user.UserID
			resp.Email = user.Email
			resp.DisplayName = user.DisplayName
			resp.PrimaryUnitID = user.PrimaryUnitID
			resp.CurrentUnitID = user.CurrentUnitID
			resp.OnboardingCompletedAt = user.OnboardingCompletedAt
			if user.ReviewStatus != "" {
				rs := user.ReviewStatus
				resp.ReviewStatus = &rs
			}
			resp.OnboardingRequired = user.OnboardingCompletedAt == nil
		case errors.Is(err, store.ErrNotFound):
			// token-only actor (DB row 미존재) — Email/DisplayName 은 token
			// claim 에서 추출 (authenticateActor 가 set).
			if v, ok := c.Get("devhub_actor_email"); ok {
				if s, ok := v.(string); ok {
					resp.Email = s
				}
			}
			if v, ok := c.Get("devhub_actor_display_name"); ok {
				if s, ok := v.(string); ok {
					resp.DisplayName = s
				}
			}
			resp.OnboardingRequired = true
		default:
			// Schema drift 등 — sentinel 만으로 response 의 missing field 가
			// onboarding_required=false 를 의미하면 spec 위반. 안전 default:
			// required=true 로 fail-safe.
			resp.OnboardingRequired = true
			httphelp.LogRequest(c, "[getMe] %q GetUser failed: %v; defaulting onboarding_required=true", actor.Login, err)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   resp,
	})
}

// patchMeRequest — API-85 PATCH /api/v1/me (§16.5). self-service display_name
// + primary_unit_id 변경. role 필드는 받지 않음.
type patchMeRequest struct {
	DisplayName   *string `json:"display_name,omitempty"`
	PrimaryUnitID *string `json:"primary_unit_id,omitempty"`
}

// patchMe — API-85 (§16.5). 본인 profile 변경 (display_name / primary_unit_id).
// primary_unit_id 변경 시 review_status 가 자동으로 pending_review 로 reset
// (REQ-FR-ONBOARD-007, UC-ONBOARD-07).
//
// 인증: OIDC + 본인 row 만. gating: onboardingGate allowlist 외 — 완료
// 사용자만 호출 (미완료는 POST /me/onboarding 으로 첫 제출).
func (h *AuthHandler) PatchMe(c *gin.Context) {
	if !h.RequireOnboardingFlag(c) {
		return
	}
	if h.cfg.OrganizationStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "organization store is not configured",
		})
		return
	}
	actor := httphelp.RequestActor(c)
	login := strings.TrimSpace(actor.Login)
	if login == "" || login == "system" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"status": "unauthenticated",
			"error":  "no authenticated user in request context",
		})
		return
	}

	var req patchMeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": "rejected",
			"code":   "invalid_payload",
			"error":  err.Error(),
		})
		return
	}
	if req.DisplayName == nil && req.PrimaryUnitID == nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": "rejected",
			"code":   "invalid_payload",
			"error":  "at least one of display_name or primary_unit_id is required",
		})
		return
	}

	input := domain.UpdateUserInput{}
	if req.DisplayName != nil {
		trimmed := strings.TrimSpace(*req.DisplayName)
		if trimmed == "" || len(trimmed) > 100 {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"status": "rejected",
				"code":   "invalid_payload",
				"error":  "display_name must be 1~100 chars",
			})
			return
		}
		input.DisplayName = &trimmed
	}
	unitChanged := false
	if req.PrimaryUnitID != nil {
		trimmed := strings.TrimSpace(*req.PrimaryUnitID)
		if trimmed == "" {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"status": "rejected",
				"code":   "invalid_payload",
				"error":  "primary_unit_id must not be empty",
			})
			return
		}
		input.PrimaryUnitID = &trimmed
		input.CurrentUnitID = &trimmed
		// REQ-FR-ONBOARD-007: unit 변경 시 review_status=pending_review 재진입.
		pendingReview := domain.ReviewStatusPendingReview
		input.ReviewStatus = &pendingReview
		unitChanged = true
	}

	user, err := h.cfg.OrganizationStore.UpdateUser(c.Request.Context(), login, input)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"status": "not_found",
				"code":   "unit_not_found",
				"error":  err.Error(),
			})
		default:
			httphelp.WriteServerError(c, err, "me.patch")
		}
		return
	}

	if unitChanged {
		h.recordAuditBestEffort(c, "account.unit_changed", "user", user.UserID, map[string]any{
			"user_id":         user.UserID,
			"primary_unit_id": user.PrimaryUnitID,
			"by_user":         login,
			"review_status":   user.ReviewStatus,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   appUserFromDomain(user),
	})
}
