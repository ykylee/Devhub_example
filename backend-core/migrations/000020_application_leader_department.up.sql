-- 000020_application_leader_department.up.sql
-- Add explicit leader/department ownership for applications.

ALTER TABLE applications
    ADD COLUMN IF NOT EXISTS leader_user_id TEXT REFERENCES users(user_id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS development_unit_id TEXT REFERENCES org_units(unit_id) ON DELETE SET NULL;

-- Backfill: leader_user_id 가 비어 있는 기존 row 는 owner_user_id 로 메운다.
-- development_unit_id 는 기본값을 추정할 수 없어 NULL 로 두며, 운영 측에서 후속 채움.
UPDATE applications
   SET leader_user_id = owner_user_id
 WHERE leader_user_id IS NULL
   AND owner_user_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS applications_leader_idx ON applications (leader_user_id);
CREATE INDEX IF NOT EXISTS applications_dev_unit_idx ON applications (development_unit_id);
