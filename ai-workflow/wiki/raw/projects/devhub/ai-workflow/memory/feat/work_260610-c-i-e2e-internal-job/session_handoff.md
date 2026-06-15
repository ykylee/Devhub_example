- 문서 목적: 본 sprint 의 작업 상태 + 핵심 사실 (e2e-internal job scope + 4 file 변경 0 + ci.yml +202 / script +5) + 다음 세션 directive.
- sprint branch: feat/work_260610-c-i-e2e-internal-job
- 대상 독자: 후속 에이전트, PR reviewer, owner
- 상태: in_progress (사용자 confirm 후 commit + push + PR 발행 대기)
- 최종 수정일: 2026-06-10
- 관련 문서: work_backlog.md, backlog/2026-06-10.md, state.json, pr_body.md, [sprint -a follow-up real-adapter session_handoff](https://github.com/ykylee/Devhub_example/blob/main/ai-workflow/memory/feat/work_260610-v1-1-sprint-a-real-adapter/session_handoff.md), [ADR-0030 §2.3 runtime injection 결정](../../../../docs/adr/0030-sso-integrations-and-auth-session-port.md)

## 1. sprint 목표 (in_progress — commit + PR 발행 대기)

sprint -a follow-up PR1 (PR #540, main `58d163f`) 의 carry-over **C-i (P2)** 의 정공법 PR. **코드 0줄 변경** (CI workflow + script 만).

sprint scope (사용자 결정 2026-06-10):
- **C-i (P2)**: E2E saovae_stub + real adapter CI matrix — DEVHUB_BUILD_TIER=internal env var + e2e shard 양쪽 정합
- **(OUT of scope)** C-j (build tag 정책 재검토, P3) / backend-integration DEVHUB_BUILD_TIER matrix / E2E spec tier-aware skip

## 2. 사용자 결정 사항 (in-session)

- **옵션 A 채택**: e2e shard 1/2/3 (saovae_stub default) + e2e-internal job 1개 (`DEVHUB_BUILD_TIER=internal`). 옵션 B (unit test cover only), C (6 matrix), D (matrix shard 1/2/3 × 2) 모두 거부.
- **Keycloak container port**: 8181 (e2e shard 의 port 8180 과 분리) — e2e shard 와 e2e-internal 동시 trigger 가능.
- **Playwright shard**: 1/1 (단일 shard, logout flow 가 real adapter wire 검증) — shard 분할 불요.
- **DEVHUB_BUILD_TIER token**: required_e2e_tokens 에 의도적 미포함 (saovae_stub default 의 env block 미설정 유지). e2e-internal 의 DEVHUB_BUILD_TIER env 정합은 actionlint + 실제 e2e run 이 검증.
- **research artifact**: `ai-workflow/memory/feat/work_260610-c-i-e2e-internal-job/research-c-i-scope.md` (artifact write EPERM 으로 inline emit only)

## 3. 완료된 작업

### 3.1 `.github/workflows/ci.yml` (e2e-internal job 신규, +202 lines) ✓

신규 job `e2e-internal` 추가. 기존 `e2e` job 의 변형:
- **PG 15 native** (e2e 와 동일 pattern, port 5432)
- **Keycloak container port 8181** (e2e 의 8180 과 분리)
- **Apply app migrations** (e2e 와 동일)
- **Validate E2E-CI Sync Contract** (e2e 와 동일)
- **Start Backend (DEVHUB_BUILD_TIER=internal — real Keycloak adapter)** — 본 PR 의 핵심
- **Start Frontend** (e2e 와 동일)
- **Wait for App Readiness** (e2e 와 동일)
- **Run E2E Tests (real adapter wire)** — Playwright shard 1/1, `DEVHUB_BUILD_TIER=internal` env
- **Upload Playwright Report** (e2e 와 동일)
- **Upload Logs on Failure** (e2e 와 동일)

총 23 step. e2e shard 1/2/3 (default) 의 env block, start command, test invocation 모두 변경 0.

### 3.2 `scripts/ci-e2e-sync-check.sh` (+5 lines comment) ✓

DEVHUB_BUILD_TIER token 의 의도적 미포함 + rationale comment 4 line. contract check 본체 변경 0.

## 4. 잔여 / 후속 작업

### 4.1 본 PR 잔여
- **사용자 confirm 후** commit + push + gh pr create + gh pr merge --squash --delete-branch.
- 본 PR 머지 후 본 PR 의 e2e-internal job 의 1차 실 실행 (CI 7/7 + e2e-internal 1/1 PASS) — workflow-lint + e2e shard 1/2/3 + e2e-internal 1/1 모두 PASS 확인.

### 4.2 carry-over (별도 PR)
- **C-j (P3)**: build tag 정책 재검토 — runtime injection (현재) ↔ build tag 전환 trade-off.
- **backend-integration DEVHUB_BUILD_TIER matrix** (backend-integration job 이 현재 && false 로 비활성화) — 필요 시 별도 PR.
- **release_v1_roadmap.md §3.5 N-13** 정합 — 본 PR 머지 후 별도 housekeeping commit.

## 5. 핵심 파일 / 라인 참조 (본 PR 시작 시점)

- `.github/workflows/ci.yml:553-755` (e2e-internal job 신규)
- `scripts/ci-e2e-sync-check.sh:21-32` (DEVHUB_BUILD_TIER comment)
- `ai-workflow/memory/feat/work_260610-c-i-e2e-internal-job/{state.json,session_handoff.md,work_backlog.md,backlog/2026-06-10.md,pr_body.md}` (sprint memory 5종)

## 6. 알아둘 trade-off (의도적 결정)

- **옵션 A 채택 (e2e shard 1/2/3 + e2e-internal 1개)**: research artifact 의 trade-off 표 기준, **CI runtime +15~20min** vs **real HTTP round-trip 검증 보장** 의 최적 trade-off. 옵션 B (real 검증 누락) / C (CI 2배) / D (중복 shard) 모두 거부.
- **Keycloak container port 8181**: e2e shard 의 8180 과 동시 trigger 가능. CI 가 둘 다 동시 실행되어도 port 충돌 0. **Keycloak infra 의존성 = 2 container (e2e + e2e-internal)**.
- **Playwright shard 1/1**: logout flow 외 다른 e2e suite 는 backend 의 build tier 무관 (auth, RBAC, application CRUD 등 모두 backend API 만 호출). **1 shard (≈ 4-5min) 만으로 충분**.
- **DEVHUB_BUILD_TIER token 미포함**: script 의 contract 가 "e2e helper 가 사용" 의미가 아닌 "ci.yml + e2e helper 양쪽에 token 존재" 의미. e2e shard 1/2/3 (saovae_stub) env block 에 DEVHUB_BUILD_TIER 추가 시 본 PR 의 의도 (saovae_stub default 유지) 와 위배. e2e-internal 의 DEVHUB_BUILD_TIER env 정합은 **actionlint (YAML structure) + 실제 e2e run (backend log 에서 "DEVHUB_BUILD_TIER=internal" 확인) 이 검증**.

## 7. 다음 세션이 가장 먼저 할 일

1. `git status` / `git log --oneline -3` / `git branch --show-current` 확인 (현재 `feat/work_260610-c-i-e2e-internal-job`).
2. **사용자 confirm 후** `git add . + git commit + git push + gh pr create --body-file pr_body.md`. PR body 의 "추적성 영향" 섹션에 변경 file 2개 + e2e-internal job 23 step + Keycloak container port 8181 + DEVHUB_BUILD_TIER=internal env 명시.
3. main flat memory 4종 동기화 (정합): `ai-workflow/memory/state.json` head_commit 갱신 + `session_handoff.md` post-PR-merge row + `work_backlog.md` §5 변경 이력 row.
4. 또는 다른 sprint 진입 (C-j build tag / N-10 RBAC E2E 6 TC / release_v1_roadmap.md 갱신).
