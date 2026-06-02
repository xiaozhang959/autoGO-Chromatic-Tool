package main

import "testing"

func TestBuildImagesAPICodeFindMultiColorsColorExport(t *testing.T) {
	oldPoints := colorPoints
	oldRectCoordEntry := rectCoordEntry
	defer func() {
		colorPoints = oldPoints
		rectCoordEntry = oldRectCoordEntry
	}()

	rectCoordEntry = nil
	colorPoints = []ColorPoint{
		{Position: "10, 20", Color: "#081029", Offset: "202020", Selected: true},
		{Position: "197, -1", Color: "#1C3A6D", Offset: "202020", Selected: true},
		{Position: "42, 148", Color: "#0D1B3C", Offset: "202020", Selected: true},
	}

	colorText, params, code := buildImagesAPICode("FindMultiColors", "0.9", "0: 从左到右，从上到下")

	wantColorText := `{0,0,0,0,"081029-202020,187,-21,1c3a6d-202020,32,128,0d1b3c-202020",0.9,0,0}`
	if colorText != wantColorText {
		t.Fatalf("color text mismatch:\nwant: %s\n got: %s", wantColorText, colorText)
	}

	wantParams := `0, 0, 0, 0, "081029-202020,187,-21,1c3a6d-202020,32,128,0d1b3c-202020", 0.9, 0, 0`
	if params != wantParams {
		t.Fatalf("params mismatch:\nwant: %s\n got: %s", wantParams, params)
	}

	wantCode := `x, y := images.FindMultiColors(0, 0, 0, 0, "081029-202020,187,-21,1c3a6d-202020,32,128,0d1b3c-202020", 0.9, 0, 0)`
	if code != wantCode {
		t.Fatalf("code mismatch:\nwant: %s\n got: %s", wantCode, code)
	}
}
