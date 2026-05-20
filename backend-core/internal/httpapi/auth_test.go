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
	// ADR-0020 sub-carve B (sprint -i): admin user pre-seed → lazy auto-create skip.
	if _, err := orgs.CreateUser(context.Background(), domain.CreateUserInput{
		UserID: "token-admin", Email: "token-admin@example.com", DisplayName: "Token Admin",
		Role: domain.AppRoleSystemAdmin, Status: domain.UserStatusActive,
		Type: domain.UserTypeHuman,
	}); err != nil {
		t.Fatalf("seed token-admin: %v", err)
	}
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
		err:                     errors.New("ERROR: column \"idp_subject\" does not exist (SQLSTATE 42703)"),
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

// TestAuthenticateActor_BackfillsIdPSubjectOnFirstLogin — sprint claude/work_260519-t
// (ADR-0019 §5.3 sprint -j codex review #9 #2 backend 확장 carve #4). user row 의
// idp_subject 가 빈 상태로 첫 로그인 시 actor.Subject 로 lazy backfill 검증.
func TestAuthenticateActor_BackfillsIdPSubjectOnFirstLogin(t *testing.T) {
	orgs := newMemoryOrganizationStore()
	// pre-existing user with empty IdPSubject (pre-migration 000030 or pre-OIDC migration state)
	if _, err := orgs.CreateUser(context.Background(), domain.CreateUserInput{
		UserID:      "alice",
		Email:       "alice@example.com",
		DisplayName: "Alice",
		Role:        domain.AppRoleDeveloper,
		Status:      domain.UserStatusActive,
		Type:        domain.UserTypeHuman,
	}); err != nil {
		t.Fatalf("seed user: %v", err)
	}

	// pre-state: idp_subject should be empty
	pre, err := orgs.GetUser(context.Background(), "alice")
	if err != nil {
		t.Fatalf("pre GetUser: %v", err)
	}
	if pre.IdPSubject != "" {
		t.Fatalf("seed user should have empty IdPSubject, got %q", pre.IdPSubject)
	}

	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login:   "alice",
		Subject: "kc-uuid-alice", // Keycloak sub claim
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

	// post-state: idp_subject should be backfilled with actor.Subject
	post, err := orgs.GetUser(context.Background(), "alice")
	if err != nil {
		t.Fatalf("post GetUser: %v", err)
	}
	if post.IdPSubject != "kc-uuid-alice" {
		t.Fatalf("IdPSubject after first login = %q; want %q (lazy backfill)", post.IdPSubject, "kc-uuid-alice")
	}
}

// TestAuthenticateActor_DoesNotResetIdPSubjectIfAlreadySet — 이미 idp_subject 가 있는
// user 는 SetIdPSubject 호출 안 함 (불필요한 DB write 회피).
func TestAuthenticateActor_DoesNotResetIdPSubjectIfAlreadySet(t *testing.T) {
	orgs := newMemoryOrganizationStore()
	if _, err := orgs.CreateUser(context.Background(), domain.CreateUserInput{
		UserID:      "bob",
		Email:       "bob@example.com",
		DisplayName: "Bob",
		Role:        domain.AppRoleManager,
		Status:      domain.UserStatusActive,
		Type:        domain.UserTypeHuman,
	}); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	// IdPSubject 를 별도 SetIdPSubject 로 사전 세팅 (CreateUserInput 에 IdPSubject 필드 없음)
	if err := orgs.SetIdPSubject(context.Background(), "bob", "kc-uuid-bob-existing"); err != nil {
		t.Fatalf("seed IdPSubject: %v", err)
	}

	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login:   "bob",
		Subject: "kc-uuid-bob-different", // 다른 sub — overwrite 시도 검증
		Role:    "manager",
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

	// idp_subject 는 변경 없음 — 이미 있던 값 유지
	post, err := orgs.GetUser(context.Background(), "bob")
	if err != nil {
		t.Fatalf("post GetUser: %v", err)
	}
	if post.IdPSubject != "kc-uuid-bob-existing" {
		t.Fatalf("IdPSubject = %q; want %q (must not overwrite)", post.IdPSubject, "kc-uuid-bob-existing")
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

func TestXDevhubActorRejectedWhenDevFallbackOff(t *testing.T) {
	// ADR-0006: inbound X-Devhub-Actor is rejected outright (400 +
	// code=x_devhub_actor_removed). Previously this asserted 401 (header
	// ignored, no Authorization → unauth). ADR-0006 makes the header
	// reject take precedence over the auth gate so client-side usage
	// surfaces immediately.
	router := NewRouter(RouterConfig{CommandStore: &memoryCommandStore{}})

	body := []byte(`{"service_id": "svc-1", "action_type": "restart", "reason": "test", "dry_run": true, "idempotency_key": "k-actor-prod"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/service-actions", bytes.NewReader(body))
	req.Header.Set("X-Devhub-Actor", "spoofed-actor")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 (ADR-0006 reject), got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "x_devhub_actor_removed") {
		t.Fatalf("expected body code=x_devhub_actor_removed, got %q", rec.Body.String())
	}
}

func TestXDevhubActorRejectedEvenWhenDevFallbackOn(t *testing.T) {
	// ADR-0006: dev fallback no longer matters — inbound X-Devhub-Actor
	// is rejected before the auth/role gate.
	commandStore := &memoryCommandStore{}
	router := testRouter(RouterConfig{CommandStore: commandStore})

	body := []byte(`{"service_id": "svc-1", "action_type": "restart", "reason": "test", "dry_run": true, "idempotency_key": "k-actor-dev"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/service-actions", bytes.NewReader(body))
	req.Header.Set("X-Devhub-Actor", "spoofed-actor")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 (ADR-0006 reject), got %d body=%s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("X-Devhub-Actor-Deprecated"); got != "" {
		t.Fatalf("X-Devhub-Actor-Deprecated must not be set, got %q", got)
	}
	if len(commandStore.commands) != 0 {
		t.Fatalf("command must not be persisted when X-Devhub-Actor is rejected, got %d", len(commandStore.commands))
	}
}

// ADR-0020 sub-carve B (sprint -i, issue #209) lazy auto-create — Keycloak
// admin console 또는 HRDB ETL 로 신규 생성된 user 의 첫 DevHub 진입 시
// `users` row 자동 생성. 이전 (sprint -ad 이전) ErrNotFound 시 token role
// claim 으로 fall-through 만 했던 동작이 본 sprint 부터 CreateUser 흐름으로 확장.

func TestAuthenticateActor_LazyAutoCreate_HappyPath(t *testing.T) {
	orgs := newMemoryOrganizationStore()
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login:       "newcomer",
		Subject:     "kc-uuid-newcomer",
		Role:        "developer",
		Email:       "newcomer@example.com",
		DisplayName: "Newcomer User",
	}}
	router := NewRouter(RouterConfig{
		OrganizationStore:   orgs,
		BearerTokenVerifier: verifier,
		AuditStore:          &memoryAuditStore{},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer t")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
	created, err := orgs.GetUser(context.Background(), "newcomer")
	if err != nil {
		t.Fatalf("lazy-created user not found: %v", err)
	}
	if created.Email != "newcomer@example.com" || created.DisplayName != "Newcomer User" ||
		created.Role != domain.AppRoleDeveloper || created.Status != domain.UserStatusActive {
		t.Errorf("lazy-created user = %+v, expected token claim values", created)
	}
	if created.IdPSubject != "kc-uuid-newcomer" {
		t.Errorf("IdPSubject = %q, want kc-uuid-newcomer (best-effort backfill)", created.IdPSubject)
	}
}

func TestAuthenticateActor_LazyAutoCreate_DefaultsRoleWhenTokenEmpty(t *testing.T) {
	orgs := newMemoryOrganizationStore()
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login:   "ghost-role",
		Subject: "kc-uuid-ghost",
		Role:    "", // empty — verifier picked nothing from realm_access.roles
	}}
	router := NewRouter(RouterConfig{
		OrganizationStore:   orgs,
		BearerTokenVerifier: verifier,
		AuditStore:          &memoryAuditStore{},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer t")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
	created, err := orgs.GetUser(context.Background(), "ghost-role")
	if err != nil {
		t.Fatalf("lazy-created user not found: %v", err)
	}
	if created.Role != domain.AppRoleDeveloper {
		t.Errorf("Role = %q, want developer (lazyAutoCreateDefaultRole, ADR-0020 §5.2.2 P1-3)", created.Role)
	}
}

func TestAuthenticateActor_LazyAutoCreate_PMOManagerAccepted(t *testing.T) {
	orgs := newMemoryOrganizationStore()
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login:   "pmo-newcomer",
		Subject: "kc-uuid-pmo",
		Role:    "pmo_manager", // migration 000021 seeded
	}}
	router := NewRouter(RouterConfig{
		OrganizationStore:   orgs,
		BearerTokenVerifier: verifier,
		AuditStore:          &memoryAuditStore{},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer t")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
	created, err := orgs.GetUser(context.Background(), "pmo-newcomer")
	if err != nil {
		t.Fatalf("lazy-created user not found: %v", err)
	}
	if string(created.Role) != "pmo_manager" {
		t.Errorf("Role = %q, want pmo_manager (token claim preserved when valid)", created.Role)
	}
}

func TestAuthenticateActor_LazyAutoCreate_FallbackDisplayNameWhenClaimsEmpty(t *testing.T) {
	orgs := newMemoryOrganizationStore()
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login:       "minimal",
		Subject:     "kc-uuid-minimal",
		Role:        "developer",
		Email:       "", // missing email claim
		DisplayName: "", // missing name claim
	}}
	router := NewRouter(RouterConfig{
		OrganizationStore:   orgs,
		BearerTokenVerifier: verifier,
		AuditStore:          &memoryAuditStore{},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer t")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
	created, err := orgs.GetUser(context.Background(), "minimal")
	if err != nil {
		t.Fatalf("lazy-created user not found: %v", err)
	}
	if created.DisplayName != "minimal" {
		t.Errorf("DisplayName = %q, want login fallback %q", created.DisplayName, "minimal")
	}
}

func TestAuthenticateActor_LazyAutoCreate_SkippedWhenStoreNil(t *testing.T) {
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login:   "no-store",
		Subject: "kc-uuid-no-store",
		Role:    "developer",
	}}
	// no OrganizationStore — lazy auto-create branch must skip without panic.
	router := NewRouter(RouterConfig{
		BearerTokenVerifier: verifier,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer t")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// /me handler returns 200 with token role even without store (auth path is
	// the focus of this test; the handler itself works with the bearer actor).
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
}
