# Strategic Threads (Standard AI Workflow)

- 문서 목적: 여러 세션에 걸쳐 진행되는 상위 전략적 목표와 설계 방향을 관리한다.
- 범위: 전략 스레드 활성화 상태, 설계 가이드라인, 태스크 연결 관계
- 대상 독자: 프로젝트 리드, AI 에이전트, 저장소 관리자
- 상태: in-progress
- 최종 수정일: 2026-04-27
- 관련 문서: `core/workflow_kit_roadmap.md`, `ai-workflow/memory/state.json`
- 운영 원칙:
    1. 모든 계획(Strategy) 수립 시 이 문서의 관련 스레드를 먼저 확인한다.
    2. 새로운 전략적 방향이 생기면 스레드를 추가한다.
    3. 스레드가 완료되거나 폐기되면 사유와 함께 상태를 업데이트한다.
    4. TASK는 반드시 특정 스레드에 귀속되어야 파편화를 방지할 수 있다.

## 1. 활성화된 전략 스레드

### [THREAD-001] SSOT 기반 문서/상태 동기화 자동화
- **목적**: `maturity_matrix.json`을 중심으로 모든 계획 문서와 코드 상태를 일치시킴.
- **상태**: In-Progress
- **설계 방향**:
    - `workflow-linter`를 통한 사후 검사.
    - JSON 데이터를 MD 표로 변환하는 유틸리티 제공.
- **연결된 TASK**: TASK-016 (매트릭스 검증), TASK-017 (로드맵 동기화)

### [THREAD-002] 실운영 파일럿 및 온보딩 마찰 제로화
- **목적**: 타 저장소 도입 시의 허들을 낮추고 실전 데이터를 확보함.
- **상태**: In-Progress
- **설계 방향**:
    - `bootstrap` 결과물 피드백 루프 구축.
    - 첫 세션 자동 브리핑(Starter kit) 강화.
- **연결된 TASK**: TASK-020 (실전 시뮬레이션 검증 완료)

### [THREAD-004] 로컬 LLM 친화적 편집 및 읽기 도구 최적화
- **목적**: 로컬 LLM의 낮은 정밀도를 보완하는 견고한(Robust) 파일 조작 체계 구축.
- **상태**: In-Progress
- **설계 방향**:
    - Aider 스타일의 Search-Replace 블록 및 퍼지 매칭 도입.
    - AST 기반의 시맨틱 심볼(함수/클래스) 추출 읽기 도구 제공.
- **연결된 TASK**: TASK-033, 034, 035, 036

## 2. 완료/보관된 스레드
- [THREAD-003] 운영 지능화 및 가버넌스 자동화 (Phase 5 완결)
    - **성과**: `git_history_summarizer`, `workflow_log_rotator`, `assess_milestone_progress` 도구 구현 및 `automated-repro-scaffold` 스킬 확보. 문서 비대화 방지와 마일스톤 진척도 관리 자동화 달성.
- [THREAD-002] 실운영 파일럿 및 온보딩 마찰 제로화 (일부 TASK 완료, THREAD 단위 종료 대기)

