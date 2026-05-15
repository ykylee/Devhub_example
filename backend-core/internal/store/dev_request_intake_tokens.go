package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/jackc/pgx/v5"
)

// LookupDevRequestIntakeToken은 hashedToken (SHA-256 hex) 으로 row 조회.
// caller 의 IP CIDR 검증과 revoke 확인은 handler 책임.
func (s *PostgresStore) LookupDevRequestIntakeToken(ctx context.Context, hashedToken string) (domain.DevRequestIntakeToken, error) {
	const query = `
SELECT token_id::text, client_label, hashed_token, allowed_ips, source_system,
       created_at, created_by, last_used_at, revoked_at
FROM dev_request_intake_tokens
WHERE hashed_token = $1`
	row := s.pool.QueryRow(ctx, query, hashedToken)

	var (
		tok          domain.DevRequestIntakeToken
		allowedIPs   []byte
		lastUsedAt   *time.Time
		revokedAt    *time.Time
	)
	if err := row.Scan(
		&tok.TokenID,
		&tok.ClientLabel,
		&tok.HashedToken,
		&allowedIPs,
		&tok.SourceSystem,
		&tok.CreatedAt,
		&tok.CreatedBy,
		&lastUsedAt,
		&revokedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.DevRequestIntakeToken{}, ErrNotFound
		}
		return domain.DevRequestIntakeToken{}, fmt.Errorf("lookup intake token: %w", err)
	}
	tok.LastUsedAt = lastUsedAt
	tok.RevokedAt = revokedAt
	if len(allowedIPs) > 0 {
		if err := json.Unmarshal(allowedIPs, &tok.AllowedIPs); err != nil {
			return domain.DevRequestIntakeToken{}, fmt.Errorf("decode allowed_ips: %w", err)
		}
	}
	return tok, nil
}

// MarkDevRequestIntakeTokenUsed는 인증 성공 시 last_used_at 갱신.
// best-effort — 실패해도 인증 자체는 통과 (audit 보존 우선).
func (s *PostgresStore) MarkDevRequestIntakeTokenUsed(ctx context.Context, tokenID string) error {
	const query = `UPDATE dev_request_intake_tokens SET last_used_at = NOW() WHERE token_id = $1::uuid`
	_, err := s.pool.Exec(ctx, query, tokenID)
	if err != nil {
		return fmt.Errorf("mark intake token used: %w", err)
	}
	return nil
}
