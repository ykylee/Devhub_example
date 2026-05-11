-- 003_seed_test_admin.sql
--
-- Manual smoke-test 용 시스템 관리자 계정 (user_id=test). e2e 시드와는
-- 별개 — 개발자가 브라우저로 빠르게 들어갈 때 사용한다. Idempotent.

INSERT INTO users (user_id, email, display_name, role, status, joined_at, user_type)
VALUES ('test', 'test@example.com', 'Test Admin', 'system_admin', 'active', '2026-01-01', 'human')
ON CONFLICT (user_id) DO NOTHING;
