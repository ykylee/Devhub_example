# ADR-0022: Keycloak 운영 image 버전 pin (25.0)

## 1. 상태
- **상태**: Draft
- **작성일**: 2026-05-22
- **수정일**: 2026-05-22
- **결정 근거 sprint**: `codex/work_260521-c-db-docker-option`
- **관련 문서**: [ADR-0019 Keycloak 단일화](./0019-keycloak-only-idp.md), [`docker-compose.deploy.yml`](../../docker-compose.deploy.yml), [`docs/setup/keycloak_operations.md`](../setup/keycloak_operations.md)

> **Draft 사유**: 본 ADR 은 26.0 → 25.0 retreat 의 결정 시점을 ADR 형식으로 명문화한다. **§3.1 retreat 사유** 항목이 placeholder 상태이며, 사용자가 사내 정합 사유 (호환성/안정성/사내 표준 등) 를 확정한 뒤 `Accepted` 로 승격한다.

## 2. 컨텍스트

### 2.1 ADR-0019 의 version 명시 부재

[ADR-0019 (2026-05-19)](./0019-keycloak-only-idp.md) 는 IdP 를 Keycloak 단일화로 결정했으나, 운영 image **버전 자체는 명시하지 않았다**. 실 운영 결정은 코드 (compose / CI / dev-up.sh) 의 image tag 만으로 표현됐다.

### 2.2 이전 상태 (26.0)

2026-05-19 ~ 2026-05-21 기간 운영 자산은 `quay.io/keycloak/keycloak:26.0` 으로 통일됐다 (`docker-compose.deploy.yml:106`, `.github/workflows/ci.yml:340`, `dev-up.sh:118`). 선정 근거는 `docker-compose.deploy.yml` 주석 — *"version 26.x = LTS 후보 (사내 결정 동반 — sprint -h failover §4 Phase 2 HA 진입 시 재평가)"*. 즉 LTS 후보로서 잠정 채택 + Phase 2 진입 시 재평가하기로 했으나, **사내 환경 실 배포 검증 단계에서 retreat 필요성 식별**.

### 2.3 retreat 식별 시점

- sprint `codex/work_260521-c-db-docker-option` (host build packaging + nginx ingress 정합 검증) 후속.
- 사내 운영 환경 (외부 100.90.113.29:13000 → 호스트 → VM → docker 단일 포트 ingress) 의 deploy smoke test 진행 중 25.0 retreat 결정.
- ADR-0019 의 핵심 결정 (Keycloak 단일 IdP) 은 reverse 하지 않음. **버전 pin 만 변경**.

## 3. 결정

**Keycloak 운영 image 를 `quay.io/keycloak/keycloak:25.0` (마이너 pin) 으로 변경**한다.

### 3.1 retreat 사유

> **(placeholder — 사용자 finalize 필요)** 사내 운영 환경의 specific 호환성/안정성/표준 사유. 후보:
> - 사내 인프라 (JDK 21 vs 17, AD/LDAP federation provider, JDBC driver) 의 25.x 정합
> - 26.x 의 known issue (구체적 이슈 ID 또는 회귀)
> - 사내 IdP 운영 표준 버전이 25.x
> - 외부 사내 Keycloak instance 가 25.x 운영이므로 dev/staging 도 동일 마이너로 정합
>
> 실제 사유 확정 후 본 §3.1 를 채워 `Accepted` 로 승격.

### 3.2 pin 등급

마이너 pin (`25.0`) 채택. 25.0.x 의 patch 자동 흡수 + 26.x 메이저 변경 차단의 balance.

### 3.3 영향 범위 (active code only)

- `docker-compose.deploy.yml:106` — image tag
- `.github/workflows/ci.yml:340` — CI E2E job
- `dev-up.sh:118` — dev 안내 echo

ADR-0019 + ADR-0020 + ADR-0021 + 기타 docs/planning 문서의 historical "26.0" 언급은 **immutable 보존** (결정 시점의 snapshot).

### 3.4 외부 ingress 포트 13000 정합

§2.3 의 retreat 식별 시점에서 짧게 언급된 **외부 ingress port 13000** 은 DevHub 사내 운영 환경의 단일 포트 reverse proxy 진입점이다. 본 sub-section 은 13000 의 코드 매핑을 명시하여 magic number 화를 방지한다.

| 위치 | 용도 |
| --- | --- |
| `infra/idp/keycloak-realm.dev.json` | `devhub-frontend` client 에 `http://localhost:13000` 6 위치 등록 — 사내 ingress 시뮬레이션 (세부 분해: `redirectUris` 3 entries L64-66 / `webOrigins` 1 entry L73 / `post.logout.redirect.uris` 2 entries L77 `##` 구분 string 안) |
| `scripts/deploy-from-env.sh:14` | `PUBLIC_ACCESS_PORT:=13000` default — 사내 deploy 시 외부 client access endpoint 의 port |
| `scripts/deploy-from-env.sh:34` | `NGINX_HTTP_PORT:=3000` default — VM 내부 nginx ingress port (13000 은 host:VM port mapping 의 host 측) |
| ADR-0022 §2.3 | "외부 100.90.113.29:13000 → 호스트 → VM → docker 단일 포트 ingress" 한 줄 reference |

**13000 의 의미**: DevHub 사내 운영 환경에서 외부 client (사내 PC) 가 `http://100.90.113.29:13000` 으로 진입 → 호스트 머신의 13000 port 가 VM 의 nginx (3000 port) 로 forward → nginx 가 frontend (`/devhub/*`) / backend (`/devhub/api/*`) / Keycloak (`/devhub/auth/keycloak/*`) 으로 reverse proxy. 단일 포트 컨셉 (ADR-0018) 정합.

> **호스트:VM port forward 의 외부 의존성**: `docker-compose.deploy.yml` 의 nginx service 는 `NGINX_HTTP_PORT` 의 host bind 만 처리한다. 호스트의 13000 ↔ VM 내부 3000 의 forward 는 **사내 인프라 측 (SSH tunnel / NAT / 호스트 측 port mapping) 별도 설정** 에 의존한다. compose 자체는 13000 → 3000 자동 forward 를 처리하지 않는다.

**사내 환경 외 (다른 운영 환경 / staging / 외부 reference)**: `PUBLIC_ACCESS_HOST` + `PUBLIC_ACCESS_PORT` env 만 변경하면 다른 host:port 로 재배치 가능. realm.dev.json 의 13000 entry 는 사내 dev/smoke 외 환경에서는 무해 (해당 origin 으로 접속 자체가 안 됨).

## 4. 결과

### 4.1 긍정

- 사내 운영 환경 정합 (§3.1 확정 후 명시).
- 25.0.x patch 자동 흡수 — 보안 패치 흐름 유지.
- Keycloak 26.x 의 silent upgrade risk 차단.
- realm export / KC config / `--proxy-headers` 등 메이저 호환성 동결.

### 4.2 부정

- 26.x 의 신규 기능 (event listener SPI 변경 / Account Console UI 개선 등) 미사용.
- 차후 사내 표준이 26.x 로 진입하면 별도 ADR-0023 (forward pin) 필요.
- 25.x EOL 시점 (Keycloak 의 release cadence 상 ~6개월 후 추정) 전 메이저 bump carve 필수.

### 4.3 ADR governance

- ADR-0019 의 본문은 수정하지 않음 (immutable history 원칙 — `feedback_adr_supersession_pattern`).
- ADR-0019 의 메타 헤더에 본 ADR 의 cross-reference 추가 권장 (선택, 사용자 결정).
- ADR-0019 의 핵심 결정 (단일 IdP) 은 reverse 되지 않으므로 supersession 아님 — **자연 확장** (`partial supersedes` 명시 불필요).

## 5. 후속 작업

| 항목 | 상태 | 비고 |
| --- | --- | --- |
| 본 ADR §3.1 retreat 사유 finalize | pending | 사용자 동반 |
| Keycloak 25.0 image pull + redeploy smoke | pending | 사내 운영자 |
| 25.0 의 realm import 정합 검증 (`infra/idp/keycloak-realm.dev.json` + `prod.json`) | pending | 26.x → 25.x 호환 |
| `--proxy-headers=xforwarded` 동작 검증 | pending | 25.x 에서도 동일 syntax 유효 확인 |
| ADR-0019 메타 헤더 cross-reference 추가 여부 | pending | 사용자 결정 |
| 25.0 patch 정기 bump carve out | open | 보안 패치 follow 절차 |

## 6. 변경 이력

| 일자 | 변경 | sprint |
| --- | --- | --- |
| 2026-05-22 | 1차 draft 발행. §3.1 retreat 사유 placeholder + active code 3 위치 정합 commit (`42b18b1`). | `codex/work_260521-c-db-docker-option` |
| 2026-05-26 | §3.4 외부 ingress 포트 13000 정합 sub-section 신규 — magic number 방지 + code 매핑 표 (realm.dev.json / deploy-from-env.sh / ADR §2.3 reference). PR #296 follow-up P1-신규-1 흡수. | `claude/work_260526-pr296-followup-p1-docs` |
