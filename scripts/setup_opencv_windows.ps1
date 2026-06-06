param(
    [string]$Source = $env:OPENCV_DIR
)

$ErrorActionPreference = "Stop"

function Resolve-FirstExistingDir {
    param([string[]]$Candidates)
    foreach ($candidate in $Candidates) {
        if ($candidate -and (Test-Path -LiteralPath $candidate -PathType Container)) {
            return (Resolve-Path -LiteralPath $candidate).Path
        }
    }
    return $null
}

function Copy-DirectoryClean {
    param(
        [string]$From,
        [string]$To
    )
    if (Test-Path -LiteralPath $To) {
        Remove-Item -LiteralPath $To -Recurse -Force
    }
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $To) | Out-Null
    Copy-Item -LiteralPath $From -Destination $To -Recurse -Force
}

function Copy-FirstFile {
    param(
        [string[]]$Patterns,
        [string]$Destination
    )
    foreach ($pattern in $Patterns) {
        $match = Get-ChildItem -Path $pattern -File -ErrorAction SilentlyContinue | Select-Object -First 1
        if ($match) {
            Copy-Item -LiteralPath $match.FullName -Destination $Destination -Force
            return $match.FullName
        }
    }
    throw "未找到文件: $($Patterns -join ', ')"
}

$repoRoot = Resolve-Path -LiteralPath (Join-Path $PSScriptRoot "..")
if (-not $Source) {
    throw "请传入 -Source <OpenCV-MinGW-root>，或先设置 OPENCV_DIR。"
}
$sourceRoot = Resolve-Path -LiteralPath $Source

$includeRoot = Resolve-FirstExistingDir @(
    (Join-Path $sourceRoot "include"),
    (Join-Path $sourceRoot "mingw64\include"),
    (Join-Path $sourceRoot "x64\mingw\include")
)
if (-not $includeRoot -or -not (Test-Path -LiteralPath (Join-Path $includeRoot "opencv2\core.hpp"))) {
    throw "OpenCV include 目录无效，未找到 opencv2/core.hpp。"
}

$libRoot = Resolve-FirstExistingDir @(
    (Join-Path $sourceRoot "lib"),
    (Join-Path $sourceRoot "mingw64\lib"),
    (Join-Path $sourceRoot "x64\mingw\lib")
)
if (-not $libRoot) {
    throw "OpenCV lib 目录无效。"
}

$binRoot = Resolve-FirstExistingDir @(
    (Join-Path $sourceRoot "bin"),
    (Join-Path $sourceRoot "mingw64\bin"),
    (Join-Path $sourceRoot "x64\mingw\bin")
)
if (-not $binRoot) {
    throw "OpenCV bin 目录无效。"
}

$targetRoot = Join-Path $repoRoot "third_party\opencv\windows-amd64"
$targetInclude = Join-Path $targetRoot "include"
$targetLib = Join-Path $targetRoot "lib"
$targetBin = Join-Path $targetRoot "bin"

New-Item -ItemType Directory -Force -Path $targetRoot, $targetLib, $targetBin | Out-Null
Copy-DirectoryClean -From (Join-Path $includeRoot "opencv2") -To (Join-Path $targetInclude "opencv2")

Copy-FirstFile -Patterns @(
    (Join-Path $libRoot "libopencv_core*.dll.a"),
    (Join-Path $libRoot "libopencv_core*.a")
) -Destination (Join-Path $targetLib "libopencv_core.dll.a") | Out-Null
Copy-FirstFile -Patterns @(
    (Join-Path $libRoot "libopencv_imgproc*.dll.a"),
    (Join-Path $libRoot "libopencv_imgproc*.a")
) -Destination (Join-Path $targetLib "libopencv_imgproc.dll.a") | Out-Null

Copy-FirstFile -Patterns @(
    (Join-Path $binRoot "libopencv_core*.dll"),
    (Join-Path $binRoot "opencv_core*.dll")
) -Destination (Join-Path $targetBin "opencv_core.dll") | Out-Null
Copy-FirstFile -Patterns @(
    (Join-Path $binRoot "libopencv_imgproc*.dll"),
    (Join-Path $binRoot "opencv_imgproc*.dll")
) -Destination (Join-Path $targetBin "opencv_imgproc.dll") | Out-Null

foreach ($runtimeDll in @("libgcc_s_seh-1.dll", "libstdc++-6.dll", "libwinpthread-1.dll")) {
    $runtimePath = $env:PATH.Split([IO.Path]::PathSeparator) |
        ForEach-Object { Join-Path $_ $runtimeDll } |
        Where-Object { Test-Path -LiteralPath $_ } |
        Select-Object -First 1
    if ($runtimePath) {
        Copy-Item -LiteralPath $runtimePath -Destination (Join-Path $targetBin $runtimeDll) -Force
    }
}

Write-Host "OpenCV 已复制到 $targetRoot"
Write-Host "下一步运行: powershell -ExecutionPolicy Bypass -File scripts/build_with_opencv_windows.ps1"
