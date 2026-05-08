# Release Notes - v0.5.0-beta (2026-05-04)

## Multi-Agent Orchestration & Workflow Intelligence

이번 릴리즈는 AI 워크플로우의 '지능형 위임'과 '자동화된 정합성 관리'에 초점을 맞추었습니다. 단순한 툴 실행을 넘어, 에이전트 군단이 서로 협력하고 프로젝트의 상태를 스스로 교정할 수 있는 인프라를 구축했습니다.

### 🚀 주요 변경 사항

#### 1. 다중 에이전트 오케스트레이션 (Multi-Agent Orchestration)
- **전문 워커 군단 도입**: `doc-worker`, `code-worker`, `validation-worker` 전용 페르소나와 지침 완성.
- **Pydantic 데이터 계약**: 에이전트 간 위임 업무(`WorkerTask`)와 결과 보고(`WorkerResponse`)를 정형화하여 통신 안정성 확보.
- **피드백 루프 시뮬레이션**: 워커의 리스크 보고에 따라 오케스트레이터가 동적으로 후속 작업을 위임하는 협업 패턴 검증.

#### 2. 워크플로우 인텔리전스 (Workflow Intelligence)
- **Workflow Linter**: `state.json`, `handoff`, `backlog` 간의 불일치를 자동 탐지하고 복구(`--apply`)하는 Skill 구현.
- **Contextual Git Resolver**: 세션 메타데이터를 분석하여 Git 충돌을 지능적으로 해결하는 전략 제안 엔진 도입.
- **Backlog Steward**: 태스크 문서의 메타데이터와 링크 구조를 자동으로 최신 표준으로 정리.

#### 3. 아키텍처 및 위생 (Architecture & Hygiene)
- **브랜치 기반 상태 격리**: `ai-workflow/memory/<branch>/` 구조를 정식 채택하여 브랜치 간 워크플로우 상태 혼선 방지.
- **내부 정합성 100% 달성**: 저장소 자체의 모든 문서 링크와 태스크 상태를 Linter를 통해 검증 완료.
- **v0.5.0-beta 범프**: 전체 코드베이스와 `workflow_kit` 패키지 버전을 동기화.

### 🛠 설치 및 적용 방법

기존 프로젝트에서 업데이트하려면 `bootstrap_workflow_kit.py`를 다시 실행하거나, `ai-workflow/` 레이어를 최신 소스로 갱신하십시오.

```bash
python3 workflow-source/scripts/bootstrap_workflow_kit.py \
  --target-root . \
  --project-slug [YOUR_PROJECT] \
  --force
```

### ⏭ 다음 단계
- **Phase 12**: 공통 유틸리티의 정식 패키지 배포 및 하네스별 최적화된 바이너리 추출.
- **실전 튜닝**: 파일럿 프로젝트 적용 결과를 바탕으로 워커 프롬프트의 정밀도 향상.

---
**Standard AI Workflow Team**
