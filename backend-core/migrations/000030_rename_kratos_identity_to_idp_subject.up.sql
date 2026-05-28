-- 000030: Rename users.kratos_identity_id -> users.idp_subject for
-- provider-agnostic identity mapping (Keycloak-only transition).

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'users' AND column_name = 'kratos_identity_id'
    ) THEN
        ALTER TABLE users RENAME COLUMN kratos_identity_id TO idp_subject;
    END IF;
END $$;

DROP INDEX IF EXISTS users_kratos_identity_id_unique;
CREATE UNIQUE INDEX IF NOT EXISTS users_idp_subject_unique
    ON users (idp_subject)
    WHERE idp_subject IS NOT NULL;
