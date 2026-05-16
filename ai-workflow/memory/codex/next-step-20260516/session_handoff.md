# Session Handoff

- 브랜치: `codex/next-step-20260516`
- 날짜: 2026-05-16
- 상태: in_progress

## 핵심 메모
- Integration API-69~75 baseline 구현 + 추적성 동기화 완료.
- API-73 verifier 1차 반영 완료:
  - `hmac_sha256:<secret>` 검증
  - shared-token 상수시간 비교
  - invalid signature 시 `401 integration_webhook_signature_invalid`
- Integration 테스트 문서/실행 보고서 반영:
  - `docs/tests/test_cases_m4_integration.md` 실행 스냅샷 업데이트
  - `docs/tests/reports/report_20260516_m4_integration.md` 추가

## 검증 스냅샷
- `cd backend-core && go test ./internal/httpapi -run 'IntegrationProviderWebhook|CreateIntegrationProvider|ListIntegrationProviders|CreateIntegrationBinding|RoutePermissionTable_CoversAllProtectedV1Routes'` 통과
- `cd backend-core && go test ./...` 통과

## 다음 액션
1. API-76~78 (HomeLab infra) 구현 착수
2. API-73 verifier provider별 전략(`provider_sdk`) 확장
3. Integration role-based E2E (`TC-INT-BINDING-02`) 및 resilience TC 수행
