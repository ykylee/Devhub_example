package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/gin-gonic/gin"
)

func TestDashboardMetricsAcceptsRoleWireFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := testRouter(RouterConfig{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/metrics?role=system_admin", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var response struct {
		Status string `json:"status"`
		Data   []struct {
			ID             string  `json:"id"`
			Label          string  `json:"label"`
			NumericValue   float64 `json:"numeric_value"`
			TrendDirection string  `json:"trend_direction"`
			Unit           string  `json:"unit"`
		} `json:"data"`
		Meta struct {
			Role  string `json:"role"`
			Count int    `json:"count"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Status != "ok" || response.Meta.Role != "system_admin" {
		t.Fatalf("unexpected response meta: %+v", response.Meta)
	}
	if len(response.Data) == 0 {
		t.Fatal("expected metrics")
	}
	if response.Data[0].ID == "" || response.Data[0].Unit == "" || response.Data[0].TrendDirection == "" {
		t.Fatalf("metric missing stable fields: %+v", response.Data[0])
	}
}

func TestDashboardMetricsUsesInjectedSnapshotProvider(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := testRouter(RouterConfig{SnapshotProvider: testSnapshotProvider{source: "test-provider"}})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/metrics?role=developer", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var response struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
		Meta struct {
			Source string `json:"source"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Meta.Source != "test-provider" {
		t.Fatalf("expected injected source, got %q", response.Meta.Source)
	}
	if len(response.Data) != 1 || response.Data[0].ID != "custom_metric" {
		t.Fatalf("unexpected metrics: %+v", response.Data)
	}
}

func TestDashboardMetricsRejectsUnknownRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := testRouter(RouterConfig{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/metrics?role=System%20Admin", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestInfraTopologyReturnsNodesAndEdges(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := testRouter(RouterConfig{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/infra/topology", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var response struct {
		Status string `json:"status"`
		Data   struct {
			Nodes []struct {
				ID          string  `json:"id"`
				Status      string  `json:"status"`
				CPUPercent  float64 `json:"cpu_percent"`
				MemoryBytes int64   `json:"memory_bytes"`
			} `json:"nodes"`
			Edges []struct {
				ID        string  `json:"id"`
				LatencyMS float64 `json:"latency_ms"`
			} `json:"edges"`
		} `json:"data"`
		Meta struct {
			NodeCount int `json:"node_count"`
			EdgeCount int `json:"edge_count"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Meta.NodeCount != len(response.Data.Nodes) || response.Meta.EdgeCount != len(response.Data.Edges) {
		t.Fatalf("meta counts do not match data: %+v", response.Meta)
	}
	if len(response.Data.Nodes) == 0 || response.Data.Nodes[0].CPUPercent == 0 || response.Data.Nodes[0].MemoryBytes == 0 {
		t.Fatalf("unexpected topology node: %+v", response.Data.Nodes)
	}
}

func TestInfraTopologyUsesInjectedSnapshotProvider(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := testRouter(RouterConfig{SnapshotProvider: testSnapshotProvider{source: "adapter"}})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/infra/topology", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var response struct {
		Data struct {
			Nodes []struct {
				ID string `json:"id"`
			} `json:"nodes"`
			Edges []struct {
				ID string `json:"id"`
			} `json:"edges"`
		} `json:"data"`
		Meta struct {
			Source string `json:"source"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Meta.Source != "adapter" {
		t.Fatalf("expected adapter source, got %q", response.Meta.Source)
	}
	if len(response.Data.Nodes) != 1 || response.Data.Nodes[0].ID != "custom-node" {
		t.Fatalf("unexpected nodes: %+v", response.Data.Nodes)
	}
	if len(response.Data.Edges) != 1 || response.Data.Edges[0].ID != "custom-edge" {
		t.Fatalf("unexpected edges: %+v", response.Data.Edges)
	}
}

func TestCIRunLogsReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := testRouter(RouterConfig{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ci-runs/unknown/logs", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestCriticalRisksReturnsActionableFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := testRouter(RouterConfig{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/risks/critical", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var response struct {
		Status string `json:"status"`
		Data   []struct {
			ID               string   `json:"id"`
			Impact           string   `json:"impact"`
			Status           string   `json:"status"`
			OwnerLogin       string   `json:"owner_login"`
			SuggestedActions []string `json:"suggested_actions"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Status != "ok" || len(response.Data) == 0 {
		t.Fatalf("unexpected response: %+v", response)
	}
	if response.Data[0].ID == "" || response.Data[0].OwnerLogin == "" || len(response.Data[0].SuggestedActions) == 0 {
		t.Fatalf("risk missing actionable fields: %+v", response.Data[0])
	}
}

func TestCriticalRisksUsesDBWhenAvailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Date(2026, 5, 3, 10, 0, 0, 0, time.UTC)
	router := testRouter(RouterConfig{DomainStore: &memoryDomainStore{
		risks: []domain.Risk{
			{
				RiskKey:          "ci_failure:502",
				Title:            "CI run failed for acme/api",
				Reason:           "CI run 502 failed on branch main",
				Impact:           "high",
				Status:           "action_required",
				SuggestedActions: []string{"inspect_logs", "rerun_ci"},
				CreatedAt:        now,
				UpdatedAt:        now,
			},
		},
	}})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/risks/critical", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
	var response struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
		Meta struct {
			Source string `json:"source"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Meta.Source != "db" || len(response.Data) != 1 || response.Data[0].ID != "ci_failure:502" {
		t.Fatalf("unexpected db-backed risks response: %+v", response)
	}
}

type testSnapshotProvider struct {
	source string
}

func (p testSnapshotProvider) DashboardMetrics(context.Context, string) ([]metricResponse, bool, error) {
	return []metricResponse{
		{ID: "custom_metric", Label: "Custom Metric", Value: "1", Trend: "Flat", TrendDirection: "flat", NumericValue: 1, Unit: "count"},
	}, true, nil
}

func (p testSnapshotProvider) InfraNodes(context.Context) ([]serviceNodeResponse, error) {
	now := time.Date(2026, 5, 3, 10, 0, 0, 0, time.UTC)
	return []serviceNodeResponse{
		{ID: "custom-node", Label: "Custom Node", Kind: "service", Status: "stable", CPUPercent: 1, MemoryBytes: 1, ActiveInstances: 1, UpdatedAt: now},
	}, nil
}

func (p testSnapshotProvider) InfraEdges(context.Context) ([]serviceEdgeResponse, error) {
	now := time.Date(2026, 5, 3, 10, 0, 0, 0, time.UTC)
	return []serviceEdgeResponse{
		{ID: "custom-edge", SourceID: "custom-node", TargetID: "custom-target", Label: "HTTP", Status: "stable", UpdatedAt: now},
	}, nil
}

func (p testSnapshotProvider) CIRuns(context.Context) ([]ciRunResponse, error) {
	now := time.Date(2026, 5, 3, 10, 0, 0, 0, time.UTC)
	return []ciRunResponse{
		{ID: "custom-run", RepositoryName: "acme/api", Branch: "main", Status: "success", StartedAt: now},
	}, nil
}

func (p testSnapshotProvider) CILogs(context.Context, string) ([]ciLogLineResponse, bool, error) {
	now := time.Date(2026, 5, 3, 10, 0, 0, 0, time.UTC)
	return []ciLogLineResponse{
		{ID: "log-1", CIRunID: "custom-run", Timestamp: now, Level: "info", Message: "ok", StepName: "test"},
	}, true, nil
}

func (p testSnapshotProvider) CriticalRisks(context.Context) ([]riskResponse, error) {
	now := time.Date(2026, 5, 3, 10, 0, 0, 0, time.UTC)
	return []riskResponse{
		{ID: "risk-custom", Title: "Custom Risk", Reason: "test", Impact: "medium", Status: "investigation", OwnerLogin: "yklee", SuggestedActions: []string{"inspect"}, CreatedAt: now, UpdatedAt: now},
	}, nil
}

func (p testSnapshotProvider) Source() string {
	return p.source
}
