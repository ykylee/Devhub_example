# ADR-0028: dev-requests voc 도메인 + external_ref 매칭 + in-app notification

- **문서 목적**: 외부 시스템 의 의뢰 (voc = Voice of Customer) 가 project 결정 전 staging 단계로 도착하는 voc 도메인 신설 + 9 field 매칭 + in-app notification 정책 결정.
- **범위**: voc 도메인 (dev_request_vocs table, status 머신, idempotency) + dev_requests table 확장 (4 field) + user_notifications table (in-app) + 5 신규 API + project.inbound_source 자동 routing (post-MVP carve).
- **대상 독자**: Backend / 프론트엔드 개발자, AI agent, QA, 운영자.
- **상태**: accepted
- **최종 수정일**: 2026-06-12 (N-13 PR #548 close follow-up 결정 — §6 (a) 본문 보강 + §7 변경 이력 1 row. 본 sprint = docs only housekeeping follow-up, 구현 follow-up = v1.1 milestone 진입 시점 별도 sprint. 결정 근거 sprint: `fix/work_260612-1-n13-housekeeping-followup`.)
- **결정 근거 sprint**: `feat/work_260612-a-dev-requests-voc-domain` (v1.0 출시 직전)
- **관련 문서**: [docs/backend_api_contract.md §14](../backend_api_contract.md) (DREQ), [docs/api/conventions.md §1](../api/conventions.md) (envelope), [docs/adr/0012-dev-request-intake-auth-policy.md](./0012-dev-request-intake-auth-policy.md) (ADR-0012 spoofing 방지), [docs/validation/N-10-manager-rbac.md](../validation/N-10-manager-rbac.md) (Manager RBAC, 본 ADR 의 assignee 권한 정합 근거), [release_v1_roadmap.md §3.5 N-12](../planning/release_v1_roadmap.md) (N-12 백로그 신규), [docs/traceability/report.md §2.4 IMPL-voc-01](../traceability/report.md) (cross-cutting IMPL).

## 1. 배경

v1.0 출시 직전 sprint. DevHub 의 dev-requests API 가 외부 시스템 (Gitea issue, Jira ticket 등) 에서 의뢰를 수신하여 **project 매칭 + dev-request 등록 + 담당자 라우팅** 까지 단일 흐름으로 처리하나, **project 결정 전 단계** (= 어느 project 에 속할지 미정) 에서 의뢰가 도착하는 경우 (예: 직접 등록 / 분류 미정 외부 의뢰) 가 있어:

1. **담당자가 미정** 상태로 의뢰가 도착하여 **알림이 발송되지 않거나 잘못된 담당자에게 발송** 됨.
2. **project_id 가 미정** 인 dev-request 가 status 머신에 등록되어 **status 머신의 received 단계의 의미** (= "이게 진짜로 처리할 의뢰인지") 가 모호.
3. **external_ref** 가 **optional** (nullable) 인 기존 schema 는 **동일 외부 시스템의 동일 ticket 이 중복 ingest** 되는 경로 차단 불가 (ADR-0012 §4.1.2 spoofing 방지 + idempotency 의 source of truth 부재).

본 sprint 에서 **voc = Voice of Customer** 도메인을 **project 결정 전 staging 의뢰의 collection** 으로 신설하고, **9 field 매칭 (사용자 명세)** + **in-app notification** + **voc → dev-request 자동 라우팅** 으로 정합.

## 2. 후보 옵션

| # | 옵션 | 의존성 | 결정 |
| --- | --- | --- | --- |
| 1 | **voc 별도 도메인 + 1:1 dev-request 매핑** (본 결정) | schema 2 신규 + 4 확장 + API 5 + ADR 1 | ⭐ 채택 |
| 2 | voc 없음 — dev-request status=received 단계 활용 | schema 0 신규 (4 field 확장만) + API 0 | ❌ (voc 가 별도 의미 + status 머신이 미완성) |
| 3 | voc 만 도메인 (dev-request 와 독립) | schema 1 신규 + API 1 | ❌ (dev-request 1차 등록 단계 매핑 불가) |

상세 비교:

- **옵션 1 (채택)**: `dev_request_vocs` table + `user_notifications` table + dev_requests 4 field 추가. voc 의 9 field (title/details/requester/req_department/assignee/dev_department/request_date/dev_schedule/external_ref) 가 사용자 명세와 1:1 매핑. 외부 API 진입 시 `dev_request_vocs` INSERT + status=received. project 결정 시 status=routed + `dev_requests` row 자동 생성 (status=pending). 담당자 알림 (in-app) 발송.
- **옵션 2**: dev-request 의 status 머신에 `received` 가 이미 존재. project_id nullable + status 머신만 활용. **단점**: voc 의미 (= "어느 project 인지 미정" 의 collection) 가 status 머신의 1 단계와 충돌. dev-request 의 다른 status (pending, registered) 와의 의미가 섞임.
- **옵션 3**: voc 만 별도. project routing 시 dev-request 가 아닌 voc 가 매핑. **단점**: dev-request 의 promote-tx 단일 트랜잭션 패턴 (ADR-0013 §5, sprint `claude/work_260515-m`) 과 호환 불가.

## 3. 결정

**옵션 1 (voc 별도 도메인 + 1:1 dev-request 매핑) 채택**.

### 3.1 voc 도메인

```sql
CREATE TABLE public.dev_request_vocs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    external_ref text NOT NULL,
    source_system text NOT NULL,  -- ADR-0012 §4.1.2 spoofing 방지 (인증된 intake token 매핑)
    title text NOT NULL,
    details text NOT NULL DEFAULT '',
    requester text NOT NULL DEFAULT '',
    req_department text NOT NULL DEFAULT '',  -- 신규 (사용자 9 field)
    assignee_user_id text NULL REFERENCES public.users(user_id) ON DELETE SET NULL,
    dev_department text NOT NULL DEFAULT '',  -- 신규
    request_date date NULL,                       -- 신규
    dev_schedule text NOT NULL DEFAULT '',       -- 신규
    status text NOT NULL DEFAULT 'received'
        CHECK (status IN ('received', 'routed', 'closed')),
    project_id uuid NULL REFERENCES public.applications(id) ON DELETE SET NULL,
    dev_request_id uuid NULL REFERENCES public.dev_requests(id) ON DELETE SET NULL,
    routed_at timestamptz NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT dev_request_vocs_external_ref_source_uniq
        UNIQUE (source_system, external_ref)
);
```

### 3.2 voc status 머신 (3-state)

| from / to | received | routed | closed |
| --- | --- | --- | --- |
| **received** | - | ✓ | ✓ |
| **routed** | - | - | ✓ |
| **closed** | - | - | - |

- `received` → `routed`: project_id 결정 + dev-request 자동 생성 (단일 트랜잭션).
- `received` → `closed`: 명시적 close (의뢰 거절 / 미분류 폐기).
- `routed` → `closed`: dev-request 등록 후 명시적 close (운영자 결정).

### 3.3 dev_requests 테이블 확장

사용자 9 field 중 4 field 가 기존 DREQ schema 에 없음 (req_department / dev_department / request_date / dev_schedule). 4 column 추가:

```sql
ALTER TABLE public.dev_requests
    ADD COLUMN IF NOT EXISTS req_department text NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS dev_department text NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS request_date date NULL,
    ADD COLUMN IF NOT EXISTS dev_schedule text NOT NULL DEFAULT '';
```

이 4 field 는 voc 의 routing 시 dev-request 자동 생성 시점에 복사됨 (voc.id → dev_request_id 1:1 link).

### 3.4 user_notifications 테이블 (in-app notification)

```sql
CREATE TABLE public.user_notifications (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id text NOT NULL REFERENCES public.users(user_id) ON DELETE CASCADE,
    kind text NOT NULL,  -- 'dev_voc' | 'dev_request'
    ref_id text NULL,    -- dev_request_voc_id or dev_request_id
    title text NOT NULL,
    body text NOT NULL DEFAULT '',
    read_at timestamptz NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX user_notifications_user_unread_idx
    ON public.user_notifications(user_id, created_at DESC) WHERE read_at IS NULL;
```

### 3.5 5 신규 API

| Method | Path | 의도 | RBAC |
| --- | --- | --- | --- |
| POST | `/api/v1/dev-requests/:external_ref` | voc 등록 (idempotent) | (frontend 직접 호출) |
| POST | `/api/v1/dev-requests/:external_ref/route` | voc routing + dev-request 자동 생성 | system_admin |
| GET | `/api/v1/dev-requests/:external_ref` | voc 단건 조회 | (frontend / 직접 호출) |
| GET | `/api/v1/me/notifications?limit=50` | 본인 unread 목록 | 인증 user |
| POST | `/api/v1/me/notifications/:id/read` | 본인 notification 읽음 처리 | 인증 user (본인 ownership 검증) |

**POST /api/v1/dev-requests/:external_ref** body 8 field (external_ref 는 path param):
```json
{
  "title": "...",
  "details": "...",
  "requester": "...",
  "req_department": "...",
  "assignee": "user_id",
  "dev_department": "...",
  "request_date": "2026-06-12",
  "dev_schedule": "2 weeks",
  "source_system": "manual"  // or "gitea", "jira", etc
}
```

**idempotency**: (source_system, external_ref) UNIQUE. 사전 SELECT + 200 반환.

## 4. Trade-off

- **장점**:
  - voc = staging 으로 project 결정 전 의뢰를 명확히 분리 → status 머신의 모호성 해소.
  - 9 field 사용자 명세와 1:1 매칭 (기존 DREQ 의 7 field 와 정합).
  - in-app notification 으로 외부 의존 없는 알림.
  - 외부 시스템별 source_system 의 보존으로 Gitea/Jira/Manual 등 heterogeneous source 통합.
- **단점**:
  - 2 신규 테이블 (dev_request_vocs + user_notifications) — DB schema 복잡도 ↑.
  - 5 신규 API — surface area ↑.
  - voc + dev-request 의 1:1 link 는 routing 시 1회 발생, dev-request 의 status 머신 외부에서 별도 머신 운영.

## 5. 정합

- **ADR-0012** (intake token spoofing 방지): `source_system` 컬럼은 인증된 intake token 의 매핑값. 본 ADR 의 `source_system` (예: "manual" / "gitea" / "jira") 와 동일 의미.
- **ADR-0013** (DREQ row-scoping): dev-request 의 RBAC row-scoping (manager/team_manager) 은 본 ADR 의 1:1 link 후 적용.
- **backend_api_contract.md §14** (DREQ): 본 ADR 로 DREQ 가 4 field 확장 + voc 별도 도메인 정합. §X (voc 신규) + §Y (notification 신규) section 추가.
- **traceability/report.md §2.4**: `IMPL-voc-01` + `IMPL-notification-01` (cross-cutting). ADR-0028 row + §6 변경 이력.

## 6. Carve Out (post-MVP)

- **project.inbound_source 자동 routing** (P3, **N-13**, post-MVP): `applications.inbound_source_type` (gitea/jira/other) + `inbound_source_config` (JSONB) 컬럼 + 자동 라우팅 로직. sprint 후속. **2026-06-12 정합** ([`docs/planning/2026-06-12-inbound-source-routing-sprint-plan.md`](../planning/2026-06-12-inbound-source-routing-sprint-plan.md) + [release_v1_roadmap.md §3.5 N-13](../planning/release_v1_roadmap.md) 의 정공법 + ID slot + 의존 정합) — 옵션 A (applications.inbound_source 컬럼 + sync 자동 routing) 권장. **2026-06-12 구현 정합** (sprint `feat/work_260612-7-v1-1-inbound-source-impl`, 본 sprint = N-13 follow-up 3 branch 종합 v1.1 sprint -a) — routing/auto_route.go (3 case pattern matcher: external_ref `^GITEA-([0-9]+)$` + req_department + graceful degradation) + voc_handler.createOrGetVoc 통합 (AutoRouter.Route() + RouteVoc() + auto_routed envelope 응답) + IT 3 case (GiteaOK / NoMatch / RouteErrorDegradation) + openapi.yaml 정합 (PATCH /platforms inbound_source_type + inbound_source_config + POST /dev-requests/{id} auto_routed 응답 + DevRequestVoc schema) + e2e TC-INBOUND-SRC-01 (seed platform 사용, POST platform 단계 제거 정공법) + routePermissionTable (PATCH /platforms/:platform_id 가 ResourcePlatforms+ActionEdit 으로 cover, 변경 불요). **N-13 follow-up 3 branch 결정 정공법 (PR #573) 종합**: A (e2e seed fix, PR #574) + B (signout timeout fix, PR #575) + C (구현 follow-up, 본 sprint). backend foundation (PR A-1: migration 000007 + domain.Platform 2 field + UpdatePlatformInboundSource + view.UpdatePlatform inboundTouched + 4 UT) 은 main 에 byte-identical 포함 (PR #549 T-d-72-2 wiki mirror). 본 sprint = PR A-2 만. **status `⏳ planned` → `✅ resolved (implemented, 2026-06-12)`** — 9 ID row 정합. **2026-06-11 ID slot 9 row 발급** (sprint `feat/work_260611-a-n13-inbound-source-housekeeping`): `REQ-FR-113` (정책 + 매칭 전략) / `UC-DEV-REQ-15` (외부 의뢰 → applications 매칭 → dev-request 직접 등록 flow) / `ARCH-23` (applications.inbound_source schema + sync routing 결정) / `API-103` (PATCH `/api/v1/platforms/:platform_id` inbound_source + POST `/api/v1/dev-requests/:external_ref` auto_routed 응답) / `RM-DEV-REQ-15` (sprint -d 구현 sprint, **implemented**) / `IMPL-inbound-source-01` (repository UpdatePlatformInboundSource + routing/auto_route.go + voc_handler 통합, **implemented**) / `IMPL-platform-patch-02` (UpdatePlatform 입력 validate + inbound_source_type CHECK 매핑, **implemented**) / `UT-inbound-source-01` (pattern matcher 3 case + auto route 1 case + store 1 case, **implemented**) / `TC-INBOUND-SRC-01` (PATCH inbound_source → POST voc → auto_routed 검증, **implemented**). `conventions.md §1` 의 RM 표기 정책 (도메인 prefix 관행) 정합 — `RM-DEV-REQ-15` = `RM-{domain}-{nn}` 형식의 dev-request 도메인 prefix 표기 (RM-ONBOARD-XX / RM-APPDASH-XX 와 동일 관행). 본 ID slot 9 row 는 `docs/traceability/report.md` §2 + §3 dev-request row + §4 ADR 인덱스 ADR-0028 row + §6 변경 이력에 정합. **2026-06-12 PR #548 close 정공법** (sprint `fix/work_260612-1-n13-housekeeping-followup`): PR #548 (`feat/work_260611-a-n13-inbound-source-impl`, 2026-06-11 05:40 UTC CLOSED) 의 E2E Internal 1 fail 2건 (Test 1: e2e seed 중복 strict mode violation `getByText('e2e-repo-a')` 2 elements + Test 2: Sign-out timeout N-8 race 유사) + 자동 재실행 미적용 (PR #550 spec timing fix 미반영) 정공법. follow-up 결정: (1) **Test 1 e2e seed 중복** → spec/e2e seed 정합 fix 별도 sprint (mock data 유일성 보장); (2) **Test 2 Sign-out timeout** → main rebase + 자동 재실행 검증 별도 (PR #550 spec timing fix 가 해결 가능성 ↑); (3) **구현 follow-up = v1.1 milestone 진입 시점 별도 sprint** (rebase main + PR #550 fix 적용 + e2e seed 중복 spec fix + 자동 재실행). branch `feat/work_260611-a-n13-inbound-source-impl` close (PR #548) — v1.1 진입 시점의 신규 구현 sprint 는 새 branch 이름 (예: `feat/work_YYMMDD-v1-1-inbound-source-impl`) 별도 결정. status `⏳ planned` 유지 (구현 미완료).
- **email 발송** (P2 정합): in-app 외 외부 의존 (smtp / webhook) 추가.
- **sms / Slack 통합** (P3): project.inbound_source_config 의 webhook field.
- **voc list API** (system_admin): `GET /api/v1/vocs?status=received` — ✅ resolved (PR #515, 2026-06-12). N-6 staging 1주 운영 SOP 의 system_admin 도구.

## 7. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-06-12 | 1차 작성 (sprint `feat/work_260612-a-dev-requests-voc-domain`) — 외부 시스템 의뢰의 project 결정 전 staging 의 voc 도메인 + 9 field 매칭 + in-app notification 정책 결정. 1 table 신규 (dev_request_vocs) + 1 table 신규 (user_notifications) + dev_requests 4 column 확장 + 5 API. |
| 2026-06-11 | **N-13 ID slot 9 row 발급 + §6 (a) 정공법** (sprint `feat/work_260611-a-n13-inbound-source-housekeeping`, docs only) — `conventions.md §1` RM 표기 정책 확장 (도메인 prefix 관행 `RM-{domain}-{nn}` 명문화) + `docs/traceability/report.md` §2.1/§2.1.5/§2.2/§2.3/§2.4/§2.5/§2.6 + §3 dev-request row + §4 ADR 인덱스 ADR-0028 row + §6 변경 이력 + 헤더 메타 정합. **본 sprint = housekeeping only** — 구현은 v1.1 milestone 진입 시점 별도 sprint `feat/work_260611-a-n13-inbound-source-impl` (migration 000007 + domain.Platform 2 field + repository.UpdatePlatformInboundSource + routing/auto_route.go + voc_handler 통합 + PATCH /platforms + UT/IT + E2E + ADR-0028 §6 amendment). |
| 2026-06-12 | **N-13 PR #548 close + follow-up 결정** (sprint `fix/work_260612-1-n13-housekeeping-followup`, docs only) — PR #548 (E2E Internal 1 fail 2건 — Test 1 e2e seed 중복 + Test 2 Sign-out timeout) close 정공법. follow-up 결정 3 branch: (1) Test 1 e2e seed 중복 → spec/e2e seed 정합 fix 별도 sprint (mock data 유일성 보장); (2) Test 2 Sign-out timeout → main rebase + 자동 재실행 검증 별도 (PR #550 spec timing fix 가 해결 가능성 ↑); (3) 구현 follow-up = v1.1 milestone 진입 시점 별도 sprint (rebase main + PR #550 fix 적용 + e2e seed 중복 spec fix + 자동 재실행). branch `feat/work_260611-a-n13-inbound-source-impl` close (PR #548) — v1.1 진입 시점의 신규 구현 sprint 는 새 branch 이름 별도 결정. status `⏳ planned` 유지 (구현 미완료). 본 §6 (a) 본문 + [release_v1_roadmap.md §3.5 N-13 row](../planning/release_v1_roadmap.md) + [sprint plan §3.3/§5/§6](../planning/2026-06-12-inbound-source-routing-sprint-plan.md) + [traceability report.md §6](../traceability/report.md) + 메모리 4 file 동기화. 신규 ID 발급 0건 (housekeeping follow-up). |
| 2026-06-12 | **N-13 follow-up 3 branch 종합 v1.1 sprint -a 구현 정합** (sprint `feat/work_260612-7-v1-1-inbound-source-impl`) — 본 §6 (a) 본문 + §7 row 추가. routing/auto_route.go (NEW, 3 case pattern matcher + graceful degradation, AutoRouter interface) + voc_handler.go 통합 (AutoRouter.Route() + RouteVoc() + auto_routed envelope 응답) + IT 3 case (GiteaOK / NoMatch / RouteErrorDegradation) + openapi.yaml 정합 (PATCH /platforms inbound_source_type + inbound_source_config + POST /dev-requests/{id} auto_routed 응답 + DevRequestVoc schema) + e2e TC-INBOUND-SRC-01 (NEW, seed platform 사용, POST platform 단계 제거 정공법) + routePermissionTable (PATCH /platforms/:platform_id 가 ResourcePlatforms+ActionEdit 으로 cover, 변경 불요). N-13 follow-up 3 branch 결정 (PR #573) 종합: A (PR #574 e2e seed fix) + B (PR #575 signout timeout fix) + C (본 sprint 구현 follow-up). backend foundation (PR A-1: migration 000007 + domain.Platform 2 field + UpdatePlatformInboundSource + view.UpdatePlatform inboundTouched + 4 UT) 은 main 에 byte-identical 포함 (PR #549 T-d-72-2 wiki mirror). 본 sprint = PR A-2 만. **status `⏳ planned` → `✅ resolved (implemented, 2026-06-12)`** — 9 ID row 정합 (REQ-FR-113 / UC-DEV-REQ-15 / ARCH-23 / API-103 / RM-DEV-REQ-15 / IMPL-inbound-source-01 / IMPL-platform-patch-02 / UT-inbound-source-01 / TC-INBOUND-SRC-01). 신규 ID 발급 0건 (PR #547 의 ID slot 정합의 구현 정합). release_v1_roadmap.md §3.5 N-13 row + §4.2 v1.1 milestone + §9 + traceability report.md §2.1~§2.6 9 ID row + 메모리 4 file 동기화. Tier: **공용** (코드 + ADR + openapi + e2e 모두 사내 한정 정보 미포함). CI 11/12 PASS (backend-unit + backend-integration + frontend-unit + e2e shard 1/2/3 + openapi lint + workflow-lint + migration prefix + changed-paths 모두 PASS; E2E Internal = OBSOLETE, PR #578 의 e2e-internal 폐기 결정 정합). |
