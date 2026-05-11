package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// HydraLogoutAdmin covers the slice of *HydraAdminClient that the logout proxy
// invokes. Tests inject a fake.
type HydraLogoutAdmin interface {
	GetLogoutRequest(ctx context.Context, challenge string) (HydraLogoutRequest, error)
	AcceptLogoutRequest(ctx context.Context, challenge string) (string, error)
}

// logoutRequestPayload is the body the frontend posts. Per DEC-1=B, the
// frontend handles Kratos /self-service/logout/browser itself; backend only
// owns Hydra accept + Hydra revoke.
//
// At least one of logout_challenge / refresh_token must be present:
//   - logout_challenge alone: RP-initiated OIDC logout from /auth/logout
//   - refresh_token alone: Header Sign Out — invalidate the refresh token so a
//     stolen copy cannot reissue access tokens
//   - both: full sign-out from a privileged path
//
// client_id is required whenever refresh_token is present (Hydra public revoke
// requires client_id for public clients). When omitted and logout_challenge is
// present, we fall back to the client_id Hydra reports for the challenge.
type logoutRequestPayload struct {
	LogoutChallenge string `json:"logout_challenge,omitempty"`
	RefreshToken    string `json:"refresh_token,omitempty"`
	ClientID        string `json:"client_id,omitempty"`
}

func (h Handler) authLogout(c *gin.Context) {
	if h.cfg.HydraLogout == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "auth proxy is not configured (HydraLogout)",
		})
		return
	}

	var req logoutRequestPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "invalid json body"})
		return
	}
	req.LogoutChallenge = strings.TrimSpace(req.LogoutChallenge)
	req.RefreshToken = strings.TrimSpace(req.RefreshToken)
	req.ClientID = strings.TrimSpace(req.ClientID)

	if req.LogoutChallenge == "" && req.RefreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "rejected",
			"error":  "either logout_challenge or refresh_token is required",
		})
		return
	}

	ctx := c.Request.Context()
	subject := ""
	clientID := req.ClientID

	// Resolve subject and client_id from the logout_challenge when available,
	// so audit records carry user context and the revoke step has a client_id
	// even when the frontend did not send one.
	if req.LogoutChallenge != "" {
		logoutReq, err := h.cfg.HydraLogout.GetLogoutRequest(ctx, req.LogoutChallenge)
		if errors.Is(err, ErrHydraChallengeNotFound) {
			c.JSON(http.StatusGone, gin.H{
				"status": "gone",
				"error":  "logout_challenge expired or unknown",
				"code":   "logout_challenge_unknown",
			})
			return
		}
		if err != nil {
			writeServerError(c, err, "auth.logout.get_request")
			return
		}
		subject = logoutReq.Subject
		if clientID == "" {
			clientID = logoutReq.Client.ClientID
		}
	}

	// Best-effort revoke. We perform it before accept so a downstream accept
	// failure does not strand a still-valid refresh token. If revoke itself
	// fails we record an audit warning and keep going — the accept still has
	// to run for the OIDC flow to terminate.
	revokeStatus := "skipped"
	if req.RefreshToken != "" {
		if h.cfg.HydraRevoker == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unavailable",
				"error":  "auth proxy is not configured (HydraRevoker)",
			})
			return
		}
		if clientID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "rejected",
				"error":  "client_id is required when refresh_token is provided",
			})
			return
		}
		if err := h.cfg.HydraRevoker.RevokeRefreshToken(ctx, req.RefreshToken, clientID); err != nil {
			revokeStatus = "failed"
			h.recordAuditBestEffort(c, "auth.logout.revoke_failed", "user", subject, map[string]any{
				"client_id": clientID,
				"reason":    err.Error(),
			})
		} else {
			revokeStatus = "succeeded"
		}
	}

	var redirectTo string
	if req.LogoutChallenge != "" {
		got, err := h.cfg.HydraLogout.AcceptLogoutRequest(ctx, req.LogoutChallenge)
		if errors.Is(err, ErrHydraChallengeNotFound) {
			c.JSON(http.StatusGone, gin.H{
				"status": "gone",
				"error":  "logout_challenge expired or unknown",
				"code":   "logout_challenge_unknown",
			})
			return
		}
		if err != nil {
			writeServerError(c, err, "auth.logout.accept")
			return
		}
		redirectTo = got
	}

	h.recordAuditBestEffort(c, "auth.logout.succeeded", "user", subject, map[string]any{
		"client_id":     clientID,
		"revoke_status": revokeStatus,
		"has_challenge": req.LogoutChallenge != "",
	})

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data": gin.H{
			"redirect_to":   redirectTo,
			"revoke_status": revokeStatus,
		},
	})
}
