-- migration 000032: audit_logs source_event_id dedup (sprint claude/work_260519-w).
--
-- ADR-0019 §5.3 (9) audit event listener Phase 2 PR-D — store-level dedup.
-- sprint -u (PR-B) 의 cursor + hash dedup 은 puller process 내부 dedup 가드라,
-- backend crash 후 재기동 + cursor revert, 또는 boundary race condition 등에서 동일
-- Keycloak event 가 audit_logs 에 중복 INSERT 될 가능성이 있다. 본 migration 의
-- partial UNIQUE INDEX 가 store layer 의 최종 가드.
--
-- partial 인 이유: 기존 audit_logs row (commands flow / Kratos hook / Gitea webhook
-- 등) 는 source_event_id 가 NULL 이라 unique 제약을 받지 않는다 (partial WHERE 조건).
-- Keycloak event listener (sprint -v PR-C wire) 처럼 deterministic source_event_id
-- 를 부여하는 신규 emitter 만 본 제약 영향. design 문서
-- docs/planning/keycloak_event_audit_integration.md §3.3 dedup state 보강.

ALTER TABLE audit_logs
    ADD COLUMN IF NOT EXISTS source_event_id TEXT;

-- partial UNIQUE INDEX — source_event_id 가 NULL 인 legacy row 에는 적용 안 됨.
-- (source_type, source_event_id) 조합으로 unique — 서로 다른 source_type 의 동일
-- event_id 충돌 회피 (예: keycloak_event 와 향후 다른 emitter 의 우연 hash 충돌).
--
-- WHERE source_type IS NOT NULL 도 명시 — PostgreSQL UNIQUE INDEX 의 NULL 비교는
-- 항상 distinct 로 취급되므로 source_type 이 NULL 이면 (NULL, 'hash-x') row 들이
-- dedup 안 됨. 운영 path 의 Keycloak emitter 는 항상 AuditSourceKeycloakEvent
-- 를 set 하지만, 잘못된 emitter (SourceType 빠뜨림) 의 회귀 방지 (sprint -w PR-D
-- Stage 3 self-review 보강).
CREATE UNIQUE INDEX IF NOT EXISTS audit_logs_source_event_id_uniq
    ON audit_logs (source_type, source_event_id)
    WHERE source_event_id IS NOT NULL AND source_type IS NOT NULL;
