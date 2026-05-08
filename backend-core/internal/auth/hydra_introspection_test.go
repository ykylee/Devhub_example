package auth

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestVerifyBearerTokenSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/admin/oauth2/introspect" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		if got := r.Header.Get("Content-Type"); got != "application/x-www-form-urlencoded" {
			t.Errorf("unexpected content-type %q", got)
		}
		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()
		if !strings.Contains(string(body), "token=valid-token") {
			t.Errorf("expected token=valid-token in body, got %q", body)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"active": true,
			"sub": "user-123",
			"username": "alice",
			"scope": "read write",
			"client_id": "devhub-frontend",
			"ext": {"role": "manager"}
		}`))
	}))
	defer srv.Close()

	v := &HydraIntrospectionVerifier{AdminURL: srv.URL}
	actor, err := v.VerifyBearerToken(context.Background(), "valid-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if actor.Login != "alice" {
		t.Errorf("expected login=alice, got %q", actor.Login)
	}
	if actor.Subject != "user-123" {
		t.Errorf("expected subject=user-123, got %q", actor.Subject)
	}
	if actor.Role != "manager" {
		t.Errorf("expected role=manager, got %q", actor.Role)
	}
}

func TestVerifyBearerTokenInactive(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"active": false}`))
	}))
	defer srv.Close()

	v := &HydraIntrospectionVerifier{AdminURL: srv.URL}
	_, err := v.VerifyBearerToken(context.Background(), "expired-token")
	if !errors.Is(err, ErrTokenInactive) {
		t.Fatalf("expected ErrTokenInactive, got %v", err)
	}
}

func TestVerifyBearerTokenNon2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal", http.StatusInternalServerError)
	}))
	defer srv.Close()

	v := &HydraIntrospectionVerifier{AdminURL: srv.URL}
	_, err := v.VerifyBearerToken(context.Background(), "any")
	if err == nil {
		t.Fatal("expected error on 5xx, got nil")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected status 500 in error, got %v", err)
	}
}

func TestVerifyBearerTokenLoginFallsBackToSubject(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"active": true,
			"sub": "user-only-subject",
			"ext": {"role": "developer"}
		}`))
	}))
	defer srv.Close()

	v := &HydraIntrospectionVerifier{AdminURL: srv.URL}
	actor, err := v.VerifyBearerToken(context.Background(), "t")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if actor.Login != "user-only-subject" {
		t.Errorf("expected login to fall back to subject, got %q", actor.Login)
	}
}

func TestVerifyBearerTokenMissingAdminURL(t *testing.T) {
	v := &HydraIntrospectionVerifier{}
	_, err := v.VerifyBearerToken(context.Background(), "t")
	if err == nil {
		t.Fatal("expected error when AdminURL is empty")
	}
}

func TestVerifyBearerTokenCustomRoleClaim(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"active": true,
			"sub": "u",
			"username": "u",
			"ext": {"groups": "system_admin"}
		}`))
	}))
	defer srv.Close()

	v := &HydraIntrospectionVerifier{AdminURL: srv.URL, RoleClaim: "ext.groups"}
	actor, err := v.VerifyBearerToken(context.Background(), "t")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if actor.Role != "system_admin" {
		t.Errorf("expected role=system_admin from ext.groups, got %q", actor.Role)
	}
}
