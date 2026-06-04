ALTER TABLE integration_providers
    DROP COLUMN IF EXISTS auth_username,
    DROP COLUMN IF EXISTS auth_client_id,
    DROP COLUMN IF EXISTS auth_token_url,
    DROP COLUMN IF EXISTS auth_secret;
