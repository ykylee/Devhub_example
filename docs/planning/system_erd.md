# DevHub 시스템 ERD 카탈로그

- 문서 목적: 코드베이스 전체 모듈의 데이터 모델을 ERD로 분리 관리한다.
- 범위: 현행 DB 스키마 + Project 확장 ERD + External Integration/HomeLab ERD 초안.
- 대상 독자: Backend 설계/구현 담당, 데이터 모델 리뷰어, 추적성 리뷰어.
- 상태: draft
- 최종 수정일: 2026-05-15 (External Integration/HomeLab ERD 초안 추가)
- 관련 문서: [Usecase 카탈로그](./system_usecases.md), [요구사항](../requirements.md), [API 계약](../backend_api_contract.md)

## 1. 기준

- 스키마 기준: `backend-core/migrations/*.up.sql`
- 현행 소스 기준: `backend-core/internal/store`, `backend-core/internal/httpapi`

## 2. 모듈별 ERD

### 2.1 Gitea Ingest / Snapshot

```mermaid
erDiagram
    REPOSITORIES ||--o{ ISSUES : has
    REPOSITORIES ||--o{ PULL_REQUESTS : has
    REPOSITORIES ||--o{ CI_RUNS : has

    WEBHOOK_EVENTS {
      bigint id PK
      text dedupe_key UK
      text event_type
      text status
      jsonb payload
    }
    REPOSITORIES {
      bigint id PK
      bigint gitea_repository_id UK
      text full_name UK
    }
    ISSUES {
      bigint id PK
      bigint repository_id FK
      bigint number
      text state
    }
    PULL_REQUESTS {
      bigint id PK
      bigint repository_id FK
      bigint number
      text state
    }
    CI_RUNS {
      bigint id PK
      text external_id UK
      bigint repository_id FK
      text status
    }
    RISKS {
      bigint id PK
      text risk_key UK
      text status
      text impact
    }
```

### 2.2 Organization / Users / RBAC

```mermaid
erDiagram
    ORG_UNITS ||--o{ ORG_UNITS : parent_child
    ORG_UNITS ||--o{ USERS : primary_current
    USERS ||--o{ UNIT_APPOINTMENTS : assigned
    ORG_UNITS ||--o{ UNIT_APPOINTMENTS : contains
    RBAC_POLICIES ||--o{ USERS : role_fk

    ORG_UNITS {
      bigint id PK
      text unit_id UK
      text parent_unit_id FK
      text unit_type
      text leader_user_id
    }
    USERS {
      bigint id PK
      text user_id UK
      text email UK
      text role FK
      text primary_unit_id FK
      text current_unit_id FK
      text kratos_identity_id UK
    }
    UNIT_APPOINTMENTS {
      bigint id PK
      text user_id FK
      text unit_id FK
      text appointment_role
    }
    RBAC_POLICIES {
      text role_id PK
      boolean is_system
      jsonb permissions
    }
```

### 2.3 Command / Audit

```mermaid
erDiagram
    COMMANDS ||--o{ AUDIT_LOGS : emits

    COMMANDS {
      bigint id PK
      text command_id UK
      text command_type
      text target_type
      text target_id
      text status
      jsonb request_payload
      jsonb result_payload
    }
    AUDIT_LOGS {
      bigint id PK
      text audit_id UK
      text actor_login
      text action
      text target_type
      text target_id
      text command_id FK
      text request_id
      text source_type
      text source_ip
    }
```

### 2.4 HRDB

```mermaid
erDiagram
    HRDB_PERSONS {
      text system_id PK
      text employee_id UK
      text name
      text department_name
      text email
      timestamptz updated_at
    }
```

### 2.5 Application/Repository/Project 확장 초안

```mermaid
erDiagram
    APPLICATIONS ||--o{ APPLICATION_REPOSITORIES : contains
    APPLICATIONS ||--o{ PROJECTS : governs
    REPOSITORIES ||--o{ PROJECTS : hosts
    SCM_PROVIDERS ||--o{ APPLICATION_REPOSITORIES : provides
    REPOSITORIES ||--o{ PR_ACTIVITIES : emits
    REPOSITORIES ||--o{ BUILD_RUNS : runs
    REPOSITORIES ||--o{ QUALITY_SNAPSHOTS : measures
    PROJECTS ||--o{ PROJECT_INTEGRATIONS : connects
    PROJECTS ||--o{ PROJECT_MEMBERS : has
    USERS ||--o{ PROJECT_MEMBERS : joins
    USERS ||--o{ APPLICATIONS : owns
    USERS ||--o{ PROJECTS : owns

    APPLICATIONS {
      uuid id PK
      text key UK
      text name
      text description
      text status
      text visibility
      text owner_user_id FK
      date start_date
      date due_date
      timestamptz archived_at
      timestamptz created_at
      timestamptz updated_at
    }
    APPLICATION_REPOSITORIES {
      uuid application_id PK
      text repo_provider PK
      text repo_full_name PK
      text external_repo_id
      timestamptz last_sync_at
      text sync_status
      text sync_error_code
      boolean sync_error_retryable
      timestamptz sync_error_at
      text role
      timestamptz linked_at
    }
    SCM_PROVIDERS {
      text provider_key PK
      text display_name
      boolean enabled
      text adapter_version
      timestamptz created_at
      timestamptz updated_at
    }
    PROJECTS {
      uuid id PK
      uuid application_id FK
      bigint repository_id FK
      text key
      text name
      text description
      text status
      text visibility
      text owner_user_id FK
      date start_date
      date due_date
      timestamptz archived_at
      timestamptz created_at
      timestamptz updated_at
    }
    PROJECT_INTEGRATIONS {
      uuid id PK
      uuid project_id FK
      text scope
      text integration_type
      text external_key
      text url
      text policy
      timestamptz created_at
      timestamptz updated_at
    }
    PROJECT_MEMBERS {
      uuid project_id PK
      text user_id PK
      text project_role
      timestamptz joined_at
    }
    PR_ACTIVITIES {
      bigint id PK
      bigint repository_id FK
      text external_pr_id
      text event_type
      text actor_login
      timestamptz occurred_at
      jsonb payload
    }
    BUILD_RUNS {
      bigint id PK
      bigint repository_id FK
      text run_external_id UK
      text branch
      text commit_sha
      text status
      integer duration_seconds
      timestamptz started_at
      timestamptz finished_at
    }
    QUALITY_SNAPSHOTS {
      bigint id PK
      bigint repository_id FK
      text tool
      text ref_name
      text commit_sha
      numeric score
      boolean gate_passed
      timestamptz measured_at
    }
```

> **합성 키 메모**:
> - `APPLICATION_REPOSITORIES.PK = (application_id, repo_provider, repo_full_name)` — 동일 `repo_full_name` 이 서로 다른 provider 에 존재할 수 있으므로 provider 를 PK 에 포함. `docs/planning/project_management_concept.md` §7 / §13.3 와 일치.
> - `PROJECT_MEMBERS.PK = (project_id, user_id)` — 동일 사용자의 중복 멤버십 차단.
> - `PROJECT_INTEGRATIONS` 는 단일 `id` PK + (`scope`, `project_id` 또는 `application_id`, `integration_type`, `external_key`) 조합에 unique 인덱스(설계 단계 확정).

### 2.6 External Integration / HomeLab 확장 초안

```mermaid
erDiagram
    INTEGRATION_PROVIDERS ||--o{ INTEGRATION_BINDINGS : binds
    INTEGRATION_PROVIDERS ||--o{ INTEGRATION_EVENTS : ingests
    INTEGRATION_PROVIDERS ||--o{ INFRA_NODES : observes
    INFRA_NODES ||--o{ INFRA_SERVICES : hosts
    APPLICATIONS ||--o{ INTEGRATION_BINDINGS : scoped
    PROJECTS ||--o{ INTEGRATION_BINDINGS : scoped

    INTEGRATION_PROVIDERS {
      uuid provider_id PK
      text provider_key UK
      text provider_type
      text display_name
      boolean enabled
      text auth_mode
      jsonb capabilities
      text sync_status
      timestamptz last_sync_at
      text last_error_code
      timestamptz created_at
      timestamptz updated_at
    }
    INTEGRATION_BINDINGS {
      uuid binding_id PK
      text scope_type
      text scope_id
      uuid provider_id FK
      text external_key
      text policy
      timestamptz created_at
      timestamptz updated_at
    }
    INTEGRATION_EVENTS {
      bigint id PK
      uuid provider_id FK
      text event_key UK
      text global_event_id UK
      text trace_id
      text resource_type
      text event_type
      text status
      jsonb payload
      timestamptz occurred_at
      timestamptz received_at
      timestamptz processed_at
    }
    INFRA_NODES {
      text node_id PK
      uuid provider_id FK
      text hostname
      text ip_address
      text environment
      text status
      jsonb metrics
      timestamptz observed_at
    }
    INFRA_SERVICES {
      text service_id PK
      text node_id FK
      text name
      text version
      int port
      text health_status
      jsonb metadata
      timestamptz observed_at
    }
```

> **스코프 FK 메모**:
> - `INTEGRATION_BINDINGS.scope_type` 이 `application` 인 경우 `scope_id -> applications.id`, `project` 인 경우 `scope_id -> projects.id` 를 의미한다.
> - 물리 FK는 polymorphic 제약으로 단일 컬럼에 직접 강제하기 어렵기 때문에 앱 레이어 + partial unique 인덱스로 보완한다.

> **이벤트 추적성 메모**:
> - `INTEGRATION_EVENTS` 는 `global_event_id`/`trace_id` 를 통해 바인딩/도메인 엔터티와의 논리 연결을 추적한다.
> - 외부 시스템 키(`external_key`, `resource_type`, `event_key`) 조합으로 연관 관계를 복원하며, 물리 FK 없이도 재처리/감사 추적이 가능하도록 설계한다.

## 3. 통합 ERD (현행 + Project 확장)

```mermaid
erDiagram
    WEBHOOK_EVENTS ||--o{ REPOSITORIES : ingest
    REPOSITORIES ||--o{ ISSUES : has
    REPOSITORIES ||--o{ PULL_REQUESTS : has
    REPOSITORIES ||--o{ CI_RUNS : has

    ORG_UNITS ||--o{ ORG_UNITS : parent
    ORG_UNITS ||--o{ USERS : belongs
    USERS ||--o{ UNIT_APPOINTMENTS : assigned
    ORG_UNITS ||--o{ UNIT_APPOINTMENTS : appointment
    RBAC_POLICIES ||--o{ USERS : role_fk

    COMMANDS ||--o{ AUDIT_LOGS : emits

    APPLICATIONS ||--o{ APPLICATION_REPOSITORIES : contains
    APPLICATIONS ||--o{ PROJECTS : governs
    REPOSITORIES ||--o{ PROJECTS : hosts
    PROJECTS ||--o{ PROJECT_INTEGRATIONS : connects
    PROJECTS ||--o{ PROJECT_MEMBERS : has
    USERS ||--o{ PROJECT_MEMBERS : joins
    USERS ||--o{ PROJECTS : owns
    USERS ||--o{ APPLICATIONS : owns
    APPLICATION_REPOSITORIES }o--|| REPOSITORIES : maps
    REPOSITORIES ||--o{ PR_ACTIVITIES : emits
    REPOSITORIES ||--o{ BUILD_RUNS : runs
    REPOSITORIES ||--o{ QUALITY_SNAPSHOTS : measures
    INTEGRATION_PROVIDERS ||--o{ INTEGRATION_BINDINGS : binds
    INTEGRATION_PROVIDERS ||--o{ INTEGRATION_EVENTS : ingests
    INTEGRATION_PROVIDERS ||--o{ INFRA_NODES : observes
    INFRA_NODES ||--o{ INFRA_SERVICES : hosts
    APPLICATIONS ||--o{ INTEGRATION_BINDINGS : scoped
    PROJECTS ||--o{ INTEGRATION_BINDINGS : scoped
```

## 4. 설계/구현 반영 규칙

1. 신규 API 계약은 대응 ERD 엔티티를 참조해야 한다.
2. 신규 마이그레이션은 본 문서의 ERD 섹션 번호를 커밋/PR에 명시한다.
3. 추적성 매트릭스에서 UC/ARCH/API/IMPL가 동일 모듈 ERD를 참조하도록 유지한다.
