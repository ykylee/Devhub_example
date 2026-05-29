package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
	"github.com/jackc/pgx/v5"
)

// IntegrationProviderListOptions parameterizes ListIntegrationProviders.
type IntegrationProviderListOptions = store.IntegrationProviderListOptions

// IntegrationBindingListOptions parameterizes ListIntegrationBindings.
type IntegrationBindingListOptions = store.IntegrationBindingListOptions

func ScanIntegrationProvider(row pgx.Row) (domain.IntegrationProvider, error) {
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
		&p.BaseURL,
		&p.APIToken,
		&p.AuthUsername,
		&p.AuthClientID,
		&p.AuthTokenURL,
		&p.AuthSecret,
		&p.WebhookSecret,
		&p.PullIntervalSeconds,
		&p.LastPulledAt,
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

func (r *IntegrationRepository) ListIntegrationProviders(ctx context.Context, opts IntegrationProviderListOptions) ([]domain.IntegrationProvider, int, error) {
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
	if err := r.store.Pool().QueryRow(ctx, countQuery, string(opts.ProviderType), opts.Enabled).Scan(&total); err != nil {
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
	updated_at,
	COALESCE(base_url, ''),
	COALESCE(api_token, ''),
	COALESCE(auth_username, ''),
	COALESCE(auth_client_id, ''),
	COALESCE(auth_token_url, ''),
	COALESCE(auth_secret, ''),
	COALESCE(webhook_secret, ''),
	COALESCE(pull_interval_seconds, 1800),
	last_pulled_at
FROM integration_providers
WHERE ($3 = '' OR provider_type = $3)
  AND ($4::boolean IS NULL OR enabled = $4::boolean)
ORDER BY created_at DESC
LIMIT $1 OFFSET $2`
	rows, err := r.store.Pool().Query(ctx, query, limit, offset, string(opts.ProviderType), opts.Enabled)
	if err != nil {
		return nil, 0, fmt.Errorf("list integration providers: %w", err)
	}
	defer rows.Close()

	out := make([]domain.IntegrationProvider, 0, limit)
	for rows.Next() {
		p, err := ScanIntegrationProvider(rows)
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

func (r *IntegrationRepository) GetIntegrationProviderByID(ctx context.Context, providerID string) (domain.IntegrationProvider, error) {
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
	updated_at,
	COALESCE(base_url, ''),
	COALESCE(api_token, ''),
	COALESCE(auth_username, ''),
	COALESCE(auth_client_id, ''),
	COALESCE(auth_token_url, ''),
	COALESCE(auth_secret, ''),
	COALESCE(webhook_secret, ''),
	COALESCE(pull_interval_seconds, 1800),
	last_pulled_at
FROM integration_providers
WHERE provider_id = $1::uuid`
	p, err := ScanIntegrationProvider(r.store.Pool().QueryRow(ctx, query, providerID))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.IntegrationProvider{}, store.ErrNotFound
	}
	if err != nil {
		return domain.IntegrationProvider{}, fmt.Errorf("get integration provider: %w", err)
	}
	return p, nil
}

func (r *IntegrationRepository) GetIntegrationProviderByKey(ctx context.Context, providerKey string) (domain.IntegrationProvider, error) {
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
	updated_at,
	COALESCE(base_url, ''),
	COALESCE(api_token, ''),
	COALESCE(auth_username, ''),
	COALESCE(auth_client_id, ''),
	COALESCE(auth_token_url, ''),
	COALESCE(auth_secret, ''),
	COALESCE(webhook_secret, ''),
	COALESCE(pull_interval_seconds, 1800),
	last_pulled_at
FROM integration_providers
WHERE provider_key = $1`
	p, err := ScanIntegrationProvider(r.store.Pool().QueryRow(ctx, query, providerKey))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.IntegrationProvider{}, store.ErrNotFound
	}
	if err != nil {
		return domain.IntegrationProvider{}, fmt.Errorf("get integration provider by key: %w", err)
	}
	return p, nil
}

func (r *IntegrationRepository) CreateIntegrationProvider(ctx context.Context, p domain.IntegrationProvider) (domain.IntegrationProvider, error) {
	caps, err := json.Marshal(p.Capabilities)
	if err != nil {
		return domain.IntegrationProvider{}, fmt.Errorf("marshal capabilities: %w", err)
	}
	if p.PullIntervalSeconds <= 0 {
		p.PullIntervalSeconds = 1800
	}
	const query = `
INSERT INTO integration_providers (
	provider_key, provider_type, display_name, enabled, auth_mode,
	credentials_ref, capabilities, sync_status, base_url, api_token,
	auth_username, auth_client_id, auth_token_url, auth_secret,
	webhook_secret, pull_interval_seconds, last_pulled_at
) VALUES (
	$1, $2, $3, $4, $5, $6, $7::jsonb, $8, NULLIF($9, ''), NULLIF($10, ''),
	NULLIF($11, ''), NULLIF($12, ''), NULLIF($13, ''), NULLIF($14, ''),
	NULLIF($15, ''), $16, $17
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
	updated_at,
	COALESCE(base_url, ''),
	COALESCE(api_token, ''),
	COALESCE(auth_username, ''),
	COALESCE(auth_client_id, ''),
	COALESCE(auth_token_url, ''),
	COALESCE(auth_secret, ''),
	COALESCE(webhook_secret, ''),
	COALESCE(pull_interval_seconds, 1800),
	last_pulled_at`
	created, err := ScanIntegrationProvider(r.store.Pool().QueryRow(
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
		p.BaseURL,
		p.APIToken,
		p.AuthUsername,
		p.AuthClientID,
		p.AuthTokenURL,
		p.AuthSecret,
		p.WebhookSecret,
		p.PullIntervalSeconds,
		p.LastPulledAt,
	))
	if store.IsUniqueViolation(err) {
		return domain.IntegrationProvider{}, store.ErrConflict
	}
	if err != nil {
		return domain.IntegrationProvider{}, fmt.Errorf("create integration provider: %w", err)
	}
	return created, nil
}

func (r *IntegrationRepository) UpdateIntegrationProvider(ctx context.Context, p domain.IntegrationProvider) (domain.IntegrationProvider, error) {
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
	sync_status = $6,
	last_sync_at = $7,
	last_error_code = NULLIF($8, ''),
	base_url = NULLIF($9, ''),
	api_token = NULLIF($10, ''),
	auth_username = NULLIF($11, ''),
	auth_client_id = NULLIF($12, ''),
	auth_token_url = NULLIF($13, ''),
	auth_secret = NULLIF($14, ''),
	webhook_secret = NULLIF($15, ''),
	pull_interval_seconds = $16,
	last_pulled_at = $17,
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
	updated_at,
	COALESCE(base_url, ''),
	COALESCE(api_token, ''),
	COALESCE(auth_username, ''),
	COALESCE(auth_client_id, ''),
	COALESCE(auth_token_url, ''),
	COALESCE(auth_secret, ''),
	COALESCE(webhook_secret, ''),
	COALESCE(pull_interval_seconds, 1800),
	last_pulled_at`
	updated, err := ScanIntegrationProvider(r.store.Pool().QueryRow(
		ctx,
		query,
		p.ID,
		p.DisplayName,
		p.Enabled,
		p.CredentialsRef,
		string(caps),
		p.SyncStatus,
		p.LastSyncAt,
		p.LastErrorCode,
		p.BaseURL,
		p.APIToken,
		p.AuthUsername,
		p.AuthClientID,
		p.AuthTokenURL,
		p.AuthSecret,
		p.WebhookSecret,
		p.PullIntervalSeconds,
		p.LastPulledAt,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.IntegrationProvider{}, store.ErrNotFound
	}
	if err != nil {
		return domain.IntegrationProvider{}, fmt.Errorf("update integration provider: %w", err)
	}
	return updated, nil
}

func (r *IntegrationRepository) DeleteIntegrationProvider(ctx context.Context, providerID string) error {
	const bindingCountQuery = `SELECT COUNT(*) FROM integration_bindings WHERE provider_id = $1::uuid`
	var bindingCount int
	if err := r.store.Pool().QueryRow(ctx, bindingCountQuery, providerID).Scan(&bindingCount); err != nil {
		return fmt.Errorf("count bindings for provider %s: %w", providerID, err)
	}
	if bindingCount > 0 {
		return store.ErrConflict
	}
	const deleteQuery = `DELETE FROM integration_providers WHERE provider_id = $1::uuid RETURNING provider_id`
	var deletedID string
	if err := r.store.Pool().QueryRow(ctx, deleteQuery, providerID).Scan(&deletedID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return store.ErrNotFound
		}
		return fmt.Errorf("delete integration provider %s: %w", providerID, err)
	}
	return nil
}

func (r *IntegrationRepository) CreateIntegrationSyncJob(ctx context.Context, providerID string, requestedBy string) (string, error) {
	const query = `
INSERT INTO integration_sync_jobs (provider_id, requested_by, status)
VALUES ($1::uuid, NULLIF($2, ''), 'queued')
RETURNING job_id::text`
	var jobID string
	if err := r.store.Pool().QueryRow(ctx, query, providerID, requestedBy).Scan(&jobID); err != nil {
		if store.IsForeignKeyViolation(err) {
			return "", store.ErrNotFound
		}
		return "", fmt.Errorf("create integration sync job: %w", err)
	}
	return jobID, nil
}

func (r *IntegrationRepository) ListIntegrationBindings(ctx context.Context, opts IntegrationBindingListOptions) ([]domain.IntegrationBinding, int, error) {
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
	if err := r.store.Pool().QueryRow(ctx, countQuery,
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
	rows, err := r.store.Pool().Query(ctx, query,
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

func (r *IntegrationRepository) CreateIntegrationBinding(ctx context.Context, b domain.IntegrationBinding) (domain.IntegrationBinding, error) {
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
	created, err := scanIntegrationBinding(r.store.Pool().QueryRow(
		ctx,
		query,
		string(b.ScopeType),
		b.ScopeID,
		b.ProviderID,
		b.ExternalKey,
		string(b.Policy),
		b.Enabled,
	))
	if store.IsUniqueViolation(err) || store.IsForeignKeyViolation(err) || store.IsCheckViolation(err, "") {
		return domain.IntegrationBinding{}, store.ErrConflict
	}
	if err != nil {
		return domain.IntegrationBinding{}, fmt.Errorf("create integration binding: %w", err)
	}
	return created, nil
}

func (r *IntegrationRepository) GetIntegrationBindingByID(ctx context.Context, bindingID string) (domain.IntegrationBinding, error) {
	const query = `
SELECT
	binding_id::text,
	scope_type,
	scope_id,
	provider_id::text,
	external_key,
	policy,
	enabled,
	created_at,
	updated_at
FROM integration_bindings
WHERE binding_id = $1::uuid`
	b, err := scanIntegrationBinding(r.store.Pool().QueryRow(ctx, query, bindingID))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.IntegrationBinding{}, store.ErrNotFound
	}
	if err != nil {
		return domain.IntegrationBinding{}, fmt.Errorf("get integration binding: %w", err)
	}
	return b, nil
}

func (r *IntegrationRepository) UpdateIntegrationBinding(ctx context.Context, b domain.IntegrationBinding) (domain.IntegrationBinding, error) {
	const query = `
UPDATE integration_bindings
SET provider_id = $2::uuid,
	external_key = $3,
	policy = $4,
	enabled = $5,
	updated_at = NOW()
WHERE binding_id = $1::uuid
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
	updated, err := scanIntegrationBinding(r.store.Pool().QueryRow(
		ctx,
		query,
		b.ID,
		b.ProviderID,
		b.ExternalKey,
		string(b.Policy),
		b.Enabled,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.IntegrationBinding{}, store.ErrNotFound
	}
	if store.IsUniqueViolation(err) || store.IsForeignKeyViolation(err) || store.IsCheckViolation(err, "") {
		return domain.IntegrationBinding{}, store.ErrConflict
	}
	if err != nil {
		return domain.IntegrationBinding{}, fmt.Errorf("update integration binding: %w", err)
	}
	return updated, nil
}

func (r *IntegrationRepository) DeleteIntegrationBinding(ctx context.Context, bindingID string) error {
	const query = `DELETE FROM integration_bindings WHERE binding_id = $1::uuid RETURNING binding_id`
	var deletedID string
	if err := r.store.Pool().QueryRow(ctx, query, bindingID).Scan(&deletedID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return store.ErrNotFound
		}
		return fmt.Errorf("delete integration binding %s: %w", bindingID, err)
	}
	return nil
}

func (r *IntegrationRepository) AcquireNextQueuedSyncJob(ctx context.Context) (string, string, error) {
	const query = `
UPDATE integration_sync_jobs
SET status = 'running'
WHERE job_id = (
	SELECT j.job_id
	FROM integration_sync_jobs j
	JOIN integration_providers p ON p.provider_id = j.provider_id
	WHERE j.status = 'queued' AND p.provider_type = 'scm'
	ORDER BY j.created_at ASC
	LIMIT 1
	FOR UPDATE OF j SKIP LOCKED
)
RETURNING job_id::text, provider_id::text`
	var jobID, providerID string
	err := r.store.Pool().QueryRow(ctx, query).Scan(&jobID, &providerID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", store.ErrNotFound
	}
	if err != nil {
		return "", "", fmt.Errorf("acquire next queued sync job: %w", err)
	}
	return jobID, providerID, nil
}

func (r *IntegrationRepository) UpdateIntegrationSyncJobStatus(ctx context.Context, jobID string, status string) error {
	const query = `
UPDATE integration_sync_jobs
SET status = $1
WHERE job_id = $2::uuid`
	tag, err := r.store.Pool().Exec(ctx, query, status, jobID)
	if err != nil {
		return fmt.Errorf("update integration sync job status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return store.ErrNotFound
	}
	return nil
}
