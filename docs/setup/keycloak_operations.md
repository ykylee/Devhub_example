# Keycloak 운영 SOP (realm / client / role / JWKS rotation / user attribute)

- 문서 목적: [ADR-0019 Keycloak 단일화](../adr/0019-keycloak-only-idp.md) §5.3 carve out (1) Keycloak realm/client/role 운영 SOP + (2) JWKS rotation 운영 SOP + (3) Keycloak ↔ HRDB sync (admin user attribute 매핑) 의 단일 통합 운영 자산.
- 범위: Keycloak realm `devhub` 의 client 2종 + role 4종 + user attribute mapper + JWKS rotation policy + local embedded vs external 모드 분기 + 운영 SOP (생성/검증/회수/장애). MFA / SSO logout chain / failover / off-boarding 즉시성 / groups → RBAC 자동 매핑은 [ADR-0019 §5.3 잔여 carve](../adr/0019-keycloak-only-idp.md#53-잔여-carve-out) 로 별도 sprint.
- 대상 독자: 운영자 (SRE / IdP), Security, Backend / Frontend / IdP 담당자.
- 상태: draft
- 최종 수정일: 2026-05-20
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
| Valid post-logout redirect URIs | local `http://localhost:3000/`, prod `https://devhub.example.com/devhub/` (sprint -s PR #187 — basePath 포함 정합). 이전 sprint -j codex review #9 #4 의 basePath 미포함 표기는 sprint -s 의 backend 확장 carve #3 으로 정합 — frontend `auth.service.ts:116` 의 `post_logout_redirect_uri` 가 `${origin}${BASE_PATH}/` 로 변경됨. |
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

### 4.3 group (composite realm role 1:1 매핑)

[docs/planning/keycloak_groups_rbac_mapping.md](../planning/keycloak_groups_rbac_mapping.md) 결정 — 옵션 B 채택. group 4개 ↔ realm role 4개 1:1 composite 매핑.

| Keycloak Group | Composite Realm Role | DevHub `users.role` |
| --- | --- | --- |
| `devhub-developers` | `developer` | `developer` |
| `devhub-managers` | `manager` | `manager` |
| `devhub-pmo-managers` | `pmo_manager` | `pmo_manager` |
| `devhub-system-admins` | `system_admin` | `system_admin` |

**Keycloak admin 설정 SOP** (1회):
1. Realm `devhub` → Groups → Create Group 4회 (위 표 group name)
2. 각 Group → Role Mappings 탭 → realm role 1개 assign (group ↔ role 매핑)
3. **Default Groups 미설정 권장** (codex review #9, [keycloak_groups_rbac_mapping §3.2](../planning/keycloak_groups_rbac_mapping.md) 결정) — Default Group 적용 시 신규 manager/pmo/admin 도 자동 `devhub-developers` 가입 → token multi-role → `extractKeycloakRole` order-dependency 위험. 명시 group 1개 가입 강제 (§8.1 step 3).

**backend 동작**: 변경 없음. group composite role 은 Keycloak 이 token 발급 시 `realm_access.roles` 에 자동 포함 → [keycloak_verifier.go:260-285](../../backend-core/internal/auth/keycloak_verifier.go) 의 추출 그대로 동작.

**user 운영**: §8.1 step 3 의 "Role Mapping" → "**Groups 탭 → group 1개 가입**" 으로 단순화 (다중 user 일괄 처리 가능).

**carve**: SCIM bridge / LDAP federation 자동 group sync + 옵션 C (groups claim mapper + multi-role) 확장 — design 문서 §8 잔여 carve.

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

[`keycloak_verifier.go:37`](../../backend-core/internal/auth/keycloak_verifier.go) 의 JWKS cache:

- TTL default = **5분** (`defaultJWKSTTL = 5 * time.Minute`)
- cache miss (TTL 만료 또는 빈 cache) 시 `${DEVHUB_OIDC_JWKS_URL}` 또는 issuer discovery 의 `/protocol/openid-connect/certs` fetch
- **kid mismatch 시 stale-while-error fallback** (sprint -r PR #186, 2026-05-19): `kid` mismatch (codex review #9 #3 carve resolved) 시 backend 가 자동 `invalidateCache` + JWKS forced refetch + 1회 retry. Keycloak key rotation 직후 새 kid 의 token 이 cache TTL 만료 전이라도 정합. signature/expired/issuer/audience 등 non-kid error 는 retry 안 함 (security 위협 회피).

**rotation 시 backend 동작** (sprint -r stale-while-error fallback 적용 후):
1. Keycloak 에서 new active key 생성 (이전 key passive 이동)
2. backend 가 새 kid token 받음 → 1차 parse 시 `errKidMismatch` → `invalidateCache` + JWKS forced refetch → 2차 parse 성공
3. **graceful window 0 — 사용자 영향 없음**
4. signature / expired / issuer / audience error 는 retry 안 함 (security)

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
| 3. role (group 가입) | **Groups 탭** | **group 1개 가입** ([§4.3](#43-group-composite-realm-role-11-매핑) 4종 중 1개) — composite realm role 자동 상속. **Default Group 미설정 권장** (codex review #9 — multi-role order-dependency 위험). |
| 4. 초기 비밀번호 | Credentials 탭 | password 설정 + "Temporary" ON (첫 로그인 시 강제 변경) |
| 5. DevHub `users` sync | **(carve — 자동 sync 미구현)** | 현재 backend [`authenticateActor`](../../backend-core/internal/auth/keycloak_verifier.go) 는 token 검증 + actor context stash 만. `SetIdPSubject` 호출은 `accounts_admin.go:123` (관리자 발급/PATCH path) 에서만. **신규 user 의 `users.idp_subject` + HRDB lookup 자동 sync 는 별도 SOP 또는 backend 확장 필요** — codex review #9 carve. 임시 SOP: admin 이 신규 user 발급 시 별도 `/api/v1/accounts` 호출 또는 직접 DB UPSERT. |

### 8.2 user 회수 (off-boarding)

자세한 즉시성 design 은 [docs/planning/keycloak_offboarding_immediacy.md](../planning/keycloak_offboarding_immediacy.md) — HR ETL ↔ Keycloak ↔ DevHub propagation chain. 본 SOP 는 Phase 1 (옵션 C HR ETL push) 운영 절차.

#### 수동 즉시 회수 (긴급, Keycloak admin 직접)

| 단계 | 위치 | 액션 |
| --- | --- | --- |
| 1. user 비활성화 | Users → 해당 user → Details → Enabled = OFF | 이후 로그인 차단 |
| 2. 활성 session 무효화 | Sessions 탭 → "Logout all sessions" | 강제 logout (refresh token 까지). access_token 은 TTL (권장 5분, [§6.2](#62-rotation-주기-권장)) 동안 유효. |
| 3. DevHub `users.status` sync | (자동) | 다음 hourly HRDB ETL cron 에서 `users.status = deactivated` |
| 4. audit 검수 | Keycloak admin event log | `USER:UPDATE` (enabled=false) + `USER:ACTION` (LOGOUT) 발생 확인 ([keycloak_event_audit_integration.md §4.2](../planning/keycloak_event_audit_integration.md#42-admin-event) 매핑 정합) |

#### 자동 회수 (운영 cron, Phase 1)

| 단계 | 위치 | 액션 |
| --- | --- | --- |
| 1. HR 시스템 | 사내 HR system | 퇴사 / 비활성화 처리 |
| 2. HR ETL cron 실행 | 사내 운영 cron (**hourly**, [Phase 1 design §3.1](../planning/keycloak_offboarding_immediacy.md)) | `scripts/hrdb_etl_sync.sh` 실행 — (a) DevHub `hrdb` schema UPSERT + (b) Keycloak Admin API user disable + (c) force logout |
| 3. access_token TTL 만료 | (자동) | 권장 5분 TTL 후 새 token 발급 차단 → 실 권한 회수 |
| 4. **worst case latency** | — | ETL 주기 (1h) + token TTL (5분) ≈ **1시간** |

#### Phase 2 carve — LDAP/AD federation 도입 시

사내 LDAP/AD 운영 중이면 Keycloak User Federation 으로 worst case ≤ 15분 ([offboarding §4](../planning/keycloak_offboarding_immediacy.md) Phase 2).

### 8.3 client_secret rotation

| 단계 | 위치 | 액션 |
| --- | --- | --- |
| 1. 신규 secret | Clients → `devhub-backend` → Credentials → Regenerate | 신규 secret 발급 |
| 2. vault 갱신 | 사내 vault | `DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET` 신규 값 |
| 3. backend 재기동 | (사내 ops) | env 재주입 + service restart |
| 4. 검증 | backend `/health` + admin API 호출 1회 | 200 OK 확인 |

### 8.4 장애 대응

자세한 failover design 은 [docs/planning/keycloak_failover.md](../planning/keycloak_failover.md). 본 SOP 는 Phase 1 (graceful degradation) 운영 절차.

#### 기본 시나리오

| 시나리오 | 즉시 대응 | 후속 |
| --- | --- | --- |
| Keycloak 자체 down (5분 미만) | best case 사용자 무영향 (JWKS cache + access_token 모두 fresh 시점 시작). **worst case 401 가능** (codex review #9 — cache 만료 + token 만료 시점 겹침) — 5분 미만이라도 alert 트리거 + 모니터링 | Keycloak 복구 후 자동 정상화 |
| Keycloak 자체 down (5-10분) | 사용자 공지 (status page) + Keycloak 복구 진행 알림 | 점진 logout 발생 — 복구 후 재로그인 안내 |
| Keycloak 자체 down (10분 이상) | Page on-call SRE + 사용자 영향 분석 | 사후 review + failover Phase 2 HA 도입 평가 ([keycloak_failover §4](../planning/keycloak_failover.md)) |
| JWKS endpoint 응답 timeout | backend cache TTL 동안 유효 (5분, [keycloak_verifier.go:37](../../backend-core/internal/auth/keycloak_verifier.go) `defaultJWKSTTL`) | cache TTL 초과 시 모든 인증 실패 — Keycloak 복구 우선 |
| `kid` mismatch (서명 검증 실패) | backend log 의 JWKS refetch 시도 확인 | 강제 cache invalidate (backend 재시작) 또는 Keycloak key rotation 검증 |
| token 유출 의심 | §6.5 비상 rotation 절차 진행 | audit_logs + Keycloak event log 분석 |

#### Phase 2 carve — Keycloak HA 도입 시

사내 인프라 결정 시 [keycloak_failover §4](../planning/keycloak_failover.md) 의 옵션 B (active-active cluster) 또는 옵션 C (active-passive) 로 RTO ≈ 0초 / 분 단위 단축. DevHub 측 변경 없음.

#### 권장 모니터링 metric (carve, [keycloak_event_audit_integration §5.3](../planning/keycloak_event_audit_integration.md) PR-C 와 정합)

- `devhub_jwks_fetch_total{status}` — JWKS fetch 성공/실패 counter
- `devhub_jwks_cache_age_seconds` — cache 잔여 시간 gauge
- alert: `devhub_jwks_fetch_total{status="failed"} > 0 for 2 minutes` → Keycloak 가용성 의심

### 8.5 SSO logout chain (RP-initiated logout)

[ADR-0019 §5.3](../adr/0019-keycloak-only-idp.md#53-잔여-carve-out) 의 SSO logout chain carve out. DevHub 가 OIDC RP (Relying Party) 로서 Keycloak 의 SSO 세션을 종료하는 [RP-initiated logout](https://openid.net/specs/openid-connect-rpinitiated-1_0.html) 표준 흐름 운영 SOP.

#### 8.5.1 현재 구현 패턴 (frontend)

`frontend/lib/services/auth.service.ts:100-126` 의 `logout()` 함수가 RP-initiated logout 표준 구현:

```ts
public async logout(): Promise<void> {
  const idToken = tokenStore.getIdToken();                                  // 1. id_token 사전 회수
  const runtimeConfig = await this.getRuntimeOIDCConfig();
  const discovery = await this.getDiscovery();

  tokenStore.clear();                                                       // 2. local 토큰 즉시 clear
  useStore.getState().setIsLoggingOut(true);                                //    (LogoutOverlay UX 활성화)
  useStore.getState().clearActor();                                         //    (actor 상태 초기화)

  const endSessionEndpoint =
    discovery.end_session_endpoint ||                                       // 3. OIDC discovery 우선
    `${runtimeConfig.oidcIssuerURL}/protocol/openid-connect/logout`;        //    fallback (Keycloak 표준 path)

  const url = new URL(endSessionEndpoint);
  url.searchParams.set("client_id", OIDC_CLIENT_ID);                        // 4. URL params 구성
  url.searchParams.set("post_logout_redirect_uri", `${window.location.origin}/`);
  if (idToken) {
    url.searchParams.set("id_token_hint", idToken);                         //    id_token_hint 권장 (silent logout)
  }
  window.location.assign(url.toString());                                   // 5. Keycloak 으로 redirect
}
```

**핵심 패턴**:

| Step | 동작 | 효과 |
| --- | --- | --- |
| 1 | `tokenStore.getIdToken()` 사전 회수 | clear 전에 `id_token_hint` 용 값 확보 |
| 2 | `tokenStore.clear()` 즉시 | local 토큰 무효화 — Keycloak redirect 실패 시에도 frontend 안전 |
| 3 | OIDC discovery → `end_session_endpoint` | Keycloak 표준 endpoint (default: `${issuer}/protocol/openid-connect/logout`) |
| 4 | `id_token_hint` + `client_id` + `post_logout_redirect_uri` 명시 | Keycloak 이 SSO 세션 종료 + post-logout redirect URI whitelist 검증 |
| 5 | `window.location.assign(url)` | Keycloak 으로 full redirect (SPA navigation 아님) |

#### 8.5.2 Keycloak admin console 설정 SOP

Keycloak 이 RP-initiated logout 을 받기 위한 client 설정:

1. Realm `devhub` → Clients → `devhub-frontend` → Settings
2. **Valid post-logout redirect URIs** 필드 (§3.1 client 정의 참조):
   - local: `http://localhost:3000/`
   - prod: `https://devhub.example.com/devhub/` (sprint -s PR #187 정합 — frontend `auth.service.ts:116` 가 `${origin}${BASE_PATH}/` 로 변경됨, ADR-0018 basePath /devhub 환경 정합).
   - **wildcard 금지** — 정확한 URI 만 등록. `*` 사용 시 open redirect 취약점.
3. **Front-channel logout** (선택): `enabled` — Keycloak 이 admin force logout 시 모든 active client 의 front-channel logout URL 호출. DevHub 가 server-side session 미사용 (JWT only) 이므로 front-channel logout URL 등록 불필요 (carve — future SOP).
4. **Backchannel logout URL** (선택): 미사용 — DevHub 는 server-side session 없음.

#### 8.5.3 logout chain order

```mermaid
sequenceDiagram
    participant U as User
    participant F as Frontend (Next.js)
    participant KC as Keycloak
    participant B as Backend-Core (Go)

    U->>F: /auth/logout 진입
    F->>F: tokenStore.clear() + setIsLoggingOut(true) + clearActor()
    F->>KC: GET /realms/devhub/protocol/openid-connect/logout?<br/>id_token_hint=...&client_id=devhub-frontend&post_logout_redirect_uri=...
    KC->>KC: id_token_hint 검증 + post_logout_redirect_uri whitelist 검증
    KC->>KC: SSO 세션 종료 (Keycloak 사용자 세션 정리)
    KC->>U: 302 redirect to post_logout_redirect_uri
    U->>F: ${origin}/ (홈)
    F->>F: AuthGuard → 미인증 → /auth/login redirect
```

**chain 순서가 보장하는 invariant**:
- local 토큰 (access/refresh/id) clear 가 **Keycloak redirect 전에 발생** → redirect 실패 시에도 frontend 가 미인증 상태로 안전
- Keycloak SSO 세션 종료가 `id_token_hint` 로 사용자 확인 → CSRF 방지 (id_token 보유자만 logout 가능)
- post_logout_redirect_uri 가 Keycloak 의 whitelist 검증 통과 → open redirect 방지

#### 8.5.4 admin force logout (front-channel)

Keycloak admin 이 사용자 강제 logout 시 (§8.2 off-boarding 또는 §6.5 비상 rotation):

1. Keycloak admin console → Users → 해당 user → Sessions 탭 → "Logout all sessions"
2. **DevHub frontend 동작**: 다음 API 호출 시 access_token 만료 또는 invalid → backend 401 → frontend AuthGuard 가 `/auth/login` redirect
   - access_token 의 `exp` 가 짧으면 (권장 5분) 강제 logout 효과가 5분 내 모든 client 에 전파
   - refresh_token 도 무효화 → 재발급 불가
3. **DevHub 측 추가 처리 불필요** — server-side session 없음 + JWT 만료 기반 자연 종료

#### 8.5.5 보안 점검

| 위협 | mitigation |
| --- | --- |
| open redirect (`post_logout_redirect_uri` 위조) | Keycloak client 의 Valid post-logout redirect URIs whitelist (§8.5.2). wildcard 금지. |
| CSRF (타 사용자 logout 강제) | `id_token_hint` 필수 — Keycloak 이 id_token 의 `aud` + `sub` 검증 후 SSO 세션 종료 |
| token replay (logout 후 재사용) | `tokenStore.clear()` 즉시 (§8.5.1 Step 2) + access_token 짧은 TTL (권장 5분) + refresh_token rotation 활성화 (Keycloak Tokens 탭 → Refresh Token Max Reuse = 0) |
| LogoutOverlay UX 우회 | `useStore.getState().setIsLoggingOut(true)` 가 navigation 중에도 UI 잠금 — sprint `gemini/dreq_e2e_260515` (PR #134) 로 도입 |

#### 8.5.6 잔여 carve out (§8.5 의 sub-carve)

- **(carve)** Front-channel logout URL 등록 — Keycloak 의 admin force logout 시 DevHub 가 즉시 알람 받는 endpoint. 현재는 access_token TTL 기반 자연 종료. SLA 가 더 짧은 즉시 종료 요구 시 carve.
- **(carve)** Backchannel logout (`logout_token` 수신) — Keycloak → DevHub backend POST. server-side session 도입 시 의미 있음. 현재 미적용.
- **(carve)** Multi-tab logout 동기화 — `tokenStore` 가 sessionStorage 인 경우 tab 간 sync 안 됨. localStorage + `storage` event listener 또는 BroadcastChannel API 로 carve.

### 8.5b Self-service 비밀번호 변경 (Keycloak Account Console 위임)

sprint `claude/work_260519-ad` (Kratos 잔재 cleanup) 이후, DevHub 는 self-service 비밀번호 변경 흐름을 **자체 proxy 하지 않는다**. 사용자는 Keycloak Account Console 에서 직접 비밀번호를 변경한다.

#### 8.5b.1 URL 구성

| 항목 | 값 |
| --- | --- |
| Account Console URL | `${OIDC_ISSUER_URL}/account/` (예: `https://idp.example.com/realms/devhub/account/`) |
| DevHub 진입점 | `/account` 페이지의 **"Open Keycloak Console"** 외부 link (`target="_blank"`). frontend `app/(dashboard)/account/page.tsx` 가 `OIDC_ISSUER_URL` 을 endpoints.ts 에서 조립한다. |
| 변경 사항 | (a) `POST /api/v1/account/password` endpoint + Kratos login/settings proxy 제거, (b) `accountService.updateMyPassword` + `SettingsFlowError` 제거, (c) `/account` 페이지의 password form 제거 + Account Console redirect button 으로 대체. |

#### 8.5b.2 사용자 흐름

1. 사용자가 DevHub 의 `/account` 페이지에서 **"Open Keycloak Console"** 버튼 클릭 (새 탭)
2. Keycloak Account Console (`{issuer}/account/`) 로 이동 — 이미 OIDC 로 sign-in 한 사용자는 자동 인증 (SSO)
3. **Signing In** 또는 **Password** 메뉴에서 비밀번호 변경 (Keycloak 의 password policy + privileged session 시간창 정합)
4. 변경 완료 후 DevHub 탭으로 돌아옴 — DevHub access_token 은 그대로 유효 (TTL 만료까지). 다음 로그인부터 새 비밀번호 적용.

#### 8.5b.3 운영 요점

- **DevHub 측 코드 변경 없음** — Keycloak Account Console 의 password policy / MFA enrollment / 활성 세션 관리 등 기능 모두 Keycloak admin 에서 컨트롤.
- **Keycloak admin 측 요구사항** — `devhub-frontend` client 의 Web Origins 가 Account Console URL 도 허용해야 함 (보통 동일 issuer 라 자동 정합).
- **MFA enrollment 도 동일 경로** — Account Console > Signing In > Authenticator. DevHub 는 MFA 시도 결과를 Keycloak 의 token 발급으로 받아 그대로 적용.
- **Audit log** — Keycloak Admin Events log 에 `UPDATE_PASSWORD` event 가 기록되며, sprint -u~-y 의 audit event listener (§8.6) 가 이를 polling 해 DevHub `audit_logs` 로 dedup-emit (source_type=`keycloak_event`, action 매핑은 §8.6 매핑 표 참조).
- **Runtime config 정합 (sprint -ad Stage 3, codex P1 + self-review P1-2 통합 fix)** — `app/(dashboard)/account/page.tsx` 의 Account Console link 는 `authService.getAccountConsoleURL()` 을 통해 `/api/runtime-config` 응답의 `oidc_issuer_url` 로부터 빌드된다. server-side env (`OIDC_ISSUER_URL` 또는 `NEXT_PUBLIC_OIDC_ISSUER_URL`) 둘 다 지원되며, login flow (`auth.service.ts:getRuntimeOIDCConfig`) 와 동일 경로를 공유한다. 즉 빌드 시점에 `NEXT_PUBLIC_*` 가 inline 되지 않은 deployment 에서도 link 가 정상 동작한다.
- **운영 로그 grep 룰 갱신 필요** — backend `identity_resolver.go` 의 IdP subject backfill log prefix 가 `[kratos-cache]` → `[idp-cache]` 로 변경됨 (sprint -ad). 기존 alert/grep 룰이 `[kratos-cache]` 를 watch 한다면 `[idp-cache]` 로 교체. 후속 backend 의 다른 cache 라인 (예: PermissionCache) 은 변경 없음.

#### 8.5b.4 잔여 carve out (§8.5b sub-carve)

- **(carve)** DevHub 의 password change 진입점에 Account Console URL 을 명시적으로 보여주지 않고 SSO redirect 만 제공 — Account Console 진입 시 추가 인증을 요구하는 Keycloak 설정인 경우 사용자에게 ID/PW 재입력이 노출. 환경별 SOP 검토.
- **(carve)** MFA enrollment 강제 (Required Actions) — Keycloak realm 의 Required Actions 에서 강제 enroll 설정 후 DevHub OIDC flow 가 그대로 MFA challenge 받음. 현재 사내 정책으로 MFA 도입 보류 (ADR-0019 §5.3 (5) excluded).

### 8.6 Keycloak event listener (audit_logs 통합) 운영 SOP

[ADR-0019 §5.3 (9) audit event listener](../adr/0019-keycloak-only-idp.md#53-잔여-carve-out) 의 Phase 2 (PR-B~PR-D, sprint -u~-w) 가 머지된 후 운영 환경에서 활성화하는 절차. design 문서: [docs/planning/keycloak_event_audit_integration.md](../planning/keycloak_event_audit_integration.md).

#### 8.6.1 활성화 사전 조건

| 항목 | 확인 |
| --- | --- |
| migration 000031 (`event_cursors` 테이블) 적용 | `psql -c "\d event_cursors"` 가 표 노출 |
| migration 000032 (`audit_logs.source_event_id` + partial UNIQUE INDEX) 적용 | `psql -c "\d audit_logs"` 의 `source_event_id` 컬럼 + `audit_logs_source_event_id_uniq` 인덱스 노출 |
| `devhub-backend` client 의 service account 권한 | realm-management role `view-events` 또는 `realm-admin` (admin event 도 필요 시 `view-events` 만으로 충분) |
| Keycloak admin console 의 Events 활성화 | Realm Settings → Events → User Events 의 Save events `ON` + Admin Events 의 Save events `ON` |
| `event_cursors` row seed | **backend 자동** — `loadCursor` (sprint -y codex hotfix #10 P1-B) 가 row 미존재 시 즉시 `last_event_at = now()` UPSERT. 운영자 사전 INSERT 불필요. 다만 시작 전 backend 가 한 번 이상 tick 을 돌렸는지 `SELECT * FROM event_cursors` 로 확인 권장. |

#### 8.6.2 Keycloak admin console 설정 (Events 활성화)

운영 staging/prod realm 에서 1회 설정. realm 별 분리 — devhub-dev 는 dev 운영팀, devhub-prod 는 운영팀.

1. Realm Settings → **Events** 탭.
2. **User events config**:
   - `Save events` → `ON`
   - `Expiration` → 최소 **운영 outage tolerance + recovery time** 보다 길게 설정 (권장 `7 days` 이상 — sprint -y codex hotfix #10 P2-E 정정). cron 이 cursor 기반 polling 이라 정상 동작 중에는 짧은 expiration 도 가능해 보이지만, **backend 장기 outage / pull 실패가 누적되면 Keycloak 측 expired event 는 backfill 불가능**. 운영자가 7d 미만으로 줄이면 그만큼의 outage 동안 audit 영구 손실. 사내 운영 SLA 검토 후 결정.
   - `Included events` → 비워둠 (전체 emit, 본 SOP 의 `defaultSkipUserEventTypes` 가 REFRESH_TOKEN / CODE_TO_TOKEN / INTROSPECT_TOKEN 을 skip)
3. **Admin events config**:
   - `Save events` → `ON`
   - `Include representation` → `OFF` (privacy 측면 권장, payload 에 user 비밀번호 / secret 등 노출 회피)
   - `Expiration` → User events 와 동일 정책 (운영 outage tolerance + recovery time, 권장 `7 days` 이상)
4. Save.

> Keycloak SPI 측 event listener (예: `jboss-logging`) 는 DevHub 운영과 무관. 별도 SPI 등록 불필요 — DevHub backend 가 Admin REST `?dateFrom=<ISO8601>` polling 으로 끌어옴.

#### 8.6.3 backend env 변수 (운영자가 set)

| env | default | 권장 운영값 | 비고 |
| --- | --- | --- | --- |
| `DEVHUB_KEYCLOAK_EVENT_LISTENER_ENABLED` | `false` | `true` (검증 후) | gate 활성화 — 사전 staging 에서 검증 후 prod 활성화 |
| `DEVHUB_KEYCLOAK_EVENT_LISTENER_INTERVAL` | `30s` | `30s` ~ `60s` | tick 주기. 짧을수록 latency ↓ + Keycloak admin REST 부하 ↑ |
| `DEVHUB_KEYCLOAK_EVENT_LISTENER_MAX_EVENTS` | `500` | `500` | 매 poll page size. Keycloak 의 events 응답 default 100 → 500 으로 확장 |
| `DEVHUB_KEYCLOAK_ADMIN_URL` / `_REALM` / `_CLIENT_ID` / `_CLIENT_SECRET` | (필수) | — | 기존 §3.2 `devhub-backend` confidential client. event polling 도 같은 service account 사용 |

활성화 후 backend 재기동. startup log 확인:

```
[keycloak-event-puller] starting (interval=30s, max_events=500)
keycloak event listener enabled (interval=30s max_events=500)
```

#### 8.6.4 Prometheus dashboard panel (Grafana)

| metric | 의미 | panel 권장 |
| --- | --- | --- |
| `devhub_keycloak_events_processed_total{kind="user",action="..."}` | user event 매핑 결과별 emit 누적 | stacked area, action 별 색상 |
| `devhub_keycloak_events_processed_total{kind="admin",action="..."}` | admin event 매핑 결과별 emit 누적 | 같음 |
| `devhub_keycloak_event_cursor_lag_seconds{cursor_key="..."}` | poll 직후 (now - cursor.LastEventAt) | line chart, cursor_key 별 |
| `devhub_keycloak_event_pull_errors_total{kind="..."}` | user / admin pull 실패 누적 | counter delta, kind 별 |

PromQL 예시:

```promql
# 5분 평균 event 처리율 (분당)
rate(devhub_keycloak_events_processed_total[5m]) * 60

# cursor lag 최대값
max(devhub_keycloak_event_cursor_lag_seconds)

# 최근 5분 pull error 건수
increase(devhub_keycloak_event_pull_errors_total[5m])
```

#### 8.6.5 알람 조건

| 알람 | 조건 | 의미 | 대응 |
| --- | --- | --- | --- |
| `cursor_lag_high` | `max(devhub_keycloak_event_cursor_lag_seconds) > 600 for 5m` | 10분 이상 새 event 미수신 — Keycloak 가용성 또는 backend cron 정지 의심 | §8.6.7 트러블슈팅 참조 |
| `cursor_lag_critical` | `max(devhub_keycloak_event_cursor_lag_seconds) > 3600 for 10m` | 1시간 이상 lag — Keycloak 측 issues 또는 cron goroutine panic | backend 재기동 + Keycloak 헬스체크 |
| `pull_error_rate` | `increase(devhub_keycloak_event_pull_errors_total[5m]) > 5` | 5분간 pull error 5건 이상 — Keycloak admin REST 401/403/timeout 의심 | service account 권한 / token 만료 확인 |

> JWKS metric (§8.4 `devhub_jwks_fetch_total` / `devhub_jwks_cache_age_seconds`) 과 함께 보면 더 정확. JWKS 실패 + event lag 동시 발생 → Keycloak full outage.

#### 8.6.6 audit_logs 정합 + dedup 동작 확인

활성화 후 1~2분 내에 audit_logs 에 `source_type='keycloak_event'` row 가 누적 시작. 확인 query:

```sql
-- 최근 emit 된 Keycloak event
SELECT id, audit_id, action, target_type, target_id, actor_login,
       source_event_id, created_at, payload->>'keycloak_event_type' AS kc_type
FROM audit_logs
WHERE source_type = 'keycloak_event'
ORDER BY created_at DESC
LIMIT 20;

-- dedup 통계 — 같은 source_event_id 가 중복 등장하지 않는지 (UNIQUE INDEX 보장)
SELECT source_event_id, COUNT(*)
FROM audit_logs
WHERE source_type = 'keycloak_event'
GROUP BY source_event_id
HAVING COUNT(*) > 1;
-- ↑ 결과 0 행이 정상 (partial UNIQUE INDEX 가 강제)

-- cursor 위치 확인
SELECT cursor_key, last_event_at, last_event_hash, updated_at FROM event_cursors;
```

#### 8.6.7 트러블슈팅

| 증상 | 1차 의심 | 검증 / 대응 |
| --- | --- | --- |
| audit_logs 에 keycloak_event row 가 전혀 없음 | gate 비활성 / service account 권한 / Events 미활성화 | (1) `DEVHUB_KEYCLOAK_EVENT_LISTENER_ENABLED=true` 확인, (2) startup log 의 `keycloak event listener enabled` 메시지 확인, (3) Keycloak admin console 의 Save events `ON` 확인, (4) service account 의 view-events 권한 확인 |
| `cursor_lag_high` 알람 | Keycloak admin REST 401/403 / timeout | (1) backend log 의 `[keycloak-event-puller] tick: ...` error 메시지 확인, (2) Keycloak admin URL 헬스체크 (`curl ${ADMIN_URL}/realms/master/.well-known/openid-configuration`), (3) service account client_secret 만료 의심 시 §8.3 rotation. (4) Keycloak Events expiration 이 너무 짧아 cursor 가 expire 된 event 를 못 잡는 케이스는 **먼저** Keycloak Events Expiration 을 7d → 30d 연장 + backend 재기동으로 회복 시도. **cursor 수동 advance (`UPDATE event_cursors SET last_event_at = NOW() WHERE cursor_key = 'keycloak.events';`) 는 advance 구간의 event 를 audit_logs 에 영구 누락시키는 최후의 수단** — 운영 incident 가 진행 중이고 lag 가 회복되지 않을 때만, 손실 감수 결정 후 실행. |
| 동일 keycloak_event 가 audit_logs 에 중복 등장 | partial UNIQUE INDEX 손상 / migration 미적용 | (1) `\d audit_logs_source_event_id_uniq` 로 인덱스 존재 확인, (2) 누락 시 migration 000032 재실행, (3) 기존 중복 row 정리: `DELETE FROM audit_logs WHERE id IN (SELECT id FROM audit_logs WHERE source_type='keycloak_event' AND id NOT IN (SELECT MIN(id) FROM audit_logs WHERE source_type='keycloak_event' GROUP BY source_event_id));` |
| `pull_error_rate` 알람 | Keycloak Admin REST 4xx/5xx | backend log 의 `keycloak admin status <code>: ...` 메시지 + Keycloak access log 동시 확인. 가장 흔한 케이스 = client_secret 만료 또는 권한 회수. |
| `keycloak.event.unknown` action 이 metric / audit 에 빈번 등장 | 새 Keycloak event type 도입 후 매핑 표 미갱신 | `internal/audit/keycloak_event_puller.go` 의 `mapUserEventToAudit` / `mapAdminEventToAudit` 에 case 추가 + design 문서 §4.1 / §4.2 표 갱신 (별도 PR) |

#### 8.6.8 disable / rollback 절차

운영 incident 발생 시 빠른 차단:

1. `DEVHUB_KEYCLOAK_EVENT_LISTENER_ENABLED=false` 로 set + backend 재기동 → cron goroutine 정지, audit_logs 신규 INSERT 중단.
2. cursor 는 그대로 유지 — 재활성화 시 마지막 cursor 부터 재개.
3. migration 000031 / 000032 자체는 rollback 불필요 (legacy row 영향 없음 — partial WHERE 가드).
4. 진짜 schema 까지 rollback 필요 시 `migrate down 2` 로 000032 → 000031 순서대로 down (운영 audit 데이터 손실 주의).

#### 8.6.9 잔여 carve out (§8.6 의 sub-carve)

- **(carve)** Keycloak event listener SPI 도입 — 현재 polling 방식이 latency 30s ~ interval. push 기반으로 전환하면 < 1s. SPI plugin 개발 + 사내 운영팀 동반 필요.
- **(carve)** `audit_logs` cold storage archival — keycloak_event row 가 매일 수천 건 누적되면 6개월 이후 cold storage 이관 SOP 필요. 본 운영 SOP 의 scope 외 — DR / 백업 정책 차원.
- **(carve)** dashboard JSON 정식 등록 — 본 SOP 의 PromQL 예시를 Grafana dashboard JSON 으로 ImportExport 화 + git 추적. 환경 별 자산이라 git 추적 외 (사내 Grafana repo).
- **(carve)** alertmanager rule YAML 정식 등록 — 본 SOP §8.6.5 의 3종 알람 (cursor_lag_high / cursor_lag_critical / pull_error_rate) 을 Prometheus `alerting_rules.yml` 정식 자산으로 등록 + 사내 alertmanager routing. 환경 별 자산이라 git 추적 외 (사내 monitoring repo).

## 9. 보안 점검

### 9.1 잠재 위협 + mitigation

| 위협 | mitigation |
| --- | --- |
| client_secret leak | vault 보관 + rotation SOP (§8.3) + Keycloak admin event log 의 `CLIENT_*` 액션 audit |
| JWKS private key leak | §6.5 비상 rotation 절차 |
| token 재사용 (replay) | Keycloak token TTL 짧게 (access 5분 권장), refresh token rotation 활성화 (Keycloak Tokens 탭 → Refresh Token Max Reuse = 0) |
| user attribute 위조 (employee_id) | Keycloak admin event log + 사내 운영팀의 attribute 변경 권한 제한 (system_admin 한정) |
| brute force login | Keycloak Realm Settings → Security Defenses → Brute Force Detection ON (default) |

### 9.2 audit log 통합 (resolved, §8.6 운영 SOP)

- Keycloak 의 event listener / admin event polling → DevHub `audit_logs` 통합은 [ADR-0019 §5.3 (9)](../adr/0019-keycloak-only-idp.md#53-잔여-carve-out) Phase 2 (PR-B~PR-E, sprint -u~-x) 로 backend + 운영 SOP 모두 resolved. 운영 활성화 절차 + dashboard panel + 알람 + 트러블슈팅은 [§8.6](#86-keycloak-event-listener-audit_logs-통합-운영-sop) 참조.

## 10. 잔여 carve out

본 SOP 의 scope 외 — 별도 sprint:

| 항목 | 관련 ADR / 위치 | 비고 |
| --- | --- | --- |
| SSO logout chain (RP-initiated logout) | ADR-0019 §5.3 | Keycloak `end_session_endpoint` + DevHub frontend redirect chain |
| MFA 도입 | ADR-0019 §5.3 | Keycloak MFA / WebAuthn 표준 정책 활성화 |
| Keycloak failover (HA 구성 또는 backup IdP) | ADR-0019 §5.3 | 단일 장애점 회피 |
| off-boarding 즉시성 | ADR-0019 §5.3 | HR 시스템 → Keycloak → DevHub propagation chain |
| `groups` claim → DevHub RBAC role 자동 매핑 | ADR-0019 §5.3 (별도 ADR 후보) | composite role 또는 mapper 로 |
| ~~Keycloak event SPI → DevHub `audit_logs` 통합~~ | ~~ADR-0019 §5.3 + ADR-0019 §4.5~~ | **resolved** — Phase 2 (PR-B~PR-E, sprint -u~-x). polling 기반 구현 + 운영 SOP §8.6. SPI push 전환은 별도 carve (§8.6.9). |
| 사내 LDAP/AD federation | ADR-0019 §5.4 RM-M4-09 | Keycloak User Federation |
| Gitea SSO via Keycloak identity broker | ADR-0019 §5.4 RM-M4-09 | M4 RM-M4 진입 시 |

## 11. 변경 이력

| 일자 | 변경 | sprint |
| --- | --- | --- |
| 2026-05-19 | 1차 draft — §2 realm + §3 client 2종 + §4 role 4종 + §5 user attribute mapper (preferred_username / email / realm_access.roles / employee_id custom) + §6 JWKS rotation 운영 SOP + §7 local embedded vs external 분기 + §8 운영 SOP (생성/회수/secret rotation/장애) + §9 보안 점검 + §10 잔여 carve out 8 항목. [ADR-0019 §5.3 carve out (1)+(2)+(3) resolved](../adr/0019-keycloak-only-idp.md#53-잔여-carve-out) 의 source-of-truth. | `claude/work_260519-c` |
| 2026-05-19 | §8.6.1 cursor seed row + §8.6.2 Expiration wording 정정 — sprint -y codex hotfix #10 (PR #189~#192) 의 P1-C (cursor bootstrap 명시) + P2-E (Expiration 이 운영 outage tolerance 보다 짧으면 audit 영구 손실 위험) 정정. P1-B backend 자동 seed 패치 (`loadCursor` 즉시 UPSERT) 와 정합. | `claude/work_260519-y` |
| 2026-05-19 | §8.6 Keycloak event listener (audit_logs 통합) 운영 SOP 신규 (9 sub-section) — [ADR-0019 §5.3 (9)](../adr/0019-keycloak-only-idp.md#53-잔여-carve-out) Phase 2 PR-E 의 마지막 carve. 활성화 사전 조건 (migration 000031 + 000032 + service account 권한 + Keycloak Events 활성화) / Keycloak admin console 설정 (User+Admin events Save + Expiration 7d) / backend env 3종 (`DEVHUB_KEYCLOAK_EVENT_LISTENER_ENABLED` / `_INTERVAL` / `_MAX_EVENTS`) / Prometheus dashboard 4 panel (events_processed_total / cursor_lag_seconds / pull_errors_total + PromQL 예시) / 알람 3종 (cursor_lag_high 600s / cursor_lag_critical 3600s / pull_error_rate 5건/5분) / audit_logs dedup 동작 확인 query (최근 emit / 중복 검사 / cursor 위치) / 트러블슈팅 5 케이스 (audit row 없음 / cursor lag / 중복 등장 / pull error / unknown action 빈번) / disable/rollback / sub-carve 3 (SPI push 전환 / cold storage archival / dashboard JSON 자산). §9.2 audit log 통합 carve → resolved 표기로 갱신 + §10 의 audit 항목 strikethrough 표기. **ADR-0019 §5.3 (9) Phase 2 모든 carve (PR-B~PR-E) resolved.** | `claude/work_260519-x` |
| 2026-05-20 | §8.5b Self-service 비밀번호 변경 (Keycloak Account Console 위임) 신규 4 sub-section — sprint -ad Kratos 잔재 residual cleanup 의 정합. (a) URL 구성 표 (`${OIDC_ISSUER_URL}/account/` + DevHub `/account` 의 "Open Keycloak Console" 외부 link), (b) 사용자 흐름 4 step, (c) 운영 요점 4건 (DevHub 측 코드 변경 없음 / Web Origins / MFA enrollment 동일 경로 / Audit log 는 sprint -u~-y event listener 가 자동 캡처), (d) sub-carve 2건 (Account Console URL 노출 / MFA enrollment 강제). DevHub 의 `POST /api/v1/account/password` proxy + Kratos login/settings client + frontend password form 모두 제거되어 ADR-0019 정합 완전 정착. | `claude/work_260519-ad` |
