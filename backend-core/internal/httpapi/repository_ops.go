package httpapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

// Repository 운영 지표 read endpoint (API-51..54, sprint claude/work_260514-c).
// pr_activities / build_runs / quality_snapshots 의 read-only 조회. write 는 ingest
// pipeline 책임 (별도 sprint).

func (h *Handler) repositoryActivity(c *gin.Context) {
	storeI, ok := h.applicationStoreOrUnavailable(c)
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
		},
	})
}

func (h *Handler) repositoryPullRequests(c *gin.Context) {
	storeI, ok := h.applicationStoreOrUnavailable(c)
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
	storeI, ok := h.applicationStoreOrUnavailable(c)
	if !ok {
		return
	}
	repoID, err := strconv.ParseInt(c.Param("repository_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "repository_id must be an integer"})
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
	runs, total, err := storeI.ListRepositoryBuildRuns(c.Request.Context(), repoID, opts)
	if err != nil {
		writeServerError(c, err, "repository.build_runs")
		return
	}
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
	storeI, ok := h.applicationStoreOrUnavailable(c)
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

// Ensure types used.
var _ domain.PRActivity
var _ store.RepositoryActivityOptions
