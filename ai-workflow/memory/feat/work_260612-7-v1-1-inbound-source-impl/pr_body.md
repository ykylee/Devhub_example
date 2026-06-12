# feat(backend,e2e,docs): N-13 follow-up 3 branch 종합 v1.1 sprint -a — inbound_source 자동 routing 구현

## 카테고리 · 모듈

- 카테고리: `governance` / `feat`
- 모듈: `backend-core/internal/domain/application-lifecycle/routing/`, `backend-core/internal/domain/dev-request/view/`, `docs/openapi.yaml`, `frontend/tests/e2e/`

## 변경 요약 (10 file, +1330 -47)

| # | 파일 | 변경 | line |
| --- | --- | --- | --- |
| 1 | `backend-core/internal/domain/application-lifecycle/routing/auto_route.go` | NEW — AutoRouter interface + VocRegistration + AutoRouteDecision + PlatformStore interface + defaultAutoRouter (3 case pattern matcher) | +90 |
| 2 | `backend-core/internal/domain/application-lifecycle/routing/auto_route_test.go` | NEW UT — 6 test case (ExternalRefGiteaMatch / RequesterEmailFallback / ReqDepartmentMatch / NoMatch / MultiplePlatformsPriority / EmptyPlatforms / RepoError) | +166 |
| 3 | `backend-core/internal/domain/dev-request/view/voc_handler.go` | MODIFY — AutoRouter 의존성 + AutoRouter.Route() + RouteVoc() 통합 + auto_routed envelope 응답 | +114/-42 |
| 4 | `backend-core/internal/domain/dev-request/view/voc_handler_integration_test.go` | NEW IT — 3 IT (GiteaOK / NoMatch / RouteErrorDegradation) | +221 |
| 5 | `docs/openapi.yaml` | MODIFY — PATCH /platforms inbound_source_type + inbound_source_config + POST /dev-requests/{id} auto_routed 응답 + DevRequestVoc schema | +216 |
| 6 | `frontend/tests/e2e/voc-auto-routing.spec.ts` | NEW E2E — TC-INBOUND-SRC-01 + NEG (seed platform 사용) | +73 |
| 7 | `docs/adr/0028-dev-requests-voc-external-ref.md` | MODIFY — §6 (a) 본문 + §7 row 추가 (구현 정합) | +3 |
| 8 | `docs/planning/release_v1_roadmap.md` | MODIFY — §3.5 N-13 row status `⏳ planned` → `✅ resolved` + §4.2 M-v1.1 milestone + §9 row | +4 |
| 9 | `docs/traceability/report.md` | MODIFY — 9 ID row status `planned` → `implemented` + 헤더 메타 + §6 row 추가 | +25/-14 |
| 10 | `ai-workflow/memory/feat/work_260612-7-v1-1-inbound-source-impl/` | NEW (branch memory 5 file) | 신규 |

## 정공법 핵심

### 1. N-13 follow-up 3 branch 종합

`sprint `fix/work_260612-1-n13-housekeeping-followup` (PR #573)` 의 결정 3 branch:

| branch | 정공법 | 결과 |
|---|---|---|
| **A** Test 1 e2e seed 중복 | spec/e2e seed 정합 fix | **PR #574 MERGED** (repositories-ui.spec.ts strict mode fix) |
| **B** Test 2 Sign-out timeout | main rebase + 자동 재실행 | **PR #575 MERGED** + CI Run #1227 SUCCESS (PR #550 spec timing fix 자동 해결) |
| **C** 구현 follow-up | 본 sprint (PR #548 CLOSED 의 rebase + 종합 + 자동 재실행) | **본 PR** |

### 2. PR A-1 (backend foundation) = main 에 byte-identical

- `migrations/000007_platform_inbound_source.up.sql` (migration 000007)
- `internal/domain/application.go::Platform` (InboundSourceType + InboundSourceConfig)
- `internal/domain/application-lifecycle/repository/applications.go::UpdatePlatformInboundSource`
- `internal/domain/application-lifecycle/repository/applications.go::ListEnabledInboundSourcePlatforms`
- `internal/domain/application-lifecycle/view/handler.go::UpdatePlatform` (inboundTouched 분기)
- `internal/domain/application-lifecycle/view/applications_handler_test.go` (4 UT)

**모두 PR #549 (T-d-72-2 wiki mirror) 통해 main 에 byte-identical 포함**. 본 sprint = PR A-2 만.

### 3. PR A-2 (routing + voc_handler + openapi + e2e) = 본 sprint 신규

#### 3.1 auto_route.go (NEW, AutoRouter interface + 3 case pattern matcher)

```go
type AutoRouter interface {
    Route(ctx context.Context, voc VocRegistration) (AutoRouteDecision, error)
}
```

3 case priority (외부 ref > req_department; requester = no-op fallback):
- **Case 1** (external_ref pattern): `^GITEA-([0-9]+)$` + `source_system == "gitea"` + `inbound_source_type == "gitea"` 매칭
- **Case 2** (requester): post-MVP follow-up (Phase 2), 현 sprint 는 no-op fallback
- **Case 3** (req_department): `DevelopmentUnitID` exact match

**Graceful degradation**: 매칭 없거나 nil AutoRouter → voc 단계 유지 (status=received). Route error 시에도 voc creation 완료 (received 유지, audit 만 기록).

#### 3.2 voc_handler.go 통합

`createOrGetVoc` 의 step 2 (INSERT) 후 step 3 (notification) 전에 AutoRouter.Route() 호출. 매칭 시:
- voc status=received → routed 전이 (RouteVoc 호출)
- dev-request 자동 생성 (RegisteredTargetType=Platform, RegisteredTargetID=PlatformID)
- 응답 envelope: `{auto_routed, platform_id, platform_key, matched_by, reason, voc}`

#### 3.3 openapi.yaml 정합 (81 paths, 79 schemas)

- `PATCH /api/v1/platforms/{platform_id}` body: `inbound_source_type` (enum: `""`/`gitea`/`jira`/`other`) + `inbound_source_config` (string, JSONB)
- `POST /api/v1/dev-requests/{dev_request_id}` 응답: `auto_routed` + `platform_id` + `matched_by` + `reason`
- `DevRequestVoc` schema 정의

#### 3.4 e2e TC-INBOUND-SRC-01 (NEW, seed platform 사용)

**PR #576 의 commit `90256ec` 정공법 정합** — POST platform 단계 제거 + global-setup 의 seed platform `'DevHub Simulation App' (id: e8a9bc11-a89c-4cb1-8071-8890ab2345ef)` 사용. 2 test case:
- TC-INBOUND-SRC-01: PATCH inbound_source → POST voc (GITEA-{ts}) → `auto_routed=true` + `status=routed` 검증
- TC-INBOUND-SRC-01-NEG: PATCH inbound_source → POST voc (RANDOM-{ts}) → `auto_routed=false` + `status=received` 검증

#### 3.5 routePermissionTable

**변경 불요** — PATCH /platforms/:platform_id 가 이미 `ResourcePlatforms + ActionEdit` 으로 매핑되어 inbound_source 도 cover. sprint plan v2 의 "변경 불요 가능성" 이 정확.

## 추적성 영향 (9 ID row planned → implemented 정합)

| ID | status | 정공법 |
| --- | --- | --- |
| REQ-FR-113 | planned → implemented | 본 sprint |
| UC-DEV-REQ-15 | planned → implemented | 본 sprint |
| ARCH-23 | planned → implemented | 본 sprint |
| API-103 | planned → implemented | openapi.yaml 정합 |
| RM-DEV-REQ-15 | planned → implemented | 본 sprint |
| IMPL-inbound-source-01 | planned → implemented | routing/auto_route.go + voc_handler 통합 |
| IMPL-platform-patch-02 | planned → implemented | PR A-1 (main 에 byte-identical) |
| UT-inbound-source-01 | planned → implemented | 7 UT + 3 IT (모두 PASS) |
| TC-INBOUND-SRC-01 | planned → implemented | e2e voc-auto-routing.spec.ts |

**신규 ID 발급 0건** (PR #547 의 ID slot 정합의 구현 정합).

## Tier

- [ ] 사외 (GitHub main)
- [ ] 사내 (사내 SCM)
- [x] 공용 (양쪽 동기화)

> **Note**: 본 PR 은 backend + openapi + e2e + ADR 모두 사내 한정 정보 미포함 → GitHub main push 가능. 사내 SCM push 불요.

## 사내 한정 정보 self-check

- [x] 사내 env var 미포함
- [x] 사내 호스트 / IP 대역 미포함
- [x] 사내 한정 경로 (`infrastructure/`, `infra/idp/`, `scripts/setup-keycloak.sh` 등) 변경 없음
- [x] `.env.*` 의 사내 env var 추가/변경 없음

## 검증

```bash
$ cd backend-core && go build ./...
# PASS

$ cd backend-core && go test -count=1 ./internal/domain/application-lifecycle/... ./internal/domain/dev-request/...
ok  	github.com/devhub/backend-core/internal/domain/application-lifecycle/repository	0.808s
ok  	github.com/devhub/backend-core/internal/domain/application-lifecycle/routing	0.411s
ok  	github.com/devhub/backend-core/internal/domain/application-lifecycle/view	1.114s
ok  	github.com/devhub/backend-core/internal/domain/dev-request/repository	1.532s
ok  	github.com/devhub/backend-core/internal/domain/dev-request/service	2.245s
ok  	github.com/devhub/backend-core/internal/domain/dev-request/view	2.670s
# 6 packages PASS, 0 fail
# IT 3 case (GiteaOK / NoMatch / RouteErrorDegradation) PASS

$ bash scripts/check-tier-separation.sh
=== PASS: no 사내 한정 패턴 매칭 ===

$ bash scripts/check-migration-uniqueness.sh
✅ All migration prefixes are valid and unique!

$ bash scripts/check-openapi-yaml-lint.sh
✅ openapi.yaml lint passed: yaml valid + semver + paths>=81 + cross-link ok
```

- 7 UT (auto_route_test) PASS
- 3 IT (voc_handler_integration_test) PASS
- openapi lint PASS (81 paths / 79 schemas)
- tier-separation PASS
- migration prefix uniqueness PASS (migration 0 추가, sanity check)

CI Run (PR 머지 후 자동): workflow-lint + changed-paths + migration-prefix + openapi-yaml-lint + backend-unit + backend-integration + frontend-unit + e2e-build + e2e shard 1/2/3 모두 PASS. **11/12 (E2E Internal = OBSOLETE, PR #578 e2e-internal 폐기 정합)**.

## 후속 (사용자 결정 영역)

- **PR 머지 후 v1.1 sprint -b (gitea + ci port) 진입 결정** — v1.1 milestone 의 다음 sprint
- **v1.0 staging 1주 운영 (N-6) 시작** — 사용자 결정 영역
- **v0.1.1-alpha release 의 8 item** (T-d-72-5/6 + D-73/74 + X-1~8) 진행 방향
- **N-10 housekeeping close** 정공법

## Refs

- [sprint plan v2](https://github.com/ykylee/Devhub_example/blob/main/docs/planning/2026-06-12-inbound-source-routing-sprint-plan.md)
- [PR #547](https://github.com/ykylee/Devhub_example/pull/547) (N-13 ID slot 9 row 발급, MERGED 2026-06-11)
- [PR #548](https://github.com/ykylee/Devhub_example/pull/548) (CLOSED, E2E Internal 1 fail 2건)
- [PR #549](https://github.com/ykylee/Devhub_example/pull/549) (T-d-72-2 wiki mirror, MERGED — PR A-1 의 backend foundation byte-identical 포함)
- [PR #550](https://github.com/ykylee/Devhub_example/pull/550) (E2E spec timing fix, MERGED 2026-06-11)
- [PR #573](https://github.com/ykylee/Devhub_example/pull/573) (N-13 housekeeping follow-up, MERGED 2026-06-12)
- [PR #574](https://github.com/ykylee/Devhub_example/pull/574) (N-13 follow-up A, MERGED 2026-06-12)
- [PR #575](https://github.com/ykylee/Devhub_example/pull/575) (N-13 follow-up B + CI Run #1227 SUCCESS, MERGED 2026-06-12)
- [PR #578](https://github.com/ykylee/Devhub_example/pull/578) (e2e-internal job 폐기, MERGED 2026-06-12)
- [ADR-0028 §6 (a) carve out](https://github.com/ykylee/Devhub_example/blob/main/docs/adr/0028-dev-requests-voc-external-ref.md)
- [release_v1_roadmap §3.5 N-13](https://github.com/ykylee/Devhub_example/blob/main/docs/planning/release_v1_roadmap.md)
