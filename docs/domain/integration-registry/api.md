# integration-registry 도메인 API

- 문서 목적: 외부 시스템 연동(Integration) Provider/Binding/HomeLab/SCM repo 연동 API 계약을 정의한다.
- 범위: API-69..78 + API-80 + API-87..90. envelope/공통 enum 은 master `docs/backend_api_contract.md` §1–§2 또는 `docs/api/conventions.md` 참조. Task Item Ingestion API 는 `task_api.md` (API-94..96) 참조.
- 대상 독자: backend / 프론트엔드 / DevOps, AI agent, QA.
- 상태: accepted
- 최종 수정일: 2026-05-29 (Phase 3 split, master §15 본문 이관)
- 관련 문서: [도메인 README](./README.md), [requirements.md](./requirements.md), [architecture.md](./architecture.md), [master API](../../backend_api_contract.md)

## 개요

본 문서는 [`./external_system_concept.md`](./external_system_concept.md) 및 [`./requirements.md`](./requirements.md) 의 API 계약이다. ID는 임시 발급(`API-69..78`)이며 상세 응답 스키마는 설계 sprint 에서 확정한다.

## 1. API ID 인덱스

| API ID | endpoint | 목적 |
| --- | --- | --- |
| API-69 | `GET /api/v1/integration/providers` | Provider catalog 조회 |
| API-70 | `POST /api/v1/integration/providers` | Provider 등록 |
| API-71 | `PATCH /api/v1/integration/providers/{provider_id}` | Provider 수정/활성화/비활성화 |
| API-72 | `POST /api/v1/integration/providers/{provider_id}/sync` | Provider 수동 재동기화 트리거 |
| API-73 | `POST /api/v1/integration/providers/{provider_key}/webhook` | Provider webhook ingest |
| API-74 | `GET /api/v1/integration/bindings` | scope별 Integration binding 조회 |
| API-75 | `POST /api/v1/integration/bindings` | scope별 Integration binding 생성 |
| API-76 | `GET /api/v1/infra/services` | 홈랩 서비스 인벤토리 조회 |
| API-77 | `POST /api/v1/infra/services/snapshot` | 홈랩 서비스 상태 스냅샷 수집 ingest |
| API-78 | `GET /api/v1/infra/topology/v2` | 노드+서비스+의존성 통합 토폴로지 조회 |
| API-80 | `DELETE /api/v1/integration/providers/{provider_id}` | Provider 삭제 (FK guard, sprint `claude/work_260518-j`) |
| API-87 | `POST /api/v1/integration/test-connection` | Provider endpoint reachability 테스트 (등록 UX 고도화 #5) |
| API-88 | `GET /api/v1/integration/providers/{provider_id}/scm-repositories` | SCM provider 원격 repository 목록 조회 (import 대상) |
| API-89 | `POST /api/v1/integration/providers/{provider_id}/import-repositories` | 선택 repository 를 시스템으로 import/연동 |
| API-90 | `POST /api/v1/integration/providers/{provider_id}/create-repository` | 선택 SCM 에 실제 저장소 생성 + 시스템 미러 (Phase C) |

> **capability 기능 gate** (sprint `claude/work_260527-scm-repo-sync` / `-phase-c`): provider `capabilities` 는 표시 라벨이 아니라 기능 gate 다. `pull` = SCM 으로부터 repository 조회/import(API-88/89) 허용, `sync` = mirror sync(API-72) 허용(`pull` 도 허용), `webhook` = inbound webhook 수신, `push` = outbound 저장소 생성(API-90) 허용. gate 미충족 시 422 `integration_capability_not_enabled`.
>
> import/create 는 현재 **Gitea REST client** 만 구현돼 있어 Gitea-compatible provider(gitea/forgejo/gogs, credentials_ref `provider_sdk:<vendor>` 기준)로 제한된다. 다른 vendor(github/gitlab/bitbucket)는 422 `integration_provider_not_gitea_compatible`.

## 2. Provider Catalog

### 2.1 API-69 `GET /api/v1/integration/providers`

- **인증**: OIDC + RBAC (`infrastructure:view` 또는 `pipelines:view`).
- **응답 — 200**: provider 목록 + `provider_type`, `enabled`, `capabilities`, `sync_status`, `last_sync_at`, `last_error_code`.

### 2.2 API-70 `POST /api/v1/integration/providers`

- **인증**: OIDC + RBAC `infrastructure:edit` (system_admin only).
- **요청**: `provider_key`, `provider_type`, `display_name`, `auth_mode`, `credentials_ref`(inbound webhook 서명 시크릿), `capabilities`, `base_url`(optional, http(s) URL — outbound sync 대상 endpoint, migration 000038), `api_token`(optional — outbound sync(REST pull) 인증용 PAT, **write-only**, migration 000040).
- **`scope` 보강 (2026-05-27)**: provider 는 scope 비종속 **catalog** 이라 현재 `createIntegrationProviderRequest`(`integration_registry.go`)에는 `scope` 필드가 **없다** — 아래 요청 줄/예시의 `scope` 는 초안(API-69..78 임시 발급) 잔재이며 전송돼도 무시된다. scope(application/project) 연결은 binding(API-75 `POST /integration/bindings`) 소관이다.
- **outbound auth 자격증명 (auth_mode 별, migration 000041, 모두 optional)**: `auth_mode` 에 따라 외부 시스템 sync/pull 시 사용하는 자격증명.
  - `token` → `api_token` (PAT). `Authorization: token <pat>`.
  - `basic` / `app_password` → `auth_username` + `auth_secret`. HTTP Basic.
  - `oauth2` → `auth_client_id` + `auth_token_url`(http(s) URL) + `auth_secret`(client_secret). client-credentials grant 후 `Authorization: Bearer`.
  - `agent` → `auth_username`(agent 식별자). 별도 agent 가 인증 (서버 직접 sync 미사용).
  - `auth_secret` 은 **write-only** (api_token 과 동일). 비밀 외 필드(`auth_username`/`auth_client_id`/`auth_token_url`)는 응답 노출.
- **응답 — 201**: 생성된 provider (`base_url`/`auth_username`/`auth_client_id`/`auth_token_url` 포함; `api_token`·`auth_secret` 은 raw 미노출, `api_token_set`/`auth_secret_set`(bool) 만 — 보안).
- **에러**: 409 `integration_provider_conflict`, 400 `invalid_provider_type`, 400 `invalid_base_url`, 400 `invalid_auth_token_url`.
- **참고**: `credentials_ref`(inbound webhook)와 outbound auth 자격증명(`api_token`/`auth_*`)은 별개 시크릿. Phase 3 (sync worker per-provider) 이후 등록 provider 의 `base_url` + auth_mode 별 자격증명이 Gitea sync / SCM repo 연동(API-88/89/90, application-lifecycle §13.9 publish)에 사용된다.
  - **env fallback 금지 (codex #358 P1 / #359)**: 명시 provider 를 대상으로 한 outbound 호출(`scmProviderClient` → `provider.ResolveOutboundAuth()`)은 worker-global env 토큰(`GITEA_TOKEN` 등)으로 **fallback 하지 않는다** — 잘못된 계정/토큰 유출 방지. 등록된 자격증명이 미설정이면 `422 integration_outbound_credentials_missing` 로 거부한다. env fallback 은 provider 미명시(legacy) sync worker 경로에서만 유효하다.

요청 예시:

```json
{
  "provider_key": "jira-main",
  "provider_type": "alm",
  "display_name": "Jira Cloud (Main)",
  "auth_mode": "oauth2",
  "credentials_ref": "secret://integrations/jira-main",
  "capabilities": ["issue.read", "epic.read", "issue.link"],
  "scope": {
    "scope_type": "project",
    "scope_id": "PRJ-001"
  }
}
```

응답 예시:

```json
{
  "status": "created",
  "data": {
    "provider_id": "8f8cdb8d-c690-458f-a243-a8b8b67f9a4d",
    "provider_key": "jira-main",
    "provider_type": "alm",
    "display_name": "Jira Cloud (Main)",
    "enabled": true,
    "auth_mode": "oauth2",
    "capabilities": ["issue.read", "epic.read", "issue.link"],
    "sync_status": "requested",
    "last_sync_at": null,
    "last_error_code": null,
    "created_at": "2026-05-15T14:00:00Z",
    "updated_at": "2026-05-15T14:00:00Z"
  }
}
```

### 2.3 API-71 `PATCH /api/v1/integration/providers/{provider_id}`

- **인증**: OIDC + RBAC `infrastructure:edit` (system_admin only).
- **요청**: `enabled`, `display_name`, `capabilities`, `credentials_ref`, `base_url`, `api_token`, `auth_username`, `auth_client_id`, `auth_token_url`, `auth_secret` 일부 수정 (전송된 키만 patch). `auth_mode` 는 등록 시 고정 — 변경 불가. write-only secret(`api_token`/`auth_secret`)은 blank/미전송 시 기존 값 유지.
- **응답 — 200**: 수정된 provider.
- **에러**: 400 `invalid_base_url`, 400 `invalid_auth_token_url`.

### 2.4 API-72 `POST /api/v1/integration/providers/{provider_id}/sync`

- **인증**: OIDC + RBAC `infrastructure:edit` (system_admin only).
- **설명**: provider 단위 수동 reconciliation job enqueue.
- **capability gate**: provider 가 `pull` 또는 `sync` capability 를 선언해야 한다. 미충족 시 422 `integration_capability_not_enabled`.
- **응답 — 202**: `{status:"accepted", job_id:"..."}`.
- **에러**: 422 `integration_sync_unsupported_provider_type` (비-SCM), 422 `integration_capability_not_enabled`.

### 2.5 API-88 `GET /api/v1/integration/providers/{provider_id}/scm-repositories`

- **인증**: OIDC + RBAC `infrastructure:view` (system_admin only).
- **설명**: SCM provider(`provider_type=scm`)의 base_url + outbound 자격증명으로 원격 repository 목록을 조회한다. 각 항목에 시스템 import 여부(`imported`)를 표시 (provider_id 로 연동된 시스템 repository 존재 여부).
- **capability gate**: `pull` 필요.
- **응답 — 200**: `{status:"ok", data:[{full_name, name, clone_url, html_url, default_branch, private, imported}], meta:{total}}`.
- **에러**: 404 `integration_provider_not_found`, 409 `integration_provider_disabled`(비활성), 422 `integration_sync_unsupported_provider_type`(비-SCM) / `integration_capability_not_enabled`(pull 없음) / `integration_provider_not_gitea_compatible` / `integration_base_url_missing` / `integration_outbound_credentials_missing`, 502 `integration_scm_unreachable`/`integration_scm_auth_failed`.

### 2.6 API-89 `POST /api/v1/integration/providers/{provider_id}/import-repositories`

- **인증**: OIDC + RBAC `infrastructure:edit` (system_admin only).
- **요청**: `{full_names: ["owner/repo", ...]}` (import 할 원격 repository full_name 목록).
- **설명**: 선택한 원격 repository 를 시스템 `repositories` 로 import/연동한다. **신뢰 가능한 SCM 데이터를 쓰기 위해 요청 payload 가 아니라 SCM 에서 다시 조회한 값**으로 upsert 한다. import 된 repository 는 `source=scm`, `provider_id` 세팅, SCM mirror 필드(clone_url/default_branch/private 등)는 이후 sync 가 갱신하고 시스템 소유 메타(`description`)는 보존된다 (소유권 분리, migration 000042).
- **capability gate**: `pull` 필요.
- **응답 — 200**: `{status:"ok", imported:N, repositories:[{full_name, name}], not_found:["..."]}`. (선택했으나 원격에 없는 full_name 은 `not_found`.)
- **에러**: 400 `integration_import_no_selection`(빈 목록), 그 외 API-88 과 동일.

### 2.7 API-90 `POST /api/v1/integration/providers/{provider_id}/create-repository`

- **인증**: OIDC + RBAC `infrastructure:edit` (system_admin only).
- **요청**: `{name(필수), owner(optional org — 비우면 인증 계정), description, private(bool), auto_init(bool)}`.
- **설명**: 시스템에서 선택 SCM(provider)에 **실제 저장소를 생성**하고 (Gitea `POST /user/repos` 또는 `/orgs/{owner}/repos`) 시스템 `repositories` 로 미러한다. 생성된 row 는 **`source=system`** (시스템이 생성을 주도) + `provider_id` 세팅 + SCM 응답값으로 mirror 필드 채움. 이후 sync 가 mirror 필드를 갱신해도 source/description 는 보존.
- **capability gate**: `push` 필요 + Gitea-compatible provider.
- **응답 — 201**: `{status:"created", repository:{full_name, name, clone_url, html_url, default_branch, private, source:"system"}}`.
- **에러**: 400 `integration_repo_name_required`, 409 `integration_provider_disabled`(비활성), 422 `integration_capability_not_enabled`(push 없음) / `integration_provider_not_gitea_compatible` / `integration_base_url_missing` / `integration_outbound_credentials_missing`, 502 `integration_scm_create_failed`(SCM 생성 실패 — 예: 이미 존재 409).

### 2.8 API-80 `DELETE /api/v1/integration/providers/{provider_id}`

- **인증**: OIDC + RBAC `infrastructure:delete` (system_admin only).
- **설명**: Provider 삭제. `integration_sync_jobs` 는 `ON DELETE CASCADE` 로 자동 정리. `integration_bindings` 는 **FK guard** — 1건 이상 존재 시 명시 차단 (실수 cascade 방지).
- **응답**:
  - 200 `{status:"ok"}` — 정상 삭제.
  - 404 `{status:"not_found", code:"integration_provider_not_found"}` — 미존재.
  - 409 `{status:"conflict", code:"integration_provider_has_bindings"}` — 활성 binding 존재. 운영자가 binding 삭제 후 재시도.
- **audit**: `integration.provider.deleted` + payload `{provider_key, provider_type, display_name}`.
- **운영 메모**: cascade binding 정리는 별도 ADR 후보 (1차 정책은 명시 차단).

### 2.9 API-87 `POST /api/v1/integration/test-connection`

- **인증**: OIDC + RBAC `infrastructure:edit` (system_admin only).
- **설명**: provider 등록 전/후 외부 시스템 endpoint reachability 검증 (등록 UX 고도화 #5). 저장된 provider 가 아니라 body 의 `base_url` 을 직접 GET (pre-save 가능). reachability 만 확인하며 자격증명 검증은 후속.
- **요청**: `{ "base_url": "https://gitea.example.com" }` (http(s) 필수).
- **동작**: GET + 5s timeout + redirect 미추적. 응답 본문은 미반환 (status_code / latency 만).
- **응답**:
  - 200 `{status:"ok", reachable:true, status_code, latency_ms}` — 도달.
  - 200 `{status:"ok", reachable:false, latency_ms, error}` — 미도달 (테스트 자체는 수행됨).
  - 400 `{status:"rejected", code:"invalid_base_url"}` — base_url 누락 또는 비-http(s).
- **보안**: SSRF — 합법적 대상이 사내 internal endpoint (Gitea/Jenkins 등) 이므로 internal IP 차단 안 함. admin 신뢰 경계 + 짧은 timeout + 본문 미반환으로 표면 최소화.

## 3. Ingest / Binding

### 3.1 API-73 `POST /api/v1/integration/providers/{provider_key}/webhook`

- **인증**: provider별 webhook 인증(header signature/token); OIDC 미적용.
- **설명**: raw event 저장 + 검증 + normalize enqueue.
- **응답**: 202 accepted / 401 invalid signature / 409 duplicate delivery.
- **검증 확장성**:
  - `Adapter Router` 가 provider별 verifier 전략을 선택해 검증한다.
  - verifier contract 예: `Verify(headers, body) -> (ok, reason)` (`hmac_sha256`, `shared_token`, `provider_sdk` 등).
- **헤더(권장 공통)**:
  - `X-Integration-Delivery`: 외부 전송 고유 ID (없으면 payload hash로 보조 dedupe)
  - `X-Integration-Event`: 이벤트 타입
  - `X-Integration-Signature`: provider 정책 기반 서명값
- **provider-native 헤더 alias**: 외부 시스템은 DevHub-native `X-Integration-*` 를 보내지 않으므로, 각 항목은 provider 고유 헤더로 fallback 한다. 현재 수용: Gitea/Forgejo `X-Gitea-Signature`/`X-Gitea-Event`/`X-Gitea-Delivery`, Gogs `X-Gogs-Signature`/`X-Gogs-Event`/`X-Gogs-Delivery`. 우선순위는 `X-Integration-*` → `X-Gitea-*` → `X-Gogs-*`. 서명 값 자체는 provider 무관 HMAC-SHA256 으로 검증한다(`hmac_sha256:` / `provider_sdk:` 전략).

### 3.2 API-74 `GET /api/v1/integration/bindings`

- **인증**: OIDC + RBAC view.
- **쿼리**: `scope_type`, `scope_id`, `provider_type`, `enabled`, `limit`, `offset`.
- **응답 — 200**: binding 목록 + pagination meta.

### 3.3 API-75 `POST /api/v1/integration/bindings`

- **인증**: OIDC + RBAC edit (system_admin only).
- **요청**: `scope_type` (`application|project`), `scope_id`, `provider_id`, `external_key`, `policy`.
- **응답 — 201**: 생성 binding.
- **에러**: 409 `integration_binding_conflict`, 422 `integration_policy_violation`.

요청 예시:

```json
{
  "scope_type": "application",
  "scope_id": "APP-001",
  "provider_id": "8f8cdb8d-c690-458f-a243-a8b8b67f9a4d",
  "external_key": "PROJ",
  "policy": "execution_system"
}
```

## 4. HomeLab Infra

### 4.1 API-76 `GET /api/v1/infra/services`

- **인증**: OIDC + RBAC `infrastructure:view`.
- **응답 — 200**: 서비스 인벤토리(`service_id`, `node_id`, `name`, `version`, `port`, `health_status`, `observed_at`).

### 4.2 API-77 `POST /api/v1/infra/services/snapshot`

- **인증**: Agent 토큰 기반 ingest 인증 (OIDC 미적용).
- **요청**: 노드/서비스 상태 스냅샷 배열.
- **응답 — 202**: 수집 accepted + ingest_id.
- **영속화 정책 (baseline)**:
  - ingest payload(`nodes`, `services`)는 `infra_service_snapshots`에 저장한다.
  - 동일 프로세스 런타임 캐시가 비어 있으면 최신 persisted snapshot을 hydrate해 조회 응답에 사용한다.

요청 예시:

```json
{
  "agent_id": "homelab-agent-a",
  "snapshot_at": "2026-05-15T14:10:00Z",
  "trace_id": "trc_01jv7w2mm4m7",
  "nodes": [
    {
      "node_id": "node-nas-01",
      "hostname": "nas-01.local",
      "ip_address": "192.168.0.20",
      "environment": "homelab",
      "status": "stable",
      "metrics": { "cpu_percent": 21.3, "mem_percent": 63.1, "disk_percent": 57.2 },
      "observed_at": "2026-05-15T14:09:58Z"
    }
  ],
  "services": [
    {
      "service_id": "svc-jenkins",
      "node_id": "node-nas-01",
      "name": "jenkins",
      "version": "2.504.1",
      "port": 8080,
      "health_status": "healthy",
      "metadata": { "runtime": "docker", "compose_project": "ci-stack" },
      "observed_at": "2026-05-15T14:09:59Z"
    }
  ]
}
```

### 4.3 API-78 `GET /api/v1/infra/topology/v2`

- **인증**: OIDC + RBAC `infrastructure:view`.
- **응답 — 200**: `nodes`, `edges`, `services`, `meta`(`snapshot_at`, `degraded_providers`).
- **상태 반영 (baseline)**:
  - `meta.snapshot_at`: 최신 ingest 시각 (런타임 캐시 또는 persisted snapshot 기준).
  - `meta.degraded_providers`: snapshot 서비스 상태가 `degraded|down`인 provider 집합.

## 5. 공통 에러 코드 (초안)

```
integration_provider_conflict
integration_provider_not_found
integration_provider_disabled
integration_binding_conflict
integration_binding_not_found
integration_policy_violation
integration_webhook_signature_invalid
integration_event_duplicate
integration_sync_job_rejected
infra_snapshot_invalid
infra_agent_unauthorized
```

## 6. 값 제약 (draft)

- `provider_type`: `alm | scm | ci_cd | doc | infra`
- `sync_status`: `requested | verifying | active | degraded | disconnected`
- `binding.policy`: `summary_only | execution_system | bidirectional_candidate`
- `infra.health_status`: `healthy | degraded | down`

## 7. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-29 | Phase 3 split — master `docs/backend_api_contract.md` §15 (Integration 본문) 을 도메인 sub-document 로 이관. ID(API-69..78, API-80, API-87..90) 보존. Task Ingestion 은 `task_api.md` 로 분리. |
