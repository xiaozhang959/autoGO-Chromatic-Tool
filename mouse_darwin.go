//go:build darwin
// +build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework CoreGraphics -framework ApplicationServices
#include <CoreGraphics/CoreGraphics.h>

void moveMouseCursor(double x, double y) {
    CGPoint point = CGPointMake(x, y);
    CGWarpMouseCursorPosition(point);
    CGAssociateMouseAndMouseCursorPosition(true);
}

void getCurrentMousePosition(double *x, double *y) {
    CGEventRef event = CGEventCreate(NULL);
    CGPoint location = CGEventGetLocation(event);
    *x = location.x;
    *y = location.y;
    CFRelease(event);
}

void moveMouseRelative(double dx, double dy) {
    double x, y;
    getCurrentMousePosition(&x, &y);
    moveMouseCursor(x + dx, y + dy);
}
*/
import "C"

// moveSystemMouse 移动系统鼠标光标到指定屏幕坐标
func moveSystemMouse(x, y float64) {
	C.moveMouseCursor(C.double(x), C.double(y))
}

// moveMouseRelative 相对移动鼠标
func moveMouseRelative(dx, dy float64) {
	C.moveMouseRelative(C.double(dx), C.double(dy))
}
