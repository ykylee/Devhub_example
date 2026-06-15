-- 000045: quality_snapshots (repository_id, ref_name) partial unique for tool='gitea-build'
-- sprint: feat/x5-gitea-pull-store-wire (X-5 production wire follow-up)
-- IMPL-GITEA-PULL-STORE-01: tool='gitea-build' 한정 ON CONFLICT upsert 정합
-- 결정: partial unique (tool 한정) — 다른 tool (e.g. sonarqube) 의 동일 ref_name 허용

BEGIN;

CREATE UNIQUE INDEX IF NOT EXISTS quality_snapshots_repo_ref_unique
ON public.quality_snapshots (repository_id, ref_name)
WHERE tool = 'gitea-build';

COMMIT;
