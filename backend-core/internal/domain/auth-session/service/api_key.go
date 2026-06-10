package service

import (
	"fmt"
	"strings"

	"github.com/devhub/backend-core/internal/shared/authkey"
)

// APIKeyStatus represents the runtime state of an API key. The DB column is
// derived (revoked_at IS NULL AND (expires_at IS NULL OR expires_at > now())),
// not stored — this keeps the schema narrow and avoids races between an
// admin revoking a key and the auth middleware evaluating it.
//
// We keep this as a Go enum for ergonomic use at the boundary; mapping to
// the DB expression happens in api_key.go (repository).
type APIKeyStatus string

const (
	// APIKeyStatusActive — key is valid. Passes auth middleware.
	APIKeyStatusActive APIKeyStatus = "active"
	// APIKeyStatusRevoked — revoked_at set. Fails auth middleware (401).
	APIKeyStatusRevoked APIKeyStatus = "revoked"
	// APIKeyStatusExpired — expires_at < now() and not revoked. Fails auth
	// middleware (401, distinct code from revoked so operators can tell the
	// difference).
	APIKeyStatusExpired APIKeyStatus = "expired"
)

// APIKey is the domain model for a managed API key. We split this from the
// repository row type so callers don't accidentally read or write the
// raw key material — that field only exists in APIKeyCreateResult and
// must be returned exactly once to the caller (see GenerateAPIKey).
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

// APIKeyCreateRequest is the input to CreateAPIKey. The caller cannot
// supply the raw key (it would defeat the point of generating it server-side)
// or the key hash (derived from the raw key).
type APIKeyCreateRequest struct {
	Name         string
	ExpiresAt    *string
	AllowedCIDRs []string
}

// APIKeyUpdateRequest is the input to UpdateAPIKey. Name is intentionally
// not updatable — renaming an API key would invalidate operator
// documentation that references the old name. To rename, revoke + create.
type APIKeyUpdateRequest struct {
	ExpiresAt    *string
	AllowedCIDRs []string
}

// APIKeyCreateResult is what CreateAPIKey returns. The RawKey field is the
// ONLY place the raw key material appears outside the request — it must be
// returned to the caller exactly once and never logged, persisted (beyond
// the hash) or echoed back in subsequent reads. The KeyHash field is the
// 32-byte sha256(rawKey) the repository persists into the api_keys table.
type APIKeyCreateResult struct {
	APIKey  APIKey
	RawKey  string
	KeyHash []byte
}

// GenerateAPIKey 는 service 의 thin re-export of authkey.GenerateAPIKey.
// view layer (admin handler) 가 service package 를 직접 import 하면
// service.keycloak_verifier → httpapi → view → service cycle 이 발생.
// generate 자체는 shared/authkey 에 두고 service 는 그 위에 status/metadata
// defaulting 만 추가.
func GenerateAPIKey(name string, createdBy string, expiresAt *string, allowedCIDRs []string) (APIKeyCreateResult, error) {
	if strings.TrimSpace(name) == "" {
		return APIKeyCreateResult{}, fmt.Errorf("api key name must not be empty")
	}
	if strings.TrimSpace(createdBy) == "" {
		return APIKeyCreateResult{}, fmt.Errorf("createdBy must not be empty")
	}
	rawKey, keyHash, keyPrefix, err := authkey.GenerateRawAPIKey()
	if err != nil {
		return APIKeyCreateResult{}, err
	}
	return APIKeyCreateResult{
		APIKey: APIKey{
			Name:         name,
			KeyPrefix:    keyPrefix,
			CreatedBy:    createdBy,
			AllowedCIDRs: allowedCIDRs,
			ExpiresAt:    expiresAt,
			Status:       APIKeyStatusActive,
		},
		RawKey:  rawKey,
		KeyHash: keyHash,
	}, nil
}
