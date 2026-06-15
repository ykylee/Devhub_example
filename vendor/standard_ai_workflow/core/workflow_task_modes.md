# Workflow Task Modes

- 문서 목적: 작업의 성격(성격)에 따라 최적화된 워크플로우 모드와 에이전트 구성을 정의한다.
- 범위: 모드 정의, 모드별 에이전트 토폴로지, 추천 스킬 및 도구 세트
- 대상 독자: AI 에이전트, 워크플로우 설계자, 프로젝트 관리자
- 상태: stable
- 최종 수정일: 2026-05-02
- 관련 문서: `workflow_agent_topology.md`, `workflow_skill_catalog.md`, `workflow-source/workflow_kit/common/modes/registry.py`

## 1. 개요

표준 워크플로우는 모든 작업에 동일한 절차를 적용하기보다, 작업의 목적(분석, 설계, 구현 등)에 따라 에이전트의 역할 분담과 도구 사용 방식을 최적화함으로써 효율성을 극대화한다. 본 체계는 `workflow_kit/common/modes/registry.py`를 통해 엔진에서 직접 관리되며 자동 감지 기능을 제공한다.

## 2. 작업 모드 정의 (Mode Definitions)

| 모드 | 목적 | 주요 산출물 | 핵심 에이전트 |
| --- | --- | --- | --- |
| **Analysis** (분석) | 코드베이스 구조, 의존성, 로직 파악 | `repository_assessment.md`, `code_index` | `doc-worker` (다수) |
| **Requirements** (요구사항) | 사용자 니즈 정의 및 제약 사항 파악 | `requirements.md`, `stakeholder_list` | `session-orchestrator` |
| **Design** (설계) | 신규 기능 청사진 및 아키텍처 수립 | `technical_spec.md`, `architecture.md` | `main`급 `doc-worker` |
| **Planning** (계획) | 목표를 실행 가능한 태스크로 분해 | `implementation_plan.md`, `backlog` | `backlog-steward` |
| **Implementation** (구현) | 실제 코드 작성 및 단위 테스트 | Code Patches, `test_results` | `code-worker`, `validation-worker` |
| **Refactoring** (리팩터링) | 기능 유지 및 코드 품질 개선 | Refactoring Diffs, Regression Report | `code-worker`, `validation-worker` |

---

## 3. 모드별 운영 가이드

### 3.1 Analysis (코드베이스 분석)
- **전략**: "넓고 얕은 탐색 후 깊은 수직 조사".
- **토폴로지**: 다수의 `small` 모델 기반 `doc-worker`를 활용하여 병렬 탐색.
- **추천 스킬**: `code-index-update`, `doc-sync`.
- **특이사항**: 파일 시스템 탐색(ls, find)과 대량 읽기에 집중.

### 3.2 Design (설계)
- **전략**: "추론 능력 극대화 및 하위 호환성 검토".
- **토폴로지**: 오케스트레이터와 `doc-worker` 모두 `main`급 모델 권장.
- **추천 스킬**: `validation-plan` (사전 검증 설계).
- **특이사항**: 기존 아키텍처와의 정합성 및 확장성 판단에 집중.

### 3.3 Implementation (구현)
- **전략**: "작업 범위 한정(Bounded) 및 즉각적 검증".
- **토폴로지**: `code-worker`와 `validation-worker` 간의 긴밀한 피드백 루프.
- **추천 스킬**: `validation-plan`, `code-index-update`.
- **특이사항**: `apply_workflow_upgrade.py`와 같은 자동화 도구 적극 활용.

### 3.4 Refactoring (리팩터링)
- **전략**: "회귀 테스트(Regression) 강화 및 무결성 보장".
- **토폴로지**: `validation-worker` 비중 확대.
- **추천 스킬**: `merge-reconcile`, `validation-plan`.
- **특이사항**: 수정 전/후의 행위 동일성 검증에 집중.

---

## 4. 모드 전환 메커니즘 (Mode Switching)

- **명시적 전환**: 사용자가 세션 시작 시 또는 백로그 등록 시 모드를 지정함.
- **암묵적 전환**: 오케스트레이터가 작업의 성격을 판단하여 서브 에이전트 토폴로지를 자동으로 조정함.
- **기록**: `state.json` 또는 `session_handoff.md`에 현재 작업의 모드를 기록하여 컨텍스트를 유지함.

## 5. 성숙도 매트릭스와의 연계

- **Alpha**: 모드별 문서 템플릿 제공.
- **Beta**: 모드별 에이전트 토폴로지 자동 할당 엔진 (Orchestrator Logic).
- **Release**: 작업 성과 측정을 통한 모드별 최적 파라미터 자동 튜닝.
