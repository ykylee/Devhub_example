# Session Handoff — claude/work_260514-b

- 브랜치: `claude/work_260514-b`
- Base: `main` @ `642d976` (PR #105 머지 직후)
- 날짜: 2026-05-14
- 상태: in_progress
- 직전 sprint: `claude/work_260514-a` (PR #105 closed) — Application Design 1차 scaffolded.

## Sprint scope

**Application Design 2차** — A1+A2+A3 묶음.

- (A1) `postgres_applications.go` 의 16 store 메서드 본체 구현 (CRUD + UpdateSync + Archive).
- (A2) API-41~50 handler body — body binding + 검증 + 상태 전이 가드 (concept §13.2.1) + immutable key + provider 검증 + audit emit.
- (A3) UT-application-store-XX + UT-application-handler-XX 1차 (happy + edge + RBAC denial).

직전 sprint 의 scaffolded (`ErrNotImplemented` stub + 501 응답) → 본 sprint 의 activated (정식 응답 + 가드 + audit).

## 진입 컨텍스트

### 직전 sprint 의 carve_out 흡수 항목
- `postgres_applications.go` body (T33)
- handler body + validation + audit emit (T34)
- UT 1차 (T35)

### 본 sprint 의 carve_out (= 후속 sprint 처리)
- API-51..58 활성화 (Repository 운영 지표 / Project CRUD / 롤업 / Integration)
- e2e TC 작성
- frontend UI
- 실 SCM 어댑터 구현
- ADR-0011 §4.2 enforceRowOwnership 1차 구현 (pmo_manager 활성화 sprint)

## 작업 계획

### 작업 순서
1. backend `internal/store/postgres_applications.go` 신규 — 16 메서드 PostgreSQL 본체. (T33)
2. backend `internal/httpapi/applications.go` body 교체 — 10 handler 실 응답 + 가드 + audit. (T34)
3. backend `internal/httpapi/applications_test.go` 신규 + `internal/store/postgres_applications_test.go` 신규 — UT 1차. (T35)
4. trace.md / backend_api_contract.md 갱신. (T36)

### 위험
- PR 크기 — 1500+ lines 예상. 리뷰 부담 ↑. 1차 평가 후 분할 가능성 검토.
- 상태 전이 가드의 일부 (active→closed 의 critical 롤업 0건) 는 롤업 store 의존. 1차에서는 가드 단순화 (활성 Repository ≥1 만), 롤업 가드는 carve out 후속.
- audit action 명명 — `application.list.requested` 등 패턴. 기존 audit 매트릭스와 정합 확인.

## 다음 세션 우선 작업 (본 sprint 결과 반영)

1. **API-51..54 활성화 (Repository 운영 지표)** — pr_activities / build_runs / quality_snapshots 조회 endpoint. ingest pipeline 연계.
2. **API-55..56 활성화 (Project CRUD)** — 본 sprint 의 ApplicationStore 인터페이스에 Project 메서드는 이미 구현 완료. handler + route 등록만 추가하면 됨.
3. **API-57 Application 롤업** — concept §13.4 weight_policy normalize 실 구현. `active → closed` 의 critical 롤업 0건 가드도 함께 흡수 (본 sprint 의 carve out).
4. **API-58 Integration CRUD** — project_integrations 의 scope polymorphism handler.
5. **postgres_applications integration test** — build-tagged or skip-if-no-DB 패턴. composite PK 위반 / FK 위반 / archived consistency 등 SQL 레벨 검증.
6. **Frontend `/admin/settings/applications`** — IA + 화면 (별도 frontend sprint).
7. **enforceRowOwnership helper (ADR-0011 §4.2)** — pmo_manager 활성화 sprint.
8. **quality_snapshots idempotency 결정** — UNIQUE 추가 vs history retention.

## 본 sprint 결과 요약

- store body 16 메서드 PostgreSQL 구현 (550+ lines)
- handler body 10 endpoint 정식 응답 (550+ lines, 가드 + audit emit 포함)
- 19 handler 단위테스트 (memoryApplicationStore + 핵심 case)
- ApplicationStore interface 추가 + main.go 와이어링
- delete repo_key path 변수: gin catch-all (`*repo_key`) + leading `/` strip
- backend `go test -count=1 ./...` 12 패키지 PASS
