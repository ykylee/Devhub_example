# Work Backlog — codex/repository-create-dialog-refactor

- 상태: in_progress

## Active

1. PR #376 리뷰 대응
- 생성 UX 변경 회귀 확인
- Create Project 실사용 피드백 반영

2. 후속 안정화 점검
- project/repository 생성 플로우 e2e 추가 여부 판단
- websocket.service 완전 제거 영향 확인

## Done

1. Admin Catalog New Repository 모달 전환
2. ProjectCreationModal 리더 초기값/비활성 사유 보강
3. ComboBox Enter/ESC 상호작용 개선
4. AuthGuard 레거시 websocket.service 연결 제거
5. lint/build 검증 완료
6. PR #376 생성 완료

## Validation

1. `cd frontend && npm run lint` (pass, warnings only)
2. `cd frontend && npm run build` (pass)
