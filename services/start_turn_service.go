package services

import (
	"log"
	"syscall"
	"time"
	"unsafe"
)

type HWND syscall.Handle

const (
	WM_SHELLHOOKMESSAGE = 49193
	NOTIFICATION_WPARAM = 32774
)

type StartTurnService struct {
	hwnd         syscall.Handle
	running      bool
	stopChan     chan struct{}
	windowService *WindowService // Dependency injection for WindowService
}

type MSG struct {
	HWnd    HWND
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}

type POINT struct {
	X, Y int32
}
// NewStartTurnService creates a new tracking service.
func NewStartTurnService(windowService *WindowService) *StartTurnService {
	return &StartTurnService{
		windowService: windowService,
	}
}

// Start starts the service.
func (sts *StartTurnService) Start() {
	if sts.running {
		log.Println("Service is already running.")
		return
	}

	sts.stopChan = make(chan struct{})
	sts.running = true
	go sts.monitorWindows()
	log.Println("Service started, monitoring for specific Windows messages.")
}

// Stop stops the service.
func (sts *StartTurnService) Stop() {
	if !sts.running {
		log.Println("Service is not running.")
		return
	}

	close(sts.stopChan)
	sts.running = false
	log.Println("Service stopped.")
}

// monitorWindows monitors for the specific Windows messages.
func (sts *StartTurnService) monitorWindows() {
	for {
		select {
		case <-sts.stopChan:
			return
		default:
			sts.processWindowsMessages()
			time.Sleep(1 * time.Second)
		}
	}
}

// processWindowsMessages listens for the specific Windows messages.
func (sts *StartTurnService) processWindowsMessages() {
	user32 := syscall.NewLazyDLL("user32.dll")
	getMessage := user32.NewProc("GetMessageW")
	translateMessage := user32.NewProc("TranslateMessage")
	dispatchMessage := user32.NewProc("DispatchMessageW")

	var msg MSG

	for {
		ret, _, _ := getMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if ret == 0 {
			break
		}

		translateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		dispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))

		if msg.Message == WM_SHELLHOOKMESSAGE && msg.WParam == NOTIFICATION_WPARAM {
			log.Printf("Received WM_SHELLHOOKMESSAGE with NOTIFICATION_WPARAM in window %d", msg.HWnd)
			sts.handleShellHookMessage(msg.HWnd)
		}
	}
}

// handleShellHookMessage handles the shell hook messages.
func (sts *StartTurnService) handleShellHookMessage(hwnd HWND) {
	// Implement the logic for when the message is detected
	log.Printf("Handling shell hook message for window %d", hwnd)
}
