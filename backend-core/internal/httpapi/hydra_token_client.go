package httpapi

import (
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

var ErrHydraTokenInvalidGrant = errors.New("hydra token invalid_grant")

// HydraTokenExchanger exchanges an OAuth2 authorization code for tokens.
type HydraTokenExchanger interface {
	ExchangeAuthorizationCode(ctx context.Context, req HydraTokenExchangeRequest) (HydraTokenExchangeResponse, error)
}

// HydraTokenRevoker revokes a refresh token via Hydra public /oauth2/revoke.
// RFC 7009: the endpoint returns 200 even when the token is unknown, so callers
// treat any non-2xx as transport-level failure rather than "wrong token".
type HydraTokenRevoker interface {
	RevokeRefreshToken(ctx context.Context, refreshToken, clientID string) error
}

type HydraTokenExchangeRequest struct {
	Code         string
	CodeVerifier string
	RedirectURI  string
	ClientID     string
}

type HydraTokenExchangeResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// HydraTokenClient calls Hydra's public /oauth2/token endpoint.
type HydraTokenClient struct {
	PublicURL string
}

func (c *HydraTokenClient) ExchangeAuthorizationCode(ctx context.Context, req HydraTokenExchangeRequest) (HydraTokenExchangeResponse, error) {
	if c.PublicURL == "" {
		return HydraTokenExchangeResponse{}, errors.New("HydraTokenClient.PublicURL is not configured")
	}

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", req.Code)
	form.Set("code_verifier", req.CodeVerifier)
	form.Set("redirect_uri", req.RedirectURI)
	form.Set("client_id", req.ClientID)

	endpoint := strings.TrimRight(c.PublicURL, "/") + "/oauth2/token"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return HydraTokenExchangeResponse{}, fmt.Errorf("build hydra token request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client().Do(httpReq)
	if err != nil {
		return HydraTokenExchangeResponse{}, fmt.Errorf("call hydra token endpoint: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return HydraTokenExchangeResponse{}, fmt.Errorf("read hydra token response: %w", err)
	}

	if resp.StatusCode == http.StatusBadRequest && strings.Contains(strings.ToLower(string(body)), "invalid_grant") {
		return HydraTokenExchangeResponse{}, ErrHydraTokenInvalidGrant
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return HydraTokenExchangeResponse{}, fmt.Errorf("hydra token status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var out HydraTokenExchangeResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return HydraTokenExchangeResponse{}, fmt.Errorf("decode hydra token response: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return HydraTokenExchangeResponse{}, errors.New("hydra token response missing access_token")
	}
	return out, nil
}

// RevokeRefreshToken posts to Hydra public /oauth2/revoke. Public PKCE clients
// do not send a client_secret; client_id alone identifies the requester.
func (c *HydraTokenClient) RevokeRefreshToken(ctx context.Context, refreshToken, clientID string) error {
	if c.PublicURL == "" {
		return errors.New("HydraTokenClient.PublicURL is not configured")
	}
	if strings.TrimSpace(refreshToken) == "" {
		return errors.New("refresh_token is required")
	}
	if strings.TrimSpace(clientID) == "" {
		return errors.New("client_id is required")
	}

	form := url.Values{}
	form.Set("token", refreshToken)
	form.Set("token_type_hint", "refresh_token")
	form.Set("client_id", clientID)

	endpoint := strings.TrimRight(c.PublicURL, "/") + "/oauth2/revoke"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("build hydra revoke request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client().Do(httpReq)
	if err != nil {
		return fmt.Errorf("call hydra revoke endpoint: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("hydra revoke status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}

func (c *HydraTokenClient) client() *http.Client {
	return &http.Client{Timeout: 10 * time.Second}
}
