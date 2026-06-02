-- 000021 down: team_manager row 삭제가 아니라 000005 seed 수준으로 metadata/policy 복구.
--
-- Fresh install 기준 team_manager 자체는 000005 에서 이미 생성되므로 rollback 이
-- role row 를 제거하면 FK/seed 불일치가 발생한다. 따라서 본 down 은 본 migration
-- 이전의 baseline permissions 로 되돌린다.

UPDATE rbac_policies
SET
    name = 'Manager',
    description = '팀 운영, risk triage, 승인 전 command 생성 권한',
    is_system = TRUE,
    permissions = '{
        "infrastructure": {"view": true, "create": false, "edit": false, "delete": false},
        "pipelines":      {"view": true, "create": false, "edit": false, "delete": false},
        "organization":   {"view": true, "create": false, "edit": false, "delete": false},
        "security":       {"view": true, "create": true,  "edit": false, "delete": false},
        "audit":          {"view": true, "create": false, "edit": false, "delete": false}
    }'::jsonb,
    updated_at = NOW()
WHERE role_id = 'team_manager';
