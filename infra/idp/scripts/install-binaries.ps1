# install-binaries.ps1
#
# 목적: Ory Hydra + Ory Kratos 의 Windows binary (SQLite 포함) 를 GitHub release
#       에서 다운로드해 PATH 가 잡히는 디렉터리에 배치한다.
#
# 배경: 두 프로젝트의 go.mod 가 `replace` 지시문을 포함해 `go install` 이 차단된다
#       (2026-05-07 확인). 따라서 release binary 다운로드를 1차 경로로 사용한다.
#       SQLite 변형은 CGO 의존이 없고 embed 된 migration 자산을 포함해 Windows
#       에서 가장 안전한 선택이다.
#
# 사용:
#   PowerShell> .\infra\idp\scripts\install-binaries.ps1
#   PowerShell> .\infra\idp\scripts\install-binaries.ps1 -Version 26.2.0
#   PowerShell> .\infra\idp\scripts\install-binaries.ps1 -BinDir "$env:USERPROFILE\bin"

[CmdletBinding()]
param(
    [string]$Version = "26.2.0",
    [string]$BinDir = "$env:USERPROFILE\go\bin"
)

$ErrorActionPreference = "Stop"

if (-not (Test-Path $BinDir)) {
    Write-Host "BinDir '$BinDir' 가 없어 새로 만든다."
    New-Item -ItemType Directory -Force -Path $BinDir | Out-Null
}

# PATH 에 BinDir 가 있는지 확인 (경고만, 강제 추가는 하지 않음)
$pathContainsBinDir = ($env:PATH -split ';') | Where-Object { $_ -ieq $BinDir }
if (-not $pathContainsBinDir) {
    Write-Warning "BinDir '$BinDir' 가 현재 PATH 에 없다. hydra/kratos 명령이 인식되지 않으면 PATH 에 추가하거나 fully qualified 경로로 호출해야 한다."
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
        Write-Error "[$Name] zip 안에서 ${Name}.exe 를 찾지 못함. zip 구조 변경 가능성."
        exit 1
    }

    $dest = Join-Path $BinDir "${Name}.exe"
    Copy-Item -Path $exe.FullName -Destination $dest -Force
    Write-Host "[$Name] installed -> $dest"

    Remove-Item -Force $zipPath
    Remove-Item -Recurse -Force $extractDir
}

Write-Host ""
Write-Host "Done. 확인:"
Write-Host "  hydra version"
Write-Host "  kratos version"
