package services

import (
	"fmt"
	"log"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	procGetMessage              = user32.NewProc("GetMessageW")
	procTranslateMessage        = user32.NewProc("TranslateMessage")
	procDispatchMessage         = user32.NewProc("DispatchMessageW")
	procRegisterShellHookWindow = user32.NewProc("RegisterShellHookWindow")
	procRegisterWindowMessage   = user32.NewProc("RegisterWindowMessageW")
	procPostQuitMessage         = user32.NewProc("PostQuitMessage")
)

// Constants for Windows messages and shell hook messages
const (
	WM_NCACTIVATE = 0x0086

	HSHELL_FLASH            = 0x8004
	HSHELL_WINDOWACTIVATED  = 0x8006
	HSHELL_WINDOWACTIVATION = 0x0006
)

// MSG structure for Windows messages
type MSG struct {
	HWnd    windows.Handle
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}

type POINT struct {
	X, Y int32
}

// StartTurnService monitors window events for a specific window
type StartTurnService struct {
	hwnd                windows.HWND
	running             bool
	mutex               sync.Mutex
	windowTitle         string
	windowSvc           *WindowService
	lastLParam          uintptr
	lastWParam          uintptr
	lastWMNCActivate    bool
	WM_SHELLHOOKMESSAGE uint32
}

// NewStartTurnService creates a new StartTurnService
func NewStartTurnService(ws *WindowService) *StartTurnService {
	return &StartTurnService{
		windowSvc: ws,
	}
}

// Start initiates the service and registers the shell hook
func (sts *StartTurnService) Start(windowTitle string) error {
	sts.mutex.Lock()
	defer sts.mutex.Unlock()

	if sts.running {
		log.Println("Service is already running.")
		return nil
	}

	log.Printf("Starting service with window title: %s", windowTitle)

	// Register WM_SHELLHOOKMESSAGE
	wmShellHookMsg, err := registerWindowMessage("SHELLHOOK")
	if err != nil || wmShellHookMsg == 0 {
		return fmt.Errorf("failed to register WM_SHELLHOOKMESSAGE: %v", err)
	}
	sts.WM_SHELLHOOKMESSAGE = wmShellHookMsg
	log.Printf("Registered WM_SHELLHOOKMESSAGE: %d", wmShellHookMsg)

	// Find the window by partial title
	log.Printf("Searching for window with title containing: %s", windowTitle)
	hwnd, err := sts.windowSvc.FindWindowByPartialTitle(windowTitle)
	if err != nil {
		return fmt.Errorf("could not find window with title %s: %v", windowTitle, err)
	}
	log.Printf("Window found: HWND = %d", hwnd)

	// Ensure the window handle is valid
	if hwnd == 0 {
		return fmt.Errorf("invalid HWND: %d", hwnd)
	}

	// Register the window to receive shell messages
	log.Printf("Registering shell hook window for HWND: %d", hwnd)
	err = registerShellHookWindow(windows.HWND(hwnd))
	if err != nil {
		return fmt.Errorf("failed to register shell hook window: %v", err)
	}
	log.Printf("Shell hook window successfully registered for HWND: %d", hwnd)

	sts.hwnd = windows.HWND(hwnd)
	sts.windowTitle = windowTitle
	sts.running = true

	// Start monitoring window events
	go sts.monitorEvents()

	return nil
}

// monitorEvents listens for Windows messages in a message loop
func (sts *StartTurnService) monitorEvents() {
	var msg MSG
	log.Println("Starting to monitor window messages")

	for {
		ret, err := getMessage(&msg)
		if ret == 0 {
			log.Println("WM_QUIT received, exiting message loop.")
			return
		} else if ret == -1 {
			log.Printf("GetMessage returned an error: %v", err)
			continue
		}

		log.Printf("Message received: HWND=%d, Message=%d, WParam=%d, LParam=%d", msg.HWnd, msg.Message, msg.WParam, msg.LParam)

		// Handle messages
		sts.handleMessage(&msg)

		translateMessage(&msg)
		dispatchMessage(&msg)
	}
}

// handleMessage processes incoming Windows messages
func (sts *StartTurnService) handleMessage(msg *MSG) {
	// Handle WM_NCACTIVATE
	if msg.Message == WM_NCACTIVATE {
		sts.lastWMNCActivate = msg.WParam != 0
		log.Printf("WM_NCACTIVATE detected: Active=%t", sts.lastWMNCActivate)
	}

	// Intercept WM_SHELLHOOKMESSAGE
	if msg.Message == sts.WM_SHELLHOOKMESSAGE {
		log.Printf("WM_SHELLHOOKMESSAGE detected: WParam=%d, LParam=%d", msg.WParam, msg.LParam)
		sts.handleShellHookMessage(msg.WParam, msg.LParam)
	}
}

// handleShellHookMessage processes WM_SHELLHOOKMESSAGE events
func (sts *StartTurnService) handleShellHookMessage(wParam, lParam uintptr) {
	log.Printf("Handling WM_SHELLHOOKMESSAGE with WParam=%d, LParam=%d", wParam, lParam)

	switch wParam {
	case HSHELL_FLASH:
		log.Printf("Notification captured: A window has requested attention, lParam=%d", lParam)

	case HSHELL_WINDOWACTIVATED:
		log.Printf("Window state change detected, lParam=%d", lParam)

		// Compare the previous values to detect a loop or repetition
		if sts.lastLParam == lParam && sts.lastWParam == wParam {
			log.Printf("Repeated shell message detected for lParam=%d, ignoring...", lParam)
			return
		}

		// Handle window state changes specifically for WM_NCACTIVATE sequence
		if sts.lastWMNCActivate {
			log.Println("Window activated/deactivated sequence detected.")
		}
		sts.lastLParam = lParam
		sts.lastWParam = wParam

	case HSHELL_WINDOWACTIVATION:
		log.Printf("Window activation/deactivation event detected, lParam=%d", lParam)

	default:
		log.Printf("Other shell event received: wParam=%d, lParam=%d", wParam, lParam)
	}
}

// Stop stops the service of message capturing
func (sts *StartTurnService) Stop() {
	sts.mutex.Lock()
	defer sts.mutex.Unlock()

	if !sts.running {
		log.Println("Service is not running.")
		return
	}

	log.Println("Stopping the service")
	postQuitMessage(0)
	sts.running = false
	log.Println("Service stopped.")
}

// Helper function to register a window message
func registerWindowMessage(lpString string) (uint32, error) {
	ret, _, err := procRegisterWindowMessage.Call(uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(lpString))))
	if ret == 0 {
		if err == nil {
			return 0, windows.GetLastError()
		}
		return 0, err
	}
	return uint32(ret), nil
}

// Helper function to register the shell hook window
func registerShellHookWindow(hwnd windows.HWND) error {
	ret, _, err := procRegisterShellHookWindow.Call(uintptr(hwnd))
	if ret == 0 {
		if err == nil {
			return windows.GetLastError()
		}
		return err
	}
	return nil
}

// Helper function to post a quit message to the message loop
func postQuitMessage(exitCode int32) {
	procPostQuitMessage.Call(uintptr(exitCode))
}

// Implementing GetMessage
func getMessage(msg *MSG) (int32, error) {
	ret, _, err := procGetMessage.Call(
		uintptr(unsafe.Pointer(msg)),
		0,
		0,
		0,
	)
	if ret == 0 {
		return 0, nil // WM_QUIT received
	}
	if ret == ^uintptr(0) { // -1 cast to uintptr
		return -1, err
	}
	return int32(ret), nil
}

// Implementing TranslateMessage
func translateMessage(msg *MSG) {
	procTranslateMessage.Call(uintptr(unsafe.Pointer(msg)))
}

// Implementing DispatchMessage
func dispatchMessage(msg *MSG) {
	procDispatchMessage.Call(uintptr(unsafe.Pointer(msg)))
}
