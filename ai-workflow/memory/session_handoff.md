# Session Handoff — main (2026-06-12, v1.0 출시 직전 finalizing — PR #514 + PR #515 4 commit + codex P2 fix)

- 문서 목적: PR #514 (voc + notification, ADR-0028) 머지 + 본 sprint `maintenance/work_260612-b-v1-finalizing` PR #515 4 commit push 상태 인계.
- 범위: 본 세션의 2 PR (PR #514 squash, PR #515 push). 옵션 A (N-12 housekeeping) + B (voc list API) + C (N-10 backend IT 3 TC) + codex P2 fix (3 layer: production router mount + routePermissionTable + gin path conflict).
- 상태: branch `maintenance/work_260612-b-v1-finalizing` HEAD `22306db` (5 commit push), PR #515 CI 진행 중 (Backend Unit Tests ✅ + 3 IN_PROGRESS).
- 최종 수정일: 2026-06-12

## 0. 본 세션 핵심 결과 (2026-06-12, v1.0 출시 직전 finalizing)

### PR 머지 / Push 결과

| PR | 상태 | 의의 |
| --- | --- | --- |
| **#514** (voc + notification, ADR-0028) | ✅ MERGED (squash) | 외부 시스템 의뢰 staging 도메인 + 9 field + in-app notification + 5 API. `(source_system, external_ref)` UNIQUE for idempotency. ADR-0028 §3 옵션 1 (별도 도메인 + 1:1 dev-request 매핑) 채택. 12 file +1043 line. main `ba7823f`. |
| **#515** (v1.0 출시 직전 finalizing) | ⏳ PUSH (CI 진행 중) | 옵션 A (N-12 housekeeping) + B (voc list API) + C (N-10 backend IT 3 TC) + codex P2 fix (PR #514 latent 회귀 3 layer 동시 fix). 5 commit (74ff06f + de94bac + 0a90782 + 2b00fe0 + 22306db). branch HEAD `22306db`. |

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

## 0a. 이전 세션 (2026-06-10, swagger UI 1차 bootstrap + housekeeping)

### PR 머지 결과 (squash, 본 housekeeping PR 추가 예정)

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
* **PR #515 CI 완료 대기** → 머지 → main HEAD `22306db` 머지.
* **N-6**: staging 1주 운영 (사용자 결정 영역).
* **N-10 IT 3 TC 완료 정합** (본 sprint): `TC-RBAC-LOGOUT-01` + `TC-RBAC-ROLE-DRIFT-01` + `TC-RBAC-LEGACY-01` ✅ verified.
* **옵션 D 검토** (`project.inbound_source` 자동 routing, ADR-0028 §6 carve a): post-MVP 후속 sprint 후보.
* **V1.1 진입 준비**: X-1/X-2 로드맵 백로그 분석.
