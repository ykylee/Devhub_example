# 01. 코드베이스 현재 상태 상세 분석

- 문서 목적: 2026-05-27 시점 DevHub 코드베이스를 백엔드/프론트엔드/인프라/테스트 축으로 전수 분석한다.
- 기준 커밋: main `cf19c94` (PR #374), 마지막 기능 커밋 `99d6edc` (PR #373).
- 작성 방식: `backend-core/internal/httpapi/router.go`, `internal/domain`, `internal/store`, `frontend/app`, `frontend/components`, `frontend/lib`, `migrations/` 전수 탐색 + 워크플로우 메모리 교차 검증.

---

## 1. 시스템 개요

DevHub 는 **역할별 기본 진입 우선순위 대시보드 + 조직/권한 관리 + 외부 시스템(SCM/ALM/CI-CD/문서/HomeLab) 연동**을 제공하는 통합 개발 관리 플랫폼이다. 3 tier:

```
[ Frontend: Next.js 16 (App Router) ]
        │  REST (/api/v1/*) + WebSocket (ticket auth)
        ▼
[ Backend Core: Go + Gin ]  ── JWKS ──▶ [ Keycloak (단일 IdP, OIDC/PKCE) ]
        │  pgx                                  ▲ event polling (audit)
        ▼                                       │
[ PostgreSQL ]   ◀── webhook/pull ──  [ Gitea / HomeLab Agent / 외부 provider ]
        ▲
        │ gRPC (예정, 미구현)
[ Backend AI: Python + FastAPI (스켈레톤) ]
```

- **인증**: Keycloak OIDC 단일 IdP (ADR-0019). 자체 Hydra/Kratos 흐름 + `/api/v1/auth/*` proxy + `/api/v1/accounts/*` admin 은 모두 폐기 완료(historical).
- **권한**: per-resource 4-boolean RBAC matrix(`{view, create, edit, delete}`) + 라우트 매핑 표 + deny-by-default (ADR-0002, ADR-0011 row-scoping).
- **실시간**: REST snapshot + WebSocket. WS 인증은 ticket 패턴(ADR-0024, `POST /api/v1/realtime/ticket` → `?ticket=`, single-use 60s, PG/in-memory).
- **배포**: native default (no-docker). 컨테이너 자산은 환경별 git 추적 외부(ADR-0003).

---

## 2. 백엔드 (backend-core, Go)

### 2.1 패키지 구조 및 역할

| 패키지 | 구현 파일 | 테스트 | 역할 |
| --- | ---: | ---: | --- |
| `internal/httpapi` | (대다수) | 37 | HTTP 라우터·핸들러·미들웨어(인증/권한/request-id/실시간/스냅샷). 42 핸들러 파일. |
| `internal/store` | 다수 | 11 | PostgreSQL 영속화 계층(pgx). 도메인별 CRUD + 트랜잭션. |
| `internal/domain` | — | 3 | 도메인 모델 struct + enum + 상태머신 + 검증. |
| `internal/gitea` | — | 6 | Gitea webhook 서명 검증(push) + REST pull sync worker(client/syncer/worker). |
| `internal/audit` | — | 4 | Keycloak event polling → `audit_logs` + user_sync + Prometheus metric. |
| `internal/auth` | — | 1 | Keycloak JWKS 검증기 + stale-while-error fallback. |
| `internal/commandworker` | — | 2 | command dry-run/live executor + `command.status.updated` publisher. |
| `internal/serviceaction` | — | 1 | service action executor(mock/compose/k8s 모드). |
| `internal/config` | — | 1 | 환경변수 로딩·검증. |
| `internal/devrequest` | — | 1 | DREQ intake token cron(만료 revoke + Prometheus gauge). |
| `internal/normalize` | — | 1 | webhook payload → domain changeset 정규화. |
| `internal/hrdb` | — | 1 | HR DB 조회(Mock + PostgreSQL adapter). |
| `internal/integrations/adapters` | — | 3 | HomeLab pull adapter(file/HTTP) + health policy + metrics collector. |
| **합계** | **89** | **72** | |

### 2.2 HTTP API 표면 (router.go, ~100 라우트)

도메인별 그룹:

| 그룹 | 대표 endpoint | 인증/권한 | API ID |
| --- | --- | --- | --- |
| 공개 | `GET /health`, `GET /metrics` | 없음 | API-01 |
| me / onboarding | `GET·PATCH /api/v1/me`, `POST /api/v1/me/onboarding`, `GET /api/v1/organizations/search` | OIDC (+ onboardingGate) | API-32/33, 83..86 |
| admin review | `POST /api/v1/admin/users/:id/review` | system_admin | API-86 |
| dashboard/events | `GET /api/v1/dashboard/metrics`, `GET /api/v1/events` | OIDC | API-04/05 |
| infra/topology | `GET /api/v1/infra/{nodes,edges,topology,services}`, `POST .../services/snapshot`, `GET .../topology/v2` | OIDC (+ agent token) | API-06/07, 76..78 |
| repository | `GET /api/v1/repositories`, `POST /api/v1/repositories`(draft), `POST .../:id/publish` | OIDC + RBAC | API-08, draft/publish(#368) |
| repository ops | `GET .../:id/{activity,pull-requests,build-runs,quality-snapshots}` | developer+ | API-51..54 |
| pipelines/risk | `GET /api/v1/{issues,pull-requests,ci-runs,risks}`, `POST /api/v1/risks/:id/mitigations` | OIDC / manager+ | API-09..13, 16 |
| audit | `GET /api/v1/audit-logs` | manager+ | API-18 |
| RBAC | `GET·PUT·POST·DELETE /api/v1/rbac/policies*` | system_admin | API-26..29, 38..40 |
| commands | `GET /api/v1/commands/:id` + 승인/거절, `POST /api/v1/admin/service-actions` | OIDC / system_admin | API-15/17/36/37 |
| users/org | `GET·POST·PATCH·DELETE /api/v1/users*`, `/api/v1/organization/{hierarchy,units,units/:id/members}` | system_admin | API-33/34 |
| applications | `GET·POST·PATCH·DELETE /api/v1/applications*` + `/repositories` link + `/rollup` | RBAC row-scoped | API-41..50, 57 |
| projects | `GET·POST·DELETE /api/v1/projects*` (+ repository scope) | RBAC | API-55/56 |
| integration legacy | `GET /api/v1/integrations` + CRUD | manager+ | API-58 |
| integration registry | `GET·POST·PATCH·DELETE /api/v1/integration/providers*` + `/sync` + `/scm-repositories` + `/import-repositories` + `/create-repository` + `/webhook` + `/test-connection` + bindings | system_admin / 토큰 | API-69..90 |
| DREQ | `POST /api/v1/dev-requests`(intake) + 목록/상세/register/reject/PATCH/DELETE + `dev-request-tokens*` | intake token / RBAC / system_admin | API-59..68, 79 |
| gitea/keycloak webhook | `POST /api/v1/integrations/gitea/webhooks`, `POST /api/v1/internal/keycloak-events` | HMAC / webhook secret | API-02 |
| realtime | `POST /api/v1/realtime/ticket`, `GET /api/v1/realtime/ws` | OIDC + ticket | API-14 |
| hr | `GET /api/v1/hr/lookup` | OIDC | — |

권한 enforcement: `internal/httpapi/permissions.go` 의 `routePermissionTable`(라우트→(resource, action) 매핑) + `enforceRoutePermission` 미들웨어 + `PermissionCache`. role rank: developer(10) < manager(20) < system_admin(30), 별도 `pmo_manager` 커스텀.

### 2.3 도메인 모델 (internal/domain)

주요 aggregate (status 머신 포함):

- **Repository** — `{Source(scm|system), ProviderID(FK), ProviderKey(derived), Status(draft|active), PublishRequestedAt, PublishedAt}`. SCM mirror + system-owned 필드 분리(#363) + draft→publish lifecycle(#368) + provider_id 단일화(#373).
- **Application** — `{Key, Status(planning|active|on_hold|closed|archived), Visibility, OwnerUserID, LeaderUserID, DevelopmentUnitID}` + ApplicationRepository link + rollup.
- **Project** — `{ApplicationID?, RepositoryID?, Key, Status, Visibility}`. standalone(repository_id NULL) 허용 + partial unique key.
- **IntegrationProvider** — `{ProviderType(alm|scm|ci_cd|doc|infra), AuthMode(token|basic|app_password|oauth2|agent), Capabilities, BaseURL, APIToken(write-only), OutboundAuth}`.
- **IntegrationBinding** — `{ScopeType(application|project), ScopeID, ProviderID, ExternalKey, Policy(summary_only|execution_system)}`.
- **DevRequest** — 6-state(`received|pending|in_review|registered|rejected|closed`) + Promote-to-Application/Project + intake token(hashed + allowed_ips + expires_at).
- **Command** — 6-state(`pending|running|succeeded|failed|rejected|cancelled`) + 승인 워크플로우 + dry-run.
- **AppUser** — `{Role, Status, IdPSubject, PrimaryUnitID, CurrentUnitID, OnboardingCompletedAt, ReviewStatus}`. Onboarding 상태머신(ADR-0021).
- **OrgUnit / UnitAppointment** — 조직 계층(company/division/team/group/part) + single-leader invariant(SQL) + 겸임/파견 모델 + total_count MV.
- **AuditLog** — `{SourceType(oidc|webhook|keycloak_event|system), SourceEventID(dedup), SourceIP, RequestID}`.
- **RealtimeTicket / EventCursor / InfraServiceSnapshot / QualitySnapshot / BuildRun / PRActivity / RepositoryActivity / SCMProvider**.

### 2.4 백그라운드 워커 (main.go)

| 워커 | 기동 조건 | 주기 | 역할 |
| --- | --- | --- | --- |
| commandworker.Worker | pgStore != nil | 2s | pending command poll + dry-run 전이 |
| commandworker.LiveWorker | ServiceActionExecutorMode 설정 | — | 실시간 service action 실행 |
| HomeLabPullLoop | HomeLabPullEnabled | 30s | homelab 상태 pull → infra snapshot |
| DREQIntakeTokenCron | DREQTokenCronEnabled | 10m | 만료 토큰 revoke + Prometheus gauge |
| KeycloakEventListener | 활성 + idpAdmin | 30s | Keycloak event polling → audit + user sync |
| OnboardingPendingReviewGauge | counter 구현 시 | 60s | pending_review 카운트 gauge |
| GiteaSyncWorker | pgStore != nil | 30s | `integration_sync_jobs` 큐 소비 + per-provider pull sync |

### 2.5 마이그레이션 (45개)

도메인별 그룹 (000001..000045):

- **Core/SCM**: 000001 webhook_events / 000002 repositories·issues·pull_requests·ci_runs·risks / 000003 commands·audit_logs / 000012 scm_providers.
- **Org/User**: 000004 org_units·users·unit_appointments / 000006-000009 users 확장 / 000010 hrdb.persons / 000011 total_count MV / 000019 single-leader / 000030 idp_subject rename / 000033 onboarding(completed_at·review_status).
- **RBAC**: 000005 rbac_policies seed / 000018·000021·000024·000026 resource 확장.
- **Application/Project**: 000013 applications / 000014 application_repositories / 000015 projects / 000016 project_members·integrations / 000017 repo ops 지표 / 000034 project_repositories / 000037·000039·000044 project key/repo nullable·unique.
- **Integration**: 000028 providers·bindings·sync_jobs / 000029 infra_service_snapshots / 000038 base_url / 000040 api_token / 000041 auth_credentials.
- **Repository 소유권/lifecycle**: 000042 source·provider_id·description / 000043 draft status / 000045 scm_provider DROP(provider_id 단일화).
- **DREQ**: 000022 dev_requests / 000023·000027 intake tokens.
- **Realtime/Event**: 000031 event_cursors / 000032 audit source_event_id / 000035 realtime_tickets.
- **기타**: 000036 applications key 완화.

> **주의(개선 후보)**: 000042가 두 PR(#363 repositories / #368 projects)에서 동시 발급되어 충돌 → #371이 #368분을 000044로 재번호한 이력이 있다. CI bypass 머지가 prefix 충돌을 통과시켰던 사례 (→ 향후 방향 §6 참조).

### 2.6 backend-ai (Python)

`backend-ai/main.py` 만 존재. FastAPI `GET /health` 1개 + gRPC AnalysisService TODO 주석. **사실상 스켈레톤** — AI Gardener/로그 분석/Weekly report 는 v2 범위로 미구현.

---

## 3. 프론트엔드 (frontend, Next.js 16 App Router)

### 3.1 기술 스택

- **Next.js 16.2.6 + React 19.2.4** (App Router, Server Components).
- 상태: **Zustand 5** (persist + subscribeWithSelector) + **@tanstack/react-query 5**.
- UI: **TailwindCSS 4** + framer-motion 12 + lucide-react + **@xyflow/react 12**(React Flow) + dagre + recharts.
- 테스트: **Vitest 4** + @testing-library + jsdom / **Playwright 1.59**.

### 3.2 페이지 (33개)

- **인증/온보딩**: `/login`, `/auth/{callback,logout,signup,error}`, `/signup`(redirect), `/onboarding`, `/` (developer redirect).
- **대시보드**: `/(dashboard)/{developer,manager,gardener,account}` + layout(AuthGuard·Header·Sidebar·OnboardingBanner).
- **도메인 페이지**: `/applications` + `[id]`, `/projects` + `[id]`, `/repositories` + `[id]`, `/dev-requests`, `/organization`.
- **admin**: `/admin`, `/admin/topology-v2`, `/admin/catalog`, `/admin/settings/{users,permissions,audit,organization,applications,dev-requests,dev-request-tokens,integrations,integration-bindings}` + settings index.

### 3.3 컴포넌트 (50개, 폴더별)

`account`(ProfileSelfEdit), `admin/users`(ConfirmReviewModal·PendingReviewPanel), `dashboard`(GardenerFeed), `dev-request`(Table·DetailModal·IntakeToken×3·Widget), `integration`(Bindings·Provider·ScmRepository·Import 모달/테이블 7), `layout`(Header·Sidebar·AuthGuard), `onboarding`(Form·Banner·OrgPicker), `organization`(OrgTree·Node·Grid·Table·Member·User모달·Permission 11), `project`(Project·Repository·Application 테이블/모달 6), `ui`(Modal·Badge·Toast·FilterBar·ComboBox·PageState·ActionMenu·DestructiveConfirm 등 공통 10).

### 3.4 서비스 레이어 (18 service)

`api-client`(401 자동 refresh) · `auth`(OIDC PKCE) · `identity`(whoAmI) · `realtime`(ticket WS) · `websocket`(legacy) · `project`/`application`/`repository` · `dev_request`/`dev_request_token` · `integration`(provider/binding/test-connection/scm-repo) · `audit` · `dashboard` · `gardener` · `infra`(topology v2) · `rbac` · `onboarding` · `risk`. + `lib/auth/{token-store,pkce,role-routing}` + `lib/config/endpoints` + `lib/store`(Zustand).

### 3.5 운영 UI 전환 상태

- mock 대시보드(Work/Quality/Sys Admin) 사이드바 비노출 + `frontend/lib/archive/` 로 분리(#334/#340).
- `PageState` 공통 컴포넌트(loading/error/empty/retry) + 에러 메시지 표준화(`error-message.ts`) + 상세 페이지 실데이터 렌더링(#342).
- application dashboard 4 stat 카드 모두 실데이터화(#369 — 'Active Regions' mock 잔재 → 'Active Applications').

---

## 4. 인프라 / 운영

- **CI**: `.github/workflows/ci.yml` — Detect Changed Paths + Workflow Lint(actionlint) + Backend Unit/Integration + Migration Prefix + Frontend Unit + E2E shard 1/2(Playwright, Keycloak/OIDC 실 연동). native PG 15.
- **인증 인프라**: Keycloak event listener SPI(Java, `infra/idp/keycloak-event-listener-spi/`) + `Dockerfile.keycloak`. ADR-0022/0023 으로 Keycloak 버전 pin(25.0→26.0).
- **reverse proxy**: nginx 단일 포트 `/devhub` prefix(ADR-0018) + X-Forwarded-Host 정합.
- **관측**: Prometheus `/metrics`(JWKS stale / intake token / onboarding pending / sync) + Alertmanager 규칙 + Grafana JSON(ADR-0016).

---

## 5. 테스트 자산

| 축 | 자산 | 비고 |
| --- | --- | --- |
| 백엔드 단위/통합 | 72 test 파일 | httpapi 37 + store 11 + gitea 6 + audit 4 등. `go test ./...` green(13 pkg). |
| 프론트 단위(Vitest) | 10 파일 | auth(pkce/token-store/role-routing) + utils + login page + AuthGuard + project.service + integration-presets + IntakeToken×2. |
| E2E(Playwright) | 28 spec | auth/rbac/onboarding/admin-*/repositories/applications/dev-requests/topology/screenshots 등. negative-path 4 spec 포함. |

**갭**: backend 핸들러 대비 frontend 단위 테스트 밀도가 낮다(서비스 18개 중 vitest 커버 2개). repository draft→publish(#368)·SCM create/import(#363/#366) 등 최신 backend 기능의 전용 E2E 는 부분적.

---

## 6. 코드베이스 건강도 요약

| 항목 | 상태 | 근거 |
| --- | --- | --- |
| 아키텍처 일관성 | 🟢 양호 | 도메인-store-handler 계층 + RBAC 게이트 + audit 일관 패턴. |
| 인증/보안 baseline | 🟢 양호 | Keycloak 단일 IdP + JWKS stale fallback + ticket WS + SetTrustedProxies(nil). |
| 평문 secret 저장 | 🟡 부채 | `credentials_ref`/`api_token`/`auth_secret` 평문 — envelope 암호화 carve(#6) 미해소. |
| 마이그레이션 hygiene | 🟡 주의 | prefix 충돌 이력(000042) — CI prefix guard 강화 필요. |
| backend-ai | 🔴 미구현 | 스켈레톤만. v2 범위. |
| 문서 drift | 🔴 부채 | 추적성/로드맵이 2026-05-16~21 정지. 본 PR 정합 대상. |
| FE 테스트 밀도 | 🟡 부채 | vitest 10 파일 — 신규 도메인 UI 커버 부족. |

상세 SDLC 정합은 [`02_sdlc_chain_status.md`](./02_sdlc_chain_status.md), 도메인별 FE/BE 정리는 [03](./03_frontend_summary.md)/[04](./04_backend_summary.md), 균형은 [05](./05_fe_be_balance.md), 향후 방향은 [06](./06_future_direction.md) 참조.
