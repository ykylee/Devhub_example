package httpapi

import (
	"context"
	"net/http"
	"strings"
	"time"
)

type RuntimeSnapshotProvider struct {
	Base         SnapshotProvider
	HealthStore  HealthStore
	GiteaURL     string
	BackendAIURL string
	HTTPClient   *http.Client
	Now          func() time.Time
}

func (p RuntimeSnapshotProvider) DashboardMetrics(ctx context.Context, role string) ([]metricResponse, bool, error) {
	return p.base().DashboardMetrics(ctx, role)
}

func (p RuntimeSnapshotProvider) InfraNodes(ctx context.Context) ([]serviceNodeResponse, error) {
	nodes, err := p.base().InfraNodes(ctx)
	if err != nil {
		return nil, err
	}

	statuses := p.runtimeStatuses(ctx)
	now := p.now()
	for i := range nodes {
		if status, ok := statuses[nodes[i].ID]; ok {
			nodes[i].Status = status
			nodes[i].UpdatedAt = now
		}
	}
	return nodes, nil
}

func (p RuntimeSnapshotProvider) InfraEdges(ctx context.Context) ([]serviceEdgeResponse, error) {
	edges, err := p.base().InfraEdges(ctx)
	if err != nil {
		return nil, err
	}

	statuses := p.runtimeStatuses(ctx)
	now := p.now()
	for i := range edges {
		sourceStatus := statuses[edges[i].SourceID]
		targetStatus := statuses[edges[i].TargetID]
		if sourceStatus == "down" || targetStatus == "down" {
			edges[i].Status = "down"
		} else if sourceStatus == "warning" || targetStatus == "warning" || sourceStatus == "degraded" || targetStatus == "degraded" {
			edges[i].Status = "warning"
		} else if sourceStatus != "" || targetStatus != "" {
			edges[i].Status = "stable"
		}
		edges[i].UpdatedAt = now
	}
	return edges, nil
}

func (p RuntimeSnapshotProvider) CIRuns(ctx context.Context) ([]ciRunResponse, error) {
	return p.base().CIRuns(ctx)
}

func (p RuntimeSnapshotProvider) CILogs(ctx context.Context, ciRunID string) ([]ciLogLineResponse, bool, error) {
	return p.base().CILogs(ctx, ciRunID)
}

func (p RuntimeSnapshotProvider) CriticalRisks(ctx context.Context) ([]riskResponse, error) {
	return p.base().CriticalRisks(ctx)
}

func (p RuntimeSnapshotProvider) Source() string {
	return "runtime"
}

func (p RuntimeSnapshotProvider) base() SnapshotProvider {
	if p.Base != nil {
		return p.Base
	}
	return StaticSnapshotProvider{}
}

func (p RuntimeSnapshotProvider) now() time.Time {
	if p.Now != nil {
		return p.Now().UTC()
	}
	return time.Now().UTC()
}

func (p RuntimeSnapshotProvider) runtimeStatuses(ctx context.Context) map[string]string {
	return map[string]string{
		"backend-core": "stable",
		"postgres":     p.postgresStatus(ctx),
		"gitea":        p.httpStatus(ctx, p.GiteaURL, ""),
		"backend-ai":   p.httpStatus(ctx, p.BackendAIURL, "/health"),
	}
}

func (p RuntimeSnapshotProvider) postgresStatus(ctx context.Context) string {
	if p.HealthStore == nil {
		return "warning"
	}
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := p.HealthStore.Ping(pingCtx); err != nil {
		return "down"
	}
	return "stable"
}

func (p RuntimeSnapshotProvider) httpStatus(ctx context.Context, baseURL, healthPath string) string {
	if strings.TrimSpace(baseURL) == "" {
		return "warning"
	}
	target := strings.TrimRight(baseURL, "/") + healthPath
	reqCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, target, nil)
	if err != nil {
		return "warning"
	}

	client := p.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 2 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return "down"
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusInternalServerError {
		return "stable"
	}
	return "down"
}
