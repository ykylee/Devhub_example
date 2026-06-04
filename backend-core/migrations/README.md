# Backend Core Migrations

PostgreSQL schema changes for the Go Core service are managed with `golang-migrate/migrate`.

- Active baseline chain:
  - `000001_initial_schema`
  - `000002_seed_system_rbac`
  - `000003_seed_bootstrap_catalog`
- Legacy historical chain (`000001_create_webhook_events` ~ `000047_normalize_team_manager_display_name`) is preserved under `backend-core/migrations-legacy-20260604/` for reference only and is no longer used by `make migrate-*` or CI.

```bash
make migrate-tools
make migrate-create NAME=add_example_feature
make migrate-up
make migrate-version
```

The default local database URL is `postgres://user:pass@localhost:5432/devhub?sslmode=disable`. Override it with `MIGRATE_DB_URL` when running against another environment.
