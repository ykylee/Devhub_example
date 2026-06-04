-- 000013: Applications — 제품 수명 주기 총괄 단위 (REQ-FR-APP-001..012, ERD §2.5, concept §13).
--
-- id        : 내부 UUID PK (API path 식별자 — `/api/v1/applications/{application_id}`)
-- key       : 사용자가 보는 immutable 10자 영문숫자 식별자 (REQ-FR-APP-003)
-- status    : 5종 상태 머신 (concept §13.2.1 가드 표)
-- visibility: 3종 공개 범위 (api §13.1 visibility 표)
-- owner_user_id : users.user_id 의 TEXT 참조 (기존 DevHub 패턴 — unit_appointments 와 동일)

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE applications (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key                TEXT NOT NULL UNIQUE,
    name               TEXT NOT NULL,
    description        TEXT,
    status             TEXT NOT NULL,
    visibility         TEXT NOT NULL,
    owner_user_id      TEXT REFERENCES users(user_id) ON DELETE SET NULL,
    start_date         DATE,
    due_date           DATE,
    archived_at        TIMESTAMPTZ,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT applications_key_format CHECK (key ~ '^[A-Za-z0-9]{10}$'),
    CONSTRAINT applications_status_check CHECK (status IN ('planning', 'active', 'on_hold', 'closed', 'archived')),
    CONSTRAINT applications_visibility_check CHECK (visibility IN ('public', 'internal', 'restricted')),
    CONSTRAINT applications_archived_consistency CHECK (
        (status = 'archived' AND archived_at IS NOT NULL)
        OR (status <> 'archived' AND archived_at IS NULL)
    ),
    CONSTRAINT applications_due_date_after_start CHECK (
        start_date IS NULL OR due_date IS NULL OR due_date >= start_date
    )
);

CREATE INDEX applications_status_idx ON applications (status);
CREATE INDEX applications_visibility_idx ON applications (visibility);
CREATE INDEX applications_owner_idx ON applications (owner_user_id);
CREATE INDEX applications_archived_at_idx ON applications (archived_at) WHERE archived_at IS NOT NULL;
