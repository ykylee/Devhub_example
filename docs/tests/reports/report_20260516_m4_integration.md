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
| TC-INT-BINDING-02 | PENDING | 역할별 E2E 권한 검증 미진입 |
| TC-INT-HOMELAB-01..03 | PENDING | API-76..78 미구현 |
| TC-INT-RESILIENCE-01 | PENDING | degraded 전이 시뮬레이션 미구현 |

## 4) 결함 및 관찰 사항

- 이번 사이클에서 blocker 결함은 없음.
- API-73 verifier는 1차 구현(`hmac_sha256:<secret>` + shared token)까지 반영.
- provider별 verifier 전략(`provider_sdk`)과 홈랩 영역(API-76..78)은 다음 단계.
