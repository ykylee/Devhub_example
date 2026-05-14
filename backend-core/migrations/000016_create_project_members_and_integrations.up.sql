-- 000016: project_members + project_integrations (REQ-FR-PROJ-003 / -005 / -006, ERD §2.5).
--
-- project_members      : composite PK = (project_id, user_id) — 동일 사용자 중복 멤버십 차단.
-- project_integrations : 단일 id PK + scope 컬럼으로 application/project 양쪽 등록 지원.
--                        external integration (jira/confluence) policy 보유. concept §13.7.

CREATE TABLE project_members (
    project_id     UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id        TEXT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    project_role   TEXT NOT NULL,
    joined_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (project_id, user_id),
    CONSTRAINT project_members_role_check CHECK (project_role IN ('lead', 'contributor', 'observer'))
);

CREATE INDEX project_members_user_idx ON project_members (user_id);

CREATE TABLE project_integrations (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scope              TEXT NOT NULL,
    project_id         UUID REFERENCES projects(id) ON DELETE CASCADE,
    application_id     UUID REFERENCES applications(id) ON DELETE CASCADE,
    integration_type   TEXT NOT NULL,
    external_key       TEXT NOT NULL,
    url                TEXT NOT NULL,
    policy             TEXT NOT NULL,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT project_integrations_scope_check CHECK (scope IN ('application', 'project')),
    CONSTRAINT project_integrations_type_check CHECK (integration_type IN ('jira', 'confluence')),
    CONSTRAINT project_integrations_policy_check CHECK (policy IN ('summary_only', 'execution_system')),
    CONSTRAINT project_integrations_scope_target_consistency CHECK (
        (scope = 'application' AND application_id IS NOT NULL AND project_id IS NULL)
        OR (scope = 'project'    AND project_id    IS NOT NULL AND application_id IS NULL)
    )
);

-- 동일 scope 내 (target, integration_type, external_key) 중복 방지
CREATE UNIQUE INDEX project_integrations_application_unique
    ON project_integrations (application_id, integration_type, external_key)
    WHERE application_id IS NOT NULL;
CREATE UNIQUE INDEX project_integrations_project_unique
    ON project_integrations (project_id, integration_type, external_key)
    WHERE project_id IS NOT NULL;
