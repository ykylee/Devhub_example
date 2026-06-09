-- 000042: api_keys 신규 테이블
-- sprint: feat/work_260609-k-api-key-management
-- ADR-0029 §6 (f) P3 multi-key 관리 + [`docs/planning/api-key-management-sprint-plan.md`](../planning/api-key-management-sprint-plan.md) §3.1
-- 결정: DB (sha256 hash + key prefix) — 회수/만료/audit/last_used_at/CIDR allowlist
-- 1-active-key + N-grace-period pattern 의 기반.

BEGIN;

-- (1) api_keys 신규 테이블
-- key_hash: sha256(raw_key) 의 32 byte binary. raw key 는 1회만 응답 (POST 응답) — 이후 절대 복구 불가.
-- key_prefix: 앞 8 char (예: "dhk_aB3x") — 운영자 식별/검색용. SELECT 시 항상 노출.
-- created_by: 발급한 actor login.
-- last_used_at: best-effort UPDATE (auth middleware 매 호출).
-- revoked_at / revoked_by: 회수 시각 + 회수한 actor login. 회수 후 미사용.
-- expires_at: 만료 시각. NULL = 무기한.
-- allowed_cidrs: CIDR allowlist (text[]). NULL = all IPs 허용.
-- key_hash active uniq: 같은 hash 의 active key 중복 방지. revoked key 는 uniq 제외
-- (revoked 후 재발급 시 hash 같을 수 없으므로 — crypto/rand 256 bit).
CREATE TABLE IF NOT EXISTS public.api_keys (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,
    key_prefix text NOT NULL,
    key_hash bytea NOT NULL,
    created_by text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    last_used_at timestamptz NULL,
    revoked_at timestamptz NULL,
    revoked_by text NULL,
    expires_at timestamptz NULL,
    allowed_cidrs text[] NULL,
    CONSTRAINT api_keys_key_prefix_not_empty CHECK (length(key_prefix) >= 4),
    CONSTRAINT api_keys_key_hash_32_bytes CHECK (octet_length(key_hash) = 32),
    CONSTRAINT api_keys_name_not_empty CHECK (length(name) >= 1)
);

-- (2) unique index — 같은 hash 의 active key 중복 방지.
-- active = revoked_at IS NULL. revoked key 는 uniq 제외 (revoked hash 재사용 시
-- uniq violation 회피 — 실제로 crypto/rand 256bit 으로 hash 같을 확률 0).
CREATE UNIQUE INDEX IF NOT EXISTS api_keys_key_hash_active_uniq
    ON public.api_keys (key_hash)
    WHERE revoked_at IS NULL;

-- (3) 검색 index — 운영자 목록 조회 시 prefix / created_by 필터.
CREATE INDEX IF NOT EXISTS api_keys_key_prefix_idx
    ON public.api_keys (key_prefix);
CREATE INDEX IF NOT EXISTS api_keys_created_by_idx
    ON public.api_keys (created_by);

-- (4) audit 정합 — rotation SOP [`docs/setup/api_key_rotation.md` §6.3] 의
-- `audit.api_key.created/revoked/used` emit 의 source_id 정합.
COMMENT ON TABLE public.api_keys IS 'multi-key API key 관리 (ADR-0029 §6 (f) P3, sprint plan §3.1). key_hash 는 sha256(raw_key) — raw key 는 POST 응답으로 1회만 노출. 이후 GET 은 key_prefix 만 노출. ADR-0029 §6 (g) metric 정합 — last_used_at best-effort UPDATE + auth.api_key.* audit 4종 emit (created/revoked/used).';
COMMENT ON COLUMN public.api_keys.key_hash IS 'sha256(raw_key) 의 32 byte binary. crypto/rand 256bit random — collision 확률 0.';
COMMENT ON COLUMN public.api_keys.key_prefix IS 'raw key 의 앞 8 char (예: dhk_aB3x) — 운영자 식별/검색용. SELECT 시 항상 노출.';
COMMENT ON COLUMN public.api_keys.allowed_cidrs IS 'CIDR allowlist (예: {"10.0.0.0/8", "192.168.0.0/16"}). NULL = all IPs 허용. auth middleware 가 ClientIP 와 CIDR match 검증.';

COMMIT;
