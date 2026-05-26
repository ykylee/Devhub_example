package gitea

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/devhub/backend-core/internal/domain"
)

// SyncJobStore defines store interactions for managing background SCM sync jobs.
type SyncJobStore interface {
	SyncStore
}

// SyncWorker periodically checks for IntegrationSyncJob requests and handles syncing.
type SyncWorker struct {
	Store      SyncJobStore
	GiteaURL   string
	GiteaToken string
}

// NewSyncWorker creates a SyncWorker.
func NewSyncWorker(store SyncJobStore, giteaURL, giteaToken string) *SyncWorker {
	return &SyncWorker{
		Store:      store,
		GiteaURL:   giteaURL,
		GiteaToken: giteaToken,
	}
}

// Run starts the background sync loop.
func (w *SyncWorker) Run(ctx context.Context, interval time.Duration) error {
	if interval <= 0 {
		interval = 30 * time.Second
	}

	log.Printf("[Gitea Sync Worker] Starting background worker with interval: %s", interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := w.ProcessOnce(ctx); err != nil {
				log.Printf("[Gitea Sync Worker] Error in sync process: %v", err)
			}
		}
	}
}

// ProcessOnce runs a single round of synchronizing repositories.
func (w *SyncWorker) ProcessOnce(ctx context.Context) error {
	// Homelab Gitea is active, let's read the credentials/bindings to sync.
	if strings.TrimSpace(w.GiteaURL) == "" || strings.TrimSpace(w.GiteaToken) == "" {
		return nil
	}

	client := NewClient(w.GiteaURL, w.GiteaToken)
	syncer := NewSyncer(w.Store)

	// Fetch all Gitea user repositories first to build local cache mapping.
	repos, err := client.ListUserRepos(ctx)
	if err != nil {
		return fmt.Errorf("failed to list user repos: %w", err)
	}

	for _, repo := range repos {
		// Sync local repositories table first.
		err = w.Store.UpsertRepository(ctx, domain.Repository{
			GiteaID:       repo.ID,
			FullName:      repo.FullName,
			Name:          repo.Name,
			CloneURL:      repo.CloneURL,
			HTMLURL:       repo.HTMLURL,
			DefaultBranch: repo.DefaultBranch,
			Private:       repo.Private,
		})
		if err != nil {
			log.Printf("[Gitea Sync Worker] Failed to upsert repo metadata for %s: %v", repo.FullName, err)
			continue
		}

		// Perform deep sync on issues and PRs for this repository.
		parts := strings.SplitN(repo.FullName, "/", 2)
		if len(parts) == 2 {
			err = syncer.SyncRepository(ctx, client, parts[0], parts[1])
			if err != nil {
				log.Printf("[Gitea Sync Worker] Failed to sync repo details for %s: %v", repo.FullName, err)
			}
		}
	}

	return nil
}
