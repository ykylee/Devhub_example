# Step 1 — 프로젝트 개요 (DevHub Example)

- 문서 목적: 학습회 발표를 위해 DevHub Example 프로젝트의 **목적·범위·아키텍처 핵심·이해관계자·핵심 결정**을 한 페이지로 정리한다.
- 범위: 전체 프로젝트 개요 (1차 source). 본 step 은 HTML 학습회 자료의 §1 §2 §3 source.
- 대상 독자: 학습회 청중 (사내 개발자/PMO/시스템 운영자), 처음 프로젝트를 접하는 신규 합류자.
- 상태: draft (학습회 source)
- 작성일: 2026-05-21
- main HEAD 기준: `d730fc6` (PR #277 codex deploy refactor 머지 후).

---

## 1. 한 줄 정의

> **DevHub Example** = 역할별 진입 페이지 우선순위로 UX 를 간접 제공하는 **통합 개발 허브**. 개발자/관리자/시스템 관리자가 동일 시스템에서 자기 역할에 최적화된 대시보드로 진입한다.

---

## 2. 핵심 가치 제안 (Value Proposition)

| 가치 | 어떻게 |
| --- | --- |
| **역할별 정보 탐색 시간 최소화** | Developer / Manager / System Admin 의 기본 진입 페이지가 각자 다른 대시보드. 같은 권한 모델 위에 UX 만 분리. |
| **단일 진실 source (SoT)** | `users` 단일 테이블 + Keycloak OIDC 단일 IdP. 자체 `accounts` 테이블 없음 — IdP 가 credential, DevHub 가 organizational metadata 만. |
| **외부 시스템 통합 single pane** | DREQ (외부 의뢰 intake) + External Integration (HomeLab/SCM/CI-CD/문서) 모두 동일 dashboard 에서 surface. |
| **자동화된 추적성** | REQ → UC → ARCH → API → IMPL → UT → TC 7단계 매트릭스. PR 마다 ID 발급 + 매트릭스 cell 동기화 강제. |

---

## 3. 시스템 컴포넌트 구조 (high-level)

```
┌────────────────────────────────────────────────────────────────┐
│                      User (Browser)                            │
└──────────────────────────────┬─────────────────────────────────┘
                               │ HTTPS (single port via nginx)
                               ▼
              ┌──────────────────────────────┐
              │   nginx (reverse proxy)      │
              │   /devhub/*   → frontend     │
              │   /api/v1/*   → backend-core │
              │   /devhub/auth/keycloak/* → KC│
              └────┬─────────┬──────────┬────┘
                   │         │          │
                   ▼         ▼          ▼
       ┌───────────────┐ ┌──────────────────┐ ┌──────────┐
       │ Next.js       │ │ Go Core          │ │ Keycloak │
       │ (frontend)    │ │ (backend-core)   │ │ (IdP)    │
       │  - Dashboard  │ │  - 5 도메인 API   │ │  - OIDC  │
       │  - RBAC menu  │ │  - RBAC matrix    │ │  - users │
       │  - Realtime WS│ │  - Audit/Realtime │ │          │
       └───────┬───────┘ └────────┬─────────┘ └────┬─────┘
               │                  │                │
               │                  ▼                │
               │       ┌──────────────────┐        │
               │       │ Postgres         │        │
               │       │  - users (org)   │        │
               │       │  - 5 도메인 model│        │
               │       │  - audit_logs    │◀───────┘
               │       │  - migration ~33 │ (event puller)
               │       └──────────────────┘
               │
               ▼
       ┌────────────────────┐
       │ Python AI          │  (gRPC client)
       │ (backend-ai, v2)   │
       └────────────────────┘
```

**핵심**:
- 외부 진입은 **단일 포트** (nginx + `/devhub` prefix, [ADR-0018](../../adr/0018-single-port-reverse-proxy-policy.md)).
- 인증은 **Keycloak 단일 IdP** ([ADR-0019](../../adr/0019-keycloak-only-idp.md)). 자체 OAuth proxy 없음.
- backend-core 의 `internal/httpapi` 가 5 도메인 API 핸들러 통합 (Auth / RBAC / DREQ / Integration / Onboarding + 기본 Domain).

---

## 4. 도메인 인벤토리 (5 + 기본)

| 도메인 | 핵심 ADR | 핵심 책임 | 상태 (2026-05-21) |
| --- | --- | --- | --- |
| **인증 / 계정 / 사용자** | [ADR-0019](../../adr/0019-keycloak-only-idp.md), [ADR-0020](../../adr/0020-account-user-management-boundary.md) | Keycloak OIDC + DevHub `users` (organizational metadata) + RBAC | ✅ M2 done, M-v1.1 정합 진행 |
| **RBAC (권한 매트릭스)** | [ADR-0002](../../adr/0002-rbac-policy-edit-api.md), [ADR-0011](../../adr/0011-rbac-row-scoping.md), [ADR-0007](../../adr/0007-rbac-cache-multi-instance.md) | 11 resource × 4 action matrix + row-level scoping | ✅ accepted + 1차 구현 |
| **Application / Project (PMO)** | [ADR-0011](../../adr/0011-rbac-row-scoping.md) | 외부 의뢰의 promote-to-application/project + team_manager role 위양 | ✅ M3 1차 backend 완성 |
| **Dev Request (DREQ)** | [ADR-0012](../../adr/0012-dreq-external-intake-auth.md), [ADR-0013](../../adr/0013-dreq-rbac-row-scoping.md), [ADR-0014](../../adr/0014-dreq-intake-token-admin.md), [ADR-0017](../../adr/0017-dreq-intake-token-operational-hardening.md) | 외부 시스템 → DevHub intake → assignee → Platform/Project promote | ✅ M5 1차 + admin UI + E2E |
| **External Integration** | [ADR-0015](../../adr/0015-homelab-adapter-pull-strategy.md), [ADR-0016](../../adr/0016-prometheus-alerts-policy.md) | ALM/SCM/CI-CD/문서/HomeLab provider 통합 + bindings/topology UI | ✅ M6 1차 + frontend |
| **Onboarding (신규)** | [ADR-0021](../../adr/0021-onboarding-self-service-unit-selection.md) | Keycloak 인증 통과 + 프로필 미완료 사용자의 self-service 초기 등록 | ⚡ **2026-05-21 Carve A backend done** (M-v1.1) |

---

## 5. 이해관계자 (Stakeholders)

| 역할 | 핵심 활동 | 진입 dashboard |
| --- | --- | --- |
| **Developer** | repo/PR/CI/risk 조회, DREQ assignee 대응, profile 등록 | `/developer` |
| **Manager** | 팀 부하 / 일정 지연 / DREQ oversight | `/manager` |
| **PMO Manager** | Platform/Project 라이프사이클 관리 (team_manager role) | `/manager` (확장) |
| **System Admin** | Gitea runner / RBAC policy / DREQ token admin / event listener | `/system-admin` |
| **외부 시스템** (Bitbucket/Bamboo/Jira/HomeLab) | API token + IP allowlist 로 DevHub intake 호출 | (백그라운드) |
| **사용자 일반** | onboarding 화면 → 소속 등록 → reviewed | `/devhub/onboarding` |

---

## 6. 핵심 ADR 인덱스 (영향도 순)

| ADR | 결정 | 영향 도메인 |
| --- | --- | --- |
| [0001](../../adr/0001-idp-selection.md) ~~Hydra+Kratos~~ → [0019](../../adr/0019-keycloak-only-idp.md) Keycloak 단일화 | IdP 결정 reversal — Keycloak single IdP + `users.idp_subject` 일반화 | 인증 (전 시스템) |
| [0002](../../adr/0002-rbac-policy-edit-api.md) RBAC policy edit API | DB-backed 11 resource × 4 action matrix | RBAC (전 시스템) |
| [0011](../../adr/0011-rbac-row-scoping.md) RBAC row-scoping | application/project/dev_request 의 row-level 권한 (assignee = 본인) | App/Proj/DREQ |
| [0012](../../adr/0012-dreq-external-intake-auth.md) DREQ intake auth | 외부 수신 = API token + IP allowlist (옵션 A) | DREQ |
| [0018](../../adr/0018-single-port-reverse-proxy-policy.md) Single port reverse proxy | nginx + `/devhub` prefix + 단일 포트 강제 | 인프라 |
| [0020](../../adr/0020-account-user-management-boundary.md) Account/User 책임 경계 | Keycloak = credential / DevHub = organizational metadata | 인증, 계정 |
| [0021](../../adr/0021-onboarding-self-service-unit-selection.md) Onboarding self-service | lazy auto-create 폐기 + onboarding 완료 시점에 row 생성 + 3 tier 접근 | 인증, Onboarding |

---

## 7. 운영 정책 핵심 5건

1. **No-Docker 정책 (default)** — 로컬 개발은 native (Go/Python/Node 직접 실행). Docker 자산은 환경별로 git 추적 외부에서 관리 ([ADR-0003](../../adr/0003-no-docker-policy-ci-scope.md)).
2. **단일 포트 진입** — 외부 진입은 `https://<host>/devhub` 만. nginx 가 backend/Keycloak 분기 ([ADR-0018](../../adr/0018-single-port-reverse-proxy-policy.md) + PR #277).
3. **자동 마이그레이션** — `db-migrate` 컨테이너가 backend boot 전 마이그레이션 적용 + 30 retry × 2s healthcheck (PR #277).
4. **추적성 자동 동기화** — 모든 PR 이 `docs/traceability/report.md` 의 §3 매트릭스 + 단계별 인덱스 갱신. ID 영구 (immutable).
5. **본인 author PR 의 4단계 self-review SOP** — diff 재검토 → gh pr comment → 보강 commit → squash merge. codex review cycle 과 병행.

---

## 8. 기술 스택 핵심

| Layer | 선택 | 이유 |
| --- | --- | --- |
| **Backend Go** | gin + pgx + ory/oathkeeper-style middleware chain | 단일 binary + low ops overhead |
| **Backend AI** | Python FastAPI + gRPC | AI 서비스 v2 carve out — v1 에는 미사용 |
| **Frontend** | Next.js 14 (App Router) + Tailwind + shadcn | semantic theme + dark/light + responsive |
| **DB** | PostgreSQL 15 + 33 migration | row-level locking + JSONB audit payload |
| **IdP** | Keycloak (외부 운영, OIDC + admin REST) | enterprise SSO + MFA + group composite role |
| **CI** | GitHub Actions + actionlint workflow lint | no-docker baseline + paths-skip |
| **Observability** | Prometheus + Grafana JSON dashboard | HomeLab puller metric + JWKS stale-while-error |

상세 — [`docs/shared/tech_stack.md`](../../shared/tech_stack.md).

---

## 9. 본 학습회 자료의 위치

본 자료는 5 step 구조:

| Step | 산출물 | 본 step 의 source |
| --- | --- | --- |
| **1** (본 문서) | 프로젝트 개요 | README + PROJECT_PROFILE + ADR 인덱스 |
| 2 | 로드맵 + 개발 문서 요약 | development_roadmap + release_v1_roadmap + 핵심 docs |
| 3 | 추적성 요약 | traceability/report (REQ → UC → ARCH → API → IMPL → UT → TC) |
| 4 | 구현 현황 | 도메인별 backend/frontend/test 활성화 상태 |
| 5 | HTML 학습회 자료 | step 1~4 의 종합 — chart.js + slide layout |

---

## 10. 다음 단계

- [`02-roadmap-and-docs.md`](./02-roadmap-and-docs.md) — 로드맵 + 개발 문서 요약 (Step 2).
- 본 step 의 ADR 인덱스 + 도메인 표는 Step 5 의 HTML 자료에서 차트 + 접기/펼치기 UI 로 시각화.
