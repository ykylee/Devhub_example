-- scripts/hrdb_etl_seed.sql
--
-- HR DB seed loader for the production `hrdb.persons` table introduced by
-- ADR-0008 + migration 000010 (sprint claude/work_260513-m). Idempotent
-- INSERT ... ON CONFLICT (system_id) DO UPDATE so the script can be re-run
-- daily without surprising side effects.
--
-- Usage (PoC / 운영 1차):
--   psql "$DEVHUB_DB_URL" -f scripts/hrdb_etl_seed.sql
--
-- For real production sync, replace the inline VALUES list with an ETL
-- worker that pulls from the upstream HR system (ADR-0008 §6). This file
-- is the documented entry point — operators are expected to maintain a
-- private fork with their real roster.
--
-- The three seeded rows mirror internal/hrdb/mock.go's MockClient so dev
-- and production share the same Sign Up flow against /api/v1/auth/signup.

BEGIN;

INSERT INTO hrdb.persons (system_id, employee_id, name, department_name, email, updated_at)
VALUES
    ('yklee',  '1001', 'YK Lee',    'Engineering',     'yklee@example.com',  NOW()),
    ('akim',   '1002', 'Alex Kim',  'Product',         'akim@example.com',   NOW()),
    ('sjones', '1003', 'Sam Jones', 'Infrastructure',  'sjones@example.com', NOW())
ON CONFLICT (system_id) DO UPDATE
SET employee_id     = EXCLUDED.employee_id,
    name            = EXCLUDED.name,
    department_name = EXCLUDED.department_name,
    email           = EXCLUDED.email,
    updated_at      = NOW();

COMMIT;
