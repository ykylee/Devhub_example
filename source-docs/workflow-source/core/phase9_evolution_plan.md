# Phase 9: System Maturity & Multi-Agent Evolution Plan

## 1. 개요
Phase 9은 워크플로우 키트의 핵심 엔진을 산업용 수준으로 안정화하고, 단일 에이전트 중심의 실행 구조를 다중 에이전트(Multi-Agent) 기반의 역할 분화 구조로 진화시키는 단계입니다.

## 2. 핵심 전략 (Core Strategies)

### 2.1 엄격한 데이터 계약 (Strict Data Contracts)
- **현황**: JSON 딕셔너리 기반의 느슨한 데이터 전달로 인해 에이전트가 필드를 오해하거나 누락하는 사례 발생.
- **개선**: 모든 스킬과 MCP 도구의 입출력을 Pydantic 모델로 정의.
- **효과**: 런타임 유효성 검사 강화 및 에이전트의 도구 사용 정확도 향상.

### 2.2 정식 MCP v1.0 SDK 전환
- **현황**: 자체 구현한 `read_only_mcp_sdk`를 사용하여 기능 확장에 제한이 있음.
- **개선**: 공식 MCP Python SDK(Bidirectional)로 전면 교체.
- **효과**: 표준 프로토콜 준수, 서버-클라이언트 간 양방향 통신(알림, 리소스) 활용 가능.

### 2.3 역할 기반 에이전트 토폴로지 실현
- **현황**: 오케스트레이터가 모든 도구 호출과 파일 읽기를 직접 수행하여 컨텍스트 비대화 발생.
- **개선**: `workflow_agent_topology.md`에 정의된 워커(Worker) 분화 로직 구현.
    - `doc-worker`: 대량 문서 분석 및 요약.
    - `code-worker`: 정밀 코드 수정 및 구현.
    - `validation-worker`: 테스트 실행 및 로그 분석.
- **연동**: Antigravity의 `sub-agent` 기능을 활용하여 워커 간 협업 자동화.

## 3. 주요 개발 로드맵 (Key Milestones)

### [TASK-050] Pydantic 기반 스키마 프레임워크 구축
- `workflow_kit.common.schemas` 모듈 신설.
- 핵심 엔티티(Backlog, Handoff, Assessment)의 데이터 모델 정의.

### [TASK-051] MCP v1.0 마이그레이션
- `mcp` 패키지 의존성 추가.
- 기존 MCP 서버들을 공식 SDK 기반으로 리팩토링.
- 양방향 통신 스모크 테스트 수행.

### [TASK-052] 워커 전용 프롬프트 번들링
- 각 워커 역할에 최적화된 시스템 프롬프트 및 도구 제한 설정.
- 워커 호출 시 필요한 최소 컨텍스트 추출 로직 구현.

### [TASK-053] git-conflict-resolver 스킬 구현
- 3-way merge 분석 및 프로젝트 특화 해결 전략 엔진 구축.
- 충돌 해결 후 자동 빌드/테스트 검증 루틴 통합.

## 4. 완료 기준 (Definition of Done)
- 모든 스킬이 Pydantic 모델을 통한 입출력 검증을 통과함.
- 공식 MCP SDK 기반으로 서버 3종 이상이 정상 동작함.
- 오케스트레이터가 sub-agent를 호출하여 복합 태스크를 완료하는 시나리오 실증.
- `git-conflict-resolver`가 실제 충돌 상황을 80% 이상의 성공률로 해결함.
