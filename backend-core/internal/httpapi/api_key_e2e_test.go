package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/devhub/backend-core/internal/domain"
)

// TestAPIKeyEndToEnd_SwaggerServes — E2E: 서버 부팅 없이 라우터 + 미들웨어 +
// 정적 자산 서빙 + API key 경로가 모두 정상 동작하는지 단일 통합 테스트.
// 운영 부팅 (DB_URL 등) 없이 in-memory store + API key 로 모든 핵심 경로를
// verify. 본 테스트는 task T-00316f81 의 (2)~(7) 시나리오의 회귀 가드.
func TestAPIKeyEndToEnd_SwaggerServes(t *testing.T) {
	router := NewRouter(RouterConfig{
		SwaggerEnabled: true,
		APIKey:         "e2e-test-key-2026-06-09",
	})

	// (a) /swagger/ → 200 + swagger-ui marker
	req := httptest.NewRequest(http.MethodGet, "/swagger/", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("GET /swagger/ got %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "swagger-ui") {
		t.Errorf("/swagger/ body missing swagger-ui marker")
	}

	// (b) /swagger/openapi.yaml → 200 + 5.6%→73% 보강 반영
	req = httptest.NewRequest(http.MethodGet, "/swagger/openapi.yaml", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /swagger/openapi.yaml got %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	// 보강된 경로 마커 검사
	wantMarkers := []string{
		"/api/v1/me:",
		"/api/v1/platforms:",
		"/api/v1/repositories:",
		"/api/v1/integration/providers:",
		"/api/v1/dev-requests:",
		"/api/v1/audit-logs:",
		"staticTokenAuth:",
	}
	for _, m := range wantMarkers {
		if !strings.Contains(body, m) {
			t.Errorf("openapi.yaml missing required marker %q (P0/P1 endpoint 보강 회귀)", m)
		}
	}
}

// TestAPIKeyEndToEnd_PublicReadEnvelopes — 공개 read-only endpoint 가 API key 로
// 인증되어 정상 흐름으로 진입 (501/503 등 store-missing 응답은 정상 — middleware
// 가 인증은 통과했다는 의미). X-Devhub-Auth=api_key header 부착 + audit_source=system
// 검증. ADR-0029 §3.1.
func TestAPIKeyEndToEnd_PublicReadEnvelopes(t *testing.T) {
	orgs := newMemoryOrganizationStore()
	// system_admin row 사전 시드 (api-key caller 가 system_admin role 로 매핑되어도
	// enforceRoutePermission 의 system_admin early return 통과).
	if _, err := orgs.CreateUser(context.Background(), domain.CreateUserInput{
		UserID: "api-key", Email: "api-key@example.com", DisplayName: "API Key",
		Role: domain.AppRoleSystemAdmin, Status: domain.UserStatusActive,
		Type: domain.UserTypeSystem,
	}); err != nil {
		t.Fatalf("seed api-key user: %v", err)
	}

	router := NewRouter(RouterConfig{
		APIKey:            "e2e-test-key-2026-06-09",
		OrganizationStore: orgs,
		// BearerTokenVerifier 비설정 — JWT 포맷 입력 시 verifier 호출 시도가
		// dev fallback 으로 fallback 되어 system_admin actor 가 됨.
		AuthDevFallback: true,
	})

	cases := []struct {
		method, path, wantAuth string
	}{
		{http.MethodGet, "/api/v1/me", "api_key"},
		{http.MethodGet, "/api/v1/dashboard/metrics", "api_key"},
		{http.MethodGet, "/api/v1/repositories", "api_key"},
		{http.MethodGet, "/api/v1/issues", "api_key"},
		{http.MethodGet, "/api/v1/risks", "api_key"},
		{http.MethodGet, "/api/v1/infra/services", "api_key"},
		{http.MethodGet, "/api/v1/organization/hierarchy", "api_key"},
		{http.MethodGet, "/api/v1/audit-logs", "api_key"},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			req.Header.Set("Authorization", "Bearer e2e-test-key-2026-06-09")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			// 모든 인증 케이스에서 미들웨어 통과 = 4xx/5xx 가 아닌 200/4xx 응답.
			// store-missing 으로 5xx 가능. 401 이면 안 됨 (auth 실패).
			if rec.Code == http.StatusUnauthorized {
				t.Errorf("API key caller at %s: got 401, want pass-through", tc.path)
			}
			if rec.Header().Get("X-Devhub-Auth") != tc.wantAuth {
				t.Errorf("API key at %s: X-Devhub-Auth=%q, want %q", tc.path, rec.Header().Get("X-Devhub-Auth"), tc.wantAuth)
			}
		})
	}
}

// TestAPIKeyEndToEnd_KeycloakJWTStillWorks — Keycloak JWT (3-part) 가 verifier
// 분기로 진입 (dev fallback 아님). 1차 PR 의 회귀 가드: API key 분기 추가로
// Keycloak path 가 깨지지 않아야 함 — JWT 가 verifier 호출로 라우팅되고,
// verifier 가 token 을 수신한 후 actor 를 반환하면 200.
func TestAPIKeyEndToEnd_KeycloakJWTStillWorks(t *testing.T) {
	orgs := newMemoryOrganizationStore()
	if _, err := orgs.CreateUser(context.Background(), domain.CreateUserInput{
		UserID: "jwt-admin", Email: "jwt-admin@example.com", DisplayName: "JWT Admin",
		Role: domain.AppRoleSystemAdmin, Status: domain.UserStatusActive,
		Type: domain.UserTypeHuman,
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}

	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login: "jwt-admin", Subject: "user-jwt-admin", Role: "system_admin",
	}}
	router := NewRouter(RouterConfig{
		APIKey:              "e2e-test-key-2026-06-09",
		BearerTokenVerifier: verifier,
		OrganizationStore:   orgs,
		// AuthDevFallback: false (default) — JWT 가 verifier 실패 시 401 로 떨어져야 함
	})

	jwt := "eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJhbGljZSJ9.signature"
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer "+jwt)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("JWT-format bearer at /me with valid verifier: got %d, want 200 (Keycloak path 회귀)", rec.Code)
	}
	if verifier.token != jwt {
		t.Errorf("verifier should have received JWT %q, got %q", jwt, verifier.token)
	}
	if rec.Header().Get("X-Devhub-Auth") == "api_key" {
		t.Errorf("JWT-format bearer was misclassified as api_key path (regression)")
	}
}

// TestAPIKeyEndToEnd_NoAuthKeycloakStillLocked — Keycloak verifier 가 dev
// fallback 없이 미설정 시 JWT 포맷 입력은 401. 본 1차 PR 의 보안 invariant.
func TestAPIKeyEndToEnd_NoAuthKeycloakStillLocked(t *testing.T) {
	router := NewRouter(RouterConfig{
		APIKey:            "e2e-test-key-2026-06-09",
		OrganizationStore: newMemoryOrganizationStore(),
		// BearerTokenVerifier: nil (default)
		// AuthDevFallback: false (default)
	})

	jwt := "eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJhbGljZSJ9.signature"
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer "+jwt)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("JWT format with no verifier/dev-fallback: got %d, want 401", rec.Code)
	}
}

// TestAPIKeyEndToEnd_StaticKeyVerifierError — API key 분기가 verifier
// 호출보다 우선 동작. verifier 가 invalid token error 를 반환해도 API key 가
// 유효하면 통과.
func TestAPIKeyEndToEnd_StaticKeyVerifierError(t *testing.T) {
	orgs := newMemoryOrganizationStore()
	if _, err := orgs.CreateUser(context.Background(), domain.CreateUserInput{
		UserID: "api-key", Email: "api-key@example.com", DisplayName: "API Key",
		Role: domain.AppRoleSystemAdmin, Status: domain.UserStatusActive,
		Type: domain.UserTypeSystem,
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}

	verifier := &fakeBearerTokenVerifier{err: errors.New("verifier would reject this token")}
	router := NewRouter(RouterConfig{
		APIKey:              "e2e-test-key-2026-06-09",
		BearerTokenVerifier: verifier,
		OrganizationStore:   orgs,
		AuthDevFallback:     true,
	})

	// 정적 API key 는 JWT 가 아니므로 verifier 가 호출되지 않고 인증 통과.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer e2e-test-key-2026-06-09")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code == http.StatusUnauthorized {
		t.Errorf("valid static API key rejected (regression): got 401, want pass-through")
	}
	if verifier.token != "" {
		t.Errorf("static API key should not reach Keycloak verifier; got token=%q", verifier.token)
	}
}
