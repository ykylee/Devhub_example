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
	Status string
	Branch string
	Limit  int
	Offset int
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

// ListRepositoryBuildRuns returns paginated build_runs rows.
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
SELECT COUNT(*) FROM build_runs
WHERE repository_id = $1
  AND ($2 = '' OR status = $2)
  AND ($3 = '' OR branch = $3)`
	var total int
	if err := s.pool.QueryRow(ctx, countQuery, repositoryID, opts.Status, opts.Branch).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count build runs: %w", err)
	}
	const query = `
SELECT id, repository_id, run_external_id, branch, commit_sha, status,
       duration_seconds, started_at, finished_at, created_at
FROM build_runs
WHERE repository_id = $3
  AND ($4 = '' OR status = $4)
  AND ($5 = '' OR branch = $5)
ORDER BY started_at DESC
LIMIT $1 OFFSET $2`
	rows, err := s.pool.Query(ctx, query, limit, offset, repositoryID, opts.Status, opts.Branch)
	if err != nil {
		return nil, 0, fmt.Errorf("list build runs: %w", err)
	}
	defer rows.Close()
	out := make([]domain.BuildRun, 0, limit)
	for rows.Next() {
		var b domain.BuildRun
		var duration *int
		var finished *time.Time
		if err := rows.Scan(&b.ID, &b.RepositoryID, &b.RunExternalID, &b.Branch, &b.CommitSHA,
			&b.Status, &duration, &b.StartedAt, &finished, &b.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan build run: %w", err)
		}
		b.DurationSeconds = duration
		b.FinishedAt = finished
		out = append(out, b)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate build runs: %w", err)
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

// --- Application 롤업 (API-57, concept §13.4) ---

// ComputeApplicationRollup aggregates connected repos' metrics into Application-level
// rollup with weight_policy normalize (concept §13.4). 1차 구현 — 정확성 우선, 성능
// 최적화 (cache / pre-aggregation) 는 carve out.
func (s *PostgresStore) ComputeApplicationRollup(ctx context.Context, applicationID string, opts domain.ApplicationRollupOptions) (domain.ApplicationRollup, error) {
	if opts.WindowFrom.IsZero() {
		opts.WindowFrom = time.Now().UTC().AddDate(0, 0, -30)
	}
	if opts.WindowTo.IsZero() {
		opts.WindowTo = time.Now().UTC()
	}
	if opts.Policy == "" {
		opts.Policy = domain.WeightPolicyEqual
	}

	// 1) link 조회 (repo_provider + repo_full_name) — local store FK lookup 후 repositories.id 매핑
	const linksQuery = `
SELECT ar.repo_provider, ar.repo_full_name, ar.role, ar.sync_status, COALESCE(ar.sync_error_code, ''),
       r.id
FROM application_repositories ar
LEFT JOIN repositories r ON r.full_name = ar.repo_full_name
WHERE ar.application_id = $1::uuid`
	rows, err := s.pool.Query(ctx, linksQuery, applicationID)
	if err != nil {
		return domain.ApplicationRollup{}, fmt.Errorf("list rollup links: %w", err)
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
			return domain.ApplicationRollup{}, fmt.Errorf("scan rollup link: %w", err)
		}
		links = append(links, l)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return domain.ApplicationRollup{}, fmt.Errorf("iterate rollup links: %w", err)
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
		// 각 role 의 인원수 측정
		roleCount := map[string]int{}
		for _, l := range contributing {
			roleCount[l.Role]++
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
		for _, l := range contributing {
			share := used[l.Role] / totalUsed / float64(roleCount[l.Role])
			rawWeights[l.FullName] = share
		}
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
			return domain.ApplicationRollup{}, fmt.Errorf("invalid weight policy: negative weight")
		}
		if sum < 1.0-domain.CustomWeightTolerance || sum > 1.0+domain.CustomWeightTolerance {
			return domain.ApplicationRollup{}, fmt.Errorf("invalid weight policy: custom weights must sum to 1.0 (got %.4f)", sum)
		}
		for _, l := range contributing {
			if w, ok := opts.CustomWeights[l.FullName]; ok {
				rawWeights[l.FullName] = w
			} else {
				// missing → equal fallback
				if n := len(contributing); n > 0 {
					fallbackW := 1.0 / float64(n)
					rawWeights[l.FullName] = fallbackW
					fallbacks = append(fallbacks, domain.RollupFallback{
						RepoFullName:  l.FullName,
						Provider:      l.Provider,
						Reason:        "custom_weight_missing",
						AppliedWeight: fallbackW,
					})
				}
			}
		}
	default:
		return domain.ApplicationRollup{}, fmt.Errorf("invalid weight policy: unknown policy %q", opts.Policy)
	}

	// 3) 각 contributing repo 의 메트릭 fetch + weighted sum
	prDistribution := map[string]int{}
	var weightedBuildSuccessRate, weightedQualityScore float64
	var weightedBuildDuration float64
	var gateFailedCount int
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
			return domain.ApplicationRollup{}, fmt.Errorf("rollup pr distribution: %w", err)
		}
		for prRows.Next() {
			var etype string
			var cnt int
			if err := prRows.Scan(&etype, &cnt); err != nil {
				prRows.Close()
				return domain.ApplicationRollup{}, fmt.Errorf("scan pr distribution: %w", err)
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
			return domain.ApplicationRollup{}, fmt.Errorf("rollup build aggregate: %w", err)
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
			return domain.ApplicationRollup{}, fmt.Errorf("rollup quality aggregate: %w", err)
		}
		if avgScore != nil {
			weightedQualityScore += *avgScore * weight
		}
		gateFailedCount += failedGates // weight 무관 합산
	}

	// 4) critical warning count — 1차 정의: gateFailedCount + buildFailureRate>0.5 인 repo 수
	criticalCount := gateFailedCount
	if weightedBuildSuccessRate < 0.5 && len(contributing) > 0 {
		criticalCount++
	}

	rollup := domain.ApplicationRollup{
		PullRequestDistribution: prDistribution,
		BuildSuccessRate:        weightedBuildSuccessRate,
		BuildAvgDurationSeconds: int(weightedBuildDuration),
		QualityScore:            weightedQualityScore,
		QualityGateFailedCount:  gateFailedCount,
		CriticalWarningCount:    criticalCount,
		Meta: domain.ApplicationRollupMeta{
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

// CountApplicationCriticalWarnings — active→closed 가드 (concept §13.2.1) 흡수.
// 1차 정의: 어떤 연결 repo 라도 quality_gate_passed=false 가 있거나 build success rate <
// 50% 이면 critical. 가드 임계치 외부화는 후속 (concept §13.2.1 운영 메모).
func (s *PostgresStore) CountApplicationCriticalWarnings(ctx context.Context, applicationID string) (int, error) {
	opts := domain.ApplicationRollupOptions{
		Policy:     domain.WeightPolicyEqual,
		WindowFrom: time.Now().UTC().AddDate(0, 0, -30),
		WindowTo:   time.Now().UTC(),
	}
	rollup, err := s.ComputeApplicationRollup(ctx, applicationID, opts)
	if err != nil {
		return 0, err
	}
	return rollup.CriticalWarningCount, nil
}
