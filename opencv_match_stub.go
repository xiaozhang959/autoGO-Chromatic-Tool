//go:build !opencv_cgo
// +build !opencv_cgo

package main

import (
	"fmt"
	"image"
)

func openCVImageMatchBackendName() string {
	return "OpenCV CGO 未启用"
}

func openCVImageMatchBackendAvailable() bool {
	return false
}

func runOpenCVImageMatchBackend(source, template *image.NRGBA, x1, y1, x2, y2 int, opts openCVMatchOptions) (openCVMatchResult, error) {
	return openCVMatchResult{Backend: openCVImageMatchBackendName()}, fmt.Errorf(
		"真实 OpenCV 后端未启用；请安装 OpenCV C++ 开发库后使用 -tags opencv_cgo 构建")
}
