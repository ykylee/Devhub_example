-- 000037 rollback: projects.repository_id NOT NULL 원복
-- 주의: NULL row 존재 시 rollback 실패 가능.

ALTER TABLE projects
  ALTER COLUMN repository_id SET NOT NULL;

