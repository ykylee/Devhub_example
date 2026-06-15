# Work Backlog

## 1. 현재 Sprint 상태

- **남은 Task (본 PR)**: PR 발행 + 4단계 self-review + squash merge
- **15 commit (4 본 sprint 통합) ready to push**

## 2. 본 sprint 통합 결과 (2026-05-29)

### Recovery 작업 (4 commit)
- **Backend test recovery** (`4b4aaed`) — Gemini cleanup sprint 의 test compile 회귀 + production code 일부 누락 + cross-package import + providerHasCapability OR semantics 회귀 정리.
- **Frontend P0 ghost import + dead file 정리** (`896d8e7`) — project.service + integration-provider-presets path 갱신, framer-motion mock type, components/{ui,layout} 14 dead file 삭제.
- **Frontend service test +369** (`e6137f5`) — audit/onboarding/dev_request/auth/realtime 신규 + rbac 재작성 + 9 file 보강. project/pkce/audit/onboarding/dev_request/rbac 100% coverage.
- **CI 임시 정리** (`17d8459`) — e2e + backend-integration job 일시 비활성화 + 복원 SOP.

### 검증
- Backend 21 패키지 `go test -short` PASS
- Frontend 29 file 431 vitest PASS / tsc 0 errors / coverage 28.03%

## 3. 머지 후 후속 carve out

| 우선순위 | 항목 | 사유 / 위치 |
|---|---|---|
| **P1** | `providerHasCapability` 3 카피 통합 (`internal/shared/integrationcaps`) | Gemini split 시 한 카피가 OR→AND 회귀했던 case. drift 재발 위험. |
| **P1** | view 컴포넌트 24개 100% coverage | 5000+ LoC, 본 sprint scope 외. 별도 sprint. |
| **P1** | CI e2e + backend-integration job 복원 | `if: ... && false` 제거. cleanup 정리 끝났을 때. |
| **P2** | `ApplicationRepository` cross-domain decouple | `*IntegrationRepository` embed 제거. review agent P1. |
| **P2** | `ApplicationStore` interface slim | 13+ integration 메서드 제거. integration domain 으로 이관. review agent P1. |
| **P3** | shared/utils/last-build.ts + lifecycle-status.ts coverage 0% 원인 조사 | 같은 디렉터리 test 존재하지만 coverage 측정 시 0%. vitest config 또는 import path 이슈 가능. |
| **P3** | PR Title Lint CI 도입 | `code-taxonomy.md` §0 prefix 컨벤션 강제. Gemini 본 sprint state.json 의 권장 사항. |
| **P3** | P0 기술 부채 (applications.go 1172 / users_units.go 1263 LoC 분할) | code-taxonomy.md §3 P0-2, P0-3, P0-4. |
