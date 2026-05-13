package hrdb_test

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/devhub/backend-core/internal/hrdb"
)

// hrdb_postgres_test pins the SQL the PostgresClient issues against the
// hrdb.persons table introduced by ADR-0008 (sprint claude/work_260513-m).
// Tests follow the store_test pattern: skip when no DEVHUB_TEST_DB_URL is
// set so CI without a seeded `hrdb` schema does not fail. ETL/integration
// sprint will exercise these against a seeded DB.

func newTestPostgresClient(t *testing.T) *hrdb.PostgresClient {
	t.Helper()
	dbURL := os.Getenv("DEVHUB_TEST_DB_URL")
	if dbURL == "" {
		t.Skip("DEVHUB_TEST_DB_URL is not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("pgxpool.New: %v", err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(ctx); err != nil {
		t.Skipf("test DB unreachable: %v", err)
	}
	return &hrdb.PostgresClient{Pool: pool, EmailFallbackDomain: "example.com"}
}

func TestPostgresClient_Lookup_MissReturnsErrPersonNotFound(t *testing.T) {
	client := newTestPostgresClient(t)

	_, _, _, err := client.Lookup(context.Background(), "this-system-id-must-not-exist-yet", "9999", "No One")
	if !errors.Is(err, hrdb.ErrPersonNotFound) {
		t.Fatalf("expected ErrPersonNotFound for empty hrdb.persons, got %v", err)
	}
}

func TestPostgresClient_Lookup_NilPoolReturnsError(t *testing.T) {
	// Pure unit test (no DB) — guard against accidentally constructing a
	// PostgresClient without wiring the pgx pool.
	client := &hrdb.PostgresClient{}
	_, _, _, err := client.Lookup(context.Background(), "x", "1", "Y")
	if err == nil || !strings.Contains(err.Error(), "Pool is not configured") {
		t.Errorf("expected nil-Pool error, got %v", err)
	}
}
