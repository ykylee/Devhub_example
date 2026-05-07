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

> ⚠️ **샌드박스 외 사용자 터미널**에서 직접 실행. AI 자동화는 binary 설치를 수행하지 않음.

### 1차 경로: GitHub release Windows binary 다운로드 (권장)

> Hydra/Kratos 의 `go.mod` 가 `replace` 지시문을 포함해 **`go install` 은 차단**된다 (2026-05-07 확인). release binary 의 SQLite 변형은 CGO 의존이 없고 embed 된 migration 자산을 포함해 Windows 에서 가장 안전한 선택이다. Ory 는 두 프로젝트에 통일된 release-train 버전을 사용한다 (예: `v26.2.0`).

자동화 스크립트로 다운로드 + 압축 해제 + 배치:

```powershell
.\infra\idp\scripts\install-binaries.ps1
# 또는 버전/위치 지정:
.\infra\idp\scripts\install-binaries.ps1 -Version 26.2.0 -BinDir "$env:USERPROFILE\go\bin"
```

설치 확인:

```powershell
hydra version
kratos version
```

`-BinDir` 가 PATH 에 없으면 스크립트가 경고를 출력한다. 기본값 `$env:USERPROFILE\go\bin` 은 Go 사용자가 보통 PATH 에 이미 추가해 둔다.

### 2차 경로 (수동 다운로드)

스크립트가 실패하면 직접:

- https://github.com/ory/hydra/releases/latest — `hydra_<ver>-windows_sqlite_64bit.zip`
- https://github.com/ory/kratos/releases/latest — `kratos_<ver>-windows_sqlite_64bit.zip`

zip 해제 후 `hydra.exe` / `kratos.exe` 를 PATH 가 잡히는 위치에 배치.

### 차단된 경로 (참고)

- `go install github.com/ory/hydra/v2@latest` → `replace` 지시문 때문에 차단.
- `go install github.com/ory/hydra/v2/cmd/hydra@latest` → `cmd/hydra` 패키지 부재 (main 이 모듈 루트에 있음).

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
