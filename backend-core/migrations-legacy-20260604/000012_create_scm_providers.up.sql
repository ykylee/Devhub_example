-- 000012: SCM Provider catalog (REQ-FR-APP-004 / REQ-FR-APP-009, ERD §2.5, concept §12).
--
-- 형상관리 도구 어댑터 카탈로그. provider_key 가 application_repositories.repo_provider 의
-- FK 후보 (현재는 application 검증 단에서 강제, ADR-0011 1차).
-- adapter_version 은 어댑터 배포 파이프라인 외 수정 금지 (concept §12.2 + api §13.1.1).

CREATE TABLE scm_providers (
    provider_key      TEXT PRIMARY KEY,
    display_name      TEXT NOT NULL,
    enabled           BOOLEAN NOT NULL DEFAULT TRUE,
    adapter_version   TEXT NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT scm_providers_key_format CHECK (provider_key ~ '^[a-z][a-z0-9_-]{0,30}$'),
    CONSTRAINT scm_providers_version_format CHECK (adapter_version ~ '^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$')
);

CREATE INDEX scm_providers_enabled_idx ON scm_providers (enabled);

-- Seed 4 providers (REQ-FR-APP-004 의 동등 지원 정책). adapter_version 은 어댑터 미구현 단계라
-- 0.0.1 placeholder. 어댑터 배포 시 배포 파이프라인이 갱신.
INSERT INTO scm_providers (provider_key, display_name, enabled, adapter_version) VALUES
    ('bitbucket', 'Bitbucket',  TRUE, '0.0.1'),
    ('gitea',     'Gitea',      TRUE, '0.0.1'),
    ('forgejo',   'Forgejo',    TRUE, '0.0.1'),
    ('github',    'GitHub',     TRUE, '0.0.1');
