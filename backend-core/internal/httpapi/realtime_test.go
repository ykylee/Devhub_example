package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func TestRealtimeHubPublishesCommandStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hub := NewRealtimeHub()
	server := httptest.NewServer(NewRouter(RouterConfig{RealtimeHub: hub}))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/v1/realtime/ws"
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()
	waitForRealtimeClient(t, hub)

	updatedAt := time.Date(2026, 5, 6, 10, 0, 0, 0, time.UTC)
	hub.PublishCommandStatus(domain.Command{
		CommandID:     "cmd_test",
		CommandType:   "service_action",
		TargetType:    "service",
		TargetID:      "backend-core",
		ActionType:    "restart",
		Status:        "succeeded",
		ActorLogin:    "yklee",
		ResultPayload: map[string]any{"dry_run": true},
		UpdatedAt:     updatedAt,
	})

	if err := conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}
	_, payload, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read websocket message: %v", err)
	}
	var event struct {
		SchemaVersion string         `json:"schema_version"`
		Type          string         `json:"type"`
		Data          map[string]any `json:"data"`
	}
	if err := json.Unmarshal(payload, &event); err != nil {
		t.Fatalf("decode websocket event: %v", err)
	}
	if event.SchemaVersion != "1" || event.Type != "command.status.updated" {
		t.Fatalf("unexpected event envelope: %+v", event)
	}
	if event.Data["command_id"] != "cmd_test" || event.Data["status"] != "succeeded" {
		t.Fatalf("unexpected command payload: %+v", event.Data)
	}
}

func TestRealtimeHubFiltersBySubscribedTypes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hub := NewRealtimeHub()
	server := httptest.NewServer(NewRouter(RouterConfig{RealtimeHub: hub}))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/v1/realtime/ws?types=risk.updated"
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()
	waitForRealtimeClient(t, hub)

	hub.Publish("command.status.updated", map[string]any{"command_id": "cmd_hidden"})
	hub.Publish("risk.updated", map[string]any{"risk_id": "risk_visible"})
	if err := conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}
	_, payload, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read risk event: %v", err)
	}
	var event struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(payload, &event); err != nil {
		t.Fatalf("decode event: %v", err)
	}
	if event.Type != "risk.updated" {
		t.Fatalf("expected risk event, got %+v", event)
	}
}

func TestRealtimeWebSocketRequiresTypesWhenRBACEnabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hub := NewRealtimeHub()
	server := httptest.NewServer(NewRouter(RouterConfig{RealtimeHub: hub, RBACPolicyStore: &memoryRBACPolicyStore{}}))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/v1/realtime/ws"
	_, resp, err := websocket.DefaultDialer.Dial(url, nil)
	if err == nil {
		t.Fatalf("expected websocket dial to fail without types")
	}
	if resp == nil || resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 response, got resp=%v err=%v", resp, err)
	}
}

func TestRealtimeWebSocketChecksRBACPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hub := NewRealtimeHub()
	server := httptest.NewServer(NewRouter(RouterConfig{RealtimeHub: hub, RBACPolicyStore: &memoryRBACPolicyStore{}}))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/v1/realtime/ws?types=command.status.updated"
	header := http.Header{}
	header.Set("X-Devhub-Role", "developer")
	_, resp, err := websocket.DefaultDialer.Dial(url, header)
	if err == nil {
		t.Fatalf("expected websocket dial to fail for insufficient command permission")
	}
	if resp == nil || resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 response, got resp=%v err=%v", resp, err)
	}

	header.Set("X-Devhub-Role", "manager")
	conn, resp, err := websocket.DefaultDialer.Dial(url, header)
	if err != nil {
		t.Fatalf("expected manager websocket dial to succeed, resp=%v err=%v", resp, err)
	}
	defer conn.Close()
}

func TestRealtimeRouteIsAbsentWithoutHub(t *testing.T) {
	gin.SetMode(gin.TestMode)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/realtime/ws", nil)
	rec := httptest.NewRecorder()

	NewRouter(RouterConfig{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected realtime route to be disabled without hub, got %d", rec.Code)
	}
}

func waitForRealtimeClient(t *testing.T, hub *RealtimeHub) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if hub.ClientCount() == 1 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("websocket client was not registered")
}
