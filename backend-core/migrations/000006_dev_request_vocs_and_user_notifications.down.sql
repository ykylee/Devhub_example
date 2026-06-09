-- 000006: dev_request_vocs + user_notifications 신규 테이블 (rollback)
-- sprint: feat/work_260612-a-dev-requests-voc-domain

BEGIN;

-- (3) → (2) → (1) 순서로 reverse (FK 의존성)
DROP TABLE IF EXISTS public.user_notifications;

ALTER TABLE public.dev_requests
    DROP COLUMN IF EXISTS dev_schedule,
    DROP COLUMN IF EXISTS request_date,
    DROP COLUMN IF EXISTS dev_department,
    DROP COLUMN IF EXISTS req_department;

DROP TABLE IF EXISTS public.dev_request_vocs;

COMMIT;
