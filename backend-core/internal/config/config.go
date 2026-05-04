package config

import "os"

type Config struct {
	Port               string
	DBURL              string
	GiteaURL           string
	GiteaToken         string
	GiteaWebhookSecret string
	BackendAIURL       string
}

func Load() Config {
	return Config{
		Port:               envOrDefault("PORT", "8080"),
		DBURL:              os.Getenv("DB_URL"),
		GiteaURL:           os.Getenv("GITEA_URL"),
		GiteaToken:         os.Getenv("GITEA_TOKEN"),
		GiteaWebhookSecret: os.Getenv("GITEA_WEBHOOK_SECRET"),
		BackendAIURL:       os.Getenv("BACKEND_AI_URL"),
	}
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
