package adapters

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

type fakeInfraSnapshotStore struct {
	saved HomeLabSnapshot
	terr  error
	load  HomeLabSnapshot
	lerr  error
}

type fakeHomeLabPuller struct {
	raw HomeLabRawSnapshot
	err error
}

func (s *fakeInfraSnapshotStore) SaveInfraSnapshot(_ context.Context, ingestID, agentID string, snapshotAt time.Time, traceID string, nodesJSON, servicesJSON []byte, degradedProviders []string) error {
	if s.terr != nil {
		return s.terr
	}
	s.saved = HomeLabSnapshot{
		IngestID:          ingestID,
		AgentID:           agentID,
		SnapshotAt:        snapshotAt,
		TraceID:           traceID,
		NodesJSON:         append([]byte(nil), nodesJSON...),
		ServicesJSON:      append([]byte(nil), servicesJSON...),
		DegradedProviders: append([]string(nil), degradedProviders...),
	}
	return nil
}

func (s *fakeInfraSnapshotStore) LoadLatestInfraSnapshot(_ context.Context) (time.Time, []byte, []byte, []string, error) {
	if s.lerr != nil {
		return time.Time{}, nil, nil, nil, s.lerr
	}
	return s.load.SnapshotAt,
		append([]byte(nil), s.load.NodesJSON...),
		append([]byte(nil), s.load.ServicesJSON...),
		append([]string(nil), s.load.DegradedProviders...),
		nil
}

func (p fakeHomeLabPuller) PullSnapshot(_ context.Context) (HomeLabRawSnapshot, error) {
	if p.err != nil {
		return HomeLabRawSnapshot{}, p.err
	}
	return p.raw, nil
}

func TestHomeLabAdapterIngestSnapshot(t *testing.T) {
	store := &fakeInfraSnapshotStore{}
	adapter := HomeLabAdapter{Store: store}
	snapshot := HomeLabSnapshot{
		IngestID:     "ing_01",
		AgentID:      "homelab-agent-a",
		SnapshotAt:   time.Date(2026, 5, 16, 11, 0, 0, 0, time.UTC),
		TraceID:      "trc_01",
		NodesJSON:    []byte(`[{"node_id":"node-1"}]`),
		ServicesJSON: []byte(`[{"service_id":"svc-1"}]`),
	}
	if err := adapter.IngestSnapshot(context.Background(), snapshot); err != nil {
		t.Fatalf("ingest snapshot: %v", err)
	}
	if store.saved.AgentID != "homelab-agent-a" || store.saved.IngestID != "ing_01" {
		t.Fatalf("saved snapshot mismatch: %+v", store.saved)
	}
}

func TestHomeLabAdapterLoadLatestSnapshot(t *testing.T) {
	store := &fakeInfraSnapshotStore{load: HomeLabSnapshot{
		SnapshotAt:        time.Date(2026, 5, 16, 12, 0, 0, 0, time.UTC),
		NodesJSON:         []byte(`[]`),
		ServicesJSON:      []byte(`[]`),
		DegradedProviders: []string{"homelab-agent"},
	}}
	adapter := HomeLabAdapter{Store: store}
	got, err := adapter.LoadLatestSnapshot(context.Background())
	if err != nil {
		t.Fatalf("load latest snapshot: %v", err)
	}
	if got.SnapshotAt.IsZero() || len(got.DegradedProviders) != 1 {
		t.Fatalf("unexpected snapshot: %+v", got)
	}
}

func TestHomeLabAdapterIngestSnapshotRejectsInvalidInput(t *testing.T) {
	adapter := HomeLabAdapter{Store: &fakeInfraSnapshotStore{}}
	err := adapter.IngestSnapshot(context.Background(), HomeLabSnapshot{})
	if !errors.Is(err, ErrInvalidHomeLabSnapshot) {
		t.Fatalf("err=%v; want ErrInvalidHomeLabSnapshot", err)
	}
}

func TestHomeLabAdapterNormalizeSnapshot(t *testing.T) {
	adapter := HomeLabAdapter{}
	raw := HomeLabRawSnapshot{
		AgentID:    "homelab-agent-a",
		SnapshotAt: "2026-05-16T12:34:56Z",
		TraceID:    "trc_01",
		Nodes:      []json.RawMessage{json.RawMessage(`{"node_id":"node-1"}`)},
		Services: []HomeLabService{
			{ServiceID: "svc-1", HealthStatus: "healthy"},
			{ServiceID: "svc-2", HealthStatus: "down"},
		},
	}
	snapshot, err := adapter.NormalizeSnapshot(raw)
	if err != nil {
		t.Fatalf("normalize snapshot: %v", err)
	}
	if snapshot.IngestID == "" || snapshot.AgentID != "homelab-agent-a" {
		t.Fatalf("unexpected normalized snapshot: %+v", snapshot)
	}
	if len(snapshot.DegradedProviders) != 1 || snapshot.DegradedProviders[0] != "homelab-agent" {
		t.Fatalf("unexpected degraded providers: %+v", snapshot.DegradedProviders)
	}
}

func TestHomeLabAdapterPullAndIngest(t *testing.T) {
	store := &fakeInfraSnapshotStore{}
	puller := fakeHomeLabPuller{
		raw: HomeLabRawSnapshot{
			AgentID:    "homelab-agent-a",
			SnapshotAt: "2026-05-16T12:34:56Z",
			Services:   []HomeLabService{{ServiceID: "svc-1", HealthStatus: "warning"}},
		},
	}
	adapter := HomeLabAdapter{Store: store, Puller: puller}
	snapshot, err := adapter.PullAndIngest(context.Background())
	if err != nil {
		t.Fatalf("pull and ingest: %v", err)
	}
	if snapshot.IngestID == "" || store.saved.IngestID == "" {
		t.Fatalf("snapshot not ingested: snapshot=%+v saved=%+v", snapshot, store.saved)
	}
}

func TestHomeLabAdapterHealthPolicyOverride(t *testing.T) {
	adapter := HomeLabAdapter{
		HealthPolicy: HomeLabHealthPolicy{
			DegradedStatuses: map[string]bool{"critical": true},
			ProviderKey:      "homelab-custom",
		},
	}
	raw := HomeLabRawSnapshot{
		AgentID:    "homelab-agent-a",
		SnapshotAt: "2026-05-16T12:34:56Z",
		Services:   []HomeLabService{{ServiceID: "svc-1", HealthStatus: "critical"}},
	}
	snapshot, err := adapter.NormalizeSnapshot(raw)
	if err != nil {
		t.Fatalf("normalize with policy override: %v", err)
	}
	if len(snapshot.DegradedProviders) != 1 || snapshot.DegradedProviders[0] != "homelab-custom" {
		t.Fatalf("unexpected degraded providers with override: %+v", snapshot.DegradedProviders)
	}
}
