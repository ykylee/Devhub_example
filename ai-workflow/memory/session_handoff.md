# Session Handoff — main (2026-05-11 EOD)

- 문서 목적: main 브랜치 기준 세션 상태와 다음 작업 진입점을 인계한다.
- 범위: M1/M2 종료 상태, TDD foundation 머지, 다음 sprint 후보
- 대상 독자: 후속 에이전트, 프로젝트 리드
- 브랜치: `main` (HEAD `adce7ec`, PR #58 squash 직후)
- 최종 수정일: 2026-05-11
- 상태: M1/M2 100% done. TDD foundation done. 다음 sprint 결정 대기.
- 관련 문서: [통합 로드맵](../../docs/development_roadmap.md), [상태 스냅샷](./state.json), [E2E 가이드](../../docs/setup/e2e-test-guide.md), [배포 가이드](../../docs/setup/test-server-deployment.md)

## 0. 2026-05-11 일과 종료 상태

- main HEAD `adce7ec` (PR #58 squash, TDD foundation).
- 오늘 머지된 PR 11건: #45 #51 #50 #49 #52 #53 #54 #55 #56 #57 #58. 자세한 목록은 `state.json` 의 `merged_prs_2026_05_11`.
- M1/M2 sprint 100% done. M3 partial (사용자/조직 admin) done.
- TDD baseline 정착: Vitest 27 단위 케이스 + Playwright 5 PASS / 1 SKIPPED (`password-change` 는 PR-L4 후속).
- 사용자 환경에서 e2e 한 사이클 처음으로 통과 (`cd frontend && npm run e2e` 한 줄).
- PoC 빠른 진입용 `test/test` 시스템 관리자 시드 추가 (운영 진입 직전 제거 예정, `test-server-deployment.md` §10).

## 1. 다음 세션 진입점 — 우선순위 후보

`state.json` 의 `next_actions` 가 단일 source-of-truth. 사용자 결정에 따라 한 축 선택.

| 후보 | 주요 작업 | 규모 | 우선 사유 |
| --- | --- | --- | --- |
| **PR-L4** Kratos session 정합 | backend `/api/v1/account/password` proxy (session_token 사용) OR browser-mode 로그인 redirect 도입 | M | password-change.spec unskip 차단 해소. M2 hygiene. |
| **PR-T3.5** e2e seed 자동화 | Playwright `globalSetup` 으로 Kratos identity 3건 + DevHub users 3행 시드 자동화 | S | 매 e2e 사이클 수동 시드 부담 해소. |
| **PR-T5** CI 도입 (DEC-4) | GitHub Actions: backend `go test` + frontend Vitest + e2e matrix/nightly | M-L | 회귀 안전망의 마지막 조각. |
| **M4 entry** | command status WebSocket UI, WebSocket 확장 (publish + replay), AI Gardener gRPC, Gitea Hourly Pull | L | 새 기능 트랙. |
| **M2 follow-up hygiene** | users.kratos_identity_id 칼럼 (FindIdentityByUserID O(1)), Kratos webhook → audit_logs, Hydra JWKS verifier 실구현 | M | 운영 진입 전 정합. |
| **PR-D follow-up** | store/postgres.go commands audit INSERT 3 곳 actor context 채우기, 모든 log 라인 request_id 부착, `DEVHUB_TRUSTED_PROXIES` env | S-M | 감사 정합 마무리. |
| **UX hygiene** | `/admin/settings/users` SearchInput 필터, `/account` Kratos privileged session 안내, Header Switch View 한계 안내 | S | UX 보강. |

## 2. 직전 sprint 인계

- [`claude/work_26_05_11-d`](./claude/work_26_05_11-d/session_handoff.md) (CLOSED, PR #58) — TDD foundation. e2e 검증 hole 5건 봉합 (DSN env, migration 5-8, identity schema, video off, seed 헬퍼). 발견된 PR-L3 구조 한계로 password-change 가 skip — PR-L4 의 unblocker.
- [`claude/work_26_05_11-c`](./claude/work_26_05_11-c/session_handoff.md) (CLOSED, PR #57) — M1 PR-D audit actor enrichment.
- [`claude/work_26_05_11-b`](./claude/work_26_05_11-b/session_handoff.md) (CLOSED, PR #56) — M1 envelope + wire/UI split + CommandStatus enum.
- [`claude/work_26_05_11`](./claude/work_26_05_11/backlog/2026-05-11.md) (CLOSED, PR #45/49/50/51/52/53/54/55) — M2 login_action + Track S.

## 3. 환경 / 운영 메모

- **5 프로세스 native 기동**: PostgreSQL(시스템 서비스) + Hydra + Kratos + backend-core + frontend. `test-server-deployment.md` §5.
- **DSN env override 필수**: `infra/idp/{hydra,kratos}.yaml` 의 `dsn` 에 credential 없음. `DSN="postgres://postgres:postgres@localhost:5432/devhub?sslmode=disable&search_path=hydra"` 식 env 로 띄움.
- **migration 적용**: `migrate` CLI 미설치 환경은 `go run ./backend-core/cmd/idp-apply-schemas -sql migrations/...up.sql` 로 1개씩. 진단은 `-query "<SQL>"`.
- **e2e 시드**: `infra/idp/sql/002_seed_e2e_users.sql` (alice/bob/charlie) + Kratos identity 3건 curl (`traits.system_id` 필수). 가이드 `e2e-test-guide.md` §2.
- **빠른 admin 진입**: `infra/idp/sql/003_seed_test_admin.sql` + `test`/`test` Kratos identity. PoC 한정.
- **사내 corp 환경 메모**: `infra/idp/ENVIRONMENT_NOTES.md` (psql 미설치, SSL inspection, PowerShell 5.1 ASCII 강제).

## 4. 잔여 결정 대기 (M0/M1 시점 보존 — 변동 없음)

- 운영 reverse proxy 환경용 `DEVHUB_TRUSTED_PROXIES` env 도입 정책 (현재는 `SetTrustedProxies(nil)` 고정).
- Hydra JWKS / introspection verifier 실구현 (M2 후속).
- PoC OIDC client 의 secrets 운영 진입 시 교체 절차 (`test-server-deployment.md` §10).

## 5. 머지 흐름 요약 (2026-05-11)

```
adce7ec PR #58 — test: TDD foundation (work_26_05_11-d squash)
4e831a3 PR #57 — feat(audit): actor enrichment + request_id middleware (PR-D, M1 final)
9b6b3ea PR #56 — feat(M1 cleanup): envelope + wire/UI split + CommandStatus enum (PR-B+PR-C)
818d54a PR #55 — feat(org): persist drag positions + leader change (PR-S4)
b935876 PR #54 — feat(account): admin endpoints (PR-S3)
2d0075e PR #53 — feat(ui): /admin/settings shell + organization 4 tabs (PR-S2)
a2f707e PR #52 — feat(ui): role-based landing + sidebar gating (PR-S1)
92b3459 PR #49 — docs(setup): test server deployment guide
205980e PR #50 — feat(account): /account password via Kratos settings (PR-L3)
858129f PR #51 — feat(auth): /auth/logout + Kratos browser logout (PR-L2)
61541da PR #45 — feat(auth): /api/v1/auth/logout proxy (PR-L1)
```
