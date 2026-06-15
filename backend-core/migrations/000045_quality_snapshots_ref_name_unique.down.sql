-- 000045 rollback: quality_snapshots (repository_id, ref_name) partial unique drop
BEGIN;

DROP INDEX IF EXISTS quality_snapshots_repo_ref_unique;

COMMIT;
