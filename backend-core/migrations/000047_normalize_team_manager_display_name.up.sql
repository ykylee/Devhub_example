-- 000047: normalize team_manager system role display label after legacy PMO rename drift.
--
-- Existing databases may still carry 000021's "PMO Manager" label even though the
-- canonical frontend/admin UX uses "Manager" for the team_manager role id.
-- Keep permissions untouched here except for re-applying the full current matrix so
-- stale installs converge on the 11-resource model.

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
        "scm_providers":            {"view": true,  "create": false, "edit": false, "delete": false},
        "dev_requests":             {"view": true,  "create": false, "edit": true,  "delete": false},
        "dev_request_intake_tokens":{"view": false, "create": false, "edit": false, "delete": false}
    }'::jsonb,
    updated_at = NOW()
WHERE role_id = 'team_manager';
