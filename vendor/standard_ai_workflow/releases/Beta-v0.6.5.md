# Beta v0.6.5 — AIDLC 패턴 차용 (Question File Format + Stage Gate)

- **릴리스 일자**: 2026-06-12
- **브랜치**: `main`
- **포함 커밋**: v0.6.3 → v0.6.5 (10 commit, ~2,600 line)
- **상태**: ✅ AIDLC (`awslabs/aidlc-workflows`) 7대 차별 메커니즘 중 2개 (Question File Format + Stage Gate Pattern) 채택 + 우리 운영 컨텍스트 적응. **35 test PASS** (7 question_format + 15 stage_gate_compliance + 13 stage_gate_runtime). breaking change 없음 (stage_completion 은 optional field, 점진 적용).

## 1. 무엇이 바뀌었나

### 1.1 Phase 회고 (v0.6.3 → v0.6.5)

| Phase | 작업 | Commit |
|---|---|---|
| v0.6.3 | (이전 baseline) | `30183c5` |
| v0.6.4 (AIDLC 분석) | 1차 출처 16 file 분석 + 1:1 비교 (강점 8 / 갭 20) + 보완안 15건 + L1 wiki topic | `2916d49`, `9946d91` |
| v0.6.4 (구현) | 신규 spec 2종 + output schema §3.4 + adoption entrypoints §7 (4 doc, 498 line) | `25756bb` |
| v0.6.4 (코드) | runtime helper 2 module + 2 smoke test (1,347 line, 22 test PASS) | `bc16d91` |
| v0.6.5 (spec) | 7 skill spec §4 출력 계약 보강 + 5 SKILL.md cross-ref + catalog §5.2 (13 file, 277 line) | `5b16517` |
| v0.6.5 (runtime) | stage_gate_runtime helper + migration guide + 13 runtime test (3 file, 644 line) | `dd98e69` |
| v0.6.5 (pilot) | automated-repro-scaffold stage_completion 통합 (1 file, 44 line) | `2fab835` |
| v0.6.5 (batch) | 6 spec 보유 skill batch stage_completion 통합 (6 file, 72 line) | `ca7a685` |

### 1.2 핵심 도입 사항

**A. Question File Format 패턴 (★★★ 권장)**
- AIDLC `common/question-format-guide.md` 차용
- multi-choice + `[Answer]:` tag + "Other" mandatory + contradiction/ambiguity auto-detection
- 8 ambiguity keyword (mix of, depends on, not sure, ...) + 6 follow-up 매핑
- runtime: `parse_answers` / `validate_answers` / `detect_ambiguity` / `detect_contradiction` / `generate_clarification_file` / `full_validation`

**C. Stage Gate 명시화 패턴 (★★★ 권장)**
- AIDLC construction phase 의 2-option completion message 차용
- **NO EMERGENT BEHAVIOR**: 3-option / 4-option ❌
- 8-field StageCompletion (stage_name/status, next_stage, requested_changes, approval_timestamp/actor, artifacts, notes)
- audit log (append-only, ISO 8601, raw user input) 통합
- auto-approval 한계 (CI/cron/P0 only, production/state/release blocked)
- runtime: 5 function (build/merge/emit_and_log/is_present/get_status)

## 2. 신규 산출물 (12 file, ~1,400 line)

### 2.1 외부 spec (4 doc)
- `workflow-source/core/question_file_format.md` (204 lines) — AIDLC 차용 spec
- `workflow-source/core/stage_gate_pattern.md` (207 lines) — gate 명시화 spec
- `workflow-source/core/stage_gate_runtime_migration.md` (166 lines) — runtime migration 가이드
- `workflow-source/core/output_schema_guide.md` §3.4 — stage_completion schema (40 line 추가)
- `workflow-source/core/workflow_adoption_entrypoints.md` §7 — v0.6.5 권장 도입 묶음 (47 line 추가)
- `workflow-source/core/workflow_skill_catalog.md` §5.2 — Stage Completion 통합 table (25+ line 추가)

### 2.2 Python module (3 file)
- `workflow_kit/common/contracts/question_format.py` (358 lines) — 6 function
- `workflow_kit/common/contracts/stage_gate.py` (335 lines) — 7 function (StageCompletion dataclass)
- `workflow_kit/common/contracts/stage_gate_runtime.py` (186 lines) — 5 function (build/merge/emit_and_log/is_present/get_status)

### 2.3 Smoke test (3 file, 35 test PASS)
- `tests/check_question_format.py` (336 lines, 7 test) — multi-choice/ambiguity/contradiction/clarification
- `tests/check_stage_gate_compliance.py` (318 lines, 15 test) — completion message/audit/2-option/auto-approval
- `tests/check_stage_gate_runtime.py` (292 lines, 13 test) — merge/emit/runtime integration

### 2.4 11종 Skill 적용 (8 file)
- **Spec 보강** (7 spec, commit 5b16517): session-start, backlog-update, doc-sync, merge-doc-reconcile, validation-plan, code-index-update, automated-repro-scaffold
- **SKILL.md cross-ref** (5 file): workflow-linter, project-status-assessment, memory-freeze, git-conflict-resolver, robust-patcher
- **Runtime 적용** (7 file, commits 2fab835 + ca7a685): 위 7 spec 의 run_*.py 성공 path 에 stage_completion merge

## 3. 호환성 / Breaking Change

**v0.6.4-5 의 breaking change 정책**:
- `stage_completion` field 는 **optional** (mandatory 아님). 기존 52 smoke test 와 즉시 호환.
- Question File Format 도 optional — 사용 안 해도 정상 운영 가능.
- 기존 status field (`"success"`, `"ok"`, `"warning"`, `"error"`) 모두 보존. runtime helper 가 `"success"` → `"ok"` 매핑.

**점진 적용 전략**:
- v0.6.5 에서 7 spec + 5 SKILL.md = 12/12 spec/cross-ref 적용 완료
- v0.6.5 runtime 은 7/7 spec 보유 skill 적용 완료
- v0.7.0+ 에서 모두 적용 후 `stage_completion` 을 required 로 격상 (선택)

## 4. AIDLC 호환

v0.6.4-5 의 우리 패턴 ↔ AIDLC construction phase 매핑:

| AIDLC 패턴 | 우리 구현 | 호환도 |
|---|---|---|
| 3-Phase Lifecycle (Inception/Construction/Operations) | mode 6종 (horizontal) | ❌ 의도적 차이 (우리 mode 시스템 유지) |
| Adaptive Stage Execution | skill EXECUTE/SKIP 미구현 (v0.6.5+ 후보) | ⚠️ 부분 (runtime 적용으로 1차) |
| Workspace Detection | state.json 으로 유사 | ⚠️ 부분 |
| Reverse Engineering 9-Artifact | repository_assessment 1개 | ❌ v0.7.0+ 후보 |
| Per-Unit Loop | (해당 없음) | ❌ 우리 스케일 부적합 |
| **Extensions 시스템** (opt-in + blocking rule) | ❌ v0.7.0+ 후보 | ❌ (follow-up) |
| **Question File Format** | ✅ v0.6.4 | ✅ 100% |
| **Append-only Audit Log** | ✅ stage_gate.append_audit_log | ✅ 100% |

## 5. Follow-up (v0.6.6+ 후보)

| 항목 | 우선순위 | Effort |
|---|---|---|
| 11종 skill spec 의 5 SKILL.md-only skill 의 runtime script 작성 (선택) | 낮음 | 2-3 ses |
| 11종 skill runtime output schema validator (v0.7.0 required 격상) | 중간 | 1-2 ses |
| Orchestrator 측 자동 emit_and_log 통합 (skill output 직후 user prompt → audit log 자동) | 중간 | 2-3 ses |
| AIDLC 의 Extensions 시스템 차용 (security-baseline 1종) | 높음 | 3-5 ses |
| AIDLC 의 3-Phase Lifecycle 부분 도입 (mode 위에 phase overlay) | 낮음 | 1 ses |
| Reverse Engineering 9-Artifact 자동 생성 | 중간 | 2-3 ses |
| ADR-NNN: Operations phase 도입 여부 | 낮음 | 1 ses |
| v0.6.4 + v0.6.5 release artifact (dist/) export + harness bundle 갱신 | 중간 | 1-2 ses |

## 6. 누적 변경 요약 (v0.5.11 → v0.6.5)

| 영역 | v0.5.11 → v0.6.5 누적 |
|---|---|
| 외부 spec | +1,170 line (4 doc 신규 + 3 doc 보강) |
| Code | +879 line (3 module 신규) |
| Smoke test | +946 line (3 file, 35 test PASS) |
| 11종 skill spec | +280 line (7 spec §4.1 stage_completion) |
| 11종 skill runtime | +116 line (7 spec batch 적용, 12-72 line each) |
| L1 wiki | +735 line (3 concept + 1 topic + log entry) |
| **총** | **~4,100 line** (35 test PASS) |

## 7. References

- AIDLC 원본: `https://github.com/awslabs/aidlc-workflows` (commit `b19c819`, 2026-06-08)
- AIDLC 분석 노트: `~/repos/standard_ai_workflow_minimax/ai-workflow/wiki/topics/aidlc-benchmark-analysis-2026-06-12.md` (L1 wiki)
- v0.6.4 (이전 baseline): `Beta-v0.6.3.md`
- Phase 11 status: in_progress
EOF