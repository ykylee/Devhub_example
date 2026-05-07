package main

import (
	"context"
	"log"
	"time"

	"github.com/devhub/backend-core/internal/commandworker"
	"github.com/devhub/backend-core/internal/config"
	"github.com/devhub/backend-core/internal/httpapi"
	"github.com/devhub/backend-core/internal/normalize"
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
	realtimeHub := httpapi.NewRealtimeHub()
	var worker *commandworker.Worker
	if cfg.DBURL != "" {
		pgStore, err := store.NewPostgresStore(ctx, cfg.DBURL)
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
		worker = &commandworker.Worker{Store: pgStore, Publisher: realtimeHub}
	} else {
		log.Println("DB_URL is not set; webhook persistence is disabled")
	}

	router := httpapi.NewRouter(httpapi.RouterConfig{
		WebhookSecret:     cfg.GiteaWebhookSecret,
		EventStore:        eventStore,
		EventProcessor:    eventProcessor,
		HealthStore:       healthStore,
		DomainStore:       domainStore,
		CommandStore:      commandStore,
		AuditStore:        auditStore,
		OrganizationStore: organizationStore,
		SnapshotProvider: httpapi.RuntimeSnapshotProvider{
			Base:         httpapi.StaticSnapshotProvider{},
			HealthStore:  healthStore,
			GiteaURL:     cfg.GiteaURL,
			BackendAIURL: cfg.BackendAIURL,
		},
		RealtimeHub: realtimeHub,
	})
	if worker != nil {
		go func() {
			if err := worker.Run(ctx, 2*time.Second); err != nil && err != context.Canceled {
				log.Printf("command worker stopped: %v", err)
			}
		}()
	}
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("run server: %v", err)
	}
}
