# DevHub v1.0 릴리즈 로드맵 — 1차 릴리즈 scope + 우선순위 + 워커 분업

- 문서 목적: DevHub 의 1차 릴리즈 (v1.0) 를 위한 단일 source-of-truth 로드맵. 잔여 carve 통합 인벤토리 + 우선순위 P0~P3 + 마일스톤 재정의 + 워커(Claude/Codex/Gemini) 분업.
- 범위: 1차 릴리즈에 포함될 기능 scope, 제외 기능, 잔여 carve 우선순위, 신규 마일스톤(M-v1.0, M-v1.1, M-v2), GitHub project + milestone 등록 plan, UI 검증 방식.
- 대상 독자: 프로젝트 리드, 모든 워커 (Claude, Codex, Gemini), 후속 작업자.
- 상태: draft
- 최종 수정일: 2026-05-21 (Onboarding 도메인 §2.3 신규 + P2-8..11 IMPL carve 4건 추가 + ADR-0021 reservation 정정 (HA Phase 2))
- 결정 근거 sprint: `claude/work_260520-f-roadmap` (본 문서)
- 관련 문서: [통합 개발 로드맵](../development_roadmap.md) (M0~M6 historical), [requirements](../requirements.md), [architecture](../architecture.md), [ADR-0019 Keycloak](../adr/0019-keycloak-only-idp.md), [ADR-0020 계정/사용자 책임 경계](../adr/0020-account-user-management-boundary.md), [traceability matrix](../traceability/report.md), [account_user_management_redesign Phase 1/2/3](./account_user_management_redesign.md), [keycloak_operations](../setup/keycloak_operations.md).

## 0. 사용 가이드

본 문서는 **2026-05-20 이후 모든 sprint 의 진입점**. 기존 [`docs/development_roadmap.md`](../development_roadmap.md) (M0~M6 historical) 는 done milestone 의 사후 명문화 자산으로 보존 — 본 문서가 v1.0 / v1.1 / v2 의 새 source-of-truth.

1. **신규 sprint 진입 전** §3 우선순위 매트릭스 확인 → P0 carve 부터 흡수
2. **워커 작업 분담** §5 매트릭스 참조 — 본 워커의 영역 + 인계 대상 확인
3. **마일스톤 진행** §4 의 M-v1.0/M-v1.1/M-v2 표에서 본 sprint 의 마일스톤 + 동반 issue 확인
4. **결정 변경** 발생 시 §9 변경 이력에 row 추가

## 1. v1.0 릴리즈 scope

### 1.1 포함 기능 (3 domain)

| Domain | scope | 현황 (2026-05-20) |
| --- | --- | --- |
| **인증/조직/대시보드** | Keycloak OIDC 로그인 + `/admin/settings/*` (users/organization/permissions + audit + dev-requests + dev-request-tokens + integrations + integration-bindings) + dashboard (developer/manager/admin) + 역할 routing | Backend done (M1/M2/M3). Frontend done. **계정/사용자 리팩토링 Phase 3 sub-carve A done** (ADR-0020, sprint -d). sub-carve B (`/api/v1/accounts/*` 폐기 + lazy auto-create) 가 v1.0 의 마지막 큰 backend 변경. |
| **Application/Repository/Project 도메인** | Application 등록/조회 + Repository 연결 + Project CRUD + rollup + 현황 페이지 + SCM provider catalog | Backend done (API-01..58 + ADR-0011, 2026-05-14). Frontend 활성 (현황 페이지 + FilterBar). |
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
7. 사내 staging 환경 1주 운영 + 외부 사용자 ≥ 5명 로그인 동작
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

### 2.2 Application/Repository/Project 도메인

| 모듈 | 위치 (backend) | 위치 (frontend) | v1.0 상태 |
| --- | --- | --- | --- |
| Application CRUD + 상태 전이 + critical_warning guard | `internal/httpapi/applications*.go` + `internal/store/applications*.go` (API-41..47) | `app/(dashboard)/admin/settings/applications/page.tsx` + `app/(dashboard)/applications/{page,[id]/page}.tsx` | ✅ done (ADR-0011) |
| ApplicationRepository link CRUD | API-48..50 + `application_repositories` 테이블 | `components/project/ApplicationCreationModal.tsx` 등 | ✅ done |
| SCM Provider catalog | API-41/42 + 4 seed (bitbucket/gitea/forgejo/github) | provider 선택 UI | ✅ done |
| Project CRUD | API-55/56 + `projects` 테이블 | `app/(dashboard)/admin/settings/applications/...` + `app/(dashboard)/projects/{page,[id]/page}.tsx` | ✅ done |
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
| IMPL carve plan + RM-ONBOARD-01..04 | [`docs/planning/onboarding_impl_plan.md`](./onboarding_impl_plan.md) | — | 본 sprint done |
| RM-ONBOARD-01 IMPL-backend (handler + middleware + migration + lazy 폐기) | `internal/httpapi/{onboarding_gate,me_onboarding,organizations_search,users_admin_review}.go` 신규 | — | ⏳ M-v1.1 |
| RM-ONBOARD-02 IMPL-frontend (page + picker + banner + gating) | — | `app/onboarding/page.tsx` + `components/onboarding/*` + `(dashboard)/layout.tsx` 확장 + `account/page.tsx` 확장 | ⏳ M-v1.1 |
| RM-ONBOARD-03 IMPL-admin (Confirm Review + pending_review filter) | — | `app/admin/settings/users/page.tsx` 확장 + `ConfirmReviewModal.tsx` 신규 | ⏳ M-v1.1 |
| RM-ONBOARD-04 IMPL-tests (UT + E2E mega lifecycle) | `internal/httpapi/onboarding{,_gate}_test.go` | `tests/e2e/onboarding.spec.ts` + 6 test seed | ⏳ M-v1.1 |

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

### 3.2 P1 — v1.0 안정성 (sprint -g/-h)

| ID | Carve | 출처 | 워커 | 비고 |
| --- | --- | --- | --- | --- |
| **P1-1** | ADR-0020 sub-carve C — Keycloak event listener 확장 (USER:UPDATE / GROUP_MEMBERSHIP / USER:DELETE 매핑) + DevHub `users` write + metric 3종 | ADR-0020 §4.1 C, design doc §5.3 | **Claude (backend)** | sprint -u~-y 자연 확장. P0-1 의 lazy auto-create 와 role 추출 로직 공유 |
| **P1-2** | ADR-0020 sub-carve D — JWKS stale-while-error expiry case 확장 | ADR-0020 §4.1 D | **Claude (backend)** | sprint -r kid mismatch fallback 자연 확장. Keycloak unreachable 시 uptime 보장 |
| **P1-3** | ADR-0019 §5.3 — Keycloak group staging-prod 적용 | session_handoff 잔여 carve | **사용자 + Codex** | Keycloak admin 1회 작업 (group 4 + composite role assign) |
| ~~**P1-4**~~ | ~~ADR-0019 §5.3 — off-boarding Phase 1 cron 실 deploy~~ | ~~session_handoff 잔여 carve~~ | — | **permanently cancelled (2026-05-20, issue #215 close)** — 외부 Keycloak 시나리오 채택. HR ↔ Keycloak sync 는 외부 IdP 팀 책임. DevHub off-boarding sync 는 sub-carve C event listener (PR #241) 가 정공법. `scripts/hrdb_etl_sync.sh` deprecation. |
| **P1-5** | ADR-0019 §5.3 — e2e Kratos → Keycloak 실 코드 전환 | session_handoff 잔여 carve | **Gemini (frontend test) + Codex (CI infra)** | sprint -m design 따름. 사내 staging Keycloak e2e 환경 동반. PR #203 의 `ci-e2e-sync-check.sh` 가 CI 단 일부 해소 |

### 3.3 P2 — v1.0 운영 안정성 + v1.1 carve

| ID | Carve | 출처 | 워커 | 비고 |
| --- | --- | --- | --- | --- |
| **P2-1** | ADR-0020 sub-carve E — service account 권한 축소 (manage-users 제거) + governance SOP `keycloak_operations.md §8.5c` | ADR-0020 §4.1 E | **Codex (infra) + Claude (docs)** | docs only + Keycloak admin SOP. P0-1 머지 후 |
| **P2-2** | ADR-0016 §6 — pull latency p95 alert + push webhook 알림 + stage→prod 임계 확정 | ADR-0016 §6 (3)+(4)+(5) | **Codex (infra)** | baseline 1주 관찰 후 |
| **P2-3** | ADR-0017 §6 (b) — PATCH expires_at + admin UI 편집 modal | ADR-0017 §6 (b) | **Gemini (frontend) + Claude (backend)** | 정책 변경 동반 (token rotation 정책) |
| **P2-4** | Bindings UI 강화 — scope_id lookup combobox + Edit/Delete + pagination | development_roadmap §6 잔여 | **Gemini (frontend+UX done ✅)** | v1.0 UI polish 동반. backend CRUD 지원 추가 완료. |
| **P2-5** | React Flow group sub-node + WebSocket 실시간 (`infra.node.updated` / `infra.service.updated`) | development_roadmap §6 잔여 | **Gemini (frontend done ✅)** | topology v2 강화. WebSocket 실시간 연동 및 Environment 그룹화 완료. |
| **P2-6** | Keycloak SPI provider JAR (PR #203 codex P2) | PR #203 codex review | **Codex (infra) + 사용자 (Java 빌드 환경)** | `infra/idp/devhub-event-listener/` Maven 또는 Gradle 빌드 + compose volume mount + 운영 SOP |
| **P2-7** | 신규 user 의 unit 초기 배치 자동화 — HRDB ETL pre-stage 가 unit 정보 동반 | ADR-0020 §5.5.2 | **Claude (backend) + 사용자 (HRDB)** | `scripts/hrdb_etl_sync.sh` 확장 |
| **P2-8** | **RM-ONBOARD-01** IMPL-backend — `users` migration (`onboarding_completed_at` + `review_status` + CHECK) + `onboardingGate` middleware + 5 handler (API-83/84/85/86 + API-32/33 확장) + lazy_auto_create.go 폐기 + audit event const 3종 | ADR-0021 §6.1, [onboarding_impl_plan.md](./onboarding_impl_plan.md) §2.1 | **Claude (backend)** | feature flag default OFF — Carve A 단독 머지 후 main 안정성. Carve B/C 진입 dependency |
| **P2-9** | **RM-ONBOARD-02** IMPL-frontend — `/onboarding` page + OrganizationPicker (typeahead + tree) + skip flag sessionStorage + dismissible banner + `(dashboard)/layout.tsx` 3-branch gating + `/account` self-service unit edit | ADR-0021 §6.1, [onboarding_impl_plan.md](./onboarding_impl_plan.md) §2.2 | **Gemini (frontend+UX)** | Carve A 머지 후 진입. Carve C 와 병행 가능 |
| **P2-10** | **RM-ONBOARD-03** IMPL-admin — `/admin/settings/users` 의 "Confirm Review" 액션 + pending_review filter + `ConfirmReviewModal.tsx` | ADR-0021 §6.1, [onboarding_impl_plan.md](./onboarding_impl_plan.md) §2.3 | **Gemini (frontend)** | Carve A 머지 후 진입. Carve B 와 병행 가능 |
| **P2-11** | **RM-ONBOARD-04** IMPL-tests — UT-onboarding-* (backend handler + middleware) + TC-ONBOARD-* 11건 (E2E mega lifecycle, REQ-NFR-ONBOARD-008 의 6 test seed) + `docs/tests/test_cases_m7_onboarding.md` | ADR-0021 §6.1, [onboarding_impl_plan.md](./onboarding_impl_plan.md) §2.4 | **Claude (UT) + Gemini (E2E)** | Carve A + B + C 모두 머지 후 |

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

## 4. 마일스톤 재정의

### 4.1 M-v1.0 — 1차 릴리즈 (target: 2026-06-15)

**완료 정의**: §1.3 의 8개 DoD 항목 모두 PASS.

**구성 sprint** (예상):
- sprint -f: P0-1 sub-carve B (`/api/v1/accounts/*` 폐기 + lazy auto-create + frontend cleanup) — Claude+Gemini 분담
- sprint -g: P0-3 Playwright screenshot mode + P0-2 UI polish 진입 — Codex+Gemini
- sprint -h: P1-1 sub-carve C event listener 확장 — Claude
- sprint -i: P1-2 sub-carve D JWKS expiry + P2-1 sub-carve E governance SOP — Claude+Codex
- sprint -j: UI polish 마무리 (P0-2 의 후속) — Gemini
- sprint -k: v1.0 e2e 종합 검증 + 운영 환경 1주 monitoring — 전 워커

### 4.2 M-v1.1 — 안정성 + 운영 강화 (target: 2026-07-31)

| 항목 | priority | 워커 |
| --- | --- | --- |
| P1-3 group staging-prod 적용 | P1 | 사용자+Codex |
| ~~P1-4 off-boarding cron deploy~~ | **cancelled (2026-05-20)** | — |
| P1-5 e2e Keycloak admin 실 코드 전환 | P1 | Gemini+Codex |
| P2-2 ADR-0016 §6 alert 임계 | P2 | Codex |
| P2-3 ADR-0017 §6 (b) PATCH expires_at | P2 | Gemini+Claude |
| P2-4 Bindings UI 강화 | P2 | Gemini |
| P2-5 React Flow group + WebSocket 실시간 | P2 | Gemini+Claude |
| P2-6 Keycloak SPI provider JAR | P2 | Codex+사용자 |
| P2-7 HRDB ETL unit pre-stage | P2 | Claude+사용자 |
| **P2-8 RM-ONBOARD-01 IMPL-backend** (handler + middleware + migration + lazy 폐기) | P2 | Claude |
| **P2-9 RM-ONBOARD-02 IMPL-frontend** (page + picker + banner + gating + /account edit) | P2 | Gemini |
| **P2-10 RM-ONBOARD-03 IMPL-admin** (Confirm Review + filter) | P2 | Gemini |
| **P2-11 RM-ONBOARD-04 IMPL-tests** (UT + E2E mega lifecycle) | P2 | Claude+Gemini |
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

### 5.1 영역별 분담

| 워커 | 영역 | 작업 스타일 | 분량 (이전 sprint 누적) |
| --- | --- | --- | --- |
| **Claude** | **Backend (Go)** + **Design (ADR + docs)** | 큰 단위 design + 분담된 backend 구현 + 4단계 self-review + codex review 응답 | 30+ sprint, 60+ PR |
| **Codex** | **Infra (Docker/Nginx/CI) + Security + Build** | docker-compose 패키징 + GitHub Actions + Keycloak SPI infra + e2e CI 정합. 외부 리뷰 (P1/P2 inline) 가장 활발 | 7+ PR (packaging + reverse proxy + Keycloak refactor + CI sync guard) |
| **Gemini** | **Frontend + UX + Test fixtures + Design polish** | Next.js page + 컴포넌트 + e2e Playwright + theme + dashboard redesign | 5+ PR (frontend redesign, dashboard, FilterBar standardization, semantic theme) |

### 5.2 v1.0 sprint 별 워커 분담 권장

| Sprint | 작업 | Claude | Codex | Gemini |
| --- | --- | --- | --- | --- |
| -f | P0-1 sub-carve B | backend handler 제거 + lazy auto-create | — | account.service.ts 폐기 + admin/settings/users 정리 + e2e TC-ACC-* |
| -g | P0-3 Playwright screenshot | — | CI artifact + workflow | screenshot fixture + page 선정 |
| -g/-h | P0-2 UI polish | (review) | — | **주도** — 모든 페이지 semantic theme + responsive |
| -i | P1-1 sub-carve C event listener | 주도 | — | (no-op) |
| -j | P1-2 sub-carve D + P2-1 sub-carve E | sub-carve D 주도 | sub-carve E governance SOP 주도 | — |
| -k | v1.0 종합 검증 | backend test | CI + staging deploy | e2e screenshot review |

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
| `v1.1 Stability` | 2026-07-31 | P1-3, P1-5, P2-2..P2-7, P3-1 (P1-4 off-boarding cron + P3-12 Sign Up 은 2026-05-20 cancelled — 외부 Keycloak / IdP 팀 책임 시나리오 채택) |
| `v2 Extension` | 2026-Q3+ | P3-2..P3-11 (단 P3-13 MFA 는 사내 정책 제외) |

### 6.3 Issue 발급 plan

각 carve (P0/P1/P2/P3) 별로 1 issue 발급. label:
- `priority/p0` `priority/p1` `priority/p2` `priority/p3`
- `worker/claude` `worker/codex` `worker/gemini` `worker/user` (사내 동반 carve)
- `domain/auth` `domain/app-repo-project` `domain/dreq` `domain/integration` `domain/ui-polish` `domain/infra`
- `type/feature` `type/refactor` `type/test` `type/docs` `type/ci`

본 PR 머지 후 별도 sprint (sprint -g-issues 또는 본 sprint -f-roadmap 의 후속 작업) 에서 `gh issue create` 일괄 발급.

## 7. 다음 sprint 진입 순서

본 PR (sprint -f-roadmap) 머지 후 권장 진입 순서:

1. **sprint -g**: P0-3 Playwright screenshot mode 도입 (Codex CI + Gemini frontend fixture) — UI 검증 자산 확보 우선. 작은 변경, 위험 낮음
2. **sprint -h** (or 동시): P0-1 sub-carve B (`/api/v1/accounts/*` 폐기) — Claude backend + Gemini frontend 분담. v1.0 의 마지막 큰 backend 변경
3. **sprint -i**: P0-2 UI polish 1차 (Gemini 주도) + P1-1 sub-carve C (Claude) 동시 — 영역 분리로 충돌 없음
4. **sprint -j**: P1-2 sub-carve D + P2-1 sub-carve E (Claude + Codex)
5. **sprint -k**: v1.0 종합 검증 + staging 1주 monitoring (전 워커)

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
    { path: "/admin/settings/applications", name: "admin-applications" },
    { path: "/admin/topology-v2", name: "admin-topology-v2" },
    { path: "/applications", name: "user-applications" },
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
| 2026-05-20 | 1차 작성 — v1.0 scope 정의 (3 domain) + 잔여 carve 통합 인벤토리 (P0/P1/P2/P3, 30+ item) + 마일스톤 재정의 (M-v1.0 / M-v1.1 / M-v2) + 워커 분업 매트릭스 (Claude=backend / Codex=infra+CI / Gemini=frontend+UX) + GitHub project + milestone plan + UI Playwright screenshot mode | `claude/work_260520-f-roadmap` |
| 2026-05-20 | **P3-12 Sign Up 영구 취소** — 사용자 결정. DevHub 가 Keycloak admin 권한이 없는 외부 IdP 운영 시나리오 (ADR-0020 결정 A 정합). user 생성은 IdP 팀 admin console 또는 HRDB ETL push 책임. §1.2 제외 기능 표 분류 'v1.1 carve' → 'permanently cancelled' + §3.4 P3-12 strikethrough + §5.2 워커 분담 표 strikethrough + §4.2 v1.1 milestone 본문 정정 + issue #235 closed | `claude/work_260520-i-cancel-signup` |
