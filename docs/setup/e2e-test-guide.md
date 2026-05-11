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

### Kratos identity 시드 (3건)

각 사용자에 대해 다음 호출 1번씩:

```sh
curl -X POST http://localhost:4434/admin/identities \
  -H "Content-Type: application/json" \
  -d '{
    "schema_id": "devhub_user",
    "traits": { "email": "alice@example.com", "display_name": "Alice" },
    "metadata_public": { "user_id": "alice" },
    "credentials": {
      "password": { "config": { "password": "ChangeMe-12345!" } }
    }
  }'
```
(`bob` / `charlie` 도 동일 패턴, role 만 다름)

### DevHub users 시드

```sh
psql -U devhub -d devhub -c "
INSERT INTO users (user_id, email, display_name, role, status, type)
VALUES
  ('alice',   'alice@example.com',   'Alice',   'developer',    'active', 'human'),
  ('bob',     'bob@example.com',     'Bob',     'manager',      'active', 'human'),
  ('charlie', 'charlie@example.com', 'Charlie', 'system_admin', 'active', 'human')
ON CONFLICT (user_id) DO NOTHING;
"
```

시드는 idempotent — 이미 존재하는 identity/user 는 무시. e2e 가 password-change 시나리오의 cleanup 단계에서 원래 비밀번호로 복귀시키므로 재실행에도 안전.

## 3. Playwright 설치 (1회)

```sh
cd frontend
npm ci  # devDependencies 에 @playwright/test 가 들어있음
npx playwright install --with-deps chromium
```

`--with-deps` 는 Linux 의 시스템 라이브러리도 함께 설치. Windows/macOS 에서는 `--with-deps` 생략 가능.

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

## 5. 시나리오 (현재 7건)

| 파일 | 시나리오 | 목적 |
| --- | --- | --- |
| `auth.spec.ts` | developer 로그인 → `/developer` | PR-S1 role-based landing |
| `auth.spec.ts` | manager 로그인 → `/manager` | PR-S1 role-based landing |
| `auth.spec.ts` | system_admin 로그인 → `/admin` | PR-S1 role-based landing |
| `auth.spec.ts` | developer 가 `/admin/settings` 직진입 → `/developer` | AuthGuard `pathRequiresSystemAdmin` |
| `signout.spec.ts` | Sign Out 후 `/login` 진입 시 password 재요청 | PR-L2 Hydra session 종료 |
| `password-change.spec.ts` | 비밀번호 변경 → Sign Out → 새 비밀번호 재로그인 → 원복 | PR-L3 Kratos settings flow |

## 6. 트러블슈팅

| 증상 | 원인 | 조치 |
| --- | --- | --- |
| `loginAs` 가 `/auth/login?login_challenge=...` 까지 못 감 | Hydra `urls.login` 이 frontend host 와 다름 | `infra/idp/hydra.yaml` 의 `urls.login` 정정 후 Hydra 재기동 |
| 로그인 폼에서 401 (invalid credentials) | Kratos identity 시드 password 가 일치 안 함 | §2 의 시드 비밀번호 재확인. password 변경 시나리오 중단 시 cleanup 실패 가능 — `kratos admin identity` 로 password 재설정 |
| `/account` 비밀번호 변경 시 "Re-authentication required" | Kratos `privileged_session_max_age=15m` 초과 | 시나리오 시간 < 15m. 환경 시간 동기화 확인. 그래도 발생 시 fixture 가 재로그인 retry 로직 추가 필요 (후속 hygiene) |
| `Sign Out` 후에도 `/login` 이 silent re-auth | Hydra session 종료 안 됨. id_token_hint 누락 가능성 | tokenStore 의 id_token 저장 확인 (PR-L2 fix-up). `/oauth2/sessions/logout` 호출 URL 확인 |
| 사용자 환경 Chromium 다운로드 실패 | 사내 SSL inspection / 외부 미러 차단 | `PLAYWRIGHT_BROWSERS_PATH` 또는 사내 미러 사용. `npx playwright install --dry-run` 으로 다운로드 URL 확인 |

## 7. 향후 확장 (PR-T4 범위)

- 조직 관리 e2e — `/admin/settings/organization` 부서 추가/이동/삭제 + 차트 drag 좌표 영속화
- 사용자 관리 e2e — 계정 발급/리셋/disable 흐름 (PR-S3)
- 권한 매트릭스 e2e — PermissionEditor 정책 변경 + audit 확인
- 시드 자동화 — pre-test hook 으로 Kratos admin API 자동 시드 (현재는 수동)

## 8. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-11 | 초판 작성 (PR-T3) |
