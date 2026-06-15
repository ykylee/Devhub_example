# Beta v0.7.0 — AIDLC Extension 시스템 + 9-Artifact (6 step 완료)

- **릴리스 일자**: 2026-06-13
- **브랜치**: `main`
- **포함 커밋**: v0.6.5 → v0.7.0 (15 commit, ~3,200 line)
- **상태**: ✅ AIDLC (`awslabs/aidlc-workflows`) 7대 차별 메커니즘 중 3개 (Extension 시스템 / Reverse Engineering 9-Artifact / 3-Phase Lifecycle) 채택. **130 test PASS** (15 security_baseline + 17 unit_of_work_template + 13 audit_log_compliance + 8 stage_completion_required + 13 stage_gate_runtime + 15 stage_gate_compliance + 7 question_format + 19 reverse_engineering + 23 extension_system). breaking change 없음 (v0.7.0 부터 stage_completion required, ensure_stage_completion() lazy fallback).

## 1. 무엇이 바뀌었나

### 1.1 Phase 회고 (v0.6.5 → v0.7.0)

| Step | 작업 | Commit | Line | Test |
|---|---|---|---|---|
| Step 1 | stage_completion required 격상 + ensure_stage_completion() helper | `6e57cf3` | +319 | 8 PASS |
| Step 6 | Reverse Engineering 9-Artifact (9 md template + 13 step guide) | `4bbd391` | +925 | 19 PASS |
| Step 7 | Extension 시스템 일반화 (SCHEMA.md + 3종 baseline) | `0052da1` | +1150 | 23 PASS |
| Step 8 | Security-baseline 1종 (6 rule + opt-in) | `dc2c22b` | +558 | 15 PASS |
| Step 9 | Unit of Work 3-layer template (system/task/atom) | `c981cac` | +622 | 17 PASS |
| Step 10 | Audit Log 표준화 (per-project audit.md + 2 latent bug fix) | `54e96a9` | +637 | 13 PASS |

### 1.2 핵심 도입 사항

**A. Stage Completion Required 격상 (Step 1)**
- v0.6.5 의 optional field → v0.7.0 부터 required
- `ensure_stage_completion()` lazy fallback (CI/cron/P0 호환)
- 8 test PASS (StageCompletion 8-field + lazy fallback + policy)

**B. Audit Log 표준화 (Step 10)**
- AIDLC 의 per-project `audit.md` (append-only, ISO 8601, raw input) 차용
- 2 latent bug fix: `append_audit_log()` 의 leading newline + `normalize_timestamp()` microsecond leak
- 13 test PASS (audit log 6 field + 2 bug fix + 6 helper + integration)

**C. Unit of Work 3-layer Template (Step 9)**
- AIDLC `inception/units-generation.md` 차용
- system-level 분해 + dependency matrix + Mermaid graph + story mapping + code organization
- 17 test PASS (UOW 5 field + dep matrix 5 + mermaid 2 + story mapping 1 + template 3 + 통합 1)
- cycle detection (DFS) + DAG validation 포함

**D. Security-baseline 1종 (Step 8)**
- AIDLC `extensions/security/baseline/` 차용, 15 rule → 우리 6 rule (workflow domain)
- 6 rule SEC-WF-01~06: Audit Log / Stage Gate Approval / Question Format / Error Handling / Dependency Integrity / R-9 Skip Marker
- opt-in pattern (A/B/X → A/B/P/X)
- 15 test PASS

**E. Reverse Engineering 9-Artifact (Step 6)**
- AIDLC `inception/reverse-engineering.md` 311 line 차용
- 9 md file (business-overview/architecture/code-structure/api/component-inventory/tech-stack/dependencies/code-quality/metadata)
- 13 step 가이드 (multi-package discovery → 9 artifact → timestamp → state sync → user message → approval)
- 19 test PASS

**F. Extension 시스템 일반화 (Step 7)**
- AIDLC 3종 extension (security/testing/resiliency) → 우리 3종 (security/testing/performance)
- `extensions/SCHEMA.md` (200 line, 11 section) — extension system SSOT
- testing-baseline: AIDLC PBT 9 rule → 우리 6 rule (PBT-05/06/08 N/A)
- performance-baseline: 우리 domain 적응 (workflow runtime, AIDLC 없음)
- security opt-in update (SCHEMA 형식 정합)
- 23 test PASS

### 1.3 우리 사용 패턴 적응 결정

| AIDLC 패턴 | 우리 적응 | 비율 |
|---|---|---|
| 15 SECURITY rule | 6 SEC-WF rule | 40% |
| 9 PBT rule | 6 TST-WF rule (PBT-05/06/08 N/A) | 67% |
| 16 RESILIENCY rule | 6 PERF-WF rule (HA/DR/Incident N/A) | 38% |
| Reverse Engineering 9 artifact | 9 artifact (구조 유지, 내용 압축) | 100% |
| Unit of Work 4종 (FD/NFR/NFRD/ID) | 3 layer (system/task/atom) | 75% |

## 2. 신규 산출물 (~3,200 line)

### 2.1 외부 spec (8 doc)
- `workflow-source/core/reverse_engineering.md` (148 lines) — 9-Artifact 13 step 가이드
- `workflow-source/extensions/SCHEMA.md` (200 lines) — Extension system SSOT
- `workflow-source/extensions/security-baseline.md` (242 lines) — 6 SEC-WF rule
- `workflow-source/extensions/security-baseline.opt-in.md` (114 lines) — opt-in prompt (Step 7 update)
- `workflow-source/extensions/testing-baseline.md` (130 lines) — 6 TST-WF rule
- `workflow-source/extensions/testing-baseline.opt-in.md` (103 lines) — opt-in prompt
- `workflow-source/extensions/performance-baseline.md` (110 lines) — 6 PERF-WF rule
- `workflow-source/extensions/performance-baseline.opt-in.md` (99 lines) — opt-in prompt
- `workflow-source/templates/unit_of_work_template.md` (208 lines) — UOW 3-layer template
- `workflow-source/core/audit_log_standard.md` (200 lines) — audit log 표준

### 2.2 Reverse Engineering 9-Artifact (9 md)
- `workflow-source/reverse-engineering/01-business-overview.md` (33 lines)
- `workflow-source/reverse-engineering/02-architecture.md` (49 lines)
- `workflow-source/reverse-engineering/03-code-structure.md` (50 lines)
- `workflow-source/reverse-engineering/04-api-documentation.md` (48 lines)
- `workflow-source/reverse-engineering/05-component-inventory.md` (55 lines)
- `workflow-source/reverse-engineering/06-technology-stack.md` (51 lines)
- `workflow-source/reverse-engineering/07-dependencies.md` (44 lines)
- `workflow-source/reverse-engineering/08-code-quality-assessment.md` (54 lines)
- `workflow-source/reverse-engineering/09-reverse-engineering-metadata.md` (54 lines)

### 2.3 Python module (3 file)
- `workflow_kit/common/contracts/stage_gate_runtime.py` (162 lines) — ensure_stage_completion() lazy fallback
- `workflow_kit/common/contracts/stage_gate.py` (update) — 2 latent bug fix (leading newline + microsecond)
- `workflow_kit/common/contracts/audit_log.py` (NEW) — audit log standard (v0.7.0 Step 10)

### 2.4 Smoke test (5 file, 95 test PASS 신규)
- `tests/check_stage_completion_required.py` (8 test) — Step 1
- `tests/check_audit_log_compliance.py` (13 test) — Step 10
- `tests/check_unit_of_work_template.py` (17 test) — Step 9
- `tests/check_security_baseline.py` (15 test) — Step 8
- `tests/check_reverse_engineering.py` (19 test) — Step 6
- `tests/check_extension_system.py` (23 test) — Step 7

## 3. 누적 검증

### 3.1 Test 통과율

| Test set | v0.6.5 | v0.7.0 |
|---|---|---|
| Total | 35 | 130 (+95) |
| 신규 | — | 95 (Step 1, 6, 7, 8, 9, 10) |
| 회귀 | 0 | 0 (모든 누적 test PASS) |

### 3.2 Coverage

- workflow_kit.common.contracts: 4 module (stage_gate / stage_gate_runtime / audit_log / question_format)
- templates: 1 (unit_of_work_template) + 기존
- extensions: 3 baseline (security + testing + performance) + SCHEMA + 3 opt-in
- reverse-engineering: 9 artifact + 1 step guide
- core spec: 6 (global_workflow_standard + stage_gate_pattern + question_file_format + audit_log_standard + reverse_engineering + stage_gate_runtime_migration)

## 4. Follow-up (v0.7.1+)

1. `workflow_kit.common.contracts.{security,testing,performance}_baseline.evaluate_compliance()` 구현 (3종)
2. session-start 에 3종 opt-in prompt 통합
3. `state.json` 의 `<name>_baseline` 필드 schema validation
4. PERF-WF-03 (memory 자동 측정) + PERF-WF-06 (profiling decorator) 구현
5. v0.8.0: Extension sub-cat 도입 (e.g. `extensions/security/auth/`)
6. v0.8.0: 4종 (resiliency) 추가 — workflow_kit health check + 장애 대응
7. v0.8.0: UOW 기반 sub-agent 위임 자동화
8. v0.7.0 packaging (5 harness) + GitHub release v0.7.0

## 5. References

- 1차 출처: AWS AIDLC `awslabs/aidlc-workflows` (commit `b19c819`, 2026-06-08)
  - `aidlc-rules/aws-aidlc-rule-details/inception/reverse-engineering.md` (311 line, 9-Artifact)
  - `aidlc-rules/aws-aidlc-rule-details/inception/units-generation.md` (188 line, UOW 4종)
  - `aidlc-rules/aws-aidlc-rule-details/extensions/security/baseline/` (307 line, 15 rule)
  - `aidlc-rules/aws-aidlc-rule-details/extensions/testing/property-based/` (284 line, 9 rule)
  - `aidlc-rules/aws-aidlc-rule-details/extensions/resiliency/baseline/` (490 line, 16 rule)
  - `aidlc-rules/aws-aidlc-rule-details/construction/build-and-test.md` (audit log)
- 우리 wiki: `~/wiki/wiki/projects/standard-ai-workflow/sources/topics-aidlc-benchmark-analysis-2026-06-12.md` (D/B/A 보완안 4종)
- 우리 log: `ai-workflow/wiki/log.md` (6 step entry + 본 release entry)

---

## v0.7.0 follow-up (2026-06-13)

본 release 직후 follow-up 3 task 완료 (commit 71de1b0):

### Task 3: v0.7.0 packaging (commit 390a6e0)

- pyproject.toml + workflow_kit/__init__.py: 0.6.3 → 0.7.0
- wheel + sdist 빌드 (159898 bytes) + twine check PASSED
- fresh venv smoke 정상
- **GitHub Release v0.7.0-beta 발행** (2 asset: wheel + tar.gz)

### Task 2: session-start opt-in 통합 (spec level)

- `core/session_start_skill_spec.md` §11 추가 (76 line)
- 3종 baseline opt-in 흐름 6 step (detect → prompt → response → file → state.json → audit)
- Default 정책: Greenfield (A/A/P) / Brownfield (기존 존중) / CI/CD (skip)
- Session-Start 출력 schema: `extension_baselines` + `warnings` field 추가
- Runtime 통합: §11.5 v0.7.1+ follow-up

### Task 1: evaluate_compliance() helper

- `workflow_kit/common/contracts/baselines.py` 신규 (516 line)
- 3종 baseline × 6 rule = 18 RuleResult runtime 평가
- 12 test PASS (smoke test `check_baselines_compliance.py`)
- 누적 **142 test PASS** (v0.7.0 130 + 12 신규)
