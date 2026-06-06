#!/usr/bin/env bash
set -euo pipefail

linkage="static"
version="4.13.0"
work_dir="${TMPDIR:-/tmp}/autogo-opencv-build"
parallel="$(sysctl -n hw.ncpu 2>/dev/null || echo 2)"
arch="${GOARCH:-$(go env GOARCH 2>/dev/null || uname -m)}"
install_to_third_party="0"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --linkage)
      linkage="$2"
      shift 2
      ;;
    --version)
      version="$2"
      shift 2
      ;;
    --work-dir)
      work_dir="$2"
      shift 2
      ;;
    --parallel)
      parallel="$2"
      shift 2
      ;;
    --arch)
      arch="$2"
      shift 2
      ;;
    --install-to-third-party)
      install_to_third_party="1"
      shift
      ;;
    *)
      echo "Unknown argument: $1" >&2
      exit 1
      ;;
  esac
done

if [[ "$linkage" != "static" && "$linkage" != "dynamic" ]]; then
  echo "--linkage must be static or dynamic" >&2
  exit 1
fi

case "$arch" in
  amd64|x86_64)
    go_arch="amd64"
    cmake_arch="x86_64"
    ;;
  arm64|aarch64)
    go_arch="arm64"
    cmake_arch="arm64"
    ;;
  *)
    echo "Unsupported macOS architecture: $arch" >&2
    exit 1
    ;;
esac

if ! command -v cmake >/dev/null 2>&1; then
  echo "Missing required command: cmake" >&2
  exit 1
fi
if ! command -v curl >/dev/null 2>&1; then
  echo "Missing required command: curl" >&2
  exit 1
fi
if ! command -v unzip >/dev/null 2>&1; then
  echo "Missing required command: unzip" >&2
  exit 1
fi

repo_root="$(cd "$(dirname "$0")/.." && pwd)"
work_root="$work_dir/darwin-$go_arch"
download_root="$work_root/downloads"
source_root="$work_root/src"
build_root="$work_root/build-$linkage-$version"
install_root="$work_root/install-$linkage-$version"
mkdir -p "$download_root" "$source_root" "$build_root" "$install_root"

opencv_zip="$download_root/opencv-$version.zip"
opencv_extract_root="$source_root/opencv-$version"
opencv_source="$opencv_extract_root/opencv-$version"

if [[ ! -f "$opencv_zip" ]]; then
  echo "Downloading OpenCV $version"
  curl -L --fail --retry 3 \
    -o "$opencv_zip" \
    "https://github.com/opencv/opencv/archive/refs/tags/$version.zip"
fi

if [[ ! -f "$opencv_source/CMakeLists.txt" ]]; then
  rm -rf "$opencv_extract_root"
  mkdir -p "$opencv_extract_root"
  unzip -q "$opencv_zip" -d "$opencv_extract_root"
fi

shared="OFF"
if [[ "$linkage" == "dynamic" ]]; then
  shared="ON"
fi

cmake_args=(
  -S "$opencv_source"
  -B "$build_root"
  -DCMAKE_BUILD_TYPE=Release
  -DCMAKE_INSTALL_PREFIX="$install_root"
  -DCMAKE_OSX_ARCHITECTURES="$cmake_arch"
  -DBUILD_LIST=core,imgproc
  -DBUILD_SHARED_LIBS="$shared"
  -DBUILD_TESTS=OFF
  -DBUILD_PERF_TESTS=OFF
  -DBUILD_EXAMPLES=OFF
  -DBUILD_DOCS=OFF
  -DBUILD_PACKAGE=OFF
  -DBUILD_opencv_apps=OFF
  -DBUILD_JAVA=OFF
  -DBUILD_opencv_java=OFF
  -DBUILD_opencv_python_bindings_generator=OFF
  -DBUILD_opencv_js=OFF
  -DBUILD_WITH_DEBUG_INFO=OFF
  -DWITH_OPENCL=OFF
  -DWITH_OPENGL=OFF
  -DWITH_COCOA=OFF
  -DWITH_AVFOUNDATION=OFF
  -DWITH_QT=OFF
  -DWITH_GTK=OFF
  -DWITH_TBB=OFF
  -DWITH_OPENMP=OFF
  -DWITH_LAPACK=OFF
  -DWITH_EIGEN=OFF
  -DWITH_IPP=OFF
  -DWITH_ITT=OFF
  -DWITH_PROTOBUF=OFF
  -DWITH_QUIRC=OFF
  -DWITH_PNG=OFF
  -DWITH_JPEG=OFF
  -DWITH_TIFF=OFF
  -DWITH_WEBP=OFF
  -DWITH_OPENJPEG=OFF
  -DWITH_JASPER=OFF
  -DWITH_FFMPEG=OFF
  -DWITH_GSTREAMER=OFF
  -DBUILD_TBB=OFF
  -DBUILD_ITT=OFF
  -DBUILD_PROTOBUF=OFF
  -DWITH_ZLIB=ON
  -DBUILD_ZLIB=ON
)

if [[ "$go_arch" == "amd64" ]]; then
  cmake_args+=(
    -DCMAKE_SYSTEM_PROCESSOR="$cmake_arch"
    -DOPENCV_WORKAROUND_CMAKE_20989=ON
    -DWITH_CAROTENE=OFF
    -DWITH_KLEIDICV=OFF
  )
fi

echo "Configuring OpenCV $version ($linkage, darwin-$go_arch)"
cmake "${cmake_args[@]}"

echo "Building OpenCV $version ($linkage, darwin-$go_arch) with $parallel parallel jobs"
cmake --build "$build_root" --target install --parallel "$parallel"

echo "OpenCV installed to: $install_root"

if [[ "$install_to_third_party" == "1" ]]; then
  target_root="$repo_root/third_party/opencv/darwin-$go_arch"
  rm -rf "$target_root"
  mkdir -p "$target_root"
  cp -R "$install_root"/. "$target_root"/
  echo "OpenCV copied to: $target_root"
fi
