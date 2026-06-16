# Session Handoff — feat/work_260610-v1-1-sprint-a-real-adapter

- **문서 목적**: 다음 세션이 본 sprint 의 작업 상태를 빠르게 복원하도록 핵심 사실만 기록한다.
- **sprint branch**: `feat/work_260610-v1-1-sprint-a-real-adapter` (코드 작업 완료, commit + PR 발행 대기)
- **시작일**: 2026-06-10 (PR1 시작 시점)
- **종료 시점**: 2026-06-10 (PR1 코드 작성 + 검증 완료, `go build` + `go test ./...` + `go vet` 모두 통과)
- **관련 문서**: [`work_backlog.md`](./work_backlog.md), [`backlog/2026-06-10.md`](./backlog/2026-06-10.md), [`state.json`](./state.json), [ADR-0030](../../../docs/adr/0030-sso-integrations-and-auth-session-port.md), [sprint -a follow-up session_handoff](../work_260610-v1-1-sprint-a-followup/session_handoff.md) (이전 sprint 의 carry-over 출처)

## 1. sprint 목표 (in_progress)

v1.1 sprint -a follow-up PR (PR1) — sprint -a follow-up 본 PR (PR #539) 의 carry-over C-a~C-f 풀번들.

scope (사용자 결정 2026-06-10):
- **C-a (P0)**: real adapter 작성 (`sso-integrations/keycloak/{verifier,admin_client}.go`)
- **C-b (P0)**: main.go event listener type assertion 정리
- **C-c (P0)**: `_ = keycloakEventPort` placeholder 제거
- **C-d (P1)**: v1.0 mirror struct 제거 (httpapi.KeycloakUserEvent/KeycloakAdminEvent → integration/ 의 struct 직접 정의)
- **C-e (P1)**: audit-ops mirror 통합 (cross-package, audit-ops 의 mirror struct → integration/ alias)
- **C-f (P1)**: `infra/idp/_archive_2026-06-10/` immutable archive

## 2. carry-over 출처

sprint `feat/work_260610-v1-1-sprint-a-followup` (PR #539) 의 carry-over 표 (`work_backlog.md` §carry-over). 본 PR1 종료 후 잔여 4건 (C-g traceability, C-h ADR-0030 timeline, C-i E2E matrix, C-j build tag 검토) 는 PR2.

## 3. baseline (PR #539 머지 직후, main HEAD `61a705bf`)

### 3.1 현재 위치 (canonical) 파일
- `backend-core/internal/sso-integrations/keycloak/saovae_stub.go` (105 lines) — 4 port stub + webhook handler
- `backend-core/internal/domain/auth-session/integration/ports.go` (100 lines) — 4 interface alias + 1 NEW KeycloakEventPort + 2 type alias (KeycloakUserEvent/KeycloakAdminEvent) + 3 sentinel error alias
- `backend-core/main.go` L154-235, L432-462 — runtime injection + 3 slot + type assertion

### 3.2 이전 위치 (real impl) — 본 PR1 에서 sso-integrations/keycloak/ 로 이전
- `backend-core/internal/domain/auth-session/service/keycloak_verifier.go` (477 lines) — `KeycloakJWKSVerifier` (real, JWKS + RS256/RS384/RS512 + stale-while-error)
- `backend-core/internal/httpapi/keycloak_admin_client.go` (410 lines) — `KeycloakAdminClient` (real, admin REST + OIDC logout + ListUserEvents/ListAdminEvents + KeycloakUserDetails + KeycloakGroup)
- `backend-core/internal/httpapi/keycloak_admin_client.go` 의 struct 정의: `KeycloakUserEvent`, `KeycloakAdminEvent` (httpapi package)

### 3.3 테스트 파일 (moved with real adapter)
- `backend-core/internal/domain/auth-session/service/keycloak_verifier_test.go` (853 lines) — verifier 테스트
- `backend-core/internal/httpapi/keycloak_admin_client_test.go` (57 lines) — admin client 기본 테스트
- `backend-core/internal/httpapi/keycloak_admin_client_events_test.go` (166 lines) — admin event list 테스트
- `backend-core/internal/httpapi/keycloak_events_webhook_test.go` (261 lines) — webhook ingest 테스트 (audit-ops 의 핸들러, httpapi 와 무관 — 정합 확인 필요)

### 3.4 audit-ops mirror (C-e 대상)
- `backend-core/internal/domain/audit-ops/service/keycloak_event_puller.go` 의 `KeycloakUserEvent`/`KeycloakAdminEvent` (mirror struct, L39-48 / L51-61) — `httpapi.KeycloakUserEvent` 와 동명 + 동 field 의 별도 struct
- `backend-core/internal/domain/audit-ops/service/keycloak_admin_adapter.go` 의 `HTTPAPIUserEvent`/`HTTPAPIAdminEvent` (mirror struct, L18-41) + `HTTPAPIEventLister` interface
- `backend-core/main.go` L515-563 의 `keycloakAdminEventLister` adapter struct (httpapi → audit 변환)
- main.go L463 의 `auditsvc.NewHTTPAPIEventListerAdapter(&keycloakAdminEventLister{kc: kc})` 호출

### 3.5 infra/idp/ (C-f 대상, archive 후보)
- `infra/idp/Dockerfile.keycloak` (active — dogfood build)
- `infra/idp/README.md` (active — mode 분기 문서)
- `infra/idp/identity.schema.json` (legacy — Kratos reference)
- `infra/idp/keycloak-event-listener-spi/` (active — SPI JAR Maven 빌드)
- `infra/idp/keycloak-realm.{ci,dev,prod}.json` (active — realm config)
- `infra/idp/sql/{001_create_idp_schemas,003_seed_test_admin}.sql` (active — schema/seed)
- `infra/idp/_archive_hydra_kratos/` (legacy — ADR-0001 시기 Hydra/Kratos)

## 4. 사용자 결정 사항 (이전 sprint -a follow-up 본 PR #539 + 본 PR1 진행 시 in-session)

- **Real adapter 위치**: `sso-integrations/keycloak/{verifier,admin_client}.go` (NEW). 기존 `service/keycloak_verifier.go` + `httpapi/keycloak_admin_client.go` 는 **삭제** (callers 이전 완료 후).
- **Build 정책**: `//go:build` tag 미사용, **runtime injection** (단일 binary, `DEVHUB_BUILD_TIER` env var). default = saovae_stub. `internal` = real adapter.
- **Tier 분류** (AGENTS.md §사외/사내 2-tier 정합): `sso-integrations/keycloak/verifier.go` + `admin_client.go` = **사내** (real Keycloak admin REST + event puller + webhook). `sso-integrations/keycloak/saovae_stub.go` = **사외** (saovae_stub, Keycloak 인프라 비의존). `domain/auth-session/integration/ports.go` = **공용** (interface 만).
- **Type alias 정책 (C-d)**: `httpapi.KeycloakUserEvent` / `httpapi.KeycloakAdminEvent` struct 정의 **삭제**. `integration.KeycloakUserEvent` / `integration.KeycloakAdminEvent` 는 `ports.go` 에서 **struct 직접 정의** (alias 가 아닌 본 정의). `sso-integrations/keycloak/admin_client.go` 의 `KeycloakAdminClient.ListUserEvents/ListAdminEvents` 가 `[]integration.KeycloakUserEvent/KeycloakAdminEvent` 반환.
- **audit-ops mirror 통합 (C-e)**: `audit-ops/service/keycloak_event_puller.go` 의 mirror struct (`KeycloakUserEvent` / `KeycloakAdminEvent`) 를 `type X = integration.X` **alias 로 변환**. main.go 의 `keycloakAdminEventLister` adapter + `audit-ops/service/keycloak_admin_adapter.go` 의 `HTTPAPIUserEvent`/`HTTPAPIAdminEvent` mirror + `HTTPAPIEventLister` interface **전부 제거** — `*keycloakadapter.KeycloakAdminClient` 가 `KeycloakEventPort` 충족하므로 `auditsvc.NewKeycloakEventPuller` 가 port 를 직접 받음 (단, 현재 `KeycloakEventPuller` 는 `KeycloakEventLister` interface 받음 — interface 통폐합 또는 adapter 1회 변환 유지 결정 필요).
- **infra/idp/ archive (C-f)**: `_archive_2026-06-10/` 디렉터리 immutable archive. 후보 = `identity.schema.json` (Kratos legacy, README §1 의 "reference" 분류 — archive 가능). active 자원은 보존. README.md 갱신 (archive 경로 추가).

## 5. 핵심 파일 / 라인 참조 (PR1 시작 시점)

- `backend-core/main.go:154-171` — `KeycloakJWKSVerifier` instantiation (PR1 에서 sso-integrations/keycloak/ 의 `NewKeycloakJWKSVerifier(cfg)` 호출로 변경)
- `backend-core/main.go:189-235` — runtime injection 분기 (PR1 에서 sso-integrations/keycloak/ 의 `NewKeycloakAdminClient(cfg)` + `NewIdentityAdminStub()`/`NewOIDCLogoutClientStub()`/`NewKeycloakEventPortStub()` 호출)
- `backend-core/main.go:432-462` — Keycloak event listener cron (PR1 에서 type assertion 제거 + `_ = keycloakEventPort` placeholder 제거 + `keycloakEventPort.ListUserEvents/ListAdminEvents` 직접 호출)
- `backend-core/main.go:515-563` — `keycloakAdminEventLister` adapter (PR1 에서 제거 — audit-ops 가 port 직접 받음)
- `backend-core/internal/domain/auth-session/integration/ports.go:63-69` — `KeycloakEventPort` interface
- `backend-core/internal/domain/auth-session/integration/ports.go:74,79` — `KeycloakUserEvent` / `KeycloakAdminEvent` alias (PR1 에서 struct 직접 정의로 변경)
- `backend-core/internal/sso-integrations/keycloak/saovae_stub.go:87-92` — stub 의 port 구현 reference

## 6. 알아둘 trade-off (의도적, ADR-0030 §2.3 명시 + 본 PR1 결정)

- **Real adapter 의 struct 정의 위치**: sso-integrations/keycloak/ 자체에 두지 않고 `integration/ports.go` 에 직접 두는 이유는, port caller (audit-ops) 가 canonical struct 만 보면 됨. type alias 가 sso-integrations → integration 단방향 (의존 방향 일치). sso-integrations 에서 자체 struct 를 정의하면 audit-ops 가 sso-integrations 를 직접 import 해야 함 (cycle 위험).
- **audit-ops KeycloakEventLister interface 의 미래**: 본 PR1 에서 audit-ops 의 `KeycloakEventLister` interface 와 `HTTPAPIEventLister` interface 를 **통폐합** — 단일 `integration.KeycloakEventPort` 가 canonical. main.go 가 `*keycloakadapter.KeycloakAdminClient` 를 그대로 `auditsvc.NewKeycloakEventPuller` 에 주입 (별도 adapter 불요).
- **Delete vs Deprecate**: 기존 `service/keycloak_verifier.go` + `httpapi/keycloak_admin_client.go` + 관련 test file 은 **삭제** (callers 이전 완료 후). deprecation shim 으로 유지하면 build time 에 검증되지 않는 dead code 가 누적. 본 PR1 의 e2e (Backend Unit + Integration) 가 import chain 검증.
- **`KeycloakAdminClient` 의 GetUserDetails / GetUserGroups 메서드**: admin event handler (audit-ops) 가 사용. 본 PR1 에서 sso-integrations/keycloak/ 의 새 `KeycloakAdminClient` 로 이전. callers = `audit-ops/view/handler.go` 또는 별도 admin event handler — grep 결과 확인 필요.

## 7. 다음 세션이 가장 먼저 할 일

1. `git status` / `git log --oneline -3` / `git branch --show-current` 확인 (현재 `feat/work_260610-v1-1-sprint-a-real-adapter`, main HEAD `61a705bf`).
2. C-a real adapter 작성: `sso-integrations/keycloak/{verifier,admin_client}.go` 신규.
3. ports.go 의 alias → struct 직접 정의로 변경.
4. main.go 의 instance 생성 + type assertion 정리 + placeholder 제거.
5. 기존 `service/keycloak_verifier.go` + `httpapi/keycloak_admin_client.go` + 관련 test 삭제.
6. audit-ops mirror → integration alias 통폐합.
7. infra/idp/ archive.
8. `go build ./...` + `go test ./...` + Backend Integration + E2E (backend-only 변경).
9. commit + push + PR 발행.

## 8. 알아둘 위험

1. **Import cycle**: `integration/ports.go` (struct 직접 정의) + `sso-integrations/keycloak/admin_client.go` (struct 사용) — sso-integrations → integration 단방향 OK. `httpapi` 가 `integration` 을 import 하면 architecture purity 문제 발생 가능 (httpapi 는 view layer). 본 PR1 에서 httpapi → integration 직접 import 회피 — `*KeycloakAdminClient` 가 sso-integrations 로 이동하여 httpapi 가 더 이상 본 struct 를 보지 않아도 됨.
2. **Test file migration**: 853 lines 의 keycloak_verifier_test.go 이동 시 build tag / package 변경 (service → sso-integrations/keycloak). import path 일관성.
3. **E2E shard 영향**: backend-only 변경이지만 integration test 가 backend core 에서 동작 — e2e shard 1/2/3 모두 영향 없음 (frontend 변경 0).
4. **audit-ops `KeycloakEventLister` interface callers**: `keycloak_event_puller.go` 내부 + test file. interface 통폐합 시 callers 모두 갱신.
5. **infra/idp/ archive 가 CI 영향**: docker build 가 `infra/idp/Dockerfile.keycloak` 사용 (dogfood) — archive 시 build context 변경 없음. `_archive_2026-06-10/` 디렉터리는 build 시 무시. 단 compose 의 `KEYCLOAK_REALM_IMPORT_PATH` 가 `keycloak-realm.{ci,dev,prod}.json` 직접 참조 — archive 후보에서 제외.
