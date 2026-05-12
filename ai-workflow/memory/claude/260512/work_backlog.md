# 작업 백로그

## [Planned]
- [ ] M2 — `frontend/lib/services/audit.service.ts` consumer UI 통합 (현재 dead code)
- [ ] `dev-up.sh` 보강 — `set -e` + Kratos/Hydra readiness probe + `make migrate-up` 자동 호출 (N1 환경 원인 제거)
- [ ] Windows `dev-up.ps1` 신규 — ASCII 영어, 메모리 정책 준수
- [ ] Kratos identity 조회 long-term — `credentials_identifier` 쿼리 또는 `users.kratos_identity_id` 캐시 우선 경로 강화 (PoC scan 대체)
- [ ] `password-change.spec`, `signout.spec` 회귀 재검증 (N2 머지 후 resolver slow path 회복)
- [ ] 운영 docs 정합 — `docs/setup/test-server-deployment.md` 와 `e2e-test-guide.md` 의 migration prereq 강조 (또는 자동화로 대체)

## [In Progress]
(none — 새 슬롯)

## [Done — 이전 세션, gemini/frontend_260510 위에 머지됨]
- [x] C1-C3 Kratos admin client revert + httptest pin (PR #63)
- [x] H3 seedLocalAdmin 로깅 + M5 postgres.go 빈줄 정리 (PR #64)
- [x] H4 dev-up.sh DB_URL 환경화 + H5 .pids/ gitignore (PR #64)
- [x] M1 console.log + M4 next pin + M3 Modal invariant 코멘트 (PR #64)
- [x] seedLocalAdmin 분기 단위 테스트 + seedOrgStore 인터페이스 (PR #64)
- [x] N1 authenticateActor GetUser 에러 surface (PR #65)
- [x] PostgresStore.GetUser sentinel (`store.ErrNotFound`) 정렬 — organization.go:203 404 경로 부산물 회복 (PR #65)
- [x] N2 Kratos 0-based 페이지네이션 (backend client + e2e globalSetup, PR #66)
- [x] E2E auth.spec 4/4 검증 (login 500 root cause 가 schema/payload 임을 확정)
