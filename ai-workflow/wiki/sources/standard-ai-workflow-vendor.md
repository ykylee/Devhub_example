---
type: topic
status: active
last_ingested_from: ai-workflow/wiki/topics/standard-ai-workflow-vendor.md
related_pages: [sources/standard-ai-workflow-vendor]
created: 2026-06-15
updated: 2026-06-15
last_touched: 2026-06-16T04:49:13Z
git_commit: cac63f35
git_branch: main
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
mirror_dirty: |
---

# Standard AI Workflow Vendor (L2 dense, in-repo)

> L1 SSOT: `ai-workflow/wiki/topics/standard-ai-workflow-vendor.md`
> 본 L2 derived view 는 in-repo retrieval 용 압축 요약.

## TL;DR

`ykylee/standard_ai_workflow` = Mavis (MiniMax Code) 의 표준화 워크플로우 발매 패키지. v0.7.17 (commit 4d09dee) = 본 저장소 의 vendor import 기준. *발매자 vs 소비자* 관계, *본 저장소 의 SSOT 1차 출처* (vendor 40+ core spec + DevHub 1차 layer 운영 문서). Mavis 4 sub-agent role + 6 task mode + SSOT 1차 출처 (maturity_matrix.json).

## 1. vendor 구조 (28M → 5.7M, 746 file)

- `ai-workflow/` (1.7M, vendor 자체 active memory + wiki mini)
- `core/` (40+ spec doc, SSOT 1차 출처: global_workflow_standard, orchestrator_subagent_contract_v1, workflow_task_modes, workflow_agent_topology, workflow_harness_distribution, maturity_matrix.json)
- `harnesses/` (6 overlay: _template, antigravity, codex, gemini-cli, **minimax-code** [Mavis], opencode, pi-dev)
- `tests/` (840K, v0.7.17 11 smoke + 70+ check)
- `workflow_kit/` (640K)
- `examples/` (290K), `scripts/` (344K), `schemas/`, `templates/`, `skills/`, `mcp_servers/`, `releases/` (Beta-v0.7.17.md 등)
- `prompts/` (code_worker, doc_worker, validation_worker)

## 2. 발매자 vs 소비자

| | standard_ai_workflow | DevHub |
|---|---|---|
| `ai-workflow/` 의미 | 도구/스킬 발매 | 운영 메모리/스크립트/스킬 인프라 |
| `workflow-source/` 의미 | raw source | (없음) |
| 1차 출처 SSOT | `core/maturity_matrix.json` | `AGENTS.md` + `MEMORY_GOVERNANCE.md` + `minimax_code_workflow.md` + `PROJECT_PROFILE.md` |
| 외부 vault `~/wiki/` | ❌ (vendor 도구 자체 미사용) | ❌ (v0.7.17 결정) |

## 3. Mavis 운영 패턴 (vendor minimax-code)

### 3.1 진입 파일
- `AGENTS.md` (Codex/OpenCode 공통) + `MiniMax.md` (MiniMax Code 전용)
- DevHub: AGENTS.md 가 본 저장소 workflow layer (worker_division, v0.1.0 roadmap, 2-tier, 워커별 메모) 흡수. Mavis 운영 = `ai-workflow/minimax_code_workflow.md` (PR #601, 12 섹션, 352 line, active).

### 3.2 sub-agent 4 role
- **orchestrator** (메인) / **doc-worker** / **code-worker** / **validation-worker** / **workflow-worker** (임시)
- 위임 입력/출력 스키마 (`orchestrator_subagent_contract_v1.md` §4-§5)
- 멀티 컴포넌트 fan-out (§4.2 v0.5.7)
- 본 저장소 매핑: mavis-team prompt 6섹션 ↔ contract v1 field 1:1

### 3.3 작업 모드 6종
Analysis / Requirements / Design / Planning / Implementation / Refactoring. 본 저장소 의 code-index-update / backlog-update / mavis-team / mini-coder-max / merge-doc-reconcile skill 의 1:1 매핑.

## 4. Release cadence

- 현재 vendor import: v0.7.17 (commit 4d09dee, 2026-06-15)
- vendor latest: v0.7.21 (commit fa329b1) — 4 release ahead
- 갱신 SOP: `git pull` + `cp -R workflow-source/. vendor/` (build/dist/pycache 제외) + 16/16 smoke 회귀 + 메모리 anchor 갱신
