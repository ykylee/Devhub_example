package serviceaction

import (
	"context"
	"errors"
	"testing"

	"github.com/devhub/backend-core/internal/domain"
)

func TestSimulationExecutorAllowsConfiguredServiceAndAction(t *testing.T) {
	executor, err := NewExecutor(SimulationMode, "runner-asia-01,backend-core", "restart,scale")
	if err != nil {
		t.Fatalf("new executor: %v", err)
	}

	result, err := executor.ExecuteServiceAction(context.Background(), domain.Command{
		CommandID:   "cmd_live",
		CommandType: "service_action",
		TargetID:    "runner-asia-01",
		ActionType:  "restart",
	})
	if err != nil {
		t.Fatalf("execute service action: %v", err)
	}
	if result["simulated"] != true || result["service_id"] != "runner-asia-01" || result["action_type"] != "restart" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestSimulationExecutorRejectsUnlistedService(t *testing.T) {
	executor, err := NewExecutor(SimulationMode, "runner-asia-01", "restart")
	if err != nil {
		t.Fatalf("new executor: %v", err)
	}

	_, err = executor.ExecuteServiceAction(context.Background(), domain.Command{
		CommandType: "service_action",
		TargetID:    "postgres",
		ActionType:  "restart",
	})
	if !errors.Is(err, ErrServiceNotAllowed) {
		t.Fatalf("expected ErrServiceNotAllowed, got %v", err)
	}
}

func TestSimulationExecutorRejectsUnlistedAction(t *testing.T) {
	executor, err := NewExecutor(SimulationMode, "runner-asia-01", "restart")
	if err != nil {
		t.Fatalf("new executor: %v", err)
	}

	_, err = executor.ExecuteServiceAction(context.Background(), domain.Command{
		CommandType: "service_action",
		TargetID:    "runner-asia-01",
		ActionType:  "rollback",
	})
	if !errors.Is(err, ErrActionNotAllowed) {
		t.Fatalf("expected ErrActionNotAllowed, got %v", err)
	}
}

func TestNewExecutorRejectsUnsupportedMode(t *testing.T) {
	_, err := NewExecutor("shell", "runner-asia-01", "restart")
	if !errors.Is(err, ErrUnsupportedExecutorMode) {
		t.Fatalf("expected ErrUnsupportedExecutorMode, got %v", err)
	}
}
