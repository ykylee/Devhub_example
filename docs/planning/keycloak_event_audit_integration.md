# Design 검토 — Keycloak event → DevHub `audit_logs` 통합

- 문서 목적: [ADR-0019 §5.3](../adr/0019-keycloak-only-idp.md#53-잔여-carve-out) 잔여 carve out 의 audit event listener / SPI → DevHub `audit_logs` 통합 design. 1차 산출물은 planning 단계 — 결정 후 ADR-0020 으로 승격은 별도 sprint.
- 범위: Keycloak event (사용자 / admin) 를 DevHub `audit_logs` 에 통합하는 옵션 비교 + 권장 + audit_logs 매핑 + 구현 단계. 실 backend 구현 (cron + Keycloak Admin Client 확장 + 단위/integration 테스트) 은 별도 후속 sprint.
- 대상 독자: 아키텍트, 운영자 (SRE / IdP), Security, Backend 담당자.
- 상태: planning (draft 1차)
- 최종 수정일: 2026-05-19
- 결정 근거 sprint: `claude/work_260519-e`
- 관련 문서: [ADR-0019 Keycloak 단일화](../adr/0019-keycloak-only-idp.md), [ADR-0008 HRDB production adapter](../adr/0008-hrdb-production-adapter.md), [ADR-0015 HomeLab pull strategy](../adr/0015-homelab-adapter-pull-strategy.md) (cron + adapter 패턴 reference), [ADR-0017 DREQ intake token cron](../adr/0017-dreq-intake-token-operational-hardening.md) (cron + metric 패턴 reference), [keycloak_operations.md](../setup/keycloak_operations.md), [backend_api_contract.md §11](../backend_api_contract.md#11-계정-및-인증-account--auth).

## 1. 컨텍스트 + 동기

### 1.1 historical context

[ADR-0001](../adr/0001-idp-selection.md) 시점에는 **Kratos webhook → DevHub `audit_logs`** 통합 패턴 (M2 PR-M2-AUDIT) 으로 인증 이벤트가 audit_logs 에 기록됐다. [ADR-0019 Keycloak 단일화](../adr/0019-keycloak-only-idp.md) (PR #167 KC-PR-D, 2026-05-18) 로 Kratos webhook 자체가 제거되면서 인증 이벤트 audit 통합 패턴이 사라진 상태.

[ADR-0019 §4.5 audit_logs 영향](../adr/0019-keycloak-only-idp.md#45-audit_logs-영향) 가 명시:

> - Kratos webhook 기반 audit 흐름 (PR-M2-AUDIT, ADR-0001 §7 의 결정) 은 제거됨
> - Keycloak 이벤트는 별도 SOP 로 후속 통합 (carve out, §5.3 참조)
> - `audit_logs.actor_login` 은 Keycloak `preferred_username` claim 으로 매핑

### 1.2 통합 필요성

DevHub `audit_logs` 의 actor enrichment + request_id 정책 ([M1 PR-D](../../ai-workflow/memory/state.json)) 은 도메인 액션 (application / project / dev_request 등) 에 대해서는 그대로 동작. 그러나 **인증 / 계정 / 세션 lifecycle 이벤트** (login / logout / signup / password change / lockout / admin user create-update-delete) 가 DevHub `audit_logs` 에서 누락 — Keycloak 이 인증 책임이지만 audit 통합 안 됨.

운영 측면 필요성:
- **regulatory 요구**: 사내 보안 감사에서 인증 이벤트도 audit trail 요구 (예: ISO 27001 / SOC 2)
- **incident response**: 의심 활동 (실패 로그인 반복 / 권한 변경 등) 의 도메인 액션 + 인증 이벤트 통합 timeline 필요
- **개인정보 access log**: 사용자 본인 인증 history 조회 요구

## 2. 통합 옵션 비교 (3종)

| 옵션 | 변경 범위 | 운영 부담 | 보안 검증 | 권장 |
| --- | --- | --- | --- | --- |
| **A. Keycloak Event Listener SPI** (Java plugin) | 큼 — Keycloak deploy 측에 custom JVM plugin (Java) 작성 + 빌드/배포 stack + DevHub backend 의 `/api/v1/internal/keycloak-event` endpoint 신설 + auth (서명 verify 또는 mTLS) | Keycloak deploy stack 유지보수 + plugin 업데이트 | event listener SPI 내부 동작 검증 + endpoint 인증 | ❌ 사내 Keycloak deploy 측 JVM 빌드/배포 능력 의존 + plugin 별도 lifecycle |
| **B. Admin event polling** (Go cron) | 중간 — DevHub backend 에 cron worker 추가 + Keycloak Admin Client 의 `/admin/realms/{realm}/events` + `/admin/realms/{realm}/events/admin` 호출 + dedup state + audit_logs INSERT | DevHub backend 만 변경 — Keycloak deploy 측 무영향 | Keycloak Admin API (이미 KC-PR-C 도입) + polling state 의 정합성 | ⭐ **권장** — DevHub git 안에서 자기완결 + Go-native cron 패턴 (ADR-0015 HomeLab pull / ADR-0017 intake token cron 정합) |
| **C. Webhook hybrid** (Event Listener SPI + HTTP webhook) | A 와 동일 (Keycloak 측 plugin) + B 의 endpoint 측면 일부 | Keycloak deploy 측 plugin + DevHub endpoint 둘 다 | A 와 동일 | △ A 의 부담 + endpoint 보안만 추가 — A 의 단점 그대로 |

## 3. 옵션 B (권장) — admin event polling 상세

### 3.1 Keycloak admin event endpoint

Keycloak Admin API 가 2 종류의 event 제공 (codex review #9 검증 — Keycloak Admin REST API 표준 path):

| Endpoint | 내용 | 권장 query |
| --- | --- | --- |
| `GET /admin/realms/{realm}/events` | 사용자 event (LOGIN / LOGOUT / REGISTER / UPDATE_PASSWORD / IDENTITY_PROVIDER_LINK 등) | `?dateFrom=<ISO8601>&max=500` (paged) |
| `GET /admin/realms/{realm}/admin-events` | admin event (USER:CREATE / USER:UPDATE / USER:DELETE / ROLE:CREATE 등) | `?dateFrom=<ISO8601>&max=500` |

**Keycloak 측 사전 활성화** (운영 SOP):
- Realm settings → Events → Login Events Settings → "Save Events" ON + Event Listeners 에 `jboss-logging` 외 `metrics-listener` (Keycloak 25+) 또는 default 활성
- Admin Events Settings → "Save Events" ON + "Include Representation" OFF (PII 측면, 별도 검토)
- Expiration: `7 days` (DevHub polling 주기 가정 + retention)

### 3.2 cron worker 패턴 (ADR-0015 / ADR-0017 정합)

```text
internal/audit/
├── keycloak_event_puller.go        // 본 design 의 신규
├── keycloak_event_puller_test.go
└── metrics.go                       // (선택) Prometheus counter

internal/audit/keycloak_event_puller.go:
  type KeycloakEventPuller struct {
      adminClient   *httpapi.KeycloakAdminClient
      store         AuditStore
      stateStore    EventCursorStore       // 마지막 처리 timestamp 영구화
      interval      time.Duration          // env DEVHUB_KEYCLOAK_EVENT_POLL_INTERVAL (default 30s)
      maxEvents     int                    // env DEVHUB_KEYCLOAK_EVENT_PULL_MAX (default 500)
      backoff       BackoffConfig          // exponential backoff on Keycloak error
  }

  func (p *KeycloakEventPuller) Run(ctx context.Context) error {
      ticker := time.NewTicker(p.interval)
      for {
          select {
          case <-ctx.Done(): return ctx.Err()
          case <-ticker.C:
              if err := p.pullOnce(ctx); err != nil {
                  // exponential backoff
              }
          }
      }
  }
```

main.go wire (ADR-0017 intake token cron wire 패턴 정합):
```go
if cfg.KeycloakEventPullEnabled {
    puller := audit.NewKeycloakEventPuller(...)
    go puller.Run(ctx)
}
```

### 3.3 state 영구화 (dedup)

`event_cursor` 테이블 신규 (migration 000022 후보):

```sql
CREATE TABLE event_cursors (
    cursor_key   TEXT PRIMARY KEY,         -- "keycloak.events" / "keycloak.events.admin"
    last_event_at TIMESTAMPTZ NOT NULL,
    last_event_id TEXT,                    -- Keycloak event 의 hash (옵션 — id 컬럼이 없으면 timestamp + type + userId 의 SHA256)
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

- 매 poll 시 `last_event_at` 이후의 event 만 처리 → dedup
- Keycloak event 가 id 없음 — `(timestamp, type, userId)` 조합으로 응용 dedup hash
- crash recovery 시 마지막 cursor 부터 재개 — at-least-once delivery (audit_logs 의 UNIQUE 제약으로 중복 INSERT 방지 carve)

### 3.4 audit_logs INSERT 매핑

audit_logs schema (10 columns):

| audit_logs 컬럼 | Keycloak event → 매핑 |
| --- | --- |
| `audit_id` | gen `aud_<hex24>` (M1 PR-D 패턴 정합) |
| `actor_login` | event `userId` → Keycloak admin lookup `preferred_username` (또는 event 의 `username` 직접) |
| `action` | event type 별 매핑 (§4 매핑 표) |
| `target_type` | `auth` / `user` / `role` (event 분류 별) |
| `target_id` | event `userId` 또는 admin event 의 `resourcePath` |
| `command_id` | NULL (Keycloak 이벤트는 command 와 무관) |
| `payload` | event raw JSON minimal subset (PII 최소화 — IP / userAgent / clientId / sessionId 등) |
| `source_ip` | event `ipAddress` |
| `request_id` | NULL (Keycloak 이 발급 X) 또는 puller 가 batch UUID 발급 (`req_keycloak_<batchUUID>`) |
| `source_type` | `keycloak_event` (신규 enum 값 — schema 정합 carve) |

### 3.5 폐기 — 보안 책임 분리

- Keycloak Admin Client 의 service account (KC-PR-C, `devhub-backend` confidential client) 가 이미 `realm-management` 의 `view-users` / `manage-users` role 보유
- event endpoint 접근 추가 role 필요 — `view-events` (realm-management) 추가 → keycloak_operations.md §3.2 갱신 carve
- secret rotation = §8.3 동일 SOP

## 4. audit_logs `action` 매핑 표

### 4.1 사용자 event (login events)

| Keycloak event type | DevHub action | target_type |
| --- | --- | --- |
| `LOGIN` | `auth.login.success` | `auth` |
| `LOGIN_ERROR` | `auth.login.failed` | `auth` |
| `LOGOUT` | `auth.logout.success` | `auth` |
| `LOGOUT_ERROR` | `auth.logout.failed` | `auth` |
| `REGISTER` | `auth.signup.success` | `user` |
| `REGISTER_ERROR` | `auth.signup.failed` | `user` |
| `UPDATE_PASSWORD` | `auth.password.changed` | `user` |
| `UPDATE_PASSWORD_ERROR` | `auth.password.change_failed` | `user` |
| `SEND_RESET_PASSWORD` | `auth.password.reset_requested` | `user` |
| `RESET_PASSWORD` | `auth.password.reset_success` | `user` |
| `REFRESH_TOKEN` | (skip — 너무 빈번, carve) | — |
| `IDENTITY_PROVIDER_LINK_ACCOUNT` | `auth.idp.linked` | `user` |
| `IDENTITY_PROVIDER_FIRST_LOGIN` | `auth.idp.first_login` | `user` |
| `VERIFY_EMAIL` | `auth.email.verified` | `user` |
| `REMOVE_TOTP` / `UPDATE_TOTP` | `auth.mfa.totp_{removed,updated}` | `user` |

### 4.2 admin event

| Keycloak admin operation (resourceType:operationType) | DevHub action | target_type |
| --- | --- | --- |
| `USER:CREATE` | `keycloak.user.created` | `user` |
| `USER:UPDATE` | `keycloak.user.updated` | `user` |
| `USER:DELETE` | `keycloak.user.deleted` | `user` |
| `USER:ACTION` (logout / disable) | `keycloak.user.action.<detail>` | `user` |
| `REALM_ROLE_MAPPING:CREATE` / `DELETE` | `keycloak.user.role.{granted,revoked}` | `user` |
| `CLIENT:UPDATE` | `keycloak.client.updated` | `client` |
| `REALM:UPDATE` | `keycloak.realm.updated` | `realm` |

### 4.3 skip 패턴 (PII / 잡음)

- `REFRESH_TOKEN` — 너무 빈번 (1분에 수십회) + 단순 token 갱신
- `CODE_TO_TOKEN` — OIDC 표준 step, audit 가치 낮음
- `INTROSPECT_TOKEN` — backend 검증 단계 (KC-PR-B), audit 가치 낮음

skip event 는 puller 의 filter list 로 명시 — 운영자가 환경 변수 (`DEVHUB_KEYCLOAK_EVENT_SKIP_TYPES`) 로 override 가능.

## 5. 구현 단계 (별도 sprint)

### 5.1 PR-A: design accepted + ADR-0020 승격 (선택)

- 본 design 의 옵션 B 채택 + ADR-0020 신규 발행 (또는 ADR-0019 §5.3 carve resolved 만으로 한정)
- design 검토 후 사용자 결정 (별도 ADR 발행 가치 평가)

### 5.2 PR-B: cron worker + cursor + audit_logs INSERT 골격

- `internal/audit/keycloak_event_puller.go` 신규
- `internal/store/event_cursors.go` + migration 000022 신규
- `KeycloakAdminClient` 확장 — `ListUserEvents` + `ListAdminEvents` 메소드
- env 추가 — `DEVHUB_KEYCLOAK_EVENT_PULL_ENABLED` (default false) + 관련 6 env (interval / max / skip types 등)
- 단위 테스트 — fake admin client + mock event payload

### 5.3 PR-C: main.go wire + integration test + Prometheus metric

- main.go env-gated cron 시작 (ADR-0017 intake token cron 패턴 정합)
- integration test (DEVHUB_TEST_DB_URL 환경) — event_cursors 영구화 + audit_logs INSERT 회귀 시나리오 (Happy / Skip / Crash recovery)
- Prometheus counter — `devhub_keycloak_events_processed_total{type, status}` + gauge `devhub_keycloak_event_cursor_lag_seconds`

### 5.4 PR-D: audit_logs schema 정합

- migration 000023 — `audit_logs.source_type` CHECK 제약에 `'keycloak_event'` 추가 (또는 NOT-NULL 제약 외 enum-like 처리)
- backend_api_contract §11 / docs/architecture §6 audit 절 갱신

### 5.5 PR-E: 운영 SOP

- `docs/setup/keycloak_operations.md` §8.6 (또는 §6 추가) — event puller 운영 가이드 + cursor 진단 + skip types 운영 SOP
- ADR-0019 §5.3 (9) audit event listener carve 실 구현 resolved 마킹

## 6. 보안 점검

### 6.1 잠재 위협 + mitigation

| 위협 | mitigation |
| --- | --- |
| Keycloak admin client 권한 과대 | `view-events` 만 추가, write 권한 추가 안 함 (admin event 는 read-only) |
| event payload PII 노출 (email / IP / userAgent) | payload 의 minimal subset 만 저장 — `INCLUDE_REPRESENTATION` OFF + 필드 allowlist (clientId / sessionId / ipAddress / userId) |
| audit_logs 의 dedup 실패 (duplicate INSERT) | event_cursor `last_event_at` 영구화 + 응용 hash `(timestamp, type, userId)` UNIQUE 제약 (carve — migration §5.4) |
| Keycloak event endpoint 응답 변경 | 단위 테스트 + integration test 의 contract 검증. Keycloak 버전 업그레이드 시 회귀 검증. |
| puller crash + cursor drift | last_event_at 영구화 + crash recovery 자동 재개. backoff exponential. |

### 6.2 audit_logs 의 `actor_login` enrichment

- Keycloak event 의 `userId` (UUID) 만 노출 → `actor_login` 으로 mapping 시 별도 Keycloak admin lookup `preferred_username` 필요
- 빈도 ↑ — admin lookup batch 또는 cache (Keycloak admin client 의 user info cache)
- cache miss / lookup 실패 시 `actor_login = "keycloak:<userId>"` fallback

## 7. cutover 절차

### 7.1 Phase 1 (본 sprint) — design only

- ✅ 본 문서 작성
- ADR-0019 §5.3 audit event listener carve resolved (design 완료) 마킹

### 7.2 Phase 2 — backend 구현 (별도 sprint -X)

- §5.2 PR-B (cron worker + cursor)
- §5.3 PR-C (main.go wire + integration test + metric)
- §5.4 PR-D (audit_logs schema 정합)
- §5.5 PR-E (운영 SOP)

### 7.3 Phase 3 — staging 검증 (별도 sprint -Y)

- staging 환경에서 1주 polling
- 처리량 / cursor lag / 잘못된 매핑 검수
- payload PII 검수 (사내 보안팀 동반)

### 7.4 Phase 4 — prod cutover

- 사내 보안팀 + Keycloak 운영팀 공동 cutover
- `DEVHUB_KEYCLOAK_EVENT_PULL_ENABLED=true`
- 모니터링 metric 확인 (cursor lag < interval)

## 8. 잔여 carve out / open question

- **(carve)** ADR-0020 승격 — design 결정 (옵션 B polling) 의 ADR governance 측면 명문화. Phase 2 진입 시 결정. ADR-0019 §5.3 carve resolved 만으로 한정해도 됨.
- **(carve)** event payload 의 PII 필드 allowlist 확정 — 사내 보안팀 검토 후 결정 (default suggestion: clientId / sessionId / ipAddress / userId 만)
- **(carve)** `audit_logs.source_type` enum 확정 — 현재 schema 가 enum 강제 아님 (실 INSERT 패턴 reference). enum 강제 결정 시 migration carve.
- **(carve)** Keycloak event 의 dedup hash — `(timestamp, type, userId)` 충돌 가능성 (동시 다발 event). 추가 nonce 또는 Keycloak 이 발급한 임시 sequence 활용 carve.
- **(open)** Keycloak admin lookup 의 cache 정책 — `preferred_username` resolve 빈도 ↑ 시 별도 read-through cache 필요 가능성 (Keycloak 25+ 의 `getUserByEvent` 가능 여부 확인 carve).
- **(open)** Phase 2 staging 진입 시점 — 사내 보안팀 + Keycloak 운영팀 동반 필요. 본 design 만 1차 진입.
- **(carve)** Event Listener SPI (옵션 A) 의 future migration path — 옵션 B 가 polling 의 latency 한계 (interval = 30s default) 가 운영 SLA 와 충돌 시 SPI 로 전환 가능. 본 sprint 는 polling 채택.

## 9. 결정 후보 (Phase 2 진입 시 ADR-0020 또는 ADR-0019 §5.3 resolved 만)

본 design 이 Phase 2 진입 시 ADR-0020 후보:

**ADR-0020**: Keycloak event → DevHub `audit_logs` 통합 정책 (옵션 B admin event polling 채택, Go-native cron, event_cursors 영구화, payload PII minimal allowlist, action 매핑 표 §4, source_type=keycloak_event 추가)

ADR §3 검토 옵션 표 + §4 결정 (옵션 B) + §5 결과 + §6 carve out 은 본 §2/§3/§5/§8 을 그대로 승격.

또는 ADR governance 단순화 차원에서 본 design 만으로 carve resolved 마킹 + 실 구현 sprint 가 ADR-0019 §5.3 의 carve 잔여 1 항목 종결 — 별도 ADR 발행 안 함.

## 10. 변경 이력

| 일자 | 변경 | sprint |
| --- | --- | --- |
| 2026-05-19 | 1차 draft — 10 section + 옵션 3종 비교 + 권장 (옵션 B polling) + audit_logs action 매핑 30 row (사용자 event 15 + admin event 7 + skip 3 + 정합 5) + 구현 단계 PR-A..E + 보안 점검 + cutover Phase 1..4 + carve out 7 항목 + ADR-0020 후보. | `claude/work_260519-e` |
