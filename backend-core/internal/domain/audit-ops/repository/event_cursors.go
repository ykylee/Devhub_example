package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/devhub/backend-core/internal/store"
	"github.com/jackc/pgx/v5"
)

// EventCursor — Keycloak event polling 의 dedup state.
type EventCursor struct {
	CursorKey     string
	LastEventAt   time.Time
	LastEventHash string
	UpdatedAt     time.Time
}

// EventCursorStore — Keycloak event polling 의 dedup state 영구화 interface.
type EventCursorStore interface {
	GetEventCursor(ctx context.Context, cursorKey string) (EventCursor, error)
	UpsertEventCursor(ctx context.Context, cursor EventCursor) error
}

// GetEventCursor — AuditRepository impl.
func (r *AuditRepository) GetEventCursor(ctx context.Context, cursorKey string) (EventCursor, error) {
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
	err := r.store.Pool().QueryRow(ctx, q, key).Scan(&c.CursorKey, &c.LastEventAt, &c.LastEventHash, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return EventCursor{}, fmt.Errorf("event_cursors %s: %w", key, store.ErrNotFound)
		}
		return EventCursor{}, fmt.Errorf("get event_cursors %s: %w", key, err)
	}
	return c, nil
}

// UpsertEventCursor — AuditRepository impl. INSERT new or UPDATE existing.
func (r *AuditRepository) UpsertEventCursor(ctx context.Context, cursor EventCursor) error {
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
	if _, err := r.store.Pool().Exec(ctx, q, key, cursor.LastEventAt, cursor.LastEventHash); err != nil {
		return fmt.Errorf("upsert event_cursors %s: %w", key, err)
	}
	return nil
}
