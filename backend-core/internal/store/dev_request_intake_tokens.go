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
       created_at, created_by, last_used_at, revoked_at, expires_at
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
		&tok.ExpiresAt,
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
INSERT INTO dev_request_intake_tokens (client_label, hashed_token, allowed_ips, source_system, created_by, expires_at)
VALUES ($1, $2, $3::jsonb, $4, $5, $6)
RETURNING token_id::text, client_label, hashed_token, allowed_ips, source_system,
          created_at, created_by, last_used_at, revoked_at, expires_at`

	row := s.pool.QueryRow(ctx, query, tok.ClientLabel, tok.HashedToken, allowedIPs, tok.SourceSystem, tok.CreatedBy, tok.ExpiresAt)
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
       created_at, created_by, last_used_at, revoked_at, expires_at
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
          created_at, created_by, last_used_at, revoked_at, expires_at`
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

// UpdateDevRequestIntakeTokenIPs는 admin allowed_ips 수정 (sprint gemini/dreq_e2e_260515 hardening).
// revoke 없이 접근 권한만 동적으로 변경할 때 사용. revoke 된 row 는 ErrConflict 반환
// (ADR-0017 §4.2 — 이미 revoked 이면 mutation 의미 없음. codex hotfix #5 P2 #3 회귀 가드).
//
// ADR-0017 §6 atomicity (sprint claude/work_260518-o) — 이전 2 query 패턴
// (UPDATE WHERE revoked_at IS NULL → ErrNoRows 시 별도 SELECT) 가 between-query
// race window 를 가졌음. 본 구현은 단일 CTE + FOR UPDATE row lock + VALUES anchor
// 패턴으로 atomic 보장:
//   (1) locked CTE 가 token row 를 FOR UPDATE 로 잠그며 revoked_at 도 함께 조회
//   (2) upd CTE 가 WHERE revoked_at IS NULL 가드로 UPDATE 시도
//   (3) (VALUES (1)) AS root(_) anchor 가 단일 row 반환 보장 — locked empty 여도
//       lock_token_id IS NULL 로 not_found 분기 가능 (codex hotfix #6 P2 정정 정합)
// handler 분기: lock_token_id IS NULL → ErrNotFound / lock_revoked_at IS NOT NULL
// → ErrConflict / upd_token_id IS NOT NULL → success.
func (s *PostgresStore) UpdateDevRequestIntakeTokenIPs(ctx context.Context, tokenID string, allowedIPs []string) (domain.DevRequestIntakeToken, error) {
	ipsJSON, err := json.Marshal(allowedIPs)
	if err != nil {
		return domain.DevRequestIntakeToken{}, fmt.Errorf("encode allowed_ips: %w", err)
	}
	const query = `
WITH locked AS (
  SELECT token_id, revoked_at FROM dev_request_intake_tokens
  WHERE token_id = $1::uuid FOR UPDATE
),
upd AS (
  UPDATE dev_request_intake_tokens
  SET allowed_ips = $2::jsonb, updated_at = NOW()
  WHERE token_id = $1::uuid AND revoked_at IS NULL
  RETURNING token_id::text, client_label, hashed_token, allowed_ips, source_system,
            created_at, created_by, last_used_at, revoked_at, expires_at
)
SELECT locked.token_id   AS lock_token_id,
       locked.revoked_at AS lock_revoked_at,
       upd.token_id      AS upd_token_id,
       upd.client_label  AS upd_client_label,
       upd.hashed_token  AS upd_hashed_token,
       upd.allowed_ips   AS upd_allowed_ips,
       upd.source_system AS upd_source_system,
       upd.created_at    AS upd_created_at,
       upd.created_by    AS upd_created_by,
       upd.last_used_at  AS upd_last_used_at,
       upd.revoked_at    AS upd_revoked_at,
       upd.expires_at    AS upd_expires_at
FROM (VALUES (1)) AS root(_)
LEFT JOIN locked ON true
LEFT JOIN upd    ON true`

	var (
		lockTokenID    *string
		lockRevokedAt  *time.Time
		updTokenID     *string
		updClientLabel *string
		updHashedToken *string
		updAllowedIPs  []byte
		updSourceSys   *string
		updCreatedAt   *time.Time
		updCreatedBy   *string
		updLastUsedAt  *time.Time
		updRevokedAt   *time.Time
		updExpiresAt   *time.Time
	)
	if scanErr := s.pool.QueryRow(ctx, query, tokenID, ipsJSON).Scan(
		&lockTokenID,
		&lockRevokedAt,
		&updTokenID,
		&updClientLabel,
		&updHashedToken,
		&updAllowedIPs,
		&updSourceSys,
		&updCreatedAt,
		&updCreatedBy,
		&updLastUsedAt,
		&updRevokedAt,
		&updExpiresAt,
	); scanErr != nil {
		return domain.DevRequestIntakeToken{}, fmt.Errorf("update intake token ips: %w", scanErr)
	}
	if lockTokenID == nil {
		return domain.DevRequestIntakeToken{}, ErrNotFound
	}
	if lockRevokedAt != nil {
		return domain.DevRequestIntakeToken{}, ErrConflict
	}
	if updTokenID == nil {
		// locked exists + not revoked + upd nil — pgx 가 다른 transaction 의 변경을
		// 동시 감지한 극단 케이스 (FOR UPDATE 가 잠그기 직전에 revoked + lock 후 다시
		// 풀린). defense-in-depth — read-after-write 불일치로 분류.
		return domain.DevRequestIntakeToken{}, fmt.Errorf("update intake token ips: unexpected upd nil for token_id=%s", tokenID)
	}
	tok := domain.DevRequestIntakeToken{
		TokenID:      derefString(updTokenID),
		ClientLabel:  derefString(updClientLabel),
		HashedToken:  derefString(updHashedToken),
		SourceSystem: derefString(updSourceSys),
		CreatedAt:    derefTime(updCreatedAt),
		CreatedBy:    derefString(updCreatedBy),
		LastUsedAt:   updLastUsedAt,
		RevokedAt:    updRevokedAt,
		ExpiresAt:    updExpiresAt,
	}
	if len(updAllowedIPs) > 0 {
		if err := json.Unmarshal(updAllowedIPs, &tok.AllowedIPs); err != nil {
			return domain.DevRequestIntakeToken{}, fmt.Errorf("decode allowed_ips: %w", err)
		}
	}
	return tok, nil
}

func derefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func derefTime(p *time.Time) time.Time {
	if p == nil {
		return time.Time{}
	}
	return *p
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
		&tok.ExpiresAt,
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
