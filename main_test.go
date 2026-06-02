package main

import "testing"

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
