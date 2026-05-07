package commandworker

import (
	"context"
	"time"

	"github.com/devhub/backend-core/internal/domain"
)

const defaultBatchLimit = 25

type Store interface {
	ListRunnableDryRunCommands(context.Context, int) ([]domain.Command, error)
	UpdateCommandStatus(context.Context, string, string, map[string]any) (domain.Command, error)
}

type Publisher interface {
	PublishCommandStatus(domain.Command)
}

type Worker struct {
	Store     Store
	Publisher Publisher
	Limit     int
}

func (w Worker) ProcessOnce(ctx context.Context) error {
	limit := w.Limit
	if limit <= 0 {
		limit = defaultBatchLimit
	}

	commands, err := w.Store.ListRunnableDryRunCommands(ctx, limit)
	if err != nil {
		return err
	}

	for _, command := range commands {
		running, err := w.Store.UpdateCommandStatus(ctx, command.CommandID, "running", map[string]any{
			"executor": "dry_run",
			"message":  "Dry-run command is being validated without external side effects.",
		})
		if err != nil {
			return err
		}
		w.publish(running)

		succeeded, err := w.Store.UpdateCommandStatus(ctx, command.CommandID, "succeeded", map[string]any{
			"executor": "dry_run",
			"message":  "Dry-run command accepted without external side effects.",
		})
		if err != nil {
			return err
		}
		w.publish(succeeded)
	}

	return nil
}

func (w Worker) Run(ctx context.Context, interval time.Duration) error {
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

func (w Worker) publish(command domain.Command) {
	if w.Publisher != nil {
		w.Publisher.PublishCommandStatus(command)
	}
}
