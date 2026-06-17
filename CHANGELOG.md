# Changelog

All notable changes to DevHub will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.1.1-alpha] - 2026-06-11

**v0.1.1-alpha release. 잔여 5 (T-d-72-5/6 + D-73/74 + X-1~8) 의 v0.1.1-alpha 격하 정공법.**

main HEAD: `356d08b7` (v0.1.0-alpha release 정합) + tag `v0.1.1-alpha` (re-tag 후 main HEAD 부착, v0.1.0-alpha release 의 follow-up patch).

### 8 item v0.1.1-alpha 격하 (v0.1.0-alpha → v0.1.1-alpha, patch release)

| ID | 항목 | 의존 | v0.1.1-alpha status |
|---|---|---|---|
| **T-d-72-5** | wiki/cross/ cross-project 종합 페이지 (1~3 page) | T-d-72-4 ✅ | ⏳ planned (v0.1.1-alpha) |
| **T-d-72-6** | wiki-lint CI integration (`ci.yml` 의 e2e shard 또는 별도 lint job) | T-d-72-3 | ⏳ planned (v0.1.1-alpha) |
| **D-73** | my_harness 측 wiki-lint `--project` + `--project-config` 옵션 | (D-72-2 ✅) | ⏳ planned (v0.1.1-alpha) |
| **D-74** | `_lint/devhub/` per-project lint report 디렉터리 셋업 | T-d-72-3 | ⏳ planned (v0.1.1-alpha) |
| **T-d-79-5** | D-79 wiki-lint 통합 (L01~L10) | D-74 | ⏳ planned (v0.1.1-alpha) |
| **T-d-80-7** | D-80 wiki-lint 통합 (L01~L10) | D-74 | ⏳ planned (v0.1.1-alpha) |
| **X-1** | System Admin 운영 대시보드 (RM-M4-07) | — | ✅ implemented (2026-06-13, sprint `feat/work_260614-x1-system-admin-dashboard` 1차 PR #583 + sprint `feat/work_260614-x1-frontend-e2e` 2차 PR — backend IntegrationRepository method 3 + httpapi admin endpoint 3 + openapi paths 3 + frontend widget 4 + admin landing page 강화 + e2e admin-x1.spec.ts 3 case + ADR-0032) |
| **X-2** | inbound webhook 정규화 깊이 (multi-provider sync 일반화) | T-d-79-2/T-d-80-2 | ✅ implemented (2026-06-13, sprint `feat/work_260614-x2-system-admin-dashboard` 1차 PR #586 + sprint `feat/work_260614-x2-webhook-adapter` 2차 PR #587 + sprint `feat/work_260614-x2-jira-generic-adapter` 3차 PR #588 + sprint `feat/work_260614-x2-openapi-frontend` 4차 PR #589 + sprint `feat/work_260614-x2-frontend-e2e` 5차 PR — backend auto_route multi-provider pattern + WebhookAdapter interface + Gitea/Jira/Generic adapters + openapi schema + frontend multi-provider 운영 UI 4 widget + e2e admin-x2 4 case + ADR-0033) |
| **X-2** | inbound webhook 정규화 깊이 (multi-provider sync 일반화) | T-d-79-2/T-d-80-2 | ⏳ planned (v0.1.1-alpha) |
| **X-3** | 평문 secret envelope 암호화 (DEK + 키관리 ADR) | ADR | ⏳ planned (v0.1.1-alpha) |
| **X-4** | Phase D — project 생성 flow ↔ SCM create 연계 | T-d-72-5/6 | ⏳ planned (v0.1.1-alpha) |
| **X-5** | Gitea Hourly Pull 정밀화 (RM-M4-06 잔여, issue #231) | — | ✅ implemented (production wire, 2026-06-15 sprint `feat/x5-gitea-pull-store-wire`) — 1차 PR #592 + a49a5660 fix (cron worker + interface + metric + audit) + 본 follow-up PR (RepositoryPullStore 9 method + ListGiteaPullTargets + migration 000045 + adapter stateToEventType + main.go production wire) |
| **KPI/Tests 위치 정공법 (Sprint A)** | Repository 상세에 KPI/Tests sub-section 통합 — kpi-tests-per-domain-scope.md 의 1차 진입 | — | ✅ implemented (sprint `feat/x-repository-kpi-tests-section`, PR #597, 2026-06-15) — 2 endpoint (GET /api/v1/repositories/:id/kpi + /test-results) + 2 frontend component (RepositoryKPISection + RepositoryTestsSection) + ManagerView 통합 + openapi.yaml +86 path + backend handler +234 line. PR #597 의 codex review 2 P1 (lucide-react import / routePermissionTable) + 2 P2 (window parsing / build-runs filter) 동시 fix. Tier: 공용. 후속: Sprint B (Project 가중치) / Sprint C (Platform sub-rollup) / Sprint D (Sidebar 재구성) / Sprint E (legacy 결정) + e2e spec 별도 PR. **N-9 잔여 polish (PR #633, sprint `chore/260617-n9-residual-build-runs-polish`)**: (3) custom hook + (4) dashboard widget frontend 2건 sub-issue 정공법. `useRepositoryBuildRuns` (TanStack Query 도입 안함, useState+useEffect) + `RepositoryBuildRunsSection` widget + ManagerView sibling 통합. **PR #633 의 codex P2 review 코멘트 2건 정공법 fix (동일 branch follow-up commit)**: (P2-1) backend `store.ListRepositoryBuildRuns` 의 3-tuple total 을 frontend 가 합성하지 않고 `service.getRepositoryBuildRunsWithMeta` (legacy `getRepositoryBuildRuns` caller 무손상 보존) 로 expose, hook 이 `meta.total` 사용 (fallback = `items.length >= pageSize`) — `hasMore` 정확. (P2-2) hook 의 cleanup boolean + ref-reset race → `AbortController` per-request token + `isCurrent` 가드 (ref 일치 + `signal.aborted`) + service 옵션 `signal` → `apiClient` plumb. `apiClient` 가 4번째 인자 `options.signal?: AbortSignal` 추가 (backward-compatible) + refresh 후 2nd fetch 동일 signal 유지. 검증: hook 9/9 + section 7/7 + service 13/13 + apiClient 17/17 + RepositoryDashboardView 8/8 PASS, `tsc --noEmit` PR 영향 7 file 모두 0. Tier: 사외. |
| **N-9 잔여 build-runs polish backend sub-1 + sub-2 (test-only)** | Repository build-runs endpoint 의 backend 2건 (sub-1 #556 RBAC 403/404 가드 + sub-2 #557 Histogram metric) 정공법 검증 test 추가 | — | ✅ implemented (sprint `chore/260617-n9-sub1-2-backend-validation`, 2026-06-17) — backend 의 `GetRepositoryByID` 404 not_found 가드 (handler line 141-148) + `routePermissionTable` 의 `ResourcePlatformRepositories, ActionView` 매핑 (line 241, 403 unauthorized middleware 자동 가드) + `metrics.go` 의 `devhub_repository_build_runs_query_duration_seconds{status_filter}` 정의 + `repository_ops.go:177` 의 observe 호출 — 모두 PR #555 (`feat: N-2 draft-publish UT + P1-6 Keycloak revocation + N-3 E2E SCM flow`, 2026-05-30) 의 main 정합 시점에 이미 구현 완료. 단 **assertion test 0건** (sub-2 Histogram label + value 정공법 검증 부재) + **routePermissionTable 회귀 test 부재** (sub-1 RBAC build-runs 매핑) 의 검증 사각지대를 본 follow-up 으로 정공법 검증. (1) `backend-core/internal/httpapi/repository_ops_test.go` (MODIFY, +91 line) — `TestRepositoryBuildRuns_HistogramObserved_WithStatusFilter` + `TestRepositoryBuildRuns_HistogramObserved_NoStatusFilter` 신규 2 case. `dto.Metric` 직접 read 로 `histogram.SampleCount >= 1` + `histogram.SampleSum >= 0` 정공법 검증 (`testutil.ToFloat64` 는 Histogram 미지원 — counter/gauge/untyped only). (2) `backend-core/internal/httpapi/permissions_test.go` (MODIFY, +6 line) — `TestEnforceRoutePermission_RoleAllowedAndDenied` 의 build-runs 회귀 row 2개 추가 (developer pass + guest deny-by-default). (3) `backend-core/go.mod` (MODIFY, +1 line) — `kylelemons/godebug v1.1.0 // indirect` (transitive). 검증: build-runs 6/6 PASS (기존 4 + 신규 2) + permissions_test 4/4 PASS (기존 2 sub-test + 신규 build-runs 2 row) + backend `go build ./...` PASS + httpapi+rbac-permissions 3 packages PASS + `go vet ./internal/httpapi/` silent. 신규 ID 발급 0 (test-only, REQ/ARCH/API/IMPL/UT/TC ID slot 무관). issue #556 + #557 close 정공법. Tier: 공용 (backend test only + docs, 사내 한정 정보 0). |
| **dev-request intake IP default fix (이슈 1)** | dev 환경 dev-request intake token 발급 시 `Allowed IPs` default 정공법 (docker / colima / WSL / load-balancer 등 다양한 host IP 미스매치 해결) | — | ✅ implemented (sprint `fix/260617-dev-request-intake-ip-default`, 2026-06-17) — 테스트 환경에서 dev-request intake UI 의 요청 발송 시 401 `auth_intake_ip_denied` client IP 에러 발생 보고. backend 의 `clientIPAllowed` 정공법 (`TestClientIPAllowed_*` 5 case main 정합, deny-by-default + CIDR + single-IP + invalid-skip 정공법) + `routePermissionTable` 의 `ResourceDevRequestIntakeTokens, ActionCreate` (admin RBAC 한정) 가 정공법. **fix** (1) `frontend/components/dev-request/IssueIntakeTokenModal.tsx` (MODIFY, +8/-1) — `useState<string[]>([""])` → `useState<string[]>(["0.0.0.0/0", "::/0"])` (전체 IPv4 + IPv6 default). admin RBAC 한정 endpoint 이므로 risk 낮음. 운영 환경 발급 시 helper text 의 "운영 환경에서는 반드시 IP 좁히기" 경고에 따라 manual 좁히기. (2) `frontend/app/(dashboard)/admin/reception-test/page.tsx` (MODIFY, +9) — `handleIntakeSubmit` 의 catch 블록에서 `res.status === 401 && body?.code === "auth_intake_ip_denied"` 감지 시 안내 toast 표시. (3) `frontend/components/dev-request/__tests__/IssueIntakeTokenModal.test.tsx` (MODIFY, +21/-2) — 기존 `submits form` 의 allowed_ips 가 `["1.1.1.1", "::/0"]` (default 의 ::/0 유지) + `prevents ESC key` 의 불필요한 `user.type` 제거 + 신규 `default allowed_ips = 0.0.0.0/0 + ::/0` 1 case. 검증: 4/4 PASS + `tsc --noEmit` 본 PR 영향 0 + backend `go build ./...` silent. **Tier**: **사외** (frontend only + docs, 사내 한정 정보 0). backend 변경 0. |
| **KPI/test 카드 graceful fallback + retry (이슈 2)** | KPI/test 5 endpoint 의 5xx/503 응답에 대한 frontend graceful fallback + retry button | — | ✅ implemented (sprint `fix/260617-kpi-test-card-graceful-fallback`, 2026-06-17) — 테스트 환경에서 플랫폼 / repository / project 페이지의 KPI/test 카드가 서버 오류로 경고 창 표시 보고. **원인**: backend 의 `platformStoreOrUnavailable` helper (router.go:753) 가 nil store 시 503 + body `{status: "unavailable", error: "..."}` 반환하나 machine-readable `code` field 부재 → frontend 가 generic 5xx error UI 표시. **fix** (1) `backend-core/internal/httpapi/router.go` (MODIFY, +13/-2) — `platformStoreOrUnavailable` 의 response body 에 `code: "platform_store_unavailable"` 추가. 5 KPI/test endpoint 가 공통 사용 → 단일 fix 로 5 endpoint 적용. (2) `backend-core/internal/httpapi/platform_kpi_test.go` (MODIFY, +22) — `TestPlatformKPI_NilStoreReturns503WithCode` 신규 1 case. (3) `frontend/shared/utils/error-message.ts` (MODIFY, +24/-12) — `toUserErrorMessage` 가 503 + `code: "platform_store_unavailable"` 매핑 → "Backend store is not initialized" 자동 안내. `ApiError` instanceof 체크를 duck-typed shape (`asApiErrorLike` helper) 로 변경 — `@/shared/api/api-client` import cycle 회피. (4) `frontend/shared/ui-foundation/components/KpiTestErrorState.tsx` (NEW, 66 line) — 6 component (Platform/Project/Repository × KPI/Tests) 공통 error state component. (5) 6 component error state 교체. (6) `frontend/shared/utils/error-message.test.ts` (MODIFY, +17) — 503 + code 매핑 + 503 generic fallback 2 case 신규. (7) `frontend/shared/ui-foundation/components/__tests__/KpiTestErrorState.test.tsx` (NEW) — 3 case. 검증: backend `TestPlatformKPI_*` 3/3 PASS + frontend 8 test file 48/48 PASS + `tsc --noEmit` 본 PR 영향 0 + `go build ./...` silent. **Tier**: **공용** (backend helper + frontend shared component 1 file NEW + 6 component error state, 사내 한정 정보 0). |
| **KPI/test 카드 detail page (이슈 3)** | KPI/test 카드를 별도 페이지로 drill-down (platforms/projects/repositories × kpi/test-results = 6 page) | — | ✅ implemented (sprint `feat/260617-kpi-test-detail-page`, 2026-06-17) — 테스트 환경에서 KPI/test 카드 뿐 아니라 상세 확인이 가능하도록 별도 페이지 지원 요청. **fix** (1) `frontend/shared/ui-foundation/components/KpiTestDetailPage.tsx` (NEW, 70 line) — 6 page 공통 wrapper component. entityType (platform/project/repository) + kind (kpi/tests) + entityId props. back link + header + children. (2) **6 신규 page** — `app/(dashboard)/{platforms,projects,repositories}/[id]/{kpi,test-results}/page.tsx` (각각 `use(params)` Promise unwrap + KpiTestDetailPage wrapper + 기존 component). (3) **drill-down link** 6 component (Platform/Project/Repository × KPI/Tests) header 의 `자세히 보기` link 추가. `usePathname()` 으로 자기 자신 페이지 진입 시 link 숨김 (loop 회피). (4) test — `KpiTestDetailPage.test.tsx` (NEW) 3 case + 6 page smoke test 신규 (sync wrapper 로 React 19 use(params) Promise unwrap 회피) + 기존 4 component 회귀. (5) `frontend/lib/test-setup.ts` (MODIFY, +16) — `vi.mock("next/link")` + `vi.mock("next/navigation")` 추가. (6) `onRetry={() => loadTests(windowDays)}` 정공법 정정 — `loadResults(window)` 로. 검증: 14/14 test file 44/44 PASS + `tsc --noEmit` 본 PR 영향 0 + backend `go build ./...` silent. **Tier**: **공용** (frontend only + docs, 사내 한정 정보 0). |
| **codex P1/P2 review 정공법 fix (PR #635/#636/#637)** | PR 3건의 codex inline review 8건 정공법 fix (P1 1건 + P2 7건) | — | ✅ implemented (sprint `fix/260617-codex-review-feedback`, 2026-06-17) — **P1 (보안, 1건)**: `IssueIntakeTokenModal` 의 `0.0.0.0/0 + ::/0` default 가 production 의 admin 모달에서 accidental submit 시 의도하지 않은 전체 IP 허용 → bearer token leak 시 모든 IP 에서 사용 가능. `useState<string[]>(["0.0.0.0/0", "::/0"])` → `useState<string[]>([""])` (deny-by-default 정공법) + **dev 환경 한정** useEffect mount 시점에 `process.env.NODE_ENV === "development"` 분기로 자동 채움. production 환경 default `[""]` → admin 의 accidental submit 시 backend `invalid_allowed_ips` reject. **P2 (2 unique issue, 7 코멘트)**: (a) `useRepositoryBuildRuns` 의 `loadMore` stale branch 가 `setLoadingMore(false)` 누락 → button 영구 disabled. (b) `Platform/Project/RepositoryTestsSection` 의 drill-down link 가 `/tests` 가리키나 page route 가 `/test-results` → 404. `href` + `isOnDetailPage` 의 path 모두 `/test-results` 로 정합. 검증: 16/16 test file 62/62 PASS + `tsc --noEmit` 본 PR 영향 0. **Tier**: **사외** (frontend only + docs, 사내 한정 정보 0). |
| **Sprint 마무리 (final-merge)** | 4 PR (#634/#636/#637/#638) 의 변경 일괄 main 머지 | — | ✅ implemented (sprint `fix/260617-final-merge`, 2026-06-17) — PR #635 의 squash commit `0b5a6cb1` 가 main 머지 시 PR #633 + PR #633 의 codex P2 fix + PR #634 + PR #635 의 4 commit 의 commit message 본문을 모두 포함했으나, PR #636/#637/#638 의 변경 (kpi-test graceful fallback, detail page, codex P1/P2 fix) 은 main 의 squash commit 의 본문에 포함되지 않음. 3 PR 의 변경 (commit hash `fd7c7159` + `083ab837` + `3091927f`) 을 main 의 superset 으로 cherry-pick (`--strategy-option=theirs`) — 8 file +89/-23 변경. PR #634 (test-only) 의 변경은 PR #635 의 squash 의 본문에 이미 들어감 → close 만. 4 PR close (#634/#636/#637/#638) + single `fix/260617-final-merge` PR 머지 (commit `7189d277`). 검증: 16/16 test file 62/62 PASS + `tsc --noEmit` 본 PR 영향 0 + `go build ./...` silent. Tier: 사외/공용. |
| **X-6 (Keycloak group staging-prod 적용) 정공법 = docs only** | X-6 의 backend/script 변경 0 — 이미 main 정합. SOP + verify-keycloak-groups.sh + release_v0-1_roadmap.md X-6 row status `⏳ planned` → `🟡 in_progress` + issue #214 close 정공법 | — | ✅ implemented (sprint `feat/260617-x6-keycloak-groups-docs`, 2026-06-17) — (1) `docs/setup/keycloak_operations.md` §4.3 (group 4종 + composite realm role 1:1 매핑 SOP, [keycloak_groups_mapping.md](../domain/rbac-permissions/keycloak_groups_mapping.md) 옵션 B 채택) + §4.4 ([`scripts/verify-keycloak-groups.sh`](../../scripts/verify-keycloak-groups.sh) 자동 검증 4 항목: realm 존재 / group 4종 존재 / composite role 1:1 매핑 / Default Groups empty + exit code + FAIL 케이스) main 정합. (2) `docs/planning/release_v0-1_roadmap.md` §3.5 NEXT block 의 X-6 row status `⏳ planned` → `🟡 in_progress (docs only, 2026-06-17)` + Keycloak admin SOP + verify-keycloak-groups.sh mention + 잔여 = 사내 Keycloak admin console 1회 작업 (사용자 결정). (3) issue #214 close 정공법 (SOP + 자동 검증 main 정합 + 사용자 admin 작업 잔여). 검증: `tsc --noEmit` 본 PR 영향 0 + `go build ./...` silent (변경 0). **Tier**: **공용** (docs only, 사내 한정 정보 0). 잔여 = 사내 Keycloak admin console 1회 작업 (group 4 + composite role assign, 사용자 결정) + 사내 staging/prod 적용 후 verify-keycloak-groups.sh 1회 실행 (사용자 결정). |
| **X-7** | ADR-0016 §6 alert 임계 확정 (P2-2) | — | ⏳ planned (v0.1.1-alpha) |
| **X-8 staging hand-off + CI e2e smoke 정공법** | staging/prod 사내 SCM hand-off + CI 환경의 SPI e2e smoke + build script + SOP (X-8 후속) | — | ✅ implemented (sprint `feat/260617-x8-staging-handoff-e2e-smoke`, 2026-06-17) — X-8 (P2-6 + P3-5) 의 PR #641 의 정공법 (compose + docs) + 본 turn 의 staging hand-off + CI e2e smoke 통합. (1) [`docs/setup/keycloak_event_listener_spi_staging.md`](../setup/keycloak_event_listener_spi_staging.md) **NEW** (9KB) — staging/prod 사내 SCM hand-off 절차 7 step (사전 준비 → build → compose mount → Keycloak 재시작 → verify 자동 실행 → 결과 보고 → 일상 운영) + 5 가지 trouble-shooting. (2) [`scripts/build-keycloak-spi.sh`](../../scripts/build-keycloak-spi.sh) **NEW** — JAR build (`mvn clean package` 또는 docker buildx) + push (harbor.internal 등 사내 registry). 3 단계 (prerequisite check / maven build / push) + 사내 maven mirror 지원 + exit 0/1. (3) `docker-compose.test.yml` keycloak service — `image: quay.io/keycloak/keycloak:26.0` → `build: context: ./infra/idp, dockerfile: Dockerfile.keycloak` 변경 (CI 환경의 SPI e2e smoke 정공법) + SPI JAR volume mount + env 추가. (4) `keycloak-realm.ci.json` eventsListeners `["jboss-logging", "devhub-event-listener"]` 추가 (의도적 미적용 → CI e2e smoke 적용). (5) [`frontend/tests/e2e/keycloak-event-listener-spi.spec.ts`](../../frontend/tests/e2e/keycloak-event-listener-spi.spec.ts) **NEW** — Playwright e2e smoke 2 case (TC-KEYCLOAK-SPI-01 push smoke latency < 1s + TC-KEYCLOAK-SPI-02 env 정합). 검증: `tsc --noEmit` 본 PR 영향 0 + backend `go build ./...` silent (변경 0) + compose YAML lint 정공법 (colima + deploy + test 셋 다 OK). **Tier**: **공용** (Java source + compose + docs + e2e spec + script 만 — 사내 한정 정보 0). 잔여 = 사내 빌드 (Java 21 + Maven 3.13+ 환경) + staging/prod 적용 + `verify-keycloak-spi.sh` 1회 실행 (사용자 결정). |

### v0.1.1-alpha release 정공법 (메모리 4 file + release_v0-1_roadmap.md + CHANGELOG.md)

- `docs/planning/release_v0-1_roadmap.md` §3.5 NEXT block 의 title `v0.1.1` → `v0.1.1-alpha` 격하.
- `ai-workflow/memory/state.json` M-v0.1.0 notes: v0.1.1-alpha release 정합 + 잔여 5 의 8 item 의 v0.1.1-alpha 격하 마킹.
- `ai-workflow/memory/work_backlog.md` status line: v0.1.1-alpha release 정합 + §5 변경 이력 row.
- `ai-workflow/memory/session_handoff.md` §0: v0.1.1-alpha release 정합 subsection.
- `CHANGELOG.md`: 본 v0.1.1-alpha release note 추가.

### v0.1.1-alpha 후속 (실제 구현, 사용자 결정 시점)

- 잔여 5 의 8 item 의 실제 구현 = 사용자 결정 시점 별도 sprint.
- v0.1.1-alpha release 후 v0.1.2-alpha 또는 v0.2.0-alpha 로 release (사용자 결정).
- v0.1.0 정식 release = v0.1.x 의 follow-up patch + 사용자 결정 시점.

### Unchanged (v0.1.0-alpha 의 8 DoD 모두 close 정합 유지)

- 8 DoD: 7 ✅ + 1 ✅ skipped (N-6) = 8 DoD 모두 close (v0.1.0-alpha 정합).
- v0.1.0-alpha release 의 CHANGELOG + 발표 자료 + release 태그 (`v0.1.0-alpha`) 정합 유지.
- main HEAD `356d08b7` = v0.1.0-alpha release 정합 + v0.1.1-alpha release 정공법.

## [v0.1.0-alpha] - 2026-06-11

**v0.1.0-alpha release. 8 DoD 모두 close (7 ✅ + 1 ✅ skipped, N-6 사용자 결정).**

main HEAD: `d860b7c9` (PR #554 squash, N-6 skip + 4 file). Git tag: `v0.1.0-alpha`.

### 8 DoD (Definition of Done) — 모두 close

| # | DoD | Status |
|---|---|---|
| 1 | 사내 Keycloak realm OIDC 로그인 동작 | ✅ done (M0~M2) |
| 2 | system_admin 9 sub-page CRUD | ✅ done (M2) |
| 3 | Application/Repository/Project CRUD + rollup + 현황 페이지 | ✅ done (PR #104~#110) |
| 4 | HomeLab provider + sync + topology v2 | ✅ done (PR #155) |
| 5 | DREQ 흐름 end-to-end | ✅ done (PR #514 + #515) |
| 6 | e2e Playwright + backend `go test ./...` + frontend `npm run build` | ✅ done (CI 11/12 PASS) |
| 7 | 사내 staging 1주 운영 + 외부 사용자 ≥5 로그인 | ✅ skipped (사용자 결정, 2026-06-11) |
| 8 | UI 디자인 polish 1차 (semantic theme + responsive + a11y) | ✅ done |

### Added

**Auth / Org / Dashboard**
- Keycloak OIDC 통합 (ADR-0019, PR #167~#171) — JWKS cache + stale-while-error fallback (PR #242, 24h MaxStaleDuration)
- RBAC PermissionCache LISTEN/NOTIFY (RM-M4-08)
- Dashboard (developer/manager/admin) + 역할 routing (defaultLandingFor + isSystemAdmin)
- Account admin: Keycloak Admin Client 위임 (`/api/v0-1/accounts/*`, PR #167 KC-PR-C)
- Audit enrichment (source_ip / request_id / source_type, PR #57) + requireRequestID middleware
- Sign Out endpoint (P1-6, PR for N-8)
- e2e Kratos legacy 제거 + dynamic idp_subject sync (PR #249)

**Application / Repository / Project 도메인**
- API-01~58 activated (PR #104~#110, 7 마이그레이션 000012~000018)
- ADR-0011 accepted + 4 신규 RBAC resource (system_admin 일임)
- Platform 등록/조회 + Repository 연결 + Project CRUD + rollup + 현황 페이지 + SCM provider catalog
- 23 integration test (P1/P2 회귀 guard)

**DREQ (Development Request) 도메인**
- API-59~68 + API-79 (PATCH allowed_ips) activated (PR #514, ADR-0028)
- ADR-0012/0013/0014/0017
- Intake token + 외부 시스템 → DevHub POST → assignee dashboard → Promote (신규 application/project 1tx)
- 2 신규 table (dev_request_vocs + user_notifications) + dev_requests 4 column 확장 + 5 신규 API
- `(source_system, external_ref)` UNIQUE for idempotency
- TC-DREQ-* 13건 + `dev-requests.spec.ts` 6 step + `test_cases_m5_dreq.md`

**External Integration 도메인**
- API-69~80 activated (PR #135~#157)
- ADR-0015 / 0016 / 0017
- HomeLab pull (file + HTTP) + Provider/Binding CRUD + topology v2 시각화 + Prometheus 통합
- bindings 관리 UI (PR #154) + topology v2 시각화 (PR #155)
- ADR-0017 §6 atomicity 실 구현 (PR #156) + ADR-0015 §6 (1)+(2) (PR #157)
- TC-INT-FRONTEND-* 10건 + TC-INT-FRONTEND-BIND-* 3건 + TC-INT-HOMELAB-03 + TC-INT-FRONTEND-TOPOLOGY-V2-* 2건

**AI Workflow**
- ai-workflow v0.5.0 → v0.5.11 동기화 (PR #545, theirs-only 1 squash, 97 file / 4562줄)
- 2-tier governance (사외/사내 형상관리, PR #531~#537) — `AGENTS.md` §사외/사내 + `docs/governance/worker_division.md` §6
- v0.1.1 sprint -a follow-up (PR #538~#543) — port interface + saovae_stub (sso-integrations/keycloak)
- D-72 Phase 1/2/3 wiki integration + D-79/D-80 thin wrapper (PR #544, #551, #552)
- Vault ingestion (T-d-72-4, 82 wiki page 신규 ingest, vault commit b1599cc, 2026-06-11)
- T-d-72-2 (D-72 Phase 1 mirror) re-sync 완료 (2026-06-11 01:45:04Z, 83 file, 1.6M)

**UI / Frontend**
- Frontend 80 files / 1033 tests PASS (FE-08 신규)
- e2e Playwright 40 TC 게이트 (PR #86) + GitHub Actions CI 도입
- Design system (semantic theme + responsive + a11y baseline)
- PermissionEditor at /admin/settings/permissions ↔ /api/v0-1/rbac/policies
- e2e strict mode violation fix (`repositories-ui.spec.ts` L42 `.first()` 추가, commit `82935f8b`)

### Changed

- **워커 분업 전면 취소** (사용자 결정, PR #500, 2026-06-09) — 모든 신규 작업은 어느 에이전트로든 자유 진행
- **2-tier governance (사외/사내 형상관리)** (PR #531~#537, 2026-06-10) — GitHub 사외 = single source-of-truth / 사내 SCM = GitHub read-only pull
- **Kratos 잔재 residual cleanup** (sprint -ad) — 11 파일 삭제 + `identity_resolver.go` 신규 + Kratos 흐름 완전 제거
- `account_password.go` + `kratos_login_client` / `settings_client` / `session_cache` / `admin_client` + `password_auth_types` 모두 삭제
- `/api/v0-1/account/password` endpoint 폐기 — Keycloak Account Console redirect 위임 (`keycloak_operations.md` §8.5b)
- main flat memory 분리 (state.json 1515 → 150 line, 90% 감소, sprint `maintenance/work_260610-b-v0-1-pre-release-housekeeping`)
- v0.5.11-beta ai-workflow 표준화 (메모리 3 file reapply 분기 theirs-only 흡수 + 백업 보존)
- 명명 재검토 (PR #532, 2026-06-10) — `DEVHUB_APP_NAME` / `DEVHUB_APP_SHORT_NAME` env var override

### Fixed

- **N-8 sign-out e2e deterministic race hotfix 4차** (issue #501, PR #502 + #503) — 502 → 204 graceful degradation + `X-Keycloak-Likely-Down: true` marker + typed error sentinel `authview.ErrOIDCConfigMissing` + `authview.ErrOIDCNetworkUnreachable`
- **N-11 CI e2e + backend-integration 복원** (issue #419, sprint 260608-a) — `.github/workflows/ci.yml` 의 `&& false` 2건 코드 레벨 복원 + main 첫 PR 의 e2e shard 1/2/3 PASS 재확인
- e2e strict mode violation (`repositories-ui.spec.ts` L42 `.first()` 추가, commit `82935f8b`, post-merge standard)
- **DB 에러 노출 (SEC-5)** mask 5xx errors (PR-A, M0)
- `SetTrustedProxies(nil)` — `audit_logs.source_ip` X-Forwarded-For 위조 방어
- CI 회귀 (2026-06-01) — `frontend/app/(dashboard)/applications/[id]/page.tsx` 중복 import 제거 + `TC-PROJ-UI-04` 환경 독립 검증
- e2e seed 정합 (E2E seed `e2e-repo-a` 1 row 만 insert, 차트 toggle 전후 strict mode 회피)

### Security

- **SEC-1~SEC-4 resolved** (M0, 2026-05-08) — Keycloak OIDC + JWKS cache + resource_access fallback
- **DB 에러 노출 (SEC-5)** mask 5xx errors (PR-A, M0)
- Audit actor enrichment + requireRequestID middleware (PR-D, M1)
- JWKS stale-while-error fallback (PR #242, sprint -l, ADR-0020 sub-carve D) — 24h MaxStaleDuration, `DEVHUB_OIDC_JWKS_MAX_STALE_DURATION` env
- `metric devhub_jwks_stale_while_error_total{result}` + `devhub_jwks_stale_age_seconds` Histogram

### Deprecated

- `infra/idp/hydra.yaml` / `kratos.yaml` + setup README/ENVIRONMENT_NOTES (PR #169) — Keycloak 단일 IdP 전환
- `/api/v0-1/account/password` endpoint — Keycloak Account Console redirect 위임
- `DEVHUB_HYDRA_*` / `DEVHUB_KRATOS_*` env — PR #167 머지로 제거
- `infra/idp/README.md` — deprecated (PR #169)

### Removed

- **Kratos 전체 흐름** (PR #167 + sprint -ad) — historical Hydra introspection + Kratos 전체 흐름
- `account_password.go` + `kratos_login_client` / `kratos_settings_client` / `kratos_session_cache` / `kratos_admin_client`
- `kratos_admin_client_test` / `kratos_identity_resolver_test` (→ `identity_resolver_test` rename) / `kratos_login_fake_test` / `kratos_session_cache_test` / `kratos_settings_client_test` / `account_password_test`

### v0.1.0-alpha 후속 (잔여 follow-up)

**잔여 3** (T-d-79-2 / T-d-80-2, my_harness 측 SSOT 작성) — 사용자 전달 후 진행 중.
**잔여 5** (T-d-72-5/6 + D-73/74 + X-1~8, v0.1.1 forward path) — v0.1.1 milestone (2026-07-31) 진입 시점 별도 sprint.
**vault Gitea remote push** — 사용자 수동.

### 8 DoD 잔여 (skipped)

- **N-6 (v0.1.0 staging 1주 운영 검증)** — ✅ skipped, 사용자 결정 (2026-06-11). v0.1.0-alpha release blocker 0건.

### 1차 종합 매트릭스 (2026-05-13, PR #89 + #90)

- REQ-FR 105 + REQ-NFR 26 + ARCH 17 + API 40 + RM 28 + IMPL 79 + UT 47 + TC 37 = 412 항목
- 도메인 그룹 13행 (auth-session, integration-registry, repository-integration 등)
- [`docs/traceability/report.md`](docs/traceability/report.md) 갱신 — REQ → UC → ARCH → API → RM → IMPL → UT → TC 19 row 매트릭스

### 1차 종합 보안 리뷰 (2026-05-08)

- 신규 0건, 기존 1건 재확인 (SEC-5 권고)
- [`ai-workflow/memory/codebase-security-review-2026-05-08.md`](ai-workflow/memory/codebase-security-review-2026-05-08.md)
