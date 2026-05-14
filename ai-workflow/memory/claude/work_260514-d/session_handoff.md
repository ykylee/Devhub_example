# Session Handoff — claude/work_260514-d (hotfix)

- 브랜치: `claude/work_260514-d`
- Base: `main` @ `f11bdbb` (PR #107 머지 직후)
- 날짜: 2026-05-14
- 상태: in_progress
- 직전 sprint: `claude/work_260514-c` (PR #107) — API-51..58 활성화. codex 외부 리뷰가 본 hotfix 의 2건 발견.

## Scope (hotfix — 2건)

### P1. ComputeApplicationRollup 의 custom weight fallback 정규화 부재
- 위치: `internal/store/repository_ops.go:341-365` (대략)
- 문제: `WeightPolicyCustom` 의 검증은 합 1.0 ±0.001 강제. 그러나 누락 repo 에 `1/n` equal fallback 할당 후 정규화 미실행 → 최종 합 > 1.0 → `build_success_rate` 등 weighted metrics 가 1.0 초과 (수치 오염).
- 정정: fallback 후 전체 가중치를 `totalRaw` 로 나눠 정규화. `fallbacks` meta 의 `AppliedWeight` 도 정규화 후 값으로 갱신.

### P2. UpdateIntegration unique violation 매핑 부재
- 위치: `internal/store/integrations.go:142-167` (`UpdateIntegration` 함수)
- 문제: external_key 변경이 partial UNIQUE 인덱스 위반 가능. 현재는 `pgx.ErrNoRows` 만 분기, unique violation 은 generic 500 으로 노출.
- 정정: `isUniqueViolation(err)` 체크 추가 → `ErrConflict` 매핑. handler `updateIntegration` 에 409 분기 추가. `memoryApplicationStore.UpdateIntegration` 에 unique 검증 추가.

## 작업 순서
1. P1 store + UT
2. P2 store + handler + memoryStore + UT
3. test + commit + push + CI + squash merge

## 다음 세션 인계
- 본 hotfix 후 다음 작업 후보는 work_260514-c carve_out 참조 (frontend UI / postgres integration test / ADR-0011 §4.2 등).
