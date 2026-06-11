-- 000007: platforms inbound_source 자동 routing (N-13, ADR-0028 §6 carve a)
-- sprint: feat/work_260611-a-n13-inbound-source-impl
-- ADR-0028 §6 (a) — applications.inbound_source 컬럼 + sync 자동 routing
-- scope: platforms 테이블에 inbound_source_type + inbound_source_config 컬럼 추가
--        + CHECK 제약 + 인덱스 + View 입력 validate 정합.
-- 의존: 없음 (platforms 테이블은 migration 000001 부터 존재).
-- 후속: 000007_idem_001_inbound_source_seed.sql 은 별도 sprint.

BEGIN;

-- (1) platforms 테이블에 inbound_source_type + inbound_source_config 컬럼 추가
ALTER TABLE public.platforms
    ADD COLUMN IF NOT EXISTS inbound_source_type text NOT NULL DEFAULT ''
        CHECK (inbound_source_type IN ('', 'gitea', 'jira', 'other')),
    ADD COLUMN IF NOT EXISTS inbound_source_config jsonb NULL;

-- (2) inbound_source_type 인덱스 (routing 매칭용 partial index)
CREATE INDEX IF NOT EXISTS platforms_inbound_source_type_idx
    ON public.platforms(inbound_source_type)
    WHERE inbound_source_type <> '';

-- (3) inbound_source_config GIN 인덱스 (jsonb key lookup)
CREATE INDEX IF NOT EXISTS platforms_inbound_source_config_gin
    ON public.platforms USING GIN (inbound_source_config)
    WHERE inbound_source_config IS NOT NULL;

-- (4) consistency constraint: type='' 일 땐 config 도 NULL 또는 '{}' 여야
ALTER TABLE public.platforms
    DROP CONSTRAINT IF EXISTS platforms_inbound_source_consistency;
ALTER TABLE public.platforms
    ADD CONSTRAINT platforms_inbound_source_consistency
        CHECK (
            (inbound_source_type = '' AND (inbound_source_config IS NULL OR inbound_source_config = '{}'::jsonb))
            OR (inbound_source_type <> '')
        );

COMMIT;
