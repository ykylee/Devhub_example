# DevHub v1.0 릴리즈 로드맵 — 1차 릴리즈 scope + 우선순위 + 워커 분업

- 문서 목적: DevHub 의 1차 릴리즈 (v1.0) 를 위한 단일 source-of-truth 로드맵. 잔여 carve 통합 인벤토리 + 우선순위 P0~P3 + 마일스톤 재정의 + 워커(Claude/Codex/Gemini/Reasonix/OpenCode) 분업.
- 범위: 1차 릴리즈에 포함될 기능 scope, 제외 기능, 잔여 carve 우선순위, 신규 마일스톤(M-v1.0, M-v1.1, M-v2), GitHub project + milestone 등록 plan, UI 검증 방식.
- 대상 독자: 프로젝트 리드, 모든 워커 (Claude, Codex, Gemini, OpenCode), 후속 작업자.
- 상태: draft
- 최종 수정일: 2026-06-11 (**N-13 housekeeping 정공법** — §3.5 N-13 row 의 status `⏳ planned` 보강 + ID slot 9 row 발급 + conventions.md §1 RM 표기 확장 + ADR-0028 §6 (a) + §7 변경 이력 + §9 본 row 추가. 본 sprint = docs only housekeeping, 구현은 v1.1 milestone 진입 시점 별도 sprint `feat/work_260611-a-n13-inbound-source-impl`.)
- 직전 결정 근거: 2026-05-27 (**코드베이스 스냅샷 정합** — Onboarding IMPL carve A/B/C/D 전부 완료 (P2-8..12 ✅, PR #278/#288/#289/#290/#291) + lazy_auto_create 폐기 + v1.1 영역 작업 일부 v1.0 전 선행 + §3.5 N/X/E 백로그 추가. 분석 근거 [docs/analysis/2026-05-27-codebase-snapshot](../analysis/2026-05-27-codebase-snapshot/README.md))
- 결정 근거 sprint: `claude/work_260520-f-roadmap` (본 문서)
- 관련 문서: [통합 개발 로드맵](../development_roadmap.md) (M0~M6 historical), [requirements](../requirements.md), [architecture](../architecture.md), [ADR-0019 Keycloak](../adr/0019-keycloak-only-idp.md), [ADR-0020 계정/사용자 책임 경계](../adr/0020-account-user-management-boundary.md), [traceability matrix](../traceability/report.md), [account_user_management_redesign Phase 1/2/3](./account_user_management_redesign.md), [keycloak_operations](../setup/keycloak_operations.md).

## 0. 사용 가이드

본 문서는 **2026-05-20 이후 모든 sprint 의 진입점**. 기존 [`docs/development_roadmap.md`](../development_roadmap.md) (M0~M6 historical) 는 done milestone 의 사후 명문화 자산으로 보존 — 본 문서가 v1.0 / v1.1 / v2 의 새 source-of-truth.

**2026-06-09 결정 — 본 문서 §5 의 워커 분담 표 + 인계 SOP + 충돌 처리 SOP 의 강제력은 무효.** Claude/Codex 자유 이용 불가로 사용자 결정. 모든 신규 sprint / PR / 작업은 **어느 에이전트로든 자유롭게** 진행 가능. 작업 우선순위 (P0~P3) 자체는 유효.

1. **신규 sprint 진입 전** §3 우선순위 매트릭스 확인 → P0 carve 부터 흡수
2. (취소) **워커 작업 분담** §5 — 2026-06-09 전면 취소. 영역/인계 강제력 없음
3. **마일스톤 진행** §4 의 M-v1.0/M-v1.1/M-v2 표에서 본 sprint 의 마일스톤 + 동반 issue 확인
4. **결정 변경** 발생 시 §9 변경 이력에 row 추가

## 1. v1.0 릴리즈 scope

### 1.1 포함 기능 (3 domain)

| Domain | scope | 현황 (2026-05-20) |
| --- | --- | --- |
| **인증/조직/대시보드** | Keycloak OIDC 로그인 + `/admin/settings/*` (users/organization/permissions + audit + dev-requests + dev-request-tokens + integrations + integration-bindings) + dashboard (developer/manager/admin) + 역할 routing | Backend done (M1/M2/M3). Frontend done. **계정/사용자 리팩토링 Phase 3 sub-carve A done** (ADR-0020, sprint -d). sub-carve B (`/api/v1/accounts/*` 폐기 + lazy auto-create) 가 v1.0 의 마지막 큰 backend 변경. |
| **Platform/Repository/Project 도메인** | Platform 등록/조회 + Repository 연결 + Project CRUD + rollup + 현황 페이지 + SCM provider catalog | Backend done (API-01..58 + ADR-0011, 2026-05-14). Frontend 활성 (현황 페이지 + FilterBar). |
| **External Integration** | HomeLab pull (file + HTTP) + Provider/Binding CRUD + topology v2 시각화 + Prometheus 통합 | Backend done (API-69..80 + ADR-0015/0016/0017). Frontend 활성 (provider + binding + topology v2). |

### 1.2 제외 기능 (v2 또는 carve)

| 기능 | 분류 | 이유 |
| --- | --- | --- |
| **AI Gardener gRPC + Suggestion Feed** | v2 P3 | 보조 기능, 1차 릴리즈 핵심 가치 외부. backend `backend-ai/` Python 모듈 + Go Core client 미구현 |
| ~~**Sign Up (셀프 가입)**~~ | **permanently cancelled (2026-05-20)** | DevHub 가 Keycloak admin 권한이 없는 외부 IdP 운영 시나리오 (ADR-0020 결정 A 정합). user 생성은 IdP 팀 admin console 또는 HRDB ETL push 책임. DevHub 셀프 가입 흐름 미운영. 이전 'v1.1 carve' 분류 → cancelled. issue #235 closed. |
| **MFA / 2FA** | v1.1 carve | ADR-0019 §5.3 (5) 사내 정책 제외 — Keycloak MFA enrollment 는 Account Console 위임. DevHub UI 무관 |
| **Weekly report worker** | v2 P3 | frontend_integration §3.4 — backend 미구현 |
| **WebSocket replay + 리소스 필터링** | v1.1 carve | RM-M4-02 — backend WebSocket 기반 인프라 + last event replay. v1.0 은 polling/refresh 로 충분 |
| **System Admin 대시보드 (Runner 상태)** | v2 P3 | RM-M4-07 — Gitea Runner 운영 view. v1.0 scope 외 |
| **Gitea Hourly Pull worker** | v2 P3 | RM-M4-06 — 외부 SCM 비동기 동기화. v1.0 은 HomeLab pull 만 |
| **외부 SSO (Gitea 연동 등)** | v1.1 carve | RM-M4-09 — Keycloak identity broker. v1.0 은 단일 Keycloak realm 전제 |

### 1.3 v1.0 완료 정의 (DoD)

1. 사내 Keycloak realm 에서 OIDC 로그인 가능 (Sign In / Sign Out / token refresh)
2. system_admin 이 `/admin/settings/*` 9 sub-page 모두 접근 + CRUD 동작
3. Application / Repository / Project 의 CRUD + rollup + 현황 페이지 동작
4. HomeLab integration provider 등록 + sync + binding 생성 + topology v2 시각화 동작
5. DREQ 흐름 — intake token 발급 + 외부 시스템 → DevHub POST → assignee dashboard 노출 → Promote (신규 application/project 1tx) → close
6. e2e Playwright 전 shard PASS + backend `go test ./...` PASS + frontend `npm run build` PASS
7. 사내 staging 환경 1주 운영 + 외부 사용자 ≥ 5명 로그인 동작 ([Onboarding 운영 SOP](../setup/onboarding_operations.md) §7 DoD 8 항목 통과) — **(N-6 skipped, 사용자 결정, 2026-06-11, v1.0 release blocker 해제)**
8. UI 디자인 polish 1차 완료 (semantic theme + responsive + a11y baseline)

## 2. 도메인 모듈 매트릭스

### 2.1 인증/조직/대시보드 (Auth/Org/Dashboard)

| 모듈 | 위치 (backend) | 위치 (frontend) | v1.0 상태 |
| --- | --- | --- | --- |
| Keycloak OIDC 토큰 검증 | `internal/auth/keycloak_verifier.go` + JWKS cache + resource_access fallback | `lib/services/auth.service.ts` + OIDC discovery + tokenStore | ✅ done (ADR-0019) |
| Keycloak Admin Client (계정 admin) | `internal/httpapi/keycloak_admin_client.go` (`Create/Update/SetState/Delete Identity`) | `lib/services/account.service.ts` 5 method | 🟡 **ADR-0020 결정 A 로 전면 폐기** — sub-carve B (sprint -f) 가 backend handler 제거 + lazy auto-create 도입 + frontend cleanup |
| `users` CRUD + appointments | `internal/httpapi/organization.go` + `internal/store/postgres_users.go` | `lib/services/identity.service.ts` (getUsers/createUser/updateUser) | ✅ done |
| 조직 단위 (units + hierarchy) | `internal/httpapi/organization.go` (`/api/v1/organization/*`) | `app/(dashboard)/admin/settings/organization/page.tsx` | ✅ done |
| RBAC policy 편집 | `internal/httpapi/rbac.go` (4 endpoint, sub-carve A 정리됨) | `app/(dashboard)/admin/settings/permissions/page.tsx` + PermissionEditor | ✅ done (ADR-0002 + ADR-0020 sub-carve A) |
| Audit log view | `internal/httpapi/audit.go` + Keycloak event listener (sprint -u~-y) | `app/(dashboard)/admin/settings/audit/page.tsx` | ✅ done |
| Dashboard (developer/manager/admin) | `internal/httpapi/dashboard_metrics.go` | `app/(dashboard)/{developer,manager,admin}/page.tsx` | ✅ done |
| Role routing + AuthGuard | — | `lib/auth/role-routing.ts` + AuthGuard layout | ✅ done |

### 2.2 Platform/Repository/Project 도메인

| 모듈 | 위치 (backend) | 위치 (frontend) | v1.0 상태 |
| --- | --- | --- | --- |
| Platform CRUD + 상태 전이 + critical_warning guard | `internal/httpapi/applications*.go` + `internal/store/applications*.go` (API-41..47) | `app/(dashboard)/admin/settings/platforms/page.tsx` + `app/(dashboard)/platforms/{page,[id]/page}.tsx` | ✅ done (ADR-0011) |
| PlatformRepository link CRUD | API-48..50 + `platform_repositories` 테이블 | `components/project/ApplicationCreationModal.tsx` 등 | ✅ done |
| SCM Provider catalog | API-41/42 + 4 seed (bitbucket/gitea/forgejo/github) | provider 선택 UI | ✅ done |
| Project CRUD | API-55/56 + `projects` 테이블 | `app/(dashboard)/admin/settings/platforms/...` + `app/(dashboard)/projects/{page,[id]/page}.tsx` | ✅ done |
| Repository ops (activity / PR / build / quality) | API-51..54 | `repositories/{page,[id]/page}.tsx` | ✅ done |
| Application rollup + custom weight | API-57 + `weight_policy` | rollup 표시 | ✅ done |
| Project Integration CRUD (legacy, separate from External Integration) | API-58 | — | ✅ done |
| **DREQ (Dev Request)** | API-59..68 + `requireIntakeToken` middleware + Promote-Tx | `app/(dashboard)/{dev-requests,admin/settings/{dev-requests,dev-request-tokens}}/page.tsx` | ✅ done (M5 closing) |

### 2.3 Onboarding 도메인 (Keycloak self-service unit selection)

| 모듈 | 위치 (backend) | 위치 (frontend) | v1.0 상태 |
| --- | --- | --- | --- |
| Concept §5.1~§5.9 + skip-and-resume | — (design doc) | — | ✅ done (PR #260 + #265) |
| Requirements §5.7 (REQ-FR-ONBOARD-001..012 + REQ-NFR-ONBOARD-001..008) | — (spec) | — | ✅ done (PR #266) |
| Usecase + Architecture + API contract | — (spec) | — | ✅ done (PR #267 — UC-ONBOARD-01..11 + ARCH-ONBOARD-01..06 + API-83..86 + API-32/33 확장) |
| ADR-0021 (책임 경계 확장 + lazy auto-create supersession) | — (ADR) | — | ✅ done (PR #269, ADR-0020 partial supersession 5 위치) |
| IMPL carve plan + RM-ONBOARD-01..04 | [`docs/domain/onboarding/impl_plan.md`](../domain/onboarding/impl_plan.md) | — | 본 sprint done |
| RM-ONBOARD-01 IMPL-backend (handler + middleware + migration) | `internal/httpapi/{onboarding_gate,me_onboarding,organizations_search,users_admin_review,onboarding_roles,onboarding_feature_flag}.go` | — | ✅ done (Carve A, PR #278) |
| RM-ONBOARD-02 IMPL-frontend (page + picker + banner + gating) | — | `app/onboarding/page.tsx` + `components/onboarding/*` + `(dashboard)/layout.tsx` + `account/page.tsx` | ✅ done (Carve B/C, PR #288) |
| RM-ONBOARD-03 IMPL-admin (Confirm Review + pending_review filter) | — | `app/admin/settings/users/page.tsx` + `components/admin/users/ConfirmReviewModal.tsx`·`PendingReviewPanel.tsx` | ✅ done (Carve B/C, PR #288) |
| RM-ONBOARD-04 IMPL-tests (UT + E2E) | `internal/httpapi/onboarding_test.go` | `tests/e2e/onboarding-first-login.spec.ts` + 6 test seed | ✅ done (Carve D, PR #289 + flag ON #290 + hotfix #291) |
| lazy_auto_create 폐기 (ADR-0021 §3.3 정공법) | `lazy_auto_create.go`·`onboarding_feature_flag.go` 삭제 + flag default ON | — | ✅ done (PR #290, issue #284 closed) |

### 2.4 External Integration

| 모듈 | 위치 (backend) | 위치 (frontend) | v1.0 상태 |
| --- | --- | --- | --- |
| Integration Provider CRUD | `internal/httpapi/integration_registry.go` (API-69..72, 80 DELETE) | `app/(dashboard)/admin/settings/integrations/page.tsx` + ProviderTable + ProviderModal | ✅ done |
| Integration Binding CRUD | API-74/75 + `integration_bindings` | `app/(dashboard)/admin/settings/integration-bindings/page.tsx` + BindingsTable + CreateBindingModal | ✅ done |
| HomeLab pull adapter (file + HTTP) | `internal/integrations/adapters/{homelab,homelab_file_puller,homelab_http_puller,homelab_pull_loop}.go` | — | ✅ done (ADR-0015) |
| Infra topology v2 | API-76/78 + `infra_service_snapshots` | `app/(dashboard)/admin/topology-v2/page.tsx` + React Flow | ✅ done |
| Prometheus metrics | `/metrics` + Counter/Gauge/Histogram | Grafana dashboard JSON (`docs/setup/grafana/homelab_dashboard.json`) | ✅ done (ADR-0016 §6 (1)+(2)) |
| Provider webhook ingest (push) | `POST /api/v1/integration/providers/:provider_id/webhook` | — | ✅ done (Bypass) |

## 3. 잔여 carve 통합 인벤토리 + 우선순위 매트릭스

### 3.1 P0 — v1.0 릴리즈 차단 (sprint -f 즉시 진입)

| ID | Carve | 출처 | 워커 | 비고 |
| --- | --- | --- | --- | --- |
| **P0-1** | ADR-0020 sub-carve B — `/api/v1/accounts/*` 4 endpoint 제거 + lazy auto-create + frontend `account.service.ts` 폐기 + admin/settings/users 페이지 정리 | sprint -d ADR-0020 §4.1 B | **Claude (backend)** + **Gemini (frontend done ✅)** | v1.0 Keycloak 단일 IdP 정합의 마지막 큰 변경. e2e TC-ACC-* 갱신 동반 |
| **P0-2** | UI 디자인 polish 1차 (semantic theme 정합 + responsive + a11y baseline) | 사용자 지시 (2026-05-20) — \"UI 띄워놓고 디자인 손보기\" | **Gemini (frontend+UX)** | ✅ done (sprint gemini/work_260520-b). PR #203 의 hardcoded color → semantic theme 패턴 확장 완료. 모든 modal + 페이지 + responsive sidebar 적용. |
| **P0-3** | Playwright screenshot mode 도입 + CI artifact 업로드 | 사용자 지시 (2026-05-20) — UI 검증 방식 | **Codex (infra+CI) + Gemini (frontend test config)** | screenshot 자산이 Gemini 의 디자인 작업 source. shard 별 캡처 |
| **P0-4** | **CI Run 생성 API 구현** (`POST /api/v1/ci-runs`) — Gitea Actions Webhook 수신 또는 직접 생성 | 2026-06-01 통합 테스트 ISSUE-05 | **Claude (backend)** | **신규 P0 (v1.0 차단)** — CI/CD 기능의 실질적 사용을 위해 필수. status validation: queued/running/success/failed/cancelled/skipped/unknown. **ID 슬롯 정합 (2026-06-06)**: 정식 GitHub issue = §3.5 N-7 ([#486](https://github.com/ykylee/Devhub_example/issues/486)). 2026-05-20 의 [#238](https://github.com/ykylee/Devhub_example/issues/238) (Docker 단일 포트, CLOSED) 도 동일 슬롯 라벨 사용 — historical 정합은 #238 유지 (P0-4 의 본래 의도), 본 carve 의 정식 추적 ID = §3.5 N-7. |

### 3.2 P1 — v1.0 안정성 (sprint -g/-h)

| ID | Carve | 출처 | 워커 | 비고 |
| --- | --- | --- | --- | --- |
| **P1-1** | ADR-0020 sub-carve C — Keycloak event listener 확장 (USER:UPDATE / GROUP_MEMBERSHIP / USER:DELETE 매핑) + DevHub `users` write + metric 3종 | ADR-0020 §4.1 C, design doc §5.3 | **Claude (backend)** | sprint -u~-y 자연 확장. P0-1 의 lazy auto-create 와 role 추출 로직 공유 |
| **P1-2** | ADR-0020 sub-carve D — JWKS stale-while-error expiry case 확장 | ADR-0020 §4.1 D | **Claude (backend)** | sprint -r kid mismatch fallback 자연 확장. Keycloak unreachable 시 uptime 보장 |
| **P1-3** | ADR-0019 §5.3 — Keycloak group staging-prod 적용 | session_handoff 잔여 carve | **사용자 + Codex** | Keycloak admin 1회 작업 (group 4 + composite role assign). GitHub issue [#214](https://github.com/ykylee/Devhub_example/issues/214) OPEN. |
| ~~**P1-4**~~ | ~~ADR-0019 §5.3 — off-boarding Phase 1 cron 실 deploy~~ | ~~session_handoff 잔여 carve~~ | — | **permanently cancelled (2026-05-20, issue #215 close)** — 외부 Keycloak 시나리오 채택. HR ↔ Keycloak sync 는 외부 IdP 팀 책임. DevHub off-boarding sync 는 sub-carve C event listener (PR #241) 가 정공법. `scripts/hrdb_etl_sync.sh` deprecation. |
| **P1-5** | ADR-0019 §5.3 — e2e Kratos → Keycloak 실 코드 전환 | session_handoff 잔여 carve | **Gemini (frontend test) + Codex (CI infra)** | sprint -m design 따름. 사내 staging Keycloak e2e 환경 동반. PR #203 의 `ci-e2e-sync-check.sh` 가 CI 단 일부 해소 |
| **P1-6** | **Sign-out endpoint 구현** (`POST /api/v1/auth/logout`) — access token 폐기 + session 종료 | 2026-06-01 통합 테스트 BUG-03 | **Claude (backend)** | **신규 P1** — 세션 관리 기본. refresh token rotate 포함 여부 결정. 정식 GitHub issue = §3.5 N-8 ([#488](https://github.com/ykylee/Devhub_example/issues/488)). |
| **P1-7** | **Repository build-runs endpoint 구현** (`GET /api/v1/repos/{id}/build-runs`) — `ci_runs` 테이블 기반 repo-scoped 조회 | 2026-06-01 통합 테스트 ISSUE-04 | **Claude (backend) + Gemini (frontend)** | **기본 구현 완료** (2026-06-11, N-9 정합). 정식 GitHub issue = §3.5 N-9 ([#487](https://github.com/ykylee/Devhub_example/issues/487)). 본 carve 의 endpoint `GET /api/v1/repositories/:repository_id/build-runs` (main 컨벤션, `/repos` 약칭이 아닌 repositories 풀패스) + `platformStoreOrUnavailable` 가드 + UT 3건 + IT 1건 + frontend `repositoryService.getRepositoryBuildRuns` + `DeveloperView` 위젯 모두 main 반영. **잔여 4건 sub-issue 분리** (v1.1 milestone 진입 시점): RBAC 403/404 가드 / Histogram metric `devhub_repository_build_runs_query_duration_seconds` / TanStack Query `useRepositoryBuildRuns` hook + status filter dropdown + skeleton / dashboard widget "Recent repository activity" 통합 + 독립 e2e spec. |
</input>

### 3.3 P2 — v1.0 운영 안정성 + v1.1 carve

| ID | Carve | 출처 | 워커 | 비고 |
| --- | --- | --- | --- | --- |
| **P2-1** | ADR-0020 sub-carve E — service account 권한 축소 (manage-users 제거) + governance SOP `keycloak_operations.md §8.5c` | ADR-0020 §4.1 E | **Codex (infra) + Claude (docs)** | docs only + Keycloak admin SOP. P0-1 머지 후 |
| **P2-2** | ADR-0016 §6 — pull latency p95 alert + push webhook 알림 + stage→prod 임계 확정 | ADR-0016 §6 (3)+(4)+(5) | **Codex (infra)** | baseline 1주 관찰 후 |
| ~~**P2-3**~~ | ADR-0017 §6 (b) — PATCH expires_at + admin UI 편집 modal | ADR-0017 §6 (b) | **Gemini (frontend) + Claude (backend)** | ✅ **resolved** (PR #137 sprint `gemini/work_260521-c-219-patch-token` + sprint `adr0017-token-mutation`). issue [#219](https://github.com/ykylee/Devhub_example/issues/219) backlog 누락분 closed (2026-05-21 검증). backend handler (`intakeTokenAdminUpdateRequest.ExpiresAt`) + frontend `EditIntakeTokenModal.tsx` + UT (`dev_requests_test.go` line 1335-1359) 모두 active. |
| **P2-4** | Bindings UI 강화 — scope_id lookup combobox + Edit/Delete + pagination | development_roadmap §6 잔여 | **Gemini (frontend+UX done ✅)** | v1.0 UI polish 동반. backend CRUD 지원 추가 완료. |
| **P2-5** | React Flow group sub-node + WebSocket 실시간 (`infra.node.updated` / `infra.service.updated`) | development_roadmap §6 잔여 | **Gemini (frontend done ✅)** | topology v2 강화. WebSocket 실시간 연동 및 Environment 그룹화 완료. |
| **P2-6** | Keycloak SPI provider JAR (PR #203 codex P2) | PR #203 codex review | **Codex (infra) + 사용자 (Java 빌드 환경)** | `infra/idp/devhub-event-listener/` Maven 또는 Gradle 빌드 + compose volume mount + 운영 SOP |
| ~~**P2-7**~~ | ~~신규 user 의 unit 초기 배치 자동화 — HRDB ETL pre-stage 가 unit 정보 동반~~ | ADR-0020 §5.5.2 | — | **cancelled (2026-05-21, issue [#223](https://github.com/ykylee/Devhub_example/issues/223) closed not-planned)** — 외부 Keycloak 시나리오 (`hrdb_etl_sync.sh` 이미 DEPRECATED, PR #257) + ADR-0021 §3.1 self-service onboarding 이 unit 매핑 cover. 두 정합 결정으로 본 carve 무효화. |
| **P2-8** | **RM-ONBOARD-01** IMPL-backend — `users` migration (`onboarding_completed_at` + `review_status` + CHECK) + `onboardingGate` middleware + 5 handler (API-83/84/85/86 + API-32/33 확장) + lazy_auto_create.go 폐기 + audit event const 3종 | ADR-0021 §6.1, [onboarding_impl_plan.md](../domain/onboarding/impl_plan.md) §2.1 | **Claude (backend)** | ✅ **resolved** (Carve A, PR #278). feature flag default OFF → Carve D 에서 ON flip. |
| **P2-9** | **RM-ONBOARD-02** IMPL-frontend — `/onboarding` page + OrganizationPicker (typeahead + tree) + skip flag sessionStorage + dismissible banner + `(dashboard)/layout.tsx` 3-branch gating + `/account` self-service unit edit | ADR-0021 §6.1, [onboarding_impl_plan.md](../domain/onboarding/impl_plan.md) §2.2 | **Gemini→Claude (override)** | ✅ **resolved** (Carve B, PR #288). |
| **P2-10** | **RM-ONBOARD-03** IMPL-admin — `/admin/settings/users` 의 "Confirm Review" 액션 + pending_review filter + `ConfirmReviewModal.tsx` | ADR-0021 §6.1, [onboarding_impl_plan.md](../domain/onboarding/impl_plan.md) §2.3 | **Gemini→Claude (override)** | ✅ **resolved** (Carve C, PR #288). |
| **P2-11** | **RM-ONBOARD-04** IMPL-tests — UT-onboarding-* (backend handler + middleware) + TC-ONBOARD-* 11건 (E2E mega lifecycle, REQ-NFR-ONBOARD-008 의 6 test seed) + `docs/domain/onboarding/test_cases.md` | ADR-0021 §6.1, [onboarding_impl_plan.md](../domain/onboarding/impl_plan.md) §2.4 | **Claude (UT) + Gemini (E2E)** | ✅ **resolved** (Carve D, PR #289 + codex hotfix #291). |
| **P2-12** | **lazy_auto_create.go deletion** — ADR-0021 §3.3 정공법 완성. `lazy_auto_create.go` + `onboarding_feature_flag.go` 2 파일 삭제 + `authenticateActor` flag 분기 제거 + `account.lazy_provisioned`/`user.role_default_assigned` audit emit 중단 + UT 정리 + ADR-0020 §4.1 sub-carve B inline banner 갱신 | ADR-0021 §3.3, [issue #284](https://github.com/ykylee/Devhub_example/issues/284) | **Claude (backend)** | ✅ **resolved** (PR #290 — flag default ON flip + `lazy_auto_create.go`/`onboarding_feature_flag.go` 삭제, issue #284 closed). 사내 staging 1주 monitoring 잔여. |

### 3.4 P3 — v2 후순위

| ID | Carve | 출처 | 워커 | 비고 |
| --- | --- | --- | --- | --- |
| **P3-1** | ADR-0020 sub-carve F — `/login` page 정리 | ADR-0020 §4.1 F | **Gemini (frontend)** | minor UX 정리. 우선순위 가장 낮음 |
| **P3-2** | ADR-0015 §6 (3) dedicated worker binary | ADR-0015 §6 | **Claude (backend)** | M4 진입 시 재평가 |
| **P3-3** | ADR-0015 §6 (4) push/pull dedup 정책 | ADR-0015 §6 | **Claude (backend)** | 별도 ADR 후보 |
| **P3-4** | ADR-0019 §5.3 — HA Phase 2 (Infinispan + shared PG + LB) | ADR-0019 §5.3 (6) | **사용자 (인프라 결정) + Codex** | 별도 ADR 후보 — ADR-0021 은 Onboarding 으로 발급됨 (2026-05-21), HA Phase 2 의 ADR 은 진입 시점에 다음 번호로 발급 |
| **P3-5** | ADR-0019 §5.3 — audit event listener SPI push 전환 | ADR-0019 §5.3 (9) §8.6.9 | **Claude (backend)** | polling latency 30s → < 1s. P2-6 의 SPI JAR 도입 후 |
| **P3-6** | RM-M4-01..03 WebSocket 확장 (replay + 리소스 필터링 + command status UI) | development_roadmap M4 | **Claude (backend) + Gemini (frontend)** | v1.1 또는 v2 |
| **P3-7** | RM-M4-04..05 AI Gardener gRPC + Suggestion Feed | development_roadmap M4 | **Claude (backend) + Gemini (frontend)** | v2 보조 기능 |
| **P3-8** | RM-M4-06 Gitea Hourly Pull worker | development_roadmap M4 | **Claude (backend)** | v2 |
| **P3-9** | RM-M4-07 System Admin 대시보드 (Runner 상태) | development_roadmap M4 | **Claude (backend) + Gemini (frontend)** | v2 |
| **P3-10** | RM-M4-08 ADR-0007 PermissionCache LISTEN/NOTIFY 실 구현 | development_roadmap M4 | **Claude (backend)** | 다중 인스턴스 시 |
| **P3-11** | RM-M4-09 외부 SSO (Gitea / AD federation) | development_roadmap M4 | **Codex (infra)** | Keycloak identity broker |
| ~~**P3-12**~~ | ~~Sign Up (셀프 가입) — 인사 DB 연동~~ | **cancelled (2026-05-20)** — DevHub Keycloak admin 권한 없음, IdP 팀 책임. issue #235 closed | — | — |
| **P3-13** | MFA / 2FA | M4 + ADR-0019 §5.3 (5) | (제외) | 사내 정책 — Keycloak Account Console 위임 |
| **P3-14** | RM-APPDASH-01 — Application 개발 대시보드 (API-93 backend + frontend) | v2 P3 | **Claude (BE) + Gemini (FE)** | GitHub issue [#384](https://github.com/ykylee/Devhub_example/issues/384) OPEN. PR #482 의 repository dashboard 개편 패턴 차용 (domain/application-lifecycle). |

### 3.5 신규 도출 백로그 (2026-05-27 코드베이스 스냅샷 §06)

> v1.1 영역 작업(Gitea SCM sync·SCM↔시스템 양방향·auth_mode full·repository draft/publish·admin catalog)이 **v1.0 전에 backend 주도로 선행**됐다. 그 결과 코드는 앞서고 테스트·운영 가시성·문서가 뒤처진 상태 → 다음 사이클은 "굳히기(harden)" 우선. 상세 분석은 [코드베이스 스냅샷 §06 향후 방향](../analysis/2026-05-27-codebase-snapshot/06_future_direction.md).

#### NOW — v1.0 마감 + 품질 굳히기

| ID | 아이템 | 영역 | 워커 |
| --- | --- | --- | --- |
| **N-1** | 문서 drift 정합 (추적성/로드맵/스펙 — 본 PR) | 문서 | Claude ✅(본 PR) |
| **N-2** | repository draft→publish UT/통합테스트 보강 (#368 무테스트) | BE | Claude |
| **N-3** | SCM import/create + draft/publish happy-path E2E | FE+BE | Gemini+Claude |
| **N-4** | 프론트 service/component 단위테스트 보강 (vitest 10→확대) | FE | Gemini |
| **N-5** | 마이그레이션 prefix uniqueness CI guard 강화 (000042 충돌 재발 방지, branch protection required check) | CI | Codex |
| **N-6** | v1.0 staging 1주 운영 검증 (외부 사용자 ≥5 로그인 + Onboarding SOP DoD 8) | 사내 | 사용자 | **status**: ✅ skipped (사용자 결정, 2026-06-11, v1.0 release blocker 해제).
| **N-7** | **CI Run 생성 API (P0-4) 구현** — 2026-06-01 통합 테스트 ISSUE-05 | BE | Claude |
| **N-8** | **Sign-out endpoint (P1-6) 구현** — 2026-06-01 통합 테스트 BUG-03 | BE | Claude |
| **N-9** | **Repository build-runs (P1-7) 구현** — 2026-06-01 통합 테스트 ISSUE-04. **2026-06-11 정합** (sprint `chore/work_260611-d-n9-status-align`): 본 carve 의 endpoint `GET /api/v1/repositories/:repository_id/build-runs` + `platformStoreOrUnavailable` 가드 + `ListRepositoryBuildRuns` (postgres.go `repository_ops.go`, `started_at DESC, id DESC` 정렬) + UT 3건 (`TestRepositoryBuildRuns_ReturnsCIRunsData` / `_NoCIRuns` / `_StatusFilter`) + IT 1건 (`TestIntegration_ListRepositoryBuildRuns_StatusFilter`) + openapi.yaml §repositories/build-runs 정의 + frontend `repositoryService.getRepositoryBuildRuns` + `DeveloperView` 위젯 (`buildRuns.length === 0` empty state + "View Logs" modal) + e2e `repository-dashboard.spec.ts` 의 inline build-runs mock 검증 모두 main 반영. **본 endpoint 의 path/query param 차이는 main 컨벤션 정합으로 의도된 차이** — `/repos/{id}` 약칭 → `/repositories/:repository_id` 풀패스, `cursor` ISO8601 → `offset` integer (repository list 와 정합). **status**: ✅ resolved (기본 구현, 2026-06-11). **잔여 4건 sub-issue 분리** (v1.1 milestone 진입 시점, 별도 carve): **(1)** RBAC 403/404 가드 (backend) — `platformStoreOrUnavailable` 외에 repository 존재/RBAC 검증; **(2)** Histogram metric `devhub_repository_build_runs_query_duration_seconds{status_filter}` (backend); **(3)** TanStack Query `useRepositoryBuildRuns` hook + status filter dropdown (queued/running/success/failed/cancelled/skipped/unknown) + skeleton + 무한 스크롤 cursor pagination (frontend); **(4)** dashboard widget "Recent repository activity" 통합 (manager/admin dashboard, frontend) + 독립 e2e spec `tests/e2e/repository-build-runs.spec.ts` (TC-E2E-BUILD-RUNS-01). | BE+FE | Claude+Gemini |
</input>
| **N-10** | **Manager role RBAC 검증** — E2E seed `bob` (team_manager) 의 권한 scope 확인 + ListProjects/ListPlatforms row filter + org unit subtree scope 검증. 검증 보고서 [docs/validation/N-10-manager-rbac.md](../validation/N-10-manager-rbac.md) (2026-06-04) — V-01..V-10 결과 + P1 follow-up 1건 (E2E spec-vs-구현 갭 6 TC: `TC-RBAC-ROW-READ-01/02` + `TC-RBAC-LOGOUT-01/02` + `TC-RBAC-ROLE-DRIFT-01` + `TC-RBAC-CODE-01` + `TC-RBAC-TRACE-01`). **2026-06-10 follow-up partial** (sprint `maintenance/work_260610-c-N10-rbac-e2e-tcs`): **E2E 4 TC active** — `TC-RBAC-LOGOUT-02` + `TC-RBAC-ROW-READ-01` + `TC-RBAC-ROW-READ-02` + `TC-RBAC-CODE-01` (frontend e2e spec `frontend/tests/e2e/rbac-data-scope.spec.ts`); **IT/UT 3 TC active** (`TC-RBAC-LOGOUT-01` / `TC-RBAC-ROLE-DRIFT-01` / `TC-RBAC-LEGACY-01` — backend `rbac_n10_integration_test.go` PR #515 옵션 C, Keycloak drift 환경 CI 검증); **Process 1 TC active** (`TC-RBAC-TRACE-01` = PR review 단계 cross-ref 검증, main flat memory + `docs/traceability/report.md` + N-10 검증 보고서 1:1 cross-ref 완료). spec 문서와 1:1 매핑 + spec file header 주석에 scope-out 명시. **status**: ✅ resolved (full — E2E 4 + IT/UT 3 + Process 1 TC 모두 verified, 2026-06-11, sprint `chore/work_260611-b-residual-housekeeping-close`). | 테스트 | Sisyphus |
| **N-11** | **CI e2e + backend-integration job 복원** — PR #407 의 `.github/workflows/ci.yml` e2e/backend-integration `&& false` 해제 + 첫 PR 두 job 실행 확인. refactor stabilize (PR #418 머지 후 main 안정) 시점. GitHub [#419](https://github.com/ykylee/Devhub_example/issues/419). **2026-06-08 코드 복원**: PR #407 cleanup-recovery (4a1942e) 후속 squash merge 4건 (5f5fdba / 9395cd9 / ce8ce7c) 으로 `&& false` 2건은 코드 레벨 복원 완료. sprint 260608-a 의 PR (코멘트 갱신만, 코드 변경 0줄) 으로 운영 정합. **2026-06-09 잔여 DoD 해소**: N-8 hotfix 4차 (PR #502 + #503) 머지 시점에 main 첫 PR 의 e2e shard 1/2/3 모두 PASS (PR #505 의 commit `6cc208a` 의 e2e hotfix 적용 후에도 e2e shard 1..3 PASS 재확인). **N-11 자체는 ✅ `resolved` — 복원 완료 + 잔여 DoD 해소 (issue #419 close)`** 정합. **status**: ✅ resolved. | CI infra | 사용자 + Codex (sprint 260608-a OpenCode 코멘트 정합) → 사용자 (N-8 hotfix 4차 follow-up) |
| **N-12** | **dev-requests voc 도메인 + 9 field + in-app notification (ADR-0028)** — v1.0 출시 직전 sprint `feat/work_260612-a-dev-requests-voc-domain` (PR #514, ADR-0028 §3 결정). 외부 시스템 의 의뢰가 project 결정 전 도착하는 경우의 staging 공간 + 9 field (title/details/requester/req_department/assignee/dev_department/request_date/dev_schedule/external_ref) 매칭 + in-app notification (`/api/v1/me/notifications`). **scope**: 2 신규 table (dev_request_vocs + user_notifications) + dev_requests 4 column 확장 + 5 신규 API (voc 등록 / routing / 조회 + notification list / mark read). `(source_system, external_ref)` UNIQUE for idempotency. **status**: ✅ resolved (backend code + ADR 정합, 2026-06-12). | Backend | OpenCode (v1.0 출시 직전 sprint) |
| **N-13** | **project.inbound_source 자동 routing (ADR-0028 §6 carve a, post-MVP)** — v1.0 출시 후 v1.1 milestone 진입 시점 sprint. ADR-0028 §3 의 voc 흐름 (외부 → voc 등록 → system_admin routing → dev-request 등록) 의 post-MVP 자동화. `applications.inbound_source_type` (gitea/jira/other) + `inbound_source_config` (JSONB) 컬럼 + sync 자동 routing 로직. **scope**: 1 migration (000007) + 1 schema 확장 + 1 API (PATCH /platforms inbound_source) + 1 worker (없음, synchronous routing) + 1 E2E + ADR-0028 §6 amendment. **2026-06-11 정공법** (sprint `feat/work_260611-a-n13-inbound-source-housekeeping`, docs only): ID slot 9 row 발급 + `conventions.md §1` RM 표기 정책 확장 (도메인 prefix 관행 `RM-{domain}-{nn}` 명문화) + `docs/traceability/report.md` §2.1/§2.1.5/§2.2/§2.3/§2.4/§2.5/§2.6 + §3 dev-request row + §4 ADR 인덱스 ADR-0028 row + §6 변경 이력 + 헤더 메타 정합. **status `⏳ planned` (housekeeping 정공법 완료, 구현은 v1.1 milestone 진입 시점 별도 sprint `feat/work_260611-a-n13-inbound-source-impl`)**. 정공법 + ID slot + 의존: [`docs/planning/2026-06-12-inbound-source-routing-sprint-plan.md`](2026-06-12-inbound-source-routing-sprint-plan.md) — 옵션 A 권장 (applications.inbound_source 컬럼 + sync 자동 routing). 검증 보고서 link = N/A (N-10 / N-13 은 code-level 정합 검증 = 구현 sprint 에서 별도).

#### NEXT — v0.1.1-alpha 운영화 + 외부 연동 깊이 정착

| ID | 아이템 | 영역 |
| --- | --- | --- |
| **X-1** | System Admin 운영 대시보드 (RM-M4-07 — Gitea sync job 큐/상태 + provider health) | FE+BE |
| **X-2** | inbound webhook 정규화 깊이 (multi-provider sync 일반화) | BE |
| **X-3** | 평문 secret envelope 암호화 (#6 — credentials_ref/api_token/auth_secret DEK + 키관리 ADR) | BE/보안 |
| **X-4** | Phase D — project 생성 flow ↔ SCM create 연계 | FE+BE |
| **X-5** | Gitea Hourly Pull 정밀화 (RM-M4-06 잔여, issue #231) | BE |
| **X-6** | Keycloak group staging-prod 적용 (P1-3, issue #214) | 사내 |
| **X-7** | ADR-0016 §6 alert 임계 확정 (P2-2) | Codex |
| **X-8** | Keycloak SPI realm events push 전환 (P2-6/P3-5) | BE/사내 |

#### LATER — v2 확장 (§3.4 P3 + 신규)

E-1 Realtime event publish(RM-M4-01) · E-2 WS replay/필터(RM-M4-02) · E-3 AI Gardener gRPC(RM-M4-04/05) · E-4 Weekly report worker · E-5 PermissionCache LISTEN/NOTIFY(RM-M4-08) · E-6 외부 SSO federation(RM-M4-09) · E-7 HomeLab dedicated worker + dedup(ADR-0015 §6) · E-8 Keycloak HA Phase 2.

## 4. 마일스톤 재정의

### 4.1 M-v1.0 — 1차 릴리즈 (target: 2026-06-15)

**완료 정의**: §1.3 의 8개 DoD 항목 모두 PASS.

**구성 sprint** (예상):
- sprint -f: P0-1 sub-carve B (`/api/v1/accounts/*` 폐기 + lazy auto-create + frontend cleanup) — Claude+Gemini 분담
- sprint -g: P0-3 Playwright screenshot mode + P0-2 UI polish 진입 — Codex+Gemini
- sprint -h: **P0-4 CI Run 생성 API (ISSUE-05)** + P1-1 sub-carve C event listener — Claude (P0-4 우선)
- sprint -i: P1-2 sub-carve D JWKS expiry + P1-6 Sign-out endpoint (BUG-03) + P1-7 Build-runs (ISSUE-04) — Claude
- sprint -j: P2-1 sub-carve E governance SOP + UI polish 마무리 — Codex+Gemini
- sprint -k: v1.0 e2e 종합 검증 + 운영 환경 1주 monitoring — 전 워커. **N-11 잔여 DoD 흡수 완료 (2026-06-09, N-8 hotfix 4차 머지 + PR #505 e2e shard 1..3 PASS 재확인)**, N-6 staging 1주 운영 검증만 잔여.

### 4.2 M-v1.1 — 안정성 + 운영 강화 (target: 2026-07-31)

| 항목 | priority | 워커 |
| --- | --- | --- |
| P1-3 group staging-prod 적용 | P1 | 사용자+Codex |
| ~~P1-4 off-boarding cron deploy~~ | **cancelled (2026-05-20)** | — |
| P1-5 e2e Keycloak admin 실 코드 전환 | P1 | Gemini+Codex |
| P2-2 ADR-0016 §6 alert 임계 | P2 | Codex |
| ~~P2-3 ADR-0017 §6 (b) PATCH expires_at~~ | **✅ resolved** (PR #137 + issue #219 closed 2026-05-21) | — |
| P2-4 Bindings UI 강화 | P2 | Gemini |
| P2-5 React Flow group + WebSocket 실시간 | P2 | Gemini+Claude |
| P2-6 Keycloak SPI provider JAR | P2 | Codex+사용자 |
| ~~P2-7 HRDB ETL unit pre-stage~~ | **cancelled (2026-05-21, issue #223)** | — |
| ~~P2-8 RM-ONBOARD-01 IMPL-backend~~ | **✅ resolved** (PR #278) | Claude |
| ~~P2-9 RM-ONBOARD-02 IMPL-frontend~~ | **✅ resolved** (PR #288) | Claude (override) |
| ~~P2-10 RM-ONBOARD-03 IMPL-admin~~ | **✅ resolved** (PR #288) | Claude (override) |
| ~~P2-11 RM-ONBOARD-04 IMPL-tests~~ | **✅ resolved** (PR #289 + #291) | Claude+Gemini |
| ~~P2-12 lazy_auto_create.go deletion~~ ([#284](https://github.com/ykylee/Devhub_example/issues/284)) | **✅ resolved** (PR #290, issue closed) | Claude |
| P1-6 Sign-out endpoint (BUG-03) | **신규 P1** | Claude |
| P1-7 Repository build-runs endpoint (ISSUE-04) | **신규 P1** | Claude+Gemini |
| P3-1 sub-carve F `/login` 정리 | P3 | Gemini |
| ~~P3-12 Sign Up 셀프 가입~~ | **cancelled (2026-05-20)** | — |

### 4.3 M-v2 — 확장 기능 (target: 2026-Q3 이후)

| 항목 | 분류 |
| --- | --- |
| P3-2 ADR-0015 §6 (3) dedicated worker | 운영 분리 |
| P3-3 ADR-0015 §6 (4) push/pull dedup | 별도 ADR |
| P3-4 HA Phase 2 (별도 ADR 후보 — ADR-0021 은 Onboarding 으로 발급됨 2026-05-21) | 인프라 |
| P3-5 audit event listener SPI push 전환 | latency |
| P3-6 RM-M4-01..03 WebSocket 확장 | 실시간 |
| P3-7 RM-M4-04..05 AI Gardener | AI 보조 |
| P3-8 RM-M4-06 Gitea Hourly Pull | 외부 SCM |
| P3-9 RM-M4-07 System Admin 대시보드 | 운영 view |
| P3-10 RM-M4-08 PermissionCache LISTEN/NOTIFY | 다중 인스턴스 |
| P3-11 RM-M4-09 외부 SSO (identity broker) | federation |

## 5. 워커 분업 매트릭스 (요약, 상세는 [docs/governance/worker_division.md](../governance/worker_division.md))

> **2026-06-09 결정 — 본 §5 의 워커 분담 표는 전면 취소되었다.** 사용자 (Owner) 의 Claude/Codex 자유 이용 불가 결정으로 §5.1 영역별 분담 + §5.2 v1.0 sprint 별 분담 + §5.3 인계 SOP 모두 무효. 모든 신규 작업은 **어느 에이전트로든 자유롭게** 진행 가능. 자세한 사항 + 유지되는 정책은 [worker_division.md §0](../governance/worker_division.md) 참조.

### 5.1 영역별 분담 — **취소 (2026-06-09)**

| 워커 | 영역 | 비고 |
| --- | --- | --- |
| (전 워커 자유) | (분담 무효) | 사용자 결정 (2026-06-09) 으로 Claude/Codex/Gemini/OpenCode 의 영역별 분담 전면 취소. 본 표는 historical 보존 |

### 5.2 v1.0 sprint 별 워커 분담 권장 — **취소 (2026-06-09)**

| Sprint | 작업 | 분담 |
| --- | --- | --- |
| (전 sprint) | (전 작업) | **자유** — 사용자/에이전트가 단독 또는 조합 결정 |

### 5.3 인계 SOP (워커 간)

1. **Claude → Codex** — Claude 가 docker-compose / CI / Keycloak SOP design 작성 → Codex 가 실 구현 PR 발급. 인계 시 design doc + 작업 범위 + 검증 기준 명시.
2. **Claude → Gemini** — Claude 가 backend handler / API contract design → Gemini 가 frontend service + page + e2e PR 발급. 인계 시 API spec + 응답 schema + 영향 받는 page 명시.
3. **Codex → Claude** — Codex 가 외부 리뷰 P1/P2 발견 시 hotfix design 또는 issue 생성 → Claude 가 hotfix PR 처리.
4. **Gemini → Claude** — Gemini 가 frontend 작업 중 backend API 누락/부정합 발견 시 issue + 가능하면 spec proposal → Claude 가 backend 진입.

## 6. GitHub project + milestone 등록 plan

### 6.1 GitHub Project

`gh project create --title "DevHub v1.0 릴리즈" --owner ykylee` — 단일 project. 다음 column:
- **Backlog** — 미진입 issue
- **Ready** — 본 sprint 진입 대상
- **In Progress** — 활성 PR
- **Review** — self-review + codex review 중
- **Done** — 머지 완료

### 6.2 Milestones (gh milestone create)

| Milestone | due | 포함 carve |
| --- | --- | --- |
| `v1.0 Release` | 2026-06-15 | P0-1, P0-2, P0-3, P1-1, P1-2, P2-1, + v1.0 종합 검증 |
| `v1.1 Stability` | 2026-07-31 | P1-3, P1-5, P2-2, P2-4..P2-6, P2-8..P2-12, P3-1 (P1-4 off-boarding cron + P2-7 HRDB ETL + P3-12 Sign Up 은 cancelled — 외부 Keycloak 시나리오 + ADR-0021 self-service 로 무효화. **P2-3 ✅ resolved** PR #137 + adr0017-token-mutation. **P2-8 ✅ resolved** PR #278 Carve A backend.) |
| `v2 Extension` | 2026-Q3+ | P3-2..P3-11 (단 P3-13 MFA 는 사내 정책 제외) |

### 6.3 Issue 발급 plan

각 carve (P0/P1/P2/P3) 별로 1 issue 발급. label:
- `priority/p0` `priority/p1` `priority/p2` `priority/p3`
- `worker/claude` `worker/codex` `worker/gemini` `worker/user` (사내 동반 carve) `worker/opencode` (OpenCode 워커, 2026-06-04 신설)
- `domain/auth` `domain/app-repo-project` `domain/dreq` `domain/integration` `domain/ui-polish` `domain/infra`
- `type/feature` `type/refactor` `type/test` `type/docs` `type/ci`

본 PR 머지 후 별도 sprint (sprint -g-issues 또는 본 sprint -f-roadmap 의 후속 작업) 에서 `gh issue create` 일괄 발급.

## 7. 다음 sprint 진입 순서

본 PR (sprint -f-roadmap) 머지 후 권장 진입 순서:

1. **sprint -g**: P0-3 Playwright screenshot mode 도입 (Codex CI + Gemini frontend fixture) — UI 검증 자산 확보 우선. 작은 변경, 위험 낮음
2. **sprint -h** (or 동시): P0-1 sub-carve B (`/api/v1/accounts/*` 폐기) + **P0-4 CI Run 생성 API** — Claude backend 우선 처리. P0-4는 CI 기능의 v1.0 출시 차단 요건
3. **sprint -i**: P0-2 UI polish 1차 (Gemini 주도) + P1-1 sub-carve C (Claude) + **P1-6 Sign-out (BUG-03)** — 영역 분리로 충돌 없음
4. **sprint -j**: P1-2 sub-carve D + P2-1 sub-carve E + **P1-7 Build-runs (ISSUE-04)** (Claude + Codex + Gemini)
5. **sprint -k**: v1.0 종합 검증 + staging 1주 monitoring — 전 워커

## 8. UI 검증 방식 — Playwright screenshot mode

### 8.1 도입 계획

`frontend/playwright.config.ts` 의 `use.screenshot` 설정 + 신규 `tests/e2e/screenshots.spec.ts`:

```typescript
test.describe("UI screenshot capture", () => {
  const pages = [
    { path: "/login", name: "login" },
    { path: "/", name: "dashboard-developer" }, // role-routed
    { path: "/admin", name: "admin-dashboard" },
    { path: "/admin/settings/users", name: "admin-users" },
    { path: "/admin/settings/organization", name: "admin-organization" },
    { path: "/admin/settings/permissions", name: "admin-permissions" },
    { path: "/admin/settings/audit", name: "admin-audit" },
    { path: "/admin/settings/dev-requests", name: "admin-dev-requests" },
    { path: "/admin/settings/dev-request-tokens", name: "admin-dev-request-tokens" },
    { path: "/admin/settings/integrations", name: "admin-integrations" },
    { path: "/admin/settings/integration-bindings", name: "admin-integration-bindings" },
    { path: "/admin/settings/platforms", name: "admin-applications" },
    { path: "/admin/topology-v2", name: "admin-topology-v2" },
    { path: "/platforms", name: "user-applications" },
    { path: "/repositories", name: "user-repositories" },
    { path: "/projects", name: "user-projects" },
    { path: "/dev-requests", name: "user-dev-requests" },
    { path: "/account", name: "user-account" },
  ];
  for (const { path, name } of pages) {
    test(`screenshot ${name}`, async ({ page }) => {
      await page.goto(path);
      await page.screenshot({ path: `test-results/screenshots/${name}.png`, fullPage: true });
    });
  }
});
```

### 8.2 CI artifact 업로드

`.github/workflows/ci.yml` 의 e2e job 에 `actions/upload-artifact` 추가 (name: `ui-screenshots-shard-N`, path: `frontend/test-results/screenshots/`).

### 8.3 디자인 review flow

1. Gemini (or Claude review mode) 가 screenshots 검토
2. 발견된 design issue 를 `domain/ui-polish` label issue 로 발급
3. P0-2 UI polish sprint 에서 일괄 처리

## 9. 변경 이력

| 일자 | 변경 | sprint |
| --- | --- | --- |
| 2026-05-27 | **코드베이스 스냅샷 정합** — Onboarding IMPL carve A/B/C/D 전부 ✅ resolved (§2.3 + §3.3 P2-8..12 + §4.2, PR #278/#288/#289/#290/#291, lazy_auto_create 폐기 issue #284 closed) + v1.1 영역 작업 v1.0 전 선행 명시(Gitea SCM sync·SCM↔시스템 양방향·auth_mode full·repository draft/publish·admin catalog) + §3.5 신규 도출 백로그(N-1..6 / X-1..8 / E-1..8, 코드베이스 스냅샷 §06 연계) 추가. 헤더 날짜 2026-05-21→2026-05-27. | `claude/work_260527-codebase-review-roadmap-refresh` |
| 2026-05-20 | 1차 작성 — v1.0 scope 정의 (3 domain) + 잔여 carve 통합 인벤토리 (P0/P1/P2/P3, 30+ item) + 마일스톤 재정의 (M-v1.0 / M-v1.1 / M-v2) + 워커 분업 매트릭스 (Claude=backend / Codex=infra+CI / Gemini=frontend+UX) + GitHub project + milestone plan + UI Playwright screenshot mode | `claude/work_260520-f-roadmap` |
| 2026-05-20 | **P3-12 Sign Up 영구 취소** — 사용자 결정. DevHub 가 Keycloak admin 권한이 없는 외부 IdP 운영 시나리오 (ADR-0020 결정 A 정합). user 생성은 IdP 팀 admin console 또는 HRDB ETL push 책임. §1.2 제외 기능 표 분류 'v1.1 carve' → 'permanently cancelled' + §3.4 P3-12 strikethrough + §5.2 워커 분담 표 strikethrough + §4.2 v1.1 milestone 본문 정정 + issue #235 closed | `claude/work_260520-i-cancel-signup` |
| 2026-06-01 | **통합 테스트 결과 반영** — §3.1 P0-4 (CI Run API) 신규 P0 carve 추가 + §3.2 P1-6 (Sign-out) P1-7 (Build-runs) 신규 P1 carve 추가 + §3.5 N-7~N-10 신규 도출 백로그 반영 + §4.1 sprint 구성 P0-4 포함 재조정 + §4.2 v1.1 milestone 신규 P1 항목 추가 + §7 sprint 진입 순서 갱신. 출처: `deepseek/test-scenarios-20260601` 브랜치 통합 테스트 | `deepseek/test-scenarios-20260601` |
| 2026-06-04 | **OpenCode 워커 부트스트랩** — §5.1 분담 표에 OpenCode 행 추가 (영역 TBD, bootstrap) + §6.3 label 에 `worker/opencode` 추가 + 헤더 메타의 워커 목록/대상 독자 갱신 | `opencode/work_260604-a-opencode-workflow-bootstrap` |
| 2026-06-04 | **OpenCode Lane 정의** — §5.1 OpenCode 행의 TBD 를 3-lane (Workflow curation / Cross-cutting validation / AI/ML prep) 으로 확정 + §9 본 row 직전 | `opencode/work_260604-b-opencode-areas` |
| 2026-06-04 | **N-10 Manager RBAC 검증** — §3.5 N-10 row 의 mgr-user-b 비공식 명칭 → E2E seed `bob` (team_manager) 으로 정정 + 검증 보고서 링크 + 1 P1 follow-up 식별. 검증 결과: backend UT 25 packages PASS, row filter SQL 정상, E2E spec-vs-구현 갭 6 TC (TC-RBAC-ROW-READ-01/02, TC-RBAC-LOGOUT-01/02, TC-RBAC-ROLE-DRIFT-01) 발견. 출처: `opencode/work_260604-c-N10-manager-rbac-validation` | `opencode/work_260604-c-N10-manager-rbac-validation` |
| 2026-06-06 | **P0/P1 미발급 carve 일괄 발급 + ID 슬롯 정합** — §3.5 N-7/8/9 GitHub issues [#486](https://github.com/ykylee/Devhub_example/issues/486) / [#487](https://github.com/ykylee/Devhub_example/issues/487) / [#488](https://github.com/ykylee/Devhub_example/issues/488) 발급 (P0-4 / P1-7 / P1-6 carve 의 정식 ID). §3.1 P0-4 row 에 ID 슬롯 정합 노트 추가 (historical [#238](https://github.com/ykylee/Devhub_example/issues/238) 유지, 정식 ID = §3.5 N-7). §3.5 N-11 (CI e2e + backend-integration job 복원, [#419](https://github.com/ykylee/Devhub_example/issues/419)) + §3.4 P3-14 (RM-APPDASH-01, [#384](https://github.com/ykylee/Devhub_example/issues/384)) 신규 row 추가. §3.2 P1-3 에 GitHub [#214](https://github.com/ykylee/Devhub_example/issues/214) cross-ref 추가. 출처: Mavis housekeeping `mvs/work_260606-a-roadmap-issue-bulk` | `mvs/work_260606-a-roadmap-issue-bulk` |
| 2026-06-08 | **N-11 CI e2e + backend-integration job 복원 완료 정합** — §3.5 N-11 row 의 `open` → `복원 완료, 잔여 DoD = main 첫 PR 두 job PASS` 정합. §4.1 v1.0 sprint 구성의 sprint -k 에 N-11 잔여 DoD 흡수 명시. §9 본 row 추가. 코드 레벨 복원은 PR #407 cleanup-recovery (4a1942e) 후속 squash merge 4건 (5f5fdba / 9395cd9 / ce8ce7c) 으로 선행 완료, 본 sprint -a 의 PR 은 코멘트 갱신 (코드 변경 0줄). §6 변경 이력 + traceability report.md §6 cross-ref. docs-only 정합 — 신규 ID 발급 없음 (cross-cutting infrastructure 운영 항목). | `opencode/work_260608-a-N11-ci-restore-verify` |
| 2026-06-09 | **워커 분업 전면 취소 (사용자 결정, Claude/Codex 자유 이용 불가)** — §0 사용 가이드에 취소 노트 추가 + §5 워커 분담 매트릭스 (5.1 영역별 분담 + 5.2 v1.0 sprint 별 분담) 전면 취소 標記. worker_division.md §0 + §1~§4 historical 標記 + §2.5 branch prefix 자유화 + §5 Owner 권한 명시. 유지 정책: §4.2 ADR reversal 의 supersession 정공법, §5 Owner 의 invoke 책임/결정 권한/review 최종 승인, 우선순위 P0~P3 강제. AGENTS.md "워커 일반 메모 (2026-06-09 전면 갱신)" 신설 + 워커별 전용 메모를 historical 標記로 격하 + branch prefix 강제 해제. 신규 ID 발급 없음 (governance 운영 결정). | `maintenance/work_260609-a-cancel-worker-division` |
| 2026-06-09 | **N-11 메모리 sync (PR #499, 5 파일 +36/-15) + §3.5 N-11 row 정합** — release_v1_roadmap.md §3.5 N-11 row 의 `잔여 DoD` 갱신 + §4.1 sprint -k (N-11 잔여 DoD 흡수) + §9 본 row 추가. 메모리 4종 (state/session_handoff/work_backlog/traceability report.md) 의 main HEAD `da7d57e` 정합. 충돌 해소 (N-11 row + 워커 분업 row 둘 다 보존). 신규 ID 없음 (cross-cutting infrastructure 운영 정합). | `opencode/work_260608-b-N11-memory-sync` |
| 2026-06-09 | **N-8 sign-out e2e deterministic race hotfix 4차 (PR #502 + #503, 3 commit, issue #501 closed)** — §3.5 N-8 race row 신규 추가 (정공법 3-step: 502→204 + X-Keycloak-Likely-Down header + typed error sentinel ErrOIDCConfigMissing/ErrOIDCNetworkUnreachable). §3.5 N-11 row 의 "잔여 DoD" → "✅ 해소" 갱신 (N-8 hotfix 4차 머지 시점에 main 첫 PR e2e shard 1..3 모두 PASS). backend 35 packages PASS, frontend 80 files / 1033 tests PASS, 4 신규 TC (TC-AUTH-LOGOUT-04/-08, TC-AUTH-LOGOUT-FE-07/-08). 신규 ID 없음 (N-8 follow-up hotfix). main HEAD `897953c`. issue #501 close. | `fix/work_260609-{a-N8-logout-graceful-degrade,b-N8-codex-p1-distinguishable}` |
| 2026-06-10 | **v1.0 출시 직전 housekeeping — §3.5 N-8/N-11 row close + §4.1 sprint -k 마킹 + §3.5 N-10 follow-up 6 TC 명시.** (1) §3.5 **N-8 race row status** `✅ resolved (issue #501 closed)` 명시. (2) §3.5 **N-11 row** "잔여 DoD" → "✅ 잔여 DoD 해소" + **status `✅ resolved`** (PR #502 + #503 + PR #505 e2e shard 1..3 PASS 재확인 정합). (3) §4.1 sprint -k 의 N-11 잔여 DoD 완료 마킹 (N-6 staging 1주 운영만 잔여). (4) §3.5 N-10 row 의 follow-up 6 TC 명시 (`TC-RBAC-ROW-READ-01/02` + `TC-RBAC-LOGOUT-01/02` + `TC-RBAC-ROLE-DRIFT-01` + `TC-RBAC-CODE-01` + `TC-RBAC-TRACE-01`) + follow-up sprint `maintenance/work_260610-c-N10-rbac-e2e-tcs` 진입 예정. (5) [docs/traceability/report.md §4](../traceability/report.md) ADR-0025 (envelope encryption, 2026-05-27 PR #350) + ADR-0026 (Keycloak role excluded decision, 2026-06-02 `deepseek/work_260602`) row 보강 — swagger UI 1차 bootstrap sprint 가 note 만 남겼던 ID slot 누락 해소. (6) §9 본 row 추가. 신규 ID 발급 없음 (governance housekeeping). main HEAD `ad8d481` (PR #505). | `maintenance/work_260610-b-v1-pre-release-housekeeping` |
| 2026-06-12 | **v1.0 출시 직전 finalizing — N-12 (voc + notification) + housekeeping + N-10 잔여 backend IT** (sprint `maintenance/work_260612-b-v1-finalizing`). (1) §3.5 N-12 row 신규 + status `✅ resolved` 마킹 — dev-requests voc 도메인 + 9 field + in-app notification (ADR-0028) 정합. (2) traceability/report.md §2.4 에 IMPL-voc-01 + IMPL-notification-01 cross-cutting row 추가 + §6 변경 이력 본 sprint row 1건. (3) [voc list API (ADR-0028 §6 carve (d))] — `GET /api/v1/vocs?status=received&assignee=&limit=&offset=` 신규 (system_admin 도구, N-6 staging 운영 SOP 정합). (4) [follow-up N-10 backend IT 3 TC] — `TC-RBAC-LOGOUT-01` + `TC-RBAC-ROLE-DRIFT-01` + `TC-RBAC-LEGACY-01` backend integration test. §9 본 row 추가. 신규 ID 발급 없음 (voc + notification 정합 + housekeeping). main HEAD `ba7823f` (PR #514). | `maintenance/work_260612-b-v1-finalizing` |
| 2026-06-12 | **v1.0 출시 직전 option D 검토 — project.inbound_source 자동 routing (ADR-0028 §6 carve a) sprint plan** (sprint `maintenance/work_260612-c-inbound-source-plan`). (1) §3.5 N-13 row 신규 + status `⏳ planned` (v1.1 milestone 진입 시점 구현). (2) [`docs/planning/2026-06-12-inbound-source-routing-sprint-plan.md`](2026-06-12-inbound-source-routing-sprint-plan.md) 신규 — 옵션 A/B/C 비교 + 옵션 A 권장 결정 (applications.inbound_source 컬럼 + sync 자동 routing) + ID slot 8건 (REQ-FR-113 + UC-DEV-REQ-15 + ARCH-23 + API-103 + RM-DEV-REQ-15 + IMPL-inbound-source-01 + IMPL-platform-patch-02 + UT-inbound-source-01 + TC-INBOUND-SRC-01) + 정공법 10 step (~9시간, 1 sprint). (3) §9 본 row 추가. 신규 ID 미발급 (v1.1 milestone 진입 시점 발급, 본 sprint = post-MVP 검토). main HEAD `f7d2705` (PR #515). | `maintenance/work_260612-c-inbound-source-plan` |
| 2026-06-11 | **N-13 P1 follow-up 정공법 — ID slot 9 row 발급 + conventions.md §1 RM 표기 확장 (docs only, post-MVP)** (sprint `feat/work_260611-a-n13-inbound-source-housekeeping`). (1) `conventions.md §1` 의 RM 표기 정책 확장 — `RM-M{n}-{nn}` legacy milestone + `RM-{domain}-{nn}` 도메인 prefix (예: `RM-ONBOARD-01..04` / `RM-APPDASH-01` / `RM-DEV-REQ-15`) 두 관행 혼용 명문화. (2) `docs/traceability/report.md` 에 N-13 ID slot 9 row 발급 — `REQ-FR-113` / `UC-DEV-REQ-15` / `ARCH-23` / `API-103` / `RM-DEV-REQ-15` / `IMPL-inbound-source-01` / `IMPL-platform-patch-02` / `UT-inbound-source-01` / `TC-INBOUND-SRC-01` (sprint `feat/work_260611-a-n13-inbound-source-impl` v1.1 진입 시점 코드 변경). (3) `docs/adr/0028-dev-requests-voc-external-ref.md` §6 (a) 의 ID slot 발급 사실 + 도메인 prefix 표기 관행 정합 + §7 변경 이력 1 row. (4) 본 §3.5 N-13 row 의 status `⏳ planned` 보강 + housekeeping 정공법 완료 마킹. (5) §9 본 row 추가. **신규 ID 발급 9 row, 코드 변경 0줄**. main HEAD `f37305d7` (PR #545 머지 후). | `feat/work_260611-a-n13-inbound-source-housekeeping` |
| 2026-06-11 | **N-9 (P1-7 Repository build-runs) 기본 구현 완료 정합 + 잔여 4건 sub-issue 분리 (docs only)** (sprint `chore/work_260611-d-n9-status-align`). (1) §3.5 **N-9 row status `✅ resolved (기본 구현, 2026-06-11)`** — 본 carve 의 endpoint `GET /api/v1/repositories/:repository_id/build-runs` (router.go:509) + `platformStoreOrUnavailable` 가드 + `ListRepositoryBuildRuns` (postgres.go `repository_ops.go`) + UT 3건 + IT 1건 + openapi.yaml §repositories/build-runs 정의 + frontend `repositoryService.getRepositoryBuildRuns` + `DeveloperView` 위젯 + e2e `repository-dashboard.spec.ts` 의 inline build-runs mock 검증 모두 main 반영. (2) §3.2 **P1-7 row 비고** 갱신 — `신규 P1` → `기본 구현 완료` + endpoint path 차이 (`/repos/{id}` → `/repositories/:repository_id`, main 컨벤션) 명시 + 잔여 4건 sub-issue 분리. (3) **잔여 4건 sub-issue 분리** (v1.1 milestone 진입 시점, 별도 carve): **(a)** RBAC 403/404 가드 (backend); **(b)** Histogram metric `devhub_repository_build_runs_query_duration_seconds{status_filter}` (backend); **(c)** TanStack Query `useRepositoryBuildRuns` hook + status filter dropdown (queued/running/success/failed/cancelled/skipped/unknown) + skeleton + 무한 스크롤 cursor pagination (frontend); **(d)** dashboard widget "Recent repository activity" 통합 (manager/admin dashboard, frontend) + 독립 e2e spec `tests/e2e/repository-build-runs.spec.ts` (TC-E2E-BUILD-RUNS-01). (4) issue [#487](https://github.com/ykylee/Devhub_example/issues/487) (정식 ID = N-9) — close + 잔여 4건 sub-issue 별도 발급 (PR 머지 후). 본 sprint = docs only 정합 — 신규 ID 발급 없음. | `chore/work_260611-d-n9-status-align` |
| 2026-06-11 | **P2 잔여 5건 일괄 처리 계획 수립 및 이슈 close (docs only)** (sprint `chore/work_260611-e-p2-residual-sprint-plan`). (1) `docs/planning/2026-06-11-p2-residual-sprint-plan.md` 신설하여 P2 잔여 이슈 5건 (#386, #382, #381, #380, #383)의 작업 범위 및 검증 기준 구체화. (2) 위 5건 이슈를 모두 close 처리하고, 후속 스프린트에서 계획을 참조하여 실 구현하도록 정합. 신규 ID 발급 0건. 코드 변경 없음. | `chore/work_260611-e-p2-residual-sprint-plan` |
</input>
</input>
