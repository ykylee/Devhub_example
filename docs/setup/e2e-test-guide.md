# E2E Test Guide (Playwright, native Hydra/Kratos)

- 문서 목적: DevHub Example 의 Playwright e2e 스위트를 사용자 환경에서 실행하기 위한 사전 조건과 절차를 정의한다.
- 범위: 사전 조건, 시드 데이터, 실행 명령, 시나리오 목록, 트러블슈팅
- 대상 독자: 본인 환경에서 회귀 검증을 돌리는 개발자, QA
- 상태: draft (PR-T3, work_26_05_11-d sprint)
- 최종 수정일: 2026-05-11
- 관련 문서: [테스트 서버 배포 가이드](./test-server-deployment.md), [Playwright config](../../frontend/playwright.config.ts), [e2e fixtures](../../frontend/tests/e2e/fixtures.ts)

## 0. 정책

- **DEC-3=A**: e2e 는 mock IdP 가 아니라 실 Hydra/Kratos 환경에서 실행한다. 운영 흐름과 동일한 OIDC 코드 흐름을 검증.
- **Single worker**: Kratos session 이 브라우저 context 별이라 시나리오 간 cross-contamination 방지 위해 1 worker.
- **사용자 native**: 5개 프로세스 (PostgreSQL + Hydra + Kratos + backend-core + frontend) 를 사용자가 직접 기동. Playwright 의 `webServer` 옵션은 의도적으로 비활성.

## 1. 사전 조건

본 가이드는 [`test-server-deployment.md`](./test-server-deployment.md) 의 §1-§5 가 이미 끝난 상태에서 시작한다. 즉:

- PostgreSQL `devhub` DB + `hydra` / `kratos` schema 가 마이그레이션 완료
- Hydra/Kratos 가 native binary 로 가동 중 (포트 4444/4445/4433/4434)
- backend-core (8080) + frontend (3000) 가동 중
- OIDC client `devhub-frontend` 가 Hydra 에 등록 완료 (`infra/idp/scripts/register-devhub-client.ps1`)

검증:
```sh
curl http://localhost:8080/health
curl http://localhost:4444/health/ready
curl http://localhost:4433/health/ready
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
| `KRATOS_ADMIN_URL` | Kratos admin endpoint (identity 생성) | `http://localhost:4434` |
| `DSN` | DevHub users 행을 INSERT 할 PostgreSQL DSN. `idp-apply-schemas` 헬퍼가 사용 | (필수) |
| `DEVHUB_E2E_SKIP_SEED` | `1` 이면 시드 단계를 건너뜀 (CI matrix 가 별도 stage 에서 시드할 때) | (미설정) |

실행 예:

```powershell
$env:DSN = "postgres://postgres:postgres@localhost:5432/devhub?sslmode=disable"
cd frontend
npm run e2e
```

자동 시드는 다음 동작을 한다:

1. Kratos admin `/admin/identities` 페이지 스캔으로 email 기준 존재 여부 확인 → 누락된 identity 만 POST
2. backend-core 의 `cmd/idp-apply-schemas -sql infra/idp/sql/002_seed_e2e_users.sql` 호출 (ON CONFLICT DO NOTHING)

따라서 두 번째 실행부터는 사실상 no-op. 이미 비밀번호가 회전된 상태 (예: password-change 시나리오 중단) 라면 자동 시드는 **변경하지 않는다** — §2.2 의 수동 절차로 원복 후 재실행.

### 2.1 수동 시드 (fallback) — Kratos identity (3건)

각 사용자에 대해 다음 호출 1번씩:

```sh
curl -X POST http://localhost:4434/admin/identities \
  -H "Content-Type: application/json" \
  -d '{
    "schema_id": "devhub_user",
    "traits": {
      "system_id": "alice",
      "email": "alice@example.com",
      "display_name": "Alice"
    },
    "metadata_public": { "user_id": "alice" },
    "credentials": {
      "password": { "config": { "password": "ChangeMe-12345!" } }
    }
  }'
```
(`bob` / `charlie` 도 동일 패턴, role 만 다름)

`traits.system_id` 가 `identity.schema.json` 의 password identifier 다 — 로그인 폼의 "System ID" 입력값이 이 값과 매칭된다. 누락하면 Kratos 가 400 `missing properties: "system_id"` 로 거절한다.

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
| `signout.spec.ts` | Sign Out 후 `/login` 진입 시 password 재요청 | PR-L2 Hydra session 종료 |
| `password-change.spec.ts` | 비밀번호 변경 → Sign Out → 재로그인 → 원복 | PR-L4 `POST /api/v1/account/password` backend proxy. 자동 시드와 함께 활성 (PR-T3.5) |

## 6. 트러블슈팅

| 증상 | 원인 | 조치 |
| --- | --- | --- |
| `loginAs` 가 `/auth/login?login_challenge=...` 까지 못 감 | Hydra `urls.login` 이 frontend host 와 다름 | `infra/idp/hydra.yaml` 의 `urls.login` 정정 후 Hydra 재기동 |
| 로그인 폼에서 401 (invalid credentials) | Kratos identity 시드 password 가 일치 안 함 | §2 의 시드 비밀번호 재확인. password 변경 시나리오 중단 시 cleanup 실패 가능 — `kratos admin identity` 로 password 재설정 |
| `/account` 비밀번호 변경 시 "Re-authentication required" | Kratos `privileged_session_max_age=15m` 초과 | PR-L4 backend proxy 가 매 호출마다 fresh api-mode 로그인을 돌려 privileged window 를 갱신하므로 정상 시나리오에서는 발생하지 않음. 그래도 노출되면 backend 의 `DEVHUB_KRATOS_PUBLIC_URL` env 누락/오설정 가능성 |
| `/account` 비밀번호 변경 시 "current password is incorrect" | 입력한 current_password 가 Kratos 시드와 불일치 | §2 의 시드 비밀번호 확인 (`ChangeMe-12345!`). password-change 시나리오가 중간에 실패해 회전이 진행된 상태라면 `kratos admin identity` 로 패스워드 원복 |
| `Sign Out` 후에도 `/login` 이 silent re-auth | Hydra session 종료 안 됨. id_token_hint 누락 가능성 | tokenStore 의 id_token 저장 확인 (PR-L2 fix-up). `/oauth2/sessions/logout` 호출 URL 확인 |
| 사용자 환경 Chromium 다운로드 실패 | 사내 SSL inspection / 외부 미러 차단 | `PLAYWRIGHT_BROWSERS_PATH` 또는 사내 미러 사용. `npx playwright install --dry-run` 으로 다운로드 URL 확인 |

## 7. 향후 확장 (PR-T4 범위)

- 조직 관리 e2e — `/admin/settings/organization` 부서 추가/이동/삭제 + 차트 drag 좌표 영속화
- 사용자 관리 e2e — 계정 발급/리셋/disable 흐름 (PR-S3)
- 권한 매트릭스 e2e — PermissionEditor 정책 변경 + audit 확인
- ~~시드 자동화 — pre-test hook 으로 Kratos admin API 자동 시드 (현재는 수동)~~ — PR-T3.5 에서 globalSetup 으로 처리

## 8. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-11 | 초판 작성 (PR-T3) |
| 2026-05-11 | PR-L4 `POST /api/v1/account/password` backend proxy 도입에 따라 password-change 시나리오 사전 조건/트러블슈팅 갱신 (work_26_05_11-e) |
| 2026-05-11 | PR-T3.5 Playwright globalSetup 자동 시드 도입 + password-change.spec unskip (work_26_05_11-e) |
