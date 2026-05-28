package gitea

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/devhub/backend-core/internal/domain"
)

// AuthorizationHeader resolves the full Authorization header value for an
// outbound call given the provider's resolved auth. The boolean reports whether
// credentials are configured (false → caller should skip; the provider has no
// usable outbound credentials for this mode). An error is returned only for
// operational failures (e.g. an oauth2 token exchange that fails).
//
//   - token            → "token <pat>"            (Gitea PAT)
//   - basic/app_password → "Basic base64(user:secret)"
//   - oauth2           → "Bearer <access_token>"  (client-credentials grant)
//   - agent            → unsupported for direct API sync (ok=false)
func AuthorizationHeader(ctx context.Context, auth domain.OutboundAuth, httpClient *http.Client) (string, bool, error) {
	switch auth.Mode {
	case domain.IntegrationAuthModeBasic, domain.IntegrationAuthModeAppPassword:
		if strings.TrimSpace(auth.Username) == "" || strings.TrimSpace(auth.Secret) == "" {
			return "", false, nil
		}
		creds := base64.StdEncoding.EncodeToString([]byte(auth.Username + ":" + auth.Secret))
		return "Basic " + creds, true, nil
	case domain.IntegrationAuthModeOAuth2:
		if strings.TrimSpace(auth.ClientID) == "" || strings.TrimSpace(auth.TokenURL) == "" || strings.TrimSpace(auth.Secret) == "" {
			return "", false, nil
		}
		accessToken, err := fetchOAuth2Token(ctx, httpClient, auth.TokenURL, auth.ClientID, auth.Secret)
		if err != nil {
			return "", false, err
		}
		return "Bearer " + accessToken, true, nil
	case domain.IntegrationAuthModeAgent:
		// Agent mode pulls via a separate agent process; this worker cannot
		// authenticate a direct API call on its behalf.
		return "", false, nil
	default: // token or unset
		if strings.TrimSpace(auth.Token) == "" {
			return "", false, nil
		}
		return "token " + auth.Token, true, nil
	}
}

type oauth2TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

// fetchOAuth2Token performs an OAuth2 client-credentials grant against tokenURL
// and returns the access token.
func fetchOAuth2Token(ctx context.Context, httpClient *http.Client, tokenURL, clientID, clientSecret string) (string, error) {
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("oauth2 token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("oauth2 token exchange: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("oauth2 token endpoint returned status %d", resp.StatusCode)
	}

	var tr oauth2TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return "", fmt.Errorf("decode oauth2 token response: %w", err)
	}
	if strings.TrimSpace(tr.AccessToken) == "" {
		return "", fmt.Errorf("oauth2 token response missing access_token")
	}
	return tr.AccessToken, nil
}

// NewClientForAuth builds a Client carrying the resolved Authorization header for
// the given outbound auth. Returns (nil, nil) when no credentials are configured
// (caller should skip), or an error on oauth2 exchange failure.
func NewClientForAuth(ctx context.Context, baseURL string, auth domain.OutboundAuth) (*Client, error) {
	c := NewClient(baseURL, "")
	header, ok, err := AuthorizationHeader(ctx, auth, c.HTTPClient)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	c.AuthHeader = header
	return c, nil
}
