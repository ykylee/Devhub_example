# Sprint plan: API key management (multi-key + admin UI)

- **문서 목적**: 1차 sprint `feat/work_260609-a-swagger-apikey-expand` 의 follow-up 으로 multi-key 관리 (DB hashed key) + 관리자 frontend UI 를 구현하기 위한 sprint plan.
- **범위**: backend `api_keys` table + CRUD endpoints + auth middleware DB lookup 확장 + frontend `/admin/api-keys` page (list / create / revoke).
- **대상 독자**: backend / frontend / AI agent / QA / 운영자.
- **상태**: planned (draft, 2026-06-09)
- **결정 근거 sprint**: 1차 PR `feat/work_260609-a-swagger-apikey-expand` 의 사용자 follow-up 결정 (multi-key + frontend admin UI + DB hashed storage).
- **관련 문서**: [ADR-0029 §6 (f) multi-key 관리](../adr/0029-api-key-auth-and-swagger-scope.md), [ADR-0029 §5.2 rotation 정책](../adr/0029-api-key-auth-and-swagger-scope.md), [docs/domain/auth-session/api.md](../domain/auth-session/api.md), [docs/domain/auth-session/requirements.md](../domain/auth-session/requirements.md), [release_v0-1_roadmap §3 N-12](../planning/release_v0-1_roadmap.md).

## 0. 배경

1차 sprint (`feat/work_260609-a-swagger-apikey-expand`) 에서 single static `DEVHUB_API_KEY` env 기반 API key 인증 미들웨어를 추가. 본 follow-up 에서는 multi-key (운영자가 키를 발급/회수/만료 관리) + frontend admin UI (관리자 페이지) + DB hashed key storage 로 확장한다.

## 1. 후보 옵션

### 1.1 storage 형식 (3종)

| # | 옵션 | 장점 | 단점 | 결정 |
| --- | --- | --- | --- | --- |
| 1 | **DB (sha256 hash + key prefix)** | 회수/만료/audit 가능, last_used_at 추적, CIDR allowlist | DB 의존성, 추가 table | ⭐ **채택** (사용자 결정) |
| 2 | env 다중 (KEY_1, KEY_2, ...) | DB 의존성 0, rotation 쉬움 | 회수 불가, last_used_at 추적 불가 | ❌ |
| 3 | Vault / Secret Manager | enterprise | 구현 부담, 외부 의존성 | ❌ |

### 1.2 frontend UI 위치

| # | 옵션 | 결정 |
| --- | --- | --- |
| 1 | **기존 `/admin` 메뉴에 "API Keys" 항목 추가** | ⭐ **채택** (사용자 의도 = "관리자 메뉴") |
| 2 | 별도 `/admin/api-keys` route (메뉴 등록) | ⭐ 채택 (메뉴 + route) |
| 3 | system settings 내부 sub-tab | ❌ (탐색 깊이) |

## 2. 결정

**옵션 1 (DB hashed) + 옵션 2 (메뉴 + route)**.

## 3. Phase 1 — Backend (P0)

### 3.1 migration 000042 — `api_keys` table

```sql
CREATE TABLE public.api_keys (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,
    key_prefix text NOT NULL,                  -- 앞 8자, 표시/검색용
    key_hash bytea NOT NULL,                   -- sha256(raw_key)
    created_by text NOT NULL,                  -- actor login
    created_at timestamptz NOT NULL DEFAULT now(),
    last_used_at timestamptz,
    revoked_at timestamptz,
    revoked_by text,
    expires_at timestamptz,
    allowed_cidrs text[]                       -- CIDR allowlist (nullable)
);
CREATE UNIQUE INDEX api_keys_key_hash_active_uniq
    ON public.api_keys (key_hash)
    WHERE revoked_at IS NULL;
CREATE INDEX api_keys_key_prefix_idx ON public.api_keys (key_prefix);
CREATE INDEX api_keys_created_by_idx ON public.api_keys (created_by);
```

### 3.2 domain layer (`internal/domain/auth-session/`)

- `schema/api_key.go`: 도메인 model + enum (status: active/revoked/expired)
- `repository/api_key.go`: `CreateAPIKey`, `ListAPIKeys`, `GetAPIKeyByHash`, `RevokeAPIKey`, `UpdateLastUsedAt`, `UpdateAPIKeyMeta`
- `service/api_key.go`: `GenerateAPIKey()` — 32 byte crypto/rand + base64url = `dhk_<43 chars>` 평문. sha256 해시 후 `key_hash` + `key_prefix` (앞 8자) 저장.

### 3.3 view layer (`internal/domain/auth-session/view/api_key.go`)

| endpoint | method | auth | 권한 |
| --- | --- | --- | --- |
| `/api/v0-1/admin/api-keys` | POST | bearerAuth | system_admin (resource: `api_keys:create`) |
| `/api/v0-1/admin/api-keys` | GET | bearerAuth | system_admin (resource: `api_keys:view`) |
| `/api/v0-1/admin/api-keys/:id` | DELETE | bearerAuth | system_admin (resource: `api_keys:delete`) |
| `/api/v0-1/admin/api-keys/:id` | PATCH | bearerAuth | system_admin (resource: `api_keys:edit`) |

응답: 생성 시 평문 key **1회만** 응답 (`key` field). 이후 GET 에서는 `key_prefix` 만 노출 (보안).

### 3.4 auth middleware 확장 (`internal/domain/auth-session/view/auth.go`)

`AuthenticateActor` 의 API key 분기 확장:
1. `cfg.APIKey` (single env) 비교 — 기존
2. `cfg.APIKeyStore != nil` 시 DB `GetAPIKeyByHash(sha256(token))` lookup — 신규
3. 조회된 row 가:
   - `revoked_at != null` → 401
   - `expires_at < now()` → 401
   - `allowed_cidrs != null` 시 ClientIP 가 CIDR match 안 함 → 401
   - 모두 통과 시 통과 + `devhub_actor_login="api-key:<key_prefix>"` + best-effort `UpdateLastUsedAt`

### 3.5 RBAC (`internal/domain/rbac-permissions/view/permissions.go`)

`routePermissionTable` 에 4 row 추가:
- `{POST, /api/v0-1/admin/api-keys}` → `Resource: api_keys, Action: create`
- `{GET, /api/v0-1/admin/api-keys}` → `Resource: api_keys, Action: view`
- `{DELETE, /api/v0-1/admin/api-keys/:id}` → `Resource: api_keys, Action: delete`
- `{PATCH, /api/v0-1/admin/api-keys/:id}` → `Resource: api_keys, Action: edit`

`DefaultPermissionMatrix` 에 `system_admin` 만 4 axis 모두 true.

### 3.6 OpenAPI 보강

- 4 신규 schema: `ApiKeyCreateRequest`, `ApiKeyCreateResponse` (key 포함), `ApiKeySummary`, `ApiKeyUpdateRequest`
- 4 신규 path 추가
- securityScheme: `bearerAuth` 만 (admin 전용, staticTokenAuth 불필요)

### 3.7 Tests

- 단위: domain layer (GenerateAPIKey 32 byte, hash + prefix 정확), repository CRUD, view handler (POST/GET/DELETE/PATCH, 권한, 1회 key 반환)
- 통합: backend `go test ./internal/domain/auth-session/...` + `./internal/httpapi/...`
- E2E: admin 로그인 → POST 발급 → 응답의 key 복사 → API key 인증 → last_used_at 갱신 검증 → DELETE 회수 → 인증 401

### 3.8 Audit (`audit.api_key.*`)

- `audit.api_key.created` (action: `created`, target_type: `api_key`)
- `audit.api_key.revoked`
- `audit.api_key.used` (high volume — sampling 또는 summary counter 사용 검토)

## 4. Phase 2 — Frontend (P0)

### 4.1 route + menu

- `frontend/app/admin/api-keys/page.tsx` 신규
- `frontend/shared/layout/Sidebar.tsx` 의 "Admin" 메뉴 group 에 "API Keys" 항목 추가 (`/admin/api-keys` link, `system_admin` role gate)

### 4.2 component (`frontend/domain/auth-session/view/`)

- `ApiKeyListPage.tsx`: DataTable (name, prefix, created_by, created_at, last_used_at, status)
- `ApiKeyCreateDialog.tsx`: name + expires_at (optional) + allowed_cidrs (optional) 입력 → 생성 후 평문 key 1회 표시 + clipboard copy + "이 창을 닫으면 다시 볼 수 없음" 경고
- `ApiKeyRevokeButton.tsx`: 확인 dialog 후 DELETE 호출
- `ApiKeyEditDialog.tsx`: expires_at / allowed_cidrs 갱신 (PATCH)

### 4.3 service (`frontend/domain/auth-session/service/api_key.service.ts`)

- `listApiKeys()`: GET /api/v0-1/admin/api-keys
- `createApiKey({ name, expires_at?, allowed_cidrs? })`: POST /api/v0-1/admin/api-keys → `ApiKeyCreateResponse` (key 포함)
- `revokeApiKey(id)`: DELETE /api/v0-1/admin/api-keys/:id
- `updateApiKey(id, { expires_at?, allowed_cidrs? })`: PATCH

### 4.4 schema (`frontend/domain/auth-session/schema/api_key.schema.ts`)

- `ApiKeySummary`, `ApiKeyCreateRequest`, `ApiKeyCreateResponse`, `ApiKeyUpdateRequest`
- Zod validation

### 4.5 Permission gate

- `useCurrentUser()` hook 의 `role === 'system_admin'` 체크
- 미충족 시 `<NotAuthorized />` component 표시

### 4.6 E2E (Playwright)

- `frontend/tests/e2e/admin-api-keys.spec.ts` 신규
- TC-ADMIN-API-KEYS-01 (목록 빈 상태 → 생성 → 1 row)
- TC-ADMIN-API-KEYS-02 (생성 후 평문 key 표시 + clipboard)
- TC-ADMIN-API-KEYS-03 (회수 → 401 응답)
- TC-ADMIN-API-KEYS-04 (만료 시각 미래 → 인증 통과, 과거 → 401)
- TC-ADMIN-API-KEYS-05 (non-admin 접근 → NotAuthorized)

## 5. Phase 3 — 운영 SOP + audit 강화 (P1)

- 운영 SOP: rotation 정책 (90일, 1-active-key + 1-grace-period pattern)
- audit `auth.api_key.used` high-volume 대응: sampling 또는 5분 단위 aggregation counter (`auth_api_key_usage_count_by_key_id_5m`)
- Prometheus metric: `devhub_api_key_active_count` (Gauge), `devhub_api_key_used_total{key_prefix}` (Counter)

## 6. carry-over (ADR-0029 §6 + 본 sprint 잔여)

- (a) enforceRoutePermission 의 `auth_source != "api_key"` 가드 — 본 sprint 와 함께 처리 (Phase 1 §3.5 의 `api_keys:*` resource 추가로 우회)
- (b) rotation 정책 SOP — Phase 3 §5
- (c) openapi P2/P3 endpoint 30+ 확장 — 별도 sprint
- (d) CI lint gate — 별도 sprint
- (e) swagger-ui system_admin 가드 — 별도 sprint
- (g) API key 사용 audit 강화 — Phase 3 §5

## 7. 작업 분해 + sprint 분할

본 sprint 는 단일 PR 로 너무 큼. 다음 2 sprint 로 분할 권장:

### Sprint 1: `feat/work_260609-b-api-key-management-backend` (Phase 1, 7~10 days)
- migration 000042
- domain layer (schema + repository + service)
- view layer (4 endpoint handler)
- auth middleware 확장 (cfg.APIKeyStore)
- RBAC 4 row 추가
- OpenAPI 4 path + 4 schema
- 단위/통합/E2E 테스트
- ADR-0029 amendment (storage 형식 결정 + auth middleware 확장)
- traceability: `IMPL-auth-05` + `IMPL-auth-key-store-01` row

### Sprint 2: `feat/work_260609-b-api-key-management-frontend` (Phase 2, 4~5 days, Sprint 1 머지 후)
- route + sidebar menu
- 4 component
- service + schema
- permission gate
- E2E 5 TC
- ADR-0029 frontend 섹션 (선택)

## 8. 성공 기준

- Backend: 운영자가 admin 로그인 → POST /admin/api-keys 호출 → 평문 key 수신 (1회) → 사용자가 Authorization: Bearer <key> 로 공개 API 호출 → 200 + `last_used_at` 갱신 → DELETE 회수 → 401.
- Frontend: `/admin/api-keys` 페이지에서 목록/생성/회수 UI 가 정상 동작, system_admin 만 접근, non-admin 은 NotAuthorized.
- 회귀 0, audit emit, RBAC 정확.

## 9. 잔여 / 모르는 점

- **key 1회 표시 UX**: 1차 생성 후 dialog 닫으면 key 다시 못 봄. backend 가 hash 만 보관하므로 의도된 동작. UI warning 명시.
- **api_key 사용량 high-volume audit**: `auth.api_key.used` audit row 가 모든 호출마다 emit 되면 audit_logs 폭증. sampling 또는 aggregation counter 검토. 본 sprint Phase 3 에서 결정.
- **last_used_at 갱신 동시성**: 매 호출마다 UPDATE 가 DB load 증가. 주기적 batch 갱신 (1분) 또는 async best-effort 검토. 본 sprint Phase 1 에서 best-effort.
- **allowed_cidrs 검증 위치**: auth middleware (인프라) vs handler (application) — middleware 가 적절 (auth 의 일부).
