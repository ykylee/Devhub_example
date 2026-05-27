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
		// Phase 3: 등록된 provider 의 base_url + auth_mode 별 자격증명 우선, 없으면 env fallback.
		baseURL, auth := w.resolveSyncConfig(ctx, providerID)
		if strings.TrimSpace(baseURL) == "" {
			log.Printf("[Gitea Sync Worker] Job %s (provider %s) skipped: base_url 미설정 (provider·env 모두)", jobID, providerID)
			_ = w.Store.UpdateIntegrationSyncJobStatus(ctx, jobID, "failed")
			return nil
		}
		client, err := NewClientForAuth(ctx, baseURL, auth)
		if err != nil {
			log.Printf("[Gitea Sync Worker] Job %s (provider %s) auth 실패 (%s mode): %v", jobID, providerID, auth.Mode, err)
			_ = w.Store.UpdateIntegrationSyncJobStatus(ctx, jobID, "failed")
			return err
		}
		if client == nil {
			log.Printf("[Gitea Sync Worker] Job %s (provider %s) skipped: %s 자격증명 미설정", jobID, providerID, auth.Mode)
			_ = w.Store.UpdateIntegrationSyncJobStatus(ctx, jobID, "failed")
			return nil
		}
		log.Printf("[Gitea Sync Worker] Acquired queued job %s (provider %s, auth=%s). Starting sync...", jobID, providerID, auth.Mode)
		if syncErr := w.syncAllWith(ctx, client, providerID); syncErr != nil {
			log.Printf("[Gitea Sync Worker] Job %s failed: %v", jobID, syncErr)
			_ = w.Store.UpdateIntegrationSyncJobStatus(ctx, jobID, "failed")
			return syncErr
		}
		log.Printf("[Gitea Sync Worker] Job %s succeeded", jobID)
		_ = w.Store.UpdateIntegrationSyncJobStatus(ctx, jobID, "succeeded")
		return nil
	}

	// 2. 큐가 비면 env(GITEA_URL/GITEA_TOKEN) 기반 주기 sync (backward compat, token mode).
	// env 미설정이면 no-op (queued job 만 per-provider 로 처리).
	if strings.TrimSpace(w.GiteaURL) == "" || strings.TrimSpace(w.GiteaToken) == "" {
		return nil
	}
	return w.syncAllWith(ctx, NewClient(w.GiteaURL, w.GiteaToken), "")
}

// resolveSyncConfig — providerID 가 명시되면 그 provider 의 base_url + auth_mode 별
// outbound 자격증명만 사용한다. providerID 가 비었거나 lookup 실패 시에만 worker 의
// env 값(token mode)으로 fallback (legacy 주기 sync).
//
// 중요 (codex #358 P1): 명시 provider 가 해석되면 env token 으로 **fallback 하지 않는다**.
// provider 고유 host(base_url)에 worker 전역 GITEA_TOKEN 을 보내면 잘못된 계정으로
// 인증하거나 다른 외부 host 로 공유 토큰이 유출될 수 있다. 자격증명이 불완전하거나
// (agent / 미완성 basic·oauth2) base_url 이 비면 ProcessOnce 가 job 을 실패 처리한다
// (NewClientForAuth 가 nil 반환).
func (w *SyncWorker) resolveSyncConfig(ctx context.Context, providerID string) (string, domain.OutboundAuth) {
	envAuth := domain.OutboundAuth{Mode: domain.IntegrationAuthModeToken, Token: w.GiteaToken}
	if strings.TrimSpace(providerID) == "" {
		return w.GiteaURL, envAuth
	}
	prov, err := w.Store.GetIntegrationProviderByID(ctx, providerID)
	if err != nil {
		log.Printf("[Gitea Sync Worker] provider %s lookup 실패, env fallback: %v", providerID, err)
		return w.GiteaURL, envAuth
	}
	// 명시 provider — 그 provider 의 base_url + 자체 자격증명만. env 혼용 금지.
	return prov.BaseURL, prov.ResolveOutboundAuth()
}

func (w *SyncWorker) syncAllWith(ctx context.Context, client *Client, providerID string) error {
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
			Source:        domain.RepositorySourceSCM,
			ProviderID:    providerID,
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
