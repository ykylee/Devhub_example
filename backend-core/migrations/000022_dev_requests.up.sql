-- 000022_dev_requests.up.sql
-- DREQ 도메인 1차 backend 활성화 (sprint claude/work_260515-i, ARCH-DREQ-05).
--
-- 외부 시스템에서 들어온 1건의 개발 작업 의뢰. 6-state lifecycle:
--   received → pending → in_review → registered | rejected | closed
-- 모든 상태 전이는 audit_logs 에 dev_request.* action 으로 기록 (REQ-NFR-DREQ-003).

CREATE TABLE dev_requests (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title                   TEXT NOT NULL,
    details                 TEXT NOT NULL DEFAULT '',
    requester               TEXT NOT NULL,
    assignee_user_id        TEXT NOT NULL REFERENCES users(user_id) ON DELETE RESTRICT,
    source_system           TEXT NOT NULL,
    external_ref            TEXT,
    status                  TEXT NOT NULL,
    registered_target_type  TEXT,
    registered_target_id    TEXT,
    rejected_reason         TEXT,
    received_at             TIMESTAMPTZ NOT NULL,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT dev_requests_status_check
        CHECK (status IN ('received', 'pending', 'in_review', 'registered', 'rejected', 'closed')),

    CONSTRAINT dev_requests_target_type_check
        CHECK (registered_target_type IS NULL
               OR registered_target_type IN ('application', 'project')),

    -- registered 상태일 때만 target_type/id 가 채워진다 (REQ-FR-DREQ-005).
    CONSTRAINT dev_requests_registered_consistency
        CHECK (
            (status = 'registered') = (registered_target_type IS NOT NULL AND registered_target_id IS NOT NULL)
        ),

    -- rejected 상태일 때만 rejected_reason 이 채워진다 (REQ-FR-DREQ-006).
    CONSTRAINT dev_requests_rejected_reason_required
        CHECK (
            (status = 'rejected') = (rejected_reason IS NOT NULL AND rejected_reason <> '')
        )
);

-- idempotency: 동일 외부 시스템에서 같은 external_ref 로 재호출 시 동일 row 반환.
-- (REQ-NFR-DREQ-002). external_ref 가 NULL 인 경우는 idempotency 대상 외.
CREATE UNIQUE INDEX dev_requests_idempotency_uniq
    ON dev_requests (source_system, external_ref)
    WHERE external_ref IS NOT NULL;

-- 담당자 dashboard 의 "내 대기 의뢰" 위젯 + 목록 필터를 위한 인덱스.
CREATE INDEX dev_requests_assignee_status_idx ON dev_requests (assignee_user_id, status);
CREATE INDEX dev_requests_status_idx ON dev_requests (status);
CREATE INDEX dev_requests_source_system_idx ON dev_requests (source_system);
