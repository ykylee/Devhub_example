# ADR-0010: 사용자의 `primary_dept` 자동 판정 알고리즘

- 문서 목적: M3 의 RM-M3-03 carve out 중 "primary_dept 자동 판정 (겸임 우선순위, 동급 시 자식 노드 수)" 항목의 결정. 사용자가 다수 unit 에 `unit_appointments` 으로 소속된 경우 어느 unit 이 "주 소속" 인지 결정하는 단일 알고리즘을 명문화한다.
- 범위: `domain.AppUser.PrimaryUnitID` 의 backfill / re-derive 알고리즘. 명시 설정 (`users.primary_unit_id` 가 admin 입력으로 채워진 경우) 은 자동 판정 대상이 아니다.
- 대상 독자: Backend 개발자, organization 도메인 stakeholder, AI agent.
- 상태: accepted
- 결정일: 2026-05-13
- 결정 근거 sprint: `claude/work_260513-n`.
- 관련 문서: [ADR-0009](./0009-org-secondary-memberships-and-total-count-mv.md), [`docs/organizational_hierarchy_spec.md`](../organizational_hierarchy_spec.md), [`docs/backend_requirements_org_hierarchy.md`](../backend_requirements_org_hierarchy.md) §1·2, [추적성 매트릭스 §2.3.1 RM-M3](../traceability/report.md#231-rm-m3-정의-sprint-claudework_260513-k).

## 1. 컨텍스트

`users.primary_unit_id` 가 비어 있거나 stale 한 상황 (예: 신규 사용자가 여러 unit 에 동시에 appointment 된 직후, 또는 admin 수동 입력 누락) 에서 시스템이 "주 소속" 을 자동 결정해야 한다. 현재 코드에는 자동 판정 알고리즘이 부재 — 매트릭스 §3 의 조직 행 + ADR-0009 §6 미해결 항목으로 인계되어 있었다.

`unit_appointments` 의 row 는 `(user_id, unit_id, appointment_role ∈ {leader, member})`. 한 사용자가 `leader` 1개 + `member` N개를 가질 수 있다. 동일 사용자가 `leader` 인 unit 이 2개 이상인 경우는 store invariant 부재 (ADR-0009 §6 carve) 이지만 본 ADR 의 알고리즘은 그 invariant 가 깨져도 deterministic 한 결과를 반환해야 한다.

## 2. 결정 동인

- **deterministic**: 같은 입력 → 같은 출력. 두 번 호출해도 결과 변경 없음.
- **명시 설정 우선**: admin 이 `users.primary_unit_id` 를 명시한 경우 알고리즘이 override 하지 않음.
- **leader 우선**: 조직 운영 측면에서 "장이 있는 부서" 가 사용자의 주 소속.
- **타이브레이커 명확**: leader 가 여러 unit 인 invariant 위반 시 또는 leader 가 없는 경우의 fallback 도 결정적이어야.

## 3. 검토한 옵션

| 옵션 | 동작 | 평가 |
| --- | --- | --- |
| **A. (선택)** leader > member, leader 동률 시 자식 노드 수 ↓ 우선, 동률 시 `unit_id` lexicographic | 본 ADR 의 algorithm. | deterministic + 운영 직관 + invariant 위반 graceful. 채택. |
| B. 가장 오래된 appointment (joined_at) 우선 | `unit_appointments.created_at` 으로 정렬. | 1차 sprint 의 시간 정보 부재 (`unit_appointments` table 에 `created_at` 컬럼은 있으나 0-second resolution). 거부. |
| C. 무조건 lexicographic `unit_id` 만 | 단순. | leader/member 구분 무시 → 운영 의도와 충돌. 거부. |
| D. random with deterministic seed | seed = user_id. | 운영자 예측 불가. 거부. |

## 4. 결정

**옵션 A 채택**. 1차 알고리즘 (Go pseudocode):

```go
func ResolvePrimaryUnit(appointments []Appointment, unitTotalCounts map[string]int) (string, bool) {
    if len(appointments) == 0 {
        return "", false
    }

    // Step 1: leader 인 appointment 만 필터링. 0 개면 member 전체로 fallback.
    candidates := filter(appointments, func(a Appointment) bool {
        return a.AppointmentRole == AppointmentRoleLeader
    })
    if len(candidates) == 0 {
        candidates = appointments
    }

    // Step 2: total_count 가 큰 (자식 노드 + 본인 멤버 수 합) unit 우선.
    sort.SliceStable(candidates, func(i, j int) bool {
        ci := unitTotalCounts[candidates[i].UnitID]
        cj := unitTotalCounts[candidates[j].UnitID]
        if ci != cj {
            return ci > cj
        }
        // Step 3: total_count 동률 시 unit_id lexicographic 으로 결정적 결과.
        return candidates[i].UnitID < candidates[j].UnitID
    })
    return candidates[0].UnitID, true
}
```

### 4.1 결정 규칙

1. **명시 설정 우선**: `users.primary_unit_id` 가 non-empty 이고 그 unit 이 사용자의 `unit_appointments` 에 존재하면 그대로 반환 — 알고리즘 미적용. (caller 측 책임)
2. **알고리즘 적용**: `users.primary_unit_id` 가 empty 또는 stale (`unit_appointments` 에 미존재) 시 위 `ResolvePrimaryUnit` 호출.
3. **빈 결과**: appointments 가 empty 인 사용자는 primary 미배정. caller 가 빈 문자열 / null 로 처리.

### 4.2 1차 구현 위치

- 함수: `internal/domain/primary_unit.go::ResolvePrimaryUnit`
- 입력: `[]Appointment` + `unitTotalCounts map[string]int` (caller 가 `org_units_total_count` MV 또는 동적 CTE 결과 주입)
- 단위테스트: 4 case (leader 1개 / leader 없음 / leader 여러개 (invariant 위반) / 동률 lexicographic)

### 4.3 통합 후속 (carve out)

- `users.primary_unit_id` backfill worker (cron 또는 admin endpoint trigger)
- ETL 시점 (`hrdb.persons` sync 직후) 의 자동 호출
- `getUser` / `listUsers` 응답에 derived primary 반영

## 5. 결과 (Consequences)

### 긍정적

- deterministic — 운영자가 같은 input 으로 같은 output 예측 가능.
- leader uniqueness store invariant 가 깨져도 graceful (lexicographic tie-break).
- 단위 함수 — DB 의존 없이 테스트 가능.

### 부정적 / 트레이드오프

- `unitTotalCounts` 가 stale 한 MV 일 때 결과가 일시적으로 부정확할 수 있음 (ADR-0009 의 cron 갱신 latency 와 동일 trade-off).
- 파견 (`is_seconded=true`, `current_unit_id ≠ primary_unit_id`) 시 본 알고리즘은 `primary_unit_id` 만 결정 — `current_unit_id` 결정은 별도 운영 정책.

### 비변경 사항

- `users.primary_unit_id` 의 admin 명시 설정은 그대로 우선.
- 본 ADR 채택 시점에 schema 변경 0.

## 6. 미해결 항목 (Open questions)

| 항목 | 후속 결정 |
| --- | --- |
| backfill 시점 (signup 직후 / cron / admin trigger) | 운영 sprint. 1차 권장: signup 직후 + admin endpoint. |
| 파견 종료 시 `primary_unit_id` 자동 복원 | 운영 sprint + ADR-0009 §6 의 `is_seconded` 자동 갱신 trigger 와 통합. |
| `unit_appointments.created_at` 활용 (옵션 B 후속) | 동률 case 가 운영에서 잦을 시 재고. |

## 7. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-13 | 1차 작성 (sprint `claude/work_260513-n`). 옵션 A 채택 + 1차 알고리즘 + 단위테스트. backfill worker 는 carve. |
