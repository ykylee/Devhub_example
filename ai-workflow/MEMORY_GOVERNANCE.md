# AI-First Memory Governance

- 문서 목적: AI 에이전트가 관리하는 운영 문서(Workflow State)의 관리 규칙과 템플릿을 정의한다.
- 범위: 상태 문서 분류, 작성 표준, 메타데이터 요구사항
- 대상 독자: AI 에이전트, 저장소 관리자
- 상태: stable
- 최종 수정일: 2026-06-04
- 관련 문서: [../ai-workflow/WORKFLOW_INDEX.md](../ai-workflow/WORKFLOW_INDEX.md), [../README.md](../README.md)

이 문서는 `ai-workflow/memory/` 하위 문서를 작성할 때 AI 에이전트가 준수해야 할 규칙과 템플릿을 정의합니다.

> Mavis (MiniMax Code) 의 본 저장소 운영 패턴은 [`./minimax_code_workflow.md`](./minimax_code_workflow.md) 단일 진입점으로 정의한다. 본 문서는 그 워크플로우 안에서 *어떤 메모리 파일을 어디에 둘지* 의 디렉터리/메타 표준만 다룬다.

## 0. 브랜치별 Memory 표준

신규 작업 상태의 source of truth는 브랜치별 디렉터리다.

- 표준 위치: `ai-workflow/memory/<agent>/<branch>/`
- 예시: `ai-workflow/memory/codex/service-action-command/`
- 예시: `ai-workflow/memory/gemini/phase6/`
- 예시: `ai-workflow/memory/opencode/work_260604-a-opencode-workflow-bootstrap/`
- 예시: `ai-workflow/memory/mvs/work_260604-a-minimax-code-workflow-setup/` (Mavis 직접 작업, 2026-06-04 추가)
- agent prefix가 없는 브랜치: `ai-workflow/memory/branches/<branch-name>/`

브랜치별 디렉터리는 아래 4종을 기본 세트로 가진다.

- `state.json`
- `session_handoff.md`
- `work_backlog.md`
- `backlog/YYYY-MM-DD.md`

공용 문서는 flat 경로에 둔다.

- `ai-workflow/memory/PROJECT_PROFILE.md`
- `ai-workflow/memory/repository_assessment.md`
- `ai-workflow/memory/environments/`
- 공용 로드맵 문서

flat `state.json`, `session_handoff.md`, `work_backlog.md`, `backlog/`는 legacy fallback 및 공용 색인 용도다. 브랜치 작업의 신규 상태 갱신은 flat 경로가 아니라 브랜치별 디렉터리에 기록한다.

### 0.5 Mavis 3-Layer 메모리 매핑 (2026-06-04 추가)

Mavis (MiniMax Code) 의 메모리는 3 layer 구조다. 본 저장소에서의 위치는 다음과 같이 매핑한다.

| Layer | 정의 | 본 저장소 위치 | 비고 |
| --- | --- | --- | --- |
| **Project memory (layer 1)** | 이 저장소/프로젝트에서만 유효. git tracked. | `AGENTS.md`, `ai-workflow/MEMORY_GOVERNANCE.md`, `ai-workflow/minimax_code_workflow.md`, `ai-workflow/memory/PROJECT_PROFILE.md`, `ai-workflow/memory/<agent>/<branch>/{state.json,session_handoff.md,work_backlog.md,backlog/YYYY-MM-DD.md}`, `ai-workflow/memory/branches/<branch-name>/`, 도메인 docs (`docs/domain/*/`) | 외부 AI worker 의 branch memory 도 모두 layer 1 (project scope) 다 |
| **Agent memory (layer 2)** | 같은 Mavis (agent) 로 다른 프로젝트에서 작업할 때도 유효 | `~/.mavis/memory/agents/mavis.md` (또는 topic file) | CLI: `mavis memory append mavis --content '...'` |
| **User memory (layer 3)** | 모든 프로젝트 + 모든 agent 에서 사용자 정체성/선호가 유효 | `~/.mavis/memory/user.md` | CLI: `mavis memory append --user --reason '<cross-project justification>' --content '...'` |
| **Scratchpad (cross-session)** | root session 공유 whiteboard, branch 세션 자동 상속 | `$MAVIS_SCRATCHPAD` = `~/.mavis/scratchpads/<rootSessionId>/scratchpad.md` | 환경변수 `MAVIS_SCRATCHPAD` 자동 주입. root 가 새 root 로 회전하면 새 scratchpad 시작 (이전 root 의 파일은 permanent archive) |

**중복 방지 규칙**:

- 한 사실은 한 곳에만 기록. (a) branch 상태는 branch memory, (b) cross-session 작업 노트는 scratchpad, (c) cross-project 운영 노트는 agent memory, (d) 사용자 정체성은 user memory. 어디에도 속하지 않으면 scratchpad.
- branch memory 4종 파일 (state.json/session_handoff.md/work_backlog.md/backlog/) 은 git tracked. commit 시 co-evolve.
- scratchpad 는 git untracked. 절대 commit 금지 (사용자/세션 privacy 정보 포함 가능).

**session-start 읽기 순서 (Mavis)**:

1. `AGENTS.md` (root 진입)
2. `ai-workflow/minimax_code_workflow.md` (Mavis 운영 단일 진입점)
3. 본 문서 `MEMORY_GOVERNANCE.md` (메모리 정책)
4. `ai-workflow/memory/PROJECT_PROFILE.md` (프로젝트 프로파일)
5. `git branch --show-current` → 해당 branch 의 `ai-workflow/memory/<agent>/<branch>/` 4종 (없으면 flat fallback)
6. layer 2: `~/.mavis/memory/agents/mavis.md` (존재 시)
7. layer 3: `~/.mavis/memory/user.md` (존재 시)
8. `$MAVIS_SCRATCHPAD` (root session 의 cross-session 메모, 존재 시)

**session-end 쓰기 순서 (Mavis)**:

1. branch 작업이면 → `ai-workflow/memory/<agent>/<branch>/` 4종 갱신
2. main 작업이면 → flat `ai-workflow/memory/` 갱신
3. cross-session 노트 → `$MAVIS_SCRATCHPAD` (branch 세션이면 같은 root scratchpad)
4. cross-project 학습 → `mavis memory append mavis ...` (layer 2)
5. 사용자 정체성/선호 학습 → `mavis memory append --user --reason '...' ...` (layer 3)
6. PR 의 "추적성 영향" 섹션 + 변경 요약 (PR 발행 시)

**외부 worker branch memory 와의 경계**:

- Mavis 는 외부 worker (claude/codex/gemini/deepseek/opencode) 의 branch memory 를 *능동적으로* read/edit 하지 않는다. cross-cut 정합이 필요할 때만 후속 PR 로 정합.
- Mavis 의 작업 결과를 외부 worker 에게 알려야 할 때는 (a) PR body 의 변경 요약, (b) 사용자 안내, (c) `worker_division.md` §7.4 cross-worker PR 패턴 — 이 3가지 채널 중 1개로만.
- 외부 worker 의 session_handoff.md 에 Mavis 가 직접 write 하는 경우는 *금지* (오너십 침범).

## 1. 작성 규칙 (Writing Rules)

- **언어**: 사용자 보고용 요약은 한국어를 사용하되, 상태 값이나 기술적 명칭은 영문 표준을 권장합니다.
- **간결성**: 중복된 설명을 피하고, 변경 사항(Diff)과 다음 행동(Next Action)에 집중합니다.
- **구조화**: Key-Value 쌍(예: `Status: in_progress`) 또는 Markdown Table을 적극 활용합니다.
- **격리 (Isolation)**: 메인 개발 브랜치(`main`) 외의 작업 브랜치에서는 `ai-workflow/memory/[agent_name]/[branch_name]/` 하위의 전용 폴더를 사용하여 상태를 관리합니다. 이는 다중 에이전트 협업 시의 병합 충돌을 방지하기 위함입니다.
- **루트의 역할 (Root Role)**: `ai-workflow/memory/` 루트의 `state.json`, `session_handoff.md` 등은 프로젝트의 최종 통합 상태(Main branch state) 또는 현재 통합 테스트 브랜치의 요약 상태를 나타냅니다.

## 2. 표준 템플릿 (Standard Templates)

### 📂 Session Handoff (`session_handoff.md`)
```markdown
# Session Handoff
- Branch: [branch_name]
- Updated: [YYYY-MM-DD HH:mm]

## 🎯 Current Focus
[현재 작업의 핵심 목표 1줄]

## 📊 Work Status
- [TASK-ID] [Title]: [Status] ([Progress %])
- [최근 수행한 핵심 변경 사항 및 결과]

## ⏭️ Next Actions
- [ ] [다음에 즉시 수행할 작업]

## ⚠️ Risks & Blockers
- [차단 요소 또는 주의가 필요한 아키텍처적 결정 사항]
```

### 📂 Task Detail (`backlog/tasks/TASK-XXX.md`)
```markdown
---
id: TASK-XXX
status: [planned|in_progress|done|blocked]
created_at: YYYY-MM-DD
---
# [Task Title]

## 📝 Description
[작업의 정의 및 범위]

## 🛠️ Implementation Log
- [YYYY-MM-DD]: [수행 내용 요약]

## ✅ Outcome
[완료 시 결과물 또는 검증 결과]
```

### 📂 Daily Backlog Index (`backlog/YYYY-MM-DD.md`)
```markdown
# YYYY-MM-DD Branch Work Backlog

- Purpose: Link task detail files for one working day.
- Status: in_progress
- Updated: YYYY-MM-DD

## Tasks

- TASK-XXX Task title: `./tasks/YYYY-MM-DD_TASK-XXX.md`
```

- `backlog/tasks/*.md` is the source of truth for detailed task state.
- `backlog/YYYY-MM-DD.md` is a tracked lightweight index. Keep it small and link-oriented.
- On merge conflicts, rebuild the daily index from task links and resolve detailed state in each task file.

## 3. 에이전트 행동 지침

- 세션 시작 시 현재 git 브랜치를 확인하고 브랜치별 memory 디렉터리를 우선 읽으십시오.
- 세션 종료 시 브랜치별 `session_handoff.md`를 위 템플릿에 맞춰 갱신하십시오.
- 새로운 작업 시작 시 브랜치별 `backlog/tasks/` 폴더에 템플릿 기반의 신규 파일을 생성하십시오.
- 날짜별 백로그에는 신규 task 파일 링크만 추가하고, 긴 상세 기록은 task 파일에 남기십시오.
- 상태 업데이트 시 자연어 서술보다는 불렛 포인트와 상태 키워드를 우선하십시오.

### 3.1 Mavis (MiniMax Code) 행동 지침 (2026-06-04 추가)

- §0.5 의 읽기/쓰기 순서를 따른다. 추가 규칙:
  - Mavis 가 *외부 worker 의 영역* (backend/infra/frontend) 으로 직접 들어가는 것은 권장하지 않는다 — cross-cut 정합 (governance/traceability/release_v1_roadmap/worker_division/AGENTS.md) 만 직접 진입.
  - branch prefix 가 `mvs/*` 이면 본 저장소에서 직접 작업, 그 외 prefix 면 외부 worker 영역으로 간주.
  - 작업 종료 시 branch memory 4종 + (필요 시) scratchpad + (학습 시) agent/user memory 동시 갱신. 한 곳만 갱신하고 다른 곳 빠뜨리지 않는다.
  - 본 워커 (`mvs`) 의 branch memory 4종은 외부 worker 의 그것과 *동일한 schema* (`schema_version`, `branch`, `status`, `last_update` 등) 를 따른다. 도구는 같고 prefix 만 다르다.
  - `mavis-team` plan 으로 분할된 subagent 의 결과는 root 의 scratchpad 에 취합, branch memory 의 `recent_handoffs` 섹션에는 *PR-level* 만 기록 (세부 결과는 scratchpad 참조).
