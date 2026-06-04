-- 000025 down: NULL assignee_user_id 인 row 제거 후 NOT NULL 복구.

ALTER TABLE dev_requests
    DROP CONSTRAINT IF EXISTS dev_requests_assignee_required_when_active;

-- NULL assignee 인 row 는 rollback 환경에서 schema 와 양립 못 함 → 삭제.
DELETE FROM dev_requests WHERE assignee_user_id IS NULL;

ALTER TABLE dev_requests ALTER COLUMN assignee_user_id SET NOT NULL;
