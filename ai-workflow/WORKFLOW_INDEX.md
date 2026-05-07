# Workflow Index

- 문서 목적: 현재 저장소에 배포된 AI workflow 문서의 진입 순서를 제공한다.
- 범위: 세션 시작, 상태 복원, 온보딩, 코어 정책 문서
- 대상 독자: Codex, 저장소 관리자, 신규 온보딩 참여자
- 상태: draft
- 최종 수정일: 2026-05-07
- 관련 문서: [README](./README.md), [Memory Governance](./MEMORY_GOVERNANCE.md)

## 1. 세션 시작 순서

1. 현재 git 브랜치를 확인한다: `git branch --show-current`
2. 브랜치별 memory 디렉터리를 연다.
   - `codex/service-action-command` → [memory/codex/service-action-command/](./memory/codex/service-action-command/)
   - `claude/init` → [memory/claude/init/](./memory/claude/init/)
   - `claude/phase13` → [memory/claude/phase13/](./memory/claude/phase13/)
3. 브랜치별 `state.json`, `session_handoff.md`, `work_backlog.md`, 최신 `backlog/YYYY-MM-DD.md`를 읽는다.
4. 공용 기준 문서를 읽는다.
   - [memory/PROJECT_PROFILE.md](./memory/PROJECT_PROFILE.md)
   - [memory/repository_assessment.md](./memory/repository_assessment.md)

## 2. 온보딩 점검 문서

- [README.md](./README.md)
- [memory/project_status_assessment.md](./memory/project_status_assessment.md)
- [core/global_workflow_standard.md](./core/global_workflow_standard.md)
- [core/workflow_adoption_entrypoints.md](./core/workflow_adoption_entrypoints.md)
- [core/workflow_skill_catalog.md](./core/workflow_skill_catalog.md)

## 3. 운영 기준

- `ai-workflow/memory/<agent>/<branch>/`는 브랜치별 workflow 상태 문서다.
- `ai-workflow/memory/PROJECT_PROFILE.md`, `repository_assessment.md`, `environments/`는 공용 기준 문서다.
- flat `ai-workflow/memory/state.json`, `session_handoff.md`, `work_backlog.md`, `backlog/`는 legacy fallback 및 공용 색인 전용이다.
- `ai-workflow/core/`는 별도 도구 없이도 수동으로 따를 수 있는 공통 운영 기준만 남긴다.
- 배포용 source bundle, scripts, skills, MCP, tests는 현재 저장소에 포함하지 않는다.

## 4. 다음 보강 후보

- `memory/PROJECT_PROFILE.md`와 루트 `AGENTS.md`의 실행 명령 TODO 정리.
- `memory/project_status_assessment.md`의 프로젝트명, 목적, 검증 명령 최신화.
- 필요 시 별도 workflow kit source bundle에서 자동화 도구를 재설치.
