# Gemini CLI Workflow Apply Guide

- 문서 목적: 기존 또는 신규 프로젝트에 표준 AI 워크플로우를 Gemini CLI 하네스 기준으로 적용하는 실제 절차를 단계별로 안내한다.
- 범위: bootstrap 실행, 생성 파일 검토, GEMINI.md 연결, 첫 세션 시작 방법
- 대상 독자: Gemini CLI 사용자, 저장소 관리자, AI workflow 설계자
- 상태: draft
- 최종 수정일: 2026-04-25
- 관련 문서: `./README.md`, `../../core/workflow_adoption_entrypoints.md`, `../../scripts/bootstrap_workflow_kit.py`

## 1. 언제 이 가이드를 쓰는가

- 프로젝트에서 Gemini CLI 를 주 하네스로 사용하려고 할 때
- 표준 workflow 문서를 Gemini CLI 의 `GEMINI.md` 진입점과 연결하려고 할 때
- 신규 프로젝트 또는 기존 프로젝트에 Gemini CLI 기준 도입을 시작하려고 할 때

## 2. 적용 전 확인

- Gemini CLI 가 프로젝트 루트의 `GEMINI.md` 를 시스템 지침보다 우선해서 읽는 흐름을 사용한다.
- 프로젝트에서 workflow 문서를 둘 위치를 `ai-workflow/` 로 유지할지 결정한다.
- 기존 프로젝트라면 기본 실행 명령과 테스트 명령을 자동 추정값과 대조할 사람이 필요하다.

## 2.1 권장 설정 계층

- 공유:
- 프로젝트 루트 `GEMINI.md` 와 `ai-workflow/` 패키지를 둔다.
- 로컬:
- `ai-workflow/memory/` 문서에서 실제 명령, 경로, backlog 상태를 관리한다.

프로젝트별 규칙은 항상 local 문서가 우선한다.

## 2.2 언어와 컨텍스트 운영 원칙

- Gemini CLI 에서 사용자에게 직접 보이는 작업 보고, 상태 요약, 문서 초안은 한국어로 작성하도록 `GEMINI.md` 에 명시한다.
- 코드, 명령어, 파일 경로, 설정 key 는 필요한 경우 원문 그대로 유지한다.
- 내부 사고 과정, 임시 분류, 중간 reasoning 은 모델이 효율적으로 처리하게 두고, 사용자에게는 필요한 결론만 짧게 전달하도록 한다.
- 진행 업데이트는 짧고 목적 지향적으로 유지하고, 이미 확인한 사실을 반복해서 길게 설명하지 않는다.
- handoff 와 backlog 에는 다음 세션에 필요한 핵심 사실만 남겨 컨텍스트 누적을 줄인다.
- Gemini CLI 에서도 가능한 경우 메인 에이전트는 조정/통합에 집중하고, bounded scope 읽기/쓰기/검증은 `invoke_agent` 를 통해 서브 에이전트로 분리하는 운영 패턴을 권장한다.
- 서브 에이전트에게는 책임 범위와 종료 조건을 명확히 넘기고, 메인 에이전트에는 핵심 사실과 결과만 다시 모은다.

## 3. 신규 프로젝트 적용 순서

1. 아래 명령으로 기본 문서 세트와 Gemini CLI overlay 를 생성한다.

```bash
python3 scripts/bootstrap_workflow_kit.py \
  --target-root /path/to/project \
  --project-slug my_project \
  --project-name "My Project" \
  --adoption-mode new \
  --harness gemini-cli \
  --copy-core-docs
```

2. 생성된 `ai-workflow/memory/PROJECT_PROFILE.md` 에 실제 명령과 검증 규칙을 채운다.
3. 루트 `GEMINI.md` 가 `ai-workflow/memory/` 문서를 먼저 읽도록 연결됐는지 확인한다.
4. 첫 세션에서 `session_handoff.md` 와 오늘 날짜 backlog 를 채운다.

## 4. 작업 중인 프로젝트 적용 순서

1. 아래 명령으로 기존 저장소 분석과 Gemini CLI overlay 를 함께 생성한다.

```bash
python3 scripts/bootstrap_workflow_kit.py \
  --target-root /path/to/project \
  --project-slug my_project \
  --project-name "My Project" \
  --adoption-mode existing \
  --harness gemini-cli \
  --copy-core-docs
```

2. `ai-workflow/memory/repository_assessment.md` 를 읽고 추정 스택, 명령, 문서 경로가 실제 저장소와 맞는지 검토한다.
3. `PROJECT_PROFILE.md` 의 설치, 실행, 테스트 명령을 실제 운영 기준으로 수정한다.
4. 루트 `GEMINI.md` 의 기본 명령과 문서 경로가 맞는지 확인한다.
5. 첫 실제 작업을 오늘 날짜 backlog 에 등록하고, 세션 종료 전에 handoff 를 갱신한다.

## 5. Gemini CLI 에서 첫 세션 시작하는 방법

- 먼저 `GEMINI.md` 를 기준으로 현재 저장소 규칙을 읽는다.
- 이어서 아래 세 문서를 순서대로 읽는다.
- `ai-workflow/memory/session_handoff.md`
- `ai-workflow/memory/work_backlog.md`
- `ai-workflow/memory/PROJECT_PROFILE.md`
- 기존 프로젝트 도입 직후라면 `ai-workflow/memory/repository_assessment.md` 도 함께 읽는다.

## 6. 적용 후 확인 체크리스트

- `GEMINI.md` 가 존재한다.
- `ai-workflow/memory/` 문서 세트가 존재한다.
- profile 문서의 명령이 실제 저장소 기준으로 채워져 있다.
- 첫 backlog 항목과 handoff 가 비어 있지 않다.
- Gemini CLI 가 읽어야 할 시작 문서 경로가 팀 내에서 합의되어 있다.

## 다음에 읽을 문서

- Gemini CLI 패키지 안내: [./README.md](./README.md)
- 하네스 허브: [../README.md](../README.md)
- 도입 분기 가이드: [../../core/workflow_adoption_entrypoints.md](../../core/workflow_adoption_entrypoints.md)
