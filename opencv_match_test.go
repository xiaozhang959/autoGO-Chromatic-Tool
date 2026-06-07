package main

import (
	"image"
	"strings"
	"testing"

	"fyne.io/fyne/v2/widget"
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

func TestOpenCVEffectiveMatchThreshold(t *testing.T) {
	if got := openCVEffectiveMatchThreshold(0.6); got != 0.8 {
		t.Fatalf("openCVEffectiveMatchThreshold(0.6) = %v, want 0.8", got)
	}
	if got := openCVEffectiveMatchThreshold(0); got != 0.9 {
		t.Fatalf("openCVEffectiveMatchThreshold(0) = %v, want 0.9", got)
	}
}

func TestFormatOpenCVImageTestResultShowsThreshold(t *testing.T) {
	result := openCVMatchResult{
		BestScore:    0.791695,
		TemplateSize: image.Pt(187, 219),
		Backend:      "OpenCV CGO",
	}

	text := formatOpenCVImageTestResult(openCVImageFuncFindImage, result, 0.6)
	for _, want := range []string{
		"sim: 0.600000",
		"threshold: 0.800000 (0.5 + sim * 0.5)",
		"bestScore: 0.791695",
		"result: -1,-1",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("formatted result missing %q:\n%s", want, text)
		}
	}
}

func TestOpenCVOptionsFromEntriesParsesCommaRange(t *testing.T) {
	rangeEntry := widget.NewEntry()
	rangeEntry.SetText("10, 20, 30, 40")
	simEntry := widget.NewEntry()
	simEntry.SetText("0.88")
	grayCheck := widget.NewCheck("", nil)
	grayCheck.SetChecked(true)
	transparentCheck := widget.NewCheck("", nil)

	opts, err := openCVOptionsFromEntries(openCVImageFuncFindImageAll, rangeEntry, simEntry, grayCheck, transparentCheck)
	if err != nil {
		t.Fatalf("openCVOptionsFromEntries returned error: %v", err)
	}
	if opts.X1 != 10 || opts.Y1 != 20 || opts.X2 != 30 || opts.Y2 != 40 {
		t.Fatalf("range mismatch: %+v", opts)
	}
	if !opts.IsGray || opts.IsTransparent || opts.Sim != 0.88 || !opts.FindAll {
		t.Fatalf("options mismatch: %+v", opts)
	}

	rangeEntry.SetText("10,20,30")
	if _, err := openCVOptionsFromEntries(openCVImageFuncFindImage, rangeEntry, simEntry, grayCheck, transparentCheck); err == nil {
		t.Fatal("expected error for invalid range format")
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
