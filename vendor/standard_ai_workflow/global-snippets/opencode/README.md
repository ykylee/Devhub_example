# OpenCode Global Snippet

- 문서 목적: OpenCode 전역 설정에 넣어도 비교적 안전한 workflow snippet 예시와 사용 원칙을 정리한다.
- 범위: additive MCP 예시, instructions 제외 원칙, 사용자 기본값 보호 원칙
- 대상 독자: OpenCode 사용자, 저장소 관리자, 운영 담당자
- 상태: draft
- 최종 수정일: 2026-04-19
- 관련 문서: `../../core/workflow_global_injection_policy.md`, `../../harnesses/opencode/apply_guide.md`, `./opencode.global.jsonc`

## 권장 원칙

- OpenCode 전역 설정에는 additive MCP 연결이나 공유 가능한 안전 기본값만 둔다.
- project-local `instructions` 는 전역 설정이 아니라 프로젝트 `opencode.json` 에 둔다.
- model, provider, top-level permission 기본값은 전역 설정이 이미 있다면 그대로 유지한다.

## 포함된 파일

- [opencode.global.jsonc](./opencode.global.jsonc)

## 사용 방법

- `~/.config/opencode/opencode.json` 과 병합 전에 diff 로 검토한다.
- 이미 쓰고 있는 provider/model 설정이 있는 경우 snippet 은 additive 항목만 추가한다.
- 프로젝트별 workflow 문서 경로는 전역 설정이 아니라 프로젝트 루트 `opencode.json` 에서 관리한다.

## 다음에 읽을 문서

- OpenCode 적용 가이드: [../../harnesses/opencode/apply_guide.md](../../harnesses/opencode/apply_guide.md)
- 비침투적 주입 정책: [../../core/workflow_global_injection_policy.md](../../core/workflow_global_injection_policy.md)
