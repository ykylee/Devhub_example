---
title: requirements
type: source
tags: [domain, requirements.md, project-devhub]
sources: [raw/projects/devhub/docs/domain/onboarding/requirements.md]
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

# onboarding 도메인 요구사항

- 문서 목적: Keycloak 인증 통과 + DevHub 프로필 미완료 사용자의 self-service 초기 등록 흐름의 기능·비기능 요구사항을 정의한다.
- 범위: REQ-FR-ONBOARD-001..012 / REQ-NFR-ONBOARD-001..008. ADR-0020 (계정/사용자 관리 책임 경계) 의 직접 후속.
- 대상 독자: backend / 프론트엔드 / DevOps, AI agent, QA.
- 상태: accepted
- 최종 수정일: 2026-05-29 (Phase 3 split, master `docs/requirements.md` §5.7 본문 이관)
- 관련 문서: [도메인 README](./README.md), [컨셉](./concept.md), [architecture.md](./architecture.md), [api.md](./api.md), [master requirements](../../requirements.md), [ADR-0019](../../adr/0019-keycloak-only-idp.md), [ADR-0020](../../adr/0020-account-user-management-boundary.md), [ADR-0021](../../adr/0021-onboarding-self-service-unit-selection.md)

## 1. 개요

본 도메인은 컨셉 문서([`./concept.md`](./concept.md), sprint `claude/keycloak-onboarding-concept-2026-05-21`, PR #260 + PR #265)에 정의된 Keycloak 인증 통과 + DevHub 프로필 미완료 사용자의 self-service 초기 등록 흐름의 요구사항을 정의한다. 본 도메인은 ADR-0020 (계정/사용자 관리 책임 경계) 의 직접 후속 — DevHub 가 책임지는 "사용자 프로필" 영역에 self-service unit selection 경로를 추가한다.

## 2. 기능 요구사항 (REQ-FR-ONBOARD)

- **REQ-FR-ONBOARD-001 (MVP):** Keycloak 인증 통과 + DevHub 프로필 미완료 사용자가 보호 리소스에 접근할 때 onboarding 경로로 유도되어야 한다.
    - "미완료" 정의: `users` row 미존재 (DB 미등록) OR `onboarding_completed_at IS NULL` (DB 등록-미완료).
    - 미완료 판단의 source-of-truth 는 backend `/api/v1/me` 응답의 `onboarding_required: true` 플래그.
- **REQ-FR-ONBOARD-002 (MVP):** Onboarding 입력 항목은 `display_name`, `primary_unit_id` 두 필드로 제한한다.
    - role 입력/선택 불가 — role 은 Keycloak token claim (`realm_access.roles`) 매핑 또는 시스템 기본값(`developer`) 으로만 결정. "소속 선택 = 권한 상승" 경로 차단.
    - 부가 항목 (사진/아바타, 닉네임, 입사일, 연락처) 본 도메인 1차 범위 제외.
- **REQ-FR-ONBOARD-003 (MVP):** Onboarding 제출은 단일 트랜잭션으로 (a) `users` row 신규 INSERT (또는 사전 등록된 row UPDATE), (b) `onboarding_completed_at = NOW()` 설정, (c) `review_status = 'pending_review'` 설정, (d) `account.onboarding_completed` audit event 기록 — 4 단계가 모두 이루어져야 한다. 부분 실패 시 모두 롤백한다.
- **REQ-FR-ONBOARD-004 (MVP):** 소속(unit) 선택 UX 는 검색(typeahead) + 트리(계층 선택기) 하이브리드를 제공한다.
    - 검색: 최소 2글자 입력 시 동작, 최대 20개 결과, 표시 포맷은 조직명만 사용.
    - 검색 endpoint: `GET /api/v1/organizations/search?q=...&limit=20` 신규 (**API-84**, [§16.4](../../backend_api_contract.md) 에서 spec staged).
    - 트리: 기존 `GET /api/v1/organization/hierarchy` 응답 재사용.
    - 권한 가드 없음 — 모든 사용자에게 모든 organization 후보가 노출된다.
- **REQ-FR-ONBOARD-005 (MVP):** 검토 상태 머신은 `pending_review → reviewed` 로 한정한다.
    - 제출 직후 자동 `pending_review`. 관리자가 검토 후 `reviewed` 로 수동 전이.
    - `pending_review` 사용자는 시스템에서 **무소속**으로 취급한다.
- **REQ-FR-ONBOARD-006 (MVP):** 미완료 사용자 접근 단계는 **3단계**로 운영한다.
    - `limited` (skip 상태, DB row 미존재): 공통 메뉴 + `/devhub/onboarding` 페이지 + `GET /api/v1/me` 만 접근.
    - `pending_review` (DB row 존재 + `onboarding_completed_at IS NOT NULL` + `review_status='pending_review'`): 공통 메뉴 + 할당된 과제/저장소/어플리케이션 접근.
    - `reviewed` (DB row 존재 + `review_status='reviewed'`): 정상 접근.
- **REQ-FR-ONBOARD-007 (MVP):** Onboarding 완료 후 사용자는 `/account` 페이지에서 자신의 `primary_unit_id` 를 self-service 로 변경할 수 있어야 한다.
    - 변경 시 `review_status` 가 `pending_review` 로 되돌려져야 한다.
    - 관리자가 재검토하여 다시 `reviewed` 로 확정해야 한다.
    - `pending_review` 재진입 기간에는 REQ-FR-ONBOARD-006 의 `pending_review` 단계 접근 정책이 적용된다.
- **REQ-FR-ONBOARD-008 (MVP):** 관리자는 사용자 사전 등록 endpoint 를 사용해 `users` row 를 사전 생성할 수 있어야 한다.
    - 사전 등록 시 입력 허용 필드: `display_name`, `primary_unit_id` (onboarding payload 와 동일 범위).
    - **role 은 사전 등록 payload 에 포함되지 않는다** — onboarding 과 동일하게 Keycloak token claim (`realm_access.roles`) 매핑 또는 시스템 기본값 (`developer`) 으로만 결정 (REQ-FR-ONBOARD-002 정합). 관리자도 사전 등록 시 role 임의 설정 불가.
    - 사전 등록 시 `onboarding_completed_at` 은 설정하지 않는다 (`NULL` 유지).
    - 사전 등록된 사용자도 첫 로그인 시 onboarding 화면에서 정보를 확인/수정한 후 제출해야 완료 처리된다.
    - 추가 필드 확장 (예: status, joined_at 의 관리자 override) 가능성은 IMPL carve 에서 결정.
- **REQ-FR-ONBOARD-009 (MVP):** Backend 는 미완료 사용자에 대해 allowlist 외 모든 endpoint 를 `403 Forbidden` + body `{ "code": "onboarding_required", ... }` 로 차단해야 한다.
    - allowlist (backend endpoint 만 — frontend 정적/공통 페이지는 backend 호출 없이 렌더되므로 본 정책과 무관):
        - Onboarding 제출 API (예: `POST /api/v1/me/onboarding`).
        - Organizations search API (예: `GET /api/v1/organizations/search`).
        - `GET /api/v1/me`.
        - 정적/공통 페이지가 호출하는 최소 backend endpoint (예: 정적 metadata, health check 등 인증 자체와 분리된 endpoint). 최종 allowlist 구성은 [ARCH §9.3](./architecture.md) 에서 확정 — `GET /api/v1/me` + `POST /api/v1/me/onboarding` + `GET /api/v1/organizations/search` + `GET /api/v1/organization/hierarchy` (트리 picker) + `/health`.
    - 기존 lazy auto-create 폐기 — `authenticateActor` 는 DB row miss 를 정상 상태 (token-only actor) 로 취급한다.
- **REQ-FR-ONBOARD-010 (MVP):** Frontend 는 `/api/v1/me` 의 `onboarding_required: true` 응답에 대해 **3분기** 로 동작해야 한다.
    - 첫 진입 (session 내 skip 액션 미실행): `/devhub/onboarding` 으로 즉시 redirect.
    - skip 액션 이후 (session-scoped skip flag set): 자동 redirect 없음 + 모든 페이지 상단에 dismissible banner 노출.
    - 보호 리소스 진입 시도 (backend `403 onboarding_required` 반환): skip 여부 무관 hard redirect.
- **REQ-FR-ONBOARD-011 (MVP):** Onboarding 화면은 "나중에 하기" (skip) 액션을 제공해야 한다.
    - skip 시 `users` row 를 생성하지 않는다.
    - skip 횟수/시간 제한 없음 — 매 로그인 시 onboarding 강제 진입이 사실상의 reminder 로 동작한다.
    - skip 자체는 audit event 를 발생시키지 않는다.
- **REQ-FR-ONBOARD-012 (후속):** Onboarding 완료 후 `/account` 에서 변경 가능한 추가 프로필 필드 (사진/아바타, 닉네임, 연락처 등) 는 본 도메인 1차 범위 밖이다. 후속 carve 에서 결정.

## 3. 비기능 / 운영 요구사항 (REQ-NFR-ONBOARD)

- **REQ-NFR-ONBOARD-001 (MVP):** UI 언어는 한국어 고정 (영문 UI 본 범위 제외).
    - 이름 표기는 단일 `display_name` 필드 자유 입력 (한글/영문/혼용 허용), 별도 영문명 필드 없음.
    - 확장성: API/DB/프론트 모델은 추후 영문 프로필 필드 (예: `display_name_en`) 가 nullable 로 추가 가능한 구조를 유지한다.
- **REQ-NFR-ONBOARD-002 (MVP):** 접근성 (a11y) 최소 기준.
    - 모든 입력 필드에 label 연결.
    - 키보드만으로 검색/선택/제출 가능.
    - 에러는 색상만으로 전달하지 않고 텍스트 메시지 병행.
    - 포커스 순서/가시성 보장.
    - organization picker 는 `combobox` role 및 ARIA 속성 (`aria-expanded`, `aria-controls`, `aria-activedescendant`) 준수.
- **REQ-NFR-ONBOARD-003 (MVP):** 제출/검증 UX.
    - 필수값 누락 시 필드별 인라인 에러 표시.
    - 제출 성공/실패 상태는 `aria-live` 영역으로 전달한다.
- **REQ-NFR-ONBOARD-004 (MVP):** 모바일 반응형은 본 범위에서 제외 (데스크탑 우선).
- **REQ-NFR-ONBOARD-005 (MVP):** 단일 포트 정합 (ADR-0018 + [`docs/reports/2026-05-20-network-docker-single-port-review.md`](../../reports/2026-05-20-network-docker-single-port-review.md)).
    - Onboarding 흐름에서 다른 host:port 로의 redirect 발생 금지 — same-origin 내부 path-relative redirect 만 허용한다.
    - Backend 의 `c.Redirect` / `Location:` 헤더 직접 작성 = 0 hit 가드 유지.
- **REQ-NFR-ONBOARD-006 (MVP):** 마이그레이션 — `users` 테이블에 다음 컬럼을 신규 추가한다.
    - `onboarding_completed_at TIMESTAMP NULL` — 완료 시점 마킹 (`NULL` = 미완료).
    - `review_status` (열거형 또는 텍스트 + CHECK 제약, 값: `pending_review`, `reviewed`) — 관리자 검토 단계.
    - 컬럼 명/타입은 [ARCH §9.5](./architecture.md) 에서 확정: `onboarding_completed_at timestamptz NULLABLE` + `review_status text NULLABLE` + bi-implication CHECK 제약 (`completed_at NULL ↔ review_status NULL`).
- **REQ-NFR-ONBOARD-007 (MVP):** Audit 정책.
    - Onboarding 완료 시점에 `account.onboarding_completed` event 를 audit_logs 에 기록한다 ([ARCH §9.6](./architecture.md) 에서 이름 확정).
    - Skip 자체는 audit event 미발생 (state 변경 없음).
    - 사용자 self-service 소속 변경은 `account.unit_changed` event 로 기록한다 ([ARCH §9.6](./architecture.md) 에서 이름 확정). 추가로 관리자 검토 transition 은 `account.review_confirmed` event 로 기록한다.
- **REQ-NFR-ONBOARD-008 (MVP):** 테스트 데이터 / 시드 세트는 단일 초기화/재적재 스크립트로 관리한다.
    - 계정 네이밍: `test_` prefix 고정.
    - 필수 시드:
        - `test_self_new_user`: DB 미등록 상태에서 첫 로그인 시 onboarding 진입 검증.
        - `test_admin_seeded_incomplete`: 관리자 사전등록 + `onboarding_completed_at=NULL` 상태 검증.
        - `test_completed_pending_review`: 완료 + `pending_review` 상태, 무소속 제한 접근 검증.
        - `test_completed_reviewed`: 완료 + `reviewed` 상태, 정상 접근 검증.
        - `test_reviewed_then_unit_change`: `reviewed` 사용자의 소속 self-service 변경 시 `pending_review` 재진입 검증.
        - `org_fixture_bulk`: organization 25개 이상 (2글자 검색 + 최대 20개 제한 + 조직명-only 표시 검증용).

## 4. 범위 경계 (Out of Scope)

- 사용자가 관리자에게 직접 문의/escalation 을 보내는 UI (concept §5.4 옵션 C 변형은 1차 범위 밖).
- Onboarding 완료 후의 추가 프로필 필드 (사진/아바타, 닉네임, 입사일, 연락처, 부서장 확인 등).
- 영문 UI / 다국어 지원 (REQ-NFR-ONBOARD-001 의 확장 nullable 필드 정의만 본 범위).
- 모바일 반응형 디자인.
- HRDB 자동 cross-check 또는 Keycloak group → unit 자동 매핑 (concept §5.4 옵션 C/D — 사전 carve 의존).
- Onboarding 완료 후의 `review_status` reversal 정책 (예: 재교육/재인증 필요 시 admin 이 `reviewed` → `pending_review` 강제 되돌리기) — 운영 정책 결정 후 후속 carve.
- MFA / 2FA — ADR-0019 §5.3 sub-carve 와 분리 (§2.5 사용자 계정 관리의 MFA 비도입 기준과 동일).

## 5. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-29 | Phase 3 split — master `docs/requirements.md` §5.7 에서 본문 그대로 이관. ID(REQ-FR-ONBOARD-001..012, REQ-NFR-ONBOARD-001..008) 보존, 신규 발급/삭제 없음. |
