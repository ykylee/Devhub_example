# Session Handoff — claude/work_260514-e

- 브랜치: `claude/work_260514-e`
- Base: `main` @ `7822a91` (PR #108 머지 직후)
- 날짜: 2026-05-14
- 상태: in_progress
- 직전 sprint: `claude/work_260514-d` (PR #108 hotfix) — P1/P2 정정.

## Sprint scope

**postgres integration test 도입** — work_260514-c carve_out 흡수.

대상:
- `internal/store/applications.go` (16 메서드)
- `internal/store/repository_ops.go` (5 메서드 — Repository ops 4 + rollup compute + critical count)
- `internal/store/integrations.go` (5 메서드 — CRUD)

CI 거동:
- `backend-unit` job 은 `make test` 만 — `DEVHUB_TEST_DB_URL` 미설정으로 `t.Skip`. 기존 postgres_integration_test.go / users_units_test.go 와 일관.
- 로컬 / 후속 CI 잡에서 DB 환경 갖춰진 상태로 실 실행.

## 작업 순서

1. fixture helper (applications/projects/integrations/scm_providers/repositories cleanup + seed)
2. Applications store + Link + SCM provider integration test
3. Repository ops + rollup compute integration test (P1 회귀 guard 핵심)
4. Projects + Integrations integration test (P2 회귀 guard)
5. trace.md §3 row 의 UT 컬럼에 integration test 발급
6. commit + push + PR + CI (skip 만 검증) + squash merge

## 위험

- 로컬 검증 — 사용자가 DEVHUB_TEST_DB_URL 설정 후 1회 검증 권장. 본 sprint 는 코드 작성 + CI skip 확인 후 머지.
- P1 회귀 guard 는 수치 정확성 검증 (custom_weights fallback 후 sum=1.0 invariant) — 본 sprint 의 핵심.

## 다음 세션 우선 작업
- (본 sprint 종료 시 갱신)
