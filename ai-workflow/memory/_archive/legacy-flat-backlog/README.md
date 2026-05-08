# Legacy flat backlog archive

- 문서 목적: PR #12 머지 전 정리 과정에서 손실 위험이 있는 flat 위치(`ai-workflow/memory/backlog/`) 의 일자별 backlog 원본을 보존한다.
- 범위: brand 별 디렉터리(`gemini/phase6/`, `codex/service-action-command/`, `antigravity/test/backend-integration/`) 로 메모리를 이전하면서 일부 트랙의 상세 본문이 새 구조에 옮겨지지 않은 항목만 archive.
- 대상 독자: 후속 회고/디버깅을 하는 개발자, AI 에이전트
- 상태: stable (read-only)
- 최종 수정일: 2026-05-08
- 관련 문서: `ai-workflow/memory/PR-12-review-actions.md` HYG-5 행, `ai-workflow/MEMORY_GOVERNANCE.md`

## 원본 위치

- `ai-workflow/memory/backlog/2026-05-04.md` (88 줄)
- `ai-workflow/memory/backlog/2026-05-06.md` (32 줄)

두 파일은 `e77a8e9 feat(frontend): implement auth guard, account UI, and websocket client for Phase 5.2 & Phase 3` 커밋에서 삭제되었다.

## 손실 추정 항목

- `2026-05-04.md` 의 **TASK-019 (Codex)** — command/audit 최소 schema 와 risk mitigation command API 1차 구현의 plan/act/validation 본문. 새 구조에서는 `gemini/phase6/state.json`, `codex/service-action-command/state.json` 의 상태 필드와 `codex/service-action-command/backlog/2026-05-07.md` 의 짧은 implementation log 로만 남았다.
- `2026-05-06.md` 의 **TASK-FRONTEND-PHASE5-ORG (Antigravity)** — 조직 관리 UI 와 멤버 할당 모달 고도화의 plan/act 본문. 새 구조의 `antigravity/test/backend-integration/backlog/2026-05-07.md` 에 일부 결과가 요약되어 있으나 plan/act 분리는 보존되지 않았다.

## 정책

- 본 archive 는 **읽기 전용 이력**이다. 새 작업 backlog 는 brand/branch 별 디렉터리에 작성한다.
- archive 항목과 새 backlog 가 충돌하면 새 backlog 를 source of truth 로 한다.
- archive 자체를 다시 정리(삭제)할 때는 별도 PR 에서 사유를 명시한다.
