-- 000006: dev_request_vocs + user_notifications 신규 테이블
-- sprint: feat/work_260612-a-dev-requests-voc-domain
-- ADR-0028: voc 도메인 도입 (project_id 미정인 의뢰 staging 공간) + in-app notification

BEGIN;

-- (1) dev_request_vocs 신규 테이블
CREATE TABLE IF NOT EXISTS public.dev_request_vocs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    external_ref text NOT NULL,
    source_system text NOT NULL,
    title text NOT NULL,
    details text NOT NULL DEFAULT '',
    requester text NOT NULL DEFAULT '',
    req_department text NOT NULL DEFAULT '',
    assignee_user_id text NULL REFERENCES public.users(user_id) ON DELETE SET NULL,
    dev_department text NOT NULL DEFAULT '',
    request_date date NULL,
    dev_schedule text NOT NULL DEFAULT '',
    status text NOT NULL DEFAULT 'received'
        CHECK (status IN ('received', 'routed', 'closed')),
    project_id uuid NULL REFERENCES public.platforms(id) ON DELETE SET NULL,
    dev_request_id uuid NULL REFERENCES public.dev_requests(id) ON DELETE SET NULL,
    routed_at timestamptz NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT dev_request_vocs_external_ref_source_uniq
        UNIQUE (source_system, external_ref)
);

CREATE INDEX IF NOT EXISTS dev_request_vocs_assignee_idx
    ON public.dev_request_vocs(assignee_user_id)
    WHERE assignee_user_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS dev_request_vocs_status_idx
    ON public.dev_request_vocs(status);
CREATE INDEX IF NOT EXISTS dev_request_vocs_project_idx
    ON public.dev_request_vocs(project_id)
    WHERE project_id IS NOT NULL;

-- (2) dev_requests 테이블 확장 (voc 의 4 field 추가)
ALTER TABLE public.dev_requests
    ADD COLUMN IF NOT EXISTS req_department text NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS dev_department text NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS request_date date NULL,
    ADD COLUMN IF NOT EXISTS dev_schedule text NOT NULL DEFAULT '';

-- (3) user_notifications 신규 테이블 (in-app notification)
CREATE TABLE IF NOT EXISTS public.user_notifications (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id text NOT NULL REFERENCES public.users(user_id) ON DELETE CASCADE,
    kind text NOT NULL,
    ref_id text NULL,
    title text NOT NULL,
    body text NOT NULL DEFAULT '',
    read_at timestamptz NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS user_notifications_user_unread_idx
    ON public.user_notifications(user_id, created_at DESC)
    WHERE read_at IS NULL;
CREATE INDEX IF NOT EXISTS user_notifications_user_all_idx
    ON public.user_notifications(user_id, created_at DESC);

COMMIT;
