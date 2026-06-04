-- 000021: ADR-0011 §4.2 / REQ-FR-PROJ-010 — team_manager system role 정책 확장
-- (sprint claude/work_260515-d, codex PR #118 P1 review 후속).
--
-- Fresh install 기준 team_manager system role 은 000005 seed 에서 이미 생성된다.
-- 따라서 본 migration 의 책임은 "생성"이 아니라 metadata + permissions 업그레이드다.
-- 기존 INSERT 구현은 신규 DB bootstrap 에서 PK 충돌을 일으켰다.
--
-- 매트릭스 (REQ-FR-PROJ-010 정책 매핑):
--   - applications:            view+edit (수정만, create/delete 는 system_admin)
--   - application_repositories: view only (link/unlink 초기 비허용)
--   - projects:                view+create+edit+delete (project.manage + members)
--   - scm_providers:           view only
--   - infrastructure/pipelines/organization/security/audit: view only
--   - audit invariant: create/edit/delete 모두 false (rbac_policies_audit_invariant CHECK)

UPDATE rbac_policies
SET
    name = 'Manager',
    description = 'Application 수정 + Project 운영/멤버 관리 위양. 시스템/계정/RBAC 변경 금지.',
    is_system = TRUE,
    permissions = '{
        "infrastructure":           {"view": true,  "create": false, "edit": false, "delete": false},
        "pipelines":                {"view": true,  "create": false, "edit": false, "delete": false},
        "organization":             {"view": true,  "create": false, "edit": false, "delete": false},
        "security":                 {"view": true,  "create": false, "edit": false, "delete": false},
        "audit":                    {"view": true,  "create": false, "edit": false, "delete": false},
        "applications":             {"view": true,  "create": false, "edit": true,  "delete": false},
        "application_repositories": {"view": true,  "create": false, "edit": false, "delete": false},
        "projects":                 {"view": true,  "create": true,  "edit": true,  "delete": true},
        "scm_providers":            {"view": true,  "create": false, "edit": false, "delete": false}
    }'::jsonb,
    updated_at = NOW()
WHERE role_id = 'team_manager';
