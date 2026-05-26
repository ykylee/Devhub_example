# Work Backlog - Keycloak Test Server & Real-time Audit SPI Integration

## Backlog Items

| ID | Task Name | Status | Target Branch | Description |
|---|---|---|---|---|
| TASK-1 | 테스트용 Keycloak 인프라 구성 | **DONE** | `gemini/keycloak-test-e2e-push-audit` | Ory 스택 제거 및 Keycloak 26.0 compose/realm 설계 완료 |
| TASK-2 | E2E 테스트 정합 (Playwright) | **DONE** | `gemini/keycloak-test-e2e-push-audit` | global-setup, fixtures, config 내 레거시 Kratos 제거 및 Keycloak 정합 완료 |
| TASK-3 | 감사 로그 Keycloak SPI 및 Webhook 연동 구현 | **DONE** | `gemini/keycloak-test-e2e-push-audit` | Event Listener SPI 구현, Dockerfile.keycloak 빌드, 백엔드 webhook endpoint 및 단위 테스트(100% PASS) 완료 |
| TASK-4 | 로컬 연동 및 E2E 테스트 검증 | **PLANNED** | `gemini/keycloak-test-e2e-push-audit` | 로컬 compose 기동 및 `npm run e2e` 실행 검증 (사용자 머신 위임) |
