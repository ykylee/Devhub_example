---
title: test_cases
type: source
tags: [domain, test_cases.md, project-devhub]
sources: [raw/projects/devhub/docs/domain/onboarding/test_cases.md]
git_commit: e91115f0
git_branch: chore/260622-wiki-drift-cleanup-2
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
last_touched: 2026-06-22T04:24:49Z
mirror_dirty: |
related: [none]
status: draft
contradictions: [none]
---

# M7 Onboarding 테스트 케이스 카탈로그

- 문서 목적: Onboarding 도메인의 테스트 범위와 우선순위 TC 카탈로그를 정의해 구현 단계의 품질 기준선 + traceability 매핑을 제공한다.
- 범위: onboardingGate middleware + POST /me/onboarding (제출) + PATCH /me (self-service) + GET /organizations/search (typeahead) + POST /admin/users/:id/review (검토 confirm) + 3-tier access state machine + token-only actor 처리.
- 대상 독자: Backend / Frontend 개발자, QA, AI 에이전트, 운영자.
- 상태: draft
- 최종 수정일: 2026-05-21
- 관련 문서: [`requirements.md`](../requirements.md) §5.7, [`domain/onboarding/concept.md`](../domain/onboarding/concept.md), [`planning/system_usecases.md`](../planning/system_usecases.md) §2.13, [`architecture.md`](../architecture.md) §9, [`backend_api_contract.md`](../backend_api_contract.md) §16, [`docs/adr/0021-onboarding-self-service-unit-selection.md`](../adr/0021-onboarding-self-service-unit-selection.md), [`domain/onboarding/impl_plan.md`](../domain/onboarding/impl_plan.md).

## 1. 기능 맵 (REQ / UC 기준)

| 기능 ID | 설명 | REQ | UC |
| --- | --- | --- | --- |
| F-ONBOARD-SUBMIT | 신규 사용자가 display_name + primary_unit_id 제출 (INSERT 또는 UPDATE 단일 트랜잭션) | REQ-FR-ONBOARD-001..003, 008 | UC-ONBOARD-01,02,08 |
| F-ONBOARD-SEARCH | 조직 검색 typeahead (q ≥ 2, limit ≤ 20) | REQ-FR-ONBOARD-004 | UC-ONBOARD-03,04 |
| F-ONBOARD-GATE | Backend gating 403 차단 + frontend 3-branch redirect / banner | REQ-FR-ONBOARD-009, 010 | UC-ONBOARD-09, 10, 11 |
| F-ONBOARD-REVIEW | 관리자 confirm transition (pending_review → reviewed) | REQ-FR-ONBOARD-005 | UC-ONBOARD-05 |
| F-ONBOARD-PATCH | Self-service primary_unit_id / display_name 변경 + review_status 자동 reset | REQ-FR-ONBOARD-006, 007 | UC-ONBOARD-06, 07 |
| F-ONBOARD-SKIP | "나중에 하기" — sessionStorage skip flag + 보호 경로 차단 유지 | REQ-FR-ONBOARD-011 | UC-ONBOARD-11 |

## 2. 테스트 계층 전략

1. **단위 테스트 (UT)** — store / handler / middleware 단위. Carve A 가 `backend-core/internal/httpapi/onboarding_test.go` (13 UT) + Carve D 가 `onboarding_carve_d_test.go` (8 UT) 로 cover. memoryOrganizationStore 가 production *store.PostgresStore 의 FK constraint mirror.
2. **통합 테스트 (IT)** — DB + handler 조합. INSERT / UPDATE branch 의 단일 트랜잭션 정합 + bi-implication CHECK constraint (migration 000033) — `internal/store/users_units_test.go` 가 source-of-truth.
3. **E2E 테스트 (TC-ONBOARD-*)** — 본 카탈로그. UI flow + Keycloak 토큰 round-trip + 3-branch gating 운영 시나리오. **본 카탈로그의 TC 는 `frontend/tests/e2e/onboarding-first-login.spec.ts` (후속 carve)** 가 source-of-truth.

## 3. 우선 테스트 케이스 (P0 / P1)

### 3.1 Happy path (lifecycle 전체)

| TC ID | 우선순위 | 계층 | 시나리오 | 기대 결과 | spec ts 위치 |
| --- | --- | --- | --- | --- | --- |
| `TC-ONBOARD-SUBMIT-01` | P0 | E2E | 신규 token-only actor 로그인 → `/devhub/onboarding` 자동 redirect → display_name 입력 + 조직 검색 (typeahead) → 선택 → 제출 | 201 + users row INSERT + onboarding_completed_at=NOW() + review_status=`pending_review` + audit `account.onboarding_completed` | `onboarding-first-login.spec.ts` step 1 |
| `TC-ONBOARD-REVIEW-01` | P0 | E2E | system_admin 이 `/admin/settings/users` 진입 → "검토 대기" panel 에 신규 사용자 노출 → `확정` 버튼 클릭 → ConfirmReviewModal → 확정 | 200 + review_status=`reviewed` + reviewed_at=NOW() + audit `account.review_confirmed` | `onboarding-first-login.spec.ts` step 2 |
| `TC-ONBOARD-SUBMITTED-ACCESS-01` | P0 | E2E | 제출 직후 (review 대기 중) 일반 페이지 (dashboard) 진입 | 정상 진입 (`onboarding_required: false`) + 일부 제한 안내 없음 | `onboarding-first-login.spec.ts` step 3 |
| `TC-ONBOARD-PATCH-UNIT-01` | P0 | E2E | reviewed 사용자가 `/account` 진입 → 조직 변경 → 저장 | 200 + review_status 자동 `pending_review` 재진입 + 사용자에게 "재검토 필요" 안내 | `onboarding-first-login.spec.ts` step 4 |

### 3.2 Skip-and-resume (§5.9)

| TC ID | 우선순위 | 계층 | 시나리오 | 기대 결과 | spec ts 위치 |
| --- | --- | --- | --- | --- | --- |
| `TC-ONBOARD-SKIP-01` | P0 | E2E | onboarding 화면에서 `나중에 하기` 클릭 | sessionStorage `devhub.onboarding.skipped=1` set + default landing redirect + 모든 페이지 상단에 dismissible banner 노출 | `onboarding-first-login.spec.ts` step 5 |
| `TC-ONBOARD-SKIP-PROTECTED-01` | P0 | E2E | skip 단계 사용자가 `/account` 진입 시도 | `/onboarding` hard redirect (REQ-FR-ONBOARD-010 분기 3) | `onboarding-first-login.spec.ts` step 6 |
| `TC-ONBOARD-SKIP-RESUME-01` | P1 | E2E | skip 후 banner 의 `지금 등록` 버튼 클릭 → onboarding 화면 → 제출 | 정상 제출 + banner 사라짐 | `onboarding-first-login.spec.ts` step 7 |
| `TC-ONBOARD-SKIP-AUDIT-01` | P1 | UT | skip 자체는 audit event 미발생 (REQ-FR-ONBOARD-011) | audit_logs query 결과 0 row | UT 단위 (E2E 검증 불필요) |

### 3.3 Admin pre-seed (API-33 확장)

| TC ID | 우선순위 | 계층 | 시나리오 | 기대 결과 | spec ts 위치 |
| --- | --- | --- | --- | --- | --- |
| `TC-ONBOARD-PRESEED-01` | P0 | E2E | admin 이 `POST /api/v1/users` 로 사전등록 (primary_unit_id 입력) → 사용자가 첫 로그인 | onboarding 화면에 display_name + primary_unit_id 채워진 상태 노출 (`fromAdmin=true` banner) | `onboarding-first-login.spec.ts` step 8 |
| `TC-ONBOARD-PRESEED-SUBMIT-01` | P0 | E2E | 사전등록된 사용자가 onboarding 제출 | row UPDATE (INSERT 아님) + onboarding_completed_at NOW + review_status=`pending_review` | `onboarding_carve_d_test.go::TestSubmitOnboarding_PreSeededUpdate` (UT cover) |

### 3.4 Negative path (제출)

| TC ID | 우선순위 | 계층 | 시나리오 | 기대 결과 | spec ts 위치 |
| --- | --- | --- | --- | --- | --- |
| `TC-ONBOARD-SUBMIT-NEG-01` | P0 | UT | display_name 누락 | 422 `invalid_payload` | `onboarding_test.go::TestSubmitOnboarding_InvalidPayload422` |
| `TC-ONBOARD-SUBMIT-NEG-02` | P0 | UT | primary_unit_id 가 organization_units 에 없음 | 404 `unit_not_found` | `onboarding_test.go::TestSubmitOnboarding_UnitNotFound404` |
| `TC-ONBOARD-SUBMIT-NEG-03` | P0 | UT | 이미 onboarding_completed_at IS NOT NULL 인 사용자 중복 제출 | 409 `onboarding_already_completed` | `onboarding_test.go::TestSubmitOnboarding_AlreadyCompleted409` |
| `TC-ONBOARD-SUBMIT-NEG-04` | P1 | UT | 요청 payload 에 role 포함 시 무시 | 201 + DB role 은 fallback (`developer`) | `onboarding_carve_d_test.go::TestSubmitOnboarding_RoleIgnored` |

### 3.5 Negative path (검색 / PATCH / review)

| TC ID | 우선순위 | 계층 | 시나리오 | 기대 결과 | spec ts 위치 |
| --- | --- | --- | --- | --- | --- |
| `TC-ONBOARD-SEARCH-NEG-01` | P0 | UT | q 1자 이하 또는 미지정 | 422 `invalid_query_params` | `onboarding_test.go::TestOrganizationsSearch_QTooShort422` |
| `TC-ONBOARD-SEARCH-NEG-02` | P1 | UT | limit > 20 | 422 `invalid_query_params` | `onboarding_carve_d_test.go::TestOrganizationsSearch_LimitOverMax422` |
| `TC-ONBOARD-PATCH-NEG-01` | P0 | UT | primary_unit_id 가 organization_units 에 없음 | 404 `unit_not_found` | `onboarding_carve_d_test.go::TestPatchMe_UnitNotFound404` |
| `TC-ONBOARD-PATCH-NEG-02` | P0 | UT | display_name + primary_unit_id 둘 다 미포함 | 422 `invalid_payload` | `onboarding_carve_d_test.go::TestPatchMe_InvalidPayload422_NoFields` |
| `TC-ONBOARD-REVIEW-NEG-01` | P0 | UT | 존재하지 않는 user id 로 review confirm | 404 `user_not_found` | `onboarding_carve_d_test.go::TestConfirmUserReview_UserNotFound404` |
| `TC-ONBOARD-REVIEW-NEG-02` | P0 | UT | 이미 `reviewed` 사용자 중복 confirm | 409 `review_already_confirmed` | `onboarding_carve_d_test.go::TestConfirmUserReview_AlreadyConfirmed409` |
| `TC-ONBOARD-REVIEW-NEG-03` | P0 | UT | onboarding_completed_at IS NULL 사용자에 대해 confirm | 422 `onboarding_not_completed` | `onboarding_carve_d_test.go::TestConfirmUserReview_OnboardingNotCompleted422` |

### 3.6 Gating (middleware + frontend)

| TC ID | 우선순위 | 계층 | 시나리오 | 기대 결과 | spec ts 위치 |
| --- | --- | --- | --- | --- | --- |
| `TC-ONBOARD-GATE-01` | P0 | UT | flag false (default) + 미완료 사용자 호출 | gate no-op (legacy lazy 동작) | `onboarding_test.go::TestOnboardingGate_FeatureFlagOff_NoOp` |
| `TC-ONBOARD-GATE-02` | P0 | UT | flag true + 미완료 사용자 + allowlist endpoint (`/me`, `/me/onboarding`, `/organizations/search`, `/organization/hierarchy`) | 정상 처리 | `onboarding_test.go::TestOnboardingGate_FeatureFlagOn_AllowlistAccess` |
| `TC-ONBOARD-GATE-03` | P0 | UT | flag true + 미완료 사용자 + allowlist 외 endpoint (예: `/dev-requests`) | 403 `onboarding_required` | `onboarding_test.go::TestOnboardingGate_FeatureFlagOn_NonAllowlistBlocks` |
| `TC-ONBOARD-FLAG-OFF-01` | P0 | UT | flag false (default) + 신규 endpoint 4건 (POST /me/onboarding 등) 호출 | 404 `onboarding_feature_disabled` (main 안정성) | `onboarding_test.go::TestOnboardingEndpoints_FlagOff404` |

### 3.7 Frontend 3-branch gating (E2E)

| TC ID | 우선순위 | 계층 | 시나리오 | 기대 결과 | spec ts 위치 |
| --- | --- | --- | --- | --- | --- |
| `TC-ONBOARD-FE-GATE-01` | P0 | E2E | 첫 진입 (sessionStorage skip flag 없음) + 미완료 사용자 | `/devhub/onboarding` 으로 즉시 redirect | `onboarding-first-login.spec.ts` step 9 |
| `TC-ONBOARD-FE-GATE-02` | P0 | E2E | skip 액션 후 일반 dashboard 페이지 진입 | 정상 진입 + 상단 dismissible banner | `onboarding-first-login.spec.ts` step 10 |
| `TC-ONBOARD-FE-GATE-03` | P0 | E2E | skip 단계 사용자가 backend 의 보호 endpoint 호출 시도 (예: `/api/v1/dev-requests`) | 403 응답 + frontend 가 `/devhub/onboarding` 으로 hard redirect | `onboarding-first-login.spec.ts` step 11 |

## 4. 우선순위 등급 정의

- **P0** — 본 도메인의 핵심 기능 / 회귀 시 main 안정성 위반.
- **P1** — 핵심 보조 기능 / negative case / monitoring 대상 (audit).
- **P2** — UX 보강 + 후속 carve 후보.

## 5. cover 매트릭스 (TC ↔ REQ ↔ UC ↔ ARCH ↔ API ↔ IMPL 매핑)

| TC ID | REQ | UC | ARCH | API | IMPL | UT |
| --- | --- | --- | --- | --- | --- | --- |
| `TC-ONBOARD-SUBMIT-01` | REQ-FR-ONBOARD-001..003, 008 | UC-ONBOARD-01,02,08 | ARCH-ONBOARD-02,06 | API-83 | IMPL-onboarding-backend, IMPL-onboarding-frontend | TestSubmitOnboarding_HappyPath |
| `TC-ONBOARD-REVIEW-01` | REQ-FR-ONBOARD-005 | UC-ONBOARD-05 | ARCH-ONBOARD-02,06 | API-86 | IMPL-onboarding-backend, IMPL-onboarding-admin | TestConfirmUserReview_HappyPath |
| `TC-ONBOARD-PATCH-UNIT-01` | REQ-FR-ONBOARD-006, 007 | UC-ONBOARD-06, 07 | ARCH-ONBOARD-02,06 | API-85 | IMPL-onboarding-backend, IMPL-onboarding-frontend | TestPatchMe_UnitChangeResetsReview |
| `TC-ONBOARD-SKIP-01` | REQ-FR-ONBOARD-010, 011 | UC-ONBOARD-10, 11 | ARCH-ONBOARD-03 | (frontend) | IMPL-onboarding-frontend | (e2e only) |
| `TC-ONBOARD-PRESEED-01` | REQ-FR-ONBOARD-008 | UC-ONBOARD-08 | ARCH-ONBOARD-02,05 | API-33 (확장) | IMPL-onboarding-backend, IMPL-onboarding-frontend | TestSubmitOnboarding_PreSeededUpdate |
| `TC-ONBOARD-GATE-01..03` | REQ-FR-ONBOARD-009 | UC-ONBOARD-09 | ARCH-ONBOARD-03 | (gating) | IMPL-onboarding-backend | TestOnboardingGate_* |

## 6. 변경 이력

- **2026-05-21** (sprint `claude/issue-275-onboarding-tests`): 본 카탈로그 신규. Carve A 의 13 UT + Carve D 의 8 UT 추가로 backend UT 21건 cover (3.1 Happy path, 3.3 admin pre-seed, 3.4~3.6 negative + gate). frontend e2e (3.1 Happy + 3.2 Skip + 3.7 FE gate) 는 후속 carve 의 `frontend/tests/e2e/onboarding-first-login.spec.ts` 가 source-of-truth.
