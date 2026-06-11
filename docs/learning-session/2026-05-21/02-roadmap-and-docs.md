# Step 2 — 로드맵 + 개발 문서 요약

- 문서 목적: DevHub Example 의 **마일스톤 진척 (M0~M7)** + **릴리즈 로드맵 (v1.0/v1.1/v2)** + **핵심 개발 문서 12건** 을 한 페이지로 요약한다.
- 범위: 학습회 §4 (마일스톤 타임라인) + §5 (v1.0 DoD + 잔여 carve) + §6 (문서 인덱스) source.
- 대상 독자: 학습회 청중 — 현재 어디까지 됐고 무엇이 남았는지 파악.
- 상태: draft (학습회 source)
- 작성일: 2026-05-21
- main HEAD 기준: `d730fc6`.

---

## 1. 마일스톤 타임라인 (M0 → M7)

```
M0 (~2026-05-07)         M1 (~2026-05-08)         M2 (~2026-05-12)        M3 (~2026-05-13)
─ initial scaffold       ─ M1 PR-A..G            ─ 인증/계정 1차 완성    ─ User/Org 관리
─ Hydra+Kratos 도입       ─ RBAC ADR-0002          ─ login/logout chain     ─ ADR-0008 HRDB
─ ADR-0001               ─ DB-backed matrix      ─ admin/settings/*       ─ ADR-0009 secondary
                                                                            memberships
M4 (planned)             M5 (2026-05-15 ~ 18)    M6 (2026-05-15 ~ 18)    M7 (2026-05-21)
─ Realtime 확장           ─ DREQ 1차 종합           ─ External Integration  ─ Onboarding self-svc
─ WebSocket replay        ─ ADR-0012 intake auth   ─ ADR-0015 HomeLab pull  ─ ADR-0021 Onboarding
─ AI Gardener (v2)       ─ ADR-0013 row-scoping  ─ ADR-0016 alerts policy ─ Carve A (PR #278)
─ PG LISTEN/NOTIFY        ─ ADR-0014 intake admin  ─ provider/binding/      ─ Carve B/C/D 후속
                          ─ ADR-0017 hardening      topology v2 UI          
                          ─ TC-DREQ-* 13건         ─ Alertmanager + Grafana

                                                                          ↓
                                                              v1.0 (target 2026-06-15)
                                                              ─ §1.3 DoD 8건 PASS
                                                              ─ M-v1.0/v1.1/v2 매트릭스
```

---

## 2. 마일스톤별 핵심 요약

### M0 — 초기 스캐폴드 (2026-05-07)
- monorepo 기반 (backend-core + backend-ai + frontend)
- IdP 결정: **ADR-0001** Hydra+Kratos (→ 2026-05-18 PR #167 + ADR-0019 Keycloak 단일화로 supersede)

### M1 — 인증 + RBAC 기반 (2026-05-08)
- M1 PR-A..G 머지 (Bearer token verifier + middleware chain + RBAC enforcement)
- **ADR-0002** DB-backed RBAC matrix (11 resource × 4 action)

### M2 — 인증/계정 1차 완성 (DONE, 2026-05-12)
- login/logout chain + signout redirect + actor enrichment
- `/admin/settings/users|organization|permissions` 9 sub-page 동작
- audit log view + Keycloak event listener (sprint -u~-y, ADR-0020 sub-carve C)

### M3 — User/Org 관리 (대부분 M2 1차 흡수, 2026-05-13)
- **ADR-0008** HRDB PostgreSQL adapter (employee/department 매핑)
- **ADR-0009** secondary memberships + total_count MV
- Sign Up (셀프 가입) — 2026-05-20 **cancelled** (외부 Keycloak 시나리오 채택, issue #235)

### M4 — Realtime/Ops 고도화 (planned, v1.1+)
- WebSocket replay + 리소스 필터링 (P3-6)
- AI Gardener gRPC + Suggestion Feed (P3-7, v2 보조)
- **ADR-0007** RBAC PermissionCache LISTEN/NOTIFY (P3-10, 다중 인스턴스 시)
- 외부 SSO Gitea/AD federation (P3-11, Keycloak identity broker)

### M5 — DREQ 1차 종합 (2026-05-15 ~ 18, 15 sprint + 8 외부 PR)
- 컨셉 → REQ → UC → ARCH → API-59..68 + API-79 → backend 1차 → frontend 1차 → admin UI → E2E
- **ADR-0012** 외부 수신 인증 (API token + IP allowlist 옵션 A)
- **ADR-0013** RBAC row-scoping (`dev_requests:view`/`edit` + assignee = 본인)
- **ADR-0014** intake token admin (plain 1회 노출 + SHA-256 hashed 저장)
- **ADR-0017** 운영 hardening (atomicity + cron revoke + Prometheus metric + PATCH allowed_ips)
- **TC-DREQ-* 13건** + `docs/domain/dev-request/test_cases.md`

### M6 — External Integration 1차 종합 closing (2026-05-15 ~ 18)
- Provider Catalog + 4 endpoint CRUD + Binding CRUD + sync + topology v2
- **ADR-0015** HomeLab pull strategy (file + HTTP puller + scheduler/metrics)
- **ADR-0016** Prometheus alerts policy (push/pull metric + Grafana dashboard JSON)
- API-69..78 + API-80 + API-81/82 (bindings PATCH/DELETE)
- 운영 자산: `docs/setup/prometheus_alertmanager_setup.md` + Grafana JSON

### M7 — Onboarding (Concept/Requirements/Design/ADR closing, 2026-05-21) ⚡
- 컨셉 → REQ §5.7 → UC-ONBOARD-01..11 → ARCH-ONBOARD-01..06 → API-83..86 → **ADR-0021**
- IMPL Carve A backend (PR #278) — migration 000033 + onboardingGate + 5 handler + UT 13
- ADR-0021 = ADR-0020 의 lazy auto-create 결정 partial supersession (5 위치)
- Carve B/C/D 는 M-v1.1 진입 예정

---

## 3. 릴리즈 로드맵 (v1.0/v1.1/v2)

### v1.0 — 1차 릴리즈 (target 2026-06-15)

**8 DoD** (모두 PASS 해야 릴리즈):
1. Keycloak realm OIDC 로그인 (Sign In / Out / token refresh)
2. system_admin `/admin/settings/*` 9 sub-page CRUD 동작
3. Application / Repository / Project CRUD + rollup + 현황 페이지
4. HomeLab integration provider 등록 + sync + binding + topology v2
5. DREQ 흐름 (intake → assignee dashboard → Promote → close)
6. E2E Playwright 전 shard PASS + backend `go test ./...` PASS + frontend `npm run build` PASS
7. 사내 staging 1주 운영 + 외부 사용자 ≥ 5명 로그인 동작
8. UI 디자인 polish 1차 (semantic theme + responsive + a11y baseline)

**잔여 D-25 (2026-06-15) 차단 carve 1건**:
- **#214 P1-3** Keycloak group staging-prod 적용 (사내 운영자 1회 작업)

### v1.1 — 안정성 + 운영 강화 (target 2026-07-31)

| 항목 | priority | worker |
| --- | --- | --- |
| P1-3 group staging-prod | P1 | 사용자 + Codex |
| P1-5 e2e Keycloak admin 전환 | P1 | Gemini + Codex |
| P2-2 ADR-0016 §6 alert 임계 | P2 | Codex |
| P2-3 ADR-0017 §6 (b) PATCH expires_at + UI | P2 | Gemini + Claude |
| P2-4 Bindings UI 강화 | P2 | Gemini |
| P2-5 React Flow group + WebSocket | P2 | Gemini + Claude |
| P2-6 Keycloak SPI provider JAR | P2 | Codex + 사용자 |
| P2-7 HRDB ETL unit pre-stage | P2 | Claude + 사용자 |
| **P2-8 RM-ONBOARD-01 backend** | P2 | Claude (✅ PR #278 머지) |
| **P2-9 RM-ONBOARD-02 frontend** | P2 | Gemini |
| **P2-10 RM-ONBOARD-03 admin UI** | P2 | Gemini |
| **P2-11 RM-ONBOARD-04 tests** | P2 | Claude (UT) + Gemini (E2E) |
| P3-1 sub-carve F `/login` 정리 | P3 | Gemini |

### v2 — 확장 기능 (target 2026-Q3 이후)

| 항목 | 분류 |
| --- | --- |
| P3-2 ADR-0015 §6 (3) dedicated worker | 운영 분리 |
| P3-3 ADR-0015 §6 (4) push/pull dedup | 별도 ADR |
| P3-4 HA Phase 2 (별도 ADR 후보) | 인프라 |
| P3-5 audit event listener SPI push 전환 | latency |
| P3-6 RM-M4-01..03 WebSocket 확장 | 실시간 |
| P3-7 RM-M4-04..05 AI Gardener | AI 보조 |
| P3-8 RM-M4-06 Gitea Hourly Pull | 외부 SCM |
| P3-9 RM-M4-07 System Admin 대시보드 | 운영 |
| P3-10 RM-M4-08 PermissionCache LISTEN/NOTIFY | RBAC |
| P3-11 RM-M4-09 외부 SSO | infra |

**cancelled**:
- ~~P1-4 off-boarding cron deploy~~ (2026-05-20, 외부 Keycloak 시나리오)
- ~~P3-12 Sign Up 셀프 가입~~ (2026-05-20, IdP 팀 책임)
- ~~P3-13 MFA / 2FA~~ (사내 정책 — Keycloak Account Console 위임)

---

## 4. 핵심 개발 문서 인덱스 (12건)

| 문서 | 위치 | 핵심 |
| --- | --- | --- |
| **요구사항 정의서** | [`docs/requirements.md`](../../requirements.md) | §2 역할별 + §5.4~§5.7 4 도메인 (App/Proj/DREQ/INT/Onboarding) — REQ-FR/NFR-* 약 200 row |
| **시스템 아키텍처** | [`docs/architecture.md`](../../architecture.md) | §6 RBAC + §7 DREQ + §8 Integration + §9 Onboarding — ARCH-* + 컴포넌트 diagram |
| **Backend API 계약** | [`docs/backend_api_contract.md`](../../backend_api_contract.md) | API-01..86 — 16 section + endpoint별 request/response/error matrix |
| **개발 로드맵** | [`docs/development_roadmap.md`](../../development_roadmap.md) | M0~M7 historical + 트랙 매핑 + 충돌 해소 표 |
| **v1.0 릴리즈 로드맵** | [`docs/planning/release_v1_roadmap.md`](../../planning/release_v1_roadmap.md) | v1.0 DoD 8건 + 도메인 모듈 매트릭스 3 + 잔여 carve P0~P3 11건 + 워커 분업 |
| **System Usecase** | [`docs/planning/system_usecases.md`](../../planning/system_usecases.md) | §2 도메인별 UC-* (Auth/Account/Org/RBAC/Gitea/CMD/Audit/RT/App/Proj/DREQ/INT/Onboarding) |
| **추적성 매트릭스** | [`docs/traceability/report.md`](../../traceability/report.md) | §3 종합 매트릭스 17 도메인 row + §4 ADR 인덱스 + §6 changelog |
| **거버넌스 표준** | [`docs/governance/document-standards.md`](../../governance/document-standards.md) | 메타 헤더 / lifecycle / ID 노출 / branch 명명 규칙 |
| **워커 분담** | [`docs/governance/worker_division.md`](../../governance/worker_division.md) | Claude (backend+design) / Gemini (frontend+UX) / Codex (infra+CI) / User (사내) |
| **기술 스택** | [`docs/shared/tech_stack.md`](../../shared/tech_stack.md) | Go + Next.js + Postgres + Keycloak + Prometheus |
| **환경 구성** | [`docs/setup/environment-setup.md`](../../setup/environment-setup.md) | native default + docker optional |
| **테스트 카탈로그** | [`docs/tests/`](../../tests/) | TC-AUTH-*/USR-*/DREQ-*/INT-* — domain별 mega lifecycle spec |

### Concept/Design Phase 1 문서 (planning/)

- `docs/domain/onboarding/concept.md` — Onboarding 컨셉 (§5.1~§5.9 + §8 #1~12 결정)
- `docs/domain/onboarding/impl_plan.md` — IMPL Carve A~D plan + RM-ONBOARD-01..04
- `docs/domain/dev-request/concept.md` — DREQ 컨셉
- `docs/domain/integration-registry/external_system_concept.md` — Integration 컨셉
- `docs/domain/auth-session/account_redesign.md` — ADR-0020 Phase 1/2 매트릭스 + 명시 결정 6건
- `docs/infrastructure/keycloak-idp/refactor_execution_plan.md` — ADR-0019 실행 계획 (PR #167 ~ KC-PR-A..F)
- `docs/infrastructure/keycloak-idp/failover.md` — Keycloak HA design (Phase 1 graceful / Phase 2 active-active)
- `docs/infrastructure/keycloak-idp/offboarding_immediacy.md` — off-boarding 즉시성 design
- `docs/domain/rbac-permissions/keycloak_groups_mapping.md` — Keycloak groups → DevHub role 매핑

### ADR 인덱스 (21건)

| Phase | ADR | Key |
| --- | --- | --- |
| 정책 / 인프라 | 0003 No-Docker, 0005 actionlint, 0018 Single port | 운영 baseline |
| 인증 / 계정 | 0001 ~ Hydra+Kratos~, 0019 Keycloak 단일화, 0020 책임 경계, 0021 Onboarding | IdP 결정 chain |
| RBAC | 0002 policy edit API, 0007 multi-instance cache, 0011 row-scoping | 권한 modeling |
| 도메인 모델 | 0008 HRDB, 0009 secondary memberships, 0010 primary_dept | 조직 model |
| DREQ | 0012 intake auth, 0013 row-scoping, 0014 admin token, 0017 hardening | DREQ chain |
| Integration | 0015 HomeLab pull, 0016 Prometheus alerts | Integration chain |
| Legacy / cleanup | 0004 ~X-Devhub-Actor~, 0006 inbound reject | Legacy removal |

---

## 5. 워커 분업 (간단)

| 워커 | 영역 | 본 학습회 자료 기준 분담 |
| --- | --- | --- |
| **Claude** | Backend (Go) + ADR / docs design + UT | Onboarding Carve A/D + Carve plan + ADR 발급 + traceability sync |
| **Gemini** | Frontend (Next.js) + UX / a11y + E2E | Onboarding Carve B/C + UI polish + e2e seed |
| **Codex** | Infra (Docker/Nginx/CI) + Security + deploy | PR #277 deploy refactor + GHCR + Keycloak realm SOP |
| **User (사내)** | 사내 인프라 / 운영자 1회 작업 | Keycloak realm 생성 + group staging-prod + SPI JAR 배포 |

상세 — [`docs/governance/worker_division.md`](../../governance/worker_division.md) + [`docs/planning/release_v1_roadmap.md §5`](../../planning/release_v1_roadmap.md).

---

## 6. 본 step 의 데이터로 시각화될 차트 후보 (Step 5)

| 차트 | 데이터 source | 의도 |
| --- | --- | --- |
| **M0~M7 마일스톤 timeline** | §2 본문의 마일스톤별 일자 | horizontal bar — 기간 + 핵심 결과 hover |
| **v1.0/1.1/2 carve 분포** | §3 의 P0~P3 11건 × milestone | stacked bar — 우선순위 색상 + worker 비율 |
| **도메인별 ADR 분포** | §4 ADR 인덱스 21건 × 도메인 7 | pie / donut — 도메인별 ADR 수 |
| **PR 누적 timeline** | 본 학습회 작성 시점까지 onboarding 도메인 9 PR | line — 일별 PR + 머지 시점 |

---

## 7. 다음 단계

- [`03-traceability-summary.md`](./03-traceability-summary.md) — 추적성 요약 (Step 3, REQ → UC → ARCH → API → IMPL → UT → TC 7단계 chain + 도메인별 row 채움 현황).
- 본 step 의 §1 timeline + §3 carve 표 + §4 ADR 분포 + §5 워커 분업은 Step 5 의 HTML 자료에서 chart.js (Gantt / stacked bar / pie) 로 시각화.
