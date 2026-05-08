# TASK-004 출력 계약 및 실패 스키마 안정화 계획

- 문서 목적: skill/MCP/runner 의 JSON 출력 계약과 실패 payload 규칙을 안정화하기 위한 장기 작업 계획을 관리한다.
- 범위: output sample, runtime contract, generated JSON Schema, error_code/source_context 규칙
- 대상 독자: 개발자, MCP 구현자, 하네스 통합 담당자, 리뷰어
- 상태: draft
- 최종 수정일: 2026-04-24
- 관련 문서: `../work_backlog.md`, `../backlog/iyeong-gyun-ui-MacBookAir.local/192.168.0.139/2026-04-24.md`, `../../../core/output_schema_guide.md`, `../../../schemas/output_sample_contracts.json`

## 1. 작업 개요

- 작업 ID:
- `TASK-004`
- 작업명:
- 출력 계약 및 실패 스키마 안정화
- 현재 상태:
- `planned`
- 작업 카테고리:
- `development`, `schema`, `workflow`
- 카테고리 설명:
- 자동화 소비자가 안정적으로 읽을 수 있도록 출력 payload 의 성공/실패 계약을 고정한다.
- 담당:
- 미정

## 2. 배경과 목표

- 배경:
- runtime contract 와 generated schema 는 존재하지만 nested payload 전체가 완전히 엄격하게 고정되지는 않았다.
- 목표:
- 같은 성격의 경고/실패 필드가 흩어지지 않게 하고, 샘플과 실제 출력의 차이를 줄인다.
- 기대 산출물:
- 갱신된 output schema guide, sample JSON, generated schema, smoke tests

## 3. 범위

- 포함 범위:
- skill 6종, MCP 6종, runner 2종의 대표 성공/실패 payload
- `error_code`, `warnings`, `source_context` shape
- 제외 범위:
- 실제 MCP transport envelope 정식화
- UI 표시 형식
- 영향 파일/문서:
- `workflow_kit/common/output_contracts.py`
- `schemas/generated_output_schemas.json`
- `schemas/output_sample_contracts.json`
- `examples/output_samples/*.json`
- `core/output_schema_guide.md`

## 4. 카테고리별 확장 기록

### 4.1 개발 계획

- 사용자/운영 시나리오:
- 하네스와 MCP client 가 runner 결과를 자동으로 읽고 실패 원인을 분류한다.
- 기능 요구사항:
- family 별 success/error required keys 를 명확히 유지한다.
- 비기능 요구사항:
- 기존 샘플과 smoke 가 같은 계약을 바라본다.
- API/UI/데이터 계약:
- `status`, `tool_version`, `warnings`, `source_context`, `error_code`

### 4.2 코드베이스 분석

- 핵심 파일/모듈:
- `workflow_kit/common/output_contracts.py`
- `scripts/generate_output_json_schema.py`
- `tests/check_output_*.py`
- 불확실한 지점:
- nested payload 를 어느 깊이까지 엄격히 고정할지

## 5. 현재까지 확인한 사실

- 현재 상태:
- generated JSON Schema 와 sample validation 경로가 있다.
- 로드맵상 우선순위 3 작업이다.
- 관련 제약:
- 계약을 너무 빨리 고정하면 프로토타입 개선 속도가 느려질 수 있다.

## 6. 결정 기록

- 확정된 결정:
- `tool_version` 은 `workflow_kit.__version__` 을 단일 출처로 유지한다.
- 보류 중인 결정:
- nested payload strictness 단계

## 7. 작업 단계

| 단계 | 상태 | 내용 | 검증/증빙 |
| --- | --- | --- | --- |
| 1 | planned | 현재 family 별 계약 표 작성 | docs |
| 2 | planned | 실패 sample 누락 확인 | output sample tests |
| 3 | planned | generated schema 갱신 | schema tests |
| 4 | planned | runner nested payload 검증 강화 | generated schema validation |

## 8. 검증 계획과 결과

- 검증 계획:
- `python3 tests/check_output_samples.py`
- `python3 tests/check_output_json_schema.py`
- `python3 tests/check_generated_schema_validation.py`
- 실행한 검증:
- 아직 없음
- 미실행 검증과 사유:
- 계획 등록 단계

## 9. 다음 세션 시작 포인트

- 먼저 읽을 파일:
- `core/output_schema_guide.md`
- `workflow_kit/common/output_contracts.py`
- 바로 할 일:
- error sample coverage 를 목록화한다.
- 주의할 점:
- schema 변경 후 export/harness 산출물에도 영향이 갈 수 있다.

## 10. 남은 리스크

- schema 가 과하게 엄격하면 실제 프로토타입 개발을 방해할 수 있다.
- sample JSON 갱신 누락 시 CI 또는 하네스 export 가 어긋날 수 있다.

## 11. 변경 이력

- `2026-04-24`: 계획 문서 생성
