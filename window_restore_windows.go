//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

const swRestore = 9

var (
	user32DLL               = syscall.NewLazyDLL("user32.dll")
	procFindWindowW         = user32DLL.NewProc("FindWindowW")
	procShowWindow          = user32DLL.NewProc("ShowWindow")
	procSetForegroundWindow = user32DLL.NewProc("SetForegroundWindow")
)

func restoreWindowByTitle(title string) {
	if title == "" {
		return
	}
	hwnd, _, _ := procFindWindowW.Call(0, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))))
	if hwnd == 0 {
		return
	}
	procShowWindow.Call(hwnd, swRestore)
	procSetForegroundWindow.Call(hwnd)
}
