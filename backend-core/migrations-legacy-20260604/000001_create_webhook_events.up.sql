CREATE TABLE webhook_events (
    id BIGSERIAL PRIMARY KEY,
    event_type TEXT NOT NULL,
    delivery_id TEXT,
    dedupe_key TEXT NOT NULL,
    repository_id BIGINT,
    repository_name TEXT,
    sender_login TEXT,
    payload JSONB NOT NULL,
    status TEXT NOT NULL DEFAULT 'received',
    error_message TEXT,
    retry_count INTEGER NOT NULL DEFAULT 0,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    validated_at TIMESTAMPTZ,
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT webhook_events_status_check
        CHECK (status IN ('received', 'validated', 'processed', 'failed', 'ignored'))
);

CREATE UNIQUE INDEX webhook_events_dedupe_key_idx ON webhook_events (dedupe_key);
CREATE INDEX webhook_events_event_type_received_at_idx ON webhook_events (event_type, received_at DESC);
CREATE INDEX webhook_events_repository_name_received_at_idx ON webhook_events (repository_name, received_at DESC);
CREATE INDEX webhook_events_status_received_at_idx ON webhook_events (status, received_at DESC);
