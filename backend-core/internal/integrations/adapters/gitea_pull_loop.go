package adapters

import (
	"context"
	"errors"
	"sync"
	"time"
)

// PullCycleResult captures the outcome of a single pull cycle.
type PullCycleResult struct {
	CycleID             string
	StartedAt           time.Time
	CompletedAt         time.Time
	RepositoriesTotal   int
	RepositoriesSynced  int
	RepositoriesErrored int
	RepositoriesPartial int
	RepositoriesSkipped int
	OverallResult       string // success | error | partial
}

// RunGiteaPullLoop periodically executes GiteaPullAdapter.PullAndIngestSince for each repository.
// `interval` is the cycle interval (default 1h). `cycleTimeout` caps a single cycle's total time.
// `concurrency` limits per-repository parallelism within a single cycle (default 4).
// `backoffCap` is the maximum exponential-backoff window (default 24h).
// `failureAlertThreshold` is the consecutive-failure count that triggers a metric event (default 5).
func RunGiteaPullLoop(
	ctx context.Context,
	adapter *GiteaPullAdapter,
	repoLister func(ctx context.Context) ([]RepositoryTarget, error),
	interval time.Duration,
	cycleTimeout time.Duration,
	concurrency int,
	backoffCap time.Duration,
	failureAlertThreshold int,
	onCycle func(PullCycleResult),
	onError func(error),
) error {
	if interval <= 0 {
		interval = 1 * time.Hour
	}
	if cycleTimeout <= 0 {
		cycleTimeout = 30 * time.Minute
	}
	if concurrency <= 0 {
		concurrency = 4
	}
	if backoffCap <= 0 {
		backoffCap = 24 * time.Hour
	}
	if failureAlertThreshold <= 0 {
		failureAlertThreshold = 5
	}

	// Run an initial cycle immediately, then on the ticker.
	runGiteaPullCycle(ctx, adapter, repoLister, cycleTimeout, concurrency, backoffCap, failureAlertThreshold, onCycle, onError)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			runGiteaPullCycle(ctx, adapter, repoLister, cycleTimeout, concurrency, backoffCap, failureAlertThreshold, onCycle, onError)
		}
	}
}

func runGiteaPullCycle(
	ctx context.Context,
	adapter *GiteaPullAdapter,
	repoLister func(ctx context.Context) ([]RepositoryTarget, error),
	cycleTimeout time.Duration,
	concurrency int,
	backoffCap time.Duration,
	failureAlertThreshold int,
	onCycle func(PullCycleResult),
	onError func(error),
) {
	cycleID := time.Now().UTC().Format("20060102T150405.000000000")
	startedAt := time.Now().UTC()

	cycleCtx, cancel := context.WithTimeout(ctx, cycleTimeout)
	defer cancel()

	auditGiteaPullStarted(cycleID)

	repos, err := repoLister(cycleCtx)
	if err != nil {
		observeGiteaPull("error", time.Since(startedAt))
		auditGiteaPullError(cycleID, "repo_lister", err.Error())
		if onError != nil {
			onError(err)
		}
		return
	}

	result := PullCycleResult{
		CycleID:           cycleID,
		StartedAt:         startedAt,
		RepositoriesTotal: len(repos),
	}

	// Filter out repos currently in backoff.
	due := make([]RepositoryTarget, 0, len(repos))
	for _, r := range repos {
		backoff, err := adapter.Store.BackoffUntil(cycleCtx, r.ID)
		if err == nil && !backoff.IsZero() && time.Now().UTC().Before(backoff) {
			result.RepositoriesSkipped++
			continue
		}
		due = append(due, r)
	}

	observeGiteaPullRepositoriesTotal(len(due))

	// Per-repository goroutine + semaphore.
	sem := NewSemaphore(concurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, repo := range due {
		wg.Add(1)
		sem.Acquire()
		go func(target RepositoryTarget) {
			defer wg.Done()
			defer sem.Release()
			pullErr := adapter.PullAndIngestSince(cycleCtx, target)

			mu.Lock()
			defer mu.Unlock()

			if pullErr == nil {
				result.RepositoriesSynced++
				return
			}

			var pe *PullError
			if errors.As(pullErr, &pe) {
				switch pe.Class {
				case "partial":
					result.RepositoriesPartial++
					// record alert metric for consecutive failures
					cf, _ := adapter.Store.IncrementConsecutiveFailures(cycleCtx, target.ID)
					auditGiteaPullPartial(cycleID, target.ID, pe.Class, pe.Message, cf)
					if cf >= failureAlertThreshold {
						observeGiteaPullAlertTriggered(target.ID, cf)
					}
					// exponential backoff (cap)
					until := time.Now().UTC().Add(backoffDuration(cf, backoffCap))
					_ = adapter.Store.SetBackoff(cycleCtx, target.ID, until)
				default:
					result.RepositoriesErrored++
					cf, _ := adapter.Store.IncrementConsecutiveFailures(cycleCtx, target.ID)
					auditGiteaPullPartial(cycleID, target.ID, pe.Class, pe.Message, cf)
					if cf >= failureAlertThreshold {
						observeGiteaPullAlertTriggered(target.ID, cf)
					}
					until := time.Now().UTC().Add(backoffDuration(cf, backoffCap))
					_ = adapter.Store.SetBackoff(cycleCtx, target.ID, until)
				}
			} else {
				result.RepositoriesErrored++
				cf, _ := adapter.Store.IncrementConsecutiveFailures(cycleCtx, target.ID)
				auditGiteaPullPartial(cycleID, target.ID, "unknown", pullErr.Error(), cf)
				if cf >= failureAlertThreshold {
					observeGiteaPullAlertTriggered(target.ID, cf)
				}
				until := time.Now().UTC().Add(backoffDuration(cf, backoffCap))
				_ = adapter.Store.SetBackoff(cycleCtx, target.ID, until)
			}
		}(repo)
	}
	wg.Wait()

	result.CompletedAt = time.Now().UTC()
	overall := "success"
	if result.RepositoriesErrored > 0 && result.RepositoriesSynced == 0 && result.RepositoriesPartial == 0 {
		overall = "error"
	} else if result.RepositoriesErrored > 0 || result.RepositoriesPartial > 0 {
		overall = "partial"
	}
	result.OverallResult = overall

	observeGiteaPull(overall, time.Since(startedAt))
	if overall == "success" {
		observeGiteaPullLastSuccess(time.Now().UTC())
	}
	auditGiteaPullSuccess(cycleID, result.RepositoriesSynced, int(time.Since(startedAt).Milliseconds()))

	if onCycle != nil {
		onCycle(result)
	}
}

// backoffDuration computes exponential backoff: 2^failures minutes, capped at backoffCap.
// failures=1 -> 2m, failures=2 -> 4m, ..., failures=10 -> ~17h, failures=11+ -> backoffCap.
// Uses integer left-shift to avoid math.Pow float overflow for large failures.
func backoffDuration(failures int, backoffCap time.Duration) time.Duration {
	if failures <= 0 {
		return 0
	}
	// Cap at 30 to keep minute value safe for int64 (2^30 minutes = ~2000 years).
	if failures > 30 {
		return backoffCap
	}
	minutes := int64(1) << uint(failures)
	d := time.Duration(minutes) * time.Minute
	if d > backoffCap {
		return backoffCap
	}
	return d
}
