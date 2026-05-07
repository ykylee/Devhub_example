# Project Workflow Maturity Assessment

- 문서 목적: 프로젝트의 AI 워크플로우 도입 수준을 자가 진단하고 개선 포인트를 도출한다.
- 범위: 기본 문서화, 도구 활용도, 프로세스 준수도, 지능화 수준
- 대상 독자: 프로젝트 리드, AI 에이전트, 온보딩 담당자
- 상태: draft
- 최종 수정일: 2026-05-06
- 관련 문서: [공통 표준](../core/global_workflow_standard.md), [세션 인계](./session_handoff.md)

## 1. 진단 요약 (Executive Summary)

- **현재 레벨**: Alpha
- **핵심 강점**: 세션 인계, 백로그, 프로젝트 프로파일, 로드맵이 분리되어 있고 주요 작업 기록이 남아 있다.
- **개선 필요 항목**: AGENTS/워크플로우 문서 경로 정합성, 최신 backlog 생성 주기, 자동 검증 명령 정착.
- **차기 목표**: `ai-workflow/memory/` 기준 운영 문서를 단일 source of truth로 고정하고 세션 종료 시 state/handoff/backlog를 일관되게 갱신한다.

---

## 2. 진단 매트릭스 (Assessment Matrix)

| 구분 | 진단 항목 | 현황 (0~3) | 비고 |
| --- | --- | --- | --- |
| **기본 문서** | `PROJECT_PROFILE.md`가 최신 상태인가? | 2 | 실제 명령과 운영 경로가 대체로 반영됨 |
| | `session_handoff.md`가 매 세션 갱신되는가? | 2 | 최근 frontend/backend 작업 인계 존재 |
| | `work_backlog.md`가 실제 작업과 동기화되는가? | 2 | TASK-020까지 반영, 일자별 최신화 필요 |
| **도구 활용** | MCP 도구를 사용하여 문서를 조회/수정하는가? | 1 | 수동 파일 갱신 중심 |
| | `workflow-linter`를 주기적으로 실행하는가? | 1 | 문서 테스트는 존재하나 자동화 미정착 |
| | `session-start` 스킬로 컨텍스트를 복원하는가? | 1 | AGENTS 기반 수동 복원 |
| **프로세스** | 작업 전 브리핑 및 계획 수립을 수행하는가? | 2 | 백로그에 Plan/Act/Validate/Result 구조 존재 |
| | 검증(Validate) 단계를 반드시 거치는가? | 2 | Go 테스트와 migration 검증 일부 기록 |
| | 작업 모드(Task Modes)를 명시하여 최적화하는가? | 1 | 표준 문서는 있으나 태스크별 적용은 제한적 |
| **품질/거버넌스** | `state.json`과 문서가 동기화되는가? | 2 | main 반영 후 project 경로 기준으로 정리 필요 |
| | 릴리즈 노트 형식을 준수하여 배포하는가? | 1 | PR/백로그 기록 중심 |

*점수 가이드: 0(미도입), 1(수동 도입), 2(부분 자동화), 3(완전 정착)*

---

## 3. 레벨별 정의 (Level Definitions)

### [Alpha] 도입 단계

- 기본 운영 문서(`project/`)가 존재함.
- 에이전트가 문서를 수동으로 읽고 갱신함.
- 검증 절차가 정의되어 있으나 누락되는 경우가 있음.

### [Beta] 가속 단계

- MCP 도구 및 표준 스킬을 적극 활용함.
- 세션 간 인계(`handoff`)가 정형화됨.
- 작업 모드(Task Modes)를 인지하여 효율적으로 작업을 분담함.

### [Stable] 최적화 단계

- 워크플로우 린트가 자동화되어 문서 정합성이 상시 보장됨.
- 성숙도 매트릭스에 기반한 지능형 작업 분배가 이루어짐.
- 프로젝트 특화 스킬 및 도구가 커스텀 개발되어 적용됨.

---

## 4. 향후 개선 계획 (Roadmap to Next Level)

- [ ] `AGENTS.md`, `ai-workflow/README.md`, `state.json`의 운영 문서 경로를 `ai-workflow/memory/` 기준으로 일치시킨다.
- [ ] 세션 종료 시 최신 backlog와 handoff를 자동/반자동으로 갱신하는 검증 절차를 정한다.
- [ ] `pytest ai-workflow/tests/check_docs.py`를 문서 변경 기본 검증으로 고정한다.

## 다음에 읽을 문서

- [공통 표준](../core/global_workflow_standard.md)
- [workflow 인덱스](../WORKFLOW_INDEX.md)
- [스킬 카탈로그](../core/workflow_skill_catalog.md)
