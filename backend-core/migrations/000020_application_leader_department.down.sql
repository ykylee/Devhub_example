-- 000020_application_leader_department.down.sql

DROP INDEX IF EXISTS applications_dev_unit_idx;
DROP INDEX IF EXISTS applications_leader_idx;

ALTER TABLE applications
    DROP COLUMN IF EXISTS development_unit_id,
    DROP COLUMN IF EXISTS leader_user_id;
