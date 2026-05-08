// Package auth provides verifier implementations that satisfy httpapi.BearerTokenVerifier. The Hydra introspection verifier calls Ory Hydra's admin /admin/oauth2/introspect endpoint and maps the response to an authenticated actor; role extraction is configurable via RoleClaim (default "ext.role").
package auth

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

	"github.com/devhub/backend-core/internal/httpapi"
)

// ErrTokenInactive is returned when Hydra reports the introspected token as inactive (expired, revoked, or unknown).
var ErrTokenInactive = errors.New("bearer token is inactive")

// HydraIntrospectionVerifier verifies bearer tokens by calling Ory Hydra's admin introspection endpoint.
type HydraIntrospectionVerifier struct {
	// AdminURL is the base URL of Hydra's admin API, e.g. http://127.0.0.1:4445.
	AdminURL string
	// HTTPClient is used to call the introspection endpoint. Defaults to a 5s-timeout client.
	HTTPClient *http.Client
	// RoleClaim is a dotted path into the introspection response that holds the actor role (default "ext.role"). Supported prefixes: top-level keys (e.g. "scope", "username") or "ext.<key>" for ext map entries.
	RoleClaim string
}

type introspectResponse struct {
	Active   bool           `json:"active"`
	Subject  string         `json:"sub"`
	Username string         `json:"username"`
	Scope    string         `json:"scope"`
	ClientID string         `json:"client_id"`
	Audience []string       `json:"aud"`
	Exp      int64          `json:"exp"`
	Ext      map[string]any `json:"ext"`
}

func (v *HydraIntrospectionVerifier) VerifyBearerToken(ctx context.Context, token string) (httpapi.AuthenticatedActor, error) {
	if strings.TrimSpace(v.AdminURL) == "" {
		return httpapi.AuthenticatedActor{}, errors.New("HydraIntrospectionVerifier.AdminURL is not configured")
	}

	form := url.Values{"token": []string{token}}.Encode()
	endpoint := strings.TrimRight(v.AdminURL, "/") + "/admin/oauth2/introspect"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form))
	if err != nil {
		return httpapi.AuthenticatedActor{}, fmt.Errorf("build introspect request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := v.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return httpapi.AuthenticatedActor{}, fmt.Errorf("call hydra introspect: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return httpapi.AuthenticatedActor{}, fmt.Errorf("read introspect response: %w", err)
	}

	if resp.StatusCode/100 != 2 {
		return httpapi.AuthenticatedActor{}, fmt.Errorf("hydra introspect status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var ir introspectResponse
	if err := json.Unmarshal(body, &ir); err != nil {
		return httpapi.AuthenticatedActor{}, fmt.Errorf("decode introspect response: %w", err)
	}
	if !ir.Active {
		return httpapi.AuthenticatedActor{}, ErrTokenInactive
	}

	actor := httpapi.AuthenticatedActor{
		Subject: ir.Subject,
		Login:   ir.Username,
		Role:    extractRole(ir, v.RoleClaim),
	}
	if strings.TrimSpace(actor.Login) == "" {
		actor.Login = ir.Subject
	}
	return actor, nil
}

func extractRole(ir introspectResponse, claim string) string {
	claim = strings.TrimSpace(claim)
	if claim == "" {
		claim = "ext.role"
	}
	const extPrefix = "ext."
	if strings.HasPrefix(claim, extPrefix) {
		key := strings.TrimPrefix(claim, extPrefix)
		if v, ok := ir.Ext[key]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
		return ""
	}
	switch claim {
	case "scope":
		return ir.Scope
	case "username":
		return ir.Username
	case "client_id":
		return ir.ClientID
	}
	return ""
}
