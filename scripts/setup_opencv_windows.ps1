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

function Resolve-CommandDir {
    param([string]$Name)
    $command = Get-Command $Name -ErrorAction SilentlyContinue
    if (-not $command) {
        return $null
    }
    return Split-Path -Parent $command.Source
}

function Clear-Directory {
    param([string]$Path)
    if (Test-Path -LiteralPath $Path) {
        Get-ChildItem -LiteralPath $Path -Force | Remove-Item -Recurse -Force
        return
    }
    New-Item -ItemType Directory -Force -Path $Path | Out-Null
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
    throw "Missing file: $($Patterns -join ', ')"
}

function Copy-MatchingFiles {
    param(
        [string[]]$Patterns,
        [string]$DestinationDir,
        [switch]$Required
    )
    $copied = @()
    foreach ($pattern in $Patterns) {
        $matches = Get-ChildItem -Path $pattern -File -ErrorAction SilentlyContinue
        foreach ($match in $matches) {
            Copy-Item -LiteralPath $match.FullName -Destination (Join-Path $DestinationDir $match.Name) -Force
            $copied += $match.FullName
        }
    }
    if ($Required -and $copied.Count -eq 0) {
        throw "Missing file: $($Patterns -join ', ')"
    }
    return $copied
}

$repoRoot = Resolve-Path -LiteralPath (Join-Path $PSScriptRoot "..")
if (-not $Source) {
    throw "Pass -Source <OpenCV-MinGW-root>, or set OPENCV_DIR first."
}
$sourceRoot = Resolve-Path -LiteralPath $Source

$includeRoot = Resolve-FirstExistingDir @(
    (Join-Path $sourceRoot "include"),
    (Join-Path $sourceRoot "mingw64\include"),
    (Join-Path $sourceRoot "x64\mingw\include")
)
$opencv2Root = $null
foreach ($candidate in @(
    (Join-Path $includeRoot "opencv2"),
    (Join-Path $includeRoot "opencv4\opencv2")
)) {
    if ($candidate -and (Test-Path -LiteralPath (Join-Path $candidate "core.hpp"))) {
        $opencv2Root = $candidate
        break
    }
}
if (-not $includeRoot -or -not $opencv2Root) {
    throw "Invalid OpenCV include directory: opencv2/core.hpp was not found."
}

$libRoot = Resolve-FirstExistingDir @(
    (Join-Path $sourceRoot "lib"),
    (Join-Path $sourceRoot "mingw64\lib"),
    (Join-Path $sourceRoot "x64\mingw\lib")
)
if (-not $libRoot) {
    throw "Invalid OpenCV lib directory."
}

$binRoot = Resolve-FirstExistingDir @(
    (Join-Path $sourceRoot "bin"),
    (Join-Path $sourceRoot "mingw64\bin"),
    (Join-Path $sourceRoot "x64\mingw\bin")
)
if (-not $binRoot) {
    throw "Invalid OpenCV bin directory."
}

$targetRoot = Join-Path $repoRoot "third_party\opencv\windows-amd64"
$targetInclude = Join-Path $targetRoot "include"
$targetLib = Join-Path $targetRoot "lib"
$targetBin = Join-Path $targetRoot "bin"

New-Item -ItemType Directory -Force -Path $targetRoot, $targetInclude | Out-Null
Clear-Directory -Path $targetLib
Clear-Directory -Path $targetBin
Copy-DirectoryClean -From $opencv2Root -To (Join-Path $targetInclude "opencv2")

Copy-FirstFile -Patterns @(
    (Join-Path $libRoot "libopencv_core*.dll.a"),
    (Join-Path $libRoot "libopencv_core*.a")
) -Destination (Join-Path $targetLib "libopencv_core.dll.a") | Out-Null
Copy-FirstFile -Patterns @(
    (Join-Path $libRoot "libopencv_imgproc*.dll.a"),
    (Join-Path $libRoot "libopencv_imgproc*.a")
) -Destination (Join-Path $targetLib "libopencv_imgproc.dll.a") | Out-Null

Copy-MatchingFiles -Patterns @(
    (Join-Path $binRoot "libopencv_core*.dll"),
    (Join-Path $binRoot "opencv_core*.dll")
) -DestinationDir $targetBin -Required | Out-Null
Copy-MatchingFiles -Patterns @(
    (Join-Path $binRoot "libopencv_imgproc*.dll"),
    (Join-Path $binRoot "opencv_imgproc*.dll")
) -DestinationDir $targetBin -Required | Out-Null

foreach ($dependencyDll in @(
    "libopenblas.dll",
    "libtbb12.dll",
    "zlib1.dll",
    "libgfortran-5.dll",
    "libquadmath-0.dll",
    "libgomp-1.dll"
)) {
    $dependencyPath = Get-ChildItem -Path $sourceRoot -Recurse -Filter $dependencyDll -File -ErrorAction SilentlyContinue |
        Select-Object -First 1 -ExpandProperty FullName
    if ($dependencyPath) {
        Copy-Item -LiteralPath $dependencyPath -Destination (Join-Path $targetBin $dependencyDll) -Force
    }
}

$runtimeSearchDirs = @(
    $binRoot,
    (Resolve-CommandDir "gcc"),
    (Resolve-CommandDir "g++")
) + $env:PATH.Split([IO.Path]::PathSeparator)

foreach ($runtimeDll in @("libgcc_s_seh-1.dll", "libstdc++-6.dll", "libwinpthread-1.dll")) {
    $runtimePath = $runtimeSearchDirs |
        Where-Object { $_ } |
        ForEach-Object { Join-Path $_ $runtimeDll } |
        Where-Object { Test-Path -LiteralPath $_ } |
        Select-Object -First 1
    if ($runtimePath) {
        Copy-Item -LiteralPath $runtimePath -Destination (Join-Path $targetBin $runtimeDll) -Force
    }
}

Write-Host "OpenCV copied to: $targetRoot"
Write-Host "Next: powershell -ExecutionPolicy Bypass -File scripts/build_with_opencv_windows.ps1"
