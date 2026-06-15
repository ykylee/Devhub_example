---
type: concept
status: active
last_ingested_from: workflow-source/core/orchestrator_subagent_contract_v1.md + workflow-source/core/workflow_agent_topology.md
related_pages: [concepts/mcp-transport, decisions/ADR-004-wiki-layer]
created: 2026-06-12
updated: 2026-06-12
---

# Orchestrator / Sub-agent Pattern

- 문서 목적: standard_ai_workflow 의 메인 오케스트레이터가 sub-agent 워커에게 작업을 위임할 때 따르는 외부 contract v1 의 핵심 패턴을 정리한다. 4개 역할 경계, 위임 가능/불가 카탈로그, 멀티 컴포넌트 fan-out/fan-in 포함.
- 범위: 4개 역할, 위임 입력/출력 스키마, MUST NOT delegate list, fan-out/fan-in
- 관련 결정: ADR-004 (wiki layer 도입, P1 review 대상)
- 최종 수정일: 2026-06-12

## §1 TL;DR  {#s1-tldr}

| # | 항목 | 값 |
|---|---|---|
| 1 | 외부 contract | `orchestrator_subagent_contract_v1.md` (v1, v0.5.4 부터, v0.5.7 멀티 컴포넌트 1차 컷) |
| 2 | 역할 수 | 4 (orchestrator, doc-worker, code-worker, validation-worker) + 임시 workflow-worker |
| 3 | 입력 키 | `contract_version`, `delegation_id`, `task.*`, `context.*` |
| 4 | 출력 키 | `worker.*`, `result.*`, `warnings`, `risks`, `sub_results` (fan-in) |
| 5 | MUST NOT delegate | handoff/backlog 직접 write, ask_user, cross-ref 갱신, fan-in 통합, parent_delegation_id 발급 |
| 6 | fan-out/fan-in | v0.5.7 부터. `sub_tasks[]` 분해, `sub_results[]` 통합 |
| 7 | Mavis enforcement | v0.5.11 P0. `enforce_subagent_response` / `enforce_fanin_response` 호출 강제 |

## §2 4개 역할 + 1 임시  {#s2-roles}

contract v1 은 4개 역할을 정의한다. 각 역할은 책임, 권한, 도구 성격이 다르며, 워커는 하나의 역할에 매핑된다.

| 역할 | 책임 | write 권한 | 금지 | 기본 모델 |
|---|---|---|---|---|
| **orchestrator** | 우선순위 결정, 결과 통합, 사용자 보고, 위험 조정, ask_user | handoff / backlog / state.json / PROJECT_PROFILE (직접) | 대량 read, bounded patch, 빌드/테스트 직접 실행 | main |
| **doc-worker** | 대량 문서 read, 비교, 허브/상태 문서 초안 | 맡은 범위 안 md/docx/html | 코드/설정 직접 수정, 빌드/테스트, 사용자-facing 결정 | small (구조 재설계 시 main) |
| **code-worker** | 범위 명확한 구현, 코드/설정 수정, 빌드/컴파일 확인, 좁은 리팩터, 테스트 보강 | 맡은 범위 안 소스/설정/스크립트 | 정책 결정, 사용자-facing 보고, handoff 직접 write | small (아키텍처 변경/광범위 리팩터/높은 회귀 위험은 main) |
| **validation-worker** | 테스트 실행, 로그 확인, 검증 증빙 수집, 실패 원인 요약 | 검증 결과 보고 (대상 파일 직접 수정 X) | 사용자-facing 보고, handoff 직접 write, 정책 결정 | small (실패 원인 복합/로그 해석 어려움/테스트 전략 재설계는 main) |
| **workflow-worker** (임시) | 분류 어려운 bounded task 또는 임시 작업 | task-scoped execution | 가능하면 doc/code/validation 으로 분해 권장 | small (여러 도메인 판단 묶음 시 main) |

권장 운영: `main orchestrator + small workers` 가 기본. 위험도/설계 복잡도가 올라가면 해당 worker 만 `main` 으로 일시 승격.

## §3 위임 가능 / 불가 카탈로그  {#s3-delegation-catalog}

오케스트레이터는 다음 카탈로그를 보고 직접 처리 vs 위임을 결정한다. 이 목록은 권장이 아니라 contract v1 의 의무 분배다.

### §3.1 MUST delegate  {#s3-1-must-delegate}

| 작업 | 권장 역할 | 비고 |
|---|---|---|
| 5개 이상 파일 read | doc-worker | bulk context read |
| 2개 이상 문서 비교/대조 | doc-worker | diff/dedup |
| 단일 문서 200줄+ 초안 작성 | doc-worker | 분량 큰 단일 산출물 |
| 5개 이상 파일 bounded patch | code-worker | multi-file change |
| 빌드/컴파일/lint/test 실행 | validation-worker | 명령 단위 검증 |
| 5개 이상 로그 파일 read/요약 | validation-worker | run result 수집 |
| 외부 저장소 분석/탐색 | doc-worker | pilot validation 등 |
| 새 파일 300줄+ 작성 | doc-worker 또는 code-worker | kind 에 따라 |

### §3.2 SHOULD delegate  {#s3-2-should-delegate}

짧은 bounded task 도 sub-agent 위임 권장. 컨텍스트 보호 목적. 단, 직접 처리도 허용 (필수 아님). 예시: 1~2개 파일 patch, 단일 파일 50~200줄 read, 단일 테스트 실행.

### §3.3 MUST NOT delegate (직접 처리)  {#s3-3-must-not-delegate}

| 작업 | 처리 주체 | 이유 |
|---|---|---|
| handoff / backlog / state.json 갱신 | orchestrator | handoff/backlog 는 orchestrator 의 단일 진실 공급원 |
| ask_user 호출 | orchestrator | 사용자 인터랙션은 오케스트레이터 책임 |
| 우선순위 결정 | orchestrator | 트레이드오프는 orchestrator 가 통합 판단 |
| sub-agent 출력 통합/리뷰 | orchestrator | 결과 통합은 orchestrator 책임 |
| PR 본문 작성 | orchestrator | 사용자-facing 메시지 |
| 짧은 triage read (1 파일, 50줄 미만) | orchestrator | 컨텍스트 부풀림 작음 |
| 메모리 layer session_handoff.md / state.json 직접 write | orchestrator | 단일 진실 공급원 |
| **cross-ref 갱신** (v0.5.7 신규) | orchestrator | cross-ref 는 단일 진실 공급원(orchestrator) 만이 일관성 보장 가능. sub-agent 가 부분 갱신 시 dead link 위험 |
| **fan-in 결과 통합 보고 작성** (v0.5.7 신규) | orchestrator | sub_results 통합은 sub-agent 출력 통합/리뷰의 멀티 컴포넌트 버전 |
| **parent_delegation_id 발급** (v0.5.7 신규) | orchestrator | fan-out 의 부모 ID 는 fan-out 결정의 일부, orchestrator 만 발급 |

## §4 위임 입력/출력 스키마  {#s4-schemas}

### §4.1 입력 핵심 필드 (orchestrator → sub-agent)  {#s4-1-input}

| 필드 | 타입 | 필수 | 설명 |
|---|---|---|---|
| `contract_version` | string | yes | `"1.0"` 고정 |
| `delegation_id` | string | yes | `del-YYYY-MM-DD-NNN` 패턴, 세션 내 유니크 |
| `issued_by.role` | enum | yes | 항상 `"orchestrator"` |
| `task.task_type` | enum | yes | `doc_draft` / `code_change` / `validation_run` / `bounded_research` |
| `task.brief` | string | yes | 1~2 문장, 무엇을 만들어야 하는지 |
| `task.expected_outputs.primary_artifact` | path | yes | 주 산출물 경로 |
| `task.validation.owner` | enum | yes | 항상 `"orchestrator"` |
| `task.required_model_tier` | enum | no | v0.5.7. `"small"` 기본 / `"main"` 강제 승격 |
| `task.sub_tasks` | object[] | no | v0.5.7 fan-out. 1개 이상이면 fan-out 모드 |

전체 payload 예시, fan-out fan-in 예시는 [원본 spec](../../workflow-source/core/orchestrator_subagent_contract_v1.md) §4·§5 참조.

### §4.2 출력 핵심 필드 (sub-agent → orchestrator)  {#s4-2-output}

| 필드 | 필수 | 설명 |
|---|---|---|
| `result.status` | yes | `ok` / `partial` / `failed` |
| `result.summary` | yes | 1~3 줄 요약 |
| `result.artifacts[]` | yes | path, kind, lines, action |
| `result.written_paths` | yes | 실제 write 한 파일 경로 |
| `result.next_step` | yes | orchestrator 가 다음에 할 일 |
| `result.sub_results[]` | no (v0.5.7) | fan-in. sub_id + status + summary + written_paths 최소 5필드 |

전체 입력 payload 예시, fan-out fan-in JSON 예시, `warnings` / `risks` / `parent_delegation_id` 등 보조 필드 정의는 [원본 spec](../../workflow-source/core/orchestrator_subagent_contract_v1.md) §4·§5·§5.2 참조.

## §5 Multi-Component Fan-Out/Fan-In (v0.5.7+)  {#s5-fanout}

복합 작업을 여러 sub-agent 가 분담할 때 사용하는 fan-out 입력. 부모 task 의 `sub_tasks` 필드에 1개 이상 sub-task 정의. 각 sub-task 는 §4 의 task 본질을 그대로 따른다.

### §5.1 fan-out 규칙  {#s5-1-fanout-rules}

| 항목 | 규칙 |
|---|---|
| `sub_tasks` 1개 이상 | 반드시 fan-out 모드 (단일 위임 아님) |
| per-sub delegation_id | `del-YYYY-MM-DD-<parent-suffix>-<sub_id>` 권장 |
| sub_task_type | 부모와 독립 가능 (e.g. 부모 `code_change` + sub `validation_run` 혼재) |
| sub_result.sub_id | 부모 fan-out 보고 내에서 유일 (caller 책임, v0.5.11 §4.2 보강) |
| sub_result.status 집계 | 모든 `ok` → `ok`, 하나라도 `failed` → `failed`, 그 외 → `partial` (자동 계산) |

### §5.2 Mavis 측 P0 enforcement (v0.5.11)  {#s5-2-mavis-p0}

orchestrator 측 `delegate_to_subagent` / `fanout_to_subs` 가 sub-agent 응답 수신 직후, 반드시 `validate_output` / `validate_fanin_output` 으로 envelope 을 검증해야 한다. 위반 시 `OutputValidationResult.raise_if_invalid()` 가 `ValueError` raise.

| 호출 | 검증 helper | 위치 |
|---|---|---|
| `delegate_to_subagent(payload)` | `enforce_subagent_response(response, expected_delegation_id=...)` | sub 응답 직후 |
| `fanout_to_subs(...)` | `enforce_fanin_response(fanin_payload, expected_parent_delegation_id=...)` | fan-in payload 직후 |

opt-out 없음 (P0). hook 우회, silent skip, marker 비활성화 모두 금지. 정의: `workflow_kit/contract_v1/output_validator.py`. re-export: `workflow_kit/contract_v1/__init__.py`.

## §6 wiki layer 와의 관계  {#s6-wiki}

wiki layer 의 모든 write 는 §3.3 의 MUST NOT delegate 영역에 속한다. ADR-004 의 wiki layer 분리 정책 (R1) 과 cross-ref 갱신 MUST NOT delegate 가 동시에 적용. `wiki-ingest` / `wiki-lint` / `wiki-query` 3종 skill (P2~P3) 은 sub-agent 가 호출 가능하지만, 실제 페이지 갱신 write 는 orchestrator 측 helper 가 수행.

## §7 다음에 읽을 문서  {#s7-next}

- 분산 규칙: [../.omo/plans/v0.5.11-plus-llm-wiki-distributed-rules.md](../.omo/plans/v0.5.11-plus-llm-wiki-distributed-rules.md)
- 운영 헌법: [../SCHEMA.md](../SCHEMA.md)
- 관련 개념: [./mcp-transport.md](./mcp-transport.md)
- 원본 contract: [../../workflow-source/core/orchestrator_subagent_contract_v1.md](../../workflow-source/core/orchestrator_subagent_contract_v1.md)
- agent 토폴로지: [../../workflow-source/core/workflow_agent_topology.md](../../workflow-source/core/workflow_agent_topology.md)

## §8 Revision Log  {#s8-revision-log}

| Date | Version | Change | Author |
|---|---|---|---|
| 2026-06-12 | 0.1.0 | 초안. 4개 역할, MUST/SHOULD/MUST NOT 카탈로그, 입력/출력 핵심 필드, fan-out/fan-in, Mavis P0 enforcement, wiki layer 관계. P1 bootstrap 의 일부 | Sisyphus (orchestrator) |
