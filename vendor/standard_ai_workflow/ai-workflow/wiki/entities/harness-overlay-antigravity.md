---
type: entity
status: active
last_ingested_from: workflow-source/harnesses/antigravity/
related_pages: [entities/standard-ai-workflow, concepts/harness-distribution, concepts/agent-topology, entities/mcp-read-only-bundle]
created: 2026-06-12
updated: 2026-06-12
---

# Antigravity Harness Overlay

[[entities/standard-ai-workflow]] 6-harness overlay 중 하나. Antigravity CLI 타겟용 bootstrap emitter 이며, [[concepts/harness-distribution]] 의 overlay 패턴 (단일 진입 파일 + MCP config 선택 emit) 을 따른다.

## Role

Antigravity 하네스에서 표준 AI 워크플로우를 운영할 때 생성되는 overlay 산출물 묶음. `workflow-source/harnesses/antigravity/` 가 source, bootstrap 시 프로젝트 루트에 emit 된다.

| 속성 | 값 |
|---|---|
| Harness slug | `antigravity` |
| Display name | Antigravity |
| Display name (legacy) | (없음, v0.5.8+ 신규 등록, `HARNESS_DEFINITIONS` 경유 X) |
| Source dir | `workflow-source/harnesses/antigravity/` |
| Source files | `README.md` (36 lines), `apply_guide.md` (87 lines), `overlay_spec.md` (46 lines) |
| Bootstrap 상태 | v0.5.10-beta 완료. `HARNESS_SPECS` + `HARNESS_FILE_BUILDERS` 등록, `check_antigravity_mode()` smoke 추가 |
| 설계 철학 | 정책 원문 X, 공통 workflow 문서로 연결하는 thin overlay |
| Overlay 두께 | 최소 (`ANTIGRAVITY.md` 1개 + 선택적 MCP config 1개) |
| v0.6.3-beta baseline | 활성 (last_ingested 2026-06-12) |

## Entry Files

단일 진입점 + 선택적 MCP config. 별도 토폴로지/에이전트 설정 파일 없음.

| 파일 | 경로 | 역할 | 필수 |
|---|---|---|---|
| `ANTIGRAVITY.md` | `<project_root>/ANTIGRAVITY.md` | 하네스 진입점. `ai-workflow/memory/active/{session_handoff,work_backlog,PROJECT_PROFILE}.md` 로 연결 | yes (항상 emit) |
| `antigravity.mcp.json` | `<project_root>/antigravity.mcp.json` | read-only MCP config snippet. `mcpServers` 키, `standardAiWorkflowReadOnly` 항목 | `--enable-mcp` 사용 시만 |
| (no config.toml) | — | Codex/OpenCode 와 달리 TOML/JSONC 설정 파일 없음. 정책은 entry md 본문으로 위임 | n/a |
| (no agent dir) | — | `.antigravity/agents/` 같은 worker 정의 디렉터리 없음 | n/a |

`overlay_spec.md` 의 "추가 overlay 파일 없음" 정책 — `ANTIGRAVITY.md` 외에는 bootstrap 시 emit 하지 않는다. 글로벌 설정(`~/.antigravity/config.json`) 의 `mcpServers` 블록은 사용자 측 수동 merge (apply_guide §4.2).

## Agent Topology

[[concepts/agent-topology]] 분류상 **단일 main agent + browser sub-agent 위임** 모델. orchestrator/worker split 없음.

| 차원 | Antigravity | OpenCode (대조) | Codex (대조) |
|---|---|---|---|
| Main agent | 1 (default 세션) | 1 orchestrator | 1 orchestrator |
| Worker 정의 | (없음) | doc/code/validation worker 분리 | bounded worker 패턴 |
| Sub-agent 종류 | browser sub-agent (Antigravity 내장) | generic + specialized (`.opencode/agents/*.md`) | bounded worker (`.codex/agents/*.md`) |
| contract v1 적용 | (대상 X, worker 정의 부재) | v0.5.6+ output_validator + delegator.choose_role | v0.5.6+ 동일 |
| 위임 라우팅 | Antigravity 런타임 자동 (browser sub-agent) | `choose_roles` (v0.5.7+ multi-component) | `choose_roles` 동등 |

핵심 차이: Antigravity 는 오버레이 측에서 worker 토폴로지를 강제하지 않고, 하네스 런타임의 browser sub-agent 가 정보 수집을 담당한다. 따라서 표준 `contract v1` 의 `sub.delegation_id` parent-prefix 규칙(`{parent_id}-st-{N}`) 적용 대상이 아니다. multi-component fan-out/in 도 worker 정의 부재로 비활성.

## MCP Config

[[entities/mcp-read-only-bundle]] 의 Antigravity emit 분기. `transport_ready=false` 이면 manual review only (v0.5.10+ 정책).

| 속성 | 값 |
|---|---|
| File | `antigravity.mcp.json` |
| Format | JSON (Gemini CLI 와 스키마 호환: `command`, `args`, `env`, `trust`, `includeTools`) |
| Emit 조건 | `--enable-mcp` 플래그 또는 `--harness antigravity` + mcp 활성 시 |
| Server key | `standardAiWorkflowReadOnly` |
| Entry point | `python3 -m workflow_kit.server.read_only_jsonrpc --stdio-lines` |
| Env | `PYTHONPATH=/ABSOLUTE/PATH/TO/standard_ai_workflow/workflow-source`, `STANDARD_AI_WORKFLOW_ROOT=/ABSOLUTE/PATH/TO/<project_root>` |
| 글로벌 merge 위치 | `~/.antigravity/config.json` 의 `mcpServers` 블록 |
| Transport 우선순위 | 1. `jsonrpc-bridge` (default, stable) / 2. `stdio-sdk` (experimental, 회귀) |
| Descriptor 위치 | `bundle/source-docs/schemas/read_only_transport_descriptors.json` (Antigravity entry) |
| descriptor `transport_ready` | v0.6.3-beta 기준 `false` (draft, manual review only) |

권장: 처음 도입은 `jsonrpc-bridge` transport. 글로벌 `~/.antigravity/mcp.json` merge 시 `jq -s '.[0].mcpServers * .[1].mcpServers'` 패턴 (apply_guide §4.1) 또는 단순 `cp`. 자세한 절차: `workflow-source/core/mcp_installation_by_harness.md`, 예시 원본: `workflow-source/examples/mcp_config_examples/antigravity-mcp.json`.

## Related

**Entities:**
- [[entities/standard-ai-workflow]] — hub entity. 6-harness overlay registry, Antigravity 포함
- [[entities/mcp-read-only-bundle]] — 6+1 read-only MCP 서버 묶음. Antigravity emit 경로/transport 정의
- [[entities/bootstrap-wiki-py]] — `bootstrap_lib/wiki.py`. Antigravity 와 직접 연결은 없지만 같은 registry/builder 패턴

**Concepts:**
- [[concepts/harness-distribution]] — 6-harness overlay registry/builder + MCP config emit 흐름의 단일 사례
- [[concepts/agent-topology]] — orchestrator/worker 패턴 분류. Antigravity 는 single main + browser sub-agent (worker 정의 부재)
- [[concepts/mcp-transport]] — `jsonrpc-bridge` vs `stdio-sdk`, failure mode F-1~F-5, Antigravity 동일 적용

**Source references:**
- `workflow-source/harnesses/antigravity/README.md` — overlay 개요, bootstrap 확장 상태
- `workflow-source/harnesses/antigravity/apply_guide.md` — 단계별 도입 절차, MCP 설치, 트러블슈팅
- `workflow-source/harnesses/antigravity/overlay_spec.md` — 식별자/예상 파일/체크리스트 (v0.4.19 draft, v0.5.10에 완료)
- `workflow-source/examples/mcp_config_examples/antigravity-mcp.json` — emit 예시 원본
