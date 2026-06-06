param(
    [string]$Output = "build\autoGo-image-helper-opencv.exe"
)

$ErrorActionPreference = "Stop"

$repoRoot = Resolve-Path -LiteralPath (Join-Path $PSScriptRoot "..")
$opencvRoot = Join-Path $repoRoot "third_party\opencv\windows-amd64"
$opencvLib = Join-Path $opencvRoot "lib"
$opencvBin = Join-Path $opencvRoot "bin"

foreach ($required in @(
    "include\opencv2\core.hpp",
    "include\opencv2\imgproc.hpp",
    "lib\libopencv_core.dll.a",
    "lib\libopencv_imgproc.dll.a"
)) {
    $path = Join-Path $opencvRoot $required
    if (-not (Test-Path -LiteralPath $path)) {
        throw "Missing OpenCV file: $path"
    }
}

foreach ($runtimePatterns in @(
    @("libopencv_core*.dll", "opencv_core*.dll"),
    @("libopencv_imgproc*.dll", "opencv_imgproc*.dll")
)) {
    $found = $false
    foreach ($pattern in $runtimePatterns) {
        if (Get-ChildItem -LiteralPath $opencvBin -Filter $pattern -File -ErrorAction SilentlyContinue | Select-Object -First 1) {
            $found = $true
            break
        }
    }
    if (-not $found) {
        throw "Missing OpenCV runtime DLL matching: $($runtimePatterns -join ', ')"
    }
}

$outPath = Join-Path $repoRoot $Output
$outDir = Split-Path -Parent $outPath
New-Item -ItemType Directory -Force -Path $outDir | Out-Null

# MinGW ld can fail on non-ASCII -L paths. Copy import libs to an ASCII temp dir
# and expose it via LIBRARY_PATH, which only affects library search paths.
$asciiLinkRoot = Join-Path $env:TEMP "autogo-opencv-link"
if (Test-Path -LiteralPath $asciiLinkRoot) {
    Remove-Item -LiteralPath $asciiLinkRoot -Recurse -Force
}
New-Item -ItemType Directory -Force -Path (Join-Path $asciiLinkRoot "lib") | Out-Null
Copy-Item -LiteralPath (Join-Path $opencvLib "libopencv_core.dll.a") -Destination (Join-Path $asciiLinkRoot "lib\libopencv_core.dll.a") -Force
Copy-Item -LiteralPath (Join-Path $opencvLib "libopencv_imgproc.dll.a") -Destination (Join-Path $asciiLinkRoot "lib\libopencv_imgproc.dll.a") -Force

$previousLibraryPath = $env:LIBRARY_PATH
$previousPath = $env:PATH
Push-Location $repoRoot
try {
    $env:CGO_ENABLED = "1"
    $env:LIBRARY_PATH = (Join-Path $asciiLinkRoot "lib")
    $env:PATH = "$opencvBin;$env:PATH"
    go build -tags opencv_cgo -o $outPath .
} finally {
    $env:LIBRARY_PATH = $previousLibraryPath
    $env:PATH = $previousPath
    Pop-Location
}

Copy-Item -Path (Join-Path $opencvBin "*.dll") -Destination $outDir -Force
Write-Host "Build completed: $outPath"
Write-Host "OpenCV DLLs copied to: $outDir"
