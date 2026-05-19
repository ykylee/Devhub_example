package main

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/devhub/backend-core/internal/audit"
	"github.com/devhub/backend-core/internal/auth"
	"github.com/devhub/backend-core/internal/commandworker"
	"github.com/devhub/backend-core/internal/config"
	"github.com/devhub/backend-core/internal/devrequest"
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
	var eventCursorStore store.EventCursorStore

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
		eventCursorStore = pgStore

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
	verifier = &auth.KeycloakJWKSVerifier{
		IssuerURL: cfg.OIDCIssuerURL,
		JWKSURL:   cfg.OIDCJWKSURL,
		ClientID:  cfg.OIDCClientID,
	}
	log.Printf("bearer token verifier: keycloak jwks (issuer=%q jwks=%q client_id=%q)", cfg.OIDCIssuerURL, cfg.OIDCJWKSURL, cfg.OIDCClientID)
	if err := cfg.Validate(verifier != nil); err != nil {
		log.Fatalf("startup refused: %v", err)
	}

	var (
		kratosAdmin httpapi.IdentityAdmin
	)
	if cfg.KeycloakAdminURL != "" && cfg.KeycloakAdminRealm != "" && cfg.KeycloakAdminClientID != "" && cfg.KeycloakAdminClientSecret != "" {
		kratosAdmin = &httpapi.KeycloakAdminClient{
			AdminURL:     cfg.KeycloakAdminURL,
			Realm:        cfg.KeycloakAdminRealm,
			ClientID:     cfg.KeycloakAdminClientID,
			ClientSecret: cfg.KeycloakAdminClientSecret,
			IssuerURL:    cfg.OIDCIssuerURL,
		}
		log.Printf("identity admin client: keycloak (admin_url=%q realm=%q client_id=%q)", cfg.KeycloakAdminURL, cfg.KeycloakAdminRealm, cfg.KeycloakAdminClientID)
	} else {
		log.Println("keycloak provider mode: account admin adapter is not fully configured")
	}

	// Seed local admin account for development using regular APIs
	if cfg.AuthDevFallback && kratosAdmin != nil && organizationStore != nil {
		seedLocalAdmin(ctx, kratosAdmin, organizationStore)
	}

	hrdbMock := hrdb.NewMockClient()
	log.Println("HR DB Mock client initialized")

	router := httpapi.NewRouter(httpapi.RouterConfig{
		WebhookSecret:              cfg.GiteaWebhookSecret,
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
		IdentityAdmin:              kratosAdmin,
		IdPProvider:                cfg.IdPProvider,
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

	// DREQ intake token cron loop — ADR-0017 §6 (a)+(c)+(d), sprint claude/work_260518-t.
	// 만료 token 자동 hard-revoke + expiring_soon/stale Prometheus gauge emit.
	// store 와 audit emitter 가 모두 준비된 경우만 활성화 — config gate 가 false 거나
	// store 가 nil 이면 skip (no-op).
	if cfg.DREQTokenCronEnabled && devRequestIntakeTokenStore != nil {
		if cronStore, ok := devRequestIntakeTokenStore.(devrequest.IntakeTokenStore); ok {
			interval := 10 * time.Minute
			if strings.TrimSpace(cfg.DREQTokenCronInterval) != "" {
				if parsed, err := time.ParseDuration(cfg.DREQTokenCronInterval); err == nil && parsed > 0 {
					interval = parsed
				} else {
					log.Printf("invalid DEVHUB_DREQ_TOKEN_CRON_INTERVAL=%q; fallback to %s", cfg.DREQTokenCronInterval, interval)
				}
			}
			expiringThreshold := 24 * time.Hour
			if strings.TrimSpace(cfg.DREQTokenExpiringSoonThreshold) != "" {
				if parsed, err := time.ParseDuration(cfg.DREQTokenExpiringSoonThreshold); err == nil && parsed > 0 {
					expiringThreshold = parsed
				}
			}
			staleThreshold := 720 * time.Hour // 30d default
			if strings.TrimSpace(cfg.DREQTokenStaleThreshold) != "" {
				if parsed, err := time.ParseDuration(cfg.DREQTokenStaleThreshold); err == nil {
					// 0 또는 음수 = disable (운영자가 stale 알림 미사용 결정 명시).
					staleThreshold = parsed
				}
			}
			emitter := buildIntakeTokenAuditEmitter(auditStore)
			opts := devrequest.IntakeTokenCronOptions{
				Interval:              interval,
				ExpiringSoonThreshold: expiringThreshold,
				StaleThreshold:        staleThreshold,
				AuditEmitter:          emitter,
			}
			go func() {
				err := devrequest.RunIntakeTokenCron(ctx, cronStore, opts)
				if err != nil && err != context.Canceled {
					log.Printf("dreq intake token cron stopped: %v", err)
				}
			}()
			log.Printf("dreq intake token cron enabled (interval=%s expiring=%s stale=%s)", interval, expiringThreshold, staleThreshold)
		}
	}

	// Keycloak event listener cron — ADR-0019 §5.3 (9), sprint claude/work_260519-v PR-C.
	// KeycloakAdminClient 로 /admin/realms/{realm}/events + /admin-events polling,
	// audit_logs 로 emit (source_type=keycloak_event). 모든 wire 의존이 모두 준비된
	// 경우에만 활성화 — config gate / KeycloakAdminClient / auditStore / eventCursorStore.
	if cfg.KeycloakEventListenerEnabled && kratosAdmin != nil && auditStore != nil && eventCursorStore != nil {
		kc, ok := kratosAdmin.(*httpapi.KeycloakAdminClient)
		if !ok {
			log.Printf("keycloak event listener skipped: identity admin is not KeycloakAdminClient (provider mode mismatch)")
		} else {
			interval := 30 * time.Second
			if strings.TrimSpace(cfg.KeycloakEventListenerInterval) != "" {
				if parsed, err := time.ParseDuration(cfg.KeycloakEventListenerInterval); err == nil && parsed > 0 {
					interval = parsed
				} else {
					log.Printf("invalid DEVHUB_KEYCLOAK_EVENT_LISTENER_INTERVAL=%q; fallback to %s", cfg.KeycloakEventListenerInterval, interval)
				}
			}
			maxEvents := cfg.KeycloakEventListenerMaxEvents
			if maxEvents <= 0 {
				maxEvents = 500
			}
			lister := audit.NewHTTPAPIEventListerAdapter(&keycloakAdminEventLister{kc: kc})
			emitter := buildKeycloakEventAuditEmitter(auditStore)
			opts := audit.KeycloakEventPullerOptions{
				Interval:     interval,
				MaxEvents:    maxEvents,
				AuditEmitter: emitter,
			}
			go func() {
				err := audit.RunKeycloakEventPuller(ctx, lister, eventCursorStore, opts)
				if err != nil && err != context.Canceled {
					log.Printf("keycloak event listener stopped: %v", err)
				}
			}()
			log.Printf("keycloak event listener enabled (interval=%s max_events=%d)", interval, maxEvents)
		}
	}

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("run server: %v", err)
	}
}

// keycloakAdminEventLister — httpapi.KeycloakAdminClient 를 audit.HTTPAPIEventLister
// 로 변환하는 thin adapter. audit ← httpapi 순방향 의존만 유지하기 위해 본 wrapper 가
// main.go 측에 존재. struct 필드는 동일하므로 named-type 변환만 수행.
type keycloakAdminEventLister struct {
	kc *httpapi.KeycloakAdminClient
}

func (a *keycloakAdminEventLister) ListUserEvents(ctx context.Context, dateFrom time.Time, max int) ([]audit.HTTPAPIUserEvent, error) {
	src, err := a.kc.ListUserEvents(ctx, dateFrom, max)
	if err != nil {
		return nil, err
	}
	out := make([]audit.HTTPAPIUserEvent, len(src))
	for i, ev := range src {
		out[i] = audit.HTTPAPIUserEvent{
			Time:     ev.Time,
			Type:     ev.Type,
			RealmID:  ev.RealmID,
			ClientID: ev.ClientID,
			UserID:   ev.UserID,
			IPAddr:   ev.IPAddr,
			Details:  ev.Details,
			Error:    ev.Error,
		}
	}
	return out, nil
}

func (a *keycloakAdminEventLister) ListAdminEvents(ctx context.Context, dateFrom time.Time, max int) ([]audit.HTTPAPIAdminEvent, error) {
	src, err := a.kc.ListAdminEvents(ctx, dateFrom, max)
	if err != nil {
		return nil, err
	}
	out := make([]audit.HTTPAPIAdminEvent, len(src))
	for i, ev := range src {
		flat := audit.HTTPAPIAdminEvent{
			Time:          ev.Time,
			RealmID:       ev.RealmID,
			OperationType: ev.OperationType,
			ResourceType:  ev.ResourceType,
			ResourcePath:  ev.ResourcePath,
			Error:         ev.Error,
		}
		if ev.AuthDetails != nil {
			flat.AuthUserID = ev.AuthDetails.UserID
			flat.AuthClientID = ev.AuthDetails.ClientID
			flat.AuthIPAddr = ev.AuthDetails.IPAddr
		}
		out[i] = flat
	}
	return out, nil
}

// buildKeycloakEventAuditEmitter — Keycloak event listener 의 best-effort audit emit
// 콜백. auditStore 가 nil 이면 nil 반환 (cron 이 audit 생략). DREQ intake token cron 의
// buildIntakeTokenAuditEmitter 패턴 정합.
func buildKeycloakEventAuditEmitter(auditStore httpapi.AuditStore) audit.AuditEmitter {
	if auditStore == nil {
		return nil
	}
	return func(ctx context.Context, action, targetType, targetID string, payload map[string]any) {
		actorLogin := "system:keycloak-event"
		if uid, ok := payload["user_id"].(string); ok && uid != "" {
			actorLogin = uid
		} else if aid, ok := payload["auth_user_id"].(string); ok && aid != "" {
			actorLogin = aid
		}
		entry := domain.AuditLog{
			ActorLogin: actorLogin,
			Action:     action,
			TargetType: targetType,
			TargetID:   targetID,
			SourceType: domain.AuditSourceKeycloakEvent,
			Payload:    payload,
		}
		// best-effort — cron 자체는 audit 실패 시에도 계속 (HomeLab pull loop 패턴).
		_, _ = auditStore.CreateAuditLog(ctx, entry)
	}
}

// buildIntakeTokenAuditEmitter — DREQ intake token cron 의 best-effort audit emit
// 콜백. auditStore 가 nil 이면 nil 반환 (cron 이 audit 생략). sprint -t.
func buildIntakeTokenAuditEmitter(auditStore httpapi.AuditStore) devrequest.AuditEmitter {
	if auditStore == nil {
		return nil
	}
	return func(ctx context.Context, action, targetType, targetID string, payload map[string]any) {
		entry := domain.AuditLog{
			ActorLogin: "system:dreq-cron",
			Action:     action,
			TargetType: targetType,
			TargetID:   targetID,
			SourceType: domain.AuditSourceSystem,
			Payload:    payload,
		}
		// best-effort — cron 자체는 audit 실패 시에도 계속 (HomeLab pull loop 패턴).
		_, _ = auditStore.CreateAuditLog(ctx, entry)
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
	SetIdPSubject(context.Context, string, string) error
}

func seedLocalAdmin(ctx context.Context, kratosAdmin httpapi.IdentityAdmin, orgStore seedOrgStore) {
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
	err = orgStore.SetIdPSubject(ctx, adminLogin, kratosID)
	if err != nil {
		log.Printf("[seedLocalAdmin] Failed to link Kratos ID: %v", err)
	} else {
		log.Printf("[seedLocalAdmin] Successfully ensured test admin '%s' via regular APIs", adminLogin)
	}
}
