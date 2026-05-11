-- 000009_add_kratos_identity_id_to_users.up.sql
--
-- L4-A (work_26_05_11-e): cache the Kratos identity_id on DevHub users so
-- FindIdentityByUserID is O(1) instead of paging through /admin/identities.
-- The column is nullable so existing rows survive the migration; new
-- account.create flows populate it eagerly and other flows backfill on
-- first lookup (httpapi.resolveKratosIdentityID).
ALTER TABLE users
    ADD COLUMN kratos_identity_id TEXT;

-- Partial unique index keeps the column nullable while still guaranteeing
-- one identity ↔ one user mapping. NULL duplicates are ignored. A PG15+
-- UNIQUE ... NULLS NOT DISTINCT would also work but the partial index is
-- portable to earlier minor versions.
CREATE UNIQUE INDEX users_kratos_identity_id_unique
    ON users (kratos_identity_id)
    WHERE kratos_identity_id IS NOT NULL;
