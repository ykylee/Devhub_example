# Alpha v0.2.0

- 날짜: 2026-04-24
- 버전: `v0.2.0-alpha`
- 상태: **Alpha**

## 1. 개요
- 빠른 세션 복원을 위한 상태 캐시(`state.json`) 시스템을 도입하고, Orchestrator-Worker 분화 운영 원칙을 정립했습니다.

## 2. 주요 변경 사항 (Key Changes)

### 🚀 기능 추가 (Features)
- **State Cache 시스템**: `state.json`을 통해 프로젝트 프로파일과 세션 상태를 빠르게 로드하는 캐시 메커니즘 추가.
- **Orchestrator 운영 최적화**: 에이전트가 직접 도구 호출보다 작업 분배와 결과 통합에 집중하도록 가이드라인 업데이트.

### 🛠 버그 수정 및 최적화 (Fixes & Refactoring)
- **Bootstrap 개선**: 하네스 가이드 및 예시 문서들이 `state.json`을 먼저 읽도록 흐름 개선.

## 3. 포함된 배포 패키지 (Assets)
- `standard-ai-workflow-codex-v0.2.0-alpha.zip`
- `standard-ai-workflow-opencode-v0.2.0-alpha.zip`

## 4. 설치 및 업그레이드 가이드
1. `scripts/generate_workflow_state.py`를 사용하여 기존 문서에서 캐시를 생성할 수 있습니다.
