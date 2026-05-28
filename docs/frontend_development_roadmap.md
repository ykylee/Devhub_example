# Frontend Development Roadmap (Phase 2+)

> ⚠ **먼저 [통합 개발 로드맵](./development_roadmap.md)을 확인하세요.** 본 문서는 그 통합 로드맵의 Frontend 트랙 세부입니다. 마일스톤(M0~M4) / 우선순위(P0~P3) / 트랙 간 의존은 통합 로드맵의 §3·§4.2 가 source-of-truth.
>
> ⚠ **v1.0/v1.1 신규 작업의 source-of-truth = [`docs/planning/release_v1_roadmap.md`](./planning/release_v1_roadmap.md).** 본 문서는 frontend phase 이력 + 잔여 추적용이며, 우선순위·마일스톤의 최신 기준은 release_v1_roadmap 이다. 현행 코드베이스 전수 분석은 [2026-05-27 codebase snapshot](./analysis/2026-05-27-codebase-snapshot/03_frontend_summary.md) 참조.

- 문서 목적: 백엔드 API 연동 및 프론트엔드 기능 고도화 로드맵을 정의한다.
- 범위: 프론트엔드 트랙의 phase별 작업 계획, 상태, 연동 의존성.
- 대상 독자: 프론트엔드 개발자, QA, 프로젝트 리드, 후속 작업자.
- 기준일: 2026-05-04
- 최종 수정일: 2026-05-27 (Keycloak 단일 IdP 전환(ADR-0019) 반영 + 현행 도메인 완성 정정 — 2026-05-27 codebase snapshot 기준. §6 자체 계정 인증 서술을 Keycloak Account Console 위임으로 정정, Phase 4/7 현황 정밀화)
- 상태: draft
- 관련 문서: [`./development_roadmap.md`](./development_roadmap.md) (통합), [`./planning/release_v1_roadmap.md`](./planning/release_v1_roadmap.md) (v1.0/v1.1 source-of-truth), [`./analysis/2026-05-27-codebase-snapshot/03_frontend_summary.md`](./analysis/2026-05-27-codebase-snapshot/03_frontend_summary.md) (현행 frontend 전수 분석), `docs/backend_api_contract.md`, `ai-workflow/memory/backend_development_roadmap.md`, [`./adr/0019-keycloak-only-idp.md`](./adr/0019-keycloak-only-idp.md) (Keycloak 단일 IdP)

## 1. 개요

프론트엔드 Phase 1에서 구축된 UI와 서비스 레이어 구조를 바탕으로, Phase 2 이후에는 백엔드 API와의 실시간 데이터 연동, 조직/계정 관리, 관리자 액션을 순차적으로 프로덕션 수준으로 끌어올린다. AI 어드바이저/AI 가드너는 v2에서 다룬다. 역할별 UX는 전용 화면 완전 분리보다 역할별 기본 진입 우선순위로 간접 제공한다.

## 2. Phase 로드맵

| Phase | 상태 | 목표 | 주요 작업 |
| --- | --- | --- | --- |
| **Phase 1** | **done** | 대시보드 UI & Mock 서비스 | 레이아웃 구축, Glassmorphism 적용, Singleton Service 패턴 도입 |
| **Phase 2** | **done** | 핵심 API 통합 | Infra Topology, Risk List, Command/Audit 연동, role wire mapping |
| **Phase 3** | **done** | 실시간성 및 CI/CD 가시화 | WebSocket 통합, CI Run/Logs 연동, 실시간 알림 피드 |
| **Phase 4** | **부분 (command status WS UI 잔여)** | 어드민 액션 고도화 | 시스템 관리자 서비스 제어 액션 실체화 ✅. **잔여: RealtimeService 구독 → command toast/status UI 연결** (Phase 4 잔여, M4) |
| **Phase 5** | **done** | 사용자 및 조직 관리 UI | 사용자 프로필, 팀/조직 단위(Org Units) 관리 UI, 멤버 할당 모달 |
| **Phase 5.1** | **done** | 조직 관리 API 통합 | 백엔드 조직 CRUD 및 멤버 할당 API 연동 |
| **Phase 5.2** | **done** | 계정 인증 및 IdP 도입 | **Keycloak OIDC 단일 IdP(ADR-0019)** 기반 인증 도입(PKCE + token refresh + AuthGuard). `/account`(ProfileSelfEdit + Account Console redirect), `/admin/settings/{users,organization,permissions}`. ~~자체 계정 발급/비밀번호 흐름~~ 은 Keycloak Account Console 위임으로 폐기. |
| **Phase 6** | **done** | 권한 관리(RBAC) UI 고도화 | PermissionEditor `/admin/settings/permissions` 완료. |
| **Phase 6.1** | **done** | RBAC API 통합 | `/api/v1/rbac/policies` 조회/편집 연동, `requirePermission` 라우트 가드 (M1 RBAC track). |
| **Phase 7** | **done** | 조직 관리 1차 완성 | 부서 CRUD, 계층 편집, 전역 감사 로그 연동 완료(`/admin/settings/{organization,users,permissions,audit}`). |
| **Phase 8 (신규)** | **done** | 도메인 페이지 + 운영 UI 전환 | Application/Project/Repository(draft·SCM import/create) + DREQ + External Integration(auth_mode 동적 입력) + Onboarding + topology v2 + admin catalog + 운영 UI 전환(mock 제거 + `PageState` + 에러 표준화). 03 snapshot §1 참조. |


## 3. Phase 2 상세 계획 (Core API Integration)

### 3.1 Infra Topology 연동

- **목표**: 인프라 그래프가 실제 런타임 상태를 반영하도록 함.
- **API**: `GET /api/v1/infra/topology`
- **핵심 로직**:
    - 백엔드의 `RuntimeSnapshotProvider`에서 제공하는 Node 상태(`stable`, `warning`, `down`)를 UI 색상 및 애니메이션으로 변환.
    - `cpu_percent`, `memory_bytes` 등 원시 데이터를 프론트엔드 유틸리티를 통해 포맷팅.

### 3.2 Risk Management 연동

- **목표**: 리스크 목록을 DB에서 가져오고, 리스크 완화 조치를 감사 가능한 형태로 처리.
- **API**: `GET /api/v1/risks`, `POST /api/v1/risks/{id}/mitigations`
- **핵심 로직**:
    - 리스크 완화 버튼 클릭 시 `idempotency_key`를 생성하여 중복 요청 방지.
    - 비동기 `command_id`를 수신하여 처리 중 상태를 UI에 표시.

### 3.3 Dashboard Metrics 연동

- **목표**: 역할별 기본 진입 페이지에서 보여줄 메트릭 카드를 실제 통계 데이터로 업데이트.
- **API**: `GET /api/v1/dashboard/metrics?role={role}`

## 4. 기술적 고려 사항

- **Error Handling**: API 호출 실패 시 Toast 메시지 노출 및 Graceful Degradation 적용.
- **Loading States**: `Framer Motion`을 활용한 스켈레톤 UI 및 로딩 애니메이션 유지.
- **Environment**: `NEXT_PUBLIC_API_URL`, `NEXT_PUBLIC_WS_URL` 환경 변수로 백엔드 주소 관리.

## 5. Phase 4 상세 계획 (어드민 액션)

- `infra.service.ts`, `risk.service.ts` 실제 API 연동은 완료.
- System Admin service action command 생성 연동은 완료.
- 백엔드 `/api/v1/realtime/ws`(**ticket 인증 — ADR-0024, `POST /api/v1/realtime/ticket` → `?ticket=`**)와 `command.status.updated` publish 경계가 구현됨. 프론트 `realtime.service`(ticket WS pub/sub)도 구현됨.
- **잔여(Phase 4)**: 기존 `RealtimeService` 구독을 **command toast/status UI 에 연결**하는 것. (infra/ci/risk event 구독 UI 는 backend RM-M4-01 publish 완성 후 = v2)
- AI Gardener suggestion API/UI 연결은 v2로 이관(backend-ai 스켈레톤 의존).

## 6. Phase 5 상세 계획 (사용자/조직 관리 + 인증)

> ⚠ **2026-05-27 정정 배너 (ADR-0019 Keycloak 단일 IdP 전환):** 본 §6 의 **자체 계정(Account) 1:1 + `/api/v1/accounts/*` 발급/잠금/재설정 + `must_change_password`** 전제는 **모두 폐기**됐다(historical). 현재는 **Keycloak OIDC 단일 IdP**가 credential master 이고, 비밀번호/계정 lifecycle 은 **Keycloak Account Console 에 위임**한다. 아래 6.1~6.3 의 원문은 historical 기록 보존용이며, 각 항목 끝에 **[현행]** 정정을 병기한다.

DevHub ~~자체 사용자 계정(Account) 1:1 컨셉~~(historical)에 따른 초기 프론트 계획. **현행: Keycloak OIDC 위임 + DevHub `users` 매핑(`GET /api/v1/me`).**

### 6.1 로그인 / 인증 흐름

- ~~**목표:** `/login` 진입점 + 인증 가드 + `must_change_password=true` 라우팅.~~ (historical)
- ~~**API:** Keycloak OIDC + `GET /api/v1/me`.~~
- ~~**핵심 로직:** 로그인 폼 `login_id` + `password` / `must_change_password=true` → `/account/password` 강제 라우팅.~~
- **[현행]** `/login` 진입점 + AuthGuard + **Keycloak OIDC PKCE**(IdP hosted login, `auth.service`) + `/auth/callback` token 영속화 + `api-client` 401 자동 refresh. **`must_change_password` 폐기** — 비밀번호 정책은 Keycloak realm 이 관리. `GET /api/v1/me` 로 DevHub actor/role 매핑.

### 6.2 내 계정 화면 (`/account`)

- ~~**목표:** 본인 로그인 ID 확인 + 비밀번호 변경.~~ (historical)
- ~~**API:** `GET /api/v1/accounts/{user_id}`, `PUT /api/v1/accounts/{user_id}/password`.~~
- ~~**핵심 로직:** login ID 읽기 전용 + 비밀번호 변경 폼 3 필드.~~
- **[현행]** `/account` 는 **ProfileSelfEdit**(표시명/온보딩 read-only self-reverse) 중심이다. **비밀번호 변경은 Keycloak Account Console redirect**(`auth.service` 의 Account Console URL)로 위임 — DevHub 는 비밀번호 폼을 직접 제공하지 않는다. `/api/v1/accounts/*` 폐기.

### 6.3 시스템 관리자 계정 관리 패널

- ~~**목표:** 사용자 row 옆 계정 발급/회수/잠금 해제/강제 재설정.~~ (historical)
- ~~**API:** `POST/PATCH/DELETE /api/v1/accounts`, `PUT /api/v1/accounts/{user_id}/password` (`force=true`).~~
- ~~**핵심 로직:** "계정 발급" 다이얼로그 + 임시 비밀번호 1회 표시 + 회수 확인 다이얼로그.~~
- **[현행]** 계정 발급/비밀번호 lifecycle 은 **Keycloak**(IdP 팀/admin)이 담당한다. DevHub 의 사용자 운영은 `/admin/settings/users`(users CRUD + role) + **Onboarding admin review**(`ConfirmReviewModal`·`PendingReviewPanel`, `POST /api/v1/admin/users/:id/review` = API-86)로 대체. `/api/v1/accounts/*` admin endpoints 폐기.

### 6.4 조직 관리 후속

- ✅ Organization UI 와 read/write API 는 완성됨(`/admin/settings/organization` + OrgTree/units/members + single-leader invariant).
- ~~사용자 신규 등록과 계정 발급을 묶는 합성 액션~~ → Keycloak 위임으로 불필요. 사용자 신규 진입은 **Onboarding self-service(ADR-0021)** 흐름.

## 7. 다음 작업 큐

> ⚠ **2026-05-27 정정**: 아래 7.1 의 M2 sprint(`claude/login_usermanagement_finish`)는 종결됐다(historical). 백엔드 짝 PR-M2-AUDIT(Kratos webhook)도 Keycloak 전환으로 폐기됨. 현행 잔여는 **7.0** 가 source-of-truth.

### 7.0 현행 잔여 큐 (2026-05-27, main `cf19c94`)

- [ ] **단위테스트 보강 (1순위, B1)** — service 18개 중 vitest 커버 2개뿐(`project.service`, `integration-provider-presets`). integration·repository·dashboard service + 신규 모달(Import/CreateScmRepository 등) 단위테스트 최소 6종 추가.
- [ ] **command status WebSocket UI (Phase 4 잔여)** — `RealtimeService` 구독 → command toast/status 연결.
- [ ] **최신 backend 기능 happy-path E2E (B2)** — repository draft→publish, SCM import/create(현재 negative-path 위주).
- [ ] **AI Gardener suggestion UI (v2)** — backend-ai gRPC AnalysisService 실구현 의존.
- [ ] **System Admin 운영 가시성 (RM-M4-07)** — Gitea sync job 큐/provider health view(backend 데이터는 있으나 운영 화면 없음).

세부 잔여·우선순위는 [05_fe_be_balance.md](./analysis/2026-05-27-codebase-snapshot/05_fe_be_balance.md) §4 + [06_future_direction.md](./analysis/2026-05-27-codebase-snapshot/06_future_direction.md) §3 참조.

### 7.1 ~~M2 1차 완성 sprint~~ (historical — 종결)

- [x] ~~PR-UX1 — `/admin/settings/users` SearchInput 실 필터링~~
- [x] ~~PR-UX2 — `/account` 재인증 세션 안내~~
- [x] ~~PR-UX3 — Header Switch View 한계 안내~~
- ~~백엔드 짝(PR-M2-AUDIT)~~ → Keycloak event polling 으로 대체(폐기).

### 7.2 완료 (참고)

- [x] 역할별 기본 진입 우선순위 라우팅 (`defaultLandingFor`)
- [x] 로그인 직후 역할 기반 기본 진입 + Header Switch View
- [x] `/login` 페이지 + AuthGuard (canonical `/login`, sub-carve F — `/auth/login` 폐기)
- [x] `/auth/callback` + tokenStore 영속화 + `api-client` 401 자동 refresh
- [x] `/account` ProfileSelfEdit + **비밀번호는 Keycloak Account Console redirect**(자체 비밀번호 변경 폼 폐기)
- [x] Header Sign Out → Keycloak OIDC 세션 종료
- [x] 온보딩(OnboardingForm/Banner/OrganizationPicker) + 도메인 페이지(Application/Project/Repository/DREQ/Integration) + topology v2 + admin catalog + 운영 UI 전환(mock 제거 + PageState)

### 7.3 후속 (별도 sprint)

- [ ] Organization 페이지의 사용자 운영 action 통합 — 현재는 `/admin/settings/users` + Onboarding admin review 로 분리 운영 (재정합 필요 시 별도 sprint)
