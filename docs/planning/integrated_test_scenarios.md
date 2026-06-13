# 통합 테스트 시나리오 카탈로그

- 문서 목적: DevHub 전체 기능에 대한 End-to-End 통합 테스트 시나리오를 정의한다. Keycloak 인증/온보딩, Platform/Project/Repository 생명주기, Gitea SCM 연동, CI/CD 빌드 상태 연동을 단일 시나리오 체인으로 연결한다.
- 범위: 시나리오 1 (Keycloak 사용자 등록/로그인/온보딩/권한 조회 범위) / 시나리오 2 (시스템 관리자 Platform/Project/Repository 등록 + Gitea 연결) / 시나리오 3 (Gitea PR/Issue 연동 및 알림) / 시나리오 4 (빌드 실패 및 대시보드 표시). 각 시나리오는 세부 서브 시나리오로 구성.
- 대상 독자: QA, Backend/Frontend 개발자, AI 에이전트, 시스템 운영자.
- 상태: draft
- 최종 수정일: 2026-06-01
- 관련 문서: [`system_usecases.md`](./system_usecases.md), [`e2e_testing_strategy.md`](../tests/e2e_testing_strategy.md), [`release_v0-1_roadmap.md`](./release_v0-1_roadmap.md), [`docs/domain/auth-session/`](../domain/auth-session/), [`docs/domain/onboarding/`](../domain/onboarding/), [`docs/domain/platform-lifecycle/`](../domain/platform-lifecycle/), [`docs/domain/repository-integration/`](../domain/repository-integration/), [`docs/domain/integration-registry/`](../domain/integration-registry/), [`docs/domain/realtime/`](../domain/realtime/), [`docs/traceability/report.md`](../traceability/report.md).

## 1. 테스트 환경

### 1.1 서비스 구성 (colima docker-compose)

| 서비스 | 엔드포인트 | 인증 |
|--------|-----------|------|
| DevHub Frontend | `http://localhost:3000/` | Keycloak OIDC (PKCE) |
| DevHub Backend API | `http://localhost:8080/` | Bearer Token (JWT) |
| DevHub AI API | `http://localhost:8000/` | (내부) |
| Keycloak IdP | `http://localhost:8180/devhub/auth/keycloak/` | admin/admin |
| Keycloak Realm | `devhub` (`/devhub/auth/keycloak/realms/devhub`) | — |
| PostgreSQL | `localhost:5433` / `db:5432` | postgres / your_password |

### 1.2 사전 준비 계정

| 계정 | 역할 | Keycloak 사용자명 | 비고 |
|------|------|-------------------|------|
| System Admin | `system_admin` | `admin` (이미 존재) | 최고 관리자 |
| Developer A | `developer` | 시나리오 1에서 생성 | 온보딩 대상 |
| Manager B | `manager` | 시나리오 1에서 생성 | 온보딩 대상 |

### 1.3 사전 준비 데이터

- `infra/idp/keycloak-realm.dev.json` 기준 Keycloak realm import 완료
- Gitea 서버 접속 가능 (`http://homelab.ddn777.synology.me/gitea`, `yklee` / `yklee12!`)
- DB 마이그레이션 46건 모두 적용 완료

---

## 2. 시나리오 1: Keycloak 사용자 첫 등록 및 첫 로그인 온보딩 + 권한별 조회 범위

### 2.1 개요

Keycloak IdP를 통한 신규 사용자 등록 → 첫 로그인 → 온보딩 게이트 통과 → 역할(Role)별 조회 권한 범위를 확인한다.

### 2.2 기능 맵

| 기능 ID | 설명 | 관련 REQ | 관련 UC |
|---------|------|----------|---------|
| F-TEST-AUTH-LOGIN | Keycloak OIDC code flow 로그인 | REQ-FR-19, 21..24 | UC-AUTH-01 |
| F-TEST-AUTH-ME | `GET /me` 토큰 기반 본인 정보 조회 | REQ-FR-61..67 | UC-AUTH-02 |
| F-TEST-ONBOARD-GATE | 온보딩 게이트 차단/통과 | REQ-FR-ONBOARD-009, 010 | UC-ONBOARD-09, 10, 11 |
| F-TEST-ONBOARD-SUBMIT | 온보딩 제출 (display_name + 조직 선택) | REQ-FR-ONBOARD-001..003 | UC-ONBOARD-01, 02 |
| F-TEST-ONBOARD-REVIEW | 관리자 온보딩 검토 승인 | REQ-FR-ONBOARD-005 | UC-ONBOARD-05 |
| F-TEST-RBAC-SCOPE | 역할별 리소스 조회 범위 차이 | REQ-FR-27, 86 | UC-RBAC-01, 03 |
| F-TEST-RBAC-ROUTE | 라우트 권한 강제 (403) | NFR-26 | UC-RBAC-03 |

### 2.3 상세 시나리오

#### SC-TEST-1.1: 신규 사용자 등록 및 첫 로그인

| 항목 | 내용 |
|------|------|
| **TC ID** | `TC-TEST-AUTH-01` |
| **우선순위** | P0 |
| **계층** | E2E |
| **전제조건** | 1) DevHub 서비스 전체 기동 2) Keycloak admin 계정 준비 3) Gitea 외부 접속 가능 |
| **시나리오** | 1. Keycloak Admin Console 접속 (`http://localhost:8180/devhub/auth/keycloak/admin/`) — admin/admin 로그인 |
| | 2. `devhub` Realm → Users → Add user: `dev-user-a`, Email `dev-a@example.com`, First/Last name 입력 |
| | 3. Credentials 탭에서 초기 비밀번호 설정 (temporary: off) |
| | 4. `devhub` Realm → Clients → `devhub-frontend` → Client scopes → "devhub-roles" 매핑 확인 |
| | 5. 신규 사용자에게 role `developer` 할당 (또는 default role 확인) |
| | — |
| | 6. **브라우저 시크릿 창**에서 `http://localhost:3000/` 접속 → Keycloak login 화면으로 redirect 확인 |
| | 7. `dev-user-a` / 설정한 비밀번호로 로그인 |
| | 8. OIDC code flow 완료 → PKCE 토큰 교환 → DevHub callback `/devhub/auth/callback` 정상 처리 |
| | 9. `GET /api/v0-1/me` 응답 확인 (토큰 claims 기반 `login`, `subject`, `role: developer`, `email`, `display_name` 초기값 null) |
| **기대 결과** | - Keycloak 사용자 등록 후 OIDC 로그인 플로우 정상 동작 |
| | - 로그인 성공 시 DevHub 세션/JWT 발급 |
| | - `GET /me` 가 OIDC claims 기반 사용자 정보 반환 |
| | - `onboarding_required: true` (온보딩 미완료 상태) |
| | - Email 중복 등록 시 오류 처리 |
| **spec ts 위치** | `frontend/tests/e2e/auth.spec.ts` (확장) |

#### SC-TEST-1.2: 온보딩 게이트 및 온보딩 제출

| 항목 | 내용 |
|------|------|
| **TC ID** | `TC-TEST-ONBOARD-01` |
| **우선순위** | P0 |
| **계층** | E2E |
| **전제조건** | SC-TEST-1.1 완료 (dev-user-a 로그인 상태) |
| **시나리오** | 1. 로그인 직후 `/devhub/` 접속 시 `/devhub/onboarding`으로 자동 redirect 확인 (onboarding_required=true) |
| | 2. 온보딩 페이지에서 "Skip" 버튼 노출 확인 → skip 시 sessionStorage skip flag 설정 → dashboard 진입 |
| | 3. 상단 dismissible banner ("온보딩을 완료해주세요") 노출 확인 |
| | 4. 보호 경로 (예: `/devhub/account`) 접근 시 `/devhub/onboarding` hard redirect 확인 |
| | — |
| | 5. 온보딩 페이지에서 display_name 입력 ("Dev User A") |
| | 6. 조직 검색 (typeahead, `q >= 2`): "Dev" 입력 → 조직 목록 드롭다운 |
| | 7. 조직 선택 후 "제출" 버튼 클릭 → `POST /api/v0-1/me/onboarding` |
| | 8. 응답 확인: 201 + `onboarding_completed_at=NOW()` + `review_status=pending_review` |
| | 9. 온보딩 이후 dashboard 정상 진입 확인 (`onboarding_required: false`) |
| **기대 결과** | - 온보딩 게이트가 403/redirect로 미완료 사용자 제어 |
| | - 온보딩 제출 시 users row INSERT/UPDATE + audit event 발생 |
| | - Skip-and-resume UX 정상 동작 (sessionStorage + banner + hard redirect) |
| | - 제출 완료 후 review_status = pending_review |
| **spec ts 위치** | `frontend/tests/e2e/onboarding-first-login.spec.ts` |

#### SC-TEST-1.3: 관리자 온보딩 검토 승인

| 항목 | 내용 |
|------|------|
| **TC ID** | `TC-TEST-ONBOARD-02` |
| **우선순위** | P0 |
| **계층** | E2E |
| **전제조건** | SC-TEST-1.2 완료 (dev-user-a review_status=pending_review) |
| **시나리오** | 1. system_admin 계정으로 로그인 |
| | 2. `/devhub/admin/settings/users` → "검토 대기" 패널에 dev-user-a 노출 확인 |
| | 3. "확정" 버튼 클릭 → ConfirmReviewModal → 확정 |
| | 4. `POST /api/v0-1/admin/users/:user_id/review` → 200 + `review_status=reviewed` + `reviewed_at=NOW()` |
| | 5. audit log `account.review_confirmed` 발행 확인 |
| | 6. 검토 완료 후 dev-user-a 계정으로 재로그인 → 모든 메뉴 정상 접근 |
| **기대 결과** | - 관리자 온보딩 검토 workflow 정상 동작 |
| | - 검토 완료 시 audit event 기록 |
| | - 검토 완료 사용자는 정상 서비스 이용 가능 |
| **spec ts 위치** | `frontend/tests/e2e/admin-users-crud.spec.ts` (확장) |

#### SC-TEST-1.4: 역할별 권한 조회 범위 (RBAC Scope Matrix)

| 항목 | 내용 |
|------|------|
| **TC ID** | `TC-TEST-RBAC-01` |
| **우선순위** | P0 |
| **계층** | API / UI |
| **전제조건** | 3개 계정 준비: Developer A(`developer`), Manager B(`manager`), System Admin(`system_admin`) |
| **시나리오** | **Developer 조회 범위** |
| | 1. `developer` 계정으로 `GET /api/v0-1/platforms` → 자신이 속한 org/project의 application만 조회 |
| | 2. `GET /api/v0-1/projects` → 동일 scope 제한 확인 |
| | 3. `GET /api/v0-1/admin/...` → 403 확인 (system_admin 전용) |
| | — |
| | **Manager 조회 범위** |
| | 4. `manager` 계정으로 `GET /api/v0-1/platforms` → 자신의 부서/프로젝트 전체 조회 가능 |
| | 5. `GET /api/v0-1/projects/:id/members` → 멤버 조회 가능 |
| | 6. `POST /api/v0-1/dev-requests` → dev-request 생성 가능 |
| | — |
| | **System Admin 조회 범위** |
| | 7. `system_admin` 계정으로 `GET /api/v0-1/admin/settings/users` → 전체 사용자 조회 가능 |
| | 8. `GET /api/v0-1/rbac/policies` → RBAC 정책 조회/편집 가능 |
| | 9. `GET /api/v0-1/integration/providers` → 전체 provider 조회 가능 |
| | 10. 시스템 대시보드 (`/system`) 접근 가능 |
| **기대 결과** | - Row-scoping (ADR-0011) 에 따라 각 role의 데이터 조회 범위가 정확히 구분됨 |
| | - developer → 자신 관련 데이터만 |
| | - manager → 부서/프로젝트 전체 |
| | - system_admin → 전체 시스템 |
| | - 권한 없는 API는 403 반환 |
| **spec ts 위치** | `frontend/tests/e2e/rbac-routes.spec.ts` (확장) |

---

## 3. 시나리오 2: 시스템 관리자 Platform/Project/Repository 등록

### 3.1 개요

시스템 관리자가 DevHub를 통해 Application → Project → Repository를 등록하고, 등록된 Repository가 Gitea에 실제 저장소로 생성되는 과정을 확인한다. Gitea 연결이 없는 경우의 연결 설정도 포함한다.

### 3.2 기능 맵

| 기능 ID | 설명 | 관련 REQ | 관련 UC |
|---------|------|----------|---------|
| F-TEST-INT-PROVIDER | Integration Provider 등록 (Gitea SCM) | REQ-FR-INT-001..005 | UC-INT-01..05 |
| F-TEST-INT-BINDING | Integration Binding 연결 | REQ-FR-INT-006..010 | UC-INT-06..10 |
| F-TEST-APP-CREATE | Application 생성 | REQ-FR-APP-001..005 | UC-APP-01..03 |
| F-TEST-PROJ-CREATE | Project 생성 | REQ-FR-PROJ-001..005 | UC-PROJ-01..04 |
| F-TEST-REPO-CREATE | Repository 등록 (Outbound Create) | REQ-FR-REPO-001..003 | UC-REPO-04, 05 |
| F-TEST-GITEA-CONNECT | Gitea SCM 연결 설정 | REQ-FR-INT-011..015 | UC-INT-15..18 |

### 3.3 상세 시나리오

#### SC-TEST-2.1: Gitea Integration Provider 등록

| 항목 | 내용 |
|------|------|
| **TC ID** | `TC-TEST-GITEA-PROVIDER-01` |
| **우선순위** | P0 |
| **계층** | E2E |
| **전제조건** | 1) Gitea 서버 접속 가능 확인 (`http://homelab.ddn777.synology.me/gitea`) 2) system_admin 계정 준비 3) Gitea 인증 정보: `yklee` / `yklee12!` |
| **시나리오** | **Gitea Provider 등록 (연결 없음 상태에서 시작)** |
| | 1. system_admin 로그인 → `/devhub/admin/settings/integrations` → "Add Provider" |
| | 2. Provider 유형: `gitea` (SCM) 선택 |
| | 3. Provider 정보 입력: |
| |    - Name: "HomeLab Gitea" |
| |    - Base URL: `http://homelab.ddn777.synology.me/gitea` |
| |    - Auth Mode: `basic` 또는 `token` |
| |    - Credentials: `yklee` / `yklee12!` |
| | 4. "Test Connection" 버튼 클릭 → `POST /api/v0-1/integration/test-connection` → 200 OK 확인 |
| | 5. 저장 → `POST /api/v0-1/integration/providers` → 201 + provider_id 반환 |
| | — |
| | **Provider 목록 확인** |
| | 6. `GET /api/v0-1/integration/providers` → 등록된 Gitea provider 조회 |
| | 7. provider 상세: base_url, auth_mode, capabilities(scm), status=active 확인 |
| **기대 결과** | - Gitea SCM Provider 등록/연결 테스트 정상 동작 |
| | - Test Connection이 base_url reachability + credential 유효성 검증 |
| | - Provider 등록 후 목록에 표시, capabilities에 scm 포함 |
| **spec ts 위치** | `frontend/tests/e2e/admin-integrations.spec.ts` (확장) |

#### SC-TEST-2.2: Application 생성

| 항목 | 내용 |
|------|------|
| **TC ID** | `TC-TEST-APP-01` |
| **우선순위** | P0 |
| **계층** | E2E |
| **전제조건** | SC-TEST-2.1 완료 (Gitea Provider 등록됨) + 사전 조직 데이터 존재 |
| **시나리오** | 1. system_admin 로그인 → `/devhub/applications` → "Create Application" |
| | 2. Application 정보 입력: |
| |    - Name: "Test Application Alpha" |
| |    - Key: `test-app-alpha` |
| |    - Leader: 조직 사용자 선택 |
| |    - Dev Unit: 조직 단위 선택 |
| |    - Visibility: `public` (기본) |
| | 3. 제출 → `POST /api/v0-1/platforms` → 201 + platform_id 반환 |
| | 4. `GET /api/v0-1/platforms` 목록에서 "Test Application Alpha" 노출 확인 |
| | 5. 상세 진입 시 입력한 정보 일치 확인 |
| **기대 결과** | - Application 생성 API 정상 동작 (API-41~43) |
| | - Key 중복 시 409 에러 |
| | - Leader/Dev Unit은 org unit과 FK 정합성 유지 |
| **spec ts 위치** | `frontend/tests/e2e/admin-applications.spec.ts` |

#### SC-TEST-2.3: Project 생성 (Application 내)

| 항목 | 내용 |
|------|------|
| **TC ID** | `TC-TEST-PROJ-01` |
| **우선순위** | P0 |
| **계층** | E2E |
| **전제조건** | SC-TEST-2.2 완료 (Test Application Alpha 존재) |
| **시나리오** | 1. Platform 상세 페이지 → "Add Project" |
| | 2. Project 정보 입력: |
| |    - Name: "Alpha Integration Sprint" |
| |    - Status: `active` |
| |    - 기간: 2026-06-01 ~ 2026-07-31 |
| | 3. 제출 → `POST /api/v0-1/projects` → 201 + project_id 반환 |
| | 4. `GET /api/v0-1/platforms/:id/projects` 목록에서 project 노출 확인 |
| | 5. Project 상세 진입: 기간, 상태, 담당자 정보 일치 확인 |
| **기대 결과** | - Project 생성 API 정상 동작 (API-55~56) |
| | - Application-Project 관계 정합성 유지 |
| **spec ts 위치** | `frontend/tests/e2e/admin-projects.spec.ts` |

#### SC-TEST-2.4: Gitea Repository Outbound 생성

| 항목 | 내용 |
|------|------|
| **TC ID** | `TC-TEST-REPO-CREATE-01` |
| **우선순위** | P0 |
| **계층** | E2E |
| **전제조건** | SC-TEST-2.1 (Gitea Provider) + SC-TEST-2.3 (Project) 완료 |
| **시나리오** | **Repository 생성 (Outbound Create → Gitea)** |
| | 1. Project 상세 페이지 → "Create Repository" |
| | 2. Repository 정보 입력: |
| |    - Repository Name: `alpha-integration-repo` |
| |    - Provider: "HomeLab Gitea" (SC-TEST-2.1에서 등록) |
| |    - Visibility: `private` |
| | 3. 제출 → `POST /api/v0-1/integration/providers/:id/create-repository` → 201 |
| | 4. 응답에 Gitea repository URL 포함 확인 |
| | — |
| | **Gitea 저장소 생성 확인** |
| | 5. Gitea 웹 UI 접속 (`http://homelab.ddn777.synology.me/gitea/yklee/alpha-integration-repo`) |
| | 6. 저장소가 정상 생성되었는지 확인 |
| | — |
| | **DevHub Binding 연결** |
| | 7. Project → "Link Repository" → 생성된 repository 선택 |
| | 8. `POST /api/v0-1/integration/bindings` 로 Project-Repository 바인딩 |
| | 9. `GET /api/v0-1/platforms/:id/dashboard` (API-93) 에 repository 연동 상태 노출 확인 |
| **기대 결과** | - DevHub에서 Gitea 저장소 Outbound 생성 정상 동작 |
| | - Gitea에 실제 저장소가 생성됨 |
| | - Project-Repository 바인딩 정상 연결 |
| | - Application dashboard에 repository 정보 표시 |
| **spec ts 위치** | `frontend/tests/e2e/repositories-publish.spec.ts` (확장) |

#### SC-TEST-2.5: Repository Draft → Publish 라이프사이클

| 항목 | 내용 |
|------|------|
| **TC ID** | `TC-TEST-REPO-LIFECYCLE-01` |
| **우선순위** | P1 |
| **계층** | API / UI |
| **전제조건** | SC-TEST-2.4 완료 |
| **시나리오** | 1. `POST /api/v0-1/repositories` (API-91) 로 draft 상태 repository 생성 |
| | 2. draft 상태에서 목록 조회 시 `status=draft` 표시 |
| | 3. `POST /api/v0-1/repositories/:id/publish` (API-92) 로 publish |
| | 4. publish 후 목록에 `status=active`로 노출 |
| | 5. Gitea에 실제 저장소 생성 확인 |
| **기대 결과** | - Repository draft → publish lifecycle 정상 동작 |
| | - Publish 전까지는 목록에만 노출, 실제 Gitea 저장소 미생성 |
| **spec ts 위치** | `frontend/tests/e2e/repositories-publish.spec.ts` |

---

## 4. 시나리오 3: Gitea PR/Issue 연동 및 알림

### 4.1 개요

Gitea에 샘플 작업물을 등록하고, PR/Issue가 DevHub에 연동되어 표시되는지 확인한다. Issue 상태 변경과 담당자 지정을 통해 알림이 정상 동작하는지 검증한다.

### 4.2 기능 맵

| 기능 ID | 설명 | 관련 REQ | 관련 UC |
|---------|------|----------|---------|
| F-TEST-GITEA-ISSUE | Gitea Issue 등록 → DevHub 연동 | REQ-FR-49..55 | UC-GITEA-01, UC-REPO-01 |
| F-TEST-GITEA-PR | Gitea PR 생성 → Webhook 수신 → DevHub 표시 | REQ-FR-49..55 | UC-GITEA-02 |
| F-TEST-ISSUE-STATUS | Issue 상태 변경 연동 (open/close/reopen) | REQ-FR-49..55 | UC-GITEA-03 |
| F-TEST-ISSUE-ASSIGN | Issue 담당자 지정 → 알림 | REQ-FR-104, 105 | UC-RT-01, 02 |
| F-TEST-WEBHOOK | Gitea Webhook 수신 | REQ-FR-56 | UC-GITEA-03 |

### 4.3 상세 시나리오

#### SC-TEST-3.1: Gitea Issue 등록 및 연동

| 항목 | 내용 |
|------|------|
| **TC ID** | `TC-TEST-ISSUE-01` |
| **우선순위** | P0 |
| **계층** | E2E |
| **전제조건** | 1) SC-TEST-2.4 완료 (alpha-integration-repo 존재) 2) Gitea Webhook이 DevHub로 전송되도록 설정 |
| **시나리오** | **Gitea에 Issue 등록** |
| | 1. Gitea 웹 UI 접속 → `alpha-integration-repo` → Issues → "New Issue" |
| | 2. Issue 작성: |
| |    - Title: "로그인 페이지 에러 수정" |
| |    - Content: "로그인 시 특정 상황에서 500 에러 발생. 원인 파악 필요." |
| |    - Assignee: `yklee` |
| |    - Label: `bug` |
| | 3. Create Issue → Gitea Webhook이 DevHub로 전송 |
| | — |
| | **DevHub Issue 표시 확인** |
| | 4. DevHub Repository 상세 페이지 진입 (`/devhub/repositories/:id`) |
| | 5. "Issues" 섹션에 등록된 Issue 표시 확인 |
| | 6. Issue 상세: title, content, author, status=`open`, label 일치 확인 |
| **기대 결과** | - Gitea Issue 생성 → Webhook → DevHub 연동 정상 |
| | - Repository 상세 페이지에 Issue 목록 노출 |
| | - Issue 메타데이터(title, status, assignee, label) 정확히 표시 |
| **spec ts 위치** | `frontend/tests/e2e/repositories-ui.spec.ts` (확장) |

#### SC-TEST-3.2: Issue 상태 변경 연동 (Open/Close/Reopen)

| 항목 | 내용 |
|------|------|
| **TC ID** | `TC-TEST-ISSUE-STATUS-01` |
| **우선순위** | P0 |
| **계층** | E2E |
| **전제조건** | SC-TEST-3.1 완료 (open 상태 Issue 존재) |
| **시나리오** | **Issue Close** |
| | 1. Gitea → Issue → Close Issue |
| | 2. DevHub Webhook 수신 확인 (`POST /api/v0-1/integrations/gitea/webhooks`) |
| | 3. DevHub Repository 페이지 Issue 상태가 `closed`로 변경 확인 |
| | — |
| | **Issue Reopen** |
| | 4. Gitea → Closed Issue → Reopen |
| | 5. DevHub Issue 상태가 다시 `open`으로 변경 확인 |
| | — |
| | **Issue 상태 집계** |
| | 6. Repository 상세 → "Issues" 섹션 open/closed count 확인 |
| | 7. Application Dashboard의 Issue 상태 요약과 일치 확인 |
| **기대 결과** | - Issue 상태 변경 시 Webhook → DevHub 실시간 연동 |
| | - Open/Close/Reopen 상태 전이 정확히 반영 |
| | - 상태 집계(open/closed count)가 dashboard와 일치 |
| **spec ts 위치** | 신규 E2E (repositories-integration.spec.ts) |

#### SC-TEST-3.3: Gitea PR 생성 및 연동

| 항목 | 내용 |
|------|------|
| **TC ID** | `TC-TEST-PR-01` |
| **우선순위** | P0 |
| **계층** | E2E |
| **전제조건** | SC-TEST-2.4 완료. Gitea 저장소에 샘플 코드 커밋 존재 |
| **시나리오** | **Gitea PR 생성** |
| | 1. Gitea `alpha-integration-repo` → branch 생성 (`fix/login-error`) |
| | 2. 샘플 코드 변경 커밋 (README 수정 등) |
| | 3. New Pull Request 생성: |
| |    - Title: "[fix] 로그인 에러 수정" |
| |    - Base: `main` ← Compare: `fix/login-error` |
| |    - Reviewers: `yklee` |
| | 4. Create PR → Webhook → DevHub |
| | — |
| | **DevHub PR 연동 확인** |
| | 5. DevHub Repository 페이지 → "Pull Requests" 섹션 확인 |
| | 6. PR 정보: title, author, source/target branch, status=`open` 일치 확인 |
| | — |
| | **PR Merge** |
| | 7. Gitea에서 PR Merge (Merge Pull Request) |
| | 8. DevHub PR 상태 `merged`로 변경 확인 |
| | 9. PR Merge 이벤트가 contributors activity에 반영 확인 |
| **기대 결과** | - PR 생성/병합 시 Webhook 연동 정상 |
| | - PR 상세 정보(title, branch, status) 정확히 표시 |
| | - Merge 후 PR 상태 `merged`로 변경 및 contributors 활동 기록 |
| **spec ts 위치** | 신규 E2E (repositories-integration.spec.ts) |

#### SC-TEST-3.4: 담당자 지정 및 알림

| 항목 | 내용 |
|------|------|
| **TC ID** | `TC-TEST-NOTIF-01` |
| **우선순위** | P0 |
| **계층** | E2E |
| **전제조건** | SC-TEST-3.2 완료. DevHub에 2명 이상 사용자 등록 (담당자/알림 수신자) |
| **시나리오** | **Issue 담당자 지정** |
| | 1. Gitea Issue → Assignee를 `dev-user-a`로 지정 |
| | 2. Webhook → DevHub 이벤트 처리 |
| | — |
| | **DevHub 알림 확인** |
| | 3. `dev-user-a` 계정으로 로그인 |
| | 4. 대시보드 상단 알림 영역에서 Issue 할당 알림 확인 |
| | 5. 알림 클릭 → 해당 Issue 상세로 이동 |
| | — |
| | **Realtime WebSocket 이벤트** |
| | 6. WebSocket 연결 (`GET /api/v0-1/realtime/ws`) 후 `notification.created` 이벤트 수신 대기 |
| | 7. Gitea에서 Issue comment 추가 → `notification.created` 이벤트 발생 확인 |
| | 8. 이벤트 payload 내 `type`, `event_id`, `occurred_at`, `data` 필드 검증 |
| | — |
| | **담당자별 필터링** |
| | 9. `/devhub/issues?assignee=dev-user-a` 로 담당자별 Issue 조회 |
| | 10. 정확히 해당 사용자에게 할당된 Issue만 표시 확인 |
| **기대 결과** | - Issue 담당자 지정 시 DevHub 알림 생성 |
| | - WebSocket `notification.created` 이벤트 실시간 수신 |
| | - 담당자별 Issue 필터링 정상 동작 |
| | - 알림 클릭 시 해당 Issue 상세로 이동 |
| **spec ts 위치** | 신규 E2E (notifications.spec.ts) |

#### SC-TEST-3.5: Gitea Webhook 이벤트 집계 (Audit)

| 항목 | 내용 |
|------|------|
| **TC ID** | `TC-TEST-WEBHOOK-01` |
| **우선순위** | P1 |
| **계층** | API |
| **전제조건** | SC-TEST-3.1~3.3 완료 (다수의 Webhook 이벤트 발생) |
| **시나리오** | 1. `POST /api/v0-1/integrations/gitea/webhooks` 의 signature 검증 로그 확인 |
| | 2. `GET /api/v0-1/webhook-events` 로 수신된 webhook 이벤트 목록 조회 |
| | 3. 이벤트 상태별 집계: `validated`, `processed`, `failed`, `ignored` 구분 확인 |
| | 4. 중복 이벤트(`X-Gitea-Delivery` 기준)가 `duplicate` 처리되는지 확인 |
| | 5. Webhook 수신 실패 시 (signature mismatch 등) 적절한 에러 응답 확인 |
| **기대 결과** | - Gitea Webhook 수신 → signature 검증 → 이벤트 저장 lifecycle 정상 |
| | - Webhook 이벤트 목록/상태 조회 가능 (API-04 등) |
| | - 중복 이벤트 idempotent 처리 |
| | - 위변조 signature는 실패 처리 |
| **spec ts 위치** | 신규 (webhook-events.spec.ts) |

---

## 5. 시나리오 4: 빌드 실패 케이스 등록 및 시스템 내 표시

### 5.1 개요

CI/CD 빌드 실패 상황을 시뮬레이션하고, DevHub 내 Platform/Project/Repository 대시보드에 빌드 상태가 정확히 표시되는지 확인한다. 다양한 실패 유형과 상태 전이를 검증한다.

### 5.2 기능 맵

| 기능 ID | 설명 | 관련 REQ | 관련 UC |
|---------|------|----------|---------|
| F-TEST-CI-RUN | CI Run 상태 등록/조회 | REQ-FR-104 | UC-RT-01 |
| F-TEST-CI-STATUS | 빌드 상태 전이 (pending→running→success/failed) | REQ-FR-104 | UC-RT-01, 02 |
| F-TEST-CI-DASHBOARD | 대시보드 빌드 상태 표시 | REQ-FR-APPDASH-001..003 | UC-APPDASH-01..04 |
| F-TEST-CI-FAILURE-TYPE | 다양한 빌드 실패 유형 처리 | REQ-FR-APPDASH-004..006 | UC-APPDASH-05..07 |
| F-TEST-CI-REALTIME | 빌드 상태 실시간 업데이트 (WebSocket) | REQ-FR-104, 105 | UC-RT-02 |

### 5.3 상세 시나리오

#### SC-TEST-4.1: CI Run 정상/실패 상태 등록

| 항목 | 내용 |
|------|------|
| **TC ID** | `TC-TEST-CI-01` |
| **우선순위** | P0 |
| **계층** | API |
| **전제조건** | SC-TEST-2.4 완료 (alpha-integration-repo 존재) |
| **시나리오** | **CI Run 데이터 주입 (API)** |
| | 1. 정상 빌드 데이터 생성 (Gitea Actions 또는 API 직접): |
| |    - Repository: `alpha-integration-repo` |
| |    - Branch: `main` |
| |    - Status: `success` |
| |    - Duration: 120초 |
| |    - Commit SHA: (실제 커밋) |
| | 2. `GET /api/v0-1/ci-runs` → 정상 빌드 목록 조회 |
| | 3. CI Run 상세: id, duration_seconds, status, html_url 필드 확인 |
| | — |
| | **빌드 실패 케이스 주입** |
| | 4. 실패 빌드 데이터 생성: |
| |    - Status: `failure` |
| |    - Branch: `feature/new-feature` |
| |    - Duration: 30초 (조기 실패) |
| |    - 에러 메시지: "go build failed: cannot find package" |
| | 5. 목록에 success/failure 모두 노출 확인 |
| | 6. 실패 빌드에 실패 원인/메시지 표시 확인 |
| **기대 결과** | - CI Run 데이터 정상 조회 (GET /api/v0-1/ci-runs) |
| | - success/failure 상태 모두 정확히 표시 |
| | - 실패 시 원인 메시지 확인 가능 |
| **spec ts 위치** | `backend-core/internal/httpapi/domain_test.go` (참조) |

#### SC-TEST-4.2: 대시보드 빌드 상태 표시

| 항목 | 내용 |
|------|------|
| **TC ID** | `TC-TEST-CI-DASHBOARD-01` |
| **우선순위** | P0 |
| **계층** | UI |
| **전제조건** | SC-TEST-4.1 완료 (success + failure CI Run 데이터 존재) |
| **시나리오** | **Repository 대시보드** |
| | 1. `/devhub/repositories/:id` → "Build Runs" 섹션 확인 |
| | 2. 최근 빌드 목록 표시 (status, branch, commit, duration, timestamp) |
| | 3. 성공/실패 상태별 배지/색상 구분 확인 |
| | 4. "Last Build" 상태 표시줄: `success` 또는 `failure` 텍스트와 색상 |
| | — |
| | **Application 대시보드** |
| | 5. `/devhub/platforms/:id/dashboard` → Application 대시보드 확인 |
| | 6. 연관 Repository들의 빌드 상태 요약 표시 |
| | 7. 전체 성공률(%) 또는 최근 빌드 상태 집계 확인 |
| | — |
| | **Project 대시보드** |
| | 8. `/devhub/projects/:id` → Project 대시보드 확인 |
| | 9. Project 내 Repository 빌드 상태 통계 노출 확인 |
| | 10. "Build Success Rate" 또는 "Last Build" 위젯 확인 |
| **기대 결과** | - Repository 상세에 Build Runs 섹션 정상 표시 |
| | - Application Dashboard에 연관 repo 빌드 상태 집계 |
| | - Project 대시보드에 빌드 통계 정보 표시 |
| | - 상태별 시각적 구분 (색상/배지) 정상 적용 |
| **spec ts 위치** | `frontend/tests/e2e/repositories-ui.spec.ts` (확장) |

#### SC-TEST-4.3: 실시간 빌드 상태 업데이트 (WebSocket)

| 항목 | 내용 |
|------|------|
| **TC ID** | `TC-TEST-CI-REALTIME-01` |
| **우선순위** | P1 |
| **계층** | E2E |
| **전제조건** | WebSocket ticket 발급 가능 + Frontend RealtimeService 연결 |
| **시나리오** | **WebSocket 연결** |
| | 1. `POST /api/v0-1/realtime/ticket` → single-use ticket 발급 (60s TTL) |
| | 2. `GET /api/v0-1/realtime/ws?ticket={ticket}&types=ci.run.updated` 연결 |
| | 3. WebSocket handshake 성공 (101) |
| | — |
| | **CI Run 이벤트 수신** |
| | 4. Gitea Actions에서 새로운 빌드 시작 → `ci.run.updated` 이벤트 발생 |
| | 5. WebSocket 메시지 검증: |
| |    - `schema_version`, `type=ci.run.updated`, `event_id`, `occurred_at` |
| |    - `data` 내 CI Run 정보 (repository, branch, status) |
| | — |
| | **UI 자동 갱신** |
| | 6. WebSocket 이벤트 수신 후 Frontend Repository 페이지 Build Runs 섹션 자동 갱신 확인 |
| | 7. 빌드 상태 변경 (running → success/failure) 이벤트 순서대로 반영 확인 |
| **기대 결과** | - WebSocket ticket 기반 연결 정상 |
| | - `ci.run.updated` 이벤트 실시간 수신 |
| | - UI가 WebSocket 이벤트에 따라 자동 갱신 |
| | - 빌드 상태 전이 이벤트 순차 반영 |
| **spec ts 위치** | `frontend/tests/e2e/realtime.spec.ts` (신규) |

#### SC-TEST-4.4: 다양한 빌드 실패 유형 처리

| 항목 | 내용 |
|------|------|
| **TC ID** | `TC-TEST-CI-FAILURE-01` |
| **우선순위** | P1 |
| **계층** | API / UI |
| **전제조건** | SC-TEST-4.1 완료 |
| **시나리오** | **컴파일 에러** |
| | 1. CI Run status=`failure`, error="compilation error: undefined variable" 등록 |
| | 2. DevHub에서 실패 유형 + 에러 메시지 표시 확인 |
| | — |
| | **테스트 실패** |
| | 3. CI Run status=`failure`, error="test failed: TestLoginHandler (2 failed)" 등록 |
| | 4. 실패 테스트 개수/이름 표시 확인 |
| | — |
| | **시간 초과** |
| | 5. CI Run status=`failure`, error="timed out after 30m" 등록 |
| | 6. timeout 메시지 표시 확인 |
| | — |
| | **취소된 빌드** |
| | 7. CI Run status=`cancelled` 등록 |
| | 8. "Cancelled" 상태 표시 (실패와 구분) 확인 |
| | — |
| | **불확정/불안정** |
| | 9. CI Run status=`unstable` 등록 (test 일부 실패 but build 성공 등) |
| | 10. "Unstable" 상태 별도 표시 확인 |
| **기대 결과** | - 다양한 빌드 실패 유형별 상태 구분 표시 |
| | - 각 유형에 맞는 에러 메시지/원인 표시 |
| | - 취소(cancelled)와 실패(failure) 시각적 구분 |
| | - 불안정(unstable) 상태 별도 처리 |
| **spec ts 위치** | `frontend/tests/e2e/repositories-detail-negative.spec.ts` (확장) |

#### SC-TEST-4.5: 빌드 상태 집계 및 트렌드

| 항목 | 내용 |
|------|------|
| **TC ID** | `TC-TEST-CI-TREND-01` |
| **우선순위** | P2 |
| **계층** | API / UI |
| **전제조건** | 10개 이상의 CI Run 데이터 존재 (success/failure 혼합) |
| **시나리오** | 1. 특정 기간(예: 최근 7일)의 빌드 성공률 계산 확인 |
| | 2. Application Dashboard의 "Build Success Rate" 위젯 확인 |
| | 3. 분할 표시: (성공/실패/취소) 각각 count 또는 percentage |
| | 4. Branch별 빌드 상태 집계 확인 (main, feature branches 구분) |
| | 5. 시간별 빌드 트렌드 (날짜별 success/failure 분포) 표시 확인 |
| **기대 결과** | - 빌드 성공률 집계 정확 |
| | - Branch별/기간별 빌드 통계 제공 |
| | - 대시보드에 시각화 차트/위젯 표시 |
| **spec ts 위치** | 신규 (dashboard-statistics.spec.ts) |

---

## 6. 테스트 계층 전략

| 계층 | 대상 | 도구/프레임워크 | 적용 시나리오 |
|------|------|-----------------|--------------|
| **UT (단위)** | Handler/Middleware/Service 로직 | Go `testing`, Jest/Vitest | SC-TEST-1.4 (RBAC), SC-TEST-3.5 (Webhook), SC-TEST-4.4 (실패유형) |
| **IT (통합)** | DB + Handler 조합 | Go `testing` + testcontainers | SC-TEST-2.2~2.4 (CRUD 정합성), SC-TEST-4.1 (CI Run) |
| **E2E** | Browser 전체 흐름 | Playwright | SC-TEST-1.1~1.3 (온보딩), SC-TEST-2.1 (Provider), SC-TEST-3.1~3.4 (Issue/PR/알림) |
| **API** | HTTP endpoint 단독 | curl / REST Client | SC-TEST-2.5 (Draft→Publish), SC-TEST-4.1 (CI 상태) |

## 7. 우선순위 등급 정의

| 등급 | 의미 | 적용 |
|------|------|------|
| **P0** | 핵심 기능 / 회귀 시 main 안정성 위반 | 모든 SC-TEST-*.01 ~ *.02 |
| **P1** | 핵심 보조 기능 / negative case / monitoring | SC-TEST-*.03 이후 서브 시나리오 |
| **P2** | UX 보강 / 집계/트렌드 / 후속 확장 | SC-TEST-4.5 등 |

## 8. Cover 매트릭스 (TC ↔ REQ ↔ UC ↔ ARCH ↔ API)

| TC ID | REQ | UC | ARCH | API | 주요 IMPL 경로 |
|-------|-----|----|------|-----|---------------|
| TC-TEST-AUTH-01 | REQ-FR-19, 21..24 | UC-AUTH-01 | ARCH-11, 12, 14 | API-19, API-32 | auth-session/view/auth.go |
| TC-TEST-ONBOARD-01 | REQ-FR-ONBOARD-001..003, 008..011 | UC-ONBOARD-01, 02, 09, 10, 11 | ARCH-ONBOARD-02, 03, 06 | API-83, API-85 | onboarding/view/ |
| TC-TEST-ONBOARD-02 | REQ-FR-ONBOARD-005 | UC-ONBOARD-05 | ARCH-ONBOARD-02, 06 | API-86 | onboarding/view/handler.go |
| TC-TEST-RBAC-01 | REQ-FR-27, 86, NFR-26 | UC-RBAC-01, 03 | ARCH-13 | API-26~29 | rbac-permissions/view/ |
| TC-TEST-GITEA-PROVIDER-01 | REQ-FR-INT-001..015 | UC-INT-01..18 | ARCH-INT-01..07 | API-69~72, 87 | integration-registry/view/ |
| TC-TEST-APP-01 | REQ-FR-APP-001..005 | UC-APP-01..03 | ARCH-10 | API-41~43 | platform-lifecycle/view/ |
| TC-TEST-PROJ-01 | REQ-FR-PROJ-001..005 | UC-PROJ-01..04 | ARCH-10 | API-55~56 | platform-lifecycle/view/ |
| TC-TEST-REPO-CREATE-01 | REQ-FR-REPO-001..003 | UC-REPO-04, 05 | ARCH-REPO-04, 05 | API-88~92 | repository-integration/view/ |
| TC-TEST-REPO-LIFECYCLE-01 | REQ-FR-REPO-001..005 | UC-REPO-06, 07 | ARCH-REPO-06, 07 | API-91, 92 | repository-integration/ |
| TC-TEST-ISSUE-01 | REQ-FR-49..55 | UC-GITEA-01, UC-REPO-01 | ARCH-06, 07 | API-02, API-51 | gitea/client.go |
| TC-TEST-ISSUE-STATUS-01 | REQ-FR-49..55 | UC-GITEA-03 | ARCH-06 | API-02 | gitea/client.go |
| TC-TEST-PR-01 | REQ-FR-49..55 | UC-GITEA-02 | ARCH-06, 07 | API-02 | gitea/client.go |
| TC-TEST-NOTIF-01 | REQ-FR-104, 105 | UC-RT-01, 02 | ARCH-05 | API-37 | realtime/view/realtime.go |
| TC-TEST-WEBHOOK-01 | REQ-FR-56 | UC-GITEA-03 | ARCH-06 | API-04 | integration-registry/view/ |
| TC-TEST-CI-01 | REQ-FR-104 | UC-RT-01 | ARCH-05 | (ci-runs) | httpapi/router.go, domain.go |
| TC-TEST-CI-DASHBOARD-01 | REQ-FR-APPDASH-001..003 | UC-APPDASH-01..04 | ARCH-APPDASH-01, 02 | API-93 | platform-lifecycle/view/ |
| TC-TEST-CI-REALTIME-01 | REQ-FR-104, 105 | UC-RT-02 | ARCH-05 | API-37 | realtime/view/realtime.go |
| TC-TEST-CI-FAILURE-01 | REQ-FR-APPDASH-004..006 | UC-APPDASH-05..07 | ARCH-APPDASH-03 | (ci-runs) | domain.go (CIRun) |
| TC-TEST-CI-TREND-01 | REQ-FR-APPDASH-001..006 | UC-APPDASH-01..07 | ARCH-APPDASH-04, 05 | API-93 | platform-lifecycle/ |

## 9. 테스트 실행 순서 (권장)

시나리오 간 의존관계를 고려한 권장 실행 순서:

```
[Phase 1] 인증/온보딩 기반
  SC-TEST-1.1 (신규 사용자 등록/로그인)
  SC-TEST-1.2 (온보딩 게이트/제출)
  SC-TEST-1.3 (관리자 온보딩 검토)
  SC-TEST-1.4 (RBAC 권한 조회 범위)

[Phase 2] 시스템 설정 및 데이터 등록
  SC-TEST-2.1 (Gitea Provider 등록)
  SC-TEST-2.2 (Application 생성)
  SC-TEST-2.3 (Project 생성)
  SC-TEST-2.4 (Repository Outbound Create + Gitea 확인)
  SC-TEST-2.5 (Repository Draft → Publish)

[Phase 3] SCM 연동 및 실시간 통신
  SC-TEST-3.1 (Gitea Issue 등록)
  SC-TEST-3.2 (Issue 상태 변경 연동)
  SC-TEST-3.3 (Gitea PR 생성)
  SC-TEST-3.4 (담당자 지정 및 알림)
  SC-TEST-3.5 (Webhook 집계)

[Phase 4] 빌드/CI 연동
  SC-TEST-4.1 (CI Run 상태 등록)
  SC-TEST-4.2 (대시보드 빌드 상태 표시)
  SC-TEST-4.3 (실시간 빌드 상태 업데이트)
  SC-TEST-4.4 (다양한 빌드 실패 유형)
  SC-TEST-4.5 (빌드 상태 집계 및 트렌드)
```

## 10. 알려진 제약 및 리스크

| 리스크 | 영향 | 대응 |
|--------|------|------|
| Gitea 외부 서비스 의존성 | SC-TEST-2.1, 3.1~3.5 차단 | Gitea 중단 시 Mock SCM provider로 대체 테스트 |
| Keycloak Realm 사전 설정 필요 | SC-TEST-1.1 사전조건 | `infra/idp/keycloak-realm.dev.json` import 확인 |
| WebSocket 인증 ticket 60s TTL | SC-TEST-4.3 시간 제약 | ticket 재발급 로직 테스트에 포함 |
| Gitea Webhook tunnel 필요 | SC-TEST-3.1~3.5 외부 webhook 수신 | ngrok/smee.io 등 tunnel 도구 사용 |
| CI Run 데이터 Gitea Actions 의존 | SC-TEST-4.1, 4.4 | API 레벨 직접 주입으로 우회 가능 |

## 11. 변경 이력

- **2026-06-01** (sprint `deepseek/test-scenarios-20260601`): 본 카탈로그 신규. 5개 시나리오, 20개 TC 정의. Phase 1~4 실행 순서 포함.
