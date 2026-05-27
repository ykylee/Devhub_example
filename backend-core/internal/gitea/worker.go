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
	AcquireNextQueuedSyncJob(ctx context.Context) (jobID string, providerID string, err error)
	UpdateIntegrationSyncJobStatus(ctx context.Context, jobID string, status string) error
	// GetIntegrationProviderByID — Phase 3: queued job 의 provider 별 base_url +
	// api_token 을 조회해 env 대신 사용 (per-provider sync 연결).
	GetIntegrationProviderByID(ctx context.Context, providerID string) (domain.IntegrationProvider, error)
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
	// 1. queued sync job 우선. AcquireNextQueuedSyncJob 은 provider_type='scm'
	// gated (codex #341 P1) — 비-SCM job 은 도달하지 않음.
	jobID, providerID, err := w.Store.AcquireNextQueuedSyncJob(ctx)
	if err == nil && jobID != "" {
		// Phase 3: 등록된 provider 의 base_url + api_token 우선, 없으면 env fallback.
		baseURL, token := w.resolveSyncConfig(ctx, providerID)
		if strings.TrimSpace(baseURL) == "" || strings.TrimSpace(token) == "" {
			log.Printf("[Gitea Sync Worker] Job %s (provider %s) skipped: base_url/api_token 미설정 (provider·env 모두)", jobID, providerID)
			_ = w.Store.UpdateIntegrationSyncJobStatus(ctx, jobID, "failed")
			return nil
		}
		log.Printf("[Gitea Sync Worker] Acquired queued job %s (provider %s). Starting sync...", jobID, providerID)
		if syncErr := w.syncAllWith(ctx, baseURL, token); syncErr != nil {
			log.Printf("[Gitea Sync Worker] Job %s failed: %v", jobID, syncErr)
			_ = w.Store.UpdateIntegrationSyncJobStatus(ctx, jobID, "failed")
			return syncErr
		}
		log.Printf("[Gitea Sync Worker] Job %s succeeded", jobID)
		_ = w.Store.UpdateIntegrationSyncJobStatus(ctx, jobID, "succeeded")
		return nil
	}

	// 2. 큐가 비면 env(GITEA_URL/GITEA_TOKEN) 기반 주기 sync (backward compat).
	// env 미설정이면 no-op (queued job 만 per-provider 로 처리).
	if strings.TrimSpace(w.GiteaURL) == "" || strings.TrimSpace(w.GiteaToken) == "" {
		return nil
	}
	return w.syncAllWith(ctx, w.GiteaURL, w.GiteaToken)
}

// resolveSyncConfig — providerID 의 base_url + api_token 을 우선 사용하고, 비어 있으면
// worker 의 env 값으로 fallback (Phase 3). provider lookup 실패 시에도 env fallback.
func (w *SyncWorker) resolveSyncConfig(ctx context.Context, providerID string) (string, string) {
	baseURL, token := w.GiteaURL, w.GiteaToken
	if strings.TrimSpace(providerID) == "" {
		return baseURL, token
	}
	prov, err := w.Store.GetIntegrationProviderByID(ctx, providerID)
	if err != nil {
		log.Printf("[Gitea Sync Worker] provider %s lookup 실패, env fallback: %v", providerID, err)
		return baseURL, token
	}
	if strings.TrimSpace(prov.BaseURL) != "" {
		baseURL = prov.BaseURL
	}
	if strings.TrimSpace(prov.APIToken) != "" {
		token = prov.APIToken
	}
	return baseURL, token
}

func (w *SyncWorker) syncAllWith(ctx context.Context, baseURL, token string) error {
	client := NewClient(baseURL, token)
	syncer := NewSyncer(w.Store)

	// Fetch all Gitea user repositories first to build local cache mapping.
	repos, err := client.ListUserRepos(ctx)
	if err != nil {
		return fmt.Errorf("failed to list user repos: %w", err)
	}

	var syncErrors []string

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
			syncErrors = append(syncErrors, fmt.Sprintf("repo metadata upsert %s: %v", repo.FullName, err))
			continue
		}

		// Perform deep sync on issues and PRs for this repository.
		parts := strings.SplitN(repo.FullName, "/", 2)
		if len(parts) == 2 {
			err = syncer.SyncRepository(ctx, client, parts[0], parts[1])
			if err != nil {
				log.Printf("[Gitea Sync Worker] Failed to sync repo details for %s: %v", repo.FullName, err)
				syncErrors = append(syncErrors, fmt.Sprintf("repo sync %s: %v", repo.FullName, err))
			}
		}
	}

	if len(syncErrors) > 0 {
		return fmt.Errorf("Gitea sync completed with errors: %s", strings.Join(syncErrors, "; "))
	}

	return nil
}
