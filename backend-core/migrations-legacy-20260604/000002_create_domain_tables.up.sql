CREATE TABLE gitea_users (
    id BIGSERIAL PRIMARY KEY,
    gitea_user_id BIGINT UNIQUE,
    login TEXT NOT NULL UNIQUE,
    display_name TEXT,
    avatar_url TEXT,
    html_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE repositories (
    id BIGSERIAL PRIMARY KEY,
    gitea_repository_id BIGINT UNIQUE,
    full_name TEXT NOT NULL UNIQUE,
    owner_login TEXT,
    name TEXT NOT NULL,
    clone_url TEXT,
    html_url TEXT,
    default_branch TEXT,
    private BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE issues (
    id BIGSERIAL PRIMARY KEY,
    gitea_issue_id BIGINT UNIQUE,
    repository_id BIGINT NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    number BIGINT NOT NULL,
    title TEXT NOT NULL,
    state TEXT NOT NULL,
    author_login TEXT,
    assignee_login TEXT,
    html_url TEXT,
    opened_at TIMESTAMPTZ,
    closed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT issues_state_check CHECK (state IN ('open', 'closed')),
    CONSTRAINT issues_repository_number_unique UNIQUE (repository_id, number)
);

CREATE TABLE pull_requests (
    id BIGSERIAL PRIMARY KEY,
    gitea_pull_request_id BIGINT UNIQUE,
    repository_id BIGINT NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    number BIGINT NOT NULL,
    title TEXT NOT NULL,
    state TEXT NOT NULL,
    author_login TEXT,
    head_branch TEXT,
    base_branch TEXT,
    head_sha TEXT,
    html_url TEXT,
    merged_at TIMESTAMPTZ,
    closed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT pull_requests_state_check CHECK (state IN ('open', 'closed', 'merged')),
    CONSTRAINT pull_requests_repository_number_unique UNIQUE (repository_id, number)
);

CREATE TABLE ci_runs (
    id BIGSERIAL PRIMARY KEY,
    external_id TEXT NOT NULL UNIQUE,
    repository_id BIGINT REFERENCES repositories(id) ON DELETE SET NULL,
    repository_name TEXT NOT NULL,
    branch TEXT,
    commit_sha TEXT,
    status TEXT NOT NULL,
    conclusion TEXT,
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    duration_seconds INTEGER,
    html_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ci_runs_status_check CHECK (status IN ('queued', 'running', 'success', 'failed', 'cancelled', 'skipped', 'unknown'))
);

CREATE TABLE risks (
    id BIGSERIAL PRIMARY KEY,
    risk_key TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    reason TEXT NOT NULL,
    impact TEXT NOT NULL,
    status TEXT NOT NULL,
    owner_login TEXT,
    source_type TEXT NOT NULL,
    source_id TEXT,
    suggested_actions JSONB NOT NULL DEFAULT '[]'::jsonb,
    detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    mitigated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT risks_impact_check CHECK (impact IN ('low', 'medium', 'high', 'critical')),
    CONSTRAINT risks_status_check CHECK (status IN ('detected', 'investigation', 'action_required', 'mitigated', 'dismissed'))
);

CREATE INDEX repositories_owner_login_idx ON repositories (owner_login);
CREATE INDEX issues_repository_state_updated_at_idx ON issues (repository_id, state, updated_at DESC);
CREATE INDEX pull_requests_repository_state_updated_at_idx ON pull_requests (repository_id, state, updated_at DESC);
CREATE INDEX ci_runs_repository_status_updated_at_idx ON ci_runs (repository_name, status, updated_at DESC);
CREATE INDEX risks_status_impact_updated_at_idx ON risks (status, impact, updated_at DESC);
