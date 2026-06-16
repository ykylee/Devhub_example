package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/devhub/backend-core/internal/domain"
)

// --- Platform sub-project rollup (Sprint C — kpi-tests-per-domain-scope.md §2.3 + §6.3) ---

// ComputePlatformWeightedKPI 는 platform 의 N개 sub-project 의 가중치 적용
// metric 을 equal average 로 종합한 PlatformWeightedKPI 를 반환.
//
// 정공법 (Sprint B 의 ProjectWeightedKPI 와 정합 — 단, weight 정공법이 다름):
//   - Sprint B: sub-repo (linked repository) 단위 contribution_weight 적용
//   - Sprint C: sub-project (linked project) 단위 **equal** average (sub-project 균등)
//   - 2-depth 가중치: 각 sub-project 의 모든 linked repository 의 raw metric 을
//     1차 종합 (quality / build / pr), 그 결과를 sub-project 균등 평균.
//
// CTE 4개 + LATERAL JOIN 3개. 단일 round-trip 정공법 (Sprint B-Tests 의 ListProjectTestResults
// 와 정합). ci_runs + quality_snapshots + pr_activities 의 read-only 집계 —
// build_runs table 미사용 (ci_runs 가 write path, ISSUES-04/P1-7).
func (s *PostgresStore) ComputePlatformWeightedKPI(ctx context.Context, platformID string, opts BuildRunListOptions) (domain.PlatformWeightedKPI, error) {
	if opts.WindowFrom.IsZero() {
		opts.WindowFrom = time.Now().UTC().AddDate(0, 0, -30)
	}
	if opts.WindowTo.IsZero() {
		opts.WindowTo = time.Now().UTC()
	}

	const query = `
WITH linked_projects AS (
  SELECT p.id, p.contribution_weight
  FROM projects p WHERE p.platform_id = $1::uuid
),
per_project_metric AS (
  SELECT
    lp.id, lp.contribution_weight,
    COALESCE(latest_quality.score, 0) AS quality_score,
    COALESCE(agg.build_success_rate, 0) AS build_success_rate,
    COALESCE(agg.build_run_count, 0) AS build_run_count,
    COALESCE(agg.active_contributor_count, 0) AS active_contributor_count,
    COALESCE(pr_stats.open_count, 0) AS open_count,
    COALESCE(pr_stats.merged_count, 0) AS merged_count
  FROM linked_projects lp
  LEFT JOIN LATERAL (
    SELECT score FROM quality_snapshots qs
    WHERE qs.repository_id IN (
      SELECT pr.repository_id FROM project_repositories pr WHERE pr.project_id = lp.id
    )
    ORDER BY measured_at DESC LIMIT 1
  ) latest_quality ON true
  LEFT JOIN LATERAL (
    SELECT
      COALESCE(COUNT(*) FILTER (WHERE br.status = 'success')::float / NULLIF(COUNT(*), 0), 0) AS build_success_rate,
      COUNT(*)::int AS build_run_count,
      COUNT(DISTINCT br.commit_author)::int AS active_contributor_count
    FROM ci_runs br
    WHERE br.repository_id IN (
      SELECT pr.repository_id FROM project_repositories pr WHERE pr.project_id = lp.id
    )
      AND ($2::timestamptz IS NULL OR COALESCE(br.started_at, br.created_at) >= $2)
      AND ($3::timestamptz IS NULL OR COALESCE(br.started_at, br.created_at) < $3)
  ) agg ON true
  LEFT JOIN LATERAL (
    SELECT
      COUNT(DISTINCT pa.number) FILTER (WHERE pa.event_type = 'opened')::int AS open_count,
      COUNT(DISTINCT pa.number) FILTER (WHERE pa.event_type = 'merged')::int AS merged_count
    FROM pr_activities pa
    WHERE pa.repository_id IN (
      SELECT pr.repository_id FROM project_repositories pr WHERE pr.project_id = lp.id
    )
      AND pa.occurred_at >= $2 AND pa.occurred_at < $3
  ) pr_stats ON true
)
SELECT
  COUNT(*)::int AS linked_project_count,
  COALESCE(AVG(quality_score), 0)::float8 AS weighted_quality_score,
  COALESCE(AVG(build_success_rate), 0)::float8 AS weighted_build_success_rate,
  COALESCE(SUM(build_run_count), 0)::int AS total_build_run_count,
  COALESCE(SUM(active_contributor_count), 0)::int AS total_active_contributors,
  COALESCE(SUM(open_count), 0)::int AS total_open_pr,
  COALESCE(SUM(merged_count), 0)::int AS total_merged_pr
FROM per_project_metric`

	var (
		linkedProjectCount   int
		weightedQuality      float64
		weightedBSR          float64
		totalBuildRuns       int
		totalContributors    int
		totalOpenPR          int
		totalMergedPR        int
	)
	wf := windowFromOrNil(opts.WindowFrom)
	wt := windowToOrNil(opts.WindowTo)
	err := s.pool.QueryRow(ctx, query, platformID, wf, wt).Scan(
		&linkedProjectCount, &weightedQuality, &weightedBSR,
		&totalBuildRuns, &totalContributors, &totalOpenPR, &totalMergedPR,
	)
	if err != nil {
		return domain.PlatformWeightedKPI{}, fmt.Errorf("compute platform weighted kpi: %w", err)
	}

	return domain.PlatformWeightedKPI{
		PlatformID:             platformID,
		WindowFrom:             opts.WindowFrom.UTC(),
		WindowTo:               opts.WindowTo.UTC(),
		WeightedQualityScore:   weightedQuality,
		WeightedBuildSuccess:   weightedBSR,
		TotalBuildRunCount:     totalBuildRuns,
		OpenPRCount:            totalOpenPR,
		MergedPRCount:          totalMergedPR,
		ActiveContributorCount: totalContributors,
		LinkedProjectCount:     linkedProjectCount,
		WeightedAt:             time.Now().UTC(),
	}, nil
}

// ListPlatformTestResults 는 platform 의 N개 sub-project 의 build_runs status 를
// 종합한 PlatformWeightedTestResults 를 반환. weighted_pass_rate 는 sub-project
// 별 pass_rate 의 equal average (sub-project 균등).
//
// CTE 3개 + LATERAL JOIN. 단일 round-trip. ci_runs table 사용 (Sprint B-Tests
// ListProjectTestResults 와 동일 정공법).
func (s *PostgresStore) ListPlatformTestResults(ctx context.Context, platformID string, opts BuildRunListOptions) (domain.PlatformWeightedTestResults, int, error) {
	if opts.WindowFrom.IsZero() {
		opts.WindowFrom = time.Now().UTC().AddDate(0, 0, -30)
	}
	if opts.WindowTo.IsZero() {
		opts.WindowTo = time.Now().UTC()
	}
	limit := opts.Limit
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	const query = `
WITH linked_projects AS (
  SELECT p.id FROM projects p WHERE p.platform_id = $1::uuid
),
per_project AS (
  SELECT
    lp.id,
    COALESCE(stats.success_count::float / NULLIF(stats.success_count + stats.failed_count, 0), 0) AS pass_rate
  FROM linked_projects lp
  LEFT JOIN LATERAL (
    SELECT
      COUNT(*) FILTER (WHERE br.status = 'success')::int AS success_count,
      COUNT(*) FILTER (WHERE br.status = 'failed')::int AS failed_count
    FROM ci_runs br
    WHERE br.repository_id IN (
      SELECT pr.repository_id FROM project_repositories pr WHERE pr.project_id = lp.id
    )
      AND ($4::timestamptz IS NULL OR COALESCE(br.started_at, br.created_at) >= $4)
      AND ($5::timestamptz IS NULL OR COALESCE(br.started_at, br.created_at) < $5)
  ) stats ON true
),
totals AS (
  SELECT br.status, COUNT(*)::int AS cnt
  FROM ci_runs br
  WHERE br.repository_id IN (
    SELECT pr.repository_id FROM project_repositories pr
    WHERE pr.project_id IN (SELECT id FROM linked_projects)
  )
    AND ($4::timestamptz IS NULL OR COALESCE(br.started_at, br.created_at) >= $4)
    AND ($5::timestamptz IS NULL OR COALESCE(br.started_at, br.created_at) < $5)
  GROUP BY br.status
),
recent AS (
  SELECT
    br.id, p.id AS project_id, p.name AS project_full_name,
    br.repository_id, r.full_name AS repository_full_name,
    br.external_id AS run_external_id, br.branch,
    COALESCE(br.commit_sha, '') AS commit_sha,
    br.status, br.duration_seconds,
    COALESCE(br.started_at, br.created_at) AS started_at, br.finished_at
  FROM ci_runs br
  JOIN repositories r ON r.id = br.repository_id
  JOIN project_repositories pr ON pr.repository_id = br.repository_id
  JOIN projects p ON p.id = pr.project_id
  WHERE p.id IN (SELECT id FROM linked_projects)
    AND ($4::timestamptz IS NULL OR COALESCE(br.started_at, br.created_at) >= $4)
    AND ($5::timestamptz IS NULL OR COALESCE(br.started_at, br.created_at) < $5)
  ORDER BY COALESCE(br.started_at, br.created_at) DESC
  LIMIT $3
),
total_count AS (
  SELECT COUNT(*)::int AS cnt
  FROM ci_runs br
  WHERE br.repository_id IN (
    SELECT pr.repository_id FROM project_repositories pr
    WHERE pr.project_id IN (SELECT id FROM linked_projects)
  )
    AND ($4::timestamptz IS NULL OR COALESCE(br.started_at, br.created_at) >= $4)
    AND ($5::timestamptz IS NULL OR COALESCE(br.started_at, br.created_at) < $5)
)
SELECT
  (SELECT AVG(pass_rate)::float8 FROM per_project) AS weighted_pass_rate,
  (SELECT json_object_agg(status, cnt) FROM totals) AS totals_json,
  (SELECT json_agg(json_build_object(
    'id', id, 'project_id', project_id, 'project_full_name', project_full_name,
    'repository_id', repository_id, 'repository_full_name', repository_full_name,
    'run_external_id', run_external_id, 'branch', branch, 'commit_sha', commit_sha,
    'status', status, 'duration_seconds', duration_seconds,
    'started_at', started_at, 'finished_at', finished_at
  )) FROM recent) AS recent_json,
  (SELECT cnt FROM total_count) AS total_count`

	var (
		weightedPassRate *float64
		totalsJSON       []byte
		recentJSON       []byte
		totalCount       int
	)
	wf := windowFromOrNil(opts.WindowFrom)
	wt := windowToOrNil(opts.WindowTo)
	err := s.pool.QueryRow(ctx, query, platformID, wf, limit, wf, wt).Scan(
		&weightedPassRate, &totalsJSON, &recentJSON, &totalCount,
	)
	if err != nil {
		return domain.PlatformWeightedTestResults{}, 0, fmt.Errorf("list platform test results: %w", err)
	}

	// totals 정규화 — 7 status 모두 0 초기화 + store 의 raw status merge.
	// Sprint A 의 RepositoryTestsSection + Sprint B-Tests 정공법과 동일.
	totals := map[string]int{
		"success":   0,
		"failed":    0,
		"running":   0,
		"cancelled": 0,
		"skipped":   0,
		"queued":    0,
		"unknown":   0,
	}
	if len(totalsJSON) > 0 {
		var rawTotals map[string]int
		if err := json.Unmarshal(totalsJSON, &rawTotals); err != nil {
			return domain.PlatformWeightedTestResults{}, 0, fmt.Errorf("decode totals json: %w", err)
		}
		for k, v := range rawTotals {
			if _, ok := totals[k]; ok {
				totals[k] = v
			} else {
				totals["unknown"] += v
			}
		}
	}

	var recent []domain.PlatformBuildRun
	if len(recentJSON) > 0 {
		if err := json.Unmarshal(recentJSON, &recent); err != nil {
			return domain.PlatformWeightedTestResults{}, 0, fmt.Errorf("decode recent json: %w", err)
		}
	}

	return domain.PlatformWeightedTestResults{
		PlatformID:       platformID,
		WindowFrom:       opts.WindowFrom.UTC(),
		WindowTo:         opts.WindowTo.UTC(),
		WeightedPassRate: weightedPassRate,
		Totals:           totals,
		Recent:           recent,
	}, totalCount, nil
}
