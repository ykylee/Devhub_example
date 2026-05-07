# Beta v0.3.2

- 날짜: 2026-04-28
- 버전: `v0.3.2-beta`
- 상태: **Beta**

## 1. 개요
- 빈 저장소 도입 시의 분석 정확도를 개선하고, 팀 내 워크플로우 전파를 위한 공유용 소개자료를 포함한 릴리스입니다.

## 2. 주요 변경 사항 (Key Changes)

### 🚀 기능 추가 (Features)
- **Phase 6 진입**: 편집 정밀화(robust-patcher) 및 지능형 읽기(smart-context-reader) 엔진이 로드맵에 공식 반영되었습니다.
- **팀 공유 자료**: `ai-workflow/docs/INTRODUCTION_DECK.md` 가 추가되어 팀 내 도입 교육용으로 활용 가능합니다.

### 🛠 버그 수정 및 최적화 (Fixes & Refactoring)
- **실행 런타임 스크립트 누락 버그 수정**: 하네스 배포 패키지 빌드 시 설정 누락으로 인해 스킬(Skill)과 MCP 서버의 파이썬 실행 스크립트(`run_*.py`)가 포함되지 않던 치명적 문제를 수정했습니다.
- **빈 프로젝트 오판 방지**: 기존 저장소 모드 도입 시 키트 내부 파일(`ai-workflow/`)을 대상 코드로 오인하던 스캔 로직을 수정했습니다 (`ignore_dirs` 적용).

### 📄 문서 및 가이드 (Docs)
- **도입 분기 가이드 업데이트**: `core/workflow_adoption_entrypoints.md`에 안정성 개선 사례 반영.
- **성숙도 매트릭스 최신화**: `core/maturity_matrix.json`을 2026-04-28 기준으로 동기화.

## 3. 포함된 배포 패키지 (Assets)
- `standard-ai-workflow-codex-v0.3.2-beta.zip`
- `standard-ai-workflow-gemini-cli-v0.3.2-beta.zip`
- `standard-ai-workflow-opencode-v0.3.2-beta.zip`
- `standard-ai-workflow-pi-dev-v0.3.2-beta.zip`

## 4. 설치 및 업그레이드 가이드
1. `scripts/bootstrap_workflow_kit.py`를 실행하여 최신 키트를 배포합니다.
2. 기존 프로젝트라면 생성된 `repository_assessment.md`를 통해 분석 결과의 정확도를 확인합니다.

---
*본 릴리스는 표준 AI 워크플로우 가버넌스를 준수하여 작성되었습니다.*
