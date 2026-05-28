# Stage 2 Plan: Go Backend 모듈형 Handler 및 API 라우트 대정비 (Declarative Grouping)

본 계획서는 사용자의 훌륭한 인사이트("API 배치도 이러한 구조로 다시 검토")를 적극 반영하여, **API 라우팅 레이아웃 자체를 도메인 단위로 수직 그룹화(Modular Grouping)하고, 미들웨어를 그룹 단위로 선언적으로 바인딩**하는 백엔드 API 배치 전면 정비 계획을 추가한 Stage 2 상세 계획서입니다.

---

## 1. API 라우트 도메인 기반 배치 검토 (Declarative Grouping)

기존 `httpapi/router.go` 에 평평하게(flat) 나열되어 있던 API 라우트들을 10대 도메인별 그룹 구조로 개편하여, 경로 규격 자체가 아키텍처 SoT를 대변하도록 설계합니다.

```text
기존 (Flat Routes):
v1.GET("/applications", handler.listApplications)
v1.GET("/rbac/policies", handler.listRBACPolicies)
v1.GET("/audit-logs", handler.listAuditLogs)

개편 (Declarative Domain Grouping):
/api/v1
├── /auth-session   -> Auth, Session Death, Realtime Ticket
├── /audit-ops      -> Logs, Keycloak webhook
├── /rbac           -> Policies CRUD
├── /organization   -> Users/Units CRUD, Hierarchy tree
├── /applications   -> App Lifecycle, Repository bindings, Rollup
└── /dev-requests   -> DREQ Intake, promoting
```

### 1.1 도메인별 라우트 그룹 및 미들웨어 조율안

#### 1) Auth & Session Group (`/api/v1/auth`, `/api/v1/me`)
- **담당**: `auth-session` 및 `onboarding`
- **구조**:
  ```go
  auth := v1.Group("/auth")
  {
      auth.POST("/logout", authHandler.Logout)
      auth.POST("/ticket", authHandler.IssueRealtimeTicket)
  }
  me := v1.Group("/me")
  {
      me.GET("", authHandler.GetMe)
      me.PATCH("", orgHandler.PatchMe)
      me.POST("/onboarding", onboardHandler.SubmitOnboarding)
  }
  ```

#### 2) Audit Group (`/api/v1/audit`)
- **담당**: `audit-ops`
- **구조**:
  ```go
  audit := v1.Group("/audit")
  {
      audit.GET("/logs", auditHandler.ListAuditLogs)
      audit.POST("/keycloak-events", auditHandler.ReceiveKeycloakEventWebhook)
  }
  ```

#### 3) RBAC Group (`/api/v1/rbac`)
- **담당**: `rbac-permissions`
- **구조**:
  ```go
  rbac := v1.Group("/rbac")
  {
      rbac.GET("/policies", rbacHandler.ListRBACPolicies)
      rbac.POST("/policies", rbacHandler.CreateRBACPolicy)
      rbac.PUT("/policies", rbacHandler.UpdateRBACPolicies)
      rbac.DELETE("/policies/:role_id", rbacHandler.DeleteRBACPolicy)
  }
  ```

#### 4) Organization & Users Group (`/api/v1/organization`, `/api/v1/users`)
- **담당**: `organization-management`
- **구조**:
  ```go
  org := v1.Group("/organization")
  {
      org.GET("/hierarchy", orgHandler.GetHierarchy)
      org.PUT("/hierarchy", orgHandler.UpdateHierarchy)
      org.POST("/units", orgHandler.CreateOrgUnit)
      org.GET("/units/:unit_id", orgHandler.GetOrgUnit)
      org.PATCH("/units/:unit_id", orgHandler.UpdateOrgUnit)
      org.DELETE("/units/:unit_id", orgHandler.DeleteOrgUnit)
      org.GET("/units/:unit_id/members", orgHandler.ListUnitMembers)
      org.PUT("/units/:unit_id/members", orgHandler.ReplaceUnitMembers)
  }
  users := v1.Group("/users")
  {
      users.GET("", orgHandler.ListUsers)
      users.POST("", orgHandler.CreateUser)
      users.GET("/:user_id", orgHandler.GetUser)
      users.PATCH("/:user_id", orgHandler.UpdateUser)
      users.DELETE("/:user_id", orgHandler.DeleteUser)
      users.POST("/:user_id/review", onboardHandler.ConfirmUserReview)
  }
  ```

#### 5) Application & Project Group (`/api/v1/applications`, `/api/v1/projects`, `/api/v1/repositories`)
- **담당**: `application-lifecycle` 및 `repository-integration`
- **구조**:
  ```go
  apps := v1.Group("/applications")
  {
      apps.GET("", appHandler.ListApplications)
      apps.POST("", appHandler.CreateApplication)
      apps.GET("/:application_id", appHandler.GetApplication)
      apps.PATCH("/:application_id", appHandler.UpdateApplication)
      apps.DELETE("/:application_id", appHandler.ArchiveApplication)
      apps.GET("/:application_id/repositories", appHandler.ListApplicationRepositories)
      apps.POST("/:application_id/repositories", appHandler.CreateApplicationRepository)
      apps.DELETE("/:application_id/repositories/*repo_key", appHandler.DeleteApplicationRepository)
      apps.GET("/:application_id/projects", appHandler.ListApplicationProjects)
      apps.POST("/:application_id/projects", appHandler.CreateApplicationProject)
      apps.GET("/:application_id/rollup", appHandler.ApplicationRollup)
      apps.GET("/:application_id/dashboard", appHandler.ApplicationDashboard)
  }
  projects := v1.Group("/projects")
  {
      projects.POST("", appHandler.CreateProjectStandalone)
      projects.GET("/standalone", appHandler.ListStandaloneProjects)
      projects.GET("/:project_id", appHandler.GetProject)
      projects.PATCH("/:project_id", appHandler.UpdateProject)
      projects.DELETE("/:project_id", appHandler.ArchiveProject)
      projects.GET("/:project_id/repositories", appHandler.ListProjectRepositories)
      projects.POST("/:project_id/repositories", appHandler.CreateProjectRepository)
      projects.DELETE("/:project_id/repositories/:repository_id", appHandler.DeleteProjectRepository)
  }
  repos := v1.Group("/repositories")
  {
      repos.GET("", appHandler.Repositories)
      repos.POST("", appHandler.CreateRepositoryDraft)
      repos.POST("/:repository_id/publish", appHandler.RequestRepositoryPublish)
      repos.GET("/:repository_id/activity", appHandler.RepositoryActivity)
      repos.GET("/:repository_id/pull-requests", appHandler.RepositoryPullRequests)
      repos.GET("/:repository_id/build-runs", appHandler.RepositoryBuildRuns)
      repos.GET("/:repository_id/quality-snapshots", appHandler.RepositoryQualitySnapshots)
  }
  ```

---

## 2. 세부 작업 체크리스트 (Stage 2 Step-by-Step)

### 3.1 [1단계] 모듈형 핸들러 구조체 정의 (각 도메인의 `view/` 내)
- [ ] `auth-session/view/` 하위에 `AuthHandler` 구조체 정의 및 리시버 메서드 리팩토링
- [ ] `audit-ops/view/` 하위에 `AuditHandler` 구조체 정의 및 리시버 메서드 리팩토링
- [ ] `rbac-permissions/view/` 하위에 `RBACHandler` 구조체 정의 및 리시버 메서드 리팩토링
- [ ] `organization-management/view/` 하위에 `OrganizationHandler` 구조체 정의 및 리시버 메서드 리팩토링
- [ ] `application-lifecycle/view/` 하위에 `ApplicationHandler` 구조체 정의 및 리시버 메서드 리팩토링
- [ ] `dev-request/view/` 하위에 `DevRequestHandler` 구조체 정의 및 리시버 메서드 리팩토링
- [ ] `integration-registry/view/` 하위에 `IntegrationHandler` 구조체 정의 및 리시버 메서드 리팩토링
- [ ] `realtime/view/` 하위에 `RealtimeHandler` 구조체 정의 및 리시버 메서드 리팩토링

### 3.2 [2단계] `httpapi/router.go` 개편 및 Declarative Grouping 구현
- [ ] `router.go` 의 단일 `Handler` 리시버 메서드 구조 소거 및 각 도메인 핸들러 패키지 임포트
- [ ] `NewRouter` 내부에서 10대 도메인별 **`v1.Group("/domain-path")`** 수직 그룹 라우팅 레이아웃 구현 및 미들웨어 선언적 바인딩

### 3.3 [3단계] 검증 및 무결성 빌드
- [ ] `/backend-core` 전체 빌드 검증 (`go build ./...`)
- [ ] 백엔드 패키지 유닛 테스트 실행 및 패스 검증 (`go test -short ./...`)

---

## 4. 진척도 추적 (Progress Tracking)

* **Stage 2 시작일**: 2026-05-28
* **Stage 2 완료 목표일**: 2026-05-28
* **Stage 2 진척률**: `5%` (파일 물리 이관 완료, 모듈형 핸들러 및 라우터 그룹 정비 계획 수립 완료)
