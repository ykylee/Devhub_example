# Session Handoff — main (2026-06-09, v1.0 출시 직전 finalizing — PR #514 + PR #515 머지 + codex P2 fix)

- 문서 목적: PR #514 (voc + notification, ADR-0028) + PR #515 (옵션 A N-12 housekeeping + B voc list + C N-10 IT 3 TC + codex P2 fix 3 layer) 머지 상태 인계.
- 범위: 본 세션의 2 PR (PR #514 + PR #515 squash). 옵션 A (N-12 housekeeping) + B (voc list API) + C (N-10 backend IT 3 TC) + codex P2 fix (3 layer: production router mount + routePermissionTable + gin path conflict).
- 상태: main `f7d2705` (PR #515 squash) + PR #514 (squash) 모두 머지 완료. main HEAD `897953c` (PR #503 housekeeping 기준) + 이후 06-09~06-10 v1.1 sprint -a follow-up PR #538/539/540/541/542/543 + tier-governance / branding / agentic-rag 등 다수 PR 머지. main 최신 HEAD `fee06d4` (2026-06-10 housekeeping `chore(memory)`).
- 최종 수정일: 2026-06-10 (handoff 본문 마지막 갱신, 다음 cross-check: 2026-06-10 23:45 KST)

## 0. 본 세션 핵심 결과 (2026-06-09, v1.0 출시 직전 finalizing)

### PR 머지 / Push 결과

| PR | 상태 | 의의 |
| --- | --- | --- |
| **#514** (voc + notification, ADR-0028) | ✅ MERGED (squash) | 외부 시스템 의뢰 staging 도메인 + 9 field + in-app notification + 5 API. `(source_system, external_ref)` UNIQUE for idempotency. ADR-0028 §3 옵션 1 (별도 도메인 + 1:1 dev-request 매핑) 채택. 12 file +1043 line. main `ba7823f`. |
| **#515** (v1.0 출시 직전 finalizing) | ✅ MERGED (squash, `f7d2705`, 06-09 07:08 UTC) | 옵션 A (N-12 housekeeping) + B (voc list API) + C (N-10 backend IT 3 TC) + codex P2 fix (PR #514 latent 회귀 3 layer 동시 fix). 5 commit (74ff06f + de94bac + 0a90782 + 2b00fe0 + 22306db). branch HEAD `22306db`. |

### Commit 4 — codex P2 fix (3 layer 동시)

**Codex review id 3378458885**: P2 — production router 에 VOC list route 미등록. `NewRouter` 가 `Handler{...}` literal 에서 `voc` field 를 init 하지 않아 `line 527` 의 `if handler.voc != nil` 체크가 production 에서 항상 `false`. PR #514 의 5 voc route 가 production 에서 mount 안 됨.

**3 layer 동시 fix (`2b00fe0`)**:
1. `router.go` `NewRouter` `Handler{...}` literal: `voc: devreqview.NewVocHandler(...)` init 추가.
2. `voc_handler.go` `RegisterVocRoutes`: `GET /dev-requests/:external_ref` → `GET /dev-requests/external/:external_ref` (gin strict duplicate path error 회피, 사용자 선택) + `POST` param name `external_ref` → `dev_request_id` 정합.
3. `permissions.go` `routePermissionTable`: 5 voc route 의 entry 추가 (PR #514 의 latent 회귀):
   - `POST /dev-requests/:dev_request_id` → `ResourceDevRequests, ActionCreate` (system_admin, 외부 intake)
   - `POST /dev-requests/:dev_request_id/route` → `ResourceDevRequests, ActionEdit` (team_manager/system_admin)
   - `GET /dev-requests/external/:external_ref` → `ResourceDevRequests, ActionView`
   - `GET /vocs` → `ResourceDevRequests, ActionView` (ADR-0028 §6 carve d, system_admin 도구)
   - `GET /me/notifications` → `Bypass` (자기 정보)
   - `POST /me/notifications/:id/read` → `Bypass` (자기 마킹)

**검증**: `go test ./...` 0 FAIL, `go build ./...` silent PASS, codex reply id 3378526106.

### 신규 ID 4건
- `IMPL-voc-01`: voc 등록 / routing / 조회
- `IMPL-notification-01`: in-app notification (`/api/v1/me/notifications`)
- `IMPL-dreq-02`: dev-request 9 field + 단일 트랜잭션 자동 생성
- `IMPL-voc-list-01`: `GET /api/v1/vocs` (ADR-0028 §6 carve d)

### 신규 TC 7건
- TC-VOC-LIST-01..03 (4 케이스)
- TC-RBAC-LOGOUT-01 + TC-RBAC-ROLE-DRIFT-01 + TC-RBAC-LEGACY-01

### PR #514 / #515 정합 후 v1.0 출시 직전 잔여
- **N-6**: staging 1주 운영 + 외부 사용자 ≥5 로그인 검증 (사용자 결정 영역)
- 옵션 D (`project.inbound_source` 자동 routing, ADR-0028 §6 carve a): post-MVP 후속 sprint 후보

## 0a. 이전 세션 (2026-06-09, swagger UI 1차 bootstrap + v1.0 직전 housekeeping)

### PR 머지 결과 (squash, 후속 housekeeping PR은 v1.0 finalizing sprint 본 §0 참조)

| PR | 상태 | 의의 |
| --- | --- | --- |
| **#505** (swagger UI 1차 bootstrap) | ✅ MERGED (squash) | 정적 HTML + CDN (swagger-ui-dist@5.17.14 unpkg) + `embed.FS` 결정. 0 Go 의존성. `DEVHUB_SWAGGER_ENABLED=true` opt-in (default false, prod-safe). 4 commit (`1b640d4` + `428c3f4` + `b48794d` + `6cc208a`) — codex P1 2건 fix + nginx `/devhub/swagger/` forward (단일 포트 정합) + e2e shard 3/3 hotfix. 신규 ID: `IMPL-swagger-01` / `ADR-0027`. main `ad8d481`. |
| **#500** (워커 분업 전면 취소, 06-09) | ✅ MERGED (squash) | `worker_division.md` §0 + §1~§4 historical 標記 + `AGENTS.md` 워커 일반 메모 갱신 + branch prefix 자유화. 사용자 결정 (Claude/Codex 자유 이용 불가). main `f99fef7`. |
| **#499** (N-11 메모리 sync) | ✅ MERGED (squash, rebase 후) | 메모리 4종 (state/handoff/work_backlog/release_v1_roadmap) + traceability report.md §3.5/§6. main `da7d57e`. |
| **#498** (ci.yml 코멘트 갱신) | ✅ CLOSED | e2e shard 2/3+3/3 fail → N-8 race 발견 → close + N-8 hotfix 4차 별도 sprint. |
| **#502** (N-8 hotfix 4차 1차: 502→204) | ✅ MERGED (squash) | backend logout handler graceful degradation. main `6654b44`. |
| **#503** (N-8 hotfix 4차 2차: codex P1 + follow-up) | ✅ MERGED (squash) | response header `X-Keycloak-Likely-Down: true` + typed error sentinel `ErrOIDCConfigMissing`/`ErrOIDCNetworkUnreachable`. main `897953c`. |

### N-8 hotfix 4차 정공법 (3 commit, 2 PR)

**근본 layer**: backend `POST /api/v1/auth/logout` 가 Keycloak 도달 실패 시 502 즉시 반환 → frontend logout() 가 OIDC skip + `window.location.assign('/login')` 강제 → AuthGuard pathname 변화 useEffect 에서 stale actor 박음 → `/developer` 진입 → `/login` 도착 못함. PR #497 의 hotfix #1/`#2/`#3 가 모두 backend 502 자체를 막지 못함 (deterministic, 32회 retry).

**PR #502 (1차)**: backend logout handler 가 502 → **204 No Content** + audit `revoke_status=unreachable` + hotfix 식별자. frontend logout() 가 정상 204 분기 진입 → OIDC end_session_endpoint 호출 → /login 정상 도착 → race close.

**PR #503 (2차 commit 066cd7b, codex P1 응답)**: "구분 가능한 응답" — 204 + response header `X-Keycloak-Likely-Down: true` + `X-Logout-Hotfix: N-8-4:graceful-degrade`. frontend 가 header 마커 conditional 확인 → OIDC skip 또는 정상 OIDC 결정. 진짜 IdP outage 시 dead IdP trap 회피.

**PR #503 (3차 commit e18b34f, codex P1 follow-up 응답)**: typed error sentinel 도입.
- `authview.ErrOIDCConfigMissing` (sentinel): backend config 결함 (missing realm/oidc_client_id/oidc_client_secret) → handler 가 **marker 미부착** + 정상 OIDC 분기 + audit `revoke_status=config_error` + `config_error_detail`
- `authview.ErrOIDCNetworkUnreachable` (sentinel): 네트워크/5xx outage (DNS 실패, conn refused, timeout, Keycloak 5xx) → handler 가 marker 부착
- 그 외 미분류 error: conservative — outage 분류

codex P1 의 핵심 우려 "reachable Keycloak SSO session is not terminated" 정공법: config error 분기에서 marker 미부착 → frontend 정상 OIDC → RP-initiated logout 시도 → SSO session 정상 종료.

### 검증

- **CI 모두 SUCCESS** (PR #503 머지 시점): workflow-lint / changed-paths / migration-prefix / Backend Unit + Integration / Frontend Unit / E2E Build / **E2E shard 1/2/3 모두 PASS**
- `go test ./...` (35 packages) PASS
- `npx vitest run` (80 files, **1033 tests**) PASS
- **신규 test 4건**:
  - TC-AUTH-LOGOUT-04 (network/5xx → 204 + marker)
  - **TC-AUTH-LOGOUT-08** (config error → 204 + marker 미부착)
  - TC-AUTH-LOGOUT-FE-07 (frontend header 마커 확인 → OIDC skip)
  - **TC-AUTH-LOGOUT-FE-08** (frontend header 없음 → 정상 OIDC)

### 잔여 DoD 해소

- **N-11 잔여 DoD** (main 첫 PR 두 job PASS, issue #419): PR #503 머지 시점에 e2e shard 1..3 모두 PASS → 해소
- **N-8 race** (issue #501): close
- **워커 분업 전면 취소** (사용자 결정): PR #500 머지로 정합. branch prefix 자유 (`maintenance/`, `chore/`, `docs/`, `fix/`, `feat/` 등)

## 1. 다음 세션 directive

### v1.0 출시 직전 — 우선순위

1. **N-6 (v1.0 staging 1주 운영)** — N-8 + N-11 + N-7 + 워커 분업 취소 + swagger UI + housekeeping 정합 완료. 사용자가 staging 환경 운영 + 외부 사용자 ≥5 로그인 검증. (사용자 결정, sprint 영역 외)
2. **N-10 Manager RBAC E2E spec-vs-구현 갭 6 TC 보강** — v1.0 출시 전 가능. **sprint `maintenance/work_260610-c-N10-rbac-e2e-tcs` 진입 예정**. validation 보고서 [docs/validation/N-10-manager-rbac.md](docs/validation/N-10-manager-rbac.md) 의 TC-RBAC-ROW-READ-01/02 + TC-RBAC-LOGOUT-01/02 + TC-RBAC-ROLE-DRIFT-01 + TC-RBAC-CODE-01 + TC-RBAC-TRACE-01 (총 6건).

### 완료 정합 (2026-06-10)

- release_v1_roadmap.md §3.5 N-8 race status `✅ resolved` + N-11 row "✅ 잔여 DoD 해소" + §4.1 sprint -k 의 N-11 잔여 DoD 완료 마킹 + §3.5 N-10 follow-up 6 TC 명시.
- docs/traceability/report.md §4 ADR-0025 (envelope encryption) + ADR-0026 (Keycloak role excluded) row 보강.
- state.json 1515 → 150 line (90% 감소) — 21 top-level key + next_actions 5 closed key archive.
- work_backlog.md §5 5월 135행 archive summary 1 line 으로 축약.
- ADR-0027 (swagger UI 결정) + IMPL-swagger-01.

### 자유 에이전트 정책 (2026-06-09 결정)

본 세션 결정으로 **누구든** 어느 sprint/영역 진입 가능. `worker_division.md` §0 + `AGENTS.md` "워커 일반 메모 (2026-06-09 전면 갱신)" 정합. branch prefix 자유.

### 자유 에이전트 정책 (2026-06-09 결정)

본 세션 결정으로 **누구든** 어느 sprint/영역 진입 가능. `worker_division.md` §0 + `AGENTS.md` "워커 일반 메모 (2026-06-09 전면 갱신)" 정합. branch prefix 자유.

## 2. 이전 핵심 스프린트 (5월/4월 historical, archive)

5월/4월 historical 5개 sprint (X-3 envelope encryption / NOW-3 SCM E2E / NOW-4 frontend unit test / NOW-5 migration prefix guard / 2026-06-01 CI 복구 / 2026-06-06 sprint -h 추적성 ID) 의 상세 본문은 [_archive/state-2026-05-pr-tracker/session_handoff_2026-05-06_archive.md](./_archive/state-2026-05-pr-tracker/session_handoff_2026-05-06_archive.md) 로 이동. canonical source 는 `git log --since=2026-04-01 --until=2026-06-08 --merges` (PR #20~#504, 470+ commit).

### 1 line summary (5월 ~ 4월)

- **X-3** (envelope encryption + KEK key management, 2026-05-27) ✅ — `internal/crypt` AES-GCM-256 봉투 + `IntegrationRepository` 자동 Encrypt/Decrypt 결합 + 6 envelope_test.go unit test PASS + 32 packages green.
- **NOW-3** (SCM import/create + draft/publish E2E) ✅ — backend 캐스팅 정정 + Gitea Mock Provider Fallback + nano-ts unique ID 매핑 + Playwright auto-wait Locator → 63 passed / 6 skipped.
- **NOW-4** (frontend unit test 보강) ✅ — Zustand store / ProviderModal / MemberTable / PermissionEditor 4 module Vitest 작성 → 962 unit test 100% PASS.
- **NOW-5** (migration prefix uniqueness CI guard) ✅ — `scripts/check-migration-uniqueness.sh` + `ci.yml` 상시 lint + `make lint-migrations` 바인딩.
- **2026-06-01 CI 복구** ✅ — `applications/[id]/page.tsx` 중복 import 제거 + `admin-projects.spec.ts` 환경 독립 검증 → CI run `26738464130` 성공.
- **2026-06-06 sprint -h 추적성 ID 발급** ✅ — PR #490, N-7 (REQ-FR-106/ARCH-18/API-98 + IMPL/UT/TC) + N-8 (REQ-FR-107/ARCH-19/API-99) + N-9 (REQ-FR-108/ARCH-20/API-100) + `integration-registry` 도메인 cross-ref 보강.

## 3. 후속 carve out / 잔여 백로그 우선순위 (current)

| 우선순위 | 항목 | 사유 |
|---|---|---|
| **N-6** | v1.0 staging 1주 운영 검증 | 외부 사용자 로그인 + Onboarding SOP DoD 8 만족 (사용자) |
| **N-10** | Manager RBAC E2E spec-vs-구현 갭 6 TC 보강 | sprint `maintenance/work_260610-c-N10-rbac-e2e-tcs` 진입 예정 |
| **X-1** | System Admin 운영 대시보드 | Gitea sync job 큐/상태 + provider health (v1.1) |
| **X-2** | inbound webhook 정규화 깊이 | multi-provider sync 일반화 (v1.1) |

## 4. 다음 세션 directive
* **PR #515 ✅ MERGED** (squash `f7d2705`) + **PR #516 ✅ MERGED** (squash `2b3c766`) + **PR #517 ✅ MERGED** (squash `97bc6bc`).
* **swagger UI 정상동작 fix 완료** — PR #508 의 silent 404 의도적 결정을 embed fallback 으로 supersede. 7 swagger TC 모두 PASS. staging env `DEVHUB_SWAGGER_ENABLED=true` 만 설정해도 openapi.yaml 정상 서빙.
* **N-6**: staging 1주 운영 (사용자 결정 영역). swagger UI 정상동작 정합.
* **N-10 IT 3 TC 완료 정합** (본 sprint): `TC-RBAC-LOGOUT-01` + `TC-RBAC-ROLE-DRIFT-01` + `TC-RBAC-LEGACY-01` ✅ verified.
* **option D 검토 완료**: N-13 + ADR-0028 §6 정합. 구현 = v1.1 milestone 진입 시점.
* **V1.1 진입 준비**: X-1/X-2 로드맵 백로그 분석.

---

## 5. 본 세션 (2026-06-10, v1.1 sprint -a follow-up — PR #539 머지)

### PR 머지 결과

| PR | 상태 | 의의 |
| --- | --- | --- |
| **#538** (sprint -a 본 — port interface, 이전 머지) | ✅ MERGED | `domain/auth-session/integration/ports.go` 신규 (4 port + 4 type alias + 3 sentinel error alias). ADR-0030. main `20b4bb3b`. |
| **#539** (sprint -a follow-up — stub + main wiring + view/ deprecation) | ✅ MERGED (squash, branch delete) | saovae_stub (4 port + webhook handler) + main.go `DEVHUB_BUILD_TIER` env var 분기 + ports.go mirror alias 통합 + view/ 3 interface deprecation. main `87e6c1f5`. CI 7/7 PASS. |

### PR #539 5 file (commit `a00793bc`, +238 -97)

1. `backend-core/internal/sso-integrations/keycloak/saovae_stub.go` (NEW, 105 lines) — 4 port stub + webhook handler. `go build ./...` PASS. no test files (의도적 — 사외 stub, keycloak infra 비의존).
2. `backend-core/main.go` (L148-235 분기 + 2 import 추가) — `var (idpAdmin httpapi.IdentityAdmin; oidcLogout httpapi.OIDCLogoutClient; keycloakEventPort integration.KeycloakEventPort)` + `DEVHUB_BUILD_TIER` env var. default (사외) = saovae_stub, `=internal` 시 real KeycloakAdminClient.
3. `backend-core/internal/domain/auth-session/integration/ports.go` — `KeycloakUserEvent` / `KeycloakAdminEvent` mirror struct → `type X = httpapi.X` alias. `*KeycloakAdminClient` 가 `KeycloakEventPort` 충족 위해 필수.
4. `backend-core/internal/domain/auth-session/view/auth.go:59` — `BearerTokenVerifier` interface deprecation comment.
5. `backend-core/internal/domain/auth-session/view/handler.go:27 + :197` — `IdentityAdmin` + `OIDCLogoutClient` interface deprecation comment.

### CI 7/7 PASS

| Check | Result | Duration |
|---|---|---|
| Backend Integration Tests | pass | 1m15s |
| Backend Unit Tests | pass | 1m8s |
| E2E Build Artifacts | pass | 1m49s |
| E2E Tests (Playwright, shard 1/3) | pass | 4m0s |
| E2E Tests (Playwright, shard 2/3) | pass | 4m2s |
| E2E Tests (Playwright, shard 3/3) | pass | 5m30s |
| Detect Changed Paths | pass | 11s |
| Migration Prefix Uniqueness | pass | 7s |
| OpenAPI YAML Lint | pass | 9s |
| Workflow Lint (actionlint) | pass | 13s |
| Frontend Unit Tests | skip (path-detect 결과 — backend-only 변경) | - |

### Tier 분류 (PR #539 self-check)

- **ports.go / view/ = 공용** (interface 만 노출)
- **sso-integrations/keycloak/ = 사외** (saovae_stub — Keycloak 인프라 비의존)
- **main.go 변경 = 공용** (runtime injection branch 만, 사내 한정 패턴 미도입 — `check-tier-separation.sh no changes` 확인)

### 사용자 결정 사항 (in-session)

- **PR scope 분리**: sprint -a follow-up 본 PR = stub + main wiring (Recommended). real adapter 별도 PR.
- **Build 정책**: `//go:build` tag 미사용, **runtime injection** (단일 binary, `DEVHUB_BUILD_TIER` env var).
- **Default = 사외 (saovae_stub)**: env var 미설정 시 saovae_stub 자동 사용.
- **`=internal` 시 real KeycloakAdminClient**: 사내 staging/prod-smoke 검증.
- **Type alias**: ports.go 의 mirror struct 를 httpapi alias 로 통합 (distinct type → alias).

### 후속 carry-over (다음 PR)

1. **C-a** real adapter: `sso-integrations/keycloak/{verifier,admin_client}.go` (P0)
2. **C-b** main.go event listener type assertion 정리 (P0)
3. **C-c** `_ = keycloakEventPort` placeholder 제거 (P0, C-b 완료 시)
4. **C-d** v1.0 mirror struct 제거: `httpapi.KeycloakUserEvent` / `httpapi.KeycloakAdminEvent` (P1)
5. **C-e** audit-ops 의 mirror 와 통합 (P1)
6. **C-f** `infra/idp/_archive_2026-06-10/` immutable archive (P1)
7. **C-g** traceability report.md IMPL-30/31/32 row 갱신 (P2)
8. **C-h** ADR-0030 §5 timeline 갱신 (P2)
9. **C-i** E2E test saovae_stub + real adapter CI matrix (P2)
10. **C-j** build tag 정책 재검토 (P3)

### Memory 갱신

- `ai-workflow/memory/feat/work_260610-v1-1-sprint-a-followup/` 신규 작성 (state.json + session_handoff.md + work_backlog.md + backlog/2026-06-10.md).
- main `state.json` status 갱신 (PR #539 entry).
- 본 `session_handoff.md` §5 append.

### 다음 세션 directive (sprint -a follow-up 완료 후)

- **real adapter PR 시작**: branch `feat/work_260610-v1-1-sprint-a-real-adapter` 분기. C-a → C-b → C-c → C-d → C-e → C-f 순서로 진행.
- **C-g traceability 갱신**: `docs/traceability/report.md` IMPL-30/31/32 row 추가.
- **또는 다른 sprint 진입**: PR #538 이전의 carry-over, ADR-0030 §5 timeline 갱신 등.

## 6. 본 세션 (2026-06-10, sprint -a follow-up PR1 (PR #540) 의 carry-over C-g + C-h 정공법 PR — docs only)

### PR 결과

| PR | 상태 | 의의 |
| --- | --- | --- |
| **#540** (sprint -a follow-up PR1 — real adapter + v1.0 mirror struct 제거) | ✅ MERGED (squash, branch delete) | `sso-integrations/keycloak/{verifier,admin_client,metrics}.go` 신규 (real KeycloakJWKSVerifier + KeycloakAdminClient 3 port 동시 충족 + raw wire → flat canonical struct 매핑) + saovae_stub 보강 (PR #539 머지 후) + v1.0 mirror struct 제거 (`httpapi.KeycloakUserEvent` + `KeycloakAdminEvent` 폐기) + audit-ops mirror 통합 (`KeycloakEventLister` interface 통폐합) + `infra/idp/_archive_2026-06-10/identity.schema.json` archive + `infra/idp/README.md` 갱신. main `58d163f`. CI 7/7 PASS. |
| **본 PR (carry-over C-g + C-h 정합 PR)** | ✅ MERGED (squash, PR #541, main `88681f4`, branch delete) | `docs/work_260610-traceability-impl-sso-keycloak` 분기 (docs only, 코드 0줄). §2.4 IMPL 개요 paragraph 갱신 + 5 row IMPL 신규 sub-table (`sso-keycloak-01` + `sso-keycloak-stub-01` + `sso-keycloak-metrics-01` + `auth-session-port-01` + `audit-ops-event-mirror-01`, conventions.md §1 kebab-case 정합) + §3.1 auth-session/audit-ops + §3.3 keycloak-idp 매트릭스 row 갱신 + §4 ADR 인덱스 ADR-0030 row + §6 변경 이력 row + ADR-0030 §5 timeline accepted/done + §9 변경 이력 row. 신규 ID 5건 (모두 IMPL, REQ/UC/ARCH/API/RM/UT/TC 신규 발급 0건). CI 4/4 PASS (docs only PR — backend/e2e/frontend skip, 4 path-detect + lint). commit `22e8c84` → squash merge `88681f4`. Tier: **공용** (문서만). |

### C-g / C-h 정합 PR scope

sprint -a follow-up PR1 (PR #540) 의 carry-over C-g + C-h 의 정공법. 본 PR 은 **문서만 변경** (코드 0줄). 5 row ID 의 정공법 = `conventions.md §1` 의 kebab-case module ID 정합 (메모리 출발점의 `IMPL-30/31/32` 표기는 형식 위반 → 정정).

### Tier 분류

본 PR 의 모든 변경 = **공용** (사내 한정 정보 미포함, `docs/traceability/report.md` + `docs/adr/0030-...` 의 문서 정합). `check-tier-separation.sh` PASS 예상.

### 후속 carry-over (C-i + C-j)

- **C-i (P2)**: E2E saovae_stub + real adapter CI matrix — DEVHUB_BUILD_TIER=internal env var + e2e shard 양쪽 정합. 본 PR 후속 별도 PR.
- **C-j (P3)**: build tag 정책 재검토 — runtime injection (현재) ↔ build tag 전환 trade-off. 별도 ADR 후보.

### Memory 갱신

- `ai-workflow/memory/docs/work_260610-traceability-impl-sso-keycloak/` 신규 (state.json + session_handoff.md + work_backlog.md + backlog/2026-06-10.md).
- main flat `state.json` status + head_commit 갱신 (PR #540 baseline = `58d163f`, 본 PR 머지 시점은 head_commit main = PR 본 PR 머지 commit).
- 본 `session_handoff.md` §6 append.

### 다음 세션 directive

- 본 PR commit + push + PR 발행 (사용자 confirm 후).
- 또는 C-i (E2E CI matrix) 진입.
- 또는 다른 sprint (N-10 RBAC E2E 6 TC / release_v1_roadmap.md 갱신).

## 7. 본 세션 (2026-06-10, sprint -a follow-up PR1 PR #540 의 carry-over C-i 정공법 PR — ci.yml + script 2 file)

### PR 결과

| PR | 상태 | 의의 |
| --- | --- | --- |
| **#542** (C-i E2E Internal job) | ✅ MERGED (squash, branch delete) | `feat/work_260610-c-i-e2e-internal-job` 분기. `.github/workflows/ci.yml` 에 `e2e-internal` job 신규 (+202 lines, 23 step, PG 15 + Keycloak container port 8181 + apply migrations + validate E2E-CI sync contract + Start Backend DEVHUB_BUILD_TIER=internal + Start Frontend + Wait + Run E2E Tests shard 1/1 + Upload Report + Upload Logs). `scripts/ci-e2e-sync-check.sh` 에 DEVHUB_BUILD_TIER 의도적 미포함 rationale comment (+5 lines). e2e shard 1/2/3 (saovae_stub default) 의 env block, start command, test invocation 모두 변경 0. main `24674b8`. CI 4/4 PASS (workflow 변경만, backend/e2e/frontend skip). Tier: **공용** (`.github/workflows/*` + `scripts/*` 모두 사내 한정 정보 미포함). |

### scope 결정 (옵션 A 채택)

e2e shard 1/2/3 (saovae_stub default) + 별도 `e2e-internal` job 1개 (`DEVHUB_BUILD_TIER=internal`) 의 CI matrix 1쌍. 옵션 B (unit test cover only) / C (6 matrix) / D (matrix shard 1/2/3 × 2) 모두 거부.

### trade-off

- **Keycloak container port 8181** (e2e shard 의 8180 과 분리) — e2e shard 와 e2e-internal 동시 trigger 가능
- **Playwright shard 1/1** (단일 shard, ≈ 4-5min) — logout flow 외 다른 e2e suite (auth, RBAC, CRUD) 는 backend 의 build tier 무관
- **DEVHUB_BUILD_TIER token** required_e2e_tokens 에 의도적 미포함 — e2e shard 1/2/3 의 saovae_stub default env block 미설정 유지. e2e-internal 의 DEVHUB_BUILD_TIER env 정합은 actionlint + 실제 e2e run 이 검증

### Memory 갱신

- `ai-workflow/memory/feat/work_260610-c-i-e2e-internal-job/` 신규 (state.json + session_handoff.md + work_backlog.md + backlog/2026-06-10.md + pr_body.md).
- main flat `state.json` head_commit = `24674b8` (PR #542 머지 baseline).
- 본 `session_handoff.md` §7 append.

### 다음 세션 directive

- **C-j (P3)**: build tag 정책 재검토 (runtime injection ↔ build tag 전환 trade-off). 별도 ADR 후보.
- **backend-integration DEVHUB_BUILD_TIER matrix** (sprint -a follow-up 본 PR #539 의 backend-integration 재활성화 + DEVHUB_BUILD_TIER=internal matrix).
- **release_v1_roadmap.md §3.5 N-13** 정합 (C-i done 마킹).

## 8. 본 세션 (2026-06-10, sprint -a follow-up PR1 PR #540 의 carry-over C-j 정공법 PR — docs/adr/ + docs/traceability/ + ai-workflow/memory/)

### PR 결과

| PR | 상태 | 의의 |
| --- | --- | --- |
| **#543** (C-j build tag 정책 재검토 PR) | ✅ MERGED (squash, branch delete) | `docs/work_260610-c-j-build-tag-review` 분기. **`docs/adr/0031-build-tag-policy-review.md` 신규 (12KB, 9 section)** — ADR-0030 §2.3 runtime injection 결정을 **정량 측정 후 confirmed** (supersede X). **결정**: 옵션 2 (런타임 injection 유지). 근거 = stub binary overhead < 5KB (전체 backend-core < 50MB 대비 0.01%) vs build tag 전환 시 CI matrix 2배 (+30~60min) + 5~10 file `//go:build` tag + 2개 binary 운영. **재검토 trigger 5건** (§5): stub code size > 250KB / stub production risk / CI axes 5+ / Phase 2 agentic RAG / stub safety — 현시점 trigger 0건. ADR-0030 §2.3 row confirmed reference 추가. `docs/traceability/report.md` §4 ADR-0031 row + §6 row 신규. main `d3488ca`. CI 4/4 PASS (docs only PR, backend/e2e/frontend skip). Tier: **공용**. |

### scope 결정

코드 0줄 변경. ADR + traceability + memory 4 file. C-j 의 정공법 = **ADR-0031 신규 + ADR-0030 §2.3 confirmed (supersede X) + 9 section 정공법** (배경 + 정량 측정 + 옵션 + 결정 + 재검토 trigger + cross-tier + risks + supersession + 변경 이력).

### trade-off (현시점 정량 측정)

| 측정 항목 | Runtime injection (현재) | Build tag (이론) | 차이 |
| --- | --- | --- | --- |
| Binary overhead | < 5KB | -6.3KB (절감) | -6.3KB (build tag 유리) |
| CI runtime | +15~20min (PR #542) | +30~60min (이론) | build tag 가 +15~40min 더 |
| CI matrix jobs | 4 (e2e 1/2/3 + e2e-internal) | 6 (e2e × 2 tags) | build tag 가 +2 jobs |
| 코드 변경 | 0 (현재 상태 유지) | 5~10 file | build tag 가 +5~10 file |
| 운영 복잡도 | 1 binary | 2 binary | build tag 가 +1 |

**결론**: build tag 의 binary size 절감 (~6KB) 은 무시 가능 수준. runtime injection 의 cost 가 build tag 의 cost 보다 본질적으로 작음. 1:5+ cost ratio.

### Memory 갱신

- `ai-workflow/memory/docs/work_260610-c-j-build-tag-review/` 신규 (state.json + session_handoff.md + work_backlog.md + backlog/2026-06-10.md + pr_body.md).
- main flat `state.json` head_commit = `d3488ca` (PR #543 머지 baseline).
- 본 `session_handoff.md` §8 append.

### 다음 세션 directive

- **backend-integration DEVHUB_BUILD_TIER matrix** (P3): sprint -a follow-up 본 PR #539 의 backend-integration 재활성화 + DEVHUB_BUILD_TIER=internal matrix.
- **release_v1_roadmap.md §3.5 N-13** 정합 (P3): C-i + C-j + C-g/C-h + C-j done 마킹. N-13 row close.
- **N-10 RBAC E2E 6 TC 보강** (sprint `maintenance/work_260610-c-N10-rbac-e2e-tcs`): v1.0 출시 직전 잔여.
- 또는 다른 sprint 진입.

## 9. 본 세션 (2026-06-10, D-72 Phase 1 — `~/wiki/` LLM Wiki 통합 의 in-repo source-of-truth + sync script)

### PR 결과

| PR | 상태 | 의의 |
| --- | --- | --- |
| **#544** (D-72 Phase 1 LLM Wiki 통합) | ✅ MERGED (squash, branch delete) | `feat/work_260610-d-72-wiki-phase-1` 분기. **`docs/llm-wiki/` 5 file 신규** (README 7.8KB + scope-and-rationale 10.6KB + mirror-list 10KB + lint-config.toml 4.4KB + operation-sop 10.7KB = 43.5KB) + **`scripts/wiki-sync-devhub.sh` 6.4KB executable** (BSD-rsync safe, 7 source 패턴, 82 file, --dry-run + vault 부재 no-op). **D-72 응답 §2 Q1~Q6 전체 적용**: 단일 vault + per-project 동거 + Q3 단순화 (lint L11 + sa-internal/ 격리 불요) + Q4 L01~L10 + L07 ADR 면제 + Q5 v1.5 동시 시작 + Q6 단일 AGENTS.md + per-project lint report. `docs/wiki/` (Public, GitHub Wiki 게시 source) vs `docs/llm-wiki/` (LLM Wiki SSOT) 의 분리. main `a96f586`. CI 4/4 PASS (docs only PR, backend/e2e/frontend skip). Tier: **공용**. |

### scope 결정

**코드 0줄 변경** (스크립트 6.4KB + docs 5 file 신규). mirror list = **core subset ~82 file** (ADR 31 + Governance 5 + Planning 26 + Setup 15 + Requirements 1 + OpenAPI 1 + AI-workflow memory 3). domain (66) + architecture + infrastructure + validation (~100 file) 은 Phase 3 (mass ingest) 의 별도 PR. **mirror 실행은 본 PR scope 외** (`~/wiki/raw/projects/devhub/` 의 out-of-repo 변경) — 사용자 confirm 후 T-d-72-5.

### trade-off

- **`docs/llm-wiki/` 선택 (vs `docs/wiki/` 또는 `docs/wiki-integration/`)**: 기존 `docs/wiki/` = **Public Wiki** (GitHub Wiki 게시 source, 인간 큐레이션, mtime 2026-05-20). 본 Phase 1 의 **LLM Wiki SSOT** 와 audience 다름. 디렉터리 이름 분리 = 두 wiki 의 명확한 구분. `docs/wiki/` (Public) ↔ `docs/llm-wiki/` (LLM) 의 cross-link 없음.
- **mirror list 의 scope = core subset ~82 file**: D-72 응답 §4 #3 의 "100~200 파일" 의 1/2. domain (66) + architecture + infrastructure + validation (~100 file) 은 **Phase 3 (mass ingest)** 에서 별도 PR. 본 PR 의 검증 가능한 정공법 (CI 4/4 + script smoke test) = 작은 core subset.
- **lint-config.toml 의 L07 ADR 면제 config 작성 (옵션 미사용)**: wiki-lint 의 `--project` + `--project-config` 옵션은 my_harness 측 D-73 의 작업. 본 PR 은 config 의 source 만 제공. 옵션 추가 후 자동 활성.

### Memory 갱신

- `ai-workflow/memory/feat/work_260610-d-72-wiki-phase-1/` 신규 (state.json + session_handoff.md + work_backlog.md + backlog/2026-06-10.md + pr_body.md).
- main flat `state.json` head_commit = `a96f586` (PR #544 머지 baseline).
- 본 `session_handoff.md` §9 append.

### 다음 세션 directive

- **T-d-72-5** (P3, 사용자 trigger): `bash scripts/wiki-sync-devhub.sh` 1회 실행 (real, dry-run 아닌) → `~/wiki/raw/projects/devhub/` 에 ~82 file mirror + `_manifest.md` 자동 생성. **본 PR 머지 후 사용자 confirm 시점**.
- **D-73** (P3, my_harness 측): wiki-lint skill 에 `--project` + `--project-config` 옵션 추가. 본 저장소 의 lint-config.toml 활성.
- **D-74** (P3, my_harness 측): my_harness 의 `_lint/my-harness/` + 본 저장소 의 `_lint/devhub/` 디렉터리 셋업.
- **Phase 3** (P3, mass ingest, 별도 PR): domain (66) + architecture + infrastructure + validation (~100 file) mirror + 30~50 wiki page.
- **wiki/cross/** (P3, Phase 3 후속): cross-project 종합 (my_harness 의 LLM Wiki 패턴 ↔ DevHub 의 ADR-0030 runtime injection).
- **v2.0** (P3, forward): LLM 호출 + BM25+vector+MCP. my_harness 의 v2.0 경험 보고 진입.
- **N-13 release_v1_roadmap §3.5 정합** (P3, housekeeping): N-13 row status = done 마킹.
- 또는 다른 sprint (backend-integration matrix / N-10 RBAC E2E 6 TC).

## 10. wiki 통합 일임 결정 (2026-06-10, yklee directive)

### 결정

**yklee 2026-06-10 directive**: wiki 통합 작업은 my_harness 측 에이전트에 일임. 본 저장소 (DevHub) 의 sprint 는 **my_harness 의 결과 통보 대기**. **동시 진행 시 꼬일 가능성 회피** (mirror 실행 / lint config 활성화 / mass ingest / wiki page 작성 / cross-project 종합 등이 양 project 에서 동시 진행 시 race condition + 정책 drift 위험).

### 본 저장소 측 follow-up

- **본 PR #544 머지로 Phase 1 의 in-repo source-of-truth 정합 완료** (docs/llm-wiki/ 5 file + scripts/wiki-sync-devhub.sh). 변경 불요.
- **본 저장소 측 mirror 실행 (T-d-72-5) 도 대기**: `bash scripts/wiki-sync-devhub.sh` 1회 실행 (real, dry-run 아닌) → `~/wiki/raw/projects/devhub/` 에 ~82 file mirror + `_manifest.md` 자동 생성. **사용자 (yklee) 의 별도 confirm 후 진행**. my_harness 의 작업 결과 통보 후 일괄 mirror 가 정합.
- **본 저장소 측의 follow-up task (carry-over, my_harness 통보 대기)**:
  - my_harness 의 D-73 (wiki-lint `--project` + `--project-config` 옵션 추가) — 본 저장소 의 lint-config.toml 자동 활성
  - my_harness 의 D-74 (`_lint/devhub/` 셋업) — 본 저장소 의 per-project lint report 정합
  - my_harness 의 Phase 3 (mass ingest, ~100 file mirror + 30~50 wiki page) — 본 저장소 의 domain + architecture + infrastructure + validation 영역
  - my_harness 의 wiki/cross/ (cross-project 종합) — my_harness 의 LLM Wiki 패턴 ↔ DevHub 의 ADR-0030 runtime injection
  - my_harness 의 v2.0 (full compile, LLM 호출 + BM25+vector+MCP) — my_harness 의 v2.0 경험 보고 진입
- **본 저장소 측 follow-up task (독립 진행 가능, yklee 별도 confirm 시)**:
  - **N-13 release_v1_roadmap.md §3.5 정합** (P3, housekeeping): D-72 + D-73 + D-74 + D-75 의 carry-over N-13 row status = done 마킹. **본 저장소 측에서 독립 진행 가능** (my_harness 결과 통보와 무관).
  - **backend-integration DEVHUB_BUILD_TIER matrix** (P3): sprint -a follow-up 본 PR #539 의 backend-integration 재활성화 + DEVHUB_BUILD_TIER=internal matrix. **본 저장소 측에서 독립 진행 가능**.
  - **N-10 RBAC E2E 6 TC 보강** (P1, sprint `maintenance/work_260610-c-N10-rbac-e2e-tcs`): v1.0 출시 직전 잔여. **본 저장소 측에서 독립 진행 가능**.

### Memory 갱신

- 본 §10 append. 다음 세션 진입 시 본 §10 의 결정 참조.

### 다음 세션 directive

- **다른 sprint 진입 (본 저장소 측의 독립 진행 가능 task)**:
  - N-13 release_v1_roadmap.md §3.5 정합 (housekeeping)
  - backend-integration DEVHUB_BUILD_TIER matrix (P3)
  - N-10 RBAC E2E 6 TC 보강 (P1)
- **또는 사용자 confirm 후 본 저장소 측의 mirror 실행 (T-d-72-5)** — my_harness 결과 통보와 무관하게 단독 진행 가능 (mirror list 가 본 저장소 의 SSOT 이므로).
- **또는 my_harness 의 결과 통보 대기** (D-73, D-74, Phase 3 등).
- 또는 N-10 RBAC E2E 6 TC 보강 / release_v1_roadmap.md 갱신.
