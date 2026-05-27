-- 000035: applications.key CHECK 완화 (10자 고정 -> 1~10자)
-- 배경:
-- - API handler validation 이 ^[A-Za-z0-9]{1,10}$ 로 완화된 이후에도
--   DB CHECK 가 ^[A-Za-z0-9]{10}$ 이면 short key 입력이 PG CHECK 위반으로 실패.
-- - 본 migration 은 DB 제약을 API 규칙과 정합시킨다.

ALTER TABLE applications
  DROP CONSTRAINT IF EXISTS applications_key_format;

ALTER TABLE applications
  ADD CONSTRAINT applications_key_format
  CHECK (key ~ '^[A-Za-z0-9]{1,10}$');

