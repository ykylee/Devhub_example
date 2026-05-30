package repository_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	intrep "github.com/devhub/backend-core/internal/domain/integration-registry/repository"
	"github.com/devhub/backend-core/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

func newIntegrationTestStore(t *testing.T) (*store.PostgresStore, *pgxpool.Pool, context.Context) {
	t.Helper()
	dbURL := os.Getenv("DEVHUB_TEST_DB_URL")
	if dbURL == "" {
		t.Skip("DEVHUB_TEST_DB_URL is not set")
	}
	ctx := context.Background()
	pgStore, err := store.NewPostgresStore(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect postgres store: %v", err)
	}
	t.Cleanup(pgStore.Close)

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect raw pool: %v", err)
	}
	t.Cleanup(pool.Close)
	return pgStore, pool, ctx
}

func TestIntegration_IntegrationRegistry_CRUD(t *testing.T) {
	pgStore, pool, ctx := newIntegrationTestStore(t)
	repo := intrep.NewIntegrationRepository(pgStore)

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	providerKey := "gitea-test-" + suffix

	// Cleanup seed
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM integration_providers WHERE provider_key = $1`, providerKey)
	}()

	// 1. Get non-existent provider by key
	if _, err := repo.GetIntegrationProviderByKey(ctx, providerKey); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound for ghost provider key, got %v", err)
	}

	// 2. Get non-existent provider by ID
	ghostUUID := "00000000-0000-0000-0000-000000000000"
	if _, err := repo.GetIntegrationProviderByID(ctx, ghostUUID); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound for ghost provider ID, got %v", err)
	}

	// 3. Seed integration_provider directly to pool
	var providerID string
	err := pool.QueryRow(ctx, `
		INSERT INTO integration_providers (provider_key, provider_type, display_name, enabled, auth_mode, capabilities, credentials_ref)
		VALUES ($1, 'scm', 'SCM Gitea', true, 'token', '["repo:sync"]', '{}')
		RETURNING provider_id::text
	`, providerKey).Scan(&providerID)
	if err != nil {
		t.Fatalf("seed integration provider: %v", err)
	}

	// 4. Get by ID
	provider, err := repo.GetIntegrationProviderByID(ctx, providerID)
	if err != nil {
		t.Fatalf("GetIntegrationProviderByID failed: %v", err)
	}
	if provider.ProviderKey != providerKey {
		t.Errorf("expected provider key %s, got %s", providerKey, provider.ProviderKey)
	}

	// 5. Get by Key
	providerByKey, err := repo.GetIntegrationProviderByKey(ctx, providerKey)
	if err != nil {
		t.Fatalf("GetIntegrationProviderByKey failed: %v", err)
	}
	if providerByKey.ID != providerID {
		t.Errorf("id mismatch: %s vs %s", providerByKey.ID, providerID)
	}

	// 6. ListIntegrationProviders
	opts := intrep.IntegrationProviderListOptions{
		ProviderType: "scm",
	}
	providers, total, err := repo.ListIntegrationProviders(ctx, opts)
	if err != nil {
		t.Fatalf("ListIntegrationProviders: %v", err)
	}
	if len(providers) == 0 || total == 0 {
		t.Errorf("expected at least 1 provider, got 0")
	}
}

func TestIntegration_ExternalTaskStore_CRUD(t *testing.T) {
	_, pool, ctx := newIntegrationTestStore(t)
	taskStore := intrep.NewPostgresExternalTaskStore(pool)

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	providerKey := "jira-task-" + suffix
	externalID := "JIRA-101-" + suffix

	// 1. Seed provider for FK constraint
	var providerID string
	err := pool.QueryRow(ctx, `
		INSERT INTO integration_providers (provider_key, provider_type, display_name, enabled, auth_mode, credentials_ref)
		VALUES ($1, 'alm', 'Jira Tracker', true, 'token', '{}')
		RETURNING provider_id::text
	`, providerKey).Scan(&providerID)
	if err != nil {
		t.Fatalf("seed integration provider for task: %v", err)
	}
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM external_task_items WHERE provider_id = $1::uuid`, providerID)
		_, _ = pool.Exec(ctx, `DELETE FROM integration_providers WHERE provider_id = $1::uuid`, providerID)
	}()

	fetchedAt := time.Now().Add(-10 * time.Minute)
	seq, err := taskStore.NextWebhookSeq(ctx)
	if err != nil {
		t.Fatalf("NextWebhookSeq failed: %v", err)
	}

	taskItem := domain.ExternalTaskItem{
		ProviderID:       providerID,
		ExternalID:       externalID,
		Title:            "Task Title " + suffix,
		Description:      "Task Desc " + suffix,
		RawStatus:        "In Progress",
		NormalizedStatus: "active",
		Priority:         "high",
		Assignee:         "alice",
		Reporter:         "bob",
		URL:              "http://jira.com/" + externalID,
		Labels:           []string{"bug", "critical"},
		WebhookSeq:       &seq,
		FetchedAt:        fetchedAt,
		RawPayload:       []byte(`{"custom":"payload"}`),
	}

	// 2. Upsert (Insert)
	created, err := taskStore.UpsertExternalTaskItem(ctx, taskItem)
	if err != nil {
		t.Fatalf("UpsertExternalTaskItem (insert) failed: %v", err)
	}
	if created.ExternalID != externalID || created.Title != taskItem.Title {
		t.Errorf("unexpected created task details: %+v", created)
	}

	// 3. Upsert (Update conflict)
	taskItem.Title = "Task Title Updated " + suffix
	updated, err := taskStore.UpsertExternalTaskItem(ctx, taskItem)
	if err != nil {
		t.Fatalf("UpsertExternalTaskItem (update) failed: %v", err)
	}
	if updated.Title != taskItem.Title {
		t.Errorf("expected updated title, got %s", updated.Title)
	}

	// 4. Get by UUID
	loaded, err := taskStore.GetExternalTaskItemByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetExternalTaskItemByID failed: %v", err)
	}
	if loaded.ID != created.ID || loaded.Title != taskItem.Title {
		t.Errorf("unexpected loaded task: %+v", loaded)
	}

	// 5. ListExternalTaskItems
	opts := intrep.ExternalTaskListOptions{
		ProviderID:       providerID,
		NormalizedStatus: "active",
		Labels:           []string{"bug"},
	}
	tasks, total, err := taskStore.ListExternalTaskItems(ctx, opts)
	if err != nil {
		t.Fatalf("ListExternalTaskItems failed: %v", err)
	}
	if len(tasks) != 1 || total != 1 {
		t.Errorf("expected exactly 1 task, got len=%d total=%d", len(tasks), total)
	}

	// 6. DetectWebhookSeqGaps
	gaps, err := taskStore.DetectWebhookSeqGaps(ctx, providerID)
	if err != nil {
		t.Fatalf("DetectWebhookSeqGaps failed: %v", err)
	}
	if gaps != 0 {
		t.Errorf("expected 0 gaps, got %d", gaps)
	}

	// 7. UpdateProviderLastPulledAt
	pulledAt := time.Now()
	err = taskStore.UpdateProviderLastPulledAt(ctx, providerID, pulledAt)
	if err != nil {
		t.Fatalf("UpdateProviderLastPulledAt failed: %v", err)
	}

	// 8. ListTaskTrackers
	trackers, err := taskStore.ListTaskTrackers(ctx)
	if err != nil {
		t.Fatalf("ListTaskTrackers failed: %v", err)
	}
	if len(trackers) == 0 {
		t.Errorf("expected at least 1 task tracker")
	}

	// 9. SoftDeleteExternalTaskItem
	err = taskStore.SoftDeleteExternalTaskItem(ctx, providerID, externalID)
	if err != nil {
		t.Fatalf("SoftDeleteExternalTaskItem failed: %v", err)
	}
	loadedDeleted, err := taskStore.GetExternalTaskItemByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetExternalTaskItemByID after delete failed: %v", err)
	}
	if loadedDeleted.DeletedAt == nil {
		t.Errorf("expected DeletedAt to be set after soft deletion")
	}

	// Soft delete non-existent
	if err := taskStore.SoftDeleteExternalTaskItem(ctx, providerID, "GHOST-999"); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound for soft deleting ghost, got %v", err)
	}
}
