package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeHydraLogout struct {
	logoutRequest HydraLogoutRequest
	getErr        error
	redirectTo    string
	acceptErr     error
	acceptCalled  int
}

func (f *fakeHydraLogout) GetLogoutRequest(_ context.Context, _ string) (HydraLogoutRequest, error) {
	if f.getErr != nil {
		return HydraLogoutRequest{}, f.getErr
	}
	return f.logoutRequest, nil
}

func (f *fakeHydraLogout) AcceptLogoutRequest(_ context.Context, _ string) (string, error) {
	f.acceptCalled++
	if f.acceptErr != nil {
		return "", f.acceptErr
	}
	return f.redirectTo, nil
}

type fakeHydraRevoker struct {
	revoked     []string // tokens that RevokeRefreshToken was called with
	clientIDs   []string
	err         error
	callCount   int
}

func (f *fakeHydraRevoker) RevokeRefreshToken(_ context.Context, refreshToken, clientID string) error {
	f.callCount++
	f.revoked = append(f.revoked, refreshToken)
	f.clientIDs = append(f.clientIDs, clientID)
	return f.err
}

func newAuthLogoutRouter(logout HydraLogoutAdmin, revoker HydraTokenRevoker, audits *memoryAuditStore) http.Handler {
	return NewRouter(RouterConfig{
		HydraLogout:  logout,
		HydraRevoker: revoker,
		AuditStore:   audits,
	})
}

func postLogout(t *testing.T, router http.Handler, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

// 1) logout_challenge 만 있고 성공 — Hydra accept 호출 + redirect_to 반환.
func TestAuthLogout_ChallengeOnlySuccess(t *testing.T) {
	audits := &memoryAuditStore{}
	logout := &fakeHydraLogout{
		logoutRequest: HydraLogoutRequest{Subject: "u1", SID: "sid-1"},
		redirectTo:    "http://localhost:3000/",
	}
	logout.logoutRequest.Client.ClientID = "devhub-frontend"
	revoker := &fakeHydraRevoker{}
	router := newAuthLogoutRouter(logout, revoker, audits)

	rec := postLogout(t, router, `{"logout_challenge":"c1"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if logout.acceptCalled != 1 {
		t.Errorf("AcceptLogoutRequest called %d times, want 1", logout.acceptCalled)
	}
	if revoker.callCount != 0 {
		t.Errorf("RevokeRefreshToken should not be called when refresh_token absent")
	}
	if !strings.Contains(rec.Body.String(), "http://localhost:3000/") {
		t.Errorf("response missing redirect_to: %s", rec.Body.String())
	}
	if len(audits.logs) != 1 || audits.logs[0].Action != "auth.logout.succeeded" {
		t.Errorf("expected auth.logout.succeeded audit, got %+v", audits.logs)
	}
	if audits.logs[0].TargetID != "u1" {
		t.Errorf("audit subject = %q, want u1", audits.logs[0].TargetID)
	}
}

// 2) refresh_token + client_id 만 있고 성공 — Hydra revoke 호출, accept 미호출, redirect_to 빈 문자열.
func TestAuthLogout_RevokeOnlySuccess(t *testing.T) {
	audits := &memoryAuditStore{}
	logout := &fakeHydraLogout{}
	revoker := &fakeHydraRevoker{}
	router := newAuthLogoutRouter(logout, revoker, audits)

	rec := postLogout(t, router, `{"refresh_token":"r1","client_id":"devhub-frontend"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if logout.acceptCalled != 0 {
		t.Errorf("Accept must not be called when challenge absent")
	}
	if revoker.callCount != 1 || revoker.revoked[0] != "r1" || revoker.clientIDs[0] != "devhub-frontend" {
		t.Errorf("revoke called wrong: %+v", revoker)
	}
	if !strings.Contains(rec.Body.String(), `"revoke_status":"succeeded"`) {
		t.Errorf("body missing revoke_status:succeeded: %s", rec.Body.String())
	}
	if len(audits.logs) != 1 || audits.logs[0].Action != "auth.logout.succeeded" {
		t.Errorf("expected auth.logout.succeeded audit, got %+v", audits.logs)
	}
}

// 3) 둘 다 — revoke + accept 모두 호출. revoke 실패해도 accept 성공이면 200 + revoke_failed audit.
func TestAuthLogout_RevokeFailedAcceptOK(t *testing.T) {
	audits := &memoryAuditStore{}
	logout := &fakeHydraLogout{
		logoutRequest: HydraLogoutRequest{Subject: "u1"},
		redirectTo:    "http://localhost:3000/",
	}
	logout.logoutRequest.Client.ClientID = "devhub-frontend"
	revoker := &fakeHydraRevoker{err: errors.New("hydra public down")}
	router := newAuthLogoutRouter(logout, revoker, audits)

	rec := postLogout(t, router, `{"logout_challenge":"c1","refresh_token":"r1"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if logout.acceptCalled != 1 {
		t.Errorf("Accept should still run after revoke failure")
	}
	if !strings.Contains(rec.Body.String(), `"revoke_status":"failed"`) {
		t.Errorf("body missing revoke_status:failed: %s", rec.Body.String())
	}
	hasRevokeFailed := false
	hasSucceeded := false
	for _, a := range audits.logs {
		if a.Action == "auth.logout.revoke_failed" {
			hasRevokeFailed = true
		}
		if a.Action == "auth.logout.succeeded" {
			hasSucceeded = true
		}
	}
	if !hasRevokeFailed || !hasSucceeded {
		t.Errorf("expected both revoke_failed and succeeded audits, got %+v", audits.logs)
	}
}

// 4) Hydra accept 5xx → 500.
func TestAuthLogout_AcceptError(t *testing.T) {
	logout := &fakeHydraLogout{
		logoutRequest: HydraLogoutRequest{Subject: "u1"},
		acceptErr:     errors.New("hydra accept boom"),
	}
	router := newAuthLogoutRouter(logout, &fakeHydraRevoker{}, &memoryAuditStore{})

	rec := postLogout(t, router, `{"logout_challenge":"c1"}`)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d, want 500 body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "internal error") {
		t.Errorf("error body should be masked, got: %s", rec.Body.String())
	}
}

// 5) logout_challenge 가 unknown → 410.
func TestAuthLogout_ChallengeUnknown(t *testing.T) {
	logout := &fakeHydraLogout{getErr: ErrHydraChallengeNotFound}
	router := newAuthLogoutRouter(logout, &fakeHydraRevoker{}, &memoryAuditStore{})

	rec := postLogout(t, router, `{"logout_challenge":"stale"}`)
	if rec.Code != http.StatusGone {
		t.Errorf("status=%d, want 410", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "logout_challenge_unknown") {
		t.Errorf("body missing code logout_challenge_unknown: %s", rec.Body.String())
	}
}

// 추가: 둘 다 비어있으면 400.
func TestAuthLogout_BothEmpty(t *testing.T) {
	router := newAuthLogoutRouter(&fakeHydraLogout{}, &fakeHydraRevoker{}, &memoryAuditStore{})
	rec := postLogout(t, router, `{}`)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status=%d, want 400 body=%s", rec.Code, rec.Body.String())
	}
}

// 추가: refresh_token 만 있고 client_id 누락 → 400.
func TestAuthLogout_RefreshWithoutClientID(t *testing.T) {
	router := newAuthLogoutRouter(&fakeHydraLogout{}, &fakeHydraRevoker{}, &memoryAuditStore{})
	rec := postLogout(t, router, `{"refresh_token":"r1"}`)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status=%d, want 400 body=%s", rec.Code, rec.Body.String())
	}
}

// 추가: HydraLogout 미주입 → 503.
func TestAuthLogout_StoreUnavailable(t *testing.T) {
	router := NewRouter(RouterConfig{}) // no HydraLogout
	rec := postLogout(t, router, `{"logout_challenge":"c1"}`)
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status=%d, want 503", rec.Code)
	}
}
