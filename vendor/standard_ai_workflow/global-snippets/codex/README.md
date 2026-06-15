# Codex Global Snippet

- 문서 목적: Codex 전역 설정에 넣어도 비교적 안전한 workflow snippet 예시와 사용 원칙을 정리한다.
- 범위: additive MCP 예시, 주의점, 프로젝트 로컬 레이어와의 관계
- 대상 독자: Codex 사용자, 저장소 관리자, 운영 담당자
- 상태: draft
- 최종 수정일: 2026-04-19
- 관련 문서: `../../core/workflow_global_injection_policy.md`, `../../harnesses/codex/apply_guide.md`, `./config.toml.snippet`

## 권장 원칙

- Codex 전역 설정에는 additive MCP 연결이나 기능 플래그만 넣는다.
- model, provider, reasoning effort 기본값은 넣지 않는다.
- 실제 프로젝트 workflow 진입은 `AGENTS.md` 와 `ai-workflow/memory/` 문서에서 수행한다.

## 포함된 파일

- [config.toml.snippet](./config.toml.snippet)

## 사용 방법

- 현재 `~/.codex/config.toml` 과 병합 전에 diff 로 검토한다.
- 이미 존재하는 MCP 서버 정의와 이름이 충돌하지 않는지 확인한다.
- 프로젝트별 규칙은 절대 전역 snippet 에 추가하지 않는다.

## 다음에 읽을 문서

- Codex 적용 가이드: [../../harnesses/codex/apply_guide.md](../../harnesses/codex/apply_guide.md)
- 비침투적 주입 정책: [../../core/workflow_global_injection_policy.md](../../core/workflow_global_injection_policy.md)
