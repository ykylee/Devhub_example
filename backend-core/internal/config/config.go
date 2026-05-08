package config

import (
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
	// AuthDevFallback enables development-only authentication fallbacks: requests with no Authorization header are allowed through, and the X-Devhub-Actor header is honoured as the actor identity. Default false (production-safe). Toggle with DEVHUB_AUTH_DEV_FALLBACK=1.
	AuthDevFallback bool
}

func Load() Config {
	return Config{
		Port:               envOrDefault("PORT", "8080"),
		DBURL:              os.Getenv("DB_URL"),
		GiteaURL:           os.Getenv("GITEA_URL"),
		GiteaToken:         os.Getenv("GITEA_TOKEN"),
		GiteaWebhookSecret: os.Getenv("GITEA_WEBHOOK_SECRET"),
		BackendAIURL:       os.Getenv("BACKEND_AI_URL"),
		AuthDevFallback:    envBool("DEVHUB_AUTH_DEV_FALLBACK"),
	}
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
