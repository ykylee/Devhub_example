package httpapi

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// TestRouter_SwaggerDisabledByDefault verifies that /swagger/* returns 404
// when SwaggerEnabled is false (zero-value default).  This is the production-safe
// default — no Swagger UI exposure without explicit opt-in.
func TestRouter_SwaggerDisabledByDefault(t *testing.T) {
	router := NewRouter(RouterConfig{})
	req := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("GET /swagger/index.html with SwaggerEnabled=false: got %d, want %d", rec.Code, http.StatusNotFound)
	}
}

// TestRouter_SwaggerEnabledServesUI verifies that the Swagger UI mount is
// active when SwaggerEnabled=true — both the route registration AND the actual
// asset lookup (codex P1 review, 2026-06-10: previously only checked the
// route registration, missing the fs.Sub sub-mount issue that returned 404
// for GET /swagger/index.html).
func TestRouter_SwaggerEnabledServesUI(t *testing.T) {
	router := NewRouter(RouterConfig{SwaggerEnabled: true})

	found := false
	for _, r := range router.Routes() {
		if r.Method == http.MethodGet && strings.HasPrefix(r.Path, "/swagger") {
			found = true
			break
		}
	}
	if !found {
		t.Error("SwaggerEnabled=true but no /swagger route registered in router.Routes()")
	}

	// Regression guard: GET /swagger/ returns 200 with embedded index.html.
	// Without fs.Sub (codex P1 review), the FS root would still contain the
	// swaggerui/ prefix and this would 404.  Trailing-slash form is used
	// because gin StaticFS 301-redirects /swagger/index.html to "./".
	req := httptest.NewRequest(http.MethodGet, "/swagger/", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("GET /swagger/index.html with SwaggerEnabled=true: got %d, want %d (likely fs.Sub missing)", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "swagger-ui") {
		body := rec.Body.String()
		preview := body
		if len(preview) > 200 {
			preview = preview[:200]
		}
		t.Errorf("GET /swagger/index.html body missing 'swagger-ui' marker; got first 200 chars: %q", preview)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TestRouter_SwaggerSpecWhenConfigured verifies that /swagger/openapi.yaml is
// served from the configured OpenAPISpecPath when both SwaggerEnabled and the
// path are set.
func TestRouter_SwaggerSpecWhenConfigured(t *testing.T) {
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "test-spec.yaml")
	content := []byte("openapi: 3.0.0\ninfo:\n  title: Test\n  version: 0.1.0\n")
	if err := os.WriteFile(yamlPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	router := NewRouter(RouterConfig{
		SwaggerEnabled:  true,
		OpenAPISpecPath: yamlPath,
	})
	req := httptest.NewRequest(http.MethodGet, "/swagger/openapi.yaml", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /swagger/openapi.yaml: got %d, want %d", rec.Code, http.StatusOK)
	}
	if !bytes.Equal(rec.Body.Bytes(), content) {
		t.Errorf("GET /swagger/openapi.yaml body mismatch\n got  %q\n want %q", rec.Body.String(), string(content))
	}
}

// TestRouter_SwaggerSpecEmptyEmbedsFallback verifies that an empty
// OpenAPISpecPath falls back to the embedded openapi.yaml (build-time
// copy of docs/openapi.yaml) so the spec is reachable out of the box in
// staging/test without DEVHUB_OPENAPI_SPEC_PATH plumbing. This supersedes
// the PR #508 silent-404 behaviour: the operator's request (option D in
// the prior conversation) is for the UI to just work in test environments.
func TestRouter_SwaggerSpecEmptyEmbedsFallback(t *testing.T) {
	router := NewRouter(RouterConfig{SwaggerEnabled: true, OpenAPISpecPath: ""})
	req := httptest.NewRequest(http.MethodGet, "/swagger/openapi.yaml", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /swagger/openapi.yaml with empty spec path: got %d, want %d (embed fallback must serve the spec)", rec.Code, http.StatusOK)
	}
	body := rec.Body.Bytes()
	if !bytes.Contains(body, []byte("openapi:")) {
		t.Fatalf("embedded openapi.yaml missing 'openapi:' preamble; body[:200]=%q", body[:min(200, len(body))])
	}
}

// TestRouter_SwaggerIndexHasDynamicSpecURL guards the relative-path
// regression. index.html must compute the spec URL from
// window.location.pathname so the same asset works behind nginx's
// /devhub/swagger/ rewrite as well as at the bare /swagger/ path, including
// nginx's 301-with-trailing-slash form (.../index.html/). PR #513 (codex P2)
// added file-vs-directory detection so directory-only pathnames are
// preserved as the base.
func TestRouter_SwaggerIndexHasDynamicSpecURL(t *testing.T) {
	router := NewRouter(RouterConfig{SwaggerEnabled: true})
	req := httptest.NewRequest(http.MethodGet, "/swagger/", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /swagger/: got %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	wantSubstrings := []string{
		`pathname.replace(/\/+$/, "") || "/"`,
		`trimmed.lastIndexOf("/")`,
		`lastSeg.includes(".")`,
		`base + "openapi.yaml"`,
	}
	for _, s := range wantSubstrings {
		if !strings.Contains(body, s) {
			t.Errorf("index.html missing required snippet %q (dynamic base spec URL). Body fragment: %q", s, body)
		}
	}
}

// TestRouter_SwaggerVendoredAssets verifies that swagger-ui vendor assets
// (CSS + bundle + standalone-preset) are embedded and served. PR #508 follow-up.
func TestRouter_SwaggerVendoredAssets(t *testing.T) {
	router := NewRouter(RouterConfig{SwaggerEnabled: true})
	for _, p := range []string{
		"/swagger/swagger-ui.css",
		"/swagger/swagger-ui-bundle.js",
		"/swagger/swagger-ui-standalone-preset.js",
	} {
		req := httptest.NewRequest(http.MethodGet, p, nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("GET %s: got %d, want %d (vendor asset must be embedded)", p, rec.Code, http.StatusOK)
		}
		body := rec.Body.String()
		if body == "" {
			t.Errorf("GET %s: empty body (vendor asset content missing)", p)
		}
	}
}
// TestTrustedProxiesFromEnv covers the DEVHUB_TRUSTED_PROXIES contract
// (PR-D follow-up, work_260512-i). Empty / "none" → nil keeps the silent
// default. "*" expands to dual-stack any. Comma lists are trimmed and
// emptied.
func TestTrustedProxiesFromEnv(t *testing.T) {
	cases := []struct {
		name string
		env  string
		want []string
	}{
		{name: "unset → nil", env: "", want: nil},
		{name: "none → nil (alias)", env: "none", want: nil},
		{name: "NONE → nil (case insensitive)", env: "NONE", want: nil},
		{name: "wildcard → dual-stack any", env: "*", want: []string{"0.0.0.0/0", "::/0"}},
		{name: "single CIDR", env: "10.0.0.0/8", want: []string{"10.0.0.0/8"}},
		{name: "comma list with whitespace", env: "10.0.0.0/8 , 192.168.1.5", want: []string{"10.0.0.0/8", "192.168.1.5"}},
		{name: "all empty entries → nil", env: " , , ", want: nil},
		{name: "leading + trailing whitespace", env: "  10.0.0.0/8  ", want: []string{"10.0.0.0/8"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("DEVHUB_TRUSTED_PROXIES", tc.env)
			got := trustedProxiesFromEnv()
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("trustedProxiesFromEnv() = %#v, want %#v", got, tc.want)
			}
		})
	}
}

// ADR-0029 §6 (e) P2 — swagger UI 자체에 system_admin gate. 미인증 → 401.
func TestRouter_SwaggerAdminGate_UnauthenticatedReturns401(t *testing.T) {
	verifier := &fakeBearerTokenVerifier{err: errors.New("not used")}
	router := NewRouter(RouterConfig{
		SwaggerEnabled:             true,
		SwaggerRequireSystemAdmin:  true,
		BearerTokenVerifier:        verifier,
		OpenAPISpecPath:            "",
	})
	req := httptest.NewRequest(http.MethodGet, "/swagger/", nil)
	req.Header.Set("Authorization", "")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d (body=%s)", rec.Code, rec.Body.String())
	}
	var resp struct {
		Code string `json:"code"`
	}
	decodeJSON(t, rec.Body.Bytes(), &resp)
	if resp.Code != "swagger_unauthenticated" {
		t.Errorf("code = %q, want swagger_unauthenticated", resp.Code)
	}
	if verifier.token != "" {
		t.Errorf("verifier should not be called when no Authorization header; got token=%q", verifier.token)
	}
}

// ADR-0029 §6 (e) P2 — non-system_admin role (developer) → 403.
func TestRouter_SwaggerAdminGate_DeveloperRoleReturns403(t *testing.T) {
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login:   "dev-user",
		Subject: "user-dev",
		Role:    "developer",
	}}
	router := NewRouter(RouterConfig{
		SwaggerEnabled:            true,
		SwaggerRequireSystemAdmin: true,
		BearerTokenVerifier:       verifier,
		OpenAPISpecPath:           "",
	})
	req := httptest.NewRequest(http.MethodGet, "/swagger/", nil)
	req.Header.Set("Authorization", "Bearer dev-jwt")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d (body=%s)", rec.Code, rec.Body.String())
	}
	var resp struct {
		Code string `json:"code"`
	}
	decodeJSON(t, rec.Body.Bytes(), &resp)
	if resp.Code != "swagger_admin_required" {
		t.Errorf("code = %q, want swagger_admin_required", resp.Code)
	}
}

// ADR-0029 §6 (e) P2 — system_admin role → swagger UI 통과. gin StaticFS 가
// /swagger/index.html → /swagger/ 로 301 redirect 응답 (trailing-slash 처리).
// gate 가 통과했으므로 301 정상 — 401/403 이면 gate 가 막힌 것.
// gin /swagger/*filepath 가 single segment 매치. /swagger/index.html 사용.
func TestRouter_SwaggerAdminGate_SystemAdminPasses(t *testing.T) {
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login:   "admin-user",
		Subject: "user-admin",
		Role:    "system_admin",
	}}
	router := NewRouter(RouterConfig{
		SwaggerEnabled:            true,
		SwaggerRequireSystemAdmin: true,
		BearerTokenVerifier:       verifier,
		OpenAPISpecPath:           "",
	})
	req := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	req.Header.Set("Authorization", "Bearer admin-jwt")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	// gin StaticFS 가 /swagger/index.html → /swagger/ 로 301 redirect 응답 (trailing-slash).
	// 301 = gate 가 system_admin 으로 통과했다는 신호 (gate 가 막혔으면 401/403).
	if rec.Code != http.StatusMovedPermanently {
		t.Fatalf("expected 301 (system_admin passes gate → StaticFS redirect), got %d (body=%s)", rec.Code, rec.Body.String())
	}
	if verifier.token != "admin-jwt" {
		t.Errorf("verifier should be called with admin-jwt; got token=%q", verifier.token)
	}
}

// ADR-0029 §6 (e) P2 — invalid bearer token (verifier fail) → 401.
func TestRouter_SwaggerAdminGate_InvalidTokenReturns401(t *testing.T) {
	verifier := &fakeBearerTokenVerifier{err: errors.New("invalid jwt")}
	router := NewRouter(RouterConfig{
		SwaggerEnabled:            true,
		SwaggerRequireSystemAdmin: true,
		BearerTokenVerifier:       verifier,
		OpenAPISpecPath:           "",
	})
	req := httptest.NewRequest(http.MethodGet, "/swagger/", nil)
	req.Header.Set("Authorization", "Bearer invalid-jwt")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d (body=%s)", rec.Code, rec.Body.String())
	}
	var resp struct {
		Code string `json:"code"`
	}
	decodeJSON(t, rec.Body.Bytes(), &resp)
	if resp.Code != "swagger_token_invalid" {
		t.Errorf("code = %q, want swagger_token_invalid", resp.Code)
	}
}
