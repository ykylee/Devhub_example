package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port               string
	DBURL              string
	GiteaURL           string
	GiteaToken         string
	GiteaWebhookSecret string
	BackendAIURL       string
	// Env selects the runtime mode. "prod" enables fail-fast guards in Config.Validate (no verifier => refuse startup; AuthDevFallback => refuse startup). Anything else is treated as dev. Toggle with DEVHUB_ENV.
	Env string
	// IdPProvider selects which identity provider integration is active.
	// Supported: "keycloak" only.
	IdPProvider string
	// AuthDevFallback enables development-only authentication fallbacks: requests with no Authorization header pass through authenticateActor, and the role guard middleware (requireMinRole) lets the request through without a role. Actor identity always falls back to "system" when no authenticated subject is present. Default false (production-safe). Toggle with DEVHUB_AUTH_DEV_FALLBACK=1.
	AuthDevFallback bool
	// OnboardingGateEnabled — RM-ONBOARD-01 (ADR-0021 §3.3). Feature flag
	// default **false** (legacy 동작 — lazy auto-create 유지, ADR-0020 sub-carve B).
	// Toggle with `DEVHUB_ONBOARDING_GATE_ENABLED=1`. true 시 신규 onboarding
	// 흐름 활성화 — authenticateActor 가 DB miss 를 token-only actor 로 취급 +
	// onboardingGate middleware 가 미완료 사용자의 allowlist 외 endpoint 차단.
	// Carve A 단독 머지 후 main 안정성 보장 — Carve D acceptance 통과 후 별도
	// hotfix PR 으로 default ON flip.
	OnboardingGateEnabled bool
	// ServiceActionExecutorMode enables the live service action worker only for supported explicit modes such as "simulation".
	ServiceActionExecutorMode string
	// ServiceActionAllowedServices is a comma-separated allowlist checked by the simulation service action executor.
	ServiceActionAllowedServices string
	// ServiceActionAllowedActions is a comma-separated allowlist checked by the simulation service action executor.
	ServiceActionAllowedActions string
	// OIDCIssuerURL is the OpenID issuer URL used by generic OIDC providers
	// such as Keycloak (for example
	// https://keycloak.example.com/realms/devhub).
	OIDCIssuerURL string
	// OIDCJWKSURL optionally overrides JWKS discovery when set.
	OIDCJWKSURL string
	// OIDCClientID and OIDCClientSecret are shared by auth/account adapters
	// that call IdP token/admin endpoints.
	OIDCClientID     string
	OIDCClientSecret string
	// OIDCJWKSMaxStaleDuration — ADR-0020 sub-carve D (sprint -l, issue #213).
	// JWKS cache TTL 만료 후 Keycloak unreachable 시 stale-while-error fallback
	// 으로 사용 가능한 최대 시간 (e.g. "24h", "12h"). 빈 값 또는 invalid 면
	// keycloak_verifier 내부 default (24h) 적용. Keycloak key rotation 주기
	// (권장 90일) 보다 짧게 운영해야 revoked key 보호 회복 한도 유지.
	OIDCJWKSMaxStaleDuration string
	// Keycloak admin access settings (used by Keycloak adapter paths).
	KeycloakAdminURL          string
	KeycloakAdminRealm        string
	KeycloakAdminClientID     string
	KeycloakAdminClientSecret string
	// InfraAgentToken is the shared secret used by HomeLab agents on
	// POST /api/v1/infra/services/snapshot (API-77). Empty value keeps the
	// endpoint unavailable (503) so ingest auth misconfiguration fails loud.
	InfraAgentToken string
	// HomeLabDegradedStatuses configures which health_status values should be
	// treated as degraded for adapter policy (comma-separated, e.g.
	// "warning,degraded,down").
	HomeLabDegradedStatuses string
	// HomeLabProviderKey is the provider identifier used in degraded_providers
	// emitted by HomeLab adapter policy.
	HomeLabProviderKey string
	// HomeLabPullEnabled enables background pull-and-ingest loop.
	HomeLabPullEnabled bool
	// HomeLabPullInterval controls pull loop interval (time.ParseDuration format).
	HomeLabPullInterval string
	// HomeLabPullFile points to a local JSON fixture file for File puller mode.
	HomeLabPullFile string
	// HomeLabPullURL points to HomeLab agent snapshot endpoint for HTTP pull mode.
	HomeLabPullURL string
	// HomeLabPullToken is an optional bearer token for HTTP pull mode.
	HomeLabPullToken string
	// HomeLabPullHTTPRetryMax is max retry count for HTTP pull mode.
	HomeLabPullHTTPRetryMax int
	// HomeLabPullHTTPRetryBackoff controls retry backoff duration (time.ParseDuration format).
	HomeLabPullHTTPRetryBackoff string
	// HomeLabPullMaxBytes caps the snapshot payload size (file or HTTP body) in
	// bytes. 0 (default) means unlimited (legacy behavior). Production-recommended
	// value is 5 MB. ADR-0015 §6 (1) — size limit + streaming decode.
	HomeLabPullMaxBytes int64
	// DREQTokenCronEnabled toggles the DREQ intake token cron loop (ADR-0017
	// §6 (a)+(c)+(d)). When true, a goroutine periodically hard-revokes expired
	// tokens and emits expiring-soon / stale Prometheus gauges. Default false.
	DREQTokenCronEnabled bool
	// DREQTokenCronInterval controls cron loop interval (time.ParseDuration
	// format). Default 10m when unset.
	DREQTokenCronInterval string
	// DREQTokenExpiringSoonThreshold marks tokens whose expires_at is within this
	// duration as "expiring soon" (time.ParseDuration). Default 24h.
	DREQTokenExpiringSoonThreshold string
	// DREQTokenStaleThreshold marks tokens with no last_used_at within this
	// duration as "stale". Empty / "0" disables the stale metric (no count). Default 720h (30d).
	DREQTokenStaleThreshold string
	// KeycloakEventListenerEnabled toggles the Keycloak Admin event listener cron
	// loop (ADR-0019 §5.3 (9), sprint claude/work_260519-v PR-C). When true and
	// KeycloakAdminURL etc. are configured, a goroutine periodically polls
	// /admin/realms/{realm}/events + /admin-events and emits to audit_logs.
	// Default false.
	KeycloakEventListenerEnabled bool
	// KeycloakEventListenerInterval controls the puller tick interval
	// (time.ParseDuration). Default 30s when unset.
	KeycloakEventListenerInterval string
	// KeycloakEventListenerMaxEvents caps the per-tick page size. Default 500
	// when unset.
	KeycloakEventListenerMaxEvents int
	// KeycloakWebhookSecret is the shared secret token to verify Keycloak SPI webhook pushes
	KeycloakWebhookSecret string
}

func Load() Config {
	return Config{
		Port:                           envOrDefault("PORT", "8080"),
		DBURL:                          os.Getenv("DB_URL"),
		GiteaURL:                       os.Getenv("GITEA_URL"),
		GiteaToken:                     os.Getenv("GITEA_TOKEN"),
		GiteaWebhookSecret:             os.Getenv("GITEA_WEBHOOK_SECRET"),
		BackendAIURL:                   os.Getenv("BACKEND_AI_URL"),
		Env:                            strings.ToLower(strings.TrimSpace(os.Getenv("DEVHUB_ENV"))),
		IdPProvider:                    normalizeIDPProvider(os.Getenv("DEVHUB_IDP_PROVIDER")),
		AuthDevFallback:                envBool("DEVHUB_AUTH_DEV_FALLBACK"),
		OnboardingGateEnabled:          envBool("DEVHUB_ONBOARDING_GATE_ENABLED"),
		ServiceActionExecutorMode:      strings.TrimSpace(os.Getenv("SERVICE_ACTION_EXECUTOR_MODE")),
		ServiceActionAllowedServices:   strings.TrimSpace(os.Getenv("SERVICE_ACTION_ALLOWED_SERVICES")),
		ServiceActionAllowedActions:    strings.TrimSpace(os.Getenv("SERVICE_ACTION_ALLOWED_ACTIONS")),
		OIDCIssuerURL:                  strings.TrimSpace(os.Getenv("DEVHUB_OIDC_ISSUER_URL")),
		OIDCJWKSURL:                    strings.TrimSpace(os.Getenv("DEVHUB_OIDC_JWKS_URL")),
		OIDCJWKSMaxStaleDuration:       strings.TrimSpace(os.Getenv("DEVHUB_OIDC_JWKS_MAX_STALE_DURATION")),
		OIDCClientID:                   strings.TrimSpace(os.Getenv("DEVHUB_OIDC_CLIENT_ID")),
		OIDCClientSecret:               strings.TrimSpace(os.Getenv("DEVHUB_OIDC_CLIENT_SECRET")),
		KeycloakAdminURL:               strings.TrimSpace(os.Getenv("DEVHUB_KEYCLOAK_ADMIN_URL")),
		KeycloakAdminRealm:             strings.TrimSpace(os.Getenv("DEVHUB_KEYCLOAK_ADMIN_REALM")),
		KeycloakAdminClientID:          strings.TrimSpace(os.Getenv("DEVHUB_KEYCLOAK_ADMIN_CLIENT_ID")),
		KeycloakAdminClientSecret:      strings.TrimSpace(os.Getenv("DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET")),
		InfraAgentToken:                strings.TrimSpace(os.Getenv("DEVHUB_INFRA_AGENT_TOKEN")),
		HomeLabDegradedStatuses:        strings.TrimSpace(os.Getenv("DEVHUB_HOMELAB_DEGRADED_STATUSES")),
		HomeLabProviderKey:             strings.TrimSpace(os.Getenv("DEVHUB_HOMELAB_PROVIDER_KEY")),
		HomeLabPullEnabled:             envBool("DEVHUB_HOMELAB_PULL_ENABLED"),
		HomeLabPullInterval:            strings.TrimSpace(os.Getenv("DEVHUB_HOMELAB_PULL_INTERVAL")),
		HomeLabPullFile:                strings.TrimSpace(os.Getenv("DEVHUB_HOMELAB_PULL_FILE")),
		HomeLabPullURL:                 strings.TrimSpace(os.Getenv("DEVHUB_HOMELAB_PULL_URL")),
		HomeLabPullToken:               strings.TrimSpace(os.Getenv("DEVHUB_HOMELAB_PULL_TOKEN")),
		HomeLabPullHTTPRetryMax:        envInt("DEVHUB_HOMELAB_PULL_HTTP_RETRY_MAX"),
		HomeLabPullHTTPRetryBackoff:    strings.TrimSpace(os.Getenv("DEVHUB_HOMELAB_PULL_HTTP_RETRY_BACKOFF")),
		HomeLabPullMaxBytes:            envInt64("DEVHUB_HOMELAB_PULL_MAX_BYTES"),
		DREQTokenCronEnabled:           envBool("DEVHUB_DREQ_TOKEN_CRON_ENABLED"),
		DREQTokenCronInterval:          strings.TrimSpace(os.Getenv("DEVHUB_DREQ_TOKEN_CRON_INTERVAL")),
		DREQTokenExpiringSoonThreshold: strings.TrimSpace(os.Getenv("DEVHUB_DREQ_TOKEN_EXPIRING_SOON_THRESHOLD")),
		DREQTokenStaleThreshold:        strings.TrimSpace(os.Getenv("DEVHUB_DREQ_TOKEN_STALE_THRESHOLD")),
		KeycloakEventListenerEnabled:   envBool("DEVHUB_KEYCLOAK_EVENT_LISTENER_ENABLED"),
		KeycloakEventListenerInterval:  strings.TrimSpace(os.Getenv("DEVHUB_KEYCLOAK_EVENT_LISTENER_INTERVAL")),
		KeycloakEventListenerMaxEvents: envInt("DEVHUB_KEYCLOAK_EVENT_LISTENER_MAX_EVENTS"),
		KeycloakWebhookSecret:          strings.TrimSpace(os.Getenv("DEVHUB_KEYCLOAK_SPI_WEBHOOK_SECRET")),
	}
}

// Validate reports whether the configuration is safe for startup given whether a bearer-token verifier has been wired up. In production (Env=="prod") it refuses startup when no verifier is configured or when AuthDevFallback is enabled. Dev mode is unconstrained. Env is normalized here so the contract holds for hand-built configs as well as those loaded via Load().
func (cfg Config) Validate(hasVerifier bool) error {
	switch normalizeIDPProvider(cfg.IdPProvider) {
	case "keycloak":
	default:
		return errors.New("DEVHUB_IDP_PROVIDER must be: keycloak")
	}

	if strings.ToLower(strings.TrimSpace(cfg.Env)) != "prod" {
		return nil
	}
	if !hasVerifier {
		return errors.New("DEVHUB_ENV=prod requires a bearer token verifier (set OIDC/Keycloak env and wire one in main.go)")
	}
	if cfg.AuthDevFallback {
		return errors.New("DEVHUB_ENV=prod must not set DEVHUB_AUTH_DEV_FALLBACK=1; remove it or change DEVHUB_ENV")
	}
	return nil
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func envBool(key string) bool {
	enabled, _ := strconv.ParseBool(strings.TrimSpace(os.Getenv(key)))
	return enabled
}

func envInt(key string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(os.Getenv(key)))
	return n
}

func normalizeIDPProvider(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	if v == "" {
		return "keycloak"
	}
	return v
}

func envInt64(key string) int64 {
	n, _ := strconv.ParseInt(strings.TrimSpace(os.Getenv(key)), 10, 64)
	return n
}
