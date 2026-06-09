package httpapi

import (
	"bytes"
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

// TestRouter_SwaggerSpecEmptyOmitsMount verifies that an empty OpenAPISpecPath
// does not mount the /swagger/openapi.yaml route (codex P2 fix, PR #508).
func TestRouter_SwaggerSpecEmptyOmitsMount(t *testing.T) {
	router := NewRouter(RouterConfig{SwaggerEnabled: true, OpenAPISpecPath: ""})
	req := httptest.NewRequest(http.MethodGet, "/swagger/openapi.yaml", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("GET /swagger/openapi.yaml with empty spec path: got %d, want %d (spec must be omitted)", rec.Code, http.StatusNotFound)
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
