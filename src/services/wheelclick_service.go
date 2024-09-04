package services

import (
	"log"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/go-vgo/robotgo"
)

const (
    WM_LBUTTONDOWN = 0x0201
    WM_LBUTTONUP   = 0x0202
    VK_MBUTTON     = 0x04 // Virtual key code for middle mouse button
)

type WheelClickService struct {
    stopChan chan struct{}
}

// GetCursorPos gets the current position of the cursor.
func GetCursorPos() (int, int) {
    return robotgo.GetMousePos()
}

// ConvertScreenToClient converts screen coordinates to client coordinates.
func ConvertScreenToClient(hWnd syscall.Handle, x, y int) (int, int) {
    var point struct {
        X, Y int32
    }
    point.X = int32(x)
    point.Y = int32(y)

    ret, _, _ := procScreenToClient.Call(uintptr(hWnd), uintptr(unsafe.Pointer(&point)))
    if ret == 0 {
        log.Printf("Erreur lors de la conversion des coordonnées écran en coordonnées client")
    }

    return int(point.X), int(point.Y)
}

// SimulateClick simulates a mouse click.
func SimulateClick(hWnd syscall.Handle, x, y int) {
    // Mettre la fenêtre au premier plan
    procSetForegroundWindow.Call(uintptr(hWnd))

    // Convertir les coordonnées d'écran en coordonnées client
    clientX, clientY := ConvertScreenToClient(hWnd, x, y)

    // Définir la position du curseur avec robotgo
    robotgo.MoveMouse(x, y)

    // Introduire un léger délai pour s'assurer que le curseur a bougé
    time.Sleep(50 * time.Millisecond)

    // Utiliser SendMessage pour simuler le clic aux coordonnées client
    log.Println("Envoi du message WM_LBUTTONDOWN")
    procSendMessage.Call(uintptr(hWnd), WM_LBUTTONDOWN, 0, uintptr(clientY<<16|clientX))
    time.Sleep(15 * time.Millisecond)
    log.Println("Envoi du message WM_LBUTTONUP")
    procSendMessage.Call(uintptr(hWnd), WM_LBUTTONUP, 0, uintptr(clientY<<16|clientX))
}

// Start detects middle mouse clicks.
func (wcs *WheelClickService) Start() {
    wcs.stopChan = make(chan struct{})
    go wcs.DetectMiddleClick()
}

// Stop stops detecting middle mouse clicks.
func (wcs *WheelClickService) Stop() {
    close(wcs.stopChan) // Signal to stop
}

// DetectMiddleClick listens for mouse button presses using robotgo.
func (wcs *WheelClickService) DetectMiddleClick() {
    for {
        select {
        case <-wcs.stopChan:
            log.Println("Stopping middle click detection")
            return
        default:
            if robotgo.AddMouse("mleft") { // Detect middle mouse click
                x, y := robotgo.GetMousePos()
                log.Printf("Middle click detected at: (%d, %d)", x, y)
                wcs.SendClickToDofusWindows(x, y)
            }
            time.Sleep(50 * time.Millisecond) // Reduce CPU usage
        }
    }
}

// SendClickToDofusWindows finds all windows containing "Dofus" and sends a click.
func (wcs *WheelClickService) SendClickToDofusWindows(x, y int) {
    windowService := &WindowService{}
    windows, err := windowService.GetWindows()
    if err != nil {
        log.Printf("Error getting windows: %v", err)
        return
    }

    for _, windowTitle := range windows {
        if strings.Contains(windowTitle, "Dofus") {
            log.Printf("Sending click to window: %s", windowTitle)
            hWnd := GetWindowHandle(windowTitle)
            log.Printf("Clicking at: X=%d, Y=%d", x, y)
            SimulateClick(hWnd, x, y)
        }
    }
}

// GetWindowHandle retrieves the handle of the window with the given title.
func GetWindowHandle(title string) syscall.Handle {
    titleUTF16, _ := syscall.UTF16PtrFromString(title)
    hWnd, _, _ := procFindWindow.Call(0, uintptr(unsafe.Pointer(titleUTF16)))
    return syscall.Handle(hWnd)
}
