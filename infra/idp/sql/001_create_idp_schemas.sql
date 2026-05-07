-- 001_create_idp_schemas.sql
--
-- 목적: ADR-0001 §8.1 결정 — 단일 devhub DB 안에 hydra, kratos schema 분리.
-- 적용:
--   psql -U postgres -d devhub -f infra/idp/sql/001_create_idp_schemas.sql
-- 또는 동일 SQL 을 직접 실행.
--
-- 비고:
-- - 본 SQL 은 schema 만 미리 만들어 둔다. 실제 테이블은
--   `hydra migrate sql` / `kratos migrate sql` 명령이 각 schema 안에 생성한다.
-- - PoC 단계는 postgres 슈퍼유저 그대로 사용. 운영 진입 시점에는
--   hydra/kratos 별 전용 role 분리를 별도로 결정한다.

CREATE SCHEMA IF NOT EXISTS hydra;
CREATE SCHEMA IF NOT EXISTS kratos;

COMMENT ON SCHEMA hydra IS 'Ory Hydra (OAuth2/OIDC server) — ADR-0001 도입';
COMMENT ON SCHEMA kratos IS 'Ory Kratos (identity & user management) — ADR-0001 도입';
