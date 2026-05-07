# Alpha v0.3.1

- 날짜: 2026-04-25
- 버전: `v0.3.1-alpha`
- 상태: **Alpha**

## 1. 개요
- 에이전트 컨텍스트 비용 최적화를 위해 템플릿을 정제하고, 하네스 지원 범위를 확장했습니다.

## 2. 주요 변경 사항 (Key Changes)

### 🛠 버그 수정 및 최적화 (Fixes & Refactoring)
- **템플릿 컨텍스트 최적화**: `templates/` 내 불필요한 메타데이터 제거로 에이전트 읽기 비용 40-50% 절감.
- **Bootstrap 로직 개선**: 하드코딩된 문자열을 제거하고 템플릿 파일 직접 참조 방식으로 전환.

### 📄 문서 및 가이드 (Docs)
- **품질 기준 수립**: `core/skill_beta_criteria.md` 추가하여 스킬 승급 기준 정의.
- **Gemini CLI 공식화**: `GEMINI.md` 지침 및 하네스 안정화.

## 3. 포함된 배포 패키지 (Assets)
- `standard-ai-workflow-codex-v0.3.1-alpha.zip`
- `standard-ai-workflow-gemini-cli-v0.3.1-alpha.zip`
- `standard-ai-workflow-opencode-v0.3.1-alpha.zip`
