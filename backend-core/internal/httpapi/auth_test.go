package httpapi

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeBearerTokenVerifier struct {
	actor AuthenticatedActor
	err   error
	token string
}

func (v *fakeBearerTokenVerifier) VerifyBearerToken(_ context.Context, token string) (AuthenticatedActor, error) {
	v.token = token
	if v.err != nil {
		return AuthenticatedActor{}, v.err
	}
	return v.actor, nil
}

func TestBearerTokenActorWritesAuditWithoutFallbackWarning(t *testing.T) {
	orgs := newMemoryOrganizationStore()
	audits := &memoryAuditStore{}
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login:   "token-admin",
		Subject: "user-token-admin",
		Role:    "admin",
	}}
	router := NewRouter(RouterConfig{
		OrganizationStore:   orgs,
		AuditStore:          audits,
		BearerTokenVerifier: verifier,
	})

	body := []byte(`{
		"user_id": "u-token",
		"email": "token@example.com",
		"display_name": "Token User",
		"role": "developer",
		"status": "active"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}
	if verifier.token != "test-token" {
		t.Fatalf("expected verifier to receive token, got %q", verifier.token)
	}
	if rec.Header().Get("X-Devhub-Actor-Deprecated") != "" {
		t.Fatalf("did not expect X-Devhub-Actor deprecation header")
	}
	if len(audits.logs) != 1 {
		t.Fatalf("expected one audit log, got %d", len(audits.logs))
	}
	log := audits.logs[0]
	if log.ActorLogin != "token-admin" {
		t.Fatalf("expected token actor, got %+v", log)
	}
	if log.Payload["actor_source"] != "authenticated_context" {
		t.Fatalf("expected authenticated actor_source, got %+v", log.Payload)
	}
}

func TestInvalidBearerTokenReturnsUnauthorized(t *testing.T) {
	verifier := &fakeBearerTokenVerifier{err: ErrInvalidBearerToken}
	router := NewRouter(RouterConfig{
		OrganizationStore:   newMemoryOrganizationStore(),
		BearerTokenVerifier: verifier,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", rec.Code, rec.Body.String())
	}
	if verifier.token != "bad-token" {
		t.Fatalf("expected verifier to receive bad token, got %q", verifier.token)
	}
}

func TestMalformedAuthorizationHeaderReturnsUnauthorized(t *testing.T) {
	router := NewRouter(RouterConfig{OrganizationStore: newMemoryOrganizationStore()})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	req.Header.Set("Authorization", "Basic abc")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestBearerTokenWithoutVerifierDoesNotSetActor(t *testing.T) {
	orgs := newMemoryOrganizationStore()
	audits := &memoryAuditStore{}
	router := NewRouter(RouterConfig{OrganizationStore: orgs, AuditStore: audits, AuthDevFallback: true})

	body := []byte(`{
		"user_id": "u-unverified",
		"email": "unverified@example.com",
		"display_name": "Unverified User",
		"role": "developer",
		"status": "active"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer unverified-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("X-Devhub-Auth") != "bearer_unverified" {
		t.Fatalf("expected unverified bearer header")
	}
	if audits.logs[0].ActorLogin != "system" || audits.logs[0].Payload["actor_source"] != "system_fallback" {
		t.Fatalf("expected system fallback audit, got %+v", audits.logs[0])
	}
}

func TestEmptyBearerActorReturnsUnauthorized(t *testing.T) {
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{}}
	router := NewRouter(RouterConfig{
		OrganizationStore:   newMemoryOrganizationStore(),
		BearerTokenVerifier: verifier,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	req.Header.Set("Authorization", "Bearer empty-actor")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestMissingAuthorizationReturnsUnauthorizedWhenDevFallbackOff(t *testing.T) {
	router := NewRouter(RouterConfig{OrganizationStore: newMemoryOrganizationStore()})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestMissingAuthorizationPassesWhenDevFallbackOn(t *testing.T) {
	router := NewRouter(RouterConfig{OrganizationStore: newMemoryOrganizationStore(), AuthDevFallback: true})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("X-Devhub-Auth") != "dev_fallback_no_header" {
		t.Fatalf("expected dev_fallback_no_header marker, got %q", rec.Header().Get("X-Devhub-Auth"))
	}
}

func TestBearerWithoutVerifierReturnsUnauthorizedWhenDevFallbackOff(t *testing.T) {
	router := NewRouter(RouterConfig{OrganizationStore: newMemoryOrganizationStore()})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	req.Header.Set("Authorization", "Bearer some-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestPublicWebhookPathBypassesAuthentication(t *testing.T) {
	router := NewRouter(RouterConfig{EventStore: &memoryEventStore{}})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/gitea/webhooks", bytes.NewReader([]byte(`{}`)))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code == http.StatusUnauthorized {
		t.Fatalf("expected webhook path to bypass authenticateActor; got 401")
	}
}

func TestXDevhubActorIgnoredWhenDevFallbackOff(t *testing.T) {
	router := NewRouter(RouterConfig{CommandStore: &memoryCommandStore{}})

	body := []byte(`{
		"service_id": "svc-1",
		"action_type": "restart",
		"reason": "test",
		"dry_run": true,
		"idempotency_key": "k-actor-prod"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/service-actions", bytes.NewReader(body))
	req.Header.Set("X-Devhub-Actor", "spoofed-actor")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 (no Authorization, dev fallback off), got %d body=%s", rec.Code, rec.Body.String())
	}
}
