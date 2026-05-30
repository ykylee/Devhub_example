# Integrated Work Backlog — gemini/work_260530-test-remediation (2026-05-30 EOD, Phase 1 In-Progress)

- 문서 목적: 백엔드 전체 커버리지 격상을 위한 Phase 1 개시 및 1단계 완수 후 백로그 상태.
- 범위: auth-session/view 90.1% 완성, integration-registry/view 24.7% 향상.
- 상태: **2026-05-30 EOD update — Phase 1 1단계 완수 및 2단계 진행 중.**
- 최종 수정일: 2026-05-30 EOD

## 1. 최근 완결된 스프린트 내역

* **sprint `gemini/work_260530-test-remediation`** — 1차 90% 및 2차 100% 테스트 커버리지 추가 보완 스프린트 개시.
  * [x] `auth-session/view/handler_test.go` 내의 AuthenticateActor, GetMe, PatchMe API 핸들러 테스트 대량 보완 (9.9% -> **90.1%** 완성)
  * [x] `integration-registry/view/handler_test.go` 내의 fakeIntegrationStore 임베딩 구축 및 List/Create/Update/Delete Integration 핸들러 테스트 추가 (11.1% -> **24.7%** 향상)

## 2. 잔여 후속 과제 (우선순위 인덱스)

| 우선순위 | 항목 | 사유 |
|---|---|---|
| **Phase 1-2** | `integration-registry/view` 남은 핸들러 보완 | external_task_handler, integration_registry.go 100% 자극 |
| **Phase 1-3** | `dev-request/view` 핸들러 테스트 신설 | CRUD 및 intake_token, metrics 인터셉터 자극 |
| **Phase 1-4** | `repository-integration/view` 핸들러 테스트 신설 | 연결 및 sync status 갱신 자극 |
| **Phase 1-5** | `rbac-permissions/view` 핸들러 테스트 신설 | 역할 할당, 권한 매트릭스 갱신 자극 |
| **Phase 1-6** | `organization-management/view` 핸들러 테스트 신설 | User CRUD, Node Hierarchy, Unit 교체 자극 |
| **Phase 1-7** | `realtime/view` 핸들러 테스트 신설 | WS connection lifecycle, ticket consume 자극 |
| **Phase 2** | Repository 및 스토어 2차 정복 | registry, org-management repo DB 통합 테스트 보강 |
| **Phase 3** | 2차 목표 100% 완전 정복 스프린트 | 남은 극소수 엣지/에러 branch 100% 커버 |
