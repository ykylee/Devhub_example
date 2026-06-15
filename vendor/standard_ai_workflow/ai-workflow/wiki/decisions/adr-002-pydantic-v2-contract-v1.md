---
type: decision
status: accepted
adr_id: ADR-002
decided_at: 2026-05-04
accepted_in: v0.5.4
alternatives_considered: [internal-spec-only, pydantic-strict, plain-dataclass, json-schema]
related_pages: [concepts/contract-v1-output-validation, concepts/orchestrator-subagent-pattern, concepts/agent-topology]
created: 2026-06-12
updated: 2026-06-12
---

# ADR-002: Pydantic v2 Contract v1 외부 Spec + Python Helper 결합

## Status

Accepted (v0.5.4, 2026-05-04). v0.5.6 (P0 enforcement: `validate_output` + `choose_role` 자동 enforce), v0.5.7 (multi-component fan-out/in: `choose_roles` + `validate_fanin_output` + `recommend_model_tier`), v0.5.10 (sub.delegation_id parent-prefix rule: `{parent_id}-st-N` 강제), v0.5.11 (`enforce_subagent_response` / `enforce_fanin_response` Mavis engine hook, opt-out 없음) 으로 incremental 발전. 정식 spec: [ADR-002 file](../../docs/architecture/ADR-002-pydantic-v2-contract-v1-external-spec.md).

## Context

v0.5.3 까지 orchestrator ↔ sub-agent 간 위임 contract 는 (a) docstring-only 정의이거나 (b) Pydantic v1 in-code 정의였다. 이로 인해 외부 consumer (Mavis, Mavis, OpenCode, Gemini CLI 통합자) 가 import 없이 spec 검토 불가, spec 위반 응답이 downstream 으로 그대로 흘러가는 enforcement 부재, docstring diff 가 spec 변경 PR review 비용 상승, Pydantic v1 의 discriminated union / strict mode / JSON schema 자동 생성 한계가 누적됐다.

해결 방향으로 두 layer 결합이 자연스러웠다. spec 의 SSOT 는 markdown 으로 외부 노출해 import 없는 review 가능, runtime enforcement 는 strict typed helper 로 자동 검사, spec ↔ helper 1:1 매핑으로 review 비용 최소화. Pydantic v2 가 v0.5.4 시점에 stable 이 되면서 strict mode + discriminated union + model_validator + JSON schema 자동 생성을 제공, 외부 markdown spec + Pydantic v2 helper 결합이 결정됐다.

상위 topology 와 4-역할 경계는 [[concepts/agent-topology]] 참조. 위임 패턴의 wiki layer 매핑은 [[concepts/orchestrator-subagent-pattern]] §3 참조.

## Decision

contract v1 = **(a) 외부 markdown spec (SSOT) + (b) Pydantic v2 기반 Python enforcement helper** 결합.

| # | 항목 | 위치 / API | 비고 |
|---|---|---|---|
| 1 | 외부 spec (SSOT) | `workflow-source/core/orchestrator_subagent_contract_v1.md` | §1 동기 ~ §8 검증 시나리오 |
| 2 | wire 가이드 | `workflow-source/core/orchestrator_contract_v1_wire_guide.md` | spec §2·§3 reference code |
| 3 | helper package | `workflow_kit/contract_v1/` | output_validator + delegator |
| 4 | single output 검증 | `output_validator.validate_output(payload, expected_delegation_id=...)` | `OutputValidationResult` 반환 (no raise) |
| 5 | fan-in output 검증 | `output_validator.validate_fanin_output(payload, expected_parent_delegation_id=...)` | §5.2 fan-in envelope |
| 6 | 단일 위임 결정 | `delegator.choose_role(task)` | `DelegationDecision(role, must_not_delegate, delegation_id, ...)` |
| 7 | 배치 위임 결정 | `delegator.choose_roles(task, strict=True)` | multi-component fan-out (v0.5.7+) |
| 8 | model tier 추천 | `delegator.recommend_model_tier(task)` | main / small 자동 (v0.5.7+) |
| 9 | P0 enforcement | `enforce_subagent_response` / `enforce_fanin_response` thin wrapper | 위반 시 `ValueError` raise, opt-out 없음 (v0.5.11 §6.5) |
| 10 | version 형식 | `contract_version: "1.0"` 고정 | v0.5.4+ 후방호환 v1.x |

`Mavis` / `Mavis` 측 wire 의무 (v0.5.11 §6.5): `delegate_to_subagent` / `fanout_to_subs` 함수는 sub-agent 응답 수신 직후 helper 호출 의무, opt-out 없음. helper 매핑 상세는 [[concepts/contract-v1-output-validation]] §2.2 참조.

## Consequences

### Positive

| 효과 | 메커니즘 |
|---|---|
| 외부 consumer 가 import 없이 spec 검토 가능 | markdown SSOT → Mavis / Mavis / OpenCode / Gemini CLI 통합자 |
| 자동 enforce (필드/타입/enum/필수여부) | Pydantic v2 strict mode + custom validation, v0.5.6+ P0 |
| spec 변경 review 가 markdown diff 로 명확 | SSOT 와 helper 1:1 매핑 |
| JSON schema 자동 export | `workflow_kit.common.output_contracts` 가 runtime output contract map |
| multi-version 호환 (v1.x) | `contract_version: "1.0"` 고정, 후방호환 보장 |
| Pydantic v2 의 strict mode + discriminated union + model_validator 활용 | v1 의 한계 해소 |

### Negative / Trade-offs

| 위험 | 완화 |
|---|---|
| spec duplication 으로 drift (markdown ↔ Pydantic 모델) | `check_contract_v1_*` 6 종 smoke test 가 양쪽 cross-check |
| helper 가 spec 의 subset 일 수 있음 (e.g. §7.4 재위임 1회 정책은 caller 책임) | helper 가 enforce 안 하는 규칙은 docstring + 주석으로 명시 |
| Pydantic v2 dependency 추가 | `workflow_kit` 만 import 하면 transitive, target Python 모두에서 v0.5.4 시점 available |
| dual maintenance 부담 | spec → helper sync smoke (`check_contract_v1_roundtrip.py`) |

## Alternatives Rejected

| Option | Reason |
|---|---|
| internal-spec-only | docstring 만으로는 external consumer review 불가, enforcement 0 |
| pydantic-strict (실제 구현은 stdlib dataclass 채택) | Pydantic v2 dependency 부담 — helper 가 stdlib only 로 전환된 후속 code reality |
| plain-dataclass (검증 부재) | runtime 타입/enum 검증 부재. spec 위반 응답 자동 검출 불가 |
| json-schema | runtime JSON schema 파싱 overhead + helper ↔ schema sync 부담 |

## 후속 결정 인용 (Subsequent Decisions)

본 ADR 이 가능케 한 후속 결정 (incremental 발전 timeline):

| 버전 | 결정 | 효과 |
|---|---|---|
| v0.5.6 | P0 enforcement: `validate_output` + `choose_role` 자동 enforce | spec 위반 시 `DelegationRejected` raise |
| v0.5.7 | multi-component fan-out/in 지원: `choose_roles` (배치), `validate_fanin_output` (§5.2), `recommend_model_tier` (main/small 자동) | sub-agent fan-out/in 패턴 정식 지원 |
| v0.5.10 | sub.delegation_id parent-prefix rule: `{parent_id}-st-N` 강제 | `choose_roles` 발급 시점 prefix enforce |
| v0.5.11 | Mavis engine hook (§6.5): `enforce_subagent_response` / `enforce_fanin_response` thin wrapper | invalid 시 `ValueError` raise, opt-out 없음 |

## Spec 구조 (외부 markdown §1-§8)

| § | 내용 | 역할 |
|---|---|---|
| §1 | 동기 | contract v1 의 필요성 + scope 정의 |
| §2 | 적용 범위 | orchestrator ↔ sub-agent 위임 envelope 한정 |
| §3 | 4개 역할 | orchestrator / doc / code / validation (workflow 는 임시) |
| §4 | 위임 입력 스키마 | single + §4.2 multi-component fan-out 입력 |
| §5 | 위임 출력 스키마 | single + §5.2 fan-in 출력 (sub_results[]) |
| §6 | 위임 가능/불가 카탈로그 | MUST / SHOULD delegate + MUST NOT 7+2 marker |
| §6.5 | Mavis engine hook | P0 enforcement (v0.5.11 추가) |
| §7 | 에러/폴백 정책 | §7.4 재위임 1회 (caller 책임) |
| §8 | 검증 시나리오 | wire 가이드의 reference code 가 §2·§3 매핑 |

## References

- 외부 spec: [`orchestrator_subagent_contract_v1.md`](../../workflow-source/core/orchestrator_subagent_contract_v1.md)
- wire 가이드: [`orchestrator_contract_v1_wire_guide.md`](../../workflow-source/core/orchestrator_contract_v1_wire_guide.md)
- helper: [`workflow_kit/contract_v1/output_validator.py`](../../workflow-source/workflow_kit/contract_v1/output_validator.py), [`workflow_kit/contract_v1/delegator.py`](../../workflow-source/workflow_kit/contract_v1/delegator.py)
- 정식 ADR: [ADR-002](../../docs/architecture/ADR-002-pydantic-v2-contract-v1-external-spec.md)
- 회귀 test 6 종: `check_contract_v1_delegator.py`, `check_contract_v1_multi_component.py`, `check_contract_v1_output_validator.py`, `check_contract_v1_roundtrip.py`, `check_contract_v1_role_mapping.py`, `check_contract_v1_direct_only.py`
- 상세 contract v1 enforcement 동작 (validate_output / validate_fanin_output / MUST NOT 7+2 marker / parent-prefix rule): [[concepts/contract-v1-output-validation]] §3·§4·§5
- 상위 topology 와 4-역할 경계: [[concepts/agent-topology]] §2
- 위임 패턴 wiki layer 매핑: [[concepts/orchestrator-subagent-pattern]] §3
