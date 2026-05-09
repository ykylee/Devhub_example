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
	// AuthDevFallback enables development-only authentication fallbacks: requests with no Authorization header pass through authenticateActor, and the role guard middleware (requireMinRole) lets the request through without a role. Actor identity always falls back to "system" when no authenticated subject is present. Default false (production-safe). Toggle with DEVHUB_AUTH_DEV_FALLBACK=1.
	AuthDevFallback bool
	// HydraAdminURL is the base URL of the Ory Hydra admin API used by the introspection verifier (for example http://127.0.0.1:4445). Empty means no Hydra verifier is wired and authentication relies on AuthDevFallback or another verifier.
	HydraAdminURL string
	// HydraRoleClaim is a dotted path into the Hydra introspection response that holds the actor role. Defaults to "ext.role" when empty. See auth.HydraIntrospectionVerifier.RoleClaim for supported paths. Toggle with DEVHUB_HYDRA_ROLE_CLAIM.
	HydraRoleClaim string
	// ServiceActionExecutorMode enables the live service action worker only for supported explicit modes such as "simulation".
	ServiceActionExecutorMode string
	// ServiceActionAllowedServices is a comma-separated allowlist checked by the simulation service action executor.
	ServiceActionAllowedServices string
	// ServiceActionAllowedActions is a comma-separated allowlist checked by the simulation service action executor.
	ServiceActionAllowedActions string
	// KratosPublicURL is the base URL of the Ory Kratos public API (default
	// http://127.0.0.1:4433) used by the /api/v1/auth/login handler to drive
	// the password self-service login flow. Empty disables the login proxy
	// (frontend cannot complete the OIDC code flow).
	KratosPublicURL string
	// KratosAdminURL is the base URL of the Ory Kratos admin API (default
	// http://127.0.0.1:4434). Required for identity creation (Sign Up).
	KratosAdminURL string
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
		AuthDevFallback:              envBool("DEVHUB_AUTH_DEV_FALLBACK"),
		HydraAdminURL:                strings.TrimSpace(os.Getenv("DEVHUB_HYDRA_ADMIN_URL")),
		HydraRoleClaim:               strings.TrimSpace(os.Getenv("DEVHUB_HYDRA_ROLE_CLAIM")),
		ServiceActionExecutorMode:    strings.TrimSpace(os.Getenv("SERVICE_ACTION_EXECUTOR_MODE")),
		ServiceActionAllowedServices: strings.TrimSpace(os.Getenv("SERVICE_ACTION_ALLOWED_SERVICES")),
		ServiceActionAllowedActions:  strings.TrimSpace(os.Getenv("SERVICE_ACTION_ALLOWED_ACTIONS")),
		KratosPublicURL:              strings.TrimSpace(os.Getenv("DEVHUB_KRATOS_PUBLIC_URL")),
		KratosAdminURL:               strings.TrimSpace(os.Getenv("DEVHUB_KRATOS_ADMIN_URL")),
	}
}

// Validate reports whether the configuration is safe for startup given whether a bearer-token verifier has been wired up. In production (Env=="prod") it refuses startup when no verifier is configured or when AuthDevFallback is enabled. Dev mode is unconstrained. Env is normalized here so the contract holds for hand-built configs as well as those loaded via Load().
func (cfg Config) Validate(hasVerifier bool) error {
	if strings.ToLower(strings.TrimSpace(cfg.Env)) != "prod" {
		return nil
	}
	if !hasVerifier {
		return errors.New("DEVHUB_ENV=prod requires a bearer token verifier (set DEVHUB_HYDRA_ADMIN_URL or wire one in main.go)")
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
