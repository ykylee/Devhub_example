# IdP Setup — 사내 환경 특수 제약 메모

- 문서 목적: Phase 13 PoC 를 진행한 환경(2026-05-07 시점) 의 특수 제약과 그에 대응한 우회 결정·도구를 따로 정리한다. 일반 환경 setup 가이드는 [README.md](./README.md) 에 있다.
- 범위: Ory Hydra + Kratos PoC 가동 시 환경 의존적으로 추가된 우회 도구·설정.
- 대상 독자: 같은 사내 환경에서 본 저장소를 처음 가동하는 신규 합류자, 외부 환경에서 본 저장소를 참고하려는 사람 (왜 이런 우회 도구가 있는지 이해 목적).
- 상태: active
- 최종 수정일: 2026-05-07

## 1. 환경 특성 (Phase 13 PoC 진행 시점)

| 항목 | 상태 |
| --- | --- |
| OS | Windows 11 Enterprise (Korean) |
| 기본 PowerShell | Windows PowerShell 5.1 (`$PSVersionTable.PSVersion = 5.1.x`) |
| PostgreSQL | 로컬 native 설치, `127.0.0.1:5432`, DB `devhub` (`postgres`/`postgres`). **CLI 도구(`psql.exe`) 가 PATH 에 없음** — GUI 또는 별도 도구로 운영. |
| Go | `C:\Program Files\Go\bin` PATH 등록. `go env GOPROXY=https://proxy.golang.org,direct`, `GOSUMDB=sum.golang.org` (기본값). |
| 네트워크 | 사내 SSL inspection 환경. **`proxy.golang.org` / `sum.golang.org` 외부 접근 차단**. GitHub release binary 직접 다운로드(`https://github.com/.../releases/download/...`) 는 가능. |
| 기타 | Docker / docker-compose 미사용 정책 (ADR-0001 §2). |

## 2. 환경 특수 제약과 우회 결정

### 2.1 사내 GoProxy / Sum DB 차단 → binary 다운로드 경로 채택

**제약**: `go install <module>` 또는 신규 모듈에 대한 `go mod tidy` 가 `proxy.golang.org` / `sum.golang.org` 에 도달하지 못해 timeout. 기존 backend-core 의 의존성은 이미 로컬 module cache 에 내려받아져 있어 `go run ./backend-core` 는 동작하지만, **신규 의존성 추가는 차단** 된다.

**Ory 도구의 일반 제약과의 구분**: Ory Hydra/Kratos 는 `go.mod` 에 `replace` 지시문을 포함하므로 `go install` 자체가 **모든 환경에서** 차단된다. 따라서 binary 다운로드 경로(Scoop / brew / GitHub release zip) 가 표준 설치 방법이다 — 본 환경의 제약과 별개로 일반적인 권장 경로다. 다만 본 환경에서는 **`go install` 시도 자체가 무의미**하다는 사실이 더 일찍 명확해졌다.

**우회 도구**: [`scripts/install-binaries.ps1`](./scripts/install-binaries.ps1) — GitHub release 의 `*-windows_sqlite_64bit.zip` 을 다운로드해 `$env:USERPROFILE\go\bin\` 에 `.exe` 를 배치. SSL inspection 통과는 GitHub.com 자체로는 문제없는 환경이라 가능. Scoop bucket 도 GitHub 기반이라 사용 가능하지만 본 저장소가 자체 헬퍼를 둔 이유는 (a) 외부 패키지 매니저 의존을 줄이기 위해, (b) 사내 mirror 로의 이전이 필요할 때 한 곳만 고치도록 추상화 지점을 만들기 위해.

### 2.2 `psql.exe` 미설치 → Go 헬퍼 신규 작성

**제약**: dev workstation 에 PostgreSQL CLI (`psql.exe`) 가 PATH 에 없다. GUI 도구로 운영 중. Schema 생성 SQL 을 일회성으로 적용할 표준 경로가 없음.

**우회 도구**: [`backend-core/cmd/idp-apply-schemas/main.go`](../../backend-core/cmd/idp-apply-schemas/main.go) — backend-core 의 기존 pgx/v5 의존성을 재사용해 SQL 파일을 실행한다. 별도 모듈을 만들면 §2.1 의 GoProxy 차단으로 신규 의존성 다운로드가 막히기 때문에, **backend-core 의 module 안에서** cmd 형태로 추가했다. 1회성 setup 도구이므로 backend-core/cmd/ 에 두는 것이 코드 재사용·유지보수 양면에서 적절했다.

> 일반 환경에서는 `psql -U postgres -d devhub -f infra/idp/sql/001_create_idp_schemas.sql` 한 줄이면 끝나며, 본 헬퍼는 사용할 필요가 없다.

### 2.3 PowerShell 5.1 + 한국어 Windows 인코딩 → `.ps1` ASCII 강제

**제약**: Windows 의 기본 PowerShell 은 5.1 이며, BOM 없는 UTF-8 `.ps1` 파일을 ANSI/CP949 (Korean code page) 로 해석한다. 한글 주석/문자열을 포함하면 바이트 시퀀스가 깨져 따옴표·중괄호 매칭이 무너지고 `TerminatorExpectedAtEndOfString` 같은 파서 에러로 실행 자체가 막힌다.

**우회 결정**: 본 저장소의 `.ps1` 파일은 **주석/메시지를 ASCII (영어) 만 사용** 한다. 한글 안내가 필요하면 같은 디렉터리의 `README.md` 에 둔다. BOM 추가 대신 ASCII 강제를 선택한 이유는 BOM 은 git core.autocrlf, 다른 에디터, 일부 도구를 거치며 손실 가능성이 있는 반면 ASCII 는 가장 견고한 invariant.

> 일반 환경(macOS/Linux 또는 Windows + PowerShell 7+) 에서는 BOM 없는 UTF-8 도 안전하게 동작하므로 이 제약은 적용되지 않는다.

## 3. 어느 도구가 환경 특수이고 어느 것이 일반 가이드인가

| 자산 | 일반 환경 | 본 저장소의 corp env |
| --- | --- | --- |
| `infra/idp/scripts/install-binaries.ps1` | 옵션 (Scoop/brew 사용 가능) | 권장 (Scoop 도 GitHub 기반이라 동등 — 우리는 자체 헬퍼 사용) |
| `backend-core/cmd/idp-apply-schemas/main.go` | 사용 안 함 (psql 사용) | 필요 (psql 미설치) |
| `infra/idp/scripts/register-devhub-client.ps1` | 권장 (Hydra v26 CLI 한계 우회) — 환경 무관 | 동일 |
| `infra/idp/sql/001_create_idp_schemas.sql` | 필요 — 환경 무관 | 동일 |
| `infra/idp/{hydra,kratos}.yaml` | 필요 — 환경 무관 | 동일 |
| `infra/idp/identity.schema.json` | 필요 — 환경 무관 | 동일 |

## 4. 신규 합류자가 다른 환경에서 시작할 때

- **외부 인터넷 접근이 자유로운 환경**: 본 문서 무시하고 [README.md](./README.md) 만 따라가면 된다. Scoop/brew 로 binary 설치 + psql 로 schema 생성 + 표준 절차.
- **macOS/Linux**: PowerShell 스크립트는 사용하지 않는다. binary 설치는 brew 또는 tar.gz 다운로드, OIDC client 등록은 admin REST API 를 직접 `curl` 로 호출하면 된다 (`register-devhub-client.ps1` 의 body 를 참고).
- **유사한 사내 corp env**: 본 문서를 그대로 따라가면 된다. 추가로 §2.1 의 사내 GoProxy 가 있다면 `GOPROXY` env 를 설정해 backend-core 의존성 재다운로드도 가능.

## 5. 관련 메모리

이 문서의 내용 중 일부는 향후 세션에서도 적용되어야 하는 정책이라 auto memory 에도 저장되어 있다 (개인 Claude memory, 저장소 외부):

- `feedback_no_docker.md` — Docker 미사용 전제.
- `feedback_powershell_ascii.md` — `.ps1` ASCII 강제.

(저장소 안의 `.claude/` 는 gitignore 됨. 본 ENVIRONMENT_NOTES.md 가 저장소 안에 남는 share 가능 사본.)
