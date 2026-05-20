package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestKeycloakAdminClient_FindAndAuthFlow — ADR-0020 sub-carve E (sprint -n)
// 후 KeycloakAdminClient 의 IdentityAdmin write methods (CreateIdentity /
// UpdateIdentityPassword / SetIdentityState / DeleteIdentity) 모두 제거됨.
// 남은 IdentityAdmin 표면은 FindIdentityByUserID (view-users role 만 요구) 만.
// 본 test 는 admin token fetch + bearer header + GET /users?username= 흐름.
func TestKeycloakAdminClient_FindAndAuthFlow(t *testing.T) {
	var gotAuthHeader string

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
		if r.Method != http.MethodGet {
			t.Fatalf("users method=%s; want GET (read-only)", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"kc-id-1","username":"alice"}]`))
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

	found, err := c.FindIdentityByUserID(ctx, "alice")
	if err != nil {
		t.Fatalf("FindIdentityByUserID err: %v", err)
	}
	if found != "kc-id-1" {
		t.Fatalf("found=%q want kc-id-1", found)
	}
	if !strings.HasPrefix(gotAuthHeader, "Bearer ") {
		t.Fatalf("missing bearer auth header: %q", gotAuthHeader)
	}
}
