-- 000018: ADR-0011 1차 — RBAC matrix 확장 (applications/application_repositories/projects/scm_providers).
--
-- 4개 신규 resource × 4 axis (view/create/edit/delete) = 16 cell × 3 system role = 48 cell.
-- 1차 정책: system_admin 만 모든 axis true, developer/manager 는 모든 axis false (REQ-FR-PROJ-000).
-- 옵션 C (handler/service 코드 검증) 채택이라 row_predicate 컬럼은 추가하지 않음 — 옵션 B
-- 채택 시 별도 마이그레이션에서 컬럼 추가 + 정책 업그레이드 (ADR-0011 §4.3).

UPDATE rbac_policies
SET permissions = permissions
        || jsonb_build_object(
            'applications',           jsonb_build_object('view', FALSE, 'create', FALSE, 'edit', FALSE, 'delete', FALSE),
            'application_repositories', jsonb_build_object('view', FALSE, 'create', FALSE, 'edit', FALSE, 'delete', FALSE),
            'projects',               jsonb_build_object('view', FALSE, 'create', FALSE, 'edit', FALSE, 'delete', FALSE),
            'scm_providers',          jsonb_build_object('view', FALSE, 'create', FALSE, 'edit', FALSE, 'delete', FALSE)
        ),
    updated_at = NOW()
WHERE role_id IN ('developer', 'manager');

UPDATE rbac_policies
SET permissions = permissions
        || jsonb_build_object(
            'applications',           jsonb_build_object('view', TRUE, 'create', TRUE, 'edit', TRUE, 'delete', TRUE),
            'application_repositories', jsonb_build_object('view', TRUE, 'create', TRUE, 'edit', TRUE, 'delete', TRUE),
            'projects',               jsonb_build_object('view', TRUE, 'create', TRUE, 'edit', TRUE, 'delete', TRUE),
            'scm_providers',          jsonb_build_object('view', TRUE, 'create', TRUE, 'edit', TRUE, 'delete', TRUE)
        ),
    updated_at = NOW()
WHERE role_id = 'system_admin';
