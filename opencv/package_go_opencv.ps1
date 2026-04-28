#Requires -Version 5.1
<#
.SYNOPSIS
  Build opencv/main.go into opencv/go_opencv/, copy OpenCV/MinGW DLLs to opencv/go_opencv/lib/.

.DESCRIPTION
  Windows does not load DLLs from .\lib automatically. Set PATH=%CD%\lib;%PATH% before running go_opencv.exe.

.PARAMETER BundleRoot
  Default: <repo>/opencv/windows

.EXAMPLE
  .\opencv\package_go_opencv.ps1 -OpenCVBin "D:\opencv\build\x64\mingw\bin" -MingwBin "C:\msys64\mingw64\bin"
#>
[CmdletBinding()]
param(
    [string] $BundleRoot = "",
    [string] $OpenCVBin = "",
    [string] $MingwBin = "",
    [string[]] $ExtraDllDirs = @(),
    [string] $MainRel = "opencv/main.go",
    [string] $ExeName = "go_opencv.exe"
)

$ErrorActionPreference = "Stop"

function Test-Dir([string] $p) {
    return ($p -and (Test-Path -LiteralPath $p -PathType Container))
}

function Resolve-MingwBin {
    if (Test-Dir $MingwBin) { return (Resolve-Path -LiteralPath $MingwBin).Path }
    if ($env:MINGW64 -and (Test-Dir $env:MINGW64)) { return (Resolve-Path -LiteralPath $env:MINGW64).Path }
    $gccCmd = Get-Command gcc -ErrorAction SilentlyContinue
    $gcc = if ($gccCmd) { $gccCmd.Source } else { $null }
    if ($gcc) {
        $bin = Split-Path -Parent $gcc
        if (Test-Dir $bin) { return (Resolve-Path -LiteralPath $bin).Path }
    }
    throw "MinGW bin not found. Pass -MingwBin, set env MINGW64, or ensure gcc is on PATH."
}

function Resolve-OpenCVBin {
    if (Test-Dir $OpenCVBin) { return (Resolve-Path -LiteralPath $OpenCVBin).Path }
    if ($env:OPENCV_BIN -and (Test-Dir $env:OPENCV_BIN)) { return (Resolve-Path -LiteralPath $env:OPENCV_BIN).Path }
    throw "OpenCV bin not found (directory must exist). Pass -OpenCVBin or set OPENCV_BIN. Tried -OpenCVBin: '$OpenCVBin' OPENCV_BIN: '$($env:OPENCV_BIN)'"
}

function Get-SystemDllRegex {
    return '^(KERNEL32|USER32|GDI32|GDIplus|SHELL32|OLE32|OLEAUT32|ADVAPI32|WS2_32|COMDLG32|WINMM|IMM32|OPENGL32|CRYPT32|bcrypt|NORMALIZ|RPCRT4|SETUPAPI|VERSION|WINHTTP|IPHLPAPI|dbghelp|USERENV|NETAPI32|SHLWAPI|COMCTL32|UxTheme|Secur32|ntdll|dwmapi|WTSAPI32|HID|PSAPI|bcryptprimitives|MSASN1|profapi|powrprof|UMPDC|kernel\.base|windows\.storage|dxgi|d3d11|dcomp|icuuc|icuin|MSVCRT|ucrtbase)\.dll$|^(api-ms-win-|ext-ms-win-)'
}

function Get-LinkedDllNames([string] $pePath, [string] $objdumpExe) {
    $names = New-Object System.Collections.Generic.HashSet[string] ([StringComparer]::OrdinalIgnoreCase)
    $raw = & $objdumpExe -p $pePath 2>$null
    if (-not $raw) { return $names }
    foreach ($line in $raw) {
        if ($line -match 'DLL Name:\s*(\S+\.dll)\s*$') {
            [void]$names.Add($Matches[1])
        }
    }
    return $names
}

function Find-DllPath([string] $dllName, [string[]] $searchDirs) {
    foreach ($d in $searchDirs) {
        if (-not $d) { continue }
        $c = Join-Path $d $dllName
        if (Test-Path -LiteralPath $c -PathType Leaf) { return $c }
    }
    return $null
}

# $PSScriptRoot is <repo>/opencv; repo root is one level up
$repoRoot = Split-Path -Parent $PSScriptRoot
if (-not $BundleRoot) {
    $BundleRoot = Join-Path $PSScriptRoot "windows"
}

$MingwBin = Resolve-MingwBin
$OpenCVBin = Resolve-OpenCVBin
$objdump = Join-Path $MingwBin "objdump.exe"
if (-not (Test-Path -LiteralPath $objdump)) {
    throw "objdump.exe not found: $objdump"
}

$searchDirs = @($OpenCVBin, $MingwBin) + $ExtraDllDirs | Where-Object { $_ -and (Test-Path -LiteralPath $_) } | ForEach-Object { (Resolve-Path -LiteralPath $_).Path }
$uniq = [ordered]@{}
foreach ($d in $searchDirs) { if (-not $uniq.Contains($d)) { $uniq[$d] = $true } }
$searchDirs = @($uniq.Keys)

$libDir = Join-Path $BundleRoot "lib"
New-Item -ItemType Directory -Force -Path $libDir | Out-Null
$bundleFull = (Resolve-Path -LiteralPath $BundleRoot).Path
$libFull = (Resolve-Path -LiteralPath $libDir).Path
$exeOut = Join-Path $bundleFull $ExeName

Push-Location $repoRoot
try {
    Write-Host "==> go build -tags cgo -> $exeOut"
    go build -tags cgo -ldflags="-s -w" -o $exeOut $MainRel
}
finally {
    Pop-Location
}

$sysRe = Get-SystemDllRegex
$copied = [ordered]@{}

function Try-CopyDll([string] $dllName) {
    if ($dllName -match $sysRe) { return $null }
    if ($copied.Contains($dllName)) { return $copied[$dllName] }
    $src = Find-DllPath $dllName $searchDirs
    if (-not $src) {
        Write-Warning "DLL not found in search path: $dllName"
        return $null
    }
    $dst = Join-Path $libFull $dllName
    Copy-Item -LiteralPath $src -Destination $dst -Force
    $copied[$dllName] = $dst
    Write-Host "  copy lib\$dllName"
    return $dst
}

Write-Host "==> Resolve exe dependencies -> windows\lib"
$exeDlls = Get-LinkedDllNames $exeOut $objdump
foreach ($n in $exeDlls) {
    [void](Try-CopyDll $n)
}

$round = 0
while ($round -lt 6) {
    $round++
    $names = @($copied.Keys)
    $added = 0
    foreach ($k in $names) {
        if ($k -notmatch '^(opencv_|lib)') { continue }
        $pe = Join-Path $libFull $k
        if (-not (Test-Path -LiteralPath $pe)) { continue }
        foreach ($dep in (Get-LinkedDllNames $pe $objdump)) {
            if ($copied.Contains($dep)) { continue }
            if ($dep -match $sysRe) { continue }
            if (Try-CopyDll $dep) { $added++ }
        }
    }
    if ($added -eq 0) { break }
}

foreach ($extra in Get-ChildItem -LiteralPath $OpenCVBin -Filter "opencv_videoio_ffmpeg*.dll" -File -ErrorAction SilentlyContinue) {
    [void](Try-CopyDll $extra.Name)
}

Write-Host "==> Done. cd opencv\windows, then: set PATH=%CD%\lib;%PATH% && $ExeName"
Write-Host "    $bundleFull"
