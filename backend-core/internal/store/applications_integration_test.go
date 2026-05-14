package store_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/devhub/backend-core/internal/domain"
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
               pr_activities, build_runs, quality_snapshots RESTART IDENTITY CASCADE;`
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

func setupApplicationsTest(t *testing.T) (*store.PostgresStore, *pgxpool.Pool, context.Context, func()) {
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
	applicationsFixture(t, ctx, pool)
	return pgStore, pool, ctx, func() {
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
