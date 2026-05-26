# Onboarding 운영 SOP (staging 1주 monitoring + rollback + incident response)

- 문서 목적: [ADR-0021 Onboarding self-service unit selection](../adr/0021-onboarding-self-service-unit-selection.md) 도메인의 운영 측면 SOP — feature flag default ON (PR #290) 직후 staging 1주 monitoring 절차 + 회귀 발견 시 rollback runbook + incident response 단계화 + DoD acceptance 정의.
- 범위: backend onboardingGate + 4 endpoint (`POST /me/onboarding` / `PATCH /me` / `GET /organizations/search` / `POST /admin/users/:id/review`) + audit event 3종 + `users.onboarding_completed_at` + `users.review_status` 컬럼 + frontend AuthGuard blocklist 모드 (PR #291). Prometheus metric backend 도입 + Grafana dashboard 자산은 별도 carve (§8).
- 대상 독자: 운영자 (SRE / 사내 운영팀), system_admin, backend / frontend / IdP 담당자, on-call.
- 상태: draft (1차)
- 최종 수정일: 2026-05-22
- 결정 근거 sprint: `claude/work_260522-onboarding-ops-sop`
- 관련 문서: [ADR-0021](../adr/0021-onboarding-self-service-unit-selection.md), [Onboarding IMPL plan §7 #6](../planning/onboarding_impl_plan.md), [Keycloak 운영 SOP](./keycloak_operations.md), [release_v1_roadmap §1.3 #7](../planning/release_v1_roadmap.md).

## 1. 배경 + 책임 분리

[ADR-0021](../adr/0021-onboarding-self-service-unit-selection.md) 가 Keycloak token 검증 후 DevHub `users` row 미존재 사용자를 token-only actor 로 정상 처리하고 onboarding 명시 제출까지 보호 endpoint 진입을 차단한다. PR #290 (2026-05-21, issue #284) 가 feature flag `DEVHUB_ONBOARDING_GATE_ENABLED` default 를 `false` → `true` 로 flip 했고 `lazy_auto_create.go` 를 폐기했다. 본 SOP 가 flip 직후 1주 monitoring 기간의 운영 자산을 source-of-truth 로 정합한다.

| 책임 | 주체 | 도구 |
| --- | --- | --- |
| Monitoring 신호 수집 + 1주 acceptance | 운영자 (SRE) | 본 SOP §4 + §7 |
| `pending_review` 사용자 검토 (`reviewed` transition) | system_admin | DevHub `/admin/settings/users` |
| Rollback 결정 + 실행 | 운영자 + on-call | 본 SOP §5 |
| Incident root cause 진단 + bug fix | backend (Claude) / frontend (Gemini) | 본 SOP §6.3 escalation |
| Keycloak claim / event listener 측면 | 사내 IdP 팀 | [keycloak_operations.md §8.6](./keycloak_operations.md#86-keycloak-event-listener-audit_logs-통합-운영-sop) |

## 2. State machine 요약 (운영 관점)

[ADR-0021 §3.2](../adr/0021-onboarding-self-service-unit-selection.md) 3-tier 상태머신:

| 단계 | DB 상태 | 접근 범위 | 운영 관찰 |
| --- | --- | --- | --- |
| `limited (skip)` | `users` row 미존재 (token-only actor) | 공통 메뉴 + onboarding 페이지 + `GET /me` 만 | 매 로그인 시 onboarding 강제 진입. skip 자체는 audit event 없음. |
| `pending_review` | row + `onboarding_completed_at IS NOT NULL` + `review_status='pending_review'` | 공통 메뉴 + 자기 자원 (무소속 처리) | admin 검토 대기. §4.3 SQL #3 으로 누적량 추적. |
| `reviewed` | row + `review_status='reviewed'` | 정상 (모든 도메인) | DoD acceptance — 1주 후 누적 reviewed 비율로 운영 정착도 판단. |

전이 4종 ↔ audit event 3종 1:1 매핑:

| 전이 | 트리거 | Audit action | 응답 코드 |
| --- | --- | --- | --- |
| `(none) → limited` | 미등록 사용자 첫 진입 | (none) | `GET /me` → 200 + `onboarding_required=true` |
| `limited → pending_review` | `POST /api/v1/me/onboarding` 성공 | `account.onboarding_completed` | 201 Created |
| `pending_review → reviewed` | `POST /api/v1/admin/users/:id/review` (system_admin) | `account.review_confirmed` | 200 OK |
| `reviewed → pending_review` | `PATCH /api/v1/me` 의 `primary_unit_id` 변경 | `account.unit_changed` | 200 OK |

## 3. Feature flag 운영

### 3.1 `DEVHUB_ONBOARDING_GATE_ENABLED`

| 상태 | onboardingGate | 4 endpoint | 사용자 영향 |
| --- | --- | --- | --- |
| `true` (default, PR #290) | 미완료 사용자의 allowlist 외 endpoint → 403 `onboarding_required` | 정상 동작 | 첫 로그인 시 `/devhub/onboarding` 강제 진입 |
| `false` (rollback path) | no-op (모든 endpoint 통과) | 404 `onboarding_feature_disabled` | gate 보호 해제 — 미완료 사용자가 보호 endpoint 접근 가능. frontend onboarding 페이지 동작 안 함. |

### 3.2 설정 위치

| 환경 | 파일 | 설정 |
| --- | --- | --- |
| local dev | `backend-core/.env` 또는 shell | `export DEVHUB_ONBOARDING_GATE_ENABLED=1` (default — unset = true) |
| staging / prod (docker) | `docs/setup/deploy.env.example` 파생 환경별 파일 | `DEVHUB_ONBOARDING_GATE_ENABLED=true` (또는 `1`) |
| 비상 rollback | 위 동일 파일 | `DEVHUB_ONBOARDING_GATE_ENABLED=0` (또는 `false`) + backend 재기동 |

`envBoolDefault` helper (`internal/config/config.go`, PR #290) 가 opt-out 패턴 — 미설정 시 default `true` 반환. `0` / `false` / `no` 만 disable.

## 4. 1주 monitoring — 무엇을 볼 것인가

### 4.1 Funnel 4 stage

신규 사용자 1명이 가는 경로:

```
1. Keycloak 로그인 성공
   → backend GET /api/v1/me 호출
   → response.onboarding_required = true
   → frontend redirect /devhub/onboarding

2. 사용자가 OrganizationPicker 에서 소속 선택 + display_name 입력
   → POST /api/v1/me/onboarding
   → 201 Created + audit 'account.onboarding_completed'
   → users row 생성 + onboarding_completed_at=NOW() + review_status='pending_review'

3. system_admin 이 /admin/settings/users 의 pending_review filter 에서 확인
   → POST /api/v1/admin/users/:user_id/review
   → 200 OK + audit 'account.review_confirmed'
   → review_status='reviewed'

4. 사용자가 정상 도메인 접근 (Application/Project/DREQ 등)
   → 403 onboarding_required 응답 0건
```

1주 acceptance 는 위 4 stage 의 dropout / latency / 회귀를 §4.3 SQL 로 추적한다.

### 4.2 핵심 신호 5종 + 임계

| # | 신호 | 임계 (rollback trigger) | 출처 |
| --- | --- | --- | --- |
| S1 | `403 onboarding_required` 응답 비율 | 정상 baseline = (`limited` 사용자 수 × 평균 호출량). baseline × 2 초과가 1시간 지속 시 §5 rollback 검토 | backend access log + audit_logs (gate block 자체는 audit emit 안 함 — log 만) |
| S2 | `POST /me/onboarding` 성공률 | < 95% (5xx 또는 invalid_payload 비율) 1시간 지속 | audit_logs + backend log |
| S3 | `pending_review` 누적량 (검토 적체) | 사내 정책 미정 — **carve**. 임시 임계: 일평균 신규 사용자 × 3 이상이 3일 지속 시 admin 알림 | SQL §4.3 #3 |
| S4 | CHECK 제약 위반 (bi-implication 깨짐) | 1건 발견 즉시 §6 incident | SQL §4.3 #6 |
| S5 | 동일 사용자 onboarding 반복 제출 | 동일 user_id 의 `account.onboarding_completed` audit 2회 이상 | SQL §4.3 #5 |

### 4.3 SQL 쿼리 모음

> 모두 read-only. 운영자가 staging 환경 PostgreSQL 에 read replica 또는 운영 DB 직접 (`psql` SET TRANSACTION READ ONLY) 으로 실행.

#### #1 — 현재 onboarding 단계별 사용자 분포

```sql
SELECT
    CASE
        WHEN onboarding_completed_at IS NULL THEN 'limited_or_admin_preseed'
        WHEN review_status = 'pending_review' THEN 'pending_review'
        WHEN review_status = 'reviewed' THEN 'reviewed'
        ELSE 'unknown'
    END AS stage,
    COUNT(*) AS user_count
FROM users
GROUP BY stage
ORDER BY user_count DESC;
```

#### #2 — 가장 오래 pending_review 인 사용자 (admin TODO)

```sql
SELECT
    user_id,
    display_name,
    primary_unit_id,
    onboarding_completed_at,
    updated_at,
    NOW() - onboarding_completed_at AS pending_age
FROM users
WHERE review_status = 'pending_review'
ORDER BY onboarding_completed_at ASC
LIMIT 50;
```

`pending_age` 가 24h 초과 사용자가 누적되면 system_admin SLA 점검.

#### #3 — 일별 onboarding 완료 / review 확정 추이

```sql
SELECT
    date_trunc('day', created_at) AS day,
    action,
    COUNT(*) AS emit_count
FROM audit_logs
WHERE action IN (
    'account.onboarding_completed',
    'account.review_confirmed',
    'account.unit_changed'
)
  AND created_at >= NOW() - INTERVAL '7 days'
GROUP BY day, action
ORDER BY day DESC, action ASC;
```

`account.onboarding_completed` 와 `account.review_confirmed` 의 일별 ratio (대기 큐 깊이) 추적.

#### #4 — onboarding 제출 실패율 (S2 신호)

backend access log 가 stdout 으로 가는 경우 (`internal/httpapi/me_onboarding.go::submitOnboarding`):

```bash
# 운영 환경 log 경로 (deploy 별로 다름)
grep '/api/v1/me/onboarding' /var/log/devhub/backend.log \
    | awk '{print $9}' \
    | sort | uniq -c | sort -rn
# expected: 201 (성공) > 422 (invalid_payload, 사용자 실수) >> 5xx
# 5xx 가 10건/시간 초과 → §6 incident response
```

(deploy log 위치 환경별 — `docker logs` / `journalctl -u devhub-backend` / k8s `kubectl logs` 등으로 치환.)

#### #5 — 동일 사용자 onboarding 반복 제출 (S5 신호)

```sql
SELECT
    target_id AS user_id,
    COUNT(*) AS emit_count,
    MIN(created_at) AS first_emit,
    MAX(created_at) AS last_emit
FROM audit_logs
WHERE action = 'account.onboarding_completed'
  AND created_at >= NOW() - INTERVAL '7 days'
GROUP BY target_id
HAVING COUNT(*) > 1
ORDER BY emit_count DESC;
```

> 정상은 0행. 1건 이상 발견 시 §6.2 pattern 3 대응. handler 가 `ErrConflict` 로 409 응답하므로 정상 흐름에서는 row 가 중복 INSERT 되지 않는다 — 발견 시 race 또는 client retry 회귀 의심.

#### #6 — CHECK 제약 위반 검증 (S4 신호)

```sql
SELECT
    user_id,
    onboarding_completed_at,
    review_status
FROM users
WHERE
    (onboarding_completed_at IS NULL AND review_status IS NOT NULL)
    OR (onboarding_completed_at IS NOT NULL AND review_status IS NULL);
```

> 정상은 0행 (migration 000033 의 `users_onboarding_review_consistency` CHECK 가 강제). 1건이라도 발견 시 §6 incident — 수동 SQL UPDATE 잔재 또는 CHECK 제약 손상 의심.

#### #7 — token-only actor 비율 (S 보조 신호)

backend log 의 `[authenticateActor] %q DB miss = token-only actor` 메시지 grep:

```bash
# 직전 1시간 token-only 사례
grep 'DB miss = token-only actor' /var/log/devhub/backend.log \
    | awk -v d="$(date -u -d '1 hour ago' --iso-8601=seconds)" '$1 > d' \
    | wc -l
```

신규 사용자 등록 시점 외에 token-only actor 가 발생하면 Keycloak ↔ DevHub idp_subject 동기화 지연 가능성 — [keycloak_operations.md §8.6](./keycloak_operations.md#86-keycloak-event-listener-audit_logs-통합-운영-sop) event listener cursor lag 점검.

### 4.4 Prometheus metric backend (2026-05-26 sprint `claude/work_260526-onboarding-prometheus-metric` 적용)

Onboarding 도메인 backend 에 Prometheus Counter/Histogram 4 종 도입 완료. SQL (§4.3) + Prometheus dashboard 둘 다 사용 가능. metric 정의는 [`backend-core/internal/httpapi/onboarding_metrics.go`](../../backend-core/internal/httpapi/onboarding_metrics.go) 참조.

| metric | type | label | 측정 시점 |
| --- | --- | --- | --- |
| `devhub_onboarding_gate_blocked_total` | Counter | `reason` (`onboarding_required`) | `onboardingGate` middleware 403 차단 직전 |
| `devhub_onboarding_submit_total` | Counter | `status` (`ok` / `rejected` / `conflict` / `not_found` / `server_error` / `unavailable` / `unauthenticated`) | `submitOnboarding` handler 의 7 분기 각각 |
| `devhub_onboarding_submit_duration_seconds` | Histogram | (없음) | `submitOnboarding` 전체 처리 시간 (`defer` 측정) — bucket 10ms ~ 10.24s |
| `devhub_onboarding_review_confirm_total` | Counter | `status` (`ok` / `rejected` / `conflict` / `not_found` / `server_error` / `unavailable` / `bad_request`) | `confirmUserReview` handler 의 7 분기 각각 |

**dashboard 매핑 (§4.2 signal 5종 ↔ metric)**:
- **S1 gate 403 spike** — `rate(devhub_onboarding_gate_blocked_total[5m])` 가 baseline 대비 3× 초과 시 alert
- **S2 submit p95 / 성공률** — `histogram_quantile(0.95, rate(devhub_onboarding_submit_duration_seconds_bucket[5m]))` + `rate(devhub_onboarding_submit_total{status="ok"}[5m]) / sum(rate(devhub_onboarding_submit_total[5m]))`
- **S3 admin review latency** — submit 시점 vs confirm 시점 차이는 DB query (§4.3 SQL #5) 로 측정 (metric 도입은 별도 carve)
- **S4 CHECK constraint violation** — `devhub_onboarding_submit_total{status="server_error"}` + backend log grep (server_error 분기 = DB constraint 또는 internal)
- **S5 token-only actor baseline** — submit Counter 의 `actor.subject` 라벨 없음 (cardinality 회피) → SQL #1 그대로 사용

backend log + SQL + metric 3 가지 채널 cross-validation 권장 (단일 채널 실패 시 fallback).

## 5. Rollback runbook

### 5.1 Rollback trigger 4 케이스

| Trigger | 신호 | 임계 |
| --- | --- | --- |
| T1 — Gate 과차단 | S1 (403 spike) | baseline × 2 가 1시간 지속, 신규 사용자가 진입 못함 |
| T2 — 제출 회귀 | S2 (성공률 저하) | 1시간 평균 < 90% 또는 5xx 비율 ≥ 10% |
| T3 — Frontend redirect loop | 사용자 보고 + AuthGuard 회귀 | redirect loop 2건 이상 동시 보고 |
| T4 — DB 정합 깨짐 | S4 (CHECK 위반) | 1건 발견 즉시 |

### 5.2 Rollback 절차 4 step

> 운영자 (SRE) 권한으로 수행. 사내 IdP 팀 / backend 담당자 알림 후 진행.

1. **환경변수 변경**: deploy 환경 파일 (`docs/setup/deploy.env.example` 파생) 의 `DEVHUB_ONBOARDING_GATE_ENABLED` 를 `0` 으로 set.
2. **backend 재기동**: docker 환경이면 `docker compose restart backend-core`. native 환경이면 systemd / supervisor 단위 restart. 재기동 후 `curl http://localhost:8080/health` 로 health check.
3. **rollback 확인**:
   - `POST /api/v1/me/onboarding` → 404 `onboarding_feature_disabled` (signal: rollback 활성)
   - `GET /api/v1/applications` (보호 endpoint) 를 미완료 사용자 token 으로 호출 → 200 (gate 풀림 신호. 진단용. 정상 사용자 흐름엔 영향 없음)
4. **영향 사용자 인벤토리**:

   ```sql
   -- rollback 기간 동안 신규 token-only actor 가 보호 endpoint 에 접근
   SELECT COUNT(*) FROM users WHERE onboarding_completed_at IS NULL;
   SELECT COUNT(*) FROM users WHERE review_status = 'pending_review';
   ```

   변동량을 rollback 전후로 비교하여 보호 풀림 동안의 영향 추적.

### 5.3 Rollback 시 영향 (사용자 + 운영)

| 영역 | 영향 |
| --- | --- |
| 보안 | onboardingGate 가 no-op → 미완료 사용자가 모든 보호 endpoint 접근 가능. token-only actor 가 `Role` claim 만으로 정상 사용자처럼 동작. **24시간 내 root cause fix + flag 복원 권장**. |
| Frontend onboarding 페이지 | `POST /me/onboarding` → 404 → 페이지에서 에러 노출. 신규 사용자는 onboarding 진입 불가. |
| 기존 `reviewed` 사용자 | 영향 없음 (정상 사용 지속) |
| Admin pending_review 검토 | `POST /admin/users/:id/review` → 404 → 검토 작업 불가. rollback 동안 누적된 pending_review 는 flag 복원 후 일괄 처리. |
| Audit 3종 | 신규 emit 중단 → 운영 데이터 공백. rollback 동안의 unit 변경 / 검토 / 제출 추적 불가. |
| Keycloak event listener | 영향 없음 — 별도 layer ([keycloak_operations.md §8.6](./keycloak_operations.md#86-keycloak-event-listener-audit_logs-통합-운영-sop)) |

### 5.4 복원 (flag 재활성화)

1. Root cause fix 가 backend / frontend PR 으로 머지된 후 staging 재배포.
2. `DEVHUB_ONBOARDING_GATE_ENABLED=1` (또는 unset 으로 default true) set + backend 재기동.
3. `curl -H 'Authorization: Bearer <admin-token>' http://localhost:8080/api/v1/admin/users/<test-user>/review` 가 200 또는 422 (정상 응답) 반환 확인 — 404 면 flag 복원 실패.
4. §4.3 SQL #6 (CHECK 위반) 재검증 — rollback 기간 동안 정합 깨지지 않았는지 확인.
5. rollback 기간 동안 unit 변경한 사용자가 있다면 (audit 공백 구간) `review_status` 수동 검증 — 본인이 변경한 unit 으로 `review_status='pending_review'` 진입했어야 정상.

### 5.5 Rollback drill (1주 monitoring 중 1회 권장)

운영 사고 발생 전 staging 환경에서 §5.2 절차 drill 1회 수행 — DoD §7 #6. drill 시간은 사용자 영향 최소화를 위해 사내 정책상 off-peak 시간대 권장.

## 6. Incident response — 회귀 발견 시 단계화

### 6.1 발견 → 격리 → 진단 → 복구 4 단계

| 단계 | 시간 목표 | 행동 |
| --- | --- | --- |
| 1. 발견 | < 5분 | §4.2 신호 임계 도달 → on-call 알림 (사내 채널) |
| 2. 격리 | < 15분 | §5.2 rollback 절차 — gate 풀어 사용자 영향 차단 후 root cause 진단 시간 확보 |
| 3. 진단 | < 4시간 | §6.2 패턴 분류 → backend / frontend / IdP / DB 영역 결정 → §6.3 escalation |
| 4. 복구 | < 24시간 | bug fix PR 머지 → staging 검증 → §5.4 flag 복원 |

### 6.2 흔한 incident pattern 5종 + 1차 대응

| # | Pattern | 1차 의심 | 검증 / 대응 |
| --- | --- | --- | --- |
| P1 | 신규 사용자가 onboarding 페이지에 진입 못함 (또는 무한 redirect) | Frontend AuthGuard 회귀 — PR #291 blocklist 패턴 깨짐 | `/onboarding` 직접 호출 결과 / `feedback_e2e_oidc_flaky` memory case 참조. 우선 §5.2 rollback. backend 담당자 (frontend AuthGuard 영역은 [`frontend/components/layout/AuthGuard.tsx`](../../frontend/components/layout/AuthGuard.tsx)) escalate. |
| P2 | pending_review 누적량 가파른 증가 (S3) | system_admin 검토 missing | §4.3 SQL #2 로 oldest pending list 확보 → system_admin 일괄 검토 요청. Rollback 불필요 — 단순 운영 backlog. |
| P3 | 동일 사용자 onboarding 반복 제출 (S5) | session race 또는 frontend 회귀 (중복 submit) | §4.3 SQL #5 로 user_id 확보 → backend log 의 해당 request_id 추적. handler 의 `ErrConflict` 분기 정상 동작 시 409 응답 — frontend 가 409 를 잘못 처리하면 retry loop 가능. frontend escalate. |
| P4 | CHECK 제약 위반 (S4) | 수동 SQL UPDATE 잔재 또는 마이그레이션 충돌 | §4.3 SQL #6 로 위반 row id 확보 → audit_logs 의 동일 user_id action grep → 수동 수정 흔적 확인. 위반 row 의 `review_status` 를 `onboarding_completed_at` IS NULL 에 맞춰 NULL 로 정합. **DB 직접 UPDATE 는 backend 동작 정합 시 audit emit 우회** — 항상 audit 후행 INSERT 동반. |
| P5 | token-only actor 비율 비정상 증가 (S 보조) | Keycloak ↔ DevHub idp_subject 동기화 지연 | event listener cursor lag 점검 ([keycloak_operations.md §8.6.5](./keycloak_operations.md#86-keycloak-event-listener-audit_logs-통합-운영-sop)). lag 회복되면 자동 복구 — onboarding 도메인 자체는 정상. |

### 6.3 Escalation path

| Level | 담당 | 영역 |
| --- | --- | --- |
| L1 | 운영자 (SRE / on-call) | §5 rollback 자체 실행, §4.3 SQL 진단 |
| L2 backend | Claude | `internal/httpapi/{onboarding_gate,me_onboarding,users_admin_review,me,auth}.go` 회귀 |
| L2 frontend | Gemini | `frontend/components/layout/AuthGuard.tsx` + `frontend/app/onboarding/page.tsx` + `frontend/components/onboarding/*` 회귀 |
| L3 | 사내 IdP 팀 | Keycloak claim / event listener / realm 정의 측면 |
| L3 | DBA | migration 000033 정합 / CHECK 제약 검증 / index 손상 |

## 7. 1주 종료 후 acceptance — DoD

[`docs/planning/onboarding_impl_plan.md §7 #6`](../planning/onboarding_impl_plan.md) 의 *"운영 환경 (staging) 1주 monitoring — 403 spike / sessionStorage flag race 등 회귀 없음"* 항목의 구체화.

| # | 항목 | Verification |
| --- | --- | --- |
| 1 | 7일간 rollback 실행 0회 (계획된 drill §5.5 제외) | env history / git log |
| 2 | `POST /me/onboarding` 성공률 ≥ 95% | §4.3 SQL #4 |
| 3 | `pending_review` 평균 검토 latency 측정값 기록 (사내 SLA 결정 input) | §4.3 SQL #2 + #3 |
| 4 | CHECK 위반 0건 | §4.3 SQL #6 |
| 5 | audit event 3종 emit 정합 (반복 제출 0건) | §4.3 SQL #5 |
| 6 | Rollback drill 1회 staging 수행 | 운영 log + 본 SOP §5.5 |
| 7 | token-only actor 비율 정상 baseline 기록 | §4.3 SQL #7 |
| 8 | 1주 종합 보고서 작성 (사내 운영 회의용) | 별도 산출물 |

DoD #1~#7 모두 통과 시 staging → prod promote 검토. 단 1개 실패해도 root cause 해소 후 1주 재시작 — 부분 통과 promote 금지.

## 8. 잔여 carve out

본 SOP 의 scope 외 — 별도 sprint:

| 항목 | 우선순위 | 비고 |
| --- | --- | --- |
| ~~Prometheus metric backend 도입 — gate 403 Counter / submit Histogram / pending_review Gauge~~ | ~~P2~~ ✅ resolved (2026-05-26, sprint `claude/work_260526-onboarding-prometheus-metric`) | 4 metric 도입 (`devhub_onboarding_gate_blocked_total{reason}` Counter / `devhub_onboarding_submit_total{status}` Counter / `devhub_onboarding_submit_duration_seconds` Histogram / `devhub_onboarding_review_confirm_total{status}` Counter). `backend-core/internal/httpapi/onboarding_metrics.go` 신규 + `audit/metrics.go` 패턴 정합 ([ADR-0019 §5.3 (9)](../adr/0019-keycloak-only-idp.md#53-잔여-carve-out) Phase 2 PR-C). pending_review Gauge 는 별도 carve (DB SELECT COUNT 부담 + cron refresh 패턴 결정 후). |
| `pending_review` count Gauge (별도 carve) | P3 | submit Counter 의 `status="ok"` 누적 + admin confirm Counter 의 `status="ok"` 차감으로 derived metric 표현 가능 — 별도 Gauge 도입 시 DB SELECT COUNT 부담. cron refresh 패턴 결정 후. |
| Grafana dashboard JSON | P3 | metric 도입 후 — `docs/setup/grafana/` 패턴. 환경 별 자산이라 git 추적 외. |
| Alertmanager rule YAML | P3 | metric 도입 후 — S1~S5 임계의 정식 자산화. 환경 별 자산이라 git 추적 외. |
| `pending_review` admin 검토 SLA 정책 | P2 | 사내 정책 결정 — 24h / 48h / 72h. 정책 결정 후 §4.2 S3 임계 구체화. |
| `pending_review` aging alert | P3 | SLA 정책 + Prometheus metric 후 — `pending_review` row 가 SLA 초과 시 admin 알림. |
| Onboarding `limited` 사용자 reminder 알림 | P3 | 매 로그인 시 강제 진입이 사실상의 reminder 인데, 일정 기간 미로그인 사용자에게 별도 채널 (이메일) reminder 발송. |
| Multi-region monitoring | v1.1+ | staging / prod / DR 환경 별 신호 통합. |

## 9. 변경 이력

| 일자 | 변경 | sprint |
| --- | --- | --- |
| 2026-05-22 | 1차 draft — §1 책임 분리 + §2 state machine 운영 관점 + §3 feature flag 운영 + §4 1주 monitoring 신호 5종 + SQL 7개 + §5 rollback runbook 4 step + drill + §6 incident response 4 단계 + 패턴 5종 + escalation 3 level + §7 DoD 8 항목 + §8 잔여 carve 7. [ADR-0021](../adr/0021-onboarding-self-service-unit-selection.md) + [Onboarding IMPL plan §7 #6](../planning/onboarding_impl_plan.md) 의 운영 측면 source-of-truth. | `claude/work_260522-onboarding-ops-sop` |
| 2026-05-26 | §4.4 Prometheus metric backend 적용 (4 metric: gate_blocked Counter / submit Counter + Histogram / review_confirm Counter). §8 carve P2 resolved + pending_review Gauge 는 별도 carve P3 로 분리. metric 정의 [`backend-core/internal/httpapi/onboarding_metrics.go`](../../backend-core/internal/httpapi/onboarding_metrics.go). dashboard 매핑 표 (S1~S5 ↔ metric). | `claude/work_260526-onboarding-prometheus-metric` |
