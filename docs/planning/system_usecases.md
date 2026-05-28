# DevHub 시스템 Usecase 카탈로그

- 문서 목적: 코드베이스 전체 모듈 기준으로 Usecase를 정의하고, 요구사항(REQ)과 설계(ARCH/API) 사이의 중간 추적 단위를 제공한다.
- 범위: backend-core 주요 모듈(auth/account/org/rbac/gitea ingest/command/audit/realtime/store) + Application/Project 도메인 + Repository SCM 연동/Lifecycle + External Integration.
- 대상 독자: 프로젝트 리드, 시스템 관리자, Backend/Frontend 설계 담당, 추적성 리뷰어.
- 상태: draft
- 최종 수정일: 2026-05-27 (§2.14 Repository SCM 연동+Lifecycle UC-REPO 신규 + §2.12 INT 에 auth_mode/test-connection/Gitea pull sync/webhook alias UC-INT-15..18 추가)
- 관련 문서: [요구사항](../requirements.md), [아키텍처](../architecture.md), [API 계약](../backend_api_contract.md), [ERD 카탈로그](./system_erd.md), [통합 로드맵](../development_roadmap.md)

## 1. 모듈 기준

| 모듈 | 코드 기준 |
| --- | --- |
| Auth/OIDC | `backend-core/internal/auth`, `backend-core/internal/httpapi/auth*.go` |
| Account | `backend-core/internal/httpapi/accounts_admin.go`, `account_password.go` |
| Organization | `backend-core/internal/httpapi/organization.go`, `backend-core/internal/store/users_units.go` |
| RBAC | `backend-core/internal/httpapi/rbac.go`, `permissions.go`, `internal/store/postgres_rbac.go` |
| Gitea Ingest/Snapshot | `backend-core/internal/gitea`, `normalize`, `store/postgres.go` |
| Command | `backend-core/internal/httpapi/commands.go`, `internal/commandworker`, `internal/serviceaction` |
| Audit | `backend-core/internal/httpapi/audit.go`, `internal/store/audit_logs.go` |
| Realtime | `backend-core/internal/httpapi/realtime.go`, `snapshot.go` |
| Application/Project | `backend-core/internal/httpapi/applications.go`, `projects.go`, `repository_ops.go`, `application_rollup.go` |
| Repository SCM 연동/Lifecycle | `backend-core/internal/httpapi/integration_scm_repositories.go`, `domain.go`(draft/publish), `backend-core/internal/gitea`(CreateRepo) |
| External Integration | `backend-core/internal/httpapi/integration_registry.go`, `integrations.go`, `infra_integrations.go`, `backend-core/internal/gitea`(pull sync) |
| Application 개발 대시보드 | `docs/requirements.md` §5.9 기준 (설계/구현 예정) |

## 2. Usecase (모듈별)

### 2.1 Auth/OIDC

| UC ID | Usecase | 성공 조건 | 관련 REQ |
| --- | --- | --- | --- |
| `UC-AUTH-01` | 로그인(code flow) 수행 | 인증 성공 시 actor context 확정 + 토큰 교환 경로 완료 | FR-19, FR-21..24 |
| `UC-AUTH-02` | 로그아웃/세션 종료 | Keycloak/OIDC 세션 무효화 + 재진입 시 재인증 | FR-22, FR-23 |
| `UC-AUTH-03` | 인증 실패 처리 | invalid 토큰/권한 부족 시 표준 에러 및 감사 추적 | NFR-3, NFR-18 |

### 2.2 Account

| UC ID | Usecase | 성공 조건 | 관련 REQ |
| --- | --- | --- | --- |
| `UC-ACCOUNT-01` | 관리자 계정 발급/회수 | 사용자-계정 1:1 정책 유지 + 상태 전이 기록 | FR-15..18, FR-61..67 |
| `UC-ACCOUNT-02` | 관리자 비밀번호 재설정 | 강제 재설정 정책 적용 + audit 기록 | FR-20, FR-22 |
| `UC-ACCOUNT-03` | 본인 비밀번호 변경 | current_password 검증 후 변경 완료 | FR-23, FR-26 |

### 2.3 Organization

| UC ID | Usecase | 성공 조건 | 관련 REQ |
| --- | --- | --- | --- |
| `UC-ORG-01` | 사용자 CRUD | 사용자 등록/조회/수정/삭제 일관 처리 | FR-68..72 |
| `UC-ORG-02` | 조직 단위 CRUD | 계층 단위 생성/수정/삭제 + cycle 방지 | FR-73..76 |
| `UC-ORG-03` | 조직 멤버 배정 | unit appointment 정합성 유지 | FR-77..79 |
| `UC-ORG-04` | 조직 계층 조회 | hierarchy + total_count 반환 | FR-80, NFR-21 |

### 2.4 RBAC

| UC ID | Usecase | 성공 조건 | 관련 REQ |
| --- | --- | --- | --- |
| `UC-RBAC-01` | 정책 조회/편집 | 정책 CRUD + 시스템 role 제약 유지 | FR-27, FR-86 |
| `UC-RBAC-02` | 사용자 role 할당 | user-role FK 정합성 보장 | FR-27 |
| `UC-RBAC-03` | 라우트 권한 강제 | deny-by-default + permission cache 반영 | NFR-26 |

### 2.5 Gitea Ingest/Snapshot

| UC ID | Usecase | 성공 조건 | 관련 REQ |
| --- | --- | --- | --- |
| `UC-GITEA-01` | webhook 수집/검증 | dedupe + 상태 추적 + raw payload 보관 | FR-49..55 |
| `UC-GITEA-02` | 도메인 정규화 | repo/issue/pr/ci/risk snapshot 최신화 | FR-49..55 |
| `UC-GITEA-03` | 조회 API 제공 | dashboard/pipeline API 일관 응답 | FR-1..13, FR-28..36 |

### 2.6 Command

| UC ID | Usecase | 성공 조건 | 관련 REQ |
| --- | --- | --- | --- |
| `UC-CMD-01` | 서비스 액션 명령 생성 | command lifecycle 시작 + idempotency 준수 | FR-58, FR-84, FR-95 |
| `UC-CMD-02` | 리스크 대응 명령 생성 | mitigation command 생성/조회 가능 | FR-59, FR-100 |
| `UC-CMD-03` | 명령 실행/상태 전이 | pending→running→종료 상태 일관성 | FR-101 |

### 2.7 Audit

| UC ID | Usecase | 성공 조건 | 관련 REQ |
| --- | --- | --- | --- |
| `UC-AUD-01` | 감사 로그 기록 | actor/request/source context 포함 저장 | FR-18, FR-26, NFR-4 |
| `UC-AUD-02` | 감사 로그 조회 | target/action/time 필터 조회 가능 | FR-102 |

### 2.8 Realtime

| UC ID | Usecase | 성공 조건 | 관련 REQ |
| --- | --- | --- | --- |
| `UC-RT-01` | 실시간 이벤트 구독 | websocket 연결/권한 검증/이벤트 전달 | FR-56, FR-57, FR-60 |
| `UC-RT-02` | 상태 변경 전파 | command/ci/risk 변화가 채널에 반영 | FR-82, FR-83, FR-104, FR-105 |

### 2.9 Application (총괄 단위)

| UC ID | Usecase | 성공 조건 | 관련 REQ |
| --- | --- | --- | --- |
| `UC-APP-01` | Application 등록 | Application 생성 + owner/KPI/기간 저장 | REQ-FR-APP-001 |
| `UC-APP-02` | Application-Repository 연결 | 1 Application : N Repository 매핑 저장 | REQ-FR-APP-002 |
| `UC-APP-03` | Application 상태 전이/보관 | status/archived 정책 적용 | REQ-FR-APP-001 |
| `UC-APP-04` | Application 상세 조회 | 메타 + 연결 Repository 목록 + Application 마일스톤 + 하위 Repository 마일스톤 롤업 조회 (Application 은 최상위 계층이므로 "상위" 는 없음) | REQ-FR-APP-001,002,004 |
| `UC-APP-05` | Application 관리 권한 검증 | `system_admin`만 쓰기 허용, `pmo_manager` 비활성 시 403 반환 | REQ-FR-PROJ-000 |
| `UC-APP-06` | Repository 운영 스냅샷 조회 | 작업현황/PR/빌드/품질 지표를 repo 단위로 조회 | REQ-FR-APP-005,006,007,008 |
| `UC-APP-07` | 외부 도구 동기화/재동기화 | webhook+pull 기반 동기화, 중복/누락 보정 처리 | REQ-FR-APP-004, REQ-NFR-PROJ-004 |
| `UC-APP-08` | SCM provider 어댑터 라우팅 | repo_provider 기준으로 적절한 어댑터를 선택해 동일 도메인 계약으로 수집/조회 처리 | REQ-FR-APP-009, REQ-NFR-PROJ-005 |
| `UC-APP-09` | Application 상태 전이 관리 | 상태 머신 규칙에 따라 유효 전이만 허용하고 invalid 전이는 거절 | REQ-FR-APP-010 |
| `UC-APP-10` | Application 롤업 메타 조회 | 롤업 지표와 함께 period/filter/weight/data_gap 메타 제공 | REQ-FR-APP-012, REQ-NFR-PROJ-006 |

### 2.10 Project (Repository 하위 운영 단위)

| UC ID | Usecase | 성공 조건 | 관련 REQ |
| --- | --- | --- | --- |
| `UC-PROJ-01` | Project 등록 | Project 생성 + owner/KPI/기간 저장 | REQ-FR-PROJ-001 |
| `UC-PROJ-02` | Project 멤버/책임자 관리 | 멤버 역할/owner 변경 + 감사 추적 | REQ-FR-PROJ-003 |
| `UC-PROJ-03` | 상/하위 마일스톤 매핑 | child->parent 매핑 저장/롤업 가능 | REQ-FR-PROJ-004 |
| `UC-PROJ-04` | Jira/Confluence 연동 정책 관리 | scope/policy 기준 integration 관리 | REQ-FR-PROJ-005,006 |
| `UC-PROJ-05` | Project 상태 전이/보관 | status/archived 정책 적용 | REQ-FR-PROJ-001,008 |
| `UC-PROJ-06` | 내 Project 조회 | 멤버십 기반 목록 + archived 토글 | REQ-FR-PROJ-002 |
| `UC-PROJ-07` | 공개 Project 조회 | visibility=public 목록 조회 | REQ-FR-PROJ-002 |
| `UC-PROJ-08` | Project 상세 조회 | 메타+마일스톤 롤업 조회 | REQ-FR-PROJ-002,004 |
| `UC-PROJ-09` | Project cadence 리포트 조회 | repo sprint 결과를 상위 롤업 | REQ-FR-PROJ-007 |
| `UC-PROJ-10` | Project 관리 권한 검증 | 관리 쓰기 권한의 RBAC/feature-flag 검증 + 감사 추적 | REQ-FR-PROJ-000,010 |

### 2.11 Dev Request (DREQ — 외부 의뢰 수신/검토/등록)

| UC ID | Usecase | 성공 조건 | 관련 REQ |
| --- | --- | --- | --- |
| `UC-DREQ-01` | 외부 시스템의 의뢰 수신 | 인증 통과 + 필수 필드 검증 + idempotency 통과 시 `pending` 으로 저장 / 검증 실패 시 `rejected` (invalid_intake) | REQ-FR-DREQ-001,002 / REQ-NFR-DREQ-001,002 |
| `UC-DREQ-02` | 담당자 dashboard 의 내 대기 의뢰 조회 | 본인의 `assignee_user_id` 와 일치하는 pending/in_review 의뢰 목록을 반환 | REQ-FR-DREQ-004 |
| `UC-DREQ-03` | 의뢰 상세 조회 | 권한 통과 시 의뢰 본문 + status 이력 + 매핑된 application/project id 반환 | REQ-FR-DREQ-004 |
| `UC-DREQ-04` | 의뢰 → Application 등록(promote) | 단일 트랜잭션으로 Application 신규 생성 + DREQ.status=registered + audit | REQ-FR-DREQ-005 |
| `UC-DREQ-05` | 의뢰 → Project 등록(promote) | 단일 트랜잭션으로 Project 신규 생성 + DREQ.status=registered + audit | REQ-FR-DREQ-005 |
| `UC-DREQ-06` | 의뢰 거절(reject) | `rejected_reason` 필수, status → rejected, audit | REQ-FR-DREQ-006 |
| `UC-DREQ-07` | 의뢰 담당자 재할당(reassign) | system_admin 만 가능, 변경 이력 audit | REQ-FR-DREQ-007 |
| `UC-DREQ-08` | 의뢰 닫기(close) | registered/rejected 만 closed 로 전이, 기타는 거부 | REQ-FR-DREQ-008 |
| `UC-DREQ-09` | 전체 의뢰 관리 (system_admin) | 모든 의뢰 조회 + 필터 + 액션 (reassign/close) | REQ-FR-DREQ-004,007,008 |
| `UC-DREQ-10` | 의뢰 상태 머신 가드 | invalid 전이 거절 + dev_request.* audit emit | REQ-FR-DREQ-003 / REQ-NFR-DREQ-003 |
| `UC-DREQ-11` | 실시간 의뢰 알림 및 배지 수신 | `realtimeService`를 통해 `dev_request.created` 실시간 이벤트를 수신하여 Header Bell의 알림 배지 및 드롭다운 카운트가 즉시 반영됨 | REQ-FR-DREQ-012 |
| `UC-DREQ-12` | 의뢰 Promote 및 프로젝트 자동 프리필 연계 | DREQ 상세 모달에서 Promote 클릭 시 `ProjectCreationModal`이 팝업되고 의뢰 메타데이터(Key, Name, Description)가 프리필됨 | REQ-FR-DREQ-013 |

### 2.12 External Integration (ALM/SCM/CI-CD/문서/홈랩)

| UC ID | Usecase | 성공 조건 | 관련 REQ |
| --- | --- | --- | --- |
| `UC-INT-01` | Provider 등록/수정/비활성화 | system_admin 가 provider 메타와 인증 모드를 관리하고 상태를 제어할 수 있다 | REQ-FR-INT-001 |
| `UC-INT-02` | Provider Catalog/Capability 조회 | provider 별 capability와 활성 상태를 조회할 수 있다 | REQ-FR-INT-002 |
| `UC-INT-03` | Webhook ingest 수집 처리 | webhook 이벤트를 검증/저장/정규화 큐로 전달한다 | REQ-FR-INT-003 / REQ-NFR-INT-003 |
| `UC-INT-04` | Scheduled pull 동기화 | 주기 실행으로 누락 데이터를 보정하고 최신 스냅샷을 갱신한다 | REQ-FR-INT-003 |
| `UC-INT-05` | SCM 정규화 처리 | bitbucket/gitea/forgejo 이벤트를 공통 Repository/PR/Activity 모델로 변환한다 | REQ-FR-INT-004 |
| `UC-INT-06` | CI/CD 정규화 처리 | bamboo/jenkins 결과를 공통 BuildRun 모델로 변환한다 | REQ-FR-INT-005 |
| `UC-INT-07` | ALM/문서 링크형 연동 | Jira key 및 Confluence 문서 메타를 Application/Project scope에 연결한다 | REQ-FR-INT-006,007 |
| `UC-INT-08` | 홈랩 Node/Service 인벤토리 관리 | 노드/서비스 등록 및 버전/포트/헬스 상태를 조회할 수 있다 | REQ-FR-INT-008 |
| `UC-INT-09` | 홈랩 토폴로지 조회 | nodes/edges/services 기반 인프라 상태 지도를 제공한다 | REQ-FR-INT-009 |
| `UC-INT-10` | 연동 상태 수명주기 관리 | requested/verifying/active/degraded/disconnected 상태를 일관되게 전이/표시한다 | REQ-FR-INT-010 |
| `UC-INT-11` | 권한 기반 통합 현황 조회 | system_admin 전체 조회, 일반 역할은 scope 기반 제한 조회가 적용된다 | REQ-FR-INT-011 |
| `UC-INT-12` | 연동 운영 감사 추적 | 연동 생성/변경/실패/복구 이벤트가 audit로 추적 가능하다 | REQ-NFR-INT-002 |
| `UC-INT-13` | Provider 장애 격리 처리 | 특정 provider 장애가 전체 수집 파이프라인 중단으로 확산되지 않는다 | REQ-NFR-INT-004 |
| `UC-INT-14` | 연동 데이터 조회 품질 보장 | 페이지네이션/필터/최신스냅샷-이력이 일관된 계약으로 제공된다 | REQ-NFR-INT-005,006 |
| `UC-INT-15` | provider auth_mode 별 자격증명 등록 | auth_mode(token/basic/oauth2/app_password/agent)별 입력을 받아 `api_token`/`auth_secret` 을 write-only(응답엔 `*_set` bool)로 저장하고, outbound 요청 시 `ResolveOutboundAuth` 로 헤더를 구성한다 (API-69..71, migration 000040/000041) | REQ-FR-INT-001 |
| `UC-INT-16` | provider 연결 테스트(test-connection) | base_url 대상으로 GET(5s timeout) 을 보내 reachability 를 확인한다. 사내 internal endpoint 가 합법 대상이라 SSRF 차단을 수용하고 응답 본문은 반환하지 않는다 (API-87) | REQ-FR-INT-001,002 |
| `UC-INT-17` | Gitea pull 동기화 (per-provider) | provider 별 자격증명을 우선 사용(env fallback 금지)해 주기/큐(`integration_sync_jobs`, SKIP LOCKED)로 SCM 데이터를 pull 한다. SCM type + pull\|sync capability + enabled 일 때만 수행, disabled 면 409 | REQ-FR-INT-003,004 |
| `UC-INT-18` | webhook 수신 헤더 alias 정규화 | 범용 ingest 가 `X-Integration-*`→`X-Gitea-*`→`X-Gogs-*` 순으로 서명/이벤트 헤더를 fallback 해석해 vendor 별 헤더 불일치를 흡수한다 (API-73) | REQ-FR-INT-003 |

### 2.13 Onboarding (Keycloak 인증 통과 사용자의 self-service 초기 등록)

| UC ID | Usecase | 성공 조건 | 관련 REQ |
| --- | --- | --- | --- |
| `UC-ONBOARD-01` | 미완료 사용자의 첫 진입 redirect | `GET /api/v1/me` 가 `onboarding_required: true` 반환 + frontend 가 `/devhub/onboarding` 으로 redirect | REQ-FR-ONBOARD-001, REQ-FR-ONBOARD-010 |
| `UC-ONBOARD-02` | Onboarding 제출 — user row 생성 + pending_review | 단일 트랜잭션으로 (INSERT users + onboarding_completed_at + review_status=pending_review + audit) 성공 | REQ-FR-ONBOARD-002, REQ-FR-ONBOARD-003, REQ-FR-ONBOARD-005 |
| `UC-ONBOARD-03` | 소속 검색 (typeahead) | 2글자 이상 입력 시 `/api/v1/organizations/search` 가 최대 20개 결과 반환 (조직명만) | REQ-FR-ONBOARD-004 |
| `UC-ONBOARD-04` | 소속 선택 (tree) | 기존 `/api/v1/organization/hierarchy` 응답 기반 트리에서 선택 → 동일 결과 | REQ-FR-ONBOARD-004 |
| `UC-ONBOARD-05` | 관리자 검토 → reviewed 전이 | system_admin 이 transition 호출 → review_status=reviewed + audit | REQ-FR-ONBOARD-005 |
| `UC-ONBOARD-06` | pending_review 사용자 제한 접근 | 공통 메뉴 + 할당된 과제/저장소/어플리케이션 접근 / 그 외 도메인 API 는 정상 응답 (무소속 처리만 적용) | REQ-FR-ONBOARD-006 |
| `UC-ONBOARD-07` | 사용자 self-service 소속 변경 → pending_review 재진입 | `/account` 에서 `PATCH /api/v1/me` 호출 → primary_unit_id 변경 + review_status=pending_review 재진입 + audit | REQ-FR-ONBOARD-007 |
| `UC-ONBOARD-08` | 관리자 사전 등록 | admin pre-register endpoint 호출 → user row 생성 + onboarding_completed_at=NULL (사용자 첫 로그인 시 onboarding 강제 진입) | REQ-FR-ONBOARD-008 |
| `UC-ONBOARD-09` | Backend gating 403 차단 | 미완료 사용자가 allowlist 외 endpoint 호출 시 403 + `code=onboarding_required` 반환 | REQ-FR-ONBOARD-009 |
| `UC-ONBOARD-10` | Skip → 한정 접근 모드 | "나중에 하기" 액션 → user row 미생성 + 공통 메뉴 + `/devhub/onboarding` + `GET /me` 만 접근 | REQ-FR-ONBOARD-011, REQ-FR-ONBOARD-006 |
| `UC-ONBOARD-11` | Skip 후 dismissible banner | session-scoped skip flag set + 모든 페이지 상단에 banner 노출 (자동 redirect 없음) | REQ-FR-ONBOARD-010 |

### 2.14 Repository SCM 연동 + Lifecycle (원격 SCM ↔ 시스템 저장소 양방향 + draft→publish)

> §2.5 Gitea 가 webhook 수집(push)을, §2.9 Application 이 Application-Repository 링크/운영지표 조회를 다루는 것과 별개로, 본 섹션은 2026-05-27 도입된 (a) 원격 SCM repo 의 inbound import, (b) 시스템→SCM outbound 생성, (c) repository draft→publish lifecycle, (d) scm-owned vs system-owned 소유권 구분을 담는다. (#363/#366/#368/#373)

| UC ID | Usecase | 성공 조건 | 관련 REQ |
| --- | --- | --- | --- |
| `UC-REPO-01` | 원격 SCM repo 목록 조회 | provider(SCM type + pull capability + enabled) 의 원격 저장소 목록을 조회하고 이미 import 된 항목에 `imported` 플래그를 표시한다 (API-88) | REQ-FR-APP-004, REQ-FR-INT-004 |
| `UC-REPO-02` | 선택 import (inbound) | 선택한 원격 repo 를 SCM 재조회 값으로 upsert 한다. source=scm 으로 표시하고 ON CONFLICT 시 SCM mirror 필드만 갱신하며 system-owned 필드는 보존한다 (API-89) | REQ-FR-APP-002,004 |
| `UC-REPO-03` | 시스템→SCM 저장소 생성 (outbound) | push capability + gitea-compat provider 에 대해 `gitea.CreateRepo` 로 원격 저장소를 생성하고 source=system 으로 등록한다 (API-90) | REQ-FR-APP-002, REQ-FR-INT-012 |
| `UC-REPO-04` | repository draft 생성 | provider_key 를 provider_id(FK, migration 000045) 로 해석해 publish 전 draft 상태 repository row 를 생성한다 (`POST /repositories`) | REQ-FR-APP-002,004 |
| `UC-REPO-05` | draft → publish 요청 | draft 상태에서만 허용. provider SCM/push/gitea-compat 검사 통과 후 `gitea.CreateRepo` → `UpsertRepository` 로 발행한다. SCM 생성 실패 시 publish_requested 마킹 + BadGateway 부분실패 경로 | REQ-FR-APP-002, REQ-FR-INT-012 |
| `UC-REPO-06` | 저장소 소유권 구분 (scm vs system) | `source`/`provider_id` 로 SCM-mirror 와 system-owned 를 구분하고, sync upsert 가 system-owned(description 등) 값을 덮어쓰지 않도록 보존한다 (in-memory fake 도 동일 parity) | REQ-FR-APP-004, REQ-FR-APP-009 |
| `UC-REPO-07` | capability 기반 기능 gate | import=pull, sync=pull\|sync, create=push + gitea-compat 로 provider capability 에 따라 기능을 제한하고, 미충족/disabled provider 는 409 로 거부한다 | REQ-FR-APP-009, REQ-FR-INT-002 |

### 2.15 Application 개발 대시보드 (APPDASH)

| UC ID | Usecase | 성공 조건 | 관련 REQ |
| --- | --- | --- | --- |
| `UC-APPDASH-01` | 실시간 타겟 브랜치 빌드 실패 상태 조회 및 딥링크 이동 | 대시보드 최상단 카드에 실패 상태인 모든 빌드가 리포지토리 슬러그, 빌드 번호, 실패 시간, 에러 요약 스니펫과 함께 노출되고, 딥링크 클릭 시 해당 빌드 로그 페이지로 성공적 이동 | REQ-FR-APPDASH-001 |
| `UC-APPDASH-02` | 다차원 코드 품질 점수 및 정적분석 미해결 이슈 조회 | 5점 만점으로 정규화된 가중 품질 스코어와 심각도별(Blocker/Critical/Major) 미해결 정적분석 이슈 건수가 표시되고, 코딩 룰 상세 내역은 저장소 상세 대시보드로 격리 조회됨 | REQ-FR-APPDASH-002 |
| `UC-APPDASH-03` | 하위 프로젝트 진척도 시각화 및 지능형 리스크 감지 | 스토리포인트(SP) 가중치 또는 개수제 방식으로 계산된 하위 프로젝트 진척율(%)이 최상단 주요 영역에 표시되고, 일정 대비 작업 비율 분석을 통해 지연 리스크 경고 배지(`Healthy 🟢`, `Warning 🟡`, `At Risk 🔴`)가 자동 제공됨 | REQ-FR-APPDASH-003 |
| `UC-APPDASH-04` | 연결된 개발 의뢰(DREQ) 목록 조회 및 프로젝트 승격 | 어플리케이션에 매핑된 모든 DREQ 리스트와 상태 조회가 가능하며, Promote 클릭 시 DREQ 메타데이터(Key, Name, Description)를 자동 상속/프리필하는 프로젝트 생성 모달이 팝업되고 단일 트랜잭션으로 연계 생성됨 | REQ-FR-APPDASH-004 |
| `UC-APPDASH-05` | CI/CD 빌드 안정성 및 품질 시계열 트렌드 조회 | 7일 및 30일 간의 평균 빌드 소요 시간 변화 추이와 빌드 성공률 추이 시계열 차트가 정상 렌더링되어 조회됨 | REQ-FR-APPDASH-005 |
| `UC-APPDASH-06` | 리포지토리 가중치 배분 및 정책 갱신 | 롤업 집계에 사용되는 리포지토리별 가중치 정책을 도넛 차트로 확인하고 변경 적용 설정을 제공함 | REQ-FR-APPDASH-006 |
| `UC-APPDASH-07` | 연동 장애 시 우아한 성능 저하(Graceful Degradation) 작동 | 특정 저장소/CI 연동 장애가 발생해도 전체 대시보드가 깨지지 않고 가용한 데이터만 롤업해 표시하며 장애 난 저장소에 대해 `data_gap` 또는 경고 표시 노출 | REQ-NFR-APPDASH-003 |

## 3. 설계/구현 반영 규칙

1. 신규/변경 API는 최소 1개 UC를 참조해야 한다.
2. 신규 마이그레이션은 대응 UC와 연결되어야 한다.
3. 추적성 매트릭스는 REQ→UC→ARCH/API→IMPL→UT/TC 순으로 유지한다.
4. 갱신 메모(2026-05-27): API-87(test-connection)/88(scm-repositories list)/89(import-repositories)/90(create-repository) 및 repository draft→publish(`POST /repositories`, `/repositories/:id/publish`) 는 위 §2.12 UC-INT-15..18 / §2.14 UC-REPO-01..07 에 매핑한다. repository draft→publish(#368) 는 무테스트로 머지되었으므로 대응 UT/TC 보강이 후속 carve 대상이다(02_sdlc_chain_status.md G4 참조). 신규 REQ-FR 는 발급하지 않고 REQ-FR-APP-002/004/009 + REQ-FR-INT-004/012 를 재사용한다.
