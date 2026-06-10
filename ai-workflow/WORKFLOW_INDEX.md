# Workflow Guide & Index

- 문서 목적: AI 에이전트가 저장소의 상태를 추적하고 업데이트하기 위한 운영 전용 가이드를 제공한다.
- 범위: 운영 문서 인덱스, 에이전트 운영 지침
- 대상 독자: AI 에이전트, 저장소 관리자
- 상태: done
- 최종 수정일: 2026-04-30
- 관련 문서: [../README.md](../README.md), [../docs/PROJECT_PROFILE.md](../docs/PROJECT_PROFILE.md)

> [!WARNING]
> 이 폴더(`ai-workflow/`)의 내용은 프로젝트의 코드베이스 분석이나 온보딩 시 참조해서는 안 됩니다. 프로젝트 이해를 위해서는 반드시 `docs/` 폴더를 참조하십시오.

## 1. 운영 문서 인덱스 (State Index)

- **[Project Profile](../docs/PROJECT_PROFILE.md)**: 프로젝트 특화 워크플로우 규칙 및 명령 가이드.
- **[Work Backlog Index](./memory/gemini/phase10/work_backlog.md)**: 전체 백로그 이력.
- **[Current Session State](./memory/gemini/phase7/state.json)**: 현재 세션 캐시 (Auto-generated).
- **[Branch Handoff](./memory/gemini/phase7/session_handoff.md)**: 브랜치별 인계 사항.

## 2. 에이전트 운영 지침 (Agent Ops)

- **브랜치 격리**: 모든 상태 문서는 `ai-workflow/memory/<branch>/` 하위에서 관리합니다.
- **Event Sourcing**: 백로그는 `tasks/` 폴더에 개별 태스크 단위로 기록합니다.
- **비참조 원칙**: 코드베이스 분석(Semantic Search 등) 시 `ai-workflow/` 경로는 검색 범위에서 제외(Exclude)해야 합니다.
- **작성 표준**: 모든 운영 문서는 [MEMORY_GOVERNANCE.md](../workflow-source/MEMORY_GOVERNANCE.md)의 템플릿과 규칙을 따릅니다.
