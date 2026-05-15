# DevHub 시스템 아키텍처 설계서

- 문서 목적: DevHub 의 시스템 구성 (Frontend / Go Core / Python AI), 서비스 간 통신 방식, 데이터 흐름, UI/UX 시각화 전략, RBAC 정책 단계화를 정의한다.
- 범위: 아키텍처 결정 본문. 구체 API 계약은 `docs/backend_api_contract.md`, 결정 근거는 `docs/adr/000X-*.md`, 도메인 모델 (조직) 은 `docs/organizational_hierarchy_spec.md` 가 source-of-truth.
- 대상 독자: Backend / 프론트엔드 / DevOps 개발자, AI agent, 아키텍처 검토자.
- 상태: accepted (sections marked Draft/Confirmed 안에서 부분 진화)
- 작성일: 2026-04-29
- 최종 수정일: 2026-05-15 (외부 시스템 연동 설계 초안, §8 ARCH-INT 추가)
- 관련 문서: [요구사항 정의서](./requirements.md), [백엔드 API 계약](./backend_api_contract.md), [ADR-0001 IdP](./adr/0001-idp-selection.md), [ADR-0002 RBAC](./adr/0002-rbac-policy-edit-api.md), [ADR-0003 No-Docker CI scope](./adr/0003-no-docker-policy-ci-scope.md), [추적성 매트릭스](./traceability/report.md), [프로젝트 프로파일](../ai-workflow/memory/PROJECT_PROFILE.md).

## 1. 개요
본 문서는 DevHub의 시스템 구성, 서비스 간 통신 방식, 데이터 흐름 및 UI/UX 시각화 전략을 상세히 정의합니다.

## 2. 시스템 컴포넌트 구조

상태 표기 기준:
- `current`: 현재 스캐폴딩 또는 health endpoint 수준으로 존재하는 구성
- `planned`: 아키텍처 계약은 확정되었지만 아직 구현 전인 구성
- `external`: DevHub 외부 시스템 또는 연동 대상

```mermaid
graph TD
    subgraph "Frontend Layer"
        NextJS[Next.js App / React 19<br/>current: scaffold]
    end

    subgraph "Backend Layer (Core)"
        GoCore[Go Core Service / Gin<br/>current: /health]
        GoCore -- "planned: Auth/Business Logic" --> NextJS
        GoCore -- "planned: WebSocket" --> NextJS
    end

    subgraph "Backend Layer (AI/Analysis)"
        PyAI[Python AI Module / FastAPI<br/>current: /health]
        GoCore -. "planned: gRPC (ProtoBuf)" .-> PyAI
        GoCore -- "planned: Analysis Request/Context" --> PyAI
        PyAI -- "planned: Analysis Result" --> GoCore
    end

    subgraph "Data Layer"
        PG[(PostgreSQL<br/>current: compose service)]
        GoCore -- "planned: SQL/JSONB" --> PG
    end

    subgraph "External Integration"
        Gitea[Gitea Server<br/>external]
        Gitea -- "planned: Webhook Events" --> GoCore
        GoCore -- "planned: REST API / Actions Control" --> Gitea
        GiteaRunner[Gitea Runner<br/>external]
        GoCore -- "planned: Health/Config" --> GiteaRunner
    end
```

## 3. 서비스 간 통신 (Internal Communication)

### 3.1 Go Core ↔ Python AI (gRPC)
- **프로토콜:** gRPC (HTTP/2 기반)
- **IDL:** Protocol Buffers (.proto)
- **계약 상태:** 내부 분석 요청/응답의 기본 통신 방식은 gRPC로 확정합니다.
- **구현 상태:** 현재 스캐폴딩에는 `proto/analysis.proto`, Go/Python 생성 명령, Python gRPC 의존성이 포함되어 있습니다. 다만 `backend-ai/main.py`는 아직 FastAPI HTTP health endpoint만 실행하며, `50051`은 Docker Compose에 예약 노출된 포트일 뿐 실제 gRPC 서버와 Go Core client/server 연동은 후속 구현 범위입니다.
- **데이터 접근 경계:** 초기 구현에서 Python AI는 PostgreSQL에 직접 접근하지 않습니다. Go Core가 Gitea 이벤트, 로그, 메트릭, 권한 필터링을 처리한 뒤 필요한 분석 입력만 gRPC로 전달합니다.
- **확장 가능성:** 대용량 분석이나 배치 처리가 필요해질 경우 Python AI의 읽기 전용 DB 접근 또는 분석 전용 view/replica를 후속 아키텍처로 검토합니다.
- **선정 이유:** 
    - Go와 Python 간의 고성능 바이너리 통신.
    - 강력한 타입 체크를 통한 인터페이스 정합성 보장.
    - 대용량 로그 데이터 전송 시 스트리밍 기능 활용 가능.

### 3.2 Backend ↔ Frontend (REST & WebSocket)
- **API:** RESTful API (Next.js Data Fetching / TanStack Query)
- **실시간 통신:** **WebSocket**을 기본 계약으로 사용합니다.
    - **용도:** Gitea Actions 빌드 상태 실시간 업데이트, 긴급 리스크 알림, 실시간 이슈 액티비티 피드.
    - **SSE 처리:** SSE는 초기 구현 범위에 포함하지 않습니다. 프록시/운영 환경 제약으로 WebSocket 유지가 어렵다고 확인될 때 별도 fallback으로 재검토합니다.

## 4. 데이터 전략 (Data Strategy)

### 4.0 SCM Provider Adapter 원칙

- 외부 형상관리 연동은 `SCM Adapter Interface`를 기준으로 provider별 구현체를 분리한다.
- Core 도메인은 provider 중립 계약(Repository/PR/Build/Quality/Event)만 사용하고, provider별 API 차이는 어댑터에서 변환한다.
- `repo_provider`를 라우팅 키로 사용해 어댑터를 선택한다.
- 신규 provider 추가는 "어댑터 등록 + 설정"으로 처리하며, 기존 도메인 API/화면 계약 변경을 최소화한다.
- 장애 격리 원칙: 특정 provider 어댑터 장애는 해당 provider ingest 파이프라인으로만 제한하고 전체 수집 파이프라인 중단을 유발하지 않는다.

### 4.1 하이브리드 동기화
- **Webhook:** Gitea의 모든 이벤트를 실시간 수집하여 즉시 반영.
- **Hourly Pull:** 매 시간 전체 상태를 체크하여 동기화 유실 방지 (Reconciliation).

### 4.2 이벤트 수집 파이프라인

Gitea 이벤트 수집은 다음 파이프라인을 기본으로 합니다.

1. **Receive:** Go Core가 Gitea Webhook 이벤트를 수신.
2. **Validate:** Webhook secret/signature를 검증하고 이벤트 타입을 식별. 알 수 없는 이벤트 타입도 원본은 저장하되 처리 상태를 구분.
3. **Persist Raw Event:** payload 원문을 JSONB로 저장하고 event type, delivery id 또는 dedupe key, repository, sender, received_at, processed_at, status를 함께 기록.
4. **Normalize:** 이슈, PR, commit, build, runner 상태 등 도메인 테이블로 정규화.
5. **Apply Domain Update:** 프로젝트/저장소/사용자/권한/상태 테이블을 갱신.
6. **Request Analysis:** 필요한 경우 Go Core가 권한 필터링을 거친 분석 입력을 Python AI에 gRPC로 전달.
7. **Publish Update:** 프론트엔드 실시간 채널에 상태 변경을 전달.

중복 처리는 Gitea delivery id를 우선 idempotency key로 사용합니다. delivery id가 없는 이벤트는 event type, repository id/name, payload hash를 조합한 보조 key를 사용하며, 같은 key는 중복 삽입 또는 중복 처리하지 않습니다.

처리 상태는 `received`, `validated`, `processed`, `failed`, `ignored`를 기본으로 하며, 실패 시 실패 사유와 retry count를 기록합니다. 반복 실패 이벤트는 수동 확인 또는 `ignored` 상태로 전환해 재처리 루프를 방지합니다.

Hourly Pull reconciliation은 Webhook 누락을 보완하는 동기화 경로이며, 가능한 한 Webhook과 동일한 정규화/갱신 경로를 사용합니다. Pull 결과가 기존 상태와 충돌하면 요구사항 문서의 데이터 상충 정책에 따라 사용자 알림 및 PL read-only 노출 기준을 적용합니다.

### 4.3 스토리지 구성
- **PostgreSQL:**
    - 정형 데이터: 사용자, 프로젝트, 권한, 저장소 매핑.
    - 비정형 데이터(JSONB): Gitea 원본 웹훅 이벤트, (v2 예정) AI 분석 리포트 요약.
    - 보존 기간: 운영 로그 1개월, 개인화 데이터(Kudos 등)는 계정 삭제 후 1개월까지 보존.

## 5. UI/UX 및 시각화 전략

### 5.1 인터랙티브 인프라 관리
- **기술:** **React Flow**
- **내용:** Gitea Runner와 프로젝트 간의 구성도를 인터랙티브 다이어그램으로 구현. 사용자가 직접 드래그, 클릭하여 노드 상태 확인 및 제어(재시작 등) 수행.

### 5.2 역할별 진입 우선순위 기반 대시보드
- **개발 대시보드 (Developer Dashboard):** 집중 시간 보호 모드, 개인화된 업무 연혁, 실시간 빌드 현황.
- **관리 대시보드 (Management Dashboard):** 리스크 탐지(7일 임계치), 진행률 시각화, 의사결정 로그.
- **시스템 대시보드 + 시스템 설정 (System Dashboard + System Settings):** 인프라 헬스체크, 알림 임계치 설정, Runner 제어 콘솔.
- **UX 제공 방식:** 역할별 UX는 전용 화면 완전 분리보다 기본 진입 페이지 우선순위로 간접 제공한다.
- **노출 정책:** 시스템 대시보드/시스템 설정은 `system_admin` 권한 사용자에게만 노출한다.

## 6. 보안 및 인증

초기 구현은 Gitea Webhook 수집과 시스템 관리자 기능의 오남용 방지를 우선하며, DevHub 자체 사용자 계정(Account) 기반 1차 인증을 도입한 뒤 Gitea SSO 통합을 후속 단계로 분리합니다. AI 가드너 기반 분석/추천 기능은 v2 범위로 분리합니다.

### 6.1 초기 구현 범위

- **Webhook 검증:** Gitea Webhook endpoint는 `GITEA_WEBHOOK_SECRET` 기반 signature 검증을 필수로 합니다. 검증 실패 이벤트는 도메인 상태를 변경하지 않으며, 원본 저장 여부는 보안 위험을 고려해 최소 metadata 중심으로 기록합니다.
- **서비스 간 권한 경계:** 모든 Gitea 이벤트와 외부 API 호출은 Go Core를 먼저 통과합니다. Python AI는 인증/권한 판단을 직접 수행하지 않고, Go Core가 필터링한 분석 입력만 처리합니다.
- **관리자 접근:** 시스템 관리자 기능은 초기 단계에서 설정 기반 allowlist 또는 seed된 system admin 계정으로 제한합니다. 일반 관리자/PM 권한과 시스템 관리자 권한은 별도 role로 분리합니다.
- **Audit Log:** Runner 제어, Gitea 계정/조직/권한 변경, 알림 임계치 변경, Webhook 재처리/무시 처리, **계정 발급/회수, 비밀번호 변경, 로그인 성공/실패**는 Audit Log 기록 대상입니다.

### 6.2 사용자(User) ↔ 계정(Account) 도메인 분리

DevHub는 사람 단위 식별(User)과 인증 자격(Account)을 분리해 관리합니다. 자세한 정책은 [요구사항 정의서 2.5절](./requirements.md#25-사용자-계정-관리-user-account-management)을 참조합니다. 본 문서는 그 정책을 만족하기 위한 데이터 모델과 인증 흐름만 정의합니다.

#### 6.2.1 데이터 모델

```text
users (이미 존재)
  user_id        text  PK
  email          text  unique
  display_name   text
  role           text  CHECK in (developer, manager, system_admin)
  status         text  CHECK in (active, pending, deactivated)
  primary_unit_id, current_unit_id, is_seconded, joined_at, ...

accounts (신규)
  id              bigserial PK
  user_id         text NOT NULL UNIQUE  REFERENCES users(user_id) ON DELETE CASCADE
  login_id        text NOT NULL UNIQUE
  password_hash   text NOT NULL
  password_algo   text NOT NULL          -- 예: 'bcrypt', 'argon2id'
  status          text NOT NULL CHECK (status IN ('active','disabled','locked','password_reset_required'))
  failed_login_attempts integer NOT NULL DEFAULT 0
  last_login_at   timestamptz
  password_changed_at timestamptz NOT NULL DEFAULT NOW()
  created_at, updated_at timestamptz NOT NULL DEFAULT NOW()
```

`accounts.user_id`의 `UNIQUE` 제약이 1:1 invariant 의 1차 방어선입니다. 도메인 레이어와 HTTP 핸들러도 이 invariant 를 함께 검사하며, 계정 생성 시 동일 사용자에 대한 중복 시도는 `409 Conflict`로 거절합니다. `users` 행이 삭제되면 `ON DELETE CASCADE` 로 계정도 함께 삭제됩니다.

#### 6.2.2 비밀번호 처리 원칙

- 비밀번호 평문은 어떤 경로로도 저장/로깅하지 않습니다. 핸들러 진입 직후 즉시 해시로 변환하고 평문 변수의 수명은 최소화합니다.
- 해시 알고리즘은 bcrypt(cost ≥ 12) 또는 argon2id 중 하나를 선택하며, 선택 결과를 `password_algo` 컬럼에 저장해 향후 알고리즘 회전을 가능하게 합니다.
- 비밀번호 강도는 운영 정책으로 별도 정의하되, 최소 길이/금지 패턴 검사는 핸들러 입력 검증 단계에서 수행합니다.
- 강제 재설정(시스템 관리자) 후 다음 로그인은 비밀번호 변경을 강제하기 위해 계정 상태를 `password_reset_required` 로 설정합니다.

#### 6.2.3 인증 흐름 (1차)

> **결정 (2026-05-07, [ADR-0001](./adr/0001-idp-selection.md))**: DevHub 의 계정/인증 구현은 자체 `accounts` 테이블이 아니라 **Ory Hydra + Ory Kratos** 를 도입한다. DevHub 자체는 Hydra 의 first-party OIDC client 로 동작하고, 다른 앱들도 동일 Hydra 를 OIDC IdP 로 사용할 수 있다. `users` 는 사람·조직 master 로 유지하고 Kratos 가 credential·세션 master 가 된다.

흐름 (사용자가 DevHub Next.js 에서 로그인하는 first-party 케이스 기준):

1. 브라우저가 DevHub Next.js `/login` 에 진입하면 Next.js 는 Hydra `/oauth2/auth` 로 Authorization Code + PKCE 흐름을 시작합니다.
2. Hydra 가 `login_challenge` 와 함께 Next.js login UI 로 redirect 하면, Next.js 는 Kratos public flow API 로 자격 증명을 검증합니다. 실패 카운터/잠금 정책은 Kratos 가 책임집니다.
3. 검증 성공 시 Next.js 는 Hydra `accept login` → first-party client 의 자동 consent 처리 → callback 에서 token endpoint 호출로 ID Token + Access Token + Refresh Token 을 받습니다.
4. Go Core 는 인입 요청의 Bearer access token 을 Hydra JWKS 또는 introspect endpoint 로 검증하고, ID Token `sub` claim 에 담긴 `users.user_id` 를 actor 로 사용합니다. `X-Devhub-Actor` fallback 헤더는 M0 SEC-4 에서 prod 코드 처리가 제거됐고 [ADR-0004](./adr/0004-x-devhub-actor-removal.md) (2026-05-13) 가 폐기 완료를 선언합니다 — 회귀 방지 테스트만 유지합니다.
5. 다른 앱은 Hydra 에 별도 OIDC client 로 등록되어 동일 표준 흐름을 사용합니다. consent UI 노출 여부는 신뢰 경계 결정(ADR-0001 §8) 에 따릅니다.

### 6.3 RBAC 단계화

| 단계 | 범위 | 기준 |
| --- | --- | --- |
| Phase 1 | Webhook secret 검증, system admin role 분리, 관리자 작업 Audit Log | TASK-007 및 초기 시스템 관리자 기능 구현 기준 |
| Phase 2 | Ory Hydra + Kratos 도입, DevHub 의 OIDC client 화, Kratos 기반 자격 증명/로그인/비밀번호 변경/계정 상태 관리, Kratos 이벤트 → DevHub audit log 매핑 | Hydra/Kratos 컨테이너 운영 진입 및 backend Phase 13 완료 시점 ([ADR-0001](./adr/0001-idp-selection.md)) |
| Phase 3 | Gitea 사용자/조직/저장소 권한 동기화, Repository 하위 Project role 매핑 | Application-Repository-Project 매핑과 관리자 대시보드 확장 시점 |
| Phase 4 | Gitea SSO 연동 기반 통합 인증, 자체 계정과의 병행/대체 정책 결정 | 운영 환경 전환 전 별도 보안 검토 후 도입 |

### 6.4 Audit Log 최소 필드

Audit Log는 최소한 `actor_id`, `actor_role`, `action`, `target_type`, `target_id`, `request_id`, `source_ip`, `result`, `reason`, `created_at`을 기록합니다. Webhook 처리 계열 작업은 `actor_id` 대신 `gitea_delivery_id` 또는 dedupe key를 함께 남겨 재처리 경로를 추적합니다. 계정/인증 계열 action 은 다음을 사용합니다.

| action | target_type | 비고 |
| --- | --- | --- |
| `account.created` | `account` | actor=발급한 시스템 관리자 |
| `account.disabled` | `account` | 회수 |
| `account.password_changed` | `account` | actor=본인 또는 시스템 관리자 |
| `account.locked` | `account` | 자동(연속 실패) 또는 수동 |
| `auth.login.succeeded` | `account` | source_ip 필수 |
| `auth.login.failed` | `account` 또는 `login_id` | login_id가 존재하지 않아도 시도는 기록 |

비밀번호 평문, 해시, 임시 비밀번호는 어떤 audit 필드에도 기록하지 않습니다.

## 7. 개발 의뢰 (Dev Request, DREQ) 도메인

외부 시스템에서 들어오는 개발 의뢰를 수신 → 담당자 검토 → application/project 등록(promote) 까지 처리하는 도메인. 컨셉 문서: [`docs/planning/development_request_concept.md`](./planning/development_request_concept.md). 요구사항: [`docs/requirements.md §5.5`](./requirements.md). Usecase: [`UC-DREQ-01..10`](./planning/system_usecases.md).

### 7.1 컴포넌트 (ARCH-DREQ-01)

```
┌──────────────────┐                       ┌──────────────────────────────────────┐
│  External System │ ──── POST /api/v1 ─▶  │  Go Core: dev_requests handler       │
│ (ops portal /    │   /dev-requests       │  ├── auth: 외부 수신용 별도 정책      │
│  ITSM / Jira /   │                       │  │   (REQ-NFR-DREQ-001, ADR 후보)     │
│  사내 워크플로우)│                       │  ├── validate: 필수 필드 + assignee   │
└──────────────────┘                       │  │   존재 / (source_system,           │
                                           │  │   external_ref) idempotency        │
                                           │  ├── store: dev_requests (Postgres)   │
                                           │  └── audit: dev_request.received      │
                                           └────────────┬─────────────────────────┘
                                                        │
                                                        ▼
                                           ┌──────────────────────────────────────┐
                                           │  Frontend: 담당자 dashboard          │
                                           │  + /admin/settings/dev-requests       │
                                           │  └── Promote-to-Application/Project  │
                                           │     (단일 트랜잭션 — REQ-FR-DREQ-005) │
                                           └──────────────────────────────────────┘
                                                        │
                                                        ▼
                                           ┌──────────────────────────────────────┐
                                           │  Application / Project 도메인        │
                                           │  (DREQ.registered_target_id 로 매핑)  │
                                           └──────────────────────────────────────┘
```

### 7.2 상태 머신 (ARCH-DREQ-02)

[컨셉 §2.3](./planning/development_request_concept.md) 의 6-상태 머신 (`received → pending → in_review → registered | rejected | closed`). 모든 전이는 `dev_request.*` audit action 으로 기록.

### 7.3 외부 수신 인증 경계 (ARCH-DREQ-03)

- 외부 수신 endpoint (`POST /api/v1/dev-requests`) 는 일반 사용자 OIDC 흐름이 아닌 **별도 인증 middleware (`requireIntakeToken`)** 를 사용. **[ADR-0012](./adr/0012-dreq-external-intake-auth.md)** 가 옵션 A (API 토큰 + IP allowlist) 를 채택. 옵션 B (HMAC) / C (OAuth client_credentials) 는 후속 단계 마이그레이션 경로.
- 검증 흐름 (ADR-0012 §4.1.2):
  - 외부 호출은 `Authorization: Bearer <plain-token>` 헤더로 도착.
  - middleware 가 `SHA-256(plain-token)` 으로 `dev_request_intake_tokens.hashed_token` lookup.
  - 매칭 없음 또는 `revoked_at IS NOT NULL` → 401.
  - caller IP 가 row 의 `allowed_ips` CIDR 범위 밖 → 401.
  - 검증 성공 시 `source_system` 컨텍스트 주입 + `last_used_at` 갱신 + audit `dev_request.intake_auth_succeeded` emit.
- 본 endpoint 는 `routePermissionTable` 의 `Bypass: true` 또는 별도 `IntakeAuth: true` 플래그로 일반 OIDC enforce 를 건너뛴다.
- 인증 성공 시 `source_system` 은 토큰의 매핑 값에서 자동 채움 (request body 의 self-claim 은 신뢰하지 않음 — spoofing 방지).
- 그 외 endpoint (GET 목록 / 상세 / Promote / Reject / Reassign / Close) 는 일반 OIDC + RBAC + 본 sprint 의 `enforceRowOwnership` 패턴([ADR-0011 §4.2](./adr/0011-rbac-row-scoping.md))으로 보호. 담당자 본인 의뢰 또는 system_admin / pmo_manager 만 가능.

### 7.4 RBAC 자원 (ARCH-DREQ-04)

- 신규 resource `dev_requests` 를 RBAC matrix 에 추가.
- 1차 정책 (MVP):
  - `system_admin`: view + create(외부 수신 server-side, frontend 에서는 미노출) + edit + delete
  - `pmo_manager`: view + edit (담당자 재할당은 제외 — system_admin 만)
  - `manager` / `developer`: view (본인 의뢰만, row-level `actor.login == assignee_user_id`)
- 정책 매핑 표는 backend 구현 sprint 의 migration (`000022_dev_requests` 또는 `000023_rbac_dev_request_resource`) 에서 확정.

### 7.5 데이터 모델 (ARCH-DREQ-05)

```text
dev_requests
  id                      uuid       PK
  title                   text       NOT NULL
  details                 text
  requester               text       NOT NULL
  assignee_user_id        text       NOT NULL  REFERENCES users(user_id) ON DELETE RESTRICT
  source_system           text       NOT NULL
  external_ref            text       NULLABLE  -- (source_system, external_ref) UNIQUE
  status                  text       NOT NULL  CHECK in (received, pending, in_review, registered, rejected, closed)
  registered_target_type  text                 CHECK in (application, project) WHEN status='registered'
  registered_target_id    text                 NULLABLE
  rejected_reason         text                 NOT NULL WHEN status='rejected'
  received_at             timestamptz NOT NULL
  created_at, updated_at  timestamptz NOT NULL DEFAULT NOW()

  CONSTRAINT dev_requests_idempotency_uniq
    UNIQUE (source_system, external_ref)
    WHERE external_ref IS NOT NULL;
  CONSTRAINT dev_requests_registered_target_consistency
    CHECK ( (status = 'registered') = (registered_target_type IS NOT NULL AND registered_target_id IS NOT NULL) );
  CONSTRAINT dev_requests_rejected_reason_required
    CHECK ( (status = 'rejected') = (rejected_reason IS NOT NULL) );
```

application / project 의 `origin_dreq_id` 역참조 컬럼 도입 여부는 REQ-FR-DREQ-009 의 ADR 후속에서 결정.

#### 외부 수신 토큰 테이블 (ADR-0012 §4.1.1)

```text
dev_request_intake_tokens
  token_id        uuid       PK
  client_label    text       NOT NULL  -- 운영용 식별자 (예: "ops_portal")
  hashed_token    text       NOT NULL  UNIQUE  -- SHA-256 hex of plain token
  allowed_ips     jsonb      NOT NULL  -- CIDR 배열
  source_system   text       NOT NULL  -- token 매핑되는 source_system 값
  created_at      timestamptz NOT NULL DEFAULT NOW()
  created_by      text       NOT NULL  REFERENCES users(user_id)
  last_used_at    timestamptz NULLABLE
  revoked_at      timestamptz NULLABLE
```

plain token 은 발급 직후 1회만 admin 에게 노출하고 어디에도 저장하지 않는다 (Kratos password issuance 패턴, [accounts_admin](../backend/) 참조).

### 7.6 Audit action 카탈로그 (ARCH-DREQ-06)

| action | target_type | 비고 |
| --- | --- | --- |
| `dev_request.received` | `dev_request` | 외부 수신, payload 에 source_system / external_ref / assignee |
| `dev_request.registered` | `dev_request` | promote 시점, payload 에 registered_target_type/id |
| `dev_request.rejected` | `dev_request` | rejected_reason 포함 |
| `dev_request.reassigned` | `dev_request` | from / to assignee |
| `dev_request.reopened` | `dev_request` | rejected → pending |
| `dev_request.closed` | `dev_request` | registered/rejected → closed |
| `dev_request.intake_auth_succeeded` | `dev_request_intake_token` | ADR-0012 §4.1.6 — payload `{token_id, client_label, source_ip}`. token plain 값은 절대 기록 안 함. |
| `dev_request.intake_auth_failed` | `dev_request_intake_token` 또는 `route` | ADR-0012 §4.1.6 — payload `{reason, source_ip, header_present, token_prefix_4chars}`. token full 값은 절대 기록 안 함. |
| `auth.row_denied` | `route` | enforceRowOwnership 패턴, 본 도메인 row 거절 |

## 8. 외부 시스템 연동 (Integration) 도메인

컨셉 문서: [`docs/planning/external_system_integration_concept.md`](./planning/external_system_integration_concept.md), 요구사항: [`docs/requirements.md §5.6`](./requirements.md), Usecase: [`UC-INT-01..14`](./planning/system_usecases.md).

### 8.1 컴포넌트 경계 (ARCH-INT-01)

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

### 8.2 동기화 전략 (ARCH-INT-02)

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

### 8.3 데이터 모델 초안 (ARCH-INT-03)

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

### 8.4 보안/권한 경계 (ARCH-INT-04)

- Provider credential 은 평문 저장을 금지한다 (encrypted at rest 또는 external secret manager 참조).
- 연동 생성/수정/비활성화는 `system_admin` 권한만 허용한다.
- 조회는 scope 기반으로 제한한다:
  - `system_admin`: 전체 조회
  - 일반 역할: 자신의 접근 가능한 Application/Project scope 한정
- 감사로그 action namespace: `integration.*`, `infra.node.*`, `infra.service.*`

### 8.5 홈랩 수집 경계 (ARCH-INT-05)

- 홈랩은 infra provider 로 취급한다 (`provider_type=infra`).
- 수집 방식은 1차에 Agent Push 를 기본 후보로 둔다.
  - Agent 가 node/service 상태를 DevHub ingest endpoint 로 전송
  - DevHub 는 마지막 스냅샷 + 상태 변경 이력을 동시 관리
- 수집 실패 시 provider 상태를 `degraded` 로 전이하고 경고를 노출한다.
- Agent payload 최소 계약:
  - `agent_id`, `snapshot_at`, `nodes[]`, `services[]`, `trace_id`
  - 각 node/service 는 `observed_at` 필수
  - 동일 `agent_id + snapshot_at` 재전송은 idempotent 처리

### 8.6 장애 격리 및 복구 (ARCH-INT-06)

- provider별 retry/backoff 정책을 독립적으로 적용한다.
- 특정 provider 의 반복 실패는 circuit-open 상태로 격리하고, 나머지 provider 파이프라인은 지속 처리한다.
- 운영자는 provider 단위로 수동 재동기화(re-sync) 요청을 트리거할 수 있어야 한다.
