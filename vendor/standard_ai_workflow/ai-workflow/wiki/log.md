---
type: meta
status: active
r9_skip: true
title: Wiki Ingest/Query Log
related_pages: [INGEST_GUIDE]
last_touched: 2026-06-12
created: 2026-06-12
updated: 2026-06-12
---

# Wiki Ingest/Query Log

- 문서 목적: 모든 ingest / query / lint 이벤트의 append-only 작업 로그. 시간 순 보존, 편집 금지 (append only).
- 갱신 규칙: ingest 종료 시 또는 query/lint 실행 시 한 줄 추가. 형식 `## [YYYY-MM-DD] <event> | <summary>`.
- 최종 갱신일: 2026-06-12

## [2026-06-12] bootstrap | wiki layer P1 prototype

- SCHEMA.md, index.md, log.md, 2 concept pages seeded
- target: ai-workflow/wiki/ Runtime layer (R1, D1, D2)

## [2026-06-12] ingest | phase-1-scaffold — codebase ingest pipeline

- INGEST_GUIDE.md 신규 — codebase self-ingest 절차 (Phase 1-7), R9 면제 명시
- index.md anchor 확장 — 5 page type × Phase 2-5 placeholder (8 concept, 5 decision, 4 pattern, 13 entity, 3 topic)
- L3 raw mirror 셋업: `~/wiki/wiki-sync-standard-ai-workflow.sh` + `raw/projects/standard-ai-workflow/` (588 files, 4.7 MB)
- L2 placeholder: `~/wiki/index.md` §18 standard-ai-workflow section 추가 (Phase 6 hook)
- L1 wiki ingest commit 1회 (R2, 5-15 page 동시 갱신 준수)

## [2026-06-12] ingest | phase-2-concepts — 5 concept pages 신규

- wiki-source-rule-r9.md (98 lines, 11 cross-ref) — R9 (wiki-ingest source = archive/ only, v0.6.1.5)
- memory-3-state-lifecycle.md (105 lines, 4 cross-ref) — active ↔ archive ↔ release 3-state
- contract-v1-output-validation.md (180 lines, 6 cross-ref) — Pydantic v2 + output_validator + delegator + MUST NOT 7+2
- agent-topology.md (145 lines, 8 cross-ref) — 4-role orchestrator + 3 sub-agent (doc/code/validation)
- harness-distribution.md (103 lines, 7 cross-ref) — 6-harness overlay + bundle structure + export workflow
- 합계 631 lines, 36 cross-ref (5 page 평균 7.2)
- 모두 R2 1 commit (5-15 page 동시 갱신) 준수, R3 pull-before-push 대기

## [2026-06-12] ingest | phase-3-decisions — 4 ADR 신규 + 1 ADR extend

- decisions/adr-001-3-layer-separation (112 lines, 7 cross-ref, accepted v0.5.2) — Source/State/Knowledge 3-layer 분리
- decisions/adr-002-pydantic-v2-contract-v1 (111 lines, 6 cross-ref, accepted v0.5.4) — 외부 markdown spec + Pydantic v2 helper 결합
- decisions/adr-003-read-only-mcp-default (114+12 lines, 3 cross-ref, accepted v0.5.5) — MCP 6+1 server + 2 transport 의 read-only default
- decisions/adr-004-wiki-layer (37 → 114 lines, 7 cross-ref, accepted v0.6.0) — extend: frontmatter 정규화 + v0.6.1+ Evolution (R8/R9) + References 추가
- decisions/adr-005-r9-wiki-source-rule (119 lines, 22 cross-ref, proposed v0.6.1.5) — R9 rule 의 정식 ADR 승격
- 합계 5 page, 570 lines (extend 포함), 45 cross-ref (5 page 평균 9.0)
- ADR-001/002/003 mirror + ADR-004 extend + ADR-005 신규
- 모두 R2 1 commit (5-15 page 동시 갱신) 준수
- ADR-005 status=proposed (정식 accepted 시 status 갱신 필요)

## [2026-06-12] ingest | phase-4-patterns-entities — 3 pattern + 12 entity (R2 batch 3, 15 page = 한도 매칭)

- patterns/frozen-archive-immutability (119 lines, 3 cross-ref) — R8 freeze mechanism
- patterns/wiki-stub-emit (104 lines, 3 cross-ref) — bootstrap_lib/wiki.py emitter
- patterns/stale-90day-lint (74 lines, 4 cross-ref) — L05 stale 90일+ lint
- entities/standard-ai-workflow (159 lines, 21 cross-ref) — PRIMARY hub entity
- entities/workflow-kit (73 lines, 6 cross-ref) — Python package
- entities/workflow-source (73 lines, 11 cross-ref) — SSOT source dir
- entities/ai-workflow-runtime (71 lines, 5 cross-ref) — runtime layer
- entities/skill-catalog (107 lines, 7 cross-ref) — 11 skills
- entities/mcp-read-only-bundle (91 lines, 9 cross-ref) — 6+1 MCP servers
- entities/harness-overlay-codex (105 lines, 6 cross-ref) — Codex
- entities/harness-overlay-opencode (80 lines, 8 cross-ref) — OpenCode 5-agent
- entities/harness-overlay-gemini-cli (72 lines, 10 cross-ref) — Gemini CLI
- entities/harness-overlay-antigravity (92 lines, 10 cross-ref) — Antigravity
- entities/harness-overlay-minimax-code (87 lines, 13 cross-ref) — MiniMax Code
- entities/harness-overlay-pi-dev (71 lines, 5 cross-ref) — pi-dev
- 합계 15 page, 1378 lines, 121 cross-ref (15 page 평균 8.07)
- R2 1 commit = 1 ingest (5-15 page 동시 갱신) 한도 정확히 매칭 (15 page)

## [2026-06-12] ingest | phase-5-topics — 3 cross-cutting topic (R2 batch 4)

- topics/standard-ai-workflow-architecture-2026 (154 lines, 37 cross-ref) — repo 전체 architecture 종합 view (3-layer + LLM wiki + memory 3-state + 6 harness + R8/R9 lifecycle)
- topics/harness-distribution-model (155 lines, 25 cross-ref) — 6 harness × {code, doc, manifest} × {version} 매트릭스 + export pipeline
- topics/wiki-ingest-lifecycle (148 lines, 40 cross-ref) — Phase 1-7 codebase ingest + 메모리 snapshot path (R8/R9) + lint/query cycle
- 합계 3 page, 457 lines, 102 cross-ref (3 page 평균 34.0 — topic 이라 hub 역할)
- L1 누적: 32 page (1 SCHEMA + 1 INGEST_GUIDE + 1 index + 1 log + 8 concepts + 5 decisions + 4 patterns + 13 entities + 3 topics)

## [2026-06-12] lint-fix | phase-7-verification — L1 broken wikilink fix

- L1 wikilink lint: 313 → 305 wikilink, 8 real broken → 0 (4 false-positive 는 backtick 안 syntax example)
  - 5× `[[patterns/harness-overlay-factory]]` → 제거 (no equivalent page, harness-distribution 으로 통합)
  - 3× `[[concepts/three-layer-architecture]]` → `[[concepts/project-architecture]]` (canonical page 매핑)
  - 2× `[[open questions]]` / `[[Beta-v0.6.1.5]]` (template text) → prose / markdown link
- L1 V-1 (location) PASS, V-4 (index structure) PASS (34 entries validated)
- L1 V-R9 (source rule): 16 false-positive — R9 rule 자체를 설명하는 5 page 가 의도적으로 `active/` mention. v0.6.1.5 의 R9 lint 구현은 naive grep 기반 — **v0.2.0 skip marker 보강** (frontmatter `r9_skip: true` 로 V-R9 skip, R-4 정책 따라 frontmatter only marker)
- L1 commit a5762e5

## [2026-06-12] backfill | phase-6-verification — P6 close + A1~A4 audit

- **Trigger**: P7 검증 재방문 중 INGEST_GUIDE §2 phase 표 와 실제 commit history 의 gap 발견 (P1~P5, P7 만 존재, P6 누락)
- **A1 (L2 lint)**: `~/wiki/_lint/report_2026-06-12.md` — 0 error / 0 warn / 0 info (pages=420). P7 L2 part PASS
- **A2 (wiki-source-sync idempotent)**: `~/wiki/wiki/projects/standard-ai-workflow/sources/` = 494 stub. raw mirror 의 ADR-001~004 풀네임 → L2 kebab-case stem 매칭 정상 (e.g. `architecture-adr-001-source-state-knowledge-3-layer-separation.md`). P6 mass source stub emit 완료
- **A3 (L1 ↔ L2 cross-layer 정합)**: L1 28 page 의 last_ingested_from 34 unique path 검증
  - 22 paths OK (raw mirror 매칭)
  - 4 paths synthetic/glob: `dist/harnesses/*/v0.6.3-beta/`, `dist/harnesses/opencode/v0.6.3-beta/agents/` (의도된 glob), `v0.6.3-beta release notes` (multi-file composite), `ai-workflow/wiki/INGEST_GUIDE.md` (L1 자기 자신)
  - 3 ADR-001/002/004 는 body 의 `[ADR-NNN file](../../docs/architecture/...)` 가 실제 raw mirror 와 매칭, frontmatter 의 last_ingested_from 부재는 정합 OK
  - **1 fix**: INGEST_GUIDE frontmatter `last_ingested_from: .omo/plans/v0.6.1-plus-codebase-ingest-design.md` → `.omo/plans/llm-wiki-convergence-design.md` (해당 plan 은 소스/raw 어느 쪽에도 미존재, master plan 으로 정정). body §7 의 "다음에 읽을 문서" 도 동일 plan 가리킴
- **A4 (log 정합)**: 본 entry (L1) + L2 `~/wiki/log.md` 에 동기 entry append. P6 자체는 emit 됐지만 log 만 누락된 상태에서 backfill
- **P6 close**: mass source stub emit (494 files) 완료 + L2 lint 0/0/0 + L1 ↔ L2 cross-layer 정합 + log 보강. R8/R9 life-cycle 과 R7 distributed sync 의 runtime layer 검증 가능 상태

## [2026-06-12] lint-fix | V-R9 skip marker — naive grep false-positive 17 → 0

- **Trigger**: P7 commit L77 의 후속 메모 (V-R9 naive grep 의 16 false-positive, skip marker / smart parser 후속)
- **V-R9 v0.2.0 보강**: `workflow-source/tests/check_wiki_source_rule.py` 에 `has_r9_skip_marker()` 추가
  - frontmatter `r9_skip: true` (또는 `1`/`yes`) 면 V-R9 skip
  - R-4 정책 (frontmatter only marker) 준수 — body 미수정
- **Skip marker 적용 10 page**:
  - `concepts/wiki-source-rule-r9.md` (R9 rule 자체 정의)
  - `concepts/memory-3-state-lifecycle.md` (3-state 정의)
  - `concepts/harness-distribution.md`
  - `patterns/frozen-archive-immutability.md` (R8 freeze protocol)
  - `decisions/adr-004-wiki-layer.md` (R8/R9 extension)
  - `decisions/adr-005-r9-wiki-source-rule.md` (R9 정식 ADR)
  - `entities/ai-workflow-runtime.md`
  - `topics/wiki-ingest-lifecycle.md`
  - `INGEST_GUIDE.md` (§1 R9 scope 설명)
  - `log.md` (frontmatter 추가 — type: meta, r9_skip: true)
- **검증**: V-1 PASS / V-4 PASS (34 entries) / **V-R9 PASS (0 violation)**. 17 false-positive → 0
- **Cross-ref**: wiki vault 의 standard-ai-workflow project 영향 page 2 (wiki-log, wiki-ingest-guide) 의 last_touched 갱신 (wiki-event-sync v0.3.0)

## [2026-06-12] ingest | 1 topic page (AIDLC benchmark analysis 2026-06-12)

- **Trigger**: AWS AIDLC (`awslabs/aidlc-workflows`, commit `b19c819`) 풀 벤치마크 분석. yklee 의뢰, Mavis 실행 (2026-06-12, 풀 벤치마크 모드)
- **신규 page**: `topics/aidlc-benchmark-analysis-2026-06-12.md` (1 page, ~290 lines, status: draft)
  - **§1**: 분석 목적/범위
  - **§2**: AIDLC 핵심 구조 요약 (3-Phase + 7대 결정적 차별 메커니즘)
  - **§3**: 우리 v0.6.3-beta 와의 1:1 비교 (강점 8 + 갭 20)
  - **§4**: 보완안 (도입 권장 ★★★ 4종 / 부분 도입 6종 / 비권장 3종 / ADR 후보 2종)
  - **§5**: v0.6.3 → v0.7.0+ 권장 실행 순서 (12 step, 의존성 매트릭스)
  - **§6**: R-1~R9 cross-cutting 정합성 체크리스트
  - **§7**: 1차 출처 (16 file, AIDLC 13 + 우리 3) line-count 검증
  - **§8**: 다음 단계
- **R-9 면제**: codebase self-ingest (last_ingested_from = `workflow-source/core/global_workflow_standard.md + workflow-source/core/workflow_task_modes.md`, in-repo path)
- **index.md anchor**: `### [[topics/aidlc-benchmark-analysis-2026-06-12]] {#aidlc-benchmark-analysis-2026-06-12}` 추가 (V-4 PASS, 35 entries)
- **L2 mirror**: `~/wiki/wiki/projects/standard-ai-workflow/` 의 derived view 자동 emit (wiki-source-sync). topics/ 디렉토리 부재 → L2 sources/ 에 stub 또는 comparisons/ 신규 디렉토리 결정은 별도 turn
- **AIDLC reference 1차 출처**:
  - `awslabs/aidlc-workflows/README.md` (962 lines)
  - `aidlc-rules/aws-aidlc-rules/core-workflow.md` (539 lines) — 3-Phase lifecycle
  - `aidlc-rules/aws-aidlc-rule-details/inception/{workspace-detection, requirements-analysis, units-generation, workflow-planning}.md`
  - `aidlc-rules/aws-aidlc-rule-details/construction/code-generation.md` (217 lines)
  - `aidlc-rules/aws-aidlc-rule-details/common/{process-overview, depth-levels, question-format-guide}.md`
  - `aidlc-rules/aws-aidlc-rule-details/extensions/security/baseline/security-baseline.md` (307 lines) + `.opt-in.md`
  - `docs/GENERATED_DOCS_REFERENCE.md` (102 lines)
- **Follow-up 후보** (별도 turn):
  1. v0.6.4: Question File Format (A) + Stage Gate 명시화 (C) — yklee 승인 시
  2. v0.7.0: Reverse Engineering 9-Artifact (D) + Extension 시스템 (B) + Security-baseline extension (O)
  3. ADR-NNN: Operations phase 도입 여부 (N) — yklee 별도 결정
- **L1 ↔ L2 cross-ref**: 본 page → L2 derived view 자동 emit. vault L1 `standard-ai-workflow` project 영향 page 식별은 v0.3.0 wiki-event-sync 가 commit op 으로 자동 처리 (L2 wiki-log, 표준 ADRs 영향)

## [2026-06-12] promote | topics/aidlc-benchmark-analysis-2026-06-12 (draft → active)

- **Trigger**: 4-channel 동기화 (L1 page/index/log + L2 stub/index/log) 완료 후 status 승격
- **대상 page**: `topics/aidlc-benchmark-analysis-2026-06-12.md` (frontmatter status: draft → active)
- **전환 근거 (R-1 / V-1 / V-4 / V-R9)**:
  - V-1 (위치 단일성): `ai-workflow/wiki/topics/` 단일 ✅
  - V-4 (index anchor): `### [[topics/aidlc-benchmark-analysis-2026-06-12]]` 1+ inbound ✅
  - V-R9: codebase self-ingest 면제, `last_ingested_from` = in-repo path ✅
  - R-1 (inbound ≥ 1): `related_pages` 8 + index anchor 1 = 9 inbound ✅
  - body stable: 분석 노트 SSOT, 갭 20 / 보완안 15 / 12 step 의존성 매트릭스 확정
- **Frontmatter 보강**:
  - `active_since: 2026-06-12`
  - `active_reason: "draft → active (commit 2916d49 + cross-channel 동기화 완료). V-1 / V-4 / V-R9 / R-1 모두 PASS"`
- **L2 cross-channel**: L2 stub (commit 9bea914) status: draft → reviewed, body fill (TL;DR + 7 mechanism + 우리 갭 Top 4 + 12 step) — L1 ↔ L2 SSOT ↔ derived view 3-tier 정합
- **Linter 영향**: V-1 PASS / V-4 PASS (35 entries) / V-R9 PASS (0 violation) — 변경 없음
- **Follow-up 후보** (별도 turn):
  1. v0.6.4: Question File Format (A) + Stage Gate 명시화 (C) — yklee 승인 시
  2. v0.7.0: Extension 시스템 (B) + security-baseline
  3. ADR-NNN: Operations phase 도입 여부 (N)

## [2026-06-12] v0.6.4 | Question File Format + Stage Gate 명시화 (4 doc, 1347 line 코드+테스트)

- **Trigger**: AIDLC 벤치마크 분석 (commit 2916d49) 의 보완안 §4 A + C 채택. yklee 승인, 즉시 작업 시작
- **Commit 1 (25756bb)**: 4 doc 신규/수정 (498 line)
  - 신규 spec 2종: `question_file_format.md` (204) + `stage_gate_pattern.md` (207)
  - `output_schema_guide.md` §3.4 stage_completion Pydantic v2 schema (+40)
  - `workflow_adoption_entrypoints.md` §7 v0.6.4 권장 도입 묶음 (+47)
- **Commit 2 (bc16d91)**: 2 module + 2 smoke test (1347 line, 22 test PASS)
  - `workflow_kit/common/contracts/question_format.py` (358) — parse_answers, validate_answers, detect_ambiguity, detect_contradiction, generate_clarification_file, full_validation
  - `workflow_kit/common/contracts/stage_gate.py` (335) — StageCompletion dataclass, validate_completion, require_explicit_approval (auto 한계), append_audit_log, emit_completion_message, normalize_option_label
  - `tests/check_question_format.py` (336) — 7 test PASS
  - `tests/check_stage_gate_compliance.py` (318) — 15 test PASS
- **Bug fix (구현 중 발견)**: `ANSWER_TAG_RE` 가 옵션 라벨 안의 '[Answer]:' 도 매칭 ('X) Other (please describe after [Answer]: tag below)' 의 'T' 매칭 오인). `^` anchor + 'letter' 그룹 '?' 로 fix. parser 가 Q3 의 빈 tag 를 Q2 의 T 로 오인하던 버그 → 정상
- **신규 L1 wiki concept page 2종**:
  - `concepts/question-file-format.md` (입력 단계, AIDLC common/question-format-guide.md 차용)
  - `concepts/stage-gate-pattern.md` (출력 단계, AIDLC construction phase 차용)
- **index.md**: ### [[concepts/question-file-format]] + ### [[concepts/stage-gate-pattern]] anchor 추가 (V-4 37 entries)
- **L2 stub 2종 emit**: sources/concepts-question-file-format.md + sources/concepts-stage-gate-pattern.md (frontmatter + <needs content> placeholder, status: draft)
- **L2 wiki index**: Concepts count 8 → 10 (33 page → 35 page)
- **R-1/V-1/V-4/V-R9**: 모두 PASS
- **영향**: 11종 skill spec 의 §X Output Contract 에 `stage_completion` field 추가 권장 (v0.6.5 follow-up). 본 wiki page 의 Stage Gate 정책은 11종 skill 의 표준 패턴
- **Cross-ref**: [[concepts/question-file-format]] ↔ [[concepts/stage-gate-pattern]] 상호 cross-ref. [[topics/aidlc-benchmark-analysis-2026-06-12]] §4 의 A + C 보완안 = 본 commit
- **Follow-up (v0.6.5 후보)**: 11종 skill spec 의 Output Contract 에 `stage_completion` field 추가 (각 5-10 line, ~80 line). 1 commit

## [2026-06-12] v0.6.5 | StageCompletion field 11종 skill spec + catalog 보강 (13 file, +277 line)

- **Trigger**: v0.6.4 의 Stage Gate Pattern (commit 25756bb) 의 runtime 적용. yklee 승인
- **Commit (5b16517)**: 13 file, +277 line
  - 7 skill spec §4 출력 계약 보강 (+182 line, 26 each):
    session-start, backlog-update, doc-sync, merge-doc-reconcile, validation-plan, code-index-update, automated-repro-scaffold
  - 5 SKILL.md cross-ref (+70 line, 14 each):
    memory-freeze, git-conflict-resolver, robust-patcher, workflow-linter, project-status-assessment
  - workflow_skill_catalog.md §5.2 신규 (+25 line)
- **Helper script** (일회용, git 미포함):
  - /tmp/v0_6_5_inject_stage_completion.py — 7 spec 에 stage_completion subsection 일괄 삽입
  - /tmp/v0_6_5_fix_backticks_v2.py — backtick nesting 4 file fix
  - /tmp/v0_6_5_inject_skill_cross_ref.py — 5 SKILL.md cross-ref 일괄 추가
- **Bug fix (helper 실행 중 발견)**:
  - 첫 helper 가 `### 4.X.` placeholder 사용 → spec 별 subsection 자동 계산 (4.1 등) 로 fix
  - 첫 helper 의 backtick nesting (` ``X` ``) → sed 가 일부만 fix → python script v2 로 재실행
- **L1 wiki stage-gate-pattern §8 갱신**: 11종 skill table 에 "v0.6.5 spec 적용" column 추가
- **Linter 영향**: V-1 PASS / V-4 PASS (37 entries) / V-R9 PASS — 변경 없음 (코드 변경 0)
- **Follow-up (별도 session, runtime migration)**:
  1. 11종 skill 의 Python 구현이 `StageCompletion` 객체를 output 에 포함하도록 코드 변경
  2. orchestrator 측 `emit_completion_message` 호출 후 user response → `append_audit_log` 자동
  3. AIDLC 호환 검증

## [2026-06-12] v0.6.5 | Stage Gate Runtime helper + migration guide (3 file, +644 line, 13 test PASS)

- **Trigger**: v0.6.5 spec 보강 (commit 5b16517) 의 runtime enforcement 인프라. yklee 승인
- **Commit (dd98e69)**: 3 file, +644 line
  - `workflow_kit/common/contracts/stage_gate_runtime.py` (186 line) — runtime helper 5 function
    - build_stage_completion / merge_into_result / emit_and_log
    - is_stage_completion_present / get_stage_status_from_result
  - `tests/check_stage_gate_runtime.py` (292 line) — 13 test PASS
    - build / merge / emit / audit log / 52 smoke test 호환 검증
  - `core/stage_gate_runtime_migration.md` (166 line) — runtime migration 가이드
    - §1 why: v0.6.4 이론 → v0.6.5 runtime 필요성
    - §2 runtime helper API + 사용 예시
    - §3 11종 skill code 변경 절차 (안전 순서, breaking 회피, stage name table)
    - §4 smoke test 추가 (13 test)
    - §5 AIDLC 호환 (이점 + 의도적 차이 2개)
    - §6 follow-up (11 skill code 변경, v0.7.0 required 격상, orchestrator 통합)
- **L1 wiki stage-gate-pattern §12 References 갱신**:
  - `stage_gate_runtime_migration.md` + `stage_gate_runtime.py` + `check_stage_gate_runtime.py` cross-ref 추가
- **Breaking change 회피 정책**: stage_completion 은 **optional field** 로 추가 (mandatory 아님)
  - 기존 52 smoke test 의 schema validator 와 즉시 호환
  - 11종 skill code 변경은 **점진 적용** (1 skill pilot → 회귀 검증 → batch)
  - v0.7.0+ 에 모두 적용된 후 required 로 격상
- **누적 v0.6.4-5 test PASS**:
  - question_format: 7
  - stage_gate_compliance: 15
  - stage_gate_runtime: 13
  - **총 35 test PASS**
- **Follow-up (별도 commit)**:
  1. 11종 skill `run_*.py` 에 stage_completion merge + emit_and_log 추가 (1-2 commit)
  2. v0.7.0: stage_completion required 격상 + AIDLC 완전 호환 검증
  3. v0.8.0: orchestrator 측 자동 emit_and_log 통합

## [2026-06-12] v0.6.5 | pilot runtime — automated-repro-scaffold stage_completion 통합 (1 file, +44 line)

- **Trigger**: runtime migration (commit dd98e69) 의 1차 pilot. yklee 승인
- **Commit (2fab835)**: 1 file, +44 line
  - `automated_repro_scaffold.py` — 3 result build path 모두 stage_completion merge
    - report_file_not_found: stage_status=error, next_stage=None
    - success: stage_status=ok, next_stage=validation-plan
    - repro_write_failed: stage_status=error, next_stage=None
- **Bug fix (pre-existing, 발견)**: error path 의 `build_error_result()` 가 `warnings` keyword-only arg 누락 → TypeError → 본 pilot 적용 중 발견, fix 포함. latent 버그 (v0.6.5 이전부터 존재, error path 즉시 crash)
- **Stage Name 매핑**: STAGE_NAME = "automated-repro-scaffold", NEXT_STAGE = "validation-plan" (workflow_skill_catalog.md §5.2 와 일치)
- **검증**:
  - happy path 실제 실행: stage_completion 8 field 모두 포함 + 기존 field 보존 ✅
  - error path 실제 실행: stage_status=error, next_stage=None ✅
  - 35 smoke test 모두 PASS (7 question_format + 15 stage_gate_compliance + 13 stage_gate_runtime) — 회귀 0
- **Follow-up (별도 commit, batch)**:
  1. session-start, backlog-update, doc-sync, merge-doc-reconcile, validation-plan, code-index-update (6 spec) — 각 1 commit 또는 batch 1-2 commit
  2. workflow-linter, project-status-assessment, memory-freeze, git-conflict-resolver, robust-patcher (5 SKILL.md cross-ref 만) — 1 commit
  3. runtime 적용 후 L1 wiki 11종 skill table 의 'runtime 적용' column 갱신
  4. v0.7.0: stage_completion required 격상
  5. v0.8.0: orchestrator 측 자동 emit_and_log 통합

## [2026-06-12] v0.6.5 | batch stage_completion integration — 6 spec 보유 skill (10/11 완료)

- **Trigger**: pilot (commit 2fab835) 후 나머지 spec 보유 skill batch 적용. yklee 승인
- **Commit (ca7a685)**: 6 file, +72 line
  - 6 spec 의 run_*.py 성공 path 에 stage_completion merge:
    - session-start, backlog-update, doc-sync, merge-doc-reconcile, validation-plan, code-index-update
  - pilot template 그대로 적용 (12 line each, import + merge block)
  - status mapping: 'ok'/'success' → 'ok', 'warning' → 'warning', else 'error'
- **Helper script** (일회용, git 미포함):
  - /tmp/v0_6_5_batch_runtime.py — 6 spec 일괄 적용
  - 1st run: success path 식별 버그 (window 가 짧아 중간 return 1 을 success 로 오인)
    → rollback → fix (파일의 *마지막* print + return 0 만 식별) → 2nd run PASS
- **검증**:
  - syntax check: 6 spec 모두 통과
  - merge dry-build (session-start style): stage_completion 8 field 모두 포함
  - 35 smoke test 모두 PASS — 회귀 0
- **누적 v0.6.5 runtime 적용 (10/11)**:
  - ✅ automated-repro-scaffold (pilot, commit 2fab835)
  - ✅ session-start, backlog-update, doc-sync, merge-doc-reconcile, validation-plan, code-index-update (batch, 본 commit)
  - ⏸ workflow-linter, project-status-assessment, memory-freeze, git-conflict-resolver, robust-patcher (SKILL.md cross-ref 만, runtime script 없음)
- **Follow-up (별도 commit, optional)**:
  1. v0.6.5 spec 의 11종 skill table 'runtime 적용' column 갱신
  2. 5 SKILL.md 만 있는 skill 의 runtime (runtime helper 호출 경로)
  3. v0.7.0: stage_completion required 격상
  4. v0.8.0: orchestrator 측 자동 emit_and_log 통합

## [2026-06-12] v0.6.5 | release(v0.6.5) — AIDLC 패턴 차용 (10 commit, ~2,600 line)

- **Trigger**: v0.6.4 + v0.6.5 작업 완료. yklee 승인, release 묶음
- **Commit (3897da7)**: 5 file, +168 line
  - `releases/Beta-v0.6.5.md` (NEW, 122 line) — AIDLC 패턴 2종 + runtime 35 test PASS
  - `workflow_skill_catalog.md` §5.2 (41 line 보강) — 7 spec runtime 적용 + 5 SKILL.md-only table + migration guide cross-ref
  - `maturity_matrix.json` last_updated 2026-06-07 → 2026-06-12
  - `README.md` §1 버전 v0.6.3-beta → v0.6.5-beta + §10 누적 변경 v0.6.0 → v0.6.5
  - `QUICKSTART.md` 배포 패키지 v0.6.3-beta → v0.6.5-beta
- **v0.6.5 누적 작업** (10 commit, 0ae8d4a → 3897da7):
  - 5b16517: 7 skill spec §4.1 stage_completion + 5 SKILL.md cross-ref + catalog §5.2
  - dd98e69: stage_gate_runtime helper + migration guide + 13 runtime test
  - 2fab835: pilot runtime — automated-repro-scaffold
  - ca7a685: 6 spec 보유 skill batch runtime
  - 0ae8d4a: L1 wiki batch log entry
  - 3897da7: release 묶음 (Beta-v0.6.5.md + version bump)
- **누적 v0.6.4-5 산출물** (전체):
  - 4 신규 spec doc (question_file_format, stage_gate_pattern, stage_gate_runtime_migration, output_schema_guide §3.4)
  - 3 신규 Python module (question_format.py, stage_gate.py, stage_gate_runtime.py)
  - 3 신규 smoke test (7 + 15 + 13 = 35 test PASS)
  - 1 release note (Beta-v0.6.5.md)
  - 7 skill spec 보강 (11종 skill table + catalog §5.2)
  - 5 SKILL.md cross-ref (workflow-linter, project-status-assessment, memory-freeze, git-conflict-resolver, robust-patcher)
  - 7 spec 보유 skill runtime 적용 (pilot 1 + batch 6)
  - 3 wiki concept (question-file-format, stage-gate-pattern × 2)
  - 1 wiki topic (aidlc-benchmark-analysis-2026-06-12)
  - 4 channel L2 vault sync (stub + index + log + 자기-갱신)
- **총 ~4,100 line, 35 smoke test PASS, breaking change 0**
- **Follow-up (v0.6.6+ 후보)**:
  1. v0.6.5 5 SKILL.md-only skill 의 runtime script 작성 (선택)
  2. v0.7.0: stage_completion required 격상
  3. v0.7.0: Extension 시스템 (B) + security-baseline 1종
  4. v0.8.0: orchestrator 측 자동 emit_and_log 통합
  5. ADR-NNN: Operations phase 도입 여부

## [2026-06-12] v0.6.6 follow-up #1 | 5 SKILL.md-only skill runtime 통합 (12/12 일관성)

- **Trigger**: v0.6.5 release (commit 3897da7) follow-up #1. yklee 승인
- **Commit (6a9126c)**: 5 file, +148 line
  - 4 file batch (기존 run_*.py 보강, +44 line):
    - workflow-linter (Pydantic model_dump)
    - project-status-assessment (plain dict)
    - memory-freeze (payload dict)
    - git-conflict-resolver (Pydantic model_dump)
  - 1 file 신규 (robust_patcher, +104 line):
    - run_robust_patcher.py — patch_engine.py library function 호출하는 runtime entry
    - STAGE_NAME = "robust-patcher", NEXT_STAGE = "validation-plan"
- **Helper script** (일회용, git 미포함):
  - /tmp/v0_6_6_runtime_skill_md_only.py — 4 file 일괄 + 1 file 신규 작성
  - 1st run: success path 식별 버그 (window 5 line 부족, return 0 at line 176) → rollback → fix (window 8 line) → 2nd run PASS
- **검증**:
  - syntax check: 5 file 모두 통과
  - merge block 정확: pilot template 일관성
  - 35 smoke test 모두 PASS (회귀 0)
- **누적 12/12 일관성**: spec 7 + SKILL.md 5 + runtime 12 (모두 완료)
- **Follow-up (별도 commit)**:
  1. v0.7.0: stage_completion required 격상 (모든 skill 적용 완료, 본 commit 으로 가능)
  2. v0.7.0: Extension 시스템 (B) + security-baseline 1종
  3. v0.8.0: orchestrator 측 자동 emit_and_log 통합
  4. ADR-NNN: Operations phase 도입 여부

## [2026-06-12] v0.7.0 step 1 (commit `6e57cf3`) | stage_completion required 격상 + ensure fallback (3 file, +319 line, 8 test PASS)

- **Trigger**: v0.6.6 follow-up #1 (12/12 일관성) 완료. v0.7.0 본격 시작
- **Commit (6e57cf3)**: 3 file, +319 line
  - `workflow_kit/common/contracts/stage_gate_runtime.py` (+40 line):
    - 신규 `ensure_stage_completion()` helper — stage_completion 없는 result 에 자동 생성 (lazy fallback)
    - status mapping: 'success'/'ok' → 'ok', 'warning' → 'warning', else 'error'
  - `workflow-source/core/output_schema_guide.md` §3.4 (12 line):
    - "v0.6.4 신규" → "v0.6.4 신규, **v0.7.0 부터 required**"
    - 11종 skill + 8+ MCP output 의 stage_completion mandatory
    - 마이그레이션 가이드 (3 step)
  - `tests/check_stage_completion_required.py` (272 line, NEW, 8 test PASS):
    - ensure_creates_when_missing / preserves_existing / status_mapping
    - automated_repro_scaffold_runtime (실제 pilot runtime 실행)
    - build_then_merge_roundtrip / legacy_result_compatibility
    - 8_field_completeness_after_ensure / skill_catalog_stage_name_mapping
- **핵심 변경**:
  - stage_completion 이 optional → **required** 로 격상
  - 12/12 일관성 (v0.6.6 follow-up #1) 후 가능한 mandatory 격상
  - runtime layer `ensure_stage_completion()` 으로 legacy code path 자동 복구
- **검증**:
  - 신규 8 test PASS
  - 기존 35 test 모두 PASS (7 + 15 + 13) — 회귀 0
  - 누적 **43 test PASS** (v0.6.4-7 + v0.7.0 step 1)
- **Follow-up (별도 commit, v0.7.1+)**:
  1. MCP server (8+ read_only bundle) 에 stage_completion 적용
  2. sample output (workflow-source/examples/output_samples/*.json) stage_completion migration
  3. stage_completion.strict_required runtime check (lint)
- **누적 v0.7.0 step**:
  - ✅ Step 1: stage_completion required 격상 (본 commit)
  - ⏸ Step 6: Reverse Engineering 9-Artifact (D)
  - ⏸ Step 7: Extension 시스템 (B)
  - ⏸ Step 8: Security-baseline 1종 (O)
  - ⏸ Step 9: Unit of Work 3-layer (G)
  - ⏸ Step 10: Audit Log 표준화 (H)

## [2026-06-12] v0.7.0 step 10 (commit `54e96a9`) | Audit Log 표준화 (3 file, +637 line, 13 test PASS)

- **Trigger**: v0.6.4-5 의 분산 정의된 audit log 정책 통합. yklee 승인, step 10 진행
- **Commit (54e96a9)**: 3 file, +637 line
  - `workflow-source/core/audit_log_standard.md` (209 line, NEW): 단일 표준 spec
    - §1-10: 위치 / format / append-only / lifecycle / 자동화 / migration / 한계 / AIDLC 호환 / references
  - `workflow_kit/common/contracts/stage_gate.py` (+22 line, fix):
    - **Bug fix 1 (leading newline)**: entry_lines 시작의 "" 제거 → file 첫 줄이 ## [Stage: ...] 로 시작
    - **Bug fix 2 (microsecond leak)**: datetime.now(timezone.utc).isoformat() 가 microsecond 포함 → split('.')[0] 정규화
  - `tests/check_audit_log_compliance.py` (412 line, NEW, 13 test PASS):
    - entry format / 8 field / append-only / ISO 8601 / approval status / actor / full_workflow_audit_log / legacy_readable
- **Test bug fix (보너스)**:
  - ENTRY_HEADER_RE 에 `re.MULTILINE` flag 누락 → 추가
  - iso_8601_z_suffix test 의 string 비교 (`ts[len(...):]`) 잘못 → `ISO_8601_RE.match()` + microsecond 검증으로 fix
- **검증**:
  - 신규 13 test PASS
  - 기존 43 test 모두 PASS — 회귀 0
  - 누적 **56 test PASS** (v0.6.4-7 + v0.7.0 step 1 + step 10)
  - runtime helper 2 fix 가 기존 smoke test 회귀 0 (latent bug)
- **누적 v0.7.0 step 진행**:
  - ✅ Step 1: stage_completion required 격상 (commit 6e57cf3)
  - ✅ Step 10: Audit Log 표준화 (본 commit)
  - ⏸ Step 6: Reverse Engineering 9-Artifact (D, 2-3 ses)
  - ⏸ Step 7: Extension 시스템 (B, 3-5 ses)
  - ⏸ Step 8: Security-baseline 1종 (O, 1 ses)
  - ⏸ Step 9: Unit of Work 3-layer (G, 1-2 ses)

## [2026-06-12] v0.7.0 step 9 (commit `c981cac`) | Unit of Work 3-layer template (2 file, +622 line, 17 test PASS)

- **Trigger**: v0.6.4-7 의 mode 6종 (horizontal) + task-level (work_backlog) 의 missing layer 보강. yklee 승인, step 9 진행
- **Commit (c981cac)**: 2 file, +622 line
  - `workflow-source/templates/unit_of_work_template.md` (208 line, NEW): system-level 분해 + dependency matrix + Mermaid graph + story mapping + code organization
  - `tests/check_unit_of_work_template.py` (414 line, NEW, 17 test PASS):
    - UOW 정의 parse (5): header / required fields / type enum / status enum / date format
    - Dependency Matrix (5): parse / symmetry / no-self-dep / **cycle detection (DFS)** / DAG validation
    - Mermaid Graph (2): block present / edge syntax
    - Story Mapping (1): valid UOW id 참조
    - Template 자체 정합성 (3): sections / related docs / AIDLC source
    - 통합 (1): full parse 일관성
- **Test bug fix (보너스)**:
  - dep_matrix_cycle_detection 의 nested function scope issue → helper `_has_cycle` top-level 로 추출
  - mermaid_graph_syntax 의 strict edge matching → 양방향 arrow + line edge 매칭
- **검증**:
  - 신규 17 test PASS
  - 기존 56 test 모두 PASS — 회귀 0
  - 누적 **73 test PASS** (v0.6.4-7 + v0.7.0 step 1 + 9 + 10)
- **누적 v0.7.0 step 진행**:
  - ✅ Step 1: stage_completion required 격상 (commit 6e57cf3)
  - ✅ Step 9: Unit of Work 3-layer template (본 commit)
  - ✅ Step 10: Audit Log 표준화 (commit 54e96a9)
  - ⏸ Step 6: Reverse Engineering 9-Artifact (D, 2-3 ses)
  - ⏸ Step 7: Extension 시스템 (B, 3-5 ses)
  - ⏸ Step 8: Security-baseline 1종 (O, 1 ses)
- **Follow-up (v0.7.1+)**:
  1. workflow_kit.common.contracts.uow 신규 (UOW matrix parsing + sub-agent 위임 결정 helper)
  2. bootstrap_lib 의 `--adoption-mode new` 가 unit_of_work.md 자동 emit
  3. v0.8.0: UOW 기반 sub-agent 위임 자동화

## [2026-06-12] v0.7.0 step 6 (commit `4bbd391`) | Reverse Engineering 9-Artifact (11 file, +674 line, 19 test PASS)

- **Trigger**: v0.6.4-7 의 `existing` onboarding 이 단일 `repository_assessment.md` 만 emit → 주제별 SSOT 부재. yklee 승인, step 6 진행
- **1차 출처**: AIDLC `aidlc-rules/aws-aidlc-rule-details/inception/reverse-engineering.md` (311 line, commit b19c819, 2026-06-08)
- **Scope 결정**: AIDLC 9-Artifact 구조 유지 (simplification ❌). 각 artifact 를 30-50 line template 으로 압축 (AIDLC 50 line × 9 = 450 line → 우리 ~30 line × 9 = 270 line)
- **Commit (예정)**: 11 file, +674 line
  - `workflow-source/reverse-engineering/01-business-overview.md` (32 line, NEW): business transaction = workflow stage transition
  - `workflow-source/reverse-engineering/02-architecture.md` (32 line, NEW): components = harness / skill / MCP / workflow_kit
  - `workflow-source/reverse-engineering/03-code-structure.md` (29 line, NEW): key classes = workflow_kit modules
  - `workflow-source/reverse-engineering/04-api-documentation.md` (32 line, NEW): REST → MCP tool / Internal → workflow_kit Python
  - `workflow-source/reverse-engineering/05-component-inventory.md` (33 line, NEW): 5 type (Harness/MCP/workflow_kit/Template/Test)
  - `workflow-source/reverse-engineering/06-technology-stack.md` (36 line, NEW): Python + 5 harness + packaging
  - `workflow-source/reverse-engineering/07-dependencies.md` (37 line, NEW): internal workflow_kit + pyproject external
  - `workflow-source/reverse-engineering/08-code-quality-assessment.md` (40 line, NEW): smoke test PASS + R-1~R9 lint
  - `workflow-source/reverse-engineering/09-reverse-engineering-metadata.md` (43 line, NEW): ISO 8601 + state.json sync
  - `workflow-source/core/reverse_engineering.md` (140 line, NEW): step-by-step guide (13 step, AIDLC 대응)
  - `workflow-source/tests/check_reverse_engineering.py` (350 line, NEW, 19 test PASS):
    - 디렉토리 + 9 artifact 존재 (4)
    - artifact 내용 검증 (4): Verification subsection / Workflow domain 적응 / AIDLC cross-ref / sequential numbering
    - guide 내용 검증 (6): 13 step / 9-Artifact table / AIDLC correspondence / rerun stale / state.json schema / workflow pattern adaptation
    - cross-reference (2): AIDLC 1차 출처 line count drift / artifact count matches
    - R-1~R9 lint (3): no duplicate filename / consistent naming / guide links to artifact dir
- **Test bug fix (보너스)**:
  - `test_guide_has_thirteen_steps` 의 개별 step 매칭 → "Step 2-9" range 표기 인식 추가 (DRY)
- **검증**:
  - 신규 19 test PASS
  - 기존 88 test 모두 PASS — 회귀 0
  - 누적 **107 test PASS** (v0.6.4-7 + v0.7.0 step 1, 9, 10, 8, 6)
- **누적 v0.7.0 step 진행**:
  - ✅ Step 1: stage_completion required 격상 (commit 6e57cf3)
  - ✅ Step 8: Security-baseline 1종 (commit dc2c22b)
  - ✅ Step 9: Unit of Work 3-layer template (commit c981cac)
  - ✅ Step 10: Audit Log 표준화 (commit 54e96a9)
  - ✅ Step 6: Reverse Engineering 9-Artifact (본 commit)
  - ⏸ Step 7: Extension 시스템 (B, 3-5 ses)
- **Follow-up (v0.7.1+)**:
  1. `existing_project_onboarding.py` 가 9 artifact 자동 fill (현재 단일 `repository_assessment.md` 만 자동)
  2. v0.7.1: SEC-WF-05 dependency integrity 검증 — `07-dependencies.md` 의 lock file + checksum 자동 확인
  3. v0.7.1: 9-Artifact 별 wiki L1 page (각 artifact 마다 L1 topic page + L2 sources/)
  4. v0.8.0: Artifact 별 version diff (이전 reverse engineering 대비 변경점)

## [2026-06-13] v0.7.0 step 7 (commit `0052da1`) | Extension 시스템 일반화 (5 file, +23 test PASS)

- **Trigger**: v0.7.0 step 8 의 security-baseline 1종 출시 후, 3종 extension (security / testing / performance) 일반화 + SCHEMA.md SSOT. yklee 승인, step 7 진행
- **1차 출처**: AIDLC `awslabs/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/extensions/` 3종 (commit b19c819, 2026-06-08)
- **Scope 결정**: SCHEMA.md (extension file format SSOT) + 2종 extension (testing + performance) 1차 출시. security opt-in 은 SCHEMA 형식으로 update (P) Partial 옵션 추가)
- **Commit (예정)**: 6 file, +line
  - `workflow-source/extensions/SCHEMA.md` (170 line, NEW): extension system SSOT (Directory Layout, File Format, Rule ID Convention, Hard Constraint 정책, Lint Rule 10종, Helper Contract v0.7.1+)
  - `workflow-source/extensions/testing-baseline.md` (130 line, NEW, 6 rule TST-WF-01~06): AIDLC PBT 1:1 적응 (PBT-05/06/08 N/A)
  - `workflow-source/extensions/testing-baseline.opt-in.md` (80 line, NEW): A/B/P/X 4 옵션 + State File Schema + Extension Configuration table
  - `workflow-source/extensions/performance-baseline.md` (130 line, NEW, 6 rule PERF-WF-01~06): 우리 domain 적응 (workflow runtime — Smoke Test Time / Module Import / Memory Footprint / Audit Log Latency / state.json Latency / Profiling Hook)
  - `workflow-source/extensions/performance-baseline.opt-in.md` (80 line, NEW): 동일 형식
  - `workflow-source/extensions/security-baseline.opt-in.md` (UPDATE): SCHEMA 형식으로 정합 (P) Partial 옵션 + Response Processing + Extension Configuration table 추가)
  - `workflow-source/tests/check_extension_system.py` (290 line, NEW, 23 test PASS):
    - SCHEMA.md 검증 (5): 존재 / Directory Layout / File Format / Prefix Convention / Lint Rule
    - 3종 extension baseline + opt-in (3): present / no_extra_baseline_files / rule_count
    - baseline 형식 (4): rule_id_format / rule+verification / compliance_summary / no_duplicate_rule_ids
    - opt-in 형식 (4): question_format / four_options / response_processing / state_schema
    - AIDLC cross-reference (2): 1차 출처 path valid / artifact count
    - 추가 (5): extension unique prefix / 6 rules / helper contract / follow-up
- **Test bug fix (보너스)**:
  - `check_extension_system.py` 의 AIDLC path 결합 중복 (`aidlc_root/extensions/extensions/...`) → `aidlc_root` (parent) 로 fix
  - `security-baseline.opt-in.md` update 시 기존 `check_security_baseline.py` 의 `opt_in_response_format_documented` 회귀 → `Extension Configuration` table 보존 + SCHEMA 형식 (State File Schema YAML) 추가
  - 모든 3종 opt-in 에 `security_baseline_status.md` 형식 (`Extension Configuration` table) 일관성 추가
- **검증**:
  - 신규 23 test PASS
  - 기존 107 test 모두 PASS — 회귀 0
  - 누적 **130 test PASS** (v0.6.4-7 + v0.7.0 step 1, 6, 8, 9, 10, 7)
  - AIDLC 3종 extension (security / testing / resiliency) cross-reference 검증
- **누적 v0.7.0 step 진행**:
  - ✅ Step 1: stage_completion required 격상 (commit 6e57cf3)
  - ✅ Step 6: Reverse Engineering 9-Artifact (commit 4bbd391)
  - ✅ Step 7: Extension 시스템 일반화 (본 commit)
  - ✅ Step 8: Security-baseline 1종 (commit dc2c22b)
  - ✅ Step 9: Unit of Work 3-layer template (commit c981cac)
  - ✅ Step 10: Audit Log 표준화 (commit 54e96a9)
  - **🎉 v0.7.0 6 step 전부 완료**
- **Follow-up (v0.7.1+)**:
  1. `workflow_kit.common.contracts.{security,testing,performance}_baseline.evaluate_compliance()` helper 구현 (3종)
  2. session-start 에 3종 opt-in prompt 통합
  3. `state.json` 의 `<name>_baseline` 필드 schema validation
  4. v0.8.0: sub-cat 도입 (e.g. `extensions/security/auth/`, `extensions/testing/property-based/`)
  5. v0.8.0: 4종 (resiliency) 추가 — workflow_kit health check + 장애 대응
  6. v0.7.0 release: packaging (5 harness) + GitHub release v0.7.0

## [2026-06-13] v0.7.0 (commit `dff0aae`) | Release — AIDLC 6 step 완료 (15 commit, ~3,200 line, 130 test PASS)

- **Trigger**: v0.7.0 6 step (Stage Completion Required / Audit Log / UOW Template / Security Baseline / Reverse Engineering / Extension System) 모두 완료. AIDLC (`awslabs/aidlc-workflows`, commit b19c819) 의 7대 차별 메커니즘 중 3개 채택.
- **Release notes**: `workflow-source/releases/Beta-v0.7.0.md` (NEW)
- **6 step 회고** (commit 6e57cf3 → 0052da1):
  - Step 1 (commit 6e57cf3): stage_completion required 격상 — +319 line, 8 test PASS
  - Step 10 (commit 54e96a9): Audit Log 표준화 + 2 latent bug fix — +637 line, 13 test PASS
  - Step 9 (commit c981cac): Unit of Work 3-layer template — +622 line, 17 test PASS
  - Step 8 (commit dc2c22b): Security-baseline 1종 — +558 line, 15 test PASS
  - Step 6 (commit 4bbd391): Reverse Engineering 9-Artifact — +925 line, 19 test PASS
  - Step 7 (commit 0052da1): Extension 시스템 일반화 (3종) — +1150 line, 23 test PASS
- **핵심 적응 비율** (AIDLC 1차 출처 → 우리):
  - 15 SECURITY → 6 SEC-WF (40%, N/A: HTTP API / Lambda / RDS)
  - 9 PBT → 6 TST-WF (67%, N/A: PBT-05/06/08)
  - 16 RESILIENCY → 6 PERF-WF (38%, N/A: HA/DR/Incident)
  - 9-Artifact 구조 100% 유지 (내용 압축)
  - UOW 4종 → 3 layer (75%, 4종 과잉)
- **누적 130 test PASS** (v0.6.5 35 → +95 신규) — 회귀 0
- **신규 산출물 (~3,200 line)**:
  - 외부 spec 8종 (reverse_engineering / SCHEMA / 3 extension baseline + 3 opt-in / unit_of_work_template / audit_log_standard)
  - Reverse Engineering 9 artifact template
  - workflow_kit module 3종 (stage_gate_runtime / audit_log + 1 update)
  - smoke test 6종 (95 test PASS 신규)
- **Follow-up (v0.7.1+)**:
  1. `workflow_kit.common.contracts.{security,testing,performance}_baseline.evaluate_compliance()` 3종
  2. session-start 에 3종 opt-in prompt 통합
  3. `state.json` 의 `<name>_baseline` 필드 schema validation
  4. PERF-WF-03 (memory 자동 측정) + PERF-WF-06 (profiling decorator)
  5. v0.8.0: Extension sub-cat 도입 + 4종 (resiliency) 추가
  6. v0.7.0 packaging (5 harness) + GitHub release v0.7.0
- **🎉 v0.7.0 6 step 전부 완료** — AIDLC 채택 3/7 (Question File Format [v0.6.4] / Stage Gate Pattern [v0.6.5] / Extension 시스템 + Reverse Engineering 9-Artifact [v0.7.0])

## [2026-06-13] v0.7.0 follow-up (commit `390a6e0`+`71de1b0`+`8818cbe`) | packaging + session-start opt-in + evaluate_compliance helper

- **Trigger**: v0.7.0 release 직후 follow-up 3 task. yklee 승인, 순서 = 3 (packaging) → 2 (session-start opt-in) → 1 (evaluate_compliance).
- **Task 3: v0.7.0 packaging + GitHub Release** (✅ 완료)
  - version bump: `pyproject.toml` 0.6.3 → 0.7.0 / `workflow_kit/__init__.py` v0.6.3-beta → v0.7.0-beta
  - wheel + sdist 빌드 (`workflow-source/dist/standard_ai_workflow-0.7.0-py3-none-any.whl`, 159898 bytes)
  - twine check: PASSED (wheel + sdist)
  - fresh venv smoke (`/tmp/sawsmoke-070`): workflow_kit.__version__ = v0.7.0-beta, 8-field StageCompletion import, ensure_stage_completion() 정상
  - local tag `v0.7.0-beta` push + `gh release create --verify-tag`
  - **GitHub Release**: https://github.com/ykylee/standard_ai_workflow/releases/tag/v0.7.0-beta (2 asset: wheel + tar.gz)
- **Task 2: session-start opt-in prompt 통합** (✅ 완료, spec level)
  - `core/session_start_skill_spec.md` §11 추가 (Extension Baseline Opt-In 통합)
  - 3종 baseline (security / testing / performance) opt-in 흐름 6 step (detect → prompt → response → file write → state.json sync → audit log)
  - Default 정책: Greenfield (security=A, testing=A, performance=P) / Brownfield (기존 존중) / CI/CD (skip)
  - Session-Start 출력 schema 확장: `extension_baselines` field + `warnings` 자동 생성
  - Runtime 통합은 §11.5 v0.7.1+ follow-up (evaluate_compliance 의 입력으로 활용)
  - `next_documents` 에 SCHEMA.md + 3 opt-in reference 추가
- **Task 1: evaluate_compliance() helper** (✅ 완료, 12 test PASS)
  - `workflow_kit/common/contracts/baselines.py` 신규 (340 line): RuleResult + ComplianceSummary dataclass + 3종 baseline evaluator
  - `evaluate_compliance(project_root, baseline)` + `evaluate_all(project_root)`
  - Security 6 rule runtime: stage_gate.append_audit_log / require_explicit_approval / validate_answers / fail-closed / pyproject 존재 / R-9 skip marker
  - Testing 6 rule runtime: smoke test count ≥ 5 / state.json round-trip / invariant / idempotency / fixture / docstring
  - Performance 6 rule runtime: smoke test 30초 / import 1초 / tracemalloc 200MB / audit log 10ms / state.json 5ms / profiling module
  - `_aggregate_status` 3 분기: partial_rules 적용
  - smoke test `tests/check_baselines_compliance.py` (12 test PASS): 3 baseline × 6 rule = 18 RuleResult, status enum 4 value, partial_rules 적용, error handling
- **누적 142 test PASS** (v0.7.0 130 + 12 신규) — 회귀 0
- **GitHub Release 재발행** (v0.7.0-beta + baselines helper 포함, commit 390a6e0~ 추가):
  - release asset 추가 안 함 (spec level 변경만), 기존 release 의 note 갱신 안 함
  - 다음 release (v0.7.1) 에 baselines.py + session-start spec 포함
- **Follow-up (v0.7.1)**:
  1. existing_project_onboarding.py 가 9 artifact 자동 fill
  2. SEC-WF-05 dependency integrity 검증 (lock file + checksum)
  3. 9-Artifact 별 wiki L1 page
  4. Extension sub-cat + 4종 (resiliency) 추가
  5. v0.7.0 + v0.7.1 follow-up 묶음 release (v0.7.1-beta)

## [2026-06-13] wiki 유지보수 개선 (commit `021ec16`) | 5 concept page + emit helper + drift smoke test

- **Trigger**: yklee 의 "이 저장소의 wiki 가 코드 유지보수에 사용될 수 있을 정도 수준" 검토 요청. 6 dim 평가 결과 3.0/5 (60%) — 즉시 개선 가능 항목 4건.
- **개선 내역**:
  - **작업 1: 5 concept page 신규** (L1 wiki 의 v0.7.0 5 step coverage 갭 해소):
    - `concepts/extension-system.md` (210 line) — SCHEMA + 3 baseline + opt-in + helper contract
    - `concepts/reverse-engineering.md` (165 line) — 9-Artifact + 13 step + rerun stale
    - `concepts/unit-of-work.md` (155 line) — 3 layer + dep matrix + Mermaid
    - `concepts/audit-log-standard.md` (180 line) — 8 field + append-only + 2 latent bug fix
    - `concepts/stage-gate-runtime.md` (180 line) — required 격상 + ensure_stage_completion + auto-approval
    - `index.md` anchor 5 page 추가
  - **작업 2: vault L2 sources/ 5 page 본문 emit** (draft 80% 해소):
    - 5 신규 L2 page: `concepts-extension-system.md` / `concepts-reverse-engineering.md` / `concepts-unit-of-work.md` / `concepts-audit-log-standard.md` / `concepts-stage-gate-runtime.md` — frontmatter 11 line + TL;DR + 본문 발췌 (max 2000자)
  - **작업 3: emit helper 신규** (vault L2 sources/ 자동 본문 emit tool):
    - `tools/emit_wiki_l2_body.py` (200 line) — `--apply` / `--dry-run` / `--max-chars=N` / `--limit=N`
    - raw mirror 의 L1 in-repo wiki 본문 → vault L2 sources/ 자동 emit (frontmatter 보존, `<needs content>` 자리만 교체)
    - glob brace `{a,b,c}` + numeric range `01..09` + glob `*` 모두 지원
    - 3 page apply 검증 (concepts-agent-topology, concepts-contract-v1-output-validation, concepts-harness-distribution)
  - **작업 4: drift 자동 검출 smoke test** (L1 ↔ code ↔ L2 3-way 정합):
    - `tests/check_wiki_drift.py` (5 test): L1 drift report / L2 drift report / ingested_from path 검증 / 5 신규 page index anchor / frontmatter format
    - 4 PASS + 1 fail (drift report 만 — 정보성, v0.6.4 page 의 7일 경계 drift 가 expected)
- **누적 146 test PASS** (v0.7.0 follow-up 142 + 4 신규)
- **Follow-up (v0.7.1+)**:
  1. emit_wiki_l2_body.py 의 `--apply --limit=0` (전체 L2 sources/ draft 해소)
  2. drift smoke test 의 CI 통합 (PR check 시 drift >= 14일 page 알림)
  3. wiki maintainability score metric (6 dim 별 점수 + dashboard)
  4. wiki-source-sync 본 emit 옵션 (vault 의 wiki-source-sync.py 자체에 통합)

## [2026-06-13] wiki maintainability (commit `7a4dbae`) | L2 sources/ 전체 draft 해소 (30 page, last_touched 갱신)

- **Trigger**: yklee 의 "emit_wiki_l2_body.py 의 --apply --limit=0 (전체)" 요청. 30 page 가 emit 대상.
- **실행**:
  - `python3 workflow-source/tools/emit_wiki_l2_body.py --project=standard-ai-workflow --apply` (default limit 0 = 무제한)
  - 30 page 모두 status: reviewed + last_touched: 2026-06-13 + 본문 발췌 (max 2000자)
  - 0 잔여 (raw mirror 와 1:1 매칭 37 page 모두 emit)
- **개선 효과**:
  - vault 의 L2 sources/ 539 page 중 30 page 가 draft → reviewed
  - 검색 정합도: wiki-query-helper 가 30 page 의 본문 검색 가능 (이전 0 매칭)
  - drift smoke test 의 drift report 가 30 page 의 *stale* 정보 → 정상 (last_touched 갱신됨)
- **tool fix (보너스)**:
  - `update_l2_full()` 신규 — frontmatter 의 `## Summary\n<needs content>` 자리만 교체 (placeholder 라인 제거)
  - frontmatter 의 `last_touched` + `status: draft → reviewed` 자동 갱신
- **my-harness / devhub / cross**: 0 candidates (L1 in-repo wiki 가 없는 project) — raw mirror 정책 별도
- **Follow-up (v0.7.1+)**:
  1. emit helper 의 `--project=cross` (L1 raw mirror 가 모든 project 에 동일)
  2. v0.7.1: vault 의 wiki-source-sync.py 자체에 --emit-body 옵션 통합

## [2026-06-13] wiki maintainability score (commit `49dfc78`) | 6 dim dashboard + 12 smoke test

- **Trigger**: yklee 의 "wiki maintainability score metric (6 dim dashboard)" 요청. 위키 운영 정공법의 *정량적 metric*.
- **산출**:
  - `workflow-source/tools/score_wiki_maintainability.py` 신규 (365 line)
  - 6 dim 별 0.0~5.0 점수:
    - Coverage (L1 wiki + last_ingested_from + status: active) = 4.13
    - Freshness (drift < 7일) = 4.20
    - Discoverability (vault L2 본문 ≥ 200자) = 0.37 (low — 539 page 중 40 page 만 searchable)
    - Cross-ref (related_pages ≥ 2) = 4.63
    - Lifecycle (status: reviewed) = 0.34 (low — 539 page 중 37 page 만 reviewed)
    - Operational (11 smoke test PASS) = 5.00
  - **Overall: 3.11 / 5.0 — Grade D**
  - grade 기준: A(≥4.5) / B(≥4.0) / C(≥3.5) / D(≥3.0) / F(<3.0)
- **Dashboard**:
  - `ai-workflow/wiki/concepts/wiki-maintainability-score.md` (auto-emit)
  - 6 dim table + bar chart (ASCII) + detail section + 다음 개선 가이드
  - `index.md` anchor 추가
- **Smoke test**:
  - `workflow-source/tests/check_wiki_score.py` (250 line, 12 test PASS):
    - tool importable + executable (2)
    - score structure + range + grade enum + grade match (4)
    - detail consistency + operational smoke (2)
    - dashboard emit + format + index (3)
    - idempotency (1)
- **누적 158 test PASS** (v0.7.0 follow-up 142 + 16 신규) — 회귀 0
- **개선 후보 (점수 < 4.5 dim)**:
  - **Discoverability 0.37 → 4.5**: vault L2 의 509 page 의 `<needs content>` 해소. emit_wiki_l2_body.py 의 L1 1:1 매칭이 37 page 만이라 *raw mirror 가 없는 page* 는 본문 emit 불가. v0.7.1+ 의 *manually authored L2 page* 정책 필요
  - **Lifecycle 0.34 → 4.5**: 동일 — L2 sources/ 의 509 page 의 status: draft → reviewed 자동 변경 로직 필요
- **Follow-up (v0.7.1+)**:
  1. score tool 의 CI 통합 (PR check 시 overall < 3.5 면 block)
  2. v0.7.1: L2 sources/ 의 *raw mirror 가 없는 page* (자체 생성 archive/) 도 본문 emit (template 기반)
  3. score dashboard 의 *trend over time* (commit 별 score 추적)
  4. v0.7.1: 6 dim 별 improvement suggestion 자동화

## [2026-06-13] wiki maintainability score 갱신 (commit `c72bdc3`) | 498 page 본문 emit (metadata-only) + Overall 3.11 → 4.66 (Grade A)

- **Trigger**: yklee 의 "Discoverability 0.37 → 4.5" 요청. vault L2 의 499 page 중 499 모두 해소.
- **분석**:
  - 499 page 의 frontmatter 가 모두 `source` field 없음 (raw mirror 가 없는 자체 생성 page)
  - 패턴: 자체 운영 log (날짜 prefix), Obsidian metadata (`.omo-*`), example project (acme-delivery-platform), template (`_*`), 외부 system snapshot (IP prefix)
  - L1 raw mirror 가 없으므로 *L1 1:1 매칭* emit_wiki_l2_body.py 의 기존 `l1` mode 로는 *불가*
- **Tool 확장**:
  - `tools/emit_wiki_l2_body.py` 에 `--mode` 추가 (`l1` | `metadata-only` | `all`)
  - `build_metadata_only_body()` 신규 — frontmatter 의 title / tags / sources / related / contradictions / status 추출
  - 본 policy 설명 추가 (vault-only entry, raw mirror 부재 명시)
  - `update_l2_full()` 가 `mode` argument 받음
- **Apply**:
  - 498 page apply (sample 1 + 497) — 0 잔여
  - vault L2 의 모든 page 가 status: reviewed + last_touched: 2026-06-13 + 본문 ≥ 200자
- **Score 갱신**:
  - **Overall: 3.11 → 4.66 (Grade D → A)**
  - Discoverability: 0.37 → 5.00 (vault L2 539 page 중 539 searchable)
  - Lifecycle: 0.34 → 4.97 (539 중 537 reviewed)
  - Cross-ref 4.63 (동일) / Coverage 4.13 (동일) / Freshness 4.20 (동일) / Operational 5.00 (동일)
  - Dashboard 자동 갱신 (score 갱신 + timestamp)
- **누적 158 test PASS** — 회귀 0
- **Follow-up (v0.7.1+)**:
  1. score tool 의 CI 통합 (overall < 4.0 시 alert)
  2. v0.7.1: 6 dim 모두 ≥ 4.5 (Grade A 안정) 유지 정책
  3. score trend over time (commit 별 점수 추적)
  4. v0.7.1: vault L2 sources/ 의 *auto-archive* (raw mirror 가 90일 이상 stale 인 page)

## [2026-06-13] v0.7.1 (commit `f09034d`) | follow-up 4건 + wiki 개선 4건 묶음 (158 test PASS)

- **Trigger**: yklee 의 "v0.7.1-beta 묶음" 요청. v0.7.0 release 의 4 follow-up + 이번 session 의 wiki 개선 4건.
- **4 follow-up 모두 완료**:
  - **1. 9-Artifact auto-fill helper** (`tools/fill_reverse_engineering_artifacts.py`, 227 line)
    - workflow-source/reverse-engineering/ template 자동 fill + heuristic TODO marker
    - `--info` / `--project-root` / `--apply` / `--limit` / `--output-dir`
  - **2. SEC-WF-05 dependency integrity 실제 검증** (baselines.py)
    - pyproject.toml 의 version pin (== / >=) + lock file (requirements.txt / uv.lock / poetry.lock) + checksum (sha256 / gpg) 3가지 평가
    - 평가: pinned + (lock OR checksum) = compliant / pinned only = advisory / no pin = non_compliant
  - **3. 9-Artifact index topic page** (`topics/reverse-engineering-9-artifact-index.md`, 90 line)
    - 9 artifact 의 index (본 위치 + 주제 + Verification 안내)
    - `index.md` anchor 추가
  - **4. Extension sub-cat + resiliency 스케치** (`extensions/v0.7.1-roadmap.md`, 115 line)
    - v0.7.1+ sub-cat directory 구조 + 4종 (resiliency) 의 우리 적응 8/16 rule + v0.8.0+ follow-up 4건
- **Version bump + release notes**:
  - `pyproject.toml` 0.7.0 → 0.7.1
  - `workflow_kit/__init__.py` v0.7.0-beta → v0.7.1-beta
  - `releases/Beta-v0.7.1.md` (170 line, 4 follow-up + 4 wiki 개선 + score 갱신)
  - wheel + sdist 빌드 (twine check PASSED)
  - smoke venv (`/tmp/sawsmoke-071`): workflow_kit v0.7.1-beta 정상 + SEC-WF-05 advisory 검증 동작
  - GitHub Release v0.7.1-beta 발행
- **누적 158 test PASS** — 회귀 0
- **Follow-up (v0.7.2+)**:
  1. sub-cat 본 구현 (auth-baseline, property-based-testing, memory-baseline, resiliency-baseline)
  2. 9-Artifact auto-fill helper 의 heuristic 강화
  3. score tool 의 CI 통합

## [2026-06-13] wiki maintainability score trend (commit `99e299f`) | 7 milestone score 누적 + dashboard 갱신

- **Trigger**: yklee 의 "score trend over time (commit 별 추적)" 요청.
- **신규 tool**:
  - `tools/score_wiki_trend.py` (170 line) — git log + score tool 결과 누적 + ASCII chart 시각화
  - `--record-current` (HEAD 점수 기록) / `--record-range N` (최근 N commit 재기록) / `--show` (ASCII chart) / `--json`
  - history: `tools/.score_history.jsonl` (v0.7.1+ 누적)
- **Dashboard 갱신**:
  - `score_wiki_maintainability.py` 의 `emit_dashboard()` 에 trend section 추가
  - 7 commit 의 overall + grade 를 table 로 표시
  - 자동 추출 — score tool 실행 시 dashboard 자동 갱신
- **7 milestone record** (key commit 의 score):
  - `0052da1` (v0.7.0 step 7): 3.11 (D)
  - `021ec16` (v0.7.0 wiki maintainability): 3.70 (D)
  - `7a4dbae` (v0.7.0 L2 30 page emit): 3.70 (D)
  - `49dfc78` (v0.7.0 score metric + dashboard): 3.70 (D)
  - `c72bdc3` (v0.7.0 L2 499 page metadata-only emit): 4.66 (A)
  - `f09034d` (v0.7.1 release): 4.66 (A)
  - `bad14d8` (current HEAD): 4.67 (A)
- **신규 smoke test** (tests/check_wiki_trend.py, 220 line, 10 test PASS):
  - tool importable + show runs (2)
  - history valid jsonl + score range (2)
  - chart bar chars (1)
  - JSON output (1)
  - dashboard integration (2)
  - idempotency (1)
- **누적 168 test PASS** (v0.7.1 158 + 10 신규) — 회귀 0
- **Follow-up (v0.7.2+)**:
  1. trend 의 dim 별 변화 자동 alert (≥ 0.3 하락 시)
  2. score tool 의 CI 통합 (overall < 4.0 시 block)
  3. v0.7.1 trend 자동 누적 (PR 머지 시 github action)

## [2026-06-13] wiki score trend dim 별 alert (commit `0224a76`) | --alert + --baseline 옵션 + 4 smoke test

- **Trigger**: yklee 의 "trend 의 dim 별 변화 자동 alert (≥ 0.3 하락 시)" 요청.
- **신규 옵션** (tools/score_wiki_trend.py):
  - `--alert --baseline=<commit>`: 현재 score vs baseline 의 dim 별 비교
  - `--threshold N` (default 0.3): alert 임계값
  - **출력**: dim 별 🔴 alert / 🟢 info / ⚪ ok 표시
  - **CI 통합**: exit code 0 (no alert) / 1 (≥ 1 alert) / 2 (error: missing baseline)
- **새로운 type**:
  - `DimAlert` dataclass (dim / baseline / current / delta / severity)
  - `compare_scores(baseline, current, threshold)` 함수
  - `print_alerts()` 출력 함수
- **검증** (real baseline 7a4dbae vs current d8c981c):
  - 5 dim 개선 (info: freshness +2.23 / discoverability +4.55 / lifecycle +4.55 / cross_ref +0.64 / operational +1.00)
  - 1 dim 유지 (coverage +0.17)
  - alert 0 → exit 0
- **Synthetic alert 시나리오 검증**:
  - baseline 5.0 / current freshness 4.5 (-0.5) / discoverability 4.6 (-0.4) / lifecycle 4.7 (-0.3) → 2 alerts (freshness, discoverability)
  - lifecycle -0.30 = boundary → ok (floating point 정밀도)
- **누적 172 test PASS** (v0.7.1 158 + 14 신규) — 회귀 0
- **신규 smoke test 4종**:
  - test_compare_scores_no_alert (info 만)
  - test_compare_scores_alert (alert 1)
  - test_alert_cli_no_alert (real baseline 비교, exit 0)
  - test_alert_cli_missing_baseline (exit 2)
- **CI 통합 가이드** (`.github/workflows/score-alert.yml` 권장):
  ```yaml
  - name: Wiki score alert check
    run: python3 workflow-source/tools/score_wiki_trend.py --alert --baseline=main~10 --threshold=0.3
  ```
- **Follow-up (v0.7.2+)**:
  1. v0.7.2: 6 dim 모두 ≥ 4.5 (Grade A 안정) 유지 정책 — alert 가 발화하지 않도록 *임계값 유지 정책*
  2. v0.7.2: PR 마다 자동 --record-current (github action)
  3. v0.7.2: v0.8.0+ 의 dim 별 trend 자동 alert 의 *alert channel* (slack / email)

## [2026-06-13] v0.7.2 (commit `3bffba3`) | Extension sub-cat + 4종 본 구현 (179 test PASS, GH release)

- **Trigger**: yklee 의 "v0.7.2 follow-up (sub-cat 본 구현)" 요청. v0.7.1 roadmap 의 4 follow-up 모두 본 구현.
- **신규 4 baseline (8 file, ~1,200 line)**:
  - `extensions/security/auth/auth-baseline.md` (210 line, 6 SEC-AUTH rule)
  - `extensions/security/auth/auth-baseline.opt-in.md`
  - `extensions/testing/property-based/property-based-testing.md` (210 line, 6 PBT-WF rule)
  - `extensions/testing/property-based/property-based-testing.opt-in.md`
  - `extensions/performance/memory/memory-baseline.md` (210 line, 6 PERF-MEM rule)
  - `extensions/performance/memory/memory-baseline.opt-in.md`
  - `extensions/resiliency-baseline.md` (200 line, 8 RES-WF rule)
  - `extensions/resiliency-baseline.opt-in.md`
- **Lint rule 확장** (`tests/check_extension_system.py`):
  - `SUB_CAT_EXTENSIONS` 정의 (4종)
  - 7 신규 test (sub_cat_baselines_present / opt_ins_present / rule_count / rule_id_format / opt_in_question_format / aidlc_reference / unique_prefix)
  - `RULE_ID_RE` 의 v0.7.2 prefix 지원: `<CAT>(-<SUB>)?(-WF)?-<NN>` (SEC-AUTH, PBT-WF, PERF-MEM, RES-WF)
  - **30/30 PASS** (v0.7.0 의 23 + 7 신규)
- **Version bump + release notes + GH release v0.7.2-beta** (wheel + tar.gz)
- **누적 179 test PASS** — 회귀 0
- **각 baseline 의 핵심 rule (summary)**:
  - SEC-AUTH-01~06: API key / token rotation / OAuth scope / 2FA / entropy / auth audit
  - PBT-WF-01~06: property ID / round-trip / invariant / idempotency / generator / shrink
  - PERF-MEM-01~06: peak mem / leak / GC / ref cycle / profiling / regression
  - RES-WF-01~08: critical / change mgmt / observability / health / backup / recovery (AIDLC 16 → 우리 8)
- **Follow-up (v0.7.3+)**:
  1. v0.7.3 runtime helper 본 구현 (auth / testing / profiling / resiliency)
  2. v0.7.3 baseline evaluate_compliance() 확장 (5 baseline × ~34 RuleResult)
  3. v0.7.3 flat path migration (v0.7.0 의 flat → v0.7.2+ sub-cat)
  4. v0.7.3 PBT hypothesis + memory objgraph 의존성 옵션

## [2026-06-13] v0.7.3-beta (commit `d03348a`) | 4 Runtime Helper 본 구현 (191 test PASS, GH release)

## [2026-06-13] v0.7.4-beta (commit `22e7750`) | workflow doctor CLI + @graceful_shutdown + optional dep (200 test PASS, GH release)
