-- Reverse PR-G3 schema change. Note: requires that no users hold a custom role
-- (rows with role NOT IN ('developer', 'manager', 'system_admin') would trip the
-- recreated CHECK). Reassign those users to a system role before downgrading.
ALTER TABLE users DROP CONSTRAINT users_role_fkey;
ALTER TABLE users
    ADD CONSTRAINT users_role_check
    CHECK (role IN ('developer', 'manager', 'system_admin'));
