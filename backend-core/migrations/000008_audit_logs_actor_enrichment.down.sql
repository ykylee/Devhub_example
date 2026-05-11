-- 000008_audit_logs_actor_enrichment.down.sql
DROP INDEX IF EXISTS audit_logs_source_type_idx;
DROP INDEX IF EXISTS audit_logs_request_id_idx;
ALTER TABLE audit_logs
    DROP COLUMN IF EXISTS source_type,
    DROP COLUMN IF EXISTS request_id,
    DROP COLUMN IF EXISTS source_ip;
