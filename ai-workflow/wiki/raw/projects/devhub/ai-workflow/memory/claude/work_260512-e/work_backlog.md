# 작업 백로그

`claude/work_260512-e` 슬롯 — 직전 슬롯 work_260512-d (PR #72) Planned #1 처리 (OIDC client 등록 자동화).

## [Planned]
(없음 — 라인 close)

## [In Progress]
(없음)

## [Done — 후속 (라인 closure)]
- [x] `gemini/frontend_260510` → `main` 머지 (PR #74, rebase, 14 commits 선형 추가, main HEAD `58f5b03`)
- [x] `gemini/frontend_260510` 브랜치 로컬·원격 모두 삭제 (사용자 합의)
- [x] 라인 close 박제 (`backlog/2026-05-12-close.md`)

## [Carried to next session]
- (선택) Fresh DB 진짜 cold-start 검증 — schema drop → dev-up 단일 실행. 사용자 DB 영향 큼.
- 결함 #2 의 frontend 부분 픽스 평가 (지금은 documented behavior).
- dev-up ergonomics 개선 (부분 시작 등) — 필요해진 시점이 오면.

## [Done — 이번 세션]

### 분석
- [x] `register-devhub-client.ps1` 내용 파악 — Hydra Admin REST API DELETE-then-POST 멱등, hydra v26 의 CLI `--client-id` 무력화 회피책.
- [x] sh 평행 필요 여부 결정 — 필요 (PS-only 상태).

### 구현
- [x] `infra/idp/scripts/register-devhub-client.sh` 신규 (bash + curl, jq 무의존, `set -euo pipefail`).
- [x] git index 에 100755 mode 등록 (unix clone 에서 바로 실행 가능).
- [x] dev-up.ps1 단계 3b — hydra Wait-ForPort 직후, 4445 reachable 체크, `DEVHUB_SKIP_OIDC_REGISTER=1` 우회. **PS-to-PS call 의 $LASTEXITCODE 미갱신 함정 발견 → inner 스크립트의 ErrorAction Stop 이 failure 전파를 보장하므로 LASTEXITCODE 체크 제거.**
- [x] dev-up.sh 단계 3b — `is_port_listening 4445` + sh 등록 스크립트 호출.
- [x] infra/idp/README.md §5 자동화 callout + bash 변형 명령 추가.

### 실머신 검증
- [x] 사전: 6 포트 free
- [x] dev-up.ps1 with skips: kratos/hydra cold spawn, "Registering OIDC client (devhub-frontend)..." 출력, DELETE existing + POST new, 5/5 서비스 ready
- [x] GET /admin/clients/devhub-frontend → 200, client_id 정확, redirect_uris / scope 정확
- [x] dev-down: 4 PID 정지 + 3000 sweep (결함 #2 픽스 무회귀)
- [x] 종료 후 6 포트 free, `.pids/` 비움

## [Carried over]
- 직전 슬롯 Planned #2 (main 머지 전략) — 이월 그대로 (이 슬롯에서도 처리 안 함)
