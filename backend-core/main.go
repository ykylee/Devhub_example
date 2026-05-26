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
	"github.com/devhub/backend-core/internal/gitea"
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
	var pgStore *store.PostgresStore

	if cfg.DBURL != "" {
		var err error
		pgStore, err = store.NewPostgresStore(ctx, cfg.DBURL)
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
	jwksVerifier := &auth.KeycloakJWKSVerifier{
		IssuerURL: cfg.OIDCIssuerURL,
		JWKSURL:   cfg.OIDCJWKSURL,
		ClientID:  cfg.OIDCClientID,
	}
	// ADR-0020 sub-carve D (sprint -l, issue #213) — stale-while-error fallback
	// 의 MaxStaleDuration env wire. 빈 값 또는 invalid 면 keycloak_verifier
	// 내부 default (24h) 적용. log 로 명시 가시화.
	if raw := strings.TrimSpace(cfg.OIDCJWKSMaxStaleDuration); raw != "" {
		if parsed, err := time.ParseDuration(raw); err == nil && parsed > 0 {
			jwksVerifier.MaxStaleDuration = parsed
			log.Printf("jwks stale-while-error max_stale_duration = %s", parsed)
		} else {
			log.Printf("invalid DEVHUB_OIDC_JWKS_MAX_STALE_DURATION=%q; fallback to default (24h)", raw)
		}
	}
	verifier = jwksVerifier
	log.Printf("bearer token verifier: keycloak jwks (issuer=%q jwks=%q client_id=%q)", cfg.OIDCIssuerURL, cfg.OIDCJWKSURL, cfg.OIDCClientID)
	if err := cfg.Validate(verifier != nil); err != nil {
		log.Fatalf("startup refused: %v", err)
	}

	var (
		idpAdmin httpapi.IdentityAdmin
	)
	if cfg.KeycloakAdminURL != "" && cfg.KeycloakAdminRealm != "" && cfg.KeycloakAdminClientID != "" && cfg.KeycloakAdminClientSecret != "" {
		idpAdmin = &httpapi.KeycloakAdminClient{
			AdminURL:     cfg.KeycloakAdminURL,
			Realm:        cfg.KeycloakAdminRealm,
			ClientID:     cfg.KeycloakAdminClientID,
			ClientSecret: cfg.KeycloakAdminClientSecret,
		}
		log.Printf("identity admin client: keycloak (admin_url=%q realm=%q client_id=%q)", cfg.KeycloakAdminURL, cfg.KeycloakAdminRealm, cfg.KeycloakAdminClientID)
	} else {
		log.Println("keycloak provider mode: account admin adapter is not fully configured")
	}

	// ADR-0020 sub-carve E (sprint -n) — seedLocalAdmin Keycloak `CreateIdentity`
	// 호출 정공법 제거. dev 운영자가 Keycloak admin console 또는 realm-export.json
	// 으로 `test` user 1회 시드. DevHub `users` row 는 infra/idp/sql/003_seed_test_admin.sql
	// (idempotent) 가 담당. 2026-05-21 lazy 폐기 sprint (issue #284) 이후 첫
	// 로그인 시 authenticateActor 의 SetIdPSubject 가 idp_subject 매핑 (DB row
	// 존재 시) 또는 token-only actor 로 처리 (DB miss 시 → onboarding 화면).

	hrdbMock := hrdb.NewMockClient()
	log.Println("HR DB Mock client initialized")

	router := httpapi.NewRouter(httpapi.RouterConfig{
		WebhookSecret:              cfg.GiteaWebhookSecret,
		KeycloakWebhookSecret:      cfg.KeycloakWebhookSecret,
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
		IdentityAdmin:              idpAdmin,
		IdPProvider:                cfg.IdPProvider,
		HRDB:                       hrdbMock,
		SnapshotProvider: httpapi.RuntimeSnapshotProvider{
			Base:         httpapi.StaticSnapshotProvider{},
			HealthStore:  healthStore,
			GiteaURL:     cfg.GiteaURL,
			BackendAIURL: cfg.BackendAIURL,
		},
		RealtimeHub:           realtimeHub,
		RealtimeTickets:       httpapi.NewRealtimeTicketStore(),
		AuthDevFallback:       cfg.AuthDevFallback,
		OnboardingGateEnabled: cfg.OnboardingGateEnabled,
		ProjectModel:          cfg.ProjectModel,
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
	if cfg.KeycloakEventListenerEnabled && idpAdmin != nil && auditStore != nil && eventCursorStore != nil {
		kc, ok := idpAdmin.(*httpapi.KeycloakAdminClient)
		if !ok {
			log.Printf("keycloak event listener skipped: identity admin is not *httpapi.KeycloakAdminClient (got %T) — provider mode mismatch or test fake", idpAdmin)
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
			// ADR-0020 sub-carve C (sprint -k, issue #212) — user_sync dispatcher
			// callback. admin event 처리 시 DevHub `users` 컬럼 자동 sync.
			// orgStore nil 인 경우 sync 생략 (이전 sprint -u~-y 동작 동등).
			var userSync audit.UserSyncCallback
			if organizationStore != nil {
				userSync = func(syncCtx context.Context, action audit.SyncUserAction, identityID, _ string) {
					start := time.Now()
					var err error
					switch action {
					case audit.SyncActionProfile:
						err = audit.SyncUserProfile(syncCtx, kc, organizationStore, identityID)
					case audit.SyncActionMembership:
						err = audit.SyncUserMembership(syncCtx, kc, organizationStore, identityID)
					case audit.SyncActionStatus:
						// USER:DELETE — Keycloak user 가 이미 gone. caller 가 username hint
						// 없이 호출하므로 GetUserDetails 가 404. MarkUserDeactivated 는
						// userID 가 빈 문자열이면 noop 반환. 이 경로는 PR #212 후속의
						// audit_logs.actor_login 캐시 lookup 으로 보강 예정 (carve out).
						err = audit.MarkUserDeactivated(syncCtx, organizationStore, identityID)
					}
					if err != nil {
						log.Printf("user_sync %s identity=%s failed: %v", action, identityID, err)
						audit.ObserveUserSyncError(action)
						return
					}
					audit.ObserveUserSync(action)
					audit.ObserveUserSyncLag(time.Since(start).Seconds())
				}
			}
			opts := audit.KeycloakEventPullerOptions{
				Interval:     interval,
				MaxEvents:    maxEvents,
				AuditEmitter: emitter,
				UserSync:     userSync,
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

	// Onboarding pending_review count Gauge cron refresh (SOP §8 P3 carve).
	// audit puller 패턴 정합 — single goroutine + ctx cancel. organizationStore 가
	// PostgresStore 인 경우 CountPendingReview 메서드 자동 구현 (interface 어설션).
	if counter, ok := organizationStore.(httpapi.OnboardingPendingReviewCounter); ok {
		go func() {
			err := httpapi.RunOnboardingPendingReviewGauge(ctx, counter, httpapi.OnboardingPendingGaugeOptions{})
			if err != nil && err != context.Canceled {
				log.Printf("onboarding pending_review gauge stopped: %v", err)
			}
		}()
		log.Printf("onboarding pending_review gauge enabled (interval=60s)")
	}

	// Gitea background sync worker
	if cfg.GiteaURL != "" && cfg.GiteaToken != "" && pgStore != nil {
		giteaWorker := gitea.NewSyncWorker(pgStore, cfg.GiteaURL, cfg.GiteaToken)
		go func() {
			if err := giteaWorker.Run(ctx, 30*time.Second); err != nil && err != context.Canceled {
				log.Printf("gitea sync worker stopped: %v", err)
			}
		}()
		log.Printf("gitea sync worker enabled (url=%s interval=30s)", cfg.GiteaURL)
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
// buildIntakeTokenAuditEmitter 패턴 정합. sourceEventID 는 puller 의 SHA256 dedup
// hash — store layer 의 partial UNIQUE INDEX (migration 000032) 가 backend crash +
// cursor revert 같은 edge 에서도 중복 INSERT 차단.
func buildKeycloakEventAuditEmitter(auditStore httpapi.AuditStore) audit.AuditEmitter {
	if auditStore == nil {
		return nil
	}
	return func(ctx context.Context, action, targetType, targetID, sourceEventID string, payload map[string]any) {
		actorLogin := "system:keycloak-event"
		if uid, ok := payload["user_id"].(string); ok && uid != "" {
			actorLogin = uid
		} else if aid, ok := payload["auth_user_id"].(string); ok && aid != "" {
			actorLogin = aid
		}
		entry := domain.AuditLog{
			ActorLogin:    actorLogin,
			Action:        action,
			TargetType:    targetType,
			TargetID:      targetID,
			SourceType:    domain.AuditSourceKeycloakEvent,
			SourceEventID: sourceEventID,
			Payload:       payload,
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
