-- RM-ONBOARD-01 (ADR-0021 §3.3, ARCH-ONBOARD-05) — users 테이블에 onboarding state 컬럼 신규.
--
-- - onboarding_completed_at: 완료 시점 (UTC). NULL = 미완료 (limited 또는 onboarding 화면 강제 진입).
-- - review_status: 'pending_review' | 'reviewed'. NULL = onboarding 미제출.
--
-- bi-implication CHECK 제약: onboarding_completed_at NULL ↔ review_status NULL.
-- (제출 시 둘 다 set / 미제출 시 둘 다 NULL).
--
-- 기존 row (lazy-auto-created 사용자) 는 NULL 로 시작 → 다음 로그인 시 onboarding 강제 진입 정합.

ALTER TABLE users
    ADD COLUMN onboarding_completed_at TIMESTAMPTZ NULL,
    ADD COLUMN review_status TEXT NULL,
    ADD CONSTRAINT users_review_status_check
        CHECK (review_status IS NULL OR review_status IN ('pending_review', 'reviewed')),
    ADD CONSTRAINT users_onboarding_review_consistency
        CHECK ((onboarding_completed_at IS NULL) = (review_status IS NULL));

-- pending_review user 조회 가속 (admin /users?review_status=pending_review filter).
CREATE INDEX users_review_status_idx ON users (review_status)
    WHERE review_status IS NOT NULL;
