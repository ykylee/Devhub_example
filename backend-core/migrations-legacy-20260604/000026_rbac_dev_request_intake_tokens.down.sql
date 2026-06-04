-- 000026_rbac_dev_request_intake_tokens.down.sql
-- Drop the dev_request_intake_tokens key from each role's permissions JSONB.

UPDATE rbac_policies
SET permissions = permissions - 'dev_request_intake_tokens',
    updated_at = NOW()
WHERE permissions ? 'dev_request_intake_tokens';
