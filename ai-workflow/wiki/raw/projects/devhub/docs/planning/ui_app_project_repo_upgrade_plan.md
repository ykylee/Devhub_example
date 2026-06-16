# UI 고도화 실행 계획 (Application / Project / Repository)

- 문서 목적: 운영 수준 UI 고도화를 위한 단계별 실행 계획을 정의한다.
- 범위: Application / Project / Repository 관리·조회 UI
- 대상 독자: Frontend, Backend, QA, 운영 담당자
- 상태: in_progress
- 최종 수정일: 2026-05-26
- 후속 E2E 정리: `docs/planning/ui_e2e_followup_after_merge.md`

## 1. 목표

- mock/임시 데이터 의존 제거
- 실제 API 기반의 안정적인 조회/관리 UX 제공
- 운영 관점의 오류/빈 상태/로딩 상태 표준화

## 2. 단계

### 2.1 1차 (즉시) — 상세 페이지 mock 제거

- Platform 상세 페이지
  - mock history 차트 제거
  - 실제 rollup / 연결 repository 데이터만 사용
- Repository 상세 페이지
  - mock activity/security 데이터 제거
  - 실제 activity 지표(`pr_event_count`, `build_run_count`, `build_success_rate`, `active_contributors`) 중심으로 재구성
 - 진행 상태: **완료 (2026-05-26)**

### 2.2 2차 — 관리 액션 완결

- Platform/Project/Repository 테이블의 no-op 액션 실제 연결
- 수정/조회/연결 해제 등 사용자 플로우 정합
 - 진행 상태: **완료 (2026-05-26)**

### 2.3 3차 — 운영 UX 표준화

- 오류/빈 상태/권한 제한/재시도 UI 공통화
- 운영 메시지 문구/로그 컨벤션 정리
 - 진행 상태: **in_progress (목록+상세 공통 상태 UI 적용 + `/devhub` docker E2E 검증 완료, 2026-05-26)**

## 3. 검증

- `cd frontend && npm run lint`
- `cd frontend && npm run build`
- 관련 E2E 시나리오 선택 실행

## 4. 진행 이력

- 2026-05-26:
  - 2.1 완료: Application/Repository 상세의 mock 의존 제거
  - 2.2 완료: 테이블 액션 no-op 제거 및 실제 플로우 연결
  - 2.3 진행: 목록+상세 화면에 공통 `PageState`(loading/error/retry/empty) 적용 및 오류 문구 표준화
  - `/devhub` docker E2E 환경 구성: `localhost:13000/devhub` + local-idp + host postgres
  - 검증: `frontend lint/build` 통과, 선택 E2E(`admin-applications`, `admin-projects`, `project-model-modes`, `repositories-ui`, `repositories-ui-negative`, `repositories-detail-negative`, `applications-projects-detail-negative`) 통과
  - 남은 E2E 후속 범위는 `ui_e2e_followup_after_merge.md` 로 분리 정리

## 5. 현재 작업 범위

- 현재 작업 범위: **2.3 3차 (운영 UX 표준화 + E2E 범위 확장)**
