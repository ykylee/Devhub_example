# Work Backlog — codex/work_260527-b-project-repo-flow

- 상태: in_progress

## Active

1. PR 최종 검토 및 게시
- diff sanity check
- commit/push
- PR draft 또는 ready 생성

2. 추적성 문서/PR 본문 동기화
- 영향 ID 갱신 필요 여부 확인
- PR 템플릿 "추적성 영향" 섹션 작성

## Done

1. Project standalone 생성 API 연결
2. Project repository optional 허용
3. Project detail N:M repository 연결 UX(+ 버튼) 구현
4. repository_create_payload 백엔드/프런트 반영
5. backend 단위 테스트 + frontend lint 검증
6. 브랜치 메모리 문서 초기화/갱신

## Validation

1. `cd backend-core && go test ./internal/httpapi ./internal/store` (pass)
2. `cd frontend && npm run lint` (pass, warnings only)
