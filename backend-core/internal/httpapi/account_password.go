package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// accountPasswordRequest is the body of POST /api/v1/account/password.
// current_password is verified server-side by running a fresh Kratos api-mode
// login (defense in depth: Kratos's privileged_session_max_age window alone
// would let a 15-minute-old session change passwords without re-typing).
type accountPasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// updateMyPassword is the self-service password-change proxy (L4-D,
// work_26_05_11-e). The route is bypassed in the RBAC matrix (12.8.1
// self-info pattern); authentication is enforced by requiring a non-system
// actor, and authorization is implicit — the handler only mutates the
// caller's own identity.
//
// Flow:
//  1. Resolve actor (DevHub user_id) and look up the email.
//  2. Verify current_password by running a Kratos api-mode login. The
//     submission yields a fresh session_token within Kratos's
//     privileged_session_max_age window — exactly the window the settings
//     flow requires for password mutations.
//  3. Open a settings flow keyed on that session_token and submit
//     method=password + new_password.
//  4. Cache the refreshed session_token so subsequent self-service calls
//     in the same browser session reuse it.
//  5. Emit audit account.password_self_change.
func (h Handler) updateMyPassword(c *gin.Context) {
	actor := requestActor(c)
	if actor.Login == "" || actor.Login == "system" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "unauthenticated",
			"error":  "no authenticated user; sign in first",
			"code":   "reauth_required",
		})
		return
	}

	if h.cfg.KratosLogin == nil || h.cfg.OrganizationStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "account password proxy requires KratosLogin + OrganizationStore",
		})
		return
	}

	var req accountPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "invalid json body"})
		return
	}
	if req.CurrentPassword == "" || req.NewPassword == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "rejected",
			"error":  "current_password and new_password are required",
		})
		return
	}
	if req.CurrentPassword == req.NewPassword {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "rejected",
			"error":  "new_password must differ from current_password",
			"code":   "validation",
		})
		return
	}

	ctx := c.Request.Context()
	user, err := h.cfg.OrganizationStore.GetUser(ctx, actor.Login)
	if err != nil {
		h.recordAuditBestEffort(c, "account.password_self_change.no_user", "user", actor.Login, map[string]any{
			"reason": err.Error(),
		})
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "unauthenticated",
			"error":  "actor not found in DevHub users",
			"code":   "reauth_required",
		})
		return
	}
	email := strings.TrimSpace(user.Email)
	if email == "" {
		writeServerError(c, errors.New("user.email empty"), "account.password.user_email")
		return
	}

	// 1) Verify current_password via a fresh api-mode login.
	flow, err := h.cfg.KratosLogin.CreateLoginFlow(ctx)
	if err != nil {
		writeServerError(c, err, "account.password.create_login")
		return
	}
	identity, err := h.cfg.KratosLogin.SubmitLogin(ctx, flow, email, req.CurrentPassword)
	switch {
	case errors.Is(err, ErrKratosInvalidCredentials):
		h.recordAuditBestEffort(c, "account.password_self_change.invalid_current", "user", actor.Login, nil)
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "unauthenticated",
			"error":  "current password is incorrect",
			"code":   "current_password_invalid",
		})
		return
	case errors.Is(err, ErrKratosFlowExpired):
		writeServerError(c, err, "account.password.login_flow_expired")
		return
	case err != nil:
		writeServerError(c, err, "account.password.submit_login")
		return
	}
	if strings.TrimSpace(identity.SessionToken) == "" {
		writeServerError(c, errors.New("kratos returned empty session_token"), "account.password.empty_session_token")
		return
	}

	// 2) Settings flow → submit new password.
	flowID, err := h.cfg.KratosLogin.CreateSettingsFlow(ctx, identity.SessionToken)
	if err != nil {
		if se := IsKratosSettingsError(err); se != nil {
			respondSettingsError(c, se, "account.password.settings_create")
			return
		}
		writeServerError(c, err, "account.password.settings_create")
		return
	}
	if err := h.cfg.KratosLogin.SubmitSettingsPassword(ctx, identity.SessionToken, flowID, req.NewPassword); err != nil {
		if se := IsKratosSettingsError(err); se != nil {
			respondSettingsError(c, se, "account.password.settings_submit")
			return
		}
		writeServerError(c, err, "account.password.settings_submit")
		return
	}

	// 3) Refresh cache so subsequent self-service calls reuse the new
	// privileged session.
	h.cfg.KratosSessionCache.Put(actor.Login, identity.SessionToken)

	h.recordAuditBestEffort(c, "account.password_self_change", "user", actor.Login, map[string]any{
		"identity_id": identity.ID,
	})

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   gin.H{"user_id": actor.Login},
	})
}

// respondSettingsError maps KratosSettingsError onto the HTTP shape the
// frontend account.service.ts.SettingsFlowError consumes.
func respondSettingsError(c *gin.Context, se *KratosSettingsError, op string) {
	switch se.Code {
	case KratosSettingsValidation:
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "rejected",
			"error":  se.Message,
			"code":   "validation",
		})
	case KratosSettingsPrivilegedRequired, KratosSettingsSessionInvalid:
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "unauthenticated",
			"error":  "re-authentication required",
			"code":   "reauth_required",
		})
	case KratosSettingsFlowExpired:
		c.JSON(http.StatusGone, gin.H{
			"status": "gone",
			"error":  "settings flow expired; retry",
			"code":   "flow_expired",
		})
	default:
		writeServerError(c, se, op)
	}
}
