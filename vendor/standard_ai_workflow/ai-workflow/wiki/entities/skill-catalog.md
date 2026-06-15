---
type: entity
status: active
last_ingested_from: workflow-source/core/workflow_skill_catalog.md
related_pages: [entities/standard-ai-workflow, entities/workflow-source, concepts/orchestrator-subagent-pattern]
created: 2026-06-12
updated: 2026-06-12
---

# Workflow Skill Catalog

## Role

표준 워크플로우의 11종 skill 카탈로그. 도입 우선순위, 역할, 입력, 기대 출력, 구현 상태, 수동 대체 절차를 단일 표로 정리한다. [[entities/workflow-source]] 의 `workflow-source/core/workflow_skill_catalog.md` (v0.5.10-beta baseline) 가 원본이며, v0.5.6 부터 contract v1 enforcement 가 적용된다.

구성:

- **1차 핵심 6종** — 세션 시작/백로그/문서 동기화/병합 정합성/검증/색인 갱신
- **2차 운영 보조 2종** — 워크플로우 린터, 프로젝트 성숙도 진단
- **3차 지능화/실전 3종** — 버그 재현 스캐폴드, 로컬 LLM 친화 패처, Git 충돌 해결

총 11종. sub-agent 로 위임될 때는 [[concepts/orchestrator-subagent-pattern]] 의 contract v1 envelope 을 따른다.

## Skill List

### 1차 핵심 (6종)

| Skill | Version | Purpose | Contract Level |
| --- | --- | --- | --- |
| `session-start` | v0.5.10-beta | handoff / 백로그 / 프로젝트 프로파일을 읽어 세션 시작 기준선 복원 | L1 — full spec + 실행형 프로토타입 + 구조화된 실패 출력 |
| `backlog-update` | v0.5.10-beta | 오늘 날짜 백로그에 작업 등록/갱신, `state.json` 자동 재생성 | L1 — full spec + 실행형 초안 생성 + 구조화된 실패 출력 |
| `doc-sync` | v0.5.10-beta | 변경 파일 영향도와 허브/색인 문서 갱신 판단 (`--apply`) | L1 — full spec + 실행형 읽기 전용 프로토타입 + 구조화된 실패 출력 |
| `merge-doc-reconcile` | v0.5.10-beta | 병합 후 handoff / 인덱스 / 허브 정합성 복구 | L1 — full spec + 실행형 정합성 자동 복구 + 구조화된 실패 출력 |
| `validation-plan` | v0.5.10-beta | 변경 유형별 검증 수준 판단, 검증 계획 scaffold (`--scaffold`) | L1 — full spec + 실행형 읽기 전용 + 구조화된 실패 출력 |
| `code-index-update` | v0.5.10-beta | 색인 문서 (README / 허브 / index) 갱신 후보 판단 (`--apply`) | L1 — full spec + 실행형 읽기 전용 + 구조화된 실패 출력 |

### 2차 운영 보조 (2종, v0.5.7+ Beta)

| Skill | Version | Purpose | Contract Level |
| --- | --- | --- | --- |
| `workflow-linter` | v0.5.7+ | state.json / handoff / 백로그 정합성 자동 검사 및 교정안 | L2 — Beta 자동 검사, 불일치 리포트 + 교정안 |
| `project-status-assessment` | v0.5.7+ | 저장소 구조/테스트/문서 기반 성숙도 진단 + 보강 추천 | L2 — Beta 자동 진단, 리포트 |

### 3차 지능화/실전 (3종)

| Skill | Version | Purpose | Contract Level |
| --- | --- | --- | --- |
| `automated-repro-scaffold` | v0.5.7+ | 버그 리포트 → `repro_*.py` 재현 테스트 자동 구축 | L3 — 프로토타입, `validation-plan` 연동 |
| `robust-patcher` | v0.5.7+ | 로컬 LLM 친화 Search-Replace + 퍼지 매칭 기반 패치 (`--apply`) | L3 — Beta, `--apply` 지원 |
| `git-conflict-resolver` | v0.5.7+ | session_handoff 문맥 기반 Git 충돌 자동 해결, 최적 병합 버전 제안 | L3 — Alpha 프로토타입 |

수동 대체: 구현 미가용 시 `global_workflow_standard.md` 와 템플릿을 읽고 같은 순서를 사람이 수행. 각 skill 의 원본 spec 은 `workflow-source/core/<skill>_skill_spec.md` 에 있다.

## Common Pattern

11종 모두 다음 두 가지를 공통으로 따른다.

### 1. `tool_version` 단일 출처

- `workflow_kit.__version__` 에서 단일 import. 코드 경로에 hardcoded version 없음.
- 출력 envelope 의 `tool_version` 필드는 이 값을 그대로 사용.
- wheel 패키징 (`tools/check_packaging.py`, v0.5.8+) 과 하네스 export 번들 파일명에 version 반영.

### 2. 구조화된 실패 JSON

모든 skill / MCP / runner 응답은 동일 envelope 을 따른다 (Pydantic v2 기반).

| Field | Type | Required | Meaning |
| --- | --- | --- | --- |
| `status` | enum | yes | `ok` / `partial` / `error` |
| `error` | string | no | 사람이 읽는 에러 설명 |
| `error_code` | string | no | 기계 판독용 코드 (e.g. `workflow_step_failed`, `contract_v1_violation`) |
| `warnings` | string[] | no | 비차단 경고. `check_workflow_linter.py` baseline warning 포함 |
| `source_context` | object | no | 실패한 step 이름, upstream `error_code`, file path, line 등 디버깅 단서 |
| `tool_version` | string | yes | `workflow_kit.__version__` |

runner 는 하위 step 이 `status: "error"` 반환 시 `workflow_step_failed` 로 wrap 하고, 실패 step 이름과 upstream `error_code` 를 `source_context` 에 남긴다. sub-agent 위임 시 동일 envelope 이 contract v1 의 `output_validator` 로 강제 검증된다.

운영 보조 원칙:

- `backlog-update`, `merge-doc-reconcile` 는 source-of-truth 문서가 준비된 경우 `state.json` 자동 재생성
- `doc-sync`, `validation-plan`, `code-index-update`, `merge-doc-reconcile` 는 `ai-workflow/` 를 workflow 메타 레이어로 보고 일반 변경 탐색에서 기본 제외
- sub-agent 응답은 contract v1 의 MUST NOT delegate 7 패턴 위반 시 거부됨

## Version

| 항목 | 값 |
| --- | --- |
| Catalog baseline | v0.5.10-beta (2026-06-09) |
| 1차 6종 spec 상태 | v0.5.10-beta. 실행형 프로토타입 + 구조화된 실패 출력 패턴 동봉 |
| 2차 2종 진입 | v0.5.7+ (Beta). 자동 검사/진단 |
| 3차 3종 진입 | v0.5.7+. `automated-repro-scaffold` 프로토타입, `robust-patcher` Beta (`--apply`), `git-conflict-resolver` Alpha |
| Contract v1 통합 | v0.5.6+ P0 enforcement |
| Multi-component fan-out | v0.5.7+ (catalog 자체는 fan-out 미사용, 위임 시 contract v1 multi-component 적용) |
| Package version | `workflow_kit.__version__` 단일 출처. wheel 포함 |

skill 자체 version string 은 workflow kit version 과 동기화한다. version bump 은 wheel 패키징과 하네스 export 번들 파일명 변경으로 즉시 반영된다.

## Related

- [[entities/standard-ai-workflow]] — 최상위 repo entity
- [[entities/workflow-source]] — `workflow-source/` 트리, catalog 원본 위치
- [[entities/workflow-kit]] — catalog 11종이 import 하는 공통 Python package
- [[concepts/orchestrator-subagent-pattern]] — sub-agent 위임 시 contract v1 enforcement
- [[concepts/contract-v1-output-validation]] — `output_validator` 가 11종 envelope 을 자동 검증
- 원본: `workflow-source/core/workflow_skill_catalog.md`
- spec 원본: `workflow-source/core/session_start_skill_spec.md`, `backlog_update_skill_spec.md`, `doc_sync_skill_spec.md`, `merge_doc_reconcile_skill_spec.md`, `validation_plan_skill_spec.md`, `code_index_update_skill_spec.md`
