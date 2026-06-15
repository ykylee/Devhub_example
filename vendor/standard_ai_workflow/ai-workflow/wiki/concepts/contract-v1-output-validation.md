---
type: concept
status: active
last_ingested_from: workflow-source/core/orchestrator_subagent_contract_v1.md + ADR-002
related_pages: [concepts/orchestrator-subagent-pattern, concepts/agent-topology, decisions/adr-002-pydantic-v2-contract-v1]
created: 2026-06-12
updated: 2026-06-12
---

# Contract v1 Output Validation (Pydantic v2)

- 문서 목적: standard_ai_workflow 의 orchestrator ↔ sub-agent 위임 contract v1 의 외부 markdown spec + Python enforcement helper 결합 정책을 정리한다. §5/§6 P0 enforcement, MUST NOT delegate 7+2 marker, fan-out/in 검증, sub.delegation_id parent-prefix rule.
- 범위: 외부 spec 구조, `output_validator` / `delegator` API, MUST NOT marker, P0 hook, parent-prefix 강제
- 최종 수정일: 2026-06-12

## §1 TL;DR  {#s1-tldr}

| # | 항목 | 값 |
|---|---|---|
| 1 | 외부 spec | `workflow-source/core/orchestrator_subagent_contract_v1.md` (v1.0, v0.5.4+) |
| 2 | 결합 정책 | 외부 markdown spec + Python helper (rationale: [[decisions/adr-002-pydantic-v2-contract-v1]]) |
| 3 | 언어 / 의존 | Python 3.10+ stdlib only (dataclass + 타입 체크; Pydantic 미사용) |
| 4 | version | `contract_version: "1.0"` 고정 (v0.5.4 부터 후방호환 v1.x) |
| 5 | 핵심 helper | `output_validator.validate_output` / `validate_fanin_output`, `delegator.choose_role` / `choose_roles` / `recommend_model_tier` |
| 6 | P0 enforce (v0.5.11) | `enforce_subagent_response` / `enforce_fanin_response` — 위반 시 `ValueError` raise, opt-out 없음 |
| 7 | MUST NOT marker | 7 base + 2 v0.5.7 신규 = 9 marker. substring match against `task.brief` + `primary_artifact` + `sub_tasks[].brief` |
| 8 | parent-prefix rule | sub.delegation_id = `{parent_id}-st-{N}` (1-based). `validate_fanin_output` 의 prefix check 가 enforce (v0.5.9+) |

## §2 Contract Spec — 외부 markdown + helper 결합  {#s2-contract-spec}

contract v1 은 두 layer 의 결합:

- **(a) 외부 markdown spec** — `workflow-source/core/orchestrator_subagent_contract_v1.md` 가 SSOT. §1 동기, §2 적용 범위, §3 4개 역할, §4 위임 입력 (single + §4.2 fan-out), §5 위임 출력 (single + §5.2 fan-in), §6 위임 가능/불가 카탈로그, §6.5 Mavis engine hook, §7 에러/폴백, §8 검증 시나리오.
- **(b) Python enforcement helper** — `workflow-source/workflow_kit/contract_v1/` 패키지. spec 의 §5/§6 을 `dataclass` + 타입 체크 + enum 으로 enforce. Pydantic 미사용 (stdlib only).

```python
from workflow_kit.contract_v1 import (
    validate_output, validate_fanin_output,
    enforce_subagent_response, enforce_fanin_response,
    choose_role, choose_roles, recommend_model_tier,
    DelegationRejected, OutputValidationResult,
)
```

### §2.1 왜 외부 spec + helper 결합인가  {#s2-1-why-external}

| 이유 | 효과 |
|---|---|
| 외부 consumer 가 import 없이 spec 검토 | Mavis / OpenCode / Gemini CLI 통합자가 markdown 으로 검토 가능 |
| 자동 enforce | v0.5.6 부터 P0. 응답 envelope 의 모든 필드/타입/enum 자동 검사 |
| spec 변경 추적 | markdown diff 로 review 명확. helper 변경이 spec 과 1:1 대응 |
| Pydantic v2 dependency 회피 | helper 가 stdlib only. Python 3.10+ 만 필요 |

상세 rationale: [[decisions/adr-002-pydantic-v2-contract-v1]].

### §2.2 외부 spec → helper 매핑  {#s2-2-mapping}

| spec § | 내용 | helper 매핑 |
|---|---|---|
| §3 | 4개 역할 + 1 임시 (orchestrator / doc / code / validation / workflow) | `delegator.TASK_TYPE_TO_ROLE` |
| §4 | 입력 스키마 (contract_version, delegation_id, task.*, context.*) | `output_validator.REQUIRED_TOP_FIELDS` |
| §4.2 | fan-out 입력 (sub_tasks[]) | `delegator._validate_sub_task` |
| §5 | 출력 스키마 (worker.*, result.*) | `output_validator.validate_output` |
| §5.2 | fan-in 출력 (sub_results[]) | `output_validator.validate_fanin_output` |
| §6.3 | MUST NOT delegate marker | `delegator.MUST_NOT_DELEGATE_MARKERS` (9 entries) |
| §6.5 | Mavis P0 hook | `enforce_subagent_response` / `enforce_fanin_response` |

## §3 Validation Pipeline — output_validator  {#s3-validation-pipeline}

sub-agent 응답을 Mavis / Mavis 측에서 수신한 직후, `enforce_subagent_response` (single) 또는 `enforce_fanin_response` (fan-in) 를 호출. 위반 시 `ValueError` raise (P0, opt-out 없음).

### §3.1 validate_output 흐름 (single)  {#s3-1-validate-output}

| 단계 | 검사 | 위반 시 |
|---|---|---|
| 1 | payload dict + REQUIRED_TOP_FIELDS 5개 (contract_version, delegation_id, completed_at, worker, result) | `field: missing` |
| 2 | `contract_version == "1.0"` | `contract_version: must be "1.0"` |
| 3 | (optional) `expected_delegation_id` 와 `payload.delegation_id` 일치 | cross-delegation leak 검출 |
| 4 | worker 3필드 (session_id, role, model_tier) + enum | `worker.X: missing` / enum 위반 |
| 5 | result 5필드 (status, summary, artifacts, written_paths, next_step) + enum | `result.X: missing` / enum 위반 |
| 6 | `result.artifacts[].kind` ∈ 7종 enum | `kind: must be one of ...` |

### §3.2 validate_fanin_output 흐름 (v0.5.7+)  {#s3-2-validate-fanin}

`validate_output` 결과를 먼저 받은 뒤, fan-in specific 규칙 4개를 추가 enforce:

| 단계 | 검사 | 위반 시 |
|---|---|---|
| 1 | `parent_delegation_id` 가 `expected_parent_delegation_id` 와 일치 | cross-delegation leak |
| 2 | `result.sub_results` 가 list, 최소 1개 | `must contain at least 1 sub-result` |
| 3 | 각 sub_result 5필드 (sub_id, delegation_id, status, summary, written_paths) | 필드별 누락 |
| 4 | sub_result.delegation_id prefix check (`startswith(parent_delegation_id)`) | `sub_delegation_id_prefix_mismatch` (v0.5.9 보강) |
| 5 | aggregated status: 모든 sub `ok` → `ok`, 하나라도 `failed` → `failed`, 그 외 → `partial` | `declared 'X' but sub_results aggregate to 'Y'` |

### §3.3 P0 Enforcement (v0.5.11 §6.5)  {#s3-3-p0}

```python
from workflow_kit.contract_v1 import enforce_subagent_response

def delegate_to_subagent(task, sub_caller):
    response = sub_caller(payload)
    # §6.5 P0: 위반 시 ValueError raise. §7.4 재위임 1회는 caller 책임.
    enforce_subagent_response(response, expected_delegation_id=decision.delegation_id)
    return response["result"]
```

opt-out 없음. hook 우회 / silent skip / `validate_*` 직접 호출 (enforce wrapper 우회) 모두 금지 — 회귀 시 catch 누락 위험.

## §4 MUST NOT Delegate 7+2 패턴  {#s4-must-not}

`delegator.MUST_NOT_DELEGATE_MARKERS` 가 `task.brief` + `task.expected_outputs.primary_artifact` + `task.sub_tasks[].brief` 를 substring 매치 (case-insensitive). 매치 시 `DelegationDecision.must_not_delegate=True` + `rejected_reason` 채워서 반환. `strict=True` 면 `DelegationRejected` raise.

| # | Marker | 이유 | 추가 |
|---|---|---|---|
| 1 | `handoff` | handoff 갱신 = orchestrator 단일 진실 공급원 | base |
| 2 | `backlog` | backlog 갱신 = orchestrator 단일 진실 공급원 | base |
| 3 | `state.json` | state.json 갱신 = orchestrator 단일 진실 공급원 | base |
| 4 | `ask_user` | 사용자 인터랙션 = orchestrator 책임 | base |
| 5 | `우선순위` | 트레이드오프 판단 = orchestrator 통합 책임 | base |
| 6 | `통합/리뷰` | sub-agent 출력 통합/리뷰 = orchestrator 책임 | base |
| 7 | `PR 본문` | 사용자-facing 메시지 = orchestrator 책임 | base |
| 8 | `cross-ref` | cross-ref 갱신 = orchestrator 단일 진실 (dead link 위험) | v0.5.7 신규 |
| 9 | `fan-in 통합` | fan-in 통합 보고 = 멀티 컴포넌트 통합 (sub-agent 출력 통합의 확장) | v0.5.7 신규 |

base 7 + v0.5.7 신규 2 = **9 marker**. spec §6.3 table 의 10행 중 `parent_delegation_id` 행은 `task.brief` 매치 대상이 아니라 fan-out 결정 시점의 orchestrator 책임이므로 marker 목록에 미포함. 상세 cross-ref: [[concepts/orchestrator-subagent-pattern]] §3.3.

## §5 Delegation API — choose_role / choose_roles  {#s5-delegation-api}

### §5.1 choose_role (single)  {#s5-1-choose-role}

`task.task_type` → role 결정 (deterministic mapping). §6.3 매치 시 `must_not_delegate=True`. `recommend_model_tier(task)` 가 main 후보 keyword ("아키텍처", "정책", "5+ 파일", "cross-cutting", "5+ source") 등장 시 `"main"`, 아니면 `"small"`. 명시 `task.required_model_tier` 우선.

| task_type | 매핑 role |
|---|---|
| `doc_draft` | `doc-worker` |
| `code_change` | `code-worker` |
| `validation_run` | `validation-worker` |
| `bounded_research` | `doc-worker` |

### §5.2 choose_roles (multi, fan-out, v0.5.7+)  {#s5-2-choose-roles}

`task.sub_tasks` 가 있으면 fan-out 모드. 각 sub-task 별 `choose_role` 호출 + 부모 decision 을 list[0] 에 prepend. 반환 형식: `[parent_decision, sub1_decision, ...]`. **v0.5.10 parent-prefix rule (필수)**: sub.delegation_id = `{parent_delegation_id}-st-{N}` (1-based). `validate_fanin_output` 의 prefix check 가 enforce.

| 항목 | 규칙 | 강제 지점 |
|---|---|---|
| sub.delegation_id 형식 | `{parent_id}-st-{N}` (1-based) | `choose_roles` 발급 시 |
| sub.delegation_id prefix | `str(sub.delegation_id).startswith(parent_id)` 필수 | `validate_fanin_output` (v0.5.9+) |
| Mavis 측 가공 | sub-agent 응답의 delegation_id 를 **그대로** 옮길 것 | wire 가이드 §8 안티패턴 #6 |
| sub_id 값 무관 | sub.sub_id 값 ("st-1" / "core-build" 등) 무관하게 st-N 형식 유지 | delegator.py:357 |

### §5.3 안티패턴 (피해야 할 wire)  {#s5-3-antipatterns}

| # | 안티패턴 | 결과 |
|---|---|---|
| 1 | `choose_role` 호출 없이 sub-agent 직접 spawn | contract v1 우회. 회귀 test fail |
| 2 | sub-agent 가 fan-in 보고서 작성 | §6.3 위반. orchestrator 만 fan-in |
| 3 | `sub_results` 무시하고 sub 응답 통째 합치기 | status consistency check 가 잡아냄 |
| 4 | `expected_delegation_id` 없이 `validate_output` 호출 | cross-delegation leak 못 잡음 |
| 5 | sub 응답의 `delegation_id` 를 Mavis 가 임의 재발급 | `sub_delegation_id_prefix_mismatch` 에러 |

## §6 Related Decisions  {#s6-related-decisions}

| 결정 | rationale | 상태 |
|---|---|---|
| [[decisions/adr-002-pydantic-v2-contract-v1]] | 외부 markdown spec + Pydantic v2 helper 결합 (실제로는 stdlib dataclass 채택) | Accepted (v0.5.4) |
| [[concepts/agent-topology]] | 4개 역할 경계 + write 권한 분배 | design complete, implementation 미포함 |
| [[concepts/orchestrator-subagent-pattern]] | contract v1 §3·§4·§5·§6 의 wiki layer 매핑 | concept 작성 완료 |

## §7 References  {#s7-references}

- 외부 spec: [contract_v1.md](../../../workflow-source/core/orchestrator_subagent_contract_v1.md) · [wire 가이드](../../../workflow-source/core/orchestrator_contract_v1_wire_guide.md)
- helper: [output_validator.py](../../../workflow-source/workflow_kit/contract_v1/output_validator.py) · [delegator.py](../../../workflow-source/workflow_kit/contract_v1/delegator.py) · [\_\_init\_\_.py](../../../workflow-source/workflow_kit/contract_v1/__init__.py)
- ADR: [ADR-002](../../../docs/architecture/ADR-002-pydantic-v2-contract-v1-external-spec.md)
- 분산 규칙: [v0.5.11-plus-llm-wiki-distributed-rules.md](../../../.omo/plans/v0.5.11-plus-llm-wiki-distributed-rules.md) · [운영 헌법](../SCHEMA.md)

## §8 Revision Log  {#s8-revision-log}

| Date | Version | Change | Author |
|---|---|---|---|
| 2026-06-12 | 0.1.0 | 초안. §1 TL;DR, §2 spec 결합, §3 validation pipeline (validate_output / validate_fanin / P0), §4 MUST NOT 7+2, §5 delegation API (choose_role / choose_roles / parent-prefix), §6 related decisions, §7 references. P1 bootstrap 의 일부 | Sisyphus (orchestrator) |
