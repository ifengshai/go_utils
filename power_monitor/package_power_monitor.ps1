#Requires -Version 5.1
<#
.SYNOPSIS
  Build power_monitor/main.go into power_monitor/windows/power_monitor.exe (static, no DLL needed).

.EXAMPLE
  .\power_monitor\package_power_monitor.ps1
  .\power_monitor\package_power_monitor.ps1 -MingwBin "C:\msys64\mingw64\bin"
#>
[CmdletBinding()]
param(
    [string] $MingwBin = "C:\msys64\mingw64\bin"
)

$ErrorActionPreference = "Stop"

$repoRoot  = Split-Path -Parent $PSScriptRoot
$outDir    = Join-Path $PSScriptRoot "windows"
$exeOut    = Join-Path $outDir "power_monitor.exe"

New-Item -ItemType Directory -Force -Path $outDir | Out-Null

$env:PATH = "$MingwBin;$env:PATH"

Push-Location $repoRoot
try {
    Write-Host "==> Building power_monitor (static, windowsgui) -> $exeOut"
    go build -ldflags="-s -w -H=windowsgui -extldflags=-static" -o $exeOut "power_monitor/main.go"
    Write-Host "==> Done: $exeOut ($([math]::Round((Get-Item $exeOut).Length/1MB, 1)) MB)"
} finally {
    Pop-Location
}
