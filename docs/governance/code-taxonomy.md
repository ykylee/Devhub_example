# 코드베이스 카테고리 · 모듈 분류 (Code Taxonomy)

- 문서 목적: DevHub Example 의 backend / frontend / infra / docs 자산을 도메인 단위 카테고리 + 카테고리 내 상세 모듈로 분류해 운영·리팩토링·신규 작업 시 일관된 참조점을 제공한다.
- 범위: backend-core / backend-ai / frontend / infra / scripts / docs / .github/workflows
- 분석 기준 커밋: main HEAD `101f4109` (2026-05-28)
- 분석 방법: Explore agent 4개 병렬 (backend / frontend / infra / docs) 결과 종합 + claude 통합
- 상태: active (신규)
- 관련 문서: [`worker_division.md`](./worker_division.md), [`document-standards.md`](./document-standards.md), [`traceability/report.md`](../traceability/report.md)

## 0. 컨벤션 — 작업 명시 형식

**모든 신규 작업 (PR, commit, backlog row) 은 영향 카테고리/모듈을 명시**한다. 충돌·중복·dead code 발견을 사전 차단하기 위함.

권장 형식:

- PR title: `[<카테고리>/<모듈>] <변경 한 줄>` — 다중 카테고리는 `[A/m1, B/m2]` 또는 prefix 생략 가능
- commit message 첫 줄: 동일 prefix 또는 conventional commit 의 scope 자리에 `<카테고리>` 사용 — 예: `feat(application-lifecycle): ...`
- traceability row: 변경 카테고리/모듈 명시 (REQ/UC/ARCH/API/IMPL ID 와 함께)
- backlog row: `[<카테고리>/<모듈>]` prefix

본 문서는 카테고리 이름 + 모듈 이름의 **단일 출처 (SoT)**. 새 도메인 추가 시 본 문서 §1 (top-level 카테고리) + §2 (모듈) + §3 (리팩토링 후보) 갱신.

## 1. Top-level 카테고리 (12개)

코드베이스는 **도메인 관심사 기준 12개 카테고리**로 분류된다. 카테고리 명은 영문 kebab-case (작업 prefix 일관성), 한글 설명 병기.

| # | 카테고리 | 한글 | 책임 (1줄) |
|---|---|---|---|
| 1 | `auth-session` | 인증·세션 | Keycloak OIDC + JWT 검증 + 토큰/세션 lifecycle |
| 2 | `audit-ops` | 감사·운영 | Keycloak event polling + audit log + 메트릭 |
| 3 | `rbac-permissions` | RBAC·권한 | 역할/리소스/액션 매트릭스 + 라우트 가드 |
| 4 | `organization-management` | 조직·사용자 | 부서 트리 + 임원진 + 사용자 + HR DB 어댑터 |
| 5 | `onboarding` | 온보딩 | 신규 사용자 first-login + gate + feature flag |
| 6 | `application-lifecycle` | 애플리케이션·프로젝트 | App/Project CRUD + 상태 머신 + 대시보드 + rollup |
| 7 | `repository-integration` | 저장소·SCM 연동 | SCM provider + repository sync + draft/publish |
| 8 | `dev-request` | 개발 의뢰 (DREQ) | 외부 intake + token + 상태 머신 + promote |
| 9 | `integration-registry` | 외부 통합 | Integration provider + binding + adapter (Gitea/Jira/...) + task ingestion |
| 10 | `realtime` | 실시간 통신 | WebSocket hub + ticket store (멀티 인스턴스) |
| 11 | `infrastructure-monitoring` | 인프라 모니터링 | Topology + risk + command worker + service action |
| 12 | `ui-foundation` | UI 공통 기반 | 공통 컴포넌트 + 레이아웃 + 토스트 + 모달 + 라우트 가드 |

추가 횡단 카테고리 (코드 외):

| # | 카테고리 | 책임 |
|---|---|---|
| X-1 | `deployment-automation` | scripts + docker-compose + nginx template + build-artifacts |
| X-2 | `ci-pipeline` | .github/workflows + actionlint + migration prefix lint |
| X-3 | `database-migration` | backend-core/migrations + golang-migrate + seed SQL |
| X-4 | `governance-docs` | docs/adr + governance + traceability + planning + setup |

## 2. 카테고리 별 모듈 매핑

각 카테고리는 backend / frontend / infra / docs 영역 별 핵심 모듈을 명시. 의존성 + lifecycle 상태도 함께.

### 2.1 `auth-session` — 인증·세션

**Backend**:
- `auth/keycloak_verifier` — JWKS 검증, stale-while-error fallback (ADR-0020 sub-carve D)
- `auth/metrics` — JWKS 메트릭
- `httpapi/auth`, `httpapi/identity_resolver`, `httpapi/keycloak_admin_client` — `BearerTokenVerifier` 계약 + Keycloak Admin API
- `httpapi/me`, `httpapi/me_onboarding` — `/api/v1/me` 응답 + onboarding flag

**Frontend**:
- `app/login`, `app/auth/callback`, `app/auth/logout` — OIDC 플로우 페이지
- `lib/auth/{token-store, refresh, refresh-scheduler, session-death, pkce, role-routing}` — 클라이언트 토큰 관리
- `lib/services/auth.service.ts` — OIDC client
- `components/layout/AuthGuard.tsx` — 페이지 진입 가드

**관련 ADR**: 0001 (deprecated → 0019), 0006, 0019, 0020, 0024
**E2E spec**: `auth.spec.ts`, `signout.spec.ts`, `onboarding-first-login.spec.ts`
**의존**: `rbac-permissions`, `audit-ops`

### 2.2 `audit-ops` — 감사·운영

**Backend**:
- `audit/{keycloak_event_puller, keycloak_admin_adapter, user_sync, metrics}` — Keycloak event polling + audit log 수집
- `httpapi/audit` — API-11 audit log list
- `httpapi/keycloak_events_webhook` — Keycloak SPI webhook 수신
- `store/audit_logs` — `audit_logs` + `event_cursors` (dedup)

**DB**: `audit_logs` (000003), `event_cursors` (000031)
**관련 ADR**: 0020 sub-carve E (event listener SPI)
**의존**: `auth-session` (actor 식별)

### 2.3 `rbac-permissions` — RBAC·권한

**Backend**:
- `httpapi/permissions` — route → resource/action 매핑 + `PermissionCache`
- `httpapi/rbac` — API-62 RBAC CRUD
- `httpapi/authz` — route policy 정의
- `store/postgres_rbac` — `rbac_policies` 쿼리
- `domain/rbac` — 상수 + 역할 enum

**Frontend**:
- `app/(dashboard)/admin/settings/permissions/page.tsx`
- `lib/services/rbac.service.ts`
- `components/organization/{PermissionEditor, PermissionMatrix}`

**DB**: `rbac_policies` (000005, 000018, 000021, 000024, 000026)
**관련 ADR**: 0002, 0007, 0011 (row-scoping)
**E2E**: `admin-permissions.spec.ts`, `rbac-routes.spec.ts`, `header-switch-view.spec.ts`

### 2.4 `organization-management` — 조직·사용자

**Backend**:
- `httpapi/organization` (1019 LoC — 분할 후보), `httpapi/organizations_search`, `httpapi/hr_lookup`
- `store/users_units` (1263 LoC — 분할 후보) — 사용자↔부서 지정
- `hrdb/postgres`, `hrdb/mock` — HR DB 어댑터
- `domain/primary_unit` — 부서 모델

**Frontend**:
- `app/(dashboard)/admin/settings/{organization, users}`
- `lib/services/identity.service.ts` (504 LoC — 분할 후보)
- `components/organization/{OrgTree, OrgNode, OrgUnitTable, OrgUnitGrid, MemberTable, *Modal}` (다수)

**DB**: `users`, `org_units`, `unit_appointments` (000004, 000019), `hrdb_persons` (000010), `org_units_total_count_mv` (000011)
**관련 ADR**: 0008 (HR adapter), 0009 (MV), 0010 (primary dept)
**E2E**: `admin-org-crud.spec.ts`, `admin-users-crud.spec.ts`, `admin-users-search.spec.ts`

### 2.5 `onboarding` — 온보딩

**Backend**:
- `httpapi/onboarding_{gate, feature_flag, roles, pending_gauge, metrics}` — gate middleware + flag + 메트릭
- `httpapi/me_onboarding` — POST API-83
- `users.onboarding_completed_at` 컬럼 (000033)

**Frontend**:
- `app/(dashboard)/onboarding/page.tsx`

**관련 ADR**: 0021
**E2E**: `onboarding-first-login.spec.ts`
**의존**: `auth-session`, `organization-management`

### 2.6 `application-lifecycle` — 애플리케이션·프로젝트

**Backend**:
- `httpapi/applications` (1066 LoC — 분할 후보), `httpapi/projects` (883 LoC), `httpapi/application_rollup`
- `store/applications` (1172 LoC — 분할 후보), `store/repository_ops` (618 LoC)
- `domain/application` (469 LoC)

**Frontend**:
- `app/(dashboard)/applications/{page, [id]/page}` (612 LoC — 분할 후보), `app/(dashboard)/projects/{page, [id]/page}` (612 LoC), `app/(dashboard)/admin/catalog/page.tsx` (587 LoC — 거대)
- `lib/services/{application.service, project.service, project.types}`
- `lib/utils/{lifecycle-status, last-build}` (헬퍼)
- `components/project/{ApplicationCreationModal, ProjectCreationModal, ApplicationTable, ProjectTable, RepositoryCreationModal, RepositoryLinkModal, RepositoryTable}`

**DB**: `applications` (000013), `application_repositories` (000014), `projects` (000015), `project_members`, `project_repositories` (000034), `repo_ops_snapshots` (000017)
**관련 ADR**: 0011 (row-scoping), 0014 (DREQ promote tx)
**E2E**: `admin-applications.spec.ts`, `admin-projects.spec.ts`, `admin-project-model-v2.spec.ts`, `applications-projects-detail-negative.spec.ts`

### 2.7 `repository-integration` — 저장소·SCM 연동

**Backend**:
- `httpapi/integration_scm_repositories` (338 LoC) — API-88/89 import/sync
- `httpapi/domain` (저장소 list/draft/publish) — API-08, 신규 draft→publish
- `store/postgres.go` repository 부분 + `store/applications.go` Get/Upsert/ListRepositories
- `gitea/*` — Gitea API client + webhook + sync worker
- `normalize/gitea` — Gitea JSON → domain 정규화

**Frontend**:
- `app/(dashboard)/repositories/{page, [id]/page}`
- `lib/services/{repository.service, integration.service, integration-provider-presets, integration.types}`
- `components/integration/{ProviderModal (663 LoC), ImportRepositoriesModal, CreateScmRepositoryModal, BindingsTable, ProviderTable, ...}`

**DB**: `repositories` (000002 + 000042+000043+000045), `scm_providers` (000012 deprecated → `integration_providers` 000028 통합), `webhook_events` (000001)
**관련 ADR**: 0024 (WS auth ticket — 일부 영향)
**E2E**: `admin-integrations.spec.ts`, `admin-integration-bindings.spec.ts`, `repositories-ui.spec.ts`, `repositories-ui-negative.spec.ts`, `repositories-detail-negative.spec.ts`

### 2.8 `dev-request` — 개발 의뢰 (DREQ)

**Backend**:
- `httpapi/dev_requests` (920 LoC — 분할 후보), `httpapi/dev_request_intake_auth`, `httpapi/dev_request_intake_tokens_admin`
- `store/{dev_requests, dev_requests_promote, dev_request_intake_tokens}`
- `devrequest/{metrics, intake_token_cron}`

**Frontend**:
- `app/(dashboard)/dev-requests`, `app/(dashboard)/admin/settings/{dev-requests, dev-request-tokens}`
- `lib/services/{dev_request.service, dev_request_token.service}`
- `components/dev-request/{DevRequestTable, DevRequestDetailModal, IntakeTokenTable, IssueIntakeTokenModal, EditIntakeTokenModal, MyPendingDevRequestsWidget}`

**DB**: `dev_requests` (000022), `dev_request_intake_tokens` (000023)
**관련 ADR**: 0012, 0013, 0014, 0017
**E2E**: `dev-requests.spec.ts`

### 2.9 `integration-registry` — 외부 통합

**Backend**:
- `httpapi/integration_registry` (857 LoC — 분할 후보), `httpapi/integrations`, `httpapi/infra_integrations`, `httpapi/external_task_handler`
- `store/{integration_registry, integrations, external_task_store}`
- `integrations/adapters/{homelab_http, homelab_file, task_item_puller}`

**Frontend**:
- (UI 진입점은 `repository-integration` 의 `components/integration/*` 와 공유 — Integration provider modal/table 이 SCM + 비-SCM 양쪽 다룸)
- `app/(dashboard)/admin/settings/integrations`

**DB**: `integration_providers`, `integration_bindings`, `integration_sync_jobs` (000028, 000040, 000041), `project_integrations` (000016 legacy), `infra_service_snapshots` (000029), `external_task_items` (000046)
**관련 ADR**: 0015 (homelab pull)

### 2.10 `realtime` — 실시간 통신

**Backend**:
- `httpapi/realtime` (WS hub), `httpapi/realtime_ticket` (ticket store — in-memory vs PG)
- `store/realtime_tickets`

**Frontend**:
- `lib/services/{realtime.service (300 LoC), websocket.service}`

**DB**: `realtime_tickets` (000035)
**관련 ADR**: 0024 (ticket pattern + ws auth)
**의존**: `auth-session` (ticket 발급 시 actor 식별), 모든 dashboard 카테고리가 subscribe 측 dep

### 2.11 `infrastructure-monitoring` — 인프라 모니터링

**Backend**:
- `httpapi/{commands, snapshot, snapshot_provider, runtime_snapshot_provider, events}` — 명령 + 스냅샷
- `commandworker/{worker, live_worker}` — 명령 polling + 실시간 실행
- `serviceaction/executor` (306 LoC) — sandbox/production 모드

**Frontend**:
- `app/(dashboard)/admin/topology-v2/page.tsx` (504 LoC — 분할 후보)
- `lib/services/{infra.service, risk.service}`

**DB**: `commands` (000003), `risks` (000002), `infra_service_snapshots` (000029)
**관련 ADR**: 0016 (Prometheus alerts)
**E2E**: `admin-topology-v2.spec.ts`

### 2.12 `ui-foundation` — UI 공통 기반

**Frontend** (전부 frontend 단일):
- `components/ui/{Modal, Badge, Toast, PageState, DashboardHeader, FilterBar, ActionMenu, ComboBox, DestructiveConfirmModal, LogoutOverlay}`
- `components/layout/{Header (369 LoC), Sidebar (232 LoC), AuthGuard}`
- `lib/utils/{cn, lifecycle-status, last-build}` — 도메인 공유 helper
- `app/globals.css`, `next.config.ts`

**의존**: 없음 (모든 다른 frontend 카테고리가 의존)

### X-1 `deployment-automation` — 배포 자동화

- `scripts/{build-artifacts.sh, deploy-from-env.sh, deploy-up.sh, deploy-preflight.sh, setup-keycloak.sh, verify-keycloak-groups.sh, nginx-conf-sync.sh, setup-test-db.sh, ci-e2e-sync-check.sh}`
- `infra/nginx/{devhub.deploy.conf.template, devhub.deploy.conf, devhub.native.conf, default.conf}`
- `docker-compose.deploy.yml` (있다면 — git 추적 외 환경별 가능)

**관련 ADR**: 0003, 0018, 0019, 0022, 0023
**Docs**: `docs/setup/{environment-setup, docker-packaging-deployment-guide, deploy_preflight_checklist, single_port_deployment, internal_network_constraints}`

### X-2 `ci-pipeline` — CI/CD

- `.github/workflows/{ci.yml, docker-image-publish.yml}`
- `scripts/ci-e2e-sync-check.sh`

**Jobs**: `changed-paths`, `workflow-lint`, `migration-prefix-lint`, `backend-unit`, `backend-integration`, `frontend-unit`, `e2e (shard 1/2)`

### X-3 `database-migration` — 마이그레이션

- `backend-core/migrations/` (46개 up/down pair, 000001 ~ 000046)
- `infra/idp/sql/{001_create_idp_schemas, 002_seed_e2e_users, 003_seed_test_admin}.sql`

**관련 도메인**: 모든 backend 카테고리가 자기 테이블 소유 (§2.1~§2.11 의 DB 항목 참조)

### X-4 `governance-docs` — 거버넌스·문서

- `docs/adr/` (24 ADR, immutable)
- `docs/governance/{README, document-standards, worker_division, keycloak_admin_responsibility, code-taxonomy(본 문서)}`
- `docs/traceability/{README, conventions, sync-checklist, report, traceability_remediation_plan_auth_org}`
- `docs/planning/` (33 문서)
- `docs/setup/` (14 문서)
- `docs/tests/` (TC 카탈로그 + 보고서)
- `docs/analysis/` (시점별 코드/도큐 분석 snapshot)
- `docs/wiki/`, `docs/README.md`, `docs/PROJECT_PROFILE.md`
- top-level: `docs/{requirements, architecture, backend_api_contract}.md`

## 3. 리팩토링 후보 (P0 ~ P3)

본 분류 작업에서 식별된 리팩토링 후보. 우선순위 + 카테고리 + 영향 모듈 명시.

### P0 — 즉시 (operational risk 또는 명확한 quick win)

| # | 카테고리/모듈 | 항목 | 효과 |
|---|---|---|---|
| P0-1 | `application-lifecycle` / `store/applications` | 거대 파일 분할 (1172 LoC) — applications + projects + repositories 혼재 | 유지보수성·테스트 용이성 |
| P0-2 | `application-lifecycle` / `httpapi/applications` | 거대 파일 분할 (1066 LoC) | 동일 |
| P0-3 | `repository-integration` + `integration-registry` / `httpapi/integration_registry` | 거대 파일 분할 (857 LoC, provider+binding+sync 혼재) | 동일 |
| P0-4 | `dev-request` / `httpapi/dev_requests` | 거대 파일 분할 (920 LoC) — handler + registration payload + 상태 머신 | 동일 |
| P0-5 | `organization-management` / `httpapi/organization` + `store/users_units` | 거대 파일 분할 (1019 + 1263 LoC) | 동일 |
| P0-6 | `auth-session` + `ui-foundation` / `components/integration/ProviderModal` | 거대 모달 (663 LoC) — `FormModal` 베이스로 추출 | 코드 재사용 |
| P0-7 | `deployment-automation` / `hrdb_etl_sync.sh` | deprecated 스크립트 명시 — `_archive_*.obsolete` 로 rename 또는 README 명시 | 운영 사고 방지 |
| P0-8 | `governance-docs` / `DOCUMENT_INDEX.md` | deprecated 배너 강화 (이미 deprecated) | 혼동 회피 |

### P1 — 1~2주 (구조 개선)

| # | 카테고리/모듈 | 항목 |
|---|---|---|
| P1-1 | `application-lifecycle` / `app/(dashboard)/applications/[id]/page.tsx` (550) + `projects/[id]/page.tsx` (612) + `admin/catalog/page.tsx` (587) + `topology-v2/page.tsx` (504) | 거대 페이지 분할 — sub-component 추출 |
| P1-2 | `application-lifecycle` + `repository-integration` / fetch 중복 | React Query/SWR 또는 Zustand store 도입으로 endpoint 중복 호출 제거 |
| P1-3 | `organization-management` / `identity.service.ts` (504 LoC) | OrgStructure/UserManagement/Permission 3개로 분리 |
| P1-4 | `auth-session` / `app/(dashboard)/admin/topology-v2/page.tsx` | React Flow 그래프 + 노드 상태 + 모달 분리 |
| P1-5 | `integration-registry` / 어댑터 패턴 표준화 | homelab 외 다른 provider 도 동일 인터페이스 따르도록 |
| P1-6 | `governance-docs` / `backend_api_contract.md` (2881 LoC) | 도메인별 분리 (`api/auth.md`, `api/application.md`, ...) |

### P2 — 정합성 + 가드

| # | 카테고리/모듈 | 항목 |
|---|---|---|
| P2-1 | (전 backend) | `httpapi` 패키지 → `handler` (HTTP 만) + `service` (비즈니스 로직) 분리 |
| P2-2 | (전 frontend) | error handling 표준화 — `ErrorBoundary` + `error-message.ts` 확장 |
| P2-3 | `ci-pipeline` / migration-prefix-lint | prefix + name slug 동시 중복 감지 (현재 prefix 만) |
| P2-4 | `auth-session` + `governance-docs` | Keycloak realm dev/prod drift 방지 — base + override 구조 |
| P2-5 | `governance-docs` / ADR | supersession chain 의 deprecation 배너 표준화 |
| P2-6 | `application-lifecycle` + `repository-integration` | `<Link>` prefetch 패턴 + idle watchdog (router idle silent fail 후속 — #405 §6.2/6.3) |

### P3 — 옵션

| # | 카테고리/모듈 | 항목 |
|---|---|---|
| P3-1 | `governance-docs` / planning | concept/design/usecase/api 4단계 의존성 명시 + carve-out 노트 강화 |
| P3-2 | `governance-docs` / `code-taxonomy.md` (본 문서) | 신규 카테고리 추가 SOP 명시 |
| P3-3 | `database-migration` / naming | 마이그레이션 이름 컨벤션 lint (action verb + entity name + obsolete naming 차단) |

## 4. 작업 prefix 컨벤션 적용 SOP

신규 작업 시:

1. **카테고리 결정** — 영향 받는 카테고리 1~3개 식별 (다중 카테고리면 primary 1개 명시 + others 표기)
2. **모듈 명시** — 카테고리 내 어느 모듈인지 (§2 의 모듈 명 사용)
3. **PR title**: `<conventional commit type>(<카테고리>): <한 줄 요약>` — 예: `feat(application-lifecycle): app/project status 자유화`
4. **commit message** 첫 줄도 동일
5. **traceability row** 의 sprint 명에 카테고리 prefix 또는 sprint name 자체에 카테고리 포함
6. **PR body** 에 `## 카테고리` 섹션 명시 — 영향 모듈 list

## 5. 본 문서 갱신 정책

- 신규 카테고리 추가 / 모듈 신설: 본 문서 §1 + §2 갱신 PR
- 리팩토링 항목 완료/추가: §3 갱신
- 컨벤션 변경: §0 + §4 갱신 + `worker_division.md` 와 cross-link
- 분석 snapshot 발행 시: `docs/analysis/<date>-codebase-snapshot/` 결과를 본 문서에 반영 (또는 본 문서가 master, snapshot 이 timestamped diff)

## 6. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-28 | 초안 — Explore agent 4개 (backend/frontend/infra/docs) 병렬 결과 종합 + 12 카테고리 + 4 횡단 카테고리 + 리팩토링 후보 P0~P3 도출. main HEAD `101f4109` 기준. |
