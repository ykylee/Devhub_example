# Session Handoff — codex/work_260527-b-project-repo-flow

- 목적: Project/Repository 생성 흐름 2차 구현 및 PR 준비
- 상태: in_progress
- 최종 수정일: 2026-05-27

## 금일 진행

1. 프로젝트 생성 모델 정리
- `POST /api/v1/projects` standalone 생성 경로 사용
- `application_id` / `repository_id` / `repository_ids` optional 처리

2. 저장소 nullable 정합
- migration `000037_projects_repository_nullable` 추가
- `projects.repository_id` nullable 전환
- store scan/insert 경로에서 nullable 정합 적용

3. 프로젝트-저장소 N:M UX 개선
- 프로젝트 상세에서 연결된 저장소를 수평 표시
- `+` 버튼으로 다중 저장소 연결 가능하도록 모달 추가

4. 저장소 동반 생성(repository_create_payload) 구현
- `repository_create_payload: { key, slug, scm_provider }` 추가
- 적용 엔드포인트:
  - `POST /api/v1/projects`
  - `POST /api/v1/applications/:application_id/projects`
- backend store에 `CreateRepositoryForProject` 추가 후 project 생성 흐름과 연결
- frontend 생성 모달에 "Create and link repository on project creation" 입력 UX 추가

5. 테스트/검증
- backend: `go test ./internal/httpapi ./internal/store` 통과
- frontend: `npm run lint` 통과 (기존 warning 3건 유지)

## 남은 작업

1. 최종 코드 리뷰(회귀/명세 정합) 1회 수행
2. 커밋/푸시/PR 생성
3. PR 본문에 변경 범위/검증/리스크/추적성 영향 기재

## 비고

- 현재 저장소 동반 생성은 내부 `repositories` row 생성/업데이트 중심이며, 외부 SCM 실제 리소스 프로비저닝은 범위 밖.
