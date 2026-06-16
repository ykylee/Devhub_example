# Session Handoff

- 브랜치: `codex/project_concept_design`
- 날짜: 2026-05-14
- 상태: in_progress

## 이번 세션 완료 사항
- `Application > Repository > Project` 기준 문서 정합화 완료.
- AI 가드너 기능을 v2 범위로 이관 (요구사항/로드맵/설계/추적성 반영).
- 역할별 메뉴/화면/API 매트릭스 문서 추가.
- Application 하위 운영 데이터(Repository 작업현황, PR/PR activity, 빌드, 품질지표) 설계 반영.
- 권한 정책 반영: `system_admin` 기본, `pmo_manager` 후보(비활성).
- `Application.key` 정책 구체화 반영:
  - 입력 정책 `^[A-Za-z0-9]{10}$`
  - DB 컬럼은 여유 길이, 앱 레벨 검증 강제
- Usecase/ERD/추적성 정합화:
  - UC 범위 `UC-APP-01..07` 반영
  - ERD 필드명 `code -> key` 통일 (Application/Project)
- API 계약 고도화:
  - `docs/backend_api_contract.md` §13 신설
  - Application/Repository/Project planned API + 공통 에러 코드 초안 추가
- Application 상세 설계 고도화:
  - 상태 전이별 권한/검증 가드 확정 (`hold_reason`, `resume_reason`, `archived_reason`, close/activate precondition)
  - 롤업 가중치 정책 확정 (`equal`, `repo_role`, `custom`) + 메타(`applied_weights`, `fallbacks`) 반영
  - `sync_error_code` 표준 사전 + retryable 정책 반영
  - API 요청/응답 JSON 예시 스키마 본문 추가
- 원격 push 완료:
  - `849dfcb` docs: defer AI gardener to v2 and add view/menu/api matrix
  - `2b35d13` docs: design application-repository-project data and repo ops metrics

## 다음 세션 우선 작업
1. planned API를 API ID 체계로 편입 (`REQ -> UC -> API-ID -> IMPL` 추적 라인 확정).
2. ERD 기반 마이그레이션/스토어 인터페이스 설계 초안 작성.
3. Application/Repository API validation/error matrix 상세화.
