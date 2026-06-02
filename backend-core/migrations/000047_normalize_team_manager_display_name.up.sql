-- 000047: normalize team_manager system role display label + permission matrix sync.
--
-- Purpose:
--   1. Stale installs that still carry 000021's old display name/permissions for
--      team_manager converge on the current DefaultPermissionMatrix (11-resource).
--   2. The permission matrix was also out of sync: organization.edit and
--      security.create were false in 000021 but are true in DefaultPermissionMatrix
--      (rbac.go:176-192). Fix both to match the canonical source of truth.
--
-- Safety: rbac_policies PK CHECK constraint only allows 'developer', 'team_manager',
-- or 'system_admin', so legacy role_ids never reach this table.  INSERT ON CONFLICT
-- guards fresh-install scenarios where 000005 seed may already have created the row.

INSERT INTO rbac_policies (role_id, name, description, is_system, permissions, created_at, updated_at)
VALUES (
    'team_manager',
    'Manager',
    'Application 수정 + Project 운영/멤버 관리 위양. 시스템/계정/RBAC 변경 금지.',
    TRUE,
    '{
        "infrastructure":           {"view": true,  "create": false, "edit": false, "delete": false},
        "pipelines":                {"view": true,  "create": false, "edit": false, "delete": false},
        "organization":             {"view": true,  "create": false, "edit": true,  "delete": false},
        "security":                 {"view": true,  "create": true,  "edit": false, "delete": false},
        "audit":                    {"view": true,  "create": false, "edit": false, "delete": false},
        "applications":             {"view": true,  "create": false, "edit": true,  "delete": false},
        "application_repositories": {"view": true,  "create": false, "edit": false, "delete": false},
        "projects":                 {"view": true,  "create": true,  "edit": true,  "delete": true},
        "scm_providers":            {"view": true,  "create": false, "edit": false, "delete": false},
        "dev_requests":             {"view": true,  "create": false, "edit": true,  "delete": false},
        "dev_request_intake_tokens":{"view": false, "create": false, "edit": false, "delete": false}
    }'::jsonb,
    NOW(),
    NOW()
)
ON CONFLICT (role_id) DO UPDATE SET
    name           = EXCLUDED.name,
    description    = EXCLUDED.description,
    is_system      = EXCLUDED.is_system,
    permissions    = EXCLUDED.permissions,
    updated_at     = NOW();
