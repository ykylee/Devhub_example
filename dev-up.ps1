# dev-up.ps1 - DevHub local services launcher (Windows PowerShell 5.1)
#
# Brings up: PostgreSQL migrations (idempotent), Kratos, Hydra, backend-core, frontend.
# Run from repo root:
#
#   .\dev-up.ps1           # start all services
#   .\dev-up.ps1 restart   # stop everything first, then start
#
# Env overrides:
#   DB_URL                  PostgreSQL DSN (default: postgres://postgres:postgres@localhost:5432/devhub?sslmode=disable)
#   DEVHUB_SKIP_MIGRATE     set to 1 to skip the migrate-up step
#   DEVHUB_SKIP_READY_WAIT  set to 1 to skip TCP readiness probes (faster, less safe)
#
# Notes:
# - ASCII-only English per repo memory (PowerShell 5.1 + non-BOM UTF-8 corrupts Korean).
# - PID files are written under .pids\<service>.pid for dev-down.ps1 to consume.
# - Errors abort the script (ErrorActionPreference = 'Stop'); readiness probes throw
#   on timeout so a misbehaving service does not silently leave the next stage waiting.

$ErrorActionPreference = 'Stop'

$RepoRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $RepoRoot

if ($args.Count -gt 0 -and $args[0] -eq 'restart') {
    Write-Host 'Restarting services (dev-down.ps1 first)...'
    & "$RepoRoot\dev-down.ps1"
    Start-Sleep -Seconds 2
}

Write-Host 'Starting DevHub local services...'

$PidDir = Join-Path $RepoRoot '.pids'
if (-not (Test-Path $PidDir)) {
    New-Item -ItemType Directory -Path $PidDir | Out-Null
}

function Wait-ForPort {
    param(
        [Parameter(Mandatory)][string]$Name,
        [Parameter(Mandatory)][int]$Port,
        [int]$TimeoutSec = 30
    )
    if ($env:DEVHUB_SKIP_READY_WAIT -eq '1') {
        Write-Host "  [skip-ready-wait] $Name on port $Port"
        return
    }
    $deadline = (Get-Date).AddSeconds($TimeoutSec)
    while ((Get-Date) -lt $deadline) {
        $client = $null
        try {
            $client = New-Object System.Net.Sockets.TcpClient
            $iar = $client.BeginConnect('127.0.0.1', $Port, $null, $null)
            $ok = $iar.AsyncWaitHandle.WaitOne(500)
            if ($ok -and $client.Connected) {
                Write-Host "  $Name ready on port $Port"
                return
            }
        } catch {
            # connection refused / not yet listening - retry
        } finally {
            if ($null -ne $client) { $client.Close() }
        }
        Start-Sleep -Milliseconds 250
    }
    throw "Timed out after ${TimeoutSec}s waiting for $Name on port $Port. Check the corresponding .log file."
}

function Start-BackgroundService {
    param(
        [Parameter(Mandatory)][string]$Name,
        [Parameter(Mandatory)][string]$Executable,
        [string[]]$Arguments = @(),
        [Parameter(Mandatory)][string]$LogFile,
        [string]$WorkingDir = $null
    )
    Write-Host "Starting $Name..."
    $absLog = if ([System.IO.Path]::IsPathRooted($LogFile)) { $LogFile } else { Join-Path $RepoRoot $LogFile }
    $params = @{
        FilePath               = $Executable
        ArgumentList           = $Arguments
        RedirectStandardOutput = $absLog
        RedirectStandardError  = "$absLog.err"
        NoNewWindow            = $true
        PassThru               = $true
    }
    if ($WorkingDir) { $params.WorkingDirectory = $WorkingDir }
    $proc = Start-Process @params
    $proc.Id | Out-File -FilePath (Join-Path $PidDir "$Name.pid") -Encoding ascii
}

function Mask-Dsn {
    param([string]$Dsn)
    return ($Dsn -replace ':[^:@/]+@', ':***@')
}

# Resolve DB_URL once so migrate-up + backend both see the same value.
if (-not $env:DB_URL) {
    $env:DB_URL = 'postgres://postgres:postgres@localhost:5432/devhub?sslmode=disable'
}
$DbUrl = $env:DB_URL

# 1. Migrations - idempotent. Skips silently if golang-migrate is not installed
#    (the operator can run `make migrate-tools` once to install it).
if ($env:DEVHUB_SKIP_MIGRATE -eq '1') {
    Write-Host '[skip-migrate] Skipping migrate-up.'
} elseif (Get-Command migrate -ErrorAction SilentlyContinue) {
    Write-Host "Applying migrations against $(Mask-Dsn $DbUrl)..."
    & migrate -path backend-core/migrations -database $DbUrl up
    if ($LASTEXITCODE -ne 0) {
        throw "migrate up failed with exit code $LASTEXITCODE"
    }
} else {
    Write-Warning 'migrate not on PATH. Run `make migrate-tools` once to install golang-migrate, or set DEVHUB_SKIP_MIGRATE=1 to suppress.'
}

# 2. Kratos
if (Get-Command kratos -ErrorAction SilentlyContinue) {
    Start-BackgroundService -Name 'kratos' -Executable 'kratos' `
        -Arguments @('serve', '-c', 'infra/idp/kratos.yaml', '--dev') `
        -LogFile 'kratos.log'
    Wait-ForPort -Name 'kratos-public' -Port 4433
    Wait-ForPort -Name 'kratos-admin'  -Port 4434
} else {
    Write-Warning 'kratos not on PATH; skipping. Backend will not be able to authenticate users.'
}

# 3. Hydra
if (Get-Command hydra -ErrorAction SilentlyContinue) {
    Start-BackgroundService -Name 'hydra' -Executable 'hydra' `
        -Arguments @('serve', 'all', '-c', 'infra/idp/hydra.yaml', '--dev') `
        -LogFile 'hydra.log'
    Wait-ForPort -Name 'hydra-public' -Port 4444
    Wait-ForPort -Name 'hydra-admin'  -Port 4445
} else {
    Write-Warning 'hydra not on PATH; skipping. OIDC code flow will not complete.'
}

# 4. backend-core
$env:AUTH_DEV_FALLBACK          = 'true'
$env:DEVHUB_AUTH_DEV_FALLBACK   = '1'
$env:DEVHUB_KRATOS_PUBLIC_URL   = 'http://localhost:4433'
$env:DEVHUB_KRATOS_ADMIN_URL    = 'http://localhost:4434'
$env:DEVHUB_HYDRA_PUBLIC_URL    = 'http://localhost:4444'
$env:DEVHUB_HYDRA_ADMIN_URL     = 'http://localhost:4445'
Start-BackgroundService -Name 'backend' -Executable 'go' `
    -Arguments @('run', '.') `
    -LogFile 'backend.log' `
    -WorkingDir (Join-Path $RepoRoot 'backend-core')
Wait-ForPort -Name 'backend' -Port 8080

# 5. frontend
Start-BackgroundService -Name 'frontend' -Executable 'npm' `
    -Arguments @('run', 'dev') `
    -LogFile 'frontend.log' `
    -WorkingDir (Join-Path $RepoRoot 'frontend')
Wait-ForPort -Name 'frontend' -Port 3000 -TimeoutSec 60

Write-Host ''
Write-Host 'All services up:'
Write-Host '  kratos      public 4433, admin 4434'
Write-Host '  hydra       public 4444, admin 4445'
Write-Host '  backend     8080  (http://localhost:8080/health)'
Write-Host '  frontend    3000  (http://localhost:3000/)'
Write-Host ''
Write-Host 'Stop with:  .\dev-down.ps1'
