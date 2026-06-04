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
               platform_repositories, applications,
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

// applicationsFixtureLockID — `go test ./...` cross-pkg race guard.
//
// 본 패키지 (store_test) 와 internal/domain/application-lifecycle/repository_test
// 양쪽이 동일 테이블 (applications, projects, project_integrations, platform_repositories,
// project_members, repositories) 을 TRUNCATE CASCADE 하는 fixture 를 동시 실행한다.
// `go test ./...` 는 패키지 단위로 별도 binary 를 병렬 실행하므로, A 패키지의 test
// 진행 중에 B 패키지 fixture 가 끼어들면 application row 가 CASCADE 로 사라지고
// 후속 INSERT (예: CreateIntegration) 가 FK violation 으로 silent fail 한다. 결과적으로
// `created.ID=""` 가 `DeleteIntegration("")` 로 흘러 SQLSTATE 22P02 ("uuid 잘못된
// 입력") 회귀가 표면화된다 (PR #437 직전 sprint case).
//
// 양쪽 setup 이 동일한 `pg_advisory_lock` id 를 잡아 fixture+test lifecycle 을
// 시리얼라이즈한다. 같은 conn 으로 session lock 잡고 teardown 에서 release.
//
// 본 상수가 변경되면 다음 파일의 대응 상수도 같이 변경해야 한다:
//   - internal/domain/application-lifecycle/repository/applications_integration_test.go
const applicationsFixtureLockID = int64(0x4150705F4C6966) // "App_Lif" ASCII

// setupApplicationsTest — store_test 자체 setup. PlatformRepository wrapper 를 wrapper 를
// 반환해 CreatePlatform / CreateProject / CreateIntegration 등의 도메인
// repository method 를 호출 가능.
func setupApplicationsTest(t *testing.T) (*apprep.PlatformRepository, *pgxpool.Pool, context.Context, func()) {
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
	repo := apprep.NewPlatformRepository(pgStore)
	return repo, pool, ctx, func() {
		_, _ = lockConn.Exec(ctx, `SELECT pg_advisory_unlock($1)`, applicationsFixtureLockID)
		lockConn.Release()
		pool.Close()
		pgStore.Close()
	}
}
