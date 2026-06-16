# Work Backlog — claude/work_260514-b

- 상태 관리: planned / in_progress / blocked / done

## A1. postgres_applications.go store body

- [planned] `(s *PostgresStore) ListApplications(ctx, opts) ([]Application, int, error)` — filter / pagination / total count
- [planned] `GetApplication(ctx, id) (Application, error)` — pgx.ErrNoRows → ErrNotFound
- [planned] `CreateApplication(ctx, app) (Application, error)` — UNIQUE key 위반 → ErrConflict
- [planned] `UpdateApplication(ctx, app) (Application, error)` — 단일 UPDATE, archived consistency 강제
- [planned] `ArchiveApplication(ctx, id, reason) (Application, error)` — status=archived + archived_at = NOW
- [planned] `ListApplicationRepositories(ctx, appID) ([]ApplicationRepository, error)`
- [planned] `CreateApplicationRepository(ctx, link) (ApplicationRepository, error)` — composite PK 충돌 → ErrConflict
- [planned] `DeleteApplicationRepository(ctx, key) error`
- [planned] `UpdateApplicationRepositorySync(ctx, key, status, errCode) error` — sync_status / sync_error_code / retryable / at 일관 갱신
- [planned] `ListSCMProviders(ctx) ([]SCMProvider, error)`
- [planned] `UpdateSCMProvider(ctx, provider) (SCMProvider, error)` — enabled / display_name 만 갱신 (adapter_version 보호)
- [planned] `ListProjects(ctx, opts) ([]Project, int, error)` — repository_id scope
- [planned] `GetProject(ctx, id) (Project, error)`
- [planned] `CreateProject(ctx, p) (Project, error)` — UNIQUE (repository_id, key)
- [planned] `UpdateProject(ctx, p) (Project, error)`
- [planned] `ArchiveProject(ctx, id, reason) (Project, error)`

## A2. handler body (API-41..50)

- [planned] `listSCMProviders` (API-41) — store.ListSCMProviders → envelope
- [planned] `updateSCMProvider` (API-42) — body binding (enabled / display_name 만), store.UpdateSCMProvider, adapter_version 변경 시 422
- [planned] `listApplications` (API-43) — query (status / include_archived / q / limit / offset), store.ListApplications → envelope + meta(total)
- [planned] `createApplication` (API-44) — body binding + key 정규식 (`^[A-Za-z0-9]{10}$`), status='planning' 강제 (또는 허용 set), store.CreateApplication, 409 application_key_conflict 매핑, audit `application.create.requested`
- [planned] `getApplication` (API-45) — store.GetApplication, ErrNotFound → 404, audit `application.get.requested` (carve out 후보)
- [planned] `updateApplication` (API-46) — body binding, key 포함 시 `422 application_key_immutable`, 상태 전이 가드 (concept §13.2.1: planning→active 의 연결 repo ≥1, on_hold→active 의 resume_reason 필수), 422 분기 별 코드, audit `application.update.requested`
- [planned] `archiveApplication` (API-47) — store.ArchiveApplication, audit `application.archive.requested`
- [planned] `listApplicationRepositories` (API-48) — store.ListApplicationRepositories
- [planned] `createApplicationRepository` (API-49) — body binding (repo_provider / repo_full_name / role), provider 등록 여부 검증 (`unsupported_repo_provider` 422), 409 repository_link_conflict, sync_status='requested' 기본, audit `application_repository.link.requested`
- [planned] `deleteApplicationRepository` (API-50) — store.DeleteApplicationRepository, audit `application_repository.unlink.requested`

## A3. 단위테스트

- [planned] `postgres_applications_test.go` — 각 메서드 happy + UNIQUE 위반 + FK 위반 + ErrNotFound. integration test 패턴 (`store.skipIfNoDB`).
- [planned] `applications_test.go` — 각 handler 의 RBAC denial (developer/manager → 403) + system_admin 정상 + invalid body (400) + 검증 실패 (422) + 충돌 (409).
- [planned] 상태 전이 가드 테스트 — planning→active 의 활성 repo 0건 케이스 → 422 application_activation_precondition_failed.

## C. 매트릭스 + API 계약 갱신

- [planned] trace.md §2.2 — Application/Repository API 표의 상태 컬럼 scaffolded → activated
- [planned] trace.md §2.4 — IMPL-application-store-01 의 책임을 stub → 본체로 갱신 + 신규 sub-IMPL (예: IMPL-application-handler-status-transition-01)
- [planned] trace.md §3 — Application/Project 도메인 row 의 UT 컬럼 채움
- [planned] trace.md §6 변경 이력 row 추가 (sprint claude/work_260514-b)
- [planned] backend_api_contract.md §13.0 — scaffolded → activated 상태 갱신
- [planned] backend_api_contract.md §13.2~§13.3 — 응답 schema 본문 구체화 (필요 시)
- [planned] PR body 추적성 영향 섹션

## D. 위험 / 검토

- [planned] 상태 전이 가드의 critical 롤업 0건 검증 — 롤업 store 의존. 1차 carve out 결정.
- [planned] audit action 명명 convention — `application.{verb}.requested` 패턴 vs 기존 `auth.role_denied` / `account.signup.partial_failure` 등과 정합.
- [planned] PR 크기 — 1500+ lines 예상. commit 단위 정리 (store / handler / UT) 후 squash merge.

## 완료
- (없음)
