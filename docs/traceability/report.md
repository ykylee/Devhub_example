# Traceability Report — 1차 종합

- 문서 목적: DevHub 의 SDLC 자산 (요구사항 → 설계 → 로드맵 → 구현 → 단위테스트 → E2E) 사이 추적 관계를 단일 매트릭스로 시각화한다.
- 범위: M0–M3 (M4 는 planned 표기). ADR 은 별도 §4 인덱스.
- 대상 독자: 모든 contributor, 후속 리뷰어, 외부 감사.
- 상태: accepted
- 최종 수정일: 2026-05-13
- 결정 근거 sprint: `claude/work_260513-c`.
- 관련 문서: [`README.md`](./README.md), [`conventions.md`](./conventions.md), [`sync-checklist.md`](./sync-checklist.md), [`../governance/document-standards.md`](../governance/document-standards.md).

## 1. 사용법

- 단계 간 추적 관계를 한 페이지에서 찾을 때 §3 종합 매트릭스 사용.
- 특정 단계의 ID 정의를 찾을 때 §2 단계별 인덱스 사용.
- ADR 은 §4 의 링크.
- 갱신 절차는 [`sync-checklist.md`](./sync-checklist.md) §3.6 참조.

## 2. 단계별 인덱스 (요약)

### 2.1 Requirements (REQ)

- **Functional**: REQ-FR-01 ~ REQ-FR-105 (총 105 항목, `docs/requirements.md` §2–§5 + `docs/backend/requirements.md` §1–§5 + `docs/backend_requirements_org_hierarchy.md` §1–§3 + `docs/frontend_integration_requirements.md` §2–§3 분포).
- **Non-functional**: REQ-NFR-01 ~ REQ-NFR-26 (총 26 항목, 보안/성능/배포 정책 + 운영 hygiene + API 표준).

### 2.2 Design (ARCH / API)

- **Architecture**: ARCH-01 ~ ARCH-17 (총 17 항목, `docs/architecture.md` + `docs/org_chart_ux_spec.md` + `docs/organizational_hierarchy_spec.md` 분포).
- **API contract**: API-01 ~ API-40 — *ID 공간 = 40*. 본 sprint `claude/work_260513-i` / `-j` 의 결정으로 일부 ID 는 composite (`API-07` = infra/edges + infra/topology, `API-13` = risks + risks/critical) 또는 결손 (`API-03` 미정의 — 후속 발급 후보) 이 존재한다. 실제 endpoint 매핑은 본 §2.2 아래 도메인별 서브 표 (RBAC / Auth / Infra+Dashboard / Pipelines / Realtime+Command+Audit / Account+Org+Me) 가 source-of-truth.

#### RBAC API §12 — endpoint 매핑 (sprint `claude/work_260513-f`, 본문 ID 노출 1차 도메인)

| API ID | 본문 위치 | 항목 |
| --- | --- | --- |
| `API-26` | §12.2 | `GET /api/v1/rbac/policies` |
| `API-27` | §12.3 | `PUT /api/v1/rbac/policies` |
| `API-28` | §12.4 | `POST /api/v1/rbac/policies` (사용자 정의 role 생성) |
| `API-29` | §12.5 | `DELETE /api/v1/rbac/policies/:role_id` (사용자 정의 role 삭제) |
| `API-30` | §12.6 | `GET /api/v1/rbac/subjects/:subject_id/roles` |
| `API-31` | §12.7 | `PUT /api/v1/rbac/subjects/:subject_id/roles` |
| `API-38` | §12.8 | 라우트 → (resource, action) 매핑 표 (정책 정의) |
| `API-39` | §12.9 | 매핑 누락 정책 (deny-by-default) |
| `API-40` | §12.10 | Cache 와 무효화 정책 |

> 본 매핑 표는 `document-standards.md` §8 우선순위 3 (본문 ID 노출) 의 RBAC 도메인 1차 적용. 다른 도메인 (`account` §11.4, `commands/audit` §9 등) 은 후속 sprint 에서 동일 패턴으로 점진 확장.
>
> **표 가독성 정책**: 정밀 매핑 (endpoint 별 본문 위치, IMPL 별 책임) 은 본 §2 서브 표가 source-of-truth. §3 종합 매트릭스의 도메인 행은 ID 범위 (`API-26–31`) + 서브 표 참조 노트만 두어 표 시인성을 유지한다.

#### Auth API §11 — endpoint 매핑 (sprint `claude/work_260513-g`, 본문 ID 노출 2차 도메인)

| API ID | 본문 위치 | 항목 |
| --- | --- | --- |
| `API-19` | §11.3 | Go Core Bearer token 경계 (`Authorization: Bearer ...` verifier interface + actor context) |
| `API-20` | §11.5 | `POST /api/v1/auth/login` (login_challenge → Kratos api-mode login + Hydra accept) |
| `API-21` | §11.5 | `POST /api/v1/auth/logout` (logout_challenge → Hydra revoke + accept) |
| `API-22` | §11.5 | `POST /api/v1/auth/token` (authorization_code → Hydra `/oauth2/token`) |
| `API-23` | §11.5 | `POST /api/v1/auth/signup` (HRDB lookup + Kratos identity 생성) |
| `API-24` | §11.5 | `GET /api/v1/auth/consent` (Hydra consent flow auto-accept) |
| `API-35` | §11.5.1 | `POST /api/v1/account/password` (self-service password change — 인증/계정 도메인 cross-cut) |

> §11.2 Hydra 표준 endpoint (외부, DevHub 재정의 0) 와 §11.4 admin identity wrapper (planned, M3 진입 시 ID 발급) 는 본 매핑 표에서 제외 — `conventions.md` §5.2 의 외부 의존성 / planned 항목 정책.

#### Infra / Dashboard API §6 — endpoint 매핑 (sprint `claude/work_260513-i`)

| API ID | 본문 위치 | 항목 |
| --- | --- | --- |
| `API-01` | §3 | `GET /health` |
| `API-02` | §4 | `POST /api/v1/integrations/gitea/webhooks` |
| `API-04` | §5 | `GET /api/v1/events` |
| `API-05` | §6.200 | `GET /api/v1/dashboard/metrics` |
| `API-06` | §6.234 | `GET /api/v1/infra/nodes` |
| `API-07` | §6.240, §6.245 | `GET /api/v1/infra/edges` + `GET /api/v1/infra/topology` (composite) |

#### Pipelines API §7 — endpoint 매핑 (sprint `claude/work_260513-i`)

| API ID | 본문 위치 | 항목 |
| --- | --- | --- |
| `API-08` | §7 | `GET /api/v1/repositories` |
| `API-09` | §7 | `GET /api/v1/issues` |
| `API-10` | §7 | `GET /api/v1/pull-requests` |
| `API-11` | §7 | `GET /api/v1/ci-runs` |
| `API-12` | §7 | `GET /api/v1/ci-runs/:ci_run_id/logs` |
| `API-13` | §7 | `GET /api/v1/risks` + `GET /api/v1/risks/critical` (composite) |

#### Realtime / Command / Audit API §8, §9 — endpoint 매핑 (sprint `claude/work_260513-i`)

| API ID | 본문 위치 | 항목 |
| --- | --- | --- |
| `API-14` | §8 | `GET /api/v1/realtime/ws` (WebSocket endpoint) |
| `API-15` | §9.1 | `POST /api/v1/admin/service-actions` |
| `API-16` | §9.2 | `POST /api/v1/risks/:risk_id/mitigations` |
| `API-17` | §9.3 | `GET /api/v1/commands/:command_id` |
| `API-18` | §9.4 | `GET /api/v1/audit-logs` |
| `API-36` | §8 envelope | `command.status.updated` WebSocket event envelope |
| `API-37` | §11.6 | command lifecycle audit 매핑 (`auth.role_denied`, command-target audit action) |

#### Account / Organization / Me API §10.1 — endpoint 매핑 (sprint `claude/work_260513-i`)

| API ID | 본문 위치 | 항목 |
| --- | --- | --- |
| `API-25` | §10.1 (spec carve out) | `/api/v1/accounts/*` admin endpoint set (POST /accounts, PUT /accounts/:id/password, PATCH /accounts/:id/status, DELETE /accounts/:id) |
| `API-32` | §10 | `GET /api/v1/me` (현재 §10 의 1차 노출, 별도 spec 절은 후속 sprint) |
| `API-33` | §10.1 (spec carve out) | `/api/v1/users` CRUD set |
| `API-34` | §10.1 (spec carve out) | `/api/v1/organization/*` set (hierarchy + units + members) |

> §10.1 에 본문 spec 부재 endpoint 의 위치 표가 명시되어 있음. 본문 spec 작성은 후속 sprint 후보 (`work_backlog.md` 의 미진입 항목).

### 2.3 Roadmap (RM)

- **M0**: RM-M0-01 (1 항목, `X-Devhub-Actor` deprecation).
- **M1**: RM-M1-01 ~ RM-M1-04 (4 항목, RBAC track + API 계약 §11 재작성).
- **M2**: RM-M2-01 ~ RM-M2-16 (16 항목, 인증/계정/조직/UX/audit + CI). 1차 완성 sprint (PR #85) 가 이전에 M3 으로 분류됐던 사용자/조직 관리 대부분을 흡수.
- **M3**: 사용자 및 조직 관리 — 대부분 M2 1차 완성에서 흡수, 잔여 정의는 §2.3.1 표 참조.
- **M4**: 실시간 + AI Gardener + 과제 추적 + 시스템 관리. 정의는 §2.3.2 표 참조.

> **drift 정합 (2026-05-13, sprint `claude/work_260513-k`)**: 본 §2.3 + `docs/development_roadmap.md` §3 가 M3/M4 정의의 single source-of-truth. 매트릭스 §3 의 도메인 행 인용 + state.json + backend_roadmap §5 모두 이 정의 기준으로 정합화. 본 sprint 이전에는 매트릭스가 RM-M3 = "Sign Up + WebSocket + AI" 로 정의해 development_roadmap.md M4 항목을 cross-cut 한 drift 상태였다.

#### 2.3.1 RM-M3 정의 (sprint `claude/work_260513-k`)

| RM ID | 항목 | 출처 |
| --- | --- | --- |
| `RM-M3-01` | Sign Up (셀프 가입) — `POST /api/v1/auth/signup` 의 hrdb lookup arm | development_roadmap §3 M3 |
| `RM-M3-02` | 인사 DB 스키마 (`name`, `system_id`, `employee_id`, `department_name`) + `internal/hrdb/` 모듈 활용 | development_roadmap §3 M3 |
| `RM-M3-03` | 조직 polish — `backend_api_contract.md` §10.4 자세한 schema + `parent_id` 검증 + primary_dept 자동 판정 (§5 백로그) | development_roadmap §3 M3 + §5 백로그 |

#### 2.3.2 RM-M4 정의 (sprint `claude/work_260513-k`, 본 sprint 가 부여)

| RM ID | 항목 | 출처 |
| --- | --- | --- |
| `RM-M4-01` | WebSocket 확장 — `infra.node.updated` / `ci.run.updated` / `risk.updated` event publish | development_roadmap §3 M4 + backend_roadmap §2 Phase 8 |
| `RM-M4-02` | WebSocket replay (last event) + 리소스 필터링 | backend_roadmap §2 Phase 8 잔여 |
| `RM-M4-03` | frontend command status WebSocket UI (Phase 4 마무리) | development_roadmap §4.2 Dashboard |
| `RM-M4-04` | AI Gardener gRPC — Python `AnalysisService` server + Go Core client | backend_roadmap §2 Phase 9 |
| `RM-M4-05` | AI Suggestion Feed 실데이터 바인딩 (frontend) | development_roadmap §4.2 Gardener |
| `RM-M4-06` | Gitea Hourly Pull worker (과제 추적 reconciliation) | backend_roadmap §2 Phase 10 |
| `RM-M4-07` | System Admin 대시보드 (Gitea Runner 상태 + 시스템 설정 관리) | development_roadmap §3 M4 |
| `RM-M4-08` | RBAC PermissionCache 다중 인스턴스 일관성 구현 ([ADR-0007](../adr/0007-rbac-cache-multi-instance.md) PG `LISTEN/NOTIFY`) | ADR-0007 §4.2 |
| `RM-M4-09` | 외부 SSO 통합 (Gitea 연동 등) | development_roadmap §3 M4 |

### 2.4 Implementation (IMPL)

- **Backend (`backend-core`)**: IMPL-auth-01..07, rbac-01..04, audit-01..02, account-01..04, org-01..04, command-01..05, serviceaction-01, domain-01..03, dashboard-01, infra-01, store-01..03, gitea-01..02, kratos-01..04, config-01, hrdb-01, realtime-01, me-01, health-01, idp-schema-01 (47 항목).

#### IMPL-rbac-XX 정의 (sprint `claude/work_260513-f`)

| IMPL ID | 코드 위치 | 책임 |
| --- | --- | --- |
| `IMPL-rbac-01` | `backend-core/internal/httpapi/rbac.go` | API-26..31 의 6 endpoint handler + §6 legacy gone (`listRBACPolicies`, `createRBACPolicy`, `updateRBACPolicies`, `deleteRBACPolicy`, `getSubjectRoles`, `setSubjectRoles`, `getRBACPolicyLegacyGone`) |
| `IMPL-rbac-02` | `backend-core/internal/store/postgres_rbac.go` | RBAC role + subject-role assignment persistence (`ListRBACRoles`, `GetRBACRole`, `CreateRBACRole`, `UpdateRBACRolePermissions`, `UpdateRBACRoleMetadata`, `DeleteRBACRole`, `GetSubjectRoles`, `SetSubjectRole`) |
| `IMPL-rbac-03` | `backend-core/internal/httpapi/permissions.go` (`routePermissionTable` + `enforceRoutePermission`) | API-38 라우트 매핑 source-of-truth + 미들웨어 enforcement + API-39 deny-by-default |
| `IMPL-rbac-04` | `backend-core/internal/httpapi/permissions.go` (`PermissionCache`) | API-40 in-memory matrix cache + Invalidate |

#### IMPL-auth-XX 정의 (sprint `claude/work_260513-g`)

| IMPL ID | 코드 위치 | 책임 |
| --- | --- | --- |
| `IMPL-auth-01` | `backend-core/internal/auth/hydra_introspection.go` + `internal/httpapi/auth.go` (`BearerTokenVerifier` interface) | §11.3 Bearer token verifier interface + Hydra introspection 구현 |
| `IMPL-auth-02` | `internal/httpapi/auth.go` (`authenticateActor` middleware + `AuthenticatedActor` context) | §11.3 검증 결과를 request context 의 actor 로 전파 |
| `IMPL-auth-03` | `internal/httpapi/auth_login.go` | API-20 `POST /api/v1/auth/login` handler |
| `IMPL-auth-04` | `internal/httpapi/auth_logout.go` | API-21 `POST /api/v1/auth/logout` handler |
| `IMPL-auth-05` | `internal/httpapi/auth_token.go` | API-22 `POST /api/v1/auth/token` handler |
| `IMPL-auth-06` | `internal/httpapi/auth_signup.go` | API-23 `POST /api/v1/auth/signup` handler |
| `IMPL-auth-07` | `internal/httpapi/auth_consent.go` | API-24 `GET /api/v1/auth/consent` handler |

> API-35 `POST /api/v1/account/password` 의 IMPL 은 account 도메인 (별도 sprint 의 `internal/httpapi/account_password.go`). 매트릭스 §3 의 인증 행 + 계정 관리 행 양쪽에 ID 가 노출되는 cross-cut.

#### IMPL-account-XX 정의 (sprint `claude/work_260513-i`)

| IMPL ID | 코드 위치 | 책임 |
| --- | --- | --- |
| `IMPL-account-01` | `internal/httpapi/accounts_admin.go` (`createAccount`) | API-25 `POST /api/v1/accounts` — Kratos identity 생성 + DevHub users 행 + temp password |
| `IMPL-account-02` | `internal/httpapi/accounts_admin.go` (`resetAccountPassword`) | API-25 `PUT /api/v1/accounts/:id/password` — admin reset |
| `IMPL-account-03` | `internal/httpapi/accounts_admin.go` (`updateAccountStatus`, `deleteAccount`) | API-25 PATCH/DELETE — disable / delete identity + users |
| `IMPL-account-04` | `internal/httpapi/account_password.go` | API-35 `POST /api/v1/account/password` — self-service password change (Kratos settings flow proxy) |

#### IMPL-org-XX 정의 (sprint `claude/work_260513-i`)

| IMPL ID | 코드 위치 | 책임 |
| --- | --- | --- |
| `IMPL-org-01` | `internal/httpapi/organization.go` (users handlers: `listUsers`, `getUser`, `createUser`, `updateUser`, `deleteUser`) | API-33 `/api/v1/users` CRUD |
| `IMPL-org-02` | `internal/httpapi/organization.go` (hierarchy + units handlers: `getOrganizationHierarchy`, `getUnit`, `createUnit`, `updateUnit`, `deleteUnit`) | API-34 hierarchy + units endpoint |
| `IMPL-org-03` | `internal/httpapi/organization.go` (members handlers: `listUnitMembers`, `replaceUnitMembers`) | API-34 unit members endpoint |
| `IMPL-org-04` | `internal/store/users_units.go` + organization store impl | persistence layer (users / org_units / unit_memberships) |

#### IMPL-command-XX 정의 (sprint `claude/work_260513-i`)

| IMPL ID | 코드 위치 | 책임 |
| --- | --- | --- |
| `IMPL-command-01` | `internal/httpapi/commands.go` (`createServiceAction`) | API-15 `POST /api/v1/admin/service-actions` handler |
| `IMPL-command-02` | `internal/httpapi/commands.go` (`createRiskMitigation`) | API-16 `POST /api/v1/risks/:id/mitigations` handler |
| `IMPL-command-03` | `internal/httpapi/commands.go` (`getCommand`) | API-17 `GET /api/v1/commands/:command_id` handler |
| `IMPL-command-04` | `internal/commandworker/*` | command worker (dry-run / live executor, `command.status.updated` publisher) |
| `IMPL-command-05` | `internal/serviceaction/*` | service-action 도메인 로직 (executor adapter 후보) |

#### IMPL-audit-XX 정의 (sprint `claude/work_260513-i`)

> auth 의 IMPL-auth-01..07 과 별도. audit 는 독립 모듈.

| IMPL ID | 코드 위치 | 책임 |
| --- | --- | --- |
| `IMPL-audit-01` | `internal/httpapi/audit.go` (`listAuditLogs`, `recordAudit`, `recordAuditBestEffort`) | API-18 `GET /api/v1/audit-logs` handler + 다른 도메인의 audit emit helper |
| `IMPL-audit-02` | `internal/store/audit_logs.go` | audit_logs persistence + source_type/request_id/source_ip 컬럼 (M1 PR-D 의 actor enrichment) |

#### IMPL-infra-XX / dashboard 정의 (sprint `claude/work_260513-i`)

| IMPL ID | 코드 위치 | 책임 |
| --- | --- | --- |
| `IMPL-infra-01` | `internal/httpapi/snapshot.go` (`getInfraNodes`, `getInfraEdges`, `getInfraTopology`) | API-06 + API-07 infra topology endpoint |
| `IMPL-dashboard-01` | `internal/httpapi/snapshot.go` (`getDashboardMetrics`) | API-05 dashboard metrics endpoint |
| `IMPL-realtime-01` | `internal/httpapi/realtime.go` + `internal/httpapi/snapshot.go` 의 critical risks / ci-runs / ci-run-logs handlers | API-14 WebSocket + API-36 `command.status.updated` publisher |
| `IMPL-me-01` | `internal/httpapi/me.go` (`getMe`) | API-32 `GET /api/v1/me` handler |
- **Frontend**: IMPL-frontend-auth-01..06, login-01..03, logout-01, dashboard-01, account-01, admin-01, admin-users-01, admin-org-01, admin-rbac-01, admin-audit-01, org-01, org-comp-01..03, role-01..03, service-auth-01, service-account-01, service-rbac-01, service-audit-01, service-org-01, service-realtime-01, service-api-01, layout-01..02, store-01 (32 항목).
- **Backend AI**: 미사용 (Phase 4 분석에서 placeholder 만 발견).

### 2.5 Unit tests (UT)

- **Backend Go**: UT-httpapi-01..24, rbac-01..03, domain-01..02, auth-01, gitea-01, normalize-01, store-01..03, commandworker-01..02, serviceaction-01, config-01, main-01 (41 파일).
- **Frontend Vitest**: UT-frontend-utils-01, frontend-auth-01..04, frontend-services-01 (6 파일).

### 2.6 E2E TC

- **M2 (28 TC)**: TC-USR-01..06, TC-USR-CRUD-01..03, TC-ACC-01..03, TC-ACC-PROFILE-01, TC-NAV-01..03 + TC-NAV-SIM-01, TC-AUD-01..02, TC-AUTH-NEG-01, TC-AUTH-NOAUTH-01, TC-AUTH-SIGNOUT-REDIR-01, TC-USER-SWITCH-01, TC-RBAC-SUB-01, TC-RBAC-MGR-01, TC-SIGNUP-01..04, TC-PERMISSIONS-SMOKE-01 (`docs/tests/test_cases_m2_auth.md`).
- **M3 (9 TC)**: TC-ORG-LIST-01..02, TC-ORG-UNIT-01..03, TC-ORG-MEM-01..02, TC-ORG-CHART-01..02 (`docs/tests/test_cases_m3_organization.md`).
- **추가** (spec 파일 안에만 정의되어 TC 카탈로그 외): `auth.spec.ts` 의 TC-AUTH-01..06 등 — 향후 TC 카탈로그로 흡수 권장.

## 3. 종합 매트릭스 (도메인 그룹 단위)

| 도메인 | REQ | ARCH / API | ROADMAP | IMPL | UT | TC |
| --- | --- | --- | --- | --- | --- | --- |
| **인증 (auth / OIDC)** | FR-19, 21–24, 65, 67; NFR-3, 18 | ARCH-11, 12; API-19–24, 35 (정밀 매핑: §2.2 Auth API 서브 표) | M1-04, M2-01, 02, 03, 09 | auth-01..07 (책임 정의: §2.4 IMPL-auth-XX 서브 표); frontend-auth-01..06; login-01..03; logout-01 | httpapi-01..04; auth-01; frontend-auth-01..04 | TC-AUTH-NEG-01, NOAUTH-01, SIGNOUT-REDIR-01; TC-USER-SWITCH-01 |
| **회원가입 (signup)** | FR-25, 61–63 | ARCH-12; API-23 (§11.5, 정밀 매핑: §2.2 Auth API 서브 표) | RM-M3-01 (Sign Up + hrdb arm, planned); RM-M3-02 (인사 DB 스키마, planned) | auth-06 (`auth_signup.go`, §2.4 IMPL-auth-XX 서브 표); hrdb-01 (placeholder); frontend-login-03 | httpapi-15 | TC-SIGNUP-01..04 |
| **RBAC** | FR-27, 86; NFR-26 | ARCH-13; API-26–31, 38–40 (정밀 매핑: §2.2 RBAC API 서브 표) | M1-01, 02, 03; M2-11 | rbac-01..04 (책임 정의: §2.4 IMPL-rbac-XX 서브 표); frontend-admin-rbac-01; frontend-service-rbac-01; org-comp-02 | rbac-01..03; domain-01 | TC-RBAC-SUB-01, MGR-01; TC-PERMISSIONS-SMOKE-01 |
| **계정 관리 (account admin + self-service)** | FR-15–18, 20, 22, 23, 26, 61–67; NFR-3, 4, 5, 7, 17, 19, 20 | ARCH-12, 14; API-25, 32, 35 (API-35 = §11.5.1 cross-cut with 인증, 정밀 매핑: §2.2 Auth API 서브 표) | M2-04, 05, 06 | account-01..04 (`account_password.go` 가 API-35 의 IMPL); frontend-account-01; frontend-admin-users-01; frontend-service-account-01 | httpapi-05, 06, 07 | TC-USR-01..06; TC-USR-CRUD-01..03; TC-ACC-01..03; TC-ACC-PROFILE-01 |
| **조직 계층 (organization)** | FR-68–80; NFR-21 | ARCH-15, 16, 17; API-33, 34 (정밀 매핑: §2.2 Account/Org/Me 서브 표, 본문 spec 은 §10.1 의 carve out — 후속 sprint 후보) | M2-07, 08 | org-01..04 (책임 정의: §2.4 IMPL-org-XX 서브 표); frontend-org-01; frontend-admin-org-01; frontend-org-comp-01..03; frontend-service-org-01 | httpapi-07; store-01 | TC-ORG-LIST-01..02; TC-ORG-UNIT-01..03; TC-ORG-MEM-01..02; TC-ORG-CHART-01..02 |
| **감사 (audit)** | FR-18, 26, 102; NFR-4 | ARCH-14; API-18, 39 (39 = §12.9 cross-cut from RBAC, 정밀 매핑: §2.2 Realtime/Command/Audit 서브 표) | M2-15 (in_progress, PR-M2-AUDIT) | audit-01, 02 (책임 정의: §2.4 IMPL-audit-XX 서브 표); kratos-03 | httpapi-19, 24; frontend (Vitest 없음 — §5.2 open) | TC-AUD-01, 02 |
| **명령 lifecycle (command / mitigation / service action)** | FR-58, 59, 84, 95, 100, 101 | API-15–17, 36, 37 (정밀 매핑: §2.2 Realtime/Command/Audit 서브 표) | M1-01..03 (envelope 정합); RM-M4-02 (replay, planned); RM-M4-03 (command status WS UI, planned) | command-01..05 (책임 정의: §2.4 IMPL-command-XX 서브 표); serviceaction-01; realtime-01 | httpapi-09, 13; commandworker-01, 02; serviceaction-01; domain-02 | (E2E 미커버 — gap §5.1, sprint `claude/work_260513-i` 에서 TC-CMD-* 카탈로그 추가) |
| **실시간 (realtime / WebSocket)** | FR-56, 57, 60, 82, 83, 104, 105; NFR-11, 22, 23 | ARCH-05; API-14, 36 (정밀 매핑: §2.2 Realtime/Command/Audit 서브 표) | RM-M4-01 (event publish 확장, planned); RM-M4-02 (replay + 리소스 필터, planned); RM-M4-03 (frontend WS UI, planned) | realtime-01 (책임 정의: §2.4 IMPL-realtime-01); frontend-service-realtime-01 | httpapi-13 | (E2E 미커버 — gap §5.1) |
| **인프라 토폴로지 (infra)** | FR-12, 13, 97, 98, 99; NFR-12, 16 | ARCH-04, 09; API-06, 07 (정밀 매핑: §2.2 Infra/Dashboard 서브 표) | RM-M4-01 (infra event publish, planned) | infra-01 (책임 정의: §2.4 IMPL-infra-XX 서브 표); frontend-role-03 (gardener) | httpapi-12 | (E2E 미커버 — gap §5.1, sprint `claude/work_260513-i` 에서 TC-INFRA-* 카탈로그 추가) |
| **Webhook + 도메인 데이터 (gitea)** | FR-49, 50, 51, 52, 53, 54, 55 | ARCH-06, 07, 08; API-02, 04, 08–13 (정밀 매핑: §2.2 Infra/Dashboard + Pipelines 서브 표) | (M0 이전 phase 완료) | gitea-01, 02; domain-01..03; normalize | httpapi-10, 14; gitea-01; normalize-01; store-03 | (E2E 미커버 — gap §5.1) |
| **대시보드 / 메트릭 / me** | FR-1–11, 28–36, 81, 85, 88, 89, 96 | ARCH-10; API-05, 32, 36, 38 (32 = `GET /api/v1/me`, 정밀 매핑: §2.2 Infra/Dashboard + Account/Org/Me + RBAC 서브 표) | (Phase 4 이전 완료) | dashboard-01; me-01 (책임 정의: §2.4 IMPL-dashboard/me 서브 표); frontend-dashboard-01; frontend-role-01..03; frontend-store-01; frontend-layout-01..02; frontend-service-api-01 | httpapi-08, 11, 22 | TC-NAV-01..03; TC-NAV-SIM-01 |
| **CI / 거버넌스** | NFR-1 (no-docker) | ADR-0001 §5; [ADR-0003](../adr/0003-no-docker-policy-ci-scope.md) | M2-16 (CI 1차, PR #86); FU-CI-1..4 (PR #87); ADR-0003 (PR #88); 거버넌스 + 매트릭스 1차 (PR #89); 갭 분석 + 메타 헤더 표준화 (본 PR / sprint `claude/work_260513-d`) | (build infra: `.github/workflows/ci.yml`, `scripts/ci-setup.sh`, `infra/idp/*.ci.yaml`); `docs/governance/*`, `docs/traceability/*` | (lint 미도입, FU-CI 향후) | (CI run 자체가 검증) |
| **M4 (planned, 정의: §2.3.2)** | FR-37–48, 56–57, 60, 90–94 (일부 — Realtime / AI / Task / Admin) | API-14 (WebSocket 확장) | RM-M4-01..09 (정의: §2.3.2 RM-M4 표) | (미진입 — sprint plan 진입 시 IMPL-ai-XX / IMPL-task-XX 발급) | (미진입) | (미진입) |

> Note — 매트릭스 셀의 ID 는 §2 의 단계별 인덱스를 줄여서 표기 (예: `auth-01..07` = `IMPL-auth-01..IMPL-auth-07`). 한 도메인이 여러 단계에 걸쳐 영향을 주므로 정확한 1:1 매핑은 §2 인덱스 + 단계별 문서 본문의 ID 노출 (`document-standards.md` §5) 로 확인.

## 4. ADR 인덱스

| ADR | 제목 | 상태 | 영향 도메인 |
| --- | --- | --- | --- |
| [ADR-0001](../adr/0001-idp-selection.md) | IdP 선정 (Ory Hydra + Kratos) | accepted (2026-05-07) | 인증, 회원가입, 계정 관리 |
| [ADR-0002](../adr/0002-rbac-policy-edit-api.md) | RBAC policy edit API (DB-backed matrix) | accepted (2026-05-08) | RBAC |
| [ADR-0003](../adr/0003-no-docker-policy-ci-scope.md) | No-Docker 정책 CI 범위 명문화 | accepted (2026-05-13, PR #88) | CI / 거버넌스 |
| [ADR-0004](../adr/0004-x-devhub-actor-removal.md) | `X-Devhub-Actor` 헤더 폐기 완료 선언 | accepted (2026-05-13, sprint `claude/work_260513-h`) | 인증 (auth) |
| [ADR-0005](../adr/0005-workflow-lint-actionlint.md) | GitHub Actions workflow lint (actionlint) CI 잡 도입 | accepted (2026-05-13, sprint `claude/work_260513-i`) | CI / 거버넌스 |
| [ADR-0006](../adr/0006-x-devhub-actor-reject-inbound.md) | inbound `X-Devhub-Actor` 헤더 명시 거부 (400) | accepted (2026-05-13, sprint `claude/work_260513-j`) | 인증 (auth) |
| [ADR-0007](../adr/0007-rbac-cache-multi-instance.md) | RBAC PermissionCache 다중 인스턴스 일관성 (PG LISTEN/NOTIFY, 구현 carve) | accepted (2026-05-13, sprint `claude/work_260513-j`) | RBAC / 운영 |

## 5. Gap 요약

### 5.1 E2E 미커버 도메인

후속 sprint 에서 보완 후보. TC 후보 ID 는 등재만 (실제 spec 작성은 별도).

| 도메인 | 현황 | 가능한 TC 후보 | 우선순위 |
| --- | --- | --- | --- |
| 명령 lifecycle / mitigation | 단위테스트 (`UT-httpapi-09`, `UT-commandworker-01..02`, `UT-domain-02`) 만, e2e UI 흐름 미커버 | TC-CMD-CREATE-01, TC-CMD-STATUS-01 — sprint `claude/work_260513-i` 가 카탈로그 추가, 실제 spec ts 작성은 후속 sprint | P2 (카탈로그 closed, spec ts open) |
| 실시간 (WebSocket) | M3 planned. 현재는 `command.status.updated` 만 publish | M3 진입 시: TC-WS-CONNECT-01, TC-WS-CMD-STATUS-01, TC-WS-RESILIENCE-01 (re-connect) | P3 (M3 의존) |
| 인프라 토폴로지 React Flow | 정적 데이터 렌더만 e2e 미커버 | TC-INFRA-RENDER-01, TC-INFRA-NODE-CLICK-01, TC-INFRA-GROUP-TOGGLE-01 — sprint `claude/work_260513-i` 가 카탈로그 추가, 실제 spec ts 작성은 후속 sprint | P2 (카탈로그 closed, spec ts open) |
| Webhook 처리 (gitea HMAC) | 단위테스트 (`UT-httpapi-14`, `UT-gitea-01`) 로 검증, 외부 영향 e2e 어려움 | E2E 후보 없음 — 통합 테스트 (Go test + 모의 webhook server) 가 자연 | P3 (E2E 외 검증으로 충분) |

### 5.2 ID 부재 / 매핑 누락

| 항목 | 상태 | 처리 |
| --- | --- | --- |
| Backend AI (`backend-ai/`) placeholder | open | M3-04 AnalysisService gRPC client 진입 시 IMPL-ai-XX 발급. |
| Frontend 컴포넌트 Vitest (Header, Sidebar, AuthGuard 등) | open | 후속 sprint 후보 (P2). |
| `auth.spec.ts` 의 TC ID 카탈로그 흡수 | **closed (2026-05-13, sprint `claude/work_260513-d`)** | 재검증 결과 spec 안의 TC-AUTH-NEG-01 + TC-AUTH-NOAUTH-01 모두 `test_cases_m2_auth.md` 의 TC 카탈로그에 이미 흡수되어 있음. 1차 분석에서 "01..06 미흡수" 라고 적은 것은 사실과 다름. |
| API §12 RBAC 정책 편집 API 의 IMPL 정밀 매핑 | **closed (2026-05-13, sprint `claude/work_260513-f`)** | §12.2~§12.10 의 9 endpoint/정책에 `(API-26..31, 38..40)` 본문 ID 노출 + §2.2 의 RBAC API 매핑 표 + §2.4 의 IMPL-rbac-01..04 책임 정의 (handler / store / enforcement / cache) + §3 RBAC 행 IMPL 컬럼 endpoint-IMPL 1:1 매핑. |
| Backend AI (`backend-ai/`) placeholder | open | (위와 동일 — 본 sprint 변경 없음) — M3-04 진입 시 IMPL-ai-XX 발급. |
| Frontend 컴포넌트 Vitest (Header, Sidebar, AuthGuard 등) | **closed (2026-05-13, sprint `claude/work_260513-i`)** | C1 작업 — Header / Sidebar / AuthGuard 단위테스트 추가. `IMPL-frontend-layout-01..02` + `IMPL-frontend-auth-XX` 의 회귀 안전망. |
| 본문 spec 부재 endpoint (`/api/v1/accounts/*`, `/api/v1/users` CRUD, `/api/v1/organization/*`) | **closed (2026-05-13, sprint `claude/work_260513-j`)** | sprint `claude/work_260513-i` 의 §10.1 carve out 표 + sprint `claude/work_260513-j` 가 `backend_api_contract.md` §10.2 (API-25 accounts admin) + §10.3 (API-33 users CRUD) + §10.4 (API-34 organization) 본문 spec 절 신설. endpoint 표 + 권한 + envelope + 1차 에러 매트릭스. 자세한 schema (모든 endpoint 의 request/response/error 매트릭스) 는 후속 spec hygiene 후보. |
| 카탈로그된 TC 가 spec 으로 구현됐는지 역검증 | open | TC-AUD-02 등 일부 TC 가 카탈로그에만 존재할 가능성 — spec 파일 grep 으로 자동 검증할 hygiene 후보. |

### 5.3 문서 ↔ 코드 불일치

| 항목 | 상태 | 처리 |
| --- | --- | --- |
| ADR-0001 vs `frontend_integration_requirements.md` §3.8 | **closed (2026-05-13, sprint `claude/work_260513-d`)** | §3.8 의 "재설계 예정" 노트를 "deprecated" 노트로 명확화 + Phase 13 머지 후 실제 endpoint (API §11.5 / §12.8.2) 로 redirect 링크. |
| X-Devhub-Actor 완전 제거 trigger | **closed (2026-05-13, sprint `claude/work_260513-h` + `-j`)** | [ADR-0004](../adr/0004-x-devhub-actor-removal.md) 가 폐기 완료를 선언 (sprint `-h`). 후속 [ADR-0006](../adr/0006-x-devhub-actor-reject-inbound.md) (sprint `-j`) 이 silent ignore → 400 명시 거부로 전환. `architecture.md` / `ADR-0001` §8 #4 / `me.go` 주석 + `backend_api_contract.md` §8/§9.1/§9.2/§11.3 + `frontend_integration_requirements.md` §3.5 + `environment-setup.md` §2.4 의 spec 잔재 7 위치 정리. 회귀 방지 4 negative 테스트는 ADR-0006 시점에 "ignore → reject" 의도로 갱신. |
| RBAC cache 다중 인스턴스 일관성 | **closed (decision, 2026-05-13, sprint `claude/work_260513-j`)** | [ADR-0007](../adr/0007-rbac-cache-multi-instance.md) 가 PostgreSQL `LISTEN`/`NOTIFY` 채택을 결정. 검토 옵션 A (선택: PG LISTEN/NOTIFY) / B (Redis pub/sub, ADR-0003 충돌로 거부) / C (폴링, latency tradeoff) / D (carve out, 의도 충돌로 거부). **구현은 M3 진입 시 carve out** — M3 sprint plan 의 명시적 항목. M1-DEFER-E 의 decision 부분 closing. |

## 6. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-13 | 1차 작성 (sprint `claude/work_260513-c`). Phase 1–6 분석 결과 통합 + 도메인 그룹 13행 매트릭스 + Gap 요약 §5. |
| 2026-05-13 | 리뷰어 모드 2-pass: §3 의 CI/거버넌스 행을 PR 단위 산출 (PR #86 / #87 / #88 / 본 PR) 로 명세화. §4 ADR-0003 행에 ※ 노트 추가 — 본 PR 매트릭스가 PR #88 미머지 상태와 정합하지 않을 가능성 자체 인지. |
| 2026-05-13 | 후속 sprint `claude/work_260513-d`: ADR-0003 가 main 에 머지된 후 §4 ADR-0003 행을 정상 link 로 활성화 + §3 의 ※ 마킹 제거. §5.1 / §5.2 / §5.3 을 표 형식으로 통일 + 상태(open/closed) 컬럼 도입. §5.2 의 auth.spec.ts TC 미흡수 항목과 §5.3 의 frontend_integration_requirements §3.8 deprecation 항목을 closed 처리. 본 sprint 의 메타 헤더 표준화 commit 도 §3 CI/거버넌스 행에 추가. |
| 2026-05-13 | sprint `claude/work_260513-e` (A 묶음, M1 PR-D 정합 마무리): backend-core 의 `writeRBACServerError` → `writeServerError` 통합 (`internal/httpapi/rbac.go` helper 제거 + 11 호출 일괄 치환). `requireRequestID` 미들웨어에 caller-supplied X-Request-ID validation 추가 — 정규식 `^[A-Za-z0-9_-]{1,128}$`, 실패 시 server-generated fallback (work_260512-j 발견 항목 closed). request_id 를 표준 ctx key (`requestIDCtxKey{}`) 에도 stash + `requestIDFromContext` / `logRequestCtx` ctx-aware helper 추가, `kratos_login_client.go` 의 untraced `log.Printf` 2건 + `kratos_identity_resolver.go` 1건을 ctx-aware 로 치환 (logRequest 의 untraced fallback 해소 항목 closed). 단위테스트 11건 추가 (validation 양/음 경로 + ctx 전파 + logRequestCtx percent-safety + 미들웨어 e2e). PR #91 (Pass 1 review 에서 `writeAuthLoginServerError` 도 같은 wrapper 발견 + 보강 commit 으로 흡수). |
| 2026-05-13 | sprint `claude/work_260513-f` (B 묶음, RBAC 1차): `backend_api_contract.md` §12.2~§12.10 의 9 헤더에 `(API-26..31, 38..40)` 본문 ID 노출 (`document-standards.md` §8 우선순위 3 RBAC 도메인 1차 적용). 본 §2.2 에 RBAC API 매핑 표 + 본 §2.4 에 IMPL-rbac-01..04 책임 정의 (handler / store / enforcement / cache) 서브 표 추가. §3 RBAC 행 IMPL 컬럼 endpoint-IMPL 1:1 매핑. §5.2 의 "RBAC API §12 IMPL 정밀 매핑" 항목 closed. Pass 1 review 보강으로 §3 RBAC 행을 ID 범위 + §2 서브 표 참조 패턴으로 정리 (PR #92). |
| 2026-05-13 | sprint `claude/work_260513-g` (B1 auth 도메인 2차): `backend_api_contract.md` §11.3 `(API-19)` + §11.5 표에 API ID 컬럼 (`API-20..24, 35`) + §11.5.1 `(API-35)` 본문 ID 노출. 본 §2.2 에 Auth API 매핑 표 + 본 §2.4 에 IMPL-auth-01..07 책임 정의 (verifier / actor / 5 endpoint handler) 서브 표 추가. §3 인증 / 회원가입 / 계정 관리 행에 §2 서브 표 참조 노트 — 회원가입과 계정 관리 행은 cross-cut (API-23 / API-35) 명시. §11.2 외부 의존성 + §11.4 planned 는 매핑 제외 (`conventions.md` §5.2). |
| 2026-05-13 | sprint `claude/work_260513-h` (B4: X-Devhub-Actor 폐기 ADR): [ADR-0004](../adr/0004-x-devhub-actor-removal.md) 발급. M0 SEC-4 + M1 PR-D Bearer token verifier 도입으로 ADR-0001 §8 #4 trigger 가 이미 충족됐음을 ex-post-facto 명문화. §4 ADR 인덱스에 ADR-0004 행 추가. §5.3 "X-Devhub-Actor 완전 제거 trigger" closed. `docs/architecture.md` line 174 + `docs/adr/0001-idp-selection.md` §8 #4 인라인 갱신 + `backend-core/internal/httpapi/me.go` line 16 주석 잔재 정리. 회귀 방지 negative 테스트 4 파일 (audit_test / commands_test / auth_test / me_test) 그대로 유지. Pass 1 review 보강으로 `backend_api_contract.md` §8/§9.1/§9.2/§11.3 + `frontend_integration_requirements.md` §3.5 + `environment-setup.md` §2.4 의 spec 잔재 6 위치도 함께 정리 + ADR-0004 §5 정정. |
| 2026-05-13 | sprint `claude/work_260513-i` (대형 묶음 B1~D5): B1 추가 5 도메인 (account / org / command / audit / infra) 본문 ID 노출 + §10.1 "본문 spec 부재 endpoint" carve out 표 + §2.2 4 신규 서브 표 (Infra/Dashboard, Pipelines, Realtime/Command/Audit, Account/Org/Me) + §2.4 6 신규 IMPL 서브 표 (account / org / command / audit / infra+dashboard+realtime+me). §3 6 행 (감사 / 명령 / 실시간 / 인프라 / Webhook+gitea / 대시보드+me + 조직) 갱신 (ID 범위 + §2 서브 표 참조). §5.1 의 명령/인프라 행은 "카탈로그 closed, spec ts open" 으로 갱신. §5.2 frontend Vitest 부재 closed (C1 1차 도입) + 본문 spec 부재 endpoint open (account / users / org). B2 archive/AGENTS.md + archive/split_checklist.md deprecated 마킹. C1 ThemeToggle Vitest 3 tests. C2 test_cases_m3_command_infra.md 신규 (TC-CMD-* / TC-INFRA-* 5 TC 카탈로그, spec ts 작성은 carve out). D5 [ADR-0005](../adr/0005-workflow-lint-actionlint.md) 발급 + `.github/workflows/ci.yml` 에 `workflow-lint` 잡 추가 (raven-actions/actionlint@v2). |
| 2026-05-13 | sprint `claude/work_260513-j` (M3 진입 전 잔여 후속 일괄): D6 inbound X-Devhub-Actor 거부 (auth.go) + [ADR-0006](../adr/0006-x-devhub-actor-reject-inbound.md) — silent ignore → 400 + `code=x_devhub_actor_removed`. 회귀 4 테스트 의도 갱신 (`me_test`, `audit_test`, `auth_test`, `commands_test`). [ADR-0007](../adr/0007-rbac-cache-multi-instance.md) RBAC cache 다중 인스턴스 일관성 결정 (PG LISTEN/NOTIFY 채택, 구현은 M3 carve out). B2-2 deprecated 추가 마킹 (backend/requirements_review.md / DOCUMENT_INDEX.md / assessment.md). 매트릭스 §2.2 nit 정정 (API-XX ID 공간 + composite + 결손 명시). `backend_api_contract.md` §10.2 (API-25 accounts admin) + §10.3 (API-33 users CRUD) + §10.4 (API-34 organization) 본문 spec 절 신설 — §5.2 "본문 spec 부재 endpoint" closed. AuthGuard smoke Vitest 1 test (loading 상태, 나머지 mock-heavy 본격은 M3+). TC-INFRA-RENDER-01 spec ts (정적 렌더 검증, 인터랙션 TC 는 carve). §4 ADR 인덱스에 ADR-0006 + ADR-0007 행 추가. §5.3 의 X-Devhub-Actor + RBAC cache 항목 closed 갱신. |
| 2026-05-13 | sprint `claude/work_260513-k` (M3/M4 drift 정합): M3/M4 정의의 3 source 사이 drift 해소. `docs/development_roadmap.md` §3 를 single source-of-truth 로 명시 (이전 매트릭스 RM-M3 정의가 development_roadmap M4 항목을 cross-cut 한 drift 상태). development_roadmap §3 의 M3 헤더 중복 + 본문 정리 — M3 잔여 = Sign Up + 인사 DB + 조직 polish (사용자/조직 관리 대부분은 M2 1차 완성 sprint PR #85 가 흡수). 매트릭스 §2.3 갱신 + §2.3.1 RM-M3 정의 표 (3 항목: Sign Up, 인사 DB, 조직 polish) + §2.3.2 RM-M4 정의 표 (9 항목 신규 발급: WebSocket 확장/replay, command status WS UI, AI Gardener, Suggestion Feed, Gitea Pull worker, System Admin 대시보드, RBAC cache LISTEN/NOTIFY, 외부 SSO). 매트릭스 §3 의 회원가입 / 명령 lifecycle / 실시간 / 인프라 행 RM-M3 → RM-M4 정합. M4 row 정의 갱신 (스코프 밖 → planned, 정의: §2.3.2). |
