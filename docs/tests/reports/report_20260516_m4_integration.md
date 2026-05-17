# M4 Integration 실행 보고서 (2026-05-16)

- 범위: API-69~75 baseline (Provider/Binding/Webhook ingest)
- 브랜치: `codex/next-step-20260516`
- 실행자: Codex
- 상태: partial-pass (IT pass, E2E pending)

## 1) 실행 명령

```bash
cd backend-core
go test ./internal/httpapi -run 'IntegrationProviderWebhook|CreateIntegrationProvider|ListIntegrationProviders|CreateIntegrationBinding|RoutePermissionTable_CoversAllProtectedV1Routes'
go test ./...
```

## 2) 결과 요약

- 두 명령 모두 PASS.
- API-69~75 baseline 경로의 handler/store/router/permission 회귀 없음.

## 3) TC 상세

| TC ID | 결과 | 비고 |
| --- | --- | --- |
| TC-INT-PROVIDER-01 | PASS | provider 생성 201 + 필드 검증 |
| TC-INT-PROVIDER-02 | PASS | provider_key 충돌 409 |
| TC-INT-INGEST-01 | PASS | webhook accepted 202 |
| TC-INT-INGEST-02 | PASS | invalid signature 401 + `integration_webhook_signature_invalid` |
| TC-INT-BINDING-01 | PASS | application scope binding 생성 201 |
| TC-INT-BINDING-02 | PASS | developer role binding 생성 차단 403 (`TestCreateIntegrationBinding_ForbiddenForDeveloperRole`) |
| TC-INT-HOMELAB-01 | PASS | snapshot ingest 202 + services 조회 반영 |
| TC-INT-HOMELAB-02 | PASS | unauthorized ingest 401 |
| TC-INT-HOMELAB-03 | PASS | topology meta + persisted snapshot hydrate |
| TC-INT-RESILIENCE-01 | PASS | duplicate delivery 충돌 409 + invalid signature 시 target provider만 degraded 전이 (`TestIntegrationProviderWebhook_InvalidSignatureMarksOnlyTargetProviderDegraded`) |

## 4) 결함 및 관찰 사항

- 이번 사이클에서 blocker 결함은 없음.
- API-73 verifier는 `hmac_sha256`, shared token, `provider_sdk:<provider>:<secret>` 전략을 지원.
- role-based binding denial(403) 및 ingest dedupe 충돌(409) 회귀 가드를 추가.
- invalid signature 누적 경로에서 provider 상태(`sync_status=degraded`, `last_error_code=webhook_signature_invalid`) 반영 + provider 격리 동작 검증 완료.
- HomeLab snapshot 영속화 baseline 반영: `infra_service_snapshots` migration + store load/save + runtime hydrate fallback.
