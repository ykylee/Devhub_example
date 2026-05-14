package httpapi

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeKratosLogin struct {
	flow              KratosLoginFlow
	flowErr           error
	identity          KratosIdentity
	submitErr         error
	submitCalled      bool
	submitCalls       int
	submitIdentifiers []string
	submitPasswords   []string
	// Settings flow knobs (L4-C/L4-D).
	settingsFlowID     string
	settingsFlowErr    error
	settingsSubmitErr  error
	settingsFlowCalls  int
	settingsSubmits    []fakeSettingsSubmit
}

type fakeSettingsSubmit struct {
	SessionToken string
	FlowID       string
	NewPassword  string
}

func (f *fakeKratosLogin) CreateLoginFlow(ctx context.Context) (KratosLoginFlow, error) {
	if f.flowErr != nil {
		return KratosLoginFlow{}, f.flowErr
	}
	return f.flow, nil
}

func (f *fakeKratosLogin) SubmitLogin(ctx context.Context, flow KratosLoginFlow, identifier, password string) (KratosIdentity, error) {
	f.submitCalled = true
	f.submitCalls++
	f.submitIdentifiers = append(f.submitIdentifiers, identifier)
	f.submitPasswords = append(f.submitPasswords, password)
	if f.submitErr != nil {
		return KratosIdentity{}, f.submitErr
	}
	return f.identity, nil
}

func (f *fakeKratosLogin) CreateSettingsFlow(_ context.Context, _ string) (string, error) {
	f.settingsFlowCalls++
	if f.settingsFlowErr != nil {
		return "", f.settingsFlowErr
	}
	if f.settingsFlowID == "" {
		return "flow-default", nil
	}
	return f.settingsFlowID, nil
}

func (f *fakeKratosLogin) SubmitSettingsPassword(_ context.Context, sessionToken, flowID, newPassword string) error {
	f.settingsSubmits = append(f.settingsSubmits, fakeSettingsSubmit{
		SessionToken: sessionToken,
		FlowID:       flowID,
		NewPassword:  newPassword,
	})
	return f.settingsSubmitErr
}

type fakeHydraAdmin struct {
	loginRequest HydraLoginRequest
	getErr       error
	redirectTo   string
	acceptErr    error
	accepted          []string // subjects that AcceptLoginRequest was called with
	acceptedChallenge []string // challenge passed to AcceptLoginRequest
}

func (f *fakeHydraAdmin) GetLoginRequest(ctx context.Context, challenge string) (HydraLoginRequest, error) {
	if f.getErr != nil {
		return HydraLoginRequest{}, f.getErr
	}
	return f.loginRequest, nil
}

func (f *fakeHydraAdmin) AcceptLoginRequest(ctx context.Context, challenge, subject string, remember bool, rememberFor int) (string, error) {
	if f.acceptErr != nil {
		return "", f.acceptErr
	}
	f.accepted = append(f.accepted, subject)
	f.acceptedChallenge = append(f.acceptedChallenge, challenge)
	return f.redirectTo, nil
}

func (f *fakeHydraAdmin) GetConsentRequest(ctx context.Context, challenge string) (HydraConsentRequest, error) {
	return HydraConsentRequest{}, nil
}

func (f *fakeHydraAdmin) AcceptConsentRequest(ctx context.Context, challenge string, grantedScope []string, remember bool, rememberFor int) (string, error) {
	return "", nil
}

func newAuthLoginRouter(kratos KratosLoginClient, hydra HydraLoginAdmin, audits *memoryAuditStore) http.Handler {
	cfg := RouterConfig{
		KratosLogin: kratos,
		HydraAdmin:  hydra,
		AuditStore:  audits,
	}
	return NewRouter(cfg)
}

func postLogin(handler http.Handler, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestAuthLogin_PasswordFlowSuccess(t *testing.T) {
	audits := &memoryAuditStore{}
	kratos := &fakeKratosLogin{
		flow:     KratosLoginFlow{ID: "flow-1", CSRFToken: "csrf"},
		identity: KratosIdentity{ID: "kratos-uuid", UserID: "u1", Email: "u1@example.com"},
	}
	hydra := &fakeHydraAdmin{
		loginRequest: HydraLoginRequest{Challenge: "c1", Skip: false},
		redirectTo:   "http://hydra/oauth2/auth?code=...",
	}
	router := newAuthLoginRouter(kratos, hydra, audits)

	rec := postLogin(router, `{"login_challenge":"c1","identifier":"u1@example.com","password":"pw"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}

	if !kratos.submitCalled {
		t.Errorf("Kratos submit was not called for non-skip flow")
	}
	if len(hydra.accepted) != 1 || hydra.accepted[0] != "u1" {
		t.Errorf("Hydra subject = %v, want [u1] (DevHub user_id from metadata_public)", hydra.accepted)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("redirect_to")) {
		t.Errorf("response body missing redirect_to: %s", rec.Body.String())
	}
	if len(audits.logs) != 1 || audits.logs[0].Action != "auth.login.succeeded" {
		t.Errorf("expected auth.login.succeeded audit, got %+v", audits.logs)
	}
}

func TestAuthLogin_SkipFastPath(t *testing.T) {
	audits := &memoryAuditStore{}
	kratos := &fakeKratosLogin{}
	hydra := &fakeHydraAdmin{
		loginRequest: HydraLoginRequest{Challenge: "c1", Skip: true, Subject: "u-cached"},
		redirectTo:   "http://hydra/cb?code=...",
	}
	router := newAuthLoginRouter(kratos, hydra, audits)

	rec := postLogin(router, `{"login_challenge":"c1"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}

	if kratos.submitCalled {
		t.Errorf("Kratos submit should be skipped when Hydra reports skip=true")
	}
	if len(hydra.accepted) != 1 || hydra.accepted[0] != "u-cached" {
		t.Errorf("Hydra accept subject = %v, want [u-cached]", hydra.accepted)
	}
}

func TestAuthLogin_SkipWithCredentials_UsesPasswordFlow(t *testing.T) {
	audits := &memoryAuditStore{}
	kratos := &fakeKratosLogin{
		flow:     KratosLoginFlow{ID: "flow-1", CSRFToken: "csrf"},
		identity: KratosIdentity{ID: "kratos-uuid", UserID: "charlie", Email: "charlie@example.com"},
	}
	hydra := &fakeHydraAdmin{
		loginRequest: HydraLoginRequest{Challenge: "c1", Skip: true, Subject: "u-cached"},
		redirectTo:   "http://hydra/cb?code=...",
	}
	router := newAuthLoginRouter(kratos, hydra, audits)

	rec := postLogin(router, `{"login_challenge":"c1","identifier":"charlie","password":"pw"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !kratos.submitCalled {
		t.Fatalf("expected Kratos submit when credentials are explicitly supplied")
	}
	if len(hydra.accepted) != 1 || hydra.accepted[0] != "charlie" {
		t.Fatalf("Hydra accept subject = %v, want [charlie]", hydra.accepted)
	}
}

func TestAuthLogin_UsesCanonicalHydraChallengeOnAccept(t *testing.T) {
	audits := &memoryAuditStore{}
	kratos := &fakeKratosLogin{
		flow:     KratosLoginFlow{ID: "flow-1", CSRFToken: "csrf"},
		identity: KratosIdentity{ID: "kratos-uuid", UserID: "u1", Email: "u1@example.com"},
	}
	hydra := &fakeHydraAdmin{
		loginRequest: HydraLoginRequest{Challenge: "canonical-c1", Skip: false},
		redirectTo:   "http://hydra/oauth2/auth?code=...",
	}
	router := newAuthLoginRouter(kratos, hydra, audits)

	rec := postLogin(router, `{"login_challenge":"opaque-from-query","identifier":"u1@example.com","password":"pw"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if len(hydra.acceptedChallenge) != 1 || hydra.acceptedChallenge[0] != "canonical-c1" {
		t.Fatalf("accept challenge = %v, want [canonical-c1]", hydra.acceptedChallenge)
	}
}

func TestAuthLogin_InvalidCredentials(t *testing.T) {
	audits := &memoryAuditStore{}
	kratos := &fakeKratosLogin{
		flow:      KratosLoginFlow{ID: "flow-1"},
		submitErr: ErrKratosInvalidCredentials,
	}
	hydra := &fakeHydraAdmin{loginRequest: HydraLoginRequest{Challenge: "c1"}}
	router := newAuthLoginRouter(kratos, hydra, audits)

	rec := postLogin(router, `{"login_challenge":"c1","identifier":"bad@example.com","password":"wrong"}`)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d, want 401 body=%s", rec.Code, rec.Body.String())
	}
	if len(hydra.accepted) != 0 {
		t.Errorf("Hydra accept must not be called on invalid credentials, got %v", hydra.accepted)
	}
	if len(audits.logs) == 0 || audits.logs[0].Action != "auth.login.failed" {
		t.Errorf("expected auth.login.failed audit, got %+v", audits.logs)
	}
}

func TestAuthLogin_LoginChallengeUnknown(t *testing.T) {
	audits := &memoryAuditStore{}
	hydra := &fakeHydraAdmin{getErr: ErrHydraChallengeNotFound}
	router := newAuthLoginRouter(&fakeKratosLogin{}, hydra, audits)

	rec := postLogin(router, `{"login_challenge":"stale","identifier":"u","password":"p"}`)
	if rec.Code != http.StatusGone {
		t.Errorf("status=%d, want 410", rec.Code)
	}
}

func TestAuthLogin_FlowExpired(t *testing.T) {
	audits := &memoryAuditStore{}
	kratos := &fakeKratosLogin{
		flow:      KratosLoginFlow{ID: "flow-1"},
		submitErr: ErrKratosFlowExpired,
	}
	hydra := &fakeHydraAdmin{loginRequest: HydraLoginRequest{Challenge: "c1"}}
	router := newAuthLoginRouter(kratos, hydra, audits)

	rec := postLogin(router, `{"login_challenge":"c1","identifier":"u","password":"p"}`)
	if rec.Code != http.StatusGone {
		t.Errorf("status=%d, want 410", rec.Code)
	}
}

func TestAuthLogin_MissingChallenge(t *testing.T) {
	router := newAuthLoginRouter(&fakeKratosLogin{}, &fakeHydraAdmin{}, &memoryAuditStore{})
	rec := postLogin(router, `{"identifier":"u","password":"p"}`)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status=%d, want 400", rec.Code)
	}
}

func TestAuthLogin_MissingCredentials(t *testing.T) {
	router := newAuthLoginRouter(&fakeKratosLogin{}, &fakeHydraAdmin{loginRequest: HydraLoginRequest{Challenge: "c1"}}, &memoryAuditStore{})
	rec := postLogin(router, `{"login_challenge":"c1"}`)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status=%d, want 400", rec.Code)
	}
}

func TestAuthLogin_StoreUnavailable(t *testing.T) {
	router := NewRouter(RouterConfig{}) // no KratosLogin / HydraAdmin
	rec := postLogin(router, `{"login_challenge":"c1","identifier":"u","password":"p"}`)
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status=%d, want 503", rec.Code)
	}
}

func TestAuthLogin_SubjectFallbackWhenMetadataMissing(t *testing.T) {
	audits := &memoryAuditStore{}
	kratos := &fakeKratosLogin{
		flow: KratosLoginFlow{ID: "flow-1"},
		// UserID empty -> handler should fall back to identity.id and audit it
		identity: KratosIdentity{ID: "kratos-uuid"},
	}
	hydra := &fakeHydraAdmin{
		loginRequest: HydraLoginRequest{Challenge: "c1"},
		redirectTo:   "http://hydra/cb",
	}
	router := newAuthLoginRouter(kratos, hydra, audits)

	rec := postLogin(router, `{"login_challenge":"c1","identifier":"u","password":"p"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if len(hydra.accepted) != 1 || hydra.accepted[0] != "kratos-uuid" {
		t.Errorf("Hydra subject = %v, want [kratos-uuid] (fallback)", hydra.accepted)
	}
	hasFallback := false
	for _, a := range audits.logs {
		if a.Action == "auth.login.subject_fallback" {
			hasFallback = true
			break
		}
	}
	if !hasFallback {
		t.Errorf("expected auth.login.subject_fallback audit, got %+v", audits.logs)
	}
}

// L4-B: 로그인 성공 시 SessionToken 이 KratosSessionCache 에 user_id 키로 저장.
func TestAuthLogin_CachesKratosSessionToken(t *testing.T) {
	audits := &memoryAuditStore{}
	kratos := &fakeKratosLogin{
		flow:     KratosLoginFlow{ID: "flow-1"},
		identity: KratosIdentity{ID: "kratos-uuid", UserID: "u1", SessionToken: "kratos-sess-1"},
	}
	hydra := &fakeHydraAdmin{
		loginRequest: HydraLoginRequest{Challenge: "c1"},
		redirectTo:   "http://hydra/cb",
	}
	cache := NewKratosSessionCache()
	router := NewRouter(RouterConfig{
		KratosLogin:        kratos,
		HydraAdmin:         hydra,
		AuditStore:         audits,
		KratosSessionCache: cache,
	})

	rec := postLogin(router, `{"login_challenge":"c1","identifier":"u","password":"p"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	got, ok := cache.Get("u1")
	if !ok || got != "kratos-sess-1" {
		t.Errorf("cache.Get(u1) = (%q, %v), want (kratos-sess-1, true)", got, ok)
	}
}

// L4-B: metadata_public.user_id 누락 → identity.id fallback 시에도 같은 키로 캐싱.
func TestAuthLogin_CacheKeyFallsBackToIdentityID(t *testing.T) {
	kratos := &fakeKratosLogin{
		flow:     KratosLoginFlow{ID: "flow-1"},
		identity: KratosIdentity{ID: "kratos-uuid-2", SessionToken: "kratos-sess-2"},
	}
	hydra := &fakeHydraAdmin{
		loginRequest: HydraLoginRequest{Challenge: "c1"},
		redirectTo:   "http://hydra/cb",
	}
	cache := NewKratosSessionCache()
	router := NewRouter(RouterConfig{
		KratosLogin:        kratos,
		HydraAdmin:         hydra,
		AuditStore:         &memoryAuditStore{},
		KratosSessionCache: cache,
	})

	rec := postLogin(router, `{"login_challenge":"c1","identifier":"u","password":"p"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if got, ok := cache.Get("kratos-uuid-2"); !ok || got != "kratos-sess-2" {
		t.Errorf("cache.Get(kratos-uuid-2) = (%q, %v), want (kratos-sess-2, true)", got, ok)
	}
	if _, ok := cache.Get(""); ok {
		t.Errorf("empty subject must not be cached")
	}
}

// AcceptLoginRequest 호출 인자 검증 — 캐시된 subject 가 fallthrough 없이 사용되는지.
func TestAuthLogin_AcceptError(t *testing.T) {
	kratos := &fakeKratosLogin{
		flow:     KratosLoginFlow{ID: "flow-1"},
		identity: KratosIdentity{ID: "kratos-uuid", UserID: "u1"},
	}
	hydra := &fakeHydraAdmin{
		loginRequest: HydraLoginRequest{Challenge: "c1"},
		acceptErr:    errors.New("hydra accept boom"),
	}
	router := newAuthLoginRouter(kratos, hydra, &memoryAuditStore{})

	rec := postLogin(router, `{"login_challenge":"c1","identifier":"u","password":"p"}`)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status=%d, want 500", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "internal error") {
		t.Errorf("error body should be masked, got: %s", rec.Body.String())
	}
}
