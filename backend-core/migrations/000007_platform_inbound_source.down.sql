-- 000007 down: platforms inbound_source 자동 routing (N-13, ADR-0028 §6 carve a) 롤백
-- scope: inbound_source_type + inbound_source_config + 인덱스 + CHECK 제약 제거

BEGIN;

ALTER TABLE public.platforms
    DROP CONSTRAINT IF EXISTS platforms_inbound_source_consistency;

DROP INDEX IF EXISTS public.platforms_inbound_source_config_gin;
DROP INDEX IF EXISTS public.platforms_inbound_source_type_idx;

ALTER TABLE public.platforms
    DROP COLUMN IF EXISTS inbound_source_config,
    DROP COLUMN IF EXISTS inbound_source_type;

COMMIT;
