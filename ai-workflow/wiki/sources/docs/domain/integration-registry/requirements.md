---
title: requirements
type: source
tags: [domain, requirements.md, project-devhub]
sources: [raw/projects/devhub/docs/domain/integration-registry/requirements.md]
git_commit: baf1cf24
git_branch: main
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
last_touched: 2026-06-22T03:37:45Z
mirror_dirty: |
related: [none]
status: draft
contradictions: [none]
---

# integration-registry 도메인 요구사항

- 문서 목적: 외부 시스템 연동(Integration) Provider 등록/카탈로그/수집/HomeLab 인벤토리 + auth_mode/base_url/webhook 헤더 alias 의 기능·비기능 요구사항을 정의한다.
- 범위: REQ-FR-INT-001..015 / REQ-NFR-INT-001..009. Task Item Ingestion 도메인은 `task_requirements.md` 참조 (REQ-FR-TASK-001..010 / REQ-NFR-TASK-001..004).
- 대상 독자: backend / 프론트엔드 / DevOps, AI agent, QA.
- 상태: accepted
- 최종 수정일: 2026-05-29 (Phase 3 split, master `docs/requirements.md` §5.6 본문 이관)
- 관련 문서: [도메인 README](./README.md), [컨셉](./external_system_concept.md), [architecture.md](./architecture.md), [api.md](./api.md), [master requirements](../../requirements.md), [ADR-0015](../../adr/0015-homelab-pull-strategy.md)

## 1. 개요

본 절은 컨셉 문서([`./external_system_concept.md`](./external_system_concept.md), 2026-05-15)에 정의된 ALM/SCM/CI-CD/문서/홈랩 인프라 통합 운영 모델의 요구사항을 정의한다.

## 2. 기능 요구사항 (REQ-FR-INT)

- **REQ-FR-INT-001 (MVP):** 시스템 관리자는 외부 연동 대상을 Provider 단위로 등록/수정/비활성화할 수 있어야 한다.
    - 최소 속성: `provider_key`, `provider_type` (alm|scm|ci_cd|doc|infra), `display_name`, `enabled`, `auth_mode`, `scope`.
- **REQ-FR-INT-002 (MVP):** DevHub 는 Provider Catalog 를 제공해야 하며, 각 provider 의 `capabilities` 를 조회할 수 있어야 한다.
    - 예: `issues.read`, `repo.read`, `pr.read`, `build.read`, `doc.meta.read`, `infra.node.read`.
- **REQ-FR-INT-003 (MVP):** 연동 데이터 수집은 `webhook ingest` 와 `scheduled pull` 을 모두 지원해야 한다.
    - Provider 별로 ingest/pull 사용 여부를 설정할 수 있어야 한다.
- **REQ-FR-INT-004 (MVP):** SCM provider (`bitbucket`, `gitea`, `forgejo` 등)는 공통 Repository/PR/Activity 모델로 정규화되어야 한다.
    - Provider 특화 필드는 확장 payload 로 보존한다.
- **REQ-FR-INT-005 (MVP):** CI/CD provider (`bamboo`, `jenkins`) 데이터는 공통 BuildRun 모델로 정규화되어야 한다.
    - 최소 필드: `external_run_id`, `status`, `branch`, `commit_sha`, `started_at`, `finished_at`, `duration_seconds`.
- **REQ-FR-INT-006 (MVP):** ALM/문서 연동은 최소 링크형 통합을 지원해야 한다.
    - Jira: project/issue key 매핑, Confluence: space/page 링크 및 메타데이터 조회.
- **REQ-FR-INT-007 (MVP):** Integration 은 Application 또는 Project scope 에 연결될 수 있어야 한다.
    - scope 별 연결 정책(`application|project`)을 명시적으로 저장해야 한다.
- **REQ-FR-INT-008 (MVP):** 홈랩 인프라 관리를 위해 Node/Service 인벤토리를 등록/조회할 수 있어야 한다.
    - Node 최소 필드: `node_id`, `hostname`, `ip`, `environment`, `owner`.
    - Service 최소 필드: `service_id`, `node_id`, `name`, `version`, `port`, `health_status`.
- **REQ-FR-INT-009 (MVP):** 홈랩 상태 수집 결과를 토폴로지 형태(`nodes`, `edges`, `services`)로 조회할 수 있어야 한다.
- **REQ-FR-INT-010 (MVP):** Provider 및 홈랩 수집 상태는 `sync_status` 로 관리되어야 한다.
    - 최소 상태: `requested`, `verifying`, `active`, `degraded`, `disconnected`.
- **REQ-FR-INT-011 (MVP):** 일반 사용자는 권한 범위 내에서 통합 운영 현황을 조회할 수 있어야 한다.
    - `system_admin`은 전체, 일반 역할은 프로젝트/소유 범위 기반 조회.
- **REQ-FR-INT-012 (후속):** DevHub 에서 외부 시스템으로의 양방향 상태 변경(write-back)은 별도 승인 정책(ADR) 후 도입한다.
- **REQ-FR-INT-013 (MVP, 추가 2026-05-27 — 등록 UX 고도화):** Provider 등록은 `auth_mode` 별 outbound 자격증명 full 모델을 지원해야 한다 (REQ-FR-INT-001 의 `auth_mode` 속성 구체화, [API-70](./api.md), migration 000040/000041).
    - 지원 `auth_mode`: `token`, `basic`, `oauth2`, `app_password`, `agent` (등록 시 고정, PATCH 로 변경 불가).
    - mode 별 자격증명: `token`→`api_token`(PAT); `basic`/`app_password`→`auth_username`+`auth_secret`; `oauth2`→`auth_client_id`+`auth_token_url`+`auth_secret`(client_secret); `agent`→`auth_username`(서버 직접 sync 미사용).
    - 자격증명 시크릿(`api_token`/`auth_secret`)은 **write-only** — 응답에 raw 미노출, `api_token_set`/`auth_secret_set` (bool) 만 노출. 비밀이 아닌 필드(`auth_username`/`auth_client_id`/`auth_token_url`)는 응답 노출.
    - inbound webhook 서명 시크릿(`credentials_ref`)과 outbound auth 자격증명은 별개 시크릿으로 분리 관리한다.
    - 미리 정의된 vendor preset(gitea/forgejo/gogs/github/gitlab/bitbucket/jenkins 등) 을 통해 `provider_type`/`capabilities`/권장 `auth_mode` 를 가이드 입력할 수 있어야 한다.
- **REQ-FR-INT-014 (MVP, 추가 2026-05-27):** Provider 는 outbound sync/pull 대상 endpoint 인 `base_url`(http(s) URL) 을 보유할 수 있어야 하며(migration 000038), 등록 전/후 endpoint **연결 테스트**(reachability) 를 제공해야 한다([API-87](./api.md) `POST /integration/test-connection`).
    - 연결 테스트는 저장 전(pre-save) body 의 `base_url` 로 직접 GET (5s timeout, redirect 미추적, 응답 본문 미반환). 결과는 reachable/status_code/latency 만 반환.
    - 사내 internal endpoint(Gitea/Jenkins 등)가 합법 대상이므로 internal IP 차단을 하지 않는다 — admin(`infrastructure:edit`) 신뢰 경계 + timeout + 응답 본문 미반환으로 SSRF 표면을 최소화한다 (의도적 수용, 운영 신뢰 경계 명시).
- **REQ-FR-INT-015 (MVP, 추가 2026-05-27):** 범용 webhook ingest([API-73](./api.md)) 는 vendor 별 서명/이벤트 헤더 이름 차이를 흡수하기 위해 헤더 alias fallback 을 지원해야 한다.
    - 우선순위: `X-Integration-*` → `X-Gitea-*` → `X-Gogs-*` (Gitea/Gogs 헤더 불일치 정정). 정규화 후 dedupe 및 sync state 갱신은 기존 정책(REQ-NFR-INT-003)을 따른다.

## 3. 비기능 / 운영 요구사항 (REQ-NFR-INT)

- **REQ-NFR-INT-001 (MVP):** 외부 연동 인증정보는 평문 저장을 금지하고, 암호화 저장 또는 외부 Secret Store 참조를 사용해야 한다.
- **REQ-NFR-INT-002 (MVP):** 모든 연동 생성/변경/비활성화/실패/복구 이벤트는 audit_logs 에 기록되어야 한다.
- **REQ-NFR-INT-003 (MVP):** 수집 파이프라인은 idempotency key 기반 중복 방지를 지원해야 한다.
- **REQ-NFR-INT-004 (MVP):** 특정 provider 장애가 전체 연동 파이프라인 중단으로 전파되지 않도록 provider 단위 격리를 보장해야 한다.
- **REQ-NFR-INT-005 (MVP):** 연동 조회 API 는 페이지네이션과 필터(Provider, Scope, Status, Time Range)를 지원해야 한다.
- **REQ-NFR-INT-006 (MVP):** 홈랩 상태 데이터는 최신 스냅샷과 변경 이력을 구분해 제공해야 한다.
- **REQ-NFR-INT-007 (후속):** Provider 별 Rate Limit 초과 대응(백오프/재시도/서킷 브레이커)은 운영 정책으로 표준화한다.
- **REQ-NFR-INT-008 (후속):** 대규모 연동 환경에서의 성능 목표(p95 응답시간, 수집 지연 SLA)는 운영 계측 후 확정한다.
- **REQ-NFR-INT-009 (MVP, 추가 2026-05-27 — secret 노출 경계):** outbound auth 자격증명(`api_token`/`auth_secret`)은 write-only 로 raw 응답 노출을 금지하고 `*_set` (bool) 만 노출해야 한다 (REQ-FR-INT-013 정합).
    - 입력 정규화: create 는 `TrimSpace`, update 는 blank/미전송 시 기존 값 유지(nil-skip), DB 저장은 `NULLIF($n,'')`.
    - **알려진 미해소 gap**: inbound webhook 시크릿 `credentials_ref` 는 현재 raw 그대로 응답에 노출되며, 저장도 평문 컬럼(`credentials_ref`/`api_token`/`auth_secret`)이다 — REQ-NFR-INT-001(평문 저장 금지)의 미충족 잔여. envelope 암호화/외부 Secret Store 참조 전환은 별도 보안 carve(#6 평문 secret)로 추적한다.

## 4. 범위 경계 (Out of Scope)

- 외부 시스템 본문 데이터의 실시간 양방향 동기화(write-back 강제 적용).
- 복잡한 승인 워크플로우(다단계 승인, 릴리즈 체인 오케스트레이션).
- 멀티 테넌트 완전 분리 모델.
- AI 기반 자동 최적화/자동 복구.

## 5. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-29 | Phase 3 split — master `docs/requirements.md` §5.6 에서 본문 그대로 이관. ID(REQ-FR-INT-001..015, REQ-NFR-INT-001..009) 보존, 신규 발급/삭제 없음. Task Ingestion 은 `task_requirements.md` 로 분리. |
