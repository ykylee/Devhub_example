# Phase 1 + 1.5 + 3 Mirror List (D-72, docs + source code + domain/architecture/infrastructure/validation mass ingest scope ~220 file)

- **문서 목적**: `scripts/wiki-sync-devhub.sh` 의 mirror source list. **Phase 1** (docs subset, 85 file) + **Phase 1.5** (소스코드 + workflow + scripts + branch memory maintenance subset, ~55 file) + **Phase 3** (domain + architecture + infrastructure + validation mass ingest, ~78 file) = **~220 file**. DevHub repo 의 **소스코드 maintenance 가 wiki 만으로 가능**하도록 mirror scope 확장 + 위키의 LLM Wiki 의 mass ingest 정공법.
- **범위**: 본 mirror list 의 scope = docs/adr + docs/governance + docs/planning + docs/setup + docs/requirements.md + docs/openapi.yaml + ai-workflow/memory (main flat + branch memory) + **.github/workflows/*.yml** + **scripts/*.sh** + **backend-core 의 maintenance-critical subset** + **frontend/tests/e2e spec subset** + **docs/domain/** (66 file) + **docs/architecture/** (1 file) + **docs/infrastructure/** (9 file) + **docs/validation/** (2 file). **scope 외**: frontend source bulk (10000+ file, 빌드 산출물/node_modules) + backend bulk.
- **대상 독자**: yklee (owner), LLM agent (코드 maintenance 작업 시 RAG source), `scripts/wiki-sync-devhub.sh` 의 자동 mirror 실행 시 source 결정.
- **상태**: active (D-72 Phase 1+1.5+3, 2026-06-13)
- **최종 수정일**: 2026-06-13 (Phase 3 추가: docs/domain + docs/architecture + docs/infrastructure + docs/validation ~78 file mass ingest, scripts 갱신, lint-config.toml 갱신, wiki page 자동 생성 정공법)
- **관련 문서**:
  - [`./scope-and-rationale.md`](./scope-and-rationale.md) (Phase 1+1.5+3 scope + D-72 Q1~Q6 적용)
  - [`./operation-sop.md`](./operation-sop.md) (sync + lint SOP, **2026-06-13 갱신: Phase 3 mass ingest SOP**)
  - [`./lint-config.toml`](./lint-config.toml) (L07 ADR 면제 + L03/L10 면제 + **L02 broken link Phase 1.5+3 갱신**)
  - `scripts/wiki-sync-devhub.sh` (**2026-06-13 갱신: Phase 1.5+3 mirror pattern 추가, 15 패턴**)

## Phase 1 (docs subset, 85 file) — 기존 정공법

기존 1.1~1.6 (ADR 31 + Governance 5 + Planning 27 + Setup 15 + Requirements + OpenAPI + memory main flat 3) = **85 file**. 본 section 의 내용은 위 historical 정공법 정합. 본 갱신은 Phase 1.5 추가가 목적.

## 1.7 Phase 1.5 (source code + workflow + scripts + branch memory maintenance subset, ~65 file) — **2026-06-13 신규 (X-5 follow-up 4 file + Sprint A 6 file 추가)**

**Phase 1.5 의 목적**: PR #578+#579+#580+#581 의 source code 변경분 + 운영/유지보수 critical file + branch memory 를 mirror 에 포함. wiki 만으로 **code maintenance 가능** (단순 SSOT 참조 + 1차 layer reasoning).

**mirror 정공법**:
- 22 file = 본 sprint 의 source code 변경분 (PR #578+#579+#580+#581 의 범위 외 file + branch memory/ 의 sprint 한정)
- ~33 file = 운영/유지보수 critical file (keycloak_verifier.go, keycloak_admin_client.go, saovae_stub.go, fixtures.ts, signin helper, critical frontend tests, ci.yml 본문, etc.)
- 100% = **찾기 쉽고 + 변경 빈도 낮고 + 문서화 가능** (lint-config.toml 의 L02 검증으로 위키 본문과 동기)

### 1.7.1 Backend Go source (maintenance critical subset, ~25 file)

**소스 경로**: `backend-core/internal/` 의 key file (본 sprint 의 verification 으로 정확한 경로 정합).

| source path | 의의 | maintenance 디테일 |
|---|---|---|
| `backend-core/main.go` | runtime injection (`DEVHUB_BUILD_TIER`) | saovae_stub default + `=internal` real |
| `backend-core/internal/sso-integrations/keycloak/saovae_stub.go` | 사외 build 의 default wiring (sprint -a follow-up PR #539) | 4 port stub + webhook handler |
| `backend-core/internal/domain/auth-session/integration/ports.go` | Port interface (ADR-0030) | 4 port + 4 type alias + 3 sentinel error |
| `backend-core/internal/domain/auth-session/view/auth.go` | BearerTokenVerifier interface (deprecation) | canonical = integration/ports.go |
| `backend-core/internal/domain/auth-session/view/handler.go` | IdentityAdmin + OIDCLogoutClient interface | canonical = integration/ports.go |
| `backend-core/internal/domain/application-lifecycle/routing/auto_route.go` | **PR #579 신규** — AutoRouter interface + 3 case pattern matcher | external_ref `^GITEA-([0-9]+)$` + req_department + graceful degradation |
| `backend-core/internal/domain/dev-request/view/voc_handler.go` | voc handler (PR #514 + PR #579 AutoRouter 통합) | createOrGetVoc + AutoRouter.Route() + RouteVoc() + auto_routed envelope |
| `backend-core/internal/domain/audit-ops/view/keycloak_events_webhook.go` | Keycloak event webhook handler (ADR-0030/0031 + PR #578 e2e-internal 폐기 정공법) | event listener type assertion + webhook routing |
| `backend-core/internal/domain/audit-ops/service/keycloak_event_puller.go` | Keycloak event puller (PR #189~#193 + #241) | 1분 polling + audit_logs 통합 |
| `backend-core/internal/domain/audit-ops/view/audit.go` | audit actor enrichment (PR-D) | source_ip + request_id + source_type + X-Request-ID |
| `backend-core/internal/domain/audit-ops/repository/audit_logs.go` | audit_logs repository | X-Request-ID + audit actor propagation |
| `backend-core/internal/domain/rbac.go` | RBAC domain (rbac_policies seeded) | system_admin / developer / manager / team_manager |
| `backend-core/internal/domain/rbac-permissions/view/rbac.go` | RBAC PermissionCache + permission cache (PR #29~#31) | LISTEN/NOTIFY (future ADR-0007) |
| `backend-core/internal/httpapi/repository_ops.go` | ListRepositoryBuildRuns (P1-7 N-9 PR #555) | devhub_repository_build_runs + platformStoreOrUnavailable guard |
| `backend-core/internal/store/repository_ops.go` | postgres store 의 repository_ops | BuildRun table CRUD + path 차이 (httpapi vs store 의 layer 분리) |
| `backend-core/internal/store/repository_pull_ingest.go` | **X-5 production wire follow-up 신규** — PostgresStore 의 `UpsertPullActivity` / `UpsertBuildRun` / `UpsertQualitySnapshot` 3 method | pr_activities / build_runs / quality_snapshots ON CONFLICT upsert (migration 000001 L402/95/505 + 000045 partial unique 정합) |
| `backend-core/internal/store/repository_pull_state.go` | **X-5 production wire follow-up 신규** — PostgresStore 의 `UpdatePullState` / `IncrementConsecutiveFailures` / `ResetConsecutiveFailures` / `SetBackoff` / `BackoffUntil` / `LastPullAt` 6 method | repository_pull_state CRUD (migration 000043), cold start 자동 upsert |
| `backend-core/internal/store/repository_pull_targets.go` | **X-5 production wire follow-up 신규** — PostgresStore 의 `ListGiteaPullTargets` + `GiteaPullTarget` type | repositories + integration_providers + repository_pull_state LEFT JOIN, Gitea SCM 한정 + backoff filter |
| `backend-core/internal/integrations/adapters/gitea_pull.go` | **X-5 1차 PR + follow-up 갱신** — GiteaClient 5 method + GiteaPullAdapter + stateToEventType helper | pr_activities.event_type enum 정합 (state="open"→"opened", state="closed"+merged=true→"merged", closed+!merged→"closed", fallback→"updated") |

**mirror 정책**: 19 file. **lint 영향**: L02 broken link 의 source code link 가 raw/ 에 존재 (mirror scope 내) → L02 PASS. L10 면제 불요 (raw/ 의 1:1 mirror).

**본 sprint 의 verification 으로 발각** (2026-06-13): 원본 Phase 1.5 의 12 file 화이트리스트 중 `keycloak_verifier.go` + `keycloak_admin_client.go` + `audit/middleware.go` + `rbac/policy_store.go` + `store/postgres/repository_ops.go` 의 경로 outdated (해당 경로에 file 부재). **fix**: §1.7.1 의 file list 의 정확한 경로 정합 (audit-ops/ + rbac-permissions/ + httpapi/ + store/ 의 실제 경로). **forward**: 새 backend file 추가 시 PR 본문에 mirror scope 추가 요청 (mirror-list.md §1.7.1 + script 의 화이트리스트 갱신).

**mirror 정책**: 25 file. **lint 영향**: L02 broken link 의 source code link 가 raw/ 에 존재 (mirror scope 내) → L02 PASS. L10 면제 불요 (raw/ 의 1:1 mirror).

**소스 경로**: `frontend/tests/e2e/`, `frontend/lib/`, `frontend/domain/`. **본 sprint 의 verification 으로 정확한 경로 정합** (2026-06-13).

| source path | 의의 | maintenance 디테일 |
|---|---|---|
| `frontend/tests/e2e/fixtures.ts` | loginAs + waitForSignInForm (PR #579 5+6차 commit + PR #580 통합) | waitForSignInForm default 30s→60s, loginAs 30s→60s, restart logic |
| `frontend/tests/e2e/voc-auto-routing.spec.ts` | **PR #579 신규** — TC-INBOUND-SRC-01 + NEG | beforeAll hook 으로 PATCH inbound_source 1회 처리 |
| `frontend/tests/e2e/signout.spec.ts` | PR #580 의 signout CI timeout 완화 | 타임아웃 상향 |
| `frontend/tests/e2e-manifests/smoke.txt` | PR #580 신규 — spec selection SSOT (smoke) | |
| `frontend/tests/e2e-manifests/quarantine.txt` | PR #580 신규 — spec selection SSOT (quarantine) | |
| `frontend/lib/store.ts` | Zustand store (NOW-4 frontend unit test 962 PASS) | auth actor + isSystemAdmin + state shape |
| `frontend/domain/auth-session/service/role-routing.ts` | RBAC routing (defaultLandingFor + isSystemAdmin + pathRequiresSystemAdmin) | system route gate |
| `frontend/components/admin/x1-widgets/SyncJobQueueWidget.tsx` | **X-1 2차 PR (2026-06-13) 신규** — RM-M4-07 widget 4종 1 (queued+running 큐, 10 row) | adminX1Service.listSyncJobs({status:queued\|running, limit:10}) 병렬 fetch |
| `frontend/components/admin/x1-widgets/SyncJobStatusWidget.tsx` | **X-1 2차 PR 신규** — widget 4종 2 (4 status 별 count grid) | adminX1Service.getStatusSummary() |
| `frontend/components/admin/x1-widgets/ProviderHealthWidget.tsx` | **X-1 2차 PR 신규** — widget 4종 3 (placeholder, API-107 별도 carve) | ADR-0032 §3 carve — v0.1.1 후속 sprint |
| `frontend/components/admin/x1-widgets/DashboardSummaryWidget.tsx` | **X-1 2차 PR 신규** — widget 4종 4 (totalJobs/queueDepth/failed/successRate) | successRate = succeeded/(succeeded+failed)*100, 소수점 1자리 |
| `frontend/domain/integration-registry/service/admin-x1.service.ts` | **X-1 2차 PR 신규** — adminX1Service class + listSyncJobs/GetSyncJob/getStatusSummary | apiClient<T> 정공법 (자동 token refresh + session death) |
| `frontend/domain/repository-integration/schema/repository-kpi.types.ts` | **Sprint A 신규** — RepositoryKPI type (quality_score + build_success_rate + open_pr_count + merged_pr_count + active_contributor_count) | GET /api/v1/repositories/:id/kpi 정합 |
| `frontend/domain/repository-integration/schema/repository-tests.types.ts` | **Sprint A 신규** — RepositoryTestResults type (totals + pass_rate + recent) | GET /api/v1/repositories/:id/test-results 정합 (build_runs 분포) |
| `frontend/domain/repository-integration/service/repository-kpi.service.ts` | **Sprint A 신규** — fetchRepositoryKPI (windowDays option) | apiClient.get<RepositoryKPIResponse> 정공법 |
| `frontend/domain/repository-integration/service/repository-tests.service.ts` | **Sprint A 신규** — fetchRepositoryTestResults (window + limit option) | apiClient.get<RepositoryTestResultsResponse> 정공법 |
| `frontend/domain/repository-integration/view/RepositoryKPISection.tsx` | **Sprint A 신규** — 4 card (Quality Score / Build Success Rate / Pull Requests / Active Contributors) + window selector | kpi-tests-per-domain-scope.md §2.1 Repository sub-section |
| `frontend/domain/repository-integration/view/RepositoryTestsSection.tsx` | **Sprint A 신규** — Pass Rate (Recharts 도넛) + Status Distribution (7 status) + Recent Runs table + window selector | kpi-tests-per-domain-scope.md §2.1 Repository sub-section |
| `frontend/components/admin/inbound-source-config/InboundSourceTypeSelector.tsx` | **X-2 5차 PR 신규** — provider_type select (Gitea/Jira/Other/Disabled 4 option) | 1차 출처 X-2 PR #586 의 InboundSourceType enum 정공법 |
| `frontend/components/admin/inbound-source-config/InboundSourceConfigEditor.tsx` | **X-2 5차 PR 신규** — JSONB textarea editor (parse error + save) | InboundSourceRoutingConfig struct 1:1 매핑 |
| `frontend/components/admin/inbound-source-config/PatternPreview.tsx` | **X-2 5차 PR 신규** — provider-specific pattern + custom regex 검증 (MATCH/NO MATCH) | auto_route.go 의 gitea/jira/github/gitlab regex 정공법 |
| `frontend/components/admin/inbound-source-config/InboundSourceManager.tsx` | **X-2 5차 PR 신규** — 4 widget 통합 view + platform selector + save + audit | system_admin /admin/inbound-source 진입 |
| `frontend/tests/e2e/admin-x2.spec.ts` | **X-2 5차 PR 신규** — TC-ADMIN-X2-01/02/03/04 (system_admin 진입 + 4 widget 렌더 + type selector 변경 + non-admin redirect + pattern preview 검증) | page.locator("#inbound-source-type-select") + textarea fill |

**mirror 정책**: 12 file (6 e2e/manifest + 1 Zustand store + 1 RBAC routing + 4 X-1 widget + 1 admin-x1 service + 1 X-1 e2e = 14 항목 중 12 core). **forward**: 새 frontend helper/page/component 추가 시 PR 본문에 mirror scope 추가 요청.

**본 sprint 의 verification 으로 발각** (2026-06-13): 원본 8 file 화이트리스트 중 `lib/auth/{tokenStore,apiClient,role-routing}.ts` 의 경로 outdated (해당 경로에 file 부재). **fix**: `frontend/lib/store.ts` (Zustand) + `frontend/domain/auth-session/service/role-routing.ts` (RBAC) 의 실제 경로 정합. **apiClient + tokenStore 는 frontend bulk source 의 mirror scope 외 (Phase 3 forward, 본 sprint 정공법의 follow-up)**.

### 1.7.3 Workflows + scripts (운영 critical, ~7 file)

**소스 경로**: `.github/workflows/*.yml`, `scripts/*.sh`.

| source path | 의의 | maintenance 디테일 |
|---|---|---|
| `.github/workflows/ci.yml` | CI fast required (PR #578 의 e2e-internal 폐기 반영) | 9 jobs (e2e shard 1/2/3 + backend + frontend + lint 4종) |
| `.github/workflows/e2e-regression.yml` | PR #580 신규 — non-quarantine full regression | |
| `.github/workflows/e2e-quarantine.yml` | PR #580 신규 — flaky/quarantine 전용 | |
| `scripts/wiki-sync-devhub.sh` | 본 wiki mirror script | BSD-rsync safe + manifest |
| `scripts/select-playwright-specs.sh` | PR #580 신규 — spec selection | |
| `scripts/ci-e2e-sync-check.sh` | CI sync check | DEVHUB_BUILD_TIER 의도적 미포함 |
| `scripts/check-migration-uniqueness.sh` | NOW-5 migration prefix CI guard | |

**mirror 정책**: 7 file.

### 1.7.4 Branch memory (sprint 한정, ~7 file)

**소스 경로**: `ai-workflow/memory/<agent>/<branch>/{state.json, session_handoff.md, work_backlog.md, backlog/YYYY-MM-DD.md, pr_body.md}`.

**mirror 정책**: 본 sprint 의 **active + 30일 이내 CLOSED** branch 만. 본 sprint 범위 = `feat/work_260612-7-v0-1-1-inbound-source-impl/` + `feat/work_260612-6-e2e-internal-removal/` + `codex/work_260612-579-ci-rearchitecture/` + `codex/work_260613-ci-retro-and-memory/`. **forward**: archive 시 (PR 머지 + 30일 후) `mavis-trash` 권장. **lint 영향**: frontmatter 형식만 검증 (raw/ 의 1:1 mirror, 본문 직접 합성 X).

### 1.7.5 Traceability + ID slot doc (link SSOT, ~5 file)

**소스 경로**: `docs/traceability/`.

| source path | 의의 | maintenance 디테일 |
|---|---|---|
| `docs/traceability/report.md` | traceability matrix (§1 REQ → §6 변경 이력) | REQ/ARCH/API/RM/IMPL/UT/TC ID + ADR 인덱스 |
| `docs/traceability/README.md` | traceability entry | workflow 정합 |
| `docs/traceability/conventions.md` | ID slot conventions (kebab-case module ID) | REQ-FR-NNN / ARCH-NN / API-NN / RM-{domain}-NN / IMPL-{module}-NN |
| `docs/traceability/sync-checklist.md` | PR traceability sync 절차 | 영향 단계 ID 발급/갱신 + matrix row + PR body |
| `docs/traceability/_archive/` | historical sync (옵션) | |

**mirror 정책**: 4 file (sync-checklist 필수).

## 1.8 Phase 3 (mass ingest, ~78 file) — **2026-06-13 신규**

**Phase 3 의 목적**: 위키의 LLM Wiki 의 mass ingest = **docs/domain + docs/architecture + docs/infrastructure + docs/validation** 의 ~78 file mirror + **30~50 wiki page 자동 생성** (concepts/entities/topics). **본 저장소 한정 + my_harness 측 wiki 일임 결정 (session §10) 해제** 후 진행.

**mirror 정책**: `find docs/{domain,architecture,infrastructure,validation} -type f -name "*.md"` 의 **78 file 모두** (Phase 1.5 의 `find` glob 동적 처리 정공법 정합).

### 1.8.1 Domain (~66 file, 10+ sub-directory)

**소스 경로**: `docs/domain/**/`

| sub-directory | file 수 | 비고 |
|---|---|---|
| `docs/domain/auth-session/` | ~10 | 인증/계정 도메인 (RELEASE Sprint, N-12 등) |
| `docs/domain/rbac-permissions/` | ~8 | RBAC 도메인 (ARCH-04..05, REQ-FR-87..) |
| `docs/domain/organization-management/` | ~6 | 조직 도메인 (REQ-FR-39..) |
| `docs/domain/application-lifecycle/` | ~8 | Application 도메인 (N-13 inbound_source 자동 routing) |
| `docs/domain/dev-request/` | ~5 | dev-request 도메인 (ADR-0028) |
| `docs/domain/onboarding/` | ~4 | onboarding 도메인 |
| `docs/domain/audit-ops/` | ~5 | audit-ops 도메인 (ADR-0030 정합) |
| `docs/domain/realtime/` | ~3 | realtime 도메인 |
| `docs/domain/auth-session-port/`, `auth-rbac/`, `auth-rbac-port/` 등 | ~17 | cross-cut 도메인 |

**mirror 정책**: `find docs/domain -type f -name "*.md"` 의 66 file 모두. **Phase 1.5 의 `find` glob 동적 처리** = script 의 §10 갱신.

### 1.8.2 Architecture (1 file)

**소스 경로**: `docs/architecture/`

| file | 비고 |
|---|---|
| `docs/architecture/README.md` | architecture 도메인 SSOT |

**mirror 정책**: 1 file.

### 1.8.3 Infrastructure (9 file)

**소스 경로**: `docs/infrastructure/`

| sub-directory | file 수 | 비고 |
|---|---|---|
| `docs/infrastructure/keycloak-idp/` | ~6 | Keycloak 운영 가이드 (single port + service account + refactor + offboarding + SSO federation + 변경 이력) |
| `docs/infrastructure/deployment-automation/` | ~2 | deployment 자동화 (single port reverse proxy + 변경 이력) |
| `docs/infrastructure/monitoring/` | ~1 | (현시점 부재, forward path) |

**mirror 정책**: 9 file 모두.

### 1.8.4 Validation (2 file)

**소스 경로**: `docs/validation/`

| file | 비고 |
|---|---|
| `docs/validation/N-10-manager-rbac.md` | N-10 RBAC 검증 보고서 (N-10 follow-up 의 검증 SOP) |
| `docs/validation/2026-06-12-n13-test2-rebase-verification.md` | N-13 follow-up B 의 검증 보고서 (PR #575) |

**mirror 정책**: 2 file 모두.

## 2. mirror 실행 정책 (script 의 source list, Phase 1.5+3 추가)

`scripts/wiki-sync-devhub.sh` 의 mirror 실행 시 다음 15 패턴으로 file 매칭 (Phase 1 = 7 패턴 + Phase 1.5 = 5 패턴 + Phase 3 = 3 패턴):

| # | 패턴 | mirror source | mirror target |
|---|---|---|---|
| 1 | ADR | `docs/adr/ADR-*.md` | `ai-workflow/wiki/raw/projects/devhub/docs/adr/ADR-*.md` |
| 2 | Governance | `docs/governance/*.md` | `ai-workflow/wiki/raw/projects/devhub/docs/governance/*.md` |
| 3 | Planning | `docs/planning/*.md` | `ai-workflow/wiki/raw/projects/devhub/docs/planning/*.md` |
| 4 | Setup | `docs/setup/*.md` | `ai-workflow/wiki/raw/projects/devhub/docs/setup/*.md` |
| 5 | Requirements | `docs/requirements.md` | `ai-workflow/wiki/raw/projects/devhub/docs/requirements.md` |
| 6 | OpenAPI | `docs/openapi.yaml` | `ai-workflow/wiki/raw/projects/devhub/docs/openapi.yaml` |
| 7 | AI-workflow memory (main flat) | `ai-workflow/memory/{state.json, session_handoff.md, work_backlog.md}` | `ai-workflow/wiki/raw/projects/devhub/ai-workflow-memory/{state.json, session_handoff.md, work_backlog.md}` |
| **8** | **Workflows (Phase 1.5)** | **`.github/workflows/*.yml`** | **`ai-workflow/wiki/raw/projects/devhub/.github/workflows/*.yml`** |
| **9** | **Scripts (Phase 1.5)** | **`scripts/*.sh` (mirror list 화이트리스트)** | **`ai-workflow/wiki/raw/projects/devhub/scripts/*.sh`** |
| **10** | **Backend critical Go (Phase 1.5)** | **`backend-core/internal/{auth,domain,httpapi,audit,rbac,store,sso-integrations}/**/*.go` (mirror list 화이트리스트)** | **`ai-workflow/wiki/raw/projects/devhub/backend-core/internal/...`** |
| **11** | **Frontend e2e critical (Phase 1.5)** | **`frontend/tests/e2e/{fixtures,signout}.ts` + `frontend/tests/e2e/voc-*.spec.ts` + `frontend/tests/e2e-manifests/*.txt` + `frontend/lib/auth/{tokenStore,apiClient,role-routing}.ts`** | **`ai-workflow/wiki/raw/projects/devhub/frontend/...`** |
| **12** | **Traceability (Phase 1.5)** | **`docs/traceability/{README,conventions,report,sync-checklist}.md`** | **`ai-workflow/wiki/raw/projects/devhub/docs/traceability/...`** |
| 13 | **Branch memory (Phase 1.5, optional)** | **`ai-workflow/memory/<agent>/<branch>/{state,session_handoff,work_backlog}.{json,md}` (active + 30일 이내 CLOSED)** | **`ai-workflow/wiki/raw/projects/devhub/ai-workflow/memory/<agent>/<branch>/...`** |
| **14** | **Domain (Phase 3, mass ingest)** | **`docs/domain/**/[*].md`** | **`ai-workflow/wiki/raw/projects/devhub/docs/domain/...`** |
| **15** | **Architecture + Infrastructure + Validation (Phase 3, mass ingest)** | **`docs/{architecture,infrastructure,validation}/[*].md`** | **`ai-workflow/wiki/raw/projects/devhub/docs/{architecture,infrastructure,validation}/...`** |

**mirror size 추정** (2026-06-13 main HEAD `32454fc` 기준):
- Phase 1: 85 file (≈ 3.5MB)
- Phase 1.5: ~55 file (≈ 1.5MB)
- Phase 3: ~78 file (≈ 2.0MB)
- **합 ≈ 7MB, ~220 file**

**Phase 3 mass ingest 정공법 (2026-06-13 추가)**:
- `find docs/{domain,architecture,infrastructure,validation} -type f -name "*.md"` 의 78 file 모두 mirror (script 의 §10/§11/§12/§13/§14/§15 갱신)
- mirror 실행 후 wiki page 자동 생성 (concepts/entities/topics) — `scripts/wiki-mass-ingest.sh` (forward path)
- 위키의 30~50 page 자동 작성 (frontmatter + raw body 1:1)

**제외 패턴** (mirror 미실시, Phase 1 + 1.5 + 3):
- 빌드 산출물: `target/`, `backend-core/main`, `frontend/.next/`, `playwright-report/`, `test-results/`, `dist/`, `build/`, `__pycache__/`, `node_modules/`
- VCS + IDE: `.git/`, `.idea/`, `.vscode/`, `.DS_Store`
- Backend source bulk: `backend-core/cmd/`, `backend-core/migrations/`, `backend-core/test/` (생성물, 위키 정합 불요)
- Frontend source bulk: `frontend/src/`, `frontend/app/`, `frontend/components/`, `frontend/node_modules/`, `frontend/.next/`, `frontend/dist/`
- Archive: `infra/idp/_archive_*/` (immutable archive)
- Public Wiki (기존): `docs/wiki/` (대외 공개용, LLM Wiki 와 cross-link 없음)
- LLM Wiki (본 Phase): `docs/llm-wiki/` (mirror 미필요, source-of-truth)
- Lint/scratch: `_lint/`, `scratch/`, `playwright-report/`

## 3. lint 영향 (Phase 1.5+3 갱신)

| L rule | 영향 | DevHub 적용 (Phase 1.5+3) |
|---|---|---|
| L01 | frontmatter 누락 | wiki page 만, raw/ 적용 X |
| L02 | broken wiki link | **Phase 1.5+3 갱신**: source code + domain/architecture/infrastructure/validation 의 raw/ link 가 mirror scope 내 → wiki page 의 source link 정상 (L02 PASS). |
| L03 | 고아 페이지 | wiki page 만, raw/ 적용 X (변동 X) |
| L04 | 중복 페이지 | wiki page 만, raw/ 적용 X |
| L05 | stale (90일+) | wiki page 만, raw/ 적용 X |
| L06 | sources: 경로 부재 | wiki page 만, raw/ 적용 X |
| L07 | 모순 (같은 title 두 페이지) | DevHub ADR-*.md 면제 (변동 X) |
| L08 | index.md 미등록 wiki 페이지 | **Phase 3 갱신**: Phase 3 의 wiki page 작성 + L08 fix 자동 (`wiki-source-sync --auto-fix`) |
| L09 | log.md 1주일+ 미갱신 | out-of-repo, my_harness 측 관리 |
| L10 | raw/ source 0 | **Phase 1.5+3 갱신**: Phase 1.5+3 mirror scope 내 file 들의 wiki page 의 raw/ source = 1:1 mirror → L10 PASS |

## 4. forward path (Phase 1 + 1.5 + 3 + N)

| 단계 | mirror list | 정공법 |
|---|---|---|
| **Phase 1 (기존)** | docs subset 85 file | 본 문서 §1.1~1.6 (2026-06-10 작성) |
| **Phase 1.5 (본 갱신)** | source code + workflow + scripts + branch memory + traceability, **~55 file 추가** | 본 문서 §1.7 (2026-06-13 추가) |
| **Phase 3 (본 갱신)** | domain (66) + architecture (1) + infrastructure (9) + validation (2), **~78 file 추가** | 본 문서 §1.8 (2026-06-13 추가) |
| **Phase N (forward)** | ai-workflow memory 의 sprint branch 별 mirror (본 Phase 1.5 의 active + 30일 이내 CLOSED branch 정공법 정합) | `scripts/wiki-sync-devhub.sh` 의 `--branch <branch>` 옵션 (선택) |

## 5. 향후 작업 지침 (wiki-only maintenance 정공법, 2026-06-13 추가)

**위키만으로 코드 maintenance 가능** 하도록 하기 위해, **모든 신규 PR**은 다음을 충족해야 한다:

1. **소스코드 변경 PR**: 본 mirror-list.md §1.7.1 / §1.7.2 / §1.7.3 / §1.7.5 의 mirror scope 내 file 변경 시, **PR 본문 + branch memory 의 pr_body.md**에 다음 명시:
   - 변경 file path
   - 변경 line / block summary (1-2 line)
   - cross-reference (이전 PR / ADR / ID)
2. **신규 file PR**: 본 §1.7 의 mirror scope 가 확장될 가능성 있는 신규 file (예: backend 새 도메인 / frontend 새 e2e spec / workflow 신규) 추가 시, **PR 본문 + pr_body.md**에 다음 명시:
   - 신규 file path
   - wiki mirror scope 추가 요청 (mirror-list.md §1.7 갱신 + lint-config.toml 갱신)
3. **위키 본문 갱신**: PR 머지 후 `wiki-source-sync` skill 의 **op=commit** 호출 (본 저장소 metadata 정합). 본문 1:1 mirror 정공법은 operation-sop.md §3 참조.

**AGENTS.md 의 "문서 tier 라벨" + 본 §5 의 지침 정합** — 향후 작업의 디폴트 위키 정합 보장.

## 6. 다음 세션 directive

1. `docs/llm-wiki/lint-config.toml` 갱신 (L02 broken link Phase 1.5 갱신 + L10 Phase 1.5 갱신).
2. `scripts/wiki-sync-devhub.sh` 갱신 (Phase 1.5 mirror pattern 6개 추가, 화이트리스트).
3. `docs/llm-wiki/operation-sop.md` 갱신 (sources/ page 본문 1:1 mirror 정공법 + Phase 1.5 SOP).
4. 본 sprint 의 22 file mirror 실행 (scripts re-run).
5. lint 검증 (4종: Detect Changed Paths / Workflow Lint / Migration Prefix / OpenAPI YAML).
6. wiki-source-sync + wiki-event-sync 호출 (raw → wiki 갱신).
7. commit + push + PR 발행 (Phase 1.5 scope).
8. main flat memory finalize (post-merge sync).

## 1. Phase 1 source list (core subset ~80 file)

### 1.1 Architecture Decision Records (ADR) — 31 file

**소스 경로**: `docs/adr/*.md` (ADR-0001..ADR-0031)

| ADR ID | 제목 | mirror target |
| --- | --- | --- |
| ADR-0001 | idp-selection | `ai-workflow/wiki/raw/projects/devhub/docs/adr/0001-idp-selection.md` |
| ADR-0002 | rbac-policy-edit-api | `ai-workflow/wiki/raw/projects/devhub/docs/adr/0002-rbac-policy-edit-api.md` |
| ... | ... | ... |
| ADR-0030 | sso-integrations-and-auth-session-port | `ai-workflow/wiki/raw/projects/devhub/docs/adr/0030-sso-integrations-and-auth-session-port.md` |
| ADR-0031 | build-tag-policy-review | `ai-workflow/wiki/raw/projects/devhub/docs/adr/0031-build-tag-policy-review.md` |

**mirror 정책**: 31 file 모두 mirror (의미 있는 SSOT 결정). `infra/idp/_archive_*/` 의 immutable archive 미포함 (sprint -a follow-up PR #540 의 immutable archive 결정 정합).

**lint 영향**: ADR-*.md 는 L07 (모순) 의 의도적 supersede (예: ADR-0030 → ADR-0031) 가 false positive 가능 → `lint-config.toml` 의 `[rules.L07].skip_paths = ["wiki/projects/devhub/sources/ADR-*.md"]` 면제 (Phase 1 의 lint config).

### 1.2 Governance — 5 file

**소스 경로**: `docs/governance/*.md`

| file | mirror target |
| --- | --- |
| `code-taxonomy.md` | `ai-workflow/wiki/raw/projects/devhub/docs/governance/code-taxonomy.md` |
| `document-standards.md` | `ai-workflow/wiki/raw/projects/devhub/docs/governance/document-standards.md` |
| `keycloak_admin_responsibility.md` | `ai-workflow/wiki/raw/projects/devhub/docs/governance/keycloak_admin_responsibility.md` |
| `README.md` | `ai-workflow/wiki/raw/projects/devhub/docs/governance/README.md` |
| `worker_division.md` | `ai-workflow/wiki/raw/projects/devhub/docs/governance/worker_division.md` |

**mirror 정책**: 5 file 모두 mirror. 단 `worker_division.md` 의 사내 한정 정보 (DEVHUB_KEYCLOAK_*, internal-registry, 172.16.0.0/12) **포함되지만** D-72 응답 §3 + yklee 결정으로 `sa-internal/` 격리 불요 + GitHub Wiki 가 아닌 in-repo (v0.7.17+) 만 push 이므로 mirror 허용.

### 1.3 Planning — 27 file

**소스 경로**: `docs/planning/*.md`

**mirror 정책**: 27 file 모두 mirror. 단 `infra/idp/_archive_*/` 의 immutable archive 미포함 (ADR-0001/0009 cross-ref 가능 archive 는 별도 위치).

**file list (자동 script 로 동적, 2026-06-11 main HEAD `f879b89` 기준, `find docs/planning -type f -name "*.md" | wc -l` = 27)**:

| file | 비고 |
|---|---|
| `2026-06-11-p2-residual-sprint-plan.md` | PR #560 의 main 정합 (P2 잔여 5건 일괄 처리 정공법) — 06-11 main 머지 완료 |
| `2026-06-12-inbound-source-routing-sprint-plan.md` | PR #547 의 N-13 housekeeping 정합 |
| `api-key-management-sprint-plan.md` | |
| `application_management_hotfix_2026-05-27.md` | |
| `external-integrations-agentic-rag-roadmap.md` | |
| `integrated_test_report_20260601.md` | |
| `integrated_test_scenarios.md` | |
| `keycloak_event_audit_integration.md` | |
| `migration_baseline_reset_plan_2026-06-04.md` | |
| `ops_ui_transition_plan.md` | |
| `project_creation_dreq_notification_concept.md` | |
| `project_operating_model_example_2026.md` | |
| `project_operating_model_template.md` | |
| `project_repository_creation_linking_plan_2026-05-27.md` | |
| `rbac-hardening-implementation-readiness-20260602.md` | |
| `release_v0-1_roadmap.md` | |
| `role-access-concept.md` | |
| `sprint-plan-20260601.md` | |
| `system_admin_catalog_plan_2026-05-27.md` | |
| `system_erd.md` | |
| `system_usecases.md` | |
| `test-findings-and-rbac-hardening-20260602.md` | |
| `ui_app_project_repo_upgrade_plan.md` | |
| `ui_e2e_followup_after_merge.md` | |
| `view_menu_screen_api_matrix.md` | |
| `ws_subprotocol_vs_ticket_poc.md` | |

**27 file 동적 mirror** (`scripts/wiki-sync-devhub.sh` 의 `find docs/planning -type f -name "*.md"` glob).

### 1.4 Setup — 15 file

**소스 경로**: `docs/setup/*.md`

**mirror 정책**: 15 file 모두 mirror. setup 의 운영 SOP (test-server-deployment, single-port-deployment, docker-packaging-deployment-guide 등) 가 LLM agent 의 RAG source 로 가치 높음. **정합 (2026-06-11 본 sprint 검증)**: code `find docs/setup -maxdepth 1 -type f -name "*.md" | wc -l` = 15, vault raw `find ai-workflow/wiki/raw/projects/devhub/docs/setup -type f | wc -l` = 15 — **drift 0** (mirror script 자동 정공법).

**file list (2026-06-11 main HEAD `f879b89` 기준, `find docs/setup -maxdepth 1 -type f -name "*.md" | wc -l` = 15)**:

| file | 비고 |
|---|---|
| `api_key_rotation.md` | |
| `deploy_preflight_checklist.md` | |
| `docker-packaging-deployment-guide.md` | |
| `e2e-test-guide.md` | |
| `environment-setup.md` | |
| `homelab_agent_token_rotation.md` | |
| `hrdb_unit_pre_stage.md` | |
| `internal_network_constraints.md` | |
| `jwks_rotation_cache_flush.md` | |
| `keycloak_operations.md` | |
| `migration_000021_conflict_resolution.md` | |
| `onboarding_operations.md` | |
| `prometheus_alertmanager_setup.md` | |
| `single_port_deployment.md` | |
| `test-server-deployment.md` | |

**15 file 동적 mirror** (정합 — code 와 raw 모두 15 file, drift 0).

### 1.5 Requirements + OpenAPI — 2 file

**소스 경로**:
- `docs/requirements.md` (DevHub 의 REQ SSOT)
- `docs/openapi.yaml` (DevHub 의 API contract)

**mirror 정책**: 2 file 모두 mirror. `openapi.yaml` 의 경로 = 81, schema = 78. **단 `openapi.yaml` 의 경로 (e.g. `internal-registry.example.com`) 가 사내 한정 정보 포함 가능** — D-72 응답 §3 의 lint L11 (사내 패턴 검출) 으로 자동 검출 권장 (D-73 작업, my_harness 측).

### 1.6 AI-workflow memory (main flat) — 3 file

**소스 경로**:
- `ai-workflow/memory/state.json` (head_commit + status)
- `ai-workflow/memory/session_handoff.md` (post-session handoff)
- `ai-workflow/memory/work_backlog.md` (변경 이력)

**mirror 정책**: 3 file 모두 mirror (main flat 만 — sprint branch 의 memory 는 본 Phase 1 의 scope 외, my_harness 의 wiki-sync-ai-workflow.sh 와 동일 pattern).

**본 저장소 의 main flat memory 의 위치**:
- `state.json` 의 `head_commit` = `76bb00580` (2026-06-11, main HEAD = PR #560 squash, 최신 PR = #514~#560)
- `session_handoff.md` 의 §0/§11 (PR #514~#560 row, main flat memory finalize)
- `work_backlog.md` 의 §0 status line + §5 변경 이력 (PR #540~#560 row)

## 2. Phase 1 scope 외 (Phase 3 mass ingest, 별도 PR)

### 2.1 Domain (66 file)

**소스 경로**: `docs/domain/**/*.md`

| sub-directory | file 수 | 비고 |
| --- | --- | --- |
| `docs/domain/auth-session/` | ~10 | 인증/계정 도메인 (RELEASE Sprint, N-12 등) |
| `docs/domain/rbac-permissions/` | ~8 | RBAC 도메인 (ARCH-04..05, REQ-FR-87..) |
| `docs/domain/organization-management/` | ~6 | 조직 도메인 (REQ-FR-39..) |
| `docs/domain/application-lifecycle/` | ~8 | Application 도메인 |
| `docs/domain/dev-request/` | ~5 | dev-request 도메인 (ADR-0028) |
| `docs/domain/onboarding/` | ~4 | onboarding 도메인 |
| `docs/domain/audit-ops/` | ~5 | audit-ops 도메인 (ADR-0030 정합) |
| `docs/domain/realtime/` | ~3 | realtime 도메인 |
| `docs/domain/auth-session-port/`, `auth-rbac/`, `auth-rbac-port/` 등 | ~17 | cross-cut 도메인 |

**Phase 3 의 mirror list** — `find docs/domain -type f -name "*.md"` 의 66 file 모두.

### 2.2 Architecture (1 file)

**소스 경로**: `docs/architecture/DETAILED_DESIGN_*.md` (총 7+ file)

**Phase 3 의 mirror list** — `find docs/architecture -type f -name "*.md"` 의 1 file.

### 2.3 Infrastructure (variable)

**소스 경로**: `docs/infrastructure/**` (sub: keycloak-idp/, integration/, monitoring/, ops/, etc.)

**Phase 3 의 mirror list** — `find docs/infrastructure -type f -name "*.md"` 의 variable file (현시점 count = `infrastructure/README.md` 만, sub-directory 별도 확인 필요).

### 2.4 Validation (1 file)

**소스 경로**: `docs/validation/N-*.md`

**Phase 3 의 mirror list** — N-12 (voc + notification) 등.

## 3. mirror 실행 정책 (script 의 source list)

`scripts/wiki-sync-devhub.sh` 의 mirror 실행 시 다음 4 패턴으로 file 매칭:

| 패턴 | mirror source | mirror target |
| --- | --- | --- |
| ADR | `docs/adr/ADR-*.md` | `ai-workflow/wiki/raw/projects/devhub/docs/adr/ADR-*.md` |
| Governance | `docs/governance/*.md` | `ai-workflow/wiki/raw/projects/devhub/docs/governance/*.md` |
| Planning | `docs/planning/*.md` | `ai-workflow/wiki/raw/projects/devhub/docs/planning/*.md` |
| Setup | `docs/setup/*.md` | `ai-workflow/wiki/raw/projects/devhub/docs/setup/*.md` |
| Requirements | `docs/requirements.md` | `ai-workflow/wiki/raw/projects/devhub/docs/requirements.md` |
| OpenAPI | `docs/openapi.yaml` | `ai-workflow/wiki/raw/projects/devhub/docs/openapi.yaml` |
| AI-workflow memory (main flat) | `ai-workflow/memory/{state.json, session_handoff.md, work_backlog.md}` | `ai-workflow/wiki/raw/projects/devhub/ai-workflow-memory/{state.json, session_handoff.md, work_backlog.md}` |

**제외 패턴** (mirror 미실시):
- 빌드 산출물: `target/`, `backend-core/main`, `frontend/.next/`, `playwright-report/`, `test-results/`, `dist/`, `build/`, `__pycache__/`, `node_modules/`
- VCS + IDE: `.git/`, `.idea/`, `.vscode/`, `.DS_Store`
- Backend runtime: `backend-core/`, `frontend/`, `backend-ai/` 의 source code (wiki 의 RAG source 가 아님, 코드 정합은 `ai-workflow/memory/{code-index, ...}.md` 등 별도)
- Archive: `infra/idp/_archive_*/` (immutable archive, wiki 정합 불요)
- Public Wiki (기존): `docs/wiki/` (대외 공개용, LLM Wiki 와 cross-link 없음)
- LLM Wiki (본 Phase 1): `docs/llm-wiki/` (본 Phase 1 의 source-of-truth, mirror 미필요)
- Lint/scratch: `_lint/`, `scratch/`, `playwright-report/`

**mirror size 추정** (2026-06-11 main HEAD `f879b89` 기준, code 1:1 mirror):
- ADR: 31 file (≈ 700KB)
- Governance: 5 file (≈ 100KB)
- Planning: 27 file (≈ 1.5MB) — PR #560 의 `2026-06-11-p2-residual-sprint-plan.md` 포함
- Setup: 15 file (≈ 800KB) — code 와 raw 모두 15 file 일치
- Requirements: 1 file (≈ 50KB)
- OpenAPI: 1 file (≈ 300KB)
- AI-workflow memory: 3 file (≈ 50KB)
- **합 ≈ 3.5MB, 83 file** (Planning 26 → 27 갱신, 총 83 file 유지)

## 4. lint 영향

**Phase 1 mirror 실행 후 wiki-lint 의 L01~L10 검증** (D-73 wiki-lint `--project` 옵션 활성 후):

| L rule | 영향 | DevHub 적용 |
| --- | --- | --- |
| L01 | frontmatter 누락 | raw/ file 은 wiki page 아님, 적용 X |
| L02 | broken wiki link | wiki page 만, raw/ 적용 X |
| L03 | 고아 페이지 | wiki page 만, raw/ 적용 X |
| L04 | 중복 페이지 | wiki page 만, raw/ 적용 X |
| L05 | stale (90일+) | wiki page 만, raw/ 적용 X |
| L06 | sources: 경로 부재 | wiki page 만, raw/ 적용 X |
| L07 | 모순 (같은 title 두 페이지) | **DevHub ADR-*.md 면제** (lint-config.toml) |
| L08 | index.md 미등록 wiki 페이지 | Phase 3 의 wiki page 작성 후 해소 |
| L09 | log.md 1주일+ 미갱신 | out-of-repo, my_harness 측 관리 |
| L10 | raw/ source 0 | wiki page 의 raw mirror source 부재, Phase 3 의 wiki page 작성 후 해소 |

**Phase 1 의 lint 검증 = 본 PR scope 외 (mirror 실행 후 사용자 confirm 후)**.

## 5. forward path

| 단계 | mirror list | 정공법 |
| --- | --- | --- |
| **Phase 1 (본 PR)** | core subset ~80 file (ADR/governance/planning/setup/requirements/openapi/ai-workflow memory) | `docs/llm-wiki/mirror-list.md` + `scripts/wiki-sync-devhub.sh` |
| **Phase 3 (별도 PR)** | domain (66) + architecture (1) + infrastructure + validation (~100 file) | `docs/llm-wiki/mirror-list-phase-3.md` (별도 작성) + `scripts/wiki-sync-devhub.sh` 의 `--phase 3` 옵션 추가 (선택) |
| **Phase N (forward)** | ai-workflow memory 의 sprint branch 별 mirror (본 Phase 1 의 main flat 외) | `scripts/wiki-sync-devhub.sh` 의 `--branch <branch>` 옵션 (선택) |

## 6. 다음 세션 directive

1. `docs/llm-wiki/lint-config.toml` 작성 (L07 ADR 면제 + lint L11/L05/L09 의 DevHub 적용 정책).
2. `docs/llm-wiki/operation-sop.md` 작성 (sync + lint SOP + forward path + 위험).
3. `scripts/wiki-sync-devhub.sh` 작성 (BSD-rsync safe, dry-run + vault-absent no-op + mirror list 의 source list 동적 + manifest 자동).
4. sprint memory 5 file + main flat memory 3 file.
5. lint 검증 (4종) + script smoke test (dry-run mode).
6. commit + push + PR 발행.
7. main flat memory finalize (post-merge sync).
