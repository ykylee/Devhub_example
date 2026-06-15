# Workflow Skill Catalog

- 문서 목적: 표준 워크플로우를 skill 로 분리할 때 우선순위와 책임을 정리한다.
- 범위: skill 후보, 역할, 입력 문서, 기대 출력, 도입 우선순위
- 대상 독자: AI agent 설계자, 개발자, 운영자
- 상태: draft
- 최종 수정일: 2026-06-09
- 관련 문서: `workflow_agent_topology.md`, `workflow_mcp_candidate_catalog.md`, `../templates/project_workflow_profile_template.md`, `./orchestrator_subagent_contract_v1.md`

## 1. 핵심 도입 skill (v0.5.10-beta 기준, 11종)

| skill | 역할 | 주요 입력 | 기대 출력 | 구현 상태 | 수동 대체 |
| --- | --- | --- | --- | --- | --- |
| `session-start` | 세션 시작 기준선 복원 | handoff, 백로그, 프로젝트 프로파일 | 현재 상태 요약, 다음 문서 경로 | **Beta** (읽기/요약 안정화) | `global_workflow_standard.md` 의 세션 시작 순서를 수동 수행 |
| `backlog-update` | 작업 등록/갱신 | 오늘 날짜 백로그, 작업 브리핑 | 신규 작업 항목 초안, 상태 갱신 문안 | **Beta** (쓰기 지원) | 백로그 템플릿을 복사해 수동 갱신 |
| `doc-sync` | 문서 영향도와 허브 갱신 판단 | 변경 파일, 기준 문서, 허브 문서 | 영향 문서 후보, 허브 갱신 체크 | **Beta** (`--apply` 지원) | 변경 파일을 기준으로 관련 허브 문서를 수동 확인 |
| `merge-doc-reconcile` | 병합 후 문서 정합성 복구 | 병합 결과, handoff, 인덱스 문서 | 병합 후 재확정 포인트 | **Beta** (정합성 자동 복구) | 병합 후 handoff, 허브, 색인 문서를 수동 재정리 |
| `validation-plan` | 변경 유형별 검증 수준 판단 | 변경 요약, 프로젝트 프로파일 | 검증 계획, 미실행 사유 | **Beta** (`--scaffold` 지원) | 프로젝트 프로파일과 공통 표준을 읽고 수동 판단 |
| `code-index-update` | 색인 문서 갱신 판단 | 변경 파일, 기존 색인 문서 | 갱신 필요 색인 후보 | **Beta** (`--apply` 지원) | 변경 경로를 기준으로 색인 문서를 수동 검토 |

## 2. 운영 보조 및 지능화 skill (Beta 진입, v0.5.7+ 확장)

| skill | 역할 | 주요 입력 | 기대 출력 | 구현 상태 | 수동 대체 |
| --- | --- | --- | --- | --- | --- |
| `workflow-linter` | 워크플로우 문서 정합성 교정 | state.json, handoff, 백로그 | 불일치 리포트 및 교정안 | **Beta** (정합성 자동 검사) | 문서 간 TASK 상태와 링크를 수동 대조 |
| `project-status-assessment` | 프로젝트 도입 성숙도 진단 | 저장소 전체 구조, 테스트, 문서 | 성숙도 리포트, 보강 추천 | **Beta** (자동 진단 및 리포팅) | `repository_assessment.md` 를 수동 작성 |
| `automated-repro-scaffold` | 버그 재현 환경 자동 구축 | 버그 리포트, 기존 테스트 코드 | 재현 테스트 파일(`repro_*.py`) | 프로토타입 (`validation-plan` 연동) | 버그 리포트를 읽고 테스트 코드를 수동 작성 |
| `robust-patcher` | 로컬 LLM 친화적 견고한 파일 수정 | 변경 대상 파일과 수정 명세 | Search-Replace + 퍼지 매칭 기반 패치 | **Beta** (`--apply` 지원) | 일반 편집 도구로 수동 수정 |
| `git-conflict-resolver` | 컨텍스트 기반 Git 충돌 자동 해결 | 충돌 파일, session_handoff 문맥 | 최적 병합 버전 선택 제안 | **Alpha** (프로토타입) | 수동 3-way merge

운영 보조 원칙:

- `backlog-update`, `merge-doc-reconcile` 는 source-of-truth 문서가 준비된 경우 `state.json` 을 자동 재생성해 빠른 세션 기준선을 맞춘다.
- `doc-sync`, `validation-plan`, `code-index-update`, `merge-doc-reconcile` 는 `ai-workflow/` 경로를 workflow 메타 레이어로 보고, 일반 프로젝트 문서/변경 탐색 범위에서는 기본적으로 제외한다.

## 3. 최소 입력 계약

- `session-start`: 현재 프로젝트의 handoff 문서, 백로그 인덱스, 프로젝트 프로파일 경로
- `backlog-update`: 오늘 날짜 백로그 경로 또는 생성 대상 날짜, 작업명, 작업 브리핑
- `doc-sync`: 변경 파일 목록, 기준 문서 후보, 허브 문서 후보
- `merge-doc-reconcile`: 병합 후 상태 문서와 허브 문서 경로

## 4. 상세 스펙 진행 상태

- `session-start`: 상세 입력/출력 계약 + 실행형 프로토타입 + 구조화된 실패 출력 시범 패턴 있음
- 참고 문서: [./session_start_skill_spec.md](./session_start_skill_spec.md)
- `backlog-update`: 상세 입력/출력 계약 + 실행형 초안 생성 프로토타입 + 구조화된 실패 출력 패턴 있음
- 참고 문서: [./backlog_update_skill_spec.md](./backlog_update_skill_spec.md)
- `doc-sync`: 상세 입력/출력 계약 + 실행형 읽기 전용 프로토타입 + 구조화된 실패 출력 패턴 있음
- 참고 문서: [./doc_sync_skill_spec.md](./doc_sync_skill_spec.md)
- `merge-doc-reconcile`: 상세 입력/출력 계약 + 실행형 읽기 전용 프로토타입 + 구조화된 실패 출력 패턴 있음
- 참고 문서: [./merge_doc_reconcile_skill_spec.md](./merge_doc_reconcile_skill_spec.md)
- `validation-plan`: 상세 입력/출력 계약 + 실행형 읽기 전용 프로토타입 + 구조화된 실패 출력 패턴 있음
- 참고 문서: [./validation_plan_skill_spec.md](./validation_plan_skill_spec.md)
- `code-index-update`: 상세 입력/출력 계약 + 실행형 읽기 전용 프로토타입 + 구조화된 실패 출력 패턴 있음
- 참고 문서: [./code_index_update_skill_spec.md](./code_index_update_skill_spec.md)

## 5. 설계 원칙

- skill 은 정책 원문이 아니라 절차와 판단 순서를 담당한다.
- skill 은 가능하면 프로젝트 프로파일을 읽고 분기해야 한다.
- tool 이 없어도 최소 수동 절차는 남아 있어야 한다.

## 5.1 Contract v1 통합 (v0.5.6+)

v0.5.6 부터 orchestrator ↔ sub-agent 위임 시 contract v1 enforcement 가 적용된다. skill 이 sub-agent 로 위임될 때 아래 원칙을 따른다.

- `choose_role` / `choose_roles` (v0.5.7 multi-component) 로 단일 또는 다중 sub-agent 선정.
- 각 sub-agent 응답은 contract v1 output envelope (`status`, `error`, `error_code`, `warnings`, `source_context`, `tool_version`) 을 준수해야 한다.
- `output_validator` 가 응답을 자동 검증하며, MUST NOT delegate 7 패턴을 위반하면 거부된다.
- multi-component fan-out 시 `validate_fanin_output` 으로 cross-ref 통합 검증을 수행한다.

상세 spec: [./orchestrator_subagent_contract_v1.md](./orchestrator_subagent_contract_v1.md), wire 가이드: [./orchestrator_contract_v1_wire_guide.md](./orchestrator_contract_v1_wire_guide.md)

## 5.2 Stage Completion 통합 (v0.6.5)

v0.6.5 부터 11종 skill 의 출력은 v0.6.4 의 [Stage Gate Pattern](./stage_gate_pattern.md) 의 `stage_completion` 필드를 포함한다. 이 필드는 다음 stage 로의 진행 gate 역할을 한다.

적용된 7 spec (§4 출력 계약 끝에 `### 4.1. stage_completion` subsection 추가, commit `5b16517`):

| Skill | Stage Name | Next Stage | Spec 보강 | Runtime 적용 |
|---|---|---|---|---|
| `session-start` | `session-start` | (workflow end) | ✅ 5b16517 | ✅ ca7a685 |
| `backlog-update` | `backlog-update` | (workflow end) | ✅ 5b16517 | ✅ ca7a685 |
| `doc-sync` | `doc-sync` | `validation-plan` | ✅ 5b16517 | ✅ ca7a685 |
| `merge-doc-reconcile` | `merge-doc-reconcile` | (workflow end) | ✅ 5b16517 | ✅ ca7a685 |
| `validation-plan` | `validation-plan` | `code-index-update` | ✅ 5b16517 | ✅ ca7a685 |
| `code-index-update` | `code-index-update` | (workflow end) | ✅ 5b16517 | ✅ ca7a685 |
| `automated-repro-scaffold` | `automated-repro-scaffold` | `validation-plan` | ✅ 5b16517 | ✅ 2fab835 (pilot) |

스펙 + runtime 모두 적용 7/7. SKILL.md cross-ref 만 있는 5 skill (runtime script 부재):

| Skill | Stage Name | Next Stage | SKILL.md cross-ref | Runtime 적용 |
|---|---|---|---|---|
| `workflow-linter` | `workflow-linter` | (workflow end) | ✅ 5b16517 | ⏸ runtime script 없음 |
| `project-status-assessment` | `project-status-assessment` | (workflow end) | ✅ 5b16517 | ⏸ runtime script 없음 |
| `memory-freeze` | `memory-freeze` | (workflow end) | ✅ 5b16517 | ⏸ runtime script 없음 |
| `git-conflict-resolver` | `git-conflict-resolver` | (workflow end) | ✅ 5b16517 | ⏸ runtime script 없음 |
| `robust-patcher` | `robust-patcher` | `validation-plan` | ✅ 5b16517 | ⏸ runtime script 없음 |

**누적 12/12 일관성** (7 spec 보강 + 5 SKILL.md cross-ref, 모두 v0.6.5 spec 수준). Runtime 적용 7/7 (spec 보유 skill 전수).

Gate 정책:
- `approval_actor: "user"` mandatory (state 문서 갱신 skill)
- `requested_changes` 비어있고 `approval_timestamp` + `approval_actor` 모두 있어야 다음 stage 진행
- auto-approval 명시적 차단 (`require_explicit_approval` 의 is_state_doc 파라미터)

Runtime migration 가이드 (11종 skill runtime 적용 절차, breaking change 회피 정책):
- [./stage_gate_runtime_migration.md](./stage_gate_runtime_migration.md) (v0.6.5 신규)
- runtime helper: [`../workflow_kit/common/contracts/stage_gate_runtime.py`](../workflow_kit/common/contracts/stage_gate_runtime.py)
- runtime smoke test: [`../tests/check_stage_gate_runtime.py`](../tests/check_stage_gate_runtime.py) (13 test PASS)

상세: [./stage_gate_pattern.md](./stage_gate_pattern.md), [./output_schema_guide.md §3.4](./output_schema_guide.md), [`../workflow_kit/common/contracts/stage_gate.py`](../workflow_kit/common/contracts/stage_gate.py)

## 6. 권장 skill 묶음

이번 릴리즈에서는 개별 skill 자체보다 “어떤 순서로 묶어 쓰는가” 를 더 중요하게 본다.

### 6.1 기존 프로젝트 첫 세션 묶음

1. `session-start`
2. `validation-plan`
3. `code-index-update`

목적:

- 현재 handoff/backlog 기준선 복원
- 첫 검증 계획 수립
- README/허브/index 재확인 후보 정리

### 6.2 일상 작업 진행 묶음

1. `backlog-update`
2. `doc-sync`

목적:

- 오늘 작업 등록/갱신
- 영향 문서 후보와 허브 갱신 포인트 정리

### 6.3 병합 후 정리 묶음

1. `merge-doc-reconcile`
2. 필요 시 `doc-sync`

목적:

- 병합 후 handoff/backlog/index 정합성 회복
- 허브 문서 누락 여부 재확인

### 6.4 이번 릴리즈 기준 제외 범위

- MCP 를 기본 진입 경로로 두는 skill 구성
- 하네스 자동 연결을 전제로 한 tool-first 소비 경로

즉, 이번 릴리즈에서는 skill 을 사람이 읽는 workflow 순서와 runner 결과 소비 순서에 맞춰 배치하는 것이 우선이다.

## 다음에 읽을 문서

- `session-start` 상세 스펙: [./session_start_skill_spec.md](./session_start_skill_spec.md)
- `backlog-update` 상세 스펙: [./backlog_update_skill_spec.md](./backlog_update_skill_spec.md)
- `doc-sync` 상세 스펙: [./doc_sync_skill_spec.md](./doc_sync_skill_spec.md)
- `merge-doc-reconcile` 상세 스펙: [./merge_doc_reconcile_skill_spec.md](./merge_doc_reconcile_skill_spec.md)
- `validation-plan` 상세 스펙: [./validation_plan_skill_spec.md](./validation_plan_skill_spec.md)
- `code-index-update` 상세 스펙: [./code_index_update_skill_spec.md](./code_index_update_skill_spec.md)
- MCP 후보 카탈로그: [./workflow_mcp_candidate_catalog.md](./workflow_mcp_candidate_catalog.md)
- agent 토폴로지: [./workflow_agent_topology.md](./workflow_agent_topology.md)
