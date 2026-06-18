---
title: requirements
type: source
tags: [domain, requirements.md, project-devhub]
sources: [raw/projects/devhub/docs/domain/repository-integration/requirements.md]
git_commit: 6c434887
git_branch: main
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
last_touched: 2026-06-18T12:08:55Z
mirror_dirty: (dirty: uncommitted changes) |
related: [none]
status: draft
contradictions: [none]
---

# repository-integration 도메인 요구사항

- 문서 목적: SCM↔시스템 Repository 연동 + Repository Lifecycle (draft→publish) 도메인의 기능·비기능 요구사항을 정의한다.
- 범위: REQ-FR-REPO-001..005 / REQ-NFR-REPO-001..003. application/project/repository 계층은 `docs/domain/platform-lifecycle/requirements.md` 참조, 외부 Provider 등록/카탈로그는 `docs/domain/integration-registry/requirements.md` 참조.
- 대상 독자: backend / 프론트엔드 / DevOps, AI agent, QA.
- 상태: accepted
- 최종 수정일: 2026-05-29 (Phase 3 split, master `docs/requirements.md` §5.8 본문 이관)
- 관련 문서: [도메인 README](./README.md), [architecture.md](./architecture.md), [api.md](./api.md), [master requirements](../../requirements.md), [integration-registry requirements](../integration-registry/requirements.md), [platform-lifecycle requirements](../platform-lifecycle/requirements.md)

## 1. 개요

본 절(신규 2026-05-27)은 SCM(Gitea 등 외부 형상관리)과 DevHub 시스템 `repositories` 사이의 **소유권 분리 + 양방향 연동(import/create) + draft→publish 생애주기** 요구사항을 정의한다. platform-lifecycle 의 `Application > Repository > Project` 계층 및 integration-registry 의 외부 연동(Integration Provider)을 전제로 하며, 코드는 PR #363(소유권 분리 + import) / #366(outbound create) / #368(draft→publish) / #373(provider_id 단일화) 으로 1차 완성됐다. 근거: [코드베이스 스냅샷](../../analysis/2026-05-27-codebase-snapshot/README.md), [API 계약](./api.md) (API-88/89/90), migration 000042/000043/000045.

> **참고**: `repositories` 의 SCM mirror(commit/PR/build/quality 등 운영지표) 수집 자체는 platform-lifecycle (REQ-FR-APP-004..009) 및 integration-registry (REQ-FR-INT-004) 에서 이미 다뤘다. 본 절은 그 위에 **"누가 repository row 를 소유하는가(SCM-owned vs system-owned) + 시스템↔SCM 양방향 생성/연동 + 게시 생애주기"** 만 추가 정의한다 (기존 ID 와 중복 발급하지 않음).

## 2. 기능 요구사항 (REQ-FR-REPO)

- **REQ-FR-REPO-001 (MVP, 확정):** 시스템 `repositories` 는 소유권 출처(`source`)를 가져야 한다.
    - 값: `scm`(외부 SCM 에서 import/sync 된 mirror) | `system`(DevHub 가 생성을 주도). 빈값/legacy 는 `scm` 로 취급한다.
    - 각 repository 는 어떤 SCM provider 에 귀속되는지 `provider_id`(외부 연동 Provider FK, integration-registry) 로 단일 식별해야 한다. 표시용 provider key 는 join 으로 derive 한다 (식별 컬럼은 FK 를 canonical 로, readable key 는 파생 — migration 000045 단일화).
    - `description` 등 **system-owned 메타데이터**는 SCM 동기화가 덮어쓰지 않고 보존되어야 한다 (소유권 분리, migration 000042).
- **REQ-FR-REPO-002 (MVP, 확정):** 시스템 관리자는 외부 SCM provider 의 원격 repository 목록을 조회([API-88](../integration-registry/api.md))하고 선택 항목을 시스템으로 **import**([API-89](../integration-registry/api.md))할 수 있어야 한다 (inbound).
    - import 는 요청 payload 가 아니라 **SCM 에서 다시 조회한 신뢰 가능한 값**으로 upsert 하며, import 된 row 는 `source=scm` + `provider_id` 가 세팅된다.
    - 목록 응답은 각 원격 repository 의 시스템 import 여부(`imported`)를 표시해야 한다.
- **REQ-FR-REPO-003 (MVP, 확정):** 시스템 관리자는 외부 SCM 에 **실제 저장소를 생성**([API-90](../integration-registry/api.md))하고 시스템으로 미러할 수 있어야 한다 (outbound, Phase C).
    - 생성된 row 는 `source=system`(시스템이 생성을 주도) + `provider_id` 세팅 + SCM 응답값으로 mirror 필드를 채운다. 이후 sync 가 mirror 필드를 갱신해도 `source`/`description` 는 보존된다.
- **REQ-FR-REPO-004 (MVP, 확정):** 시스템 주도 repository 는 **draft → active 게시 생애주기**를 가져야 한다 (`repository_status`, migration 000043).
    - `POST /api/v1/repositories` 로 `draft`(source=system) 상태 row 를 생성하고, `POST /api/v1/repositories/{repository_id}/publish` 로 게시를 요청하면 외부 SCM 에 실제 저장소를 생성한 뒤 `active` 로 전이한다. `publish_requested_at`/`published_at` 시점을 기록한다.
    - SCM 에서 import/sync 된 repository(REQ-FR-REPO-002)는 `active` 상태로 직행한다 (draft 단계 없음).
    - publish 요청 가능 상태는 `draft` 뿐이다. (위 endpoint 들은 본 절 기준 master `docs/backend_api_contract.md` §13.9 에서 본문 spec staged.)
- **REQ-FR-REPO-005 (MVP, 확정):** import/create/sync 동작은 provider 의 `capabilities` 를 **기능 gate** 로 사용해야 한다.
    - `pull` = 원격 조회/import(API-88/89) 허용, `sync` = mirror sync(API-72) 허용(`pull` 포함), `push` = outbound 저장소 생성(API-90) 허용. gate 미충족 시 `422 integration_capability_not_enabled` 로 거절한다.

## 3. 비기능 / 운영 요구사항 (REQ-NFR-REPO)

- **REQ-NFR-REPO-001 (MVP):** SCM sync upsert 는 **멱등**해야 하며 ON CONFLICT 시 SCM mirror 필드(clone_url/default_branch/private 등)만 갱신하고 system-owned 필드(`source`/`description`)는 보존해야 한다. in-memory fake(테스트) 도 production store 와 동일하게 보존 미러를 구현한다(parity).
- **REQ-NFR-REPO-002 (MVP):** outbound create/publish 는 현재 **Gitea REST client** 만 구현돼 있어 Gitea-compatible provider(gitea/forgejo/gogs)로 제한된다. 비-호환 vendor(github/gitlab/bitbucket)는 `422 integration_provider_not_gitea_compatible` 로 거절해야 한다. 비활성(`disabled`) provider 대상 연동은 `409` 로 거절한다. (provider 추상화 확장 시 다른 vendor 어댑터 추가로 확장 — REQ-FR-APP-009 정합.)
- **REQ-NFR-REPO-003 (MVP):** 소유권 전이/import/create/publish 는 audit 가능해야 한다. publish 흐름은 외부 SCM 생성 실패 시 부분 실패 경로(생성 실패 + draft 보존)를 가지므로, 본 lifecycle 의 자동화 테스트(UT/E2E) 보강이 후속 과제로 추적되어야 한다 (현재 draft→publish 핸들러는 무테스트로 머지 — [SDLC 체인 점검 G4](../../analysis/2026-05-27-codebase-snapshot/02_sdlc_chain_status.md)).

## 4. 범위 경계 (Out of Scope)

- DevHub → SCM 의 양방향 상태 동기화(이슈/PR write-back) — integration-registry REQ-FR-INT-012 와 동일하게 별도 승인 정책 후.
- Gitea 외 vendor(github/gitlab/bitbucket)에 대한 outbound create/publish 어댑터.
- 신규 Platform 등록 시 Gitea 저장소 자동 생성/브랜치 보호/멤버 초대 자동 오케스트레이션(platform-lifecycle 후속).
- 평문 secret(`credentials_ref`/`api_token`/`auth_secret`) 의 envelope 암호화 — integration-registry REQ-NFR-INT-009 의 #6 carve 로 추적.

## 5. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-29 | Phase 3 split — master `docs/requirements.md` §5.8 에서 본문 그대로 이관. ID(REQ-FR-REPO-001..005, REQ-NFR-REPO-001..003) 보존, 신규 발급/삭제 없음. |
