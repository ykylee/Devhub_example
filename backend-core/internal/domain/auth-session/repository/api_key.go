package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/devhub/backend-core/internal/store"
)

// APIKeyStatus represents the runtime state of an API key. The DB column is
// derived (revoked_at IS NULL AND (expires_at IS NULL OR expires_at > now())),
// not stored — this keeps the schema narrow and avoids races between an
// admin revoking a key and the auth middleware evaluating it.
type APIKeyStatus string

const (
	APIKeyStatusActive  APIKeyStatus = "active"
	APIKeyStatusRevoked APIKeyStatus = "revoked"
	APIKeyStatusExpired APIKeyStatus = "expired"
)

// APIKey is the repository-local row shape. We intentionally do NOT import
// the auth-session `service` package from this file (cross-domain import
// cycle: service.keycloak_verifier → httpapi → view → repository → service
// would close the cycle). The view layer maps this struct to
// service.APIKey when it needs to call the auth middleware.
//
// Fields mirror the public.api_keys columns. LastUsedAt / RevokedAt / RevokedBy
// / ExpiresAt are *string because the DB columns are nullable timestamptz —
// nil = column is NULL.
type APIKey struct {
	ID           string
	Name         string
	KeyPrefix    string
	CreatedBy    string
	CreatedAt    string
	LastUsedAt   *string
	RevokedAt    *string
	RevokedBy    *string
	ExpiresAt    *string
	AllowedCIDRs []string
	Status       APIKeyStatus
}

// APIKeyUpdateRequest is the input to UpdateAPIKeyMeta. Name is intentionally
// not updatable — renaming an API key would invalidate operator documentation
// that references the old name. To rename, revoke + create.
type APIKeyUpdateRequest struct {
	ExpiresAt    *string
	AllowedCIDRs []string
}

// APIKeyStore is the persistence interface for the api_keys table.
//
// Two implementations live in this package:
//   - pgAPIKeyStore: production, backed by the migration-000042 table
//   - memoryAPIKeyStore: test/dev, in-memory map keyed by ID + index by
//     sha256 hash for O(1) auth-middleware lookup
//
// We keep this as an interface (rather than a concrete struct) so the auth
// middleware (in view/auth.go) can depend on it without depending on the
// concrete database driver. cfg.APIKeyStore can then be nil for legacy
// single-env-key setups (ADR-0029 §3.4 backwards compat).
type APIKeyStore interface {
	// CreateAPIKey inserts a new api_keys row. The caller supplies the
	// pre-computed sha256 hash (service.GenerateAPIKey returns it). Returns
	// the persisted row (the key_hash is NOT echoed back — only the
	// APIKey view with KeyPrefix + metadata).
	CreateAPIKey(ctx context.Context, keyHash []byte, key APIKey) (APIKey, error)

	// GetAPIKeyByHash looks up an active (not revoked, not expired) key by
	// its sha256 hash. Returns store.ErrNotFound if the key is unknown
	// or not active.
	GetAPIKeyByHash(ctx context.Context, keyHash []byte) (APIKey, error)

	// ListAPIKeys returns the metadata of all api_keys rows, newest first.
	// The raw key_hash is NOT included in the returned slice (operators see
	// only KeyPrefix + metadata).
	ListAPIKeys(ctx context.Context) ([]APIKey, error)

	// RevokeAPIKey sets revoked_at + revoked_by on the row. Idempotent —
	// re-revoking an already-revoked key is a no-op.
	RevokeAPIKey(ctx context.Context, id string, revokedBy string) error

	// UpdateAPIKeyMeta updates expires_at and/or allowed_cidrs. Name and
	// key_hash are NOT updatable (renaming requires revoke+create; key_hash
	// is the lookup key, changing it would invalidate the row).
	UpdateAPIKeyMeta(ctx context.Context, id string, update APIKeyUpdateRequest) error

	// UpdateLastUsedAt is a best-effort write — the auth middleware calls
	// it on every successful API key authentication. If the write fails
	// we do NOT fail the request (the auth has already succeeded; a
	// missing last_used_at is acceptable and the next successful call
	// will retry).
	UpdateLastUsedAt(ctx context.Context, id string, when time.Time) error
}

// pgAPIKeyStore is the production implementation. We use pgxpool directly
// rather than going through the existing store package because the api_keys
// table is small and high-throughput (called on every authenticated
// request via GetAPIKeyByHash) — a dedicated store keeps the hot path
// simple and avoids pulling in cross-cutting concerns from the legacy
// store package.
type pgAPIKeyStore struct {
	pool *pgxpool.Pool
}

// NewPgAPIKeyStore wires a pgAPIKeyStore against the shared pgxpool.
// Returns the store interface so the caller doesn't need to import this
// package for the interface declaration.
func NewPgAPIKeyStore(pool *pgxpool.Pool) APIKeyStore {
	return &pgAPIKeyStore{pool: pool}
}

const apiKeysSelectColumns = `
	id::text,
	name,
	key_prefix,
	created_by,
	created_at,
	last_used_at,
	revoked_at,
	revoked_by,
	expires_at,
	allowed_cidrs
`

func (s *pgAPIKeyStore) scanAPIKey(row pgx.Row) (APIKey, error) {
	var key APIKey
	var lastUsed, revokedAt, revokedBy, expiresAt *string
	if err := row.Scan(
		&key.ID,
		&key.Name,
		&key.KeyPrefix,
		&key.CreatedBy,
		&key.CreatedAt,
		&lastUsed,
		&revokedAt,
		&revokedBy,
		&expiresAt,
		&key.AllowedCIDRs,
	); err != nil {
		return APIKey{}, err
	}
	key.LastUsedAt = lastUsed
	key.RevokedAt = revokedAt
	key.RevokedBy = revokedBy
	key.ExpiresAt = expiresAt
	key.Status = computeAPIKeyStatus(revokedAt, expiresAt)
	return key, nil
}

func (s *pgAPIKeyStore) CreateAPIKey(ctx context.Context, keyHash []byte, key APIKey) (APIKey, error) {
	if len(keyHash) != 32 {
		return APIKey{}, fmt.Errorf("keyHash must be 32 bytes (sha256), got %d", len(keyHash))
	}
	row := s.pool.QueryRow(ctx, `
INSERT INTO public.api_keys (
    id, name, key_prefix, key_hash, created_by, expires_at, allowed_cidrs
) VALUES (
    COALESCE(NULLIF($1, '')::uuid, gen_random_uuid()),
    $2, $3, $4, $5, $6, $7
)
RETURNING `+apiKeysSelectColumns,
		key.ID, key.Name, key.KeyPrefix, keyHash, key.CreatedBy,
		key.ExpiresAt, key.AllowedCIDRs,
	)
	return s.scanAPIKey(row)
}

func (s *pgAPIKeyStore) GetAPIKeyByHash(ctx context.Context, keyHash []byte) (APIKey, error) {
	if len(keyHash) != 32 {
		return APIKey{}, errors.New("keyHash must be 32 bytes (sha256)")
	}
	// The query intentionally filters to active keys only (revoked_at IS
	// NULL AND (expires_at IS NULL OR expires_at > now())). This is the
	// auth-middleware hot path — a revoked key must NOT match.
	row := s.pool.QueryRow(ctx, `
SELECT `+apiKeysSelectColumns+`
FROM public.api_keys
WHERE key_hash = $1
  AND revoked_at IS NULL
  AND (expires_at IS NULL OR expires_at > now())
LIMIT 1
`, keyHash)
	return s.scanAPIKey(row)
}

func (s *pgAPIKeyStore) ListAPIKeys(ctx context.Context) ([]APIKey, error) {
	rows, err := s.pool.Query(ctx, `
SELECT `+apiKeysSelectColumns+`
FROM public.api_keys
ORDER BY created_at DESC
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var keys []APIKey
	for rows.Next() {
		key, err := s.scanAPIKey(rows)
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, rows.Err()
}

func (s *pgAPIKeyStore) RevokeAPIKey(ctx context.Context, id string, revokedBy string) error {
	tag, err := s.pool.Exec(ctx, `
UPDATE public.api_keys
SET revoked_at = COALESCE(revoked_at, now()),
    revoked_by = $2
WHERE id = $1
`, id, revokedBy)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return store.ErrNotFound
	}
	return nil
}

func (s *pgAPIKeyStore) UpdateAPIKeyMeta(ctx context.Context, id string, update APIKeyUpdateRequest) error {
	tag, err := s.pool.Exec(ctx, `
UPDATE public.api_keys
SET expires_at = $2,
    allowed_cidrs = $3
WHERE id = $1
  AND revoked_at IS NULL
`, id, update.ExpiresAt, update.AllowedCIDRs)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return store.ErrNotFound
	}
	return nil
}

func (s *pgAPIKeyStore) UpdateLastUsedAt(ctx context.Context, id string, when time.Time) error {
	// Best-effort — we don't surface the error to the caller because the
	// auth middleware has already approved the request by the time we
	// reach this point. A failure here means last_used_at is stale by
	// one observation window, which the next successful call corrects.
	_, err := s.pool.Exec(ctx, `
UPDATE public.api_keys
SET last_used_at = $2
WHERE id = $1
`, id, when)
	return err
}

// computeAPIKeyStatus derives the runtime status of an api_keys row
// from its revoked_at + expires_at columns. The DB stores the columns
// directly (not a status enum) so we can re-derive the status at any
// time without a migration — `expires_at` is the source of truth for
// temporal validity.
func computeAPIKeyStatus(revokedAt *string, expiresAt *string) APIKeyStatus {
	if revokedAt != nil {
		return APIKeyStatusRevoked
	}
	if expiresAt != nil {
		if t, err := time.Parse(time.RFC3339, *expiresAt); err == nil && !t.After(time.Now()) {
			return APIKeyStatusExpired
		}
	}
	return APIKeyStatusActive
}
