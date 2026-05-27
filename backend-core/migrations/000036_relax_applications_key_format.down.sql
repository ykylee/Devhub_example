-- 000036 rollback: applications.key CHECK 원복 (1~10자 -> 10자 고정)

ALTER TABLE applications
  DROP CONSTRAINT IF EXISTS applications_key_format;

ALTER TABLE applications
  ADD CONSTRAINT applications_key_format
  CHECK (key ~ '^[A-Za-z0-9]{10}$');
