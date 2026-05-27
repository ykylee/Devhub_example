package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
)

// ADR-0024 §3.2 ticket pattern + §6 carve 6 (multi-instance PG 백킹) 단위테스트.

func TestRealtimeTicketStore_InMemory_IssueConsumeSingleUse(t *testing.T) {
	s := NewRealtimeTicketStore()
	ctx := context.Background()

	ticket, err := s.issue(ctx, "alice", "developer", domain.AuditSourceOIDC)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if ticket == "" {
		t.Fatal("expected non-empty ticket")
	}

	entry, ok, err := s.consume(ctx, ticket)
	if err != nil {
		t.Fatalf("consume err: %v", err)
	}
	if !ok {
		t.Fatal("first consume should succeed")
	}
	if entry.actorLogin != "alice" || entry.actorRole != "developer" || entry.sourceType != domain.AuditSourceOIDC {
		t.Fatalf("unexpected entry: %+v", entry)
	}

	// single-use: 두 번째 consume 은 miss.
	if _, ok, _ := s.consume(ctx, ticket); ok {
		t.Fatal("second consume should miss (single-use)")
	}
}

func TestRealtimeTicketStore_InMemory_UnknownAndExpired(t *testing.T) {
	s := NewRealtimeTicketStore()
	ctx := context.Background()

	if _, ok, err := s.consume(ctx, "does-not-exist"); ok || err != nil {
		t.Fatalf("unknown ticket should miss with no error, got ok=%v err=%v", ok, err)
	}

	// 만료 ticket 은 miss (직접 expiresAt 과거로 주입).
	s.mu.Lock()
	s.tickets["expired"] = &realtimeTicket{
		actorLogin: "bob",
		expiresAt:  time.Now().Add(-time.Second),
	}
	s.mu.Unlock()
	if _, ok, _ := s.consume(ctx, "expired"); ok {
		t.Fatal("expired ticket should miss")
	}
}

func TestNewRealtimeTicketStoreFor_NilUsesInMemory(t *testing.T) {
	// 명시적 typed-nil *store.PostgresStore → in-memory fallback (typed-nil
	// interface pitfall 회피 검증).
	var pg *store.PostgresStore
	got := NewRealtimeTicketStoreFor(pg)
	if _, ok := got.(*RealtimeTicketStore); !ok {
		t.Fatalf("expected *RealtimeTicketStore, got %T", got)
	}
}

// fakeRealtimeTicketDB — DBRealtimeTicketStore 단위테스트용 in-memory fake.
type fakeRealtimeTicketDB struct {
	rows         map[string]store.RealtimeTicket
	insertErr    error
	consumeErr   error
	deleteCalls  int
	insertCalls  int
	consumeCalls int
}

func newFakeRealtimeTicketDB() *fakeRealtimeTicketDB {
	return &fakeRealtimeTicketDB{rows: make(map[string]store.RealtimeTicket)}
}

func (f *fakeRealtimeTicketDB) InsertRealtimeTicket(_ context.Context, t store.RealtimeTicket) error {
	f.insertCalls++
	if f.insertErr != nil {
		return f.insertErr
	}
	f.rows[t.Ticket] = t
	return nil
}

func (f *fakeRealtimeTicketDB) ConsumeRealtimeTicket(_ context.Context, ticket string) (store.RealtimeTicket, bool, error) {
	f.consumeCalls++
	if f.consumeErr != nil {
		return store.RealtimeTicket{}, false, f.consumeErr
	}
	row, ok := f.rows[ticket]
	if !ok {
		return store.RealtimeTicket{}, false, nil
	}
	if row.ExpiresAt.Before(time.Now()) {
		return store.RealtimeTicket{}, false, nil
	}
	delete(f.rows, ticket) // single-use
	return row, true, nil
}

func (f *fakeRealtimeTicketDB) DeleteExpiredRealtimeTickets(_ context.Context) (int64, error) {
	f.deleteCalls++
	var n int64
	now := time.Now()
	for k, v := range f.rows {
		if v.ExpiresAt.Before(now) {
			delete(f.rows, k)
			n++
		}
	}
	return n, nil
}

func TestDBRealtimeTicketStore_IssueConsumeSingleUse(t *testing.T) {
	fake := newFakeRealtimeTicketDB()
	s := NewDBRealtimeTicketStore(fake)
	ctx := context.Background()

	ticket, err := s.issue(ctx, "carol", "manager", domain.AuditSourceOIDC)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if fake.insertCalls != 1 {
		t.Fatalf("expected 1 insert, got %d", fake.insertCalls)
	}
	// issue 가 opportunistic cleanup 을 호출.
	if fake.deleteCalls != 1 {
		t.Fatalf("expected 1 opportunistic cleanup, got %d", fake.deleteCalls)
	}

	entry, ok, err := s.consume(ctx, ticket)
	if err != nil {
		t.Fatalf("consume err: %v", err)
	}
	if !ok {
		t.Fatal("first consume should succeed")
	}
	if entry.actorLogin != "carol" || entry.actorRole != "manager" || entry.sourceType != domain.AuditSourceOIDC {
		t.Fatalf("unexpected entry: %+v", entry)
	}

	if _, ok, _ := s.consume(ctx, ticket); ok {
		t.Fatal("second consume should miss (single-use)")
	}
}

// codex review PR #344 — store fault 를 miss (401) 로 collapse 하지 않고 error 를
// 전파해야 auth flow 가 503 으로 구분 가능.
func TestDBRealtimeTicketStore_ConsumeErrorIsPropagated(t *testing.T) {
	fake := newFakeRealtimeTicketDB()
	fake.consumeErr = errors.New("db down")
	s := NewDBRealtimeTicketStore(fake)

	entry, ok, err := s.consume(context.Background(), "anything")
	if err == nil {
		t.Fatal("store fault should be propagated as error (not silent miss)")
	}
	if ok || entry != nil {
		t.Fatalf("store fault must not honor a ticket, got ok=%v entry=%v", ok, entry)
	}
}

func TestDBRealtimeTicketStore_IssuePropagatesInsertError(t *testing.T) {
	fake := newFakeRealtimeTicketDB()
	fake.insertErr = errors.New("insert failed")
	s := NewDBRealtimeTicketStore(fake)

	if _, err := s.issue(context.Background(), "dave", "developer", domain.AuditSourceOIDC); err == nil {
		t.Fatal("expected issue to propagate insert error")
	}
}

// faultyTicketStore — consume 가 항상 store fault 를 반환.
type faultyTicketStore struct{}

func (faultyTicketStore) issue(_ context.Context, _, _ string, _ domain.AuditSourceType) (string, error) {
	return "", nil
}
func (faultyTicketStore) consume(_ context.Context, _ string) (*realtimeTicket, bool, error) {
	return nil, false, errors.New("postgres connection lost")
}

// codex review PR #344 — auth flow 가 ticket store fault (5xx) 와 invalid ticket
// (401) 를 구분해야 한다. store fault 시 valid ticket 이 401 로 거부되거나 infra
// 신호가 사라지면 안 됨.
func TestRealtimeWS_TicketStoreFault_Returns503(t *testing.T) {
	router := NewRouter(RouterConfig{
		RealtimeHub:     NewRealtimeHub(),
		RealtimeTickets: faultyTicketStore{},
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/realtime/ws?ticket=maybe-valid", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("ticket store fault should return 503, got %d", rec.Code)
	}
}

// 진짜 miss (unknown/expired) 는 401 — store fault 와 구분.
func TestRealtimeWS_TicketMiss_Returns401(t *testing.T) {
	router := NewRouter(RouterConfig{
		RealtimeHub:     NewRealtimeHub(),
		RealtimeTickets: NewRealtimeTicketStore(), // 비어 있음 → consume miss
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/realtime/ws?ticket=unknown", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unknown ticket should return 401, got %d", rec.Code)
	}
}
