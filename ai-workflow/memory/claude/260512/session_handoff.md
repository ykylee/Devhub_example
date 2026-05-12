# 세션 인계 문서 (2026-05-12)

## 현재 상태
- **브랜치**: `claude/260512` (base `gemini/frontend_260510` @ a59f887)
- **이전 세션 결과**: gemini 의 frontend_260510 회귀 4건 정리, 모두 머지 완료. e2e `auth.spec` 4/4 통과로 login 500 원인이 hotfix 로 해소됨을 확정.

## 머지 완료 PR (gemini/frontend_260510 위)

| PR | 내용 |
| --- | --- |
| #63 | C1-C3: Kratos admin client revert + httptest 16 cases |
| #64 | H3-H5/M1/M3-M5: seedLocalAdmin 로깅, dev-up.sh DB_URL 환경화, .pids/ gitignore, console.log/next pin/Modal invariant |
| #65 | N1: authenticateActor 의 GetUser 비-NotFound 에러 surface + PostgresStore.GetUser sentinel (`store.ErrNotFound`) 정렬 |
| #66 | N2: Kratos /admin/identities 0-based 페이지네이션 (backend + e2e globalSetup) |

## 핵심 컨텍스트
- 로컬 PostgreSQL DSN: `postgres://postgres:postgres@localhost:5432/devhub?sslmode=disable` (사용자 머신 확인됨)
- 마이그레이션 000009 (kratos_identity_id) 는 이 머신에 직접 적용되어 있어야 GetUser 가 작동. dev-up.sh 가 자동 적용 안 함 (follow-up 후보)
- Kratos v26.2.0 (admin port 4434) — 0-based 페이지네이션 검증
- E2E 시드 사용자: alice (developer), bob (manager), charlie (system_admin), 패스워드 `ChangeMe-12345!`

## 다음 작업 후보 (work_backlog 참조)
1. **M2** — `frontend/lib/services/audit.service.ts` 의 consumer (admin 감사로그 화면) 통합
2. **dev-up.sh 보강** — `set -e`, Kratos/Hydra readiness probe, `make migrate-up` 자동 호출
3. **Windows dev-up.ps1** — ASCII 영어, 메모리 정책 준수
4. **Kratos identity O(1) lookup** — page scan 대신 `credentials_identifier` 쿼리 또는 캐시 우선
5. **password-change.spec / signout.spec 회귀 재검증** — N2 머지 후 slow path 가 회복됐는지 확인

## 작업 원칙 리마인더 (CLAUDE.md)
- 사용자 보고는 한국어
- bounded 작업은 sub-agent 위임 가능 (`workflow-doc-worker`, `workflow-code-worker`, `workflow-validation-worker`)
- 파괴적 작업은 사용자 확인 후
