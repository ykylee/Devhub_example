# TASK-002 기존 프로젝트 온보딩 자동 루틴 강화 계획

- 문서 목적: 기존 프로젝트 도입 후 `assessment -> session-start -> validation-plan -> code-index-update` 흐름을 안정화하기 위한 장기 작업 계획을 기록한다.
- 범위: existing bootstrap, onboarding runner, session/backlog/profile 연결, 하네스 소비 가이드
- 대상 독자: 개발자, 운영자, AI agent, 리뷰어
- 상태: draft
- 최종 수정일: 2026-04-24
- 관련 문서: `../work_backlog.md`, `../backlog/iyeong-gyun-ui-MacBookAir.local/192.168.0.139/2026-04-24.md`, `../../../core/workflow_kit_roadmap.md`, `../../../core/existing_project_onboarding_contract.md`

## 1. 작업 개요

- 작업 ID:
- `TASK-002`
- 작업명:
- 기존 프로젝트 온보딩 자동 루틴 강화
- 현재 상태:
- `planned`
- 작업 카테고리:
- `workflow`, `development`, `documentation`
- 카테고리 설명:
- 기존 코드베이스에 workflow kit 을 도입할 때 첫 세션을 안정적으로 복원하고 후속 작업을 제안하는 흐름을 강화한다.
- 담당:
- 미정

## 2. 배경과 목표

- 배경:
- 현재 `run_existing_project_onboarding.py` 는 존재하지만, 실제 프로젝트 다건 적용 전이라 분기와 출력 소비 방식이 더 검증되어야 한다.
- 목표:
- 기존 프로젝트 도입 직후 사용자가 어떤 문서와 skill 결과를 어떤 순서로 봐야 하는지 안정화한다.
- 기대 산출물:
- 보강된 onboarding runner, 업데이트된 계약 문서, smoke test, 첫 세션 소비 가이드

## 3. 범위

- 포함 범위:
- `scripts/run_existing_project_onboarding.py`
- `core/existing_project_onboarding_contract.md`
- `core/workflow_adoption_entrypoints.md`
- `tests/check_existing_project_onboarding.py`
- 제외 범위:
- MCP server 정식 승격
- 실제 외부 저장소 파일럿 자체
- 영향 파일/문서:
- `scripts/run_existing_project_onboarding.py`
- `workflow_kit/common/runner.py`
- `examples/output_samples/existing_project_onboarding*.json`
- `schemas/generated_output_schemas.json`

## 4. 카테고리별 확장 기록

### 4.1 개발 계획

- 사용자/운영 시나리오:
- 기존 프로젝트에 bootstrap 을 적용한 직후, agent 가 runner 결과를 읽고 다음 작업을 바로 제안한다.
- 기능 요구사항:
- assessment 추정값, latest backlog, validation-plan, code-index-update 결과를 한 payload 에서 일관되게 연결한다.
- 비기능 요구사항:
- 누락 문서가 있어도 구조화된 error 로 실패해야 한다.
- API/UI/데이터 계약:
- `status`, `warnings`, `source_context`, `execution_trace`, `onboarding_summary` 유지
- 구현 단위:
- runner 입력 검증, step 생략 조건, sample 갱신

### 4.2 문서/워크플로우

- 갱신할 문서:
- `core/existing_project_onboarding_contract.md`
- `core/workflow_adoption_entrypoints.md`
- 사용자 안내 필요 여부:
- 첫 세션에서 `status -> onboarding_summary -> warnings -> validation_plan -> code_index_update` 순서로 읽는 가이드를 유지한다.

## 5. 현재까지 확인한 사실

- 현재 상태:
- 기존 runner 와 smoke test 는 존재하며 통과한다.
- 로드맵상 우선순위 1 작업이다.
- 관련 제약:
- 실제 외부 프로젝트 적용 결과가 부족하다.

## 6. 결정 기록

- 확정된 결정:
- 기존 프로젝트 온보딩은 MCP 연결보다 workflow/skill 묶음 소비를 우선한다.
- 보류 중인 결정:
- 하네스별 소비 UI/요약 형식을 어디까지 표준화할지

## 7. 작업 단계

| 단계 | 상태 | 내용 | 검증/증빙 |
| --- | --- | --- | --- |
| 1 | planned | runner 출력과 계약 문서 차이 점검 | `tests/check_existing_project_onboarding.py` |
| 2 | planned | 누락/부분 입력 시나리오 보강 | error sample |
| 3 | planned | 하네스 소비 순서 가이드 보강 | docs smoke |
| 4 | planned | output sample/schema 갱신 | output/schema tests |

## 8. 검증 계획과 결과

- 검증 계획:
- `python3 tests/check_existing_project_onboarding.py`
- `python3 tests/check_demo_workflow.py`
- `python3 tests/check_output_samples.py`
- 실행한 검증:
- 아직 이 계획 문서 작성 후 별도 구현 검증 없음
- 미실행 검증과 사유:
- 계획 등록 단계이므로 구현 검증은 후속

## 9. 다음 세션 시작 포인트

- 먼저 읽을 파일:
- `core/existing_project_onboarding_contract.md`
- `scripts/run_existing_project_onboarding.py`
- 바로 할 일:
- 계약 문서와 runner output keys 를 비교한다.
- 주의할 점:
- 실제 적용 검증 작업과 섞지 않는다.

## 10. 남은 리스크

- 실제 프로젝트마다 문서 구조가 달라 자동 추정값이 부정확할 수 있다.
- runner 가 너무 많은 세부 정보를 반환하면 첫 세션 브리핑이 장황해질 수 있다.

## 11. 변경 이력

- `2026-04-24`: 계획 문서 생성
