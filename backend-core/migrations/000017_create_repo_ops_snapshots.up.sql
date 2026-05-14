-- 000017: Repository 운영 지표 스냅샷 (REQ-FR-APP-005..008, ERD §2.5).
--
-- pr_activities     : PR/PR Activity 타임라인 이벤트 (event_type 단위 row, idempotent via composite UK)
-- build_runs        : 빌드 실행 이력 (run_external_id UNIQUE — provider 별 run ID)
-- quality_snapshots : 정적분석/품질 점수/게이트 결과

CREATE TABLE pr_activities (
    id                 BIGSERIAL PRIMARY KEY,
    repository_id      BIGINT NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    external_pr_id     TEXT NOT NULL,
    event_type         TEXT NOT NULL,
    actor_login        TEXT,
    occurred_at        TIMESTAMPTZ NOT NULL,
    payload            JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT pr_activities_event_type_check CHECK (
        event_type IN ('opened', 'reviewed', 'commented', 'closed', 'merged', 'reopened', 'updated')
    ),
    CONSTRAINT pr_activities_event_unique UNIQUE (repository_id, external_pr_id, event_type, occurred_at)
);

CREATE INDEX pr_activities_repository_occurred_at_idx ON pr_activities (repository_id, occurred_at DESC);
CREATE INDEX pr_activities_external_pr_idx ON pr_activities (repository_id, external_pr_id);

CREATE TABLE build_runs (
    id                 BIGSERIAL PRIMARY KEY,
    repository_id      BIGINT NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    run_external_id    TEXT NOT NULL UNIQUE,
    branch             TEXT NOT NULL,
    commit_sha         TEXT NOT NULL,
    status             TEXT NOT NULL,
    duration_seconds   INTEGER,
    started_at         TIMESTAMPTZ NOT NULL,
    finished_at        TIMESTAMPTZ,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT build_runs_status_check CHECK (
        status IN ('queued', 'running', 'success', 'failed', 'cancelled', 'skipped', 'unknown')
    ),
    CONSTRAINT build_runs_finished_consistency CHECK (
        finished_at IS NULL OR finished_at >= started_at
    )
);

CREATE INDEX build_runs_repository_started_at_idx ON build_runs (repository_id, started_at DESC);
CREATE INDEX build_runs_repository_status_idx ON build_runs (repository_id, status);

CREATE TABLE quality_snapshots (
    id                 BIGSERIAL PRIMARY KEY,
    repository_id      BIGINT NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    tool               TEXT NOT NULL,
    ref_name           TEXT NOT NULL,
    commit_sha         TEXT,
    score              NUMERIC(6,2),
    gate_passed        BOOLEAN,
    metric_payload     JSONB NOT NULL DEFAULT '{}'::jsonb,
    measured_at        TIMESTAMPTZ NOT NULL,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT quality_snapshots_score_range CHECK (score IS NULL OR (score >= 0 AND score <= 100))
);

CREATE INDEX quality_snapshots_repository_measured_at_idx ON quality_snapshots (repository_id, measured_at DESC);
CREATE INDEX quality_snapshots_repository_tool_idx ON quality_snapshots (repository_id, tool);
