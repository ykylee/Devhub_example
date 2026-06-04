-- migration 000032 down: drop partial UNIQUE INDEX + source_event_id column.

DROP INDEX IF EXISTS audit_logs_source_event_id_uniq;

ALTER TABLE audit_logs
    DROP COLUMN IF EXISTS source_event_id;
