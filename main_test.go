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

func distinctPointColorBuckets(t *testing.T, img image.Image, points []image.Point) map[int]bool {
	t.Helper()

	classes := make(map[int]bool)
	for _, point := range points {
		class, ok := pickColorBucketAt(img, point.X, point.Y)
		if !ok {
			t.Fatalf("point has no color bucket: %v", point)
		}
		classes[class] = true
	}
	return classes
}

func assertStringSliceEqual(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("slice length mismatch: want %d got %d (%v)", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("slice item %d mismatch: want %q got %q (%v)", i, want[i], got[i], got)
		}
	}
}

func TestAndroidCapDexMainArgsUsesClasspathAppProcess(t *testing.T) {
	got := androidCapDexMainArgs("2", "7", "/sdcard/screenshot.png")
	want := []string{
		"shell",
		"CLASSPATH=" + androidCapDexDevicePath,
		"app_process",
		"/",
		androidCapDexMainClass,
		"2",
		"7",
		"/sdcard/screenshot.png",
	}
	assertStringSliceEqual(t, got, want)
}

func TestAndroidScreenshotTempPathsPreferDataLocalTmp(t *testing.T) {
	assertStringSliceEqual(t, androidScreenshotTempPaths(), []string{
		"/data/local/tmp/screenshot_temp.png",
		"/sdcard/screenshot_temp.png",
	})
}

func TestParseVirtualDisplayIDs(t *testing.T) {
	output := "\n0\n2\nabc\n2\n 15 \n"
	assertStringSliceEqual(t, parseVirtualDisplayIDs(output), []string{"2", "15"})
}

func TestParseDumpsysVirtualDisplayIDs(t *testing.T) {
	output := `
  mNextNonDefaultDisplayId=17
  mDefaultViewport=DisplayViewport{valid=true, displayId=0}
    mDisplayId=0
    mDisplayId=3
    mDisplayId=3
    mDisplayId=12
`
	assertStringSliceEqual(t, parseDumpsysVirtualDisplayIDs(output), []string{"3", "12"})
}

func TestBuildDeviceDisplayOptions(t *testing.T) {
	devices := []string{"emulator-5554", "emulator-5554[2]", "device-1[9]", "device-1", "emulator-5554[2]"}

	baseDevices, displaysByDevice := buildDeviceDisplayOptions(devices)

	assertStringSliceEqual(t, baseDevices, []string{"emulator-5554", "device-1"})
	assertStringSliceEqual(t, displaysByDevice["emulator-5554"], []string{"0", "2"})
	assertStringSliceEqual(t, displaysByDevice["device-1"], []string{"0", "9"})
}

func TestSelectOptionsWithPrompt(t *testing.T) {
	assertStringSliceEqual(t, selectOptionsWithPrompt("--虚拟屏选择--", []string{"0", "2"}), []string{"--虚拟屏选择--", "0", "2"})
}

func TestFormatAndroidDeviceID(t *testing.T) {
	if got := formatAndroidDeviceID("emulator-5554", "0"); got != "emulator-5554" {
		t.Fatalf("primary display device mismatch: %q", got)
	}
	if got := formatAndroidDeviceID("emulator-5554", "7"); got != "emulator-5554[7]" {
		t.Fatalf("virtual display device mismatch: %q", got)
	}
	if got := formatAndroidDeviceID(" ", "7"); got != "" {
		t.Fatalf("empty base device should stay empty, got %q", got)
	}
}

func TestParseShortcutText(t *testing.T) {
	tests := []struct {
		name             string
		text             string
		wantFormatted    string
		wantShortcutName string
		wantErr          bool
	}{
		{
			name:             "custom combo",
			text:             "Ctrl + Shift + S",
			wantFormatted:    "Ctrl+Shift+S",
			wantShortcutName: "CustomDesktop:Shift+Control+S",
		},
		{
			name:             "builtin select all",
			text:             "Ctrl+A",
			wantFormatted:    "Ctrl+A",
			wantShortcutName: "SelectAll",
		},
		{
			name:             "builtin undo",
			text:             "Ctrl+Z",
			wantFormatted:    "Ctrl+Z",
			wantShortcutName: "Undo",
		},
		{
			name:             "function key combo",
			text:             "Ctrl+F1",
			wantFormatted:    "Ctrl+F1",
			wantShortcutName: "CustomDesktop:Control+F1",
		},
		{
			name:    "missing modifier",
			text:    "A",
			wantErr: true,
		},
		{
			name: "empty disables shortcut",
			text: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shortcut, formatted, err := parseShortcutText(tt.text)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if formatted != tt.wantFormatted {
				t.Fatalf("formatted shortcut mismatch: want %q got %q", tt.wantFormatted, formatted)
			}
			if tt.wantShortcutName == "" {
				if shortcut != nil {
					t.Fatalf("expected nil shortcut, got %s", shortcut.ShortcutName())
				}
				return
			}
			if shortcut == nil {
				t.Fatalf("expected shortcut")
			}
			if got := shortcut.ShortcutName(); got != tt.wantShortcutName {
				t.Fatalf("shortcut name mismatch: want %q got %q", tt.wantShortcutName, got)
			}
		})
	}
}

func TestShortcutTextFromKeyboardShortcut(t *testing.T) {
	tests := []struct {
		name     string
		shortcut fyne.KeyboardShortcut
		want     string
		wantOK   bool
	}{
		{
			name:     "custom shortcut",
			shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: fyne.KeyModifierControl | fyne.KeyModifierShift},
			want:     "Ctrl+Shift+S",
			wantOK:   true,
		},
		{
			name:     "builtin undo",
			shortcut: &fyne.ShortcutUndo{},
			want:     "Ctrl+Z",
			wantOK:   true,
		},
		{
			name:     "missing modifier",
			shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyS},
		},
		{
			name:     "unsupported key",
			shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyMinus, Modifier: fyne.KeyModifierControl},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := shortcutTextFromKeyboardShortcut(tt.shortcut)
			if ok != tt.wantOK {
				t.Fatalf("ok mismatch: want %v got %v", tt.wantOK, ok)
			}
			if got != tt.want {
				t.Fatalf("shortcut text mismatch: want %q got %q", tt.want, got)
			}
		})
	}
}

func TestShortcutCaptureEntryRestoresValueWhenUnchanged(t *testing.T) {
	fynetest.NewTempApp(t)

	entry := newShortcutCaptureEntry()
	entry.SetText("Ctrl+Z")
	entry.FocusGained()
	if entry.Text != "" {
		t.Fatalf("focused shortcut entry should show placeholder by clearing display text, got %q", entry.Text)
	}

	entry.FocusLost()
	if entry.Text != "Ctrl+Z" {
		t.Fatalf("unchanged shortcut entry should restore previous value, got %q", entry.Text)
	}
}

func TestShortcutCaptureEntryKeepsCapturedShortcut(t *testing.T) {
	fynetest.NewTempApp(t)

	entry := newShortcutCaptureEntry()
	entry.SetText("Ctrl+Z")
	entry.FocusGained()
	entry.TypedShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: fyne.KeyModifierControl | fyne.KeyModifierShift})
	entry.FocusLost()

	if entry.Text != "Ctrl+Shift+S" {
		t.Fatalf("captured shortcut mismatch: %q", entry.Text)
	}
}

func TestShortcutCaptureEntryKeepsClearedShortcut(t *testing.T) {
	fynetest.NewTempApp(t)

	entry := newShortcutCaptureEntry()
	entry.SetText("Ctrl+Z")
	entry.FocusGained()
	entry.TypedKey(&fyne.KeyEvent{Name: fyne.KeyDelete})
	entry.FocusLost()

	if entry.Text != "" {
		t.Fatalf("cleared shortcut should stay empty, got %q", entry.Text)
	}
}

func TestNormalizeShortcutConfigPreservesBlankOverrides(t *testing.T) {
	got := normalizeShortcutConfig(map[string]string{
		shortcutActionScreenshot: "",
		shortcutActionRange:      "Ctrl+Shift+R",
	})

	if got[shortcutActionScreenshot] != "" {
		t.Fatalf("blank shortcut override should be preserved, got %q", got[shortcutActionScreenshot])
	}
	if got[shortcutActionRange] != "Ctrl+Shift+R" {
		t.Fatalf("custom shortcut mismatch: %q", got[shortcutActionRange])
	}
	if got[shortcutActionImport] != defaultShortcutTexts[shortcutActionImport] {
		t.Fatalf("missing shortcut should use default, got %q", got[shortcutActionImport])
	}
}

func TestNormalizeShortcutConfigAddsCommandDefaults(t *testing.T) {
	got := normalizeShortcutConfig(nil)

	if got[shortcutActionScreenshot] != "Ctrl+Z" {
		t.Fatalf("screenshot default shortcut mismatch: %q", got[shortcutActionScreenshot])
	}
	if value, ok := got[commandCopyCode]; !ok || value != "" {
		t.Fatalf("new command shortcut should default to disabled, got value=%q ok=%v", value, ok)
	}
}

func TestNodeAndAppInfoCommandsAreConfigurable(t *testing.T) {
	ids := []string{
		commandNodeCaptureSimple,
		commandNodeSearch,
		commandNodePrevious,
		commandNodeNext,
		commandNodeSelectAll,
		commandNodeClearSelected,
		commandNodeTestSelector,
		commandNodeGenerateCode,
		commandNodeCopyCode,
		commandNodeCopyParams,
		commandNodeFormat,
		commandAppInfoQuery,
		commandAppInfoCopyName,
		commandAppInfoCopyPkg,
		commandAppInfoCopyLaunch,
		commandAppInfoCopyActs,
	}
	shortcuts := normalizeShortcutConfig(nil)

	for _, id := range ids {
		if _, ok := commandDefinitionByID(id); !ok {
			t.Fatalf("missing command definition for %q", id)
		}
		if value, ok := shortcuts[id]; !ok || value != "" {
			t.Fatalf("command %q should default to disabled shortcut, got value=%q ok=%v", id, value, ok)
		}
	}
}

func TestNormalizeToolButtonConfigsDefaults(t *testing.T) {
	got := normalizeToolButtonConfigs(nil)
	want := []string{
		shortcutActionClearAll,
		commandCopyCoords,
		commandPasteCoords,
		commandApplyOffset,
		commandClearOffset,
	}

	if len(got) != len(want) {
		t.Fatalf("default tool button count mismatch: want %d got %d", len(want), len(got))
	}
	for i, id := range want {
		if got[i].CommandID != id {
			t.Fatalf("default tool button %d mismatch: want %q got %q", i, id, got[i].CommandID)
		}
		if !got[i].Visible {
			t.Fatalf("default tool button %q should be visible", id)
		}
		if got[i].Order != i+1 {
			t.Fatalf("default tool button %q order mismatch: %d", id, got[i].Order)
		}
	}
}

func TestNormalizeToolButtonConfigsPreservesHiddenAndCustomLabel(t *testing.T) {
	got := normalizeToolButtonConfigs([]ToolButtonConfig{
		{CommandID: shortcutActionClearAll, Visible: false, Order: 2},
		{CommandID: commandCopyCode, Label: "出码", Visible: true, Order: 1},
		{CommandID: "unknown", Visible: true, Order: 3},
	})

	if got[0].CommandID != commandCopyCode || got[0].Label != "出码" || !got[0].Visible {
		t.Fatalf("custom command should stay first and visible with label, got %+v", got[0])
	}
	if got[1].CommandID != shortcutActionClearAll || got[1].Visible {
		t.Fatalf("hidden default command should stay hidden, got %+v", got[1])
	}
	for _, button := range got {
		if button.CommandID == "unknown" {
			t.Fatalf("unknown command should be removed")
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

func TestAutoPickHighlightPointsPreferBrightRegion(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 20, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: 0x18, G: 0x18, B: 0x18, A: 0xff})
		}
	}
	brightRect := image.Rect(6, 7, 14, 15)
	for y := brightRect.Min.Y; y < brightRect.Max.Y; y++ {
		for x := brightRect.Min.X; x < brightRect.Max.X; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: 0xf0, G: 0xf0, B: 0xf0, A: 0xff})
		}
	}

	points := autoPickPoints(autoPickRequest{
		Image:       img,
		Rect:        image.Rect(0, 0, 20, 20),
		Count:       8,
		Mode:        autoPickModeHighlight,
		MinDistance: 2,
	})

	if len(points) == 0 {
		t.Fatal("expected highlight points, got none")
	}
	for _, point := range points {
		if !point.In(brightRect) {
			t.Fatalf("highlight point should be inside bright region %v, got %v in %v", brightRect, point, points)
		}
	}
}

func TestAutoPickHighlightPointsPureDarkReturnsEmpty(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: 0x10, G: 0x10, B: 0x10, A: 0xff})
		}
	}

	points := autoPickPoints(autoPickRequest{
		Image:       img,
		Rect:        image.Rect(0, 0, 10, 10),
		Count:       5,
		Mode:        autoPickModeHighlight,
		MinDistance: 2,
	})

	if len(points) != 0 {
		t.Fatalf("expected no highlight points for pure dark image, got %v", points)
	}
}

func TestAutoPickHighSaturationPointsPreferColorfulRegion(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 24, 12))
	for y := 0; y < 12; y++ {
		for x := 0; x < 24; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: 0x80, G: 0x80, B: 0x80, A: 0xff})
		}
	}

	redRect := image.Rect(2, 2, 9, 9)
	for y := redRect.Min.Y; y < redRect.Max.Y; y++ {
		for x := redRect.Min.X; x < redRect.Max.X; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: 0xf0, G: 0x10, B: 0x10, A: 0xff})
		}
	}
	whiteRect := image.Rect(14, 2, 21, 9)
	for y := whiteRect.Min.Y; y < whiteRect.Max.Y; y++ {
		for x := whiteRect.Min.X; x < whiteRect.Max.X; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff})
		}
	}

	points := autoPickPoints(autoPickRequest{
		Image:       img,
		Rect:        image.Rect(0, 0, 24, 12),
		Count:       6,
		Mode:        autoPickModeHighSaturation,
		MinDistance: 2,
	})

	if len(points) == 0 {
		t.Fatal("expected high saturation points, got none")
	}
	for _, point := range points {
		if !point.In(redRect) {
			t.Fatalf("high saturation point should be inside colorful region %v, got %v in %v", redRect, point, points)
		}
	}
}

func TestAutoPickHighSaturationPointsIgnoreTransparentPixels(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: 0xff, A: 0})
		}
	}

	points := autoPickPoints(autoPickRequest{
		Image:       img,
		Rect:        image.Rect(0, 0, 10, 10),
		Count:       5,
		Mode:        autoPickModeHighSaturation,
		MinDistance: 2,
	})

	if len(points) != 0 {
		t.Fatalf("expected no high saturation points for transparent image, got %v", points)
	}
}

func TestAutoPickColorClassRandomPointsCoverMainColors(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 24, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 24; x++ {
			switch {
			case x < 8:
				img.SetNRGBA(x, y, color.NRGBA{R: 0xf0, G: 0x10, B: 0x10, A: 0xff})
			case x < 16:
				img.SetNRGBA(x, y, color.NRGBA{R: 0x10, G: 0xf0, B: 0x10, A: 0xff})
			default:
				img.SetNRGBA(x, y, color.NRGBA{R: 0x10, G: 0x10, B: 0xf0, A: 0xff})
			}
		}
	}

	points := autoPickPoints(autoPickRequest{
		Image:       img,
		Rect:        image.Rect(0, 0, 24, 10),
		Count:       6,
		Mode:        autoPickModeColorClassRandom,
		MinDistance: 1,
		Rand:        rand.New(rand.NewSource(2)),
	})

	if len(points) != 6 {
		t.Fatalf("point count mismatch: want 6 got %d (%v)", len(points), points)
	}
	classes := distinctPointColorBuckets(t, img, points)
	if len(classes) != 3 {
		t.Fatalf("expected points to cover 3 color classes, got %d (%v)", len(classes), points)
	}
}

func TestAutoPickColorClassContourPointsPreferColorBoundaries(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 24, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 24; x++ {
			switch {
			case x < 8:
				img.SetNRGBA(x, y, color.NRGBA{R: 0xf0, G: 0x10, B: 0x10, A: 0xff})
			case x < 16:
				img.SetNRGBA(x, y, color.NRGBA{R: 0x10, G: 0xf0, B: 0x10, A: 0xff})
			default:
				img.SetNRGBA(x, y, color.NRGBA{R: 0x10, G: 0x10, B: 0xf0, A: 0xff})
			}
		}
	}

	points := autoPickPoints(autoPickRequest{
		Image:       img,
		Rect:        image.Rect(0, 0, 24, 10),
		Count:       6,
		Mode:        autoPickModeColorClassContour,
		MinDistance: 1,
	})

	if len(points) == 0 {
		t.Fatal("expected color class contour points, got none")
	}
	for _, point := range points {
		if point.X != 7 && point.X != 8 && point.X != 15 && point.X != 16 {
			t.Fatalf("color class contour point should be on color boundary, got %v in %v", point, points)
		}
	}
	classes := distinctPointColorBuckets(t, img, points)
	if len(classes) < 2 {
		t.Fatalf("expected contour points to cover multiple color classes, got %d (%v)", len(classes), points)
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

func TestInitialRightPanelSplitOffsetUsesSavedValue(t *testing.T) {
	config := defaultUserConfig()
	config.RightPanelSplitOffset = 0.73

	got := initialRightPanelSplitOffset(config, 1000, 190, 340)
	if got != 0.73 {
		t.Fatalf("saved split offset was not used: got %v", got)
	}
}

func TestInitialRightPanelSplitOffsetFallsBackForInvalidValue(t *testing.T) {
	config := defaultUserConfig()
	config.RightPanelSplitOffset = 1.2

	got := initialRightPanelSplitOffset(config, 1000, 190, 340)
	want := splitOffsetForFixedRightWidth(1000, 190, 340)
	if got != want {
		t.Fatalf("split offset fallback mismatch: want %v got %v", want, got)
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
