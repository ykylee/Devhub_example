---
type: concept
status: active
last_ingested_from: workflow-source/core/stage_gate_pattern.md + workflow_kit/common/contracts/stage_gate.py
related_pages: [concepts/question-file-format, concepts/contract-v1-output-validation, decisions/adr-001-3-layer-separation, patterns/r4-anchor-index, topics/aidlc-benchmark-analysis-2026-06-12]
created: 2026-06-12
updated: 2026-06-12
---

# Stage Gate Pattern (v0.6.4, AIDLC 차용)

- 문서 목적: standard_ai_workflow v0.6.4 의 Stage Gate 명시화 패턴 (AIDLC construction phase 의 2-option completion message 차용) 의 외부 markdown spec + Python enforcement helper 결합 정책을 정리한다. NO EMERGENT BEHAVIOR (3-option / 4-option ❌) + audit log append-only + auto-approval 한계.
- 범위: 외부 spec 구조, `StageCompletion` dataclass, `validate_completion` / `require_explicit_approval` / `append_audit_log` / `emit_completion_message` API, gate 위반 시 행동
- 최종 수정일: 2026-06-12

## §1 TL;DR  {#s1-tldr}

| # | 항목 | 값 |
|---|---|---|
| 1 | 외부 spec | `workflow-source/core/stage_gate_pattern.md` (207 lines, v0.6.4 stable) |
| 2 | Python helper | `workflow_kit/common/contracts/stage_gate.py` (335 lines) |
| 3 | smoke test | `workflow-source/tests/check_stage_gate_compliance.py` (318 lines, 15 test PASS) |
| 4 | Source 1차 출처 | AIDLC `awslabs/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/construction/code-generation.md` §14 (2026-06-08 commit `b19c819`) |
| 5 | 도입 버전 | v0.6.4 (commit `25756bb` + `bc16d91`) |
| 6 | Output schema | `output_schema_guide.md` §3.4 stage_completion field |
| 7 | 관련 컨셉 | [[concepts/question-file-format]] (input 단계), [[concepts/contract-v1-output-validation]] (orchestrator-subagent contract) |
| 8 | 관련 토픽 | [[topics/aidlc-benchmark-analysis-2026-06-12]] §4 보완안 (C) |

## §2 왜 Stage Gate 가 필요한가  {#s2-why}

기존 implicit next-step 패턴의 약점:

- skill output 후 다음 stage 가 implicit 진행 → 사용자 검토 시점 모호
- 실수 / 오판 / 누락 늦게 발견 → rework 비용 ↑
- audit log 부재 → approval 추적 어려움

AIDLC 의 construction phase 가 해결:

- **명시적 approval**: 모든 stage 끝에 2-option (Request Changes / Continue) 사용자 결정
- **Audit 가능**: 모든 approval 이 ISO 8601 timestamp + audit log append-only
- **NO EMERGENT BEHAVIOR**: 3-option / 4-option ❌. 2-option 표준 강제

## §3 Stage Completion Message 표준  {#s3-completion-message}

### §3.1 2-Option 형식 (mandatory)

```markdown
# ✅ [Stage Name] Complete

[1-3 line AI summary]

## 📋 Review Required
Please examine the artifacts:
- `[primary output]`
- (optional) `[secondary output]`

## 🚀 What's Next?
You may:

🔧 **Request Changes** - Ask for modifications to [stage name] based on your review

✅ **Continue to Next Stage** - Approve [stage name] and proceed to **[next stage]**
```

### §3.2 3-Option 절대 금지 (NO EMERGENT BEHAVIOR)

다음 패턴은 ❌:
- "Approve & Continue", "Request Changes", "Skip to Next" (3-option)
- 4-option: "Approve", "Request Changes", "Skip", "Save and Exit"
- 사용자 정의 옵션 추가

근거: 일관성 + 자동화 단순성 + audit binary record.

## §4 Output Schema (output_schema_guide.md §3.4)  {#s4-schema}

### §4.1 StageCompletion 8 field

| Field | Type | Description |
|---|---|---|
| `stage_name` | `str` | stage 식별자 (예: `code-generation`) |
| `stage_status` | `Literal["ok", "warning", "error"]` | stage 실행 결과 |
| `next_stage` | `str \| None` | 다음 stage 이름. workflow 끝이면 `None` |
| `requested_changes` | `list[str]` | user 요청 변경 사항 |
| `approval_timestamp` | `str \| None` | ISO 8601 timestamp |
| `approval_actor` | `Literal["user", "orchestrator", "auto"]` | 승인 주체 |
| `artifacts` | `list[str]` | 검토 대상 artifact path list |
| `notes` | `list[str]` | AI 1-3 line summary |

### §4.2 Pydantic v2 schema

```python
class StageCompletion(BaseModel):
    stage_name: str
    stage_status: Literal["ok", "warning", "error"]
    next_stage: str | None = None
    requested_changes: list[str] = Field(default_factory=list)
    approval_timestamp: str | None = None
    approval_actor: Literal["user", "orchestrator", "auto"] | None = None
    artifacts: list[str] = Field(default_factory=list)
    notes: list[str] = Field(default_factory=list)
```

## §5 Python Enforcement  {#s5-python}

### §5.1 API 표면

| 함수 | 용도 |
|---|---|
| `StageCompletion(...)` | dataclass (8 field) |
| `is_approved()` | gate 통과 여부 (requested_changes 비어있고 timestamp/actor 모두 있어야) |
| `validate_completion(completion)` | 8 field 검증 |
| `require_explicit_approval(completion, env, is_production_code, is_state_doc, is_release)` | auto-approval 한계 검사 |
| `append_audit_log(audit_path, completion)` | append-only 기록 (overwrite ❌) |
| `emit_completion_message(stage_name, artifacts, next_stage, notes, stage_status)` | 정공법 format emit |
| `normalize_option_label(label)` | AIDLC "Approve & Continue" → "continue" 정규화 |

### §5.2 Auto-approval 정책

| Env | Allowed? | 비고 |
|---|---|---|
| `dev` | ❌ (user mandatory) | local edit |
| `ci` / `ci/cd` | ✅ only if not production code / state doc / release | timeout |
| `cron` | ✅ | scheduled job |
| `p0_hotfix` | ✅ only if notes 에 "P0 hotfix" keyword | 긴급 패치 |

**Hard blocked auto-approval** (어떤 env):
- Production 코드 변경
- State 문서 (handoff, backlog, state.json) 갱신
- Release / version tag

## §6 Audit Log 통합  {#s6-audit}

`ai-workflow/memory/active/audit.md` (또는 project 가 정의한 audit log) 에 append-only 기록:

```markdown
## [Stage: code-generation] [2026-06-12T22:30:15Z]
**Stage**: code-generation
**Status**: ok
**Artifacts**:
- src/services/user.py (created)
- src/services/user.py.test (created)
**Approval**: approved
**Actor**: user (yklee)
**Notes**: 3-method implementation, all unit tests pass

---
```

Audit log 정책:
- 기존 entry overwrite ❌
- 새 entry 항상 append
- ISO 8601 timestamp mandatory
- raw user input 그대로 (요약 ❌)
- Stage context 명시

## §7 Stage Gate 위반 시 행동  {#s7-violation}

다음 중 1개:

- **Error**: "stage gate not approved: [stage_name]. Explicit approval required."
- **Wait**: agent 가 사용자에게 prompt 반복 (max 1회)
- **Skip (CI only)**: CI/CD 환경에서 timeout 시 → `approval_actor: "auto"`, `notes` 에 사유 명시

## §8 11종 Skill 별 Stage Name  {#s8-skill-stages}

| Skill | Stage Name | Next Stage (typical) | v0.6.5 spec 적용 |
|---|---|---|---|
| `session-start` | `session-start` | (없음 — task selection) | ✅ spec 적용 (commit `5b16517`) |
| `backlog-update` | `backlog-update` | (없음 — task execution) | ✅ spec 적용 |
| `doc-sync` | `doc-sync` | `validation-plan` 또는 다음 skill | ✅ spec 적용 |
| `merge-doc-reconcile` | `merge-doc-reconcile` | (없음) | ✅ spec 적용 |
| `validation-plan` | `validation-plan` | `code-index-update` 또는 task execution | ✅ spec 적용 |
| `code-index-update` | `code-index-update` | (없음) | ✅ spec 적용 |
| `workflow-linter` | `workflow-linter` | (없음) | ✅ SKILL.md cross-ref |
| `project-status-assessment` | `project-status-assessment` | (없음) | ✅ SKILL.md cross-ref |
| `git-conflict-resolver` | `git-conflict-resolver` | (없음) | ✅ SKILL.md cross-ref |
| `robust-patcher` | `robust-patcher` | `validation-plan` | ✅ SKILL.md cross-ref |
| `automated-repro-scaffold` | `automated-repro-scaffold` | `validation-plan` | ✅ spec 적용 |
| `memory-freeze` (v0.6.1+ 신규) | `memory-freeze` | (없음) | ✅ SKILL.md cross-ref |

v0.6.5 (commit `5b16517`) 의 일괄 적용:
- 7종 spec (§4 출력 계약 끝에 `### 4.1. stage_completion` subsection 추가, 26 line each)
- 5종 SKILL.md cross-ref (14 line each)
- `workflow_skill_catalog.md` §5.2 신규 (+25 line)

## §9 Question File Format 과의 결합  {#s9-binding}

| 단계 | 패턴 | Gate |
|---|---|---|
| Stage 시작 | Question File Format (multi-choice + tag + ambiguity/contradiction auto) | 모든 `[Answer]:` 채워질 때까지 시작 ❌ |
| Stage 실행 | (skill/MCP) | (중간) |
| Stage 종료 | Stage Gate (2-option + audit log) | approval_timestamp/actor 모두 있어야 다음 stage 진행 |

AIDLC 의 audit.md 가 이 결합을 자연스럽게 support.

## §10 한계와 예외  {#s10-limitations}

- **긴급 hotfix (P0)**: stage gate skip 가능, `notes: ["P0 hotfix: skip gate, reason: ..."]` mandatory
- **Read-only operation**: question audit 안 함 (예: `read_only` MCP family)
- **CI/CD 자동화**: timeout 시 auto-approval 정책 별도 정의 (project profile 에 명시)

## §11 Related Decisions / Patterns  {#s11-related}

- [[concepts/question-file-format]] — input 단계 패턴
- [[concepts/contract-v1-output-validation]] — orchestrator-subagent contract v1 (StageCompletion 과 별도)
- [[decisions/adr-001-3-layer-separation]] — Source/Runtime/Project Docs 3-layer
- [[patterns/r4-anchor-index]] — wiki anchor 기반 index
- [[topics/aidlc-benchmark-analysis-2026-06-12]] — AIDLC 벤치마크 분석 (C.Stage Gate 도입 근거)

## §12 References  {#s12-references}

- 외부 spec: [`../../workflow-source/core/stage_gate_pattern.md`](../../workflow-source/core/stage_gate_pattern.md)
- Output schema: [`../../workflow-source/core/output_schema_guide.md`](../../workflow-source/core/output_schema_guide.md) §3.4
- Python helper: [`../../workflow-source/workflow_kit/common/contracts/stage_gate.py`](../../workflow-source/workflow_kit/common/contracts/stage_gate.py)
- **v0.6.5 Runtime migration**: [`../../workflow-source/core/stage_gate_runtime_migration.md`](../../workflow-source/core/stage_gate_runtime_migration.md), [`../../workflow-source/workflow_kit/common/contracts/stage_gate_runtime.py`](../../workflow-source/workflow_kit/common/contracts/stage_gate_runtime.py)
- smoke test: [`../../workflow-source/tests/check_stage_gate_compliance.py`](../../workflow-source/tests/check_stage_gate_compliance.py) (15 test) + [`../../workflow-source/tests/check_stage_gate_runtime.py`](../../workflow-source/tests/check_stage_gate_runtime.py) (13 test, v0.6.5 신규)
- AIDLC 원본: `/Users/yklee/repos/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/construction/code-generation.md` §14
