-- 000025_dev_requests_assignee_nullable.up.sql
-- DREQ codex hotfix #3 P1 (sprint claude/work_260515-k).
--
-- 외부 수신 의뢰가 검증 실패 (invalid_intake) 시에도 row 자체는 저장해야
-- audit 보존 가능 (REQ-FR-DREQ-002, drop 안 함). 그러나 기존 schema 는
-- assignee_user_id 가 NOT NULL + FK 라 미존재 assignee 면 insert 자체 실패
-- → row drop. 본 마이그레이션이 NULL 허용으로 alter 하여 invalid_intake
-- 행이 assignee_user_id = NULL 로 저장 가능하게 한다.
--
-- 유효한 의뢰는 application 단의 검증 + (status='pending' 이면 assignee 필수)
-- CHECK 제약으로 무결성 유지.

ALTER TABLE dev_requests ALTER COLUMN assignee_user_id DROP NOT NULL;

-- status='pending'/'in_review'/'registered' 인 row 는 assignee_user_id 가 필수.
-- 'rejected'/'closed' 행만 NULL 허용. 'received' 는 transient 라 거의 발생 안 함.
ALTER TABLE dev_requests
    ADD CONSTRAINT dev_requests_assignee_required_when_active
    CHECK (
        status IN ('rejected', 'closed', 'received')
        OR assignee_user_id IS NOT NULL
    );
