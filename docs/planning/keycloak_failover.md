# Design 검토 — Keycloak failover (HA + DR + DevHub graceful degradation)

- 문서 목적: [ADR-0019 §5.3](../adr/0019-keycloak-only-idp.md#53-잔여-carve-out) 잔여 carve out 의 Keycloak failover design. 1차 산출물은 planning 단계 — Phase 2 HA 진입 시 별도 ADR 승격 결정.
- 범위: Keycloak 단일 장애점 회피 + 가용성 향상. DevHub 측 graceful degradation 동작 분석 + Keycloak HA / DR 옵션 비교 + 권장 + cutover. backup IdP fallback 은 ADR-0019 Keycloak-only 결정과 충돌하므로 명시 제외.
- 대상 독자: 아키텍트, 운영자 (SRE / IdP), Security, 인프라 담당자, Backend 담당자.
- 상태: planning (draft 1차)
- 최종 수정일: 2026-05-19
- 결정 근거 sprint: `claude/work_260519-h`
- 관련 문서: [ADR-0019 Keycloak 단일화](../adr/0019-keycloak-only-idp.md), [ADR-0003 No-Docker policy CI scope](../adr/0003-no-docker-policy-ci-scope.md), [keycloak_operations.md](../setup/keycloak_operations.md) §6 JWKS rotation + §8.4 장애 대응, [backend-core/internal/auth/keycloak_verifier.go](../../backend-core/internal/auth/keycloak_verifier.go) (JWKS cache).

## 1. 컨텍스트 + 동기

### 1.1 현재 (옵션 A — 단일 Keycloak instance)

- ADR-0019 채택으로 Keycloak 이 DevHub 의 **단일 IdP**
- 단일 instance 운영 시 **SPOF (single point of failure)** — Keycloak down 시 영향:
  - 신규 로그인 / signup / password reset 차단
  - JWKS endpoint 응답 불가 → backend cache TTL 만료 후 JWT 검증 불가
  - admin API 호출 불가 → off-boarding 자동화 (sprint -g §3.1 HR ETL push) 일시 차단
- 사내 정책 측면 — SLA 요구 (예: 99.9% uptime) 와 충돌 가능

### 1.2 DevHub 측 graceful degradation (현재 자연 동작)

`backend-core/internal/auth/keycloak_verifier.go:37` 의 `defaultJWKSTTL = 5 minute` + Keycloak 권장 access_token TTL 5분 ([keycloak_operations §6.2](../setup/keycloak_operations.md#62-rotation-주기-권장)) 결합:

```text
Keycloak down 발생
  ↓ 0-5분
backend JWKS cache 유효 — 기존 access_token 검증 정상
  ↓ 5분 후
JWKS cache TTL 만료 → fetch 시도 실패 → JWT 검증 401
  ↓ 또는 access_token 만료
frontend refresh 시도 → Keycloak unreachable → logout 또는 retry
  ↓ Keycloak 복구 후
정상 동작 자동 복귀 (cache 자동 fetch)
```

**worst case graceful degradation window** = max(JWKS cache TTL 5분, access_token TTL 5분) ≈ **5-10분**.

→ 5-10분 window 내 Keycloak 복구 시 사용자 무영향. 그 이상 down 시 점진 logout.

### 1.3 가용성 목표

- **Phase 1 (단기)**: Keycloak instance uptime + graceful degradation 명문화. SLA = Keycloak instance uptime, 5-10분 미만 down 시 사용자 무영향.
- **Phase 2 (중기)**: Keycloak HA (active-active 또는 active-passive). uptime ≥ 99.9%.
- **Phase 3 (장기, carve)**: DR site (별도 datacenter + DB replication + DNS failover). RTO ≤ N분 / RPO ≤ N분 (사내 SRE 결정).

## 2. 통합 옵션 비교 (6종)

| 옵션 | 변경 범위 | DevHub backend 영향 | uptime / RTO | 권장 |
| --- | --- | --- | --- | --- |
| **A. 단일 Keycloak instance (현행)** | 없음 | 없음 (graceful degradation 자연 동작) | instance uptime ≈ 99.5% | △ SPOF — 부족 |
| **B. Keycloak HA (active-active cluster)** | Keycloak 다중 instance + Infinispan cache replication + shared PostgreSQL + 사내 load balancer | 없음 (DevHub 는 OIDC issuer URL 하나만 신뢰) | ≥ 99.9%, RTO ~ 0초 | ⭐ **Phase 2 권장** |
| **C. Keycloak active-passive (standby)** | primary Keycloak + standby (cold/warm/hot) + DB replication + DNS 또는 LB failover | 없음 | 99.9%, RTO 분-수십분 | △ active-active 보다 단순하나 cutover latency |
| **D. DevHub graceful degradation (명문화)** | DevHub 측 변경 없음 (이미 JWKS cache + access_token TTL 자연 graceful) | 없음 | 5-10분 graceful window | ⭐ **Phase 1 권장** (현행 인프라 그대로) |
| **E. backup IdP fallback** (자체 password method 부활) | ADR-0019 Keycloak-only 결정 회귀 — Kratos 또는 자체 auth backend 재도입 | 큼 — 옵션 B Kratos federation 회귀 의미 | — | ❌ **명시 제외** (ADR-0019 결정 충돌) |
| **F. DR site (별도 datacenter)** | 별도 datacenter Keycloak instance + DB replication + DNS failover + 사내 datacenter 다중화 | 없음 (DNS 통한 자동 failover 가정) | RTO 수분-수십분 / RPO 초-분 | △ **Phase 3 carve** (사내 datacenter 다중화 시) |

## 3. Phase 1 (단기 권장) — 옵션 D graceful degradation 명문화

### 3.1 현재 동작 분석

DevHub backend 의 자연 graceful degradation:

| 구성 요소 | TTL / lifespan | Keycloak down 시 동작 |
| --- | --- | --- |
| backend JWKS cache | 5분 ([keycloak_verifier.go:37](../../backend-core/internal/auth/keycloak_verifier.go) `defaultJWKSTTL`) | cache hit 시 정상 검증, cache miss 시 fetch 실패 → 401 |
| access_token | 5분 (권장, [keycloak_operations §6.2](../setup/keycloak_operations.md#62-rotation-주기-권장)) | 유효 기간 동안 정상 동작 |
| refresh_token | 12h (권장) | 사용자가 refresh 시도 시 Keycloak unreachable → frontend logout 또는 retry |
| SSO Session Idle | 30분 ([keycloak_operations §2](../setup/keycloak_operations.md)) | Keycloak 측 session 유지 — 복구 후 정상 재개 |

**핵심**: backend 측 코드 변경 없음 — JWKS cache + JWT TTL 기반 자연 graceful degradation 이 이미 동작.

### 3.2 운영 정책

- **5분 미만 down**: 사용자 무영향 (cache + token 유효). 모니터링만, 사용자 공지 불필요.
- **5-10분 down**: 점진 logout 발생. 사용자 공지 (예: status page) + Keycloak 복구 진행 알림.
- **10분 이상 down**: 전체 신규 인증 차단. Page on-call SRE + 사용자 영향 분석 + 사후 review.

### 3.3 backend 변경 없음

옵션 D 는 design 명문화 + 운영 SOP 만. backend / frontend 코드 변경 없음.

권장 운영 모니터링 metric (carve, [keycloak_event_audit_integration §5.3](./keycloak_event_audit_integration.md) PR-C 와 정합):
- `devhub_jwks_fetch_total{status}` — JWKS fetch 성공/실패 counter
- `devhub_jwks_cache_age_seconds` — cache 잔여 시간 gauge
- alert: `devhub_jwks_fetch_total{status="failed"} > 0 for 2 minutes` → Keycloak 가용성 의심

## 4. Phase 2 (중기 권장) — 옵션 B Keycloak HA (active-active)

### 4.1 Keycloak HA 표준 구성

Keycloak 25+ 권장 구성 (사내 인프라 결정 영역):

```text
┌─────────────────┐       ┌─────────────────┐
│ Keycloak inst-1 │ ────► │   PostgreSQL    │ ◄──── Keycloak inst-2
└────────┬────────┘       │   (shared HA)   │       └────────┬────────
         │                └─────────────────┘                │
         │                                                   │
         ▼                                                   ▼
   ┌──────────────────────────────────────────────────────────┐
   │      Infinispan distributed cache (session replication)  │
   └──────────────────────────────────────────────────────────┘
                              ▲
                              │ (LB front)
                   ┌──────────┴──────────┐
                   │ nginx / haproxy LB  │  ◄── DevHub frontend + backend
                   └─────────────────────┘
```

핵심 구성 요소:
- **다중 Keycloak instance** (최소 2) — active-active. 사내 보안 정책에 따라 N instance.
- **Shared PostgreSQL HA** — Keycloak DB 의 가용성. patroni / pg_auto_failover / 사내 DB 운영 표준.
- **Infinispan distributed cache** — session replication (logout / token revocation 일관성).
- **사내 load balancer** — sticky session 또는 stateless (Infinispan 가 보장).
- **단일 issuer URL** — `https://devhub.example.com/devhub/auth/keycloak/realms/devhub` ([ADR-0018](../adr/0018-single-port-reverse-proxy-policy.md) 단일 포트 정합).

### 4.2 DevHub 측 변경 없음

- backend `DEVHUB_OIDC_ISSUER_URL` = 단일 LB endpoint (현재 운영 그대로)
- backend JWKS cache + token verifier 동작 동일
- HA 도입은 Keycloak 운영팀 + 사내 인프라 책임

### 4.3 trade-off

| 측면 | HA active-active | HA active-passive (옵션 C) |
| --- | --- | --- |
| uptime | ≥ 99.9% | 99.9% (cutover latency 영향) |
| RTO (장애 → 복구) | ~ 0초 (LB 자동 라우팅) | 분-수십분 (DNS / LB failover 수동 또는 자동) |
| 운영 부담 | 다중 instance + Infinispan + DB HA | primary + standby + DB replication |
| 비용 | ↑ (다중 instance) | 중간 |
| 복잡도 | ↑ (Infinispan tuning) | 중간 |

권장 = **옵션 B (active-active)** — 사내 보안/규제가 99.9%+ 요구 시. 99% 수준이면 옵션 C (active-passive) 도 acceptable.

### 4.4 cutover

- staging 환경에서 HA 구성 1주 검수 (failover 시뮬레이션 포함)
- 사내 SRE + IdP 운영팀 동반
- DevHub 측 변경 없으므로 PR 불필요

## 5. Phase 3 (장기 carve) — 옵션 F DR site

### 5.1 DR site 구성

- 별도 datacenter 의 Keycloak instance + DB replication (async or sync)
- DNS failover 또는 GSLB (Global Server Load Balancing)
- RTO / RPO 사내 SRE 결정 (예: RTO ≤ 30분 / RPO ≤ 5분)

### 5.2 적용 시점

- 사내 datacenter 다중화 정책 결정 후
- Keycloak HA (Phase 2) 와 별도 — Phase 2 가 single-datacenter HA, Phase 3 가 cross-datacenter HA

Phase 3 carve — 사내 인프라 결정 동반.

## 6. 옵션 E 명시 제외 — backup IdP fallback

### 6.1 옵션 E 가 의미하는 것

- Keycloak 장애 시 별도 IdP (예: 자체 password method, Kratos 재도입, 또는 emergency local auth) 활성화
- 사용자가 Keycloak 외 backup IdP 로 로그인 가능

### 6.2 ADR-0019 결정과의 충돌

- ADR-0019 §3.1 = "Keycloak 단일 IdP 로 일원화"
- ADR-0019 §3.3 결정 근거 = "운영 단순성 + SSO 사용자 공존 부담 회피 + MFA-WebAuthn-recovery 표준 흡수 + `users.idp_subject` 일반화"
- backup IdP fallback 도입 시:
  - ADR-0019 §3.3 (1) 운영 단순성 → multi-stack 운영 복귀 (역행)
  - ADR-0019 §3.3 (2) SSO 사용자 공존 부담 → 회귀
  - keycloak_sso_federation.md (rejected 옵션 B) 의 dual stack 패턴으로 회귀

**결정**: 옵션 E 는 ADR-0019 와 충돌. **명시 제외**. failover 는 Keycloak 자체의 HA / DR 로만 해결.

### 6.3 대안

backup IdP 가 필요한 진짜 case 가 발생하면 ADR-0019 재검토 (옵션 B Kratos federation 등) 가 정답 — 별도 sprint + ADR 후보.

## 7. 보안 점검

### 7.1 잠재 위협 + mitigation

| 위협 | mitigation |
| --- | --- |
| Phase 1 graceful degradation 의 5-10분 window 동안 신규 로그인 차단 | 사용자 공지 (status page) + Keycloak 복구 SLA 단축 |
| HA 구성의 split-brain (Infinispan partition) | Keycloak quorum 설정 + monitoring |
| DB HA 의 replication lag → session 불일치 | sync replication (latency ↑) 또는 sticky session (실패 시 무재인증) |
| DR site 의 DB replication 지연 → 데이터 손실 | RPO 명문화 + 사내 SRE 정책 |
| LB SPOF | LB 자체 HA (active-active LB 또는 anycast) |
| Keycloak HA 의 secret 공유 (admin password / vault key) | 사내 vault 의 multi-instance 공유 정책 + secret rotation SOP (keycloak_operations §8.3) |

### 7.2 audit log 영향

- HA / DR 환경에서 audit_logs 의 `source_ip` enrichment 정합 — LB 통과 후의 X-Forwarded-For 신뢰 정책 (ADR-0018 §3.5 trusted proxies)
- Keycloak admin event log 가 multi-instance 의 LB 가 통과 가정 — Keycloak Infinispan 이 admin event 도 replicate

## 8. cutover 절차

### 8.1 Phase 1 (단기, 본 sprint) — design + 운영 SOP 명문화

- ✅ 본 design 문서
- ADR-0019 §5.3 (6) failover carve resolved (design)
- keycloak_operations §8.4 장애 대응 표 확장 (graceful degradation SOP)
- 권장 모니터링 metric carve

### 8.2 Phase 2 (중기, 사내 인프라 결정 시) — Keycloak HA 도입

- 사내 SRE / IdP 운영팀 + DB 운영팀 + 보안팀 합의
- staging 환경 HA 구성 + failover 시뮬레이션 1주
- prod HA 도입 + 모니터링
- DevHub backend 변경 없음

### 8.3 Phase 3 (장기 carve) — DR site

- 사내 datacenter 다중화 정책 결정 후
- 별도 sprint + ADR-0022 후보

## 9. ADR governance 결정

본 design 의 ADR governance 측면:

- Phase 1 (옵션 D graceful degradation) = 운영 SOP 수준 — 별도 ADR 발행 가치 낮음
- Phase 2 (옵션 B HA) = Keycloak HA 도입 결정 — ADR-0021 후보 (현재 별도 ADR 없음, ADR-0019 §5.3 carve resolved 만으로 충분 가능)
- Phase 3 (옵션 F DR site) = 별도 ADR (ADR-0022+) 후보

**1차 결정**: ADR-0019 §5.3 (6) carve resolved (design) 만으로 한정. Phase 2 진입 시 ADR-0021 재평가. ADR-0019 의 옵션 E 명시 제외도 본 design 에서 결정.

## 10. 잔여 carve out / open question

- **(carve)** Phase 2 Keycloak HA 도입 — 사내 SRE + IdP 운영팀 동반. staging 환경 검수 + prod cutover.
- **(carve)** Phase 3 DR site — 사내 datacenter 다중화 정책 결정 동반.
- **(carve)** 운영 모니터링 metric 도입 — `devhub_jwks_fetch_total` + `devhub_jwks_cache_age_seconds`. sprint -e PR-C 와 정합.
- **(carve)** status page (사용자 공지) 운영 SOP — Keycloak 5분 이상 down 시 자동 공지 정책.
- **(open)** Keycloak HA SLA 목표 — 99.9% / 99.95% / 99.99% 결정. 사내 보안팀.
- **(open)** access_token TTL 단축 (현재 권장 5분) 시 graceful degradation window 축소 영향. trade-off 평가.

## 11. 변경 이력

| 일자 | 변경 | sprint |
| --- | --- | --- |
| 2026-05-19 | 1차 draft — 11 section + 옵션 6종 비교 (A 단일 / B HA active-active / C HA active-passive / D DevHub graceful degradation / E backup IdP fallback 명시 제외 / F DR site) + Phase 1 권장 옵션 D 상세 (JWKS cache 5분 + access_token 5분 = 5-10분 graceful window + 운영 정책 + 권장 모니터링 metric) + Phase 2 권장 옵션 B Keycloak HA (Infinispan + shared PG + LB + 단일 issuer URL) + Phase 3 carve 옵션 F DR site + 옵션 E ADR-0019 충돌 명시 제외 + 보안 6 위협 + cutover Phase 1..3 + ADR governance 결정 (1차 별도 ADR 없음, Phase 2 ADR-0021 후보) + carve 4 + open 2. | `claude/work_260519-h` |
