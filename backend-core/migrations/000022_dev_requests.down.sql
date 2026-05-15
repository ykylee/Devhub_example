-- 000022 down: dev_requests 테이블 제거.

DROP INDEX IF EXISTS dev_requests_source_system_idx;
DROP INDEX IF EXISTS dev_requests_status_idx;
DROP INDEX IF EXISTS dev_requests_assignee_status_idx;
DROP INDEX IF EXISTS dev_requests_idempotency_uniq;

DROP TABLE IF EXISTS dev_requests;
