package services

import (
	"fmt"
	"log"
	"runtime"
	"sync"
	"syscall"
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
	procRegisterClassEx         = user32.NewProc("RegisterClassExW")
	procCreateWindowEx          = user32.NewProc("CreateWindowExW")
	procDefWindowProc           = user32.NewProc("DefWindowProcW")
)

// Constants for Windows messages and shell hook messages
const (
	WM_NCACTIVATE = 0x0086
	WM_DESTROY    = 0x0002

	HSHELL_HIGHBIT = 0x8000

	HSHELL_WINDOWCREATED       = 1
	HSHELL_WINDOWDESTROYED     = 2
	HSHELL_ACTIVATESHELLWINDOW = 3
	HSHELL_WINDOWACTIVATED     = 4
	HSHELL_GETMINRECT          = 5
	HSHELL_REDRAW              = 6
	HSHELL_FLASH               = HSHELL_REDRAW
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

type WNDCLASSEX struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     windows.Handle
	HIcon         windows.Handle
	HCursor       windows.Handle
	HbrBackground windows.Handle
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       windows.Handle
}

// StartTurnService monitors window events for a specific window
type StartTurnService struct {
	hwnd                windows.HWND
	targetHwnd          windows.HWND
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
	sts.windowTitle = windowTitle
	log.Printf("Starting service with window title: %s", windowTitle)

	// Register WM_SHELLHOOKMESSAGE
	wmShellHookMsg, err := registerWindowMessage("SHELLHOOK")
	if err != nil || wmShellHookMsg == 0 {
		return fmt.Errorf("failed to register WM_SHELLHOOKMESSAGE: %v", err)
	}
	sts.WM_SHELLHOOKMESSAGE = wmShellHookMsg
	log.Printf("Registered WM_SHELLHOOKMESSAGE: %d", wmShellHookMsg)

	// Find the target window by partial title
	log.Printf("Searching for window with title containing: %s", windowTitle)
	hwndTarget, err := sts.windowSvc.FindWindowByPartialTitle(windowTitle)
	if err != nil {
		return fmt.Errorf("could not find window with title %s: %v", windowTitle, err)
	}
	log.Printf("Window found: HWND = %d", hwndTarget)
	sts.targetHwnd = windows.HWND(hwndTarget)

	// Start monitoring window events
	go sts.monitorEvents()

	return nil
}

// monitorEvents listens for Windows messages in a message loop
func (sts *StartTurnService) monitorEvents() {
	runtime.LockOSThread()

	// Create message-only window
	hwnd, err := createMessageOnlyWindow()
	if err != nil {
		log.Fatalf("failed to create message-only window: %v", err)
	}

	// Register shell hook window
	log.Printf("Registering shell hook window for HWND: %d", hwnd)
	err = registerShellHookWindow(hwnd)
	if err != nil {
		log.Fatalf("failed to register shell hook window: %v", err)
	}
	log.Printf("Shell hook window successfully registered for HWND: %d", hwnd)

	sts.hwnd = hwnd
	sts.running = true

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

	// Extraire le code du message en masquant HSHELL_HIGHBIT
	messageCode := uint32(wParam & ^uintptr(HSHELL_HIGHBIT))
	log.Printf("Extracted message code: %d", messageCode)

	if windows.HWND(lParam) != sts.targetHwnd {
		log.Printf("Received shell hook message for a different window: %d", lParam)
		return
	}

	switch messageCode {
	case HSHELL_WINDOWACTIVATED:
		log.Printf("Target window activated, lParam=%d", lParam)
		// Vous pouvez ajouter un traitement ici si nécessaire

	case HSHELL_REDRAW:
		log.Printf("Notification captured: The target window has requested attention (HSHELL_REDRAW), lParam=%d", lParam)
		// Appeler FocusWindowWithTitle pour mettre la fenêtre au premier plan
		err := sts.windowSvc.FocusWindowWithTitle(sts.windowTitle)
		if err != nil {
			log.Printf("Failed to focus window: %v", err)
		} else {
			log.Println("Window focused successfully.")
		}

	default:
		log.Printf("Other shell event received: messageCode=%d, lParam=%d", messageCode, lParam)
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

// Create a message-only window
func createMessageOnlyWindow() (windows.HWND, error) {
	var className = windows.StringToUTF16Ptr("MessageOnlyWindowClass")

	var wcex WNDCLASSEX
	wcex.CbSize = uint32(unsafe.Sizeof(wcex))
	wcex.LpfnWndProc = syscall.NewCallback(messageOnlyWndProc)
	wcex.HInstance = windows.Handle(0)
	wcex.LpszClassName = className

	atom, _, err := procRegisterClassEx.Call(uintptr(unsafe.Pointer(&wcex)))
	if atom == 0 {
		return 0, err
	}

	hwnd, _, err := procCreateWindowEx.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		0,
		0,
		0,
		0,
		0,
		0,
		uintptr(HWND_MESSAGE), // Parent window
		0,
		0,
		0,
	)
	if hwnd == 0 {
		return 0, err
	}

	return windows.HWND(hwnd), nil
}

// Window procedure for the message-only window
func messageOnlyWndProc(hwnd windows.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_DESTROY:
		postQuitMessage(0)
		return 0
	default:
		ret, _, _ := procDefWindowProc.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
		return ret
	}
}

// Constants
const (
	HWND_MESSAGE = windows.Handle(^uintptr(2)) // Define HWND_MESSAGE as (HWND)-3
)
