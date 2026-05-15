-- 000021 down: pmo_manager role + CHECK constraint 복구.
--
-- users.role 이 rbac_policies(role_id) FK (migration 000006) 를 참조하므로,
-- 운영 중 pmo_manager 로 지정된 사용자가 있는 환경에서는 DELETE 가 FK 위반으로
-- 실패한다 (rollback 자체가 불가능). 따라서 DELETE 직전에 pmo_manager 사용자를
-- 안전한 default 인 'developer' 로 재할당한다 (회복 가능한 변환). 운영 진입 후
-- 다른 default 가 정해지면 그에 맞춰 본 down 을 갱신.
-- codex PR #119 review P1 (sprint claude/work_260515-h).

UPDATE users SET role = 'developer' WHERE role = 'pmo_manager';

DELETE FROM rbac_policies WHERE role_id = 'pmo_manager';

ALTER TABLE rbac_policies
    DROP CONSTRAINT IF EXISTS rbac_policies_role_id_format;

ALTER TABLE rbac_policies
    ADD CONSTRAINT rbac_policies_role_id_format CHECK (
        role_id IN ('developer', 'manager', 'system_admin')
        OR role_id ~ '^custom-[a-z0-9][a-z0-9_-]{0,62}$'
    );
