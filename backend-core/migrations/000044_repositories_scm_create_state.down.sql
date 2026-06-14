-- 000044 rollback: repositories sc_create_state 컬럼 drop
BEGIN;

DROP INDEX IF EXISTS idx_repositories_sc_create_pending;
DROP INDEX IF EXISTS idx_repositories_sc_create_failed;
ALTER TABLE public.repositories
    DROP COLUMN IF EXISTS sc_create_at,
    DROP COLUMN IF EXISTS sc_external_id,
    DROP COLUMN IF EXISTS sc_create_error,
    DROP COLUMN IF EXISTS sc_create_status;

COMMIT;
