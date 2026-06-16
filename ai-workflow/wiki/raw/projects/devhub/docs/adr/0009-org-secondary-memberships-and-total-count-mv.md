# ADR-0009: 파견/겸임 (secondary memberships) 모델 + `total_count` Materialized View

- 문서 목적: M3 후속 RM-M3-03 의 "1:N 파견/겸임 테이블" + "total_count Materialized View" 항목의 결정 ADR. 1:N 파견/겸임은 이미 `unit_appointments` 가 cover — 본 ADR 이 그 사실 명문화. `total_count` 는 MV 신설 결정만, 구현 carve.
- 범위: organization 도메인의 파견/겸임 spec 정합 + sub-tree member count 캐싱 전략.
- 대상 독자: Backend 개발자, organization 도메인 stakeholder, AI agent.
- 상태: accepted
- 결정일: 2026-05-13
- 결정 근거 sprint: `claude/work_260513-m`.
- 관련 문서: [`migrations/000004_create_users_units.up.sql`](../../backend-core/migrations/000004_create_users_units.up.sql), [`docs/organizational_hierarchy_spec.md`](../organizational_hierarchy_spec.md), [`docs/backend_requirements_org_hierarchy.md`](../backend_requirements_org_hierarchy.md), [추적성 매트릭스 §2.3.1 RM-M3](../traceability/report.md#231-rm-m3-정의-sprint-claudework_260513-k).

## 1. 컨텍스트

`development_roadmap.md` §5 백로그 와 `backend_requirements_org_hierarchy.md` §1·2 가 4 항목을 미해결로 인계:

1. `parent_id` cycle 검증 — **본 sprint 의 RM-M3-03 1차에서 처리** (`updateOrgUnit::cycle_detected`, 직접 self-reference + ancestor chain). 본 ADR 범위 외.
2. `primary_dept` 자동 판정 — 알고리즘 carve out (별도 ADR 후보).
3. 파견/겸임 1:N 테이블 — **본 ADR 의 결정 #1**.
4. `total_count` Materialized View — **본 ADR 의 결정 #2**.

### 1.1 파견/겸임 현황

`migrations/000004_create_users_units.up.sql` 가 이미 다음을 cover:

- `users.primary_unit_id` (소속), `users.current_unit_id` (현재 근무지, 파견 시 다름), `users.is_seconded` (파견 boolean).
- `unit_appointments(user_id, unit_id, appointment_role)` N:N — `appointment_role ∈ {'leader', 'member'}`, `UNIQUE(user_id, unit_id)`.
- `domain.AppUser.Appointments` wire format 으로 사용자별 unit 목록 노출 (`appUserResponse.appointments`).

즉 1:N 파견/겸임은 `unit_appointments` 가 이미 cover. 별도 `unit_secondary_memberships` 테이블 신설 불요.

### 1.2 `total_count` 현황

`orgUnitResponse.total_count` 가 wire 에 노출되지만 schema 레벨 컬럼은 부재. backend code 가 동적으로 sub-tree 멤버 수를 RECURSIVE CTE 로 계산. 성능 측면에서 매트릭스 응답마다 재계산 — 조직 규모가 커질수록 latency 증가.

## 2. 결정 동인

- **파견/겸임 모델 정합**: 별도 테이블 신설은 schema bloat. 기존 `unit_appointments` + `users` 의 3 컬럼 (`primary_unit_id`, `current_unit_id`, `is_seconded`) 이 spec 의 모든 요구 cover.
- **`total_count` 성능**: 조직 100+ unit 환경에서 hierarchy 조회마다 RECURSIVE CTE 가 50-200ms latency. MV 로 캐시하면 ms 단위.
- **MV 갱신 정책**: trigger (즉시 갱신, write 비용) vs cron (1-5분 stale, write 비용 0). organization 변경 빈도가 낮아 cron 충분.

## 3. 검토한 옵션

### 3.1 결정 #1 — 파견/겸임 모델

| 옵션 | 설명 | 평가 |
| --- | --- | --- |
| **A. (선택)** `unit_appointments` 유지 + spec 명문화 | 기존 N:N + `users.is_seconded` 사용. 본 ADR 이 cover 완료 사실 명시. | 추가 schema 0. 채택. |
| B. `unit_secondary_memberships` 별도 테이블 신설 | secondary 만 별도 테이블. | schema bloat, 기존과 cross-cut. 거부. |

### 3.2 결정 #2 — `total_count` MV

| 옵션 | 설명 | 갱신 방식 | 평가 |
| --- | --- | --- | --- |
| **A. (선택)** `org_units_total_count` MV + cron 갱신 | RECURSIVE CTE 결과를 MV 캐시. 1-5분 cron 으로 `REFRESH MATERIALIZED VIEW CONCURRENTLY`. | cron | 성능 + 운영 단순. 채택. |
| B. trigger 기반 즉시 갱신 | unit_appointments / org_units INSERT/UPDATE/DELETE 마다 trigger 가 MV refresh. | trigger | write 비용 + concurrency 복잡. 거부. |
| C. 매번 동적 RECURSIVE CTE (현재) | MV 도입 없이 매 호출마다 재계산. | 즉시 | 조직 규모 확장 시 latency. M3 후속 carve 와 충돌. |
| D. application-level 캐시 (in-memory) | backend process 가 hierarchy 캐싱. | invalidate 패턴 | RBAC cache (ADR-0007) 와 동일 패턴 — 다중 인스턴스 일관성 문제. 거부. |

## 4. 결정

### 4.1 파견/겸임 = 옵션 A — `unit_appointments` 유지 + 모델 cover 사실 명문화

추가 schema 없음. spec 측면에서는 다음과 같이 정합 확인:

- **primary membership** (소속 부서): `users.primary_unit_id`. 1:1.
- **current unit** (현 근무지, 파견 시 임시): `users.current_unit_id`. 1:1.
- **secondary memberships** (겸임): `unit_appointments.appointment_role='member'` 인 row 들. user 1 = unit N. `appointment_role='leader'` 가 정확히 0..1 보장은 store invariant (현재 unique constraint 는 `(user_id, unit_id)` 단위) — 별도 leader uniqueness 검증은 후속 carve.

### 4.2 `total_count` = 옵션 A — `org_units_total_count` MV + cron 갱신

migration `000012_create_total_count_mv.up.sql` (carve out, M3 후속 sprint):

```sql
CREATE MATERIALIZED VIEW org_units_total_count AS
WITH RECURSIVE descendants(root_unit_id, descendant_unit_id) AS (
    SELECT unit_id, unit_id FROM org_units
    UNION ALL
    SELECT d.root_unit_id, child.unit_id
    FROM descendants d
    JOIN org_units child ON child.parent_unit_id = d.descendant_unit_id
)
SELECT u.unit_id AS unit_id, COUNT(DISTINCT ua.user_id) AS total_count
FROM org_units u
LEFT JOIN descendants d ON d.root_unit_id = u.unit_id
LEFT JOIN unit_appointments ua ON ua.unit_id = d.descendant_unit_id
GROUP BY u.unit_id
WITH NO DATA;

CREATE UNIQUE INDEX org_units_total_count_pk ON org_units_total_count (unit_id);

REFRESH MATERIALIZED VIEW CONCURRENTLY org_units_total_count;
```

cron 갱신 (carve, 운영 sprint):
- 5분 주기 `REFRESH MATERIALIZED VIEW CONCURRENTLY org_units_total_count`.
- 매 RBAC policy / org_unit 변경 직후 명시 호출 (best-effort).

backend code 변경 (carve):
- `getHierarchy` 의 `total_count` 컬럼 채우기 → MV join 으로 교체.
- RECURSIVE CTE 동적 계산은 fallback (MV 가 stale 일 가능성 대비).

## 5. 결과 (Consequences)

### 긍정적

- 파견/겸임 schema 가 이미 cover — 별도 변경 0.
- MV 도입 후 hierarchy 응답 latency 감소.
- cron 갱신은 운영 단순, ADR-0007 (RBAC LISTEN/NOTIFY) 와는 다른 trade-off (organization 변경 빈도가 낮음).

### 부정적 / 트레이드오프

- MV 가 stale 한 windowed time (1-5분) 동안 sub-tree count 가 부정확. UI 가 즉시 정확성 요구하면 carve out fallback (RECURSIVE CTE) 사용.
- `unit_appointments` 의 leader uniqueness 검증이 store invariant 부재 — `(user_id, unit_id, appointment_role='leader')` 가 unit 마다 정확히 0..1 인지는 application 책임. 후속 ADR 후보.

### 비변경 사항

- 본 ADR 채택 시점에 코드 / migration 변경 0.
- M3 후속 sprint 의 implementation step:
  - `migrations/000012_create_total_count_mv.up.sql` 신규
  - backend `getHierarchy` 의 MV join
  - cron 갱신 (운영 sprint)
  - leader uniqueness store invariant 후속 ADR

## 6. 미해결 항목 (Open questions)

| 항목 | 후속 결정 |
| --- | --- |
| MV `CONCURRENTLY` refresh 의 transaction isolation | 구현 시점. 1차 = `CONCURRENTLY`, 5분 cron. |
| Leader uniqueness store invariant (`unit_appointments` 의 `appointment_role='leader'` 가 unit 마다 0..1) | 별도 ADR 후보. application 측 검증이 1차. |
| `primary_dept` 자동 판정 알고리즘 (겸임 우선순위, 동급 시 자식 노드 수) | 본 ADR 범위 외. `organizational_hierarchy_spec.md` §3 의 결정 후 별도 ADR 후보. |
| 파견 종료 (`is_seconded=false` + `current_unit_id` 정합) 자동 갱신 trigger / worker | 운영 sprint. |

## 7. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-13 | 1차 작성 (sprint `claude/work_260513-m`). 파견/겸임 모델 cover 사실 + `total_count` MV 채택. 구현은 모두 carve out. |
