# 2026-05-21 배포/E2E 실패 분석 + 보완 계획

- 문서 목적: 로컬 배포 스택 기동 및 Playwright e2e 실행 중 확인된 실패 원인을 정리하고 재발 방지 보완 계획을 정의한다.
- 범위: DB 마이그레이션, OIDC issuer/JWKS 정합, Keycloak seed 권한, 로컬/배포 compose 개선.
- 대상 독자: 운영자, 인프라 담당, QA.
- 상태: active
- 최종 수정일: 2026-05-21
- 관련 문서: `docker-compose.deploy.yml`, `docker-compose.local.yml`, `docs/setup/docker-packaging-deployment-guide.md`, `docs/setup/deploy.env.example`

## 1. 실패 요약

### 1.1 DB 스키마 없음으로 e2e seed 실패

- 증상: `relation "users" does not exist`
- 원인: 테스트 DB가 비어 있는 상태에서 app schema 마이그레이션이 자동 수행되지 않음.

### 1.2 OIDC 로그인 루프 / `/api/v1/me` 401

- 증상: 로그인 후 `/developer` 진입 실패, `/auth/login` 재진입 반복.
- backend 로그: `token has invalid issuer`
- 원인: Keycloak token issuer(claim)와 backend 검증용 issuer 환경변수가 불일치.

### 1.3 e2e global setup 의 Keycloak user 생성 403

- 증상: `Keycloak create user alice failed 403`
- 원인: `devhub-backend` service account 최소권한 정책(view/query 중심) 상태에서 테스트 seed가 요구하는 user write 권한 부족.

## 2. 적용한 보완

### 2.1 배포 compose 자동 마이그레이션

- `docker-compose.deploy.yml`에 `db-migrate` 서비스 추가.
- `backend-core`는 `db-migrate` 성공 완료 후 기동하도록 의존성 추가.
- 결과:
  - 빈 DB: 초기 마이그레이션 자동 적용.
  - 기존 DB: 증분 마이그레이션만 적용(데이터 보존).

### 2.2 로컬 compose 정합 개선

- `docker-compose.local.yml`에 `db-migrate` 추가.
- `backend-core`:
  - `DEVHUB_OIDC_ISSUER_URL`를 브라우저 기준 issuer(`localhost:8180`)로 정합.
  - `DEVHUB_OIDC_JWKS_URL`를 internal DNS(`keycloak:8080`)로 분리 설정.
- 목적: issuer claim 검증은 public 값과 맞추고, JWKS fetch는 backend가 접근 가능한 내부 경로 사용.

### 2.3 운영 템플릿/가이드 보강

- `docs/setup/deploy.env.example`:
  - `DEVHUB_OIDC_JWKS_URL` 권장 사용 주석 추가.
- `docs/setup/docker-packaging-deployment-guide.md`:
  - 자동 마이그레이션 정책 명시.
  - issuer(public)/jwks(internal) 분리 권장 명시.

## 3. 재발 방지 운영 체크리스트

1. `docker compose ... config` 단계에서 `DB_URL`, `DEVHUB_OIDC_ISSUER_URL`, `OIDC_REDIRECT_URI` 필수값 확인.
2. `db-migrate` 성공 여부 확인 후 `backend-core` 기동 확인.
3. runtime config 확인:
   - `curl <base>/api/runtime-config` 의 `oidc_issuer_url`, `oidc_redirect_uri` 기대값 검증.
4. backend 인증 로그 확인:
   - `token has invalid issuer` 발생 시 issuer/JWKS 분리 설정 재검토.
5. Keycloak client redirect allowlist 점검:
   - wildcard 금지, callback URI 정확 일치.

## 4. 후속 작업

- e2e 전용 service account 전략 정리:
  - 옵션 A: 테스트 환경에서만 임시 write 권한 부여.
  - 옵션 B: e2e seed 전용 client 분리(권장).
- 배포 preflight 스크립트화:
  - env 검사 + issuer/JWKS reachability 검사 자동화.
