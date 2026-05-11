package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/domain"
)

// staticActorAuth 는 모든 요청에 actor.Login=defaultActor 를 설정하는 미들웨어.
// authenticateActor 의 production 분기를 호출하면 BearerTokenVerifier 가 필요한데,
// 본 sprint 의 password handler 단위 테스트는 actor 추출만 검증한다.
type staticActorAuth struct {
	actorLogin string
}

func (s staticActorAuth) VerifyBearerToken(_ context.Context, _ string) (AuthenticatedActor, error) {
	return AuthenticatedActor{Login: s.actorLogin, Role: "developer"}, nil
}

func newAccountPasswordRouter(actorLogin string, orgStore OrganizationStore, kratos KratosLoginClient, audits *memoryAuditStore, cache *KratosSessionCache) http.Handler {
	return NewRouter(RouterConfig{
		KratosLogin:         kratos,
		OrganizationStore:   orgStore,
		AuditStore:          audits,
		BearerTokenVerifier: staticActorAuth{actorLogin: actorLogin},
		KratosSessionCache:  cache,
	})
}

func postAccountPassword(handler http.Handler, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/account/password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer fake")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func seedActorUser(t *testing.T, orgStore *memoryOrganizationStore, userID, email string) {
	t.Helper()
	if _, err := orgStore.CreateUser(context.Background(), domain.CreateUserInput{
		UserID:      userID,
		Email:       email,
		DisplayName: userID,
		Role:        domain.AppRoleDeveloper,
		Status:      domain.UserStatusActive,
		Type:        domain.UserTypeHuman,
		JoinedAt:    time.Now().UTC(),
	}); err != nil {
		t.Fatalf("seed user %s: %v", userID, err)
	}
}

// 1) Happy path: backend proxy 가 fresh login → settings flow → cache 갱신 + audit.
func TestAccountPassword_Success(t *testing.T) {
	orgStore := newMemoryOrganizationStore()
	seedActorUser(t, orgStore, "alice", "alice@example.com")

	kratos := &fakeKratosLogin{
		flow:           KratosLoginFlow{ID: "login-flow-1"},
		identity:       KratosIdentity{ID: "kratos-uuid-alice", UserID: "alice", SessionToken: "fresh-sess-1"},
		settingsFlowID: "settings-flow-1",
	}
	audits := &memoryAuditStore{}
	cache := NewKratosSessionCache()
	router := newAccountPasswordRouter("alice", orgStore, kratos, audits, cache)

	rec := postAccountPassword(router,
		`{"current_password":"old-pass-12!","new_password":"new-pass-12!"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if kratos.submitCalls != 1 || kratos.submitIdentifiers[0] != "alice@example.com" {
		t.Errorf("SubmitLogin called with %+v / %+v, want alice@example.com",
			kratos.submitIdentifiers, kratos.submitPasswords)
	}
	if len(kratos.settingsSubmits) != 1 || kratos.settingsSubmits[0].NewPassword != "new-pass-12!" {
		t.Errorf("settings submit unexpected: %+v", kratos.settingsSubmits)
	}
	if kratos.settingsSubmits[0].SessionToken != "fresh-sess-1" {
		t.Errorf("settings flow ran with stale session_token: %s", kratos.settingsSubmits[0].SessionToken)
	}
	if tok, ok := cache.Get("alice"); !ok || tok != "fresh-sess-1" {
		t.Errorf("cache.Get(alice) = (%q, %v), want (fresh-sess-1, true)", tok, ok)
	}
	hasAudit := false
	for _, a := range audits.logs {
		if a.Action == "account.password_self_change" {
			hasAudit = true
		}
	}
	if !hasAudit {
		t.Errorf("expected account.password_self_change audit, got %+v", audits.logs)
	}
}

// 2) current_password 누락 → 400.
func TestAccountPassword_RejectsMissingFields(t *testing.T) {
	orgStore := newMemoryOrganizationStore()
	seedActorUser(t, orgStore, "alice", "alice@example.com")
	router := newAccountPasswordRouter("alice", orgStore, &fakeKratosLogin{}, &memoryAuditStore{}, NewKratosSessionCache())

	rec := postAccountPassword(router, `{"new_password":"only"}`)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("missing current → status=%d, want 400", rec.Code)
	}
	rec = postAccountPassword(router, `{"current_password":"only"}`)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("missing new → status=%d, want 400", rec.Code)
	}
}

// 3) current_password == new_password → 400 validation.
func TestAccountPassword_RejectsSamePassword(t *testing.T) {
	orgStore := newMemoryOrganizationStore()
	seedActorUser(t, orgStore, "alice", "alice@example.com")
	router := newAccountPasswordRouter("alice", orgStore, &fakeKratosLogin{}, &memoryAuditStore{}, NewKratosSessionCache())

	rec := postAccountPassword(router, `{"current_password":"same-pass","new_password":"same-pass"}`)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status=%d, want 400", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "validation") {
		t.Errorf("error code missing: %s", rec.Body.String())
	}
}

// 4) DevHub users 에 actor 없음 → 401 reauth.
func TestAccountPassword_UnknownDevhubUser(t *testing.T) {
	orgStore := newMemoryOrganizationStore() // no users seeded
	audits := &memoryAuditStore{}
	router := newAccountPasswordRouter("ghost", orgStore, &fakeKratosLogin{}, audits, NewKratosSessionCache())

	rec := postAccountPassword(router, `{"current_password":"a","new_password":"b"}`)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status=%d, want 401 body=%s", rec.Code, rec.Body.String())
	}
	hasAudit := false
	for _, a := range audits.logs {
		if a.Action == "account.password_self_change.no_user" {
			hasAudit = true
		}
	}
	if !hasAudit {
		t.Errorf("expected no_user audit, got %+v", audits.logs)
	}
}

// 5) Kratos 가 current_password 거절 → 401 current_password_invalid.
func TestAccountPassword_InvalidCurrent(t *testing.T) {
	orgStore := newMemoryOrganizationStore()
	seedActorUser(t, orgStore, "alice", "alice@example.com")
	kratos := &fakeKratosLogin{
		flow:      KratosLoginFlow{ID: "f"},
		submitErr: ErrKratosInvalidCredentials,
	}
	audits := &memoryAuditStore{}
	router := newAccountPasswordRouter("alice", orgStore, kratos, audits, NewKratosSessionCache())

	rec := postAccountPassword(router, `{"current_password":"wrong","new_password":"new-pass-12!"}`)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d, want 401 body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "current_password_invalid") {
		t.Errorf("body should expose current_password_invalid code: %s", rec.Body.String())
	}
	if len(kratos.settingsSubmits) != 0 {
		t.Errorf("settings flow must not run when current_password is rejected")
	}
	hasAudit := false
	for _, a := range audits.logs {
		if a.Action == "account.password_self_change.invalid_current" {
			hasAudit = true
		}
	}
	if !hasAudit {
		t.Errorf("expected invalid_current audit, got %+v", audits.logs)
	}
}

// 6) Settings flow 가 privileged required → 401 reauth_required.
func TestAccountPassword_PrivilegedRequiredFromSettingsFlow(t *testing.T) {
	orgStore := newMemoryOrganizationStore()
	seedActorUser(t, orgStore, "alice", "alice@example.com")
	kratos := &fakeKratosLogin{
		flow:            KratosLoginFlow{ID: "f"},
		identity:        KratosIdentity{ID: "id", UserID: "alice", SessionToken: "sess"},
		settingsFlowErr: &KratosSettingsError{Code: KratosSettingsPrivilegedRequired},
	}
	router := newAccountPasswordRouter("alice", orgStore, kratos, &memoryAuditStore{}, NewKratosSessionCache())

	rec := postAccountPassword(router, `{"current_password":"old","new_password":"new-pass-12!"}`)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d, want 401 body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "reauth_required") {
		t.Errorf("body should expose reauth_required: %s", rec.Body.String())
	}
}

// 7) Validation from Kratos (weak new password) → 400 validation.
func TestAccountPassword_ValidationFromSettingsSubmit(t *testing.T) {
	orgStore := newMemoryOrganizationStore()
	seedActorUser(t, orgStore, "alice", "alice@example.com")
	kratos := &fakeKratosLogin{
		flow:              KratosLoginFlow{ID: "f"},
		identity:          KratosIdentity{ID: "id", UserID: "alice", SessionToken: "sess"},
		settingsFlowID:    "settings-1",
		settingsSubmitErr: &KratosSettingsError{Code: KratosSettingsValidation, Message: "password too short"},
	}
	router := newAccountPasswordRouter("alice", orgStore, kratos, &memoryAuditStore{}, NewKratosSessionCache())

	rec := postAccountPassword(router, `{"current_password":"old","new_password":"shrt"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d, want 400 body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "password too short") {
		t.Errorf("validation message missing: %s", rec.Body.String())
	}
}

// 8) KratosLogin 미주입 → 503.
func TestAccountPassword_Unavailable(t *testing.T) {
	orgStore := newMemoryOrganizationStore()
	seedActorUser(t, orgStore, "alice", "alice@example.com")
	router := NewRouter(RouterConfig{
		OrganizationStore:   orgStore,
		BearerTokenVerifier: staticActorAuth{actorLogin: "alice"},
		// KratosLogin intentionally nil
	})
	rec := postAccountPassword(router, `{"current_password":"a","new_password":"b"}`)
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status=%d, want 503", rec.Code)
	}
}

// 9) Empty session_token from Kratos → 500 (handler 가 invariant 깨지면 마스킹).
func TestAccountPassword_EmptySessionTokenSurface500(t *testing.T) {
	orgStore := newMemoryOrganizationStore()
	seedActorUser(t, orgStore, "alice", "alice@example.com")
	kratos := &fakeKratosLogin{
		flow:     KratosLoginFlow{ID: "f"},
		identity: KratosIdentity{ID: "id", UserID: "alice", SessionToken: ""}, // empty
	}
	router := newAccountPasswordRouter("alice", orgStore, kratos, &memoryAuditStore{}, NewKratosSessionCache())

	rec := postAccountPassword(router, `{"current_password":"old","new_password":"new-pass-12!"}`)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status=%d, want 500 body=%s", rec.Code, rec.Body.String())
	}
}
