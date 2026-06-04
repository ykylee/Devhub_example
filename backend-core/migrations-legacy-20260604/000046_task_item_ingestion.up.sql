-- 000046: Task Item Ingestion 도메인 — external_task_items 테이블 + integration_providers 확장.
--
-- 배경:
--   외부 ALM/SCM/Issue Tracker(Jira, GitHub Issues, GitLab)의 작업 항목을
--   Webhook(실시간) + Pull(주기 동기화) 혼합 방식으로 수집한다.
--   기존 integration_providers 에 task_tracker type 을 추가하고,
--   webhook_secret/pull_interval_seconds/last_pulled_at 컬럼을 확장한다.
--
-- 참조:
--   docs/planning/task_item_ingestion_concept.md
--   docs/architecture.md §12

-- === 1. Webhook SEQ 전용 시퀀스 ===
CREATE SEQUENCE IF NOT EXISTS task_webhook_seq
  AS BIGINT
  INCREMENT BY 1
  MINVALUE 1
  NO MAXVALUE
  START WITH 1
  CACHE 1;

-- === 2. integration_providers 확장 ===
ALTER TABLE integration_providers
  ADD COLUMN IF NOT EXISTS webhook_secret TEXT,
  ADD COLUMN IF NOT EXISTS pull_interval_seconds INTEGER NOT NULL DEFAULT 1800,
  ADD COLUMN IF NOT EXISTS last_pulled_at TIMESTAMPTZ;

-- === 3. external_task_items 테이블 ===
CREATE TABLE IF NOT EXISTS external_task_items (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_id       UUID NOT NULL REFERENCES integration_providers(provider_id) ON DELETE CASCADE,
    external_id       TEXT NOT NULL,

    title             TEXT NOT NULL,
    description       TEXT,
    raw_status        TEXT NOT NULL,
    normalized_status TEXT,
    priority          TEXT,
    assignee          TEXT,
    reporter          TEXT,
    url               TEXT,
    labels            TEXT[],

    raw_payload       JSONB,
    webhook_seq       BIGINT,

    fetched_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ,

    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (provider_id, external_id)
);

-- webhook_seq uniqueness (nullable: partial index)
CREATE UNIQUE INDEX IF NOT EXISTS external_task_items_webhook_seq_uniq
    ON external_task_items (provider_id, webhook_seq)
    WHERE webhook_seq IS NOT NULL;
