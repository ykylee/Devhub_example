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

	authview "github.com/devhub/backend-core/internal/domain/auth-session/view"
)

// ErrIdentityNotFound is returned by IdentityAdmin implementations when no
// identity matches the supplied DevHub user_id (or the underlying GET returns
// 404). Defined here because KeycloakAdminClient is the source-of-truth impl
// (ADR-0019).
var ErrIdentityNotFound = errors.New("identity not found")

// KeycloakAdminClient maps account-admin operations to Keycloak Admin API +
// OIDC user-facing endpoints. It satisfies IdentityAdmin and OIDCLogoutClient.
//
// 두 client 구분 — RFC 6749 §4.1.3 / Keycloak 의 token binding 정책:
//   - ClientID/ClientSecret: **admin client** (e.g. devhub-backend). Admin
//     REST endpoint (/admin/realms/{realm}/users/{id}/logout) 호출용.
//   - OIDCClientID/OIDCClientSecret: **OIDC client** (e.g. devhub-frontend).
//     User-facing OIDC endpoint (/realms/{realm}/protocol/openid-connect/logout)
//     호출용. Keycloak 은 token 발급 client 와 다른 client 의 logout 을 거부함
//     (production 401 회피).
type KeycloakAdminClient struct {
	AdminURL        string
	Realm           string
	ClientID        string
	ClientSecret    string
	IssuerURL       string
	OIDCClientID    string
	OIDCClientSecret string
	HTTPClient      *http.Client
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

// LogoutUserSession terminates all active sessions for the given Keycloak
// identity via Keycloak Admin REST: POST /admin/realms/{realm}/users/{id}/logout.
// This is a best-effort revocation — the frontend also triggers the OIDC
// end_session redirect independently. If the identity doesn't exist (HTTP 404),
// the method returns nil (already logged out).
func (c *KeycloakAdminClient) LogoutUserSession(ctx context.Context, identityID string) error {
	_, _, err := c.adminJSON(ctx, http.MethodPost, "users/"+url.PathEscape(identityID)+"/logout", nil)
	if errors.Is(err, ErrIdentityNotFound) {
		return nil
	}
	return err
}

// OIDCLogout — sprint mvs/work_260608-i-488-sign-out (N-8 / P1-6). Keycloak
// OIDC /protocol/openid-connect/logout endpoint 로 refresh_token 폐기.
// issue #488 spec: client_id + client_secret + refresh_token 을 form-urlencoded
// body 로 POST. Keycloak 4xx (invalid/expired token) → nil (idempotent).
// network / 5xx error → error (502 매핑용).
//
// ⚠ OIDCClientID / OIDCClientSecret 사용 (admin ClientID 와 분리). RFC 6749
// §4.1.3 / Keycloak token binding 정책 — token 발급 client 와 logout client
// 가 동일해야 Keycloak 이 logout 을 수락. production 에서 frontend 가
// devhub-frontend client 로 token 발급 → logout 도 devhub-frontend client
// 자격증명 필요. devhub-backend admin client 자격증명 사용 시 Keycloak 이
// 401 반환 (codex P1 review #2 정합).
//
// N-8 hotfix 4차 codex P1 follow-up: error 카테고리 구분. sentinel errors:
//   - authview.ErrOIDCConfigMissing : backend config 결함 (missing realm/oidc_client_id/secret) —
//                                    OIDC 호출 *전* 발견. handler 가 marker 미부착
//                                    + 정상 OIDC 분기 결정.
//   - authview.ErrOIDCNetworkUnreachable : 네트워크 error (c.client().Do) + 5xx —
//                                         Keycloak 도달 실패. handler 가 marker
//                                         부착 + frontend OIDC skip 결정.
func (c *KeycloakAdminClient) OIDCLogout(ctx context.Context, refreshToken string) error {
	if strings.TrimSpace(refreshToken) == "" {
		return errors.New("OIDC logout requires non-empty refresh_token")
	}
	if strings.TrimSpace(c.Realm) == "" || strings.TrimSpace(c.OIDCClientID) == "" || strings.TrimSpace(c.OIDCClientSecret) == "" {
		// sentinel: ErrOIDCConfigMissing — handler 가 marker 미부착 결정.
		return fmt.Errorf("%w: KeycloakAdminClient requires realm, oidc_client_id, oidc_client_secret for OIDC logout (separate from admin client)", authview.ErrOIDCConfigMissing)
	}
	logoutURL := c.oidcLogoutEndpoint()
	form := url.Values{}
	form.Set("client_id", c.OIDCClientID)
	form.Set("client_secret", c.OIDCClientSecret)
	form.Set("refresh_token", refreshToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, logoutURL, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("build keycloak logout request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.client().Do(req)
	if err != nil {
		// sentinel: ErrOIDCNetworkUnreachable — handler 가 marker 부착 결정.
		return fmt.Errorf("%w: call keycloak logout endpoint: %v", authview.ErrOIDCNetworkUnreachable, err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	// 200/204 = success. 400/401 = invalid token — caller 에서 idempotent 204
	// 반환 (이미 revoke 된 토큰 재시도 등). 5xx = server error, propagate.
	if resp.StatusCode/100 == 2 {
		return nil
	}
	if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnauthorized {
		return nil
	}
	// 5xx: sentinel: ErrOIDCNetworkUnreachable — Keycloak 서버 측 outage.
	return fmt.Errorf("%w: keycloak logout status %d", authview.ErrOIDCNetworkUnreachable, resp.StatusCode)
}

// oidcLogoutEndpoint — issuer url 기반 derive. tokenEndpoint 와 같은 패턴.
func (c *KeycloakAdminClient) oidcLogoutEndpoint() string {
	if issuer := strings.TrimRight(strings.TrimSpace(c.IssuerURL), "/"); issuer != "" {
		return issuer + "/protocol/openid-connect/logout"
	}
	u, err := url.Parse(strings.TrimRight(c.AdminURL, "/"))
	if err != nil {
		return strings.TrimRight(c.AdminURL, "/") + "/realms/" + url.PathEscape(c.Realm) + "/protocol/openid-connect/logout"
	}
	u.Path = path.Join(u.Path, "realms", c.Realm, "protocol", "openid-connect", "logout")
	return u.String()
}

// ADR-0020 sub-carve E (sprint -n) — Keycloak admin = 별도 운영팀 (PoLP).
// write methods (CreateIdentity / UpdateIdentityPassword / SetIdentityState /
// DeleteIdentity) 는 정공법 제거. service account 는 view-users + view-events
// 만 요구. password reset / state change / delete 는 Keycloak admin console 가
// 책임.

// KeycloakUserDetails — sprint -k (#212 P1-1, ADR-0020 sub-carve C) — admin event
// listener 의 USER:UPDATE / USER:DELETE 처리 시 Keycloak 의 최신 user state 를
// fetch 해 DevHub `users` 컬럼 sync 에 사용한다. Keycloak Admin REST:
// GET /admin/realms/{realm}/users/{id} 응답의 평탄화.
type KeycloakUserDetails struct {
	ID         string              `json:"id"`
	Username   string              `json:"username"`
	Email      string              `json:"email,omitempty"`
	FirstName  string              `json:"firstName,omitempty"`
	LastName   string              `json:"lastName,omitempty"`
	Enabled    bool                `json:"enabled"`
	Attributes map[string][]string `json:"attributes,omitempty"`
}

// GetUserDetails fetches the current user record from Keycloak. Returns
// ErrIdentityNotFound when the user has already been deleted (HTTP 404).
// admin event handler 가 본 메서드로 최신 email/name/enabled 를 가져와
// DevHub users 컬럼 sync 에 사용 (ADR-0020 §5.3.2).
func (c *KeycloakAdminClient) GetUserDetails(ctx context.Context, identityID string) (KeycloakUserDetails, error) {
	body, headers, err := c.adminGET(ctx, "/users/"+url.PathEscape(identityID))
	if err != nil {
		if headers != nil && http.StatusText(http.StatusNotFound) != "" {
			// adminGET 가 HTTP error 를 wrapping 한 경우 404 식별을 위해 body
			// 가 nil 이면서 status 가 명시되어야 한다. 현재 adminJSON/adminGET
			// 의 error 형식이 status 명시이므로 ErrIdentityNotFound 변환은
			// caller 가 처리 (admin event 의 ResourcePath 가 유효하면 404 는
			// 이미 삭제된 race 상황이라 caller 가 USER:DELETE 흐름으로 분기).
		}
		return KeycloakUserDetails{}, err
	}
	var details KeycloakUserDetails
	if err := json.Unmarshal(body, &details); err != nil {
		return KeycloakUserDetails{}, fmt.Errorf("decode keycloak user details: %w", err)
	}
	return details, nil
}

// GetUserGroups fetches the group membership list for the given identity.
// admin event handler 가 GROUP_MEMBERSHIP:CREATE / DELETE 처리 시 사용한다 —
// Keycloak group composite role 매핑 (keycloak_groups_rbac_mapping.md 권장 B)
// 의 결과 role 을 DevHub `users.role` 에 sync (ADR-0020 §5.3.1).
type KeycloakGroup struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path,omitempty"`
}

func (c *KeycloakAdminClient) GetUserGroups(ctx context.Context, identityID string) ([]KeycloakGroup, error) {
	body, _, err := c.adminGET(ctx, "/users/"+url.PathEscape(identityID)+"/groups")
	if err != nil {
		return nil, err
	}
	var groups []KeycloakGroup
	if err := json.Unmarshal(body, &groups); err != nil {
		return nil, fmt.Errorf("decode keycloak user groups: %w", err)
	}
	return groups, nil
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
	// admin operations: always use AdminURL. NOT IssuerURL — codex P1
	// review (PR #496) 정합. admin endpoint 와 user-facing OIDC endpoint
	// 는 deployment 별로 다른 host 일 수 있음 (public ingress vs internal
	// docker). IssuerURL 은 oidcLogoutEndpoint() 전용. 본 함수는 admin
	// token endpoint 만.
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

