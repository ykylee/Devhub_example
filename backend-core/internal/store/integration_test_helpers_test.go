package store_test

import (
	"context"
	"os"
	"testing"

	apprep "github.com/devhub/backend-core/internal/domain/application-lifecycle/repository"
	"github.com/devhub/backend-core/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

// store_test 패키지 자체 integration test helpers. application-lifecycle/repository_test
// 의 setup/fixture 는 다른 패키지이므로 cross-package 호출 불가 — 본 패키지에서
// repository 계층 method 를 호출하는 store_test 가 같은 fixture 의존성을 가지므로
// 본 helper 로 동등한 setup 을 제공한다.

const (
	testAppKey1 = "TESTAPP001"
	testAppKey2 = "TESTAPP002"
	testRepoID1 = int64(99001)
	testRepoID2 = int64(99002)
)

// applicationsFixture — application-lifecycle/repository_test 의 동명 함수와 동일한
// 의미. truncate + test repository 2개 seed.
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

// setupApplicationsTest — store_test 자체 setup. ApplicationRepository wrapper 를
// 반환해 CreateApplication / CreateProject / CreateIntegration 등의 도메인
// repository method 를 호출 가능.
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
	applicationsFixture(t, ctx, pool)
	repo := apprep.NewApplicationRepository(pgStore)
	return repo, pool, ctx, func() {
		pool.Close()
		pgStore.Close()
	}
}
