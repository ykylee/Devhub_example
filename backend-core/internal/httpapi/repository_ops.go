package httpapi

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

var errRepositoryNotFound = "repository_not_found"

// Repository 운영 지표 read endpoint (API-51..54, sprint claude/work_260514-c).
// pr_activities / build_runs / quality_snapshots 의 read-only 조회. write 는 ingest
// pipeline 책임 (별도 sprint).

func (h *Handler) repositoryActivity(c *gin.Context) {
	storeI, ok := h.platformStoreOrUnavailable(c)
	if !ok {
		return
	}
	repoID, err := strconv.ParseInt(c.Param("repository_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "repository_id must be an integer"})
		return
	}
	opts := store.RepositoryActivityOptions{}
	if v := c.Query("from"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "from must be RFC3339"})
			return
		}
		opts.WindowFrom = t
	}
	if v := c.Query("to"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "to must be RFC3339"})
			return
		}
		opts.WindowTo = t
	}
	activity, err := storeI.ListRepositoryActivity(c.Request.Context(), repoID, opts)
	if err != nil {
		writeServerError(c, err, "repository.activity")
		return
	}
	// 마지막 빌드 상태 + 시각 — REQ-FR-APPDASH-001 (단순 % 보다 broken/red 즉시 표기).
	// 빈 status 는 "unknown" 정규화. 시각은 RFC3339 또는 nil.
	lastBuildStatus := activity.LastBuildStatus
	if lastBuildStatus == "" {
		lastBuildStatus = "unknown"
	}
	var lastBuildAt any
	if activity.LastBuildAt != nil {
		lastBuildAt = activity.LastBuildAt.UTC().Format(time.RFC3339)
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data": gin.H{
			"repository_id":       activity.RepositoryID,
			"window_from":         activity.WindowFrom.UTC().Format(time.RFC3339),
			"window_to":           activity.WindowTo.UTC().Format(time.RFC3339),
			"pr_event_count":      activity.PREventCount,
			"active_contributors": activity.ActiveContributors,
			"build_run_count":     activity.BuildRunCount,
			"build_success_rate":  activity.BuildSuccessRate,
			"last_build_status":   lastBuildStatus,
			"last_build_at":       lastBuildAt,
		},
	})
}

func (h *Handler) repositoryPullRequests(c *gin.Context) {
	storeI, ok := h.platformStoreOrUnavailable(c)
	if !ok {
		return
	}
	repoID, err := strconv.ParseInt(c.Param("repository_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "repository_id must be an integer"})
		return
	}
	opts := store.PRActivityListOptions{EventType: c.Query("event_type")}
	if v := c.Query("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 200 {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "limit must be 1..200"})
			return
		}
		opts.Limit = n
	}
	if v := c.Query("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "offset must be >= 0"})
			return
		}
		opts.Offset = n
	}
	events, total, err := storeI.ListRepositoryPullRequests(c.Request.Context(), repoID, opts)
	if err != nil {
		writeServerError(c, err, "repository.pull_requests")
		return
	}
	resp := make([]gin.H, 0, len(events))
	for _, e := range events {
		resp = append(resp, gin.H{
			"id":             e.ID,
			"repository_id":  e.RepositoryID,
			"external_pr_id": e.ExternalPRID,
			"event_type":     e.EventType,
			"actor_login":    e.ActorLogin,
			"occurred_at":    e.OccurredAt.UTC().Format(time.RFC3339),
			"payload":        e.Payload,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   resp,
		"meta":   gin.H{"total": total},
	})
}

func (h *Handler) repositoryBuildRuns(c *gin.Context) {
	storeI, ok := h.platformStoreOrUnavailable(c)
	if !ok {
		return
	}
	repoID, err := strconv.ParseInt(c.Param("repository_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "repository_id must be an integer"})
		return
	}
	// #556: RBAC guard — repository 존재 여부 확인 (부재 시 404)
	if _, err := storeI.GetRepositoryByID(c.Request.Context(), repoID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status": "not_found",
			"error":  "repository not found",
			"code":   errRepositoryNotFound,
		})
		return
	}
	opts := store.BuildRunListOptions{
		Status: c.Query("status"),
		Branch: c.Query("branch"),
	}
	if v := c.Query("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 200 {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "limit must be 1..200"})
			return
		}
		opts.Limit = n
	}
	if v := c.Query("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "offset must be >= 0"})
			return
		}
		opts.Offset = n
	}
	queryStart := time.Now()
	runs, total, err := storeI.ListRepositoryBuildRuns(c.Request.Context(), repoID, opts)
	queryDuration := time.Since(queryStart)
	if err != nil {
		writeServerError(c, err, "repository.build_runs")
		return
	}
	// #557: Prometheus histogram metric 관측 (status_filter label)
	observeBuildRunsQueryDuration(opts.Status, queryDuration)
	resp := make([]gin.H, 0, len(runs))
	for _, r := range runs {
		resp = append(resp, gin.H{
			"id":               r.ID,
			"repository_id":    r.RepositoryID,
			"run_external_id":  r.RunExternalID,
			"branch":           r.Branch,
			"commit_sha":       r.CommitSHA,
			"status":           r.Status,
			"duration_seconds": r.DurationSeconds,
			"started_at":       r.StartedAt.UTC().Format(time.RFC3339),
			"finished_at":      formatTimePtr(r.FinishedAt),
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   resp,
		"meta":   gin.H{"total": total},
	})
}

func (h *Handler) repositoryQualitySnapshots(c *gin.Context) {
	storeI, ok := h.platformStoreOrUnavailable(c)
	if !ok {
		return
	}
	repoID, err := strconv.ParseInt(c.Param("repository_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "repository_id must be an integer"})
		return
	}
	opts := store.QualitySnapshotListOptions{Tool: c.Query("tool")}
	if v := c.Query("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 200 {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "limit must be 1..200"})
			return
		}
		opts.Limit = n
	}
	if v := c.Query("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "offset must be >= 0"})
			return
		}
		opts.Offset = n
	}
	snapshots, total, err := storeI.ListRepositoryQualitySnapshots(c.Request.Context(), repoID, opts)
	if err != nil {
		writeServerError(c, err, "repository.quality_snapshots")
		return
	}
	resp := make([]gin.H, 0, len(snapshots))
	for _, q := range snapshots {
		resp = append(resp, gin.H{
			"id":              q.ID,
			"repository_id":   q.RepositoryID,
			"tool":            q.Tool,
			"ref_name":        q.RefName,
			"commit_sha":      q.CommitSHA,
			"score":           q.Score,
			"gate_passed":     q.GatePassed,
			"metric_payload":  q.MetricPayload,
			"measured_at":     q.MeasuredAt.UTC().Format(time.RFC3339),
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   resp,
		"meta":   gin.H{"total": total},
	})
}
// repositoryKPI handler — 단일 repository 의 raw KPI 종합 (가중치 없음, weight=1).
// Sprint A — kpi-tests-per-domain-scope.md §2.1 (Repository sub-section, raw).
// 데이터 소스: quality_snapshots (1 row, score 평균) + build_runs (success rate) + pr_activities (open/merged).
// 가중치 미적용 — 단일 repository 의 raw metric.
func (h *Handler) repositoryKPI(c *gin.Context) {
	storeI, ok := h.platformStoreOrUnavailable(c)
	if !ok {
		return
	}
	repoID, err := strconv.ParseInt(c.Param("repository_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "repository_id must be an integer"})
		return
	}
	opts := store.RepositoryActivityOptions{}
	if v := c.Query("from"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "from must be RFC3339"})
			return
		}
		opts.WindowFrom = t
	}
	if v := c.Query("to"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "to must be RFC3339"})
			return
		}
		opts.WindowTo = t
	}
	// quality_snapshots 의 평균 score + 마지막 measured_at 별도 조회 (ListRepositoryActivity 가
	// quality 는 종합하지 않으므로). 단일 row 의 score (avg of 1) 이 raw quality score.
	qualityOpts := store.QualitySnapshotListOptions{Limit: 1}
	snaps, _, err := storeI.ListRepositoryQualitySnapshots(c.Request.Context(), repoID, qualityOpts)
	if err != nil {
		writeServerError(c, err, "repository.kpi.quality")
		return
	}
	activity, err := storeI.ListRepositoryActivity(c.Request.Context(), repoID, opts)
	if err != nil {
		writeServerError(c, err, "repository.kpi.activity")
		return
	}
	openPR, mergedPR, err := storeI.CountOpenAndMergedPRs(c.Request.Context(), repoID, opts.WindowFrom, opts.WindowTo)
	if err != nil {
		writeServerError(c, err, "repository.kpi.pr")
		return
	}
	var qualityScore *float64
	var qualityMeasuredAt *time.Time
	if len(snaps) > 0 {
		qualityScore = snaps[0].Score
		qualityMeasuredAt = &snaps[0].MeasuredAt
	}

	var qs any
	if qualityScore != nil {
		qs = *qualityScore
	} else {
		qs = nil
	}
	var qmAt any
	if qualityMeasuredAt != nil {
		qmAt = qualityMeasuredAt.UTC().Format(time.RFC3339)
	}
	var bsr = activity.BuildSuccessRate
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data": gin.H{
			"repository_id":              repoID,
			"window_from":                activity.WindowFrom.UTC().Format(time.RFC3339),
			"window_to":                  activity.WindowTo.UTC().Format(time.RFC3339),
			"quality_score":              qs,
			"quality_score_measured_at":  qmAt,
			"build_success_rate":         bsr,
			"build_run_count":            activity.BuildRunCount,
			"open_pr_count":              openPR,
			"merged_pr_count":            mergedPR,
			"active_contributor_count":   len(activity.ActiveContributors),
		},
	})
}

// repositoryTestResults handler — 단일 repository 의 build_runs 기반 test results 분포.


// Sprint A — kpi-tests-per-domain-scope.md §2.1 (Repository sub-section, raw).
// 별도 repository_tests table 미존재 → build_runs 의 status 분포로 test results 표현.
// 별도 repository_tests table 도입은 후속 (ADR-0003 §6 carve out).
func (h *Handler) repositoryTestResults(c *gin.Context) {
	storeI, ok := h.platformStoreOrUnavailable(c)
	if !ok {
		return
	}
	repoID, err := strconv.ParseInt(c.Param("repository_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "repository_id must be an integer"})
		return
	}
	windowFrom, windowTo, err := parseTestResultsWindow(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": err.Error()})
		return
	}
	limit := 20
	if v := c.Query("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 50 {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "limit must be 1..50"})
			return
		}
		limit = n
	}
	// build_runs 전체 (status 필터 없음) — 분포 계산용
	allOpts := store.BuildRunListOptions{Limit: 200}
	all, total, err := storeI.ListRepositoryBuildRuns(c.Request.Context(), repoID, allOpts)
	if err != nil {
		writeServerError(c, err, "repository.test_results.list")
		return
	}
	totals := map[string]int{
		"success":   0,
		"failed":    0,
		"running":   0,
		"cancelled": 0,
		"skipped":   0,
		"queued":    0,
		"unknown":   0,
	}
	for _, r := range all {
		key := strings.ToLower(strings.TrimSpace(r.Status))
		if _, ok := totals[key]; !ok {
			totals["unknown"]++
			continue
		}
		totals[key]++
	}
	denom := totals["success"] + totals["failed"]
	var passRate *float64
	if denom > 0 {
		v := float64(totals["success"]) / float64(denom)
		passRate = &v
	}
	// recent — 최신순 limit 개. ListRepositoryBuildRuns 가 status filter 없으면
	// ORDER BY started_at DESC 가정 (기존 repositoryBuildRuns 정합).
	recent := make([]gin.H, 0, limit)
	for i, r := range all {
		if i >= limit {
			break
		}
		recent = append(recent, gin.H{
			"id":              r.ID,
			"run_external_id": r.RunExternalID,
			"commit_sha":      r.CommitSHA,
			"status":          r.Status,
			"branch":          r.Branch,
			"started_at":      r.StartedAt.UTC().Format(time.RFC3339),
			"finished_at":     formatTimePtr(r.FinishedAt),
		})
	}
	var pr any
	if passRate != nil {
		pr = *passRate
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data": gin.H{
			"repository_id": repoID,
			"window_from":   windowFrom.UTC().Format(time.RFC3339),
			"window_to":     windowTo.UTC().Format(time.RFC3339),
			"totals":        totals,
			"pass_rate":     pr,
			"recent":        recent,
		},
		"meta": gin.H{"total": total, "limit": limit},
	})
}

// parseTestResultsWindow extracts window_from / window_to from query string.
// Default 30d, range [1d, 365d]. window short string 도 parse (e.g. "7d", "30d", "90d", "1y").
func parseTestResultsWindow(c *gin.Context) (time.Time, time.Time, error) {
	now := time.Now().UTC()
	windowTo := now
	windowFrom := now.AddDate(0, 0, -30)
	if v := c.Query("window"); v != "" {
		dur, err := parseWindowShort(v)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		windowFrom = now.Add(-dur)
	} else {
		if v := c.Query("from"); v != "" {
			t, err := time.Parse(time.RFC3339, v)
			if err != nil {
				return time.Time{}, time.Time{}, errors.New("from must be RFC3339")
			}
			windowFrom = t
		}
		if v := c.Query("to"); v != "" {
			t, err := time.Parse(time.RFC3339, v)
			if err != nil {
				return time.Time{}, time.Time{}, errors.New("to must be RFC3339")
			}
			windowTo = t
		}
	}
	return windowFrom, windowTo, nil
}

// parseWindowShort parses "Nd"/"Nw"/"Nm"/"Ny" into time.Duration.
func parseWindowShort(s string) (time.Duration, error) {
	if len(s) < 2 {
		return 0, errors.New("window must be Nd/Nw/Nm/Ny (e.g. 30d)")
	}
	n, err := strconv.Atoi(s[:len(s)-1])
	if err != nil || n <= 0 {
		return 0, errors.New("window number must be a positive integer")
	}
	switch s[len(s)-1] {
	case 'd':
		return time.Duration(n) * 24 * time.Hour, nil
	case 'w':
		return time.Duration(n) * 7 * 24 * time.Hour, nil
	case 'm':
		return time.Duration(n) * 30 * 24 * time.Hour, nil
	case 'y':
		return time.Duration(n) * 365 * 24 * time.Hour, nil
	default:
		return 0, errors.New("window suffix must be d/w/m/y")
	}
}

// Ensure types used.
var _ domain.PRActivity
var _ store.RepositoryActivityOptions
