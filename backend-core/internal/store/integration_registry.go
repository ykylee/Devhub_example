package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/jackc/pgx/v5"
)

type IntegrationProviderListOptions struct {
	ProviderType domain.IntegrationProviderType
	Enabled      *bool
	Limit        int
	Offset       int
}

type IntegrationBindingListOptions struct {
	ScopeType    domain.IntegrationScopeType
	ScopeID      string
	ProviderType domain.IntegrationProviderType
	Enabled      *bool
	Limit        int
	Offset       int
}

func scanIntegrationProvider(row pgx.Row) (domain.IntegrationProvider, error) {
	var p domain.IntegrationProvider
	var capsJSON []byte
	if err := row.Scan(
		&p.ID,
		&p.ProviderKey,
		&p.ProviderType,
		&p.DisplayName,
		&p.Enabled,
		&p.AuthMode,
		&p.CredentialsRef,
		&capsJSON,
		&p.SyncStatus,
		&p.LastSyncAt,
		&p.LastErrorCode,
		&p.CreatedAt,
		&p.UpdatedAt,
	); err != nil {
		return domain.IntegrationProvider{}, err
	}
	if len(capsJSON) > 0 {
		if err := json.Unmarshal(capsJSON, &p.Capabilities); err != nil {
			return domain.IntegrationProvider{}, fmt.Errorf("decode capabilities: %w", err)
		}
	}
	return p, nil
}

func scanIntegrationBinding(row pgx.Row) (domain.IntegrationBinding, error) {
	var b domain.IntegrationBinding
	if err := row.Scan(
		&b.ID,
		&b.ScopeType,
		&b.ScopeID,
		&b.ProviderID,
		&b.ExternalKey,
		&b.Policy,
		&b.Enabled,
		&b.CreatedAt,
		&b.UpdatedAt,
	); err != nil {
		return domain.IntegrationBinding{}, err
	}
	return b, nil
}

func (s *PostgresStore) ListIntegrationProviders(ctx context.Context, opts IntegrationProviderListOptions) ([]domain.IntegrationProvider, int, error) {
	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	const countQuery = `
SELECT COUNT(*)
FROM integration_providers
WHERE ($1 = '' OR provider_type = $1)
  AND ($2::boolean IS NULL OR enabled = $2::boolean)`
	var total int
	if err := s.pool.QueryRow(ctx, countQuery, string(opts.ProviderType), opts.Enabled).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count integration providers: %w", err)
	}

	const query = `
SELECT
	provider_id::text,
	provider_key,
	provider_type,
	display_name,
	enabled,
	auth_mode,
	credentials_ref,
	capabilities::text,
	sync_status,
	last_sync_at,
	COALESCE(last_error_code, ''),
	created_at,
	updated_at
FROM integration_providers
WHERE ($3 = '' OR provider_type = $3)
  AND ($4::boolean IS NULL OR enabled = $4::boolean)
ORDER BY created_at DESC
LIMIT $1 OFFSET $2`
	rows, err := s.pool.Query(ctx, query, limit, offset, string(opts.ProviderType), opts.Enabled)
	if err != nil {
		return nil, 0, fmt.Errorf("list integration providers: %w", err)
	}
	defer rows.Close()

	out := make([]domain.IntegrationProvider, 0, limit)
	for rows.Next() {
		p, err := scanIntegrationProvider(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan integration provider: %w", err)
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate integration providers: %w", err)
	}
	return out, total, nil
}

func (s *PostgresStore) GetIntegrationProviderByID(ctx context.Context, providerID string) (domain.IntegrationProvider, error) {
	const query = `
SELECT
	provider_id::text,
	provider_key,
	provider_type,
	display_name,
	enabled,
	auth_mode,
	credentials_ref,
	capabilities::text,
	sync_status,
	last_sync_at,
	COALESCE(last_error_code, ''),
	created_at,
	updated_at
FROM integration_providers
WHERE provider_id = $1::uuid`
	p, err := scanIntegrationProvider(s.pool.QueryRow(ctx, query, providerID))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.IntegrationProvider{}, ErrNotFound
	}
	if err != nil {
		return domain.IntegrationProvider{}, fmt.Errorf("get integration provider: %w", err)
	}
	return p, nil
}

func (s *PostgresStore) GetIntegrationProviderByKey(ctx context.Context, providerKey string) (domain.IntegrationProvider, error) {
	const query = `
SELECT
	provider_id::text,
	provider_key,
	provider_type,
	display_name,
	enabled,
	auth_mode,
	credentials_ref,
	capabilities::text,
	sync_status,
	last_sync_at,
	COALESCE(last_error_code, ''),
	created_at,
	updated_at
FROM integration_providers
WHERE provider_key = $1`
	p, err := scanIntegrationProvider(s.pool.QueryRow(ctx, query, providerKey))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.IntegrationProvider{}, ErrNotFound
	}
	if err != nil {
		return domain.IntegrationProvider{}, fmt.Errorf("get integration provider by key: %w", err)
	}
	return p, nil
}

func (s *PostgresStore) CreateIntegrationProvider(ctx context.Context, p domain.IntegrationProvider) (domain.IntegrationProvider, error) {
	caps, err := json.Marshal(p.Capabilities)
	if err != nil {
		return domain.IntegrationProvider{}, fmt.Errorf("marshal capabilities: %w", err)
	}
	const query = `
INSERT INTO integration_providers (
	provider_key, provider_type, display_name, enabled, auth_mode,
	credentials_ref, capabilities, sync_status
) VALUES (
	$1, $2, $3, $4, $5, $6, $7::jsonb, $8
)
RETURNING
	provider_id::text,
	provider_key,
	provider_type,
	display_name,
	enabled,
	auth_mode,
	credentials_ref,
	capabilities::text,
	sync_status,
	last_sync_at,
	COALESCE(last_error_code, ''),
	created_at,
	updated_at`
	created, err := scanIntegrationProvider(s.pool.QueryRow(
		ctx,
		query,
		p.ProviderKey,
		string(p.ProviderType),
		p.DisplayName,
		p.Enabled,
		string(p.AuthMode),
		p.CredentialsRef,
		string(caps),
		p.SyncStatus,
	))
	if isUniqueViolation(err) {
		return domain.IntegrationProvider{}, ErrConflict
	}
	if err != nil {
		return domain.IntegrationProvider{}, fmt.Errorf("create integration provider: %w", err)
	}
	return created, nil
}

func (s *PostgresStore) UpdateIntegrationProvider(ctx context.Context, p domain.IntegrationProvider) (domain.IntegrationProvider, error) {
	caps, err := json.Marshal(p.Capabilities)
	if err != nil {
		return domain.IntegrationProvider{}, fmt.Errorf("marshal capabilities: %w", err)
	}
	const query = `
UPDATE integration_providers
SET display_name = $2,
	enabled = $3,
	credentials_ref = $4,
	capabilities = $5::jsonb,
	updated_at = NOW()
WHERE provider_id = $1::uuid
RETURNING
	provider_id::text,
	provider_key,
	provider_type,
	display_name,
	enabled,
	auth_mode,
	credentials_ref,
	capabilities::text,
	sync_status,
	last_sync_at,
	COALESCE(last_error_code, ''),
	created_at,
	updated_at`
	updated, err := scanIntegrationProvider(s.pool.QueryRow(
		ctx,
		query,
		p.ID,
		p.DisplayName,
		p.Enabled,
		p.CredentialsRef,
		string(caps),
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.IntegrationProvider{}, ErrNotFound
	}
	if err != nil {
		return domain.IntegrationProvider{}, fmt.Errorf("update integration provider: %w", err)
	}
	return updated, nil
}

func (s *PostgresStore) CreateIntegrationSyncJob(ctx context.Context, providerID string, requestedBy string) (string, error) {
	const query = `
INSERT INTO integration_sync_jobs (provider_id, requested_by, status)
VALUES ($1::uuid, NULLIF($2, ''), 'queued')
RETURNING job_id::text`
	var jobID string
	if err := s.pool.QueryRow(ctx, query, providerID, requestedBy).Scan(&jobID); err != nil {
		if isForeignKeyViolation(err) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("create integration sync job: %w", err)
	}
	return jobID, nil
}

func (s *PostgresStore) ListIntegrationBindings(ctx context.Context, opts IntegrationBindingListOptions) ([]domain.IntegrationBinding, int, error) {
	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}
	const countQuery = `
SELECT COUNT(*)
FROM integration_bindings b
JOIN integration_providers p ON p.provider_id = b.provider_id
WHERE ($1 = '' OR b.scope_type = $1)
  AND ($2 = '' OR b.scope_id = $2)
  AND ($3 = '' OR p.provider_type = $3)
  AND ($4::boolean IS NULL OR b.enabled = $4::boolean)`
	var total int
	if err := s.pool.QueryRow(ctx, countQuery,
		string(opts.ScopeType), opts.ScopeID, string(opts.ProviderType), opts.Enabled).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count integration bindings: %w", err)
	}

	const query = `
SELECT
	b.binding_id::text,
	b.scope_type,
	b.scope_id,
	b.provider_id::text,
	b.external_key,
	b.policy,
	b.enabled,
	b.created_at,
	b.updated_at
FROM integration_bindings b
JOIN integration_providers p ON p.provider_id = b.provider_id
WHERE ($3 = '' OR b.scope_type = $3)
  AND ($4 = '' OR b.scope_id = $4)
  AND ($5 = '' OR p.provider_type = $5)
  AND ($6::boolean IS NULL OR b.enabled = $6::boolean)
ORDER BY b.created_at DESC
LIMIT $1 OFFSET $2`
	rows, err := s.pool.Query(ctx, query,
		limit, offset, string(opts.ScopeType), opts.ScopeID, string(opts.ProviderType), opts.Enabled)
	if err != nil {
		return nil, 0, fmt.Errorf("list integration bindings: %w", err)
	}
	defer rows.Close()

	out := make([]domain.IntegrationBinding, 0, limit)
	for rows.Next() {
		b, err := scanIntegrationBinding(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan integration binding: %w", err)
		}
		out = append(out, b)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate integration bindings: %w", err)
	}
	return out, total, nil
}

func (s *PostgresStore) CreateIntegrationBinding(ctx context.Context, b domain.IntegrationBinding) (domain.IntegrationBinding, error) {
	const query = `
INSERT INTO integration_bindings (
	scope_type, scope_id, provider_id, external_key, policy, enabled
) VALUES (
	$1, $2, $3::uuid, $4, $5, $6
)
RETURNING
	binding_id::text,
	scope_type,
	scope_id,
	provider_id::text,
	external_key,
	policy,
	enabled,
	created_at,
	updated_at`
	created, err := scanIntegrationBinding(s.pool.QueryRow(
		ctx,
		query,
		string(b.ScopeType),
		b.ScopeID,
		b.ProviderID,
		b.ExternalKey,
		string(b.Policy),
		b.Enabled,
	))
	if isUniqueViolation(err) || isForeignKeyViolation(err) || isCheckViolation(err, "") {
		return domain.IntegrationBinding{}, ErrConflict
	}
	if err != nil {
		return domain.IntegrationBinding{}, fmt.Errorf("create integration binding: %w", err)
	}
	return created, nil
}
