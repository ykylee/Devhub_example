package httpapi

import (
	"context"
	"errors"
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

	entry, ok := s.consume(ctx, ticket)
	if !ok {
		t.Fatal("first consume should succeed")
	}
	if entry.actorLogin != "alice" || entry.actorRole != "developer" || entry.sourceType != domain.AuditSourceOIDC {
		t.Fatalf("unexpected entry: %+v", entry)
	}

	// single-use: 두 번째 consume 은 miss.
	if _, ok := s.consume(ctx, ticket); ok {
		t.Fatal("second consume should miss (single-use)")
	}
}

func TestRealtimeTicketStore_InMemory_UnknownAndExpired(t *testing.T) {
	s := NewRealtimeTicketStore()
	ctx := context.Background()

	if _, ok := s.consume(ctx, "does-not-exist"); ok {
		t.Fatal("unknown ticket should miss")
	}

	// 만료 ticket 은 miss (직접 expiresAt 과거로 주입).
	s.mu.Lock()
	s.tickets["expired"] = &realtimeTicket{
		actorLogin: "bob",
		expiresAt:  time.Now().Add(-time.Second),
	}
	s.mu.Unlock()
	if _, ok := s.consume(ctx, "expired"); ok {
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

	entry, ok := s.consume(ctx, ticket)
	if !ok {
		t.Fatal("first consume should succeed")
	}
	if entry.actorLogin != "carol" || entry.actorRole != "manager" || entry.sourceType != domain.AuditSourceOIDC {
		t.Fatalf("unexpected entry: %+v", entry)
	}

	if _, ok := s.consume(ctx, ticket); ok {
		t.Fatal("second consume should miss (single-use)")
	}
}

func TestDBRealtimeTicketStore_ConsumeErrorIsMiss(t *testing.T) {
	fake := newFakeRealtimeTicketDB()
	fake.consumeErr = errors.New("db down")
	s := NewDBRealtimeTicketStore(fake)

	if _, ok := s.consume(context.Background(), "anything"); ok {
		t.Fatal("consume error should report miss, not panic/honor")
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
