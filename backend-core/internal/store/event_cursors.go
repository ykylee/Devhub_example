package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// EventCursor — Keycloak event polling 의 dedup state.
// migration 000031 (sprint claude/work_260519-u) 가 도입한 event_cursors 테이블.
// design: docs/planning/keycloak_event_audit_integration.md §3.3.
type EventCursor struct {
	CursorKey     string
	LastEventAt   time.Time
	LastEventHash string
	UpdatedAt     time.Time
}

// EventCursorStore — Keycloak event polling 의 dedup state 영구화 interface.
// KeycloakEventPuller (internal/audit/keycloak_event_puller.go) 가 의존.
type EventCursorStore interface {
	// GetEventCursor returns the cursor for the given key. If no row exists,
	// returns ErrNotFound — caller treats this as "start from epoch" (or
	// `time.Now()` for first run safety).
	GetEventCursor(ctx context.Context, cursorKey string) (EventCursor, error)
	// UpsertEventCursor stores the cursor advance — INSERT new row or UPDATE
	// existing row. Idempotent.
	UpsertEventCursor(ctx context.Context, cursor EventCursor) error
}

// GetEventCursor — PostgresStore impl.
func (s *PostgresStore) GetEventCursor(ctx context.Context, cursorKey string) (EventCursor, error) {
	key := strings.TrimSpace(cursorKey)
	if key == "" {
		return EventCursor{}, errors.New("cursor_key is required")
	}
	const q = `
SELECT cursor_key, last_event_at, COALESCE(last_event_hash, ''), updated_at
FROM event_cursors
WHERE cursor_key = $1
`
	var c EventCursor
	err := s.pool.QueryRow(ctx, q, key).Scan(&c.CursorKey, &c.LastEventAt, &c.LastEventHash, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return EventCursor{}, fmt.Errorf("event_cursors %s: %w", key, ErrNotFound)
		}
		return EventCursor{}, fmt.Errorf("get event_cursors %s: %w", key, err)
	}
	return c, nil
}

// UpsertEventCursor — PostgresStore impl. INSERT new or UPDATE existing.
// updated_at 은 DB DEFAULT NOW() 사용 (last_event_at 은 caller 가 명시 — Keycloak
// event 의 timestamp 와 정합).
func (s *PostgresStore) UpsertEventCursor(ctx context.Context, cursor EventCursor) error {
	key := strings.TrimSpace(cursor.CursorKey)
	if key == "" {
		return errors.New("cursor_key is required")
	}
	if cursor.LastEventAt.IsZero() {
		return errors.New("last_event_at is required")
	}
	const q = `
INSERT INTO event_cursors (cursor_key, last_event_at, last_event_hash, updated_at)
VALUES ($1, $2, NULLIF($3, ''), NOW())
ON CONFLICT (cursor_key) DO UPDATE
SET last_event_at = EXCLUDED.last_event_at,
    last_event_hash = NULLIF(EXCLUDED.last_event_hash, ''),
    updated_at = NOW()
`
	if _, err := s.pool.Exec(ctx, q, key, cursor.LastEventAt, cursor.LastEventHash); err != nil {
		return fmt.Errorf("upsert event_cursors %s: %w", key, err)
	}
	return nil
}
