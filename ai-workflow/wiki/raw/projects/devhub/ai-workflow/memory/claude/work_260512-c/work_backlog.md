# 작업 백로그

`claude/work_260512-c` 슬롯 — 직전 슬롯 `claude/work_260512-b` (PR #70) 의 Planned #1 (결함 #2 평가) 처리.

## [Planned]
1. kratos / hydra `migrate sql` 자동화 평가 — fresh DB cold-start 시나리오의 마지막 수동 단계 (README §2). dev-up 통합 vs README 가이드 유지 결정.
2. `gemini/frontend_260510` → `main` 머지 전략 검토 — 이 feature branch 위에 PR #63-#70 + 본 PR (총 6+ 건) 이 누적. 단일 squash 머지 vs 개별 cherry-pick.

## [In Progress]
(없음)

## [Done — 이번 세션]

### 결함 #2 평가 및 픽스
- [x] backend (`go run .`): launcher 가 grandchild 로 server 를 띄워 PID 불일치 → `go build -o dev-bin/backend-core(.exe)` 후 직접 실행으로 픽스. PID == port owner 일치.
- [x] frontend (`npm.cmd run dev`): npm 컨벤션 우회의 ergonomic 비용이 커서 유지. port-sweep 안전망이 정확히 이 케이스 용 documented behavior 로 close.

### 스크립트 변경
- [x] dev-up.ps1: backend 블록에서 `go build` → `dev-bin/backend-core.exe` 실행. `Start-BackgroundService` 헬퍼가 빈 ArgumentList 거부 문제를 conditional 로 픽스 (Arguments.Count > 0 일 때만 ArgumentList splat 포함).
- [x] dev-up.sh: 동일 의미론을 bash 로 — `( cd $REPO_ROOT/backend-core && go build -o "$backend_bin" . )` 후 binary 직접 실행. backend-core 가 자체 go.mod 라 빌드는 해당 디렉터리 안에서.
- [x] .gitignore: `dev-bin/` 추가 (.pids/ 다음 줄).

### Scenario C 검증
- [x] cold-start round-trip 통과: 5/5 ready
- [x] `backend.pid == 8080 owner` 일치 (25984 == 25984) ✓
- [x] backend `/health` 200
- [x] dev-down 시 8080 sweep 메시지 없음 (PID-kill 만으로 종료)
- [x] frontend 의 3000 sweep 은 documented 동작으로 박제 (npm.cmd grandchild)
- [x] 종료 후 6 포트 free, `.pids/` 비움
- [x] `dev-bin/backend-core.exe` (38MB) gitignore 검증

## [Carried over]
- 직전 슬롯 Planned #2 (kratos / hydra migrate sql 자동화) — 본 슬롯 Planned #1 으로 이월
