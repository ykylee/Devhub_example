# ADR-0021: Onboarding self-service unit selection + lazy auto-create supersession

## 1. 상태
- **상태**: Accepted
- **작성일**: 2026-05-21
- **수정일**: 2026-05-21
- **결정 근거 sprint**: `claude/keycloak-user-onboarding-concept` (PR #260, 컨셉 1차) + `claude/keycloak-onboarding-concept-2026-05-21` (PR #265, §5.9 skip-and-resume) + `claude/onboarding-requirements-2026-05-21` (PR #266, REQ §5.7) + `claude/onboarding-arch-2026-05-21` (PR #267, ARCH §9 + API §16) + `claude/onboarding-adr-2026-05-21` (본 ADR 발급)
- **partial supersedes**: [ADR-0020 외부 Keycloak 가정 하의 계정/사용자 관리 책임 경계 (2026-05-20)](./0020-account-user-management-boundary.md) — §3.2 의 "신규 user 의 unit 초기 배치" row 의 lazy-auto-create-후 정책 + §4.1 sub-carve B 의 lazy auto-create 실 구현 결정 + §4.2 의 lazy auto-create 관련 보안 영향 + §6.2 carve out 의 동일 항목. ADR-0020 의 **핵심 결정** (옵션 A 책임 경계 — Keycloak account vs DevHub user 분리) 은 reverse 하지 않고 **자연 확장**.
- **관련 문서**: [`docs/planning/keycloak_user_onboarding_concept.md`](../planning/keycloak_user_onboarding_concept.md), [`docs/requirements.md §5.7`](../requirements.md), [`docs/architecture.md §9`](../architecture.md), [`docs/backend_api_contract.md §16`](../backend_api_contract.md), [`docs/planning/system_usecases.md §2.13`](../planning/system_usecases.md), [ADR-0019 Keycloak 단일화](./0019-keycloak-only-idp.md), [ADR-0020 계정/사용자 책임 경계](./0020-account-user-management-boundary.md)

## 2. 컨텍스트

### 2.1 ADR-0020 후속 — DevHub user 의 "첫 진입" UX 단절

[ADR-0020 (2026-05-20)](./0020-account-user-management-boundary.md) 가 외부 Keycloak 시나리오에서의 책임 경계를 옵션 A (자체 accounts 테이블 폐기 + Keycloak 단일 IdP) 로 확정하면서, 신규 사용자의 첫 진입 시점에 **lazy auto-create** 흐름을 도입 (sub-carve B, PR #239) 했다.

- `authenticateActor` 가 Keycloak token 검증 후 `GetUser` miss 시 자동으로 `users` row 를 생성 (`primary_unit_id=NULL`)
- 사용자가 인지 없이 시스템 진입 가능 — 첫 API call 차단 없음

이 흐름은 ADR-0020 의 "운영 거버넌스 명확화" + "service account 권한 최소화" 결정과 정합했으나, **사용자 UX 측면의 3개 gap** 이 후속 검토에서 발견됐다 ([`docs/planning/keycloak_user_onboarding_concept.md` §2.2](../planning/keycloak_user_onboarding_concept.md)):

1. **소속 미배정 상태가 사용자 인지 없이 무기한 잔존** — UI 의 user list 에 "소속 없음" 으로만 표기. lazy-created user 가 자신의 소속을 시스템에 알릴 self-service 경로 부재.
2. **`display_name` 정확성 의존** — Keycloak `name` claim 비어있으면 login (user_id) fallback. 사용자가 자신의 표시명을 명시적으로 확인/수정한 적 없음.
3. **사용자 첫 진입 UX 단절** — Keycloak 로그인 → 대시보드 진입 → 자신의 소속이 "미배정" 인 상태로 사용 시작. 어느 admin 에게 요청해야 unit 이 배정되는지 알기 어려움.

### 2.2 컨셉 → 요구사항 → 설계 — 3 phase 누적 결정

본 ADR 은 다음 3 phase 의 결정을 종합 명문화한다:

- **컨셉 phase** (PR #260 + PR #265): 9 sub-section (§5.1~§5.9) + 12 open question (§8 #1~#12) 결정. 마지막 open question (§8 #7 skip-and-resume) 까지 PR #265 에서 종결.
- **요구사항 phase** (PR #266): `docs/requirements.md` §5.7 신규 — REQ-FR-ONBOARD-001..012 + REQ-NFR-ONBOARD-001..008 (20 row).
- **설계 phase** (PR #267): `docs/planning/system_usecases.md` §2.13 (UC-ONBOARD-01..11) + `docs/architecture.md` §9 (ARCH-ONBOARD-01..06) + `docs/backend_api_contract.md` §16 (API-83..86 신규 + API-32 / API-33 확장).

### 2.3 ADR 발급 필요성

설계 산출물이 ADR-0020 의 보조 결정 (lazy auto-create) 을 **부분 reverse** 한다 — 단순 carve out 의 자연 후속이 아닌 결정 reversal. 핵심 결정 (책임 경계 옵션 A) 은 유지하되 보조 결정의 명시 supersession 필요.

또한 self-service unit selection 자체가 ADR-0020 §3.2 의 책임 경계 표를 **확장** — "DevHub admin 단독" → "사용자 self-service onboarding + admin 검토 (`reviewed` transition)". ADR governance 측면에서 ADR-0021 발급으로 결정의 가시성 확보.

## 3. 결정

### 3.1 책임 경계 확장 — self-service unit selection 허용

ADR-0020 §3.2 의 "user 조직 unit assignment" 책임 주체를 다음과 같이 확장:

| 운영 동작 | 책임 주체 | 도구 |
| --- | --- | --- |
| **신규 user 의 unit 초기 배치** | **사용자 self-service onboarding** | DevHub `/devhub/onboarding` (`POST /api/v1/me/onboarding`, API-83) |
| **기존 user 의 unit 변경 (self-service)** | **사용자 본인** | DevHub `/account` (`PATCH /api/v1/me`, API-85) — `review_status='pending_review'` 자동 재진입 |
| **사용자 등록 후 검토 (`pending_review → reviewed`)** | **DevHub admin** | DevHub `/admin/settings/users` (`POST /api/v1/admin/users/:user_id/review`, API-86) |
| **사용자 사전 등록** | **DevHub admin** | DevHub `/admin/settings/users` (`POST /api/v1/users`, API-33 확장) — `onboarding_completed_at=NULL`, 사용자 첫 로그인 시 onboarding 강제 진입 |
| **`users.role` 직접 수정** | **금지** (Keycloak claim 매핑 + event listener 자동 sync, ADR-0020 §3.2 결정 유지) | — |

- Onboarding payload (`POST /me/onboarding` + `PATCH /me` + `POST /users` 모두) 에 **role 필드 비포함** — concept §5.8 결정. "소속 선택 = 권한 상승" 경로 차단.

### 3.2 미완료 사용자 접근 단계 — 3 tier 상태머신

미완료 사용자의 접근 권한을 3 tier 로 운영 (ARCH-ONBOARD-02 + REQ-FR-ONBOARD-006):

| 단계 | 조건 | 접근 범위 |
| --- | --- | --- |
| `limited (skip)` | `users` row 미존재 | 공통 메뉴 + `/devhub/onboarding` 페이지 + `GET /api/v1/me` 만 |
| `pending_review` | row 존재 + `onboarding_completed_at IS NOT NULL` + `review_status='pending_review'` | 공통 메뉴 + 할당된 과제/저장소/어플리케이션 (무소속 처리) |
| `reviewed` | row 존재 + `review_status='reviewed'` | 정상 접근 (모든 도메인 API) |

전이:

| 전이 | 트리거 | Audit |
| --- | --- | --- |
| `(none) → limited` | 미등록 사용자의 첫 진입 | (none) |
| `limited → pending_review` | `POST /api/v1/me/onboarding` 성공 | `account.onboarding_completed` |
| `pending_review → reviewed` | `POST /api/v1/admin/users/:id/review` (system_admin) | `account.review_confirmed` |
| `reviewed → pending_review` | `PATCH /api/v1/me` 의 `primary_unit_id` 변경 | `account.unit_changed` |

### 3.3 Lazy auto-create 폐기 (ADR-0020 부분 supersession)

ADR-0020 §3.2 / §4.1 sub-carve B / §4.2 / §6.2 의 **lazy auto-create 결정을 supersede**:

- `authenticateActor` 는 `GetUser` miss 시 user row 를 **생성하지 않는다** — DB row miss 를 정상 상태 (token-only actor) 로 취급한다.
- `AuthenticatedActor` 의 `Email` / `DisplayName` 은 Keycloak token claim 에서 직접 추출 (PR #239 의 `keycloak_verifier.go::extractDisplayName` 재사용 — 함수 자체는 보존, 호출 시점만 변경).
- `users` row 의 첫 INSERT 는 **onboarding 제출 시점** (`POST /api/v1/me/onboarding`) 에 단일 트랜잭션으로 수행.
- 폐기되는 audit event:
  - `account.lazy_provisioned` (ADR-0020 sub-carve B 신규) — 신규 emit **중단**. 기존 emit 이력은 audit_logs 에 보존 (immutable).
  - `user.role_default_assigned` (ADR-0020 sub-carve B 신규) — 신규 emit **중단**. event listener (ADR-0020 sub-carve C, PR #241) 의 group → role 매핑이 정공법.

### 3.4 Skip-and-resume 정책

Onboarding 화면에 "나중에 하기" 액션 제공 (concept §5.9 + REQ-FR-ONBOARD-011):

- skip 시 `users` row 미생성 (token-only actor 로 한정 접근 모드 진입 — §3.2 의 `limited` 단계).
- skip 횟수/시간 제한 없음 — 매 로그인 시 onboarding 강제 진입이 사실상의 reminder 로 동작.
- skip flag 는 frontend 의 `sessionStorage` 에 저장 (탭 닫기 = reset, 매 로그인 시 reminder 재진입). localStorage 미사용 (영구 skip 회피).
- skip 자체는 audit event 미발생 (state 변경 없음).

### 3.5 Backend gating + Frontend 3분기

Backend (source of truth, REQ-FR-ONBOARD-009 + ARCH §9.3):

- `onboardingGate` middleware 가 미완료 사용자에 대해 allowlist 외 모든 endpoint 를 `403 Forbidden` + `{ code: "onboarding_required" }` 차단.
- Allowlist (backend endpoint 만 — frontend 정적 페이지는 본 정책과 무관):
  - `GET /api/v1/me` (API-32 확장 — `onboarding_required` flag 반환)
  - `POST /api/v1/me/onboarding` (API-83)
  - `GET /api/v1/organizations/search` (API-84)
  - `GET /api/v1/organization/hierarchy` (기존)
  - 정적/health endpoint

Frontend (UX layer, REQ-FR-ONBOARD-010):

- 첫 진입 (session-scoped skip flag 미설정): `/devhub/onboarding` 으로 즉시 redirect.
- skip 액션 이후 (sessionStorage `devhub.onboarding.skipped=true`): 자동 redirect 없음 + 모든 페이지 상단 dismissible banner.
- 보호 리소스 진입 시도 (backend `403 onboarding_required`): skip 여부 무관 hard redirect to `/devhub/onboarding`.

## 4. 결과 / 영향

### 4.1 ADR-0020 supersession scope

본 ADR 이 supersede 하는 ADR-0020 결정 (partial):

| ADR-0020 위치 | 결정 | 본 ADR 의 supersede 방식 |
| --- | --- | --- |
| §3.2 "신규 user 의 unit 초기 배치" row | lazy auto-create 후 `primary_unit_id=NULL` 로 잔존 | §3.1 + §3.3 — onboarding 완료 시점에 row INSERT + unit 설정 |
| §3.2 "user 조직 unit assignment" row 의 책임 주체 | DevHub admin 단독 | §3.1 — DevHub admin + 사용자 self-service onboarding + 관리자 검토 |
| §4.1 sub-carve B 항목 "authenticateActor lazy auto-create 실 구현" | lazy auto-create 도입 | §3.3 — lazy 폐기 + token-only actor 흐름 |
| §4.2 의 lazy auto-create 보안 영향 | "token 검증 성공한 user 만 lazy create" | §3.3 — lazy 자체 폐기, 보안 영향 항목 무효 |
| §6.2 carve out "authenticateActor lazy auto-create 실 구현" | 진행 항목 | §3.3 — 항목 자체 deprecated (이미 PR #239 머지 후 본 ADR 로 reversal) |

ADR-0020 의 **핵심 결정** (옵션 A — Keycloak account vs DevHub user 책임 경계, `rbac_subject_roles` 제거, service account 권한 축소) 는 **변경 없이 유지**.

### 4.2 코드 영향

본 ADR 의 실 구현은 별도 IMPL carve 에서 진행 — 본 ADR 발급 시점에는 spec only:

| 영역 | 위치 | 변경 유형 |
| --- | --- | --- |
| Backend gating middleware | `internal/httpapi/onboarding_gate.go` (신규) | 신규 — `onboardingGate` middleware |
| Onboarding 제출 handler | `internal/httpapi/me_onboarding.go` (신규) | 신규 — `POST /me/onboarding` (API-83) |
| Self-service 프로필 변경 | `internal/httpapi/me.go` (확장) | 신규 endpoint — `PATCH /me` (API-85) |
| 조직 검색 handler | `internal/httpapi/organizations_search.go` (신규) | 신규 — `GET /organizations/search` (API-84) |
| 검토 transition handler | `internal/httpapi/users_admin.go` (확장 또는 신규) | 신규 endpoint — `POST /admin/users/:id/review` (API-86) |
| Lazy auto-create | `internal/httpapi/lazy_auto_create.go` (폐기) | **폐기** — `authenticateActor` 가 DB miss 정상 처리 |
| Audit action const | `internal/domain/domain.go` (확장) | 신규 — `account.onboarding_completed` + `account.review_confirmed` + `account.unit_changed` 3종 |
| 마이그레이션 | `backend-core/migrations/0000XX_user_onboarding_state.up.sql` (신규) | 신규 — `onboarding_completed_at` + `review_status` 컬럼 + CHECK 제약 |
| Frontend | `frontend/app/onboarding/page.tsx` + `OrganizationPicker.tsx` + `(dashboard)/layout.tsx` + `account/page.tsx` 확장 | 신규/확장 |

### 4.3 보안 영향

- Backend gating (`onboardingGate`) — 미완료 사용자의 도메인 API 접근 차단 (이전 lazy auto-create 정책의 "차단 안 함" → 본 ADR 의 "allowlist 외 403"). enumeration 방어 강화.
- Role 권한 상승 차단 — onboarding payload 에 role 필드 거부 (`POST /me/onboarding` + `PATCH /me` + `POST /users`). Keycloak claim 매핑 + event listener (ADR-0020 sub-carve C) 가 유일한 role 결정 경로.
- Skip flag 의 sessionStorage 저장 — 탭 단위 격리 + 매 로그인 시 reminder 재진입. localStorage 미사용으로 영구 skip 회피.

### 4.4 운영 영향

- DevHub admin 의 `/admin/settings/users` UI 에 새 액션 — "검토 확정 (Confirm Review)" 버튼 (pending_review user 에 대해).
- 사용자의 `/account` 페이지에 새 form — 소속 (primary_unit_id) self-service 변경 (변경 시 pending_review 재진입 안내).
- HRDB ETL pre-stage (ADR-0020 §3.2 (a) 옵션) 의 의미 변경 — 사전 등록은 가능하되 `onboarding_completed_at=NULL` 유지, 사용자가 첫 로그인 시 onboarding 화면에서 확인/수정 후 제출해야 완료.

### 4.5 ADR governance

- ADR-0020 의 메타 헤더에 "partial supersession by ADR-0021" 명시.
- ADR-0020 의 §3.2 / §4.1 / §4.2 / §6.2 의 lazy auto-create 결정 위치에 inline supersession banner 추가 (메모리 `feedback_adr_supersession_pattern` 패턴).
- `docs/traceability/report.md` §4 ADR 인덱스 + §6 changelog 갱신.

## 5. 대안 / 거부된 옵션

### 5.1 옵션 A — Force completion (concept §5.9 옵션 A)

- 사용자가 onboarding 완료 전까지 모든 진입 차단 (로그아웃은 가능, skip 불가).
- **거부 이유**: UX 압박 + 사용자 포기 위험. 사용자가 onboarding 화면을 만나는 첫 순간에 다른 작업으로 이동할 가능성 있음 — DevHub 진입 자체를 포기할 위험.

### 5.2 옵션 C — Admin escalation 단독 (concept §5.9 옵션 C)

- 사용자가 자신의 소속을 모르는 경우 관리자에게 문의하는 경로만 제공.
- **거부 이유**: 중도 이탈 자체 미허용 — UX 측면에서 옵션 A 와 동일한 문제. 사용자 능동성 차단.

### 5.3 Lazy auto-create 유지 (concept §5.2 옵션 A)

- ADR-0020 의 lazy auto-create 흐름을 그대로 유지 + onboarding 화면은 별도 단계로 추가.
- **거부 이유**: "row 존재 = 등록 완료" 의미가 깨짐. lazy create 후 unit NULL 상태가 onboarding 완료까지 잔존 — UI / 통계 / audit 에서 노이즈 발생. ADR-0020 결정 동인 ("divergence 원천 차단") 과 부분 충돌.

### 5.4 신규 컬럼 `profile_status ENUM(...)` (concept §5.1 옵션 C)

- 단순 `onboarding_completed_at TIMESTAMP NULL` 대신 다단계 상태 머신 (`incomplete`, `submitted`, `verified`, `complete` 등) ENUM 컬럼.
- **거부 이유**: 현 시점에 과한 모델링. `onboarding_completed_at IS NOT NULL` + `review_status (pending_review|reviewed)` 2 컬럼으로 2 차원 표현 충분. 다단계 onboarding 도입 시점에 재고.

### 5.5 ADR-0020 §3.2 partial 수정 (옵션 A — 본 ADR 의 거부된 대안)

- ADR-0020 §3.2 표의 row 를 partial 수정 + 신규 sub-section 추가로 종결.
- **거부 이유**: ADR supersession 패턴 (메모리 `feedback_adr_supersession_pattern`) 위반 가능. PR #167 의 ADR-0001 partial 수정 → sprint -a (PR #169) 가 ADR-0019 발행으로 정정한 전례 존재. 본문 immutable 보존 원칙 + 결정 가시성 확보 측면에서 신규 ADR 발급이 정공법.

## 6. 미해결 / 후속 작업

### 6.1 IMPL carve

본 ADR 의 결정을 실 구현하는 sprint 분담 — 상세 plan 은 [`docs/planning/onboarding_impl_plan.md`](../planning/onboarding_impl_plan.md) (2026-05-21 sprint `claude/onboarding-impl-carve-plan-2026-05-21` 신규).

| Carve | RM ID | 영역 | Worker | Milestone | 진입 조건 |
| --- | --- | --- | --- | --- | --- |
| **IMPL-onboarding-backend** | RM-ONBOARD-01 | migration + onboardingGate middleware + 5 endpoint handler (API-83/84/85/86 + API-32/33 확장) + audit event const + lazy_auto_create 폐기. Feature flag default OFF. | Claude | M-v1.1 | 없음 (단독 진입 가능) |
| **IMPL-onboarding-frontend** | RM-ONBOARD-02 | `/onboarding` page + OrganizationPicker (typeahead + tree) + skip flag sessionStorage + banner + `(dashboard)/layout` 3-branch gating + `/account` unit edit | Gemini | M-v1.1 | Carve A 머지 후 |
| **IMPL-onboarding-admin** | RM-ONBOARD-03 | `/admin/settings/users` 의 "Confirm Review" 액션 + pending_review user list filter + `ConfirmReviewModal` | Gemini | M-v1.1 | Carve A 머지 후. Carve B 와 병행 가능 |
| **IMPL-onboarding-tests** | RM-ONBOARD-04 | UT-onboarding-* (handler 단위) + TC-ONBOARD-* (E2E mega lifecycle, concept §8 #12 의 6 시드 활용) + `docs/tests/test_cases_m7_onboarding.md` | Claude (UT) + Gemini (E2E) | M-v1.1 | Carve A + B + C 모두 머지 후 |

GitHub issue 등록: P2-8 ~ P2-11 (4건, `release_v1_roadmap.md` §3.3 매트릭스 참조).

### 6.2 후속 carve (concept §5.4 옵션 C/D)

- **HRDB cross-check** — 사용자가 입력한 소속과 HRDB 의 직원-부서 매핑 비교 + 불일치 시 경고. ADR-0008 deprecated (외부 Keycloak 시나리오 채택, issue #215 cancelled) 후 사실 정합 필요.
- **Keycloak group → unit 자동 매핑** — 사용자의 Keycloak group membership 으로 unit 자동 유추. ADR-0019 §5.3 sub-carve 의 group staging-prod (잔여 1건, v1.0 release gate) 완료 후 결합.
- **Review status reversal** — admin 이 `reviewed → pending_review` 강제 되돌리기 (재교육/재인증). 운영 정책 결정 후 carve.

### 6.3 부가 프로필 필드 확장 (REQ-FR-ONBOARD-012 후속)

- 사진/아바타, 닉네임, 입사일, 연락처 등 onboarding 완료 후 `/account` 에서 변경 가능한 필드는 본 ADR 의 1차 범위 밖. 별도 design carve.

## 7. 변경 이력

| 일자 | 변경 | sprint |
| --- | --- | --- |
| 2026-05-21 | 본 ADR 발급. ADR-0020 의 lazy auto-create 결정 (§3.2 / §4.1 sub-carve B / §4.2 / §6.2) partial supersession + self-service unit selection 책임 경계 확장. 3 tier 접근 상태머신 + skip-and-resume + role 권한 상승 차단 정합. 컨셉 (PR #260 + #265) + 요구사항 (PR #266) + 설계 (PR #267) 의 누적 결정 명문화. | `claude/onboarding-adr-2026-05-21` |
| 2026-05-21 | codex review hotfix (PR #270) — ADR-0020 메타 헤더 3 line (상태 / partial superseded by / supersession 안내 box) 의 supersession scope "4 위치" → "5 위치" + §6.1 explicit 정합. 본 ADR 본문은 변경 없음. | `claude/onboarding-codex-hotfix-2026-05-21` |
| 2026-05-21 | §6.1 IMPL carve 표 확장 — RM-ONBOARD-01..04 + worker / milestone / 진입 조건 column 추가 + [`onboarding_impl_plan.md`](../planning/onboarding_impl_plan.md) cross-link. GitHub issue 매핑 (P2-8..11). | `claude/onboarding-impl-carve-plan-2026-05-21` |
