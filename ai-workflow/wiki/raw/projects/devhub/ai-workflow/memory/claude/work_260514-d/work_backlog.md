# Work Backlog — claude/work_260514-d (hotfix)

- 상태 관리: planned / in_progress / blocked / done

## A. P1 ComputeApplicationRollup custom weight 정규화

- [planned] custom fallback 후 sum=1.0 재정규화 (totalRaw 분모)
- [planned] fallbacks meta 의 AppliedWeight 정규화 후 값으로 갱신
- [planned] UT 회귀 guard: custom_weights={"team/a": 1.0} + contributing 에 "team/b" 존재 → 정규화 후 합 1.0 검증

## B. P2 UpdateIntegration unique violation

- [planned] store: isUniqueViolation(err) → ErrConflict
- [planned] handler: errors.Is(err, store.ErrConflict) → 409 integration_conflict
- [planned] memoryApplicationStore.UpdateIntegration: unique 검증 추가
- [planned] UT 회귀 guard: 2 integration 생성 후 1 의 external_key 를 다른 1 의 값으로 PATCH → 409

## C. 매트릭스 + 문서

- [planned] trace.md §6 변경 이력 row 추가
- [planned] PR body 추적성 영향 명시 (REQ/ARCH/API 변경 없음, IMPL fix)

## D. 머지

- [planned] commit + push + CI wait + squash merge
