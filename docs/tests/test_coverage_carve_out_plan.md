# 단위테스트 코드 커버리지 carve out plan

- 문서 목적: 2026-05-29 sprint `claude/work_260529-{a..p}` 머지 후 backend/frontend 단위테스트 커버리지 현황 점검 + 잔여 carve out 후보 정리.
- 범위: backend-core (Go) + frontend (Vitest) — service / view / repository / shared / infrastructure 영역.
- 대상 독자: 후속 sprint 담당 (Claude / Gemini / Codex), 프로젝트 리드.
- 상태: draft
- 최종 수정일: 2026-05-29
- 관련 문서: [governance/code-taxonomy.md](../governance/code-taxonomy.md), [governance/document-standards.md](../governance/document-standards.md), [domain/README.md](../domain/README.md), [traceability/report.md](../traceability/report.md)

## 1. 현황 측정 (main HEAD `a55bfb0`)

### 1.1 Frontend (`npx vitest run --coverage`)

| 지표 | 값 | 변화 |
|---|---|---|
| Statements | **61.75%** (2185/3538) | sprint 시작 28.03% → +33.72%p |
| Branches | 60.57% (1561/2577) | |
| Functions | 55.08% (563/1022) | |
| Lines | **63.43%** (1992/3140) | |

#### 도메인별 service 계층 — 거의 100% (sprint #407)

| 영역 | Lines % |
|---|---|
| `domain/rbac-permissions/service` (rbac.service) | 100% |
| `domain/integration-registry/service` (infra.service, integration-provider-presets) | 96~100% |
| `domain/repository-integration/service` (repository.service) | 96.66% |
| `domain/auth-session/service` (token-store/pkce/refresh-scheduler/session-death/refresh) | 88~100% |
| `domain/dev-request/service` | ~100% |
| `domain/audit-ops/service` + `domain/onboarding/service` | ~100% |
| `domain/realtime/service` (realtime + websocket) | 92~96% |

#### 도메인별 view 계층 — 평균 90%+ (PR #412 + #424)

| 영역 | Lines % |
|---|---|
| `domain/platform-lifecycle/view` (ApplicationCreationModal 96% / ProjectCreationModal 92% / ApplicationTable 89% / ProjectTable 92%) | **94%** (PR #424) |
| `domain/integration-registry/view` (Provider/Binding modal/table) | 82~96% |
| `domain/rbac-permissions/view` (PermissionMatrix/Editor) | 85~92% |
| `domain/repository-integration/view` (RepositoryTable/Link) | 95% |
| `domain/dev-request/view` (DevRequest detail/table/widget + IntakeToken) | 91~92% |
| `domain/organization-management/view` (MemberTable/MemberManagementModal/OrgNode/OrgUnitGrid/UserCreationModal) | 91~92% |

#### 잔여 carve out 영역 — Lines 0~50%

| 영역 | Lines % | 분류 |
|---|---|---|
| `shared/ui-foundation/components/Modal` | 0% | A |
| `shared/ui-foundation/components/Toast` | 0% | A |
| `shared/ui-foundation/components/PageState` | 0% | A |
| `shared/ui-foundation/components/FilterBar` | 0% | A |
| `shared/ui-foundation/components/DashboardHeader` | 0% | A |
| `shared/ui-foundation/components/DestructiveConfirmModal` | 0% | A |
| `shared/ui-foundation/components/LogoutOverlay` | 0% | A |
| `shared/ui-foundation/components/ActionMenu` | 64% | A (보강) |
| `shared/ui-foundation/components/ComboBox` | 60% | A (보강) |
| `shared/ui-foundation/layout/Header` | 0% | A |
| `shared/ui-foundation/layout/Sidebar` | 0% | A |
| `shared/ui-foundation/layout/AuthGuard` | 87% | (충분) |
| `lib/services/dashboard.service` | 0% | B (legacy 이관 미완) |
| `lib/services/gardener.service` | 0% | B |
| `lib/services/integration.service` | 0% | B |
| `lib/services/risk.service` | 0% | B |
| `lib/services/error-message` | 0% | B |
| `lib/services/api-client` | 12.5% | B (cross-cutting) |
| `lib/services/dev_request_token.service` | 0% | B |
| `lib/storage/onboardingSkip` | 0% | B (legacy) |
| `lib/archive/mock-ui-legacy` | 0% | C (dead) |

### 1.2 Backend (`go test ./... -count=1 -short -cover`)

#### 도달 영역

| 패키지 | Coverage |
|---|---|
| `internal/shared/integrationcaps` | **100%** (PR #409 신규) |
| `internal/domain/dev-request/service` | 88.7% |
| `internal/infrastructure/serviceaction` | 86.4% |
| `internal/infrastructure/gitea` | 76.7% |
| `internal/normalize` | 68.5% |
| `internal/httpapi` (router + cross-domain test) | 63.5% |
| `internal/integrations/adapters` | 61.9% |
| `internal/infrastructure/commandworker` | 55.2% |
| `internal/shared/config` | 37.1% |
| `internal/domain/onboarding/view` | 27.5% |
| `internal/domain/realtime/view` | 21.7% |
| `internal/domain/rbac-permissions/view` | 10.5% |
| `internal/infrastructure/hrdb` | 10.5% |

#### 0% 영역 — 측정 한계 vs 실제 미커버

**측정 한계** (env 미설정 skip + 패키지 외 test 위치):
- `internal/store` 0% — `*_integration_test.go` 가 `DEVHUB_TEST_DB_URL` env 미설정 시 `t.Skip` (testing.Short 와 무관). 실제는 PR #109/#110 의 23 integration test (PG 15 native) 가 cover. CI 의 `backend-integration` job (env 주입 + PG 5432 native) 으로 검증 (현재 임시 disable, #419).
- `internal/domain/<도메인>/{view,repository}` 대부분 0% — 도메인 view test 가 `internal/httpapi/` 에 위치 (Phase 3 의 view cross-package import 회피 결과 sprint -f). 실 cover 는 `httpapi` 63.5% 에 포함.

**실제 미커버**:
- `internal/shared/httphelp` 0% — test 자체 없음 (PR #407 신규). 실제 exported API: `ParseBoundedInt` + `WriteServerError` (errors.go) + `GenerateRequestID` + `ValidateCallerRequestID` + `RequireRequestID` + `RequestIDFrom` + `RequestIDFromContext` + `SourceTypeFrom` + `ClientIPFrom` + `LogRequest` (request_context.go).
- `internal/domain/<도메인>/service` 일부 — auth-session/service / audit-ops/service / organization-management/repository 등은 production code 자체가 thin (interface 정의 + thin pass-through) 이라 실 test 작성 후보 한정.

#### `-coverpkg=./...` 재측정 (B-2 완료, sprint `claude/work_260529-t`)

`go test ./... -count=1 -short -coverpkg=./... -coverprofile=/tmp/cover.out`:
- **Backend total = 43.0%** (cross-package call cover 포함)
- 패키지별 self-coverage 와 -coverpkg 결과 차이: -coverpkg 는 각 패키지 코드를 어디서든 호출됐는지 기준 — 도메인 view 패키지가 httpapi test 에서 호출되면 cover 측정.
- 도메인 source level 분포 (`go tool cover -func` sample): `domain/dev_request.go::IsActive` 100% / `IsValidDevRequestTransition` 75% / `domain/rbac.go::ValidateRoleID` 100% / `EnforceAuditInvariant` 100% / `Allows` 77.8% / `DefaultPermissionMatrix` 100% / `domain/primary_unit.go::ResolvePrimaryUnit` 100% / `domain/application.go::ResolveOutboundAuth` 72.7% / `IsRetryableSyncError` 66.7% / `domain/domain.go::IsTerminal` 100% / `CanTransitionTo` 100%.

상세 패키지별 분석 (도메인 view per-package 도달률) 은 후속 carve out.

#### B-1 완료 — `internal/shared/httphelp` test 신규

sprint `claude/work_260529-t` 진행: `errors_test.go` (4 test) + `request_context_test.go` (10 test) = 14 test 신규, **coverage 98.6%** (잔존 1.4% 는 `GenerateRequestID` 의 `rand.Read` 실패 분기 — 실행 불가).

#### B-5 (부분) 완료 — store integration test 실 DB 실행 (sprint `claude/work_260529-aa`)

`DEVHUB_TEST_DB_URL=postgres://postgres:postgres@localhost:5432/devhub_test` + 46 migration 적용 후 실 PostgreSQL 18 환경에서 integration test 실행:

| 패키지 | -short cover | 실 DB cover |
|---|---|---|
| `internal/store` | 0% (env skip) | **20.2%** |
| `internal/domain/platform-lifecycle/repository` | 0% | **43.6%** |
| `internal/domain/dev-request/repository` | 0% | **24.1%** |
| Backend total (`-coverpkg=./...`) | 43.0% | **54.4%** (+11.4%p) |

CI `backend-integration` job 복원 시 동일 상승 예상. 잔존:
- `integrations_integration_test.go:145` FAIL (uuid input "" → SQLSTATE 22P02) — 별도 hotfix 후보, 본 PR scope 외.
- `internal/domain/audit-ops/service` integration test 없음 (TestIntegration_* 매치 0건).

## 2. 잔여 carve out 후보

### 2.1 P1 — 즉시 진행 가능

#### F-1. shared/ui-foundation components/layout test 보강

- **scope**: Modal / Toast / PageState / FilterBar / DashboardHeader / DestructiveConfirmModal / LogoutOverlay / Header / Sidebar (9 component, 모두 0%) + ActionMenu (64% → 90%) + ComboBox (60% → 90%) 보강.
- **이유**: Shared 레이어 (code-taxonomy §2.2) 가 도메인 비결합 공통 UI 라 회귀 위험 큼. 모든 도메인 페이지가 import.
- **예상 작업**: ~15 test file (component 별 ~12 test), 추정 +120 test, +2000 LoC.
- **기대**: frontend overall coverage 63.43% → **75%+** 예상.

#### B-1. backend `internal/shared/httphelp` test 신규

- **scope**: `errors.go` (`ParseBoundedInt`, `WriteServerError`) + `request_context.go` (`GenerateRequestID`, `ValidateCallerRequestID`, `RequireRequestID`, `RequestIDFrom`, `RequestIDFromContext`, `SourceTypeFrom`, `ClientIPFrom`, `LogRequest`).
- **이유**: Shared 레이어 helper 가 모든 handler 에서 호출. 회귀 risk + 100% 달성 용이 (작은 file).
- **예상 작업**: 1 test file, ~15 test, ~150 LoC.
- **기대**: 0% → **100%**.

#### B-2. backend `-coverpkg=./...` 재측정 + 도메인 view 실 coverage 명문화

- **scope**: `go test ./... -coverpkg=./... -coverprofile=cover.out` + cover report 분석. 도메인 view 패키지의 실 cover 확정 (현재 0% 표기는 측정 한계).
- **이유**: 현 0% 표기 패키지 중 다수가 httpapi cross-package test 로 실제 cover 됨. 정확한 수치 필요.
- **예상 작업**: 1 measurement run + 분석 문서 1 file (~100 LoC).
- **기대**: 실 backend coverage 명확화 — 보강 대상 우선순위 정렬.

### 2.2 P2 — 영역 정리 동반

#### F-2. lib/services legacy 도메인 이관 + test

- **scope**: `lib/services/dashboard.service` / `gardener.service` / `integration.service` / `risk.service` / `error-message` / `api-client` / `dev_request_token.service` / `wire.ts` 를 적절한 `domain/<도메인>/service/` 또는 `shared/` 로 이관 후 test.
- **이유**: code-taxonomy SoT 의 4 계층 매핑 정합 + 0% coverage 해소.
- **예상 작업**: 이관 8 file + test 8 file, ~+150 test, +1500 LoC.
- **기대**: frontend overall +5%p coverage.

#### B-3. backend domain/<도메인>/view 자체 unit test (handler.go shim)

- **scope**: 10 도메인의 `view/handler.go` 가 config wire + DI 만 — 단순 helper test. 실 핸들러 cover 는 httpapi test 가 담당.
- **이유**: 패키지별 self-coverage 0% 해소 (검증 가시성).
- **예상 작업**: 도메인별 1 test file × 10 = ~50 test, ~500 LoC.
- **기대**: 도메인 view 패키지 self-coverage 30%+ (그 외는 httpapi cover).

#### B-4. infrastructure 보강

- **scope**: hrdb (10.5% → 60%+, postgres adapter / mock branch), shared/config (37.1% → 80%+, env 분기).
- **이유**: 운영 영향 큰 영역 (HRDB integration + env 설정).
- **예상 작업**: 2 test file 보강 + 신규 + ~30 test.
- **기대**: 두 영역 60~80% 도달.

### 2.3 P3 — 별도 sprint 권장 (대규모)

#### F-3. lib/storage / lib/archive 정리

- **scope**: `lib/storage/onboardingSkip` 사용처 확인 후 도메인 이관 또는 dead 정리. `lib/archive/mock-ui-legacy` dead 정리 (이미 archive).
- **이유**: legacy + dead code 정리.

#### B-5. integration test CI 복원 + store 영역 cover

- **scope**: [issue #419](https://github.com/ykylee/Devhub_example/issues/419) — CI `backend-integration` job 복원. `internal/store/*_integration_test.go` (PR #109/#110 의 23 test) 실 실행.
- **이유**: store 영역 0% 해소 + 회귀 가드.

## 3. 우선순위 권장 작업 순서

| 순서 | carve | 분류 | 예상 시간 | 기대 효과 |
|---|---|---|---|---|
| 1 | F-1 | P1 | ~3 hours | frontend overall 63% → 75% |
| 2 | B-1 | P1 | ~30 min | shared/httphelp 100% |
| 3 | B-2 | P1 | ~30 min | backend coverage 가시성 |
| 4 | F-2 | P2 | ~4 hours | lib/services legacy 정리 + 5%p |
| 5 | B-3 | P2 | ~2 hours | 도메인 view self-coverage 가시성 |
| 6 | B-4 | P2 | ~2 hours | hrdb + config 60~80% |
| 7 | F-3 | P3 | ~1 hour | legacy 정리 |
| 8 | B-5 | P3 | 사내 결정 | store 영역 cover (CI 복원 #419 후) |

총 예상 — **~13 hours** (사내 작업 #419 제외).

## 4. 측정 method 표준화 권장

### Backend

```bash
cd backend-core
go test ./... -count=1 -short -coverpkg=./... -coverprofile=cover.out
go tool cover -func=cover.out | grep total
```

`-coverpkg=./...` 로 cross-package call 도 cover 측정 — 정확한 실 coverage.

### Frontend

```bash
cd frontend
npx vitest run --coverage --coverage.reporter=text-summary
```

기존 vitest.config.ts 의 coverage include 패턴 (`lib/**/*.ts`, `components/**/*.tsx`, `domain/**/*.ts(x)`, `shared/**/*.ts(x)`) 정합 확인.

## 5. 관련 issue

- [#419](https://github.com/ykylee/Devhub_example/issues/419) — CI e2e + backend-integration 복원 (P1, 사내)
- [#420](https://github.com/ykylee/Devhub_example/issues/420) — view 큰 modal 70%+ (closed PR #424)
- [#421](https://github.com/ykylee/Devhub_example/issues/421) — PlatformRepository decouple (P2, 진행 중)
- [#422](https://github.com/ykylee/Devhub_example/issues/422) — ApplicationStore slim (P2, 진행 중)
- [#423](https://github.com/ykylee/Devhub_example/issues/423) — traceability §2 인덱스 정합 (P3, 진행 중)

## 6. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-29 | 초안 작성 (sprint `claude/work_260529-p`). main HEAD `a55bfb0` 기준 backend + frontend 점검 + 8건 carve out 후보 정리. |
