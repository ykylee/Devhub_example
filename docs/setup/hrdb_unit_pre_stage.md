# HRDB ETL unit pre-stage 가이드 (사내 운영)

- 문서 목적: 사내 HRDB 의 사용자 unit 매핑 정보를 Keycloak user attribute 또는 DevHub `hrdb.persons` schema 로 사전 stage 하는 운영 절차를 정의한다. [ADR-0020 §3.2 + §6.3](../adr/0020-account-user-management-boundary.md) 의 "HRDB ETL push 의 unit 매핑 정보 stage" 사내 동반 carve.
- 범위: HRDB → Keycloak Admin REST 또는 DevHub PostgreSQL `hrdb.persons` schema 의 ETL 패턴 + DevHub onboarding 흐름과의 정합 + cross-check 후속 carve. 외부 Keycloak 시나리오 (사내 IdP 팀이 별도 운영하는 Keycloak) 가정.
- 대상 독자: 사내 HRDB / DBA 팀, 사내 IdP 팀, DevHub 운영자.
- 상태: draft (1차)
- 최종 수정일: 2026-05-22
- 결정 근거 sprint: `claude/work_260522-internal-coordinated-carve-docs`
- 관련 문서: [ADR-0008 HRDB production adapter](../adr/0008-hrdb-production-adapter.md), [ADR-0020 §3.2 + §6.3](../adr/0020-account-user-management-boundary.md), [ADR-0021 Onboarding §3 + §6.2](../adr/0021-onboarding-self-service-unit-selection.md), [Keycloak 운영 SOP §5.2](./keycloak_operations.md#52-custom-claim--employee_id-hrdb-sync-핵심), [keycloak_admin_responsibility](../governance/keycloak_admin_responsibility.md), [keycloak_offboarding_immediacy](../planning/keycloak_offboarding_immediacy.md).

## 1. 배경

[ADR-0020](../adr/0020-account-user-management-boundary.md) 가 외부 Keycloak 시나리오를 채택 후 (sprint -q PR #257 의 HRDB ETL cron 폐기 정합), HRDB 의 사용자 정보 (unit 매핑 포함) 는 **사내 IdP 팀이 운영하는 ETL → Keycloak Admin REST → user attribute** 경로로 stage 된다. DevHub backend 는 이를 token claim 또는 lookup 으로 활용.

[ADR-0021 Onboarding](../adr/0021-onboarding-self-service-unit-selection.md) 의 self-service 흐름에선 사용자 본인이 `/devhub/onboarding` 에서 unit 을 명시 선택한다. HRDB pre-stage 가 있으면 onboarding picker 의 **default 값을 자동 제공** + admin 검토 시 HRDB cross-check 활성. pre-stage 가 없으면 사용자 입력만으로 동작 (현 1차 운영 모드).

본 가이드가 두 시나리오의 운영 절차를 명문화 — 사내 IdP 팀의 ETL 설계 input + DevHub 운영자의 cross-check 후속 carve 안내.

### 1.1 외부 Keycloak 시나리오 정착 (2026-05-20)

[state.json sprint -q (PR #257)](../../ai-workflow/memory/state.json) 결정 — DevHub 자체의 `scripts/hrdb_etl_sync.sh` cron 폐기. HRDB ↔ Keycloak sync 책임이 외부 IdP 팀으로 이관. 본 가이드는 사내 IdP 팀이 운영할 ETL 의 설계 reference + DevHub 측 정합 요구사항.

## 2. 데이터 경로 옵션

### 2.1 옵션 A — HRDB → Keycloak user attribute (권장)

```
사내 HRDB ──(사내 IdP 팀 ETL)──▶ Keycloak Admin REST
                                    │
                                    ▼
                          Keycloak user attribute
                          (`employee_id`, `unit_id`, `department_name`)
                                    │
                                    ▼
                          Token claim 매핑 (User Attribute Mapper)
                                    │
                                    ▼
                          DevHub backend (token claim 우선) 또는
                          KeycloakAdminClient.GetUserDetails lookup
```

**장점**:
- 단일 source-of-truth (Keycloak)
- DevHub PG 안에 HRDB 사본 보관 불필요 — schema 격리 (ADR-0020 §3.2 정합)
- Off-boarding sync 자연 통합 ([keycloak_offboarding_immediacy.md](../planning/keycloak_offboarding_immediacy.md) Phase 1 옵션 C)

**단점**:
- Keycloak user attribute 의 schema 변경이 사내 정책 결정 동반 (custom mapper 추가)
- token size 증가 (claim 추가)

**적용 시기**: 외부 Keycloak 시나리오의 정공법. **권장**.

### 2.2 옵션 B — HRDB → DevHub `hrdb.persons` schema (보조)

[ADR-0008](../adr/0008-hrdb-production-adapter.md) 의 결정 — DevHub PG 안의 별도 `hrdb` schema. 사내 HR ETL 이 daily SQL push 또는 DevHub 가 직접 lookup (사내 HR REST API).

```
사내 HRDB ──(daily SQL push 또는 ETL tool)──▶ DevHub PG `hrdb.persons`
                                                      │
                                                      ▼
                                       DevHub backend (`internal/hrdb/postgres.go`)
                                                      │
                                                      ▼
                                       Onboarding cross-check (carve)
```

**장점**:
- DevHub 가 HRDB 의 read 전용 사본 보유 → cross-check 단순
- Keycloak 측 schema 변경 불필요

**단점**:
- 이중 source-of-truth (Keycloak claim + DevHub hrdb.persons) — divergence 위험
- HR 마스터 데이터 ownership 모호 (ADR-0008 §5)
- daily ETL latency

**적용 시기**: 옵션 A 사내 인프라 미동반 시 보조. 또는 cross-check 후속 carve 의 입력.

### 2.3 옵션 C — 사용자 self-service only (현 1차 운영 모드)

ADR-0021 onboarding 흐름의 자체 동작 — 사용자가 `OrganizationPicker` 에서 명시 선택. HRDB pre-stage 미동반.

**장점**: 사내 인프라 변경 0
**단점**: 사용자가 자기 unit 을 잘못 입력해도 admin 검토 (`pending_review → reviewed`) 에서만 catch — pre-stage 가 있으면 input 시점에 detect 가능

**적용 시기**: pre-stage 미준비 시 default. 옵션 A 또는 B 도입 후 자동 fill-in 으로 보강.

## 3. 옵션 A 운영 절차 (권장)

### 3.1 사내 HRDB 측 (사내 IdP 팀 운영)

| 단계 | 작업 |
| --- | --- |
| 1 | HRDB 의 사용자 (`employee_id`, `unit_id`, `department_name`) 일자별 diff 추출 |
| 2 | Keycloak Admin REST 의 `PUT /admin/realms/devhub/users/:id` 호출 — user attribute 갱신 |
| 3 | 신규 사용자 시 `POST /admin/realms/devhub/users` 로 생성 + attribute set |
| 4 | 퇴사자 시 `PUT .../users/:id` 의 `enabled=false` ([keycloak_offboarding_immediacy.md](../planning/keycloak_offboarding_immediacy.md) Phase 1) |
| 5 | ETL 실패 시 사내 alerting (재시도 + 에스컬레이션) |

**Keycloak user attribute key 명명**:

| HRDB 컬럼 | Keycloak user attribute key | DevHub 사용처 |
| --- | --- | --- |
| `employee_id` | `employee_id` | [Keycloak 운영 SOP §5.2](./keycloak_operations.md#52-custom-claim--employee_id-hrdb-sync-핵심) 정합 — token claim |
| `unit_id` | `unit_id` | DevHub `users.primary_unit_id` 자동 매핑 (본 가이드 §3.3) |
| `department_name` | `department_name` | DevHub 표시용 (admin filter / 사용자 검색 보조) |

### 3.2 Keycloak User Attribute Mapper 추가

[Keycloak 운영 SOP §5.2](./keycloak_operations.md#52-custom-claim--employee_id-hrdb-sync-핵심) 패턴 따라 `unit_id` mapper 추가:

1. Keycloak admin console → Clients → `devhub-frontend` → Client Scopes → `profile` → Mappers → Add Mapper → "By configuration"
2. Mapper type = **User Attribute**
3. 설정:
   - Name: `unit_id`
   - User Attribute: `unit_id`
   - Token Claim Name: `unit_id`
   - Claim JSON Type: `String`
   - Add to ID token: `ON`
   - Add to access token: `ON`
   - Add to userinfo: `ON`
4. 저장 후 jwt.io 등으로 access_token decode 해 `unit_id` claim 확인

### 3.3 DevHub backend 자동 매핑 (carve, P2)

본 carve 진입 시 `internal/auth/keycloak_verifier.go` 에 `unit_id` claim 추출 + `AuthenticatedActor` 확장:

```go
type AuthenticatedActor struct {
    Login         string
    Subject       string
    Role          string
    Email         string
    DisplayName   string
    // 본 가이드 §3.3 carve — HRDB pre-stage 시 token claim 으로 unit_id propagate.
    PrimaryUnitID string  // unit_id claim, NULL 시 사용자 self-service onboarding 입력
}
```

- `authenticateActor` 의 `users.idp_subject` lazy backfill 흐름에서 `unit_id` claim 이 있고 DB `primary_unit_id` 이 NULL 이면 자동 매핑
- ADR-0021 §3.2 의 `limited → pending_review` transition 시 사용자 입력 default 값으로 사용

본 carve 는 별도 sprint — pre-stage 인프라 (옵션 A) 가 사내에서 실 운영 시작된 후 진입.

## 4. 옵션 B 운영 절차 (보조)

### 4.1 사내 HRDB → DevHub PG ETL

[ADR-0008 §4.2](../adr/0008-hrdb-production-adapter.md) 의 schema 정합:

```sql
CREATE SCHEMA IF NOT EXISTS hrdb;
CREATE TABLE IF NOT EXISTS hrdb.persons (
    system_id        TEXT PRIMARY KEY,
    employee_id      TEXT NOT NULL UNIQUE,
    name             TEXT NOT NULL,
    department_name  TEXT NOT NULL,
    email            TEXT,
    unit_id          TEXT,            -- (carve) ADR-0008 §4.2 확장 — unit 매핑 추가
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX persons_unit_id_idx ON hrdb.persons (unit_id);
```

> **ADR-0008 §4.2 의 schema 확장 carve** — 현재 ADR-0008 의 spec 은 `unit_id` 컬럼 미포함. 본 가이드의 옵션 B 채택 시 ADR-0008 §4.2 schema 확장 + migration 신규 필요. ADR governance 절차 (`docs/governance/document-standards.md`) 따라 ADR-0008 변경 이력 row + migration `0000XX_add_unit_id_to_hrdb_persons.sql` 동반.

### 4.2 daily ETL cron 패턴

```bash
# 사내 ETL 도구 또는 cron job
#!/bin/bash
# HRDB 사용자 diff 추출 + DevHub PG INSERT/UPDATE
psql "${DEVHUB_DB_URL}" <<SQL
INSERT INTO hrdb.persons (system_id, employee_id, name, department_name, email, unit_id, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW())
ON CONFLICT (system_id) DO UPDATE
SET employee_id = EXCLUDED.employee_id,
    name = EXCLUDED.name,
    department_name = EXCLUDED.department_name,
    email = EXCLUDED.email,
    unit_id = EXCLUDED.unit_id,
    updated_at = NOW();
SQL
```

### 4.3 DevHub backend `internal/hrdb/postgres.go` (carve)

ADR-0008 §4.1 의 `PostgresClient` 구현체 신규. `HRDBClient` interface 의 `Lookup` 메서드 — onboarding cross-check 시 호출.

## 5. Onboarding cross-check 후속 carve

[ADR-0021 §6.2](../adr/0021-onboarding-self-service-unit-selection.md) 의 "HRDB cross-check" 후속 carve — 사용자가 입력한 unit 과 HRDB 매핑 비교 + 불일치 시 경고.

### 5.1 흐름

```
사용자 ─[/devhub/onboarding]─▶ POST /api/v1/me/onboarding (unit_id = "X")
                                            │
                                            ▼
                            backend ─[HRDB lookup (옵션 A claim or 옵션 B DB)]─▶ unit_id = "Y"
                                            │
                            ┌───────────────┴───────────────┐
                            ▼                                ▼
                       (X == Y)                          (X != Y)
                            │                                │
                            ▼                                ▼
                  정상 INSERT + audit                 422 + 경고 또는 audit warn
                                                     ('account.onboarding_hrdb_mismatch')
```

### 5.2 결정 옵션

| 옵션 | 동작 |
| --- | --- |
| **차단** | 사용자가 HRDB 의 unit 외 선택 시 422 — 강제 정합 |
| **경고** | 사용자 입력 그대로 통과 + audit warn — admin 검토 시 cross-check 결과 표시 |
| **사용자 확인** | 사용자 입력이 HRDB 와 다르면 modal — 사용자 본인 확인 후 진행 |

권장 — **경고 + admin 검토 시 표시** (옵션 2). 사용자 의도적 변경 (조직 개편 직후 등) 허용 + admin 가 최종 검토. 사내 정책 결정 후 carve 진입.

## 6. 잔여 carve

| 항목 | 우선순위 | 비고 |
| --- | --- | --- |
| 옵션 A 의 `unit_id` Keycloak user attribute mapper 도입 | P2 | 사내 IdP 팀 동반 |
| 옵션 A backend `AuthenticatedActor.PrimaryUnitID` 자동 매핑 | P2 | 옵션 A 인프라 시작 후 |
| 옵션 B `hrdb.persons.unit_id` 컬럼 ADR-0008 §4.2 schema 확장 | P3 | 옵션 B 채택 시 |
| 옵션 B `internal/hrdb/postgres.go` 구현 | P3 | ADR-0008 §4.1 의 carve |
| ADR-0021 §6.2 의 onboarding cross-check 실 구현 | P3 | 옵션 A 또는 B 인프라 동반 |
| HRDB ETL audit trail 표준 | P3 | 사내 HRDB 팀 정책 |
| Off-boarding 즉시성 Phase 2 (LDAP federation) | P3 | [keycloak_offboarding_immediacy.md](../planning/keycloak_offboarding_immediacy.md) Phase 2 결정 후 |

## 7. 변경 이력

| 일자 | 변경 | sprint |
| --- | --- | --- |
| 2026-05-22 | 1차 draft — §1 외부 Keycloak 시나리오 정착 + §2 데이터 경로 3 옵션 (A Keycloak user attribute 권장 / B DevHub hrdb.persons 보조 / C self-service only 현 1차) + §3 옵션 A 운영 절차 (사내 HRDB 측 5 step + Keycloak User Attribute Mapper + backend 자동 매핑 carve) + §4 옵션 B 운영 절차 (ADR-0008 §4.2 schema 확장 carve + daily ETL cron + PostgresClient carve) + §5 onboarding cross-check 후속 carve (3 결정 옵션 — 차단/경고/확인) + §6 잔여 carve 7건. ADR-0020 §6.3 사내 동반 carve "HRDB ETL push 의 unit 매핑 정보 stage" 의 docs 초안. 사내 실 적용은 IdP 팀 + HRDB 팀 + 사내 정책 결정 동반. | `claude/work_260522-internal-coordinated-carve-docs` |
