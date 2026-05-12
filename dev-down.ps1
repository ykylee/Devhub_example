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
    # Returns $true when a PID file existed (regardless of whether the process
    # was still alive) so the caller can decide to add this service's ports to
    # the sweep list. When no PID file exists, dev-up did not start the
    # service - leave any externally-managed listener intact.
    param([string]$Name)
    $pidFile = Join-Path $PidDir "$Name.pid"
    if (-not (Test-Path $pidFile)) {
        return $false
    }
    $svcPid = (Get-Content $pidFile -ErrorAction SilentlyContinue).Trim()
    if (-not $svcPid) {
        Remove-Item $pidFile -ErrorAction SilentlyContinue
        return $true
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
    return $true
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

# Reverse start order: frontend -> backend -> hydra -> kratos. Only the
# services dev-up actually spawned (PID file present) contribute their ports
# to the sweep list, so externally-managed Kratos/Hydra/backend/frontend
# instances are left alone.
$servicePorts = [ordered]@{
    'frontend' = @(3000)
    'backend'  = @(8080)
    'hydra'    = @(4444, 4445)
    'kratos'   = @(4433, 4434)
}

$sweepPorts = New-Object System.Collections.Generic.List[int]
foreach ($name in $servicePorts.Keys) {
    if (Stop-ServiceByPid $name) {
        foreach ($p in $servicePorts[$name]) { [void]$sweepPorts.Add($p) }
    } else {
        $ports = ($servicePorts[$name] -join ', ')
        Write-Host "  $name not started by this script; leaving any listener on port(s) $ports intact"
    }
}

# Safety net only for services we actually started.
if ($sweepPorts.Count -gt 0) {
    Stop-ByPort -Ports $sweepPorts.ToArray()
}

Write-Host 'All DevHub services stopped.'
