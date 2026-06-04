package repository_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	auditrep "github.com/devhub/backend-core/internal/domain/audit-ops/repository"
	"github.com/devhub/backend-core/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

func newAuditTestStore(t *testing.T) (*store.PostgresStore, *pgxpool.Pool, context.Context) {
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

func TestIntegration_AuditLogs_CRUD(t *testing.T) {
	pgStore, pool, ctx := newAuditTestStore(t)
	repo := auditrep.NewAuditRepository(pgStore)

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	sourceEventID := "evt-" + suffix

	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM audit_logs WHERE source_event_id = $1`, sourceEventID)
	}()

	logInput := domain.AuditLog{
		ActorLogin:    "alice",
		Action:        "create",
		TargetType:    "platform",
		TargetID:      "test-app-id",
		SourceIP:      "127.0.0.1",
		RequestID:     "req-" + suffix,
		SourceType:    domain.AuditSourceKeycloakEvent,
		SourceEventID: sourceEventID,
		Payload:       map[string]any{"meta": "data"},
	}

	// 1. CreateAuditLog
	created, err := repo.CreateAuditLog(ctx, logInput)
	if err != nil {
		t.Fatalf("CreateAuditLog failed: %v", err)
	}
	if created.ActorLogin != "alice" || created.Action != "create" {
		t.Errorf("unexpected created audit log details: %+v", created)
	}

	// 2. CreateAuditLog ON CONFLICT (Dedup)
	duplicate, err := repo.CreateAuditLog(ctx, logInput)
	if err != nil {
		t.Fatalf("CreateAuditLog duplicate failed: %v", err)
	}
	if duplicate.ID != created.ID {
		t.Errorf("expected duplicate dedup to return same ID, got %d vs %d", duplicate.ID, created.ID)
	}

	// 3. ListAuditLogs
	opts := store.ListAuditLogsOptions{
		ActorLogin: "alice",
		Action:     "create",
		TargetType: "platform",
		TargetID:   "test-app-id",
	}
	logs, err := repo.ListAuditLogs(ctx, opts)
	if err != nil {
		t.Fatalf("ListAuditLogs failed: %v", err)
	}
	if len(logs) == 0 {
		t.Errorf("expected at least 1 audit log, got 0")
	}

	// 4. Create empty payload and command ID checks
	logEmpty := domain.AuditLog{
		Action: "empty-test",
	}
	createdEmpty, err := repo.CreateAuditLog(ctx, logEmpty)
	if err != nil {
		t.Fatalf("CreateAuditLog with empty payload failed: %v", err)
	}
	if createdEmpty.ActorLogin != "system" {
		t.Errorf("expected default actor 'system', got %s", createdEmpty.ActorLogin)
	}
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM audit_logs WHERE audit_id = $1`, createdEmpty.AuditID)
	}()
}

func TestIntegration_EventCursors_CRUD(t *testing.T) {
	pgStore, pool, ctx := newAuditTestStore(t)
	repo := auditrep.NewAuditRepository(pgStore)

	cursorKey := "keycloak-test-cursor-" + fmt.Sprintf("%d", time.Now().UnixNano())
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM event_cursors WHERE cursor_key = $1`, cursorKey)
	}()

	// 1. Get non-existent cursor
	if _, err := repo.GetEventCursor(ctx, cursorKey); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound for ghost cursor, got %v", err)
	}

	// 2. UpsertEventCursor
	cursor := auditrep.EventCursor{
		CursorKey:     cursorKey,
		LastEventAt:   time.Now().Add(-1 * time.Hour),
		LastEventHash: "hash123",
	}
	err := repo.UpsertEventCursor(ctx, cursor)
	if err != nil {
		t.Fatalf("UpsertEventCursor failed: %v", err)
	}

	// 3. GetEventCursor
	loaded, err := repo.GetEventCursor(ctx, cursorKey)
	if err != nil {
		t.Fatalf("GetEventCursor failed: %v", err)
	}
	if loaded.CursorKey != cursorKey || loaded.LastEventHash != "hash123" {
		t.Errorf("unexpected loaded cursor details: %+v", loaded)
	}

	// 4. Upsert existing (ON CONFLICT UPDATE)
	cursor.LastEventHash = "hash456"
	err = repo.UpsertEventCursor(ctx, cursor)
	if err != nil {
		t.Fatalf("UpsertEventCursor update failed: %v", err)
	}

	loadedUpdate, err := repo.GetEventCursor(ctx, cursorKey)
	if err != nil {
		t.Fatalf("GetEventCursor update failed: %v", err)
	}
	if loadedUpdate.LastEventHash != "hash456" {
		t.Errorf("expected updated hash to be 'hash456', got '%s'", loadedUpdate.LastEventHash)
	}

	// 5. Validation errors
	if _, err := repo.GetEventCursor(ctx, ""); err == nil {
		t.Errorf("expected error for empty cursor key")
	}
	if err := repo.UpsertEventCursor(ctx, auditrep.EventCursor{}); err == nil {
		t.Errorf("expected error for empty upsert cursor")
	}
}
