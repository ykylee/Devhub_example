# 2026-05-29 Test Coverage Sprint 결과 보고서

- 문서 목적: 2026-05-29 단일 세션에 진행된 backend + frontend test coverage 보강 sprint 의 결과 + silent production bug 해소 + test infrastructure 강화 내역을 정리한다.
- 범위: backend-core 의 `internal/store` / `internal/domain/application-lifecycle/*` / `internal/domain/rbac-permissions/*` + frontend 의 `components/dashboard,organization` / `lib/config` / `vitest.config.ts` + `docs/tests/test_coverage_carve_out_plan.md`
- 대상 독자: 저장소 관리자, codex 외부 리뷰어, 후속 carve out 진행자
- 상태: published
- 최종 수정일: 2026-05-29
- 관련 문서: `docs/tests/test_coverage_carve_out_plan.md`, `docs/governance/code-taxonomy.md`, `docs/adr/0019-keycloak-only.md`, `docs/adr/0020-keycloak-event-listener-spi.md`

## 1. 요약

- 단일 세션 **8 PR main 머지** (#435 → #442). main HEAD `d150dfc` (2026-05-29 KST).
- backend 실 DB 기준 `-coverpkg=./...` total **43.0% → 54.4%** (+11.4%p).
- `internal/domain/application-lifecycle/view` self-coverage **7.0% → 90.2%** (+83.2%p).
- frontend overall **Lines 74.39% → 81.8%** (+7.41%p).
- **silent production bug 1건 해소** (RBAC handler 의 422 매핑 회귀, `errors.Is` sentinel duplicate 가 silent fail).
- **test infrastructure 강화**: cross-package fixture race 차단 (PostgreSQL advisory lock), silent err 무시 패턴 정정, vitest coverage exclude 정정.
- **신규 test 362건** (backend 174 + frontend 188), 신규 LoC ~6300 (production code 변경 0).

## 2. 머지된 PR 목록 (시계열)

| # | PR | 영역 | 효과 | main commit |
|---|---|---|---|---|
| 1 | [#435](https://github.com/ykylee/Devhub_example/pull/435) | backend / docs | B-5 carve out — store integration test 실 DB cover 측정 (43.0% → 54.4%) | `db2b065` |
| 2 | [#436](https://github.com/ykylee/Devhub_example/pull/436) | backend / view test | B-3 9 도메인 handler shim test 142건 (audit-ops 47.6%, application-lifecycle 7.0% 미달) | `f2a2ee2` |
| 3 | [#437](https://github.com/ykylee/Devhub_example/pull/437) | backend / store test | uuid empty FAIL 회귀 — silent err 무시 패턴 정정 | `084a6a4` |
| 4 | [#438](https://github.com/ykylee/Devhub_example/pull/438) | backend / RBAC | **silent production bug** — sentinel duplicate 통합 (handler 422 매핑 회귀 hotfix) | `00afc2c` |
| 5 | [#439](https://github.com/ykylee/Devhub_example/pull/439) | backend / test infra | cross-pkg fixture race 차단 — pg_advisory_lock (`0x4150705F4C6966` "App_Lif") | `11e7b78` |
| 6 | [#440](https://github.com/ykylee/Devhub_example/pull/440) | backend / view test | application-lifecycle view 보강 — self-coverage **7.0% → 90.2%** (178 신규 test) | `3f1b2b9` |
| 7 | [#441](https://github.com/ykylee/Devhub_example/pull/441) | frontend / config | F-1 잔여 — dead code 제거 + vitest coverage exclude (Lines 74.4% → 75.1%) | `50f7481` |
| 8 | [#442](https://github.com/ykylee/Devhub_example/pull/442) | frontend / 도메인 component | D 옵션 — 0% 도메인 component 보강 (Lines 75.1% → 81.8%) | `d150dfc` |

## 3. Coverage 변화

### 3.1 Backend (`-coverpkg=./...` 실 DB 기준)

| 패키지 | before | after | delta |
|---|---|---|---|
| `-coverpkg=./...` total | 43.0% | **54.4%** | +11.4%p |
| `internal/store` | 0% (env skip) | **20.2%** | +20.2%p |
| `internal/domain/application-lifecycle/repository` | 0% | **43.6%** | +43.6%p |
| `internal/domain/application-lifecycle/view` | 0% | **90.2%** | +90.2%p |
| `internal/domain/dev-request/repository` | 0% | **24.1%** | +24.1%p |
| `internal/domain/audit-ops/view` | 0% | **47.6%** | +47.6%p |
| `internal/domain/onboarding/view` | 27.5% | **42.3%** | +14.8%p |
| `internal/domain/organization-management/view` | 0% | **15.4%** | +15.4%p |
| `internal/domain/rbac-permissions/view` | 10.5% | **15.2%** | +4.7%p |
| `internal/domain/repository-integration/view` | 0% | **13.5%** | +13.5%p |
| `internal/domain/dev-request/view` | 0% | **14.0%** | +14.0%p |
| `internal/domain/integration-registry/view` | 0% | **11.1%** | +11.1%p |
| `internal/domain/realtime/view` | 21.7% | **27.8%** | +6.1%p |

### 3.2 Frontend (전체 `npm run test:coverage`)

| metric | before | after | delta |
|---|---|---|---|
| **Lines** | 74.39% | **81.8%** | +7.41%p |
| Statements | 72.49% | 80.34% | +7.85%p |
| Branches | 72.75% | 76.03% | +3.28%p |
| Functions | 68.68% | 77.25% | +8.57%p |

도메인 component 추가 cover:

| component | before | after |
|---|---|---|
| `components/dashboard/GardenerFeed.tsx` | 0% | **96.96%** |
| `components/organization/OrgTree.tsx` | 0% | **75.27%** |
| `components/organization/OrgUnitTable.tsx` | 0% | 정상 (42 test PASS, v8 표 truncate) |

## 4. Silent Production Bug 해소 (#438)

### 4.1 Root cause

`internal/domain/rbac-permissions/repository/postgres_rbac.go` 의 `ErrSystemRoleImmutable` / `ErrRoleInUse` (production code 가 `fmt.Errorf("...: %w", ErrXxx)` 로 wrap 하던 sentinel) 와 `internal/store/options.go` 의 동명 sentinel 이 **같은 message 의 서로 다른 `*errors.errorString` instance**. `errors.Is` 는 pointer 비교 + `Is()` 메서드 사용 — string 비교 X. 따라서 `internal/domain/rbac-permissions/view/rbac.go` 의 handler 가 `errors.Is(err, store.ErrSystemRoleImmutable)` / `store.ErrRoleInUse` 시도 → 항상 false → default 경로 (500 internal error) 로 떨어짐.

추가로 `store.ErrRoleInUse` 의 message (`"role is in use"`) 가 production wrap message (`"role is still assigned to subjects"`) 와 다른 잔재.

### 4.2 사용자 facing impact (해소 전 → 해소 후)

| endpoint | before | after |
|---|---|---|
| `DELETE /api/v1/rbac/policies/<system_role>` | 500 internal error | **422 system_role_not_deletable** |
| `DELETE /api/v1/rbac/policies/<custom-in-use>` | 500 internal error | **422 role_in_use** |
| PATCH 시스템 role metadata (via `update_policies` bulk) | 500 internal error | **422 system_role_immutable** |

### 4.3 발견 경로

직전 sprint 의 실 DB integration test 두 건 (`TestRBAC_SystemRoleImmutable`, `TestRBAC_DeleteCustomRoleInUse`) 이 우연히 발견. 단위테스트로 cover 되지 않은 영역의 silent 회귀가 실 DB cover 측정 sprint 의 부수효과로 표면화. **B-5 + B-3 sprint 의 직접 산출 외에 가장 큰 가치**.

### 4.4 Fix

- `internal/store/options.go::ErrRoleInUse` message 를 production wrap 과 일치 (`"role is still assigned to subjects"`)
- `internal/domain/rbac-permissions/repository/postgres_rbac.go` 의 dup sentinel 제거 + `store.Err*` 로 wrap 통일

## 5. Test Infrastructure 강화

### 5.1 Cross-pkg fixture race 차단 (#439)

`go test ./...` 가 패키지 단위로 별도 binary 를 병렬 실행 → `internal/store_test` 와 `internal/domain/application-lifecycle/repository_test` 의 `setupApplicationsTest` 가 동일 테이블 (`applications`, `projects`, `project_integrations`, `application_repositories`, `project_members`, `repositories`) 을 TRUNCATE CASCADE 동시 실행 → A 패키지 test 진행 중 B 패키지 fixture 가 끼어들면 row 가 사라져 silent fail 회귀.

직전 sprint **#437** 의 SQLSTATE 22P02 ("uuid 잘못된 입력") 회귀가 본 race 가 silent err 무시 패턴을 통과한 결과.

**Fix**: 양쪽 `setupApplicationsTest` 에 동일 `applicationsFixtureLockID = int64(0x4150705F4C6966)` ("App_Lif" ASCII). dedicated conn 으로 `pg_advisory_lock(<id>)` session-level 잡고 fixture + test lifecycle 진행, teardown 에서 release. cross-pkg binary 가 같은 lock id 시도 시 block → 자연 시리얼라이즈.

**검증**: 두 패키지 동시 실행 `-count=3 -parallel 8` stress PASS (~37s / ~31s).

### 5.2 Silent err 무시 패턴 정정 (#437)

`integrations_integration_test.go` 의 5 test 가 모두 `app, _ := CreateApplication(...)` + `created, _ := CreateIntegration(...)` 패턴으로 err 무시. cross-pkg race 시 `app.ID=""` 가 `DeleteIntegration("")` 까지 흘러 SQLSTATE 22P02 로 표면화.

5 test 모두 `err :=` + `t.Fatalf("seed ...: %v", err)` 명시 → 회귀 시 root cause 메시지 즉시 표시.

### 5.3 Frontend vitest coverage exclude 정정 (#441)

- `lib/config/` (import 0건 dead code) 2 file 제거
- `vitest.config.ts` coverage.exclude 에 `lib/__mocks__/**` 추가 (react-dom/test-utils redirect 전용 test infrastructure)
- `vitest.config.ts` coverage.exclude 에 `lib/archive/**` 추가 (`ENABLE_LEGACY_MOCK_UI=false` 로 lock 된 dead path)

## 6. 핵심 산출

### 6.1 신규 test 분포

| 영역 | 신규 test 수 | 신규 LoC |
|---|---|---|
| backend B-3 9 도메인 view shim (`internal/domain/*/view/handler_test.go`) | 142 | ~3170 |
| backend application-lifecycle view endpoint (`applications_handler_test.go` + `projects_handler_test.go` + `fake_store_test.go`) | 178 | +2862 |
| frontend domain component (`GardenerFeed`, `OrgTree`, `OrgUnitTable`) | 42 | +825 |
| **합계** | **362** | **~6857** |

### 6.2 Production code 변경

| PR | 변경 |
|---|---|
| #438 | `internal/store/options.go` 1 line (message 일치) + `internal/domain/rbac-permissions/repository/postgres_rbac.go` 5 line (sentinel wrap 정정), 합 6 line |
| 다른 7 PR | 0 line |

## 7. 잔여 Carve Out 후보

본 sprint 의 종결조건 충족 후 미진입 영역:

| 후보 | 영역 | 추정 효과 |
|---|---|---|
| **ApplicationDashboard 50% 잔여** | backend application-lifecycle view | self-cover 90.2% → 95%+, `build_runs` / `quality_snapshots` fixture seeding 보강 |
| **OrgTree 75.27% 잔여** | frontend organization component | Focus Selection / onConnect / handleNodesChange position 동기화 분기 |
| **internal/store 20.2% 잔여** | backend store integration test | 27 메서드 `ApplicationStore` 의 미커버 method |
| **audit-ops/service integration test 신규** | backend domain | `TestIntegration_*` 매치 0건, 신규 carve |
| **운영 UI 전환 마무리** | frontend | `ENABLE_LEGACY_MOCK_UI` flag + `lib/archive/` 자체 제거 (정책 결정 필요) |
| **ui-foundation 미세 분기 보강** | frontend shared | Sidebar 83% → 90%+, AuthGuard 87% → 95%+ |
| **`integrations_integration_test.go:145` uuid empty FAIL** | backend store integration | 본 sprint 의 silent err 정정 (#437) + fixture race lock (#439) 으로 회귀 검출 가능, 다만 fail 자체는 잔존 가능성 (root cause 별도 조사) |

## 8. 학습 / 회고

1. **silent err 무시 패턴의 위험성**: `_, _ := f()` 가 명시 fail 대신 silent corruption 으로 흘러 root cause 진단을 어렵게 함. test 의 silent err 무시 는 production-grade 코드와 동일 기준으로 정정 권장.
2. **sentinel duplicate 의 silent production bug 위험**: `var ErrXxx = errors.New("same message")` 가 두 패키지에 각각 정의되면 `errors.Is` 매치 silent 실패. **단일 source 의 sentinel + import 통일** 이 정공법. message 비교는 fragile.
3. **cross-pkg fixture race 의 진단 어려움**: 단독 실행 PASS + 전체 실행 FAIL 패턴. PostgreSQL advisory lock 이 가장 가벼운 시리얼라이즈 수단. session-level lock 으로 lifecycle 보장.
4. **integration test cover 의 silent production bug 발견 효과**: 단위테스트 cover 가 일정 수준 도달 후의 잔여 carve 는 **integration cover 가 silent 회귀를 발견하는 catch-net 역할**. 본 sprint 의 RBAC 422 매핑 회귀 발견 case.
5. **sub-agent stall 회복**: sub-agent 가 `no progress 600s` 로 stall 시 working tree 의 partial 작업이 충분히 가치 있을 수 있음. stop 후 commit 가능 여부 검증 → 본 sprint 의 #442 case (sub-agent stall 했지만 42 test 완료 상태였음).
6. **branch race 의 working tree 흡수 패턴**: background sub-agent 가 working tree 를 자기 brunch 로 switch 하면 다른 brunch 에서 진행하던 작업이 sub-agent 의 brunch 에 commit. cherry-pick + force-push 로 분리. 본 sprint 의 #437 / #439 case.

## 9. 추적성 영향

- **IMPL**: #438 의 1건 (사용자 facing API 매핑 회귀 hotfix).
- **UT**: 362 신규 (회귀 가드 강화).
- **TC**: N/A (단위테스트 중심).
- **REQ / ARCH / API**: N/A (production code 무변경 + API contract 의 422 매핑은 기존 정의 그대로).
- 매트릭스 row 갱신 N/A (cover 보강은 IMPL 행에 영향 없음).

## 10. 변경 이력

- 2026-05-29: 초안 작성 (8 PR 머지 완결 직후).
