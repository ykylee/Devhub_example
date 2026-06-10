# DevHub IdP 자산 — Keycloak 단일화 (ADR-0019)

- 문서 목적: DevHub 의 IdP 자산 디렉터리 (`infra/idp/`) 의 현행 구성과 사용 방법을 정의한다.
- 범위: 로컬 테스트 모드 (compose `local-idp` profile) 와 외부 Keycloak 모드의 분기, realm import 파일 사용 SOP.
- 대상 독자: DevHub 개발자, 운영자, 신규 합류자.
- 상태: active
- 최종 수정일: 2026-06-10 (Kratos legacy identity.schema.json archive 이동)
- 관련 문서:
  - [ADR-0019 — Keycloak 단일화](../../docs/adr/0019-keycloak-only-idp.md)
  - [ADR-0018 — 단일 외부 포트 역프록시 정책](../../docs/adr/0018-single-port-reverse-proxy-policy.md)
  - [ADR-0022 — Keycloak 25.0 pin + §3.4 외부 ingress 포트 13000 정합](../../docs/adr/0022-keycloak-version-pin-25-0.md)
  - [docs/setup/keycloak_operations.md](../../docs/setup/keycloak_operations.md)
  - [docs/setup/single_port_deployment.md](../../docs/setup/single_port_deployment.md)
  - [_archive_hydra_kratos/](./_archive_hydra_kratos/) — ADR-0001 시기 Hydra/Kratos 자산 (archive)
  - [_archive_2026-06-10/](./_archive_2026-06-10/) — 2026-06-10 v1.1 sprint -a follow-up PR1 의 KRATOS 잔재 archive (identity.schema.json)

## 1. 구성 파일

| 파일 | 용도 | 모드 |
| --- | --- | --- |
| `keycloak-realm.dev.json` | 로컬 테스트 / CI / smoke 용 realm import. `displayName: "DevHub Local Testing"`. localhost wildcard (`:3000` native + `:8080` 단일포트 시뮬 + `:13000` 사내 ingress reference, [ADR-0022 §3.4](../../docs/adr/0022-keycloak-version-pin-25-0.md#34-외부-ingress-포트-13000-정합)) 포함. | 로컬 모드 only |
| `keycloak-realm.prod.json` | 외부 Keycloak 운영팀 reference 템플릿. 실제 운영 hostname 만 redirect_uri 허용. | 외부 모드 template |
| `sql/003_seed_test_admin.sql` | smoke test 용 `test` 계정 시드 (`users` 테이블, DevHub schema). Idempotent. | dev / smoke |
| `_archive_hydra_kratos/` | ADR-0001 시기 자산 (deprecated, ADR-0019 supersedes). | archive only |
| `_archive_2026-06-10/` | 2026-06-10 v1.1 sprint -a follow-up PR1 의 KRATOS 잔재 archive (`identity.schema.json`). Keycloak 전환 후 미사용. | archive only |

## 2. 두 가지 모드

### 2.1 로컬 테스트 모드 (compose `local-idp` profile)
- 용도: 개발자 워크스테이션 또는 CI 에서 짧게 띄우는 일회용 Keycloak 인스턴스.
- 가동: `docker-compose -f docker-compose.deploy.yml --profile local-idp --profile local-db up`
- 자동 import: `keycloak-realm.dev.json` (compose 의 `KEYCLOAK_REALM_IMPORT_PATH` default 가 dev.json).
- 외부 hostname: `KEYCLOAK_HOSTNAME` 으로 nginx 의 외부 hostname 과 일치시킨다.
- 관계 path: `KC_HTTP_RELATIVE_PATH=/devhub/auth/keycloak` (compose default) — 단일 포트 reverse proxy 와 정합.
- realm 의 redirect_uris 가 localhost wildcard 를 허용하므로 **운영 환경에서는 절대 사용 금지**.

### 2.1.a dogfood 전용 예외
- `docker-compose.colima.yml` 기반 dogfood 스택은 `infra/idp/Dockerfile.keycloak` 을 사용해 커스텀 Keycloak 이미지를 빌드한다.
- 이 경로의 build context 는 `./infra/idp` 로 고정되어 있으며, `keycloak-event-listener-spi/` 자산은 이 디렉터리 기준 상대경로를 전제로 한다.
- 반면 현재 `docker-compose.deploy.yml` 과 GitHub Actions CI 는 stock `quay.io/keycloak/keycloak:26.0` 이미지를 유지한다.
- 즉, 커스텀 SPI 포함 이미지는 현재 dogfood 검증용 경로에만 적용된다.

### 2.2 외부 Keycloak 모드 (compose `local-idp` profile 미활성)
- 용도: 사내 운영팀이 별도로 관리하는 Keycloak 인스턴스 사용.
- compose 의 `keycloak` 서비스는 시작되지 않는다.
- nginx 의 Keycloak upstream 은 `KEYCLOAK_UPSTREAM` 환경변수로 외부 host:port 지정. 예:
  ```
  KEYCLOAK_UPSTREAM=kc.internal.example.com:8443
  ```
  자세한 절차는 [infra/nginx/README.md](../nginx/README.md).
- realm 구성: 운영팀이 `keycloak-realm.prod.json` 을 reference 로 자체 realm 을 발급 / 관리. 본 저장소는 추적성 외에 운영 realm 을 git 으로 관리하지 않는다 (CLAUDE.md "환경 특화 자산은 git 추적 외" 정책 정합).
- backend / frontend env:
  - `DEVHUB_OIDC_ISSUER_URL=https://<external-keycloak>/<relative-path>/realms/devhub`
  - `DEVHUB_KEYCLOAK_ADMIN_URL=https://<external-keycloak>/<relative-path>` (또는 internal-only URL)
  - `NEXT_PUBLIC_OIDC_ISSUER_URL=https://<devhub-host>/devhub/auth/keycloak/realms/devhub` (사용자 브라우저는 항상 DevHub origin 을 거쳐 도달 — 단일 포트 정합)

## 3. realm import 정합 SOP

`docs/setup/keycloak_operations.md` 의 §2 ~ §4 절차를 따른다. 본 README 는 모드 분기와 파일 매핑만 정의한다.

## 4. 단일 포트 컨셉 정합 가드

- 운영 realm 의 `redirectUris` / `webOrigins` 는 단일 포트 entry origin (`https://<devhub-host>/devhub/*`) 만 허용해야 한다.
- localhost wildcard 또는 `["*"]` 는 dev 외 환경에서 금지.
- 자세한 정합성 리뷰: [docs/reports/2026-05-20-network-docker-single-port-review.md](../../docs/reports/2026-05-20-network-docker-single-port-review.md).
