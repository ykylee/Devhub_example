INSERT INTO public.rbac_policies (role_id, name, description, is_system, permissions)
VALUES
    (
        'system_admin',
        'System Admin',
        '시스템 설정, 조직/사용자 관리, 운영 command 관리 권한',
        TRUE,
        '{
            "audit": {"edit": false, "view": true, "create": false, "delete": false},
            "projects": {"edit": true, "view": true, "create": true, "delete": true},
            "security": {"edit": true, "view": true, "create": true, "delete": true},
            "pipelines": {"edit": true, "view": true, "create": true, "delete": true},
            "platforms": {"edit": true, "view": true, "create": true, "delete": true},
            "dev_requests": {"edit": true, "view": true, "create": true, "delete": true},
            "organization": {"edit": true, "view": true, "create": true, "delete": true},
            "scm_providers": {"edit": true, "view": true, "create": true, "delete": true},
            "infrastructure": {"edit": true, "view": true, "create": true, "delete": true},
            "platform_repositories": {"edit": true, "view": true, "create": true, "delete": true},
            "dev_request_intake_tokens": {"edit": true, "view": true, "create": true, "delete": true}
        }'::jsonb
    ),
    (
        'developer',
        'Developer',
        '개발자 대시보드, 본인 관련 repository/CI/risk 조회 권한',
        TRUE,
        '{
            "audit": {"edit": false, "view": false, "create": false, "delete": false},
            "projects": {"edit": false, "view": false, "create": false, "delete": false},
            "security": {"edit": false, "view": true, "create": false, "delete": false},
            "pipelines": {"edit": false, "view": true, "create": false, "delete": false},
            "platforms": {"edit": false, "view": false, "create": false, "delete": false},
            "dev_requests": {"edit": false, "view": true, "create": false, "delete": false},
            "organization": {"edit": false, "view": true, "create": false, "delete": false},
            "scm_providers": {"edit": false, "view": false, "create": false, "delete": false},
            "infrastructure": {"edit": false, "view": true, "create": false, "delete": false},
            "platform_repositories": {"edit": false, "view": false, "create": false, "delete": false},
            "dev_request_intake_tokens": {"edit": false, "view": false, "create": false, "delete": false}
        }'::jsonb
    ),
    (
        'team_manager',
        'Manager',
        'Application 수정 + Project 운영/멤버 관리 위양. 시스템/계정/RBAC 변경 금지.',
        TRUE,
        '{
            "audit": {"edit": false, "view": true, "create": false, "delete": false},
            "projects": {"edit": true, "view": true, "create": true, "delete": true},
            "security": {"edit": false, "view": true, "create": true, "delete": false},
            "pipelines": {"edit": false, "view": true, "create": false, "delete": false},
            "platforms": {"edit": true, "view": true, "create": false, "delete": false},
            "dev_requests": {"edit": true, "view": true, "create": false, "delete": false},
            "organization": {"edit": true, "view": true, "create": false, "delete": false},
            "scm_providers": {"edit": false, "view": true, "create": false, "delete": false},
            "infrastructure": {"edit": false, "view": true, "create": false, "delete": false},
            "platform_repositories": {"edit": false, "view": true, "create": false, "delete": false},
            "dev_request_intake_tokens": {"edit": false, "view": false, "create": false, "delete": false}
        }'::jsonb
    );
