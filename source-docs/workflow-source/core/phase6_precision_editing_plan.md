# Phase 6: 로컬 LLM 편집 정밀화 및 지능형 읽기 도구 최적화 계획

- 문서 목적: 로컬 LLM 환경에서 `edit` 도구의 낮은 신뢰도를 해결하기 위한 robust-patcher 스킬과 semantic-reader MCP 도구의 설계 및 구현 계획을 정리한다.
- 범위: Search-Replace 블록 파서, 퍼지 매칭 엔진, AST 기반 함수/클래스 추출기
- 대상 독자: 워크플로우 설계자, 에이전트 개발자
- 상태: planned
- 최종 수정일: 2026-04-27
- 관련 문서: `core/workflow_kit_roadmap.md`, `core/strategic_threads.md`

## 1. 배경 및 문제 의식
로컬 LLM(7B~32B)은 줄 번호 계산이나 정확한 공백 유지가 필요한 표준 `edit` 도구 사용 시 실패율이 높음. 이는 전체 세션의 품질 저하와 반복적인 재시도로 인한 토큰 낭비를 초래함.

## 2. 해결 전략

### A. [Skill] robust-patcher (Aider-style Search-Replace)
- **방식**: `<<<<<<< SEARCH`, `=======`, `>>>>>>> REPLACE` 태그를 사용하여 수정 범위를 명시.
- **지능화**: 텍스트가 100% 일치하지 않더라도 퍼지 매칭을 통해 가장 유사한 블록을 탐색.
- **안전성**: 적용 전후 문법 검증(Syntax Check) 및 자동 롤백.

### B. [MCP] smart-context-reader (Semantic Reader)
- **방식**: 파일 전체를 읽는 대신 특정 심볼(함수, 클래스)만 추출하여 읽기.
- **이점**: LLM의 컨텍스트 비대화를 막고 핵심 로직에 집중하게 하여 환각 방지.

## 3. 마일스톤 및 태스크
- **TASK-033**: Phase 6 전략 수립 및 `robust-patcher` 상세 설계 (In-Progress)
- **TASK-034**: `robust-patcher` 핵심 엔진 및 스킬 구현
- **TASK-035**: `smart-context-reader` MCP 도구 구현
- **TASK-036**: 하네스(opencode, pi-dev) 지침 업데이트 및 실전 검증

## 4. 성공 기준
- 기존 `edit` 도구가 실패하던 "공백/들여쓰기 오차" 상황에서도 `robust-patcher`가 성공적으로 수정을 수행함.
- `smart-context-reader`를 통해 대규모 파일 작업 시 컨텍스트 점유율을 50% 이상 절감.
