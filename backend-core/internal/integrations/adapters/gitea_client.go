package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// GiteaClient is a minimal Gitea API client (REST) used by X-4 (SCM create) and
// shared with X-5 (Gitea pull, PR #592). This is the X-4 reference minimal client;
// when X-5 PR #592 is merged, the richer client (5 method for pull cycle) replaces this.
type GiteaClient struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

// NewGiteaClient constructs a GiteaClient with sensible defaults.
func NewGiteaClient(baseURL, token string) *GiteaClient {
	return &GiteaClient{
		BaseURL:    strings.TrimRight(baseURL, "/"),
		Token:      token,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// doJSON is a small helper for GET/POST with bearer token + JSON decode.
func (c *GiteaClient) doJSON(ctx context.Context, method, url string, body io.Reader, out interface{}) error {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return fmt.Errorf("gitea: build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "token "+c.Token)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("gitea: http: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("gitea: status %d: %s", resp.StatusCode, strings.TrimSpace(string(errBody)))
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
