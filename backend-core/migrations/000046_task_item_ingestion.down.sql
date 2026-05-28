-- 000046 rollback: Task Item Ingestion 도메인 철회.

DROP TABLE IF EXISTS external_task_items CASCADE;

ALTER TABLE integration_providers
  DROP COLUMN IF EXISTS webhook_secret,
  DROP COLUMN IF EXISTS pull_interval_seconds,
  DROP COLUMN IF EXISTS last_pulled_at;

DROP SEQUENCE IF EXISTS task_webhook_seq;
