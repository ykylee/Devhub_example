-- 000014: Application-Repository 연결 (REQ-FR-APP-002 / 004 / 011, ERD §2.5, concept §13.3).
--
-- composite PK = (application_id, repo_provider, repo_full_name) — 동일 repo_full_name 이 서로
-- 다른 provider 에 존재할 수 있으므로 provider 를 PK 에 포함 (concept §13.3 + ERD §2.5 합성 키 메모).
-- sync_status / sync_error_code 는 link 단위 최신 1건 캐시 (api §13.3 운영 룰). event 단위 상세는
-- webhook_events (현행) 또는 후속 adapter_event_logs 에 보관.

CREATE TABLE application_repositories (
    application_id        UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    repo_provider         TEXT NOT NULL REFERENCES scm_providers(provider_key) ON DELETE RESTRICT,
    repo_full_name        TEXT NOT NULL,
    external_repo_id      TEXT,
    role                  TEXT NOT NULL,
    sync_status           TEXT NOT NULL DEFAULT 'requested',
    sync_error_code       TEXT,
    sync_error_retryable  BOOLEAN,
    sync_error_at         TIMESTAMPTZ,
    last_sync_at          TIMESTAMPTZ,
    linked_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (application_id, repo_provider, repo_full_name),
    CONSTRAINT application_repositories_role_check CHECK (role IN ('primary', 'sub', 'shared')),
    CONSTRAINT application_repositories_sync_status_check CHECK (
        sync_status IN ('requested', 'verifying', 'active', 'degraded', 'disconnected')
    ),
    CONSTRAINT application_repositories_sync_error_code_check CHECK (
        sync_error_code IS NULL OR sync_error_code IN (
            'provider_unreachable',
            'auth_invalid',
            'permission_denied',
            'rate_limited',
            'webhook_signature_invalid',
            'payload_schema_mismatch',
            'resource_not_found',
            'internal_adapter_error'
        )
    ),
    CONSTRAINT application_repositories_sync_error_consistency CHECK (
        (sync_error_code IS NULL AND sync_error_retryable IS NULL AND sync_error_at IS NULL)
        OR (sync_error_code IS NOT NULL AND sync_error_retryable IS NOT NULL AND sync_error_at IS NOT NULL)
    )
);

CREATE INDEX application_repositories_provider_repo_idx ON application_repositories (repo_provider, repo_full_name);
CREATE INDEX application_repositories_sync_status_idx ON application_repositories (sync_status);
CREATE INDEX application_repositories_external_repo_id_idx ON application_repositories (repo_provider, external_repo_id)
    WHERE external_repo_id IS NOT NULL;
