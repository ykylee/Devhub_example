---
title: architecture
type: source
tags: [domain, architecture.md, project-devhub]
sources: [raw/projects/devhub/docs/domain/repository-integration/architecture.md]
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

# repository-integration 도메인 아키텍처

- 문서 목적: SCM↔시스템 Repository 소유권 분리·연동·lifecycle 아키텍처를 정의한다.
- 범위: ARCH-REPO-01..07. 일반 Integration architecture / capability gate 일반 정책은 `docs/domain/integration-registry/architecture.md` (ARCH-INT-01..07) 참조.
- 대상 독자: backend / 프론트엔드 / DevOps, AI agent, 아키텍처 검토자.
- 상태: accepted
- 최종 수정일: 2026-05-29 (Phase 3 split, master `docs/architecture.md` §10 본문 이관)
- 관련 문서: [도메인 README](./README.md), [requirements.md](./requirements.md), [api.md](./api.md), [master architecture](../../architecture.md), [integration-registry architecture](../integration-registry/architecture.md), [platform-lifecycle architecture](../platform-lifecycle/architecture.md)

## 개요

DevHub `repositories` 는 (a) 외부 SCM(Gitea 등)에서 webhook/pull 로 미러된 row 와 (b) DevHub 운영자가 시스템 내에서 직접 생성·관리하는 row 가 같은 테이블에 공존한다. 본 도메인은 이 두 출처를 구분하는 소유권 모델, 양방향(import/create) 연동, draft→publish lifecycle, 그리고 SCM provider 참조의 canonical 단일화를 정의한다. 도입 PR: #363(소유권 분리 + import) / #366(create) / #368(draft→publish) / #371(충돌 정정) / #373(provider_id 단일화). 관련 마이그레이션: 000042 / 000043 / 000044 / 000045.

## 1. 소유권 분리 모델 (ARCH-REPO-01)

`domain.Repository` 는 SCM mirror 필드와 시스템 메타를 분리한다.

```text
repositories  (소유권/연동 관련 컬럼만 발췌)
  source            text      -- 'scm' | 'system' (빈값 = legacy 'scm' 취급)
  provider_id       uuid  FK  -- integration_providers(provider_id) (scm type), canonical SCM 참조
  provider_key      text      -- (read-only, LEFT JOIN integration_providers 로 derive — 표시용)
  description       text      -- system-owned 메타 (SCM sync 가 절대 덮어쓰지 않음)
  repository_status text      -- 'draft' | 'active' (CHECK repositories_status_check)
  publish_requested_at, published_at  timestamptz NULL
```

- `source='scm'`: 외부 SCM 이 원천(SoT). webhook/pull sync 가 mirror.
- `source='system'`: DevHub 가 원천. draft 생성 또는 outbound create(§4)로 발생.
- `provider_id` 가 SCM provider 참조의 **단일 출처(FK)**, `provider_key` 는 사람이 읽기 위한 derived 값일 뿐 저장 식별자가 아니다(§7).

## 2. SCM mirror vs system-owned 필드 보존 (ARCH-REPO-02)

`UpsertRepository` 의 `ON CONFLICT (full_name) DO UPDATE` 는 sync 가 외부 값으로 덮어써도 되는 **SCM mirror 필드만** `EXCLUDED` 로 갱신하고, system-owned 필드는 보존한다.

- 덮어쓰는 SCM mirror: `owner_login / name / clone_url / html_url / default_branch / private / gitea_repository_id`.
- 보존(기존 우선): `source = COALESCE(기존, EXCLUDED)`, `provider_id = COALESCE(기존, EXCLUDED)`.
- 보존(SET 절에서 아예 제외): `description` — system-owned 메타라 sync 가 절대 갱신하지 않음.
- INSERT 분기는 `source = COALESCE(NULLIF($n,''),'scm')`, `repository_status='active'`, `published_at=NOW()` 로 채운다.

이 규약 덕분에 운영자가 SCM-mirror row 에 부여한 분류/설명 메타가 다음 sync 에 의해 유실되지 않는다. in-memory fake 도 동일 미러 보존을 흉내내 production parity 를 맞춘다.

## 3. inbound import (ARCH-REPO-03)

운영자가 등록된 SCM provider 의 원격 저장소를 DevHub 로 끌어오는 경로.

- `GET /api/v1/integration/providers/:provider_id/scm-repositories` (API-88) — provider 의 원격 repo 목록 + DevHub 내 import 여부(`imported` 플래그, `ListRepositoriesByProvider`).
- `POST /api/v1/integration/providers/:provider_id/import-repositories` (API-89) — 선택 repo 를 SCM 에서 **재조회한 값**으로 `UpsertRepository`(`source='scm'`). request body 의 stale 값이 아니라 provider 에 직접 조회한 결과를 신뢰한다.
- 권한: `infrastructure:edit`. capability gate: `pull`(§6).

## 4. outbound create (ARCH-REPO-04)

DevHub 에서 신규 저장소를 외부 SCM 에 생성하는 경로(시스템→SCM).

- `POST /api/v1/integration/providers/:provider_id/create-repository` (API-90) — `gitea.Client.CreateRepo`(owner 비면 `POST /user/repos`, 있으면 `POST /orgs/{owner}/repos`)로 원격 생성 후 DevHub row 를 `source='system'` 으로 기록.
- 권한: `infrastructure:edit`. capability gate: `push`(§6) + provider 가 gitea-compatible vendor 여야 함(`isGiteaCompatibleProvider` — 비-gitea vendor 거부).
- §3 + §4 로 SCM ↔ 시스템 repository **양방향(import + create) 연동**이 완성된다.

## 5. draft→publish 상태머신 (ARCH-REPO-05)

system-owned repository 는 외부 SCM 에 즉시 만들지 않고 DevHub 내 draft 로 먼저 등록한 뒤 publish 시점에 SCM 에 생성할 수 있다.

```
  POST /api/v1/repositories
  (createRepositoryDraft)
        │  source='system', repository_status='draft'
        │  provider_key → provider_id FK 해석 (migration 000045)
        ▼
  ┌──────────────┐    POST /api/v1/repositories/:id/publish    ┌──────────────┐
  │   draft      │ ─── (requestRepositoryPublish, draft only) ─▶│   active     │
  │              │     provider SCM/push/gitea-compat 검사       │              │
  └──────────────┘     → gitea.CreateRepo → UpsertRepository     └──────────────┘

  (SCM webhook/pull sync 로 인입되는 row 는 draft 를 거치지 않고 repository_status='active' 직행)
```

- `createRepositoryDraft`(API: `POST /repositories`, RBAC `platform_repositories:create`): `source='system'`, `repository_status='draft'` row INSERT. provider_key 를 provider_id FK 로 해석.
- `requestRepositoryPublish`(API: `POST /repositories/:id/publish`, RBAC `platform_repositories:edit`): `repository_status='draft'` 인 row 만 대상. provider 의 SCM type + push capability + gitea-compat 검사 후 `gitea.CreateRepo` → `UpsertRepository`. SCM 생성 실패 시 `MarkRepositoryDraftPublishRequested`(publish_requested_at set) 후 502(BadGateway) 반환하는 부분 실패 경로가 있다.
- **검증 공백(부채)**: draft→publish 핸들러·store 메서드(`CreateRepositoryDraft`/`MarkRepositoryDraftPublishRequested`)는 #368(codex)이 **무테스트로 머지**했고 #373 이 그 위를 수정 — 단위/통합 테스트 보강이 후속 directive 다.

## 6. capability gate (ARCH-REPO-06)

SCM 연동 endpoint(import/create/sync)는 공통 게이트 `scmProviderForCapability` 를 통과해야 한다.

- 게이트 검사: provider **exists** + **enabled**(disabled provider 는 409 거부, #371 정정) + `provider_type='scm'` + 요청 capability 보유 + gitea-compat.
- capability ↔ 기능 매핑: `import` = `pull`, `sync` = `pull | sync`, `create` = `push`.
- 이로써 provider 가 선언한 capability 범위 밖의 동작이 차단된다(예: pull-only provider 에 create 거부).

## 7. provider_id 단일화 (ARCH-REPO-07)

도입 과정에서 SCM provider 참조가 두 컬럼으로 중복됐다 — #368 의 `scm_provider`(provider_key TEXT, migration 000043 ADD) 와 #363 의 `provider_id`(FK UUID, migration 000042). 동일 SCM 참조를 의미 중복하던 것을 #373(migration 000045)이 정리했다.

- **canonical = `provider_id`(FK)**. `scm_provider` 컬럼은 provider_key→provider_id backfill 후 DROP.
- 표시용 `provider_key` 는 저장하지 않고 `GetRepositoryByID`/`ListRepositories` 가 `LEFT JOIN integration_providers` 로 derive.
- 패턴: **중복 식별 컬럼은 FK 를 canonical 로 두고 readable key 는 join 으로 derive**(SCM-owned vs system-owned 보존 규약 §2 와 인접 원칙).
- 부수 메모: project-companion 흐름의 `RepositoryCreatePayload.SCMProvider` 는 placeholder 로 별개라 유지된다(draft→publish 의 provider_id 해석과 의미가 겹치는 경미한 부채).
- 운영 주의: `scm_provider` 는 000043 ADD → 000045 DROP 으로 2 마이그레이션만 존재한 short-lived 컬럼이다. 000045 의 down 은 컬럼 재추가 + provider_id→provider_key best-effort backfill(매칭 실패 시 NULL)로 비대칭이므로 rollback 시 사전 점검이 필요하다.

## 8. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-29 | Phase 3 split — master `docs/architecture.md` §10 (Repository 본문) 을 도메인 sub-document 로 이관. ID(ARCH-REPO-01..07) 보존, 신규 발급/삭제 없음. |
