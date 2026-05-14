# Work Backlog — claude/work_260514-f (hotfix)

- 상태 관리: planned / in_progress / blocked / done

## A. B1 fixture cleanup SQL 분리

- [planned] applications_integration_test.go::applicationsFixture 의 cleanup 을 cleanupStatic + cleanupRepos 2 statement 로 분리
- [planned] TestIntegration_FixtureCleanupSanity 신규 — fixture 가 정상 동작하는지 확인 + repos seed 검증

## B. I1 test DB 마이그레이션 절차

- [planned] scripts/setup-test-db.sh 신규 — DEVHUB_TEST_DB_URL 환경 가정 + migrate up 호출
- [planned] docs/setup/test-server-deployment.md 또는 별도 section 에 절차 명시
- [planned] handoff 에 절차 링크

## C. I3 CI backend-integration job

- [planned] .github/workflows/ci.yml 에 backend-integration job 신설
- [planned] PG 15 setup (e2e job 의 step 재사용 — 복사)
- [planned] scripts/ci-setup.sh 일부 재사용 (마이그레이션만)
- [planned] DEVHUB_TEST_DB_URL 환경에서 go test -run 'TestIntegration_' 실행

## D. 매트릭스 + 머지

- [planned] trace.md §6 변경 이력
- [planned] commit + push + CI (5/5 SUCCESS — backend-integration 신규) + squash merge
