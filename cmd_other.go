//go:build !windows
// +build !windows

package main

import (
	"os"
	"os/exec"
)

func screenSizePixels() (int, int) {
	return 0, 0
}

func screenScale() float32 {
	return 1
}

func adbExec(str ...string) string {
	cmd := exec.Command(adb, str...)
	output, err := cmd.Output()
	if err != nil {
		return err.Error()
	}
	if len(output) > 0 {
		if output[len(output)-1] == 10 {
			output = output[:len(output)-1]
		}
		if output[len(output)-1] == 13 {
			output = output[:len(output)-1]
		}
	}
	return string(output)
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
