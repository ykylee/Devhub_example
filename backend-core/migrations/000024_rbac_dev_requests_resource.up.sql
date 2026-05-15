-- 000024_rbac_dev_requests_resource.up.sql
-- DREQ 도메인의 RBAC 자원 `dev_requests` 를 rbac_policies.permissions JSONB 매트릭스에 추가.
-- ARCH-DREQ-04 정책 매핑 (sprint claude/work_260515-i):
--   - developer: view (route gate 만, handler 가 row-level filter)
--   - manager: view
--   - system_admin: view + create + edit + delete
--   - pmo_manager: view + edit (create/delete false; close/reassign 은 handler 가 추가 검증)
--
-- domain.DefaultPermissionMatrix() 와 byte-for-byte 정합.

UPDATE rbac_policies
SET permissions = permissions
        || jsonb_build_object(
            'dev_requests',
            jsonb_build_object('view', TRUE, 'create', FALSE, 'edit', FALSE, 'delete', FALSE)
        ),
    updated_at = NOW()
WHERE role_id IN ('developer', 'manager');

UPDATE rbac_policies
SET permissions = permissions
        || jsonb_build_object(
            'dev_requests',
            jsonb_build_object('view', TRUE, 'create', TRUE, 'edit', TRUE, 'delete', TRUE)
        ),
    updated_at = NOW()
WHERE role_id = 'system_admin';

UPDATE rbac_policies
SET permissions = permissions
        || jsonb_build_object(
            'dev_requests',
            jsonb_build_object('view', TRUE, 'create', FALSE, 'edit', TRUE, 'delete', FALSE)
        ),
    updated_at = NOW()
WHERE role_id = 'pmo_manager';
