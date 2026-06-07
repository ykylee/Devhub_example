-- 000005_ci_runs_idempotency_and_runner.up.sql
-- Sprint: mvs/work_260607-h-486-ci-runs-api (N-7 / P0-4)
-- Issue: #486 — POST /api/v1/ci-runs 의 v1.0 spec 정합
--   - 신규 column: runner (Gitea runner 식별자, optional)
--   - 신규 unique index: (repository_id, commit_sha, status, started_at)
--     → Gitea Actions webhook 중복 ingest 안전 (idempotency)
--     → 기존 external_id UNIQUE 은 UpsertCIRun / ciRunLogs GET 호환 위해 유지
-- 주의: commit_sha / started_at nullable. Postgres unique index 는 NULL 을 distinct
--       로 처리하므로, (repo, NULL commit, NULL start) 조합은 다중 허용 — webhook
--       미제공 케이스 호환.
-- 주의: 기존 external_id UNIQUE 가 살아있으므로, 본 index 가 추가 충돌을 유발하지
--       는지 운영 staging 1주 monitoring 필요 (PR #476 이후 본 PR 머지 시점).

BEGIN;

-- 1) runner column 추가 (optional)
ALTER TABLE public.ci_runs
    ADD COLUMN IF NOT EXISTS runner text;

-- 2) 신규 idempotency unique index
--    동일 (repository, commit, status, start) 조합은 CI Run 1 row 만 허용.
--    Gitea webhook 재시도 / 중복 delivery 가 같은 row 로 dedup.
CREATE UNIQUE INDEX IF NOT EXISTS ci_runs_idempotency_idx
    ON public.ci_runs (repository_id, commit_sha, status, started_at);

COMMIT;
