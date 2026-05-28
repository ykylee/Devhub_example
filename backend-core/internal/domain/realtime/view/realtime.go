package view

import (
	"github.com/devhub/backend-core/internal/shared/httphelp"
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

func (h *RealtimeHandler) HandleRealtimeWebSocket(c *gin.Context) {
	if h.cfg.RealtimeHub == nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "realtime hub is not configured"})
		return
	}
	eventTypes := parseRealtimeTypes(c.Query("types"))
	if !httphelp.DevFallbackEnabled(c) && len(eventTypes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "types query is required"})
		return
	}
	if !httphelp.DevFallbackEnabled(c) {
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
			allowed, err := h.cfg.PermissionCache.Allows(c.Request.Context(), actorRole, resource, action)
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
	h.cfg.RealtimeHub.HandleWebSocket(c, eventTypes)
}

// serverHeartbeatInterval — 서버측 WebSocket ping 주기. nginx/proxy/CDN 의 idle
// timeout (보통 60s) 보다 짧게 두 방향(client↔upstream) 모두 트래픽을 유지한다.
// (#392 codex P1 정합 — client-only heartbeat 는 upstream→client 방향의 nginx
//  proxy_read_timeout 을 갱신하지 못하므로 서버측에서 ping 프레임을 추가 발행.)
const serverHeartbeatInterval = 25 * time.Second

// serverHeartbeatWriteTimeout — ping frame 의 쓰기 deadline.
const serverHeartbeatWriteTimeout = 10 * time.Second

func (h *RealtimeHub) HandleWebSocket(c *gin.Context, eventTypes []string) {
	conn, err := websocketUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	client := h.add(conn, eventTypes)
	defer h.remove(conn)

	// 서버측 heartbeat — WebSocket ping 프레임을 주기적으로 발행. 브라우저가
	// 자동으로 pong 응답하므로 양방향에 트래픽이 흘러 nginx 등 중간 프록시의
	// idle timeout 을 갱신한다. client.writePing() 은 writeMu 로 직렬화되어
	// Publish 의 writeJSON 과 동시 접근해도 안전.
	heartbeatDone := make(chan struct{})
	go func() {
		ticker := time.NewTicker(serverHeartbeatInterval)
		defer ticker.Stop()
		for {
			select {
			case <-heartbeatDone:
				return
			case <-ticker.C:
				if err := client.writePing(); err != nil {
					// 쓰기 실패 = 연결 단절 신호. read 루프가 자체적으로 종료하므로
					// 여기선 단순 return — heartbeatDone close 는 메인 루프가 담당.
					return
				}
			}
		}
	}()

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			close(heartbeatDone)
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

func (h *RealtimeHub) add(conn *websocket.Conn, eventTypes []string) *realtimeClient {
	h.mu.Lock()
	defer h.mu.Unlock()
	client := &realtimeClient{
		conn:         conn,
		subscription: realtimeSubscription{types: realtimeTypeSet(eventTypes)},
	}
	h.clients[conn] = client
	return client
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

// writePing — WebSocket ping 프레임을 발행. writeJSON 과 동일한 writeMu 로
// 직렬화 (Gorilla websocket Conn 의 동시 write 미지원 정합). 브라우저가 자동으로
// pong 으로 응답하므로 양방향 트래픽이 흘러 중간 프록시(nginx 등) 의 idle
// timeout (보통 60s) 을 갱신한다.
func (c *realtimeClient) writePing() error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return c.conn.WriteControl(
		websocket.PingMessage,
		[]byte{},
		time.Now().Add(serverHeartbeatWriteTimeout),
	)
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
