package httpapi

import (
	"net/http/httptest"
	"strings"
	"testing"
)

// TestRouter_SwaggerStagingSimulation simulates a staging env where the
// operator has only set DEVHUB_SWAGGER_ENABLED=true and left
// DEVHUB_OPENAPI_SPEC_PATH unset. The UI must serve out of the box via
// the embedded openapi.yaml fallback.
func TestRouter_SwaggerStagingSimulation(t *testing.T) {
	router := NewRouter(RouterConfig{SwaggerEnabled: true, OpenAPISpecPath: ""})

	// 1) /swagger/ → 200 + index.html
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/swagger/", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("/swagger/ → %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "DevHub API") {
		end := 120
		if rec.Body.Len() < end {
			end = rec.Body.Len()
		}
		t.Errorf("/swagger/ missing 'DevHub API' title; body[:120]=%q", rec.Body.String()[:end])
	}

	// 2) /swagger/openapi.yaml → 200 + openapi: preamble (embed fallback)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/swagger/openapi.yaml", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("/swagger/openapi.yaml → %d, want 200 (embed fallback)", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "openapi:") {
		end := 200
		if rec.Body.Len() < end {
			end = rec.Body.Len()
		}
		t.Errorf("openapi.yaml missing 'openapi:' preamble; body[:200]=%q", rec.Body.String()[:end])
	}

	// 3) /swagger/swagger-ui.css → 200
	rec = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/swagger/swagger-ui.css", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Errorf("/swagger/swagger-ui.css → %d, want 200", rec.Code)
	}

	// 4) /swagger/swagger-ui-bundle.js → 200
	rec = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/swagger/swagger-ui-bundle.js", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Errorf("/swagger/swagger-ui-bundle.js → %d, want 200", rec.Code)
	}
}
