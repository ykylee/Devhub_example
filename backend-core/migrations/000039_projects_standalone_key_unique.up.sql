-- 000039: standalone(repository_id NULL) project 의 key 중복 방지 (codex #349 P2).
-- 배경: 000037 이 repository_id 를 nullable 로 바꾼 뒤, 기존 UNIQUE(repository_id, key)
-- 제약이 PostgreSQL 의 NULL distinct 규칙 때문에 repository_id IS NULL row 들에 대해
-- key 중복을 막지 못한다. handler 는 project_key_conflict 를 invariant 로 취급하므로
-- repository 없는 standalone project 의 key 를 전역 유일하게 강제하는 partial unique
-- index 를 추가한다. (repository 연결 project 는 기존 UNIQUE(repository_id, key) 가 계속 담당.)

CREATE UNIQUE INDEX IF NOT EXISTS projects_standalone_key_uniq
  ON projects (key)
  WHERE repository_id IS NULL;
