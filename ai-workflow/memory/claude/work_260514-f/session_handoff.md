# Session Handoff — claude/work_260514-f (hotfix)

- 브랜치: `claude/work_260514-f`
- Base: `main` @ `1e38c4d` (PR #109 머지 직후)
- 날짜: 2026-05-14
- 상태: in_progress
- Trigger: PR #109 머지 후 codex 외부 리뷰 (P1 inline) + 본인 리뷰어 종합 리뷰가 발견한 핵심 + 보강 항목.

## Scope (B1 + I1 + I3)

### B1 (codex P1 = blocking)
`applicationsFixture` 의 cleanup SQL 이 multi-statement + bind args 라 pgx prepared execution 이 거부 → 23 test 가 fixture 단계에서 panic. 정정: TRUNCATE 와 DELETE 를 별도 `pool.Exec` 로 분리.

### I1 (handoff/문서 보강)
test DB 마이그레이션 적용 절차 명시 + `scripts/setup-test-db.sh` placeholder.

### I3 (CI backend-integration job 신설)
`.github/workflows/ci.yml` 에 새 job 추가:
- e2e job 의 PG 15 setup 패턴 재사용 (Setup PostgreSQL native)
- `scripts/ci-setup.sh` 로 마이그레이션 적용
- `cd backend-core && go test -count=1 -run 'TestIntegration_' ./internal/store/...` 실행
- DEVHUB_TEST_DB_URL 설정

CI 잡 수: 4 → 5 (Workflow Lint / Backend Unit / Frontend Unit / E2E / **Backend Integration**)

## 검증 흐름

- B1 정정 후: 23 test 가 fixture 통과
- CI backend-integration job 이 본 sprint 부터 23 test 실행 → P1/P2 회귀 guard 가 실제 검증

## 다음 세션 우선 작업

- (본 sprint 종료 시 갱신)
