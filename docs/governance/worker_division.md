# 워커 분업 — **취소 (2026-06-09)**

- 문서 목적: **본 문서는 2026-06-09 사용자 결정으로 워커 분업 정책이 전면 취소되었음을 명시한다.** 이전 분업 (Claude/Codex/Gemini/OpenCode) 표기는 historical record 로만 보존.
- 범위: 취소 결정 + 자유 에이전트 이용 정책 + 잔여 운영 결정.
- 대상 독자: 모든 contributor (사람 + AI agent).
- 상태: **superseded (2026-06-09 사용자 결정)**
- 최종 수정일: 2026-06-09
- 결정 근거: 사용자 (Owner) 의 자유 에이전트 이용 결정 — Claude/Codex 의 자유로운 이용 불가로 분배 무효화
- 관련 문서: [v1.0 릴리즈 로드맵 §5](../../planning/release_v1_roadmap.md) (역시 워커 분담 표 갱신 대상), [AGENTS.md](../../../AGENTS.md) (워크플로우 진입점), `ai-workflow/MEMORY_GOVERNANCE.md`.

## 0. 2026-06-09 결정 — 워커 분업 전면 취소

**사용자 (Owner) 결정**: Claude 및 Codex 의 자유로운 이용이 불가한 상황이므로, 본 문서 (§1~§5) + [`AGENTS.md`](../../../AGENTS.md) + [`docs/planning/release_v1_roadmap.md` §5](../../planning/release_v1_roadmap.md) 의 **워커별 영역 분담 / sprint 별 분담 / 인계 SOP / 충돌 처리 SOP** 를 **모두 취소**한다.

**취소 적용 범위**:
- §1.1 Claude — Backend (Go) + Design: **무효**
- §1.2 Codex — Infra + Security + Build: **무효**
- §1.3 Gemini — Frontend + UX + Test: **무효**
- §1.4 OpenCode (Sisyphus) — Workflow curation + Cross-cutting validation + AI/ML prep: **무효**
- §2 v1.0 sprint 별 분담 (Claude/Codex/Gemini 주도 워커 칼럼): **무효**
- §3 인계 SOP 4 패턴 (Claude→Codex / Claude→Gemini / Codex→Claude / Gemini→Claude): **무효**
- §4 충돌 처리 (§4.1 같은 파일 동시 작업 / §4.2 ADR reversal / §4.3 우선순위 충돌): **무효**
- §5 사용자 (Owner) 의 invoke 책임 / 결정 권한 / review 최종 승인: **유지** (Owner 결정 권한 + PR 머지 권한)

**유지되는 정책 (영역 무관)**:
- §4.2 ADR 결정 reversal (immutable history + supersession): **유지** — 이건 워커 분담과 무관한 문서 governance 의 정공법
- §2.5 Branch 명명 규칙의 `<worker>` prefix 부분: **취소** — `maintenance/` / `chore/` / `docs/` 등 자유 prefix 허용. 단, 식별성을 위해 `<role>/work_<YYMMDD>-<sprint-seq>-<issue-num>-<short-key>` 패턴은 권장 유지
- 모든 워커가 보존하는 history (이전 sprint 의 `<worker>/` prefix branch): **보존** (rename 금지)
- AGENTS.md / `ai-workflow/MEMORY_GOVERNANCE.md` 의 브랜치별 memory 디렉터리 패턴 (`ai-workflow/memory/<agent>/<branch>/`): **유지** — historical 보존, 신규 진입 시 자유
- §6.3 release_v1_roadmap.md 의 `worker/<X>` label: **유지** (historical 분류용), 단 신규 PR 의 라벨 부착 강제 없음

**취소 결정의 운영 효과**:
- 모든 신규 sprint / PR / 작업은 **어느 에이전트로든 자유롭게** 진행 가능
- 영역 분담 / 인계 SOP / 충돌 처리 SOP 의 강제력 소멸
- 사용자 (Owner) 가 단일 워커로 전체 작업 가능 (sprint -a/-b 연속 PR 도 1 워커가 흡수 가능)
- `main` PR 머지 권한 + ADR 결정 권한 + 우선순위 변경 권한은 **여전히 Owner 만 보유**

## 1. (Historical) 영역별 분담 — 2026-05-20 ~ 2026-06-08

> 본 절은 2026-05-20 ~ 2026-06-08 사이 운영된 워커 분업의 historical record 다. 2026-06-09 결정으로 **무효**가 되었으나, 이전 sprint 의 PR 이력 추적 + 신규 인원의 onboarding reference 용으로 보존한다.

### 1.1 (Historical) Claude — Backend (Go) + Design (ADR + docs)

**주요 책임** (historical):
- Go Core API (`backend-core/`)
- ADR 발급 + design doc 작성
- 추적성 매트릭스 (`docs/traceability/report.md`) 갱신
- M3+ 백엔드 기능 (HRDB, organization, RBAC, Application, DREQ, Audit event listener)
- 외부 워커의 design 리뷰 (PR review mode)
- workflow memory 관리 (`ai-workflow/memory/`)

**작업 스타일** (historical):
- 큰 단위 design 우선 (현황 파악 → 옵션 비교 → 결정 → 실 구현) Phase 분리
- 4단계 self-review (diff 재검토 → gh pr comment → 보강 commit → squash merge)
- codex 외부 review 후 hotfix PR 즉시 진입
- ADR governance 엄격 준수 (immutable history, supersession 패턴)

**누적 이력 (2026-05-20 ~ 2026-06-08, 30+ sprint, 60+ PR)**: M1 RBAC track + M2 1차 완성 + M3 HRDB/Sign Up + M5 DREQ + M6 External Integration design + ADR-0019 §5.3 전체 design 완결 + ADR-0020 sub-carve A 등.

### 1.2 (Historical) Codex — Infra (Docker/Nginx/CI) + Security + Build

**주요 책임** (historical):
- Docker packaging (`docker-compose.deploy.yml`, `Dockerfile`, infra/nginx/)
- GitHub Actions workflow (`.github/workflows/ci.yml`)
- Keycloak infra (realm.json, SPI plugin, admin SOP)
- Security review (외부 리뷰 P1/P2 발견 가장 활발)
- Build / packaging hardening
- e2e CI 정합 (`scripts/ci-e2e-sync-check.sh`)

**누적 이력 (2026-05-20 ~ 2026-06-08, 7+ PR)**: PR #135 (External Integration concept), PR #139 (backend 1차), PR #166 (reverse proxy 실 구현 ADR-0018), PR #167 (Keycloak-only refactor KC-PR-A..F), PR #201 (Keycloak E2E CI 정합), PR #203 SPI webhook 등. review cycle: hotfix #1..#12 누적 (codex 외부 리뷰 inline P1/P2 → claude hotfix PR).

### 1.3 (Historical) Gemini — Frontend + UX + Test fixtures + Design polish

**주요 책임** (historical):
- Next.js frontend (`frontend/app/`, `frontend/components/`, `frontend/lib/`)
- e2e Playwright (`frontend/tests/e2e/`)
- Semantic theme (`frontend/app/globals.css` + tailwind variables)
- Dashboard / modal / FilterBar 재설계
- UI/UX polish + responsive + a11y

**누적 이력 (2026-05-20 ~ 2026-06-08, 5+ PR)**: PR #115 (light theme + dropdown + endpoints), PR #134 (dashboard UI + LogoutOverlay), PR #138 (dashboard rebrand + Applications/Repositories/Projects 현황 페이지 + FilterBar), PR #140 (FilterBar standardize + DestructiveConfirmModal), PR #203 (semantic theme) 등.

### 1.4 (Historical) OpenCode (Sisyphus / MiniMax-M3) — Workflow curation + Cross-cutting validation + AI/ML prep

**정체성** (historical): 메인 에이전트 조정/통합 specialist (Sisyphus).

**3-lane 영역** (historical, 2026-06-04 확정):
1. **Workflow / governance curation** (1순위) — `ai-workflow/` 메타 + `docs/governance/` cross-cut 정합
2. **Cross-cutting validation & test infrastructure** (2순위) — multi-file 회귀 검증
3. **AI/ML service prep** (3순위, v1.1/v2) — `backend-ai/` Python + gRPC

**누적 이력 (2026-06-04 ~ 2026-06-08)**: bootstrap sprint `opencode/work_260604-a-opencode-workflow-bootstrap` + areas 정의 `opencode/work_260604-b-opencode-areas` + N-10 Manager RBAC 검증 보고서 (`opencode/work_260604-c-N10-manager-rbac-validation`) + N-11 CI 복원 운영 정합 (sprint 260608-a, PR #498 + 260608-b PR #499, 2026-06-08).

## 2. (Historical) v1.0 sprint 별 분담

> 2026-05-20 ~ 2026-06-08 운영. 2026-06-09 결정으로 **무효**. 실제 분담 이력은 [release_v1_roadmap.md §9](../../planning/release_v1_roadmap.md) 변경 이력 참조.

## 2.5 (Historical) Branch 명명 규칙

**2026-05-20 ~ 2026-06-08 운영 규칙** (이후 **취소**):

```
<worker>/work_<YYMMDD>-<sprint-seq>-<issue-num>-<short-key>
```

- `<worker>` prefix: `claude` / `codex` / `gemini` / `deepseek` (Reasonix) / `opencode` (Sisyphus) / `mvs` (Mavis)

**2026-06-09 결정 — Branch 명명 자유화**:
- `<worker>` prefix: **무효** (강제력 없음)
- 권장 패턴 유지: `<role>/work_<YYMMDD>-<sprint-seq>-<issue-num>-<short-key>` — 예: `maintenance/work_260609-a-cancel-worker-division` (본 PR), `chore/...`, `docs/...`, `fix/...`, `feat/...`
- 식별성 (`<short-key>`) + 날짜 (`<YYMMDD>`) + 이슈 번호 (`<issue-num>`) 는 권장 유지
- 이전 sprint 의 `<worker>/` prefix branch (예: `claude/work_260519-ad`): **보존** (rename 금지)
- 신규 진입: **자유** prefix 사용 가능

## 3. (Historical) 인계 SOP

> 2026-05-20 ~ 2026-06-08 운영. 2026-06-09 결정으로 **무효**.

본래 4 패턴 (Claude→Codex / Claude→Gemini / Codex→Claude / Gemini→Claude) + §4 충돌 처리 SOP 가 정의됐으나, 워커 분업 취소로 인계 SOP 자체가 무효. 다중 워커 협업 시의 일반 원칙:

- 같은 파일 동시 작업 발생 시 **후착자가 rebase + conflict 해결** 책임 (운영 원칙)
- **ADR 결정 reversal 의 immutable history + supersession 패턴은 유지** (§4.2 의 governance 정공법)
- 우선순위 충돌 (P0 > P1 > P2 > P3): 유지 (release_v1_roadmap.md §0.1)

## 4. (Historical) 충돌 처리

### 4.1 같은 파일 동시 작업 (Historical)

- 후착 워커가 main rebase
- 충돌 영역 분석 + 양 쪽 의도 보존
- fix commit + force-push (lease) + PR review comment 로 충돌 해소 명시

### 4.2 ADR 결정 reversal — **유지 (governance 정공법)**

immutable history 패턴 따름 — 본문 partial 수정 **금지**, 새 ADR 분리 발행 + supersede 대상 ADR 메타 헤더 갱신 + inline supersession banner. 표준 절차 (변경 없음):

1. **새 ADR 발행** — `docs/adr/NNNN-<topic>.md` 신규. 메타 헤더에 `supersedes: [ADR-XXXX](./XXXX-*.md)` 명시
2. **supersede 대상 ADR 갱신** — 메타 헤더 `상태: superseded by [ADR-NNNN](./NNNN-*.md)` + §0 / 각 § heading 에 inline supersession banner 추가
3. **본문 immutable 보존** — supersede 된 ADR 의 본문은 historical context 로 그대로 유지
4. **traceability §4 row** — 매트릭스 ADR 인덱스에 supersession 관계 명시
5. **관련 문서 정합 (≥5개 docs)** — architecture / requirements / api_contract / setup / planning 등에서 supersede 된 ADR 참조 위치를 새 ADR 로 redirect

canonical 사례: [ADR-0019 Keycloak 단일화](../../adr/0019-keycloak-only-idp.md) 가 [ADR-0001 Hydra+Kratos](../../adr/0001-idp-selection.md) supersession (sprint `claude/work_260519-a`, PR #169).

### 4.3 우선순위 충돌 (Historical)

P0 > P1 > P2 > P3 강제 (governance 정합). P0 carve 진행 중 P2 carve 진입 금지.

## 5. 사용자 (Owner) 의 역할 — **유지**

- **invoke 책임**: 워커는 사용자가 invoke. 자동 트리거 없음
- **사내 동반 carve**: `worker/user` label 항목은 사내 인프라/운영팀 동반 작업 (Keycloak admin, HRDB ETL deploy, HA Phase 2 등)
- **결정 권한**: ADR 결정 + 우선순위 변경 + 마일스톤 변경 + 워커 분업 변경 (본 결정 2026-06-09 의 예)
- **review 최종 승인**: 모든 PR 의 squash merge 권한

## 6. 변경 이력

| 일자 | 변경 | sprint |
| --- | --- | --- |
| 2026-06-09 | **§0 신규 + §1~§4 전면 改 編 (historical 標記) + §2.5 branch prefix 자유화 + §5 Owner 권한 명시 (워커 분업 전면 취소)** — 사용자 결정 (Claude/Codex 자유 이용 불가) 으로 §1.1~1.4 의 영역별 분담 / §2 sprint 별 분담 / §3 인계 SOP 4 패턴 / §4.1·4.3 충돌 처리 SOP 모두 무효. §4.2 ADR reversal 의 supersession 정공법 + §5 Owner 의 결정 권한은 유지. branch `<worker>` prefix 강제 해제, 권장 패턴만 유지 | `maintenance/work_260609-a-cancel-worker-division` |
| 2026-05-20 | 1차 작성 — Claude (backend+design) / Codex (infra+CI+security) / Gemini (frontend+UX) 분담 + 인계 SOP 4 패턴 + 충돌 처리 SOP + 사용자 역할 명시 | `claude/work_260520-f-roadmap` |
| 2026-05-20 | codex review hotfix (P2) — §4.2 ADR reversal 의 dead link 정정 + 5 step 표준 절차 본문 명시 + canonical 사례 ADR-0019/ADR-0001 supersession 인용 | `claude/work_260520-g-codex-hotfix` |
| 2026-05-20 | §2.5 신규 — Branch 명명 규칙 (`<worker>/work_<YYMMDD>-<sprint-seq>-<issue-num>-<short-key>`) | `claude/work_260520-i-209-accounts-deprecation` |
| 2026-06-01 | §2.5 `<worker>` 목록에 `deepseek` (Reasonix) 추가 | `deepseek/construct_workflow_for_deepseek` |
| 2026-06-04 | **§1.4 OpenCode 신설** — 영역 TBD placeholder + bootstrap 노트 / §2.5 `<worker>` 목록에 `opencode` 추가 | `opencode/work_260604-a-opencode-workflow-bootstrap` |
| 2026-06-04 | **§1.4 OpenCode 본문 정의** — 3-lane 확정 + release_v1_roadmap §5.1 정합 | `opencode/work_260604-b-opencode-areas` |
| 2026-06-04 | **Mavis (MiniMax Code) 오케스트레이션 레이어 신설** — `ai-workflow/minimax_code_workflow.md` 신규 | `mvs/work_260604-a-minimax-code-workflow-setup` |
