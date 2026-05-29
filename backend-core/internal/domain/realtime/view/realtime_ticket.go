package view

import (
	"github.com/devhub/backend-core/internal/shared/httphelp"
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

// realtimeTicketTTL — WebSocket handshake 직전 사용을 위한 단기 ticket 수명.
// ADR-0024 §3.2 (ticket pattern). 60 초 = network jitter + retry 1회 여유.
const realtimeTicketTTL = 60 * time.Second

// RealtimeTicketStore — issue/consume 추상화.
type RealtimeTicketStore interface {
	Issue(ctx context.Context, actorLogin, actorRole string, sourceType domain.AuditSourceType) (string, error)
	Consume(ctx context.Context, ticket string) (store.RealtimeTicket, bool, error)
}

func newRealtimeTicket() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// InMemoryRealtimeTicketStore — single-instance in-memory ticket store.
type InMemoryRealtimeTicketStore struct {
	mu      sync.Mutex
	tickets map[string]store.RealtimeTicket
}

func NewInMemoryRealtimeTicketStore() *InMemoryRealtimeTicketStore {
	return &InMemoryRealtimeTicketStore{tickets: make(map[string]store.RealtimeTicket)}
}

func (s *InMemoryRealtimeTicketStore) Issue(_ context.Context, actorLogin, actorRole string, sourceType domain.AuditSourceType) (string, error) {
	ticket, err := newRealtimeTicket()
	if err != nil {
		return "", err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sweepLocked(time.Now())
	s.tickets[ticket] = store.RealtimeTicket{
		Ticket:     ticket,
		ActorLogin: actorLogin,
		ActorRole:  actorRole,
		SourceType: string(sourceType),
		ExpiresAt:  time.Now().Add(realtimeTicketTTL),
	}
	return ticket, nil
}

func (s *InMemoryRealtimeTicketStore) Consume(_ context.Context, ticket string) (store.RealtimeTicket, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	s.sweepLocked(now)
	entry, ok := s.tickets[ticket]
	if !ok {
		return store.RealtimeTicket{}, false, nil
	}
	delete(s.tickets, ticket)
	if entry.ExpiresAt.Before(now) {
		return store.RealtimeTicket{}, false, nil
	}
	return entry, true, nil
}

// realtimeTicketDB — PG 백킹 ticket store 가 의존하는 store 메서드.
type realtimeTicketDB interface {
	InsertRealtimeTicket(ctx context.Context, t store.RealtimeTicket) error
	ConsumeRealtimeTicket(ctx context.Context, ticket string) (store.RealtimeTicket, bool, error)
	DeleteExpiredRealtimeTickets(ctx context.Context) (int64, error)
}

// DBRealtimeTicketStore — multi-instance 안전 PG 백킹 ticket store (ADR-0024 §6 carve 6).
type DBRealtimeTicketStore struct {
	db realtimeTicketDB
}

func NewDBRealtimeTicketStore(db realtimeTicketDB) *DBRealtimeTicketStore {
	return &DBRealtimeTicketStore{db: db}
}

func NewRealtimeTicketStoreFor(pg *store.PostgresStore) RealtimeTicketStore {
	if pg == nil {
		return NewInMemoryRealtimeTicketStore()
	}
	return NewDBRealtimeTicketStore(pg)
}

func (s *DBRealtimeTicketStore) Issue(ctx context.Context, actorLogin, actorRole string, sourceType domain.AuditSourceType) (string, error) {
	ticket, err := newRealtimeTicket()
	if err != nil {
		return "", err
	}
	_, _ = s.db.DeleteExpiredRealtimeTickets(ctx)
	if err := s.db.InsertRealtimeTicket(ctx, store.RealtimeTicket{
		Ticket:     ticket,
		ActorLogin: actorLogin,
		ActorRole:  actorRole,
		SourceType: string(sourceType),
		ExpiresAt:  time.Now().Add(realtimeTicketTTL),
	}); err != nil {
		return "", err
	}
	return ticket, nil
}

func (s *DBRealtimeTicketStore) Consume(ctx context.Context, ticket string) (store.RealtimeTicket, bool, error) {
	row, ok, err := s.db.ConsumeRealtimeTicket(ctx, ticket)
	if err != nil {
		return store.RealtimeTicket{}, false, err
	}
	if !ok {
		return store.RealtimeTicket{}, false, nil
	}
	return row, true, nil
}

func (s *InMemoryRealtimeTicketStore) sweepLocked(now time.Time) {
	for k, v := range s.tickets {
		if v.ExpiresAt.Before(now) {
			delete(s.tickets, k)
		}
	}
}

// IssueRealtimeTicket — POST /api/v1/realtime/ticket.
func (h *RealtimeHandler) IssueRealtimeTicket(c *gin.Context) {
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

	sourceTypeRaw, _ := c.Get(httphelp.CtxKeySourceType)
	source, _ := sourceTypeRaw.(domain.AuditSourceType)
	if source == "" {
		source = domain.AuditSourceOIDC
	}

	ticket, err := h.cfg.RealtimeTickets.Issue(c.Request.Context(), login, role, source)
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
