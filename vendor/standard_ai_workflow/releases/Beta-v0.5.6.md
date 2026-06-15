# Beta v0.5.6 — Contract v1 §5/§6 P0 Enforcement (validator + delegator)

- **릴리스 일자**: 2026-06-07
- **브랜치**: `release/v0.5.6`
- **포함 커밋**: v0.5.5 (PR #19) 이후 1 squash
- **상태**: ✅ 1개 TASK 완료 (TASK-V056-001)

## 1. 하이라이트

v0.5.5 Phase 11 pilot 에서 도출된 P0 priorities 2개를 묶어서 1 PR 로 enforce:

- **TASK-V056-001-A**: sub-agent 측 §5 출력 스키마 validator — `workflow_kit.contract_v1.output_validator.validate_output(payload, expected_delegation_id=None) -> OutputValidationResult`
- **TASK-V056-001-B**: Mavis 측 §6.1 자동 위임 결정 delegator — `workflow_kit.contract_v1.delegator.choose_role(task, strict=False) -> DelegationDecision`
- **신규 회귀 2개**: `check_contract_v1_output_validator.py` (4 pilot baseline + 8 violation + 4 pilot doc §5 output) · `check_contract_v1_delegator.py` (4 task_type mapping + 7 §6.3 rejection + strict + primary_artifact marker + non-match baseline)
- **contract v1 spec §5/§6 갭 발견 & 수정**: `artifact_kind` enum 에 `code` 추가 (v0.5.5 pilot 의 PR #492, #491 에서 사용)

## 2. TASK-V056-001-A: §5 출력 Validator

### 동기

v0.5.5 pilot 의 4 시나리오에서 contract v1 §4/§5 JSON round-trip 이 모두 PASS 했지만, **회귀 검증**이 자동화되어 있지 않음. v0.5.5 P0 도출: sub-agent 응답이 spec 과 fit 한지 자동 검증. 검증 실패 시 §7.4 "출력 스키마 위반" 정책 적용 (orchestrator 가 위반 보고 + 재위임 또는 직접 처리).

### 산출물

[`workflow_kit/contract_v1/output_validator.py`](../workflow_kit/contract_v1/output_validator.py) (신규, ~180줄, pure Python)

핵심 API:

```python
from workflow_kit.contract_v1 import validate_output, OutputValidationError, OutputValidationResult

result = validate_output(
    payload,                          # sub-agent 가 응답한 dict
    expected_delegation_id="del-...", # orchestrator 가 위임한 delegation_id
)
if not result.is_valid:
    # §7.4 정책: 위반 보고 + 재위임 1회 → 직접 처리
    for err in result.errors:
        log(f"{err.field}: {err.reason}")
```

검증 항목 (총 9):
1. `contract_version == "1.0"`
2. 필수 top-level 필드 5개: `contract_version`, `delegation_id`, `completed_at`, `worker`, `result`
3. `worker.role` ∈ `{doc-worker, code-worker, validation-worker, workflow-worker}`
4. `worker.model_tier` ∈ `{small, main}`
5. `result.status` ∈ `{ok, partial, failed}`
6. `result.artifacts[i].kind` ∈ `{markdown, python, json, toml, text, code, other}` (v0.5.6 갭 수정)
7. `delegation_id` 가 expected 와 일치 (cross-delegation leak 방지)
8. `result.validation_result.ran` is bool + `validation_result.status` ∈ `{pass, fail, skipped}` when `ran=true`
9. `result.next_step` is non-empty string

### 회귀

[`tests/check_contract_v1_output_validator.py`](../tests/check_contract_v1_output_validator.py) (신규, 13 check)

- 4 pilot baseline (PR #493/#492/#491/#490) — 모두 valid
- 8 violation: missing field / wrong contract_version / invalid status enum / invalid role enum / invalid validation_result.status / delegation_id mismatch / empty next_step / invalid artifact kind
- v0.5.5 pilot doc (`pilot_phase11_devhub_contract_v1.md`) 의 실제 §5 JSON 4개 — 모두 valid (regression baseline)

## 3. TASK-V056-001-B: §6.1 / §6.3 Delegator

### 동기

v0.5.4 contract v1 spec 의 §6 카탈로그는 "권장" 이라 강제력이 없었음. P0 도출: orchestrator 가 매번 손으로 §6.1/§6.3 을 판단하지 말고, 자동 enforce. v0.5.5 pilot 의 §6 정합성 85% 도 자동화면 100% 가능.

### 산출물

[`workflow_kit/contract_v1/delegator.py`](../workflow_kit/contract_v1/delegator.py) (신규, ~130줄, pure Python)

핵심 API:

```python
from workflow_kit.contract_v1 import choose_role, DelegationDecision, DelegationRejected

decision = choose_role(task)  # task 는 §4 입력 dict
if decision.must_not_delegate:
    # §6.3 매치 — orchestrator 직접 처리
    log(f"Rejected: {decision.rejected_reason}")
else:
    sub_agent_prompt(decision.delegation_id, decision.role, task)

# strict 모드: §6.3 매치 시 raise
try:
    choose_role(task, strict=True)
except DelegationRejected as e:
    log(f"Must not delegate: {e.decision.rejected_reason}")
```

검증 항목:
- `task.task_type` ∈ `{doc_draft, code_change, validation_run, bounded_research}`
- `task_type` → `role` 매핑 (deterministic, §6.1):
  - `doc_draft` / `bounded_research` → `doc-worker`
  - `code_change` → `code-worker`
  - `validation_run` → `validation-worker`
- §6.3 MUST-NOT-delegate 7개 marker 거부: `handoff`, `backlog`, `state.json`, `ask_user`, `우선순위`, `통합/리뷰`, `PR 본문` (brief 와 primary_artifact 양쪽 검사)
- `delegation_id` 자동 생성: `del-YYYY-MM-DD-xxxxxxxx` (8자리 hex suffix)

### 회귀

[`tests/check_contract_v1_delegator.py`](../tests/check_contract_v1_delegator.py) (신규, 6 check)

- 4 task_type → role 매핑 + delegation_id 형식
- 7 §6.3 marker 거부 (각 marker 의 reason 메시지 검증)
- `strict=True` 모드에서 `DelegationRejected` raise
- 알 수 없는 task_type → `ValueError`
- primary_artifact 의 marker 도 거부
- negative control: "비교" 가 "통합/리뷰" 의 일부 substring 이지만 reject 안 됨 (false positive 방지)

## 4. contract v1 spec 갭 수정 (v0.5.6 discovery)

v0.5.5 pilot doc 의 `kind: "code"` 가 §5 enum 에 없어 validator 가 거부. spec 갭:

| 필드 | Before | After |
| --- | --- | --- |
| `task.expected_outputs.artifact_kind` (contract v1 §4) | `markdown` / `python` / `json` / `toml` / `text` / `other` | + `code` 추가 |
| `ALLOWED_ARTIFACT_KINDS` (validator) | 동상 | + `code` 추가 |

PR #492 (Go code) + PR #491 (frontend TS) 의 artifact 가 `kind: "code"` 이므로, contract v1 spec + validator enum 둘 다 수정. v0.5.5 pilot 의 4 시나리오 모두 다시 validate PASS 확인.

## 5. 패키지화

[`workflow-source/pyproject.toml`](../pyproject.toml) 에 `workflow_kit.contract_v1` sub-package 추가. `pip install -e workflow-source` 후 `from workflow_kit.contract_v1 import validate_output, choose_role` 가능.

## 6. 회귀 테스트 종합

| 테스트 | 결과 |
| --- | --- |
| `check_bootstrap.py` (8 모드) | ✅ 8/8 |
| `check_bootstrap_mcp_roundtrip.py` (5 하네스) | ✅ 5/5 |
| `check_docs.py` | ✅ 108+ markdown PASS |
| `check_contract_v1_roundtrip.py` (S1) | ✅ |
| `check_contract_v1_role_mapping.py` (S2) | ✅ |
| `check_contract_v1_direct_only.py` (S3) | ✅ |
| `check_contract_v1_output_validator.py` (v0.5.6 신규) | ✅ (4 pilot + 8 violation + 4 pilot doc) |
| `check_contract_v1_delegator.py` (v0.5.6 신규) | ✅ (4 mapping + 7 rejection + strict + 2 baseline) |
| `check_pilot_phase11_contract_v1.py` (v0.5.5) | ✅ |
| Workflow linter | ✅ 0 issues |

## 7. 메모리 layer

v0.5.6 의 모든 작업은 `ai-workflow/memory/release/v0.5.6/` 에 기록:
- `PROJECT_PROFILE.md` — v0.5.6 프로젝트 프로파일 (validator/delegator 명령 추가)
- `session_handoff.md` — 세션 인계
- `backlog/2026-06-07.md` — TASK 상세
- `state.json` — workflow 상태 캐시

## 8. 다음 단계 (v0.5.7 후보)

- [ ] **P1**: contract v2 fan-out/in (§11 1차 컷) — `delegator.choose_role` 의 멀티 sub-task 지원, parent orchestrator 가 fan-in
- [ ] **P1**: §6.1 "cross-ref 갱신" 명시 row 추가
- [ ] **P2**: `task.required_model_tier` 필드 (delegator 가 자동 main/small 결정)
- [ ] **P2**: `artifacts` 의 `added`/`removed`/`total` 분리 (validator 갭 발견 시 추가)
- [ ] Mavis 측 `delegator.choose_role` 을 실제 위임 결정에 wire (v0.5.6 는 모듈만, 런타임 wire 은 별도)
- [ ] PyPI 배포 (v0.5.2 부터 보류 중)
- [ ] Phase 11 case study 추가 (다른 외부 저장소, my_harness 등)

## 9. v0.5.0 → v0.5.6 변경 통계

```
v0.5.0 (PR #13): 42/42 smoke + MiniMax Code harness overlay
v0.5.1 (PR #14): per-harness MCP install + auto-emit + guide
v0.5.1 (PR #15): check_bootstrap_mcp_roundtrip + 5 harnesses round-trip
v0.5.2 (PR #16): bootstrap 풀 리팩터 + 패키지화 + pilot validation
v0.5.3 (PR #17): antigravity MCP 표준화 + cross-language stack 표시
v0.5.4 (PR #18): orchestrator ↔ sub-agent contract v1 + maturity sync + baseline sync
v0.5.5 (PR #19): Phase 11 본격 pilot (Devhub Example × Contract v1)
v0.5.6 (PR #20): contract v1 §5/§6 P0 enforcement (validator + delegator)
```

총 8 PR.

## 10. 이슈 트래커

- ✅ [issue #1](https://github.com/ykylee/standard_ai_workflow/issues/1) 영구 해결 (v0.5.4)
- 0 open issues
