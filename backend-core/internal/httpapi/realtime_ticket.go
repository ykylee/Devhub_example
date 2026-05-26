package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/gin-gonic/gin"
)

// realtimeTicketTTL — WebSocket handshake 직전 사용을 위한 단기 ticket 수명.
// ADR-0024 §3.2 (ticket pattern). 60 초 = network jitter + retry 1회 여유.
const realtimeTicketTTL = 60 * time.Second

type realtimeTicket struct {
	actorLogin string
	actorRole  string
	sourceType domain.AuditSourceType
	expiresAt  time.Time
}

// RealtimeTicketStore — single-instance in-memory ticket store.
// ADR-0024 §6 carve 1 (ticket pattern) minimal 구현.
// multi-instance (sticky session 없이 horizontal scale) 의 경우 Redis/PG
// 백킹 store 로 마이그레이션 필요 — 후속 carve.
type RealtimeTicketStore struct {
	mu      sync.Mutex
	tickets map[string]*realtimeTicket
}

func NewRealtimeTicketStore() *RealtimeTicketStore {
	return &RealtimeTicketStore{tickets: make(map[string]*realtimeTicket)}
}

func (s *RealtimeTicketStore) issue(actorLogin, actorRole string, sourceType domain.AuditSourceType) (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	ticket := hex.EncodeToString(buf)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sweepLocked(time.Now())
	s.tickets[ticket] = &realtimeTicket{
		actorLogin: actorLogin,
		actorRole:  actorRole,
		sourceType: sourceType,
		expiresAt:  time.Now().Add(realtimeTicketTTL),
	}
	return ticket, nil
}

// consume returns the ticket entry and removes it (single-use).
// Expired entries are also removed but reported as miss.
func (s *RealtimeTicketStore) consume(ticket string) (*realtimeTicket, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	s.sweepLocked(now)
	entry, ok := s.tickets[ticket]
	if !ok {
		return nil, false
	}
	delete(s.tickets, ticket)
	if entry.expiresAt.Before(now) {
		return nil, false
	}
	return entry, true
}

// sweepLocked removes expired entries. Caller holds s.mu.
// Opportunistic cleanup — TTL 60s + low expected concurrency (one ticket per
// WS connect) keeps the map size bounded without a background goroutine.
func (s *RealtimeTicketStore) sweepLocked(now time.Time) {
	for k, v := range s.tickets {
		if v.expiresAt.Before(now) {
			delete(s.tickets, k)
		}
	}
}

// issueRealtimeTicket — POST /api/v1/realtime/ticket.
// authenticateActor middleware 가 actor 를 c.Set 한 후 호출. Bearer auth 통과
// = 인증된 actor 로 발급 권한 인정 (RBAC bypass). ADR-0024 §3.2.
func (h Handler) issueRealtimeTicket(c *gin.Context) {
	if h.cfg.RealtimeTickets == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "realtime ticket store not configured",
		})
		return
	}
	actorLoginRaw, _ := c.Get("devhub_actor_login")
	login, _ := actorLoginRaw.(string)
	if login == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "unauthenticated",
			"error":  "authenticated actor required",
		})
		return
	}
	actorRoleRaw, _ := c.Get("devhub_actor_role")
	role, _ := actorRoleRaw.(string)

	sourceTypeRaw, _ := c.Get(ctxKeySourceType)
	source, _ := sourceTypeRaw.(domain.AuditSourceType)
	if source == "" {
		source = domain.AuditSourceOIDC
	}

	ticket, err := h.cfg.RealtimeTickets.issue(login, role, source)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "failed",
			"error":  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":             "ok",
		"ticket":             ticket,
		"expires_in_seconds": int(realtimeTicketTTL.Seconds()),
	})
}
