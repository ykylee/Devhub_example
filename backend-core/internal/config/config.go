package config

import "os"

type Config struct {
	Port                         string
	DBURL                        string
	GiteaURL                     string
	GiteaToken                   string
	GiteaWebhookSecret           string
	BackendAIURL                 string
	AuthDevFallback              bool
	ServiceActionExecutorMode    string
	ServiceActionAllowedServices string
	ServiceActionAllowedActions  string
}

func Load() Config {
	return Config{
		Port:                         envOrDefault("PORT", "8080"),
		DBURL:                        os.Getenv("DB_URL"),
		GiteaURL:                     os.Getenv("GITEA_URL"),
		GiteaToken:                   os.Getenv("GITEA_TOKEN"),
		GiteaWebhookSecret:           os.Getenv("GITEA_WEBHOOK_SECRET"),
		BackendAIURL:                 os.Getenv("BACKEND_AI_URL"),
		AuthDevFallback:              envBool("DEVHUB_AUTH_DEV_FALLBACK"),
		ServiceActionExecutorMode:    os.Getenv("SERVICE_ACTION_EXECUTOR_MODE"),
		ServiceActionAllowedServices: os.Getenv("SERVICE_ACTION_ALLOWED_SERVICES"),
		ServiceActionAllowedActions:  os.Getenv("SERVICE_ACTION_ALLOWED_ACTIONS"),
	}
}

func envBool(key string) bool {
	switch os.Getenv(key) {
	case "1", "true", "TRUE", "yes", "YES", "on", "ON":
		return true
	default:
		return false
	}
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
