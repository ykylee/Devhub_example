// admin_client_oidc_test.go — KeycloakAdminClient 의 OIDC logout + admin
// service-account token endpoint 라우팅 검증. sprint mvs/work_260608-i-488-sign-out
// (N-8) 의 codex P1 review 정합 회귀 가드.
//
// v1.1 sprint -a follow-up PR1 (real adapter 이전) 에서 httpapi/auth_test.go
// 의 동일 test 들을 본 file 로 이동. KeycloakAdminClient 가 sso-integrations/keycloak/
// 으로 이전됨에 따라 same-package 접근 필요.
package keycloak

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TC-AUTH-LOGOUT-06 — codex P1 review 정합: OIDC logout 이 OIDC (frontend)
// client 자격증명을 사용하고 admin (backend) client 자격증명을 사용하지
// 않음을 보장. Keycloak token binding 정책 위반 시 production 401.
func TestKeycloakAdminClient_OIDCLogout_UsesOIDCClientCreds(t *testing.T) {
	// httptest.NewServer 로 mock Keycloak OIDC logout endpoint 제공.
	// form body 의 client_id 가 OIDCClientID (frontend) 인지 검증.
	var capturedClientID string
	mux := http.NewServeMux()
	mux.HandleFunc("/realms/test/protocol/openid-connect/logout", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		capturedClientID = r.FormValue("client_id")
		w.WriteHeader(http.StatusOK)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	// admin client (devhub-backend) 와 OIDC client (devhub-frontend) 가 분리.
	// OIDCClientID/OIDCClientSecret 가 ClientID/ClientSecret 와 다르게 설정.
	c := &KeycloakAdminClient{
		AdminURL:         srv.URL,
		Realm:            "test",
		ClientID:         "devhub-backend-admin", // admin client
		ClientSecret:     "admin-secret",
		OIDCClientID:     "devhub-frontend", // OIDC client (frontend)
		OIDCClientSecret: "frontend-secret",
	}

	if err := c.OIDCLogout(context.Background(), "rt-1"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if capturedClientID != "devhub-frontend" {
		t.Errorf("expected client_id=devhub-frontend (OIDC), got %q (admin ClientID leaked)", capturedClientID)
	}
}

// TC-AUTH-LOGOUT-07 — OIDC client 자격증명 미설정 시 명확한 에러 반환
// (production silent skip 방지).
func TestKeycloakAdminClient_OIDCLogout_MissingOIDCClientCreds(t *testing.T) {
	c := &KeycloakAdminClient{
		AdminURL:     "http://localhost:0",
		Realm:        "test",
		ClientID:     "admin-client",
		ClientSecret: "admin-secret",
		// OIDCClientID/OIDCClientSecret 미설정 — OIDC logout 호출 불가
	}
	err := c.OIDCLogout(context.Background(), "rt-1")
	if err == nil {
		t.Fatal("expected error when OIDC client creds missing")
	}
	if !strings.Contains(err.Error(), "oidc_client_id") || !strings.Contains(err.Error(), "oidc_client_secret") {
		t.Errorf("expected error mentioning OIDC client fields, got %v", err)
	}
}

// TC-AUTH-LOGOUT-08 — codex P1 review #3 (PR #496) 정합 회귀 가드: admin
// service-account token (accessToken) 은 AdminURL 로만 가야 한다. IssuerURL
// 이 public ingress 라도 admin endpoint 와 host 가 다르면 (docker internal
// vs public ingress) IssuerURL 사용 시 service-account token 이 public
// network 로 새서 admin-network deployment 가 깨진다. accessToken() →
// tokenEndpoint() 가 AdminURL 만 사용하는지 직접 검증.
func TestKeycloakAdminClient_AccessToken_UsesAdminURLOnly(t *testing.T) {
	// admin endpoint: 별도 host. issuer endpoint: 또 다른 host.
	// token endpoint 가 admin 으로 가야 함.
	var tokenPath string
	mux := http.NewServeMux()
	mux.HandleFunc("/realms/test/protocol/openid-connect/token", func(w http.ResponseWriter, r *http.Request) {
		tokenPath = r.URL.Path
		_ = r.ParseForm()
		// 응답: 어떤 admin client 로 호출됐는지 검증
		if r.FormValue("client_id") != "devhub-backend-admin" {
			t.Errorf("expected admin client_id=devhub-backend-admin, got %q", r.FormValue("client_id"))
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"access_token":"admin-tok","token_type":"Bearer","expires_in":60}`)
	})
	adminSrv := httptest.NewServer(mux)
	defer adminSrv.Close()

	// IssuerURL 은 adminSrv 와 다른 host 라고 가정 (실제 production 시나리오
	// 모사: admin = internal docker, issuer = public ingress). 본 테스트에선
	// adminSrv 만 띄우므로, 만약 tokenEndpoint() 가 IssuerURL 로 가면 404
	// 받게 됨. 따라서 tokenPath == "/realms/.../token" 이면 AdminURL 사용 ✅.
	// tokenPath 가 비어있거나 (네트워크 에러) / 다른 path 면 ❌.
	c := &KeycloakAdminClient{
		AdminURL:         adminSrv.URL, // admin (internal)
		Realm:            "test",
		ClientID:         "devhub-backend-admin",
		ClientSecret:     "admin-secret",
		IssuerURL:        "https://public-ingress.example.com", // admin 과 다른 host — 절대 token endpoint 로 사용되면 안 됨
		OIDCClientID:     "devhub-frontend",
		OIDCClientSecret: "frontend-secret",
	}

	tok, err := c.accessToken(context.Background())
	if err != nil {
		t.Fatalf("accessToken failed (tokenEndpoint 가 잘못된 URL 로 갔을 수 있음): %v", err)
	}
	if tok != "admin-tok" {
		t.Errorf("expected access_token=admin-tok, got %q", tok)
	}
	if tokenPath != "/realms/test/protocol/openid-connect/token" {
		t.Errorf("expected token endpoint hit at adminSrv, got path=%q (IssuerURL 로 새었을 수 있음)", tokenPath)
	}
}
