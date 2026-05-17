package adapters

import (
	"context"
	"encoding/json"
	"time"
)

// InfraSnapshotStore defines the persistence boundary for HomeLab snapshot ingest.
type InfraSnapshotStore interface {
	SaveInfraSnapshot(ctx context.Context, ingestID, agentID string, snapshotAt time.Time, traceID string, nodesJSON, servicesJSON []byte, degradedProviders []string) error
	LoadLatestInfraSnapshot(ctx context.Context) (snapshotAt time.Time, nodesJSON, servicesJSON []byte, degradedProviders []string, err error)
}

// HomeLabSnapshot is the normalized adapter payload before store marshalling.
type HomeLabSnapshot struct {
	IngestID          string
	AgentID           string
	SnapshotAt        time.Time
	TraceID           string
	NodesJSON         []byte
	ServicesJSON      []byte
	DegradedProviders []string
}

// HomeLabPuller defines pull-based collector contract for future reconciliation flows.
type HomeLabPuller interface {
	PullSnapshot(ctx context.Context) (HomeLabRawSnapshot, error)
}

// HomeLabRawSnapshot is the transport-level shape before adapter normalization.
type HomeLabRawSnapshot struct {
	AgentID    string            `json:"agent_id"`
	SnapshotAt string            `json:"snapshot_at"`
	TraceID    string            `json:"trace_id"`
	Nodes      []json.RawMessage `json:"nodes"`
	Services   []HomeLabService  `json:"services"`
}

// HomeLabService is the minimal service contract used by health policy normalization.
type HomeLabService struct {
	ServiceID    string `json:"service_id"`
	HealthStatus string `json:"health_status"`
}
