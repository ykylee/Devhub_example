-- 000026_rbac_dev_request_intake_tokens.up.sql
-- DREQ intake token admin (sprint claude/work_260515-o, ADR-0014).
-- 신규 RBAC 자원 `dev_request_intake_tokens` 를 rbac_policies.permissions JSONB 매트릭스에 추가.
-- 정책: system_admin 일임 — 모든 4축 true. developer / manager / pmo_manager 는 모두 false.
--
-- domain.DefaultPermissionMatrix() 와 byte-for-byte 정합.

UPDATE rbac_policies
SET permissions = permissions
        || jsonb_build_object(
            'dev_request_intake_tokens',
            jsonb_build_object('view', FALSE, 'create', FALSE, 'edit', FALSE, 'delete', FALSE)
        ),
    updated_at = NOW()
WHERE role_id IN ('developer', 'manager', 'pmo_manager');

UPDATE rbac_policies
SET permissions = permissions
        || jsonb_build_object(
            'dev_request_intake_tokens',
            jsonb_build_object('view', TRUE, 'create', TRUE, 'edit', TRUE, 'delete', TRUE)
        ),
    updated_at = NOW()
WHERE role_id = 'system_admin';
