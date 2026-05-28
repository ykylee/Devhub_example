# `docs/backend_api_contract.md` ↔ 라우트/핸들러 대조 분석

- 문서 목적: `docs/backend_api_contract.md` 의 API 인벤토리를 현재 코드(`backend-core/internal/httpapi/router.go` + 핸들러)와 전수 대조해, 누락·불일치·stale 항목을 근거와 함께 기록한다. 본 분석은 동일 sprint 의 contract 갱신(Edit) 근거 문서다.
- 분석 기준 시점: 2026-05-27 (main `cf19c94`).
- 1차 근거: `backend-core/internal/httpapi/router.go`(라우트 등록), `domain.go`(repository draft/publish), `integration_registry.go`(API-69..75/80/87), `integration_scm_repositories.go`(API-88/89/90). 라우트 인벤토리는 `docs/analysis/2026-05-27-codebase-snapshot/code/backend/httpapi.md` §3 와 교차검증.
- 결론 요약: contract §15(Integration)는 **이미 매우 최신** — API-87/88/89/90 spec + API-70/71 auth_mode 자격증명 + write-only(`*_set`) 규칙 + API-73 헤더 alias 모두 기재돼 있다. 실제 **유일한 구조적 누락은 repository draft→publish 2 endpoint**(`POST /repositories`, `POST /repositories/:id/publish`)다. 그 외 2건은 코드와 어긋난 stale prose(env fallback 서술 / 제거된 `scope` 필드).

---

## 1. 대조 대상 라우트 인벤토리 (router.go)

본 분석이 다루는 범위(컨텍스트로 지정된 API-70/71/73/87/88/89/90 + repository draft/publish)에 한정한 라우트표.

| # | 메서드 | 경로 | 핸들러 | RBAC (routePermissionTable) | contract 본문 | 상태 |
| --- | --- | --- | --- | --- | --- | --- |
| 1 | POST | `/api/v1/repositories` | `createRepositoryDraft` | `application_repositories:create` | **없음** | **누락 — 신규 추가 필요** |
| 2 | POST | `/api/v1/repositories/:repository_id/publish` | `requestRepositoryPublish` | `application_repositories:edit` | **없음** | **누락 — 신규 추가 필요** |
| 3 | POST | `/api/v1/integration/providers` | `createIntegrationProvider` | `infrastructure:edit` | §15.2 API-70 | 기재됨(보강 1건) |
| 4 | PATCH | `/api/v1/integration/providers/:provider_id` | `updateIntegrationProvider` | `infrastructure:edit` | §15.2 API-71 | 기재됨(정합) |
| 5 | POST | `/api/v1/integration/providers/:provider_id/webhook` | `ingestIntegrationProviderWebhook` | Bypass(public path)+서명 | §15.3 API-73 | 기재됨(헤더 alias 포함, 정합) |
| 6 | POST | `/api/v1/integration/test-connection` | `testIntegrationConnection` | `infrastructure:edit` | §15.2 API-87 | 기재됨(정합) |
| 7 | GET | `/api/v1/integration/providers/:provider_id/scm-repositories` | `listSCMRepositories` | `infrastructure:view` | §15.2 API-88 | 기재됨(정합) |
| 8 | POST | `/api/v1/integration/providers/:provider_id/import-repositories` | `importSCMRepositories` | `infrastructure:edit` | §15.2 API-89 | 기재됨(정합) |
| 9 | POST | `/api/v1/integration/providers/:provider_id/create-repository` | `createSCMRepository` | `infrastructure:edit` | §15.2 API-90 | 기재됨(정합) |

> RBAC 컬럼은 `permissions.go::routePermissionTable` 값(httpapi.md §3 교차검증). API-70/71/87/89/90 은 모두 `infrastructure:edit`, API-88 은 `infrastructure:view`, API-80 은 `infrastructure:delete`. 등록되지 않은 v1 라우트는 `enforceRoutePermission` 의 deny-by-default(403 `auth_policy_unmapped`)로 fail-loud — silent 불일치는 구조적으로 차단됨.

---

## 2. 누락 (contract 에 spec 없음 — 신규 추가)

### M-1. Repository draft→publish 2 endpoint (HIGH — 구조적 누락)

`POST /api/v1/repositories`(`createRepositoryDraft`)와 `POST /api/v1/repositories/:repository_id/publish`(`requestRepositoryPublish`)가 라우터에 등록·구현돼 있으나 contract 전 절에서 spec 부재. grep 결과 `repositories/:repository_id/publish`/`createRepositoryDraft`/`requestRepositoryPublish` 0건. #368(draft→publish lifecycle, codex 머지) + #373(provider_id 단일화)로 도입.

- 근거: `router.go:238-239`; `domain.go:141-287`.
- 코드 동작(`domain.go`):
  - **draft 생성**(`createRepositoryDraft`, `domain.go:141`): 요청 `{key, slug, provider_key?}`. `provider_key` 가 주어지면 `GetIntegrationProviderByKey` 로 `integration_providers` FK(`provider_id`)로 해석(migration 000045 — 구 `scm_provider` 통합). provider 가 SCM type 아니면 422 `integration_sync_unsupported_provider_type`. 빈 값이면 provider 미지정 draft. `CreateRepositoryDraft(key, slug, providerID)` → `201 {status:"ok", data: repositoryResponse}`. key/slug 누락 시 400, 충돌 시 409 conflict, DomainStore 가 `repositoryDraftStore` 미충족이면 503.
  - **publish 요청**(`requestRepositoryPublish`, `domain.go:191`): path `repository_id`(positive int). draft 만 허용(아니면 409 conflict). `provider_id` 필수(없으면 400 `integration_provider_required`). provider lookup → SCM type + push capability + gitea-compatible 검사(각 422) → `scmProviderClient` → `gitea.CreateRepo`(owner/name/description/private, AutoInit=true, DefaultBranch="main") → 성공 시 `UpsertRepository(source=system, provider_id)` + reload → `200 {status:"ok", data}`. **SCM 생성 실패 시 `MarkRepositoryDraftPublishRequested` 만 호출하고 502 `integration_scm_create_failed`** (부분 실패 경로 — httpapi.md F-1: 무테스트 머지라 검증 공백).
- 응답 schema(`repositoryResponse`, `domain.go:17-33`): `id, gitea_repository_id?, full_name, owner_login?, name, clone_url?, html_url?, default_branch?, private, status, provider_id?, provider_key?, publish_requested_at?, published_at?, updated_at`. `status` 는 `draft|published`(repository_status). `provider_key` 는 Get/List 가 `LEFT JOIN integration_providers` 로 derive(표시용, migration 000045).
- **권고**: §13 에 신규 sub-section(예 §13.9 Repository Draft/Publish) + API ID **API-91**(POST /repositories) / **API-92**(POST /repositories/:id/publish) 발급. §13.0 인덱스 row 추가.

---

## 3. 불일치 / stale (기재돼 있으나 코드와 어긋남 — 보강/정정)

### D-1. API-70 §15.2 의 "env fallback, token mode" 서술 (MEDIUM — stale, 보안 오해 소지)

§15.2 API-70 의 "**참고**" 줄(`2012`)이 *"Phase 3 (sync worker per-provider) 이후 등록 provider 의 base_url + auth_mode 별 자격증명이 Gitea sync 에 사용됨 (**env fallback, token mode**)"* 라고 적었다. 그러나 명시 provider 의 outbound(`scmProviderClient`, `integration_scm_repositories.go:88`)는 `provider.ResolveOutboundAuth()` 만 쓰고 **worker-global env 토큰 fallback 을 금지**한다(codex #358 P1 / #359 — 잘못된 계정/토큰 유출 방지). 미설정 시 422 `integration_outbound_credentials_missing`. env fallback 은 provider 미명시(legacy) sync worker 경로에서만 유효.

- 근거: `integration_scm_repositories.go:88-103`; 메모리 `feedback_no_env_credential_fallback_for_explicit_provider`.
- **권고**: §15.2 API-70 참고 줄에서 "env fallback" 표현이 명시 provider outbound 에 적용되는 것처럼 읽히지 않도록 정정 — 명시 provider 는 등록된 자격증명만 사용, 미설정 시 거부(422)임을 명시.

### D-2. API-70 §15.2 요청/예시의 `scope` 필드 (LOW — 코드에 없는 필드)

§15.2 API-70 요청 줄(`2003`)과 예시(`2024-2027`)가 `scope: {scope_type, scope_id}` 를 포함한다. 그러나 실제 `createIntegrationProviderRequest`(`integration_registry.go:221-236`)에는 `scope` 필드가 없다 — provider 는 scope 비종속(catalog), scope 는 binding(API-75)에만 존재. provider create 요청에 `scope` 를 보내도 무시된다.

- 근거: `integration_registry.go:221-236`(필드: provider_key/provider_type/display_name/auth_mode/credentials_ref/capabilities/base_url/api_token/auth_username/auth_client_id/auth_token_url/auth_secret) vs `2003`/`2024-2027`.
- **영향도 LOW**: 초안 시점(API-69..78 임시 발급) 잔재. 기존 ID/prose 삭제 금지 원칙상 본 sprint 에선 **보강 노트로만 명시**(scope 는 provider 가 아니라 binding 소관) 권장 — 예시 본문 제거는 별도 정리 sprint 후보.

### D-3. API-69 RBAC 표기 "infrastructure:view 또는 pipelines:view" (LOW — 부정확)

§15.2 API-69(`1997`)가 RBAC 를 "`infrastructure:view` 또는 `pipelines:view`" 로 적었으나 routePermissionTable 상 `GET /integration/providers` 는 `infrastructure:view` 단독(httpapi.md §3:169). `pipelines:view` 는 적용되지 않음.

- 근거: `permissions.go::routePermissionTable`(httpapi.md §3); `router.go:314`.
- **영향도 LOW**: 본 sprint 컨텍스트(API-70/71/73/87/88/89/90 + draft/publish) 밖 — 정정은 선택. 기록만 남김.

---

## 4. 정합 확인 (코드와 일치 — 변경 불요)

| 항목 | contract 위치 | 코드 근거 | 비고 |
| --- | --- | --- | --- |
| API-87 test-connection (method/path/요청/응답/SSRF 수용/에러) | §15.2 `2105-2115` | `integration_registry.go:418-461` | reachable bool + status_code + latency_ms, 400 `invalid_base_url`. 정합 |
| API-88 scm-repositories (응답 필드/capability=pull/에러 매트릭스) | §15.2 `2068-2074` | `integration_scm_repositories.go:134-177` | `imported` 플래그 + 404/409/422/502 코드 모두 정합 |
| API-89 import-repositories (요청 `full_names`/SCM 재조회 upsert source=scm/not_found/에러) | §15.2 `2076-2083` | `integration_scm_repositories.go:179-260` | `{imported, repositories, not_found}` 응답 정합 |
| API-90 create-repository (요청/source=system/push gate+gitea-compat/응답 201/에러) | §15.2 `2085-2092` | `integration_scm_repositories.go:262-338` | 정합 |
| API-70 auth_mode 자격증명 필드 (`base_url`/`api_token`(write-only)/`auth_username`/`auth_client_id`/`auth_token_url`/`auth_secret`(write-only)) | §15.2 `2003-2012` | `integration_registry.go:221-321`, response `:28-52` | `api_token_set`/`auth_secret_set` bool 응답 규칙 정합 (D-1/D-2 외) |
| API-71 PATCH (전송 키만 patch / auth_mode 변경 불가 / write-only blank=keep) | §15.2 `2056` | `integration_registry.go:323-416` | 정합. auth_mode 는 update 요청 구조체에 부재(고정) |
| API-73 헤더 alias (`X-Integration-*`→`X-Gitea-*`→`X-Gogs-*` fallback, HMAC 무관 검증) | §15.3 `2127-2131` | `integration_registry.go:596-617`(`firstHeader`) | 우선순위/서명전략 정합 |
| capability 기능 gate (pull/sync/push, 422 `integration_capability_not_enabled`) | §15.1 `1989-1991` | `integration_scm_repositories.go:33-78` | gitea-compat 제약 포함 정합 |
| API-80 DELETE provider FK guard (409 `integration_provider_has_bindings`) | §15.2 `2094-2103` | `integration_registry.go:518-574` | 정합 |

> 컨텍스트가 "spec 누락/미흡 가능" 으로 지목한 API-87/88/89/90, API-70/71 auth_mode, API-73 헤더 alias 는 **이미 §15 에 충실히 기재됨**. 본 sprint 갱신의 실질 작업은 (a) repository draft/publish 신규 추가(M-1), (b) D-1/D-2 stale prose 보강. (c) 메타 헤더 날짜/이력 갱신.

---

## 5. 갱신 계획 (contract Edit 범위 — 추가/보강만, 삭제·재번호 금지)

1. **§13** 신규 sub-section **§13.9 Repository Draft → Publish** 추가 — API-91/92 발급, method/path/RBAC/요청·응답 schema/에러코드/capability·gitea-compat gate/부분실패 경로 명시. §13.0 인덱스에 API-91/92 row 추가. §13.8 공통 에러 코드에 신규 코드(`integration_provider_required`, `integration_scm_create_failed` 등) 누락분 보강.
2. **§15.2 API-70** 참고 줄 D-1 정정(env fallback → 명시 provider 는 등록 자격증명만, 미설정 422) + D-2 보강 노트(`scope` 는 binding 소관).
3. 메타 헤더 `최종 수정일` → 2026-05-27 + 변경 이력 1줄.
