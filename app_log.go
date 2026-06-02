package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

var appLogFile *os.File
var appLogPath string

func setupAppLogging() {
	appLogPath = filepath.Join(os.TempDir(), "autogo-color-helper.log")

	file, err := os.OpenFile(appLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("打开日志文件失败: %v", err)
		return
	}

	appLogFile = file
	log.SetOutput(io.MultiWriter(os.Stderr, file))
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
	log.Printf("日志文件: %s", appLogPath)
}

func logPreview(value string, maxRunes int) string {
	runes := []rune(value)
	if maxRunes <= 0 || len(runes) <= maxRunes {
		return value
	}
	if maxRunes <= 3 {
		return string(runes[:maxRunes])
	}
	return string(runes[:maxRunes-3]) + "..."
}
