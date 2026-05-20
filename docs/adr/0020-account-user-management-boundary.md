# ADR-0020: 외부 Keycloak 가정 하의 계정/사용자 관리 책임 경계

## 1. 상태
- **상태**: Accepted
- **작성일**: 2026-05-20
- **수정일**: 2026-05-20
- **결정 근거 sprint**: `claude/work_260520-a` (Phase 1 현황 파악) + `claude/work_260520-b` (Phase 2 명시 결정 6건 확정) + `claude/work_260520-d` (Phase 3 실 구현 진입 + 본 ADR 발급)
- **관련 문서**: [docs/planning/account_user_management_redesign.md](../planning/account_user_management_redesign.md) (Phase 1 매트릭스 + Phase 2 design + Phase 3 실행 계획), [ADR-0019 Keycloak 단일화](./0019-keycloak-only-idp.md), [ADR-0011 RBAC row-scoping](./0011-rbac-row-scoping.md), [ADR-0008 HRDB production adapter](./0008-hrdb-production-adapter.md), [keycloak_operations.md (§8.5b self-service + §8.5c governance, 후속 sprint 신규)](../setup/keycloak_operations.md), [keycloak_groups_rbac_mapping.md](../planning/keycloak_groups_rbac_mapping.md)

## 2. 컨텍스트

### 2.1 ADR-0019 후속 — 외부 Keycloak 시나리오의 책임 경계 모호성

[ADR-0019 (2026-05-19)](./0019-keycloak-only-idp.md) 가 DevHub 의 IdP 를 Keycloak 단일화로 확정한 후, **계정 관리** + **사용자 관리** 의 코드 + UI + DB 책임이 4 경로로 분산된 상태가 남아 있었다:

1. **`/api/v1/accounts/*`** — Keycloak Admin REST 의 proxy (service account 가 manage-users role 사용)
2. **`/api/v1/users/*`** — 조직 메타데이터 (status / role / unit assignment) CRUD
3. **`/api/v1/organization/*`** — 조직 단위 (units / hierarchy)
4. **`/api/v1/rbac/*`** — RBAC policy + subject-role assignment

본 4 경로 의 책임이 **외부 Keycloak (사내 IdP 팀이 별도 운영)** 시나리오에서 모호하다.

### 2.2 Phase 1 — 현황 파악 매트릭스 (sprint `-a`, PR #199, 2026-05-20)

`docs/planning/account_user_management_redesign.md` 신규 (235 lines) — 책임 분산 매트릭스 13 row + 17 backend endpoint + 4 frontend page + DB schema 정합. **핵심 발견 5건**:

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

| sub-carve | 영역 | sprint |
| --- | --- | --- |
| **A** | ADR-0020 발급 + design doc §6 실행 계획 + `rbac_subject_roles` 완전 제거 (결정 D, §5.8 따름) | `-d` (본 sprint) |
| **B** | `/api/v1/accounts/*` 4 endpoint 제거 + `KeycloakAdminClient` write 메서드 호출처 제거 + `authenticateActor` lazy auto-create 실 구현 (결정 A + §5.2) + frontend `account.service.ts` 폐기 | `-e` (후속) |
| **C** | event listener 확장 — USER:UPDATE / GROUP_MEMBERSHIP / USER:DELETE 매핑 + DevHub `users` write + metric 3종 (결정 C + §5.3) | `-f` (후속) |
| **D** | JWKS stale-while-error expiry case 확장 (결정 F + §5.6) | `-g` (후속) |
| **E** | service account 권한 축소 + governance 협약 SOP (`keycloak_operations.md §8.5c` 신규, §5.5) | `-h` (후속) |
| **F** | `/login` page 정리 (결정 B + §5.7) | `-i` (후속, 우선순위 낮음) |

### 4.2 보안 영향

- service account 권한 축소 (`manage-users` 제거) → 최소 권한 원칙 정합. service account compromised 시 user lifecycle 조작 불가
- JWKS stale-while-error expiry case 확장 → Keycloak unreachable 시 uptime 90일 까지 (key rotation 주기). signature 검증 유지로 token forgery 위험 없음 — rotation 직후 backend cache flush SOP 필요 (별도 carve)
- lazy auto-create → token 검증 성공한 user 만 lazy create (Keycloak user lifecycle 정책이 1차 필터). enumeration 위협 없음

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

## 6. 미해결 / 후속 작업

### 6.1 후속 ADR 후보
- (없음 — 본 ADR 이 Phase 2 결정 6건 모두 cover)

### 6.2 carve out (Phase 3 sprint 영역)
- backend `/api/v1/accounts/*` 4 endpoint 제거 (§4.1 sub-carve B)
- `authenticateActor` lazy auto-create 실 구현 (§4.1 sub-carve B)
- event listener 확장 (§4.1 sub-carve C)
- JWKS stale-while-error expiry case 확장 (§4.1 sub-carve D)
- service account 권한 축소 + governance SOP (§4.1 sub-carve E)
- frontend `account.service.ts` 폐기 + admin/settings/users UI 정리 (§4.1 sub-carve B 통합)
- `/login` page 정리 (§4.1 sub-carve F)

### 6.3 사내 동반 carve
- HRDB ETL push 의 unit 매핑 정보 stage — sprint -p `hrdb_etl_sync.sh` 확장 (신규 user unit 자동 매핑)
- Keycloak admin 운영 SOP 협약 — IdP 팀 + DevHub 운영자 간 책임 분리 협약 문서화 (§3.2 표를 사내 정책 문서로 승격)
- JWKS rotation 직후 backend cache flush SOP — rotation 직후 backend 강제 재시작 또는 cache flush endpoint

## 7. 변경 이력

| 일자 | 변경 | sprint |
| --- | --- | --- |
| 2026-05-20 | draft + Accepted — Phase 2 명시 결정 6건 종합 (옵션 A 전면 폐기) + Phase 3 sub-carve 6건 분리 plan | `claude/work_260520-d` |
