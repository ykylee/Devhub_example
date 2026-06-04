-- 000048: DOWN — Revert "application" → "platform" rename.
--
-- Reverse all changes: rename tables/columns back, restore old FK names,
-- revert RBAC JSONB keys, and revert scope values.

-- ============================================
-- 1. Drop FK constraints referencing platforms.id
-- ============================================

ALTER TABLE projects
    DROP CONSTRAINT IF EXISTS projects_platform_id_fkey;

ALTER TABLE project_integrations
    DROP CONSTRAINT IF EXISTS project_integrations_platform_id_fkey;

ALTER TABLE platform_repositories
    DROP CONSTRAINT IF EXISTS platform_repositories_platform_id_fkey;

-- ============================================
-- 2. Drop indexes with new names
-- ============================================

DROP INDEX IF EXISTS projects_platform_idx;
DROP INDEX IF EXISTS project_integrations_platform_unique;
DROP INDEX IF EXISTS platform_repositories_provider_repo_idx;
DROP INDEX IF EXISTS platform_repositories_sync_status_idx;
DROP INDEX IF EXISTS platform_repositories_external_repo_id_idx;

-- ============================================
-- 3. Rename tables back
-- ============================================

ALTER TABLE platforms RENAME TO applications;
ALTER TABLE platform_repositories RENAME TO application_repositories;

-- ============================================
-- 4. Rename columns back (platform_id → application_id)
-- ============================================

ALTER TABLE projects RENAME COLUMN platform_id TO application_id;
ALTER TABLE project_integrations RENAME COLUMN platform_id TO application_id;
ALTER TABLE application_repositories RENAME COLUMN platform_id TO application_id;

-- ============================================
-- 5. Rename constraints back
-- ============================================

ALTER TABLE applications RENAME CONSTRAINT platforms_pkey TO applications_pkey;
ALTER TABLE applications RENAME CONSTRAINT platforms_key_format TO applications_key_format;
ALTER TABLE applications RENAME CONSTRAINT platforms_status_check TO applications_status_check;
ALTER TABLE applications RENAME CONSTRAINT platforms_visibility_check TO applications_visibility_check;
ALTER TABLE applications RENAME CONSTRAINT platforms_archived_consistency TO applications_archived_consistency;
ALTER TABLE applications RENAME CONSTRAINT platforms_due_date_after_start TO applications_due_date_after_start;

ALTER TABLE application_repositories RENAME CONSTRAINT platform_repositories_pkey TO application_repositories_pkey;
ALTER TABLE application_repositories RENAME CONSTRAINT platform_repositories_role_check TO application_repositories_role_check;
ALTER TABLE application_repositories RENAME CONSTRAINT platform_repositories_sync_status_check TO application_repositories_sync_status_check;
ALTER TABLE application_repositories RENAME CONSTRAINT platform_repositories_sync_error_code_check TO application_repositories_sync_error_code_check;
ALTER TABLE application_repositories RENAME CONSTRAINT platform_repositories_sync_error_consistency TO application_repositories_sync_error_consistency;

-- ============================================
-- 6. Rename indexes back
-- ============================================

ALTER INDEX IF EXISTS platforms_status_idx RENAME TO applications_status_idx;
ALTER INDEX IF EXISTS platforms_visibility_idx RENAME TO applications_visibility_idx;
ALTER INDEX IF EXISTS platforms_owner_idx RENAME TO applications_owner_idx;
ALTER INDEX IF EXISTS platforms_archived_at_idx RENAME TO applications_archived_at_idx;
ALTER INDEX IF EXISTS platforms_leader_idx RENAME TO applications_leader_idx;
ALTER INDEX IF EXISTS platforms_dev_unit_idx RENAME TO applications_dev_unit_idx;

-- ============================================
-- 7. Recreate original FK constraints
-- ============================================

ALTER TABLE projects
    ADD CONSTRAINT projects_application_id_fkey
    FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE SET NULL;

ALTER TABLE project_integrations
    ADD CONSTRAINT project_integrations_application_id_fkey
    FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE;

ALTER TABLE application_repositories
    ADD CONSTRAINT application_repositories_application_id_fkey
    FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE;

-- ============================================
-- 8. Recreate original indexes
-- ============================================

CREATE INDEX projects_application_idx ON projects (application_id);
CREATE UNIQUE INDEX project_integrations_application_unique
    ON project_integrations (application_id, integration_type, external_key)
    WHERE application_id IS NOT NULL;
CREATE INDEX application_repositories_provider_repo_idx ON application_repositories (repo_provider, repo_full_name);
CREATE INDEX application_repositories_sync_status_idx ON application_repositories (sync_status);
CREATE INDEX application_repositories_external_repo_id_idx
    ON application_repositories (repo_provider, external_repo_id)
    WHERE external_repo_id IS NOT NULL;

-- ============================================
-- 9. Revert scope column + CHECK constraints
-- ============================================

-- project_integrations.scope: 'platform' → 'application'
UPDATE project_integrations SET scope = 'application' WHERE scope = 'platform';
ALTER TABLE project_integrations DROP CONSTRAINT IF EXISTS project_integrations_scope_check;
ALTER TABLE project_integrations ADD CONSTRAINT project_integrations_scope_check
    CHECK (scope IN ('application', 'project'));

ALTER TABLE project_integrations DROP CONSTRAINT IF EXISTS project_integrations_scope_target_consistency;
ALTER TABLE project_integrations ADD CONSTRAINT project_integrations_scope_target_consistency
    CHECK (
        (scope = 'application' AND application_id IS NOT NULL AND project_id IS NULL)
        OR (scope = 'project'    AND project_id    IS NOT NULL AND application_id IS NULL)
    );

-- dev_requests.registered_target_type: 'platform' → 'application'
UPDATE dev_requests SET registered_target_type = 'application' WHERE registered_target_type = 'platform';
ALTER TABLE dev_requests DROP CONSTRAINT IF EXISTS dev_requests_target_type_check;
ALTER TABLE dev_requests ADD CONSTRAINT dev_requests_target_type_check
    CHECK (registered_target_type IS NULL
           OR registered_target_type IN ('application', 'project'));

-- integration_bindings.scope_type: 'platform' → 'application'
UPDATE integration_bindings SET scope_type = 'application' WHERE scope_type = 'platform';
ALTER TABLE integration_bindings DROP CONSTRAINT IF EXISTS integration_bindings_scope_type_check;
ALTER TABLE integration_bindings ADD CONSTRAINT integration_bindings_scope_type_check
    CHECK (scope_type IN ('application', 'project'));

-- ============================================
-- 10. Revert RBAC JSONB resource keys
--     'platforms' → 'applications'
--     'platform_repositories' → 'application_repositories'
-- ============================================

UPDATE rbac_policies
SET permissions = (
    SELECT jsonb_object_agg(
        CASE key
            WHEN 'platforms' THEN 'applications'
            WHEN 'platform_repositories' THEN 'application_repositories'
            ELSE key
        END,
        value
    )
    FROM jsonb_each(permissions)
)
WHERE permissions ? 'platforms' OR permissions ? 'platform_repositories';
