DROP INDEX IF EXISTS repositories_status_updated_at_idx;

ALTER TABLE repositories
DROP CONSTRAINT IF EXISTS repositories_status_check;

ALTER TABLE repositories
DROP COLUMN IF EXISTS published_at,
DROP COLUMN IF EXISTS publish_requested_at,
DROP COLUMN IF EXISTS scm_provider,
DROP COLUMN IF EXISTS repository_status;
