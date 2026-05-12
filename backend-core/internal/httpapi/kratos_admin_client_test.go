package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// These tests pin the wire contract of KratosAdminClient against the Ory
// Kratos Admin API. They exist because a prior rewrite turned
// UpdateIdentityPassword / SetIdentityState / DeleteIdentity into noop
// stubs and shipped a CreateIdentity payload that did not match the
// devhub_user schema — both regressions were masked by tests that only
// exercised MockKratosAdmin. Anything that touches this file must keep
// these HTTP-level assertions intact.

func TestKratosAdminClient_CreateIdentity_SuccessPayload(t *testing.T) {
	var captured map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/admin/identities" {
			t.Errorf("path = %s, want /admin/identities", r.URL.Path)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &captured); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"kratos-uuid-1"}`))
	}))
	defer srv.Close()

	c := &KratosAdminClient{AdminURL: srv.URL}
	id, err := c.CreateIdentity(context.Background(), "alice@example.com", "Alice", "alice", "Pass1234567!")
	if err != nil {
		t.Fatalf("CreateIdentity: %v", err)
	}
	if id != "kratos-uuid-1" {
		t.Errorf("id = %q, want kratos-uuid-1", id)
	}

	// devhub_user schema + traits/metadata_public mapping must match
	// identity.schema.json (ADR-0001 §5). A previous regression sent
	// schema_id="default" with traits {email, name}; this test exists to
	// pin the contract.
	if got, _ := captured["schema_id"].(string); got != "devhub_user" {
		t.Errorf("schema_id = %q, want devhub_user", got)
	}
	if got, _ := captured["state"].(string); got != "active" {
		t.Errorf("state = %q, want active", got)
	}
	traits, _ := captured["traits"].(map[string]any)
	if traits == nil {
		t.Fatalf("traits missing from payload")
	}
	if traits["system_id"] != "alice" {
		t.Errorf("traits.system_id = %v, want alice", traits["system_id"])
	}
	if traits["email"] != "alice@example.com" {
		t.Errorf("traits.email = %v, want alice@example.com", traits["email"])
	}
	if traits["display_name"] != "Alice" {
		t.Errorf("traits.display_name = %v, want Alice", traits["display_name"])
	}
	meta, _ := captured["metadata_public"].(map[string]any)
	if meta == nil || meta["user_id"] != "alice" {
		t.Errorf("metadata_public.user_id = %v, want alice", meta["user_id"])
	}
	creds, _ := captured["credentials"].(map[string]any)
	pwd, _ := creds["password"].(map[string]any)
	cfg, _ := pwd["config"].(map[string]any)
	if cfg["password"] != "Pass1234567!" {
		t.Errorf("credentials.password.config.password = %v", cfg["password"])
	}
}

func TestKratosAdminClient_CreateIdentity_NonCreatedSurfacesStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"error":{"id":"conflict"}}`))
	}))
	defer srv.Close()

	c := &KratosAdminClient{AdminURL: srv.URL}
	_, err := c.CreateIdentity(context.Background(), "a@b.c", "A", "a", "Pass1234567!")
	if err == nil {
		t.Fatalf("expected error on 409")
	}
	if !strings.Contains(err.Error(), "409") {
		t.Errorf("error did not surface status: %v", err)
	}
}

func TestKratosAdminClient_CreateIdentity_RejectsEmptyAdminURL(t *testing.T) {
	c := &KratosAdminClient{}
	_, err := c.CreateIdentity(context.Background(), "a@b.c", "A", "a", "p")
	if err == nil {
		t.Fatalf("expected error for empty AdminURL")
	}
}

func TestKratosAdminClient_FindIdentityByUserID_MatchesMetadataPublic(t *testing.T) {
	// Regression guard: the previous rewrite decoded into KratosIdentity
	// (no JSON tags) and checked ident.UserID — Kratos returns user_id
	// under metadata_public.user_id, so every match silently failed.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/admin/identities" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("method = %s", r.Method)
		}
		_, _ = w.Write([]byte(`[
			{"id":"uuid-A","metadata_public":{"user_id":"other"}},
			{"id":"uuid-B","metadata_public":{"user_id":"alice"}},
			{"id":"uuid-C","metadata_public":{"user_id":"bob"}}
		]`))
	}))
	defer srv.Close()

	c := &KratosAdminClient{AdminURL: srv.URL}
	id, err := c.FindIdentityByUserID(context.Background(), "alice")
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if id != "uuid-B" {
		t.Errorf("id = %q, want uuid-B", id)
	}
}

func TestKratosAdminClient_FindIdentityByUserID_Paginates(t *testing.T) {
	// Kratos /admin/identities pagination is 0-based — verified empirically
	// against v26.2.0 (page=0 returns first batch, page=1 returns second).
	// The earlier 1-based start silently returned empty first page and
	// short-circuited to ErrKratosIdentityNotFound.
	var pages int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pages++
		// Pin the page size — if the client ever silently lowers per_page,
		// the pagination cap math (page > 39 → 10k identities) goes off and
		// large Kratos deployments would silently truncate scans.
		if got := r.URL.Query().Get("per_page"); got != "250" {
			t.Errorf("per_page = %q, want 250", got)
		}
		page := r.URL.Query().Get("page")
		switch page {
		case "0":
			// Return a full page (250) of non-matching identities so the
			// client must request page 1.
			var batch []map[string]any
			for i := 0; i < 250; i++ {
				batch = append(batch, map[string]any{
					"id":              fmt.Sprintf("uuid-page0-%d", i),
					"metadata_public": map[string]any{"user_id": fmt.Sprintf("other-%d", i)},
				})
			}
			_ = json.NewEncoder(w).Encode(batch)
		case "1":
			_, _ = w.Write([]byte(`[{"id":"uuid-target","metadata_public":{"user_id":"target"}}]`))
		default:
			t.Errorf("unexpected page param: %q (want 0 or 1)", page)
		}
	}))
	defer srv.Close()

	c := &KratosAdminClient{AdminURL: srv.URL}
	id, err := c.FindIdentityByUserID(context.Background(), "target")
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if id != "uuid-target" {
		t.Errorf("id = %q, want uuid-target", id)
	}
	if pages != 2 {
		t.Errorf("pages fetched = %d, want 2", pages)
	}
}

func TestKratosAdminClient_FindIdentityByUserID_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	c := &KratosAdminClient{AdminURL: srv.URL}
	_, err := c.FindIdentityByUserID(context.Background(), "ghost")
	if !errors.Is(err, ErrKratosIdentityNotFound) {
		t.Errorf("err = %v, want ErrKratosIdentityNotFound", err)
	}
}

func TestKratosAdminClient_FindIdentityByUserID_EmptyUserID(t *testing.T) {
	c := &KratosAdminClient{AdminURL: "http://unused"}
	_, err := c.FindIdentityByUserID(context.Background(), "")
	if err == nil {
		t.Fatalf("expected error for empty user_id")
	}
}

func TestKratosAdminClient_FindIdentityByUserID_StopsAt10kCap(t *testing.T) {
	// FindIdentityByUserID caps the PoC scan at 40 pages (pages 0..39 =
	// 10k identities at per_page=250). The cap matters because Kratos has
	// no server-side metadata filter; without it, a misconfigured
	// deployment (user_id never populated) would pull every identity each
	// call and either time out or hammer the admin endpoint. Pin the cap
	// so a future refactor cannot drop it silently.
	var requests int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		// Always return a full page so the < pageSize early-exit never fires.
		var batch []map[string]any
		for i := 0; i < 250; i++ {
			batch = append(batch, map[string]any{
				"id":              fmt.Sprintf("uuid-r%d-i%d", requests, i),
				"metadata_public": map[string]any{"user_id": "noone"},
			})
		}
		_ = json.NewEncoder(w).Encode(batch)
	}))
	defer srv.Close()

	c := &KratosAdminClient{AdminURL: srv.URL}
	_, err := c.FindIdentityByUserID(context.Background(), "missing")
	if !errors.Is(err, ErrKratosIdentityNotFound) {
		t.Fatalf("err = %v, want ErrKratosIdentityNotFound", err)
	}
	if requests != 40 {
		t.Errorf("requests = %d, want exactly 40 (10k identity PoC cap)", requests)
	}
}

func TestKratosAdminClient_UpdateIdentityPassword_RoundTrips(t *testing.T) {
	// Kratos has no first-class admin "set password"; the contract is
	// GET → mutate credentials.password.config → PUT. A prior regression
	// silently returned nil from this method, so this test pins both
	// requests and the mutated body.
	var (
		getCount int
		putBody  map[string]any
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const wantPath = "/admin/identities/id-1"
		if r.URL.Path != wantPath {
			t.Errorf("path = %s, want %s", r.URL.Path, wantPath)
		}
		switch r.Method {
		case http.MethodGet:
			getCount++
			_, _ = w.Write([]byte(`{
				"id":"id-1",
				"state":"active",
				"traits":{"email":"a@b.c"},
				"credentials":{"password":{"config":{"password":"old"}}}
			}`))
		case http.MethodPut:
			body, _ := io.ReadAll(r.Body)
			if err := json.Unmarshal(body, &putBody); err != nil {
				t.Fatalf("decode PUT body: %v", err)
			}
			_, _ = w.Write([]byte(`{"id":"id-1"}`))
		default:
			t.Errorf("unexpected method: %s", r.Method)
		}
	}))
	defer srv.Close()

	c := &KratosAdminClient{AdminURL: srv.URL}
	if err := c.UpdateIdentityPassword(context.Background(), "id-1", "NewPass-2026!"); err != nil {
		t.Fatalf("UpdateIdentityPassword: %v", err)
	}
	if getCount != 1 {
		t.Errorf("GET count = %d, want 1", getCount)
	}
	creds, _ := putBody["credentials"].(map[string]any)
	pwd, _ := creds["password"].(map[string]any)
	cfg, _ := pwd["config"].(map[string]any)
	if cfg["password"] != "NewPass-2026!" {
		t.Errorf("PUT credentials.password.config.password = %v, want NewPass-2026!", cfg["password"])
	}
	// Untouched fields must survive the round-trip — Kratos requires the
	// full identity body on PUT and rejects partial documents.
	if traits, _ := putBody["traits"].(map[string]any); traits["email"] != "a@b.c" {
		t.Errorf("PUT lost traits.email: %v", putBody["traits"])
	}
}

func TestKratosAdminClient_UpdateIdentityPassword_NotFoundOnGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := &KratosAdminClient{AdminURL: srv.URL}
	err := c.UpdateIdentityPassword(context.Background(), "missing", "Pass1234567!")
	if !errors.Is(err, ErrKratosIdentityNotFound) {
		t.Errorf("err = %v, want ErrKratosIdentityNotFound", err)
	}
}

func TestKratosAdminClient_UpdateIdentityPassword_RejectsEmptyPassword(t *testing.T) {
	c := &KratosAdminClient{AdminURL: "http://unused"}
	err := c.UpdateIdentityPassword(context.Background(), "id", "")
	if err == nil {
		t.Fatalf("expected error for empty password")
	}
}

func TestKratosAdminClient_SetIdentityState_ActiveAndInactive(t *testing.T) {
	cases := []struct {
		name      string
		active    bool
		wantState string
	}{
		{"activate", true, "active"},
		{"deactivate", false, "inactive"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var putBody map[string]any
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodGet:
					_, _ = w.Write([]byte(`{"id":"id-1","state":"active","traits":{}}`))
				case http.MethodPut:
					body, _ := io.ReadAll(r.Body)
					_ = json.Unmarshal(body, &putBody)
					_, _ = w.Write([]byte(`{"id":"id-1"}`))
				}
			}))
			defer srv.Close()

			c := &KratosAdminClient{AdminURL: srv.URL}
			if err := c.SetIdentityState(context.Background(), "id-1", tc.active); err != nil {
				t.Fatalf("SetIdentityState: %v", err)
			}
			if putBody["state"] != tc.wantState {
				t.Errorf("PUT state = %v, want %s", putBody["state"], tc.wantState)
			}
		})
	}
}

func TestKratosAdminClient_DeleteIdentity_Success(t *testing.T) {
	var called bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		if r.URL.Path != "/admin/identities/id-1" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := &KratosAdminClient{AdminURL: srv.URL}
	if err := c.DeleteIdentity(context.Background(), "id-1"); err != nil {
		t.Fatalf("DeleteIdentity: %v", err)
	}
	if !called {
		t.Errorf("server was not called")
	}
}

func TestKratosAdminClient_DeleteIdentity_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := &KratosAdminClient{AdminURL: srv.URL}
	err := c.DeleteIdentity(context.Background(), "ghost")
	if !errors.Is(err, ErrKratosIdentityNotFound) {
		t.Errorf("err = %v, want ErrKratosIdentityNotFound", err)
	}
}

func TestKratosAdminClient_DeleteIdentity_SurfacesUnexpectedStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"boom"}`))
	}))
	defer srv.Close()

	c := &KratosAdminClient{AdminURL: srv.URL}
	err := c.DeleteIdentity(context.Background(), "id-1")
	if err == nil {
		t.Fatalf("expected error on 500")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("err did not surface status: %v", err)
	}
}

func TestKratosAdminClient_DeleteIdentity_EscapesIdentityID(t *testing.T) {
	// Identity IDs are UUIDs in production, but the URL builder uses
	// url.PathEscape; a future caller might pass a value with slashes
	// (e.g. composite IDs in tests) and we want that to stay scoped to
	// a single path segment.
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := &KratosAdminClient{AdminURL: srv.URL}
	if err := c.DeleteIdentity(context.Background(), "id/with/slash"); err != nil {
		t.Fatalf("DeleteIdentity: %v", err)
	}
	if gotPath != "/admin/identities/id%2Fwith%2Fslash" {
		t.Errorf("escaped path = %s", gotPath)
	}
}
