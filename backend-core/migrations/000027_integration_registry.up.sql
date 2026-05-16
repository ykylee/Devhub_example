CREATE TABLE integration_providers (
    provider_id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_key      TEXT NOT NULL UNIQUE,
    provider_type     TEXT NOT NULL,
    display_name      TEXT NOT NULL,
    enabled           BOOLEAN NOT NULL DEFAULT TRUE,
    auth_mode         TEXT NOT NULL,
    credentials_ref   TEXT NOT NULL,
    capabilities      JSONB NOT NULL DEFAULT '[]'::jsonb,
    sync_status       TEXT NOT NULL DEFAULT 'requested',
    last_sync_at      TIMESTAMPTZ NULL,
    last_error_code   TEXT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT integration_providers_type_check CHECK (provider_type IN ('alm', 'scm', 'ci_cd', 'doc', 'infra')),
    CONSTRAINT integration_providers_auth_mode_check CHECK (auth_mode IN ('token', 'basic', 'oauth2', 'app_password', 'agent')),
    CONSTRAINT integration_providers_sync_status_check CHECK (sync_status IN ('requested', 'verifying', 'active', 'degraded', 'disconnected'))
);

CREATE TABLE integration_bindings (
    binding_id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scope_type        TEXT NOT NULL,
    scope_id          TEXT NOT NULL,
    provider_id       UUID NOT NULL REFERENCES integration_providers(provider_id) ON DELETE CASCADE,
    external_key      TEXT NOT NULL,
    policy            TEXT NOT NULL,
    enabled           BOOLEAN NOT NULL DEFAULT TRUE,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT integration_bindings_scope_type_check CHECK (scope_type IN ('application', 'project')),
    CONSTRAINT integration_bindings_policy_check CHECK (policy IN ('summary_only', 'execution_system')),
    UNIQUE (scope_type, scope_id, provider_id, external_key)
);

CREATE TABLE integration_sync_jobs (
    job_id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_id       UUID NOT NULL REFERENCES integration_providers(provider_id) ON DELETE CASCADE,
    requested_by      TEXT NULL,
    status            TEXT NOT NULL DEFAULT 'queued',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT integration_sync_jobs_status_check CHECK (status IN ('queued', 'running', 'succeeded', 'failed'))
);

CREATE INDEX integration_providers_type_enabled_idx ON integration_providers (provider_type, enabled);
CREATE INDEX integration_bindings_scope_idx ON integration_bindings (scope_type, scope_id);
CREATE INDEX integration_bindings_provider_idx ON integration_bindings (provider_id);
CREATE INDEX integration_sync_jobs_provider_created_idx ON integration_sync_jobs (provider_id, created_at DESC);
