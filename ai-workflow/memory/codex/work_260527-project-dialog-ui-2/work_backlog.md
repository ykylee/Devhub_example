# Work Backlog — codex/work_260527-project-dialog-ui-2

- 상태: in_progress

## Active

1. PR #368 리뷰/CI 대응
2. 머지 후 다음 과제 브랜치 준비

## Done

1. repository draft 상태 모델/마이그레이션 추가
2. draft 생성/발행 API 구현
3. draft publish와 SCM 생성 연동
4. Admin Catalog repository UX 개선
5. backend 테스트 + frontend build 검증
6. PR #368 생성

## Validation

1. `cd backend-core && go test ./internal/httpapi ./internal/store` (pass)
2. `cd frontend && npm run -s build` (pass)

