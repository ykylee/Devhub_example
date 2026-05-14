-- 000018 down: 4개 신규 resource (applications/application_repositories/projects/scm_providers) 제거.

UPDATE rbac_policies
SET permissions = permissions
        - 'applications'
        - 'application_repositories'
        - 'projects'
        - 'scm_providers',
    updated_at = NOW()
WHERE role_id IN ('developer', 'manager', 'system_admin');
