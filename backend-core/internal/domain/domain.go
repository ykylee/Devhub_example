package domain

import (
	"context"
	"time"
)

type Repository struct {
	GiteaID       int64
	FullName      string
	OwnerLogin    string
	Name          string
	CloneURL      string
	HTMLURL       string
	DefaultBranch string
	Private       bool
}

type User struct {
	GiteaID     int64
	Login       string
	DisplayName string
	AvatarURL   string
	HTMLURL     string
}

type Issue struct {
	GiteaID           int64
	RepositoryGiteaID int64
	RepositoryName    string
	Number            int64
	Title             string
	State             string
	AuthorLogin       string
	AssigneeLogin     string
	HTMLURL           string
	OpenedAt          *time.Time
	ClosedAt          *time.Time
}

type PullRequest struct {
	GiteaID           int64
	RepositoryGiteaID int64
	RepositoryName    string
	Number            int64
	Title             string
	State             string
	AuthorLogin       string
	HeadBranch        string
	BaseBranch        string
	HeadSHA           string
	HTMLURL           string
	MergedAt          *time.Time
	ClosedAt          *time.Time
}

type CIRun struct {
	ExternalID      string
	RepositoryName  string
	Branch          string
	CommitSHA       string
	Status          string
	Conclusion      string
	StartedAt       *time.Time
	FinishedAt      *time.Time
	DurationSeconds *int
	HTMLURL         string
}

type ChangeSet struct {
	Repository  *Repository
	Sender      *User
	Issue       *Issue
	PullRequest *PullRequest
	CIRun       *CIRun
	Ignored     bool
	Reason      string
}

type Sink interface {
	UpsertRepository(context.Context, Repository) error
	UpsertUser(context.Context, User) error
	UpsertIssue(context.Context, Issue) error
	UpsertPullRequest(context.Context, PullRequest) error
	UpsertCIRun(context.Context, CIRun) error
	MarkWebhookEventProcessed(context.Context, int64) error
	MarkWebhookEventIgnored(context.Context, int64, string) error
	MarkWebhookEventFailed(context.Context, int64, string) error
}
