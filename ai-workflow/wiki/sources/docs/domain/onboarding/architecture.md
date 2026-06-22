---
title: architecture
type: source
tags: [domain, architecture.md, project-devhub]
sources: [raw/projects/devhub/docs/domain/onboarding/architecture.md]
git_commit: baf1cf24
git_branch: main
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
last_touched: 2026-06-22T03:37:45Z
mirror_dirty: |
related: [none]
status: draft
contradictions: [none]
---

# onboarding 도메인 아키텍처

- 문서 목적: onboarding 도메인의 컴포넌트·상태머신·gating 정책·RBAC 매핑·데이터 모델·audit catalog 를 정의한다.
- 범위: ARCH-ONBOARD-01..06. cross-cutting 3대 레이어 / OIDC 일반 정책은 master `docs/architecture.md` §1–§4, §6 참조.
- 대상 독자: backend / 프론트엔드 / DevOps, AI agent, 아키텍처 검토자.
- 상태: accepted
- 최종 수정일: 2026-05-29 (Phase 3 split, master `docs/architecture.md` §9 본문 이관)
- 관련 문서: [도메인 README](./README.md), [컨셉](./concept.md), [requirements.md](./requirements.md), [api.md](./api.md), [master architecture](../../architecture.md), [ADR-0019](../../adr/0019-keycloak-only-idp.md), [ADR-0020](../../adr/0020-account-user-management-boundary.md), [ADR-0021](../../adr/0021-onboarding-self-service-unit-selection.md)

## 개요

Keycloak 인증 통과 + DevHub 프로필 미완료 사용자의 self-service 초기 등록 흐름을 처리하는 도메인. 컨셉 문서: [`./concept.md`](./concept.md). 요구사항: [`./requirements.md`](./requirements.md). Usecase: [`UC-ONBOARD-01..11`](../../planning/system_usecases.md).

## 1. 컴포넌트 (ARCH-ONBOARD-01)

```
┌──────────────────┐         ┌────────────────────────────────────────────┐
│  User (Browser)  │ ──────▶ │  Frontend (Next.js)                        │
└──────────────────┘         │  ├── /devhub/onboarding (form + skip)       │
                             │  ├── (dashboard)/layout (3-branch gating)  │
                             │  ├── /account (self-service unit change)   │
                             │  └── OrganizationPicker (typeahead + tree) │
                             └────────────┬───────────────────────────────┘
                                          │ Authorization: Bearer <kc-token>
                                          ▼
┌────────────────────────────────────────────────────────────────────────┐
│                    Go Core: Onboarding handlers                        │
│                                                                         │
│  authenticateActor                                                      │
│  ├── token verify (Keycloak JWKS, ARCH-12)                              │
│  ├── GetUser(idp_subject) — DB row miss = token-only actor (lazy        │
│  │   auto-create 폐기, REQ-FR-ONBOARD-009 정합)                          │
│  └── attach actor to context                                            │
│                                                                         │
│  onboardingGate middleware (allowlist 외 enforce)                       │
│  └── DB row 없음 OR onboarding_completed_at IS NULL → 403 +             │
│      { code: onboarding_required }                                      │
│                                                                         │
│  Handlers:                                                              │
│  ├── GET    /api/v1/me                — onboarding_required flag        │
│  ├── POST   /api/v1/me/onboarding     — 제출 (row INSERT + 완료 + audit) │
│  ├── PATCH  /api/v1/me                — self-service unit change         │
│  ├── GET    /api/v1/organizations/    — typeahead search (≤20 results)  │
│  │          search?q=...&limit=20                                       │
│  ├── POST   /api/v1/users             — admin 사전 등록 (API-33 확장)     │
│  └── POST   /api/v1/admin/users/      — review_status pending→reviewed   │
│             :user_id/review             전이 (system_admin)              │
└────────────┬───────────────────────────────────────────────────────────┘
             │
             ▼
┌────────────────────────────────────────────────────────────────────────┐
│  Postgres                                                               │
│  ├── users (onboarding_completed_at, review_status 신규 컬럼 — §5)      │
│  ├── organization_units (search/tree 소스, 기존 ARCH-15..17 재사용)     │
│  └── audit_logs (account.onboarding_completed, account.unit_changed,    │
│                  account.review_confirmed — §6)                         │
└────────────────────────────────────────────────────────────────────────┘
```

## 2. 상태 머신 (ARCH-ONBOARD-02)

미완료 사용자 접근 단계는 **3 tier** (concept §5.9 + REQ-FR-ONBOARD-006):

```
                       (첫 로그인, DB row 없음)
                                 │
                                 ▼
                   ┌─────────────────────────┐
                   │      limited (skip)     │  ← user row 미존재
                   │  공통 메뉴 + /onboarding │      onboarding_required=true
                   │  + GET /me              │      session skip flag → banner
                   └────────┬────────┬──────┘
                            │        │
                  ("나중에  │        │ (form 제출 — POST /me/onboarding)
                   하기")   │        ▼
                            │  ┌─────────────────────────┐
                            │  │   pending_review        │  ← row INSERT + completed_at
                            │  │ 무소속 처리 + 할당 리소스 │      review_status=pending_review
                            │  │  + 공통 메뉴 접근        │
                            │  └────────────┬────────────┘
                            │               │
                            │   (admin POST /admin/users/:id/review)
                            │               ▼
                            │  ┌─────────────────────────┐
                            │  │       reviewed          │  ← review_status=reviewed
                            │  │   정상 접근 (모든 API)   │
                            │  └────────────┬────────────┘
                            │               │
                            │   (self-service PATCH /me — primary_unit_id 변경)
                            │               │
                            │               ▼
                            │      review_status → pending_review (재진입)
                            ▼
                       (재로그인 또는 banner 해제)
                       매 로그인 시 limited 상태 재진입
```

전이 규칙:

| 전이 | 트리거 | 결과 컬럼 | Audit |
| --- | --- | --- | --- |
| `(none) → limited` | 미등록 사용자의 첫 진입 | (row 미생성) | (none) |
| `limited → pending_review` | `POST /api/v1/me/onboarding` 성공 | row INSERT + `onboarding_completed_at=NOW()` + `review_status='pending_review'` | `account.onboarding_completed` |
| `pending_review → reviewed` | `POST /api/v1/admin/users/:id/review` (system_admin) | `review_status='reviewed'` | `account.review_confirmed` |
| `reviewed → pending_review` | `PATCH /api/v1/me` 의 `primary_unit_id` 변경 | `review_status='pending_review'` | `account.unit_changed` |

## 3. Gating 정책 (ARCH-ONBOARD-03)

- **Backend** — source of truth.
  - `onboardingGate` middleware 가 미완료 사용자에 대해 allowlist 외 모든 endpoint 를 `403 Forbidden` + `{ code: onboarding_required }` 로 차단 (REQ-FR-ONBOARD-009).
  - Allowlist (backend endpoint 만 — frontend 정적 페이지는 backend 호출 없이 렌더되므로 본 정책과 무관):
    - `GET /api/v1/me`
    - `POST /api/v1/me/onboarding`
    - `GET /api/v1/organizations/search`
    - `GET /api/v1/organization/hierarchy` (트리 picker 소스, 기존 endpoint)
    - 정적/health endpoint (예: `GET /health`)
  - lazy auto-create 폐기 — `authenticateActor` 는 DB row miss 를 정상 상태 (token-only actor) 로 취급. AuthenticatedActor 의 Email/DisplayName 은 token claim 에서 직접 추출.
- **Frontend** — UX layer (3분기, REQ-FR-ONBOARD-010).
  - 첫 진입 (session-scoped skip flag 미설정): `/devhub/onboarding` 으로 즉시 redirect.
  - skip 액션 이후 (sessionStorage 의 `devhub.onboarding.skipped=true`): 자동 redirect 없음 + 모든 페이지 상단에 dismissible banner.
  - 보호 리소스 진입 시도 (backend `403 onboarding_required`): skip 여부 무관 hard redirect to `/devhub/onboarding`.
  - sessionStorage 선택 이유: 세션 단위로만 보존 (탭 닫기 = reset), 매 로그인 시 onboarding 재강제 (REQ-FR-ONBOARD-011 의 "사실상의 reminder" 정합).

## 4. RBAC / Route permission 정책 (ARCH-ONBOARD-04)

- 신규 RBAC resource **추가 없음**. 본 도메인의 권한 분기는 route-level 만으로 cover.
- Route permission table 갱신:

| Endpoint | RBAC 요구 | onboardingGate | 비고 |
| --- | --- | --- | --- |
| `GET /api/v1/me` | (인증만, 모든 role) | **allowlist** | onboarding_required 분기 |
| `POST /api/v1/me/onboarding` | (인증만, 모든 role — token-only actor 도 호출 가능) | **allowlist** | 미완료 사용자가 제출 가능해야 하므로 gate bypass |
| `GET /api/v1/organizations/search` | (인증만, 모든 role) | **allowlist** | 모든 사용자에게 모든 조직 노출 (REQ-FR-ONBOARD-004) |
| `GET /api/v1/organization/hierarchy` | (인증만, 모든 role, 기존 endpoint) | **allowlist** | 트리 picker 소스 (§3 allowlist 정합) |
| `PATCH /api/v1/me` | (인증만, 본인) | **외 (차단)** | 완료 사용자만 호출 — `pending_review` 재진입 부수 효과. 미완료 사용자는 `POST /me/onboarding` 으로 첫 제출 |
| `POST /api/v1/users` | `users:create` (system_admin) | **외 (차단)** | admin 사전 등록 — admin 자신은 항상 완료 사용자 |
| `POST /api/v1/admin/users/:user_id/review` | `users:edit` (system_admin) | **외 (차단)** | review_status transition |

- pending_review 사용자의 "무소속" 처리는 RBAC 레벨이 아닌 **business logic 레벨** — 할당 리소스 조회 시 `user.primary_unit_id` 를 검토 상태에 따라 NULL 로 취급하거나 query filter 적용. 정확한 구현은 IMPL carve 에서.

## 5. 데이터 모델 (ARCH-ONBOARD-05)

`users` 테이블에 다음 컬럼 신규 추가 (REQ-NFR-ONBOARD-006):

```text
users  (기존 컬럼 + 신규 2개)
  ... (기존 컬럼은 ARCH-12, ARCH-15 등 참조)
  onboarding_completed_at   timestamptz   NULLABLE   -- 완료 시점 마킹 (NULL = 미완료)
  review_status             text          NULLABLE   -- 'pending_review' | 'reviewed'

  CONSTRAINT users_review_status_check
    CHECK ( review_status IS NULL OR review_status IN ('pending_review', 'reviewed') );
  CONSTRAINT users_onboarding_review_consistency
    CHECK ( (onboarding_completed_at IS NULL) = (review_status IS NULL) );
```

- `onboarding_completed_at IS NULL` ↔ `review_status IS NULL` — onboarding 미완료 사용자는 검토 단계가 적용되지 않는다는 의미. CHECK 제약으로 데이터 무결성 보장.
- `review_status='pending_review'` row 는 onboarding 제출 직후 자동 생성. `review_status='reviewed'` 는 system_admin 의 명시 transition.
- 마이그레이션 파일: `backend-core/migrations/0000XX_user_onboarding_state.up.sql` (번호는 다음 sequential — 기존 마지막 마이그레이션 + 1, IMPL carve 에서 확정).
- 기존 row 처리 — 본 컬럼 추가 이전에 lazy auto-created 되어 있던 row 는 `onboarding_completed_at=NULL` + `review_status=NULL` (기본 NULLABLE) 로 시작. 사용자가 다음 로그인 시 onboarding 강제 진입을 통해 정상 흐름 진입 — backfill 불요 (REQ-FR-ONBOARD-001 의 "row 미존재 OR completed_at NULL = 미완료" 정합).

## 6. Audit action 카탈로그 (ARCH-ONBOARD-06)

| action | target_type | payload | 트리거 |
| --- | --- | --- | --- |
| `account.onboarding_completed` | `user` | `{ user_id, primary_unit_id, display_name }` | `POST /api/v1/me/onboarding` 성공 |
| `account.review_confirmed` | `user` | `{ user_id, primary_unit_id, reviewed_by }` | `POST /api/v1/admin/users/:user_id/review` 성공 |
| `account.unit_changed` | `user` | `{ user_id, primary_unit_id_from, primary_unit_id_to, by_user }` | `PATCH /api/v1/me` 의 primary_unit_id 변경 또는 admin 의 unit reassignment. 부수 효과로 `review_status=pending_review` 재진입. |

- **Skip 자체는 audit emit 안 함** — state 변경 없음 (REQ-FR-ONBOARD-011 정합). 매 로그인 시 onboarding 화면 강제 진입이 사실상의 reminder 역할.
- 기존 `account.lazy_provisioned` event (ADR-0020 sub-carve B PR #239) 는 lazy auto-create 폐기와 함께 **deprecated** — 신규 row 는 모두 `account.onboarding_completed` 로 기록. 기존 emit 이력은 audit_logs 에 보존 (immutable).
- ADR-0019 §5.3 (9) Keycloak admin event listener 와의 관계: Keycloak group/role 변경은 `audit/user_sync.go` (sub-carve C PR #241) 가 별도 audit emit. 본 도메인의 `account.unit_changed` 는 **DevHub 내 self-service / admin transition** 만 발급.

## 7. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-29 | Phase 3 split — master `docs/architecture.md` §9 (Onboarding 본문) 을 도메인 sub-document 로 이관. ID(ARCH-ONBOARD-01..06) 보존, 신규 발급/삭제 없음. |
