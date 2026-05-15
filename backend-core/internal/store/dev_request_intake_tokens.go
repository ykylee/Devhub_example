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

// CreateDevRequestIntakeToken은 admin 발급 흐름 (sprint claude/work_260515-o, ADR-0014).
// caller (handler) 가 이미 plain 토큰을 SHA-256(hex) 으로 hash 해서 HashedToken 에 전달한다.
// AllowedIPs 는 JSONB 컬럼 — caller 가 비어 있지 않은 CIDR 배열을 보장한다.
// CreatedBy 가 들어가지 않은 row 는 운영 추적이 어려우므로 caller 가 채워야 한다.
// UNIQUE (hashed_token) 위반 시 ErrConflict.
func (s *PostgresStore) CreateDevRequestIntakeToken(ctx context.Context, tok domain.DevRequestIntakeToken) (domain.DevRequestIntakeToken, error) {
	allowedIPs, err := json.Marshal(tok.AllowedIPs)
	if err != nil {
		return domain.DevRequestIntakeToken{}, fmt.Errorf("encode allowed_ips: %w", err)
	}
	const query = `
INSERT INTO dev_request_intake_tokens (client_label, hashed_token, allowed_ips, source_system, created_by)
VALUES ($1, $2, $3::jsonb, $4, $5)
RETURNING token_id::text, client_label, hashed_token, allowed_ips, source_system,
          created_at, created_by, last_used_at, revoked_at`

	row := s.pool.QueryRow(ctx, query, tok.ClientLabel, tok.HashedToken, allowedIPs, tok.SourceSystem, tok.CreatedBy)
	created, err := scanIntakeToken(row)
	if isUniqueViolation(err) {
		return domain.DevRequestIntakeToken{}, ErrConflict
	}
	if err != nil {
		return domain.DevRequestIntakeToken{}, fmt.Errorf("create intake token: %w", err)
	}
	return created, nil
}

// ListDevRequestIntakeTokens는 admin 목록 조회. revoked 포함, created_at DESC.
// hashed_token 은 도메인 객체에 들어가지만 handler 가 응답 매핑 시 제외한다.
func (s *PostgresStore) ListDevRequestIntakeTokens(ctx context.Context) ([]domain.DevRequestIntakeToken, error) {
	const query = `
SELECT token_id::text, client_label, hashed_token, allowed_ips, source_system,
       created_at, created_by, last_used_at, revoked_at
FROM dev_request_intake_tokens
ORDER BY created_at DESC`
	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list intake tokens: %w", err)
	}
	defer rows.Close()
	out := make([]domain.DevRequestIntakeToken, 0)
	for rows.Next() {
		tok, err := scanIntakeToken(rows)
		if err != nil {
			return nil, fmt.Errorf("scan intake token: %w", err)
		}
		out = append(out, tok)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate intake tokens: %w", err)
	}
	return out, nil
}

// RevokeDevRequestIntakeToken은 admin revoke (set revoked_at). 이미 revoked 인 row 는
// 그대로 두고 현재 row 반환 (idempotent). 존재하지 않으면 ErrNotFound.
func (s *PostgresStore) RevokeDevRequestIntakeToken(ctx context.Context, tokenID string) (domain.DevRequestIntakeToken, error) {
	const query = `
UPDATE dev_request_intake_tokens
SET revoked_at = COALESCE(revoked_at, NOW())
WHERE token_id = $1::uuid
RETURNING token_id::text, client_label, hashed_token, allowed_ips, source_system,
          created_at, created_by, last_used_at, revoked_at`
	row := s.pool.QueryRow(ctx, query, tokenID)
	tok, err := scanIntakeToken(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.DevRequestIntakeToken{}, ErrNotFound
	}
	if err != nil {
		return domain.DevRequestIntakeToken{}, fmt.Errorf("revoke intake token: %w", err)
	}
	return tok, nil
}

func scanIntakeToken(row pgx.Row) (domain.DevRequestIntakeToken, error) {
	var (
		tok        domain.DevRequestIntakeToken
		allowedIPs []byte
		lastUsedAt *time.Time
		revokedAt  *time.Time
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
		return domain.DevRequestIntakeToken{}, err
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
