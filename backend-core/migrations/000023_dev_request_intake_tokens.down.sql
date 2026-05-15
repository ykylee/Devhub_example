-- 000023 down: dev_request_intake_tokens 테이블 제거.

DROP INDEX IF EXISTS dev_request_intake_tokens_active_idx;
DROP TABLE IF EXISTS dev_request_intake_tokens;
