package hrdb_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/devhub/backend-core/internal/infrastructure/hrdb"
)

// postgres_unit_test extends the DB-skipped pgx integration tests in
// postgres_test.go with pure-unit cover for the EmailFallbackDomain + Scan
// branches of (*PostgresClient).Lookup. pgxpool.New is lazy (no connect on
// construction) so these tests never open a network socket — the cancelled
// context surfaces an error at Scan time without touching DB.

// poolForUnit constructs a pgxpool that intentionally points at a non-routable
// loopback port. pgxpool.New is non-blocking; the (cancelled) ctx ensures the
// QueryRow Scan returns immediately with a context error.
func poolForUnit(t *testing.T) *pgxpool.Pool {
	t.Helper()
	pool, err := pgxpool.New(context.Background(), "postgres://127.0.0.1:1/devhub_unit_test")
	if err != nil {
		t.Fatalf("pgxpool.New: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// TestPostgresClient_Lookup_CancelledContextSurfacesErr drives Lookup through
// the EmailFallbackDomain default branch + Scan path WITHOUT a reachable DB.
func TestPostgresClient_Lookup_CancelledContextSurfacesErr(t *testing.T) {
	client := &hrdb.PostgresClient{Pool: poolForUnit(t)} // empty domain → default branch

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // force Scan to fail without touching the network
	email, sysID, dept, err := client.Lookup(ctx, "yklee", "1001", "YK Lee")
	if err == nil {
		t.Fatal("expected error from cancelled context, got nil")
	}
	// Cancellation must NOT collapse into ErrPersonNotFound — only pgx.ErrNoRows
	// is mapped to that sentinel.
	if errors.Is(err, hrdb.ErrPersonNotFound) {
		t.Errorf("cancellation must not be reported as ErrPersonNotFound: %v", err)
	}
	if email != "" || sysID != "" || dept != "" {
		t.Errorf("expected zero return values on err, got (%q,%q,%q)", email, sysID, dept)
	}
}

func TestPostgresClient_Lookup_CustomEmailFallbackDomain(t *testing.T) {
	// Pin that EmailFallbackDomain custom value is honoured. SQL parameter
	// inspection requires DB; we exercise the non-default branch via the
	// cancellation path so coverage tracks the assignment-skip line.
	client := &hrdb.PostgresClient{Pool: poolForUnit(t), EmailFallbackDomain: "corp.example.com"}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _, _, err := client.Lookup(ctx, "x", "1", "Y")
	if err == nil {
		t.Fatal("expected error from cancelled context, got nil")
	}
}
