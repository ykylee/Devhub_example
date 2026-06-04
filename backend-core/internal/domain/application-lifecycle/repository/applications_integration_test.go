package repository_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	apprep "github.com/devhub/backend-core/internal/domain/application-lifecycle/repository"
	"github.com/devhub/backend-core/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Application 도메인 PostgreSQL integration test (sprint claude/work_260514-e).
//
// CI backend-unit job 은 DEVHUB_TEST_DB_URL 미설정으로 t.Skip — 기존 postgres_*_test.go
// 패턴과 일관. 로컬 / 후속 CI 분리 잡에서 마이그레이션 000012..000018 가 적용된 DB
// 환경에서 실 실행.

const (
	testAppKey1 = "TESTAPP001"
	testAppKey2 = "TESTAPP002"
	testRepoID1 = int64(99001) // fixture 가 생성
	testRepoID2 = int64(99002)
)

// applicationsFixture 는 본 sprint 의 테스트가 의존하는 DB 상태를 보장한다:
//   - applications / application_repositories / projects / project_members / project_integrations
//     의 모든 row 를 cleanup (TRUNCATE CASCADE)
//   - test repository 2개 추가 (testRepoID1, testRepoID2)
//   - SCM provider 카탈로그는 migration 000012 의 seed 그대로 유지
//
// PR #109 codex review P1 정정 (sprint claude/work_260514-f) — cleanup 을 multi-
// statement + bind args 단일 호출에서 두 statement 분리 호출로 변경. pgx 의
// prepared execution 은 multi-command query 를 거부하므로 이전 버전은 fixture
// 단계에서 panic.
func applicationsFixture(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	const cleanupStatic = `
TRUNCATE TABLE project_members, project_integrations, projects,
               application_repositories, applications,
               pr_activities, build_runs, ci_runs, quality_snapshots RESTART IDENTITY CASCADE;`
	if _, err := pool.Exec(ctx, cleanupStatic); err != nil {
		t.Fatalf("cleanup static tables: %v", err)
	}
	const cleanupRepos = `DELETE FROM repositories WHERE id IN ($1, $2);`
	if _, err := pool.Exec(ctx, cleanupRepos, testRepoID1, testRepoID2); err != nil {
		t.Fatalf("cleanup test repositories: %v", err)
	}
	const seedRepos = `
INSERT INTO repositories (id, gitea_repository_id, full_name, name)
VALUES ($1, 8001, 'team/devhub-core', 'devhub-core'),
       ($2, 8002, 'team/devhub-web',  'devhub-web')`
	if _, err := pool.Exec(ctx, seedRepos, testRepoID1, testRepoID2); err != nil {
		t.Fatalf("seed repositories: %v", err)
	}
}

// PR #109 codex P1 회귀 guard — fixture 가 정상 동작하는지 sanity check.
// applicationsFixture 가 panic 없이 끝나고, repositories 가 seed 되었는지만 확인.
// 이전 버전 (multi-statement + bind args) 이 회귀하면 setupApplicationsTest 안에서
// applicationsFixture 가 t.Fatalf 로 실패하므로 본 test 도 즉시 실패.
func TestIntegration_FixtureCleanupSanity(t *testing.T) {
	pgStore, pool, ctx, teardown := setupApplicationsTest(t)
	defer teardown()
	_ = pgStore

	// repositories 가 seed 되었는지 raw pool 로 확인.
	var count int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM repositories WHERE id IN ($1, $2)`,
		testRepoID1, testRepoID2).Scan(&count); err != nil {
		t.Fatalf("count repositories: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 seeded repositories, got %d", count)
	}
}

// applicationsFixtureLockID — `go test ./...` cross-pkg race guard.
//
// 본 패키지 (application-lifecycle/repository_test) 와 internal/store_test 양쪽이
// 동일 테이블 (applications, projects, project_integrations, application_repositories,
// project_members, repositories) 을 TRUNCATE CASCADE 하는 fixture 를 동시 실행한다.
// `go test ./...` 는 패키지 단위로 별도 binary 를 병렬 실행하므로, A 패키지의 test
// 진행 중에 B 패키지 fixture 가 끼어들면 row 가 사라져 silent fail 회귀 (예: PR
// #437 의 SQLSTATE 22P02 uuid empty case) 가 발생한다.
//
// 양쪽 setup 이 동일한 `pg_advisory_lock` id 를 잡아 fixture+test lifecycle 을
// 시리얼라이즈한다. 같은 conn 으로 session lock 잡고 teardown 에서 release.
//
// 본 상수가 변경되면 다음 파일의 대응 상수도 같이 변경해야 한다:
//   - internal/store/integration_test_helpers_test.go
const applicationsFixtureLockID = int64(0x4150705F4C6966) // "App_Lif" ASCII

func setupApplicationsTest(t *testing.T) (*apprep.ApplicationRepository, *pgxpool.Pool, context.Context, func()) {
	t.Helper()
	dbURL := os.Getenv("DEVHUB_TEST_DB_URL")
	if dbURL == "" {
		t.Skip("DEVHUB_TEST_DB_URL is not set")
	}
	ctx := context.Background()
	pgStore, err := store.NewPostgresStore(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect postgres store: %v", err)
	}
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		pgStore.Close()
		t.Fatalf("connect raw pool: %v", err)
	}
	lockConn, err := pool.Acquire(ctx)
	if err != nil {
		pool.Close()
		pgStore.Close()
		t.Fatalf("acquire advisory lock conn: %v", err)
	}
	if _, err := lockConn.Exec(ctx, `SELECT pg_advisory_lock($1)`, applicationsFixtureLockID); err != nil {
		lockConn.Release()
		pool.Close()
		pgStore.Close()
		t.Fatalf("pg_advisory_lock: %v", err)
	}
	applicationsFixture(t, ctx, pool)
	repo := apprep.NewApplicationRepository(pgStore)
	return repo, pool, ctx, func() {
		_, _ = lockConn.Exec(ctx, `SELECT pg_advisory_unlock($1)`, applicationsFixtureLockID)
		lockConn.Release()
		pool.Close()
		pgStore.Close()
	}
}

// --- Applications ---

func TestIntegration_CreateApplication_Happy(t *testing.T) {
	pgStore, _, ctx, teardown := setupApplicationsTest(t)
	defer teardown()

	app := domain.Application{
		Key: testAppKey1, Name: "Test App", Status: domain.ApplicationStatusPlanning,
		Visibility: domain.ApplicationVisibilityInternal, OwnerUserID: "u1",
	}
	created, err := pgStore.CreateApplication(ctx, app)
	if err != nil {
		t.Fatalf("create application: %v", err)
	}
	if created.Key != testAppKey1 {
		t.Errorf("key = %q, want %q", created.Key, testAppKey1)
	}
	if created.ID == "" {
		t.Errorf("ID should be generated, got empty")
	}
}

func TestIntegration_CreateApplication_DuplicateKey(t *testing.T) {
	pgStore, _, ctx, teardown := setupApplicationsTest(t)
	defer teardown()

	app := domain.Application{
		Key: testAppKey1, Name: "App", Status: domain.ApplicationStatusPlanning,
		Visibility: domain.ApplicationVisibilityInternal, OwnerUserID: "u1",
	}
	if _, err := pgStore.CreateApplication(ctx, app); err != nil {
		t.Fatalf("seed: %v", err)
	}
	_, err := pgStore.CreateApplication(ctx, app)
	if !errors.Is(err, store.ErrConflict) {
		t.Errorf("expected ErrConflict on duplicate key, got %v", err)
	}
}

func TestIntegration_GetApplication_NotFound(t *testing.T) {
	pgStore, _, ctx, teardown := setupApplicationsTest(t)
	defer teardown()

	_, err := pgStore.GetApplication(ctx, "00000000-0000-0000-0000-000000000000")
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestIntegration_UpdateApplication_ArchivedConsistency(t *testing.T) {
	pgStore, _, ctx, teardown := setupApplicationsTest(t)
	defer teardown()

	app := domain.Application{
		Key: testAppKey1, Name: "App", Status: domain.ApplicationStatusActive,
		Visibility: domain.ApplicationVisibilityInternal, OwnerUserID: "u1",
	}
	created, err := pgStore.CreateApplication(ctx, app)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	// status='archived' 로 update 시 archived_at 자동 채워짐
	created.Status = domain.ApplicationStatusArchived
	updated, err := pgStore.UpdateApplication(ctx, created)
	if err != nil {
		t.Fatalf("update archived: %v", err)
	}
	if updated.ArchivedAt == nil {
		t.Errorf("archived_at should be set when status=archived")
	}
	// status='active' 로 revert 시 archived_at NULL 로 reset
	updated.Status = domain.ApplicationStatusActive
	reverted, err := pgStore.UpdateApplication(ctx, updated)
	if err != nil {
		t.Fatalf("update revert: %v", err)
	}
	if reverted.ArchivedAt != nil {
		t.Errorf("archived_at should be NULL when status != archived")
	}
}

func TestIntegration_ArchiveApplication_SetsTimestamp(t *testing.T) {
	pgStore, _, ctx, teardown := setupApplicationsTest(t)
	defer teardown()

	app := domain.Application{
		Key: testAppKey1, Name: "App", Status: domain.ApplicationStatusActive,
		Visibility: domain.ApplicationVisibilityInternal, OwnerUserID: "u1",
	}
	created, _ := pgStore.CreateApplication(ctx, app)
	archived, err := pgStore.ArchiveApplication(ctx, created.ID, "test reason")
	if err != nil {
		t.Fatalf("archive: %v", err)
	}
	if archived.Status != domain.ApplicationStatusArchived {
		t.Errorf("status = %q, want archived", archived.Status)
	}
	if archived.ArchivedAt == nil {
		t.Errorf("archived_at should be set")
	}
}

func TestIntegration_ListApplications_Filter(t *testing.T) {
	pgStore, _, ctx, teardown := setupApplicationsTest(t)
	defer teardown()

	statuses := []domain.ApplicationStatus{
		domain.ApplicationStatusPlanning,
		domain.ApplicationStatusActive,
		domain.ApplicationStatusArchived,
	}
	for i, status := range statuses {
		key := fmt.Sprintf("TEST00000%d", i+1)
		_, err := pgStore.CreateApplication(ctx, domain.Application{
			Key: key, Name: "X", Status: status,
			Visibility: domain.ApplicationVisibilityInternal, OwnerUserID: "u1",
		})
		if err != nil {
			t.Fatalf("seed %d: %v", i, err)
		}
		if status == domain.ApplicationStatusArchived {
			// CreateApplication 은 status='archived' 도 허용하므로 archived_at 직접 set
			apps, _, _ := pgStore.ListApplications(ctx, store.ApplicationListOptions{IncludeArchived: true})
			for _, a := range apps {
				if a.Key == key {
					a.Status = domain.ApplicationStatusArchived
					if _, err := pgStore.UpdateApplication(ctx, a); err != nil {
						t.Fatalf("set archived: %v", err)
					}
				}
			}
		}
	}

	// default: archived 제외 → 2건
	_, total, err := pgStore.ListApplications(ctx, store.ApplicationListOptions{})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 2 {
		t.Errorf("default list total = %d, want 2 (archived 제외)", total)
	}
	// include_archived=true → 3건
	_, total, _ = pgStore.ListApplications(ctx, store.ApplicationListOptions{IncludeArchived: true})
	if total != 3 {
		t.Errorf("include_archived total = %d, want 3", total)
	}
	// status=active → 1건
	_, total, _ = pgStore.ListApplications(ctx, store.ApplicationListOptions{Status: "active"})
	if total != 1 {
		t.Errorf("status=active total = %d, want 1", total)
	}
}

// --- Application-Repository link + sync ---

func TestIntegration_CreateApplicationRepository_CompositeKey(t *testing.T) {
	pgStore, _, ctx, teardown := setupApplicationsTest(t)
	defer teardown()

	app, _ := pgStore.CreateApplication(ctx, domain.Application{
		Key: testAppKey1, Name: "X", Status: domain.ApplicationStatusPlanning,
		Visibility: domain.ApplicationVisibilityInternal, OwnerUserID: "u1",
	})
	link := domain.ApplicationRepository{
		ApplicationID: app.ID, RepoProvider: "gitea", RepoFullName: "team/devhub-core",
		Role: domain.ApplicationRepositoryRolePrimary, SyncStatus: domain.SyncStatusRequested,
	}
	if _, err := pgStore.CreateApplicationRepository(ctx, link); err != nil {
		t.Fatalf("create link: %v", err)
	}
	// 동일 composite key 중복 → ErrConflict
	_, err := pgStore.CreateApplicationRepository(ctx, link)
	if !errors.Is(err, store.ErrConflict) {
		t.Errorf("expected ErrConflict on duplicate composite key, got %v", err)
	}
	// 다른 provider 같은 repo_full_name → OK (composite PK 포함)
	link.RepoProvider = "bitbucket"
	if _, err := pgStore.CreateApplicationRepository(ctx, link); err != nil {
		t.Errorf("different provider should not conflict: %v", err)
	}
}

func TestIntegration_CountActiveApplicationRepositories(t *testing.T) {
	pgStore, _, ctx, teardown := setupApplicationsTest(t)
	defer teardown()

	app, _ := pgStore.CreateApplication(ctx, domain.Application{
		Key: testAppKey1, Name: "X", Status: domain.ApplicationStatusPlanning,
		Visibility: domain.ApplicationVisibilityInternal, OwnerUserID: "u1",
	})
	// 0개 — count=0
	count, err := pgStore.CountActiveApplicationRepositories(ctx, app.ID)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 0 {
		t.Errorf("initial count = %d, want 0", count)
	}
	// 1 active 추가
	_, _ = pgStore.CreateApplicationRepository(ctx, domain.ApplicationRepository{
		ApplicationID: app.ID, RepoProvider: "gitea", RepoFullName: "team/devhub-core",
		Role: domain.ApplicationRepositoryRolePrimary, SyncStatus: domain.SyncStatusActive,
	})
	count, _ = pgStore.CountActiveApplicationRepositories(ctx, app.ID)
	if count != 1 {
		t.Errorf("after 1 active, count = %d, want 1", count)
	}
	// 1 requested (활성 아님) 추가
	_, _ = pgStore.CreateApplicationRepository(ctx, domain.ApplicationRepository{
		ApplicationID: app.ID, RepoProvider: "bitbucket", RepoFullName: "team/devhub-web",
		Role: domain.ApplicationRepositoryRoleSub, SyncStatus: domain.SyncStatusRequested,
	})
	count, _ = pgStore.CountActiveApplicationRepositories(ctx, app.ID)
	if count != 1 {
		t.Errorf("requested not counted, count = %d, want 1", count)
	}
}

func TestIntegration_UpdateApplicationRepositorySync_ErrorReset(t *testing.T) {
	pgStore, _, ctx, teardown := setupApplicationsTest(t)
	defer teardown()

	app, _ := pgStore.CreateApplication(ctx, domain.Application{
		Key: testAppKey1, Name: "X", Status: domain.ApplicationStatusPlanning,
		Visibility: domain.ApplicationVisibilityInternal, OwnerUserID: "u1",
	})
	_, _ = pgStore.CreateApplicationRepository(ctx, domain.ApplicationRepository{
		ApplicationID: app.ID, RepoProvider: "gitea", RepoFullName: "team/devhub-core",
		Role: domain.ApplicationRepositoryRolePrimary, SyncStatus: domain.SyncStatusRequested,
	})
	key := store.ApplicationRepositoryLinkKey{
		ApplicationID: app.ID, RepoProvider: "gitea", RepoFullName: "team/devhub-core",
	}
	// 에러 발생
	if err := pgStore.UpdateApplicationRepositorySync(ctx, key, domain.SyncStatusDegraded, domain.SyncErrorRateLimited); err != nil {
		t.Fatalf("set error: %v", err)
	}
	links, _ := pgStore.ListApplicationRepositories(ctx, app.ID)
	if len(links) != 1 || links[0].SyncErrorCode != domain.SyncErrorRateLimited {
		t.Fatalf("link state after error: %+v", links)
	}
	if links[0].SyncErrorRetryable == nil || !*links[0].SyncErrorRetryable {
		t.Errorf("rate_limited should be retryable=true")
	}
	// 에러 reset (errorCode 빈 문자열) → retryable / at NULL
	if err := pgStore.UpdateApplicationRepositorySync(ctx, key, domain.SyncStatusActive, ""); err != nil {
		t.Fatalf("reset error: %v", err)
	}
	links, _ = pgStore.ListApplicationRepositories(ctx, app.ID)
	if links[0].SyncErrorCode != "" || links[0].SyncErrorRetryable != nil || links[0].SyncErrorAt != nil {
		t.Errorf("error fields should be NULL after reset, got %+v", links[0])
	}
}

// --- SCM provider ---

func TestIntegration_UpdateSCMProvider_AdapterVersionPreserved(t *testing.T) {
	pgStore, _, ctx, teardown := setupApplicationsTest(t)
	defer teardown()

	// store UpdateSCMProvider 는 adapter_version 을 갱신하지 않음 (handler 가 막아주지만
	// store 가 unconditionally 갱신하면 위험)
	providers, _ := pgStore.ListSCMProviders(ctx)
	var gitea domain.SCMProvider
	for _, p := range providers {
		if p.ProviderKey == "gitea" {
			gitea = p
			break
		}
	}
	if gitea.ProviderKey != "gitea" {
		t.Fatalf("gitea seed not found")
	}
	originalVersion := gitea.AdapterVersion
	gitea.AdapterVersion = "9.9.9-injected"
	gitea.DisplayName = "Gitea Renamed"
	updated, err := pgStore.UpdateSCMProvider(ctx, gitea)
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.AdapterVersion != originalVersion {
		t.Errorf("adapter_version should be preserved, got %q (want %q)", updated.AdapterVersion, originalVersion)
	}
	if updated.DisplayName != "Gitea Renamed" {
		t.Errorf("display_name not updated: %q", updated.DisplayName)
	}
}

// --- Extra Comprehensive Coverage Test Cases ---

func TestIntegration_Applications_ExtraCRUD(t *testing.T) {
	pgStore, _, ctx, teardown := setupApplicationsTest(t)
	defer teardown()

	ghostUUID := "00000000-0000-0000-0000-000000000000"

	// 1. GetApplication not found
	if _, err := pgStore.GetApplication(ctx, ghostUUID); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound for ghost UUID, got %v", err)
	}

	// 2. GetApplicationByKey not found
	if _, err := pgStore.GetApplicationByKey(ctx, "GHOSTKEY"); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound for GHOSTKEY, got %v", err)
	}

	// 3. UpdateApplication not found
	appGhost := domain.Application{ID: ghostUUID, Key: "GHOST", Name: "Ghost"}
	if _, err := pgStore.UpdateApplication(ctx, appGhost); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound for UpdateApplication ghost, got %v", err)
	}

	// 4. ArchiveApplication not found
	if _, err := pgStore.ArchiveApplication(ctx, ghostUUID, "reason"); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound for ArchiveApplication ghost, got %v", err)
	}

	// 5. DeleteApplication not found
	if err := pgStore.DeleteApplication(ctx, ghostUUID); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound for DeleteApplication ghost, got %v", err)
	}

	// 6. DeleteApplicationRepository not found
	keyGhost := store.ApplicationRepositoryLinkKey{ApplicationID: ghostUUID, RepoProvider: "gitea", RepoFullName: "ghost/repo"}
	if err := pgStore.DeleteApplicationRepository(ctx, keyGhost); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound for DeleteApplicationRepository, got %v", err)
	}

	// 7. UpdateSCMProvider not found
	pGhost := domain.SCMProvider{ProviderKey: "ghostprovider", DisplayName: "Ghost"}
	if _, err := pgStore.UpdateSCMProvider(ctx, pGhost); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound for UpdateSCMProvider, got %v", err)
	}

	// 8. UpdateApplicationRepositorySync not found
	if err := pgStore.UpdateApplicationRepositorySync(ctx, keyGhost, domain.SyncStatusActive, ""); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound for UpdateApplicationRepositorySync, got %v", err)
	}

	// 9. ListApplications options coverage
	opts := store.ApplicationListOptions{
		Status:          "planning",
		IncludeArchived: true,
		Query:           "non-existent-query-string",
	}
	apps, total, err := pgStore.ListApplications(ctx, opts)
	if err != nil {
		t.Fatalf("ListApplications: %v", err)
	}
	if len(apps) != 0 || total != 0 {
		t.Errorf("expected 0 results, got len=%d total=%d", len(apps), total)
	}
}

func TestIntegration_Projects_ExtraCRUD(t *testing.T) {
	pgStore, pool, ctx, teardown := setupApplicationsTest(t)
	defer teardown()

	ghostUUID := "00000000-0000-0000-0000-000000000000"

	// 1. GetProject not found
	if _, err := pgStore.GetProject(ctx, ghostUUID); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound for GetProject ghost, got %v", err)
	}

	// 2. CreateProject and Duplicate check
	pKey := fmt.Sprintf("PRJ-%d", time.Now().UnixNano())
	defer func() { _, _ = pool.Exec(ctx, `DELETE FROM projects WHERE key = $1`, pKey) }()

	p := domain.Project{
		Key:         pKey,
		Name:        "Test Project",
		Status:      domain.ApplicationStatusActive,
		Visibility:  domain.ApplicationVisibilityInternal,
		OwnerUserID: "u1",
		ProjectMembers: []domain.ProjectMember{
			{UserID: "member1", ProjectRole: "contributor"},
			{UserID: "member2", ProjectRole: "observer"},
		},
	}
	created, err := pgStore.CreateProject(ctx, p)
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}
	if created.Key != pKey {
		t.Errorf("expected project key %s, got %s", pKey, created.Key)
	}

	// Fetch to verify project members creation
	fetched, err := pgStore.GetProject(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetProject failed to verify members: %v", err)
	}
	if len(fetched.ProjectMembers) != 2 {
		t.Errorf("expected 2 project members, got %d", len(fetched.ProjectMembers))
	}

	// Duplicate key conflict check
	_, err = pgStore.CreateProject(ctx, p)
	if !errors.Is(err, store.ErrConflict) {
		t.Errorf("expected ErrConflict for duplicate project key, got %v", err)
	}

	// 3. UpdateProject
	created.Name = "Test Project Updated"
	created.ProjectMembers = []domain.ProjectMember{
		{UserID: "member1", ProjectRole: "lead"},
		{UserID: "member3", ProjectRole: "contributor"},
	}
	updated, err := pgStore.UpdateProject(ctx, created)
	if err != nil {
		t.Fatalf("UpdateProject failed: %v", err)
	}
	if updated.Name != "Test Project Updated" {
		t.Errorf("expected name updated, got %s", updated.Name)
	}

	// Fetch to verify synced project members
	fetchedUpdated, err := pgStore.GetProject(ctx, updated.ID)
	if err != nil {
		t.Fatalf("GetProject failed to verify updated members: %v", err)
	}
	if len(fetchedUpdated.ProjectMembers) != 2 {
		t.Errorf("expected 2 project members after update, got %d", len(fetchedUpdated.ProjectMembers))
	}
	var hasMember3 bool
	for _, m := range fetchedUpdated.ProjectMembers {
		if m.UserID == "member3" {
			hasMember3 = true
		}
	}
	if !hasMember3 {
		t.Errorf("expected member3 to be in updated project members")
	}

	// Update ghost project
	pGhost := domain.Project{ID: ghostUUID, Key: "GHOST", Name: "Ghost"}
	if _, err := pgStore.UpdateProject(ctx, pGhost); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound for UpdateProject ghost, got %v", err)
	}

	// 4. ListProjects filter coverage
	opts := store.ProjectListOptions{
		Status:         "active",
		StandaloneOnly: true,
	}
	_, _, err = pgStore.ListProjects(ctx, opts)
	if err != nil {
		t.Errorf("ListProjects error: %v", err)
	}

	// 5. Project Repository Links CRUD
	link := domain.ProjectRepository{
		ProjectID:    created.ID,
		RepositoryID: testRepoID1,
		Role:         "linked",
	}
	createdLink, err := pgStore.CreateProjectRepository(ctx, link)
	if err != nil {
		t.Fatalf("CreateProjectRepository failed: %v", err)
	}
	if createdLink.ProjectID != created.ID || createdLink.RepositoryID != testRepoID1 {
		t.Errorf("unexpected created link details: %+v", createdLink)
	}

	// Duplicate link conflict check
	_, err = pgStore.CreateProjectRepository(ctx, link)
	if !errors.Is(err, store.ErrConflict) {
		t.Errorf("expected ErrConflict for duplicate project repository link, got %v", err)
	}

	// List
	links, err := pgStore.ListProjectRepositories(ctx, created.ID)
	if err != nil {
		t.Fatalf("ListProjectRepositories: %v", err)
	}
	if len(links) != 1 || links[0].RepositoryID != testRepoID1 {
		t.Errorf("expected 1 project repository link, got %+v", links)
	}

	// Delete Project Repository Link
	if err := pgStore.DeleteProjectRepository(ctx, created.ID, testRepoID1); err != nil {
		t.Fatalf("DeleteProjectRepository failed: %v", err)
	}
	// Delete non-existent link
	if err := pgStore.DeleteProjectRepository(ctx, created.ID, 99999); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound for non-existent link deletion, got %v", err)
	}

	// 6. ArchiveProject and DeleteProject
	archived, err := pgStore.ArchiveProject(ctx, created.ID, "some reason")
	if err != nil {
		t.Fatalf("ArchiveProject failed: %v", err)
	}
	if archived.Status != domain.ApplicationStatusArchived {
		t.Errorf("expected status archived, got %s", archived.Status)
	}

	// Archive ghost
	if _, err := pgStore.ArchiveProject(ctx, ghostUUID, "reason"); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound for ArchiveProject ghost, got %v", err)
	}

	// Delete Project
	if err := pgStore.DeleteProject(ctx, created.ID); err != nil {
		t.Fatalf("DeleteProject failed: %v", err)
	}
	// Delete ghost
	if err := pgStore.DeleteProject(ctx, ghostUUID); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound for DeleteProject ghost, got %v", err)
	}
}

func TestIntegration_RepositoryDrafts_ExtraCRUD(t *testing.T) {
	pgStore, pool, ctx, teardown := setupApplicationsTest(t)
	defer teardown()

	slug := fmt.Sprintf("drafts/repo-%d", time.Now().UnixNano())
	key := fmt.Sprintf("draft-%d", time.Now().UnixNano())
	defer func() { _, _ = pool.Exec(ctx, `DELETE FROM repositories WHERE full_name = $1`, slug) }()

	// 1. Create Repository Draft
	draft, err := pgStore.CreateRepositoryDraft(ctx, key, slug, "")
	if err != nil {
		t.Fatalf("CreateRepositoryDraft failed: %v", err)
	}
	if draft.FullName != slug || draft.Status != "draft" {
		t.Errorf("unexpected draft details: %+v", draft)
	}

	// Duplicate slug conflict check
	_, err = pgStore.CreateRepositoryDraft(ctx, key, slug, "")
	if !errors.Is(err, store.ErrConflict) {
		t.Errorf("expected ErrConflict for duplicate slug, got %v", err)
	}

	// Create with empty inputs
	if _, err := pgStore.CreateRepositoryDraft(ctx, "", slug, ""); !errors.Is(err, store.ErrConflict) {
		t.Errorf("expected ErrConflict for empty key, got %v", err)
	}

	// 2. MarkRepositoryDraftPublishRequested
	updated, err := pgStore.MarkRepositoryDraftPublishRequested(ctx, draft.ID)
	if err != nil {
		t.Fatalf("MarkRepositoryDraftPublishRequested failed: %v", err)
	}
	if updated.PublishRequestedAt == nil {
		t.Errorf("expected PublishRequestedAt to be set")
	}

	// Mark ghost publish requested
	if _, err := pgStore.MarkRepositoryDraftPublishRequested(ctx, 999999); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound for ghost publish requested, got %v", err)
	}

	// 3. GetRepositoryByID
	loaded, err := pgStore.GetRepositoryByID(ctx, draft.ID)
	if err != nil {
		t.Fatalf("GetRepositoryByID failed: %v", err)
	}
	if loaded.ID != draft.ID || loaded.FullName != slug {
		t.Errorf("loaded repo details mismatch: %+v", loaded)
	}

	// Get ghost by ID
	if _, err := pgStore.GetRepositoryByID(ctx, 999999); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound for ghost repository id, got %v", err)
	}
}

// TestIntegration_UpdateRepositoryDraft — draft 만 부분 갱신 가능.
func TestIntegration_UpdateRepositoryDraft(t *testing.T) {
	pgStore, pool, ctx, teardown := setupApplicationsTest(t)
	defer teardown()

	slug := fmt.Sprintf("drafts/update-%d", time.Now().UnixNano())
	key := fmt.Sprintf("upd-%d", time.Now().UnixNano())
	defer func() { _, _ = pool.Exec(ctx, `DELETE FROM repositories WHERE full_name = $1`, slug) }()

	// ensure integration provider 'gitea' exists (E2E seed 가 이미 있지만 idempotent 보장)
	if _, err := pool.Exec(ctx, `
INSERT INTO integration_providers (provider_id, provider_key, provider_type, display_name, enabled, auth_mode, credentials_ref, capabilities, sync_status, base_url, api_token)
VALUES ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'gitea', 'scm', 'Local Gitea', true, 'token', 'cr', '["push"]'::jsonb, 'active', 'http://localhost:3000', 'tok')
ON CONFLICT (provider_key) DO NOTHING
`); err != nil {
		t.Fatalf("seed gitea provider: %v", err)
	}

	draft, err := pgStore.CreateRepositoryDraft(ctx, key, slug, "")
	if err != nil {
		t.Fatalf("CreateRepositoryDraft failed: %v", err)
	}
	if draft.ProviderID != "" {
		t.Errorf("expected empty ProviderID for provider-less draft, got %q", draft.ProviderID)
	}

	// 1. nil params → no-op (unchanged)
	if _, err := pgStore.UpdateRepositoryDraft(ctx, draft.ID, store.RepositoryUpdateDraftParams{}); err != nil {
		t.Errorf("no-op update failed: %v", err)
	}

	// 2. update key + slug
	newKey := key + "-X"
	newSlug := slug + "-renamed"
	updated, err := pgStore.UpdateRepositoryDraft(ctx, draft.ID, store.RepositoryUpdateDraftParams{
		Key:  &newKey,
		Slug: &newSlug,
	})
	if err != nil {
		t.Fatalf("UpdateRepositoryDraft key+slug failed: %v", err)
	}
	if updated.Name != newKey || updated.FullName != newSlug {
		t.Errorf("key/slug not updated: name=%q full_name=%q", updated.Name, updated.FullName)
	}
	// codex P2 fix — slug 변경 시 owner_login 재계산
	if updated.OwnerLogin != "drafts" {
		t.Errorf("owner_login not recomputed from new slug: got %q, want %q", updated.OwnerLogin, "drafts")
	}
	if updated.ProviderKey != "" {
		t.Errorf("ProviderKey should still be empty (unchanged): got %q", updated.ProviderKey)
	}

	// 2.5 codex P1 fix — slug 변경 시 application_repositories link cascade update
	linkAppID := "00000000-0000-0000-0000-000000000910"
	defer func() { _, _ = pool.Exec(ctx, `DELETE FROM applications WHERE id = $1`, linkAppID) }()
	if _, err := pool.Exec(ctx, `
INSERT INTO applications (id, key, name, status, visibility, owner_user_id, created_at, updated_at)
VALUES ($1, 'cascade-app', 'Cascade App', 'active', 'internal', 'u-cascade', NOW(), NOW())
ON CONFLICT (id) DO NOTHING
`, linkAppID); err != nil {
		t.Fatalf("seed cascade app: %v", err)
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO application_repositories (application_id, repo_provider, repo_full_name, role, sync_status, link_source, linked_at)
VALUES ($1, 'gitea', $2, 'primary', 'active', 'manual', NOW())
`, linkAppID, newSlug); err != nil {
		t.Fatalf("seed application_repositories link: %v", err)
	}
	defer func() { _, _ = pool.Exec(ctx, `DELETE FROM application_repositories WHERE application_id = $1`, linkAppID) }()
	cascadedSlug := newSlug + "-v2"
	updated, err = pgStore.UpdateRepositoryDraft(ctx, draft.ID, store.RepositoryUpdateDraftParams{Slug: &cascadedSlug})
	if err != nil {
		t.Fatalf("UpdateRepositoryDraft cascade failed: %v", err)
	}
	if updated.FullName != cascadedSlug {
		t.Errorf("slug not updated: %q", updated.FullName)
	}
	var linkCount int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM application_repositories WHERE application_id = $1 AND repo_full_name = $2`, linkAppID, cascadedSlug).Scan(&linkCount); err != nil {
		t.Fatalf("verify cascade: %v", err)
	}
	if linkCount != 1 {
		t.Errorf("application_repositories link not cascaded: count=%d", linkCount)
	}
	var staleCount int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM application_repositories WHERE application_id = $1 AND repo_full_name = $2`, linkAppID, newSlug).Scan(&staleCount); err != nil {
		t.Fatalf("verify stale: %v", err)
	}
	if staleCount != 0 {
		t.Errorf("stale link left behind: count=%d", staleCount)
	}

	// 2.6 codex P1 fix — slug conflict 시 ErrConflict + tx rollback
	conflictAppID := "00000000-0000-0000-0000-000000000911"
	defer func() { _, _ = pool.Exec(ctx, `DELETE FROM applications WHERE id = $1`, conflictAppID) }()
	if _, err := pool.Exec(ctx, `
INSERT INTO applications (id, key, name, status, visibility, owner_user_id, created_at, updated_at)
VALUES ($1, 'conflict-app', 'Conflict App', 'active', 'internal', 'u-conflict', NOW(), NOW())
ON CONFLICT (id) DO NOTHING
`, conflictAppID); err != nil {
		t.Fatalf("seed conflict app: %v", err)
	}
	conflictSlug2 := cascadedSlug + "-conflict"
	if _, err := pool.Exec(ctx, `
INSERT INTO application_repositories (application_id, repo_provider, repo_full_name, role, sync_status, link_source, linked_at)
VALUES ($1, 'gitea', $2, 'sub', 'active', 'manual', NOW())
`, conflictAppID, conflictSlug2); err != nil {
		t.Fatalf("seed conflict link: %v", err)
	}
	defer func() { _, _ = pool.Exec(ctx, `DELETE FROM application_repositories WHERE application_id = $1`, conflictAppID) }()
	if _, err := pgStore.UpdateRepositoryDraft(ctx, draft.ID, store.RepositoryUpdateDraftParams{Slug: &conflictSlug2}); !errors.Is(err, store.ErrConflict) {
		t.Errorf("expected ErrConflict for slug conflict with existing link, got %v", err)
	}
	// conflict 시 tx rollback — draft.FullName 이 cascadedSlug 그대로 유지
	current, err := pgStore.GetRepositoryByID(ctx, draft.ID)
	if err != nil {
		t.Fatalf("verify rollback: %v", err)
	}
	if current.FullName != cascadedSlug {
		t.Errorf("rollback failed: full_name=%q want %q", current.FullName, cascadedSlug)
	}

	// 3. set provider via ProviderID
	giteaID := "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"
	updated, err = pgStore.UpdateRepositoryDraft(ctx, draft.ID, store.RepositoryUpdateDraftParams{
		ProviderID: &giteaID,
	})
	if err != nil {
		t.Fatalf("UpdateRepositoryDraft set provider failed: %v", err)
	}
	if updated.ProviderID != giteaID || updated.ProviderKey != "gitea" {
		t.Errorf("provider not set: provider_id=%q provider_key=%q", updated.ProviderID, updated.ProviderKey)
	}

	// 4. unlink provider (empty string = SET NULL)
	empty := ""
	updated, err = pgStore.UpdateRepositoryDraft(ctx, draft.ID, store.RepositoryUpdateDraftParams{
		ProviderID: &empty,
	})
	if err != nil {
		t.Fatalf("UpdateRepositoryDraft unlink provider failed: %v", err)
	}
	if updated.ProviderID != "" || updated.ProviderKey != "" {
		t.Errorf("provider not unlinked: provider_id=%q provider_key=%q", updated.ProviderID, updated.ProviderKey)
	}

	// 5. unique violation — create second draft, try to set its slug to first's
	slug2 := fmt.Sprintf("drafts/other-%d", time.Now().UnixNano())
	draft2, err := pgStore.CreateRepositoryDraft(ctx, "other-"+fmt.Sprint(time.Now().UnixNano()), slug2, "")
	if err != nil {
		t.Fatalf("CreateRepositoryDraft #2 failed: %v", err)
	}
	defer func() { _, _ = pool.Exec(ctx, `DELETE FROM repositories WHERE id = $1`, draft2.ID) }()
	conflictSlug := newSlug // already used by draft
	if _, err := pgStore.UpdateRepositoryDraft(ctx, draft2.ID, store.RepositoryUpdateDraftParams{Slug: &conflictSlug}); !errors.Is(err, store.ErrConflict) {
		t.Errorf("expected ErrConflict for duplicate slug, got %v", err)
	}

	// 6. status guard — published/active 는 업데이트 불가
	if _, err := pgStore.MarkRepositoryDraftPublishRequested(ctx, draft.ID); err != nil {
		t.Fatalf("MarkRepositoryDraftPublishRequested failed: %v", err)
	}
	otherKey := "AFTER-PUBLISH"
	if _, err := pgStore.UpdateRepositoryDraft(ctx, draft.ID, store.RepositoryUpdateDraftParams{Key: &otherKey}); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound for update on published repo, got %v", err)
	}
}

// TestIntegration_DeleteRepository — draft 만 삭제 가능, FK reference 가드.
func TestIntegration_DeleteRepository(t *testing.T) {
	pgStore, pool, ctx, teardown := setupApplicationsTest(t)
	defer teardown()

	slug := fmt.Sprintf("drafts/del-%d", time.Now().UnixNano())
	key := fmt.Sprintf("del-%d", time.Now().UnixNano())
	draft, err := pgStore.CreateRepositoryDraft(ctx, key, slug, "")
	if err != nil {
		t.Fatalf("CreateRepositoryDraft failed: %v", err)
	}

	// 1. happy path: delete draft
	if err := pgStore.DeleteRepository(ctx, draft.ID); err != nil {
		t.Fatalf("DeleteRepository failed: %v", err)
	}
	if _, err := pgStore.GetRepositoryByID(ctx, draft.ID); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}

	// 2. ghost delete
	if err := pgStore.DeleteRepository(ctx, 999999); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound for ghost delete, got %v", err)
	}

	// 3. status guard — published/active 는 삭제 불가
	slug2 := fmt.Sprintf("drafts/del-pub-%d", time.Now().UnixNano())
	draft2, err := pgStore.CreateRepositoryDraft(ctx, "del-pub-"+fmt.Sprint(time.Now().UnixNano()), slug2, "")
	if err != nil {
		t.Fatalf("CreateRepositoryDraft #2 failed: %v", err)
	}
	defer func() { _, _ = pool.Exec(ctx, `DELETE FROM repositories WHERE id = $1`, draft2.ID) }()
	if _, err := pgStore.MarkRepositoryDraftPublishRequested(ctx, draft2.ID); err != nil {
		t.Fatalf("MarkRepositoryDraftPublishRequested failed: %v", err)
	}
	if err := pgStore.DeleteRepository(ctx, draft2.ID); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound for delete on published repo, got %v", err)
	}

	// 4. FK guard — application_repositories link → ErrConflict
	slug3 := fmt.Sprintf("drafts/del-fk-%d", time.Now().UnixNano())
	draft3, err := pgStore.CreateRepositoryDraft(ctx, "del-fk-"+fmt.Sprint(time.Now().UnixNano()), slug3, "")
	if err != nil {
		t.Fatalf("CreateRepositoryDraft #3 failed: %v", err)
	}
	defer func() { _, _ = pool.Exec(ctx, `DELETE FROM application_repositories WHERE repo_full_name = $1`, slug3) }()
	defer func() { _, _ = pool.Exec(ctx, `DELETE FROM repositories WHERE id = $1`, draft3.ID) }()
	appID := "00000000-0000-0000-0000-000000000999"
	defer func() { _, _ = pool.Exec(ctx, `DELETE FROM applications WHERE id = $1`, appID) }()
	if _, err := pool.Exec(ctx, `
INSERT INTO applications (id, key, name, status, visibility, owner_user_id, created_at, updated_at)
VALUES ($1, $2, $3, 'active', 'internal', 'u-del', NOW(), NOW())
ON CONFLICT (id) DO NOTHING
`, appID, "del-fk-app", "Del FK App"); err != nil {
		t.Fatalf("seed app: %v", err)
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO application_repositories (application_id, repo_provider, repo_full_name, role, sync_status, link_source, linked_at)
VALUES ($1, 'gitea', $2, 'primary', 'active', 'manual', NOW())
`, appID, slug3); err != nil {
		t.Fatalf("seed application_repositories: %v", err)
	}
	if err := pgStore.DeleteRepository(ctx, draft3.ID); !errors.Is(err, store.ErrConflict) {
		t.Errorf("expected ErrConflict for delete with FK link, got %v", err)
	}
}

