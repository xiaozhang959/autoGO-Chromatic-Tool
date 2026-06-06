param(
    [string]$Output = "build\autoGo-image-helper-opencv.exe"
)

$ErrorActionPreference = "Stop"
$repoRoot = Resolve-Path -LiteralPath (Join-Path $PSScriptRoot "..")
$opencvRoot = Join-Path $repoRoot "third_party\opencv\windows-amd64"

foreach ($required in @(
    "include\opencv2\core.hpp",
    "include\opencv2\imgproc.hpp",
    "lib\libopencv_core.dll.a",
    "lib\libopencv_imgproc.dll.a",
    "bin\opencv_core.dll",
    "bin\opencv_imgproc.dll"
)) {
    $path = Join-Path $opencvRoot $required
    if (-not (Test-Path -LiteralPath $path)) {
        throw "缺少 OpenCV 文件: $path。请先运行 scripts/setup_opencv_windows.ps1。"
    }
}

$outPath = Join-Path $repoRoot $Output
$outDir = Split-Path -Parent $outPath
New-Item -ItemType Directory -Force -Path $outDir | Out-Null

Push-Location $repoRoot
try {
    $env:CGO_ENABLED = "1"
    go build -tags opencv_cgo -o $outPath .
} finally {
    Pop-Location
}

Copy-Item -Path (Join-Path $opencvRoot "bin\*.dll") -Destination $outDir -Force
Write-Host "构建完成: $outPath"
Write-Host "已复制 OpenCV DLL 到: $outDir"
