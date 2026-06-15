# Audit Log Standard (v0.7.0)

- 문서 목적: v0.6.4-5 의 Stage Gate Pattern + Runtime Migration 가이드에 분산 정의된 audit log 정책을 **단일 표준** 으로 통합. v0.7.0 부터 모든 skill/MCP output 의 stage_completion 승인 내역 + Q&A 결정 + state 변경을 audit log 에 일관되게 기록.
- 범위: 파일 위치, format, lifecycle 정책, append-only 강제, 자동화 hook
- 대상 독자: workflow 설계자, AI agent, 운영자, compliance 검토자
- 상태: stable (v0.7.0 도입)
- 최종 수정일: 2026-06-12
- 관련 문서: `./stage_gate_pattern.md` §6 (분산 정의), `./stage_gate_runtime_migration.md` §5, `./output_schema_guide.md` §3.4, [`../workflow_kit/common/contracts/stage_gate.py`](../workflow_kit/common/contracts/stage_gate.py) `append_audit_log`
- 1차 출처: AIDLC `awslabs/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/common/process-overview.md` + `inception/requirements-analysis.md` 의 audit.md 정책 (commit `b19c819`, 2026-06-08)

## 1. 왜 Audit Log 표준이 필요한가

v0.6.4-5 에서 stage_completion 도입 후 audit log 가 *선택적* 정책이었음. 결과:
- 12/12 skill runtime 적용 후에도 audit log 가 항상 기록되는지 불확실
- file 위치가 skill 별로 다를 수 있음 (`audit.md`, `state_audit.md`, `state_changes.log` 등)
- format 이 skill 별로 다를 수 있음
- overwrite 가능성 (단순 `print()` 기반 기록 시)

v0.7.0 의 mandatory stage_completion 정책과 정합 — **audit log 도 mandatory** 로 표준화.

## 2. Audit Log 파일 위치 (Standard)

### 2.1 Primary 위치

`<project_root>/ai-workflow/memory/active/audit.md`

이유:
- `ai-workflow/memory/active/` 는 workflow state meta layer (global_workflow_standard.md §1.2, R-1 정책)
- `state.json`, `session_handoff.md`, `work_backlog.md` 와 같은 layer
- `gitignore` 의 selective 추적 (사용자 state 는 git 추적 X)
- 프로젝트 bootstrap 시 자동 생성 (bootstrap_lib)

### 2.2 Fallback / Alternative 위치

| Case | 위치 | 이유 |
|---|---|---|
| Multi-agent / orchestrator | `ai-workflow/memory/active/audit/orchestrator.md` | sub-agent 별 분리 |
| Cron / P0 hotfix | `ai-workflow/memory/active/audit/auto-approvals.md` | auto-approval 만 별도 추적 (compliance) |
| Per-skill | `ai-workflow/memory/active/audit/<skill_name>.md` | skill 별 detail log (optional) |

기본: single `audit.md`. alternative 은 project 가 명시적으로 결정.

## 3. Audit Log Format (Standard)

### 3.1 Entry Format (Mandatory)

```markdown
## [Stage: <stage_name>] [<ISO_8601_timestamp>]
**Stage**: <stage_name>
**Status**: <ok|warning|error>
**Artifacts**:
- <artifact_path_1>
- <artifact_path_2>
**Approval**: <approved|pending>
**Actor**: <user|orchestrator|auto>
**Notes**: <one-line summary>

---
```

### 3.2 Field 정의

| Field | 필수 | 형식 | 비고 |
|---|---|---|---|
| Header line | ✅ | `## [Stage: X] [ISO_8601]` | 둘 다 있어야 grep 가능 |
| Stage | ✅ | stage_name | `code-generation`, `session-start`, ... |
| Status | ✅ | `ok` / `warning` / `error` | stage 실행 결과 |
| Artifacts | (선택) | path list | 검토 대상 |
| Approval | ✅ | `approved` / `pending` | gate 통과 여부 |
| Actor | ✅ | `user` / `orchestrator` / `auto` | 승인 주체 |
| Notes | ✅ | 1-line summary | AI 가 사용자에게 보여주는 summary |
| separator | ✅ | `---` | entry 간 visual 구분 |
| timestamp | ✅ | ISO 8601 (YYYY-MM-DDTHH:MM:SSZ) | UTC 권장 |

### 3.3 Raw User Input 캡처 (AIDLC 준수)

사용자 응답 (raw text, 요약 X) 그대로 보존. 우리 `stage_gate.emit_and_log` 또는 `append_audit_log` 가 `notes` field 에 1-line summary 만 적고, **full raw text 는 stage_completion.requested_changes + 별도 Q&A file 에 보존**. audit log 자체는 *summary-only* 정책.

**이유**:
- audit log 가 매 entry 마다 user 의 full chat message 를 포함하면 비대해짐
- 우리 분리 정책: *decision* (raw text) 는 Q&A file / requested_changes field, *event* (what happened) 는 audit log

**Q&A 의 경우** 별도 채널:
- `ai-workflow/memory/active/questions/<phase>-questions.md` (full Q&A + answers, append-only)
- audit log 에는 "Q&A performed" 한 줄 + reference

## 4. Append-Only 강제 (CRITICAL)

### 4.1 정책

- **절대 overwrite 금지** — 기존 entry 보존
- 새 entry 는 항상 append (file 끝에 추가)
- 잘못된 entry 시: 별도 `<entry-id>.correction.md` 로 추가 (in-place fix ❌)
- ISO 8601 timestamp mandatory — chronological ordering 보장
- `completed` entry 의 `requested_changes` 가 비어있지 않으면 별도 `changelog` 또는 같은 audit log 의 *next* entry 가 변경 내역 기록

### 4.2 우리 구현

`workflow_kit.common.contracts.stage_gate.append_audit_log(audit_path, completion)`:
```python
def append_audit_log(audit_path: Path | str, completion: StageCompletion) -> None:
    # 기존 file read → append → write (overwrite 절대 안 함)
    existing = ""
    if p.exists():
        existing = p.read_text(encoding="utf-8")
    # 새 entry 생성 (ts 포함)
    entry = _format_entry(completion, ts=completion.approval_timestamp or now_utc_iso())
    p.write_text(existing + entry, encoding="utf-8")
```

### 4.3 검증

`tests/check_stage_gate_compliance.py` (v0.6.4, 15 test):
- `audit_log_append_only` test — 두 번 호출 시 첫 entry 보존
- ISO 8601 timestamp 형식
- `requested_changes` / `approval_timestamp` / `approval_actor` 정합성

## 5. Lifecycle 정책

### 5.1 Retention

- **Default**: 무기한 보존 (append-only)
- **Rotation**: file size > 10 MB 시 `audit.<timestamp>.md` 로 archive + 새 `audit.md` 시작
  - 우리 구현: `audit.2026-07-01T00-00-00Z.md` 형식 (Z → '-' 변환)
  - 자동 rotation 미구현 (v0.7.0+ 후보) — manual 또는 cron

### 5.2 Freeze / Archive

R8 (Memory Raw Freeze) 와 함께: session 종료 시 `ai-workflow/memory/active/` 의 모든 state 가 `archive/<date>/` 로 freeze. **audit.md 도 freeze 대상**:
- `archive/2026-06-12/audit.md` (immutable copy)
- `ai-workflow/memory/active/audit.md` (mutable, 다음 session 부터 append)

이유: audit log 가 active/ 에 누적되면 state-doc 비대화. archive 로 periodic 이동.

### 5.3 Privacy / Redaction

- Sensitive data (PII, secrets, tokens, internal host names) 는 **stage_completion.notes 에 적지 말 것**
- 필요시 `redacted: <reason>` placeholder
- Compliance 검토 시: 별도 [PII] tag 적용

## 6. 자동화 Hook (v0.7.0+ 권장)

### 6.1 Runtime Side (이미 적용)

`stage_gate.emit_and_log()` — 호출 시 자동 audit log append (optional `audit_log_path`):
```python
emit_and_log(
    stage_name="code-generation",
    artifacts=["src/x.py"],
    next_stage="build-and-test",
    audit_log_path="ai-workflow/memory/active/audit.md",  # optional
    approval_timestamp=result.get("approval_timestamp"),
    approval_actor=result.get("approval_actor"),
)
```

### 6.2 Orchestrator Side (v0.8.0+ 후보)

Orchestrator 가 user response 수신 후 자동 호출:
- user prompt 직전: `emit_completion_message()` → 2-option prompt
- user response 수신: `append_audit_log(audit_path, response.completion)`
- 자동화 policy: `skill output 직후 user prompt → approval → audit log 자동` (stage_gate_runtime_migration.md §6)

### 6.3 Linter (v0.7.0+ 신규)

`tests/check_audit_log_compliance.py` (v0.7.0+ 후보):
- 모든 skill/MCP runtime script 가 audit log 호출 (또는 명시적 skip)
- `audit.md` 가 존재하고 ISO 8601 형식 + append-only
- `requested_changes` 가 비어있지 않은 entry 가 다음 entry 에서 resolved 표시

## 7. 마이그레이션 가이드 (v0.6.x → v0.7.0)

| Step | 작업 | Effort |
|---|---|---|
| 1 | 모든 skill 의 `run_*.py` success path 에 `emit_and_log(audit_log_path="ai-workflow/memory/active/audit.md")` 추가 (이미 v0.6.5+ batch 적용) | (이미 완료) |
| 2 | 기존 `audit.md` 가 있다면 §3.1 format 검증 (필요시 normalize script) | 1 ses |
| 3 | Orchestrator 측 자동 호출 (v0.8.0+ 후보) | 1 ses |
| 4 | Linter (audit log compliance check) | 1 ses |

## 8. 한계 / 예외

- **긴급 hotfix (P0)**: `notes: ["P0 hotfix: skip gate, reason: ..."]` mandatory. `approval_actor: "auto"` 만 허용.
- **Read-only operation**: `read_only` MCP family 는 audit log 안 함 (read-only = no state change).
- **CI/CD timeout**: `notes: ["auto-approved: CI timeout"]` + `approval_actor: "auto"` (env=ci 만).
- **Test fixture**: smoke test 의 `audit_path = Path(tempfile.gettempdir()) / "test-*.md"` — production audit log 와 격리.

## 9. AIDLC 호환

| AIDLC 정책 | 우리 정책 | 호환도 |
|---|---|---|
| `aidlc-docs/audit.md` (in-repo, append-only) | `ai-workflow/memory/active/audit.md` (state meta layer, append-only) | ✅ 100% format 일치 |
| ISO 8601 timestamp | 동일 | ✅ |
| Stage context 명시 | 동일 (`## [Stage: X]`) | ✅ |
| Raw user input 그대로 | 분리 (Q&A file / requested_changes vs summary in notes) | ⚠️ 부분 (의도적) |
| Auto-approval | 동일 정책 (CI/cron/P0) | ✅ |
| Lint / Validation | `tests/check_audit_log_compliance.py` (v0.7.0+ 후보) | ⏸ follow-up |

차이점 (의도적):
- 위치: in-repo (AIDLC) vs state meta layer (우리). 우리 3-layer 분리 (Source/Runtime/Project Docs) 정책 준수.
- Raw input: AIDLC 는 in-place, 우리 는 분리 (Q&A file).

## 10. 다음에 읽을 문서

- [`./stage_gate_pattern.md`](./stage_gate_pattern.md) §6 (분산 정의, 본 spec 으로 통합)
- [`./stage_gate_runtime_migration.md`](./stage_gate_runtime_migration.md) §5
- [`./output_schema_guide.md` §3.4](./output_schema_guide.md) — stage_completion + audit log
- [`../workflow_kit/common/contracts/stage_gate.py`](../workflow_kit/common/contracts/stage_gate.py) `append_audit_log` + `emit_and_log`
- [`../tests/check_stage_gate_compliance.py`](../tests/check_stage_gate_compliance.py) (15 test)
- AIDLC 원본: `/Users/yklee/repos/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/common/process-overview.md` + `inception/requirements-analysis.md`
