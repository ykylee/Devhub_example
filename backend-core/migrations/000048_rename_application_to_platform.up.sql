-- 000048: Rename "application" to "platform" across all schemas.
--
-- Baseline reset (#477) 가 initial schema 자체를 platform naming 으로 생성하게
-- 변경했으므로 본 migration 은 no-op. 유지보수/rollback 경로를 위해 파일만 유지.
--
-- Down: 000048_rename_application_to_platform.down.sql
SELECT 1;
