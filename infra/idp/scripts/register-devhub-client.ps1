# register-devhub-client.ps1
#
# Purpose: Register the DevHub frontend as a first-party OIDC client on Hydra,
#          per ADR-0001 §8.2 (silent consent, PKCE, public client).
#
# Implementation note: Hydra v26 CLI does not honor a --client-id flag and
# always assigns a random UUID. To make the PoC reproducible, this script
# calls the Hydra Admin REST API directly so it can set client_id explicitly.
#
# Prerequisites:
#   - hydra Admin URL is reachable (default http://localhost:4445).
#   - PowerShell 5.1 or later (Invoke-RestMethod).
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
# Operational note: replace redirect/post-logout URIs with production hosts and
# decide on secret rotation before going to production.
#
# Note: ASCII-only on purpose (PowerShell 5.1 + BOM-less UTF-8 corrupts non-ASCII).

$ErrorActionPreference = "Stop"

$AdminUrl = if ($env:HYDRA_ADMIN_URL) { $env:HYDRA_ADMIN_URL } else { "http://localhost:4445" }
$ClientId = if ($env:DEVHUB_OIDC_CLIENT_ID) { $env:DEVHUB_OIDC_CLIENT_ID } else { "devhub-frontend" }
$RedirectUri = if ($env:DEVHUB_OIDC_REDIRECT_URI) { $env:DEVHUB_OIDC_REDIRECT_URI } else { "http://localhost:3000/auth/callback" }
$PostLogoutUri = if ($env:DEVHUB_OIDC_POST_LOGOUT_URI) { $env:DEVHUB_OIDC_POST_LOGOUT_URI } else { "http://localhost:3000/" }

$body = @{
    client_id                  = $ClientId
    client_name                = "DevHub Frontend (first-party)"
    grant_types                = @("authorization_code", "refresh_token")
    response_types             = @("code")
    redirect_uris              = @($RedirectUri)
    post_logout_redirect_uris  = @($PostLogoutUri)
    scope                      = "openid offline_access email profile"
    token_endpoint_auth_method = "none"
    skip_consent               = $true
    skip_logout_consent        = $true
}

Write-Host "Hydra admin URL: $AdminUrl"
Write-Host "Registering OIDC client: $ClientId"

# If the client already exists, replace it (DELETE then POST). The Admin API
# also supports PUT /admin/clients/<id> for upsert, but DELETE+POST is simpler
# and keeps the request body identical to a fresh creation.
try {
    $existing = Invoke-RestMethod -Method Get -Uri "$AdminUrl/admin/clients/$ClientId" -ErrorAction Stop
    if ($existing) {
        Write-Host "Existing client '$ClientId' found; deleting before recreate."
        Invoke-RestMethod -Method Delete -Uri "$AdminUrl/admin/clients/$ClientId" | Out-Null
    }
} catch {
    if ($_.Exception.Response.StatusCode.value__ -ne 404) {
        Write-Warning "Existence check failed (non-404). Continuing. Detail: $($_.Exception.Message)"
    }
}

$created = Invoke-RestMethod -Method Post `
    -Uri "$AdminUrl/admin/clients" `
    -ContentType "application/json" `
    -Body ($body | ConvertTo-Json -Depth 4)

Write-Host ""
Write-Host "Created:"
$created | Select-Object client_id, client_name, grant_types, response_types, redirect_uris, scope, token_endpoint_auth_method, skip_consent | Format-List
Write-Host "Verify:"
Write-Host "  hydra list oauth2-clients --endpoint $AdminUrl"
