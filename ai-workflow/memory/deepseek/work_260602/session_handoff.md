# Session Handoff

- PR `#462` CI E2E failures reproduced locally in fresh env.
- Reproduced failing specs:
  - `frontend/tests/e2e/admin-applications.spec.ts` `TC-APP-SEARCH-01`
  - `frontend/tests/e2e/repositories-publish.spec.ts` `TC-REPO-PUBLISH-01`
- Local/CI parity fix applied: `scripts/setup-keycloak.sh` now grants `manage-users` to `devhub-backend` service account, matching CI so fresh local E2E seed does not fail with 403.
- Important local-only pitfall found: Next.js frontend rewrites are build-time bound to `BACKEND_API_URL`. Rebuilding frontend with `BACKEND_API_URL=http://localhost:18080` was required; otherwise `/api/v1/me` proxied to stale `localhost:8080` and auth looped to `/login`.
