# Mock UI Archive (2026-05-26)

## 아카이브 목적
- 운영 화면에서 제거한 단순 mock 블록의 원본 의미를 보존하고, 추후 고도화 시 재사용 근거로 활용.

## 아카이브 위치
- 코드 데이터: `frontend/lib/archive/mock-ui-legacy.ts`
- 전환 계획: `docs/planning/ops_ui_transition_plan.md`

## 아카이브된 단순 mock 항목
- Developer: Active Stream, Deployment Pipeline(mock build logs), AI Gardener/Recognition/Infrastructure
- Manager: Velocity(mock), Talent Load Balancing(mock), Decision Audit
- Project Detail: Recent Activity(mock), Active Tasks(mock)

## 복원 방법
- `frontend/lib/config/mock-ui.ts` 의 `ENABLE_LEGACY_MOCK_UI` 값을 `true`로 변경
- 운영 전환 테스트 목적 외 상시 복원 금지
