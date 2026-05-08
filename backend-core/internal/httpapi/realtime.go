package httpapi

import (
	"net/http"
	"strings"
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
	clients map[*websocket.Conn]*realtimeClient
}

type realtimeClient struct {
	conn         *websocket.Conn
	subscription realtimeSubscription
	writeMu      sync.Mutex
}

type realtimeSubscription struct {
	types map[string]bool
}

func NewRealtimeHub() *RealtimeHub {
	return &RealtimeHub{clients: map[*websocket.Conn]*realtimeClient{}}
}

func (handler Handler) handleRealtimeWebSocket(c *gin.Context) {
	if handler.cfg.RealtimeHub == nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "realtime hub is not configured"})
		return
	}
	eventTypes := parseRealtimeTypes(c.Query("types"))
	if !devFallbackEnabled(c) && len(eventTypes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "types query is required"})
		return
	}
	if !devFallbackEnabled(c) {
		actorValue, _ := c.Get("devhub_actor_role")
		actorRole, _ := actorValue.(string)
		if actorRole == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthenticated", "error": "authenticated actor role is required"})
			return
		}
		for _, eventType := range eventTypes {
			resource, action, ok := realtimeEventPermission(eventType)
			if !ok {
				c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "unsupported realtime event type"})
				return
			}
			allowed, err := handler.cfg.PermissionCache.Allows(c.Request.Context(), actorRole, resource, action)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
				return
			}
			if !allowed {
				c.JSON(http.StatusForbidden, gin.H{"status": "forbidden", "error": "permission denied"})
				return
			}
		}
	}
	handler.cfg.RealtimeHub.HandleWebSocket(c, eventTypes)
}

func (h *RealtimeHub) HandleWebSocket(c *gin.Context, eventTypes []string) {
	conn, err := websocketUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	h.add(conn, eventTypes)
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

	clients := h.clientsFor(eventType)
	for _, client := range clients {
		if err := client.writeJSON(event); err != nil {
			h.removeClient(client)
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

func (h *RealtimeHub) add(conn *websocket.Conn, eventTypes []string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[conn] = &realtimeClient{
		conn:         conn,
		subscription: realtimeSubscription{types: realtimeTypeSet(eventTypes)},
	}
}

func (h *RealtimeHub) remove(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, conn)
	_ = conn.Close()
}

func (h *RealtimeHub) removeClient(client *realtimeClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if current, ok := h.clients[client.conn]; ok && current == client {
		delete(h.clients, client.conn)
	}
	_ = client.conn.Close()
}

func (h *RealtimeHub) clientsFor(eventType string) []*realtimeClient {
	h.mu.RLock()
	defer h.mu.RUnlock()
	clients := make([]*realtimeClient, 0, len(h.clients))
	for _, client := range h.clients {
		if client.subscription.allows(eventType) {
			clients = append(clients, client)
		}
	}
	return clients
}

func (c *realtimeClient) writeJSON(event realtimeEvent) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return c.conn.WriteJSON(event)
}

var websocketUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func prefixedEventID() string {
	return "evt_" + time.Now().UTC().Format("20060102150405.000000000")
}

func parseRealtimeTypes(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, part := range parts {
		eventType := strings.TrimSpace(part)
		if eventType == "" || seen[eventType] {
			continue
		}
		seen[eventType] = true
		out = append(out, eventType)
	}
	return out
}

func realtimeTypeSet(eventTypes []string) map[string]bool {
	if len(eventTypes) == 0 {
		return nil
	}
	out := make(map[string]bool, len(eventTypes))
	for _, eventType := range eventTypes {
		out[eventType] = true
	}
	return out
}

func (s realtimeSubscription) allows(eventType string) bool {
	return len(s.types) == 0 || s.types[eventType]
}

func realtimeEventPermission(eventType string) (domain.Resource, domain.Action, bool) {
	switch eventType {
	case "command.status.updated":
		return domain.ResourceInfrastructure, domain.ActionView, true
	case "risk.critical.created", "risk.updated":
		return domain.ResourceSecurity, domain.ActionView, true
	case "ci.run.updated", "ci.log.appended":
		return domain.ResourcePipelines, domain.ActionView, true
	case "infra.node.updated", "infra.edge.updated", "notification.created":
		return domain.ResourceInfrastructure, domain.ActionView, true
	default:
		return "", "", false
	}
}
