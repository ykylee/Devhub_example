-- 002_seed_e2e_users.sql
--
-- E2E 시나리오 (frontend/tests/e2e/fixtures.ts SEEDED) 가 요구하는
-- 3 사용자를 DevHub users 테이블에 시드한다. Kratos identity 시드는
-- 별개 (admin REST API curl, docs/setup/e2e-test-guide.md §2).
--
-- Idempotent: 같은 user_id 가 이미 있으면 무시 (ON CONFLICT). e2e 회귀
-- 사이클마다 재실행해도 안전.

INSERT INTO users (user_id, email, display_name, role, status, joined_at, user_type)
VALUES
    ('alice',   'alice@example.com',   'Alice',   'developer',    'active', '2026-01-01', 'human'),
    ('bob',     'bob@example.com',     'Bob',     'manager',      'active', '2026-01-01', 'human'),
    ('charlie', 'charlie@example.com', 'Charlie', 'system_admin', 'active', '2026-01-01', 'human')
ON CONFLICT (user_id) DO NOTHING;
