# 테스트 서버 배포 가이드 (Native, no-docker)

- 문서 목적: DevHub Example 을 단일 테스트 서버에 native binary 로 빌드·배포·기동하는 표준 절차를 정의한다. 5개 컴포넌트 (PostgreSQL, Hydra, Kratos, backend-core, frontend) 를 Docker 없이 직접 실행한다.
- 범위: 사전 준비, 빌드(`go build` + `npm run build`), 환경변수, 기동 순서, OIDC client 등록, 시드 사용자, 헬스체크, 로그인 e2e, 재기동/롤백.
- 대상 독자: 테스트 서버 운영자, QA, 신규 환경 부트스트랩 담당.
- 상태: draft (sprint `claude/work_26_05_11`)
- 최종 수정일: 2026-05-11
- 관련 문서: [개발 환경 구성](./environment-setup.md) (개발용), [ADR-0001 IdP](../adr/0001-idp-selection.md), [Hydra 설정](../../infra/idp/hydra.yaml), [Kratos 설정](../../infra/idp/kratos.yaml), [backend API 계약](../backend_api_contract.md)

## 0. 정책

- **Docker 미사용** — 모든 프로세스는 host OS 의 native binary 또는 시스템 서비스로 기동한다 (`CLAUDE.md` 의 "No Docker policy", ADR-0001 §5·§7.2 결정).
- **PoC 1차** — Windows Service / systemd unit 등록은 본 문서 범위 밖. 별도 창/백그라운드 프로세스로 직접 실행 (ADR-0001 §8.7).
- **PoC secrets** — `infra/idp/{hydra,kratos}.yaml` 의 `secrets` 필드는 PoC placeholder. 운영 진입 시 반드시 교체 (ADR-0001 §1·§7.2 의 후속 hygiene).

## 1. 사전 준비

### 1.1 호스트 런타임

| 항목 | 권장 버전 | 확인 명령 |
| --- | --- | --- |
| Go | 1.22+ | `go version` |
| Node.js | 20 LTS | `node --version` |
| npm | 10+ | `npm --version` |
| PostgreSQL | 15 | `psql --version` |
| Ory Hydra | v2.x | `hydra version` |
| Ory Kratos | v1.x | `kratos version` |
| `migrate` CLI | v4.19.1 | `migrate -version` |

### 1.2 Hydra/Kratos binary 설치 (필요 시)

ADR-0001 §8.6: 사용자 터미널에서 직접 실행. 사내 GoProxy mirror 가 잡혀있는 환경에서:

```sh
go install github.com/ory/hydra/v2/cmd/hydra@latest
go install github.com/ory/kratos/cmd/kratos@latest
```

또는 PowerShell 환경에서 `infra/idp/scripts/install-binaries.ps1` 참조.

### 1.3 PostgreSQL 데이터베이스/스키마

DevHub 본 데이터베이스 1개 + Hydra/Kratos 용 schema 2개를 같은 인스턴스에 둔다 (ADR-0001 §8.1 결정).

```sh
createdb -U postgres devhub
psql -U postgres -d devhub -c "CREATE SCHEMA IF NOT EXISTS hydra;"
psql -U postgres -d devhub -c "CREATE SCHEMA IF NOT EXISTS kratos;"
```

DSN 예시 (운영 자격으로 교체):

- DevHub: `postgres://devhub:<pw>@<host>:5432/devhub?sslmode=disable`
- Hydra: `postgres://devhub:<pw>@<host>:5432/devhub?sslmode=disable&search_path=hydra`
- Kratos: `postgres://devhub:<pw>@<host>:5432/devhub?sslmode=disable&search_path=kratos`

## 2. 빌드

### 2.1 backend-core (Go)

```sh
cd backend-core
go build -o bin/devhub-backend .
```

생성물: `backend-core/bin/devhub-backend` (Linux/macOS) 또는 `.exe` (Windows).

### 2.2 frontend (Next.js)

```sh
cd frontend
npm ci
npm run build
```

생성물: `frontend/.next/` (정적 산출 + 서버 번들). Next.js 16 의 `npm run start` 가 이를 사용한다.

### 2.3 마이그레이션 (DevHub 스키마)

```sh
make migrate-tools     # 최초 1회
MIGRATE_DB_URL="postgres://devhub:<pw>@<host>:5432/devhub?sslmode=disable" make migrate-up
```

적용 후 `rbac_policies` 시스템 role 3건(`developer`/`manager`/`system_admin`) 이 자동 시드된다 (M1 PR-G3).

### 2.4 IdP 마이그레이션

```sh
hydra  migrate sql --yes "postgres://devhub:<pw>@<host>:5432/devhub?sslmode=disable&search_path=hydra"
kratos migrate sql --yes "postgres://devhub:<pw>@<host>:5432/devhub?sslmode=disable&search_path=kratos"
```

## 3. 환경변수 (테스트 서버 기본값)

값은 `localhost` 가정. 다른 호스트로 쪼갤 때는 모든 URL 의 host 를 함께 갱신.

### 3.1 backend-core

| 변수 | 권장값 | 설명 |
| --- | --- | --- |
| `PORT` | `8080` | gin 바인딩 포트 |
| `DEVHUB_ENV` | `prod` | `prod` 시 verifier 미주입/dev fallback 활성 시 startup 거부 |
| `DB_URL` | `postgres://devhub:<pw>@localhost:5432/devhub?sslmode=disable` | DevHub master DB |
| `DEVHUB_HYDRA_ADMIN_URL` | `http://localhost:4445` | Bearer token introspection + login/logout/consent admin |
| `DEVHUB_HYDRA_PUBLIC_URL` | `http://localhost:4444` | `/oauth2/token` (PKCE 교환), `/oauth2/revoke` (refresh token revoke) |
| `DEVHUB_KRATOS_PUBLIC_URL` | `http://localhost:4433` | login self-service flow |
| `DEVHUB_KRATOS_ADMIN_URL` | `http://localhost:4434` | identity 발급 (Sign Up) |
| `DEVHUB_HYDRA_ROLE_CLAIM` | `ext.role` (default) | introspection 응답에서 role 추출 경로 |
| `DEVHUB_AUTH_DEV_FALLBACK` | (미설정) | `prod` 에서는 절대 1 로 두지 않는다 |
| `BACKEND_AI_URL` | `http://localhost:8000` | AI 서비스 미사용 시 비워도 됨 |
| `GITEA_URL` / `GITEA_TOKEN` / `GITEA_WEBHOOK_SECRET` | (환경별) | Gitea 연동 시만 |
| `SERVICE_ACTION_EXECUTOR_MODE` | `simulation` 또는 비움 | live executor 활성 여부 |

### 3.2 frontend

`.env.production` 또는 `npm run start` 직전 export.

| 변수 | 권장값 | 설명 |
| --- | --- | --- |
| `BACKEND_API_URL` | `http://localhost:8080` | Next.js rewrite 의 backend |
| `NEXT_PUBLIC_OIDC_AUTH_URL` | `http://localhost:4444/oauth2/auth` | Hydra public authorize |
| `NEXT_PUBLIC_OIDC_CLIENT_ID` | `devhub-frontend` | OIDC client id (§4 에서 등록) |
| `NEXT_PUBLIC_OIDC_REDIRECT_URI` | `http://localhost:3000/auth/callback` | (생략 시 `window.location.origin` 자동 산출) |
| `NEXT_PUBLIC_OIDC_SCOPE` | `openid offline_access email profile` | refresh_token 발급에 `offline_access` 필수 |
| `NEXT_PUBLIC_KRATOS_PUBLIC_URL` | `http://localhost:4433` | `/account` 비밀번호 변경 + Sign Out 의 Kratos browser logout |

### 3.3 Hydra/Kratos 설정 파일

저장소의 `infra/idp/hydra.yaml`, `infra/idp/kratos.yaml` 을 그대로 사용하거나 호스트/포트 조정.

운영 hygiene:

- `secrets.system` (hydra), `secrets.cookie` / `secrets.cipher` (kratos) — 운영용 임의값으로 교체
- `--dev` 플래그 제거 (HTTPS 강제)
- `urls.{login,logout,error,consent,post_logout_redirect}` 호스트가 frontend 호스트와 정확히 일치해야 함
- `selfservice.allowed_return_urls` 에 frontend origin 추가

## 4. OIDC client + 사용자 시드

### 4.1 OIDC client 등록 (1회)

Hydra 가동 후:

```powershell
# Windows PowerShell
.\infra\idp\scripts\register-devhub-client.ps1
```

리눅스 등 PowerShell 미설치 환경에서는 동일 페이로드를 `curl` 로 등록 (`POST <hydra-admin>/admin/clients`).

### 4.2 Kratos identity 시드 (테스트 사용자 1명)

`metadata_public.user_id` 를 DevHub `users.user_id` 와 1:1 매칭시킨다. 누락 시 backend 가 `auth.login.subject_fallback` 로그를 남기고 Kratos identity.id (UUID) 를 subject 로 사용 → RBAC 가 인식 못 함.

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

`traits.system_id` 가 `identity.schema.json` (`infra/idp/identity.schema.json`) 의 password identifier 다. 로그인 폼 "System ID" 입력값이 이 값과 매칭된다. 누락 시 Kratos 가 `400 missing properties: "system_id"` 로 거절한다.

같은 user_id 로 DevHub `users` 행도 추가:

```sh
psql -U devhub -d devhub -c "
INSERT INTO users (user_id, email, display_name, role, status)
VALUES ('alice', 'alice@example.com', 'Alice', 'system_admin', 'active');
"
```

(role 은 `developer` / `manager` / `system_admin` 중 1)

### 4.3 PoC 빠른 진입용 test 계정 (선택)

브라우저 smoke 가 자주 필요할 때 `test`/`test` 한 쌍을 시스템 관리자로 시드해 두면 편하다. Kratos identity + DevHub users 양쪽 모두 idempotent:

```sh
# Kratos identity
curl -X POST http://localhost:4434/admin/identities \
  -H "Content-Type: application/json" \
  -d '{
    "schema_id": "devhub_user",
    "traits": { "system_id": "test", "email": "test@example.com", "display_name": "Test Admin" },
    "metadata_public": { "user_id": "test" },
    "credentials": { "password": { "config": { "password": "test" } } }
  }'

# DevHub users (psql 미설치 환경은 idp-apply-schemas 헬퍼)
psql -U postgres -d devhub -f infra/idp/sql/003_seed_test_admin.sql
# 또는: go run ./backend-core/cmd/idp-apply-schemas -sql infra/idp/sql/003_seed_test_admin.sql
```

운영 진입 전에 §10 체크리스트의 "test/test 제거" 항목으로 일소.

## 5. 기동 순서

각 프로세스를 별도 창/터미널에서 실행. 한 창이 죽으면 그 컴포넌트만 재기동.

```sh
# 1) PostgreSQL — 시스템 서비스로 이미 기동되어 있다고 가정

# 2) Hydra — yaml 의 dsn 에 credential 이 없으므로 DSN env 로 override
DSN="postgres://devhub:<pw>@localhost:5432/devhub?sslmode=disable&search_path=hydra" \
hydra serve all --config infra/idp/hydra.yaml

# 3) Kratos (저장소 루트에서 실행 — identity.schema.json 의 file:// 상대경로 때문)
DSN="postgres://devhub:<pw>@localhost:5432/devhub?sslmode=disable&search_path=kratos" \
kratos serve --config infra/idp/kratos.yaml

# 4) backend-core
cd backend-core
DEVHUB_ENV=prod \
DB_URL="postgres://devhub:<pw>@localhost:5432/devhub?sslmode=disable" \
DEVHUB_HYDRA_ADMIN_URL=http://localhost:4445 \
DEVHUB_HYDRA_PUBLIC_URL=http://localhost:4444 \
DEVHUB_KRATOS_PUBLIC_URL=http://localhost:4433 \
DEVHUB_KRATOS_ADMIN_URL=http://localhost:4434 \
./bin/devhub-backend

# 5) frontend
cd frontend
NEXT_PUBLIC_OIDC_AUTH_URL=http://localhost:4444/oauth2/auth \
NEXT_PUBLIC_OIDC_CLIENT_ID=devhub-frontend \
NEXT_PUBLIC_OIDC_SCOPE="openid offline_access email profile" \
NEXT_PUBLIC_KRATOS_PUBLIC_URL=http://localhost:4433 \
BACKEND_API_URL=http://localhost:8080 \
npm run start
```

PowerShell 사용자는 `export` 대신 `$env:NAME = "value"` 후 실행 또는 `Start-Process` 로 별도 창 분리.

## 6. 헬스체크

```sh
curl http://localhost:8080/health                     # backend-core
curl http://localhost:4444/health/ready                # Hydra public
curl http://localhost:4445/health/ready                # Hydra admin
curl http://localhost:4433/health/ready                # Kratos public
curl http://localhost:4434/health/ready                # Kratos admin
curl -I http://localhost:3000/                         # frontend
```

모두 200 이면 다음 단계.

## 7. 로그인 e2e 검증

브라우저로:

1. `http://localhost:3000/login` 접속 → 자동으로 Hydra `/oauth2/auth` 진입
2. Hydra 가 `/auth/login?login_challenge=...` 로 redirect → password 폼 노출
3. `alice@example.com` / `ChangeMe-12345!` 입력
4. `/auth/callback` 으로 돌아옴 → token 교환 → 역할 기반 페이지(`/developer` / `/manager` / `/admin`) 로 이동
5. 우측 상단 사용자 메뉴 → **Account Settings** 진입 → 비밀번호 변경 시도 (Kratos `/self-service/settings` 흐름)
6. **Sign Out** 클릭 → Hydra `/oauth2/sessions/logout` 으로 navigate → `/auth/logout?logout_challenge=...` → backend Hydra accept + Kratos cookie kill → `/` 복귀
7. 다시 `/login` → password 재입력 요구 (Hydra session 종료 확인)

backend 로그에서:
- `auth.login.succeeded` 1건
- `auth.logout.succeeded` 1건 (revoke_status=succeeded)

DB:
```sh
psql -U devhub -d devhub -c "SELECT action, target_type, target_id, created_at FROM audit_logs ORDER BY created_at DESC LIMIT 10;"
```

## 8. 재기동 / 롤백

| 시나리오 | 절차 |
| --- | --- |
| backend 코드 hotfix | `cd backend-core && go build -o bin/devhub-backend . && (기존 프로세스 종료 후) ./bin/devhub-backend` |
| frontend 코드 hotfix | `cd frontend && npm run build && (기존 npm run start 종료 후) npm run start` |
| DB 스키마 롤백 | `MIGRATE_DB_URL=... make migrate-down` |
| Hydra/Kratos 재시작 | 해당 프로세스만 재기동. DB 는 유지 |
| 시크릿 교체 | `infra/idp/{hydra,kratos}.yaml` 의 `secrets.*` 갱신 → 두 프로세스 재기동. session 모두 무효화됨 |
| OIDC client 재생성 | `register-devhub-client.ps1` 재실행 (idempotent — 같은 `client_id` 면 PUT 으로 갱신) |

## 9. 트러블슈팅

| 증상 | 원인 | 조치 |
| --- | --- | --- |
| `/login` 진입 시 무한 redirect | Hydra `urls.login` 호스트가 frontend 와 다름 | `infra/idp/hydra.yaml` 의 `urls.login` 정정 후 Hydra 재기동 |
| 로그인 후 `/api/v1/me` 401 | backend 가 token 검증 못 함 | `DEVHUB_HYDRA_ADMIN_URL` 셋·도달 가능 여부 확인 |
| 로그인 후 RBAC 권한 거부 (auth.policy_unmapped, auth.role_denied) | Kratos identity 의 `metadata_public.user_id` 누락 → subject 가 UUID 로 fallback | §4.2 의 identity payload 에 `metadata_public.user_id` 포함 + DevHub `users` 같은 user_id 로 row 존재 확인 |
| Sign Out 후 다시 `/login` 가도 password 재입력 안 함 | Hydra session 종료 안 됨 (id_token 미저장 fallback) | frontend 가 token-store 에 id_token 저장하는지(=callback 흐름이 정상이었는지) 확인. 또는 Hydra cookie 직접 삭제 |
| `/account` 비밀번호 변경 시 "Sign In Again" 노출 | Kratos privileged_session_max_age=15m 초과 | 사용자가 다시 로그인하면 자동 해소. 무한 발생 시 Kratos 세션 cookie 도메인/SameSite 정책 점검 |
| 사내 SSL inspection 으로 `go install` 실패 | 외부 mirror 차단 | 사내 GoProxy 사용 (`GOPROXY=https://...` export). docker 이미지 mirror 가 아님 — release binary mirror 또는 소스 빌드 |
| Kratos identity admin 호출 401 | `/admin` 포트(4434) 가 외부에 노출됨 | 운영에서는 admin 포트를 내부망/방화벽 안으로 제한 |

## 10. 운영 체크리스트 (테스트 → 운영 전환 시)

- [ ] PoC secrets (`secrets.system`, `secrets.cookie`, `secrets.cipher`) 운영용 임의 값으로 교체
- [ ] Hydra/Kratos `--dev` 플래그 제거, HTTPS 강제
- [ ] Hydra/Kratos `urls.*` 와 frontend host 가 동일한 origin 으로 통일 (또는 SameSite/CORS 정책 일관)
- [ ] OIDC client 의 `redirect_uris` / `post_logout_redirect_uris` 에 운영 host 만 포함
- [ ] backend `DEVHUB_AUTH_DEV_FALLBACK` 가 unset (또는 0)
- [ ] backend `DEVHUB_ENV=prod` 설정 (verifier 미주입 시 startup 거부)
- [ ] PostgreSQL 사용자/롤 분리 (DevHub / Hydra / Kratos 각자)
- [ ] Hydra/Kratos `/admin` 포트(4445/4434) 외부 차단
- [ ] backend `audit_logs` 모니터링 (`auth.login.subject_fallback` / `auth.logout.revoke_failed` / `auth.policy_unmapped` / `auth.role_denied` 알림 연결)
- [ ] 시스템 서비스 등록 (Windows Service / systemd / launchd) — 후속 phase
- [ ] PoC 빠른 진입용 `test`/`test` 시스템 관리자 계정 제거 — Kratos identity 삭제 (`DELETE /admin/identities/{id}`) + DevHub `users` 의 `user_id='test'` 행 삭제 + `infra/idp/sql/003_seed_test_admin.sql` 운영 시드 경로에서 제외
