# Keycloak Admin 책임 분리 협약 (IdP 팀 ↔ DevHub 운영자)

- 문서 목적: 외부 Keycloak 시나리오 하에서 사내 IdP 팀과 DevHub 운영자 간의 책임 경계를 governance 정책 문서로 명문화한다. [ADR-0020 §3.2](../adr/0020-account-user-management-boundary.md) 표를 사내 정책으로 승격 + ADR-0021 §3.1 의 onboarding 흐름 확장 반영.
- 범위: 사용자 lifecycle (생성/비밀번호/disable/삭제/group membership/unit assignment/role 변경/RBAC policy) + JWKS rotation + Keycloak realm 자산 + audit trail. 본 문서가 책임 매트릭스 + escalation path 의 single source-of-truth.
- 대상 독자: 사내 IdP 팀, DevHub 운영자 (SRE), system_admin, security, 외부 감사.
- 상태: draft (1차 — 사내 검토 동반 필요)
- 최종 수정일: 2026-05-22
- 결정 근거 sprint: `claude/work_260522-internal-coordinated-carve-docs`
- 관련 문서: [ADR-0019 Keycloak 단일화](../adr/0019-keycloak-only-idp.md), [ADR-0020 계정/사용자 책임 경계](../adr/0020-account-user-management-boundary.md), [ADR-0021 Onboarding self-service](../adr/0021-onboarding-self-service-unit-selection.md), [Keycloak 운영 SOP](../setup/keycloak_operations.md), [JWKS rotation cache flush SOP](../setup/jwks_rotation_cache_flush.md), [HRDB unit pre-stage 가이드](../setup/hrdb_unit_pre_stage.md), [governance/README](./README.md).

## 1. 배경

[ADR-0019](../adr/0019-keycloak-only-idp.md) 가 DevHub 의 IdP 를 Keycloak 단일화로 결정한 후, [ADR-0020](../adr/0020-account-user-management-boundary.md) 가 외부 Keycloak 시나리오 (사내 IdP 팀이 별도 운영하는 Keycloak instance) 의 책임 경계를 명시했다. 본 문서가 ADR-0020 §3.2 표를 **사내 정책 문서로 승격** — ADR (architecture decision record) 의 결정을 운영 협약 (operational contract) 으로 전환.

[ADR-0021](../adr/0021-onboarding-self-service-unit-selection.md) 의 onboarding self-service 흐름은 ADR-0020 의 책임 경계를 **확장** (reversal 아님) — 사용자 self-service onboarding + DevHub admin 검토 path 추가. 본 문서가 두 ADR 의 결정을 통합 표로 정리.

### 1.1 외부 Keycloak 시나리오 정의

DevHub 가 사용하는 Keycloak instance 는 **사내 IdP 팀이 별도 운영**한다. DevHub 운영자는 Keycloak admin console 권한을 가지지 않는다 (사내 정책). DevHub backend 의 service account (`devhub-backend` client) 는 `view-users` + `view-events` realm role 만 보유 (ADR-0020 sub-carve E PR #244 정합).

이 가정 하에서 사용자 lifecycle / RBAC / audit trail 의 각 동작이 어느 팀의 책임인지를 본 문서가 정의.

## 2. 책임 매트릭스 (Master)

### 2.1 사용자 lifecycle

| 운영 동작 | 책임 주체 | 도구 | Audit |
| --- | --- | --- | --- |
| user 생성 (account.create) | **IdP 팀** | Keycloak admin console **또는** HRDB ETL push (사내 ETL → Keycloak Admin REST) | Keycloak audit event log + DevHub audit_logs (event listener sync, [Keycloak 운영 SOP §8.6](../setup/keycloak_operations.md)) |
| user 비밀번호 reset | **IdP 팀** | Keycloak admin console (Credentials 탭) | Keycloak audit event log |
| user 본인 비밀번호 변경 | **사용자 본인** | Keycloak Account Console (DevHub `/account` 에 "Open Keycloak Console" 외부 link, [keycloak_operations §8.5b](../setup/keycloak_operations.md)) | Keycloak audit event log |
| user status disable / enable | **IdP 팀** | Keycloak admin console **또는** HRDB ETL push | Keycloak audit event log + DevHub `users.status` event sync ([Keycloak 운영 SOP §8.6](../setup/keycloak_operations.md)) |
| user 삭제 | **IdP 팀** | Keycloak admin console | Keycloak audit event log + DevHub `users.status='deactivated'` soft delete (ADR-0020 sub-carve C) |
| group membership 변경 (role 변경) | **IdP 팀** | Keycloak admin console (User detail → Groups 탭) | Keycloak audit event log + DevHub `users.role` event sync |

> **`users.role` 직접 수정 금지** — Keycloak group composite role 매핑 + event listener 자동 sync 가 sole policy. DevHub API 또는 DB direct UPDATE 로 `users.role` 수정 안 됨.

### 2.2 사용자 조직 unit assignment

[ADR-0021 §3.1](../adr/0021-onboarding-self-service-unit-selection.md) 의 책임 경계 확장 반영.

| 운영 동작 | 책임 주체 | 도구 | Audit |
| --- | --- | --- | --- |
| **신규 user 의 unit 초기 배치** | **사용자 self-service onboarding** | DevHub `/devhub/onboarding` (`POST /api/v1/me/onboarding`, API-83) | DevHub audit_logs `account.onboarding_completed` |
| **기존 user 의 unit 변경 (self-service)** | **사용자 본인** | DevHub `/account` (`PATCH /api/v1/me`, API-85) — `review_status='pending_review'` 자동 재진입 | DevHub audit_logs `account.unit_changed` |
| **사용자 등록 후 검토 (`pending_review → reviewed`)** | **DevHub system_admin** | DevHub `/admin/settings/users` (`POST /api/v1/admin/users/:user_id/review`, API-86) | DevHub audit_logs `account.review_confirmed` |
| **사용자 사전 등록** | **DevHub system_admin** | DevHub `/admin/settings/users` (`POST /api/v1/users`, API-33 확장) — `onboarding_completed_at=NULL`, 사용자 첫 로그인 시 onboarding 강제 진입 | DevHub audit_logs `user.created` |
| **HRDB ETL pre-stage 자동 unit 매핑** (carve) | **IdP 팀 + 사내 HRDB 운영팀** | HRDB ETL → Keycloak user attribute `unit_id` → DevHub onboarding cross-check ([HRDB unit pre-stage 가이드](../setup/hrdb_unit_pre_stage.md)) | HRDB ETL log + Keycloak audit event log |

### 2.3 RBAC + DevHub 조직 자체

| 운영 동작 | 책임 주체 | 도구 | Audit |
| --- | --- | --- | --- |
| 조직 unit (department/team) CRUD | **DevHub system_admin** | DevHub `/admin/settings/organization` (`/api/v1/organization/*`) | DevHub audit_logs |
| RBAC policy (role × resource × action) 편집 | **DevHub system_admin** | DevHub `/admin/settings/permissions` (`/api/v1/rbac/policies`) | DevHub audit_logs |
| RBAC role definition (role_id, description) | **DevHub system_admin** | DevHub `/admin/settings/permissions` (`/api/v1/rbac/roles`) | DevHub audit_logs |

### 2.4 IdP 자산 (Keycloak realm / client / JWKS)

| 운영 동작 | 책임 주체 | 도구 | Audit |
| --- | --- | --- | --- |
| Keycloak realm 정의 (`devhub`) + client 정의 (`devhub-frontend`, `devhub-backend`) | **IdP 팀** | Keycloak admin console | Keycloak audit event log (CLIENT_INFO_*) |
| Keycloak realm role 정의 (`developer`/`manager`/`team_manager`/`system_admin`) + composite group 매핑 | **IdP 팀** | Keycloak admin console (Roles + Groups 탭) | Keycloak audit event log |
| JWKS rotation (signing key 변경) | **IdP 팀** | Keycloak admin console (Realm settings → Keys) | Keycloak audit event log + DevHub backend cache flush 동반 ([JWKS rotation cache flush SOP](../setup/jwks_rotation_cache_flush.md)) |
| `devhub-backend` service account 의 client_secret rotation | **IdP 팀 + DevHub 운영자** | Keycloak admin console (client secret 발급) + DevHub vault 갱신 | Keycloak audit event log + DevHub vault audit |
| Valid Redirect URIs / Post Logout Redirect URIs allowlist | **IdP 팀** (요청은 DevHub 운영자) | Keycloak admin console (client Settings 탭) | Keycloak audit event log |

### 2.5 운영 monitoring + incident response

| 운영 동작 | 책임 주체 | 도구 | 비고 |
| --- | --- | --- | --- |
| Keycloak 서비스 가용성 모니터링 | **IdP 팀** | Keycloak operational dashboard | DevHub 가 ADR-0019 §3.1 따라 단일 IdP 의존 — Keycloak down = DevHub 로그인 차단 |
| Keycloak ↔ DevHub audit_logs sync (event listener) | **DevHub 운영자** | [Keycloak 운영 SOP §8.6](../setup/keycloak_operations.md) | Polling 기반 (`audit/keycloak_event_puller.go`, ADR-0019 §5.3 (9) Phase 2 PR-E) |
| JWKS cache 상태 (5분 TTL + 24h stale-while-error) | **DevHub 운영자** | Prometheus metric `devhub_jwks_fetch_total` + `devhub_jwks_cache_age_seconds` ([Keycloak 운영 SOP §8.4](../setup/keycloak_operations.md)) | Stale fallback 사용 시 alarm |
| DevHub 도메인 API monitoring | **DevHub 운영자** | DevHub `/admin/settings/audit` + Prometheus | ADR-0011 RBAC + 도메인별 audit action |
| Onboarding flow monitoring | **DevHub 운영자** | [Onboarding 운영 SOP](../setup/onboarding_operations.md) §4 + DoD §7 | Feature flag `DEVHUB_ONBOARDING_GATE_ENABLED` rollback path 동반 |

## 3. Escalation Path

### 3.1 Level 1 — DevHub 운영자 (SRE / on-call) 자체 처리

- DevHub 도메인 API 회귀 / 운영 신호 임계 도달 — 본 문서 §2.5 의 monitoring 도구로 1차 진단
- Onboarding SOP §5 의 rollback (`DEVHUB_ONBOARDING_GATE_ENABLED=0`) 실행
- JWKS rotation cache flush ([JWKS rotation cache flush SOP](../setup/jwks_rotation_cache_flush.md)) 의 backend 재기동 실행

### 3.2 Level 2 — DevHub backend / frontend 담당 (Claude / Gemini)

- backend 로직 회귀 — `internal/httpapi/*`, `internal/auth/*`, `internal/audit/*` 의 bug fix PR
- frontend 회귀 — `app/`, `components/`, `lib/` 의 bug fix PR
- ADR-0020 / ADR-0021 의 결정 정합 검토 — `docs/adr/` + ADR governance

### 3.3 Level 3 — 사내 IdP 팀

- Keycloak instance 가용성 / realm config / JWKS rotation / event listener 측면
- `KC_BOOTSTRAP_ADMIN_USERNAME` vs `KEYCLOAK_ADMIN` env 호환성 ([ADR-0022](../adr/0022-keycloak-version-pin-25-0.md) Keycloak 25.0 pin)
- 사용자 lifecycle (생성/disable/group 변경/삭제) — DevHub 운영자가 admin console 권한 없음

### 3.4 Level 3 — 사내 HRDB / DBA 팀

- HRDB ETL push → Keycloak user attribute 매핑 경로 ([HRDB unit pre-stage 가이드](../setup/hrdb_unit_pre_stage.md))
- DevHub PostgreSQL `users` 테이블 schema 정합 — migration 충돌 / CHECK 제약 손상
- ADR-0008 의 `hrdb` schema 정합

## 4. 운영 정책 — 명시 금지 동작

다음 동작은 **누구도 수행 금지**:

| 금지 동작 | 이유 |
| --- | --- |
| DevHub DB 의 `users.role` 컬럼 직접 UPDATE | Keycloak group composite role 매핑 + event listener sync 가 sole policy. 직접 UPDATE 시 audit emit 우회 + 다음 event sync 에서 덮어쓰임. |
| DevHub DB 의 `users.onboarding_completed_at` 컬럼 직접 UPDATE | `users_onboarding_review_consistency` CHECK 제약 손상 위험 + audit emit 우회. ADR-0021 §3 의 onboarding 흐름이 정공법. |
| Keycloak realm `master` 의 admin user 권한 DevHub backend 에 부여 | Service account 권한 최소화 (ADR-0020 sub-carve E) 위반. `devhub-backend` 는 `view-users` + `view-events` 만 보유. |
| Keycloak realm `devhub` 의 backup IdP fallback 도입 | ADR-0019 §3.1 단일 IdP 결정 위반 (옵션 E 명시 제외, ADR-0019 §5.3 (10) carve). |
| `/auth/login` URL 사용 (deprecated 2026-05-22 PR #295) | `/login` 이 canonical entry. 외부 link / bookmark / Keycloak post.logout.redirect.uris 모두 `/login` 으로 정합. |

## 5. 변경 절차

본 문서의 책임 매트릭스 (§2) 변경은 다음 절차 따름:

1. **변경 제안** — 사내 IdP 팀 + DevHub 운영자 간 사전 협의
2. **ADR 검토** — 변경이 ADR-0019/0020/0021 의 결정 영향 시 신규 ADR 발급 (`docs/adr/00XX-*.md`)
3. **본 문서 갱신** — §2 표 row 수정 + §6 변경 이력 row 추가
4. **사내 검토** — IdP 팀 + Security + 외부 감사 (필요 시) sign-off
5. **운영 적용** — Keycloak admin console 작업 / DevHub 운영자 작업 (각 측면)

## 6. 변경 이력

| 일자 | 변경 | sprint |
| --- | --- | --- |
| 2026-05-22 | 1차 draft — ADR-0020 §3.2 + ADR-0021 §3.1 표 통합 + §2 책임 매트릭스 5 sub-section (사용자 lifecycle / 조직 unit assignment / RBAC + DevHub 조직 / IdP 자산 / monitoring + incident) + §3 escalation 4 level + §4 명시 금지 5건 + §5 변경 절차. 사내 검토 동반 필요 — Draft → Accepted 승격은 IdP 팀 + Security sign-off 후. | `claude/work_260522-internal-coordinated-carve-docs` |
