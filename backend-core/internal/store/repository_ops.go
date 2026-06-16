package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/jackc/pgx/v5"
)

// Repository 운영 지표 store 메서드 (API-51..54, sprint claude/work_260514-c).
//
// pr_activities / build_runs / quality_snapshots 는 migration 000017 의 테이블.
// 본 sprint 는 read-only 조회만 — write 는 ingest pipeline (별도 sprint).

// RepositoryActivityOptions parameterizes ListRepositoryActivity.
type RepositoryActivityOptions struct {
	WindowFrom time.Time
	WindowTo   time.Time
}

// PRActivityListOptions parameterizes ListRepositoryPullRequests.
type PRActivityListOptions struct {
	EventType string
	Limit     int
	Offset    int
}

// BuildRunListOptions parameterizes ListRepositoryBuildRuns.
type BuildRunListOptions struct {
	Status     string
	Branch     string
	Limit      int
	Offset     int
	WindowFrom time.Time
	WindowTo   time.Time
}

// QualitySnapshotListOptions parameterizes ListRepositoryQualitySnapshots.
type QualitySnapshotListOptions struct {
	Tool   string
	Limit  int
	Offset int
}

// ListRepositoryActivity aggregates pr_activities + build_runs for a window.
// 1차 구현 — commit 활동량은 후속 (ingest pipeline 에 commit 이벤트 도입 시점).
func (s *PostgresStore) ListRepositoryActivity(ctx context.Context, repositoryID int64, opts RepositoryActivityOptions) (domain.RepositoryActivity, error) {
	if opts.WindowFrom.IsZero() {
		opts.WindowFrom = time.Now().UTC().AddDate(0, 0, -30) // 기본 최근 30일
	}
	if opts.WindowTo.IsZero() {
		opts.WindowTo = time.Now().UTC()
	}

	const prAggQuery = `
SELECT COUNT(*),
       COALESCE(array_agg(DISTINCT actor_login) FILTER (WHERE actor_login IS NOT NULL AND actor_login <> ''), '{}')
FROM pr_activities
WHERE repository_id = $1 AND occurred_at >= $2 AND occurred_at < $3`

	activity := domain.RepositoryActivity{
		RepositoryID: repositoryID,
		WindowFrom:   opts.WindowFrom,
		WindowTo:     opts.WindowTo,
	}
	if err := s.pool.QueryRow(ctx, prAggQuery, repositoryID, opts.WindowFrom, opts.WindowTo).
		Scan(&activity.PREventCount, &activity.ActiveContributors); err != nil {
		return domain.RepositoryActivity{}, fmt.Errorf("aggregate pr activity: %w", err)
	}

	const buildAggQuery = `
SELECT COUNT(*),
       COUNT(*) FILTER (WHERE status = 'success')::float / NULLIF(COUNT(*), 0)
FROM build_runs
WHERE repository_id = $1 AND started_at >= $2 AND started_at < $3`
	var successRate *float64
	if err := s.pool.QueryRow(ctx, buildAggQuery, repositoryID, opts.WindowFrom, opts.WindowTo).
		Scan(&activity.BuildRunCount, &successRate); err != nil {
		return domain.RepositoryActivity{}, fmt.Errorf("aggregate build runs: %w", err)
	}
	if successRate != nil {
		activity.BuildSuccessRate = *successRate
	}

	// 마지막 빌드 상태 + 시각 (REQ-FR-APPDASH-001 — 단순 % 보다 broken/red 즉시 표기).
	// window 무관하게 build_runs 의 가장 최근 1건. 없으면 LastBuildStatus="" 유지 → handler
	// 가 "unknown" 으로 노출.
	const lastBuildQuery = `
SELECT status, started_at
FROM build_runs
WHERE repository_id = $1
ORDER BY started_at DESC
LIMIT 1`
	var lbStatus string
	var lbStartedAt time.Time
	switch err := s.pool.QueryRow(ctx, lastBuildQuery, repositoryID).Scan(&lbStatus, &lbStartedAt); {
	case err == nil:
		activity.LastBuildStatus = lbStatus
		t := lbStartedAt
		activity.LastBuildAt = &t
	case errors.Is(err, pgx.ErrNoRows):
		// build_runs 무 → unknown 처리는 handler 측에서.
	default:
		return domain.RepositoryActivity{}, fmt.Errorf("last build run: %w", err)
	}

	return activity, nil
}

// ListRepositoryPullRequests returns paginated pr_activities rows for a Repository.
func (s *PostgresStore) ListRepositoryPullRequests(ctx context.Context, repositoryID int64, opts PRActivityListOptions) ([]domain.PRActivity, int, error) {
	limit := opts.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}
	const countQuery = `
SELECT COUNT(*) FROM pr_activities
WHERE repository_id = $1 AND ($2 = '' OR event_type = $2)`
	var total int
	if err := s.pool.QueryRow(ctx, countQuery, repositoryID, opts.EventType).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count pr activities: %w", err)
	}
	const query = `
SELECT id, repository_id, external_pr_id, event_type,
       COALESCE(actor_login, ''), occurred_at,
       payload, created_at
FROM pr_activities
WHERE repository_id = $3 AND ($4 = '' OR event_type = $4)
ORDER BY occurred_at DESC
LIMIT $1 OFFSET $2`
	rows, err := s.pool.Query(ctx, query, limit, offset, repositoryID, opts.EventType)
	if err != nil {
		return nil, 0, fmt.Errorf("list pr activities: %w", err)
	}
	defer rows.Close()
	out := make([]domain.PRActivity, 0, limit)
	for rows.Next() {
		var a domain.PRActivity
		var payload []byte
		if err := rows.Scan(&a.ID, &a.RepositoryID, &a.ExternalPRID, &a.EventType,
			&a.ActorLogin, &a.OccurredAt, &payload, &a.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan pr activity: %w", err)
		}
		if len(payload) > 0 {
			_ = json.Unmarshal(payload, &a.Payload)
		}
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate pr activities: %w", err)
	}
	return out, total, nil
}

// ListRepositoryBuildRuns returns paginated build_runs rows sourced from ci_runs.
// ISSUE-04/P1-7: ci_runs is the write path (UpsertCIRun/CreateCIRun); build_runs
// table has no ingest pipeline, so we read from ci_runs directly.
func (s *PostgresStore) ListRepositoryBuildRuns(ctx context.Context, repositoryID int64, opts BuildRunListOptions) ([]domain.BuildRun, int, error) {
	limit := opts.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}
	const countQuery = `
SELECT COUNT(*) FROM ci_runs
WHERE repository_id = $1
  AND ($2 = '' OR status = $2)
  AND ($3 = '' OR branch = $3)
  AND ($4::timestamptz IS NULL OR COALESCE(started_at, created_at) >= $4)
  AND ($5::timestamptz IS NULL OR COALESCE(started_at, created_at) < $5)`
	var total int
	if err := s.pool.QueryRow(ctx, countQuery, repositoryID, opts.Status, opts.Branch, windowFromOrNil(opts.WindowFrom), windowToOrNil(opts.WindowTo)).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count ci_runs: %w", err)
	}
	const query = `
SELECT id, repository_id, external_id, branch, commit_sha, status,
       duration_seconds, COALESCE(started_at, created_at), finished_at, created_at
FROM ci_runs
WHERE repository_id = $3
  AND ($4 = '' OR status = $4)
  AND ($5 = '' OR branch = $5)
  AND ($6::timestamptz IS NULL OR COALESCE(started_at, created_at) >= $6)
  AND ($7::timestamptz IS NULL OR COALESCE(started_at, created_at) < $7)
ORDER BY COALESCE(started_at, created_at) DESC
LIMIT $1 OFFSET $2`
	rows, err := s.pool.Query(ctx, query, limit, offset, repositoryID, opts.Status, opts.Branch, windowFromOrNil(opts.WindowFrom), windowToOrNil(opts.WindowTo))
	if err != nil {
		return nil, 0, fmt.Errorf("list ci_runs as build runs: %w", err)
	}
	defer rows.Close()
	out := make([]domain.BuildRun, 0, limit)
	for rows.Next() {
		var b domain.BuildRun
		var duration *int
		var finished *time.Time
		if err := rows.Scan(&b.ID, &b.RepositoryID, &b.RunExternalID, &b.Branch, &b.CommitSHA,
			&b.Status, &duration, &b.StartedAt, &finished, &b.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan ci_run as build run: %w", err)
		}
		b.DurationSeconds = duration
		b.FinishedAt = finished
		out = append(out, b)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate ci_runs: %w", err)
	}
	return out, total, nil
}

// ListRepositoryQualitySnapshots returns paginated quality_snapshots rows.
func (s *PostgresStore) ListRepositoryQualitySnapshots(ctx context.Context, repositoryID int64, opts QualitySnapshotListOptions) ([]domain.QualitySnapshot, int, error) {
	limit := opts.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}
	const countQuery = `
SELECT COUNT(*) FROM quality_snapshots
WHERE repository_id = $1 AND ($2 = '' OR tool = $2)`
	var total int
	if err := s.pool.QueryRow(ctx, countQuery, repositoryID, opts.Tool).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count quality snapshots: %w", err)
	}
	const query = `
SELECT id, repository_id, tool, ref_name, COALESCE(commit_sha, ''),
       score, gate_passed, metric_payload, measured_at, created_at
FROM quality_snapshots
WHERE repository_id = $3 AND ($4 = '' OR tool = $4)
ORDER BY measured_at DESC
LIMIT $1 OFFSET $2`
	rows, err := s.pool.Query(ctx, query, limit, offset, repositoryID, opts.Tool)
	if err != nil {
		return nil, 0, fmt.Errorf("list quality snapshots: %w", err)
	}
	defer rows.Close()
	out := make([]domain.QualitySnapshot, 0, limit)
	for rows.Next() {
		var q domain.QualitySnapshot
		var score *float64
		var gate *bool
		var payload []byte
		if err := rows.Scan(&q.ID, &q.RepositoryID, &q.Tool, &q.RefName, &q.CommitSHA,
			&score, &gate, &payload, &q.MeasuredAt, &q.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan quality snapshot: %w", err)
		}
		q.Score = score
		q.GatePassed = gate
		if len(payload) > 0 {
			_ = json.Unmarshal(payload, &q.MetricPayload)
		}
		out = append(out, q)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate quality snapshots: %w", err)
	}
	return out, total, nil
}

// CountOpenAndMergedPRs returns distinct PR number count for event_type='opened' (open)
// and event_type='merged' (merged). Sprint A — kpi-tests-per-domain-scope.md §6.1 repository
// KPI 종합 정공법. state="closed" + merged_at IS NOT NULL row 는 event_type='merged' 로
// upsert 되므로 본 query 가 정확. 그 외 (state="closed" + not merged) 는 합산에서 제외
// (raw kpi 가치 낮음).
func (s *PostgresStore) CountOpenAndMergedPRs(ctx context.Context, repositoryID int64, from, to time.Time) (int, int, error) {
	const query = `
SELECT
  COUNT(DISTINCT number) FILTER (WHERE event_type = 'opened')::int AS open_count,
  COUNT(DISTINCT number) FILTER (WHERE event_type = 'merged')::int AS merged_count
FROM pr_activities
WHERE repository_id = $1 AND occurred_at >= $2 AND occurred_at < $3`
	var open, merged int
	if err := s.pool.QueryRow(ctx, query, repositoryID, from, to).Scan(&open, &merged); err != nil {
		return 0, 0, fmt.Errorf("count open/merged prs: %w", err)
	}
	return open, merged, nil
}

// --- Project 가중치 rollup (Sprint B — kpi-tests-per-domain-scope.md §2.2 + §6.2) ---

// ComputeProjectWeightedKPI 는 project 의 N개 linked repository 의 raw metric 을
// contribution_weight 로 가중평균한 ProjectWeightedKPI 를 반환.
//
// 가중치 정공법:
//   - WeightedQualityScore = Σ(quality_score_i × weight_i) / Σ(weight_i)
//   - WeightedBuildSuccess = Σ(build_success_rate_i × weight_i) / Σ(weight_i)
//   - TotalBuildRunCount = Σ(build_run_count_i) — 단순 합산
//   - ActiveContributorCount = Σ(distinct contributors_i) — 단순 합산
//
// linked_repository_count = 0 인 경우: 모든 가중치 metric = 0 (NULLIF → division by 0 회피).
// quality_snapshot 없거나 build_run 없는 repo 는 해당 metric 만 NULL → 0 으로 COALESCE.
func (s *PostgresStore) ComputeProjectWeightedKPI(ctx context.Context, projectID string, opts RepositoryActivityOptions) (domain.ProjectWeightedKPI, error) {
	if opts.WindowFrom.IsZero() {
		opts.WindowFrom = time.Now().UTC().AddDate(0, 0, -30)
	}
	if opts.WindowTo.IsZero() {
		opts.WindowTo = time.Now().UTC()
	}
	const query = `
SELECT
  COUNT(*)::int AS linked_repo_count,
  COALESCE(SUM(latest_quality.score * pr.contribution_weight) / NULLIF(SUM(pr.contribution_weight), 0), 0)::float8 AS weighted_quality_score,
  COALESCE(SUM(activity.build_success_rate * pr.contribution_weight) / NULLIF(SUM(pr.contribution_weight), 0), 0)::float8 AS weighted_build_success_rate,
  COALESCE(SUM(activity.build_run_count), 0)::int AS total_build_run_count,
  COALESCE(SUM(activity.active_contributor_count), 0)::int AS total_active_contributors
FROM project_repositories pr
JOIN repositories r ON r.id = pr.repository_id
LEFT JOIN LATERAL (
  SELECT score FROM quality_snapshots qs
  WHERE qs.repository_id = r.id
  ORDER BY measured_at DESC LIMIT 1
) latest_quality ON true
LEFT JOIN LATERAL (
  SELECT
    COALESCE(SUM(CASE WHEN br.status = 'success' THEN 1 ELSE 0 END)::float / NULLIF(COUNT(*), 0), 0) AS build_success_rate,
    COUNT(*)::int AS build_run_count,
    COUNT(DISTINCT br.commit_author)::int AS active_contributor_count
  FROM build_runs br
  WHERE br.repository_id = r.id
    AND COALESCE(br.started_at, br.created_at) >= $2
    AND COALESCE(br.started_at, br.created_at) < $3
) activity ON true
WHERE pr.project_id = $1::uuid`

	var (
		linkedRepoCount   int
		weightedQuality   float64
		weightedBSR       float64
		totalBuildRuns    int
		totalContributors int
	)
	if err := s.pool.QueryRow(ctx, query, projectID, opts.WindowFrom, opts.WindowTo).Scan(
		&linkedRepoCount, &weightedQuality, &weightedBSR, &totalBuildRuns, &totalContributors,
	); err != nil {
		return domain.ProjectWeightedKPI{}, fmt.Errorf("compute project weighted kpi: %w", err)
	}

	return domain.ProjectWeightedKPI{
		ProjectID:              projectID,
		WindowFrom:             opts.WindowFrom.UTC(),
		WindowTo:               opts.WindowTo.UTC(),
		WeightedQualityScore:   weightedQuality,
		WeightedBuildSuccess:   weightedBSR,
		TotalBuildRunCount:     totalBuildRuns,
		ActiveContributorCount: totalContributors,
		LinkedRepositoryCount:  linkedRepoCount,
		WeightedAt:             time.Now().UTC(),
	}, nil
}

// CountProjectOpenAndMergedPRs 는 project 의 N개 linked repository 의 open/merged
// PR 을 contribution_weight 로 가중평균. Sprint B §2.2.
//
// 정공법: 각 repo 의 (open_i, merged_i) × weight_i → Σ(weighted). 단순 카운트가 아닌
// 가중치 적용 (정수 반올림). linked_repo 0 → (0, 0).
func (s *PostgresStore) CountProjectOpenAndMergedPRs(ctx context.Context, projectID string, from, to time.Time) (int, int, error) {
	const query = `
SELECT
  COALESCE(SUM(pr_stats.open_count * pr.contribution_weight), 0)::float8 AS weighted_open,
  COALESCE(SUM(pr_stats.merged_count * pr.contribution_weight), 0)::float8 AS weighted_merged
FROM project_repositories pr
LEFT JOIN LATERAL (
  SELECT
    COUNT(DISTINCT number) FILTER (WHERE event_type = 'opened')::float8 AS open_count,
    COUNT(DISTINCT number) FILTER (WHERE event_type = 'merged')::float8 AS merged_count
  FROM pr_activities pa
  WHERE pa.repository_id = pr.repository_id
    AND pa.occurred_at >= $2 AND pa.occurred_at < $3
) pr_stats ON true
WHERE pr.project_id = $1::uuid`

	var weightedOpen, weightedMerged float64
	if err := s.pool.QueryRow(ctx, query, projectID, from, to).Scan(&weightedOpen, &weightedMerged); err != nil {
		return 0, 0, fmt.Errorf("count project open/merged prs: %w", err)
	}
	// 가중치 적용 결과는 소수점 (Σ(count × weight) 의 정확한 가중치값). 정수 반올림은
	// handler 에서 (UI 표시용). store 는 raw value 반환.
	return int(weightedOpen + 0.5), int(weightedMerged + 0.5), nil
}

// --- Application 롤업 (API-57, concept §13.4) ---

// ComputePlatformRollup aggregates connected repos' metrics into Application-level
// rollup with weight_policy normalize (concept §13.4). 1차 구현 — 정확성 우선, 성능
// 최적화 (cache / pre-aggregation) 는 carve out.
func (s *PostgresStore) ComputePlatformRollup(ctx context.Context, platformID string, opts domain.PlatformRollupOptions) (domain.PlatformRollup, error) {
	if opts.WindowFrom.IsZero() {
		opts.WindowFrom = time.Now().UTC().AddDate(0, 0, -30)
	}
	if opts.WindowTo.IsZero() {
		opts.WindowTo = time.Now().UTC()
	}
	if opts.Policy == "" {
		opts.Policy = domain.WeightPolicyEqual
	}

	// 1) link 조회 (repo_provider + repo_full_name) — 직접 연결 + 프로젝트 경유 간접 연결 합집합 (중복 배제)
	const linksQuery = `
WITH combined_repos AS (
  SELECT 
    ar.repo_provider AS provider, 
    ar.repo_full_name AS full_name, 
    ar.role AS role, 
    ar.sync_status AS sync_status, 
    COALESCE(ar.sync_error_code, '') AS sync_error_code,
    r.id AS repo_id
  FROM platform_repositories ar
  LEFT JOIN repositories r ON r.full_name = ar.repo_full_name
  WHERE ar.platform_id = $1::uuid

  UNION ALL

  SELECT 
    COALESCE(p_prov.provider_key, 'gitea') AS provider,
    r.full_name AS full_name,
    pr.role AS role,
    'active' AS sync_status,
    '' AS sync_error_code,
    r.id AS repo_id
  FROM projects p
  JOIN project_repositories pr ON pr.project_id = p.id
  JOIN repositories r ON r.id = pr.repository_id
  LEFT JOIN integration_providers p_prov ON p_prov.provider_id = r.provider_id
  WHERE p.platform_id = $1::uuid
)
SELECT DISTINCT ON (full_name) provider, full_name, role, sync_status, sync_error_code, repo_id
FROM combined_repos
ORDER BY full_name, CASE WHEN sync_status = 'active' THEN 0 ELSE 1 END ASC`
	rows, err := s.pool.Query(ctx, linksQuery, platformID)
	if err != nil {
		return domain.PlatformRollup{}, fmt.Errorf("list rollup links: %w", err)
	}
	type linkRow struct {
		Provider, FullName, Role, SyncStatus, SyncErrCode string
		RepoID                                            *int64 // repositories.id (NULL 이면 미수집)
	}
	links := make([]linkRow, 0, 8)
	for rows.Next() {
		var l linkRow
		if err := rows.Scan(&l.Provider, &l.FullName, &l.Role, &l.SyncStatus, &l.SyncErrCode, &l.RepoID); err != nil {
			rows.Close()
			return domain.PlatformRollup{}, fmt.Errorf("scan rollup link: %w", err)
		}
		links = append(links, l)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return domain.PlatformRollup{}, fmt.Errorf("iterate rollup links: %w", err)
	}

	// 2) data_gaps + weight normalize 계산
	gaps := make([]domain.RollupDataGap, 0)
	fallbacks := make([]domain.RollupFallback, 0)
	rawWeights := make(map[string]float64) // repo_full_name → weight (정규화 전)
	contributing := make([]linkRow, 0, len(links))

	for _, l := range links {
		if l.RepoID == nil || l.SyncStatus != "active" {
			reason := "no_data_in_window"
			if l.SyncStatus == "degraded" || l.SyncErrCode != "" {
				reason = "provider_unreachable"
			} else if l.SyncStatus == "disconnected" {
				reason = "disconnected"
			}
			gaps = append(gaps, domain.RollupDataGap{
				RepoFullName: l.FullName, Provider: l.Provider, Reason: reason,
			})
			continue
		}
		contributing = append(contributing, l)
	}

	switch opts.Policy {
	case domain.WeightPolicyEqual:
		if n := len(contributing); n > 0 {
			w := 1.0 / float64(n)
			for _, l := range contributing {
				rawWeights[l.FullName] = w
			}
		}
	case domain.WeightPolicyRepoRole:
		roleSums := map[string]float64{"primary": 0.6, "sub": 0.3, "shared": 0.1}
		// 각 role 의 인원수 측정 — known role 만 집계 (unknown role 은 fallback 대상).
		roleCount := map[string]int{}
		unknownRoleLinks := make([]linkRow, 0)
		for _, l := range contributing {
			if _, known := roleSums[l.Role]; known {
				roleCount[l.Role]++
			} else {
				unknownRoleLinks = append(unknownRoleLinks, l)
			}
		}
		// 비어 있는 role 의 가중치를 다른 role 에 비례 재분배 (concept §13.4)
		used := map[string]float64{}
		totalUsed := 0.0
		for role, weight := range roleSums {
			if roleCount[role] > 0 {
				used[role] = weight
				totalUsed += weight
			}
		}
		// PR #107 self-review B1 — division-by-zero 안전망 + unknown role fallback.
		// totalUsed=0 시나리오 (모든 contributing repo 의 role 이 catalogue 외 값) 에서
		// equal fallback 으로 정규화. unknown role 는 `unknown_role_fallback_equal`
		// RollupFallback 기록.
		if totalUsed == 0 || len(contributing) == 0 {
			if n := len(contributing); n > 0 {
				w := 1.0 / float64(n)
				for _, l := range contributing {
					rawWeights[l.FullName] = w
					fallbacks = append(fallbacks, domain.RollupFallback{
						RepoFullName:  l.FullName,
						Provider:      l.Provider,
						Reason:        "unknown_role_fallback_equal",
						AppliedWeight: w,
					})
				}
			}
			break
		}
		for _, l := range contributing {
			if _, known := roleSums[l.Role]; !known {
				// catalogue 외 role — equal fallback (다른 known role 의 평균).
				w := 1.0 / float64(len(contributing))
				rawWeights[l.FullName] = w
				fallbacks = append(fallbacks, domain.RollupFallback{
					RepoFullName:  l.FullName,
					Provider:      l.Provider,
					Reason:        "unknown_role_fallback_equal",
					AppliedWeight: w,
				})
				continue
			}
			share := used[l.Role] / totalUsed / float64(roleCount[l.Role])
			rawWeights[l.FullName] = share
		}
		_ = unknownRoleLinks // 위 루프에서 처리 — placeholder.
	case domain.WeightPolicyCustom:
		// 합계 검증
		sum := 0.0
		negative := false
		for _, w := range opts.CustomWeights {
			if w < 0 {
				negative = true
				break
			}
			sum += w
		}
		if negative {
			return domain.PlatformRollup{}, fmt.Errorf("invalid weight policy: negative weight")
		}
		if sum < 1.0-domain.CustomWeightTolerance || sum > 1.0+domain.CustomWeightTolerance {
			return domain.PlatformRollup{}, fmt.Errorf("invalid weight policy: custom weights must sum to 1.0 (got %.4f)", sum)
		}
		missingCount := 0
		for _, l := range contributing {
			if w, ok := opts.CustomWeights[l.FullName]; ok {
				rawWeights[l.FullName] = w
			} else {
				// missing → equal fallback (정규화는 아래에서)
				if n := len(contributing); n > 0 {
					fallbackW := 1.0 / float64(n)
					rawWeights[l.FullName] = fallbackW
					fallbacks = append(fallbacks, domain.RollupFallback{
						RepoFullName:  l.FullName,
						Provider:      l.Provider,
						Reason:        "custom_weight_missing",
						AppliedWeight: fallbackW,
					})
					missingCount++
				}
			}
		}
		// PR #107 codex review P1 — fallback 후 sum=1.0 재정규화.
		// custom_weights 가 이미 합 1.0 으로 검증된 상태에서 누락 repo 에 1/N equal
		// fallback 을 부여하면 최종 합이 1.0 초과 → build_success_rate 같은 weighted
		// metric 가 1.0 초과 (수치 오염). missing 이 있을 때만 전체 정규화하고
		// fallbacks meta 의 AppliedWeight 도 정규화 후 값으로 갱신.
		if missingCount > 0 {
			total := 0.0
			for _, w := range rawWeights {
				total += w
			}
			if total > 0 {
				for k := range rawWeights {
					rawWeights[k] /= total
				}
				for i := range fallbacks {
					if w, ok := rawWeights[fallbacks[i].RepoFullName]; ok {
						fallbacks[i].AppliedWeight = w
					}
				}
			}
		}
	default:
		return domain.PlatformRollup{}, fmt.Errorf("invalid weight policy: unknown policy %q", opts.Policy)
	}

	// 3) 각 contributing repo 의 메트릭 fetch + weighted sum
	prDistribution := map[string]int{}
	var weightedBuildSuccessRate, weightedQualityScore float64
	var weightedBuildDuration float64
	var gateFailedCount int
	// targetBranchBuildStatus 도출 — REQ-FR-APPDASH-001 (단순 % 보다 broken/red 즉시 표기).
	//   - 어떤 contributing repo 의 last build = failed|cancelled → "broken"
	//   - 모두 success|skipped → "healthy"
	//   - 그 외 (running/queued/unknown/데이터 없음) 한 건이라도 + broken 없음 → "unknown"
	//   - contributing 전체 비어있음 → "unknown"
	var (
		sawBroken bool
		sawHealthy bool
		sawIndeterminate bool
	)
	for _, l := range contributing {
		if l.RepoID == nil {
			continue
		}
		repoID := *l.RepoID
		weight := rawWeights[l.FullName]
		if weight == 0 {
			continue
		}

		// PR distribution: pr_activities event_type 집계
		const prDistQuery = `
SELECT event_type, COUNT(*)
FROM pr_activities
WHERE repository_id = $1 AND occurred_at >= $2 AND occurred_at < $3
GROUP BY event_type`
		prRows, err := s.pool.Query(ctx, prDistQuery, repoID, opts.WindowFrom, opts.WindowTo)
		if err != nil {
			return domain.PlatformRollup{}, fmt.Errorf("rollup pr distribution: %w", err)
		}
		for prRows.Next() {
			var etype string
			var cnt int
			if err := prRows.Scan(&etype, &cnt); err != nil {
				prRows.Close()
				return domain.PlatformRollup{}, fmt.Errorf("scan pr distribution: %w", err)
			}
			prDistribution[etype] += cnt // PR 분포는 weight 무관 합산 (raw count)
		}
		prRows.Close()

		// Build aggregate
		const buildAggQuery = `
SELECT COUNT(*),
       COUNT(*) FILTER (WHERE status = 'success')::float / NULLIF(COUNT(*), 0),
       AVG(duration_seconds)::float
FROM build_runs
WHERE repository_id = $1 AND started_at >= $2 AND started_at < $3`
		var buildCount int
		var rate, avgDur *float64
		if err := s.pool.QueryRow(ctx, buildAggQuery, repoID, opts.WindowFrom, opts.WindowTo).
			Scan(&buildCount, &rate, &avgDur); err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return domain.PlatformRollup{}, fmt.Errorf("rollup build aggregate: %w", err)
		}
		if rate != nil {
			weightedBuildSuccessRate += *rate * weight
		}
		if avgDur != nil {
			weightedBuildDuration += *avgDur * weight
		}

		// Quality latest snapshot per (repo, tool)
		const qualityQuery = `
WITH latest AS (
  SELECT DISTINCT ON (tool) tool, score, gate_passed
  FROM quality_snapshots
  WHERE repository_id = $1 AND measured_at >= $2 AND measured_at < $3
  ORDER BY tool, measured_at DESC
)
SELECT AVG(score)::float, COUNT(*) FILTER (WHERE gate_passed = false)
FROM latest`
		var avgScore *float64
		var failedGates int
		if err := s.pool.QueryRow(ctx, qualityQuery, repoID, opts.WindowFrom, opts.WindowTo).
			Scan(&avgScore, &failedGates); err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return domain.PlatformRollup{}, fmt.Errorf("rollup quality aggregate: %w", err)
		}
		if avgScore != nil {
			weightedQualityScore += *avgScore * weight
		}
		gateFailedCount += failedGates // weight 무관 합산

		// last build status — window 무관, 본 repo 의 가장 최근 build_run 1건.
		const lastBuildQuery = `
SELECT status FROM build_runs
WHERE repository_id = $1
ORDER BY started_at DESC
LIMIT 1`
		var lbStatus string
		switch err := s.pool.QueryRow(ctx, lastBuildQuery, repoID).Scan(&lbStatus); {
		case err == nil:
			switch lbStatus {
			case "failed", "cancelled":
				sawBroken = true
			case "success", "skipped":
				sawHealthy = true
			default:
				sawIndeterminate = true
			}
		case errors.Is(err, pgx.ErrNoRows):
			sawIndeterminate = true
		default:
			return domain.PlatformRollup{}, fmt.Errorf("rollup last build: %w", err)
		}
	}

	// 4) critical warning count — 1차 정의: gateFailedCount + buildFailureRate>0.5 인 repo 수
	criticalCount := gateFailedCount
	if weightedBuildSuccessRate < 0.5 && len(contributing) > 0 {
		criticalCount++
	}

	// 5) target branch build status derive — broken 우선 (안전 측 표기).
	targetBranchBuildStatus := "unknown"
	switch {
	case sawBroken:
		targetBranchBuildStatus = "broken"
	case sawHealthy && !sawIndeterminate:
		targetBranchBuildStatus = "healthy"
	}

	rollup := domain.PlatformRollup{
		PullRequestDistribution: prDistribution,
		BuildSuccessRate:        weightedBuildSuccessRate,
		BuildAvgDurationSeconds: int(weightedBuildDuration),
		QualityScore:            weightedQualityScore,
		QualityGateFailedCount:  gateFailedCount,
		CriticalWarningCount:    criticalCount,
		TargetBranchBuildStatus: targetBranchBuildStatus,
		Meta: domain.PlatformRollupMeta{
			Period:         domain.RollupPeriod{From: opts.WindowFrom, To: opts.WindowTo},
			Filters:        map[string]any{},
			WeightPolicy:   opts.Policy,
			AppliedWeights: rawWeights,
			Fallbacks:      fallbacks,
			DataGaps:       gaps,
		},
	}
	return rollup, nil
}

// CountPlatformCriticalWarnings — active→closed 가드 (concept §13.2.1) 흡수.
// 1차 정의: 어떤 연결 repo 라도 quality_gate_passed=false 가 있거나 build success rate <
// 50% 이면 critical. 가드 임계치 외부화는 후속 (concept §13.2.1 운영 메모).
func (s *PostgresStore) CountPlatformCriticalWarnings(ctx context.Context, platformID string) (int, error) {
	opts := domain.PlatformRollupOptions{
		Policy:     domain.WeightPolicyEqual,
		WindowFrom: time.Now().UTC().AddDate(0, 0, -30),
		WindowTo:   time.Now().UTC(),
	}
	rollup, err := s.ComputePlatformRollup(ctx, platformID, opts)
	if err != nil {
		return 0, err
	}
	return rollup.CriticalWarningCount, nil
}
// windowFromOrNil / windowToOrNil — pgx 가 time.Time zero value 를 nil 로
// cast 하기 위한 helper. WindowFrom/WindowTo 가 zero 면 SQL $N::timestamptz IS NULL
// 분기 → filter 비활성. ListRepositoryBuildRuns 의 window 정합.
func windowFromOrNil(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t.UTC()
}
func windowToOrNil(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t.UTC()
}
