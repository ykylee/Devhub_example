# Workflow Index

- 문서 목적: 현재 저장소에 배포된 AI workflow 문서의 진입 순서를 제공한다.
- 범위: 세션 시작, 상태 복원, 온보딩, 코어 정책 문서
- 대상 독자: Codex, 저장소 관리자, 신규 온보딩 참여자
- 상태: active
- 최종 수정일: 2026-05-06
- 관련 문서: [README.md](./README.md), [memory/state.json](./memory/state.json)

## 1. 세션 시작 순서

1. [memory/state.json](./memory/state.json)
2. [memory/session_handoff.md](./memory/session_handoff.md)
3. [memory/work_backlog.md](./memory/work_backlog.md)
4. 최신 날짜 backlog: [memory/backlog/2026-05-04.md](./memory/backlog/2026-05-04.md)
5. [PROJECT_PROFILE.md](../docs/PROJECT_PROFILE.md)

## 2. 온보딩 점검 문서

- [README.md](./README.md)
- [memory/project_status_assessment.md](./memory/project_status_assessment.md)
- [core/global_workflow_standard.md](./core/global_workflow_standard.md)
- [core/workflow_adoption_entrypoints.md](./core/workflow_adoption_entrypoints.md)
- [core/workflow_skill_catalog.md](./core/workflow_skill_catalog.md)

## 3. 운영 기준

- `ai-workflow/memory/`는 현재 프로젝트의 workflow 상태 문서다.
- `ai-workflow/core/`는 별도 도구 없이도 수동으로 따를 수 있는 공통 운영 기준만 남긴다.
- 배포용 source bundle, scripts, skills, MCP, tests는 현재 저장소에 포함하지 않는다.

## 4. 다음 보강 후보

- `memory/PROJECT_PROFILE.md`와 루트 `AGENTS.md`의 실행 명령 TODO 정리.
- `memory/project_status_assessment.md`의 프로젝트명, 목적, 검증 명령 최신화.
- 필요 시 별도 workflow kit source bundle에서 자동화 도구를 재설치.
