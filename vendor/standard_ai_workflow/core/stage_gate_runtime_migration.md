# Stage Gate Runtime Migration Guide (v0.6.5)

- 문서 목적: v0.6.4 의 Stage Gate Pattern (이론) → v0.6.5 의 runtime helper + 11종 skill code 변경 절차.
- 범위: helper API, 기존 skill code 변경 절차, breaking change 회피, AIDLC 호환
- 대상 독자: workflow_kit 구현자, skill implementor, 운영자
- 상태: stable (v0.6.5 도입)
- 최종 수정일: 2026-06-12
- 관련 문서: `./stage_gate_pattern.md` (이론), `./question_file_format.md` (입력 단계), `../output_schema_guide.md` §3.4 (schema), AIDLC `awslabs/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/construction/code-generation.md` §14
- 1차 출처: AIDLC construction phase 의 completion message (2026-06-08 commit `b19c819`)

## 1. 왜 Runtime Migration 이 필요한가

v0.6.4 의 Stage Gate Pattern 은 **이론 spec** (markdown 207 lines) + **핵심 helper** (`stage_gate.py` 335 lines) 만 제공. v0.6.5 의 11종 skill spec 보강 (commit 5b16517) 으로 *스펙* 은 SSOT 화됐지만, **runtime** 에서:

- 각 skill 의 `run_*.py` 가 `print(json.dumps(result))` 시 `stage_completion` field 를 자동 merge 하지 못함
- user 가 stage gate 를 통과해도 (approval_timestamp + approval_actor) audit log 자동 기록 안 됨
- orchestrator 측에서 `emit_completion_message` 자동 호출 안 됨

→ **runtime migration** 필요: helper 1개 (자동 merge + emit + audit log) + 11종 skill code 변경.

## 2. Runtime Helper (v0.6.5 신규)

### 2.1 `stage_gate_runtime.py` API

| 함수 | 용도 |
|---|---|
| `build_stage_completion(stage_name, stage_status, artifacts, next_stage, notes, ...)` | 8-field dict 생성 |
| `merge_into_result(result, stage_completion, *, overwrite=False)` | 기존 skill result dict 에 merge (기존 field 보존, idempotent) |
| `emit_and_log(stage_name, artifacts, next_stage, *, stage_status, notes, audit_log_path, ...)` | 2-option message emit + (optional) audit log append |
| `is_stage_completion_present(result)` | 8-field 모두 있는 dict 인지 검증 |
| `get_stage_status_from_result(result)` | stage_completion.stage_status 우선, 없으면 legacy 'status' fallback |

### 2.2 사용 예시 (skill output 끝에 적용)

```python
# 기존 run_*.py 의 마지막 부분
import sys
from pathlib import Path
sys.path.insert(0, str(Path(__file__).resolve().parents[4]))
from workflow_kit.common.contracts.stage_gate_runtime import (
    build_stage_completion, merge_into_result, emit_and_log,
)

# ... existing code that builds result dict ...

# 1) stage_completion build + merge
sc = build_stage_completion(
    stage_name="session-start",
    stage_status="ok",
    artifacts=["ai-workflow/memory/active/state.json", "ai-workflow/memory/active/session_handoff.md"],
    next_stage=None,
    notes=["session restored, 3 in_progress items"],
)
result = merge_into_result(result, sc)

# 2) stdout emit (2-option completion message)
print(json.dumps(result, ensure_ascii=False, indent=2))
print()
print(emit_and_log(
    stage_name="session-start",
    artifacts=sc["artifacts"],
    next_stage=None,
    notes=sc["notes"],
    stage_status="ok",
    audit_log_path="ai-workflow/memory/active/audit.md",  # optional
    approval_timestamp=result.get("approval_timestamp"),  # orchestrator 가 채움
    approval_actor=result.get("approval_actor"),
))
```

## 3. 11종 Skill Code 변경 절차

### 3.1 변경 순서 (안전 순서)

1. **변경 전 baseline smoke test 52개 모두 PASS 확인** — `python3 workflow-source/tests/check_*.py` (mock 사용 가능 환경에서)
2. **1 skill 부터 pilot** (가장 simple 한 `automated-repro-scaffold` 권장) — stage_completion merge + emit_and_log 추가
3. **1 skill 변경 후 52 smoke test 회귀 검증** — 1 commit 으로 관리
4. **나머지 10 skill 순차 적용** — 각 1 commit
5. **L1 wiki + L2 vault 동기화** — 변경 사실 wiki log + concept page 갱신

### 3.2 Breaking Change 회피 정책

**중요**: stage_completion 은 **optional field** 로 추가 (mandatory 아님). 이유:

- 기존 52 smoke test 의 schema validator (e.g., `check_output_samples.py`, `check_generated_schema_validation.py`) 가 `required_keys` 검사. `stage_completion` 을 required 로 추가하면 즉시 break.
- v0.6.5 의 spec 적용 (`commit 5b16517`) 도 spec doc 에 "optional" 명시. runtime 도 동일 정책.
- **점진적 적용**: 1 skill 씩 변경하면서 회귀 검증. 모두 적용된 후 (v0.7.0?) required 로 격상.

### 3.3 변경 template (per skill)

각 skill 의 `scripts/run_*.py` 끝부분에 다음 3-7 line 추가:

```python
# === v0.6.5 stage_completion integration ===
from workflow_kit.common.contracts.stage_gate_runtime import (
    build_stage_completion, merge_into_result, emit_and_log,
)

STAGE_NAME = "<skill-name>"  # 예: "session-start", "backlog-update"
NEXT_STAGE = "<next>" or None  # None 또는 다음 stage 이름
ARTIFACTS = ["<primary output path>", ...]  # 검토 대상

result = merge_into_result(result, build_stage_completion(
    stage_name=STAGE_NAME,
    stage_status="ok" if result["status"] == "ok" else result["status"],
    artifacts=ARTIFACTS,
    next_stage=NEXT_STAGE,
    notes=[result.get("summary", "")[:200]] if result.get("summary") else [],
))
```

Stage name 매핑 (skill_catalog.md §5.2 와 일치):

| Skill | Stage Name | Next Stage | Artifacts (예시) |
|---|---|---|---|
| `session-start` | `session-start` | (None) | `state.json`, `session_handoff.md` |
| `backlog-update` | `backlog-update` | (None) | `backlog/<target_date>.md`, `work_backlog.md` |
| `doc-sync` | `doc-sync` | `validation-plan` | `session_handoff.md` |
| `merge-doc-reconcile` | `merge-doc-reconcile` | (None) | `session_handoff.md`, `work_backlog.md` |
| `validation-plan` | `validation-plan` | `code-index-update` | `backlog/<target_date>.md` |
| `code-index-update` | `code-index-update` | (None) | `README.md`, `session_handoff.md` |
| `automated-repro-scaffold` | `automated-repro-scaffold` | `validation-plan` | `repro_<bug_id>.py` |
| `workflow-linter` | `workflow-linter` | (None) | (lint report 자체) |
| `project-status-assessment` | `project-status-assessment` | (None) | (assessment report 자체) |
| `memory-freeze` | `memory-freeze` | (None) | `archive/YYYY-MM-DD/` |
| `git-conflict-resolver` | `git-conflict-resolver` | (None) | (resolved files) |
| `robust-patcher` | `robust-patcher` | `validation-plan` | (patched files) |

## 4. Smoke Test 추가 (v0.6.5 신규)

`workflow-source/tests/check_stage_gate_runtime.py` (13 test, 모두 PASS):
- build_stage_completion_basic / with_approval
- merge_into_result_preserves_existing / idempotent / overwrite
- is_stage_completion_present_valid / missing
- get_stage_status_prefers_stage_completion / falls_back_to_legacy
- emit_and_log_basic / with_audit_log / audit_log_pending
- existing_result_without_stage_completion_still_valid (52 smoke test 호환)

## 5. AIDLC 호환

본 runtime migration 후 AIDLC 의 construction phase 와 동등:

- AIDLC §14: 2-option completion message (Request Changes / Continue to Next Stage) ✅
- AIDLC §15: Wait for Explicit Approval (gate 위반 시 자동 진행 ❌) ✅
- AIDLC audit log: append-only, ISO 8601 timestamp, raw user input ✅
- AIDLC NO EMERGENT BEHAVIOR: 2-option 만 ✅

차이점 (의도적):
- 우리 11종 skill 의 **stage_completion 은 optional field** (v0.6.5). AIDLC 는 mandatory.
- 우리 v0.6.5+ 적용 후 점진적으로 required 로 격상 (모두 적용 후 v0.7.0+).
- 우리 audit log 는 `ai-workflow/memory/active/audit.md` 채널 (project state meta layer). AIDLC 는 `aidlc-docs/audit.md` (in-repo). 위치 차이만.

## 6. Follow-up

- **v0.6.5 다음 commit**: 11종 skill code 변경 (1-2 commit, 1 skill 씩 또는 batch)
- **v0.7.0**: stage_completion 을 required 로 격상 + AIDLC 완전 호환 검증
- **v0.8.0** (별도): orchestrator 측 자동 emit_and_log 통합 (skill output 직후 user prompt → approval → audit log 자동)

## 7. 다음에 읽을 문서

- [`./stage_gate_pattern.md`](./stage_gate_pattern.md) — 이론 (StageCompletion 8 field, gate 정책, audit log)
- [`./output_schema_guide.md` §3.4](./output_schema_guide.md) — stage_completion output schema
- [`../workflow_kit/common/contracts/stage_gate.py`](../workflow_kit/common/contracts/stage_gate.py) — 핵심 helper
- [`../workflow_kit/common/contracts/stage_gate_runtime.py`](../workflow_kit/common/contracts/stage_gate_runtime.py) — runtime migration helper
- [`../tests/check_stage_gate_runtime.py`](../tests/check_stage_gate_runtime.py) — 13 test
- AIDLC 원본: `/Users/yklee/repos/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/construction/code-generation.md` §14
