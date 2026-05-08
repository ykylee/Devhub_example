-- ADR-0002 / PR-G3: replace static role CHECK with a foreign key into rbac_policies
-- so SetSubjectRole can assign custom roles validated by the DB.
ALTER TABLE users DROP CONSTRAINT users_role_check;
ALTER TABLE users
    ADD CONSTRAINT users_role_fkey
    FOREIGN KEY (role) REFERENCES rbac_policies(role_id);
