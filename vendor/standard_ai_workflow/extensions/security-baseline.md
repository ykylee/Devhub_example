# Security Baseline Extension (v0.7.0, AIDLC 차용)

- 문서 목적: standard_ai_workflow 의 cross-cutting security rule baseline. AIDLC `extensions/security/baseline/` 차용. v0.7.0 의 Extension 시스템 (B) 1차 출시 — security-baseline 1종.
- 범위: 6 rule (SEC-WF-01 ~ SEC-WF-06), opt-in pattern, audit log 통합, workflow_kit helper
- 대상 독자: workflow 설계자, AI agent, 운영자, compliance 검토자
- 상태: stable (v0.7.0 도입)
- 최종 수정일: 2026-06-12
- 관련 문서: [`../core/stage_gate_pattern.md`](../core/stage_gate_pattern.md) (gate 정책), [`../core/audit_log_standard.md`](../core/audit_log_standard.md) (audit log 표준), [`../workflow_kit/common/contracts/stage_gate.py`](../workflow_kit/common/contracts/stage_gate.py) `require_explicit_approval`, [`../workflow_kit/common/contracts/question_format.py`](../workflow_kit/common/contracts/question_format.py) `validate_answers`
- 1차 출처: AIDLC `awslabs/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/extensions/security/baseline/security-baseline.md` (307 line, 2026-06-08 commit `b19c819`)

## 1. 왜 Security Baseline 이 필요한가

v0.6.4-5 에서 stage_completion + audit log + question format 등 runtime 안전장치 도입. 그러나 **cross-cutting security rule** (encryption, dependency integrity, error handling, fail-closed 등) 의 SSOT 가 없었음.

AIDLC 의 extensions 시스템:
- baseline `<name>.md` (rules) + `<name>.opt-in.md` (사용자 선택 prompt)
- 워크플로우 시작 시 opt-in prompt → 사용자가 enable/disable 결정
- enabled rule 은 **hard constraint** (stage completion 에서 위반 시 "Request Changes" 만 표시, "Continue" 옵션 사라짐)

본 1차 출시: **security-baseline 1종**, 우리 운영 컨텍스트에 applicable 한 6 rule. opt-in pattern 적용.

## 2. Extension 시스템 (AIDLC 패턴)

### 2.1 baseline + opt-in file 페어

| File | 역할 |
|---|---|
| `extensions/security-baseline.md` (본 doc) | 6 rule 의 hard constraint 정의 |
| `extensions/security-baseline.opt-in.md` (별도) | 사용자 opt-in prompt (workflow 시작 시 1회 표시) |

### 2.2 opt-in prompt 형식

사용자 prompt:
> Question: Should security baseline extension rules be enforced for this project?
>
> A) Yes — enforce all SEC-WF rules as blocking constraints (recommended for production-grade applications)
>
> B) No — skip all SEC-WF rules (suitable for PoCs, prototypes, experimental projects)
>
> X) Other (please describe after [Answer]: tag below)
>
> [Answer]:

답변: `ai-workflow/memory/active/security_baseline_status.md` (또는 project state) 에 기록. `aidlc-state.md` 의 `## Extension Configuration` 표 형태.

### 2.3 Hard Constraint 정책 (enabled 시)

- stage completion message 에 "Security Compliance" 섹션 추가
- 각 rule 마다 compliant / non-compliant / N/A 표시
- 1+ rule 이 non-compliant 이면 "Continue" 옵션 사라짐, "Request Changes" 만 표시
- 0% compliance 위반 시 (모두 N/A 또는 모두 compliant) 정상 진행

## 3. Rule 정의 (6 rule)

### 3.1 Rule SEC-WF-01: Audit Log Append-Only + ISO 8601

**Rule**: `ai-workflow/memory/active/audit.md` 는 append-only, ISO 8601 timestamp, raw user input 분리 정책 준수.

**Verification**:
- `audit.md` 첫 줄이 `## [Stage: X] [ISO_8601]` 형식
- `audit.md` 가 *overwrite* 된 흔적 없음 (read-modify-write 만)
- microsecond leak 없음
- 자동 검증: `tests/check_audit_log_compliance.py` (13 test, commit 54e96a9)

**Cross-ref**: [`../core/audit_log_standard.md`](../core/audit_log_standard.md), [`../tests/check_audit_log_compliance.py`](../tests/check_audit_log_compliance.py)

### 3.2 Rule SEC-WF-02: Stage Gate Approval Mandatory

**Rule**: 모든 skill/MCP output 의 `stage_completion.approval_actor` 는 `user`, `orchestrator`, `auto` 만 허용. auto-approval 은 CI/cron/P0_hotfix env 에서만 허용. production 코드 / state 문서 / release 변경은 user approval mandatory.

**Verification**:
- `require_explicit_approval()` 의 `is_production_code` / `is_state_doc` / `is_release` 가 True 일 때 auto-approval 거부
- `notes` 에 auto-approval 사유 keyword 포함 (`p0 hotfix` / `ci timeout` / `cron` / `auto-approved`)
- 자동 검증: `tests/check_stage_gate_compliance.py` 의 `test_auto_approval_*` (15 test 중 5)

**Cross-ref**: [`../core/stage_gate_pattern.md` §4](../core/stage_gate_pattern.md), [`../workflow_kit/common/contracts/stage_gate.py` `require_explicit_approval`](../workflow_kit/common/contracts/stage_gate.py)

### 3.3 Rule SEC-WF-03: Question Format Validation (Q&A 모호성 검출)

**Rule**: 모든 사용자 결정 (multi-choice question) 은 [Question File Format](./question_file_format.md) 준수. 모호 응답 자동 검출.

**Verification**:
- 모든 [Answer]: tag 가 valid letter (A-Z)
- 모호 keyword (`mix of`, `depends on`, `not sure`, `tbd`, ...) 자동 검출
- Contradiction (cross-question 모순) 검출
- Clarification file 자동 emit
- 자동 검증: `tests/check_question_format.py` (7 test, commit bc16d91)

**Cross-ref**: [`../core/question_file_format.md`](../core/question_file_format.md), [`../workflow_kit/common/contracts/question_format.py`](../workflow_kit/common/contracts/question_format.py)

### 3.4 Rule SEC-WF-04: Error Handling Fail-Closed

**Rule**: 모든 skill 의 error path 는 **fail-closed** (다음 stage 진행 ❌). user explicit approval 후에만 진행. stage_completion.stage_status = "error" 시 next_stage = None.

**Verification**:
- 모든 skill 의 `run_*.py` error path 가 `result = build_error_result(...)` + `print(json.dumps(...))` + `return 1` 패턴
- error path 의 stage_completion.stage_status = "error", next_stage = None
- `ensure_stage_completion()` fallback 이 error path 에서도 적용
- 자동 검증: `tests/check_stage_completion_required.py` `test_ensure_status_mapping` (8 test 중 1)

**Cross-ref**: [`../core/output_schema_guide.md` §3.4](../core/output_schema_guide.md), [`../workflow_kit/common/contracts/stage_gate_runtime.py` `ensure_stage_completion`](../workflow_kit/common/contracts/stage_gate_runtime.py)

### 3.5 Rule SEC-WF-05: Dependency Integrity (Lock File + Checksum)

**Rule**: 모든 Python dependency 는 `requirements.txt` / `requirements-dev.txt` / `pyproject.toml` 의 lock file 로 관리. wheel install 시 SHA-256 checksum 검증 (optional).

**Verification**:
- `requirements*.txt` 가 commit tracked
- `pip install --require-hashes` 옵션 (production 환경) 권장
- `pyproject.toml` 에 dependency version pinned (e.g., `>=0.6.0,<0.7.0`)
- 자동 검증: 본 commit 범위 외 (별도 session, v0.7.1+)

**Cross-ref**: [`../../workflow_kit/pyproject.toml`](../../workflow_kit/pyproject.toml), [`../../../requirements.txt`](../../../requirements.txt), [`../../../requirements-dev.txt`](../../../requirements-dev.txt)

### 3.6 Rule SEC-WF-06: Documentation Compliance (R-9 Skip Marker)

**Rule**: wiki in-repo (L1) 의 모든 page 는 R-9 (raw source = `archive/` only) 준수. R-9 rule 정의/설명 page 는 `r9_skip: true` frontmatter marker 적용 (V-R9 lint skip).

**Verification**:
- L1 wiki page 의 frontmatter `r9_skip: true` (또는 `1` / `yes`) 면 V-R9 skip
- V-R9 naive grep false-positive 0건
- 자동 검증: `tests/check_wiki_source_rule.py` (v0.6.1.5+ 기존, 50 line)

**Cross-ref**: [`../concepts/wiki-source-rule-r9.md`](../concepts/wiki-source-rule-r9.md), [`../tests/check_wiki_source_rule.py`](../tests/check_wiki_source_rule.py)

## 4. Compliance Summary (Stage Completion 시)

모든 stage 의 completion message 에 다음 섹션 추가 (enabled 시):

```markdown
## Security Compliance (security-baseline)

| Rule | Status | Notes |
|---|---|---|
| SEC-WF-01 (Audit Log) | ✅ compliant | audit.md format 검증 PASS |
| SEC-WF-02 (Stage Gate) | ✅ compliant | user approval recorded |
| SEC-WF-03 (Question Format) | ⚠️ N/A | 이 stage 에 Q&A 없음 |
| SEC-WF-04 (Error Handling) | ✅ compliant | error path fail-closed |
| SEC-WF-05 (Dependency) | ⏸ deferred | v0.7.1+ 검증 |
| SEC-WF-06 (R-9 Skip Marker) | ✅ compliant | V-R9 0 violation |

**Summary**: 4 compliant, 1 N/A, 1 deferred. **Hard constraint: PASS** (모든 enabled rule 의 verification 충족 또는 N/A).
```

### 4.1 Hard Constraint 정책 (재강조)

- 1+ rule 의 Status = ❌ non-compliant 이면 **gate 정지**:
  - "Continue to Next Stage" 옵션 제거
  - "Request Changes" 옵션만 표시
  - audit log 에 blocking finding 기록
- 1+ rule 이 ⏸ deferred 이면 **gate 통과** (v0.7.1+ 후속 검증)
- 모두 N/A 또는 ✅ compliant 이면 **gate 통과**

## 5. 우리 운영 적용

### 5.1 opt-in prompt 위치

`workflow_kit/skills/session-start/scripts/run_session_start.py` (v0.7.1+ follow-up):
- 새 session 시작 시 1회 prompt 표시
- 응답을 `ai-workflow/memory/active/security_baseline_status.md` 에 저장
- 다음 session 부터 status read → enabled/disabled 적용

### 5.2 workflow_kit helper (v0.7.0 신규)

- `workflow_kit.common.contracts.security_baseline.evaluate_compliance(stage_name, artifacts)` — stage 의 6 rule compliance 자동 평가
- `check_security_baseline.py` smoke test (v0.7.0 신규, 본 commit)

### 5.3 Linter 통합 (v0.7.1+ 후보)

- `check_security_baseline.py` 가 `tests/check_*.py` 사이클에 자동 포함
- 0% non-compliant rule 검증 (모든 stage)
- V-R1 / V-4 / V-R9 와 정합 (wiki lint)

## 6. 한계 / 예외

- **PoC / Prototype**: opt-in prompt 에서 "B) No" 선택 시 rule 미적용
- **N/A 처리**: 특정 stage 에 rule 이 적용 불가 시 (e.g., SEC-WF-03 question format 는 Q&A 없는 stage 에서 N/A) — non-compliant 가 아님
- **Deferred**: v0.7.0 시점 미구현 검증 로직 (e.g., SEC-WF-05 dependency integrity) — blocking 아님, follow-up 추적
- **Stage 별 다른 적용**: 일부 stage (e.g., session-start) 는 SEC-WF-04 만 relevant. compliance summary 가 stage 별 동적 적용.

## 7. AIDLC 호환

| AIDLC 정책 | 우리 정책 | 호환도 |
|---|---|---|
| 15 SECURITY rule (SECURITY-01 ~ 15) | 6 SEC-WF rule (workflow domain specific) | ⚠️ 부분 (도메인 적응) |
| `<name>.md` + `<name>.opt-in.md` 페어 | 동일 | ✅ |
| opt-in prompt 형식 | 동일 | ✅ |
| Hard constraint (blocking finding) | 동일 | ✅ |
| N/A 처리 | 동일 | ✅ |
| Audit log 통합 | 동일 (audit_log_standard.md 활용) | ✅ |

차이점 (의도적):
- 15 rule → 6 rule (workflow 도메인 specific; AWS-특화 rule 은 제외)
- AIDLC 의 SECURITY-04 (HTTP headers) 는 우리 workflow 가 HTTP 서비스 아니므로 N/A 처리 (default: skip)

## 8. Follow-up

- v0.7.1: workflow_kit.common.contracts.security_baseline.evaluate_compliance 구현
- v0.7.1: session-start 에 opt-in prompt 통합
- v0.7.1: SEC-WF-05 dependency integrity 검증 (lock file + checksum)
- v0.7.1: SEC-WF-06 R-9 skip marker 자동 검증 (run test 시)
- v0.8.0: Extension 시스템 2종 (testing-baseline, performance-baseline) 추가

## 9. 다음에 읽을 문서

- [`../core/stage_gate_pattern.md`](../core/stage_gate_pattern.md) — gate 정책
- [`../core/audit_log_standard.md`](../core/audit_log_standard.md) — audit log 표준
- [`../core/question_file_format.md`](../core/question_file_format.md) — question format
- [`../workflow_kit/common/contracts/stage_gate.py`](../workflow_kit/common/contracts/stage_gate.py) — stage_completion
- [`../workflow_kit/common/contracts/question_format.py`](../workflow_kit/common/contracts/question_format.py) — question format
- AIDLC 원본: `/Users/yklee/repos/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/extensions/security/baseline/security-baseline.md` (307 line)
