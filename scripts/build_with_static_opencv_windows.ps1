param(
    [string]$Output = "build\autoGo-image-helper-opencv-static.exe",
    [string]$OpenCVRoot = $env:OPENCV_STATIC_DIR,
    [string]$Version = "4.13.0",
    [int]$Parallel = [Math]::Max(1, [Environment]::ProcessorCount - 1),
    [switch]$BuildOpenCVIfMissing
)

$ErrorActionPreference = "Stop"

function Invoke-NativeCommand {
    param(
        [string]$FilePath,
        [string[]]$ArgumentList,
        [string]$Label = $FilePath
    )
    & $FilePath @ArgumentList
    if ($LASTEXITCODE -ne 0) {
        throw "$Label failed with exit code $LASTEXITCODE"
    }
}

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

function Copy-FirstStaticLib {
    param(
        [string]$SourceDir,
        [string]$Pattern,
        [string]$Destination
    )
    $match = Get-ChildItem -LiteralPath $SourceDir -Filter $Pattern -File -ErrorAction SilentlyContinue |
        Where-Object { $_.Name -notlike "*.dll.a" } |
        Select-Object -First 1
    if (-not $match) {
        throw "Missing static library matching $Pattern in $SourceDir"
    }
    Copy-Item -LiteralPath $match.FullName -Destination $Destination -Force
}

$repoRoot = Resolve-Path -LiteralPath (Join-Path $PSScriptRoot "..")
if (-not $OpenCVRoot) {
    $OpenCVRoot = Join-Path $env:TEMP "autogo-opencv-build\windows-amd64\install-static-$Version"
}

if (-not (Test-Path -LiteralPath $OpenCVRoot -PathType Container)) {
    if (-not $BuildOpenCVIfMissing) {
        throw "Static OpenCV not found: $OpenCVRoot. Run scripts/build_minimal_opencv_windows.ps1 -Linkage static first, or pass -BuildOpenCVIfMissing."
    }
    & (Join-Path $PSScriptRoot "build_minimal_opencv_windows.ps1") -Linkage static -Version $Version -Parallel $Parallel
}
$OpenCVRoot = (Resolve-Path -LiteralPath $OpenCVRoot).Path

$includeRoot = Resolve-FirstExistingDir @(
    (Join-Path $OpenCVRoot "include"),
    (Join-Path $OpenCVRoot "mingw64\include"),
    (Join-Path $OpenCVRoot "x64\mingw\include")
)
if (-not $includeRoot -or -not (Test-Path -LiteralPath (Join-Path $includeRoot "opencv2\core.hpp"))) {
    throw "Invalid static OpenCV include directory: opencv2/core.hpp was not found."
}

$libRoot = Resolve-FirstExistingDir @(
    (Join-Path $OpenCVRoot "x64\mingw\staticlib"),
    (Join-Path $OpenCVRoot "mingw64\staticlib"),
    (Join-Path $OpenCVRoot "staticlib"),
    (Join-Path $OpenCVRoot "lib")
)
if (-not $libRoot) {
    throw "Invalid static OpenCV lib directory."
}

$outPath = Join-Path $repoRoot $Output
$outDir = Split-Path -Parent $outPath
New-Item -ItemType Directory -Force -Path $outDir | Out-Null

# Keep import libraries in an ASCII path; MinGW ld can fail on non-ASCII repo paths.
$asciiLinkRoot = Join-Path $env:TEMP "autogo-opencv-static-link"
if (Test-Path -LiteralPath $asciiLinkRoot) {
    Remove-Item -LiteralPath $asciiLinkRoot -Recurse -Force
}
New-Item -ItemType Directory -Force -Path (Join-Path $asciiLinkRoot "lib") | Out-Null
Copy-FirstStaticLib -SourceDir $libRoot -Pattern "libopencv_core*.a" -Destination (Join-Path $asciiLinkRoot "lib\libopencv_core.a")
Copy-FirstStaticLib -SourceDir $libRoot -Pattern "libopencv_imgproc*.a" -Destination (Join-Path $asciiLinkRoot "lib\libopencv_imgproc.a")
Copy-FirstStaticLib -SourceDir $libRoot -Pattern "libzlib*.a" -Destination (Join-Path $asciiLinkRoot "lib\libzlib.a")

$staleRuntimePatterns = @(
    "libopencv*.dll",
    "opencv_*.dll",
    "libopenblas.dll",
    "libtbb*.dll",
    "zlib1.dll",
    "libgcc_s*.dll",
    "libstdc++-6.dll",
    "libgfortran*.dll",
    "libquadmath*.dll",
    "libgomp*.dll",
    "libwinpthread-1.dll"
)
foreach ($pattern in $staleRuntimePatterns) {
    Get-ChildItem -LiteralPath $outDir -Filter $pattern -File -ErrorAction SilentlyContinue |
        Remove-Item -Force
}

$previousLibraryPath = $env:LIBRARY_PATH
$previousCGOFlags = $env:CGO_LDFLAGS
$previousCXXFlags = $env:CGO_CXXFLAGS
$previousPath = $env:PATH
Push-Location $repoRoot
try {
    $env:CGO_ENABLED = "1"
    $env:LIBRARY_PATH = (Join-Path $asciiLinkRoot "lib")
    $env:CGO_CXXFLAGS = "-I$includeRoot"
    $env:CGO_LDFLAGS = "-lopencv_imgproc -lopencv_core -lzlib -lole32 -luuid -lws2_32 -lpsapi -lwinpthread"
    $compilerDir = Resolve-CommandDir "gcc"
    if ($compilerDir) {
        $env:PATH = "$compilerDir;$env:PATH"
    }
    $goArgs = @(
        "build",
        "-tags", "opencv_cgo",
        "-ldflags", '-linkmode external -extldflags "-static"',
        "-o", $outPath,
        "."
    )
    Invoke-NativeCommand -FilePath "go" -ArgumentList $goArgs -Label "go build"
} finally {
    $env:LIBRARY_PATH = $previousLibraryPath
    $env:CGO_LDFLAGS = $previousCGOFlags
    $env:CGO_CXXFLAGS = $previousCXXFlags
    $env:PATH = $previousPath
    Pop-Location
}

Write-Host "Static build completed: $outPath"
Write-Host "No OpenCV DLLs are required next to this EXE."
