-- 000010: HR DB schema + persons table (ADR-0008, RM-M3-02).
--
-- DevHub PostgreSQL 안에 별도 schema `hrdb` 를 두고, 인사 DB 마스터 데이터의
-- 사본을 보관한다. 외부 HR 시스템에서 daily / on-demand ETL 로 sync. 자세한
-- 결정은 docs/adr/0008-hrdb-production-adapter.md.

CREATE SCHEMA IF NOT EXISTS hrdb;

CREATE TABLE IF NOT EXISTS hrdb.persons (
    system_id        TEXT PRIMARY KEY,
    employee_id      TEXT NOT NULL UNIQUE,
    name             TEXT NOT NULL,
    department_name  TEXT NOT NULL,
    email            TEXT,
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS persons_employee_id_idx ON hrdb.persons (employee_id);
CREATE INDEX IF NOT EXISTS persons_name_lower_idx ON hrdb.persons (lower(name));
