# Stage Gate Pattern

- 문서 목적: 표준 AI 워크플로우의 모든 skill/MCP output 후 사용자의 explicit 2-option approval 을 gate 로 강제한다. AIDLC 의 construction phase completion message 패턴 차용.
- 범위: completion message 표준 형식, `output_schema_guide.md` 의 `stage_completion` 필드, audit log 통합, gate 위반 시 행동
- 대상 독자: 워크플로우 skill/MCP 구현자, AI agent, 운영자
- 상태: stable (v0.6.4 도입)
- 최종 수정일: 2026-06-12
- 관련 문서: `./workflow_adoption_entrypoints.md` §7.2, `./question_file_format.md`, `./output_schema_guide.md` §3.2, AIDLC `awslabs/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/construction/code-generation.md` (1차 출처)
- 1차 출처: AIDLC construction phase 의 "Standardized 2-option completion message" (2026-06-08 commit `b19c819`)

## 1. 왜 Stage Gate 가 필요한가

기존 패턴:
- skill output 후 다음 stage 가 implicit 하게 진행
- 사용자 검토 시점이 모호 (output 직후? 다음 session?)
- 실수 / 오판 / 누락을 늦게 발견 → rework 비용 ↑

AIDLC 의 construction phase 가 해결하는 문제:
- **명시적 approval**: 모든 stage 끝에 2-option (Request Changes / Continue) 사용자 결정
- **Audit 가능**: 모든 approval 이 ISO 8601 timestamp 와 함께 audit log 에 append
- **NO EMERGENT BEHAVIOR**: 3-option 또는 4-option 메뉴 ❌. 2-option 표준 강제 (일관성)

## 2. Stage Completion Message 표준

### 2.1 2-Option 형식 (mandatory)

모든 stage 완료 시 아래 형식으로 completion message 를 emit 한다.

```markdown
# ✅ [Stage Name] Complete

[1-3 line AI-generated summary of what was accomplished]

## 📋 Review Required
Please examine the [artifact path / output location]:
- `[primary output path]`
- (optional) `[secondary output path]`

## 🚀 What's Next?
You may:

🔧 **Request Changes** - Ask for modifications to [stage name] based on your review

✅ **Continue to Next Stage** - Approve [stage name] and proceed to **[next stage name]**
```

### 2.2 3-Option 절대 금지 (NO EMERGENT BEHAVIOR)

다음 패턴은 ❌:
- "Approve & Continue", "Request Changes", "Skip to Next"
- 4-option: "Approve", "Request Changes", "Skip", "Save and Exit"
- 사용자 정의 옵션 추가

근거:
- 일관성: 모든 stage 가 같은 옵션 → 사용자 학습 비용 ↓
- 자동화: 2-option parser 가 단순 (binary decision)
- Audit: approval record 가 binary (approved | changes_requested)

### 2.3 Stage Completion Message 의 위치

기본: `skill output 의 stdout` (chat 또는 terminal)
선택: `ai-workflow/memory/active/audit.md` 에 append-only 기록 (audit log 통합 시)

## 3. Output Schema (output_schema_guide.md §3.2)

### 3.1 신규 필드

`output_schema_guide.md` §3 의 공통 필드에 `stage_completion` 추가:

```python
class StageCompletion(BaseModel):
    """Skill/MCP output 의 stage completion 승인 필드 (v0.6.4 신규)"""
    stage_name: str                     # 예: "code-generation", "requirements-analysis"
    stage_status: Literal["ok", "warning", "error"]
    next_stage: str | None              # 다음 stage 이름. None 이면 workflow end
    requested_changes: list[str]        # user 가 요청한 변경 사항 (free text, list)
    approval_timestamp: str | None      # ISO 8601. user 가 승인한 시각
    approval_actor: str | None          # user / orchestrator / auto (CI/CD)
    artifacts: list[str]                # 검토 대상 artifact path list
    notes: list[str]                    # AI summary bullet
```

### 3.2 적용 위치

- **모든 skill spec 11종** 의 §X "Output Contract" 에 `stage_completion` 필드 추가
- **모든 MCP tool** 의 response schema 에 `stage_completion` 필드 추가
- **Orchestrator worker delegation** (contract v1) 의 `sub_payloads` 에도 `stage_completion` 포함

### 3.3 11종 skill 별 적용

| Skill | Stage Name | Next Stage (typical) |
|---|---|---|
| `session-start` | `session-start` | (없음 — task selection) |
| `backlog-update` | `backlog-update` | (없음 — task execution) |
| `doc-sync` | `doc-sync` | `validation-plan` 또는 다음 skill |
| `merge-doc-reconcile` | `merge-doc-reconcile` | (없음) |
| `validation-plan` | `validation-plan` | `code-index-update` 또는 task execution |
| `code-index-update` | `code-index-update` | (없음) |
| `workflow-linter` | `workflow-linter` | (없음 — auto fix 권고) |
| `project-status-assessment` | `project-status-assessment` | (없음) |
| `git-conflict-resolver` | `git-conflict-resolver` | (없음) |
| `robust-patcher` | `robust-patcher` | `validation-plan` |
| `automated-repro-scaffold` | `automated-repro-scaffold` | `validation-plan` |
| `memory-freeze` (v0.6.1+ 신규) | `memory-freeze` | (없음) |

## 4. Gate 정책

### 4.1 Gate 위반 시 행동

Stage gate 가 위반되면 (사용자 approval 없음) 다음 stage 자동 진행 ❌. 다음 중 1개 발생:

- **Error**: "stage gate not approved: [stage_name]. Explicit approval required."
- **Wait**: agent 가 사용자에게 prompt 반복 (max 1회)
- **Skip (CI only)**: CI/CD 환경에서 timeout 시 → `approval_actor: "auto"`, `notes: ["gate-skipped: CI timeout"]` 명시 후 진행

### 4.2 Auto-approval 의 한계

다음은 **auto-approval 가능**:
- CI/CD 환경의 명시적 timeout
- cron job (audit log 에 `approval_actor: "auto"` 필수)
- 백업 / hotfix (긴급 패치, `notes` 에 사유 명시)

다음은 **explicit user approval 필수**:
- Production 코드 변경
- State 문서 (handoff, backlog, state.json) 갱신
- Release / version tag
- Cross-project 영향 작업

### 4.3 Audit Log 통합

`ai-workflow/memory/active/audit.md` (또는 project 가 정의한 audit log) 에 append-only:

```markdown
## [Stage: code-generation] [2026-06-12T22:30:15Z]
**Stage**: code-generation
**Status**: ok
**Artifacts**:
- src/services/user-service.ts (created)
- src/services/user-service.test.ts (created)
**Approval**: approved
**Actor**: user (yklee)
**Notes**: 3-method implementation, all unit tests pass

---
```

Audit log 의 append-only 정책:
- 기존 entry overwrite ❌
- 새 entry 항상 append
- ISO 8601 timestamp mandatory
- raw user input 그대로 (요약 ❌)
- Stage context 명시

## 5. 구현 위치

### 5.1 문서

- 본 spec: `workflow-source/core/stage_gate_pattern.md` (이 문서)
- 적용: 각 skill spec 11종 의 §X "Output Contract" 보강, `output_schema_guide.md` §3.2
- cross-ref: `workflow_adoption_entrypoints.md` §7.2, `question_file_format.md` (gate 와의 결합)

### 5.2 코드 (v0.6.4 신규)

- `workflow_kit/common/contracts/stage_gate.py`
  - `StageCompletion` Pydantic v2 model
  - `validate_completion(completion: dict) -> list[ValidationError]`
  - `require_explicit_approval(completion: dict, env: str) -> bool`
  - `append_audit_log(audit_path: Path, completion: StageCompletion)` (append-only)
  - `emit_completion_message(stage_name, artifacts, next_stage) -> str` (정공법 format)

### 5.3 테스트

- `workflow-source/tests/check_stage_gate_compliance.py`
  - 11종 skill spec 의 output schema 에 `stage_completion` 필드 존재 → PASS
  - 2-option 만 사용 (3-option / 4-option / no-option → FAIL)
  - audit log 가 append-only (overwrite 시도 → FAIL)
  - explicit approval 필수 stage 에서 auto-approval → FAIL

## 6. Question File Format 과의 결합

Stage Gate 와 Question File Format 은 **상호 보강**:
- **Question File Format** = stage 시작 전 사용자의 결정 입력 (multi-choice, clarification 가능)
- **Stage Gate** = stage 종료 후 사용자의 결정 입력 (binary, 2-option)

결합 패턴:
1. Stage 시작 → Question File Format 으로 사용자 결정
2. Stage 실행
3. Stage 종료 → Stage Gate 2-option 으로 사용자 결정
4. (선택) 다음 stage 가 다시 Question File Format 으로 시작

AIDLC 의 audit.md 가 이 결합을 자연스럽게 support:
- 모든 question 의 `[Answer]:` 기록
- 모든 stage approval 의 timestamp/actor 기록
- ISO 8601 + raw input

## 7. 한계와 예외

- **긴급 hotfix (P0)**: stage gate skip 가능, 단 `notes: ["P0 hotfix: skip gate, reason: <text>"]` mandatory
- **Read-only operation**: question audit 안 함 (ex: `read_only` MCP family)
- **CI/CD 자동화**: timeout 시 auto-approval 정책 별도 정의 (project profile 에 명시)

## 8. 다음에 읽을 문서

- [`./question_file_format.md`](./question_file_format.md) — Stage Gate 의 input 단계 패턴
- [`./workflow_adoption_entrypoints.md` §7.2](./workflow_adoption_entrypoints.md#72-stage-gate-명시화-패턴--권장)
- [`./output_schema_guide.md`](./output_schema_guide.md) — stage_completion field 의 output schema 위치
- AIDLC 원본: `/Users/yklee/repos/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/construction/code-generation.md` §14
