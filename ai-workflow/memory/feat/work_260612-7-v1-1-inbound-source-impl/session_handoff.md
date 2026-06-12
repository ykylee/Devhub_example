# Session Handoff — feat/work_260612-7-v1-1-inbound-source-impl (2026-06-12, N-13 follow-up 종합 v1.1 sprint -a)

- 문서 목적: N-13 follow-up 결정 3 branch (A: e2e seed fix, B: signout timeout fix, C: 구현 follow-up) 의 종합 정공법 + 본 sprint 의 v1.1 sprint -a 진입 정공법.
- 범위: backend-core (routing/auto_route.go + voc_handler 통합) + openapi.yaml + e2e spec + ADR-0028 §6 amendment + release_v1_roadmap.md §3.5 + traceability/report.md 9 ID row 정합.
- 상태: in_progress. main HEAD `0d2dd89` (PR #578 e2e-internal job 폐기 + 메모리 finalize).
- 최종 수정일: 2026-06-12

## 0. 본 sprint 결정 사항 (2026-06-12)

### N-13 follow-up 3 branch 종합 결정

sprint `fix/work_260612-1-n13-housekeeping-followup` (PR #573, 2026-06-12 MERGED) 의 follow-up 결정 3 branch:

| branch | 정공법 | 결과 |
|---|---|---|
| **A** Test 1 e2e seed 중복 | spec/e2e seed 정합 fix 별도 sprint | **PR #574 MERGED** (2026-06-12, `fix/work_260612-2-e2e-seed-strict-mode-fix`) — repositories-ui.spec.ts 의 `getByText('e2e-repo-a')` 2 elements strict mode violation fix |
| **B** Test 2 Sign-out timeout | main rebase + 자동 재실행 검증 별도 (PR #550 spec timing fix 가 해결 가능성 ↑) | **PR #575 MERGED** + **CI Run #1227 SUCCESS** (2026-06-12, `fix/work_260612-3-n13-followup-b-test2-rebase`) — PR #550 spec timing fix 가 main 에 머지 후 자동 해결 |
| **C** 구현 follow-up | v1.1 milestone 진입 시점 별도 sprint (rebase main + PR #550 fix + e2e seed 정합 fix + 자동 재실행 종합) | **본 sprint** (`feat/work_260612-7-v1-1-inbound-source-impl`) — 종합 v1.1 sprint -a 진입 |

### PR #548 close 정공법 (N-13 1차 구현)

PR #548 (`feat/work_260611-a-n13-inbound-source-impl`, 2026-06-11 05:40 UTC CLOSED) 의 E2E Internal 1 fail 2건 (Test 1 + Test 2) + 자동 재실행 미적용 정공법. 본 sprint 의 정공법 = PR #548 의 **rebase main (post-merge PR #574 + #575 + #578) + PR A-1 의 backend foundation 유지 + PR A-2 (routing + voc_handler + openapi + e2e) 신규 구현 + 자동 재실행 종합**.

### PR A-1 vs PR A-2 scope 결정

sprint plan v2 의 결정:
- **PR A-1 (backend foundation)**: no-op (skip) — 이미 main 에 byte-identical 로 존재 (PR #549 T-d-72-2 wiki mirror).
- **PR A-2 (routing + voc_handler + openapi + e2e)**: 5-6 file 신규 작성 — 본 sprint 의 **실질적 scope**.

본 sprint 의 scope = **PR A-2 만** (5-6 file 신규 + 2-3 file 수정, +600~800 lines).

## 1. 변경 요약 (5-6 file 신규 + 2-3 file 수정)

### 1.1 Backend (3 file)

| # | 파일 | 변경 | line |
| --- | --- | --- | --- |
| 1 | `backend-core/internal/domain/application-lifecycle/routing/auto_route.go` | 신규 — 3 case pattern matcher + auto route 1 case | +126 |
| 2 | `backend-core/internal/domain/application-lifecycle/routing/auto_route_test.go` | 신규 UT — 6 test case (ExternalRef / Requester / ReqDepartment / NoMatch / MultiplePlatforms / EmptyPlatforms) | +166 |
| 3 | `backend-core/internal/domain/dev-request/view/voc_handler.go` | 수정 — createOrGetVoc 에 AutoRouter.Route() 호출 + RouteVoc() 통합 + auto_routed 응답 | +114/-42 |

### 1.2 Integration Test (1 file)

| # | 파일 | 변경 | line |
| --- | --- | --- | --- |
| 4 | `backend-core/internal/domain/dev-request/view/voc_handler_integration_test.go` | 신규 IT — 3 IT (GiteaOK / NoMatch / RouteErrorDegradation) | +221 |

### 1.3 OpenAPI + E2E (2 file)

| # | 파일 | 변경 | line |
| --- | --- | --- | --- |
| 5 | `docs/openapi.yaml` | 정합 — PATCH /platforms inbound_source + POST /dev-requests/{dev_request_id} auto_routed 응답 + DevRequestVoc schema | +216 |
| 6 | `frontend/tests/e2e/voc-auto-routing.spec.ts` | 신규 E2E — TC-INBOUND-SRC-01 (PATCH → POST → auto_routed 검증, seed platform 사용) | +157 |

### 1.4 Docs (3 file)

| # | 파일 | 변경 | line |
| --- | --- | --- | --- |
| 7 | `docs/adr/0028-dev-requests-voc-external-ref.md` | §6 (a) amendment (구현 정합) | +20/-10 |
| 8 | `docs/planning/release_v1_roadmap.md` | §3.5 N-13 row status `⏳ planned` → `✅ resolved (implemented, 2026-06-12)` + §4.2 v1.1 milestone + §9 | +15/-3 |
| 9 | `docs/traceability/report.md` | §2.1~§2.6 9 ID row status `planned` → `implemented` | +0/-0 (cell fill 만) |

## 2. 정공법 핵심

### 2.1 auto_route.go (pattern matcher)

**3 case pattern matcher**:
- **Case 1 (external_ref pattern)**: `^GITEA-([0-9]+)$` regex match + `source_system == "gitea"` → `inbound_source_type == "gitea"` platform 매칭
- **Case 2 (requester email)**: requester string → DevelopmentUnitID exact match
- **Case 3 (req_department)**: req_department string → DevelopmentUnitID exact match
- **우선순위**: external_ref > requester > req_department. **첫 매칭 return**.

**graceful degradation**: 매칭 없으면 → voc 단계 유지 (현행 동작 보존). 라우팅 결정은 sync (worker 아님 — sprint plan v2 의 "post-MVP 검토에서 worker vs sync 결정" → sync 결정).

### 2.2 voc_handler 통합 (createOrGetVoc)

voc INSERT 후 `AutoRouter.Route()` 호출 → 매칭 시 `RouteVoc()` 으로 dev-request 자동 생성 + voc status=routed 전이. **graceful degradation**: route error 시에도 voc creation 은 완료 (received 상태 유지).

응답 envelope 정합:
```json
{
  "auto_routed": true|false,
  "platform_id": "<uuid>",
  "dev_request_id": "<uuid>",
  "reason": "external_ref pattern matched platform_id=..." | "no match",
  "voc": {...}  // 기존 voc 응답
}
```

### 2.3 openapi.yaml 정합

- `PATCH /api/v1/platforms/{platform_id}` body 에 `inbound_source_type` (enum: `""` | `gitea` | `jira` | `other`) + `inbound_source_config` (string, JSONB) 추가.
- `POST /api/v1/dev-requests/{dev_request_id}` 응답에 `auto_routed` + `platform_id` + `dev_request_id` + `reason` 필드 추가.
- `DevRequestVoc` schema 정의.
- Envelope oneOf 등록.

### 2.4 e2e TC-INBOUND-SRC-01

**`seed platform` 사용** (PR #548 의 1차 spec 의 POST platform 단계 제거 정공법 + e2e seed 중복 회피). 단계:
1. `PATCH /api/v1/platforms/{seed_platform_id}` body `inbound_source_type=gitea` + `inbound_source_config={...}` (e2e 환경의 systemAdmin token).
2. `POST /api/v1/dev-requests/GITEA-1234` (external_ref pattern match) + body 9 field.
3. 응답 envelope 의 `auto_routed == true` + `dev_request_id` 검증.
4. (negative) `POST /api/v1/dev-requests/MANUAL-XYZ` (no match) → `auto_routed == false` + voc status=received.

**frontend spec 의 strict mode violation bypass** 정합 (PR #574 의 Test 1 fix 의 영향 안 받음 — 별도 spec). **N-8 race-free** 정합 (signout timeout 영향 없음 — logout flow 무관).

**PR #579 1차 commit (1차 CI) E2E shard 3/3 fail 정공법 (옵션 B)**:
- **근본 layer**: shard 3/3 의 Keycloak container 가 initial start-up race 로 늦게 ready (GitHub Actions runner 의 transient network issue, `curl: (56) Recv failure: Connection reset by peer` 9 회). beforeEach 의 loginAs 가 systemAdmin token 받기 전에 backend 401 → patchResp 4xx.
- **정공법**: PATCH inbound_source 를 `test.beforeAll` hook 으로 이동 (Keycloak startup race 회피) + retry 3 회 with backoff (5s + 10s + 15s, 최대 4 attempts). loginAs 도 beforeAll 에서 1 회 (Keycloak ready 보장). 2 test case 의 PATCH 단계 제거 — 검증만.
- **beforeAll 의 context.close()** 명시 — leak 방지. throw 시 last error 메시지 + 4 attempts 명시.
- **N-13 follow-up 3 branch 결정 (PR #573) 의 종합 정공법** 정합: PR #574 (Test 1) + PR #575 (Test 2) + 본 sprint (Test 1+2+구현 종합, N-13 follow-up C 의 정공법). 본 sprint 의 beforeAll 정공법 = **N-13 follow-up 3 branch 결정의 종합 + N-13 follow-up C 의 정식 구현 + Keycloak race 정공법** 정합.

## 3. 추적성 영향 (9 ID row planned → implemented 정합)

| ID | status | 정공법 |
| --- | --- | --- |
| REQ-FR-113 | planned → implemented | 본 sprint |
| UC-DEV-REQ-15 | planned → implemented | 본 sprint |
| ARCH-23 | planned → implemented | 본 sprint |
| API-103 | planned → implemented | 본 sprint (openapi.yaml 정합) |
| RM-DEV-REQ-15 | planned → implemented | 본 sprint |
| IMPL-inbound-source-01 | planned → implemented | 본 sprint (routing/auto_route.go + voc_handler 통합) |
| IMPL-platform-patch-02 | planned → implemented | PR A-1 (main 에 byte-identical 포함) |
| UT-inbound-source-01 | planned → implemented | 본 sprint (UT 6 + IT 3) |
| TC-INBOUND-SRC-01 | planned → implemented | 본 sprint (e2e voc-auto-routing.spec.ts) |

## 4. 후속 (사용자 결정 영역)

- PR 머지 후 v1.1 sprint -b (gitea + ci port) 진입 결정
- v1.0 staging 1주 운영 (N-6) 시작 (사용자 결정)
- ADR-0028 §6 (a) 의 implementation follow-up status `⏳ planned` → `✅ resolved (implemented, 2026-06-12)` 정합 (본 sprint 의 §7 docs 갱신에 포함)

## 5. 검증 (PR 머지 직전)

- [ ] `cd backend-core && go build ./...` PASS
- [ ] `cd backend-core && go test ./internal/domain/application-lifecycle/routing/ ./internal/domain/dev-request/view/` ALL PASS (6 UT + 3 IT)
- [ ] `cd backend-core && go vet ./...` PASS (변경 package)
- [ ] `cd backend-core && go test ./...` PASS (전체 회귀 0)
- [ ] `cd frontend && tsc --noEmit` pre-existing errors only (변경 file 무관)
- [ ] `bash scripts/check-tier-separation.sh` PASS
- [ ] `bash scripts/check-openapi-yaml-lint.sh` PASS (openapi.yaml 정합)
- [ ] `bash scripts/check-migration-uniqueness.sh` PASS (migration 변경 없음, sanity check)
- [ ] workflow-lint + e2e shard 1/2/3 PASS (CI)
- [ ] `git diff --stat` = 9 file 변경 (3 backend + 1 IT + 1 openapi + 1 e2e + 3 docs)
- [ ] `git log -1 --format='%an %ae'` = 본 세션 author

## 6. E2E shard 3/3 fail 정공법 (옵션 B, PR #579 2차 commit + 옵션 A 재시도 + 3차 commit)

**근본 layer**: PR #579 1차 commit 의 E2E shard 3/3 fail 2건 (TC-INBOUND-SRC-01 + NEG). shard 3/3 의 Keycloak container 가 initial start-up race 로 늦게 ready (GitHub Actions runner 의 transient network issue, `curl: (56) Recv failure: Connection reset by peer` 9 회). beforeEach 의 loginAs 가 systemAdmin token 받기 전에 backend 401 → patchResp 4xx.

**정공법 (옵션 B, 2차 commit)**: PATCH inbound_source 를 `test.beforeAll` hook 으로 이동 (Keycloak startup race 회피). retry 3 회 with backoff (0s + 5s + 10s + 15s, 최대 4 attempts). loginAs 도 beforeAll 에서 1 회 (Keycloak ready 보장). 2 test case 의 PATCH 단계 제거 — 검증만.

**옵션 A (재시도 1회, 3차 rerun)**: 동일 shard 3/3 fail (6m12s) — 1차 (5m32s) + 2차 (8m30s) + 3차 (6m12s) 모두 fail. **3회 연속 shard 3/3 fail 의 chronic flake** 확인. 옵션 A 한계 도달.

**근본 layer 정밀 확정 (3회 fail log 분석)**:
- `Wait for imported Keycloak realm` step: curl: (56) Recv failure × 8회 + curl: (52) Empty reply × 3회 = 11회. step 은 timeout 120s 안에 결국 PASS.
- `Wait for App Readiness` step: timeout 60s + 120s 모두 PASS.
- **`beforeAll` hook timeout default 30000ms** 가 부족 — loginAs 1회 (page.waitForURL timeout 30s) + retry 3회 (5s+10s+15s=30s) 합쳐서 60s 가 필요. **Playwright 의 beforeAll hook default 30s 안** 에 안 들어감.
- **error message**: `"beforeAll" hook timeout of 30000ms exceeded.`

**3차 commit 정공법 (근본 fix)**:
- `test.beforeAll` 의 timeout 30000ms → **180000ms (3분) 명시** (Playwright `{ timeout: 180_000 }` option)
- loginAs 의 page.waitForURL timeout 30000ms → 60000ms (within beforeAll) — fixtures.ts 변경 OR loginAs 호출자 옵션
- retry 3 회 backoff (5s+10s+15s=30s) 유지
- 2 test case 의 PATCH 단계 유지 (beforeAll 에서 1 회 PATCH 통합)

**근본 layer 종합 (4-step)**:
- PR #548 (1차) → PR #574/575/576 (fix 1차 fail 2건) →
- PR #579 1차 (구현 + 1차 fail 2건 fix 의 종합) →
- PR #579 2차 (beforeAll fix, **timeout 부족으로 2차 fail**) →
- PR #579 3차 (beforeAll timeout 명시, **3분 버퍼**)
- **e2e spec 변경 0** (코드 변경 없이 spec 의 timeout option 만)

**상세 변경**:
- `frontend/tests/e2e/voc-auto-routing.spec.ts` (110 lines → 112 lines) — `test.beforeAll(async () => {...}, { timeout: 180_000 })` option 추가
- e2e spec 의 PATCH retry 3 회 backoff 정공법은 유지
- `ai-workflow/memory/feat/work_260612-7-v1-1-inbound-source-impl/session_handoff.md` — 본 §6 append (정공법 + 4-step 종합 정합)
- `ai-workflow/memory/feat/work_260612-7-v1-1-inbound-source-impl/work_backlog.md` — T-7c row 추가

**검증 (PR #579 3차 commit, 머지 직전)**:
- [ ] `git diff --stat` = 1 file 변경 (e2e spec)
- [ ] `git log -1 --format='%an %ae'` = 본 세션 author
- [ ] CI e2e shard 3/3 PASS (beforeAll timeout 180s 명시로 Keycloak race 회피)
