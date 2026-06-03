package main

import (
	_ "embed"
	"log"
	"os"
	"path/filepath"
	"sync"
)

//go:embed cap.dex
var capDexData []byte

// capDexPath 存储临时目录中 cap.dex 的路径
var capDexPath string

const androidCapDexDevicePath = "/data/local/tmp/cap.dex"
const androidCapDexMainClass = "com.autogo.vdm.Main"
const androidCapDexAppInfoClass = "com.autogo.vdm.AppInfo"

// pushedDevices 记录已推送过 cap.dex 的设备，避免重复推送
var pushedDevices = make(map[string]bool)
var pushedDevicesMutex sync.RWMutex

// extractCapDex 将嵌入的 cap.dex 释放到系统临时目录
// 应在程序启动时调用
func extractCapDex() error {
	tempDir := os.TempDir()
	capDexPath = filepath.Join(tempDir, "cap.dex")

	// 写入临时文件
	err := os.WriteFile(capDexPath, capDexData, 0644)
	if err != nil {
		log.Printf("释放 cap.dex 到临时目录失败: %v", err)
		return err
	}

	log.Printf("cap.dex 已释放到: %s", capDexPath)
	return nil
}

// ensureCapDexOnDevice 确保 cap.dex 已推送到指定设备
// deviceID 应为基础设备ID（不含虚拟屏后缀）
func ensureCapDexOnDevice(deviceID string) {
	// 检查是否已推送
	pushedDevicesMutex.RLock()
	if pushedDevices[deviceID] {
		pushedDevicesMutex.RUnlock()
		return
	}
	pushedDevicesMutex.RUnlock()

	// 确保 capDexPath 已设置
	if capDexPath == "" {
		log.Printf("cap.dex 未释放，跳过推送: %s", deviceID)
		return
	}

	// 执行 adb push
	log.Printf("正在向设备 %s 推送 cap.dex...", deviceID)
	output, err := adbExecCombined("-s", deviceID, "push", capDexPath, androidCapDexDevicePath)
	if err != nil {
		log.Printf("cap.dex 推送失败: %s err=%v", deviceID, adbErrorWithOutput(err, output))
		return
	}

	pushedDevicesMutex.Lock()
	pushedDevices[deviceID] = true
	pushedDevicesMutex.Unlock()
	log.Printf("cap.dex 推送完成: %s output=%q", deviceID, logPreview(output, 500))
}

func androidCapDexMainArgs(mode string, args ...string) []string {
	return androidCapDexClassArgs(androidCapDexMainClass, append([]string{mode}, args...)...)
}

func androidCapDexAppInfoArgs(args ...string) []string {
	return androidCapDexClassArgs(androidCapDexAppInfoClass, args...)
}

func androidCapDexClassArgs(mainClass string, args ...string) []string {
	command := []string{
		"shell",
		"CLASSPATH=" + androidCapDexDevicePath,
		"app_process",
		"/",
		mainClass,
	}
	return append(command, args...)
}
