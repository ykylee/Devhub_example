package adapters

import (
	"context"
	"time"
)

// RunHomeLabPullLoop periodically executes PullAndIngest until ctx cancellation.
func RunHomeLabPullLoop(ctx context.Context, adapter HomeLabAdapter, interval time.Duration, onError func(error)) error {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	runHomeLabPullOnce(ctx, adapter, interval, onError)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			runHomeLabPullOnce(ctx, adapter, interval, onError)
		}
	}
}

func runHomeLabPullOnce(ctx context.Context, adapter HomeLabAdapter, timeout time.Duration, onError func(error)) {
	start := time.Now()
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if _, err := adapter.PullAndIngest(runCtx); err != nil {
		observeHomeLabPull("error", time.Since(start))
		if onError != nil {
			onError(err)
		}
		return
	}
	observeHomeLabPull("success", time.Since(start))
	observeHomeLabPullSuccess(time.Now())
}
