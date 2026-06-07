package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"strings"

	"golang.org/x/image/bmp"
)

const (
	openCVImageFuncFindImage          = "FindImage"
	openCVImageFuncFindImageFromImage = "FindImageFromImage"
	openCVImageFuncFindImageAll       = "FindImageAll"
	defaultOpenCVImageMaxResults      = 1000
)

type openCVMatchOptions struct {
	X1            int
	Y1            int
	X2            int
	Y2            int
	IsGray        bool
	IsTransparent bool
	Sim           float32
	FindAll       bool
	MaxResults    int
}

type openCVMatch struct {
	Point image.Point
	Score float32
}

type openCVMatchResult struct {
	Matches      []openCVMatch
	TemplateSize image.Point
	BestScore    float32
	Backend      string
}

func normalizeOpenCVImageFunctionName(name string) string {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "findimageall":
		return openCVImageFuncFindImageAll
	case "findimagefromimage":
		return openCVImageFuncFindImageFromImage
	default:
		return openCVImageFuncFindImage
	}
}

func runOpenCVImageMatch(source image.Image, templateBytes []byte, opts openCVMatchOptions) (openCVMatchResult, error) {
	if source == nil {
		return openCVMatchResult{}, fmt.Errorf("请先截图或载入待查找图片")
	}
	if len(templateBytes) == 0 {
		return openCVMatchResult{}, fmt.Errorf("请先裁剪或载入模板图片")
	}

	templateImg, err := decodeOpenCVTemplateBytes(templateBytes)
	if err != nil {
		return openCVMatchResult{}, err
	}

	sourceNRGBA := openCVImageToNRGBA(source)
	templateNRGBA := openCVImageToNRGBA(templateImg)
	templateBounds := templateNRGBA.Bounds()
	if templateBounds.Empty() {
		return openCVMatchResult{}, fmt.Errorf("模板图片为空")
	}

	x1, y1, x2, y2, ok := imageSearchBounds(sourceNRGBA, opts.X1, opts.Y1, opts.X2, opts.Y2)
	if !ok {
		return openCVMatchResult{}, fmt.Errorf("查找范围无效")
	}
	if templateBounds.Dx() > x2-x1 || templateBounds.Dy() > y2-y1 {
		return openCVMatchResult{}, fmt.Errorf("模板尺寸 %dx%d 大于查找范围 %dx%d",
			templateBounds.Dx(), templateBounds.Dy(), x2-x1, y2-y1)
	}

	if opts.Sim <= 0 {
		opts.Sim = 0.8
	}
	if opts.MaxResults <= 0 {
		opts.MaxResults = defaultOpenCVImageMaxResults
	}

	result, err := runOpenCVImageMatchBackend(sourceNRGBA, templateNRGBA, x1, y1, x2, y2, opts)
	if err != nil {
		return result, err
	}
	result.TemplateSize = image.Pt(templateBounds.Dx(), templateBounds.Dy())
	result.Backend = openCVImageMatchBackendName()
	return result, nil
}

func decodeOpenCVTemplateBytes(data []byte) (image.Image, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("模板图片为空")
	}

	if img, err := png.Decode(bytes.NewReader(data)); err == nil {
		return img, nil
	}
	if img, _, err := image.Decode(bytes.NewReader(data)); err == nil {
		return img, nil
	}
	if img, err := bmp.Decode(bytes.NewReader(data)); err == nil {
		return img, nil
	}

	return nil, fmt.Errorf("模板图片解码失败")
}

func encodeOpenCVTemplatePNG(img image.Image) ([]byte, error) {
	if img == nil {
		return nil, fmt.Errorf("模板图片为空")
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("模板图片编码失败: %v", err)
	}
	return buf.Bytes(), nil
}

func openCVImageToNRGBA(src image.Image) *image.NRGBA {
	if src == nil {
		return image.NewNRGBA(image.Rect(0, 0, 0, 0))
	}

	bounds := src.Bounds()
	dst := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			dst.Set(x, y, src.At(bounds.Min.X+x, bounds.Min.Y+y))
		}
	}
	return dst
}

func openCVMatchPoints(matches []openCVMatch) []image.Point {
	points := make([]image.Point, 0, len(matches))
	for _, match := range matches {
		points = append(points, match.Point)
	}
	return points
}

func openCVMatchHighlightRects(result openCVMatchResult) []MarkRect {
	rects := make([]MarkRect, 0, len(result.Matches))
	for _, match := range result.Matches {
		rects = append(rects, MarkRect{
			X1:    match.Point.X,
			Y1:    match.Point.Y,
			X2:    match.Point.X + result.TemplateSize.X,
			Y2:    match.Point.Y + result.TemplateSize.Y,
			Color: findTestMarkColor,
		})
	}
	return rects
}

func setOpenCVFindTestHighlightRects(viewer *ImageViewer, rects []MarkRect) {
	if viewer == nil || viewer.nodeToolOnly {
		return
	}
	viewer.findTestRects = rects
	if viewer.image != nil {
		viewer.Refresh()
	}
}
