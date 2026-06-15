# Mavis / Mavis Orchestrator → Contract v1 Wire 가이드 (v0.5.7)

> standard_ai_workflow 의 contract v1 enforcement 헬퍼(`workflow_kit.contract_v1`)를 Mavis / Mavis 같은 메인 오케스트레이터 런타임이 **실제 위임 결정 경로**에 묶는 패턴 가이드. v0.5.6 부터 모듈은 stable, v0.5.7 부터 fan-out/in wire 까지 포함.

## 1. 왜 wire 가이드가 필요한가

v0.5.6 의 `choose_role` / `validate_output` 은 **순수 Python 모듈** — 자동 enforce 가능하지만 orchestrator 가 명시적으로 호출해야 동작. v0.5.5 pilot 에서 4 시나리오 round-trip 은 PASS 였지만, **회귀 검증**이 자동화되어 있지 않아 "원칙적으로 강제" 만 되고 "실제 매 호출마다 강제" 되진 않았음.

v0.5.6 부터 회귀 스크립트(`check_contract_v1_delegator.py` / `check_contract_v1_output_validator.py`) 가 "강제 안 하면 PASS" 인 baseline 을 자동 검증. v0.5.7 부터 fan-out/in 도 동일.

본 가이드는 Mavis / Mavis 가 `choose_role` / `choose_roles` / `validate_output` / `validate_fanin_output` 을 **실제 sub-agent 호출 직전·직후**에 어떻게 wire 하는지 보여준다.

## 2. Single-task 위임 패턴

```python
from workflow_kit.contract_v1 import (
    choose_role,
    enforce_subagent_response,
    DelegationRejected,
)


def delegate_to_subagent(task: dict, sub_agent_caller) -> dict:
    """Standard single-task delegation with contract v1 enforcement."""
    # 1) 위임 결정 (§6)
    decision = choose_role(task)
    if decision.must_not_delegate:
        # §6.3 — orchestrator 직접 처리
        raise DelegationRejected(decision)

    # 2) sub-agent 호출 (§4 입력 페이로드에 decision.delegation_id 주입)
    payload = {
        "contract_version": "1.0",
        "delegation_id": decision.delegation_id,
        # ... (issued_by, task, context 등 §4 스키마 채움)
    }
    response = sub_agent_caller(payload)

    # 3) 출력 검증 (§5) — v0.5.11 §6.5: enforce helper 가 invalid 시
    # ValueError raise. §7.4 정책 (재위임 1회) 은 caller 책임.
    enforce_subagent_response(
        response,
        expected_delegation_id=decision.delegation_id,
    )

    return response["result"]
```

## 3. Multi-component fan-out/in 패턴 (v0.5.7)

```python
from workflow_kit.contract_v1 import (
    choose_roles,
    enforce_fanin_response,
)


def fanout_to_subs(task: dict, sub_agent_caller) -> dict:
    """v0.5.7: 멀티 컴포넌트 fan-out + parent fan-in 패턴."""
    # 1) fan-out 결정 (§4.2) — strict=True 면 §6.3 위반 시 raise
    decisions = choose_roles(task, strict=True)
    parent_decision = decisions[0]
    sub_decisions = decisions[1:]

    # 2) sub-agent 들을 병렬로 호출
    sub_payloads: list[dict] = []
    sub_responses: list[dict] = []
    for sub_decision in sub_decisions:
        sub_payload = {
            "contract_version": "1.0",
            "delegation_id": sub_decision.delegation_id,
            "parent_delegation_id": parent_decision.delegation_id,
            "task": {
                "sub_id": sub_decision.sub_id,
                "task_type": sub_decision.task_type,
                # ...
            },
        }
        sub_payloads.append(sub_payload)
        sub_responses.append(sub_agent_caller(sub_payload))

    # 3) fan-in 보고서 작성 (sub-agent 가 아니라 orchestrator)
    fanin_payload = {
        "contract_version": "1.0",
        "delegation_id": parent_decision.delegation_id,
        "parent_delegation_id": parent_decision.delegation_id,
        "completed_at": now_iso(),
        "worker": {
            "session_id": orchestrator_session_id(),
            "role": parent_decision.role,
            "model_tier": parent_decision.recommended_model_tier or "small",
        },
        "result": {
            "status": aggregate_status([r["result"]["status"] for r in sub_responses]),
            "summary": fanin_summary(sub_responses),
            "artifacts": fanin_artifacts(sub_responses),
            "written_paths": fanin_written_paths(sub_responses),
            "sub_results": [
                {
                    "sub_id": sub["task"]["sub_id"],
                    "delegation_id": sub["delegation_id"],
                    "status": r["result"]["status"],
                    "summary": r["result"]["summary"],
                    "written_paths": r["result"]["written_paths"],
                }
                for sub, r in zip(sub_payloads, sub_responses)
            ],
            "next_step": orchestrator_next_step(sub_responses),
        },
    }

    # 4) fan-in 검증 (§5.2) — v0.5.11 §6.5: enforce helper 가 invalid 시
    # ValueError raise. aggregated status mismatch / sub delegation_id
    # prefix mismatch 등을 한 줄로 enforce.
    enforce_fanin_response(
        fanin_payload,
        expected_parent_delegation_id=parent_decision.delegation_id,
    )

    return fanin_payload
```

## 4. model_tier 결정 (v0.5.7)

```python
from workflow_kit.contract_v1 import choose_role, recommend_model_tier


def delegate_with_model_tier(task: dict) -> DelegationDecision:
    """v0.5.7: 자동 main/small 결정 + 명시 override 존중."""
    decision = choose_role(task)
    if decision.must_not_delegate:
        return decision
    # decision.recommended_model_tier 가 자동 결정값 (small / main)
    # task.required_model_tier 가 명시되어 있으면 그게 우선
    actual_tier = decision.recommended_model_tier
    return decision


# main 후보 keyword (delegator 내부 정의):
# - "아키텍처", "정책", "5+ 파일", "cross-cutting", "5+ source"
# main 이 default 로 매핑되는 경우: 위 keyword 가 task.brief / constraints /
# must_include 에 등장하면 main, 아니면 small.
```

## 5. Mavis 측 권장 wiring 지점

Mavis / Mavis 가 sub-agent 위임을 결정하는 함수는 보통 다음 두 곳 중 하나:

| 지점 | Mavis 함수 (개념) | wire |
| --- | --- | --- |
| Task 라우팅 (delegation 결정) | `should_delegate(task) → bool` | `choose_role(task)` 호출로 대체. `must_not_delegate=True` 면 False (직접 처리) |
| Tool 사용 결정 | `execute_tool(name, args)` | 도구 호출 직전에 task 를 합성해서 `choose_role` 통과 여부 확인 |

### 5.1 minimal wire (Mavis 측 의사 코드)

```python
# mavis/orchestrator.py (개념)
from workflow_kit.contract_v1 import choose_role, DelegationRejected


def mavis_delegate(task: dict) -> DelegationDecision:
    """Mavis 가 task 를 받으면 무조건 choose_role 부터 호출."""
    decision = choose_role(task)
    if decision.must_not_delegate:
        # §6.3: orchestrator 직접 처리 영역
        return direct_handle(task, decision.rejected_reason)
    # fan-out 후보: task.sub_tasks 가 있으면 choose_roles
    if task.get("sub_tasks"):
        sub_decisions = choose_roles(task, strict=True)
        return fanout_to_subs(task, sub_decisions)
    # 일반 single 위임
    return delegate_to_subagent(task, decision)
```

## 6. 회귀 검증 (Mavis 통합 테스트)

v0.5.6 의 `check_contract_v1_delegator.py` / `check_contract_v1_output_validator.py`, v0.5.7 의 `check_contract_v1_multi_component.py` 가 자동 enforce baseline. Mavis 측은 추가로:

```bash
# 위 3개 회귀 + bootstrap + mcp_roundtrip + pilot 을 CI 에서 매 PR 마다 실행
python3 workflow-source/tests/check_contract_v1_delegator.py
python3 workflow-source/tests/check_contract_v1_output_validator.py
python3 workflow-source/tests/check_contract_v1_multi_component.py
python3 workflow-source/tests/check_pilot_phase11_contract_v1.py
```

## 7. v0.5.7 신규 wire 포인트 요약

| 신규 API | wire 시점 | 효과 |
| --- | --- | --- |
| `choose_roles(task, strict=...)` | sub-agent fan-out 결정 시 | sub_task 별 role + parent-prefix delegation_id 발급 (형식: `{parent_delegation_id}-st-{i}`, e.g. `del-2026-06-08-c6cc8da7-st-1`) |
| `validate_fanin_output(payload, expected_parent_delegation_id=...)` | sub-agent fan-in 보고 수신 시 | aggregated status 일관성 + sub delegation_id prefix enforce |
| `recommend_model_tier(task)` | `choose_role` 결과 확인 시 | task keyword 기반 main/small 자동 결정 |
| `decision.recommended_model_tier` | sub-agent 호출 페이로드 생성 시 | worker.model_tier 자동 주입 |
| `decision.sub_id` | sub-agent 호출 결과 routing 시 | multi-component 결과 → 원본 sub 매핑 |
| **sub.delegation_id prefix rule (v0.5.9 보강)** | fan-in 보고서 `sub_results[].delegation_id` 작성 시 | sub.delegation_id 는 반드시 parent.delegation_id 의 prefix 여야 한다 (e.g. `del-PARENT` → `del-PARENT-st-1`). `validate_fanin_output` 가 enforce — 위반 시 `sub_delegation_id_prefix_mismatch` 에러. **Mavis 측 sub-agent 응답을 가공할 때 절대 prefix 를 떼거나 재발급하지 말 것.** |

## 8. 안티패턴 (피해야 할 wire)

1. **`choose_role` 호출 없이 sub-agent 직접 spawn** — contract v1 우회. 회귀가 빨간불이어야 함.
2. **sub-agent 가 fan-in 보고서를 작성** — §6.3 의 "fan-in 통합 보고는 orchestrator" 위반. orchestrator 만 fan-in.
3. **`sub_results` 를 무시하고 sub 응답을 통째로 합치기** — §5.2 의 aggregated status 계산이 무의미해짐. `validate_fanin_output` 의 status consistency check 가 잡아냄.
4. **`expected_delegation_id` 없이 `validate_output` 호출** — cross-delegation leak (sub-agent 가 다른 delegation_id 응답) 못 잡음.
5. **model_tier 를 sub-agent 가 자기 결정** — §4.1 의 `required_model_tier` 가 orchestrator 결정값. sub-agent 가 무시 가능하지만 자동 검증.
6. **sub 응답의 `delegation_id` 를 orchestrator 가 임의 재발급** (v0.5.9 보강) — Mavis 가 sub-agent 응답을 fan-in 보고서 `sub_results[].delegation_id` 에 옮길 때 새 UUID / 새 prefix 부여하면 `validate_fanin_output` 가 `sub_delegation_id_prefix_mismatch` 로 거절한다. sub-agent 가 보낸 `delegation_id` 를 **그대로** 옮길 것. sub-agent 가 빈 값으로 보낸 경우에만 `sub_decision.delegation_id` (= `choose_roles` 가 미리 발급한 parent-prefix 값) 로 fill — 그 외 재가공 금지.

## 9. 다음 단계

- Mavis 측 `mavis_team` (engine) 의 `mavis communication send --command spawn` 호출 직전에 `choose_role` 자동 wire
- mavis-team 의 `delegate_to_subagent` hook 에서 `validate_output` enforce
- v0.5.8 (Beta) — GitHub Releases 만 배포 (PyPI 미배포 확정, `docs/RELEASE.md` 참조). v0.5.8 에서 `bootstrap_lib` 인터랙티브 `--harness` picker + packaging smoke 자동화 추가
- v0.5.9 (예정) — 본 문서의 §7/§8 보강 + 위 hook 들의 Mavis 표준 wiring. sub.delegation_id prefix 룰 자동 회귀 test (이미 `check_contract_v1_multi_component.py` `check_validate_fanin_sub_delegation_id_prefix_mismatch` 로 enforce)
