-- 000005_ci_runs_idempotency_and_runner.down.sql
-- Sprint: mvs/work_260607-h-486-ci-runs-api (N-7 / P0-4)
-- DOWN: drop idempotency index + drop runner column.

BEGIN;

DROP INDEX IF EXISTS public.ci_runs_idempotency_idx;

ALTER TABLE public.ci_runs
    DROP COLUMN IF EXISTS runner;

COMMIT;
