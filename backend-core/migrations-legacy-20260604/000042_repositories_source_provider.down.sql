DROP INDEX IF EXISTS idx_repositories_provider_id;

ALTER TABLE repositories
    DROP COLUMN IF EXISTS source,
    DROP COLUMN IF EXISTS provider_id,
    DROP COLUMN IF EXISTS description;
