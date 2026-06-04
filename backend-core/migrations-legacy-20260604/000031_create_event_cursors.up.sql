-- migration 000031: event_cursors table for Keycloak event polling (sprint claude/work_260519-u).
--
-- ADR-0019 §5.3 audit event listener / SPI → DevHub audit_logs 통합 의 Phase 2 PR-B
-- 구현. design 문서 docs/planning/keycloak_event_audit_integration.md §3.3 의 state
-- 영구화 (dedup) 패턴. golang-migrate sequence 정합 (sprint -l 의 000021 충돌
-- 정정 + CI lint sprint -o 의 회귀 차단 적용 후 신규).
--
-- 사용 시나리오: KeycloakEventPuller (internal/audit/keycloak_event_puller.go) 가 매
-- poll 주기마다 cursor_key 의 last_event_at 이후의 Keycloak event 만 처리. crash
-- recovery 시 마지막 cursor 부터 재개 (at-least-once delivery — audit_logs 의
-- UNIQUE 제약으로 중복 INSERT 방지 carve).

CREATE TABLE IF NOT EXISTS event_cursors (
    cursor_key      TEXT PRIMARY KEY,
    last_event_at   TIMESTAMPTZ NOT NULL,
    last_event_hash TEXT,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- cursor_key 값 reference (운영자가 manual 조회 시):
--   'keycloak.events'       — user events (LOGIN / LOGOUT / REGISTER / UPDATE_PASSWORD 등)
--   'keycloak.events.admin' — admin events (USER:CREATE / USER:UPDATE / USER:DELETE / ROLE:CREATE 등)

CREATE INDEX IF NOT EXISTS event_cursors_updated_at_idx ON event_cursors (updated_at);
