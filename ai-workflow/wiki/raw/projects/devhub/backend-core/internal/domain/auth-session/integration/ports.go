// Package integration — auth-session 도메인의 port interface 정의.
//
// v1.1 sprint -a 결정 (docs/planning/external-integrations-agentic-rag-roadmap.md §0.4).
// auth-session 도메인 layer 가 IdP (Keycloak) 와의 결합을 끊기 위한 Adapter
// pattern 적용. 본 package 는 **interface + canonical struct 만** 노출하며
// 실제 구현은 `sso-integrations/keycloak/` 에 위치.
//
// 역사:
//   - v1.0 (이전) — `view/auth.go` 에 `BearerTokenVerifier` interface + `view/handler.go` 에
//     `IdentityAdmin` / `OIDCLogoutClient` interface 정의. domain layer 가
//     IdP 와 직접 결합 (KeycloakJWKSVerifier / KeycloakAdminClient 가
//     domain/auth-session/service/ + httpapi/ 에 위치).
//   - v1.1 sprint -a (PR #538) — 본 package 로 interface 이동. view/ 의 interface 는
//     deprecated alias 로 유지 (backward compat). 신규 호출은 본 package
//     alias 사용.
//   - v1.1 sprint -a follow-up (PR #539) — `sso-integrations/keycloak/` stub + main wiring
//     + view/ deprecation. 본 PR 에서 port 의 KeycloakUserEvent/KeycloakAdminEvent 가
//     **type alias** 상태.
//   - v1.1 sprint -a follow-up (PR1) — real adapter 가 `sso-integrations/keycloak/`
//     로 이동. canonical struct 직접 정의 (alias 해제).
//
// Tier 분류: **공용** (사외 + 사내 양쪽 build 가 import 가능). interface 와
// canonical struct 만 노출하므로 사내 한정 정보 미포함. 사외 build 시
// saovae stub (build tag) 자동 사용.

package integration

import (
	"context"
	"time"

	"github.com/devhub/backend-core/internal/domain/auth-session/view"
	"github.com/devhub/backend-core/internal/shared/httphelp"
)

// --- Public type aliases (re-export from view for port callers) ---

// AuthenticatedActor — JWT claim 추출 결과. v1.0 의 view.AuthenticatedActor 의
// type alias. v1.1 sprint -a follow-up 에서 view.AuthenticatedActor 정의가
// 본 package 로 이전될 예정 (별도 PR).
type AuthenticatedActor = view.AuthenticatedActor

// --- Core port interfaces (migrated from view/) ---

// BearerTokenVerifier — OIDC Bearer JWT 검증. view/auth.go:65 의 동일 interface.
// auth middleware 가 hot path 에서 호출. 구현: sso-integrations/keycloak/ (JWKS 기반).
type BearerTokenVerifier = view.BearerTokenVerifier

// IdentityAdmin — Keycloak admin REST (user lifecycle + logout). view/handler.go:31
// 의 동일 interface. 사내 IdP lifecycle 관리용. 구현: sso-integrations/keycloak/.
type IdentityAdmin = view.IdentityAdmin

// OIDCLogoutClient — OIDC /protocol/openid-connect/logout endpoint wrapper.
// view/handler.go:205 의 동일 interface. user-facing OIDC endpoint (admin
// endpoint 와 분리). 구현: sso-integrations/keycloak/.
type OIDCLogoutClient = view.OIDCLogoutClient

// --- NEW port (defined in integration/, not view/) ---

// KeycloakEventPort — Keycloak user/admin event polling + webhook ingest.
// audit-ops/service/keycloak_event_puller.go 의 KeycloakEventLister 와
// 동등. **NEW port** (sprint -a 에서 신규 정의; 기존엔 interface 없이 직접
// struct 사용).
//
// 구현: sso-integrations/keycloak/ 의 KeycloakAdminClient.ListUserEvents /
// ListAdminEvents.
type KeycloakEventPort interface {
	// ListUserEvents — Keycloak user events (login, logout, register, ...) polling.
	ListUserEvents(ctx context.Context, dateFrom time.Time, max int) ([]KeycloakUserEvent, error)

	// ListAdminEvents — Keycloak admin events (CREATE_USER, UPDATE_ROLE, ...) polling.
	ListAdminEvents(ctx context.Context, dateFrom time.Time, max int) ([]KeycloakAdminEvent, error)
}

// KeycloakUserEvent — Keycloak user event 의 port canonical. v1.0 의
// httpapi.KeycloakUserEvent 와 field-compatible. v1.1 sprint -a follow-up PR1
// 에서 struct 직접 정의 (alias 해제). 호출 site (audit-ops puller, integration
// caller) 는 본 struct 만 import (sso-integrations 은 노출하지 않음).
//
// Keycloak wire format:
//   - Time: unix ms
//   - Type: LOGIN / LOGIN_ERROR / LOGOUT / REGISTER / UPDATE_PASSWORD 등
//   - Details: sessionId / redirect_uri 등 key=value metadata
//
// sso-integrations/keycloak/admin_client.go 의 ListUserEvents 가 본 struct 반환
// (wire format 매핑은 internal raw struct 으로 처리).
type KeycloakUserEvent struct {
	Time     int64
	Type     string
	RealmID  string
	ClientID string
	UserID   string
	IPAddr   string
	Details  map[string]string
	Error    string
}

// KeycloakAdminEvent — Keycloak admin event 의 port canonical. v1.0 의
// httpapi.KeycloakAdminEvent (nested AuthDetails) 와 다르게 audit-ops port
// caller 가 일관되게 flat access 할 수 있도록 **평탄화**. 사내/사외 port
// caller 가 nested struct 탐색 없이 바로 .AuthUserID / .AuthClientID /
// .AuthIPAddr 로 접근 가능.
//
// sso-integrations/keycloak/admin_client.go 의 ListAdminEvents 가 본 struct
// 반환 (nested authDetails → flat 매핑은 internal raw struct 으로 처리).
type KeycloakAdminEvent struct {
	Time          int64
	RealmID       string
	OperationType string // CREATE / UPDATE / DELETE / ACTION
	ResourceType  string // USER / GROUP / ROLE / CLIENT / REALM 등
	ResourcePath  string
	AuthUserID    string
	AuthClientID  string
	AuthIPAddr    string
	Error         string
}

// --- Sentinel errors (shared with view/handler.go) ---

// ErrOIDCConfigMissing — backend config 결함 (missing realm/oidc_client_id/
// oidc_client_secret 등) — Keycloak 자체는 reachable 한데 OIDC 호출 자체를
// 못 함. handler 가 marker 미부착 + 정상 OIDC 분기 + audit `revoke_status=
// config_error` + `config_error_detail` 처리.
//
// v1.0 의 view/handler.go:ErrOIDCConfigMissing 와 동일 (value alias).
var ErrOIDCConfigMissing = view.ErrOIDCConfigMissing

// ErrOIDCNetworkUnreachable — 네트워크/5xx outage (DNS 실패, connection refused,
// timeout, Keycloak 5xx 등) — Keycloak 도달 실패. handler 가 marker 부착
// (X-Keycloak-Likely-Down: true) + frontend 가 OIDC skip + 강제 /login.
//
// v1.0 의 view/handler.go:ErrOIDCNetworkUnreachable 와 동일 (value alias).
var ErrOIDCNetworkUnreachable = view.ErrOIDCNetworkUnreachable

// ErrIdentityNotFound — Keycloak admin REST 가 404 반환 시. v1.0 의
// httphelp.ErrIdentityNotFound 와 동일 (value alias).
var ErrIdentityNotFound = httphelp.ErrIdentityNotFound
