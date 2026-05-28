package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// RealtimeTicket — ADR-0024 §3.2 ticket pattern 의 영구화 row (migration 000035).
type RealtimeTicket struct {
	Ticket     string
	ActorLogin string
	ActorRole  string
	SourceType string
	ExpiresAt  time.Time
}

// InsertRealtimeTicket stores a freshly issued ticket.
func (r *RealtimeRepository) InsertRealtimeTicket(ctx context.Context, t RealtimeTicket) error {
	ticket := strings.TrimSpace(t.Ticket)
	if ticket == "" {
		return errors.New("ticket is required")
	}
	if strings.TrimSpace(t.ActorLogin) == "" {
		return errors.New("actor_login is required")
	}
	if t.ExpiresAt.IsZero() {
		return errors.New("expires_at is required")
	}
	const q = `
INSERT INTO realtime_tickets (ticket, actor_login, actor_role, source_type, expires_at)
VALUES ($1, $2, $3, $4, $5)
`
	if _, err := r.store.Pool().Exec(ctx, q, ticket, t.ActorLogin, t.ActorRole, t.SourceType, t.ExpiresAt); err != nil {
		return fmt.Errorf("insert realtime_tickets: %w", err)
	}
	return nil
}

// ConsumeRealtimeTicket atomically deletes and returns a non-expired ticket.
func (r *RealtimeRepository) ConsumeRealtimeTicket(ctx context.Context, ticket string) (RealtimeTicket, bool, error) {
	ticket = strings.TrimSpace(ticket)
	if ticket == "" {
		return RealtimeTicket{}, false, nil
	}
	const q = `
DELETE FROM realtime_tickets
WHERE ticket = $1 AND expires_at > NOW()
RETURNING ticket, actor_login, actor_role, source_type, expires_at
`
	var t RealtimeTicket
	err := r.store.Pool().QueryRow(ctx, q, ticket).Scan(&t.Ticket, &t.ActorLogin, &t.ActorRole, &t.SourceType, &t.ExpiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return RealtimeTicket{}, false, nil
		}
		return RealtimeTicket{}, false, fmt.Errorf("consume realtime_tickets: %w", err)
	}
	return t, true, nil
}

// DeleteExpiredRealtimeTickets removes expired rows so the table stays bounded.
func (r *RealtimeRepository) DeleteExpiredRealtimeTickets(ctx context.Context) (int64, error) {
	const q = `DELETE FROM realtime_tickets WHERE expires_at <= NOW()`
	tag, err := r.store.Pool().Exec(ctx, q)
	if err != nil {
		return 0, fmt.Errorf("delete expired realtime_tickets: %w", err)
	}
	return tag.RowsAffected(), nil
}
