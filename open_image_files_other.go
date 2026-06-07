//go:build !windows && !darwin
// +build !windows,!darwin

package main

import nativedialog "github.com/sqweek/dialog"

func openImageFiles() ([]string, error) {
	filePath, err := nativedialog.File().
		Filter("图片文件", "png", "jpg", "jpeg", "bmp").
		Title("选择图片文件").
		Load()
	if err != nil {
		return nil, err
	}
	return []string{filePath}, nil
}
