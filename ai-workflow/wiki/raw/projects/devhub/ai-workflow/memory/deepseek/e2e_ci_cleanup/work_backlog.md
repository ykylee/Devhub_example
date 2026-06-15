# Work Backlog — deepseek/e2e_ci_cleanup

- 상태: PR 준비 완료

## Done

1. ci-e2e-sync-check.sh 중복 문장형 항목 제거
2. infra-topology.spec.ts 삭제 (무효화된 TC)
3. dashboard-retry-empty-state.spec.ts: describe.skip 해제 + admin test 제거
4. test_cases_m7_onboarding.md: spec 파일명 12개 참조 정정
5. global-setup.ts + onboarding-first-login.spec.ts: 기본 client ID devhub-backend로 통일
6. ci.yml: stale ci-setup.sh 참조 2건 정정
7. test_cases_m4_integration.md: 상태 draft→accepted
8. e2e_testing_strategy.md: stale ci-setup.sh 참조 정정

## Validation
- `git diff --stat`: 10 files changed, 20 insertions(+), 55 deletions(-)
- 변경 내용 확인 완료
