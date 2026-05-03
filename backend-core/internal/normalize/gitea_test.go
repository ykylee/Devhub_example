package normalize

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
)

func TestNormalizeIssueEvent(t *testing.T) {
	event := fixtureEvent(t, "issues", "issue_opened.json")

	changeSet, err := Normalize(event)
	if err != nil {
		t.Fatalf("normalize issue event: %v", err)
	}
	if changeSet.Repository == nil || changeSet.Repository.FullName != "acme/api" {
		t.Fatalf("unexpected repository: %+v", changeSet.Repository)
	}
	if changeSet.Sender == nil || changeSet.Sender.Login != "yklee" {
		t.Fatalf("unexpected sender: %+v", changeSet.Sender)
	}
	if changeSet.Issue == nil {
		t.Fatal("expected issue change")
	}
	if changeSet.Issue.Number != 17 || changeSet.Issue.State != "open" || changeSet.Issue.AuthorLogin != "yklee" {
		t.Fatalf("unexpected issue: %+v", changeSet.Issue)
	}
}

func TestNormalizePullRequestEvent(t *testing.T) {
	event := fixtureEvent(t, "pull_request", "pull_request_opened.json")

	changeSet, err := Normalize(event)
	if err != nil {
		t.Fatalf("normalize pull request event: %v", err)
	}
	if changeSet.PullRequest == nil {
		t.Fatal("expected pull request change")
	}
	if changeSet.PullRequest.Number != 23 || changeSet.PullRequest.HeadBranch != "feature/domain-normalizer" {
		t.Fatalf("unexpected pull request: %+v", changeSet.PullRequest)
	}
	if changeSet.PullRequest.BaseBranch != "main" || changeSet.PullRequest.State != "open" {
		t.Fatalf("unexpected pull request branch/state: %+v", changeSet.PullRequest)
	}
}

func TestNormalizeActionRunEvent(t *testing.T) {
	event := fixtureEvent(t, "action_run", "action_run_completed.json")

	changeSet, err := Normalize(event)
	if err != nil {
		t.Fatalf("normalize action run event: %v", err)
	}
	if changeSet.CIRun == nil {
		t.Fatal("expected ci run change")
	}
	if changeSet.CIRun.ExternalID != "501" || changeSet.CIRun.Status != "success" {
		t.Fatalf("unexpected ci run: %+v", changeSet.CIRun)
	}
	if changeSet.CIRun.RepositoryName != "acme/api" || changeSet.CIRun.Branch != "main" {
		t.Fatalf("unexpected ci run repository/branch: %+v", changeSet.CIRun)
	}
}

func TestNormalizeUnsupportedEventIsIgnored(t *testing.T) {
	event := fixtureEvent(t, "release", "issue_opened.json")

	changeSet, err := Normalize(event)
	if err != nil {
		t.Fatalf("normalize unsupported event: %v", err)
	}
	if !changeSet.Ignored || changeSet.Reason == "" {
		t.Fatalf("expected ignored change set, got %+v", changeSet)
	}
}

func TestProcessorMarksProcessed(t *testing.T) {
	event := fixtureEvent(t, "pull_request", "pull_request_opened.json")
	event.ID = 7
	sink := &memorySink{}
	processor := Processor{Sink: sink}

	if err := processor.Process(context.Background(), event); err != nil {
		t.Fatalf("process event: %v", err)
	}
	if len(sink.repositories) != 1 || len(sink.users) != 1 || len(sink.pullRequests) != 1 {
		t.Fatalf("unexpected sink writes: %+v", sink)
	}
	if sink.processedID != 7 {
		t.Fatalf("expected processed id 7, got %d", sink.processedID)
	}
}

func fixtureEvent(t *testing.T, eventType, name string) store.WebhookEvent {
	t.Helper()
	payload, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	now := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)
	return store.WebhookEvent{
		EventType:   eventType,
		DedupeKey:   eventType + "-" + name,
		Payload:     payload,
		Status:      "validated",
		ReceivedAt:  now,
		ValidatedAt: &now,
	}
}

type memorySink struct {
	repositories []domain.Repository
	users        []domain.User
	issues       []domain.Issue
	pullRequests []domain.PullRequest
	ciRuns       []domain.CIRun
	processedID  int64
	ignoredID    int64
	failedID     int64
}

func (s *memorySink) UpsertRepository(_ context.Context, value domain.Repository) error {
	s.repositories = append(s.repositories, value)
	return nil
}

func (s *memorySink) UpsertUser(_ context.Context, value domain.User) error {
	s.users = append(s.users, value)
	return nil
}

func (s *memorySink) UpsertIssue(_ context.Context, value domain.Issue) error {
	s.issues = append(s.issues, value)
	return nil
}

func (s *memorySink) UpsertPullRequest(_ context.Context, value domain.PullRequest) error {
	s.pullRequests = append(s.pullRequests, value)
	return nil
}

func (s *memorySink) UpsertCIRun(_ context.Context, value domain.CIRun) error {
	s.ciRuns = append(s.ciRuns, value)
	return nil
}

func (s *memorySink) MarkWebhookEventProcessed(_ context.Context, id int64) error {
	s.processedID = id
	return nil
}

func (s *memorySink) MarkWebhookEventIgnored(_ context.Context, id int64, _ string) error {
	s.ignoredID = id
	return nil
}

func (s *memorySink) MarkWebhookEventFailed(_ context.Context, id int64, _ string) error {
	s.failedID = id
	return nil
}
