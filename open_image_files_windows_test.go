//go:build windows
// +build windows

package main

import (
	"path/filepath"
	"testing"
	"unicode/utf16"
)

func TestParseWindowsOpenFileNamesSingle(t *testing.T) {
	filePath := filepath.Join("C:\\", "images", "one.png")

	got := parseWindowsOpenFileNames(utf16OpenFileBuffer(filePath))

	if len(got) != 1 || got[0] != filePath {
		t.Fatalf("got %q, want [%q]", got, filePath)
	}
}

func TestParseWindowsOpenFileNamesMultipleKeepsBufferOrder(t *testing.T) {
	dir := filepath.Join("C:\\", "images")
	want := []string{
		filepath.Join(dir, "b.png"),
		filepath.Join(dir, "a.jpg"),
		filepath.Join(dir, "c.bmp"),
	}

	got := parseWindowsOpenFileNames(utf16OpenFileBuffer(dir, "b.png", "a.jpg", "c.bmp"))

	if len(got) != len(want) {
		t.Fatalf("got %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %q, want %q", got, want)
		}
	}
}

func utf16OpenFileBuffer(parts ...string) []uint16 {
	buf := make([]uint16, 0)
	for _, part := range parts {
		buf = append(buf, utf16.Encode([]rune(part))...)
		buf = append(buf, 0)
	}
	buf = append(buf, 0)
	return buf
}
