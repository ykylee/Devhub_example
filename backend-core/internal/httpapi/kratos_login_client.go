package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ErrInvalidCredentials is returned when Kratos rejects the password.
// The handler maps this to a 401 with a generic message so callers cannot
// distinguish "user does not exist" from "wrong password" (timing/oracle
// hardening per OWASP ASVS V2.1.10).
var ErrInvalidCredentials = errors.New("invalid credentials")

// ErrLoginFlowExpired signals that a previously created login flow lifetime
// (kratos.yaml selfservice.flows.login.lifespan, default 10m) has elapsed.
// The frontend should restart the OIDC flow.
var ErrLoginFlowExpired = errors.New("kratos login flow expired")

// PasswordAuthHTTPClient drives the password self-service login flow on the Ory Kratos
// public API. It only owns the steps needed by /api/v1/auth/login; settings
// (password change), recovery, and registration flows live in their own
// helpers introduced by later PRs.
type PasswordAuthHTTPClient struct {
	// PublicURL is the base URL of the Kratos public API
	// (e.g. http://127.0.0.1:4433). Required.
	PublicURL string
	// HTTPClient is used for all calls; defaults to a 5s-timeout client.
	HTTPClient *http.Client
}

// IDPIdentity is the slice of an identity that the login proxy needs to
// authenticate the caller and pass a stable subject to Hydra.
type IDPIdentity struct {
	// ID is the Kratos UUID (identity.id). Stable for the lifetime of the
	// identity but not the value DevHub stores on rbac/users — that is
	// metadata_public.user_id below.
	ID string
	// UserID is the DevHub users.user_id pulled from
	// metadata_public.user_id (ADR-0001 §5). Empty when the operator has
	// not populated the metadata yet — login still proceeds, but the
	// caller must reject anonymous DevHub users.
	UserID string
	// Email, DisplayName and SystemID come from traits per identity.schema.json.
	SystemID    string
	Email       string
	DisplayName string
	// SessionToken carries the Kratos api-mode session token returned by
	// /self-service/login/api submission (L4-B, work_26_05_11-e). The
	// authLogin handler caches it in IDPSessionCache so the later
	// /api/v1/account/password proxy can drive Kratos settings flow on
	// behalf of the user without holding a browser cookie.
	SessionToken string
}

// IDPLoginFlow is the minimum subset of the Kratos login flow response
// that /api/v1/auth/login round-trips: the flow id (used as the form action
// id) plus the CSRF token nested in the password method node.
type IDPLoginFlow struct {
	ID        string
	CSRFToken string
	ExpiresAt time.Time
}

// CreateLoginFlow opens an api-mode login flow. We use api-mode (refresh=true,
// return_to optional) instead of browser-mode because the DevHub login proxy
// drives the flow server-to-server; browser-mode would require us to mirror
// Kratos cookies back to the user agent.
func (c *PasswordAuthHTTPClient) CreateLoginFlow(ctx context.Context) (IDPLoginFlow, error) {
	if strings.TrimSpace(c.PublicURL) == "" {
		return IDPLoginFlow{}, errors.New("PasswordAuthHTTPClient.PublicURL is not configured")
	}
	endpoint := strings.TrimRight(c.PublicURL, "/") + "/self-service/login/api"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return IDPLoginFlow{}, fmt.Errorf("build kratos login flow request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.client().Do(req)
	if err != nil {
		return IDPLoginFlow{}, fmt.Errorf("call kratos login flow: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return IDPLoginFlow{}, fmt.Errorf("read kratos login flow: %w", err)
	}
	if resp.StatusCode/100 != 2 {
		return IDPLoginFlow{}, fmt.Errorf("kratos login flow status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return parseLoginFlow(body)
}

// SubmitLogin posts password credentials against an existing flow id. The
// caller is responsible for creating the flow with CreateLoginFlow first.
func (c *PasswordAuthHTTPClient) SubmitLogin(ctx context.Context, flow IDPLoginFlow, identifier, password string) (IDPIdentity, error) {
	if strings.TrimSpace(c.PublicURL) == "" {
		return IDPIdentity{}, errors.New("PasswordAuthHTTPClient.PublicURL is not configured")
	}
	endpoint := strings.TrimRight(c.PublicURL, "/") + "/self-service/login?flow=" + flow.ID
	payload := map[string]any{
		"method":     "password",
		"identifier": identifier,
		"password":   password,
	}
	if flow.CSRFToken != "" {
		payload["csrf_token"] = flow.CSRFToken
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return IDPIdentity{}, fmt.Errorf("encode kratos login payload: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(encoded))
	if err != nil {
		return IDPIdentity{}, fmt.Errorf("build kratos login submit: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.client().Do(req)
	if err != nil {
		return IDPIdentity{}, fmt.Errorf("call kratos login submit: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return IDPIdentity{}, fmt.Errorf("read kratos login submit: %w", err)
	}

	switch {
	case resp.StatusCode == http.StatusOK:
		return parseLoginSuccess(body)
	case resp.StatusCode == http.StatusBadRequest:
		logRequestCtx(ctx, "[PasswordAuthHTTPClient] SubmitLogin 400: %s", string(body))
		// 400 is Kratos's response for both schema errors and rejected
		// credentials; treat the latter as the dominant case to keep
		// callers from leaking which arm fired.
		return IDPIdentity{}, ErrInvalidCredentials
	case resp.StatusCode == http.StatusGone || resp.StatusCode == http.StatusUnauthorized:
		return IDPIdentity{}, ErrLoginFlowExpired
	default:
		logRequestCtx(ctx, "[PasswordAuthHTTPClient] SubmitLogin status %d: %s", resp.StatusCode, string(body))
		return IDPIdentity{}, fmt.Errorf("kratos login submit status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
}

func (c *PasswordAuthHTTPClient) client() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return &http.Client{Timeout: 5 * time.Second}
}

// kratosLoginFlowResponse mirrors the Kratos public login flow JSON
// shape that we depend on. Fields outside this struct are ignored by
// encoding/json, so a Kratos minor-version field rename only breaks if it
// touches one of these paths.
type kratosLoginFlowResponse struct {
	ID        string    `json:"id"`
	ExpiresAt time.Time `json:"expires_at"`
	UI        struct {
		Nodes []struct {
			Attributes struct {
				Name  string `json:"name"`
				Value any    `json:"value"`
			} `json:"attributes"`
		} `json:"nodes"`
	} `json:"ui"`
}

func parseLoginFlow(body []byte) (IDPLoginFlow, error) {
	var raw kratosLoginFlowResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return IDPLoginFlow{}, fmt.Errorf("decode kratos login flow: %w", err)
	}
	if raw.ID == "" {
		return IDPLoginFlow{}, errors.New("kratos login flow missing id")
	}
	flow := IDPLoginFlow{ID: raw.ID, ExpiresAt: raw.ExpiresAt}
	for _, node := range raw.UI.Nodes {
		if node.Attributes.Name == "csrf_token" {
			if s, ok := node.Attributes.Value.(string); ok {
				flow.CSRFToken = s
			}
			break
		}
	}
	return flow, nil
}

// kratosLoginSuccessResponse covers the api-mode 200 response (session +
// session_token + identity). browser mode returns a redirect_browser_to
// URL instead — we do not use that path.
type kratosLoginSuccessResponse struct {
	SessionToken string `json:"session_token"`
	Session      struct {
		Identity kratosIdentityRaw `json:"identity"`
	} `json:"session"`
}

type kratosIdentityRaw struct {
	ID             string `json:"id"`
	Traits         struct {
		SystemID    string `json:"system_id"`
		Email       string `json:"email"`
		DisplayName string `json:"display_name"`
	} `json:"traits"`
	MetadataPublic struct {
		UserID string `json:"user_id"`
	} `json:"metadata_public"`
}

func parseLoginSuccess(body []byte) (IDPIdentity, error) {
	var raw kratosLoginSuccessResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return IDPIdentity{}, fmt.Errorf("decode kratos login success: %w", err)
	}
	identity := raw.Session.Identity
	if identity.ID == "" {
		return IDPIdentity{}, errors.New("kratos login success missing identity.id")
	}
	return IDPIdentity{
		ID:           identity.ID,
		UserID:       strings.TrimSpace(identity.MetadataPublic.UserID),
		SystemID:     identity.Traits.SystemID,
		Email:        identity.Traits.Email,
		DisplayName:  identity.Traits.DisplayName,
		SessionToken: raw.SessionToken,
	}, nil
}
