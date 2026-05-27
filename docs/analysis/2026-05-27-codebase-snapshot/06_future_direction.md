# 06. 향후 개발 방향 + 도출 아이템

- 문서 목적: 본 스냅샷 분석(01~05) + 추적성/로드맵 점검을 토대로 향후 개발 방향을 정리하고 실행 가능한 백로그 아이템을 도출한다.
- 기준: main `cf19c94` (PR #374).
- 연계: 본 도출 결과는 [v1.0 릴리즈 로드맵](../../planning/release_v1_roadmap.md) §3/§4 와 [통합 개발 로드맵](../../development_roadmap.md) 에 반영한다.

---

## 1. 현재 위치 진단

- v1.0 scope 3 도메인(인증·조직·대시보드 / Application·Repository·Project / External Integration) + DREQ + Onboarding 은 **기능 완성**. v1.0 DoD 8 항목 중 코드 측면(1~6, 8)은 충족, 잔여는 **사내 운영 검증(7: staging 1주 + 외부 사용자 ≥5 로그인)** 위주.
- 2026-05-26~27 에는 v1.0 scope 를 넘어 **외부 연동 깊이(Gitea sync·SCM 양방향·auth_mode)** 로 자발적 확장이 진행됐다. 이는 사실상 **v1.1 영역 작업이 v1.0 전에 선행**된 상태.
- 가장 큰 부채는 **문서 drift**(본 PR 해소) + **FE 테스트 밀도** + **운영 가시성**.

## 2. 개발 방향 (3 축)

1. **굳히기(Harden)** — 신규 기능 속도를 잠시 늦추고, 따라잡기(E2E happy-path + FE 단위테스트 + #368 무테스트 보강 + 마이그레이션 hygiene)로 품질 부채를 상환한다.
2. **운영화(Operationalize)** — 백엔드가 만든 데이터(sync job 큐, provider health, audit, onboarding pending)를 보는 **System Admin 운영 대시보드**로 BE↔FE 균형을 회복하고, v1.0 staging 운영 검증을 통과한다.
3. **보안 정합(Secure)** — 평문 secret envelope 암호화(#6) + outbound 자격증명 정합 + CI prefix/secret guard 강화.

이후 v2 의 **확장(Extend)** — Realtime event publish/replay, AI Gardener, 외부 SSO federation.

## 3. 도출 백로그 (우선순위)

### 3.1 NOW — v1.0 마감 + 품질 굳히기 (다음 1~2 sprint)

| ID | 아이템 | 영역 | 출처 |
| --- | --- | --- | --- |
| **N-1** | **문서 drift 정합** — 추적성 ADR-0022/0023 + API-87..90 + migration 34~45 + repository/SCM 도메인 행 + Onboarding 완료 반영. 로드맵 ⏳→✅ 정정 | 문서 | 본 PR (Step 3/4) |
| **N-2** | **repository draft→publish UT/통합테스트 보강** (#368 무테스트 머지분) | BE | 02 §G4 / 04 §6 |
| **N-3** | **SCM import/create + draft/publish happy-path E2E** | FE+BE | 05 §B2 |
| **N-4** | **프론트 service/component 단위테스트 보강** (integration·repository·dashboard service + 신규 모달 최소 6종) | FE | 05 §B1 |
| **N-5** | **마이그레이션 prefix uniqueness CI guard 강화** (000042 충돌 재발 방지 — `uniq -d` gate + CI bypass 시에도 잡히게) | CI | 04 §6 / [[feedback_concurrent_migration_prefix_collision]] |
| **N-6** | **v1.0 staging 1주 운영 검증** (외부 사용자 ≥5 로그인 + Onboarding SOP DoD 8) | 사내/운영 | release_v1 §1.3 (7) |

### 3.2 NEXT — v1.1 운영화 + 외부 연동 깊이 정착 (target 2026-07)

| ID | 아이템 | 영역 | 출처 |
| --- | --- | --- | --- |
| **X-1** | **System Admin 운영 대시보드** (RM-M4-07) — Gitea sync job 큐/상태 + provider health + integration_sync_jobs 가시화 | FE+BE | 05 §B3 |
| **X-2** | **inbound webhook 정규화 깊이** — multi-provider sync 일반화 (Gitea 외 forgejo/gogs/github normalize 확장) | BE | session_handoff directive |
| **X-3** | **평문 secret envelope 암호화** (#6) — credentials_ref/api_token/auth_secret DEK 암호화 + 키 관리 ADR | BE/보안 | 04 §6 / 02 §G6 |
| **X-4** | **Phase D — project 생성 flow ↔ SCM create 연계** (현재 provider-scoped 독립 endpoint → project 생성 시 SCM repo 동반 생성) | FE+BE | session_handoff directive |
| **X-5** | **Gitea Hourly Pull 정밀화** (RM-M4-06 잔여) — hourly 스케줄 + reconciliation 정책 (issue #231) | BE | 추적성 RM-M4-06 |
| **X-6** | **Keycloak group staging-prod 적용** (P1-3, issue #214) | 사내/Codex | release_v1 §3.2 |
| **X-7** | **ADR-0016 §6 alert 임계 확정** (P2-2) — pull latency p95 + push webhook 알림 + stage→prod | Codex/infra | release_v1 §3.3 |
| **X-8** | **Keycloak SPI realm events push 전환** (P2-6/P3-5) — polling 30s → <1s, SPI JAR 빌드·배포 | BE/사내 | release_v1 §3.3/§3.4 |

### 3.3 LATER — v2 확장 (target 2026-Q3+)

| ID | 아이템 | 영역 | 출처 |
| --- | --- | --- | --- |
| **E-1** | **Realtime event publish 확장** (RM-M4-01) — infra.node/ci.run/risk.updated publish + FE 구독 UI | FE+BE | 02 §2.6 |
| **E-2** | **WebSocket replay + 리소스 필터링** (RM-M4-02) — last event replay + scope filter | BE | 추적성 RM-M4-02 |
| **E-3** | **AI Gardener gRPC** (RM-M4-04/05) — Python AnalysisService + Go client + suggestion feed UI | BE+FE | backend-ai 스켈레톤 |
| **E-4** | **Weekly report worker** — cron vs scheduled command 결정 후 구현 | BE | development_roadmap 백로그 |
| **E-5** | **RBAC PermissionCache LISTEN/NOTIFY** (RM-M4-08, ADR-0007) — 다중 인스턴스 일관성 | BE | release_v1 §3.4 |
| **E-6** | **외부 SSO federation** (RM-M4-09) — Keycloak identity broker (Gitea/AD/LDAP) | Codex/사내 | release_v1 §3.4 |
| **E-7** | **HomeLab dedicated worker binary + push/pull dedup** (ADR-0015 §6 (3)(4)) | BE | release_v1 §3.4 |
| **E-8** | **Keycloak HA Phase 2** (Infinispan + shared PG + LB) | 사내/Codex | release_v1 §3.4 |

### 3.4 명시 제외 (cancelled / 사내 정책)

- ~~Sign Up 셀프 가입~~ (cancelled, 외부 IdP 시나리오 — IdP 팀/HRDB ETL 책임).
- ~~off-boarding cron deploy~~ / ~~HRDB ETL unit pre-stage~~ (cancelled, event listener 정공법 + Onboarding self-service).
- MFA/2FA (Keycloak Account Console 위임 — 사내 정책).

## 4. 권장 진입 순서

1. **본 PR 머지** → 문서 drift 정합(N-1) 완료, 후속 작업의 기준선 확보.
2. **굳히기 sprint** (N-2/N-3/N-4/N-5) — 신규 기능 휴지기, 테스트·hygiene 상환. 작은 변경·낮은 위험.
3. **운영화 sprint** (X-1 운영 대시보드 + X-6 group staging-prod) — v1.0 staging 검증(N-6)과 병행.
4. **보안/깊이 sprint** (X-3 secret 암호화 + X-2 webhook 정규화 + X-4 Phase D).
5. **v2 확장** — E-1..E-8 은 우선순위·인프라 결정 후 진입.

## 5. 마일스톤 매핑 (갱신 후)

| 마일스톤 | 포함 | 상태 |
| --- | --- | --- |
| **M-v1.0** (target 06-15) | 3 도메인 + DREQ + Onboarding 기능 완성 + N-1..N-6 | 코드 완성, 운영 검증(N-6) 잔여 |
| **M-v1.1** (target 07-31) | X-1..X-8 (운영화 + 외부 연동 깊이 정착 + 보안) | 진입 전 (일부 X-2..X-5 는 v1.0 전 선행됨) |
| **M-v2** (Q3+) | E-1..E-8 (Realtime/AI/SSO/HA 확장) | backlog |

> 본 §3/§5 의 도출 아이템은 [release_v1_roadmap.md](../../planning/release_v1_roadmap.md) §3 우선순위 매트릭스에 신규 row(N-1..N-6 + X-1..X-4)로 통합한다.
