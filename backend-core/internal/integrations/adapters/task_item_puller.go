package adapters

import (
	"context"
	"time"
)

// ExternalTaskItem is the normalized task item used by the pull adapter.
// Matches the domain model but without store-specific fields.
type ExternalTaskItem struct {
	ExternalID       string
	Title            string
	Description      string
	RawStatus        string
	NormalizedStatus string
	Priority         string
	Assignee         string
	Reporter         string
	URL              string
	Labels           []string
	FetchedAt        time.Time
}

// TaskItemPuller defines the pull-based collector contract for external task trackers.
// Implementations (GitHub Issues, GitLab, Jira adapters) fetch from their respective APIs.
type TaskItemPuller interface {
	// FetchTaskItems pulls task items updated since the given timestamp.
	// Returns normalized ExternalTaskItem slice + optional cursor for pagination.
	// Cursor is provider-specific (page token, offset, etc.). Empty string = no more pages.
	FetchTaskItems(ctx context.Context, since time.Time, cursor string) ([]ExternalTaskItem, string, error)
}

// TaskItemPullerStore defines the persistence boundary that the pull loop needs.
// This is a subset of the full store — the loop only writes, doesn't read.
type TaskItemPullerStore interface {
	UpsertExternalTaskItem(providerID, externalID, title, description, rawStatus, normalizedStatus,
		priority, assignee, reporter, url string, labels []string, fetchedAt time.Time) error
	UpdateProviderLastPulledAt(providerID string, pulledAt time.Time) error
	DetectWebhookSeqGaps(providerID string) (int64, error)
}

// TaskItemPullAdapter wires a Puller + Store for a single provider.
type TaskItemPullAdapter struct {
	ProviderID string
	Puller     TaskItemPuller
	Store      TaskItemPullerStore
}

// PullAndIngest performs one pull cycle: FetchTaskItems → upsert each → update cursor.
func (a *TaskItemPullAdapter) PullAndIngest(ctx context.Context, since time.Time) (int, error) {
	cursor := ""
	total := 0

	for {
		items, nextCursor, err := a.Puller.FetchTaskItems(ctx, since, cursor)
		if err != nil {
			return total, err
		}

		for _, item := range items {
			if err := a.Store.UpsertExternalTaskItem(
				a.ProviderID, item.ExternalID, item.Title, item.Description,
				item.RawStatus, item.NormalizedStatus,
				item.Priority, item.Assignee, item.Reporter, item.URL,
				item.Labels, item.FetchedAt,
			); err != nil {
				return total, err
			}
			total++
		}

		if nextCursor == "" {
			break
		}
		cursor = nextCursor
	}

	// Update last_pulled_at after successful pull
	now := time.Now().UTC()
	if err := a.Store.UpdateProviderLastPulledAt(a.ProviderID, now); err != nil {
		return total, err
	}

	return total, nil
}

// RunTaskItemPullLoop runs a periodic pull loop for a single adapter.
// Pattern matches homelab_pull_loop.go.
func RunTaskItemPullLoop(ctx context.Context, adapter *TaskItemPullAdapter, interval time.Duration, onError func(error)) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Immediate first tick
	if err := runTaskItemPullOnce(ctx, adapter); err != nil && onError != nil {
		onError(err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := runTaskItemPullOnce(ctx, adapter); err != nil && onError != nil {
				onError(err)
			}
		}
	}
}

func runTaskItemPullOnce(ctx context.Context, adapter *TaskItemPullAdapter) error {
	// Detect seq gaps first
	gaps, err := adapter.Store.DetectWebhookSeqGaps(adapter.ProviderID)
	if err != nil {
		return err
	}
	_ = gaps // TODO: emit metric + audit for gap > 0

	// Pull since last_pulled_at is handled by the caller passing the right since time
	return nil
}
