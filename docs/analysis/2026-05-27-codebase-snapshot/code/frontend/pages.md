# Frontend 분석 — Pages (App Router)

- 문서 목적: `frontend/app/**` 의 33 `page.tsx` + 4 `layout.tsx` 를 라우트별로 역할·데이터 소스·권한 가드·운영 UI 전환 여부 기준으로 인벤토리화한다.
- 범위: `frontend/app/` (node_modules / .next / coverage 제외). 스택 = Next.js 16.2.6 (App Router) + React 19.2.4 + Tailwind 4 + framer-motion + Zustand.
- 최종 작성일: 2026-05-27 (main `cf19c94`)
- 관련 문서: `components.md`, `services.md`, `frontend_platform.md`

## 1. 라우트 그룹 구조

- `app/layout.tsx` — Root layout. `<html>`/`<body>`, `ToastContainer`, `LogoutOverlay`, 그리고 paint 전 theme bootstrap inline `<script>` (FOUC 회피, `localStorage["devhub-theme"]` 공유).
- `app/(dashboard)/layout.tsx` — 인증 영역 shell. `AuthGuard` 로 감싼 뒤 `Sidebar` + `Header` + `OnboardingBanner` + `<main>`. 모든 dashboard 라우트가 여기를 통과한다.
- `app/(dashboard)/admin/settings/layout.tsx` — system_admin 전용 settings nested layout. 9개 sub-tab 을 3 카테고리(Access Control / App & Requests / Integrations & Audit)로 분류한 좌측 nav + 모바일 glass dropdown + collapse 토글(`localStorage["admin-settings-sidebar-collapsed"]`). `isSystemAdmin(actor?.role)` defense-in-depth 가드 → 비-admin 은 `defaultLandingFor` 로 replace, `if (!allowed) return null`.
- 인증 외 라우트(`/login`, `/auth/*`, `/onboarding`, `/signup`)는 그룹 밖 → AuthGuard 미적용.

## 2. 인증/공개 라우트

| 라우트 | 역할 | 데이터 소스 | 가드 | 운영/비고 |
|---|---|---|---|---|
| `app/page.tsx` | 루트. `redirect("/developer")` (server) | — | 없음 | 실질 진입은 role별 landing 으로 재라우팅 |
| `login/page.tsx` | OIDC 진입점(단일 canonical). 무에러 진입 시 자동 `authService.getAuthorizeURL()` → `window.location.assign`. `?error=` 있으면 머무르며 메시지 노출 | `authService` (PKCE) | 없음 | `resolveErrorMessage` export (vitest 대상). `redirectStartedRef` 로 이중 redirect 방지 |
| `auth/callback/page.tsx` | OIDC code→token 교환 후 identity 해석 → `defaultLandingFor(role)` replace. 실패 시 에러 카드 + `/login?error=login_failed` | `authService.exchangeCode` + `resolveIdentity` | 없음 | `processedRef` 로 1회 실행. `Suspense` (useSearchParams) |
| `auth/logout/page.tsx` | mount 시 `authService.logout()` (RP-initiated logout) | `authService` | 없음 | id_token_hint + post_logout_redirect_uri (basePath 포함) |
| `auth/error/page.tsx` | IdP 에러 redirect landing. provider별 query key 정규화 후 "Restart sign-in" | searchParams only | 없음 | static 표시 |
| `auth/signup/page.tsx` | "Sign Up Unavailable" 안내(Keycloak 마이그레이션 중 self-signup 비활성) | 없음 (static) | 없음 | 운영 정책 명문화 |
| `signup/page.tsx` | legacy alias. `redirect("/auth/signup")` | — | 없음 | deprecation redirect |
| `onboarding/page.tsx` | 첫 로그인 프로필 등록. `whoAmI` → `onboarding_required` false 면 landing replace, true 면 `OnboardingForm` 노출 | `identityService.whoAmI` | 401 → `/login?error=session_expired` | `fromAdmin` = admin pre-seed 여부(display_name+unit 둘 다 존재) |

## 3. 역할별 대시보드 (운영 UI 전환 = 대부분 아카이브)

| 라우트 | 역할 | 데이터 소스 | 가드 | 운영 전환 상태 |
|---|---|---|---|---|
| `(dashboard)/developer/page.tsx` | "Work Status Archived" 안내 + `MyPendingDevRequestsWidget` | dev_request.service (위젯) | AuthGuard | **아카이브** — 본문은 Projects 로 유도. 위젯만 실데이터 |
| `(dashboard)/manager/page.tsx` | "Quality Status Archived" 안내 (Projects 유도) | 없음 | AuthGuard | **아카이브** — 순수 static. 구 manager velocity/team-load UI 폐기 |
| `(dashboard)/admin/page.tsx` | "Sys Admin Dashboard Archived" 안내 (Topology v2 / Settings 유도) | 없음 | AuthGuard + (admin landing) | **아카이브** — static |
| `(dashboard)/gardener/page.tsx` | AI Gardener 피드 + Autonomous Stats | `GardenerFeed`(gardener.service) | client `userRole !== "System Admin"` → `/developer` push | 부분 운영. 우측 "Autonomous Stats"(94/100, Locked) **하드코딩 mock 잔재** |

> dashboard.service.ts(developer stream/builds, manager velocity/team-load/decisions), risk.service.ts 는 백엔드 호출 코드가 남아있으나 위 아카이브 페이지들이 더 이상 호출하지 않는다 → service dead-path (services.md 참조).

## 4. 운영 도메인 페이지 (PageState 공통화 적용)

공통 패턴: `useState(loading/error)` + `loadData()` async + `PageLoading`/`PageError(onRetry)`/`PageEmpty` (`components/ui/PageState.tsx`) + `DashboardHeader` + `FilterBar`. 데이터는 service 직접 호출, 클라이언트 측 필터링.

| 라우트 | 역할 | 데이터 소스 | 가드 | 비고 |
|---|---|---|---|---|
| `(dashboard)/applications/page.tsx` | Application 현황 목록 + rollup stat 카드 4종 | `applicationService.listApplications` + `getApplicationRollup` (N병렬) | AuthGuard | stat 카드 "Active Applications" = `status==="active"` 실집계 (#369 mock 교체 후) |
| `(dashboard)/applications/[id]/page.tsx` | Application 상세 + rollup BarChart + repositories/projects | application.service + project.service | AuthGuard | rollup 실패 시 PageError+retry (TC-APP-DETAIL-ERR-01) |
| `(dashboard)/projects/page.tsx` | Project 현황. 전 repo 의 project 집계 | `repositoryService.listRepositories` → `projectService.listAllProjects` | AuthGuard | progress bar 는 status proxy (실 진척도 아님 — stale UI) |
| `(dashboard)/projects/[id]/page.tsx` | Project 상세 + activity + tasks + repo link | project.service + identity.service + repository.service | AuthGuard | `ENABLE_LEGACY_MOCK_UI`(false) + legacyMock* import — **mock 잔재 분기** |
| `(dashboard)/repositories/page.tsx` | Repository 활동성. activity + build-runs(50)로 buildHealth 판정 | `repositoryService` (activity + build-runs N병렬) | AuthGuard | `evaluateBuildHealth` branch별 terminal status 집계 |
| `(dashboard)/repositories/[id]/page.tsx` | Repository 상세 + activity 카드 | repository.service | AuthGuard | activity 실패 시 PageError+retry (TC-REPO-DETAIL-ERR-01) |
| `(dashboard)/dev-requests/page.tsx` | 내 개발 의뢰 목록 + 상세 모달 | `devRequestService.list` | AuthGuard | `isSystemAdmin(actor.role)` 로 모달 액션 권한 전달 |
| `(dashboard)/account/page.tsx` | Account Settings + `ProfileSelfEdit` + Keycloak Console 링크 | `useStore.actor` + `authService.getAccountConsoleURL` | AuthGuard + AuthGuard limited-mode blocklist(`/account`→`/onboarding`) | email `{login}@example.com` + "MFA Status: Disabled" **하드코딩 잔재** |
| `(dashboard)/organization/page.tsx` | legacy. `replace("/admin/settings/users")` | — | AuthGuard(+ pathRequiresSystemAdmin) | deprecation redirect (PR-S2) |

## 5. Admin (system_admin 전용)

가드 2중: AuthGuard 의 `pathRequiresSystemAdmin` + admin/settings layout 의 `isSystemAdmin` 재검.

| 라우트 | 역할 | 데이터 소스 | 비고 |
|---|---|---|---|
| `admin/catalog/page.tsx` | Applications/Repositories/Projects 통합 카탈로그 (탭+검색, URL query 동기화) | application/repository/project service | repository **draft 생성/publish** UI 가 `prompt()` 기반 (handleCreateRepositoryDraft / requestRepositoryPublish) — primitive UI |
| `admin/topology-v2/page.tsx` | React Flow 기반 infra topology v2 (Grouped/Flat, node detail modal) | `infraService.getTopologyV2` + `realtimeService` 구독 | **유일한 ticket WS(realtime.service) 실시간 소비처** (`infra.node.updated`/`infra.service.updated`) |
| `admin/settings/page.tsx` | `replace("/admin/settings/users")` | — | index redirect |
| `admin/settings/users/page.tsx` | 사용자 관리 + 검색 + `PendingReviewPanel` | identity.service + rbac.service | onboarding pending_review 검토 패널 포함 |
| `admin/settings/organization/page.tsx` | 조직 단위 CRUD (list/grid/chart 뷰) | identity.service (units/members/hierarchy) | OrgTree(React Flow) chart 뷰 |
| `admin/settings/permissions/page.tsx` | RBAC policy 매트릭스 편집 | rbac.service | lazy clone 으로 baseline 보존 |
| `admin/settings/applications/page.tsx` | Application + Repository + Project CRUD | project.service | 다중 모달 조합 |
| `admin/settings/dev-requests/page.tsx` | 전체 의뢰 관리 (source_system 필터) | dev_request.service | system_admin 액션 활성 |
| `admin/settings/dev-request-tokens/page.tsx` | intake token 발급/편집/revoke | dev_request_token.service | IssueModal(plain 1회 reveal) + DestructiveConfirmModal |
| `admin/settings/integrations/page.tsx` | 외부 연동 provider 관리 + SCM import/create-repo | integration.service | ProviderModal/Import/CreateScmRepo 모달. mount effect dep 의도적 비움(codex #6 P1) |
| `admin/settings/integration-bindings/page.tsx` | provider↔scope binding 관리 (페이지네이션) | integration.service | Create/Edit/Destructive 모달 |
| `admin/settings/audit/page.tsx` | 감사 로그 (필터 + 50건 페이지네이션) | audit.service | draft/applied filter 분리 |

## 6. 핵심 가드 로직 — AuthGuard 3-branch onboarding gating

`components/layout/AuthGuard.tsx` 가 dashboard 진입 전 `whoAmI()` → store 갱신 후 다음 분기 (REQ-FR-ONBOARD-010):

1. `onboarding_required && !skipped` → `/onboarding` 강제 redirect (첫 진입).
2. `onboarding_required && skipped && 일반 페이지` → 통과 + `OnboardingBanner` 노출.
3. `onboarding_required && skipped && /account` → `/onboarding` hard redirect (limited-mode blocklist).

과거 whitelist 방식(`["/onboarding","/auth/"]`)이 default landing 까지 막아 무한 redirect 를 유발 → **blocklist 전환**(`LIMITED_MODE_BLOCKED_PREFIXES = ["/account"]`, `AuthGuard.tsx:17`). 추가로 `pathRequiresSystemAdmin && !isSystemAdmin(resolved.role)` → default landing replace. 실패 시 401 → `/login?error=session_expired`, 그 외 → `?error=login_failed`. `actor.role`(서버 진실)로 가드하고 zustand `role`(Header Switch View 시뮬레이션 가능)은 쓰지 않는다.

## 발견 사항 (불일치/stale/부채)

- **역할 대시보드 전면 아카이브 vs 잔존 service 호출 코드**: `developer`/`manager`/`admin` page.tsx 가 모두 static 안내 화면(`developer/page.tsx:15`, `manager/page.tsx:13`, `admin/page.tsx:13`)인데, 이를 채우던 `dashboard.service.ts`(developer stream/builds, manager velocity/team-load/decisions) + `risk.service.ts` 의 백엔드 호출 코드는 그대로 남아 dead-path 가 됐다. services.md 참조.
- **account 페이지 하드코딩 mock 잔재**: email 을 `{actor?.login}@example.com` 으로 합성(`account/page.tsx:66`) — actor.email 이 store 에 있는데도 무시. "MFA Status: Disabled" 도 하드코딩(`account/page.tsx:84`). 운영 UI 전환이 account 페이지에는 미적용.
- **gardener 우측 stat 하드코딩**: "Optimization Score 94/100", "Security Posture Locked"(`gardener/page.tsx:61`) 는 실데이터 미연동 mock. 좌측 `GardenerFeed` 는 실 service.
- **projects/[id] mock 분기 잔재**: `ENABLE_LEGACY_MOCK_UI`(상수 false) + `legacyMockProjectActivity/Tasks` import(`projects/[id]/page.tsx:25-26`). 죽은 분기지만 import 가 남아 lint/번들 노이즈.
- **repository draft/publish UI 가 `prompt()` 기반**: 최신 backend draft→publish lifecycle(#368/#373)의 프론트 진입점이 `admin/catalog/page.tsx:185-189` 의 `prompt()` 3연속 — 정식 모달 미구현. happy-path E2E 도 후행(repositories-ui.spec.ts 는 목록/검색/상세만 cover, draft/publish 시나리오 없음).
- **command status 실시간 UI 부재 (Phase 4 잔여)**: `realtime.service.ts:11` 의 `DEFAULT_EVENT_TYPES` 에 `command.status.updated` 가 구독 대상으로 들어있으나 이를 렌더하는 페이지/컴포넌트가 없다(grep 결과 service 정의 외 소비처 0). service-action(infra/gardener/risk)의 비동기 command 진행 UI 가 미완.
- **projects 진척도 가짜 표시**: progress bar 가 status 기반 고정 비율(`active`→2/3, `closed`→full, else 1/4)로 `projects/page.tsx:155-160`. 주석에도 "using status as a proxy" 명시 — 실 진척 데이터 미연동 stale UI.
- **admin 라우트 3중 가드 중복**: pathRequiresSystemAdmin(AuthGuard) + admin/settings/layout 의 isSystemAdmin + 일부 페이지(gardener) 자체 client redirect 가 산재. 의도적 defense-in-depth 지만 가드 진실원이 분산.
