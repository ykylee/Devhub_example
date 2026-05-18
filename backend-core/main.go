package main

import (
	"context"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/devhub/backend-core/internal/auth"
	"github.com/devhub/backend-core/internal/commandworker"
	"github.com/devhub/backend-core/internal/config"
	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/hrdb"
	"github.com/devhub/backend-core/internal/httpapi"
	"github.com/devhub/backend-core/internal/integrations/adapters"
	"github.com/devhub/backend-core/internal/normalize"
	"github.com/devhub/backend-core/internal/serviceaction"
	"github.com/devhub/backend-core/internal/store"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	var eventStore httpapi.WebhookEventStore
	var eventProcessor httpapi.WebhookEventProcessor
	var healthStore httpapi.HealthStore
	var domainStore httpapi.DomainStore
	var commandStore httpapi.CommandStore
	var auditStore httpapi.AuditStore
	var organizationStore httpapi.OrganizationStore
	var applicationStore httpapi.ApplicationStore
	var devRequestStore httpapi.DevRequestStore
	var devRequestIntakeTokenStore httpapi.IntakeTokenStore
	var rbacStore httpapi.RBACStore
	realtimeHub := httpapi.NewRealtimeHub()
	var worker *commandworker.Worker
	var liveWorker *commandworker.LiveWorker
	var homeLabAdapterStore adapters.InfraSnapshotStore

	if cfg.DBURL != "" {
		pgStore, err := store.NewPostgresStore(ctx, cfg.DBURL)
		if err != nil {
			log.Fatalf("connect postgres: %v", err)
		}
		defer pgStore.Close()
		eventStore = pgStore
		eventProcessor = normalize.Processor{Sink: pgStore}
		healthStore = pgStore
		domainStore = pgStore
		commandStore = pgStore
		auditStore = pgStore
		organizationStore = pgStore
		applicationStore = pgStore
		devRequestStore = pgStore
		devRequestIntakeTokenStore = pgStore
		rbacStore = pgStore
		homeLabAdapterStore = pgStore

		worker = &commandworker.Worker{Store: pgStore, Publisher: realtimeHub}
		if cfg.ServiceActionExecutorMode != "" {
			executor, err := serviceaction.NewExecutor(
				cfg.ServiceActionExecutorMode,
				cfg.ServiceActionAllowedServices,
				cfg.ServiceActionAllowedActions,
			)
			if err != nil {
				log.Fatalf("configure service action executor: %v", err)
			}
			liveWorker = &commandworker.LiveWorker{Store: pgStore, Executor: executor, Publisher: realtimeHub}
			log.Printf("service action executor enabled in %s mode", cfg.ServiceActionExecutorMode)
		}
	} else {
		log.Fatalf("DB_URL is not set; startup refused")
	}

	var verifier httpapi.BearerTokenVerifier
	if cfg.HydraAdminURL != "" {
		parsed, err := url.Parse(cfg.HydraAdminURL)
		if err != nil {
			log.Fatalf("startup refused: DEVHUB_HYDRA_ADMIN_URL is not a valid URL: %v", err)
		}
		verifier = &auth.HydraIntrospectionVerifier{
			AdminURL:  cfg.HydraAdminURL,
			RoleClaim: cfg.HydraRoleClaim,
		}
		log.Printf("bearer token verifier: hydra introspection at %s (role_claim=%q)", parsed.Redacted(), cfg.HydraRoleClaim)
	}
	if err := cfg.Validate(verifier != nil); err != nil {
		log.Fatalf("startup refused: %v", err)
	}

	var (
		hydraAdmin   httpapi.HydraLoginAdmin
		hydraLogout  httpapi.HydraLogoutAdmin
		hydraToken   httpapi.HydraTokenExchanger
		hydraRevoker httpapi.HydraTokenRevoker
		kratosLogin  httpapi.KratosLoginClient
		kratosAdmin  httpapi.KratosAdmin
	)
	if cfg.HydraAdminURL != "" {
		adminClient := &httpapi.HydraAdminClient{AdminURL: cfg.HydraAdminURL}
		hydraAdmin = adminClient
		hydraLogout = adminClient
	}
	if cfg.HydraPublicURL != "" {
		tokenClient := &httpapi.HydraTokenClient{PublicURL: cfg.HydraPublicURL}
		hydraToken = tokenClient
		hydraRevoker = tokenClient
	}
	if cfg.KratosPublicURL != "" {
		kratosLogin = &httpapi.KratosClient{PublicURL: cfg.KratosPublicURL}
	}
	if cfg.KratosAdminURL != "" {
		kratosAdmin = &httpapi.KratosAdminClient{AdminURL: cfg.KratosAdminURL}
	} else {
		kratosAdmin = &httpapi.MockKratosAdmin{}
		log.Println("Kratos Admin URL not set; using MockKratosAdmin for development")
	}

	// Seed local admin account for development using regular APIs
	if cfg.AuthDevFallback && kratosAdmin != nil && organizationStore != nil {
		seedLocalAdmin(ctx, kratosAdmin, organizationStore)
	}

	hrdbMock := hrdb.NewMockClient()
	log.Println("HR DB Mock client initialized")

	router := httpapi.NewRouter(httpapi.RouterConfig{
		WebhookSecret:              cfg.GiteaWebhookSecret,
		KratosWebhookToken:         cfg.KratosWebhookToken,
		InfraAgentToken:            cfg.InfraAgentToken,
		HomeLabProviderKey:         cfg.HomeLabProviderKey,
		HomeLabDegradedRaw:         cfg.HomeLabDegradedStatuses,
		EventStore:                 eventStore,
		EventProcessor:             eventProcessor,
		HealthStore:                healthStore,
		DomainStore:                domainStore,
		CommandStore:               commandStore,
		AuditStore:                 auditStore,
		OrganizationStore:          organizationStore,
		ApplicationStore:           applicationStore,
		DevRequestStore:            devRequestStore,
		DevRequestIntakeTokenStore: devRequestIntakeTokenStore,
		RBACStore:                  rbacStore,
		BearerTokenVerifier:        verifier,
		KratosLogin:                kratosLogin,
		HydraAdmin:                 hydraAdmin,
		HydraLogout:                hydraLogout,
		HydraToken:                 hydraToken,
		HydraRevoker:               hydraRevoker,
		KratosAdmin:                kratosAdmin,
		HRDB:                       hrdbMock,
		SnapshotProvider: httpapi.RuntimeSnapshotProvider{
			Base:         httpapi.StaticSnapshotProvider{},
			HealthStore:  healthStore,
			GiteaURL:     cfg.GiteaURL,
			BackendAIURL: cfg.BackendAIURL,
		},
		RealtimeHub:     realtimeHub,
		AuthDevFallback: cfg.AuthDevFallback,
	})

	if worker != nil {
		go func() {
			if err := worker.Run(ctx, 2*time.Second); err != nil && err != context.Canceled {
				log.Printf("command worker stopped: %v", err)
			}
		}()
	}
	if liveWorker != nil {
		go func() {
			if err := liveWorker.Run(ctx, 2*time.Second); err != nil && err != context.Canceled {
				log.Printf("live service action worker stopped: %v", err)
			}
		}()
	}
	if cfg.HomeLabPullEnabled && homeLabAdapterStore != nil {
		policy := buildHomeLabHealthPolicy(cfg.HomeLabProviderKey, cfg.HomeLabDegradedStatuses)
		var (
			puller     adapters.HomeLabPuller
			pullerDesc string
		)
		maxBytes := cfg.HomeLabPullMaxBytes
		if maxBytes < 0 {
			maxBytes = 0
		}
		if filePath := strings.TrimSpace(cfg.HomeLabPullFile); filePath != "" {
			puller = adapters.HomeLabFilePuller{Path: filePath, MaxBytes: maxBytes}
			pullerDesc = "file=" + filePath
		} else if endpoint := strings.TrimSpace(cfg.HomeLabPullURL); endpoint != "" {
			retryBackoff := time.Second
			if strings.TrimSpace(cfg.HomeLabPullHTTPRetryBackoff) != "" {
				if parsed, err := time.ParseDuration(cfg.HomeLabPullHTTPRetryBackoff); err == nil && parsed > 0 {
					retryBackoff = parsed
				} else {
					log.Printf("invalid DEVHUB_HOMELAB_PULL_HTTP_RETRY_BACKOFF=%q; fallback to %s", cfg.HomeLabPullHTTPRetryBackoff, retryBackoff)
				}
			}
			retryMax := cfg.HomeLabPullHTTPRetryMax
			if retryMax < 0 {
				retryMax = 0
			}
			puller = adapters.HomeLabHTTPPuller{
				URL:          endpoint,
				Token:        cfg.HomeLabPullToken,
				RetryMax:     retryMax,
				RetryBackoff: retryBackoff,
				MaxBytes:     maxBytes,
			}
			pullerDesc = "url=" + endpoint
		}
		if puller == nil {
			log.Printf("homelab pull loop skipped: set DEVHUB_HOMELAB_PULL_FILE or DEVHUB_HOMELAB_PULL_URL")
		} else {
			adapter := adapters.HomeLabAdapter{
				Store:        homeLabAdapterStore,
				Puller:       puller,
				HealthPolicy: policy,
			}
			interval := 30 * time.Second
			if strings.TrimSpace(cfg.HomeLabPullInterval) != "" {
				if parsed, err := time.ParseDuration(cfg.HomeLabPullInterval); err == nil && parsed > 0 {
					interval = parsed
				} else {
					log.Printf("invalid DEVHUB_HOMELAB_PULL_INTERVAL=%q; fallback to %s", cfg.HomeLabPullInterval, interval)
				}
			}
			go func() {
				err := adapters.RunHomeLabPullLoop(ctx, adapter, interval, func(runErr error) {
					log.Printf("homelab pull loop error: %v", runErr)
				})
				if err != nil && err != context.Canceled {
					log.Printf("homelab pull loop stopped: %v", err)
				}
			}()
			log.Printf("homelab pull loop enabled (%s interval=%s)", pullerDesc, interval)
		}
	}
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("run server: %v", err)
	}
}

func buildHomeLabHealthPolicy(providerKey, degradedRaw string) adapters.HomeLabHealthPolicy {
	statuses := map[string]bool{}
	for _, item := range strings.Split(degradedRaw, ",") {
		status := strings.ToLower(strings.TrimSpace(item))
		if status == "" {
			continue
		}
		statuses[status] = true
	}
	if len(statuses) == 0 {
		statuses = map[string]bool{"warning": true, "degraded": true, "down": true}
	}
	key := strings.TrimSpace(providerKey)
	if key == "" {
		key = "homelab-agent"
	}
	return adapters.HomeLabHealthPolicy{
		DegradedStatuses: statuses,
		ProviderKey:      key,
	}
}

// seedOrgStore is the narrow subset of httpapi.OrganizationStore that
// seedLocalAdmin actually drives. Keeping it local lets the seed unit
// tests use a 2-method fake instead of stubbing the full 14-method
// interface.
type seedOrgStore interface {
	CreateUser(context.Context, domain.CreateUserInput) (domain.AppUser, error)
	SetKratosIdentityID(context.Context, string, string) error
}

func seedLocalAdmin(ctx context.Context, kratosAdmin httpapi.KratosAdmin, orgStore seedOrgStore) {
	const (
		adminLogin    = "test"
		adminEmail    = "test@example.com"
		adminName     = "Test Admin"
		adminPassword = "test"
	)

	// 1. Kratos Identity — CreateIdentity 가 409 (already exists) 등으로
	// 실패하면 기존 identity 를 찾아 재사용. 양쪽 모두 실패하면 시딩
	// 포기 (Kratos down / 스키마 미스매치 같은 운영 이슈를 묻히지 않도록
	// Find 에러까지 같이 surface).
	kratosID, err := kratosAdmin.CreateIdentity(ctx, adminEmail, adminName, adminLogin, adminPassword)
	if err != nil {
		log.Printf("[seedLocalAdmin] CreateIdentity for %q failed: %v; falling back to FindIdentityByUserID", adminLogin, err)
		existing, findErr := kratosAdmin.FindIdentityByUserID(ctx, adminLogin)
		if findErr != nil {
			log.Printf("[seedLocalAdmin] FindIdentityByUserID for %q also failed: %v; aborting seed", adminLogin, findErr)
			return
		}
		kratosID = existing
	}
	if kratosID == "" {
		log.Printf("[seedLocalAdmin] resolved empty Kratos ID for %q; aborting seed", adminLogin)
		return
	}

	// 2. DevHub User
	_, err = orgStore.CreateUser(ctx, domain.CreateUserInput{
		UserID:      adminLogin,
		Email:       adminEmail,
		DisplayName: adminName,
		Role:        domain.AppRoleSystemAdmin,
		Status:      domain.UserStatusActive,
		Type:        domain.UserTypeHuman,
	})
	if err != nil {
		log.Printf("[seedLocalAdmin] DB User creation failed or skipped: %v", err)
	}

	// 3. Link
	err = orgStore.SetKratosIdentityID(ctx, adminLogin, kratosID)
	if err != nil {
		log.Printf("[seedLocalAdmin] Failed to link Kratos ID: %v", err)
	} else {
		log.Printf("[seedLocalAdmin] Successfully ensured test admin '%s' via regular APIs", adminLogin)
	}
}
