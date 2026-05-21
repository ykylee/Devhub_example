package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

// RM-ONBOARD-01 — API-86 POST /api/v1/admin/users/:user_id/review (§16.7).
// system_admin 의 명시 transition pending_review → reviewed. 사용자의
// onboarding 이 이미 완료된 상태에서만 동작 (onboarding_completed_at IS NOT NULL
// AND review_status='pending_review').
//
// 인증: OIDC + RBAC `users:edit` (system_admin). gating: onboardingGate
// allowlist 외 — admin 자신은 항상 완료된 사용자.

func (h Handler) confirmUserReview(c *gin.Context) {
	if h.cfg.OrganizationStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "organization store is not configured",
		})
		return
	}

	userID := strings.TrimSpace(c.Param("user_id"))
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "rejected",
			"error":  "user_id is required",
		})
		return
	}

	user, err := h.cfg.OrganizationStore.ConfirmUserReview(c.Request.Context(), userID)
	if err == nil {
		auditLog := h.recordAuditBestEffort(c, "account.review_confirmed", "user", user.UserID, map[string]any{
			"user_id":         user.UserID,
			"primary_unit_id": user.PrimaryUnitID,
			"reviewed_by":     requestActor(c).Login,
		})
		response := gin.H{
			"status": "ok",
			"data": gin.H{
				"user_id":         user.UserID,
				"review_status":   user.ReviewStatus,
				"reviewed_at":     user.UpdatedAt,
				"reviewed_by":     requestActor(c).Login,
				"primary_unit_id": user.PrimaryUnitID,
			},
		}
		addAuditMeta(response, auditLog)
		c.JSON(http.StatusOK, response)
		return
	}

	// store.ConfirmUserReview 가 ErrNotFound 반환 (affected=0) — 정확한 분기를
	// 위해 GetUser 로 재확인.
	if !errors.Is(err, store.ErrNotFound) {
		writeServerError(c, err, "onboarding.confirm_review")
		return
	}
	current, getErr := h.cfg.OrganizationStore.GetUser(c.Request.Context(), userID)
	if getErr != nil {
		if errors.Is(getErr, store.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"status": "not_found",
				"code":   "user_not_found",
				"error":  "user not found",
			})
			return
		}
		writeServerError(c, getErr, "onboarding.confirm_review.lookup")
		return
	}
	switch {
	case current.OnboardingCompletedAt == nil:
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": "rejected",
			"code":   "onboarding_not_completed",
			"error":  "user has not submitted onboarding yet",
		})
	case current.ReviewStatus == "reviewed":
		c.JSON(http.StatusConflict, gin.H{
			"status": "conflict",
			"code":   "review_already_confirmed",
			"error":  "user is already reviewed",
		})
	default:
		writeServerError(c, err, "onboarding.confirm_review.unknown")
	}
}
