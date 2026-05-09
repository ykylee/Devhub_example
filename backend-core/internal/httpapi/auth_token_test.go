package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeHydraTokenExchanger struct {
	gotReq HydraTokenExchangeRequest
	resp   HydraTokenExchangeResponse
	err    error
}

func (f *fakeHydraTokenExchanger) ExchangeAuthorizationCode(_ context.Context, req HydraTokenExchangeRequest) (HydraTokenExchangeResponse, error) {
	f.gotReq = req
	if f.err != nil {
		return HydraTokenExchangeResponse{}, f.err
	}
	return f.resp, nil
}

func newAuthTokenRouter(exchanger HydraTokenExchanger) http.Handler {
	return NewRouter(RouterConfig{
		HydraToken: exchanger,
	})
}

func postAuthToken(t *testing.T, router http.Handler, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/token", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestAuthTokenSuccess(t *testing.T) {
	fake := &fakeHydraTokenExchanger{
		resp: HydraTokenExchangeResponse{
			AccessToken:  "access-1",
			TokenType:    "Bearer",
			ExpiresIn:    1800,
			RefreshToken: "refresh-1",
			IDToken:      "id-1",
			Scope:        "openid profile offline_access",
		},
	}
	rec := postAuthToken(t, newAuthTokenRouter(fake), `{"code":"c1","code_verifier":"v1","redirect_uri":"http://localhost:3000/auth/callback","client_id":"devhub-frontend"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if fake.gotReq.Code != "c1" || fake.gotReq.CodeVerifier != "v1" || fake.gotReq.ClientID != "devhub-frontend" {
		t.Fatalf("unexpected request forwarded: %+v", fake.gotReq)
	}
}

func TestAuthTokenInvalidGrant(t *testing.T) {
	fake := &fakeHydraTokenExchanger{err: ErrHydraTokenInvalidGrant}
	rec := postAuthToken(t, newAuthTokenRouter(fake), `{"code":"c1","code_verifier":"v1","redirect_uri":"http://localhost:3000/auth/callback","client_id":"devhub-frontend"}`)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAuthTokenBadRequestMissingFields(t *testing.T) {
	rec := postAuthToken(t, newAuthTokenRouter(&fakeHydraTokenExchanger{}), `{"code":"c1"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAuthTokenUnavailableWhenClientNotWired(t *testing.T) {
	rec := postAuthToken(t, NewRouter(RouterConfig{}), `{"code":"c1","code_verifier":"v1","redirect_uri":"http://localhost:3000/auth/callback","client_id":"devhub-frontend"}`)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAuthTokenServerError(t *testing.T) {
	fake := &fakeHydraTokenExchanger{err: errors.New("hydra down")}
	rec := postAuthToken(t, newAuthTokenRouter(fake), `{"code":"c1","code_verifier":"v1","redirect_uri":"http://localhost:3000/auth/callback","client_id":"devhub-frontend"}`)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}
