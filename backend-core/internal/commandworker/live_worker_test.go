package commandworker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/domain"
)

type fakeLiveStore struct {
	pending []domain.Command
	updates []domain.Command
}

func (s *fakeLiveStore) ListRunnableLiveServiceActionCommands(context.Context, int) ([]domain.Command, error) {
	out := []domain.Command{}
	for _, command := range s.pending {
		if command.CommandType == "service_action" && command.Status == "pending" && !command.DryRun && !command.RequiresApproval {
			out = append(out, command)
		}
	}
	return out, nil
}

func (s *fakeLiveStore) UpdateCommandStatus(_ context.Context, commandID, status string, resultPayload map[string]any) (domain.Command, error) {
	for i, command := range s.pending {
		if command.CommandID == commandID {
			command.Status = status
			command.ResultPayload = resultPayload
			command.UpdatedAt = time.Date(2026, 5, 7, 12, len(s.updates), 0, 0, time.UTC)
			s.pending[i] = command
			s.updates = append(s.updates, command)
			return command, nil
		}
	}
	return domain.Command{}, nil
}

type fakeServiceActionExecutor struct {
	commands []domain.Command
	result   map[string]any
	err      error
}

func (e *fakeServiceActionExecutor) ExecuteServiceAction(_ context.Context, command domain.Command) (map[string]any, error) {
	e.commands = append(e.commands, command)
	if e.err != nil {
		return nil, e.err
	}
	return e.result, nil
}

func TestLiveWorkerExecutesOnlyApprovedLiveServiceActions(t *testing.T) {
	store := &fakeLiveStore{
		pending: []domain.Command{
			{
				CommandID:        "cmd_approved",
				CommandType:      "service_action",
				TargetType:       "service",
				TargetID:         "runner-asia-01",
				ActionType:       "restart",
				Status:           "pending",
				DryRun:           false,
				RequiresApproval: false,
			},
			{
				CommandID:        "cmd_waiting",
				CommandType:      "service_action",
				Status:           "pending",
				DryRun:           false,
				RequiresApproval: true,
			},
			{
				CommandID:        "cmd_dry",
				CommandType:      "service_action",
				Status:           "pending",
				DryRun:           true,
				RequiresApproval: false,
			},
		},
	}
	executor := &fakeServiceActionExecutor{result: map[string]any{"message": "Restart accepted"}}
	publisher := &fakePublisher{}

	err := LiveWorker{Store: store, Executor: executor, Publisher: publisher}.ProcessOnce(context.Background())
	if err != nil {
		t.Fatalf("process live command: %v", err)
	}

	if len(executor.commands) != 1 || executor.commands[0].CommandID != "cmd_approved" {
		t.Fatalf("expected only approved live command execution, got %+v", executor.commands)
	}
	if len(store.updates) != 2 || store.updates[0].Status != "running" || store.updates[1].Status != "succeeded" {
		t.Fatalf("unexpected status updates: %+v", store.updates)
	}
	if store.updates[1].ResultPayload["executor"] != "service_action" {
		t.Fatalf("expected executor marker, got %+v", store.updates[1].ResultPayload)
	}
	if len(publisher.events) != 2 || publisher.events[1].Status != "succeeded" {
		t.Fatalf("unexpected published events: %+v", publisher.events)
	}
}

func TestLiveWorkerMarksFailureWhenExecutorFails(t *testing.T) {
	store := &fakeLiveStore{
		pending: []domain.Command{
			{
				CommandID:        "cmd_approved",
				CommandType:      "service_action",
				Status:           "pending",
				DryRun:           false,
				RequiresApproval: false,
			},
		},
	}
	executor := &fakeServiceActionExecutor{err: errors.New("executor unavailable")}
	publisher := &fakePublisher{}

	err := LiveWorker{Store: store, Executor: executor, Publisher: publisher}.ProcessOnce(context.Background())
	if err != nil {
		t.Fatalf("process live command: %v", err)
	}
	if len(store.updates) != 2 || store.updates[0].Status != "running" || store.updates[1].Status != "failed" {
		t.Fatalf("unexpected status updates: %+v", store.updates)
	}
	if len(publisher.events) != 2 || publisher.events[1].Status != "failed" {
		t.Fatalf("unexpected published events: %+v", publisher.events)
	}
}
