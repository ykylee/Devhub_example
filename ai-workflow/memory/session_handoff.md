# Session Handoff — main (2026-05-15 post-EOD final + docker packaging merge)

- 문서 목적: main 브랜치 기준 세션 상태와 다음 작업 진입점을 인계한다.
- 범위: 2026-05-15 직전 final EOD (sprint l) 이후의 후속 세션 종료. 추가 5 PR 흡수 (#128 m / #129 n / #130 o / #131 p / 본 q housekeeping). **DREQ carve out 1/4 + 2/4 전체 완료** (RBAC-ADR / Promote-Tx / codex hotfix #4 / Admin-UI backend / Admin-UI frontend).
- 대상 독자: 후속 에이전트, 프로젝트 리드, 다음 세션 진입자.
- 상태: M1/M2/M3 1차 closing (이전). Application 도메인 backend 1차 (2026-05-14). DREQ 도메인 1차 완성 (sprint l 종료, 2026-05-15). 본 후속 세션 — **DREQ carve out 1/4 (RBAC-ADR + Promote-Tx) + 2/4 (Admin-UI backend + frontend) 완료**. 4 sprint (m/n/o/p), 본인 4단계 리뷰 4회 clean, codex review cycle 1회 (hotfix #4). ADR-0013 + ADR-0014 누적. API-66..68 + 신규 RBAC resource `dev_request_intake_tokens` activated. /admin/settings/dev-request-tokens 페이지.
- 최종 수정일: 2026-05-15 (PR #133 merge 반영)
- 관련 문서: [통합 로드맵](../../docs/development_roadmap.md), [상태 스냅샷](./state.json), [거버넌스](../../docs/governance/README.md), [추적성 매트릭스](../../docs/traceability/report.md), [Project 도메인 컨셉](../../docs/planning/project_management_concept.md), [Dev Request 도메인 컨셉](../../docs/planning/development_request_concept.md), [ADR-0011 RBAC row-scoping](../../docs/adr/0011-rbac-row-scoping.md), [ADR-0012 DREQ 외부 수신 인증](../../docs/adr/0012-dreq-external-intake-auth.md), [ADR-0013 DREQ RBAC row-scoping](../../docs/adr/0013-dreq-rbac-row-scoping.md), [ADR-0014 DREQ intake token admin](../../docs/adr/0014-dreq-intake-token-admin.md).
- 브랜치: `main` (HEAD `4892a78`, PR #133 squash merge 반영).

## 본 후속 세션 (2026-05-15 post-EOD) 누적 머지 — 7 PR

| PR | sha | sprint | 작업 |
| --- | --- | --- | --- |
| #128 | 1f9ec50 | claude/work_260515-m | DREQ-Promote-Tx 단일 트랜잭션 + ADR-0013 RBAC row-scoping |
| #129 | 5546a41 | claude/work_260515-n | codex review hotfix #4 — PR #128 의 P1 (CHECK 매핑) + P2 (SCM gate) + self-review P2 #1 (rejected_reason NULL) |
| #130 | 0bdf299 | claude/work_260515-o | DREQ-Admin-UI backend — intake token admin (API-66..68) + ADR-0014 |
| #131 | 2147d6d | claude/work_260515-p | DREQ-Admin-UI frontend — /admin/settings/dev-request-tokens 페이지 + plain-1회 modal |
| #132 | 253063e | claude/work_260515-q | post-EOD housekeeping (main flat memory sync + sprint finalize) |
| #133 | 4892a78 | codex/docker-packaging-guide | Docker deploy 패키지 안정화 + runtime-config/OIDC/토큰 모달/권한 UI 리뷰 반영 + 보안 placeholder 강화 |

## 본 후속 세션 도입 핵심 (재참조 가능)

### 1. DREQ carve out 1/4 — RBAC-ADR + Promote-Tx (sprint m, PR #128)

- **ADR-0013** — `dev_requests` resource 의 RBAC row-scoping 정책 사후 명문화. ADR-0011 §4.2 helper `enforceRowOwnership(c, dr.AssigneeUserID, "pmo_manager")` 의 dev_requests 적용 사례. handler wire-up 은 PR #124 (sprint i) 에서 이미 도입.
- **Promote-Tx**: `store.RegisterDevRequestWithNewApplication` + `RegisterDevRequestWithNewProject` 신규 — `pool.BeginTx` → INSERT target (+ optional `application_repositories`) → UPDATE dev_requests → Commit. **REQ-FR-DREQ-005 정합 완성**.
- handler 분기: `target_id` (legacy) / `application_payload` / `project_payload` mutual exclusion.
- SQL drift 방지: `applications.go` 의 INSERT SQL 들을 const 로 추출.

### 2. codex hotfix #4 (sprint n, PR #129)

- **P1**: primary_repo 분기의 `application_repositories_role_check` CHECK 위반이 500 으로 surface → handler `validApplicationRepoRoles` gate + store `isCheckViolation` defense-in-depth.
- **P2**: promote primary_repo path 의 SCM provider enablement gate 우회 → handler `ListSCMProviders` + Enabled 검증.
- **self-review P2 #1**: `MarkDevRequestRegistered` + `dreqMarkRegisteredUpdateQuery` 에 `rejected_reason = NULL` clear 추가.
- 5 회귀 가드 test.

### 3. DREQ carve out 2/4 Admin-UI backend (sprint o, PR #130)

- **ADR-0014** — dev_request_intake_tokens resource RBAC + plain-1회-노출 + idempotent revoke 정책. accounts_admin temp_password 패턴과 정합.
- **신규 RBAC resource** `dev_request_intake_tokens` (system_admin 일임). migration 000026.
- **신규 endpoint 3**: API-66 POST (발급) / API-67 GET (목록) / API-68 DELETE (revoke). server 32-byte base64url plain 생성 → SHA-256(hex) 저장 + audit 에 plain/hashed 둘 다 미포함 + revoke `COALESCE(revoked_at, NOW())`.
- 8 신규 unit test.

### 4. DREQ carve out 2/4 Admin-UI frontend (sprint p, PR #131)

- **/admin/settings/dev-request-tokens** 페이지 (system_admin 보호 via AuthGuard + `/admin/*` prefix + layout subTabs 에 Intake Tokens 추가).
- **IssueIntakeTokenModal**: 2 phase (form → reveal). reveal phase 의 outside-click + ESC **차단** — 실수로 plain token 분실 방지. clipboard API copy.
- **IntakeTokenTable**: client/source + allowed_ips chips + Active/Revoked badge + revoke action.
- `dev_request_token.{service,types}.ts` — thin wrapper.
- npm run build PASS (26 static pages) + vitest 41 tests PASS.

## 다음 세션 directive

**DREQ carve out 3/4 — DREQ-E2E** (sprint q' 또는 다음 세션 진입).

| 작업 | scope |
| --- | --- |
| Playwright spec | intake auth (token bearer) → admin issue token → dashboard widget (assignee 본인 의뢰) → promote (신규 application/project 생성 단일 tx) → revoke token. TC-DREQ-* 발급 |
| Vitest unit | IntakeTokenTable / IssueIntakeTokenModal 두 phase (form / reveal) + clipboard mock + outside-click 차단 검증 |
| P2 carve out 흡수 후보 (누적 6건) | (1) promote-tx race 가드 (UPDATE WHERE status IN ...) — sprint m P2 #2; (2) memoryDevRequestStore.failPromote dead field — sprint m P2 #3; (3) window.confirm 대신 destructive confirm dialog — sprint p P2 #2; (4) plain_token reveal Show/Hide toggle — sprint p P2 #3; (5) token rotation policy (expires_at + cron) — ADR-0014 §6; (6) allowed_ips mutation endpoint — ADR-0014 §6 |

본 4건 carve out 중 1/4 + 2/4 완료. 3/4 만 남음 (4/4 는 본 sprint plan 에서 carve 2 가 2 개 sprint 로 분할되며 자연 흡수, 별도 1 carve 가 아니라 묶음의 일부).

## 다음 세션 directive (사용자 지시)

> "다음 작업 사항 모두 묶어서 진행할 거야" — DREQ carve out 4건을 한 sprint plan 으로 진입.

| Carve | 의존 | scope |
| --- | --- | --- |
| **DREQ-RBAC-ADR** | (독립) | pmo_manager 위양 정책 ADR — ADR-0011 §4.2 패턴 따라 dev_requests resource 의 row-level 위양 명문화 |
| **DREQ-Promote-Tx** | backend (Promote-Tx ↔ Admin-UI 둘은 backend 가 의존하지 않음) | store.RegisterDevRequest 가 신규 application/project 생성 + dev_request 상태 갱신 단일 트랜잭션 — REQ-FR-DREQ-005 정합 완성. 현재 handler 는 기존 target_id 매핑만 |
| **DREQ-Admin-UI** | backend admin endpoint 신설 → frontend UI | intake token 발급/revoke endpoint (`POST /api/v1/dev-request-tokens` 등) + `/admin/settings/dev-request-tokens` 페이지. accounts_admin 의 password issuance 패턴 (plain 1회 노출) 따름 |
| **DREQ-E2E** | 다른 3건 완료 후 | Playwright spec (intake → dashboard widget → register → close 흐름) + Vitest unit. TC-DREQ-* 발급. |

권장 진입 순서: **RBAC-ADR + Promote-Tx 병행 → Admin-UI → E2E**. 한 sprint 묶음 또는 4개 PR 로 나누는 선택은 진입 시점에 결정.

## 본 세션 (2026-05-15) 도입 핵심 (재참조 가능)

### 1. 도메인 / 인프라 4 패턴
1. `frontend/lib/config/endpoints.ts` — 모든 서비스 URL default 단일 진실 소스 (native default + env override)
2. `app/layout.tsx` inline script — theme FOUC 방지
3. `next.config.ts output: standalone` — NEXT_OUTPUT env gate
4. ADR-0011 §4.2 `enforceRowOwnership` + audit `auth.row_denied` + pmo_manager seed migration 000021

### 2. DREQ 도메인 4 결정 (sprint f~k)
- **컨셉** (sprint f, PR #121) — `docs/planning/development_request_concept.md` + REQ-FR-DREQ-001..011 + UC-DREQ-01..10 + ARCH-DREQ-01..06 + API-59..65 spec
- **AuthADR** (sprint g, PR #122) — ADR-0012 옵션 A (API 토큰 + IP allowlist) + `dev_request_intake_tokens` 테이블 스펙
- **Backend** (sprint i, PR #124) — 7 endpoint activated + `requireIntakeToken` middleware + 19 신규 unit test + migration 000022/023/024
- **Frontend** (sprint j, PR #125) — `/admin/settings/dev-requests` (system_admin 전체관리) + `/dev-requests` (일반 사용자, codex #125 hotfix) + DevRequestTable / DevRequestDetailModal / MyPendingDevRequestsWidget + developer/manager dashboard 통합

### 3. codex review cycle 3회
- hotfix #1 (sprint d, PR #119) — PR #114/#118 의 P1×3 + P2×1
- hotfix #2 (sprint h, PR #123) — PR #119/#120/#121 의 P1×2 + P2×2 (migration 000021 down FK + API-65 close 권한 + REQ source_system + sprint memory finalize)
- hotfix #3 (sprint k, PR #126) — PR #122/#124/#125 의 P1×2 + P2×2 (assignee FK rejected row + limit/offset + 일반 사용자 페이지 + session_handoff header)

## 본 세션 (2026-05-15) 누적 머지 — 15 PR

| PR | sha | sprint | 작업 |
| --- | --- | --- | --- |
| #112 | 3f387cd | codex/frontend_color_review | (흡수) Admin UI + ActionMenu + iPad |
| #115 | b669bc7 | gemini/frontend_redesign_260514 | Light theme + dropdown + endpoints 통일 |
| #114 | 25f97ba | codex/260514-a | Application leader/dev_unit + search 확장 |
| #116 | cbc36b0 | claude/work_260515-a | sprint a housekeeping |
| #117 | 68f031e | claude/work_260515-b | 모달 token 정책 sweep |
| #118 | 519a508 | claude/work_260515-c | enforceRowOwnership helper + ADR-0011 §4.2 |
| #119 | bca612e | claude/work_260515-d | codex hotfix #1 |
| #120 | feac299 | claude/work_260515-e | sprint e housekeeping |
| #121 | 52f6ad8 | claude/work_260515-f | DREQ 도메인 컨셉~설계 staged |
| #122 | 4d0277f | claude/work_260515-g | ADR-0012 DREQ AuthADR |
| #123 | 1d24acf | claude/work_260515-h | codex hotfix #2 |
| #124 | 333edc9 | claude/work_260515-i | DREQ Backend 1차 |
| #125 | 58033d2 | claude/work_260515-j | DREQ Frontend 1차 |
| #126 | bb164c4 | claude/work_260515-k | codex hotfix #3 |

## 0. 2026-05-15 머지 흐름 (7 PR)

```
bca612e PR #119 — fix(frontend,backend,migrations): codex review hotfix — PR #114 + PR #118 (claude/work_260515-d)
519a508 PR #118 — feat(httpapi,adr): enforceRowOwnership helper — ADR-0011 §4.2 + REQ-FR-PROJ-009 활성화 (claude/work_260515-c, 본인 4단계 리뷰)
68f031e PR #117 — refactor(frontend): PR #114 신규 컴포넌트 token 정책 일관 sweep (claude/work_260515-b, 본인 4단계 리뷰)
cbc36b0 PR #116 — docs(memory): 2026-05-15 sprint claude/work_260515-a housekeeping (claude/work_260515-a)
25f97ba PR #114 — feat(application): leader/dev_unit 모델 + search 확장 + auth_login canonical (codex/260514-a, 본인 4단계 리뷰)
b669bc7 PR #115 — feat(frontend): Light theme + dropdown refactor + endpoints 통일 (gemini/frontend_redesign_260514, 본인 4단계 리뷰)
3f387cd PR #112 — feat(frontend/org): Admin UI 개선 + iPad 터치 + 백엔드 트랜잭션 (codex/frontend_color_review, 2026-05-14 머지 누락 흡수)
```

## 1. 본 세션 도입 핵심 4 패턴 (재사용 가능)

### 1.1 endpoints 통일 모듈 (`frontend/lib/config/endpoints.ts`)
모든 서비스 URL default 단일 진실 소스. 정책: 코드 default = native(localhost), docker 는 env override (CLAUDE.md "native default, docker optional" + 메모리 [docker env-specific]). 사용처 8개 service 일괄 갱신 (`next.config`, `infra`, `rbac`, `realtime`, `websocket`, `auth`, `kratos-logout`). `frontend/.env.example` + `.gitignore` 의 .env.example 예외 nested 까지 확장.

### 1.2 theme FOUC 방지 (`app/layout.tsx` inline script)
paint 전 `<script>` 가 localStorage 의 `devhub-theme` 을 읽어 `theme-dark` class 적용. `Header.tsx` 의 useState 는 `document.documentElement.classList.contains` 로 lazy initialize. dead code `ThemeToggle.tsx` 제거 (Header dropdown 단일 진입점).

### 1.3 standalone gate (`next.config.ts`)
`output: "standalone"` 이 `next start` 와 호환되지 않아 CI/native dev 의 e2e webServer 가 깨졌던 문제. `NEXT_OUTPUT === "standalone"` env 일 때만 활성화 — docker 빌드 단계에서만 켠다.

### 1.4 ADR-0011 row-level 위양 (`enforceRowOwnership` + audit `auth.row_denied`)
- 시그니처: `func (h Handler) enforceRowOwnership(c *gin.Context, ownerUserID string, allowedRoles ...string) bool`
- allow 규칙: (1) `actor.role == system_admin`, (2) `actor.role ∈ allowedRoles`, (3) `actor.login == ownerUserID` (owner-self; ownerUserID == "" 자동 비활성화)
- deny: audit `auth.row_denied` + payload `{actor_role, owner_user_id, resource, action, denied_reason="owner_mismatch"}` + 403 `code=auth_row_denied`
- 운영 wire-up: `updateApplication / archiveApplication / updateProject / archiveProject` 4개 handler (PR #119)
- pmo_manager RBAC seed migration `000021_rbac_pmo_manager` — REQ-FR-PROJ-010 정책 매핑 (applications view+edit only, projects 전체, application_repositories view only)
- `devFallbackEnabled` 환경(test) 에서는 bypass — `enforceRoutePermission` 과 일관

## 2. 본 세션 리뷰어 모드 / codex review cycle 학습

### 2.1 본인 PR 4단계 리뷰 4회 (#114/#115/#117/#118)
diff 재검토 → gh pr comment (P0/P1/P2 분류) → 보강 commit (필요 시) → squash merge. **PR #115 의 E2E 가 2회 실패** 한 사례에서 **서비스 부팅 워닝 (Layer 1) + 실패 테스트 이름 (Layer 2)** 두 layer 분리 분석 패턴 정착:
- artifact `frontend.log` 의 "next start does not work with output standalone" → standalone gate fix
- raw job log 의 `header-switch-view.spec.ts:8/22/46/55` → Switch View 의도적 제거에 따른 e2e 정리

### 2.2 codex 외부 리뷰 hotfix 1회 (#119)
머지 후 codex 가 4 PR 에 inline 리뷰 (P1×3 + P2×1). PR #119 에서 일괄 처리:
- PR #114 P1: edit payload key 제외 (immutable 검증 회귀)
- PR #114 P2: date-only `new Date()` → `parseISO` (timezone shift 회귀)
- PR #115 P1: 이미 해소 (보강 commit `f621189`) — no action
- PR #118 P1: helper dead code → pmo_manager seed + handler wiring

본 sprint 가 정착한 패턴: **scope 와 effective production 효과의 불일치는 정당한 codex 지적**. ADR 의 carve out 이라도 PR body 가 "활성화" 라고 표기하면 effective 효과까지 가야 일관.

## 3. 다음 세션 진입 후보 (우선순위 순)

1. **owner-self route gate 활성화** — 현재 route-level RBAC gate 가 `applications:edit / projects:edit` 을 `system_admin/pmo_manager` 만 통과시키므로 owner-self 가 route gate 에서 막힘. owner-self effective 활성화는 route gate 정책 변경 (예: bypass owner-self for own row) + ADR 갱신 동반. ADR-0011 §4.2 의 자연 후속.
2. **critical_warning_count 임계치 외부화** — concept §13.2.1 운영 정책 테이블 신설 + handler lookup 동적화.
3. **CountApplicationCriticalWarnings lightweight SQL** — 현재는 전체 rollup compute 재호출, 성능 분리.
4. **Repository commit activity ingest** — pr_activities 외 commit 단위 이벤트.
5. **Project active→closed 가드 정책** — Application 만 critical 가드, Project 는 단순 transition.
6. **pr_activities.payload sanitization** — system_admin 외 노출 방어.
7. **quality_snapshots idempotency** — UNIQUE 추가 vs history retention.
8. **M4 RM-M4-XX 본격 진입** — WebSocket 확장 / replay / System Admin 대시보드 등.
9. **traceability follow-up** — TC-NAV-01/02/SIM-01 row 정리 (PR #115), endpoints 모듈 도입 ARCH 정책 1줄 (PR #115), pmo_manager seed traceability row (PR #119).

## 0. 2026-05-14 머지 흐름 (PR #104~#110, 7건)

```
d29f2ac PR #110 — fix(test,ci,store): PR #109 codex review hotfix — B1 fixture + I1 setup + I3 backend-integration job + 2 SQL fix + GITHUB_STEP_SUMMARY (claude/work_260514-f)
1e38c4d PR #109 — test(store): postgres integration test 도입 + P1/P2 회귀 guard 23 test (claude/work_260514-e)
7822a91 PR #108 — fix(application,store,httpapi): PR #107 codex review hotfix — P1 custom weight normalize + P2 UpdateIntegration unique mapping (claude/work_260514-d)
f11bdbb PR #107 — feat(application,store,httpapi,docs): API-51..58 세트 활성화 + active→closed critical 가드 흡수 (claude/work_260514-c)
66ab5ff PR #106 — feat(application,store,httpapi,docs): Application Design 2차 — A1+A2+A3 (API-41~50 activated) (claude/work_260514-b)
642d976 PR #105 — feat(application,adr,docs): Application Design 1차 + ADR-0011 결정 (claude/work_260514-a)
63a7ea2 PR #104 — docs: Application/Repository/Project 설계 고도화 + SCM 어댑터 모델 + self-review 13건 보강 (codex/project_concept_design)
```

## 0.1 Application 도메인 1차 완성 인벤토리

| 항목 | 결과 |
| --- | --- |
| API | API-01~58 전체 activated. 본 세션이 API-41..58 (18 endpoint group) 모두 활성화. |
| 마이그레이션 | 000012~000018 (7건) — scm_providers + applications + application_repositories + projects + project_members + project_integrations + pr_activities + build_runs + quality_snapshots + RBAC seed |
| ADR | ADR-0011 RBAC row-scoping accepted (옵션 C 1차 + 옵션 B 단계적 확장 옵션) |
| RBAC | 4 신규 resource (`applications` / `application_repositories` / `projects` / `scm_providers`) — system_admin only |
| Frontend | `PermissionMatrix` 9 resource 확장 (PR #105 self-review B1) |
| Domain types | Application / ApplicationRepository / SCMProvider / Project / ProjectMember / ProjectIntegration / PRActivity / BuildRun / QualitySnapshot / RepositoryActivity + WeightPolicy / Rollup* |
| Store interface | 27 메서드 (ApplicationStore) |
| Handler | 14 endpoint with 상태 전이 머신 + 가드 + audit emit (과거형 시제) |
| Unit test | 43 application 관련 (handler 25 + project 8 + integration 6 + rollup 4) |
| Integration test | 25 (Applications 9 + Repository ops 8 + Projects+Integrations 8 — DEVHUB_TEST_DB_URL 환경) |
| CI job | 4 → 5 (Workflow Lint / Backend Unit / Frontend Unit / E2E / **Backend Integration 신설**) |
| 본인 리뷰 | 4단계 × 5회 (#104, #105, #106, #107, #109/#110) — 모두 diff 재검토 → 코멘트 → 보강 → 머지 |
| codex 외부 리뷰 흡수 | 2회 (#107 → #108 hotfix / #109 → #110 hotfix) |

## 0.2 codex review cycle 학습 (다음 세션 적용)

머지 후 codex 외부 리뷰가 inline P1/P2 로 도착하는 시나리오가 본 세션에 2회 발생:
- PR #107 → P1 custom weight normalize 미실행 + P2 UpdateIntegration unique 매핑 누락 → PR #108 hotfix
- PR #109 → P1 fixture cleanup SQL multi-statement + bind args → PR #110 hotfix

본 세션이 정착한 패턴:
1. 머지 후 codex 외부 리뷰 확인
2. inline P1/P2 발견 시 hotfix PR 진입 (별도 브랜치)
3. 정정 + 회귀 guard test 추가 (integration test 가 결정적)
4. self-review 4단계 일관 유지

## 1. 다음 세션 진입 후보 (우선순위 순)

1. **frontend `/admin/settings/applications` UI** — IA 설계 + 화면 흐름. PR #105 self-review B1 의 PermissionMatrix 확장만 했고 페이지 자체는 미생성.
2. **ADR-0011 §4.2 enforceRowOwnership helper** — Owner 위양 2차 단계 (pmo_manager 활성화 sprint). 시그니처는 ADR-0011 §6 에 명시.
3. **critical_warning_count 임계치 외부화** — concept §13.2.1 운영 정책 테이블 신설 + handler lookup.
4. **CountApplicationCriticalWarnings lightweight SQL** — 성능 (현재는 전체 rollup compute 재호출).
5. **Repository commit activity ingest** — pr_activities 외 commit 단위 이벤트.
6. **Project active→closed 가드 정책 결정** — Application 만 critical 가드, Project 는 단순 transition.
7. **pr_activities.payload sanitization** — system_admin 외 노출 방어.
8. **quality_snapshots idempotency 결정** — UNIQUE 추가 vs history retention.
9. **M4 RM-M4-XX 본격 진입** — WebSocket 확장 / replay / System Admin 대시보드 등.
10. **세션 진입 시점에 사용자 환경 ops verification** — `scripts/setup-test-db.sh` 1회 실행 권장 (integration test 가 CI 외에 로컬에서도 정합).

## 0. 2026-05-13 머지 흐름

```
244f6b1 PR #102 — docs(planning): Project 도메인 컨셉 1차 — CRUD + 등록 + 조회 MVP scope (claude/work_260513-p)
118899b PR #101 — docs(memory): 2026-05-13 세션 종료 housekeeping sync (claude/work_260513-o)
4134b37 PR #100 — feat(domain,migrations),docs(adr,traceability),scripts: M4 전 잔여 일괄 (claude/work_260513-n)
c7c2f35 PR #99 — feat(hrdb,org,signup),docs(adr,traceability),test: M3 후속 1-4 일괄 (claude/work_260513-m)
b1268ce PR #98 — feat(auth),docs(adr,api,traceability),test: M3 진입 1차 — RM-M3-01..03 (claude/work_260513-l)
3d7d5a2 PR #97 — docs(roadmap,traceability): M3/M4 drift 정합 (claude/work_260513-k)
f551e6a PR #96 — feat(auth),docs(adr,api,traceability),test: M3 진입 전 잔여 후속 일괄 (claude/work_260513-j)
cb9e6d5 PR #95 — docs(traceability,adr,ci),test(frontend): 대형 묶음 B1~D5 (claude/work_260513-i)
ceb0f6f PR #94 — docs(adr): ADR-0004 X-Devhub-Actor 폐기 완료 선언 (B4) (claude/work_260513-h)
594be74 PR #93 — docs(traceability,api): B1 auth 도메인 2차 — §11 본문 ID 노출 + IMPL-auth-01..07 책임 정의 (claude/work_260513-g)
a73dba1 PR #92 — docs(traceability,api): B 묶음 — RBAC API §12 IMPL 정밀 매핑 + 본문 ID 노출 (claude/work_260513-f)
ae8b459 PR #91 — feat(backend): A 묶음 — M1 PR-D 정합 마무리 (writeRBACServerError + writeAuthLoginServerError 통합 + X-Request-ID validation + ctx 표준 request_id 전파) (claude/work_260513-e)
ea8df91 PR #90 — docs(governance,traceability): 갭 분석 정리 + 메타 헤더 표준화 + main flat sync (claude/work_260513-d)
7fac5bf PR #89 — docs: governance + traceability 체계 도입 + 1차 종합 매트릭스 (claude/work_260513-c)
9268227 PR #88 — docs(adr): ADR-0003 no-docker policy CI scope — drop services: postgres + native PG 15 (claude/work_260513-b)
e86f38f PR #87 — ci: FU-CI-2/3/4 (playwright scope, GHA cache, frontend readiness) + main flat memory sync (claude/work_260513-a)
450cc24 PR #86 — ci: E2E 테스트 안정화 및 GitHub Actions CI 파이프라인 구축 (gemini/prepare-github-action)
```

세부:
- **PR #86** — GitHub Actions CI 1차 도입. backend-unit + frontend-unit + e2e 3잡. 리뷰어 모드 2-pass 로 5 blocker + 5 follow-on fix.
- **PR #87** — FU-CI-2/3/4 처리. playwright install scope (chromium 단일), `actions/cache@v4`, frontend readiness 120s, install split (cache-hit 시 browser install skip).
- **PR #88** — ADR-0003 (no-docker 정책 CI 범위 명문화). `services: postgres:15` sidecar 제거 + pgdg native PG 15 native 설치 step. PG-14 dropcluster + `--port=5432` 강제.
- **PR #89** — `docs/governance/` (README + document-standards.md) + `docs/traceability/` (README + conventions + sync-checklist + report) + PR template + AGENTS/GEMINI 진입점. 1차 종합 매트릭스: REQ-FR 105 + REQ-NFR 26 + ARCH 17 + API 40 + RM 28 + IMPL 79 + UT 47 + TC 37 = 412 항목.
- **PR #90** — 매트릭스 §5 갭 표 형식 통일 + §5.2 auth.spec.ts TC 미흡수 + §5.3 ADR-0001 vs §3.8 모두 closed. document-standards §2 메타 헤더 9 문서 (누락 4 + 변형 4 + 부분 1) 일괄 적용. PR #87/#88/#89 누적 main flat memory sync.
- **PR #91** — A 묶음 (M1 PR-D 정합 마무리). `writeRBACServerError` (11 호출) + `writeAuthLoginServerError` (5 호출, Pass 1 review 발견) → `writeServerError` 일괄 통합. `requireRequestID` 미들웨어에 `validateCallerRequestID` (정규식 + 길이 + control char) 추가. request_id 를 `requestIDCtxKey{}` typed ctx key 에도 stash + `requestIDFromContext`/`logRequestCtx` ctx-aware helper. kratos_login_client.go 2건 + kratos_identity_resolver.go 1건 ctx-aware 치환. 단위테스트 11건 추가.
- **PR #92** — B 묶음 RBAC 1차. `backend_api_contract.md` §12.2~§12.10 의 9 헤더에 `(API-26..31, 38..40)` 본문 ID 노출. 매트릭스 §2.2 RBAC API + §2.4 IMPL-rbac-01..04 책임 정의 (handler / store / enforcement / cache) 서브 표 도입. §5.2 RBAC IMPL 정밀 매핑 항목 closed. Pass 1 review 보강으로 §3 RBAC 행을 ID 범위 + §2 서브 표 참조 패턴으로 정리 ("표 가독성 정책" 명문화).
- **PR #93** — B1 auth 도메인 2차. `backend_api_contract.md` §11.3 `(API-19)` + §11.5 표에 API ID 컬럼 (`API-20..24, 35`) + §11.5.1 `(API-35)` 본문 ID 노출. 매트릭스 §2.2 Auth API + §2.4 IMPL-auth-01..07 책임 정의 (verifier / actor / 5 endpoint handler) 서브 표 도입. §3 인증/회원가입/계정 관리 행 정리 (cross-cut API-23 / API-35 명시).
- **PR #101** — 2026-05-13 세션 종료 housekeeping sync 1차. main flat memory 의 PR #100 흡수 + 다음 세션 진입 후보 명문화 (RM-M4-01..09 + M3 carve out 6 항목).
- **PR #102** — Project 도메인 컨셉 1차. `docs/planning/project_management_concept.md` 신규 (10 절: 도메인 정의 / 행위자 × usecase / MVP scope / 데이터 모델 초안 / 다른 도메인 연계 / UI 컨셉 / 미해결 항목 / 후속 4 sprint hook). 일반 사용자 = 조회 중심, 시스템 관리자 = 등록·관리 전용 분리. RBAC row-scoping (ADR-0011 후보) 는 Design sprint 보류. `docs/planning/README.md §5.1` 도메인 컨셉 인덱스 신설 + `docs/development_roadmap.md §5` 백로그 1행 추가. 추적성 ID 미발급 (컨셉 단계).

## 1. 세션 종료 — 다음 세션 진입 후보

본 세션 (2026-05-13) 종료. 다음 세션 진입자는 본 §1 + §2 를 기준으로 작업 선택.

### 1.1 RM-M4 진입 후보 (매트릭스 §2.3.2 참조)

| RM ID | 항목 | 후보 sprint plan 단위 |
| --- | --- | --- |
| `RM-M4-01` | WebSocket 확장 — infra/ci/risk event publish | M4-WS sprint |
| `RM-M4-02` | WebSocket replay + 리소스 필터 | M4-WS (RM-M4-01 과 묶음 가능) |
| `RM-M4-03` | frontend command status WebSocket UI | M4-WS-UI |
| `RM-M4-04` | AI Gardener gRPC (Python AnalysisService + Go Core client) | M4-AI |
| `RM-M4-05` | AI Suggestion Feed 실데이터 바인딩 | M4-AI (RM-M4-04 직속 후속) |
| `RM-M4-06` | Gitea Hourly Pull worker (Phase 10) | M4-task |
| `RM-M4-07` | System Admin 대시보드 | M4-admin |
| `RM-M4-08` | RBAC PermissionCache LISTEN/NOTIFY ([ADR-0007](../../docs/adr/0007-rbac-cache-multi-instance.md)) | M4-infra (격리된 backend 작업) |
| `RM-M4-09` | 외부 SSO 통합 (Gitea 등) | M4-SSO |

### 1.1.b Project 도메인 후속 진입 후보 (PR #102 컨셉 1차 머지 후, 본 세션 신규)

본 세션에서 [`docs/planning/project_management_concept.md`](../../docs/planning/project_management_concept.md) 컨셉 1차 머지 완료. 후속은 컨셉 §9 의 4 sprint hook 을 따른다.

| 후속 sprint | 산출물 | 진입 조건 |
| --- | --- | --- |
| **Project-Req** | `docs/requirements.md §5.7` 확장 또는 `§5.8 Project 도메인` 신설, REQ-FR-* 일괄 발급 (개별 usecase 단위), NFR (응답시간/페이지네이션) 정의, 매트릭스 row 추가 | 컨셉 머지 직후 (즉시 가능) |
| **Project-Usecase** | 행위자 × usecase 매트릭스 + 핵심 시퀀스 (등록·조회·멤버 변경 3종) + RBAC 매트릭스 확장 후보 → ADR-0011 초안 | Project-Req 머지 직후 |
| **Project-Design (backend)** | `architecture.md` Project 컴포넌트 추가, `backend_api_contract.md` 신규 § `/api/v1/projects/*`, 마이그레이션 (`000012_projects.sql`) | Project-Usecase 머지 직후 |
| **Project-Design (frontend)** | `frontend_development_roadmap.md` 새 phase, 진입 경로 / 컴포넌트 / store 모델 초안 | Backend design 와 병행 |
| **Project-Impl** | IMPL-project-* / UT-project-* / TC-PROJ-* 발급 + 실 구현 | 모든 design 머지 후 |

### 1.2 M3 carve out (M4 와 병행 가능)

- `getHierarchy` MV join 코드 변경 ([ADR-0009](../../docs/adr/0009-org-secondary-memberships-and-total-count-mv.md) §4.2)
- ETL daily cron 운영 entry ([ADR-0008](../../docs/adr/0008-hrdb-production-adapter.md) §6)
- `primary_dept` backfill worker (signup 직후 + admin trigger, [ADR-0010](../../docs/adr/0010-primary-dept-resolution.md) §4.3)
- 파견 종료 (`is_seconded` 자동 갱신) trigger
- 본격 Vitest (AuthGuard / Header / Sidebar mock-heavy)
- TC-CMD/INFRA 인터랙션 spec ts (`test_cases_m3_command_infra.md` §4)

### 1.3 본 세션 머지 영역 요약 (PR #86~#102, 17건)

- **거버넌스/추적성** 체계 도입 + 1차 종합 매트릭스 + 갭 정리 + 메타 헤더 표준화 + M3/M4 drift 정합.
- **Project 도메인 컨셉 staged** (PR #102): 신규 1차 도메인의 컨셉 단계 문서 신규 작성, 후속 4 sprint (Req → Usecase → Design backend/frontend → Impl) 진입 hook 명문화. 추적성 ID 미발급 (컨셉 단계).
- **M1 PR-D 정합** (A 묶음): writeRBACServerError + writeAuthLoginServerError → writeServerError 통합, X-Request-ID validation, ctx 표준 request_id 전파.
- **X-Devhub-Actor 폐기**: ADR-0004 + ADR-0006 (inbound 400 거부) + 회귀 4 테스트 의도 갱신.
- **5 도메인 본문 ID 노출**: RBAC §12 / auth §11 / account §10.2 / users CRUD §10.3 / organization §10.4 / accounts admin / signup §11.5.2.
- **CI / lint**: GitHub Actions 4 잡 (workflow-lint + backend + frontend + e2e) + ADR-0003 no-docker + ADR-0005 actionlint.
- **M3 1차 closing**: Sign Up audit emit + 단위테스트 + §11.5.2 본격 spec / hrdb.MockClient → PostgresClient (ADR-0008) + migration 000010 / parent_id cycle 검증 + ResolvePrimaryUnit (ADR-0010) + 5 단위테스트 / total_count MV migration 000011 (ADR-0009) / scripts/hrdb_etl_seed.sql / frontend signup form alias.

### 1.4 ADR 인덱스 (본 세션 신규 5건)

| ADR | 제목 |
| --- | --- |
| ADR-0006 | inbound `X-Devhub-Actor` 헤더 명시 거부 (400) |
| ADR-0007 | RBAC PermissionCache 다중 인스턴스 일관성 (PG LISTEN/NOTIFY, 구현 carve) |
| ADR-0008 | HR DB production 어댑터 (PostgreSQL `hrdb` schema) |
| ADR-0009 | 파견/겸임 모델 + `total_count` Materialized View |
| ADR-0010 | `users.primary_unit_id` 자동 판정 알고리즘 |

### 1.5 다음 세션 진입 안내

다음 세션 첫 액션 권장:

1. 본 `session_handoff.md` + `state.json` + `work_backlog.md` 읽고 main HEAD 가 본 housekeeping 직후 commit 인지 확인.
2. 본 §1.1 RM-M4 / §1.1.b Project 도메인 후속 / §1.2 M3 carve out 중 우선순위 결정.
3. 새 sprint branch (`claude/work_260514-a` 또는 `claude/project-req-1` 등) + sprint memory 초기화 → 진행.

## 2. 다음 진입점 — 우선순위 후보

`state.json#next_actions` 참조. 본 sprint 머지 후 후속:

| 후보 | 주요 작업 | 규모 |
| --- | --- | --- |
| document-standards §8 우선순위 3 | 본문 ID 명기 (요구사항/설계 문서에 REQ-FR/ARCH/API 옆 backtick ID) | M |
| document-standards §8 우선순위 4 | deprecated 문서 식별 + 마킹 | S |
| E2E 신규 TC 작성 | TC-CMD-*, TC-INFRA-* — 매트릭스 §5.1 의 후보 구현 | M |
| frontend 컴포넌트 Vitest | Header, Sidebar, AuthGuard 등 | S |
| RBAC API §12 IMPL 정밀 매핑 | endpoint 별 IMPL ID 명시 | S |
| X-Devhub-Actor 폐기 ADR | architecture.md §6.2.3 의 완전 제거 trigger | S |
| RBAC cache 다중 인스턴스 일관성 | M1-DEFER-E | M-L |
| actionlint / workflow lint | ADR-0003 §6 의 후속 ADR | S |
| M3/M4 진입 | command status WebSocket UI, WebSocket 확장, AI Gardener gRPC, Gitea Hourly Pull worker | L |

## 3. 환경 / 운영 메모

- **CI 환경**: GitHub Actions ubuntu-24.04. native PostgreSQL 15 (pgdg) + native Ory Hydra/Kratos v26.2.0. `services:` 컨테이너 사용 안 함 (ADR-0003).
- **5 프로세스 native 기동** (prod / dev-server): PostgreSQL(시스템 서비스) + Hydra + Kratos + backend-core + frontend.
- **e2e 자동 시드**: `cd frontend && npm run e2e` 한 줄.

## 4. 잔여 결정 대기

- 본 sprint 후속 항목 (위 §2 표) 의 우선순위 결정.
- 운영 진입 전 hygiene: PoC `test/test` 시드 제거 (test-server-deployment.md §10).

## 5. 거버넌스 / 추적성 진입점

- `docs/governance/README.md` — 두 축 (문서 관리 + 추적성) 인덱스.
- `docs/governance/document-standards.md` — 문서 표준 (메타 헤더, lifecycle, 단계별 유형).
- `docs/traceability/conventions.md` — ID 컨벤션.
- `docs/traceability/sync-checklist.md` — 매 PR 동기화 절차.
- `docs/traceability/report.md` — 1차 종합 매트릭스.
- `.github/pull_request_template.md` — PR body 의 "추적성 영향" 섹션 자동 안내.
