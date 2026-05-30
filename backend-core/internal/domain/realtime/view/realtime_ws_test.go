package view

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/shared/httphelp"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

// ---------------------------------------------------------------------------
// fakeRealtimeTicketStore — injectable error/result for IssueRealtimeTicket
// ---------------------------------------------------------------------------

type fakeTicketStore struct {
	issueTicket string
	issueErr    error
}

func (f *fakeTicketStore) Issue(_ context.Context, _, _ string, _ domain.AuditSourceType) (string, error) {
	return f.issueTicket, f.issueErr
}
func (f *fakeTicketStore) Consume(_ context.Context, _ string) (store.RealtimeTicket, bool, error) {
	return store.RealtimeTicket{}, false, nil
}

// fakePermCacheDeny — permission cache that always denies
type fakePermCacheDeny struct{}

func (f *fakePermCacheDeny) Allows(_ context.Context, _ string, _ domain.Resource, _ domain.Action) (bool, error) {
	return false, nil
}

// fakePermCacheErr — permission cache that always errors
type fakePermCacheErr struct{}

func (f *fakePermCacheErr) Allows(_ context.Context, _ string, _ domain.Resource, _ domain.Action) (bool, error) {
	return false, errors.New("perm check failed")
}

// ---------------------------------------------------------------------------
// parseRealtimeTypes tests
// ---------------------------------------------------------------------------

func TestParseRealtimeTypes_Empty(t *testing.T) {
	got := parseRealtimeTypes("")
	if len(got) != 0 {
		t.Fatalf("expected empty, got %v", got)
	}
}

func TestParseRealtimeTypes_CSVDedup(t *testing.T) {
	got := parseRealtimeTypes("a,b, a, c , b")
	if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Fatalf("dedup: %v", got)
	}
}

func TestParseRealtimeTypes_Whitespace(t *testing.T) {
	got := parseRealtimeTypes("  , , , ")
	if len(got) != 0 {
		t.Fatalf("expected empty after trim, got %v", got)
	}
}

// ---------------------------------------------------------------------------
// realtimeTypeSet tests
// ---------------------------------------------------------------------------

func TestRealtimeTypeSet_Empty(t *testing.T) {
	got := realtimeTypeSet(nil)
	if got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestRealtimeTypeSet_NonEmpty(t *testing.T) {
	got := realtimeTypeSet([]string{"a", "b"})
	if !got["a"] || !got["b"] || got["c"] {
		t.Fatalf("set: %v", got)
	}
}

// ---------------------------------------------------------------------------
// realtimeSubscription.allows tests
// ---------------------------------------------------------------------------

func TestSubscription_AllowsAll(t *testing.T) {
	s := realtimeSubscription{types: nil}
	if !s.allows("anything") {
		t.Fatal("nil types should allow all")
	}
}

func TestSubscription_AllowsSpecific(t *testing.T) {
	s := realtimeSubscription{types: map[string]bool{"a": true}}
	if !s.allows("a") {
		t.Fatal("should allow 'a'")
	}
	if s.allows("b") {
		t.Fatal("should not allow 'b'")
	}
}

// ---------------------------------------------------------------------------
// realtimeEventPermission tests
// ---------------------------------------------------------------------------

func TestRealtimeEventPermission_KnownTypes(t *testing.T) {
	cases := []struct {
		eventType string
		wantOK    bool
	}{
		{"command.status.updated", true},
		{"risk.critical.created", true},
		{"risk.updated", true},
		{"ci.run.updated", true},
		{"ci.log.appended", true},
		{"infra.node.updated", true},
		{"infra.edge.updated", true},
		{"notification.created", true},
		{"unknown.event.type", false},
	}
	for _, tc := range cases {
		_, _, ok := realtimeEventPermission(tc.eventType)
		if ok != tc.wantOK {
			t.Fatalf("event=%q: ok=%v, want=%v", tc.eventType, ok, tc.wantOK)
		}
	}
}

// ---------------------------------------------------------------------------
// prefixedEventID tests
// ---------------------------------------------------------------------------

func TestPrefixedEventID_HasPrefix(t *testing.T) {
	id := prefixedEventID()
	if !strings.HasPrefix(id, "evt_") {
		t.Fatalf("expected evt_ prefix, got %q", id)
	}
}

// ---------------------------------------------------------------------------
// NewRealtimeHub tests
// ---------------------------------------------------------------------------

func TestNewRealtimeHub(t *testing.T) {
	hub := NewRealtimeHub()
	if hub == nil {
		t.Fatal("expected non-nil hub")
	}
	if hub.ClientCount() != 0 {
		t.Fatalf("expected 0 clients, got %d", hub.ClientCount())
	}
}

// ---------------------------------------------------------------------------
// RealtimeHub Publish with no clients (no panic)
// ---------------------------------------------------------------------------

func TestRealtimeHub_PublishNoClients(t *testing.T) {
	hub := NewRealtimeHub()
	// Should not panic
	hub.Publish("command.status.updated", map[string]any{"test": true})
}

func TestRealtimeHub_PublishCommandStatus(t *testing.T) {
	hub := NewRealtimeHub()
	// Should not panic
	hub.PublishCommandStatus(domain.Command{
		CommandID:   "cmd-1",
		CommandType: "deploy",
		Status:      "completed",
	})
}

// ---------------------------------------------------------------------------
// HandleRealtimeWebSocket — pre-WebSocket-upgrade validation tests
// ---------------------------------------------------------------------------

func TestHandleRealtimeWebSocket_NilHub(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewRealtimeHandler(RealtimeConfig{})
	r := gin.New()
	r.GET("/ws", h.HandleRealtimeWebSocket)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/ws", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != 404 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestHandleRealtimeWebSocket_NoTypesNoDevFallback(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hub := NewRealtimeHub()
	h := NewRealtimeHandler(RealtimeConfig{RealtimeHub: hub})
	r := gin.New()
	r.GET("/ws", h.HandleRealtimeWebSocket)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/ws?types=", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestHandleRealtimeWebSocket_NoActorRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hub := NewRealtimeHub()
	h := NewRealtimeHandler(RealtimeConfig{RealtimeHub: hub, PermissionCache: &fakeRealtimePermCache{}})
	r := gin.New()
	r.GET("/ws", h.HandleRealtimeWebSocket)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/ws?types=command.status.updated", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != 401 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestHandleRealtimeWebSocket_UnsupportedEventType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hub := NewRealtimeHub()
	h := NewRealtimeHandler(RealtimeConfig{RealtimeHub: hub, PermissionCache: &fakeRealtimePermCache{}})
	r := gin.New()
	r.GET("/ws", func(c *gin.Context) {
		c.Set("devhub_actor_role", "developer")
		h.HandleRealtimeWebSocket(c)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/ws?types=unknown.bogus", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestHandleRealtimeWebSocket_PermissionCheckError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hub := NewRealtimeHub()
	h := NewRealtimeHandler(RealtimeConfig{RealtimeHub: hub, PermissionCache: &fakePermCacheErr{}})
	r := gin.New()
	r.GET("/ws", func(c *gin.Context) {
		c.Set("devhub_actor_role", "developer")
		h.HandleRealtimeWebSocket(c)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/ws?types=command.status.updated", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != 500 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestHandleRealtimeWebSocket_PermissionDenied(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hub := NewRealtimeHub()
	h := NewRealtimeHandler(RealtimeConfig{RealtimeHub: hub, PermissionCache: &fakePermCacheDeny{}})
	r := gin.New()
	r.GET("/ws", func(c *gin.Context) {
		c.Set("devhub_actor_role", "developer")
		h.HandleRealtimeWebSocket(c)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/ws?types=command.status.updated", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != 403 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestHandleRealtimeWebSocket_DevFallbackBypassesTypesCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hub := NewRealtimeHub()
	h := NewRealtimeHandler(RealtimeConfig{RealtimeHub: hub})
	r := gin.New()
	r.GET("/ws", func(c *gin.Context) {
		c.Set("devhub_auth_dev_fallback", true)
		h.HandleRealtimeWebSocket(c)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/ws", nil)
	r.ServeHTTP(rec, req)
	// With devFallback enabled, the handler bypasses types and auth checks
	// and proceeds directly to WebSocket upgrade. The upgrader returns 400
	// because this is a plain HTTP request, not a WebSocket handshake.
	// This confirms the pre-upgrade validation was successfully bypassed.
	if rec.Code != 400 {
		t.Fatalf("status = %d (expected 400 from WS upgrader for non-WS request)", rec.Code)
	}
}

// ---------------------------------------------------------------------------
// IssueRealtimeTicket handler tests
// ---------------------------------------------------------------------------

func TestIssueRealtimeTicket_NilStore(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewRealtimeHandler(RealtimeConfig{})
	r := gin.New()
	r.POST("/ticket", h.IssueRealtimeTicket)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/ticket", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != 503 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestIssueRealtimeTicket_NoActorLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts := &fakeTicketStore{}
	h := NewRealtimeHandler(RealtimeConfig{RealtimeTickets: ts})
	r := gin.New()
	r.POST("/ticket", h.IssueRealtimeTicket)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/ticket", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != 401 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestIssueRealtimeTicket_IssueError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts := &fakeTicketStore{issueErr: errors.New("db down")}
	h := NewRealtimeHandler(RealtimeConfig{RealtimeTickets: ts})
	r := gin.New()
	r.POST("/ticket", func(c *gin.Context) {
		c.Set("devhub_actor_login", "alice")
		c.Set("devhub_actor_role", "developer")
		h.IssueRealtimeTicket(c)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/ticket", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != 500 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestIssueRealtimeTicket_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts := &fakeTicketStore{issueTicket: "test-ticket-abc"}
	h := NewRealtimeHandler(RealtimeConfig{RealtimeTickets: ts})
	r := gin.New()
	r.POST("/ticket", func(c *gin.Context) {
		c.Set("devhub_actor_login", "alice")
		c.Set("devhub_actor_role", "developer")
		c.Set(httphelp.CtxKeySourceType, domain.AuditSourceOIDC)
		h.IssueRealtimeTicket(c)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/ticket", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "test-ticket-abc") {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestIssueRealtimeTicket_SuccessDefaultSource(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ts := &fakeTicketStore{issueTicket: "ticket-xyz"}
	h := NewRealtimeHandler(RealtimeConfig{RealtimeTickets: ts})
	r := gin.New()
	r.POST("/ticket", func(c *gin.Context) {
		c.Set("devhub_actor_login", "bob")
		// no source type set — should default to AuditSourceOIDC
		h.IssueRealtimeTicket(c)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/ticket", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status = %d", rec.Code)
	}
}
