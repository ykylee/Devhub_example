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
