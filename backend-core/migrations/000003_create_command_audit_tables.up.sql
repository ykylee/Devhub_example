CREATE TABLE commands (
    id BIGSERIAL PRIMARY KEY,
    command_id TEXT NOT NULL UNIQUE,
    command_type TEXT NOT NULL,
    target_type TEXT NOT NULL,
    target_id TEXT NOT NULL,
    action_type TEXT NOT NULL,
    status TEXT NOT NULL,
    actor_login TEXT NOT NULL,
    reason TEXT NOT NULL,
    dry_run BOOLEAN NOT NULL DEFAULT TRUE,
    requires_approval BOOLEAN NOT NULL DEFAULT FALSE,
    idempotency_key TEXT UNIQUE,
    request_payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    result_payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT commands_type_check CHECK (command_type IN ('risk_mitigation', 'service_action', 'weekly_report')),
    CONSTRAINT commands_status_check CHECK (status IN ('pending', 'running', 'succeeded', 'failed', 'rejected', 'cancelled')),
    CONSTRAINT commands_target_check CHECK (target_type IN ('risk', 'service', 'report'))
);

CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    audit_id TEXT NOT NULL UNIQUE,
    actor_login TEXT NOT NULL,
    action TEXT NOT NULL,
    target_type TEXT NOT NULL,
    target_id TEXT NOT NULL,
    command_id TEXT REFERENCES commands(command_id) ON DELETE SET NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX commands_status_updated_at_idx ON commands (status, updated_at DESC);
CREATE INDEX commands_target_updated_at_idx ON commands (target_type, target_id, updated_at DESC);
CREATE INDEX commands_idempotency_key_idx ON commands (idempotency_key) WHERE idempotency_key IS NOT NULL;
CREATE INDEX audit_logs_command_id_idx ON audit_logs (command_id);
CREATE INDEX audit_logs_target_created_at_idx ON audit_logs (target_type, target_id, created_at DESC);
