# 세션 인계 문서 (2026-05-12 work continuation slot — defect #2)

## 세션 목표

직전 슬롯 `claude/work_260512-b` 의 Planned #1 — 결함 #2 (PID grandchild drift) 평가 + 픽스. 직전 슬롯의 보너스 발견(`kratos/hydra` 는 PID-kill 만으로 종료) 으로 범위가 backend / frontend launcher 로 축소된 상태에서 시작.

## 평가 결과

| 컴포넌트 | 평가 | 결정 |
| --- | --- | --- |
| backend (`go run .`) | launcher 가 grandchild 로 컴파일된 binary 실행 → PID 불일치. 빌드 산출물을 직접 실행하면 해소. | **픽스 — `go build -o dev-bin/backend-core` + binary 실행** |
| frontend (`npm.cmd run dev`) | `npm.cmd` 는 cmd → node → next 의 wrapper chain. 우회하려면 `node node_modules/next/dist/bin/next dev` 등 npm 컨벤션을 벗어남. ergonomic 비용이 이득보다 큼. | **유지 — port-sweep 안전망이 정확히 이 케이스용 documented behavior** |

## 픽스 내용

### 1. `dev-up.ps1` backend 블록
- `Get-Command go` 의존성은 그대로 (이미 backend 가 go 가 필요했음).
- `dev-bin/` 디렉터리 생성 → `go build -o dev-bin/backend-core.exe .` (backend-core/ 안에서 실행. root go.mod 가 없고 backend-core 가 자체 모듈).
- 빌드 산출물 직접 `Start-BackgroundService` → PID 가 진짜 server PID.
- 빌드 실패 시 `$ErrorActionPreference=Stop` 으로 즉시 abort.

### 2. `dev-up.sh` backend 블록
- bash 평행 구현: `mkdir -p $REPO_ROOT/dev-bin`, `( cd $REPO_ROOT/backend-core && go build -o ... )`, `run_service "backend" "$backend_bin" ...`.
- bash 의 `set -euo pipefail` 이 빌드 실패 시 abort 보장.

### 3. `Start-BackgroundService` 헬퍼 픽스
- `Start-Process -ArgumentList @()` (빈 array) 가 PowerShell parameter validation 으로 거부됨 — "element of the argument collection contains a null value".
- 새 backend 호출은 인자가 없음 → 헬퍼가 `Arguments.Count -gt 0` 일 때만 ArgumentList 를 splat 에 포함하도록 조건부.

### 4. `.gitignore`
- `.pids/` 다음 줄에 `dev-bin/` 추가. 빌드 산출물 (~38MB) 추적 회피.

## 실머신 검증 (Scenario C)

1. **사전 정리** — `dev-down` 으로 직전 잔재 (kratos/hydra) 정리. 6 포트 free.
2. **cold-start** — `dev-up.ps1`:
   - kratos / hydra cold spawn → 4 포트 ready
   - "Compiling backend..." → `go build` → "Starting backend..." → "backend ready on port 8080"
   - frontend ready on 3000
3. **PID 일치성 검증**:
   - `backend.pid = 25984` == `8080 owner pid = 25984` ✓
   - `frontend.pid = 16000` ≠ `3000 owner pid = 18784` (예상된 mismatch, npm.cmd 패턴)
   - `kratos.pid` / `hydra.pid` == port owner (직접 binary)
4. **healthz** — backend `/health` 200 `{"db":"ok","service":"backend-core","status":"ok"}`.
5. **dev-down**:
   ```
   frontend stopped (PID 16000)
   backend stopped (PID 25984)
   hydra stopped (PID 28788)
   kratos stopped (PID 30848)
   swept PID 18784 on port 3000 (node)
   ```
   - 8080 sweep 없음 — backend PID-kill 만으로 종료 ✓
   - 3000 sweep 한 줄만 (frontend grandchild orphan) — documented 동작
6. **종료 후** — 6 포트 free, `.pids/` 비움.

## 부수 개선 검증

빌드 시간 비교 (단순 측정):
- `go run .` cold: 매 호출마다 컴파일 (~5-10s)
- `go build -o ... .` cold: 첫 빌드 ~10s, 캐시 후 ~1-3s
실질적 ergonomic 영향 없음. dev-up 한 번 실행 후 dev-down/dev-up 사이클에서는 두 번째부터 incremental build 로 빠름.

## 다음 슬롯 출발점 후보

1. **이번 묶음 commit + PR + 본인 리뷰 모드 머지** — `.gitignore`, `dev-up.ps1`, `dev-up.sh` + 본 메모리 3 파일. base=`gemini/frontend_260510`.
2. **kratos / hydra `migrate sql` 자동화 평가** — 직전 슬롯 backlog Planned #2 그대로 이월. fresh DB 시나리오의 마지막 수동 단계.
3. **gemini/frontend_260510 → main 머지 전략** — 이 라인이 PR #63-#70 + 본 PR 이 누적된 상태. main 으로 통합할 시점이 다가오고 있음. 단일 squash merge vs PR 별 cherry-pick 검토.

## 환경 정리 상태 (현 시점)

- 6 포트 모두 free
- `.pids/` 비움
- log 잔재 파일 모두 제거
- `dev-bin/backend-core.exe` (38MB) — gitignore 됨, commit 안 됨, 다음 dev-up 에서 재사용 가능
- 작업 트리: `.gitignore` / `dev-up.ps1` / `dev-up.sh` + 본 슬롯 3 파일
