# Keycloak 운영 SOP (realm / client / role / JWKS rotation / user attribute)

- 문서 목적: [ADR-0019 Keycloak 단일화](../adr/0019-keycloak-only-idp.md) §5.3 carve out (1) Keycloak realm/client/role 운영 SOP + (2) JWKS rotation 운영 SOP + (3) Keycloak ↔ HRDB sync (admin user attribute 매핑) 의 단일 통합 운영 자산.
- 범위: Keycloak realm `devhub` 의 client 2종 + role 4종 + user attribute mapper + JWKS rotation policy + local embedded vs external 모드 분기 + 운영 SOP (생성/검증/회수/장애). MFA / SSO logout chain / failover / off-boarding 즉시성 / groups → RBAC 자동 매핑은 [ADR-0019 §5.3 잔여 carve](../adr/0019-keycloak-only-idp.md#53-잔여-carve-out) 로 별도 sprint.
- 대상 독자: 운영자 (SRE / IdP), Security, Backend / Frontend / IdP 담당자.
- 상태: draft (1차)
- 최종 수정일: 2026-05-19
- 결정 근거 sprint: `claude/work_260519-c`
- 관련 문서: [ADR-0019 Keycloak 단일화](../adr/0019-keycloak-only-idp.md), [ADR-0001 IdP (Hydra+Kratos, superseded)](../adr/0001-idp-selection.md), [keycloak_only_refactor_execution_plan §6](../planning/keycloak_only_refactor_execution_plan.md#6-keycloak-서버-구성-계획), [ADR-0008 HRDB production adapter](../adr/0008-hrdb-production-adapter.md), [test-server-deployment](./test-server-deployment.md), [environment-setup](./environment-setup.md).

## 1. 배경 + 책임 분리

DevHub 는 [ADR-0019](../adr/0019-keycloak-only-idp.md) 채택으로 Keycloak 을 **단일 IdP** 로 운영한다. 본 SOP 는 Keycloak 인스턴스 자체의 운영 자산 (realm 정의 / client / role / mapper / JWKS key) 의 결정 + 변경 SOP 를 source-of-truth 로 정합한다.

- **DevHub 저장소 (`devhub`)**: 의도 + 환경설정 계약 (env 변수) + 본 SOP 의 운영 정책 source-of-truth. realm/client 의 실 instance config (admin password, client_secret) 는 외부.
- **운영 자산 저장소 (별도)**: realm export JSON / client_secret / TLS 인증서 / agent 운영 스크립트가 deploy 되는 곳. **devhub git 외부**, 사내 vault 또는 ops git.

CLAUDE.md 의 docker = env-specific 정책 정합 — 환경 특화 자산 (Keycloak host 주소, realm 비밀번호, client_secret, JWKS private key) 은 devhub 저장소 외부.

## 2. realm 구조

| 항목 | 값 | 메모 |
| --- | --- | --- |
| realm name | `devhub` | local embedded + external 모드 동일 |
| display name | `DevHub` | 로그인 화면 표시 (Keycloak admin > Realm settings > General) |
| frontend URL | (env 별) — local `http://localhost:8080`, prod `https://devhub.example.com/devhub/auth/keycloak` ([ADR-0018](../adr/0018-single-port-reverse-proxy-policy.md) 단일 포트 정합) | issuer = `${frontend_url}/realms/devhub` |
| login theme | `keycloak` (default) | 사내 브랜딩은 별도 carve |
| session timeout | SSO Session Idle = `30 min`, SSO Session Max = `12 h` | 사내 정책 기반 (carve — 사내 보안팀 확정) |
| password policy | `length(8) and notUsername` 최소 | 사내 정책 (carve — Keycloak 표준 password policy 강도 확정) |

### 2.1 issuer URL 일관성

- backend `DEVHUB_OIDC_ISSUER_URL` + frontend `NEXT_PUBLIC_OIDC_ISSUER_URL` 가 동일한 issuer 가리킴
- ADR-0018 단일 포트 환경에서는 `https://devhub.example.com/devhub/auth/keycloak/realms/devhub` 형태
- nginx 가 `/devhub/auth/keycloak/*` prefix 를 Keycloak `:8080` 으로 strip-and-proxy

## 3. client 정의 (2종)

### 3.1 `devhub-frontend` (public, PKCE required)

| 항목 | 값 |
| --- | --- |
| Client ID | `devhub-frontend` |
| Client type | `OpenID Connect` |
| Access type | `public` (no client secret) |
| Standard flow | `enabled` (Authorization Code) |
| Direct access grants | `disabled` (보안 — Resource Owner Password Credentials 차단) |
| Service accounts | `disabled` |
| **PKCE** | `S256` **required** |
| Valid redirect URIs | local `http://localhost:3000/auth/callback`, prod `https://devhub.example.com/devhub/auth/callback` |
| Valid post-logout redirect URIs | local `http://localhost:3000/`, prod `https://devhub.example.com/devhub/` |
| Web Origins | local `http://localhost:3000`, prod `https://devhub.example.com` |
| Front-channel logout | `enabled` |

### 3.2 `devhub-backend` (confidential, service account)

| 항목 | 값 |
| --- | --- |
| Client ID | `devhub-backend` |
| Client type | `OpenID Connect` |
| Access type | `confidential` (client secret 발급) |
| Standard flow | `disabled` |
| Direct access grants | `disabled` |
| **Service accounts** | `enabled` (`client_credentials` grant) |
| Service account roles | realm-management 의 `view-users`, `manage-users` (admin operations 용 — Keycloak Admin Client) |
| Valid redirect URIs | (불필요 — service account only) |

client_secret 은 사내 vault 보관. 정기 rotation SOP 는 §6 JWKS rotation 과 별도 (vault 의 secret rotation 정책 따름).

## 4. role / group 정의

### 4.1 realm role 4종

[ADR-0011 RBAC row-scoping](../adr/0011-rbac-row-scoping.md) + DevHub `rbac_policies` seed (migration `000004_seed_rbac.up.sql`) 와 1:1 매핑:

| Role | Description | Keycloak naming |
| --- | --- | --- |
| `developer` | 일반 개발자 (default) | `devhub-developer` 또는 `developer` (DevHub backend role wire 와 일치) |
| `manager` | 부서 manager | `devhub-manager` 또는 `manager` |
| `pmo_manager` | PMO manager (Application/Project Owner 위양) | `devhub-pmo-manager` 또는 `pmo_manager` |
| `system_admin` | 시스템 관리자 (모든 권한) | `devhub-system-admin` 또는 `system_admin` |

**naming 결정**: DevHub backend 의 [`internal/auth/keycloak_verifier.go`](../../backend-core/internal/auth/keycloak_verifier.go) 가 token claim 의 role 을 그대로 wire format 으로 사용하므로, Keycloak role name 도 **`developer` / `manager` / `pmo_manager` / `system_admin`** 으로 prefix 없이 둔다. 사내 정책으로 prefix 필요 시 mapper 에서 strip 처리.

### 4.2 role mapping 위치

| 위치 | mapping |
| --- | --- |
| Realm Role | 위 4종 |
| Client Role (`devhub-frontend`) | (사용 안 함 — realm role 만 활용) |
| Composite Role | (carve — 사내 권한 위계 결정 후) |

### 4.3 group (carve)

- group 기능은 본 SOP scope 외 — `groups` claim → DevHub RBAC role 자동 매핑은 [ADR-0019 §5.3 잔여 carve](../adr/0019-keycloak-only-idp.md#53-잔여-carve-out) (별도 ADR 후보).

## 5. user attribute mapper (token claim 매핑)

### 5.1 표준 claim (Keycloak 기본 제공)

| claim | source | DevHub 사용처 |
| --- | --- | --- |
| `sub` | Keycloak user UUID | `users.idp_subject` 1:1 매핑 (migration `000021_rename_kratos_identity_to_idp_subject`) |
| `preferred_username` | Keycloak username | `audit_logs.actor_login` ([ADR-0019 §4.5](../adr/0019-keycloak-only-idp.md#45-audit_logs-영향)) |
| `email` | Keycloak email attribute | `users.email` 동기화 |
| `email_verified` | Keycloak | (사내 정책 — 미검증 email 거부 carve) |
| `name` | Keycloak first+last name | `users.display_name` |
| `realm_access.roles` | Keycloak realm role | backend [`keycloak_verifier.go`](../../backend-core/internal/auth/keycloak_verifier.go) 의 role extraction |
| `resource_access.{client_id}.roles` | Keycloak client role | fallback (PR #167 KC-PR-B `resource_access` fallback 로직) |

### 5.2 custom claim — `employee_id` (HRDB sync 핵심)

[ADR-0008 HRDB production adapter](../adr/0008-hrdb-production-adapter.md) + [ADR-0019 §5.3 carve](../adr/0019-keycloak-only-idp.md#53-잔여-carve-out) 의 Keycloak ↔ HRDB sync (employee_id strict link).

**Keycloak admin console 설정**:

1. Realm `devhub` → Clients → `devhub-frontend` → Client Scopes 탭
2. 또는 Client Scopes → `profile` (default scope) → Mappers → Add Mapper → "By configuration"
3. Mapper type = **User Attribute**
4. 설정:
   - Name: `employee_id`
   - User Attribute: `employee_id`
   - Token Claim Name: `employee_id`
   - Claim JSON Type: `String`
   - Add to ID token: `ON`
   - Add to access token: `ON`
   - Add to userinfo: `ON`
5. 저장 후 jwt.io 등으로 access_token decode 해 `employee_id` claim 확인

**user attribute 입력 경로**:

- 운영팀 매뉴얼: Keycloak admin console → Users → 해당 user → Attributes → `employee_id` = HRDB primary key
- 자동화: SCIM bridge 또는 사내 HRDB ETL → Keycloak Admin API (carve — sprint 별도)

### 5.3 token claim 검증 체크리스트 (backend 측)

[`keycloak_verifier.go`](../../backend-core/internal/auth/keycloak_verifier.go) 의 verifier 가 검증하는 claim:

- `iss` ↔ `DEVHUB_OIDC_ISSUER_URL`
- `aud` ↔ `DEVHUB_OIDC_AUDIENCE` (또는 client_id)
- `exp` / `nbf` / `iat` ↔ clock skew 허용 (default 60s, env override 가능)
- `sub` ↔ `users.idp_subject` lookup
- `realm_access.roles` → DevHub role wire (fallback: `resource_access.{client_id}.roles`)
- `preferred_username` → audit_logs actor_login

## 6. JWKS rotation 운영 SOP

### 6.1 Keycloak 의 key rotation 기본 동작

- Keycloak realm 의 keys 는 **여러 active key 동시 보유** 가능 (active + passive)
- token 발급 시 active key 로 서명
- token 검증 시 모든 active + passive key 의 JWKS 가 노출 → 이전 active key 로 발급된 token 도 유효 기간 동안 검증 가능
- key rotation = active key 를 새로 생성 + 이전 key 를 passive 로 이동 (token TTL 까지 검증 유효)

### 6.2 rotation 주기 권장

| 환경 | active key rotation | passive key 보관 |
| --- | --- | --- |
| local embedded | 갱신 불필요 (개발) | — |
| staging | 90일 | 14일 (token TTL 보다 길게) |
| prod | 90일 (사내 보안 정책에 따름) | 14일 |

권장 = 90일 active rotation, 14일 passive retain. 단 사내 보안팀 정책이 더 짧으면 따름.

### 6.3 DevHub backend JWKS cache invalidation

[`keycloak_verifier.go`](../../backend-core/internal/auth/keycloak_verifier.go) 의 JWKS cache:

- TTL default = 10분 (PR #167 KC-PR-B 의 cache 구현)
- cache miss 시 `${DEVHUB_OIDC_JWKS_URL}` 또는 issuer discovery 의 `/protocol/openid-connect/certs` fetch
- `kid` (key ID) mismatch → 강제 refetch + cache 갱신

**rotation 시 backend 동작**:
1. Keycloak 에서 new active key 생성 (이전 key passive 이동)
2. backend 가 새 token 의 `kid` 받으면 cache miss → JWKS refetch → 새 key 흡수
3. cache TTL 만료 후에도 새 key 자동 흡수 보장 — **수동 backend 재시작 불필요**

### 6.4 rotation 운영 SOP (D-Day)

1. **사전 준비** (D-7)
   - 사내 보안팀에 rotation 일정 공지
   - backend cache TTL (10분) + token TTL (access 5분 권장) 확인
2. **rotation 당일 (D-Day)**
   - Keycloak admin console → Realm settings → Keys → Providers
   - 신규 key provider 추가 (rsa-generated, priority 100) — 자동 active 전환
   - 이전 key provider priority 감소 (passive)
3. **검증** (D+0, 30분 후)
   - backend `/health` + 1 test login → 새 token 의 `kid` 확인 (jwt.io decode)
   - backend log 의 JWKS refetch 발생 확인
4. **passive cleanup** (D+14, token TTL 보다 길게)
   - 이전 key provider disable (passive 종료)
   - 잔여 token 검증 실패 시 강제 재로그인 (모니터링)

### 6.5 비상 rotation (key 유출 의심 시)

1. **즉시 회수**: Keycloak admin → 의심 key provider disable
2. **모든 token 무효화**: Realm settings → Sessions → "Revoke all sessions" — 모든 활성 token 무효화 + 강제 재로그인
3. **신규 key**: 새 rsa-generated 추가 (priority 100)
4. **사후 audit**: Keycloak event listener log 의 `LOGIN_ERROR` + `INVALID_TOKEN` 패턴 추적

## 7. local embedded vs external 모드 분기

[`keycloak_only_refactor_execution_plan §6`](../planning/keycloak_only_refactor_execution_plan.md#6-keycloak-서버-구성-계획) 가 정의한 2 모드.

### 7.1 local embedded 모드

- 용도: 개발 환경 (개발자 로컬)
- 기동: 환경 특화 (docker-compose 또는 native binary — devhub git 외부)
- realm: `devhub` (위 §2 정의)
- clients: `devhub-frontend` + `devhub-backend` (위 §3)
- 사용자 seed: `alice` / `bob` / `charlie` (개발 편의)
- env (`.env.local`):
  ```
  DEVHUB_IDP_PROVIDER=keycloak
  DEVHUB_OIDC_ISSUER_URL=http://localhost:8080/realms/devhub
  DEVHUB_OIDC_CLIENT_ID=devhub-frontend
  # backend confidential
  DEVHUB_KEYCLOAK_ADMIN_URL=http://localhost:8080
  DEVHUB_KEYCLOAK_ADMIN_REALM=devhub
  DEVHUB_KEYCLOAK_ADMIN_CLIENT_ID=devhub-backend
  DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET=<vault>
  ```

### 7.2 external 모드 (운영)

- 용도: staging / prod
- 기동: 사내 Keycloak (별도 운영팀 관리)
- 앱은 외부 issuer/discovery URL 신뢰 — local Keycloak 기동 안 함
- env: 사내 vault 주입 — realm/client/secret/issuer 모두 외부
- 운영 체크리스트:
  - `iss` / `aud` mismatch 검증 (backend log + 1 test login)
  - clock skew 허용 범위 (default 60s, 사내 NTP sync 상태 확인)
  - JWKS rotation/cache 설정 (§6)
  - TLS/CA 신뢰체인 (사내 SSL inspection 환경 시 root CA 주입 — `SSL_CERT_FILE` 또는 `/etc/ssl/certs` 에 사내 CA 추가)
  - [ADR-0018 단일 포트 reverse proxy](../adr/0018-single-port-reverse-proxy-policy.md) 환경에서는 nginx 의 `/devhub/auth/keycloak/*` strip-and-proxy 정합 검증

## 8. 운영 SOP (생성 / 검증 / 회수 / 장애 대응)

### 8.1 신규 user 생성

| 단계 | 위치 | 액션 |
| --- | --- | --- |
| 1. Keycloak admin | Users → Add user | username (= DevHub `users.user_id` 와 일치 권장), email, first/last name |
| 2. user attribute | Attributes 탭 | `employee_id` = HRDB primary key |
| 3. role | Role Mapping 탭 | realm role 1개 선택 (developer / manager / pmo_manager / system_admin) |
| 4. 초기 비밀번호 | Credentials 탭 | password 설정 + "Temporary" ON (첫 로그인 시 강제 변경) |
| 5. DevHub `users` sync | (자동) | 첫 로그인 시 backend 가 `sub` ↔ `idp_subject` 매핑 + HRDB lookup 으로 `users` row 보강 |

### 8.2 user 회수 (off-boarding)

| 단계 | 위치 | 액션 |
| --- | --- | --- |
| 1. user 비활성화 | Users → 해당 user → Details → Enabled = OFF | 이후 로그인 차단 |
| 2. 활성 session 무효화 | Sessions 탭 → "Logout all sessions" | 강제 logout (refresh token 까지) |
| 3. DevHub `users.status` sync | (자동/수동 carve) | HRDB ETL cron 이 비활성 동기화 — `users.status = deactivated` |

### 8.3 client_secret rotation

| 단계 | 위치 | 액션 |
| --- | --- | --- |
| 1. 신규 secret | Clients → `devhub-backend` → Credentials → Regenerate | 신규 secret 발급 |
| 2. vault 갱신 | 사내 vault | `DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET` 신규 값 |
| 3. backend 재기동 | (사내 ops) | env 재주입 + service restart |
| 4. 검증 | backend `/health` + admin API 호출 1회 | 200 OK 확인 |

### 8.4 장애 대응

| 시나리오 | 즉시 대응 | 후속 |
| --- | --- | --- |
| Keycloak 자체 down | 사내 운영팀 alert + DevHub 사용자 공지 (로그인 전체 불가) | failover 정책 별도 carve ([ADR-0019 §5.3](../adr/0019-keycloak-only-idp.md#53-잔여-carve-out)) |
| JWKS endpoint 응답 timeout | backend cache TTL 동안 유효 (10분) — 그 사이 Keycloak 복구 | cache TTL 초과 시 모든 인증 실패 — Keycloak 복구 우선 |
| `kid` mismatch (서명 검증 실패) | backend log 의 JWKS refetch 시도 확인 | 강제 cache invalidate (backend 재시작) 또는 Keycloak key rotation 검증 |
| token 유출 의심 | §6.5 비상 rotation 절차 진행 | audit_logs + Keycloak event log 분석 |

## 9. 보안 점검

### 9.1 잠재 위협 + mitigation

| 위협 | mitigation |
| --- | --- |
| client_secret leak | vault 보관 + rotation SOP (§8.3) + Keycloak admin event log 의 `CLIENT_*` 액션 audit |
| JWKS private key leak | §6.5 비상 rotation 절차 |
| token 재사용 (replay) | Keycloak token TTL 짧게 (access 5분 권장), refresh token rotation 활성화 (Keycloak Tokens 탭 → Refresh Token Max Reuse = 0) |
| user attribute 위조 (employee_id) | Keycloak admin event log + 사내 운영팀의 attribute 변경 권한 제한 (system_admin 한정) |
| brute force login | Keycloak Realm Settings → Security Defenses → Brute Force Detection ON (default) |

### 9.2 audit log 통합 (carve)

- Keycloak 의 event listener / admin event SPI 를 DevHub `audit_logs` 와 통합하는 SOP 는 별도 carve — [ADR-0019 §5.3 잔여 carve](../adr/0019-keycloak-only-idp.md#53-잔여-carve-out) 의 audit 항목.

## 10. 잔여 carve out

본 SOP 의 scope 외 — 별도 sprint:

| 항목 | 관련 ADR / 위치 | 비고 |
| --- | --- | --- |
| SSO logout chain (RP-initiated logout) | ADR-0019 §5.3 | Keycloak `end_session_endpoint` + DevHub frontend redirect chain |
| MFA 도입 | ADR-0019 §5.3 | Keycloak MFA / WebAuthn 표준 정책 활성화 |
| Keycloak failover (HA 구성 또는 backup IdP) | ADR-0019 §5.3 | 단일 장애점 회피 |
| off-boarding 즉시성 | ADR-0019 §5.3 | HR 시스템 → Keycloak → DevHub propagation chain |
| `groups` claim → DevHub RBAC role 자동 매핑 | ADR-0019 §5.3 (별도 ADR 후보) | composite role 또는 mapper 로 |
| Keycloak event SPI → DevHub `audit_logs` 통합 | ADR-0019 §5.3 + ADR-0019 §4.5 | event listener |
| 사내 LDAP/AD federation | ADR-0019 §5.4 RM-M4-09 | Keycloak User Federation |
| Gitea SSO via Keycloak identity broker | ADR-0019 §5.4 RM-M4-09 | M4 RM-M4 진입 시 |

## 11. 변경 이력

| 일자 | 변경 | sprint |
| --- | --- | --- |
| 2026-05-19 | 1차 draft — §2 realm + §3 client 2종 + §4 role 4종 + §5 user attribute mapper (preferred_username / email / realm_access.roles / employee_id custom) + §6 JWKS rotation 운영 SOP + §7 local embedded vs external 분기 + §8 운영 SOP (생성/회수/secret rotation/장애) + §9 보안 점검 + §10 잔여 carve out 8 항목. [ADR-0019 §5.3 carve out (1)+(2)+(3) resolved](../adr/0019-keycloak-only-idp.md#53-잔여-carve-out) 의 source-of-truth. | `claude/work_260519-c` |
