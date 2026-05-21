# Step 3 — 추적성 요약

- 문서 목적: DevHub Example 의 **7단계 추적성 chain** (REQ → UC → ARCH → API → IMPL → UT → TC) 과 도메인별 채움 현황을 한 페이지로 요약한다.
- 범위: 학습회 §7 (추적성 chain 개념) + §8 (도메인별 매트릭스 시각화) source.
- 대상 독자: 학습회 청중 — 요구사항 ID 부터 e2e test 까지 어떻게 연결되는지.
- 상태: draft (학습회 source)
- 작성일: 2026-05-21
- main HEAD 기준: `d730fc6`. 1차 source = [`docs/traceability/report.md`](../../traceability/report.md).

---

## 1. 7단계 추적성 chain (개념)

```
┌──────────────┐  →  ┌──────────────┐  →  ┌──────────────┐  →  ┌──────────────┐
│  REQ-FR-*    │     │  UC-*        │     │  ARCH-*      │     │  API-*       │
│  REQ-NFR-*   │     │  (Usecase)   │     │  (Component) │     │  (Endpoint)  │
│  requirements│     │  system_     │     │  architecture│     │  backend_api │
│              │     │  usecases    │     │              │     │  _contract   │
└──────────────┘     └──────────────┘     └──────────────┘     └──────┬───────┘
                                                                       │
                ┌──────────────────────────────────────────────────────┘
                ▼
┌──────────────┐  →  ┌──────────────┐  →  ┌──────────────┐
│  IMPL-*      │     │  UT-*        │     │  TC-*        │
│  (Code path) │     │  (Unit test) │     │  (E2E spec)  │
│  router/     │     │  go test     │     │  Playwright  │
│  handler/    │     │              │     │  spec.ts     │
│  store       │     │              │     │              │
└──────────────┘     └──────────────┘     └──────────────┘
                          │                       │
                          └──────┬────────────────┘
                                 ▼
                       ┌─────────────────────┐
                       │  CI gate (PR 머지)   │
                       │  + Audit (운영)      │
                       └─────────────────────┘
```

**원칙**:
- 부여된 ID 는 **재사용/재번호 안 함** (immutable). 항목 삭제 시 "deprecated" 마킹만 — chain 깨짐 방지.
- 모든 PR 이 영향 받는 단계의 ID 를 발급 + `docs/traceability/report.md` 의 인덱스 + §3 매트릭스 row 동기화.
- 신규 도메인은 `REQ → UC → ARCH/API` 순서를 기본 체인으로 — 기존 도메인은 점진 전환.

---

## 2. ID prefix + 영역 (총 7 + 1 ADR)

| 단계 | 접두사 | 형식 | 예시 | 위치 |
| --- | --- | --- | --- | --- |
| 요구사항 (functional) | `REQ-FR-` | `REQ-FR-{nn}` 또는 `REQ-FR-{DOMAIN}-{nnn}` | `REQ-FR-01`, `REQ-FR-ONBOARD-001` | `docs/requirements.md` |
| 요구사항 (non-functional) | `REQ-NFR-` | 동일 | `REQ-NFR-03`, `REQ-NFR-ONBOARD-001` | 동일 |
| Usecase | `UC-` | `UC-{DOMAIN}-{nn}` | `UC-APP-01`, `UC-ONBOARD-01` | `docs/planning/system_usecases.md` |
| Architecture | `ARCH-` | `ARCH-{nn}` 또는 `ARCH-{DOMAIN}-{nn}` | `ARCH-02`, `ARCH-DREQ-01`, `ARCH-ONBOARD-01` | `docs/architecture.md` |
| API contract | `API-` | `API-{nn}` | `API-07`, `API-86` | `docs/backend_api_contract.md` |
| Roadmap | `RM-` | `RM-M{0..3}-{nn}` 또는 `RM-{DOMAIN}-{nn}` | `RM-M2-04`, `RM-ONBOARD-01` | `docs/development_roadmap.md` |
| Implementation | `IMPL-` | `IMPL-{module}-{nn}` | `IMPL-auth-01`, `IMPL-onboarding-gate-01` | `backend-core/internal/...` |
| Unit test | `UT-` | `UT-{pkg}-{nn}` | `UT-httpapi-05`, `UT-onboarding-submit-01` | `*_test.go` 파일 |
| E2E | `TC-` | `TC-{FEATURE}-{nn}` | `TC-AUTH-01`, `TC-DREQ-INTAKE-AUTH-01` | `frontend/tests/e2e/*.spec.ts` |
| **ADR** (별개 식별) | `ADR-` | `ADR-{nnnn}` | `ADR-0019`, `ADR-0021` | `docs/adr/` |

---

## 3. 단계별 ID 발급 수 (2026-05-21 기준)

| 단계 | 카탈로그 ID 수 | 분포 |
| --- | --- | --- |
| **REQ-FR-*** | **~ 150 row** | 일반 (REQ-FR-01..105) + APP/PROJ (12+11+6) + DREQ (11) + INT (12) + Onboarding (12) |
| **REQ-NFR-*** | **~ 55 row** | 일반 (REQ-NFR-01..26) + PROJ (6) + DREQ (6) + INT (8) + Onboarding (8) |
| **UC-*** | **~ 75 항목** | AUTH 3 / ACCOUNT 3 / ORG 4 / RBAC 3 / GITEA 3 / CMD 3 / AUD 2 / RT 2 / APP 10 / PROJ 10 / DREQ 10 / INT 14 / ONBOARD 11 |
| **ARCH-*** | **35 항목** | 일반 (ARCH-01..17) + DREQ (6) + INT (6) + ONBOARD (6) |
| **API-*** | **86 항목** | API-01..86 (composite/결손 정책 일부 유지). 활성화 endpoint 80여 개 + DREQ/INT/Onboarding 신규 28 |
| **RM-*** | **~ 25 항목** | M-M0..M3 + M4 planned (9) + DREQ (carve out) + INT (deferred) + ONBOARD (4: ONBOARD-01..04) |
| **IMPL-*** | **~ 80 항목** | auth (7) / rbac (4) / account/org (8) / command (5) / audit (2) / dreq (10+8 frontend) / int (8+5 frontend) / onboarding (8, Carve A) 등 |
| **UT-*** | **~ 50 항목** | httpapi (40+) / store / domain / frontend Vitest |
| **TC-*** | **~ 50 항목** | AUTH (NEG/NOAUTH/SIGNOUT) + USR-CRUD (3) + ORG (LIST/UNIT/MEM/CHART) + ACC (PROFILE) + DREQ (13) + INT (12+1 frontend negative) + ONBOARD (11 후보) |
| **ADR** | **21건** | 0001~0021 (0001 superseded by 0019; 0004 retired; 0006 reject inbound; 0020 partial superseded by 0021) |

---

## 4. 도메인별 chain 채움률 (현 시점)

매트릭스 — 각 단계가 채워졌으면 ✅ / partial 이면 🟡 / 미진입은 ⚪.

| 도메인 | REQ | UC | ARCH | API | RM | IMPL | UT | TC |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| **인증 (Auth/OIDC)** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **회원가입 (signup)** | ✅ | ✅ | ✅ | 🟡 (legacy 폐기) | 🟡 (cancelled) | 🟡 | ⚪ | ✅ |
| **RBAC** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **계정 관리 (admin + self)** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **조직 계층 (org units)** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **감사 (audit)** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **명령 lifecycle** | ✅ | ✅ | ⚪ (planned) | ✅ | 🟡 (M4) | ✅ | ✅ | ⚪ (gap §5.1) |
| **실시간 (WebSocket)** | ✅ | ✅ | ✅ | ✅ | 🟡 (M4) | ✅ | ✅ | ⚪ (gap §5.1) |
| **인프라 토폴로지** | ✅ | ✅ | ✅ | ✅ | 🟡 (M4) | ✅ | ✅ | 🟡 (정적만, gap §5.1) |
| **Webhook (gitea)** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ⚪ (UT 로 충분) |
| **대시보드 / me** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **CI / 거버넌스** | ✅ (NFR-1) | ✅ | ✅ (ADR-0003) | ⚪ | ✅ | ✅ | ⚪ | ⚪ (CI run 자체가 검증) |
| **Application/Project (PMO)** | ✅ | ✅ | ✅ | ✅ | 🟡 (M3) | ✅ | ✅ | 🟡 (FilterBar TC 부분) |
| **Dev Request (DREQ)** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ (13 TC 정식) |
| **External Integration** | ✅ | ✅ | ✅ | ✅ | 🟡 (deferred) | ✅ | ✅ | ✅ (12 TC frontend + UT only neg) |
| **Onboarding ⚡** | ✅ | ✅ | ✅ | ✅ | ✅ | 🟡 (Carve A) | 🟡 (Carve A UT 13) | ⚪ (Carve D 후속) |
| **M4 (planned)** | ✅ | ✅ | 🟡 (확장 예정) | 🟡 (WS) | ✅ | ⚪ | ⚪ | ⚪ |

**범례**:
- ✅ = 발급 완료 + 활성화
- 🟡 = partial (1차 발급 / 일부 미진입 / 후속 carve)
- ⚪ = 미진입 (M-v1.1/v2 후속 또는 의도된 gap)

---

## 5. 도메인 chain 예시 (DREQ + Onboarding)

### 5.1 DREQ 도메인 — 완전 chain 예시

```
REQ-FR-DREQ-001~011 (외부 시스템 의뢰 수신 + 흐름)
        │
        ▼
UC-DREQ-01~10 (의뢰 수신/조회/Promote/Reject/Reassign/Close/...)
        │
        ▼
ARCH-DREQ-01~06 (컴포넌트 / 상태머신 / 외부 수신 인증 / RBAC / 데이터 모델 / Audit)
        │
        ▼
API-59~68 (DREQ §14) + API-79 (PATCH allowed_ips & expires_at)
        │
        ▼
ADR-0012 (intake auth) + ADR-0013 (row-scoping) + ADR-0014 (admin token) + ADR-0017 (hardening)
        │
        ▼
IMPL-dreq-{domain,store,handler,router,rbac,intake_auth,promote_tx,intake_admin,token_expiry,allowed_ips_mutation}-01
+ IMPL-dreq-frontend-{page,widget,table,modal,service,intake_admin,...}-01
+ migration 000022 (dev_requests) + 000023 (dev_request_intake_tokens) + 000024 (RBAC seed) + 000026/000027
        │
        ▼
UT-dreq-handler-XX (15 + 5 promote + 8 intake admin = 28 test) + UT-dreq-intake_auth + UT-clientIP
+ UT-dreq-token-expiry + UT-dreq-allowed-ips-mutation + UT-frontend-IntakeTokenTable
        │
        ▼
TC-DREQ-* 13건 (test_cases_m5_dreq.md) — mega lifecycle + RBAC negative + intake auth negative
```

### 5.2 Onboarding 도메인 — 2026-05-21 완성된 phase 1~4 chain

```
REQ-FR-ONBOARD-001..012 + REQ-NFR-ONBOARD-001..008 (PR #266)
        │                                  ← concept (PR #260 + #265, §5.9 skip-and-resume)
        ▼
UC-ONBOARD-01..11 (PR #267, system_usecases.md §2.13)
        │
        ▼
ARCH-ONBOARD-01..06 (PR #267, architecture.md §9 — 컴포넌트 / 3-tier 상태머신 / Gating / RBAC route / 데이터 모델 / Audit)
        │
        ▼
API-83..86 신규 + API-32/API-33 확장 (PR #267, backend_api_contract.md §16)
        │
        ▼
ADR-0021 (PR #269 — Onboarding self-service + lazy auto-create supersession)
        │     ↳ ADR-0020 partial supersession (5 위치 inline banner)
        ▼
IMPL Carve plan (PR #271, onboarding_impl_plan.md) — RM-ONBOARD-01..04 발급
        │
        ▼
RM-ONBOARD-01 Carve A backend (PR #278, ⚡ 2026-05-21 머지)
        ├── migration 000033 (users.onboarding_completed_at + review_status + bi-implication CHECK)
        ├── domain model 확장 (AppUser + UpdateUserInput + OnboardingSubmitInput + ReviewStatus enum)
        ├── store (4 SELECT 갱신 + SubmitOnboarding + ConfirmUserReview + SearchOrgUnits + UpdateUser ReviewStatus)
        ├── authenticateActor 의 feature flag conditional (false=lazy 유지 / true=token-only actor)
        ├── onboardingGate middleware (allowlist 4 endpoint + flag conditional)
        ├── 5 handler (POST /me/onboarding API-83 / GET /organizations/search API-84 /
        │              PATCH /me API-85 / POST /admin/users/:id/review API-86 / GET /me API-32 확장)
        ├── handler-level flag guard (P1 fix)
        ├── permissions.go (4 entry) + main.go env wire (DEVHUB_ONBOARDING_GATE_ENABLED default OFF)
        └── UT-onboarding-* 13건 (gate / submit happy-preseeded-409-404-422 / search / review confirm / patch me)

Carve B (frontend, #273) + C (admin UI, #274) + D (E2E, #275) → M-v1.1 후속
```

---

## 6. ADR 인덱스 (영향 도메인 × 핵심 결정)

| ADR | 도메인 | 결정 | 상태 |
| --- | --- | --- | --- |
| 0001 | 인증 / 계정 | ~Hydra+Kratos~ → ADR-0019 supersession | superseded |
| 0002 | RBAC | DB-backed 11 × 4 matrix | accepted |
| 0003 | CI / 운영 | No-Docker CI scope | accepted |
| 0004 / 0006 | Legacy | X-Devhub-Actor removal / inbound reject | accepted |
| 0005 | CI | workflow lint (actionlint) | accepted |
| 0007 | RBAC | PermissionCache multi-instance (PG LISTEN/NOTIFY) | accepted (impl carve out) |
| 0008 | 회원가입 | HRDB production adapter (PostgreSQL schema) | accepted (Phase 2 cancelled) |
| 0009 / 0010 | 조직 | secondary memberships / primary_dept resolution | accepted |
| 0011 | RBAC | row-scoping (assignee = 본인) | accepted |
| 0012 | DREQ | external intake auth (옵션 A — API token + IP allowlist) | accepted |
| 0013 | DREQ | RBAC row-scoping for dev_requests | accepted |
| 0014 | DREQ | intake token admin (plain 1회 노출 + SHA-256 hashed) | accepted |
| 0015 | Integration | HomeLab pull strategy (file + HTTP) | accepted |
| 0016 | Integration | Prometheus alerts policy | accepted |
| 0017 | DREQ | intake token operational hardening (atomicity + cron + metric) | accepted |
| 0018 | 인프라 | single port reverse proxy (`/devhub` prefix) | accepted |
| 0019 | 인증 | Keycloak 단일화 (Hydra+Kratos 폐기, ADR-0001 supersede) | accepted |
| 0020 | 인증 / 계정 | account/user 책임 경계 (Keycloak + DevHub `users`) | partial superseded by 0021 |
| 0021 | Onboarding | self-service unit selection + lazy auto-create supersession | accepted (2026-05-21) ⚡ |

---

## 7. Gap 요약 (현 시점, 의도된 미진입)

| 도메인 / 항목 | Gap | 우선순위 | 보완 plan |
| --- | --- | --- | --- |
| **명령 lifecycle E2E** | TC 카탈로그 closed, spec ts open | P2 | M3+ sprint 에서 spec 작성 |
| **인프라 토폴로지 E2E (interactive)** | 정적 render TC 만, click/group toggle open | P2 | M4 진입 시 |
| **실시간 (WebSocket) E2E** | M4 planned, current = `command.status.updated` 만 publish | P3 (M4 의존) | RM-M4-01..03 |
| **Onboarding E2E (TC-ONBOARD-*)** | Carve A 머지 후, Carve D (frontend E2E mega lifecycle) 후속 | P2 (M-v1.1) | Carve B/C 머지 후 Carve D 진입 |
| **Backend AI (gRPC client)** | IMPL-ai-XX 미발급 (v2 placeholder) | P3 (v2) | v2 AI Gardener carve |
| **카탈로그된 TC 의 spec 역검증** | spec 파일 grep 자동 hygiene 부재 | P3 | CI lint hook 후보 |

상세 — [`docs/traceability/report.md §5`](../../traceability/report.md).

---

## 8. 본 step 의 데이터로 시각화될 차트 (Step 5)

| 차트 | 데이터 source | 의도 |
| --- | --- | --- |
| **단계별 ID 발급 수** | §3 표 | horizontal bar — 7단계 + ADR. 도메인 색상 구분 |
| **도메인별 chain 채움률** | §4 매트릭스 17 도메인 × 8 단계 | heatmap — ✅/🟡/⚪ 색상 |
| **ADR 도메인 분포** | §6 ADR 인덱스 21건 × 8 도메인 | donut — 도메인별 ADR 수 |
| **Onboarding chain timeline** | §5.2 의 PR 시퀀스 (PR #260 ~ #278) | vertical timeline — 일별 PR + chain 단계 표시 |

---

## 9. 핵심 메시지 (학습회용)

1. **추적성 = 자동화된 정합성**. PR 마다 7단계 ID 발급 + matrix 갱신 강제 — 어느 단계에서 시작하든 chain 전체 추적 가능.
2. **ID 영구 (immutable)**. 삭제 시에도 deprecated 마킹만 — 9개월 후 신규 합류자가 PR #260 → ADR-0021 → migration 000033 → handler 까지 따라갈 수 있음.
3. **신규 도메인 추가 = chain 전체 + ADR**. Onboarding 도메인 (2026-05-21) = REQ 20 + UC 11 + ARCH 6 + API 4 + ADR 1 + Carve plan + Carve A backend = **7 PR / 1 week sprint**.
4. **Gap 은 의도된 미진입**. 17 도메인 × 8 단계 = 136 cell 중 ~85% 채워짐. M4 (Realtime/AI) + v2 (gRPC AI) 는 의도된 후속.

---

## 10. 다음 단계

- [`04-implementation-status.md`](./04-implementation-status.md) — 구현 현황 (Step 4 — 도메인별 backend/frontend/test 활성화 + 운영 자산 + CI 상태).
- 본 step 의 §3 ID 발급 수 + §4 chain 채움률 매트릭스 + §6 ADR 인덱스는 Step 5 HTML 자료의 핵심 차트 source.
