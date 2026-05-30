package repository_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	repo "github.com/devhub/backend-core/internal/domain/repository-integration/repository"
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

func TestIntegration_Repository_UpsertAndList(t *testing.T) {
	pgStore, pool, ctx := newIntegrationTestStore(t)
	integRepo := repo.NewIntegrationRepository(pgStore)

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	providerKey := "scm-test-" + suffix

	// ── Seed integration_provider ──────────────────────────────────────────
	const seedProvider = `
		INSERT INTO integration_providers (provider_key, provider_type, display_name, enabled, auth_mode, capabilities, credentials_ref)
		VALUES ($1, 'scm', 'SCM Test', true, 'token', '["repo:sync"]', '{}')
		RETURNING provider_id::text`
	var providerID string
	if err := pool.QueryRow(ctx, seedProvider, providerKey).Scan(&providerID); err != nil {
		t.Fatalf("seed integration provider: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM repositories WHERE provider_id = $1::uuid`, providerID)
		_, _ = pool.Exec(ctx, `DELETE FROM integration_providers WHERE provider_id = $1::uuid`, providerID)
	})

	t.Run("UpsertRepository_InsertNew", func(t *testing.T) {
		repo := domain.Repository{
			FullName:      "org/test-repo-" + suffix,
			OwnerLogin:    "org",
			Name:          "test-repo-" + suffix,
			CloneURL:      "https://scm.test/org/test-repo.git",
			HTMLURL:       "https://scm.test/org/test-repo",
			DefaultBranch: "main",
			Private:       false,
			Source:        "scm",
			ProviderID:    providerID,
			Description:   "Test repository",
		}
		if err := integRepo.UpsertRepository(ctx, repo); err != nil {
			t.Fatalf("UpsertRepository (insert) failed: %v", err)
		}

		// Verify the row was inserted
		var fullName, ownerLogin, name, status string
		err := pool.QueryRow(ctx, `SELECT full_name, owner_login, name, repository_status FROM repositories WHERE full_name = $1`, repo.FullName).
			Scan(&fullName, &ownerLogin, &name, &status)
		if err != nil {
			t.Fatalf("query inserted row: %v", err)
		}
		if fullName != repo.FullName || name != repo.Name || status != "active" {
			t.Fatalf("unexpected row: full_name=%q name=%q status=%q", fullName, name, status)
		}
	})

	t.Run("UpsertRepository_UpdateExisting", func(t *testing.T) {
		repo := domain.Repository{
			FullName:      "org/test-repo-" + suffix, // same full_name → ON CONFLICT
			OwnerLogin:    "org",
			Name:          "test-repo-" + suffix,
			CloneURL:      "https://scm.test/org/test-repo-updated.git",
			HTMLURL:       "https://scm.test/org/test-repo-updated",
			DefaultBranch: "develop",
			Private:       true,
			Source:        "scm",
			ProviderID:    providerID,
			Description:   "Updated description",
		}
		if err := integRepo.UpsertRepository(ctx, repo); err != nil {
			t.Fatalf("UpsertRepository (update) failed: %v", err)
		}

		var cloneURL, defaultBranch string
		err := pool.QueryRow(ctx, `SELECT clone_url, default_branch FROM repositories WHERE full_name = $1`, repo.FullName).
			Scan(&cloneURL, &defaultBranch)
		if err != nil {
			t.Fatalf("query updated row: %v", err)
		}
		if cloneURL != repo.CloneURL || defaultBranch != repo.DefaultBranch {
			t.Fatalf("expected updated fields: clone_url=%q default_branch=%q", cloneURL, defaultBranch)
		}
	})

	t.Run("UpsertRepository_InsertWithMinimalFields", func(t *testing.T) {
		suffix2 := fmt.Sprintf("%d", time.Now().UnixNano())
		minimal := domain.Repository{
			FullName:   "org/minimal-" + suffix2,
			Name:       "minimal-" + suffix2,
			ProviderID: providerID,
			Source:     "scm",
		}
		if err := integRepo.UpsertRepository(ctx, minimal); err != nil {
			t.Fatalf("UpsertRepository (minimal) failed: %v", err)
		}

		var name, status, defaultBranch string
		err := pool.QueryRow(ctx, `SELECT name, repository_status, COALESCE(default_branch, '') FROM repositories WHERE full_name = $1`, minimal.FullName).
			Scan(&name, &status, &defaultBranch)
		if err != nil {
			t.Fatalf("query minimal row: %v", err)
		}
		if name != minimal.Name || status != "active" || defaultBranch != "" {
			t.Fatalf("unexpected minimal row: name=%q status=%q default_branch=%q", name, status, defaultBranch)
		}
	})

	t.Run("ListRepositoriesByProvider_ReturnsRepos", func(t *testing.T) {
		repos, err := integRepo.ListRepositoriesByProvider(ctx, providerID)
		if err != nil {
			t.Fatalf("ListRepositoriesByProvider failed: %v", err)
		}
		if len(repos) == 0 {
			t.Fatal("expected at least one repository, got 0")
		}
		// All returned repos should belong to the provider
		for _, r := range repos {
			if r.ProviderID != providerID {
				t.Fatalf("repo %q has provider_id=%q, expected %q", r.FullName, r.ProviderID, providerID)
			}
		}
	})

	t.Run("ListRepositoriesByProvider_EmptyForUnknownProvider", func(t *testing.T) {
		ghostUUID := "00000000-0000-0000-0000-000000000000"
		repos, err := integRepo.ListRepositoriesByProvider(ctx, ghostUUID)
		if err != nil {
			t.Fatalf("ListRepositoriesByProvider for ghost provider: %v", err)
		}
		if len(repos) != 0 {
			t.Fatalf("expected 0 repos for ghost provider, got %d", len(repos))
		}
	})
}
