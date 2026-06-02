//go:build !windows
// +build !windows

package main

import (
	"context"
	"fmt"
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

	cmd := exec.CommandContext(ctx, adb, str...)
	var output []byte
	var err error
	if combined {
		output, err = cmd.CombinedOutput()
	} else {
		output, err = cmd.Output()
	}

	text := strings.TrimRight(string(output), "\r\n")
	if ctx.Err() == context.DeadlineExceeded {
		return text, fmt.Errorf("adb 命令超时（%s）", adbCommandTimeout)
	}
	if err != nil {
		return text, err
	}
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
