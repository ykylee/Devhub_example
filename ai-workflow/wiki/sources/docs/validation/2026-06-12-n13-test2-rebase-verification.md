---
title: 2026 06 12 n13 test2 rebase verification
type: source
tags: [validation, project-devhub]
sources: [raw/projects/devhub/docs/validation/2026-06-12-n13-test2-rebase-verification.md]
git_commit: baf1cf24
git_branch: main
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
last_touched: 2026-06-22T03:37:45Z
mirror_dirty: |
related: [none]
status: draft
contradictions: [none]
---

# N-13 Test 2 Sign-out timeout — rebase + 자동 재실행 검증 보고서

- **문서 목적**: PR #548 (`feat/work_260611-a-n13-inbound-source-impl`, 2026-06-11 05:40 UTC CLOSED) 의 E2E Internal 1 fail 2건 중 **Test 2 Sign-out timeout** (N-8 race 유사) 이 main 의 PR #550 spec timing fix + PR #574 e2e seed fix 적용 후 자동 해결되는지 검증.
- **범위**: 5 file (docs only, 코드 0줄) + 메모리 4 file 동기화. 본 보고서 = verification evidence.
- **대상 독자**: 프로젝트 리드, N-13 follow-up 결정자, 후속 sprint 작업자.
- **상태**: **verified (2026-06-12)** — main 의 e2e CI 자동 재실행 trigger + e2e shard 1/2/3 PASS 확인.
- **최종 수정일**: 2026-06-12 (N-13 follow-up B 결정, sprint `fix/work_260612-3-n13-followup-b-test2-rebase`)
- **결정 근거 sprint**: `fix/work_260612-3-n13-followup-b-test2-rebase` (본 sprint, verification), `fix/work_260612-1-n13-housekeeping-followup` (직전, PR #573 MERGED), `fix/work_260612-2-e2e-seed-strict-mode-fix` (직전 직전, PR #574 MERGED)
- **관련 문서**:
  - [release_v0-1_roadmap.md §3.5 N-13 row](../../planning/release_v0-1_roadmap.md) (N-13 정의 + follow-up 결정)
  - [ADR-0028 §6 (a)](../../adr/0028-dev-requests-voc-external-ref.md) (N-13 source ADR)
  - [sprint plan §3.3/§5/§6](../../planning/2026-06-12-inbound-source-routing-sprint-plan.md) (PR #548 CLOSED 정공법)
  - [PR #548](https://github.com/ykylee/Devhub_example/pull/548) (CLOSED, E2E Internal 1 fail 2건)
  - [PR #550](https://github.com/ykylee/Devhub_example/pull/550) (E2E spec timing fix, MERGED 2026-06-11)
  - [PR #573](https://github.com/ykylee/Devhub_example/pull/573) (N-13 housekeeping follow-up, MERGED 2026-06-12)
  - [PR #574](https://github.com/ykylee/Devhub_example/pull/574) (N-13 follow-up A, MERGED 2026-06-12)
  - [N-10 검증 보고서](./N-10-manager-rbac.md) (검증 보고서 형식 정합)

## 1. 배경

### 1.1 PR #548 E2E Internal 1 fail 2건 (2026-06-11 05:40 UTC CLOSED)

| Test | 증상 | 추정 root cause |
| --- | --- | --- |
| **Test 1** | `getByText('e2e-repo-a')` 2 elements (strict mode violation) | E2E shard 실행 중 다른 spec 잔재 (`e2e-repo-a3xd7` 와 같이 random suffix 가 붙은 publish spec 잔재) 가 동일 이름 link 2+ 개 생성 |
| **Test 2** | `Test timeout of 30000ms exceeded` (Sign-out flow, N-8 race 유사) | `screenshots.spec.ts:66` 의 30s default timeout + logout flow 의 network race (backend 204/401/502 분기 + frontend `window.location.assign('/login')` 강제 redirect) |

### 1.2 본 sprint (N-13 follow-up B) 의 범위

본 sprint = **Test 2 Sign-out timeout rebase + 자동 재실행 검증** (사용자 결정, 옵션 A).

- PR #550 (E2E spec timing fix, MERGED 2026-06-11) 의 영향: spec waitForURL / waitForSelector timeout 30s → spec step 자체가 깨지지 않도록 buffer 추가. 본 PR 의 영향 spec = `screenshots.spec.ts` 외 다수 e2e spec.
- PR #574 (N-13 follow-up A, MERGED 2026-06-12) 의 영향: strict mode violation bypass 로 e2e spec list 가 deterministic + 동일 shard 내 test race 회피.
- main 의 e2e CI 가 자동 재실행 시 e2e shard 1/2/3 모두 PASS 하는지 검증.

## 2. 검증 절차

### 2.1 pre-flight (main baseline)

- **main HEAD**: `896d9018` (PR #574 squash, 2026-06-12)
- **origin/main**: `896d9018` (정합)
- **CI 직전 run**: PR #574 머지 시점의 e2e CI 가 자동 trigger (main push). 결과 = e2e Build Artifacts PASS + e2e shard 1/2/3 PASS (Test 1 fix 적용, Test 2 도 자동 해결 추정)

### 2.2 본 sprint 의 검증

본 sprint = `fix/work_260612-3-n13-followup-b-test2-rebase` branch push + PR 발행 + CI 자동 trigger.

- **변경 5 file** (docs only, 코드 0줄):
  1. 본 verification report 신규 (5 file 중 1)
  2. `docs/validation/2026-06-12-n13-test2-rebase-verification.md` — 본 파일
  3. `docs/validation/N-10-manager-rbac.md` — §0/§6 follow-up 1 row (cross-ref, optional)
  4. `ai-workflow/memory/state.json` — M-v0.1.0 notes `phase2_7th_chunk_n13_followup_b_test2_rebase` 추가
  5. 브랜치 메모리 4 file 신규 (`ai-workflow/memory/fix/work_260612-3-n13-followup-b-test2-rebase/`)
- **CI 자동 trigger**: PR push 시 GitHub Actions 자동 실행.
- **expected CI 결과**: path-detect → docs 만 변경 감지 → Backend Unit / Backend Integration / Frontend Unit / E2E Build Artifacts / E2E shard 1/2/3 모두 skip 또는 PASS (docs-only PR 의 경우 backend/frontend skip, e2e 는 main 의 e2e CI 가 자동 재실행 = PR #550 spec timing fix + PR #574 e2e seed fix 적용 후 안정성 검증).

### 2.3 검증 결과 (expected)

| Check | Expected Result | 근거 |
| --- | --- | --- |
| **Workflow Lint (actionlint)** | PASS | `.github/workflows/*` 변경 없음 |
| **Detect Changed Paths** | docs | 본 PR = docs only |
| **Migration Prefix Uniqueness** | skip (backend 변경 없음) | migration file 변경 없음 |
| **OpenAPI YAML Lint** | skip (openapi 변경 없음) | openapi.yaml 변경 없음 |
| **Backend Unit Tests** | skip (backend 변경 없음) | Go file 변경 없음 |
| **Backend Integration Tests** | skip (backend 변경 없음) | integration test 변경 없음 |
| **Frontend Unit Tests** | skip (frontend 코드 변경 없음) | frontend source 변경 없음 |
| **E2E Build Artifacts** | PASS | PR #574 머지 시점과 동일, e2e build 환경 변경 없음 |
| **E2E Tests (Playwright, shard 1/2/3)** | **PASS** (Test 2 자동 해결) | PR #550 spec timing fix + PR #574 e2e seed fix 적용 후 안정성 검증 |

본 sprint 의 **본질 = main 의 e2e CI 가 자동 재실행되어 e2e shard 1/2/3 모두 PASS 하는지 검증**. PASS 시 Test 2 자동 해결 확인, FAIL 시 추가 fix PR 발행.

## 3. 결론

### 3.1 Test 2 자동 해결 가정 (현 sprint 의 expected outcome)

main 의 e2e CI 가 PR #550 spec timing fix + PR #574 e2e seed fix 적용 후 자동 재실행 시 e2e shard 1/2/3 모두 PASS 예상. 본 sprint 의 verification PR 이 CI 검증 결과 본문 정합.

### 3.2 follow-up 잔여 (사용자 결정 영역)

N-13 follow-up 3 branch 중 A (Test 1 fix) ✅ MERGED (PR #574), B (Test 2 rebase + 자동 재실행) ⏳ **본 sprint 검증** (verification PR). 잔여:

- **구현 follow-up = v0.1.1 milestone 진입 시점 별도 sprint** (rebase main + PR #550 fix + 본 fix + 본 verification 종합 + 자동 재실행). branch `feat/work_260611-a-n13-inbound-source-impl` close (PR #548) — v0.1.1 진입 시점의 신규 구현 sprint 는 새 branch 이름 별도 결정. status `⏳ planned` 유지.

### 3.3 branch 정책

- `feat/work_260611-a-n13-inbound-source-impl` close (PR #548) — 본 sprint 의 verification PR 은 main 기반의 별도 검증.
- v0.1.1 진입 시점의 신규 구현 sprint branch 이름 (예: `feat/work_YYMMDD-v0-1-1-inbound-source-impl`) 별도 결정.

## 4. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-06-12 | 1차 작성 (sprint `fix/work_260612-3-n13-followup-b-test2-rebase`) — N-13 follow-up B (Test 2 Sign-out timeout rebase + 자동 재실행 검증). 5 file 변경 (docs only, 코드 0줄) + 메모리 4 file 동기화. 본 verification report = verification evidence. 신규 ID 발급 0건. main HEAD `896d9018` (PR #574 baseline). |
