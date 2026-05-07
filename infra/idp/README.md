# DevHub IdP — PoC 환경 가이드

- 문서 목적: ADR-0001 결정에 따라 Ory Hydra + Ory Kratos 를 native binary 로 가동하는 1차 PoC 절차를 정의한다.
- 범위: Phase 13 P1 1단계 (PoC) — sub-step (a) binary 설치 → (b) DB schema → (c) 설정 → (d) identity schema → (e) OIDC client 등록 → (f) round-trip 검증.
- 대상 독자: DevHub 개발자, AI Agent.
- 상태: in_progress (Phase 13 P1 1단계)
- 최종 수정일: 2026-05-07
- 관련 문서: [ADR-0001](../../docs/adr/0001-idp-selection.md), [backend_development_roadmap.md P1](../../ai-workflow/memory/backend_development_roadmap.md)

## 0. 전제

- 로컬 PostgreSQL 가 `127.0.0.1:5432` 에 가동 중이고 db `devhub` 에 접근 가능 (`postgres`/`postgres`).
- 마이그레이션 `000001~000004` 가 적용된 상태.
- DevHub backend-core (`go run ./backend-core`) 와 frontend (`npm run dev`) 가 가동 가능한 상태.
- **Docker / docker-compose 미사용** (정책, ADR-0001 §2 / `MEMORY.md` `feedback_no_docker`).

## 1. 사용자 수동 단계 — Hydra/Kratos binary 설치

> ⚠️ **샌드박스 외 사용자 터미널**에서 직접 실행. AI 자동화는 binary 설치를 수행하지 않음 (사내 GoProxy 미러 통과 필요).

두 프로젝트 모두 `main.go` 가 모듈 루트에 있다. install path 에 `/cmd/...` 를 붙이면 실패한다 (`cmd/` 디렉터리는 라이브러리 서브패키지일 뿐 binary entry 가 아님).

```powershell
# 사내 GoProxy 미러가 환경에 이미 설정되어 있다고 가정
go install github.com/ory/hydra/v2@latest
go install github.com/ory/kratos@latest
```

설치 확인:

```powershell
hydra version
kratos version
```

`go install` 의 binary 출력 디렉터리(`$env:GOPATH\bin` 또는 `$env:USERPROFILE\go\bin`)가 `PATH` 에 포함되어 있어야 한다.

### 폴백: GitHub release binary 다운로드

`go install` 이 embed 자산 누락(SQLite 미지원 등) 으로 실패할 경우 GitHub release 의 Windows binary (가능하면 `*-windows_sqlite_64bit.zip`) 를 다운로드해 zip 해제 후 PATH 가 잡히는 위치(예: `$env:USERPROFILE\go\bin`) 에 배치한다.

- https://github.com/ory/hydra/releases (latest stable)
- https://github.com/ory/kratos/releases (latest stable)

## 2. DB schema 생성

```powershell
psql -U postgres -d devhub -f infra/idp/sql/001_create_idp_schemas.sql
```

확인:

```powershell
psql -U postgres -d devhub -c "\dn"
```

`hydra`, `kratos` schema 가 출력되면 성공.

## 3. Hydra/Kratos 자체 마이그레이션

각 도구가 자기 schema 안에 테이블을 생성한다.

```powershell
$DSN_HYDRA  = "postgres://postgres:postgres@127.0.0.1:5432/devhub?sslmode=disable&search_path=hydra"
$DSN_KRATOS = "postgres://postgres:postgres@127.0.0.1:5432/devhub?sslmode=disable&search_path=kratos"

hydra migrate sql --yes $DSN_HYDRA
kratos migrate sql --yes $DSN_KRATOS
```

## 4. Hydra/Kratos 가동

별도 PowerShell 창 2개에서 각각 실행 (PoC 단계는 직접 실행, ADR-0001 §8.7 결정).

```powershell
# 창 A — Kratos
kratos serve --config infra/idp/kratos.yaml

# 창 B — Hydra (개발용 --dev: HTTPS 강제 해제)
hydra serve all --config infra/idp/hydra.yaml --dev
```

가동 확인:

- Hydra public: <http://localhost:4444/.well-known/openid-configuration>
- Hydra admin: <http://localhost:4445/health/ready>
- Kratos public: <http://localhost:4433/health/ready>
- Kratos admin: <http://localhost:4434/health/ready>

## 5. DevHub OIDC client 등록

Hydra 가 가동 중인 상태에서 한 번만 실행.

```powershell
.\infra\idp\scripts\register-devhub-client.ps1
```

확인:

```powershell
hydra list oauth2-clients --endpoint http://localhost:4445
```

`devhub-frontend` 가 목록에 보이면 성공.

## 6. round-trip 검증 (1단계 (f))

> 이 단계는 frontend `/auth/login`, `/auth/consent`, `/auth/callback` 라우트 구현이 필요하다. PoC 1단계 (f) 진입 시점에 본 README 와 frontend 라우트 구현 PR 을 별도로 진행.

대략의 흐름:
1. 브라우저에서 `http://localhost:3000/auth/login` 진입 → DevHub Next.js 가 Hydra `/oauth2/auth?client_id=devhub-frontend&...&code_challenge=...` 로 redirect.
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
| PostgreSQL | 127.0.0.1:5432 (db=devhub) |

## 운영 환경 전환 시 검토 항목

PoC 가동 후 운영 진입 전 별도 결정·교체 필요:

- `hydra.yaml` `secrets.system`, `kratos.yaml` `secrets.cookie` / `secrets.cipher` 운영 secret 으로 교체 (예: 사내 vault).
- Hydra `--dev` 플래그 제거 + 운영 도메인 HTTPS 인증서.
- Hydra/Kratos 호스트 시스템 서비스 등록 (Windows Service / systemd) — ADR-0001 §8.7.
- Hydra/Kratos 별 전용 DB role 분리 (현재 `postgres` 슈퍼유저 그대로).
- `register-devhub-client.ps1` 의 redirect URI / post-logout URI 를 운영 도메인으로 교체.
- `kratos.yaml` `courier.smtp` 를 실제 메일 서버 (예: 사내 SMTP relay) 로 교체. recovery flow 활성 결정.
- Hydra 서명 키 회전 정책 결정 (PoC 는 default JWKS 자동 생성).
- 외부 SaaS client 추가 시 consent UI 구현 (ADR-0001 §8.2).
