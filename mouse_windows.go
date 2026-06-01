//go:build windows
// +build windows

package main

import (
	"syscall"
	"unsafe"
)

var (
	user32           = syscall.NewLazyDLL("user32.dll")
	setCursorPosProc = user32.NewProc("SetCursorPos")
	getCursorPosProc = user32.NewProc("GetCursorPos")
)

type point struct {
	X, Y int32
}

// moveSystemMouse 移动系统鼠标光标到指定屏幕坐标
func moveSystemMouse(x, y float64) {
	setCursorPosProc.Call(uintptr(int(x)), uintptr(int(y)))
}

// moveMouseRelative 相对移动鼠标
func moveMouseRelative(dx, dy float64) {
	var pt point
	getCursorPosProc.Call(uintptr(unsafe.Pointer(&pt)))
	setCursorPosProc.Call(uintptr(pt.X+int32(dx)), uintptr(pt.Y+int32(dy)))
}
