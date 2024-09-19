package services

import (
	"log"
	"syscall"
	"unsafe"
)

// Constantes Windows
const (
	WM_HOTKEY = 0x0312
)

// DLL et Procédures Windows
var (
	user32                     = syscall.NewLazyDLL("user32.dll")
	procGetWindowRect          = user32.NewProc("GetWindowRect")
	procSetForegroundWindow    = user32.NewProc("SetForegroundWindow")
	procShowWindow             = user32.NewProc("ShowWindow")
	procEnumWindows            = user32.NewProc("EnumWindows")
	procGetWindowTextW         = user32.NewProc("GetWindowTextW")
	procUnregisterHotKey       = user32.NewProc("UnregisterHotKey")
	procSetCursorPos           = user32.NewProc("SetCursorPos")
	procSendMessage            = user32.NewProc("SendMessageA")
	procFindWindow             = user32.NewProc("FindWindowW")
	procGetCursorPos           = user32.NewProc("GetCursorPos")
	procScreenToClient         = user32.NewProc("ScreenToClient")
	procCreateCompatibleDC     = syscall.NewLazyDLL("gdi32.dll").NewProc("CreateCompatibleDC")
	procCreateCompatibleBitmap = syscall.NewLazyDLL("gdi32.dll").NewProc("CreateCompatibleBitmap")
	procSelectObject           = syscall.NewLazyDLL("gdi32.dll").NewProc("SelectObject")
	procPrintWindow            = syscall.NewLazyDLL("user32.dll").NewProc("PrintWindow")
	procGetDIBits              = syscall.NewLazyDLL("gdi32.dll").NewProc("GetDIBits")
	procDeleteObject           = syscall.NewLazyDLL("gdi32.dll").NewProc("DeleteObject")
	procDeleteDC               = syscall.NewLazyDLL("gdi32.dll").NewProc("DeleteDC")

	procGetWindowTextLengthW     = user32.NewProc("GetWindowTextLengthW")
	procIsIconic                 = user32.NewProc("IsIconic")
	procGetForegroundWindow      = user32.NewProc("GetForegroundWindow")
	procAttachThreadInput        = user32.NewProc("AttachThreadInput")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
)

type Point struct {
	X int32
	Y int32
}

func GetWindowTextLength(hwnd syscall.Handle) int {
	ret, _, _ := procGetWindowTextLengthW.Call(uintptr(hwnd))
	return int(ret)
}
func GetForegroundWindow() syscall.Handle {
	ret, _, _ := procGetForegroundWindow.Call()
	return syscall.Handle(ret)
}

func GetWindowText(hwnd syscall.Handle) string {
	length := GetWindowTextLength(hwnd)
	if length == 0 {
		return ""
	}
	buf := make([]uint16, length+1)
	procGetWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), uintptr(length+1))
	return syscall.UTF16ToString(buf)
}
func EnumWindows(enumFunc func(hwnd syscall.Handle, lParam uintptr) uintptr, lParam uintptr) error {
	ret, _, err := procEnumWindows.Call(syscall.NewCallback(enumFunc), lParam)
	if ret == 0 {
		return err
	}
	return nil
}

func IsIconic(hwnd syscall.Handle) bool {
	ret, _, _ := procIsIconic.Call(uintptr(hwnd))
	return ret != 0
}

func ShowWindow(hwnd syscall.Handle, cmdShow int32) {
	procShowWindow.Call(uintptr(hwnd), uintptr(cmdShow))
}

func SetForegroundWindow(hwnd syscall.Handle) bool {
	fgWindow := GetForegroundWindow()
	fgThreadID, _ := GetWindowThreadProcessId(fgWindow)
	targetThreadID, _ := GetWindowThreadProcessId(hwnd)

	if AttachThreadInput(fgThreadID, targetThreadID, true) {
		defer AttachThreadInput(fgThreadID, targetThreadID, false)

		// Appeler l'API Windows directement ici, via procSetForegroundWindow.Call
		ret, _, _ := procSetForegroundWindow.Call(uintptr(hwnd))
		if ret == 0 {
			log.Println("Failed to bring the window to the foreground.")
			return false
		}

		log.Println("Window successfully brought to the foreground.")
		return true
	} else {
		log.Println("Failed to attach thread input.")
		return false
	}
}

func AttachThreadInput(idAttach, idAttachTo uint32, fAttach bool) bool {
	ret, _, _ := procAttachThreadInput.Call(
		uintptr(idAttach),
		uintptr(idAttachTo),
		uintptr(boolToBOOL(fAttach)),
	)
	return ret != 0
}

func GetWindowThreadProcessId(hwnd syscall.Handle) (threadID uint32, processID uint32) {
	var pid uint32
	tid, _, _ := procGetWindowThreadProcessId.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&pid)))
	return uint32(tid), pid
}

func boolToBOOL(value bool) int32 {
	if value {
		return 1
	}
	return 0
}
