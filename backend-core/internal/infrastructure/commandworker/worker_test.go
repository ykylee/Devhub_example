package commandworker

import (
	"context"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/domain"
)

type fakeStore struct {
	pending []domain.Command
	updates []domain.Command
}

func (s *fakeStore) ListRunnableDryRunCommands(context.Context, int) ([]domain.Command, error) {
	return s.pending, nil
}

func (s *fakeStore) UpdateCommandStatus(_ context.Context, commandID, status string, resultPayload map[string]any) (domain.Command, error) {
	for i, command := range s.pending {
		if command.CommandID == commandID {
			command.Status = status
			command.ResultPayload = resultPayload
			command.UpdatedAt = time.Date(2026, 5, 6, 11, len(s.updates), 0, 0, time.UTC)
			s.pending[i] = command
			s.updates = append(s.updates, command)
			return command, nil
		}
	}
	return domain.Command{}, nil
}

type fakePublisher struct {
	events []domain.Command
}

func (p *fakePublisher) PublishCommandStatus(command domain.Command) {
	p.events = append(p.events, command)
}

func TestWorkerTransitionsDryRunCommandsAndPublishes(t *testing.T) {
	store := &fakeStore{
		pending: []domain.Command{
			{
				CommandID:        "cmd_test",
				CommandType:      "service_action",
				TargetType:       "service",
				TargetID:         "backend-core",
				ActionType:       "restart",
				Status:           "pending",
				DryRun:           true,
				RequiresApproval: false,
			},
		},
	}
	publisher := &fakePublisher{}

	err := Worker{Store: store, Publisher: publisher}.ProcessOnce(context.Background())
	if err != nil {
		t.Fatalf("process commands: %v", err)
	}

	if len(store.updates) != 2 {
		t.Fatalf("expected running and succeeded updates, got %+v", store.updates)
	}
	if store.updates[0].Status != "running" || store.updates[1].Status != "succeeded" {
		t.Fatalf("unexpected status updates: %+v", store.updates)
	}
	if len(publisher.events) != 2 || publisher.events[1].Status != "succeeded" {
		t.Fatalf("unexpected published events: %+v", publisher.events)
	}
	if publisher.events[1].ResultPayload["executor"] != "dry_run" {
		t.Fatalf("expected dry-run result payload, got %+v", publisher.events[1].ResultPayload)
	}
}
