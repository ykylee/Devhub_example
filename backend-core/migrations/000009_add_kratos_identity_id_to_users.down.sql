-- 000009_add_kratos_identity_id_to_users.down.sql
DROP INDEX IF EXISTS users_kratos_identity_id_unique;
ALTER TABLE users DROP COLUMN IF EXISTS kratos_identity_id;
