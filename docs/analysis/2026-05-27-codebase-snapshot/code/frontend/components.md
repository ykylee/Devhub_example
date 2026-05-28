# Frontend 분석 — Components

- 문서 목적: `frontend/components/` 의 50 파일을 폴더(account/admin/dashboard/dev-request/integration/layout/onboarding/organization/project/ui)별로 역할·사용처 기준으로 인벤토리화한다.
- 범위: `frontend/components/` (테스트 `*.test.tsx` 4건은 frontend_platform.md 에 기재). 모든 컴포넌트는 `"use client"`.
- 최종 작성일: 2026-05-27 (main `cf19c94`)
- 관련 문서: `pages.md`, `services.md`, `frontend_platform.md`

## 1. ui/ (공통 프리미티브 — 11)

| 컴포넌트 | 역할 | 사용처 |
|---|---|---|
| `PageState.tsx` | `PageLoading`/`PageError(onRetry)`/`PageEmpty` 3 export — 페이지 로딩·에러·빈 상태 공통화 | applications/projects/repositories 목록·상세, admin/catalog 등 운영 페이지 전반 |
| `Toast.tsx` | `ToastContainer`(root layout) + `useToast()` hook(toast/success/error/warning/info) | root layout + 거의 모든 admin/모달 페이지 (store.addToast 래핑) |
| `Modal.tsx` | 범용 모달 shell(size/title/onClose, ESC/backdrop) | topology-v2 node detail 등 |
| `DestructiveConfirmModal.tsx` | 파괴적 액션 확인(typed confirm) | intake token revoke, integration provider/binding delete |
| `FilterBar.tsx` | 검색 input + status select 콤보 | applications/projects/repositories/dev-requests/admin users·dev-requests |
| `DashboardHeader.tsx` | titlePrefix+gradient+subtitle 헤더 | 운영 도메인 목록 페이지 공통 헤더 |
| `Badge.tsx` | status 배지(variant success/warning/danger/secondary/glass, dot) | 테이블·상세 전반 |
| `ComboBox.tsx` | 검색형 select(autocomplete) | 사용자/단위 선택 모달 |
| `ActionMenu.tsx` | row-level dropdown 액션 메뉴 | 테이블 행 액션 |
| `LogoutOverlay.tsx` | `isLoggingOut` store 구독 전체 화면 오버레이 | root layout |
| `Badge`/`ComboBox` 등은 lint cleanup(PR #332) 에서 onClick/discriminator 보강 대상이었음 | — | — |

## 2. layout/ (4, 테스트 1 별도)

| 컴포넌트 | 역할 | 사용처 |
|---|---|---|
| `AuthGuard.tsx` | dashboard 진입 가드 — whoAmI → store 갱신 + onboarding 3-branch + system-route 가드 + WS 구독(레거시 websocketService) | `(dashboard)/layout.tsx` |
| `Header.tsx` | 상단바 — 실시간 연결 표시(realtimeService `status.changed`), 알림 dropdown(pending DREQ), user 메뉴(theme/account/settings/logout), DREQ 상세→Project 승격 모달 | dashboard layout |
| `Sidebar.tsx` | 좌측 nav — base menu + `showSystem`(isSystemAdmin) 조건 system menu, collapse, useSyncExternalStore 로 SSR hydration | dashboard layout |
| `AuthGuard.test.tsx` | (vitest) | — |

## 3. account/ (1)

| 컴포넌트 | 역할 | 사용처 |
|---|---|---|
| `ProfileSelfEdit.tsx` | self-service display_name/primary_unit 편집(onboardingService.patchMe) + review_status 안내 | `/account` page |

## 4. admin/users/ (2)

| 컴포넌트 | 역할 | 사용처 |
|---|---|---|
| `PendingReviewPanel.tsx` | onboarding pending_review 멤버 목록 패널 | admin/settings/users page |
| `ConfirmReviewModal.tsx` | 멤버 검토 확정(onboardingService.confirmUserReview) | PendingReviewPanel |

## 5. dashboard/ (1)

| 컴포넌트 | 역할 | 사용처 |
|---|---|---|
| `GardenerFeed.tsx` | AI 제안 피드(gardenerService.getSuggestions/applySuggestion) | `/gardener` page |

## 6. dev-request/ (8, 테스트 2 별도)

| 컴포넌트 | 역할 | 사용처 |
|---|---|---|
| `DevRequestTable.tsx` | 의뢰 목록 테이블(상태 배지, onSelect) | /dev-requests, admin/settings/dev-requests |
| `DevRequestDetailModal.tsx` | 의뢰 상세 + 액션(view/register/reject/reassign). `isSystemAdmin` 으로 액션 게이팅, `onPromote` 콜백으로 Project 승격 위임 | /dev-requests, admin/settings/dev-requests, Header |
| `MyPendingDevRequestsWidget.tsx` | 담당자 "내 대기 의뢰" 위젯(getMyPending) | developer dashboard |
| `IntakeTokenTable.tsx` | intake token 테이블(active/revoked 배지, revoke/edit, revokingTokenID disable) | admin/settings/dev-request-tokens |
| `IssueIntakeTokenModal.tsx` | token 발급 — form→reveal 2 phase, plain 토큰 1회 노출(reveal 중 ESC 차단) | admin/settings/dev-request-tokens |
| `EditIntakeTokenModal.tsx` | token allowed_ips/메타 편집(updateIPs) | admin/settings/dev-request-tokens |
| `__tests__/IntakeTokenTable.test.tsx`, `__tests__/IssueIntakeTokenModal.test.tsx` | (vitest) | — |

## 7. integration/ (8)

| 컴포넌트 | 역할 | 사용처 |
|---|---|---|
| `ProviderModal.tsx` | provider 등록/편집 — vendor 템플릿(7종) + **auth_mode 동적 입력** + capability 체크박스 + base_url + 연결 테스트 + webhook 자격증명 분리 입력 | admin/settings/integrations |
| `ProviderTable.tsx` | provider 목록(sync_status, edit/sync/delete + scm+pull 시 import/create-repo) | admin/settings/integrations |
| `ImportRepositoriesModal.tsx` | SCM 원격 repo 목록(API-88) + 선택 import(API-89) | admin/settings/integrations |
| `CreateScmRepositoryModal.tsx` | SCM 에 신규 repo 생성(API-90, push+gitea-compat) | admin/settings/integrations |
| `BindingsTable.tsx` | provider↔scope binding 목록 | admin/settings/integration-bindings |
| `CreateBindingModal.tsx` | binding 생성(scope/provider/external_key/policy) | admin/settings/integration-bindings |
| `EditBindingModal.tsx` | binding 편집(policy/enabled) | admin/settings/integration-bindings |

### ProviderModal auth_mode 동적 입력 (핵심 로직)
`authMode`(token/basic/oauth2/app_password/agent)에 따라 outbound auth 필드가 전환된다(`ProviderModal.tsx:175,429-519`):
- `token` → API Token(SecretField, write-only blank=keep).
- `basic`/`app_password` → Username + Password/App Password(secret).
- `oauth2` → Client ID + Token URL + Client Secret(secret).
- `agent` → Agent Identifier(secret 없음).
secret 류는 모두 write-only(`api_token`/`auth_secret`, blank=keep, `*_set` bool 로 설정 여부 표시). vendor 템플릿(known vendor) 선택 시 type/auth/signature 가 고정되고 generic select 는 read-only 요약으로 숨겨짐(고정메뉴 Phase 2b). webhook credentials 는 strategy+vendor+secret → `composeCredentialsRef` 로 조합(내부 인코딩 직접 입력 회피).

## 8. onboarding/ (3)

| 컴포넌트 | 역할 | 사용처 |
|---|---|---|
| `OnboardingForm.tsx` | 프로필 등록(display_name+unit) submit / skip(markOnboardingSkipped). ApiError code 별 한국어 메시지 | /onboarding page |
| `OnboardingBanner.tsx` | limited-mode dismissible 배너(skip 후 미완료 사용자) | dashboard layout |
| `OrganizationPicker.tsx` | 조직 단위 선택(search + tree, onboardingService.searchOrganizations) | OnboardingForm, ProfileSelfEdit |

## 9. organization/ (13)

| 컴포넌트 | 역할 | 사용처 |
|---|---|---|
| `MemberTable.tsx` | 사용자 목록 테이블(role 변경, create/edit/delete 모달 연결) | admin/settings/users |
| `MemberManagementModal.tsx` | 단위 멤버 배정 관리(replaceUnitMembers) | admin/settings/organization |
| `OrgUnitTable.tsx` / `OrgUnitGrid.tsx` | 조직 단위 list/grid 뷰 | admin/settings/organization |
| `OrgTree.tsx` | React Flow 조직도(chart 뷰, 노드 레이아웃) | admin/settings/organization |
| `OrgNode.tsx` | React Flow 커스텀 노드(memo) | OrgTree |
| `UnitManagementModal.tsx` | 단위 생성/편집(createUnit/updateUnit) | admin/settings/organization |
| `UserCreationModal.tsx` | 사용자 생성(HR lookup autofill) | MemberTable |
| `UserEditModal.tsx` | 사용자 편집(role/status/unit) | MemberTable |
| `PermissionEditor.tsx` | RBAC role 선택 + 매트릭스 + custom role 생성/저장 | admin/settings/permissions |
| `PermissionMatrix.tsx` | resource×permission 매트릭스(readOnly/lockedCells, deep-clone toggle) | PermissionEditor |

## 10. project/ (6)

| 컴포넌트 | 역할 | 사용처 |
|---|---|---|
| `ApplicationTable.tsx` | application 목록(edit/archive/repo 연결) | admin/settings/applications |
| `ApplicationCreationModal.tsx` | application 생성(initialData prefill) | admin/settings/applications, admin/catalog |
| `ProjectTable.tsx` | project 목록 | admin/settings/applications |
| `ProjectCreationModal.tsx` | project 생성 — standalone/application-scoped, repository_ids N:M + repository_create_payload | admin/settings/applications, admin/catalog, Header(DREQ 승격) |
| `RepositoryTable.tsx` | application 연동 repository 목록(disconnect) | admin/settings/applications |
| `RepositoryLinkModal.tsx` | application↔repository 연결(connectRepository) | admin/settings/applications |

## 발견 사항 (불일치/stale/부채)

- **컴포넌트 단위테스트 희박**: 50개 중 vitest 가 있는 건 dev-request 2개(IntakeTokenTable/IssueIntakeTokenModal) + AuthGuard 뿐. ProviderModal(auth_mode 5분기 + write-only secret + vendor 템플릿 — 가장 복잡)·PermissionMatrix(deep-clone toggle)·OrgTree·DevRequestDetailModal(4 mode 액션) 등 분기 많은 컴포넌트가 무테스트.
- **GardenerFeed mock fallback 의존**: gardenerService.getSuggestions 가 에러 시 하드코딩 mock 반환(services.md) → GardenerFeed 가 백엔드 장애를 정상 제안으로 렌더. gardener 페이지 우측 stat(94/100 등)도 하드코딩(pages.md).
- **Header WS UI 의 stale 구독**: `Header.tsx:81` 의 `dev_request.created` 구독이 realtime DEFAULT_EVENT_TYPES 에 없어 영구 미발화(services.md). 알림 dropdown 갱신은 `status.changed`(reconnect) 시점에만 fetchDreqs 로 이뤄짐.
- **AuthGuard 가 레거시 websocketService 사용**: 컴포넌트 레이어에서 ticket 기반 realtimeService 가 아닌 `websocket.service`(`?access_token=`)를 connect(`AuthGuard.tsx:96`) → ADR-0024 ticket-only 컷오버 미반영(services.md 중복 항목).
- **command status WS 컴포넌트 부재 (Phase 4 잔여)**: service-action(infra/gardener/risk)의 비동기 command 진행을 표시하는 컴포넌트가 없다. topology-v2 의 "Live WS Active" 패널은 연결 표시일 뿐 command 상태 추적 아님.
- **repository draft/publish 전용 컴포넌트 없음**: 최신 #368/#373 draft→publish lifecycle 의 UI 가 별도 모달 없이 admin/catalog 의 `prompt()` 로만 존재(pages.md) → integration/CreateScmRepositoryModal 같은 정식 모달 패턴 미적용.
- **lint disable 산재 컴포넌트**: Header.tsx(set-state-in-effect), OrgTree.tsx(exhaustive-deps) 등 disable 주석. Sidebar 는 useSyncExternalStore 정공법을 쓰는데 Header/layout 은 disable 우회로 일관성 없음(frontend_platform.md).
- **organization mock 헬퍼 — 일부 활성 / 일부 dead**: `identityService.mockHierarchy()`(identity.service.ts:479)는 OrgTree.tsx:355 에서 fallback 으로 호출됨(실 데이터 실패 시 mock 조직도 렌더 — 운영 UI 전환 위배). 반면 `mockUsers()`(identity.service.ts:431)는 `private` 이고 호출처가 없어 **dead code**.
