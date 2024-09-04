package services

import (
	"syscall"
)

// Constantes Windows
const (
	WM_HOTKEY = 0x0312
)

// DLL et Procédures Windows
var (
	user32                   = syscall.NewLazyDLL("user32.dll")
	procGetWindowRect       = user32.NewProc("GetWindowRect")
	procSetForegroundWindow   = user32.NewProc("SetForegroundWindow")
	procShowWindow            = user32.NewProc("ShowWindow")
	procEnumWindows           = user32.NewProc("EnumWindows")
	procGetWindowTextW        = user32.NewProc("GetWindowTextW")
	procUnregisterHotKey      = user32.NewProc("UnregisterHotKey")
	procSetCursorPos          = user32.NewProc("SetCursorPos")
	procSendMessage           = user32.NewProc("SendMessageA")
	procFindWindow            = user32.NewProc("FindWindowW")
	procGetCursorPos          = user32.NewProc("GetCursorPos")
	procScreenToClient        = user32.NewProc("ScreenToClient")
	procCreateCompatibleDC    = syscall.NewLazyDLL("gdi32.dll").NewProc("CreateCompatibleDC")
	procCreateCompatibleBitmap = syscall.NewLazyDLL("gdi32.dll").NewProc("CreateCompatibleBitmap")
	procSelectObject          = syscall.NewLazyDLL("gdi32.dll").NewProc("SelectObject")
	procPrintWindow           = syscall.NewLazyDLL("user32.dll").NewProc("PrintWindow")
	procGetDIBits            = syscall.NewLazyDLL("gdi32.dll").NewProc("GetDIBits")
	procDeleteObject          = syscall.NewLazyDLL("gdi32.dll").NewProc("DeleteObject")
	procDeleteDC              = syscall.NewLazyDLL("gdi32.dll").NewProc("DeleteDC")
)

type Point struct {
	X int32
	Y int32
}

