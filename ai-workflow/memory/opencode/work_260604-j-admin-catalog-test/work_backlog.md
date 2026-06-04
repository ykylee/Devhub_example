# Work Backlog — opencode/work_260604-j-admin-catalog-test

- Branch: `opencode/work_260604-j-admin-catalog-test`
- Agent: opencode (Sisyphus)
- Updated: 2026-06-04

## 0. 본 sprint 의 목표

admin/catalog page.test.tsx 3차 시도 (pivot: 구조 중심 테스트).

## 1. 작업 단위 분해

| ID | 작업 | 상태 |
| --- | --- | --- |
| WB-01 | 브랜치 + memory set up | done |
| WB-02 | comprehensive mock setup (pivot) | done |
| WB-03 | 8 tests 작성 (structural focus) | done |
| WB-04 | vitest + tsc + lint 검증 | done |
| WB-05 | 메모리 갱신 | done |

## 2. pivot 전략

### 실패 분석 (sprint J 1/2차)
| 문제 | 영향 | pivot 해결 |
| --- | --- | --- |
| loadAll → 4 service → Promise.all race | test 간 상태 충돌 | structural test로 전환, 데이터는 단순화 |
| tab URL mock(setTab → router.replace) | 탭 전환 시 searchParams 변경 race | testid 기반 tab 존재 검증 |
| useToast ToastProvider context | 모든 test error | vi.mock 으로 전역 stub |

### mock 전략 (3차)
```
vi.mock("framer-motion")        → Proxy pattern
vi.mock("next/navigation")      → useSearchParams + useRouter
vi.mock("@/shared/.../Toast")   → useToast: () => ({ toast: vi.fn() })
vi.mock("@/domain/.../ApplicationCreationModal")  → null passthrough
vi.mock("@/shared/.../PageState")  → PageLoading/PageError/PageEmpty passthrough
vi.mock("@/domain/.../application.service")  → vi.hoisted mocks
vi.mock("@/domain/.../repository.service")   → vi.hoisted mocks
vi.mock("@/domain/.../project.service")      → vi.hoisted mocks
vi.mock("@/domain/.../identity.service")     → vi.hoisted mocks
```

## 3. 검증 결과

| 항목 | 결과 |
| --- | --- |
| vitest (신규 8 tests) | 8/8 PASS |
| vitest (전체) | 1015/1015 PASS (기존 1007 + 8, 회귀 0) |
| tsc | no new errors |
