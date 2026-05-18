package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestKeycloakAdminClient_CRUDFlow(t *testing.T) {
	var (
		gotAuthHeader string
		gotEnabled    *bool
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/realms/devhub/protocol/openid-connect/token", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("token method=%s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"kc-token"}`))
	})
	mux.HandleFunc("/admin/realms/devhub/users", func(w http.ResponseWriter, r *http.Request) {
		gotAuthHeader = r.Header.Get("Authorization")
		switch r.Method {
		case http.MethodPost:
			w.Header().Set("Location", "http://x/admin/realms/devhub/users/kc-id-1")
			w.WriteHeader(http.StatusCreated)
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[{"id":"kc-id-1","username":"alice"}]`))
		default:
			t.Fatalf("unexpected users method=%s", r.Method)
		}
	})
	mux.HandleFunc("/admin/realms/devhub/users/kc-id-1/reset-password", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("reset method=%s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/admin/realms/devhub/users/kc-id-1", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			var payload map[string]any
			_ = json.NewDecoder(r.Body).Decode(&payload)
			if v, ok := payload["enabled"].(bool); ok {
				gotEnabled = &v
			}
			w.WriteHeader(http.StatusNoContent)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected user-id method=%s", r.Method)
		}
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := &KeycloakAdminClient{
		AdminURL:     srv.URL,
		Realm:        "devhub",
		ClientID:     "devhub-admin",
		ClientSecret: "secret",
		IssuerURL:    srv.URL + "/realms/devhub",
	}
	ctx := context.Background()

	id, err := c.CreateIdentity(ctx, "alice@example.com", "Alice", "alice", "TmpPass123!")
	if err != nil {
		t.Fatalf("CreateIdentity err: %v", err)
	}
	if id != "kc-id-1" {
		t.Fatalf("id=%q want kc-id-1", id)
	}
	if !strings.HasPrefix(gotAuthHeader, "Bearer ") {
		t.Fatalf("missing bearer auth header: %q", gotAuthHeader)
	}

	found, err := c.FindIdentityByUserID(ctx, "alice")
	if err != nil {
		t.Fatalf("FindIdentityByUserID err: %v", err)
	}
	if found != "kc-id-1" {
		t.Fatalf("found=%q want kc-id-1", found)
	}

	if err := c.UpdateIdentityPassword(ctx, "kc-id-1", "NewPass123!"); err != nil {
		t.Fatalf("UpdateIdentityPassword err: %v", err)
	}
	if err := c.SetIdentityState(ctx, "kc-id-1", false); err != nil {
		t.Fatalf("SetIdentityState err: %v", err)
	}
	if gotEnabled == nil || *gotEnabled {
		t.Fatalf("enabled payload not set to false")
	}
	if err := c.DeleteIdentity(ctx, "kc-id-1"); err != nil {
		t.Fatalf("DeleteIdentity err: %v", err)
	}
}

