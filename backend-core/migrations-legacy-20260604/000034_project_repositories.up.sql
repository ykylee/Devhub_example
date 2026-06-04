-- 000034: project_repositories
--
-- Project 와 Repository 의 N:M 연결을 지원하는 조인 테이블.
-- 기존 projects.repository_id 는 legacy 호환(primary repository) 용도로 유지한다.

CREATE TABLE project_repositories (
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    repository_id   BIGINT NOT NULL REFERENCES repositories(id) ON DELETE RESTRICT,
    role            TEXT NOT NULL DEFAULT 'linked',
    linked_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (project_id, repository_id),
    CONSTRAINT project_repositories_role_check CHECK (role IN ('primary', 'linked', 'shared'))
);

CREATE INDEX project_repositories_repository_idx ON project_repositories (repository_id);

-- 기존 데이터 백필: legacy projects.repository_id 를 primary link 로 승격
INSERT INTO project_repositories (project_id, repository_id, role)
SELECT id, repository_id, 'primary'
FROM projects
ON CONFLICT (project_id, repository_id) DO NOTHING;

