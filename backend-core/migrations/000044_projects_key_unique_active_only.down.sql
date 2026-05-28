-- 000044 rollback: 기존 unique 정책 복원 (archived 포함 충돌).

DROP INDEX IF EXISTS projects_standalone_key_active_uniq;
DROP INDEX IF EXISTS projects_repository_key_active_uniq;

ALTER TABLE projects
  ADD CONSTRAINT projects_repository_key_unique UNIQUE (repository_id, key);

CREATE UNIQUE INDEX IF NOT EXISTS projects_standalone_key_uniq
  ON projects (key)
  WHERE repository_id IS NULL;

