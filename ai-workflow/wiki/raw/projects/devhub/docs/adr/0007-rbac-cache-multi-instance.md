# ADR-0007: RBAC PermissionCache 다중 인스턴스 일관성 결정

- 문서 목적: M1-DEFER-E ("`PermissionCache` 다중 인스턴스 일관성") 의 후속 결정. M3 진입 시 backend-core 가 단일 인스턴스를 넘어 N 개로 확장될 때 RBAC matrix cache 의 일관성을 어떻게 보장할지 결정한다. 본 ADR 채택 시점에 구현은 carve out — M3 진입 시 본 ADR 의 결정에 따라 추가 코드 작성.
- 범위: backend-core 의 RBAC matrix in-memory cache. RBAC policy edit (POST/PUT/DELETE `/api/v1/rbac/policies`) 시 다른 인스턴스의 cache 가 stale 한 상태로 남는 문제를 해소.
- 대상 독자: Backend 개발자, 운영자, AI agent.
- 상태: accepted
- 결정일: 2026-05-13
- 결정 근거 sprint: `claude/work_260513-j`.
- 관련 문서: [`backend_api_contract.md` §12.10](../backend_api_contract.md#1210-cache-와-무효화-api-40), [`backend-core/internal/httpapi/permissions.go`](../../backend-core/internal/httpapi/permissions.go), [ADR-0002](./0002-rbac-policy-edit-api.md), [추적성 매트릭스 §5.2](../traceability/report.md#52-id-부재--매핑-누락).

## 1. 컨텍스트

`IMPL-rbac-04` (`PermissionCache`) 는 process-local in-memory matrix cache. RBAC policy edit 핸들러 (POST/PUT/DELETE `/api/v1/rbac/policies`, PUT `/api/v1/rbac/subjects/.../roles`) 가 머지 후 동일 process 의 `PermissionCache.Invalidate()` 를 호출하여 next request 가 store 에서 재로드한다.

**문제**: backend-core 가 N 개 인스턴스로 가동되면 한 인스턴스의 edit 핸들러가 자기 cache 만 invalidate 하고, 다른 N-1 인스턴스의 cache 는 stale 한 매트릭스로 enforcement 를 계속한다. RBAC permission grant/revoke 가 즉시 모든 인스턴스에 반영되지 않는 일관성 회귀.

`backend_api_contract.md` §12.10 (API-40) 이 이 한계를 명시:
> 다중 인스턴스 환경의 cache 일관성은 §6 미해결 — 운영 phase 진입 시 pub/sub 또는 polling 으로 보강.

`M1-PR-review-actions.md` 의 M1-DEFER-E 가 같은 항목을 후속으로 인계. PR #92 의 §3 RBAC 행이 IMPL-rbac-04 의 책임을 "API-40 in-memory matrix cache + Invalidate" 로 정의했지만, 다중 인스턴스 일관성은 본 ADR 의 결정 대상.

## 2. 결정 동인

- **현재 PoC = 단일 인스턴스**: 운영 진입 (M3+) 시점에 N 인스턴스 확장 가능성. PoC 시점에 구현 미보유 → drift 위험.
- **인프라 영향**: pub/sub (Redis 등) 도입은 인프라 추가. PostgreSQL LISTEN/NOTIFY 는 기존 PG 의존성 안에서 해결. 폴링은 RBAC matrix 변경 latency tradeoff.
- **본 ADR 시점의 구현 carve**: M3 진입 결정과 함께 실제 코드 작성. 본 ADR 은 *결정* 만.

## 3. 검토한 옵션

| 옵션 | 설명 | 일관성 latency | 인프라 영향 | 평가 |
| --- | --- | --- | --- | --- |
| **A. (선택)** PostgreSQL `LISTEN`/`NOTIFY` | edit 핸들러가 `NOTIFY rbac_invalidate` 발행, 각 인스턴스가 `LISTEN` 으로 즉시 cache invalidate. | 거의 즉시 (PG round-trip) | 0 (기존 PG 의존성 안) | no-docker 정책 (ADR-0003) 정합. 채택. |
| B. Redis pub/sub | edit 핸들러가 Redis channel publish, 인스턴스가 subscribe. | 즉시 (Redis round-trip) | Redis 신규 인프라 의존성 추가 | ADR-0003 의 no-docker / 외부 의존성 최소화 원칙에 충돌. 거부. |
| C. 폴링 (N 초마다 store re-load) | 각 인스턴스가 일정 주기로 strore re-load. | N 초 (조정 가능) | 0 | latency tradeoff + DB load 증가. RBAC matrix 변경 빈도 (낮음) 대비 비효율. |
| D. carve out (M3 진입 후 별도 결정) | 본 ADR 채택 시점에는 결정 보류. | N/A | N/A | 본 ADR 의 의도와 충돌 — 결정 명문화가 목적. |

## 4. 결정

**옵션 A 채택** — PostgreSQL `LISTEN`/`NOTIFY` 기반 cache invalidate. **구현은 M3 진입 시 carve out**.

### 4.1 설계 개요 (구현 시 가이드)

- `IMPL-rbac-04::PermissionCache` 에 `NOTIFY` 발행 + `LISTEN` 리스너 추가.
- 새 채널 이름: `devhub_rbac_invalidate` (단일 채널, payload 는 invalidate event timestamp 또는 trigger source 정보).
- edit 핸들러 (`createRBACPolicy`, `updateRBACPolicies`, `deleteRBACPolicy`, `setSubjectRoles`) 가 store commit 직후 `tx.Exec("NOTIFY devhub_rbac_invalidate, $1", payload)`.
- 각 인스턴스 부팅 시 별도 goroutine 이 `LISTEN devhub_rbac_invalidate` 로 구독. notification 수신 시 `PermissionCache.Invalidate()` 호출.
- 단일 인스턴스 (PoC) 환경에서도 동일 흐름 — `LISTEN` 이 자기 자신의 `NOTIFY` 도 받지만 idempotent.

### 4.2 구현 carve out

- 본 ADR 채택 시점 (2026-05-13) 의 구현은 변경 없음. PoC 가 단일 인스턴스라 일관성 문제 발현 없음.
- M3 진입 sprint plan 의 명시적 항목: "ADR-0007 결정에 따라 PostgreSQL `LISTEN`/`NOTIFY` cache invalidate 도입 + 통합 테스트 (2 인스턴스 시뮬레이션)".
- 통합 테스트 위치 후보: `backend-core/internal/httpapi/permissions_integration_test.go` (다중 connection / NOTIFY round-trip 검증).

## 5. 결과 (Consequences)

### 긍정적

- M3 진입 시점에 RBAC enforcement 일관성이 거의 즉시 보장.
- 인프라 추가 없이 기존 PG 의존성 안에서 해결 — ADR-0003 no-docker 정책 정합.

### 부정적 / 트레이드오프

- PG `LISTEN`/`NOTIFY` 는 connection-bound — 인스턴스가 long-lived connection 을 보유해야 함 (connection pool 의 dedicated slot).
- `pq.NewListener` 또는 동등 dependency 의 reconnect 로직 필요 — PG restart / network blip 회복.

### 비변경 사항

- 본 ADR 채택 시점에 코드 변경 0.
- 단일 인스턴스 PoC 환경 동작 변경 0.

## 6. 미해결 항목 (Open questions)

| 항목 | 후속 결정 |
| --- | --- |
| `NOTIFY` payload 형식 (timestamp / role_id / 모두) | 구현 시점 결정. `IMPL-rbac-04` re-load 가 전체 매트릭스이므로 payload 는 informational. |
| `LISTEN` connection 의 failure handling | 구현 시점 결정. backoff retry + warning log + 단일 인스턴스 fallback. |
| 통합 테스트 환경 (2 인스턴스 시뮬레이션) | CI 잡 신규 또는 기존 backend-unit 확장 — 구현 시점 결정. |

## 7. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-13 | 1차 작성 (sprint `claude/work_260513-j`). M1-DEFER-E closing. 구현은 M3 진입 시 carve out. |
