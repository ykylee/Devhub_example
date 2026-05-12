# dev-down.ps1 - DevHub local services shutdown (Windows PowerShell 5.1)
#
# Stops services launched by dev-up.ps1 in reverse order (frontend first so
# clients drain before the backend goes away), then sweeps the known ports as
# a safety net for stale processes whose PID files were lost.
#
# Run from repo root:
#
#   .\dev-down.ps1
#
# ASCII-only English per repo memory (PowerShell 5.1 + non-BOM UTF-8 corrupts Korean).

$ErrorActionPreference = 'Continue'  # one stuck service should not block the others

$RepoRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $RepoRoot

$PidDir = Join-Path $RepoRoot '.pids'

function Stop-ServiceByPid {
    param([string]$Name)
    $pidFile = Join-Path $PidDir "$Name.pid"
    if (-not (Test-Path $pidFile)) {
        Write-Host "  $Name not tracked (no PID file)"
        return
    }
    $svcPid = (Get-Content $pidFile -ErrorAction SilentlyContinue).Trim()
    if (-not $svcPid) {
        Remove-Item $pidFile -ErrorAction SilentlyContinue
        return
    }
    $proc = Get-Process -Id $svcPid -ErrorAction SilentlyContinue
    if ($null -ne $proc) {
        try {
            Stop-Process -Id $svcPid -Force -ErrorAction Stop
            Write-Host "  $Name stopped (PID $svcPid)"
        } catch {
            Write-Warning "  $Name (PID $svcPid) stop failed: $_"
        }
    } else {
        Write-Host "  $Name (PID $svcPid) already gone"
    }
    Remove-Item $pidFile -ErrorAction SilentlyContinue
}

function Stop-ByPort {
    # Safety sweep: any leftover process holding a known DevHub port gets killed.
    # Triggers when a previous dev-up was killed before writing its PID files,
    # or when a service was launched outside dev-up.
    param([int[]]$Ports)
    foreach ($port in $Ports) {
        $conns = Get-NetTCPConnection -LocalPort $port -State Listen -ErrorAction SilentlyContinue
        foreach ($c in $conns) {
            $procId = $c.OwningProcess
            if (-not $procId -or $procId -eq 0) { continue }
            $proc = Get-Process -Id $procId -ErrorAction SilentlyContinue
            if ($null -ne $proc) {
                try {
                    Stop-Process -Id $procId -Force -ErrorAction Stop
                    Write-Host "  swept PID $procId on port $port ($($proc.ProcessName))"
                } catch {
                    Write-Warning "  PID $procId on port $port stop failed: $_"
                }
            }
        }
    }
}

Write-Host 'Stopping DevHub local services...'

# Reverse start order: frontend -> backend -> hydra -> kratos.
Stop-ServiceByPid 'frontend'
Stop-ServiceByPid 'backend'
Stop-ServiceByPid 'hydra'
Stop-ServiceByPid 'kratos'

# Safety net for stale binaries holding well-known ports.
Stop-ByPort -Ports @(3000, 8080, 4433, 4434, 4444, 4445)

Write-Host 'All DevHub services stopped.'
