# ADR-0020: 외부 Keycloak 가정 하의 계정/사용자 관리 책임 경계

## 1. 상태
- **상태**: Accepted (**partial supersession 2026-05-21** — §3.2 "신규 user 의 unit 초기 배치" row + §4.1 sub-carve B 의 lazy auto-create 실 구현 결정 + §4.2 의 lazy auto-create 보안 영향 + §6.1 후속 ADR 후보 row + §6.2 carve out 의 동일 항목 (5 위치) 은 [ADR-0021](./0021-onboarding-self-service-unit-selection.md) 가 supersede. **핵심 결정** (옵션 A — Keycloak account vs DevHub user 책임 경계, `rbac_subject_roles` 제거, service account 권한 축소) 은 **변경 없이 유지**.)
- **작성일**: 2026-05-20
- **수정일**: 2026-05-20
- **결정 근거 sprint**: `claude/work_260520-a` (Phase 1 현황 파악) + `claude/work_260520-b` (Phase 2 명시 결정 6건 확정) + `claude/work_260520-d` (Phase 3 실 구현 진입 + 본 ADR 발급)
- **partial superseded by**: [ADR-0021 Onboarding self-service unit selection + lazy auto-create supersession (2026-05-21)](./0021-onboarding-self-service-unit-selection.md) — lazy auto-create 결정 (§3.2 + §4.1 sub-carve B + §4.2 + §6.1 후속 ADR 후보 row + §6.2, **5 위치**) 의 partial supersession. §3.2 의 "user 조직 unit assignment" row 의 책임 주체는 "DevHub admin 단독" → "사용자 self-service onboarding + DevHub admin 검토" 로 확장 (reversal 아닌 확장).
- **관련 문서**: [docs/domain/auth-session/account_redesign.md](../domain/auth-session/account_redesign.md) (Phase 1 매트릭스 + Phase 2 design + Phase 3 실행 계획), [ADR-0019 Keycloak 단일화](./0019-keycloak-only-idp.md), [ADR-0021 Onboarding (partial supersedes 본 ADR)](./0021-onboarding-self-service-unit-selection.md), [ADR-0011 RBAC row-scoping](./0011-rbac-row-scoping.md), [ADR-0008 HRDB production adapter](./0008-hrdb-production-adapter.md), [keycloak_operations.md (§8.5b self-service + §8.5c governance, 후속 sprint 신규)](../setup/keycloak_operations.md), [keycloak_groups_rbac_mapping.md](../domain/rbac-permissions/keycloak_groups_mapping.md)

> **⚠️ partial supersession 안내 (2026-05-21)**: 본 ADR 의 §3.2 의 "신규 user 의 unit 초기 배치" row + §4.1 sub-carve B 의 "authenticateActor lazy auto-create 실 구현" + §4.2 의 lazy auto-create 보안 영향 + §6.1 후속 ADR 후보 row + §6.2 carve out 의 동일 항목 (5 위치) 은 [ADR-0021 §3.3](./0021-onboarding-self-service-unit-selection.md#33-lazy-auto-create-폐기-adr-0020-부분-supersession) 가 supersede 한다. 현재 운영의 lazy auto-create 결정은 **폐기** — onboarding 완료 시점에 user row 가 INSERT 된다. §3.2 의 "user 조직 unit assignment" 책임 주체도 [ADR-0021 §3.1](./0021-onboarding-self-service-unit-selection.md#31-책임-경계-확장--self-service-unit-selection-허용) 의 확장 표로 supersede (사용자 self-service onboarding + admin 검토). 본 ADR 의 **핵심 결정** (옵션 A 책임 경계, `rbac_subject_roles` 제거, service account 권한 축소) 은 변경 없이 유지. 본문 본문은 immutable 보존.

## 2. 컨텍스트

### 2.1 ADR-0019 후속 — 외부 Keycloak 시나리오의 책임 경계 모호성

[ADR-0019 (2026-05-19)](./0019-keycloak-only-idp.md) 가 DevHub 의 IdP 를 Keycloak 단일화로 확정한 후, **계정 관리** + **사용자 관리** 의 코드 + UI + DB 책임이 4 경로로 분산된 상태가 남아 있었다:

1. **`/api/v1/accounts/*`** — Keycloak Admin REST 의 proxy (service account 가 manage-users role 사용)
2. **`/api/v1/users/*`** — 조직 메타데이터 (status / role / unit assignment) CRUD
3. **`/api/v1/organization/*`** — 조직 단위 (units / hierarchy)
4. **`/api/v1/rbac/*`** — RBAC policy + subject-role assignment

본 4 경로 의 책임이 **외부 Keycloak (사내 IdP 팀이 별도 운영)** 시나리오에서 모호하다.

### 2.2 Phase 1 — 현황 파악 매트릭스 (sprint `-a`, PR #199, 2026-05-20)

`docs/domain/auth-session/account_redesign.md` 신규 (235 lines) — 책임 분산 매트릭스 13 row + 17 backend endpoint + 4 frontend page + DB schema 정합. **핵심 발견 5건**:

1. **role 이중 source-of-truth** — Keycloak group composite (`token realm_access.roles`) 가 backend 평가 1차, `users.role` 컬럼은 화면 표시 cache. sync mechanism 없음 → divergent 위험
2. **status 이중 source** — Keycloak `enabled` + DevHub `users.status`. `/api/v1/accounts/:id` PATCH 만 atomic. Keycloak admin console 직접 변경 시 DevHub stale
3. **POST `/accounts` vs POST `/users` 중복** — 둘 다 신규 user 생성, atomic 단일 작업 아님
4. **dead frontend code** — `unlockAccount` + `deleteAccount` UI 미사용
5. **`PUT /rbac/subjects/:id/roles` UI 미구현** — backend-only, 실 권한은 token claim 우선이라 의미 약함

§4.5 에서 Phase 2 입력 옵션 A~D 후보 도출:
- **A** — DevHub admin endpoint 전면 폐기, Keycloak admin 직접 + HRDB ETL push + token 검증 시 lazy users row 자동 생성
- **B** — 현재 상태 유지 (Admin Client proxy)
- **C** — Hybrid (write 일부만 — password reset / disable 만)
- **D** — Read-only DevHub admin (GET 만 + SCIM bridge)

### 2.3 Phase 2 — 명시 결정 6건 (sprint `-b`, PR #200, 2026-05-20)

사용자 Q&A 6 round 토론 후 6건 결정 + Phase 2 책임 분리 design (§5 신규 10 sub-section).

## 3. 결정

### 3.1 명시 결정 6건 종합

| # | 영역 | 결정 | 근거 |
| --- | --- | --- | --- |
| **A** | `/api/v1/accounts/*` 4 endpoint 향방 | **전면 폐기** (옵션 A) | 사용자 조건: IdP 팀 별도 운영, DevHub 운영자 manage-users 권한 없음. 자주 쓰는 동작 (user 목록 view + 조직 unit/role assignment) 은 `/api/v1/users` 계열 (Keycloak 무관). divergence 원천 차단 + service account 권한 축소 (`manage-users` 제거) + dead UI 메서드 자연 정리 |
| **B** | `/login` page 향방 | **entry minimal page 유지** | DevHub brand 첫인상 + 명시 로그인 step + error message 표시 + `/login` ↔ `/auth/login` 중복 정리 |
| **C** | role/status sync mechanism | **event listener 확장 (sprint -u~-y 자연 확장)** + lazy backfill/auto-create hot path 1회 | (a) token 검증 write-through, (b) stale 비교 모두 token claim 한계 (group_membership / status change 감지 불가). event listener 만 정확. hot path 영향 0, latency 30s 는 access_token 5분 stale 안에 묻힘 |
| **D** | `rbac_subject_roles` (endpoint + store + interface) | **완전 제거** | 발견: 별도 테이블 없음, `users.role` 컬럼 직접 write. 결정 C 와 충돌 (event listener 가 곧 덮어쓰기). `PATCH /api/v1/users/:id` 와 기능 중복. ADR-0011 row-scoping 과 무관 |
| **E** | read-only 모드 carve (Keycloak down 시 GET grace period) | **도입 안 함** (self-reverse) | signature 검증 skip 은 OIDC 표준 위반 + token forgery 위험 + revoked / audit forgery. 명시 결정 F 가 진짜 정공법 |
| **F** | JWKS stale-while-error 확장 (expiry mismatch) | **확장 도입** | sprint -r (PR #186) kid mismatch fallback 의 자연 확장. signature 검증 유지 (token forgery 위험 없음) + uptime → key rotation period (90일) 까지. revoked key 시나리오는 별도 mitigation |

### 3.2 책임 경계 결정 (옵션 A 채택)

> **[2026-05-21 partial supersession]**: 본 §3.2 표의 "user 조직 unit assignment" row 의 책임 주체 (DevHub admin 단독) 는 [ADR-0021 §3.1](./0021-onboarding-self-service-unit-selection.md#31-책임-경계-확장--self-service-unit-selection-허용) 가 **확장 (reversal 아님)** — 사용자 self-service onboarding + DevHub admin 검토 (`reviewed` transition) 으로 분기. 본 §3.2 표의 "신규 user 의 unit 초기 배치" row 의 lazy-auto-create-후 정책은 [ADR-0021 §3.3](./0021-onboarding-self-service-unit-selection.md#33-lazy-auto-create-폐기-adr-0020-부분-supersession) 가 **supersede** — lazy auto-create 폐기 + onboarding 제출 시점에 row INSERT. 본 §3.2 표의 다른 row (Keycloak admin 책임 / `users.role` 직접 수정 금지 / 조직 unit CRUD / RBAC policy 편집) 는 변경 없이 유지.

| 운영 동작 | 책임 주체 | 도구 |
| --- | --- | --- |
| user 생성 (account.create) | **Keycloak admin** (IdP 팀) | Keycloak admin console **또는** HRDB ETL push (`scripts/hrdb_etl_sync.sh`) |
| user 비밀번호 reset | **Keycloak admin** (IdP 팀) | Keycloak admin console (Credentials 탭) |
| user status disable / enable | **Keycloak admin** (IdP 팀) | Keycloak admin console **또는** HRDB ETL push |
| user 삭제 | **Keycloak admin** (IdP 팀) | Keycloak admin console |
| group membership (role 변경) | **Keycloak admin** (IdP 팀) | Keycloak admin console (User detail → Groups) |
| **user 조직 unit assignment** | **DevHub admin** | DevHub `/admin/settings/users` (PATCH `/api/v1/users/:id`) |
| **신규 user 의 unit 초기 배치** | **DevHub admin** (lazy auto-create 후) | (a) HRDB ETL pre-stage 동반 자동 매핑, (b) 미동반 시 unit 미할당 (`primary_unit_id=NULL`) 으로 lazy create — 첫 API call 차단 안 함 |
| **`users.role` 직접 수정** | **금지** (event listener 가 자동 sync) | — |
| 조직 unit (department/team) CRUD | **DevHub admin** | DevHub `/admin/settings/organization` (Keycloak 무관) |
| RBAC policy (role × resource × action) 편집 | **DevHub admin** | DevHub `/admin/settings/permissions` |

### 3.3 결정 동인

- **divergence 원천 차단** — single source-of-truth (Keycloak)
- **운영 거버넌스 명확화** — IdP 팀 vs DevHub admin 책임 분리
- **보안 권한 최소화** — service account 의 `manage-users` role 제거 → `view-users` + `view-events` 만
- **OIDC 표준 정합** — signature 검증 의무 (결정 E self-reverse)

## 4. 결과 / 영향

### 4.1 backend code 변경 (Phase 3 sub-carve)

> **[2026-05-21 partial supersession]**: 본 §4.1 표의 **sub-carve B (backend)** 항목의 "authenticateActor lazy auto-create 실 구현" 결정은 [ADR-0021 §3.3](./0021-onboarding-self-service-unit-selection.md#33-lazy-auto-create-폐기-adr-0020-부분-supersession) 가 supersede — lazy 폐기 + token-only actor 흐름 + onboarding 완료 시점 row INSERT 로 전환. sub-carve B 의 다른 변경 (`/api/v1/accounts/*` 4 endpoint 제거 + `KeycloakAdminClient` write 메서드 호출처 제거) 은 변경 없이 유지. 신규 audit event `account.lazy_provisioned` / `user.role_default_assigned` 의 신규 emit 은 본 ADR 시점부터 중단되나 audit_logs 의 기존 row 는 immutable 보존.

| sub-carve | 영역 | sprint |
| --- | --- | --- |
| **A** | ADR-0020 발급 + design doc §6 실행 계획 + `rbac_subject_roles` 완전 제거 (결정 D, §5.8 따름) | ✅ `-d` (PR #205 `f2a389a`) |
| **B (backend)** | `/api/v1/accounts/*` 4 endpoint 제거 + `KeycloakAdminClient` write 메서드 호출처 제거 + `authenticateActor` lazy auto-create 실 구현 (결정 A + §5.2) + audit action `account.lazy_provisioned` / `user.role_default_assigned` 신규 | ✅ `-i` (PR TBD, sprint `claude/work_260520-i-209-accounts-deprecation`) |
| **B (frontend)** | frontend `account.service.ts` 폐기 + admin/settings/users page 의 admin actions 제거 + e2e TC-ACC-* 갱신 | `-?` (Gemini 별도 sprint) |
| **C** | event listener 확장 — USER:UPDATE / GROUP_MEMBERSHIP / USER:DELETE 매핑 + DevHub `users` write + metric 3종 (결정 C + §5.3) | ✅ `-k` (sprint `claude/work_260520-k-212-event-listener-users-sync`, PR TBD) |
| **D** | JWKS stale-while-error expiry case 확장 (결정 F + §5.6) | ✅ `-l` (sprint `claude/work_260520-l-213-jwks-stale-expiry`, PR TBD) |
| **E** | service account 권한 축소 + governance 협약 SOP (`docs/infrastructure/keycloak-idp/service_account_min_role.md` 신규, §5.5). 옵션 A 정공법 (호출처 전면 제거) 채택. organization.go POST /users password 분기 제거 + seedLocalAdmin 함수 + KeycloakAdminClient write methods 4건 + IdentityAdmin interface 정리 | ✅ `-n` (sprint `claude/work_260520-n-214-service-account-min-role`, PR TBD) |
| **F** | `/login` page 정리 (결정 B + §5.7) | ✅ `claude/work_260522-adr-0020-subcarve-f-login` (PR TBD) — `/login` canonical page swap + `?error=` 처리 + `/auth/login` stub 보존 + AuthGuard 401 fallback `/login?error=session_expired` |

#### 4.1.1 Phase 3 closing status (2026-05-22)

본 ADR 발급 이후 sub-carve 진행 결과 — **8 carve 중 7 closed + 1 사내 동반**.

| Sub-carve | 머지 | SHA | 비고 |
| --- | --- | --- | --- |
| A | PR #205 | `f2a389a` | sprint `-d` |
| B (backend) | PR #239 | `d21e801` | sprint `-i`. lazy auto-create 부분은 PR #290 `fa042c5` 가 ADR-0021 §3.3 따라 폐기 |
| B (frontend) | PR #246 | `b1e34bd` | gemini `work_260520-a-209-accounts-cleanup` |
| C | PR #241 | `9ea7e1c` | sprint `-k` |
| D | PR #242 | `cb6646d` | sprint `-l` |
| E | PR #244 | `6810384` | sprint `-n` |
| **F** | PR TBD | TBD | sprint `claude/work_260522-adr-0020-subcarve-f-login` — `/login` canonical page swap (`/auth/login` 96 LoC → `/login` + `?error=` 처리) + **`/auth/login` 완전 제거** (사용자 결정 옵션 B, 운영 정합 일괄) + AuthGuard 401 fallback `/login?error=session_expired` + 비-401 fallback `/login?error=login_failed` + 호출처 8 위치 sync + infra/scripts/5 docs 의 `post.logout.redirect.uris` allowlist URI 정합 (`/devhub/auth/login` → `/devhub/login`). **사내 운영자 1회 작업 필요** — Keycloak admin console 의 client `devhub-frontend` 의 `Valid Post Logout Redirect URIs` 에서 `/devhub/auth/login` 제거 + `/devhub/login` 추가. realm.prod.json 재 import 또는 setup-keycloak.sh 재실행으로 자동 정합 가능. worker_division 사용자 명시 override 진입 (`feedback_worker_division_override`). |
| **SPI provider JAR** | — | — | **사내 동반 P2** — Keycloak SPI Java 빌드 + Maven/Gradle 자산. 사내 인프라 결정 동반. |

**핵심 결정 (옵션 A 책임 경계 / `rbac_subject_roles` 제거 / service account 권한 축소)** 은 6 closed carve 로 완전 적용. 추가 F closing 으로 frontend UX 정리도 완료 — 잔여는 SPI JAR (사내 동반) 만.

§6.3 사내 동반 carve 3건 (HRDB ETL push unit stage / Keycloak admin 운영 SOP 승격 / JWKS rotation 직후 cache flush SOP) 은 본 ADR 의 결정 적용 후속 운영 자산 — claude 가 docs 초안 작성 가능, 사내 실 적용은 IdP 팀 / 운영팀 동반.

### 4.2 보안 영향

- service account 권한 축소 (`manage-users` 제거) → 최소 권한 원칙 정합. service account compromised 시 user lifecycle 조작 불가
- JWKS stale-while-error expiry case 확장 → Keycloak unreachable 시 uptime 90일 까지 (key rotation 주기). signature 검증 유지로 token forgery 위험 없음 — rotation 직후 backend cache flush SOP 필요 (별도 carve)
- ~~lazy auto-create → token 검증 성공한 user 만 lazy create (Keycloak user lifecycle 정책이 1차 필터). enumeration 위협 없음~~ — **[2026-05-21 supersession]** [ADR-0021 §3.3](./0021-onboarding-self-service-unit-selection.md#33-lazy-auto-create-폐기-adr-0020-부분-supersession) 의 lazy 폐기로 본 항목 무효. 현재 운영은 `onboardingGate` middleware 가 미완료 사용자의 도메인 API 접근을 403 차단 — enumeration 방어 더 강화. ADR-0021 §4.3 참조.

### 4.3 운영 영향

- DevHub admin UI 의 "Issue Account" / "Reset Password" / "Disable Account" 액션 **제거** — IdP 팀 책임으로 이관
- 신규 user 의 첫 진입 시점에 unit 미할당 상태로 lazy create — DevHub admin 이 후속 `/admin/settings/users` filter (unit_id=null) 로 식별 + 배치
- HRDB ETL push 가 unit 정보 pre-stage 한 경우 → 자동 매핑 (sprint -p `hrdb_etl_sync.sh` 확장 carve)

### 4.4 ADR governance

- ADR-0019 의 §5.3 (8) groups → RBAC role 자동 매핑 design (sprint -f PR #174) 의 후속 — event listener 확장 매핑이 본 결정의 자연 통합
- ADR-0011 row-scoping 과 무관 (rbac_subject_roles 제거가 row-scoping 정책에 영향 없음 — `users.role` 컬럼 자체는 보존)

## 5. 대안 / 거부된 옵션

### 5.1 옵션 B — 현재 상태 유지 (Admin Client proxy)
- **거부 이유**: source-of-truth 이중화 issue 재발 + service account `manage-users` 권한 유지 (보안 위험)

### 5.2 옵션 C — Hybrid
- **거부 이유**: 부분 정합 — `account.disable` 만 DevHub UI 유지하면 DevHub `users.status` 와 Keycloak `enabled` 의 sync 가 일부 경로만 정합. divergence 가능성 축소되나 제거 안 됨

### 5.3 옵션 D — Read-only mirror
- **거부 이유**: 사용자 조건 (IdP 팀 별도 운영, manage-users 권한 없음) 으로 read 도 token 검증만으로 충분 — Keycloak admin REST 호출 자체 불필요. SCIM bridge 도입은 over-engineering

### 5.4 결정 E self-reverse — read-only mode carve (Keycloak down 시)
- **본인 제안 → self-reverse** — signature 검증 skip 은 token forgery 안티패턴. 진짜 정공법은 결정 F (JWKS stale-while-error expiry case 확장)

## 5.5 codex review hotfix (sprint -d Stage 3, P1 응답)

PR #205 의 codex review (P1, 2026-05-20) 가 본 ADR 결정 D (`rbac_subject_roles` 폐기) 의 backward compatibility 회귀 발견:

> Removing the subject-role routes here eliminates the only backend path that could assign arbitrary existing RBAC roles to users. The remaining user update path still validates role against a fixed allowlist (`developer|manager|system_admin` in `organization.go`), so `custom-*` roles and even `team_manager` can no longer be assigned via API.

### 응답 — 결정 D 의 호환성 보강

본 ADR 의 결정 C (event listener 확장, sprint -f 후속) 가 완성되면 DevHub UI 의 role assignment 가 폐기되고 Keycloak admin console group membership 변경 → event listener sync 만이 유일 경로. 그러나 sprint -d 만 머지된 상태에서는 sub-carve C 미완성으로 일시적 호환 필요.

#### sprint -d Stage 3 hotfix

`backend-core/internal/httpapi/organization.go` 의 `validAppRoles` map 에 `team_manager` 추가 (이전: `developer/manager/system_admin` 3개만, sprint -d 까지는 `/rbac/subjects/:id/roles` 가 backup 경로였음). error message 도 4 role 명시.

| 영역 | sprint -d 이전 | sprint -d Stage 3 후 | sub-carve C (sprint -f) 완성 후 |
| --- | --- | --- | --- |
| `PATCH /api/v1/users/:id` role allowlist | `developer/manager/system_admin` | `developer/manager/team_manager/system_admin` | (sub-carve B 진입 시 본 endpoint 자체 폐기 검토) |
| `PUT /api/v1/rbac/subjects/:id/roles` (custom role 포함 임의 role 허용) | ✅ | ❌ 폐기 (sprint -d) | (변경 없음) |
| Keycloak admin console group membership → event listener sync → `users.role` write | (없음) | (없음 — 사내 운영자가 token refresh 기다림) | ✅ 자동 sync |

#### custom role 처리

sprint -d 이후 custom role (예: `pmo_director`, `qa_lead` 등 `rbac_policies` 의 user-defined role) 은 DevHub API 로 직접 할당 불가. 사내 운영자가 임시로 사용 시 옵션:
1. **권장 (Phase 3 정공법)** — Keycloak admin console 에서 group composite role 매핑 + DevHub event listener (sprint -f 후속) 가 자동 sync
2. **임시 우회 (sprint -f 미완성 동안)** — DB direct UPDATE `users.role` (`users.role` FK to `rbac_policies.role_id` 가 그대로 보호). 운영자 SOP 동반 carve
3. **신규 API endpoint 발급** — `validAppRoles` 를 `rbac_policies` dynamic lookup 으로 변경 (handler + store 변경 필요). 본 ADR 의 결정 C 와 충돌 (event listener 가 곧 덮어쓰기) → **거부**

본 sprint -d 의 Stage 3 hotfix 는 옵션 (1) 의 임시 호환만 보장 (`team_manager` 명시 추가) + 옵션 (2) 의 DB direct UPDATE SOP 는 sub-carve E (governance SOP, sprint -h) 가 명문화 예정.

#### 회귀 test (sprint -d Stage 3)

- `TestCreateUserAcceptsPMOManager` — `POST /api/v1/users` 의 `role: team_manager` 가 201 반환
- `TestUpdateUserAcceptsPMOManagerRole` — `PATCH /api/v1/users/:id` 의 role 필드를 `team_manager` 로 변경 시 200 + role 갱신 확인

## 6. 미해결 / 후속 작업

### 6.1 후속 ADR 후보
- ~~(없음 — 본 ADR 이 Phase 2 결정 6건 모두 cover)~~ — **[2026-05-21]** [ADR-0021 Onboarding self-service unit selection](./0021-onboarding-self-service-unit-selection.md) 가 발급되어 본 ADR 의 lazy auto-create 결정 (§3.2 / §4.1 / §4.2 / §6.2) 을 partial supersede.

### 6.2 carve out (Phase 3 sprint 영역)
- backend `/api/v1/accounts/*` 4 endpoint 제거 (§4.1 sub-carve B)
- ~~`authenticateActor` lazy auto-create 실 구현 (§4.1 sub-carve B)~~ — **[2026-05-21 supersession]** [ADR-0021 §3.3](./0021-onboarding-self-service-unit-selection.md#33-lazy-auto-create-폐기-adr-0020-부분-supersession) 가 lazy 폐기 + onboarding 완료 시점 row INSERT 로 supersede.
- event listener 확장 (§4.1 sub-carve C)
- JWKS stale-while-error expiry case 확장 (§4.1 sub-carve D)
- service account 권한 축소 + governance SOP (§4.1 sub-carve E)
- frontend `account.service.ts` 폐기 + admin/settings/users UI 정리 (§4.1 sub-carve B 통합)
- `/login` page 정리 (§4.1 sub-carve F)

### 6.3 사내 동반 carve

> **[2026-05-22 docs 초안 resolved]** sprint `claude/work_260522-internal-coordinated-carve-docs` 가 3 carve 의 docs 초안 신규. 사내 실 적용 (IdP 팀 / HRDB 팀 / 운영팀 동반) 은 별도.

- ✅ resolved (docs 초안) — [HRDB ETL push 의 unit 매핑 정보 stage](../setup/hrdb_unit_pre_stage.md) (옵션 A Keycloak user attribute 권장 / 옵션 B DevHub hrdb.persons 보조 / 옵션 C self-service only)
- ✅ resolved (docs 초안) — [Keycloak admin 운영 SOP 협약](../governance/keycloak_admin_responsibility.md) (§3.2 표를 사내 정책 문서로 승격 — IdP 팀 vs DevHub 운영자 책임 매트릭스 5 sub-section + escalation 4 level + 명시 금지 5건)
- ✅ resolved (docs 초안) — [JWKS rotation 직후 backend cache flush SOP](../setup/jwks_rotation_cache_flush.md) (강제 재기동 정공법 4 환경 + 검증 4 step + cache flush endpoint carve P3)

## 7. 변경 이력

| 일자 | 변경 | sprint |
| --- | --- | --- |
| 2026-05-20 | draft + Accepted — Phase 2 명시 결정 6건 종합 (옵션 A 전면 폐기) + Phase 3 sub-carve 6건 분리 plan | `claude/work_260520-d` |
| 2026-05-20 | Stage 3 보강 — PR #205 codex review P1 응답. `validAppRoles` 에 `team_manager` 추가 (sprint -d 시점 backward compat) + §5.5 신규 hotfix 섹션 + 회귀 test 2건 (TestCreateUserAcceptsPMOManager + TestUpdateUserAcceptsPMOManagerRole). custom role 처리는 옵션 (1) Keycloak admin console + event listener (sprint -f 권장) / 옵션 (2) DB direct UPDATE (sub-carve E 의 SOP 동반) 으로 명시. | `claude/work_260520-d` |
| 2026-05-20 | **sub-carve B (backend) resolved** — `/api/v1/accounts/*` 4 endpoint 제거 + `authenticateActor` lazy auto-create 실 구현 (`lazy_auto_create.go::lazyAutoCreateUser`) + `AuthenticatedActor` Email/DisplayName 필드 추가 + `keycloak_verifier::extractDisplayName` + 신규 audit action 2종 (`account.lazy_provisioned` / `user.role_default_assigned`) + 회귀 test 5건 + 기존 test 3건 admin pre-seed fix. §4.1 sub-carve 표 갱신 (B backend done, B frontend 별도 sprint, C/D/E/F sprint label shift). | `claude/work_260520-i-209-accounts-deprecation` |
| 2026-05-20 | **sub-carve D resolved** — JWKS stale-while-error expiry case 확장. 5 commit: (1) cache struct 확장 (`cachedAt time.Time` + `MaxStaleDuration` 필드 + `defaultJWKSMaxStale = 24h`) + `readStaleCachedKeys` helper. (2) `fetchJWKS` 흐름 변경 — `fetchAndCacheJWKS` 별도 함수로 분리 + network fetch 실패 시 stale fallback 분기 + log WARN. (3) `internal/auth/metrics.go` 신규 — `devhub_jwks_stale_while_error_total{result}` CounterVec + `devhub_jwks_stale_age_seconds` Histogram (ExponentialBuckets 1m~4096m). (4) Config `OIDCJWKSMaxStaleDuration` (env `DEVHUB_OIDC_JWKS_MAX_STALE_DURATION`) + main.go wire (parse → verifier.MaxStaleDuration set + log + invalid fallback). (5) 회귀 test 4건 — StaleWhileError_KeycloakUnreachable (Keycloak 500 시 stale 통과) / StaleExpired_Fails401 (MaxStaleDuration 초과 시 401) / FreshCache_NoStaleFallback (cache 안 fresh 시 network 0 회) / StaleFallback_DefaultMaxStale (env 미설정 시 24h default 적용). §4.1 sub-carve 표 D done 마킹. | `claude/work_260520-l-213-jwks-stale-expiry` |
| 2026-05-20 | **sub-carve C resolved** — Keycloak admin event 처리 시 DevHub `users` 컬럼 자동 sync. 5 commit. (1) `KeycloakAdminClient.GetUserDetails` + `GetUserGroups` 신규. (2) `audit/user_sync.go` 신규 — `SyncUserProfile`/`SyncUserMembership`/`MarkUserDeactivated` + helper (`composeDisplayName`/`pickHighestPriorityRole`/`groupNameToRole`/`ParseIdentityIDFromResourcePath`) + `UserSyncOrgStore`/`UserSyncAdminClient` narrow interface + `SyncUserAction` enum. (3) `keycloak_event_puller.go` 확장 — `KeycloakEventPullerOptions.UserSync UserSyncCallback` 신규 + `classifyAdminEventForSync` helper + `mapAdminEventToAudit` 에 `GROUP_MEMBERSHIP:CREATE/DELETE` 2 row 추가 (10 row 총) + admin event loop 분기 추가 (audit emit + sync callback). (4) `audit/metrics.go` 확장 — 신규 metric 3종 (`devhub_keycloak_user_sync_total{action}` + `_errors_total` + `_lag_seconds` Histogram) + 4 observe helpers + 회귀 test 4건 (GROUP_MEMBERSHIP 매핑 + classifyAdminEventForSync 5 case + InvokesUserSyncCallback + NilUserSync backward compat). (5) `main.go` wire — `UserSync` callback dispatcher (action 별 SyncUserProfile / Membership / MarkUserDeactivated 호출 + metric observe + error log). backward compatible (UserSync nil = sprint -u~-y 동작 동등). §4.1 sub-carve 표 갱신 (C done). | `claude/work_260520-k-212-event-listener-users-sync` |
| 2026-05-20 | **sub-carve E resolved** — service account 권한 축소 정공법 (옵션 A 전면 호출처 제거). 5 commit. (1) `organization.go` `POST /api/v1/users` `req.Password` 분기 제거 + `createUserRequest.Password` field 폐기 + `audit_logs.details` dead key `kratos_id` 제거. (2) `main.go` `seedLocalAdmin` 함수 + 호출 + `seedOrgStore` interface 완전 제거. `main_test.go` 전체 삭제 (seedLocalAdmin 전용 test 3건 + `idpAdminFake` + `orgStoreFake`). (3) `KeycloakAdminClient.CreateIdentity` / `UpdateIdentityPassword` / `SetIdentityState` / `DeleteIdentity` 4 method 제거 + `keycloakIDFromLocation` dead helper 제거. `IdentityAdmin` interface 의 write method 4건 제거 (`FindIdentityByUserID` 만 view-users role 유지). `MockIdentityAdmin` + `keycloak_admin_client_test.go` 정리. (4) `docs/infrastructure/keycloak-idp/service_account_min_role.md` 신규 (현황 매트릭스 13 row + 옵션 A/B/C 비교 + 옵션 A 채택 + 운영 SOP 5 sub-section) + §4.1 sub-carve E done 마킹. (5) traceability + memory. backend service account 가 view-users + view-events realm role 만 요구 — Keycloak 운영자가 `manage-users` 제거 가능 (사내 운영팀 후속). | `claude/work_260520-n-214-service-account-min-role` |
| 2026-05-21 | **partial supersession by ADR-0021** — [ADR-0021](./0021-onboarding-self-service-unit-selection.md) 가 본 ADR 의 lazy auto-create 결정 (§3.2 신규 user unit 초기 배치 row + §4.1 sub-carve B 의 lazy auto-create 실 구현 + §4.2 lazy auto-create 보안 영향 + §6.2 carve out 의 동일 항목) 을 supersede. §3.2 의 "user 조직 unit assignment" 책임 주체는 사용자 self-service onboarding + admin 검토로 **확장**. 본 ADR 의 핵심 결정 (옵션 A 책임 경계 / `rbac_subject_roles` 제거 / service account 권한 축소) 은 변경 없이 유지. 메타 헤더 + §3.2/§4.1/§4.2/§6.1/§6.2 4 위치에 inline supersession banner 추가 (메모리 `feedback_adr_supersession_pattern` 패턴). | `claude/onboarding-adr-2026-05-21` |
| 2026-05-22 | **Phase 3 closing status 명문화** — §4.1.1 신규 sub-section (sub-carve 8 closing 표 + 핵심 결정 적용 완료 + 잔여 F/SPI active 표기). main flat memory directive 의 "ADR-0020 Phase 3 8 carve 진입" 표현이 misleading 했던 점 (잔여 active 는 F + SPI 만) 정정. | `claude/work_260522-adr-0020-phase3-closing-housekeeping` |
| 2026-05-22 | **sub-carve F resolved** — `/login` 이 canonical entry page (결정 B). `/auth/login` 96 LoC → `/login` 본문 swap + `?error=` query 처리 (`session_expired/login_failed/unauthorized + error_description` 5 케이스) + **`/auth/login` 완전 제거** (사용자 결정 옵션 B, 사내 운영자 Keycloak admin console allowlist 1회 갱신 동반) + AuthGuard 401 fallback `/login?error=session_expired` + 비-401 fallback `/login?error=login_failed` + 호출처 8 위치 sync (AuthGuard 2 + onboarding 1 + auth/callback 1 + auth/error 1 + auth/signup 1 + role-routing.test 1 + 1 신규 `/login` test) + infra (realm.prod.json + nginx template + setup-keycloak.sh) 3건 + 5 docs (docker-packaging / environment-setup / keycloak_operations / single_port_deployment / e2e-test-guide) 의 `/devhub/auth/login` allowlist URI 정합. vitest 회귀 가드 신규 (`resolveErrorMessage` unit 6건 + AuthGuard fallback 2건). §4.1.1 표 F → done 표기. | `claude/work_260522-adr-0020-subcarve-f-login` |
| 2026-05-22 | **§6.3 사내 동반 carve 3건 docs 초안 resolved** — sprint `claude/work_260522-internal-coordinated-carve-docs`. (1) [`docs/governance/keycloak_admin_responsibility.md`](../governance/keycloak_admin_responsibility.md) 신규 (§3.2 표를 사내 정책 문서로 승격 — IdP 팀 vs DevHub 운영자 책임 매트릭스 5 sub-section + escalation 4 level + 명시 금지 5건 + 변경 절차). (2) [`docs/setup/jwks_rotation_cache_flush.md`](../setup/jwks_rotation_cache_flush.md) 신규 (revoked key 위협 대응 — backend 강제 재기동 4 환경별 + 검증 4 step + cache flush endpoint carve P3). (3) [`docs/setup/hrdb_unit_pre_stage.md`](../setup/hrdb_unit_pre_stage.md) 신규 (옵션 A Keycloak user attribute 권장 / 옵션 B DevHub hrdb.persons 보조 / 옵션 C self-service only + ADR-0021 §6.2 onboarding cross-check 결정 옵션 3가지). 사내 실 적용 (IdP 팀 / HRDB 팀 / Security sign-off) 은 별도. §6.3 carve list 3건 모두 docs 초안 마킹. | `claude/work_260522-internal-coordinated-carve-docs` |
