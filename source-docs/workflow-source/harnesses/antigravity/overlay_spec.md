# Antigravity Overlay Spec

- 문서 목적: `Antigravity` 하네스 오버레이를 구현하기 전에 필요한 파일 목록과 연결 전략을 정리한다.
- 범위: 진입 파일, 설정 파일, 공통 workflow 참조 경로, 권한 정책 초안
- 대상 독자: 저장소 관리자, AI workflow 설계자, 하네스 통합 담당자
- 상태: draft
- 최종 수정일: 2026-04-19
- 관련 문서: `./README.md`, `../../core/workflow_harness_distribution.md`, `../../scripts/bootstrap_workflow_kit.py`

## 1. 하네스 식별자

- slug:
- `antigravity`
- display name:
- `Antigravity`

## 2. 예상 진입 파일

- 루트 진입 파일:
- `ANTIGRAVITY.md`
- 설정 파일:
- `TODO: 하네스 전용 설정 파일 경로`
- 추가 overlay 파일:
- TODO

## 3. 공통 workflow 연결 규칙

- 항상 `ai-workflow/memory/session_handoff.md` 를 우선 읽게 연결한다.
- 항상 `ai-workflow/memory/work_backlog.md` 를 참조하게 연결한다.
- 항상 `ai-workflow/memory/PROJECT_PROFILE.md` 를 참조하게 연결한다.
- 필요하면 `ai-workflow/memory/repository_assessment.md` 를 adoption 단계 보조 문서로 사용한다.
- read-only MCP draft descriptor 는 `bundle/source-docs/schemas/read_only_transport_descriptors.json` 위치를 기준으로 검토한다.

## 4. 구현 체크리스트

- bootstrap 생성 함수 추가
- 레지스트리 등록
- smoke test 추가
- README 문서 갱신
- descriptor export 산출물 포함 여부 확인

## 다음에 읽을 문서

- 하네스 패키지 안내: [./README.md](./README.md)
- 하네스 템플릿: [../_template/README.md](../_template/README.md)
