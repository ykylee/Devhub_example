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
	"path"
	"strings"
	"time"
)

// KeycloakAdminClient maps account-admin operations to Keycloak Admin API.
// It intentionally satisfies IdentityAdmin so existing handlers can be reused
// during the migration window.
type KeycloakAdminClient struct {
	AdminURL     string
	Realm        string
	ClientID     string
	ClientSecret string
	IssuerURL    string
	HTTPClient   *http.Client
}

func (c *KeycloakAdminClient) CreateIdentity(ctx context.Context, email, name, userID, password string) (string, error) {
	payload := map[string]any{
		"username":    userID,
		"email":       email,
		"enabled":     true,
		"firstName":   name,
		"attributes":  map[string][]string{"devhub_user_id": {userID}},
		"credentials": []map[string]any{{"type": "password", "value": password, "temporary": true}},
	}
	respBody, headers, err := c.adminJSON(ctx, http.MethodPost, "/users", payload)
	if err != nil {
		return "", err
	}
	_ = respBody
	if id := keycloakIDFromLocation(headers.Get("Location")); id != "" {
		return id, nil
	}
	id, err := c.FindIdentityByUserID(ctx, userID)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (c *KeycloakAdminClient) FindIdentityByUserID(ctx context.Context, userID string) (string, error) {
	q := url.Values{}
	q.Set("username", userID)
	body, _, err := c.adminGET(ctx, "/users?"+q.Encode())
	if err != nil {
		return "", err
	}
	var users []struct {
		ID       string `json:"id"`
		Username string `json:"username"`
	}
	if err := json.Unmarshal(body, &users); err != nil {
		return "", fmt.Errorf("decode keycloak users lookup: %w", err)
	}
	for _, u := range users {
		if strings.EqualFold(strings.TrimSpace(u.Username), strings.TrimSpace(userID)) && strings.TrimSpace(u.ID) != "" {
			return strings.TrimSpace(u.ID), nil
		}
	}
	return "", ErrIdentityNotFound
}

func (c *KeycloakAdminClient) UpdateIdentityPassword(ctx context.Context, identityID, password string) error {
	payload := map[string]any{
		"type":      "password",
		"value":     password,
		"temporary": true,
	}
	_, _, err := c.adminJSON(ctx, http.MethodPut, "/users/"+url.PathEscape(identityID)+"/reset-password", payload)
	return err
}

func (c *KeycloakAdminClient) SetIdentityState(ctx context.Context, identityID string, active bool) error {
	payload := map[string]any{"enabled": active}
	_, _, err := c.adminJSON(ctx, http.MethodPut, "/users/"+url.PathEscape(identityID), payload)
	return err
}

func (c *KeycloakAdminClient) DeleteIdentity(ctx context.Context, identityID string) error {
	_, _, err := c.adminJSON(ctx, http.MethodDelete, "/users/"+url.PathEscape(identityID), nil)
	return err
}

// KeycloakUserEvent — sprint -u (PR-B) audit event listener 의 user events
// (LOGIN / LOGOUT / REGISTER / UPDATE_PASSWORD 등). design 문서 §3.1 + §4.1 매핑 표.
// Keycloak Admin REST: GET /admin/realms/{realm}/events?dateFrom=...
type KeycloakUserEvent struct {
	Time     int64             `json:"time"` // unix ms
	Type     string            `json:"type"` // LOGIN / LOGIN_ERROR / LOGOUT / REGISTER 등
	RealmID  string            `json:"realmId,omitempty"`
	ClientID string            `json:"clientId,omitempty"`
	UserID   string            `json:"userId,omitempty"`
	IPAddr   string            `json:"ipAddress,omitempty"`
	Details  map[string]string `json:"details,omitempty"`
	Error    string            `json:"error,omitempty"`
}

// KeycloakAdminEvent — admin events (USER:CREATE / USER:UPDATE / USER:DELETE /
// ROLE:CREATE 등). design 문서 §3.1 + §4.2 매핑 표. Keycloak Admin REST:
// GET /admin/realms/{realm}/admin-events?dateFrom=... (codex review #9 정정 정합).
type KeycloakAdminEvent struct {
	Time        int64  `json:"time"`
	RealmID     string `json:"realmId,omitempty"`
	AuthDetails *struct {
		RealmID  string `json:"realmId,omitempty"`
		ClientID string `json:"clientId,omitempty"`
		UserID   string `json:"userId,omitempty"`
		IPAddr   string `json:"ipAddress,omitempty"`
	} `json:"authDetails,omitempty"`
	OperationType string `json:"operationType"` // CREATE / UPDATE / DELETE / ACTION
	ResourceType  string `json:"resourceType"`  // USER / GROUP / ROLE / CLIENT / REALM 등
	ResourcePath  string `json:"resourcePath,omitempty"`
	Error         string `json:"error,omitempty"`
}

// ListUserEvents — Keycloak user events polling (sprint -u PR-B).
// dateFrom 은 ISO8601 (codex review #9 정정 정합 — `?dateFrom=` 이지 `?from=` 아님).
// max 는 page size (Keycloak default 100, 권장 500).
func (c *KeycloakAdminClient) ListUserEvents(ctx context.Context, dateFrom time.Time, max int) ([]KeycloakUserEvent, error) {
	q := url.Values{}
	if !dateFrom.IsZero() {
		q.Set("dateFrom", dateFrom.UTC().Format("2006-01-02T15:04:05Z"))
	}
	if max <= 0 {
		max = 500
	}
	q.Set("max", fmt.Sprintf("%d", max))
	body, _, err := c.adminGET(ctx, "/events?"+q.Encode())
	if err != nil {
		return nil, err
	}
	var events []KeycloakUserEvent
	if err := json.Unmarshal(body, &events); err != nil {
		return nil, fmt.Errorf("decode keycloak user events: %w", err)
	}
	return events, nil
}

// ListAdminEvents — Keycloak admin events polling (sprint -u PR-B).
// dateFrom 은 ISO8601. path = `/admin-events` (codex review #9 정정 정합 — `/events/admin` 아님).
func (c *KeycloakAdminClient) ListAdminEvents(ctx context.Context, dateFrom time.Time, max int) ([]KeycloakAdminEvent, error) {
	q := url.Values{}
	if !dateFrom.IsZero() {
		q.Set("dateFrom", dateFrom.UTC().Format("2006-01-02T15:04:05Z"))
	}
	if max <= 0 {
		max = 500
	}
	q.Set("max", fmt.Sprintf("%d", max))
	body, _, err := c.adminGET(ctx, "/admin-events?"+q.Encode())
	if err != nil {
		return nil, err
	}
	var events []KeycloakAdminEvent
	if err := json.Unmarshal(body, &events); err != nil {
		return nil, fmt.Errorf("decode keycloak admin events: %w", err)
	}
	return events, nil
}

func (c *KeycloakAdminClient) adminGET(ctx context.Context, adminPath string) ([]byte, http.Header, error) {
	return c.adminCall(ctx, http.MethodGet, adminPath, nil)
}

func (c *KeycloakAdminClient) adminJSON(ctx context.Context, method, adminPath string, payload any) ([]byte, http.Header, error) {
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return nil, nil, fmt.Errorf("encode keycloak request: %w", err)
		}
		body = bytes.NewReader(encoded)
	}
	return c.adminCall(ctx, method, adminPath, body)
}

func (c *KeycloakAdminClient) adminCall(ctx context.Context, method, adminPath string, body io.Reader) ([]byte, http.Header, error) {
	if strings.TrimSpace(c.AdminURL) == "" || strings.TrimSpace(c.Realm) == "" || strings.TrimSpace(c.ClientID) == "" || strings.TrimSpace(c.ClientSecret) == "" {
		return nil, nil, errors.New("KeycloakAdminClient requires admin_url, realm, client_id, client_secret")
	}
	token, err := c.accessToken(ctx)
	if err != nil {
		return nil, nil, err
	}

	base := strings.TrimRight(c.AdminURL, "/")
	endpoint := base + "/admin/realms/" + url.PathEscape(c.Realm) + "/" + strings.TrimLeft(adminPath, "/")
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, nil, fmt.Errorf("build keycloak request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client().Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("call keycloak admin: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("read keycloak admin response: %w", err)
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, resp.Header, ErrIdentityNotFound
	}
	if resp.StatusCode/100 != 2 {
		return nil, resp.Header, fmt.Errorf("keycloak admin status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	return respBody, resp.Header, nil
}

func (c *KeycloakAdminClient) accessToken(ctx context.Context) (string, error) {
	tokenURL := c.tokenEndpoint()
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", c.ClientID)
	form.Set("client_secret", c.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("build keycloak token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.client().Do(req)
	if err != nil {
		return "", fmt.Errorf("call keycloak token endpoint: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read keycloak token response: %w", err)
	}
	if resp.StatusCode/100 != 2 {
		return "", fmt.Errorf("keycloak token status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var raw struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return "", fmt.Errorf("decode keycloak token response: %w", err)
	}
	if strings.TrimSpace(raw.AccessToken) == "" {
		return "", errors.New("keycloak token response missing access_token")
	}
	return raw.AccessToken, nil
}

func (c *KeycloakAdminClient) tokenEndpoint() string {
	if issuer := strings.TrimRight(strings.TrimSpace(c.IssuerURL), "/"); issuer != "" {
		return issuer + "/protocol/openid-connect/token"
	}
	u, err := url.Parse(strings.TrimRight(c.AdminURL, "/"))
	if err != nil {
		return strings.TrimRight(c.AdminURL, "/") + "/realms/" + url.PathEscape(c.Realm) + "/protocol/openid-connect/token"
	}
	u.Path = path.Join(u.Path, "realms", c.Realm, "protocol", "openid-connect", "token")
	return u.String()
}

func (c *KeycloakAdminClient) client() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return &http.Client{Timeout: 5 * time.Second}
}

func keycloakIDFromLocation(location string) string {
	location = strings.TrimSpace(location)
	if location == "" {
		return ""
	}
	if parsed, err := url.Parse(location); err == nil {
		location = parsed.Path
	}
	parts := strings.Split(strings.TrimRight(location, "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	return strings.TrimSpace(parts[len(parts)-1])
}
