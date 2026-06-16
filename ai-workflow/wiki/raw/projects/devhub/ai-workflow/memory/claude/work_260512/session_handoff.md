# 세션 인계 문서 (2026-05-12 work slot, mid-session)

## 세션 목표

직전 슬롯 `claude/260512` handoff 의 "다음 세션 출발점 후보 #1" — `dev-up.ps1` / `dev-up.sh` 실머신 실행 검증을 본 슬롯에서 진행한다. 검증 과정에서 발견한 결함 #1 (`npm` Start-Process) 과 #3 (외부 IdP 충돌) 까지 픽스하고 검증한다. 결함 #2 (grandchild PID) 와 신규 결함 #4 (kratos DSN cold-start) 는 다음 슬롯으로 이월.

## 실머신 검증 결과 (1차 — clean fix path)

| 항목 | 결과 |
| --- | --- |
| 환경 | Win11 Enterprise + PowerShell 5.1, kratos / hydra / postgres 가 user 관리로 사전 기동 (preflight kratos pid=29004, hydra pid=16872) |
| dev-up 1차 실행 | frontend launch 에서 `Start-Process -FilePath 'npm'` 이 `npm.ps1` 로 resolve → `%1 is not a valid Win32 application` 으로 거부 → 결함 #1 |
| 결함 #1 픽스 | `npm` → `npm.cmd` (3줄, `dev-up.ps1` 의 frontend block). `.cmd` 는 정상적 Win32 launcher 라 CreateProcess 통과. |
| dev-up 2차 실행 | 5/5 서비스 ready, backend `/health` 200 `{"db":"ok","service":"backend-core","status":"ok"}`, frontend `/` 200 (23KB) |
| dev-down (1차) | 6/6 포트 free, `.pids/` 비움 — 단, kratos pid=29004 와 hydra pid=16872 (user managed) 까지 port-sweep 으로 함께 종료 → 결함 #3 재현 |

## 결함 박제

### 결함 #1 [FIXED — 이번 브랜치]
`Start-Process -FilePath 'npm'` 이 PowerShell 5.1 에서 `npm.ps1` 로 resolve, CreateProcess 가 `.ps1` 을 Win32 exe 로 인정하지 않아 거부. `npm.cmd` 직접 지정으로 해결.

### 결함 #2 [미수정 — 이월]
backend 의 `go run .` 와 frontend 의 `npm.cmd` 가 launcher 로 동작하고 실제 서버는 grandchild 프로세스. `dev-down.ps1` 의 `Stop-Process` 는 launcher PID 만 종료하고 grandchild 는 살아남음. port-sweep 이 실질적 종료 경로. PID 추적 의미가 약해짐. 픽스 옵션: `go run` → `go build` + 빌드 산출물 실행, 혹은 grandchild PID 까지 같이 추적.

### 결함 #3 [FIXED — 이번 브랜치]
**증상**: dev-up 이 포트 점유 검사 없이 `Get-Command kratos/hydra` 만으로 spawn → user-managed 인스턴스가 이미 4433/4444 를 잡고 있으면 새 instance 는 bind fail 후 zombie 가 되고 `.pids/<svc>.pid` 가 zombie 를 가리킴. 이어서 dev-down 의 port-sweep 4433/4434/4444/4445 가 user-managed 인스턴스까지 죽임. handoff 정책 "kratos / hydra 는 직접 띄운 그대로 유지" 와 정면 충돌.

**픽스**: 4개 파일 모두 동일 의미론으로 수정 (dev-up.ps1, dev-down.ps1, dev-up.sh, dev-down.sh). 핵심 원칙: **"dev-up 이 spawn 한 서비스만 PID 파일이 생기고, dev-down 은 PID 파일이 있는 서비스만 정리한다 (port-sweep 도 그 서비스의 포트만)."**
- `dev-up`: 각 서비스(kratos/hydra/backend/frontend) 시작 전 메인 포트 LISTEN 체크 → 이미 떠 있으면 spawn / Wait-ForPort / PID 작성 모두 skip + "external instance detected" 로그
- `dev-down`: PID 파일 있는 서비스만 stop + 그 서비스의 포트만 sweep 대상 누적. PID 파일 없으면 "not started by this script; leaving any listener on port(s) X intact" 로그.

### 결함 #4 [신규, 미수정 — 이월]
clean cold-start 시 kratos 가 4433 에 bind 실패. 원인: `infra/idp/kratos.yaml` 의 DSN `postgres://localhost:5432/devhub?sslmode=disable&search_path=kratos` 에 user/password 없어 libpq 가 OS 유저(현 머신에서는 `sam`) 로 폴백 → `failed SASL auth ... password authentication failed`. 픽스 옵션: (a) DSN 에 `postgres://postgres:postgres@...` 명시, (b) dev-up 이 PGUSER/PGPASSWORD env export, (c) kratos 전용 wrapper 스크립트로 env 주입.

## 결함 #3 픽스 검증 — Scenario A (외부 IdP 보존)

수동으로 PGUSER=postgres / PGPASSWORD=postgres 환경에서 kratos / hydra 를 띄워 (4433/4444 LISTEN, pid 27136 / 17464) 다음을 확인.

| 단계 | 기대 | 결과 |
| --- | --- | --- |
| dev-up 출력 | "external instance detected on port 4433; using existing kratos" + 4444 동일 | ✅ |
| `.pids/` 내용 | backend.pid + frontend.pid 만 (kratos/hydra 없음) | ✅ |
| 헬스 체크 | backend `/health` 200, frontend `/` 200 | ✅ |
| dev-down 출력 | "hydra not started by this script; leaving any listener on port(s) 4444, 4445 intact" + kratos 동일. backend/frontend sweep. | ✅ |
| 종료 후 포트 | 3000/8080 free, 4433/4434/4444/4445 여전히 LISTEN pid 27136 / 17464 | ✅ |

이후 테스트용 kratos/hydra 는 수동으로 정리 (Stop-Process). 최종 상태: 6 포트 모두 free, `.pids/` 비움, log 잔재 파일 모두 제거.

## 다음 세션 출발점 후보

1. **이번 슬롯 작업물 commit + PR** — dev-up.ps1 + dev-down.ps1 + dev-up.sh + dev-down.sh (+ 본 메모리 파일 3개). 단일 commit / 단일 PR 권장. 픽스 범위가 #1 + #3 두 결함에 걸쳐 있고 모두 "외부 환경/서비스 존중" 이라는 한 가지 의도로 묶이므로 자연스러움.
2. **결함 #4 픽스** — `infra/idp/kratos.yaml` DSN 에 credential 명시 또는 dev-up env 주입. 그 다음 진짜 cold-start (Scenario B) 검증.
3. **`/healthz` → `/health` 표기 정정** — `CLAUDE.md` 와 `dev-up.{ps1,sh}` 최종 출력 라인 (`http://localhost:8080/health` 표기는 일관되나, CLAUDE.md "프로젝트 실행 기본값" 섹션의 `curl http://localhost:8080/healthz` 가 틀림)
4. **결함 #2 픽스 평가** — 빌드 산출물 실행 전환 vs grandchild 추적 방식 trade-off 검토.

## 환경 정리 상태 (현 시점)

- 모든 포트 free (3000, 8080, 4433, 4434, 4444, 4445)
- `.pids/` empty
- log 잔재 파일 모두 제거 (kratos.log, kratos.log.err, hydra.log*, backend.log*, frontend.log*, kratos-ext.log*, hydra-ext.log*)
- 작업 트리 변경: dev-up.ps1 / dev-down.ps1 / dev-up.sh / dev-down.sh + 본 메모리 슬롯 3 파일
