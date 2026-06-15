# Sprint Plan: project.inbound_source 자동 routing (ADR-0028 §6 carve a) 검토

- 문서 목적: ADR-0028 §6 의 (a) carve out — `project.inbound_source` 자동 routing 의 구현 sprint 진입 전 **정공법 + 스코프 + ID slot + 의존** 정합. v0.1.0 출시 후 v0.1.1 milestone 진입 후보.
- 범위: options 3종 비교 + 결정 후보 + DB schema (migration) + backend API + frontend 운영 + RBAC + ID 발급 + 추적성 영향.
- 대상 독자: Backend / 프론트엔드 개발자, AI agent, QA, 운영자, product owner.
- 상태: **draft** (v0.1.0 출시 직전 검토, 구현은 v0.1.1 milestone 진입 시점)
- 최종 수정일: 2026-06-12
- 결정 근거 sprint: `maintenance/work_260612-c-inbound-source-plan` (option D plan 본 검토)
- 관련 문서:
  - [ADR-0028 §6 carve a](../adr/0028-dev-requests-voc-external-ref.md) (source ADR)
  - [release_v0-1_roadmap.md §3.5 N-12](../planning/release_v0-1_roadmap.md) (voc + notification 정합)
  - [release_v0-1_roadmap.md §3.5 N-11 (CI e2e 복원)](../planning/release_v0-1_roadmap.md) (CI 정합)
  - [docs/domain/dev-request/api.md](../domain/dev-request/api.md) (voc API 표)
  - [docs/architecture.md §6.3 integration-registry](../architecture.md) (integration-registry cross-cut)
  - [docs/planning/external_system_integration_concept.md](./external_system_integration_concept.md) (외부 시스템 어댑터 모델)
  - [docs/adr/0012-dev-request-intake-auth-policy.md](../adr/0012-dev-request-intake-auth-policy.md) (intake token spoofing 방지)

## 1. 배경 (ADR-0028 §6)

ADR-0028 §3 의 본 결정 (옵션 1 채택, voc 별도 도메인 + 1:1 dev-request 매핑) 으로 voc 등록 시 `source_system` 컬럼에 heterogeneous source 통합 (manual / gitea / jira 등) 가능. **post-MVP carve** 항목:

- **(a) `project.inbound_source` 자동 routing**: `applications.inbound_source_type` + `inbound_source_config` 컬럼 + 자동 라우팅 로직. **본 sprint plan 의 대상**.
- (b) email 발송 (P2).
- (c) sms / Slack 통합 (P3, inbound_source_config.webhook).
- (d) voc list API (✅ PR #515 에서 이미 구현, §6 의 d 항목 close).

**현재 voc 등록 흐름**:
1. 외부 시스템 → `POST /api/v0-1/dev-requests/:external_ref` (intake token 인증) → `dev_request_vocs` INSERT (status=received) → `assignee` 에 in-app notification.
2. 담당자/관리자 dashboard → `POST /api/v0-1/dev-requests/:external_ref/route` (system_admin) → `dev_request_vocs` UPDATE (status=routed) + `dev_requests` INSERT (단일 트랜잭션) → `assignee` 에 in-app notification (의뢰 라우팅 완료).

**post-MVP 의 자동 routing 의 의미**:
- 외부 시스템에서 의뢰 도착 시 (step 1) — `project` (= `applications.id`) 가 자동 결정되면 voc 단계 skip → 직접 dev-request 등록.
- `applications.inbound_source_type` + `inbound_source_config` 가 source_system + 외부 시스템 ID 매핑의 source-of-truth.
- 의뢰의 **req_department** 또는 **requester** 또는 **external_ref pattern** 매칭으로 project 자동 결정.

## 2. 후보 옵션 (구현 정공법)

| # | 옵션 | 결정 후보 | 의존성 |
| --- | --- | --- | --- |
| **A** | **`applications.inbound_source` 컬럼 + 자동 routing** (ADR-0028 §6 carve a) | ⭐ 권장 | migration 1 신규 + schema 1 확장 + API 1 신규 + worker 1 신규 |
| **B** | `integration_providers` 와 매핑 (provider_type='alm'/'scm'/'ci_cd' + provider config 재활용) | 권장 보류 | migration 0 신규 + schema 0 확장 (재활용) + API 0 신규 + worker 1 신규 |
| **C** | voc 단계 유지 + 자동 routing + 자동 ingestion 후 명시적 promote (voc 의 status=received 단계 흡수) | 권장 보류 | migration 1 신규 + schema 0 확장 + worker 1 신규 |

### 2.1 옵션 A 상세 (권장)

**DB schema**:
```sql
-- 000007_application_inbound_source.up.sql
ALTER TABLE public.applications
    ADD COLUMN inbound_source_type text NULL,  -- 'gitea' | 'jira' | 'other' | NULL (수동 routing)
    ADD COLUMN inbound_source_config jsonb NOT NULL DEFAULT '{}'::jsonb,
    ADD CONSTRAINT applications_inbound_source_type_check
        CHECK (inbound_source_type IS NULL OR inbound_source_type IN ('gitea', 'jira', 'other'));
```

**backend 구현**:
1. `internal/domain/platform-lifecycle/repository/applications.go` 의 `CreatePlatform` / `UpdatePlatform` 에 inbound_source 2 field 추가.
2. `internal/domain/platform-lifecycle/routing/auto_route.go` 신규 (기존 platform-lifecycle domain path 정합) — 외부 시스템 ID 매칭 + project 자동 결정. **worker** 가 아니고 **synchronous routing** (voc 등록 시점에 즉시 적용) — **post-MVP 검토에서 worker vs sync 결정 필요**.
3. `internal/domain/dev-request/view/voc_handler.go` `createOrGetVoc` 에 자동 routing 호출 — match 시 voc skip + dev-request 직접 등록.

**API**:
- `PATCH /api/v0-1/platforms/:platform_id` body 에 `inbound_source_type` + `inbound_source_config` 추가 (system_admin).
- `POST /api/v0-1/dev-requests/:dev_request_id` (voc 등록) 의 응답에 `auto_routed: true|false` + `dev_request_id: <uuid>` 추가 (자동 routing 적용 시).

**RBAC**: `inbound_source` 설정은 system_admin 일임, voc 자동 routing 결정 결과는 운영자 dashboard 에서 audit 확인.

**매칭 전략** (worker / sync):
- `external_ref pattern` 매칭: `^GITEA-([0-9]+)$` → gitea issue ID lookup.
- `requester` 매칭: Keycloak `email` 또는 `login` 매핑.
- `req_department` 매칭: organization hierarchy 매칭.
- **매칭 없으면 → voc 단계 유지** (현행 동작 보존).

### 2.2 옵션 B 상세 (권장 보류)

- `integration_providers` (provider_type='alm'|'scm'|'ci_cd', provider_id, config) 와 1:N 매핑.
- applications 에 column 추가 없이 provider 의 config 의 `application_ids` 또는 별도 join table 로 매핑.
- **단점**: integration_providers 의 provider_type 이 5종 (alm/scm/ci_cd/doc/infra) 으로 분류되어 dev-request 의 외부 source (Gitea issue / Jira ticket) 가 어느 provider_type 에 속하는지 모호. scm 으로 분류하면 Gitea repository 와의 매핑과 충돌.
- **결론**: option B 는 integration-registry cross-cut 검토 필요. sprint 후속.

### 2.3 옵션 C 상세 (권장 보류)

- voc 단계 유지 + 자동 routing + 자동 ingestion 후 명시적 promote.
- voc 의 status 머신에 `auto_routed` 단계 추가.
- **단점**: voc 의 3-state 머신 (received → routed → closed) 이 4-state 로 확장. PR #514 의 ADR-0028 §3 결정과 충돌.
- **결론**: option A 의 자동 routing (skip) 과 비교해 audit trail 손실 (voc 의 received 단계의 기록이 남지 않음).

## 3. 결정 (sprint 진입 시)

**옵션 A (applications.inbound_source 컬럼 + sync 자동 routing)** 권장.

### 3.1 ID slot (구현 sprint 진입 시 발급)

| ID 종류 | ID | 스코프 |
| --- | --- | --- |
| REQ | `REQ-FR-113` | inbound_source 자동 routing 정책 + 매칭 전략 |
| UC | `UC-DEV-REQ-15` | 외부 시스템 의뢰 → applications 매칭 → dev-request 직접 등록 flow |
| ARCH | `ARCH-23` | applications.inbound_source schema + sync routing 결정 |
| API | `API-103` | PATCH /api/v0-1/platforms/:platform_id inbound_source + POST /api/v0-1/dev-requests/:external_ref auto_routed 응답 |
| RM | `RM-DEV-REQ-15` | sprint -d (구현 sprint) |
| IMPL | `IMPL-inbound-source-01` | applications inbound_source + auto route worker + voc integration |
| IMPL | `IMPL-platform-patch-02` | PATCH /platforms/:id inbound_source 확장 (ADR-0028 §6 정합) |
| UT | `UT-inbound-source-01` | pattern matcher unit test + integration test |
| TC | `TC-INBOUND-SRC-01` | E2E: PATCH inbound_source → POST voc → 자동 routing 결과 검증 |

### 3.2 마일스톤 (release_v0-1_roadmap 정합)

- **본 sprint plan 의 역할**: v0.1.0 출시 후 **v0.1.1 milestone 진입 시점** 의 sprint 진입 후보 정합.
- `release_v0-1_roadmap.md` §4.2 v0.1.1 milestone 에 본 plan 의 옵션 A 의 ID slot 추가.
- `release_v0-1_roadmap.md` §3.5 N-13 row 신규 (post-MVP carve a 의 정식 ID).

### 3.3 의존

| 의존 | 상태 | 비고 |
| --- | --- | --- |
| ADR-0028 (voc + notification) | ✅ resolved (PR #514 + #515) | 본 plan 의 source |
| `applications` table | ✅ stable | `migrations/000001_initial_schema.up.sql:48` |
| `integration_providers` table | ✅ stable | pattern 매칭 strategy 의 보완 후보 |
| `intake token` 인증 (ADR-0012) | ✅ resolved | 외부 시스템 인증 |
| Keycloak user lookup | ✅ stable | requester → user_id 매핑 |
| organization hierarchy | ✅ stable | req_department → unit_id 매핑 |
| **PR #548 (`feat/work_260611-a-n13-inbound-source-impl`)** | ❌ **CLOSED (2026-06-11 05:40 UTC)** | E2E Internal 1 fail — Test 1 (e2e seed 중복 strict mode violation) + Test 2 (Sign-out timeout). 자동 재실행 미적용 (PR #550 spec timing fix 미반영). follow-up: rebase main + PR #550 fix + e2e seed 정합 fix + 자동 재실행 = v0.1.1 milestone 진입 시점 별도 sprint (sprint `fix/work_260612-1-n13-housekeeping-followup` 결정). |
| **PR #548 (`feat/work_260611-a-n13-inbound-source-impl`)** | ❌ **CLOSED (2026-06-11 05:40 UTC)** | E2E Internal 1 fail — Test 1 (e2e seed 중복 strict mode violation) + Test 2 (Sign-out timeout). 자동 재실행 미적용 (PR #550 spec timing fix 미반영). follow-up: rebase main + PR #550 fix + e2e seed 정합 fix + 자동 재실행 = v0.1.1 milestone 진입 시점 별도 sprint (sprint `fix/work_260612-1-n13-housekeeping-followup` 결정). |

### 3.4 정합 항목

- **traceability §2.1 REQ-FR-113 + §2.1.5 UC-DEV-REQ-15 + §2.2 ARCH-23 + §2.2 API-103 + §2.3 RM-DEV-REQ-15 + §2.4 IMPL-inbound-source-01 + §2.4 IMPL-platform-patch-02 + §2.5 UT-inbound-source-01 + §2.6 TC-INBOUND-SRC-01**: 9 row 추가.
- **release_v0-1_roadmap §3.5 N-13 row 신규** + §4.2 v0.1.1 milestone 정합.
- **ADR-0028 §6 amendment**: 본 plan 결정 사항을 ADR-0028 §6 의 (a) carve out 에 정합 노트 (sprint 진입 시점).

## 4. 정공법 (구현 sprint 진입 시)

| 순번 | 산출물 | 의존 | 예상 |
| --- | --- | --- | --- |
| 1 | **migration 000007** (applications.inbound_source 2 field + CHECK) | (없음) | 30분 |
| 2 | **domain.Platform** 2 field 추가 | (없음) | 30분 |
| 3 | **platform-lifecycle repository** 2 method (`UpdatePlatformInboundSource` + view API) | 1, 2 | 1시간 |
| 4 | **routing/auto_route.go** 신규 — pattern matcher + 3 매칭 전략 | 2 | 2시간 |
| 5 | **voc_handler.createOrGetVoc** 자동 routing 호출 + 응답 | 3, 4 | 1시간 |
| 6 | **routePermissionTable** 1 entry 추가 (PATCH /platforms inbound_source) | 3 | 15분 |
| 7 | **UT + IT** (pattern matcher 3 case + auto route 1 case + store 1 case) | 4, 5 | 2시간 |
| 8 | **E2E TC-INBOUND-SRC-01** (PATCH → POST → auto_routed 검증) | 5, 6 | 1시간 |
| 9 | **traceability §2.1~§2.6 8 row** + release_v0-1_roadmap §3.5 N-13 + §4.2 | (없음) | 30분 |
| 10 | **ADR-0028 §6 amendment** (본 plan 결정 사항) | 1-9 | 15분 |
| **합계** | | | **~9시간** (1 sprint) |

## 5. 결정 보류 사유 (현 sprint = post-MVP 검토)

본 plan 은 **option D 검토 = 정공법 + ID slot + 의존 정합** 만 수행. **구현 = v0.1.1 milestone 진입 시점** 의 별도 sprint 에서 결정.

**현 sprint (v0.1.0 출시 직전) 의 잔여 = N-6 staging 1주 운영** (사용자 결정 영역). 본 plan 의 구현은 v0.1.0 staging 운영 후 v0.1.1 진입 시점.

**2026-06-12 보강** (sprint `fix/work_260612-1-n13-housekeeping-followup`): 본 plan 의 1차 구현 시도 (sprint `feat/work_260611-a-n13-inbound-source-impl`, PR #548) 가 E2E Internal 1 fail 2건으로 CLOSED 결정. follow-up 결정 3 branch: (1) Test 1 e2e seed 중복 → spec/e2e seed 정합 fix 별도 sprint; (2) Test 2 Sign-out timeout → main rebase + 자동 재실행 검증 별도 (PR #550 spec timing fix 가 해결 가능성 ↑); (3) 구현 follow-up = v0.1.1 milestone 진입 시점 별도 sprint (rebase main + PR #550 fix + e2e seed 정합 fix + 자동 재실행 종합). 본 plan 의 구현 = 1차 시도 보류 + follow-up 종합 검증 후 v0.1.1 진입 시점 재진입.

## 6. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-06-12 | 1차 작성 (sprint `maintenance/work_260612-c-inbound-source-plan`) — ADR-0028 §6 carve (a) 의 정공법 + ID slot + 의존 정합. 옵션 A 권장 (applications.inbound_source 컬럼 + sync 자동 routing). 구현은 v0.1.1 milestone 진입 시점. |
| 2026-06-12 | **N-13 PR #548 close follow-up 결정** (sprint `fix/work_260612-1-n13-housekeeping-followup`, docs only) — 본 plan 의 1차 구현 시도 (PR #548) 가 E2E Internal 1 fail 2건 (Test 1 e2e seed 중복 + Test 2 Sign-out timeout) 으로 CLOSED. §3.3 의존 표에 PR #548 CLOSED row 추가 + §5 결정 보류 사유 보강 (3 branch follow-up 결정) + 본 §6 row 추가. ADR-0028 §6 (a) + release_v0-1_roadmap.md §3.5 N-13 row + traceability/report.md §6 + 메모리 4 file 동기화. 신규 ID 발급 0건 (housekeeping follow-up). |
