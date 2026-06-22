---
title: e2e_migration
type: source
tags: [infrastructure, e2e_migration.md, project-devhub]
sources: [raw/projects/devhub/docs/infrastructure/keycloak-idp/e2e_migration.md]
git_commit: 046e0c81
git_branch: chore/260622-wiki-drift-cleanup-4
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
last_touched: 2026-06-22T06:22:35Z
mirror_dirty: |
related: [none]
status: draft
contradictions: [none]
---

# Design 검토 — Frontend e2e Kratos → Keycloak admin API migration

- 문서 목적: [ADR-0019](../adr/0019-keycloak-only-idp.md) Keycloak 단일화 후 frontend e2e (Playwright) 의 잔재 Kratos admin API 의존을 Keycloak admin API 로 전환하는 design.
- 범위: `frontend/tests/e2e/global-setup.ts` (170 line) + `fixtures.ts` (130 line) + `.env.example` L41-42 의 `KRATOS_ADMIN_URL` 잔재. e2e 시작 시 user seed + signup.spec cleanup 패턴.
- 대상 독자: Frontend QA, Backend, IdP 운영팀.
- 상태: draft
- 최종 수정일: 2026-05-20
- 결정 근거 sprint: `claude/work_260519-m`
- 관련 문서: [ADR-0019 Keycloak 단일화](../adr/0019-keycloak-only-idp.md), [keycloak_operations.md](../setup/keycloak_operations.md), [keycloak_offboarding_immediacy.md](./keycloak_offboarding_immediacy.md) (admin REST path pattern reference).

## 1. 컨텍스트

### 1.1 현재 잔재 (sprint -k discovery)

[ADR-0019](../adr/0019-keycloak-only-idp.md) Keycloak 단일화 후에도 frontend e2e 측면에 Kratos admin API 의존 잔재:

| 파일 | 잔재 사용 |
| --- | --- |
| `frontend/tests/e2e/global-setup.ts` | `KRATOS_ADMIN_URL` env (default `localhost:4434`) + `POST/PUT/GET /admin/identities` 호출 — alice/bob/charlie 3 user seed. DevHub users row 별도 `infra/idp/sql/002_seed_e2e_users.sql` (Kratos identity_id 기반 가정) |
| `frontend/tests/e2e/fixtures.ts` | `getKratosIdentityIdByEmail` + `deleteKratosIdentityByEmail` — signup.spec cleanup 등 |
| `frontend/.env.example` L41-42 | `KRATOS_ADMIN_URL=http://localhost:4434` — Playwright global-setup 용 |

### 1.2 전환 복잡성

단순 rename 불가능 — Kratos 와 Keycloak admin API 의 차이:

| 항목 | Kratos | Keycloak |
| --- | --- | --- |
| Admin endpoint base | `KRATOS_ADMIN_URL/admin/identities` | `DEVHUB_KEYCLOAK_ADMIN_URL/admin/realms/{realm}/users` |
| 인증 | unauthenticated (localhost bind) | `client_credentials` grant — service account access_token |
| User schema | `traits.{email,system_id,display_name}` + `metadata_public.user_id` + `credentials.password.config.password` | `username` + `email` + `firstName/lastName` + `attributes.{employee_id,user_id}` + `credentials[].{type:password, value}` |
| List query | `?page=N&per_page=250` | `?email={email}&first={N}&max={N}` |
| 상태 | `state: active/inactive` | `enabled: true/false` |
| Password set | identity payload 의 `credentials.password.config.password` | 별도 `PUT /users/{id}/reset-password` 호출 |
| ID 식별자 | `id` (UUID) | `id` (UUID) |
| 이메일 검색 | pagination scan 필요 | `?email=` query 직접 지원 |

## 2. 통합 옵션 비교 (3종)

| 옵션 | 변경 범위 | DevHub backend 영향 | e2e 동작 변경 | 권장 |
| --- | --- | --- | --- | --- |
| **A. 현행 잔재 그대로 — 사내 Kratos staging 의존** | 없음 | 없음 | 사내 e2e 환경에 Kratos 가동 시까지 동작 | △ ADR-0019 정합 안 됨 |
| **B. Keycloak admin API 전환 — Playwright global-setup 재작성** | global-setup.ts + fixtures.ts + .env.example + e2e 가이드. backend 의 `idp-apply-schemas + 002_seed_e2e_users.sql` 도 Keycloak sub 기반으로 정합 | 없음 (backend code 그대로 — admin client 는 이미 KC-PR-C 도입됨) | 사내 Keycloak e2e 환경 staging 동반 필요 | ⭐ **권장** (Phase 2 사내 환경 staging 진입 시) |
| **C. e2e 자체 폐기 — 사내 manual QA 로 전환** | Playwright suite 전체 제거 | 없음 | CI 의 e2e job 제거 | ❌ regression 가드 손실 |

## 3. 옵션 B (권장) — Keycloak admin API 전환 상세

### 3.1 admin token 획득 (인증 추가)

Keycloak admin API 는 `client_credentials` grant 의 access_token 필요:

```ts
// global-setup.ts 신규 헬퍼
async function fetchKeycloakAdminToken(): Promise<string> {
  const tokenURL = `${KC_ISSUER_URL}/protocol/openid-connect/token`;
  const resp = await fetch(tokenURL, {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: new URLSearchParams({
      grant_type: "client_credentials",
      client_id: KC_ADMIN_CLIENT_ID,        // 'devhub-backend' (KC-PR-C service account)
      client_secret: KC_ADMIN_CLIENT_SECRET,
    }),
  });
  if (!resp.ok) throw new Error(`Keycloak token ${resp.status}: ${await resp.text()}`);
  const json = await resp.json() as { access_token: string };
  return json.access_token;
}
```

env 변경:
- `KRATOS_ADMIN_URL` → 제거
- 신규 `KC_ISSUER_URL` (또는 `KEYCLOAK_BASE_URL` + realm 분리)
- 신규 `KC_ADMIN_CLIENT_ID` (default `devhub-backend`)
- 신규 `KC_ADMIN_CLIENT_SECRET` (사내 vault)
- 신규 `KC_ADMIN_REALM` (default `devhub`)

### 3.2 user seed 패턴 변경

```ts
// 1. Find or create
async function findKeycloakUserByEmail(token: string, email: string): Promise<KeycloakUser | null> {
  const url = `${KC_ADMIN_URL}/admin/realms/${KC_ADMIN_REALM}/users?email=${encodeURIComponent(email)}&exact=true`;
  const resp = await fetch(url, { headers: { Authorization: `Bearer ${token}`, Accept: "application/json" } });
  if (!resp.ok) throw new Error(`Keycloak list users ${resp.status}: ${await resp.text()}`);
  const users = await resp.json() as KeycloakUser[];
  return users[0] ?? null;
}

async function createKeycloakUser(token: string, seed: KratosSeed /* rename → KeycloakSeed */): Promise<string> {
  const payload = {
    username: seed.user_id,
    email: seed.email,
    firstName: seed.display_name.split(" ")[0],
    lastName: seed.display_name.split(" ").slice(1).join(" "),
    enabled: true,
    emailVerified: true,
    attributes: {
      employee_id: [seed.user_id],
      // (e2e seed 의 user_id 는 Keycloak sub 가 아니라 username — Keycloak 이 발급한 sub 가 users.idp_subject 와 매핑)
    },
    credentials: [{ type: "password", value: seed.password, temporary: false }],
  };
  const resp = await fetch(`${KC_ADMIN_URL}/admin/realms/${KC_ADMIN_REALM}/users`, {
    method: "POST",
    headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  if (resp.status !== 201) throw new Error(`Keycloak create user ${seed.user_id} → ${resp.status}: ${await resp.text()}`);
  // Keycloak 201 응답은 Location header 에 신규 user URL 포함 → ID 추출
  const location = resp.headers.get("location") ?? "";
  const id = location.split("/").pop();
  if (!id) throw new Error("Keycloak create user: missing Location header");
  return id;
}

// 2. Reset password (이미 존재하는 user 의 password 재설정 — Kratos PUT 패턴 정합)
async function resetKeycloakPassword(token: string, userId: string, password: string): Promise<void> {
  const url = `${KC_ADMIN_URL}/admin/realms/${KC_ADMIN_REALM}/users/${userId}/reset-password`;
  const resp = await fetch(url, {
    method: "PUT",
    headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
    body: JSON.stringify({ type: "password", value: password, temporary: false }),
  });
  if (!resp.ok) throw new Error(`Keycloak reset password ${userId} → ${resp.status}: ${await resp.text()}`);
}

// 3. group 가입 (keycloak_groups_rbac_mapping §3.2 정합)
async function assignKeycloakGroup(token: string, userId: string, groupName: string): Promise<void> {
  // group ID 조회
  const groupURL = `${KC_ADMIN_URL}/admin/realms/${KC_ADMIN_REALM}/groups?search=${encodeURIComponent(groupName)}&exact=true`;
  const groupResp = await fetch(groupURL, { headers: { Authorization: `Bearer ${token}` } });
  const groups = await groupResp.json() as Array<{ id: string }>;
  if (!groups[0]) throw new Error(`Keycloak group ${groupName} not found`);
  const groupId = groups[0].id;
  // user 를 group 에 가입
  const joinResp = await fetch(
    `${KC_ADMIN_URL}/admin/realms/${KC_ADMIN_REALM}/users/${userId}/groups/${groupId}`,
    { method: "PUT", headers: { Authorization: `Bearer ${token}` } },
  );
  if (!joinResp.ok) throw new Error(`Keycloak group join ${userId} → ${groupName}: ${joinResp.status}`);
}
```

### 3.3 DevHub users row seed 변경

`backend-core/cmd/idp-apply-schemas` + `infra/idp/sql/002_seed_e2e_users.sql` 의 패턴이 Kratos identity_id 기반 → Keycloak sub 기반으로 정합:

- 옵션 1: Keycloak admin API 로 user 생성 → 응답 ID (Keycloak sub) 추출 → DevHub `users.idp_subject = sub` 로 직접 INSERT/UPSERT
- 옵션 2: backend 의 `authenticateActor` 확장 (자동 첫 로그인 sync, sprint -j codex hotfix #2 의 backend 확장 carve) 후 — 첫 로그인 시 자동 sync
- 옵션 3: 기존 SQL seed 의 user_id 매핑을 Keycloak admin API 호출 결과로 정합 (idp-apply-schemas 확장)

권장 = **옵션 1** — Playwright global-setup 의 단순 chain (create user → reset password → join group → INSERT users row).

### 3.4 fixtures.ts helper 변경

```ts
export async function getKeycloakUserIdByEmail(email: string): Promise<string | null> {
  const token = await fetchKeycloakAdminToken();
  const user = await findKeycloakUserByEmail(token, email);
  return user?.id ?? null;
}

export async function deleteKeycloakUserByEmail(email: string): Promise<void> {
  const token = await fetchKeycloakAdminToken();
  const id = await getKeycloakUserIdByEmail(email);
  if (!id) return;
  const resp = await fetch(`${KC_ADMIN_URL}/admin/realms/${KC_ADMIN_REALM}/users/${id}`, {
    method: "DELETE",
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!resp.ok && resp.status !== 404) throw new Error(`Keycloak delete user ${email}: ${resp.status}`);
}
```

## 4. cutover 절차

### 4.1 Phase 1 (본 sprint) — design + deprecation 명시

- ✅ 본 design 문서
- ✅ `.env.example` L41-42 deprecation 주석 추가
- ✅ `global-setup.ts` + `fixtures.ts` 상단 deprecation note
- ✅ ADR-0019 §5.3 carve out 추가 — e2e Kratos → Keycloak migration

### 4.2 Phase 2 — staging Keycloak e2e 환경 진입 (별도 sprint)

- 사내 Keycloak realm `devhub` 가 staging 환경에 존재 (keycloak_operations.md §2 정합)
- group 4 + composite role 4 적용 (keycloak_groups_rbac_mapping §3.2 SOP)
- `devhub-backend` service account 의 realm-management view-users/manage-users role 권한 확보
- Playwright global-setup.ts + fixtures.ts 전환 (§3.1~§3.4)
- DevHub users seed SQL 정합 (§3.3 옵션 1)

### 4.3 Phase 3 — prod cutover

- staging 1주 검수 후 prod e2e CI 환경 전환
- KRATOS_ADMIN_URL env 완전 제거
- 사내 Kratos staging instance shutdown (사내 운영팀)

## 5. 보안 점검

### 5.1 잠재 위협 + mitigation

| 위협 | mitigation |
| --- | --- |
| `KC_ADMIN_CLIENT_SECRET` 노출 (Playwright global-setup 시) | 사내 vault + CI secret 주입. dev local 은 `.env.local` (git untracked) + secret rotation SOP (keycloak_operations §8.3) |
| e2e user seed 의 사내 prod 환경 누출 | e2e 환경 분리 (별도 realm 또는 별도 Keycloak instance). prod 환경 e2e 금지 |
| admin token 의 권한 과대 | service account 의 realm-management view-users/manage-users 만 — admin 전체 권한 회피 |
| user create 후 cleanup 누락 시 e2e 환경 누적 | global-setup 의 idempotent UPSERT 패턴 (find-or-create + reset-password) + signup.spec cleanup (deleteKeycloakUserByEmail) |

## 6. 잔여 carve out

- **(carve)** Phase 2 staging 진입 — 사내 Keycloak e2e 환경 동반
- **(carve)** Phase 3 prod cutover — 사내 Kratos staging shutdown 동반
- **(carve)** backend `authenticateActor` 확장 (자동 첫 로그인 HRDB sync) — sprint -j codex hotfix #2 의 backend 확장 carve 와 정합. 옵션 2 채택 시 Playwright global-setup 의 INSERT users row 단계 생략 가능.
- **(carve)** idp-apply-schemas + 002_seed_e2e_users.sql 의 Keycloak sub 기반 정합

## 7. ADR governance 결정

본 design 은 ADR-0019 §5.3 의 잔여 carve — e2e Kratos → Keycloak migration. Phase 2 staging 진입 시 별도 ADR 발행 가치 낮음 — design 완료 + Phase 2 가 실 코드 전환 PR 으로 처리.

## 8. 변경 이력

| 일자 | 변경 | sprint |
| --- | --- | --- |
| 2026-05-19 | 1차 draft — 8 section + 옵션 3종 비교 (현행 잔재 / Keycloak admin 전환 / e2e 폐기) + Phase 1 design + Phase 2 staging + Phase 3 prod + admin API mapping 표 (인증/path/payload/state/password/검색 6 항목) + 신규 user seed + 그룹 가입 + DevHub users 정합 3 옵션 + 보안 4 위협 + carve 4 항목. ADR-0019 §5.3 잔여 carve 추가. | `claude/work_260519-m` |
