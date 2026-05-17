package adapters

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var ErrInvalidHomeLabSnapshot = errors.New("invalid homelab snapshot")

// HomeLabAdapter provides a minimal persistence-oriented adapter skeleton.
type HomeLabAdapter struct {
	Store        InfraSnapshotStore
	Puller       HomeLabPuller
	HealthPolicy HomeLabHealthPolicy
}

type HomeLabHealthPolicy struct {
	DegradedStatuses map[string]bool
	ProviderKey      string
}

func defaultHomeLabHealthPolicy() HomeLabHealthPolicy {
	return HomeLabHealthPolicy{
		DegradedStatuses: map[string]bool{
			"warning":  true,
			"degraded": true,
			"down":     true,
		},
		ProviderKey: "homelab-agent",
	}
}

func (a HomeLabAdapter) IngestSnapshot(ctx context.Context, snapshot HomeLabSnapshot) error {
	if a.Store == nil {
		return ErrInvalidHomeLabSnapshot
	}
	if strings.TrimSpace(snapshot.IngestID) == "" || strings.TrimSpace(snapshot.AgentID) == "" {
		return ErrInvalidHomeLabSnapshot
	}
	if snapshot.SnapshotAt.IsZero() {
		return ErrInvalidHomeLabSnapshot
	}
	if len(snapshot.NodesJSON) == 0 && len(snapshot.ServicesJSON) == 0 {
		return ErrInvalidHomeLabSnapshot
	}
	err := a.Store.SaveInfraSnapshot(
		ctx,
		snapshot.IngestID,
		snapshot.AgentID,
		snapshot.SnapshotAt.UTC(),
		snapshot.TraceID,
		snapshot.NodesJSON,
		snapshot.ServicesJSON,
		snapshot.DegradedProviders,
	)
	if err == nil {
		observeHomeLabSnapshot(snapshot)
	}
	return err
}

func (a HomeLabAdapter) LoadLatestSnapshot(ctx context.Context) (HomeLabSnapshot, error) {
	if a.Store == nil {
		return HomeLabSnapshot{}, ErrInvalidHomeLabSnapshot
	}
	snapshotAt, nodesJSON, servicesJSON, degradedProviders, err := a.Store.LoadLatestInfraSnapshot(ctx)
	if err != nil {
		return HomeLabSnapshot{}, err
	}
	return HomeLabSnapshot{
		SnapshotAt:        snapshotAt.UTC(),
		NodesJSON:         nodesJSON,
		ServicesJSON:      servicesJSON,
		DegradedProviders: degradedProviders,
	}, nil
}

func (a HomeLabAdapter) PullAndIngest(ctx context.Context) (HomeLabSnapshot, error) {
	if a.Puller == nil {
		return HomeLabSnapshot{}, ErrInvalidHomeLabSnapshot
	}
	raw, err := a.Puller.PullSnapshot(ctx)
	if err != nil {
		return HomeLabSnapshot{}, err
	}
	snapshot, err := a.NormalizeSnapshot(raw)
	if err != nil {
		return HomeLabSnapshot{}, err
	}
	if err := a.IngestSnapshot(ctx, snapshot); err != nil {
		return HomeLabSnapshot{}, err
	}
	return snapshot, nil
}

func (a HomeLabAdapter) NormalizeSnapshot(raw HomeLabRawSnapshot) (HomeLabSnapshot, error) {
	agentID := strings.TrimSpace(raw.AgentID)
	if agentID == "" {
		return HomeLabSnapshot{}, ErrInvalidHomeLabSnapshot
	}
	snapshotAt, err := time.Parse(time.RFC3339, strings.TrimSpace(raw.SnapshotAt))
	if err != nil {
		return HomeLabSnapshot{}, ErrInvalidHomeLabSnapshot
	}
	nodesJSON, err := json.Marshal(raw.Nodes)
	if err != nil {
		return HomeLabSnapshot{}, ErrInvalidHomeLabSnapshot
	}
	servicesJSON, err := json.Marshal(raw.Services)
	if err != nil {
		return HomeLabSnapshot{}, ErrInvalidHomeLabSnapshot
	}
	ingestID := "ing_" + strings.ReplaceAll(snapshotAt.UTC().Format("20060102T150405.000000000"), ".", "")
	return HomeLabSnapshot{
		IngestID:          ingestID,
		AgentID:           agentID,
		SnapshotAt:        snapshotAt.UTC(),
		TraceID:           strings.TrimSpace(raw.TraceID),
		NodesJSON:         nodesJSON,
		ServicesJSON:      servicesJSON,
		DegradedProviders: a.collectDegradedProviders(raw.Services),
	}, nil
}

func (a HomeLabAdapter) collectDegradedProviders(services []HomeLabService) []string {
	if len(services) == 0 {
		return nil
	}
	policy := a.HealthPolicy
	if len(policy.DegradedStatuses) == 0 {
		policy = defaultHomeLabHealthPolicy()
	}
	if strings.TrimSpace(policy.ProviderKey) == "" {
		policy.ProviderKey = defaultHomeLabHealthPolicy().ProviderKey
	}
	seen := false
	for _, svc := range services {
		status := strings.ToLower(strings.TrimSpace(svc.HealthStatus))
		if policy.DegradedStatuses[status] {
			seen = true
			break
		}
	}
	if !seen {
		return nil
	}
	return []string{policy.ProviderKey}
}
