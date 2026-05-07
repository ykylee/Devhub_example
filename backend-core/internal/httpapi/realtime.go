package httpapi

import (
	"net/http"
	"sync"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type realtimeEvent struct {
	SchemaVersion string         `json:"schema_version"`
	Type          string         `json:"type"`
	EventID       string         `json:"event_id"`
	OccurredAt    time.Time      `json:"occurred_at"`
	Data          map[string]any `json:"data"`
}

type RealtimeHub struct {
	mu      sync.RWMutex
	clients map[*websocket.Conn]struct{}
}

func NewRealtimeHub() *RealtimeHub {
	return &RealtimeHub{clients: map[*websocket.Conn]struct{}{}}
}

func (h *RealtimeHub) HandleWebSocket(c *gin.Context) {
	conn, err := websocketUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	h.add(conn)
	defer h.remove(conn)

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (h *RealtimeHub) Publish(eventType string, data map[string]any) {
	event := realtimeEvent{
		SchemaVersion: "1",
		Type:          eventType,
		EventID:       prefixedEventID(),
		OccurredAt:    time.Now().UTC(),
		Data:          data,
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	for conn := range h.clients {
		if err := conn.WriteJSON(event); err != nil {
			_ = conn.Close()
		}
	}
}

func (h *RealtimeHub) PublishCommandStatus(command domain.Command) {
	h.Publish("command.status.updated", map[string]any{
		"command_id":     command.CommandID,
		"command_type":   command.CommandType,
		"target_type":    command.TargetType,
		"target_id":      command.TargetID,
		"action_type":    command.ActionType,
		"status":         command.Status,
		"actor_login":    command.ActorLogin,
		"result_payload": command.ResultPayload,
		"updated_at":     command.UpdatedAt,
	})
}

func (h *RealtimeHub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

func (h *RealtimeHub) add(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[conn] = struct{}{}
}

func (h *RealtimeHub) remove(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, conn)
	_ = conn.Close()
}

var websocketUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func prefixedEventID() string {
	return "evt_" + time.Now().UTC().Format("20060102150405.000000000")
}
