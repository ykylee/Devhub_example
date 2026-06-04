-- 000024 down: rbac_policies.permissions 에서 dev_requests 키 제거.

UPDATE rbac_policies
SET permissions = permissions - 'dev_requests',
    updated_at = NOW()
WHERE permissions ? 'dev_requests';
