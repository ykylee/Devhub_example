-- 000008_audit_logs_actor_enrichment.up.sql
--
-- T-M1-04 (work_26_05_11-c): enrich audit_logs with operator-actor context
-- so security audit and incident triage can reconstruct who/where/which-request.
--
-- Per DEC-1 (NULL-allowed) the existing rows stay untouched -- the new columns
-- are populated for new writes only. Per DEC-2 the source_type vocabulary is
-- limited to oidc | webhook | system at this stage; future actor classes (cli,
-- api_token, ...) get added when they become real.
ALTER TABLE audit_logs
    ADD COLUMN source_ip   TEXT,
    ADD COLUMN request_id  TEXT,
    ADD COLUMN source_type TEXT;

-- Two lookup-friendly indexes. request_id is the primary correlation key when
-- chasing a single user-visible incident; source_type is useful when pivoting
-- by class of caller (oidc vs webhook vs system).
CREATE INDEX audit_logs_request_id_idx ON audit_logs (request_id) WHERE request_id IS NOT NULL;
CREATE INDEX audit_logs_source_type_idx ON audit_logs (source_type) WHERE source_type IS NOT NULL;
