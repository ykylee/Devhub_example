package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetMeReturnsAuthenticatedActor(t *testing.T) {
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login:   "alice",
		Subject: "user-alice",
		Role:    "manager",
	}}
	router := NewRouter(RouterConfig{BearerTokenVerifier: verifier})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer t")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var body struct {
		Status string     `json:"status"`
		Data   meResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body.Data.Login != "alice" || body.Data.Subject != "user-alice" || body.Data.Role != "manager" {
		t.Errorf("unexpected actor in response: %+v", body.Data)
	}
	if body.Data.Source != "authenticated_context" {
		t.Errorf("expected source=authenticated_context, got %q", body.Data.Source)
	}
}

func TestGetMeReturns401WithoutAuthentication(t *testing.T) {
	router := NewRouter(RouterConfig{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGetMeReturns401WhenDevFallbackButNoActor(t *testing.T) {
	router := testRouter(RouterConfig{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 (no X-Devhub-Actor, dev fallback resolves to system), got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGetMeRejectsXDevhubActorHeader(t *testing.T) {
	// SEC-4 close + ADR-0006 (2026-05-13): /api/v1/me must reject inbound
	// X-Devhub-Actor outright (400 + code=x_devhub_actor_removed) rather
	// than silently ignore it. The negative test originally asserted 401
	// (header is ignored, dev fallback resolves actor to "system"); ADR-0006
	// turns that into an explicit reject so client-side usage of the dead
	// header surfaces immediately.
	router := testRouter(RouterConfig{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("X-Devhub-Actor", "dev-user")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 (ADR-0006 reject), got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "x_devhub_actor_removed") {
		t.Fatalf("expected body code=x_devhub_actor_removed, got %q", rec.Body.String())
	}
	if got := rec.Header().Get("X-Devhub-Actor-Deprecated"); got != "" {
		t.Fatalf("X-Devhub-Actor-Deprecated must not be set, got %q", got)
	}
}
