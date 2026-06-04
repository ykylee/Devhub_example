# MiniMax Code (Mavis) Workflow Setup

- 문서 목적: 본 저장소에서 Mavis (MiniMax Code, 모델 `MiniMax-M3`) 가 작업을 운영할 때 따르는 단일 진입점. 기존 5-워커 워크플로우 (Claude/Codex/Gemini/Reasonix/OpenCode, [`docs/governance/worker_division.md`](../docs/governance/worker_division.md)) 와 공존하며, **Mavis 가 본 저장소 안에서 어떻게 움직이는지** 만 정의한다.
- 범위: 세션 모델, 작업 라우팅, 메모리 3-layer, communication, cron, 사용 가능 도구 (skill/agent/MCP), 브랜치 컨벤션, 인계 SOP, 하드 룰.
- 대상 독자: Mavis (root/branch 세션), 본 저장소에서 Mavis 를 호출하는 사용자, 후속 AI 에이전트.
- 상태: draft
- 최종 수정일: 2026-06-04
- 결정 근거 sprint: `mvs/work_260604-a-minimax-code-workflow-setup`
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

## 1. 세션 모델 (Session Model)

Mavis 는 **root session ↔ branch session** 트리로 동작한다.

- **root session**: 사용자가 처음 진입한 세션. `agentRole: orchestrator`. 프로젝트 전체 컨텍스트 + 사용자 요청을 받아 작업 라우팅 결정 + 결과 통합. 절대 종료되지 않고 다음 사용자 요청을 대기 (idle = `finished` 라고 부르며 routable 한 상태).
- **branch session**: root 가 spawn 한 작업 단위 세션. 부모 root 의 워크스페이스를 상속 + cross-session communication 으로 보고. 일반적으로 작업 완료 후 close.
- **hard rule**: branch 세션에서 다시 branch mavis 세션을 spawn 하지 않는다 (사용자 명시 요청 시 예외).

본 저장소에서는:

- **opencode 워커의 branch 메모리** (`ai-workflow/memory/opencode/<branch>/`) 와 **Mavis branch session 메모리** (`$MAVIS_SCRATCHPAD` = `~/.mavis/scratchpads/<rootSessionId>/scratchpad.md`) 가 별개 레이어.
- Mavis 가 작성하는 cross-session 노트는 scratchpad, 본 저장소의 workflow state 는 opencode branch memory. 두 위치는 중복하지 않고 역할 분리.

## 2. 작업 라우팅 (Task Routing)

사용자 요청이 들어오면 Mavis 는 다음 분기로 즉시 결정한다.

| 조건 | 결정 |
| --- | --- |
| 대화/Q&A/추천, 단순 정보 조회, 단일 파일 read, lightweight op | **Handle it yourself** — 직접 처리, branch spawn 안 함 |
| 컨텍스트 안에 deliverable 을 끝까지 그릴 수 있는 bounded scope 작업 (단일 파일 fix, bulk rename, 설정/문서/프롬프트 편집, quick draft) | **Handle it yourself** — `mavis-team` 로드 금지 |
| 3+ 독립 tracks, 멀티 소스/툴, high error cost, 멀티 스테이지 delivery chain | **`mavis-team` plan** — skill 로드 후 plan 실행 |
| 기존 deliverable 의 review/test/verify/audit (코드 리뷰, 테스트, 검증) | **Spawn single-shot worker** (`mavis communication send --command spawn`) — verifier 전용 채널 |
| Producer 작업 (코드/리팩토링/feature/bug fix) 을 single-spawn 으로 처리하고 싶을 때 | **금지** — 본인 또는 `mavis-team` 으로 라우팅 |

분기 결정은 작업 시작 전 1 회. 작업 중 복잡도가 escalate 되면 새 분기로 re-route 가능.

## 3. 메모리 모델 (3-Layer Memory)

Mavis 의 메모리는 3 layer, 가장 좁은 것부터 선택:

1. **Project memory** (`AGENTS.md` 또는 repo 내 토픽 파일 + `changelogs/`) — 이 저장소/프로젝트에서만 유효. 직접 edit, no CLI.
2. **Agent memory** (`mavis memory append mavis --content '...'` ) — 다른 프로젝트에서도 같은 Mavis 로 일할 때 유효.
3. **User memory** (`mavis memory append --user --reason '<cross-project justification>' --content '...'` ) — `--reason` 필수, 모든 프로젝트에서 유효할 때만.
4. **Scratchpad** (`$MAVIS_SCRATCHPAD` = `~/.mavis/scratchpads/<rootSessionId>/scratchpad.md`) — cross-session whiteboard, branch 세션이 자동 상속. git untracked.

규칙:

- append 는 새 entry 만 추가. 수정/삭제 는 memory 파일 직접 edit.
- 작업 종료 전 "Did I learn anything reusable?" 1 회 — 해당되면 즉시 기록.
- **언어**: 사용자와 같은 언어 (이 저장소는 한국어). 코드/명령어/경로/외부 시스템 명칭은 원문 유지.
- 메모리는 hint — 액션 전 verify. 프로젝트 메모리는 git tracked, agent/user 메모리는 `~/.mavis/memory/` 아래.

본 저장소에서는:

- **Project memory (layer 1)**: `AGENTS.md`, `ai-workflow/MEMORY_GOVERNANCE.md` (정책), `ai-workflow/minimax_code_workflow.md` (본 문서), `ai-workflow/memory/PROJECT_PROFILE.md` (프로젝트 프로파일), `ai-workflow/memory/<agent>/<branch>/` (외부 5-worker + mavis branch state, **layer 1 의 일부**), 도메인별 `requirements.md` / `architecture.md`.
- **Agent memory (layer 2)**: `~/.mavis/memory/agents/mavis.md`. Mavis 가 다른 repo 에서도 활용할 cross-project 운영 노트.
- **User memory (layer 3)**: `~/.mavis/memory/user.md`. 사용자의 역할/관심사/워크플로우 선호 (자연스러운 대화로 채움).
- **Scratchpad**: 본 root session ID `mvs_8952f7f57f9749a68171434a78f89960` 의 scratchpad 가 branch 세션과 공유.

**구체적 session-start read / session-end write 절차** + **중복 방지 규칙** + **외부 worker branch memory 와의 경계** 는 [`MEMORY_GOVERNANCE.md` §0.5](./MEMORY_GOVERNANCE.md) 의 단일 source-of-truth 를 따른다. 본 섹션은 그것의 *요약* 일 뿐.

## 4. Communication + Cron

### 4.1 Cross-session communication

`mavis communication send` API:

- root → branch, branch → root 보고, branch ↔ branch 모두 지원.
- report-back 실패 시: 5 초 후 1 회 재시도, 실패 시 scratchpad + IM notify.

### 4.2 Self-reminder (cron)

비동기 핸드오프 (CI, batch job, MR auto-merge, 외부 API, human reply 대기) 후 **반드시** cron 등록:

```bash
mavis cron self <name> --every <interval> --prompt "<text>"
```

예외: `mavis team plan` 자체는 heartbeat + CycleReport + unresponsive alert 가 내장되어 cron 등록 불필요.

### 4.3 외부 communication

- Feishu/Lark: `lark-tools` skill (또는 `lark-*` 세부 skill).
- X/Twitter: `x-link-reader` skill.
- IM 채널 라우팅: `mavis im` 로 hook/agent 매핑 관리.

## 5. 사용 가능 도구 (Tools)

### 5.1 Skills (직접 호출)

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

### 5.2 Sub-agents (직접 호출 가능한 native agent)

- `explore` — fast code exploration, glob/grep/Read 위주.
- `general` — 범용 multi-step worker.

### 5.3 MCP servers

`mavis mcp ls` 로 동적 확인. 본 저장소에서 자주 쓰는 것:

- `matrix` — VIDEO/AUDIO 이해, IMAGE/VIDEO/오디오 생성, **WEB_SEARCH** (네이티브 tool, `web_search` 직접 호출).
- `cu` (Computer Use) — 데스크탑 마우스/키보드 자동화. 25개 desktop_* 도구. 좌표는 0~1000 정규화.
- `playwright` — 브라우저 자동화. 공개 페이지 → `mavis-browser` 우선, SPA/anti-bot → playwright MCP.
- `trash` — recoverable delete (OS Trash). `mavis-trash` CLI 직접 호출.

### 5.4 Hard limits

- **single-spawn 은 verifier 전용**. producer work 를 single-spawn 으로 처리하지 말 것.
- **`mavis-team` 을 low-complexity 에 로드 금지** — 한 번에 끝낼 수 있는 일에 팀 plan 금지.
- **사용자에게 clarify 하지 말 것** — 답을 모를 때만, 그리고 정말 outcome 이 달라질 때만 물어본다.
- **in-scope collateral issue 는 묶어서 fix** — 작은 부수 발견이라도 사용자에게 되묻지 말고 같이 fix.
- **verification 없는 결과는 done 으로 확정 금지** — deliverable 후 verify 단계 동행.
- **`mavis communication send --command prompt`** 를 기존 worker session 에 producer 작업 단축으로 쓰지 말 것 (mavis-team 또는 본인).

## 6. 브랜치 / 커밋 / PR 컨벤션

### 6.1 브랜치 prefix

| prefix | 의미 | 비고 |
| --- | --- | --- |
| `mvs/` | Mavis root session 의 직접 작업 | 본 워커 prefix. 형식: `mvs/work_<YYMMDD>-<sprint-seq>-<short-key>` |
| `claude/` `codex/` `gemini/` `deepseek/` `opencode/` | 외부 AI 워커 | [`docs/governance/worker_division.md` §2.5](../docs/governance/worker_division.md) 정합 |
| `feat/` `fix/` `chore/` `docs/` `test/` | 일반 prefix | agent prefix 가 없는 브랜치. `ai-workflow/memory/branches/<branch-name>/` 사용 |

### 6.2 Mavis branch 예외

- housekeeping / docs hotfix / memory sync 같이 GitHub issue 없는 작업: `mvs/work_<YYMMDD>-<sprint-seq>-<short-key>` (issue 번호 생략).
- 외부 contribution (mavis-team 의 worker 가 spawn 한 결과): 자유 branch 명 허용. Mavis 가 인수 시 그대로 작업 후 PR.

### 6.3 Commit / PR

- commit message: `type(scope): summary` (Conventional Commits, 기존 5-워커 컨벤션 정합).
- PR body "추적성 영향" 섹션 ([`docs/traceability/sync-checklist.md`](../docs/traceability/sync-checklist.md) §3.7 정합).
- 모든 문서 변경은 PR 기반. 약식 hotfix 만 squash merge.
- 직접 commit 은 사용자 명시 요청 시. 그 외엔 PR 으로 흐름.

## 7. 인계 SOP (Handoff)

Mavis 의 인계는 5-워커 패턴과 다른 차원:

### 7.1 Mavis → 사용자

- 작업 완료 후 짧은 한국어 요약 (1~3 bullet) + 다음 행동 제안.
- deliverable (파일/이미지/zip) 은 `<media src="..."/>` 태그로 첨부. 경로만 print 금지.
- 사용자 follow-up 이 필요하면 그 1 줄 + 옵션 (있으면) 함께.

### 7.2 Mavis → mavis-team

- 본인 orchestrator 가 owner. mavis-team plan 으로 3+ tracks 분할.
- 각 track 의 subagent prompt 에 6섹션 표준 (TASK/EXPECTED OUTCOME/REQUIRED TOOLS/MUST DO/MUST NOT DO/CONTEXT) + reference file + 검증 기준 명시.
- verifier 가 결과 채점 → CycleReport → accept/retry 결정은 root 가.

### 7.3 Mavis → single-shot verifier

- `mavis communication send --command spawn --content '{"agent": "code-reviewer", "prompt": "..."}'`.
- parent session 상속. verifier 결과는 parent 에 report.
- producer 작업은 절대 single-spawn 으로 처리하지 않음.

### 7.4 Mavis → opencode branch (cross-worker)

- opencode 워커의 sprint branch 에서 작업 중인 deliverable 의 cross-cutting 정합 (governance / traceability) 갱신이 필요할 때:
  1. opencode 가 PR `opencode/work_*` 머지 후 close.
  2. Mavis 가 후속 `mvs/work_*` branch 에서 cross-cut 파일 (AGENTS.md, worker_division.md, release_v1_roadmap.md) 갱신 + PR.
  3. 사용자 confirm 후 merge.

## 8. 본 저장소에서 Mavis 의 기본 행동 (Day-1 Baseline)

본 sprint (`mvs/work_260604-a-minimax-code-workflow-setup`) 종료 시점 기준 Mavis 의 day-1 baseline:

1. **세션 시작**: 워크스페이스 인식 + 이 문서 + [`AGENTS.md`](../AGENTS.md) + 현재 git branch 의 `ai-workflow/memory/<agent>/<branch>/` (있으면) 4 종 파일 read. main branch 면 flat memory + `PROJECT_PROFILE.md` read.
2. **사용자 요청 수신**: §2 라우팅 분기. low-complexity 면 직접, complex 면 `mavis-team`.
3. **작업 수행**:
   - read-heavy / single-fix / doc-edit / prompt-tweak → 직접.
   - producer + 멀티파일 + 검증 필요 → `mavis-team` 또는 scratchpad 에 sub-todo 분해 후 직접.
   - verifier 필요 → single-spawn.
4. **작업 종료**:
   - 본 워커 branch (있다면) 의 `state.json` / `session_handoff.md` / `work_backlog.md` / 최신 `backlog/YYYY-MM-DD.md` 갱신.
   - PR body 에 추적성 영향 + 변경 요약.
   - deliverable 은 `<media />` 첨부.
   - 비동기 핸드오프 시작했으면 cron 등록.
5. **세션 idle (다음 사용자 요청 대기)**: `finished` 상태 유지. routable.

## 9. 기존 5-워커 워크플로우와의 관계

| 차원 | 5-워커 워크플로우 (worker_division.md) | Mavis 워크플로우 (본 문서) |
| --- | --- | --- |
| 단위 | 외부 AI worker (Claude/Codex/Gemini/Reasonix/OpenCode) | Mavis 세션 (root/branch) |
| 분할 기준 | 영역 (backend/infra/frontend/workflow/AI-ML) | 작업 복잡도 + track 수 + 검증 필요성 |
| 인계 | worker → worker handoff SOP | Mavis → 사용자 / mavis-team / single-spawn verifier / opencode branch |
| 메모리 | `ai-workflow/memory/<agent>/<branch>/` | Mavis scratchpad + project/agent/user memory 3-layer |
| 브랜치 prefix | 5 개 (`claude/`, `codex/`, `gemini/`, `deepseek/`, `opencode/`) | `mvs/` + 5 개 보조 prefix (cross-worker 정합 시) |
| 트리거 | 사용자 invoke | 사용자 메시지 (자동) |

**공존 원칙**: 본 문서는 5-워커 워크플로우를 대체하지 않는다. Mavis 가 본 저장소에서 움직이는 방식을 정의하며, cross-worker 정합이 필요할 때 §7.4 의 패턴으로 5-워커 PR 에 후속 정합 PR 을 발행한다.

## 10. 변경 이력

| 일자 | 변경 | sprint |
| --- | --- | --- |
| 2026-06-04 | 1차 작성 — Mavis (MiniMax Code, MiniMax-M3) 의 본 저장소 운영 패턴 단일 진입점. 세션 모델, 작업 라우팅, 메모리 3-layer, communication/cron, skills/agents/MCP, hard limits, 브랜치/PR 컨벤션, 인계 SOP, day-1 baseline, 5-워커 워크플로우와의 관계 명시 | `mvs/work_260604-a-minimax-code-workflow-setup` |
