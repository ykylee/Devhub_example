# Session Handoff

- 브랜치: `codex/next-step-20260516`
- 날짜: 2026-05-17
- 상태: in_progress

## 핵심 메모
- Integration API-69~75 baseline 구현 + 추적성 동기화 완료.
- API-73 verifier 확장 반영 완료:
  - `hmac_sha256:<secret>` 검증
  - shared-token 상수시간 비교
  - `provider_sdk:<provider>:<secret>` 전략 라우팅
  - invalid signature 시 `401 integration_webhook_signature_invalid`
- API-76~78(HomeLab infra) baseline 반영 완료:
  - `GET /api/v1/infra/services`
  - `POST /api/v1/infra/services/snapshot` (agent token 인증)
  - `GET /api/v1/infra/topology/v2`
  - `routePermissionTable` + `publicAPIPaths` 동기화
  - snapshot 영속화 baseline:
    - migration `000028_infra_service_snapshots`
    - `PostgresStore.SaveInfraSnapshot/LoadLatestInfraSnapshot`
    - runtime cache 비어 있을 때 persisted snapshot hydrate
  - adapter skeleton:
    - `internal/integrations/adapters/contract.go`
    - `internal/integrations/adapters/homelab.go`
    - `internal/integrations/adapters/homelab_test.go`
  - adapter 고도화 1차:
    - `HomeLabPuller` contract
    - `NormalizeSnapshot` + health policy override
    - `PullAndIngest` 경로 및 단위테스트 추가
  - runtime config 연동:
    - `DEVHUB_HOMELAB_DEGRADED_STATUSES`
    - `DEVHUB_HOMELAB_PROVIDER_KEY`
    - config → main → router → adapter health policy 주입 경로 반영
  - file puller 1차:
    - `internal/integrations/adapters/homelab_file_puller.go`
    - `internal/integrations/adapters/homelab_file_puller_test.go`
- pull scheduler 연결:
    - `DEVHUB_HOMELAB_PULL_ENABLED`, `DEVHUB_HOMELAB_PULL_INTERVAL`, `DEVHUB_HOMELAB_PULL_FILE`
    - `main.go` 에 feature-flag 기반 `RunHomeLabPullLoop` 실행 경로 연결
- Prometheus 1차:
    - `/metrics` endpoint 노출
    - pull run/latency/snapshot/degraded/last_success 지표 계측 추가
- HTTP puller(P1) 구현:
  - `internal/integrations/adapters/homelab_http_puller.go`
  - `DEVHUB_HOMELAB_PULL_URL`, `DEVHUB_HOMELAB_PULL_TOKEN` config 반영
  - file mode 우선, 미설정 시 HTTP mode fallback 선택 로직(`main.go`) 반영
  - `homelab_http_puller_test.go`로 success/invalid payload/http error/timeout 검증
  - retry/backoff 고도화:
    - `DEVHUB_HOMELAB_PULL_HTTP_RETRY_MAX`
    - `DEVHUB_HOMELAB_PULL_HTTP_RETRY_BACKOFF`
    - 5xx 재시도 후 성공/재시도 소진 테스트 추가
- Prometheus 운영 초안 문서:
  - `docs/planning/prometheus_homelab_alerts.md` (alert rule/dashboard draft)
  - stage/prod 임계값 프로파일 + 튜닝 체크리스트 반영
- Integration 테스트 문서/실행 보고서 반영:
  - `docs/tests/test_cases_m4_integration.md` 실행 스냅샷 업데이트
  - `docs/tests/reports/report_20260516_m4_integration.md` 추가
- role/resilience 테스트 보강:
  - binding 생성 권한 거부(403) 회귀 가드 추가
  - webhook duplicate delivery 충돌(409) 회귀 가드 추가
  - invalid signature 시 target provider만 degraded 전이 + last_error_code 반영 검증
- 추적성 동기화:
  - `docs/traceability/report.md` API-76~78 상태 `activated (baseline)` 반영
  - API-76~78 영속화 반영 (`IMPL-int-store-02`, `IMPL-int-migration-02`, TC/Gap 상태 갱신)
  - HomeLab adapter baseline/후속 범위 (`docs/architecture.md §8.5`) 명시
  - RM/IMPL/UT ID 확장 점검 결과 반영 (RM-INT deferred 유지, IMPL/UT 확장 반영)
  - `infra_integrations` handler에 HomeLab adapter normalize/ingest/load 단계 연결
  - HomeLab pull source 도입 전략 문서 추가 (`docs/planning/homelab_adapter_pull_strategy.md`)

## 검증 스냅샷
- `cd backend-core && go test ./internal/httpapi -run 'IntegrationProviderWebhook|CreateIntegrationProvider|ListIntegrationProviders|CreateIntegrationBinding|RoutePermissionTable_CoversAllProtectedV1Routes'` 통과
- `cd backend-core && go test ./internal/httpapi -run 'CreateIntegrationBinding_ForbiddenForDeveloperRole|IntegrationProviderWebhook_DuplicateDeliveryConflict|IntegrationProviderWebhook'` 통과
- `cd backend-core && go test ./internal/httpapi -run 'IntegrationProviderWebhook_InvalidSignatureMarksOnlyTargetProviderDegraded|IntegrationProviderWebhook'` 통과
- `cd backend-core && go test ./internal/httpapi -run 'InfraServices|InfraTopologyV2|RoutePermissionTable_CoversAllProtectedV1Routes'` 통과
- `cd backend-core && go test ./internal/httpapi -run 'InfraServices|InfraTopologyV2|InfraServicesHydratesFromPersistedSnapshot'` 통과
- `cd backend-core && go test ./...` 통과

## 다음 액션
1. PR 본문(추적성 영향 섹션 포함) 작성 및 제출
2. 운영 환경 Prometheus rule 파일/Alertmanager 라우팅에 임계값 적용
3. Grafana 대시보드 JSON 초안 생성 및 패널별 쿼리 고정
