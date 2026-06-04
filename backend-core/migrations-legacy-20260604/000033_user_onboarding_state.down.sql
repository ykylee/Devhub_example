DROP INDEX IF EXISTS users_review_status_idx;

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_onboarding_review_consistency,
    DROP CONSTRAINT IF EXISTS users_review_status_check,
    DROP COLUMN IF EXISTS review_status,
    DROP COLUMN IF EXISTS onboarding_completed_at;
