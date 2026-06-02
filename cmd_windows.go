//go:build windows
// +build windows

package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

const adbCommandTimeout = 12 * time.Second

func screenSizePixels() (int, int) {
	user32 := syscall.NewLazyDLL("user32.dll")
	getSystemMetrics := user32.NewProc("GetSystemMetrics")

	width, _, _ := getSystemMetrics.Call(0)  // SM_CXSCREEN
	height, _, _ := getSystemMetrics.Call(1) // SM_CYSCREEN
	return int(width), int(height)
}

func screenScale() float32 {
	user32 := syscall.NewLazyDLL("user32.dll")
	getDpiForSystem := user32.NewProc("GetDpiForSystem")

	dpi, _, _ := getDpiForSystem.Call()
	if dpi <= 0 {
		return 1
	}
	return float32(dpi) / 96
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
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

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
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		localAdbPath := filepath.Join(exeDir, "adb.exe")
		if _, err := os.Stat(localAdbPath); err == nil {
			return localAdbPath
		}
	}

	if path, err := exec.LookPath("adb"); err == nil {
		return path
	}

	processes, err := process.Processes()
	if err == nil {
		for _, p := range processes {
			name, err := p.Name()
			if err != nil || strings.ToLower(name) != "adb.exe" {
				continue
			}
			path, err := p.Exe()
			if err == nil && path != "" {
				return path
			}
		}
	}

	return "adb"
}
