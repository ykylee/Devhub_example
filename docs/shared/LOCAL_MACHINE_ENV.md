# Local Machine Environment Setup (Untracked)

This document describes the environment setup for this specific Mac.

## Components Installed via Brew
- **PostgreSQL 14**: Database at `localhost:5432` (no password).
- **Ory Hydra**: OIDC Provider.
    - Public: `http://localhost:4444`
    - Admin: `http://localhost:4445`
- **Ory Kratos**: Identity Management.
    - Public: `http://localhost:4433`
    - Admin: `http://localhost:4434`

## Setup Commands Performed
1. `brew install ory/tap/hydra ory/tap/kratos postgresql@14`
2. `brew services start postgresql@14`
3. Create DB and Schemas:
   ```bash
   createdb devhub
   psql -d devhub -f infra/idp/sql/001_create_idp_schemas.sql
   ```
4. Run Migrations:
   ```bash
   hydra migrate sql up --yes "postgres://localhost:5432/devhub?sslmode=disable&search_path=hydra"
   kratos migrate sql up --yes "postgres://localhost:5432/devhub?sslmode=disable&search_path=kratos"
   ```
5. Register OIDC Client:
   ```bash
   curl -X POST http://localhost:4445/admin/clients ... (see infra/idp/scripts/register-devhub-client.ps1)
   ```

## Running Services
### Hydra
```bash
hydra serve all --config infra/idp/hydra.yaml --dev
```
### Kratos
```bash
kratos serve --config infra/idp/kratos.yaml
```

## Backend Environment Variables for Local Testing
```bash
export DEVHUB_KRATOS_PUBLIC_URL=http://localhost:4433
export DEVHUB_KRATOS_ADMIN_URL=http://localhost:4434
export DEVHUB_HYDRA_ADMIN_URL=http://localhost:4445
export PORT=8080
export DEVHUB_AUTH_DEV_FALLBACK=1 # Optional, set to 0 to test full RBAC
```
