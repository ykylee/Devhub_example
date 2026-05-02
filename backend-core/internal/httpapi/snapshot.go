package httpapi

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type metricResponse struct {
	ID             string  `json:"id"`
	Label          string  `json:"label"`
	Value          string  `json:"value"`
	Trend          string  `json:"trend"`
	TrendDirection string  `json:"trend_direction"`
	NumericValue   float64 `json:"numeric_value,omitempty"`
	Unit           string  `json:"unit,omitempty"`
}

type serviceNodeResponse struct {
	ID              string    `json:"id"`
	Label           string    `json:"label"`
	Kind            string    `json:"kind"`
	Status          string    `json:"status"`
	Region          string    `json:"region,omitempty"`
	CPUPercent      float64   `json:"cpu_percent"`
	MemoryBytes     int64     `json:"memory_bytes"`
	ActiveInstances int       `json:"active_instances"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type serviceEdgeResponse struct {
	ID            string    `json:"id"`
	SourceID      string    `json:"source_id"`
	TargetID      string    `json:"target_id"`
	Label         string    `json:"label"`
	Status        string    `json:"status"`
	LatencyMS     float64   `json:"latency_ms"`
	ThroughputRPS float64   `json:"throughput_rps"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ciRunResponse struct {
	ID              string     `json:"id"`
	RepositoryName  string     `json:"repository_name"`
	Branch          string     `json:"branch"`
	CommitSHA       string     `json:"commit_sha"`
	Status          string     `json:"status"`
	DurationSeconds int        `json:"duration_seconds"`
	StartedAt       time.Time  `json:"started_at"`
	FinishedAt      *time.Time `json:"finished_at,omitempty"`
}

type ciLogLineResponse struct {
	ID        string    `json:"id"`
	CIRunID   string    `json:"ci_run_id"`
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	StepName  string    `json:"step_name"`
}

type riskResponse struct {
	ID               string    `json:"id"`
	Title            string    `json:"title"`
	Reason           string    `json:"reason"`
	Impact           string    `json:"impact"`
	Status           string    `json:"status"`
	OwnerLogin       string    `json:"owner_login"`
	SuggestedActions []string  `json:"suggested_actions"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (h Handler) dashboardMetrics(c *gin.Context) {
	role := c.Query("role")
	if role == "" {
		role = "developer"
	}

	metrics, ok := snapshotMetrics()[role]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "rejected",
			"error":  "role must be one of developer, manager, system_admin",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   metrics,
		"meta": gin.H{
			"role":  role,
			"count": len(metrics),
		},
	})
}

func (h Handler) infraNodes(c *gin.Context) {
	nodes := snapshotNodes()
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   nodes,
		"meta": gin.H{
			"count": len(nodes),
		},
	})
}

func (h Handler) infraEdges(c *gin.Context) {
	edges := snapshotEdges()
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   edges,
		"meta": gin.H{
			"count": len(edges),
		},
	})
}

func (h Handler) infraTopology(c *gin.Context) {
	nodes := snapshotNodes()
	edges := snapshotEdges()
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data": gin.H{
			"nodes": nodes,
			"edges": edges,
		},
		"meta": gin.H{
			"node_count": len(nodes),
			"edge_count": len(edges),
		},
	})
}

func (h Handler) ciRuns(c *gin.Context) {
	runs := snapshotCIRuns()
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   runs,
		"meta": gin.H{
			"count": len(runs),
			"scope": c.DefaultQuery("scope", "all"),
		},
	})
}

func (h Handler) ciRunLogs(c *gin.Context) {
	ciRunID := c.Param("ci_run_id")
	logsByRun := snapshotCILogs()
	logs, ok := logsByRun[ciRunID]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"status": "not_found",
			"error":  "ci run logs not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   logs,
		"meta": gin.H{
			"ci_run_id": ciRunID,
			"count":     len(logs),
		},
	})
}

func (h Handler) criticalRisks(c *gin.Context) {
	risks := snapshotRisks()
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   risks,
		"meta": gin.H{
			"count": len(risks),
		},
	})
}

func snapshotTime() time.Time {
	return time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)
}

func snapshotMetrics() map[string][]metricResponse {
	return map[string][]metricResponse{
		"developer": {
			{ID: "active_tasks", Label: "Active Tasks", Value: "3", Trend: "On Track", TrendDirection: "flat", NumericValue: 3, Unit: "count"},
			{ID: "build_success", Label: "Build Success", Value: "98%", Trend: "+2%", TrendDirection: "up", NumericValue: 98, Unit: "percent"},
			{ID: "code_review", Label: "Code Review", Value: "2", Trend: "Pending", TrendDirection: "flat", NumericValue: 2, Unit: "count"},
		},
		"manager": {
			{ID: "completion", Label: "Completion", Value: "72%", Trend: "+5%", TrendDirection: "up", NumericValue: 72, Unit: "percent"},
			{ID: "team_velocity", Label: "Team Velocity", Value: "48", Trend: "+12%", TrendDirection: "up", NumericValue: 48, Unit: "points"},
			{ID: "open_risks", Label: "Open Risks", Value: "2", Trend: "High", TrendDirection: "up", NumericValue: 2, Unit: "count"},
			{ID: "avg_cycle_time", Label: "Avg Cycle Time", Value: "4.2d", Trend: "-0.5d", TrendDirection: "down", NumericValue: 4.2, Unit: "days"},
		},
		"system_admin": {
			{ID: "availability", Label: "Availability", Value: "99.99%", Trend: "Stable", TrendDirection: "flat", NumericValue: 99.99, Unit: "percent"},
			{ID: "active_runners", Label: "Active Runners", Value: "12/12", Trend: "Full", TrendDirection: "flat", NumericValue: 12, Unit: "count"},
			{ID: "ai_engine_load", Label: "AI Engine Load", Value: "24%", Trend: "Low", TrendDirection: "down", NumericValue: 24, Unit: "percent"},
			{ID: "storage", Label: "Storage", Value: "1.2TB", Trend: "82%", TrendDirection: "up", NumericValue: 82, Unit: "percent"},
		},
	}
}

func snapshotNodes() []serviceNodeResponse {
	now := snapshotTime()
	return []serviceNodeResponse{
		{ID: "backend-core", Label: "Go Core Service", Kind: "service", Status: "stable", Region: "asia-01", CPUPercent: 12.4, MemoryBytes: 1288490189, ActiveInstances: 1, UpdatedAt: now},
		{ID: "gitea", Label: "Gitea Instance", Kind: "external", Status: "stable", Region: "asia-01", CPUPercent: 8.1, MemoryBytes: 858993459, ActiveInstances: 1, UpdatedAt: now},
		{ID: "backend-ai", Label: "Python AI Engine", Kind: "service", Status: "warning", Region: "asia-01", CPUPercent: 45.2, MemoryBytes: 4509715661, ActiveInstances: 1, UpdatedAt: now},
		{ID: "postgres", Label: "PostgreSQL Cluster", Kind: "database", Status: "stable", Region: "homelab", CPUPercent: 5.2, MemoryBytes: 2576980378, ActiveInstances: 1, UpdatedAt: now},
	}
}

func snapshotEdges() []serviceEdgeResponse {
	now := snapshotTime()
	return []serviceEdgeResponse{
		{ID: "gitea-backend-core", SourceID: "gitea", TargetID: "backend-core", Label: "WEBHOOK", Status: "stable", LatencyMS: 28.5, ThroughputRPS: 2.4, UpdatedAt: now},
		{ID: "backend-core-backend-ai", SourceID: "backend-core", TargetID: "backend-ai", Label: "gRPC", Status: "warning", LatencyMS: 42.7, ThroughputRPS: 0.8, UpdatedAt: now},
		{ID: "backend-core-postgres", SourceID: "backend-core", TargetID: "postgres", Label: "SQL", Status: "stable", LatencyMS: 9.3, ThroughputRPS: 12.1, UpdatedAt: now},
	}
}

func snapshotCIRuns() []ciRunResponse {
	now := snapshotTime()
	finished101 := now.Add(-2 * time.Minute)
	finished102 := now.Add(-7 * time.Minute)
	return []ciRunResponse{
		{ID: "101", RepositoryName: "devhub-core", Branch: "main", CommitSHA: "8a2f1b4", Status: "success", DurationSeconds: 134, StartedAt: now.Add(-5 * time.Minute), FinishedAt: &finished101},
		{ID: "102", RepositoryName: "devhub-core", Branch: "feat/auth", CommitSHA: "3c91a22", Status: "success", DurationSeconds: 105, StartedAt: now.Add(-9 * time.Minute), FinishedAt: &finished102},
		{ID: "103", RepositoryName: "devhub-core", Branch: "fix/deadlock", CommitSHA: "54ef9d0", Status: "failed", DurationSeconds: 190, StartedAt: now.Add(-3 * time.Minute)},
	}
}

func snapshotCILogs() map[string][]ciLogLineResponse {
	now := snapshotTime()
	return map[string][]ciLogLineResponse{
		"101": {
			{ID: "101-1", CIRunID: "101", Timestamp: now.Add(-4 * time.Minute), Level: "info", Message: "checkout completed", StepName: "checkout"},
			{ID: "101-2", CIRunID: "101", Timestamp: now.Add(-3 * time.Minute), Level: "info", Message: "go test ./... passed", StepName: "test"},
		},
		"102": {
			{ID: "102-1", CIRunID: "102", Timestamp: now.Add(-8 * time.Minute), Level: "info", Message: "frontend lint passed", StepName: "lint"},
		},
		"103": {
			{ID: "103-1", CIRunID: "103", Timestamp: now.Add(-2 * time.Minute), Level: "error", Message: "deadlock regression test failed", StepName: "test"},
		},
	}
}

func snapshotRisks() []riskResponse {
	now := snapshotTime()
	return []riskResponse{
		{ID: "risk-001", Title: "Gitea Migration Blocked", Reason: "Access token expiration and scope mismatch detected in logs.", Impact: "high", Status: "action_required", OwnerLogin: "alex.k", SuggestedActions: []string{"rotate_token", "verify_scopes"}, CreatedAt: now.Add(-24 * time.Hour), UpdatedAt: now},
		{ID: "risk-002", Title: "Frontend CI Pipeline Delay", Reason: "Average build time increased by 45% in last 24h.", Impact: "medium", Status: "investigation", OwnerLogin: "yklee", SuggestedActions: []string{"scale_runners", "inspect_cache"}, CreatedAt: now.Add(-6 * time.Hour), UpdatedAt: now},
	}
}
