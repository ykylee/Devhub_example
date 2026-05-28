-- 000039: standalone(repository_id NULL) project 의 key 중복 방지 (codex #349 P2).
-- 배경: 000037 이 repository_id 를 nullable 로 바꾼 뒤, 기존 UNIQUE(repository_id, key)
-- 제약이 PostgreSQL 의 NULL distinct 규칙 때문에 repository_id IS NULL row 들에 대해
-- key 중복을 막지 못한다. handler 는 project_key_conflict 를 invariant 로 취급하므로
-- repository 없는 standalone project 의 key 를 전역 유일하게 강제하는 partial unique
-- index 를 추가한다. (repository 연결 project 는 기존 UNIQUE(repository_id, key) 가 계속 담당.)

-- preflight (codex #354 P2): 000037~본 migration 사이 window 에서 이미 생성된 중복
-- standalone key 가 있으면 CREATE UNIQUE INDEX 가 즉시 실패해 (정작 버그 영향 받은)
-- 배포만 업그레이드 불가가 된다. 데이터 손실 없이 deterministic 재명명으로 unblock —
-- key 별 가장 먼저 생성된 row 는 유지, 나머지는 key 에 '-dup-<uuid>' suffix.
-- 운영자는 마이그레이션 후 -dup- 접미사 row 를 reconcile 한다.
WITH ranked AS (
  SELECT id, key,
         ROW_NUMBER() OVER (PARTITION BY key ORDER BY created_at, id) AS rn
  FROM projects
  WHERE repository_id IS NULL
)
UPDATE projects p
SET key = p.key || '-dup-' || p.id::text
FROM ranked r
WHERE p.id = r.id AND r.rn > 1;

CREATE UNIQUE INDEX IF NOT EXISTS projects_standalone_key_uniq
  ON projects (key)
  WHERE repository_id IS NULL;
