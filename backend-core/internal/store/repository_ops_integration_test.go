package store_test

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository ops + rollup compute integration test (sprint claude/work_260514-e).
// 핵심: P1 회귀 guard — ComputeApplicationRollup 의 custom_weights 정규화 (sum=1.0
// invariant). PR #108 hotfix 가 정정한 부분이 다시 깨지지 않는지 검증.

func seedRepoOpsData(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	now := time.Now().UTC()
	const seedPR = `
INSERT INTO pr_activities (repository_id, external_pr_id, event_type, actor_login, occurred_at, payload)
VALUES
  ($1, 'PR-001', 'opened',  'alice',   $3, '{}'::jsonb),
  ($1, 'PR-001', 'merged',  'alice',   $4, '{}'::jsonb),
  ($1, 'PR-002', 'opened',  'bob',     $3, '{}'::jsonb),
  ($2, 'PR-010', 'opened',  'charlie', $3, '{}'::jsonb)`
	if _, err := pool.Exec(ctx, seedPR, testRepoID1, testRepoID2, now.Add(-24*time.Hour), now); err != nil {
		t.Fatalf("seed pr_activities: %v", err)
	}
	const seedBuild = `
INSERT INTO build_runs (repository_id, run_external_id, branch, commit_sha, status, duration_seconds, started_at, finished_at)
VALUES
  ($1, 'run-001', 'main', 'abc111', 'success', 120, $3, $4),
  ($1, 'run-002', 'main', 'abc222', 'failed',   60, $3, $4),
  ($2, 'run-010', 'main', 'def111', 'success', 240, $3, $4)`
	if _, err := pool.Exec(ctx, seedBuild, testRepoID1, testRepoID2, now.Add(-1*time.Hour), now); err != nil {
		t.Fatalf("seed build_runs: %v", err)
	}
	const seedQuality = `
INSERT INTO quality_snapshots (repository_id, tool, ref_name, commit_sha, score, gate_passed, metric_payload, measured_at)
VALUES
  ($1, 'sonar', 'main', 'abc222', 85.5, true,  '{}'::jsonb, $3),
  ($2, 'sonar', 'main', 'def111', 75.0, false, '{}'::jsonb, $3)`
	if _, err := pool.Exec(ctx, seedQuality, testRepoID1, testRepoID2, now); err != nil {
		t.Fatalf("seed quality_snapshots: %v", err)
	}
}

func TestIntegration_ListRepositoryActivity(t *testing.T) {
	pgStore, pool, ctx, teardown := setupApplicationsTest(t)
	defer teardown()
	seedRepoOpsData(t, ctx, pool)

	activity, err := pgStore.ListRepositoryActivity(ctx, testRepoID1, store.RepositoryActivityOptions{
		WindowFrom: time.Now().UTC().Add(-48 * time.Hour),
		WindowTo:   time.Now().UTC().Add(1 * time.Hour),
	})
	if err != nil {
		t.Fatalf("activity: %v", err)
	}
	if activity.PREventCount != 3 {
		t.Errorf("pr event count = %d, want 3 (PR-001 opened+merged, PR-002 opened)", activity.PREventCount)
	}
	if activity.BuildRunCount != 2 {
		t.Errorf("build run count = %d, want 2", activity.BuildRunCount)
	}
	if math.Abs(activity.BuildSuccessRate-0.5) > 0.01 {
		t.Errorf("build success rate = %.4f, want ≈0.5 (1 success / 2 total)", activity.BuildSuccessRate)
	}
}

func TestIntegration_ListRepositoryPullRequests_EventFilter(t *testing.T) {
	pgStore, pool, ctx, teardown := setupApplicationsTest(t)
	defer teardown()
	seedRepoOpsData(t, ctx, pool)

	events, total, err := pgStore.ListRepositoryPullRequests(ctx, testRepoID1, store.PRActivityListOptions{
		EventType: "merged",
	})
	if err != nil {
		t.Fatalf("list pr: %v", err)
	}
	if total != 1 || len(events) != 1 || events[0].EventType != "merged" {
		t.Errorf("merged filter result: total=%d, events=%+v", total, events)
	}
}

func TestIntegration_ListRepositoryBuildRuns_StatusFilter(t *testing.T) {
	pgStore, pool, ctx, teardown := setupApplicationsTest(t)
	defer teardown()
	seedRepoOpsData(t, ctx, pool)

	runs, total, err := pgStore.ListRepositoryBuildRuns(ctx, testRepoID1, store.BuildRunListOptions{
		Status: "failed",
	})
	if err != nil {
		t.Fatalf("list runs: %v", err)
	}
	if total != 1 || len(runs) != 1 || runs[0].Status != "failed" {
		t.Errorf("failed filter result: total=%d, runs=%+v", total, runs)
	}
}

func TestIntegration_ListRepositoryQualitySnapshots(t *testing.T) {
	pgStore, pool, ctx, teardown := setupApplicationsTest(t)
	defer teardown()
	seedRepoOpsData(t, ctx, pool)

	snaps, total, err := pgStore.ListRepositoryQualitySnapshots(ctx, testRepoID1, store.QualitySnapshotListOptions{})
	if err != nil {
		t.Fatalf("list quality: %v", err)
	}
	if total != 1 || len(snaps) != 1 || snaps[0].Score == nil || *snaps[0].Score != 85.5 {
		t.Errorf("quality snapshot: total=%d, snaps=%+v", total, snaps)
	}
}

// --- rollup compute ---

// seedAppWith2ActiveLinks 는 testRepoID1/2 를 active 상태로 application 에 연결.
func seedAppWith2ActiveLinks(t *testing.T, ctx context.Context, pgStore *store.PostgresStore, key string) domain.Application {
	t.Helper()
	app, err := pgStore.CreateApplication(ctx, domain.Application{
		Key: key, Name: "X", Status: domain.ApplicationStatusActive,
		Visibility: domain.ApplicationVisibilityInternal, OwnerUserID: "u1",
	})
	if err != nil {
		t.Fatalf("seed app: %v", err)
	}
	if _, err := pgStore.CreateApplicationRepository(ctx, domain.ApplicationRepository{
		ApplicationID: app.ID, RepoProvider: "gitea", RepoFullName: "team/devhub-core",
		Role: domain.ApplicationRepositoryRolePrimary, SyncStatus: domain.SyncStatusActive,
	}); err != nil {
		t.Fatalf("seed link 1: %v", err)
	}
	if _, err := pgStore.CreateApplicationRepository(ctx, domain.ApplicationRepository{
		ApplicationID: app.ID, RepoProvider: "gitea", RepoFullName: "team/devhub-web",
		Role: domain.ApplicationRepositoryRoleSub, SyncStatus: domain.SyncStatusActive,
	}); err != nil {
		t.Fatalf("seed link 2: %v", err)
	}
	return app
}

func TestIntegration_ComputeApplicationRollup_Equal(t *testing.T) {
	pgStore, pool, ctx, teardown := setupApplicationsTest(t)
	defer teardown()
	seedRepoOpsData(t, ctx, pool)
	app := seedAppWith2ActiveLinks(t, ctx, pgStore, testAppKey1)

	rollup, err := pgStore.ComputeApplicationRollup(ctx, app.ID, domain.ApplicationRollupOptions{
		Policy:     domain.WeightPolicyEqual,
		WindowFrom: time.Now().UTC().Add(-48 * time.Hour),
		WindowTo:   time.Now().UTC().Add(1 * time.Hour),
	})
	if err != nil {
		t.Fatalf("rollup: %v", err)
	}
	if len(rollup.Meta.AppliedWeights) != 2 {
		t.Fatalf("expected 2 weighted repos, got %+v", rollup.Meta.AppliedWeights)
	}
	sum := 0.0
	for _, w := range rollup.Meta.AppliedWeights {
		sum += w
	}
	if math.Abs(sum-1.0) > 0.001 {
		t.Errorf("equal sum = %.4f, want 1.0 ±0.001", sum)
	}
	for repo, w := range rollup.Meta.AppliedWeights {
		if math.Abs(w-0.5) > 0.001 {
			t.Errorf("repo %s weight = %.4f, want 0.5", repo, w)
		}
	}
}

// P1 codex review 회귀 guard — custom_weights 가 일부 repo 만 cover 할 때
// 누락 repo 의 equal fallback 후 전체 sum 이 1.0 으로 정규화되는지 검증.
//
// 시나리오: contributing 에 team/devhub-core 와 team/devhub-web 2개. custom_weights
// 가 team/devhub-core 에만 weight=1.0 부여. 정규화 전 합 = 1.0 + 0.5 = 1.5.
// hotfix #108 의 정정으로 합이 1.0 으로 정규화되어야 함.
func TestIntegration_ComputeApplicationRollup_CustomNormalization_P1(t *testing.T) {
	pgStore, pool, ctx, teardown := setupApplicationsTest(t)
	defer teardown()
	seedRepoOpsData(t, ctx, pool)
	app := seedAppWith2ActiveLinks(t, ctx, pgStore, testAppKey1)

	rollup, err := pgStore.ComputeApplicationRollup(ctx, app.ID, domain.ApplicationRollupOptions{
		Policy: domain.WeightPolicyCustom,
		CustomWeights: map[string]float64{
			"team/devhub-core": 1.0, // sum check: exact 1.0
			// team/devhub-web: missing → equal fallback
		},
		WindowFrom: time.Now().UTC().Add(-48 * time.Hour),
		WindowTo:   time.Now().UTC().Add(1 * time.Hour),
	})
	if err != nil {
		t.Fatalf("rollup: %v", err)
	}
	sum := 0.0
	for _, w := range rollup.Meta.AppliedWeights {
		sum += w
	}
	if math.Abs(sum-1.0) > 0.001 {
		t.Errorf("PR #108 P1 회귀: custom fallback 후 sum = %.4f, want 1.0 ±0.001 (정규화 부재 시 1.5)", sum)
	}
	// fallbacks meta 의 AppliedWeight 도 정규화 후 값이어야 함.
	if len(rollup.Meta.Fallbacks) != 1 {
		t.Fatalf("expected 1 fallback (team/devhub-web missing), got %+v", rollup.Meta.Fallbacks)
	}
	fb := rollup.Meta.Fallbacks[0]
	if fb.RepoFullName != "team/devhub-web" || fb.Reason != "custom_weight_missing" {
		t.Errorf("fallback metadata: %+v", fb)
	}
	if math.Abs(fb.AppliedWeight-rollup.Meta.AppliedWeights["team/devhub-web"]) > 0.0001 {
		t.Errorf("fallback applied_weight = %.4f vs map = %.4f (P1 정정: 정규화 후 값으로 갱신되어야)",
			fb.AppliedWeight, rollup.Meta.AppliedWeights["team/devhub-web"])
	}
}

func TestIntegration_ComputeApplicationRollup_RepoRoleNormalize(t *testing.T) {
	pgStore, pool, ctx, teardown := setupApplicationsTest(t)
	defer teardown()
	seedRepoOpsData(t, ctx, pool)
	app := seedAppWith2ActiveLinks(t, ctx, pgStore, testAppKey1)
	// primary + sub 의 카테고리 정규화: primary 0.6, sub 0.3, shared 없으면 shared 0.1
	// 을 다른 카테고리에 비례 재분배 → primary=0.6/(0.9), sub=0.3/(0.9) → 약 0.667 / 0.333.

	rollup, err := pgStore.ComputeApplicationRollup(ctx, app.ID, domain.ApplicationRollupOptions{
		Policy:     domain.WeightPolicyRepoRole,
		WindowFrom: time.Now().UTC().Add(-48 * time.Hour),
		WindowTo:   time.Now().UTC().Add(1 * time.Hour),
	})
	if err != nil {
		t.Fatalf("rollup: %v", err)
	}
	sum := 0.0
	for _, w := range rollup.Meta.AppliedWeights {
		sum += w
	}
	if math.Abs(sum-1.0) > 0.001 {
		t.Errorf("repo_role sum = %.4f, want 1.0 ±0.001", sum)
	}
	primary := rollup.Meta.AppliedWeights["team/devhub-core"]
	sub := rollup.Meta.AppliedWeights["team/devhub-web"]
	if primary <= sub {
		t.Errorf("primary (%.4f) should be > sub (%.4f) after redistribution", primary, sub)
	}
}

func TestIntegration_CountApplicationCriticalWarnings(t *testing.T) {
	pgStore, pool, ctx, teardown := setupApplicationsTest(t)
	defer teardown()
	seedRepoOpsData(t, ctx, pool)
	app := seedAppWith2ActiveLinks(t, ctx, pgStore, testAppKey1)

	count, err := pgStore.CountApplicationCriticalWarnings(ctx, app.ID)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	// seedRepoOpsData 에 따라:
	//   - quality_snapshots: testRepoID1 gate_passed=true / testRepoID2 gate_passed=false → 1건
	//   - build success rate: testRepoID1 0.5 / testRepoID2 1.0 → weighted ≈ 0.75 (>= 0.5)
	// 따라서 critical = 1 (gate failed) + 0 (build rate ok) = 1
	if count != 1 {
		t.Errorf("critical count = %d, want 1 (gate_passed=false 1건)", count)
	}
}
