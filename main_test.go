package main

import (
	"image"
	"image/color"
	"math/rand"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
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

func assertImagePoints(t *testing.T, got, want []image.Point) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("point count mismatch: want %d got %d (%v)", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("point %d mismatch: want %v got %v", i, want[i], got[i])
		}
	}
}

func TestParsePickCount(t *testing.T) {
	tests := []struct {
		text string
		want int
	}{
		{text: "20个", want: 20},
		{text: "20", want: 20},
		{text: "", want: defaultAutoPickCount},
		{text: "abc", want: defaultAutoPickCount},
		{text: "0个", want: 0},
	}

	for _, tt := range tests {
		if got := parsePickCount(tt.text); got != tt.want {
			t.Fatalf("parsePickCount(%q) mismatch: want %d got %d", tt.text, tt.want, got)
		}
	}
}

func TestAutoPickRandomPointsStayInRectAndUnique(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 20, 20))
	rect := image.Rect(2, 3, 12, 13)

	points := autoPickPoints(autoPickRequest{
		Image:       img,
		Rect:        rect,
		Count:       10,
		Mode:        autoPickModeRandom,
		MinDistance: 1,
		Rand:        rand.New(rand.NewSource(1)),
	})

	if len(points) != 10 {
		t.Fatalf("point count mismatch: want 10 got %d (%v)", len(points), points)
	}
	seen := make(map[image.Point]struct{}, len(points))
	for _, point := range points {
		if !point.In(rect) {
			t.Fatalf("point out of rect: %v not in %v", point, rect)
		}
		if _, exists := seen[point]; exists {
			t.Fatalf("duplicate point: %v in %v", point, points)
		}
		seen[point] = struct{}{}
	}
}

func TestAutoPickContourPointsPreferBoundary(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 20, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 20; x++ {
			if x < 10 {
				img.SetNRGBA(x, y, color.NRGBA{A: 0xff})
			} else {
				img.SetNRGBA(x, y, color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff})
			}
		}
	}

	points := autoPickPoints(autoPickRequest{
		Image:       img,
		Rect:        image.Rect(0, 0, 20, 10),
		Count:       5,
		Mode:        autoPickModeContour,
		MinDistance: 2,
	})

	if len(points) == 0 {
		t.Fatal("expected contour points, got none")
	}
	for _, point := range points {
		if point.X < 9 || point.X > 10 {
			t.Fatalf("contour point should be near boundary x=10, got %v in %v", point, points)
		}
	}
}

func TestAutoPickContourPointsPureColorDoesNotPanic(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: 0x80, G: 0x80, B: 0x80, A: 0xff})
		}
	}

	points := autoPickPoints(autoPickRequest{
		Image:       img,
		Rect:        image.Rect(0, 0, 10, 10),
		Count:       5,
		Mode:        autoPickModeContour,
		MinDistance: 2,
	})

	if len(points) != 0 {
		t.Fatalf("expected no contour points for pure color image, got %v", points)
	}
}

func TestNormalizePickRectClampsAndNormalizes(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 10, 8))

	got := normalizePickRect(img, image.Rect(8, 6, -3, 2))
	want := image.Rect(0, 2, 8, 6)

	if got != want {
		t.Fatalf("rect mismatch: want %v got %v", want, got)
	}
}

func TestImageViewerAddPointsBatchRefreshesOnce(t *testing.T) {
	fynetest.NewTempApp(t)

	oldColorPoints := colorPoints
	oldImageViewer := imageViewer
	oldRefreshColorList := refreshColorList
	oldDefaultOffset := defaultColorPointOffset
	t.Cleanup(func() {
		colorPoints = oldColorPoints
		imageViewer = oldImageViewer
		refreshColorList = oldRefreshColorList
		defaultColorPointOffset = oldDefaultOffset
	})

	img := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	img.SetNRGBA(1, 1, color.NRGBA{R: 0x11, G: 0x22, B: 0x33, A: 0xff})
	img.SetNRGBA(2, 2, color.NRGBA{R: 0x44, G: 0x55, B: 0x66, A: 0xff})
	viewer := NewImageViewer()
	viewer.SetImage(img)
	imageViewer = viewer
	colorPoints = nil
	defaultColorPointOffset = "ABCDEF"

	refreshCount := 0
	refreshColorList = func() {
		refreshCount++
	}

	viewer.AddPoints([]image.Point{
		image.Pt(1, 1),
		image.Pt(2, 2),
		image.Pt(9, 9),
	})

	if refreshCount != 1 {
		t.Fatalf("refresh count mismatch: want 1 got %d", refreshCount)
	}
	if len(colorPoints) != 2 {
		t.Fatalf("color point count mismatch: want 2 got %d", len(colorPoints))
	}
	if len(viewer.markPoints) != 2 {
		t.Fatalf("mark point count mismatch: want 2 got %d", len(viewer.markPoints))
	}
	if colorPoints[0].ID != 0 || colorPoints[0].Position != "1, 1" || colorPoints[0].Color != "#112233" || colorPoints[0].Offset != "ABCDEF" {
		t.Fatalf("first color point mismatch: %+v", colorPoints[0])
	}
	if colorPoints[1].ID != 1 || colorPoints[1].Position != "2, 2" || colorPoints[1].Color != "#445566" {
		t.Fatalf("second color point mismatch: %+v", colorPoints[1])
	}
}

func TestImageViewerReplacePointsClearsListAndKeepsRange(t *testing.T) {
	fynetest.NewTempApp(t)

	oldColorPoints := colorPoints
	oldImageViewer := imageViewer
	oldRefreshColorList := refreshColorList
	oldRectCoordEntry := rectCoordEntry
	t.Cleanup(func() {
		colorPoints = oldColorPoints
		imageViewer = oldImageViewer
		refreshColorList = oldRefreshColorList
		rectCoordEntry = oldRectCoordEntry
	})

	rectCoordEntry = widget.NewEntry()
	rectCoordEntry.SetText("1,2,3,4")
	img := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	img.SetNRGBA(3, 3, color.NRGBA{R: 0xaa, G: 0xbb, B: 0xcc, A: 0xff})
	viewer := NewImageViewer()
	viewer.SetImage(img)
	viewer.markPoints = append(viewer.markPoints, MarkPoint{X: 1, Y: 1, Color: color.White})
	viewer.markRects = append(viewer.markRects, MarkRect{
		X1:    1,
		Y1:    2,
		X2:    3,
		Y2:    4,
		Color: color.RGBA{255, 0, 0, 255},
	})
	viewer.manualRectSelected = true
	imageViewer = viewer
	colorPoints = []ColorPoint{
		{ID: 0, Position: "1, 1", Color: "#111111", Selected: true},
		{ID: 1, Position: "2, 2", Color: "#222222", Selected: true},
	}

	refreshCount := 0
	refreshColorList = func() {
		refreshCount++
	}

	viewer.ReplacePoints([]image.Point{image.Pt(3, 3)})

	if refreshCount != 1 {
		t.Fatalf("refresh count mismatch: want 1 got %d", refreshCount)
	}
	if len(colorPoints) != 1 {
		t.Fatalf("color point count mismatch: want 1 got %d", len(colorPoints))
	}
	if colorPoints[0].ID != 0 || colorPoints[0].Position != "3, 3" || colorPoints[0].Color != "#AABBCC" {
		t.Fatalf("replacement color point mismatch: %+v", colorPoints[0])
	}
	if len(viewer.markPoints) != 1 || viewer.markPoints[0].X != 3 || viewer.markPoints[0].Y != 3 {
		t.Fatalf("replacement mark points mismatch: %+v", viewer.markPoints)
	}
	if rectCoordEntry.Text != "1,2,3,4" {
		t.Fatalf("range text was overwritten: got %q", rectCoordEntry.Text)
	}
	if len(viewer.markRects) != 1 || viewer.markRects[0].X1 != 1 || viewer.markRects[0].Y1 != 2 || viewer.markRects[0].X2 != 3 || viewer.markRects[0].Y2 != 4 {
		t.Fatalf("range rect was overwritten: %+v", viewer.markRects)
	}
}

func TestAutoPickRangeCallbackDoesNotOverwriteRangeSelection(t *testing.T) {
	fynetest.NewTempApp(t)

	oldRectCoordEntry := rectCoordEntry
	oldImageViewer := imageViewer
	t.Cleanup(func() {
		rectCoordEntry = oldRectCoordEntry
		imageViewer = oldImageViewer
	})

	rectCoordEntry = widget.NewEntry()
	rectCoordEntry.SetText("1,2,3,4")
	viewer := NewImageViewer()
	viewer.SetImage(image.NewNRGBA(image.Rect(0, 0, 20, 20)))
	viewer.markRects = append(viewer.markRects, MarkRect{
		X1:    1,
		Y1:    2,
		X2:    3,
		Y2:    4,
		Color: color.RGBA{255, 0, 0, 255},
	})
	viewer.manualRectSelected = true
	imageViewer = viewer

	var callbackRect image.Rectangle
	viewer.SetRangeSelectModeWithCallback(func(rect image.Rectangle) {
		callbackRect = rect
	})
	viewer.mouseDownX = 5
	viewer.mouseDownY = 5
	viewer.isDragging = true
	viewer.dragMode = imageDragRange
	viewer.tempRect = &MarkRect{
		X1:    5,
		Y1:    5,
		X2:    10,
		Y2:    10,
		Color: color.RGBA{255, 0, 0, 255},
	}

	viewer.MouseUp(&desktop.MouseEvent{
		PointEvent: fyne.PointEvent{Position: fyne.NewPos(10, 10)},
		Button:     desktop.MouseButtonPrimary,
	})

	if callbackRect != image.Rect(5, 5, 11, 11) {
		t.Fatalf("callback rect mismatch: got %v", callbackRect)
	}
	if rectCoordEntry.Text != "1,2,3,4" {
		t.Fatalf("range text was overwritten: got %q", rectCoordEntry.Text)
	}
	if len(viewer.markRects) != 1 || viewer.markRects[0].X1 != 1 || viewer.markRects[0].Y1 != 2 || viewer.markRects[0].X2 != 3 || viewer.markRects[0].Y2 != 4 {
		t.Fatalf("range rect was overwritten: %+v", viewer.markRects)
	}
}

func TestSplitOffsetForFixedRightWidth(t *testing.T) {
	got := splitOffsetForFixedRightWidth(1000, 190, 340)
	want := float64(470) / float64(810)
	if got < want-0.000001 || got > want+0.000001 {
		t.Fatalf("split offset mismatch: want %v got %v", want, got)
	}
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

func TestRunImageFindTestHighlightPointsFindColor(t *testing.T) {
	withColorPointsForTest(t, []ColorPoint{
		{Position: "0, 0", Color: "#AABBCC", Offset: "000000", Selected: true},
	})
	img := image.NewNRGBA(image.Rect(0, 0, 6, 6))
	img.SetNRGBA(4, 1, color.NRGBA{R: 0xaa, G: 0xbb, B: 0xcc, A: 0xff})

	got := runImageFindTestHighlightPoints(img, "FindColor", "1.0", "0: 从左到右，从上到下")

	assertImagePoints(t, got, []image.Point{image.Pt(4, 1)})
}

func TestFindTestHighlightRectsUsesLargerBox(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 50, 50))

	rects := findTestHighlightRects(img, []image.Point{image.Pt(25, 25), image.Pt(2, 3)})

	if len(rects) != 2 {
		t.Fatalf("rect count mismatch: got %d", len(rects))
	}
	if rects[0].X1 != 15 || rects[0].Y1 != 15 || rects[0].X2 != 36 || rects[0].Y2 != 36 {
		t.Fatalf("center rect mismatch: got %+v", rects[0])
	}
	if rects[1].X1 != 0 || rects[1].Y1 != 0 || rects[1].X2 != 13 || rects[1].Y2 != 14 {
		t.Fatalf("clamped rect mismatch: got %+v", rects[1])
	}
}

func TestNearestColorPointIndex(t *testing.T) {
	points := []ColorPoint{
		{Position: "10, 10"},
		{Position: "20, 20"},
		{Position: "bad"},
	}

	if got := nearestColorPointIndex(points, 18, 19, 12); got != 1 {
		t.Fatalf("nearest point mismatch: got %d", got)
	}
	if got := nearestColorPointIndex(points, 40, 40, 12); got != -1 {
		t.Fatalf("out of range point mismatch: got %d", got)
	}
}

func TestLinkedPointHighlightRects(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 60, 60))

	rects := linkedPointHighlightRects(img, []image.Point{image.Pt(30, 30)})

	if len(rects) != 1 {
		t.Fatalf("rect count mismatch: got %d", len(rects))
	}
	if rects[0].X1 != 16 || rects[0].Y1 != 16 || rects[0].X2 != 45 || rects[0].Y2 != 45 {
		t.Fatalf("linked rect mismatch: got %+v", rects[0])
	}
}

func TestRunImageFindTestHighlightPointsFindMultiColorsAll(t *testing.T) {
	withColorPointsForTest(t, []ColorPoint{
		{Position: "0, 0", Color: "#112233", Offset: "000000", Selected: true},
		{Position: "1, 0", Color: "#445566", Offset: "000000", Selected: true},
	})
	img := image.NewNRGBA(image.Rect(0, 0, 6, 6))
	img.SetNRGBA(1, 1, color.NRGBA{R: 0x11, G: 0x22, B: 0x33, A: 0xff})
	img.SetNRGBA(2, 1, color.NRGBA{R: 0x44, G: 0x55, B: 0x66, A: 0xff})
	img.SetNRGBA(3, 2, color.NRGBA{R: 0x11, G: 0x22, B: 0x33, A: 0xff})
	img.SetNRGBA(4, 2, color.NRGBA{R: 0x44, G: 0x55, B: 0x66, A: 0xff})

	got := runImageFindTestHighlightPoints(img, "FindMultiColorsAll", "1.0", "0: 从左到右，从上到下")

	assertImagePoints(t, got, []image.Point{image.Pt(1, 1), image.Pt(3, 2)})
}

func TestRunImageFindTestHighlightPointsCmpColor(t *testing.T) {
	withColorPointsForTest(t, []ColorPoint{
		{Position: "2, 3", Color: "#AABBCC", Offset: "000000", Selected: true},
	})
	img := image.NewNRGBA(image.Rect(0, 0, 6, 6))
	img.SetNRGBA(2, 3, color.NRGBA{R: 0xaa, G: 0xbb, B: 0xcc, A: 0xff})

	got := runImageFindTestHighlightPoints(img, "CmpColor", "1.0", "0: 从左到右，从上到下")

	assertImagePoints(t, got, []image.Point{image.Pt(2, 3)})
}

func TestRunImageFindTestHighlightPointsNotFound(t *testing.T) {
	withColorPointsForTest(t, []ColorPoint{
		{Position: "0, 0", Color: "#AABBCC", Offset: "000000", Selected: true},
	})
	img := image.NewNRGBA(image.Rect(0, 0, 6, 6))

	got := runImageFindTestHighlightPoints(img, "FindColor", "1.0", "0: 从左到右，从上到下")

	assertImagePoints(t, got, nil)
}

func TestSplitCodeTestArgsSupportsQuotedCommasAndFunctionCall(t *testing.T) {
	args, err := splitCodeTestArgs(`images.FindMultiColors(0, 0, 0, 0, "112233,1,1,445566", 1.0, 0, 0)`)
	if err != nil {
		t.Fatalf("split args failed: %v", err)
	}
	if len(args) != 8 {
		t.Fatalf("arg count mismatch: got %d (%v)", len(args), args)
	}
	colorText, err := unquoteCodeTestArg(args[4])
	if err != nil {
		t.Fatalf("unquote failed: %v", err)
	}
	if colorText != "112233,1,1,445566" {
		t.Fatalf("color arg mismatch: got %s", colorText)
	}
}

func TestRunCodeTestForImageFindMultiColors(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 6, 6))
	img.SetNRGBA(2, 3, color.NRGBA{R: 0x11, G: 0x22, B: 0x33, A: 0xff})
	img.SetNRGBA(3, 4, color.NRGBA{R: 0x44, G: 0x55, B: 0x66, A: 0xff})

	got := runCodeTestForImage(img, "FindMultiColors", `0, 0, 0, 0, "112233,1,1,445566", 1.0, 0, 0`)

	if got != "2,3" {
		t.Fatalf("code test result mismatch: got %s", got)
	}
}

func TestRunCodeTestForImageFindMultiColorsAll(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 6, 6))
	img.SetNRGBA(1, 1, color.NRGBA{R: 0x11, G: 0x22, B: 0x33, A: 0xff})
	img.SetNRGBA(2, 1, color.NRGBA{R: 0x44, G: 0x55, B: 0x66, A: 0xff})
	img.SetNRGBA(3, 2, color.NRGBA{R: 0x11, G: 0x22, B: 0x33, A: 0xff})
	img.SetNRGBA(4, 2, color.NRGBA{R: 0x44, G: 0x55, B: 0x66, A: 0xff})

	got := runCodeTestForImage(img, "findMultiColorAll", `0, 0, 0, 0, "112233,1,0,445566", 1.0, 0, 0`)

	want := "[\n    {1 1}\n    {3 2}\n]"
	if got != want {
		t.Fatalf("code test find all mismatch:\nwant: %s\n got: %s", want, got)
	}
}

func TestRunCodeTestForImageCmpColor(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 6, 6))
	img.SetNRGBA(2, 3, color.NRGBA{R: 0xaa, G: 0xbb, B: 0xcc, A: 0xff})

	got := runCodeTestForImage(img, "CmpColor", `2, 3, "AABBCC-000000", 1.0, 0`)

	if got != "true" {
		t.Fatalf("code test cmp mismatch: got %s", got)
	}
}

func TestCodeTestHighlightPointsFromResult(t *testing.T) {
	tests := []struct {
		name     string
		function string
		params   string
		result   string
		want     []image.Point
	}{
		{
			name:     "FindMultiColors",
			function: "FindMultiColors",
			result:   "2,3",
			want:     []image.Point{image.Pt(2, 3)},
		},
		{
			name:     "FindMultiColors not found",
			function: "FindMultiColors",
			result:   "-1,-1",
		},
		{
			name:     "FindMultiColorsAll",
			function: "FindMultiColorsAll",
			result:   "[\n    {1 1}\n    {3 2}\n]",
			want:     []image.Point{image.Pt(1, 1), image.Pt(3, 2)},
		},
		{
			name:     "CmpColor",
			function: "CmpColor",
			params:   `2, 3, "AABBCC-000000", 1.0, 0`,
			result:   "true",
			want:     []image.Point{image.Pt(2, 3)},
		},
		{
			name:     "error result",
			function: "FindColor",
			result:   "参数错误：x1 必须是整数",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := codeTestHighlightPointsFromResult(tt.function, tt.params, tt.result)
			assertImagePoints(t, got, tt.want)
		})
	}
}
