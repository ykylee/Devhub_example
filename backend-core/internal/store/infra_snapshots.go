package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

func (s *PostgresStore) SaveInfraSnapshot(ctx context.Context, ingestID, agentID string, snapshotAt time.Time, traceID string, nodesJSON, servicesJSON []byte, degradedProviders []string) error {
	degradedJSON, err := json.Marshal(degradedProviders)
	if err != nil {
		return fmt.Errorf("marshal degraded providers: %w", err)
	}
	const query = `
INSERT INTO infra_service_snapshots (
	ingest_id,
	agent_id,
	snapshot_at,
	trace_id,
	nodes_payload,
	services_payload,
	degraded_providers
) VALUES ($1, $2, $3, NULLIF($4, ''), $5::jsonb, $6::jsonb, $7::jsonb)`
	_, err = s.pool.Exec(ctx, query, ingestID, agentID, snapshotAt.UTC(), traceID, string(nodesJSON), string(servicesJSON), string(degradedJSON))
	if err != nil {
		return fmt.Errorf("save infra snapshot: %w", err)
	}
	return nil
}

func (s *PostgresStore) LoadLatestInfraSnapshot(ctx context.Context) (snapshotAt time.Time, nodesJSON, servicesJSON []byte, degradedProviders []string, err error) {
	const query = `
SELECT snapshot_at, nodes_payload::text, services_payload::text, degraded_providers::text
FROM infra_service_snapshots
ORDER BY snapshot_at DESC, created_at DESC
LIMIT 1`
	var nodesRaw, servicesRaw, degradedRaw string
	if err := s.pool.QueryRow(ctx, query).Scan(&snapshotAt, &nodesRaw, &servicesRaw, &degradedRaw); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return time.Time{}, nil, nil, nil, ErrNotFound
		}
		return time.Time{}, nil, nil, nil, fmt.Errorf("load latest infra snapshot: %w", err)
	}
	nodesJSON = []byte(nodesRaw)
	servicesJSON = []byte(servicesRaw)
	if strings.TrimSpace(degradedRaw) != "" {
		if uErr := json.Unmarshal([]byte(degradedRaw), &degradedProviders); uErr != nil {
			return time.Time{}, nil, nil, nil, fmt.Errorf("decode degraded providers: %w", uErr)
		}
	}
	return snapshotAt, nodesJSON, servicesJSON, degradedProviders, nil
}
