package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/devhub/backend-core/internal/domain"
)

// --- Project 가중치 적용 test results (Sprint B-Tests — kpi-tests-per-domain-scope.md §2.2 follow-up) ---

// ListProjectTestResults 는 project 의 N개 linked repository 의 build_runs
// status 를 종합한 ProjectWeightedTestResults 를 반환.
//
// 정공법 (CTE 3개, 단일 round-trip):
//   - linked_repos: project_repositories + contribution_weight
//   - weighted: Σ(success × weight) / Σ((success+failed) × weight) — denom=0
//     인 경우 weightedPassRate = nil (linked repo 0 또는 모든 build 가
//     queued/running/cancelled/skipped/unknown).
//   - totals_rows: build_runs.status 별 count (가중치 무관 단순 합산).
//   - recent_rows: 모든 linked repo 의 build_runs 최신순 limit (multi-repo,
//     repository_full_name 추가).
//
// opts.Limit = recent limit (default 20, max 50 — handler 측 가드 정합).
// opts.WindowFrom/WindowTo 가 zero 면 default 30d (Sprint A 정공법 정합).
// opts.Status/Branch/Offset 는 본 endpoint 에서 무시 (multi-repo 통합 응답).
func (s *PostgresStore) ListProjectTestResults(ctx context.Context, projectID string, opts BuildRunListOptions) (domain.ProjectWeightedTestResults, int, error) {
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

	// CTE 3개 + json_agg. ci_runs 를 ListRepositoryBuildRuns 와 동일하게 사용
	// (ISSUES-04/P1-7 — build_runs 는 ingest pipeline 미보유, ci_runs 가 write path).
	const query = `
WITH linked_repos AS (
  SELECT pr.repository_id, pr.contribution_weight
  FROM project_repositories pr
  WHERE pr.project_id = $1::uuid
),
weighted AS (
  SELECT
    COALESCE(SUM(stats.success_count * lr.contribution_weight), 0)::float8 AS num,
    COALESCE(SUM((stats.success_count + stats.failed_count) * lr.contribution_weight), 0)::float8 AS denom
  FROM linked_repos lr
  LEFT JOIN LATERAL (
    SELECT
      COUNT(*) FILTER (WHERE br.status = 'success')::int AS success_count,
      COUNT(*) FILTER (WHERE br.status = 'failed')::int AS failed_count
    FROM ci_runs br
    WHERE br.repository_id = lr.repository_id
      AND ($4::timestamptz IS NULL OR COALESCE(br.started_at, br.created_at) >= $4)
      AND ($5::timestamptz IS NULL OR COALESCE(br.started_at, br.created_at) < $5)
  ) stats ON true
),
totals_rows AS (
  SELECT br.status, COUNT(*)::int AS cnt
  FROM ci_runs br
  WHERE br.repository_id IN (SELECT repository_id FROM linked_repos)
    AND ($4::timestamptz IS NULL OR COALESCE(br.started_at, br.created_at) >= $4)
    AND ($5::timestamptz IS NULL OR COALESCE(br.started_at, br.created_at) < $5)
  GROUP BY br.status
),
recent_rows AS (
  SELECT
    br.id, br.repository_id, r.full_name AS repository_full_name,
    br.external_id AS run_external_id, br.branch,
    COALESCE(br.commit_sha, '') AS commit_sha,
    br.status, br.duration_seconds,
    COALESCE(br.started_at, br.created_at) AS started_at, br.finished_at
  FROM ci_runs br
  JOIN repositories r ON r.id = br.repository_id
  WHERE br.repository_id IN (SELECT repository_id FROM linked_repos)
    AND ($4::timestamptz IS NULL OR COALESCE(br.started_at, br.created_at) >= $4)
    AND ($5::timestamptz IS NULL OR COALESCE(br.started_at, br.created_at) < $5)
  ORDER BY COALESCE(br.started_at, br.created_at) DESC
  LIMIT $3
),
total_count AS (
  SELECT COUNT(*)::int AS cnt
  FROM ci_runs br
  WHERE br.repository_id IN (SELECT repository_id FROM linked_repos)
    AND ($4::timestamptz IS NULL OR COALESCE(br.started_at, br.created_at) >= $4)
    AND ($5::timestamptz IS NULL OR COALESCE(br.started_at, br.created_at) < $5)
)
SELECT
  (SELECT num FROM weighted) AS weighted_num,
  (SELECT denom FROM weighted) AS weighted_denom,
  (SELECT json_object_agg(status, cnt) FROM totals_rows) AS totals_json,
  (SELECT json_agg(json_build_object(
    'id', id,
    'repository_id', repository_id,
    'repository_full_name', repository_full_name,
    'run_external_id', run_external_id,
    'branch', branch,
    'commit_sha', commit_sha,
    'status', status,
    'duration_seconds', duration_seconds,
    'started_at', started_at,
    'finished_at', finished_at
  )) FROM recent_rows) AS recent_json,
  (SELECT cnt FROM total_count) AS total_count`

	var (
		weightedNum   float64
		weightedDenom float64
		totalsJSON    []byte
		recentJSON    []byte
		totalCount    int
	)
	wf := windowFromOrNil(opts.WindowFrom)
	wt := windowToOrNil(opts.WindowTo)
	err := s.pool.QueryRow(ctx, query, projectID, wf, limit, wf, wt).Scan(
		&weightedNum, &weightedDenom, &totalsJSON, &recentJSON, &totalCount,
	)
	if err != nil {
		return domain.ProjectWeightedTestResults{}, 0, fmt.Errorf("list project test results: %w", err)
	}

	// totals 정규화 — 7 status 모두 0 초기화 + store 의 raw status merge.
	// Sprint A 의 RepositoryTestsSection 정공법과 동일.
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
			return domain.ProjectWeightedTestResults{}, 0, fmt.Errorf("decode totals json: %w", err)
		}
		for k, v := range rawTotals {
			// 7 status 외의 status (e.g. "weird_status") 는 unknown 으로 합산.
			if _, ok := totals[k]; ok {
				totals[k] = v
			} else {
				totals["unknown"] += v
			}
		}
	}

	var recent []domain.ProjectBuildRun
	if len(recentJSON) > 0 {
		if err := json.Unmarshal(recentJSON, &recent); err != nil {
			return domain.ProjectWeightedTestResults{}, 0, fmt.Errorf("decode recent json: %w", err)
		}
	}

	var passRate *float64
	if weightedDenom > 0 {
		v := weightedNum / weightedDenom
		passRate = &v
	}

	return domain.ProjectWeightedTestResults{
		ProjectID:        projectID,
		WindowFrom:       opts.WindowFrom.UTC(),
		WindowTo:         opts.WindowTo.UTC(),
		WeightedPassRate: passRate,
		Totals:           totals,
		Recent:           recent,
	}, totalCount, nil
}
