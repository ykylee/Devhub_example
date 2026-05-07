# Package Contents

- 문서 목적: `standard-ai-workflow-antigravity` 배포 패키지에 무엇이 들어 있는지와 어떤 파일부터 읽어야 하는지 빠르게 안내한다.
- 범위: 패키지 레이어, 포함 파일, 기본 제외 항목, 권장 진입점
- 대상 독자: 배포 패키지를 받는 AI agent 운영자, 저장소 관리자, 하네스 통합 담당자
- 상태: draft
- 최종 수정일: 2026-05-01
- 관련 문서: `./manifest.json`, `./APPLY_GUIDE.md`

## 1. 패키지 식별자

- 패키지명: `standard-ai-workflow-antigravity`
- 하네스: `antigravity`
- 버전: `v0.4.1-beta`
- 배포 프로필: `agent_runtime_minimal`

## 2. 포함 레이어

- 공통 runtime workflow 레이어
- 하네스 runtime overlay 레이어
- 패키지 메타데이터 레이어

## 3. 공통 runtime workflow 파일

- `bundle/ai-workflow/README.md`
- `bundle/ai-workflow/core/global_workflow_standard.md`
- `bundle/ai-workflow/core/workflow_adoption_entrypoints.md`
- `bundle/ai-workflow/core/workflow_skill_catalog.md`
- `bundle/ai-workflow/memory/PROJECT_PROFILE.md`
- `bundle/ai-workflow/memory/state.json`
- `bundle/ai-workflow/memory/session_handoff.md`
- `bundle/ai-workflow/memory/work_backlog.md`
- `bundle/ai-workflow/memory/backlog/2026-04-23.md`

## 4. 하네스 runtime overlay 파일

- `bundle/ANTIGRAVITY.md`

## 5. 패키지 메타데이터

- `manifest.json`
- `PACKAGE_CONTENTS.md`
- `APPLY_GUIDE.md`

## 6. 기본 제외 항목

- 개발 참고용 source docs: 기본 제외
- 전역 설정 snippet 예시: 기본 제외
- draft MCP descriptor/fixture/reference docs: 기본 제외

기본 프로필은 실제 AI agent 가 읽는 런타임 파일만 남겨 컨텍스트 낭비를 줄이는 것을 목표로 한다.

## 7. 권장 진입점

- `bundle/ANTIGRAVITY.md`
- `bundle/ai-workflow/README.md`
- `bundle/ai-workflow/memory/state.json`
- `bundle/ai-workflow/memory/session_handoff.md`
- `bundle/ai-workflow/memory/work_backlog.md`
- `bundle/ai-workflow/memory/PROJECT_PROFILE.md`
- `bundle/ai-workflow/core/workflow_adoption_entrypoints.md`
- `bundle/ai-workflow/core/workflow_skill_catalog.md`

## 8. 다음 단계

- 먼저 [APPLY_GUIDE.md](./APPLY_GUIDE.md) 를 읽고 대상 저장소에 복사할 파일 범위를 정한다.
- 이후 `manifest.json` 으로 실제 포함 파일과 제외 정책을 다시 확인한다.
