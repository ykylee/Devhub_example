---
title: test_cases
type: source
tags: [domain, test_cases.md, project-devhub]
sources: [raw/projects/devhub/docs/domain/integration-registry/test_cases.md]
git_commit: fb3894f7
git_branch: chore/260622-wiki-drift-cleanup-3
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
last_touched: 2026-06-22T06:03:34Z
mirror_dirty: |
related: [none]
status: draft
contradictions: [none]
---

# M4 External Integration 테스트 전략/케이스 초안

- 문서 목적: 외부 시스템 연동(Integration) 도메인의 테스트 범위와 우선순위 TC를 정의해 구현 단계의 품질 기준선을 제공한다.
- 범위: Provider/Biding/Ingest/HomeLab API 및 동기화 파이프라인의 단위/통합/E2E 테스트 초안.
- 대상 독자: Backend/Frontend 개발자, QA, 운영 담당자, AI 에이전트.
- 상태: accepted
- 최종 수정일: 2026-05-16
- 관련 문서: [requirements.md](../requirements.md), [planning/system_usecases.md](../planning/system_usecases.md), [architecture.md](../architecture.md), [backend_api_contract.md](../backend_api_contract.md), [e2e_testing_strategy.md](./e2e_testing_strategy.md)

## 1. 기능 맵 (REQ/UC 기준)

| 기능 ID | 설명 | REQ | UC |
| --- | --- | --- | --- |
| F-INT-PROVIDER | Provider 등록/수정/비활성화/조회 | REQ-FR-INT-001,002,010 | UC-INT-01,02,10 |
| F-INT-INGEST | Webhook/Pull 수집과 중복 처리 | REQ-FR-INT-003,004,005 | UC-INT-03,04,05,06 |
| F-INT-BINDING | Scope(Platform/Project) binding 정책 | REQ-FR-INT-007,011 | UC-INT-11 |
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
| TC-INT-HOMELAB-03 | P1 | E2E | 토폴로지 조회 화면에서 node/service 표시 | nodes/services/edges 일관 노출 (sprint `claude/work_260518-n` 활성화 — `/admin/topology-v2`) |
| TC-INT-RESILIENCE-01 | P1 | IT | 특정 provider 연속 실패 | 해당 provider만 `degraded`, 타 provider 정상 |
| TC-INT-FRONTEND-LIST-01 | P1 | E2E | system_admin 이 `/admin/settings/integrations` 접근 → Provider 목록 노출 | ProviderTable 렌더 + provider_type/auth_mode/sync_status badge 정합 |
| TC-INT-FRONTEND-CREATE-01 | P1 | E2E | Register Provider 모달에서 provider_key/type/display_name/auth_mode/credentials_ref 입력 후 등록 | API-70 호출 → 목록 row 추가 + sync_status `requested` 노출 |
| TC-INT-FRONTEND-EDIT-01 | P1 | E2E | Provider row 의 Edit → display_name/credentials_ref/enabled 갱신 | API-71 호출 → row 즉시 갱신 (provider_key/type/auth_mode 는 immutable) |
| TC-INT-FRONTEND-SYNC-01 | P1 | E2E | Provider row 의 Sync 버튼 click | API-72 호출 → sync_status badge `requested` 로 즉시 전이 + last_sync_at 갱신 |
| TC-INT-FRONTEND-RBAC-01 | P2 | E2E | non-system_admin (developer / manager) 이 `/admin/settings/integrations` 직접 접근 | AuthGuard + layout.tsx 의 `isSystemAdmin` 가드로 default landing 으로 redirect |
| TC-INT-FRONTEND-DELETE-01 | P1 | E2E | Provider row 의 Delete 버튼 → DestructiveConfirmModal 확인 → binding 없는 provider 삭제 | API-80 호출 + 200 응답 + table row 제거 + 성공 toast |
| TC-INT-FRONTEND-DELETE-NEG-01 | P1 | E2E | binding 이 있는 provider 삭제 시도 | API-80 호출 → 409 `integration_provider_has_bindings` → 에러 toast + row 유지 |
| TC-INT-FRONTEND-BIND-LIST-01 | P1 | E2E | system_admin 이 `/admin/settings/integration-bindings` 접근 → BindingsTable 또는 empty state 렌더 | scope/provider/external_key/policy 컬럼 노출, scope filter dropdown 동작 |
| TC-INT-FRONTEND-BIND-CREATE-01 | P1 | E2E | Create Binding 모달에서 scope_type/scope_id/provider/external_key/policy 입력 후 등록 | API-75 호출 → 200/201 → table row 추가 + scope 별 badge + provider display_name 정합 |
| TC-INT-FRONTEND-BIND-RBAC-01 | P2 | E2E | non-system_admin (developer / manager) 이 `/admin/settings/integration-bindings` 직접 접근 | AuthGuard + layout.tsx 의 `isSystemAdmin` 가드로 default landing 으로 redirect |
| TC-INT-FRONTEND-TOPOLOGY-V2-NAV-01 | P2 | E2E | `/admin` 페이지의 "Topology v2" link → /admin/topology-v2 navigation | URL 변경 + v2 헤딩 노출 |
| TC-INT-FRONTEND-TOPOLOGY-V2-RBAC-01 | P2 | E2E | non-system_admin 의 `/admin/topology-v2` 직접 접근 | AuthGuard + role-routing default landing 으로 redirect |

### 3.1 Frontend 카버리지 매핑 (sprint `claude/work_260518-h`)

`frontend/tests/e2e/admin-integrations.spec.ts` 가 source-of-truth. LIST/CREATE/EDIT/SYNC 4건은 mega lifecycle test 한 케이스로 묶고 `test.step()` 으로 분리. RBAC 는 별도 test (developer 로그인 컨텍스트 분리 필요).

| TC ID | spec ts test 명 / step | 상태 |
| --- | --- | --- |
| `TC-INT-FRONTEND-LIST-01` | `provider lifecycle` step 1 | ✅ active |
| `TC-INT-FRONTEND-CREATE-01` | `provider lifecycle` step 2 (Register 모달 → API-70) | ✅ active |
| `TC-INT-FRONTEND-EDIT-01` | `provider lifecycle` step 3 (Edit 모달 + immutable field 가드 검증 — provider_key 입력 부재 + type/auth_mode disabled) | ✅ active |
| `TC-INT-FRONTEND-SYNC-01` | `provider lifecycle` step 4 (Sync 버튼) | ✅ active |
| `TC-INT-FRONTEND-RBAC-01` | `non-system_admin redirect` | ✅ active (developer 로그인 → `/developer` 로 redirect 검증) |
| `TC-INT-FRONTEND-DELETE-01` | `provider lifecycle` step 5 (DELETE 버튼 → DestructiveConfirmModal → API-80) | ✅ active (sprint `claude/work_260518-j`, mega test cleanup → 명시 delete 로 전환) |
| `TC-INT-FRONTEND-DELETE-NEG-01` | UT/IT 영역 (binding 시드 후 DELETE → 409) — backend `TestDeleteIntegrationProvider_HasBindings` 가 cover. E2E carve out. | 🟡 UT only (backend test 활성) |
| `TC-INT-FRONTEND-BIND-LIST-01` | `admin-integration-bindings.spec.ts` 의 `bindings lifecycle` step 2 (BindingsTable / empty state 렌더) | ✅ active (sprint `claude/work_260518-m`) |
| `TC-INT-FRONTEND-BIND-CREATE-01` | `admin-integration-bindings.spec.ts` 의 `bindings lifecycle` step 3 (Create Binding 모달 → API-75 호출 검증 + table row + display_name 매핑) | ✅ active (sprint `claude/work_260518-m`) |
| `TC-INT-FRONTEND-BIND-RBAC-01` | `admin-integration-bindings.spec.ts` 의 `non-system_admin redirect` test | ✅ active (sprint `claude/work_260518-m`) |
| `TC-INT-HOMELAB-03` | `admin-topology-v2.spec.ts` 의 `topology v2 + nodes/services/snapshot 메타 렌더` test — React Flow canvas 또는 empty state (`.or()` chain), `Services (N)` heading, `Last snapshot:` 메타 노출 검증 | ✅ active (sprint `claude/work_260518-n`) |
| `TC-INT-FRONTEND-TOPOLOGY-V2-NAV-01` | `admin-topology-v2.spec.ts` 의 `/admin → /admin/topology-v2 navigation` test | ✅ active (sprint `claude/work_260518-n`) |
| `TC-INT-FRONTEND-TOPOLOGY-V2-RBAC-01` | `admin-topology-v2.spec.ts` 의 `non-system_admin redirect` test | ✅ active (sprint `claude/work_260518-n`) |

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
| 2026-05-18 | sprint `claude/work_260518-g` — TC-INT-FRONTEND-{LIST,CREATE,EDIT,SYNC,RBAC}-01 신규 발급 (External Integration frontend 진입점 1차, `/admin/settings/integrations` 페이지). E2E spec ts 작성은 후속 carve out. |
| 2026-05-18 | sprint `claude/work_260518-h` — TC-INT-FRONTEND-* 5건의 E2E spec ts 활성화. `frontend/tests/e2e/admin-integrations.spec.ts` 신규 (2 test: mega lifecycle 4 step + RBAC negative). §3.1 카버리지 매핑 표 신설 — TC ID ↔ spec ts step 1:1. carve out → active 전환. |
| 2026-05-18 | sprint `claude/work_260518-j` — **API-80 DELETE endpoint + frontend Delete UI**. TC-INT-FRONTEND-DELETE-01 (E2E active — mega test step 5) + DELETE-NEG-01 (UT only — `TestDeleteIntegrationProvider_HasBindings`) 발급. backend handler 4 신규 unit test (Happy / NotFound / HasBindings / RBAC). frontend: service.deleteProvider + ProviderTable Delete 버튼 + DestructiveConfirmModal. spec ts cleanup 부분이 명시 DELETE 호출로 전환. |

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
| TC-INT-BINDING-02 | PASS | `TestCreateIntegrationBinding_ForbiddenForDeveloperRole` (403) |
| TC-INT-HOMELAB-01 | PASS | `TestInfraServicesSnapshotIngestAndRead` (202 + 서비스 조회 반영) |
| TC-INT-HOMELAB-02 | PASS | `TestInfraServicesSnapshotRejectsUnauthorized` (401) |
| TC-INT-HOMELAB-03 | PASS | `TestInfraTopologyV2ContainsMeta`, `TestInfraServicesHydratesFromPersistedSnapshot` (meta + 영속 스냅샷 복원) |
| TC-INT-RESILIENCE-01 | PASS | `TestIntegrationProviderWebhook_DuplicateDeliveryConflict`(409) + `TestIntegrationProviderWebhook_InvalidSignatureMarksOnlyTargetProviderDegraded`(provider 격리 degraded 전이) |
