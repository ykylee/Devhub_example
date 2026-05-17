CREATE TABLE IF NOT EXISTS infra_service_snapshots (
    ingest_id text PRIMARY KEY,
    agent_id text NOT NULL,
    snapshot_at timestamptz NOT NULL,
    trace_id text NULL,
    nodes_payload jsonb NOT NULL,
    services_payload jsonb NOT NULL,
    degraded_providers jsonb NOT NULL DEFAULT '[]'::jsonb,
    created_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_infra_service_snapshots_snapshot_at
    ON infra_service_snapshots (snapshot_at DESC, created_at DESC);
