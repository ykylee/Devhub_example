package httpapi

import (
	"context"
	"time"
)

type SnapshotProvider interface {
	DashboardMetrics(context.Context, string) ([]metricResponse, bool, error)
	InfraNodes(context.Context) ([]serviceNodeResponse, error)
	InfraEdges(context.Context) ([]serviceEdgeResponse, error)
	CIRuns(context.Context) ([]ciRunResponse, error)
	CILogs(context.Context, string) ([]ciLogLineResponse, bool, error)
	CriticalRisks(context.Context) ([]riskResponse, error)
	Source() string
}

type StaticSnapshotProvider struct{}

func (StaticSnapshotProvider) DashboardMetrics(_ context.Context, role string) ([]metricResponse, bool, error) {
	metrics, ok := staticSnapshotMetrics()[role]
	return metrics, ok, nil
}

func (StaticSnapshotProvider) InfraNodes(context.Context) ([]serviceNodeResponse, error) {
	return staticSnapshotNodes(), nil
}

func (StaticSnapshotProvider) InfraEdges(context.Context) ([]serviceEdgeResponse, error) {
	return staticSnapshotEdges(), nil
}

func (StaticSnapshotProvider) CIRuns(context.Context) ([]ciRunResponse, error) {
	return staticSnapshotCIRuns(), nil
}

func (StaticSnapshotProvider) CILogs(_ context.Context, ciRunID string) ([]ciLogLineResponse, bool, error) {
	logs, ok := staticSnapshotCILogs()[ciRunID]
	return logs, ok, nil
}

func (StaticSnapshotProvider) CriticalRisks(context.Context) ([]riskResponse, error) {
	return staticSnapshotRisks(), nil
}

func (StaticSnapshotProvider) Source() string {
	return "static"
}

func staticSnapshotTime() time.Time {
	return time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)
}

func staticSnapshotMetrics() map[string][]metricResponse {
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

func staticSnapshotNodes() []serviceNodeResponse {
	now := staticSnapshotTime()
	return []serviceNodeResponse{
		{ID: "backend-core", Label: "Go Core Service", Kind: "service", Status: "stable", Region: "asia-01", CPUPercent: 12.4, MemoryBytes: 1288490189, ActiveInstances: 1, UpdatedAt: now},
		{ID: "gitea", Label: "Gitea Instance", Kind: "external", Status: "stable", Region: "asia-01", CPUPercent: 8.1, MemoryBytes: 858993459, ActiveInstances: 1, UpdatedAt: now},
		{ID: "backend-ai", Label: "Python AI Engine", Kind: "service", Status: "warning", Region: "asia-01", CPUPercent: 45.2, MemoryBytes: 4509715661, ActiveInstances: 1, UpdatedAt: now},
		{ID: "postgres", Label: "PostgreSQL Cluster", Kind: "database", Status: "stable", Region: "homelab", CPUPercent: 5.2, MemoryBytes: 2576980378, ActiveInstances: 1, UpdatedAt: now},
	}
}

func staticSnapshotEdges() []serviceEdgeResponse {
	now := staticSnapshotTime()
	return []serviceEdgeResponse{
		{ID: "gitea-backend-core", SourceID: "gitea", TargetID: "backend-core", Label: "WEBHOOK", Status: "stable", LatencyMS: 28.5, ThroughputRPS: 2.4, UpdatedAt: now},
		{ID: "backend-core-backend-ai", SourceID: "backend-core", TargetID: "backend-ai", Label: "gRPC", Status: "warning", LatencyMS: 42.7, ThroughputRPS: 0.8, UpdatedAt: now},
		{ID: "backend-core-postgres", SourceID: "backend-core", TargetID: "postgres", Label: "SQL", Status: "stable", LatencyMS: 9.3, ThroughputRPS: 12.1, UpdatedAt: now},
	}
}

func staticSnapshotCIRuns() []ciRunResponse {
	now := staticSnapshotTime()
	finished101 := now.Add(-2 * time.Minute)
	finished102 := now.Add(-7 * time.Minute)
	return []ciRunResponse{
		{ID: "101", RepositoryName: "devhub-core", Branch: "main", CommitSHA: "8a2f1b4", Status: "success", DurationSeconds: 134, StartedAt: now.Add(-5 * time.Minute), FinishedAt: &finished101},
		{ID: "102", RepositoryName: "devhub-core", Branch: "feat/auth", CommitSHA: "3c91a22", Status: "success", DurationSeconds: 105, StartedAt: now.Add(-9 * time.Minute), FinishedAt: &finished102},
		{ID: "103", RepositoryName: "devhub-core", Branch: "fix/deadlock", CommitSHA: "54ef9d0", Status: "failed", DurationSeconds: 190, StartedAt: now.Add(-3 * time.Minute)},
	}
}

func staticSnapshotCILogs() map[string][]ciLogLineResponse {
	now := staticSnapshotTime()
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

func staticSnapshotRisks() []riskResponse {
	now := staticSnapshotTime()
	return []riskResponse{
		{ID: "risk-001", Title: "Gitea Migration Blocked", Reason: "Access token expiration and scope mismatch detected in logs.", Impact: "high", Status: "action_required", OwnerLogin: "alex.k", SuggestedActions: []string{"rotate_token", "verify_scopes"}, CreatedAt: now.Add(-24 * time.Hour), UpdatedAt: now},
		{ID: "risk-002", Title: "Frontend CI Pipeline Delay", Reason: "Average build time increased by 45% in last 24h.", Impact: "medium", Status: "investigation", OwnerLogin: "yklee", SuggestedActions: []string{"scale_runners", "inspect_cache"}, CreatedAt: now.Add(-6 * time.Hour), UpdatedAt: now},
	}
}
