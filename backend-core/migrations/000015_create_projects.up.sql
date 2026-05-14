-- 000015: Projects — Repository 하위 기간성 운영 단위 (REQ-FR-PROJ-001..010, ERD §2.5).
--
-- id            : 내부 UUID PK
-- application_id: 총괄 Application FK (optional — Repository 단독 Project 도 허용 가능, 후속 결정)
-- repository_id : 실행 Repository FK (existing repositories.id BIGSERIAL)
-- key           : Repository 내 unique key (UNIQUE (repository_id, key))
-- status        : 5종 상태 머신 (Application 과 동일 vocabulary)

CREATE TABLE projects (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id     UUID REFERENCES applications(id) ON DELETE SET NULL,
    repository_id      BIGINT NOT NULL REFERENCES repositories(id) ON DELETE RESTRICT,
    key                TEXT NOT NULL,
    name               TEXT NOT NULL,
    description        TEXT,
    status             TEXT NOT NULL,
    visibility         TEXT NOT NULL,
    owner_user_id      TEXT REFERENCES users(user_id) ON DELETE SET NULL,
    start_date         DATE,
    due_date           DATE,
    archived_at        TIMESTAMPTZ,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT projects_status_check CHECK (status IN ('planning', 'active', 'on_hold', 'closed', 'archived')),
    CONSTRAINT projects_visibility_check CHECK (visibility IN ('public', 'internal', 'restricted')),
    CONSTRAINT projects_archived_consistency CHECK (
        (status = 'archived' AND archived_at IS NOT NULL)
        OR (status <> 'archived' AND archived_at IS NULL)
    ),
    CONSTRAINT projects_due_date_after_start CHECK (
        start_date IS NULL OR due_date IS NULL OR due_date >= start_date
    ),
    CONSTRAINT projects_repository_key_unique UNIQUE (repository_id, key)
);

CREATE INDEX projects_application_idx ON projects (application_id);
CREATE INDEX projects_repository_idx ON projects (repository_id);
CREATE INDEX projects_status_idx ON projects (status);
CREATE INDEX projects_visibility_idx ON projects (visibility);
CREATE INDEX projects_owner_idx ON projects (owner_user_id);
