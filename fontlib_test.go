package main

import (
	"image"
	"image/color"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

func newSolidFontTestImage(w, h int, c color.NRGBA) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetNRGBA(x, y, c)
		}
	}
	return img
}

func fillFontTestRect(img *image.NRGBA, rect image.Rectangle, c color.NRGBA) {
	rect = rect.Intersect(img.Bounds())
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			img.SetNRGBA(x, y, c)
		}
	}
}

func assertColorClose(t *testing.T, got, want color.NRGBA, maxDistanceSq int) {
	t.Helper()
	if fontColorDistanceSq(got, want) > maxDistanceSq {
		t.Fatalf("color mismatch: got %v, want near %v", got, want)
	}
}

func assertRectContains(t *testing.T, outer, inner image.Rectangle) {
	t.Helper()
	if !inner.Min.In(outer) || !image.Pt(inner.Max.X-1, inner.Max.Y-1).In(outer) {
		t.Fatalf("rect %v does not contain %v", outer, inner)
	}
}

func TestAutoPreprocessFontImageEstimatesForeground(t *testing.T) {
	cases := []struct {
		name string
		bg   color.NRGBA
		fg   color.NRGBA
	}{
		{name: "light background dark text", bg: color.NRGBA{245, 245, 245, 255}, fg: color.NRGBA{5, 5, 5, 255}},
		{name: "dark background light text", bg: color.NRGBA{12, 12, 12, 255}, fg: color.NRGBA{240, 240, 240, 255}},
		{name: "colored text", bg: color.NRGBA{250, 250, 250, 255}, fg: color.NRGBA{220, 20, 40, 255}},
	}

	textRect := image.Rect(6, 4, 17, 11)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			img := newSolidFontTestImage(28, 18, tc.bg)
			fillFontTestRect(img, textRect, tc.fg)

			result, ok := autoPreprocessFontImage(img)
			if !ok {
				t.Fatal("auto preprocess returned false")
			}
			assertColorClose(t, result.Foreground, tc.fg, 20*20)
			assertRectContains(t, result.CropRect, textRect)
			if result.CropRect == img.Bounds() {
				t.Fatalf("expected crop rect smaller than full image, got %v", result.CropRect)
			}
			if result.ForegroundPixels != textRect.Dx()*textRect.Dy() {
				t.Fatalf("foreground pixels = %d, want %d", result.ForegroundPixels, textRect.Dx()*textRect.Dy())
			}
		})
	}
}

func TestAutoPreprocessFontImageRejectsEmptyBackground(t *testing.T) {
	img := newSolidFontTestImage(20, 10, color.NRGBA{255, 255, 255, 255})
	if _, ok := autoPreprocessFontImage(img); ok {
		t.Fatal("expected empty background image to be rejected")
	}
}

func TestAutoPreprocessFontImageClampsBoundaryCrop(t *testing.T) {
	img := newSolidFontTestImage(12, 10, color.NRGBA{255, 255, 255, 255})
	textRect := image.Rect(0, 0, 4, 5)
	fillFontTestRect(img, textRect, color.NRGBA{0, 0, 0, 255})

	result, ok := autoPreprocessFontImage(img)
	if !ok {
		t.Fatal("auto preprocess returned false")
	}
	assertRectContains(t, result.CropRect, textRect)
	if result.CropRect.Min.X != 0 || result.CropRect.Min.Y != 0 {
		t.Fatalf("crop rect should be clamped to image origin, got %v", result.CropRect)
	}
	if !result.CropRect.In(img.Bounds()) {
		t.Fatalf("crop rect %v exceeds bounds %v", result.CropRect, img.Bounds())
	}
}

func TestEstimateFontColorFromReferencePointsCombinesMultipleSamples(t *testing.T) {
	bg := color.NRGBA{245, 245, 245, 255}
	img := newSolidFontTestImage(48, 24, bg)
	fillFontTestRect(img, image.Rect(1, 1, 22, 18), color.NRGBA{0, 0, 0, 255})
	shadeA := color.NRGBA{48, 90, 140, 255}
	shadeB := color.NRGBA{86, 120, 166, 255}
	shadeC := color.NRGBA{62, 105, 152, 255}
	fillFontTestRect(img, image.Rect(28, 4, 34, 10), shadeA)
	fillFontTestRect(img, image.Rect(36, 6, 42, 12), shadeB)
	fillFontTestRect(img, image.Rect(31, 14, 39, 19), shadeC)

	_, automaticForeground, ok := estimateFontForegroundColor(img)
	if !ok {
		t.Fatal("automatic foreground estimation failed")
	}
	if fontColorDistanceSq(automaticForeground, shadeA) <= 20*20 {
		t.Fatalf("test setup invalid: automatic foreground already selected target color %v", automaticForeground)
	}

	_, referencedForeground, tolerance, ok := estimateFontColorFromReferencePoints(img, []fontColorReferencePoint{
		{Point: image.Pt(29, 5), Color: shadeA},
		{Point: image.Pt(37, 7), Color: shadeB},
		{Point: image.Pt(32, 15), Color: shadeC},
	})
	if !ok {
		t.Fatal("reference point color estimation failed")
	}
	for _, sample := range []color.NRGBA{shadeA, shadeB, shadeC} {
		if !isColorMatch(sample, referencedForeground, tolerance) {
			t.Fatalf("sample %v not covered by foreground %v tolerance %v", sample, referencedForeground, tolerance)
		}
	}
	if isColorMatch(bg, referencedForeground, tolerance) {
		t.Fatalf("background %v should not match foreground %v tolerance %v", bg, referencedForeground, tolerance)
	}
}

func TestFindCharacterBBoxesRespondsToGapParams(t *testing.T) {
	img := newSolidFontTestImage(12, 8, color.NRGBA{255, 255, 255, 255})
	fillFontTestRect(img, image.Rect(1, 2, 4, 6), color.NRGBA{0, 0, 0, 255})
	fillFontTestRect(img, image.Rect(6, 2, 9, 6), color.NRGBA{0, 0, 0, 255})

	binary := createBinaryPreview(img, "000000-101010")
	separate := findCharacterBBoxes(binary, 0, 1)
	if len(separate) != 2 {
		t.Fatalf("gap 0 detected %d boxes, want 2", len(separate))
	}
	merged := findCharacterBBoxes(binary, 2, 1)
	if len(merged) != 1 {
		t.Fatalf("gap 2 detected %d boxes, want 1", len(merged))
	}
}

func TestFontLibExportParseRoundTrip(t *testing.T) {
	bitmap := [][]bool{{true, false, true}, {false, true, false}}
	hexData, whitePixels := encodeBitmapHex(bitmap)
	chars := []FontChar{{
		Char:        "字",
		Width:       3,
		Height:      2,
		HexData:     hexData,
		WhitePixels: whitePixels,
		Bitmap:      bitmap,
	}}

	parsed := parseFontLib(exportFontLib(chars, "000000-101010"))
	if len(parsed) != 1 {
		t.Fatalf("parsed %d chars, want 1", len(parsed))
	}
	if parsed[0].Char != chars[0].Char || parsed[0].Width != chars[0].Width || parsed[0].HexData != chars[0].HexData || parsed[0].WhitePixels != chars[0].WhitePixels {
		t.Fatalf("round trip mismatch: got %+v, want %+v", parsed[0], chars[0])
	}
}

func TestFontLibMatchKeyIgnoresWhitespacePadding(t *testing.T) {
	compact := [][]bool{
		{true, false, true},
		{false, true, false},
	}
	padded := [][]bool{
		{false, false, false, false, false},
		{false, true, false, true, false},
		{false, false, true, false, false},
		{false, false, false, false, false},
	}

	compactHex, compactWp := encodeBitmapHex(compact)
	paddedHex, paddedWp := encodeBitmapHex(padded)
	imported := FontChar{
		Char:        "字",
		Width:       len(padded[0]),
		Height:      len(padded),
		HexData:     strings.ToUpper(paddedHex),
		WhitePixels: paddedWp,
		Bitmap:      decodeBitmapHex(strings.ToUpper(paddedHex), len(padded[0]), len(padded)),
	}
	current := FontChar{
		Char:        "字",
		Width:       len(compact[0]),
		Height:      len(compact),
		HexData:     compactHex,
		WhitePixels: compactWp,
		Bitmap:      compact,
	}

	if fontCharMatchKey(imported) != fontCharMatchKey(current) {
		t.Fatalf("match key mismatch: imported %q, current %q", fontCharMatchKey(imported), fontCharMatchKey(current))
	}
}

func TestBitmapSimilarityExactMatchIgnoresWhitespacePadding(t *testing.T) {
	compact := [][]bool{
		{true, false, true},
		{false, true, false},
	}
	padded := [][]bool{
		{false, false, false, false, false},
		{false, true, false, true, false},
		{false, false, true, false, false},
		{false, false, false, false, false},
	}

	got := bitmapSimilarity(compact, padded)
	if got != 1 {
		t.Fatalf("similarity = %v, want 1", got)
	}
}

func TestBitmapSimilarityAllowsSmallDifference(t *testing.T) {
	a := [][]bool{
		{true, false},
		{false, true},
	}
	b := [][]bool{
		{true, false},
		{true, true},
	}

	got := bitmapSimilarity(a, b)
	if got != 0.75 {
		t.Fatalf("similarity = %v, want 0.75", got)
	}
}

func TestBitmapSimilarityResizesTemplate(t *testing.T) {
	cell := [][]bool{
		{true, true, false, false},
		{true, true, false, false},
		{false, false, true, true},
		{false, false, true, true},
	}
	template := [][]bool{
		{true, false},
		{false, true},
	}

	got := bitmapSimilarity(cell, template)
	if got != 1 {
		t.Fatalf("similarity = %v, want 1", got)
	}
}

func TestBestFontCharMatchUsesThreshold(t *testing.T) {
	cell := [][]bool{
		{true, false},
		{false, true},
	}
	near := [][]bool{
		{true, false},
		{true, true},
	}
	nearHex, nearWhite := encodeBitmapHex(near)
	chars := []FontChar{{
		Char:        "近",
		Width:       len(near[0]),
		Height:      len(near),
		HexData:     strings.ToUpper(nearHex),
		WhitePixels: nearWhite,
	}}

	name, score, ok := bestFontCharMatch(cell, chars, 0.7)
	if !ok || name != "近" || score != 0.75 {
		t.Fatalf("match = %q %.2f %v, want 近 0.75 true", name, score, ok)
	}
	if _, _, ok := bestFontCharMatch(cell, chars, 0.8); ok {
		t.Fatal("expected match below threshold to be rejected")
	}
}

func TestFontImageViewerImagePositionRejectsOutsideDisplayedImage(t *testing.T) {
	img := newSolidFontTestImage(4, 3, color.NRGBA{255, 255, 255, 255})
	viewer := newFontImageViewer()
	viewer.SetImage(img)

	point, ok := viewer.imagePosition(fyne.NewPos(2.9, 1.2))
	if !ok || point != image.Pt(2, 1) {
		t.Fatalf("imagePosition inside = %v, %v; want (2,1), true", point, ok)
	}
	if _, ok := viewer.imagePosition(fyne.NewPos(4, 1)); ok {
		t.Fatal("expected x at displayed image edge to be rejected")
	}
	if _, ok := viewer.imagePosition(fyne.NewPos(1, 3)); ok {
		t.Fatal("expected y at displayed image edge to be rejected")
	}
}

func TestFontImageViewerImagePositionUsesRoundedDisplaySize(t *testing.T) {
	img := newSolidFontTestImage(10, 3, color.NRGBA{255, 255, 255, 255})
	viewer := newFontImageViewer()
	viewer.SetImage(img)
	viewer.SetZoom(1.1)

	displayW, displayH := viewer.displayPixelSize()
	if displayW != 11 || displayH != 4 {
		t.Fatalf("display size = %dx%d, want 11x4", displayW, displayH)
	}
	point, ok := viewer.imagePosition(fyne.NewPos(10.9, 3.9))
	if !ok || point != image.Pt(9, 2) {
		t.Fatalf("imagePosition at rounded display edge = %v, %v; want (9,2), true", point, ok)
	}
	if _, ok := viewer.imagePosition(fyne.NewPos(11, 1)); ok {
		t.Fatal("expected x at rounded display edge to be rejected")
	}
}

func TestFontImageViewerUsesCrosshairCursor(t *testing.T) {
	viewer := newFontImageViewer()
	if got := viewer.Cursor(); got != desktop.CrosshairCursor {
		t.Fatalf("cursor = %v, want %v", got, desktop.CrosshairCursor)
	}
}

func TestFontSourceInitialZoomMakesSmallImagesVisible(t *testing.T) {
	zoom := fontSourceInitialZoom(image.Rect(0, 0, 131, 66))
	if zoom <= 1 {
		t.Fatalf("small source zoom = %v, want > 1", zoom)
	}
	if zoom > maxImageZoom {
		t.Fatalf("small source zoom = %v, exceeds max %v", zoom, maxImageZoom)
	}
	if float32(131)*zoom < fontSourceInitialDisplayWidth-0.5 {
		t.Fatalf("scaled width = %v, want at least %d", float32(131)*zoom, fontSourceInitialDisplayWidth)
	}
}

func TestFontSourceInitialZoomKeepsLargeImagesAtOne(t *testing.T) {
	zoom := fontSourceInitialZoom(image.Rect(0, 0, 800, 300))
	if zoom != 1 {
		t.Fatalf("large source zoom = %v, want 1", zoom)
	}
}

func TestFontPreviewZoomMatchesSourcePixelScale(t *testing.T) {
	sourceZoom := float32(2.7)
	got := fontPreviewZoomForSourceZoom(sourceZoom)
	want := sourceZoom / float32(dotCellSize+1)
	if got < want-0.0001 || got > want+0.0001 {
		t.Fatalf("preview zoom = %v, want %v", got, want)
	}
}

func TestFontPreviewZoomClampsToMinimum(t *testing.T) {
	got := fontPreviewZoomForSourceZoom(0.1)
	if got != minImageZoom {
		t.Fatalf("preview zoom = %v, want min %v", got, minImageZoom)
	}
}
