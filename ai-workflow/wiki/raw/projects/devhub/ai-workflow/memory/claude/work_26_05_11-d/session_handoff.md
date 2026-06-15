# Session Handoff — claude/work_26_05_11-d (TDD foundation) — CLOSED

- 문서 목적: `claude/work_26_05_11-d` sprint 종료 기록
- 범위: TDD 기반 마련 — Frontend 단위 테스트 인프라 (Vitest) + E2E (Playwright) 인프라 + 첫 시나리오 세트
- 대상 독자: 후속 에이전트, 프로젝트 리드
- 상태: closed
- 머지: PR #58 squash → main `adce7ec` (2026-05-11)
- 최종 수정일: 2026-05-11

## 0. Outcome

- PR-T1/T2/T3 단일 PR (#58) 로 squash. 6 follow-up commit 으로 e2e 사용자 검증 회귀 hole 모두 봉합.
- E2E baseline: **5 PASS / 1 SKIPPED / 11.5s** (Playwright list reporter 직접 실행 확인).
- Vitest unit suite: 27 케이스 (lib/auth/ statement 95.04% / function 100% / line 98.91%).
- 사용자 환경 (Windows + corp SSL inspection) 에서 한 사이클 처음으로 e2e 가 통과한 sprint — 사전 조건 (DSN env override, migration 5-8, Kratos identity schema) 의 차이가 모두 가이드/시드 헬퍼로 흡수됨.

## 1. 결정 4건 (확정 2026-05-11)

- **DEC-1=Vitest** / **DEC-2=Playwright** / **DEC-3=A 사용자 native E2E** / **DEC-4=B 별도 sprint CI**.

## 2. 시나리오 결과

| 시나리오 | 결과 | 비고 |
| --- | --- | --- |
| developer 로그인 → /developer | PASS | 3.7s |
| manager 로그인 → /manager | PASS | 1.1s |
| system_admin 로그인 → /admin | PASS | 3.0s |
| AuthGuard /admin/settings bounce | PASS | 1.0s |
| Sign Out → 재로그인 password 요청 | PASS | 1.5s |
| /account 비밀번호 변경 round-trip | **SKIPPED** | `test.skip` — PR-L4 후속. 사유: login 이 api-mode 라 `ory_kratos_session` cookie 없음 → settings/browser flow 거절 |

## 3. 진단 중 발견된 hole (모두 봉합)

- `infra/idp/{hydra,kratos}.yaml` 의 `dsn` 에 credential 누락 → `DSN` env override 안내 추가 (`test-server-deployment.md` §5)
- `migrate` CLI 미설치 환경에서 migration 5-8 (`rbac_policies`/`users.role` FK/`users.user_type`/`audit_logs` actor) 미적용 → `idp-apply-schemas` 헬퍼로 적용 가능, `-query` 플래그 신설로 진단도 가능
- 가이드의 Kratos identity 페이로드가 `identity.schema.json` 의 `traits.system_id` required 와 불일치 → 두 가이드 모두 정정
- Playwright `channel: "chrome"` 환경에서 video 캡처가 bundled ffmpeg 를 요구 → `video: "off"` (trace+screenshot 으로 충분)
- DevHub users seed 가 매번 수동 → idempotent `infra/idp/sql/002_seed_e2e_users.sql` + 헬퍼 호출 가이드

## 4. 후속 작업 (별도 sprint 후보)

- **PR-L4** (Kratos session 정합): backend `/api/v1/account/password` proxy (session_token 사용) OR browser-mode 로그인 redirect 도입. unskip 조건 = password-change.spec.ts.
- **PR-T3.5** (e2e seed 자동화): Playwright `globalSetup` 으로 Kratos identity 3건 + DevHub users 3행 시드 자동화 (현재 가이드 수동 절차).
- **PR-T5** (CI 도입, DEC-4): GitHub Actions 에서 backend test + frontend Vitest + e2e (matrix 또는 nightly).
- **운영 진입 hygiene**: `test/test` PoC 시드 제거 (`test-server-deployment.md` §10 체크리스트 신설).

## 5. 머지 후 main 영향

- main HEAD `adce7ec` (squash). 코드 변경: `backend-core/cmd/idp-apply-schemas`, `frontend/{vitest.config,playwright.config,lib/test-setup,...}`, `infra/idp/sql/00{2,3}_*.sql`, docs 2건.
- 직접적인 런타임 영향 없음 (테스트 인프라 + 시드 SQL + 가이드). 기존 backend/frontend 동작 변화 없음.

## 6. 다음 세션 진입 포인트

- main `state.json` / `session_handoff.md` 갱신본 → 후속 sprint 우선순위 결정.
- 권장 다음 작업 — 사용자 결정에 따라 PR-L4 (Kratos hygiene) 또는 M2 entry (시스템 설정 - 사용자/조직 관리) 중 하나.
