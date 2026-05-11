package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateSettingsFlow_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/self-service/settings/api" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.Header.Get("X-Session-Token") != "sess-1" {
			t.Errorf("X-Session-Token = %q, want sess-1", r.Header.Get("X-Session-Token"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"flow-abc","state":"show_form","ui":{"nodes":[],"messages":[]}}`))
	}))
	defer srv.Close()

	c := &KratosClient{PublicURL: srv.URL}
	id, err := c.CreateSettingsFlow(context.Background(), "sess-1")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if id != "flow-abc" {
		t.Errorf("flow_id = %q, want flow-abc", id)
	}
}

func TestCreateSettingsFlow_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"id":"session_inactive","reason":"no session"}}`))
	}))
	defer srv.Close()

	c := &KratosClient{PublicURL: srv.URL}
	_, err := c.CreateSettingsFlow(context.Background(), "bogus")
	se := IsKratosSettingsError(err)
	if se == nil {
		t.Fatalf("err = %v, want KratosSettingsError", err)
	}
	if se.Code != KratosSettingsSessionInvalid {
		t.Errorf("code = %v, want session_invalid", se.Code)
	}
}

func TestCreateSettingsFlow_PrivilegedRequired(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":{"id":"session_refresh_required","reason":"reauth"}}`))
	}))
	defer srv.Close()

	c := &KratosClient{PublicURL: srv.URL}
	_, err := c.CreateSettingsFlow(context.Background(), "sess-old")
	se := IsKratosSettingsError(err)
	if se == nil || se.Code != KratosSettingsPrivilegedRequired {
		t.Errorf("err = %v, want privileged_required", err)
	}
}

func TestSubmitSettingsPassword_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("flow") != "flow-abc" {
			t.Errorf("flow query = %q", r.URL.Query().Get("flow"))
		}
		if r.Header.Get("X-Session-Token") != "sess-1" {
			t.Errorf("X-Session-Token = %q", r.Header.Get("X-Session-Token"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"flow-abc","state":"success","ui":{}}`))
	}))
	defer srv.Close()

	c := &KratosClient{PublicURL: srv.URL}
	if err := c.SubmitSettingsPassword(context.Background(), "sess-1", "flow-abc", "NewPass-2026!"); err != nil {
		t.Errorf("submit: %v", err)
	}
}

func TestSubmitSettingsPassword_ValidationOn400(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"id":"f","state":"show_form","ui":{"nodes":[{"messages":[{"text":"password too short"}]}]}}`))
	}))
	defer srv.Close()

	c := &KratosClient{PublicURL: srv.URL}
	err := c.SubmitSettingsPassword(context.Background(), "sess", "f", "short")
	se := IsKratosSettingsError(err)
	if se == nil || se.Code != KratosSettingsValidation {
		t.Fatalf("err = %v, want validation", err)
	}
	if !strings.Contains(se.Message, "password too short") {
		t.Errorf("message did not surface validation text: %q", se.Message)
	}
}

func TestSubmitSettingsPassword_ValidationOn200WithMessages(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"f","state":"show_form","ui":{"messages":[{"text":"data breach list match"}]}}`))
	}))
	defer srv.Close()

	c := &KratosClient{PublicURL: srv.URL}
	err := c.SubmitSettingsPassword(context.Background(), "sess", "f", "pass")
	se := IsKratosSettingsError(err)
	if se == nil || se.Code != KratosSettingsValidation || !strings.Contains(se.Message, "breach") {
		t.Errorf("err = %v, want validation w/ breach msg", err)
	}
}

func TestSubmitSettingsPassword_PrivilegedRequired(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":{"id":"session_refresh_required"}}`))
	}))
	defer srv.Close()

	c := &KratosClient{PublicURL: srv.URL}
	err := c.SubmitSettingsPassword(context.Background(), "sess", "f", "Pass1234567!")
	se := IsKratosSettingsError(err)
	if se == nil || se.Code != KratosSettingsPrivilegedRequired {
		t.Errorf("err = %v, want privileged_required", err)
	}
}

func TestSubmitSettingsPassword_FlowExpired(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusGone)
		_, _ = w.Write([]byte(`{"error":{"id":"self_service_flow_expired"}}`))
	}))
	defer srv.Close()

	c := &KratosClient{PublicURL: srv.URL}
	err := c.SubmitSettingsPassword(context.Background(), "sess", "f", "Pass1234567!")
	se := IsKratosSettingsError(err)
	if se == nil || se.Code != KratosSettingsFlowExpired {
		t.Errorf("err = %v, want flow_expired", err)
	}
}

func TestSubmitSettingsPassword_5xxPropagatesAsRawError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"kratos boom"}`))
	}))
	defer srv.Close()

	c := &KratosClient{PublicURL: srv.URL}
	err := c.SubmitSettingsPassword(context.Background(), "sess", "f", "Pass1234567!")
	if err == nil {
		t.Fatalf("expected error on 5xx")
	}
	if IsKratosSettingsError(err) != nil {
		t.Errorf("5xx should not coerce to KratosSettingsError; got %v", err)
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("err did not surface status code: %v", err)
	}
}

func TestSubmitSettingsPassword_EmptySessionToken(t *testing.T) {
	c := &KratosClient{PublicURL: "http://unused"}
	err := c.SubmitSettingsPassword(context.Background(), "", "f", "Pass1234567!")
	se := IsKratosSettingsError(err)
	if se == nil || se.Code != KratosSettingsSessionInvalid {
		t.Errorf("err = %v, want session_invalid", err)
	}
}

// errors.Is interop sanity — wrapping behaves as expected.
func TestKratosSettingsError_ErrorsAs(t *testing.T) {
	want := &KratosSettingsError{Code: KratosSettingsValidation, Message: "x"}
	wrapped := errors.New("outer")
	if errors.As(wrapped, &want) {
		t.Errorf("unrelated wrapped error must not unwrap to KratosSettingsError")
	}
}
