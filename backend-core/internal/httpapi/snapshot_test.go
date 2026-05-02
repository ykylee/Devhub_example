package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestDashboardMetricsAcceptsRoleWireFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(RouterConfig{})

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

func TestDashboardMetricsRejectsUnknownRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(RouterConfig{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/metrics?role=System%20Admin", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestInfraTopologyReturnsNodesAndEdges(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(RouterConfig{})

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

func TestCIRunLogsReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(RouterConfig{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ci-runs/unknown/logs", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestCriticalRisksReturnsActionableFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(RouterConfig{})

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
