# AGENTS.md

- 문서 목적: 모든 AI 에이전트(어떤 워커든)가 이 저장소에서 먼저 읽어야 할 workflow 진입 규칙과 기본 작업 원칙을 제공한다.
- 범위: 세션 복원, workflow state docs 참조 순서, 사용자 보고 언어, 기본 실행/검증 명령, **v0.1.0 릴리즈 로드맵 (워커 분업 전면 취소 결정 2026-06-09 반영)**, **사외/사내 2-tier 형상관리 분리 (2026-06-10 결정)**
- 대상 독자: 모든 AI 워커 (Claude/Codex/Gemini/Reasonix/OpenCode/Mavis/기타), 저장소 관리자, workflow 설계자
- 상태: active
- 최종 수정일: 2026-06-23 (`backend-knowledge` → 외부 repo `yklee/ai_library` extraction 결정, vendor import pattern 역방향 — 본 저장소에서 `backend-knowledge/` 디렉터리 삭제됨 + ADR-0037/0038 §6 supersession row + umbrella §1.2 G1 redirect + child doc §8 row + 본 AGENTS.md line 29 redirect. §15 ADR supersession 정공법 정합 — historical supersession (extraction, 외부 repo 이전) + M-v0.2.3+ 부터 supersession 가능 + docs/governance/worker_division.md §4.2 1:1 정합. **상세**: [`docs/planning/release_v0-2_roadmap.md`](docs/planning/release_v0-2_roadmap.md) §9 변경 이력 2026-06-23 row 정공법) + 2026-06-17 (v0.2.0 umbrella + ADR-0037 OKF + ADR-0038 backend-knowledge 추가, Q&A 11/11 결정 완료)
- 관련 문서: `ai-workflow/MEMORY_GOVERNANCE.md`, `ai-workflow/memory/<agent>/<branch>/state.json`, `ai-workflow/memory/PROJECT_PROFILE.md`, `docs/governance/README.md` (거버넌스 진입점), `docs/governance/document-standards.md`, `docs/governance/worker_division.md` (**§0 워커 분업 전면 취소 + §6 사외/사내 2-tier 분업**), `docs/planning/release_v0-1_roadmap.md` (**v0.1.0 릴리즈 로드맵**), `docs/planning/release_v0-2_roadmap.md` (**v0.2.0 릴리즈 로드맵 — 외부 시스템 연동 + OKF 기반 AI Agent Library**, 2026-06-17 accepted, [ADR-0037 OKF v0.1 채택](./adr/0037-okf-adoption.md) + [ADR-0038 backend-knowledge 신설](./adr/0038-backend-knowledge-creation.md), Q&A 11/11 결정 완료), `docs/traceability/README.md`

## v0.1.0 릴리즈 로드맵

**2026-06-09 결정 — 워커 분업 전면 취소**. 본 AGENTS.md 의 v0.1.0 릴리즈 로드맵 진입 시 워커 분담 / 인계 SOP / 충돌 처리 SOP 의 강제력을 모두 무효로 한다. 모든 신규 sprint / PR / 작업은 **어느 에이전트로든 자유롭게** 진행 가능.

모든 신규 sprint 진입 전 다음 1 문서 확인:

- [`docs/planning/release_v0-1_roadmap.md`](docs/planning/release_v0-1_roadmap.md) — v0.1.0 scope + 잔여 carve 우선순위 (P0~P3) + 마일스톤 (워커 분담 표 §5 는 2026-06-09 취소, 작업 우선순위/P0~P3 자체는 유효)

> 참고: [`docs/governance/worker_division.md`](docs/governance/worker_division.md) 는 2026-06-09 전면 취소 결정의 historical record + 유지되는 정책 (ADR supersession 정공법, Owner 권한) + **2026-06-10 §6 사외/사내 2-tier 분업** 만 보존. 강제력 없음.

## v0.2.0 릴리즈 로드맵 (2026-06-17 결정)

**v0.2.0 = 외부 시스템 연동 + 데이터 취합을 별도의 백엔드(`backend-knowledge`)로 모으고, Google OKF v0.1 기반 AI Agent Library 로 통합.** 사용자가 2026-06-17 결정.

모든 신규 sprint / 작업 진입 전 다음 1 문서 확인:

- [`docs/planning/release_v0-2_roadmap.md`](docs/planning/release_v0-2_roadmap.md) — v0.2.0 scope (G1~G7: backend-knowledge 단일화 / 외부 연동 흡수 / OKF 형 bundle / 3가지 기능 / 1차 raw API / 완전 독립 운영 / Q&A 11/11 결정) + 마일스톤 (M-v0.2.0-alpha~v0.3.0 6 단계) + cross-section 정합 (round 1/1.1/2 + §3/§4/§5/§6/§7/§8 self-review)

핵심 결정:
- **신규 백엔드**: `backend-knowledge/` → **2026-06-23 extraction**: 외부 repo [`yklee/ai_library`](https://homelab.ddn777.synology.me/gitea/yklee/ai_library) (Gitea private). 본 저장소에서 `backend-knowledge/` 디렉터리 삭제됨. (Python 3.13+ / FastAPI / OKF / Pi (pi.dev) v0.79.6 LLM enrich, **완전 standalone** — 다른 backend 연결 ❌, vendored re-introduce 가능)
- **ADR-0037**: [OKF v0.1 채택](./adr/0037-okf-adoption.md) (1차 출처: Google SPEC.md / README.md, Apache 2.0, 1 concept = 1 .md, frontmatter `type` 1개 필수, 8종 type enum)
- **ADR-0038**: [`backend-knowledge` 신설](./adr/0038-backend-knowledge-creation.md) (외부 시스템 7종 source 만 단방향, M-v0.2.3 운영 기준: Gitea 4 sub-plugin gitea_repo_pull / gitea_issue / gitea_wiki / gitea_action + homelab + metrics + hrdb, 2026-06-17 A/A 결정 + 2026-06-18 supersession 정합)
- **backend-ai/ 폐기** (M-v0.2.2, placeholder 정리)
- **Q&A 11/11 결정 완료** (Q1~Q11, release_v0-2_roadmap.md §7)

sprint 진입 checklist (release_v0-2_roadmap.md §5.3) 6 항목 중 4 항목 완료 (2026-06-17): umbrella doc publish ✅ / child doc active 전환 ✅ / state.json M-v0.2.0 row 발급 ✅ / OKF SPEC.md 1차 정독 ✅. 2 항목 별도: `backend-knowledge/` 디렉터리 skeleton (별도 PR) / GitHub milestone v0.2.0 (별도).

## 사외 / 사내 2-tier 형상관리 분리 (2026-06-10 결정)

**배경**: GitHub (사외) 가 단일 source-of-truth (push-only), 사내 형상관리 툴은 GitHub 에서 read-only pull. 사내 한정 코드/문서/시크릿이 GitHub `main` 으로 push 되면 사내 동기화 시 충돌 또는 사내 한정 정보 노출 위험.

**3-tier 정책** ([`docs/governance/worker_division.md` §6](../docs/governance/worker_division.md) 본문):

| Tier | 의미 | Push 대상 |
|---|---|---|
| **사외** | 사내 인프라 의존 없음. GitHub `main` push. | GitHub (single source-of-truth) |
| **사내** | 사내 호스트/시크릿/사내 IdP 팀 SOP. 사내 SCM 에만 push. | 사내 SCM (GitHub 에서 pull 만) |
| **공용** | 양쪽 byte-identical 유지 필수. drift 시 governance/agent prompt/추적성 ID 회귀. | GitHub (synchronization) |

**PR 작성 시 self-check** (사외 PR 의 경우):
- `DEVHUB_KEYCLOAK_*` / `GITEA_URL` / `HR_EXPORT_CMD` / `internal-registry.example.com` / `kc.internal.example.com` / `devhub.example.com` / `172.16.0.0/12` 등 사내 한정 패턴 누락 여부
- `infrastructure/`, `infra/idp/`, `scripts/setup-keycloak.sh`, `docker-compose.{local,test,deploy,colima}.yml` 등 사내 한정 경로 변경 여부
- `.env.deploy`, `.env.test`, `frontend/.env.example` 의 사내 env var 추가/변경 여부

**PR template 의 Tier 필드** (`.github/pull_request_template.md`): 본 §6.5 도입 예정. 도입 시 본인이 push 하는 PR 의 tier 명시 필수.

**자세한 디렉터리/문서 tier 매핑**: `docs/governance/worker_division.md` §6.3 (SoT).

## 목적

이 저장소에서는 표준 AI 워크플로우를 기준으로 작업한다. 세션 시작, backlog 갱신, 문서 동기화, 세션 종료는 `ai-workflow/` 아래 문서를 우선 기준으로 삼는다.

## 항상 먼저 읽을 문서

1. 현재 git 브랜치를 확인한다: `git branch --show-current`
2. 브랜치별 memory 디렉터리를 우선 읽는다.
   - 브랜치 prefix `codex/` 이면 `ai-workflow/memory/codex/<branch-suffix>/`
   - 브랜치 prefix `claude/` 이면 `ai-workflow/memory/claude/<branch-suffix>/`
   - 브랜치 prefix `gemini/...`이면 `ai-workflow/memory/gemini/<branch-suffix>/`
   - 브랜치 prefix `deepseek/...`이면 `ai-workflow/memory/deepseek/<branch-suffix>/` (Reasonix 포함)
   - 브랜치 prefix `opencode/...`이면 `ai-workflow/memory/opencode/<branch-suffix>/`
   - 브랜치 prefix `mvs/...`이면 `ai-workflow/memory/mvs/<branch-suffix>/`
   - prefix 없는 브랜치는 `ai-workflow/memory/branches/<branch-name>/` 사용
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
- **Tier 분리**: 본인의 PR 이 사외/사내/공용 어느 tier 인지 식별하고, 사내 한정 정보 누락 여부를 self-review 한다 (§사외/사내 2-tier 형상관리 분리 참조).

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
- **문서 tier 라벨**: `docs/` 하위 신규/수정 문서는 `docs/governance/document-standards.md` §2 메타 헤더에 `Tier: 사외 / 사내 / 공용` 필드 명시 필수 (다음 `document-standards.md` 갱신 시 정식 도입)
- **위키 1:1 mirror 정공법 (2026-06-13, Phase 1.5+3 추가)**: 본 저장소 의 mirror scope (Phase 1+1.5+3, ~220 file) — `docs/llm-wiki/mirror-list.md` §1.7 + §1.8 의 위키의 source code + workflow + scripts + branch memory + traceability + **domain/architecture/infrastructure/validation mass ingest** 의 raw mirror 1:1 byte-identical 정공법. **위키만으로 코드 maintenance 가능**. 모든 신규 PR은 mirror scope 갱신 요청 (mirror-list.md §1.7/§1.8 + lint-config.toml + scripts/wiki-sync-devhub.sh 화이트리스트 정합) 필수. **PR 머지 후 `bash scripts/wiki-sync-devhub.sh` 1회 실행** (real mode) 으로 `~/wiki/raw/projects/devhub/` 의 1:1 mirror 갱신. **상세 SOP**: `docs/llm-wiki/operation-sop.md` §0+§9 의 정공법 + 본 저장소 의 main flat memory + branch memory 정합. **위반 시점 drift**: mirror byte-identical 검증 script 의 `Total: ~196, Diff: 0` 미충족 시 즉시 fix. **provenance tracking 정공법**: mirror script 의 manifest 의 commit hash + version + last sync 자동 capture + 위키 page 의 frontmatter 의 git_commit + version_system + version_workflow + last_touched 자동 반영. **status 검증 command**: `bash scripts/wiki-status-check.sh` (4 mode: all / --stale / --diff / --json). **mass ingest 정공법** (Phase 3): `bash scripts/wiki-mass-ingest.sh --apply` (78 file 의 wiki page 자동 생성 + index.md 78 line append).

## 워커 일반 메모 (2026-06-09 전면 갱신)

**2026-06-09 결정 — 워커 분업 전면 취소** (사용자 결정, Claude/Codex 자유 이용 불가). 본 AGENTS.md 의 **이하 모든 워커별 전용 메모는 historical record** 로서만 보존되며, 강제력 없음. 모든 신규 작업은 어느 에이전트로든 자유롭게 진행 가능. 세부:

- **분기 prefix**: 역사적 보존 (`codex/` / `claude/` / `gemini/` / `deepseek/` / `opencode/` / `mvs/` + 자유 prefix). 신규 진입 시 `maintenance/` / `chore/` / `docs/` / `fix/` / `feat/` 등 자유 prefix 허용. 단 식별성을 위해 `<role>/work_<YYMMDD>-<sprint-seq>-<issue-num>-<short-key>` 패턴은 권장 유지.
- **메모리 디렉터리 패턴**: `<agent>/<branch>/` historical 보존, 신규 진입 시 자유 (예: `ai-workflow/memory/maintenance/<branch-suffix>/`).
- **GitHub PR label 의 `worker/<X>`**: historical 분류용으로 유지, 신규 PR 의 강제 부착 없음.
- **유지 정책** (워커 무관): §4.2 ADR supersession 정공법, §5 Owner 권한, 우선순위 P0~P3 — 자세한 사항은 [worker_division.md §0](../docs/governance/worker_division.md) 참조.

### Codex 전용 메모 (Historical)

- Codex 는 프로젝트 루트의 `AGENTS.md` 를 읽으므로, 상세 정책은 본 문서에서 시작하고 세부 운영 기준은 `ai-workflow/` 문서를 참조한다.
- OpenAI 관련 질문이 나오면 OpenAI 문서 MCP 를 우선 사용하는 구성을 권장한다.
- 가능한 경우 메인 에이전트는 조정과 통합에 집중하고, bounded scope 의 읽기/쓰기/검증 작업은 worker 성격의 서브 에이전트로 분리하는 패턴을 권장한다.
- worker 에게는 책임 파일과 종료 조건을 명확히 넘기고, 메인 에이전트에는 핵심 사실과 결과만 다시 모은다.
- `main`/`small` 모델을 함께 운영한다면, 메인 에이전트는 난도 높은 판단과 통합에, worker 는 bounded scope 탐색/초안/검증에 우선 배치하는 편이 효율적이다.
- 신규 프로젝트 기준 초안이다. 프로젝트 고유의 실행 명령과 문서 구조가 정확한지 확인해야 한다.

### Reasonix (deepseek-v4) 전용 메모 (Historical)

- Reasonix 는 Codex 와 동일한 workflow 레이어(`ai-workflow/`)를 따르며, 브랜치별 memory 디렉터리 패턴을 동일하게 사용한다.
- 표준 sprint branch 명명 규칙은 `docs/governance/worker_division.md` §2.5 를 따른다: `deepseek/work_<YYMMDD>-<sprint-seq>-<issue-num>-<short-key>`
- 브랜치 prefix `deepseek/` → `ai-workflow/memory/deepseek/<branch-suffix>/`
- Reasonix 의 기본 모델은 `deepseek-v4-flash`이며, 복잡한 cross-file 리팩토링 시 자동으로 `deepseek-v4-pro` 로 escalation 된다.
- Reasonix 의 기본 도구 세트에는 GitHub MCP, Filesystem MCP, Memory KG, Puppeteer 가 포함되어 있다.
- `run_command` 사용 시 `cd` 가 체인에서 거부되므로, cwd 가 필요한 명령은 `cwd` 인자를 직접 전달하거나 `--prefix`/`-C` 플래그를 사용한다.
- 현재 브랜치가 `main`이 아닐 때는 `ai-workflow/` 메타 레이어를 기본 탐색 범위에 포함하지 말고, workflow 문서 갱신이나 세션 복원 시에만 참조한다.
- 프로젝트 실행 기본값(TODO 항목들)은 아직 설정되지 않았다 (`TODO: ...` 상태). Reasonix 세션 시작 시 `PROJECT_PROFILE.md` §3 기본 명령을 직접 참조하여 실행한다.

### OpenCode (Sisyphus / MiniMax-M3) 전용 메모 (Historical)

- OpenCode 워커는 Codex / Claude / Gemini / Reasonix 와 동일한 workflow 레이어(`ai-workflow/`)를 따르며, 브랜치별 memory 디렉터리 패턴을 동일하게 사용한다.
- 표준 sprint branch 명명 규칙은 `docs/governance/worker_division.md` §2.5 를 따른다: `opencode/work_<YYMMDD>-<sprint-seq>-<issue-num>-<short-key>`
- 브랜치 prefix `opencode/` → `ai-workflow/memory/opencode/<branch-suffix>/`
- OpenCode 의 기본 에이전트 식별자는 **Sisyphus** 이며, 기본 모델은 `MiniMax-M3` 다. 복잡한 cross-file 리팩토링·아키텍처 결정은 Oracle 같은 specialist 호출로 escalate 한다.
- OpenCode 는 메인 에이전트 조정/통합에 집중하고, bounded scope 의 읽기/쓰기/검증 작업은 `explore` / `librarian` / `Sisyphus-Junior` 같은 worker 성격 서브 에이전트에 위임하는 패턴을 권장한다.
- 사용자에게 보이는 작업 보고, handoff, backlog, 사용자 안내 문구는 기본 한국어로 작성한다 (Reasonix 와 동일). 코드/명령어/경로/외부 시스템 명칭은 원문 유지.
- 현재 브랜치가 `main`이 아닐 때는 `ai-workflow/` 메타 레이어를 기본 탐색 범위에 포함하지 말고, workflow 문서 갱신이나 세션 복원 시에만 참조한다.

### Mavis (MiniMax Code / MiniMax-M3) 전용 메모 (Historical)

- 본 저장소에서 Mavis 는 외부 AI 워커와 다른 오케스트레이션 레이어로 동작한다. 단일 진입점: [`ai-workflow/minimax_code_workflow.md`](ai-workflow/minimax_code_workflow.md).
- Mavis 의 day-1 baseline, 작업 라우팅 (self / mavis-team / single-spawn verifier), 메모리 3-layer (project/agent/user), 사용 가능 skills/agents/MCP, hard limits, 인계 SOP 은 위 문서에서 단일 source-of-truth 로 관리한다.
- 표준 sprint branch 명명 규칙은 `docs/governance/worker_division.md` §2.5 를 따른다: `mvs/work_<YYMMDD>-<sprint-seq>-<issue-num>-<short-key>`
- 브랜치 prefix `mvs/` → `ai-workflow/memory/mvs/<branch-suffix>/`
- Mavis 는 5-워커 워크플로우를 대체하지 않는다 — cross-cut 정합 (governance / traceability / release_v0-1_roadmap / worker_division) 이 필요할 때 후속 PR 로만 개입.
- 사용자에게 보이는 작업 보고, handoff, backlog, 사용자 안내 문구는 기본 한국어로 작성한다 (Reasonix/OpenCode 와 동일). 코드/명령어/경로/외부 시스템 명칭은 원문 유지.

