package main

import (
	"image"
	"strings"
	"testing"
)

func TestNormalizeOpenCVImageFunctionName(t *testing.T) {
	tests := map[string]string{
		"":                   openCVImageFuncFindImage,
		"FindImage":          openCVImageFuncFindImage,
		"findimageall":       openCVImageFuncFindImageAll,
		"FindImageFromImage": openCVImageFuncFindImageFromImage,
		"FindMultiColors":    openCVImageFuncFindImage,
	}

	for input, want := range tests {
		if got := normalizeOpenCVImageFunctionName(input); got != want {
			t.Fatalf("normalizeOpenCVImageFunctionName(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestBuildOpenCVImageTestCodeUsesNewDocSignature(t *testing.T) {
	opts := openCVMatchOptions{
		X1:            1,
		Y1:            2,
		X2:            30,
		Y2:            40,
		IsGray:        true,
		IsTransparent: true,
		Sim:           0.85,
	}

	code := buildOpenCVImageTestCode(openCVImageFuncFindImage, opts, 3, "button.png")
	for _, want := range []string{
		`os.ReadFile("button.png")`,
		`opencv.FindImage(1, 2, 30, 40, &templateBytes, true, true, 0.85, 3)`,
	} {
		if !strings.Contains(code, want) {
			t.Fatalf("generated code missing %q:\n%s", want, code)
		}
	}

	code = buildOpenCVImageTestCode(openCVImageFuncFindImageFromImage, opts, 0, "button.png")
	if !strings.Contains(code, `opencv.FindImageFromImage(img, &templateBytes, true, true, 0.85)`) {
		t.Fatalf("FindImageFromImage code uses wrong signature:\n%s", code)
	}
}

func TestOpenCVMatchHighlightRectsUseTemplateSize(t *testing.T) {
	result := openCVMatchResult{
		Matches: []openCVMatch{
			{Point: image.Pt(10, 20), Score: 0.99},
		},
		TemplateSize: image.Pt(7, 9),
	}

	rects := openCVMatchHighlightRects(result)
	if len(rects) != 1 {
		t.Fatalf("len(rects) = %d, want 1", len(rects))
	}
	got := rects[0]
	if got.X1 != 10 || got.Y1 != 20 || got.X2 != 17 || got.Y2 != 29 {
		t.Fatalf("rect = %+v, want (10,20)-(17,29)", got)
	}
}
