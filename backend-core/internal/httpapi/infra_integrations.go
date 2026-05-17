package httpapi

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/devhub/backend-core/internal/integrations/adapters"
	"github.com/gin-gonic/gin"
)

type infraServiceRecord struct {
	ServiceID    string `json:"service_id"`
	NodeID       string `json:"node_id"`
	Name         string `json:"name"`
	Version      string `json:"version"`
	Port         int    `json:"port"`
	HealthStatus string `json:"health_status"`
	ObservedAt   string `json:"observed_at"`
}

type infraNodeRecord struct {
	NodeID      string         `json:"node_id"`
	Hostname    string         `json:"hostname"`
	IPAddress   string         `json:"ip_address"`
	Environment string         `json:"environment"`
	Status      string         `json:"status"`
	Metrics     map[string]any `json:"metrics,omitempty"`
	ObservedAt  string         `json:"observed_at"`
}

type infraSnapshotRequest struct {
	AgentID    string               `json:"agent_id"`
	SnapshotAt string               `json:"snapshot_at"`
	TraceID    string               `json:"trace_id"`
	Nodes      []infraNodeRecord    `json:"nodes"`
	Services   []infraServiceRecord `json:"services"`
}

type infraSnapshotState struct {
	mu           sync.RWMutex
	snapshotAt   string
	nodes        []infraNodeRecord
	services     []infraServiceRecord
	degradedFrom []string
}

var runtimeInfraSnapshots infraSnapshotState

type infraSnapshotPersistence interface {
	SaveInfraSnapshot(ctx context.Context, ingestID, agentID string, snapshotAt time.Time, traceID string, nodesJSON, servicesJSON []byte, degradedProviders []string) error
	LoadLatestInfraSnapshot(ctx context.Context) (snapshotAt time.Time, nodesJSON, servicesJSON []byte, degradedProviders []string, err error)
}

func (h *Handler) listInfraServices(c *gin.Context) {
	h.hydrateRuntimeInfraSnapshot(c.Request.Context())
	services := currentInfraServices()
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   services,
		"meta":   gin.H{"count": len(services)},
	})
}

func (h *Handler) ingestInfraServicesSnapshot(c *gin.Context) {
	token := strings.TrimSpace(h.cfg.InfraAgentToken)
	if token == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "infra agent ingest is not configured",
		})
		return
	}
	authorization := strings.TrimSpace(c.GetHeader("Authorization"))
	if !strings.HasPrefix(authorization, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "rejected",
			"error":  "infra agent unauthorized",
			"code":   "infra_agent_unauthorized",
		})
		return
	}
	presented := strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer "))
	if subtle.ConstantTimeCompare([]byte(presented), []byte(token)) != 1 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "rejected",
			"error":  "infra agent unauthorized",
			"code":   "infra_agent_unauthorized",
		})
		return
	}

	var req infraSnapshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": err.Error(), "code": "infra_snapshot_invalid"})
		return
	}
	if strings.TrimSpace(req.AgentID) == "" || strings.TrimSpace(req.SnapshotAt) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "rejected",
			"error":  "agent_id and snapshot_at are required",
			"code":   "infra_snapshot_invalid",
		})
		return
	}
	if _, err := time.Parse(time.RFC3339, req.SnapshotAt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "rejected",
			"error":  "snapshot_at must be RFC3339",
			"code":   "infra_snapshot_invalid",
		})
		return
	}
	if len(req.Nodes) == 0 && len(req.Services) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "rejected",
			"error":  "nodes or services must be provided",
			"code":   "infra_snapshot_invalid",
		})
		return
	}

	ingestID := "ing_" + strings.ReplaceAll(time.Now().UTC().Format("20060102T150405.000000000"), ".", "")
	degradedProviders := collectDegradedProviders(req.Services)
	if adapter, ok := h.homeLabAdapter(); ok {
		raw, err := toHomeLabRawSnapshot(req)
		if err == nil {
			normalized, nErr := adapter.NormalizeSnapshot(raw)
			if nErr == nil {
				ingestID = normalized.IngestID
				degradedProviders = normalized.DegradedProviders
				_ = adapter.IngestSnapshot(c.Request.Context(), normalized)
			}
		}
	}
	runtimeInfraSnapshots.mu.Lock()
	runtimeInfraSnapshots.snapshotAt = req.SnapshotAt
	runtimeInfraSnapshots.nodes = append([]infraNodeRecord(nil), req.Nodes...)
	runtimeInfraSnapshots.services = append([]infraServiceRecord(nil), req.Services...)
	runtimeInfraSnapshots.degradedFrom = degradedProviders
	runtimeInfraSnapshots.mu.Unlock()
	if _, ok := h.homeLabAdapter(); !ok {
		h.persistInfraSnapshotBestEffort(c.Request.Context(), ingestID, req, degradedProviders)
	}

	c.JSON(http.StatusAccepted, gin.H{
		"status":    "accepted",
		"ingest_id": ingestID,
	})
}

func (h *Handler) infraTopologyV2(c *gin.Context) {
	h.hydrateRuntimeInfraSnapshot(c.Request.Context())
	nodes, edges := currentInfraTopologyV2()
	services := currentInfraServices()
	runtimeInfraSnapshots.mu.RLock()
	snapshotAt := runtimeInfraSnapshots.snapshotAt
	degradedProviders := append([]string(nil), runtimeInfraSnapshots.degradedFrom...)
	runtimeInfraSnapshots.mu.RUnlock()
	if snapshotAt == "" {
		snapshotAt = time.Now().UTC().Format(time.RFC3339)
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data": gin.H{
			"nodes":    nodes,
			"edges":    edges,
			"services": services,
		},
		"meta": gin.H{
			"snapshot_at":        snapshotAt,
			"degraded_providers": degradedProviders,
		},
	})
}

func (h *Handler) persistInfraSnapshotBestEffort(ctx context.Context, ingestID string, req infraSnapshotRequest, degradedProviders []string) {
	saver, ok := h.cfg.ApplicationStore.(infraSnapshotPersistence)
	if !ok {
		return
	}
	snapshotAt, err := time.Parse(time.RFC3339, req.SnapshotAt)
	if err != nil {
		return
	}
	nodesJSON, err := json.Marshal(req.Nodes)
	if err != nil {
		return
	}
	servicesJSON, err := json.Marshal(req.Services)
	if err != nil {
		return
	}
	_ = saver.SaveInfraSnapshot(ctx, ingestID, strings.TrimSpace(req.AgentID), snapshotAt, strings.TrimSpace(req.TraceID), nodesJSON, servicesJSON, degradedProviders)
}

func (h *Handler) hydrateRuntimeInfraSnapshot(ctx context.Context) {
	runtimeInfraSnapshots.mu.RLock()
	hasRuntime := runtimeInfraSnapshots.snapshotAt != "" || len(runtimeInfraSnapshots.nodes) > 0 || len(runtimeInfraSnapshots.services) > 0
	runtimeInfraSnapshots.mu.RUnlock()
	if hasRuntime {
		return
	}
	if adapter, ok := h.homeLabAdapter(); ok {
		snapshot, err := adapter.LoadLatestSnapshot(ctx)
		if err == nil {
			var nodes []infraNodeRecord
			var services []infraServiceRecord
			if len(snapshot.NodesJSON) > 0 {
				if uErr := json.Unmarshal(snapshot.NodesJSON, &nodes); uErr != nil {
					return
				}
			}
			if len(snapshot.ServicesJSON) > 0 {
				if uErr := json.Unmarshal(snapshot.ServicesJSON, &services); uErr != nil {
					return
				}
			}
			runtimeInfraSnapshots.mu.Lock()
			runtimeInfraSnapshots.snapshotAt = snapshot.SnapshotAt.UTC().Format(time.RFC3339)
			runtimeInfraSnapshots.nodes = nodes
			runtimeInfraSnapshots.services = services
			runtimeInfraSnapshots.degradedFrom = append([]string(nil), snapshot.DegradedProviders...)
			runtimeInfraSnapshots.mu.Unlock()
			return
		}
	}
	loader, ok := h.cfg.ApplicationStore.(infraSnapshotPersistence)
	if !ok {
		return
	}
	snapshotAt, nodesJSON, servicesJSON, degradedProviders, err := loader.LoadLatestInfraSnapshot(ctx)
	if err != nil {
		return
	}
	var nodes []infraNodeRecord
	var services []infraServiceRecord
	if len(nodesJSON) > 0 {
		if err := json.Unmarshal(nodesJSON, &nodes); err != nil {
			return
		}
	}
	if len(servicesJSON) > 0 {
		if err := json.Unmarshal(servicesJSON, &services); err != nil {
			return
		}
	}
	runtimeInfraSnapshots.mu.Lock()
	runtimeInfraSnapshots.snapshotAt = snapshotAt.UTC().Format(time.RFC3339)
	runtimeInfraSnapshots.nodes = nodes
	runtimeInfraSnapshots.services = services
	runtimeInfraSnapshots.degradedFrom = append([]string(nil), degradedProviders...)
	runtimeInfraSnapshots.mu.Unlock()
}

func (h *Handler) homeLabAdapter() (adapters.HomeLabAdapter, bool) {
	store, ok := h.cfg.ApplicationStore.(infraSnapshotPersistence)
	if !ok {
		return adapters.HomeLabAdapter{}, false
	}
	return adapters.HomeLabAdapter{
		Store:        store,
		HealthPolicy: homeLabHealthPolicyFromConfig(h.cfg.HomeLabProviderKey, h.cfg.HomeLabDegradedRaw),
	}, true
}

func homeLabHealthPolicyFromConfig(providerKey, degradedRaw string) adapters.HomeLabHealthPolicy {
	statuses := map[string]bool{}
	for _, item := range strings.Split(degradedRaw, ",") {
		status := strings.ToLower(strings.TrimSpace(item))
		if status == "" {
			continue
		}
		statuses[status] = true
	}
	if len(statuses) == 0 {
		statuses = map[string]bool{"warning": true, "degraded": true, "down": true}
	}
	key := strings.TrimSpace(providerKey)
	if key == "" {
		key = "homelab-agent"
	}
	return adapters.HomeLabHealthPolicy{
		DegradedStatuses: statuses,
		ProviderKey:      key,
	}
}

func toHomeLabRawSnapshot(req infraSnapshotRequest) (adapters.HomeLabRawSnapshot, error) {
	nodes := make([]json.RawMessage, 0, len(req.Nodes))
	for _, n := range req.Nodes {
		raw, err := json.Marshal(n)
		if err != nil {
			return adapters.HomeLabRawSnapshot{}, err
		}
		nodes = append(nodes, json.RawMessage(raw))
	}
	services := make([]adapters.HomeLabService, 0, len(req.Services))
	for _, s := range req.Services {
		services = append(services, adapters.HomeLabService{
			ServiceID:    s.ServiceID,
			HealthStatus: s.HealthStatus,
		})
	}
	return adapters.HomeLabRawSnapshot{
		AgentID:    req.AgentID,
		SnapshotAt: req.SnapshotAt,
		TraceID:    req.TraceID,
		Nodes:      nodes,
		Services:   services,
	}, nil
}

func currentInfraServices() []infraServiceRecord {
	runtimeInfraSnapshots.mu.RLock()
	services := append([]infraServiceRecord(nil), runtimeInfraSnapshots.services...)
	runtimeInfraSnapshots.mu.RUnlock()
	if len(services) > 0 {
		return services
	}

	// Fallback: build a lightweight inventory from existing snapshot provider nodes.
	provider := StaticSnapshotProvider{}
	nodes, _ := provider.InfraNodes(nil)
	out := make([]infraServiceRecord, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, infraServiceRecord{
			ServiceID:    n.ID,
			NodeID:       n.ID,
			Name:         n.Label,
			Version:      "",
			Port:         0,
			HealthStatus: normalizeHealthStatus(n.Status),
			ObservedAt:   n.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}
	return out
}

func currentInfraTopologyV2() ([]infraNodeRecord, []serviceEdgeResponse) {
	runtimeInfraSnapshots.mu.RLock()
	nodes := append([]infraNodeRecord(nil), runtimeInfraSnapshots.nodes...)
	runtimeInfraSnapshots.mu.RUnlock()
	if len(nodes) > 0 {
		return nodes, []serviceEdgeResponse{}
	}
	provider := StaticSnapshotProvider{}
	pNodes, _ := provider.InfraNodes(nil)
	pEdges, _ := provider.InfraEdges(nil)
	outNodes := make([]infraNodeRecord, 0, len(pNodes))
	for _, n := range pNodes {
		outNodes = append(outNodes, infraNodeRecord{
			NodeID:      n.ID,
			Hostname:    n.Label,
			Environment: n.Region,
			Status:      n.Status,
			ObservedAt:  n.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}
	return outNodes, pEdges
}

func normalizeHealthStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "stable", "healthy":
		return "healthy"
	case "warning", "degraded":
		return "degraded"
	case "down":
		return "down"
	default:
		return "degraded"
	}
}

func collectDegradedProviders(services []infraServiceRecord) []string {
	if len(services) == 0 {
		return nil
	}
	seen := map[string]bool{}
	out := make([]string, 0)
	for _, s := range services {
		hs := normalizeHealthStatus(s.HealthStatus)
		if hs != "degraded" && hs != "down" {
			continue
		}
		key := "homelab-agent"
		if !seen[key] {
			seen[key] = true
			out = append(out, key)
		}
	}
	return out
}
