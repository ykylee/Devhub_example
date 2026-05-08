CREATE TABLE rbac_policy_versions (
    id BIGSERIAL PRIMARY KEY,
    policy_version TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL DEFAULT 'active',
    actor_login TEXT NOT NULL,
    reason TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT rbac_policy_versions_status_check CHECK (status IN ('active', 'archived'))
);

CREATE UNIQUE INDEX rbac_policy_versions_single_active_idx
    ON rbac_policy_versions (status)
    WHERE status = 'active';

CREATE TABLE rbac_policy_rules (
    id BIGSERIAL PRIMARY KEY,
    policy_version_id BIGINT NOT NULL REFERENCES rbac_policy_versions(id) ON DELETE CASCADE,
    role TEXT NOT NULL,
    resource TEXT NOT NULL,
    permission TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT rbac_policy_rules_role_check CHECK (role IN ('developer', 'manager', 'system_admin')),
    CONSTRAINT rbac_policy_rules_permission_check CHECK (permission IN ('none', 'read', 'write', 'admin')),
    CONSTRAINT rbac_policy_rules_unique UNIQUE (policy_version_id, role, resource)
);

CREATE INDEX rbac_policy_rules_role_resource_idx ON rbac_policy_rules (role, resource);

INSERT INTO rbac_policy_versions (policy_version, status, actor_login, reason)
VALUES ('2026-05-07.default', 'active', 'system', 'Seed default RBAC policy');

INSERT INTO rbac_policy_rules (policy_version_id, role, resource, permission)
SELECT v.id, r.role, r.resource, r.permission
FROM rbac_policy_versions v
CROSS JOIN (VALUES
    ('developer', 'repositories', 'read'),
    ('developer', 'ci_runs', 'read'),
    ('developer', 'risks', 'read'),
    ('developer', 'commands', 'none'),
    ('developer', 'organization', 'none'),
    ('developer', 'system_config', 'none'),
    ('manager', 'repositories', 'write'),
    ('manager', 'ci_runs', 'read'),
    ('manager', 'risks', 'write'),
    ('manager', 'commands', 'write'),
    ('manager', 'organization', 'read'),
    ('manager', 'system_config', 'none'),
    ('system_admin', 'repositories', 'admin'),
    ('system_admin', 'ci_runs', 'admin'),
    ('system_admin', 'risks', 'admin'),
    ('system_admin', 'commands', 'admin'),
    ('system_admin', 'organization', 'admin'),
    ('system_admin', 'system_config', 'admin')
) AS r(role, resource, permission)
WHERE v.policy_version = '2026-05-07.default';
