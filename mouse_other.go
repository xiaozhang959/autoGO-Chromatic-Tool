//go:build !darwin && !windows
// +build !darwin,!windows

package main

// moveSystemMouse 移动系统鼠标光标（其他平台暂不支持）
func moveSystemMouse(x, y float64) {
	// 其他平台暂不支持移动鼠标
}

// moveMouseRelative 相对移动鼠标（其他平台暂不支持）
func moveMouseRelative(dx, dy float64) {
	// 其他平台暂不支持移动鼠标
}
