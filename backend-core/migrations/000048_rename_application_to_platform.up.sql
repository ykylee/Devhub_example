-- 000048: Rename "application" to "platform" across all schemas.
--
-- Renames tables, columns, constraints, indexes, CHECK constraints, and
-- RBAC JSONB resource keys from "application" nomenclature to "platform".
-- Also updates project_integrations.scope and integration_bindings.scope_type
-- values from 'application' → 'platform'.
--
-- Down: 000048_rename_application_to_platform.down.sql

-- ============================================
-- 1. Drop FK constraints referencing applications.id
-- ============================================

ALTER TABLE projects
    DROP CONSTRAINT IF EXISTS projects_application_id_fkey;

ALTER TABLE project_integrations
    DROP CONSTRAINT IF EXISTS project_integrations_application_id_fkey;

ALTER TABLE application_repositories
    DROP CONSTRAINT IF EXISTS application_repositories_application_id_fkey;

-- ============================================
-- 2. Drop indexes referencing old column/table names
-- ============================================

DROP INDEX IF EXISTS projects_application_idx;
DROP INDEX IF EXISTS project_integrations_application_unique;
DROP INDEX IF EXISTS application_repositories_provider_repo_idx;
DROP INDEX IF EXISTS application_repositories_sync_status_idx;
DROP INDEX IF EXISTS application_repositories_external_repo_id_idx;

-- ============================================
-- 3. Rename tables
-- ============================================

ALTER TABLE applications RENAME TO platforms;
ALTER TABLE application_repositories RENAME TO platform_repositories;

-- ============================================
-- 4. Rename columns (application_id → platform_id)
-- ============================================

ALTER TABLE projects RENAME COLUMN application_id TO platform_id;
ALTER TABLE project_integrations RENAME COLUMN application_id TO platform_id;
ALTER TABLE platform_repositories RENAME COLUMN application_id TO platform_id;

-- ============================================
-- 5. Rename constraints on renamed tables
-- ============================================

ALTER TABLE platforms RENAME CONSTRAINT applications_pkey TO platforms_pkey;
ALTER TABLE platforms RENAME CONSTRAINT applications_key_format TO platforms_key_format;
ALTER TABLE platforms RENAME CONSTRAINT applications_status_check TO platforms_status_check;
ALTER TABLE platforms RENAME CONSTRAINT applications_visibility_check TO platforms_visibility_check;
ALTER TABLE platforms RENAME CONSTRAINT applications_archived_consistency TO platforms_archived_consistency;
ALTER TABLE platforms RENAME CONSTRAINT applications_due_date_after_start TO platforms_due_date_after_start;

ALTER TABLE platform_repositories RENAME CONSTRAINT application_repositories_pkey TO platform_repositories_pkey;
ALTER TABLE platform_repositories RENAME CONSTRAINT application_repositories_role_check TO platform_repositories_role_check;
ALTER TABLE platform_repositories RENAME CONSTRAINT application_repositories_sync_status_check TO platform_repositories_sync_status_check;
ALTER TABLE platform_repositories RENAME CONSTRAINT application_repositories_sync_error_code_check TO platform_repositories_sync_error_code_check;
ALTER TABLE platform_repositories RENAME CONSTRAINT application_repositories_sync_error_consistency TO platform_repositories_sync_error_consistency;

-- ============================================
-- 6. Rename indexes on renamed tables
-- ============================================

ALTER INDEX IF EXISTS applications_status_idx RENAME TO platforms_status_idx;
ALTER INDEX IF EXISTS applications_visibility_idx RENAME TO platforms_visibility_idx;
ALTER INDEX IF EXISTS applications_owner_idx RENAME TO platforms_owner_idx;
ALTER INDEX IF EXISTS applications_archived_at_idx RENAME TO platforms_archived_at_idx;
ALTER INDEX IF EXISTS applications_leader_idx RENAME TO platforms_leader_idx;
ALTER INDEX IF EXISTS applications_dev_unit_idx RENAME TO platforms_dev_unit_idx;

-- ============================================
-- 7. Recreate FK constraints with new names
-- ============================================

ALTER TABLE projects
    ADD CONSTRAINT projects_platform_id_fkey
    FOREIGN KEY (platform_id) REFERENCES platforms(id) ON DELETE SET NULL;

ALTER TABLE project_integrations
    ADD CONSTRAINT project_integrations_platform_id_fkey
    FOREIGN KEY (platform_id) REFERENCES platforms(id) ON DELETE CASCADE;

ALTER TABLE platform_repositories
    ADD CONSTRAINT platform_repositories_platform_id_fkey
    FOREIGN KEY (platform_id) REFERENCES platforms(id) ON DELETE CASCADE;

-- ============================================
-- 8. Recreate indexes with new names
-- ============================================

CREATE INDEX projects_platform_idx ON projects (platform_id);
CREATE UNIQUE INDEX project_integrations_platform_unique
    ON project_integrations (platform_id, integration_type, external_key)
    WHERE platform_id IS NOT NULL;
CREATE INDEX platform_repositories_provider_repo_idx ON platform_repositories (repo_provider, repo_full_name);
CREATE INDEX platform_repositories_sync_status_idx ON platform_repositories (sync_status);
CREATE INDEX platform_repositories_external_repo_id_idx
    ON platform_repositories (repo_provider, external_repo_id)
    WHERE external_repo_id IS NOT NULL;

-- ============================================
-- 9. Update scope column values + CHECK constraints
-- ============================================

-- project_integrations.scope: 'application' → 'platform'
UPDATE project_integrations SET scope = 'platform' WHERE scope = 'application';
ALTER TABLE project_integrations DROP CONSTRAINT IF EXISTS project_integrations_scope_check;
ALTER TABLE project_integrations ADD CONSTRAINT project_integrations_scope_check
    CHECK (scope IN ('platform', 'project'));

ALTER TABLE project_integrations DROP CONSTRAINT IF EXISTS project_integrations_scope_target_consistency;
ALTER TABLE project_integrations ADD CONSTRAINT project_integrations_scope_target_consistency
    CHECK (
        (scope = 'platform' AND platform_id IS NOT NULL AND project_id IS NULL)
        OR (scope = 'project'   AND project_id    IS NOT NULL AND platform_id IS NULL)
    );

-- dev_requests.registered_target_type: 'application' → 'platform'
UPDATE dev_requests SET registered_target_type = 'platform' WHERE registered_target_type = 'application';
ALTER TABLE dev_requests DROP CONSTRAINT IF EXISTS dev_requests_target_type_check;
ALTER TABLE dev_requests ADD CONSTRAINT dev_requests_target_type_check
    CHECK (registered_target_type IS NULL
           OR registered_target_type IN ('platform', 'project'));

-- integration_bindings.scope_type: 'application' → 'platform'
UPDATE integration_bindings SET scope_type = 'platform' WHERE scope_type = 'application';
ALTER TABLE integration_bindings DROP CONSTRAINT IF EXISTS integration_bindings_scope_type_check;
ALTER TABLE integration_bindings ADD CONSTRAINT integration_bindings_scope_type_check
    CHECK (scope_type IN ('platform', 'project'));

-- ============================================
-- 10. Update RBAC JSONB resource keys
--     'applications' → 'platforms'
--     'application_repositories' → 'platform_repositories'
-- ============================================

UPDATE rbac_policies
SET permissions = (
    SELECT jsonb_object_agg(
        CASE key
            WHEN 'applications' THEN 'platforms'
            WHEN 'application_repositories' THEN 'platform_repositories'
            ELSE key
        END,
        value
    )
    FROM jsonb_each(permissions)
)
WHERE permissions ? 'applications' OR permissions ? 'application_repositories';
