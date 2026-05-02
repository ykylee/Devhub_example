package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrDuplicateEvent = errors.New("duplicate webhook event")

type WebhookEvent struct {
	ID             int64
	EventType      string
	DeliveryID     string
	DedupeKey      string
	RepositoryID   *int64
	RepositoryName string
	SenderLogin    string
	Payload        []byte
	Status         string
	ReceivedAt     time.Time
	ValidatedAt    *time.Time
}

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(ctx context.Context, dbURL string) (*PostgresStore, error) {
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return &PostgresStore{pool: pool}, nil
}

func (s *PostgresStore) Close() {
	s.pool.Close()
}

func (s *PostgresStore) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

func (s *PostgresStore) SaveWebhookEvent(ctx context.Context, event WebhookEvent) (int64, error) {
	const query = `
INSERT INTO webhook_events (
	event_type,
	delivery_id,
	dedupe_key,
	repository_id,
	repository_name,
	sender_login,
	payload,
	status,
	received_at,
	validated_at
) VALUES ($1, NULLIF($2, ''), $3, $4, NULLIF($5, ''), NULLIF($6, ''), $7::jsonb, $8, $9, $10)
ON CONFLICT (dedupe_key) DO NOTHING
RETURNING id`

	var id int64
	err := s.pool.QueryRow(
		ctx,
		query,
		event.EventType,
		event.DeliveryID,
		event.DedupeKey,
		event.RepositoryID,
		event.RepositoryName,
		event.SenderLogin,
		string(event.Payload),
		event.Status,
		event.ReceivedAt,
		event.ValidatedAt,
	).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrDuplicateEvent
	}
	if err != nil {
		return 0, err
	}
	return id, nil
}

type ListWebhookEventsOptions struct {
	Limit  int
	Offset int
}

func (s *PostgresStore) ListWebhookEvents(ctx context.Context, opts ListWebhookEventsOptions) ([]WebhookEvent, error) {
	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	const query = `
SELECT
	id,
	event_type,
	COALESCE(delivery_id, ''),
	dedupe_key,
	repository_id,
	COALESCE(repository_name, ''),
	COALESCE(sender_login, ''),
	payload,
	status,
	received_at,
	validated_at
FROM webhook_events
ORDER BY received_at DESC, id DESC
LIMIT $1 OFFSET $2`

	rows, err := s.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]WebhookEvent, 0, limit)
	for rows.Next() {
		var event WebhookEvent
		if err := rows.Scan(
			&event.ID,
			&event.EventType,
			&event.DeliveryID,
			&event.DedupeKey,
			&event.RepositoryID,
			&event.RepositoryName,
			&event.SenderLogin,
			&event.Payload,
			&event.Status,
			&event.ReceivedAt,
			&event.ValidatedAt,
		); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}
