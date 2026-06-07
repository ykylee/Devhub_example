# Work Backlog — mvs/work_260607-h-486-ci-runs-api (N-7 / P0-4)

- 문서 목적: 본 sprint 의 작업 항목 진행 상태와 결정 기록을 추적한다.
- 범위: `POST /api/v1/ci-runs` 신규 handler + audit/metric + UT + traceability 정합.
- 대상 독자: 본 sprint 진입자 (Mavis / MiniMax-Code, Claude 인계 전), 후속 리뷰어 (Codex 외부 리뷰 + Claude self-review).
- 상태: in_progress
- 최종 수정일: 2026-06-07
- 관련 문서: [session_handoff.md](./session_handoff.md), [state.json](./state.json), [Issue #486](https://github.com/ykylee/Devhub_example/issues/486), [release_v1_roadmap.md §3.1 P0-4 + §3.5 N-7](../../../../docs/planning/release_v1_roadmap.md).
- 스프린트 목표: v1.0 release 차단 carve (P0-4 / N-7) 해소. Gitea Actions webhook / 직접 생성 경로의 CI Run ingest endpoint.

## 진행 상태

- [x] sprint scaffold 3 파일 초기화 (state.json + work_backlog.md + session_handoff.md)
- [x] 브랜치 mvs/work_260607-h-486-ci-runs-api 생성 (origin/main @ 46568a8 base)
- [x] Issue #486 assign ykylee + milestone v1.0 Release
- [ ] 기존 handler 패턴 탐색 (P1-7 build-runs GET, 다른 idempotent POST 예시)
- [ ] traceability 정합 — report.md / API contract / architecture / requirements 4 문서 동시 갱신
- [ ] internal/httpapi/ci_runs.go 구현
- [ ] internal/httpapi/ci_runs_test.go 작성 (TC-CI-RUN-01..05)
- [ ] route 등록 + audit emit + Prometheus Counter/Histogram 부착
- [ ] go test ./... + go build PASS
- [ ] commit + push + PR 생성

## 결정 기록

- **2026-06-07 브랜치 prefix** — 사용자 redirect. `claude/` 가 아닌 `mvs/` (Mavis/MiniMax-Code 관행) 사용. 본 session 이 Claude 인계 전 단계.
- **2026-06-07 API ID 슬롯** — issue #486 housekeeping comment 의 2차 정정에 따라 `API-98` 사용. `API-94` (Task Item Ingestion) / `API-97` (realtime 도메인 표) 모두 점유.
- **(예정) idempotency key** — issue #486 본문 spec 대로 `(repository_id, commit_sha, status, started_at)` 사용. commit_sha / started_at null 가능 여부 결정 필요 (webhook dedup vs 동일 started_at race).
- **(예정) RBAC scope** — `platform:read` + 동일 org unit subtree scope. 기존 P1-7 build-runs GET handler 의 RBAC 패턴 차용.

## 다음 sprint 진입 후보 (본 sprint 머지 후)

1. **P1-1 sub-carve C** (Keycloak event listener 확장) — Claude sprint-i
2. **N-8 / #488 P1-6 Sign-out endpoint** — Claude sprint-i 동시
3. **N-9 / #487 P1-7 Repository build-runs GET** (이미 main 머지됨, 후속 dashboard widget = Gemini)
4. **N-11 / #419 CI e2e+backend-integration job 복원** — Codex + 사용자
