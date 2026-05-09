package main

import (
	"context"
	"log"
	"net/url"
	"time"

	"github.com/devhub/backend-core/internal/auth"
	"github.com/devhub/backend-core/internal/commandworker"
	"github.com/devhub/backend-core/internal/config"
	"github.com/devhub/backend-core/internal/hrdb"
	"github.com/devhub/backend-core/internal/httpapi"
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
	var rbacStore httpapi.RBACStore
	realtimeHub := httpapi.NewRealtimeHub()
	var worker *commandworker.Worker
	var liveWorker *commandworker.LiveWorker
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
		rbacStore = pgStore
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
		log.Println("DB_URL is not set; webhook persistence is disabled")
	}

	var verifier httpapi.BearerTokenVerifier
	if cfg.HydraAdminURL != "" {
		parsed, err := url.Parse(cfg.HydraAdminURL)
		if err != nil {
			log.Fatalf("startup refused: DEVHUB_HYDRA_ADMIN_URL is not a valid URL: %v", err)
		}
		if parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			log.Fatalf("startup refused: DEVHUB_HYDRA_ADMIN_URL must be an absolute http(s) URL: got %s", parsed.Redacted())
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

	// Auth proxy clients are only wired when both Hydra admin and Kratos
	// public URLs are configured. Assigning typed nil pointers to interface
	// fields would defeat the handler's `cfg.KratosLogin == nil` guard, so
	// we leave the fields untouched when either env var is missing.
	var (
		hydraAdmin  httpapi.HydraLoginAdmin
		hydraToken  httpapi.HydraTokenExchanger
		kratosLogin httpapi.KratosLoginClient
		kratosAdmin httpapi.KratosAdmin
	)
	if cfg.HydraAdminURL != "" {
		hydraAdmin = &httpapi.HydraAdminClient{AdminURL: cfg.HydraAdminURL}
		log.Printf("hydra admin client wired: %s", cfg.HydraAdminURL)
	}
	if cfg.HydraPublicURL != "" {
		hydraToken = &httpapi.HydraTokenClient{PublicURL: cfg.HydraPublicURL}
		log.Printf("hydra public token client wired: %s", cfg.HydraPublicURL)
	}
	if cfg.KratosPublicURL != "" {
		kratosLogin = &httpapi.KratosClient{PublicURL: cfg.KratosPublicURL}
		log.Printf("kratos public client wired: %s", cfg.KratosPublicURL)
	}
	if cfg.KratosAdminURL != "" {
		kratosAdmin = &httpapi.KratosAdminClient{AdminURL: cfg.KratosAdminURL}
		log.Printf("kratos admin client wired: %s", cfg.KratosAdminURL)
	} else {
		kratosAdmin = &httpapi.MockKratosAdmin{}
		log.Println("Kratos Admin URL not set; using MockKratosAdmin for development")
	}

	hrdbMock := hrdb.NewMockClient()
	log.Println("HR DB Mock client initialized")

	router := httpapi.NewRouter(httpapi.RouterConfig{
		WebhookSecret:       cfg.GiteaWebhookSecret,
		EventStore:          eventStore,
		EventProcessor:      eventProcessor,
		HealthStore:         healthStore,
		DomainStore:         domainStore,
		CommandStore:        commandStore,
		AuditStore:          auditStore,
		OrganizationStore:   organizationStore,
		RBACStore:           rbacStore,
		BearerTokenVerifier: verifier,
		KratosLogin:         kratosLogin,
		HydraAdmin:          hydraAdmin,
		HydraToken:          hydraToken,
		KratosAdmin:         kratosAdmin,
		HRDB:                hrdbMock,
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
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("run server: %v", err)
	}
}
