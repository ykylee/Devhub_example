-- rollback: scm_provider 컬럼 복원 + provider_id 의 provider_key 로 backfill (best-effort).
ALTER TABLE repositories ADD COLUMN IF NOT EXISTS scm_provider TEXT;

UPDATE repositories r
SET scm_provider = p.provider_key
FROM integration_providers p
WHERE r.provider_id IS NOT NULL
  AND p.provider_id = r.provider_id;
