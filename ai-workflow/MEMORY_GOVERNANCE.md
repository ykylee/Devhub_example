# AI-First Memory Governance

- 문서 목적: AI 에이전트가 관리하는 운영 문서(Workflow State)의 관리 규칙과 템플릿을 정의한다.
- 범위: 상태 문서 분류, 작성 표준, 메타데이터 요구사항
- 대상 독자: AI 에이전트, 저장소 관리자
- 상태: stable
- 최종 수정일: 2026-04-30
- 관련 문서: [../ai-workflow/WORKFLOW_INDEX.md](../ai-workflow/WORKFLOW_INDEX.md), [../README.md](../README.md)

이 문서는 `ai-workflow/memory/` 하위 문서를 작성할 때 AI 에이전트가 준수해야 할 규칙과 템플릿을 정의합니다.

## 1. 작성 규칙 (Writing Rules)

- **언어**: 사용자 보고용 요약은 한국어를 사용하되, 상태 값이나 기술적 명칭은 영문 표준을 권장합니다.
- **간결성**: 중복된 설명을 피하고, 변경 사항(Diff)과 다음 행동(Next Action)에 집중합니다.
- **구조화**: Key-Value 쌍(예: `Status: in_progress`) 또는 Markdown Table을 적극 활용합니다.
- **격리**: 문서 간의 의존성을 최소화하고, 각 파일이 독립적인 컨텍스트를 완결성 있게 담도록 합니다.

## 2. 표준 템플릿 (Standard Templates)

### 📂 Session Handoff (`session_handoff.md`)
```markdown
# Session Handoff
- Branch: [branch_name]
- Updated: [YYYY-MM-DD HH:mm]

## 🎯 Current Focus
[현재 작업의 핵심 목표 1줄]

## 📊 Work Status
- [TASK-ID] [Title]: [Status] ([Progress %])
- [최근 수행한 핵심 변경 사항 및 결과]

## ⏭️ Next Actions
- [ ] [다음에 즉시 수행할 작업]

## ⚠️ Risks & Blockers
- [차단 요소 또는 주의가 필요한 아키텍처적 결정 사항]
```

### 📂 Task Detail (`backlog/tasks/TASK-XXX.md`)
```markdown
---
id: TASK-XXX
status: [planned|in_progress|done|blocked]
created_at: YYYY-MM-DD
---
# [Task Title]

## 📝 Description
[작업의 정의 및 범위]

## 🛠️ Implementation Log
- [YYYY-MM-DD]: [수행 내용 요약]

## ✅ Outcome
[완료 시 결과물 또는 검증 결과]
```

### 📂 Daily Backlog Index (`backlog/YYYY-MM-DD.md`)
```markdown
# YYYY-MM-DD Branch Work Backlog

- Purpose: Link task detail files for one working day.
- Status: in_progress
- Updated: YYYY-MM-DD

## Tasks

- TASK-XXX Task title: `./tasks/YYYY-MM-DD_TASK-XXX.md`
```

- `backlog/tasks/*.md` is the source of truth for detailed task state.
- `backlog/YYYY-MM-DD.md` is a tracked lightweight index. Keep it small and link-oriented.
- On merge conflicts, rebuild the daily index from task links and resolve detailed state in each task file.

## 3. 에이전트 행동 지침

- 세션 종료 시 `session_handoff.md`를 위 템플릿에 맞춰 갱신하십시오.
- 새로운 작업 시작 시 `tasks/` 폴더에 템플릿 기반의 신규 파일을 생성하십시오.
- 날짜별 백로그에는 신규 task 파일 링크만 추가하고, 긴 상세 기록은 task 파일에 남기십시오.
- 상태 업데이트 시 자연어 서술보다는 불렛 포인트와 상태 키워드를 우선하십시오.
