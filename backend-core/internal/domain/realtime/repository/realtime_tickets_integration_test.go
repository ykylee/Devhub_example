package repository_test

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/domain/realtime/repository"
	rtview "github.com/devhub/backend-core/internal/domain/realtime/view"
	"github.com/devhub/backend-core/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

// realtime_tickets integration test (sprint claude/work_260527-adr0024-ticket-store).
//
// CI backend-unit job 은 DEVHUB_TEST_DB_URL 미설정으로 t.Skip — postgres_*_test.go 패턴 정합.
// migration 000035 (realtime_tickets 테이블) 가 적용된 DB 환경 필요.
// ADR-0024 §6 carve 6 — multi-instance single-use 보장 (DELETE ... RETURNING) 검증.

func newTicketTestStore(t *testing.T) (*store.PostgresStore, *pgxpool.Pool, context.Context) {
	t.Helper()
	dbURL := os.Getenv("DEVHUB_TEST_DB_URL")
	if dbURL == "" {
		t.Skip("DEVHUB_TEST_DB_URL is not set")
	}
	ctx := context.Background()
	pgStore, err := store.NewPostgresStore(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect postgres store: %v", err)
	}
	t.Cleanup(pgStore.Close)

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect raw pool: %v", err)
	}
	t.Cleanup(pool.Close)
	return pgStore, pool, ctx
}

// TestIntegration_RealtimeTickets_InsertConsumeSingleUse — INSERT 후 첫 consume
// 성공 + 두 번째 consume miss (single-use, DELETE 로 row 제거).
func TestIntegration_RealtimeTickets_InsertConsumeSingleUse(t *testing.T) {
	pgStore, pool, ctx := newTicketTestStore(t)
	repo := repository.NewRealtimeRepository(pgStore)

	ticket := fmt.Sprintf("test-su-%d", time.Now().UnixNano())
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM realtime_tickets WHERE ticket = $1`, ticket)
	}()

	if err := repo.InsertRealtimeTicket(ctx, repository.RealtimeTicket{
		Ticket:     ticket,
		ActorLogin: "alice",
		ActorRole:  "developer",
		SourceType: "oidc",
		ExpiresAt:  time.Now().Add(60 * time.Second),
	}); err != nil {
		t.Fatalf("insert: %v", err)
	}

	row, ok, err := repo.ConsumeRealtimeTicket(ctx, ticket)
	if err != nil {
		t.Fatalf("consume: %v", err)
	}
	if !ok {
		t.Fatal("first consume should succeed")
	}
	if row.ActorLogin != "alice" || row.ActorRole != "developer" || row.SourceType != "oidc" {
		t.Fatalf("unexpected row: %+v", row)
	}

	_, ok, err = repo.ConsumeRealtimeTicket(ctx, ticket)
	if err != nil {
		t.Fatalf("second consume err: %v", err)
	}
	if ok {
		t.Fatal("second consume should miss (single-use)")
	}
}

// TestIntegration_RealtimeTickets_ExpiredNotConsumed — 만료 ticket 은 consume miss.
func TestIntegration_RealtimeTickets_ExpiredNotConsumed(t *testing.T) {
	pgStore, pool, ctx := newTicketTestStore(t)
	repo := repository.NewRealtimeRepository(pgStore)

	ticket := fmt.Sprintf("test-exp-%d", time.Now().UnixNano())
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM realtime_tickets WHERE ticket = $1`, ticket)
	}()

	if err := repo.InsertRealtimeTicket(ctx, repository.RealtimeTicket{
		Ticket:     ticket,
		ActorLogin: "bob",
		SourceType: "oidc",
		ExpiresAt:  time.Now().Add(-time.Second), // 이미 만료
	}); err != nil {
		t.Fatalf("insert: %v", err)
	}

	_, ok, err := repo.ConsumeRealtimeTicket(ctx, ticket)
	if err != nil {
		t.Fatalf("consume: %v", err)
	}
	if ok {
		t.Fatal("expired ticket should not be consumed")
	}

	// DeleteExpiredRealtimeTickets 가 만료 row 회수.
	if _, err := repo.DeleteExpiredRealtimeTickets(ctx); err != nil {
		t.Fatalf("delete expired: %v", err)
	}
	var cnt int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM realtime_tickets WHERE ticket = $1`, ticket).Scan(&cnt); err != nil {
		t.Fatalf("count: %v", err)
	}
	if cnt != 0 {
		t.Fatalf("expired row should be reaped, found %d", cnt)
	}
}

// TestIntegration_RealtimeTickets_ConcurrentConsumeHonoredOnce — N goroutine 이
// 같은 ticket 을 동시에 consume 해도 정확히 1회만 성공 (DELETE ... RETURNING 원자성).
// multi-instance horizontal scale 의 single-use invariant 핵심 검증.
func TestIntegration_RealtimeTickets_ConcurrentConsumeHonoredOnce(t *testing.T) {
	pgStore, pool, ctx := newTicketTestStore(t)
	repo := repository.NewRealtimeRepository(pgStore)

	ticket := fmt.Sprintf("test-race-%d", time.Now().UnixNano())
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM realtime_tickets WHERE ticket = $1`, ticket)
	}()

	if err := repo.InsertRealtimeTicket(ctx, repository.RealtimeTicket{
		Ticket:     ticket,
		ActorLogin: "carol",
		SourceType: "oidc",
		ExpiresAt:  time.Now().Add(60 * time.Second),
	}); err != nil {
		t.Fatalf("insert: %v", err)
	}

	const n = 8
	var wg sync.WaitGroup
	var mu sync.Mutex
	successes := 0
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			_, ok, err := repo.ConsumeRealtimeTicket(ctx, ticket)
			if err != nil {
				return
			}
			if ok {
				mu.Lock()
				successes++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if successes != 1 {
		t.Fatalf("ticket must be honored exactly once under concurrency, got %d", successes)
	}
}

// ---------------------------------------------------------------------------
// DBRealtimeTicketStore integration tests — view-layer adapter wrapping
// store methods with side-effect (DeleteExpiredRealtimeTickets on Issue).
// ---------------------------------------------------------------------------

func TestIntegration_DBRealtimeTicketStore_IssueConsume(t *testing.T) {
	pgStore, pool, ctx := newTicketTestStore(t)
	store2 := rtview.NewDBRealtimeTicketStore(pgStore)

	issued, err := store2.Issue(ctx, "alice", "developer", domain.AuditSourceOIDC)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if issued == "" {
		t.Fatal("expected non-empty ticket")
	}
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM realtime_tickets WHERE ticket = $1`, issued)
	}()

	// Consume the issued ticket
	row, ok, err := store2.Consume(ctx, issued)
	if err != nil {
		t.Fatalf("Consume: %v", err)
	}
	if !ok {
		t.Fatal("first consume should succeed")
	}
	if row.ActorLogin != "alice" || row.ActorRole != "developer" || row.SourceType != string(domain.AuditSourceOIDC) {
		t.Fatalf("unexpected ticket row: %+v", row)
	}

	// Second consume must miss (single-use)
	_, ok, err = store2.Consume(ctx, issued)
	if err != nil {
		t.Fatalf("second Consume err: %v", err)
	}
	if ok {
		t.Fatal("second consume should miss (single-use)")
	}
}

func TestIntegration_DBRealtimeTicketStore_ExpiredTicketNotConsumed(t *testing.T) {
	pgStore, pool, ctx := newTicketTestStore(t)
	store2 := rtview.NewDBRealtimeTicketStore(pgStore)

	// Manually insert an expired ticket via raw pool to test Consume expiry
	ticket := fmt.Sprintf("test-db-exp-%d", time.Now().UnixNano())
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM realtime_tickets WHERE ticket = $1`, ticket)
	}()

	_, err := pool.Exec(ctx, `
		INSERT INTO realtime_tickets (ticket, actor_login, actor_role, source_type, expires_at)
		VALUES ($1, 'bob', 'manager', 'oidc', NOW() - INTERVAL '1 second')
	`, ticket)
	if err != nil {
		t.Fatalf("seed expired ticket: %v", err)
	}

	// Consume must miss for expired ticket
	_, ok, err := store2.Consume(ctx, ticket)
	if err != nil {
		t.Fatalf("Consume expired: %v", err)
	}
	if ok {
		t.Fatal("expired ticket should not be consumed")
	}
}
