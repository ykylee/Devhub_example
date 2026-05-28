# Frontend 분석 — Platform (auth / config / storage / store + 테스트 + 스택)

- 문서 목적: `frontend/lib/{auth,config,storage,store.ts}` 의 플랫폼 레이어와 테스트 자산(Vitest 10 + Playwright 28 spec), package.json 스택을 인벤토리화한다.
- 범위: `frontend/lib/auth`, `lib/config`, `lib/storage`, `lib/store.ts`, `lib/utils.ts`, `tests/`, `package.json`.
- 최종 작성일: 2026-05-27 (main `cf19c94`)
- 관련 문서: `pages.md`, `components.md`, `services.md`

## 1. Auth 레이어 (lib/auth)

### token-store.ts
브라우저 `sessionStorage` 기반 토큰 보관(`devhub_access_token`/`devhub_refresh_token`/`devhub_id_token`). in-memory 캐시 + lazy `ensureLoaded()`(cold 시 sessionStorage fallthrough). `save`/`clear` 가 메모리+스토리지 동기. `getIdToken` 은 RP-initiated logout 의 `id_token_hint` 용. SSR(`typeof window === "undefined"`) no-op.

### pkce.ts (OIDC PKCE flow)
- `createPkceState()` → `{state, codeChallenge, codeChallengeMethod:"S256"}`. random verifier(32B)+state(UUID, 비-secure context 면 수동 UUID), challenge = SHA-256(verifier) base64url.
- **multi-flight 지원**: `oidc_pkce_map`(state→verifier map) 으로 동시 OIDC start 허용. legacy 단일 key(`oidc_state`/`oidc_verifier`)도 backward-compat 유지.
- `consumeVerifier(state)` — map 에서 매칭 verifier 소비(미매칭 시 CSRF 에러 throw), legacy fallback 분기.
- **`sha256Fallback`** — `crypto.subtle` 미가용(plain HTTP non-localhost) 환경용 순수 JS SHA-256. RFC 7636 Appendix B 벡터 준수(테스트로 pin).

### role-routing.ts
- `defaultLandingFor(role)` — System Admin→`/admin`, Manager→`/manager`, Developer/그 외→`/developer`.
- `isSystemAdmin(role)`, `pathRequiresSystemAdmin(path)` — `/admin`·`/admin/*`·`/organization`(legacy) 게이팅. Sidebar/AuthGuard/admin-layout 공유 predicate.

## 2. Config 레이어 (lib/config)

### endpoints.ts
env 기반 주소 해석(native default, docker override 정책). 핵심 export:
- `BASE_PATH`(NEXT_PUBLIC_BASE_PATH 정규화), `API_BASE_URL`(basePath 또는 same-origin).
- `WS_BASE_URL` — 브라우저 protocol(ws/wss)+host+basePath 동적 해석 → `/api/v1/realtime/ws`.
- `OIDC_ISSUER_URL`/`OIDC_AUTH_URL`/`OIDC_REDIRECT_URI`(origin+basePath+`/auth/callback`).
- `getKCAdminConsoleUrl()` — env 우선, issuer pathname 의 realm 추출 fallback, 실패 시 `null`(caller 가 링크 숨김).
- `BACKEND_API_URL_SERVER` — server-only(next.config rewrites).

### mock-ui.ts
`ENABLE_LEGACY_MOCK_UI = false` 단일 상수. projects/[id] 가 import 하나 항상 false → 죽은 분기.

## 3. Storage 레이어 (lib/storage)

### onboardingSkip.ts
`sessionStorage["devhub.onboarding.skipped"]` 기반 `isOnboardingSkipped`/`markOnboardingSkipped`/`clearOnboardingSkip`. AuthGuard 의 limited-mode 3-branch gating 과 OnboardingForm 의 skip 액션이 공유. try/catch 로 private-mode/quota 방어.

## 4. 전역 상태 — store.ts (Zustand)

`useStore` = `create + subscribeWithSelector + persist`. 상태:
- `role`/`actor`(AuthenticatedActor: login/subject/role/source/display_name/email/primary_unit_id/onboarding_required/onboarding_completed_at/review_status) + `setActor`(actor.role→role 동기)/`clearActor`/`setRole`.
- `isDeepFocus`, `notifications`(초기 3 — UI 시드값), toasts(`addToast` 5s 자동 제거), `isLoggingOut`, `isSidebarOpen`/`isSidebarCollapsed`.
- **persist**: `devhub-storage`(localStorage), `partialize` 로 `isLoggingOut`/`toasts`/`isSidebarOpen` 제외(나머지 persist). `subscribeWithSelector` 로 realtime.service 가 identity 변화 구독.

`utils.ts` — `cn`(clsx+tailwind-merge), `formatBytes`.

## 5. 테스트 — Vitest (10 파일)

| 파일 | 시나리오 (1줄 요약) |
|---|---|
| `lib/auth/pkce.test.ts` | createPkceState+consumeVerifier(저장/소비/CSRF mismatch/verifier 누락/overlapping flow) + challengeFromVerifier RFC7636 벡터 + sha256Fallback subtle 일치(50 random) |
| `lib/auth/token-store.test.ts` | save/getter 라운드트립, refresh·id 없을 때 키 제거, clear 메모리+스토리지, cold cache sessionStorage fallthrough |
| `lib/auth/role-routing.test.ts` | defaultLandingFor 3 role + null fallback, isSystemAdmin, pathRequiresSystemAdmin(admin/legacy organization/비-게이트 경로) |
| `lib/services/integration-provider-presets.test.ts` | preset 인벤토리/type 매핑/unknown fallback, composeCredentialsRef 3전략, parseCredentialsRef 라운드트립 + secret 비노출 |
| `lib/services/project.service.test.ts` | getApplications query 직렬화, repo id 기반 project flatten, per-repo 에러 swallow |
| `lib/utils.test.ts` | cn 병합/조건부/tailwind override/undefined·null |
| `app/login/page.test.tsx` | resolveErrorMessage(code/description 우선순위, session_expired/login_failed/unauthorized 매핑, raw fallback) |
| `components/layout/AuthGuard.test.tsx` | loading state, skip gating(default landing 통과 / `/account`→`/onboarding` / skip 미실행 redirect / 완료 통과), 401→session_expired / non-401→login_failed |
| `components/dev-request/__tests__/IntakeTokenTable.test.tsx` | empty state, active/revoked 배지, onRevoke 호출, revokingTokenID 매칭 시 버튼 disable |
| `components/dev-request/__tests__/IssueIntakeTokenModal.test.tsx` | form phase, submit→reveal phase 전환, reveal 중 ESC close 차단 |

setup: `lib/test-setup.ts`(jest-dom matcher + RTL cleanup). 스크립트 `npm test`=`vitest run`.

## 6. 테스트 — Playwright E2E (28 spec, 각 1줄)

| spec | 시나리오 |
|---|---|
| `auth.spec.ts` | role별 landing(developer/manager/admin) + developer `/admin/settings` 바운스 + 오답 비밀번호 시 sign-in 유지 |
| `rbac-routes.spec.ts` | developer 가 모든 `/admin/settings/*` 바운스 + manager `/admin` 바운스 |
| `signout.spec.ts` | Sign Out 후 재로그인 자격증명 요구 + 보호 라우트 직진입 `/login` 바운스 + 사용자 전환(alice→bob) |
| `signup.spec.ts` | `/auth/signup` migration notice + sign-in 링크 |
| `account.spec.ts` | `/account` 프로필 info 노출 + Open Keycloak Console 버튼이 issuer/account/ 로 |
| `onboarding-first-login.spec.ts` | token-only actor keycloak 로그인 → `/onboarding` landing |
| `header-switch-view.spec.ts` | Header dropdown → Account Profile → `/account` 이동 |
| `dashboard-retry-empty-state.spec.ts` | developer/manager/admin metrics 실패 시 retry 노출 |
| `dev-requests.spec.ts` | Intake→Promote→Revoke lifecycle + invalid bearer 401 + token allowed_ips PATCH + revoke Cancel 유지 + Intake→Promote→Project |
| `admin-users-crud.spec.ts` | Invite Member 모달 open/close + Role select 옵션 |
| `admin-users-search.spec.ts` | users 검색(name 부분일치+복귀 / email / role) |
| `admin-org-crud.spec.ts` | Org 리스트/검색/부서 생성 모달/멤버 관리 모달/Org Chart 뷰 전환 |
| `admin-permissions.spec.ts` | permission 진입+role 선택+matrix + Custom role 생성·권한편집·저장 |
| `admin-applications.spec.ts` | Applications 탭 진입+New 버튼/모달 open-close/필수값 검증/검색 |
| `admin-projects.spec.ts` | Projects 진입+New 버튼/모달/필수값/CRUD(N:M 연결) |
| `admin-project-model-v2.spec.ts` | application→project→project_repositories(N:M) v2 flow |
| `project-model-modes.spec.ts` | legacy/v2 route 가용성이 DEVHUB_PROJECT_MODEL 따름 |
| `admin-integrations.spec.ts` | provider lifecycle(LIST/CREATE/EDIT/SYNC) + non-admin redirect |
| `admin-integration-bindings.spec.ts` | bindings lifecycle(LIST+CREATE) + non-admin redirect |
| `admin-catalog.spec.ts` | system_admin 3탭 전환+검색 + Applications 드릴다운 + non-admin 차단 |
| `admin-topology-v2.spec.ts` | system_admin 노드/서비스/snapshot 메타 렌더 + nav link + non-admin redirect |
| `infra-topology.spec.ts` | system_admin `/admin` topology 뷰 mount |
| `audit.spec.ts` | system_admin audit 탭 + 로그 entry |
| `repositories-ui.spec.ts` | 저장소 목록 진입 + 검색 필터 + 상세 활동 카드 |
| `repositories-ui-negative.spec.ts` | 목록 조회 실패 retry + 빈 목록 empty state |
| `repositories-detail-negative.spec.ts` | activity 실패 시 에러+retry |
| `applications-projects-detail-negative.spec.ts` | rollup 실패 retry + activity 부분 실패 경고 배너 |
| `screenshots.spec.ts` | 디자인 리뷰용 캡처(admin 12 / login / user 6 / manager 1) |

## 7. 스택 (package.json)

- 런타임: next 16.2.6, react/react-dom 19.2.4, zustand 5, @tanstack/react-query 5(의존성 등록), @xyflow/react 12 + dagre(React Flow topology/org-tree), recharts 3, framer-motion 12, date-fns 4, lucide-react, clsx + tailwind-merge.
- 빌드/테스트: tailwindcss 4(+postcss), typescript 5, eslint 9 + eslint-config-next, vitest 4 + @vitest/coverage-v8 + jsdom + @testing-library(react/jest-dom/user-event), @playwright/test 1.59.
- 스크립트: `dev`/`build`/`start`/`lint`/`test`(vitest run)/`test:coverage`/`e2e`(playwright)/`e2e:ui`/`e2e:report`.
- 주의: `frontend/AGENTS.md` 가 "이 Next.js 는 breaking change 가 있다 — 코드 작성 전 `node_modules/next/dist/docs/` 확인" 경고.

## 발견 사항 (불일치/stale/부채)

- **테스트 커버리지 편중**: Vitest 10 파일이 auth(pkce/token/role) + 순수함수 + UI 모달 2개에 집중. service 18 중 단위테스트 2개(services.md 참조). E2E 28 spec 도 admin/auth/dev-request/repository 에 집중하고 최신 SCM import·create·repository draft/publish happy-path 미커버.
- **`notifications` 초기값 3 시드**: `store.ts:59` 가 `notifications: 3` 하드코딩 — 실 알림 수가 아닌 디자인 시드값. Header 가 mount 시 `fetchDreqs()` 로 덮어쓰지만(`Header.tsx:54`) 초기 paint 에 가짜 배지 3 노출.
- **`ENABLE_LEGACY_MOCK_UI` dead 상수**: `config/mock-ui.ts` 가 false 단일 export, projects/[id] 만 import → 죽은 분기 + import 잔재.
- **legacy WS endpoint 가 endpoints.ts 에 잔존**: `WS_BASE_URL`(endpoints.ts:34)은 ticket 기반 realtime.service 가 쓰지만, `websocket.service.ts` 도 같은 base 에 `?access_token=` 를 붙임 → 컷오버 미완(services.md 참조).
- **lint disable 회귀 패턴**: `eslint-disable` 8 파일 9건(set-state-in-effect 5, exhaustive-deps 등). React 19/Next 16 의 set-state-in-effect 룰을 mount-only effect 마다 disable 주석으로 우회(`admin/settings/layout.tsx:72`, `projects/page.tsx:49`, `admin/catalog/page.tsx:76`, `Header.tsx:72`, `admin/topology-v2/page.tsx:202`). 직전 lint cleanup(PR #332, 29→0) 이후 Onboarding/Project/SCM 누적으로 재발 가능 — useSyncExternalStore(Sidebar.tsx:42) 같은 정공법 미적용 잔여.
- **AGENTS.md breaking-change 경고 vs 검증 부재**: "this is NOT the Next.js you know" 경고가 있으나 코드 주석/문서에 어떤 API 가 달라졌는지 추적이 없어, 신규 작업자가 stale 가정으로 진입할 위험.
