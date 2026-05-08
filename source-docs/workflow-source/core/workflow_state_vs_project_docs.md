# Workflow State Vs Project Docs

- 문서 목적: `ai-workflow/memory/*` 와 실제 프로젝트 문서(`docs/...`)의 역할 차이를 한 번에 이해할 수 있게 정리한다.
- 범위: 용어 정의, 읽기 우선순위, 쓰기 대상, 탐색 경계, 대표 예시
- 대상 독자: 개발자, 운영자, AI agent, 하네스 통합 담당자
- 상태: draft
- 최종 수정일: 2026-04-24
- 관련 문서: `./global_workflow_standard.md`, `./workflow_configuration_layers.md`, `./workflow_adoption_entrypoints.md`, `../scripts/README.md`

## 1. 한 줄 요약

- `ai-workflow/memory/*` 는 workflow state docs 다.
- `PROJECT_PROFILE.md` 안의 `docs/...` 경로는 실제 프로젝트 문서 위치다.
- 둘은 모두 참조 대상이지만, 같은 레이어가 아니다.

## 2. 용어 정의

### 2.1 Workflow State Docs

현재 세션 기준선, backlog 상태, handoff, 빠른 복원용 cache 를 담는 문서다.

대표 경로:

- `ai-workflow/memory/PROJECT_PROFILE.md`
- `ai-workflow/memory/session_handoff.md`
- `ai-workflow/memory/work_backlog.md`
- `ai-workflow/memory/backlog/YYYY-MM-DD.md`
- `ai-workflow/memory/state.json`

### 2.2 Project Docs

실제 저장소 운영 문서, 허브 문서, runbook, README, 운영 절차 문서를 뜻한다.

대표 경로:

- `docs/README.md`
- `docs/operations/`
- `docs/operations/backlog/`
- `docs/operations/session_handoff.md`
- 프로젝트별 기타 운영/도메인 문서

## 3. 언제 무엇을 먼저 읽는가

세션 시작과 상태 복원에서는 workflow state docs 를 먼저 읽는다.

권장 순서:

1. `ai-workflow/memory/state.json`
2. `ai-workflow/memory/session_handoff.md`
3. `ai-workflow/memory/work_backlog.md`
4. `ai-workflow/memory/PROJECT_PROFILE.md`
5. 필요 시 최신 `ai-workflow/memory/backlog/YYYY-MM-DD.md`

그 다음 실제 project docs 를 읽는다.

즉:

- workflow state docs 는 "지금 무엇을 하고 있는가" 를 복원하는 레이어
- project docs 는 "프로젝트가 실제로 어떻게 동작하는가" 를 확인하는 레이어

## 4. 무엇을 어디에 쓰는가

기본 원칙:

- 작업 상태 기록, handoff, 세션 기준선, state cache 는 `ai-workflow/memory/*` 에 쓴다.
- 실제 runbook, 허브, README, 운영 절차 문서 수정은 profile 이 가리키는 `docs/...` 에 쓴다.

대표 예시:

- `backlog-update --apply`
  - 기본 write target: `ai-workflow/memory/backlog/*.md`
  - 함께 동기화: `ai-workflow/memory/work_backlog.md`, `ai-workflow/memory/session_handoff.md`, `ai-workflow/memory/state.json`
- `doc-sync`, `code-index-update`
  - 기본 탐색 대상: profile 이 가리키는 실제 project docs
- `merge-doc-reconcile --apply`
  - 제한적 write target: `ai-workflow/memory/session_handoff.md` 운영 메모

## 5. 탐색 경계

`ai-workflow/` 를 아예 무시하면 안 된다. 다만 일반 project discovery 와는 분리해야 한다.

정리하면:

- 세션 복원, workflow 문서 갱신:
  - `ai-workflow/` 적극 참조
- runbook 추천, 허브 stale 탐색, 실제 문서 영향도 판단:
  - project docs 우선
  - `ai-workflow/` 는 메타 레이어로 분리

즉, `ai-workflow/` 는 "참조하지 말라" 가 아니라 "일반 project docs 와 같은 방식으로 섞어 탐색하지 말라" 가 정확한 표현이다.

## 6. 실수하기 쉬운 케이스

- `PROJECT_PROFILE.md` 안에 `백로그 위치: docs/...` 가 적혀 있다고 해서 workflow backlog write target 이 `docs/...` 로 바뀌는 것은 아니다.
- `ai-workflow/memory/backlog/*.md` 는 workflow state backlog 다.
- `docs/operations/backlog/*.md` 는 실제 프로젝트 운영 문서일 수 있지만, 기본 workflow state write target 과는 별개다.

## 7. 구현 기준

- 경로 해석은 두 기준으로 나눈다.
- workflow state docs:
  - `ai-workflow/memory/` 기준
- project docs:
  - 저장소 루트 기준

- runner 와 helper 는 이 둘을 섞어 해석하지 않아야 한다.
- 문서와 하네스 가이드도 같은 용어를 유지해야 한다.

## 다음에 읽을 문서

- 공통 표준: [./global_workflow_standard.md](./global_workflow_standard.md)
- 설정 계층: [./workflow_configuration_layers.md](./workflow_configuration_layers.md)
- 도입 분기: [./workflow_adoption_entrypoints.md](./workflow_adoption_entrypoints.md)
- 스크립트 안내: [../scripts/README.md](../scripts/README.md)
