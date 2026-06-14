-- 000043 rollback: repository_pull_state drop
BEGIN;

DROP TRIGGER IF EXISTS trg_repository_pull_state_updated_at ON public.repository_pull_state;
DROP FUNCTION IF EXISTS public.touch_repository_pull_state_updated_at();
DROP INDEX IF EXISTS idx_repository_pull_state_consecutive_failures;
DROP INDEX IF EXISTS idx_repository_pull_state_backoff;
DROP TABLE IF EXISTS public.repository_pull_state;

COMMIT;
