package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// KratosLoginClient covers the slice of *KratosClient that the login proxy
// invokes. Tests inject a fake. The two settings flow methods (L4-C,
// work_26_05_11-e) sit here too so the /api/v1/account/password handler can
// depend on a single dependency.
type KratosLoginClient interface {
	CreateLoginFlow(ctx context.Context) (KratosLoginFlow, error)
	SubmitLogin(ctx context.Context, flow KratosLoginFlow, identifier, password string) (KratosIdentity, error)
	CreateSettingsFlow(ctx context.Context, sessionToken string) (string, error)
	SubmitSettingsPassword(ctx context.Context, sessionToken, flowID, newPassword string) error
}

// HydraLoginAdmin covers the slice of *HydraAdminClient that the login proxy
// invokes. Tests inject a fake.
type HydraLoginAdmin interface {
	GetLoginRequest(ctx context.Context, challenge string) (HydraLoginRequest, error)
	AcceptLoginRequest(ctx context.Context, challenge, subject string, remember bool, rememberFor int) (string, error)
	GetConsentRequest(ctx context.Context, challenge string) (HydraConsentRequest, error)
	AcceptConsentRequest(ctx context.Context, challenge string, grantedScope []string, remember bool, rememberFor int) (string, error)
}

// loginRequestPayload is the body the frontend posts from /auth/login. The
// login_challenge comes from Hydra's redirect; identifier+password from the
// password form.
type loginRequestPayload struct {
	LoginChallenge string `json:"login_challenge"`
	Identifier     string `json:"identifier"`
	Password       string `json:"password"`
	Remember       bool   `json:"remember,omitempty"`
}

const defaultRememberForSeconds = 3600

// authLogin drives the password leg of the OIDC code flow. The frontend has
// already been redirected here from Hydra with a login_challenge; this
// handler exchanges credentials with Kratos and tells Hydra the outcome.
//
// docs/backend_api_contract.md section 11 will describe this endpoint once
// PR-LOGIN-2 wires the frontend; until then the spec lives here in code.
func (h Handler) authLogin(c *gin.Context) {
	if h.cfg.KratosLogin == nil || h.cfg.HydraAdmin == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "auth proxy is not configured (KratosLogin / HydraAdmin)",
		})
		return
	}

	var req loginRequestPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "invalid json body"})
		return
	}
	req.LoginChallenge = strings.TrimSpace(req.LoginChallenge)
	req.Identifier = strings.TrimSpace(req.Identifier)
	if req.LoginChallenge == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "login_challenge is required"})
		return
	}

	ctx := c.Request.Context()

	// Fast path: when Hydra reports skip=true the browser already has a
	// valid login session for this client; we accept the cached subject
	// without round-tripping Kratos.
	hydraReq, err := h.cfg.HydraAdmin.GetLoginRequest(ctx, req.LoginChallenge)
	if errors.Is(err, ErrHydraChallengeNotFound) {
		c.JSON(http.StatusGone, gin.H{"status": "gone", "error": "login_challenge expired or unknown", "code": "login_challenge_unknown"})
		return
	}
	if err != nil {
		writeServerError(c, err, "auth.login.get_request")
		return
	}
	if hydraReq.Skip {
		redirectTo, err := h.cfg.HydraAdmin.AcceptLoginRequest(ctx, req.LoginChallenge, hydraReq.Subject, req.Remember, defaultRememberForSeconds)
		if err != nil {
			writeServerError(c, err, "auth.login.accept_skip")
			return
		}
		h.recordAuditBestEffort(c, "auth.login.succeeded", "user", hydraReq.Subject, map[string]any{
			"flow":      "skip",
			"client_id": hydraReq.Client.ClientID,
		})
		c.JSON(http.StatusOK, gin.H{"status": "ok", "data": gin.H{"redirect_to": redirectTo}})
		return
	}
	if req.Identifier == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "identifier and password are required"})
		return
	}

	logRequest(c, "[authLogin] Creating login flow for challenge %s", req.LoginChallenge)
	flow, err := h.cfg.KratosLogin.CreateLoginFlow(ctx)
	if err != nil {
		writeServerError(c, err, "auth.login.create_flow")
		return
	}

	logRequest(c, "[authLogin] Submitting login for identifier %s", req.Identifier)
	identity, err := h.cfg.KratosLogin.SubmitLogin(ctx, flow, req.Identifier, req.Password)
	switch {
	case errors.Is(err, ErrKratosInvalidCredentials):
		h.recordAuditBestEffort(c, "auth.login.failed", "login_id", req.Identifier, map[string]any{
			"reason":    "invalid_credentials",
			"client_id": hydraReq.Client.ClientID,
		})
		c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized", "error": "invalid credentials"})
		return
	case errors.Is(err, ErrKratosFlowExpired):
		c.JSON(http.StatusGone, gin.H{"status": "gone", "error": "login flow expired; restart sign-in", "code": "login_flow_expired"})
		return
	case err != nil:
		writeServerError(c, err, "auth.login.submit")
		return
	}

	subject := identity.UserID
	if subject == "" {
		// Fall back to the Kratos identity.id when the operator has not
		// populated metadata_public.user_id yet. Hydra introspect will
		// surface a UUID instead of users.user_id, which RBAC does not
		// recognise — log it so operators can fix the mapping.
		subject = identity.ID
		h.recordAuditBestEffort(c, "auth.login.subject_fallback", "user", subject, map[string]any{
			"reason":    "missing_metadata_public_user_id",
			"client_id": hydraReq.Client.ClientID,
		})
	}

	// Cache the Kratos session_token under the resolved subject so the
	// /api/v1/account/password proxy can drive the settings flow without
	// holding a browser cookie (L4-B, work_26_05_11-e). Put is a no-op
	// when SessionToken is empty (api-mode response with no token, or
	// fakes that did not set one).
	h.cfg.KratosSessionCache.Put(subject, identity.SessionToken)

	logRequest(c, "[authLogin] Accepting login request for subject %s", subject)
	redirectTo, err := h.cfg.HydraAdmin.AcceptLoginRequest(ctx, req.LoginChallenge, subject, req.Remember, defaultRememberForSeconds)
	if err != nil {
		writeServerError(c, err, "auth.login.accept")
		return
	}

	h.recordAuditBestEffort(c, "auth.login.succeeded", "user", subject, map[string]any{
		"flow":      "password",
		"client_id": hydraReq.Client.ClientID,
		"identity":  identity.ID,
	})

	c.JSON(http.StatusOK, gin.H{"status": "ok", "data": gin.H{"redirect_to": redirectTo}})
}
