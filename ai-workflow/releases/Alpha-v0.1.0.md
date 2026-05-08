# Alpha v0.1.0

- 날짜: 2026-04-23
- 버전: `v0.1.0-alpha`
- 상태: **Alpha**

## 1. 개요
- 표준 AI 워크플로우의 첫 번째 프로토타입으로, 핵심 스킬 묶음과 하네스별 최소 실행 환경(Minimal Runtime) 배포를 목표로 합니다.

## 2. 주요 변경 사항 (Key Changes)

### 🚀 기능 추가 (Features)
- **Workflow/Skill 초기 도입**: `bootstrap`, `onboarding runner`, `skill catalog` 등 핵심 진입점 구축.
- **Harness Minimal Runtime**: AI 에이전트 컨텍스트 절약을 위해 필요한 파일만 추출하여 배포하는 `export` 로직 구현.
- **MCP 준비**: read-only JSON-RPC bridge 및 공식 MCP SDK candidate 검증 기반 마련.

### 📄 문서 및 가이드 (Docs)
- **도입 가이드**: `core/workflow_adoption_entrypoints.md` 및 파일럿 적용 템플릿 추가.
- **배포 규격**: `core/workflow_release_spec.md` 정의.

## 3. 포함된 배포 패키지 (Assets)
- `standard-ai-workflow-codex-v0.1.0-alpha.zip`
- `standard-ai-workflow-opencode-v0.1.0-alpha.zip`

## 4. 설치 및 업그레이드 가이드
1. `scripts/bootstrap_workflow_kit.py`를 실행하여 초기 문서를 생성합니다.
2. 상세 가이드는 `ai-workflow/README.md`를 참고하십시오.
