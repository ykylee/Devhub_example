package httpapi

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
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
		Role:    "system_admin",
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

// orgStoreGetUserError wraps memoryOrganizationStore and overrides GetUser
// to inject store errors that aren't ErrNotFound (schema drift, connection
// failures, etc.). Used to pin authenticateActor's surface-the-error path.
type orgStoreGetUserError struct {
	*memoryOrganizationStore
	err error
}

func (s *orgStoreGetUserError) GetUser(_ context.Context, _ string) (domain.AppUser, error) {
	return domain.AppUser{}, s.err
}

// Regression guard: when GetUser fails with a schema/connection error, the
// middleware must (a) still complete the request using the token role
// claim, and (b) log loud enough that operators can spot a missing
// migration. The silent-fallback bug once routed every actor to a single
// default role until we found the underlying SQL error by accident.
func TestAuthenticateActor_LogsNonNotFoundGetUserError(t *testing.T) {
	var buf bytes.Buffer
	oldOut := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(oldOut)

	orgs := &orgStoreGetUserError{
		memoryOrganizationStore: newMemoryOrganizationStore(),
		err:                     errors.New("ERROR: column \"kratos_identity_id\" does not exist (SQLSTATE 42703)"),
	}
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login:   "bob",
		Subject: "user-bob",
		Role:    "manager", // token claim says manager
	}}
	router := NewRouter(RouterConfig{
		OrganizationStore:   orgs,
		BearerTokenVerifier: verifier,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer t")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	// Token-claim role must survive the GetUser failure (no silent
	// collapse to actor.Role default of "").
	if !strings.Contains(rec.Body.String(), `"role":"manager"`) {
		t.Errorf("expected role to fall back to token claim 'manager', body = %s", rec.Body.String())
	}
	if !strings.Contains(buf.String(), `[authenticateActor] GetUser "bob" failed`) {
		t.Errorf("expected GetUser error to be logged; got: %q", buf.String())
	}
}

// memoryOrganizationStore.GetUser returns ErrNotFound for users that have
// not yet been onboarded. That's a normal state for a freshly-issued
// Hydra token, not a misconfiguration — must not generate log noise.
func TestAuthenticateActor_DoesNotLogGetUserNotFound(t *testing.T) {
	var buf bytes.Buffer
	oldOut := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(oldOut)

	orgs := &orgStoreGetUserError{
		memoryOrganizationStore: newMemoryOrganizationStore(),
		err:                     fmt.Errorf("user new-user: %w", store.ErrNotFound),
	}
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login:   "new-user",
		Subject: "new-user",
		Role:    "developer",
	}}
	router := NewRouter(RouterConfig{
		OrganizationStore:   orgs,
		BearerTokenVerifier: verifier,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer t")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(buf.String(), "GetUser") {
		t.Errorf("ErrNotFound is a normal state and must not log; got: %q", buf.String())
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

func TestXDevhubActorIgnoredEvenWhenDevFallbackOn(t *testing.T) {
	commandStore := &memoryCommandStore{}
	router := testRouter(RouterConfig{CommandStore: commandStore})

	body := []byte(`{
		"service_id": "svc-1",
		"action_type": "restart",
		"reason": "test",
		"dry_run": true,
		"idempotency_key": "k-actor-dev"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/service-actions", bytes.NewReader(body))
	req.Header.Set("X-Devhub-Actor", "spoofed-actor")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202 (dev fallback bypasses role gate), got %d body=%s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("X-Devhub-Actor-Deprecated"); got != "" {
		t.Fatalf("X-Devhub-Actor-Deprecated must not be set after SEC-4 removal, got %q", got)
	}
	if len(commandStore.commands) != 1 {
		t.Fatalf("expected one command, got %d", len(commandStore.commands))
	}
	if commandStore.commands[0].ActorLogin != "system" {
		t.Errorf("X-Devhub-Actor must be ignored, expected actor=system, got %q", commandStore.commands[0].ActorLogin)
	}
}
