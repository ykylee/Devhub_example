# ADR-0017: Dev Request intake token 운영 hardening (expires_at + PATCH allowed_ips)

- 문서 목적: [ADR-0014](./0014-dreq-intake-token-admin.md) §6 의 2 carve out 항목 (token rotation `expires_at` + `allowed_ips` mutation endpoint) 활성화 결정. PR #137 (`gemini/dreq_e2e_260515`, sha `72bf265`) 으로 backend 가 1차 활성화한 사실의 사후 명문화.
- 범위: `dev_request_intake_tokens` 테이블의 `expires_at` / `updated_at` 컬럼 추가 (migration 000027), `requireIntakeToken` middleware 의 만료 체크 + `auth_intake_token_expired` 에러 코드, 신규 endpoint `PATCH /api/v1/dev-request-tokens/:token_id` (API-79). RBAC / plain 노출 정책 (system_admin, plain 1회) 은 ADR-0014 가 source-of-truth.
- 대상 독자: Backend 개발자, 운영자, 외부 system (HomeLab 외 의뢰 caller — 만료된 token 의 에러 처리).
- 상태: accepted
- 작성일: 2026-05-18
- 결정일: 2026-05-16 (PR #137 활성화), 2026-05-18 사후 명문화 (sprint `claude/work_260518-c`)
- 결정 근거 sprint: `gemini/dreq_e2e_260515` (PR #137, sha `72bf265`).
- 관련 문서: [ADR-0012 DREQ 외부 수신 인증](./0012-dreq-external-intake-auth.md), [ADR-0014 DREQ intake token admin](./0014-dreq-intake-token-admin.md), [`docs/backend_api_contract.md`](../backend_api_contract.md) §14.10 (API-79), [추적성 매트릭스 §3 Dev Request 행 + §4 ADR](../traceability/report.md).

## 1. 컨텍스트

ADR-0014 (sprint `claude/work_260515-o`, PR #130) 는 `dev_request_intake_tokens` resource 의 admin endpoint 3종 (POST/GET/DELETE) 을 결정하면서 §6 에 두 가지 항목을 명시적으로 carve out:

> **(carve)** token rotation policy — 만료 자동화 (`expires_at` 컬럼 + cron revoke). 운영 빈도 확인 후 결정.
> **(carve)** allowed_ips 의 mutation endpoint — 현재는 발급 후 변경 불가, revoke + 재발급으로 우회. **필요 시 별도 ADR**.

PR #137 이 두 항목 중 일부를 활성화했다:

1. `expires_at` 컬럼 추가 + middleware 만료 체크 — **활성화**. 단 자동 cron revoke 는 carve out 유지.
2. `PATCH /api/v1/dev-request-tokens/:token_id` 신규 endpoint — **활성화**. revoke + 재발급 우회 가능.

본 ADR 은 두 결정을 사후 명문화. ADR-0014 §6 의 "필요 시 별도 ADR" 명시 정합.

## 2. 결정 동인

### 2.1 expires_at 도입

- **토큰 수명 명시화**: 운영자가 token 발급 시 만료 시각을 결정해 두면, 잊혀진 token 이 영구히 작동하는 시나리오 차단. `accounts_admin` 의 `temp_password_expires_at` 패턴과 정합 (backend_api_contract §10.2).
- **자동화 vs 수동의 단계적 도입**: 본 결정은 컬럼 + middleware 체크만 활성화. 자동 cron revoke 는 운영 빈도 확인 후 별도 결정 (carve out 유지).
- **error code 분리**: 만료된 token 은 `auth_invalid_token` 과 의미가 다름. caller 가 "재발급 필요" vs "권한 부재" 구분 가능해야 함.

### 2.2 PATCH allowed_ips 도입

- **revoke+재발급 우회**: ADR-0014 §6 carve out 시점 ("revoke + 재발급으로 우회") 의 가정은 IP 변경 빈도가 낮을 것이라는 추정. 실제 운영에서는 source system 의 IP 가 자주 바뀌고, 매번 plain token 재배포는 운영 부담.
- **immutable 정책 vs mutation 정책 트레이드오프**: token 자체 (hashed_token) 는 immutable 유지. `allowed_ips` 만 변경 허용. `client_label`, `source_system` 은 변경 불가 (audit anchor 유지).
- **plain token 무영향**: PATCH 는 plain token 노출 없음 (이미 발급 시 1회 노출 완료). caller 의 인증 자체에는 영향 없음.

## 3. 검토 옵션

### 3.1 expires_at 정책

| 항목 | 옵션 | 채택 |
| --- | --- | --- |
| 컬럼 nullable | NOT NULL (모든 token 만료 필수) vs **NULL 허용** (만료 없음 가능) | **NULL 허용** (기존 token 호환 + 영구 token 옵션 보존) |
| 만료 체크 위치 | client 호출 시 (lazy) vs scheduled cron (eager) | **client lazy** (현재) + cron carve out |
| 만료 에러 코드 | `auth_invalid_token` 재사용 vs **`auth_intake_token_expired` 신규** | **신규** (caller UX 분리) |
| 만료 후 동작 | 401 reject + audit `dev_request_intake.expired` vs 401 reject only | **401 reject + audit** (운영 visibility) |

### 3.2 PATCH allowed_ips 정책

| 항목 | 옵션 | 채택 |
| --- | --- | --- |
| 변경 가능 필드 | hashed_token + allowed_ips + label vs **allowed_ips only** | **allowed_ips only** (audit anchor 유지) |
| 변경 방식 | merge (add/remove) vs **replace (전체 갱신)** | **replace** (UX 단순, allowed_ips dedupe 자연) |
| 빈 배열 처리 | 허용 (deny-all) vs **거절** (`invalid_allowed_ips`) | **거절** (ADR-0014 §4.2 와 동일 정책 일관) |
| audit action | `dev_request_intake_token.updated` 신규 vs `revoked + reissued` 시뮬레이션 | **`updated` 신규** (의미 명확) |

## 4. 결정

### 4.1 expires_at 컬럼 + 만료 체크

**migration 000027** (`add_expires_at_to_intake_tokens.up.sql`):

```sql
ALTER TABLE dev_request_intake_tokens ADD COLUMN expires_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE dev_request_intake_tokens ADD COLUMN updated_at TIMESTAMP WITH TIME ZONE;
```

- `expires_at` NULL = 영구 토큰 (기존 호환).
- `updated_at` 은 PATCH 시 갱신.

**middleware 만료 체크** (`requireIntakeToken` in `dev_request_intake_auth.go`):

```
if expires_at IS NOT NULL AND expires_at <= NOW():
    audit dev_request_intake.expired
    401 { code: "auth_intake_token_expired", ... }
```

**error code** — `auth_intake_token_expired` (401). caller 는 운영자에게 새 token 발급 요청.

**POST `/api/v1/dev-request-tokens` 요청 body 확장**:

```json
{
  "client_label": "ops_portal",
  "source_system": "ops",
  "allowed_ips": ["..."],
  "expires_at": "2026-12-31T23:59:59Z"   // optional, RFC3339
}
```

response 의 `expires_at` 필드 추가.

### 4.2 PATCH endpoint — API-79

```
PATCH /api/v1/dev-request-tokens/:token_id
Authorization: Bearer <admin OIDC>
RBAC: dev_request_intake_tokens:edit (system_admin)
```

**요청 body**:
```json
{
  "allowed_ips": ["10.0.0.0/24", "192.0.2.10"]
}
```

- `allowed_ips` 필수. 빈 배열 거절 (`invalid_allowed_ips` 400).
- CIDR / IP 검증은 POST 와 동일 (ADR-0014 §4.2 의 `invalid_allowed_ips` 정책 재사용).
- 다른 필드 (`client_label`, `source_system`, `hashed_token`) 는 변경 거부.

**처리**:
1. token_id 로 row 조회. 미존재 → 404 `intake_token_not_found`.
2. 이미 revoked → 409 `intake_token_revoked` (변경 의미 없음).
3. UPDATE `allowed_ips`, `updated_at = NOW()`.
4. audit `dev_request_intake_token.updated` + payload `{client_label, source_system, allowed_ips_before, allowed_ips_after}`.

**응답 — 200**: revoke 응답과 동일 schema (`plain_token` 없음).

### 4.3 error 카탈로그 확장

ADR-0014 §4.3 의 카탈로그에 다음 추가:

| code | HTTP | 의미 |
| --- | --- | --- |
| `auth_intake_token_expired` | 401 | `requireIntakeToken` 만료 체크 실패 (`expires_at <= NOW()`) |
| `intake_token_not_found` | 404 | PATCH/DELETE 대상 token_id 미존재 |
| `intake_token_revoked` | 409 | 이미 revoke 된 token 의 PATCH 시도 |

### 4.4 audit 정합

ADR-0014 §4.4 의 audit action 에 추가:

| action | target_type | payload 핵심 |
| --- | --- | --- |
| `dev_request_intake_token.updated` | `dev_request_intake_token` | `client_label`, `source_system`, `allowed_ips_before`, `allowed_ips_after` |
| `dev_request_intake.expired` | `dev_request_intake_token` | `client_label`, `source_system`, `expires_at` |

ADR-0014 §4.4 의 정책 ("plain / hashed token 자체는 audit 에 절대 기록하지 않는다") 동일 유지.

## 5. 결과

- migration `000027_add_expires_at_to_intake_tokens` (up/down) — `expires_at` + `updated_at` 컬럼.
- `backend-core/internal/store/dev_request_intake_tokens.go` — `Create/List/Revoke` 시 `expires_at` 매핑 + `UpdateAllowedIPs` 메서드 신규.
- `backend-core/internal/httpapi/dev_request_intake_auth.go` — middleware 만료 체크 + `auth_intake_token_expired` 분기 + audit `dev_request_intake.expired`.
- `backend-core/internal/httpapi/dev_request_intake_tokens_admin.go` — PATCH handler 신규 (`API-79`).
- `backend-core/internal/httpapi/permissions.go` — PATCH route 의 `dev_request_intake_tokens:edit` policy 추가.
- `backend-core/internal/httpapi/router.go` — PATCH endpoint 등록.
- `docs/backend_api_contract.md` §14.10 — API-66/67/68/79 활성화 + token expires_at 명세.
- unit test — `UT-dreq-token-expiry-01` (middleware 만료 체크 4 case: 정상 / 막 만료 / 1초 전 만료 / NULL), `UT-dreq-allowed-ips-mutation-01` (PATCH 4 case: happy / not_found / revoked / invalid_allowed_ips).

## 6. 후속 작업

- **(carve, ADR-0014 §6 유지)** 자동 cron revoke — `expires_at <= NOW()` 인 token 을 주기적으로 hard-revoke 처리. 현재는 middleware lazy 체크만. 운영 빈도 확인 후 별도 결정.
- **(carve)** PATCH 의 `expires_at` 갱신 — 본 ADR 은 `allowed_ips` 만 변경 허용. `expires_at` 갱신은 revoke + 재발급 우회 (1차 정책).
- **(carve)** 토큰 만료 알림 — `expires_at` 임박 시 운영자에게 알림 (Prometheus metric `devhub_intake_token_expiring_soon` 또는 audit + 대시보드).
- **(carve)** `last_used_at` 기반 staleness alert — `last_used_at` 컬럼은 ADR-0014 §4.2 응답 schema 에 이미 존재. N일 미사용 token 자동 정리는 별도 ADR.
- **(carve)** `UpdateDevRequestIntakeTokenIPs` 의 atomicity 강화 — PR #146 (codex hotfix #5) 가 추가한 revoked guard 는 2 query 패턴: (1) `UPDATE ... WHERE revoked_at IS NULL RETURNING ...`, (2) ErrNoRows 시 별도 `SELECT revoked_at WHERE token_id` 로 not_found vs revoked 분기. 두 query 가 별도 pgxpool connection 에서 실행될 수 있어 between-query race window 존재 — UPDATE 가 ErrNoRows 후 다른 transaction 이 row 를 삭제/revoke 한 경우 SELECT 결과가 사실과 다른 분기 (ErrNotFound vs ErrConflict) 를 낼 수 있음. 실제 영향은 의미상 둘 다 mutation 차단으로 같지만 (false positive 가 보안 영향 없음), 정확한 status code 가 client/runbook 의 분기 로직에 영향 줄 수 있음. 권장 해소: 단일 query CTE 또는 `tx.Begin()` + `SELECT ... FOR UPDATE` + UPDATE 로 atomic 보장. PR #146 self-review P2.

```sql
-- 단일 query CTE 예시 (carve out 의 reference)
WITH locked AS (
  SELECT token_id, revoked_at FROM dev_request_intake_tokens
  WHERE token_id = $1::uuid FOR UPDATE
),
upd AS (
  UPDATE dev_request_intake_tokens
  SET allowed_ips = $2::jsonb, updated_at = NOW()
  WHERE token_id = $1::uuid AND revoked_at IS NULL
  RETURNING token_id::text, client_label, hashed_token, allowed_ips, source_system,
            created_at, created_by, last_used_at, revoked_at, expires_at
)
-- 한 행 anchor (root) 위에서 두 CTE 를 LEFT JOIN — locked 가 empty (token 미존재)
-- 이라도 row 1개는 반드시 반환되어 `lock_token_id IS NULL → 404` 분기 가능.
-- codex hotfix #6 P2 (PR #147): 원래 `FROM locked LEFT JOIN upd` 패턴은 locked
-- 가 empty 일 때 zero rows 를 내서 not-found 매핑이 불가했음.
SELECT locked.token_id  AS lock_token_id,
       locked.revoked_at AS lock_revoked_at,
       upd.token_id     AS upd_token_id,
       upd.client_label,
       upd.allowed_ips,
       upd.updated_at
FROM (VALUES (1)) AS root(_)
LEFT JOIN locked ON true
LEFT JOIN upd    ON true;
```

handler 는 단일 row 결과를 보고 `lock_token_id IS NULL` → 404, `lock_revoked_at IS NOT NULL` → 409, `upd_token_id IS NOT NULL` → 200 으로 분기. 트랜잭션 격리 (READ COMMITTED 기본) + `FOR UPDATE` row lock 으로 atomic 보장.

## 7. 변경 이력

| 일자 | 변경 | 메모 |
| --- | --- | --- |
| 2026-05-16 | PR #137 활성화 — migration 000027 + middleware 만료 체크 + PATCH endpoint (API-79). | sprint `gemini/dreq_e2e_260515` |
| 2026-05-18 | accepted — ADR 형식으로 사후 명문화. expires_at NULL 허용 (기존 호환), 만료 체크는 lazy (cron carve out), PATCH 는 allowed_ips only (audit anchor 보존), 에러 코드 3종 + audit action 2종 추가. ADR-0014 §6 의 두 carve out 중 본 ADR 가 활성화한 부분 명시. | sprint `claude/work_260518-c` (PR #143) |
| 2026-05-18 | §6 atomicity carve out 추가 — PR #146 (codex hotfix #5) 가 store 에 revoked guard 를 추가했지만 2 query 패턴이라 between-query race window 존재. PR #146 self-review P2 가 발견. CTE 단일 query + `FOR UPDATE` row lock 으로 atomic 보장하는 reference SQL 본문 추가. 실제 false positive 영향은 mutation 차단으로 같지만 정확한 status code 보장은 별도 sprint 후속. | sprint `claude/work_260518-f` (PR #147) |
| 2026-05-18 | §6 atomicity reference CTE 정정 — codex hotfix #6 P2 (PR #147 review) 가 원래 `FROM locked LEFT JOIN upd ON true` 패턴이 locked empty 시 zero rows 를 내서 `lock_token_id IS NULL → 404` 분기가 불가함을 지적. `(VALUES (1)) AS root(_) LEFT JOIN locked LEFT JOIN upd` anchor 패턴으로 정정 — token 미존재 / revoked / 정상 갱신 3 분기 모두 단일 row 결과로 분류 가능. | sprint `claude/work_260518-i` |
