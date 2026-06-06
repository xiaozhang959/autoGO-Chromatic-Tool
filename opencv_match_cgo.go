//go:build opencv_cgo
// +build opencv_cgo

package main

/*
#cgo CXXFLAGS: -std=c++17
#cgo windows,amd64 CXXFLAGS: -I${SRCDIR}/third_party/opencv/windows-amd64/include
#cgo windows,amd64 LDFLAGS: -lopencv_core -lopencv_imgproc
#include <stdlib.h>
#include "opencv_match_bridge.h"
*/
import "C"

import (
	"errors"
	"image"
	"unsafe"
)

func openCVImageMatchBackendName() string {
	return "OpenCV CGO"
}

func openCVImageMatchBackendAvailable() bool {
	return true
}

func runOpenCVImageMatchBackend(source, template *image.NRGBA, x1, y1, x2, y2 int, opts openCVMatchOptions) (openCVMatchResult, error) {
	if source == nil || template == nil || len(source.Pix) == 0 || len(template.Pix) == 0 {
		return openCVMatchResult{Backend: openCVImageMatchBackendName()}, errors.New("OpenCV 输入图片为空")
	}

	maxResults := opts.MaxResults
	if !opts.FindAll || maxResults <= 0 {
		maxResults = 1
	}
	points := make([]C.OpenCVMatchPoint, maxResults)
	sourceBounds := source.Bounds()
	templateBounds := template.Bounds()

	cResult := C.autogo_match_template(
		(*C.uchar)(unsafe.Pointer(&source.Pix[0])),
		C.int(sourceBounds.Dx()),
		C.int(sourceBounds.Dy()),
		C.int(source.Stride),
		(*C.uchar)(unsafe.Pointer(&template.Pix[0])),
		C.int(templateBounds.Dx()),
		C.int(templateBounds.Dy()),
		C.int(template.Stride),
		C.int(x1),
		C.int(y1),
		C.int(x2),
		C.int(y2),
		C.int(boolToCInt(opts.IsGray)),
		C.int(boolToCInt(opts.IsTransparent)),
		C.float(opts.Sim),
		C.int(boolToCInt(opts.FindAll)),
		(*C.OpenCVMatchPoint)(unsafe.Pointer(&points[0])),
		C.int(maxResults),
	)

	result := openCVMatchResult{
		BestScore: cFloatToFloat32(cResult.best_score),
		Backend:   openCVImageMatchBackendName(),
	}
	if cResult.error_code != 0 {
		return result, errors.New(C.GoString((*C.char)(unsafe.Pointer(&cResult.error[0]))))
	}

	count := int(cResult.count)
	if count > len(points) {
		count = len(points)
	}
	result.Matches = make([]openCVMatch, 0, count)
	for i := 0; i < count; i++ {
		point := points[i]
		result.Matches = append(result.Matches, openCVMatch{
			Point: image.Pt(int(point.x), int(point.y)),
			Score: cFloatToFloat32(point.score),
		})
	}
	return result, nil
}

func boolToCInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func cFloatToFloat32(value C.float) float32 {
	return float32(value)
}
