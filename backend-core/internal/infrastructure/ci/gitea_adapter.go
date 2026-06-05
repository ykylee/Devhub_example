package ci

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/devhub/backend-core/internal/infrastructure/gitea"
)

type GiteaActionsAdapter struct {
	client *gitea.Client
}

func NewGiteaActionsAdapter(client *gitea.Client) *GiteaActionsAdapter {
	return &GiteaActionsAdapter{client: client}
}

func (a *GiteaActionsAdapter) GetRuns(ctx context.Context, owner, repo string) ([]Run, error) {
	runs, err := a.client.ListActionRuns(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("list action runs from gitea: %w", err)
	}

	result := make([]Run, 0, len(runs))
	for _, run := range runs {
		status := normalizeGiteaStatus(run.Status, run.Conclusion)
		
		var duration int
		if !run.UpdatedAt.IsZero() && !run.CreatedAt.IsZero() {
			duration = int(run.UpdatedAt.Sub(run.CreatedAt).Seconds())
		}

		createdAtCopy := run.CreatedAt
		updatedAtCopy := run.UpdatedAt

		result = append(result, Run{
			ID:              strconv.FormatInt(run.ID, 10),
			RepositoryName:  owner + "/" + repo,
			Branch:          run.HeadBranch,
			CommitSHA:       run.HeadSHA,
			Status:          status,
			DurationSeconds: duration,
			StartedAt:       &createdAtCopy,
			FinishedAt:      &updatedAtCopy,
			HTMLURL:         run.HTMLURL,
		})
	}
	return result, nil
}

func (a *GiteaActionsAdapter) GetRunLogs(ctx context.Context, owner, repo, runID string) ([]LogLine, error) {
	idVal, err := strconv.ParseInt(runID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse runID %q: %w", runID, err)
	}

	logText, err := a.client.GetActionRunLogs(ctx, owner, repo, idVal)
	if err != nil {
		return nil, fmt.Errorf("get action run logs from gitea: %w", err)
	}

	lines := strings.Split(logText, "\n")
	result := make([]LogLine, 0, len(lines))
	
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Parse simple console log format: "2026-06-05T13:10:00Z [info] step: message" or similar.
		var ts *time.Time
		level := "info"
		message := trimmed
		stepName := "run"

		// Basic whitespace separation
		parts := strings.SplitN(trimmed, " ", 2)
		if len(parts) > 0 {
			if t, err := time.Parse(time.RFC3339, parts[0]); err == nil {
				ts = &t
				if len(parts) > 1 {
					message = parts[1]
				}
			}
		}

		// Level deduction
		lowerMsg := strings.ToLower(message)
		if strings.Contains(lowerMsg, "[error]") || strings.Contains(lowerMsg, "error:") {
			level = "error"
		} else if strings.Contains(lowerMsg, "[warning]") || strings.Contains(lowerMsg, "warning:") {
			level = "warning"
		}

		result = append(result, LogLine{
			Timestamp: ts,
			Level:     level,
			Message:   message,
			StepName:  stepName,
		})
		
		// Limit to 500 lines for efficiency
		if i >= 500 {
			break
		}
	}

	return result, nil
}

func (a *GiteaActionsAdapter) TriggerRun(ctx context.Context, owner, repo, branch string) error {
	// Gitea Actions trigger manually (e.g. dispatch or webhook mock)
	// Currently returns nil as stub integration.
	return nil
}

func normalizeGiteaStatus(status, conclusion string) string {
	status = strings.ToLower(status)
	conclusion = strings.ToLower(conclusion)
	
	if status == "completed" && conclusion != "" {
		status = conclusion
	}

	switch status {
	case "queued", "running", "success", "failed", "cancelled", "skipped":
		return status
	case "failure", "error":
		return "failed"
	default:
		return "unknown"
	}
}
