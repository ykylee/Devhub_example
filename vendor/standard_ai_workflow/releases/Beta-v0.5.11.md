# Beta v0.5.11 — Mavis engine hook + ADR 정식 기록 + 비대화형 가이드 보강

- **릴리스 일자**: 2026-06-09 (1차), 2026-06-16 (2차 — PR-3b ADR 3건 머지 후)
- **브랜치**: `main`
- **포함 커밋**: 5 (1차) + 1 (2차, PR-3b ADR 3건 + architecture/README 갱신. 7-day cool-down 후)
  - 1차 (#1 + #2 + #3 + #4 governance + release note): Mavis engine hook + 회귀 test 강화 + 비대화형 가이드 + governance 갱신
  - 2차 (#4 ADR 3건): ADR-001/002/003 정식 기록 (PR-3b, 7-day cool-down)
- **상태**: ✅ P0 enforcement 자동화 + ✅ 영구 지식 자산 정식 기록 + ✅ docs governance 1인 dev 환경 명시. breaking change 없음.

## 1. 무엇이 바뀌었나

### 1.1 Mavis engine hook 자동 enforce (#1, P0)

contract v1 §6.5 (신규) 가 Mavis 측 `delegate_to_subagent` / `fanout_to_subs` 호출 시 **반드시** `validate_output` / `validate_fanin_output` 으로 응답 envelope 을 enforce 하도록 강제. 이전엔 docstring-level 가이드였으나, v0.5.11 부터 thin wrapper helper 가 enforce.

```python
from workflow_kit.contract_v1 import (
    choose_role,
    enforce_subagent_response,  # 신규
)
from workflow_kit.contract_v1 import (
    choose_roles,
    enforce_fanin_response,  # 신규
)
```

`enforce_subagent_response(payload, expected_delegation_id=...)` / `enforce_fanin_response(payload, expected_parent_delegation_id=...)` 가 위반 시 `ValueError` raise. **opt-out 없음 (P0)**.

#### wire 가이드 §2/§3 reference code 갱신

`workflow-source/core/orchestrator_contract_v1_wire_guide.md` 의 §2/§3 example 이 본 helper 사용으로 갱신. 비표준 식별자 (`OutputValidationFailed`, `log_violation`) **완전 제거** — 표준 contract_v1 public API 만 사용.

`check_wire_guide_v059.py` 의 stub namespace 도 helper 호환 버전으로 동시 갱신 (PR-1 의 commit 1건). `except _OutputValidationFailed` → `except ValueError` 로 변경. 0줄 누락 시 wire 가이드 walkthrough 가 `ValueError` 그대로 propagate → test fail.

### 1.2 회귀 test 강화 (#2, P0)

- `validate_fanin_output` 의 raise path 회귀 test 2종 추가 (`check_contract_v1_multi_component.py`):
  - **`check_validate_fanin_parent_delegation_id_missing_raises`** — fan-in payload 에 `parent_delegation_id` 누락 시 `is_valid=False`
  - **`check_validate_fanin_sub_id_duplicate_with_status_inconsistency_detected`** — 동일 sub_id + status 불일치 동반 시 aggregated mismatch 로 **indirect 검출**

**Known limitation (시나리오 B)**: 동일 sub_id + 동일 status 의 중복은 aggregated check 가 통과함. 옵션 (b) — `choose_roles` 에 dedup check 추가 — 는 **v0.5.12+ 별도 plan** (breaking change 가능).

- `check_contract_v1_output_validator.py` 의 `check_violation_wrong_contract_version` 는 pre-existing (v0.5.6 부터). 본 릴리스의 contract_version test 추가 0건. **DoD 의 "신규 1개" 는 redundant** 로 판정, 변경 없음.

- spec `orchestrator_subagent_contract_v1.md` §4.2 에 sub_id uniqueness = caller 책임 + 시나리오 B 미검출 한계 1줄 명시.

### 1.3 `--no-interactive` 비대화형 가이드 보강 (#3, P2)

v0.5.8 부터 비대화형 환경에서 `--harness` 미지정 시 `SystemExit(1)` + 6개 harness 목록 fail-fast 동작은 변함 없음. 본 릴리스는 **문서만 보강**:

- `workflow-source/core/mcp_installation_by_harness.md` §3 끝에 비대화형 정책 단락 + bash 예시 1개
- `docs/INSTALLATION_AND_USAGE.md` §7.1 에 옵션 C (CI/스크립트 환경 권장 호출) 추가

신규 코드 0줄.

### 1.4 ADR 정식 기록 (#4, P1)

`docs/architecture/` 가 **작성 예정** 상태에서 **ADR 3건 정식 기록** 상태로 전환:

- **ADR-001: Source/State/Knowledge 3-layer 분리** — v0.5.2 commit `96431f1` 의 rationale. 3-layer 책임, 갱신 정책, PRESERVE_RELATIVE_PATHS 정의. v0.5.10.1 hotfix 의 smart update 와의 결합 (PRESERVE 가 hotfix 의 정책 우선순위에서 가장 강함) 명시.
- **ADR-002: Pydantic v2 contract v1 외부 spec + Python helper 결합** — v0.5.4 의 rationale. 외부 markdown spec 의 장점 + Pydantic v2 helper 의 자동 enforce. v0.5.6 / 5.7 / 5.10 / 5.11 의 후속 결정 인용 (Consequences 섹션). 압축 형식.
- **ADR-003: Read-only MCP 우선 정책** — v0.5.7 의 rationale. 에이전트 자율성 vs 안전성 균형. `create_backlog_entry` 의 draft-only 예외 명시. transport 우선순위 (jsonrpc-bridge default / stdio-sdk opt-in).

`docs/architecture/README.md` §3 컴포넌트 명세에 3 ADR link 추가 + §4 "향후 작성 예정" 에서 3건 완료 표시.

#### governance 갱신 (옵션 A)

`docs/README.md` §1 (현 분류) + §3 (PR 가이드라인) 에 1인 dev 환경 명시:

> 본 저장소는 현재 1인 dev 환경 (Sisyphus, Mavis/MiniMax M3 기반 AI agent) 으로 운영되며, 본인 (`Sisyphus`) 은 `문서 관리자` 로 자임한다. 모든 `docs/` PR 은 self-review + 24h cool-down 후 머지하며, 추후 reviewer 가 추가되면 본 단락을 갱신한다.

> **Governance 변경 PR** (예: 본 README §3 의 1인 dev 명시 자체) 은 **7-day cool-down** 적용. governance 의 meta 순환 (governance 가 governance 변경을 self-merge) 위험을 "fresh eyes" interval 로 완화.

## 2. 변경 diff 요약

| 파일 | 변경 종류 | 라인 |
| --- | --- | --- |
| `workflow-source/core/orchestrator_subagent_contract_v1.md` | §6.5 Mavis engine hook (신규 절) + §4.2 sub_id uniqueness (1줄) | +60 / -2 |
| `workflow-source/core/orchestrator_contract_v1_wire_guide.md` | §2/§3 reference code helper 사용 + 비표준 식별자 제거 | +6 / -8 |
| `workflow-source/workflow_kit/contract_v1/output_validator.py` | `enforce_subagent_response`, `enforce_fanin_response` 신규 helper | +30 |
| `workflow-source/workflow_kit/contract_v1/__init__.py` | 신규 helper 2개 re-export | +5 / -2 |
| `workflow-source/core/mcp_installation_by_harness.md` | §3 비대화형 정책 단락 + bash 예시 1개 | +15 |
| `docs/INSTALLATION_AND_USAGE.md` | §7.1 옵션 C (CI 환경) | +12 / -2 |
| `docs/architecture/ADR-001-source-state-knowledge-3-layer-separation.md` | **신규** | +180 |
| `docs/architecture/ADR-002-pydantic-v2-contract-v1-external-spec.md` | **신규** | +150 |
| `docs/architecture/ADR-003-read-only-mcp-default-policy.md` | **신규** | +130 |
| `docs/architecture/README.md` | §3 ADR link, §4 backlog 갱신 | +6 / -3 |
| `docs/README.md` | §1 architecture 상태, §3 1인 dev 명시 + 7-day cool-down | +5 / -3 |
| `workflow-source/tests/check_contract_v1_delegator.py` | `enforce_subagent_response` test 2종 | +30 |
| `workflow-source/tests/check_contract_v1_multi_component.py` | `enforce_fanin_response` test 1종 + #2 회귀 test 2종 | +50 |
| `workflow-source/tests/check_wire_guide_v059.py` | stub namespace helper 호환 + `except ValueError` 변경 | +10 / -15 |

## 3. 검증 (actual run, fresh venv)

### 3.1 신규 회귀 test

```text
$ python3 workflow-source/tests/check_contract_v1_delegator.py
Contract v1 §6.1/§6.3 delegator smoke check passed
  (4 task_type mappings, 9 must-not-delegate rejections [7 v0.5.6 + 2 v0.5.7],
  strict mode, unknown type, primary_artifact marker, non-match baseline,
  model_tier main keywords, model_tier default small, model_tier explicit
  override, choose_role propagates tier, enforce_subagent_response happy + violation).

$ python3 workflow-source/tests/check_contract_v1_multi_component.py
Contract v1 §4.2/§5.2 multi-component smoke check passed
  (..., enforce_fanin_response prefix mismatch,
  parent_delegation_id missing raise,
  sub_id duplicate status-inconsistency).
```

### 3.2 audit 5-step (Rev 4) — wire 가이드 stub 결합

```
1. delegator.py:        0 matches for validate_output|validate_fanin_output  PASS
2. output_validator.py: 1 match each for enforce_subagent_response|enforce_fanin_response  PASS
3. __init__.py:         5 matches for enforce_*  (>= 2)  PASS
4. wire_guide.md:       4 matches for enforce_*  (>= 1)  PASS
5. OutputValidationFailed|log_violation:  0 in wire_guide.md  PASS
```

### 3.3 전체 smoke (회귀 0)

```text
$ for t in workflow-source/tests/check_*.py; do python3 "$t"; done
[신규 #1 + #2 test 5종, ADR 3건의 smoke, wire 가이드 walkthrough,
 문서 metadata 124 files 모두 PASS]
```

## 4. 사용 예시

### 4.1 Mavis 측 wire (신규)

```python
# mavis/orchestrator.py
from workflow_kit.contract_v1 import (
    choose_role, enforce_subagent_response, DelegationRejected,
)

def delegate_to_subagent(task: dict, sub_agent_caller) -> dict:
    decision = choose_role(task)
    if decision.must_not_delegate:
        raise DelegationRejected(decision)
    payload = {
        "contract_version": "1.0",
        "delegation_id": decision.delegation_id,
        # ...
    }
    response = sub_agent_caller(payload)
    enforce_subagent_response(
        response, expected_delegation_id=decision.delegation_id,
    )
    return response["result"]
```

### 4.2 fan-out/in (신규)

```python
from workflow_kit.contract_v1 import choose_roles, enforce_fanin_response

def fanout_to_subs(task: dict, sub_agent_caller) -> dict:
    decisions = choose_roles(task, strict=True)
    parent = decisions[0]
    sub_payloads = []
    sub_responses = []
    for sub in decisions[1:]:
        sub_payload = {
            "contract_version": "1.0",
            "delegation_id": sub.delegation_id,
            "parent_delegation_id": parent.delegation_id,
            # ...
        }
        sub_payloads.append(sub_payload)
        sub_responses.append(sub_agent_caller(sub_payload))
    fanin_payload = build_fanin(parent, sub_payloads, sub_responses)
    enforce_fanin_response(
        fanin_payload, expected_parent_delegation_id=parent.delegation_id,
    )
    return fanin_payload
```

### 4.3 CI 환경 bootstrap

```bash
python3 -m bootstrap_lib \
  --target-root "$REPO" \
  --project-slug "$SLUG" \
  --harness opencode \
  --no-interactive \
  --adoption-mode existing
```

## 5. 의도적 비-변경

- `--no-interactive` 동작: v0.5.8 의 SystemExit(1) + 6개 harness 목록 fail-fast 그대로 유지. 본 릴리스는 문서 보강만.
- bootstrap / apply_workflow_upgrade / MCP CLI 기존 flag: 변경 없음.
- v0.5.10.1 smart update 정책 / PRESERVE_RELATIVE_PATHS: 변경 없음.
- Mavis 측 opt-out 시나리오: **금지**. P0 enforcement 은 자동.

## 6. Known limitations (v0.5.11 범위 외)

- **sub_id uniqueness 옵션 (b)**: `choose_roles` 에 dedup check 추가. v0.5.12+ 별도 plan.
- **contract v2 streaming/observability**: §7 spec 설계 + PoC. v0.6+ 별도 plan.
- **추가 하네스 오버레이** (Claude Code, Cursor, Windsurf): HARNESS_SPECS 확장. v0.6+ 별도 plan.
- **`--report <path>` / `--preserve-user-edits`** 등 v0.5.10.1 의 후속 옵션: v0.5.12+.

## 7. 다음 단계

- **v0.5.12** (단기): sub_id uniqueness 옵션 (b) PoC + Mavis 측 integrate_test 추가. Beta-v0.5.12.md.
- **v0.6** (장기): contract v2 streaming, 추가 하네스 오버레이. Beta-v0.6.md.
