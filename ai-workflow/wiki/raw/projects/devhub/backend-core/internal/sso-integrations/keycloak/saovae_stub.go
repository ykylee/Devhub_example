// saovae_stub.go — 사외 build 용 Keycloak adapter stub.
//
// 정책 (`docs/governance/worker_division.md` §6 + ADR-0030 §2.3):
//   - runtime injection: 단일 binary, main.go 에서 DEVHUB_BUILD_TIER
//     env var 로 stub/real 선택.
//   - 본 stub 은 **사외 build 시 default** (DEVHUB_BUILD_TIER=external 또는
//     unset). 사내 build 시 main.go 가 sso-integrations/keycloak/real.go
//     의 real adapter 로 wiring (sprint -a follow-up 별도 PR).
//
// 빌드 tag 정책:
//   - `//go:build saovae` — 명시적 saovae build
//   - `//go:build saovae || dev` — saovae + dev (default) 시
//   - **default = 사외 = stub** — runtime injection 으로 사내 시 real
//     adapter 가 stub 의 의존성 대체.
//
// stub 의 한계 (production 사용 금지):
//   - JWT 서명 검증 없음 — 모든 token 이 `dev-stub-user` 로 검증 (중간자 공격 가능)
//   - Keycloak admin REST 호출 없음 — 모든 admin 메서드가 mock identity 반환
//   - Event polling 없음 — 빈 list 반환
//   - Webhook ingest 없음 — handler 가 nil
//
// stub 의 의도: 사외 build (CI / e2e / frontend / 일반 dev) 가 Keycloak
// 인프라 의존성 0 으로 통과. 사내 staging 검증은 main.go 의 real adapter
// wiring 으로 별도 환경에서 실행.

package keycloak

import (
	"context"
	"net/http"
	"time"

	"github.com/devhub/backend-core/internal/domain/auth-session/integration"
)

// --- Port implementations (사외 build stub) ---

// stubBearerTokenVerifier — 모든 JWT 를 dev actor 로 검증. signature/expiry/aud/iss
// 무시. dev fixture.
type stubBearerTokenVerifier struct{}

func (stubBearerTokenVerifier) VerifyBearerToken(_ context.Context, _ string) (integration.AuthenticatedActor, error) {
	return integration.AuthenticatedActor{
		Login:       "dev-stub-user",
		Subject:     "sub-stub-001",
		Role:        "system_admin",
		Email:       "dev-stub@devhub.local",
		DisplayName: "Dev Stub User",
	}, nil
}

// NewBearerTokenVerifierStub — main.go 가 runtime injection 으로 호출.
func NewBearerTokenVerifierStub() integration.BearerTokenVerifier {
	return stubBearerTokenVerifier{}
}

// stubIdentityAdmin — admin lifecycle 메서드가 mock identity 반환.
type stubIdentityAdmin struct{}

func (stubIdentityAdmin) FindIdentityByUserID(_ context.Context, userID string) (string, error) {
	// 사외 stub: userID 를 그대로 identityID 로 사용 (Keycloak UUID 형식 무시).
	return "stub-identity-" + userID, nil
}

func (stubIdentityAdmin) LogoutUserSession(_ context.Context, _ string) error {
	return nil
}

func NewIdentityAdminStub() integration.IdentityAdmin {
	return stubIdentityAdmin{}
}

// stubOIDCLogoutClient — OIDC logout 을 success 로 stub.
type stubOIDCLogoutClient struct{}

func (stubOIDCLogoutClient) OIDCLogout(_ context.Context, _ string) error {
	return nil
}

func NewOIDCLogoutClientStub() integration.OIDCLogoutClient {
	return stubOIDCLogoutClient{}
}

// stubKeycloakEventPort — event polling + webhook 의 stub. 빈 list 반환.
type stubKeycloakEventPort struct{}

func (stubKeycloakEventPort) ListUserEvents(_ context.Context, _ time.Time, _ int) ([]integration.KeycloakUserEvent, error) {
	return nil, nil
}

func (stubKeycloakEventPort) ListAdminEvents(_ context.Context, _ time.Time, _ int) ([]integration.KeycloakAdminEvent, error) {
	return nil, nil
}

func NewKeycloakEventPortStub() integration.KeycloakEventPort {
	return stubKeycloakEventPort{}
}

// HTTPHandler — webhook ingest handler 의 stub. 503 Service Unavailable.
func NewWebhookHandlerStub() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"error":"keycloak webhook ingest disabled in saovae build"}`))
	})
}
