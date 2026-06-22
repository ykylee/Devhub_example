---
title: api
type: source
tags: [domain, api.md, project-devhub]
sources: [raw/projects/devhub/docs/domain/repository-integration/api.md]
git_commit: fb3894f7
git_branch: chore/260622-wiki-drift-cleanup-3
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
last_touched: 2026-06-22T06:03:34Z
mirror_dirty: |
related: [none]
status: draft
contradictions: [none]
---

# repository-integration 도메인 API

- 문서 목적: Repository 운영 지표 조회 + Repository draft→publish lifecycle API 계약을 정의한다.
- 범위: API-51..54 (repository 운영 지표) + API-91/92 (draft→publish). 외부 SCM provider 의 원격 repository import/create endpoint(API-88/89/90)는 `docs/domain/integration-registry/api.md` 참조. application/project ↔ repository 연결 endpoint(API-55/56A/56B)는 `docs/domain/platform-lifecycle/api.md` 참조.
- 대상 독자: backend / 프론트엔드 / DevOps, AI agent, QA.
- 상태: accepted
- 최종 수정일: 2026-05-29 (Phase 3 split, master §13.4 + §13.9 본문 이관)
- 관련 문서: [도메인 README](./README.md), [requirements.md](./requirements.md), [architecture.md](./architecture.md), [master API](../../backend_api_contract.md), [integration-registry api](../integration-registry/api.md), [platform-lifecycle api](../platform-lifecycle/api.md)

## 개요

본 도메인의 API 는 다음 두 묶음으로 구성된다.

1. **Repository 운영 지표 조회**(§2, API-51..54) — Application 대시보드/롤업이 소비하는 read-only 조회 endpoint.
2. **Repository Draft → Publish lifecycle**(§3, API-91/92) — 시스템 주도 등록 흐름.

외부 SCM provider 컨텍스트의 원격 repository 목록 조회/import/생성(API-88/89/90)은 integration-registry 도메인 api 에서 정의된다.

## 1. API ID 인덱스

| API ID | endpoint | 목적 |
| --- | --- | --- |
| API-51 | `GET /api/v1/repositories/{repository_id}/activity` | commit/contributor/작업 추이 |
| API-52 | `GET /api/v1/repositories/{repository_id}/pull-requests` | PR 타임라인 |
| API-53 | `GET /api/v1/repositories/{repository_id}/build-runs` | 빌드 이력 |
| API-54 | `GET /api/v1/repositories/{repository_id}/quality-snapshots` | 정적분석/품질 점수 |
| API-91 | `POST /api/v1/repositories` | repository draft 생성 (시스템 주도) |
| API-92 | `POST /api/v1/repositories/{repository_id}/publish` | draft → 실제 SCM 생성 + published 전환 |

> Cross-domain endpoints: API-55 (`/repositories/{id}/projects`), API-56A (`/platforms/{id}/projects`), API-56B (`/projects/{id}/repositories`) 의 본문은 [platform-lifecycle api.md](../platform-lifecycle/api.md) 참조. API-88/89/90 의 본문은 [integration-registry api.md](../integration-registry/api.md) 참조.

## 2. Repository 운영 지표 조회

### 2.1 `GET /api/v1/repositories/{repository_id}/activity` (API-51 (planned))

- 설명: commit/contributor/작업 추이 조회.
- 응답 data 필드:
  - `pr_event_count`, `active_contributors`, `build_run_count`, `build_success_rate` (window 내 가중평균)
  - `last_build_status`: build_runs 의 가장 **최근 1건** status (window 무관) — `queued|running|success|failed|cancelled|skipped|unknown`. 없으면 `unknown` (REQ-FR-APPDASH-001 — 단순 % 보다 broken/red 상태 즉시 표기).
  - `last_build_at`: 최근 빌드 `started_at` (RFC3339) 또는 `null`.

### 2.2 `GET /api/v1/repositories/{repository_id}/pull-requests` (API-52 (planned))

- 설명: PR 상태/타임라인/활동 지표 조회.

### 2.3 `GET /api/v1/repositories/{repository_id}/build-runs` (API-53 (planned))

- 설명: 빌드 실행 이력/상태/소요 시간 조회.

### 2.4 `GET /api/v1/repositories/{repository_id}/quality-snapshots` (API-54 (planned))

- 설명: 정적분석/품질 점수/게이트 결과 조회.

## 3. Repository Draft → Publish (API-91 / API-92)

시스템 내에서 repository 를 먼저 **draft** 로 등록한 뒤, 등록된 SCM provider 에 실제 저장소를 생성하며 **published** 로 전환하는 2단계 lifecycle. #368(draft→publish lifecycle, repository_status/publish_* 컬럼 migration 000043) 도입 + #373(provider 참조를 `provider_id` FK 로 단일화, migration 000045). API-90(`create-repository`)이 "provider 컨텍스트에서 즉시 생성+미러" 인 데 반해, 본 흐름은 "draft 로 먼저 잡아두고 별도 단계에서 publish" 하는 시스템 주도 등록 경로다.

- 쓰기 권한: `POST /repositories` = `platform_repositories:create`, `POST /repositories/{repository_id}/publish` = `platform_repositories:edit` (기본 `system_admin`).
- `repository_status`(응답 `status`): `draft` → `published`. draft 상태에서만 publish 가능.
- `provider_key` ↔ `provider_id`: 입력은 사람이 읽는 `provider_key`(예 `gitea-main`)를 받고, 핸들러가 `integration_providers` 의 FK(`provider_id` UUID)로 해석해 저장한다(migration 000045 — 구 `scm_provider` TEXT 통합). 조회/목록 응답은 `LEFT JOIN integration_providers` 로 `provider_key` 를 표시용 derive 한다.

### 3.1 `POST /api/v1/repositories` (API-91)

- 설명: repository **draft** 생성. SCM 에는 아직 아무것도 만들지 않는다 (메타데이터 등록만).
- 요청 body 필드:
  - `key` (required): repository 식별 key.
  - `slug` (required): repository slug.
  - `provider_key` (optional): 등록된 SCM provider 의 key. 주어지면 SCM type provider 로 해석해 `provider_id` FK 로 저장. 빈 값이면 provider 미지정 draft (publish 전 별도 지정 필요).
- 응답 (`201 Created`): `{ "status": "ok", "data": { /* repositoryResponse */ } }`.
- 응답 schema (`repositoryResponse`):
  - `id`, `gitea_repository_id?`, `full_name`, `owner_login?`, `name`, `clone_url?`, `html_url?`, `default_branch?`, `private`, `status`(`draft|published`), `provider_id?`, `provider_key?`(derive), `publish_requested_at?`, `published_at?`, `updated_at`.
- 에러:
  - `400` — body parse 실패, `key`/`slug` 누락.
  - `404 integration_provider_not_found` — `provider_key` 가 등록 provider 와 매칭 안 됨.
  - `422 integration_sync_unsupported_provider_type` — `provider_key` 가 SCM type 이 아님.
  - `409 conflict` — `key` 또는 `slug` 중복.
  - `503` — `DomainStore` 가 draft store(`CreateRepositoryDraft`/`MarkRepositoryDraftPublishRequested`/`GetRepositoryByID`) 미충족 또는 application store 미설정.

요청 예시:

```json
{
  "key": "devhub-core",
  "slug": "devhub-core",
  "provider_key": "gitea-main"
}
```

### 3.2 `POST /api/v1/repositories/{repository_id}/publish` (API-92)

- 설명: draft repository 를 등록된 SCM provider 에 **실제 생성**하고 시스템 미러를 갱신하며 **published** 로 전환한다.
- path param: `repository_id` (positive integer).
- 요청 body: 없음.
- 동작: draft 검증 → provider lookup → SCM type + `push` capability + Gitea-compatible 검사 → `gitea.CreateRepo`(owner=full_name 의 org, name, description, `private`, `default_branch=main`, `auto_init=true`) → 성공 시 `UpsertRepository`(`source=system`, `provider_id` 세팅, SCM 응답값으로 mirror) + reload.
- capability gate: `push` 필요 + Gitea-compatible provider(gitea/forgejo/gogs).
- 응답 (`200 OK`): `{ "status": "ok", "data": { /* repositoryResponse (published) */ } }`.
- 에러:
  - `400` — `repository_id` 가 positive int 아님.
  - `400 integration_provider_required` — draft 에 `provider_id` 미지정.
  - `404 not_found` — draft repository 미존재.
  - `409 conflict` — draft 상태가 아님 (이미 published 등 — draft 만 publish 가능).
  - `404 integration_provider_not_found` — provider_id 가 가리키는 provider 미존재.
  - `422 integration_sync_unsupported_provider_type` / `integration_capability_not_enabled`(push 없음) / `integration_provider_not_gitea_compatible` / `integration_base_url_missing` / `integration_outbound_credentials_missing`.
  - `502 integration_scm_create_failed` — SCM 저장소 생성 실패. **이 경로에서는 `publish_requested_at` 만 기록하고(draft 보존) BadGateway 반환** — 부분 실패 후 재시도 가능.
  - `502 integration_scm_auth_failed` — outbound 자격증명으로 SCM 인증 실패.

> **구현 메모**: `createRepositoryDraft`/`requestRepositoryPublish`(`backend-core/internal/httpapi/domain.go`)는 #368 에서 무테스트로 머지된 후 #373 이 그 위를 수정했다. publish 의 부분 실패 경로(SCM 생성 실패 → `MarkRepositoryDraftPublishRequested` 만 호출) 검증 공백이 알려진 부채다 — 후속 테스트 보강 carve 후보.

## 4. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-29 | Phase 3 split — master `docs/backend_api_contract.md` §13.4 (Repository 운영 지표) + §13.9 (Draft → Publish) 를 도메인 sub-document 로 이관. ID(API-51..54, API-91, API-92) 보존, 신규 발급/삭제 없음. cross-domain endpoint (API-55/56A/56B/88/89/90)은 link 만. |
