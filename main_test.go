package main

import (
	"image"
	"image/color"
	"testing"

	fynetest "fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func withColorPointsForTest(t *testing.T, points []ColorPoint) {
	t.Helper()

	oldPoints := colorPoints
	oldRectCoordEntry := rectCoordEntry
	oldTemplates := apiFormatTemplates
	t.Cleanup(func() {
		colorPoints = oldPoints
		rectCoordEntry = oldRectCoordEntry
		apiFormatTemplates = oldTemplates
	})

	rectCoordEntry = nil
	colorPoints = points
	apiFormatTemplates = defaultAPIFormatTemplates()
}

func TestBuildImagesAPICodeColorExportUsesOfficialParamOrder(t *testing.T) {
	withColorPointsForTest(t, []ColorPoint{
		{Position: "10, 20", Color: "#081029", Offset: "202020", Selected: true},
		{Position: "197, -1", Color: "#1C3A6D", Offset: "202020", Selected: true},
		{Position: "42, 148", Color: "#0D1B3C", Offset: "202020", Selected: true},
	})

	tests := []struct {
		name      string
		function  string
		wantColor string
		wantParam string
		wantCode  string
	}{
		{
			name:      "FindColor",
			function:  "FindColor",
			wantColor: `{0,0,0,0,"081029-202020|1c3a6d-202020|0d1b3c-202020",0.9,0,0}`,
			wantParam: `0, 0, 0, 0, "081029-202020|1c3a6d-202020|0d1b3c-202020", 0.9, 0, 0`,
			wantCode:  `x, y := images.FindColor(0, 0, 0, 0, "081029-202020|1c3a6d-202020|0d1b3c-202020", 0.9, 0, 0)`,
		},
		{
			name:      "FindMultiColors",
			function:  "FindMultiColors",
			wantColor: `{0,0,0,0,"081029-202020,187,-21,1c3a6d-202020,32,128,0d1b3c-202020",0.9,0,0}`,
			wantParam: `0, 0, 0, 0, "081029-202020,187,-21,1c3a6d-202020,32,128,0d1b3c-202020", 0.9, 0, 0`,
			wantCode:  `x, y := images.FindMultiColors(0, 0, 0, 0, "081029-202020,187,-21,1c3a6d-202020,32,128,0d1b3c-202020", 0.9, 0, 0)`,
		},
		{
			name:      "FindMultiColorsAll",
			function:  "FindMultiColorsAll",
			wantColor: `{0,0,0,0,"081029-202020,187,-21,1c3a6d-202020,32,128,0d1b3c-202020",0.9,0,0}`,
			wantParam: `0, 0, 0, 0, "081029-202020,187,-21,1c3a6d-202020,32,128,0d1b3c-202020", 0.9, 0, 0`,
			wantCode:  `points := images.FindMultiColorsAll(0, 0, 0, 0, "081029-202020,187,-21,1c3a6d-202020,32,128,0d1b3c-202020", 0.9, 0, 0)`,
		},
		{
			name:      "CmpColor",
			function:  "CmpColor",
			wantColor: `{10,20,"081029-202020|1c3a6d-202020|0d1b3c-202020",0.9,0}`,
			wantParam: `10, 20, "081029-202020|1c3a6d-202020|0d1b3c-202020", 0.9, 0`,
			wantCode:  `matched := images.CmpColor(10, 20, "081029-202020|1c3a6d-202020|0d1b3c-202020", 0.9, 0)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			colorText, params, code := buildImagesAPICode(tt.function, "0.9", "0: 从左到右，从上到下")

			if colorText != tt.wantColor {
				t.Fatalf("color text mismatch:\nwant: %s\n got: %s", tt.wantColor, colorText)
			}
			if params != tt.wantParam {
				t.Fatalf("params mismatch:\nwant: %s\n got: %s", tt.wantParam, params)
			}
			if code != tt.wantCode {
				t.Fatalf("code mismatch:\nwant: %s\n got: %s", tt.wantCode, code)
			}
		})
	}
}

func TestBuildImagesAPICodeUsesCustomFormatTemplate(t *testing.T) {
	withColorPointsForTest(t, []ColorPoint{
		{Position: "10, 20", Color: "#081029", Offset: "202020", Selected: true},
		{Position: "197, -1", Color: "#1C3A6D", Offset: "202020", Selected: true},
	})
	apiFormatTemplates = normalizeAPIFormatTemplates(map[string]string{
		"findmulticolor": "call([颜色参数]) params=[参数]",
	})

	_, _, code := buildImagesAPICode("FindMultiColors", "0.9", "0: 从左到右，从上到下")

	want := `call({0,0,0,0,"081029-202020,187,-21,1c3a6d-202020",0.9,0,0}) params=0, 0, 0, 0, "081029-202020,187,-21,1c3a6d-202020", 0.9, 0, 0`
	if code != want {
		t.Fatalf("custom code mismatch:\nwant: %s\n got: %s", want, code)
	}
}

func TestInsertTextAtEntryCursor(t *testing.T) {
	fynetest.NewTempApp(t)

	entry := widget.NewMultiLineEntry()
	entry.SetText("first\nsecond")
	entry.CursorRow = 1
	entry.CursorColumn = 3

	insertTextAtEntryCursor(entry, "[参数]")

	want := "first\nsec[参数]ond"
	if entry.Text != want {
		t.Fatalf("entry text mismatch:\nwant: %s\n got: %s", want, entry.Text)
	}
	if entry.CursorRow != 1 || entry.CursorColumn != 7 {
		t.Fatalf("cursor mismatch: row=%d column=%d", entry.CursorRow, entry.CursorColumn)
	}
}

func TestRunImageFindTestFindMultiColors(t *testing.T) {
	withColorPointsForTest(t, []ColorPoint{
		{Position: "0, 0", Color: "#112233", Offset: "000000", Selected: true},
		{Position: "1, 1", Color: "#445566", Offset: "000000", Selected: true},
	})
	img := image.NewNRGBA(image.Rect(0, 0, 6, 6))
	img.SetNRGBA(3, 2, color.NRGBA{R: 0x11, G: 0x22, B: 0x33, A: 0xff})
	img.SetNRGBA(4, 3, color.NRGBA{R: 0x44, G: 0x55, B: 0x66, A: 0xff})

	x, y := runImageFindTest(img, "FindMultiColors", "1.0", "0: 从左到右，从上到下")

	if x != 3 || y != 2 {
		t.Fatalf("find multi result mismatch: got %d,%d", x, y)
	}
}

func TestRunImageFindTestFindColor(t *testing.T) {
	withColorPointsForTest(t, []ColorPoint{
		{Position: "0, 0", Color: "#AABBCC", Offset: "000000", Selected: true},
	})
	img := image.NewNRGBA(image.Rect(0, 0, 6, 6))
	img.SetNRGBA(4, 1, color.NRGBA{R: 0xaa, G: 0xbb, B: 0xcc, A: 0xff})

	x, y := runImageFindTest(img, "FindColor", "1.0", "0: 从左到右，从上到下")

	if x != 4 || y != 1 {
		t.Fatalf("find color result mismatch: got %d,%d", x, y)
	}
}

func TestRunImageFindTestNotFound(t *testing.T) {
	withColorPointsForTest(t, []ColorPoint{
		{Position: "0, 0", Color: "#AABBCC", Offset: "000000", Selected: true},
	})
	img := image.NewNRGBA(image.Rect(0, 0, 6, 6))

	x, y := runImageFindTest(img, "FindColor", "1.0", "0: 从左到右，从上到下")

	if x != -1 || y != -1 {
		t.Fatalf("not found result mismatch: got %d,%d", x, y)
	}
}

func TestRunImageFindTestResultFindMultiColorsAll(t *testing.T) {
	withColorPointsForTest(t, []ColorPoint{
		{Position: "0, 0", Color: "#112233", Offset: "000000", Selected: true},
		{Position: "1, 0", Color: "#445566", Offset: "000000", Selected: true},
	})
	img := image.NewNRGBA(image.Rect(0, 0, 6, 6))
	img.SetNRGBA(1, 1, color.NRGBA{R: 0x11, G: 0x22, B: 0x33, A: 0xff})
	img.SetNRGBA(2, 1, color.NRGBA{R: 0x44, G: 0x55, B: 0x66, A: 0xff})
	img.SetNRGBA(3, 2, color.NRGBA{R: 0x11, G: 0x22, B: 0x33, A: 0xff})
	img.SetNRGBA(4, 2, color.NRGBA{R: 0x44, G: 0x55, B: 0x66, A: 0xff})

	got := runImageFindTestResult(img, "FindMultiColorsAll", "1.0", "0: 从左到右，从上到下")

	want := "[\n    {1 1}\n    {3 2}\n]"
	if got != want {
		t.Fatalf("find all result mismatch:\nwant: %s\n got: %s", want, got)
	}
}

func TestRunImageFindTestResultCmpColor(t *testing.T) {
	withColorPointsForTest(t, []ColorPoint{
		{Position: "2, 3", Color: "#AABBCC", Offset: "000000", Selected: true},
	})
	img := image.NewNRGBA(image.Rect(0, 0, 6, 6))
	img.SetNRGBA(2, 3, color.NRGBA{R: 0xaa, G: 0xbb, B: 0xcc, A: 0xff})

	got := runImageFindTestResult(img, "CmpColor", "1.0", "0: 从左到右，从上到下")

	if got != "true" {
		t.Fatalf("cmp color result mismatch: got %s", got)
	}
}
