# 작업 백로그

`claude/work_260512-b` 슬롯 — 직전 슬롯 `claude/work_260512` (PR #69) 의 Planned #2, #3 처리.

## [Planned]
1. 결함 #2 평가 — 범위가 backend/frontend launcher 로 축소되었음 (보너스 발견). `go run .` → 빌드 산출물 실행으로 grandchild 제거할지, 현 PID-kill + port-sweep 조합을 그대로 둘지 결정.
2. kratos / hydra `migrate sql` 자동화 평가 — fresh DB cold-start 의 마지막 수동 단계 (README §2). dev-up 통합 vs README 가이드 유지.

## [In Progress]
(없음)

## [Done — 이번 세션]

### 결함 #4 픽스 (Ory DSN env override)
- [x] dev-up.ps1: `Get-IdpDsn` 헬퍼 + kratos / hydra spawn 직전 `$env:DSN` 주입, hydra 후 `Remove-Item Env:DSN`
- [x] dev-up.sh: `idp_dsn` 헬퍼 + `export DSN=...; run_service ...; unset DSN` 패턴 (bash 의 `VAR=val func` 가 subshell 까지 안전하게 전파되는지 버전·구조 의존이라 명시적 export 채택)
- [x] credential 평문은 repo 밖 유지 — yaml `dsn` 은 그대로 (search_path 만), DB_URL 에 들어 있는 credential 을 dev-up 가 spawn 시점에만 주입

### healthz 표기 정정 (후보 #2)
- [x] `docs/setup/environment-setup.md:96` `curl http://localhost:8080/healthz` → `/health` (tracked, commit)
- [x] 루트 `CLAUDE.md:52` — `.gitignore:11` 에 의해 무시되는 개인 로컬 파일임을 확인. 로컬에만 정정 적용, commit 대상 아님. 메모리에 사실 박제.

### Scenario B 콜드부팅 검증
- [x] 사전: 6 포트 free, `.pids/` 없음
- [x] dev-up.ps1: 5/5 서비스 ready
- [x] PID 파일: kratos / hydra / backend / frontend 4개 모두 작성
- [x] 헬스: backend `/health` 200, frontend `/` 200, kratos `/health/ready` 200, hydra `/health/ready` 200
- [x] dev-down.ps1: 4 PID 정지 + 3000 / 8080 port-sweep (kratos/hydra 는 PID-kill 한 번에 종료 — 결함 #2 범위 축소 발견)
- [x] 종료 후: 6 포트 free, `.pids/` 비움

## [Carried over — 이전 슬롯에서 미해결]
- 결함 #2 — Planned #1 로 이월 (범위 축소된 상태)
