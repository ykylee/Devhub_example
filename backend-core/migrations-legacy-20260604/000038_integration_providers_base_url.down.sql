-- 000038 rollback: integration_providers.base_url 제거.

ALTER TABLE integration_providers
  DROP COLUMN IF EXISTS base_url;
