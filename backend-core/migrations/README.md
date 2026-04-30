# Backend Core Migrations

PostgreSQL schema changes for the Go Core service are managed with `golang-migrate/migrate`.

```bash
make migrate-tools
make migrate-create NAME=create_webhook_events
make migrate-up
make migrate-version
```

The default local database URL is `postgres://user:pass@localhost:5432/devhub?sslmode=disable`. Override it with `MIGRATE_DB_URL` when running against another environment.
