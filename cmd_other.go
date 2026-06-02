//go:build !windows
// +build !windows

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

const adbCommandTimeout = 12 * time.Second

func screenSizePixels() (int, int) {
	return 0, 0
}

func screenScale() float32 {
	return 1
}

func adbExec(str ...string) string {
	output, err := runADBCommand(false, str...)
	if err != nil {
		return err.Error()
	}
	return output
}

func adbExecCombined(str ...string) (string, error) {
	return runADBCommand(true, str...)
}

func runADBCommand(combined bool, str ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), adbCommandTimeout)
	defer cancel()

	start := time.Now()
	log.Printf("[adb] start path=%q args=%q combined=%v", adb, str, combined)

	cmd := exec.CommandContext(ctx, adb, str...)
	var output []byte
	var err error
	if combined {
		output, err = cmd.CombinedOutput()
	} else {
		output, err = cmd.Output()
	}

	text := strings.TrimRight(string(output), "\r\n")
	elapsed := time.Since(start)
	if ctx.Err() == context.DeadlineExceeded {
		log.Printf("[adb] timeout path=%q args=%q elapsed=%s output=%q", adb, str, elapsed, logPreview(text, 500))
		return text, fmt.Errorf("adb 命令超时（%s）", adbCommandTimeout)
	}
	if err != nil {
		log.Printf("[adb] failed path=%q args=%q elapsed=%s err=%v output=%q", adb, str, elapsed, err, logPreview(text, 500))
		return text, err
	}
	log.Printf("[adb] ok path=%q args=%q elapsed=%s output=%q", adb, str, elapsed, logPreview(text, 500))
	return text, nil
}

// 查找 ADB 可执行文件路径
func findADBPath() string {
	commonPaths := []string{
		"/usr/local/bin/adb",
		"/opt/homebrew/bin/adb",
		os.Getenv("HOME") + "/Library/Android/sdk/platform-tools/adb",
		os.Getenv("HOME") + "/Android/sdk/platform-tools/adb",
		"/usr/bin/adb",
	}

	// 先检查常见路径
	for _, path := range commonPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// 尝试从 PATH 中查找
	if path, err := exec.LookPath("adb"); err == nil {
		return path
	}

	return "adb"
}
