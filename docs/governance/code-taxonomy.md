# 코드베이스 카테고리 · 모듈 분류 (Code Taxonomy)

- 문서 목적: DevHub Example의 backend-core / backend-ai / frontend 자산을 도메인 기반 3대 레이어(**Domain**, **Shared**, **Infrastructure**) 및 도메인 내부 **4대 계층**(`view`, `service`, `repository`, `schema`)으로 구조화하여, 개발·리팩토링·에이전트 협업 시 일관된 아키텍처 SoT를 제공한다.
- 범위: backend-core / backend-ai / frontend / infra / scripts / docs / .github/workflows
- 분석 기준 커밋: pr-406 HEAD `536886a` (2026-05-28)
- 상태: active (고도화)
- 관련 문서: [`worker_division.md`](./worker_division.md), [`document-standards.md`](./document-standards.md), [`traceability/report.md`](../traceability/report.md), [`docs/architecture.md`](../architecture.md)

---

## 0. 아키텍처 컨벤션 — 3대 레이어 및 작업 명시

모든 신규 작업(PR, Commit, Backlog, Traceability)은 영향받는 **[레이어 / 도메인 / 계층]**을 명시한다.
*   **PR Title & Commit Prefix**: `<type>(<레이어>/<도메인>-<계층>): <설명>`
    *   예: `feat(domain/application-lifecycle-service): App 상태 전이 규칙 검증 추가`
    *   예: `fix(infra/gitea-scm-service): 백그라운드 동기화 큐 SKIP LOCKED 데드락 해결`
    *   예: `style(shared/ui-foundation-view): 모달 공통 애니메이션 개선`
*   **Traceability & Backlog**: 변경 사항이 속한 세부 컴포넌트(예: `domain/rbac-permissions/repository`)를 기재하여 추적의 해상도를 확보한다.

---

## 1. Top-level 아키텍처 레이어 (3대 분류)

프로젝트 자산은 비즈니스 핵심성 및 외부 기술 종속 여부에 따라 **3개 최상위 레이어**로 격리·관리된다.

```text
┌─────────────────────────────────────────────────────────────┐
│                       1. DOMAIN                             │
│  (Pure Business Logic: Auth, Audit, RBAC, App, DREQ, etc.)  │
│  ├── view (API/UI 접점)                                      │
│  ├── service (비즈니스 제어/연산)                             │
│  ├── repository (영속성 추상화)                              │
│  └── schema (엔티티/DTO 스키마)                              │
└──────────────────────────────┬──────────────────────────────┘
                               │ (추상화 인터페이스를 통한 의존)
                               ▼
┌─────────────────────────────────────────────────────────────┐
│                     3. INFRASTRUCTURE                       │
│  (Concrete Tech Implementations: Keycloak, Gitea, DB, etc.) │
└─────────────────────────────────────────────────────────────┘
  ▲
  │ (공통 유틸리티 및 UI 프레임워크 활용)
┌─────────────────────────────────────────────────────────────┐
│                        2. SHARED                            │
│  (Common Utilities, Configs, Layouts, Design System Tokens)  │
└─────────────────────────────────────────────────────────────┘
```

1.  **DOMAIN (비즈니스 핵심)**: 
    외부 기술이나 특정 데이터베이스 구체 엔진에 종속되지 않는 비즈니스 의사결정 규칙의 핵심이다. 도메인 내부에는 `view`, `service`, `repository`, `schema` 계층이 수직적으로 격리되어 존재한다.
2.  **SHARED (공통 기반)**:
    특정 비즈니스 도메인에 결합되지 않고 시스템 전반에서 유틸리티, 전역 환경 설정, 공통 레이아웃 및 UI 빌딩 블록으로 기능하는 재사용 모듈이다.
3.  **INFRASTRUCTURE (기술 구현 및 외부 연동)**:
    인증 서버(Keycloak), SCM(Gitea), 외부 인사망(HR DB) 등 타사 플랫폼과의 통신, 원시 데이터베이스 마이그레이션(golang-migrate), 그리고 배포/CI 인프라 같은 실제 하드웨어/운영 연동에 결합되는 구체 레이어다.

---

## 2. 레이어별 상세 모듈 및 컴포넌트 매핑 SoT

### 2.1 DOMAIN (비즈니스 핵심 영역)

#### 2.1.1 `auth-session` — 인증·세션 도메인
Keycloak OIDC 연동을 통한 브라우저 토큰 라이프사이클 및 로그인 플로우를 통제한다.
*   **view**: 
    *   Backend: `httpapi/auth.go`, `httpapi/me.go`, `httpapi/identity_resolver.go` (`BearerTokenVerifier` 연계)
    *   Frontend: `app/login`, `app/auth/callback`, `app/auth/logout`, `components/layout/AuthGuard.tsx` (진입 가드)
*   **service**: 
    *   Frontend: `lib/auth/refresh-scheduler.ts` (토큰 만료 재스케줄러), `lib/auth/session-death.ts`, `lib/services/auth.service.ts`
*   **repository**: 없음 (자격 증명 및 세션 수명은 전적으로 Keycloak IdP가 소유)
*   **schema**: 토큰 Claims 구조체 및 OIDC/PKCE 데이터 교환 모델
*   *의존*: `rbac-permissions`, `audit-ops`
*   *관련 ADR*: ADR-0006, ADR-0019, ADR-0020, ADR-0024
*   *E2E Spec*: `auth.spec.ts`, `signout.spec.ts`

#### 2.1.2 `audit-ops` — 감사·운영 도메인
사용자 및 시스템에 의해 트리거된 주요 변경 사항을 캡처하여 감사용 영속 로그를 발행하고 메트릭을 수집한다.
*   **view**: 
    *   Backend: `httpapi/audit.go` (감사 목록), `httpapi/keycloak_events_webhook.go` (webhook push 채널)
*   **service**: 
    *   Backend: `audit/keycloak_event_puller.go` (cron poll 동기화), `audit/user_sync.go` (인사 프로필 동기화 규칙)
*   **repository**: 
    *   Backend: `store/audit_logs.go` (`audit_logs` 테이블 쿼리), `store/postgres.go` 내 cursor 저장소
*   **schema**: 
    *   Backend: `domain/audit.go` (AuditLog 엔티티)
    *   DB: `audit_logs` (000003), `event_cursors` (000031)
*   *의존*: `auth-session` (요청 actor 식별)
*   *관련 ADR*: ADR-0020 sub-carve E

#### 2.1.3 `rbac-permissions` — 권한 통제 도메인
역할(Role), 자원(Resource), 액션(Action) 매트릭스를 기반으로 한 다차원 접근 제어를 수행한다.
*   **view**: 
    *   Backend: `httpapi/permissions.go` (라우트 접근 가드), `httpapi/rbac.go` (RBAC 관리 API), `httpapi/authz.go`
    *   Frontend: `/admin/settings/permissions` 페이지, `components/organization/PermissionEditor.tsx`
*   **service**: 
    *   Backend: `httpapi/permissions.go` 내 `PermissionCache` (권한 메모리 캐시 및 평가 엔진)
    *   Frontend: `lib/services/rbac.service.ts`
*   **repository**: 
    *   Backend: `store/postgres_rbac.go` (RBAC 정책 쿼리)
*   **schema**: 
    *   Backend: `domain/rbac.go` (역할 및 권한 매핑 데이터 모델, 역할 상수가 정의된 곳)
    *   DB: `rbac_policies` (000005, 000018, 000021, 000024, 000026)
*   *관련 ADR*: ADR-0002, ADR-0007, ADR-0011 (Row-scoping)
*   *E2E Spec*: `admin-permissions.spec.ts`, `rbac-routes.spec.ts`

#### 2.1.4 `organization-management` — 조직·사용자 도메인
인사 조직 트리, 직무 Appointments, 사용자 마스터 프로필의 변경 및 조회를 제어한다.
*   **view**: 
    *   Backend: `httpapi/organization.go` (부서 조회 API), `httpapi/organizations_search.go`, `httpapi/hr_lookup.go`
    *   Frontend: `/admin/settings/organization`, `/admin/settings/users` 페이지, `components/organization/OrgTree.tsx`
*   **service**: 
    *   Backend: 조직 노드 유효성(순환 의존성 검사), 임원 할당 규칙 제어
    *   Frontend: `lib/services/identity.service.ts`
*   **repository**: 
    *   Backend: `store/users_units.go` (사용자-부서 맵 및 인사 정보 쿼리)
*   **schema**: 
    *   Backend: `domain/primary_unit.go` (주부서 매핑 모델), `domain/user.go`
    *   DB: `users`, `org_units`, `unit_appointments` (000004, 000019), `org_units_total_count_mv` (000011)
*   *관련 ADR*: ADR-0008, ADR-0009, ADR-0010
*   *E2E Spec*: `admin-org-crud.spec.ts`, `admin-users-crud.spec.ts`

#### 2.1.5 `onboarding` — 신규 사용자 온보딩 도메인
인사 등록 후 최초 로그인하는 신규 사용자의 승인 절차, 온보딩 게이트 가드 및 초기 가이드를 통제한다.
*   **view**: 
    *   Backend: `httpapi/me_onboarding.go` (온보딩 신청 API), `/api/v1/me` 응답 게이트 미들웨어 (`onboardingGate`)
    *   Frontend: `/onboarding` 게이트 페이지
*   **service**: 온보딩 게이트 통과 규칙(필수 정보 수집 검증 및 관리자 통보)
*   **repository**: (해당 없음 - `organization-management` 유저 테이블 공유)
*   **schema**: 온보딩 폼 입력 필드 및 가벨 플래그 모델
*   *의존*: `auth-session`, `organization-management`
*   *관련 ADR*: ADR-0021
*   *E2E Spec*: `onboarding-first-login.spec.ts`

#### 2.1.6 `application-lifecycle` — 애플리케이션·프로젝트 관리 도메인
핵심 비즈니스 엔티티인 Application과 Project의 CRUD, 상태 전이 머신, 그리고 롤업 요약 데이터 생성을 담당한다.
*   **view**: 
    *   Backend: `httpapi/applications.go`, `httpapi/projects.go`, `httpapi/application_rollup.go`
    *   Frontend: `/applications`, `/projects` 페이지, `components/project/ProjectCreationModal.tsx`
*   **service**: Application/Project의 생명주기 상태 머신(활성, 유휴, 폐기), 롤업 계산 규칙
*   **repository**: 
    *   Backend: `store/applications.go` (애플리케이션 및 프로젝트 DB 쿼리), `store/repository_ops.go` (상태 스냅샷)
*   **schema**: 
    *   Backend: `domain/application.go`
    *   DB: `applications` (000013), `projects` (000015), `project_members`, `repo_ops_snapshots` (000017)
*   *관련 ADR*: ADR-0011, ADR-0014
*   *E2E Spec*: `admin-applications.spec.ts`, `admin-projects.spec.ts`

#### 2.1.7 `repository-integration` — 저장소 관리 도메인
SCM 저장소를 프로젝트와 연결하고, 가져오기(Import) 및 코드 자산 매핑을 수행한다.
*   **view**: 
    *   Backend: `httpapi/integration_scm_repositories.go` (저장소 import API), `httpapi/domain.go`
    *   Frontend: `/repositories` 페이지, `components/project/RepositoryLinkModal.tsx`
*   **service**: SCM 저장소와 DevHub 내부 프로젝트 간의 맵핑 검증, 강제 동기화 비즈니스 규칙
*   **repository**: `store/applications.go` 내 Get/Upsert/ListRepositories 함수군
*   **schema**: 
    *   Backend: SCM Repository 도메인 모델
    *   DB: `repositories` (000002, 000042)
*   *E2E Spec*: `repositories-ui.spec.ts`, `repositories-detail-negative.spec.ts`

#### 2.1.8 `dev-request` — 개발 의뢰 (DREQ) 도메인
외부 연동 채널 또는 사내 채널을 통해 들어온 신규 시스템 개발 의뢰(DREQ)의 인입, 검토, promote(Application/Project 자동 생성) 프로세스를 관리한다.
*   **view**: 
    *   Backend: `httpapi/dev_requests.go` (목록/조정), `httpapi/dev_request_intake_auth.go` (인입 전용), `httpapi/dev_request_intake_tokens_admin.go`
    *   Frontend: `/dev-requests` 페이지, `components/dev-request/DevRequestDetailModal.tsx`
*   **service**: 
    *   Backend: DREQ 6단계 상태 머신 전이 규칙, promote 트랜잭션 비즈니스 규칙, intake 토큰 만료 처리
    *   Frontend: `lib/services/dev_request.service.ts`
*   **repository**: 
    *   Backend: `store/dev_requests.go`, `store/dev_request_intake_tokens.go`
*   **schema**: 
    *   Backend: `dev_requests` 및 `dev_request_intake_tokens` 엔티티 정의
    *   DB: `dev_requests` (000022), `dev_request_intake_tokens` (000023)
*   *관련 ADR*: ADR-0012, ADR-0013, ADR-0014, ADR-0017
*   *E2E Spec*: `dev-requests.spec.ts`

#### 2.1.9 `integration-registry` — 통합 등록 도메인
SCM(Gitea) 및 비-SCM(Jira, Confluence, Homelab 등) 연동 공급자와 바인딩을 통합 관리하며, **ProviderModal** UI 공통 컴포넌트의 소유권을 가집니다.
*   **view**: 
    *   Backend: `httpapi/integration_registry.go`, `httpapi/integrations.go`, `httpapi/external_task_handler.go`
    *   Frontend: `/admin/settings/integrations` 페이지, **`components/integration/ProviderModal.tsx`** (Codex 리뷰를 수렴하여 `auth-session`에서 이관)
*   **service**: 외부 Preset 매핑 규칙, Sync 작업 스케줄링 큐 비즈니스 로직, Task Ingestion 라우팅 규칙
*   **repository**: 
    *   Backend: `store/integration_registry.go` (통합 프로바이더/바인딩), `store/external_task_store.go`
*   **schema**: 
    *   Backend: `integration_providers` 및 `integration_bindings` 스키마, `external_task_items` 스키마
    *   DB: `integration_providers` (000028), `integration_bindings` (000040), `external_task_items` (000046)
*   *관련 ADR*: ADR-0015
*   *E2E Spec*: `admin-integrations.spec.ts`, `admin-integration-bindings.spec.ts`

#### 2.1.10 `realtime` — 실시간 통신 도메인
WebSocket을 통한 백그라운드 이벤트 전송 및 단일-사용(single-use) WebSocket 인증 티켓 발행/검증 생명주기를 다룬다.
*   **view**: 
    *   Backend: `httpapi/realtime.go` (WS Hub 연결), `httpapi/realtime_ticket.go` (티켓 발급 API)
    *   Frontend: `lib/services/websocket.service.ts`
*   **service**: WebSocket 브로드캐스트 이벤트 필터링 규칙(RBAC 권한 재검사), 티켓 60s TTL 만료 검증 로직
*   **repository**: 
    *   Backend: `store/realtime_tickets.go` (티켓 원자적 영속 및 소비)
*   **schema**: 티켓 및 실시간 이벤트 프레임
    *   DB: `realtime_tickets` (000035)
*   *의존*: `auth-session`
*   *관련 ADR*: ADR-0024

---

### 2.2 SHARED (공통 기능 영역)

*   `config`: 전역 환경 설정 로더 (`backend-core/internal/config`) 및 프론트엔드 `.env` 매핑
*   `logger`: 시스템 표준 로그 수집 어댑터
*   `utils`: 공통 유틸리티 헬퍼 (`lib/utils/cn.ts`, `lifecycle-status.ts` 등)
*   `ui-foundation`: 프론트엔드 공통 UI 및 레이아웃 컴포넌트 전체
    *   `components/ui/*`: `Modal`, `Badge`, `Toast`, `PageState`, `FilterBar`, `ComboBox`, `DestructiveConfirmModal`
    *   `components/layout/*`: `Header.tsx`, `Sidebar.tsx`, `AuthGuard.tsx`
    *   `app/globals.css`, `next.config.ts`

---

### 2.3 INFRASTRUCTURE (외부 기술 및 연동 구현체)

#### 2.3.1 `keycloak-idp` — Keycloak IDP 연동
*   `infra/idp/keycloak-event-listener-spi/`: Keycloak SPI Java 이벤트 리스너 플러그인 소스
*   `internal/auth/keycloak_verifier`: Keycloak JWKS 검증 및 stale-while-error fallback (ADR-0020 sub-carve D) 구현체
*   `infra/idp/sql/`: Keycloak 시드 데이터 (`002_seed_e2e_users.sql` 등)
*   *관련 ADR*: ADR-0022, ADR-0023

#### 2.3.2 `gitea-scm` — Gitea SCM 연동 구현체
*   `gitea/*`: Gitea API 클라이언트 및 백그라운드 sync 워커, webhook 서명 검증 어댑터 구현체
*   `normalize/gitea`: Gitea webhook JSON 데이터를 DevHub 도메인 모델로 정규화하는 모듈
*   *관련 ADR*: ADR-0003

#### 2.3.3 `hrdb` — 인사망 데이터 어댑터
*   `hrdb/postgres`: 실 PostgreSQL 인사 데이터 연동 구현체
*   `hrdb/mock`: 유닛/E2E 테스트용 Mock 인사 데이터 어댑터
*   *관련 ADR*: ADR-0008, ADR-0010

#### 2.3.4 `commandworker` — 명령어 실행기
*   `commandworker/{worker, live_worker}`: 실시간 인프라 명령어 폴링 및 실행 에이전트 구체 구현
*   `serviceaction/executor`: 명령어 실행용 sandbox 및 프로덕션 샌드박스 연동 구현

#### 2.3.5 `database-migration` — 데이터베이스 스키마 마이그레이션
*   `backend-core/migrations/`: `000001_webhook_events`부터 `000046_add_external_task_items`까지 전체 golang-migrate SQL 파일

#### 2.3.6 `deployment-automation` — 배포 스크립트 및 인프라 템플릿
*   `scripts/`: 배포 전처리 스크립트군 (`deploy-up.sh`, `verify-keycloak-groups.sh`, `setup-test-db.sh` 등)
*   `infra/nginx/`: Nginx 역프록시 템플릿 (`devhub.deploy.conf.template`, `devhub.native.conf`)
*   `docker-compose.deploy.yml`, `docker-compose.yml`
*   *관련 ADR*: ADR-0018

---

## 3. 리팩토링 로드맵 (3대 아키텍처 기준)

계층형 구조 전환으로 인해 식별된 리팩토링 후보군.

### P0 — 즉시 해결 (계층 위반 수정 및 소유권 오류 정정)
1.  **P0-1. ProviderModal 소유권 이관 완료**:
    *   *내용*: `components/integration/ProviderModal`를 `auth-session`에서 `integration-registry` 도메인 소속으로 SoT를 전면 수정하고 관련 이력을 연계한다. (PR #406 Codex 피드백 수렴 완료)
2.  **P0-2. `store/applications` 파일 분할 (LoC 1172)**:
    *   *내용*: 애플리케이션 및 프로젝트 CRUD 쿼리가 혼재되어 있음. `domain/application-lifecycle/repository` 계층 하위로 완전히 분할한다.
3.  **P0-3. `httpapi/applications` 파일 분할 (LoC 1066)**:
    *   *내용*: 애플리케이션 핸들러가 단일 파일에 밀집되어 있음. `domain/application-lifecycle/view` 계층 하위로 쪼갠다.
4.  **P0-4. `httpapi/organization` 및 `store/users_units` 분할 (LoC 1019 + 1263)**:
    *   *내용*: 조직 관리 부서 트리 제어가 복잡함. `domain/organization-management` 하위 `view` 및 `repository` 계층으로 역할을 명확히 쪼갠다.

### P1 — 1~2주 (레이어 격리 고도화)
1.  **P1-1. backend-core `httpapi` 패키지 해체**:
    *   *내용*: 비즈니스 로직과 HTTP 규격 처리가 혼재된 `httpapi`를 `handler` (View) 및 `service` (Service) 레이어로 명확하게 수직 격리한다.
2.  **P1-2. 외부 연동 어댑터 패턴 표준화**:
    *   *내용*: Homelab 외 Gitea, Keycloak 등 외부 연동 인프라 기술을 `infrastructure/` 하위 어댑터 구조로 완전히 격리하고, `domain`은 인터페이스만 소유하게 만든다.

---

## 4. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-28 | **초안 발행** (Claude Explore agent 종합 결과). 평면적인 12+4 카테고리 분류 수립. |
| 2026-05-28 | **고도화 개편** (Gemini). 평면형 카테고리를 **[Domain / Shared / Infrastructure]** 3대 레이어로 개편하고, 도메인 내부를 `service, repository, schema, view` 4대 계층으로 컴포넌트 세분화. PR #406의 Codex 인라인 리뷰 피드백을 수용하여 `ProviderModal`의 오너십 오분류 정정 완료. |
