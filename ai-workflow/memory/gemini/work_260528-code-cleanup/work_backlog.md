# Work Backlog

## 1. 현재 Sprint 백로그 상태
*   **남아있는 Task**: 없음 (All Completed)
*   **성공 조건 달성도**: 100%

## 2. 최근 완료 (2026-05-28)

### integration-registry 도메인 단위테스트 (28 tests)
| 파일 | 테스트 수 | 검증 내용 |
|------|:--------:|-----------|
| `infra.service.test.ts` | 10 | getMetrics/getNodes/getTopology/getTopologyV2/controlService |
| `ProviderTable.test.tsx` | 11 | empty/렌더링/Sync/Edit/Delete/syncing/deleting/Import/NewRepo |
| `BindingsTable.test.tsx` | 5 | empty/렌더링/ActionMenu edit-delete/provider fallback |

### 환경 설정 변경
- `vitest.config.ts` include 패턴: `domain/**/*`, `shared/**/*` 추가
- vitest 환경: `jsdom` → `happy-dom` 전환 (React 19 createRoot 호환)
- `frontend/scripts/postinstall.js`: React 19 `act` polyfill (react.production.js에 flushSync 기반 act 추가)
- `frontend/lib/test-setup.ts`: framer-motion jsdom mock 전역 설정

## 3. 향후 권장 과제
*   **PR Title Lint CI 도입** — code-taxonomy.md prefix 강제
*   **P0 기술 부채 리팩토링** — applications.go, users_units.go 분할
*   **도메인 단위테스트 확장** — service 계층 (rbac, application-lifecycle, realtime 등)
