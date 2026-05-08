# Schemas

- 문서 목적: 출력 샘플과 smoke test 가 공유하는 정적 계약 파일의 위치와 역할을 설명한다.
- 범위: 현재 포함된 schema 초안, 사용 방식, 한계
- 대상 독자: AI workflow 설계자, 구현자, 테스트 작성자
- 상태: draft
- 최종 수정일: 2026-04-23
- 관련 문서: `../core/output_schema_guide.md`, `../tests/README.md`, `../workflow_kit/README.md`

## 1. 목적

이 디렉터리는 현재 저장소의 실행형 프로토타입 출력 계약을 정적 파일과 generated JSON Schema 형태로 함께 유지하기 위한 위치다.

현재 단계에서는 런타임 계약을 단일 출처로 두고, 사람이 읽기 쉬운 정적 계약과 기계 검증용 generated schema 를 함께 유지하는 것을 우선한다.

- output sample 과 smoke test 가 같은 계약 파일을 공유
- contract 변경이 문서와 코드 사이에서 쉽게 보이도록 유지
- runtime contract 에서 generated JSON Schema draft 를 재생성할 수 있게 유지
- 이후 JSON Schema 또는 MCP/server 응답 계약으로 승격하기 쉬운 중간 산출물 확보

## 2. 현재 포함된 schema 초안

- [output_sample_contracts.json](./output_sample_contracts.json)
- [generated_output_schemas.json](./generated_output_schemas.json)
- [read_only_transport_descriptors.json](./read_only_transport_descriptors.json)
- [read_only_harness_mcp_examples.json](./read_only_harness_mcp_examples.json)
- [read_only_jsonrpc_fixtures.json](./read_only_jsonrpc_fixtures.json)

이 파일들은 아래 정보를 담는다.

- 공통 필수 필드
- sample family 별 success/error 필수 필드
- read-only MCP 5종과 주요 skill/runner family 의 nested output field shape
- `tool_version` 단일 출처가 `workflow_kit.__version__` 임을 가리키는 메타데이터
- runtime 계약에서 자동 생성한 JSON Schema draft 묶음
- read-only descriptor 를 하네스별 MCP 설정 예시로 변환한 manual-review-only draft
- read-only JSON-RPC draft bridge 의 대표 request/response fixture

추가 메모:

- 현재 정적 계약 파일은 `tool_version` 값을 직접 반복 정의하지 않고, `workflow_kit.__version__` 를 단일 출처로 삼는다는 사실만 메타데이터로 기록한다.
- 실제 JSON 샘플 값 일치 여부는 `tests/check_output_samples.py` 가 런타임 `workflow_kit.__version__` 과 직접 비교해 검증한다.

## 3. 현재 사용 방식

- `tests/check_output_samples.py` 가 이 파일을 읽어 대표 JSON 샘플의 핵심 필드 계약을 검증한다.
- `workflow_kit/common/output_contracts.py` 는 같은 계약을 Python 코드에서 재사용하기 위한 런타임 표현이자 단일 출처다.
- 현재는 read-only MCP 5종과 `session_start`, `backlog_update`, `create_backlog_entry`, `doc_sync`, `merge_doc_reconcile`, `code_index_update`, `validation_plan`, `demo_workflow`, `existing_project_onboarding` 의 nested list/object shape 도 정적 schema 와 런타임 validator 가 함께 검증한다.
- 주요 error payload 의 `source_context` shape 는 `error_field_shapes` 로 분리해 관리하며, 정의된 family 에 한해 runtime validator 와 generated JSON Schema 가 함께 검증한다.
- read-only bundle entrypoint 자체 오류는 `read_only_entrypoint` family 로 분리해 `action`, `tool`, `payload_json` 중심의 `source_context` shape 를 검증한다.
- `scripts/generate_output_json_schema.py` 는 런타임 계약에서 [generated_output_schemas.json](./generated_output_schemas.json) 을 생성한다.
- `scripts/generate_read_only_transport_descriptors.py` 는 read-only registry 에서 [read_only_transport_descriptors.json](./read_only_transport_descriptors.json) 을 생성한다.
- `scripts/generate_read_only_harness_mcp_examples.py` 는 descriptor 기준으로 [read_only_harness_mcp_examples.json](./read_only_harness_mcp_examples.json) 을 생성한다.
- `scripts/generate_read_only_jsonrpc_fixtures.py` 는 JSON-RPC draft bridge 기준으로 [read_only_jsonrpc_fixtures.json](./read_only_jsonrpc_fixtures.json) 을 생성한다.
- `tests/check_output_json_schema.py` 는 generated schema 파일과 생성 스크립트 출력이 런타임 계약과 같은지 검증한다.
- `tests/check_read_only_transport_descriptors.py` 는 read-only transport descriptor 파일과 생성 스크립트 출력이 registry 와 같은지 검증한다.
- `tests/check_read_only_harness_mcp_examples.py` 는 하네스별 MCP 예시 파일과 생성 스크립트 출력이 descriptor 기준과 같은지 검증한다.
- `tests/check_read_only_jsonrpc_fixtures.py` 는 체크인된 JSON-RPC fixture 와 생성 스크립트 출력이 runtime bridge 결과와 같은지 검증한다.
- `tests/check_generated_schema_validation.py` 는 generated schema draft 로 실제 sample JSON 이 통과하는지 검증한다.
- 두 표현이 어긋나면 smoke test 가 실패하도록 유지하는 것이 권장된다.

## 4. 현재 한계

- generated JSON Schema draft 는 포함되며, `tests/check_generated_schema_validation.py` 가 Draft 2020-12 validator 로 대표 sample 검증까지 수행한다.
- read-only MCP bundle manifest 는 runtime 계약에서 생성한 tool 별 JSON Schema draft 를 `output_schema.json_schema` 로 직접 노출한다.
- 하네스별 MCP 예시는 `workflow_kit.server.read_only_jsonrpc --stdio-lines` draft bridge 를 가리키지만 실제 설정 파일이 아니라 `transport_ready=false` 인 참고 산출물이며, 정식 MCP SDK 서버 루프가 붙기 전에는 자동 적용 대상으로 보지 않는다.
- JSON-RPC fixture 는 SDK transport 승격 전 request/response envelope 비교용 기준선이며, 실제 MCP client 호환성 보장을 뜻하지 않는다.
- 타입 수준 검증은 이전보다 넓어졌지만, 여전히 family 별 핵심 필드와 주요 nested object 중심이다.
- error `source_context` shape 검증은 대표 skill/runner family 와 read-only entrypoint family 에 우선 적용돼 있다.
- nested payload 전체 구조와 enum 제약을 모든 family 에 대해 완전 검증하지는 않는다.

## 5. 다음 확장 후보

- family 별 허용 상태값과 nested object 필드 제약 추가
- runner/skill/MCP 공통 에러 코드 분류 체계 문서화
- generated schema 를 manifest 바깥의 외부 소비 지점에서 읽는 방식 정리

## 다음에 읽을 문서

- 출력 스키마 가이드: [../core/output_schema_guide.md](../core/output_schema_guide.md)
- 테스트 가이드: [../tests/README.md](../tests/README.md)
- package 루트: [../workflow_kit/README.md](../workflow_kit/README.md)
