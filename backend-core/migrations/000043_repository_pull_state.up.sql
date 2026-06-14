-- 000043: repository_pull_state 신규 테이블 (X-5 Gitea Hourly Pull 정밀화)
-- sprint: feat/x5-gitea-hourly-pull
-- ADR-0034 §1.2 (per-repository state 추적 결정)
-- 결정: per-repository last_pull_at + consecutive_failures + exponential backoff (24h cap) + 5회 연속 실패 시 alert
-- v1.0 webhook only 의 한계 보완 (누락 복구 / historical backfill / 운영 가시성).

BEGIN;

-- repository_pull_state
-- repository_id: PK, FK to repositories (one-to-one). Gitea SCM provider 한정 (homelab 는 자체 metric 보유).
-- last_pull_at: 마지막 successful pull 시각. NULL = never pulled.
-- last_pull_status: success|error|partial. last_pull_error 와 짝.
-- last_pull_error: error_class + error_message. NULL when last_pull_status='success'.
-- consecutive_failures: 연속 error/partial 횟수. success 시 0 reset.
-- backoff_until: now() < backoff_until 시 skip (exponential: 2^failures, cap 24h).
-- updated_at: trigger or manual.
-- last_alert_at: 5회 연속 실패 시 alert emit 시각 (idempotent, 같은 incident 의 중복 alert 방지). NULL = never alerted.
CREATE TABLE IF NOT EXISTS public.repository_pull_state (
    repository_id uuid PRIMARY KEY REFERENCES public.repositories(id) ON DELETE CASCADE,
    last_pull_at timestamptz,
    last_pull_status text,
    last_pull_error text,
    consecutive_failures integer NOT NULL DEFAULT 0 CHECK (consecutive_failures >= 0),
    backoff_until timestamptz,
    updated_at timestamptz NOT NULL DEFAULT now(),
    last_alert_at timestamptz
);

-- index: backoff 조회 (전체 repo 의 due-or-not scan 용)
CREATE INDEX IF NOT EXISTS idx_repository_pull_state_backoff
    ON public.repository_pull_state (backoff_until)
    WHERE backoff_until IS NOT NULL;

-- index: consecutive_failures 조회 (alert 임계 scan 용)
CREATE INDEX IF NOT EXISTS idx_repository_pull_state_consecutive_failures
    ON public.repository_pull_state (consecutive_failures)
    WHERE consecutive_failures >= 1;

-- updated_at auto-update trigger
CREATE OR REPLACE FUNCTION public.touch_repository_pull_state_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_repository_pull_state_updated_at ON public.repository_pull_state;
CREATE TRIGGER trg_repository_pull_state_updated_at
    BEFORE UPDATE ON public.repository_pull_state
    FOR EACH ROW EXECUTE FUNCTION public.touch_repository_pull_state_updated_at();

COMMIT;
