package httpapi

import (
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

type realtimeTicket struct {
	actorLogin string
	actorRole  string
	sourceType domain.AuditSourceType
	expiresAt  time.Time
}

// realtimeTicketStore — issue/consume 추상화. single-instance 는 in-memory
// (RealtimeTicketStore), multi-instance (sticky session 없이 horizontal scale,
// ADR-0024 §6 carve 6) 는 PG 백킹 (DBRealtimeTicketStore) 구현이 주입된다.
type realtimeTicketStore interface {
	issue(ctx context.Context, actorLogin, actorRole string, sourceType domain.AuditSourceType) (string, error)
	// consume returns (entry, true, nil) on hit, (nil, false, nil) on a genuine
	// miss (unknown/expired/already-used), and (nil, false, err) on a store
	// fault. Callers MUST distinguish err != nil (infra outage — 5xx) from a
	// miss (401), so a valid ticket is not rejected as unauthenticated during a
	// transient DB failure (codex review PR #344).
	consume(ctx context.Context, ticket string) (*realtimeTicket, bool, error)
}

func newRealtimeTicket() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// RealtimeTicketStore — single-instance in-memory ticket store.
// ADR-0024 §6 carve 1 (ticket pattern) minimal 구현. DB 미연결 (in-memory event
// store) 배포에서 fallback 으로 사용. multi-instance 는 DBRealtimeTicketStore.
type RealtimeTicketStore struct {
	mu      sync.Mutex
	tickets map[string]*realtimeTicket
}

func NewRealtimeTicketStore() *RealtimeTicketStore {
	return &RealtimeTicketStore{tickets: make(map[string]*realtimeTicket)}
}

func (s *RealtimeTicketStore) issue(_ context.Context, actorLogin, actorRole string, sourceType domain.AuditSourceType) (string, error) {
	ticket, err := newRealtimeTicket()
	if err != nil {
		return "", err
	}
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
func (s *RealtimeTicketStore) consume(_ context.Context, ticket string) (*realtimeTicket, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	s.sweepLocked(now)
	entry, ok := s.tickets[ticket]
	if !ok {
		return nil, false, nil
	}
	delete(s.tickets, ticket)
	if entry.expiresAt.Before(now) {
		return nil, false, nil
	}
	return entry, true, nil
}

// realtimeTicketDB — PG 백킹 ticket store 가 의존하는 store 메서드.
// store.PostgresStore (internal/store/realtime_tickets.go) 가 구현.
type realtimeTicketDB interface {
	InsertRealtimeTicket(ctx context.Context, t store.RealtimeTicket) error
	ConsumeRealtimeTicket(ctx context.Context, ticket string) (store.RealtimeTicket, bool, error)
	DeleteExpiredRealtimeTickets(ctx context.Context) (int64, error)
}

// DBRealtimeTicketStore — multi-instance 안전 PG 백킹 ticket store (ADR-0024 §6
// carve 6). issue/consume 가 realtime_tickets 테이블을 통하므로 어느 인스턴스가
// 발급해도 다른 인스턴스가 single-use 로 소비 가능 (DELETE ... RETURNING 원자성).
type DBRealtimeTicketStore struct {
	db realtimeTicketDB
}

func NewDBRealtimeTicketStore(db realtimeTicketDB) *DBRealtimeTicketStore {
	return &DBRealtimeTicketStore{db: db}
}

// NewRealtimeTicketStoreFor selects the ticket store implementation: PG-backed
// (multi-instance safe, ADR-0024 §6 carve 6) when a PostgresStore is available,
// else the in-memory fallback (single-instance / DB 미연결 배포). The concrete
// *store.PostgresStore nil check avoids the typed-nil interface pitfall.
func NewRealtimeTicketStoreFor(pg *store.PostgresStore) realtimeTicketStore {
	if pg == nil {
		return NewRealtimeTicketStore()
	}
	return NewDBRealtimeTicketStore(pg)
}

func (s *DBRealtimeTicketStore) issue(ctx context.Context, actorLogin, actorRole string, sourceType domain.AuditSourceType) (string, error) {
	ticket, err := newRealtimeTicket()
	if err != nil {
		return "", err
	}
	// opportunistic cleanup — best-effort, background goroutine 불요.
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

func (s *DBRealtimeTicketStore) consume(ctx context.Context, ticket string) (*realtimeTicket, bool, error) {
	row, ok, err := s.db.ConsumeRealtimeTicket(ctx, ticket)
	if err != nil {
		// store fault — surface so the caller can return 5xx instead of
		// conflating an infra outage with an invalid ticket (401).
		return nil, false, err
	}
	if !ok {
		return nil, false, nil
	}
	return &realtimeTicket{
		actorLogin: row.ActorLogin,
		actorRole:  row.ActorRole,
		sourceType: domain.AuditSourceType(row.SourceType),
		expiresAt:  row.ExpiresAt,
	}, true, nil
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

	ticket, err := h.cfg.RealtimeTickets.issue(c.Request.Context(), login, role, source)
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
