#!/usr/bin/env bash
set -euo pipefail

output="build/autoGo-image-helper-opencv-static"
opencv_root="${OPENCV_STATIC_DIR:-}"
work_dir="${OPENCV_BUILD_DIR:-${TMPDIR:-/tmp}/autogo-opencv-build}"
version="4.13.0"
arch="${GOARCH:-$(go env GOARCH 2>/dev/null || uname -m)}"
parallel="$(sysctl -n hw.ncpu 2>/dev/null || echo 2)"
go_ldflags="-s -w"
build_opencv_if_missing="0"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --output)
      output="$2"
      shift 2
      ;;
    --opencv-root)
      opencv_root="$2"
      shift 2
      ;;
    --work-dir)
      work_dir="$2"
      shift 2
      ;;
    --version)
      version="$2"
      shift 2
      ;;
    --arch)
      arch="$2"
      shift 2
      ;;
    --parallel)
      parallel="$2"
      shift 2
      ;;
    --ldflags)
      go_ldflags="$2"
      shift 2
      ;;
    --build-opencv-if-missing)
      build_opencv_if_missing="1"
      shift
      ;;
    *)
      echo "Unknown argument: $1" >&2
      exit 1
      ;;
  esac
done

case "$arch" in
  amd64|x86_64)
    go_arch="amd64"
    ;;
  arm64|aarch64)
    go_arch="arm64"
    ;;
  *)
    echo "Unsupported macOS architecture: $arch" >&2
    exit 1
    ;;
esac

repo_root="$(cd "$(dirname "$0")/.." && pwd)"
if [[ -z "$opencv_root" ]]; then
  opencv_root="$work_dir/darwin-$go_arch/install-static-$version"
fi

if [[ ! -d "$opencv_root" ]]; then
  if [[ "$build_opencv_if_missing" != "1" ]]; then
    echo "Static OpenCV not found: $opencv_root" >&2
    echo "Run scripts/build_minimal_opencv_macos.sh --linkage static --arch $go_arch first, or pass --build-opencv-if-missing." >&2
    exit 1
  fi
  bash "$repo_root/scripts/build_minimal_opencv_macos.sh" \
    --linkage static \
    --version "$version" \
    --work-dir "$work_dir" \
    --arch "$go_arch" \
    --parallel "$parallel"
fi

opencv_root="$(cd "$opencv_root" && pwd)"
include_root=""
for candidate in \
  "$opencv_root/include" \
  "$opencv_root/include/opencv4"; do
  if [[ -f "$candidate/opencv2/core.hpp" ]]; then
    include_root="$candidate"
    break
  fi
done
if [[ -z "$include_root" ]]; then
  echo "Invalid static OpenCV include directory: opencv2/core.hpp was not found." >&2
  exit 1
fi

find_static_lib() {
  local pattern="$1"
  find "$opencv_root" -type f -name "$pattern" | sort | head -n 1
}

core_lib="$(find_static_lib 'libopencv_core*.a')"
imgproc_lib="$(find_static_lib 'libopencv_imgproc*.a')"
zlib_lib="$(find_static_lib 'libzlib*.a')"
tegra_hal_lib="$(find_static_lib 'libtegra_hal*.a')"
kleidicv_hal_lib="$(find_static_lib 'libkleidicv_hal*.a')"
kleidicv_lib="$(find_static_lib 'libkleidicv.a')"
kleidicv_thread_lib="$(find_static_lib 'libkleidicv_thread*.a')"

if [[ -z "$core_lib" || -z "$imgproc_lib" ]]; then
  echo "Missing static OpenCV core/imgproc libraries in: $opencv_root" >&2
  exit 1
fi

link_root="${TMPDIR:-/tmp}/autogo-opencv-static-link-darwin-$go_arch"
rm -rf "$link_root"
mkdir -p "$link_root/lib"
cp "$core_lib" "$link_root/lib/libopencv_core.a"
cp "$imgproc_lib" "$link_root/lib/libopencv_imgproc.a"
link_flags=(-L"$link_root/lib" -lopencv_imgproc -lopencv_core)
copy_optional_static_lib() {
  local source="$1"
  local link_name="$2"
  if [[ -z "$source" ]]; then
    return
  fi
  cp "$source" "$link_root/lib/lib$link_name.a"
  link_flags+=("-l$link_name")
}
copy_optional_static_lib "$tegra_hal_lib" "tegra_hal"
copy_optional_static_lib "$kleidicv_hal_lib" "kleidicv_hal"
copy_optional_static_lib "$kleidicv_lib" "kleidicv"
copy_optional_static_lib "$kleidicv_thread_lib" "kleidicv_thread"
if [[ -n "$zlib_lib" ]]; then
  cp "$zlib_lib" "$link_root/lib/libzlib.a"
  link_flags+=(-lzlib)
fi
link_flags+=(-lc++ -lpthread)

mkdir -p "$(dirname "$repo_root/$output")"

previous_cxxflags="${CGO_CXXFLAGS:-}"
previous_ldflags="${CGO_LDFLAGS:-}"
previous_cgo_enabled="${CGO_ENABLED:-}"
previous_goos="${GOOS:-}"
previous_goarch="${GOARCH:-}"
cleanup() {
  export CGO_CXXFLAGS="$previous_cxxflags"
  export CGO_LDFLAGS="$previous_ldflags"
  export CGO_ENABLED="$previous_cgo_enabled"
  export GOOS="$previous_goos"
  export GOARCH="$previous_goarch"
}
trap cleanup EXIT

export CGO_ENABLED=1
export GOOS=darwin
export GOARCH="$go_arch"
export CGO_CXXFLAGS="-I$include_root"
printf -v joined_link_flags '%q ' "${link_flags[@]}"
export CGO_LDFLAGS="$joined_link_flags"

(
  cd "$repo_root"
  go build \
    -tags opencv_cgo \
    -ldflags "$go_ldflags" \
    -o "$output" \
    .
)

if [[ ! -f "$repo_root/$output" ]]; then
  echo "macOS static OpenCV executable was not generated: $repo_root/$output" >&2
  exit 1
fi

echo "Static macOS OpenCV build completed: $repo_root/$output"
echo "No OpenCV dylibs are required next to this executable."
