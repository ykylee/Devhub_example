package config

import (
	"strings"
	"testing"
)

// load_test covers Config.Load + the unexported env helpers
// (envOrDefault / envBool / envBoolDefault / envInt / envInt64 +
// normalizeIDPProvider / normalizeProjectModel). It manipulates env via
// t.Setenv so cleanup is automatic per the go 1.17+ contract.

// clearAllConfigEnv resets every env var Load consults so individual cases
// can rely on a deterministic baseline regardless of host env.
func clearAllConfigEnv(t *testing.T) {
	t.Helper()
	keys := []string{
		"PORT",
		"DB_URL",
		"GITEA_URL",
		"GITEA_TOKEN",
		"GITEA_WEBHOOK_SECRET",
		"BACKEND_AI_URL",
		"DEVHUB_ENV",
		"DEVHUB_IDP_PROVIDER",
		"DEVHUB_AUTH_DEV_FALLBACK",
		"DEVHUB_ONBOARDING_GATE_ENABLED",
		"DEVHUB_PROJECT_MODEL",
		"SERVICE_ACTION_EXECUTOR_MODE",
		"SERVICE_ACTION_ALLOWED_SERVICES",
		"SERVICE_ACTION_ALLOWED_ACTIONS",
		"DEVHUB_OIDC_ISSUER_URL",
		"DEVHUB_OIDC_JWKS_URL",
		"DEVHUB_OIDC_JWKS_MAX_STALE_DURATION",
		"DEVHUB_OIDC_CLIENT_ID",
		"DEVHUB_OIDC_CLIENT_SECRET",
		"DEVHUB_KEYCLOAK_ADMIN_URL",
		"DEVHUB_KEYCLOAK_ADMIN_REALM",
		"DEVHUB_KEYCLOAK_ADMIN_CLIENT_ID",
		"DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET",
		"DEVHUB_INFRA_AGENT_TOKEN",
		"DEVHUB_HOMELAB_DEGRADED_STATUSES",
		"DEVHUB_HOMELAB_PROVIDER_KEY",
		"DEVHUB_HOMELAB_PULL_ENABLED",
		"DEVHUB_HOMELAB_PULL_INTERVAL",
		"DEVHUB_HOMELAB_PULL_FILE",
		"DEVHUB_HOMELAB_PULL_URL",
		"DEVHUB_HOMELAB_PULL_TOKEN",
		"DEVHUB_HOMELAB_PULL_HTTP_RETRY_MAX",
		"DEVHUB_HOMELAB_PULL_HTTP_RETRY_BACKOFF",
		"DEVHUB_HOMELAB_PULL_MAX_BYTES",
		"DEVHUB_DREQ_TOKEN_CRON_ENABLED",
		"DEVHUB_DREQ_TOKEN_CRON_INTERVAL",
		"DEVHUB_DREQ_TOKEN_EXPIRING_SOON_THRESHOLD",
		"DEVHUB_DREQ_TOKEN_STALE_THRESHOLD",
		"DEVHUB_KEYCLOAK_EVENT_LISTENER_ENABLED",
		"DEVHUB_KEYCLOAK_EVENT_LISTENER_INTERVAL",
		"DEVHUB_KEYCLOAK_EVENT_LISTENER_MAX_EVENTS",
		"DEVHUB_KEYCLOAK_SPI_WEBHOOK_SECRET",
	}
	for _, k := range keys {
		t.Setenv(k, "")
	}
}

func TestLoad_Defaults_AllUnset(t *testing.T) {
	clearAllConfigEnv(t)
	cfg := Load()

	if cfg.Port != "8080" {
		t.Errorf("Port default = %q, want 8080", cfg.Port)
	}
	if cfg.DBURL != "" {
		t.Errorf("DBURL = %q, want empty", cfg.DBURL)
	}
	if cfg.Env != "" {
		t.Errorf("Env = %q, want empty (dev mode)", cfg.Env)
	}
	if cfg.IdPProvider != "keycloak" {
		t.Errorf("IdPProvider default = %q, want keycloak", cfg.IdPProvider)
	}
	if cfg.AuthDevFallback {
		t.Error("AuthDevFallback default = true, want false")
	}
	if !cfg.OnboardingGateEnabled {
		t.Error("OnboardingGateEnabled default = false, want true (2026-05-21 flip)")
	}
	if cfg.ProjectModel != "hybrid" {
		t.Errorf("ProjectModel default = %q, want hybrid", cfg.ProjectModel)
	}
	if cfg.HomeLabPullEnabled {
		t.Error("HomeLabPullEnabled default = true, want false")
	}
	if cfg.HomeLabPullHTTPRetryMax != 0 {
		t.Errorf("HomeLabPullHTTPRetryMax default = %d, want 0", cfg.HomeLabPullHTTPRetryMax)
	}
	if cfg.HomeLabPullMaxBytes != 0 {
		t.Errorf("HomeLabPullMaxBytes default = %d, want 0", cfg.HomeLabPullMaxBytes)
	}
	if cfg.DREQTokenCronEnabled {
		t.Error("DREQTokenCronEnabled default = true, want false")
	}
	if cfg.KeycloakEventListenerEnabled {
		t.Error("KeycloakEventListenerEnabled default = true, want false")
	}
	if cfg.KeycloakEventListenerMaxEvents != 0 {
		t.Errorf("KeycloakEventListenerMaxEvents default = %d, want 0", cfg.KeycloakEventListenerMaxEvents)
	}
}

func TestLoad_PortFromEnv(t *testing.T) {
	clearAllConfigEnv(t)
	t.Setenv("PORT", "9090")
	cfg := Load()
	if cfg.Port != "9090" {
		t.Errorf("Port = %q, want 9090", cfg.Port)
	}
}

func TestLoad_PortEmptyFallsBackToDefault(t *testing.T) {
	clearAllConfigEnv(t)
	t.Setenv("PORT", "")
	cfg := Load()
	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want 8080 (empty falls back)", cfg.Port)
	}
}

func TestLoad_EnvNormalisation(t *testing.T) {
	cases := []struct {
		raw  string
		want string
	}{
		{"prod", "prod"},
		{"PROD", "prod"},
		{" Prod ", "prod"},
		{"dev", "dev"},
		{"", ""},
	}
	for _, tc := range cases {
		t.Run("env="+tc.raw, func(t *testing.T) {
			clearAllConfigEnv(t)
			t.Setenv("DEVHUB_ENV", tc.raw)
			cfg := Load()
			if cfg.Env != tc.want {
				t.Errorf("Env = %q, want %q", cfg.Env, tc.want)
			}
		})
	}
}

func TestLoad_IdPProviderNormalisation(t *testing.T) {
	cases := []struct {
		raw  string
		want string
	}{
		{"keycloak", "keycloak"},
		{"KEYCLOAK", "keycloak"},
		{" Keycloak ", "keycloak"},
		{"", "keycloak"}, // empty defaults to keycloak (ADR-0019)
	}
	for _, tc := range cases {
		t.Run("provider="+tc.raw, func(t *testing.T) {
			clearAllConfigEnv(t)
			t.Setenv("DEVHUB_IDP_PROVIDER", tc.raw)
			cfg := Load()
			if cfg.IdPProvider != tc.want {
				t.Errorf("IdPProvider = %q, want %q", cfg.IdPProvider, tc.want)
			}
		})
	}
}

func TestLoad_AuthDevFallbackParsing(t *testing.T) {
	cases := []struct {
		raw  string
		want bool
	}{
		{"1", true},
		{"true", true},
		{"TRUE", true},
		{"0", false},
		{"false", false},
		{"", false},
		{"not-a-bool", false}, // ParseBool err → zero value
	}
	for _, tc := range cases {
		t.Run("fallback="+tc.raw, func(t *testing.T) {
			clearAllConfigEnv(t)
			t.Setenv("DEVHUB_AUTH_DEV_FALLBACK", tc.raw)
			cfg := Load()
			if cfg.AuthDevFallback != tc.want {
				t.Errorf("AuthDevFallback = %v, want %v", cfg.AuthDevFallback, tc.want)
			}
		})
	}
}

func TestLoad_OnboardingGateEnabled_OptOutOnZero(t *testing.T) {
	// envBoolDefault is opt-out: empty → true, "0"/"false" → false, garbage → true.
	cases := []struct {
		raw  string
		want bool
	}{
		{"", true},
		{"  ", true}, // trimmed to empty
		{"0", false},
		{"false", false},
		{"FALSE", false},
		{"1", true},
		{"true", true},
		{"garbage", true}, // ParseBool err → default (true)
	}
	for _, tc := range cases {
		t.Run("gate="+tc.raw, func(t *testing.T) {
			clearAllConfigEnv(t)
			t.Setenv("DEVHUB_ONBOARDING_GATE_ENABLED", tc.raw)
			cfg := Load()
			if cfg.OnboardingGateEnabled != tc.want {
				t.Errorf("OnboardingGateEnabled raw=%q = %v, want %v", tc.raw, cfg.OnboardingGateEnabled, tc.want)
			}
		})
	}
}

func TestLoad_ProjectModelNormalisation(t *testing.T) {
	cases := []struct {
		raw  string
		want string
	}{
		{"legacy", "legacy"},
		{"v2", "v2"},
		{"hybrid", "hybrid"},
		{"LEGACY", "legacy"},
		{" v2 ", "v2"},
		{"", "hybrid"},        // empty → default hybrid
		{"unknown", "hybrid"}, // invalid → default hybrid
	}
	for _, tc := range cases {
		t.Run("model="+tc.raw, func(t *testing.T) {
			clearAllConfigEnv(t)
			t.Setenv("DEVHUB_PROJECT_MODEL", tc.raw)
			cfg := Load()
			if cfg.ProjectModel != tc.want {
				t.Errorf("ProjectModel = %q, want %q", cfg.ProjectModel, tc.want)
			}
		})
	}
}

func TestLoad_PassThroughTrimmedStrings(t *testing.T) {
	clearAllConfigEnv(t)
	// All TrimSpace-wrapped getters should drop leading/trailing whitespace.
	t.Setenv("DEVHUB_OIDC_ISSUER_URL", "  https://kc.example.com/realms/devhub  ")
	t.Setenv("DEVHUB_OIDC_JWKS_URL", " https://kc.example.com/jwks ")
	t.Setenv("DEVHUB_OIDC_CLIENT_ID", "\tdevhub-frontend\t")
	t.Setenv("DEVHUB_OIDC_CLIENT_SECRET", "\nsekret\n")
	t.Setenv("DEVHUB_INFRA_AGENT_TOKEN", " agent-secret ")
	t.Setenv("DEVHUB_HOMELAB_PULL_FILE", " /tmp/snap.json ")
	t.Setenv("SERVICE_ACTION_EXECUTOR_MODE", " simulation ")
	t.Setenv("DEVHUB_OIDC_JWKS_MAX_STALE_DURATION", " 12h ")

	cfg := Load()
	if cfg.OIDCIssuerURL != "https://kc.example.com/realms/devhub" {
		t.Errorf("OIDCIssuerURL not trimmed: %q", cfg.OIDCIssuerURL)
	}
	if cfg.OIDCJWKSURL != "https://kc.example.com/jwks" {
		t.Errorf("OIDCJWKSURL not trimmed: %q", cfg.OIDCJWKSURL)
	}
	if cfg.OIDCClientID != "devhub-frontend" {
		t.Errorf("OIDCClientID not trimmed: %q", cfg.OIDCClientID)
	}
	if cfg.OIDCClientSecret != "sekret" {
		t.Errorf("OIDCClientSecret not trimmed: %q", cfg.OIDCClientSecret)
	}
	if cfg.InfraAgentToken != "agent-secret" {
		t.Errorf("InfraAgentToken not trimmed: %q", cfg.InfraAgentToken)
	}
	if cfg.HomeLabPullFile != "/tmp/snap.json" {
		t.Errorf("HomeLabPullFile not trimmed: %q", cfg.HomeLabPullFile)
	}
	if cfg.ServiceActionExecutorMode != "simulation" {
		t.Errorf("ServiceActionExecutorMode not trimmed: %q", cfg.ServiceActionExecutorMode)
	}
	if cfg.OIDCJWKSMaxStaleDuration != "12h" {
		t.Errorf("OIDCJWKSMaxStaleDuration not trimmed: %q", cfg.OIDCJWKSMaxStaleDuration)
	}
}

func TestLoad_HomeLabPullIntsAndBools(t *testing.T) {
	clearAllConfigEnv(t)
	t.Setenv("DEVHUB_HOMELAB_PULL_ENABLED", "true")
	t.Setenv("DEVHUB_HOMELAB_PULL_HTTP_RETRY_MAX", "5")
	t.Setenv("DEVHUB_HOMELAB_PULL_MAX_BYTES", "5242880") // 5 MiB
	t.Setenv("DEVHUB_HOMELAB_PULL_INTERVAL", " 30s ")
	t.Setenv("DEVHUB_HOMELAB_PULL_URL", " https://homelab.example.com ")
	t.Setenv("DEVHUB_HOMELAB_PULL_TOKEN", " homelab-token ")
	t.Setenv("DEVHUB_HOMELAB_PULL_HTTP_RETRY_BACKOFF", " 500ms ")

	cfg := Load()
	if !cfg.HomeLabPullEnabled {
		t.Error("HomeLabPullEnabled = false, want true")
	}
	if cfg.HomeLabPullHTTPRetryMax != 5 {
		t.Errorf("HomeLabPullHTTPRetryMax = %d, want 5", cfg.HomeLabPullHTTPRetryMax)
	}
	if cfg.HomeLabPullMaxBytes != 5242880 {
		t.Errorf("HomeLabPullMaxBytes = %d, want 5242880", cfg.HomeLabPullMaxBytes)
	}
	if cfg.HomeLabPullInterval != "30s" {
		t.Errorf("HomeLabPullInterval = %q, want 30s", cfg.HomeLabPullInterval)
	}
	if cfg.HomeLabPullURL != "https://homelab.example.com" {
		t.Errorf("HomeLabPullURL = %q", cfg.HomeLabPullURL)
	}
	if cfg.HomeLabPullToken != "homelab-token" {
		t.Errorf("HomeLabPullToken = %q", cfg.HomeLabPullToken)
	}
	if cfg.HomeLabPullHTTPRetryBackoff != "500ms" {
		t.Errorf("HomeLabPullHTTPRetryBackoff = %q", cfg.HomeLabPullHTTPRetryBackoff)
	}
}

func TestLoad_HomeLabPullIntInvalidFallsToZero(t *testing.T) {
	clearAllConfigEnv(t)
	t.Setenv("DEVHUB_HOMELAB_PULL_HTTP_RETRY_MAX", "not-an-int")
	t.Setenv("DEVHUB_HOMELAB_PULL_MAX_BYTES", "not-an-int64")
	t.Setenv("DEVHUB_KEYCLOAK_EVENT_LISTENER_MAX_EVENTS", "abc")

	cfg := Load()
	if cfg.HomeLabPullHTTPRetryMax != 0 {
		t.Errorf("HomeLabPullHTTPRetryMax invalid = %d, want 0", cfg.HomeLabPullHTTPRetryMax)
	}
	if cfg.HomeLabPullMaxBytes != 0 {
		t.Errorf("HomeLabPullMaxBytes invalid = %d, want 0", cfg.HomeLabPullMaxBytes)
	}
	if cfg.KeycloakEventListenerMaxEvents != 0 {
		t.Errorf("KeycloakEventListenerMaxEvents invalid = %d, want 0", cfg.KeycloakEventListenerMaxEvents)
	}
}

func TestLoad_DREQAndKeycloakEventListenerKnobs(t *testing.T) {
	clearAllConfigEnv(t)
	t.Setenv("DEVHUB_DREQ_TOKEN_CRON_ENABLED", "1")
	t.Setenv("DEVHUB_DREQ_TOKEN_CRON_INTERVAL", " 5m ")
	t.Setenv("DEVHUB_DREQ_TOKEN_EXPIRING_SOON_THRESHOLD", " 24h ")
	t.Setenv("DEVHUB_DREQ_TOKEN_STALE_THRESHOLD", " 720h ")
	t.Setenv("DEVHUB_KEYCLOAK_EVENT_LISTENER_ENABLED", "true")
	t.Setenv("DEVHUB_KEYCLOAK_EVENT_LISTENER_INTERVAL", " 30s ")
	t.Setenv("DEVHUB_KEYCLOAK_EVENT_LISTENER_MAX_EVENTS", "500")
	t.Setenv("DEVHUB_KEYCLOAK_SPI_WEBHOOK_SECRET", " spi-secret ")

	cfg := Load()
	if !cfg.DREQTokenCronEnabled {
		t.Error("DREQTokenCronEnabled false")
	}
	if cfg.DREQTokenCronInterval != "5m" {
		t.Errorf("DREQTokenCronInterval = %q", cfg.DREQTokenCronInterval)
	}
	if cfg.DREQTokenExpiringSoonThreshold != "24h" {
		t.Errorf("DREQTokenExpiringSoonThreshold = %q", cfg.DREQTokenExpiringSoonThreshold)
	}
	if cfg.DREQTokenStaleThreshold != "720h" {
		t.Errorf("DREQTokenStaleThreshold = %q", cfg.DREQTokenStaleThreshold)
	}
	if !cfg.KeycloakEventListenerEnabled {
		t.Error("KeycloakEventListenerEnabled false")
	}
	if cfg.KeycloakEventListenerInterval != "30s" {
		t.Errorf("KeycloakEventListenerInterval = %q", cfg.KeycloakEventListenerInterval)
	}
	if cfg.KeycloakEventListenerMaxEvents != 500 {
		t.Errorf("KeycloakEventListenerMaxEvents = %d", cfg.KeycloakEventListenerMaxEvents)
	}
	if cfg.KeycloakWebhookSecret != "spi-secret" {
		t.Errorf("KeycloakWebhookSecret = %q", cfg.KeycloakWebhookSecret)
	}
}

func TestLoad_RawPassThroughDBAndGiteaAndAI(t *testing.T) {
	// Raw os.Getenv (no TrimSpace) — pin that legacy keys are not trimmed.
	clearAllConfigEnv(t)
	t.Setenv("DB_URL", " postgres://x ")
	t.Setenv("GITEA_URL", " https://gitea ")
	t.Setenv("GITEA_TOKEN", " gitea-token ")
	t.Setenv("GITEA_WEBHOOK_SECRET", " webhook-secret ")
	t.Setenv("BACKEND_AI_URL", " http://ai ")

	cfg := Load()
	if !strings.HasPrefix(cfg.DBURL, " ") {
		t.Errorf("DBURL is unexpectedly trimmed: %q", cfg.DBURL)
	}
	if !strings.HasPrefix(cfg.GiteaURL, " ") {
		t.Errorf("GiteaURL is unexpectedly trimmed: %q", cfg.GiteaURL)
	}
	if !strings.HasPrefix(cfg.GiteaToken, " ") {
		t.Errorf("GiteaToken is unexpectedly trimmed: %q", cfg.GiteaToken)
	}
	if !strings.HasPrefix(cfg.GiteaWebhookSecret, " ") {
		t.Errorf("GiteaWebhookSecret is unexpectedly trimmed: %q", cfg.GiteaWebhookSecret)
	}
	if !strings.HasPrefix(cfg.BackendAIURL, " ") {
		t.Errorf("BackendAIURL is unexpectedly trimmed: %q", cfg.BackendAIURL)
	}
}

// Direct coverage for the unexported helper functions — they're trivial but
// the low-level branches (empty/invalid/numeric edge) are most likely to drift
// without a unit guard.

func TestEnvOrDefault(t *testing.T) {
	t.Setenv("UNIT_TEST_ENV_OR_DEFAULT", "")
	if got := envOrDefault("UNIT_TEST_ENV_OR_DEFAULT", "fallback"); got != "fallback" {
		t.Errorf("empty env should yield fallback, got %q", got)
	}
	t.Setenv("UNIT_TEST_ENV_OR_DEFAULT", "explicit")
	if got := envOrDefault("UNIT_TEST_ENV_OR_DEFAULT", "fallback"); got != "explicit" {
		t.Errorf("set env should win, got %q", got)
	}
}

func TestEnvBool(t *testing.T) {
	cases := []struct {
		raw  string
		want bool
	}{
		{"1", true}, {"true", true}, {"TRUE", true}, {"True", true},
		{"0", false}, {"false", false}, {"", false}, {"garbage", false},
		{"  true  ", true}, // trimmed before parse
	}
	for _, tc := range cases {
		t.Run("raw="+tc.raw, func(t *testing.T) {
			t.Setenv("UNIT_TEST_ENV_BOOL", tc.raw)
			if got := envBool("UNIT_TEST_ENV_BOOL"); got != tc.want {
				t.Errorf("envBool(%q) = %v, want %v", tc.raw, got, tc.want)
			}
		})
	}
}

func TestEnvBoolDefault(t *testing.T) {
	cases := []struct {
		raw  string
		def  bool
		want bool
	}{
		{"", true, true},          // empty → default true
		{"", false, false},        // empty → default false
		{"  ", true, true},        // trimmed empty → default
		{"garbage", true, true},   // parse err → default
		{"garbage", false, false}, // parse err → default
		{"1", false, true},        // explicit overrides default
		{"0", true, false},        // explicit overrides default
		{"true", false, true},
		{"false", true, false},
	}
	for _, tc := range cases {
		t.Run(tc.raw, func(t *testing.T) {
			t.Setenv("UNIT_TEST_ENV_BOOL_DEFAULT", tc.raw)
			if got := envBoolDefault("UNIT_TEST_ENV_BOOL_DEFAULT", tc.def); got != tc.want {
				t.Errorf("envBoolDefault(%q, %v) = %v, want %v", tc.raw, tc.def, got, tc.want)
			}
		})
	}
}

func TestEnvInt(t *testing.T) {
	cases := []struct {
		raw  string
		want int
	}{
		{"0", 0}, {"42", 42}, {"-7", -7},
		{"", 0}, {"abc", 0}, {"  100  ", 100},
	}
	for _, tc := range cases {
		t.Run(tc.raw, func(t *testing.T) {
			t.Setenv("UNIT_TEST_ENV_INT", tc.raw)
			if got := envInt("UNIT_TEST_ENV_INT"); got != tc.want {
				t.Errorf("envInt(%q) = %d, want %d", tc.raw, got, tc.want)
			}
		})
	}
}

func TestEnvInt64(t *testing.T) {
	cases := []struct {
		raw  string
		want int64
	}{
		{"0", 0}, {"5242880", 5242880}, {"-1", -1},
		{"", 0}, {"abc", 0}, {"  100  ", 100},
		{"9223372036854775807", 9223372036854775807}, // max int64
	}
	for _, tc := range cases {
		t.Run(tc.raw, func(t *testing.T) {
			t.Setenv("UNIT_TEST_ENV_INT64", tc.raw)
			if got := envInt64("UNIT_TEST_ENV_INT64"); got != tc.want {
				t.Errorf("envInt64(%q) = %d, want %d", tc.raw, got, tc.want)
			}
		})
	}
}

func TestNormalizeIDPProvider_Direct(t *testing.T) {
	cases := []struct {
		raw  string
		want string
	}{
		{"", "keycloak"},
		{"keycloak", "keycloak"},
		{"KEYCLOAK", "keycloak"},
		{"  Keycloak  ", "keycloak"},
		{"oidc", "oidc"},          // pass-through (Validate rejects)
		{"unknown", "unknown"},    // pass-through (Validate rejects)
	}
	for _, tc := range cases {
		t.Run(tc.raw, func(t *testing.T) {
			if got := normalizeIDPProvider(tc.raw); got != tc.want {
				t.Errorf("normalizeIDPProvider(%q) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}

func TestNormalizeProjectModel_Direct(t *testing.T) {
	cases := []struct {
		raw  string
		want string
	}{
		{"legacy", "legacy"},
		{"v2", "v2"},
		{"hybrid", "hybrid"},
		{"  LEGACY  ", "legacy"},
		{"V2", "v2"},
		{"", "hybrid"},
		{"unknown", "hybrid"},
		{"prod", "hybrid"},
	}
	for _, tc := range cases {
		t.Run(tc.raw, func(t *testing.T) {
			if got := normalizeProjectModel(tc.raw); got != tc.want {
				t.Errorf("normalizeProjectModel(%q) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}
