package store_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

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

func TestIntegration_RealtimeTickets_InsertConsumeSingleUse(t *testing.T) {
	pgStore, pool, ctx := newTicketTestStore(t)

	ticket := fmt.Sprintf("test-su-%d", time.Now().UnixNano())
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM realtime_tickets WHERE ticket = $1`, ticket)
	}()

	if err := pgStore.InsertRealtimeTicket(ctx, store.RealtimeTicket{
		Ticket:     ticket,
		ActorLogin: "alice",
		ActorRole:  "developer",
		SourceType: "oidc",
		ExpiresAt:  time.Now().Add(60 * time.Second),
	}); err != nil {
		t.Fatalf("insert: %v", err)
	}

	row, ok, err := pgStore.ConsumeRealtimeTicket(ctx, ticket)
	if err != nil {
		t.Fatalf("consume: %v", err)
	}
	if !ok {
		t.Fatal("first consume should succeed")
	}
	if row.ActorLogin != "alice" || row.ActorRole != "developer" || row.SourceType != "oidc" {
		t.Fatalf("unexpected row: %+v", row)
	}

	_, ok, err = pgStore.ConsumeRealtimeTicket(ctx, ticket)
	if err != nil {
		t.Fatalf("second consume err: %v", err)
	}
	if ok {
		t.Fatal("second consume should miss (single-use)")
	}
}

func TestIntegration_RealtimeTickets_ExpiredNotConsumed(t *testing.T) {
	pgStore, pool, ctx := newTicketTestStore(t)

	ticket := fmt.Sprintf("test-exp-%d", time.Now().UnixNano())
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM realtime_tickets WHERE ticket = $1`, ticket)
	}()

	if err := pgStore.InsertRealtimeTicket(ctx, store.RealtimeTicket{
		Ticket:     ticket,
		ActorLogin: "bob",
		SourceType: "oidc",
		ExpiresAt:  time.Now().Add(-time.Second), // 이미 만료
	}); err != nil {
		t.Fatalf("insert: %v", err)
	}

	_, ok, err := pgStore.ConsumeRealtimeTicket(ctx, ticket)
	if err != nil {
		t.Fatalf("consume: %v", err)
	}
	if ok {
		t.Fatal("expired ticket should not be consumed")
	}

	// DeleteExpiredRealtimeTickets 가 만료 row 회수.
	if _, err := pgStore.DeleteExpiredRealtimeTickets(ctx); err != nil {
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
