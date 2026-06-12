# Sprint Plan v2: N-13 follow-up C (구현 follow-up, v1.1 milestone 진입 시점) — PR A-1 + A-2 통합본

- **문서 목적**: PR #548 (CLOSED, 2026-06-11) 의 E2E Internal 1 fail 2건 중 follow-up 3 branch 중 마지막 branch (구현 follow-up) 의 정공법 + scope + ID slot + 의존 + sprint step 정합. 본 sprint = PR A-1 (backend foundation) + PR A-2 (routing + voc_handler 통합 + openapi + e2e) 통합 1 PR.
- **범위**: backend 9 file (PR A-1) + 4 file (PR A-2 신규) + docs 1 file + e2e 1 file = **총 15 file** (코드 ~800~ line 추정, frontend X).
- **대상 독자**: Backend / 프론트엔드 개발자, AI agent, QA, 운영자, product owner.
- **상태**: **draft → active** (sprint `feat/work_260612-4-v1-1-inbound-source-impl` 분기 후 active, 사용자 confirm 후 머지).
- **최종 수정일**: 2026-06-12 (sprint plan v2 정공법 결정, sprint `feat/work_260612-4-v1-1-inbound-source-impl`).
- **결정 근거 sprint**: `feat/work_260612-4-v1-1-inbound-source-impl` (본 sprint, PR A-1 + A-2 통합), `feat/work_260611-a-n13-inbound-source-housekeeping` (직전 sprint, ID slot 9 row 발급, PR #547 MERGED), `feat/work_260611-a-n13-inbound-source-impl` (직전 직전 sprint, PR A-1 의 backend foundation 1차 시도, PR #548 CLOSED).
- **관련 문서**:
  - [ADR-0028 §6 (a)](../adr/0028-dev-requests-voc-external-ref.md) (N-13 source ADR)
  - [release_v1_roadmap.md §3.5 N-13 row + §4.2 v1.1 milestone](../planning/release_v1_roadmap.md) (N-13 정의 + v1.1 정공법)
  - [sprint plan v1 (2026-06-12 PR #548 close 정공법)](./2026-06-12-inbound-source-routing-sprint-plan.md) (정공법 + ID slot + 의존)
  - [verification report v1 (2026-06-12 Test 2 verification)](../validation/2026-06-12-n13-test2-rebase-verification.md)
  - [PR #547](https://github.com/ykylee/Devhub_example/pull/547) (ID slot 9 row 발급, MERGED 2026-06-11)
  - [PR #548](https://github.com/ykylee/Devhub_example/pull/548) (CLOSED, E2E Internal 1 fail 2건)
  - [PR #550](https://github.com/ykylee/Devhub_example/pull/550) (E2E spec timing fix, MERGED 2026-06-11) — Test 2 자동 해결에 기여
  - [PR #574](https://github.com/ykylee/Devhub_example/pull/574) (N-13 follow-up A, Test 1 e2e seed fix, MERGED 2026-06-12) — strict mode violation bypass
  - [PR #575](https://github.com/ykylee/Devhub_example/pull/575) (N-13 follow-up B, Test 2 verification, MERGED 2026-06-12) — CI Run #1227 SUCCESS

## 1. 배경 (PR #548 CLOSED + follow-up 3 branch)

PR #548 (`feat/work_260611-a-n13-inbound-source-impl`, 2026-06-11 05:40 UTC CLOSED) 의 E2E Internal 1 fail 2건 + 자동 재실행 미적용 정공법. follow-up 3 branch 결정:

- **A: Test 1 e2e seed 중복 strict mode violation fix** — ✅ MERGED (PR #574, 2026-06-12, squash `896d9018`)
- **B: Test 2 Sign-out timeout rebase + 자동 재실행 검증** — ✅ MERGED (PR #575, 2026-06-12, squash `8d0e2e88`) + CI Run #1227 SUCCESS (`54eb8391`)
- **C: 구현 follow-up = v1.1 milestone 진입 시점 별도 sprint** — ⏳ **본 sprint** (PR A-1 + A-2 통합)

본 sprint = PR A-1 (backend foundation) + PR A-2 (routing + voc_handler 통합 + openapi + e2e) 통합 1 PR. PR #548 의 backend foundation 코드 (9 file, 340 line) + 본 sprint 의 routing + voc_handler 신규 작성 (4 file) + openapi.yaml 정합 + e2e TC-INBOUND-SRC-01 spec.

**branch**: `feat/work_260611-a-n13-inbound-source-impl` close (PR #548) — 본 sprint 의 신규 branch `feat/work_260612-4-v1-1-inbound-source-impl` 별도 결정.

## 2. 후보 옵션 (본 sprint scope 결정)

| # | 옵션 | 결정 |
| --- | --- | --- |
| **C-1** | **PR A-1 + PR A-2 통합 1 PR** | ⭐ **채택** (사용자 결정, 2026-06-12) |
| C-2 | PR A-1 만 (backend foundation 만) | 권장 보류 — PR A-2 별도 sprint 시 routing/voc_handler 미완성 |
| C-3 | PR A-1 + A-2 + ADR-0028 §6 amendment | 권장 보류 — ADR amendment 는 별도 sprint 가능 (ADR-0028 §6 (a) 의 본문 follow-up 결정 노트 추가 수준) |

### 2.1 C-1 상세 (채택) — **현실적 scope 재조정**

**2026-06-12 06:00 KST 정밀 분석 결과**: PR #548 의 backend foundation (PR A-1) 코드 9 file 은 **이미 main 에 byte-identical 로 존재** (`43feccfb` via PR #549 T-d-72-2 wiki mirror). `git diff ff4022f6 origin/main -- backend-core/` → no output. 따라서 **PR A-1 의 9 file 은 no-op (skip)**.

본 sprint 의 **실질적 scope = PR A-2 4 file 신규 작성 + openapi.yaml 정합 + e2e spec 신규 = 총 6 file**.

**backend (PR A-2, 4 file 신규)**:
1. `backend-core/internal/domain/application-lifecycle/routing/auto_route.go` (신규) — pattern matcher 3 case (external_ref pattern / requester email / req_department organization hierarchy) + auto route 1 case (synchronous routing at voc registration time)
2. `backend-core/internal/domain/application-lifecycle/routing/auto_route_test.go` (신규 UT) — pattern matcher 3 case + auto route 1 case + no match case
3. `backend-core/internal/domain/dev-request/view/voc_handler.go` (수정) — createOrGetVoc 에 자동 routing 호출 추가 + 응답 (auto_routed: true|false + dev_request_id: uuid)
4. `backend-core/internal/domain/dev-request/view/voc_handler_integration_test.go` (신규 IT) — TC-INBOUND-SRC-01 backend IT (PATCH inbound_source → POST voc → auto_routed 검증)

**docs (1 file 정합)**:
5. `docs/openapi.yaml` — PATCH /api/v1/platforms/:platform_id body inbound_source (inbound_source_type + inbound_source_config) + POST /api/v1/dev-requests/:external_ref 응답 auto_routed: true|false + dev_request_id: uuid 추가

**e2e (1 file 신규)**:
6. `frontend/tests/e2e/voc-auto-routing.spec.ts` (신규) — TC-INBOUND-SRC-01 E2E: PATCH inbound_source → POST voc → 자동 routing 결과 검증

**메모리**: state.json M-v1.0 `phase2_8th_chunk_n13_followup_c_v1_1_impl` + work_backlog.md + session_handoff.md + 브랜치 메모리 4 file

**PR A-1 (no-op, 이미 main 에 존재) — 작성/수정 불요**:
- `backend-core/migrations/000007_platform_inbound_source.up.sql` (이미 존재)
- `backend-core/migrations/000007_platform_inbound_source.down.sql` (이미 존재)
- `backend-core/internal/domain/application.go` (InboundSourceType/Config field 이미 존재)
- `backend-core/internal/domain/application-lifecycle/repository/applications.go` (COALESCE + UpdatePlatformInboundSource + ListEnabledInboundSourcePlatforms 이미 존재)
- `backend-core/internal/domain/application-lifecycle/view/handler.go` (PlatformStore interface +2 method 이미 존재)
- `backend-core/internal/domain/application-lifecycle/view/applications.go` (inbound_source 별도 처리 이미 존재)
- `backend-core/internal/domain/application-lifecycle/view/applications_handler_test.go` (4 UT 이미 존재)
- `backend-core/internal/domain/application-lifecycle/view/fake_store_test.go` (fake store impl 이미 존재)
- `backend-core/internal/httpapi/applications_test.go` (memory store +2 method 이미 존재)

### 2.2 PR A-1 정공법 핵심 (재정공법)

1. **scope = backend foundation + routing + voc_handler + openapi + e2e** (C-1 통합). PR A-1 의 PR #548 backend foundation + PR A-2 의 routing + voc_handler 통합 + openapi + e2e.
2. **inbound_source 격리**: `inboundTouched` 일 때 `UpdatePlatform` 호출 skip + `GetPlatform` 로 row 확인 + `UpdatePlatformInboundSource` 단독 호출. **inbound_source 부분 실패가 다른 필드 변경에 영향 없도록 격리**.
3. **migration 000007 의 CHECK 제약**: type='' 일 때 config NULL/'{}' 만 허용 (consistency). type whitelist = `''|gitea|jira|other` (4 종).
4. **routePermissionTable 변경 불요** — 기존 PATCH `/api/v1/platforms/:platform_id` entry 가 `ResourcePlatforms + ActionEdit` 으로 inbound_source 도 cover.
5. **fake store parity**: fakeViewPlatformStore + memoryPlatformStore 모두 production UpdatePlatformInboundSource 동작 (json.Valid check + ErrInvalidInboundSourceType/Config 매핑) mirror.
6. **PR A-2 routing 정공법**: voc 등록 시점에 `ListEnabledInboundSourcePlatforms` 조회 + 3 case pattern matcher (external_ref / requester / req_department) + match 시 voc skip + dev-request 직접 등록 (단일 트랜잭션). 매칭 없으면 voc 단계 유지 (현행 동작 보존).
7. **e2e TC-INBOUND-SRC-01**: PATCH inbound_source → POST voc → 자동 routing 결과 검증. PR #574 (Test 1 e2e seed fix) 의 strict mode violation bypass 정합 + PR #550 (E2E spec timing fix) 의 spec waitForURL buffer 정합.

## 3. 결정 (sprint 진입 시)

**옵션 C-1 (PR A-1 + PR A-2 통합) 채택**.

### 3.1 ID slot (구현 sprint 진입 시 발급)

PR #547 (N-13 housekeeping) 의 9 ID slot 정합. 본 sprint = 코드 변경 본 (planned → implemented).

| ID 종류 | ID | 스코프 | 본 sprint 의 status |
| --- | --- | --- | --- |
| REQ | `REQ-FR-113` | inbound_source 자동 routing 정책 + 매칭 전략 | implemented (planned → implemented) |
| UC | `UC-DEV-REQ-15` | 외부 시스템 의뢰 → applications 매칭 → dev-request 직접 등록 flow | implemented |
| ARCH | `ARCH-23` | applications.inbound_source schema + sync routing 결정 | implemented |
| API | `API-103` | PATCH `/api/v1/platforms/:platform_id` inbound_source + POST `/api/v1/dev-requests/:external_ref` auto_routed 응답 | implemented (openapi.yaml 정합) |
| RM | `RM-DEV-REQ-15` | sprint -d (구현 sprint) | implemented (본 sprint) |
| IMPL | `IMPL-inbound-source-01` | repository.UpdatePlatformInboundSource + routing/auto_route.go + voc_handler 통합 | implemented |
| IMPL | `IMPL-platform-patch-02` | UpdatePlatform 입력 validate + inbound_source_type CHECK 매핑 | implemented |
| UT | `UT-inbound-source-01` | pattern matcher 3 case + auto route 1 case + store 1 case + voc_handler IT | implemented |
| TC | `TC-INBOUND-SRC-01` | E2E: PATCH inbound_source → POST voc → auto_routed 검증 | implemented (e2e spec 추가) |

### 3.2 마일스톤 (release_v1_roadmap 정합)

- **본 sprint 의 역할**: v1.1 milestone (M-v1.1, target 2026-07-31) 진입 sprint.
- `release_v1_roadmap.md` §3.5 N-13 row status `⏳ planned` → `✅ resolved (implemented, 2026-06-12)` 정합.
- `release_v1_roadmap.md` §4.2 v1.1 milestone 에 본 sprint 의 ID slot 추가.
- `release_v1_roadmap.md` §9 변경 이력 row 추가.

### 3.3 의존

| 의존 | 상태 | 비고 |
| --- | --- | --- |
| ADR-0028 (voc + notification) | ✅ resolved (PR #514 + #515) | 본 sprint 의 source |
| `applications` table | ✅ stable | `migrations/000001_initial_schema.up.sql:48` |
| `integration_providers` table | ✅ stable | pattern 매칭 strategy 의 보완 후보 |
| `intake token` 인증 (ADR-0012) | ✅ resolved | 외부 시스템 인증 |
| Keycloak user lookup | ✅ stable | requester → user_id 매칭 |
| organization hierarchy | ✅ stable | req_department → unit_id 매칭 |
| PR #550 (E2E spec timing fix) | ✅ resolved | Test 2 자동 해결 |
| PR #574 (Test 1 e2e seed fix) | ✅ resolved | strict mode violation bypass |
| PR #548 (close) | ❌ CLOSED | 본 sprint 의 PR A-1 + A-2 통합본 = 신규 branch |

### 3.4 정합 항목

- **traceability §2.1 REQ-FR-113 + §2.1.5 UC-DEV-REQ-15 + §2.2 ARCH-23 + §2.2 API-103 + §2.3 RM-DEV-REQ-15 + §2.4 IMPL-inbound-source-01 + §2.4 IMPL-platform-patch-02 + §2.5 UT-inbound-source-01 + §2.6 TC-INBOUND-SRC-01**: 9 row status `planned` → `implemented` 갱신.
- **release_v1_roadmap §3.5 N-13 row** status `⏳ planned` → `✅ resolved (implemented, 2026-06-12)` 마킹.
- **release_v1_roadmap §4.2 v1.1 milestone** 본 sprint 의 ID slot 추가.
- **release_v1_roadmap §9** 변경 이력 row 추가.
- **ADR-0028 §6 (a)** 본문 정공법 (PR A-1 + A-2 통합본 = 본 sprint) + §7 변경 이력 row.
- **sprint plan v1** §6 변경 이력 row (본 sprint = v2).

## 4. 정공법 (구현 sprint 진입 시)

| 순번 | 산출물 | 의존 | 예상 |
| --- | --- | --- | --- |
| 1 | **migration 000007** (applications.inbound_source 2 field + CHECK) | (없음) | 30분 |
| 2 | **domain.Platform** 2 field 추가 | (없음) | 30분 |
| 3 | **platform-lifecycle repository** 2 method (UpdatePlatformInboundSource + ListEnabledInboundSourcePlatforms) | 1, 2 | 1시간 |
| 4 | **view.UpdatePlatform** inbound_source 별도 처리 (inboundTouched 시 UpdatePlatform skip + UpdatePlatformInboundSource 단독 호출) | 3 | 1시간 |
| 5 | **routing/auto_route.go** 신규 — pattern matcher 3 case + auto route 1 case (synchronous routing at voc registration time) | 2 | 2시간 |
| 6 | **voc_handler.createOrGetVoc** 자동 routing 호출 + 응답 (auto_routed: true|false + dev_request_id) | 4, 5 | 1시간 |
| 7 | **routePermissionTable** 1 entry 추가 (POST voc auto_routed 응답) | 4 | 15분 |
| 8 | **UT 4건 (GiteaOK / InvalidType400 / InvalidConfig400 / DisableEmpty)** (PR A-1 의 4 UT) | 3, 4 | 1시간 |
| 9 | **IT (voc_handler_integration_test.go)** pattern matcher 3 case + auto route 1 case + store 1 case | 5, 6 | 1.5시간 |
| 10 | **E2E TC-INBOUND-SRC-01** (voc-auto-routing.spec.ts) — PATCH inbound_source → POST voc → 자동 routing 결과 검증 | 6, 7 | 1.5시간 |
| 11 | **openapi.yaml** 정합 (PATCH /platforms inbound_source + POST voc auto_routed 응답) | 4, 6 | 30분 |
| 12 | **fake store parity** (fakeViewPlatformStore + memoryPlatformStore UpdatePlatformInboundSource 동작 mirror) | 3, 4 | 30분 |
| 13 | **traceability §2.1~§2.6 9 row status `planned` → `implemented` 갱신** | (없음) | 15분 |
| 14 | **release_v1_roadmap §3.5 N-13 row + §4.2 v1.1 + §9** 갱신 | 13 | 30분 |
| 15 | **ADR-0028 §6 (a) + §7** 갱신 | 14 | 15분 |
| 16 | **sprint plan v1** §6 변경 이력 row | 14 | 15분 |
| **합계** | | | **~12시간** (1 sprint, 2~3 PR 가능) |

본 sprint = 1 PR (C-1 통합본). 코드 ~800 line + 9 ID row implemented 정합.

## 5. Tier / CI

- **Tier**: **사내** (backend 코드 변경 + openapi.yaml, main branch push-only)
- **CI 11/12 PASS 예상** (path-detect → backend + openapi 변경 감지):
  - Backend Unit Tests → PASS
  - Backend Integration Tests → PASS
  - Frontend Unit Tests → PASS (e2e spec 추가 영향)
  - E2E Build Artifacts → PASS
  - E2E Tests (Playwright, shard 1/3) → **PASS** (Test 1 + Test 2 자동 해결)
  - E2E Tests (Playwright, shard 2/3) → **PASS**
  - E2E Tests (Playwright, shard 3/3) → **PASS**
  - Detect Changed Paths → success
  - Migration Prefix Uniqueness → success
  - OpenAPI YAML Lint → **PASS** (openapi.yaml 정합)
  - Workflow Lint (actionlint) → success
  - E2E Internal (real Keycloak adapter) → skip (PR A-1 의 saovae_stub default)

## 6. 결정 보류 사유

본 sprint = C-1 통합본 (PR A-1 + A-2). 본 sprint 의 코드 변경 후 main 의 e2e CI 가 자동 trigger 되어 e2e shard 1/2/3 모두 PASS 가정 (PR #574 + PR #550 정합).

## 7. 후속 (v1.1 milestone 진입 시점)

- 본 sprint = v1.1 milestone 의 첫 sprint
- v1.1 의 다른 항목 (X-1, X-2, X-3, X-4, X-5, X-6, X-7, X-8) 의 sprint 진입은 본 sprint 완료 후 결정
- ADR-0028 §6 (a) 의 implementation follow-up = 본 sprint 완료로 close

## 8. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-06-12 | 1차 작성 (sprint `feat/work_260612-4-v1-1-inbound-source-impl`) — PR #548 CLOSED 정공법 + follow-up 3 branch 중 C (구현 follow-up) 정공법. C-1 (PR A-1 + A-2 통합 1 PR) 채택. 15 file + ~800 line. 9 ID row planned → implemented. |
