# register-devhub-client.ps1
#
# Purpose: Register the DevHub frontend as a first-party OIDC client on Hydra,
#          per ADR-0001 §8.2 (silent consent, PKCE, public client).
#
# Prerequisites:
#   - hydra binary on PATH (install via infra/idp/scripts/install-binaries.ps1).
#     `go install` is blocked by replace directives in Hydra/Kratos go.mod.
#   - Hydra is running on http://localhost:4445 (admin URL).
#
# Usage:
#   PowerShell> .\infra\idp\scripts\register-devhub-client.ps1
#
# Result:
#   - client_id: devhub-frontend
#   - response_type: code
#   - grant_types: authorization_code, refresh_token
#   - token_endpoint_auth_method: none (public client; PKCE replaces secret)
#   - skip_consent: true (first-party silent consent)
#
# Operational note: Replace redirect/post-logout URIs with production hostnames
# and decide on secret rotation before going to production.
#
# Note: ASCII-only on purpose (PowerShell 5.1 + BOM-less UTF-8 corrupts non-ASCII).

$ErrorActionPreference = "Stop"

$Endpoint = if ($env:HYDRA_ADMIN_URL) { $env:HYDRA_ADMIN_URL } else { "http://localhost:4445" }
$ClientId = "devhub-frontend"
$RedirectUri = "http://localhost:3000/auth/callback"
$PostLogoutUri = "http://localhost:3000/"

Write-Host "Hydra admin endpoint: $Endpoint"
Write-Host "Registering OIDC client: $ClientId"

# Hydra v2 CLI: `hydra create oauth2-client` (alias: `hydra create client`).
# Flag names may vary slightly across v2 minor versions. If this fails, run
# `hydra help create oauth2-client` and adjust the flags below.
& hydra create oauth2-client `
  --endpoint $Endpoint `
  --name "DevHub Frontend (first-party)" `
  --grant-type "authorization_code,refresh_token" `
  --response-type "code" `
  --redirect-uri $RedirectUri `
  --post-logout-callback $PostLogoutUri `
  --scope "openid,offline_access,email,profile" `
  --token-endpoint-auth-method "none" `
  --skip-consent `
  --skip-logout-consent

if ($LASTEXITCODE -ne 0) {
    Write-Error "Hydra client registration failed. Check (1) hydra binary on PATH, (2) Hydra admin :4445 is up, (3) v2 CLI flag compatibility."
    exit 1
}

Write-Host ""
Write-Host "Done. Client '$ClientId' registered."
Write-Host "Verify:"
Write-Host "  hydra list oauth2-clients --endpoint $Endpoint"
