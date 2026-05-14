-- 000020_application_leader_department.up.sql
-- Add explicit leader/department ownership for applications.

ALTER TABLE applications
    ADD COLUMN IF NOT EXISTS leader_user_id TEXT REFERENCES users(user_id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS development_unit_id TEXT REFERENCES org_units(unit_id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS applications_leader_idx ON applications (leader_user_id);
CREATE INDEX IF NOT EXISTS applications_dev_unit_idx ON applications (development_unit_id);
