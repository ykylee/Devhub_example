package ci

import (
	"context"
	"time"
)

type Run struct {
	ID              string     `json:"id"`
	RepositoryName  string     `json:"repository_name"`
	Branch          string     `json:"branch"`
	CommitSHA       string     `json:"commit_sha"`
	Status          string     `json:"status"`
	DurationSeconds int        `json:"duration_seconds"`
	StartedAt       *time.Time `json:"started_at"`
	FinishedAt      *time.Time `json:"finished_at"`
	HTMLURL         string     `json:"html_url"`
}

type LogLine struct {
	Timestamp *time.Time `json:"timestamp"`
	Level     string     `json:"level"`
	Message   string     `json:"message"`
	StepName  string     `json:"step_name"`
}

type Adapter interface {
	GetRuns(ctx context.Context, owner, repo string) ([]Run, error)
	GetRunLogs(ctx context.Context, owner, repo, runID string) ([]LogLine, error)
	TriggerRun(ctx context.Context, owner, repo, branch string) error
}
