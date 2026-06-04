# AGENTS.md

- 문서 목적: 모든 AI 에이전트(Codex, Reasonix 등)가 이 저장소에서 먼저 읽어야 할 workflow 진입 규칙과 기본 작업 원칙을 제공한다.
- 범위: 세션 복원, workflow state docs 참조 순서, 사용자 보고 언어, 기본 실행/검증 명령, **v1.0 릴리즈 로드맵 + 워커 분업**
- 대상 독자: Codex, Reasonix (deepseek-v4), 저장소 관리자, workflow 설계자
- 상태: active
- 최종 수정일: 2026-06-04 (OpenCode Lane 정의 보강)
- 관련 문서: `ai-workflow/MEMORY_GOVERNANCE.md`, `ai-workflow/memory/<agent>/<branch>/state.json`, `ai-workflow/memory/PROJECT_PROFILE.md`, `docs/governance/README.md` (거버넌스 진입점), `docs/governance/document-standards.md`, `docs/governance/worker_division.md` (**워커 분업 — Codex/Reasonix 영역**), `docs/planning/release_v1_roadmap.md` (**v1.0 릴리즈 로드맵**), `docs/traceability/README.md`

## v1.0 릴리즈 로드맵 + 워커 분업

모든 신규 sprint 진입 전 다음 2 문서 확인:

- [`docs/planning/release_v1_roadmap.md`](docs/planning/release_v1_roadmap.md) — v1.0 scope + 잔여 carve 우선순위 (P0~P3) + 마일스톤 + sprint 별 워커 분담
- [`docs/governance/worker_division.md`](docs/governance/worker_division.md) — Codex 영역 (infra + CI + security) + 인계 SOP + 충돌 처리

## 목적

이 저장소에서는 표준 AI 워크플로우를 기준으로 작업한다. 세션 시작, backlog 갱신, 문서 동기화, 세션 종료는 `ai-workflow/` 아래 문서를 우선 기준으로 삼는다.

## 항상 먼저 읽을 문서

1. 현재 git 브랜치를 확인한다: `git branch --show-current`
2. 브랜치별 memory 디렉터리를 우선 읽는다.
   - 브랜치명이 `codex/service-action-command`이면 `ai-workflow/memory/codex/service-action-command/`
   - 브랜치명이 `claude/phase13`이면 `ai-workflow/memory/claude/phase13/`
   - 브랜치명이 `gemini/...`이면 `ai-workflow/memory/gemini/<branch-suffix>/`
   - 브랜치명이 `deepseek/...`이면 `ai-workflow/memory/deepseek/<branch-suffix>/` (Reasonix 포함)
   - agent prefix가 없는 브랜치는 `ai-workflow/memory/branches/<branch-name>/`를 사용한다.
3. 브랜치별 디렉터리에서 아래 문서를 먼저 읽는다.
   - `state.json`
   - `session_handoff.md`
   - `work_backlog.md`
   - 최신 `backlog/YYYY-MM-DD.md`
4. 공용 기준 문서를 읽는다.
   - `ai-workflow/memory/PROJECT_PROFILE.md`
   - `ai-workflow/memory/repository_assessment.md`

`ai-workflow/memory/state.json`, `ai-workflow/memory/session_handoff.md`, `ai-workflow/memory/work_backlog.md`, `ai-workflow/memory/backlog/` 같은 flat 경로는 legacy fallback 및 공용 색인이다. 신규 작업 상태는 브랜치별 memory 디렉터리에 기록한다.

`ai-workflow/` 는 세션 복원과 workflow 상태 관리용 메타 레이어다. 프로젝트 코드나 프로젝트 문서를 탐색할 때는 이 경로를 기본 탐색 범위에 넣지 말고, workflow 문서 자체를 갱신하거나 현재 세션 상태를 복원할 때만 예외적으로 참조한다.

## 작업 원칙

- 작업을 시작하기 전에 목적, 범위, 영향 문서를 짧게 정리한다.
- 작업 상태는 `planned`, `in_progress`, `blocked`, `done` 중 하나로 관리한다.
- 검증하지 않은 결과는 완료로 확정하지 않는다. 모든 신규 기능은 `docs/tests/e2e_testing_strategy.md` 지침에 따라 E2E 테스트를 작성/수행해야 한다.
- 세션 종료 전에는 브랜치별 `state.json`, `session_handoff.md`, `work_backlog.md`, 최신 backlog 를 갱신한다.
- **추적성 동기화**: 모든 PR 은 `docs/traceability/sync-checklist.md` 절차를 따른다. 영향 받는 단계의 ID (REQ/ARCH/API/RM/IMPL/UT/TC) 발급 또는 갱신 + `docs/traceability/report.md` 매트릭스 row 갱신 + PR body 의 "추적성 영향" 섹션 채움 (`.github/pull_request_template.md` 참조).

## 언어와 컨텍스트 원칙

- 사용자에게 직접 보이는 작업 보고, 상태 요약, 문서 갱신 문안은 기본적으로 한국어로 작성한다.
- 코드, 명령어, 파일 경로, 설정 key, 외부 시스템 고유 명칭은 필요할 때 원문 그대로 유지한다.
- 내부 사고 과정과 임시 분류는 모델이 가장 효율적인 방식으로 처리하되, 사용자에게는 필요한 결론과 다음 행동만 짧게 전달한다.
- 장문의 중간 reasoning, 중복 요약, 불필요한 자기 설명을 피한다.
- handoff 와 backlog 에는 다음 세션에 필요한 핵심 사실만 남겨 불필요한 컨텍스트 누적을 줄인다.

## 프로젝트 실행 기본값

- 설치: `TODO: 설치 명령 입력`
- 로컬 실행: `TODO: 로컬 실행 명령 입력`
- 빠른 테스트: `TODO: 빠른 테스트 명령 입력`
- 격리 테스트: `TODO: 격리 테스트 명령 입력`
- 실행 확인: `TODO: 실행 확인 명령 입력`

## 문서 작업 기준

- 문서 위키 홈: `README.md`
- 운영 문서 위치: `ai-workflow/memory/`
- 브랜치별 운영 문서 위치: `ai-workflow/memory/<agent>/<branch>/`
- backlog 위치: `ai-workflow/memory/<agent>/<branch>/backlog/`
- session handoff 위치: `ai-workflow/memory/<agent>/<branch>/session_handoff.md`
- flat memory 위치: legacy fallback 및 공용 색인 전용
- 문서 포맷 원칙: 원본은 Markdown(`.md`) 유지, HTML은 보고/취합용 파생 산출물로만 사용 (`docs/governance/document-standards.md` §0)

## Codex 전용 메모

- Codex 는 프로젝트 루트의 `AGENTS.md` 를 읽으므로, 상세 정책은 본 문서에서 시작하고 세부 운영 기준은 `ai-workflow/` 문서를 참조한다.
- OpenAI 관련 질문이 나오면 OpenAI 문서 MCP 를 우선 사용하는 구성을 권장한다.
- 가능한 경우 메인 에이전트는 조정과 통합에 집중하고, bounded scope 의 읽기/쓰기/검증 작업은 worker 성격의 서브 에이전트로 분리하는 패턴을 권장한다.
- worker 에게는 책임 파일과 종료 조건을 명확히 넘기고, 메인 에이전트에는 핵심 사실과 결과만 다시 모은다.
- `main`/`small` 모델을 함께 운영한다면, 메인 에이전트는 난도 높은 판단과 통합에, worker 는 bounded scope 탐색/초안/검증에 우선 배치하는 편이 효율적이다.
- 신규 프로젝트 기준 초안이다. 프로젝트 고유의 실행 명령과 문서 구조가 정확한지 확인해야 한다.

## Reasonix (deepseek-v4) 전용 메모

- Reasonix 는 Codex 와 동일한 workflow 레이어(`ai-workflow/`)를 따르며, 브랜치별 memory 디렉터리 패턴을 동일하게 사용한다.
- **브랜치 생성 시 반드시 `deepseek/` prefix 를 사용한다.** 브랜치명 예: `deepseek/construct_workflow_for_deepseek`
- 표준 sprint branch 명명 규칙은 `docs/governance/worker_division.md` §2.5 를 따른다: `deepseek/work_<YYMMDD>-<sprint-seq>-<issue-num>-<short-key>`
- 브랜치 prefix `deepseek/` → `ai-workflow/memory/deepseek/<branch-suffix>/`
- Reasonix 의 기본 모델은 `deepseek-v4-flash`이며, 복잡한 cross-file 리팩토링 시 자동으로 `deepseek-v4-pro` 로 escalation 된다.
- Reasonix 의 기본 도구 세트에는 GitHub MCP, Filesystem MCP, Memory KG, Puppeteer 가 포함되어 있다.
- `run_command` 사용 시 `cd` 가 체인에서 거부되므로, cwd 가 필요한 명령은 `cwd` 인자를 직접 전달하거나 `--prefix`/`-C` 플래그를 사용한다.
- 현재 브랜치가 `main`이 아닐 때는 `ai-workflow/` 메타 레이어를 기본 탐색 범위에 포함하지 말고, workflow 문서 갱신이나 세션 복원 시에만 참조한다.
- 프로젝트 실행 기본값(TODO 항목들)은 아직 설정되지 않았다 (`TODO: ...` 상태). Reasonix 세션 시작 시 `PROJECT_PROFILE.md` §3 기본 명령을 직접 참조하여 실행한다.

## OpenCode (Sisyphus / MiniMax-M3) 전용 메모

- OpenCode 워커는 Codex / Claude / Gemini / Reasonix 와 동일한 workflow 레이어(`ai-workflow/`)를 따르며, 브랜치별 memory 디렉터리 패턴을 동일하게 사용한다.
- **브랜치 생성 시 반드시 `opencode/` prefix 를 사용한다.** 브랜치명 예: `opencode/work_260604-a-opencode-workflow-bootstrap`
- 표준 sprint branch 명명 규칙은 `docs/governance/worker_division.md` §2.5 를 따른다: `opencode/work_<YYMMDD>-<sprint-seq>-<issue-num>-<short-key>`
- 브랜치 prefix `opencode/` → `ai-workflow/memory/opencode/<branch-suffix>/`
- OpenCode 의 기본 에이전트 식별자는 **Sisyphus** 이며, 기본 모델은 `MiniMax-M3` 다. 복잡한 cross-file 리팩토링·아키텍처 결정은 Oracle 같은 specialist 호출로 escalate 한다.
- OpenCode 는 메인 에이전트 조정/통합에 집중하고, bounded scope 의 읽기/쓰기/검증 작업은 `explore` / `librarian` / `Sisyphus-Junior` 같은 worker 성격 서브에이전트에 위임하는 패턴을 권장한다.
- **Lane 정의 (`worker_division.md` §1.4, 2026-06-04 확정)**: Lane 1 = workflow/governance curation, Lane 2 = cross-cutting validation + test infrastructure, Lane 3 = AI/ML service prep (v1.1/v2). Lane 1·2 는 즉시 carve 진입 가능, Lane 3 는 v1.0 출시 후.
- 사용자에게 보이는 작업 보고, handoff, backlog, 사용자 안내 문구는 기본 한국어로 작성한다 (Reasonix 와 동일). 코드/명령어/경로/외부 시스템 명칭은 원문 유지.
- 현재 브랜치가 `main`이 아닐 때는 `ai-workflow/` 메타 레이어를 기본 탐색 범위에 포함하지 말고, workflow 문서 갱신이나 세션 복원 시에만 참조한다.
- 첫 sprint (`opencode/work_260604-a-opencode-workflow-bootstrap`) 는 governance 부트스트랩에 한정. 두 번째 sprint (`opencode/work_260604-b-opencode-areas`) 가 §1.4 본문 정의를 다룸.
