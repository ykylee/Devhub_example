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
	start := time.Now()
	if _, err := adapter.PullAndIngest(ctx); err != nil {
		observeHomeLabPull("error", time.Since(start))
		if onError != nil {
			onError(err)
		}
	} else {
		observeHomeLabPull("success", time.Since(start))
		observeHomeLabPullSuccess(time.Now())
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			start := time.Now()
			if _, err := adapter.PullAndIngest(ctx); err != nil {
				observeHomeLabPull("error", time.Since(start))
				if onError != nil {
					onError(err)
				}
			} else {
				observeHomeLabPull("success", time.Since(start))
				observeHomeLabPullSuccess(time.Now())
			}
		}
	}
}
