package main

import (
	"context"
	"log"

	"github.com/devhub/backend-core/internal/config"
	"github.com/devhub/backend-core/internal/httpapi"
	"github.com/devhub/backend-core/internal/store"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	var eventStore httpapi.WebhookEventStore
	var healthStore httpapi.HealthStore
	if cfg.DBURL != "" {
		pgStore, err := store.NewPostgresStore(ctx, cfg.DBURL)
		if err != nil {
			log.Fatalf("connect postgres: %v", err)
		}
		defer pgStore.Close()
		eventStore = pgStore
		healthStore = pgStore
	} else {
		log.Println("DB_URL is not set; webhook persistence is disabled")
	}

	router := httpapi.NewRouter(httpapi.RouterConfig{
		WebhookSecret: cfg.GiteaWebhookSecret,
		EventStore:    eventStore,
		HealthStore:   healthStore,
	})
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("run server: %v", err)
	}
}
