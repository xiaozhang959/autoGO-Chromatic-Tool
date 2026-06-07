//go:build windows
// +build windows

package main

import (
	"fmt"
	"path/filepath"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"github.com/TheTitanrain/w32"
	nativedialog "github.com/sqweek/dialog"
)

func openImageFiles() ([]string, error) {
	buf := make([]uint16, 65536)
	filters := windowsImageFileFilters()
	title, err := syscall.UTF16PtrFromString("选择图片文件")
	if err != nil {
		return nil, err
	}

	ofn := w32.OPENFILENAME{
		File:    &buf[0],
		MaxFile: uint32(len(buf)),
		Filter:  &filters[0],
		Title:   title,
		Flags: w32.OFN_FILEMUSTEXIST |
			w32.OFN_NOCHANGEDIR |
			w32.OFN_EXPLORER |
			w32.OFN_ALLOWMULTISELECT,
	}
	ofn.StructSize = uint32(unsafe.Sizeof(ofn))

	if w32.GetOpenFileName(&ofn) {
		return parseWindowsOpenFileNames(buf), nil
	}

	if errCode := w32.CommDlgExtendedError(); errCode != 0 {
		return nil, fmt.Errorf("CommDlgExtendedError: %#x", errCode)
	}
	return nil, nativedialog.ErrCancelled
}

func windowsImageFileFilters() []uint16 {
	filters := make([]uint16, 0)
	filters = append(filters, utf16.Encode([]rune("图片文件"))...)
	filters = append(filters, 0)
	filters = append(filters, utf16.Encode([]rune("*.png;*.jpg;*.jpeg;*.bmp"))...)
	filters = append(filters, 0, 0)
	return filters
}

func parseWindowsOpenFileNames(buf []uint16) []string {
	parts := make([]string, 0)
	start := 0
	for i, r := range buf {
		if r != 0 {
			continue
		}
		if i == start {
			break
		}
		parts = append(parts, string(utf16.Decode(buf[start:i])))
		start = i + 1
	}

	if len(parts) <= 1 {
		return parts
	}

	dir := parts[0]
	filePaths := make([]string, 0, len(parts)-1)
	for _, name := range parts[1:] {
		filePaths = append(filePaths, filepath.Join(dir, name))
	}
	return filePaths
}
