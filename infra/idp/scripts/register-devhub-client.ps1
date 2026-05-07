# register-devhub-client.ps1
#
# 목적: ADR-0001 §8.2 결정에 따라 DevHub frontend 를 first-party OIDC client 로
#       Hydra 에 등록한다. PKCE + skip_consent (silent consent).
#
# 전제:
#   - hydra binary 가 PATH 에 있어야 한다 (사용자가 샌드박스 외 터미널에서
#     `go install github.com/ory/hydra/v2/cmd/hydra@latest` 실행 후).
#   - Hydra 가 admin URL (http://localhost:4445) 에서 가동 중이어야 한다.
#
# 사용:
#   PowerShell> .\infra\idp\scripts\register-devhub-client.ps1
#
# 결과:
#   - client_id: devhub-frontend
#   - response_type: code
#   - grant_types: authorization_code, refresh_token
#   - token_endpoint_auth_method: none (public client, PKCE 가 client_secret 역할)
#   - skip_consent: true (first-party silent consent)
#
# 운영 진입 시점에는 redirect URI 를 운영 도메인으로 변경하고 secret 회전 정책을
# 별도 결정한다.

$ErrorActionPreference = "Stop"

$Endpoint = if ($env:HYDRA_ADMIN_URL) { $env:HYDRA_ADMIN_URL } else { "http://localhost:4445" }
$ClientId = "devhub-frontend"
$RedirectUri = "http://localhost:3000/auth/callback"
$PostLogoutUri = "http://localhost:3000/"

Write-Host "Hydra admin endpoint: $Endpoint"
Write-Host "Registering OIDC client: $ClientId"

# Hydra v2 CLI: `hydra create oauth2-client` (alias: `hydra create client`).
# 일부 플래그 이름은 v2 마이너 버전에 따라 달라질 수 있다. 실패 시 `hydra help create
# oauth2-client` 로 확인하고 본 스크립트의 플래그를 수정한다.
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
    Write-Error "Hydra client 등록 실패. (1) hydra binary PATH, (2) Hydra admin :4445 가동 여부, (3) v2 CLI 플래그 호환성 확인."
    exit 1
}

Write-Host ""
Write-Host "Done. Client '$ClientId' 등록 완료."
Write-Host "확인:"
Write-Host "  hydra list oauth2-clients --endpoint $Endpoint"
