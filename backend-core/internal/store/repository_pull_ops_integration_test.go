package store

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// X-5 production wire follow-up integration test (TC-GITEA-PULL-STORE-01).
//
// CI backend-unit job 은 DEVHUB_TEST_DB_URL 미설정으로 t.Skip — store 패키지 정합 패턴.
// setup 자체는 inline (cross-package unexported helper 사용 불가).
//
// 검증 범위:
//   - repository_pull_state 6 method (UpdatePullState, IncrementConsecutiveFailures,
//     ResetConsecutiveFailures, SetBackoff, BackoffUntil, LastPullAt)
//   - pr_activities / build_runs / quality_snapshots upsert 3 method
//   - ListGiteaPullTargets query 의 filter (provider_type='scm' + key='gitea' +
//     status='active' + gitea_repository_id IS NOT NULL + backoff filtering)
//
// 시드 cleanup: 각 test 가 자체 row 를 만들며 test 끝에서 repository_pull_state /
// pr_activities / build_runs / quality_snapshots 의 해당 repository_id 의 row 만
// cleanup. repositories / integration_providers 는 cleanup 안 함 (다른 test 와 공유
// 가능). 그러나 본 file 의 test 는 unique repository 를 생성해 격리.

// setupPullOpsTest creates a PostgresStore + a unique test repository (gitea provider
// + repositories row) and returns them. The repository is registered with
// gitea_repository_id in [100000, 200000) to avoid collisions with seed data.
func setupPullOpsTest(t *testing.T) (*PostgresStore, *pgxpool.Pool, context.Context, int64, string) {
	t.Helper()
	dbURL := os.Getenv("DEVHUB_TEST_DB_URL")
	if dbURL == "" {
		t.Skip("DEVHUB_TEST_DB_URL is not set")
	}
	ctx := context.Background()
	pgStore, err := NewPostgresStore(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect postgres: %v", err)
	}
	pool := pgStore.Pool()

	// Look up an existing Gitea scm provider (created by setupApplicationsTest or
	// application seeding in TestPostgresStoreUpsertRepoOps). When the CI
	// Backend Integration Tests job runs only this test package (`./internal/store/...`),
	// TestPostgresStoreUpsertRepoOps may not have run yet → integration_providers
	// has no Gitea row → t.Skip the test (defensive — sibling test owns the seed).
	//
	// Sprint A 정공법: 통합 test 의 사전 seed 부재 시 t.Skip (cross-test pollution 회피).
	// local 에서 본 PR 의 3개 test 만 단독 실행할 때도 Gitea provider 가 없는 환경이면 skip.
	var providerID string
	err = pool.QueryRow(ctx, `
SELECT provider_id::text
FROM integration_providers
WHERE provider_type = 'scm' AND provider_key = 'gitea'
LIMIT 1`).Scan(&providerID)
	if err != nil {
		t.Skipf("Gitea scm provider seed missing (run TestPostgresStoreUpsertRepoOps first to seed integration_providers): %v", err)
	}

	// Create a unique repository for this test. Use a uniquified full_name and
	// gitea_repository_id derived from the current nanosecond.
	suffix := fmt.Sprintf("pull-ops-%d", time.Now().UnixNano())
	var repoID int64
	err = pool.QueryRow(ctx, `
INSERT INTO repositories (
    full_name, name, owner_login, gitea_repository_id, default_branch,
    repository_status, provider_id, source
) VALUES (
    $1, $2, 'test-org', $3, 'main', 'active', $4::uuid, 'gitea'
)
RETURNING id`, suffix, suffix, time.Now().UnixNano()%1_000_000+100_000, providerID).Scan(&repoID)
	if err != nil {
		t.Fatalf("create test repository: %v", err)
	}

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM quality_snapshots WHERE repository_id = $1`, repoID)
		_, _ = pool.Exec(ctx, `DELETE FROM build_runs WHERE repository_id = $1`, repoID)
		_, _ = pool.Exec(ctx, `DELETE FROM pr_activities WHERE repository_id = $1`, repoID)
		_, _ = pool.Exec(ctx, `DELETE FROM repository_pull_state WHERE repository_id = $1`, repoID)
		_, _ = pool.Exec(ctx, `DELETE FROM repositories WHERE id = $1`, repoID)
		pgStore.Close()
	})

	return pgStore, pool, ctx, repoID, providerID
}

func TestIntegration_RepositoryPullState_AllMethods(t *testing.T) {
	pgStore, _, ctx, repoID, _ := setupPullOpsTest(t)
	repoIDStr := fmt.Sprintf("%d", repoID)

	// LastPullAt (cold start, no row)
	at, err := pgStore.LastPullAt(ctx, repoIDStr)
	if err != nil {
		t.Fatalf("LastPullAt cold start: %v", err)
	}
	if !at.IsZero() {
		t.Errorf("LastPullAt cold start should be zero, got %v", at)
	}

	// BackoffUntil (cold start, no row)
	bu, err := pgStore.BackoffUntil(ctx, repoIDStr)
	if err != nil {
		t.Fatalf("BackoffUntil cold start: %v", err)
	}
	if !bu.IsZero() {
		t.Errorf("BackoffUntil cold start should be zero, got %v", bu)
	}

	// UpdatePullState (success)
	now := time.Now().UTC()
	if err := pgStore.UpdatePullState(ctx, repoIDStr, "success", "", now); err != nil {
		t.Fatalf("UpdatePullState success: %v", err)
	}
	at, err = pgStore.LastPullAt(ctx, repoIDStr)
	if err != nil || at.IsZero() {
		t.Errorf("LastPullAt after success should be non-zero: %v %v", at, err)
	}

	// IncrementConsecutiveFailures (3회)
	for i := 1; i <= 3; i++ {
		n, err := pgStore.IncrementConsecutiveFailures(ctx, repoIDStr)
		if err != nil {
			t.Fatalf("Increment #%d: %v", i, err)
		}
		if n != i {
			t.Errorf("Increment #%d = %d; want %d", i, n, i)
		}
	}

	// SetBackoff
	until := time.Now().UTC().Add(2 * time.Hour)
	if err := pgStore.SetBackoff(ctx, repoIDStr, until); err != nil {
		t.Fatalf("SetBackoff: %v", err)
	}
	bu, err = pgStore.BackoffUntil(ctx, repoIDStr)
	if err != nil {
		t.Fatalf("BackoffUntil after SetBackoff: %v", err)
	}
	if bu.IsZero() || bu.Unix() != until.Unix() {
		t.Errorf("BackoffUntil = %v; want ~%v", bu, until)
	}

	// UpdatePullState (error with errMsg)
	if err := pgStore.UpdatePullState(ctx, repoIDStr, "error", "test_error", time.Time{}); err != nil {
		t.Fatalf("UpdatePullState error: %v", err)
	}

	// ResetConsecutiveFailures
	if err := pgStore.ResetConsecutiveFailures(ctx, repoIDStr); err != nil {
		t.Fatalf("ResetConsecutiveFailures: %v", err)
	}
	// After reset, IncrementConsecutiveFailures should start from 1 (cold insert path).
	// Actually the row exists, so ON CONFLICT path increments current (0) → 1.
	n, err := pgStore.IncrementConsecutiveFailures(ctx, repoIDStr)
	if err != nil {
		t.Fatalf("Increment after reset: %v", err)
	}
	if n != 1 {
		t.Errorf("Increment after reset = %d; want 1", n)
	}
	// BackoffUntil should also be cleared by ResetConsecutiveFailures.
	bu, err = pgStore.BackoffUntil(ctx, repoIDStr)
	if err != nil {
		t.Fatalf("BackoffUntil after reset: %v", err)
	}
	if !bu.IsZero() {
		t.Errorf("BackoffUntil after reset = %v; want zero", bu)
	}
}

func TestIntegration_RepositoryPullIngest_UpsertPRBuildQuality(t *testing.T) {
	pgStore, _, ctx, repoID, _ := setupPullOpsTest(t)
	repoIDStr := fmt.Sprintf("%d", repoID)

	// UpsertPullActivity — adapter 가 state="closed" + merged=true → "merged" 결정
	now := time.Now().UTC()
	if err := pgStore.UpsertPullActivity(ctx, repoIDStr, 1234, 42, "merged", "title", "body", "deadbeef", "alice", now); err != nil {
		t.Fatalf("UpsertPullActivity: %v", err)
	}
	// upsert same key again — should not error (ON CONFLICT DO UPDATE)
	if err := pgStore.UpsertPullActivity(ctx, repoIDStr, 1234, 42, "merged", "title-updated", "body", "deadbeef", "alice", now); err != nil {
		t.Fatalf("UpsertPullActivity idempotent: %v", err)
	}

	// UpsertBuildRun
	if err := pgStore.UpsertBuildRun(ctx, repoIDStr, 5678, "feedface", "pull_request", "success", "success", now); err != nil {
		t.Fatalf("UpsertBuildRun: %v", err)
	}

	// UpsertQualitySnapshot
	if err := pgStore.UpsertQualitySnapshot(ctx, repoIDStr, "feedface", now); err != nil {
		t.Fatalf("UpsertQualitySnapshot: %v", err)
	}
	// Idempotent upsert
	if err := pgStore.UpsertQualitySnapshot(ctx, repoIDStr, "feedface", now.Add(time.Hour)); err != nil {
		t.Fatalf("UpsertQualitySnapshot idempotent: %v", err)
	}

	// verify rows exist via ListRepositoryPullRequests / ListRepositoryBuildRuns /


	builds, btotal, err := pgStore.ListRepositoryBuildRuns(ctx, repoID, BuildRunListOptions{})
	if err != nil {
		t.Fatalf("ListRepositoryBuildRuns: %v", err)
	}
	if btotal != 1 || len(builds) != 1 || builds[0].Status != "success" {
		t.Errorf("build_runs = %+v (total=%d); want 1 row with status=success", builds, btotal)
	}

	snaps, stotal, err := pgStore.ListRepositoryQualitySnapshots(ctx, repoID, QualitySnapshotListOptions{Tool: "gitea-build"})
	if err != nil {
		t.Fatalf("ListRepositoryQualitySnapshots: %v", err)
	}
	if stotal != 1 || len(snaps) != 1 || snaps[0].Tool != "gitea-build" {
		t.Errorf("quality_snapshots = %+v (total=%d); want 1 row with tool=gitea-build", snaps, stotal)
	}
}

func TestIntegration_ListGiteaPullTargets_Filter(t *testing.T) {
	pgStore, _, ctx, repoID, _ := setupPullOpsTest(t)
	repoIDStr := fmt.Sprintf("%d", repoID)

	// Our test repo is created with gitea provider and active status and
	// non-null gitea_repository_id → it should appear in the listing.
	targets, err := pgStore.ListGiteaPullTargets(ctx)
	if err != nil {
		t.Fatalf("ListGiteaPullTargets: %v", err)
	}
	found := false
	for _, tgt := range targets {
		if tgt.ID == repoID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("test repoID %d not in lister output (%d targets)", repoID, len(targets))
	}

	// Now set a backoff_until in the future and verify the lister excludes it.
	futureBackoff := time.Now().UTC().Add(2 * time.Hour)
	if err := pgStore.SetBackoff(ctx, repoIDStr, futureBackoff); err != nil {
		t.Fatalf("SetBackoff: %v", err)
	}
	targets2, err := pgStore.ListGiteaPullTargets(ctx)
	if err != nil {
		t.Fatalf("ListGiteaPullTargets (backoff set): %v", err)
	}
	for _, tgt := range targets2 {
		if tgt.ID == repoID {
			t.Errorf("test repoID %d should be excluded due to backoff_until=%v, but appeared in lister", repoID, futureBackoff)
		}
	}

	// BackoffUntil should match what we set (within ±1s tolerance).
	bu, err := pgStore.BackoffUntil(ctx, repoIDStr)
	if err != nil {
		t.Fatalf("BackoffUntil: %v", err)
	}
	if bu.IsZero() {
		t.Errorf("BackoffUntil should be set after SetBackoff")
	}
	if absDuration(bu.Sub(futureBackoff)) > 2*time.Second {
		t.Errorf("BackoffUntil = %v; want ~%v (±2s)", bu, futureBackoff)
	}
}

// absDuration returns the absolute value of a duration (helper for time comparison).
func absDuration(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}
