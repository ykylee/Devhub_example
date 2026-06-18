---
title: offboarding_immediacy
type: source
tags: [infrastructure, offboarding_immediacy.md, project-devhub]
sources: [raw/projects/devhub/docs/infrastructure/keycloak-idp/offboarding_immediacy.md]
git_commit: 01f1969c
git_branch: main
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
last_touched: 2026-06-18T07:11:15Z
mirror_dirty: |
related: [none]
status: draft
contradictions: [none]
---

# Design 검토 — Off-boarding 즉시성 (HR ↔ Keycloak ↔ DevHub propagation chain)

> ⚠ **2026-05-20 결정 변경 (issue #215 cancelled, sprint claude/work_260520-q-215-hrdb-cancel)**: 사용자가 DevHub 외부 Keycloak 시나리오 채택 — 사내 IdP 팀이 별도 운영. **HR ↔ Keycloak sync 책임이 외부 IdP 팀 (Keycloak User Federation 또는 사내 ETL → Keycloak Admin REST) 로 이관**. 본 design 의 **Phase 1 cron (§3.1 옵션 C)** 은 사용자 결정으로 **폐기**. 대체 흐름:
> 1. 외부 IdP 팀이 HR 'terminated' → Keycloak user disable (Keycloak admin console 또는 사내 ETL)
> 2. DevHub backend 의 Keycloak event listener (sub-carve C, PR #241) 가 admin event polling — USER:UPDATE event 감지
> 3. `user_sync.go::SyncUserProfile` 가 DevHub `users.status='deactivated'` 자동 sync
>
> 따라서 본 design 의 Phase 1 cron 부분 + `scripts/hrdb_etl_sync.sh` (sprint -p PR #184) 는 **historical reference 로 보존**. Phase 2 LDAP/AD federation 또는 외부 Keycloak provider 가 운영 시나리오. issue #215 close (not planned).

- 문서 목적: [ADR-0019 §5.3](../adr/0019-keycloak-only-idp.md#53-잔여-carve-out) 잔여 carve out 의 off-boarding 즉시성 design + [ADR-0008 §6](../adr/0008-hrdb-production-adapter.md#6-미해결-항목-open-questions) 의 "실시간 sync 요구 시 별도 ADR" 항목 통합. 1차 산출물은 planning 단계 — Phase 2 LDAP/AD federation 진입 시 별도 ADR 승격 결정.
- 범위: HR 시스템에서 사용자 비활성화 (퇴사 / 부서 이동 / 권한 회수) 가 Keycloak + DevHub 까지 전파되는 chain 의 latency 최소화. SCIM bridge / 실시간 webhook 등 고급 통합은 Phase 3 carve.
- 대상 독자: 아키텍트, 운영자 (SRE / IdP), Security, HR 시스템 담당자, Backend 담당자.
- 상태: draft
- 최종 수정일: 2026-05-20
- 결정 근거 sprint: `claude/work_260519-g`
- 관련 문서: [ADR-0019 Keycloak 단일화](../adr/0019-keycloak-only-idp.md), [ADR-0008 HRDB production adapter](../adr/0008-hrdb-production-adapter.md), [ADR-0010 primary_dept resolution](../adr/0010-primary-dept-resolution.md), [keycloak_operations.md](../setup/keycloak_operations.md) §8.2, [keycloak_event_audit_integration.md](./keycloak_event_audit_integration.md) (admin event listener — off-boarding event 감지 활용).

## 1. 컨텍스트 + 동기

### 1.1 현재 chain (worst case)

```text
HR 시스템 (퇴사 처리)
  ↓ daily ETL cron (24h latency)
DevHub `hrdb` schema (users.status, ADR-0008)
  ↓ Keycloak admin 별도 작업 (수동, latency 가변)
Keycloak user.enabled = false
  ↓ access_token TTL (default 1h, keycloak_operations §6.2 권장 5분)
실 권한 차단 (새 token 발급 차단)
```

**worst case latency** (default 값 가정):
- HR ETL daily cron 주기 = 24h
- Keycloak admin 수동 작업 = 수 시간 ~ 1일 (운영 SLA 의존)
- access_token TTL = 1h (default) 또는 5분 (권장)
- **총 worst case ≈ 24h + N hours + 5분 ≈ 24-48h**

### 1.2 문제

- 사용자 퇴사 후에도 최대 1~2일 DevHub 접근 가능 — 정보 유출 / 권한 남용 위험
- HR → DevHub `hrdb` schema sync 와 HR → Keycloak sync 가 **독립** — 두 시스템 동기화 보장 안 됨 (예: HR ETL 만 실행, Keycloak 비활성화 누락)
- 운영 SLA 요구 (예: 1시간 이내 차단) 와 충돌

### 1.3 즉시성 향상 목표

- **Phase 1 (단기, 본 design)**: worst case **≤ 1h** — HR push 도입 + access_token TTL 단축
- **Phase 2 (중기)**: worst case **≤ 15분** — LDAP/AD federation 자동 sync
- **Phase 3 (장기, carve)**: worst case **≤ 1분** — SCIM bridge + 실시간 webhook

## 2. 통합 옵션 비교 (6종)

| 옵션 | 변경 범위 | DevHub backend 영향 | latency (worst case) | 권장 |
| --- | --- | --- | --- | --- |
| **A. 현행 — daily HR ETL** | 없음 | 없음 | 24-48h | △ 부족 |
| **B. Keycloak admin 직접 force logout** | Keycloak admin 작업 SOP 만 | 없음 | 수동 SLA (가변) | △ 운영 의존 |
| **C. HR ETL → Keycloak Admin API push** | ETL script 확장 (psql + Keycloak Admin REST 호출) | 없음 | ETL 주기 (단기 1h 권장) + token TTL (5분) ≈ 1h | ⭐ **Phase 1 권장** |
| **D. HR system → DevHub webhook** | HR system 측 webhook 도입 + DevHub backend `/internal/hr-event` endpoint | backend 변경 (신규 endpoint + auth + Keycloak admin client 호출) | 즉시 (~ 5분 token TTL) | △ HR system 측 변경 부담 |
| **E. LDAP/AD federation** | Keycloak User Federation 활성화 (사내 LDAP/AD 운영 중일 때) | 없음 | LDAP cache TTL (15분) + token TTL (5분) ≈ 20분 | ⭐ **Phase 2 권장** |
| **F. SCIM bridge (HR → Keycloak SCIM API)** | 사내 SCIM bridge 인프라 + Keycloak Identity Provider 측 SCIM 활성화 | 없음 | 실시간 (~ token TTL 5분) | △ Phase 3 carve — 인프라 부담 |

## 3. Phase 1 (단기 권장) — 옵션 C 상세

### 3.1 ETL script 확장

`scripts/hrdb_etl_seed.sql` 패턴 ([ADR-0008 §6](../adr/0008-hrdb-production-adapter.md#6-미해결-항목-open-questions)) 기반의 운영 ETL cron 확장. **실 구현**: [`scripts/hrdb_etl_sync.sh`](../../scripts/hrdb_etl_sync.sh) (sprint `claude/work_260519-p`, PR #184). 아래 skeleton 은 design 시점 sample — 실제 동작 script 는 sprint -p hotfix 이후 codex review #9 정정 (admin REST `/admin/realms/{realm}/...` + `users.status` 직접 UPDATE) 정합.

```bash
# scripts/hrdb_etl_sync.sh (신규 운영 script — 사내 cron 운영)
# 1. HR 시스템에서 active=false 사용자 list export
# 2. DevHub PG hrdb schema sync (UPSERT)
# 3. Keycloak Admin API 로 user disable + force logout

HR_DISABLED_USERS=$(hr_export_disabled.sh)  # 사내 HR export 도구

for user in $HR_DISABLED_USERS; do
  # Step 1: DevHub users 비활성화 (codex review #9 정정 — hrdb 의 actual schema 는 hrdb.persons + active 컬럼 없음.
  # ADR-0008 §3 / migration 000010_create_hrdb_persons.up.sql 정합. 비활성화는 hrdb 가 아닌 users.status 직접 UPDATE.)
  psql -c "UPDATE users SET status = 'deactivated', updated_at = NOW() WHERE system_id = '$user'"
  # (hrdb 가 인사 master 의 사본 — 퇴사자는 hrdb.persons row 가 HR 시스템에서 사라지면 daily ETL upsert 의 자연 결과로 미동기.
  # 즉 hrdb 는 활성 인사 의 사본만 보관. 회수 액션은 users.status 직접 UPDATE 가 자연.)

  # Step 2 (신규): Keycloak admin disable + force logout (codex review #9 정정 — Admin REST base path /admin/realms/...)
  KC_TOKEN=$(curl -s -X POST "$KC_ISSUER/protocol/openid-connect/token" \
    -d "grant_type=client_credentials&client_id=devhub-backend&client_secret=$KC_SECRET" | jq -r .access_token)

  KC_USER_ID=$(curl -s -H "Authorization: Bearer $KC_TOKEN" \
    "$KC_ADMIN/admin/realms/devhub/users?username=$user" | jq -r '.[0].id')

  # User disable
  curl -s -X PUT -H "Authorization: Bearer $KC_TOKEN" -H "Content-Type: application/json" \
    "$KC_ADMIN/admin/realms/devhub/users/$KC_USER_ID" \
    -d '{"enabled": false}'

  # Force logout (모든 active session 종료)
  curl -s -X POST -H "Authorization: Bearer $KC_TOKEN" \
    "$KC_ADMIN/admin/realms/devhub/users/$KC_USER_ID/logout"
done
```

> **codex review #9 정정 사항**:
> - Admin REST base path = `$KC_ADMIN/admin/realms/{realm}/...` (이전 표기 `$KC_ADMIN/realms/...` 부정확). [`keycloak_admin_client.go`](../../backend-core/internal/httpapi/keycloak_admin_client.go) 의 `base + "/admin/realms/" + realm` 패턴과 정합.
> - HRDB schema = `hrdb.persons` (no `active` column, migration 000010). 회수는 `users.status = 'deactivated'` 직접 UPDATE — `hrdb.employees` + `active = false` 패턴은 실 schema 와 불일치. 향후 hrdb.persons 의 비활성 컬럼 추가는 별도 migration carve.

**핵심 변경**:
- ETL cron 주기 = **daily → hourly** (Phase 1 latency 목표 1h)
- Keycloak Admin API 호출 추가 — user disable + force logout
- DevHub backend 변경 없음

### 3.2 access_token TTL 단축

[keycloak_operations.md §6.2](../setup/keycloak_operations.md) 권장 적용:

- Keycloak admin → Realm Settings → Tokens → "Access Token Lifespan" = **5 minutes**
- "SSO Session Idle" = 30 minutes (logout overlay UX 정합)
- 권장 운영: refresh_token = 12h, refresh_token rotation 활성화

5분 TTL 적용 시 worst case = (1h ETL 주기) + (5분 token TTL) = **최대 65분 ≈ 1h**

### 3.3 audit_logs 통합

[keycloak_event_audit_integration.md §4.2](./keycloak_event_audit_integration.md#42-admin-event) 의 admin event 매핑 활용 — `USER:UPDATE` (enabled=false) + `USER:ACTION` (force logout) 이 자동 audit_logs 로 통합 (Phase 2 event listener 구현 후).

Phase 1 단순 운영 audit = ETL script log + Keycloak admin event log (별도 검수).

### 3.4 운영 SOP (keycloak_operations.md §8.2 갱신 정합)

| 단계 | 위치 | 액션 |
| --- | --- | --- |
| 1. HR ETL cron 실행 | 사내 운영 cron (hourly) | scripts/hrdb_etl_sync.sh 실행 (DevHub hrdb sync + Keycloak disable + force logout) |
| 2. (실시간 긴급) | Keycloak admin console (수동) | Users → 해당 user → Details → Enabled = OFF + Sessions → "Logout all sessions" |
| 3. DevHub `users.status` sync | (자동) | 다음 hourly ETL 에서 갱신 |
| 4. audit 검수 | Keycloak admin event log | `USER:UPDATE` (enabled=false) + `USER:ACTION` (LOGOUT) 발생 확인 |

## 4. Phase 2 (중기 권장) — 옵션 E LDAP/AD federation

### 4.1 사용 시나리오

사내 LDAP / Active Directory 가 운영 중일 때 Keycloak 의 표준 User Federation 활성화:

- Keycloak admin → User Federation → Add provider → `ldap`
- 사내 LDAP/AD endpoint + Bind credentials + Search base + User attribute mapping (sAMAccountName / mail / displayName / employeeID)
- Cache policy: `DEFAULT` (Keycloak default) 또는 `EVICT_DAILY` (정확도 우선)

### 4.2 동작

- LDAP/AD 에서 user 비활성화 (직급 변경 / 퇴사) → Keycloak federation cache TTL 만료 후 자동 반영
- Keycloak cache TTL = default `NO_CACHE` (실시간) 또는 1h (성능 trade-off)
- `NO_CACHE` 적용 시 worst case latency = LDAP query time (~ ms) + token TTL (5분) ≈ **5분**

### 4.3 trade-off

| 측면 | LDAP federation | HR ETL push (Phase 1) |
| --- | --- | --- |
| latency | 5-20분 | 1시간 |
| 운영 부담 | LDAP/AD bind credentials + sync 정책 | ETL script 유지 |
| 의존성 | 사내 LDAP/AD 운영 중 가정 | 사내 HR 시스템 + 운영팀 cron |
| Keycloak 측 변경 | User Federation 활성화 + LDAP mapper | (변경 없음) |
| HR master 보존 | LDAP/AD master + Keycloak read-only sync | HR 시스템 master + Keycloak read-only |
| DevHub backend | 변경 없음 | 변경 없음 |

### 4.4 cutover

- staging 환경에서 LDAP federation 1주 검수 (특히 mapping + cache TTL)
- 사내 보안팀 + LDAP/AD 운영팀 동반
- HR ETL Phase 1 과 **공존 가능** — federation 이 primary, ETL 이 backup

## 5. Phase 3 (장기 carve) — SCIM bridge + 실시간 webhook

### 5.1 SCIM bridge

- HR 시스템 → SCIM 2.0 API → Keycloak Identity Provider
- 사내 SCIM bridge 미들웨어 운영 필요 (예: Okta SCIM connector, ForgeRock IDM, 자체 구현)
- Phase 3 carve — 사내 IdP 인프라 결정 동반

### 5.2 실시간 webhook (옵션 D 변형)

- HR 시스템에 user disable webhook 추가 → DevHub backend `/api/v1/internal/hr-event` 직접 수신
- DevHub backend 가 Keycloak Admin Client 로 즉시 disable + force logout
- backend 변경 큼 (신규 endpoint + auth + Keycloak Admin Client 확장 + idempotency + retry)
- Phase 3 carve — HR system 측 webhook 도입 의존

## 6. 보안 점검

### 6.1 잠재 위협 + mitigation

| 위협 | mitigation |
| --- | --- |
| ETL script 의 Keycloak admin client_secret 노출 | 사내 vault (keycloak_operations §8.3) + script 실행 환경의 secret 주입 (env var 또는 OS keyring) |
| ETL 실패 시 partial sync (DevHub hrdb sync 완료, Keycloak disable 미완) | script idempotent + retry + 실패 시 alert (사내 monitoring + Slack/Email) + 다음 cron 에서 재시도 |
| 잘못된 user disable (HR data error) | HR 시스템의 정확성 의존 + Keycloak admin event log 의 USER:UPDATE 추적 + 잘못 disable 시 admin 복구 SOP |
| force logout 으로 인한 active 작업 손실 (예: 작업 중인 명령) | 사용자 영향 — Keycloak admin force logout 은 강제 동작. 사내 정책으로 사전 공지 SOP (24h 전 공지) |
| Phase 2 LDAP federation 의 LDAP 가용성 의존 | LDAP/AD HA 구성 (사내 인프라) + Keycloak federation cache fallback |

### 6.2 audit log 통합 (sprint -e design 정합)

[keycloak_event_audit_integration.md](./keycloak_event_audit_integration.md) 의 admin event 매핑:
- `USER:UPDATE` (enabled=false) → `keycloak.user.updated` action + payload 의 `enabled` field 표기
- `USER:ACTION` (LOGOUT) → `keycloak.user.action.logout` action

본 design 의 운영 audit 가 sprint -e 의 audit event listener 구현 후 자동 통합.

## 7. cutover 절차

### 7.1 Phase 1 (단기, 본 sprint 후 별도 운영 sprint)

- ✅ 본 design 문서 작성
- access_token TTL 5분 적용 (Keycloak admin Realm Settings → Tokens)
- ETL cron 주기 daily → hourly + Keycloak Admin API 호출 추가
- 모니터링: ETL log + Keycloak admin event log

### 7.2 Phase 2 (중기, 사내 LDAP/AD 도입 시)

- LDAP/AD federation 검토 (사내 IdP 운영팀 동반)
- staging 1주 검수
- prod 도입 + HR ETL Phase 1 backup 유지

### 7.3 Phase 3 (장기, carve)

- SCIM bridge 도입 평가 (사내 IdP 인프라 결정)
- 실시간 webhook 도입 평가 (HR system 측 변경 동반)

## 8. ADR governance 결정

### 8.1 별도 ADR 발행 여부

본 design 은 ADR-0019 §5.3 carve + ADR-0008 §6 의 "실시간 sync" 항목 통합. ADR governance 측면:

- Phase 1 (옵션 C HR ETL push) = 운영 SOP 수준 — 별도 ADR 발행 가치 낮음
- Phase 2 (옵션 E LDAP federation) = Keycloak User Federation 도입 결정 — **별도 ADR 후보** (ADR-0021 은 Onboarding 으로 2026-05-21 발급됨 — Phase 2 진입 시점에 다음 번호 사용)
- Phase 3 (SCIM / webhook) = 별도 ADR (ADR-0022 또는 이후) 후보

**1차 결정**: ADR-0019 §5.3 (7) carve resolved (design) 만으로 한정. ADR-0008 §6 "실시간 sync" 항목에 design link 추가. Phase 2 진입 시 별도 ADR 재평가.

### 8.2 ADR-0008 §6 와의 정합

- ADR-0008 §6 의 "실시간 sync 요구 (퇴사자 즉시 차단) 시 별도 ADR" → 본 design 으로 1차 명문화 (별도 ADR 발행 보류, Phase 2 재평가)
- ADR-0008 §6 의 "ETL 책임 (수동 / cron / 외부 tool)" 부분 결정 → 본 design 의 §3.1 hourly cron 으로 명확화

## 9. 잔여 carve out / open question

- **(carve)** Phase 1 hourly ETL cron 의 실 운영 script (`scripts/hrdb_etl_sync.sh`) 작성 — 별도 운영 sprint
- **(carve)** Phase 2 LDAP/AD federation 도입 — 사내 LDAP/AD 운영 중일 때 가능, IdP 운영팀 동반
- **(carve)** Phase 3 SCIM bridge — 사내 IdP 인프라 결정 동반
- **(carve)** Phase 3 실시간 webhook — HR system 측 webhook 도입 + backend `/internal/hr-event` endpoint 신설
- **(carve)** keycloak_event_audit_integration.md §4.2 admin event 매핑 표에 `USER:UPDATE` (enabled=false 특화) row 추가 — design + audit 통합 sprint 정합
- **(open)** 1h ETL cron 의 사내 정책 동의 — 운영팀의 cron 부담 vs latency trade-off
- **(open)** force logout 사전 공지 정책 — 24h 전 공지 vs 즉시 — 사내 보안팀 결정

## 10. 변경 이력

| 일자 | 변경 | sprint |
| --- | --- | --- |
| 2026-05-19 | 1차 draft — 10 section + 옵션 6종 비교 (A 현행 daily / B Keycloak admin force / C HR ETL push / D HR webhook / E LDAP federation / F SCIM bridge) + Phase 1 권장 옵션 C 상세 (hourly ETL + access_token TTL 5분 + 운영 SOP) + Phase 2 권장 옵션 E LDAP federation (5분 latency) + Phase 3 carve (SCIM / webhook) + 보안 5 위협 + cutover Phase 1..3 + ADR governance 결정 (1차 별도 ADR 없음, Phase 2 ADR-0021 후보) + ADR-0008 §6 와 정합 + carve 5 + open 2 항목. | `claude/work_260519-g` |
