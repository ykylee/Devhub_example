# Session Handoff — codex/work_260526-main-sync-new-task

- 문서 목적: 목업 UI 정리 및 운영 UI 전환 작업의 현재 상태를 다음 세션에 인계한다.
- 범위: mock 비노출 처리, 아카이브, 운영 API 연결 1차, 에러 처리 표준화
- 대상 독자: 메인 에이전트, 리뷰어, 다음 구현 세션 담당자
- 상태: in_progress
- 최종 수정일: 2026-05-26

## 이번 세션 핵심 결과

1. 단순 mock UI를 기본 비노출 처리 (`ENABLE_LEGACY_MOCK_UI=false`)
2. 기존 mock 데이터 아카이브 파일/문서화
3. 운영 전환 계획 문서 및 백엔드 연동 체크리스트 작성
4. Developer/Manager/Project 상세에 운영 API 기반 데이터 렌더링 1차 반영
5. `infraService.getMetrics` mock fallback 제거, API 실패 시 empty-state + retry 적용
6. 공통 사용자 메시지 유틸(`toUserErrorMessage`) 추가 및 대시보드/프로젝트 적용

## 현재 상태

- 프런트 lint 통과
- API 응답 스키마는 alias 매핑으로 완충 처리됨
- 일부 endpoint가 미구현/무응답이면 화면은 empty-state를 정상 표출

## 다음 세션 우선순위

1. 백엔드 실제 응답 샘플 기준으로 alias 축소/타입 고정
2. Playwright E2E에 retry/empty-state 시나리오 추가
3. 추적성(sync-checklist/report) 최종 동기화 점검
