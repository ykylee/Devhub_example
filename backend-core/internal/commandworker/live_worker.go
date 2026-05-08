package commandworker

import (
	"context"
	"time"

	"github.com/devhub/backend-core/internal/domain"
)

type LiveStore interface {
	ListRunnableLiveServiceActionCommands(context.Context, int) ([]domain.Command, error)
	UpdateCommandStatus(context.Context, string, string, map[string]any) (domain.Command, error)
}

type ServiceActionExecutor interface {
	ExecuteServiceAction(context.Context, domain.Command) (map[string]any, error)
}

type LiveWorker struct {
	Store     LiveStore
	Executor  ServiceActionExecutor
	Publisher Publisher
	Limit     int
}

func (w LiveWorker) ProcessOnce(ctx context.Context) error {
	if w.Store == nil || w.Executor == nil {
		return nil
	}
	limit := w.Limit
	if limit <= 0 {
		limit = defaultBatchLimit
	}

	commands, err := w.Store.ListRunnableLiveServiceActionCommands(ctx, limit)
	if err != nil {
		return err
	}

	for _, command := range commands {
		running, err := w.Store.UpdateCommandStatus(ctx, command.CommandID, "running", map[string]any{
			"executor": "service_action",
			"message":  "Service action execution started.",
		})
		if err != nil {
			return err
		}
		w.publish(running)

		result, err := w.Executor.ExecuteServiceAction(ctx, command)
		if err != nil {
			failed, updateErr := w.Store.UpdateCommandStatus(ctx, command.CommandID, "failed", map[string]any{
				"executor": "service_action",
				"error":    err.Error(),
			})
			if updateErr != nil {
				return updateErr
			}
			w.publish(failed)
			continue
		}
		if result == nil {
			result = map[string]any{}
		}
		result["executor"] = "service_action"
		succeeded, err := w.Store.UpdateCommandStatus(ctx, command.CommandID, "succeeded", result)
		if err != nil {
			return err
		}
		w.publish(succeeded)
	}

	return nil
}

func (w LiveWorker) Run(ctx context.Context, interval time.Duration) error {
	if interval <= 0 {
		interval = time.Second
	}

	if err := w.ProcessOnce(ctx); err != nil {
		return err
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := w.ProcessOnce(ctx); err != nil {
				return err
			}
		}
	}
}

func (w LiveWorker) publish(command domain.Command) {
	if w.Publisher != nil {
		w.Publisher.PublishCommandStatus(command)
	}
}
