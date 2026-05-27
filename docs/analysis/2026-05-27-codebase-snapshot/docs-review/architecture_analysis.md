# architecture.md ↔ 코드 정합 분석 (2026-05-27 스냅샷)

- 문서 목적: `docs/architecture.md`(현행 618줄, 최종 수정일 2026-05-21)를 2026-05-27 코드 스냅샷(main `cf19c94`, 마지막 기능 커밋 `99d6edc`/PR #373)과 대조해 누락·불일치 구간을 줄 단위 근거로 식별하고, 비파괴(추가 + inline 정정) 갱신 계획을 확정한다.
- 범위: architecture.md 의 §1~§9 구조 점검 + 신규 §10 설계 + §8/§3/§6 보강 계획. 코드 수정은 다루지 않는다.
- 대상 독자: architecture.md 갱신 작업자, 후속 리뷰어, 추적성 동기화 담당.
- 상태: accepted (분석 스냅샷)
- 최종 수정일: 2026-05-27
- 관련 문서: [`../01_codebase_state_analysis.md`](../01_codebase_state_analysis.md), [`../code/backend/httpapi.md`](../code/backend/httpapi.md), [`../code/backend/store.md`](../code/backend/store.md), [`../code/backend/domain.md`](../code/backend/domain.md), [`../code/backend/migrations.md`](../code/backend/migrations.md), [`../code/backend/support_packages.md`](../code/backend/support_packages.md), [`../../../architecture.md`](../../../architecture.md)

## 1. 현행 architecture.md 구조 요약

| § | 제목 | ARCH ID | 코드 정합 상태 |
| --- | --- | --- | --- |
| 1 | 개요 | — | 정합 |
| 2 | 시스템 컴포넌트 구조 (mermaid) | — | **stale** — 거의 모든 노드가 `current: scaffold`/`planned`. 실제로는 backend 기능 완성, Gitea pull/webhook 양방향 가동 |
| 3 | 서비스 간 통신 (gRPC / REST·WS) | — | §3.2 WebSocket 계약은 맞으나 **ticket 인증(ADR-0024) 누락** |
| 4 | 데이터 전략 (SCM adapter / 하이브리드 동기화 / 파이프라인 / 스토리지) | — | 정합 (원칙 수준). 단 §4.1 "Hourly Pull" 은 실제 30s tick 와 표현 차 |
| 5 | UI/UX 시각화 (React Flow / 역할 대시보드) | — | 정합 |
| 6 | 보안 및 인증 (Webhook / User↔Account / OIDC / RBAC 단계화 / Audit) | — | §6.2.3 Keycloak 흐름 정합. **Keycloak 버전 pin(ADR-0022/0023) + SPI event listener 누락** |
| 7 | DREQ 도메인 | ARCH-DREQ-01..06 | 정합 (intake token expires_at 후속은 코드에 반영됐으나 §7 미언급 — 경미) |
| 8 | 외부 연동(Integration) 도메인 | ARCH-INT-01..06 | **부분 stale** — auth_mode 5종/base_url/api_token write-only/sync job 큐/webhook 헤더 alias 누락. auth_mode 표(§8.3)는 구 4종 |
| 9 | Onboarding 도메인 | ARCH-ONBOARD-01..06 | 정합 (lazy 폐기 후 token-only actor 반영됨) |
| — | (없음) | — | **Repository 소유권·연동·lifecycle 아키텍처 전체 누락** → 신규 §10 |

## 2. 누락/불일치 상세 (architecture.md:줄 근거)

### A. Repository 소유권·연동·lifecycle 전면 누락 (신규 §10 필요)

architecture.md 어디에도 다음 코드 사실이 반영돼 있지 않다.

- **소유권 분리(`source`/`provider_id`)**: `domain.Repository.{Source, ProviderID, ProviderKey, Description}`(`domain.md:23-28`), migration 000042(`migrations.md:54`). architecture.md §4.0 의 SCM adapter 원칙은 provider 라우팅 키(`repo_provider`)만 언급(`architecture.md:80`)하고 system-owned vs SCM-owned 분리 개념이 없음.
- **SCM mirror vs system-owned 보존**: `UpsertRepository` ON CONFLICT 가 SCM mirror 필드만 갱신하고 `source`/`provider_id`(COALESCE 기존우선)·`description`(SET 절 부재)을 보존(`store.md:133-140`). 이 보존 규약이 §4.2 정규화 파이프라인(`architecture.md:88-104`)에 없음.
- **inbound import (API-88/89)**: `listSCMRepositories`/`importSCMRepositories`(`httpapi.md:174-175,335`) — SCM 재조회 값으로 `source=scm` upsert. architecture.md 미기재.
- **outbound create (API-90, gitea CreateRepo)**: `createSCMRepository`(`httpapi.md:176`) + `gitea.Client.CreateRepo`(`support_packages.md:19`) — push capability + gitea-compat → `source=system`. 미기재.
- **draft→publish 상태머신**: `createRepositoryDraft`/`requestRepositoryPublish`(`httpapi.md:106-107`, `domain.md:119-121`) — `repository_status` draft|active, migration 000043. POST `/repositories` + `/publish` endpoint. 미기재.
- **capability gate**: `scmProviderForCapability`(exists+enabled+scm type+capability+gitea-compat, `httpapi.md:284`). import=pull, sync=pull|sync, create=push. 미기재.
- **provider_id 단일화(000045)**: `scm_provider`(key TEXT, #368) ↔ `provider_id`(FK UUID, #363) 의미 중복 → provider_id canonical, provider_key 는 LEFT JOIN derive(`store.md:45,172`, migration 000045). 미기재.

→ **결론**: §9 뒤에 신규 **§10 "Repository 소유권·연동·lifecycle 아키텍처"** (ARCH-REPO-01..07) 추가.

### B. §8 Integration 도메인 부분 stale

- **auth_mode 표 구식**: `architecture.md:371` `integration_providers.auth_mode` 주석은 `token | basic | oauth2 | app_password | agent` 5종으로 이미 맞으나, ARCH-INT 본문 어디에도 **`OutboundAuth`/`ResolveOutboundAuth` 자격증명 해석 모델**(`domain.md:130`, `application.go:265`)과 **mode 별 Authorization 헤더 산출**(`support_packages.md:23-28`, gitea `auth.go`)이 없음.
- **base_url / api_token write-only 누락**: §8.3 데이터 모델(`architecture.md:362-407`)에 `base_url`(migration 000038)·`api_token`(000040, write-only `api_token_set`)·auth_credentials 4컬럼(000041)이 없음. 현행 모델은 `auth_mode`/`capabilities`/`sync_status`만.
- **sync job 큐 누락**: `integration_sync_jobs`(migration 000028) + `AcquireNextQueuedSyncJob`(SKIP LOCKED, `provider_type='scm'` gate, `store.md:87`) + 백그라운드 `GiteaSyncWorker`(30s, `support_packages.md:35-43`)가 §8.2 동기화 전략(`architecture.md:347-360`)에 없음. §8.2 는 webhook/pull 병행 원칙만 추상 기술.
- **webhook 헤더 alias 누락**: §8.1 Adapter Router(`architecture.md:343-345`)는 `Verify(headers, body)` 추상 계약만. 실제 ingest(API-73)는 `X-Integration-*`→`X-Gitea-*`→`X-Gogs-*` fallback(`httpapi.md:281`, `support_packages.md:255`). 미기재.

→ **결론**: §8 에 **새 서브섹션 §8.7 "Gitea SCM pull sync + sync job 큐 + auth_mode/OutboundAuth + webhook 헤더 alias" (ARCH-INT-07)** 추가(기존 §8.1~§8.6 유지). §8.3 데이터 모델 표에는 신규 컬럼을 inline 정정 배너로 보강.

### C. §3 / §6 보안·인증 보강

- **WS ticket 인증 누락**: §3.2(`architecture.md:70-72`)는 WebSocket 을 실시간 계약으로만 규정. 실제 인증은 ticket 패턴 — `POST /api/v1/realtime/ticket` → `?ticket=` single-use 60s, PG(`DELETE...RETURNING`)/in-memory(`httpapi.md:195-196,252`, `store.md:142-144`, ADR-0024). legacy `?access_token=` query fallback 은 ticket-only 컷오버(ADR-0024 §6 carve 5)로 제거(`httpapi.md:221`). architecture.md 전체에서 ADR-0024 미참조.
- **Keycloak 버전 pin 누락**: ADR-0022(25.0) → ADR-0023(26.0) 버전 pin(`01_codebase_state_analysis.md:168`). §6 본문 + §6 관련 ADR 링크에 없음.
- **SPI event listener 누락**: Keycloak event listener SPI(Java, `infra/idp/keycloak-event-listener-spi/`) + push endpoint `POST /api/v1/internal/keycloak-events`(`X-Webhook-Secret`, `httpapi.md:87`) + polling 워커(`internal/audit`, 30s, `support_packages.md:51-87`). §6 Audit(`architecture.md:183-196`)에 Keycloak event → audit_logs 경로가 없음.

→ **결론**: §3.2 끝에 ticket 인증 inline 보강 배너 + §6 끝에 **§6.5 "Keycloak 버전 pin + SPI event listener + WS ticket 인증 (보강)"** 서브섹션 추가.

### D. §2 / §4 표현 차 (경미, inline 정정만)

- §2 mermaid 의 `current: scaffold`/`planned` 라벨(`architecture.md:24-51`)은 backend 기능 완성 현실과 괴리. 다이어그램 전면 재작성은 scope 외 → §2 직하에 "현행 구현 상태" inline 정정 배너 1개로 갈음.
- §4.1 "Hourly Pull"(`architecture.md:85-86`) 표현은 실제 Gitea sync 워커 30s tick(`support_packages.md:45-47`)·HomeLab 30s pull 과 상이. 원칙 수준이라 stale 배너 대신 §10/§8.7 의 실 주기 명시로 간접 정정.

### E. stale 용어 (Hydra/Kratos 등)

- §6.2.3(`architecture.md:164`)은 이미 ADR-0001(Hydra+Kratos, superseded) → ADR-0019(Keycloak) supersession 을 명시하고 있어 추가 배너 불요. 단 §6.2.3 흐름 1번(`architecture.md:168`)의 `/login` 진입 기술은 ADR-0019 §sub-carve F(`/login` canonical, `/auth/login` 제거)와 정합 — 정상.
- 관련 문서 링크(`architecture.md:9`)에 ADR-0001 을 "superseded" 로 명시하고 있어 정합. ADR-0022/0023/0024 링크만 추가 필요.

## 3. 갱신 계획 (비파괴 원칙)

| 항목 | 위치 | 방식 | 신규 ID |
| --- | --- | --- | --- |
| 메타 헤더 | `architecture.md:8-9` | 최종 수정일 2026-05-27 + 코드 스냅샷 링크 + ADR-0022/0023/0024 링크 추가 | — |
| §2 현행 상태 | `architecture.md:52` 직후 | inline 정정 배너(다이어그램 라벨 stale 고지) | — |
| §3.2 ticket 인증 | `architecture.md:72` 직후 | inline 보강 배너(ADR-0024) | — |
| §6.5 신규 | §6.4 뒤(`architecture.md:196` 직후) | 서브섹션 추가(Keycloak pin + SPI + WS ticket) | — |
| §8.3 신규 컬럼 | `architecture.md:376` 직후 | inline 정정 배너(base_url/api_token/auth_credentials) | — |
| §8.7 신규 | §8.6 뒤(`architecture.md:454` 직후) | 서브섹션 추가(sync 워커 + 큐 + OutboundAuth + 헤더 alias) | ARCH-INT-07 |
| §10 신규 | §9 뒤(`architecture.md:618` 직후) | 도메인 섹션 추가 | ARCH-REPO-01..07 |

**원칙 준수**:
- 기존 ARCH-DREQ-01..06 / ARCH-INT-01..06 / ARCH-ONBOARD-01..06 ID·prose **삭제·재번호 금지**.
- 신규 ID 는 미사용 namespace `ARCH-REPO-*` + 기존 `ARCH-INT-*` 의 다음 sequential(`-07`)만 사용.
- stale 표현은 본문 수정 대신 `> **정정(2026-05-27)**:` inline 배너로 처리.
- document-standards.md §2 메타 헤더 + §5 ARCH ID 본문 노출(섹션 제목/표 헤더) 준수.

### 3.1 신규 ARCH-REPO ID 배정안 (§10)

| ID | 소제목 |
| --- | --- |
| ARCH-REPO-01 | 소유권 분리 모델(source / provider_id / provider_key) |
| ARCH-REPO-02 | SCM mirror vs system-owned 필드 보존(UpsertRepository ON CONFLICT) |
| ARCH-REPO-03 | inbound import(API-88/89, SCM 재조회 upsert source=scm) |
| ARCH-REPO-04 | outbound create(API-90, gitea CreateRepo, source=system) |
| ARCH-REPO-05 | draft→publish 상태머신(repository_status, POST /repositories·/publish) |
| ARCH-REPO-06 | capability gate(scmProviderForCapability) + gitea-compat |
| ARCH-REPO-07 | provider_id 단일화(000045, scm_provider DROP, key join derive) |

## 4. 추적성 영향

- 신규 ARCH-REPO-01..07 + ARCH-INT-07 발급 → `docs/traceability/report.md` ARCH 단계 row 보강 대상(별도 sync sprint 또는 본 PR 추적성 섹션에서 처리). 본 분석 문서는 ID 발급 근거를 제공.
- 관련 API ID(API-88/89/90)·migration(000042/000043/000045) 는 기 발급(코드 스냅샷 분석 기준) — ARCH 단계만 신규.
