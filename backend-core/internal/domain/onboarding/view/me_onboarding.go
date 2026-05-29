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

// RM-ONBOARD-01 — API-83 POST /api/v1/me/onboarding (ARCH-ONBOARD-01..06,
// §16.3). 사용자 onboarding 제출 — display_name + primary_unit_id 설정 + row
// INSERT (DB 미등록) 또는 UPDATE (admin pre-seeded 미완료) + onboarding_completed_at
// + review_status='pending_review' + audit emit.
//
// 인증: OIDC (token-only actor 도 호출 가능). gating: onboardingGate allowlist
// 포함 (REQ-FR-ONBOARD-009).
//
// role 필드는 받지 않음 (REQ-FR-ONBOARD-002 / ADR-0021 §3.1). Keycloak claim
// 매핑 또는 fallback `developer` 만 사용.

type onboardingSubmitRequest struct {
	DisplayName   string `json:"display_name"`
	PrimaryUnitID string `json:"primary_unit_id"`
}

func (h *OnboardingHandler) SubmitOnboarding(c *gin.Context) {
	start := time.Now()
	defer func() {
		ObserveOnboardingSubmitDuration(time.Since(start).Seconds())
	}()

	if !h.RequireOnboardingFlag(c) {
		// RequireOnboardingFlag 가 404 (onboarding_feature_disabled) 응답.
		// codex review (#313 P2-1) — "unavailable" 은 OrganizationStore nil 시점용,
		// flag disabled 는 명시 분리. `feature_disabled` label.
		ObserveOnboardingSubmit("feature_disabled")
		return
	}
	if h.cfg.OrganizationStore == nil {
		ObserveOnboardingSubmit("unavailable")
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "organization store is not configured",
		})
		return
	}

	actor := httphelp.RequestActor(c)
	login := strings.TrimSpace(actor.Login)
	if login == "" || login == "system" {
		ObserveOnboardingSubmit("unauthenticated")
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"status": "unauthenticated",
			"error":  "no authenticated user in request context",
		})
		return
	}

	var req onboardingSubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ObserveOnboardingSubmit("rejected")
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": "rejected",
			"code":   "invalid_payload",
			"error":  err.Error(),
		})
		return
	}
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	req.PrimaryUnitID = strings.TrimSpace(req.PrimaryUnitID)
	if req.DisplayName == "" || len(req.DisplayName) > 100 {
		ObserveOnboardingSubmit("rejected")
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": "rejected",
			"code":   "invalid_payload",
			"error":  "display_name is required (1~100 chars)",
		})
		return
	}
	if req.PrimaryUnitID == "" {
		ObserveOnboardingSubmit("rejected")
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": "rejected",
			"code":   "invalid_payload",
			"error":  "primary_unit_id is required",
		})
		return
	}

	// token claim 정보 (authenticateActor 가 context 에 set).
	email := ""
	if v, ok := c.Get("devhub_actor_email"); ok {
		if s, ok := v.(string); ok {
			email = s
		}
	}
	subject := ""
	if v, ok := c.Get("devhub_actor_subject"); ok {
		if s, ok := v.(string); ok {
			subject = s
		}
	}
	roleStr := ""
	if v, ok := c.Get("devhub_actor_role"); ok {
		if s, ok := v.(string); ok {
			roleStr = s
		}
	}
	fallbackRole := domain.AppRole(roleStr)
	if !onboardingValidRole(fallbackRole) {
		fallbackRole = domain.AppRoleDeveloper
	}

	input := domain.OnboardingSubmitInput{
		UserID:        login,
		Email:         email,
		DisplayName:   req.DisplayName,
		PrimaryUnitID: req.PrimaryUnitID,
		IdPSubject:    subject,
		FallbackRole:  fallbackRole,
	}

	user, err := h.cfg.OrganizationStore.SubmitOnboarding(c.Request.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrConflict):
			ObserveOnboardingSubmit("conflict")
			c.JSON(http.StatusConflict, gin.H{
				"status": "conflict",
				"code":   "onboarding_already_completed",
				"error":  err.Error(),
			})
		case errors.Is(err, store.ErrNotFound):
			ObserveOnboardingSubmit("not_found")
			c.JSON(http.StatusNotFound, gin.H{
				"status": "not_found",
				"code":   "unit_not_found",
				"error":  err.Error(),
			})
		default:
			ObserveOnboardingSubmit("server_error")
			httphelp.WriteServerError(c, err, "onboarding.submit")
		}
		return
	}

	auditLog := h.recordAuditBestEffort(c, "account.onboarding_completed", "user", user.UserID, map[string]any{
		"user_id":         user.UserID,
		"primary_unit_id": user.PrimaryUnitID,
		"display_name":    user.DisplayName,
	})

	response := gin.H{
		"status": "ok",
		"data":   appUserFromDomain(user),
	}
	addAuditMeta(response, auditLog)
	ObserveOnboardingSubmit("ok")
	c.JSON(http.StatusCreated, response)
}
