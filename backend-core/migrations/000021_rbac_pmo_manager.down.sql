-- 000021 down: pmo_manager role + CHECK constraint 복구.

DELETE FROM rbac_policies WHERE role_id = 'pmo_manager';

ALTER TABLE rbac_policies
    DROP CONSTRAINT IF EXISTS rbac_policies_role_id_format;

ALTER TABLE rbac_policies
    ADD CONSTRAINT rbac_policies_role_id_format CHECK (
        role_id IN ('developer', 'manager', 'system_admin')
        OR role_id ~ '^custom-[a-z0-9][a-z0-9_-]{0,62}$'
    );
