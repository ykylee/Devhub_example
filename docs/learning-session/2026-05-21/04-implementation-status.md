# Step 4 — 구현 현황

- 문서 목적: DevHub Example 의 **도메인별 backend/frontend/test 활성화 상태**, **운영 자산**, **CI 상태**, **최근 sprint 활동** 을 정리한다.
- 범위: 학습회 §9 (구현 현황 매트릭스) + §10 (운영 자산) + §11 (CI / 최근 활동) source.
- 대상 독자: 학습회 청중 — 무엇이 실 동작하고 무엇이 남았는가.
- 상태: draft (학습회 source)
- 작성일: 2026-05-21
- main HEAD 기준: `d730fc6`.

---

## 1. 한눈에 — 코드/문서 통계 (2026-05-21)

| 항목 | 수치 | 비고 |
| --- | --- | --- |
| **Backend Go 패키지** | 14 | `internal/{audit, auth, commandworker, config, devrequest, domain, gitea, hrdb, httpapi, integrations/adapters, normalize, serviceaction, store}` |
| **DB Migration** | **33 개** | `000001` ~ `000033_user_onboarding_state` |
| **Backend API endpoint** | **80 + 신규 4** (Onboarding Carve A) | `/api/v1/*` routes (router.go v1 group) |
| **Frontend Next.js app route** | 20+ | `app/(dashboard)/admin/settings/*` 9 페이지 + 도메인별 dashboard + onboarding 페이지 (Carve B) |
| **Playwright E2E spec** | **20 spec** | `tests/e2e/*.spec.ts` — auth/account/users-crud/admin-*/dev-requests/topology-v2/integrations 등 |
| **ADR 발급** | **21건** | `0001` ~ `0021` |
| **누적 docs** | ~ 100+ md | `docs/` + `docs/adr/` + `docs/planning/` + `docs/setup/` + `docs/tests/` + `docs/traceability/` |
| **2026-05-21 누적 머지 PR** | **10건** | #265 ~ #270 + #271 + #277 + #278 (codex deploy refactor + Onboarding 전 phase + Carve A backend) |

---

## 2. 도메인별 구현 현황 매트릭스

### 2.1 인증 / 조직 / 대시보드 (Auth + Org + Dashboard)

| 모듈 | Backend | Frontend | E2E | v1.0 |
| --- | --- | --- | --- | --- |
| Keycloak OIDC 토큰 검증 | `internal/auth/keycloak_verifier.go` + JWKS cache + stale-while-error | `auth.service.ts` OIDC discovery + tokenStore | TC-AUTH-* (auth.spec.ts) | ✅ done |
| Keycloak Admin Client (read-only) | `internal/httpapi/keycloak_admin_client.go` (`FindIdentityByUserID` 만) | — | — | ✅ done (ADR-0020 sub-carve E) |
| `users` CRUD + appointments | `organization.go` + `internal/store/postgres_users.go` (4 SELECT) | `identity.service.ts` getUsers/createUser/updateUser | admin-users-crud.spec.ts | ✅ done |
| 조직 단위 (units + hierarchy) | `organization.go` `/api/v1/organization/*` | `app/(dashboard)/admin/settings/organization/page.tsx` | admin-org-crud.spec.ts | ✅ done |
| RBAC policy 편집 | `internal/httpapi/rbac.go` (4 endpoint, sub-carve A 후) | `admin/settings/permissions/page.tsx` + PermissionEditor | admin-permissions.spec.ts | ✅ done |
| Audit log view + Keycloak event listener | `internal/httpapi/audit.go` + `internal/audit/keycloak_event_puller.go` + metric 3종 | `admin/settings/audit/page.tsx` | audit.spec.ts | ✅ done |
| Sign Up (셀프 가입) | `auth_signup.go` (legacy) | `signup/page.tsx` (legacy) | ⚪ | 🟡 **cancelled** (외부 Keycloak, 2026-05-20) |

### 2.2 Application / Repository / Project 도메인 (PMO)

| 모듈 | Backend | Frontend | E2E |
| --- | --- | --- | --- |
| Platform CRUD + rollup | `applications.go` (API-43..50, 57) + store | `admin/settings/platforms/page.tsx` + FilterBar | admin-applications.spec.ts |
| Repository CRUD + activity/PR/build-runs/quality | `repository_ops.go` (API-51..54) | application detail page (activity widgets) | — |
| Project CRUD | `applications.go` project endpoint (API-55..56) | application/project nested page | — |
| Integration (project scope, API-58) | `applications.go` integration endpoint | — | — |
| SCM Provider catalog | API-41/42 | provider 페이지 | — |
| ADR-0011 RBAC row-scoping (assignee = 본인) | `permissions.go` + `enforceRowOwnership` | role gate | covered by AUTH/RBAC |

### 2.3 Dev Request (DREQ) 도메인 — M5 종합

| 모듈 | Backend | Frontend | E2E |
| --- | --- | --- | --- |
| 외부 수신 (POST /dev-requests, API-59) | `internal/httpapi/dev_request_intake_auth.go` (token + IP CIDR) | — | TC-DREQ-INTAKE-AUTH-01 + NEG |
| DREQ CRUD (API-60..65) | `dev_requests.go` + store | `app/(dashboard)/dev-requests/page.tsx` + widget | dev-requests.spec.ts (mega lifecycle) |
| Promote (Platform/Project, API-62) | `dev_requests_promote.go` 단일 트랜잭션 | promote modal | TC-DREQ-PROMOTE-TX-01 |
| Intake Token admin (API-66..68 + 79) | intake token admin handler | `admin/settings/dev-request-tokens/page.tsx` + IssueIntakeTokenModal (plain-1회) | TC-DREQ-ADMIN-TOKEN-* (PATCH-01, REVOKE-01) |
| ADR-0017 hardening | atomicity (`UpdateDevRequestIntakeTokenIPs` FOR UPDATE) + cron revoke + Prometheus metric 3종 | — | — |

### 2.4 External Integration 도메인 — M6 종합

| 모듈 | Backend | Frontend | E2E |
| --- | --- | --- | --- |
| Provider Catalog (API-69..72, 80) | `integration_registry.go` + store | `admin/settings/integrations/page.tsx` + ProviderTable + ProviderModal | admin-integrations.spec.ts (CREATE/EDIT/SYNC/RBAC/DELETE) |
| Provider Webhook (API-73) | provider webhook handler + signature verify | — | — |
| Bindings (API-74/75/81/82) | binding handler + store + 4-tuple unique guard | `admin/settings/integration-bindings/page.tsx` + BindingsTable + CreateBindingModal | admin-integration-bindings.spec.ts |
| HomeLab pull adapter | `internal/integrations/adapters/{homelab_file_puller,homelab_http_puller,homelab_pull_loop}.go` | — | — (TC-INT-HOMELAB-* 3건 active) |
| Infra topology v2 (API-76..78) | `infra_snapshots.go` + topology v2 endpoint | `admin/topology-v2/page.tsx` + React Flow | admin-topology-v2.spec.ts |
| Prometheus metric + alerts | `/metrics` + Counter/Gauge/Histogram + Grafana JSON | Grafana dashboard JSON | — |

### 2.5 Onboarding 도메인 — 2026-05-21 Carve A backend ⚡

| 모듈 | Backend | Frontend | E2E | 상태 |
| --- | --- | --- | --- | --- |
| Migration 000033 (`users.onboarding_completed_at + review_status + CHECK`) | ✅ | — | — | ✅ Carve A PR #278 |
| `onboardingGate` middleware (allowlist + flag conditional) | ✅ | — | — | ✅ Carve A |
| API-83 POST /me/onboarding (제출) | ✅ `me_onboarding.go` (single tx + 201) | ⏳ Carve B | ⏳ Carve D | 🟡 backend done |
| API-84 GET /organizations/search (typeahead) | ✅ `organizations_search.go` (q ≥ 2, limit ≤ 20) | ⏳ Carve B | ⏳ Carve D | 🟡 |
| API-85 PATCH /me (self-service unit change) | ✅ `me.go::patchMe` (review_status auto reset) | ⏳ Carve B (`/account` page) | ⏳ Carve D | 🟡 |
| API-86 POST /admin/users/:id/review (transition) | ✅ `users_admin_review.go` | ⏳ Carve C (admin UI Confirm Review) | ⏳ Carve D | 🟡 |
| API-32 확장 (GET /me 응답 shape) | ✅ `me.go::getMe` (onboarding_required + nullable review_status) | ⏳ Carve B | ⏳ Carve D | 🟡 |
| Feature flag `DEVHUB_ONBOARDING_GATE_ENABLED` | ✅ default OFF (main 안정성) | n/a | n/a | ✅ Carve A |
| UT-onboarding-* (13건) | ✅ `onboarding_test.go` | n/a | n/a | ✅ Carve A |

**다음 단계** = Carve B (frontend, #273) + C (admin UI, #274) + D (E2E, #275) M-v1.1 진입.

### 2.6 인프라 / 운영 / CI

| 모듈 | 활성화 |
| --- | --- |
| 단일 포트 nginx reverse proxy | ✅ `infra/nginx/devhub.deploy.conf` (PR #277) |
| Docker compose 배포 (image only) | ✅ `docker-compose.deploy.yml` + db-migrate auto |
| Deploy preflight + up scripts | ✅ `scripts/deploy-{preflight,up}.sh` (PR #277) |
| Keycloak realm SOP | ✅ `scripts/setup-keycloak.sh` (devhub-frontend + devhub-backend + devhub-e2e-seeder client 3) |
| Prometheus metric + Grafana dashboard | ✅ `docs/setup/grafana/homelab_dashboard.json` |
| Alertmanager guide | ✅ `docs/setup/prometheus_alertmanager_setup.md` |
| HRDB ETL script | 🟡 **deprecated** (외부 Keycloak 시나리오 채택) — `scripts/hrdb_etl_sync.sh` deprecation banner |

---

## 3. CI 상태 (.github/workflows/ci.yml)

```
┌─────────────────────────────────────────────────────────────┐
│  GitHub Actions CI Pipeline                                  │
├─────────────────────────────────────────────────────────────┤
│  1. Detect Changed Paths  (~ 8s)                             │
│     ├── path filter: backend / frontend / docs / .github     │
│     └── paths-skip 으로 docs-only PR 은 9 job 자동 SKIP       │
│                                                               │
│  2. Workflow Lint (actionlint)  (~ 12s)                      │
│     └── ADR-0005 — workflow yml syntax 보장                  │
│                                                               │
│  3. Migration Prefix Uniqueness  (~ 6s)                      │
│     └── migrations/000XXX_ 중복 차단                          │
│                                                               │
│  4. Backend Unit Tests  (~ 20s)                              │
│     └── go test ./... (14 패키지)                             │
│                                                               │
│  5. Backend Integration Tests  (~ 1m)                        │
│     └── PG-backed integration suite (postgres docker)        │
│                                                               │
│  6. Frontend Unit Tests                                      │
│     └── Vitest (skipped on docs-only PR)                     │
│                                                               │
│  7. E2E Tests (Playwright, shard 1/2 + 2/2)  (~ 3-4m each)  │
│     └── seed via global-setup.ts → 51+ tests                 │
└─────────────────────────────────────────────────────────────┘
```

**최근 PR 통과율 (2026-05-21)**:
- PR #277 / #278 — 전 job PASS (E2E 51 passed)
- PR #265 ~ #271, #276 (docs-only) — 5 job paths-skip + 2 job pass

---

## 4. 운영 자산 (v1.0 배포 대상)

### 4.1 환경별 설정 템플릿

| 파일 | 용도 |
| --- | --- |
| `docs/setup/deploy.env.example` | 기본 deploy template |
| `docs/setup/deploy.stage.env.example` | stage 환경 (PR #277) |
| `docs/setup/deploy.prod.env.example` | production 환경 (PR #277) |
| `frontend/.env.example` | frontend dev/test |

**핵심 env** (PR #277 도입):
- `DEVHUB_PUBLIC_BASE_URL` — 외부 진입 host (https://devhub.example.com)
- `DEVHUB_OIDC_ISSUER_URL` vs `DEVHUB_OIDC_JWKS_URL` — issuer/JWKS 분리 시나리오 지원
- `DEVHUB_KEYCLOAK_ADMIN_URL` + admin client secret
- `DEVHUB_ONBOARDING_GATE_ENABLED` (default OFF — Carve A 머지 후 main 안정성)

### 4.2 운영 SOP 문서

| 문서 | 핵심 |
| --- | --- |
| `docs/setup/environment-setup.md` | native default + docker optional |
| `docs/setup/docker-packaging-deployment-guide.md` | image build + compose deploy + nginx |
| `docs/setup/keycloak_operations.md` | realm SOP §8 — JWKS rotation / user disable / event listener / logout chain |
| `docs/setup/prometheus_alertmanager_setup.md` | Alertmanager 통합 + 알람 정책 |
| `docs/setup/homelab_agent_token_rotation.md` | HomeLab agent token rotation |
| `docs/setup/keycloak_service_account_min_role.md` | service account `view-users + view-events` 만 |
| `docs/setup/grafana/homelab_dashboard.json` | Grafana 4-panel dashboard import |

### 4.3 Test catalog

| 도메인 | 카탈로그 위치 |
| --- | --- |
| Auth (M2) | `docs/domain/auth-session/test_cases.md` |
| Command / Infra (M3) | `docs/infrastructure/commandworker/test_cases.md` |
| Integration (M4) | `docs/domain/integration-registry/test_cases.md` |
| DREQ (M5) | `docs/domain/dev-request/test_cases.md` (13 TC) |

---

## 5. 2026-05-21 누적 sprint 활동 (10 머지 PR)

| PR | sprint / 도메인 | 머지 SHA | 핵심 |
| --- | --- | --- | --- |
| #265 | onboarding-concept-2026-05-21 | `e9b7543` | concept §5.9 skip-and-resume + §8 #7 결정 |
| #266 | onboarding-requirements-2026-05-21 | `4d882d5` | REQ-FR/NFR-ONBOARD-* 20개 발급 (§5.7) |
| #267 | onboarding-arch-2026-05-21 | `105b835` | UC-ONBOARD-01..11 + ARCH-ONBOARD-01..06 + API-83..86 |
| #269 | onboarding-adr-2026-05-21 | `a2e751a` | ADR-0021 발급 + ADR-0020 partial supersession (5 위치) |
| #270 | onboarding-codex-hotfix-2026-05-21 | `175bf9a` | codex P1 #16.3 INSERT/UPDATE + P2 §6.1 scope |
| #271 | onboarding-impl-carve-plan-2026-05-21 | `759f101` | IMPL carve plan + RM-ONBOARD-01..04 + ADR-0021 reservation fix (14 위치) |
| #276 | onboarding-codex-hotfix2-2026-05-21 | `703b0f3` | codex P2 ADR-0021 scope 5 위치 정합 + carve plan path |
| #278 | issue-272-onboarding-backend | `4a77d08` | **Carve A backend 실 구현 (PR #278)** — migration 000033 + 5 handler + UT 13 |
| #277 | codex/work_260521-a-next-work | `d730fc6` | Codex deploy refactor (DEVHUB_PUBLIC_BASE_URL + db-migrate + preflight) + claude follow-up (P2) |
| (#268) | (auto-closed, replaced by #269) | — | base branch 삭제로 auto-close — 학습 기록 |

**오늘의 학습 3건** (메모리 저장):
- ADR supersession 의 cross-document scope sync (PR #270 ↔ #276 round)
- Stacked PR + base merge = auto-close (PR #268 → #269 recreate)
- Codex review cycle 의 2 round 패턴 (self-review → codex round 1 → round 2)

---

## 6. 잔여 carve (v1.0 차단 + v1.1 진입)

### 6.1 v1.0 (target 2026-06-15) 차단 1건

- **#214 P1-3** Keycloak group staging-prod 적용 — 사내 운영자 1회 작업 (group 4 생성 + composite role assign).

### 6.2 v1.1 (target 2026-07-31) Onboarding carve 잔여

| Issue | Carve | Worker | 진입 조건 |
| --- | --- | --- | --- |
| #273 | RM-ONBOARD-02 frontend (page + picker + banner + 3-branch gating + /account edit) | Gemini | Carve A 머지 후 ✅ |
| #274 | RM-ONBOARD-03 admin UI (Confirm Review + filter) | Gemini | Carve A 머지 후 ✅ (병행 가능) |
| #275 | RM-ONBOARD-04 tests (UT + E2E mega lifecycle + 6 seed + TC 카탈로그) | Claude (UT) + Gemini (E2E) | Carve A+B+C 모두 머지 후 |

**Feature flag default ON flip** (별도 hotfix PR): Carve D acceptance 통과 + 1주 staging monitoring 후.

### 6.3 v2 P3 carve (10건)

ADR-0015 §6 (3)+(4) worker / push-pull dedup, HA Phase 2, audit event listener SPI push, WebSocket replay, AI Gardener, Gitea Hourly Pull, System Admin dashboard, PermissionCache LISTEN/NOTIFY, 외부 SSO. (cancelled: Sign Up, MFA, off-boarding cron)

---

## 7. 핵심 메시지 (학습회용)

1. **Onboarding 도메인 1차 종합 closing** — 2026-05-21 단일 일자에 8 PR 누적으로 concept → REQ → ARCH/API → ADR → IMPL plan → Carve A backend 완성. M-v1.1 의 P2-8 ✅, P2-9/10/11 잔여.
2. **Feature flag 안전망** — Carve A 머지 후에도 `DEVHUB_ONBOARDING_GATE_ENABLED=false` (default) 로 main 동작 변경 없음. Carve D acceptance + 1주 monitoring 후 별도 hotfix PR 으로 default ON flip.
3. **17 도메인 × 8 chain 단계 ≈ 85% 채움**. M4 (Realtime/AI) + v2 (gRPC AI) 는 의도된 후속.
4. **codex/claude/gemini 3-worker 분업**. 본 학습회 자료의 10 PR 중 9건 claude (Onboarding backend + plan + docs) + 1건 codex (deploy refactor). Carve B/C 는 Gemini 영역.
5. **운영 자산 모두 준비 — v1.0 D-25 차단 1건만**. 사내 Keycloak group staging-prod 1회 작업이 release blocker.

---

## 8. 본 step 의 데이터로 시각화될 차트 (Step 5)

| 차트 | 데이터 source | 의도 |
| --- | --- | --- |
| **도메인별 backend/frontend/test 활성화 표** | §2 매트릭스 6 도메인 × 4 axis | grouped bar — 도메인별 row 채움 |
| **2026-05-21 누적 PR timeline** | §5 표 10건 (시간 순) | vertical timeline — PR 번호 + 도메인 색상 + 머지 SHA |
| **Migration 누적** | 000001 ~ 000033 + 도메인별 분포 | stacked area — 시간 축 + 도메인 색상 |
| **CI job 통과 시간** | §3 의 7 job (Detect/Lint/Migration/Backend Unit/Integration/Frontend/E2E×2) | horizontal bar — 평균 시간 |
| **v1.0 / v1.1 / v2 잔여 carve 분포** | §6 의 v1.0 (1) + v1.1 (12 → 3 onboarding 잔여) + v2 (10) | stacked donut — milestone × priority |

---

## 9. 다음 단계

- [`05-learning-session.html`](./05-learning-session.html) — HTML 최종 자료 (Step 5 — chart.js + slide layout + 접기/펼치기 UI).
- 본 step 의 §2 매트릭스 + §3 CI + §5 PR timeline + §6 잔여 carve 는 Step 5 의 핵심 시각화 source.
