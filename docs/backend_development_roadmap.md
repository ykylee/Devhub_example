# 백엔드 개발 로드맵

> ⚠ **먼저 [통합 개발 로드맵](./development_roadmap.md)을 확인하세요.** 본 문서는 그 통합 로드맵의 Backend 트랙 세부입니다. 마일스톤(M0~M4) / 우선순위(P0~P3) / 트랙 간 의존은 통합 로드맵의 §3·§4 가 source-of-truth.
>
> ⚠ **v1.0/v1.1 신규 작업의 source-of-truth = [`docs/planning/release_v1_roadmap.md`](./planning/release_v1_roadmap.md).** 본 문서는 backend phase 이력 + 잔여 추적용이며, 우선순위·마일스톤·잔여 carve 의 최신 기준은 release_v1_roadmap 이다.
>
> ⚠ **2026-05-29 정합 (SDLC 재정비 sprint #408~#416)**: backend 코드는 [`docs/governance/code-taxonomy.md`](./governance/code-taxonomy.md) §2.1 의 **`backend-core/internal/domain/<도메인>/{view,service,repository,schema}` 4 계층** 으로 재정렬됨. PR #409 가 `providerHasCapability` 3 카피 → `internal/shared/integrationcaps/` 통합 (11 unit test). 현행 코드베이스 전수 분석은 [2026-05-27 codebase snapshot](./analysis/2026-05-27-codebase-snapshot/04_backend_summary.md) 참조.

- 문서 목적: DevHub 백엔드 구현 범위, 순서, 진척 상태를 추적한다.
- 범위: backend-core phase 로드맵, 완료 범위, 다음 작업 큐, 차단 항목, `backend-core/internal/domain/<도메인>/` 4 계층 매핑.
- 대상 독자: 백엔드 개발자, 프론트엔드 연동 담당자, AI agent
- 기준일: 2026-05-07
- 상태: in_progress
- 최종 수정일: 2026-05-29 (SDLC 재정비 sprint #408~#416 — backend domain 4 계층 매핑 + Shared `integrationcaps` 도입 (PR #409) + Infrastructure 진입점 명시 + 후속 carve out (PlatformRepository decouple / ApplicationStore slim))
- 관련 문서: [`docs/development_roadmap.md`](./development_roadmap.md) (통합), [`docs/planning/release_v1_roadmap.md`](./planning/release_v1_roadmap.md) (v1.0/v1.1 source-of-truth), [`docs/governance/code-taxonomy.md`](./governance/code-taxonomy.md) (SoT — 10 도메인 + 4 계층), [`docs/domain/`](./domain/README.md) (도메인 SDLC 진입점), [`docs/shared/`](./shared/README.md) (Shared 진입점 — integrationcaps 포함), [`docs/infrastructure/`](./infrastructure/README.md) (Infrastructure 진입점), [`docs/analysis/2026-05-27-codebase-snapshot/04_backend_summary.md`](./analysis/2026-05-27-codebase-snapshot/04_backend_summary.md) (현행 backend 전수 분석), `docs/requirements.md`, `docs/architecture.md`, `docs/shared/tech_stack.md`, `docs/backend_api_contract.md`, [`docs/adr/0019-keycloak-only-idp.md`](./adr/0019-keycloak-only-idp.md) (current IdP), [`docs/adr/0001-idp-selection.md`](./adr/0001-idp-selection.md) (Hydra+Kratos, **superseded** by ADR-0019), [`docs/adr/0024-websocket-auth-query-token.md`](./adr/0024-websocket-auth-query-token.md) (WS ticket 인증)
- 현재 브랜치: `main`
- 현재 기준선: main `273d9d4` (PR #415 머지 후, sprint `claude/work_260529-k` 진입 기준). **Keycloak 단일 IdP 전환 완료(ADR-0019 — Hydra/Kratos 전면 제거).** Application·Repository·Project / DREQ / External Integration / Onboarding / Gitea SCM sync worker / Repository draft→publish + SCM 양방향 도메인 모두 1차 완성.

## 1. 개발 원칙

- Go Core는 Gitea Webhook/API 연동, 권한 관리, 데이터 저장, 시스템 관리자 기능의 중심 서비스로 둔다.
- Python AI는 초기 단계에서 PostgreSQL에 직접 접근하지 않고, Go Core가 필터링한 입력을 gRPC로 전달받는다.
- 프론트 대상 실시간 API는 gRPC stream이 아니라 REST snapshot + WebSocket event로만 계약한다.
- 프론트엔드 UI 표시명과 API wire format은 분리한다. 역할 값은 API에서 `developer`, `manager`, `system_admin`을 기본으로 한다.
- 역할별 UX 제공은 프론트의 기본 진입 우선순위로 간접 제공하며, 백엔드는 role wire format과 권한 검사 일관성을 보장한다.
- 명령성 액션은 boolean 결과가 아니라 `command_id`와 command status lifecycle로 관리하고 audit log를 남긴다.
- 운영 actor는 `X-Devhub-Actor` header(폐기)가 아니라 **Keycloak OIDC 토큰의 JWT claim**에서 도출한다(JWKS 검증 + stale-while-error fallback, `internal/auth/keycloak_verifier.go`). ADR-0004/0006 으로 `X-Devhub-Actor` inbound 는 명시 거부.
- 조직/사용자 도메인의 master data는 DevHub `users`/`org_units`가 담당하고, **credential/session master는 Keycloak(단일 IdP, ADR-0019)**가 담당한다. (이전 Hydra/Kratos 전제는 ADR-0019 로 폐기 — historical.)
- 검증하지 않은 단계는 `done`으로 전환하지 않는다.
- 세션 상태 문서는 브랜치별 memory 경로(`ai-workflow/memory/<agent>/<branch>/`)를 source of truth로 사용한다.

## 1.1 도메인 4 계층 매핑 (2026-05-29 SDLC 재정비 정합)

backend 코드는 [`docs/governance/code-taxonomy.md`](./governance/code-taxonomy.md) §2.1 의 **10 core 도메인 × 4 계층** 구조 (`backend-core/internal/domain/<도메인>/{view,service,repository,schema}`) 를 따른다. 각 도메인 SDLC 진입점은 [`docs/domain/<도메인>/README.md`](./domain/README.md) 참조.

| 도메인 | view | service | repository | schema |
| --- | --- | --- | --- | --- |
| `auth-session` | `internal/httpapi/{auth,me,identity_resolver}.go` | `internal/auth/keycloak_verifier.go` | (없음 — Keycloak 소유) | `domain/user.go` (idp_subject 캐시) |
| `audit-ops` | `internal/httpapi/{audit,keycloak_events_webhook}.go` | `internal/audit/{keycloak_event_puller,user_sync}.go` | `internal/store/audit_logs.go` + `event_cursors` | `domain/audit.go` (000003, 000031) |
| `rbac-permissions` | `internal/httpapi/{permissions,rbac,authz}.go` | `internal/httpapi/permissions.go` (PermissionCache) | `internal/store/postgres_rbac.go` | `domain/rbac.go` (000005, 000018, 000021, 000024, 000026) |
| `organization-management` | `internal/httpapi/{organization,organizations_search,hr_lookup}.go` | (조직 유효성 + 임명 규칙) | `internal/store/users_units.go` | `domain/{user,primary_unit}.go` (000004, 000019, 000011) |
| `onboarding` | `internal/httpapi/me_onboarding.go` + `onboardingGate` middleware | (게이트 통과 규칙) | (organization-management 공유) | onboarding payload (000033) |
| `platform-lifecycle` | `internal/httpapi/{applications,projects,application_rollup}.go` | (상태머신 + rollup) | `internal/store/{applications,repository_ops}.go` | `domain/application.go` (000013, 000015, 000017) |
| `repository-integration` | `internal/httpapi/{integration_scm_repositories,domain}.go` | (SCM ↔ DevHub 맵핑 검증) | `internal/store/applications.go` (Get/Upsert/ListRepositories) | repositories (000002, 000042, 000043, 000045) |
| `dev-request` | `internal/httpapi/{dev_requests,dev_request_intake_auth,dev_request_intake_tokens_admin}.go` | (상태머신 + promote-tx + 만료 cron) | `internal/store/{dev_requests,dev_request_intake_tokens}.go` | dev_requests (000022), dev_request_intake_tokens (000023, 000027) |
| `integration-registry` | `internal/httpapi/{integration_registry,integrations,external_task_handler}.go` | (preset 매핑 + sync 큐 + Task ingestion) | `internal/store/{integration_registry,external_task_store}.go` | integration_providers (000028, 000038, 000040, 000041) / integration_bindings (000040) / external_task_items (000046) |
| `realtime` | `internal/httpapi/{realtime,realtime_ticket}.go` | (브로드캐스트 필터 + ticket TTL) | `internal/store/realtime_tickets.go` | realtime_tickets (000035) |

## 1.2 Shared 레이어 (`backend-core/internal/shared/`)

도메인 비결합 공통 모듈. 자세한 진입점은 [`docs/shared/README.md`](./shared/README.md).

| 모듈 | 코드 | 주 책임 |
| --- | --- | --- |
| `config` | `internal/shared/config/` | 전역 환경 설정 로더 (`DEVHUB_*` env) |
| `httphelp` | `internal/shared/httphelp/` | 공통 utility helper (에러 응답, 요청 파싱) |
| `integrationcaps` | `internal/shared/integrationcaps/` (PR #409, 2026-05-29) | `providerHasCapability` OR semantics 공용 helper — 3 카피 → 단일 구현 + 11 unit test |

## 1.3 Infrastructure 레이어 (`backend-core/internal/infrastructure/`)

외부 기술 어댑터. 자세한 진입점은 [`docs/infrastructure/README.md`](./infrastructure/README.md).

| 모듈 | 코드 | 주 책임 | ADR |
| --- | --- | --- | --- |
| `keycloak-idp` | `internal/auth/keycloak_verifier.go`, `infra/idp/keycloak-event-listener-spi/`, `infra/idp/sql/` | Keycloak JWKS + admin client + event listener SPI | 0019, 0020, 0022, 0023 |
| `gitea-scm` | `internal/infrastructure/gitea/`, `internal/normalize/gitea/` | Gitea REST 클라이언트 + sync worker (#341) + webhook signature + JSON 정규화 | 0003 |
| `hrdb` | `internal/infrastructure/hrdb/{postgres,mock}.go` | 인사망 어댑터 (실 PG / mock) | 0008, 0010 |
| `commandworker` | `internal/infrastructure/commandworker/{worker,live_worker}.go`, `serviceaction/executor.go` | 명령어 폴링/실행 + sandbox (mock/compose/k8s) | — |
| `database-migration` | `backend-core/migrations/000001~000046_*.sql` | golang-migrate SQL | — |
| `deployment-automation` | `scripts/`, `infra/nginx/`, `docker-compose.{deploy,}.yml` | 배포 전처리 + Nginx 역프록시 + compose | 0018 |

## 2. Phase 로드맵

| Phase | 상태 | 목표 | 주요 산출물 | 다음 판정 기준 |
| --- | --- | --- | --- | --- |
| Phase 1 | done | Go Core 기반 구조 정리 | `internal/config`, `internal/httpapi`, `internal/gitea`, `internal/store` 분리 | `cd backend-core && go test ./...` |
| Phase 2 | done | PostgreSQL 초기 스키마 | `webhook_events` migration | migration 적용 검증 |
| Phase 3 | done | Gitea Webhook raw 수신부 | `POST /api/v1/integrations/gitea/webhooks`, signature 검증, dedupe 처리 | handler 단위 테스트 |
| Phase 4 | done | 프론트 연동 계약 안정화 1차 | role wire format, REST snapshot/WebSocket envelope, integration requirements, 역할별 기본 진입 우선순위 지원 계약 | API 계약 문서화 및 smoke/lint 통과 |
| Phase 5 | done | 프론트 snapshot API 1차 | metrics, infra topology, ci-runs/logs, risk 조회 API, runtime snapshot provider | handler 테스트 및 fallback 동작 확인 |
| Phase 6 | done | 도메인 정규화 1차 | repository/user/issue/pull_request/ci_run/risk 기초 테이블 및 normalize processor | fixture 및 store 테스트 |
| Phase 7 | done | command/audit 기반 액션 API | service action(dry-run + live executor + approval/reject), risk mitigation, command status, idempotency, audit log(Keycloak event polling → `audit_logs` + user sync) | ✅ `commandworker/*`·`serviceaction/*`(mock/compose/k8s) + `internal/audit/*`. actor = Keycloak JWT claim |
| Phase 8 | 부분 (ticket auth + command publish done, infra/ci/risk publish + replay 잔여) | WebSocket 실시간 채널 | `/api/v1/realtime/ws` + **ticket 인증(ADR-0024)** + `command.status.updated` publish + `types` subscription/RBAC filter | **잔여: RM-M4-01** infra/ci/risk event publish + **RM-M4-02** replay/resource scope filter |
| Phase 9 | planned | Python AI gRPC 연결 | Go gRPC client, Python `AnalysisService`, build log summary/risk detection | gRPC 통합 테스트 |
| Phase 10 | planned | Hourly Pull Reconciliation | Gitea REST client, 누락 이벤트 보정 worker | dry-run 및 idempotency 테스트 |
| Phase 11 | planned | 시스템 관리자 기능 고도화 | Runner/server adapter, config 조회, allowlist/seed admin | 권한/audit/health adapter 테스트 |
| Phase 12 | done | 조직/사용자 관리 API | `users`, `org_units`, appointments, hierarchy, unit members CRUD | handler/store 테스트 및 프론트 연동 |
| Phase 13 | done (Keycloak 단일 IdP 전환, ADR-0019) | IdP 도입 — **Keycloak OIDC 단일 IdP** (이전 Hydra/Kratos PoC 는 historical, 전면 제거) | Keycloak OIDC code flow + PKCE, JWKS 검증 + stale-while-error fallback(`internal/auth/keycloak_verifier.go`), Keycloak event polling → `audit_logs` + user sync(`internal/audit/*`), `SetTrustedProxies(nil)`. WS 는 ticket 인증(ADR-0024). | ✅ done. ~~자체 `/api/v1/auth/*` proxy + `/api/v1/accounts/*` admin + Hydra introspection/JWKS verifier~~ 는 **모두 폐기**(historical) — credential/session 은 Keycloak 이 master. |

## 3. 현재 완료 범위

- Gitea webhook raw 저장, signature 검증, dedupe 처리, event 조회 API를 구현했다.
- repository/user/issue/pull_request/ci_run/risk 정규화 테이블과 normalize processor를 구현했다.
- DB-backed domain 조회 API를 구현했다: repositories, issues, pull-requests, ci-runs, risks.
- snapshot handler를 `SnapshotProvider` 경계로 분리하고 runtime/static fallback을 제공한다.
- command/audit migration을 추가했다: `commands`, `audit_logs`.
- `POST /api/v1/admin/service-actions`, `POST /api/v1/risks/{risk_id}/mitigations`, `GET /api/v1/commands/{command_id}`를 구현했다.
- idempotency replay와 command 조회 테스트를 추가했다.
- 승인 불필요 dry-run command worker를 추가해 `pending -> running -> succeeded`로 자동 전이한다.
- `/api/v1/realtime/ws`와 in-process `RealtimeHub`를 추가했고 `command.status.updated`를 publish한다.
- WebSocket `types` query 기반 subscription filtering과 event type별 RBAC read permission check 1차 구현을 추가했다.
- Phase 12 조직/사용자 CRUD API를 구현했다: users CRUD, org unit CRUD, hierarchy, unit members replace/list.
- `GET /api/v1/audit-logs`를 추가했고 조직/사용자 CRUD와 멤버 교체에 audit log 생성을 연결했다.
- `X-Devhub-Actor` 사용 시 deprecation 응답 헤더를 추가해 Phase 13 token actor 전환 경로를 노출했다.
- ~~`docs/backend_api_contract.md` §11을 Hydra/Kratos 기준으로 재작성~~ → **Keycloak 단일 IdP(ADR-0019) 로 정정됨**. Go Core 의 토큰 검증은 Keycloak JWKS verifier(`internal/auth`)다.
- `GET /api/v1/rbac/policy`를 추가하고 프론트 Permissions 화면이 backend policy를 조회하도록 준비했다.
- RBAC policy version table과 `PUT /api/v1/rbac/policy`를 추가해 전체 matrix 교체와 audit log 기록 경계를 만들었다.
- `PUT /api/v1/rbac/policy`에 `system_config: admin` RBAC enforcement를 적용했다.
- `GET /api/v1/me`를 추가해 인증 actor를 DevHub `users`와 매핑하고 effective permissions를 반환한다.
- service action, risk mitigation, audit 조회, 조직/사용자 쓰기 API에 RBAC enforcement를 적용했다.
- ~~Phase 13 Ory Hydra/Kratos PoC scaffold가 main에 반영됐다~~ → **Keycloak 단일 IdP 전환(ADR-0019) 으로 Hydra/Kratos 자산 전면 제거됨**(historical). 현재 IdP 인프라는 Keycloak realm/client + event listener SPI(`infra/idp/keycloak-event-listener-spi/`).
- 브랜치별 memory 구조를 적용해 현재 브랜치 상태 문서는 `ai-workflow/memory/<agent>/<branch>/` 아래에서 관리한다.

### 3.1 2026-05-12 이후 신규 완성 도메인 (ADR-0019 Keycloak 전환 이후, main `cf19c94` 기준)

> 상세 패키지/마이그레이션 근거는 [04_backend_summary.md](./analysis/2026-05-27-codebase-snapshot/04_backend_summary.md) §1·§3·§5 참조.

- **인증 — Keycloak 단일 IdP(ADR-0019)**: 자체 Hydra/Kratos 흐름 + `/api/v1/auth/*` proxy + `/api/v1/accounts/*` admin 전면 폐기. JWKS 검증(`internal/auth/keycloak_verifier.go`, TTL 5분 + stale-while-error fallback). audit 는 Keycloak event polling(`internal/audit/*`, event_cursors dedup + Prometheus).
- **Application / Repository / Project**: CRUD + 상태전이 + rollup + RBAC row-scoping(ADR-0011). `applications.go`·`projects.go`·`repository_ops.go`·`application_rollup.go`.
- **Repository draft→publish lifecycle (#368, migration 000043)**: draft 생성 + publish 요청(`repository_ops.go`). ⚠ **무테스트 머지** — UT/통합테스트 보강 잔여(N-2).
- **SCM↔시스템 repository 양방향 (#363/#366/#373)**: 소유권 분리(source=scm|system, migration 000042) + import(API-89) + create(API-90, gitea CreateRepo) + capability gate + provider_id 단일화(000045). `integration_scm_repositories.go` + UpsertRepository ON CONFLICT(mirror 만 갱신).
- **DREQ (Dev Request)**: intake auth(API token + IP allowlist, ADR-0012) + promote-tx(단일 트랜잭션) + intake token admin(ADR-0014) + 만료 token revoke cron(`internal/devrequest/`).
- **External Integration**: provider/binding registry + auth_mode full 모델(token/basic/app_password/oauth2/agent, migration 000041 OutboundAuth) + base_url(000038) + api_token write-only(000040) + 연결 테스트(API-87) + 범용 webhook ingest(X-Gitea/X-Gogs alias) + HomeLab pull adapter(`internal/integrations/adapters/*`).
- **Onboarding (ADR-0021)**: gate middleware + submit/search/admin review(API-83..86) + onboarding 상태머신(migration 000033). lazy_auto_create 폐기.
- **Gitea SCM 동기화 워커 (#341)**: REST pull(repos/issues/PRs) 정규화 upsert + `integration_sync_jobs` 큐(SKIP LOCKED) + per-provider sync config(`internal/gitea/{client,syncer,worker}.go`).
- **Realtime WS ticket 인증(ADR-0024, #344/#348)**: `POST /api/v1/realtime/ticket` → `?ticket=` single-use 60s(PG/in-memory store). `?access_token=` query 전면 제거(ticket-only cutover).

## 4. 재검토 결과

### 방향성 충돌 없음

- Phase 12 조직/사용자 관리와 현재 command/realtime 작업은 충돌하지 않는다. 둘 다 `backend-core` 라우터와 Postgres store에 공존 가능하다.
- Phase 13 IdP 는 **Keycloak 단일 IdP(ADR-0019)**로 종결됐다. `X-Devhub-Actor` 는 inbound 명시 거부(ADR-0006), production actor 경계는 Keycloak OIDC JWT claim 이다.
- service action command의 dry-run 자동 성공 전이는 “실제 executor 도입 전 안전한 시뮬레이션”으로 유지하되, live executor(mock/compose/k8s 모드)도 구현됐다.

### 조정이 필요한 전제 (2026-05-27 갱신)

- WebSocket: ticket 인증(ADR-0024) + command publish + `types` subscription/RBAC filter 까지 done. **잔여는 infra/ci/risk event publish(RM-M4-01) + replay/resource scope filter(RM-M4-02)** 뿐 — Phase 8 부분.
- Command/Audit: command 생성·상태 전이·approval/reject·live executor + audit(Keycloak event polling) 까지 done → Phase 7 done.
- ~~Phase 13 계정/인증 계약은 Hydra/Kratos 기준 + JWKS/introspection verifier·admin identity wrapper 잔여~~ → **폐기**(historical). Keycloak 단일 IdP 전환으로 JWKS verifier 만 사용하고 자체 accounts/auth proxy 는 제거됨. admin identity wrapper 불필요.
- 로컬 검증 명령은 Go/NPM/native PostgreSQL 중심(no-docker default, ADR-0003). Docker Compose 자산은 환경별 git 추적 외부.

## 5. 기능 단위별 우선순위 계획 (Functional Priorities)

> ⚠ **2026-05-27 갱신**: 본 §5 는 2026-05-12 시점의 M2~M4 계획을 담고 있었으나 ADR-0019 Keycloak 전환 + 도메인 다수 완성으로 대부분 종결됐다. 우선순위·잔여 carve 의 **최신 source-of-truth 는 [`docs/planning/release_v1_roadmap.md`](./planning/release_v1_roadmap.md)** 다. 아래는 backend 트랙의 완성/잔여 요약으로 정정한다.

### [P0] M2: 인증 및 사용자 기반 — ✅ 종결 (Keycloak 단일 IdP 전환)

M2 인증 기반은 **Keycloak 단일 IdP(ADR-0019)로 종결**됐다. 이전 Hydra/Kratos/accounts 항목은 historical.

- **Keycloak OIDC + JWKS verifier**: ✅ done — `internal/auth/keycloak_verifier.go`(TTL 5분 + stale-while-error fallback). ~~Hydra introspection verifier~~ 폐기.
- **`X-Request-ID` middleware + audit enrichment**: ✅ done — `source_ip`/`request_id`/`source_type`.
- ~~**Accounts admin endpoints** (`POST/PUT/PATCH/DELETE /api/v1/accounts`)~~: **폐기**(historical) — 계정 lifecycle 은 Keycloak + Onboarding admin review(API-86)로 대체.
- ~~**Kratos self-service webhook → `audit_logs`** (PR-M2-AUDIT)~~: **폐기**(historical) — audit 는 Keycloak event polling(`internal/audit/*`)으로 대체 완성.
- ~~**Hydra JWKS verifier / Identity Admin Wrapper**~~: **폐기**(historical) — Keycloak JWKS verifier 단일화.

### [P1] M3: 사용자 및 조직 관리 — ✅ 완성

- **User/Org CRUD + RBAC**: ✅ done — users CRUD + org unit 계층/임명 + single-leader invariant(SQL) + RBAC 4-boolean matrix + row-scoping(ADR-0011). `organization.go`·`users_units.go`·`permissions.go`.
- ~~**Sign Up Service** (`POST /api/v1/auth/signup` + Kratos identity)~~: **cancelled** — 외부 IdP 시나리오(IdP 팀/HRDB ETL 책임). Onboarding self-service(ADR-0021)로 대체.
- **Onboarding**: ✅ done — gate + submit/search/admin review(API-83..86) + 상태머신.

### [P2] M4: 실시간 대시보드 및 AI — 🟡 부분 (잔여 = RM-M4)

- **WebSocket**: ticket 인증(ADR-0024) + command publish done. **잔여: `infra.node.updated`/`ci.run.updated`/`risk.updated` publish(RM-M4-01) + replay/resource scope filter(RM-M4-02)**.
- **gRPC AnalysisService**: 🔴 미구현 — `backend-ai/main.py` 스켈레톤(`/health` 만). **v2 범위**(AI Gardener/로그 분석/Weekly report).

### [P3] M4: 외부 연동·SCM·시스템 관리 — ✅ 대부분 완성

- **External Integration**: ✅ done — provider/binding registry + auth_mode full 모델 + base_url + api_token(write-only) + 연결 테스트 + 범용 webhook ingest + HomeLab pull.
- **Gitea SCM 동기화**: ✅ done — REST pull sync worker + `integration_sync_jobs` 큐(SKIP LOCKED). **잔여: RM-M4-06 Hourly Pull 정밀화(reconciliation 스케줄, issue #231)**.
- **Repository draft→publish + SCM 양방향**: ✅ done(#368/#363/#366/#373). **잔여: #368 무테스트 보강(N-2)**.
- **System Admin 운영 가시성**: 🟡 잔여 — sync job 큐/provider health 운영 대시보드(RM-M4-07, BE 데이터는 있으나 운영 view 없음).

### [보안 부채] (release_v1 §3.3 / snapshot 04 §6)

- **평문 secret envelope 암호화(#6)**: 🟡 미해소 — `credentials_ref`/`api_token`/`auth_secret` 평문 저장. 신규 secret 은 write-only 응답 패턴(`<field>_set` bool)이나 저장 자체는 평문. DEK/KMS 암호화 + 키 관리 ADR 잔여.
- **마이그레이션 prefix uniqueness CI guard**: 🟡 — 000042 동시 발급 충돌 이력(CI bypass 통과). `uniq -d` gate 강화 필요(N-5).
- **Keycloak SPI realm events push 전환(P3-5)**: polling 30s → <1s, SPI JAR 빌드·배포(사내 동반).

---

## 6. 다음 작업 큐 (Next Tasks Queue)
- [x] ~~API 계약 §11 Hydra/Kratos 재작성~~ (당시 done — 이후 Keycloak 단일 IdP(ADR-0019)로 재정정됨)
- [x] Bearer token 검증 middleware 설계 및 최소 구현 (현행 = Keycloak JWKS verifier)
- [x] RBAC policy 조회 API 및 프론트 Permissions 연동 준비
- [x] RBAC policy persistence/edit API와 audit 경계
- [x] RBAC policy edit enforcement (`system_config: admin`)
- [x] `GET /api/v1/me` 및 DevHub user-role lookup
- [x] service action/risk/audit/organization RBAC enforcement
- [x] 인증 actor 미매핑/비활성 시 role fallback 우회 차단
- [x] `X-Devhub-Actor` deprecation warning 경로 추가
- [x] audit log 조회 API와 organization CRUD audit 연결
- [x] WebSocket 인증/구독 필터 1차 구현
- [x] WebSocket publish lock 개선
- [x] service action approval/reject API 및 audit boundary
- [x] approved live service action query 및 executor adapter boundary
- [x] simulation service action executor 및 명시적 main 주입 설정
- [x] ~~Hydra introspection verifier (admin endpoint)~~ → **Keycloak JWKS verifier 로 대체**(ADR-0019, `internal/auth/keycloak_verifier.go`)
- [x] Bearer token middleware + audit `source_ip`/`request_id` enrichment
- [x] ~~Accounts admin endpoints 4종~~ → **폐기**(historical) — Keycloak + Onboarding admin review(API-86)로 대체
- [x] ~~PR-M2-AUDIT (Kratos self-service webhook → `audit_logs`)~~ → **폐기**(historical) — Keycloak event polling(`internal/audit/*`)으로 audit 완성
- [x] Platform/Repository/Project CRUD + 상태전이 + rollup + row-scoping
- [x] DREQ intake auth + promote-tx + token admin + 만료 cron
- [x] External Integration provider/binding + auth_mode full + base_url + api_token + 연결 테스트 + 범용 webhook ingest + HomeLab pull
- [x] Onboarding gate + submit/search/admin review(API-83..86)
- [x] Gitea SCM pull sync worker + `integration_sync_jobs` 큐
- [x] Repository draft→publish(#368) + SCM 양방향 import/create(API-89/90) + provider_id 단일화(#373)
- [x] Realtime WS ticket 인증(ADR-0024, `?access_token=` 제거)
- [ ] **N-2**: repository draft→publish(#368) UT/통합테스트 보강 (무테스트 머지분)
- [ ] **N-3**: SCM import/create + draft/publish happy-path E2E
- [ ] **RM-M4-01**: WebSocket infra/ci/risk event publish
- [ ] **RM-M4-02**: WebSocket replay + resource scope filter
- [ ] **#6**: 평문 secret envelope 암호화 (credentials_ref/api_token/auth_secret DEK + 키 관리 ADR)
- [ ] **N-5**: 마이그레이션 prefix uniqueness CI guard 강화
- [ ] **RM-M4-06**: Gitea Hourly Pull 정밀화 (reconciliation 스케줄, issue #231)
- [ ] **RM-M4-07**: System Admin 운영 대시보드용 sync job 큐/provider health 노출
- [ ] **v2**: Python AI gRPC AnalysisService (backend-ai 스켈레톤 → 실구현)
- [ ] **carve (P2, 2026-05-29 SDLC 재정비)**: `PlatformRepository` cross-domain decouple — `*IntegrationRepository` embed 제거 (review agent P1)
- [ ] **carve (P2, 2026-05-29 SDLC 재정비)**: `ApplicationStore` interface slim — 13+ integration 메서드 → integration-registry 도메인 이관
- [ ] **carve (P1, 2026-05-29 SDLC 재정비)**: CI `backend-integration` job 복원 (refactor stabilize 후 `&& false` gate 제거)
- [ ] **P0 carve (code-taxonomy §3 P0-2/P0-3/P0-4)**: `store/applications` (LoC 1172) + `httpapi/applications` (LoC 1066) + `httpapi/organization` (LoC 1019) + `store/users_units` (LoC 1263) 도메인 4 계층 하위로 file split

## 6.1 변경 이력

| 일자 | 변경 | 메모 |
| --- | --- | --- |
| 2026-05-29 | SDLC 재정비 sprint #408~#416 정합 — §1.1 도메인 4 계층 매핑 + §1.2 Shared (`integrationcaps` 신규 PR #409) + §1.3 Infrastructure 진입점 명시 + §6 carve out 4건 추가 (PlatformRepository decouple / ApplicationStore slim / CI 복원 / P0 file split). main HEAD `273d9d4`. | sprint `claude/work_260529-k` |

## 7. Blocked 항목

- 현재 백엔드 코드 진행을 막는 hard blocker는 없다.
- ~~Phase 13 round-trip 검증은 Hydra/Kratos native binary 준비 필요~~ → **해소**(historical). Keycloak OIDC E2E 는 CI shard 1/2 에서 실 연동(Keycloak realm/client 설정) 검증된다.
- **Keycloak group staging-prod 적용(P1-3, issue #214)** + **SPI realm events push 전환(P3-5)** 은 사내 Keycloak admin/배포 동반 항목으로 backend 코드 외부 의존이다.
- 외부 네트워크/사내 SSL inspection 환경에서는 Go module, npm package, font 다운로드가 막힐 수 있으므로 mirror 또는 사내 CA 설정을 사용한다.

## 8. 진척 관리 방식

- 이 문서의 Phase 상태는 `planned`, `in_progress`, `blocked`, `done` 중 하나로만 관리한다.
- 코드 변경이 포함된 Phase는 테스트 또는 실행 검증 결과를 남긴 뒤 `done`으로 전환한다.
- 세션 종료 전 현재 브랜치별 `ai-workflow/memory/<agent>/<branch>/state.json`, `session_handoff.md`, 최신 backlog에서 이 문서와 현재 Phase를 함께 갱신한다.
