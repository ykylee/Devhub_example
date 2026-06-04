-- 000037: 프로젝트 독립 생성 지원 — projects.repository_id nullable 허용

ALTER TABLE projects
  ALTER COLUMN repository_id DROP NOT NULL;

