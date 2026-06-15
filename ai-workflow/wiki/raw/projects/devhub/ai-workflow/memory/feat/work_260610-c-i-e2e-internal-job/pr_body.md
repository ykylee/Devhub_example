# ci(workflow): E2E Internal (real Keycloak adapter) job — DEVHUB_BUILD_TIER=internal matrix 1쌍 (sprint -a follow-up PR1 PR #540 의 carry-over C-i)

## 목적

sprint -a follow-up PR1 (PR #540, `feat/work_260610-v1-1-sprint-a-real-adapter`, main `58d163f`) 의 carry-over **C-i** (P2) 의 정공법 PR. **e2e shard 1/2/3 (saovae_stub default) + 별도 e2e-internal job 1개 (`DEVHUB_BUILD_TIER=internal`)** 의 CI matrix 1쌍 정합.

- C-i (P2): E2E saovae_stub + real adapter CI matrix (ADR-0030 §2.3 runtime injection + sprint -a follow-up PR1 carry-over)

## 변경 요약 (2 file)

### 1. `.github/workflows/ci.yml` (+202 lines)

신규 job `e2e-internal` 추가. 기존 `e2e` job 의 변형으로, **`DEVHUB_BUILD_TIER=internal` 만 env block 에 추가** + Keycloak container 를 별도 port `8181` 로 시작 (e2e shard 의 port 8180 과 충돌 회피) + Playwright shard `1/1` 단일 (logout flow 1 e2e suite 가 real adapter wire 검증).

- `e2e-internal` job = 23 step:
  - PG 15 native (e2e 와 동일, port 5432 공유)
  - Go + Node setup (e2e 와 동일)
  - Keycloak container (port `8181` — e2e 의 port 8180 과 분리)
  - Apply app migrations (e2e 와 동일)
  - Validate E2E-CI Sync Contract (e2e 와 동일)
  - **Start Backend (DEVHUB_BUILD_TIER=internal — real Keycloak adapter)** ← 본 PR 의 핵심
  - Start Frontend (e2e 와 동일)
  - Wait for App Readiness (e2e 와 동일)
  - **Run E2E Tests (real adapter wire) — DEVHUB_BUILD_TIER=internal env 주입** ← 본 PR 의 핵심
  - Upload Playwright Report + Upload Logs on Failure

e2e shard 1/2/3 (default) 의 env block, start command, test invocation 모두 변경 0. **saovae_stub path 그대로 유지**.

### 2. `scripts/ci-e2e-sync-check.sh` (+5 lines, comment)

`required_e2e_tokens` 는 변경 없음. 본 PR 의 의도 명시 comment 만 추가:
- `DEVHUB_BUILD_TIER` 는 **의도적으로** required_e2e_tokens 에 미포함
- e2e shard 1/2/3 (saovae_stub default) env block 에는 미설정
- e2e-internal job (DEVHUB_BUILD_TIER=internal) 만 별도 token 노출
- e2e-internal 의 DEVHUB_BUILD_TIER env 정합은 **actionlint + 실제 e2e run 이 검증**

## trade-off 결정 (옵션 A 권장)

research artifact: `ai-workflow/memory/feat/work_260610-c-i-e2e-internal-job/research-c-i-scope.md` (option 4건 비교)

- **옵션 A (권장)**: e2e shard 1/2/3 (default saovae_stub) + e2e-internal job 1개 (`internal`) = 1쌍
  - CI runtime: e2e shard × 1 (e2e-internal) = +15~20min
  - Keycloak infra: e2e 의 container + e2e-internal 의 container = 2 container (port 8180 + 8181 분리)
  - 코드 영향 0, ci.yml + script 2 file
  - **본 PR 의 정공법**

- 옵션 B: shard 1/2/3 only + unit test cover → real HTTP round-trip 검증 누락 (production 회귀 catch 안됨) — 거부
- 옵션 C: 6 matrix (default/internal × shard 1/2/3) → CI runtime 2배, logout port 만 차이 — 거부
- 옵션 D: e2e shard 를 default 1/2/3 → internal 1/2/3 로 matrix 화 → logout flow 가 shard 모두 동일 = 중복 — 거부

## scope 결정

### 변경 file (2)
- `.github/workflows/ci.yml` — `e2e-internal` job 신규 (+202 lines)
- `scripts/ci-e2e-sync-check.sh` — DEVHUB_BUILD_TIER 의도적 미포함 comment 추가 (+5 lines)

### 변경 안 함
- backend code: 0 변경 (real adapter wire 가 main PR #540 머지로 이미 정합)
- frontend code: 0 변경 (frontend e2e env 의 `DEVHUB_BUILD_TIER` 추가 불요 — backend build tier 가 frontend e2e logic 무관, frontend 의 logout flow 가 backend 의 logout endpoint 만 호출, real vs stub 결정은 backend)
- e2e spec: 0 변경 (동일 e2e spec 가 real adapter wire 에서도 동작 검증)
- docs: 0 변경
- Makefile: 0 변경
- ADR: 0 변경 (ADR-0030 §2.3 의 runtime injection 결정 + 1.1b 의 single binary 가 본 PR 의 matrix 로 정합)
- migration: 0 변경

## Tier 분류 (self-check)

| 변경 영역 | Tier | 근거 |
| --- | --- | --- |
| `.github/workflows/ci.yml` | **공용** | GitHub-hosted CI workflow, 사내 한정 정보 미포함 |
| `scripts/ci-e2e-sync-check.sh` | **공용** | 검증 script, 사내 한정 정보 미포함 |

**본 PR 의 모든 변경 = 공용**. `check-tier-separation.sh` no changes between origin/main and HEAD 확인.

## 검증 (run on this branch)

- `bash scripts/check-tier-separation.sh` — ✅ no changes between origin/main and HEAD
- `bash scripts/check-openapi-yaml-lint.sh` — ✅ passed (openapi.yaml 변경 0)
- `bash scripts/check-migration-uniqueness.sh` — ✅ valid and unique (migration 변경 0)
- `bash scripts/ci-e2e-sync-check.sh` — ✅ E2E-CI sync contract check passed
- Python YAML parse: ✅ jobs = 10 (changed-paths, workflow-lint, migration-prefix-lint, openapi-yaml-lint, backend-unit, backend-integration, frontend-unit, e2e-build, e2e, **e2e-internal**), e2e-internal 23 step + DEVHUB_BUILD_TIER env 1+ 곳
- GitHub Actions run (PR 머지 후): workflow-lint job 의 actionlint 통과 + e2e-internal job 의 실 실행 PASS (이 PR 머지 후 1차 PR 의 CI 가 본 PR 의 effect 검증)

## 의도적 trade-off

- **e2e-internal 의 Keycloak container 를 port 8181 로 분리** — e2e shard 의 port 8180 와의 충돌 회피. e2e-internal 단독 실행 + e2e shard 와 동시 실행 모두 가능. e2e shard 와 `e2e-internal` 의 `if` 조건이 동일 (`e2e == 'true' || workflow == 'true'`) 이므로 동시 trigger 가능.
- **DEVHUB_BUILD_TIER token 을 required_e2e_tokens 에 미포함** — script 의 contract 가 "e2e helper 가 사용" 의미가 아닌 "ci.yml + e2e helper 양쪽에 token 존재" 의미. e2e shard 1/2/3 (saovae_stub) 의 env block 에 DEVHUB_BUILD_TIER 추가 시 본 PR 의 의도 (saovae_stub default 유지) 와 위배. e2e-internal 의 DEVHUB_BUILD_TIER env 정합은 **actionlint (YAML structure) + 실제 e2e run (backend 가 DEVHUB_BUILD_TIER=internal 인지 log 확인) 이 검증**.
- **Playwright shard 1/1** — logout flow 의 e2e suite 가 1 shard 에 다 들어감. logout flow 외 다른 e2e suite 는 backend 의 build tier 무관 (auth, RBAC, application CRUD 등 모두 backend API 만 호출, backend 가 real vs stub 이 무관). 따라서 **1 shard (≈ 4-5min) 만으로 충분**.

## Out of scope (별도 PR carry-over)

- **C-j (P3)**: build tag 정책 재검토 — runtime injection (현재) ↔ build tag 전환 trade-off. 본 PR 의 C-i matrix 와는 별개.
- **backend-integration DEVHUB_BUILD_TIER** — backend-integration job (현재 && false 로 비활성화) 의 DEVHUB_BUILD_TIER matrix. 필요 시 별도 PR.
- **E2E spec tier-aware skip** — 동일 spec 그대로 real wire 에서 검증. e2e spec 변경 불요.
- **ADR-0030 §5 timeline + traceability report.md** — 본 PR 의 scope 외. C-h PR #541 에서 1.1a + 1.1b accepted/done 정합 완료.
- **release_v1_roadmap.md §3.5 N-13** — 본 PR 의 C-i 의 본 release_v1_roadmap row N-13 정합은 별도 housekeeping commit 으로 진행.

## Refs

- Sprint -a follow-up PR1 PR #540 (`feat/work_260610-v1-1-sprint-a-real-adapter`, main `58d163f`) — real adapter + v1.0 mirror struct 제거. 본 PR 의 carry-over source.
- Sprint -a follow-up 본 PR #539 (`feat/work_260610-v1-1-sprint-a-followup`, main `87e6c1f5`) — saovae_stub + main.go DEVHUB_BUILD_TIER env var 분기 + view/ deprecation. 본 PR 의 stub path 가 main 에 머지.
- Sprint -a 본 PR #538 (`feat/work_260610-v1-1-sprint-a-sso-integrations`, main `20b4bb3`) — port interface canonical 위치 = `domain/auth-session/integration/`.
- [ADR-0030 §2.1 port interface canonical 위치 결정](../adr/0030-sso-integrations-and-auth-session-port.md) — option 2 = `domain/auth-session/integration/`
- [ADR-0030 §2.2 real adapter 분리 결정](../adr/0030-sso-integrations-and-auth-session-port.md) — option 2 = `sso-integrations/keycloak/`
- [ADR-0030 §2.3 runtime injection 결정](../adr/0030-sso-integrations-and-auth-session-port.md) — option 2 = `DEVHUB_BUILD_TIER` env var, build tag 미사용. **본 PR 의 matrix 가 본 결정의 code-level 적용**.
- [ADR-0030 §5 timeline](../../adr/0030-sso-integrations-and-auth-session-port.md) — 1.1a + 1.1b status = accepted/done (PR #541 머지로 정합)
- [release_v1_roadmap.md §3.5 N-13](../planning/release_v1_roadmap.md) — v1.1 sprint -a follow-up carry-over N-13

## Base / target

- **base branch**: `main` (HEAD `88681f4`, PR #541 머지 후)
- **target branch**: `main`
- **merge strategy**: squash
- **branch name**: `feat/work_260610-c-i-e2e-internal-job`
