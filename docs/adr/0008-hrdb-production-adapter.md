# ADR-0008: HR DB production 어댑터 결정 (PostgreSQL `hrdb` schema)

- 문서 목적: M3 의 Sign Up 흐름 (`POST /api/v1/auth/signup`) 이 의존하는 HR DB 조회의 production 어댑터를 결정한다. 현재 `internal/hrdb/mock.go` 의 `MockClient` 는 PoC 수준 (in-memory 3 시드 인원). 실 인사 DB 의 source 와 연동 방식을 본 ADR 이 명문화.
- 범위: `HRDBClient` interface 의 production 구현. interface 자체는 그대로 두고 구현체만 교체.
- 대상 독자: Backend 개발자, 인프라 운영자, AI agent.
- 상태: accepted
- 결정일: 2026-05-13
- 결정 근거 sprint: `claude/work_260513-l`.
- 관련 문서: [`backend_api_contract.md` §11.5.2](../backend_api_contract.md#1152-post-apiv1authsignup-api-23-rm-m3-01), [`internal/hrdb/mock.go`](../../backend-core/internal/hrdb/mock.go), [ADR-0003 No-Docker CI](./0003-no-docker-policy-ci-scope.md), [추적성 매트릭스 §2.3.1 RM-M3](../traceability/report.md#231-rm-m3-정의-sprint-claudework_260513-k).

## 1. 컨텍스트

`POST /api/v1/auth/signup` (API-23, RM-M3-01) 은 사용자가 자발적으로 가입할 때 인사 DB 의 `(name, system_id, employee_id, department_name)` 4 필드를 조회해 등록 가능 인원만 통과시킨다. 현재 `internal/httpapi/router.go::HRDBClient` interface 는 다음과 같다:

```go
type HRDBClient interface {
    Lookup(ctx context.Context, systemID, employeeID, name string) (string, string, string, error)
    // returns email, userID, dept, err
}
```

`internal/hrdb/mock.go::MockClient` 가 유일 구현체. 3 시드 인원 (`yklee`, `akim`, `sjones`) 을 보유한 in-memory store. PoC 동작은 정합하지만:

- 실 인사 DB 의 마스터 데이터와 동기화 안 됨.
- 단일 process binary 외부에서 변경 불가.
- 운영 환경에서 가입 가능 인원 갱신이 코드 배포 의존.

따라서 production 어댑터가 필요하다. RM-M3-02 (인사 DB 스키마) 와 함께 본 ADR 이 어댑터 종류를 결정.

## 2. 결정 동인

- **데이터 source 위치**: DevHub 운영 환경의 인사 DB 가 어디 있는가?
  - 사내 별도 HR 시스템 (REST / LDAP / Oracle) — 외부 의존.
  - DevHub PostgreSQL 안의 별도 schema — 내부 의존, sync 필요.
- **ADR-0003 no-docker 정책**: 외부 의존성 (Redis, MongoDB 등) 추가는 ADR-0003 의 운영 단순화 원칙과 충돌.
- **테스트 격리**: 단위테스트가 외부 시스템 의존 없이 가능해야 한다. interface 추상화는 이미 충족.
- **운영 갱신 속도**: 인사 DB 업데이트 (입사/퇴사/부서 이동) 가 어떤 주기로 DevHub 에 반영되어야 하는가? — daily batch / realtime / on-demand 의 trade-off.

## 3. 검토한 옵션

| 옵션 | 설명 | 외부 의존 | 갱신 latency | 평가 |
| --- | --- | --- | --- | --- |
| **A. (선택)** PostgreSQL `hrdb` schema | DevHub PG 안에 별도 schema `hrdb` (또는 `hr` table) + `PostgresClient` 구현. 외부 HR 시스템에서 daily / on-demand ETL 로 sync. | 0 (기존 PG) | ETL 주기 (1차: daily) | ADR-0003 정합 + 운영 단순. 채택. |
| B. REST API 어댑터 | 사내 HR REST API 직접 호출. token / VPN 필요. | 외부 HR 서비스 | 거의 실시간 | 신뢰성 / latency 외부 의존. REST API 가 항상 존재하는지 불확실. |
| C. LDAP / Active Directory 어댑터 | AD 그룹 조회. | 외부 AD | 거의 실시간 | LDAP 클라이언트 의존성 + AD 인프라 가정. 도입 비용 큼. |
| D. MockClient 유지 (PoC 영구화) | 코드 시드 만으로 운영. | 0 | (코드 배포 시) | 운영에서 가입 가능 인원 갱신이 PR 머지 의존 — 비효율. 채택 거부. |

## 4. 결정

**옵션 A 채택** — PostgreSQL `hrdb` schema + `PostgresClient` 구현. **실 구현은 carve out** (본 sprint 의 결정 + 별도 sprint 의 구현).

### 4.1 구현 가이드 (carve out 시점)

- **schema 위치**: 기존 `devhub` DB 안에 `hrdb` schema. Hydra / Kratos 와 같은 schema-isolation 패턴 (ADR-0001 §8 #1).
- **table**: `hrdb.persons` — column = `system_id (PK)`, `employee_id (UNIQUE)`, `name`, `department_name`, `email`, `updated_at`.
- **migration**: `backend-core/migrations/0000XX_create_hrdb_persons.sql` 신규.
- **PostgresClient**: `internal/hrdb/postgres.go` 신규. `HRDBClient` interface 의 두 번째 구현체. `MockClient` 는 dev / 테스트 용도로 그대로 유지.
- **RouterConfig**: prod 환경에서 `PostgresClient` 주입, dev / 테스트 환경은 `MockClient` 유지.
- **ETL strategy**: 1차 = daily cron + `INSERT ... ON CONFLICT (system_id) DO UPDATE`. 후속 = realtime sync 후속 ADR 결정.
- **email 컬럼**: HR DB 가 email 을 보유하지 않으면 `system_id @ <도메인>` 형식 fallback. fallback 정책은 운영 env var 로 결정.

### 4.2 schema migration 1차 안

```sql
CREATE SCHEMA IF NOT EXISTS hrdb;

CREATE TABLE IF NOT EXISTS hrdb.persons (
    system_id        TEXT PRIMARY KEY,
    employee_id      TEXT NOT NULL UNIQUE,
    name             TEXT NOT NULL,
    department_name  TEXT NOT NULL,
    email            TEXT,
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS persons_employee_id_idx ON hrdb.persons (employee_id);
CREATE INDEX IF NOT EXISTS persons_name_idx ON hrdb.persons (lower(name));
```

`Lookup` 쿼리:

```sql
SELECT COALESCE(email, system_id || '@' || $4) AS email, system_id, department_name
FROM hrdb.persons
WHERE lower(system_id) = lower($1) AND employee_id = $2 AND lower(name) = lower($3);
```

`$4` = 운영 env var (`DEVHUB_HR_EMAIL_FALLBACK_DOMAIN`, 기본 `example.com`).

## 5. 결과 (Consequences)

### 긍정적

- 외부 인프라 추가 0 — 기존 DevHub PG 안에서 해결. ADR-0003 정합.
- `MockClient` / `PostgresClient` 의 두 구현체로 dev / prod 분리 — 테스트 격리 유지.
- HR 시스템 변경 시에도 schema 만 일관되면 ETL 만 갱신.

### 부정적 / 트레이드오프

- ETL 의 sync latency (1차 daily). realtime 요구는 후속 ADR 후보.
- HR 마스터 데이터의 ownership 모호 — DevHub 가 HR 의 사본을 보유. 운영 정책 (HR 변경 → DevHub sync 책임자) 결정 필요.

### 비변경 사항

- 본 ADR 채택 시점에 코드 변경 0 — `MockClient` 유지.
- M3 후속 sprint 의 implementation step 으로 carve out.

## 6. 미해결 항목 (Open questions)

| 항목 | 후속 결정 |
| --- | --- |
| ETL 책임 (수동 SQL / cron / 외부 ETL tool) | **부분 결정 (2026-05-13, sprint `claude/work_260513-n`)**: 1차 PoC seed 는 `scripts/hrdb_etl_seed.sql` (idempotent `INSERT ... ON CONFLICT`). 운영 cron 은 별도 운영 sprint 의 결정으로 위임 (1차 권장: daily `psql -f` cron + 사내 HR 시스템 export → sync 패턴). |
| `email` 컬럼 fallback 정책 (env var vs 별도 ADR) | 구현 시점. 1차 env var. |
| 실시간 sync 요구 (퇴사자 즉시 차단) 시 별도 ADR | 운영 진입 후 정책. |
| `MockClient` 의 dev / 테스트 영구 유지 vs deprecation | 구현 시점. 1차 = 그대로 유지 (dev / 테스트 격리용). |

## 7. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-13 | 1차 작성 (sprint `claude/work_260513-l`). RM-M3-02 의 어댑터 결정 + 구현 carve out. |
