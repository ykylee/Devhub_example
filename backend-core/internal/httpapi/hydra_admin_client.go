package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ErrHydraChallengeNotFound is returned when Hydra reports the supplied
// login_challenge as unknown — typically a stale browser tab or a tampered
// query parameter. The handler should redirect the user back to /login.
var ErrHydraChallengeNotFound = errors.New("hydra login_challenge not found")

// HydraAdminClient invokes Hydra's admin OAuth2 request endpoints used by the
// /api/v1/auth/login proxy. Token introspection lives on a sibling type
// (HydraIntrospectionVerifier) so the verifier can stay swappable for tests.
type HydraAdminClient struct {
	// AdminURL is Hydra's admin base URL (default http://127.0.0.1:4445).
	AdminURL string
	// HTTPClient is reused for all admin calls; defaults to a 5s-timeout client.
	HTTPClient *http.Client
}

// HydraLoginRequest is the slice of GET /admin/oauth2/auth/requests/login
// that the proxy needs to validate the incoming challenge before asking
// Kratos for credentials.
type HydraLoginRequest struct {
	Challenge      string
	Skip           bool
	Subject        string
	RequestedScope []string
	Client         struct {
		ClientID  string
		ClientName string
	}
	RequestURL string
}

// AcceptLoginRequest tells Hydra that the supplied subject is authenticated
// for this challenge. Hydra responds with redirect_to which the frontend
// follows to complete the OIDC code flow (consent + callback).
//
// remember=true asks Hydra to set a long-lived login session cookie so the
// next /oauth2/auth from the same browser short-circuits this whole flow.
// rememberFor is the cookie lifetime in seconds.
func (c *HydraAdminClient) AcceptLoginRequest(ctx context.Context, challenge, subject string, remember bool, rememberFor int) (string, error) {
	if strings.TrimSpace(c.AdminURL) == "" {
		return "", errors.New("HydraAdminClient.AdminURL is not configured")
	}
	if strings.TrimSpace(challenge) == "" {
		return "", errors.New("login_challenge is required")
	}
	if strings.TrimSpace(subject) == "" {
		return "", errors.New("subject is required")
	}

	endpoint := strings.TrimRight(c.AdminURL, "/") +
		"/admin/oauth2/auth/requests/login/accept?login_challenge=" + url.QueryEscape(challenge)
	payload := map[string]any{
		"subject":      subject,
		"remember":     remember,
		"remember_for": rememberFor,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("encode hydra accept payload: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build hydra accept request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.client().Do(req)
	if err != nil {
		return "", fmt.Errorf("call hydra accept: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read hydra accept response: %w", err)
	}

	switch {
	case resp.StatusCode == http.StatusOK:
		var out struct {
			RedirectTo string `json:"redirect_to"`
		}
		if err := json.Unmarshal(respBody, &out); err != nil {
			return "", fmt.Errorf("decode hydra accept response: %w", err)
		}
		if strings.TrimSpace(out.RedirectTo) == "" {
			return "", errors.New("hydra accept returned empty redirect_to")
		}
		return out.RedirectTo, nil
	case resp.StatusCode == http.StatusNotFound:
		return "", ErrHydraChallengeNotFound
	default:
		return "", fmt.Errorf("hydra accept status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
}

// GetLoginRequest fetches the metadata Hydra associates with a given
// login_challenge. The handler uses Skip=true as a fast path: when Hydra
// already remembers the subject (skip=true), we skip Kratos and accept the
// existing subject directly.
func (c *HydraAdminClient) GetLoginRequest(ctx context.Context, challenge string) (HydraLoginRequest, error) {
	if strings.TrimSpace(c.AdminURL) == "" {
		return HydraLoginRequest{}, errors.New("HydraAdminClient.AdminURL is not configured")
	}
	if strings.TrimSpace(challenge) == "" {
		return HydraLoginRequest{}, errors.New("login_challenge is required")
	}
	endpoint := strings.TrimRight(c.AdminURL, "/") +
		"/admin/oauth2/auth/requests/login?login_challenge=" + url.QueryEscape(challenge)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return HydraLoginRequest{}, fmt.Errorf("build hydra get login request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.client().Do(req)
	if err != nil {
		return HydraLoginRequest{}, fmt.Errorf("call hydra get login: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return HydraLoginRequest{}, fmt.Errorf("read hydra get login: %w", err)
	}
	if resp.StatusCode == http.StatusNotFound {
		return HydraLoginRequest{}, ErrHydraChallengeNotFound
	}
	if resp.StatusCode/100 != 2 {
		return HydraLoginRequest{}, fmt.Errorf("hydra get login status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var raw struct {
		Challenge      string   `json:"challenge"`
		Skip           bool     `json:"skip"`
		Subject        string   `json:"subject"`
		RequestedScope []string `json:"requested_scope"`
		Client         struct {
			ClientID   string `json:"client_id"`
			ClientName string `json:"client_name"`
		} `json:"client"`
		RequestURL string `json:"request_url"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return HydraLoginRequest{}, fmt.Errorf("decode hydra get login: %w", err)
	}
	out := HydraLoginRequest{
		Challenge:      raw.Challenge,
		Skip:           raw.Skip,
		Subject:        raw.Subject,
		RequestedScope: raw.RequestedScope,
		RequestURL:     raw.RequestURL,
	}
	out.Client.ClientID = raw.Client.ClientID
	out.Client.ClientName = raw.Client.ClientName
	return out, nil
}

func (c *HydraAdminClient) client() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return &http.Client{Timeout: 5 * time.Second}
}
