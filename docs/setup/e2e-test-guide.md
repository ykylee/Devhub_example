# E2E Test Guide (Playwright, Keycloak/OIDC)

> ⚠ 2026-05-20 정정: 본 문서는 Keycloak-only 운영 기준으로 갱신되었다.

- 문서 목적: DevHub Example 의 Playwright e2e 스위트를 사용자 환경에서 실행하기 위한 사전 조건과 절차를 정의한다.
- 범위: 사전 조건, 시드 데이터, 실행 명령, 시나리오 목록, 트러블슈팅
- 대상 독자: 본인 환경에서 회귀 검증을 돌리는 개발자, QA
- 상태: active
- 최종 수정일: 2026-05-20
- 관련 문서: [테스트 서버 배포 가이드](./test-server-deployment.md), [Playwright config](../../frontend/playwright.config.ts), [e2e fixtures](../../frontend/tests/e2e/fixtures.ts)

## 0. 정책

- **DEC-3=A**: e2e 는 mock IdP 가 아니라 실 Keycloak/OIDC 환경에서 실행한다. 운영 흐름과 동일한 OIDC 코드 흐름을 검증.
- **Single worker**: IdP session state 충돌 방지를 위해 E2E 테스트는 1 worker 유지.
- **사용자 native 또는 compose**: PostgreSQL + Keycloak + backend-core + frontend 가 먼저 기동되어야 한다. Playwright 의 `webServer` 옵션은 비활성.

## 1. 사전 조건

본 가이드는 [`test-server-deployment.md`](./test-server-deployment.md) 의 §1-§5 가 이미 끝난 상태에서 시작한다. 즉:

- PostgreSQL `devhub` DB 마이그레이션 완료
- Keycloak/OIDC 가 가동 중 (`/devhub/auth/keycloak` 경로 또는 issuer URL 접근 가능)
- backend-core + frontend 가동 중
- OIDC client `devhub-frontend` 가 IdP(Keycloak)에 등록 완료

검증:
```sh
curl http://localhost:8080/health
curl http://localhost:8180/devhub/auth/keycloak/realms/devhub/.well-known/openid-configuration
curl -I http://localhost:3000/
```
모두 200/OK 이면 다음 단계.

## 2. 시드 데이터

e2e 픽스처 (`frontend/tests/e2e/fixtures.ts` 의 `SEEDED`) 가 가정하는 3 사용자:

| user_id | email | password | role | landing |
| --- | --- | --- | --- | --- |
| alice | alice@example.com | ChangeMe-12345! | developer | /developer |
| bob | bob@example.com | ChangeMe-12345! | manager | /manager |
| charlie | charlie@example.com | ChangeMe-12345! | system_admin | /admin |

### 2.0 자동 시드 (기본, PR-T3.5)

Playwright `globalSetup` (`frontend/tests/e2e/global-setup.ts`) 이 매 `npm run e2e` 실행 직전에 위 표를 idempotent 하게 시드한다. 필요한 환경변수:

| 변수 | 의미 | 기본값 |
| --- | --- | --- |
| `DEVHUB_KEYCLOAK_ADMIN_URL` | Keycloak admin base URL (seed API 호출) | `http://localhost:8180/devhub/auth/keycloak` |
| `DEVHUB_KEYCLOAK_ADMIN_REALM` | realm 명 | `devhub` |
| `DEVHUB_KEYCLOAK_ADMIN_CLIENT_ID` | e2e seed 전용 service account client id | `devhub-e2e-seeder` |
| `DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET` | service account client secret | (필수) |
| `DSN` | DevHub users 행을 INSERT 할 PostgreSQL DSN. `idp-apply-schemas` 헬퍼가 사용 | (필수) |
| `DEVHUB_E2E_SKIP_SEED` | `1` 이면 시드 단계를 건너뜀 (CI matrix 가 별도 stage 에서 시드할 때) | (미설정) |

실행 예:

```powershell
$env:DSN = "postgres://postgres:postgres@localhost:5432/devhub?sslmode=disable"
cd frontend
npm run e2e
```

자동 시드는 다음 동작을 한다:

1. Keycloak Admin API 기준으로 email 기준 사용자 존재 여부 확인
   - 누락된 identity → Keycloak 사용자 생성
   - 기존 identity → 테스트 비밀번호/속성 재시드
2. backend-core 의 `cmd/idp-apply-schemas -sql infra/idp/sql/002_seed_e2e_users.sql` 호출 (ON CONFLICT DO NOTHING).

두 번째 실행부터는 DB 측은 no-op, Keycloak 측은 동일 값 재적용이라 사실상 no-op.

`scripts/setup-keycloak.sh` 실행 시 아래 출력이 함께 제공된다:
- `DEVHUB_E2E_KEYCLOAK_ADMIN_CLIENT_ID=devhub-e2e-seeder`
- `DEVHUB_E2E_KEYCLOAK_ADMIN_CLIENT_SECRET=<secret>`

### 2.1 수동 시드 (fallback) — Keycloak identity (3건)

수동 시드는 [keycloak_operations.md](./keycloak_operations.md)의 사용자 생성 절차를 따른다.

### 2.2 수동 시드 (fallback) — DevHub users

저장소에 idempotent 한 시드 SQL (`infra/idp/sql/002_seed_e2e_users.sql`) 이 포함되어 있다. `psql` 이 PATH 에 있으면:

```sh
psql -U postgres -d devhub -f infra/idp/sql/002_seed_e2e_users.sql
```

`psql` 미설치 환경 (사내 Windows 등) 에서는 backend-core 의 헬퍼를 사용:

```powershell
cd backend-core
go run ./cmd/idp-apply-schemas -sql ../infra/idp/sql/002_seed_e2e_users.sql
```

헬퍼는 backend-core 의 pgx/v5 의존성을 재사용한다 (`infra/idp/ENVIRONMENT_NOTES.md` §2.2). `-query "<SELECT ...>"` 플래그로 임의 SELECT 결과도 출력 가능 — 디버깅용.

시드는 idempotent — 이미 존재하는 identity/user 는 무시. e2e 가 password-change 시나리오의 cleanup 단계에서 원래 비밀번호로 복귀시키므로 재실행에도 안전.

## 3. Playwright 설치 (1회)

```sh
cd frontend
npm ci  # devDependencies 에 @playwright/test 가 들어있음
```

본 sprint 의 `playwright.config.ts` 는 chromium project 에 `channel: "chrome"` 을 지정해 **시스템에 이미 설치된 Chrome 을 재사용**한다. 따라서 별도 `npx playwright install` 단계가 필요 없다. Windows/macOS 에서는 보통 Chrome 이 기본 설치되어 있고, Linux 는 패키지 매니저 (`apt install google-chrome-stable` 등) 로 설치한다.

추가로 `video` 캡처는 `off`. Playwright 의 video 녹화는 bundled ffmpeg 바이너리를 요구하는데, 이 역시 `npx playwright install` 로만 받을 수 있다. 시스템 Chrome 재사용 정책과 일관성을 위해 비활성. 실패 시 진단은 trace (zip) + screenshot 으로 충분.

### 대안 — bundled Chromium 사용 (사내 SSL inspection 환경 등)

`channel: "chrome"` 을 빼고 Playwright 의 bundled Chromium 으로 가려면:

- **사내 CA 신뢰**: `NODE_EXTRA_CA_CERTS=/path/to/corp-ca.crt npx playwright install chromium` (가장 안전, CA 인증서는 사내 IT 가이드 참조)
- **Playwright mirror**: `PLAYWRIGHT_DOWNLOAD_HOST=https://mirror.your-corp.local/playwright npx playwright install chromium`
- **TLS 검증 비활성** (임시 — 보안 약함): `NODE_TLS_REJECT_UNAUTHORIZED=0 npx playwright install chromium`

bundled 로 전환 시 `playwright.config.ts` 에서 `channel: "chrome"` 줄을 제거.

## 4. 실행

```sh
cd frontend
npm run e2e            # CI mode: 전체 실행 + HTML report
npm run e2e:ui         # 인터랙티브 UI 모드 (시나리오 선택 + step 별 inspect)
npm run e2e:report     # 직전 실행의 HTML report 열기
```

기본 base URL 은 `http://localhost:3000`. 다른 host 사용 시:
```sh
PLAYWRIGHT_BASE_URL=http://10.0.0.5:3000 npm run e2e
```

## 5. 시나리오 (현재 6건)

| 파일 | 시나리오 | 목적 |
| --- | --- | --- |
| `auth.spec.ts` | developer 로그인 → `/developer` | PR-S1 role-based landing |
| `auth.spec.ts` | manager 로그인 → `/manager` | PR-S1 role-based landing |
| `auth.spec.ts` | system_admin 로그인 → `/admin` | PR-S1 role-based landing |
| `auth.spec.ts` | developer 가 `/admin/settings` 직진입 → `/developer` | AuthGuard `pathRequiresSystemAdmin` |
| `signout.spec.ts` | Sign Out 후 `/login` 진입 시 password 재요청 | PR-L2 IdP session 종료 |
| `password-change.spec.ts` | 비밀번호 변경 → Sign Out → 재로그인 → 원복 | PR-L4 `POST /api/v1/account/password` backend proxy. 자동 시드와 함께 활성 (PR-T3.5) |

## 6. 트러블슈팅

| 증상 | 원인 | 조치 |
| --- | --- | --- |
| `loginAs` 가 `/login` 까지 못 감 | IdP redirect/callback URI 또는 frontend host 불일치 | IdP client 설정의 redirect URI 및 frontend origin 재확인 |
| 로그인 폼에서 401 (invalid credentials) | Keycloak 사용자 시드 password 불일치 | `npm run e2e` 재실행 후 globalSetup 시드 단계 확인 |
| 로그인 성공 직후 `/api/v1/me` 401 + `/login?error=session_expired` 반복 | backend verifier issuer/JWKS 경로 불일치 (`token has invalid issuer` 또는 JWKS fetch 실패) | `DEVHUB_OIDC_ISSUER_URL` 을 browser 토큰 `iss` 와 동일한 public URL 로 맞추고, `DEVHUB_OIDC_JWKS_URL` 은 backend 컨테이너에서 도달 가능한 주소(`host.docker.internal` 등)로 명시 |
| `/account` 관련 시나리오 실패 | Keycloak Account Console/redirect 설정 불일치 | `NEXT_PUBLIC_OIDC_REDIRECT_URI`, issuer, Keycloak client redirect URI 정합 확인 |
| `Sign Out` 후에도 `/login` 이 silent re-auth | IdP session 종료 안 됨. id_token_hint 누락 가능성 | tokenStore 의 `id_token` 저장 여부와 end-session endpoint 호출 URL 확인 |
| 사용자 환경 Chromium 다운로드 실패 | 사내 SSL inspection / 외부 미러 차단 | `PLAYWRIGHT_BROWSERS_PATH` 또는 사내 미러 사용. `npx playwright install --dry-run` 으로 다운로드 URL 확인 |

## 7. 향후 확장 (PR-T4 범위)

- 조직 관리 e2e — `/admin/settings/organization` 부서 추가/이동/삭제 + 차트 drag 좌표 영속화
- 사용자 관리 e2e — 계정 발급/리셋/disable 흐름 (PR-S3)
- 권한 매트릭스 e2e — PermissionEditor 정책 변경 + audit 확인
- 시드 자동화 고도화 — Keycloak admin rate limit/재시도 정책 강화

## 8. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-11 | 초판 작성 (PR-T3) |
| 2026-05-11 | PR-L4 `POST /api/v1/account/password` backend proxy 도입에 따라 password-change 시나리오 사전 조건/트러블슈팅 갱신 (work_26_05_11-e) |
| 2026-05-11 | PR-T3.5 Playwright globalSetup 자동 시드 도입 + password-change.spec unskip (work_26_05_11-e) |
| 2026-05-12 | PR-T3.5 hardening — globalSetup 이 기존 identity 의 비밀번호를 PUT 으로 force-reset, stale rotation 자동 복구 (work_260512-f) |
| 2026-05-20 | Keycloak-only 운영 기준으로 사전조건/시드/트러블슈팅 전면 정정 (Hydra/Kratos 절차 제거) |
