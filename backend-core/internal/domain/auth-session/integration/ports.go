// Package integration — auth-session 도메인의 port interface 정의.
//
// v1.1 sprint -a 결정 (docs/planning/external-integrations-agentic-rag-roadmap.md §0.4).
// auth-session 도메인 layer 가 IdP (Keycloak) 와의 결합을 끊기 위한 Adapter
// pattern 적용의 1단계. 본 package 는 **interface 만** 노출하며 실제 구현은
// `sso-integrations/keycloak/` 에 위치.
//
// 역사:
//   - v1.0 (이전) — `view/auth.go` 에 `BearerTokenVerifier` interface + `view/handler.go` 에
//     `IdentityAdmin` / `OIDCLogoutClient` interface 정의. domain layer 가
//     IdP 와 직접 결합 (KeycloakJWKSVerifier / KeycloakAdminClient 가
//     domain/auth-session/service/ + httpapi/ 에 위치).
//   - v1.1 sprint -a (현재) — 본 package 로 interface 이동. view/ 의 interface 는
//     deprecated alias 로 유지 (backward compat). 신규 호출은 본 package
//     alias 사용.
//   - v1.1 sprint -a follow-up (예정) — `sso-integrations/keycloak/` 로 실제 구현 이동.
//     v1.1 sprint -a PR 은 **interface + saovae stub + main wiring** 만.
//
// Tier 분류: **공용** (사외 + 사내 양쪽 build 가 import 가능). interface 만
// 노출하므로 사내 한정 정보 미포함. 사외 build 시 saovae stub (build tag) 자동 사용.

package integration

import (
	"context"
	"time"

	"github.com/devhub/backend-core/internal/domain/auth-session/view"
	"github.com/devhub/backend-core/internal/httpapi"
	"github.com/devhub/backend-core/internal/shared/httphelp"
)

// --- Public type aliases (re-export from view for port callers) ---

// AuthenticatedActor — JWT claim 추출 결과. v1.0 의 view.AuthenticatedActor 의
// type alias. v1.1 sprint -a follow-up 에서 view.AuthenticatedActor 정의가
// 본 package 로 이전될 예정 (별도 PR).
type AuthenticatedActor = view.AuthenticatedActor

// --- Core port interfaces (migrated from view/) ---

// BearerTokenVerifier — OIDC Bearer JWT 검증. view/auth.go:59 의 동일 interface.
// auth middleware 가 hot path 에서 호출. 구현: sso-integrations/keycloak/ (JWKS 기반).
type BearerTokenVerifier = view.BearerTokenVerifier

// IdentityAdmin — Keycloak admin REST (user lifecycle + logout). view/handler.go:27
// 의 동일 interface. 사내 IdP lifecycle 관리용. 구현: sso-integrations/keycloak/.
type IdentityAdmin = view.IdentityAdmin

// OIDCLogoutClient — OIDC /protocol/openid-connect/logout endpoint wrapper.
// view/handler.go:197 의 동일 interface. user-facing OIDC endpoint (admin
// endpoint 와 분리). 구현: sso-integrations/keycloak/.
type OIDCLogoutClient = view.OIDCLogoutClient

// --- NEW port (defined in integration/, not view/) ---

// KeycloakEventPort — Keycloak user/admin event polling + webhook ingest.
// audit-ops/service/keycloak_event_puller.go:30 의 KeycloakEventLister 와
// 동등. **NEW port** (sprint -a 에서 신규 정의; 기존엔 interface 없이 직접
// struct 사용).
//
// 구현: sso-integrations/keycloak/ 의 event_puller.go + webhook.go.
type KeycloakEventPort interface {
	// ListUserEvents — Keycloak user events (login, logout, register, ...) polling.
	ListUserEvents(ctx context.Context, dateFrom time.Time, max int) ([]KeycloakUserEvent, error)

	// ListAdminEvents — Keycloak admin events (CREATE_USER, UPDATE_ROLE, ...) polling.
	ListAdminEvents(ctx context.Context, dateFrom time.Time, max int) ([]KeycloakAdminEvent, error)
}

// KeycloakUserEvent — Keycloak event 의 user-side mirror. v1.0 의
// httpapi.KeycloakUserEvent 와 field-compatible. sprint -a follow-up 에서 v1.0
// 의 mirror 들은 본 alias 로 통합 (별도 PR — 본 PR 에서 alias 만).
type KeycloakUserEvent = httpapi.KeycloakUserEvent

// KeycloakAdminEvent — Keycloak event 의 admin-side mirror. v1.0 의
// httpapi.KeycloakAdminEvent 와 field-compatible. sprint -a follow-up 에서
// alias 로 통합.
type KeycloakAdminEvent = httpapi.KeycloakAdminEvent

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
