# 외부 시스템 연동 + Agentic RAG 통합 Roadmap (2026-06-10 결정)

- **문서 목적**: 현재 `backend-core/internal/infrastructure/` (Keycloak, Gitea, CI, commandworker, HRDB, serviceaction) + `backend-core/internal/integrations/adapters/` (homelab, task_item_puller, metrics) 의 외부 시스템 연동 기능을 (1) `agentic-integrations/` 로 명명 재구성 + (2) agentic RAG 기반 long-term 진화 방향을 정리한다.
- **범위**: backend domain layer 와의 의존 방향, adapter 패턴, agentic RAG 통합 비전, P0~P3 마일스톤.
- **대상 독자**: backend / frontend / AI / ops 트랙 담당자, 후속 sprint 작업자, owner.
- **상태**: planned (draft, 2026-06-10) — 사용자 결정 (2026-06-10) 의 long-term 비전 정리.
- **결정 근거**: [PR #531 §6.7 명명 재검토](../governance/worker_division.md#67-명명-재검토-2026-06-10) + 사용자 2026-06-10 결정 (외부 시스템 연동 = agentic RAG 와 함께 발전).
- **관련 문서**: [`docs/governance/worker_division.md` §6 사외/사내 2-tier 분업](../governance/worker_division.md), [`docs/governance/worker_division.md` §6.3.3 directory tier SoT](../governance/worker_division.md#633-tier-매핑-sot--1차-분류), [`docs/planning/external_integration_capability_matrix.md`](./external_integration_capability_matrix.md), [`docs/planning/external_system_integration_concept.md`](./external_system_integration_concept.md), [ADR-0019 Keycloak 단일화](../adr/0019-keycloak-only-idp.md), [ADR-0025 봉투 암호화](../adr/0025-envelope-encryption-key-management.md).

## 0. 배경

### 0.1 현재 상태

`backend-core/internal/infrastructure/` 디렉터리는 외부 시스템 어댑터 (사내 한정 tier=internal):
- `keycloak-idp` (향후 도메인 이전) — Keycloak admin REST / event puller
- `gitea-scm` — Gitea REST + webhook HMAC + sync worker
- `ci` — Gitea Actions adapter (CI run / log)
- `commandworker` — approved command 실행 (현재 dry-run only)
- `hrdb` — 사내 HR DB adapter (PostgreSQL `hrdb.persons`)
- `serviceaction` — service action executor (현재 simulation mode)
- `database-migration` — golang-migrate chain
- `deployment-automation` — deploy scripts

`backend-core/internal/integrations/adapters/` 는 cross-cutting adapter:
- `homelab*` — 사내 HomeLab agent polling
- `task_item_puller` — 외부 task (Jira 등) pull
- `metrics` — Prometheus

### 0.2 한계

1. **명명 일관성 부재**: `infrastructure/` 와 `integrations/adapters/` 의 경계 모호. "외부 시스템 연동" 이라는 동일 카테고리가 두 디렉터리에 분산.
2. **도메인 의존 직접**: `domain/integration-registry/view/` 가 `infrastructure/gitea` / `infrastructure/hrdb` 를 직접 import. 도메인 layer 의 "외부 기술 무관" 원칙 위배 가능 ([code-taxonomy §1](../governance/code-taxonomy.md)).
3. **Agentic 확장성 부재**: 현재 모든 integration 은 **동기적 REST/HTTP 호출**. 운영자가 "최근 Gitea sync 가 느려졌나?" 같은 자연어 쿼리를 던질 수 없고, integration 이 context (운영 history + monitoring + spec) 를 retrieve 해서 자동화할 수 없음.

### 0.3 사용자 결정 (2026-06-10)

> "외부 시스템 연동 기능은 따로 분리해서 추후 agentic rag와 같이 발전시킬 계획이 있어. 이 계획에 따라 분리 계획을 세워보자."

→ **현재 `infrastructure/` 와 `integrations/adapters/` 의 통합** + **agentic RAG 기반 long-term 진화** = 본 design doc 의 목표.

## 1. 후보 옵션

### 1.1 명명 + 디렉터리 구조

| # | 옵션 | 장점 | 단점 | 결정 |
| --- | --- | --- | --- | --- |
| 1 | **현재 유지** (`infrastructure/` + `integrations/adapters/`) | 변경 없음, 기존 sprint 즉시 가능 | 모호한 boundary, agentic RAG 확장 시 일관성 결여 | ❌ |
| 2 | **`agentic-integrations/` 통합** (현재 `infrastructure/` + `integrations/adapters/` 모두 흡수, 단일 트리) | 단일 진입점, agentic RAG 진화 시 일관된 module 경계, 신규 `agentic/` sub-tree 추가 용이 | 기존 import path 변경 (40+ 파일), migration sprint 필요 | ⭐ **채택** (사용자 의도) |
| 3 | **`integrations/` 통합 (agentic- 미포함)** | 단순 통합 | agentic RAG 진화 시 재구조 필요 | ❌ |

### 1.2 도메인 의존 방향 (adapter pattern)

| # | 옵션 | 장점 | 단점 | 결정 |
| --- | --- | --- | --- | --- |
| 1 | **현재 유지** (domain → infrastructure 직접 import) | 단순함 | 도메인 layer 가 사내 시스템 종속 (code-taxonomy §1 위배) | ❌ (장기 부적합) |
| 2 | **Adapter pattern** — domain layer 는 `IntegrationPort` interface 만 의존, 구현은 `agentic-integrations/<system>/` adapter | 도메인 layer purity 유지, 사외 build 시 stub adapter 로 도메인만 build 가능, agentic RAG layer 가 adapter 로 invoke | adapter 파일 수 증가, build tag / DI discipline 필요 | ⭐ **채택** (long-term) |
| 3 | **Hexagonal architecture (full ports & adapters)** | 가장 pure | over-engineering, 현 규모에서 비용 큼 | ❌ (장기 옵션) |

### 1.3 agentic RAG 진화 깊이

| # | 옵션 | 장점 | 단점 | 결정 |
| --- | --- | --- | --- | --- |
| 1 | **RAG retrieve 만** (외부 시스템 spec/history retrieve) — agentic plan 없음 | 단순, 빠른 구현 | 자동화 한계 | ❌ (사용자 의도와 부합도 낮음) |
| 2 | **RAG + agentic plan + tool invocation** (단일 system scope) | multi-step 자동화 | 구현 복잡 | ⭐ **채택** (사용자 2026-06-10 결정 = "agentic RAG 자동화") |
| 3 | **RAG + multi-agent + cross-system orchestration** | full automation | 현 시점 over-engineering | ❌ (장기 옵션, v2.0) |

## 2. 결정

**옵션 1.1-2 (agentic-integrations/ 통합) + 옵션 1.2-2 (adapter pattern) + 옵션 1.3-2 (RAG + agentic + tool invocation)**.

즉:
- `backend-core/internal/infrastructure/` + `backend-core/internal/integrations/adapters/` → `backend-core/internal/agentic-integrations/` 로 단일 트리 통합.
- 각 system (gitea, keycloak, hrdb, ...) 은 sub-module. `domain/<name>/integration/ports/` 에 interface 정의, `agentic-integrations/<system>/` 에 구현.
- 신규 `agentic-integrations/_agentic/` sub-tree 가 RAG retrieve + agentic plan + tool invocation layer 책임. v1.0 release 직후 v1.1 에서.

## 3. Phase 1 — 디렉터리 통합 + Adapter Pattern 적용 (P1, v1.1)

> **Scope**: 코드 마이그레이션 + adapter interface 정의. agentic RAG 자체는 v1.2 이후.

### 3.1 새 디렉터리 구조

```
backend-core/internal/
├── agentic-integrations/                          # 통합 외부 시스템 연동 (현재 infrastructure/ + integrations/adapters/ 흡수)
│   ├── _archive_2026-06-10/                       # 구 infrastructure/ + integrations/adapters/ 의 history 보존
│   │
│   ├── keycloak/                                  # 사내
│   │   ├── admin_client.go                         # Keycloak Admin REST
│   │   ├── event_puller.go                         # Keycloak event poll
│   │   ├── port.go                                # KeycloakPort interface (domain layer 가 의존)
│   │   └── saovae_stub.go                         # 사외 build 용 stub (build tag 또는 runtime injection)
│   │
│   ├── gitea/                                     # 사내
│   │   ├── client.go                              # Gitea REST
│   │   ├── worker.go                              # sync worker
│   │   ├── webhook_hmac.go                         # HMAC verify
│   │   ├── port.go                                # GiteaPort interface
│   │   └── saovae_stub.go
│   │
│   ├── hrdb/                                      # 사내
│   │   ├── postgres.go
│   │   ├── mock.go
│   │   ├── port.go
│   │   └── saovae_stub.go
│   │
│   ├── homelab/                                   # 사내
│   │   ├── file_puller.go
│   │   ├── http_puller.go
│   │   ├── pull_loop.go
│   │   ├── port.go
│   │   └── saovae_stub.go
│   │
│   ├── commandworker/                             # 사내
│   │   ├── worker.go                              # dry-run executor
│   │   ├── live_worker.go                         # real executor (gated)
│   │   └── port.go
│   │
│   ├── serviceaction/                             # 사내
│   │   ├── executor.go                            # simulation mode + real mode
│   │   └── port.go
│   │
│   ├── ci/                                        # 사내 (Gitea Actions adapter)
│   │   ├── gitea_adapter.go
│   │   └── port.go
│   │
│   ├── metrics/                                   # 사외 (Prometheus)
│   │   └── counter.go
│   │
│   └── _ports/                                    # 도메인 layer 가 의존하는 통합 interface
│       ├── keycloak_port.go                        # type KeycloakPort interface { ... }
│       ├── gitea_port.go
│       ├── hrdb_port.go
│       ├── homelab_port.go
│       └── ...
│
├── domain/
│   ├── auth-session/                              # 사내 (Keycloak)
│   │   ├── view/auth.go
│   │   ├── view/handler.go
│   │   └── integration/                            # NEW: 도메인 integration port
│   │       └── ports.go                            # import "agentic-integrations/_ports" 만. 사외 build 시 stub 자동 주입.
│   │
│   ├── integration-registry/                       # 사외 (core) / 사내 (credentials)
│   │   ├── view/integration.go
│   │   ├── repository/                             # DB layer (DevHub DB)
│   │   └── integration/
│   │       └── ports.go                            # import "agentic-integrations/_ports" 만
│   │
│   └── ... (기타 8 domain 동일 패턴)
```

### 3.2 Port interface 예시 (sketch)

```go
// backend-core/internal/agentic-integrations/_ports/keycloak_port.go
package ports

import "context"

type KeycloakPort interface {
    // Admin REST
    GetUser(ctx context.Context, userID string) (*KeycloakUser, error)
    CreateUser(ctx context.Context, user KeycloakUserCreate) (string, error)
    UpdateUser(ctx context.Context, userID string, patch KeycloakUserPatch) error
    DeleteUser(ctx context.Context, userID string) error

    // Event stream
    PollEvents(ctx context.Context, since time.Time) ([]KeycloakEvent, error)

    // OIDC client
    GetServiceAccountToken(ctx context.Context) (string, error)
}

type KeycloakUser struct { ... }
type KeycloakUserCreate struct { ... }
type KeycloakUserPatch struct { ... }
type KeycloakEvent struct { ... }
```

```go
// backend-core/internal/domain/auth-session/integration/ports.go
package integration

import "github.com/devhub/backend-core/internal/agentic-integrations/_ports"

type KeycloakAdmin = ports.KeycloakPort  // type alias for brevity

// 도메인 layer 는 interface 만 의존. 구현은 main.go 에서 주입.
```

```go
// backend-core/internal/agentic-integrations/keycloak/saovae_stub.go
//go:build saovae || dev

package keycloak

import "context"

func NewKeycloakPort() ports.KeycloakPort {
    return &stubKeycloakPort{
        users: map[string]ports.KeycloakUser{ /* dev fixture */ },
    }
}

type stubKeycloakPort struct { ... }
```

```go
// backend-core/main.go (wiring)
import (
    keycloakadapter "github.com/devhub/backend-core/internal/agentic-integrations/keycloak"
    "github.com/devhub/backend-core/internal/domain/auth-session/integration"
)

func main() {
    if os.Getenv("DEVHUB_BUILD_TIER") == "internal" {
        // 사내 build: real Keycloak adapter
        integration.SetKeycloakAdmin(keycloakadapter.NewKeycloakPortFromEnv())
    } else {
        // 사외 build: stub adapter (dev/test)
        integration.SetKeycloakAdmin(keycloakadapter.NewKeycloakPort())
    }
    ...
}
```

### 3.3 Migration sprint 분할

| Sprint | Scope | Risk | DoD |
|---|---|---|---|
| **v1.1 sprint -a** | 신규 `agentic-integrations/_ports/` 정의 + `keycloak/` 만 새 구조로 이전 + `domain/auth-session/integration/ports.go` 추가 + `main.go` wiring + 1 saovae stub | 낮음 (auth-session 만, 1 system) | `go build` 양쪽 tier, `go test` 100% |
| **v1.1 sprint -b** | `gitea/` + `ci/` 이전 + `domain/integration-registry` + `domain/repository-integration` ports 추가 | 중간 (multi-system, gitea sync worker 영향) | PR #518 admin api-keys + PR #528 영향 없음, e2e 통과 |
| **v1.1 sprint -c** | `hrdb/` + `commandworker/` + `serviceaction/` 이전 + `domain/organization-management` + `domain/audit-ops` ports 추가 | 중간 (HRDB adapter 가 사내 production 의존) | staging dogfood 검증 |
| **v1.1 sprint -d** | `homelab/` + `integrations/adapters/{homelab*,task_item_puller*}` + `metrics/` 이전 + `domain/infra-topology` ports 추가 + `infrastructure/` 디렉터리 legacy archive | 낮음 (legacy archival, 신규 build) | 새 구조로 100% build, legacy `_archive_2026-06-10/` immutable 보존 |

각 sprint 후 `go test ./...` + `go build -tags saovae` + `go build -tags internal` 모두 통과. **기존 `infrastructure/` + `integrations/adapters/` 디렉터리는 `agentic-integrations/_archive_2026-06-10/` 로 이전 (immutable, §4.2 ADR)**.

### 3.4 Build tag / Runtime injection

**Option A (build tag)**:
```bash
# 사외 build (default)
go build ./...
# 사내 build
go build -tags internal ./...
```

**Option B (runtime injection)**:
```go
// main.go
tier := os.Getenv("DEVHUB_BUILD_TIER")  // external | internal
if tier == "internal" {
    integration.SetKeycloakAdmin(realKeycloakAdapter)
} else {
    integration.SetKeycloakAdmin(stubKeycloakAdapter)  // default
}
```

**결정: Option B (runtime injection)**. 이유:
- 단일 binary 로 두 환경 deploy 가능 (CI 단순)
- 사내 deploy 시 env var 하나만 추가하면 됨
- Build tag 의 `saovae` / `internal` 분리는 Go test 에서만 사용 (테스트 fixture)

## 4. Phase 2 — Agentic RAG Layer (P1, v1.2, 후속)

> **Scope**: Phase 1 의 port interface 위에 RAG retrieve + agentic plan + tool invocation layer 추가. **본 phase 는 code change 없음, design 만**. 실제 구현은 v1.2 release 후.

### 4.1 Architecture (3-layer agentic)

```
[Operator 자연어 query]
       │
       ▼
┌────────────────────────────────────────┐
│  Agent Layer (Python service, NEW)     │  ← RAG-aware planning + tool selection
│  ┌──────────────────────────────────┐  │
│  │ RAG Retriever                     │  ← 사내 RAG index (외부 시스템 spec + history + monitoring)
│  │ - keycloak spec / changelog      │  │
│  │ - gitea API reference + history  │  │
│  │ - hrdb schema + department list   │  │
│  │ - homelab service map + alerts   │  │
│  └──────────────────────────────────┘  │
│  ┌──────────────────────────────────┐  │
│  │ Agentic Planner                   │  ← LLM + tool registry
│  │ - multi-step plan                 │  │
│  │ - tool selection (per query)     │  │
│  │ - approval gate (for mutation)   │  │
│  └──────────────────────────────────┘  │
│  ┌──────────────────────────────────┐  │
│  │ Tool Invoker (Phase 1 port 재사용) │  ← keycloakPort, giteaPort, hrdbPort, ...
│  │ - HTTP/gRPC call                  │  │
│  │ - result normalize                │  │
│  │ - audit emit                      │  │
│  └──────────────────────────────────┘  │
└────────────────────────────────────────┘
       │
       ▼
[agentic-integrations/<system>/]  (Phase 1 에서 구축)
       │
       ▼
[사내 Keycloak / Gitea / HRDB / HomeLab]
```

### 4.2 Use case 예시 (long-term vision, v1.2+)

| Operator query | Agent plan | Tool invocation |
|---|---|---|
| "최근 7일 Gitea sync latency" | retrieve monitoring + sync history | giteaPort.GetSyncMetrics(7d) |
| "Keycloak 의 staging realm user 가 dev realm 으로 leak 됐나?" | cross-check + history retrieve | keycloakPort.ListUsers(staging) ∩ keycloakPort.ListUsers(dev) |
| "신규 입사자 onboarding SOP 자동화" | retrieve RBAC + org unit default + devhub convention | hrdbPort.GetNewHires(week) → rbac matrix lookup → keycloakPort.CreateUser |
| "이 PR 의 Keycloak config 변경이 운영 realm 에 영향?" | diff retrieve + RAG lookup | (approval gate) |
| "HomeLab service 가 degraded 일 때 자동 restart" | alert retrieve + runbook RAG | homelabPort.GetServiceStatus → serviceactionPort.RestartService (approval) |

### 4.3 Safety + governance

- **Mutation tool = approval gate 필수** (사내 operator 명시적 confirm)
- **Read-only tool = 자동 실행** (예: metrics / history / list)
- **Audit trail**: 모든 tool invocation 은 `audit.agentic_invocation` 으로 audit log emit
- **RBAC**: agentic layer 가 operator 의 RBAC 을 위임 받음 (impersonation)
- **Tier**: agentic RAG 자체는 **사내 한정** (RAG index 가 사내 spec/history/secret 보유)

## 5. Cross-tier impact

### 5.1 사외/사내 tier 매핑 (Phase 1 적용 후)

| Module | Tier | 비고 |
|---|---|---|
| `agentic-integrations/_ports/` | **사외 (interface only)** | Go interface, no I/O. 사외 build 에서 import 가능. |
| `agentic-integrations/keycloak/` 등 (각 system 구현) | **사내** | 사내 시스템 어댑터. 사내 build 시에만 wiring. |
| `agentic-integrations/keycloak/saovae_stub.go` | **사외** (build tag) | 사외 build/test 용 stub |
| `agentic-integrations/metrics/` | **사외** | Prometheus adapter (generic) |
| `agentic-integrations/_agentic/` (Phase 2) | **사내** | RAG + agentic + tool invoker. RAG index 가 사내 spec 보유. |

### 5.2 .gitignore / CI / ADR 업데이트 필요

- **.gitignore** (`agentic-integrations/_archive_2026-06-10/` 는 추적, `agentic-integrations/keycloak/saovae_stub.go` 는 build tag 로 분리 — 추적 유지)
- **CI**: `go build` (사외), `go build -tags internal` (사내 runner), e2e 는 양쪽 동일
- **ADR**: 새 ADR 후보 — "external integration agentic RAG path" (Phase 2 진입 시점)
- **tier lint** (`scripts/check-tier-separation.sh`): `agentic-integrations/_ports/` 가 `DEVHUB_KEYCLOAK_*` 같은 사내 한정 env var 를 직접 read 안 함 (port interface 만 노출) → lint 통과

## 6. 결정 timeline (장기)

| Phase | Sprint (v1.1) | Sprint (v1.2) | Status |
|---|---|---|---|
| 1.1 (keycloak port + stub) | v1.1 sprint -a | — | planned (P1) |
| 1.2 (gitea + ci port) | v1.1 sprint -b | — | planned (P1) |
| 1.3 (hrdb + commandworker + serviceaction) | v1.1 sprint -c | — | planned (P1) |
| 1.4 (homelab + adapters 통합 + legacy archive) | v1.1 sprint -d | — | planned (P1) |
| 1.5 (build tag 정리 + saovae_stub 100% coverage) | v1.1 sprint -e | — | planned (P1) |
| 2.1 (RAG index 구축 — 사내 spec/history vector DB) | — | v1.2 sprint -a | planned (P1) |
| 2.2 (Agentic planner + tool registry + approval gate) | — | v1.2 sprint -b | planned (P1) |
| 2.3 (operator UI: 자연어 query → agentic plan → result) | — | v1.2 sprint -c | planned (P2) |
| 2.4 (multi-agent + cross-system orchestration) | — | v2.0 | future |

## 7. Risks + Open questions

### 7.1 Risks

1. **Migration risk**: 40+ file 의 import path 변경. 4 sprint 분할로 완화, 각 sprint 후 `go build` + `go test` 100% 통과 필수.
2. **agentic RAG safety**: mutation tool 의 approval gate 가 사내 operator 의 workflow 와 정합 필수. Sprint -b 단계에서 RBAC 위임 정책 결정.
3. **RAG index staleness**: 사내 spec/history 가 vector DB 에 stale 로 들어가면 agentic plan 이 잘못된 tool 호출. 주기적 index refresh + staleness audit.
4. **observability cost**: agentic invocation 의 audit + metric + trace 가 사내 observability (Prometheus) 에 부하. 비용 추정 후 v1.2 진입 시점에 결정.

### 7.2 Open questions

1. **Build tag vs runtime injection**: 본 design 은 runtime injection 채택. v1.0 release 직전 재검토 (CI 단순성 vs binary size trade-off).
2. **Port granularity**: `KeycloakPort` 같은 interface 가 너무 크면 mock 부담. sub-port (`KeycloakAdminPort`, `KeycloakEventPort`, `KeycloakOIDCClientPort`) 분리 여부는 v1.1 sprint -a 에서 결정.
3. **Legacy archive 보존 기간**: `agentic-integrations/_archive_2026-06-10/` 의 immutable 보존 기간. v1.5 이후 삭제 가능? (사내 정책 결정)
4. **agentic RAG 의 LLM 선택**: 사내에서 자체 호스팅 (Llama / Mistral) vs 외부 API (OpenAI / Anthropic). v1.2 sprint -a 에서 결정. **Tier 영향**: 외부 API 사용 시 API key 가 사내 secret 으로 추가.

## 8. 변경 이력

| 일자 | 변경 | sprint |
| --- | --- | --- |
| 2026-06-10 | 1차 작성 — 사용자 2026-06-10 결정 (외부 시스템 연동 = agentic RAG 와 함께 발전) 의 long-term 비전 + Phase 1~2 마일스톤 + 옵션 결정. | `docs/work_260610-tier-onboarding-and-agentic-rag-design` |
