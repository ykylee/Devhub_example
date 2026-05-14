# Session Handoff — claude/work_260514-c

- 브랜치: `claude/work_260514-c`
- Base: `main` @ `66ab5ff` (PR #106 머지 직후)
- 날짜: 2026-05-14
- 상태: in_progress
- 직전 sprint: `claude/work_260514-b` (PR #106 closed) — Application Design 2차 (API-41..50 activated).

## Sprint scope

**API-51..58 세트 활성화** — 4개 도메인 일괄.

1. **API-51..54 Repository 운영 지표 (read-only)**:
   - `GET /api/v1/repositories/:repository_id/activity` (commit 활동량, contributor)
   - `GET /api/v1/repositories/:repository_id/pull-requests` (PR 활동 타임라인)
   - `GET /api/v1/repositories/:repository_id/build-runs` (빌드 이력)
   - `GET /api/v1/repositories/:repository_id/quality-snapshots` (정적분석 결과)
2. **API-55..56 Project CRUD**:
   - `GET /api/v1/repositories/:repository_id/projects` (list)
   - `POST /api/v1/repositories/:repository_id/projects` (create)
   - `GET /api/v1/projects/:project_id` (detail)
   - `PATCH /api/v1/projects/:project_id` (update)
   - `DELETE /api/v1/projects/:project_id` (archive)
3. **API-57 Application 롤업**:
   - `GET /api/v1/applications/:application_id/rollup`
   - concept §13.4 weight_policy normalize 룰 (equal / repo_role / custom ± 0.001) 실 구현
   - meta (period / filters / weight_policy / applied_weights / fallbacks / data_gaps)
   - **active→closed critical 가드 흡수** (직전 sprint carve_out)
4. **API-58 Integration CRUD**:
   - `GET /api/v1/integrations` (list)
   - `POST /api/v1/integrations` (create, scope=application|project)
   - `PATCH /api/v1/integrations/:integration_id`
   - `DELETE /api/v1/integrations/:integration_id`

## 작업 순서
1. API-51..54 Repository 운영 지표 — store + handler + route + RBAC + UT
2. API-55..56 Project CRUD — handler + route + RBAC + UT (store 는 이미 준비)
3. API-57 Application 롤업 — store rollup + handler + active→closed 가드
4. API-58 Integration CRUD — store + handler + route + UT
5. 매트릭스 동기 + commit + push + PR

## 위험
- PR 크기 — 2500+ lines 예상. squash merge 단일 commit. 리뷰 부담 ↑.
- active→closed 가드 흡수 — 기존 호출자에 영향. self-review 시 명시 + 운영 문서 정합화.
- 롤업 계산 — 성능 최적화는 carve out, 1차 정확성 우선.

## 다음 세션 우선 작업

1. **postgres_applications integration test** — build-tagged 또는 skip-if-no-DB 패턴으로 store body 의 PostgreSQL 검증.
2. **frontend `/admin/settings/applications`** — IA + 화면 (별도 frontend sprint).
3. **ADR-0011 §4.2 enforceRowOwnership helper** — pmo_manager 활성화 sprint.
4. **롤업 cache / pre-aggregation** — 성능 최적화 (compute 가 다중 repo 의 N+1 query 사용 중).
5. **Repository 운영 지표의 commit activity** — 실제 commit 이벤트 ingest 도입 후 RepositoryActivity 보강.
6. **Integration policy 강화** — execution_system 정책의 cross-cut 검증 (실 외부 시스템 연동).
7. **quality_snapshots idempotency 결정** — UNIQUE 추가 vs history retention (PR #105 sequel).

## 본 sprint 결과 요약

- API-51..58 모두 activated (8 신규 endpoint group, 17 신규 route)
- store: repository_ops.go (Repository ops 4 메서드 + rollup compute + critical count) + integrations.go (CRUD 5 메서드)
- handler: repository_ops.go (4) + projects.go (5) + application_rollup.go (1) + integrations.go (4) = 14 신규
- ApplicationStore interface +11 메서드 (27 total)
- active→closed critical 가드 흡수 완료 (직전 sprint carve_out close)
- domain: PRActivity/BuildRun/QualitySnapshot/RepositoryActivity/WeightPolicy/RollupOptions/Meta/Period/Fallback/DataGap/Rollup + CustomWeightTolerance
- UT: projects(8) + integrations(6) + rollup(4) = 18 신규
- backend `go test -count=1 ./...` 12 패키지 PASS
