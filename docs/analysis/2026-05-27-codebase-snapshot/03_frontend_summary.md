# 03. 프론트엔드 개발 사항 정리

- 문서 목적: 프론트엔드(Next.js) 개발 현황을 도메인/영역별로 정리한다.
- 기준: main `cf19c94` / `frontend/` (node_modules·.next 제외).
- 스택: Next.js 16.2.6 · React 19.2.4 · Zustand 5 · @tanstack/react-query 5 · TailwindCSS 4 · @xyflow/react 12 · framer-motion 12 · Vitest 4 · Playwright 1.59.

---

## 1. 영역별 완성 현황

| 영역 | 상태 | 비고 |
| --- | :-: | --- |
| 인증/세션 (OIDC PKCE + token refresh + AuthGuard) | ✅ done | `auth.service` + `token-store` + `pkce` + `api-client` 401 refresh. ticket WS. |
| 역할 라우팅 (defaultLandingFor + system route gate) | ✅ done | `lib/auth/role-routing.ts`. |
| 온보딩 (page + OrganizationPicker + Banner + gating + /account edit) | ✅ done | RM-ONBOARD-02/03 완료(#288). 3-branch gating(layout). |
| 조직/사용자 관리 (Org tree/units/members + Users CRUD + Permission matrix) | ✅ done | `/admin/settings/{organization,users,permissions}`. |
| 감사 로그 뷰 | ✅ done | `/admin/settings/audit`. |
| Application/Repository/Project 현황 + admin CRUD | ✅ done | 목록/상세 + admin/settings/applications + admin/catalog. |
| Repository draft/publish + SCM import/create UI | ✅ done | ImportRepositoriesModal + CreateScmRepositoryModal + ProviderTable 액션. |
| DREQ (목록/상세/위젯 + intake token admin) | ✅ done | `/dev-requests` + `/admin/settings/{dev-requests,dev-request-tokens}`. |
| External Integration (provider/binding + auth_mode 동적 입력 + vendor 템플릿) | ✅ done | `/admin/settings/{integrations,integration-bindings}` + ProviderModal. |
| Infra Topology v2 (React Flow + WebSocket 실시간 + 그룹화) | ✅ done | `/admin/topology-v2`. |
| 운영 UI 전환 (mock 제거 + PageState + 에러 표준화) | ✅ done(1차) | #334/#340/#342/#369. |
| Command status WebSocket UI (toast/status) | 🟡 부분 | RealtimeService 구독은 있으나 command toast 연결 미완(Phase 4 잔여). |
| 대시보드 실시간 로그 스트리밍 | 🟡 부분 | gardener feed 위주, 로그 스트리밍 UI 미완. |
| AI Gardener suggestion UI | 🔴 v2 | backend 미구현 의존. |

## 2. 페이지 인벤토리 (33)

- **진입/인증**: `/`(developer redirect), `/login`, `/signup`(redirect), `/auth/{callback,logout,signup,error}`, `/onboarding`.
- **대시보드 그룹** `(dashboard)`: layout(AuthGuard+Header+Sidebar+OnboardingBanner) / `developer` / `manager` / `gardener` / `account`.
- **도메인**: `applications`(+`[id]`) / `projects`(+`[id]`) / `repositories`(+`[id]`) / `dev-requests` / `organization`.
- **admin**: `admin` / `admin/topology-v2` / `admin/catalog` / `admin/settings`(+ users/permissions/audit/organization/applications/dev-requests/dev-request-tokens/integrations/integration-bindings).

## 3. 컴포넌트 (50, 폴더별 요약)

| 폴더 | 수 | 핵심 |
| --- | --: | --- |
| `ui` | 10 | Modal·Badge·Toast·FilterBar·ComboBox·PageState·ActionMenu·DestructiveConfirmModal·DashboardHeader·LogoutOverlay (공통 디자인 시스템) |
| `organization` | 11 | OrgTree·OrgNode·OrgUnitGrid/Table·MemberTable·User/Member/Unit 모달·PermissionEditor·PermissionMatrix |
| `integration` | 7 | ProviderTable/Modal·BindingsTable·Create/EditBindingModal·ImportRepositoriesModal·CreateScmRepositoryModal |
| `dev-request` | 6+ | DevRequestTable·DetailModal·IntakeToken(Table/Issue/Edit Modal)·MyPendingDevRequestsWidget (+__tests__) |
| `project` | 6 | Project/Repository/Application Table + 각 CreationModal + RepositoryLinkModal |
| `onboarding` | 3 | OnboardingForm·OnboardingBanner·OrganizationPicker |
| `layout` | 3 | Header·Sidebar·AuthGuard |
| `account`/`admin/users`/`dashboard` | 4 | ProfileSelfEdit·ConfirmReviewModal·PendingReviewPanel·GardenerFeed |

## 4. 서비스 레이어 (18)

`api-client`(401 자동 refresh) · `base` · `auth`(OIDC PKCE/discovery/Keycloak Account Console URL) · `identity`(whoAmI) · `realtime`(ticket WS pub/sub) · `websocket`(legacy) · `project` · `application` · `repository` · `dev_request` · `dev_request_token` · `integration`(provider/binding/test-connection/scm-repo import·create) · `audit` · `dashboard` · `gardener` · `infra`(topology v2) · `rbac` · `onboarding` · `risk`.

보조: `lib/auth/{token-store,pkce,role-routing}`, `lib/config/endpoints`, `lib/storage/onboardingSkip`, `lib/store`(Zustand: actor/role/toasts/notifications/sidebar), `lib/services/{error-message,integration-provider-presets,wire,types}`.

## 5. 테스트

- **Vitest 10**: `pkce` / `token-store` / `role-routing` / `utils` / `login/page` / `AuthGuard` / `project.service` / `integration-provider-presets` / `IntakeTokenTable` / `IssueIntakeTokenModal`.
- **Playwright 28**: auth/signup/signout/account/onboarding-first-login/rbac-routes/header-switch-view + admin-(topology-v2/permissions/org-crud/users-crud/users-search/projects/applications/integrations/integration-bindings/catalog/project-model-v2) + dev-requests/audit + repositories-(ui/ui-negative/detail-negative) + applications-projects-detail-negative + project-model-modes + infra-topology + dashboard-retry-empty-state + screenshots.

## 6. 최근 주요 변화 (2026-05-26~27)

1. **운영 UI 전환** — mock 대시보드 archive 분리 + 실데이터 렌더링 + `PageState` 공통화 + 에러 메시지 표준화.
2. **admin catalog** (#357/#361) — `/admin/catalog` 중심 project/application/repository 운영 동선 + archived 포함 listing.
3. **외부 연동 등록 UX** (#352) — vendor 템플릿 7종 + 가이드 자격증명(strategy+secret 분리) + capability 체크박스 + base_url + 연결 테스트.
4. **auth_mode 동적 입력** (#358) — ProviderModal 이 auth_mode 별 자격증명 필드(token/basic/app_password/oauth2/agent) 렌더 + write-only secret.
5. **SCM repository UI** (#363/#366/#373) — Import/Create 모달 + ProviderTable 액션 + provider_id 단일화 반영.
6. **고정메뉴 Phase 2b** (#362) — known vendor 등록 시 generic select 숨김.
7. **draft repository publish** (#368) — admin catalog/projects/repositories 의 draft→publish UI.

## 7. 프론트엔드 부채

- **단위 테스트 밀도 낮음**: service 18 중 vitest 커버는 `project.service` 1개 + presets. 컴포넌트 단위 테스트는 AuthGuard/IntakeToken 2종뿐.
- **최신 backend 기능 E2E 후행**: repository draft→publish, SCM create/import 의 전용 happy-path E2E 부분적(negative 위주).
- **Command status WS UI 미완**: Phase 4 잔여 — RealtimeService 구독을 command toast/status 에 연결 필요.
- **lint 회귀 패턴**: 신규 코드 누적 시 lint 경고 재증가(과거 18→29→0 복구 이력). pre-commit/CI lint gate 강화 여지.
