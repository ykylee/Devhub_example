---
type: topic
status: active
last_ingested_from: workflow-source/core/global_workflow_standard.md + workflow-source/core/workflow_task_modes.md
related_pages: [concepts/project-architecture, concepts/agent-topology, concepts/orchestrator-subagent-pattern, decisions/adr-001-3-layer-separation, decisions/adr-004-wiki-layer, patterns/r4-anchor-index, topics/standard-ai-workflow-architecture-2026, topics/wiki-ingest-lifecycle]
created: 2026-06-12
updated: 2026-06-12
active_since: 2026-06-12
active_reason: "draft → active (commit 2916d49 + cross-channel 동기화 완료). V-1 (위치) / V-4 (anchor) / V-R9 (면제) / R-1 (inbound ≥ 1) 모두 PASS"
---

# AIDLC 벤치마크 분석 (2026-06-12)

## 1. 분석 목적 및 범위

AWS AIDLC(AI-DLC, AI-Driven Development Life Cycle) 워크플로우의 설계 철학과 메커니즘을 벤치마킹하여, 우리 `standard_ai_workflow_minimax` (v0.6.3-beta) 의 보완 후보를 도출한다.

- **레퍼런스**: `https://github.com/awslabs/aidlc-workflows` (commit `b19c819`, 2026-06-08)
- **체크아웃 위치**: `/Users/yklee/repos/aidlc-workflows/`
- **분석 일자**: 2026-06-12
- **분석자**: Mavis (yklee 의뢰)
- **분석 모드**: 풀 벤치마크 (yklee 명시)

## 2. AIDLC 핵심 구조 (요약)

### 2.1 3-Phase 라이프사이클

| Phase | 목적 | Stage (총 13개) | 핵심 |
|---|---|---|---|
| 🔵 INCEPTION | WHAT/WHY (계획·아키텍처) | 7개: Workspace Detection, Reverse Engineering, Requirements Analysis, User Stories, Workflow Planning, Application Design, Units Generation | 무조건 3개 (WD/RA/WP) + 조건부 4개 |
| 🟢 CONSTRUCTION | HOW (설계·구현·빌드) | 6개: Functional Design, NFR Requirements, NFR Design, Infrastructure Design, Code Generation, Build & Test | per-unit loop + build&test |
| 🟡 OPERATIONS | DEPLOY/RUN (placeholder) | 1개 (placeholder) | 미래 확장용 |

### 2.2 Adaptive Execution 메커니즘

**"The workflow adapts to the work, not the other way around."**

핵심 차별점:
- **Stage Selection (binary)**: EXECUTE vs SKIP. Workflow Planning stage 가 결정.
- **Detail Level (adaptive)**: Stage 가 실행되면 *모든* 정의된 artifact 가 생성됨. 단 detail level 이 complexity 에 따라 minimal/standard/comprehensive 로 조정.
- **No fixed sequence**: 13개 stage 가 의존성 그래프 기반으로 자유롭게 실행.

### 2.3 결정적 차별 메커니즘 7가지

1. **Workspace Detection (ALWAYS)**: `aidlc-state.md` 존재 여부로 resume/신규 판정. brownfield/greenfield 자동 분류.
2. **Reverse Engineering (CONDITIONAL, brownfield only)**: 기존 코드베이스 → 9개 artifact 자동 생성 (architecture, code-structure, api, component-inventory, tech-stack, dependencies, business-overview, code-quality, timestamp).
3. **Requirements Analysis (ALWAYS, adaptive depth)**: minimal/standard/comprehensive 3 depth. **결정적**: `requirement-verification-questions.md` 에 모든 질문을 파일로 작성, [Answer]: tag 로 응답.
4. **Workflow Planning (ALWAYS)**: Mermaid 다이어그램으로 실행 plan 시각화. 사용자 승인 게이트.
5. **Application Design (CONDITIONAL)**: 4개 artifact (components.md, component-methods.md, services.md, component-dependency.md).
6. **Units Generation (CONDITIONAL)**: 시스템을 unit of work 로 분해. `unit-of-work.md` / `unit-of-work-dependency.md` / `unit-of-work-story-map.md` 3종.
7. **Per-Unit Loop + Build & Test**: 각 unit 별로 design 4종 + code generation (Part 1 Planning, Part 2 Generation) 완료 후 다음 unit.

### 2.4 Audit & State 추적

- `aidlc-state.md`: stage progress 체크박스 + Extension Configuration
- `audit.md`: **append-only**. 모든 user input 을 ISO 8601 timestamp 와 함께 raw text 그대로 기록 (요약 ❌). critical rule: 파일 overwrite 절대 금지, append/edit 만 허용.
- `aidlc-docs/` 디렉토리 구조: inception/, construction/, operations/ 하위에 plans/, reverse-engineering/, requirements/, user-stories/, application-design/, build-and-test/ 등.

### 2.5 Extensions 시스템 (★ 핵심 차별점 ★)

`extensions/` 디렉토리에 3종 baseline extension:
- `security/baseline/` (SECURITY-01 ~ SECURITY-NN)
- `testing/property-based/` (PROPERTY-01 ~)
- `resiliency/baseline/`

**메커니즘**:
- 각 extension = `<name>.md` (rules) + `<name>.opt-in.md` (사용자 opt-in 질문)
- 워크플로우 시작 시 `*.opt-in.md` 만 load (가벼움). full rule 파일은 opt-in 후에만 load.
- **blocking constraint**: extension rule 은 hard constraint. 미준수 시 stage completion message 에 "Request Changes" 만 표시, "Continue" 옵션 자체가 사라짐.
- **N/A handling**: rule 이 현재 project 에 적용 안 되면 N/A 로 마킹, blocking 아님.
- **Extension Configuration**: `aidlc-state.md` 에 enable/disable 상태 기록.

### 2.6 Question Format (강제)

**`common/question-format-guide.md`** 가 다음을 mandatory 로 강제:
- ❌ 채팅으로 질문 ❌ — 반드시 별도 question 파일에 작성
- ✅ A/B/C/D/X multiple choice (X = Other)
- ✅ `[Answer]:` tag 로 응답
- ✅ 모든 옵션 사이에 빈 줄 (CommonMark strict renderer 호환)
- ✅ Contradiction/ambiguity detection 후 clarification file 자동 생성

### 2.7 Audit Log 정책 (강제)

- **Append-only** (절대 overwrite 금지)
- **Raw user input 그대로** 기록 (요약/의역 ❌)
- **ISO 8601 timestamp** (YYYY-MM-DDTHH:MM:SSZ)
- 모든 interaction 기록 (승인만 ❌)
- Stage context 명시

## 3. 우리 워크플로우 구조 (요약)

### 3.1 핵심 컴포넌트 (v0.6.3-beta)

| 영역 | 우리 | 비고 |
|---|---|---|
| Lifecycle | 없음 (무한 loop, 작업 단위) | INCEPTION/CONSTRUCTION 같은 phase 구분 ❌ |
| Task Mode | 6 mode (Analysis/Requirements/Design/Planning/Implementation/Refactoring) | `workflow_task_modes.md`, `workflow_kit/common/modes/registry.py` |
| State 추적 | `state.json` + `session_handoff.md` + `work_backlog.md` + `daily_backlog_*.md` | machine-readable + human-readable |
| Audit Log | `audit.md` 개념 있으나 raw input capture 미강제 | `wiki-prompt-log` 가 raw mirror 담당 (별도 layer) |
| Quality Gate | `workflow-linter` (52개 smoke test) | `tools/check_packaging.py` 포함 |
| Skill | 11종 (session-start, backlog-update, doc-sync, validation-plan, merge-doc-reconcile, workflow-linter, code-index-update, project-status-assessment, git-conflict-resolver, robust_patcher, automated-repro-scaffold) | v0.5.7+ 프로토타입 |
| MCP | jsonrpc-bridge (default, stable) + stdio-sdk (experimental) | read-only 기본 |
| Harness | 6종 (Codex, OpenCode, Gemini, Antigravity, MiniMax Code, pi-dev) | bootstrap 자동화 |
| Orchestration | orchestrator + sub-agent (doc/code/validation worker) contract v1 | Pydantic v2 schema, output_validator + choose_roles |
| Task 분해 | work_backlog.md 의 작업 항목 단위 | unit-of-work 같은 system-level decomposition ❌ |
| Brownfield 분석 | `repository_assessment.md` (bootstrap --adoption-mode existing) | 1회성, RE 처럼 9개 artifact 자동 생성 ❌ |
| Cross-cutting rule | 없음 (extension 시스템 ❌) | security/testing/resiliency baseline ❌ |

### 3.2 강점 (AIDLC 대비)

1. **한국어 + 다국어 친화**: §1.1 언어 원칙 명시. AIDLC 는 영어 강제.
2. **Harness 추상화**: 6개 harness 에 overlay 패턴으로 동일 룰 적용. AIDLC 는 6개 platform 별로 README 1쌍 (`AGENTS.md` + `.aidlc-rule-details/`).
3. **Contract v1 (orchestrator ↔ sub-agent)**: 외부 spec + Pydantic v2 enforcement. AIDLC 는 agent 간 contract 표준 ❌.
4. **Workflow Kit + Bootstrap Lib**: Python package 로서 실제로 import 해서 쓸 수 있는 모듈. AIDLC 는 순수 markdown 룰.
5. **MCP 통합**: read-only mcp transport 가 1급 시민. AIDLC 는 tool 통합 ❌.
6. **Packaging**: wheel + GitHub Releases 듀얼 채널, packaging smoke 자동화. AIDLC 는 zip 1종.
7. **Wiki 운영 통합**: R-1 ~ R-9 lifecycle 이 자체 SSOT. AIDLC 는 wiki 개념 자체가 없음.
8. **Memory 3-state lifecycle** (active/frozen/archive): R-8/R-10 freeze 정책. AIDLC 는 단일 `aidlc-state.md`.

### 3.3 약점 / 갭 (AIDLC 대비)

| # | AIDLC 기능 | 우리 상태 | 영향도 | 보완 권장 |
|---|---|---|---|---|
| 1 | **3-Phase Lifecycle** (Inception/Construction/Operations) | ❌ 없음. mode 6종이 horizontal 분류 | 중 | 부분 도입 (mode 위에 phase overlay) |
| 2 | **Adaptive Execution** (Stage Selection + Detail Level) | ❌ 없음. skill 이 무조건 실행 | 중 | 부분 도입 (skill 마다 EXECUTE/SKIP gate) |
| 3 | **Stage Gate** (명시적 user approval) | ⚠️ `validation-plan`, `merge-doc-reconcile` 가 유사 | 중 | 적용 권장 (artifact 별 명시적 approval prompt) |
| 4 | **Workspace Detection** (resume vs 신규 자동 판정) | ⚠️ `state.json` 존재로 유사하나 명시적 protocol ❌ | 중 | 도입 권장 (`aidlc-state.md` 유사한 `workflow_state.md` 도입) |
| 5 | **Reverse Engineering** (9개 artifact 자동 생성) | ❌ `repository_assessment.md` 1개만 | 상 | 도입 권장 (우리도 RE artifact 9종 정의) |
| 6 | **Unit of Work 분해** (unit-of-work.md, dependency matrix, story map) | ⚠️ `work_backlog.md` 가 task 단위 분해 담당 | 상 | 부분 도입 (system-level 분해 vs task-level 분리) |
| 7 | **Question File Format** (강제 + contradiction detection) | ❌ 없음. inline Q&A 만 | 상 | **도입 권장 (★★★)** — 효과 큰 패턴 |
| 8 | **Extension 시스템** (opt-in + blocking rule) | ❌ 없음 | 상 | **도입 권장 (★★★)** — cross-cutting rule 의 정공법 |
| 9 | **Append-only Audit Log** (raw input + ISO 8601) | ⚠️ `wiki-prompt-log` 가 raw mirror 만, raw 자체 ❌ | 중 | 부분 도입 (per-project audit.md 채널 정의) |
| 10 | **Workflow Plan 시각화** (Mermaid) | ❌ 없음. `workflow_task_modes.md` 만 정적 | 중 | 부분 도입 (per-session plan 자동 생성) |
| 11 | **Build & Test 별도 phase** (instructions 7종) | ⚠️ `validation-plan` 스킬 1개 | 중 | 부분 도입 (build/test/e2e/contract/security 5종 분화) |
| 12 | **Per-Unit Design Loop** (FD/NFR/NFRD/ID 4종) | ❌ 없음 | 하 | 비권장 (overhead > benefit for our scale) |
| 13 | **NFR Design 별도 stage** | ❌ 없음 (Design mode 일부 흡수) | 하 | 비권장 |
| 14 | **Operations phase** (deployment/monitoring) | ❌ 없음 | 중 | 별도 ADR 로 검토 (DevOps 영역 침범 여부) |
| 15 | **Property-based testing** baseline | ❌ 없음 | 하 | extension 후보 |
| 16 | **Security baseline** (SECURITY-01 ~ NN) | ❌ 없음 | 상 | **도입 권장 (★★★)** — 우리도 project profile 에 security profile 추가 |
| 17 | **Resiliency baseline** | ❌ 없음 | 중 | extension 후보 |
| 18 | **"No duplication"** 룰 (SSOT 강제) | ⚠️ ADR-001 (3-layer) + ADR-004 (wiki layer) 로 부분 충족 | - | 이미 좋음 |
| 19 | **"Agnostic"** 룰 (tool/vendor 독립) | ✅ 강점 | - | 이미 좋음 |
| 20 | **Welcome Message** (workflow start banner) | ❌ 없음 | 하 | 부분 도입 (entry point banner 표준화) |

## 4. 보완안 (도입 권장 / 부분 도입 / 비권장)

### 4.1 도입 권장 (★★★ 효과 큼, 즉시 검토)

#### A. Question File Format 패턴 (AIDLC § question-format-guide)
- **현재**: inline Q&A, 모호한 응답 처리 X
- **개선**: AIDLC 의 multi-choice + `[Answer]:` tag + clarification file 패턴 도입
- **적용 위치**:
  - `workflow_adoption_entrypoints.md` 에 question format 섹션 추가
  - `project_workflow_profile_template.md` 에 "프로젝트별 질문 파일" 가이드
  - `ask_user` 호출 빈도 줄이고, 사용자 응답 후 contradiction/ambiguity 자동 점검
- **Effort**: 1-2 session (문서 + 1-2 sample)
- **영향**: 큰 효과. yklee 가 multi-question popup 보다 file 기반을 선호할 가능성 있음 (cross-session grep, 위키 ingest 호환).

#### B. Extension 시스템 (cross-cutting rule 의 정공법)
- **현재**: cross-cutting rule (security/testing/perf) 위치 ❌
- **개선**: `workflow-source/extensions/` 디렉토리 + `<name>.md` + `<name>.opt-in.md` 페어
- **적용 위치**:
  - `workflow-source/extensions/security-baseline.md` (v0.7.0 신규)
  - `workflow-source/extensions/testing-baseline.md` (v0.7.0 신규)
  - `workflow-source/extensions/perf-baseline.md` (v0.7.0 신규)
- **Effort**: 3-5 session (초안 + R-9 검증 + wiki 연동 + 1-2 실전 pilot)
- **영향**: 매우 큰 효과. project 별로 cross-cutting rule 을 enable/disable 하는 표준 메커니즘 부재가 가장 큰 갭.

#### C. Stage Gate 명시화
- **현재**: skill output 후 implicit next-step. 명시적 approval prompt 없음.
- **개선**: 각 stage 완료 시 표준 2-option message (Request Changes / Continue)
- **Effort**: 1-2 session (전체 skill spec 11종 review + 표준화)
- **영향**: 중간. agent 행동 일관성 ↑, yklee 의 self-review 패턴과 정합.

#### D. Reverse Engineering 9-Artifact 자동 생성
- **현재**: `repository_assessment.md` 1개. 토폴로지/의존성/API 별도 생성 ❌
- **개선**: bootstrap 시 9종 artifact 자동 생성 (architecture, code-structure, api, component-inventory, tech-stack, dependencies, business-overview, code-quality, timestamp)
- **Effort**: 2-3 session (bootstrap_lib 확장 + smoke test 추가)
- **영향**: 큰 효과. brownfield project 온보딩 시 1회성 → 영구 SSOT.

### 4.2 부분 도입 (선택적, 모드별 흡수 검토)

#### E. 3-Phase Lifecycle (Inception/Construction/Operations) 부분 도입
- **현재**: 6 mode (horizontal)
- **개선**: mode 위에 phase overlay (vertical). Requirements/Design/Planning → INCEPTION. Implementation/Refactoring → CONSTRUCTION. Analysis 는 양쪽.
- **Effort**: 1 session (workflow_task_modes.md 업데이트 + sample)
- **영향**: 중간. "지금 어느 phase?" 명시성 ↑, Operations phase 는 별도 검토.

#### F. Adaptive Stage Execution (skill EXECUTE/SKIP gate)
- **현재**: skill 11종 모두 무조건 사용 가능. 자동 선택 로직 약함.
- **개선**: 모드별로 skill EXECUTE/SKIP 매트릭스. task 가 "bug fix" 면 design-planning 스킬 SKIP.
- **Effort**: 1-2 session (`modes/registry.py` 확장)
- **영향**: 큰 효과. 불필요한 stage 제거 → 컨텍스트 절약.

#### G. Unit of Work 분해 (system-level)
- **현재**: `work_backlog.md` 의 task 단위
- **개선**: work_backlog 위에 unit-of-work.md 추가. system → unit → task 3-layer.
- **Effort**: 1-2 session
- **영향**: 중간. 마이크로서비스/멀티 모듈 project 에서 가치 큼. 단일 모듈 project 에는 overhead.

#### H. Audit Log 표준화 (per-project audit.md)
- **현재**: `wiki-prompt-log` 가 L1 raw mirror 만. in-project audit 채널 ❌
- **개선**: `ai-workflow/memory/active/audit.md` 채널 정의. append-only, ISO 8601 timestamp, raw input 그대로.
- **Effort**: 1 session (정책 + 1 sample)
- **영향**: 중간. yklee 의 `wiki-prompt-log` 와 role 충돌 주의 — in-project 는 short-term, vault 는 long-term archive 로 분리.

#### I. Build & Test 분화 (5종 instructions)
- **현재**: `validation-plan` 1개
- **개선**: build / unit / integration / contract / e2e 5종 분화. 각 스킬 output schema 표준화.
- **Effort**: 2-3 session
- **영향**: 중간. contract v1 spec 의 build/test 관점 강화.

#### J. Workflow Plan Mermaid 자동 생성
- **현재**: `workflow_task_modes.md` 정적 문서
- **개선**: session-start 시 현재 task 의 plan 을 mermaid 로 자동 생성 + `state.json` 에 attach
- **Effort**: 1 session
- **영향**: 중간. 시각화 + 사용자 이해 ↑.

### 4.3 비권장 (오버헤드 > 효과, 우리 스케일 부적합)

#### K. Per-Unit Design Loop (FD/NFR/NFRD/ID 4종)
- 우리 사용 패턴 (단일 모듈 풀스택 generalist) 에서 4종 분화는 과잉. yklee 의 "1 session = 1 chapter" 패턴과 충돌.

#### L. NFR Design 별도 stage
- Design mode 에 흡수. 별도 stage 불필요.

#### M. Welcome Message banner
- overhead 대비 효과 적음. session-start 가 이미 비슷한 역할.

### 4.4 별도 결정 필요 (yklee ADR 요청)

#### N. Operations phase 도입 여부
- DevOps 영역 (deployment, monitoring, incident response) 까지 워크플로우로 다룰지.
- AWS AIDLC 는 placeholder. 우리도 동일하게 placeholder 로 두는 게 안전.
- **권장**: ADR 로 결정. 기본값 = placeholder (AIDLC 추종).

#### O. 9종 extension 사전 정의 vs 사용자 opt-in
- AIDLC 는 3종 (security/testing/resiliency) baseline + 사용자 opt-in.
- 우리는 baseline 도 없음. **권장**: v0.7.0 에서 1종 (security-baseline) 만 출시 + opt-in 패턴, 나머지는 v0.8+ 점진.

## 5. 권장 실행 순서 (v0.6.3 → v0.7.0 로드맵)

| 순서 | 작업 | 권장 버전 | Effort | 의존성 |
|---|---|---|---|---|
| 1 | Question File Format 패턴 (A) | v0.6.4 | 1-2 ses | 없음 |
| 2 | Stage Gate 명시화 (C) | v0.6.4 | 1-2 ses | A 완료 후 |
| 3 | 3-Phase Lifecycle 부분 도입 (E) | v0.6.5 | 1 ses | 없음 |
| 4 | Adaptive Stage Execution (F) | v0.6.5 | 1-2 ses | E 완료 후 |
| 5 | Workflow Plan Mermaid (J) | v0.6.5 | 1 ses | E 완료 후 |
| 6 | Reverse Engineering 9-Artifact (D) | v0.7.0 | 2-3 ses | 없음 |
| 7 | Extension 시스템 (B) | v0.7.0 | 3-5 ses | D 와 병행 가능 |
| 8 | Security-baseline extension (O) | v0.7.0 | 1 ses (B 와 함께) | B 완료 후 |
| 9 | Unit of Work 3-layer (G) | v0.7.1 | 1-2 ses | B 완료 후 |
| 10 | Audit Log 표준화 (H) | v0.7.1 | 1 ses | 없음 |
| 11 | Build & Test 5종 분화 (I) | v0.8.0 | 2-3 ses | contract v1 강화 |
| 12 | Operations phase ADR (N) | 별도 | 1 ses (ADR) | 없음 |

## 6. Cross-cutting 정합성 체크리스트 (우리 L1 wiki R-1~R9)

분석 결과 L1 wiki 에 추가/수정 시 영향받는 rule:
- **R-1 (위키 SSOT)**: 본 topic page 추가. status: draft → stable 전환 필요 (1 세션 후).
- **R-2 (vault 정합)**: 본 page → L2 `~/wiki/wiki/projects/standard-ai-workflow/` 의 derived view (L2 `topics/` 디렉토리 미존재 — `comparisons/` 또는 신규 디렉토리 결정 필요).
- **R-3 (pull-before-push)**: commit 직전 `git fetch && rebase origin/main` 필수.
- **R-7 (merge-res)**: 신규 page merge 시 깨진 wikilink 확인.
- **R-8 (freeze)**: 본 page 는 freeze 불필요 (status: draft, in-repo codebase self-ingest).
- **R-9 (raw source `archive/` only)**: 본 page 는 codebase self-ingest (R-9 면제 대상) — `last_ingested_from` 가 in-repo path (`workflow-source/core/*`). V-R9 lint skip 불필요.
- **V-1 (위치 단일성)**: `ai-workflow/wiki/topics/` 만 사용 ✅.
- **V-4 (index anchor)**: `index.md` 에 `### [[topics/aidlc-benchmark-analysis-2026-06-12]]` anchor 추가 필요.
- **V-2 (commit = 1 ingest)**: 단일 page 추가는 1 commit 으로.

## 7. 검증 방법 (analyst's evidence)

본 분석은 다음 파일들을 직접 읽고 1차 출처 기반으로 작성됨:
- `/Users/yklee/repos/aidlc-workflows/README.md` (962 lines)
- `/Users/yklee/repos/aidlc-workflows/AGENTS.md` (182 lines)
- `/Users/yklee/repos/aidlc-workflows/aidlc-rules/aws-aidlc-rules/core-workflow.md` (539 lines)
- `/Users/yklee/repos/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/inception/workspace-detection.md` (97 lines)
- `/Users/yklee/repos/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/inception/requirements-analysis.md` (190 lines)
- `/Users/yklee/repos/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/inception/units-generation.md` (188 lines)
- `/Users/yklee/repos/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/inception/workflow-planning.md` (469 lines, 1-120 read)
- `/Users/yklee/repos/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/construction/code-generation.md` (217 lines)
- `/Users/yklee/repos/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/common/process-overview.md` (141 lines)
- `/Users/yklee/repos/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/common/depth-levels.md` (73 lines)
- `/Users/yklee/repos/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/common/question-format-guide.md` (369 lines)
- `/Users/yklee/repos/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/extensions/security/baseline/security-baseline.md` (307 lines, 1-80 read)
- `/Users/yklee/repos/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/extensions/security/baseline/security-baseline.opt-in.md` (20 lines)
- `/Users/yklee/repos/aidlc-workflows/docs/GENERATED_DOCS_REFERENCE.md` (102 lines)

우리 측:
- `/Users/yklee/repos/standard_ai_workflow_minimax/README.md` (329 lines)
- `/Users/yklee/repos/standard_ai_workflow_minimax/QUICKSTART.md` (139 lines)
- `/Users/yklee/repos/standard_ai_workflow_minimax/workflow-source/core/global_workflow_standard.md` (149 lines)
- `/Users/yklee/repos/standard_ai_workflow_minimax/workflow-source/core/workflow_task_modes.md` (65 lines)

## 8. 다음 단계

1. **즉시 가능 (이번 session 마무리)**: 본 page L1 wiki commit + L2 mirror + raw mirror 동기화
2. **다음 session 후보**: v0.6.4 작업 시작 (Question File Format + Stage Gate)
3. **ADR 후보**: Operations phase 도입 여부 (yklee 별도 요청 시)
4. **Extension 1차**: security-baseline 초안 작성 (별도 1-2 session)

## 9. v0.6.4-6 follow-up 실적 (2026-06-12, 17 commit, ~3,800 line, 35 test PASS, 0 breaking change)

### 9.1 적용된 보완안 (§4 도입 권장)

| 보완안 | 상태 | commit | line | test |
|---|---|---|---|---|
| **A. Question File Format** (★★★) | ✅ 완료 | 25756bb, bc16d91 | 562 (spec) + 358 (code) + 336 (test) | 7 PASS |
| **B. Extension 시스템** (★★★) | ⏸ v0.7.0+ 후보 | (예정) | - | - |
| **C. Stage Gate 명시화** (★★★) | ✅ 완료 | 25756bb, bc16d91, dd98e69, 2fab835, ca7a685, 6a9126c | 552 (spec) + 335+186 (code) + 318+292 (test) | 28 PASS |
| **D. Reverse Engineering 9-Artifact** (★★★) | ⏸ v0.7.0+ 후보 | (예정) | - | - |

### 9.2 부분 도입 (selected)

- **E. 3-Phase Lifecycle overlay**: mode 6종 위 phase overlay ⏸ v0.7.0+ (필요 시)
- **F. Adaptive Stage Execution**: skill EXECUTE/SKIP ⏸ v0.7.0+ (mode × skill matrix 필요)
- **G. Unit of Work 3-layer**: system-level 분해 ⏸ v0.7.1+
- **H. Audit Log 표준화**: stage_gate.append_audit_log 로 partial ✅
- **I. Build & Test 5종 분화**: ⏸ v0.8.0+ (contract v1 강화 후)
- **J. Workflow Plan Mermaid 자동 생성**: ⏸ v0.6.5+ (별도)

### 9.3 ADR 후보

- **N. Operations phase 도입 여부**: ⏸ yklee 별도 결정 (placeholder 유지)
- **O. Extension 1차 출시 범위**: ⏸ v0.7.0+ (security-baseline 1종만, opt-in 패턴)

### 9.4 비권장 (오버헤드 > 효과)

- **K. Per-Unit Design Loop**: 우리 스케일 부적합
- **L. NFR Design 별도 stage**: Design mode 흡수
- **M. Welcome Message banner**: session-start 가 유사 역할

### 9.5 v0.6.5 release + v0.6.6 follow-up + v0.6.7 packaging

- **v0.6.5** (commit 3897da7): Beta-v0.6.5.md release note + version bump v0.6.3-beta → v0.6.5-beta. 10 commit 묶음.
- **v0.6.6 follow-up #1** (commit 6a9126c): 5 SKILL.md-only skill 의 runtime 적용. 12/12 일관성 (workers 의 5 agent roster 는 design-excluded).
- **v0.6.7 packaging** (예정, 본 session): export_harness_package.py 실행 → dist/harnesses/<harness>/v0.6.6-beta/ + GitHub release draft.

### 9.6 §5 권장 실행 순서 진행 현황

| Step | v0.6.x | Effort | Status |
|---|---|---|---|
| 1. Question Format (A) | v0.6.4 | 1-2 ses | ✅ commit 25756bb, bc16d91 |
| 2. Stage Gate (C) | v0.6.4 | 1-2 ses | ✅ commit 25756bb, bc16d69 |
| 3. 3-Phase overlay (E) | v0.6.5 | 1 ses | ⏸ optional |
| 4. Adaptive Stage (F) | v0.6.5 | 1-2 ses | ⏸ optional |
| 5. Mermaid plan (J) | v0.6.5 | 1 ses | ⏸ optional |
| 6. Reverse Engineering (D) | v0.7.0 | 2-3 ses | ⏸ follow-up |
| 7. Extension (B) | v0.7.0 | 3-5 ses | ⏸ follow-up |
| 8. Security-baseline (O) | v0.7.0 | 1 ses | ⏸ follow-up |
| 9. Unit of Work (G) | v0.7.1 | 1-2 ses | ⏸ follow-up |
| 10. Audit Log (H) | v0.7.1 | 1 ses | ⏸ follow-up |
| 11. Build/Test 5종 (I) | v0.8.0 | 2-3 ses | ⏸ follow-up |
| 12. Operations ADR (N) | 별도 | 1 ses | ⏸ yklee 결정 |
