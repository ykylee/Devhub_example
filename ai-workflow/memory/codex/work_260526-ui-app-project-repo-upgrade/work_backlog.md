# Work Backlog — codex/work_260526-ui-app-project-repo-upgrade

- 상태: in_progress

## Active

1. UI 운영 고도화 2.3
- 상세 정보 카드 mock 텍스트 실데이터화 후보 정리
2. docker `/devhub` E2E 확장
- repositories/detail 관련 선택 시나리오 추가 검증
- local-idp compose 스택 재기동 절차 정리

## Done

1. 2.1 상세 페이지 mock 제거
2. 2.2 테이블 액션 no-op 제거
3. 2.3 목록 3페이지 PageState 적용
4. 2.3 상세 3페이지 PageState + 오류 메시지 표준화 적용
5. 상세 카드 정적 텍스트 일부 실데이터화
6. `/devhub` docker E2E 환경 구성 + 선택 스펙 3종 통과

## Validation

1. `cd frontend && npm run lint` (pass)
2. `cd frontend && npm run build` (pass)
3. `cd frontend && PLAYWRIGHT_BASE_URL=http://localhost:13000 PLAYWRIGHT_BASE_PATH=/devhub ... npm run e2e -- tests/e2e/admin-applications.spec.ts` (pass)
4. `cd frontend && PLAYWRIGHT_BASE_URL=http://localhost:13000 PLAYWRIGHT_BASE_PATH=/devhub ... npm run e2e -- tests/e2e/admin-projects.spec.ts tests/e2e/project-model-modes.spec.ts` (pass)
