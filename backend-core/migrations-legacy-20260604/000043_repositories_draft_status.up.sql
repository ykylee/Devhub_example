ALTER TABLE repositories
ADD COLUMN repository_status TEXT NOT NULL DEFAULT 'active',
ADD COLUMN scm_provider TEXT,
ADD COLUMN publish_requested_at TIMESTAMPTZ,
ADD COLUMN published_at TIMESTAMPTZ;

ALTER TABLE repositories
ADD CONSTRAINT repositories_status_check CHECK (repository_status IN ('draft', 'active'));

UPDATE repositories
SET published_at = COALESCE(published_at, updated_at)
WHERE repository_status = 'active' AND published_at IS NULL;

CREATE INDEX repositories_status_updated_at_idx ON repositories (repository_status, updated_at DESC);
