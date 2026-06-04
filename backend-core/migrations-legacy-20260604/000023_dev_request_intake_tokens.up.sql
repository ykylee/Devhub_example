-- 000023_dev_request_intake_tokens.up.sql
-- DREQ 외부 수신 인증 토큰 (ADR-0012 §4.1.1, sprint claude/work_260515-i).
--
-- 외부 시스템마다 발급되는 long-lived bearer token. plain token 은 발급 직후
-- 1회만 admin 에게 노출하고 어디에도 저장하지 않는다. DB 에는 SHA-256 hex 만.
-- IP allowlist (CIDR 배열) 가 2차 방어선.

CREATE TABLE dev_request_intake_tokens (
    token_id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_label    TEXT NOT NULL,
    hashed_token    TEXT NOT NULL UNIQUE,
    allowed_ips     JSONB NOT NULL DEFAULT '[]'::jsonb,
    source_system   TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by      TEXT NOT NULL REFERENCES users(user_id) ON DELETE RESTRICT,
    last_used_at    TIMESTAMPTZ,
    revoked_at      TIMESTAMPTZ,

    CONSTRAINT dev_request_intake_tokens_label_format
        CHECK (client_label ~ '^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$'),

    -- allowed_ips 는 CIDR 문자열 배열만 허용 (JSON array of strings).
    -- 정밀한 CIDR validation 은 application 단에서 처리 (PG 의 cidr type 변환 비용 회피).
    CONSTRAINT dev_request_intake_tokens_allowed_ips_array
        CHECK (jsonb_typeof(allowed_ips) = 'array')
);

CREATE INDEX dev_request_intake_tokens_active_idx
    ON dev_request_intake_tokens (revoked_at)
    WHERE revoked_at IS NULL;
