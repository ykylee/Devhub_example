# M4 External Integration 테스트 전략/케이스 초안

- 문서 목적: 외부 시스템 연동(Integration) 도메인의 테스트 범위와 우선순위 TC를 정의해 구현 단계의 품질 기준선을 제공한다.
- 범위: Provider/Biding/Ingest/HomeLab API 및 동기화 파이프라인의 단위/통합/E2E 테스트 초안.
- 대상 독자: Backend/Frontend 개발자, QA, 운영 담당자, AI 에이전트.
- 상태: draft
- 최종 수정일: 2026-05-16
- 관련 문서: [requirements.md](../requirements.md), [planning/system_usecases.md](../planning/system_usecases.md), [architecture.md](../architecture.md), [backend_api_contract.md](../backend_api_contract.md), [e2e_testing_strategy.md](./e2e_testing_strategy.md)

## 1. 기능 맵 (REQ/UC 기준)

| 기능 ID | 설명 | REQ | UC |
| --- | --- | --- | --- |
| F-INT-PROVIDER | Provider 등록/수정/비활성화/조회 | REQ-FR-INT-001,002,010 | UC-INT-01,02,10 |
| F-INT-INGEST | Webhook/Pull 수집과 중복 처리 | REQ-FR-INT-003,004,005 | UC-INT-03,04,05,06 |
| F-INT-BINDING | Scope(Application/Project) binding 정책 | REQ-FR-INT-007,011 | UC-INT-11 |
| F-INT-HOMELAB | Node/Service snapshot 수집 및 토폴로지 조회 | REQ-FR-INT-008,009 | UC-INT-08,09 |
| F-INT-RESILIENCE | 장애 격리/복구 및 감사 추적 | REQ-NFR-INT-002,004,005 | UC-INT-12,13,14 |

## 2. 테스트 계층 전략

1. 단위 테스트(UT): adapter normalize, idempotency key 생성, 상태 전이 규칙 검증.
2. 통합 테스트(IT): DB 저장소 + handler 조합으로 Provider/Binding/HomeLab API 계약 검증.
3. E2E 테스트(TC): UI 또는 API 흐름 기준으로 등록→동기화→조회→장애표시 시나리오 검증.

## 3. 우선 테스트 케이스 (P0/P1)

| TC ID | 우선순위 | 계층 | 시나리오 | 기대 결과 |
| --- | --- | --- | --- | --- |
| TC-INT-PROVIDER-01 | P0 | IT | `POST /api/v1/integration/providers` 정상 등록 | 201 + `sync_status=requested` |
| TC-INT-PROVIDER-02 | P0 | IT | 동일 `provider_key` 중복 등록 | 409 `integration_provider_conflict` |
| TC-INT-INGEST-01 | P0 | UT/IT | 동일 delivery 재수신 | 202/409 정책에 맞게 idempotent 처리 |
| TC-INT-INGEST-02 | P0 | IT | webhook 서명 오류 | 401 `integration_webhook_signature_invalid` |
| TC-INT-BINDING-01 | P0 | IT | Application scope binding 생성 | 201 + scope/provider 매핑 저장 |
| TC-INT-BINDING-02 | P1 | IT | 권한 없는 role의 binding 생성 시도 | 403 |
| TC-INT-HOMELAB-01 | P0 | IT | snapshot ingest 정상 입력 | 202 + ingest_id 반환 |
| TC-INT-HOMELAB-02 | P0 | IT | 필수 필드 누락 snapshot | 422 `infra_snapshot_invalid` |
| TC-INT-HOMELAB-03 | P1 | E2E | 토폴로지 조회 화면에서 node/service 표시 | nodes/services/edges 일관 노출 |
| TC-INT-RESILIENCE-01 | P1 | IT | 특정 provider 연속 실패 | 해당 provider만 `degraded`, 타 provider 정상 |

## 4. E2E 시나리오 초안

| 시나리오 ID | 설명 | 사전 조건 |
| --- | --- | --- |
| E2E-INT-01 | system_admin 이 Provider 등록 후 목록에서 상태 확인 | OIDC 로그인(system_admin), 기본 시드 |
| E2E-INT-02 | webhook ingest 후 Integration 이벤트 히스토리 반영 확인 | test provider + 서명키 시드 |
| E2E-INT-03 | HomeLab snapshot 전송 후 topology v2 조회 | agent 토큰 시드 + 노드/서비스 fixture |
| E2E-INT-04 | provider 장애 발생 시 degraded 배지 노출 | 실패 응답을 반환하는 mock provider |

## 5. 데이터/환경 지침

1. 테스트 provider key는 `*-test` suffix를 사용해 운영 키와 분리한다.
2. webhook replay 테스트는 고정 `delivery_id` fixture를 재사용한다.
3. HomeLab snapshot fixture는 최소 1 node + 2 service 조합을 기본으로 둔다.

## 6. 보고서 연계

- 실행 결과 보고서는 `docs/tests/reports/report_YYYYMMDD_m4_integration.md` 형식으로 기록한다.
- 실패 케이스는 `TC ID | 실패 원인 | 재현 절차 | 조치` 4열을 반드시 포함한다.

## 7. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-15 | 초안 작성 — Integration 도메인 테스트 계층/우선 TC/E2E 시나리오 정의. |
| 2026-05-16 | API-69~75 baseline 구현 기준 실행 스냅샷 반영 (IT 중심), E2E 미진입 항목 명시. |

## 8. 실행 스냅샷 (2026-05-16)

- 실행 환경: `backend-core` 로컬 테스트
- 실행 명령:
  - `go test ./internal/httpapi -run 'IntegrationProviderWebhook|CreateIntegrationProvider|ListIntegrationProviders|CreateIntegrationBinding|RoutePermissionTable_CoversAllProtectedV1Routes'`
  - `go test ./...`
- 결과: PASS

| TC ID | 결과 | 근거 |
| --- | --- | --- |
| TC-INT-PROVIDER-01 | PASS | `TestCreateIntegrationProvider_Happy` |
| TC-INT-PROVIDER-02 | PASS | `TestCreateIntegrationProvider_Duplicate` |
| TC-INT-INGEST-01 | PASS | `TestIntegrationProviderWebhook_Happy` (ingest accepted) |
| TC-INT-INGEST-02 | PASS | `TestIntegrationProviderWebhook_InvalidSignature` (401) |
| TC-INT-BINDING-01 | PASS | `TestCreateIntegrationBinding_Happy` |
| TC-INT-BINDING-02 | 미진입 | role별 e2e 권한 시나리오는 UI/플로우 구현 후 수행 |
| TC-INT-HOMELAB-01..03 | 미진입 | API-76..78 draft 단계 (미구현) |
| TC-INT-RESILIENCE-01 | 미진입 | provider failover/degraded 시뮬레이션 후속 |
