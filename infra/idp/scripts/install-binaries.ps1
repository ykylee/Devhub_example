# install-binaries.ps1
#
# Purpose: Download Ory Hydra + Ory Kratos Windows binaries (SQLite-enabled)
#          from GitHub releases and place them on a PATH-visible directory.
#
# Background: Both projects' go.mod contains `replace` directives, which makes
#             `go install` impossible (verified 2026-05-07). The SQLite-enabled
#             Windows builds embed migration assets and require no CGO toolchain,
#             which is the safest choice on Windows. See infra/idp/README.md.
#
# Usage:
#   PowerShell> .\infra\idp\scripts\install-binaries.ps1
#   PowerShell> .\infra\idp\scripts\install-binaries.ps1 -Version 26.2.0
#   PowerShell> .\infra\idp\scripts\install-binaries.ps1 -BinDir "$env:USERPROFILE\bin"
#
# Note: This file is intentionally ASCII-only. PowerShell 5.1 (default Windows)
#       reads BOM-less UTF-8 as ANSI / CP949, which corrupts non-ASCII characters
#       and breaks the parser.

[CmdletBinding()]
param(
    [string]$Version = "26.2.0",
    [string]$BinDir = "$env:USERPROFILE\go\bin"
)

$ErrorActionPreference = "Stop"

if (-not (Test-Path $BinDir)) {
    Write-Host "BinDir '$BinDir' does not exist. Creating."
    New-Item -ItemType Directory -Force -Path $BinDir | Out-Null
}

# Warn if BinDir is not on PATH (do not modify PATH automatically).
$pathContainsBinDir = ($env:PATH -split ';') | Where-Object { $_ -ieq $BinDir }
if (-not $pathContainsBinDir) {
    Write-Warning "BinDir '$BinDir' is not on PATH. After install, either add it to PATH or invoke hydra/kratos with a fully qualified path."
}

foreach ($Name in @("hydra", "kratos")) {
    $assetName = "${Name}_${Version}-windows_sqlite_64bit.zip"
    $url = "https://github.com/ory/${Name}/releases/download/v${Version}/${assetName}"
    $zipPath = Join-Path $env:TEMP $assetName
    $extractDir = Join-Path $env:TEMP "ory-${Name}-${Version}"

    Write-Host ""
    Write-Host "[$Name] downloading $url"
    Invoke-WebRequest -Uri $url -OutFile $zipPath -UseBasicParsing

    if (Test-Path $extractDir) { Remove-Item -Recurse -Force $extractDir }
    Expand-Archive -Path $zipPath -DestinationPath $extractDir -Force

    $exe = Get-ChildItem -Path $extractDir -Filter "${Name}.exe" -Recurse | Select-Object -First 1
    if (-not $exe) {
        Write-Error "[$Name] could not find ${Name}.exe inside the zip. Asset structure may have changed."
        exit 1
    }

    $dest = Join-Path $BinDir "${Name}.exe"
    Copy-Item -Path $exe.FullName -Destination $dest -Force
    Write-Host "[$Name] installed -> $dest"

    Remove-Item -Force $zipPath
    Remove-Item -Recurse -Force $extractDir
}

Write-Host ""
Write-Host "Done. Verify with:"
Write-Host "  hydra version"
Write-Host "  kratos version"
