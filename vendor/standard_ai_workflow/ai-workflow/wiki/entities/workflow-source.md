---
type: entity
status: active
last_ingested_from: workflow-source/README.md
related_pages: [entities/standard-ai-workflow, entities/ai-workflow-runtime, entities/workflow-kit, decisions/adr-001-3-layer-separation]
created: 2026-06-12
updated: 2026-06-12
---

# workflow-source/

## Role

[[entities/standard-ai-workflow]] 의 **Source layer (SSOT)**. v0.5.2 부터 [[decisions/adr-001-3-layer-separation]] 으로 source / state / knowledge 3-layer 로 분리된 구조의 최상위 layer. 모든 skill / MCP / template / helper / doc / harness overlay 의 원본이 위치하며, `bootstrap_lib` 와 `apply_workflow_upgrade` 가 이 디렉토리에서 [[entities/ai-workflow-runtime]] (`ai-workflow/`) 과 `docs/` 로 렌더링한다.

핵심 책임:

- workflow kit 의 **단일 진실 공급원** (SSOT)
- contributor PR + maintainer approve 로 merge
- 사용자 / AI agent 가 직접 수정하지 않음 — bootstrap 이 덮어쓴다
- [[entities/workflow-kit]] (`workflow-source/workflow_kit/`) 와 동일한 디렉토리에 공존하는 reusable Python package 루트 포함

## Subdirectories

| Subdirectory | Contents |
|---|---|
| `core/` | 공통 표준 문서 (`global_workflow_standard.md`, `workflow_skill_catalog.md`, `workflow_mcp_candidate_catalog.md`, `workflow_agent_topology.md` 등) + skill spec + phase plan + contract v1 외부 spec |
| `templates/` | 프로젝트 프로파일 / 세션 handoff / work backlog / daily backlog / release note / pilot adoption record 템플릿 + `prompts/` worker prompt 모음 |
| `skills/` | 11종 skill 프로토타입 (`session-start`, `backlog-update`, `doc-sync`, `merge-doc-reconcile`, `code-index-update`, `project-status-assessment`, `validation-plan`, `automated-repro-scaffold`, `robust_patcher`, `git-conflict-resolver`, `workflow-linter`) + `workers/` (orchestrator/worker overlay) + `memory-freeze/` |
| `mcp_servers/` | 11종 read-only MCP 프로토타입 (`check-doc-links`, `latest-backlog`, `create-backlog-entry`, `create-session-handoff-draft`, `smart-context-reader`, `git-history-summarizer`, `apply_robust_patch`, `suggest-impacted-docs`, `check-doc-metadata`, `create-environment-record-stub`, `check-quickstart-stale-links`) + `lib/` |
| `examples/` | 샘플 프로젝트 (`acme_delivery_platform`, `research_eval_hub`) + end-to-end demo (`end_to_end_skill_demo.md`, `end_to_end_mcp_demo.md`, `orchestration_demo.py`, `orchestrator_feedback_loop_demo.py`) + MCP config 예시 5종 + output samples + pilot adoption 사례 |
| `global-snippets/` | 하네스 전역 설정용 비침투적 snippet 예시 (`codex/`, `opencode/`) |
| `harnesses/` | 6개 하네스 overlay (`codex/`, `opencode/`, `gemini-cli/`, `antigravity/`, `minimax-code/`, `pi-dev/`) + `_template/` 신규 하네스 추가용 stub |
| `scripts/` | end-to-end runner + bootstrap 도구 + `bootstrap_lib/` (wiki/mcp/harnesses 서브모듈 포함) + 출력 스키마/contract/transport descriptor 생성기 + harness export + `setup_gitignore.py` |
| `tools/` | packaging smoke 검증 도구 (`check_packaging.py`) |
| `tests/` | 52종 smoke test (bootstrap, harness export, contract v1 multi-component, wire guide, output sample, validation/code-index, onboarding runner, read-only MCP bundle 회귀) |
| `releases/` | 릴리스 노트 (`Alpha-v0.1.0` ~ `Beta-v0.6.3`) + pre-release 기록 (`prototype-v1-pre-release.md`, `prototype-v2-pre-release.md`) |
| `workflow_kit/` | reusable Python package 루트 (`common/`, `contract_v1/`, `harness/`, `server/`, `constants.py`, `gitignore.py`, `upgrade_diff.py`, `pyproject.toml`) — v0.5.2+ 본격 추출, v0.5.6+ contract_v1 추가 |
| `prompts/` | worker prompt 정의 (`code_worker_prompt.md`, `doc_worker_prompt.md`, `validation_worker_prompt.md`) — orchestrator → sub-agent 위임용 |
| `schemas/` | 생성된 Pydantic v2 schema (`generated_output_schemas.json`), output sample contract (`output_sample_contracts.json`), read-only MCP descriptor / JSON-RPC fixture / transport descriptor |

> 참고: `dist/`, `docs/`, `MEMORY_GOVERNANCE.md`, `pyproject.toml`, `standard_ai_workflow.egg-info/` 도 SSOT 트리에 포함되지만, 이 표는 위키가 추적하는 14개 semantic subdirectory 만 정리한다.

## Source-of-Truth Policy

`workflow-source/` 는 [[decisions/adr-001-3-layer-separation]] 의 **Layer 1 (Source)** 이며 다음 정책으로 운용된다.

| 항목 | 정책 |
|---|---|
| Owner | maintainer (Sisyphus) |
| 변경 빈도 | 릴리스 단위 (주~월 단위) |
| Review 정책 | contributor PR + maintainer approve |
| Git tracked | ✅ (전체) |
| 직접 수정 가능? | ❌ — contributor 만 PR 로 |
| 사용자 행동 | bootstrap / `apply_workflow_upgrade` 로만 반영 |
| 재실행 시 | smart update 정책 (v0.5.10.1+) + `PRESERVE_RELATIVE_PATHS` 로 state/knowledge 영역 보호 |

핵심 경계:

- `workflow-source/` 를 직접 편집해서 다시 bootstrap 하면 [[entities/ai-workflow-runtime]] 으로 그대로 렌더링된다. 반대로 `ai-workflow/` 나 `docs/` 를 수정해서 source 에 반영하려면 PR 을 거쳐야 한다.
- [[entities/workflow-kit]] 의 package code (`workflow-source/workflow_kit/`) 도 같은 SSOT 정책에 속한다. v0.5.2+ 부터 source layer 의 일부로 관리되며, CI smoke test (`workflow-source/tests/`) 가 52개 baseline 을 강제한다.
- `workflow_kit.constants.PRESERVE_RELATIVE_PATHS` 가 `ai-workflow/memory/`, `ai-workflow/WORKFLOW_INDEX.md`, `ai-workflow/README.md` 를 명시. upgrade / bootstrap 시 이 경로는 절대 덮어쓰지 않음.

## Related

- [[entities/standard-ai-workflow]] — 최상위 프로젝트, 본 entity 가 속한 저장소
- [[entities/ai-workflow-runtime]] — 본 SSOT 로부터 bootstrap 되는 runtime layer (`ai-workflow/`)
- [[entities/workflow-kit]] — `workflow-source/workflow_kit/` 에 위치한 reusable Python package
- [[decisions/adr-001-3-layer-separation]] — 본 entity 가 source layer 로 분리된 architectural 결정
- `workflow-source/README.md` — 본 entity 가 직접 ingest 하는 canonical source
- `workflow-source/MEMORY_GOVERNANCE.md` — memory layer 와의 경계를 명시한 governance 문서
- `workflow-source/core/workflow_state_vs_project_docs.md` — state doc vs project doc 경계 가이드
- `workflow-source/core/workflow_kit_roadmap.md` — kit 진화 roadmap
