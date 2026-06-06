OpenCV Windows AMD64 bundle placeholder.

Purpose:
  This directory is the project-local OpenCV dependency used by the
  opencv_cgo build tag. End users should receive the built EXE together
  with DLLs from bin/, so they do not need to install OpenCV manually.

Expected layout:
  include/opencv2/core.hpp
  include/opencv2/imgproc.hpp
  lib/libopencv_core.dll.a
  lib/libopencv_imgproc.dll.a
  bin/libopencv_core*.dll or bin/opencv_core*.dll
  bin/libopencv_imgproc*.dll or bin/opencv_imgproc*.dll

Populate this folder with:
  powershell -ExecutionPolicy Bypass -File scripts/setup_opencv_windows.ps1 -Source <OpenCV-MinGW-root>

Build a distributable app with:
  powershell -ExecutionPolicy Bypass -File scripts/build_with_opencv_windows.ps1

Build a smaller local OpenCV bundle from source with:
  powershell -ExecutionPolicy Bypass -File scripts/build_minimal_opencv_windows.ps1 -Linkage dynamic -InstallToThirdParty

Build static OpenCV libraries for later single-EXE work with:
  powershell -ExecutionPolicy Bypass -File scripts/build_minimal_opencv_windows.ps1 -Linkage static

Build a single EXE with static OpenCV after building static libraries with:
  powershell -ExecutionPolicy Bypass -File scripts/build_with_static_opencv_windows.ps1

Notes:
  - The import libraries are normalized to unversioned names so cgo can link
    with -lopencv_core and -lopencv_imgproc.
  - Keep the OpenCV license files from your binary distribution when shipping
    a release package.
