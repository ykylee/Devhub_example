package serviceaction

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/devhub/backend-core/internal/domain"
)

const SimulationMode = "simulation"

var ErrUnsupportedExecutorMode = errors.New("unsupported service action executor mode")
var ErrServiceNotAllowed = errors.New("service action target is not allowed")
var ErrActionNotAllowed = errors.New("service action type is not allowed")

type Executor struct {
	Mode            string
	AllowedServices map[string]bool
	AllowedActions  map[string]bool
}

func NewExecutor(mode, allowedServices, allowedActions string) (Executor, error) {
	mode = strings.TrimSpace(mode)
	if mode == "" {
		return Executor{}, ErrUnsupportedExecutorMode
	}
	if mode != SimulationMode {
		return Executor{}, fmt.Errorf("%w: %s", ErrUnsupportedExecutorMode, mode)
	}
	return Executor{
		Mode:            mode,
		AllowedServices: csvSet(allowedServices),
		AllowedActions:  csvSet(allowedActions),
	}, nil
}

func (e Executor) ExecuteServiceAction(_ context.Context, command domain.Command) (map[string]any, error) {
	if command.CommandType != "service_action" {
		return nil, fmt.Errorf("unsupported command type %q", command.CommandType)
	}
	if !e.allowed(e.AllowedServices, command.TargetID) {
		return nil, fmt.Errorf("%w: %s", ErrServiceNotAllowed, command.TargetID)
	}
	if !e.allowed(e.AllowedActions, command.ActionType) {
		return nil, fmt.Errorf("%w: %s", ErrActionNotAllowed, command.ActionType)
	}

	return map[string]any{
		"mode":        e.Mode,
		"simulated":   e.Mode == SimulationMode,
		"service_id":  command.TargetID,
		"action_type": command.ActionType,
		"message":     "Service action passed executor policy and was simulated without external side effects.",
		"executed_at": time.Now().UTC().Format(time.RFC3339Nano),
	}, nil
}

func (e Executor) allowed(allowlist map[string]bool, value string) bool {
	if len(allowlist) == 0 {
		return false
	}
	return allowlist[strings.TrimSpace(value)]
}

func csvSet(raw string) map[string]bool {
	out := map[string]bool{}
	for _, part := range strings.Split(raw, ",") {
		value := strings.TrimSpace(part)
		if value != "" {
			out[value] = true
		}
	}
	return out
}
