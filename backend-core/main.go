package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	apprep "github.com/devhub/backend-core/internal/domain/application-lifecycle/repository"
	auditrep "github.com/devhub/backend-core/internal/domain/audit-ops/repository"
	auditsvc "github.com/devhub/backend-core/internal/domain/audit-ops/service"
	"github.com/devhub/backend-core/internal/domain/auth-session/integration"
	authapikeyrep "github.com/devhub/backend-core/internal/domain/auth-session/repository"
	authview "github.com/devhub/backend-core/internal/domain/auth-session/view"
	devreqrep "github.com/devhub/backend-core/internal/domain/dev-request/repository"
	devreqsvc "github.com/devhub/backend-core/internal/domain/dev-request/service"
	devreqview "github.com/devhub/backend-core/internal/domain/dev-request/view"
	intgregrep "github.com/devhub/backend-core/internal/domain/integration-registry/repository"
	onboardview "github.com/devhub/backend-core/internal/domain/onboarding/view"
	orgrep "github.com/devhub/backend-core/internal/domain/organization-management/repository"
	rbacrep "github.com/devhub/backend-core/internal/domain/rbac-permissions/repository"
	realtimeview "github.com/devhub/backend-core/internal/domain/realtime/view"
	notifrep "github.com/devhub/backend-core/internal/domain/user-notification/repository"
	"github.com/devhub/backend-core/internal/httpapi"
	"github.com/devhub/backend-core/internal/infrastructure/commandworker"
	"github.com/devhub/backend-core/internal/infrastructure/gitea"
	"github.com/devhub/backend-core/internal/infrastructure/hrdb"
	"github.com/devhub/backend-core/internal/infrastructure/serviceaction"
	"github.com/devhub/backend-core/internal/integrations/adapters"
	"github.com/devhub/backend-core/internal/normalize"
	"github.com/devhub/backend-core/internal/shared/config"
	keycloakadapter "github.com/devhub/backend-core/internal/sso-integrations/keycloak"
	"github.com/devhub/backend-core/internal/store"
)

// apiKeyViewStoreAdapter — repository.APIKeyStore (repository-local struct
// 반환) 를 view.APIKeyStore (view-local APIKeyView 반환) 로 wrap. cross-domain
// import cycle 회피 + struct 정의 1쌍 (repository + view) 의 단일 adapter.
// auth.go 의 APIKeyStore 가 APIKeyView 만 보면 되므로 변환 1회로 충분.
type apiKeyViewStoreAdapter struct {
	inner authapikeyrep.APIKeyStore
}

func (a *apiKeyViewStoreAdapter) GetAPIKeyByHash(ctx context.Context, keyHash []byte) (authview.APIKeyView, error) {
	k, err := a.inner.GetAPIKeyByHash(ctx, keyHash)
	if err != nil {
		return authview.APIKeyView{}, err
	}
	return authview.APIKeyView{
		ID:           k.ID,
		Name:         k.Name,
		KeyPrefix:    k.KeyPrefix,
		CreatedBy:    k.CreatedBy,
		ExpiresAt:    k.ExpiresAt,
		AllowedCIDRs: k.AllowedCIDRs,
	}, nil
}

func (a *apiKeyViewStoreAdapter) UpdateLastUsedAt(ctx context.Context, id string, when time.Time) error {
	return a.inner.UpdateLastUsedAt(ctx, id, when)
}

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
	var platformStore httpapi.PlatformStore
	// integrationStore — issue #421/#422 (sprint claude/work_260529-n).
	// IntegrationRepository 를 ApplicationRepository embed 에서 분리 후
	// 별도 inject 해 cross-domain decouple.
	var integrationStore httpapi.IntegrationStore
	var devRequestStore httpapi.DevRequestStore
	var devRequestIntakeTokenStore httpapi.IntakeTokenStore
	// ADR-0028: voc + notification
	var vocStore devreqview.VocStore
	var notificationStore devreqview.NotificationStore
	var rbacStore httpapi.RBACStore
	realtimeHub := realtimeview.NewRealtimeHub()
	var worker *commandworker.Worker
	var liveWorker *commandworker.LiveWorker
	var homeLabAdapterStore adapters.InfraSnapshotStore
	var eventCursorStore auditrep.EventCursorStore
	var pgStore *store.PostgresStore
	// apiKeyStore — ADR-0029 §6 (f) P3 multi-key. nil 이면 admin endpoints 503.
	var apiKeyStore authapikeyrep.APIKeyStore
	var apiKeyStoreAdmin authapikeyrep.APIKeyStore
	var apiKeyViewStore authview.APIKeyStore

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
		auditStore = auditrep.NewAuditRepository(pgStore)
		organizationStore = orgrep.NewOrganizationRepository(pgStore)
		platformStore = apprep.NewPlatformRepository(pgStore)
		// ADR-0029 §6 (f) P3 — multi-key 관리 store. auth middleware (read-only
		// hot path) + admin handler (write) 양쪽 동일 store 사용. pgStore.Pool()
		// 가 nil 아닌 경우에만 wire — cfg.DBURL 이 빈 경우 (sqlite/in-memory
		// 환경) nil 로 두고 AuthConfig.APIKeyStore/Admin 둘 다 nil → admin
		// endpoints 가 503 응답.
		apiKeyStore = authapikeyrep.NewPgAPIKeyStore(pgStore.Pool())
		apiKeyStoreAdmin = apiKeyStore
		// view-local APIKeyStore (read-only) adapter — view 가 repository
		// package 의존하지 않도록 cross-domain import cycle 회피. view 의
		// auth.go 가 APIKeyView 를 받고, admin handler 가 repository.APIKey
		// 를 받음 — 두 adapter 가 동일 underlying store 를 wrap.
		apiKeyViewStore = &apiKeyViewStoreAdapter{inner: apiKeyStore}
		integrationStore = intgregrep.NewIntegrationRepository(pgStore)
		devRequestRepository := devreqrep.NewDevRequestRepository(pgStore)
		devRequestStore = devRequestRepository
		devRequestIntakeTokenStore = devRequestRepository
		// ADR-0028: voc + notification repository (sprint work_260612-a)
		vocStore = devreqrep.NewDevRequestVocRepository(pgStore)
		notificationStore = notifrep.NewUserNotificationRepository(pgStore)
		rbacStore = rbacrep.NewRBACRepository(pgStore)
		homeLabAdapterStore = pgStore
		eventCursorStore = auditrep.NewAuditRepository(pgStore)

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
	jwksVerifier := &keycloakadapter.KeycloakJWKSVerifier{
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
		idpAdmin          httpapi.IdentityAdmin
		oidcLogout        httpapi.OIDCLogoutClient
		keycloakEventPort integration.KeycloakEventPort
	)
	// v1.1 sprint -a follow-up (ADR-0030 §2.3) — runtime injection via
	// DEVHUB_BUILD_TIER. default = saovae_stub (외부 환경, Keycloak 인프라
	// 의존성 0). `internal` 시 real KeycloakAdminClient (현 path) 사용 —
	// 사내 staging/prod-smoke 검증. keycloakAdminClient 가 3 개 port
	// (IdentityAdmin + OIDCLogoutClient + KeycloakEventPort via
	// ListUserEvents/ListAdminEvents) 모두 충족하므로 단일 instance 가
	// 3 슬롯에 동시 주입.
	//
	// v1.1 sprint -a follow-up PR1 (real adapter 이전): KeycloakAdminClient 가
	// sso-integrations/keycloak/ 으로 이동. canonical struct (integration.*) 가
	// wire 시 type assertion 없이 port 를 직접 충족하므로 별도 adapter 불요.
	if strings.EqualFold(os.Getenv("DEVHUB_BUILD_TIER"), "internal") {
		// 사내 build: real KeycloakAdminClient. sso-integrations/keycloak/ 패키지의
		// real adapter instance (sprint -a follow-up PR1 에서 이전).
		if cfg.KeycloakAdminURL != "" && cfg.KeycloakAdminRealm != "" && cfg.KeycloakAdminClientID != "" && cfg.KeycloakAdminClientSecret != "" {
			kc := &keycloakadapter.KeycloakAdminClient{
				AdminURL:     cfg.KeycloakAdminURL,
				Realm:        cfg.KeycloakAdminRealm,
				ClientID:     cfg.KeycloakAdminClientID,
				ClientSecret: cfg.KeycloakAdminClientSecret,
				// IssuerURL 은 OIDC logout endpoint URL 결정 전용
				// (oidcLogoutEndpoint 만 사용). tokenEndpoint() (admin service-
				// account token) 는 절대 IssuerURL 을 보지 않음 — admin endpoint 와
				// user-facing OIDC endpoint 가 deployment 별로 다른 host 일 수 있음
				// (DEVHUB_KEYCLOAK_ADMIN_URL = internal docker vs
				// DEVHUB_OIDC_ISSUER_URL = public ingress). codex P1 review #3 정합.
				IssuerURL:        cfg.OIDCIssuerURL,
				OIDCClientID:     cfg.OIDCClientID,
				OIDCClientSecret: cfg.OIDCClientSecret,
				// OIDC logout 은 token 발급 client (frontend) 자격증명 사용
				// (RFC 6749 §4.1.3 / Keycloak token binding). admin client 와
				// 분리 — codex P1 review #2 정합 (sprint -i fix).
			}
			idpAdmin = kc
			// KeycloakAdminClient 가 OIDCLogoutClient 인터페이스도 충족하므로
			// 동일 인스턴스를 두 슬롯에 주입. (codex P1 review #1 정합)
			oidcLogout = kc
			// KeycloakEventPort (ListUserEvents/ListAdminEvents) — KeycloakAdminClient
			// 가 본 port 도 충족. 단일 instance 가 3 슬롯에 동시 주입. type assertion
			// 불요 (canonical struct 정합).
			keycloakEventPort = kc
			log.Printf("identity admin + OIDC logout + keycloak event port: real keycloak (admin_client=%q oidc_client=%q realm=%q)",
				cfg.KeycloakAdminClientID, cfg.OIDCClientID, cfg.KeycloakAdminRealm)
		} else {
			log.Println("keycloak provider mode (internal tier): account admin adapter is not fully configured")
		}
	} else {
		// 사외 build (default): saovae_stub. Keycloak 인프라 의존성 0.
		idpAdmin = keycloakadapter.NewIdentityAdminStub()
		oidcLogout = keycloakadapter.NewOIDCLogoutClientStub()
		keycloakEventPort = keycloakadapter.NewKeycloakEventPortStub()
		log.Println("identity admin + OIDC logout + keycloak event port: saovae_stub (DEVHUB_BUILD_TIER not set or != internal)")
	}

	// ADR-0020 sub-carve E (sprint -n) — seedLocalAdmin Keycloak `CreateIdentity`
	// 호출 정공법 제거. dev 운영자가 Keycloak admin console 또는 realm-export.json
	// 으로 `test` user 1회 시드. DevHub `users` row 는 infra/idp/sql/003_seed_test_admin.sql
	// (idempotent) 가 담당. 2026-05-21 lazy 폐기 sprint (issue #284) 이후 첫
	// 로그인 시 authenticateActor 의 SetIdPSubject 가 idp_subject 매핑 (DB row
	// 존재 시) 또는 token-only actor 로 처리 (DB miss 시 → onboarding 화면).

	hrdbMock := hrdb.NewMockClient()
	log.Println("HR DB Mock client initialized")

	if cfg.AuthDevFallback {
		log.Println("[WARNING] DEVHUB_AUTH_DEV_FALLBACK is enabled. Development-only authentication fallbacks are ACTIVE.")
	}

	// Swagger UI 1차 bootstrap (ADR-0027) — opt-in via DEVHUB_SWAGGER_ENABLED.
	// OpenAPISpecPath 는 명시적 절대경로 env (예: /etc/devhub/openapi.yaml) 가
	// 설정되면 disk 파일을 서빙하고, 미설정 시 embed.FS 의 swaggerui/asset/openapi.yaml
	// (docs/openapi.yaml 의 build-time copy) 로 fallback. codex P2 fix (PR #508):
	// cwd-relative default 가 운영 cwd 가 repo root 가 아닐 때 silent 미존재하는
	// 함정 회피 — env 미설정 = 명시적 disk 경로 없음, embed fallback 으로 전환.
	swaggerSpecPath := strings.TrimSpace(os.Getenv("DEVHUB_OPENAPI_SPEC_PATH"))

	router := httpapi.NewRouter(httpapi.RouterConfig{
		SwaggerEnabled: cfg.SwaggerEnabled,
		// ADR-0029 §6 (e) P2 — 운영 환경 default = true (swagger UI 자체에
		// system_admin gate). 로컬 dev / e2e test 는 false 로 명시적 override
		// 가능 (env DEVHUB_SWAGGER_REQUIRE_SYSTEM_ADMIN).
		SwaggerRequireSystemAdmin:  cfg.SwaggerRequireSystemAdmin,
		OpenAPISpecPath:            swaggerSpecPath,
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
		PlatformStore:              platformStore,
		IntegrationStore:           integrationStore,
		DevRequestStore:            devRequestStore,
		DevRequestIntakeTokenStore: devRequestIntakeTokenStore,
		// ADR-0028: voc + notification
		VocStore:            vocStore,
		NotificationStore:   notificationStore,
		RBACStore:           rbacStore,
		BearerTokenVerifier: verifier,
		APIKey:              cfg.APIKey,
		APIKeyAdminOnly:     cfg.APIKeyAdminOnly,
		// ADR-0029 §6 (f) P3 — multi-key store wire. nil 이면 multi-key 비활성.
		APIKeyStore:      apiKeyViewStore,
		APIKeyStoreAdmin: apiKeyStoreAdmin,
		IdentityAdmin:    idpAdmin,
		OIDCLogoutClient: oidcLogout,
		IdPProvider:      cfg.IdPProvider,
		HRDB:             hrdbMock,
		SnapshotProvider: httpapi.RuntimeSnapshotProvider{
			Base:         httpapi.StaticSnapshotProvider{},
			HealthStore:  healthStore,
			GiteaURL:     cfg.GiteaURL,
			BackendAIURL: cfg.BackendAIURL,
		},
		RealtimeHub:     realtimeHub,
		RealtimeTickets: realtimeview.NewRealtimeTicketStoreFor(pgStore),
		// codex P1 (#390) — task item ingestion 의 PostgresExternalTaskStore wire.
		// pgStore 는 위 cfg.DBURL gate 에서 fatal 처리되므로 여기서는 non-nil 보장.
		ExternalTaskStore:     intgregrep.NewPostgresExternalTaskStoreFor(pgStore),
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

	// X-5 Gitea Hourly Pull 정밀화 (ADR-0034, sprint feat/x5-gitea-hourly-pull).
	// opt-in via DEVHUB_GITEA_PULL_ENABLED. cycle interval default 1h.
	// Gitea API → DevHub pr_activities / build_runs / quality_snapshots since-based sync.
	if cfg.GiteaPullEnabled {
		giteaBase := strings.TrimSpace(cfg.GiteaAPIBaseURL)
		if giteaBase == "" {
			log.Printf("gitea pull loop skipped: set DEVHUB_GITEA_API_BASE_URL (sprint feat/x5-gitea-hourly-pull)")
		} else {
			giteaInterval := 1 * time.Hour
			if strings.TrimSpace(cfg.GiteaPullInterval) != "" {
				if parsed, err := time.ParseDuration(cfg.GiteaPullInterval); err == nil && parsed > 0 {
					giteaInterval = parsed
				} else {
					log.Printf("invalid DEVHUB_GITEA_PULL_INTERVAL=%q; fallback to %s", cfg.GiteaPullInterval, giteaInterval)
				}
			}
			giteaCycleTimeout := 30 * time.Minute
			if strings.TrimSpace(cfg.GiteaPullCycleTimeout) != "" {
				if parsed, err := time.ParseDuration(cfg.GiteaPullCycleTimeout); err == nil && parsed > 0 {
					giteaCycleTimeout = parsed
				} else {
					log.Printf("invalid DEVHUB_GITEA_PULL_CYCLE_TIMEOUT=%q; fallback to %s", cfg.GiteaPullCycleTimeout, giteaCycleTimeout)
				}
			}
			giteaConcurrency := cfg.GiteaPullConcurrency
			if giteaConcurrency <= 0 {
				giteaConcurrency = 4
			}
			giteaBackoffCap := 24 * time.Hour
			if strings.TrimSpace(cfg.GiteaPullBackoffCap) != "" {
				if parsed, err := time.ParseDuration(cfg.GiteaPullBackoffCap); err == nil && parsed > 0 {
					giteaBackoffCap = parsed
				} else {
					log.Printf("invalid DEVHUB_GITEA_PULL_BACKOFF_CAP=%q; fallback to %s", cfg.GiteaPullBackoffCap, giteaBackoffCap)
				}
			}
			giteaFailureAlertThreshold := cfg.GiteaPullFailureAlertThreshold
			if giteaFailureAlertThreshold <= 0 {
				giteaFailureAlertThreshold = 5
			}

			giteaClient := adapters.NewGiteaClient(giteaBase, cfg.GiteaAPIToken)
			// NOTE: the production RepositoryPullStore wiring is provided by a follow-up PR.
			// In this PR we wire only the loop with a nil store; the loop will fail-fast on cycle
			// and emit error audit. Operators are expected to provide a store implementation
			// alongside DEVHUB_GITEA_PULL_ENABLED=true.
			adapter := &adapters.GiteaPullAdapter{
				Client:         giteaClient,
				Store:          nil, // follow-up: wire to repository store
				MaxItemsPerCall: 200,
			}
			repoLister := func(ctx context.Context) ([]adapters.RepositoryTarget, error) {
				// follow-up: query repositories table where provider_type='gitea' and not in backoff.
				return nil, nil
			}
			onCycle := func(r adapters.PullCycleResult) {
				log.Printf("gitea pull cycle %s result=%s synced=%d errored=%d partial=%d skipped=%d",
					r.CycleID, r.OverallResult, r.RepositoriesSynced, r.RepositoriesErrored, r.RepositoriesPartial, r.RepositoriesSkipped)
			}
			onError := func(err error) {
				log.Printf("gitea pull loop error: %v", err)
			}
			go func() {
				err := adapters.RunGiteaPullLoop(ctx, adapter, repoLister, giteaInterval, giteaCycleTimeout, giteaConcurrency, giteaBackoffCap, giteaFailureAlertThreshold, onCycle, onError)
				if err != nil && err != context.Canceled {
					log.Printf("gitea pull loop stopped: %v", err)
				}
			}()
			log.Printf("gitea pull loop enabled (interval=%s cycle_timeout=%s concurrency=%d backoff_cap=%s alert_threshold=%d)",
				giteaInterval, giteaCycleTimeout, giteaConcurrency, giteaBackoffCap, giteaFailureAlertThreshold)
		}
	}

	// DREQ intake token cron loop — ADR-0017 §6 (a)+(c)+(d), sprint claude/work_260518-t.
	// 만료 token 자동 hard-revoke + expiring_soon/stale Prometheus gauge emit.
	// store 와 audit emitter 가 모두 준비된 경우만 활성화 — config gate 가 false 거나
	// store 가 nil 이면 skip (no-op).
	if cfg.DREQTokenCronEnabled && devRequestIntakeTokenStore != nil {
		if cronStore, ok := devRequestIntakeTokenStore.(devreqsvc.IntakeTokenStore); ok {
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
			opts := devreqsvc.IntakeTokenCronOptions{
				Interval:              interval,
				ExpiringSoonThreshold: expiringThreshold,
				StaleThreshold:        staleThreshold,
				AuditEmitter:          emitter,
			}
			go func() {
				err := devreqsvc.RunIntakeTokenCron(ctx, cronStore, opts)
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
	// 경우에만 활성화 — config gate / KeycloakEventPort / auditStore / eventCursorStore.
	//
	// v1.1 sprint -a follow-up PR1 (real adapter 이전): type assertion 제거.
	// integration.KeycloakEventPort (canonical port) 를 keycloakEventPort 가 직접
	// 충족하므로 idpAdmin 의 type assertion 불요. event listener 가 port 를 직접 받음.
	if cfg.KeycloakEventListenerEnabled && keycloakEventPort != nil && auditStore != nil && eventCursorStore != nil {
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
		emitter := buildKeycloakEventAuditEmitter(auditStore)
		opts := auditsvc.KeycloakEventPullerOptions{
			Interval:     interval,
			MaxEvents:    maxEvents,
			AuditEmitter: emitter,
		}
		go func() {
			err := auditsvc.RunKeycloakEventPuller(ctx, keycloakEventPort, eventCursorStore, opts)
			if err != nil && err != context.Canceled {
				log.Printf("keycloak event listener stopped: %v", err)
			}
		}()
		log.Printf("keycloak event listener enabled (interval=%s max_events=%d)", interval, maxEvents)
	}

	// Onboarding pending_review count Gauge cron refresh (SOP §8 P3 carve).
	// audit puller 패턴 정합 — single goroutine + ctx cancel. organizationStore 가
	// PostgresStore 인 경우 CountPendingReview 메서드 자동 구현 (interface 어설션).
	if counter, ok := organizationStore.(onboardview.OnboardingPendingReviewCounter); ok {
		go func() {
			err := onboardview.RunOnboardingPendingReviewGauge(ctx, counter, onboardview.OnboardingPendingGaugeOptions{})
			if err != nil && err != context.Canceled {
				log.Printf("onboarding pending_review gauge stopped: %v", err)
			}
		}()
		log.Printf("onboarding pending_review gauge enabled (interval=60s)")
	}

	// Gitea background sync worker. Phase 3: pgStore 가 있으면 항상 기동해 queued
	// sync job 을 provider 별 base_url+api_token 으로 처리 (env GITEA_URL/TOKEN 은
	// 큐 빈 주기 sync 의 fallback). env 미설정이어도 UI 로 등록한 Gitea provider 의
	// sync job 은 동작.
	if pgStore != nil {
		giteaWorker := gitea.NewSyncWorker(intgregrep.NewIntegrationRepository(pgStore), cfg.GiteaURL, cfg.GiteaToken)
		go func() {
			if err := giteaWorker.Run(ctx, 30*time.Second); err != nil && err != context.Canceled {
				log.Printf("gitea sync worker stopped: %v", err)
			}
		}()
		log.Printf("gitea sync worker enabled (env url=%q, per-provider sync, interval=30s)", cfg.GiteaURL)
	}

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("run server: %v", err)
	}
}

// buildKeycloakEventAuditEmitter — Keycloak event listener 의 best-effort audit emit
// 콜백. auditStore 가 nil 이면 nil 반환 (cron 이 audit 생략). DREQ intake token cron 의
// buildIntakeTokenAuditEmitter 패턴 정합. sourceEventID 는 puller 의 SHA256 dedup
// hash — store layer 의 partial UNIQUE INDEX (migration 000032) 가 backend crash +
// cursor revert 같은 edge 에서도 중복 INSERT 차단.
func buildKeycloakEventAuditEmitter(auditStore httpapi.AuditStore) auditsvc.AuditEmitter {
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
func buildIntakeTokenAuditEmitter(auditStore httpapi.AuditStore) devreqsvc.AuditEmitter {
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
