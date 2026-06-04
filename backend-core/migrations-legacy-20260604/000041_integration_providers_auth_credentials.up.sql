-- Structured outbound auth credentials per auth_mode (auth_mode-driven UI).
--
-- token mode continues to use api_token (migration 000040) — Authorization: token.
-- basic / app_password use auth_username + auth_secret (HTTP Basic).
-- oauth2 uses auth_client_id + auth_token_url (non-secret) + auth_secret (client_secret).
-- agent uses auth_username (agent identifier).
--
-- auth_secret is treated write-only at the API layer (response exposes
-- auth_secret_set bool, never the raw value) — same pattern as api_token.
-- Non-secret fields (auth_username, auth_client_id, auth_token_url) are returnable.
ALTER TABLE integration_providers
    ADD COLUMN IF NOT EXISTS auth_username  TEXT,
    ADD COLUMN IF NOT EXISTS auth_client_id TEXT,
    ADD COLUMN IF NOT EXISTS auth_token_url TEXT,
    ADD COLUMN IF NOT EXISTS auth_secret    TEXT;
