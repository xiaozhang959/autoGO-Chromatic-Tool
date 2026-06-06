param(
    [ValidateSet("dynamic", "static")]
    [string]$Linkage = "dynamic",
    [string]$Version = "4.13.0",
    [string]$WorkDir = (Join-Path $env:TEMP "autogo-opencv-build"),
    [int]$Parallel = [Math]::Max(1, [Environment]::ProcessorCount - 1),
    [switch]$InstallToThirdParty,
    [string]$CMakeExe
)

$ErrorActionPreference = "Stop"

function Resolve-CommandPath {
    param([string]$Name)
    $command = Get-Command $Name -ErrorAction SilentlyContinue
    if (-not $command) {
        throw "Missing required command: $Name"
    }
    return $command.Source
}

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

function Invoke-Download {
    param(
        [string]$Url,
        [string]$Output
    )
    if (Test-Path -LiteralPath $Output) {
        return
    }
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $Output) | Out-Null
    Write-Host "Downloading $Url"
    Invoke-NativeCommand -FilePath "curl.exe" -ArgumentList @("-L", "--fail", "--retry", "3", "--output", $Output, $Url) -Label "curl"
}

function Expand-ZipIfNeeded {
    param(
        [string]$ZipPath,
        [string]$Destination,
        [string]$ExpectedPath
    )
    if (Test-Path -LiteralPath $ExpectedPath) {
        return
    }
    if (Test-Path -LiteralPath $Destination) {
        Remove-Item -LiteralPath $Destination -Recurse -Force
    }
    New-Item -ItemType Directory -Force -Path $Destination | Out-Null
    Expand-Archive -LiteralPath $ZipPath -DestinationPath $Destination -Force
}

$repoRoot = Resolve-Path -LiteralPath (Join-Path $PSScriptRoot "..")
$workRoot = Join-Path $WorkDir "windows-amd64"
$downloadRoot = Join-Path $workRoot "downloads"
$sourceRoot = Join-Path $workRoot "src"
$toolsRoot = Join-Path $workRoot "tools"
New-Item -ItemType Directory -Force -Path $downloadRoot, $sourceRoot, $toolsRoot | Out-Null

$gcc = Resolve-CommandPath "gcc"
$gxx = Resolve-CommandPath "g++"
$make = Resolve-CommandPath "mingw32-make"

if (-not $CMakeExe) {
    $cmakeCommand = Get-Command cmake -ErrorAction SilentlyContinue
    if ($cmakeCommand) {
        $CMakeExe = $cmakeCommand.Source
    }
}
if (-not $CMakeExe) {
    $cmakeVersion = "3.30.5"
    $cmakeZip = Join-Path $downloadRoot "cmake-$cmakeVersion-windows-x86_64.zip"
    $cmakeExtractRoot = Join-Path $toolsRoot "cmake-$cmakeVersion"
    $cmakeExpected = Join-Path $cmakeExtractRoot "cmake-$cmakeVersion-windows-x86_64\bin\cmake.exe"
    Invoke-Download -Url "https://github.com/Kitware/CMake/releases/download/v$cmakeVersion/cmake-$cmakeVersion-windows-x86_64.zip" -Output $cmakeZip
    Expand-ZipIfNeeded -ZipPath $cmakeZip -Destination $cmakeExtractRoot -ExpectedPath $cmakeExpected
    $CMakeExe = $cmakeExpected
}
if (-not (Test-Path -LiteralPath $CMakeExe)) {
    throw "CMake executable not found: $CMakeExe"
}

$opencvZip = Join-Path $downloadRoot "opencv-$Version.zip"
$opencvExtractRoot = Join-Path $sourceRoot "opencv-$Version"
$opencvSource = Join-Path $opencvExtractRoot "opencv-$Version"
Invoke-Download -Url "https://github.com/opencv/opencv/archive/refs/tags/$Version.zip" -Output $opencvZip
Expand-ZipIfNeeded -ZipPath $opencvZip -Destination $opencvExtractRoot -ExpectedPath (Join-Path $opencvSource "CMakeLists.txt")

$shared = if ($Linkage -eq "dynamic") { "ON" } else { "OFF" }
$buildRoot = Join-Path $workRoot "build-$Linkage-$Version"
$installRoot = Join-Path $workRoot "install-$Linkage-$Version"
New-Item -ItemType Directory -Force -Path $buildRoot, $installRoot | Out-Null

$cmakeArgs = @(
    "-S", $opencvSource,
    "-B", $buildRoot,
    "-G", "MinGW Makefiles",
    "-DCMAKE_BUILD_TYPE=Release",
    "-DCMAKE_INSTALL_PREFIX=$installRoot",
    "-DCMAKE_C_COMPILER=$gcc",
    "-DCMAKE_CXX_COMPILER=$gxx",
    "-DCMAKE_MAKE_PROGRAM=$make",
    "-DBUILD_LIST=core,imgproc",
    "-DBUILD_SHARED_LIBS=$shared",
    "-DBUILD_TESTS=OFF",
    "-DBUILD_PERF_TESTS=OFF",
    "-DBUILD_EXAMPLES=OFF",
    "-DBUILD_DOCS=OFF",
    "-DBUILD_PACKAGE=OFF",
    "-DBUILD_opencv_apps=OFF",
    "-DBUILD_JAVA=OFF",
    "-DBUILD_opencv_java=OFF",
    "-DBUILD_opencv_python_bindings_generator=OFF",
    "-DBUILD_opencv_js=OFF",
    "-DBUILD_ANDROID_PROJECTS=OFF",
    "-DBUILD_WITH_DEBUG_INFO=OFF",
    "-DWITH_OPENCL=OFF",
    "-DWITH_OPENGL=OFF",
    "-DWITH_DIRECTX=OFF",
    "-DWITH_TBB=OFF",
    "-DWITH_OPENMP=OFF",
    "-DWITH_LAPACK=OFF",
    "-DWITH_EIGEN=OFF",
    "-DWITH_IPP=OFF",
    "-DWITH_ITT=OFF",
    "-DWITH_PROTOBUF=OFF",
    "-DWITH_QUIRC=OFF",
    "-DWITH_ZLIB=OFF",
    "-DWITH_PNG=OFF",
    "-DWITH_JPEG=OFF",
    "-DWITH_TIFF=OFF",
    "-DWITH_WEBP=OFF",
    "-DWITH_OPENJPEG=OFF",
    "-DWITH_JASPER=OFF",
    "-DWITH_FFMPEG=OFF",
    "-DWITH_GSTREAMER=OFF",
    "-DWITH_MSMF=OFF",
    "-DWITH_DSHOW=OFF",
    "-DBUILD_ZLIB=OFF",
    "-DBUILD_TBB=OFF",
    "-DBUILD_ITT=OFF",
    "-DBUILD_PROTOBUF=OFF"
)

if ($Linkage -eq "static") {
    $cmakeArgs += @(
        "-DCMAKE_EXE_LINKER_FLAGS=-static-libgcc -static-libstdc++",
        "-DCMAKE_SHARED_LINKER_FLAGS=-static-libgcc -static-libstdc++"
    )
}

Write-Host "Configuring OpenCV $Version ($Linkage)"
Invoke-NativeCommand -FilePath $CMakeExe -ArgumentList $cmakeArgs -Label "cmake configure"

Write-Host "Building OpenCV $Version ($Linkage) with $Parallel parallel jobs"
Invoke-NativeCommand -FilePath $CMakeExe -ArgumentList @("--build", $buildRoot, "--target", "install", "--parallel", "$Parallel") -Label "cmake build"

Write-Host "OpenCV installed to: $installRoot"

if ($InstallToThirdParty) {
    if ($Linkage -ne "dynamic") {
        throw "-InstallToThirdParty currently installs the dynamic runtime used by opencv_cgo. Static libraries are left in: $installRoot"
    }
    & (Join-Path $PSScriptRoot "setup_opencv_windows.ps1") -Source $installRoot
}
