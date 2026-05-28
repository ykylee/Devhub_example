package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ExternalTaskListOptions for filtering GET /api/v1/external-tasks.
type ExternalTaskListOptions = store.ExternalTaskListOptions

func scanExternalTaskItem(row pgx.Row) (domain.ExternalTaskItem, error) {
	var t domain.ExternalTaskItem
	var rawPayloadJSON []byte
	if err := row.Scan(
		&t.ID,
		&t.ProviderID,
		&t.ExternalID,
		&t.Title,
		&t.Description,
		&t.RawStatus,
		&t.NormalizedStatus,
		&t.Priority,
		&t.Assignee,
		&t.Reporter,
		&t.URL,
		&t.Labels,
		&rawPayloadJSON,
		&t.WebhookSeq,
		&t.FetchedAt,
		&t.DeletedAt,
		&t.CreatedAt,
		&t.UpdatedAt,
	); err != nil {
		return domain.ExternalTaskItem{}, fmt.Errorf("scan external task item: %w", err)
	}
	if len(rawPayloadJSON) > 0 {
		t.RawPayload = rawPayloadJSON
	}
	return t, nil
}

func scanExternalTaskItemSlice(rows pgx.Rows) ([]domain.ExternalTaskItem, error) {
	var items []domain.ExternalTaskItem
	for rows.Next() {
		item, err := scanExternalTaskItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// PostgresExternalTaskStore implements ExternalTaskStore.
type PostgresExternalTaskStore struct {
	pool *pgxpool.Pool
}

// NewPostgresExternalTaskStore creates a new store.
func NewPostgresExternalTaskStore(pool *pgxpool.Pool) *PostgresExternalTaskStore {
	return &PostgresExternalTaskStore{pool: pool}
}

// NewPostgresExternalTaskStoreFor produces a task store sharing the supplied
// PostgresStore's pool.
func NewPostgresExternalTaskStoreFor(pg *store.PostgresStore) *PostgresExternalTaskStore {
	if pg == nil {
		return nil
	}
	return &PostgresExternalTaskStore{pool: pg.Pool()}
}

// UpsertExternalTaskItem inserts or updates a task item by (provider_id, external_id).
func (s *PostgresExternalTaskStore) UpsertExternalTaskItem(ctx context.Context, t domain.ExternalTaskItem) (domain.ExternalTaskItem, error) {
	labels := t.Labels
	if labels == nil {
		labels = []string{}
	}

	const query = `
INSERT INTO external_task_items (
    provider_id, external_id, title, description, raw_status, normalized_status,
    priority, assignee, reporter, url, labels, raw_payload, webhook_seq,
    fetched_at, deleted_at
) VALUES (
    $1, $2, $3, $4, $5, $6,
    NULLIF($7, ''), NULLIF($8, ''), NULLIF($9, ''), NULLIF($10, ''), $11, $12::jsonb, $13,
    $14, $15
)
ON CONFLICT (provider_id, external_id) DO UPDATE SET
    title            = COALESCE(NULLIF(EXCLUDED.title, ''), external_task_items.title),
    description      = COALESCE(NULLIF(EXCLUDED.description, ''), external_task_items.description),
    raw_status       = EXCLUDED.raw_status,
    normalized_status= COALESCE(NULLIF(EXCLUDED.normalized_status, ''), external_task_items.normalized_status),
    priority         = EXCLUDED.priority,
    assignee         = EXCLUDED.assignee,
    reporter         = EXCLUDED.reporter,
    url              = COALESCE(NULLIF(EXCLUDED.url, ''), external_task_items.url),
    labels           = EXCLUDED.labels,
    raw_payload      = EXCLUDED.raw_payload,
    webhook_seq      = COALESCE(EXCLUDED.webhook_seq, external_task_items.webhook_seq),
    fetched_at       = EXCLUDED.fetched_at,
    deleted_at       = EXCLUDED.deleted_at,
    updated_at       = NOW()
RETURNING
    id, provider_id, external_id, title, description, raw_status, normalized_status,
    priority, assignee, reporter, url, labels, raw_payload, webhook_seq,
    fetched_at, deleted_at, created_at, updated_at
`
	return scanExternalTaskItem(s.pool.QueryRow(ctx, query,
		t.ProviderID, t.ExternalID, t.Title, t.Description, t.RawStatus, t.NormalizedStatus,
		t.Priority, t.Assignee, t.Reporter, t.URL,
		labels, t.RawPayload, t.WebhookSeq,
		t.FetchedAt, t.DeletedAt,
	))
}

// SoftDeleteExternalTaskItem marks a task item as deleted (deleted_at = NOW()).
func (s *PostgresExternalTaskStore) SoftDeleteExternalTaskItem(ctx context.Context, providerID, externalID string) error {
	const query = `
UPDATE external_task_items
SET deleted_at = NOW(), updated_at = NOW()
WHERE provider_id = $1 AND external_id = $2
`
	ct, err := s.pool.Exec(ctx, query, providerID, externalID)
	if err != nil {
		return fmt.Errorf("soft delete external task item: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return store.ErrNotFound
	}
	return nil
}

// ListExternalTaskItems returns a filtered list of task items.
func (s *PostgresExternalTaskStore) ListExternalTaskItems(ctx context.Context, opts ExternalTaskListOptions) ([]domain.ExternalTaskItem, int, error) {
	where := "WHERE 1=1"
	args := []any{}
	arg := 1

	if opts.ProviderID != "" {
		where += fmt.Sprintf(" AND provider_id = $%d", arg)
		args = append(args, opts.ProviderID)
		arg++
	}
	if opts.RawStatus != "" {
		where += fmt.Sprintf(" AND raw_status = $%d", arg)
		args = append(args, opts.RawStatus)
		arg++
	}
	if opts.NormalizedStatus != "" {
		where += fmt.Sprintf(" AND normalized_status = $%d", arg)
		args = append(args, opts.NormalizedStatus)
		arg++
	}
	if opts.Assignee != "" {
		where += fmt.Sprintf(" AND assignee = $%d", arg)
		args = append(args, opts.Assignee)
		arg++
	}
	if len(opts.Labels) > 0 {
		where += fmt.Sprintf(" AND labels && $%d::text[]", arg)
		args = append(args, opts.Labels)
		arg++
	}
	if !opts.IncludeDeleted {
		where += " AND deleted_at IS NULL"
	}

	limit := opts.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	countQuery := "SELECT COUNT(*) FROM external_task_items " + where
	var total int
	if err := s.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count external task items: %w", err)
	}

	listQuery := `
SELECT id, provider_id, external_id, title, description, raw_status, normalized_status,
       priority, assignee, reporter, url, labels, raw_payload, webhook_seq,
       fetched_at, deleted_at, created_at, updated_at
FROM external_task_items ` + where + fmt.Sprintf(`
ORDER BY fetched_at DESC
LIMIT $%d OFFSET $%d`, arg, arg+1)
	args = append(args, limit, offset)

	rows, err := s.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list external task items: %w", err)
	}
	defer rows.Close()

	items, err := scanExternalTaskItemSlice(rows)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// GetExternalTaskItemByID returns a single task item by its UUID.
func (s *PostgresExternalTaskStore) GetExternalTaskItemByID(ctx context.Context, id string) (domain.ExternalTaskItem, error) {
	const query = `
SELECT id, provider_id, external_id, title, description, raw_status, normalized_status,
       priority, assignee, reporter, url, labels, raw_payload, webhook_seq,
       fetched_at, deleted_at, created_at, updated_at
FROM external_task_items
WHERE id = $1
`
	item, err := scanExternalTaskItem(s.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ExternalTaskItem{}, store.ErrNotFound
		}
		return domain.ExternalTaskItem{}, fmt.Errorf("get external task item by id: %w", err)
	}
	return item, nil
}

// NextWebhookSeq acquires the next value from the task_webhook_seq sequence.
func (s *PostgresExternalTaskStore) NextWebhookSeq(ctx context.Context) (int64, error) {
	var seq int64
	if err := s.pool.QueryRow(ctx, "SELECT nextval('task_webhook_seq')").Scan(&seq); err != nil {
		return 0, fmt.Errorf("nextval task_webhook_seq: %w", err)
	}
	return seq, nil
}

// DetectWebhookSeqGaps returns the count of missing webhook_seq values for a provider.
func (s *PostgresExternalTaskStore) DetectWebhookSeqGaps(ctx context.Context, providerID string) (int64, error) {
	const query = `
WITH seq_range AS (
    SELECT MIN(webhook_seq) AS min_seq, MAX(webhook_seq) AS max_seq
    FROM external_task_items
    WHERE provider_id = $1 AND webhook_seq IS NOT NULL
),
all_seqs AS (
    SELECT generate_series(min_seq, max_seq) AS seq
    FROM seq_range
),
present_seqs AS (
    SELECT webhook_seq AS seq
    FROM external_task_items
    WHERE provider_id = $1 AND webhook_seq IS NOT NULL
)
SELECT COUNT(*)::bigint
FROM all_seqs
WHERE seq NOT IN (SELECT seq FROM present_seqs)
`
	var gapCount int64
	if err := s.pool.QueryRow(ctx, query, providerID, providerID).Scan(&gapCount); err != nil {
		return 0, fmt.Errorf("detect webhook seq gaps: %w", err)
	}
	return gapCount, nil
}

// UpdateProviderLastPulledAt updates last_pulled_at for a provider.
func (s *PostgresExternalTaskStore) UpdateProviderLastPulledAt(ctx context.Context, providerID string, pulledAt time.Time) error {
	const query = `UPDATE integration_providers SET last_pulled_at = $1 WHERE provider_id = $2::uuid`
	_, err := s.pool.Exec(ctx, query, pulledAt, providerID)
	if err != nil {
		return fmt.Errorf("update provider last_pulled_at: %w", err)
	}
	return nil
}

// ListTaskTrackers returns all task_tracker integration providers.
func (s *PostgresExternalTaskStore) ListTaskTrackers(ctx context.Context) ([]domain.IntegrationProvider, error) {
	const query = `
SELECT provider_id::text, provider_key, provider_type, display_name, enabled, auth_mode,
       credentials_ref, capabilities::text, sync_status, last_sync_at,
       COALESCE(last_error_code, ''), created_at, updated_at,
       COALESCE(base_url, ''), COALESCE(api_token, ''),
       COALESCE(auth_username, ''), COALESCE(auth_client_id, ''), COALESCE(auth_token_url, ''), COALESCE(auth_secret, ''),
       COALESCE(webhook_secret, ''), COALESCE(pull_interval_seconds, 1800), last_pulled_at
FROM integration_providers
WHERE provider_type = 'task_tracker' AND enabled = true
`
	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list task trackers: %w", err)
	}
	defer rows.Close()

	var providers []domain.IntegrationProvider
	for rows.Next() {
		p, err := ScanIntegrationProvider(rows)
		if err != nil {
			return nil, err
		}
		providers = append(providers, p)
	}
	return providers, rows.Err()
}
