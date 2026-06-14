-- 000044: repositories 테이블 SCM create state 컬럼 추가 (X-4)
-- sprint: feat/x4-project-scm-create
-- ADR-0035 §3.6 (X-4 Project ↔ SCM create 연계)
-- 결정: post-commit Gitea API call 의 best-effort 보상 + 운영 가시성. sc_create_status 4 state.

BEGIN;

-- repositories 신규 컬럼 4개
-- sc_create_status: pending | success | failed | retry_scheduled
--   pending: SCM create 호출 전 (tx commit 직후, hook 실행 전)
--   success: Gitea API 200 + sc_external_id 설정
--   failed: Gitea API 4xx/5xx/timeout/network (운영자 manual retry 대기)
--   retry_scheduled: 운영자 retry API 호출됨 (follow-up)
-- sc_create_error: sc_create_status='failed' 일 때 error_class + message
-- sc_external_id: Gitea repo ID (success 시)
-- sc_create_at: success 시 now(). failed 시 last attempt 시각.
ALTER TABLE public.repositories
    ADD COLUMN IF NOT EXISTS sc_create_status text,
    ADD COLUMN IF NOT EXISTS sc_create_error text,
    ADD COLUMN IF NOT EXISTS sc_external_id bigint,
    ADD COLUMN IF NOT EXISTS sc_create_at timestamptz;

-- index: failed 상태 sc 조회 (운영자 retry queue)
CREATE INDEX IF NOT EXISTS idx_repositories_sc_create_failed
    ON public.repositories (sc_create_status, sc_create_at)
    WHERE sc_create_status = 'failed';

-- index: pending 상태 sc 조회 (운영 가시성)
CREATE INDEX IF NOT EXISTS idx_repositories_sc_create_pending
    ON public.repositories (sc_create_at)
    WHERE sc_create_status = 'pending';

COMMIT;
