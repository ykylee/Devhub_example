package main

import (
	"context"
	"log"
	"net/url"
	"time"

	"github.com/devhub/backend-core/internal/auth"
	"github.com/devhub/backend-core/internal/commandworker"
	"github.com/devhub/backend-core/internal/config"
	"github.com/devhub/backend-core/internal/domain"
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
		WebhookSecret:       cfg.GiteaWebhookSecret,
		KratosWebhookToken:  cfg.KratosWebhookToken,
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
		HydraLogout:         hydraLogout,
		HydraToken:          hydraToken,
		HydraRevoker:        hydraRevoker,
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
