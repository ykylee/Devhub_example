CREATE TABLE rbac_policies (
    role_id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    permissions JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT rbac_policies_role_id_format CHECK (
        role_id IN ('developer', 'manager', 'system_admin')
        OR role_id ~ '^custom-[a-z0-9][a-z0-9_-]{0,62}$'
    ),
    CONSTRAINT rbac_policies_audit_invariant CHECK (
        COALESCE((permissions -> 'audit' ->> 'create')::boolean, FALSE) = FALSE
        AND COALESCE((permissions -> 'audit' ->> 'edit')::boolean, FALSE) = FALSE
        AND COALESCE((permissions -> 'audit' ->> 'delete')::boolean, FALSE) = FALSE
    )
);

CREATE INDEX rbac_policies_is_system_idx ON rbac_policies (is_system);

-- Seed system roles. Matrices match docs/backend_api_contract.md section 12.1 default policy
-- and preserve M0 requireMinRole enforcement byte-for-byte. Descriptions match
-- internal/domain/rbac.go SystemRoles() so PR-G3 store seed can rely on either source.
INSERT INTO rbac_policies (role_id, name, description, is_system, permissions) VALUES
    (
        'developer',
        'Developer',
        '개발자 대시보드, 본인 관련 repository/CI/risk 조회 권한',
        TRUE,
        '{
            "infrastructure": {"view": true,  "create": false, "edit": false, "delete": false},
            "pipelines":      {"view": true,  "create": false, "edit": false, "delete": false},
            "organization":   {"view": true,  "create": false, "edit": false, "delete": false},
            "security":       {"view": true,  "create": false, "edit": false, "delete": false},
            "audit":          {"view": false, "create": false, "edit": false, "delete": false}
        }'::jsonb
    ),
    (
        'manager',
        'Manager',
        '팀 운영, risk triage, 승인 전 command 생성 권한',
        TRUE,
        '{
            "infrastructure": {"view": true, "create": false, "edit": false, "delete": false},
            "pipelines":      {"view": true, "create": false, "edit": false, "delete": false},
            "organization":   {"view": true, "create": false, "edit": false, "delete": false},
            "security":       {"view": true, "create": true,  "edit": false, "delete": false},
            "audit":          {"view": true, "create": false, "edit": false, "delete": false}
        }'::jsonb
    ),
    (
        'system_admin',
        'System Admin',
        '시스템 설정, 조직/사용자 관리, 운영 command 관리 권한',
        TRUE,
        '{
            "infrastructure": {"view": true, "create": true, "edit": true, "delete": true},
            "pipelines":      {"view": true, "create": true, "edit": true, "delete": true},
            "organization":   {"view": true, "create": true, "edit": true, "delete": true},
            "security":       {"view": true, "create": true, "edit": true, "delete": true},
            "audit":          {"view": true, "create": false, "edit": false, "delete": false}
        }'::jsonb
    );
