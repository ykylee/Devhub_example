# TASK-006 read-only MCP SDK transport 승격 계획

- 문서 목적: read-only MCP draft bridge 를 공식 MCP SDK transport 로 승격하기 위한 장기 작업 계획을 기록한다.
- 범위: read-only registry, descriptor, JSON-RPC fixture, optional SDK candidate, stdio 검증
- 대상 독자: MCP server 구현자, 하네스 통합 담당자, 리뷰어
- 상태: draft
- 최종 수정일: 2026-04-24
- 관련 문서: `../work_backlog.md`, `../backlog/iyeong-gyun-ui-MacBookAir.local/192.168.0.139/2026-04-24.md`, `../../../core/read_only_mcp_transport_promotion.md`, `../../../core/prototype_promotion_scope.md`

## 1. 작업 개요

- 작업 ID:
- `TASK-006`
- 작업명:
- read-only MCP SDK transport 승격
- 현재 상태:
- `planned`
- 작업 카테고리:
- `development`, `mcp`, `integration`, `release`
- 카테고리 설명:
- 현재 draft/direct-call/read-only bundle 을 공식 MCP SDK 기반 server 경로로 승격할지 판단하고 구현한다.
- 담당:
- 미정

## 2. 배경과 목표

- 배경:
- read-only MCP 는 registry, direct-call entrypoint, JSON-RPC draft bridge, optional SDK candidate 까지 준비돼 있다.
- 목표:
- 유지할 descriptor/output 계약과 바뀔 수 있는 SDK envelope 를 분리한 상태로 공식 SDK transport 를 검증한다.
- 기대 산출물:
- SDK server 경로, stdio smoke, descriptor/fixture 갱신, 승격 판단 기록

## 3. 범위

- 포함 범위:
- read-only MCP 1차 tool 묶음
- official SDK stdio round-trip
- transport descriptor/harness example 검증
- 제외 범위:
- 쓰기 성격 MCP
- 하네스 자동 연결 기본 활성화
- 영향 파일/문서:
- `workflow_kit/server/read_only_mcp_sdk.py`
- `workflow_kit/server/read_only_registry.py`
- `schemas/read_only_*.json`
- `tests/check_read_only_*.py`

## 4. 카테고리별 확장 기록

### 4.1 개발 계획

- 사용자/운영 시나리오:
- 하네스 또는 MCP client 가 read-only tool 을 stdio server 로 호출한다.
- 기능 요구사항:
- tool 목록, input/output schema, readOnlyHint 유지
- 비기능 요구사항:
- draft JSON-RPC fixture 와 SDK envelope 차이를 허용 범위 안에서 설명한다.
- API/UI/데이터 계약:
- registry descriptor 를 단일 출처로 유지한다.

### 4.2 운영/릴리스

- 배포 단위:
- prototype-vNext package 후보
- 롤백 기준:
- SDK transport 가 기존 direct-call 계약의 payload 를 손실하면 승격 보류

## 5. 현재까지 확인한 사실

- 현재 상태:
- optional official SDK candidate 와 stdio smoke 가 존재한다.
- 이번 릴리즈에서는 기본 활성화가 제외됐다.
- 관련 제약:
- MCP SDK 설치와 Python 3.11 환경이 필요하다.

## 6. 결정 기록

- 확정된 결정:
- read-only descriptor 는 registry 를 단일 출처로 유지한다.
- 보류 중인 결정:
- SDK transport 를 기본 경로로 활성화할 시점

## 7. 작업 단계

| 단계 | 상태 | 내용 | 검증/증빙 |
| --- | --- | --- | --- |
| 1 | planned | 기존 candidate 와 승격 spec 재검토 | docs |
| 2 | planned | SDK server 경로 보강 | sdk smoke |
| 3 | planned | fixture 허용 변화 정리 | fixture diff |
| 4 | planned | 하네스 예시 갱신 | harness MCP example tests |

## 8. 검증 계획과 결과

- 검증 계획:
- `python3 tests/check_read_only_mcp_server.py`
- `python3 tests/check_read_only_jsonrpc_bridge.py`
- `python3 tests/check_read_only_jsonrpc_fixtures.py`
- `python3 tests/check_read_only_mcp_sdk_candidate.py`
- `python3 tests/check_read_only_mcp_sdk_stdio.py`
- 실행한 검증:
- 아직 없음
- 미실행 검증과 사유:
- 계획 등록 단계

## 9. 다음 세션 시작 포인트

- 먼저 읽을 파일:
- `core/read_only_mcp_transport_promotion.md`
- `workflow_kit/server/read_only_mcp_sdk.py`
- 바로 할 일:
- SDK candidate 가 현재 registry/output schema 와 어긋나는지 확인한다.
- 주의할 점:
- draft JSON-RPC bridge 를 정식 SDK behavior 로 오해하지 않는다.

## 10. 남은 리스크

- SDK envelope 변화가 fixture 를 크게 흔들 수 있다.
- 하네스별 MCP 설정 방식이 아직 수동 검토 상태다.

## 11. 변경 이력

- `2026-04-24`: 계획 문서 생성
