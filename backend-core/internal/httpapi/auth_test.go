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
	authview "github.com/devhub/backend-core/internal/domain/auth-session/view"
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
		Role:    "team_manager", // token claim says manager
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
	if !strings.Contains(rec.Body.String(), `"role":"team_manager"`) {
		t.Errorf("expected role to fall back to token claim 'team_manager', body = %s", rec.Body.String())
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
		Role:        domain.AppRoleTeamManager,
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
		Role:    "team_manager",
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

// ADR-0021 §3.3 (2026-05-21 lazy 폐기 sprint, issue #284): authenticateActor 가
// GetUser ErrNotFound 시 token-only actor 로 처리 (`devhub_onboarding_required`
// flag 만 set, users row 자동 생성 없음). 이전 (sprint -i ~ -ad) 의 lazy
// auto-create 흐름 (ADR-0020 sub-carve B) 은 폐기됨.

func TestAuthenticateActor_TokenOnlyActor_NoLazyCreate(t *testing.T) {
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
	// users row 자동 생성 금지 — DB miss 가 token-only actor 로 끝나야 함.
	if _, err := orgs.GetUser(context.Background(), "newcomer"); err == nil {
		t.Fatalf("users row 가 자동 생성되었음 — lazy 폐기 (issue #284) 위반")
	}
}

func TestAuthenticateActor_TokenOnlyActor_WorksWithoutStore(t *testing.T) {
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login:   "no-store",
		Subject: "kc-uuid-no-store",
		Role:    "developer",
	}}
	// no OrganizationStore — auth 가 panic 없이 token role 으로 진행해야 함.
	router := NewRouter(RouterConfig{
		BearerTokenVerifier: verifier,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer t")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
}

func TestAuthenticateActor_RoleDriftFailsClosedOnProtectedRoute(t *testing.T) {
	orgs := newMemoryOrganizationStore()
	if _, err := orgs.CreateUser(context.Background(), domain.CreateUserInput{
		UserID:      "bob",
		Email:       "bob@example.com",
		DisplayName: "Bob",
		Role:        domain.AppRoleDeveloper,
		Status:      domain.UserStatusActive,
		Type:        domain.UserTypeHuman,
	}); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	audits := &memoryAuditStore{}
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login:   "bob",
		Subject: "user-bob",
		Role:    "team_manager",
	}}
	router := NewRouter(RouterConfig{
		OrganizationStore:   orgs,
		AuditStore:          audits,
		BearerTokenVerifier: verifier,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	req.Header.Set("Authorization", "Bearer t")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "auth.role_sync_required") {
		t.Fatalf("expected auth.role_sync_required, body=%s", rec.Body.String())
	}
	if len(audits.logs) != 1 || audits.logs[0].Action != "auth.role_sync_required" {
		t.Fatalf("expected auth.role_sync_required audit, got %+v", audits.logs)
	}
}

func TestAuthenticateActor_GenericKeycloakRoleDoesNotTriggerDriftFailClosed(t *testing.T) {
	for _, tokenRole := range []string{"user", "default-roles-devhub"} {
		t.Run(tokenRole, func(t *testing.T) {
			orgs := newMemoryOrganizationStore()
			if _, err := orgs.CreateUser(context.Background(), domain.CreateUserInput{
				UserID:      "charlie",
				Email:       "charlie@example.com",
				DisplayName: "Charlie",
				Role:        domain.AppRoleSystemAdmin,
				Status:      domain.UserStatusActive,
				Type:        domain.UserTypeHuman,
			}); err != nil {
				t.Fatalf("seed user: %v", err)
			}
			audits := &memoryAuditStore{}
			verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
				Login:   "charlie",
				Subject: "user-charlie",
				Role:    tokenRole,
			}}
			router := NewRouter(RouterConfig{
				OrganizationStore:   orgs,
				AuditStore:          audits,
				BearerTokenVerifier: verifier,
			})

			req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
			req.Header.Set("Authorization", "Bearer t")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
			}
			for _, log := range audits.logs {
				if log.Action == "auth.role_sync_required" {
					t.Fatalf("did not expect auth.role_sync_required audit for generic token role %q: %+v", tokenRole, audits.logs)
				}
			}
		})
	}
}

func TestLogoutEndpoint_ClearsCookiesAndWritesAudit(t *testing.T) {
	audits := &memoryAuditStore{}
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login:   "alice",
		Subject: "user-alice",
		Role:    "developer",
	}}
	oidc := &fakeOIDCLogoutClient{}
	router := NewRouter(RouterConfig{
		OrganizationStore:   newMemoryOrganizationStore(),
		AuditStore:          audits,
		BearerTokenVerifier: verifier,
		OIDCLogoutClient:    oidc,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", strings.NewReader(`{"refresh_token":"rt-1","id_token":"id-1"}`))
	req.Header.Set("Authorization", "Bearer t")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// TC-AUTH-LOGOUT-01 — 204 No Content (spec) + cookies cleared + audit.
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d body=%s", rec.Code, rec.Body.String())
	}
	cookieHeader := strings.Join(rec.Header().Values("Set-Cookie"), "\n")
	for _, name := range []string{"devhub_session", "devhub_access_token", "devhub_refresh_token", "devhub_id_token"} {
		if !strings.Contains(cookieHeader, name+"=") {
			t.Fatalf("expected cookie %q to be cleared, headers=%q", name, cookieHeader)
		}
	}
	if len(audits.logs) != 1 || audits.logs[0].Action != "auth.logout" {
		t.Fatalf("expected auth.logout audit, got %+v", audits.logs)
	}
	// OIDC logout 호출 검증 (revoke_status=ok)
	if len(oidc.calls) != 1 || oidc.calls[0] != "rt-1" {
		t.Fatalf("expected OIDCLogout(rt-1), got %v", oidc.calls)
	}
	revokeStatus, _ := audits.logs[0].Payload["revoke_status"].(string)
	if revokeStatus != "ok" {
		t.Fatalf("expected revoke_status=ok, got %v", revokeStatus)
	}
}

func TestLogoutEndpoint_RevokesKeycloakSessions(t *testing.T) {
	audits := &memoryAuditStore{}
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login:   "alice",
		Subject: "user-alice",
		Role:    "developer",
	}}
	oidc := &fakeOIDCLogoutClient{}
	router := NewRouter(RouterConfig{
		OrganizationStore:   newMemoryOrganizationStore(),
		AuditStore:          audits,
		BearerTokenVerifier: verifier,
		OIDCLogoutClient:    oidc,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", strings.NewReader(`{"refresh_token":"rt-1","id_token":"id-1"}`))
	req.Header.Set("Authorization", "Bearer t")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// TC-AUTH-LOGOUT-02 — OIDC logout 호출 + 204 + audit revoke_status=ok.
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d body=%s", rec.Code, rec.Body.String())
	}
	if len(audits.logs) != 1 || audits.logs[0].Action != "auth.logout" {
		t.Fatalf("expected auth.logout audit, got %+v", audits.logs)
	}
	if len(oidc.calls) != 1 || oidc.calls[0] != "rt-1" {
		t.Fatalf("expected OIDCLogout(rt-1), got %v", oidc.calls)
	}
	revokeStatus, _ := audits.logs[0].Payload["revoke_status"].(string)
	if revokeStatus != "ok" {
		t.Fatalf("expected revoke_status=ok, got %v", revokeStatus)
	}
}

// TC-AUTH-LOGOUT-03 — idempotency: 같은 refresh_token 두번 호출 → 두번 다
// 204. OIDCLogout 가 4xx 를 nil 로 정규화하므로 second call 도 정상.
func TestLogoutEndpoint_Idempotent(t *testing.T) {
	audits := &memoryAuditStore{}
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login:   "alice",
		Subject: "user-alice",
		Role:    "developer",
	}}
	oidc := &fakeOIDCLogoutClient{}
	router := NewRouter(RouterConfig{
		OrganizationStore:   newMemoryOrganizationStore(),
		AuditStore:          audits,
		BearerTokenVerifier: verifier,
		OIDCLogoutClient:    oidc,
	})

	body := `{"refresh_token":"rt-1"}`
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer t")
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("call %d: expected 204, got %d body=%s", i+1, rec.Code, rec.Body.String())
		}
	}
	// OIDCLogout 가 두번 호출되어야 함 (idempotency 는 핸들러 레벨 — Keycloak
	// 자체가 idempotent).
	if len(oidc.calls) != 2 {
		t.Errorf("expected 2 OIDCLogout calls, got %d", len(oidc.calls))
	}
	if len(audits.logs) != 2 {
		t.Errorf("expected 2 audit rows, got %d", len(audits.logs))
	}
}

// TC-AUTH-LOGOUT-04 — Keycloak unreachable → **204 No Content + response
// header `X-Keycloak-Likely-Down: true`** (N-8 hotfix 4차, issue #501,
// 2026-06-09). spec #488 의 "정합 우선" 분기 (502) 가 graceful degradation
// 으로 변경 — frontend logout() 가 정상 204 분기 진입 + response header 의
// `X-Keycloak-Likely-Down` 마커 확인 → OIDC end_session_endpoint 호출 skip
// + 강제 /login (IdP outage 시 dead IdP trap 회피). 204 No Content 정합
// (HTTP spec) — body 없이 header 마커 사용. OIDC 정상 시 (revoke_status=ok
// 분기) 본 분기 진입 안 함.
func TestLogoutEndpoint_KeycloakUnreachable_204(t *testing.T) {
	audits := &memoryAuditStore{}
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login:   "alice",
		Subject: "user-alice",
		Role:    "developer",
	}}
	oidc := &fakeOIDCLogoutClient{err: errors.New("keycloak timeout")}
	router := NewRouter(RouterConfig{
		OrganizationStore:   newMemoryOrganizationStore(),
		AuditStore:          audits,
		BearerTokenVerifier: verifier,
		OIDCLogoutClient:    oidc,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", strings.NewReader(`{"refresh_token":"rt-1"}`))
	req.Header.Set("Authorization", "Bearer t")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 (N-8 hotfix 4차 graceful degradation), got %d body=%s", rec.Code, rec.Body.String())
	}
	// 204 No Content 정합 — body 비어 있어야 함.
	if rec.Body.Len() != 0 {
		t.Errorf("expected empty body for 204, got %s", rec.Body.String())
	}
	// response header 마커 — frontend 가 OIDC skip 결정용.
	if got := rec.Header().Get("X-Keycloak-Likely-Down"); got != "true" {
		t.Errorf("expected X-Keycloak-Likely-Down=true, got %q", got)
	}
	if got := rec.Header().Get("X-Logout-Hotfix"); got != "N-8-4:graceful-degrade" {
		t.Errorf("expected X-Logout-Hotfix=N-8-4:graceful-degrade, got %q", got)
	}
	// audit 는 unreachable 상태로 emit 되어야 함 (handler 가 204 반환 전 audit).
	if len(audits.logs) != 1 {
		t.Fatalf("expected 1 audit row (unreachable), got %d", len(audits.logs))
	}
	revokeStatus, _ := audits.logs[0].Payload["revoke_status"].(string)
	if revokeStatus != "unreachable" {
		t.Errorf("expected revoke_status=unreachable, got %v", revokeStatus)
	}
	if hotfix, _ := audits.logs[0].Payload["hotfix"].(string); hotfix != "N-8-4:graceful-degrade" {
		t.Errorf("expected hotfix=N-8-4:graceful-degrade, got %v", hotfix)
	}
}

// TC-AUTH-LOGOUT-08 — backend config error (e.g. missing OIDC client
// credentials, devhub-frontend public client 시 DEVHUB_OIDC_CLIENT_SECRET
// 없음) → **204 No Content + marker 미부착** (N-8 hotfix 4차 codex P1
// follow-up, 2026-06-09). sentinel error `authview.ErrOIDCConfigMissing` 로
// wrap 되어야 handler 가 config_error 분기 진입. frontend logout() 가
// marker 없음 → 정상 OIDC 분기 → RP-initiated logout 시도 → Keycloak
// SSO session 정상 종료. codex P1 의 "reachable Keycloak SSO session is
// not terminated" 문제 회피.
func TestLogoutEndpoint_OIDCConfigError_204_NoMarker(t *testing.T) {
	audits := &memoryAuditStore{}
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login:   "alice",
		Subject: "user-alice",
		Role:    "developer",
	}}
	// sentinel error wrap — handler 가 config_error 분기 진입 결정.
	wrappedErr := fmt.Errorf("%w: KeycloakAdminClient requires realm, oidc_client_id, oidc_client_secret for OIDC logout", authview.ErrOIDCConfigMissing)
	oidc := &fakeOIDCLogoutClient{err: wrappedErr}
	router := NewRouter(RouterConfig{
		OrganizationStore:   newMemoryOrganizationStore(),
		AuditStore:          audits,
		BearerTokenVerifier: verifier,
		OIDCLogoutClient:    oidc,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", strings.NewReader(`{"refresh_token":"rt-1"}`))
	req.Header.Set("Authorization", "Bearer t")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 (config error 도 정상 응답), got %d body=%s", rec.Code, rec.Body.String())
	}
	if rec.Body.Len() != 0 {
		t.Errorf("expected empty body for 204, got %s", rec.Body.String())
	}
	// config error 분기 — marker **미부착**. frontend 가 정상 OIDC 분기
	// (RP-initiated logout) 결정. codex P1 follow-up 핵심.
	if got := rec.Header().Get("X-Keycloak-Likely-Down"); got != "" {
		t.Errorf("expected X-Keycloak-Likely-Down unset for config error, got %q", got)
	}
	if got := rec.Header().Get("X-Logout-Hotfix"); got != "" {
		t.Errorf("expected X-Logout-Hotfix unset for config error, got %q", got)
	}
	// audit 는 config_error 상태로 emit.
	if len(audits.logs) != 1 {
		t.Fatalf("expected 1 audit row (config_error), got %d", len(audits.logs))
	}
	revokeStatus, _ := audits.logs[0].Payload["revoke_status"].(string)
	if revokeStatus != "config_error" {
		t.Errorf("expected revoke_status=config_error, got %v", revokeStatus)
	}
	if hotfix, _ := audits.logs[0].Payload["hotfix"].(string); hotfix != "N-8-4:graceful-degrade" {
		t.Errorf("expected hotfix=N-8-4:graceful-degrade, got %v", hotfix)
	}
	// config_error_detail (codex P1 follow-up: config error 의 detail 정보
	// audit trace — Keycloak admin 가 디버깅용).
	if detail, _ := audits.logs[0].Payload["config_error_detail"].(string); detail == "" {
		t.Errorf("expected config_error_detail to be set, got empty")
	}
}

// TC-AUTH-LOGOUT-05 — refresh_token 없이 호출 → 204 + audit
// revoke_status=skipped_no_refresh_token. OIDC logout skip, cookies
// clear + audit 만 emit.
func TestLogoutEndpoint_NoRefreshToken_204(t *testing.T) {
	audits := &memoryAuditStore{}
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login:   "alice",
		Subject: "user-alice",
		Role:    "developer",
	}}
	oidc := &fakeOIDCLogoutClient{}
	router := NewRouter(RouterConfig{
		OrganizationStore:   newMemoryOrganizationStore(),
		AuditStore:          audits,
		BearerTokenVerifier: verifier,
		OIDCLogoutClient:    oidc,
	})

	// empty body — refresh_token 없이
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer t")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d body=%s", rec.Code, rec.Body.String())
	}
	// OIDCLogout 호출 안 됨
	if len(oidc.calls) != 0 {
		t.Errorf("expected 0 OIDCLogout calls (no refresh_token), got %d", len(oidc.calls))
	}
	// audit revoke_status=skipped_no_refresh_token
	if len(audits.logs) != 1 {
		t.Fatalf("expected 1 audit row, got %d", len(audits.logs))
	}
	revokeStatus, _ := audits.logs[0].Payload["revoke_status"].(string)
	if revokeStatus != "skipped_no_refresh_token" {
		t.Errorf("expected revoke_status=skipped_no_refresh_token, got %v", revokeStatus)
	}
}

// fakeOIDCLogoutClient — OIDCLogoutClient 인터페이스 구현. 테스트 fixture.
type fakeOIDCLogoutClient struct {
	err   error
	calls []string // 호출된 refresh_token 들
}

func (f *fakeOIDCLogoutClient) OIDCLogout(_ context.Context, refreshToken string) error {
	f.calls = append(f.calls, refreshToken)
	return f.err
}

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
		ClientID:         "devhub-backend-admin",                  // admin client
		ClientSecret:     "admin-secret",
		OIDCClientID:     "devhub-frontend",                       // OIDC client (frontend)
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
		AdminURL:         adminSrv.URL,     // admin (internal)
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

// ADR-0029 — API key 인증 테스트. public API 호출 시 Keycloak JWT 대신
// Authorization: Bearer <DEVHUB_API_KEY> 로 인증. Keycloak 도달 불필요.
func TestAPIKeyAuthentication_ValidKeyPassesThrough(t *testing.T) {
	const key = "test-static-api-key-2026-06-09"
	router := NewRouter(RouterConfig{
		APIKey:            key,
		OrganizationStore: newMemoryOrganizationStore(),
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/metrics", nil)
	req.Header.Set("Authorization", "Bearer "+key)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/dashboard/metrics with valid API key: got %d, want 200. body=%s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("X-Devhub-Auth") != "api_key" {
		t.Errorf("expected X-Devhub-Auth=api_key, got %q", rec.Header().Get("X-Devhub-Auth"))
	}
}

// API key 가 잘못되면 401. 단 timing attack 회피용 constant-time 비교.
func TestAPIKeyAuthentication_InvalidKeyReturns401(t *testing.T) {
	router := NewRouter(RouterConfig{
		APIKey:            "correct-key",
		OrganizationStore: newMemoryOrganizationStore(),
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/metrics", nil)
	req.Header.Set("Authorization", "Bearer wrong-key")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("invalid API key: got %d, want 401", rec.Code)
	}
}

// JWT 형식 토큰 + APIKey 설정 시 Keycloak verifier 가 우선 호출되어야 함
// (admin endpoints 등). 미설정 verifier 라면 verifier-not-configured 응답.
func TestAPIKeyAuthentication_JWTFormatGoesToKeycloakVerifier(t *testing.T) {
	verifier := &fakeBearerTokenVerifier{err: errors.New("not a real jwt")}
	router := NewRouter(RouterConfig{
		APIKey:              "static-key",
		BearerTokenVerifier: verifier,
		OrganizationStore:   newMemoryOrganizationStore(),
	})

	jwt := "header.payload.signature"
	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/metrics", nil)
	req.Header.Set("Authorization", "Bearer "+jwt)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if verifier.token != jwt {
		t.Errorf("expected verifier to receive JWT %q, got %q", jwt, verifier.token)
	}
	if rec.Header().Get("X-Devhub-Auth") == "api_key" {
		t.Errorf("JWT token must not be classified as api_key path")
	}
}

// APIKey 가 비어있으면 Bearer 의 비-JWT 가 그대로 401 로 떨어져야 함
// (기존 Keycloak-only 동작 회귀 가드).
func TestAPIKeyAuthentication_EmptyKeyDoesNotActivate(t *testing.T) {
	router := NewRouter(RouterConfig{
		APIKey:            "",
		OrganizationStore: newMemoryOrganizationStore(),
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/metrics", nil)
	req.Header.Set("Authorization", "Bearer some-static-string")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("non-JWT bearer with empty API key: got %d, want 401 (regression: must not fall through to api_key branch)", rec.Code)
	}
}

// looksLikeJWT 단위 테스트.
func TestLooksLikeJWT(t *testing.T) {
	// looksLikeJWT 는 view 패키지 내부 함수이므로, 본 테스트는 view 패키지
	// 자체의 동작 (auth.go 의 API key 분기 결과) 으로 검증한다. 비-JWT 입력
	// 이 API key 분기에서 401 로 떨어지지 않고 (X-Devhub-Auth 미부착) 그대로
	// verifier 호출로 이어지는지만 확인.
	verifier := &fakeBearerTokenVerifier{err: errors.New("not configured")}
	router := NewRouter(RouterConfig{
		APIKey:              "static-key",
		BearerTokenVerifier: verifier,
		OrganizationStore:   newMemoryOrganizationStore(),
	})

	// 비-JWT + invalid key → api_key 분기에서 deny (verifier 호출 X).
	req1 := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/metrics", nil)
	req1.Header.Set("Authorization", "Bearer only.two")
	rec1 := httptest.NewRecorder()
	router.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusUnauthorized {
		t.Errorf("looksLikeJWT path: 2-segment token should be treated as non-JWT, got %d", rec1.Code)
	}
	if verifier.token != "" {
		t.Errorf("non-JWT token should not reach Keycloak verifier; got token=%q", verifier.token)
	}
}

// ADR-0029 §6 (g) P2 — API key 인증 성공 시 `auth.api_key_authenticated` audit + payload enrichment 검증.
func TestAPIKeyAuthentication_SuccessAuditEnriched(t *testing.T) {
	const key = "test-api-key-audit-success-2026-06-09"
	audits := &memoryAuditStore{}
	router := NewRouter(RouterConfig{
		APIKey:            key,
		OrganizationStore: newMemoryOrganizationStore(),
		AuditStore:        audits,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/metrics", nil)
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("X-Forwarded-For", "203.0.113.42")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body=%s)", rec.Code, rec.Body.String())
	}

	// audit row 검증.
	var found *domain.AuditLog
	for i, log := range audits.logs {
		if log.Action == "auth.api_key_authenticated" {
			found = &audits.logs[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("expected auth.api_key_authenticated audit row, got logs=%+v", audits.logs)
	}
	if found.TargetType != "auth" || found.TargetID != "api-key" {
		t.Errorf("audit target mismatch: got targetType=%q targetID=%q, want auth/api-key", found.TargetType, found.TargetID)
	}
	if found.Payload["actor_role"] != "system_admin" {
		t.Errorf("payload.actor_role = %v, want system_admin", found.Payload["actor_role"])
	}
	if found.Payload["path"] != "/api/v1/dashboard/metrics" {
		t.Errorf("payload.path = %v, want /api/v1/dashboard/metrics", found.Payload["path"])
	}
	if found.Payload["method"] != http.MethodGet {
		t.Errorf("payload.method = %v, want GET", found.Payload["method"])
	}
	if found.Payload["client_ip"] == "" || found.Payload["client_ip"] == nil {
		t.Errorf("payload.client_ip = %v, want non-empty (X-Forwarded-For honored)", found.Payload["client_ip"])
	}
	if found.Payload["request_id"] == "" || found.Payload["request_id"] == nil {
		t.Errorf("payload.request_id = %v, want non-empty", found.Payload["request_id"])
	}
}

// ADR-0029 §6 (g) P2 — API key 인증 거부 (invalid key) 시 `auth.api_key_denied` audit + reason=envelope 검증.
func TestAPIKeyAuthentication_InvalidKeyAuditEnriched(t *testing.T) {
	audits := &memoryAuditStore{}
	router := NewRouter(RouterConfig{
		APIKey:            "correct-key",
		OrganizationStore: newMemoryOrganizationStore(),
		AuditStore:        audits,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/metrics", nil)
	req.Header.Set("Authorization", "Bearer wrong-key")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}

	var found *domain.AuditLog
	for i, log := range audits.logs {
		if log.Action == "auth.api_key_denied" {
			found = &audits.logs[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("expected auth.api_key_denied audit row, got logs=%+v", audits.logs)
	}
	if found.Payload["reason"] != "invalid_key" {
		t.Errorf("payload.reason = %v, want invalid_key", found.Payload["reason"])
	}
	if found.Payload["path"] != "/api/v1/dashboard/metrics" {
		t.Errorf("payload.path = %v, want /api/v1/dashboard/metrics", found.Payload["path"])
	}
}
