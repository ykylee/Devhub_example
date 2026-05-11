package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeHRDB struct {
	email, userID, dept string
	err                 error
}

func (f *fakeHRDB) Lookup(_ context.Context, _, _, _ string) (string, string, string, error) {
	if f.err != nil {
		return "", "", "", f.err
	}
	return f.email, f.userID, f.dept, nil
}

// TestEnvelope_NewlyFixedHandlersHaveStatusKey is the regression net for the
// PR-B1 envelope alignment work. Every JSON body the handler emits must carry
// the {status, ...} key per docs/backend_api_contract.md §1. Earlier sprints
// silently shipped a few status-less bodies (auth/signup, hr/lookup); this
// test makes sure they stay aligned.
func TestEnvelope_NewlyFixedHandlersHaveStatusKey(t *testing.T) {
	cases := []struct {
		name         string
		method       string
		path         string
		body         string
		setupRouter  func() http.Handler
		wantStatuses map[int]bool // expected HTTP status -> envelope status key required
	}{
		{
			name:   "auth/signup invalid payload -> rejected",
			method: http.MethodPost,
			path:   "/api/v1/auth/signup",
			body:   `{}`,
			setupRouter: func() http.Handler {
				return NewRouter(RouterConfig{AuthDevFallback: true})
			},
			wantStatuses: map[int]bool{http.StatusBadRequest: true},
		},
		{
			name:   "auth/signup unavailable when HRDB nil -> unavailable",
			method: http.MethodPost,
			path:   "/api/v1/auth/signup",
			body:   `{"name":"a","system_id":"b","employee_id":"c","password":"d"}`,
			setupRouter: func() http.Handler {
				return NewRouter(RouterConfig{AuthDevFallback: true})
			},
			wantStatuses: map[int]bool{http.StatusServiceUnavailable: true},
		},
		{
			name:   "hr/lookup missing query -> rejected",
			method: http.MethodGet,
			path:   "/api/v1/hr/lookup",
			body:   "",
			setupRouter: func() http.Handler {
				return NewRouter(RouterConfig{AuthDevFallback: true})
			},
			wantStatuses: map[int]bool{http.StatusBadRequest: true},
		},
		{
			name:   "hr/lookup unavailable when HRDB nil -> unavailable",
			method: http.MethodGet,
			path:   "/api/v1/hr/lookup?system_id=u1",
			body:   "",
			setupRouter: func() http.Handler {
				return NewRouter(RouterConfig{AuthDevFallback: true})
			},
			wantStatuses: map[int]bool{http.StatusServiceUnavailable: true},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			router := tc.setupRouter()
			var bodyReader *strings.Reader
			if tc.body != "" {
				bodyReader = strings.NewReader(tc.body)
			} else {
				bodyReader = strings.NewReader("")
			}
			req := httptest.NewRequest(tc.method, tc.path, bodyReader)
			if tc.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if !tc.wantStatuses[rec.Code] {
				t.Fatalf("status=%d unexpected (want one of %v); body=%s", rec.Code, tc.wantStatuses, rec.Body.String())
			}
			var payload map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
				t.Fatalf("response is not JSON: %v body=%s", err, rec.Body.String())
			}
			if _, ok := payload["status"]; !ok {
				t.Errorf("response missing 'status' envelope key: %s", rec.Body.String())
			}
		})
	}
}

// TestEnvelope_HrLookupSuccessShape locks the new envelope shape for the
// hr/lookup happy path: {status: "ok", data: {email, user_id, ...}}. The mock
// HRDB returns canned values so this exercises the response shape only.
func TestEnvelope_HrLookupSuccessShape(t *testing.T) {
	router := NewRouter(RouterConfig{
		HRDB:            &fakeHRDB{email: "a@x", userID: "u1", dept: "Eng"},
		AuthDevFallback: true,
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/hr/lookup?system_id=u1", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var payload struct {
		Status string `json:"status"`
		Data   struct {
			Email      string `json:"email"`
			UserID     string `json:"user_id"`
			Department string `json:"department"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode: %v body=%s", err, rec.Body.String())
	}
	if payload.Status != "ok" {
		t.Errorf("status=%q want ok", payload.Status)
	}
	if payload.Data.Email != "a@x" || payload.Data.UserID != "u1" || payload.Data.Department != "Eng" {
		t.Errorf("data shape mismatch: %+v", payload.Data)
	}
}
