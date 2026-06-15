# Session Handoff — mvs/work_260607-h-486-ci-runs-api

- 문서 목적: 본 sprint 의 진입 상태 + 산출물 + 다음 sprint 진입점을 인계한다.
- 범위: `POST /api/v1/ci-runs` 신규 handler (Gitea Actions webhook / 직접 생성 통합 endpoint).
- 대상 독자: 후속 sprint 진입자 (Claude 정식 인계 시), 리뷰어 (Codex 외부 리뷰).
- 상태: in_progress (scaffold 단계)
- 최종 수정일: 2026-06-07
- 관련 문서: [`work_backlog.md`](./work_backlog.md), [`state.json`](./state.json), [Issue #486](https://github.com/ykylee/Devhub_example/issues/486), [release_v1_roadmap.md §3.1 P0-4 + §3.5 N-7](../../../../docs/planning/release_v1_roadmap.md).
- 브랜치: `mvs/work_260607-h-486-ci-runs-api` (base `origin/main` @ `46568a8`).

## 1. 본 sprint 작업 목표

v1.0 release 차단 carve (P0-4) 해소. CI/CD 기능의 실질적 사용을 위한 CI Run ingest endpoint.

### Background
- `ci_runs` 테이블은 P1-7 sprint 에서 이미 main 머지됨 (repository-scoped GET 의 기반 자원)
- 그러나 **외부에서 신규 CI Run 을 등록하는 API 부재** — Gitea Actions webhook 이나 직접 생성 경로 없음
- 통합 테스트 ISSUE-05 시나리오 (2026-06-01) 가 본 API 없이는 stub 으로만 검증됨

### In-scope (본 sprint)
- `POST /api/v1/ci-runs` handler (backend Go, `internal/httpapi/ci_runs.go`)
- Request body: `{ repository_id, ref, status, commit_sha?, runner?, started_at?, finished_at? }`
- Status enum: `queued` / `running` / `success` / `failed` / `cancelled` / `skipped` / `unknown` (7종)
- Status validation: 위 7종 외 422
- Repository 가드: 존재 검증 + RBAC (`platform:read` + 동일 org unit subtree scope)
- Idempotency: `(repository_id, commit_sha, status, started_at)` unique key. 중복 시 409 + 기존 row 반환.
- Audit emit: `ci_run.created` (actor, repo, status, ref)
- Metric emit: `devhub_ci_runs_total{status}` Counter + `devhub_ci_run_ingest_duration_seconds` Histogram
- UT 5건: `ci_runs_test.go` (status 7종 PASS + 1종 422, idempotency 201+409, repository 404, RBAC 403)
- Traceability 정합: 4 문서 (report.md / API contract §11 / architecture §7 / requirements §5) 동시 갱신

### Out-of-scope (본 sprint)
- E2E `tests/e2e/ci-runs.spec.ts` — sprint -h 동시 작성 (Gemini 또는 Claude 본인이 진행 시 후속)
- webhook → handler fan-out 옵션 (기존 webhook entrypoint 보강) — v1.0 후속
- Frontend (repository detail 의 'Recent runs' widget) — Gemini 후속

## 2. 산출물 (예정)

- `internal/httpapi/ci_runs.go` — handler (status validation, RBAC, idempotency, audit, metric)
- `internal/httpapi/ci_runs_test.go` — UT 5건
- `internal/httpapi/router.go` (또는 동등한 route registration) — `POST /api/v1/ci-runs` 1 line 추가
- `docs/traceability/report.md` — §3.1 P0-4 row 갱신 (REQ-FR-106, ARCH-18, API-98, IMPL-ci-runs-01, UT-ci-runs-01, TC-CI-RUN-01..05)
- `docs/backend_api_contract.md §11` — API-98 entry 추가
- `docs/architecture.md §7` — ARCH-18 row 갱신
- `docs/requirements.md §5` — REQ-FR-106 row 갱신
- PR 1건 — 본 branch → main

## 3. 다음 sprint 진입 안내

본 sprint 머지 후:
1. **N-11 / #419 CI e2e+backend-integration job 복원** — Codex + 사용자 (의존성 깨면 본 sprint 의 go test 가 CI 에서 실행 안 됨)
2. **N-8 / #488 P1-6 Sign-out endpoint** — Claude sprint-i
3. **N-9 frontend widget** — Gemini (P1-7 GET endpoint 는 main 머지됨)

## 4. 인계 노트 (Claude 정식 인계 시)

- 본 session 은 Mavis (MiniMax-Code) 가 사용자 redirect 로 진입. Claude 정식 인계 시 본 scaffold + state.json 의 todo_done / todo 부터 이어가면 됨.
- `docs/governance/worker_division.md` §1.4 OpenCode Lane 정의에 따라 Mavis 가 직접 backend handler 작성은 Lane 2 (cross-cutting validation) 영역 외. **Lane 1 (workflow curation) 진입** 으로 분류. Claude 정식 진입 시 Lane 2/3 으로 재분류 필요.
- 본 sprint 의 핵심 위험: status enum 7종 + idempotency unique key 조합의 동시성 race. UT 의 parallel POST test 가 없으면 1차 CI 에서 통과해도 운영 race 가능. **시드 데이터 1 set + 2개 동시 POST goroutine race test** 권장.
