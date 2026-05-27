-- 000042: archived project row 는 key 재사용 가능하도록 unique 범위를 active row 로 제한.
-- 배경: 현재 Delete는 hard delete가 아니라 archive(status='archived', archived_at=NOW())이며,
-- 기존 UNIQUE(repository_id, key) + standalone partial unique(repository_id IS NULL on key)
-- 는 archived row도 충돌 대상으로 포함한다.
-- 결과적으로 archive 후 동일 key 재생성이 project_key_conflict로 막힌다.
--
-- 정책 변경:
-- - active row(archived_at IS NULL)에서만 key uniqueness 강제
-- - repository 연결 project: (repository_id, key) unique where archived_at IS NULL
-- - standalone project: key unique where repository_id IS NULL AND archived_at IS NULL

ALTER TABLE projects
  DROP CONSTRAINT IF EXISTS projects_repository_key_unique;

DROP INDEX IF EXISTS projects_standalone_key_uniq;

CREATE UNIQUE INDEX IF NOT EXISTS projects_repository_key_active_uniq
  ON projects (repository_id, key)
  WHERE repository_id IS NOT NULL AND archived_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS projects_standalone_key_active_uniq
  ON projects (key)
  WHERE repository_id IS NULL AND archived_at IS NULL;

