# MiniMax Code (Mavis) Workflow Setup

- 문서 목적: 본 저장소에서 Mavis (MiniMax Code, 모델 `MiniMax-M3`) 가 작업을 운영할 때 따르는 단일 진입점. 기존 5-워커 워크플로우 (Claude/Codex/Gemini/Reasonix/OpenCode, [`docs/governance/worker_division.md`](../docs/governance/worker_division.md)) 와 공존하며, **Mavis 가 본 저장소 안에서 어떻게 움직이는지** 만 정의한다.
- 범위: 세션 모델, 작업 라우팅, 메모리 3-layer, communication, cron, 사용 가능 도구 (skill/agent/MCP), 브랜치 컨벤션, 인계 SOP, 하드 룰, **v0.7.17 vendor SSOT 동기화 (global_workflow_standard + orchestrator_subagent_contract_v1 + workflow_task_modes)**.
- 대상 독자: Mavis (root/branch 세션), 본 저장소에서 Mavis 를 호출하는 사용자, 후속 AI 에이전트.
- 상태: **active** (v0.7.17 vendor 동기화 완료, 2026-06-15)
- 최종 수정일: 2026-06-15 (v0.7.17 vendor SSOT 동기화 — §1.1 언어 원칙 + §1.2 컨텍스트 절약 + §1.3 작업 모드 6종 + §4 SSOT (maturity_matrix) + §5 sub-agent contract v1 + §7 day-1 baseline 의 vendor global_workflow_standard §1-§9 매핑)
- 결정 근거 sprint: `feat/sync-minimax-workflow-v0.7.17` (본 갱신)
- 1차 출처: [`vendor/standard_ai_workflow/harnesses/minimax-code/README.md`](../../vendor/standard_ai_workflow/harnesses/minimax-code/README.md), [`vendor/standard_ai_workflow/core/global_workflow_standard.md`](../../vendor/standard_ai_workflow/core/global_workflow_standard.md), [`vendor/standard_ai_workflow/core/orchestrator_subagent_contract_v1.md`](../../vendor/standard_ai_workflow/core/orchestrator_subagent_contract_v1.md), [`vendor/standard_ai_workflow/core/workflow_task_modes.md`](../../vendor/standard_ai_workflow/core/workflow_task_modes.md).
- 관련 문서: [`README.md`](./README.md), [`MEMORY_GOVERNANCE.md`](./MEMORY_GOVERNANCE.md), [`memory/PROJECT_PROFILE.md`](./memory/PROJECT_PROFILE.md), [`../AGENTS.md`](../AGENTS.md), [`../docs/governance/worker_division.md`](../docs/governance/worker_division.md).

## 0. 핵심 정의

| 항목 | 값 |
| --- | --- |
| 런타임 | MiniMax Code (coding agent / agentic coding workspace) |
| 에이전트 식별자 (display) | **Mavis** |
| 에이전트 ID (CLI/routing/storage) | `mavis` |
| 에이전트 타입 | `orchestrator` (root session) |
| 기본 모델 | `MiniMax-M3` |
| 세션 ID 형식 | `mvs_<uuid>` |
| 워크스페이스 기본값 | OS 환경 `MAVIS_WORKSPACE` / 사용자 선택 디렉터리 |
| 코드 ID prefix (브랜치) | `mvs/<work-name>` |
| sub-agent 역할 4종 (vendor contract v1) | `doc-worker` / `code-worker` / `validation-worker` / `workflow-worker` (임시) |
| 작업 모드 6종 (vendor §1.3) | Analysis / Requirements / Design / Planning / Implementation / Refactoring |
| SSOT 1차 출처 (vendor §7.1) | `vendor/standard_ai_workflow/core/maturity_matrix.json` |

## 1. 세션 모델 (Session Model)

Mavis 는 **root session ↔ branch session** 트리로 동작한다.

- **root session**: 사용자가 처음 진입한 세션. `agentRole: orchestrator`. 프로젝트 전체 컨텍스트 + 사용자 요청을 받아 작업 라우팅 결정 + 결과 통합. 절대 종료되지 않고 다음 사용자 요청을 대기 (idle = `finished` 라고 부르며 routable 한 상태).
- **branch session**: root 가 spawn 한 작업 단위 세션. 부모 root 의 워크스페이스를 상속 + cross-session communication 으로 보고. 일반적으로 작업 완료 후 close.
- **hard rule**: branch 세션에서 다시 branch mavis 세션을 spawn 하지 않는다 (사용자 명시 요청 시 예외).

본 저장소에서는:

- **opencode 워커의 branch 메모리** (`ai-workflow/memory/opencode/<branch>/`) 와 **Mavis branch session 메모리** (`$MAVIS_SCRATCHPAD` = `~/.mavis/scratchpads/<rootSessionId>/scratchpad.md`) 가 별개 레이어.
- Mavis 가 작성하는 cross-session 노트는 scratchpad, 본 저장소의 workflow state 는 opencode branch memory. 두 위치는 중복하지 않고 역할 분리.

## 1.1 언어와 보고 원칙 (vendor §1.1, 2026-06-09)

- 사용자에게 직접 보여지는 작업 보고, 상태 요약, 문서 초안, handoff, backlog 갱신 문안은 기본적으로 **한국어**로 작성한다.
- 저장소 표준 문서와 템플릿도 별도 예외가 없으면 한국어를 기본 언어로 유지한다.
- 코드, 명령어, 파일 경로, 설정 key, 외부 시스템 고유 명칭은 필요할 때 원문 그대로 유지할 수 있다.
- 프로젝트 특성상 영어 산출물이 꼭 필요한 경우에는 [`PROJECT_PROFILE.md`](./memory/PROJECT_PROFILE.md) 에 예외를 명시한다.

## 1.2 컨텍스트 절약 원칙 (vendor §1.2, v0.7.17)

- 사용자에게 보이지 않는 내부 처리, 중간 분류, 임시 사고 과정은 모델이 가장 효율적인 형태로 수행한다.
- 중간 reasoning 을 장문으로 반복 출력하지 않는다.
- 이미 확인한 사실을 매 단계 길게 재서술하지 않고, 필요한 결론과 다음 행동만 짧게 남긴다.
- 작업 중 누적되는 컨텍스트는 현재 의사결정과 다음 행동에 필요한 정보 중심으로 유지한다.
- 긴 원문 인용, 중복 요약, 불필요한 체크리스트 복제를 피한다.
- 세션 문서에는 최종 결정, 검증 결과, 다음 세션에 필요한 사실만 남기고 내부 탐색 흔적은 최소화한다.
- orchestrator 와 worker 를 나눠 운영할 수 있는 하네스에서는 **메인 orchestrator 가 직접 도구 호출을 떠안기보다 task delegation 과 결과 통합에 집중**하는 구성을 기본값으로 둔다.
- 실제 탐색, 수정, 검증은 bounded scope worker 에 맡기고, ask 는 genuinely blocking decision 이나 위험한 외부 작업으로만 좁히는 편을 기본 원칙으로 둔다.
- `ai-workflow/` 는 세션 복원과 workflow 상태 관리용 메타 레이어로 보고, 프로젝트 코드/문서 탐색 범위에는 기본적으로 포함하지 않는다.
- 메인 orchestrator 와 sub-agent 간 위임은 [vendor §5 sub-agent contract v1](./minimax_code_workflow.md#5-sub-agent-위임-contract-v1-vendor-orchestrator_subagent_contract_v1md-동기화) 의 외부 contract v1 을 따른다 (v0.5.4 부터 적용).

## 1.3 작업 모드 (Task Modes, vendor §1.3, 6종)

작업의 성격에 따라 최적화된 워크플로우를 제공한다. 모드별 에이전트 토폴로지 + 추천 skill 은 vendor [`workflow_task_modes.md`](../../vendor/standard_ai_workflow/core/workflow_task_modes.md) §3 의 SSOT.

| 모드 | 목적 | 주요 산출물 | 핵심 sub-agent | 본 저장소 매핑 |
| --- | --- | --- | --- | --- |
| **Analysis** | 코드베이스 구조/의존성/로직 파악 | `repository_assessment.md`, `code_index` | `doc-worker` (다수) | 본 저장소 `ai-workflow/skills/code-index-update/` + `mavis-team` 의 `explore` sub-agent |
| **Requirements** | 사용자 니즈 정의/제약 | `requirements.md`, `stakeholder_list` | `session-orchestrator` | 본 저장소 `mavis-team` 의 `general` sub-agent + `ask_user` |
| **Design** | 신규 기능 청사진/아키텍처 | `technical_spec.md`, `architecture.md` | `main`급 `doc-worker` | 본 저장소 `docs/governance/adr/` + `docs/architecture.md` + `mavis-team` 의 main-tier `general` |
| **Planning** | 실행 가능한 태스크로 분해 | `implementation_plan.md`, `backlog` | `backlog-steward` | 본 저장소 `ai-workflow/skills/backlog-update/` + `ai-workflow/memory/feat/*/work_backlog.md` |
| **Implementation** | 실제 코드 작성/단위 테스트 | code patches, `test_results` | `code-worker`, `validation-worker` | 본 저장소 `mavis-team` 의 `mini-coder-max` + `general` + `validation-plan` skill |
| **Refactoring** | 기능 유지/코드 품질 개선 | refactoring diffs, regression report | `code-worker`, `validation-worker` | 본 저장소 `mavis-team` 의 `mini-coder-max` + `merge-doc-reconcile` skill |

운영 원칙:

- 세션 오케스트레이터는 현재 작업의 성격을 판단하여 모드를 전환하고, 해당 모드에 최적화된 에이전트 토폴로지를 구성한다.
- 모드 전환은 명시적 (사용자 backlog 등록 시) + 암묵적 (orchestrator 의 자동 판단) 둘 다 가능. 모드는 `state.json` 또는 `session_handoff.md` 의 "현재 작업 모드" 섹션에 기록.

## 2. 작업 라우팅 (Task Routing)

사용자 요청이 들어오면 Mavis 는 다음 분기로 즉시 결정한다.

| 조건 | 결정 | vendor 매핑 |
| --- | --- | --- |
| 대화/Q&A/추천, 단순 정보 조회, 단일 파일 read, lightweight op | **Handle it yourself** — 직접 처리, branch spawn 안 함 | orchestrator 직접 도구 호출 (소규모) |
| 컨텍스트 안에 deliverable 을 끝까지 그릴 수 있는 bounded scope 작업 (단일 파일 fix, bulk rename, 설정/문서/프롬프트 편집, quick draft) | **Handle it yourself** — `mavis-team` 로드 금지 | orchestrator 직접 (bounded scope) |
| 3+ 독립 tracks, 멀티 소스/툴, high error cost, 멀티 스테이지 delivery chain | **`mavis-team` plan** — skill 로드 후 plan 실행 | orchestrator → sub-agent (doc/code/validation 분업, vendor contract v1) |
| 기존 deliverable 의 review/test/verify/audit (코드 리뷰, 테스트, 검증) | **Spawn single-shot worker** (`mavis communication send --command spawn`) — verifier 전용 채널 | orchestrator → validation-worker (vendor role) |
| Producer 작업 (코드/리팩토링/feature/bug fix) 을 single-spawn 으로 처리하고 싶을 때 | **금지** — 본인 또는 `mavis-team` 으로 라우팅 | vendor §3 role 경계 정합 |
| Analysis 모드 (코드베이스 구조 분석) | `mavis-team` plan + sub-agent 5+ doc-worker 병렬 | vendor §1.3 Analysis 모드 |
| Implementation 모드 (코드 작성/단위 테스트) | `mavis-team` plan + code-worker + validation-worker fan-out | vendor §1.3 Implementation 모드 + §4.2 멀티 컴포넌트 fan-out |

분기 결정은 작업 시작 전 1 회. 작업 중 복잡도가 escalate 되면 새 분기로 re-route 가능. 모드 전환 시 vendor §1.3 의 mode switching 절차 (state.json 기록) 적용.

## 3. 메모리 모델 (3-Layer Memory)

Mavis 의 메모리는 3 layer, 가장 좁은 것부터 선택:

1. **Project memory** (`AGENTS.md` 또는 repo 내 토픽 파일 + `changelogs/`) — 이 저장소/프로젝트에서만 유효. 직접 edit, no CLI.
2. **Agent memory** (`mavis memory append mavis --content '...'` ) — 다른 프로젝트에서도 같은 Mavis 로 일할 때 유효.
3. **User memory** (`mavis memory append --user --reason '<cross-project justification>' --content '...'` ) — `--reason` 필수, 모든 프로젝트에서 유효할 때만.
4. **Scratchpad** (`$MAVIS_SCRATCHPAD` = `~/.mavis/scratchpads/<rootSessionId>/scratchpad.md`) — cross-session whiteboard, branch 세션이 자동 상속. git untracked.

규칙:

- append 는 새 entry 만 추가. 수정/삭제 는 memory 파일 직접 edit.
- 작업 종료 전 "Did I learn anything reusable?" 1 회 — 해당되면 즉시 기록.
- **언어**: 사용자와 같은 언어 (이 저장소는 한국어, §1.1 vendor SSOT 정합). 코드/명령어/경로/외부 시스템 명칭은 원문 유지.
- 메모리는 hint — 액션 전 verify. 프로젝트 메모리는 git tracked, agent/user 메모리는 `~/.mavis/memory/` 아래.

본 저장소에서는:

- **Project memory (layer 1)**: `AGENTS.md`, `ai-workflow/MEMORY_GOVERNANCE.md` (정책), `ai-workflow/minimax_code_workflow.md` (본 문서), `ai-workflow/memory/PROJECT_PROFILE.md` (프로젝트 프로파일), `ai-workflow/memory/<agent>/<branch>/` (외부 5-worker + mavis branch state, **layer 1 의 일부**), 도메인별 `requirements.md` / `architecture.md`.
- **Agent memory (layer 2)**: `~/.mavis/memory/agents/mavis.md`. Mavis 가 다른 repo 에서도 활용할 cross-project 운영 노트.
- **User memory (layer 3)**: `~/.mavis/memory/user.md`. 사용자의 역할/관심사/워크플로우 선호 (자연스러운 대화로 채움).
- **Scratchpad**: 본 root session ID `mvs_8952f7f57f9749a68171434a78f89960` 의 scratchpad 가 branch 세션과 공유.

**구체적 session-start read / session-end write 절차** + **중복 방지 규칙** + **외부 worker branch memory 와의 경계** 는 [`MEMORY_GOVERNANCE.md` §0.5](./MEMORY_GOVERNANCE.md) 의 단일 source-of-truth 를 따른다. 본 섹션은 그것의 *요약* 일 뿐.

## 4. SSOT (Single Source of Truth) + 상태 동기화 (vendor §7, v0.5.10-beta)

v0.7.17 vendor 의 [`global_workflow_standard.md` §7](https://github.com/ykylee/standard_ai_workflow/blob/v0.7.17/core/global_workflow_standard.md#7) 의 SSOT 정책 정합.

### 4.1 단일 진실 공급원 (SSOT, vendor §7.1)

- 모든 skill, MCP, milestone 의 공식 상태는 **SSOT 1차 출처 = `vendor/standard_ai_workflow/core/maturity_matrix.json`** 에서 관리.
- 본 저장소 의 [`PROJECT_PROFILE.md`](./memory/PROJECT_PROFILE.md) + [`ai-workflow/MEMORY_GOVERNANCE.md`](./MEMORY_GOVERNANCE.md) 가 *본 저장소 측* 의 SSOT (vendor maturity_matrix 와 별개 레이어).
- v0.7.17 vendor 동기화 시 vendor 의 skill/MCP/milestone 변경분 → 우리 PROJECT_PROFILE 의 cross-reference 자동 갱신.

### 4.2 동기화 루틴 (vendor §7.2)

- **Skill 승급 시**: 코드 구현 완료 후 우리 DevHub 의 skill 의 stage 변경 → PROJECT_PROFILE.md 의 skill catalog 갱신.
- **TASK 완료 시**: 세션 종료 전, 완료된 TASK 가 본 저장소 의 milestone/skill 상태에 영향 주는지 확인 → PROJECT_PROFILE.md + memory 의 work_backlog.md + state.json 동시 갱신.

### 4.3 자동 검증 (Workflow Linter, vendor §7.3)

- 본 저장소 의 [`ai-workflow/skills/workflow-linter/`](./skills/workflow-linter/) 가 v0.7.17 vendor 의 workflow-linter 와 정합.
- 검증 항목: (1) 선언된 `test_path` 파일의 실제 존재, (2) 구현 완료(`stable`/`beta`)로 선언된 항목의 실제 코드/스크립트 존재, (3) 본 저장소 의 로드맵 단계와 PROJECT_PROFILE 의 milestone 단계 일치.
- 본 저장소 의 workflow-linter 실행: `bash ai-workflow/skills/workflow-linter/scripts/run_lint.sh` (또는 v0.7.17 vendor 의 `python3 vendor/standard_ai_workflow/tests/check_*.py` 5종 smoke 회귀 — §4.4).

### 4.4 v0.7.17 vendor smoke 회귀 (본 저장소 측 정합)

| Smoke test | PASS 기준 | 본 저장소 측 |
| --- | --- | --- |
| `check_v0_7_17_wiki_in_repo_isolation.py` (vendor) | 11/11 PASS | 외부 vault `~/wiki/` reference 0 |
| `check_v0_7_17_devhub_wiki_in_repo_invariant.py` (본 저장소) | 5/5 PASS | 본 저장소 script/doc 의 in-repo path 정합 |

본 smoke 들은 매 vendor 갱신 시 회귀 검증 필수 (`bash scripts/check_vendor_smoke.sh` 1회 실행).

## 5. sub-agent 위임 contract v1 (vendor `orchestrator_subagent_contract_v1.md` 동기화)

본 저장소 의 `mavis-team` plan + single-spawn verifier 의 sub-agent 위임은 vendor 의 **contract v1** 을 따른다. v0.5.4 부터 적용, v0.5.7 부터 멀티 컴포넌트 fan-out.

### 5.1 4개 역할 (vendor §3, 본 저장소 매핑)

| Role (vendor) | 책임 (요약) | 본 저장소 매핑 |
| --- | --- | --- |
| **orchestrator** (메인) | 우선순위 결정, 결과 통합, 사용자 보고, 위험 조정, ask_user 호출 | Mavis root session (§1) |
| **doc-worker** | 대량 문서 read, 문서 비교, 허브/상태 문서 초안 작성 | `mavis-team` 의 `general` sub-agent (small model, read-heavy) |
| **code-worker** | 범위 명확 구현, 코드/설정 수정, 빌드/컴파일 확인, 좁은 리팩터 | `mavis-team` 의 `mini-coder-max` (bounded edit) |
| **validation-worker** | 테스트 실행, 로그 확인, 검증 증빙 수집, 실패 원인 요약 | `single-spawn` verifier (`mavis communication send --command spawn --content '{"agent": "tester", ...}'`) |
| **workflow-worker** (임시/예외) | 위 3개로 분류 어려운 bounded task | `mavis-team` 의 `general` sub-agent (fallback) |

### 5.2 위임 입력 스키마 (vendor §4, 본 저장소 prompt 매핑)

`mavis-team` plan 의 subagent prompt 6섹션 표준 (TASK/EXPECTED OUTCOME/REQUIRED TOOLS/MUST DO/MUST NOT DO/CONTEXT) 은 vendor 의 위임 입력 스키마와 1:1 매핑:

| mavis-team 섹션 | vendor contract v1 필드 | 비고 |
| --- | --- | --- |
| TASK | `task.brief` + `task.constraints` | 1~2 문장 brief + scope 제한 |
| EXPECTED OUTCOME | `task.expected_outputs` (primary_artifact + artifact_kind + must_include) | 주 산출물 + 종류 + 필수 포함 항목 |
| REQUIRED TOOLS | `task.inputs.files` + `task.inputs.context_paths` | 읽어야 할 file + 참고 dir |
| MUST DO | `task.must_include` 또는 sub-task 의 `must_include` | mandatory 체크 |
| MUST NOT DO | `task.constraints.do_not_touch` | scope 외 금지 |
| CONTEXT | `context.branch` + `context.memory_layer_root` + `context.project_root` | git context |

`delegation_id`, `issued_at`, `issued_by.session_id`, `task.task_id`, `task.task_type` (`doc_draft` | `code_change` | `validation_run` | `bounded_research`), `validation.required` + `validation.criteria` + `validation.owner` 는 본 저장소 의 mavis-team engine 이 자동 생성. `task.required_model_tier` (`small` | `main`) 도 engine 자동 결정 (main 후보: 아키텍처 변경, 정책 결정, 5+ 파일 cross-cutting, bounded_research 5+ source).

### 5.3 멀티 컴포넌트 fan-out (vendor §4.2, v0.5.7)

`mavis-team` plan 의 3+ tracks 가 fan-out 결정 시 vendor §4.2 의 sub-task 패턴 적용. 부모 task 의 `sub_tasks` 필드에 N개 sub-task 정의, 각 sub-task 는 §4 의 task 본질 + `parent_delegation_id` + `sub_id` (`st-N`).

### 5.4 위임 출력 스키마 (vendor §5, 본 저장소 report 매핑)

sub-agent 가 root 에 보고할 때 vendor §5 의 출력 스키마 (`contract_version` + `delegation_id` + `completed_at` + `worker.session_id` + `worker.role` + `worker.model_tier` + `result.status` + `result.summary` + `result.artifact_paths` + `result.metrics` + `issues` + `verifications`) 정합. mavis-team 의 CycleReport 가 본 스키마를 따름.

## 6. Communication + Cron

### 6.1 Cross-session communication

`mavis communication send` API:

- root → branch, branch → root 보고, branch ↔ branch 모두 지원.
- report-back 실패 시: 5 초 후 1 회 재시도, 실패 시 scratchpad + IM notify.
- 모든 branch → root 보고는 vendor §5 의 위임 출력 스키마 정합.

### 6.2 Self-reminder (cron)

비동기 핸드오프 (CI, batch job, MR auto-merge, 외부 API, human reply 대기) 후 **반드시** cron 등록:

```bash
mavis cron self <name> --every <interval> --prompt "<text>"
```

예외: `mavis team plan` 자체는 heartbeat + CycleReport + unresponsive alert 가 내장되어 cron 등록 불필요.

### 6.3 외부 communication

- Feishu/Lark: `lark-tools` skill (또는 `lark-*` 세부 skill).
- X/Twitter: `x-link-reader` skill.
- IM 채널 라우팅: `mavis im` 로 hook/agent 매핑 관리.

## 7. 사용 가능 도구 (Tools)

### 7.1 Skills (직접 호출)

- `mavis` — Mavis runtime entry point (skill/agent/session/cron/hook 관리).
- `mavis-team` — 병렬 팀 plan 실행 또는 신규 agent 생성 라우팅. **3+ 독립 track / multi-source / high error cost 일 때만**.
- `mavis-doctor` — 세션/agent/daemon 디버깅. 키워드: 排查, 调试, 卡住, why, log, debug, inspect, retry, recovery.
- `deep-research` — 5-step Deep Research pipeline (Mavis Team Engine 기반).
- `init` — coding project bootstrap (root `AGENTS.md` + `.harness/`). cold-start 또는 사용자 명시 호출 시.
- `create-agent` — 새 agent 디스크 생성 (mavis-team 이 plan 분석 후 호출, 또는 사용자 명시).
- `mcp-cli`, `mcp-onboarding` — MCP 관리/온보딩.
- `llm-call` — 설정된 LLM 직접 호출.
- `plan-mode` — 실행 전 계획 (모호성 / 다중 옵션 / 사용자 명시).
- `skill-creator` / `skill-refiner` / `skill-evolution` — skill lifecycle.
- `visual-page` — diagram/data/webpage 시각화.
- `docx` / `pdf` / `pptx` / `xlsx` — 문서 deliverable.
- `lark-tools` (+ `lark-*` 세부) — Feishu/Lark full capability.
- `x-link-reader` — X/Twitter 우회 fetch.
- `playwright` — 브라우저 자동화 (dev/test/CI 자동화 시).

### 7.2 Sub-agents (직접 호출 가능한 native agent)

- `explore` — fast code exploration, glob/grep/Read 위주.
- `general` — 범용 multi-step worker.
- (vendor 의 doc-worker / code-worker / validation-worker / workflow-worker 4 role 은 **mavis-team 내부 분업** 으로 처리, 직접 호출 가능한 native agent 가 아님 — §5.1 매핑 참조).

### 7.3 MCP servers

`mavis mcp ls` 로 동적 확인. 본 저장소에서 자주 쓰는 것:

- `matrix` — VIDEO/AUDIO 이해, IMAGE/VIDEO/오디오 생성, **WEB_SEARCH** (네이티브 tool, `web_search` 직접 호출).
- `cu` (Computer Use) — 데스크탑 마우스/키보드 자동화. 25개 desktop_* 도구. 좌표는 0~1000 정규화.
- `playwright` — 브라우저 자동화. 공개 페이지 → `mavis-browser` 우선, SPA/anti-bot → playwright MCP.
- `trash` — recoverable delete (OS Trash). `mavis-trash` CLI 직접 호출.

### 7.4 Hard limits

- **single-spawn 은 verifier 전용** (vendor §3 validation-worker 정합). producer work 를 single-spawn 으로 처리하지 말 것.
- **`mavis-team` 을 low-complexity 에 로드 금지** — 한 번에 끝낼 수 있는 일에 팀 plan 금지.
- **사용자에게 clarify 하지 말 것** — 답을 모를 때만, 그리고 정말 outcome 이 달라질 때만 물어본다.
- **in-scope collateral issue 는 묶어서 fix** — 작은 부수 발견이라도 사용자에게 되묻지 말고 같이 fix.
- **verification 없는 결과는 done 으로 확정 금지** — deliverable 후 verify 단계 동행.
- **`mavis communication send --command prompt`** 를 기존 worker session 에 producer 작업 단축으로 쓰지 말 것 (mavis-team 또는 본인).
- **sub-agent 는 본인 도구 호출 회피** (vendor §1.2 정합) — bounded scope 위임 + 결과 통합에 집중.

## 8. 브랜치 / 커밋 / PR 컨벤션

### 8.1 브랜치 prefix

| prefix | 의미 | 비고 |
| --- | --- | --- |
| `mvs/` | Mavis root session 의 직접 작업 | 본 워커 prefix. 형식: `mvs/work_<YYMMDD>-<sprint-seq>-<short-key>` |
| `claude/` `codex/` `gemini/` `deepseek/` `opencode/` | 외부 AI 워커 | [`docs/governance/worker_division.md` §2.5](../docs/governance/worker_division.md) 정합 |
| `feat/` `fix/` `chore/` `docs/` `test/` | 일반 prefix | agent prefix 가 없는 브랜치. `ai-workflow/memory/branches/<branch-name>/` 사용 |

### 8.2 Mavis branch 예외

- housekeeping / docs hotfix / memory sync 같이 GitHub issue 없는 작업: `mvs/work_<YYMMDD>-<sprint-seq>-<short-key>` (issue 번호 생략).
- 외부 contribution (mavis-team 의 worker 가 spawn 한 결과): 자유 branch 명 허용. Mavis 가 인수 시 그대로 작업 후 PR.

### 8.3 Commit / PR

- commit message: `type(scope): summary` (Conventional Commits, 기존 5-워커 컨벤션 정합).
- PR body "추적성 영향" 섹션 ([`docs/traceability/sync-checklist.md`](../docs/traceability/sync-checklist.md) §3.7 정합).
- 모든 문서 변경은 PR 기반. 약식 hotfix 만 squash merge.
- 직접 commit 은 사용자 명시 요청 시. 그 외엔 PR 으로 흐름.

## 9. 인계 SOP (Handoff)

Mavis 의 인계는 5-워커 패턴과 다른 차원. vendor §5 의 위임 출력 스키마 + 본 저장소 의 4 방향:

### 9.1 Mavis → 사용자

- 작업 완료 후 짧은 한국어 요약 (1~3 bullet) + 다음 행동 제안.
- deliverable (파일/이미지/zip) 은 `<media src="..."/>` 태그로 첨부. 경로만 print 금지.
- 사용자 follow-up 이 필요하면 그 1 줄 + 옵션 (있으면) 함께.
- §1.1 의 한국어 보고 원칙 정합.

### 9.2 Mavis → mavis-team

- 본인 orchestrator 가 owner. mavis-team plan 으로 3+ tracks 분할.
- 각 track 의 subagent prompt 6섹션 표준 = vendor §4 위임 입력 스키마 1:1 매핑 (§5.2). reference file + 검증 기준 명시.
- verifier 가 결과 채점 → CycleReport → accept/retry 결정은 root 가.

### 9.3 Mavis → single-shot verifier

- `mavis communication send --command spawn --content '{"agent": "code-reviewer", "prompt": "..."}'`.
- parent session 상속. verifier 결과는 parent 에 report.
- producer 작업은 절대 single-spawn 으로 처리하지 않음. vendor §3 validation-worker 정합.
- 결과 보고는 vendor §5 위임 출력 스키마 정합.

### 9.4 Mavis → opencode branch (cross-worker)

- opencode 워커의 sprint branch 에서 작업 중인 deliverable 의 cross-cutting 정합 (governance / traceability) 갱신이 필요할 때:
  1. opencode 가 PR `opencode/work_*` 머지 후 close.
  2. Mavis 가 후속 `mvs/work_*` branch 에서 cross-cut 파일 (AGENTS.md, worker_division.md, release_v1_roadmap.md) 갱신 + PR.
  3. 사용자 confirm 후 merge.

## 10. 본 저장소에서 Mavis 의 기본 행동 (Day-1 Baseline, vendor §1-§9 동기화)

본 sprint (`feat/sync-minimax-workflow-v0.7.17`, 2026-06-15) 종료 시점 기준 Mavis 의 day-1 baseline:

1. **세션 시작** (vendor §2): 워크스페이스 인식 + 이 문서 + [`AGENTS.md`](../AGENTS.md) + 현재 git branch 의 `ai-workflow/memory/<agent>/<branch>/` (있으면) 4 종 파일 read. main branch 면 flat memory + `PROJECT_PROFILE.md` read.
2. **작업 모드 결정** (vendor §1.3, §1.3 매핑): 사용자 요청 → §1.3 의 6종 모드 중 1개 결정 → state.json / session_handoff.md 의 "현재 작업 모드" 섹션에 기록.
3. **사용자 요청 수신** → §2 라우팅 분기. low-complexity 면 직접, complex 면 `mavis-team`, verify 면 `single-spawn`.
4. **작업 수행** (vendor §1.2 컨텍스트 절약 + §5 sub-agent contract v1):
   - read-heavy / single-fix / doc-edit / prompt-tweak → 직접 (orchestrator 본인).
   - producer + 멀티파일 + 검증 필요 → `mavis-team` 또는 scratchpad 에 sub-todo 분해 후 직접.
   - verifier 필요 → `single-spawn` (validation-worker).
   - 3+ tracks → `mavis-team` fan-out (vendor §4.2).
5. **작업 종료** (vendor §8):
   - 본 워커 branch (있다면) 의 `state.json` / `session_handoff.md` / `work_backlog.md` / 최신 `backlog/YYYY-MM-DD.md` 갱신.
   - **vendor §7.2 동기화 루틴**: PROJECT_PROFILE.md + memory 의 work_backlog.md + state.json 동시 갱신.
   - PR body 에 추적성 영향 + 변경 요약.
   - deliverable 은 `<media />` 첨부.
   - 비동기 핸드오프 시작했으면 cron 등록.
   - **vendor §8 의 workflow-linter 실행**: `bash ai-workflow/skills/workflow-linter/scripts/run_lint.sh` — 문서 간 불일치 없는지 확인.
   - **v0.7.17 vendor smoke 회귀**: §4.4 의 2종 smoke 16/16 PASS 확인 (vendor 갱신 시점만).
6. **세션 idle (다음 사용자 요청 대기)**: `finished` 상태 유지. routable.

## 11. 기존 5-워커 워크플로우와의 관계

| 차원 | 5-워커 워크플로우 (worker_division.md) | Mavis 워크플로우 (본 문서) |
| --- | --- | --- |
| 단위 | 외부 AI worker (Claude/Codex/Gemini/Reasonix/OpenCode) | Mavis 세션 (root/branch) |
| 분할 기준 | 영역 (backend/infra/frontend/workflow/AI-ML) | 작업 복잡도 + track 수 + 검증 필요성 + 모드 (vendor §1.3) |
| 인계 | worker → worker handoff SOP | Mavis → 사용자 / mavis-team / single-spawn verifier / opencode branch (vendor §5 sub-agent contract v1) |
| 메모리 | `ai-workflow/memory/<agent>/<branch>/` | Mavis scratchpad + project/agent/user memory 3-layer |
| SSOT | per-agent branch memory | vendor maturity_matrix.json + PROJECT_PROFILE.md + MEMORY_GOVERNANCE.md |
| 브랜치 prefix | 5 개 (`claude/`, `codex/`, `gemini/`, `deepseek/`, `opencode/`) | `mvs/` + 5 개 보조 prefix (cross-worker 정합 시) |
| 트리거 | 사용자 invoke | 사용자 메시지 (자동) |

**공존 원칙**: 본 문서는 5-워커 워크플로우를 대체하지 않는다. Mavis 가 본 저장소에서 움직이는 방식을 정의하며, cross-worker 정합이 필요할 때 §9.4 의 패턴으로 5-워커 PR 에 후속 정합 PR 을 발행한다.

## 12. 변경 이력

| 일자 | 변경 | sprint |
| --- | --- | --- |
| 2026-06-04 | 1차 작성 — Mavis (MiniMax Code, MiniMax-M3) 의 본 저장소 운영 패턴 단일 진입점. 세션 모델, 작업 라우팅, 메모리 3-layer, communication/cron, skills/agents/MCP, hard limits, 브랜치/PR 컨벤션, 인계 SOP, day-1 baseline, 5-워커 워크플로우와의 관계 명시 | `mvs/work_260604-a-minimax-code-workflow-setup` |
| 2026-06-15 | **v0.7.17 vendor 동기화** — §0 상태 draft → active, §1.1 언어 원칙 (vendor §1.1), §1.2 컨텍스트 절약 (vendor §1.2), §1.3 작업 모드 6종 (vendor §1.3), §2 라우팅 표에 vendor contract v1 매핑 추가, §4 SSOT (vendor §7), §5 sub-agent contract v1 (vendor `orchestrator_subagent_contract_v1.md` §3-§5), §10 day-1 baseline 의 vendor §1-§9 동기화. 한국어 표준어 통일 (위임/스킬/플랜/마일스톤/SSOT). 본 문서를 Mavis 운영 패턴 SSOT 로 active 전환 | `feat/sync-minimax-workflow-v0.7.17` |
