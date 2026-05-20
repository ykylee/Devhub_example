# Session Handoff — gemini/work_260520-c-216-kratos-keycloak-e2e

- 문서 목적: P1-5 E2E Kratos → Keycloak 실 코드 전환 완료 상태를 인계한다.
- 범위: legacy Ory 잔재 제거, Keycloak 기반 local/CI 환경 정합, dynamic user seed.

## 작업 완료 사항

1. **Legacy Ory 제거**:
   - `scripts/ci-setup.sh`, `infra/idp/scripts/register-devhub-client.*`, `infra/idp/sql/001_create_idp_schemas.sql`, `infra/idp/sql/002_seed_e2e_users.sql` 삭제.
   - `dev-up.sh/ps1`, `dev-down.sh/ps1` 에서 Kratos/Hydra 관련 로직 및 환경변수 제거.
2. **Local/CI 환경 정합**:
   - `docker-compose.local.yml` 을 Keycloak 기반으로 갱신.
   - `ci.yml` 의 e2e path filter 에서 `ci-setup.sh` 제거.
   - `idp-apply-schemas/main.go` 에서 더 이상 유효하지 않은 `hydra`/`kratos` 스키마 체크 로직 제거.
3. **Dynamic User Seed (Option 1)**:
   - `global-setup.ts` 에서 Keycloak Admin API 로 생성된 `sub` (ID) 를 획득.
   - 획득한 ID 를 포함하여 DevHub `users` 테이블에 `idp_subject` 를 직접 UPSERT 하는 dynamic SQL 실행 로직 추가.
   - 이로써 E2E 테스트가 "첫 로그인 시 자동 sync" 에 의존하지 않고도 결정론적(deterministic)으로 동작함.

## 다음 작업 제언

- 본 브랜치의 변경사항을 `main` 에 병합.
- `P1-1` ~ `P1-4` 항목 (Claude/Codex 분담) 진행 상황 모니터링 및 필요 시 지원.
