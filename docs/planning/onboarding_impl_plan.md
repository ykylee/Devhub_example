# Keycloak ↔ DevHub Onboarding — IMPL Carve Plan

- 문서 목적: Onboarding 도메인의 IMPL (구현) phase 를 4 carve 로 분할하고, 각 carve 의 scope / dependency / 진입 순서 / 검증 기준 / GitHub issue 매핑을 정의한다.
- 범위: 4 carve (backend / frontend / admin UI / tests) — 각 carve 가 1 sprint = 1 PR.
- 대상 독자: Claude / Gemini / Codex 워커, PR 작성자, 후속 carve 담당자.
- 상태: draft
- 작성일: 2026-05-21
- 결정 근거 sprint: `claude/onboarding-impl-carve-plan-2026-05-21`
- 관련 문서:
  - [컨셉 `keycloak_user_onboarding_concept.md`](./keycloak_user_onboarding_concept.md)
  - [요구사항 `docs/requirements.md §5.7`](../requirements.md) (REQ-FR-ONBOARD-001..012 / REQ-NFR-ONBOARD-001..008)
  - [Usecase `system_usecases.md §2.13`](./system_usecases.md) (UC-ONBOARD-01..11)
  - [Architecture `docs/architecture.md §9`](../architecture.md) (ARCH-ONBOARD-01..06)
  - [API contract `docs/backend_api_contract.md §16`](../backend_api_contract.md) (API-83..86 + API-32/33 확장)
  - [ADR-0021 Onboarding self-service](../adr/0021-onboarding-self-service-unit-selection.md)
  - [worker_division.md](../governance/worker_division.md)
  - [release_v1_roadmap.md](./release_v1_roadmap.md)

## 1. Carve 분할 원칙

| 원칙 | 적용 |
| --- | --- |
| **워커 책임 경계 최대 정합** | backend (Claude) ↔ frontend / admin UI (Gemini) ↔ tests (Claude UT + Gemini E2E) 자연 분리 |
| **PR 크기 ≤ 800 LoC** | 각 carve 의 backend 또는 frontend 변경분 한정 |
| **dependency-first** | backend (carve A) 가 frontend (carve B/C) 의 contract 제공 — A 머지 후 B/C 진입 |
| **테스트 동반** | 각 carve 가 자체 UT + 회귀 test 포함. E2E (carve D) 는 carve A+B+C 머지 후 종합 |
| **roll-back-safe** | onboardingGate middleware 가 feature flag (env `DEVHUB_ONBOARDING_GATE_ENABLED=false` default) 로 disable 가능 — carve A 의 단독 머지 후에도 main 안정성 유지 |

## 2. Carve 인벤토리

### 2.1 Carve A — IMPL-onboarding-backend (Claude)

**RM ID**: `RM-ONBOARD-01`

**Scope**: backend handler + middleware + migration + audit event const + `lazy_auto_create.go` 폐기.

**파일 변경**:

| 파일 | 변경 유형 | 비고 |
| --- | --- | --- |
| `backend-core/migrations/0000XX_user_onboarding_state.up.sql` | 신규 | `users.onboarding_completed_at timestamptz NULL` + `users.review_status text NULL` + bi-implication CHECK 제약 (ARCH §9.5) |
| `backend-core/migrations/0000XX_user_onboarding_state.down.sql` | 신규 | 컬럼 + CHECK rollback |
| `backend-core/internal/httpapi/onboarding_gate.go` | 신규 | `onboardingGate` middleware — allowlist 외 미완료 user 403 (ARCH §9.3) |
| `backend-core/internal/httpapi/me_onboarding.go` | 신규 | `POST /api/v1/me/onboarding` handler (API-83, §16.3) |
| `backend-core/internal/httpapi/me.go` | 확장 | `GET /api/v1/me` 응답 shape 확장 (API-32 §16.2) + `PATCH /api/v1/me` handler (API-85 §16.5) |
| `backend-core/internal/httpapi/organizations_search.go` | 신규 | `GET /api/v1/organizations/search` handler (API-84 §16.4) |
| `backend-core/internal/httpapi/users_admin_review.go` | 신규 | `POST /api/v1/admin/users/:user_id/review` handler (API-86 §16.7) |
| `backend-core/internal/httpapi/organization.go` | 확장 | `POST /api/v1/users` admin 사전등록 의미 명시 (API-33 §16.6) + role/email/onboarding_completed_at 정합 |
| `backend-core/internal/httpapi/lazy_auto_create.go` | **삭제** | ADR-0021 §3.3 lazy auto-create 폐기 — `authenticateActor` 가 DB miss 정상 처리 |
| `backend-core/internal/httpapi/auth.go` | 확장 | `authenticateActor` 의 lazy 분기 제거 + `AuthenticatedActor` 의 token-only 동작 |
| `backend-core/internal/httpapi/router.go` | 확장 | 5 신규 endpoint 라우팅 + `onboardingGate` middleware wire |
| `backend-core/internal/httpapi/permissions.go` | 확장 | route permission 표에 5 endpoint 추가 (ARCH §9.4) |
| `backend-core/internal/domain/domain.go` | 확장 | audit event const 3종 신규 (`account.onboarding_completed` / `account.review_confirmed` / `account.unit_changed`) — ARCH §9.6 |
| `backend-core/internal/store/postgres_users.go` | 확장 | `UpsertUser` / `UpdateUserUnit` / `SetUserReviewed` method 신규 또는 확장 |
| `backend-core/cmd/devhub-core/main.go` | 확장 | env `DEVHUB_ONBOARDING_GATE_ENABLED` 읽기 + `onboardingGate` wire (default: false — feature flag) |

**예상 LoC**: backend +800 / -150 (lazy_auto_create.go 삭제 -100 포함)

**Acceptance criteria**:

| Criterion | Verification |
| --- | --- |
| Migration up/down 모두 성공 | `make migrate-up` + `make migrate-down` |
| `GET /api/v1/me` 응답이 `onboarding_required` flag 반환 | UT-onboarding-me-01 |
| `POST /me/onboarding` 성공 시 single tx (INSERT + completed_at + pending_review + audit emit) | UT-onboarding-submit-01..04 |
| `PATCH /me` 의 primary_unit_id 변경이 review_status 를 pending_review 로 재진입 | UT-onboarding-patch-01..02 |
| `GET /organizations/search` 가 q ≥ 2, limit ≤ 20 정합 | UT-onboarding-search-01..03 |
| `POST /admin/users/:id/review` 가 pending_review → reviewed transition | UT-onboarding-review-01..03 |
| `authenticateActor` 가 DB miss 시 token-only actor 정상 반환 (lazy 폐기 정합) | UT-onboarding-actor-01..02 |
| `onboardingGate` 가 미완료 user 의 allowlist 외 endpoint 호출 시 403 | UT-onboarding-gate-01..04 |
| Feature flag disable 시 onboardingGate 가 no-op (main 안정성) | UT-onboarding-gate-disabled-01 |

**Dependencies**: 없음 (단독 진입 가능).

**Sprint name 예시**: `claude/work_260522-x-onboarding-backend` (또는 issue # 기반).

### 2.2 Carve B — IMPL-onboarding-frontend (Gemini)

**RM ID**: `RM-ONBOARD-02`

**Scope**: onboarding page + OrganizationPicker + 3-branch gating + dismissible banner + `/account` self-service unit edit.

**파일 변경**:

| 파일 | 변경 유형 | 비고 |
| --- | --- | --- |
| `frontend/app/onboarding/page.tsx` | 신규 | onboarding form (display_name + primary_unit_id) + "나중에 하기" 액션 |
| `frontend/components/onboarding/OnboardingForm.tsx` | 신규 | form + submit + skip 버튼 |
| `frontend/components/organization/OrganizationPicker.tsx` | 신규 | typeahead (≥2 chars, ≤20 results) + tree (hierarchy endpoint 재사용) 하이브리드 |
| `frontend/lib/services/onboarding.service.ts` | 신규 | `submitOnboarding` / `searchOrganizations` API client |
| `frontend/lib/services/identity.service.ts` | 확장 | `getMe()` response shape 확장 (`onboarding_required` / `onboarding_completed_at` / `review_status`) |
| `frontend/app/(dashboard)/layout.tsx` | 확장 | 3-branch gating — 첫 진입 redirect / skip 후 banner / 보호 리소스 hard redirect (ARCH §9.3) |
| `frontend/components/onboarding/OnboardingBanner.tsx` | 신규 | dismissible banner — "프로필을 완료해 주세요" + onboarding 페이지 링크 |
| `frontend/lib/storage/onboardingSkip.ts` | 신규 | sessionStorage `devhub.onboarding.skipped` set/get/clear helper |
| `frontend/app/account/page.tsx` | 확장 | self-service primary_unit_id 변경 form + pending_review 재진입 안내 |
| `frontend/lib/services/account.service.ts` | 확장 | `patchMe` (display_name + primary_unit_id 변경) |

**예상 LoC**: frontend +900 / -50

**Acceptance criteria**:

| Criterion | Verification |
| --- | --- |
| 미완료 user 첫 진입 시 `/devhub/onboarding` redirect | TC-ONBOARD-FIRST-ENTRY-01 |
| "나중에 하기" 클릭 시 sessionStorage flag set + dashboard 진입 + banner 노출 | TC-ONBOARD-SKIP-01 |
| 보호 리소스 진입 시도 시 backend 403 → hard redirect | TC-ONBOARD-BACKEND-403-01 |
| OrganizationPicker 의 typeahead 가 2 chars 미만 시 disable | TC-ONBOARD-ORG-SEARCH-01 |
| OrganizationPicker 의 tree 가 hierarchy endpoint 정합 | TC-ONBOARD-ORG-TREE-01 |
| onboarding 제출 후 dashboard redirect + banner 미노출 | TC-ONBOARD-SUBMIT-01 |
| `/account` 에서 primary_unit_id 변경 시 pending_review 재진입 안내 | TC-ONBOARD-UNIT-CHANGE-01 |
| 키보드만으로 검색/선택/제출 가능 (REQ-NFR-ONBOARD-002) | TC-ONBOARD-A11Y-01 |
| sessionStorage 가 탭 닫기 후 reset | TC-ONBOARD-SKIP-RESET-01 |

**Dependencies**: Carve A 머지 후 진입 (`onboardingGate` + 5 endpoint contract 필요).

**Sprint name 예시**: `gemini/work_260523-y-onboarding-frontend`.

### 2.3 Carve C — IMPL-onboarding-admin (Gemini)

**RM ID**: `RM-ONBOARD-03`

**Scope**: `/admin/settings/users` 페이지의 "Confirm Review" 액션 + pending_review user list filter.

**파일 변경**:

| 파일 | 변경 유형 | 비고 |
| --- | --- | --- |
| `frontend/app/admin/settings/users/page.tsx` | 확장 | review_status column + "Confirm Review" 액션 버튼 (system_admin 만) + pending_review filter |
| `frontend/components/admin/users/ConfirmReviewModal.tsx` | 신규 | confirm dialog — 사용자 정보 + primary_unit_id 표시 + 확정 버튼 |
| `frontend/lib/services/users.service.ts` | 확장 | `confirmUserReview(user_id)` API client (API-86) |
| `frontend/lib/types/user.types.ts` | 확장 | `User` interface 에 `review_status` / `onboarding_completed_at` 추가 |

**예상 LoC**: frontend +300 / -10

**Acceptance criteria**:

| Criterion | Verification |
| --- | --- |
| admin 이 pending_review user list 필터링 가능 | TC-ONBOARD-PENDING-REVIEW-LIST-01 |
| "Confirm Review" 버튼 클릭 → modal 표시 → 확정 시 reviewed transition + list refresh | TC-ONBOARD-REVIEWED-01 |
| 이미 reviewed 사용자에 대해 버튼 disable 또는 미노출 | TC-ONBOARD-REVIEWED-DISABLED-01 |
| 비-system_admin 은 버튼 미노출 | TC-ONBOARD-REVIEWED-RBAC-01 |
| admin 사전등록 endpoint (`POST /users`) 호출 시 onboarding_completed_at NULL 정합 | TC-ONBOARD-ADMIN-PRE-SEED-01 |

**Dependencies**: Carve A 머지 후 진입 (API-86 + admin endpoint 필요). Carve B 와 병행 가능 (서로 독립 paths).

**Sprint name 예시**: `gemini/work_260523-z-onboarding-admin`.

### 2.4 Carve D — IMPL-onboarding-tests (Claude UT + Gemini E2E)

**RM ID**: `RM-ONBOARD-04`

**Scope**: backend UT 보강 (Carve A 의 UT 외 추가 회귀 + edge case) + frontend E2E (TC-ONBOARD-* 11~13건 mega lifecycle spec).

**파일 변경**:

| 파일 | 변경 유형 | 비고 |
| --- | --- | --- |
| `backend-core/internal/httpapi/onboarding_test.go` | 신규 | mega test (single login → submit → review → unit change → re-review) |
| `backend-core/internal/httpapi/onboarding_gate_test.go` | 신규 | allowlist 정합 + 403 응답 shape |
| `backend-core/internal/httpapi/auth_test.go` | 확장 | token-only actor (DB miss = 정상) 회귀 |
| `frontend/tests/e2e/onboarding.spec.ts` | 신규 | TC-ONBOARD-FIRST-ENTRY/SKIP/SUBMIT/PENDING-REVIEW/REVIEWED/UNIT-CHANGE/ADMIN-PRE-SEED/ORG-SEARCH/A11Y/BACKEND-403/FRONTEND-BANNER 11건 |
| `frontend/tests/e2e/seeds/onboarding.ts` | 신규 | 6 test seed (REQ-NFR-ONBOARD-008 정합 — test_self_new_user 등) + org_fixture_bulk |
| `docs/tests/test_cases_m7_onboarding.md` | 신규 | TC 카탈로그 (DREQ §M5 / INT §M6 패턴) |

**예상 LoC**: backend test +400 / frontend test +700 + docs +200

**Acceptance criteria**:

| Criterion | Verification |
| --- | --- |
| Backend UT coverage 90%+ (handler + middleware) | `go test ./internal/httpapi/... -coverprofile` |
| E2E mega lifecycle 1 spec 5+ step PASS | `npm run test:e2e -- onboarding.spec.ts` |
| Seed 6건 모두 정상 hydrate + skip-and-resume flow 회귀 | seed 스크립트 + spec |
| TC-ONBOARD-A11Y-01 (keyboard navigation 전체 path) | Playwright axe-core 또는 manual review |
| TC 카탈로그 doc 가 traceability §3 Onboarding row 의 TC cell 채움 | traceability/report.md row update |

**Dependencies**: Carve A + B + C 모두 머지 후 진입 (E2E 가 풀스택 의존).

**Sprint name 예시**: `claude/work_260524-aa-onboarding-tests-backend` + `gemini/work_260524-bb-onboarding-tests-e2e`.

## 3. 진입 순서 + 의존 그래프

```
                       ┌───────────────────────────┐
                       │  Carve A — Backend         │
                       │  (Claude, RM-ONBOARD-01)   │
                       │  feature flag default OFF  │
                       └─────────┬─────────────────┘
                                 │ merge → contract 노출
                ┌────────────────┼────────────────┐
                ▼                                 ▼
   ┌────────────────────────────┐  ┌────────────────────────────┐
   │  Carve B — Frontend         │  │  Carve C — Admin UI         │
   │  (Gemini, RM-ONBOARD-02)    │  │  (Gemini, RM-ONBOARD-03)    │
   │  병행 가능                  │  │  병행 가능                  │
   └─────────────┬──────────────┘  └─────────────┬───────────────┘
                 │                                 │
                 └────────────┬───────────────────┘
                              │ B + C 모두 머지 후
                              ▼
                ┌──────────────────────────────────┐
                │  Carve D — Tests                  │
                │  (Claude UT + Gemini E2E,         │
                │   RM-ONBOARD-04)                  │
                │  E2E mega lifecycle               │
                └─────────────┬────────────────────┘
                              │ E2E PASS + traceability fill
                              ▼
                ┌──────────────────────────────────┐
                │  Feature flag default ON          │
                │  (별도 hotfix PR — Carve A 의     │
                │  main.go 의 default flip)         │
                └──────────────────────────────────┘
```

## 4. Milestone 매핑

| Carve | Milestone | Worker | Issue Label | 비고 |
| --- | --- | --- | --- | --- |
| **Carve A** | M-v1.1 | Claude | `priority/p2 worker/claude domain/auth type/feature` | 새 도메인 — v1.0 scope 외 |
| **Carve B** | M-v1.1 | Gemini | `priority/p2 worker/gemini domain/auth domain/ui-polish type/feature` | Carve A 머지 후 진입 |
| **Carve C** | M-v1.1 | Gemini | `priority/p2 worker/gemini domain/auth domain/ui-polish type/feature` | Carve A 머지 후 진입 |
| **Carve D** | M-v1.1 | Claude (UT) + Gemini (E2E) | `priority/p2 worker/claude worker/gemini domain/auth type/test` | Carve A + B + C 머지 후 진입 |

**Feature flag default flip** (별도 hotfix PR): M-v1.1 의 e2e PASS 후. **flip 자체는 별도 issue 로 등록 안 함** — Carve D 의 acceptance criteria 통과 후 메인테이너가 single-line PR 으로 enable.

## 5. ADR-0019 §5.3 잔여 carve 와의 관계

ADR-0021 §6.2 의 후속 carve (Onboarding 1차 범위 외) 와 ADR-0019 §5.3 잔여 carve 가 일부 overlap:

| ADR-0019 §5.3 잔여 carve | ADR-0021 §6.2 후속 carve | 관계 |
| --- | --- | --- |
| (8) Keycloak `groups` → DevHub RBAC role 매핑 (staging-prod 실 적용) | Keycloak group → unit 자동 매핑 | 본 IMPL carve 4 가 머지된 후 별도 carve 진입. group → role 적용이 먼저 (P1-3 issue #214). |
| (5) MFA / 2FA | (해당 없음 — REQ §5.7.3 Out of Scope) | 본 IMPL scope 외 |
| (7) off-boarding 즉시성 Phase 2 | (해당 없음) | 본 IMPL scope 외 |

본 IMPL carve 4건은 ADR-0019 §5.3 (8) group staging-prod 적용과 **독립** — 두 carve 가 병행 가능.

## 6. Risk + 회귀 가드

| Risk | 완화 |
| --- | --- |
| Carve A 의 lazy_auto_create 폐기가 production 영향 | **Feature flag default OFF** + 별도 hotfix PR 로 flip. Carve A 단독 머지 후 main 안정성 검증 1주 |
| Carve B + C 가 병행 진입 시 conflict | 서로 독립 paths (`frontend/app/onboarding/` vs `frontend/app/admin/settings/users/`) — conflict 위험 낮음 |
| ADR-0019 §5.3 잔여 carve (event listener / group staging) 와 RBAC 충돌 | 본 carve 4 는 신규 RBAC resource 도입 없음 (ARCH §9.4) — 충돌 없음 |
| Migration 충돌 (다른 sprint 의 migration 과 sequence 번호) | Carve A 진입 직전 main HEAD 의 마지막 migration 번호 + 1 사용 |
| sessionStorage skip flag 가 XSS 영향 | session-scoped + onboarding flag 만 — token 등 secret 미포함 |

## 7. Acceptance — IMPL phase 완료 정의 (DoD)

| # | 항목 | Verification |
| --- | --- | --- |
| 1 | Carve A + B + C + D 4 sprint 모두 머지 | git log |
| 2 | Feature flag default ON 됨 (별도 hotfix PR) | env 또는 main.go default 값 |
| 3 | UT-onboarding-* + TC-ONBOARD-* 의 traceability cell 모두 채워짐 | `docs/traceability/report.md §3` |
| 4 | concept §9 next-phase #4 (IMPL carve) + #5 (매트릭스 cell) 모두 done 마크 | concept doc §9 |
| 5 | ADR-0021 §6.1 IMPL carve 4건 모두 resolved 마킹 | ADR-0021 doc + §7 변경 이력 |
| 6 | 운영 환경 (staging) 1주 monitoring — 403 spike / sessionStorage flag race 등 회귀 없음 | 운영 metric / log |

## 8. 후속 carve (Onboarding 도메인 1차 범위 외)

ADR-0021 §6.2 의 후속 carve — 본 IMPL phase 완료 후 별도 sprint:

- **HRDB cross-check** — 사용자 입력 소속 vs HRDB 매핑 비교 (ADR-0008 deprecated 후 사실 정합 필요)
- **Keycloak group → unit 자동 매핑** — ADR-0019 §5.3 (8) group staging-prod 완료 후 결합
- **Review status reversal** — admin 의 `reviewed → pending_review` 강제 되돌리기 (재교육/재인증)
- **부가 프로필 필드** (REQ-FR-ONBOARD-012) — 사진/아바타, 닉네임, 연락처 등

## 9. 변경 이력

| 일자 | 변경 | sprint |
| --- | --- | --- |
| 2026-05-21 | 1차 draft — 4 carve (Backend / Frontend / Admin / Tests) 정의 + RM-ONBOARD-01..04 발급 + 진입 순서 + milestone 매핑 + feature flag 안정성 보장 + ADR-0019 §5.3 잔여 carve 와의 관계 정리. | `claude/onboarding-impl-carve-plan-2026-05-21` |
