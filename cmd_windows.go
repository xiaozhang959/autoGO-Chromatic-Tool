//go:build windows
// +build windows

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/shirou/gopsutil/v3/process"
)

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
	cmd := exec.Command(adb, str...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
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
	processes, err := process.Processes()
	if err == nil {
		for _, p := range processes {
			name, err := p.Name()
			if err != nil {
				continue
			}
			if strings.ToLower(name) == "adb.exe" {
				path, err := p.Exe()
				if err != nil {
					continue
				}
				return path
			}
		}
	}

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

	return "adb"
}
