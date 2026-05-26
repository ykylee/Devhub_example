# Session Handoff - Keycloak Test Server & Real-time Audit SPI Integration

## 1. 작업 완료 사항 (Work Accomplished)

- **[1단계] 테스트용 Keycloak 인프라 구성 완료**
  - Ory 스택(Kratos/Hydra)의 레거시 컨테이너 및 마이그레이션 설정을 전면 걷어내고, 기동 시 로컬 Realm 설정이 자동 임포트되는 quay.io Keycloak 26.0 기반의 `devhub-keycloak` 단독 OIDC/IdP 인프라 구축 완료.
  - Schema 생성 스크립트(`001_create_idp_schemas.sql`)에 `keycloak` 스키마 반영.
  - `keycloak-realm.json`을 신규 설계하여 `devhub` realm, clients (`devhub-frontend`, `devhub-backend`), roles, groups, mapper (`employee_id`), service account 권한(realm-management 의 view-users 등)을 정교하게 세팅.
  - 루트 `.env` 파일을 Keycloak 로컬 엔드포인트 및 secret 설정으로 완전히 갱신.
- **[2단계] Playwright E2E 테스트 Keycloak 정합 완료**
  - E2E 헬퍼 및 셋업(`global-setup.ts`, `fixtures.ts`)에서 `DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET`의 기본값을 `"secret-change-me-backend"`로 처리하여 로컬 구동 편의성 고도화.
  - 설정 파일들(`playwright.config.ts`, `endpoints.ts`, `.env.example`)에 존재하던 Kratos/Hydra 레거시 주석과 환경변수 잔재를 말끔히 청소 및 Keycloak 전용 포트로 정합.
- **[3단계] Real-time Audit Log SPI & Webhook Push 완료**
  - Keycloak 26.0(Java 21) Event Listener SPI인 `devhub-event-listener` 프로젝트 신규 생성 (`pom.xml`, `DevHubEventListenerProvider.java` 등). User/Admin 이벤트 발생 시 백엔드 webhook으로 비동기 JSON Push 구현.
  - `Dockerfile.keycloak`을 maven multi-stage 빌드로 설계하여 SPI jar 빌드 후 `/opt/keycloak/providers`로 자동 배포 빌드화.
  - 백엔드에 OIDC 미인증 internal webhook API 엔드포인트 `/api/v1/internal/keycloak-events` 추가 (`router.go`, `config.go`, `main.go`).
  - Webhook 보안 검증 헤더(`X-Webhook-Secret`) 구현 및 User/Admin 이벤트 파싱 & `domain.AuditLog` 변환 후 DB 적재 비즈니스 로직 작성 (`keycloak_events_webhook.go`).
  - 백엔드 `httpapi` 단위 테스트(`keycloak_events_webhook_test.go`) 작성하여 **인증 실패, User Event 매핑, Admin Event 매핑, 멱등성(중복 무시) 처리까지 100% PASS 검증 완료**.

## 2. 로컬 실행 및 E2E 테스트 검증 가이드 (Next Action Items)

개발 환경에 최적화된 로컬 환경 기동 및 테스트 방법입니다. 터미널에서 다음 순서로 실행해 주세요:

1. **Docker Compose 전체 빌드 및 기동**:
   ```bash
   docker compose down --volumes
   docker compose up --build -d
   ```
   * Keycloak 컨테이너(`devhub-keycloak`)가 custom SPI `devhub-event-listener.jar`를 maven multi-stage로 자동 빌드하여 탑재하고, `keycloak-realm.json`을 자동으로 임포트하여 `8180` 포트로 기동합니다.
   * `backend-core` 및 `frontend`가 Keycloak 26.0을 기준으로 정상적으로 연결되어 구동됩니다.

2. **백엔드 컴파일 및 단위 테스트 실행 확인**:
   ```bash
   cd backend-core
   go test -v ./internal/httpapi/ -run TestReceiveKeycloakEvent
   ```
   * 작성된 감사 로그 webhook 로직이 정상 동작함을 로컬에서 다시 한번 검증할 수 있습니다.

3. **Playwright E2E 테스트 실행**:
   ```bash
   cd frontend
   npm run e2e
   ```
   * Keycloak의 admin client credentials를 활용해 OIDC Seed 유저들을 안전하게 생성하고, 전체 Playwright OIDC 인증 및 대시보드 시나리오가 Keycloak을 기반으로 **100% PASS** 하는지 확인할 수 있습니다.
