package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInfraServicesSnapshotIngestAndRead(t *testing.T) {
	runtimeInfraSnapshots = infraSnapshotState{}
	router := NewRouter(RouterConfig{
		AuthDevFallback: true,
		InfraAgentToken: "agent-secret",
	})

	payload := `{
		"agent_id":"homelab-agent-a",
		"snapshot_at":"2026-05-15T14:10:00Z",
		"services":[
			{
				"service_id":"svc-jenkins",
				"node_id":"node-nas-01",
				"name":"jenkins",
				"version":"2.504.1",
				"port":8080,
				"health_status":"healthy",
				"observed_at":"2026-05-15T14:09:59Z"
			}
		]
	}`
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/infra/services/snapshot", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer agent-secret")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("snapshot ingest status=%d body=%s", rec.Code, rec.Body.String())
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/infra/services", nil)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("infra services status=%d body=%s", getRec.Code, getRec.Body.String())
	}
	if !bytes.Contains(getRec.Body.Bytes(), []byte(`"service_id":"svc-jenkins"`)) {
		t.Fatalf("expected ingested service, body=%s", getRec.Body.String())
	}
}

func TestInfraServicesSnapshotRejectsUnauthorized(t *testing.T) {
	runtimeInfraSnapshots = infraSnapshotState{}
	router := NewRouter(RouterConfig{
		AuthDevFallback: true,
		InfraAgentToken: "agent-secret",
	})
	rec := doJSON(t, router, http.MethodPost, "/api/v1/infra/services/snapshot",
		`{"agent_id":"homelab-agent-a","snapshot_at":"2026-05-15T14:10:00Z","services":[{"service_id":"svc-jenkins","node_id":"node-nas-01","name":"jenkins","health_status":"healthy","observed_at":"2026-05-15T14:09:59Z"}]}`)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestInfraTopologyV2ContainsMeta(t *testing.T) {
	runtimeInfraSnapshots = infraSnapshotState{}
	router := NewRouter(RouterConfig{AuthDevFallback: true})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/infra/topology/v2", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var response struct {
		Meta struct {
			SnapshotAt string   `json:"snapshot_at"`
			Degraded   []string `json:"degraded_providers"`
		} `json:"meta"`
		Data struct {
			Services []json.RawMessage `json:"services"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Meta.SnapshotAt == "" {
		t.Fatalf("expected snapshot_at in meta: %s", rec.Body.String())
	}
	if len(response.Data.Services) == 0 {
		t.Fatalf("expected services in topology v2: %s", rec.Body.String())
	}
}

func TestInfraServicesHydratesFromPersistedSnapshot(t *testing.T) {
	runtimeInfraSnapshots = infraSnapshotState{}
	appStore := newMemoryApplicationStore()
	router := NewRouter(RouterConfig{
		ApplicationStore: appStore,
		AuthDevFallback:  true,
		InfraAgentToken:  "agent-secret",
	})

	payload := `{
		"agent_id":"homelab-agent-a",
		"snapshot_at":"2026-05-15T14:10:00Z",
		"services":[
			{
				"service_id":"svc-gitea",
				"node_id":"node-nas-01",
				"name":"gitea",
				"version":"1.22.0",
				"port":3000,
				"health_status":"degraded",
				"observed_at":"2026-05-15T14:09:59Z"
			}
		]
	}`
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/infra/services/snapshot", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer agent-secret")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("snapshot ingest status=%d body=%s", rec.Code, rec.Body.String())
	}

	// emulate process restart: runtime cache is empty, but store snapshot should hydrate.
	runtimeInfraSnapshots = infraSnapshotState{}
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/infra/services", nil)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("infra services status=%d body=%s", getRec.Code, getRec.Body.String())
	}
	if !bytes.Contains(getRec.Body.Bytes(), []byte(`"service_id":"svc-gitea"`)) {
		t.Fatalf("expected persisted service after hydrate, body=%s", getRec.Body.String())
	}
}

func TestHomeLabHealthPolicyFromConfig_Defaults(t *testing.T) {
	p := homeLabHealthPolicyFromConfig("", "")
	if p.ProviderKey != "homelab-agent" {
		t.Fatalf("provider key=%q want homelab-agent", p.ProviderKey)
	}
	if !p.DegradedStatuses["warning"] || !p.DegradedStatuses["degraded"] || !p.DegradedStatuses["down"] {
		t.Fatalf("default degraded statuses mismatch: %+v", p.DegradedStatuses)
	}
}

func TestHomeLabHealthPolicyFromConfig_Override(t *testing.T) {
	p := homeLabHealthPolicyFromConfig("lab-custom", "critical,down")
	if p.ProviderKey != "lab-custom" {
		t.Fatalf("provider key=%q want lab-custom", p.ProviderKey)
	}
	if !p.DegradedStatuses["critical"] || !p.DegradedStatuses["down"] {
		t.Fatalf("override statuses missing: %+v", p.DegradedStatuses)
	}
	if p.DegradedStatuses["warning"] {
		t.Fatalf("warning should not be enabled in override: %+v", p.DegradedStatuses)
	}
}
