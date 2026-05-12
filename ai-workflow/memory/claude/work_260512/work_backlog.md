# 작업 백로그

`claude/work_260512` 슬롯. 직전 슬롯 `claude/260512` handoff next-step #1 (dev-up 실머신 검증) 을 수행하면서 결함을 식별하고, 결함 #1 + #3 까지 픽스해 4개 파일 수정. 커밋 + PR 직전 상태.

## [Planned]
1. healthz 경로 표기 정정 — `CLAUDE.md` 와 dev-up 스크립트 헤더에서 `/healthz` → `/health` (backend-core 는 `/health` 만 응답하고 `/healthz` 는 404)
2. 결함 #2 (PID grandchild) 픽스 평가 — `go run .` 대신 빌드 산출물 실행, 또는 grandchild PID 까지 추적해 launcher PID 기반 정리가 의미를 갖게
3. 결함 #4 (cold-start kratos DSN) — `infra/idp/kratos.yaml` 의 DSN `postgres://localhost:5432/devhub?...` 에 user/password 없어 libpq 가 OS 유저로 폴백 → SASL 인증 실패. PGUSER/PGPASSWORD env 주입, 또는 DSN 에 명시 credential, 또는 dev-up 에서 env 자동 export 중 하나.

## [In Progress]
(없음)

## [Done — 이번 세션]

### dev-up 1차 실행 + 결함 #1 박제·픽스
- [x] `dev-up.ps1` 1차 실행 — frontend launch 직전 `Start-Process` 가 `npm.ps1` 을 Win32 exe 가 아니라며 거부, 재현 확인
- [x] 결함 #1 픽스: `Start-Process -FilePath 'npm'` → `'npm.cmd'` (3줄)
- [x] 재실행 — 5/5 서비스 ready, backend `/health` 200 `{"db":"ok","service":"backend-core","status":"ok"}`, frontend `/` 200 (23KB)
- [x] `dev-down.ps1` — 6/6 포트 free, `.pids/` 비움

### 결함 #2, #3 식별
- [x] #2: `go run .` 와 `npm.cmd` 가 grandchild 로 포트 점유. dev-down 의 PID-kill 은 launcher 만 죽임. port-sweep 이 실질적 종료 경로.
- [x] #3: dev-up 이 포트 점유 체크 없이 `Get-Command kratos/hydra` 만으로 항상 spawn → user-managed 인스턴스 있을 때 새 instance 는 bind fail zombie 가 되고, dev-down 의 port-sweep 이 user-managed 인스턴스까지 죽임.

### 결함 #3 픽스 (이번 슬롯 완료)
- [x] dev-up.ps1: `Test-PortListening` 헬퍼 + kratos/hydra/backend/frontend 4개 서비스 모두 메인 포트 LISTEN 시 spawn / Wait-ForPort / PID 파일 작성 모두 skip
- [x] dev-down.ps1: `Stop-ServiceByPid` 가 PID 파일 존재 여부 bool 반환. `$servicePorts` ordered map 으로 서비스→포트 매핑. PID 파일 있는 서비스만 sweep 대상 누적, 없으면 "leaving any listener on port(s) X intact" 로그
- [x] dev-up.sh, dev-down.sh: 동일 의미론을 bash 로 (is_port_listening 헬퍼, associative array 기반 sweep 누적)
- [x] `.ps1` 두 파일 non-ASCII 스캔 통과 (ASCII-only 정책 유지)

### Scenario A 검증 (external IdP 보존)
- [x] kratos/hydra 를 PGUSER=postgres / PGPASSWORD=postgres 수동 기동 (4433/4444 LISTEN)
- [x] dev-up.ps1 → "external instance detected on port 4433/4444; using existing" 메시지 정확히 출력
- [x] backend/frontend 만 spawn, `.pids/` 에 backend.pid + frontend.pid 만 작성 (kratos/hydra 미작성)
- [x] backend `/health` 200, frontend `/` 200
- [x] dev-down.ps1 → "hydra/kratos not started by this script; leaving any listener on port(s) … intact" 출력, 4433/4434/4444/4445 sweep 생략, 3000/8080 만 sweep
- [x] dev-down 후에도 kratos pid=27136 / hydra pid=17464 살아남음 확인 (테스트 후 수동 정리)

### Scenario B 검증 (clean cold-start) — 부분 완료
- [x] backend/frontend 콜드 부팅 정상 동작 (Scenario A 의 backend/frontend spawn 으로 사실상 검증)
- [ ] kratos cold-start — 결함 #4 (kratos.yaml DSN credential 누락) 로 인해 본 슬롯에서 미완. Planned #3 로 이월.
