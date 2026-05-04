package normalize

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
)

var ErrIgnoredEvent = errors.New("ignored webhook event")

type Processor struct {
	Sink domain.Sink
}

func (p Processor) Process(ctx context.Context, event store.WebhookEvent) error {
	changeSet, err := Normalize(event)
	if err != nil {
		if p.Sink != nil && event.ID != 0 {
			_ = p.Sink.MarkWebhookEventFailed(ctx, event.ID, err.Error())
		}
		return err
	}
	if p.Sink == nil {
		return nil
	}
	if changeSet.Ignored {
		if event.ID != 0 {
			return p.Sink.MarkWebhookEventIgnored(ctx, event.ID, changeSet.Reason)
		}
		return nil
	}
	if changeSet.Repository != nil {
		if err := p.Sink.UpsertRepository(ctx, *changeSet.Repository); err != nil {
			return p.fail(ctx, event.ID, err)
		}
	}
	if changeSet.Sender != nil {
		if err := p.Sink.UpsertUser(ctx, *changeSet.Sender); err != nil {
			return p.fail(ctx, event.ID, err)
		}
	}
	if changeSet.Issue != nil {
		if err := p.Sink.UpsertIssue(ctx, *changeSet.Issue); err != nil {
			return p.fail(ctx, event.ID, err)
		}
	}
	if changeSet.PullRequest != nil {
		if err := p.Sink.UpsertPullRequest(ctx, *changeSet.PullRequest); err != nil {
			return p.fail(ctx, event.ID, err)
		}
	}
	if changeSet.CIRun != nil {
		if err := p.Sink.UpsertCIRun(ctx, *changeSet.CIRun); err != nil {
			return p.fail(ctx, event.ID, err)
		}
	}
	if changeSet.Risk != nil {
		if err := p.Sink.UpsertRisk(ctx, *changeSet.Risk); err != nil {
			return p.fail(ctx, event.ID, err)
		}
	}
	if event.ID != 0 {
		return p.Sink.MarkWebhookEventProcessed(ctx, event.ID)
	}
	return nil
}

func (p Processor) fail(ctx context.Context, eventID int64, err error) error {
	if eventID != 0 {
		_ = p.Sink.MarkWebhookEventFailed(ctx, eventID, err.Error())
	}
	return err
}

func Normalize(event store.WebhookEvent) (domain.ChangeSet, error) {
	var payload webhookPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return domain.ChangeSet{}, fmt.Errorf("decode webhook payload: %w", err)
	}

	changeSet := domain.ChangeSet{
		Repository: payload.repository(),
		Sender:     payload.sender(),
	}

	switch event.EventType {
	case "issues", "issue":
		issue := payload.issue()
		if issue == nil {
			return domain.ChangeSet{}, errors.New("issue event missing issue payload")
		}
		if changeSet.Repository != nil {
			issue.RepositoryGiteaID = changeSet.Repository.GiteaID
			issue.RepositoryName = changeSet.Repository.FullName
		}
		changeSet.Issue = issue
	case "pull_request":
		pullRequest := payload.pullRequest()
		if pullRequest == nil {
			return domain.ChangeSet{}, errors.New("pull_request event missing pull_request payload")
		}
		if changeSet.Repository != nil {
			pullRequest.RepositoryGiteaID = changeSet.Repository.GiteaID
			pullRequest.RepositoryName = changeSet.Repository.FullName
		}
		changeSet.PullRequest = pullRequest
	case "action_run", "workflow_run":
		ciRun := payload.ciRun(changeSet.Repository)
		if ciRun == nil {
			return domain.ChangeSet{}, errors.New("action run event missing run payload")
		}
		changeSet.CIRun = ciRun
		changeSet.Risk = riskFromCIRun(ciRun)
	case "push":
		changeSet.Reason = "push events only refresh repository and sender metadata in normalization phase 1"
	default:
		changeSet.Ignored = true
		changeSet.Reason = "unsupported event type: " + event.EventType
	}

	if changeSet.Repository == nil && !changeSet.Ignored {
		return domain.ChangeSet{}, errors.New("event missing repository payload")
	}
	return changeSet, nil
}

func riskFromCIRun(run *domain.CIRun) *domain.Risk {
	if run == nil || run.Status != "failed" {
		return nil
	}
	sourceID := run.ExternalID
	repositoryName := run.RepositoryName
	title := "CI run failed"
	if repositoryName != "" {
		title = "CI run failed for " + repositoryName
	}
	reason := "CI run " + sourceID + " failed"
	if run.Branch != "" {
		reason += " on branch " + run.Branch
	}
	detectedAt := time.Now().UTC()
	if run.FinishedAt != nil {
		detectedAt = *run.FinishedAt
	} else if run.StartedAt != nil {
		detectedAt = *run.StartedAt
	}
	return &domain.Risk{
		RiskKey:          "ci_failure:" + sourceID,
		Title:            title,
		Reason:           reason,
		Impact:           "high",
		Status:           "action_required",
		SourceType:       "ci_run",
		SourceID:         sourceID,
		SuggestedActions: []string{"inspect_logs", "rerun_ci"},
		DetectedAt:       detectedAt,
	}
}

type webhookPayload struct {
	Repository  *repositoryPayload  `json:"repository"`
	Sender      *userPayload        `json:"sender"`
	Issue       *issuePayload       `json:"issue"`
	PullRequest *pullRequestPayload `json:"pull_request"`
	ActionRun   *ciRunPayload       `json:"action_run"`
	WorkflowRun *ciRunPayload       `json:"workflow_run"`
}

type repositoryPayload struct {
	ID            int64        `json:"id"`
	FullName      string       `json:"full_name"`
	Name          string       `json:"name"`
	CloneURL      string       `json:"clone_url"`
	HTMLURL       string       `json:"html_url"`
	DefaultBranch string       `json:"default_branch"`
	Private       bool         `json:"private"`
	Owner         *userPayload `json:"owner"`
}

type userPayload struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	UserName  string `json:"username"`
	FullName  string `json:"full_name"`
	AvatarURL string `json:"avatar_url"`
	HTMLURL   string `json:"html_url"`
}

type issuePayload struct {
	ID        int64        `json:"id"`
	Index     int64        `json:"index"`
	Number    int64        `json:"number"`
	Title     string       `json:"title"`
	State     string       `json:"state"`
	User      *userPayload `json:"user"`
	Assignee  *userPayload `json:"assignee"`
	HTMLURL   string       `json:"html_url"`
	CreatedAt *time.Time   `json:"created_at"`
	ClosedAt  *time.Time   `json:"closed_at"`
}

type pullRequestPayload struct {
	ID       int64          `json:"id"`
	Index    int64          `json:"index"`
	Number   int64          `json:"number"`
	Title    string         `json:"title"`
	State    string         `json:"state"`
	User     *userPayload   `json:"user"`
	Head     *branchPayload `json:"head"`
	Base     *branchPayload `json:"base"`
	HTMLURL  string         `json:"html_url"`
	MergedAt *time.Time     `json:"merged_at"`
	ClosedAt *time.Time     `json:"closed_at"`
}

type branchPayload struct {
	Ref string `json:"ref"`
	SHA string `json:"sha"`
}

type ciRunPayload struct {
	ID              any        `json:"id"`
	RunNumber       any        `json:"run_number"`
	Status          string     `json:"status"`
	Conclusion      string     `json:"conclusion"`
	HeadBranch      string     `json:"head_branch"`
	HeadSHA         string     `json:"head_sha"`
	HTMLURL         string     `json:"html_url"`
	RunStartedAt    *time.Time `json:"run_started_at"`
	StartedAt       *time.Time `json:"started_at"`
	UpdatedAt       *time.Time `json:"updated_at"`
	CompletedAt     *time.Time `json:"completed_at"`
	DurationSeconds *int       `json:"duration_seconds"`
}

func (p webhookPayload) repository() *domain.Repository {
	if p.Repository == nil {
		return nil
	}
	fullName := p.Repository.FullName
	if fullName == "" && p.Repository.Owner != nil && p.Repository.Name != "" {
		fullName = loginOf(p.Repository.Owner) + "/" + p.Repository.Name
	}
	if fullName == "" {
		fullName = p.Repository.Name
	}
	name := p.Repository.Name
	if name == "" && strings.Contains(fullName, "/") {
		parts := strings.Split(fullName, "/")
		name = parts[len(parts)-1]
	}
	repository := &domain.Repository{
		GiteaID:       p.Repository.ID,
		FullName:      fullName,
		Name:          name,
		CloneURL:      p.Repository.CloneURL,
		HTMLURL:       p.Repository.HTMLURL,
		DefaultBranch: p.Repository.DefaultBranch,
		Private:       p.Repository.Private,
	}
	if p.Repository.Owner != nil {
		repository.OwnerLogin = loginOf(p.Repository.Owner)
	} else if strings.Contains(fullName, "/") {
		repository.OwnerLogin = strings.SplitN(fullName, "/", 2)[0]
	}
	return repository
}

func (p webhookPayload) sender() *domain.User {
	if p.Sender == nil || loginOf(p.Sender) == "" {
		return nil
	}
	return userFromPayload(p.Sender)
}

func (p webhookPayload) issue() *domain.Issue {
	if p.Issue == nil {
		return nil
	}
	return &domain.Issue{
		GiteaID:       p.Issue.ID,
		Number:        firstInt64(p.Issue.Number, p.Issue.Index),
		Title:         p.Issue.Title,
		State:         normalizeIssueState(p.Issue.State),
		AuthorLogin:   loginOf(p.Issue.User),
		AssigneeLogin: loginOf(p.Issue.Assignee),
		HTMLURL:       p.Issue.HTMLURL,
		OpenedAt:      p.Issue.CreatedAt,
		ClosedAt:      p.Issue.ClosedAt,
	}
}

func (p webhookPayload) pullRequest() *domain.PullRequest {
	if p.PullRequest == nil {
		return nil
	}
	pullRequest := &domain.PullRequest{
		GiteaID:     p.PullRequest.ID,
		Number:      firstInt64(p.PullRequest.Number, p.PullRequest.Index),
		Title:       p.PullRequest.Title,
		State:       normalizePullRequestState(p.PullRequest.State, p.PullRequest.MergedAt),
		AuthorLogin: loginOf(p.PullRequest.User),
		HTMLURL:     p.PullRequest.HTMLURL,
		MergedAt:    p.PullRequest.MergedAt,
		ClosedAt:    p.PullRequest.ClosedAt,
	}
	if p.PullRequest.Head != nil {
		pullRequest.HeadBranch = p.PullRequest.Head.Ref
		pullRequest.HeadSHA = p.PullRequest.Head.SHA
	}
	if p.PullRequest.Base != nil {
		pullRequest.BaseBranch = p.PullRequest.Base.Ref
	}
	return pullRequest
}

func (p webhookPayload) ciRun(repository *domain.Repository) *domain.CIRun {
	run := p.ActionRun
	if run == nil {
		run = p.WorkflowRun
	}
	if run == nil {
		return nil
	}
	externalID := stringify(run.ID)
	if externalID == "" {
		externalID = stringify(run.RunNumber)
	}
	if externalID == "" {
		return nil
	}
	repositoryName := ""
	if repository != nil {
		repositoryName = repository.FullName
	}
	return &domain.CIRun{
		ExternalID:      externalID,
		RepositoryName:  repositoryName,
		Branch:          run.HeadBranch,
		CommitSHA:       run.HeadSHA,
		Status:          normalizeCIStatus(run.Status, run.Conclusion),
		Conclusion:      run.Conclusion,
		StartedAt:       firstTime(run.StartedAt, run.RunStartedAt),
		FinishedAt:      firstTime(run.CompletedAt, run.UpdatedAt),
		DurationSeconds: run.DurationSeconds,
		HTMLURL:         run.HTMLURL,
	}
}

func userFromPayload(user *userPayload) *domain.User {
	return &domain.User{
		GiteaID:     user.ID,
		Login:       loginOf(user),
		DisplayName: user.FullName,
		AvatarURL:   user.AvatarURL,
		HTMLURL:     user.HTMLURL,
	}
}

func loginOf(user *userPayload) string {
	if user == nil {
		return ""
	}
	if user.Login != "" {
		return user.Login
	}
	return user.UserName
}

func firstInt64(values ...int64) int64 {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}

func firstTime(values ...*time.Time) *time.Time {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}

func stringify(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case float64:
		return strconv.FormatInt(int64(typed), 10)
	default:
		return fmt.Sprint(typed)
	}
}

func normalizeIssueState(state string) string {
	state = strings.ToLower(state)
	if state == "closed" {
		return "closed"
	}
	return "open"
}

func normalizePullRequestState(state string, mergedAt *time.Time) string {
	if mergedAt != nil {
		return "merged"
	}
	state = strings.ToLower(state)
	if state == "closed" {
		return "closed"
	}
	return "open"
}

func normalizeCIStatus(status, conclusion string) string {
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
	case "":
		return "unknown"
	default:
		return "unknown"
	}
}
