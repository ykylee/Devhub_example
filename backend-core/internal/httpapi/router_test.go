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
// active when SwaggerEnabled=true.  The embedded asset directory currently
// contains only .gitkeep (index.html is a separate task), so the body-content
// assertion is deferred.  This test validates that the /swagger/* route group
// is registered.
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
