# ADR-0014: Dev Request intake token 의 admin 발급/조회/revoke 정책

- 문서 목적: `dev_request_intake_tokens` resource 의 admin endpoint 정책 결정 — token 발급 (plain 1회 노출), 목록 조회 (hashed 미노출), revoke (idempotent). ADR-0012 (외부 수신 인증) 의 운영 측면 carve out.
- 범위: `POST /api/v1/dev-request-tokens`, `GET /api/v1/dev-request-tokens`, `DELETE /api/v1/dev-request-tokens/:token_id` 3 endpoint 의 RBAC + 응답 schema + audit 정책. middleware `requireIntakeToken` (token consumer side) 은 ADR-0012 가 다룬다.
- 대상 독자: Backend 개발자, frontend admin UI 담당자 (`/admin/settings/dev-request-tokens` 페이지 sprint p), RBAC 정책 stakeholder, 운영 admin.
- 상태: accepted
- 작성일: 2026-05-15
- 결정일: 2026-05-15 (sprint `claude/work_260515-o`)
- 결정 근거 sprint: `claude/work_260515-o` — DREQ-Admin-UI backend (carve 2/4 part 1).
- 관련 문서: [ADR-0012 DREQ 외부 수신 인증](./0012-dreq-external-intake-auth.md), [ADR-0013 DREQ RBAC row-scoping](./0013-dreq-rbac-row-scoping.md), [`docs/planning/development_request_concept.md`](../planning/development_request_concept.md) §10, [`docs/requirements.md`](../requirements.md) §5.5, [추적성 매트릭스 §3 Dev Request 행 + §4 ADR](../traceability/report.md).

## 1. 컨텍스트

ADR-0012 가 채택한 옵션 A (API 토큰 + IP allowlist) 는 `dev_request_intake_tokens` 테이블 + `requireIntakeToken` middleware 를 sprint `i` (PR #124) 에서 도입. 그러나 **token 의 발급 / 조회 / revoke admin endpoint** 는 carve out 으로 남아 있었다. 본 ADR 가 그 carve 를 결정한다.

token 의 운영 상 핵심 제약:
- **plain token 은 단 한번만 노출**. 발급 응답에 1회 포함 후 server-side 어디에도 보관하지 않는다 (DB 컬럼 0, audit 0, log 0).
- **server 는 SHA-256(plain) hex** 만 `dev_request_intake_tokens.hashed_token` 컬럼에 저장한다. ADR-0012 §4.1.1 의 인증 검증 흐름 (hash → lookup) 과 정합.
- 운영자가 잃어버린 token 은 복구 불가 — 새로 발급해야 한다 (`accounts_admin` password issuance 와 동일한 정책).

## 2. 결정 동인

- **노출 최소화**: plain 은 1회, hashed 는 절대 응답에 노출하지 않는다. operator 입장에서 hashed 의 가치는 0 이고, 노출 시 brute-force 표면이 커진다.
- **RBAC 단순**: system_admin 일임. 발급/조회/revoke 셋 모두 운영자 책임이며, 다른 role 에 위양할 필요가 없다. ADR-0013 의 row-level scope 와는 다른 결정 — token 자체는 row 단위로 위양되지 않는다.
- **accounts_admin 패턴 일관**: `POST /api/v1/accounts` 가 server-generated temp password 를 1회 응답에 노출하는 것과 동일한 패턴. 운영자 UX 일관.
- **deny-by-default**: `allowed_ips` 빈 배열은 admin 단계에서 거절 (`invalid_allowed_ips`). production 에서 모든 caller 를 차단하는 token 을 만들 의도는 없다고 가정 — 실수 방지.
- **idempotent revoke**: 같은 token 을 두 번 revoke 해도 `revoked_at` 이 변하지 않는다 (COALESCE 패턴). operator 가 retry 해도 정합.

## 3. 검토 옵션

본 ADR 는 ADR-0012 의 채택 (옵션 A) 위의 운영 carve 이므로 인증 방식 자체는 다시 평가하지 않는다. 본 ADR 는 다음 결정 항목만:

| 항목 | 옵션 | 채택 |
| --- | --- | --- |
| token 생성 주체 | client 가 plain 송부 vs server 생성 | server 생성 (32-byte base64url, 256-bit entropy) |
| plain 응답 노출 | 절대 안 함 vs 1회 노출 vs 매번 노출 | **1회 노출** (accounts_admin temp_password 와 정합) |
| hashed 응답 노출 | 항상 vs 발급 시만 vs 절대 | **절대 안 함** (operator 가치 0, brute-force 표면 ↓) |
| RBAC | 신규 resource `dev_request_intake_tokens` vs `security` resource 재사용 | **신규 resource** (의미 일관 + matrix 명확) |
| RBAC matrix | system_admin only vs 다른 role 위양 | **system_admin only** (운영 단일 책임) |
| allowed_ips 빈 배열 | 통과 (deny-all) vs 거절 | **거절** (`invalid_allowed_ips`, 실수 방지) |
| revoke 시 hard delete vs soft (revoked_at) | hard vs soft | **soft** — audit 보존 + 동일 token 재발급 시 hashed 충돌 검출 가능 |

## 4. 결정

### 4.1 RBAC 매트릭스 — `dev_request_intake_tokens` 신규 resource

| role | view | create | edit | delete |
| --- | --- | --- | --- | --- |
| `developer` | ❌ | ❌ | ❌ | ❌ |
| `manager` | ❌ | ❌ | ❌ | ❌ |
| `pmo_manager` | ❌ | ❌ | ❌ | ❌ |
| `system_admin` | ✅ | ✅ | ✅ | ✅ |

migration `000026_rbac_dev_request_intake_tokens` (sprint `claude/work_260515-o`) 가 4 role 의 `rbac_policies.permissions` JSONB 에 위 셀을 seed. `edit` 은 본 ADR 의 4 endpoint 범위에 없지만 matrix 자축 통일성을 위해 등재 (`dev_requests` resource 와 동일 패턴).

### 4.2 endpoint 3종

#### POST `/api/v1/dev-request-tokens` — 발급

**요청 body**:
```json
{
  "client_label": "ops_portal",
  "source_system": "ops",
  "allowed_ips": ["10.0.0.0/24", "192.0.2.7"]
}
```

- `client_label`: 운영 식별자 (UNIQUE 제약 없음 — 같은 source 의 다중 client 가능). 필수.
- `source_system`: intake 요청의 `source_system` 자동 채움 값 (ADR-0012 §4.1.2 spoofing 방지의 anchor). 필수.
- `allowed_ips`: CIDR 또는 단일 IP. 빈 배열 거절. 중복 자동 dedupe (순서 보존).

**처리**:
1. server 가 32-byte `crypto/rand` → `base64.RawURLEncoding` → plain token 생성 (43 chars, 256-bit entropy).
2. SHA-256(plain) hex 계산 → `hashed_token`.
3. `actor.login` 을 `created_by` 로 매핑 (없으면 `"system"`).
4. INSERT. UNIQUE (hashed_token) 위반 시 `intake_token_collision` 409 (사실상 발생 불가).
5. audit `dev_request_intake_token.issued` + payload `{client_label, source_system, allowed_ips}`. **plain/hashed 둘 다 audit 미포함**.

**응답 — 201**:
```json
{
  "status": "ok",
  "data": {
    "token_id": "uuid",
    "client_label": "ops_portal",
    "source_system": "ops",
    "allowed_ips": ["10.0.0.0/24", "192.0.2.7"],
    "created_at": "...",
    "created_by": "alice",
    "last_used_at": null,
    "revoked_at": null,
    "plain_token": "<43-char base64url>"
  }
}
```
`plain_token` 은 본 응답에서만 노출. operator 가 즉시 안전한 저장소로 옮기는 책임.

#### GET `/api/v1/dev-request-tokens` — 목록

응답: 위와 동일 schema, 단 `plain_token` 키 부재. **`hashed_token` 도 노출하지 않는다**. revoked 행 포함, `created_at DESC` 정렬.

#### DELETE `/api/v1/dev-request-tokens/:token_id` — revoke

`revoked_at = COALESCE(revoked_at, NOW())` UPDATE — 이미 revoked 면 변경 없음 (idempotent). 200 응답에 row 자체 (`plain_token` 없음). audit `dev_request_intake_token.revoked` + payload `{client_label, source_system}`.

### 4.3 error 카탈로그

| code | HTTP | 의미 |
| --- | --- | --- |
| `invalid_allowed_ips` | 400 | 빈 배열 또는 잘못된 CIDR/IP |
| `intake_token_collision` | 409 | hashed_token UNIQUE 위반 (사실상 발생 불가) |
| `auth_role_denied` | 403 | system_admin 아닌 role 의 호출 (route gate) |

기존 일반 코드 (`unauthorized` 등) 는 인증 layer 가 발급.

### 4.4 audit 정합

| action | target_type | payload 핵심 |
| --- | --- | --- |
| `dev_request_intake_token.issued` | `dev_request_intake_token` | `client_label`, `source_system`, `allowed_ips` |
| `dev_request_intake_token.revoked` | `dev_request_intake_token` | `client_label`, `source_system` |

**plain / hashed token 자체는 audit 에 절대 기록하지 않는다** — log 누출 시 인증 우회 가능성 차단.

## 5. 결과

- `backend-core/internal/httpapi/dev_request_intake_tokens_admin.go` 신규 — 3 handler.
- `backend-core/internal/store/dev_request_intake_tokens.go` 의 `IntakeTokenStore` 를 admin CRUD 메서드 (Create/List/Revoke) 까지 확장.
- `backend-core/internal/domain/rbac.go` 에 `ResourceDevRequestIntakeTokens` 추가 + `AllResources()` / `DefaultPermissionMatrix()` 4 role 모두 갱신.
- `backend-core/internal/httpapi/permissions.go` 의 `routePermissionTable` 에 3 endpoint 추가.
- `backend-core/internal/httpapi/router.go` 에 endpoint 등록.
- migration `000026_rbac_dev_request_intake_tokens` (up/down).
- 신규 unit test 8건 (handler happy / validation 4 / list excludes hashed / revoke happy + idempotent + not_found / RBAC route policy gate).

## 6. 후속 작업

- **(sprint `p`)** frontend `/admin/settings/dev-request-tokens` 페이지 — `accounts_admin` `/admin/settings/users` 의 plain-1회-노출 modal 패턴 재사용. carve 2/4 part 2.
- **(carve)** token rotation policy — 만료 자동화 (`expires_at` 컬럼 + cron revoke). 운영 빈도 확인 후 결정.
- **(carve)** allowed_ips 의 mutation endpoint — 현재는 발급 후 변경 불가, revoke + 재발급으로 우회. 필요 시 별도 ADR.

## 7. 변경 이력

| 일자 | 변경 | 메모 |
| --- | --- | --- |
| 2026-05-15 | accepted — `dev_request_intake_tokens` resource 신규 도입, system_admin 일임. plain 1회 노출 + hashed 절대 미노출 + idempotent revoke. accounts_admin password issuance 패턴과 정합. | sprint `claude/work_260515-o` (PR TBD) |
