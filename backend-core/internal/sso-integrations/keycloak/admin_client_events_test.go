package keycloak

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// fakeKeycloakServer — token endpoint + events / admin-events handler 를 1개 process 안에서 제공.
// codex review #9 정정 정합 (dateFrom + /admin-events 경로) 검증용.
func newFakeKeycloakServer(t *testing.T, h func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/realms/test/protocol/openid-connect/token", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "stub-bearer"})
	})
	mux.HandleFunc("/", h)
	return httptest.NewServer(mux)
}

func TestKeycloakAdminClient_ListUserEvents_QueryAndDecode(t *testing.T) {
	var capturedPath, capturedQuery, capturedAuth string
	srv := newFakeKeycloakServer(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/protocol/openid-connect/token") {
			return // already handled by /realms/test/... mux entry above
		}
		capturedPath = r.URL.Path
		capturedQuery = r.URL.RawQuery
		capturedAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{
				"time":      int64(1747641600000),
				"type":      "LOGIN",
				"realmId":   "test",
				"clientId":  "devhub-frontend",
				"userId":    "u1",
				"ipAddress": "10.0.0.1",
				"details":   map[string]string{"sessionId": "sess-1"},
			},
			{
				"time":   int64(1747641700000),
				"type":   "LOGIN_ERROR",
				"userId": "u2",
				"error":  "invalid_grant",
			},
		})
	})
	defer srv.Close()

	c := &KeycloakAdminClient{
		AdminURL:     srv.URL,
		Realm:        "test",
		ClientID:     "devhub-admin-cli",
		ClientSecret: "secret",
		IssuerURL:    srv.URL + "/realms/test",
	}
	dateFrom := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	events, err := c.ListUserEvents(context.Background(), dateFrom, 100)
	if err != nil {
		t.Fatalf("ListUserEvents: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("events = %d; want 2", len(events))
	}
	if events[0].Type != "LOGIN" || events[0].UserID != "u1" {
		t.Fatalf("events[0] = %+v", events[0])
	}
	if events[1].Type != "LOGIN_ERROR" || events[1].Error != "invalid_grant" {
		t.Fatalf("events[1] = %+v", events[1])
	}

	if !strings.Contains(capturedPath, "/admin/realms/test/events") {
		t.Fatalf("path = %q; want admin/realms/test/events", capturedPath)
	}
	if !strings.Contains(capturedQuery, "dateFrom=2026-05-19T12") {
		t.Fatalf("query = %q; want dateFrom=2026-05-19T12...", capturedQuery)
	}
	if !strings.Contains(capturedQuery, "max=100") {
		t.Fatalf("query = %q; want max=100", capturedQuery)
	}
	if capturedAuth != "Bearer stub-bearer" {
		t.Fatalf("auth = %q; want Bearer stub-bearer", capturedAuth)
	}
}

func TestKeycloakAdminClient_ListAdminEvents_PathIsAdminEvents(t *testing.T) {
	var capturedPath string
	srv := newFakeKeycloakServer(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/protocol/openid-connect/token") {
			return
		}
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{
				"time":          int64(1747641600000),
				"realmId":       "test",
				"operationType": "CREATE",
				"resourceType":  "USER",
				"resourcePath":  "users/u1",
			},
		})
	})
	defer srv.Close()

	c := &KeycloakAdminClient{
		AdminURL:     srv.URL,
		Realm:        "test",
		ClientID:     "devhub-admin-cli",
		ClientSecret: "secret",
		IssuerURL:    srv.URL + "/realms/test",
	}
	dateFrom := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	events, err := c.ListAdminEvents(context.Background(), dateFrom, 50)
	if err != nil {
		t.Fatalf("ListAdminEvents: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("events = %d; want 1", len(events))
	}
	if events[0].ResourceType != "USER" || events[0].OperationType != "CREATE" {
		t.Fatalf("events[0] = %+v", events[0])
	}
	// codex review #9 정정 정합: path 는 /admin-events 여야 함 (NOT /events/admin).
	if !strings.HasSuffix(capturedPath, "/admin/realms/test/admin-events") {
		t.Fatalf("path = %q; want suffix /admin/realms/test/admin-events", capturedPath)
	}
}

func TestKeycloakAdminClient_ListUserEvents_DefaultMax(t *testing.T) {
	var capturedQuery string
	srv := newFakeKeycloakServer(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/protocol/openid-connect/token") {
			return
		}
		capturedQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
	})
	defer srv.Close()

	c := &KeycloakAdminClient{
		AdminURL:     srv.URL,
		Realm:        "test",
		ClientID:     "devhub-admin-cli",
		ClientSecret: "secret",
		IssuerURL:    srv.URL + "/realms/test",
	}
	_, err := c.ListUserEvents(context.Background(), time.Time{}, 0)
	if err != nil {
		t.Fatalf("ListUserEvents: %v", err)
	}
	if !strings.Contains(capturedQuery, "max=500") {
		t.Fatalf("query = %q; want max=500 default", capturedQuery)
	}
	// zero-value dateFrom 일 때 query 에 dateFrom 미포함.
	if strings.Contains(capturedQuery, "dateFrom=") {
		t.Fatalf("query = %q; should NOT include dateFrom when zero", capturedQuery)
	}
}
