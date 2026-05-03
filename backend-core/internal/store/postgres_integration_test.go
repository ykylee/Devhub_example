package store_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/normalize"
	"github.com/devhub/backend-core/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresStoreProcessesNormalizedWebhookEvent(t *testing.T) {
	dbURL := os.Getenv("DEVHUB_TEST_DB_URL")
	if dbURL == "" {
		t.Skip("DEVHUB_TEST_DB_URL is not set")
	}

	ctx := context.Background()
	pgStore, err := store.NewPostgresStore(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect postgres store: %v", err)
	}
	defer pgStore.Close()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect cleanup pool: %v", err)
	}
	defer pool.Close()

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	repositoryName := "devhub-test/api-" + suffix
	senderLogin := "devhub-test-user-" + suffix
	dedupeKey := "integration-" + suffix
	defer cleanupIntegrationRows(ctx, t, pool, repositoryName, senderLogin, dedupeKey)

	payload := []byte(fmt.Sprintf(`{
		"repository": {
			"id": 0,
			"full_name": %q,
			"name": "api-%s",
			"default_branch": "main",
			"owner": {"login": "devhub-test"}
		},
		"sender": {
			"id": 0,
			"login": %q
		},
		"issue": {
			"id": 0,
			"number": 1,
			"title": "Integration issue",
			"state": "open",
			"user": {"login": %q},
			"created_at": "2026-05-03T10:00:00Z"
		}
	}`, repositoryName, suffix, senderLogin, senderLogin))

	now := time.Now().UTC()
	eventID, err := pgStore.SaveWebhookEvent(ctx, store.WebhookEvent{
		EventType:      "issues",
		DedupeKey:      dedupeKey,
		RepositoryName: repositoryName,
		SenderLogin:    senderLogin,
		Payload:        payload,
		Status:         "validated",
		ReceivedAt:     now,
		ValidatedAt:    &now,
	})
	if err != nil {
		t.Fatalf("save webhook event: %v", err)
	}

	event := store.WebhookEvent{
		ID:             eventID,
		EventType:      "issues",
		DedupeKey:      dedupeKey,
		RepositoryName: repositoryName,
		SenderLogin:    senderLogin,
		Payload:        payload,
		Status:         "validated",
		ReceivedAt:     now,
		ValidatedAt:    &now,
	}
	processor := normalize.Processor{Sink: pgStore}
	if err := processor.Process(ctx, event); err != nil {
		t.Fatalf("process event: %v", err)
	}

	events, err := pgStore.ListWebhookEvents(ctx, store.ListWebhookEventsOptions{Limit: 100})
	if err != nil {
		t.Fatalf("list webhook events: %v", err)
	}
	for _, listed := range events {
		if listed.DedupeKey == dedupeKey {
			if listed.Status != "processed" {
				t.Fatalf("expected processed status, got %q", listed.Status)
			}
			return
		}
	}
	t.Fatalf("saved event with dedupe key %q was not listed", dedupeKey)
}

func cleanupIntegrationRows(ctx context.Context, t *testing.T, pool *pgxpool.Pool, repositoryName, senderLogin, dedupeKey string) {
	t.Helper()
	queries := []string{
		`DELETE FROM webhook_events WHERE dedupe_key = $1`,
		`DELETE FROM issues WHERE repository_id IN (SELECT id FROM repositories WHERE full_name = $1)`,
		`DELETE FROM pull_requests WHERE repository_id IN (SELECT id FROM repositories WHERE full_name = $1)`,
		`DELETE FROM ci_runs WHERE repository_name = $1`,
		`DELETE FROM repositories WHERE full_name = $1`,
		`DELETE FROM gitea_users WHERE login = $1`,
	}
	args := [][]any{
		{dedupeKey},
		{repositoryName},
		{repositoryName},
		{repositoryName},
		{repositoryName},
		{senderLogin},
	}
	for i, query := range queries {
		if _, err := pool.Exec(ctx, query, args[i]...); err != nil && !strings.Contains(err.Error(), "does not exist") {
			t.Fatalf("cleanup query failed: %v", err)
		}
	}
}
