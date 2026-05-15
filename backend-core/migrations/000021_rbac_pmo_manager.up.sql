-- 000021: ADR-0011 §4.2 / REQ-FR-PROJ-010 — pmo_manager system role 도입
-- (sprint claude/work_260515-d, codex PR #118 P1 review 후속).
--
-- Application/Project 운영 위양 role. handler 단의 enforceRowOwnership helper
-- 와 함께 작동: pmo_manager 는 row owner 가 아니어도 update 가능, owner-self 는
-- 그대로 통과 (단 owner-self 활성화는 route-level RBAC gate 의 정책 변경 carve out).
--
-- 매트릭스 (REQ-FR-PROJ-010 정책 매핑):
--   - applications:            view+edit (수정만, create/delete 는 system_admin)
--   - application_repositories: view only (link/unlink 초기 비허용)
--   - projects:                view+create+edit+delete (project.manage + members)
--   - scm_providers:           view only
--   - infrastructure/pipelines/organization/security/audit: view only
--   - audit invariant: create/edit/delete 모두 false (rbac_policies_audit_invariant CHECK)

ALTER TABLE rbac_policies
    DROP CONSTRAINT IF EXISTS rbac_policies_role_id_format;

ALTER TABLE rbac_policies
    ADD CONSTRAINT rbac_policies_role_id_format CHECK (
        role_id IN ('developer', 'manager', 'system_admin', 'pmo_manager')
        OR role_id ~ '^custom-[a-z0-9][a-z0-9_-]{0,62}$'
    );

INSERT INTO rbac_policies (role_id, name, description, is_system, permissions) VALUES
    (
        'pmo_manager',
        'PMO Manager',
        'Application 수정 + Project 운영/멤버 관리 위양. 시스템/계정/RBAC 변경 금지.',
        TRUE,
        '{
            "infrastructure":           {"view": true,  "create": false, "edit": false, "delete": false},
            "pipelines":                {"view": true,  "create": false, "edit": false, "delete": false},
            "organization":             {"view": true,  "create": false, "edit": false, "delete": false},
            "security":                 {"view": true,  "create": false, "edit": false, "delete": false},
            "audit":                    {"view": true,  "create": false, "edit": false, "delete": false},
            "applications":             {"view": true,  "create": false, "edit": true,  "delete": false},
            "application_repositories": {"view": true,  "create": false, "edit": false, "delete": false},
            "projects":                 {"view": true,  "create": true,  "edit": true,  "delete": true},
            "scm_providers":            {"view": true,  "create": false, "edit": false, "delete": false}
        }'::jsonb
    );
