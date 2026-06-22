---
title: architecture
type: source
tags: [domain, architecture.md, project-devhub]
sources: [raw/projects/devhub/docs/domain/integration-registry/architecture.md]
git_commit: 71c0d2cd
git_branch: chore/260622-wiki-drift-cleanup
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
last_touched: 2026-06-22T03:47:55Z
mirror_dirty: |
related: [none]
status: draft
contradictions: [none]
---

# integration-registry 도메인 아키텍처

- 문서 목적: 외부 시스템 연동(Integration) 도메인의 컴포넌트·동기화 전략·데이터 모델·보안·홈랩 수집·장애 격리 아키텍처를 정의한다.
- 범위: ARCH-INT-01..07. Task Item Ingestion 은 `task_architecture.md` (ARCH-TASK-01..06) 참조. cross-cutting 3대 레이어 / SCM Provider Adapter 원칙은 master `docs/architecture.md` §1–§4 참조.
- 대상 독자: backend / 프론트엔드 / DevOps, AI agent, 아키텍처 검토자.
- 상태: accepted
- 최종 수정일: 2026-05-29 (Phase 3 split, master `docs/architecture.md` §8 본문 이관)
- 관련 문서: [도메인 README](./README.md), [컨셉](./external_system_concept.md), [requirements.md](./requirements.md), [api.md](./api.md), [master architecture](../../architecture.md), [ADR-0015](../../adr/0015-homelab-pull-strategy.md)

## 개요

컨셉 문서: [`./external_system_concept.md`](./external_system_concept.md), 요구사항: [`./requirements.md`](./requirements.md), Usecase: [`UC-INT-01..14`](../../planning/system_usecases.md).

## 1. 컴포넌트 경계 (ARCH-INT-01)

```
┌──────────────────────────────────────────────────────────────────────┐
│                    Go Core Integration Domain                        │
│                                                                      │
│  Provider Registry ──┬── Adapter Router ──┬── Ingest Pipeline        │
│  (type,capability,   │                    │   (webhook/pull)         │
│   enabled,auth,scope)│                    │                           │
│                      │                    └── Normalize Pipeline       │
│                      │                        (repo/pr/build/doc/infra)│
│                      │                                                │
│                      └── Health/Status Manager (sync_status)         │
└──────────────────────────────────────────────────────────────────────┘
           │                         │                           │
           ▼                         ▼                           ▼
   External ALM/SCM/CI         External Doc System          HomeLab Agents
 (Jira/Bitbucket/Gitea/...)    (Confluence 등)             (node/service telemetry)
```

- Core 는 provider 중립 계약만 유지하고, provider-specific API 차이는 Adapter 내부에서 흡수한다.
- provider 장애는 격리 경계로 취급해 전체 파이프라인 중단으로 확산되지 않게 한다.
- `Adapter Router` 는 provider별 webhook 검증 전략을 분리한다.
  - 예: HMAC-SHA256, token compare, provider SDK verifier
  - 공통 contract: `Verify(headers, body) -> (ok, reason)` 를 제공하고 API-73 ingest 전에 실행

## 2. 동기화 전략 (ARCH-INT-02)

- 두 경로를 병행한다.
  - 실시간 경로: webhook ingest
  - 보정 경로: scheduled pull (reconciliation)
- 동일 자원에 대해 idempotency key를 사용해 중복 처리/중복 저장을 방지한다.
- 정규화 결과는 snapshot + event history 로 분리 저장한다.
- 동기화 우선순위 규칙:
  - 동일 `resource_type + external_id` 에 대해 `occurred_at` 이 더 최신인 이벤트를 우선한다.
  - `occurred_at` 이 같으면 `ingested_at` 이 더 늦은 이벤트를 최종 반영한다.
  - pull 경로는 webhook 미수신 구간 보정만 수행하며, 최신 watermark 이후 데이터만 처리한다.
- 충돌 정책:
  - 외부 SoT 필드와 DevHub 내부 주석성 필드가 충돌할 때 SoT 필드는 외부 원천값 우선.
  - 충돌 감지 시 `integration.conflict.detected` audit 을 기록하고 운영 화면에 경고 배지를 노출한다.

## 3. 데이터 모델 초안 (ARCH-INT-03)

```text
integration_providers
  provider_id          uuid PK
  provider_key         text UNIQUE            -- jira, confluence, gitea, forgejo, bitbucket, jenkins, bamboo, homelab
  provider_type        text NOT NULL          -- alm | scm | ci_cd | doc | infra
  display_name         text NOT NULL
  enabled              boolean NOT NULL
  auth_mode            text NOT NULL          -- token | basic | oauth2 | app_password | agent
  capabilities         jsonb NOT NULL         -- ["repo.read","pr.read",...]
  sync_status          text NOT NULL          -- requested | verifying | active | degraded | disconnected
  last_sync_at         timestamptz NULL
  last_error_code      text NULL
  created_at, updated_at timestamptz NOT NULL

integration_bindings
  binding_id           uuid PK
  scope_type           text NOT NULL          -- application | project
  scope_id             text NOT NULL
  provider_id          uuid NOT NULL REFERENCES integration_providers(provider_id)
  external_key         text NOT NULL
  policy               text NOT NULL          -- summary_only | execution_system | bidirectional_candidate
  created_at, updated_at timestamptz NOT NULL
  UNIQUE(scope_type, scope_id, provider_id, external_key)

infra_nodes
  node_id              text PK
  provider_id          uuid NOT NULL REFERENCES integration_providers(provider_id)
  hostname             text NOT NULL
  ip_address           text NOT NULL
  environment          text NOT NULL          -- homelab | stage | prod
  status               text NOT NULL          -- stable | warning | down
  metrics              jsonb NOT NULL         -- cpu/mem/disk/load
  observed_at          timestamptz NOT NULL

infra_services
  service_id           text PK
  node_id              text NOT NULL REFERENCES infra_nodes(node_id)
  name                 text NOT NULL
  version              text NULL
  port                 int NULL
  health_status        text NOT NULL          -- healthy | degraded | down
  metadata             jsonb NOT NULL
  observed_at          timestamptz NOT NULL
```

> **정정/보강 (2026-05-27)**: 위 `integration_providers` 초안에는 outbound 연동에 필요한 자격증명 컬럼이 빠져 있다. 현행 스키마는 다음 컬럼이 추가됐다(상세 모델은 §7):
> - `base_url text NULL` (migration 000038) — provider API 엔드포인트. `auth_token_url` 과 함께 `http(s)+host` 검증.
> - `api_token text NULL` (migration 000040) — outbound PAT. **write-only** — 응답에는 raw 미노출, `api_token_set` bool 만.
> - `auth_username / auth_client_id / auth_token_url / auth_secret` (migration 000041) — `auth_mode` 별 구조화 자격증명. `auth_secret` 도 write-only(`auth_secret_set` bool).
> - `credentials_ref text` — inbound webhook 서명 검증용 시크릿. **현재 GET 응답에 평문 노출되는 알려진 보안 gap**(#6 평문 secret 저장 carve, envelope 암호화 미적용).
>
> `auth_mode` 값은 본 §3 표대로 `token | basic | oauth2 | app_password | agent` 5종이며, mode 별 Authorization 헤더 산출(`OutboundAuth`/`ResolveOutboundAuth`)은 §7 참조. `integration_sync_jobs` 큐 테이블(migration 000028)도 본 초안에는 누락 — §7 에서 정의한다.

- `capabilities` 는 provider type 별 최소 표준 키를 포함한다.
  - `alm`: `issue.read`, `epic.read`, `issue.link`
  - `scm`: `repo.read`, `pr.read`, `branch.read`, `webhook.ingest`
  - `ci_cd`: `build.read`, `deploy.read`, `job.rerun`
  - `doc`: `page.read`, `space.read`, `doc.link`
  - `infra`: `node.read`, `service.read`, `snapshot.ingest`
- `integration_bindings.policy` 는 scope-연동 책임을 의미한다.
  - `summary_only`: 읽기 전용 요약
  - `execution_system`: 실행/상태 판단의 기준 시스템
  - `bidirectional_candidate`: write-back 후보(ADR 승인 전 비활성)

## 4. 보안/권한 경계 (ARCH-INT-04)

- Provider credential 은 평문 저장을 금지한다 (encrypted at rest 또는 external secret manager 참조).
- 연동 생성/수정/비활성화는 `system_admin` 권한만 허용한다.
- 조회는 scope 기반으로 제한한다:
  - `system_admin`: 전체 조회
  - 일반 역할: 자신의 접근 가능한 Platform/Project scope 한정
- 감사로그 action namespace: `integration.*`, `infra.node.*`, `infra.service.*`

## 5. 홈랩 수집 경계 (ARCH-INT-05)

- 홈랩은 infra provider 로 취급한다 (`provider_type=infra`).
- 수집 방식은 1차에 Agent Push 를 기본 후보로 둔다.
  - Agent 가 node/service 상태를 DevHub ingest endpoint 로 전송
  - DevHub 는 마지막 스냅샷 + 상태 변경 이력을 동시 관리
- 수집 실패 시 provider 상태를 `degraded` 로 전이하고 경고를 노출한다.
- Agent payload 최소 계약:
  - `agent_id`, `snapshot_at`, `nodes[]`, `services[]`, `trace_id`
  - 각 node/service 는 `observed_at` 필수
  - 동일 `agent_id + snapshot_at` 재전송은 idempotent 처리
- Adapter 연동 범위 (baseline):
  - API-77 ingest payload를 `infra_service_snapshots`로 영속화하고, API-76/API-78 조회 시 최신 persisted snapshot hydrate를 지원한다.
  - adapter 계약은 `save_snapshot/load_latest_snapshot` 읽기·쓰기 경계까지만 포함한다.
- Adapter 연동 범위 (후속):
  - provider별 delta upsert (`infra_nodes`, `infra_services`)와 변경 이력(event log) 분리 저장.
  - pull/reconciliation 경로와 push ingest 간 watermark 정합, 충돌 해결 정책 적용.

## 6. 장애 격리 및 복구 (ARCH-INT-06)

- provider별 retry/backoff 정책을 독립적으로 적용한다.
- 특정 provider 의 반복 실패는 circuit-open 상태로 격리하고, 나머지 provider 파이프라인은 지속 처리한다.
- 운영자는 provider 단위로 수동 재동기화(re-sync) 요청을 트리거할 수 있어야 한다.
- `degraded` 전이 임계값은 설정 가능(configurable)해야 한다.
  - 기본 예시: `failure_threshold=3`, `window=5m`, `cooldown=10m`
  - 홈랩/사내망 환경 특성에 맞춰 provider별 override 를 허용한다.

## 7. Gitea SCM pull sync 워커 · sync job 큐 · auth_mode/OutboundAuth · webhook 헤더 alias (ARCH-INT-07)

> 본 절은 §1~§6 의 provider-중립 연동 원칙을, 2026-05-21 이후 코드에 구현된 **Gitea(및 Forgejo/Gogs 호환) SCM 연동의 구체 아키텍처**로 보강한다. 기존 ARCH-INT-01..06 은 변경 없이 유지된다.

### 7.1 SCM pull sync 워커 + sync job 큐

- **데이터 모델(보강)**: `integration_sync_jobs`(migration 000028) — provider 단위 sync 작업 큐. status(`queued | running | succeeded | failed`).
- **큐 소비**: store 의 `AcquireNextQueuedSyncJob` 가 `provider_type='scm'` gate + `FOR UPDATE ... SKIP LOCKED` 로 단건 acquire 한다. 비-SCM job 은 store 레이어에서 차단되어 워커에 도달하지 않는다. multi-instance 에서 같은 job 을 두 워커가 잡지 않도록 SKIP LOCKED 가 직렬화한다.
- **백그라운드 워커**: `internal/gitea` 의 SCM sync 워커가 `main.go` 의 30s 주기 goroutine 으로 기동(`pgStore != nil` 일 때 항상). `ProcessOnce` 순서:
  1. queued sync job 을 우선 acquire.
  2. `resolveSyncConfig` 로 provider 의 `base_url` + `auth_mode` 별 자격을 해석. base_url 없거나 자격 미설정(또는 `agent` mode) → job `failed`.
  3. `ListUserRepos` → repo 마다 `UpsertRepository`(`source=scm`, `provider_id` 기록) + repo 단위 deep sync(issues open/closed + PRs open/closed upsert).
  4. 큐가 비면 env(`GITEA_URL`/`GITEA_TOKEN`) 기반 legacy 주기 sync(둘 다 있을 때만).
- **보안 핵심 — env fallback 금지**: 명시 provider 를 해석할 때 worker-global env 토큰으로 **fallback 하지 않는다**(provider 고유 host 에 잘못된 계정 토큰이 유출되는 것을 차단). env fallback 은 provider 미명시(legacy) 경로에만 허용된다.
- **운영 가시성 부채**: 본 워커는 현재 Prometheus metric 이 없어 진행/실패가 로그로만 노출되고, 30s 주기는 env override 불가(하드코딩)다 — 후속 hardening 후보.

### 7.2 auth_mode 모델 + OutboundAuth/ResolveOutboundAuth

- `IntegrationProvider.auth_mode` 5종에 대해, provider receiver `ResolveOutboundAuth()` 가 active mode 별 자격증명 컬럼을 `OutboundAuth` 구조로 해석한다.

| auth_mode | 사용 컬럼 | 산출 Authorization 헤더 |
| --- | --- | --- |
| `token` / unset | `api_token` | `token <pat>` |
| `basic` | `auth_username` + `auth_secret` | `Basic base64(user:secret)` |
| `app_password` | `auth_username` + `auth_secret` | `Basic base64(user:secret)` |
| `oauth2` | `auth_client_id` + `auth_token_url` + `auth_secret` | client-credentials grant 교환 후 `Bearer <token>` |
| `agent` | (직접 API sync 불가) | — (skip) |

- 자격 누락 시 `ok=false` 로 skip(워커는 job `failed` 처리), oauth2 토큰 교환 실패는 error 로 전파한다.
- `api_token`/`auth_secret` 은 API 레이어에서 write-only(`*_set` bool)로 가려지지만 store/도메인 레이어는 raw 평문을 그대로 보관한다(at-rest 암호화 부재, #6 carve).

### 7.3 webhook 헤더 alias (inbound ingest)

- 범용 ingest endpoint `POST /api/v1/integration/providers/:id/webhook`(API-73)는 서명 헤더를 `X-Integration-Signature` → `X-Gitea-Signature` → `X-Gogs-Signature` 순으로 fallback 수용한다(Gitea 가 `X-Gitea-Signature` 를 보내는데 초기 코드가 `X-Integration-Signature` 만 보던 헤더 불일치를 정정).
- 서명 검증은 provider 의 `credentials_ref` 전략별(`hmac_sha256:<secret>` / `provider_sdk:<vendor>:<secret>` / shared token)로 수행하고, 통과 시 dedupe `SaveWebhookEvent` + sync state best-effort 갱신.
- 별도 전용 Gitea webhook 핸들러 `POST /api/v1/integrations/gitea/webhooks`(API-02)는 `X-Gitea-Signature`/`X-Gogs-Signature` 만 수용한다(`X-Integration-*` 미수용) — 두 경로의 헤더 수용 범위가 달라 일관성 부채로 남아 있다.

## 8. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-29 | Phase 3 split — master `docs/architecture.md` §8 (External Integration 본문) 을 도메인 sub-document 로 이관. ID(ARCH-INT-01..07) 보존, 신규 발급/삭제 없음. Task Ingestion 은 `task_architecture.md` 로 분리. |
