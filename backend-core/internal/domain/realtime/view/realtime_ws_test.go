package view

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/shared/httphelp"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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

// ---------------------------------------------------------------------------
// WebSocket integration tests — httptest.Server + gorilla/websocket.Dial
// ---------------------------------------------------------------------------

// wsTestHarness sets up a Gin router + httptest.Server with DevFallback
// enabled to bypass auth checks for the WS endpoint.
func wsTestHarness(t *testing.T, handler *RealtimeHandler) (*httptest.Server, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/ws", func(c *gin.Context) {
		c.Set("devhub_auth_dev_fallback", true)
		handler.HandleRealtimeWebSocket(c)
	})
	ts := httptest.NewServer(r)
	t.Cleanup(ts.Close)
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	return ts, wsURL
}

func TestRealtimeHub_WebSocket_PublishSubscribe(t *testing.T) {
	hub := NewRealtimeHub()
	h := NewRealtimeHandler(RealtimeConfig{RealtimeHub: hub})
	_, wsURL := wsTestHarness(t, h)

	dialer := &websocket.Dialer{}
	conn, _, err := dialer.Dial(wsURL+"?types=command.status.updated", nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	// give the goroutine time to register the client
	time.Sleep(50 * time.Millisecond)

	hub.Publish("command.status.updated", map[string]any{"key": "val"})

	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read message: %v", err)
	}

	var ev struct {
		Type    string         `json:"type"`
		EventID string         `json:"event_id"`
		Data    map[string]any `json:"data"`
	}
	if err := json.Unmarshal(msg, &ev); err != nil {
		t.Fatalf("unmarshal: %v\nbody: %s", err, string(msg))
	}
	if ev.Type != "command.status.updated" {
		t.Fatalf("event type = %q", ev.Type)
	}
	if ev.Data["key"] != "val" {
		t.Fatalf("data = %v", ev.Data)
	}
	if !strings.HasPrefix(ev.EventID, "evt_") {
		t.Fatalf("event_id = %q", ev.EventID)
	}
}

func TestRealtimeHub_WebSocket_SubscriptionFilter(t *testing.T) {
	hub := NewRealtimeHub()
	h := NewRealtimeHandler(RealtimeConfig{RealtimeHub: hub})
	_, wsURL := wsTestHarness(t, h)

	// subscribe only to "events.a"
	dialer := &websocket.Dialer{}
	conn, _, err := dialer.Dial(wsURL+"?types=events.a", nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	time.Sleep(50 * time.Millisecond)

	// publish a non-matching event — should be filtered by subscription
	hub.Publish("events.b", map[string]any{"nope": true})
	time.Sleep(100 * time.Millisecond)

	// publish a matching event
	hub.Publish("events.a", map[string]any{"match": true})

	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("should receive events.a: %v", err)
	}
	var ev struct {
		Type string         `json:"type"`
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(msg, &ev); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if ev.Type != "events.a" {
		t.Fatalf("expected type 'events.a', got %q — subscription filter leaked events.b", ev.Type)
	}
	if ev.Data["match"] != true {
		t.Fatalf("data = %v", ev.Data)
	}
}

func TestRealtimeHub_WebSocket_MultipleClients(t *testing.T) {
	hub := NewRealtimeHub()
	h := NewRealtimeHandler(RealtimeConfig{RealtimeHub: hub})
	_, wsURL := wsTestHarness(t, h)

	dialer := &websocket.Dialer{}
	conn1, _, err := dialer.Dial(wsURL+"?types=test.event", nil)
	if err != nil {
		t.Fatalf("dial conn1: %v", err)
	}
	t.Cleanup(func() { conn1.Close() })

	conn2, _, err := dialer.Dial(wsURL+"?types=test.event", nil)
	if err != nil {
		t.Fatalf("dial conn2: %v", err)
	}
	t.Cleanup(func() { conn2.Close() })
	time.Sleep(80 * time.Millisecond)

	if hub.ClientCount() != 2 {
		t.Fatalf("expected 2 clients, got %d", hub.ClientCount())
	}

	hub.Publish("test.event", map[string]any{"broadcast": true})

	// both clients should receive the event
	for i, conn := range []*websocket.Conn{conn1, conn2} {
		_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		_, msg, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("client %d read: %v", i+1, err)
		}
		var ev struct {
			Type string         `json:"type"`
			Data map[string]any `json:"data"`
		}
		if err := json.Unmarshal(msg, &ev); err != nil {
			t.Fatalf("client %d unmarshal: %v", i+1, err)
		}
		if ev.Type != "test.event" {
			t.Fatalf("client %d type = %q", i+1, ev.Type)
		}
	}
}

func TestRealtimeHub_WebSocket_ClientDisconnect(t *testing.T) {
	hub := NewRealtimeHub()
	h := NewRealtimeHandler(RealtimeConfig{RealtimeHub: hub})
	_, wsURL := wsTestHarness(t, h)

	dialer := &websocket.Dialer{}
	conn, _, err := dialer.Dial(wsURL+"?types=test.event", nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	time.Sleep(50 * time.Millisecond)

	if hub.ClientCount() != 1 {
		t.Fatalf("expected 1 client after connect, got %d", hub.ClientCount())
	}

	// disconnect the client — this causes handleWebSocket read loop to exit
	// and defer h.remove(conn) to fire.
	conn.Close()
	time.Sleep(100 * time.Millisecond)

	if hub.ClientCount() != 0 {
		t.Fatalf("expected 0 clients after disconnect, got %d", hub.ClientCount())
	}
}

func TestRealtimeHub_WebSocket_PublishToSubscribeAll(t *testing.T) {
	// When no types query param is provided with DevFallback,
	// subscription allows all events.
	hub := NewRealtimeHub()
	h := NewRealtimeHandler(RealtimeConfig{RealtimeHub: hub})
	_, wsURL := wsTestHarness(t, h)

	dialer := &websocket.Dialer{}
	conn, _, err := dialer.Dial(wsURL, nil) // no ?types=
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	time.Sleep(50 * time.Millisecond)

	hub.Publish("any.event.type", map[string]any{"catchall": true})

	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var ev struct {
		Type string         `json:"type"`
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(msg, &ev); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if ev.Type != "any.event.type" {
		t.Fatalf("type = %q", ev.Type)
	}
}

func TestRealtimeHub_WebSocket_PublishClientWriteFail(t *testing.T) {
	// When Publish encounters a client whose write fails (e.g. disconnected),
	// it should call removeClient to clean up the stale entry.
	hub := NewRealtimeHub()
	h := NewRealtimeHandler(RealtimeConfig{RealtimeHub: hub})
	_, wsURL := wsTestHarness(t, h)

	conn, _, err := (&websocket.Dialer{}).Dial(wsURL+"?types=test.event", nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	time.Sleep(50 * time.Millisecond)
	if hub.ClientCount() != 1 {
		t.Fatalf("expected 1 client, got %d", hub.ClientCount())
	}

	// close the client connection, then publish immediately — there's a race
	// between HandleWebSocket's deferred h.remove(conn) and our Publish().
	// If Publish wins, writeJSON fails and removeClient cleans up.
	conn.Close()
	// publish without delay to hit the window before HandleWebSocket cleanup
	hub.Publish("test.event", map[string]any{"after": "disconnect"})

	// in either case (removeClient or normal remove), the client should be gone
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if hub.ClientCount() == 0 {
			return // success
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("client not removed after 500ms — count = %d", hub.ClientCount())
}


