package gitea

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/devhub/backend-core/internal/domain"
)

// SyncStore contract for inserting normalized Gitea SCM entities.
type SyncStore interface {
	UpsertRepository(ctx context.Context, repository domain.Repository) error
	UpsertUser(ctx context.Context, user domain.User) error
	UpsertIssue(ctx context.Context, issue domain.Issue) error
	UpsertPullRequest(ctx context.Context, pr domain.PullRequest) error
}

// Syncer orchestrates the pulling and upserting of Gitea repos, issues, and PRs.
type Syncer struct {
	Store SyncStore
}

// NewSyncer creates a new Syncer.
func NewSyncer(store SyncStore) *Syncer {
	return &Syncer{Store: store}
}

// SyncRepository fetches all metadata, open/closed issues, and PRs for a repo and upserts them.
func (s *Syncer) SyncRepository(ctx context.Context, client *Client, owner, repoName string) error {
	log.Printf("[Gitea Sync] Starting sync for repository: %s/%s", owner, repoName)

	// 1. Sync Issues (Open & Closed)
	for _, state := range []string{"open", "closed"} {
		issues, err := client.ListIssues(ctx, owner, repoName, state)
		if err != nil {
			return fmt.Errorf("failed to fetch Gitea issues (%s): %w", state, err)
		}

		for _, issue := range issues {
			if issue.User != nil {
				_ = s.Store.UpsertUser(ctx, domain.User{
					GiteaID:     issue.User.ID,
					Login:       issue.User.Login,
					DisplayName: issue.User.FullName,
					AvatarURL:   issue.User.AvatarURL,
					HTMLURL:     issue.User.HTMLURL,
				})
			}
			if issue.Assignee != nil {
				_ = s.Store.UpsertUser(ctx, domain.User{
					GiteaID:     issue.Assignee.ID,
					Login:       issue.Assignee.Login,
					DisplayName: issue.Assignee.FullName,
					AvatarURL:   issue.Assignee.AvatarURL,
					HTMLURL:     issue.Assignee.HTMLURL,
				})
			}

			// We need repo full name / ID mapped in the normalize phase
			err = s.Store.UpsertIssue(ctx, domain.Issue{
				GiteaID:            issue.ID,
				RepositoryName:     owner + "/" + repoName,
				Number:             issue.Number,
				Title:              issue.Title,
				State:              issue.State,
				AuthorLogin:        getStringUserLogin(issue.User),
				AssigneeLogin:      getStringUserLogin(issue.Assignee),
				HTMLURL:            issue.HTMLURL,
				OpenedAt:           &issue.CreatedAt,
				ClosedAt:           issue.ClosedAt,
			})
			if err != nil {
				log.Printf("[Gitea Sync] Warning: failed to upsert issue #%d: %v", issue.Number, err)
			}
		}
	}

	// 2. Sync Pull Requests (Open & Closed)
	for _, state := range []string{"open", "closed"} {
		pulls, err := client.ListPullRequests(ctx, owner, repoName, state)
		if err != nil {
			return fmt.Errorf("failed to fetch Gitea PRs (%s): %w", state, err)
		}

		for _, pr := range pulls {
			if pr.User != nil {
				_ = s.Store.UpsertUser(ctx, domain.User{
					GiteaID:     pr.User.ID,
					Login:       pr.User.Login,
					DisplayName: pr.User.FullName,
					AvatarURL:   pr.User.AvatarURL,
					HTMLURL:     pr.User.HTMLURL,
				})
			}

			headBranch := ""
			headSHA := ""
			if pr.Head != nil {
				headBranch = pr.Head.Ref
				headSHA = pr.Head.SHA
			}
			baseBranch := ""
			if pr.Base != nil {
				baseBranch = pr.Base.Ref
			}

			err = s.Store.UpsertPullRequest(ctx, domain.PullRequest{
				GiteaID:            pr.ID,
				RepositoryName:     owner + "/" + repoName,
				Number:             pr.Number,
				Title:              pr.Title,
				State:              normalizePullRequestState(pr.State, pr.MergedAt),
				AuthorLogin:        getStringUserLogin(pr.User),
				HeadBranch:         headBranch,
				BaseBranch:         baseBranch,
				HeadSHA:            headSHA,
				HTMLURL:            pr.HTMLURL,
				MergedAt:           pr.MergedAt,
				ClosedAt:           pr.ClosedAt,
			})
			if err != nil {
				log.Printf("[Gitea Sync] Warning: failed to upsert PR #%d: %v", pr.Number, err)
			}
		}
	}

	log.Printf("[Gitea Sync] Successfully completed sync for repository: %s/%s", owner, repoName)
	return nil
}

func getStringUserLogin(u *GiteaUser) string {
	if u == nil {
		return ""
	}
	return u.Login
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
