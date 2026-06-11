# Skill: project-status-assessment

- 문서 목적: 저장소의 기술 스택, 구조, 테스트, 문서화 수준을 진단하여 워크플로우 도입 성숙도를 평가하는 스킬을 설명한다.
- 범위: 입력 및 출력 계약, 진단 항목, 실행 방법
- 대상 독자: AI 에이전트, 워크플로우 온보딩 담당자
- 상태: beta
- 최종 수정일: 2026-04-26
- 관련 문서: `core/project_status_assessment.md`, `core/workflow_skill_catalog.md`

## 1. 개요

새로운 프로젝트에 워크플로우를 도입하기 전, 해당 프로젝트의 현재 상태를 객관적으로 파악하는 것이 중요하다. 이 스킬은 파일 시스템을 탐색하여 주요 설정 파일, 소스 디렉토리, 테스트 코드의 존재 여부를 확인하고 표준화된 진단 리포트를 제공한다.

## 2. 입력 및 출력

### 입력 (Inputs)
- `project-root`: 진단 대상 프로젝트 루트 경로 (기본값: `.`)
- `output-path`: 리포트 생성 경로 (기본값: `ai-workflow/memory/repository_assessment.md`)
- `apply`: 실제 파일 생성/갱신 여부

### 출력 (Outputs)
- `status`: "ok"
- `assessment`:
    - `primary_stack`: 감지된 주요 기술 스택 (예: Python, Node.js, Go)
    - `structure_score`: 구조 체계화 점수
    - `test_coverage_hint`: 테스트 존재 여부 및 유형
    - `docs_score`: 문서화 수준
- `recommended_actions`: 워크플로우 도입을 위한 우선 보강 사항

## 3. 주요 진단 로직

1. **스택 감지**: `package.json`, `requirements.txt`, `go.mod`, `Cargo.toml` 등을 통해 주력 언어와 프레임워크 파악.
2. **구조 분석**: `src/`, `lib/`, `tests/`, `docs/` 등 표준 디렉토리 구성 확인.
3. **명령어 추정**: `scripts/` 또는 설정 파일을 분석하여 설치(`install`), 실행(`run`), 테스트(`test`) 명령어 후보 추출.
4. **성숙도 평가**: 진단 결과에 따라 1~4단계 성숙도 판정.

## 4. 실행 방법

```bash
python3 skills/project-status-assessment/scripts/run_project_status_assessment.py \
  --project-root /path/to/target/project \
  --apply
```
