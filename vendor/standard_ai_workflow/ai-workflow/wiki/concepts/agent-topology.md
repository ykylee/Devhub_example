---
type: concept
status: active
last_ingested_from: workflow-source/core/workflow_agent_topology.md + dist/harnesses/opencode/v0.6.3-beta/agents/
related_pages: [concepts/orchestrator-subagent-pattern, concepts/contract-v1-output-validation, concepts/harness-distribution]
created: 2026-06-12
updated: 2026-06-12
---

# Agent Topology: 4-Role Orchestrator + 3-Worker Pattern (v0.5.7+)

- 문서 목적: standard_ai_workflow 의 표준 agent 토폴로지 (1 main orchestrator + 3 sub-agent workers) 의 구조, 권한 경계, 하네스별 변형을 정리한다.
- 범위: 4 역할, fan-out/fan-in, OpenCode/Codex 변형, 위임 메커니즘
- 관련 결정: [[concepts/orchestrator-subagent-pattern]] (위임 contract v1), [[concepts/contract-v1-output-validation]] (출력 envelope 검증)
- 최종 수정일: 2026-06-12

## §1 TL;DR  {#s1-tldr}

| 역할 | 인스턴스 수 | sub-agent 종류 | 주 책임 |
|---|---|---|---|
| **session-orchestrator** (main) | 1 | n/a (orchestrator 자체) | 우선순위 결정, 결과 통합, 사용자 보고, ask_user, fan-out/fan-in |
| **workflow-doc-worker** | 0~N | subagent (`mode: subagent`) | 대량 문서 read, 비교, 허브/상태 문서 초안 |
| **workflow-code-worker** | 0~N | subagent (`mode: subagent`) | 범위 명확한 코드/설정 수정, bounded patch, 빌드/컴파일 |
| **workflow-validation-worker** | 0~N | subagent (`mode: subagent`) | 테스트/로그 실행, 검증 증빙 수집, 실패 원인 요약 |

v0.5.7 부터 multi-component fan-out/fan-in 정식 지원 (`sub_tasks[]` 분해, `sub_results[]` 통합). v0.5.7+ 부터 `main`/`small` 2-tier 모델 분배 권장: orchestrator 는 `main`, workers 는 기본 `small`.

## §2 Role Topology  {#s2-role-topology}

### §2.1 main orchestrator (1개 고정)  {#s2-1-main-orchestrator}

| 항목 | 값 |
|---|---|
| 모드 | `primary` (OpenCode) / task-only bounded (Codex) |
| 기본 모델 | `main` (조정, 통합, 우선순위 판단에 깊은 문맥 유지) |
| 직접 read 범위 | `state.json`, `session_handoff.md`, `work_backlog.md`, `PROJECT_PROFILE.md` + 단일 파일 triage |
| 직접 edit | deny 권장. 모든 write 는 worker 경유 |
| ask_user | orchestrator 단독 권한 |
| fan-out 발급 | `parent_delegation_id` 발급 + `choose_roles` 호출 |
| 하네스 overlay | `dist/harnesses/opencode/v0.6.3-beta/bundle/.opencode/agents/workflow-orchestrator.md` |

### §2.2 workflow-doc-worker (0~N fan-out)  {#s2-2-doc-worker}

| 항목 | 값 |
|---|---|
| 모드 | `subagent` |
| 권한 | `edit: allow`, `bash: allow`, `webfetch: allow` (OpenCode overlay) |
| 기본 모델 | `small` (구조 재설계/정책 충돌 조정 시 `main` 승격) |
| write scope | 맡은 문서 범위 한정 (md/docx/html) |
| 금지 | 코드/설정 직접 수정, 빌드/테스트 실행, 사용자-facing 결정 |
| 하네스 overlay | `dist/harnesses/opencode/v0.6.3-beta/bundle/.opencode/agents/workflow-doc-worker.md` |

### §2.3 workflow-code-worker (0~N fan-out)  {#s2-3-code-worker}

| 항목 | 값 |
|---|---|
| 모드 | `subagent` |
| 권한 | `edit: allow`, `bash: allow`, `webfetch: allow` (OpenCode overlay) |
| 기본 모델 | `small` (아키텍처 변경/광범위 리팩터/높은 회귀 위험 시 `main` 승격) |
| write scope | 맡은 소스/설정/스크립트 한정 |
| 금지 | 정책 결정, 사용자-facing 보고, handoff 직접 write |
| 하네스 overlay | `dist/harnesses/opencode/v0.6.3-beta/bundle/.opencode/agents/workflow-code-worker.md` |

### §2.4 workflow-validation-worker (0~N fan-out)  {#s2-4-validation-worker}

| 항목 | 값 |
|---|---|
| 모드 | `subagent` |
| 권한 | `edit: allow`, `bash: allow`, `webfetch: allow` (OpenCode overlay) |
| 기본 모델 | `small` (실패 원인 복합/로그 해석 어려움/테스트 전략 재설계 시 `main` 승격) |
| write scope | 검증 결과 보고. 대상 파일 직접 수정 X |
| 금지 | 사용자-facing 보고, handoff 직접 write, 정책 결정 |
| 하네스 overlay | `dist/harnesses/opencode/v0.6.3-beta/bundle/.opencode/agents/workflow-validation-worker.md` |

## §3 Per-Harness Variation  {#s3-per-harness}

| 차원 | OpenCode (v0.6.3-beta) | Codex (v0.6.3-beta) |
|---|---|---|
| 오케스트레이터 정의 | `.opencode/agents/workflow-orchestrator.md` (TUI frontmatter `mode: primary`, `permission: edit/bash/webfetch = deny`) | `AGENTS.md` 본문 (TUI 없음, AGENTS.md 가 진입점) |
| worker 정의 | `.opencode/agents/workflow-{doc,code,validation}-worker.md` 3종 overlay (`mode: subagent`, `permission: allow`) | `AGENTS.md` 의 "Codex 전용 메모" 섹션 (banned pattern 문서화, overlay 파일 없음) |
| fan-out 메커니즘 | TUI `task` 위임 + `workflow_kit.contract_v1.delegator.choose_roles` 호출 (config-driven) | 동일 `choose_roles` 사용 + bounded scope 의 `task` 위임 (수동 라우팅) |
| 모델 분배 | per-agent model selection 지원 (TUI `model:` field). orchestrator=`main`, workers=`small` 기본 | `AGENTS.md` 텍스트 권장. 실제 per-agent 분배는 harness 제약 따라 |
| MUST NOT delegate | overlay frontmatter + 본문에 명시 | `AGENTS.md` 작업 원칙에 명시 |
| 배포 경로 | `dist/harnesses/opencode/v0.6.3-beta/bundle/.opencode/agents/` | `dist/harnesses/codex/v0.6.3-beta/bundle/AGENTS.md` |
| Export 도구 | `workflow-source/scripts/export_harness_package.py --harness opencode` | 동일 도구 `--harness codex` |

핵심 차이: OpenCode 는 overlay 파일 단위로 4-역할을 TUI 가 직접 인식, Codex 는 `AGENTS.md` 텍스트 패턴으로 동일 운영 원칙을 문서화. 양쪽 모두 contract v1 fan-out/fan-in 호출은 `workflow_kit.contract_v1` helper 로 강제.

## §4 Delegation Mechanics  {#s4-delegation-mechanics}

위임 입력/출력 envelope, `sub.delegation_id` parent-prefix 규칙, fan-in status 집계 등 메커니즘은 [[concepts/orchestrator-subagent-pattern]] 의 §4·§5 가 정식 spec. 본 절은 토폴로지 관점 요약.

### §4.1 단일 위임 (v0.5.4+)  {#s4-1-single}

| 단계 | 주체 | 호출 | 검증 |
|---|---|---|---|
| 1 | orchestrator | `delegate_to_subagent(payload)` (envelope: `contract_version`, `delegation_id`, `task.*`, `context.*`) | 입력 envelope: `contract_v1.input_schema.validate` |
| 2 | worker | scoped 실행 + `result.{status, summary, artifacts[], written_paths, next_step}` 반환 | 출력 envelope: `enforce_subagent_response(response, expected_delegation_id=...)` (P0) |
| 3 | orchestrator | 결과 통합, 사용자 보고 | `OutputValidationResult.raise_if_invalid()` 위반 시 `ValueError` |

### §4.2 multi-component fan-out/fan-in (v0.5.7+)  {#s4-2-fanout}

| 항목 | 규칙 |
|---|---|
| fan-out trigger | 부모 `task.sub_tasks` 길이 ≥ 1 |
| per-sub `delegation_id` | `{parent_id}-st-{N}` 형식 parent-prefix (v0.5.10 spec 정합) |
| sub_result 필수 필드 | `sub_id`, `status`, `summary`, `written_paths` 최소 4종 |
| status 집계 (자동) | 모든 `ok` → `ok`, 하나라도 `failed` → `failed`, 그 외 → `partial` |
| Mavis P0 enforcement | `fanout_to_subs(...)` 직후 `enforce_fanin_response(fanin_payload, expected_parent_delegation_id=...)` (opt-out 없음) |
| 정의 위치 | `workflow_kit/contract_v1/output_validator.py` + `delegator.choose_roles` |

### §4.3 통합 (orchestrator 책임)  {#s4-3-fanin}

| 동작 | 담당 | cross-ref |
|---|---|---|
| sub_results[] 취합 + status 집계 | orchestrator | [[concepts/contract-v1-output-validation]] §fanin |
| cross-ref 갱신 | orchestrator (sub-agent 직접 write 금지, contract v1 §3.3 MUST NOT) | [[concepts/orchestrator-subagent-pattern]] §3.3 |
| 사용자 보고 작성 | orchestrator | — |
| handoff/backlog/state.json 갱신 | orchestrator (단일 진실 공급원) | [[concepts/orchestrator-subagent-pattern]] §3.3 |

## §5 Related Decisions  {#s5-related-decisions}

- [[concepts/orchestrator-subagent-pattern]] — 외부 contract v1, 4 역할 경계, MUST/SHOULD/MUST NOT 카탈로그, fan-out/fan-in
- [[concepts/contract-v1-output-validation]] — `enforce_subagent_response` / `enforce_fanin_response` P0 enforcement, Pydantic v2 envelope
- [[concepts/harness-distribution]] — 6개 하네스 overlay 배포 (`Codex`, `OpenCode`, `Gemini CLI`, `Antigravity`, `MiniMax Code`, `pi-dev`)
- ADR-001 (planned): agent 토폴로지 역할/권한 경계 (source: `workflow-source/core/workflow_agent_topology.md` §3 권한 매트릭스)
- ADR-004: wiki layer (cross-ref 단일 진실 공급원 정책)

## §6 References  {#s6-references}

- 토폴로지 원본 spec: [`workflow-source/core/workflow_agent_topology.md`](../../workflow-source/core/workflow_agent_topology.md) (v0.5.4+, 2026-06-07)
- 외부 contract v1: [`workflow-source/core/orchestrator_subagent_contract_v1.md`](../../workflow-source/core/orchestrator_subagent_contract_v1.md)
- Contract v1 wire 가이드: [`workflow-source/core/orchestrator_contract_v1_wire_guide.md`](../../workflow-source/core/orchestrator_contract_v1_wire_guide.md)
- OpenCode orchestrator overlay: `dist/harnesses/opencode/v0.6.3-beta/bundle/.opencode/agents/workflow-orchestrator.md`
- OpenCode doc worker: `dist/harnesses/opencode/v0.6.3-beta/bundle/.opencode/agents/workflow-doc-worker.md`
- OpenCode code worker: `dist/harnesses/opencode/v0.6.3-beta/bundle/.opencode/agents/workflow-code-worker.md`
- OpenCode validation worker: `dist/harnesses/opencode/v0.6.3-beta/bundle/.opencode/agents/workflow-validation-worker.md`
- Codex AGENTS.md: `dist/harnesses/codex/v0.6.3-beta/bundle/AGENTS.md`
- Harness export 도구: [`workflow-source/scripts/export_harness_package.py`](../../workflow-source/scripts/export_harness_package.py)

## §7 Revision Log  {#s7-revision-log}

| Date | Version | Change | Author |
|---|---|---|---|
| 2026-06-12 | 0.1.0 | 초안. 4-role topology, OpenCode vs Codex 변형, fan-out/fan-in 메커니즘, contract v1 cross-ref. P2 wiki-ingest 의 일부 | Sisyphus (orchestrator) |
