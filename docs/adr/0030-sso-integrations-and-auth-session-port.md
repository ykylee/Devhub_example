# ADR-0030: auth-session 도메인의 Port interface 도입 + sso-integrations/ 분리 (v1.1 sprint -a)

- **문서 목적**: auth-session 도메인 layer 가 사내 IdP (Keycloak) 와 결합을 끊고 **Adapter pattern** 으로 port interface 만 의존하도록 분리하는 결정을 명문화한다. 실제 구현은 `sso-integrations/keycloak/` (사내 IdP infra 통합) 트리로 이전. Phase 2 (v1.2) 의 agentic RAG layer 가 본 port 를 invoke 가능.
- **범위**: `backend-core/internal/domain/auth-session/integration/ports.go` (NEW, interface 만) + `backend-core/internal/sso-integrations/` (NEW 트리, sprint -a follow-up 에서 구현) + `backend-core/main.go` runtime injection (sprint -a follow-up) + `infra/idp/` 의 immutable archive.
- **대상 독자**: Backend / IdP / AI agent / ops 트랙 담당자, 후속 sprint 작업자, owner.
- **상태**: accepted (sprint -a, 2026-06-10)
- **최종 수정일**: 2026-06-10
- **결정 근거 sprint**: `feat/work_260610-v1-1-sprint-a-sso-integrations` (v1.1 sprint -a)
- **Tier**: **공용** (interface 만 노출, 사내 한정 정보 미포함)
- **관련 문서**: [`docs/governance/worker_division.md` §6 사외/사내 2-tier 분업](../governance/worker_division.md), [`docs/governance/worker_division.md` §6.7 명명 재검토](../governance/worker_division.md#67-명명-재검토-2026-06-10), [`docs/planning/external-integrations-agentic-rag-roadmap.md` §0.4 + §3 + §6](../planning/external-integrations-agentic-rag-roadmap.md) (장기 비전), [ADR-0019](./0019-keycloak-only-idp.md) (Keycloak 단일화 — 본 ADR 의 supersession 후보), [ADR-0020](./0020-account-user-management-boundary.md) (사내 IdP 팀 ↔ DevHub 운영자 책임 매트릭스), [code-taxonomy.md §1](../code-taxonomy.md) (Domain layer purity 원칙).

## 1. 배경

### 1.1 v1.0 의 Keycloak 결합

v1.0 에서 DevHub 의 인증 layer 는 다음 3 개의 interface + 2 개의 implementation 으로 구성:

**Interface (v1.0, view/ 패키지에 정의)**:
- `backend-core/internal/domain/auth-session/view/auth.go:59` — `BearerTokenVerifier` (`VerifyBearerToken(ctx, token) (AuthenticatedActor, error)`)
- `backend-core/internal/domain/auth-session/view/handler.go:27` — `IdentityAdmin` (`FindIdentityByUserID`, `LogoutUserSession`)
- `backend-core/internal/domain/auth-session/view/handler.go:197` — `OIDCLogoutClient` (`OIDCLogout`)

**Implementation (v1.0, 사내 IdP 종속)**:
- `backend-core/internal/domain/auth-session/service/keycloak_verifier.go:23` — `KeycloakJWKSVerifier` (JWKS + RS256/RS384/RS512)
- `backend-core/internal/httpapi/keycloak_admin_client.go:35` — `KeycloakAdminClient` (Keycloak Admin REST + OIDC user logout)
- `backend-core/internal/domain/audit-ops/view/keycloak_events_webhook.go:41` — `ReceiveKeycloakEventWebhook` (Keycloak SPI webhook ingest)

**Wiring (`main.go`)**:
- L152-169 — `KeycloakJWKSVerifier` instantiation + assignment to `httpapi.BearerTokenVerifier` variable
- L179-201 — `KeycloakAdminClient` instantiation + assignment to `httpapi.IdentityAdmin` + `httpapi.OIDCLogoutClient`
- L410-440 — Event listener goroutine (Keycloak SPI event puller) with type assertion to `*httpapi.KeycloakAdminClient`

### 1.2 한계

1. **Domain layer 가 IdP 와 직접 결합**: `KeycloakJWKSVerifier` 가 `domain/auth-session/service/` 에 위치. service layer 가 외부 IdP 시스템에 종속 (Keycloak 교체 시 service 코드 직접 수정 필요).
2. **AuthSession → httpapi cycle 위험**: `KeycloakJWKSVerifier` (service) 가 `httpapi.AuthenticatedActor` (view) 를 import. `KeycloakAdminClient` (httpapi) 가 다시 service/wiring 에서 사용. cycle 위험.
3. **Agentic RAG 확장성 부재**: Phase 2 (v1.2) 의 agentic RAG 가 IdP user 자동 생성 / RBAC sync 위해 `KeycloakAdmin` 의 모든 메서드를 invoke 해야 하나, 현재는 `KeycloakAdminClient` 의 sub-set 만 노출. port 가 없으면 RAG 가 구현체에 직접 결합.
4. **Test fixture 중복**: `view/handler_test.go` 와 `view/auth_test.go` (sprint 후속) 에 `fakeBearerTokenVerifier`, `fakeIdentityAdmin`, `fakeOIDCLogoutClient` 등 **fixture 가 view/ 패키지 내부에 정의**. 본 port 의 다른 caller (audit-ops, integration-registry 등) 도 동일한 fixture 가 필요하나 중복 발생.

### 1.3 사용자 결정 (2026-06-10)

> "외부 시스템 연동 기능은 따로 분리해서 추후 agentic rag와 같이 발전시킬 계획이 있어. 이 계획에 따라 분리 계획을 세워보자."

→ **PR #535 design doc** + **2026-06-10 후속 결정** (PR #537 §0.4): Keycloak 은 `agentic-integrations/` scope 에서 제외. Keycloak 의 정확한 위치 = **`domain/auth-session/integration/ports.go` (port interface) + `sso-integrations/keycloak/` (구현)**. 본 ADR 은 그 결정을 code-level 로 적용.

## 2. 후보 옵션

### 2.1 Port interface 위치

| # | 옵션 | 장점 | 단점 | 결정 |
| --- | --- | --- | --- | --- |
| 1 | **현재 유지** (view/ 패키지에 interface 정의) | 변경 없음 | domain layer 가 IdP 와 결합, agentic RAG 확장 시 service/ view 간 cycle 위험 | ❌ |
| 2 | **domain/auth-session/integration/ports.go (NEW)** — type alias 로 view/ 의 interface re-export, 신규 `KeycloakEventPort` 만 ports.go 에 정의 | interface 의 canonical 위치, saovae stub 위치 자연스러움, agentic RAG 가 port 만 invoke | view/ 와 import cycle (auth-session 도메인이 view 를 re-export 해야 함 — 현재는 가능) | ⭐ **채택** (사용자 의도) |
| 3 | **shared/contracts/ (NEW)** — 모든 도메인의 port interface 단일 트리 | 일관된 module boundary | cross-domain import 시 over-engineering, auth-session 만 별도 위치 | ❌ (장기 옵션) |

### 2.2 실제 구현 이전 (sprint -a follow-up)

| # | 옵션 | 장점 | 단점 | 결정 |
| --- | --- | --- | --- | --- |
| 1 | **sprint -a PR 에서 모두 이전** (40+ file import path 변경) | 즉시 깨끗한 architecture | risk 큼, e2e 회귀 가능성, main.go wiring + 5+ test file 동시 변경 | ❌ (over-scope) |
| 2 | **sprint -a PR 은 port interface + saovae stub + main wiring 만**, 실제 구현 (`KeycloakJWKSVerifier` + `KeycloakAdminClient` 이동) 은 sprint -a follow-up PR | risk 최소화, 단계적 검증, 각 sprint 후 `go build` + `go test` 100% 통과 | sprint -a PR 의 본질 = "도메인 layer 가 port interface 만 의존" (구현 이동은 부수적) | ⭐ **채택** |
| 3 | **sprint -a PR 에서 구현 이동만** (port interface 미정의) | 한 PR 에 집중 | 도메인 layer 의 IdP 결합이 본 PR 이후도 지속 | ❌ |

### 2.3 saovae stub 전략 (sprint -a follow-up)

| # | 옵션 | 장점 | 단점 | 결정 |
| --- | --- | --- | --- | --- |
| 2 | **Runtime injection** (main.go 에서 `DEVHUB_BUILD_TIER` env var 로 분기) | 단일 binary, CI 단순, saovae default | stub 이 binary 에 항상 포함 (binary size) | ⭐ **채택** (PR #535 design doc §3.4 권고) — **2026-06-10 confirmed (ADR-0031 §4 재평가)**: sprint -a follow-up PR #540 (real adapter) + PR #542 (e2e-internal job) 머지 후 정량 측정 결과, binary overhead < 5KB (전체 backend-core < 50MB 대비 0.01%) vs build tag 전환 시 CI matrix 2배 + +5~10 file 변경. runtime injection 유지 결정 confirmed. **2026-06-12 partial supersession (baseline 변경 정공법)** (sprint `fix/work_260612-6-e2e-internal-removal`): e2e-internal job 폐기 (사용자 결정: 사내 환경용 셋팅) → runtime injection 결정과 독립. ADR-0031 §2.3 / §3.1 / §3.2 / §4 / §5 / §6.1 / §7.1 / §8 baseline 변경 (CI matrix 1쌍 → 단일 matrix). 결론 자체는 변동 0건 (runtime injection 유지). 사내 staging/prod-smoke 가 real adapter 검증 책임. |

## 3. 결정

**옵션 2.1-2 + 옵션 2.2-2 + 옵션 2.3-2**.

즉:
- `domain/auth-session/integration/ports.go` (NEW) — 4 interface (3 type alias + 1 NEW `KeycloakEventPort`) + 3 sentinel error (value alias)
- view/ 의 기존 interface 는 **deprecated alias** 로 유지 (backward compat)
- sprint -a follow-up PR 에서: `sso-integrations/keycloak/` 실제 구현 + saovae_stub + main.go runtime injection + `infra/idp/_archive_2026-06-10/` immutable archive

## 4. 적용 — sprint -a (본 PR)

### 4.1 새 파일: `domain/auth-session/integration/ports.go`

**Type aliases** (view/ 의 interface re-export, backward compat):
- `type AuthenticatedActor = view.AuthenticatedActor`
- `type BearerTokenVerifier = view.BearerTokenVerifier`
- `type IdentityAdmin = view.IdentityAdmin`
- `type OIDCLogoutClient = view.OIDCLogoutClient`

**NEW port** (ports.go 에서 직접 정의):
- `type KeycloakEventPort interface { ListUserEvents(...); ListAdminEvents(...) }`
- `type KeycloakUserEvent struct {...}` (mirror — audit-ops 의 mirror 와 동일 정의, circular import 회피)
- `type KeycloakAdminEvent struct {...}`

**Sentinel error aliases** (view/handler.go 및 httphelp 의 value re-export):
- `var ErrOIDCConfigMissing = view.ErrOIDCConfigMissing`
- `var ErrOIDCNetworkUnreachable = view.ErrOIDCNetworkUnreachable`
- `var ErrIdentityNotFound = httphelp.ErrIdentityNotFound`

### 4.2 변경 안 함 (sprint -a PR 의 한계)

- **view/ 의 interface 삭제** 안 함 (backward compat). `deprecation` comment 만 추가.
- **`KeycloakJWKSVerifier` 의 `domain/auth-session/service/` → `sso-integrations/keycloak/` 이전** 안 함. sprint -a follow-up.
- **`KeycloakAdminClient` 의 `httpapi/` → `sso-integrations/keycloak/` 이전** 안 함. sprint -a follow-up. **cycle 회피** 중요 — 현재 httpapi 가 service 에서 import 되는 것 의 reverse 위험.
- **`main.go` runtime injection** 안 함. sprint -a follow-up.
- **`infra/idp/_archive_2026-06-10/` immutable archive** 안 함. sprint -a follow-up (별도 PR).

| Phase | Sprint | Status |
|---|---|---|
| 1.1a (Keycloak port interface + interface 의 canonical 위치) — 본 PR (sprint -a) | v1.1 sprint -a | **accepted (P1), done (PR #538 머지, 2026-06-10)** |
| 1.1b (sso-integrations/ 실제 구현 + saovae stub + main wiring + infra/idp archive) | v1.1 sprint -a follow-up | **accepted (P1), done (PR #539 + PR #540 머지, 2026-06-10)** |
| 1.2 (gitea + ci port) | v1.1 sprint -b | planned (P1) |
| 1.3 (hrdb + commandworker + serviceaction) | v1.1 sprint -c | planned (P1) |
| 1.4 (homelab + adapters 통합 + legacy archive) | v1.1 sprint -d | planned (P1) |
| 2.2 (Agentic planner + tool registry + ssoKeycloakPort 도 invoke) | v1.2 sprint -b | planned (P1) |
| C-h (ADR-0030 §5 timeline + traceability report.md IMPL row 갱신) | `docs/work_260610-traceability-impl-sso-keycloak` PR | **done (2026-06-10)** — §5 1.1a + 1.1b status accepted/done 명시 + `docs/traceability/report.md` §2.4 + §3.1/§3.3 matrix 갱신 + §4 ADR-0030 row 신규 + §6 변경 이력 row 신규. |

## 6. Cross-tier impact

### 6.1 tier 매핑

| Module | Tier | 비고 |
|---|---|---|
| `domain/auth-session/integration/ports.go` | **공용** | interface 만, no I/O. 사외 + 사내 양쪽 build 가 import 가능. |
| `sso-integrations/keycloak/` (sprint -a follow-up) | **사내** | Keycloak admin REST + event puller + webhook. 사내 build 시 wiring. |
| `sso-integrations/keycloak/saovae_stub.go` (sprint -a follow-up) | **사외** (build tag) | 사외 build/test 용 stub (real Keycloak 없이도 작동) |
| `domain/auth-session/integration/saovae_stub.go` (sprint -a follow-up) | **사외** (build tag) | 본 port 의 BearerTokenVerifier 사외 stub |

### 6.2 .gitignore / CI / ADR 업데이트 필요

- **.gitignore**: `sso-integrations/_archive_2026-06-10/` (sprint -a follow-up) 는 추적 (immutable, §4.2 ADR).
- **CI**: `go build ./...` (사외), `go build -tags internal` (사내 runner, sprint -a follow-up), e2e 는 양쪽 동일.
- **ADR**: 새 ADR 후보 — "sprint -a follow-up: sso-integrations/ 실제 구현 + saovae_stub + main wiring" (별도 PR).
- **tier lint** (`scripts/check-tier-separation.sh`): `domain/auth-session/integration/ports.go` 가 `DEVHUB_KEYCLOAK_*` 같은 사내 한정 env var 를 직접 read 안 함 → lint 통과.

## 7. Risks + Open questions

### 7.1 Risks

1. **Import cycle**: `domain/auth-session/integration/` 가 `domain/auth-session/view/` 를 re-export. **view 가 integration 을 import** 하면 cycle. `audit-ops/service/keycloak_event_puller.go` 의 `KeycloakEventLister` 가 본 port 와 정합 시 cycle 검증 필수. sprint -a follow-up 에서 `KeycloakEventLister` → `KeycloakEventPort` migration 시 확인.
2. **sprint -a follow-up 의 cycle 위험**: `KeycloakAdminClient` 의 `httpapi/` → `sso-integrations/keycloak/` 이동 시. `httpapi` 가 본 admin client 를 import 하는 다른 caller (audit-ops/view/keycloak_events_webhook.go 의 type assertion) 가 있을 수 있음. **type assertion 제거 후에만 이동 가능**. sprint -a follow-up 의 1차 step 으로 type assertion 정리.
3. **Sprint -a PR 의 가시성**: 본 PR 의 실질적 코드 변경 = ports.go 신규 1 file. **큰 architecture 결정**임에도 **diff 는 작음**. 후속 reviewer 가 "왜 이 결정이 필요한가" 를 PR body 로 충분히 설명해야 함.

### 7.2 Open questions

1. **KeycloakUserEvent / KeycloakAdminEvent 의 canonical 위치**: ports.go 에 정의 vs `sso-integrations/keycloak/events.go` (adapter) + port 는 interface 만. **현재**: ports.go 에 정의 (mirror 와 동일). sprint -a follow-up 에서 v1.0 의 httpapi/ mirror 와 통합 결정.
2. **Sprint -a follow-up 의 검증 환경**: `sso-integrations/keycloak/` 가 import 되는 시점에 `go build` (사외, default) 가 stub 으로 통과해야 함. **stub 의 정확성** (sprint -a follow-up 의 1차 review item) 이 매우 중요.
3. **Phase 2 (v1.2) 의 ssoKeycloakPort 노출**: agentic RAG layer 가 user 자동 생성 시 `ssoKeycloakPort.CreateUser` 등 호출. Phase 1 의 port 가 RBAC impersonation 까지 포함할지 (sprint -a follow-up 의 `IdentityAdmin` 확장 여부) 결정 필요.

## 8. 결정 supersession

본 ADR 자체는 supersede 되지 않음. 다만 다음 sprint -a follow-up PR 이 본 ADR 의 §4.2 (실제 구현 이전) 를 실행할 때, **별도 ADR 추가 불필요** — 본 ADR 의 §5 timeline 에 이미 "1.1b (sso-integrations/ 실제 구현 + saovae stub + main wiring + infra/idp archive)" 로 등재.

후속 Phase 2 (v1.2) 의 agentic RAG 진입 시 새 ADR — "external integration agentic RAG path" 가 후보 (PR #535 design doc §6.2.4).

## 9. 변경 이력

| 일자 | 변경 | sprint |
| --- | --- | --- |
| 2026-06-10 | 1차 작성 — auth-session 도메인의 port interface 도입 + sso-integrations/ 분리 결정. view/ 의 interface 는 deprecated alias 로 backward compat. sprint -a follow-up 에서 실제 구현 이전. 사용자 2026-06-10 결정 (외부 시스템 연동 = agentic RAG 와 함께 발전) + PR #537 §0.4 (Keycloak 분류 재정의) 의 code-level 적용. | `feat/work_260610-v1-1-sprint-a-sso-integrations` |
| 2026-06-10 | §5 결정 timeline 갱신 — 1.1a (sprint -a, port interface) status = **accepted/done** (PR #538 머지) + 1.1b (sprint -a follow-up, real adapter + saovae_stub + main wiring + infra/idp archive) status = **accepted/done** (PR #539 + PR #540 머지). C-h (ADR timeline + traceability 정합 PR) row 신규 — `docs/work_260610-traceability-impl-sso-keycloak` PR 의 후속 정합. | `docs/work_260610-traceability-impl-sso-keycloak` |
| 2026-06-10 | §2.3 결정 (옵션 2 runtime injection) row 에 [ADR-0031 §4 재평가](./0031-build-tag-policy-review.md) confirmed reference 추가. sprint -a follow-up PR #540 (real adapter) + PR #542 (e2e-internal job) 머지 후 정량 측정 결과 본 결정 유지. supersession X. | `docs/work_260610-c-j-build-tag-review` |
| 2026-06-12 | **partial supersession (baseline 변경 정공법)** — e2e-internal job 폐기 결정 (사용자: "사내 환경용 셋팅, GitHub Action 으로 체크 불요") 정합. 본 ADR §2.3 runtime injection 결정과 e2e-internal job 폐기 = **독립 결정** (runtime injection 유지, e2e-internal 만 폐기). ADR-0031 partial supersession 정공법 정합. CI matrix 1쌍 → 단일 matrix (e2e shard 1/2/3 saovae_stub default 만). 사내 staging/prod-smoke 가 real adapter 검증 책임. **결론 변동 0건** (runtime injection 결정 = confirmed). 신규 ID 발급 0건 (housekeeping follow-up 정공법). | `fix/work_260612-6-e2e-internal-removal` |
