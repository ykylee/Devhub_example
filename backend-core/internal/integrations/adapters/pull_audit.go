package adapters

import (
	"context"
	"log"
	"time"
)

// PullAuditHook is the optional contract for emitting structured audit events from the pull loop.
// Implementations may write to the audit_logs table, ship to a SIEM, or simply log.
// The default implementation logs structured key=value lines; production wiring will replace this
// with the audit-ops emitter (sprint -h follow-up).
type PullAuditHook interface {
	PullStarted(ctx context.Context, cycleID string, repositoryCount int)
	PullSuccess(ctx context.Context, cycleID string, repositoriesSynced int, durationMs int)
	PullError(ctx context.Context, cycleID string, errorClass, errorMessage string)
	PullPartial(ctx context.Context, cycleID, repositoryID, errorClass, errorMessage string, consecutiveFailures int)
}

// LogPullAuditHook is the default log-only audit hook (no DB, no SIEM). Used when no hook is wired.
type LogPullAuditHook struct{}

func (LogPullAuditHook) PullStarted(_ context.Context, cycleID string, repositoryCount int) {
	log.Printf("audit pull_started cycle_id=%s repository_count=%d ts=%s", cycleID, repositoryCount, time.Now().UTC().Format(time.RFC3339Nano))
}

func (LogPullAuditHook) PullSuccess(_ context.Context, cycleID string, repositoriesSynced int, durationMs int) {
	log.Printf("audit pull_success cycle_id=%s repositories_synced=%d duration_ms=%d ts=%s", cycleID, repositoriesSynced, durationMs, time.Now().UTC().Format(time.RFC3339Nano))
}

func (LogPullAuditHook) PullError(_ context.Context, cycleID string, errorClass, errorMessage string) {
	log.Printf("audit pull_error cycle_id=%s error_class=%s error_message=%q ts=%s", cycleID, errorClass, errorMessage, time.Now().UTC().Format(time.RFC3339Nano))
}

func (LogPullAuditHook) PullPartial(_ context.Context, cycleID, repositoryID, errorClass, errorMessage string, consecutiveFailures int) {
	log.Printf("audit pull_partial cycle_id=%s repository_id=%s error_class=%s error_message=%q consecutive_failures=%d ts=%s", cycleID, repositoryID, errorClass, errorMessage, consecutiveFailures, time.Now().UTC().Format(time.RFC3339Nano))
}

// Package-level audit hook (overridable). Defaults to log-only.
var globalPullAuditHook PullAuditHook = LogPullAuditHook{}

// SetPullAuditHook replaces the default audit hook. Call from main.go wire.
func SetPullAuditHook(h PullAuditHook) {
	if h != nil {
		globalPullAuditHook = h
	}
}

func auditGiteaPullStarted(cycleID string) {
	if globalPullAuditHook == nil {
		return
	}
	// cycle_start is not in the hook signature; emit via PullStarted with repository_count=0 placeholder.
	// We use a synthetic repositoryCount to convey "cycle started" state when no lister result is known.
	globalPullAuditHook.PullStarted(context.Background(), cycleID, 0)
}

func auditGiteaPullSuccess(cycleID string, repositoriesSynced int, durationMs int) {
	if globalPullAuditHook == nil {
		return
	}
	globalPullAuditHook.PullSuccess(context.Background(), cycleID, repositoriesSynced, durationMs)
}

func auditGiteaPullError(cycleID string, errorClass, errorMessage string) {
	if globalPullAuditHook == nil {
		return
	}
	globalPullAuditHook.PullError(context.Background(), cycleID, errorClass, errorMessage)
}

func auditGiteaPullPartial(cycleID, repositoryID, errorClass, errorMessage string, consecutiveFailures int) {
	if globalPullAuditHook == nil {
		return
	}
	globalPullAuditHook.PullPartial(context.Background(), cycleID, repositoryID, errorClass, errorMessage, consecutiveFailures)
}
