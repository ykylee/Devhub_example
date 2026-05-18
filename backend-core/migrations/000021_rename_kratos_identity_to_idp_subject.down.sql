-- 000021 rollback: Rename users.idp_subject -> users.kratos_identity_id.

DROP INDEX IF EXISTS users_idp_subject_unique;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'users' AND column_name = 'idp_subject'
    ) THEN
        ALTER TABLE users RENAME COLUMN idp_subject TO kratos_identity_id;
    END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS users_kratos_identity_id_unique
    ON users (kratos_identity_id)
    WHERE kratos_identity_id IS NOT NULL;
