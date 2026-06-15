# TASK-005 공통 라이브러리 추출 및 스크립트 정리 계획

- 문서 목적: `workflow_kit/common` 을 중심으로 parser/classifier/recommendation 로직을 재사용 가능한 패키지 단위로 정리하는 장기 작업 계획을 기록한다.
- 범위: 공통 파서, 문서 분류, 추천 로직, runner helper, 기존 CLI 재구성
- 대상 독자: 개발자, MCP server 구현자, 리뷰어
- 상태: draft
- 최종 수정일: 2026-04-24
- 관련 문서: `../work_backlog.md`, `../backlog/iyeong-gyun-ui-MacBookAir.local/192.168.0.139/2026-04-24.md`, `../../../core/prototype_promotion_scope.md`, `../../../workflow_kit/README.md`

## 1. 작업 개요

- 작업 ID:
- `TASK-005`
- 작업명:
- 공통 라이브러리 추출 및 스크립트 정리
- 현재 상태:
- `planned`
- 작업 카테고리:
- `development`, `refactor`, `architecture`
- 카테고리 설명:
- CLI, skill, MCP, runner 에 흩어진 공통 로직을 안정적인 package API 로 모은다.
- 담당:
- 미정

## 2. 배경과 목표

- 배경:
- `workflow_kit/common` 추출은 이미 시작됐지만, 일부 로직은 여전히 스크립트에 남아 있다.
- 목표:
- 공통 파서/분류/추천 로직부터 좁게 package API 로 정리한다.
- 기대 산출물:
- 정리된 common module, 함수 단위 테스트 후보, 중복 감소한 CLI scripts

## 3. 범위

- 포함 범위:
- project profile / handoff / backlog parser
- changed-file classification
- doc-sync, validation-plan, code-index 추천 helper
- 제외 범위:
- output contract 전체 재설계
- MCP transport 구현
- 영향 파일/문서:
- `workflow_kit/common/*.py`
- `skills/*/scripts/*.py`
- `mcp/*/scripts/*.py`
- `tests/check_*.py`

## 4. 카테고리별 확장 기록

### 4.1 리팩터링/마이그레이션

- 변경 전략:
- 스크립트 동작을 유지하면서 내부 로직만 common 함수로 이동한다.
- 단계별 전환 계획:
- parser/classifier 부터 시작하고 orchestration helper 는 후순위로 둔다.
- 호환성/롤백 계획:
- 기존 smoke test 를 그대로 통과해야 한다.
- 데이터 또는 설정 영향:
- JSON output 계약은 유지한다.

### 4.2 개발 계획

- 구현 단위:
- parser API 정리
- recommendation helper 정리
- tests 보강
- 호환성 고려:
- CLI 인자와 출력 payload 는 바꾸지 않는다.

## 5. 현재까지 확인한 사실

- 현재 상태:
- `workflow_kit/common` 에 경로, Markdown, docs, project_docs, planning, doc_sync, runner helper 가 있다.
- 승격 범위 문서는 parser/classifier/recommendation 우선 추출을 권장한다.
- 관련 제약:
- 너무 넓은 추출은 출력 계약 안정화 전에는 위험하다.

## 6. 결정 기록

- 확정된 결정:
- 공통 라이브러리 확장은 parser/classifier/recommendation 중심으로 좁게 진행한다.
- 보류 중인 결정:
- 함수 단위 테스트 디렉터리 구조

## 7. 작업 단계

| 단계 | 상태 | 내용 | 검증/증빙 |
| --- | --- | --- | --- |
| 1 | planned | 중복 로직 목록화 | rg/code review |
| 2 | planned | parser/helper API 정리 | smoke tests |
| 3 | planned | skill scripts 의 common 호출 비율 확대 | skill checks |
| 4 | planned | MCP scripts 재사용 정리 | MCP checks |

## 8. 검증 계획과 결과

- 검증 계획:
- `for t in tests/check_*.py; do python3 "$t" || exit 1; done`
- 실행한 검증:
- 아직 없음
- 미실행 검증과 사유:
- 계획 등록 단계

## 9. 다음 세션 시작 포인트

- 먼저 읽을 파일:
- `core/prototype_promotion_scope.md`
- `workflow_kit/common/`
- 바로 할 일:
- 스크립트별 중복 parser/classifier 로직을 목록화한다.
- 주의할 점:
- unrelated refactor 로 확장하지 않는다.

## 10. 남은 리스크

- 리팩터링 범위가 넓어 smoke 실패 원인 추적이 어려워질 수 있다.
- output contract 안정화 작업과 충돌할 수 있다.

## 11. 변경 이력

- `2026-04-24`: 계획 문서 생성
