# 워커 분업 — Claude / Codex / Gemini

- 문서 목적: DevHub 의 4 워커 (Claude, Codex, Gemini, OpenCode) 가 영역별 작업을 분담하고 인계하는 규칙. 사용자 1명이 모든 워커를 invoke 하지만, 워커별 강점/이력/스타일이 다름.
- 범위: 워커별 책임 영역, 작업 스타일, 인계 SOP, 충돌 처리.
- 대상 독자: 모든 워커, 사용자.
- 상태: draft
- 최종 수정일: 2026-06-04
- 결정 근거 sprint: `claude/work_260520-f-roadmap`
- 관련 문서: [v1.0 릴리즈 로드맵 §5 분업 매트릭스](../planning/release_v1_roadmap.md), [governance/README](./README.md), [document-standards](./document-standards.md), `AGENTS.md`.

## 1. 영역별 분담

### 1.1 Claude — Backend (Go) + Design (ADR + docs)

**주요 책임**:
- Go Core API (`backend-core/`)
- ADR 발급 + design doc 작성
- 추적성 매트릭스 (`docs/traceability/report.md`) 갱신
- M3+ 백엔드 기능 (HRDB, organization, RBAC, Application, DREQ, Audit event listener)
- 외부 워커의 design 리뷰 (PR review mode)
- workflow memory 관리 (`ai-workflow/memory/`)

**작업 스타일**:
- 큰 단위 design 우선 (현황 파악 → 옵션 비교 → 결정 → 실 구현) Phase 분리
- 4단계 self-review (diff 재검토 → gh pr comment → 보강 commit → squash merge)
- codex 외부 review 후 hotfix PR 즉시 진입
- ADR governance 엄격 준수 (immutable history, supersession 패턴)

**누적 이력 (2026-05-20 기준)**:
- 30+ sprint, 60+ PR
- ADR 0002, 0003, 0004, 0005, 0006, 0007, 0008, 0009, 0010, 0011, 0013, 0014, 0017 (일부 codex 도 발급) + 0019, 0020 발급
- M1 RBAC track + M2 1차 완성 + M3 HRDB/Sign Up + M5 DREQ + M6 External Integration design + ADR-0019 §5.3 전체 design 완결 + ADR-0020 sub-carve A

### 1.2 Codex — Infra (Docker/Nginx/CI) + Security + Build

**주요 책임**:
- Docker packaging (`docker-compose.deploy.yml`, `Dockerfile`, infra/nginx/)
- GitHub Actions workflow (`.github/workflows/ci.yml`)
- Keycloak infra (realm.json, SPI plugin, admin SOP)
- Security review (외부 리뷰 P1/P2 발견 가장 활발)
- Build / packaging hardening
- e2e CI 정합 (`scripts/ci-e2e-sync-check.sh`)

**작업 스타일**:
- 외부 리뷰 우선 (PR 머지 직후 P1/P2 inline 발견)
- Infrastructure-as-Code (compose / nginx config / realm.json)
- 운영 SOP 동반 docs

**누적 이력 (2026-05-20 기준)**:
- 7+ PR — packaging guide / reverse-proxy / Keycloak-only refactor / next-step External Integration backend / E2E fix / memory housekeeping / e2e CI 정합
- 주요 contribution: PR #135 (External Integration concept), PR #139 (backend 1차), PR #166 (reverse proxy 실 구현 ADR-0018), PR #167 (Keycloak-only refactor KC-PR-A..F), PR #201 (Keycloak E2E CI 정합), PR #203 SPI webhook (gemini 가 머지차단, claude 가 인수)
- review cycle: hotfix #1..#12 누적 (codex 외부 리뷰 inline P1/P2 → claude hotfix PR)

### 1.3 Gemini — Frontend + UX + Test fixtures + Design polish

**주요 책임**:
- Next.js frontend (`frontend/app/`, `frontend/components/`, `frontend/lib/`)
- e2e Playwright (`frontend/tests/e2e/`)
- Semantic theme (`frontend/app/globals.css` + tailwind variables)
- Dashboard / modal / FilterBar 재설계
- UI/UX polish + responsive + a11y

**작업 스타일**:
- 큰 frontend redesign sweep
- 다수 modal 일괄 theme 정합
- e2e selector 정합 + flaky fix

**누적 이력 (2026-05-20 기준)**:
- 5+ PR — frontend redesign / dashboard UI + LogoutOverlay / FilterBar 표준화 / dev-requests + audit log redesign / DREQ E2E 안정화 / Keycloak test login + semantic theme (PR #203, claude 가 인수해서 머지)
- 주요 contribution: PR #115 (light theme + dropdown + endpoints), PR #134 (dashboard UI + LogoutOverlay), PR #138 (dashboard rebrand + Applications/Repositories/Projects 현황 페이지 + FilterBar), PR #140 (FilterBar standardize + DestructiveConfirmModal), PR #203 (semantic theme)

### 1.4 OpenCode — 영역 TBD (governance bootstrap 진행 중)

**주요 책임**:
- (영역 미정 — 첫 sprint 종료 후 후속 sprint 의 backlog 로 정의)
- 1차 sprint (`opencode/work_260604-a-opencode-workflow-bootstrap`) 는 governance 부트스트랩 한정 — 본 문서 §1.4 placeholder + `AGENTS.md` 의 "OpenCode 전용 메모" 섹션 + `MEMORY_GOVERNANCE.md` prefix 예시 + `WORKFLOW_INDEX.md` 진입 예시 정합

**작업 스타일**:
- 메인 에이전트 조정/통합 중심 (Sisyphus 정체성)
- bounded scope 의 읽기/쓰기/검증은 `explore` / `librarian` / `Sisyphus-Junior` 같은 worker 성격 서브에이전트로 위임
- 복잡한 cross-file 리팩토링·아키텍처 결정은 Oracle 호출로 escalate

**누적 이력 (2026-06-04 기준)**:
- 0 PR (bootstrap sprint)
- 결정: 첫 sprint 는 governance 한정. 영역 분담 + 인계 SOP 은 §1.4 본문 채우기 sprint 에서 확정

## 2. v1.0 sprint 별 분담

[release_v1_roadmap.md §5.2](../planning/release_v1_roadmap.md) 참조. 요약:

| sprint | 작업 | 주도 워커 | 보조 워커 |
| --- | --- | --- | --- |
| -f | sub-carve B (`/api/v1/accounts/*` 폐기) | Claude (backend) | Gemini (frontend cleanup) |
| -g | Playwright screenshot mode | Codex (CI artifact) | Gemini (fixture) |
| -g/-h | UI polish 1차 | **Gemini** | — |
| -i | sub-carve C event listener 확장 | Claude (backend) | — |
| -j | sub-carve D + E | Claude (D) + Codex (E SOP) | — |
| -k | v1.0 종합 검증 | 전 워커 | — |

## 2.5 Branch 명명 규칙 (2026-05-20 신규)

작업 식별성 + 추적성 강화를 위해 모든 신규 sprint branch 는 다음 패턴 따른다:

```
<worker>/work_<YYMMDD>-<sprint-seq>-<issue-num>-<short-key>
```

| 요소 | 설명 | 예시 |
| --- | --- | --- |
| `<worker>` | 워커 prefix | `claude`, `codex`, `gemini`, `deepseek` (Reasonix), `opencode` (Sisyphus) |
| `<YYMMDD>` | 작업 시작 날짜 (KST) | `260520` |
| `<sprint-seq>` | 알파벳 sequence (당일 본 워커의 N번째 sprint) | `a`, `b`, ..., `z`, `aa`, `ab`, ... |
| `<issue-num>` | GitHub issue 번호 (해당 sprint 의 핵심 작업) | `209`, `238` |
| `<short-key>` | 키워드 식별자 (소문자 + 하이픈) | `accounts-deprecation`, `docker-single-port`, `screenshot` |

### 예시

| Branch | 의미 |
| --- | --- |
| `claude/work_260520-i-209-accounts-deprecation` | Claude 2026-05-20 의 i번째 sprint, issue #209 (sub-carve B backend) |
| `codex/work_260520-a-238-docker-single-port` | Codex 2026-05-20 의 a번째 sprint, issue #238 |
| `gemini/work_260521-a-210-ui-polish` | Gemini 2026-05-21 의 a번째 sprint, issue #210 |
| `deepseek/work_260601-a-construct-workflow` | DeepSeek (Reasonix) 2026-06-01 의 a번째 sprint, Reasonix 워크플로우 구성 |
| `opencode/work_260604-a-opencode-workflow-bootstrap` | OpenCode (Sisyphus) 2026-06-04 의 a번째 sprint, governance 부트스트랩 (이슈 없는 housekeeping 예외) |

### 예외 — issue 없는 작업

- housekeeping / memory sync / docs hotfix 같이 GitHub issue 없는 작업: issue 번호 생략 + key 만 명시. 예: `claude/work_260520-c-housekeeping`, `claude/work_260520-g-codex-hotfix`
- codex/gemini 외부 contribution: 자유 branch 명 허용 (예: 본 사례의 `gemini/keycloak-test-e2e-push-audit`). 본인 인수 시 그대로 작업 후 PR 머지.
- Reasonix 환경에서 브랜치 생성 시 반드시 `deepseek/` prefix 를 사용한다. (Reasonix 는 `AGENTS.md` "Reasonix 전용 메모" 및 "항상 먼저 읽을 문서" 섹션의 deepseek 패턴을 따른다.)
- OpenCode 환경에서 브랜치 생성 시 반드시 `opencode/` prefix 를 사용한다. (OpenCode 는 `AGENTS.md` "OpenCode 전용 메모" 및 "항상 먼저 읽을 문서" 섹션의 opencode 패턴을 따른다.)

### 적용 시점

- **2026-05-20 sprint -i 이후 모든 신규 sprint** 에 본 규칙 적용
- 이전 branch (예: `claude/work_260519-ad`, `claude/work_260520-h-screenshot`) 는 historical 보존 — rename 금지

### 효과

- branch 만 봐도 어떤 issue / 어떤 작업인지 즉시 식별
- GitHub project 와 1:1 매핑 (issue # → branch → PR)
- 다중 워커 협업 시 충돌 영역 즉시 인지

## 3. 인계 SOP

### 3.1 Claude → Codex

**예시 시나리오**: Claude 가 ADR-0019 §5.3 (9) audit event listener design 작성 → Codex 가 (a) Keycloak admin event polling SOP 운영 자산, (b) infra/ 자산 (Keycloak realm 설정 export) 실 구현.

**인계 자산**:
- design doc (예: `docs/planning/keycloak_event_audit_integration.md`)
- 작업 범위 명시 (어떤 파일 / 어떤 SOP / 어떤 운영 작업)
- 검증 기준 (Prometheus metric / Grafana panel / 운영 SOP 항목)

**인계 형식**:
1. Claude 가 design doc + ADR 발급 PR 머지
2. issue 생성 — `worker/codex` label + 작업 범위 본문
3. Codex 가 issue claim + PR 발급

### 3.2 Claude → Gemini

**예시 시나리오**: Claude 가 `/api/v1/accounts/*` 4 endpoint 제거 PR 머지 (backend) → Gemini 가 frontend `account.service.ts` 폐기 + admin/settings/users page 정리 + e2e TC-ACC-* 갱신.

**인계 자산**:
- API spec (제거된 endpoint + 대체 흐름)
- 응답 schema (변경된 경우)
- 영향 받는 page + service + test 명시

**인계 형식**:
1. Claude 가 backend PR 머지 + frontend 영향 영역 issue 생성
2. issue label `worker/gemini` + `domain/<area>` + `type/refactor`
3. Gemini 가 issue claim + PR 발급
4. Gemini 가 backend API 누락/부정합 발견 시 PR review comment + 신규 issue

### 3.3 Codex → Claude

**예시 시나리오**: Codex 가 PR review 에서 P1 발견 (예: PR #205 의 `team_manager` 누락 회귀) → Claude 가 hotfix PR.

**인계 자산**:
- codex review comment URL
- P1/P2 마킹 + finding 본문
- 권장 fix 옵션 (codex 가 보통 1~3 옵션 제시)

**인계 형식**:
1. Codex 가 GitHub PR comment 로 review 게시 (inline + P1/P2 badge)
2. Claude 가 review 인지 → 옵션 분석 + 사용자 confirm 필요 시 question → fix PR 진입
3. fix PR 의 commit message 에 codex review URL 인용 + P1/P2 응답 명시

### 3.4 Gemini → Claude

**예시 시나리오**: Gemini 가 admin/settings/users 정리 중 backend GET /api/v1/users 의 response schema 가 frontend type 과 부정합 발견 → Claude 가 backend handler 갱신.

**인계 자산**:
- frontend 가 기대하는 schema (type 정의)
- backend 의 실제 response (curl 또는 spec inspect)
- 차이점 정리

**인계 형식**:
1. Gemini 가 PR review comment 또는 신규 issue
2. label `worker/claude` + `type/refactor` + `domain/<area>`
3. Claude 가 backend 진입 + 검증

## 4. 충돌 처리

### 4.1 같은 파일 동시 작업

**원칙**: 영역별 분담 따르면 같은 파일 동시 작업 거의 없음. 발생 시:
- `router.go` / `permissions.go` / `state.json` 같은 cross-cut 파일이 가장 잦음
- 후착 워커가 rebase 책임 + 충돌 해소

**SOP**:
1. 후착 워커가 main rebase
2. 충돌 영역 분석 + 양 쪽 의도 보존
3. 본인 fix commit + force-push (lease) + PR review comment 로 충돌 해소 사실 명시

### 4.2 ADR 결정 reversal

immutable history 패턴 따름 — 본문 partial 수정 **금지**, 새 ADR 분리 발행 + supersede 대상 ADR 메타 헤더 갱신 + inline supersession banner. 표준 절차:

1. **새 ADR 발행** — `docs/adr/NNNN-<topic>.md` 신규. 메타 헤더에 `supersedes: [ADR-XXXX](./XXXX-*.md)` 명시
2. **supersede 대상 ADR 갱신** — 메타 헤더 `상태: superseded by [ADR-NNNN](./NNNN-*.md)` + §0 / 각 § heading 에 inline supersession banner 추가
3. **본문 immutable 보존** — supersede 된 ADR 의 본문은 historical context 로 그대로 유지
4. **traceability §4 row** — 매트릭스 ADR 인덱스에 supersession 관계 명시
5. **관련 문서 정합 (≥5개 docs)** — architecture / requirements / api_contract / setup / planning 등에서 supersede 된 ADR 참조 위치를 새 ADR 로 redirect

**canonical 사례**: [ADR-0019 Keycloak 단일화](../adr/0019-keycloak-only-idp.md) 가 [ADR-0001 Hydra+Kratos](../adr/0001-idp-selection.md) supersession (sprint `claude/work_260519-a`, PR #169). 직전 PR #167 (codex) 가 ADR-0001 본문 partial 수정 → sprint -a 가 ADR-0019 발행으로 정공법 정정 + 14 추가 정합 docs.

### 4.3 우선순위 충돌

P0 > P1 > P2 > P3 강제. P0 carve 진행 중 P2 carve 진입 금지 (예외: 같은 워커가 idle 상태일 때만).

## 5. 사용자 (Owner) 의 역할

- **invoke 책임**: 모든 워커는 사용자가 invoke. 자동 트리거 없음
- **사내 동반 carve**: `worker/user` label 항목은 사내 인프라/운영팀 동반 작업 (Keycloak admin 작업, HRDB ETL deploy, HA Phase 2 등)
- **결정 권한**: ADR 결정 + 우선순위 변경 + 마일스톤 변경
- **review 최종 승인**: 모든 PR 의 squash merge 권한

## 6. 변경 이력

| 일자 | 변경 | sprint |
| --- | --- | --- |
| 2026-05-20 | 1차 작성 — Claude (backend+design) / Codex (infra+CI+security) / Gemini (frontend+UX) 분담 + 인계 SOP 4 패턴 + 충돌 처리 SOP + 사용자 역할 명시 | `claude/work_260520-f-roadmap` |
| 2026-05-20 | codex review hotfix (P2) — §4.2 ADR reversal 의 dead link (`[feedback_adr_supersession_pattern.md](#)`) 정정. per-user auto-memory 파일이라 repo 에 없음 → 5 step 표준 절차 본문 명시 + canonical 사례 ADR-0019/ADR-0001 supersession (PR #169) 인용으로 대체 | `claude/work_260520-g-codex-hotfix` |
| 2026-05-20 | §2.5 신규 — Branch 명명 규칙 (`<worker>/work_<YYMMDD>-<sprint-seq>-<issue-num>-<short-key>`). 사용자 지시 (2026-05-20) 따라 작업 식별성 강화. 예외 (issue 없는 housekeeping/hotfix / 외부 contribution) 명시. 2026-05-20 sprint -i 이후 적용, 이전 branch 는 historical 보존 | `claude/work_260520-i-209-accounts-deprecation` (본 sprint 가 적용 첫 사례) |
| 2026-06-01 | §2.5 `<worker>` 목록에 `deepseek` (Reasonix) 추가 + 예시 row + 예외에 Reasonix 환경 deepseek/ prefix 규칙 명시 | `deepseek/construct_workflow_for_deepseek` |
| 2026-06-04 | **§1.4 OpenCode 신설** — 영역 TBD placeholder + bootstrap 노트 / §2.5 `<worker>` 목록에 `opencode` (Sisyphus) 추가 + 예시 row + 예외에 OpenCode 환경 opencode/ prefix 규칙 명시 / 문서 헤더 워커 수 3 → 4 갱신 | `opencode/work_260604-a-opencode-workflow-bootstrap` |
