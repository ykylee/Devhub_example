# 운영 UI 전환 계획 (Mock 정리)

- 작성일: 2026-05-26
- 대상 브랜치: `codex/work_260526-main-sync-new-task`
- 목적: 대시보드/상세 화면에서 단순 mock UI 노출을 중단하고, 운영 API 연동 중심으로 전환 준비

## 1) 분류 요약

### 즉시 운영 전환 가능
- KPI fallback mock (`infraService.getMetrics` 실패 시)
- Application/Project/Repository 상세 차트 mock 데이터
- OrgTree hierarchy fallback mock

### 단순 mock (즉시 비노출 대상)
- Developer: Active Stream, Deployment Pipeline(mock build logs), AI Gardener/Recognition/Infrastructure 위젯
- Manager: Velocity(mock), Talent Load Balancing(mock), Decision Audit(mock)
- Project Detail: Recent Activity(mock), Active Tasks(mock)

## 2) 이번 변경 범위
- `ENABLE_LEGACY_MOCK_UI=false` 기본값으로 단순 mock 섹션 렌더링 차단
- 단순 mock 데이터는 `frontend/lib/archive/mock-ui-legacy.ts`로 아카이브
- 운영 전환 가능 항목은 화면 구조를 유지하고 API 데이터 주입 준비 상태로 유지

## 3) 백엔드 연동 준비 체크리스트
- Dashboard
  - `GET /api/v1/dashboard/developer/stream` (active work stream)
  - `GET /api/v1/dashboard/developer/builds` (pipeline/build log)
  - `GET /api/v1/dashboard/manager/velocity` (quality/security 시계열)
  - `GET /api/v1/dashboard/manager/team-load` (인력 부하)
  - `GET /api/v1/dashboard/manager/decisions` (의사결정 이력)
- Project Detail
  - `GET /api/v1/projects/{id}/activity`
  - `GET /api/v1/projects/{id}/tasks?status=open,in_progress,review`

## 4) 운영 전환 원칙
- API 실패 시에도 임의 mock 데이터 자동 표출 금지
- 실패 상태는 `empty state + retry + 에러 토스트`로 통일
- mock 복원은 `ENABLE_LEGACY_MOCK_UI=true`로만 제한적으로 허용
