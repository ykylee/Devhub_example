-- 000040 rollback: integration_providers.api_token 제거.

ALTER TABLE integration_providers
  DROP COLUMN IF EXISTS api_token;
