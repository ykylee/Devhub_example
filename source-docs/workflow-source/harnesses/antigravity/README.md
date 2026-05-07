# Antigravity Harness Package

- 문서 목적: 표준 AI 워크플로우를 `Antigravity` 하네스에 맞춰 배포할 때 생성할 파일과 검토 포인트를 정리한다.
- 범위: 루트 진입 파일, 설정 파일, 공통 workflow 문서 연결 방식
- 대상 독자: Antigravity 사용자, 저장소 관리자, AI workflow 설계자
- 상태: draft
- 최종 수정일: 2026-04-19
- 관련 문서: `../../core/workflow_harness_distribution.md`, `../../scripts/bootstrap_workflow_kit.py`, `../_template/README.md`

## 생성 대상

- `ANTIGRAVITY.md`
- `TODO: 하네스 전용 설정 파일 경로`
- TODO: 하네스가 요구하는 추가 overlay 파일

## 구성 원칙

- `ANTIGRAVITY.md` 는 하네스의 실제 진입 파일로 사용한다.
- 상세 정책은 `ai-workflow/memory/` 문서를 먼저 읽도록 연결한다.
- 설정 파일은 가능한 한 최소 예시만 두고, 긴 정책 본문은 공통 문서로 위임한다.
- export bundle 의 `bundle/source-docs/schemas/read_only_transport_descriptors.json` 는 read-only MCP 연결 검토용 draft descriptor 로 취급한다.
- descriptor 의 `transport_ready` 값이 `false` 인 동안에는 실제 MCP 연결보다 참고 산출물로 두는 편이 안전하다.

## bootstrap 확장 TODO

- `scripts/bootstrap_workflow_kit.py` 에 `antigravity` 하네스 생성 함수를 추가한다.
- `SUPPORTED_HARNESSES`, `HARNESS_DEFINITIONS`, `HARNESS_FILE_BUILDERS` 에 `antigravity` 를 등록한다.
- `tests/check_bootstrap.py` 에 `antigravity` 생성 검증을 추가한다.
- descriptor export 위치와 draft 사용 범위를 문서화한다.

## 다음에 읽을 문서

- 하네스 허브: [../README.md](../README.md)
- 하네스 템플릿: [../_template/README.md](../_template/README.md)
- 배포 전략: [../../core/workflow_harness_distribution.md](../../core/workflow_harness_distribution.md)
