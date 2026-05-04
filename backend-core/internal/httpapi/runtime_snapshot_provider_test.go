package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type pingStore struct {
	err error
}

func (s pingStore) Ping(context.Context) error {
	return s.err
}

func TestRuntimeSnapshotProviderUpdatesInfraNodeStatuses(t *testing.T) {
	giteaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer giteaServer.Close()

	aiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Fatalf("expected /health path, got %q", r.URL.Path)
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer aiServer.Close()

	now := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)
	provider := RuntimeSnapshotProvider{
		Base:         StaticSnapshotProvider{},
		HealthStore:  pingStore{},
		GiteaURL:     giteaServer.URL,
		BackendAIURL: aiServer.URL,
		Now:          func() time.Time { return now },
	}

	nodes, err := provider.InfraNodes(context.Background())
	if err != nil {
		t.Fatalf("infra nodes: %v", err)
	}
	statuses := nodeStatuses(nodes)
	if statuses["backend-core"] != "stable" {
		t.Fatalf("expected backend-core stable, got %q", statuses["backend-core"])
	}
	if statuses["postgres"] != "stable" {
		t.Fatalf("expected postgres stable, got %q", statuses["postgres"])
	}
	if statuses["gitea"] != "stable" {
		t.Fatalf("expected gitea stable, got %q", statuses["gitea"])
	}
	if statuses["backend-ai"] != "down" {
		t.Fatalf("expected backend-ai down, got %q", statuses["backend-ai"])
	}
	if updatedAt := nodeUpdatedAt(nodes, "backend-ai"); !updatedAt.Equal(now) {
		t.Fatalf("expected updated_at %s, got %s", now, updatedAt)
	}
}

func TestRuntimeSnapshotProviderMarksEdgesWarningOrDown(t *testing.T) {
	provider := RuntimeSnapshotProvider{
		Base:        StaticSnapshotProvider{},
		HealthStore: pingStore{err: errors.New("db down")},
		Now:         func() time.Time { return time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC) },
	}

	edges, err := provider.InfraEdges(context.Background())
	if err != nil {
		t.Fatalf("infra edges: %v", err)
	}
	statuses := edgeStatuses(edges)
	if statuses["backend-core-postgres"] != "down" {
		t.Fatalf("expected backend-core-postgres down, got %q", statuses["backend-core-postgres"])
	}
	if statuses["gitea-backend-core"] != "warning" {
		t.Fatalf("expected gitea-backend-core warning when gitea url missing, got %q", statuses["gitea-backend-core"])
	}
}

func TestRuntimeSnapshotProviderDelegatesMetrics(t *testing.T) {
	provider := RuntimeSnapshotProvider{Base: testSnapshotProvider{source: "inner"}}

	metrics, ok, err := provider.DashboardMetrics(context.Background(), "developer")
	if err != nil {
		t.Fatalf("dashboard metrics: %v", err)
	}
	if !ok || len(metrics) != 1 || metrics[0].ID != "custom_metric" {
		t.Fatalf("expected delegated metrics, got ok=%v metrics=%+v", ok, metrics)
	}
	if provider.Source() != "runtime" {
		t.Fatalf("expected runtime source, got %q", provider.Source())
	}
}

func nodeStatuses(nodes []serviceNodeResponse) map[string]string {
	statuses := make(map[string]string, len(nodes))
	for _, node := range nodes {
		statuses[node.ID] = node.Status
	}
	return statuses
}

func nodeUpdatedAt(nodes []serviceNodeResponse, id string) time.Time {
	for _, node := range nodes {
		if node.ID == id {
			return node.UpdatedAt
		}
	}
	return time.Time{}
}

func edgeStatuses(edges []serviceEdgeResponse) map[string]string {
	statuses := make(map[string]string, len(edges))
	for _, edge := range edges {
		statuses[edge.ID] = edge.Status
	}
	return statuses
}
