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
}

func Load() Config {
	return Config{
		Port:                         envOrDefault("PORT", "8080"),
		DBURL:                        os.Getenv("DB_URL"),
		GiteaURL:                     os.Getenv("GITEA_URL"),
		GiteaToken:                   os.Getenv("GITEA_TOKEN"),
		GiteaWebhookSecret:           os.Getenv("GITEA_WEBHOOK_SECRET"),
		BackendAIURL:                 os.Getenv("BACKEND_AI_URL"),
		Env:                          strings.ToLower(strings.TrimSpace(os.Getenv("DEVHUB_ENV"))),
		IdPProvider:                  normalizeIDPProvider(os.Getenv("DEVHUB_IDP_PROVIDER")),
		AuthDevFallback:              envBool("DEVHUB_AUTH_DEV_FALLBACK"),
		ServiceActionExecutorMode:    strings.TrimSpace(os.Getenv("SERVICE_ACTION_EXECUTOR_MODE")),
		ServiceActionAllowedServices: strings.TrimSpace(os.Getenv("SERVICE_ACTION_ALLOWED_SERVICES")),
		ServiceActionAllowedActions:  strings.TrimSpace(os.Getenv("SERVICE_ACTION_ALLOWED_ACTIONS")),
		OIDCIssuerURL:                strings.TrimSpace(os.Getenv("DEVHUB_OIDC_ISSUER_URL")),
		OIDCJWKSURL:                  strings.TrimSpace(os.Getenv("DEVHUB_OIDC_JWKS_URL")),
		OIDCClientID:                 strings.TrimSpace(os.Getenv("DEVHUB_OIDC_CLIENT_ID")),
		OIDCClientSecret:             strings.TrimSpace(os.Getenv("DEVHUB_OIDC_CLIENT_SECRET")),
		KeycloakAdminURL:             strings.TrimSpace(os.Getenv("DEVHUB_KEYCLOAK_ADMIN_URL")),
		KeycloakAdminRealm:           strings.TrimSpace(os.Getenv("DEVHUB_KEYCLOAK_ADMIN_REALM")),
		KeycloakAdminClientID:        strings.TrimSpace(os.Getenv("DEVHUB_KEYCLOAK_ADMIN_CLIENT_ID")),
		KeycloakAdminClientSecret:    strings.TrimSpace(os.Getenv("DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET")),
		InfraAgentToken:              strings.TrimSpace(os.Getenv("DEVHUB_INFRA_AGENT_TOKEN")),
		HomeLabDegradedStatuses:      strings.TrimSpace(os.Getenv("DEVHUB_HOMELAB_DEGRADED_STATUSES")),
		HomeLabProviderKey:           strings.TrimSpace(os.Getenv("DEVHUB_HOMELAB_PROVIDER_KEY")),
		HomeLabPullEnabled:           envBool("DEVHUB_HOMELAB_PULL_ENABLED"),
		HomeLabPullInterval:          strings.TrimSpace(os.Getenv("DEVHUB_HOMELAB_PULL_INTERVAL")),
		HomeLabPullFile:              strings.TrimSpace(os.Getenv("DEVHUB_HOMELAB_PULL_FILE")),
		HomeLabPullURL:               strings.TrimSpace(os.Getenv("DEVHUB_HOMELAB_PULL_URL")),
		HomeLabPullToken:             strings.TrimSpace(os.Getenv("DEVHUB_HOMELAB_PULL_TOKEN")),
		HomeLabPullHTTPRetryMax:      envInt("DEVHUB_HOMELAB_PULL_HTTP_RETRY_MAX"),
		HomeLabPullHTTPRetryBackoff:  strings.TrimSpace(os.Getenv("DEVHUB_HOMELAB_PULL_HTTP_RETRY_BACKOFF")),
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
