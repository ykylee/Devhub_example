# 2026-06-15 PR #598 CI 복구 결과 및 레슨런

- 문서 목적: PR #598 CI 실패의 실제 원인, 수정 조치, 검증 결과, 재사용 가능한 대응 절차를 기록한다.
- 범위: `actions/setup-go` cache restore 실패, 최신 `main` 리베이스 후 드러난 회귀, 로컬/PR 검증 결과, 후속 운영 규칙.
- 대상 독자: DevHub contributor, 리뷰어, CI 운영 담당자, AI agent.
- 상태: active
- 최종 수정일: 2026-06-15
- Tier: 공용
- 관련 문서: [2026-06-13 CI 재구성 결과 및 레슨런](./2026-06-13-ci-rearchitecture-retrospective.md), [Planning Documentation](./README.md), [문서 작성·관리 표준](../governance/document-standards.md)

## 1. 요약

PR #598의 CI 실패는 한 가지 원인만 있는 상태가 아니었다.

- 최근 PR 공통 실패의 직접 원인은 `actions/setup-go@v5` implicit cache가 저장소 루트에서 `go.sum`을 찾다가 실패하는 workflow 설정 문제였다.
- 이 문제를 고친 뒤 PR #598을 최신 `main`에 리베이스하자, base branch에 이미 들어와 있던 다른 회귀가 새로 드러났다.
- 최종적으로는 workflow 수정 1건과 최신 `main` 회귀 수정 여러 건을 함께 정리한 뒤 PR이 다시 성공 상태로 회복되었다.

즉, 이번 사례는 "PR 체크가 실패했다"는 하나의 증상 뒤에 **workflow 설정 결함 + 최신 base regression**이 겹쳐 있을 수 있음을 보여준다.

## 2. 실제 원인

### 2.1 1차 원인: `setup-go` implicit cache와 모듈 경로 불일치

최근 실패 PR들(#595, #596, #597 포함)은 공통적으로 아래 증상으로 초기에 중단됐다.

- `Restore cache failed: Dependencies file is not found ... go.sum`

원인은 `actions/setup-go@v5`가 기본 cache 동작을 통해 저장소 루트 기준 `go.sum`을 찾으려 했기 때문이다. 하지만 이 저장소의 Go module은 루트가 아니라 `backend-core/go.sum`을 사용한다.

기존 workflow에는 이미 explicit `actions/cache@v4`가 있었기 때문에, `setup-go`의 implicit cache는 중복이면서도 잘못된 경로를 바라보는 상태였다.

### 2.2 2차 원인: 최신 `main` 리베이스 후 드러난 base regression

PR #598은 처음에 stale local `main` 기반에서 만들어졌다. 이후 `origin/main`을 fetch해 보니 base branch가 이미 더 앞서 있었고, 새 SHA 기준으로는 다음 회귀가 드러났다.

- backend fake store가 새 인터페이스(`CountOpenAndMergedPRs`)를 따라가지 못함
- RBAC route permission table에 신규 admin integration endpoint 3개가 누락됨
- frontend integration provider type가 중복 선언됨
- repository-integration 서비스가 현재 `apiClient<T>(method, path)` 시그니처와 어긋남
- KPI window 타입이 component/service 사이에서 느슨하게 연결됨
- provider type 확장 후 `ProviderTable` badge variant mapping이 exhaustiveness를 만족하지 못함
- `httpapi` 테스트용 `memoryPlatformStore`가 최신 `IntegrationStore` 인터페이스 전체를 구현하지 못해 `integration store is not configured` 503으로 다수 테스트가 붕괴함
- 최신 UI 문자열/DOM 구조와 오래된 frontend test assertion이 어긋남

핵심은, **workflow fix만으로는 새 base 위에서 보이는 실패가 끝나지 않았다**는 점이다.

## 3. 무엇을 수정했는가

### 3.1 workflow 조치

다음 workflow의 `actions/setup-go@v5`에 `cache: false`를 명시했다.

- `.github/workflows/ci.yml`
- `.github/workflows/e2e-regression.yml`
- `.github/workflows/e2e-quarantine.yml`

의도는 명확했다.

- Go dependency cache는 기존 explicit cache 단계만 사용한다.
- `setup-go` implicit cache가 루트 `go.sum`을 찾다가 실패하는 경로를 제거한다.

### 3.2 backend 조치

다음 회귀를 함께 정리했다.

- `backend-core/internal/domain/application-lifecycle/view/fake_store_test.go`
  - `CountOpenAndMergedPRs` fake 구현 추가
- `backend-core/internal/domain/rbac-permissions/view/permissions.go`
  - admin integration sync/summary GET endpoint 3개를 permission table에 추가
- `backend-core/internal/httpapi/applications_test.go`
  - `memoryPlatformStore`에 integration sync job fake 저장소/조회/count 구현 추가
  - `newPlatformsRouter()`가 테스트 fake를 `IntegrationStore`로 주입하도록 보강
  - compile-time guard `var _ IntegrationStore = (*memoryPlatformStore)(nil)` 추가

이 조합으로 `httpapi` 대량 실패의 공통 원인이던 `integration store is not configured` 503이 해소됐다.

### 3.3 frontend 조치

다음 정합성 문제를 함께 수정했다.

- `frontend/domain/integration-registry/schema/integration.types.ts`
  - `IntegrationProviderType` 중복 선언 제거, 확장 union을 단일 source-of-truth로 통합
- `frontend/domain/integration-registry/view/ProviderTable.tsx`
  - `task_tracker`, `other`에 대한 badge variant 매핑 추가
- `frontend/domain/repository-integration/service/repository-kpi.service.ts`
  - `apiClient.get(...)` → `apiClient("GET", ...)`
- `frontend/domain/repository-integration/service/repository-tests.service.ts`
  - 동일 수정
- `frontend/domain/repository-integration/view/RepositoryKPISection.tsx`
  - `KPIWindowDays` 타입을 state / handler / selector에 일관되게 적용
- `frontend/components/admin/inbound-source-config/inbound-source-config.test.tsx`
  - 최신 안내 문구 및 split text node에 맞게 assertion 보정
- `frontend/components/admin/x1-widgets/x1-widgets.test.tsx`
  - 중복 텍스트 매칭 대신 heading role 기반 assertion으로 보정

## 4. 검증 결과

최종 수정 후 아래 검증을 통과했다.

- `bash ./scripts/ci-e2e-sync-check.sh .github/workflows/ci.yml`
- `bash ./scripts/ci-e2e-sync-check.sh .github/workflows/e2e-regression.yml`
- `bash ./scripts/ci-e2e-sync-check.sh .github/workflows/e2e-quarantine.yml`
- `frontend`: `npm test`
- `frontend`: `npm run build`
- `backend-core`: `go test ./internal/httpapi`
- `backend-core`: `go test ./...`

PR 기준으로는 다음 순서로 회복됐다.

1. workflow cache restore failure 제거
2. 최신 `main` 리베이스
3. 리베이스 후 드러난 최신 base regression 정리
4. 새 SHA `8f274ac5`를 PR #598에 push
5. GitHub checks 재실행 후 성공 확인

## 5. 레슨런

### 5.1 stale local `main`에서 만든 PR은 분석을 왜곡할 수 있다

처음에는 "내 수정이 아직 부족한가?"처럼 보였지만, 실제로는 오래된 base 위에서 브랜치를 만든 탓에 최신 `main`의 회귀가 가려져 있었다.

운영상 먼저 확인할 것:

- `origin/main`과 현재 브랜치의 공통 조상 시점
- PR 생성 시점 이후 base branch에 머지된 변경
- 새 SHA 기준에서 체크가 정말 같은 failure class인지 여부

### 5.2 workflow fix와 base regression fix를 분리해서 봐야 한다

이번 사례처럼 첫 실패 원인은 workflow였지만, 그것만 고치면 끝나지 않을 수 있다.

- 1차: workflow / runner / cache / path 문제
- 2차: 최신 base branch의 실제 코드 회귀

이 둘을 섞어 해석하면 대응 순서가 흔들린다. 이번에는 **cache 문제를 먼저 제거한 뒤** 최신 base 회귀를 별도로 정리한 것이 효과적이었다.

### 5.3 테스트 fake store는 interface drift의 빠른 경보 역할을 한다

`memoryPlatformStore`가 새 `IntegrationStore` 메서드 집합을 다 구현하지 못하면서 503이 대량으로 발생했다. 이건 불편하지만 좋은 신호이기도 하다.

- production wiring과 테스트 wiring의 계약 차이를 빨리 드러낸다.
- compile-time guard를 붙여두면 다음 drift를 더 이르게 잡을 수 있다.

따라서 테스트 fake는 임시 우회보다 **실제 interface를 완전히 구현하는 방향**이 장기적으로 안전하다.

### 5.4 frontend build와 test는 failure class가 다르다

이번에는 `npm test`와 `npm run build`가 서로 다른 문제를 잡아냈다.

- test는 UI 문구 / DOM 구조 회귀를 발견했다.
- build는 타입 중복과 exhaustiveness 문제를 발견했다.

즉, PR 회복 시에는 둘 중 하나만 green인 상태로 판단하면 부족하다.

## 6. 다음번 동일 증상 대응 절차

비슷한 PR CI 실패가 다시 나오면 아래 순서가 안전하다.

1. 여러 PR에서 같은 early failure가 반복되는지 먼저 본다.
2. 공통 증상이면 workflow / action / cache / path 같은 infra 계층부터 확인한다.
3. PR 브랜치가 최신 `origin/main` 기반인지 확인한다.
4. workflow 수정 후에는 PR 브랜치에 새 SHA를 만들어 다시 태운다.
5. 리베이스 후 새로 드러난 실패는 "원래 문제의 잔재"가 아니라 최신 base regression일 수 있음을 전제로 다시 분류한다.
6. frontend는 `npm test`와 `npm run build`를 둘 다 본다.
7. backend는 representative package 확인 후 `go test ./...`로 전체 수렴을 본다.

## 7. 결론

PR #598 복구는 단순한 rerun 성공이 아니라, 다음 운영 규칙을 확인한 사례였다.

- 공통 조기 실패는 workflow 계층에서 먼저 잡아야 한다.
- base branch가 앞서 있으면 PR 실패 해석이 달라질 수 있다.
- 최신 `main` 리베이스 뒤 드러난 회귀는 별도 정리 대상으로 봐야 한다.
- fake/test harness도 production interface 변화에 맞춰 유지해야 한다.

앞으로는 최근 PR들이 한꺼번에 비슷하게 실패할 때, **workflow 공통 원인 확인 → 최신 `main` 정렬 → 새 base regression 분리 대응**을 기본 복구 패턴으로 삼는 편이 좋다.
