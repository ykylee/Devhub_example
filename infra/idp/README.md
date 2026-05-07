# DevHub IdP — Ory Hydra + Kratos Setup Guide

- 문서 목적: ADR-0001 결정에 따라 Ory Hydra + Ory Kratos 를 native binary 로 가동하는 일반 환경 setup 가이드.
- 범위: binary 설치 → DB schema → 마이그레이션 → 가동 → OIDC client 등록 → round-trip 검증.
- 대상 독자: DevHub 개발자 (신규 합류자 포함), 운영자.
- 상태: active
- 최종 수정일: 2026-05-07
- 관련 문서:
    - [ADR-0001 — IdP 선택](../../docs/adr/0001-idp-selection.md)
    - [ENVIRONMENT_NOTES.md — 이 저장소가 가정하는 사내 corp 환경 특수 제약](./ENVIRONMENT_NOTES.md)
    - [backend_development_roadmap.md P1](../../ai-workflow/memory/backend_development_roadmap.md)

> ℹ️ 본 가이드는 보통의 개발 환경(외부 인터넷 접근, Scoop/brew 등 표준 패키지 매니저 사용 가능)을 가정한다. 사내 SSL inspection·GoProxy 미러 제약이 있는 환경에 대한 우회 절차는 [ENVIRONMENT_NOTES.md](./ENVIRONMENT_NOTES.md) 를 별도로 본다.

## 0. 전제

- 로컬 PostgreSQL 가 `127.0.0.1:5432` 에 가동 중이고 DB `devhub` 에 접근 가능 (`postgres`/`postgres`, 또는 동등 사용자).
- DevHub backend-core 마이그레이션 `000001~000004` 가 적용된 상태.
- DevHub backend-core (`go run ./backend-core`) 와 frontend (`npm run dev`) 가 가동 가능한 상태.
- **Docker / docker-compose 미사용** (정책, ADR-0001 §2).

## 1. Hydra / Kratos binary 설치

> Ory Hydra/Kratos 의 `go.mod` 가 `replace` 지시문을 포함하므로 **`go install` 은 사용할 수 없다** (Go 도구의 일반 제약). 표준 설치 경로는 OS 별 패키지 매니저 또는 GitHub release binary.

### 1.1 Windows (권장: Scoop 또는 헬퍼 스크립트)

옵션 A — **Scoop**:

```powershell
scoop bucket add ory https://github.com/ory/scoop.git
scoop install ory/hydra ory/kratos
```

옵션 B — **본 저장소 헬퍼 스크립트** (Scoop 사용 불가 시):

```powershell
.\infra\idp\scripts\install-binaries.ps1
# 버전/위치 지정:
.\infra\idp\scripts\install-binaries.ps1 -Version 26.2.0 -BinDir "$env:USERPROFILE\go\bin"
```

스크립트는 GitHub release 에서 `<name>_<ver>-windows_sqlite_64bit.zip` 을 받아 `$env:USERPROFILE\go\bin\` 에 `.exe` 를 배치한다. SQLite 변형은 CGO 의존이 없고 embed 된 migration 자산을 포함한다.

### 1.2 macOS

```bash
brew install ory/tap/hydra ory/tap/kratos
```

### 1.3 Linux

GitHub release 의 tar.gz 를 다운로드해 PATH 에 풀거나, 사용 중인 distro 의 패키지 매니저를 사용한다.

- https://github.com/ory/hydra/releases/latest
- https://github.com/ory/kratos/releases/latest

### 1.4 설치 확인

```powershell
hydra version
kratos version
```

두 명령 모두 동일한 release-train 버전 (예: `v26.2.0`) 을 출력해야 한다.

## 2. DB schema 생성

Hydra/Kratos 는 `?search_path=...` DSN 옵션이 가리키는 schema 안에 자체 테이블을 만든다. 하지만 schema 자체는 미리 만들어 둬야 한다 (ADR-0001 §8.1 결정 — 단일 `devhub` DB 안 분리).

### 2.1 표준 경로 — psql

```powershell
psql -U postgres -d devhub -f infra/idp/sql/001_create_idp_schemas.sql
psql -U postgres -d devhub -c "\dn"
```

`hydra`, `kratos` schema 가 출력되면 성공.

### 2.2 폴백 — psql 없는 환경용 Go 헬퍼

dev workstation 에 `psql` 이 설치되어 있지 않을 경우 본 저장소의 헬퍼를 사용한다:

```powershell
Push-Location backend-core
go run ./cmd/idp-apply-schemas -sql ..\infra\idp\sql\001_create_idp_schemas.sql
Pop-Location
```

헬퍼는 backend-core 의 pgx/v5 의존성을 재사용해 SQL 파일을 실행한다. 자세한 배경은 [ENVIRONMENT_NOTES.md](./ENVIRONMENT_NOTES.md) 참조.

## 3. Hydra / Kratos 자체 마이그레이션

각 도구가 자기 schema 안에 테이블을 생성한다.

```powershell
$DSN_HYDRA  = "postgres://postgres:postgres@127.0.0.1:5432/devhub?sslmode=disable&search_path=hydra"
$DSN_KRATOS = "postgres://postgres:postgres@127.0.0.1:5432/devhub?sslmode=disable&search_path=kratos"

hydra migrate sql up --yes $DSN_HYDRA
kratos migrate sql up --yes $DSN_KRATOS
```

> Hydra v26 부터 `hydra migrate sql` (목적어 없는 형태) 는 deprecated. 위처럼 `up` 을 명시한다.

상태 확인:

```powershell
hydra migrate sql status $DSN_HYDRA | Select-Object -Last 5
kratos migrate sql status $DSN_KRATOS | Select-Object -Last 5
```

모든 행이 `Applied` 상태여야 한다.

## 4. Hydra / Kratos 가동

별도 창 2개에서 각각 실행 (운영 환경에서는 시스템 서비스로 등록 — ADR-0001 §8.7).

> ⚠️ Kratos 는 `identity.schemas[].url=file://./...` 를 **process working directory 기준** 으로 해석한다. 반드시 **저장소 루트**에서 launch 한다.

```powershell
# 창 A — Kratos (저장소 루트에서)
kratos serve --config infra\idp\kratos.yaml

# 창 B — Hydra (개발용 --dev: HTTPS 강제 해제)
hydra serve all --config infra\idp\hydra.yaml --dev
```

가동 확인:

```powershell
Invoke-WebRequest http://localhost:4444/.well-known/openid-configuration -UseBasicParsing | ConvertFrom-Json | Select-Object issuer, authorization_endpoint, token_endpoint, jwks_uri
Invoke-WebRequest http://localhost:4445/health/ready -UseBasicParsing
Invoke-WebRequest http://localhost:4433/health/ready -UseBasicParsing
Invoke-WebRequest http://localhost:4434/health/ready -UseBasicParsing
```

Hydra `:4444 / :4445`, Kratos `:4433 / :4434` 가 모두 응답하면 OK.

## 5. DevHub OIDC client 등록

Hydra 가 가동 중인 상태에서 한 번만 실행한다 (멱등 — 기존 `devhub-frontend` client 가 있으면 삭제 후 재생성).

```powershell
.\infra\idp\scripts\register-devhub-client.ps1
```

확인:

```powershell
hydra list oauth2-clients --endpoint http://localhost:4445
```

`devhub-frontend` 가 목록에 보여야 한다.

> 본 스크립트는 Hydra Admin REST API (`POST /admin/clients`) 를 호출해 `client_id=devhub-frontend` 를 명시적으로 지정한다. Hydra v26 의 `hydra create oauth2-client` CLI 는 `--client-id` 플래그를 무시하고 UUID 를 자동 발급하므로 재현 가능한 PoC 에는 부적합하다.

## 6. round-trip 검증

> 이 단계는 frontend `/auth/login`, `/auth/consent`, `/auth/callback` 라우트 구현이 선행되어야 한다 (Phase 5). frontend 라우트가 준비되면 본 절차로 검증한다.

흐름:
1. 브라우저에서 `http://localhost:3000/auth/login` 진입 → DevHub Next.js 가 Hydra `/oauth2/auth?client_id=devhub-frontend&code_challenge=...&...` 로 redirect.
2. Hydra 가 `login_challenge` 를 붙여 Kratos 의 login UI URL (Next.js `/auth/login`) 로 redirect.
3. 사용자가 Kratos public flow 로 자격 증명 검증.
4. Next.js 가 Hydra `accept login` (admin API) → `skip_consent=true` 로 자동 consent → `/auth/callback?code=...` 로 redirect.
5. Next.js 가 Hydra `/oauth2/token` 으로 code 교환 → ID Token + Access Token + Refresh Token 수신.

## 포트 요약

| 서비스 | URL |
| --- | --- |
| Frontend (Next.js) | http://localhost:3000 |
| Backend Core (Go Core) | http://localhost:8080 |
| Hydra public | http://localhost:4444 |
| Hydra admin | http://localhost:4445 |
| Kratos public | http://localhost:4433 |
| Kratos admin | http://localhost:4434 |
| PostgreSQL | 127.0.0.1:5432 (DB=devhub) |

## 일반적인 함정 (general gotchas)

`go install` 미지원 외에도 환경에 무관하게 적용되는 Ory 도구의 quirks:

- **Hydra v26 CLI `--client-id` 무시**: `hydra create oauth2-client` 는 client_id 를 자동 UUID 로 발급. 재현 가능한 등록은 admin REST API 사용 (본 저장소의 `register-devhub-client.ps1` 가 이 경로).
- **`hydra migrate sql` deprecated**: v26 부터 `hydra migrate sql up` 으로 명시.
- **Kratos identity schema URL 의 cwd 의존**: `identity.schemas[].url=file://./...` 가 process cwd 기준 해석. 저장소 루트에서 launch 해야 한다.
- **Kratos `secrets.cipher` 길이 ≤ 32 byte**: AES-256 의 max key length. 더 길면 부팅 실패.

## 운영 환경 전환 시 검토 항목

PoC 가동 후 운영 진입 전 별도 결정·교체 필요:

- `hydra.yaml` `secrets.system`, `kratos.yaml` `secrets.cookie` / `secrets.cipher` 운영 secret 으로 교체 (예: 사내 vault).
- Hydra `--dev` 플래그 제거 + 운영 도메인 HTTPS 인증서.
- Hydra/Kratos 호스트 시스템 서비스 등록 (Windows Service / systemd / launchd) — ADR-0001 §8.7.
- Hydra/Kratos 별 전용 DB role 분리 (현재 `postgres` 슈퍼유저 그대로).
- `register-devhub-client.ps1` 의 redirect URI / post-logout URI 를 운영 도메인으로 교체.
- `kratos.yaml` `courier.smtp` 를 실제 메일 서버 (예: 사내 SMTP relay) 로 교체. recovery flow 활성 결정.
- Hydra 서명 키 회전 정책 결정 (PoC 는 default JWKS 자동 생성).
- 외부 SaaS client 추가 시 consent UI 구현 (ADR-0001 §8.2).
