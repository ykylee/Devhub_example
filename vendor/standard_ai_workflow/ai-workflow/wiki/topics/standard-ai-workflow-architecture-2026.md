---
type: topic
status: active
last_ingested_from: README.md + workflow-source/README.md + v0.6.3-beta release notes
related_pages: [entities/standard-ai-workflow, entities/workflow-source, entities/ai-workflow-runtime, entities/workflow-kit, decisions/adr-001-3-layer-separation, decisions/adr-004-wiki-layer, decisions/adr-005-r9-wiki-source-rule, concepts/project-architecture, concepts/harness-distribution, concepts/memory-3-state-lifecycle]
created: 2026-06-12
updated: 2026-06-12
---

# Standard AI Workflow — Architecture (v0.6.3-beta, 2026-06-12)

## TL;DR

v0.6.3-beta 의 `standard-ai-workflow` 는 **Source (workflow-source) → Runtime (ai-workflow + 6 harness overlays) → Project Docs (docs/)** 의 3-layer 분리 (ADR-001) 위에 **Memory 3-State Lifecycle** (ADR-005), **LLM Wiki** (ADR-004), **Contract v1 Multi-Component Fan-Out/In** 을 누적한 cross-cutting 표준 워크플로우 패키지다.

| # | 항목 | 값 |
|---|---|---|
| 1 | 현재 버전 | v0.6.3-beta (2026-06-12) |
| 2 | 이전 베이스라인 | v0.6.0-beta (Codex/OpenCode 패키지 동시 배포) |
| 3 | Source layer | `workflow-source/` (PR-reviewed SSOT) |
| 4 | Runtime layer | `ai-workflow/` (memory gitignore, wiki git) |
| 5 | Project Docs | `docs/` (PR-reviewed) |
| 6 | Wiki | `ai-workflow/wiki/` (5 types, R1~R7, V-1~V-8) |
| 7 | 지원 하네스 | 6종 (codex, opencode, gemini-cli, antigravity, minimax-code, pi-dev) |
| 8 | Python | 3.10+ (3.11 권장), editable install `workflow_kit` |
| 9 | Smoke CI | `python 3.11` + `PYTHONPATH=workflow-source` 52개 `check_*.py` |
| 10 | 안정 MCP transport | `jsonrpc-bridge` (default) — `stdio-sdk` 은 experimental |

### Cross-cutting invariants

| Invariant | 적용 범위 | 강제 rule |
|---|---|---|
| 3-layer separation (Source/Runtime/Project Docs) | repo 전체 | ADR-001 + D1 (PR review) |
| Memory immutability after freeze | `memory/archive/**` | R8 freeze + R10 lint (broken 파일/누락 marker = error) |
| Wiki ingest source = `archive/` only | `ai-workflow/wiki/**` | R9 wiki-source-rule |
| 1 commit = 1 ingest (5~15 page atomic) | `ai-workflow/wiki/**` | R2 page atomicity |
| Wiki location 단일 | `ai-workflow/wiki/` | R1 (다른 위치 사용 금지, V-1) |
| Additive merge (충돌 시 양쪽 보존) | wiki merge 시 | R5 + R7 (LLM reviewer 의무) |
| Push 직전 `fetch && rebase` | wiki sync | R3 pull-before-push + `.ingest_lock` |
| ID 영구 보존 | R/A/V/P/ADR 번호 | SCHEMA §1.2 (개정 시 ID 유지, version 만 bump) |
| Harness 추가 = 2 line | `bootstrap_lib/harnesses/__init__.py` + `bootstrap_lib/__main__.py` | |
| `ai-workflow/` 는 일반 탐색 제외 | export bundle, project code search | README §8 (메타 레이어 명시) |

## Layered View

3-layer separation ([[decisions/adr-001-3-layer-separation]]) — 각 layer 는 위치·추적·소유자·내용이 직교한다. Wiki layer (v0.6.0+, [[decisions/adr-004-wiki-layer]]) 와 Memory layer (v0.6.1+, [[decisions/adr-005-r9-wiki-source-rule]]) 는 Runtime 내부의 sub-layer 로 본다. 상세: [[concepts/project-architecture]].

| Layer | Location | Git | Owner | Content |
|---|---|---|---|---|
| **Source** | `workflow-source/` | tracked, PR review | maintainer | `core/`, `templates/`, `skills/`, `mcp_servers/`, `harnesses/`, `examples/`, `tests/`, `workflow_kit/`, `releases/` |
| **Runtime** | `ai-workflow/` | mixed (memory gitignore, wiki tracked) | session AI agent | `memory/{active,archive,release}/`, `wiki/`, `core/` (sync from source) |
| **Runtime → Wiki** | `ai-workflow/wiki/` | tracked, R1~R7 | LLM + human | 5 types: `entities/`, `concepts/`, `decisions/`, `patterns/`, `queries/` |
| **Runtime → Memory** | `ai-workflow/memory/` | gitignore, R8/R9/R10 | session → freeze → release | 3-state: `active/` → `archive/YYYY-MM-DD/` → `release/v0.X.Y/` |
| **Project Docs** | `docs/`, root `*.md` | tracked, PR review | project maintainer | `PROJECT_PROFILE.md`, `INSTALLATION_AND_USAGE.md`, `RELEASE.md`, runbooks |

## Component Map

| Component | Type | Version | Role |
|---|---|---|---|
| `workflow-source/` | source SSOT | v0.6.3-beta | PR-reviewed core/templates/skills; [[entities/workflow-source]] |
| `workflow-source/core/` | governance docs | v0.6.3-beta | `global_workflow_standard.md`, `orchestrator_subagent_contract_v1.md`, `workflow_harness_distribution.md` 등 |
| `workflow-source/templates/` | templates | v0.6.3-beta | project profile, session handoff, work backlog, daily backlog |
| `workflow-source/skills/` | skill impls | 11종 (1차 6 + 2차 2 + 3차 3) | `backlog-update`, `merge-doc-reconcile`, `wiki-ingest`, `wiki-query`, `wiki-lint`, `memory-freeze` 등 |
| `workflow-source/mcp_servers/` | MCP impls | v0.6.3-beta | `jsonrpc-bridge` (stable) + `stdio-sdk` (experimental) |
| `workflow_kit/` | Python pkg | v0.5.7.1+ wheel | `common` (30+ submodules) + `contract_v1/` + `bootstrap_lib`; editable install `pip install -e`; [[entities/workflow-kit]] |
| `docs/` | project docs | v0.6.3-beta | `INSTALLATION_AND_USAGE.md`, `RELEASE.md`, `PROJECT_PROFILE.md` |
| `ai-workflow/` | runtime layer | v0.6.3-beta | sync from source; [[entities/ai-workflow-runtime]] |
| Harness: **codex** | overlay | v0.6.3-beta | `AGENTS.md` + `.codex/config.toml.example`; [[entities/harness-overlay-codex]] |
| Harness: **opencode** | overlay | v0.6.3-beta | `AGENTS.md` + `opencode.json` + `.opencode/agents/{orchestrator,worker,doc,code,validation}` |
| Harness: **gemini-cli** | overlay | v0.6.3-beta | `GEMINI.md` (단일 진입점, `invoke_agent` 분리) |
| Harness: **antigravity** | overlay | v0.6.3-beta | `ANTIGRAVITY.md` (단일 진입점, browser sub-agent 위임) |
| Harness: **minimax-code** | overlay | v0.6.3-beta | `AGENTS.md` + `MiniMax.md` + `MiniMax_config.example.json` + `.MiniMax/agents/*` (Codex/OpenCode 와 동일 4 worker) |
| Harness: **pi-dev** | overlay | v0.6.3-beta | `AGENTS.md` + `SYSTEM.md` (페르소나 기반) |
| `ai-workflow/wiki/` | LLM wiki | v0.6.0+ | 5 types, 7 rules (R1~R7), 8 validations (V-1~V-8), 4 phases (P1~P4) |
| `ai-workflow/memory/` | session state | v0.6.1+ | 3-state lifecycle: active ↔ archive ↔ release; R8/R9/R10 |
| `dist/harnesses/<h>/<v>/` | export bundle | v0.6.3-beta | runtime + overlay + manifest + zip; pattern: harness factory (HARNESS_SPECS + builder) |
| `workflow-source/scripts/` | runners | v0.6.3-beta | `run_demo_workflow.py`, `run_existing_project_onboarding.py`, `export_harness_package.py`, `run_workflow_linter.py`, `run_git_conflict_resolver.py`, `run_memory_freeze.py`, `generate_workflow_state.py` |
| `workflow-source/global-snippets/` | 비침투적 snippet | v0.6.3-beta | 하네스 전역 설정에 복사 가능한 snippet 예시 (default exclude) |
| `workflow-source/examples/` | 샘플 + demo | v0.6.3-beta | `acme_delivery_platform/`, `end_to_end_skill_demo.md`, `end_to_end_mcp_demo.md`, `mcp_config_examples/` (5종), `output_samples/` |
| `workflow-source/tests/` | smoke baseline | 52 `check_*.py` | docs / bootstrap / harness export / output sample / generated schema / validation / code-index / onboarding runner / read-only MCP bundle / contract v1 multi-component / wire guide 회귀 |
| `tools/check_packaging.py` | packaging smoke | v0.5.8+ | `export_harness_package.py` 산출물 자동 검증 |

상위 entity: [[entities/standard-ai-workflow]]. 6-harness 모델: [[concepts/harness-distribution]]. 4-worker 페르소나: [[concepts/agent-topology]].

## Lifecycle

End-to-end: source 변경이 wiki 페이지까지 도달하는 8단계. 각 단계에 rule / trigger / owner 가 매핑된다.

| # | Stage | Trigger | Owner | Rule / Tool | Output |
|---|---|---|---|---|---|
| 1 | **Source edit** | PR opened | maintainer | PR review (D1) | `workflow-source/**` merge to main |
| 2 | **Bootstrap sync** | post-merge | maintainer / script | `python3 -m bootstrap_lib --target-root . --harness <h> --copy-core-docs --force` | `ai-workflow/core/`, harness overlay, `state.json` |
| 3 | **Runtime write** | session start | session AI agent | D2 (session write) | `ai-workflow/memory/active/{session_handoff,work_backlog,backlog,state}.*` |
| 4 | **Freeze (R8)** | session end | session AI agent | `memory-freeze` skill, [[concepts/memory-3-state-lifecycle]] | `memory/archive/YYYY-MM-DD/` + `.frozen` marker |
| 5 | **Wiki ingest (R9)** | per-session ingest | LLM + human | `wiki-ingest` skill, R2 (1 commit = 1 ingest) | atomic page updates (5~15) + `log.md` append |
| 6 | **Lint (V-1~V-8)** | weekly / pre-release | linter skill | `check_wiki_*.py` (5항목: contradiction, stale, orphan, missing, broken) | error/warning report |
| 7 | **Query** | user ask | LLM | `wiki-query` skill, R4 index load | synthesized answer; ≥30 lines = `queries/<date>-<slug>.md` |
| 8 | **Release** | release time | release manager | `RELEASE.md` 절차, deep freeze | `memory/release/v0.X.Y/` + zip + release note |

Bootstrap 의 R9 source rule: [[concepts/wiki-source-rule-r9]]. ADR 배경: [[concepts/project-architecture]].

## Version

| Version | Date | Highlight |
|---|---|---|
| v0.5.4 | 2026-Q2 | contract v1 외부 spec + `workflow_kit/contract_v1/` (issue #1 영구 해결) |
| v0.6.0 | 2026-06-12 | LLM Wiki layer (ADR-004 accepted), 5 page types, R1~R7; Codex + OpenCode 동시 패키지 |
| v0.6.1 | 2026-06-12 | Memory 3-State Lifecycle (ADR-005 proposed), R8 freeze, raw-layer policy |
| v0.6.1.5 | 2026-06-12 | wiki-source-rule R9 (ingest source = `archive/` only) |
| **v0.6.3-beta** | **2026-06-12** | **현재 baseline.** 6-harness export 정렬, 11종 skill, 52 smoke CI, contract v1 multi-component fan-out/in, R10 immutability lint |

Cadence: 메이저 (X.0) 분기 단위, 마이너 (X.Y) 월 단위, 패치 (X.Y.Z) 주 단위. Beta 단계는 API 변경 가능 (R1/R2 rule ID 자체는 영구 보존 — SCHEMA §1.2). Release procedure: `docs/RELEASE.md`.

## Related

### Concepts

- [[concepts/project-architecture]] — Source / Runtime / Project Docs 3-layer 정의
- [[concepts/project-architecture]] — 3-Layer + LLM Wiki + Memory 3-State 통합 요약
- [[concepts/harness-distribution]] — 6-harness overlay model
- [[concepts/memory-3-state-lifecycle]] — active ↔ archive ↔ release, R8 freeze 메커니즘
- [[concepts/agent-topology]] — orchestrator + 4 worker (worker/doc/code/validation) 페르소나
- [[concepts/mcp-transport]] — jsonrpc-bridge (stable) + stdio-sdk (experimental)
- [[concepts/contract-v1-output-validation]] — output_validator + MUST NOT delegate 7 패턴
- [[concepts/orchestrator-subagent-pattern]] — multi-sub 위임 + validate_fanin_output
- [[concepts/wiki-source-rule-r9]] — ingest source = `archive/...` only

### Decisions

- [[decisions/adr-001-3-layer-separation]] — Source / Runtime / Project Docs 분리 (v0.5.2)
- [[decisions/adr-004-wiki-layer]] — Karpathy LLM Wiki 패턴 채택 (v0.6.0 accepted)
- [[decisions/adr-005-r9-wiki-source-rule]] — Memory as Raw Layer (v0.6.1+ proposed)

### Entities

- [[entities/standard-ai-workflow]] — 저장소 entity (top-level)
- [[entities/workflow-source]] — Source layer entity
- [[entities/ai-workflow-runtime]] — Runtime layer entity
- [[entities/workflow-kit]] — `workflow_kit` Python package entity
- [[entities/harness-overlay-codex]] / [[entities/harness-overlay-opencode]] / [[entities/harness-overlay-minimax-code]] — 하네스 overlay entity (3종 대표)

### Patterns / Topics

- — `HARNESS_SPECS` 한 줄 + `register_harness_builder` 한 줄로 harness 추가
- [[topics/harness-distribution-model]] — 6-harness 모델의 상위 topic 페이지

### External refs (NOT wiki pages)

- `workflow-source/README.md` — workflow-source layer 진입점
- `ai-workflow/wiki/SCHEMA.md` — wiki 운영 헌법 (R1~R7, A1~A4, V-1~V-8, P1~P4)
- `ai-workflow/wiki/index.md` — wiki index (R4 anchor 기반)
- `ai-workflow/wiki/log.md` — ingest 이벤트 로그 (D5 append-only)
- `docs/INSTALLATION_AND_USAGE.md` — 설치/editable install/스모크 테스트 가이드
- `docs/RELEASE.md` — 릴리스 절차 (v0.5.7+)
